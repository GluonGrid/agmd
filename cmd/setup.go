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
	Long: `Create the ~/.agmd directory for storing reusable content.

The registry starts empty. Create items with any type you want:
  agmd new rule:my-rule
  agmd new framework:my-framework
  agmd new prompt:coding-assistant

Examples:
  agmd setup              # Initialize registry
  agmd setup --force      # Reinitialize`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupForce, "force", false, "Force reinitialize")
}

func runSetup(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	if reg.Exists() && !setupForce {
		fmt.Printf("%s Registry already exists: %s\n", yellow("!"), reg.BasePath)
		fmt.Println("\nUse --force to reinitialize")
		return nil
	}

	if setupForce && reg.Exists() {
		fmt.Printf("%s Reinitializing: %s\n", yellow("!"), reg.BasePath)
	} else {
		fmt.Printf("%s Creating: %s\n", blue("->"), reg.BasePath)
	}

	if err := reg.Setup(setupForce); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	fmt.Printf("\n%s Registry ready!\n", green("ok"))
	fmt.Println("\nBundled:")
	fmt.Println("  guide:agmd       — quick reference for AI agents")
	fmt.Println("  skill:agmd-migrate — step-by-step migration guide for AI agents")
	fmt.Println("\nNext steps:")
	fmt.Println("  agmd migrate CLAUDE.md        # Migrate existing AI instructions → directives.md")
	fmt.Println("  agmd init                     # Start fresh in a new project")
	fmt.Println("  agmd skill link agmd-migrate  # Link the migration skill to your agent dirs")

	return nil
}
