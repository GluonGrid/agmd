package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"agmd/pkg/registry"

	"gopkg.in/yaml.v3"
)

// Task storage operations shared by the CLI and the web board. Callers that
// change existing files must hold the project lock (see task_lock.go).

// validateName rejects project and task names that are unsafe as a single
// path segment under the data directory.
func validateName(kind, name string) error {
	switch {
	case strings.TrimSpace(name) == "":
		return newTaskError("invalid_input", "%s name is empty", kind)
	case name == "." || name == ".." || strings.HasPrefix(name, "."):
		return newTaskError("invalid_input", "invalid %s name '%s': must not start with '.'", kind, name)
	case strings.ContainsAny(name, `/\`):
		return newTaskError("invalid_input", "invalid %s name '%s': must not contain '/' or '\\'", kind, name)
	}
	return nil
}

// encodeTask renders a task as YAML frontmatter followed by its content.
func encodeTask(task *Task) ([]byte, error) {
	frontmatter := struct {
		ID        string   `yaml:"id,omitempty"`
		Subject   string   `yaml:"subject"`
		Status    string   `yaml:"status"`
		Priority  int      `yaml:"priority"`
		Type      string   `yaml:"type,omitempty"`
		Feature   string   `yaml:"feature,omitempty"`
		DependsOn []string `yaml:"depends_on"`
	}{
		ID:        task.ID,
		Subject:   task.Subject,
		Status:    task.Status,
		Priority:  task.Priority,
		Type:      task.Type,
		Feature:   task.Feature,
		DependsOn: task.DependsOn,
	}
	if frontmatter.DependsOn == nil {
		frontmatter.DependsOn = []string{}
	}
	if frontmatter.Type == "task" {
		frontmatter.Type = ""
	}
	fmBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("---\n%s---\n\n%s\n", string(fmBytes), task.Content)), nil
}

type newTaskSpec struct {
	Name      string
	Subject   string // defaults to the name, title-cased
	Content   string
	Priority  int
	Type      string
	Feature   string
	DependsOn []string // task names or ids in the same project
}

// createTask validates spec and writes a new pending task. It fails with
// "invalid_input" if a task of that name already exists.
func createTask(reg *registry.Registry, projectName string, spec newTaskSpec) (*Task, error) {
	if err := validateName("project", projectName); err != nil {
		return nil, err
	}
	if err := validateName("task", spec.Name); err != nil {
		return nil, err
	}
	if spec.Priority < 0 || spec.Priority > 4 {
		return nil, newTaskError("invalid_input", "invalid priority %d. Use 0-4 (0=critical, 4=backlog)", spec.Priority)
	}
	if spec.Type == "" {
		spec.Type = "task"
	}
	if !validTaskTypes[spec.Type] {
		return nil, newTaskError("invalid_input", "invalid type '%s'. Use: bug, feature, task, or chore", spec.Type)
	}
	dependsOn, err := resolveDependencyNames(reg, projectName, spec.Name, spec.DependsOn)
	if err != nil {
		return nil, err
	}
	subject := strings.TrimSpace(spec.Subject)
	if subject == "" {
		subject = strings.Title(strings.ReplaceAll(spec.Name, "-", " "))
	}

	taskDir := getTaskDir(reg, projectName)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create task directory: %w", err)
	}
	task := &Task{
		Name:        spec.Name,
		ID:          computeTaskID(projectName, spec.Name),
		Subject:     subject,
		Status:      "pending",
		Priority:    spec.Priority,
		Type:        spec.Type,
		Feature:     strings.TrimSpace(spec.Feature),
		DependsOn:   dependsOn,
		Content:     strings.TrimSpace(spec.Content),
		FilePath:    filepath.Join(taskDir, spec.Name+".md"),
		ProjectName: projectName,
	}
	data, err := encodeTask(task)
	if err != nil {
		return nil, err
	}
	if err := writeFileExclusive(task.FilePath, data, 0o644); err != nil {
		if os.IsExist(err) {
			return nil, newTaskError("invalid_input", "task:%s already exists in project '%s'", spec.Name, projectName)
		}
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

// resolveDependencyNames maps dependency refs to task names, de-duplicated.
func resolveDependencyNames(reg *registry.Registry, projectName, taskName string, refs []string) ([]string, error) {
	names := []string{}
	seen := map[string]bool{}
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		dep, err := resolveTaskRef(reg, projectName, ref)
		if err != nil {
			return nil, newTaskError("not_found", "dependency task '%s' not found in project '%s'", ref, projectName)
		}
		if dep.Name == taskName {
			return nil, newTaskError("invalid_input", "task '%s' cannot depend on itself", taskName)
		}
		if !seen[dep.Name] {
			seen[dep.Name] = true
			names = append(names, dep.Name)
		}
	}
	return names, nil
}

var slugUnsafe = regexp.MustCompile(`[^a-z0-9]+`)

// uniqueTaskName derives a file-safe task name from a subject, adding a
// numeric suffix when the name is taken.
func uniqueTaskName(reg *registry.Registry, projectName, subject string) string {
	base := strings.Trim(slugUnsafe.ReplaceAllString(strings.ToLower(subject), "-"), "-")
	if len(base) > 48 {
		base = strings.TrimRight(base[:48], "-")
	}
	if base == "" {
		base = "task"
	}
	name := base
	for i := 2; ; i++ {
		if _, err := os.Stat(getTaskPath(reg, projectName, name)); os.IsNotExist(err) {
			return name
		}
		name = fmt.Sprintf("%s-%d", base, i)
	}
}

// deleteTask removes a task with its claim and comments. The caller must hold
// the project lock.
func deleteTask(reg *registry.Registry, projectName string, task *Task) error {
	if err := os.Remove(task.FilePath); err != nil {
		if os.IsNotExist(err) {
			return newTaskError("not_found", "task '%s' was already deleted from project '%s'", task.Name, projectName)
		}
		return fmt.Errorf("failed to delete task: %w", err)
	}
	if err := dropTaskAssignments(reg, projectName, task.ID); err != nil {
		return fmt.Errorf("failed to update assignments: %w", err)
	}
	if err := removeTaskComments(reg, projectName, task.ID); err != nil {
		return fmt.Errorf("failed to remove comments: %w", err)
	}
	return nil
}

// releaseClaim removes a task's claim and puts it back to pending if it was
// in progress. The caller must hold the project lock.
func releaseClaim(reg *registry.Registry, projectName string, task *Task) error {
	assignments, err := loadTaskAssignments(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}
	if _, held := assignments[task.ID]; !held {
		return newTaskError("not_found", "task '%s' is not assigned in project '%s'", task.Name, projectName)
	}
	delete(assignments, task.ID)
	if err := saveTaskAssignments(reg, projectName, assignments); err != nil {
		return fmt.Errorf("failed to save assignments: %w", err)
	}
	return releaseTask(task)
}
