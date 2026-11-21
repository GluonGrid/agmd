package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
		build := exec.Command("go", "build", "-o", agmdBinPath, ".")
		build.Dir = root
		out, err := build.CombinedOutput()
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

func TestTaskConcurrency_ClaimNeverHandsOutAHumanTask(t *testing.T) {
	registry := setup(t)
	const agentTasks, n = 6, 10
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "-p", "0", "--content", "Only the human can merge")
	for i := 0; i < agentTasks; i++ {
		run(t, "task", "new", fmt.Sprintf("t%02d", i), "--project", "p", "-p", "3", "--content", "work")
	}

	results := runParallel(t, n, func(i int) []string {
		return []string{"task", "next", "--claim", "--workflow-id", fmt.Sprintf("agent-%d", i), "--project", "p", "--json"}
	})

	claimed := map[string]int{}
	notFound := 0
	for i, r := range results {
		if r.err != nil {
			if exitErr, ok := r.err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
				notFound++
				continue
			}
			t.Fatalf("agent-%d failed: %v", i, r.err)
		}
		var payload struct {
			Task struct {
				Name  string `json:"name"`
				Human bool   `json:"human"`
			} `json:"task"`
		}
		if err := json.Unmarshal(r.stdout, &payload); err != nil {
			t.Fatalf("agent-%d output is not JSON: %v\n%s", i, err, r.stdout)
		}
		if payload.Task.Name == "merge-pr" || payload.Task.Human {
			t.Errorf("agent-%d was handed the human task", i)
		}
		claimed[payload.Task.Name]++
	}
	if len(claimed) != agentTasks || notFound != n-agentTasks {
		t.Errorf("claimed %d distinct tasks with %d not_found, want %d and %d: %v", len(claimed), notFound, agentTasks, n-agentTasks, claimed)
	}
	for name, count := range claimed {
		if count != 1 {
			t.Errorf("task %s claimed %d times", name, count)
		}
	}
	for _, a := range readAssignments(t, registry, "p") {
		if a["task_name"] == "merge-pr" {
			t.Errorf("assignments.json holds the human task: %v", a)
		}
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

func TestTaskJSONErrorIsOneLineOnStderr(t *testing.T) {
	setup(t)
	cmd := exec.Command(agmdBinary(t), "task", "next", "--claim", "--workflow-id", "x", "--project", "empty", "--json")
	cmd.Env = os.Environ()
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		t.Fatal("expected a non-zero exit")
	}
	got := strings.TrimSpace(stderr.String())
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if strings.Contains(got, "\n") || json.Unmarshal([]byte(got), &payload) != nil || payload.Error.Code != "not_found" {
		t.Errorf("stderr should be exactly one JSON error line, got:\n%s", got)
	}
}

// A reopen racing parallel claims of its dependent: either the reopen wins
// and nobody gets the dependent, or one claim wins and the reopen is refused.
func TestTaskConcurrency_ReopenRacesClaimOfDependent(t *testing.T) {
	setup(t)
	const rounds = 12
	reopenWins, claimWins := 0, 0
	for round := 0; round < rounds; round++ {
		claimers := 1 + round%4 // fewer rivals in some rounds, so the reopen wins some
		project := fmt.Sprintf("p%d", round)
		// a is the human's step (ticked in the app); b is the only agent task.
		run(t, "task", "new", "a", "--project", project, "--human", "--content", "first")
		run(t, "task", "new", "b", "--project", project, "--blocked-by", "a", "--content", "second")
		run(t, "task", "status", "a", "completed", "--project", project)

		// Process 0 reopens, the rest claim; all start at once, and several
		// rounds let scheduling produce both orderings.
		results := runParallel(t, claimers+1, func(i int) []string {
			if i == 0 {
				return []string{"task", "reopen", "a", "--project", project, "--json"}
			}
			return []string{"task", "next", "--claim", "--workflow-id", fmt.Sprintf("agent-%d", i), "--project", project, "--json"}
		})

		exitCode := func(err error) int {
			if err == nil {
				return 0
			}
			if exitErr, ok := err.(*exec.ExitError); ok {
				return exitErr.ExitCode()
			}
			t.Fatalf("round %d: %v", round, err)
			return -1
		}
		reopened := false
		switch code := exitCode(results[0].err); code {
		case 0:
			reopened = true
		case 6:
		default:
			t.Fatalf("round %d: reopen exit %d, want 0 or 6", round, code)
		}
		claims := 0
		for i := 1; i <= claimers; i++ {
			switch code := exitCode(results[i].err); code {
			case 0:
				if !strings.Contains(string(results[i].stdout), `"name":"b"`) {
					t.Errorf("round %d: agent-%d claimed something else: %s", round, i, results[i].stdout)
				}
				claims++
			case 3:
			default:
				t.Fatalf("round %d: agent-%d exit %d", round, i, code)
			}
		}
		switch {
		case reopened && claims == 0:
			reopenWins++
		case !reopened && claims == 1:
			claimWins++
		default:
			t.Errorf("round %d: reopened=%v with %d claims of b; want reopen xor one claim", round, reopened, claims)
		}
	}
	t.Logf("reopen won %d rounds, a claim won %d", reopenWins, claimWins)
}

// Parallel updates and dependency links on one task: no write may undo
// another, and the claim and comments must survive.
func TestTaskConcurrency_UpdateKeepsParallelDepAdds(t *testing.T) {
	registry := setup(t)
	const n = 8
	run(t, "task", "new", "hub", "--project", "p", "--content", "start")
	run(t, "task", "comment", "hub", "keep me", "--project", "p", "--author", "x")
	run(t, "task", "assign", "hub", "--project", "p", "--workflow-id", "wf-1")
	for i := 0; i < n; i++ {
		run(t, "task", "new", fmt.Sprintf("d%d", i), "--project", "p", "--content", "dep")
	}

	results := runParallel(t, 2*n, func(i int) []string {
		if i%2 == 0 {
			return []string{"task", "update", "hub", "--project", "p", "--content", fmt.Sprintf("text-%d", i/2)}
		}
		return []string{"task", "dep", "add", "hub", fmt.Sprintf("d%d", i/2), "--project", "p"}
	})
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("process %d failed: %v", i, r.err)
		}
	}

	hub := showFullTask(t, "hub")
	if len(hub.DependsOn) != n {
		t.Errorf("hub has %d dependencies, want %d (lost updates): %v", len(hub.DependsOn), n, hub.DependsOn)
	}
	if !strings.HasPrefix(hub.Content, "text-") {
		t.Errorf("hub content = %q, want one of the updates", hub.Content)
	}
	if len(hub.Comments) != 1 {
		t.Errorf("comments = %+v", hub.Comments)
	}
	if a := readAssignments(t, registry, "p"); len(a) != 1 {
		t.Errorf("claim lost: %v", a)
	}
}

func TestTaskConcurrency_ParallelDepRemovesAreNotLost(t *testing.T) {
	setup(t)
	const n = 8
	deps := []string{}
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("d%d", i)
		run(t, "task", "new", name, "--project", "p", "--content", "dep")
		deps = append(deps, name)
	}
	run(t, "task", "new", "hub", "--project", "p", "--content", "hub", "--blocked-by", strings.Join(deps, ","))

	results := runParallel(t, n, func(i int) []string {
		return []string{"task", "dep", "remove", "hub", fmt.Sprintf("d%d", i), "--project", "p"}
	})
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("process %d failed: %v", i, r.err)
		}
	}
	if left := showFullTask(t, "hub").DependsOn; len(left) != 0 {
		t.Errorf("dependencies left after parallel removes: %v", left)
	}
}
