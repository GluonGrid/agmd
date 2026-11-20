package cmd

import (
	"strings"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

// completeTaskName provides completion for task names
func completeTaskName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// For commands that take exactly one task name
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectName, err := getProjectName()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, t := range tasks {
		if strings.HasPrefix(t.Name, toComplete) {
			completions = append(completions, t.Name)
		}
		if t.ID != "" && strings.HasPrefix(t.ID, toComplete) {
			completions = append(completions, t.ID)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeTaskStatus provides completion for task status values
func completeTaskStatus(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Task status command takes: task-name status
	// Complete task name first, then status
	if len(args) == 0 {
		return completeTaskName(cmd, args, toComplete)
	}
	if len(args) == 1 {
		statuses := []string{"pending", "in_progress", "completed"}
		var completions []string
		for _, s := range statuses {
			if strings.HasPrefix(s, toComplete) {
				completions = append(completions, s)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeTaskDependency provides completion for task dependency commands (blocked-by, unblock)
func completeTaskDependency(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// These commands take: task-name dependency-name
	// Complete task names for both arguments
	if len(args) < 2 {
		return completeTaskName(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeListStatusFlag provides completion for --status flag values
func completeListStatusFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	statuses := []string{"ready", "blocked", "in_progress", "completed"}
	var completions []string
	for _, s := range statuses {
		if strings.HasPrefix(s, toComplete) {
			completions = append(completions, s)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completePriorityFlag provides completion for --priority flag values
func completePriorityFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	priorities := []string{
		"0\tcritical",
		"1\thigh",
		"2\tmedium (default)",
		"3\tlow",
		"4\tbacklog",
	}
	var completions []string
	for _, p := range priorities {
		if strings.HasPrefix(p, toComplete) || strings.HasPrefix("P"+p, strings.ToUpper(toComplete)) {
			completions = append(completions, p)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeTypeFlag provides completion for --type flag values
func completeTypeFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	types := []string{
		"bug\tDefect to fix",
		"feature\tNew functionality",
		"task\tGeneral task (default)",
		"chore\tMaintenance work",
	}
	var completions []string
	for _, t := range types {
		if strings.HasPrefix(t, toComplete) {
			completions = append(completions, t)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
