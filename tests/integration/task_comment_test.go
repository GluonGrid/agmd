package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"agmd/cmd"
	"agmd/pkg/registry"
)

type commentItem struct {
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

func showComments(t *testing.T, project, task string) []commentItem {
	t.Helper()
	out, err := runE(t, "task", "show", task, "--project", project, "--json")
	if err != nil {
		t.Fatalf("task show: %v", err)
	}
	var item struct {
		Comments []commentItem `json:"comments"`
	}
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("decode show: %v\n%s", err, out)
	}
	return item.Comments
}

func TestTaskComment_AddAndShow(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")

	run(t, "task", "comment", "a", "started the schema", "--author", "agent-1", "--project", "p")
	t.Setenv("AGMD_AUTHOR", "agent-2")
	run(t, "task", "comment", "a", "--content", "reviewed", "--project", "p")

	comments := showComments(t, "p", "a")
	if len(comments) != 2 {
		t.Fatalf("got %d comments, want 2: %+v", len(comments), comments)
	}
	if comments[0].Author != "agent-1" || comments[0].Body != "started the schema" {
		t.Errorf("first comment = %+v", comments[0])
	}
	if comments[1].Author != "agent-2" || comments[1].CreatedAt == "" {
		t.Errorf("second comment = %+v", comments[1])
	}

	out := run(t, "task", "show", "a", "--project", "p")
	if !strings.Contains(out, "started the schema") || !strings.Contains(out, "comments (2)") {
		t.Errorf("text show is missing comments:\n%s", out)
	}
}

func TestTaskComment_EmptyBodyRejected(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")

	_, err := runE(t, "task", "comment", "a", "   ", "--project", "p")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected empty-body error, got %v", err)
	}
}

func TestTaskComment_DeletedWithTask(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")
	run(t, "task", "comment", "a", "note", "--project", "p")

	commentsDir := filepath.Join(registryPath, "task", "p", "comments")
	run(t, "task", "delete", "a", "--project", "p", "--force")
	entries, _ := filepath.Glob(filepath.Join(commentsDir, "*.json"))
	if len(entries) != 0 {
		t.Errorf("comments left behind after delete: %v", entries)
	}
}

func TestTaskComment_ParallelCommentsAreNotLost(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")

	const n = 12
	results := runParallel(t, n, func(i int) []string {
		return []string{"task", "comment", "a", fmt.Sprintf("note %d", i), "--author", fmt.Sprintf("agent-%d", i), "--project", "p"}
	})
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("comment %d failed: %v", i, r.err)
		}
	}
	if got := len(showComments(t, "p", "a")); got != n {
		t.Errorf("got %d comments, want %d (lost update)", got, n)
	}
}

func TestTaskWeb_CommentEndpoint(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	status, data := webPost(t, srv.URL+"/api/tasks/proj-a/a/comments", map[string]interface{}{"body": "from the board"})
	if status != http.StatusOK {
		t.Fatalf("POST comment = %d, body=%s", status, data)
	}
	var item struct {
		Comments []commentItem `json:"comments"`
	}
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(item.Comments) != 1 || item.Comments[0].Author != "user" {
		t.Errorf("response comments = %+v", item.Comments)
	}

	status, _ = webPost(t, srv.URL+"/api/tasks/proj-a/a/comments", map[string]interface{}{"body": ""})
	if status != http.StatusBadRequest {
		t.Errorf("empty comment status = %d, want 400", status)
	}

	if got := showComments(t, "proj-a", "a"); len(got) != 1 || got[0].Body != "from the board" {
		t.Errorf("CLI sees comments %+v", got)
	}
}
