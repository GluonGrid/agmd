package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"agmd/pkg/importer"
	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract rules/workflows/guidelines from AGENTS.md into local registry",
	Long: `Extract parses AGENTS.md and directives.md to import items into your local registry.

This is useful when:
- You clone a project and want to use its rules locally
- You accidentally deleted your registry and need to recover
- You want to add someone else's rules to your personal collection

The command reads both directives.md (for structure) and AGENTS.md (for content)
and imports the items into your ~/.agmd/ registry.

Examples:
  agmd extract              # Extract all items from current project
  agmd extract --all        # Extract all without prompts (skip conflicts)
  agmd extract --overwrite  # Overwrite existing items without asking
`,
	RunE: runExtract,
}

var (
	extractAll       bool
	extractOverwrite bool
)

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().BoolVar(&extractAll, "all", false, "Extract all items without prompting")
	extractCmd.Flags().BoolVar(&extractOverwrite, "overwrite", false, "Overwrite existing items without asking")
}

func runExtract(cmd *cobra.Command, args []string) error {
	// Check if directives.md and AGENTS.md exist
	directivesPath := "directives.md"
	agentsPath := "AGENTS.md"

	if _, err := os.Stat(directivesPath); os.IsNotExist(err) {
		return fmt.Errorf("directives.md not found in current directory")
	}

	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md not found in current directory")
	}

	// Read files
	directivesContent, err := os.ReadFile(directivesPath)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		return fmt.Errorf("failed to read AGENTS.md: %w", err)
	}

	// Parse and match items
	fmt.Println("→ Analyzing directives.md and AGENTS.md...")
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
		fmt.Println("ℹ No items found to extract")
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

	// Extract items
	fmt.Println("\n→ Extracting to local registry...")
	imported := 0
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
				if extractOverwrite {
					// Overwrite without asking
					if err := writeItemToRegistry(reg, item); err != nil {
						fmt.Printf("✗ Failed to overwrite %s:%s: %v\n", itemType, item.Name, err)
					} else {
						overwritten++
						fmt.Printf("✓ Overwritten %s:%s\n", itemType, item.Name)
					}
				} else if extractAll {
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
							fmt.Printf("✗ Failed to extract %s:%s: %v\n", itemType, item.Name, err)
						} else {
							imported++
							fmt.Printf("✓ Extracted %s:%s (renamed)\n", itemType, item.Name)
						}
					}
				}
			} else {
				// Item doesn't exist, extract it
				if err := writeItemToRegistry(reg, item); err != nil {
					fmt.Printf("✗ Failed to extract %s:%s: %v\n", itemType, item.Name, err)
				} else {
					imported++
					fmt.Printf("✓ Extracted %s:%s\n", itemType, item.Name)
				}
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("✓ Extraction complete!\n")
	fmt.Printf("  • Extracted: %d\n", imported)
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
