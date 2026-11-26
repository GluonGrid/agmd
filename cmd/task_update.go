package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// `task update` changes a task's fields and text in one locked step. It
// rewrites only the task file: id, dependencies, comments, claim and
// created_at stay as they are, and saveTask stamps updated_at. `task edit`
// is the older name, kept for callers that use it.

var taskEditPriority int = -1 // -1 means not set
var taskEditType string
var taskUpdateSubject string
var taskUpdateContentFile string

var taskUpdateCmd = &cobra.Command{
	Use:   "update <task-id-or-name>",
	Short: "Change a task's text, subject, priority, type, feature, human flag or check",
	Long: `Change a task's text and fields in one locked step.

The task keeps its id, dependencies, comments, claim and created_at;
updated_at is set to now. Dependencies are changed with 'agmd task dep'.

The new text comes from --content, from --content-file (a path, or - for
stdin), or from stdin when it is piped and no other change is given.
--subject sets the title, -p/--priority 0-4, -t/--type bug|feature|task|chore,
--feature (--feature "" clears it), --human (--human=false hands the task
back to agents), --check sets a human task's check (--check "" clears it).
--set run.<key>=<value> sets a free run field (profile, model, effort, ...;
repeatable, an empty value clears it) and --unset run.<key> clears one.

A completed task is refused with "invalid_transition" (status 5) unless
--force is given.

Examples:
  agmd task update setup-db --content "Create the schema and seed data"
  agmd task update setup-db --content-file notes.md
  cat notes.md | agmd task update setup-db
  agmd task update setup-db --subject "Set up the database" -p 1 -t feature
  agmd task update setup-db --feature ""      # Clear the feature
  agmd task update merge-pr --human --check "gh pr view 12 --json state -q .state | grep -qx MERGED"
  agmd task update setup-db --set run.profile=frontend --unset run.model
  agmd task update setup-db --content "Fix the typo" --force   # Completed task`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskUpdate(cmd, args, "update")
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit <task-id-or-name>",
	Short: "Alias of 'update' (compatibility)",
	Long: `Compatibility alias of 'agmd task update', kept for existing callers.

Unlike 'update' it also changes completed tasks without --force and never
reads stdin. Its JSON reports "action": "edit". Use 'agmd task update'.

Examples:
  agmd task edit setup-db -p 0 -t bug       # Same as: agmd task update setup-db -p 0 -t bug`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskUpdate(cmd, args, "edit")
	},
}

func init() {
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskEditCmd)
	for _, c := range []*cobra.Command{taskUpdateCmd, taskEditCmd} {
		c.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
		c.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
		c.Flags().IntVarP(&taskEditPriority, "priority", "p", -1, "New priority (0=critical, 1=high, 2=medium, 3=low, 4=backlog)")
		c.Flags().StringVarP(&taskEditType, "type", "t", "", "New type (bug, feature, task, chore)")
		c.Flags().StringVar(&taskFeature, "feature", "", "New feature (empty string clears it)")
		c.Flags().BoolVar(&taskHuman, "human", false, "Mark as a task for the human (--human=false clears it)")
		c.Flags().StringVar(&taskCheck, "check", "", "Shell command whose exit 0 means the human step is done (--check \"\" clears it)")
		c.Flags().StringVar(&taskUpdateSubject, "subject", "", "New subject (title)")
		c.Flags().StringVar(&taskContent, "content", "", "New task text (replaces the description)")
		c.Flags().StringVar(&taskUpdateContentFile, "content-file", "", "Read the new task text from this file (- for stdin)")
		c.Flags().BoolVarP(&taskForce, "force", "f", false, "Allow changing a completed task")
		c.Flags().StringArrayVar(&taskSetFields, "set", nil, "Set a run field, run.<key>=<value> (repeatable; an empty value clears it)")
		c.Flags().StringArrayVar(&taskUnsetFields, "unset", nil, "Clear a run field, run.<key> (repeatable)")
		c.RegisterFlagCompletionFunc("priority", completePriorityFlag)
		c.RegisterFlagCompletionFunc("type", completeTypeFlag)
	}
}

// taskUpdate is the set of changes one update applies; nil means unchanged.
type taskUpdate struct {
	Subject  *string
	Content  *string
	Priority *int
	Type     *string
	Feature  *string
	Human    *bool
	Check    *string
	Run      runFieldChanges
}

func (u taskUpdate) empty() bool {
	return u.Subject == nil && u.Content == nil && u.Priority == nil && u.Type == nil &&
		u.Feature == nil && u.Human == nil && u.Check == nil && u.Run.empty()
}

// readTaskUpdate collects the changes from the flags (and, for update,
// stdin) and validates them before the lock is taken.
func readTaskUpdate(cmd *cobra.Command, action string) (taskUpdate, error) {
	var u taskUpdate
	flags := cmd.Flags()
	if flags.Changed("subject") {
		subject := strings.TrimSpace(taskUpdateSubject)
		if subject == "" {
			return u, newTaskError("invalid_input", "--subject cannot be empty")
		}
		u.Subject = &subject
	}
	if flags.Changed("priority") {
		if taskEditPriority < 0 || taskEditPriority > 4 {
			return u, newTaskError("invalid_input", "invalid priority %d. Use 0-4 (0=critical, 4=backlog)", taskEditPriority)
		}
		priority := taskEditPriority
		u.Priority = &priority
	}
	if flags.Changed("type") {
		if !validTaskTypes[taskEditType] {
			return u, newTaskError("invalid_input", "invalid type '%s'. Use: bug, feature, task, or chore", taskEditType)
		}
		taskType := taskEditType
		u.Type = &taskType
	}
	if flags.Changed("feature") {
		feature := strings.TrimSpace(taskFeature)
		u.Feature = &feature
	}
	if flags.Changed("human") {
		human := taskHuman
		u.Human = &human
	}
	if flags.Changed("check") {
		check := strings.TrimSpace(taskCheck)
		u.Check = &check
	}
	run, err := parseRunFieldFlags(taskSetFields, taskUnsetFields)
	if err != nil {
		return u, err
	}
	u.Run = run

	contentSet := flags.Changed("content")
	fileSet := flags.Changed("content-file")
	if contentSet && fileSet {
		return u, newTaskError("invalid_input", "use either --content or --content-file, not both")
	}
	switch {
	case contentSet:
		content := taskContent
		u.Content = &content
	case fileSet:
		content, err := readContentFile(taskUpdateContentFile)
		if err != nil {
			return u, err
		}
		u.Content = &content
	case action == "update" && u.empty() && !isTerminal(os.Stdin):
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return u, fmt.Errorf("failed to read from stdin: %w", err)
		}
		if strings.TrimSpace(string(data)) != "" {
			content := string(data)
			u.Content = &content
		}
	}

	if u.empty() {
		return u, newTaskError("invalid_input", "specify at least one of --content, --content-file, --subject, --priority, --type, --feature, --human, --check, --set or --unset")
	}
	return u, nil
}

// readContentFile reads task text from a path, or from stdin for "-".
func readContentFile(path string) (string, error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", newTaskError("invalid_input", "cannot read --content-file: %v", err)
	}
	return string(data), nil
}

// applyTaskUpdate applies u to task and returns a description of each
// change. It refuses a check on a task that is not for the human.
func applyTaskUpdate(task *Task, u taskUpdate) ([]string, error) {
	var changes []string
	if u.Subject != nil {
		task.Subject = *u.Subject
		changes = append(changes, "subject")
	}
	if u.Content != nil {
		task.Content = strings.TrimSpace(*u.Content)
		changes = append(changes, "content")
	}
	if u.Priority != nil {
		changes = append(changes, fmt.Sprintf("priority P%d → P%d", task.Priority, *u.Priority))
		task.Priority = *u.Priority
	}
	if u.Type != nil {
		oldType := task.Type
		if oldType == "" {
			oldType = "task"
		}
		changes = append(changes, fmt.Sprintf("type %s → %s", oldType, *u.Type))
		task.Type = *u.Type
	}
	if u.Feature != nil {
		changes = append(changes, fmt.Sprintf("feature %s → %s", featureLabel(task.Feature), featureLabel(*u.Feature)))
		task.Feature = *u.Feature
	}
	if u.Human != nil {
		task.Human = *u.Human
		if task.Human {
			changes = append(changes, "for: human")
		} else {
			changes = append(changes, "for: agents")
		}
	}
	if u.Check != nil {
		task.Check = *u.Check
		if task.Check == "" {
			changes = append(changes, "check cleared")
		} else {
			changes = append(changes, "check set")
		}
	}
	changes = append(changes, applyRunFields(task, u.Run)...)
	if task.Check != "" && !task.Human {
		return nil, newTaskError("invalid_input", "task '%s' has a check but is not for the human: a check only applies to human tasks (add --human or clear it with --check \"\")", task.Name)
	}
	return changes, nil
}

// runTaskUpdate implements `task update` and its alias `task edit`.
func runTaskUpdate(cmd *cobra.Command, args []string, action string) error {
	u, err := readTaskUpdate(cmd, action)
	if err != nil {
		return err
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, taskRef, err := resolveTaskProjectAndRef(args[0])
	if err != nil {
		return err
	}

	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return err
	}
	defer unlock()

	task, err := resolveTaskRef(reg, projectName, taskRef)
	if err != nil {
		return err
	}
	if action == "update" && task.Status == "completed" && !taskForce {
		return newTaskError("invalid_transition", "task '%s' is completed; pass --force to change it anyway", task.Name)
	}
	changes, err := applyTaskUpdate(task, u)
	if err != nil {
		return err
	}
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  action,
			Project: projectName,
			Task:    task.Name,
			ID:      task.ID,
			Status:  task.Status,
		})
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Updated task '%s': %s\n", green("✓"), task.Name, strings.Join(changes, ", "))
	return nil
}

// featureLabel names a feature in edit messages; an empty one reads "none".
func featureLabel(feature string) string {
	if feature == "" {
		return "none"
	}
	return feature
}
