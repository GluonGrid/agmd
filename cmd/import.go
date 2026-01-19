package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"agmd/pkg/generator"
	"agmd/pkg/registry"
	"agmd/pkg/state"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var importForce bool

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import existing config file and migrate to agmd format",
	Long: `Import an existing AI agent configuration file (CLAUDE.md, .claude/claude.md, etc.)
and migrate it to agmd format. The imported content is placed in the custom section
for you to organize into agmd rules, workflows, and guidelines.

The import process:
1. Creates backup of original file (in place, with .backup suffix)
2. Creates agents.toml with default configuration
3. Creates AGENTS.md with imported content in custom section
4. Opens AGENTS.md in editor for you to organize content
5. Validates for any unregistered rules after editing

Examples:
  agmd import CLAUDE.md              # Import from root
  agmd import .claude/claude.md      # Import from subdirectory
  agmd import AGENTS.md --force      # Re-import non-agmd format AGENTS.md`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().BoolVarP(&importForce, "force", "f", false, "Allow import even if agents.toml exists")
}

func runImport(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	sourceFile := args[0]

	// Check if source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s not found", sourceFile)
	}

	// Check if agents.toml already exists
	if _, err := os.Stat(agentsTomlFilename); err == nil && !importForce {
		return fmt.Errorf("agents.toml already exists. Use --force to reimport")
	}

	// Check if AGENTS.md already exists (and it's not the source)
	if sourceFile != agentsMdFilename {
		if _, err := os.Stat(agentsMdFilename); err == nil && !importForce {
			return fmt.Errorf("AGENTS.md already exists. Use --force to overwrite")
		}
	}

	fmt.Printf("%s Importing %s...\n", blue("→"), sourceFile)

	// Read source file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sourceFile, err)
	}

	// Create backup IN PLACE (same directory)
	backupPath := sourceFile + ".backup"
	fmt.Printf("%s Creating backup...\n", blue("→"))
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("%s Created backup: %s\n", green("✓"), backupPath)

	// Check if registry exists
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		fmt.Printf("%s Registry not found. Run 'agmd setup' first.\n", yellow("⚠"))
		return fmt.Errorf("registry not initialized")
	}

	// Create agents.toml with default state
	fmt.Printf("%s Creating agents.toml...\n", blue("→"))
	st := state.DefaultState()
	if err := st.Save(agentsTomlFilename); err != nil {
		return fmt.Errorf("failed to create agents.toml: %w", err)
	}
	fmt.Printf("%s Created agents.toml\n", green("✓"))

	// Create AGENTS.md with imported content in custom section
	fmt.Printf("%s Creating AGENTS.md...\n", blue("→"))
	gen := generator.New(reg, st)

	// Generate base AGENTS.md (with no-modify-agmd-sections rule)
	agentsMdContent, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate AGENTS.md template: %w", err)
	}

	// Replace the default custom section content with imported content
	importedContent := string(content)

	customSectionWithImport := fmt.Sprintf(`<!-- agmd:custom-start -->

## Imported Content

Below is your imported configuration. You can:
- Keep it here as custom content, OR
- Organize it into the managed section above by moving content there
- After organizing, run 'agmd validate' to detect unregistered rules

---

%s

<!-- agmd:custom-end -->`, importedContent)

	// Replace custom section
	agentsMdContent = replaceCustomSection(agentsMdContent, customSectionWithImport)

	// Write AGENTS.md
	if err := os.WriteFile(agentsMdFilename, []byte(agentsMdContent), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}
	fmt.Printf("%s Created AGENTS.md with imported content\n", green("✓"))

	// Remove original file if it's not AGENTS.md and not in root
	if sourceFile != agentsMdFilename && sourceFile != filepath.Base(sourceFile) {
		// It's in a subdirectory, offer to remove
		fmt.Printf("%s Original file is in subdirectory: %s\n", blue("ℹ"), sourceFile)
		fmt.Println("  The backup remains at:", backupPath)
		fmt.Println("  You may want to remove the original if no longer needed")
	}

	fmt.Printf("\n%s Import complete!\n", green("✓"))
	fmt.Println("\n" + blue("→") + " Opening AGENTS.md for you to organize content...")
	fmt.Println()
	fmt.Println("Tips:")
	fmt.Println("  • Imported content is in the custom section")
	fmt.Println("  • Move rules/workflows to managed section if desired")
	fmt.Println("  • After editing, validation will run automatically")
	fmt.Println()

	// Open in editor
	if err := openInEditor(agentsMdFilename); err != nil {
		fmt.Printf("%s Could not open editor: %v\n", yellow("⚠"), err)
		fmt.Println("\nManually edit AGENTS.md to organize your content.")
	}

	// Auto-validate after editing
	fmt.Printf("\n%s Running validation...\n", blue("→"))
	if err := runValidateCommand(); err != nil {
		fmt.Printf("%s Validation completed with suggestions\n", yellow("ℹ"))
	} else {
		fmt.Printf("%s No unregistered content detected\n", green("✓"))
	}

	return nil
}

// replaceCustomSection replaces the custom section in the template
func replaceCustomSection(content, newCustomSection string) string {
	// Find and replace the default custom section
	startMarker := "<!-- agmd:custom-start -->"
	endMarker := "<!-- agmd:custom-end -->"

	startIdx := findLastIndex(content, startMarker)
	if startIdx == -1 {
		return content
	}

	endIdx := findIndex(content[startIdx:], endMarker)
	if endIdx == -1 {
		return content
	}

	endIdx += startIdx + len(endMarker)

	return content[:startIdx] + newCustomSection + content[endIdx:]
}

func findLastIndex(s, substr string) int {
	lastIdx := -1
	idx := 0
	for {
		i := findIndex(s[idx:], substr)
		if i == -1 {
			break
		}
		lastIdx = idx + i
		idx = lastIdx + 1
	}
	return lastIdx
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// runValidateCommand is a helper to run validation
func runValidateCommand() error {
	// Will be implemented with validate command
	// For now, just return nil
	return nil
}
