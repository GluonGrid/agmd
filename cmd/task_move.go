package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// `task move` puts a task in another project without re-creating it: the
// file, its id, comments and timestamps go along, in one step that holds the
// locks of both projects.

var taskMoveToProject string
var taskMoveAuthor string

var taskMoveCmd = &cobra.Command{
	Use:   "move <task> --to-project <project>",
	Short: "Move a task to another project, keeping its id and comments",
	Long: `Move a task to another project. It keeps its name, id, text, fields
(human flag, check, priority, ...), comments and created/completed times;
updated_at records the move. The task gets a comment "Moved from project
'<from>'." (author from --author, $AGMD_AUTHOR, or "user").

Refused with:
  invalid_input (2)       --to-project is missing, invalid or the same project
  conflict (6)            the task has dependencies or dependents (they would
                          cross projects: remove them with 'task dep remove'
                          first), the target already has a task with that name
                          or id, or another workflow holds a claim on it
                          (pass its --workflow-id to move your own claim along)

Examples:
  agmd task move fix-proxy --project proxmox --to-project agents
  agmd task move proxmox/fix-proxy --to-project agents --json`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskMove,
}

func init() {
	taskCmd.AddCommand(taskMoveCmd)
	taskMoveCmd.Flags().StringVar(&taskProject, "project", "", "Project the task is in (stable selector, overrides automatic resolution)")
	taskMoveCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskMoveCmd.Flags().StringVar(&taskMoveToProject, "to-project", "", "Project to move the task to (required)")
	taskMoveCmd.Flags().StringVar(&taskWorkflowID, "workflow-id", "", "Your workflow id: a claim you hold moves with the task")
	taskMoveCmd.Flags().StringVar(&taskMoveAuthor, "author", "", "Who moves it, recorded in a comment (default $AGMD_AUTHOR or \"user\")")
}

// lockTwoProjects takes both project locks, always in name order, so two
// moves in opposite directions cannot deadlock.
func lockTwoProjects(reg *registry.Registry, a, b string) (func(), error) {
	names := []string{a, b}
	sort.Strings(names)
	unlockFirst, err := lockProject(reg, names[0])
	if err != nil {
		return nil, err
	}
	unlockSecond, err := lockProject(reg, names[1])
	if err != nil {
		unlockFirst()
		return nil, err
	}
	return func() {
		unlockSecond()
		unlockFirst()
	}, nil
}

// moveTask checks and applies a move. The caller must hold the locks of both
// projects. The task is written in the target before it is removed from the
// source, so a crash leaves a duplicate rather than losing it.
func moveTask(reg *registry.Registry, fromProject, toProject string, task *Task, workflowID, author string) error {
	if task.ProjectName != fromProject {
		return fmt.Errorf("task '%s' was loaded from project '%s', not '%s'", task.Name, task.ProjectName, fromProject)
	}
	if len(task.DependsOn) > 0 {
		return newTaskError("conflict", "cannot move '%s': it depends on %s, which would then be in another project; remove the dependencies first", task.Name, strings.Join(task.DependsOn, ", "))
	}
	sourceTasks, err := loadProjectTasks(reg, fromProject)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}
	var dependents []string
	for _, t := range sourceTasks {
		if t.Name != task.Name && dependsOnTask(t, task) {
			dependents = append(dependents, t.Name)
		}
	}
	if len(dependents) > 0 {
		sort.Strings(dependents)
		return newTaskError("conflict", "cannot move '%s': %s depend on it and would then wait on another project; remove those dependencies first", task.Name, strings.Join(dependents, ", "))
	}

	targetPath := getTaskPath(reg, toProject, task.Name)
	if _, err := os.Lstat(targetPath); err == nil {
		return newTaskError("conflict", "project '%s' already has a task named '%s'", toProject, task.Name)
	}
	targetTasks, err := loadProjectTasks(reg, toProject)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}
	for _, t := range targetTasks {
		if t.ID == task.ID {
			return newTaskError("conflict", "project '%s' already has a task with id %s ('%s')", toProject, task.ID, t.Name)
		}
	}

	sourceAssignments, err := loadTaskAssignments(reg, fromProject)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}
	claim, claimed := sourceAssignments[task.ID]
	if claimed && claim.WorkflowID != workflowID {
		return newTaskError("conflict", "cannot move '%s': it is claimed by workflow '%s'; unassign it first", task.Name, claim.WorkflowID)
	}

	comments, err := loadTaskComments(reg, fromProject, task.ID)
	if err != nil {
		return err
	}

	// Write the task in the target; create-once, so a racing 'task new' of
	// the same name there wins cleanly instead of being overwritten.
	task.UpdatedAt = taskTimestamp(taskNow())
	data, err := encodeTask(task)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(getTaskDir(reg, toProject), 0o755); err != nil {
		return err
	}
	if err := writeFileExclusive(targetPath, data, 0o644); err != nil {
		if os.IsExist(err) {
			return newTaskError("conflict", "project '%s' already has a task named '%s'", toProject, task.Name)
		}
		return fmt.Errorf("failed to write task: %w", err)
	}
	sourcePath := task.FilePath
	task.FilePath = targetPath
	task.ProjectName = toProject

	if len(comments) > 0 {
		if err := saveTaskComments(reg, toProject, task.ID, comments); err != nil {
			return err
		}
	}
	if _, err := addTaskComment(reg, toProject, task.ID, author, fmt.Sprintf("Moved from project '%s'.", fromProject)); err != nil {
		return err
	}
	if claimed {
		targetAssignments, err := loadTaskAssignments(reg, toProject)
		if err != nil {
			return fmt.Errorf("failed to load assignments: %w", err)
		}
		targetAssignments[task.ID] = claim
		if err := saveTaskAssignments(reg, toProject, targetAssignments); err != nil {
			return fmt.Errorf("failed to save assignments: %w", err)
		}
	}

	if err := os.Remove(sourcePath); err != nil {
		return fmt.Errorf("task copied to '%s' but not removed from '%s': %w", toProject, fromProject, err)
	}
	if err := removeTaskComments(reg, fromProject, task.ID); err != nil {
		return err
	}
	return dropTaskAssignments(reg, fromProject, task.ID)
}

func runTaskMove(cmd *cobra.Command, args []string) error {
	toProject := strings.TrimSpace(taskMoveToProject)
	if toProject == "" {
		return newTaskError("invalid_input", "--to-project <project> is required")
	}
	if err := validateName("project", toProject); err != nil {
		return err
	}
	author := taskMoveAuthor
	if author == "" {
		author = os.Getenv("AGMD_AUTHOR")
	}
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	fromProject, taskRef, err := resolveTaskProjectAndRef(args[0])
	if err != nil {
		return err
	}
	if fromProject == toProject {
		return newTaskError("invalid_input", "task is already in project '%s'", toProject)
	}
	unlock, err := lockTwoProjects(reg, fromProject, toProject)
	if err != nil {
		return err
	}
	defer unlock()

	task, err := resolveTaskRef(reg, fromProject, taskRef)
	if err != nil {
		return err
	}
	if err := moveTask(reg, fromProject, toProject, task, taskWorkflowID, author); err != nil {
		return err
	}

	if taskJSON {
		return printJSON(struct {
			taskMutationJSONPayload
			FromProject string `json:"from_project"`
		}{
			taskMutationJSONPayload: taskMutationJSONPayload{
				OK:      true,
				Action:  "move",
				Project: toProject,
				Task:    task.Name,
				ID:      task.ID,
				Status:  task.Status,
			},
			FromProject: fromProject,
		})
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Moved task '%s' from '%s' to '%s'\n", green("✓"), task.Name, fromProject, toProject)
	return nil
}
