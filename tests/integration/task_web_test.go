package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agmd/cmd"
	"agmd/pkg/registry"
)

// webTaskItem mirrors the JSON shape returned by the task web API.
type webTaskItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Project        string `json:"project"`
	Subject        string `json:"subject"`
	Status         string `json:"status"`
	ComputedStatus string `json:"computed_status"`
	Priority       int    `json:"priority"`
	Type           string `json:"type"`
	Feature        string `json:"feature"`
}

type webStateResp struct {
	CurrentProject string `json:"currentProject"`
	Projects       []struct {
		Name   string `json:"name"`
		Counts struct {
			Total      int `json:"total"`
			Ready      int `json:"ready"`
			InProgress int `json:"in_progress"`
			Completed  int `json:"completed"`
		} `json:"counts"`
	} `json:"projects"`
	Tasks []webTaskItem `json:"tasks"`
}

type webErrResp struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func webGetState(t *testing.T, base string) webStateResp {
	t.Helper()
	resp, err := http.Get(base + "/api/state")
	if err != nil {
		t.Fatalf("GET /api/state: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/state status = %d", resp.StatusCode)
	}
	var state webStateResp
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	return state
}

func webPatch(t *testing.T, base, project, id string, body map[string]interface{}) (int, []byte) {
	t.Helper()
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPatch, base+"/api/tasks/"+project+"/"+id, bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new PATCH request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s/%s: %v", project, id, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func findWebTask(tasks []webTaskItem, project, name string) (webTaskItem, bool) {
	for _, task := range tasks {
		if task.Project == project && task.Name == name {
			return task, true
		}
	}
	return webTaskItem{}, false
}

func webPost(t *testing.T, url string, body map[string]interface{}) (int, []byte) {
	t.Helper()
	raw, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func webGet(t *testing.T, url string) (int, []byte) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

// writeStubClaude installs a fake `claude` binary (exits 0, does nothing) and
// points AGMD_CLAUDE_BIN at it so run tests never spawn the real CLI.
func writeStubClaude(t *testing.T) {
	t.Helper()
	stub := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write stub claude: %v", err)
	}
	t.Setenv("AGMD_CLAUDE_BIN", stub)
}

func TestTaskWeb_RunLaunches(t *testing.T) {
	setup(t)
	chdir(t)
	t.Setenv("HOME", t.TempDir())
	writeStubClaude(t)

	run(t, "task", "new", "runnable", "--project", "proj-a", "--content", "Implement runnable.")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	status, data := webPost(t, srv.URL+"/api/run", map[string]interface{}{
		"tasks":          []map[string]string{{"project": "proj-a", "id": "runnable"}},
		"permissionMode": "bypassPermissions",
	})
	if status != http.StatusOK {
		t.Fatalf("POST /api/run = %d, body=%s", status, data)
	}

	var info struct {
		SessionID      string `json:"sessionId"`
		TranscriptPath string `json:"transcriptPath"`
		ResumeHint     string `json:"resumeHint"`
		PermissionMode string `json:"permissionMode"`
		Tasks          []struct {
			Name string `json:"name"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatalf("decode run info: %v (body=%s)", err, data)
	}
	if info.SessionID == "" {
		t.Error("response missing sessionId")
	}
	if !strings.HasSuffix(info.TranscriptPath, info.SessionID+".jsonl") {
		t.Errorf("transcriptPath %q should end with <sessionId>.jsonl", info.TranscriptPath)
	}
	if !strings.Contains(info.ResumeHint, info.SessionID) {
		t.Errorf("resumeHint %q should contain the session id", info.ResumeHint)
	}
	if info.PermissionMode != "bypassPermissions" {
		t.Errorf("permissionMode = %q, want bypassPermissions", info.PermissionMode)
	}
	if len(info.Tasks) != 1 || info.Tasks[0].Name != "runnable" {
		t.Errorf("tasks = %+v, want [runnable]", info.Tasks)
	}

	// The launched task should have been flipped to in_progress.
	state := webGetState(t, srv.URL)
	task, ok := findWebTask(state.Tasks, "proj-a", "runnable")
	if !ok {
		t.Fatal("runnable task missing from state")
	}
	if task.Status != "in_progress" {
		t.Errorf("launched task status = %q, want in_progress", task.Status)
	}

	// The run should be queryable by id.
	rstatus, rdata := webGet(t, srv.URL+"/api/run/"+info.SessionID)
	if rstatus != http.StatusOK {
		t.Fatalf("GET /api/run/{id} = %d, body=%s", rstatus, rdata)
	}
}

func TestTaskWeb_RunTranscriptEvents(t *testing.T) {
	setup(t)
	chdir(t)
	t.Setenv("HOME", t.TempDir())
	writeStubClaude(t)

	run(t, "task", "new", "tscript", "--project", "proj-a", "--content", "x")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	_, data := webPost(t, srv.URL+"/api/run", map[string]interface{}{
		"tasks": []map[string]string{{"project": "proj-a", "id": "tscript"}},
	})
	var info struct {
		SessionID      string `json:"sessionId"`
		TranscriptPath string `json:"transcriptPath"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatalf("decode run info: %v", err)
	}

	// Simulate Claude appending to its session transcript.
	if err := os.MkdirAll(filepath.Dir(info.TranscriptPath), 0o755); err != nil {
		t.Fatalf("mkdir transcript dir: %v", err)
	}
	line := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hello from claude"}]}}` + "\n"
	if err := os.WriteFile(info.TranscriptPath, []byte(line), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	_, rdata := webGet(t, srv.URL+"/api/run/"+info.SessionID)
	var rs struct {
		Events []struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rdata, &rs); err != nil {
		t.Fatalf("decode run status: %v (body=%s)", err, rdata)
	}
	if len(rs.Events) == 0 {
		t.Fatalf("expected transcript events, got none (body=%s)", rdata)
	}
	last := rs.Events[len(rs.Events)-1]
	if !strings.Contains(last.Text, "hello from claude") {
		t.Errorf("last event text = %q, want it to contain transcript output", last.Text)
	}
}

func TestTaskWeb_RunList(t *testing.T) {
	setup(t)
	chdir(t)
	t.Setenv("HOME", t.TempDir())
	writeStubClaude(t)

	run(t, "task", "new", "task-one", "--project", "proj-a", "--content", "one")
	run(t, "task", "new", "task-two", "--project", "proj-a", "--content", "two")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	// Empty list before any runs.
	_, data := webGet(t, srv.URL+"/api/runs")
	var empty struct {
		Runs []json.RawMessage `json:"runs"`
	}
	json.Unmarshal(data, &empty) //nolint:errcheck
	if len(empty.Runs) != 0 {
		t.Errorf("expected no runs initially, got %d", len(empty.Runs))
	}

	// Launch two separate runs.
	var ids []string
	for _, name := range []string{"task-one", "task-two"} {
		_, rd := webPost(t, srv.URL+"/api/run", map[string]interface{}{
			"tasks": []map[string]string{{"project": "proj-a", "id": name}},
		})
		var info struct {
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(rd, &info); err != nil {
			t.Fatalf("decode run info: %v", err)
		}
		ids = append(ids, info.SessionID)
	}

	_, data = webGet(t, srv.URL+"/api/runs")
	var list struct {
		Runs []struct {
			SessionID string `json:"sessionId"`
			Tasks     []struct {
				Name string `json:"name"`
			} `json:"tasks"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		t.Fatalf("decode runs list: %v (body=%s)", err, data)
	}
	if len(list.Runs) != 2 {
		t.Fatalf("expected 2 runs, got %d (body=%s)", len(list.Runs), data)
	}
	seen := map[string]bool{}
	for _, r := range list.Runs {
		seen[r.SessionID] = true
	}
	for _, id := range ids {
		if !seen[id] {
			t.Errorf("runs list missing session %s", id)
		}
	}
}

func TestTaskWeb_RunValidation(t *testing.T) {
	setup(t)
	chdir(t)
	writeStubClaude(t)

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	// Empty task list -> 400.
	if status, _ := webPost(t, srv.URL+"/api/run", map[string]interface{}{"tasks": []map[string]string{}}); status != http.StatusBadRequest {
		t.Errorf("empty tasks: status = %d, want 400", status)
	}

	// Invalid permission mode -> 400.
	run(t, "task", "new", "vtask", "--project", "proj-a", "--content", "v")
	status, data := webPost(t, srv.URL+"/api/run", map[string]interface{}{
		"tasks":          []map[string]string{{"project": "proj-a", "id": "vtask"}},
		"permissionMode": "wideOpen",
	})
	if status != http.StatusBadRequest {
		t.Errorf("invalid permission mode: status = %d, want 400 (body=%s)", status, data)
	}

	// Unknown run id -> 404.
	if status, _ := webGet(t, srv.URL+"/api/run/does-not-exist"); status != http.StatusNotFound {
		t.Errorf("unknown run: status = %d, want 404", status)
	}
}

func TestTaskWeb_StateAcrossProjects(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "alpha-one", "--project", "proj-a", "-t", "feature", "-p", "1", "--content", "Alpha one.")
	run(t, "task", "new", "alpha-two", "--project", "proj-a", "--content", "Alpha two.")
	run(t, "task", "new", "beta-one", "--project", "proj-b", "-t", "bug", "--content", "Beta one.")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	state := webGetState(t, srv.URL)

	if state.CurrentProject != "proj-a" {
		t.Errorf("currentProject = %q, want proj-a", state.CurrentProject)
	}

	projects := map[string]int{}
	for _, p := range state.Projects {
		projects[p.Name] = p.Counts.Total
	}
	if projects["proj-a"] != 2 {
		t.Errorf("proj-a total = %d, want 2", projects["proj-a"])
	}
	if projects["proj-b"] != 1 {
		t.Errorf("proj-b total = %d, want 1", projects["proj-b"])
	}

	if _, ok := findWebTask(state.Tasks, "proj-a", "alpha-one"); !ok {
		t.Error("state.tasks missing proj-a/alpha-one")
	}
	if _, ok := findWebTask(state.Tasks, "proj-b", "beta-one"); !ok {
		t.Error("state.tasks missing proj-b/beta-one")
	}
}

func TestTaskWeb_PatchPersists(t *testing.T) {
	registryPath := setup(t)
	chdir(t)

	run(t, "task", "new", "editable", "--project", "proj-a", "--content", "Editable task.")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	// PATCH status -> in_progress, returns the updated item.
	status, data := webPatch(t, srv.URL, "proj-a", "editable", map[string]interface{}{"status": "in_progress"})
	if status != http.StatusOK {
		t.Fatalf("PATCH status = %d, body=%s", status, data)
	}
	var updated webTaskItem
	if err := json.Unmarshal(data, &updated); err != nil {
		t.Fatalf("decode patch response: %v", err)
	}
	if updated.Status != "in_progress" || updated.ComputedStatus != "in_progress" {
		t.Errorf("after status patch: status=%q computed=%q", updated.Status, updated.ComputedStatus)
	}

	// PATCH priority -> 0 (P0).
	status, data = webPatch(t, srv.URL, "proj-a", "editable", map[string]interface{}{"priority": 0})
	if status != http.StatusOK {
		t.Fatalf("PATCH priority = %d, body=%s", status, data)
	}

	// Re-GET to confirm persistence through a fresh disk read.
	state := webGetState(t, srv.URL)
	task, ok := findWebTask(state.Tasks, "proj-a", "editable")
	if !ok {
		t.Fatal("editable task missing after patch")
	}
	if task.Status != "in_progress" {
		t.Errorf("persisted status = %q, want in_progress", task.Status)
	}
	if task.Priority != 0 {
		t.Errorf("persisted priority = %d, want 0", task.Priority)
	}

	// Confirm the change reached the underlying task file on disk.
	taskFile := filepath.Join(registryPath, "task", "proj-a", "editable.md")
	if got := readFile(t, taskFile); !strings.Contains(got, "status: in_progress") {
		t.Errorf("task file does not contain persisted status:\n%s", got)
	}
}

func TestTaskWeb_RejectsCompletedTransition(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "locked", "--project", "proj-a", "--content", "Locked task.")

	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	defer srv.Close()

	if status, data := webPatch(t, srv.URL, "proj-a", "locked", map[string]interface{}{"status": "completed"}); status != http.StatusOK {
		t.Fatalf("completing task: status=%d body=%s", status, data)
	}

	status, data := webPatch(t, srv.URL, "proj-a", "locked", map[string]interface{}{"status": "pending"})
	if status >= 200 && status < 300 {
		t.Fatalf("expected completed->pending to be rejected, got status %d", status)
	}
	var errResp webErrResp
	if err := json.Unmarshal(data, &errResp); err != nil {
		t.Fatalf("decode error response: %v (body=%s)", err, data)
	}
	if errResp.Error.Code != "invalid_transition" {
		t.Errorf("error code = %q, want invalid_transition", errResp.Error.Code)
	}
}
