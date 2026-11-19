package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agmd/cmd"
	"agmd/pkg/registry"
)

// requireCode fails unless err is a coded agmd error with the given code and exit status.
func requireCode(t *testing.T, err error, code string, exit int) {
	t.Helper()
	type coded interface {
		ErrorCode() string
		ExitStatus() int
	}
	c, ok := err.(coded)
	if !ok {
		t.Fatalf("expected coded error %q, got %T (%v)", code, err, err)
	}
	if c.ErrorCode() != code || c.ExitStatus() != exit {
		t.Fatalf("error = %s/%d, want %s/%d (%v)", c.ErrorCode(), c.ExitStatus(), code, exit, err)
	}
}

type humanListItem struct {
	Name    string `json:"name"`
	Feature string `json:"feature"`
	Human   bool   `json:"human"`
}

func listNames(t *testing.T, args ...string) map[string]humanListItem {
	t.Helper()
	out := run(t, append([]string{"task", "list", "--project", "p", "--json", "--all"}, args...)...)
	var payload struct {
		Tasks []humanListItem `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode list: %v\n%s", err, out)
	}
	items := map[string]humanListItem{}
	for _, it := range payload.Tasks {
		items[it.Name] = it
	}
	return items
}

func TestTaskHuman_NewWritesFieldOnlyWhenTrue(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "--content", "Merge it")
	run(t, "task", "new", "code", "--project", "p", "--content", "Write it")

	if raw := readFile(t, filepath.Join(registryPath, "task", "p", "merge-pr.md")); !strings.Contains(raw, "human: true") {
		t.Errorf("human task file lacks 'human: true':\n%s", raw)
	}
	if raw := readFile(t, filepath.Join(registryPath, "task", "p", "code.md")); strings.Contains(raw, "human") {
		t.Errorf("agent task file should not mention human:\n%s", raw)
	}

	items := listNames(t)
	if !items["merge-pr"].Human || items["code"].Human {
		t.Errorf("JSON human flags wrong: %+v", items)
	}
	// The raw JSON omits the field on agent tasks.
	out := run(t, "task", "show", "code", "--project", "p", "--json")
	if strings.Contains(out, `"human"`) {
		t.Errorf("agent task JSON should omit human: %s", out)
	}
}

func TestTaskHuman_ListFiltersAndTextMarks(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "--content", "Merge it")
	run(t, "task", "new", "code", "--project", "p", "--content", "Write it")

	if all := listNames(t); len(all) != 2 {
		t.Errorf("default list should show both, got %v", all)
	}
	if only := listNames(t, "--human"); len(only) != 1 || !only["merge-pr"].Human {
		t.Errorf("--human should show only merge-pr, got %v", only)
	}
	if none := listNames(t, "--no-human"); len(none) != 1 || none["code"].Name == "" {
		t.Errorf("--no-human should show only code, got %v", none)
	}
	_, err := runE(t, "task", "list", "--project", "p", "--human", "--no-human", "--json")
	requireCode(t, err, "invalid_input", 2)

	text := run(t, "task", "list", "--project", "p")
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "merge-pr") && !strings.Contains(line, "[human]") {
			t.Errorf("list line for human task lacks [human]: %q", line)
		}
		if strings.HasSuffix(line, " code") && strings.Contains(line, "[human]") {
			t.Errorf("agent task marked human: %q", line)
		}
	}
	if show := run(t, "task", "show", "merge-pr", "--project", "p"); !strings.Contains(show, "for: human") {
		t.Errorf("task show lacks 'for: human':\n%s", show)
	}
	if show := run(t, "task", "show", "code", "--project", "p"); strings.Contains(show, "for:") {
		t.Errorf("agent task show should not have a for: line:\n%s", show)
	}
}

func TestTaskHuman_EditSetsAndClearsHumanAndFeature(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "login", "--project", "p", "--feature", "auth", "--content", "Log in")

	if _, err := runE(t, "task", "edit", "login", "--project", "p", "--human"); err != nil {
		t.Fatalf("edit --human: %v", err)
	}
	if it := listNames(t)["login"]; !it.Human || it.Feature != "auth" {
		t.Errorf("after --human: %+v", it)
	}

	if _, err := runE(t, "task", "edit", "login", "--project", "p", "--feature", "deploy"); err != nil {
		t.Fatalf("edit --feature: %v", err)
	}
	if it := listNames(t)["login"]; !it.Human || it.Feature != "deploy" {
		t.Errorf("--feature should keep human and set feature: %+v", it)
	}

	if _, err := runE(t, "task", "edit", "login", "--project", "p", "--human=false", "--feature", ""); err != nil {
		t.Fatalf("edit --human=false --feature '': %v", err)
	}
	if it := listNames(t)["login"]; it.Human || it.Feature != "" {
		t.Errorf("after clearing: %+v", it)
	}
	if raw := readFile(t, filepath.Join(registryPath, "task", "p", "login.md")); strings.Contains(raw, "human") || strings.Contains(raw, "feature") {
		t.Errorf("cleared fields should be gone from the file:\n%s", raw)
	}

	_, err := runE(t, "task", "edit", "login", "--project", "p")
	requireCode(t, err, "invalid_input", 2)
}

func TestTaskHuman_NextSkipsHumanTasks(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "-p", "0", "--content", "Merge it")

	// Only a human task is ready: nothing to hand out, peek or claim.
	_, err := runE(t, "task", "next", "--project", "p", "--json")
	requireCode(t, err, "not_found", 3)
	_, err = claim(t, "agent-1")
	requireCode(t, err, "not_found", 3)

	run(t, "task", "new", "code", "--project", "p", "-p", "3", "--content", "Write it")
	out := run(t, "task", "next", "--project", "p", "--json")
	if !strings.Contains(out, `"name":"code"`) {
		t.Errorf("peek should return the agent task, got %s", out)
	}
	resp, err := claim(t, "agent-1")
	if err != nil || resp.Task.Name != "code" || resp.Action != "claim" {
		t.Fatalf("claim = %+v, %v; want code/claim", resp, err)
	}
	// Resume still works for agent tasks.
	if resp, err = claim(t, "agent-1"); err != nil || resp.Action != "resume" || resp.Task.Name != "code" {
		t.Errorf("resume = %+v, %v", resp, err)
	}
}

func TestTaskHuman_StaleTakeoverAndHeldHumanTask(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "code", "--project", "p", "--content", "Write it")
	if _, err := claim(t, "agent-1"); err != nil {
		t.Fatal(err)
	}
	ageClaims(t, registryPath, 3*time.Hour)
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "-p", "0", "--content", "Merge it")

	resp, err := claim(t, "agent-2")
	if err != nil || resp.Action != "takeover" || resp.Task.Name != "code" {
		t.Fatalf("takeover = %+v, %v; want takeover of code", resp, err)
	}

	// A task marked human after an agent claimed it is not resumed.
	run(t, "task", "edit", "code", "--project", "p", "--human")
	_, err = claim(t, "agent-2")
	requireCode(t, err, "not_found", 3)
}

func TestTaskHuman_AssignRefusedEvenWithForce(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "--content", "Merge it")

	for _, extra := range [][]string{nil, {"--force"}} {
		args := append([]string{"task", "assign", "merge-pr", "--project", "p", "--workflow-id", "agent-1", "--json"}, extra...)
		_, err := runE(t, args...)
		requireCode(t, err, "conflict", 6)
		if !strings.Contains(err.Error(), "for the human") {
			t.Errorf("message should say the task is for the human: %v", err)
		}
	}
	if out := run(t, "task", "show", "merge-pr", "--project", "p", "--json"); strings.Contains(out, "assignment") {
		t.Errorf("human task got an assignment: %s", out)
	}
}

func TestTaskHuman_CompletingReleasesDependents(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "login", "--project", "p", "--human", "--content", "Log in")
	run(t, "task", "new", "deploy", "--project", "p", "--blocked-by", "login", "--content", "Deploy")

	_, err := claim(t, "agent-1")
	requireCode(t, err, "not_found", 3)

	if _, err := runE(t, "task", "status", "login", "completed", "--project", "p"); err != nil {
		t.Fatalf("status on human task: %v", err)
	}
	resp, err := claim(t, "agent-1")
	if err != nil || resp.Task.Name != "deploy" {
		t.Fatalf("after the human finished, claim = %+v, %v; want deploy", resp, err)
	}
}

func TestTaskHuman_WebStateAndRunRefused(t *testing.T) {
	setup(t)
	chdir(t)
	t.Setenv("HOME", t.TempDir())
	writeStubClaude(t)
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "--content", "Merge it")
	run(t, "task", "new", "code", "--project", "p", "--content", "Write it")

	reg, err := registry.New()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "p"))
	defer srv.Close()

	_, data := webGet(t, srv.URL+"/api/state")
	var state struct {
		Tasks []humanListItem `json:"tasks"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	for _, task := range state.Tasks {
		if task.Human != (task.Name == "merge-pr") {
			t.Errorf("web task %s human = %v", task.Name, task.Human)
		}
	}

	status, body := webPost(t, srv.URL+"/api/run", map[string]interface{}{
		"tasks":          []map[string]string{{"project": "p", "id": "code"}, {"project": "p", "id": "merge-pr"}},
		"permissionMode": "acceptEdits",
	})
	if status != http.StatusConflict || !strings.Contains(string(body), "for the human") {
		t.Fatalf("run on a human task = %d %s, want 409 with a clear message", status, body)
	}
	// Nothing was claimed: the run was refused before any task was touched.
	if out := run(t, "task", "show", "code", "--project", "p", "--json"); strings.Contains(out, "assignment") || !strings.Contains(out, `"status":"pending"`) {
		t.Errorf("refused run still changed the agent task: %s", out)
	}
}
