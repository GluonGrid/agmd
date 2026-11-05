package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"agmd/pkg/registry"
)

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

// claimTaskForRun re-reads the task under its project lock, so a CLI agent's
// concurrent change is not overwritten, flips it to in_progress unless it is
// already completed, and claims it for workflowID unless another workflow
// holds it. Failures are ignored: the run starts regardless.
func claimTaskForRun(reg *registry.Registry, task *Task, workflowID string) {
	unlock, err := lockProject(reg, task.ProjectName)
	if err != nil {
		return
	}
	defer unlock()
	current, err := loadTask(task.FilePath)
	if err != nil || current.Status == "completed" {
		return
	}
	if setTaskStatus(current, "in_progress") != nil || saveTask(current) != nil {
		return
	}
	assignments, err := loadTaskAssignments(reg, task.ProjectName)
	if err != nil {
		return
	}
	if _, held := assignments[current.ID]; held {
		return
	}
	assignments[current.ID] = taskAssignment{
		TaskID:     current.ID,
		TaskName:   current.Name,
		WorkflowID: workflowID,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	_ = saveTaskAssignments(reg, task.ProjectName, assignments)
}

// heartbeatRun keeps a run's claims alive until done is closed.
func heartbeatRun(reg *registry.Registry, projects []string, workflowID string, done <-chan struct{}) {
	ttl, err := claimTTL()
	if err != nil || ttl <= 0 {
		return
	}
	interval := ttl / 3
	if interval > 5*time.Minute {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			for _, project := range projects {
				_ = heartbeatWorkflow(reg, project, workflowID)
			}
		}
	}
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

	sessionID, err := newSessionID()
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", fmt.Sprintf("failed to generate session id: %v", err))
		return
	}
	workflowID := "web-" + sessionID

	// Resolve refs + claim under the file mutex; released before spawn.
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
	runProjects := []string{}
	for _, task := range tasks {
		claimTaskForRun(h.reg, task, workflowID)
		if !slices.Contains(runProjects, task.ProjectName) {
			runProjects = append(runProjects, task.ProjectName)
		}
	}
	h.mu.Unlock()

	prompt := buildGoalPrompt(tasks)

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

	runDone := make(chan struct{})
	go heartbeatRun(h.reg, runProjects, workflowID, runDone)
	go func() {
		werr := cmd.Wait()
		close(runDone)
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
