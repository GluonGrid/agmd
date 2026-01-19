package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/generator"
	"agmd/pkg/registry"
	"agmd/pkg/state"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	agentsMdFilename  = "AGENTS.md"
	agentsTomlFilename = "agents.toml"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate AGENTS.md from registry and state",
	Long: `Generate or regenerate AGENTS.md based on:
- Active rules/workflows/guidelines in agents.toml
- Content from ~/.agmd/ registry
- Preserving custom sections

Examples:
  agmd generate           # Regenerate AGENTS.md`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Generating AGENTS.md...\n", blue("→"))

	// Load state
	fmt.Printf("%s Loading agents.toml...\n", blue("→"))
	st, err := state.Load(agentsTomlFilename)
	if err != nil {
		return fmt.Errorf("failed to load agents.toml: %w\nRun 'agmd init' first", err)
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Create generator
	gen := generator.New(reg, st)

	// Check if AGENTS.md exists
	var content string
	if _, err := os.Stat(agentsMdFilename); err == nil {
		// Exists - preserve custom section
		fmt.Printf("%s Preserving custom sections...\n", blue("→"))
		content, err = gen.GeneratePreservingCustom(agentsMdFilename)
		if err != nil {
			return fmt.Errorf("failed to generate AGENTS.md: %w", err)
		}
	} else {
		// Doesn't exist - create new
		content, err = gen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate AGENTS.md: %w", err)
		}
	}

	// Write AGENTS.md
	if err := os.WriteFile(agentsMdFilename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("\n%s Generated AGENTS.md successfully!\n", green("✓"))

	// Show summary
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
