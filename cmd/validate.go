package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate directives.md for :::new blocks ready to promote",
	Long: `Validate directives.md by detecting :::new blocks that define
new rules, workflows, or guidelines that can be promoted to the registry.

This is useful after:
- Importing existing configuration (agmd import)
- Manually adding :::new blocks to directives.md
- Organizing content from imported files

The validator will detect patterns like:
- :::new rule:name ... ::: (new rule definitions)
- :::new workflow:name ... ::: (new workflow definitions)
- :::new guideline:name ... ::: (new guideline definitions)

And suggest promoting them to the registry using 'agmd promote'.

Examples:
  agmd validate           # Check for :::new blocks in directives.md`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s Validating directives.md...\n", blue("→"))

	// Check if directives.md exists
	if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("directives.md not found. Run 'agmd init' first")
	}

	// Read directives.md
	content, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Detect :::new blocks
	newBlocks := detectNewBlocks(string(content))

	if len(newBlocks.Rules) == 0 && len(newBlocks.Workflows) == 0 && len(newBlocks.Guidelines) == 0 {
		fmt.Printf("\n%s No :::new blocks found in directives.md\n", green("✓"))
		fmt.Println("\nAll content is either referenced from registry or inline.")
		return nil
	}

	// Report findings
	fmt.Printf("\n%s Found :::new blocks ready to promote:\n\n", yellow("ℹ"))

	if len(newBlocks.Rules) > 0 {
		fmt.Printf("%s Rules (%d):\n", cyan("●"), len(newBlocks.Rules))
		for _, name := range newBlocks.Rules {
			fmt.Printf("  • %s\n", name)
		}
		fmt.Println()
	}

	if len(newBlocks.Workflows) > 0 {
		fmt.Printf("%s Workflows (%d):\n", cyan("●"), len(newBlocks.Workflows))
		for _, name := range newBlocks.Workflows {
			fmt.Printf("  • %s\n", name)
		}
		fmt.Println()
	}

	if len(newBlocks.Guidelines) > 0 {
		fmt.Printf("%s Guidelines (%d):\n", cyan("●"), len(newBlocks.Guidelines))
		for _, name := range newBlocks.Guidelines {
			fmt.Printf("  • %s\n", name)
		}
		fmt.Println()
	}

	fmt.Println("Next step:")
	fmt.Printf("  • Promote these to the registry: %s\n", blue("agmd promote"))
	fmt.Println()
	fmt.Printf("%s Tip: 'agmd promote' will save them to ~/.agmd/ and replace :::new blocks with @references\n", blue("ℹ"))

	return nil
}

type NewBlocksContent struct {
	Rules      []string
	Workflows  []string
	Guidelines []string
}

// detectNewBlocks scans directives.md for :::new markers
func detectNewBlocks(content string) NewBlocksContent {
	result := NewBlocksContent{
		Rules:      []string{},
		Workflows:  []string{},
		Guidelines: []string{},
	}

	// Regex to match :::new:TYPE name=value blocks (parser syntax)
	// Example: :::new:rule name=simple-test
	re := regexp.MustCompile(`(?m)^:::new:(rule|workflow|guideline)\s+name=([a-z0-9/_-]+)\s*$`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		itemType := match[1]
		name := match[2]

		switch itemType {
		case "rule":
			if !contains(result.Rules, name) {
				result.Rules = append(result.Rules, name)
			}
		case "workflow":
			if !contains(result.Workflows, name) {
				result.Workflows = append(result.Workflows, name)
			}
		case "guideline":
			if !contains(result.Guidelines, name) {
				result.Guidelines = append(result.Guidelines, name)
			}
		}
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
