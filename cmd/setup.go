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
	fmt.Println("\nNext steps:")
	fmt.Println("  agmd init              # Create directives.md in a project")
	fmt.Println("  agmd new type:name     # Create a reusable item")
	fmt.Println("  agmd list              # See your registry")

	return nil
}
