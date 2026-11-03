package cmd

import (
	"context"
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

//go:embed webui
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
	Emoji  string    `json:"emoji,omitempty"`
	Counts webCounts `json:"counts"`
}

type webStatePayload struct {
	CurrentProject string `json:"currentProject"`
	// ClaimTTLSeconds is how long a claim survives without a heartbeat
	// (0 = claims never expire); the board uses it to flag stale claims.
	ClaimTTLSeconds int64            `json:"claimTtlSeconds"`
	Projects        []webProjectInfo `json:"projects"`
	Tasks           []taskJSONItem   `json:"tasks"`
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

	ttl, err := claimTTL()
	if err != nil {
		return state, err
	}
	state.ClaimTTLSeconds = int64(ttl.Seconds())

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

		meta, err := loadProjectMeta(reg, name)
		if err != nil {
			return state, err
		}
		state.Projects = append(state.Projects, webProjectInfo{
			Name:  name,
			Emoji: meta.Emoji,
			Counts: webCounts{
				Total:      payload.Counts.Total,
				Ready:      payload.Counts.Ready,
				Blocked:    payload.Counts.Blocked,
				InProgress: payload.Counts.InProgress,
				Completed:  payload.Counts.Completed,
			},
		})
		for i := range payload.Tasks {
			if err := attachComments(reg, &payload.Tasks[i]); err != nil {
				return state, err
			}
		}
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
	mux.HandleFunc("GET /api/changes", h.handleChanges)
	mux.HandleFunc("POST /api/tasks", h.handleCreate)
	mux.HandleFunc("PATCH /api/tasks/{project}/{id}", h.handlePatch)
	mux.HandleFunc("DELETE /api/tasks/{project}/{id}", h.handleDelete)
	mux.HandleFunc("DELETE /api/tasks/{project}/{id}/claim", h.handleRelease)
	mux.HandleFunc("POST /api/tasks/{project}/{id}/comments", h.handleComment)
	mux.HandleFunc("POST /api/run", h.handleRun)
	mux.HandleFunc("GET /api/runs", h.handleRunList)
	mux.HandleFunc("GET /api/run/{id}", h.handleRunStatus)
	mux.Handle("GET /", staticHandler())
	return localOnly(mux)
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
	Status    *string      `json:"status"`
	Priority  *webPriority `json:"priority"`
	Type      *string      `json:"type"`
	Feature   *string      `json:"feature"`
	Subject   *string      `json:"subject"`
	Content   *string      `json:"content"`
	DependsOn *[]string    `json:"depends_on"`
}

func (h *taskWebHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	project, ok := projectParam(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")

	var body taskPatchBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid JSON body: %v", err))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	unlock, err := lockProject(h.reg, project)
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	defer unlock()

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
	if body.DependsOn != nil {
		deps, err := resolveDependencyNames(h.reg, project, task.Name, *body.DependsOn)
		if err != nil {
			status, code := webErrorStatus(err)
			writeWebError(w, status, code, err.Error())
			return
		}
		task.DependsOn = deps
	}

	if err := saveTask(task); err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", fmt.Sprintf("failed to save task: %v", err))
		return
	}

	h.writeTaskItem(w, project, task)
}

type commentBody struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

func (h *taskWebHandler) handleComment(w http.ResponseWriter, r *http.Request) {
	project, ok := projectParam(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")

	var body commentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid JSON body: %v", err))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	unlock, err := lockProject(h.reg, project)
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	defer unlock()

	task, err := resolveTaskRef(h.reg, project, id)
	if err != nil {
		status, code := webErrorStatus(err)
		writeWebError(w, status, code, err.Error())
		return
	}
	if _, err := addTaskComment(h.reg, project, task.ID, body.Author, body.Body); err != nil {
		status, code := webErrorStatus(err)
		writeWebError(w, status, code, err.Error())
		return
	}
	h.writeTaskItem(w, project, task)
}

// writeTaskItem responds with the task as the board shows it. It reloads the
// project so computed_status reflects sibling task states.
func (h *taskWebHandler) writeTaskItem(w http.ResponseWriter, project string, task *Task) {
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
	if err := attachComments(h.reg, &item); err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
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
		case "conflict":
			return http.StatusConflict, "conflict"
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

// --- server lifecycle ---

func runTaskWeb(cmd *cobra.Command, args []string) error {
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
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
