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

var addCmd = &cobra.Command{
	Use:   "add [type] [name]",
	Short: "Add a rule, workflow, or guideline to the project",
	Long: `Add a rule, workflow, or guideline from the registry to the current project.
This updates agents.toml and regenerates AGENTS.md.

Types:
  rule       - Add a rule
  workflow   - Add a workflow
  guideline  - Add a guideline

Examples:
  agmd add rule typescript       # Add TypeScript rule
  agmd add workflow release      # Add release workflow
  agmd add guideline code-style  # Add code style guideline`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	itemType := strings.ToLower(args[0])
	name := args[1]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Validate type
	if itemType != "rule" && itemType != "workflow" && itemType != "guideline" {
		return fmt.Errorf("invalid type '%s'. Must be: rule, workflow, or guideline", itemType)
	}

	fmt.Printf("%s Adding %s: %s\n", blue("→"), itemType, name)

	// Check if registry exists
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Verify item exists in registry
	fmt.Printf("%s Verifying %s exists in registry...\n", blue("→"), itemType)
	switch itemType {
	case "rule":
		if _, err := reg.GetRule(name); err != nil {
			return fmt.Errorf("%s '%s' not found in registry\nRun 'agmd new rule %s' to create it", itemType, name, name)
		}
	case "workflow":
		if _, err := reg.GetWorkflow(name); err != nil {
			return fmt.Errorf("%s '%s' not found in registry\nRun 'agmd new workflow %s' to create it", itemType, name, name)
		}
	case "guideline":
		if _, err := reg.GetGuideline(name); err != nil {
			return fmt.Errorf("%s '%s' not found in registry\nRun 'agmd new guideline %s' to create it", itemType, name, name)
		}
	}

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

	// Check if already added
	var alreadyAdded bool
	switch itemType {
	case "rule":
		alreadyAdded = st.HasRule(name)
	case "workflow":
		alreadyAdded = st.HasWorkflow(name)
	case "guideline":
		alreadyAdded = st.HasGuideline(name)
	}

	if alreadyAdded {
		fmt.Printf("%s %s '%s' is already in agents.toml\n", yellow("ℹ"), strings.Title(itemType), name)
		return nil
	}

	// Add to state
	switch itemType {
	case "rule":
		st.AddRule(name)
	case "workflow":
		st.AddWorkflow(name)
	case "guideline":
		st.AddGuideline(name)
	}

	// Save state
	fmt.Printf("%s Updating agents.toml...\n", blue("→"))
	if err := st.Save(agentsTomlFilename); err != nil {
		return fmt.Errorf("failed to save agents.toml: %w", err)
	}

	fmt.Printf("%s Updated agents.toml\n", green("✓"))

	// Regenerate AGENTS.md
	fmt.Printf("%s Regenerating AGENTS.md...\n", blue("→"))
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

	fmt.Printf("\n%s Added %s '%s' to project!\n", green("✓"), itemType, name)

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

	return nil
}
