package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	showMerged bool
	showJSON   bool
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show agent configuration",
	Long: `Display the agent configuration file.

Use --merged to show the effective configuration after resolving
all inheritance (universal → profiles → project → overrides).

Examples:
  agmd show                # Show raw AGENTS.md
  agmd show --merged       # Show fully resolved config
  agmd show --merged --json # Show as JSON`,
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVar(&showMerged, "merged", false, "Show merged config (after inheritance)")
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")
}

func runShow(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Find AGENTS.md in current directory
	agentFile := "AGENTS.md"
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md not found in current directory. Run 'agmd init' first")
	}

	if !showMerged {
		// Show raw file
		content, err := os.ReadFile(agentFile)
		if err != nil {
			return fmt.Errorf("failed to read AGENTS.md: %w", err)
		}
		fmt.Print(string(content))
		return nil
	}

	// Show merged config
	fmt.Printf("%s Resolving inheritance...\n", blue("→"))

	resolved, err := config.Resolve(agentFile)
	if err != nil {
		return fmt.Errorf("failed to resolve config: %w", err)
	}

	// Show inheritance chain
	fmt.Printf("\n%s Inheritance Chain:\n", blue("→"))
	if resolved.Universal != nil {
		fmt.Printf("  %s Universal: %s\n", green("✓"), resolved.Universal.Path)
	}
	for _, profile := range resolved.Profiles {
		fmt.Printf("  %s Profile: %s\n", green("✓"), profile.Path)
	}
	fmt.Printf("  %s Project: %s\n", green("✓"), resolved.Project.Path)

	if len(resolved.Project.Frontmatter.Overrides) > 0 {
		fmt.Printf("  %s Overrides: %d applied\n", green("✓"), len(resolved.Project.Frontmatter.Overrides))
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Show merged content
	if showJSON {
		// TODO: Implement JSON output
		fmt.Printf("%s JSON output not yet implemented\n", yellow("⚠"))
		return nil
	}

	// Show as markdown
	fmt.Printf("# Effective Configuration\n\n")
	fmt.Print(resolved.Merged.Content)

	return nil
}
