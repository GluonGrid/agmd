package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/generator"
	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync and generate AGENTS.md from directives.md",
	Long: `Read directives.md and generate expanded AGENTS.md for AI agents.

The command:
1. Checks for unpromoted :::new blocks (errors if found)
2. Reads directives.md (source file with directives)
3. Expands all :::include and :::list directives with content from registry
4. Writes expanded output to AGENTS.md

All non-directive content is preserved.

Note: If you have :::new blocks, run 'agmd promote' first to add them to
your registry with proper metadata (name, description).

Examples:
  agmd sync               # Generate AGENTS.md from directives.md`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("%s Generating AGENTS.md from directives.md...\n", blue("→"))

	// Check if directives.md exists
	if _, err := os.Stat(directivesMdFilename); err != nil {
		return fmt.Errorf("directives.md not found\nRun 'agmd init' first")
	}

	// Load registry
	fmt.Printf("%s Loading registry...\n", blue("→"))
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Check for :::new blocks - they must be promoted first
	directivesBytes, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}
	newBlocks := detectNewBlocks(string(directivesBytes))
	if len(newBlocks.Items) > 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s Found %d :::new blocks that need to be promoted first\n", yellow("⚠"), len(newBlocks.Items))
		fmt.Println("\nRun 'agmd promote' to add them to your registry with proper metadata.")
		fmt.Println("This ensures each item has a name and description in the frontmatter.")
		return fmt.Errorf("cannot sync with unpromoted :::new blocks")
	}

	// Auto-sync filenames with frontmatter names
	autoSyncRegistryFilenames(reg)

	// Create generator
	gen := generator.New(reg, nil)

	// Parse and expand directives from directives.md
	fmt.Printf("%s Parsing and expanding directives...\n", blue("→"))
	content, err := gen.ParseAndExpand(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to parse and expand directives.md: %w", err)
	}

	// Write expanded output to AGENTS.md
	if err := os.WriteFile(agentsMdFilename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("\n%s Generated AGENTS.md successfully!\n", green("✓"))
	fmt.Printf("%s Source: %s → Output: %s\n", blue("ℹ"), directivesMdFilename, agentsMdFilename)

	return nil
}

