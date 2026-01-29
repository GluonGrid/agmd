package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"agmd/pkg/importer"
	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect rules from an agmd-compatible project into your registry",
	Long: `Collect rules, workflows, and guidelines from an agmd-compatible project into your personal registry.

This command is for projects that already use agmd (have directives.md with :::include directives).
It parses directives.md to find referenced items and extracts their content from AGENTS.md.

Use this when:
- You clone an agmd project and want its rules in your registry
- You find a project with great rules you want to reuse
- You want to build your registry from existing agmd projects

For migrating a project that doesn't use agmd yet (raw CLAUDE.md/AGENTS.md),
use 'agmd migrate' instead.

Examples:
  agmd collect                    # Collect from AGENTS.md (default)
  agmd collect --file CLAUDE.md   # Collect from CLAUDE.md instead
  agmd collect --all              # Collect all without prompts (skip conflicts)
  agmd collect --overwrite        # Overwrite existing items without asking
`,
	RunE: runCollect,
}

var (
	collectAll       bool
	collectOverwrite bool
	collectFile      string
)

func init() {
	rootCmd.AddCommand(collectCmd)
	collectCmd.Flags().BoolVar(&collectAll, "all", false, "Collect all items without prompting")
	collectCmd.Flags().BoolVar(&collectOverwrite, "overwrite", false, "Overwrite existing items without asking")
	collectCmd.Flags().StringVarP(&collectFile, "file", "f", "AGENTS.md", "Source file to collect from (default: AGENTS.md)")
}

func runCollect(cmd *cobra.Command, args []string) error {
	// Check if directives.md and agents file exist
	directivesPath := "directives.md"
	agentsPath := collectFile

	if _, err := os.Stat(directivesPath); os.IsNotExist(err) {
		return fmt.Errorf("directives.md not found in current directory")
	}

	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return fmt.Errorf("%s not found in current directory", agentsPath)
	}

	// Read files
	directivesContent, err := os.ReadFile(directivesPath)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", agentsPath, err)
	}

	// Parse and match items
	fmt.Printf("→ Analyzing directives.md and %s...\n", agentsPath)
	items, warnings, err := importer.MatchDirectivesWithAgents(
		string(directivesContent),
		string(agentsContent),
	)
	if err != nil {
		return fmt.Errorf("failed to match items: %w", err)
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("⚠ %s\n", warning)
		}
	}

	// Count total items
	totalItems := 0
	for _, itemList := range items {
		totalItems += len(itemList)
	}

	if totalItems == 0 {
		fmt.Println("ℹ No items found to collect")
		return nil
	}

	// Display what was found
	fmt.Printf("\n✓ Found %d items:\n", totalItems)
	for itemType, itemList := range items {
		fmt.Printf("  • %d %ss\n", len(itemList), itemType)
		for _, item := range itemList {
			fmt.Printf("    - %s\n", item.Name)
		}
	}

	// Initialize registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s. Run 'agmd setup' first", reg.GetBasePath())
	}

	// Collect items
	fmt.Println("\n→ Collecting to local registry...")
	collected := 0
	skipped := 0
	overwritten := 0

	for itemType, itemList := range items {
		for _, item := range itemList {
			// Check if item already exists in registry
			itemPath := filepath.Join(reg.Paths().Base, itemType, item.Name+".md")
			exists := false
			if _, err := os.Stat(itemPath); err == nil {
				exists = true
			}

			if exists {
				if collectOverwrite {
					// Overwrite without asking
					if err := writeItemToRegistry(reg, item); err != nil {
						fmt.Printf("✗ Failed to overwrite %s:%s: %v\n", itemType, item.Name, err)
					} else {
						overwritten++
						fmt.Printf("✓ Overwritten %s:%s\n", itemType, item.Name)
					}
				} else if collectAll {
					// Skip without asking
					skipped++
					fmt.Printf("⊘ Skipped %s:%s (already exists)\n", itemType, item.Name)
				} else {
					// Ask user what to do
					choice := promptConflict(itemType, item.Name)
					switch choice {
					case "overwrite":
						if err := writeItemToRegistry(reg, item); err != nil {
							fmt.Printf("✗ Failed to overwrite %s:%s: %v\n", itemType, item.Name, err)
						} else {
							overwritten++
							fmt.Printf("✓ Overwritten %s:%s\n", itemType, item.Name)
						}
					case "skip":
						skipped++
						fmt.Printf("⊘ Skipped %s:%s\n", itemType, item.Name)
					case "rename":
						newName := promptNewName(itemType, item.Name)
						item.Name = newName
						if err := writeItemToRegistry(reg, item); err != nil {
							fmt.Printf("✗ Failed to collect %s:%s: %v\n", itemType, item.Name, err)
						} else {
							collected++
							fmt.Printf("✓ Collected %s:%s (renamed)\n", itemType, item.Name)
						}
					}
				}
			} else {
				// Item doesn't exist, collect it
				if err := writeItemToRegistry(reg, item); err != nil {
					fmt.Printf("✗ Failed to collect %s:%s: %v\n", itemType, item.Name, err)
				} else {
					collected++
					fmt.Printf("✓ Collected %s:%s\n", itemType, item.Name)
				}
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("✓ Collection complete!\n")
	fmt.Printf("  • Collected: %d\n", collected)
	if overwritten > 0 {
		fmt.Printf("  • Overwritten: %d\n", overwritten)
	}
	if skipped > 0 {
		fmt.Printf("  • Skipped: %d\n", skipped)
	}

	return nil
}

// writeItemToRegistry writes an extracted item to the registry
func writeItemToRegistry(reg *registry.Registry, item importer.ImportedItem) error {
	// Determine target directory
	var targetDir string
	switch item.Type {
	case "rule":
		targetDir = reg.Paths().Rules
	case "workflow":
		targetDir = reg.Paths().Workflows
	case "guideline":
		targetDir = reg.Paths().Guidelines
	default:
		// Custom type - create directory if needed
		targetDir = filepath.Join(reg.GetBasePath(), item.Type)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for type '%s': %w", item.Type, err)
		}
	}

	// Create file path
	filePath := filepath.Join(targetDir, item.Name+".md")

	// Generate frontmatter
	frontmatter := fmt.Sprintf(`---
name: %s
category: extracted
description: Extracted from project
---

`, item.Name)

	// Write file
	content := frontmatter + item.Content
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}

// promptConflict asks user what to do with conflicting item
func promptConflict(itemType, itemName string) string {
	fmt.Printf("\n⚠ %s:%s already exists in registry\n", itemType, itemName)
	fmt.Println("  [O]verwrite  [S]kip  [R]ename")
	fmt.Print("  Choice: ")

	var input string
	fmt.Scanln(&input)

	switch input {
	case "O", "o", "overwrite":
		return "overwrite"
	case "S", "s", "skip", "":
		return "skip"
	case "R", "r", "rename":
		return "rename"
	default:
		return "skip"
	}
}

// promptNewName asks user for a new name
func promptNewName(itemType, oldName string) string {
	fmt.Printf("New name for %s:%s (default: %s-extracted): ", itemType, oldName, oldName)

	var input string
	fmt.Scanln(&input)

	if input == "" {
		return oldName + "-extracted"
	}
	return input
}
