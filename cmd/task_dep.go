package cmd

import (
	"fmt"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// `task dep add|remove` change dependency links, one per call, under the
// project lock. `task blocked-by` and `task unblock` are the older names,
// kept for callers that use them; they report their own JSON action.

var taskDepCmd = &cobra.Command{
	Use:   "dep",
	Short: "Add or remove task dependencies",
	Long: `Add or remove a dependency between two tasks of the same project.

A task with an unfinished dependency is blocked and never handed out.

Examples:
  agmd task dep add create-api setup-db       # create-api waits for setup-db
  agmd task dep remove create-api setup-db    # create-api no longer waits`,
}

var taskDepAddCmd = &cobra.Command{
	Use:   "add <task-id-or-name> <dependency-id-or-name>",
	Short: "Make a task depend on another",
	Long: `Make <task> depend on <dependency>: it stays blocked until the dependency is done.

Fails with "invalid_input" (status 2) if the link exists already or the task
would depend on itself, and with "conflict" (status 6) if the dependency
already depends on the task (a cycle).

Examples:
  agmd task dep add create-api setup-db       # create-api depends on setup-db`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskDependency,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDepAdd(args, "dep-add")
	},
}

var taskDepRemoveCmd = &cobra.Command{
	Use:     "remove <task-id-or-name> <dependency-id-or-name>",
	Aliases: []string{"rm"},
	Short:   "Remove a dependency from a task",
	Long: `Remove the dependency of <task> on <dependency>.

Fails with "not_found" (status 3) if there is no such link.

Examples:
  agmd task dep remove create-api setup-db    # Remove setup-db dependency from create-api`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskDependency,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDepRemove(args, "dep-remove")
	},
}

var taskBlockedByCmd = &cobra.Command{
	Use:   "blocked-by <task-id-or-name> <dependency-id-or-name>",
	Short: "Alias of 'dep add' (compatibility)",
	Long: `Compatibility alias of 'agmd task dep add'. Its JSON reports "action": "blocked-by".

Examples:
  agmd task blocked-by create-api setup-db    # Same as: agmd task dep add create-api setup-db`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskDependency,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDepAdd(args, "blocked-by")
	},
}

var taskUnblockCmd = &cobra.Command{
	Use:   "unblock <task-id-or-name> <dependency-id-or-name>",
	Short: "Alias of 'dep remove' (compatibility)",
	Long: `Compatibility alias of 'agmd task dep remove'. Its JSON reports "action": "unblock".

Examples:
  agmd task unblock create-api setup-db       # Same as: agmd task dep remove create-api setup-db`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeTaskDependency,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDepRemove(args, "unblock")
	},
}

func init() {
	taskCmd.AddCommand(taskDepCmd)
	taskDepCmd.AddCommand(taskDepAddCmd)
	taskDepCmd.AddCommand(taskDepRemoveCmd)
	taskCmd.AddCommand(taskBlockedByCmd)
	taskCmd.AddCommand(taskUnblockCmd)
	for _, c := range []*cobra.Command{taskDepAddCmd, taskDepRemoveCmd, taskBlockedByCmd, taskUnblockCmd} {
		c.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
		c.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	}
}

// withDependencyPair resolves a task and a dependency of the same project,
// then calls change under the project lock and saves the task.
func withDependencyPair(args []string, change func(task, dep *Task, reg *registry.Registry, projectName string) error) (string, *Task, *Task, error) {
	reg, err := registry.New()
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, taskRef, err := resolveTaskProjectAndRef(args[0])
	if err != nil {
		return "", nil, nil, err
	}
	dependencyProject, dependencyRef, hasDependencyProject := parseTaskRef(args[1])
	if hasDependencyProject && dependencyProject != projectName {
		return "", nil, nil, newTaskError("project_mismatch", "dependency reference '%s' does not match project '%s'", dependencyProject, projectName)
	}

	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return "", nil, nil, err
	}
	defer unlock()

	task, err := resolveTaskRef(reg, projectName, taskRef)
	if err != nil {
		return "", nil, nil, err
	}
	dep, err := resolveTaskRef(reg, projectName, dependencyRef)
	if err != nil {
		return "", nil, nil, newTaskError("not_found", "dependency task '%s' not found in project '%s'", args[1], projectName)
	}
	if err := change(task, dep, reg, projectName); err != nil {
		return "", nil, nil, err
	}
	if err := saveTask(task); err != nil {
		return "", nil, nil, fmt.Errorf("failed to save task: %w", err)
	}
	return projectName, task, dep, nil
}

func runTaskDepAdd(args []string, action string) error {
	projectName, task, dep, err := withDependencyPair(args, func(task, dep *Task, reg *registry.Registry, projectName string) error {
		if dep.Name == task.Name {
			return newTaskError("invalid_input", "task '%s' cannot depend on itself", task.Name)
		}
		if containsString(task.DependsOn, dep.Name) {
			return newTaskError("invalid_input", "task '%s' already depends on '%s'", task.Name, dep.Name)
		}
		tasks, err := loadProjectTasks(reg, projectName)
		if err != nil {
			return err
		}
		taskMap := make(map[string]*Task, len(tasks))
		for _, t := range tasks {
			taskMap[t.Name] = t
		}
		if dependsTransitively(dep, task, taskMap, map[string]bool{}) {
			return newTaskError("conflict", "'%s' already depends on '%s': adding this link would create a cycle", dep.Name, task.Name)
		}
		task.DependsOn = append(task.DependsOn, dep.Name)
		return nil
	})
	if err != nil {
		return err
	}
	if taskJSON {
		return printDepJSON(action, projectName, task)
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Added dependency: '%s' is now blocked by '%s'\n", green("✓"), task.Name, dep.Name)
	return nil
}

func runTaskDepRemove(args []string, action string) error {
	projectName, task, dep, err := withDependencyPair(args, func(task, dep *Task, reg *registry.Registry, projectName string) error {
		if !containsString(task.DependsOn, dep.Name) {
			return newTaskError("not_found", "task '%s' does not depend on '%s'", task.Name, dep.Name)
		}
		remaining := []string{}
		for _, name := range task.DependsOn {
			if name != dep.Name {
				remaining = append(remaining, name)
			}
		}
		task.DependsOn = remaining
		return nil
	})
	if err != nil {
		return err
	}
	if taskJSON {
		return printDepJSON(action, projectName, task)
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Removed dependency: '%s' is no longer blocked by '%s'\n", green("✓"), task.Name, dep.Name)
	return nil
}

func printDepJSON(action, projectName string, task *Task) error {
	return printJSON(taskMutationJSONPayload{
		OK:      true,
		Action:  action,
		Project: projectName,
		Task:    task.Name,
		ID:      task.ID,
		Status:  task.Status,
	})
}
