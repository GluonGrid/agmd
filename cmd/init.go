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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project with agents.toml and AGENTS.md",
	Long: `Initialize a new project by creating:
- agents.toml (with default no-modify-agmd-sections rule)
- AGENTS.md (generated from registry and state)

The no-modify-agmd-sections rule is always included to ensure
AI assistants understand the managed vs custom section boundaries.

Examples:
  agmd init           # Initialize in current directory`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("%s Initializing agmd project...\n", blue("→"))

	// Check if registry exists
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\n%s\nRun 'agmd setup' first to initialize the registry",
			red(reg.BasePath),
			red("✗"))
	}

	// Check if agents.toml already exists
	if _, err := os.Stat(agentsTomlFilename); err == nil {
		return fmt.Errorf("agents.toml already exists in current directory")
	}

	// Check if AGENTS.md already exists
	if _, err := os.Stat(agentsMdFilename); err == nil {
		return fmt.Errorf("AGENTS.md already exists in current directory")
	}

	// Create default state (always includes no-modify-agmd-sections)
	fmt.Printf("%s Creating agents.toml with default configuration...\n", blue("→"))
	st := state.DefaultState()

	// Save agents.toml
	if err := st.Save(agentsTomlFilename); err != nil {
		return fmt.Errorf("failed to create agents.toml: %w", err)
	}

	fmt.Printf("%s Generated agents.toml\n", green("✓"))

	// Generate AGENTS.md
	fmt.Printf("%s Generating AGENTS.md...\n", blue("→"))
	gen := generator.New(reg, st)
	content, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate AGENTS.md: %w", err)
	}

	if err := os.WriteFile(agentsMdFilename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("%s Generated AGENTS.md\n", green("✓"))

	fmt.Printf("\n%s Project initialized successfully!\n", green("✓"))
	fmt.Println("\nCreated:")
	fmt.Printf("  • %s - Project configuration\n", agentsTomlFilename)
	fmt.Printf("  • %s - AI agent instructions\n", agentsMdFilename)

	fmt.Println("\nDefault rules included:")
	fmt.Println("  • no-modify-agmd-sections (required for agmd to work correctly)")

	fmt.Println("\nNext steps:")
	fmt.Println("  • Edit the custom section in AGENTS.md to add project-specific notes")
	fmt.Println("  • Run 'agmd add rule <name>' to add more rules")
	fmt.Println("  • Run 'agmd new rule <name>' to create custom rules")

	return nil
}
