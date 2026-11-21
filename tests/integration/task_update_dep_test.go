package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type shownTask struct {
	ID         string   `json:"id"`
	Subject    string   `json:"subject"`
	Status     string   `json:"status"`
	Priority   int      `json:"priority"`
	Type       string   `json:"type"`
	Feature    string   `json:"feature"`
	Human      bool     `json:"human"`
	Check      string   `json:"check"`
	DependsOn  []string `json:"depends_on"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	Content    string   `json:"content"`
	Assignment *struct {
		WorkflowID string `json:"workflow_id"`
	} `json:"assignment"`
	Comments []struct {
		Body string `json:"body"`
	} `json:"comments"`
}

func showFullTask(t *testing.T, name string) shownTask {
	t.Helper()
	out, err := runE(t, "task", "show", name, "--project", "p", "--json")
	if err != nil {
		t.Fatalf("show %s: %v", name, err)
	}
	var task shownTask
	if err := json.Unmarshal([]byte(out), &task); err != nil {
		t.Fatalf("show %s output is not JSON: %v\n%s", name, err, out)
	}
	return task
}

// withStdin runs fn with os.Stdin reading the given text.
func withStdin(t *testing.T, text string, fn func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	old := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = old }()
	fn()
}

var timestampLine = regexp.MustCompile(`(?m)^(created_at|updated_at): .*$`)

// ageTimestamps rewrites a task's created_at/updated_at to a fixed old time.
func ageTimestamps(t *testing.T, registry, name string) {
	t.Helper()
	path := filepath.Join(registry, "task", "p", name+".md")
	data := timestampLine.ReplaceAllString(readFile(t, path), `$1: "2020-01-01T00:00:00Z"`)
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTaskUpdate_EachFieldKeepsIdentityCommentsDepsAndClaim(t *testing.T) {
	registry := setup(t)
	run(t, "task", "new", "dep", "--project", "p", "--content", "Dep")
	run(t, "task", "new", "work", "--project", "p", "--blocked-by", "dep", "--content", "Old text")
	run(t, "task", "comment", "work", "keep me", "--project", "p", "--author", "x")
	run(t, "task", "assign", "work", "--project", "p", "--workflow-id", "wf-1")
	ageTimestamps(t, registry, "work")
	before := showFullTask(t, "work")

	steps := []struct {
		args  []string
		check func(shownTask) bool
	}{
		{[]string{"--subject", "New subject"}, func(s shownTask) bool { return s.Subject == "New subject" }},
		{[]string{"--content", "New text"}, func(s shownTask) bool { return s.Content == "New text" }},
		{[]string{"-p", "0"}, func(s shownTask) bool { return s.Priority == 0 }},
		{[]string{"-t", "bug"}, func(s shownTask) bool { return s.Type == "bug" }},
		{[]string{"--feature", "auth"}, func(s shownTask) bool { return s.Feature == "auth" }},
		{[]string{"--human"}, func(s shownTask) bool { return s.Human }},
		{[]string{"--check", "true"}, func(s shownTask) bool { return s.Check == "true" }},
		{[]string{"--check", "", "--human=false", "--feature", ""}, func(s shownTask) bool { return !s.Human && s.Check == "" && s.Feature == "" }},
	}
	for _, step := range steps {
		args := append([]string{"task", "update", "work", "--project", "p", "--json"}, step.args...)
		out, err := runE(t, args...)
		if err != nil {
			t.Fatalf("update %v: %v", step.args, err)
		}
		if !strings.Contains(out, `"action":"update"`) || !strings.Contains(out, `"ok":true`) {
			t.Errorf("update %v JSON = %s", step.args, out)
		}
		if got := showFullTask(t, "work"); !step.check(got) {
			t.Errorf("update %v not applied: %+v", step.args, got)
		}
	}

	after := showFullTask(t, "work")
	if after.ID != before.ID {
		t.Errorf("id changed: %s → %s", before.ID, after.ID)
	}
	if len(after.DependsOn) != 1 || after.DependsOn[0] != "dep" {
		t.Errorf("depends_on = %v, want [dep]", after.DependsOn)
	}
	if len(after.Comments) != 1 || after.Comments[0].Body != "keep me" {
		t.Errorf("comments = %+v", after.Comments)
	}
	if after.Assignment == nil || after.Assignment.WorkflowID != "wf-1" {
		t.Errorf("claim lost: %+v", after.Assignment)
	}
	if after.CreatedAt != "2020-01-01T00:00:00Z" {
		t.Errorf("created_at changed to %s", after.CreatedAt)
	}
	if after.UpdatedAt == "2020-01-01T00:00:00Z" || after.UpdatedAt == "" {
		t.Errorf("updated_at not stamped: %s", after.UpdatedAt)
	}
}

func TestTaskUpdate_ContentFromFileAndStdin(t *testing.T) {
	setup(t)
	run(t, "task", "new", "work", "--project", "p", "--content", "Old")

	file := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(file, []byte("# From file\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runE(t, "task", "update", "work", "--project", "p", "--content-file", file); err != nil {
		t.Fatal(err)
	}
	if got := showFullTask(t, "work").Content; got != "# From file\n\nbody" {
		t.Errorf("content from file = %q", got)
	}

	withStdin(t, "From piped stdin\n", func() {
		if _, err := runE(t, "task", "update", "work", "--project", "p"); err != nil {
			t.Fatal(err)
		}
	})
	if got := showFullTask(t, "work").Content; got != "From piped stdin" {
		t.Errorf("content from stdin = %q", got)
	}

	withStdin(t, "From content-file -", func() {
		if _, err := runE(t, "task", "update", "work", "--project", "p", "--content-file", "-", "-p", "1"); err != nil {
			t.Fatal(err)
		}
	})
	if got := showFullTask(t, "work"); got.Content != "From content-file -" || got.Priority != 1 {
		t.Errorf("content-file - = %q, priority %d", got.Content, got.Priority)
	}

	// Stdin is only read when nothing else is changed.
	withStdin(t, "must not be read", func() {
		run(t, "task", "update", "work", "--project", "p", "--subject", "S")
	})
	if got := showFullTask(t, "work").Content; got != "From content-file -" {
		t.Errorf("stdin read although --subject was given: %q", got)
	}
}

func TestTaskUpdate_RefusesInvalidInput(t *testing.T) {
	setup(t)
	run(t, "task", "new", "work", "--project", "p", "--content", "Old")
	cases := [][]string{
		{},                  // nothing to change (stdin empty)
		{"-p", "7"},         // bad priority
		{"-t", "epic"},      // bad type
		{"--subject", " "},  // empty subject
		{"--check", "true"}, // check on a non-human task
		{"--content", "a", "--content-file", "x"},
		{"--content-file", "/no/such/file"},
	}
	for _, extra := range cases {
		args := append([]string{"task", "update", "work", "--project", "p", "--json"}, extra...)
		var err error
		withStdin(t, "", func() { _, err = runE(t, args...) })
		if exitCode(err) != 2 {
			t.Errorf("update %v: err = %v (exit %d), want invalid_input", extra, err, exitCode(err))
		}
	}
	if _, err := runE(t, "task", "update", "nope", "--project", "p", "--subject", "x"); exitCode(err) != 3 {
		t.Errorf("missing task: exit %d, want 3", exitCode(err))
	}
	if got := showFullTask(t, "work"); got.Content != "Old" || got.Priority != 2 {
		t.Errorf("refused update changed the task: %+v", got)
	}
}

func TestTaskUpdate_CompletedNeedsForce(t *testing.T) {
	setup(t)
	run(t, "task", "new", "done", "--project", "p", "--content", "Old")
	run(t, "task", "status", "done", "completed", "--project", "p")

	_, err := runE(t, "task", "update", "done", "--project", "p", "--content", "New")
	if exitCode(err) != 5 {
		t.Fatalf("update of completed task: err %v (exit %d), want invalid_transition (5)", err, exitCode(err))
	}
	if got := showFullTask(t, "done").Content; got != "Old" {
		t.Errorf("refused update changed content: %q", got)
	}
	if _, err := runE(t, "task", "update", "done", "--project", "p", "--content", "New", "--force"); err != nil {
		t.Fatal(err)
	}
	got := showFullTask(t, "done")
	if got.Content != "New" || got.Status != "completed" {
		t.Errorf("forced update: %+v", got)
	}
}

func TestTaskEdit_AliasStillWorks(t *testing.T) {
	setup(t)
	run(t, "task", "new", "work", "--project", "p", "--content", "Old")
	run(t, "task", "status", "work", "completed", "--project", "p")
	// edit keeps its old behaviour: completed tasks without --force, "edit" action.
	out, err := runE(t, "task", "edit", "work", "--project", "p", "-p", "1", "--feature", "f", "--human", "--check", "true", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"action":"edit"`) {
		t.Errorf("edit JSON = %s", out)
	}
	got := showFullTask(t, "work")
	if got.Priority != 1 || got.Feature != "f" || !got.Human || got.Check != "true" {
		t.Errorf("edit not applied: %+v", got)
	}
}

func TestTaskDep_AddRemoveAndAliases(t *testing.T) {
	setup(t)
	for _, name := range []string{"a", "b", "c"} {
		run(t, "task", "new", name, "--project", "p", "--content", name)
	}

	out, err := runE(t, "task", "dep", "add", "b", "a", "--project", "p", "--json")
	if err != nil || !strings.Contains(out, `"action":"dep-add"`) {
		t.Fatalf("dep add: %v %s", err, out)
	}
	out, err = runE(t, "task", "blocked-by", "c", "b", "--project", "p", "--json")
	if err != nil || !strings.Contains(out, `"action":"blocked-by"`) {
		t.Fatalf("blocked-by: %v %s", err, out)
	}
	if deps := showFullTask(t, "c").DependsOn; len(deps) != 1 || deps[0] != "b" {
		t.Errorf("c depends_on = %v", deps)
	}

	// Refusals: duplicate, self, cycle (direct and transitive), missing.
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"b", "a"}, 2},
		{[]string{"a", "a"}, 2},
		{[]string{"a", "b"}, 6},
		{[]string{"a", "c"}, 6},
		{[]string{"a", "nope"}, 3},
	} {
		args := append([]string{"task", "dep", "add"}, tc.args...)
		_, err := runE(t, append(args, "--project", "p", "--json")...)
		if exitCode(err) != tc.code {
			t.Errorf("dep add %v: exit %d, want %d (%v)", tc.args, exitCode(err), tc.code, err)
		}
	}
	if deps := showFullTask(t, "a").DependsOn; len(deps) != 0 {
		t.Errorf("refused links were added: %v", deps)
	}

	out, err = runE(t, "task", "dep", "remove", "b", "a", "--project", "p", "--json")
	if err != nil || !strings.Contains(out, `"action":"dep-remove"`) {
		t.Fatalf("dep remove: %v %s", err, out)
	}
	if _, err := runE(t, "task", "dep", "rm", "b", "a", "--project", "p"); exitCode(err) != 3 {
		t.Errorf("removing a missing link: exit %d, want 3", exitCode(err))
	}
	out, err = runE(t, "task", "unblock", "c", "b", "--project", "p", "--json")
	if err != nil || !strings.Contains(out, `"action":"unblock"`) {
		t.Fatalf("unblock: %v %s", err, out)
	}
	if deps := showFullTask(t, "c").DependsOn; len(deps) != 0 {
		t.Errorf("c depends_on after unblock = %v", deps)
	}
}

func TestTaskNew_DependsOnSynonym(t *testing.T) {
	setup(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "a")
	run(t, "task", "new", "b", "--project", "p", "--content", "b")
	run(t, "task", "new", "c", "--project", "p", "--content", "c", "--blocked-by", "a", "--depends-on", "b")
	if deps := showFullTask(t, "c").DependsOn; fmt.Sprint(deps) != "[a b]" {
		t.Errorf("c depends_on = %v, want [a b]", deps)
	}
}

// exitCode returns the exit status a coded task error maps to (1 otherwise, 0 for nil).
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if coded, ok := err.(interface{ ExitStatus() int }); ok {
		return coded.ExitStatus()
	}
	return 1
}
