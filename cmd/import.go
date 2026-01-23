package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var importForce bool

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import existing config file and migrate to agmd format",
	Long: `Import an existing AI agent configuration file (CLAUDE.md, .claude/claude.md, etc.)
and append it to directives.md for you to organize into rules, workflows, and guidelines.

The import process:
1. Creates backup of original file (in place, with .backup suffix)
2. Initializes project if needed (creates agents.toml and directives.md)
3. Appends imported content to directives.md with :::new markers
4. Opens directives.md in editor for you to organize content
5. You organize content using :::new markers to create new registry items

Examples:
  agmd import CLAUDE.md              # Import from root
  agmd import .claude/claude.md      # Import from subdirectory
  agmd import existing.md --force    # Re-import with force`,
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

	// Initialize project if needed
	projectInitialized := false
	if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
		fmt.Printf("%s Project not initialized. Initializing...\n", blue("→"))
		if err := runInitCommand(); err != nil {
			return fmt.Errorf("failed to initialize project: %w", err)
		}
		projectInitialized = true
		fmt.Printf("%s Project initialized\n", green("✓"))
	}

	// Append imported content to directives.md with :::new markers
	fmt.Printf("%s Appending imported content to directives.md...\n", blue("→"))

	importedContent := string(content)
	importSection := fmt.Sprintf(`

---

## Imported Content (Organize Below)

Below is your imported configuration. Use :::new markers to create new rules/workflows:

:::new rule:my-rule-name
# Rule: My Rule Name
Content here...
:::

:::new workflow:my-workflow
# Workflow: My Workflow
Steps here...
:::

---

%s

`, importedContent)

	// Read existing directives.md
	existingDirectives, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Append import section
	newDirectives := string(existingDirectives) + importSection
	if err := os.WriteFile(directivesMdFilename, []byte(newDirectives), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}
	fmt.Printf("%s Imported content appended to directives.md\n", green("✓"))

	fmt.Printf("\n%s Import complete!\n", green("✓"))
	fmt.Println("\n" + blue("→") + " Opening directives.md for you to organize content...")
	fmt.Println()
	fmt.Println("Tips:")
	fmt.Println("  • Imported content is at the bottom of directives.md")
	fmt.Println("  • Wrap sections with :::new markers to create new registry items:")
	fmt.Println("    :::new rule:typescript")
	fmt.Println("    # Rule: TypeScript Standards")
	fmt.Println("    Your content here...")
	fmt.Println("    :::")
	fmt.Println("  • Use @rule:name or @workflow:name to reference existing items")
	if !projectInitialized {
		fmt.Println("  • Run 'agmd generate' after organizing to update AGENTS.md")
	}
	fmt.Println()

	// Open in editor
	if err := openInEditor(directivesMdFilename); err != nil {
		fmt.Printf("%s Could not open editor: %v\n", yellow("⚠"), err)
		fmt.Println("\nManually edit directives.md to organize your content.")
	}

	return nil
}

// runInitCommand is a helper to run project initialization
func runInitCommand() error {
	// Call the init command logic
	return runInit(nil, nil)
}
