package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate agent configuration",
	Long: `Validate the AGENTS.md configuration file and check:
- YAML frontmatter is valid
- Universal shared config exists (if specified)
- All profiles exist and are valid
- Overrides reference existing sections
- File structure is correct

Examples:
  agmd validate           # Validate current directory's AGENTS.md`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Find AGENTS.md in current directory
	agentFile := "AGENTS.md"
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md not found in current directory")
	}

	fmt.Printf("Validating %s...\n\n", agentFile)

	// Run validation
	result := config.Validate(agentFile)

	// Show errors
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Printf("%s %s: %s\n", red("✗"), err.Type, err.Message)
			if err.Path != "" && err.Path != agentFile {
				fmt.Printf("  Path: %s\n", err.Path)
			}
		}
		fmt.Println()
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			fmt.Printf("%s %s\n", yellow("!"), warning)
		}
		fmt.Println()
	}

	// Show result
	if result.Valid {
		fmt.Printf("%s Configuration is valid\n", green("✓"))
		if len(result.Warnings) > 0 {
			fmt.Printf("\n%d warning(s) found\n", len(result.Warnings))
		}
		return nil
	}

	fmt.Printf("%s Configuration is invalid\n", red("✗"))
	fmt.Printf("\n%d error(s), %d warning(s)\n", len(result.Errors), len(result.Warnings))
	os.Exit(1)
	return nil
}
