package cmd

import (
	"fmt"
	"os"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// A task that is replaced by another (renamed, split, merged) must not stay
// open: `task replace` completes it, links both tasks and moves its
// dependents to the replacement in one locked step.

var taskReplaceBy string
var taskReplaceAuthor string

var taskReplaceCmd = &cobra.Command{
	Use:   "replace <old-task> --by <new-task>",
	Short: "Close a task as replaced by another and move its dependents",
	Long: `Close a task because another task replaces it.

The old task is completed and records "replaced_by: <new>"; the new task
records the old one under "replaces". Tasks that depend on the old task now
depend on the new one instead, so they stay blocked until the replacement is
done. A claim on the old task is released. Both tasks get a comment (author
from --author, $AGMD_AUTHOR, or "user").

Fails with "invalid_transition" (status 5) if the old task is already
completed, and with "conflict" (status 6) if moving a dependent would make
the new task depend on itself.

Examples:
  agmd task replace human-merge-all --by human-merge-pr-12
  agmd task replace old-name --by new-name --author app --json`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskReplace,
}

func init() {
	taskCmd.AddCommand(taskReplaceCmd)
	taskReplaceCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskReplaceCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskReplaceCmd.Flags().StringVar(&taskReplaceBy, "by", "", "The task that replaces it (required)")
	taskReplaceCmd.Flags().StringVar(&taskReplaceAuthor, "author", "", "Who replaces it, recorded in the comments (default $AGMD_AUTHOR or \"user\")")
}

// taskReplaceResult is what a replace changed, for output.
type taskReplaceResult struct {
	Old   *Task
	New   *Task
	Moved []string // dependents now blocked by the new task
}

// replaceTask checks and applies a replace. The caller must hold the project
// lock, so concurrent claimers see the old and new state, never a mix.
func replaceTask(reg *registry.Registry, projectName string, oldTask, newTask *Task, author string) (*taskReplaceResult, error) {
	if oldTask.Name == newTask.Name {
		return nil, newTaskError("invalid_input", "task '%s' cannot replace itself", oldTask.Name)
	}
	if oldTask.Status == "completed" {
		return nil, newTaskError("invalid_transition", "task '%s' is already completed; only open tasks can be replaced", oldTask.Name)
	}
	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	taskMap := buildTaskMap(tasks)

	var dependents []*Task
	for _, t := range tasks {
		if t.Name == oldTask.Name || t.Name == newTask.Name || !dependsOnTask(t, oldTask) {
			continue
		}
		if dependsTransitively(newTask, t, taskMap, map[string]bool{}) {
			return nil, newTaskError("conflict", "cannot replace '%s' by '%s': '%s' depends on '%s' and '%s' depends on it", oldTask.Name, newTask.Name, t.Name, oldTask.Name, newTask.Name)
		}
		dependents = append(dependents, t)
	}

	result := &taskReplaceResult{Old: oldTask, New: newTask}
	for _, t := range dependents {
		t.DependsOn = moveDependency(t.DependsOn, oldTask, newTask.Name)
		if err := saveTask(t); err != nil {
			return nil, fmt.Errorf("failed to save task: %w", err)
		}
		result.Moved = append(result.Moved, t.Name)
	}

	oldTask.Status = "completed"
	oldTask.CompletedAt = taskTimestamp(taskNow())
	oldTask.CompletionCountsAt = ""
	oldTask.ReplacedBy = newTask.Name
	if err := saveTask(oldTask); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}
	if !containsString(newTask.Replaces, oldTask.Name) {
		newTask.Replaces = append(newTask.Replaces, oldTask.Name)
	}
	if err := saveTask(newTask); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}
	if err := dropTaskAssignments(reg, projectName, oldTask.ID); err != nil {
		return nil, fmt.Errorf("failed to update assignments: %w", err)
	}

	oldNote := fmt.Sprintf("Replaced by %s; closed.", newTask.Name)
	if len(result.Moved) > 0 {
		oldNote += fmt.Sprintf(" Dependents moved to %s: %s.", newTask.Name, strings.Join(result.Moved, ", "))
	}
	if _, err := addTaskComment(reg, projectName, oldTask.ID, author, oldNote); err != nil {
		return nil, err
	}
	if _, err := addTaskComment(reg, projectName, newTask.ID, author, fmt.Sprintf("Replaces %s.", oldTask.Name)); err != nil {
		return nil, err
	}
	return result, nil
}

// moveDependency swaps every reference to oldTask for newName, without
// duplicating newName.
func moveDependency(deps []string, oldTask *Task, newName string) []string {
	moved := make([]string, 0, len(deps))
	seen := map[string]bool{}
	for _, ref := range deps {
		if ref == oldTask.Name || (oldTask.ID != "" && ref == oldTask.ID) {
			ref = newName
		}
		if !seen[ref] {
			seen[ref] = true
			moved = append(moved, ref)
		}
	}
	return moved
}

// dependsTransitively reports whether from depends, directly or through
// other tasks, on target.
func dependsTransitively(from, target *Task, taskMap map[string]*Task, seen map[string]bool) bool {
	if seen[from.Name] {
		return false
	}
	seen[from.Name] = true
	for _, ref := range from.DependsOn {
		dep := taskMap[ref]
		if dep == nil {
			continue
		}
		if dep.Name == target.Name || dependsTransitively(dep, target, taskMap, seen) {
			return true
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func runTaskReplace(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(taskReplaceBy) == "" {
		return newTaskError("invalid_input", "--by <new-task> is required")
	}
	author := taskReplaceAuthor
	if author == "" {
		author = os.Getenv("AGMD_AUTHOR")
	}
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, oldRef, err := resolveTaskProjectAndRef(args[0])
	if err != nil {
		return err
	}
	newProject, newRef, explicit := parseTaskRef(taskReplaceBy)
	if explicit && newProject != projectName {
		return newTaskError("project_mismatch", "replacement '%s' is in project '%s', not '%s'", taskReplaceBy, newProject, projectName)
	}
	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return err
	}
	defer unlock()

	oldTask, err := resolveTaskRef(reg, projectName, oldRef)
	if err != nil {
		return err
	}
	newTask, err := resolveTaskRef(reg, projectName, newRef)
	if err != nil {
		return err
	}
	result, err := replaceTask(reg, projectName, oldTask, newTask, author)
	if err != nil {
		return err
	}

	if taskJSON {
		return printJSON(struct {
			taskMutationJSONPayload
			ReplacedBy string   `json:"replaced_by"`
			Moved      []string `json:"moved_dependents"`
		}{
			taskMutationJSONPayload: taskMutationJSONPayload{
				OK:      true,
				Action:  "replace",
				Project: projectName,
				Task:    oldTask.Name,
				ID:      oldTask.ID,
				Status:  oldTask.Status,
			},
			ReplacedBy: newTask.Name,
			Moved:      append([]string{}, result.Moved...),
		})
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Closed task '%s', replaced by '%s'\n", green("✓"), oldTask.Name, newTask.Name)
	if len(result.Moved) > 0 {
		fmt.Printf("  dependents moved to '%s': %s\n", newTask.Name, strings.Join(result.Moved, ", "))
	}
	return nil
}
