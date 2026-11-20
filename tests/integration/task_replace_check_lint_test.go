package integration

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

type showPItem struct {
	Status     string   `json:"status"`
	Human      bool     `json:"human"`
	Check      string   `json:"check"`
	ReplacedBy string   `json:"replaced_by"`
	Replaces   []string `json:"replaces"`
	DependsOn  []string `json:"depends_on"`
	Comments   []struct {
		Author string `json:"author"`
		Body   string `json:"body"`
	} `json:"comments"`
}

func showP(t *testing.T, name string) showPItem {
	t.Helper()
	out := run(t, "task", "show", name, "--project", "p", "--json")
	var item showPItem
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("decode show: %v\n%s", err, out)
	}
	return item
}

func TestTaskReplace_ClosesLinksAndMovesDependents(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "human-old", "--project", "p", "--human", "--content", "Bundle")
	run(t, "task", "new", "human-new", "--project", "p", "--human", "--content", "One action")
	run(t, "task", "new", "deploy", "--project", "p", "--blocked-by", "human-old", "--content", "Deploy")

	if _, err := runE(t, "task", "replace", "human-old", "--by", "human-new", "--project", "p", "--author", "me"); err != nil {
		t.Fatalf("replace: %v", err)
	}

	old := showP(t, "human-old")
	if old.Status != "completed" || old.ReplacedBy != "human-new" {
		t.Errorf("old task = %+v, want completed and replaced_by human-new", old)
	}
	if len(old.Comments) != 1 || !strings.Contains(old.Comments[0].Body, "Replaced by human-new") || old.Comments[0].Author != "me" {
		t.Errorf("old comments = %+v", old.Comments)
	}
	nw := showP(t, "human-new")
	if len(nw.Replaces) != 1 || nw.Replaces[0] != "human-old" || nw.Status != "pending" {
		t.Errorf("new task = %+v, want replaces [human-old]", nw)
	}
	dep := showP(t, "deploy")
	if len(dep.DependsOn) != 1 || dep.DependsOn[0] != "human-new" {
		t.Errorf("deploy depends_on = %v, want [human-new]", dep.DependsOn)
	}
	// deploy stays blocked: the replacement is not done.
	if out := run(t, "task", "list", "--project", "p", "--status", "blocked"); !strings.Contains(out, "deploy") {
		t.Errorf("deploy should stay blocked:\n%s", out)
	}

	if out := run(t, "task", "show", "human-old", "--project", "p"); !strings.Contains(out, "replaced by: human-new") {
		t.Errorf("show should print replaced by:\n%s", out)
	}
	if out := run(t, "task", "list", "--project", "p", "--all"); !strings.Contains(out, "replaced by human-new") {
		t.Errorf("list should print replaced by:\n%s", out)
	}
}

func TestTaskReplace_Refusals(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "A")
	run(t, "task", "new", "b", "--project", "p", "--content", "B")
	run(t, "task", "new", "c", "--project", "p", "--blocked-by", "a", "--content", "C")
	run(t, "task", "blocked-by", "b", "c", "--project", "p")

	_, err := runE(t, "task", "replace", "a", "--project", "p")
	requireCode(t, err, "invalid_input", 2)
	_, err = runE(t, "task", "replace", "a", "--by", "a", "--project", "p")
	requireCode(t, err, "invalid_input", 2)
	// b depends on c, c depends on a: moving c to b would be a cycle.
	_, err = runE(t, "task", "replace", "a", "--by", "b", "--project", "p")
	requireCode(t, err, "conflict", 6)
	if showP(t, "a").Status != "pending" {
		t.Error("refused replace must not change the old task")
	}

	run(t, "task", "status", "b", "completed", "--project", "p")
	_, err = runE(t, "task", "replace", "b", "--by", "a", "--project", "p")
	requireCode(t, err, "invalid_transition", 5)
}

func TestTaskReplace_ReleasesClaim(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "A")
	run(t, "task", "new", "b", "--project", "p", "--content", "B")
	run(t, "task", "assign", "a", "--project", "p", "--workflow-id", "w1")
	run(t, "task", "replace", "a", "--by", "b", "--project", "p")
	out := run(t, "task", "show", "a", "--project", "p", "--json")
	if strings.Contains(out, `"assignment"`) {
		t.Errorf("claim on replaced task should be released: %s", out)
	}
}

func TestTaskCheck_NewEditShowAndValidation(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	_, err := runE(t, "task", "new", "x", "--project", "p", "--check", "true", "--content", "X")
	requireCode(t, err, "invalid_input", 2)

	run(t, "task", "new", "merge", "--project", "p", "--human", "--check", "test -f merged", "--content", "Merge")
	if raw := readFile(t, filepath.Join(registryPath, "task", "p", "merge.md")); !strings.Contains(raw, "check: test -f merged") {
		t.Errorf("task file lacks check:\n%s", raw)
	}
	if out := run(t, "task", "show", "merge", "--project", "p"); !strings.Contains(out, "check: test -f merged") {
		t.Errorf("show should print the check:\n%s", out)
	}
	run(t, "task", "edit", "merge", "--project", "p", "--check", "true")
	if got := showP(t, "merge").Check; got != "true" {
		t.Errorf("check after edit = %q", got)
	}
	// Turning it into an agent task while it has a check is refused.
	_, err = runE(t, "task", "edit", "merge", "--project", "p", "--human=false")
	requireCode(t, err, "invalid_input", 2)
	run(t, "task", "edit", "merge", "--project", "p", "--check", "")
	if got := showP(t, "merge").Check; got != "" {
		t.Errorf("check after clear = %q", got)
	}

	run(t, "task", "new", "agent", "--project", "p", "--content", "A")
	_, err = runE(t, "task", "edit", "agent", "--project", "p", "--check", "true")
	requireCode(t, err, "invalid_input", 2)
}

func TestTaskCheck_CompletesPassingHumanTasks(t *testing.T) {
	registryPath := setup(t)
	dir := chdir(t)
	run(t, "task", "new", "pass", "--project", "p", "--human", "--check", "echo MR merged; echo second", "--content", "P")
	run(t, "task", "new", "fail", "--project", "p", "--human", "--check", "echo open; exit 1", "--content", "F")
	run(t, "task", "new", "slow", "--project", "p", "--human", "--check", "sleep 5", "--content", "S")
	run(t, "task", "new", "env", "--project", "p", "--human", "--check", `test "$AGMD_TASK" = env && test "$AGMD_PROJECT" = p`, "--content", "E")
	run(t, "task", "new", "deploy", "--project", "p", "--blocked-by", "pass", "--content", "D")
	// An agent task with a check written by hand must never be run.
	marker := filepath.Join(dir, "ran")
	run(t, "task", "new", "sneaky", "--project", "p", "--content", "S")
	sneaky := filepath.Join(registryPath, "task", "p", "sneaky.md")
	raw := readFile(t, sneaky)
	writeFile(t, sneaky, strings.Replace(raw, "depends_on:", "check: touch "+marker+"\ndepends_on:", 1))

	out, err := runE(t, "task", "check", "--project", "p", "--timeout", "500ms")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	lines := strings.Fields(strings.TrimSpace(out))
	if strings.Join(lines, " ") != "p/pass p/env" && strings.Join(lines, " ") != "p/env p/pass" {
		t.Errorf("check output = %q, want p/env and p/pass only", out)
	}
	if fileExists(marker) {
		t.Error("check ran for a task not flagged human")
	}

	pass := showP(t, "pass")
	if pass.Status != "completed" {
		t.Errorf("pass status = %s", pass.Status)
	}
	if len(pass.Comments) != 1 || pass.Comments[0].Body != "verified automatically: MR merged" {
		t.Errorf("pass comments = %+v", pass.Comments)
	}
	for _, name := range []string{"fail", "slow", "sneaky"} {
		if s := showP(t, name).Status; s != "pending" {
			t.Errorf("%s status = %s, want pending", name, s)
		}
	}
	if out := run(t, "task", "list", "--project", "p", "--status", "ready"); !strings.Contains(out, "deploy") {
		t.Errorf("deploy should be released:\n%s", out)
	}

	// A second run finds nothing new; JSON lists the remaining failures.
	out = run(t, "task", "check", "--project", "p", "--timeout", "500ms", "--json")
	var payload struct {
		Completed []struct{ Task string } `json:"completed"`
		NotDone   []struct {
			Task     string `json:"task"`
			TimedOut bool   `json:"timed_out"`
		} `json:"not_done"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode: %v\n%s", err, out)
	}
	if len(payload.Completed) != 0 || len(payload.NotDone) != 2 {
		t.Errorf("second run = %+v", payload)
	}
	for _, r := range payload.NotDone {
		if r.Task == "slow" && !r.TimedOut {
			t.Error("slow check should time out")
		}
	}
}

func TestTaskLint_FindsUnflaggedAndReplacedTasks(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "human-legacy", "--project", "p", "--content", "Old style")
	run(t, "task", "new", "human-ok", "--project", "p", "--human", "--content", "New style")
	run(t, "task", "new", "renamed", "--project", "p", "--content", "R")
	run(t, "task", "comment", "renamed", "Replaced by better-name", "--project", "p")
	run(t, "task", "new", "closed", "--project", "p", "--content", "C")
	run(t, "task", "comment", "closed", "superseded by x", "--project", "p")
	run(t, "task", "status", "closed", "completed", "--project", "p")
	run(t, "task", "new", "a", "--project", "p", "--content", "A")
	run(t, "task", "new", "b", "--project", "p", "--content", "B")
	run(t, "task", "replace", "a", "--by", "b", "--project", "p")

	out := run(t, "task", "lint", "--project", "p")
	want := []string{"p/human-legacy human-name-not-flagged", "p/renamed replaced-but-open"}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("lint output lacks %q:\n%s", w, out)
		}
	}
	for _, unwanted := range []string{"human-ok", "p/closed", "p/a ", "p/b "} {
		if strings.Contains(out, unwanted) {
			t.Errorf("lint output should not mention %q:\n%s", unwanted, out)
		}
	}
	if n := len(strings.Split(strings.TrimSpace(out), "\n")); n != 2 {
		t.Errorf("lint found %d lines, want 2:\n%s", n, out)
	}
}
