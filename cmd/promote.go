package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var promoteAll bool

var promoteCmd = &cobra.Command{
	Use:   "promote [type:name]",
	Short: "Promote :::new blocks from directives.md to registry",
	Long: `Promote :::new blocks from directives.md to the registry.
This extracts content from :::new markers, saves it to ~/.agmd/,
and replaces the :::new block with :::include directive.

Interactive mode (no arguments):
  agmd promote

Promote all at once:
  agmd promote --all

Specific item:
  agmd promote rule:custom-auth
  agmd promote prompt:code-review

Examples:
  agmd promote               # Interactive: select which items to promote
  agmd promote --all         # Promote all :::new blocks automatically
  agmd promote rule:auth     # Promote specific item`,
	RunE: runPromote,
}

func init() {
	rootCmd.AddCommand(promoteCmd)
	promoteCmd.Flags().BoolVarP(&promoteAll, "all", "a", false, "Promote all :::new blocks automatically")
}

func runPromote(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	// Check if directives.md exists
	if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("directives.md not found. Run 'agmd init' first")
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Read directives.md
	content, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	directivesContent := string(content)

	// Detect :::new blocks
	newBlocks := detectNewBlocks(directivesContent)

	if len(newBlocks.Items) == 0 {
		fmt.Printf("%s No :::new blocks found\n", green("✓"))
		return nil
	}

	// If specific item provided (type:name format)
	if len(args) == 1 {
		parts := strings.SplitN(args[0], ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
		}
		itemType := strings.ToLower(parts[0])
		name := parts[1]
		return promoteSingle(itemType, name, directivesContent, reg)
	}

	// If --all flag, promote everything
	if promoteAll {
		return promoteAllBlocks(newBlocks, directivesContent, reg)
	}

	// Interactive mode - promote multiple
	return promoteInteractive(newBlocks, directivesContent, reg)
}

func promoteAllBlocks(newBlocks NewBlocksContent, directivesContent string, reg *registry.Registry) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	if len(newBlocks.Items) == 0 {
		fmt.Printf("%s No :::new blocks to promote\n", green("✓"))
		return nil
	}

	fmt.Printf("%s Found %d :::new blocks to promote\n", blue("→"), len(newBlocks.Items))
	fmt.Println()

	// Promote all
	promoted := 0
	updatedContent := directivesContent

	for _, item := range newBlocks.Items {
		fmt.Printf("%s Promoting %s:%s\n", blue("→"), item.Type, item.Name)
		newContent, err := promoteSingleToRegistry(item.Type, item.Name, updatedContent, reg)
		if err != nil {
			fmt.Printf("%s Failed: %v\n", yellow("⚠"), err)
		} else {
			updatedContent = newContent
			promoted++
		}
	}

	// Save updated directives.md
	if promoted > 0 {
		if err := os.WriteFile(directivesMdFilename, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("failed to write directives.md: %w", err)
		}
		fmt.Printf("\n%s Complete! %d/%d items promoted to registry and directives.md updated.\n", green("✓"), promoted, len(newBlocks.Items))
		fmt.Printf("%s Run 'agmd sync' to update AGENTS.md\n", blue("ℹ"))
	}

	return nil
}

func promoteInteractive(newBlocks NewBlocksContent, directivesContent string, reg *registry.Registry) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	reader := bufio.NewReader(os.Stdin)

	// Show what's available
	fmt.Printf("%s Found :::new blocks:\n\n", blue("→"))

	for i, item := range newBlocks.Items {
		fmt.Printf("  %d. %s:%s\n", i+1, item.Type, item.Name)
	}
	fmt.Println()

	// Ask which to promote
	fmt.Print("Promote which items? (comma-separated numbers, 'all', or 'none'): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "none" || response == "" {
		fmt.Println("Cancelled.")
		return nil
	}

	var selectedItems []NewBlock

	if response == "all" {
		selectedItems = newBlocks.Items
	} else {
		// Parse numbers
		parts := strings.Split(response, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			num, err := strconv.Atoi(part)
			if err != nil || num < 1 || num > len(newBlocks.Items) {
				fmt.Printf("%s Invalid selection: %s\n", yellow("⚠"), part)
				continue
			}
			selectedItems = append(selectedItems, newBlocks.Items[num-1])
		}
	}

	if len(selectedItems) == 0 {
		fmt.Println("No items selected.")
		return nil
	}

	// Promote each selected item
	promoted := 0
	updatedContent := directivesContent

	for _, item := range selectedItems {
		fmt.Printf("\n%s Promoting %s:%s\n", blue("→"), item.Type, item.Name)
		newContent, err := promoteSingleToRegistry(item.Type, item.Name, updatedContent, reg)
		if err != nil {
			fmt.Printf("%s Failed: %v\n", yellow("⚠"), err)
		} else {
			updatedContent = newContent
			promoted++
		}
	}

	// Save updated directives.md
	if promoted > 0 {
		if err := os.WriteFile(directivesMdFilename, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("failed to write directives.md: %w", err)
		}
		fmt.Printf("\n%s Complete! %d items promoted to registry and directives.md updated.\n", green("✓"), promoted)
		fmt.Printf("%s Run 'agmd sync' to update AGENTS.md\n", blue("ℹ"))
	}

	return nil
}

func promoteSingle(itemType, name string, directivesContent string, reg *registry.Registry) error {
	// Promote and save updated directives.md
	updatedContent, err := promoteSingleToRegistry(itemType, name, directivesContent, reg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(directivesMdFilename, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("\n%s Promoted successfully!\n", green("✓"))
	fmt.Printf("%s Run 'agmd sync' to update AGENTS.md\n", blue("ℹ"))

	return nil
}

func promoteSingleToRegistry(itemType, name string, directivesContent string, reg *registry.Registry) (string, error) {
	green := color.New(color.FgGreen).SprintFunc()

	// Extract :::new TYPE:NAME block content (parser syntax)
	// Example: :::new rule:simple-test
	// Match the full block including :::end
	blockPattern := fmt.Sprintf(`(?s):::new\s+%s:%s\s*\n(.*?)\n:::end`, regexp.QuoteMeta(itemType), regexp.QuoteMeta(name))
	re := regexp.MustCompile(blockPattern)
	match := re.FindStringSubmatch(directivesContent)

	if match == nil {
		return "", fmt.Errorf("could not find :::new %s:%s block", itemType, name)
	}

	blockContent := strings.TrimSpace(match[1])
	fullMatch := match[0]

	fmt.Printf("%s Extracted content from :::new block\n", green("✓"))

	// Check if already exists in registry
	basePath := reg.TypePath(itemType)

	filePath := fmt.Sprintf("%s/%s.md", basePath, name)

	// Check if exists
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("%s:%s already exists in registry at %s", itemType, name, filePath)
	}

	// Create type directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create type directory: %w", err)
	}

	// Create subdirectories if needed (e.g., auth/custom-auth)
	if strings.Contains(name, "/") {
		// Extract directory path without filename
		lastSlash := strings.LastIndex(name, "/")
		fileDir := fmt.Sprintf("%s/%s", basePath, name[:lastSlash])
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create subdirectories: %w", err)
		}
		fmt.Printf("%s Created subdirectory at %s\n", green("✓"), fileDir)
	}

	// Create with frontmatter (empty description field for user to fill)
	fullContent := fmt.Sprintf(`---
name: %s
description: ""
---

%s`, name, blockContent)

	if err := os.WriteFile(filePath, []byte(fullContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write to registry: %w", err)
	}

	fmt.Printf("%s Created %s at %s\n", green("✓"), itemType, filePath)

	// Replace :::new block with :::include directive in directives.md
	replacement := fmt.Sprintf(":::include %s:%s", itemType, name)
	updatedContent := strings.Replace(directivesContent, fullMatch, replacement, 1)

	fmt.Printf("%s Replaced :::new block with :::include %s:%s\n", green("✓"), itemType, name)

	return updatedContent, nil
}

