package cmd

import (
	"fmt"
	"io"
	"os"
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
	Subject     string   `yaml:"subject"`
	Status      string   `yaml:"status"`
	Feature     string   `yaml:"feature,omitempty"`
	DependsOn   []string `yaml:"depends_on"`
	Content     string   `yaml:"-"`
	FilePath    string   `yaml:"-"`
	ProjectName string   `yaml:"-"`
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
var taskFeature string
var taskAll bool
var taskForce bool
var taskContent string
var taskBlockedBy string
var taskNoEditor bool
var taskRaw bool
var taskStatus string
var taskTree bool

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage project tasks",
	Long: `Manage project tasks with dependencies and status tracking.

Tasks are organized by project (based on current directory name).
Use --feature to scope tasks to a specific feature/session.

Subcommands:
  list        List tasks for current project
  new         Create a new task
  show        Show task content
  delete      Delete a task
  status      Update task status
  blocked-by  Add a dependency
  unblock     Remove a dependency

Examples:
  agmd task list                                    # List all tasks
  agmd task list --feature auth                     # List tasks for "auth" feature
  agmd task new setup-db --content "Set up DB"      # Create task
  agmd task new setup-db --feature auth             # Create task scoped to feature
  agmd task show setup-db                           # Show task
  agmd task delete setup-db                         # Delete task
  agmd task status setup-db completed               # Update status
  agmd task blocked-by create-api setup-db          # Add dependency
  agmd task unblock create-api setup-db             # Remove dependency`,
}

var taskListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tasks for current project",
	Long: `List tasks for the current project, auto-sorted by status.

Tasks are sorted: ready → in_progress → blocked → completed.
Completed tasks are hidden by default (use --all to show).
Use --feature to filter tasks by feature/session.
Use --status to filter by computed status (ready, blocked, in_progress, completed).
Use --tree to show dependency tree visualization.

Examples:
  agmd task list                            # List active tasks
  agmd task ls                              # Same (alias)
  agmd task list --all                      # Include completed tasks
  agmd task list --feature auth             # Only tasks for "auth" feature
  agmd task list --status ready             # Only ready tasks
  agmd task list --status blocked           # Only blocked tasks
  agmd task list --tree                     # Show dependency tree
  agmd task list --project myproj           # List tasks for specific project`,
	RunE: runTaskList,
}

var taskNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new task",
	Long: `Create a new task for the current project.

Examples:
  agmd task new setup-db --content "Set up database"
  agmd task new create-api --content "Create API" --blocked-by "setup-db"
  agmd task new setup-db --feature auth --content "Set up auth DB"
  agmd task new my-task --project other-project
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
	RunE: runTaskShow,
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
	Args: cobra.ExactArgs(1),
	RunE: runTaskDelete,
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
	Args: cobra.ExactArgs(2),
	RunE: runTaskStatus,
}

var taskBlockedByCmd = &cobra.Command{
	Use:   "blocked-by <task-name> <dependency>",
	Short: "Add a dependency to a task",
	Long: `Add a dependency to a task.

This makes <task-name> depend on <dependency>.

Examples:
  agmd task blocked-by create-api setup-db    # create-api depends on setup-db`,
	Args: cobra.ExactArgs(2),
	RunE: runTaskBlockedBy,
}

var taskUnblockCmd = &cobra.Command{
	Use:   "unblock <task-name> <dependency>",
	Short: "Remove a dependency from a task",
	Long: `Remove a dependency from a task.

Examples:
  agmd task unblock create-api setup-db    # Remove setup-db dependency from create-api`,
	Args: cobra.ExactArgs(2),
	RunE: runTaskUnblock,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskNewCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskBlockedByCmd)
	taskCmd.AddCommand(taskUnblockCmd)

	// Add --project and --feature flags to subcommands that need them
	taskListCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
	taskListCmd.Flags().StringVar(&taskFeature, "feature", "", "Filter tasks by feature")
	taskListCmd.Flags().BoolVarP(&taskAll, "all", "a", false, "Include completed tasks")
	taskListCmd.Flags().StringVar(&taskStatus, "status", "", "Filter by computed status (ready, blocked, in_progress, completed)")
	taskListCmd.Flags().BoolVar(&taskTree, "tree", false, "Show dependency tree visualization")

	taskNewCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
	taskNewCmd.Flags().StringVar(&taskFeature, "feature", "", "Feature/session name for this task")
	taskNewCmd.Flags().StringVar(&taskContent, "content", "", "Task content/description")
	taskNewCmd.Flags().StringVar(&taskBlockedBy, "blocked-by", "", "Comma-separated list of task dependencies")
	taskNewCmd.Flags().BoolVar(&taskNoEditor, "no-editor", false, "Don't open editor after creating")

	taskShowCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
	taskShowCmd.Flags().StringVar(&taskFeature, "feature", "", "Filter tasks by feature")
	taskShowCmd.Flags().BoolVarP(&taskAll, "all", "a", false, "Show all tasks for project")
	taskShowCmd.Flags().BoolVar(&taskRaw, "raw", false, "Include frontmatter in output")

	taskDeleteCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
	taskDeleteCmd.Flags().BoolVarP(&taskForce, "force", "f", false, "Skip confirmation prompt")

	taskStatusCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
	taskBlockedByCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
	taskUnblockCmd.Flags().StringVar(&taskProject, "project", "", "Project name (default: current directory name)")
}

// getProjectName returns the project name (from flag or cwd)
func getProjectName() (string, error) {
	if taskProject != "" {
		return taskProject, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Base(cwd), nil
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
		Subject   string   `yaml:"subject"`
		Status    string   `yaml:"status"`
		Feature   string   `yaml:"feature,omitempty"`
		DependsOn []string `yaml:"depends_on"`
	}{
		Subject:   task.Subject,
		Status:    task.Status,
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
		task.ProjectName = projectName
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

// sortTasksByStatus sorts tasks: ready -> in_progress -> blocked -> completed
func sortTasksByStatus(tasks []*Task, allTasks map[string]*Task) []*Task {
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

	// Sort by status priority
	statusPriority := map[ComputedStatus]int{
		StatusReady:      0,
		StatusInProgress: 1,
		StatusBlocked:    2,
		StatusCompleted:  3,
	}

	sort.SliceStable(tasksWithStatus, func(i, j int) bool {
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

// buildDependencyTree prints tasks in a tree format showing dependency chains
func printDependencyTree(tasks []*Task, taskMap map[string]*Task, showAll bool, featureFilter string) {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	// Find root tasks (tasks with no dependencies or whose deps are all outside the displayed set)
	taskSet := make(map[string]bool)
	for _, t := range tasks {
		taskSet[t.Name] = true
	}

	// Build children map (reverse of depends_on)
	children := make(map[string][]string)
	roots := []string{}
	for _, t := range tasks {
		isRoot := true
		for _, dep := range t.DependsOn {
			if taskSet[dep] {
				children[dep] = append(children[dep], t.Name)
				isRoot = false
			}
		}
		if isRoot {
			roots = append(roots, t.Name)
		}
	}

	// Sort roots by status priority
	sort.SliceStable(roots, func(i, j int) bool {
		ti := taskMap[roots[i]]
		tj := taskMap[roots[j]]
		si := computeTaskStatus(ti, taskMap)
		sj := computeTaskStatus(tj, taskMap)
		priority := map[ComputedStatus]int{StatusReady: 0, StatusInProgress: 1, StatusBlocked: 2, StatusCompleted: 3}
		return priority[si] < priority[sj]
	})

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
		if status == StatusCompleted && !showAll {
			return
		}

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

		// Feature tag
		featureTag := ""
		if t.Feature != "" && featureFilter == "" {
			featureTag = " " + dim("("+t.Feature+")")
		}

		if isRoot {
			fmt.Printf("%s %s%s\n", indicator, t.Name, featureTag)
		} else {
			connector := "├── "
			if isLast {
				connector = "└── "
			}
			fmt.Printf("%s%s%s %s%s\n", prefix, connector, indicator, t.Name, featureTag)
		}

		// Get children and sort them
		kids := children[name]
		sort.SliceStable(kids, func(i, j int) bool {
			ti := taskMap[kids[i]]
			tj := taskMap[kids[j]]
			if ti == nil || tj == nil {
				return false
			}
			si := computeTaskStatus(ti, taskMap)
			sj := computeTaskStatus(tj, taskMap)
			priority := map[ComputedStatus]int{StatusReady: 0, StatusInProgress: 1, StatusBlocked: 2, StatusCompleted: 3}
			return priority[si] < priority[sj]
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

	for _, root := range roots {
		printNode(root, "", false, true)
	}

	// Legend
	fmt.Printf("\n%s  ready  %s  in_progress  %s  blocked  %s  completed\n",
		green("●"), blue("●"), red("●"), dim("✓"))
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

	// Sort by dependency status
	sorted := sortTasksByStatus(tasks, taskMap)

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

	// Print tasks
	for _, t := range sorted {
		status := computeTaskStatus(t, taskMap)

		// Skip completed unless --all
		if status == StatusCompleted && !taskAll {
			continue
		}

		// Status badge
		var badge string
		switch status {
		case StatusReady:
			badge = green("[ready]")
		case StatusInProgress:
			badge = blue("[in_progress]")
		case StatusBlocked:
			badge = red("[blocked]")
		case StatusCompleted:
			badge = dim("[completed] ✓")
		}

		// Show feature tag when not filtering by feature
		if t.Feature != "" && taskFeature == "" {
			fmt.Printf("%s %s %s\n", badge, t.Name, dim("("+t.Feature+")"))
		} else {
			fmt.Printf("%s %s\n", badge, t.Name)
		}

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

	// Build path: ~/.agmd/task/<project>/<name>.md
	taskDir := filepath.Join(reg.BasePath, "task", projectName)
	filePath := filepath.Join(taskDir, name+".md")

	// Check if exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("task:%s already exists in project '%s'", name, projectName)
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
				return fmt.Errorf("dependency task '%s' not found in project '%s'", dep, projectName)
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

	// Build feature line (optional)
	featureLine := ""
	if taskFeature != "" {
		featureLine = fmt.Sprintf("feature: %s\n", taskFeature)
	}

	fileContent := fmt.Sprintf("---\nsubject: %s\nstatus: pending\n%sdepends_on: %s\n---\n\n%s",
		subject, featureLine, dependsOnYAML, strings.TrimSpace(content))

	if err := os.WriteFile(filePath, []byte(fileContent+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("%s Created task:%s (project: %s)\n", green("ok"), name, projectName)

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
		return fmt.Errorf("specify a task name or use --all")
	}

	name := args[0]

	// Check if name includes project (project/task-name)
	projectName := taskProject
	taskName := name
	if strings.Contains(name, "/") {
		parts := strings.SplitN(name, "/", 2)
		projectName = parts[0]
		taskName = parts[1]
	}

	if projectName == "" {
		pn, err := getProjectName()
		if err != nil {
			return err
		}
		projectName = pn
	}

	taskPath := getTaskPath(reg, projectName, taskName)
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return fmt.Errorf("task '%s' not found in project '%s'", taskName, projectName)
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

	fmt.Printf("%s %s\n", dim("subject:"), task.Subject)
	fmt.Printf("%s %s\n", dim("status:"), task.Status)
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
		if taskFeature != "" {
			return fmt.Errorf("no tasks found for project '%s' with feature '%s'", projectName, taskFeature)
		}
		return fmt.Errorf("no tasks found for project '%s'", projectName)
	}

	// Build task map and sort
	taskMap := make(map[string]*Task)
	for _, t := range tasks {
		taskMap[t.Name] = t
	}
	sorted := sortTasksByStatus(tasks, taskMap)

	fmt.Printf("Tasks for: %s\n\n", cyan(projectName))

	for i, t := range sorted {
		status := computeTaskStatus(t, taskMap)

		fmt.Printf("%s %s [%s]\n", dim("---"), t.Name, string(status))
		fmt.Printf("%s %s\n", dim("subject:"), t.Subject)
		fmt.Printf("%s %s\n", dim("status:"), t.Status)
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
		return fmt.Errorf("task '%s' not found in project '%s'", name, projectName)
	}

	// Show what will be deleted
	fmt.Printf("%s Deleting task:%s (project: %s)\n", blue("→"), name, projectName)
	fmt.Printf("  Path: %s\n", taskPath)

	// Confirmation prompt (unless --force)
	if !taskForce {
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

	fmt.Printf("%s Deleted task:%s\n", green("✓"), name)

	// Clean up empty project directory
	taskDir := getTaskDir(reg, projectName)
	entries, err := os.ReadDir(taskDir)
	if err == nil && len(entries) == 0 {
		os.Remove(taskDir)
	}

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
		return fmt.Errorf("invalid status '%s'. Use: pending, in_progress, or completed", newStatus)
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
		return fmt.Errorf("task '%s' not found in project '%s'", taskName, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	task.Status = newStatus
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
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
		return fmt.Errorf("task '%s' not found in project '%s'", taskName, projectName)
	}

	// Check dependency exists
	depPath := getTaskPath(reg, projectName, dependency)
	if _, err := os.Stat(depPath); os.IsNotExist(err) {
		return fmt.Errorf("dependency task '%s' not found in project '%s'", dependency, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	// Check if already depends on it
	for _, dep := range task.DependsOn {
		if dep == dependency {
			return fmt.Errorf("task '%s' already depends on '%s'", taskName, dependency)
		}
	}

	task.DependsOn = append(task.DependsOn, dependency)
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
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
		return fmt.Errorf("task '%s' not found in project '%s'", taskName, projectName)
	}

	task, err := loadTask(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

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
		return fmt.Errorf("task '%s' does not depend on '%s'", taskName, dependency)
	}

	task.DependsOn = newDeps
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Printf("%s Removed dependency: '%s' is no longer blocked by '%s'\n", green("✓"), taskName, dependency)
	return nil
}
