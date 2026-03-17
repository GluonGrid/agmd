package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Task represents a task with its metadata
type Task struct {
	Name        string   `yaml:"-"`
	ID          string   `yaml:"id,omitempty"`
	Subject     string   `yaml:"subject"`
	Status      string   `yaml:"status"`
	Priority    int      `yaml:"priority,omitempty"` // 0-4 (P0=critical, P4=backlog)
	Type        string   `yaml:"type,omitempty"`     // bug, feature, task, chore
	Feature     string   `yaml:"feature,omitempty"`
	DependsOn   []string `yaml:"depends_on"`
	Content     string   `yaml:"-"`
	FilePath    string   `yaml:"-"`
	ProjectName string   `yaml:"-"`
}

// Valid task types
var validTaskTypes = map[string]bool{
	"bug":     true,
	"feature": true,
	"task":    true,
	"chore":   true,
}

// ComputedStatus represents the computed status of a task
type ComputedStatus string

const (
	StatusReady      ComputedStatus = "ready"
	StatusBlocked    ComputedStatus = "blocked"
	StatusInProgress ComputedStatus = "in_progress"
	StatusCompleted  ComputedStatus = "completed"
)

// Shared flags for task subcommands
var taskProject string
var taskRepoPath string
var taskJSON bool
var taskFeature string
var taskAll bool
var taskForce bool
var taskContent string
var taskBlockedBy string
var taskNoEditor bool
var taskRaw bool
var taskStatus string
var taskTree bool
var taskPriority int
var taskType string
var taskFilterPriority string
var taskFilterType string

type codedTaskError struct {
	Code     string
	Message  string
	ExitCode int
}

func (e *codedTaskError) Error() string {
	return e.Message
}

func (e *codedTaskError) ErrorCode() string {
	return e.Code
}

func (e *codedTaskError) ExitStatus() int {
	if e.ExitCode <= 0 {
		return 1
	}
	return e.ExitCode
}

func newTaskError(code, format string, args ...interface{}) error {
	exitCode := 1
	switch code {
	case "invalid_input":
		exitCode = 2
	case "not_found":
		exitCode = 3
	case "project_mismatch":
		exitCode = 4
	case "invalid_transition":
		exitCode = 5
	}
	return &codedTaskError{
		Code:     code,
		Message:  fmt.Sprintf(format, args...),
		ExitCode: exitCode,
	}
}

func printJSON(payload interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(payload)
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage project tasks",
	Long: `Manage project tasks with dependencies, priorities, and status tracking.

Tasks are organized by project (derived from repository identity).
In Git worktrees, agmd resolves to the canonical repository project.
Each task has a priority (P0-P4) and type (bug/feature/task/chore).
Tasks are auto-sorted by priority then status.

Priority levels:
  P0  critical    Urgent issues requiring immediate attention
  P1  high        Important tasks to address soon
  P2  medium      Default priority for normal tasks
  P3  low         Can be done when time permits
  P4  backlog     Ideas and future work

Task types:
  bug      Defect to fix
  feature  New functionality
  task     General task (default)
  chore    Maintenance work

Subcommands:
  list        List tasks (sorted by priority, filterable)
  new         Create a new task with priority/type
  edit        Edit task priority or type
  show        Show task content
  delete      Delete a task
  status      Update task status
  blocked-by  Add a dependency
  unblock     Remove a dependency

Examples:
  agmd task list                                    # List all tasks
  agmd task list --type bug --priority 0           # Critical bugs only
  agmd task new fix-auth -t bug -p 0 --content "Fix login"
  agmd task new setup-db --content "Set up DB"      # Create task (P2, task)
  agmd task edit setup-db -p 0 -t bug              # Change priority and type
  agmd task status setup-db completed               # Update status
  agmd task blocked-by create-api setup-db          # Add dependency`,
}

var taskListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tasks for current project",
	Long: `List tasks for the current project, auto-sorted by priority then status.

Tasks are sorted: P0 → P4, then ready → in_progress → blocked → completed.
Completed tasks are hidden by default (use --all to show).

Filters:
  --feature     Filter by feature/session
  --status      Filter by computed status (ready, blocked, in_progress, completed)
  --priority    Filter by priority (0-4 or P0-P4)
  --type        Filter by type (bug, feature, task, chore)

Examples:
  agmd task list                            # List active tasks
  agmd task ls                              # Same (alias)
  agmd task list --all                      # Include completed tasks
  agmd task list --feature auth             # Only tasks for "auth" feature
  agmd task list --status ready             # Only ready tasks
  agmd task list --priority 0               # Only critical tasks (P0)
  agmd task list --type bug                 # Only bugs
  agmd task list --type bug --priority 0    # Critical bugs only
  agmd task list --tree                     # Show dependency tree
  agmd task list --project myproj           # List tasks for specific project`,
	RunE: runTaskList,
}

var taskNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new task",
	Long: `Create a new task for the current project.

Priority levels: 0=critical, 1=high, 2=medium (default), 3=low, 4=backlog
Task types: bug, feature, task (default), chore

Examples:
  agmd task new setup-db --content "Set up database"
  agmd task new fix-auth -t bug -p 0 --content "Critical auth bug"
  agmd task new add-dark-mode -t feature -p 2 --content "Add dark mode"
  agmd task new create-api --content "Create API" --blocked-by "setup-db"
  agmd task new setup-db --feature auth --content "Set up auth DB"
  echo "Task description" | agmd task new setup-db`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskNew,
}

var taskShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show task content",
	Long: `Display the content of a task.

Examples:
  agmd task show setup-db                       # Show single task
  agmd task show setup-db --raw                 # Include frontmatter
  agmd task show --all                          # Show all tasks for project
  agmd task show --all --feature auth           # Show all tasks for feature
  agmd task show --all --project myproj         # Show all tasks for specific project`,
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskShow,
}

var taskDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a task",
	Long: `Delete a task from the current project.

Examples:
  agmd task delete setup-db             # Delete with confirmation
  agmd task rm setup-db                 # Same (alias)
  agmd task delete setup-db --force     # Skip confirmation
  agmd task delete setup-db --project x # Delete from specific project`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskDelete,
}

var taskStatusCmd = &cobra.Command{
	Use:   "status <task-name> <status>",
	Short: "Update task status",
	Long: `Update the status of a task.

Valid statuses: pending, in_progress, completed

Examples:
  agmd task status setup-db pending
  agmd task status setup-db in_progress
  agmd task status setup-db completed`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskStatus,
	RunE:              runTaskStatus,
}

var taskBlockedByCmd = &cobra.Command{
	Use:   "blocked-by <task-name> <dependency>",
	Short: "Add a dependency to a task",
	Long: `Add a dependency to a task.

This makes <task-name> depend on <dependency>.

Examples:
  agmd task blocked-by create-api setup-db    # create-api depends on setup-db`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskDependency,
	RunE:              runTaskBlockedBy,
}

var taskUnblockCmd = &cobra.Command{
	Use:   "unblock <task-name> <dependency>",
	Short: "Remove a dependency from a task",
	Long: `Remove a dependency from a task.

Examples:
  agmd task unblock create-api setup-db    # Remove setup-db dependency from create-api`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskDependency,
	RunE:              runTaskUnblock,
}

var taskEditPriority int = -1 // -1 means not set
var taskEditType string

var taskCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Delete all completed tasks",
	Long: `Delete all completed tasks for the current project.

This permanently removes all tasks with status "completed".
Use --force to skip the confirmation prompt.

Examples:
  agmd task clean                # Delete completed tasks (with confirmation)
  agmd task clean --force        # Skip confirmation`,
	RunE: runTaskClean,
}

var taskEditCmd = &cobra.Command{
	Use:   "edit <task-name>",
	Short: "Edit task priority or type",
	Long: `Edit a task's priority or type.

Use -p/--priority to change priority (0-4 or P0-P4).
Use -t/--type to change type (bug, feature, task, chore).

Examples:
  agmd task edit setup-db -p 0              # Set to critical priority
  agmd task edit setup-db -t bug            # Change type to bug
  agmd task edit setup-db -p 1 -t feature   # Change both`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskEdit,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskNewCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskCleanCmd)
	taskCmd.AddCommand(taskBlockedByCmd)
	taskCmd.AddCommand(taskUnblockCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.PersistentFlags().BoolVar(&taskJSON, "json", false, "Output machine-readable JSON")

	// Add --project/--repo-path and --feature flags to subcommands that need them
	taskListCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskListCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskListCmd.Flags().StringVar(&taskFeature, "feature", "", "Filter tasks by feature")
	taskListCmd.Flags().BoolVarP(&taskAll, "all", "a", false, "Include completed tasks")
	taskListCmd.Flags().StringVar(&taskStatus, "status", "", "Filter by computed status (ready, blocked, in_progress, completed)")
	taskListCmd.RegisterFlagCompletionFunc("status", completeListStatusFlag)
	taskListCmd.Flags().BoolVar(&taskTree, "tree", false, "Show dependency tree visualization")
	taskListCmd.Flags().StringVarP(&taskFilterPriority, "priority", "p", "", "Filter by priority (0-4 or P0-P4)")
	taskListCmd.Flags().StringVarP(&taskFilterType, "type", "t", "", "Filter by type (bug, feature, task, chore)")
	taskListCmd.RegisterFlagCompletionFunc("priority", completePriorityFlag)
	taskListCmd.RegisterFlagCompletionFunc("type", completeTypeFlag)

	taskNewCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskNewCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskNewCmd.Flags().StringVar(&taskFeature, "feature", "", "Feature/session name for this task")
	taskNewCmd.Flags().StringVar(&taskContent, "content", "", "Task content/description")
	taskNewCmd.Flags().StringVar(&taskBlockedBy, "blocked-by", "", "Comma-separated list of task dependencies")
	taskNewCmd.Flags().BoolVar(&taskNoEditor, "no-editor", false, "Don't open editor after creating")
	taskNewCmd.Flags().IntVarP(&taskPriority, "priority", "p", 2, "Priority (0=critical, 1=high, 2=medium, 3=low, 4=backlog)")
	taskNewCmd.Flags().StringVarP(&taskType, "type", "t", "task", "Type (bug, feature, task, chore)")
	taskNewCmd.RegisterFlagCompletionFunc("priority", completePriorityFlag)
	taskNewCmd.RegisterFlagCompletionFunc("type", completeTypeFlag)

	taskShowCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskShowCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskShowCmd.Flags().StringVar(&taskFeature, "feature", "", "Filter tasks by feature")
	taskShowCmd.Flags().BoolVarP(&taskAll, "all", "a", false, "Show all tasks for project")
	taskShowCmd.Flags().BoolVar(&taskRaw, "raw", false, "Include frontmatter in output")

	taskDeleteCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskDeleteCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskDeleteCmd.Flags().BoolVarP(&taskForce, "force", "f", false, "Skip confirmation prompt")

	taskStatusCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskStatusCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskBlockedByCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskBlockedByCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskUnblockCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskUnblockCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")

	taskEditCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskEditCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskEditCmd.Flags().IntVarP(&taskEditPriority, "priority", "p", -1, "New priority (0=critical, 1=high, 2=medium, 3=low, 4=backlog)")
	taskEditCmd.Flags().StringVarP(&taskEditType, "type", "t", "", "New type (bug, feature, task, chore)")
	taskEditCmd.RegisterFlagCompletionFunc("priority", completePriorityFlag)
	taskEditCmd.RegisterFlagCompletionFunc("type", completeTypeFlag)

	taskCleanCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskCleanCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskCleanCmd.Flags().BoolVarP(&taskForce, "force", "f", false, "Skip confirmation prompt")
}

// getProjectName returns the project name from explicit selector or repository path.
func getProjectName() (string, error) {
	if taskProject != "" {
		return taskProject, nil
	}

	targetPath := taskRepoPath
	if targetPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		targetPath = cwd
	}

	canonicalRepoPath, err := resolveCanonicalRepoPath(targetPath)
	if err != nil {
		absPath, absErr := filepath.Abs(targetPath)
		if absErr != nil {
			return "", fmt.Errorf("failed to resolve project path '%s': %w", targetPath, absErr)
		}
		return filepath.Base(absPath), nil
	}
	return filepath.Base(canonicalRepoPath), nil
}

// resolveCanonicalRepoPath resolves path to canonical repo root.
// In worktrees, this maps to the main repo path from git common dir.
func resolveCanonicalRepoPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	commonDir, err := gitRevParse(absPath, "--path-format=absolute", "--git-common-dir")
	if err == nil && commonDir != "" {
		normalized := filepath.Clean(commonDir)
		if filepath.Base(normalized) == ".git" {
			return filepath.Dir(normalized), nil
		}
	}

	topLevel, topErr := gitRevParse(absPath, "--path-format=absolute", "--show-toplevel")
	if topErr == nil && topLevel != "" {
		return filepath.Clean(topLevel), nil
	}

	if err != nil {
		return "", err
	}
	return "", topErr
}

func gitRevParse(path string, args ...string) (string, error) {
	gitArgs := append([]string{"-C", path, "rev-parse"}, args...)
	out, err := exec.Command("git", gitArgs...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git rev-parse failed for '%s': %s", path, msg)
	}
	return strings.TrimSpace(string(out)), nil
}

func computeTaskID(projectName, taskName string) string {
	sum := sha256.Sum256([]byte(projectName + "::" + taskName))
	return "tsk_" + hex.EncodeToString(sum[:8])
}

func ensureTaskIdentity(task *Task, projectName string) {
	task.ProjectName = projectName
	if task.ID == "" {
		task.ID = computeTaskID(projectName, task.Name)
	}
}

type taskJSONItem struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Project             string   `json:"project"`
	Subject             string   `json:"subject"`
	Status              string   `json:"status"`
	ComputedStatus      string   `json:"computed_status"`
	Priority            int      `json:"priority"`
	Type                string   `json:"type"`
	Feature             string   `json:"feature,omitempty"`
	DependsOn           []string `json:"depends_on"`
	PendingDependencies []string `json:"pending_dependencies,omitempty"`
	Content             string   `json:"content,omitempty"`
}

type taskListJSONPayload struct {
	Project string         `json:"project"`
	Feature string         `json:"feature,omitempty"`
	Tasks   []taskJSONItem `json:"tasks"`
	Counts  struct {
		Total      int `json:"total"`
		Ready      int `json:"ready"`
		Blocked    int `json:"blocked"`
		InProgress int `json:"in_progress"`
		Completed  int `json:"completed"`
	} `json:"counts"`
}

type taskMutationJSONPayload struct {
	OK      bool   `json:"ok"`
	Action  string `json:"action"`
	Project string `json:"project"`
	Task    string `json:"task,omitempty"`
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
}

func taskToJSONItem(task *Task, taskMap map[string]*Task, includeContent bool) taskJSONItem {
	taskType := task.Type
	if taskType == "" {
		taskType = "task"
	}
	status := computeTaskStatus(task, taskMap)
	item := taskJSONItem{
		ID:             task.ID,
		Name:           task.Name,
		Project:        task.ProjectName,
		Subject:        task.Subject,
		Status:         task.Status,
		ComputedStatus: string(status),
		Priority:       task.Priority,
		Type:           taskType,
		Feature:        task.Feature,
		DependsOn:      task.DependsOn,
	}
	if status == StatusBlocked {
		item.PendingDependencies = getPendingDependencies(task, taskMap)
	}
	if includeContent {
		item.Content = task.Content
	}
	return item
}

func buildTaskListJSON(projectName, feature string, tasks []*Task, taskMap map[string]*Task, includeContent bool) taskListJSONPayload {
	payload := taskListJSONPayload{
		Project: projectName,
		Feature: feature,
		Tasks:   make([]taskJSONItem, 0, len(tasks)),
	}
	for _, task := range tasks {
		status := computeTaskStatus(task, taskMap)
		switch status {
		case StatusReady:
			payload.Counts.Ready++
		case StatusBlocked:
			payload.Counts.Blocked++
		case StatusInProgress:
			payload.Counts.InProgress++
		case StatusCompleted:
			payload.Counts.Completed++
		}
		payload.Tasks = append(payload.Tasks, taskToJSONItem(task, taskMap, includeContent))
	}
	payload.Counts.Total = len(tasks)
	return payload
}

func parseTaskRef(name string) (string, string, bool) {
	if !strings.Contains(name, "/") {
		return "", name, false
	}
	parts := strings.SplitN(name, "/", 2)
	return parts[0], parts[1], true
}

// getTaskPath returns the path to a task file
func getTaskPath(reg *registry.Registry, projectName, taskName string) string {
	return filepath.Join(reg.BasePath, "task", projectName, taskName+".md")
}

// getTaskDir returns the path to a project's task directory
func getTaskDir(reg *registry.Registry, projectName string) string {
	return filepath.Join(reg.BasePath, "task", projectName)
}

// loadTask loads a task from file
func loadTask(filePath string) (*Task, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	task := &Task{
		FilePath: filePath,
		Name:     strings.TrimSuffix(filepath.Base(filePath), ".md"),
	}

	// Parse frontmatter
	frontmatter, body, err := extractTaskFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, task); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
	}

	// Set defaults
	if task.Status == "" {
		task.Status = "pending"
	}
	if task.DependsOn == nil {
		task.DependsOn = []string{}
	}

	task.Content = strings.TrimSpace(string(body))
	ensureTaskIdentity(task, filepath.Base(filepath.Dir(filePath)))
	return task, nil
}

// extractTaskFrontmatter extracts YAML frontmatter from content
func extractTaskFrontmatter(content []byte) ([]byte, []byte, error) {
	if len(content) < 4 || string(content[:4]) != "---\n" {
		return nil, content, nil
	}

	// Find closing ---
	rest := content[4:]
	idx := strings.Index(string(rest), "\n---")
	if idx == -1 {
		return nil, content, nil
	}

	frontmatter := rest[:idx]
	body := rest[idx+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	return frontmatter, body, nil
}

// saveTask saves a task to file
func saveTask(task *Task) error {
	// Build frontmatter
	frontmatter := struct {
		ID        string   `yaml:"id,omitempty"`
		Subject   string   `yaml:"subject"`
		Status    string   `yaml:"status"`
		Priority  int      `yaml:"priority,omitempty"`
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

	fmBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("---\n%s---\n\n%s\n", string(fmBytes), task.Content)
	return os.WriteFile(task.FilePath, []byte(content), 0644)
}

// loadProjectTasks loads all tasks for a project
func loadProjectTasks(reg *registry.Registry, projectName string) ([]*Task, error) {
	taskDir := getTaskDir(reg, projectName)

	entries, err := os.ReadDir(taskDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Task{}, nil
		}
		return nil, err
	}

	var tasks []*Task
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		taskPath := filepath.Join(taskDir, entry.Name())
		task, err := loadTask(taskPath)
		if err != nil {
			continue
		}
		ensureTaskIdentity(task, projectName)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// computeTaskStatus computes the effective status of a task
func computeTaskStatus(task *Task, allTasks map[string]*Task) ComputedStatus {
	// If already completed or in_progress, return as-is
	if task.Status == "completed" {
		return StatusCompleted
	}
	if task.Status == "in_progress" {
		return StatusInProgress
	}

	// Check if blocked by any pending dependencies
	for _, dep := range task.DependsOn {
		depTask, exists := allTasks[dep]
		if !exists {
			// Dependency doesn't exist, consider it blocking
			return StatusBlocked
		}
		if depTask.Status != "completed" {
			return StatusBlocked
		}
	}

	return StatusReady
}

// sortTasksByPriorityAndStatus sorts tasks: P0 -> P4, then ready -> in_progress -> blocked -> completed
func sortTasksByPriorityAndStatus(tasks []*Task, allTasks map[string]*Task) []*Task {
	// Compute status for each task
	type taskWithStatus struct {
		task   *Task
		status ComputedStatus
	}

	tasksWithStatus := make([]taskWithStatus, len(tasks))
	for i, t := range tasks {
		tasksWithStatus[i] = taskWithStatus{
			task:   t,
			status: computeTaskStatus(t, allTasks),
		}
	}

	// Sort by priority first, then by status
	statusPriority := map[ComputedStatus]int{
		StatusReady:      0,
		StatusInProgress: 1,
		StatusBlocked:    2,
		StatusCompleted:  3,
	}

	sort.SliceStable(tasksWithStatus, func(i, j int) bool {
		// First compare by priority (lower is higher priority)
		if tasksWithStatus[i].task.Priority != tasksWithStatus[j].task.Priority {
			return tasksWithStatus[i].task.Priority < tasksWithStatus[j].task.Priority
		}
		// Then by status
		return statusPriority[tasksWithStatus[i].status] < statusPriority[tasksWithStatus[j].status]
	})

	// Extract sorted tasks
	sorted := make([]*Task, len(tasks))
	for i, ts := range tasksWithStatus {
		sorted[i] = ts.task
	}
	return sorted
}

// getPendingDependencies returns the list of pending dependencies
func getPendingDependencies(task *Task, allTasks map[string]*Task) []string {
	var pending []string
	for _, dep := range task.DependsOn {
		depTask, exists := allTasks[dep]
		if !exists || depTask.Status != "completed" {
			pending = append(pending, dep)
		}
	}
	return pending
}

// filterTasksByFeature returns only tasks matching the given feature
func filterTasksByFeature(tasks []*Task, feature string) []*Task {
	var filtered []*Task
	for _, t := range tasks {
		if strings.EqualFold(t.Feature, feature) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// parsePriority parses priority from string (0-4 or P0-P4)
func parsePriority(s string) (int, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	// Handle P0-P4 format
	if len(s) == 2 && s[0] == 'P' {
		s = string(s[1])
	}
	// Parse as integer
	if len(s) == 1 && s[0] >= '0' && s[0] <= '4' {
		return int(s[0] - '0'), nil
	}
	return 0, fmt.Errorf("invalid priority '%s'. Use 0-4 or P0-P4", s)
}

// formatPriorityType returns a colored "P0:bug" style tag
// Only shows non-default values (P2 and task are defaults)
func formatPriorityType(priority int, taskType string) string {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	// Normalize type
	if taskType == "" {
		taskType = "task"
	}

	// Check what's non-default
	hasPriority := priority != 2
	hasType := taskType != "task"

	if !hasPriority && !hasType {
		return ""
	}

	// Build the tag
	var tag string
	if hasPriority && hasType {
		tag = fmt.Sprintf("P%d:%s", priority, taskType)
	} else if hasPriority {
		tag = fmt.Sprintf("P%d", priority)
	} else {
		tag = taskType
	}

	// Color based on priority (or type if no priority)
	switch {
	case priority == 0:
		return red(tag)
	case priority == 1:
		return yellow(tag)
	case priority == 3 || priority == 4:
		return dim(tag)
	case taskType == "bug":
		return red(tag)
	case taskType == "feature":
		return magenta(tag)
	case taskType == "chore":
		return dim(tag)
	default:
		return tag
	}
}

// buildDependencyTree prints tasks in a tree format showing dependency chains
func printDependencyTree(tasks []*Task, taskMap map[string]*Task, showAll bool, featureFilter string) {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Find root tasks (tasks with no dependencies or whose deps are all outside the displayed set)
	taskSet := make(map[string]bool)
	for _, t := range tasks {
		taskSet[t.Name] = true
	}

	// Build children map (reverse of depends_on)
	children := make(map[string][]string)
	rootSet := make(map[string]bool)
	for _, t := range tasks {
		isRoot := true
		for _, dep := range t.DependsOn {
			if taskSet[dep] {
				children[dep] = append(children[dep], t.Name)
				isRoot = false
			}
		}
		if isRoot {
			rootSet[t.Name] = true
		}
	}

	// Sort roots by priority first, then status
	sortRoots := func(roots []string) {
		sort.SliceStable(roots, func(i, j int) bool {
			ti := taskMap[roots[i]]
			tj := taskMap[roots[j]]
			// First by priority (lower is higher priority)
			if ti.Priority != tj.Priority {
				return ti.Priority < tj.Priority
			}
			// Then by status
			si := computeTaskStatus(ti, taskMap)
			sj := computeTaskStatus(tj, taskMap)
			statusOrder := map[ComputedStatus]int{StatusReady: 0, StatusInProgress: 1, StatusBlocked: 2, StatusCompleted: 3}
			return statusOrder[si] < statusOrder[sj]
		})
	}

	// Print tree recursively
	printed := make(map[string]bool)
	var printNode func(name string, prefix string, isLast bool, isRoot bool)
	printNode = func(name string, prefix string, isLast bool, isRoot bool) {
		if printed[name] {
			return
		}
		printed[name] = true

		t := taskMap[name]
		if t == nil {
			return
		}

		status := computeTaskStatus(t, taskMap)
		// In tree view, always show completed tasks - they're needed to understand the dependency chain

		// Status indicator
		var indicator string
		switch status {
		case StatusReady:
			indicator = green("●")
		case StatusInProgress:
			indicator = blue("●")
		case StatusBlocked:
			indicator = red("●")
		case StatusCompleted:
			indicator = dim("✓")
		}

		// Priority:type tag (e.g., "P0:bug", "P1", "chore")
		prioType := formatPriorityType(t.Priority, t.Type)
		if prioType != "" {
			prioType = " " + prioType
		}

		if isRoot {
			fmt.Printf("  %s%s %s\n", indicator, prioType, t.Name)
		} else {
			connector := "├── "
			if isLast {
				connector = "└── "
			}
			fmt.Printf("  %s%s%s%s %s\n", prefix, connector, indicator, prioType, t.Name)
		}

		// Get children and sort them by priority then status
		kids := children[name]
		sort.SliceStable(kids, func(i, j int) bool {
			ti := taskMap[kids[i]]
			tj := taskMap[kids[j]]
			if ti == nil || tj == nil {
				return false
			}
			// First by priority (lower is higher priority)
			if ti.Priority != tj.Priority {
				return ti.Priority < tj.Priority
			}
			// Then by status
			si := computeTaskStatus(ti, taskMap)
			sj := computeTaskStatus(tj, taskMap)
			statusOrder := map[ComputedStatus]int{StatusReady: 0, StatusInProgress: 1, StatusBlocked: 2, StatusCompleted: 3}
			return statusOrder[si] < statusOrder[sj]
		})

		// Child prefix
		var childPrefix string
		if isRoot {
			childPrefix = ""
		} else if isLast {
			childPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "│   "
		}

		for i, child := range kids {
			printNode(child, childPrefix, i == len(kids)-1, false)
		}
	}

	// Group ALL tasks by feature (not just roots)
	featureGroups := make(map[string][]*Task)
	var featureOrder []string
	seenFeatures := make(map[string]bool)

	for _, t := range tasks {
		feature := t.Feature
		// If feature filter is applied, skip grouping
		if featureFilter != "" {
			feature = featureFilter
		}
		if !seenFeatures[feature] {
			seenFeatures[feature] = true
			featureOrder = append(featureOrder, feature)
		}
		featureGroups[feature] = append(featureGroups[feature], t)
	}

	// Sort features: non-empty first (alphabetically), then empty
	sort.SliceStable(featureOrder, func(i, j int) bool {
		fi, fj := featureOrder[i], featureOrder[j]
		if fi == "" && fj != "" {
			return false
		}
		if fi != "" && fj == "" {
			return true
		}
		return fi < fj
	})

	// Print each feature group
	for i, feature := range featureOrder {
		featureTasks := featureGroups[feature]

		// Build local task set and children map for this feature
		localTaskSet := make(map[string]bool)
		for _, t := range featureTasks {
			localTaskSet[t.Name] = true
		}

		// Find roots within this feature (tasks whose deps are outside this feature)
		localChildren := make(map[string][]string)
		var localRoots []string
		for _, t := range featureTasks {
			isLocalRoot := true
			for _, dep := range t.DependsOn {
				if localTaskSet[dep] {
					localChildren[dep] = append(localChildren[dep], t.Name)
					isLocalRoot = false
				}
			}
			if isLocalRoot {
				localRoots = append(localRoots, t.Name)
			}
		}

		// Sort local roots
		sortRoots(localRoots)

		// Print feature header (skip if feature filter is applied)
		if featureFilter == "" {
			if feature != "" {
				fmt.Printf("%s\n", yellow(feature+"/"))
			} else {
				fmt.Printf("%s\n", dim("(no feature)"))
			}
		}

		// Override children map for local printing
		oldChildren := children
		children = localChildren

		// Print trees for this feature
		for _, root := range localRoots {
			printNode(root, "", false, true)
		}

		// Restore children map
		children = oldChildren

		// Reset printed map for next feature group
		printed = make(map[string]bool)

		// Add spacing between feature groups
		if i < len(featureOrder)-1 {
			fmt.Println()
		}
	}

	// Legend
	fmt.Printf("\n%s  ready  %s  in_progress  %s  blocked  %s  completed\n",
		green("●"), blue("●"), red("●"), dim("✓"))
}

// printTasksGroupedByFeature prints tasks grouped by feature
func printTasksGroupedByFeature(tasks []*Task, taskMap map[string]*Task, showAll bool,
	green, blue, red, yellow, dim func(a ...interface{}) string) {

	// Group tasks by feature
	featureGroups := make(map[string][]*Task)
	var featureOrder []string
	seenFeatures := make(map[string]bool)

	for _, t := range tasks {
		feature := t.Feature
		if feature == "" {
			feature = "" // empty string for no feature
		}
		if !seenFeatures[feature] {
			seenFeatures[feature] = true
			featureOrder = append(featureOrder, feature)
		}
		featureGroups[feature] = append(featureGroups[feature], t)
	}

	// Sort feature order: non-empty features first (alphabetically), then empty
	sort.SliceStable(featureOrder, func(i, j int) bool {
		if featureOrder[i] == "" && featureOrder[j] != "" {
			return false
		}
		if featureOrder[i] != "" && featureOrder[j] == "" {
			return true
		}
		return featureOrder[i] < featureOrder[j]
	})

	// Print each feature group
	for _, feature := range featureOrder {
		groupTasks := featureGroups[feature]

		// Count active tasks in this group
		activeInGroup := 0
		for _, t := range groupTasks {
			status := computeTaskStatus(t, taskMap)
			if status != StatusCompleted {
				activeInGroup++
			}
		}

		// Skip empty groups (all completed and not showing all)
		if activeInGroup == 0 && !showAll {
			continue
		}

		// Print feature header
		if feature != "" {
			fmt.Printf("%s\n", yellow(feature+"/"))
		} else {
			// Only print header if there are also feature groups
			hasFeatureGroups := false
			for _, f := range featureOrder {
				if f != "" {
					hasFeatureGroups = true
					break
				}
			}
			if hasFeatureGroups {
				fmt.Printf("%s\n", dim("(no feature)"))
			}
		}

		// Print tasks in this group
		for _, t := range groupTasks {
			status := computeTaskStatus(t, taskMap)

			// Skip completed unless --all
			if status == StatusCompleted && !showAll {
				continue
			}

			// Status badge
			var statusBadge string
			switch status {
			case StatusReady:
				statusBadge = green("[ready]")
			case StatusInProgress:
				statusBadge = blue("[in_progress]")
			case StatusBlocked:
				statusBadge = red("[blocked]")
			case StatusCompleted:
				statusBadge = dim("[completed]") + " ✓"
			}

			// Priority:type tag
			prioType := formatPriorityType(t.Priority, t.Type)
			if prioType != "" {
				prioType = " " + prioType
			}

			// Indent if under a feature group
			indent := ""
			if feature != "" || len(featureOrder) > 1 {
				indent = "  "
			}

			fmt.Printf("%s%s%s %s\n", indent, statusBadge, prioType, t.Name)

			// Content preview (first line)
			if t.Content != "" {
				lines := strings.SplitN(t.Content, "\n", 2)
				preview := strings.TrimSpace(lines[0])
				if preview != "" {
					fmt.Printf("%s  %s\n", indent, dim(preview))
				}
			}

			// Pending dependencies
			if status == StatusBlocked {
				pending := getPendingDependencies(t, taskMap)
				if len(pending) > 0 {
					fmt.Printf("%s  %s waiting: %s\n", indent, yellow("↳"), strings.Join(pending, ", "))
				}
			}
		}

		fmt.Println()
	}
}

// --- Subcommand implementations ---

func runTaskList(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter by feature if specified
	if taskFeature != "" {
		tasks = filterTasksByFeature(tasks, taskFeature)
	}

	if len(tasks) == 0 {
		if taskJSON {
			emptyTaskMap := map[string]*Task{}
			return printJSON(buildTaskListJSON(projectName, taskFeature, tasks, emptyTaskMap, false))
		}
		if taskFeature != "" {
			fmt.Printf("%s No tasks for project '%s' with feature '%s'\n", yellow("!"), projectName, taskFeature)
		} else {
			fmt.Printf("%s No tasks for project '%s'\n", yellow("!"), projectName)
		}
		fmt.Println("\nCreate a task:")
		fmt.Printf("  agmd task new my-task --content \"Description\"\n")
		return nil
	}

	// Build task map for dependency resolution (use all project tasks for dep resolution)
	allTasks, _ := loadProjectTasks(reg, projectName)
	taskMap := make(map[string]*Task)
	for _, t := range allTasks {
		taskMap[t.Name] = t
	}

	// Filter by status if specified
	if taskStatus != "" {
		validStatuses := map[string]ComputedStatus{
			"ready":       StatusReady,
			"blocked":     StatusBlocked,
			"in_progress": StatusInProgress,
			"completed":   StatusCompleted,
		}
		target, ok := validStatuses[strings.ToLower(taskStatus)]
		if !ok {
			return fmt.Errorf("invalid status '%s'. Use: ready, blocked, in_progress, or completed", taskStatus)
		}
		var filtered []*Task
		for _, t := range tasks {
			if computeTaskStatus(t, taskMap) == target {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered

		// Implicitly show completed when filtering for them
		if target == StatusCompleted {
			taskAll = true
		}
	}

	// Filter by priority if specified
	if taskFilterPriority != "" {
		priority, err := parsePriority(taskFilterPriority)
		if err != nil {
			return err
		}
		var filtered []*Task
		for _, t := range tasks {
			if t.Priority == priority {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Filter by type if specified
	if taskFilterType != "" {
		if !validTaskTypes[taskFilterType] {
			return fmt.Errorf("invalid type '%s'. Use: bug, feature, task, or chore", taskFilterType)
		}
		var filtered []*Task
		for _, t := range tasks {
			// Empty type defaults to "task"
			taskType := t.Type
			if taskType == "" {
				taskType = "task"
			}
			if taskType == taskFilterType {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Sort by priority then status
	sorted := sortTasksByPriorityAndStatus(tasks, taskMap)
	if taskJSON && taskTree {
		return newTaskError("invalid_input", "cannot use --tree with --json")
	}

	visible := sorted
	if !taskAll {
		filtered := make([]*Task, 0, len(sorted))
		for _, task := range sorted {
			if computeTaskStatus(task, taskMap) != StatusCompleted {
				filtered = append(filtered, task)
			}
		}
		visible = filtered
	}
	if taskJSON {
		return printJSON(buildTaskListJSON(projectName, taskFeature, visible, taskMap, false))
	}

	// Count by status
	completedCount := 0
	for _, t := range sorted {
		if computeTaskStatus(t, taskMap) == StatusCompleted {
			completedCount++
		}
	}

	// Header
	activeCount := len(tasks) - completedCount
	if taskFeature != "" {
		fmt.Printf("\nTasks for: %s [%s] (%d active", cyan(projectName), yellow(taskFeature), activeCount)
	} else {
		fmt.Printf("\nTasks for: %s (%d active", cyan(projectName), activeCount)
	}
	if completedCount > 0 {
		fmt.Printf(", %d completed", completedCount)
	}
	fmt.Printf(")\n\n")

	// Tree view
	if taskTree {
		printDependencyTree(tasks, taskMap, taskAll, taskFeature)
		return nil
	}

	// Group tasks by feature (only when not filtering by feature)
	if taskFeature == "" {
		printTasksGroupedByFeature(sorted, taskMap, taskAll, green, blue, red, yellow, dim)
		return nil
	}

	// Print tasks (when filtering by feature - no grouping needed)
	for _, t := range sorted {
		status := computeTaskStatus(t, taskMap)

		// Skip completed unless --all
		if status == StatusCompleted && !taskAll {
			continue
		}

		// Status badge
		var statusBadge string
		switch status {
		case StatusReady:
			statusBadge = green("[ready]")
		case StatusInProgress:
			statusBadge = blue("[in_progress]")
		case StatusBlocked:
			statusBadge = red("[blocked]")
		case StatusCompleted:
			statusBadge = dim("[completed]") + " ✓"
		}

		// Priority:type tag (e.g., "P0:bug", "P1", "chore")
		prioType := formatPriorityType(t.Priority, t.Type)
		if prioType != "" {
			prioType = " " + prioType
		}

		fmt.Printf("%s%s %s\n", statusBadge, prioType, t.Name)

		// Subject (if different from name)
		if t.Subject != "" && t.Subject != strings.Title(strings.ReplaceAll(t.Name, "-", " ")) {
			fmt.Printf("  %s\n", t.Subject)
		}

		// Content preview (first line)
		if t.Content != "" {
			lines := strings.SplitN(t.Content, "\n", 2)
			preview := strings.TrimSpace(lines[0])
			if preview != "" {
				fmt.Printf("  %s\n", dim(preview))
			}
		}

		// Pending dependencies
		if status == StatusBlocked {
			pending := getPendingDependencies(t, taskMap)
			if len(pending) > 0 {
				fmt.Printf("  %s waiting: %s\n", yellow("↳"), strings.Join(pending, ", "))
			}
		}

		fmt.Println()
	}

	return nil
}

func runTaskNew(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	name := args[0]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	// Validate priority
	if taskPriority < 0 || taskPriority > 4 {
		return newTaskError("invalid_input", "invalid priority %d. Use 0-4 (0=critical, 4=backlog)", taskPriority)
	}

	// Validate type
	if !validTaskTypes[taskType] {
		return newTaskError("invalid_input", "invalid type '%s'. Use: bug, feature, task, or chore", taskType)
	}

	// Build path: ~/.agmd/task/<project>/<name>.md
	taskDir := filepath.Join(reg.BasePath, "task", projectName)
	filePath := filepath.Join(taskDir, name+".md")

	// Check if exists
	if _, err := os.Stat(filePath); err == nil {
		return newTaskError("invalid_input", "task:%s already exists in project '%s'", name, projectName)
	}

	// Create directory
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Determine content source
	var content string
	if taskContent != "" {
		content = taskContent
	} else if !isTerminal(os.Stdin) {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = string(stdinContent)
	}

	// Parse blocked-by dependencies
	var dependsOn []string
	if taskBlockedBy != "" {
		deps := strings.Split(taskBlockedBy, ",")
		for _, dep := range deps {
			dep = strings.TrimSpace(dep)
			if dep != "" {
				dependsOn = append(dependsOn, dep)
			}
		}

		// Validate that all dependencies exist
		for _, dep := range dependsOn {
			depPath := getTaskPath(reg, projectName, dep)
			if _, err := os.Stat(depPath); os.IsNotExist(err) {
				return newTaskError("not_found", "dependency task '%s' not found in project '%s'", dep, projectName)
			}
		}
	}

	// Build task frontmatter
	subject := strings.Title(strings.ReplaceAll(name, "-", " "))

	// Format depends_on as YAML array
	dependsOnYAML := "[]"
	if len(dependsOn) > 0 {
		dependsOnYAML = "[" + strings.Join(dependsOn, ", ") + "]"
	}

	// Build optional lines
	var optionalLines string
	if taskFeature != "" {
		optionalLines += fmt.Sprintf("feature: %s\n", taskFeature)
	}
	// Always write priority so it round-trips correctly (0 = P0 critical, 2 = P2 default)
	optionalLines += fmt.Sprintf("priority: %d\n", taskPriority)
	// Only include type if non-default (not "task")
	if taskType != "task" {
		optionalLines += fmt.Sprintf("type: %s\n", taskType)
	}

	taskID := computeTaskID(projectName, name)
	fileContent := fmt.Sprintf("---\nid: %s\nsubject: %s\nstatus: pending\n%sdepends_on: %s\n---\n\n%s",
		taskID, subject, optionalLines, dependsOnYAML, strings.TrimSpace(content))

	if err := os.WriteFile(filePath, []byte(fileContent+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Build info message
	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "new",
			Project: projectName,
			Task:    name,
			ID:      taskID,
			Status:  "pending",
		})
	}

	infoMsg := fmt.Sprintf("%s Created task:%s (project: %s", green("ok"), name, projectName)
	if taskPriority != 2 {
		infoMsg += fmt.Sprintf(", P%d", taskPriority)
	}
	if taskType != "task" {
		infoMsg += fmt.Sprintf(", %s", taskType)
	}
	infoMsg += ")"
	fmt.Println(infoMsg)

	// Open editor unless --no-editor or content was provided
	if taskNoEditor || taskContent != "" || !isTerminal(os.Stdin) {
		fmt.Printf("%s %s\n", blue("->"), filePath)
		return nil
	}

	fmt.Printf("%s Opening editor...\n", blue("->"))
	return openInEditor(filePath)
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	dim := color.New(color.Faint).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Handle "agmd task show --all"
	if taskAll {
		return runTaskShowAll(reg)
	}

	if len(args) == 0 {
		return newTaskError("invalid_input", "specify a task name or use --all")
	}

	name := args[0]

	// Check if name includes project (project/task-name)
	projectNameFromName, taskName, hasProjectInName := parseTaskRef(name)
	if hasProjectInName && taskProject != "" && projectNameFromName != taskProject {
		return newTaskError(
			"project_mismatch",
			"task reference '%s' does not match --project '%s'",
			projectNameFromName,
			taskProject,
		)
	}

	projectName := taskProject
	if hasProjectInName {
		projectName = projectNameFromName
	}
	if projectName == "" {
		resolved, err := getProjectName()
		if err != nil {
			return err
		}
		projectName = resolved
	}

	taskPath := getTaskPath(reg, projectName, taskName)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return newTaskError("not_found", "task '%s' not found in project '%s'", taskName, projectName)
	}

	if taskRaw {
		raw, err := os.ReadFile(taskPath)
		if err != nil {
			return fmt.Errorf("failed to read task: %w", err)
		}
		fmt.Print(string(raw))
		return nil
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}
	ensureTaskIdentity(task, projectName)
	taskMap := map[string]*Task{
		task.Name: task,
	}

	if taskJSON {
		return printJSON(taskToJSONItem(task, taskMap, true))
	}

	fmt.Printf("%s %s\n", dim("subject:"), task.Subject)
	fmt.Printf("%s %s\n", dim("status:"), task.Status)
	// Always show priority for non-default (P2 is default)
	if task.Priority != 2 {
		fmt.Printf("%s P%d\n", dim("priority:"), task.Priority)
	}
	if task.Type != "" && task.Type != "task" {
		fmt.Printf("%s %s\n", dim("type:"), task.Type)
	}
	if task.Feature != "" {
		fmt.Printf("%s %s\n", dim("feature:"), task.Feature)
	}
	if len(task.DependsOn) > 0 {
		fmt.Printf("%s %s\n", dim("depends_on:"), strings.Join(task.DependsOn, ", "))
	}
	if task.Content != "" {
		fmt.Printf("\n%s\n", task.Content)
	}

	return nil
}

func runTaskShowAll(reg *registry.Registry) error {
	dim := color.New(color.Faint).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter by feature if specified
	if taskFeature != "" {
		tasks = filterTasksByFeature(tasks, taskFeature)
	}

	if len(tasks) == 0 {
		if taskJSON {
			emptyTaskMap := map[string]*Task{}
			return printJSON(buildTaskListJSON(projectName, taskFeature, tasks, emptyTaskMap, true))
		}
		if taskFeature != "" {
			return newTaskError("not_found", "no tasks found for project '%s' with feature '%s'", projectName, taskFeature)
		}
		return newTaskError("not_found", "no tasks found for project '%s'", projectName)
	}

	// Build task map and sort
	taskMap := make(map[string]*Task)
	for _, t := range tasks {
		taskMap[t.Name] = t
	}
	sorted := sortTasksByPriorityAndStatus(tasks, taskMap)
	if taskJSON {
		return printJSON(buildTaskListJSON(projectName, taskFeature, sorted, taskMap, true))
	}

	fmt.Printf("Tasks for: %s\n\n", cyan(projectName))

	for i, t := range sorted {
		status := computeTaskStatus(t, taskMap)

		fmt.Printf("%s %s [%s]\n", dim("---"), t.Name, string(status))
		fmt.Printf("%s %s\n", dim("subject:"), t.Subject)
		fmt.Printf("%s %s\n", dim("status:"), t.Status)
		if t.Priority != 0 {
			fmt.Printf("%s P%d\n", dim("priority:"), t.Priority)
		}
		if t.Type != "" && t.Type != "task" {
			fmt.Printf("%s %s\n", dim("type:"), t.Type)
		}
		if t.Feature != "" {
			fmt.Printf("%s %s\n", dim("feature:"), t.Feature)
		}
		if len(t.DependsOn) > 0 {
			fmt.Printf("%s %s\n", dim("depends_on:"), strings.Join(t.DependsOn, ", "))
		}
		if t.Content != "" {
			fmt.Printf("\n%s\n", t.Content)
		}
		if i < len(sorted)-1 {
			fmt.Println()
		}
	}

	return nil
}

func runTaskClean(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Find completed tasks
	var completedTasks []*Task
	for _, t := range tasks {
		if t.Status == "completed" {
			completedTasks = append(completedTasks, t)
		}
	}

	if len(completedTasks) == 0 {
		if taskJSON {
			return printJSON(taskMutationJSONPayload{
				OK:      true,
				Action:  "clean",
				Project: projectName,
			})
		}
		fmt.Printf("%s No completed tasks to clean in project '%s'\n", yellow("!"), projectName)
		return nil
	}

	// Show what will be deleted
	if !taskJSON {
		fmt.Printf("Found %d completed task(s) to delete:\n", len(completedTasks))
		for _, t := range completedTasks {
			fmt.Printf("  - %s\n", t.Name)
		}
	}

	// Confirmation prompt (unless --force)
	if !taskForce {
		if taskJSON {
			return newTaskError("invalid_input", "task clean requires --force when --json is enabled")
		}
		fmt.Printf("\n%s This will permanently delete these tasks.\n", yellow("⚠"))
		fmt.Print("\nAre you sure? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("\nCancelled.")
			return nil
		}
	}

	// Delete the tasks
	deleted := 0
	for _, t := range completedTasks {
		if err := os.Remove(t.FilePath); err != nil {
			fmt.Printf("  Failed to delete %s: %v\n", t.Name, err)
		} else {
			deleted++
		}
	}

	// Clean up empty project directory
	taskDir := getTaskDir(reg, projectName)
	entries, err := os.ReadDir(taskDir)
	if err == nil && len(entries) == 0 {
		os.Remove(taskDir)
	}

	if taskJSON {
		return printJSON(struct {
			OK      bool   `json:"ok"`
			Action  string `json:"action"`
			Project string `json:"project"`
			Deleted int    `json:"deleted"`
		}{
			OK:      true,
			Action:  "clean",
			Project: projectName,
			Deleted: deleted,
		})
	}

	fmt.Printf("%s Deleted %d completed task(s)\n", green("✓"), deleted)

	return nil
}

func runTaskEdit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	taskName := args[0]

	// Check if at least one flag is provided
	if taskEditPriority == -1 && taskEditType == "" {
		return newTaskError("invalid_input", "specify at least one of --priority or --type")
	}

	// Validate priority if provided
	if taskEditPriority != -1 && (taskEditPriority < 0 || taskEditPriority > 4) {
		return newTaskError("invalid_input", "invalid priority %d. Use 0-4 (0=critical, 4=backlog)", taskEditPriority)
	}

	// Validate type if provided
	if taskEditType != "" && !validTaskTypes[taskEditType] {
		return newTaskError("invalid_input", "invalid type '%s'. Use: bug, feature, task, or chore", taskEditType)
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	taskPath := getTaskPath(reg, projectName, taskName)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return newTaskError("not_found", "task '%s' not found in project '%s'", taskName, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}
	ensureTaskIdentity(task, projectName)

	// Track changes for output message
	var changes []string

	// Update priority if specified
	if taskEditPriority != -1 {
		oldPriority := task.Priority
		task.Priority = taskEditPriority
		changes = append(changes, fmt.Sprintf("priority P%d → P%d", oldPriority, taskEditPriority))
	}

	// Update type if specified
	if taskEditType != "" {
		oldType := task.Type
		if oldType == "" {
			oldType = "task"
		}
		task.Type = taskEditType
		changes = append(changes, fmt.Sprintf("type %s → %s", oldType, taskEditType))
	}

	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "edit",
			Project: projectName,
			Task:    taskName,
			ID:      task.ID,
			Status:  task.Status,
		})
	}

	fmt.Printf("%s Updated task '%s': %s\n", green("✓"), taskName, strings.Join(changes, ", "))
	return nil
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	name := args[0]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	taskPath := getTaskPath(reg, projectName, name)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return newTaskError("not_found", "task '%s' not found in project '%s'", name, projectName)
	}

	taskID := computeTaskID(projectName, name)

	// Show what will be deleted
	if !taskJSON {
		fmt.Printf("%s Deleting task:%s (project: %s)\n", blue("→"), name, projectName)
		fmt.Printf("  Path: %s\n", taskPath)
	}

	// Confirmation prompt (unless --force)
	if !taskForce {
		if taskJSON {
			return newTaskError("invalid_input", "task delete requires --force when --json is enabled")
		}
		fmt.Printf("\n%s This will permanently delete this task.\n", yellow("⚠"))
		fmt.Print("\nAre you sure? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("\nCancelled.")
			return nil
		}
	}

	// Delete the file
	if err := os.Remove(taskPath); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Clean up empty project directory
	taskDir := getTaskDir(reg, projectName)
	entries, err := os.ReadDir(taskDir)
	if err == nil && len(entries) == 0 {
		os.Remove(taskDir)
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "delete",
			Project: projectName,
			Task:    name,
			ID:      taskID,
		})
	}

	fmt.Printf("%s Deleted task:%s\n", green("✓"), name)
	return nil
}

func runTaskStatus(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	taskName := args[0]
	newStatus := strings.ToLower(args[1])

	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
	}
	if !validStatuses[newStatus] {
		return newTaskError("invalid_input", "invalid status '%s'. Use: pending, in_progress, or completed", newStatus)
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	taskPath := getTaskPath(reg, projectName, taskName)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return newTaskError("not_found", "task '%s' not found in project '%s'", taskName, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}
	ensureTaskIdentity(task, projectName)

	if task.Status == "completed" && newStatus != "completed" {
		return newTaskError(
			"invalid_transition",
			"invalid status transition for task '%s': completed -> %s",
			taskName,
			newStatus,
		)
	}

	task.Status = newStatus
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "status",
			Project: projectName,
			Task:    taskName,
			ID:      task.ID,
			Status:  newStatus,
		})
	}

	fmt.Printf("%s Updated task '%s' status to '%s'\n", green("✓"), taskName, newStatus)
	return nil
}

func runTaskBlockedBy(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	taskName := args[0]
	dependency := args[1]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	taskPath := getTaskPath(reg, projectName, taskName)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return newTaskError("not_found", "task '%s' not found in project '%s'", taskName, projectName)
	}

	// Check dependency exists
	depPath := getTaskPath(reg, projectName, dependency)
	if _, err := os.Stat(depPath); os.IsNotExist(err) {
		return newTaskError("not_found", "dependency task '%s' not found in project '%s'", dependency, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}
	ensureTaskIdentity(task, projectName)

	// Check if already depends on it
	for _, dep := range task.DependsOn {
		if dep == dependency {
			return newTaskError("invalid_input", "task '%s' already depends on '%s'", taskName, dependency)
		}
	}

	task.DependsOn = append(task.DependsOn, dependency)
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "blocked-by",
			Project: projectName,
			Task:    taskName,
			ID:      task.ID,
			Status:  task.Status,
		})
	}

	fmt.Printf("%s Added dependency: '%s' is now blocked by '%s'\n", green("✓"), taskName, dependency)
	return nil
}

func runTaskUnblock(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	taskName := args[0]
	dependency := args[1]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	taskPath := getTaskPath(reg, projectName, taskName)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return newTaskError("not_found", "task '%s' not found in project '%s'", taskName, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}
	ensureTaskIdentity(task, projectName)

	// Remove dependency
	found := false
	newDeps := []string{}
	for _, dep := range task.DependsOn {
		if dep == dependency {
			found = true
		} else {
			newDeps = append(newDeps, dep)
		}
	}

	if !found {
		return newTaskError("not_found", "task '%s' does not depend on '%s'", taskName, dependency)
	}

	task.DependsOn = newDeps
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "unblock",
			Project: projectName,
			Task:    taskName,
			ID:      task.ID,
			Status:  task.Status,
		})
	}

	fmt.Printf("%s Removed dependency: '%s' is no longer blocked by '%s'\n", green("✓"), taskName, dependency)
	return nil
}
