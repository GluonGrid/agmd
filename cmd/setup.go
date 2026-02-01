package cmd

import (
	"fmt"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var setupForce bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize the agmd registry",
	Long: `Create the ~/.agmd directory structure and install default templates.

This command sets up:
- ~/.agmd/rule/ (coding rules)
- ~/.agmd/workflow/ (process workflows)
- ~/.agmd/guideline/ (best practices)
- ~/.agmd/profile/ (project templates)

Examples:
  agmd setup              # Initialize registry
  agmd setup --force      # Reinitialize (overwrites existing)`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupForce, "force", false, "Force reinitialize (overwrites existing)")
}

func runSetup(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Create registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Check if already exists
	if reg.Exists() && !setupForce {
		fmt.Printf("%s Registry already exists at: %s\n", yellow("ℹ"), reg.BasePath)
		fmt.Println("\nUse --force to reinitialize (this will overwrite existing files)")
		return nil
	}

	if setupForce && reg.Exists() {
		fmt.Printf("%s Reinitializing registry at: %s\n", yellow("⚠"), reg.BasePath)
	} else {
		fmt.Printf("%s Initializing registry at: %s\n", blue("→"), reg.BasePath)
	}

	// Setup registry
	if err := reg.Setup(setupForce); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Success message
	fmt.Printf("\n%s Registry initialized successfully!\n", green("✓"))

	paths := reg.Paths()
	fmt.Println("\nCreated:")
	fmt.Printf("  • %s\n", paths.Rules)
	fmt.Printf("  • %s\n", paths.Workflows)
	fmt.Printf("  • %s\n", paths.Guidelines)
	fmt.Printf("  • %s\n", paths.Profiles)

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'agmd init' in a project directory to create directives.md")
	fmt.Println("  2. Run 'agmd new rule <name>' to create custom rules")
	fmt.Println("  3. Run 'agmd add rule <name>' to add rules to a project")
	fmt.Printf("\n%s Default profile created:\n", blue("ℹ"))
	fmt.Println("  • 'agmd init' will use the 'default' profile automatically")
	fmt.Println("  • Customize: agmd edit profile:default")

	return nil
}
