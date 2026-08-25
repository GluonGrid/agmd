package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

//go:embed webui/index.html
var webuiFS embed.FS

// Web command flags (project/repo-path are shared with the other subcommands).
var (
	taskWebPort   int
	taskWebHost   string
	taskWebNoOpen bool
)

var taskWebCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch a local web UI to browse and edit tasks",
	Long: `Launch a local web UI to browse and edit tasks across all projects.

Starts a small HTTP server (by default on a free ephemeral port bound to
127.0.0.1) and opens it in your browser. The UI aggregates tasks from every
project in your registry and defaults the project filter to the current
repository's project. Filter by project/status/priority/type/feature, edit
status and priority inline, and edit subject/type/feature/content in a detail
panel.

Changes are written straight back to the task files, so they are immediately
reflected in 'agmd task list'.

Examples:
  agmd task web                  # serve on a free port and open the browser
  agmd task web --no-open        # serve without opening a browser
  agmd task web --port 8765      # serve on a fixed port
  agmd task web --host 0.0.0.0   # bind all interfaces (use with care)`,
	Args: cobra.NoArgs,
	RunE: runTaskWeb,
}

func init() {
	taskCmd.AddCommand(taskWebCmd)
	taskWebCmd.Flags().IntVar(&taskWebPort, "port", 0, "Port to listen on (0 = pick a free port)")
	taskWebCmd.Flags().StringVar(&taskWebHost, "host", "127.0.0.1", "Host/interface to bind")
	taskWebCmd.Flags().BoolVar(&taskWebNoOpen, "no-open", false, "Do not open a browser automatically")
	taskWebCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskWebCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
}

// --- web state payloads ---

type webCounts struct {
	Total      int `json:"total"`
	Ready      int `json:"ready"`
	Blocked    int `json:"blocked"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
}

type webProjectInfo struct {
	Name   string    `json:"name"`
	Counts webCounts `json:"counts"`
}

type webStatePayload struct {
	CurrentProject string           `json:"currentProject"`
	Projects       []webProjectInfo `json:"projects"`
	Tasks          []taskJSONItem   `json:"tasks"`
}

// listProjectNames returns the sorted set of project names that have a task
// directory under <registry>/task.
func listProjectNames(reg *registry.Registry) ([]string, error) {
	taskRoot := filepath.Join(reg.BasePath, "task")
	entries, err := os.ReadDir(taskRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// buildTaskMap indexes tasks by both name and ID so dependency lookups resolve
// regardless of which form is stored in depends_on.
func buildTaskMap(tasks []*Task) map[string]*Task {
	taskMap := make(map[string]*Task, len(tasks)*2)
	for _, t := range tasks {
		taskMap[t.Name] = t
		if t.ID != "" {
			taskMap[t.ID] = t
		}
	}
	return taskMap
}

// loadWebState aggregates tasks and per-project counts across every project.
func loadWebState(reg *registry.Registry, currentProject string) (webStatePayload, error) {
	state := webStatePayload{
		CurrentProject: currentProject,
		Projects:       []webProjectInfo{},
		Tasks:          []taskJSONItem{},
	}

	names, err := listProjectNames(reg)
	if err != nil {
		return state, err
	}

	for _, name := range names {
		tasks, err := loadProjectTasks(reg, name)
		if err != nil {
			return state, err
		}
		assignments, err := loadTaskAssignments(reg, name)
		if err != nil {
			return state, err
		}

		taskMap := buildTaskMap(tasks)
		payload := buildTaskListJSON(name, "", tasks, taskMap, assignments, true)

		state.Projects = append(state.Projects, webProjectInfo{
			Name: name,
			Counts: webCounts{
				Total:      payload.Counts.Total,
				Ready:      payload.Counts.Ready,
				Blocked:    payload.Counts.Blocked,
				InProgress: payload.Counts.InProgress,
				Completed:  payload.Counts.Completed,
			},
		})
		state.Tasks = append(state.Tasks, payload.Tasks...)
	}

	return state, nil
}

// --- HTTP handler ---

type taskWebHandler struct {
	reg            *registry.Registry
	currentProject string
	mu             sync.Mutex // serializes task file reads/writes

	claudeBin string
	runMu     sync.Mutex // guards runs
	runs      map[string]*runInfo
}

// NewTaskWebHandler builds the task web HTTP handler. It carries no global
// state, so it can be exercised directly via httptest in integration tests.
func NewTaskWebHandler(reg *registry.Registry, currentProject string) http.Handler {
	claudeBin := os.Getenv("AGMD_CLAUDE_BIN")
	if claudeBin == "" {
		claudeBin = "claude"
	}
	h := &taskWebHandler{
		reg:            reg,
		currentProject: currentProject,
		claudeBin:      claudeBin,
		runs:           map[string]*runInfo{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", h.handleState)
	mux.HandleFunc("PATCH /api/tasks/{project}/{id}", h.handlePatch)
	mux.HandleFunc("POST /api/run", h.handleRun)
	mux.HandleFunc("GET /api/runs", h.handleRunList)
	mux.HandleFunc("GET /api/run/{id}", h.handleRunStatus)
	mux.HandleFunc("GET /", h.handleIndex)
	return mux
}

func (h *taskWebHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := webuiFS.ReadFile("webui/index.html")
	if err != nil {
		http.Error(w, "web UI asset missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data) //nolint:errcheck
}

func (h *taskWebHandler) handleState(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	state, err := loadWebState(h.reg, h.currentProject)
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeWebJSON(w, http.StatusOK, state)
}

// webPriority accepts a JSON number (0-4) or a string ("P0".."P4" / "0".."4").
type webPriority struct {
	value int
}

func (p *webPriority) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		p.value = n
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("priority must be a number or string")
	}
	parsed, err := parsePriority(s)
	if err != nil {
		return err
	}
	p.value = parsed
	return nil
}

type taskPatchBody struct {
	Status   *string      `json:"status"`
	Priority *webPriority `json:"priority"`
	Type     *string      `json:"type"`
	Feature  *string      `json:"feature"`
	Subject  *string      `json:"subject"`
	Content  *string      `json:"content"`
}

func (h *taskWebHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	id := r.PathValue("id")

	var body taskPatchBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid JSON body: %v", err))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	task, err := resolveTaskRef(h.reg, project, id)
	if err != nil {
		status, code := webErrorStatus(err)
		writeWebError(w, status, code, err.Error())
		return
	}

	if body.Status != nil {
		if err := setTaskStatus(task, *body.Status); err != nil {
			status, code := webErrorStatus(err)
			writeWebError(w, status, code, err.Error())
			return
		}
	}
	if body.Priority != nil {
		if body.Priority.value < 0 || body.Priority.value > 4 {
			writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid priority %d. Use 0-4", body.Priority.value))
			return
		}
		task.Priority = body.Priority.value
	}
	if body.Type != nil {
		newType := strings.ToLower(strings.TrimSpace(*body.Type))
		if newType != "" && !validTaskTypes[newType] {
			writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid type '%s'. Use: bug, feature, task, chore", newType))
			return
		}
		task.Type = newType
	}
	if body.Feature != nil {
		task.Feature = strings.TrimSpace(*body.Feature)
	}
	if body.Subject != nil {
		task.Subject = strings.TrimSpace(*body.Subject)
	}
	if body.Content != nil {
		task.Content = strings.TrimSpace(*body.Content)
	}

	if err := saveTask(task); err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", fmt.Sprintf("failed to save task: %v", err))
		return
	}

	// Reload the project so computed_status reflects sibling task states.
	tasks, err := loadProjectTasks(h.reg, project)
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	assignments, err := loadTaskAssignments(h.reg, project)
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	taskMap := buildTaskMap(tasks)
	item := taskToJSONItem(task, taskMap, assignments, true)
	writeWebJSON(w, http.StatusOK, item)
}

// webErrorStatus maps a coded task error to an HTTP status + machine code.
func webErrorStatus(err error) (int, string) {
	if coded, ok := err.(codedError); ok {
		switch coded.ErrorCode() {
		case "invalid_input":
			return http.StatusBadRequest, "invalid_input"
		case "not_found":
			return http.StatusNotFound, "not_found"
		case "project_mismatch":
			return http.StatusConflict, "project_mismatch"
		case "invalid_transition":
			return http.StatusConflict, "invalid_transition"
		default:
			return http.StatusBadRequest, coded.ErrorCode()
		}
	}
	return http.StatusInternalServerError, "internal"
}

func writeWebJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload) //nolint:errcheck
}

func writeWebError(w http.ResponseWriter, status int, code, message string) {
	writeWebJSON(w, status, map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

// --- batch run via Claude Code ---

type runTaskRef struct {
	Project string `json:"project"`
	ID      string `json:"id"`
}

type runRequest struct {
	Tasks          []runTaskRef `json:"tasks"`
	PermissionMode string       `json:"permissionMode,omitempty"`
}

type runTaskSummary struct {
	Project string `json:"project"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
}

type runInfo struct {
	SessionID      string           `json:"sessionId"`
	TranscriptPath string           `json:"transcriptPath"`
	LogPath        string           `json:"logPath,omitempty"`
	Cwd            string           `json:"cwd"`
	PermissionMode string           `json:"permissionMode"`
	ResumeHint     string           `json:"resumeHint"`
	Prompt         string           `json:"prompt"`
	Tasks          []runTaskSummary `json:"tasks"`
	StartedAt      string           `json:"startedAt"`
	Done           bool             `json:"done"`
	ExitError      string           `json:"exitError,omitempty"`

	cmd *exec.Cmd
}

type runEvent struct {
	Kind string `json:"kind"`
	Role string `json:"role,omitempty"`
	Text string `json:"text"`
}

// claude --permission-mode accepts these modes.
var validPermissionModes = map[string]bool{
	"default":           true,
	"acceptEdits":       true,
	"plan":              true,
	"bypassPermissions": true,
}

// newSessionID returns a random RFC-4122 v4 UUID string.
func newSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// buildGoalPrompt assembles the /goal prompt that instructs a Claude Code
// session to complete the selected tasks.
func buildGoalPrompt(tasks []*Task) string {
	var b strings.Builder
	b.WriteString("/goal the following agmd tasks should be completed:\n")
	for _, t := range tasks {
		subject := t.Subject
		if subject == "" {
			subject = t.Name
		}
		ref := t.Name
		if t.ProjectName != "" {
			ref = t.ProjectName + "/" + t.Name
		}
		fmt.Fprintf(&b, "\n- %s — %s\n", ref, subject)
		content := strings.TrimSpace(t.Content)
		if content != "" {
			for _, line := range strings.Split(content, "\n") {
				b.WriteString("  ")
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}
	}
	b.WriteString("\nImplement each task, then run 'agmd task status <id> completed' to mark it done once it is truly complete and verified.")
	return b.String()
}

// transcriptDirSlug encodes an absolute path the way Claude Code names its
// per-project transcript directory: path separators become dashes.
func transcriptDirSlug(absPath string) string {
	return strings.ReplaceAll(absPath, string(os.PathSeparator), "-")
}

// sessionTranscriptPath returns the deterministic transcript JSONL path for a
// session started in cwd.
func sessionTranscriptPath(cwd, sessionID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects", transcriptDirSlug(abs), sessionID+".jsonl"), nil
}

// locateTranscript returns the transcript path, falling back to a glob across
// all project dirs in case the directory-name encoding differs from our guess.
func locateTranscript(expected, sessionID string) string {
	if _, err := os.Stat(expected); err == nil {
		return expected
	}
	if home, err := os.UserHomeDir(); err == nil {
		matches, _ := filepath.Glob(filepath.Join(home, ".claude", "projects", "*", sessionID+".jsonl"))
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return expected
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// summarizeMessage extracts (role, text) from a transcript line's message,
// handling both string content and content-block arrays.
func summarizeMessage(raw json.RawMessage) (string, string) {
	if len(raw) == 0 {
		return "", ""
	}
	var msg struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return "", ""
	}

	var asString string
	if err := json.Unmarshal(msg.Content, &asString); err == nil {
		return msg.Role, asString
	}

	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		return msg.Role, ""
	}

	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		switch b.Type {
		case "text":
			if strings.TrimSpace(b.Text) != "" {
				parts = append(parts, b.Text)
			}
		case "tool_use":
			parts = append(parts, "🔧 "+b.Name)
		case "tool_result":
			parts = append(parts, "↩ tool result")
		}
	}
	return msg.Role, strings.Join(parts, "\n")
}

func transcriptLineToEvent(line string) (runEvent, bool) {
	var rawLine struct {
		Type    string          `json:"type"`
		Summary string          `json:"summary"`
		Message json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal([]byte(line), &rawLine); err != nil {
		return runEvent{}, false
	}

	switch rawLine.Type {
	case "summary":
		if strings.TrimSpace(rawLine.Summary) == "" {
			return runEvent{}, false
		}
		return runEvent{Kind: "summary", Text: rawLine.Summary}, true
	case "user", "assistant":
		role, text := summarizeMessage(rawLine.Message)
		text = strings.TrimSpace(text)
		if text == "" {
			return runEvent{}, false
		}
		return runEvent{Kind: rawLine.Type, Role: role, Text: truncate(text, 4000)}, true
	default:
		return runEvent{}, false
	}
}

// parseTranscript reads the session JSONL and returns up to max compact events
// (most recent last). Missing/partial files yield an empty slice.
func parseTranscript(path string, max int) []runEvent {
	f, err := os.Open(path)
	if err != nil {
		return []runEvent{}
	}
	defer f.Close()

	events := make([]runEvent, 0, 64)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024) // tolerate long JSONL lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if ev, ok := transcriptLineToEvent(line); ok {
			events = append(events, ev)
		}
	}
	if max > 0 && len(events) > max {
		events = events[len(events)-max:]
	}
	return events
}

func (h *taskWebHandler) handleRun(w http.ResponseWriter, r *http.Request) {
	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid JSON body: %v", err))
		return
	}
	if len(req.Tasks) == 0 {
		writeWebError(w, http.StatusBadRequest, "invalid_input", "no tasks selected")
		return
	}
	permMode := req.PermissionMode
	if permMode == "" {
		permMode = "bypassPermissions"
	}
	if !validPermissionModes[permMode] {
		writeWebError(w, http.StatusBadRequest, "invalid_input",
			fmt.Sprintf("invalid permission mode '%s'. Use: default, acceptEdits, plan, bypassPermissions", permMode))
		return
	}

	// Resolve refs + flip to in_progress under the file mutex; released before spawn.
	h.mu.Lock()
	tasks := make([]*Task, 0, len(req.Tasks))
	summaries := make([]runTaskSummary, 0, len(req.Tasks))
	for _, ref := range req.Tasks {
		task, err := resolveTaskRef(h.reg, ref.Project, ref.ID)
		if err != nil {
			h.mu.Unlock()
			status, code := webErrorStatus(err)
			writeWebError(w, status, code, err.Error())
			return
		}
		tasks = append(tasks, task)
		summaries = append(summaries, runTaskSummary{
			Project: task.ProjectName,
			ID:      task.ID,
			Name:    task.Name,
			Subject: task.Subject,
		})
	}
	for _, task := range tasks {
		if task.Status != "completed" {
			if err := setTaskStatus(task, "in_progress"); err == nil {
				_ = saveTask(task)
			}
		}
	}
	h.mu.Unlock()

	prompt := buildGoalPrompt(tasks)

	sessionID, err := newSessionID()
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", fmt.Sprintf("failed to generate session id: %v", err))
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	transcriptPath, _ := sessionTranscriptPath(cwd, sessionID)

	cmd := exec.Command(h.claudeBin,
		"-p",
		"--session-id", sessionID,
		"--permission-mode", permMode,
		prompt,
	)
	cmd.Dir = cwd
	// Detach into its own process group so Ctrl+C on the agmd server does not
	// also terminate an in-flight run.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logPath := filepath.Join(os.TempDir(), "agmd-run-"+sessionID+".log")
	logFile, lerr := os.Create(logPath)
	if lerr != nil {
		logFile = nil
		logPath = ""
	}
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		writeWebError(w, http.StatusInternalServerError, "internal",
			fmt.Sprintf("failed to start claude: %v (is the 'claude' CLI installed and on PATH?)", err))
		return
	}

	info := &runInfo{
		SessionID:      sessionID,
		TranscriptPath: transcriptPath,
		LogPath:        logPath,
		Cwd:            cwd,
		PermissionMode: permMode,
		ResumeHint:     "claude --resume " + sessionID,
		Prompt:         prompt,
		Tasks:          summaries,
		StartedAt:      time.Now().UTC().Format(time.RFC3339),
		cmd:            cmd,
	}
	h.runMu.Lock()
	h.runs[sessionID] = info
	h.runMu.Unlock()

	go func() {
		werr := cmd.Wait()
		if logFile != nil {
			logFile.Close()
		}
		h.runMu.Lock()
		info.Done = true
		if werr != nil {
			info.ExitError = werr.Error()
		}
		h.runMu.Unlock()
	}()

	writeWebJSON(w, http.StatusOK, info)
}

// runInfoToMap renders the shared (event-free) view of a run.
func runInfoToMap(info *runInfo) map[string]interface{} {
	return map[string]interface{}{
		"sessionId":      info.SessionID,
		"transcriptPath": info.TranscriptPath,
		"logPath":        info.LogPath,
		"cwd":            info.Cwd,
		"permissionMode": info.PermissionMode,
		"resumeHint":     info.ResumeHint,
		"tasks":          info.Tasks,
		"startedAt":      info.StartedAt,
		"done":           info.Done,
		"exitError":      info.ExitError,
		"running":        !info.Done,
	}
}

func (h *taskWebHandler) handleRunStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	h.runMu.Lock()
	info, ok := h.runs[id]
	var snapshot runInfo
	if ok {
		snapshot = *info
	}
	h.runMu.Unlock()

	if !ok {
		writeWebError(w, http.StatusNotFound, "not_found", "no run with id "+id)
		return
	}

	path := locateTranscript(snapshot.TranscriptPath, id)
	events := parseTranscript(path, 200)

	resp := runInfoToMap(&snapshot)
	resp["events"] = events
	writeWebJSON(w, http.StatusOK, resp)
}

// handleRunList returns every tracked run (most recent first), without events.
func (h *taskWebHandler) handleRunList(w http.ResponseWriter, r *http.Request) {
	h.runMu.Lock()
	list := make([]map[string]interface{}, 0, len(h.runs))
	for _, info := range h.runs {
		list = append(list, runInfoToMap(info))
	}
	h.runMu.Unlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i]["startedAt"].(string) > list[j]["startedAt"].(string)
	})
	writeWebJSON(w, http.StatusOK, map[string]interface{}{"runs": list})
}

// --- server lifecycle ---

func runTaskWeb(cmd *cobra.Command, args []string) error {
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Best-effort current project; may be empty when run outside a repo.
	currentProject := ""
	if name, err := getProjectName(); err == nil {
		currentProject = name
	}

	handler := NewTaskWebHandler(reg, currentProject)

	addr := net.JoinHostPort(taskWebHost, strconv.Itoa(taskWebPort))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	url := fmt.Sprintf("http://%s", listener.Addr().String())
	server := &http.Server{Handler: handler}

	green := color.New(color.FgGreen).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()
	fmt.Printf("%s agmd task web serving at %s\n", green("✓"), url)
	if currentProject != "" {
		fmt.Printf("  %s %s\n", dim("current project:"), currentProject)
	}
	fmt.Printf("  %s\n", dim("press Ctrl+C to stop"))

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	if !taskWebNoOpen {
		if err := openBrowser(url); err != nil {
			fmt.Printf("  (could not open browser automatically: %v)\n", err)
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	case <-sigCh:
		fmt.Printf("\n%s shutting down…\n", dim("·"))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}

// openBrowser opens url in the default browser. Failure is non-fatal.
func openBrowser(url string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{url}
	case "windows":
		name, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		name, args = "xdg-open", []string{url}
	}
	return exec.Command(name, args...).Start()
}
