package integration

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

type movedTask struct {
	ID          string `json:"id"`
	Project     string `json:"project"`
	Status      string `json:"status"`
	Human       bool   `json:"human"`
	Check       string `json:"check"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at"`
	Assignment  *struct {
		WorkflowID string `json:"workflow_id"`
	} `json:"assignment"`
	Comments []struct {
		Author string `json:"author"`
		Body   string `json:"body"`
	} `json:"comments"`
}

func showIn(t *testing.T, project, name string) movedTask {
	t.Helper()
	out, err := runE(t, "task", "show", name, "--project", project, "--json")
	if err != nil {
		t.Fatalf("show %s/%s: %v", project, name, err)
	}
	var task movedTask
	if err := json.Unmarshal([]byte(out), &task); err != nil {
		t.Fatalf("show output is not JSON: %v\n%s", err, out)
	}
	return task
}

func TestTaskMove_KeepsIdentityTextCommentsAndFlags(t *testing.T) {
	registry := setup(t)
	chdir(t)
	run(t, "task", "new", "merge-pr", "--project", "from", "--human", "--check", "true", "--content", "Merge the PR")
	run(t, "task", "comment", "merge-pr", "first note", "--project", "from", "--author", "a1")
	before := showIn(t, "from", "merge-pr")

	out, err := runE(t, "task", "move", "merge-pr", "--project", "from", "--to-project", "to", "--author", "mover", "--json")
	if err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	var payload struct {
		OK          bool   `json:"ok"`
		Action      string `json:"action"`
		Project     string `json:"project"`
		FromProject string `json:"from_project"`
		Task        string `json:"task"`
		ID          string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("move output is not JSON: %v\n%s", err, out)
	}
	if !payload.OK || payload.Action != "move" || payload.Project != "to" || payload.FromProject != "from" || payload.Task != "merge-pr" || payload.ID != before.ID {
		t.Errorf("payload = %+v", payload)
	}

	after := showIn(t, "to", "merge-pr")
	if after.ID != before.ID || after.Content != "Merge the PR" || !after.Human || after.Check != "true" || after.CreatedAt != before.CreatedAt {
		t.Errorf("moved task changed: before %+v, after %+v", before, after)
	}
	if len(after.Comments) != 2 || after.Comments[0].Body != "first note" || after.Comments[1].Body != "Moved from project 'from'." || after.Comments[1].Author != "mover" {
		t.Errorf("comments = %+v", after.Comments)
	}
	if fileExists(filepath.Join(registry, "task", "from", "merge-pr.md")) {
		t.Error("task file still in the source project")
	}
	if fileExists(filepath.Join(registry, "task", "from", "comments", before.ID+".json")) {
		t.Error("comments still in the source project")
	}
	if _, err := runE(t, "task", "show", "merge-pr", "--project", "from"); err == nil {
		t.Error("task still found in the source project")
	}
	// The project/task form works as well, and a completed task keeps its completion time.
	run(t, "task", "status", "merge-pr", "completed", "--project", "to")
	done := showIn(t, "to", "merge-pr")
	run(t, "task", "move", "to/merge-pr", "--to-project", "third")
	if moved := showIn(t, "third", "merge-pr"); moved.Status != "completed" || moved.CompletedAt != done.CompletedAt {
		t.Errorf("completed task after move = %+v", moved)
	}
}

func TestTaskMove_Refusals(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "base", "--project", "from", "--content", "B")
	run(t, "task", "new", "top", "--project", "from", "--blocked-by", "base", "--content", "T")
	run(t, "task", "new", "taken", "--project", "to", "--content", "exists there")
	run(t, "task", "new", "taken", "--project", "from", "--content", "and here")

	_, err := runE(t, "task", "move", "top", "--project", "from", "--to-project", "to")
	requireCode(t, err, "conflict", 6)
	if err == nil || !strings.Contains(err.Error(), "depends on base") {
		t.Errorf("dependency refusal should name it: %v", err)
	}
	_, err = runE(t, "task", "move", "base", "--project", "from", "--to-project", "to")
	requireCode(t, err, "conflict", 6)
	if !strings.Contains(err.Error(), "top depend on it") {
		t.Errorf("dependent refusal should name it: %v", err)
	}
	_, err = runE(t, "task", "move", "taken", "--project", "from", "--to-project", "to")
	requireCode(t, err, "conflict", 6)
	_, err = runE(t, "task", "move", "taken", "--project", "from", "--to-project", "from")
	requireCode(t, err, "invalid_input", 2)
	_, err = runE(t, "task", "move", "taken", "--project", "from")
	requireCode(t, err, "invalid_input", 2)
	_, err = runE(t, "task", "move", "missing", "--project", "from", "--to-project", "to")
	requireCode(t, err, "not_found", 3)

	// Nothing was moved by the refusals.
	for _, name := range []string{"base", "top", "taken"} {
		showIn(t, "from", name)
	}
	if got := showIn(t, "to", "taken"); got.Content != "exists there" {
		t.Errorf("target task overwritten: %+v", got)
	}
	// Once the dependency is removed, both can move.
	run(t, "task", "dep", "remove", "top", "base", "--project", "from")
	if _, err := runE(t, "task", "move", "top", "--project", "from", "--to-project", "to"); err != nil {
		t.Errorf("move after dep remove: %v", err)
	}
}

func TestTaskMove_ClaimByAnotherWorkflowRefusedOwnClaimMoves(t *testing.T) {
	registry := setup(t)
	chdir(t)
	run(t, "task", "new", "job", "--project", "from", "--content", "J")
	run(t, "task", "assign", "job", "--project", "from", "--workflow-id", "wf-1")

	_, err := runE(t, "task", "move", "job", "--project", "from", "--to-project", "to")
	requireCode(t, err, "conflict", 6)
	_, err = runE(t, "task", "move", "job", "--project", "from", "--to-project", "to", "--workflow-id", "wf-2")
	requireCode(t, err, "conflict", 6)

	if _, err := runE(t, "task", "move", "job", "--project", "from", "--to-project", "to", "--workflow-id", "wf-1"); err != nil {
		t.Fatalf("move by the claim holder: %v", err)
	}
	moved := showIn(t, "to", "job")
	if moved.Assignment == nil || moved.Assignment.WorkflowID != "wf-1" {
		t.Errorf("claim did not move along: %+v", moved)
	}
	if a := readAssignments(t, registry, "from"); len(a) != 0 {
		t.Errorf("claim left in the source project: %v", a)
	}
}

func TestTaskCommentDelete(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "c", "--project", "p", "--content", "C")
	run(t, "task", "comment", "c", "one", "--project", "p", "--author", "a1")
	run(t, "task", "comment", "c", "two with a handle", "--project", "p", "--author", "a2")
	run(t, "task", "comment", "c", "three", "--project", "p", "--author", "a1")

	if out := run(t, "task", "show", "c", "--project", "p"); !strings.Contains(out, "#2") {
		t.Errorf("show should number comments:\n%s", out)
	}
	_, err := runE(t, "task", "comment", "c", "--delete", "2", "--project", "p", "--author", "a1")
	requireCode(t, err, "conflict", 6)
	_, err = runE(t, "task", "comment", "c", "--delete", "4", "--project", "p", "--author", "a1")
	requireCode(t, err, "not_found", 3)
	_, err = runE(t, "task", "comment", "c", "body", "--delete", "1", "--project", "p", "--author", "a1")
	requireCode(t, err, "invalid_input", 2)

	out, err := runE(t, "task", "comment", "c", "--delete", "2", "--project", "p", "--author", "a2", "--json")
	if err != nil {
		t.Fatalf("delete own comment: %v", err)
	}
	if !strings.Contains(out, `"action":"comment_delete"`) || !strings.Contains(out, `"two with a handle"`) {
		t.Errorf("delete JSON = %s", out)
	}
	if _, err := runE(t, "task", "comment", "c", "--delete", "1", "--project", "p", "--author", "someone", "--force"); err != nil {
		t.Fatalf("forced delete: %v", err)
	}
	got := showIn(t, "p", "c").Comments
	if len(got) != 1 || got[0].Body != "three" {
		t.Errorf("comments left = %+v", got)
	}
	run(t, "task", "comment", "c", "--delete", "1", "--project", "p", "--author", "a1")
	if got := showIn(t, "p", "c").Comments; len(got) != 0 {
		t.Errorf("comments left = %+v", got)
	}
}
