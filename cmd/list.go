package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/registry"
	"agmd/pkg/state"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and active rules, workflows, and guidelines",
	Long: `List all rules, workflows, and guidelines in the registry.
Shows which items are active in the current project (if agents.toml exists).

Examples:
  agmd list           # List all available content`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Load project state if it exists
	var st *state.ProjectState
	hasProject := false
	if _, err := os.Stat(agentsTomlFilename); err == nil {
		st, err = state.Load(agentsTomlFilename)
		if err == nil {
			hasProject = true
		}
	}

	fmt.Printf("%s Registry location: %s\n\n", blue("ℹ"), reg.BasePath)

	// List rules
	rules, err := reg.ListRules()
	if err != nil {
		return fmt.Errorf("failed to list rules: %w", err)
	}

	fmt.Printf("%s Rules (%d):\n", cyan("●"), len(rules))
	if len(rules) == 0 {
		fmt.Printf("  %s No rules in registry\n", yellow("ℹ"))
	} else {
		for _, rule := range rules {
			isActive := hasProject && st.HasRule(rule.Name)
			if isActive {
				fmt.Printf("  %s %s (active)\n", green("✓"), rule.Name)
			} else {
				fmt.Printf("    %s\n", rule.Name)
			}
			if rule.Description != "" {
				fmt.Printf("      %s\n", rule.Description)
			}
		}
	}
	fmt.Println()

	// List workflows
	workflows, err := reg.ListWorkflows()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	fmt.Printf("%s Workflows (%d):\n", cyan("●"), len(workflows))
	if len(workflows) == 0 {
		fmt.Printf("  %s No workflows in registry\n", yellow("ℹ"))
	} else {
		for _, workflow := range workflows {
			isActive := hasProject && st.HasWorkflow(workflow.Name)
			if isActive {
				fmt.Printf("  %s %s (active)\n", green("✓"), workflow.Name)
			} else {
				fmt.Printf("    %s\n", workflow.Name)
			}
			if workflow.Description != "" {
				fmt.Printf("      %s\n", workflow.Description)
			}
		}
	}
	fmt.Println()

	// List guidelines
	guidelines, err := reg.ListGuidelines()
	if err != nil {
		return fmt.Errorf("failed to list guidelines: %w", err)
	}

	fmt.Printf("%s Guidelines (%d):\n", cyan("●"), len(guidelines))
	if len(guidelines) == 0 {
		fmt.Printf("  %s No guidelines in registry\n", yellow("ℹ"))
	} else {
		for _, guideline := range guidelines {
			isActive := hasProject && st.HasGuideline(guideline.Name)
			if isActive {
				fmt.Printf("  %s %s (active)\n", green("✓"), guideline.Name)
			} else {
				fmt.Printf("    %s\n", guideline.Name)
			}
			if guideline.Description != "" {
				fmt.Printf("      %s\n", guideline.Description)
			}
		}
	}
	fmt.Println()

	if hasProject {
		fmt.Printf("%s Active in current project: %d rules, %d workflows, %d guidelines\n",
			blue("ℹ"), len(st.Rules), len(st.Workflows), len(st.Guidelines))
	} else {
		fmt.Printf("%s No project found in current directory (run 'agmd init' to create one)\n", yellow("ℹ"))
	}

	return nil
}
