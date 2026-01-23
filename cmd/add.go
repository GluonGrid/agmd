package cmd

import (
	"fmt"
	"os"
	"strings"

	"agmd/pkg/markdown"
	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	agentsTomlFilename = "agents.toml" // Legacy - for backwards compatibility
)

var addCmd = &cobra.Command{
	Use:   "add [type:name]",
	Short: "Add a rule, workflow, or guideline to directives.md",
	Long: `Add a rule, workflow, or guideline from the registry to directives.md.
This inserts a directive:
- Appends to first existing :::list TYPE block if found
- Otherwise creates ## Section with :::include:TYPE name

Run 'agmd sync' after adding to update AGENTS.md.

Examples:
  agmd add rule:typescript       # Add TypeScript rule
  agmd add workflow:release      # Add release workflow
  agmd add guideline:code-style  # Add code style guideline`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	// Add directive to directives.md
	fmt.Printf("%s Inserting directive...\n", blue("→"))
	updated, err := markdown.AddToDirective(content, itemType, name)
	if err != nil {
		return fmt.Errorf("failed to add directive: %w", err)
	}

	// Write updated directives.md
	if err := os.WriteFile(directivesMdFilename, updated, 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}

	fmt.Printf("\n%s Added %s '%s' to directives.md!\n", green("✓"), itemType, name)
	fmt.Printf("\nRun 'agmd sync' to update AGENTS.md.\n")

	return nil
}
