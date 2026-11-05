package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"agmd/cmd"
	"agmd/pkg/registry"
)

func boardServer(t *testing.T) *httptest.Server {
	t.Helper()
	reg, err := registry.New()
	if err != nil {
		t.Fatalf("registry.New: %v", err)
	}
	srv := httptest.NewServer(cmd.NewTaskWebHandler(reg, "proj-a"))
	t.Cleanup(srv.Close)
	return srv
}

func webDo(t *testing.T, method, url string, body interface{}, header map[string]string) (int, []byte) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		if k == "Host" {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

type boardTask struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Project        string   `json:"project"`
	Subject        string   `json:"subject"`
	Status         string   `json:"status"`
	ComputedStatus string   `json:"computed_status"`
	Priority       int      `json:"priority"`
	DependsOn      []string `json:"depends_on"`
	Assignment     *struct {
		WorkflowID string `json:"workflow_id"`
	} `json:"assignment"`
}

func decodeTask(t *testing.T, data []byte) boardTask {
	t.Helper()
	var task boardTask
	if err := json.Unmarshal(data, &task); err != nil {
		t.Fatalf("decode task: %v\n%s", err, data)
	}
	return task
}

func TestTaskWeb_CreateTask(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "dep", "--project", "proj-a", "--content", "first")
	srv := boardServer(t)

	status, data := webDo(t, "POST", srv.URL+"/api/tasks", map[string]interface{}{
		"project": "proj-a", "subject": "Rate limit: export", "content": "429 + Retry-After",
		"priority": 1, "type": "feature", "depends_on": []string{"dep"},
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("create = %d %s", status, data)
	}
	task := decodeTask(t, data)
	if task.Name != "rate-limit-export" || task.Subject != "Rate limit: export" || task.ComputedStatus != "blocked" || task.Priority != 1 {
		t.Errorf("created task = %+v", task)
	}
	if got := readFile(t, filepath.Join(registryPath, "task", "proj-a", "rate-limit-export.md")); !strings.Contains(got, "'Rate limit: export'") {
		t.Errorf("subject with a colon must be quoted in YAML:\n%s", got)
	}

	// Same title again gets a unique name.
	_, data = webDo(t, "POST", srv.URL+"/api/tasks", map[string]interface{}{"project": "proj-a", "subject": "Rate limit: export"}, nil)
	if task := decodeTask(t, data); task.Name != "rate-limit-export-2" || task.Priority != 2 {
		t.Errorf("second task = %+v", task)
	}

	// A brand-new project is created on first task.
	status, _ = webDo(t, "POST", srv.URL+"/api/tasks", map[string]interface{}{"project": "fresh", "subject": "Hello"}, nil)
	if status != http.StatusOK {
		t.Errorf("create in new project = %d", status)
	}

	for _, bad := range []map[string]interface{}{
		{"project": "../escape", "subject": "x"},
		{"project": ".hidden", "subject": "x"},
		{"project": "proj-a", "subject": "  "},
		{"project": "proj-a", "subject": "x", "depends_on": []string{"missing"}},
		{"project": "proj-a", "subject": "x", "priority": 9},
	} {
		if status, data := webDo(t, "POST", srv.URL+"/api/tasks", bad, nil); status < 400 || status >= 500 {
			t.Errorf("create %v = %d %s, want 4xx", bad, status, data)
		}
	}
}

func TestTaskWeb_DeleteTaskDropsClaimAndComments(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")
	run(t, "task", "next", "--claim", "--workflow-id", "agent-1", "--project", "proj-a")
	run(t, "task", "comment", "a", "note", "--project", "proj-a")
	srv := boardServer(t)

	if status, data := webDo(t, "DELETE", srv.URL+"/api/tasks/proj-a/a", nil, nil); status != http.StatusOK {
		t.Fatalf("delete = %d %s", status, data)
	}
	if fileExists(filepath.Join(registryPath, "task", "proj-a", "a.md")) {
		t.Error("task file still exists")
	}
	if got := readFile(t, filepath.Join(registryPath, "task", "proj-a", "assignments.json")); strings.Contains(got, "agent-1") {
		t.Errorf("claim left behind: %s", got)
	}
	if status, _ := webDo(t, "DELETE", srv.URL+"/api/tasks/proj-a/a", nil, nil); status != http.StatusNotFound {
		t.Errorf("second delete = %d, want 404", status)
	}
}

func TestTaskWeb_ReleaseClaim(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")
	run(t, "task", "next", "--claim", "--workflow-id", "agent-1", "--project", "proj-a")
	srv := boardServer(t)

	status, data := webDo(t, "DELETE", srv.URL+"/api/tasks/proj-a/a/claim", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("release = %d %s", status, data)
	}
	task := decodeTask(t, data)
	if task.Assignment != nil || task.Status != "pending" || task.ComputedStatus != "ready" {
		t.Errorf("after release = %+v", task)
	}
	if status, _ := webDo(t, "DELETE", srv.URL+"/api/tasks/proj-a/a/claim", nil, nil); status != http.StatusNotFound {
		t.Errorf("releasing an unclaimed task = %d, want 404", status)
	}
	if out := run(t, "task", "next", "--claim", "--workflow-id", "agent-2", "--project", "proj-a", "--json"); !strings.Contains(out, `"action":"claim"`) {
		t.Errorf("released task should be claimable again:\n%s", out)
	}
}

func TestTaskWeb_PatchDependencies(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")
	run(t, "task", "new", "b", "--project", "proj-a", "--content", "work")
	srv := boardServer(t)

	status, data := webPatch(t, srv.URL, "proj-a", "b", map[string]interface{}{"depends_on": []string{"a"}})
	if status != http.StatusOK || decodeTask(t, data).ComputedStatus != "blocked" {
		t.Fatalf("add dependency = %d %s", status, data)
	}
	if status, _ := webPatch(t, srv.URL, "proj-a", "b", map[string]interface{}{"depends_on": []string{"b"}}); status != http.StatusBadRequest {
		t.Errorf("self dependency = %d, want 400", status)
	}
	status, data = webPatch(t, srv.URL, "proj-a", "b", map[string]interface{}{"depends_on": []string{}})
	if status != http.StatusOK || decodeTask(t, data).ComputedStatus != "ready" {
		t.Errorf("clear dependencies = %d %s", status, data)
	}
}

func TestTaskWeb_ChangesTracksCLIWrites(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")
	srv := boardServer(t)

	version := func() string {
		_, data := webDo(t, "GET", srv.URL+"/api/changes", nil, nil)
		var v struct{ Version string }
		json.Unmarshal(data, &v) //nolint:errcheck
		return v.Version
	}
	v1 := version()
	if v1 == "" || version() != v1 {
		t.Fatalf("version should be stable while nothing changes: %q", v1)
	}
	run(t, "task", "comment", "a", "progress", "--author", "agent-1", "--project", "proj-a")
	v2 := version()
	if v2 == v1 {
		t.Error("a CLI comment should change the version")
	}
	run(t, "task", "next", "--claim", "--workflow-id", "agent-1", "--project", "proj-a")
	if version() == v2 {
		t.Error("a CLI claim should change the version")
	}
}

func TestTaskWeb_StateReportsTTLAndEmoji(t *testing.T) {
	setup(t)
	chdir(t)
	t.Setenv("AGMD_CLAIM_TTL", "45m")
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")
	run(t, "task", "emoji", "🚀", "--project", "proj-a")
	srv := boardServer(t)

	_, data := webDo(t, "GET", srv.URL+"/api/state", nil, nil)
	var state struct {
		ClaimTTLSeconds int `json:"claimTtlSeconds"`
		Projects        []struct {
			Name  string `json:"name"`
			Emoji string `json:"emoji"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	if state.ClaimTTLSeconds != 45*60 {
		t.Errorf("claimTtlSeconds = %d, want 2700", state.ClaimTTLSeconds)
	}
	if len(state.Projects) != 1 || state.Projects[0].Emoji != "🚀" {
		t.Errorf("projects = %+v", state.Projects)
	}
}

func TestTaskWeb_RejectsCrossSiteRequests(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "proj-a", "--content", "work")
	srv := boardServer(t)

	body := map[string]interface{}{"project": "proj-a", "subject": "x"}
	if status, _ := webDo(t, "POST", srv.URL+"/api/tasks", body, map[string]string{"Origin": "http://evil.example"}); status != http.StatusForbidden {
		t.Errorf("foreign Origin = %d, want 403", status)
	}
	if status, _ := webDo(t, "GET", srv.URL+"/api/state", nil, map[string]string{"Host": "evil.example:80"}); status != http.StatusForbidden {
		t.Errorf("DNS-rebinding Host = %d, want 403", status)
	}
	if status, _ := webDo(t, "POST", srv.URL+"/api/tasks", body, map[string]string{"Origin": srv.URL}); status != http.StatusOK {
		t.Errorf("same-origin create = %d, want 200", status)
	}
}

func TestTaskWeb_ServesBoardAssets(t *testing.T) {
	setup(t)
	srv := boardServer(t)
	for path, want := range map[string]string{"/": "board.js", "/board.js": "App.loadState", "/app.css": "--accent", "/panel.js": "App.Panel", "/runs.js": "App.Runs", "/ui.js": "App.confirm"} {
		status, data := webDo(t, "GET", srv.URL+path, nil, nil)
		if status != http.StatusOK || !strings.Contains(string(data), want) {
			t.Errorf("GET %s = %d, missing %q", path, status, want)
		}
	}
}
