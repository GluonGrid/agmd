package cmd

import (
	"fmt"
	"os"
	"strings"

	"agmd/pkg/markdown"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [type:name]",
	Short: "Remove a rule, workflow, or guideline from directives.md",
	Long: `Remove a rule, workflow, or guideline from directives.md.
This removes the directive:
- Removes from :::list:TYPE block if present
- Removes :::include:TYPE name if present

Run 'agmd sync' after removing to update AGENTS.md.

Note: This only removes the directive from directives.md.
The item remains in the registry and can be added again later.

Examples:
  agmd remove rule:typescript       # Remove TypeScript rule
  agmd remove workflow:release      # Remove release workflow
  agmd remove guideline:code-style  # Remove code style guideline`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	// Parse type:name
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	name := parts[1]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Validate type
	if itemType != "rule" && itemType != "workflow" && itemType != "guideline" {
		return fmt.Errorf("invalid type '%s'. Must be: rule, workflow, or guideline", itemType)
	}

	fmt.Printf("%s Removing %s: %s\n", blue("→"), itemType, name)

	// Check if directives.md exists
	if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("directives.md not found\nRun 'agmd init' first to initialize project")
	}

	// Read directives.md
	fmt.Printf("%s Reading directives.md...\n", blue("→"))
	content, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Remove directive from directives.md
	fmt.Printf("%s Removing directive...\n", blue("→"))
	updated, err := markdown.RemoveFromDirective(content, itemType, name)
	if err != nil {
		return err
	}

	// Write updated directives.md
	if err := os.WriteFile(directivesMdFilename, updated, 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}

	fmt.Printf("\n%s Removed %s '%s' from directives.md!\n", green("✓"), itemType, name)
	fmt.Printf("\nRun 'agmd sync' to update AGENTS.md.\n")

	return nil
}
