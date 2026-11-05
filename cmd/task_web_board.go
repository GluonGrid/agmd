package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Board endpoints beyond the basic state/patch/comment ones: creating and
// deleting tasks, releasing claims, and a cheap change feed for live updates.

// localOnly rejects requests that a web page on another site could forge
// against the board: a non-IP Host (DNS rebinding) or a foreign Origin.
func localOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		host = strings.Trim(host, "[]")
		if host != "localhost" && net.ParseIP(host) == nil {
			writeWebError(w, http.StatusForbidden, "forbidden", "the board only answers on localhost or an IP address")
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
			writeWebError(w, http.StatusForbidden, "forbidden", "cross-origin requests are not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// staticHandler serves the embedded board UI.
func staticHandler() http.Handler {
	sub, err := fs.Sub(webuiFS, "webui")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

// projectParam returns the {project} path value, answering 400 when it is not
// a safe project name.
func projectParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	project := r.PathValue("project")
	if err := validateName("project", project); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return "", false
	}
	return project, true
}

type createTaskBody struct {
	Project   string       `json:"project"`
	Name      string       `json:"name"`
	Subject   string       `json:"subject"`
	Content   string       `json:"content"`
	Priority  *webPriority `json:"priority"`
	Type      string       `json:"type"`
	Feature   string       `json:"feature"`
	DependsOn []string     `json:"depends_on"`
}

func (h *taskWebHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body createTaskBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("invalid JSON body: %v", err))
		return
	}
	project := strings.TrimSpace(body.Project)
	if err := validateName("project", project); err != nil {
		writeWebError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	if strings.TrimSpace(body.Subject) == "" && strings.TrimSpace(body.Name) == "" {
		writeWebError(w, http.StatusBadRequest, "invalid_input", "a task needs a title")
		return
	}
	priority := 2
	if body.Priority != nil {
		priority = body.Priority.value
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	unlock, err := lockProject(h.reg, project)
	if err != nil {
		writeWebError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	defer unlock()

	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = uniqueTaskName(h.reg, project, body.Subject)
	}
	task, err := createTask(h.reg, project, newTaskSpec{
		Name:      name,
		Subject:   body.Subject,
		Content:   body.Content,
		Priority:  priority,
		Type:      strings.ToLower(strings.TrimSpace(body.Type)),
		Feature:   body.Feature,
		DependsOn: body.DependsOn,
	})
	if err != nil {
		status, code := webErrorStatus(err)
		writeWebError(w, status, code, err.Error())
		return
	}
	h.writeTaskItem(w, project, task)
}

func (h *taskWebHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	h.mutateTask(w, r, func(project string, task *Task) bool {
		if err := deleteTask(h.reg, project, task); err != nil {
			status, code := webErrorStatus(err)
			writeWebError(w, status, code, err.Error())
			return false
		}
		writeWebJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": task.ID})
		return false
	})
}

func (h *taskWebHandler) handleRelease(w http.ResponseWriter, r *http.Request) {
	h.mutateTask(w, r, func(project string, task *Task) bool {
		if err := releaseClaim(h.reg, project, task); err != nil {
			status, code := webErrorStatus(err)
			writeWebError(w, status, code, err.Error())
			return false
		}
		return true
	})
}

// mutateTask resolves {project}/{id} under both locks and runs fn. When fn
// returns true the updated task is written as the response.
func (h *taskWebHandler) mutateTask(w http.ResponseWriter, r *http.Request, fn func(project string, task *Task) bool) {
	project, ok := projectParam(w, r)
	if !ok {
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

	task, err := resolveTaskRef(h.reg, project, r.PathValue("id"))
	if err != nil {
		status, code := webErrorStatus(err)
		writeWebError(w, status, code, err.Error())
		return
	}
	if fn(project, task) {
		h.writeTaskItem(w, project, task)
	}
}

// handleChanges returns a fingerprint of every task, claim, comment and
// project file. The board polls it and reloads state only when it changes, so
// edits made by agents through the CLI show up within a poll interval.
func (h *taskWebHandler) handleChanges(w http.ResponseWriter, r *http.Request) {
	sum := sha256.New()
	root := filepath.Join(h.reg.BasePath, "task")
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { //nolint:errcheck
		if err != nil {
			return nil
		}
		// Skip the lock file and in-flight temp files from atomic writes.
		if strings.HasPrefix(d.Name(), ".") && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Fprintf(sum, "%s\x00%d\x00%d\n", path, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if _, err := os.Stat(root); err != nil {
		sum.Write([]byte("missing"))
	}
	writeWebJSON(w, http.StatusOK, map[string]string{"version": hex.EncodeToString(sum.Sum(nil))[:16]})
}
