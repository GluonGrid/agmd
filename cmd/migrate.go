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
	Short: "Migrate a CLAUDE.md or AGENTS.md to agmd format",
	Long: `Migrate an existing AI configuration file to the agmd format.

The migration process:
1. Creates a backup of your file (file.backup)
2. Copies content to directives.md with a guide header
3. Opens your editor to organize content with :::new markers

After editing:
  agmd promote    # Save :::new blocks to your registry
  agmd sync       # Generate AGENTS.md

Examples:
  agmd migrate CLAUDE.md
  agmd migrate .cursor/rules.md
  agmd migrate existing.md --force`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVarP(&migrateForce, "force", "f", false, "Overwrite existing directives.md")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	sourceFile := args[0]

	// Check source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", sourceFile)
	}

	// Check if directives.md already exists
	if _, err := os.Stat(directivesMdFilename); err == nil && !migrateForce {
		return fmt.Errorf("directives.md already exists\nUse --force to overwrite")
	}

	// Check registry exists
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	fmt.Printf("%s Migrating %s...\n", blue("->"), sourceFile)

	// Read source file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create backup
	backupPath := sourceFile + ".backup"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("%s Backup created: %s\n", green("ok"), backupPath)

	// Create directives.md with frontmatter + guide + content
	directives := fmt.Sprintf(`---
# agmd directives file
# Edit this file, then run: agmd promote && agmd sync
---

# Directives

<!--
Wrap sections with :::new to create registry items:

:::new rule:my-rule
Your rule content here...
:::

:::new workflow:my-workflow
Your workflow steps...
:::

After editing, run:
  agmd promote   (save to registry)
  agmd sync      (generate AGENTS.md)
-->

%s
`, string(content))

	if err := os.WriteFile(directivesMdFilename, []byte(directives), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}
	fmt.Printf("%s Created directives.md\n", green("ok"))

	// Open in editor
	fmt.Printf("\n%s Opening editor...\n", blue("->"))
	if err := openInEditor(directivesMdFilename); err != nil {
		fmt.Printf("%s Could not open editor: %v\n", yellow("!"), err)
		fmt.Println("Edit directives.md manually.")
	}

	fmt.Printf("\n%s Migration complete!\n", green("ok"))
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Wrap sections with :::new markers")
	fmt.Println("  2. Run 'agmd promote' to save to registry")
	fmt.Println("  3. Run 'agmd sync' to generate AGENTS.md")

	return nil
}
