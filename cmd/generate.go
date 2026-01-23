package cmd

import (
	"fmt"
	"os"

	"agmd/pkg/generator"
	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	directivesMdFilename = "directives.md" // Source file with directives
	agentsMdFilename     = "AGENTS.md"     // Generated output
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate AGENTS.md from directives.md",
	Long: `Read directives.md and generate expanded AGENTS.md for AI agents.

The command:
1. Reads directives.md (source file with directives)
2. Expands all directives with content from registry
3. Writes expanded output to AGENTS.md

Processes:
- :::include:TYPE name directives
- :::list:TYPE blocks
- :::new:TYPE name=value blocks

All non-directive content is preserved.

Examples:
  agmd generate           # Generate AGENTS.md from directives.md`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
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
