package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func TestTask_List(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "task-alpha", "--content", "Alpha.")
	run(t, "task", "new", "task-beta", "--content", "Beta.")

	out := run(t, "task", "list")
	if !strings.Contains(out, "task-alpha") {
		t.Error("task list should show task-alpha")
	}
	if !strings.Contains(out, "task-beta") {
		t.Error("task list should show task-beta")
	}
}

func TestTask_Status(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "my-task", "--content", "Do something.")
	run(t, "task", "status", "my-task", "completed")

	out := run(t, "task", "list", "--all")
	if !strings.Contains(out, "completed") {
		t.Error("task should show as completed")
	}
}

func TestTask_Show(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "show-task", "--content", "Show this content.")

	out := run(t, "task", "show", "show-task")
	if !strings.Contains(out, "Show this content.") {
		t.Error("task show should display task content")
	}
}

func TestTask_New(t *testing.T) {
	setup(t)
	chdir(t)

	out := run(t, "task", "new", "my-task", "--content", "Do the thing.")
	if !strings.Contains(out, "my-task") {
		t.Error("expected task name in output")
	}
}

func TestTask_NewWithTypeAndPriority(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "critical-bug", "-t", "bug", "-p", "0", "--content", "Critical bug.")

	out := run(t, "task", "list")
	if !strings.Contains(out, "critical-bug") {
		t.Error("task list should show critical-bug")
	}
	if !strings.Contains(out, "P0") {
		t.Error("task list should show P0 priority")
	}
}

func TestTask_Delete(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "del-task", "--content", "To be deleted.")
	run(t, "task", "delete", "del-task", "--force")

	out := run(t, "task", "list")
	if strings.Contains(out, "del-task") {
		t.Error("deleted task should not appear in list")
	}
}

func TestTask_Clean(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "done-task", "--content", "Completed.")
	run(t, "task", "status", "done-task", "completed")
	run(t, "task", "new", "active-task", "--content", "Still active.")

	run(t, "task", "clean", "--force")

	out := run(t, "task", "list", "--all")
	if strings.Contains(out, "done-task") {
		t.Error("clean should remove completed tasks")
	}
	if !strings.Contains(out, "active-task") {
		t.Error("clean should keep active tasks")
	}
}

func TestTask_Edit(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "my-task", "--content", "Do this.")
	run(t, "task", "edit", "my-task", "-p", "0", "-t", "bug")

	out := run(t, "task", "list")
	if !strings.Contains(out, "P0") {
		t.Error("task list should show P0 after edit")
	}
	if !strings.Contains(out, "bug") {
		t.Error("task list should show bug type after edit")
	}
}

func TestTask_ListStatusFilter(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "ready-task", "--content", "Ready.")
	run(t, "task", "new", "done-task", "--content", "Done.")
	run(t, "task", "status", "done-task", "completed")

	outReady := run(t, "task", "list", "--status", "ready")
	if !strings.Contains(outReady, "ready-task") {
		t.Error("--status ready should show ready-task")
	}
	if strings.Contains(outReady, "done-task") {
		t.Error("--status ready should not show done-task")
	}
}

func TestTask_BlockedByAndUnblock(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "dep-task", "--content", "Dependency.")
	run(t, "task", "new", "main-task", "--content", "Needs dep.", "--blocked-by", "dep-task")

	out := run(t, "task", "list")
	if !strings.Contains(out, "blocked") {
		t.Errorf("main-task should show as blocked, got:\n%s", out)
	}

	run(t, "task", "unblock", "main-task", "dep-task")

	out = run(t, "task", "list")
	if strings.Contains(out, "blocked") {
		t.Errorf("main-task should not be blocked after unblock, got:\n%s", out)
	}
}

func TestTask_ListFilters(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "bug-one", "-t", "bug", "-p", "0", "--content", "Bug.")
	run(t, "task", "new", "feat-one", "-t", "feature", "-p", "2", "--content", "Feature.")

	outBug := run(t, "task", "list", "--type", "bug")
	if !strings.Contains(outBug, "bug-one") {
		t.Error("--type bug should show bug-one")
	}
	if strings.Contains(outBug, "feat-one") {
		t.Error("--type bug should not show feat-one")
	}

	outP0 := run(t, "task", "list", "--priority", "0")
	if !strings.Contains(outP0, "bug-one") {
		t.Error("--priority 0 should show bug-one")
	}
	if strings.Contains(outP0, "feat-one") {
		t.Error("--priority 0 should not show feat-one (P2)")
	}
}

func TestTask_ShowAll(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "show-a", "--content", "Content A.")
	run(t, "task", "new", "show-b", "--content", "Content B.")

	out := run(t, "task", "show", "--all")
	if !strings.Contains(out, "Content A.") {
		t.Error("show --all should include Content A.")
	}
	if !strings.Contains(out, "Content B.") {
		t.Error("show --all should include Content B.")
	}
}

func TestTask_WorktreeUsesCanonicalProject(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	registry := setup(t)
	_ = registry

	baseDir := t.TempDir()
	mainRepo := filepath.Join(baseDir, "canonical-project")
	worktreeRepo := filepath.Join(baseDir, "feature-worktree")
	if err := os.MkdirAll(mainRepo, 0o755); err != nil {
		t.Fatalf("mkdir main repo: %v", err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(orig)

	if err := os.Chdir(mainRepo); err != nil {
		t.Fatalf("chdir main repo: %v", err)
	}
	runGit(t, mainRepo, "init")
	runGit(t, mainRepo, "config", "user.email", "test@example.com")
	runGit(t, mainRepo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("seed"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "init")
	runGit(t, mainRepo, "worktree", "add", "-b", "feature-x", worktreeRepo)

	if err := os.Chdir(worktreeRepo); err != nil {
		t.Fatalf("chdir worktree: %v", err)
	}
	run(t, "task", "new", "worktree-task", "--content", "from worktree")

	if err := os.Chdir(mainRepo); err != nil {
		t.Fatalf("chdir main repo: %v", err)
	}
	out := run(t, "task", "list")
	if !strings.Contains(out, "worktree-task") {
		t.Fatalf("task created in worktree should be visible in canonical project, got:\n%s", out)
	}

	out = run(t, "task", "list", "--project", "feature-worktree")
	if strings.Contains(out, "worktree-task") {
		t.Fatalf("task should not be stored under worktree folder name, got:\n%s", out)
	}
}

func TestTask_RepoPathExplicitTargeting(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	setup(t)

	baseDir := t.TempDir()
	repoPath := filepath.Join(baseDir, "project-alpha")
	otherDir := filepath.Join(baseDir, "elsewhere")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatalf("mkdir other: %v", err)
	}

	runGit(t, repoPath, "init")
	runGit(t, repoPath, "config", "user.email", "test@example.com")
	runGit(t, repoPath, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("seed"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	runGit(t, repoPath, "add", "README.md")
	runGit(t, repoPath, "commit", "-m", "init")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(orig)
	if err := os.Chdir(otherDir); err != nil {
		t.Fatalf("chdir other: %v", err)
	}

	run(t, "task", "new", "repo-task", "--repo-path", repoPath, "--content", "explicit repo target")
	out := run(t, "task", "list", "--repo-path", repoPath)
	if !strings.Contains(out, "repo-task") {
		t.Fatalf("repo-targeted task should be listed via --repo-path, got:\n%s", out)
	}

	out = run(t, "task", "list")
	if strings.Contains(out, "repo-task") {
		t.Fatalf("task should not appear for unrelated cwd project, got:\n%s", out)
	}
}

func TestTask_ListJSON(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "json-task", "--content", "json payload")
	out := run(t, "task", "list", "--json")

	var payload struct {
		Project string `json:"project"`
		Tasks   []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("expected valid JSON payload, got error: %v\noutput:\n%s", err, out)
	}
	if payload.Project == "" {
		t.Fatal("expected project in JSON payload")
	}
	if len(payload.Tasks) == 0 {
		t.Fatal("expected at least one task in JSON payload")
	}
	if payload.Tasks[0].ID == "" {
		t.Fatal("expected task id in JSON payload")
	}
}

func TestTask_InvalidTransitionCodedError(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "transition-task", "--content", "transition")
	run(t, "task", "status", "transition-task", "completed")
	_, err := runE(t, "task", "status", "transition-task", "pending", "--json")
	if err == nil {
		t.Fatal("expected invalid transition error")
	}

	type coded interface {
		ErrorCode() string
		ExitStatus() int
	}
	codedErr, ok := err.(coded)
	if !ok {
		t.Fatalf("expected coded error, got %T", err)
	}
	if codedErr.ErrorCode() != "invalid_transition" {
		t.Fatalf("expected code invalid_transition, got %s", codedErr.ErrorCode())
	}
	if codedErr.ExitStatus() != 5 {
		t.Fatalf("expected exit status 5, got %d", codedErr.ExitStatus())
	}
}

func TestTask_WorktreeIDStable(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	setup(t)

	baseDir := t.TempDir()
	mainRepo := filepath.Join(baseDir, "stable-id-repo")
	worktreeRepo := filepath.Join(baseDir, "stable-id-worktree")
	if err := os.MkdirAll(mainRepo, 0o755); err != nil {
		t.Fatalf("mkdir main repo: %v", err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(orig)

	if err := os.Chdir(mainRepo); err != nil {
		t.Fatalf("chdir main repo: %v", err)
	}
	runGit(t, mainRepo, "init")
	runGit(t, mainRepo, "config", "user.email", "test@example.com")
	runGit(t, mainRepo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("seed"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "init")
	runGit(t, mainRepo, "worktree", "add", "-b", "feature-y", worktreeRepo)

	if err := os.Chdir(worktreeRepo); err != nil {
		t.Fatalf("chdir worktree: %v", err)
	}
	run(t, "task", "new", "stable-task", "--content", "stable id")

	worktreeOut := run(t, "task", "show", "stable-task", "--json")
	if err := os.Chdir(mainRepo); err != nil {
		t.Fatalf("chdir main repo: %v", err)
	}
	mainOut := run(t, "task", "show", "stable-task", "--json")

	var worktreeTask struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(worktreeOut), &worktreeTask); err != nil {
		t.Fatalf("invalid worktree json: %v", err)
	}
	var mainTask struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(mainOut), &mainTask); err != nil {
		t.Fatalf("invalid main json: %v", err)
	}

	if worktreeTask.ID == "" || mainTask.ID == "" {
		t.Fatalf("expected non-empty ids, got worktree=%q main=%q", worktreeTask.ID, mainTask.ID)
	}
	if worktreeTask.ID != mainTask.ID {
		t.Fatalf("expected stable id across worktree and main repo, got %q vs %q", worktreeTask.ID, mainTask.ID)
	}
}
