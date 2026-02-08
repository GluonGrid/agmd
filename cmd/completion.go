package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

// completeTypeName provides completion for type:name arguments
// It completes types first, then items within a type after the colon
func completeTypeName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// If we already have an argument, no more completions
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Check if we're completing after a colon (type:name)
	if strings.Contains(toComplete, ":") {
		parts := strings.SplitN(toComplete, ":", 2)
		typeName := parts[0]
		namePrefix := parts[1]

		// Complete item names within the type
		items, err := reg.ListItems(typeName)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var completions []string
		for _, item := range items {
			if strings.HasPrefix(item.Name, namePrefix) {
				completions = append(completions, typeName+":"+item.Name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Complete type names (with colon suffix to indicate more to come)
	types, err := reg.ListTypes()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, t := range types {
		// Skip reserved types (they have their own subcommands)
		if t == "task" || t == "doc" || t == "file" {
			continue
		}
		if strings.HasPrefix(t, toComplete) {
			// Add colon to indicate there's more to complete
			completions = append(completions, t+":")
		}
	}

	// Don't add space after completion (user needs to type the name)
	return completions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

// completeTypeOnly provides completion for type-only arguments (like agmd list <type>)
func completeTypeOnly(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	types, err := reg.ListTypes()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, t := range types {
		// Skip reserved types (they have their own subcommands)
		if t == "task" || t == "doc" || t == "file" {
			continue
		}
		if strings.HasPrefix(t, toComplete) {
			completions = append(completions, t)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeTaskName provides completion for task names
func completeTaskName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// For commands that take exactly one task name
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
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

// completeProfileName provides completion for profile names
func completeProfileName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	profiles, err := reg.ListProfiles()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, p := range profiles {
		name := "profile:" + p.Name
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeFileName provides completion for file names in the registry
func completeFileName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	fileDir := filepath.Join(reg.BasePath, "file")
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := "file:" + entry.Name()
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
