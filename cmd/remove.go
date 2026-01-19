package cmd

import (
	"fmt"
	"os"
	"strings"

	"agmd/pkg/generator"
	"agmd/pkg/registry"
	"agmd/pkg/state"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [type] [name]",
	Short: "Remove a rule, workflow, or guideline from the project",
	Long: `Remove a rule, workflow, or guideline from the current project.
This updates agents.toml and regenerates AGENTS.md.

Note: This only removes the item from the project configuration.
The item remains in the registry and can be added again later.

Types:
  rule       - Remove a rule
  workflow   - Remove a workflow
  guideline  - Remove a guideline

Examples:
  agmd remove rule typescript       # Remove TypeScript rule
  agmd remove workflow release      # Remove release workflow
  agmd remove guideline code-style  # Remove code style guideline`,
	Args: cobra.ExactArgs(2),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	itemType := strings.ToLower(args[0])
	name := args[1]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Validate type
	if itemType != "rule" && itemType != "workflow" && itemType != "guideline" {
		return fmt.Errorf("invalid type '%s'. Must be: rule, workflow, or guideline", itemType)
	}

	fmt.Printf("%s Removing %s: %s\n", blue("→"), itemType, name)

	// Check if agents.toml exists
	if _, err := os.Stat(agentsTomlFilename); os.IsNotExist(err) {
		return fmt.Errorf("agents.toml not found\nRun 'agmd init' first to initialize project")
	}

	// Load state
	fmt.Printf("%s Loading project state...\n", blue("→"))
	st, err := state.Load(agentsTomlFilename)
	if err != nil {
		return fmt.Errorf("failed to load agents.toml: %w", err)
	}

	// Check if item exists in project
	var exists bool
	switch itemType {
	case "rule":
		exists = st.HasRule(name)
	case "workflow":
		exists = st.HasWorkflow(name)
	case "guideline":
		exists = st.HasGuideline(name)
	}

	if !exists {
		fmt.Printf("%s %s '%s' is not in agents.toml\n", yellow("ℹ"), strings.Title(itemType), name)
		return nil
	}

	// Prevent removing the required no-modify-agmd-sections rule
	if itemType == "rule" && name == "no-modify-agmd-sections" {
		return fmt.Errorf("%s Cannot remove 'no-modify-agmd-sections' rule\nThis rule is required for agmd to work correctly with AI assistants", red("✗"))
	}

	// Remove from state
	switch itemType {
	case "rule":
		st.RemoveRule(name)
	case "workflow":
		st.RemoveWorkflow(name)
	case "guideline":
		st.RemoveGuideline(name)
	}

	// Save state
	fmt.Printf("%s Updating agents.toml...\n", blue("→"))
	if err := st.Save(agentsTomlFilename); err != nil {
		return fmt.Errorf("failed to save agents.toml: %w", err)
	}

	fmt.Printf("%s Updated agents.toml\n", green("✓"))

	// Regenerate AGENTS.md
	fmt.Printf("%s Regenerating AGENTS.md...\n", blue("→"))
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	gen := generator.New(reg, st)

	var content string
	if _, err := os.Stat(agentsMdFilename); err == nil {
		// Preserve custom section
		content, err = gen.GeneratePreservingCustom(agentsMdFilename)
	} else {
		content, err = gen.Generate()
	}

	if err != nil {
		return fmt.Errorf("failed to generate AGENTS.md: %w", err)
	}

	if err := os.WriteFile(agentsMdFilename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("%s Regenerated AGENTS.md\n", green("✓"))

	fmt.Printf("\n%s Removed %s '%s' from project\n", green("✓"), itemType, name)

	// Show updated summary
	fmt.Println("\nActive content:")
	if len(st.Rules) > 0 {
		fmt.Printf("  • Rules: %d\n", len(st.Rules))
		for _, rule := range st.Rules {
			fmt.Printf("    - %s\n", rule)
		}
	}
	if len(st.Workflows) > 0 {
		fmt.Printf("  • Workflows: %d\n", len(st.Workflows))
		for _, workflow := range st.Workflows {
			fmt.Printf("    - %s\n", workflow)
		}
	}
	if len(st.Guidelines) > 0 {
		fmt.Printf("  • Guidelines: %d\n", len(st.Guidelines))
		for _, guideline := range st.Guidelines {
			fmt.Printf("    - %s\n", guideline)
		}
	}

	if len(st.Rules) == 0 && len(st.Workflows) == 0 && len(st.Guidelines) == 0 {
		fmt.Printf("  %s No active content. Use 'agmd add rule <name>' to add rules.\n", yellow("ℹ"))
	}

	return nil
}
