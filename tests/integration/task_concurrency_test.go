package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// These tests run a real agmd binary as parallel subprocesses: in-process
// ExecuteArgs shares package-level flag variables and cannot run concurrently.

var (
	agmdBinOnce sync.Once
	agmdBinPath string
	agmdBinErr  error
)

func agmdBinary(t *testing.T) string {
	t.Helper()
	agmdBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "agmd-bin-")
		if err != nil {
			agmdBinErr = err
			return
		}
		agmdBinPath = filepath.Join(dir, "agmd")
		_, thisFile, _, _ := runtime.Caller(0)
		root := filepath.Join(filepath.Dir(thisFile), "..", "..")
		out, err := exec.Command("go", "build", "-o", agmdBinPath, root).CombinedOutput()
		if err != nil {
			agmdBinErr = fmt.Errorf("go build failed: %v\n%s", err, out)
		}
	})
	if agmdBinErr != nil {
		t.Fatal(agmdBinErr)
	}
	return agmdBinPath
}

type procResult struct {
	stdout []byte
	err    error
}

// runParallel starts n agmd processes at once; argsFor(i) gives each one's args.
func runParallel(t *testing.T, n int, argsFor func(i int) []string) []procResult {
	t.Helper()
	bin := agmdBinary(t)
	results := make([]procResult, n)
	var start, done sync.WaitGroup
	start.Add(1)
	for i := 0; i < n; i++ {
		done.Add(1)
		go func(i int) {
			defer done.Done()
			cmd := exec.Command(bin, argsFor(i)...)
			cmd.Env = os.Environ()
			start.Wait()
			out, err := cmd.Output()
			results[i] = procResult{stdout: out, err: err}
		}(i)
	}
	start.Done()
	done.Wait()
	return results
}

func readAssignments(t *testing.T, registry, project string) map[string]map[string]string {
	t.Helper()
	raw := readFile(t, filepath.Join(registry, "task", project, "assignments.json"))
	assignments := map[string]map[string]string{}
	if err := json.Unmarshal([]byte(raw), &assignments); err != nil {
		t.Fatalf("assignments.json is not valid JSON: %v\n%s", err, raw)
	}
	return assignments
}

func TestTaskConcurrency_ClaimNeverHandsOutATaskTwice(t *testing.T) {
	registry := setup(t)
	const n = 12
	for i := 0; i < n; i++ {
		run(t, "task", "new", fmt.Sprintf("t%02d", i), "--project", "p", "--content", "work")
	}

	results := runParallel(t, n, func(i int) []string {
		return []string{"task", "next", "--claim", "--workflow-id", fmt.Sprintf("agent-%d", i), "--project", "p", "--json"}
	})

	claimed := map[string]int{}
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("agent-%d failed: %v", i, r.err)
		}
		var payload struct {
			Action string `json:"action"`
			Task   struct {
				Name string `json:"name"`
			} `json:"task"`
		}
		if err := json.Unmarshal(r.stdout, &payload); err != nil {
			t.Fatalf("agent-%d output is not JSON: %v\n%s", i, err, r.stdout)
		}
		if payload.Action != "claim" {
			t.Errorf("agent-%d action = %q, want claim", i, payload.Action)
		}
		claimed[payload.Task.Name]++
	}
	if len(claimed) != n {
		t.Errorf("expected %d distinct tasks claimed, got %d: %v", n, len(claimed), claimed)
	}
	for name, count := range claimed {
		if count != 1 {
			t.Errorf("task %s claimed %d times", name, count)
		}
	}
	if got := len(readAssignments(t, registry, "p")); got != n {
		t.Errorf("assignments.json has %d entries, want %d (lost update)", got, n)
	}
}

func TestTaskConcurrency_ParallelAssignsAreNotLost(t *testing.T) {
	registry := setup(t)
	const n = 12
	for i := 0; i < n; i++ {
		run(t, "task", "new", fmt.Sprintf("t%02d", i), "--project", "p", "--content", "work")
	}

	results := runParallel(t, n, func(i int) []string {
		return []string{"task", "assign", fmt.Sprintf("t%02d", i), "--workflow-id", fmt.Sprintf("agent-%d", i), "--project", "p"}
	})
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("assign %d failed: %v", i, r.err)
		}
	}
	if got := len(readAssignments(t, registry, "p")); got != n {
		t.Errorf("assignments.json has %d entries, want %d (lost update)", got, n)
	}
}

func TestTaskConcurrency_ParallelNewSameNameCreatesOnce(t *testing.T) {
	setup(t)
	const n = 8

	results := runParallel(t, n, func(i int) []string {
		return []string{"task", "new", "shared", "--project", "p", "--content", fmt.Sprintf("from agent %d", i)}
	})
	succeeded := 0
	for _, r := range results {
		if r.err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Errorf("%d processes created task 'shared', want exactly 1", succeeded)
	}
}
