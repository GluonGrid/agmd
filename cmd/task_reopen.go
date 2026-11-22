package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// A completed task is final once something was built on it. Until then the
// human can undo a completion with `task reopen`, and `task status completed
// --hold` delays the moment a completion releases dependents, so a mis-tap
// can be undone before any agent claims the next task.

// maxCompletionHold caps --hold: it is an undo window, not a scheduler.
const maxCompletionHold = time.Hour

// taskNow is the clock for timestamps and hold checks.
var taskNow = time.Now

var taskHold string
var taskReopenAuthor string

var taskReopenCmd = &cobra.Command{
	Use:   "reopen <task-id-or-name>",
	Short: "Undo a completion: put a completed task back to pending",
	Long: `Put a completed task back to pending, for example after a mis-tap.

Allowed only while nothing was built on the completion: no task that depends
on it is claimed, in progress or completed. Dependents that are merely ready
become blocked again. Otherwise the command fails with code "conflict"
(status 6) naming the dependent. A task that is not completed fails with
"invalid_transition" (status 5).

The reopen is recorded as a comment on the task (author from --author,
$AGMD_AUTHOR, or "user"). Agents should not reopen tasks they did not
complete themselves.

Examples:
  agmd task reopen merge-pr
  agmd task reopen merge-pr --author app --json`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskReopen,
}

func init() {
	taskCmd.AddCommand(taskReopenCmd)
	taskReopenCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskReopenCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskReopenCmd.Flags().StringVar(&taskReopenAuthor, "author", "", "Who reopens it, recorded in the comment (default $AGMD_AUTHOR or \"user\")")

	taskStatusCmd.Flags().StringVar(&taskHold, "hold", "", "With completed: release dependents only after this delay, e.g. 10s or 2m (max 1h)")
}

// dependencySatisfied answers "is this dependency done" for every reader:
// computed status, next/claim, list filters, show and the web board. A
// missing dependency is not done, and neither is a completion still inside
// its --hold window.
func dependencySatisfied(dep *Task, now time.Time) bool {
	if dep == nil || dep.Status != "completed" {
		return false
	}
	_, held := completionHeldUntil(dep, now)
	return !held
}

// completionHeldUntil reports whether a completed task's hold is still
// running at now, and until when. An unreadable time counts as no hold.
func completionHeldUntil(task *Task, now time.Time) (time.Time, bool) {
	if task.Status != "completed" || task.CompletionCountsAt == "" {
		return time.Time{}, false
	}
	until, err := time.Parse(time.RFC3339, task.CompletionCountsAt)
	if err != nil || !now.Before(until) {
		return time.Time{}, false
	}
	return until, true
}

func taskTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// parseCompletionHold validates --hold. Zero means no hold.
func parseCompletionHold(raw, newStatus string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	hold, err := time.ParseDuration(raw)
	if err != nil || hold < 0 {
		return 0, newTaskError("invalid_input", "invalid --hold '%s': use a duration like 10s or 2m", raw)
	}
	if hold > maxCompletionHold {
		return 0, newTaskError("invalid_input", "--hold %s is longer than the maximum of 1h", raw)
	}
	if hold > 0 && newStatus != "completed" {
		return 0, newTaskError("invalid_input", "--hold only applies to status completed")
	}
	return hold, nil
}

// applyCompletionHold stores when the completion starts to count, rounded
// up to the second so the stored RFC 3339 time never shortens the hold.
func applyCompletionHold(task *Task, hold time.Duration, now time.Time) {
	until := now.Add(hold)
	if t := until.Truncate(time.Second); t.Before(until) {
		until = t.Add(time.Second)
	}
	task.CompletionCountsAt = taskTimestamp(until)
}

// formatHoldNote is the short "counts from" note shown next to a held task.
func formatHoldNote(until time.Time) string {
	return "counts from " + until.Local().Format("15:04:05")
}

// reopenBlocker returns why task cannot be reopened, or "" if it can: a
// direct dependent that is claimed, in progress or completed. The caller
// must hold the project lock and pass fresh data.
func reopenBlocker(task *Task, tasks []*Task, assignments map[string]taskAssignment) string {
	for _, t := range tasks {
		if t.Name == task.Name || !dependsOnTask(t, task) {
			continue
		}
		switch {
		case t.Status == "completed":
			return fmt.Sprintf("dependent task '%s' is already completed", t.Name)
		case t.Status == "in_progress":
			return fmt.Sprintf("dependent task '%s' is in progress", t.Name)
		}
		if a, claimed := assignments[t.ID]; claimed {
			return fmt.Sprintf("dependent task '%s' is claimed by workflow '%s'", t.Name, a.WorkflowID)
		}
	}
	return ""
}

func dependsOnTask(t, dep *Task) bool {
	for _, ref := range t.DependsOn {
		if ref == dep.Name || (dep.ID != "" && ref == dep.ID) {
			return true
		}
	}
	return false
}

// reopenTask checks and applies a reopen. The caller must hold the project
// lock, so the check and the write are one step for concurrent claimers.
func reopenTask(reg *registry.Registry, projectName string, task *Task, author string) error {
	if task.Status != "completed" {
		return newTaskError("invalid_transition", "task '%s' is %s, not completed; only completed tasks can be reopened", task.Name, task.Status)
	}
	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}
	assignments, err := loadTaskAssignments(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}
	if why := reopenBlocker(task, tasks, assignments); why != "" {
		return newTaskError("conflict", "cannot reopen task '%s': %s", task.Name, why)
	}

	task.Status = "pending"
	task.CompletedAt = ""
	task.CompletionCountsAt = ""
	if err := saveTask(task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	if _, err := addTaskComment(reg, projectName, task.ID, author, "Reopened: completed -> pending."); err != nil {
		return err
	}
	return nil
}

func runTaskReopen(cmd *cobra.Command, args []string) error {
	author := taskReopenAuthor
	if author == "" {
		author = os.Getenv("AGMD_AUTHOR")
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
	if err := reopenTask(reg, projectName, task, author); err != nil {
		return err
	}

	if taskJSON {
		return printJSON(taskMutationJSONPayload{
			OK:      true,
			Action:  "reopen",
			Project: projectName,
			Task:    task.Name,
			ID:      task.ID,
			Status:  task.Status,
		})
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Reopened task '%s' (back to pending)\n", green("✓"), task.Name)
	return nil
}
