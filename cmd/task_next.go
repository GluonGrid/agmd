package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var taskClaim bool
var taskStaleAfter string

// defaultClaimTTL is how long a claim survives without a heartbeat before
// another workflow may take the task over.
const defaultClaimTTL = 2 * time.Hour

var taskNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show or claim the next ready task",
	Long: `Show the highest-priority task that is ready (all dependencies completed)
and not assigned to any workflow.

With --claim, atomically assign that task to --workflow-id and mark it
in_progress, so parallel agents never pick the same task. If the workflow
already holds an unfinished task, that task is returned instead, so an agent
that restarts resumes its work rather than claiming more.

A claim expires when its workflow shows no sign of life for --stale-after
(default $AGMD_CLAIM_TTL, else 2h; 0 disables expiry). Claiming, resuming,
'agmd task heartbeat', and commenting with --author set to the workflow id
all count as signs of life. An expired claim's task can be claimed again; the
takeover is recorded as a comment on the task.

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
	taskNextCmd.Flags().StringVar(&taskStaleAfter, "stale-after", "", "Treat claims idle this long as expired, e.g. 30m (default $AGMD_CLAIM_TTL or 2h; 0 = never)")

	taskCmd.AddCommand(taskHeartbeatCmd)
	taskHeartbeatCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskHeartbeatCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskHeartbeatCmd.Flags().StringVar(&taskWorkflowID, "workflow-id", "", "Workflow whose claims to keep alive")
	taskHeartbeatCmd.MarkFlagRequired("workflow-id") //nolint:errcheck
}

var taskHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat [task-id-or-name]",
	Short: "Keep a workflow's claims from expiring",
	Long: `Refresh the claims held by --workflow-id so other agents don't treat them as
stale. With a task argument, only that claim is refreshed, and the command
fails with code "conflict" (status 6) if another workflow now holds the task,
or "not_found" (status 3) if nobody does: the agent has lost its claim.

Examples:
  agmd task heartbeat --workflow-id agent-1
  agmd task heartbeat setup-db --workflow-id agent-1 --json`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskHeartbeat,
}

func runTaskNext(cmd *cobra.Command, args []string) error {
	workflowID := strings.TrimSpace(taskWorkflowID)
	if taskClaim && workflowID == "" {
		return newTaskError("invalid_input", "--workflow-id is required with --claim")
	}

	ttl, err := claimTTL()
	if err != nil {
		return err
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
	now := time.Now().UTC()
	if picked == nil {
		picked = nextReadyTask(tasks, taskMap, assignments, taskFeature, ttl, now)
	}
	if picked == nil {
		return newTaskError("not_found", "no ready, unassigned task in project '%s'", projectName)
	}

	var previous *taskAssignment
	if prev, held := assignments[picked.ID]; held && action == "claim" {
		previous, action = &prev, "takeover"
	}

	switch action {
	case "resume":
		refreshClaim(assignments, picked.ID, now)
		if err := saveTaskAssignments(reg, projectName, assignments); err != nil {
			return fmt.Errorf("failed to save assignments: %w", err)
		}
	case "claim", "takeover":
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
			UpdatedAt:    now.Format(time.RFC3339),
		}
		if err := saveTaskAssignments(reg, projectName, assignments); err != nil {
			return fmt.Errorf("failed to save assignments: %w", err)
		}
		if previous != nil {
			note := fmt.Sprintf("Took over this task from workflow '%s' (no sign of life since %s).", previous.WorkflowID, previous.UpdatedAt)
			if _, err := addTaskComment(reg, projectName, picked.ID, workflowID, note); err != nil {
				return err
			}
		}
	}

	item := taskToJSONItem(picked, taskMap, assignments, true)
	if err := attachComments(reg, &item); err != nil {
		return fmt.Errorf("failed to load comments: %w", err)
	}
	if taskJSON {
		payload := struct {
			OK            bool         `json:"ok"`
			Action        string       `json:"action"`
			Project       string       `json:"project"`
			TakenOverFrom string       `json:"taken_over_from,omitempty"`
			Task          taskJSONItem `json:"task"`
		}{
			OK:      true,
			Action:  action,
			Project: projectName,
			Task:    item,
		}
		if previous != nil {
			payload.TakenOverFrom = previous.WorkflowID
		}
		return printJSON(payload)
	}

	green := color.New(color.FgGreen).SprintFunc()
	switch action {
	case "claim":
		fmt.Printf("%s Claimed task '%s' for workflow '%s'\n", green("✓"), picked.Name, workflowID)
	case "resume":
		fmt.Printf("%s Workflow '%s' already holds task '%s'\n", green("✓"), workflowID, picked.Name)
	case "takeover":
		fmt.Printf("%s Took over stale task '%s' from workflow '%s' for '%s'\n", green("✓"), picked.Name, previous.WorkflowID, workflowID)
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

// nextReadyTask returns the highest-priority task that is either ready and
// unassigned, or unfinished with an expired claim.
func nextReadyTask(tasks []*Task, taskMap map[string]*Task, assignments map[string]taskAssignment, feature string, ttl time.Duration, now time.Time) *Task {
	if feature != "" {
		tasks = filterTasksByFeature(tasks, feature)
	}
	for _, t := range sortTasksByPriorityAndStatus(tasks, taskMap) {
		if a, assigned := assignments[t.ID]; assigned {
			if t.Status != "completed" && claimIsStale(a, ttl, now) {
				return t
			}
			continue
		}
		if computeTaskStatus(t, taskMap) == StatusReady {
			return t
		}
	}
	return nil
}

// claimTTL resolves --stale-after, then $AGMD_CLAIM_TTL, then the default.
// Zero disables expiry.
func claimTTL() (time.Duration, error) {
	raw := strings.TrimSpace(taskStaleAfter)
	source := "--stale-after"
	if raw == "" {
		raw, source = strings.TrimSpace(os.Getenv("AGMD_CLAIM_TTL")), "AGMD_CLAIM_TTL"
	}
	if raw == "" {
		return defaultClaimTTL, nil
	}
	if raw == "0" {
		return 0, nil
	}
	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl < 0 {
		return 0, newTaskError("invalid_input", "invalid %s '%s': use a duration like 30m or 2h, or 0 to disable", source, raw)
	}
	return ttl, nil
}

// claimIsStale reports whether a claim has gone ttl without a sign of life.
// Claims with an unreadable timestamp never expire.
func claimIsStale(a taskAssignment, ttl time.Duration, now time.Time) bool {
	if ttl <= 0 {
		return false
	}
	updated, err := time.Parse(time.RFC3339, a.UpdatedAt)
	if err != nil {
		return false
	}
	return now.Sub(updated) >= ttl
}

// refreshClaim records a sign of life on a claim.
func refreshClaim(assignments map[string]taskAssignment, taskID string, now time.Time) {
	if a, ok := assignments[taskID]; ok {
		a.UpdatedAt = now.UTC().Format(time.RFC3339)
		assignments[taskID] = a
	}
}

func runTaskHeartbeat(cmd *cobra.Command, args []string) error {
	workflowID := strings.TrimSpace(taskWorkflowID)
	if workflowID == "" {
		return newTaskError("invalid_input", "--workflow-id is required")
	}
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	var projectName, taskRef string
	if len(args) == 1 {
		projectName, taskRef, err = resolveTaskProjectAndRef(args[0])
	} else {
		projectName, err = getProjectName()
	}
	if err != nil {
		return err
	}
	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return err
	}
	defer unlock()

	assignments, err := loadTaskAssignments(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}
	now := time.Now().UTC()
	var refreshed []string
	if taskRef != "" {
		task, err := resolveTaskRef(reg, projectName, taskRef)
		if err != nil {
			return err
		}
		a, held := assignments[task.ID]
		if !held {
			return newTaskError("not_found", "task '%s' is not claimed by any workflow", task.Name)
		}
		if a.WorkflowID != workflowID {
			return newTaskError("conflict", "task '%s' is now held by workflow '%s'", task.Name, a.WorkflowID)
		}
		refreshClaim(assignments, task.ID, now)
		refreshed = append(refreshed, task.Name)
	} else {
		for id, a := range assignments {
			if a.WorkflowID == workflowID {
				refreshClaim(assignments, id, now)
				refreshed = append(refreshed, a.TaskName)
			}
		}
		if len(refreshed) == 0 {
			return newTaskError("not_found", "workflow '%s' holds no tasks in project '%s'", workflowID, projectName)
		}
	}
	if err := saveTaskAssignments(reg, projectName, assignments); err != nil {
		return fmt.Errorf("failed to save assignments: %w", err)
	}
	sort.Strings(refreshed)

	if taskJSON {
		return printJSON(struct {
			OK         bool     `json:"ok"`
			Action     string   `json:"action"`
			Project    string   `json:"project"`
			WorkflowID string   `json:"workflow_id"`
			Tasks      []string `json:"tasks"`
		}{
			OK:         true,
			Action:     "heartbeat",
			Project:    projectName,
			WorkflowID: workflowID,
			Tasks:      refreshed,
		})
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Refreshed %d claim(s) for workflow '%s': %s\n", green("✓"), len(refreshed), workflowID, strings.Join(refreshed, ", "))
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
