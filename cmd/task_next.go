package cmd

import (
	"fmt"
	"strings"
	"time"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var taskClaim bool

var taskNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show or claim the next ready task",
	Long: `Show the highest-priority task that is ready (all dependencies completed)
and not assigned to any workflow.

With --claim, atomically assign that task to --workflow-id and mark it
in_progress, so parallel agents never pick the same task. If the workflow
already holds an unfinished task, that task is returned instead, so an agent
that restarts resumes its work rather than claiming more.

Exits with code "not_found" (status 3) when there is nothing to pick.

Examples:
  agmd task next                                   # Peek at the next task
  agmd task next --claim --workflow-id agent-1 --json
  agmd task next --claim --workflow-id agent-2 --feature auth`,
	Args: cobra.NoArgs,
	RunE: runTaskNext,
}

func init() {
	taskCmd.AddCommand(taskNextCmd)
	taskNextCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskNextCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskNextCmd.Flags().StringVar(&taskFeature, "feature", "", "Only consider tasks in this feature")
	taskNextCmd.Flags().BoolVar(&taskClaim, "claim", false, "Assign the task to --workflow-id and mark it in_progress")
	taskNextCmd.Flags().StringVar(&taskWorkflowID, "workflow-id", "", "Workflow claiming the task (required with --claim)")
	taskNextCmd.Flags().StringVar(&taskWorktreePath, "worktree-path", "", "Optional worktree path recorded on the claim")
}

func runTaskNext(cmd *cobra.Command, args []string) error {
	workflowID := strings.TrimSpace(taskWorkflowID)
	if taskClaim && workflowID == "" {
		return newTaskError("invalid_input", "--workflow-id is required with --claim")
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	if taskClaim {
		unlock, err := lockProject(reg, projectName)
		if err != nil {
			return err
		}
		defer unlock()
	}

	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}
	assignments, err := loadTaskAssignments(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}
	taskMap := make(map[string]*Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.Name] = t
	}

	action := "next"
	var picked *Task
	if taskClaim {
		action = "claim"
		if held := heldTask(tasks, assignments, workflowID); held != nil {
			picked, action = held, "resume"
		}
	}
	if picked == nil {
		picked = nextReadyTask(tasks, taskMap, assignments, taskFeature)
	}
	if picked == nil {
		return newTaskError("not_found", "no ready, unassigned task in project '%s'", projectName)
	}

	if action == "claim" {
		if err := setTaskStatus(picked, "in_progress"); err != nil {
			return err
		}
		if err := saveTask(picked); err != nil {
			return fmt.Errorf("failed to save task: %w", err)
		}
		assignments[picked.ID] = taskAssignment{
			TaskID:       picked.ID,
			TaskName:     picked.Name,
			WorkflowID:   workflowID,
			WorktreePath: strings.TrimSpace(taskWorktreePath),
			UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
		if err := saveTaskAssignments(reg, projectName, assignments); err != nil {
			return fmt.Errorf("failed to save assignments: %w", err)
		}
	}

	item := taskToJSONItem(picked, taskMap, assignments, true)
	if err := attachComments(reg, &item); err != nil {
		return fmt.Errorf("failed to load comments: %w", err)
	}
	if taskJSON {
		return printJSON(struct {
			OK      bool         `json:"ok"`
			Action  string       `json:"action"`
			Project string       `json:"project"`
			Task    taskJSONItem `json:"task"`
		}{
			OK:      true,
			Action:  action,
			Project: projectName,
			Task:    item,
		})
	}

	green := color.New(color.FgGreen).SprintFunc()
	switch action {
	case "claim":
		fmt.Printf("%s Claimed task '%s' for workflow '%s'\n", green("✓"), picked.Name, workflowID)
	case "resume":
		fmt.Printf("%s Workflow '%s' already holds task '%s'\n", green("✓"), workflowID, picked.Name)
	default:
		fmt.Printf("Next task: %s\n", picked.Name)
	}
	fmt.Printf("  %s\n", strings.TrimSpace(formatPriorityType(picked.Priority, picked.Type)+" "+picked.Subject))
	if picked.Content != "" {
		fmt.Printf("\n%s\n", picked.Content)
	}
	printTaskComments(item.Comments)
	return nil
}

// nextReadyTask returns the highest-priority ready task nobody is assigned to.
func nextReadyTask(tasks []*Task, taskMap map[string]*Task, assignments map[string]taskAssignment, feature string) *Task {
	if feature != "" {
		tasks = filterTasksByFeature(tasks, feature)
	}
	for _, t := range sortTasksByPriorityAndStatus(tasks, taskMap) {
		if _, assigned := assignments[t.ID]; assigned {
			continue
		}
		if computeTaskStatus(t, taskMap) == StatusReady {
			return t
		}
	}
	return nil
}

// heldTask returns an unfinished task already assigned to workflowID.
func heldTask(tasks []*Task, assignments map[string]taskAssignment, workflowID string) *Task {
	for _, t := range tasks {
		if a, ok := assignments[t.ID]; ok && a.WorkflowID == workflowID && t.Status != "completed" {
			return t
		}
	}
	return nil
}
