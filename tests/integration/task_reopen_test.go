package integration

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agmd/cmd"
	"agmd/pkg/registry"
)

type reopenTaskItem struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	ComputedStatus string   `json:"computed_status"`
	PendingDeps    []string `json:"pending_dependencies"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	CompletedAt    string   `json:"completed_at"`
	HoldUntil      string   `json:"hold_until"`
}

func showTask(t *testing.T, name string) reopenTaskItem {
	t.Helper()
	out, err := runE(t, "task", "show", name, "--project", "p", "--json")
	if err != nil {
		t.Fatalf("show %s: %v", name, err)
	}
	var item reopenTaskItem
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("decode show: %v\n%s", err, out)
	}
	return item
}

// chainAB creates a (done first) and b, which depends on a.
func chainAB(t *testing.T) {
	t.Helper()
	run(t, "task", "new", "a", "--project", "p", "--content", "first")
	run(t, "task", "new", "b", "--project", "p", "--blocked-by", "a", "--content", "second")
}

func TestTaskReopen_ReadyDependentBecomesBlockedAgain(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	run(t, "task", "status", "a", "completed", "--project", "p")
	if got := showTask(t, "b").ComputedStatus; got != "ready" {
		t.Fatalf("b = %s before reopen, want ready", got)
	}

	if _, err := runE(t, "task", "reopen", "a", "--project", "p", "--author", "app"); err != nil {
		t.Fatalf("reopen: %v", err)
	}
	a := showTask(t, "a")
	if a.Status != "pending" || a.CompletedAt != "" {
		t.Errorf("a after reopen = %+v, want pending without completed_at", a)
	}
	if got := showTask(t, "b").ComputedStatus; got != "blocked" {
		t.Errorf("b = %s after reopen, want blocked", got)
	}
	out := run(t, "task", "show", "a", "--project", "p")
	if !strings.Contains(out, "Reopened") || !strings.Contains(out, "app") {
		t.Errorf("show should list the reopen comment by app:\n%s", out)
	}
}

func TestTaskReopen_RefusedWhenDependentIsClaimed(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	run(t, "task", "status", "a", "completed", "--project", "p")
	run(t, "task", "assign", "b", "--workflow-id", "agent-1", "--project", "p")

	_, err := runE(t, "task", "reopen", "a", "--project", "p")
	requireCode(t, err, "conflict", 6)
	if !strings.Contains(err.Error(), "'b'") {
		t.Errorf("error should name the dependent b: %v", err)
	}
	if got := showTask(t, "a").Status; got != "completed" {
		t.Errorf("a = %s after refused reopen, want completed", got)
	}
}

func TestTaskReopen_RefusedWhenDependentInProgress(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	run(t, "task", "status", "a", "completed", "--project", "p")
	run(t, "task", "status", "b", "in_progress", "--project", "p")

	_, err := runE(t, "task", "reopen", "a", "--project", "p")
	requireCode(t, err, "conflict", 6)
	if !strings.Contains(err.Error(), "'b'") {
		t.Errorf("error should name the dependent b: %v", err)
	}
}

func TestTaskReopen_RefusedWhenDependentCompleted(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	run(t, "task", "status", "a", "completed", "--project", "p")
	run(t, "task", "status", "b", "completed", "--project", "p")

	_, err := runE(t, "task", "reopen", "a", "--project", "p")
	requireCode(t, err, "conflict", 6)
}

func TestTaskReopen_Chain(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	run(t, "task", "new", "c", "--project", "p", "--blocked-by", "b", "--content", "third")
	run(t, "task", "status", "a", "completed", "--project", "p")
	run(t, "task", "status", "b", "completed", "--project", "p")

	// a released b and b is completed: a is final.
	_, err := runE(t, "task", "reopen", "a", "--project", "p")
	requireCode(t, err, "conflict", 6)
	// c is only ready, so b can still be undone.
	if _, err := runE(t, "task", "reopen", "b", "--project", "p"); err != nil {
		t.Fatalf("reopen b: %v", err)
	}
	if got := showTask(t, "c").ComputedStatus; got != "blocked" {
		t.Errorf("c = %s, want blocked", got)
	}
	// Now nothing completed depends on a.
	if _, err := runE(t, "task", "reopen", "a", "--project", "p"); err != nil {
		t.Fatalf("reopen a after b: %v", err)
	}
}

func TestTaskReopen_NotCompletedIsInvalidTransition(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	_, err := runE(t, "task", "reopen", "a", "--project", "p")
	requireCode(t, err, "invalid_transition", 5)
	// status still cannot leave completed; reopen is the only way back.
	run(t, "task", "status", "a", "completed", "--project", "p")
	_, err = runE(t, "task", "status", "a", "pending", "--project", "p")
	requireCode(t, err, "invalid_transition", 5)
}

func claimNext(t *testing.T, workflow string) (string, error) {
	t.Helper()
	out, err := runE(t, "task", "next", "--claim", "--workflow-id", workflow, "--project", "p", "--json")
	if err != nil {
		return "", err
	}
	var payload struct {
		Task struct {
			Name string `json:"name"`
		} `json:"task"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode next: %v\n%s", err, out)
	}
	return payload.Task.Name, nil
}

func TestTaskHold_BlocksDependentsUntilItPasses(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	chainAB(t)

	out, err := runE(t, "task", "status", "a", "completed", "--hold", "1s", "--project", "p", "--json")
	if err != nil {
		t.Fatalf("status --hold: %v", err)
	}
	var res struct {
		HoldUntil string `json:"hold_until"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil || res.HoldUntil == "" {
		t.Fatalf("status --hold JSON should carry hold_until: %v\n%s", err, out)
	}
	until, err := time.Parse(time.RFC3339, res.HoldUntil)
	if err != nil {
		t.Fatalf("hold_until: %v", err)
	}

	a := showTask(t, "a")
	if a.Status != "completed" || a.ComputedStatus != "completed" || a.HoldUntil == "" || a.CompletedAt == "" {
		t.Errorf("held a = %+v, want completed with hold_until and completed_at", a)
	}
	if text := run(t, "task", "show", "a", "--project", "p"); !strings.Contains(text, "counts from") {
		t.Errorf("show should note the hold:\n%s", text)
	}
	if b := showTask(t, "b"); b.ComputedStatus != "blocked" || len(b.PendingDeps) != 1 {
		t.Errorf("b during hold = %+v, want blocked on a", b)
	}
	if out := run(t, "task", "list", "--status", "ready", "--project", "p", "--json"); strings.Contains(out, `"name":"b"`) {
		t.Errorf("list --status ready shows b during the hold:\n%s", out)
	}
	_, err = runE(t, "task", "next", "--project", "p")
	requireCode(t, err, "not_found", 3)
	_, err = claimNext(t, "agent-1")
	requireCode(t, err, "not_found", 3)

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "p"))
	defer srv.Close()
	if b, _ := findWebTask(webGetState(t, srv.URL).Tasks, "p", "b"); b.ComputedStatus != "blocked" {
		t.Errorf("board shows b as %s during the hold, want blocked", b.ComputedStatus)
	}

	time.Sleep(time.Until(until) + 100*time.Millisecond)

	if b, _ := findWebTask(webGetState(t, srv.URL).Tasks, "p", "b"); b.ComputedStatus != "ready" {
		t.Errorf("board shows b as %s after the hold, want ready", b.ComputedStatus)
	}
	if got := showTask(t, "a").HoldUntil; got != "" {
		t.Errorf("hold_until = %q after the hold, want none", got)
	}
	name, err := claimNext(t, "agent-1")
	if err != nil || name != "b" {
		t.Fatalf("claim after hold = %q, %v; want b", name, err)
	}
	raw := readFile(t, filepath.Join(registryPath, "task", "p", "a.md"))
	if !strings.Contains(raw, "completion_counts_at:") {
		t.Errorf("a.md should store completion_counts_at:\n%s", raw)
	}
}

func TestTaskHold_ReopenDuringHoldSucceeds(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	run(t, "task", "status", "a", "completed", "--hold", "10m", "--project", "p")
	// Nobody can claim b, so nothing blocks the undo.
	_, err := claimNext(t, "agent-1")
	requireCode(t, err, "not_found", 3)
	if _, err := runE(t, "task", "reopen", "a", "--project", "p"); err != nil {
		t.Fatalf("reopen during hold: %v", err)
	}
	if a := showTask(t, "a"); a.Status != "pending" || a.HoldUntil != "" {
		t.Errorf("a = %+v, want pending without hold", a)
	}
	// Completing again without --hold releases b at once.
	run(t, "task", "status", "a", "completed", "--project", "p")
	if got := showTask(t, "b").ComputedStatus; got != "ready" {
		t.Errorf("b = %s, want ready", got)
	}
}

func TestTaskHold_Validation(t *testing.T) {
	setup(t)
	chdir(t)
	chainAB(t)
	_, err := runE(t, "task", "status", "a", "completed", "--hold", "61m", "--project", "p")
	requireCode(t, err, "invalid_input", 2)
	_, err = runE(t, "task", "status", "a", "completed", "--hold", "soon", "--project", "p")
	requireCode(t, err, "invalid_input", 2)
	_, err = runE(t, "task", "status", "a", "in_progress", "--hold", "1m", "--project", "p")
	requireCode(t, err, "invalid_input", 2)
	run(t, "task", "status", "a", "completed", "--project", "p")
	_, err = runE(t, "task", "status", "a", "completed", "--hold", "1m", "--project", "p")
	requireCode(t, err, "invalid_transition", 5)
}

func TestTaskTimestamps(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "x")
	a := showTask(t, "a")
	if _, err := time.Parse(time.RFC3339, a.CreatedAt); err != nil || !strings.HasSuffix(a.CreatedAt, "Z") {
		t.Errorf("created_at = %q, want RFC 3339 UTC", a.CreatedAt)
	}
	if a.UpdatedAt != a.CreatedAt || a.CompletedAt != "" {
		t.Errorf("new task = %+v, want updated_at = created_at and no completed_at", a)
	}
	run(t, "task", "status", "a", "completed", "--project", "p")
	if a = showTask(t, "a"); a.CompletedAt == "" || a.UpdatedAt == "" {
		t.Errorf("completed task = %+v, want completed_at and updated_at", a)
	}
	run(t, "task", "reopen", "a", "--project", "p")
	if a = showTask(t, "a"); a.CompletedAt != "" {
		t.Errorf("completed_at = %q after reopen, want cleared", a.CompletedAt)
	}

	// A task written before timestamps existed keeps working and gets no
	// invented created_at.
	old := "---\nsubject: Old\nstatus: pending\npriority: 2\ndepends_on: []\n---\n\nold task\n"
	if err := os.WriteFile(filepath.Join(registryPath, "task", "p", "old.md"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if o := showTask(t, "old"); o.CreatedAt != "" || o.UpdatedAt != "" {
		t.Errorf("old task = %+v, want no timestamps", o)
	}
	run(t, "task", "edit", "old", "-p", "1", "--project", "p")
	if o := showTask(t, "old"); o.CreatedAt != "" || o.UpdatedAt == "" {
		t.Errorf("edited old task = %+v, want updated_at only", o)
	}
}

func TestTaskList_SortByAge(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "young", "--project", "p", "-p", "0", "--content", "x")
	run(t, "task", "new", "elder", "--project", "p", "-p", "3", "--content", "x")
	run(t, "task", "new", "legacy", "--project", "p", "-p", "0", "--content", "x")
	setCreated := func(name, created string) {
		path := filepath.Join(registryPath, "task", "p", name+".md")
		raw := readFile(t, path)
		lines := []string{}
		for _, l := range strings.Split(raw, "\n") {
			if strings.HasPrefix(l, "created_at:") {
				if created == "" {
					continue
				}
				l = "created_at: \"" + created + "\""
			}
			lines = append(lines, l)
		}
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	setCreated("elder", "2026-01-01T00:00:00Z")
	setCreated("young", "2026-06-01T00:00:00Z")
	setCreated("legacy", "")

	names := func(args ...string) []string {
		out := run(t, append([]string{"task", "list", "--project", "p", "--json"}, args...)...)
		var payload struct {
			Tasks []struct {
				Name string `json:"name"`
			} `json:"tasks"`
		}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("decode: %v\n%s", err, out)
		}
		var got []string
		for _, it := range payload.Tasks {
			got = append(got, it.Name)
		}
		return got
	}
	if got := strings.Join(names("--sort", "age"), ","); got != "elder,young,legacy" {
		t.Errorf("--sort age = %s, want elder,young,legacy", got)
	}
	if got := names(); got[len(got)-1] != "elder" {
		t.Errorf("default order should stay by priority, got %v", got)
	}
	_, err := runE(t, "task", "list", "--project", "p", "--sort", "name")
	requireCode(t, err, "invalid_input", 2)
}
