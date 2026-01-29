package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var migrateForce bool

var migrateCmd = &cobra.Command{
	Use:   "migrate <file>",
	Short: "Migrate a raw CLAUDE.md or AGENTS.md to agmd format",
	Long: `Migrate an unstructured AI agent configuration file to the agmd format.

Use this for projects that don't use agmd yet - raw CLAUDE.md or AGENTS.md files
with freeform content that needs to be organized into rules, workflows, and guidelines.

The migration process:
1. Creates backup of original file (with .backup suffix)
2. Initializes project if needed (creates directives.md)
3. Appends content to directives.md with :::new markers
4. Opens directives.md in editor for you to organize content
5. Use 'agmd promote' to save organized items to your registry

For collecting rules from a project that already uses agmd (has directives.md),
use 'agmd collect' instead.

Examples:
  agmd migrate CLAUDE.md              # Migrate from CLAUDE.md
  agmd migrate .claude/claude.md      # Migrate from subdirectory
  agmd migrate existing.md --force    # Re-migrate with force`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVarP(&migrateForce, "force", "f", false, "Allow migration even if directives.md exists")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	sourceFile := args[0]

	// Check if source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s not found", sourceFile)
	}

	fmt.Printf("%s Migrating %s...\n", blue("→"), sourceFile)

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

	// Append migrated content to directives.md with :::new markers
	fmt.Printf("%s Appending migrated content to directives.md...\n", blue("→"))

	migratedContent := string(content)
	migrateSection := fmt.Sprintf(`

---

## Migrated Content (Organize Below)

Below is your migrated configuration. Use :::new markers to create new rules/workflows:

:::new rule:my-rule-name
# Rule: My Rule Name
Content here...
:::end

:::new workflow:my-workflow
# Workflow: My Workflow
Steps here...
:::end

---

%s

`, migratedContent)

	// Read existing directives.md
	existingDirectives, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Append migrate section
	newDirectives := string(existingDirectives) + migrateSection
	if err := os.WriteFile(directivesMdFilename, []byte(newDirectives), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}
	fmt.Printf("%s Migrated content appended to directives.md\n", green("✓"))

	fmt.Printf("\n%s Migration complete!\n", green("✓"))
	fmt.Println("\n" + blue("→") + " Opening directives.md for you to organize content...")
	fmt.Println()
	fmt.Println("Tips:")
	fmt.Println("  • Migrated content is at the bottom of directives.md")
	fmt.Println("  • Wrap sections with :::new markers to create new registry items:")
	fmt.Println("    :::new rule:typescript")
	fmt.Println("    # Rule: TypeScript Standards")
	fmt.Println("    Your content here...")
	fmt.Println("    :::end")
	fmt.Println("  • Run 'agmd promote' to save items to your registry")
	if !projectInitialized {
		fmt.Println("  • Run 'agmd sync' after organizing to update AGENTS.md")
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
