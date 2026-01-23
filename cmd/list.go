package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available registry items including custom types",
	Long: `List all items in the registry including rules, workflows, guidelines, profiles, and custom types.
Shows which items are active in the current project (if directives.md exists).

Examples:
  agmd list           # List all available content`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Load active items from directives.md if it exists
	activeRules := make(map[string]bool)
	activeWorkflows := make(map[string]bool)
	activeGuidelines := make(map[string]bool)
	activeCustomTypes := make(map[string]map[string]bool) // type -> name -> bool
	hasProject := false

	if _, err := os.Stat(directivesMdFilename); err == nil {
		hasProject = true
		content, err := os.ReadFile(directivesMdFilename)
		if err == nil {
			extractActiveItems(string(content), activeRules, activeWorkflows, activeGuidelines, activeCustomTypes)
		}
	}

	fmt.Printf("%s Registry location: %s\n\n", blue("ℹ"), reg.BasePath)

	// List rules
	rules, err := reg.ListRules()
	if err != nil {
		return fmt.Errorf("failed to list rules: %w", err)
	}

	fmt.Printf("%s Rules (%d):\n", cyan("●"), len(rules))
	if len(rules) == 0 {
		fmt.Printf("  %s No rules in registry\n", yellow("ℹ"))
	} else {
		for _, rule := range rules {
			isActive := hasProject && activeRules[rule.Name]
			if isActive {
				fmt.Printf("  %s %s (active)\n", green("✓"), rule.Name)
			} else {
				fmt.Printf("    %s\n", rule.Name)
			}
			if rule.Description != "" {
				fmt.Printf("      %s\n", rule.Description)
			}
		}
	}
	fmt.Println()

	// List workflows
	workflows, err := reg.ListWorkflows()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	fmt.Printf("%s Workflows (%d):\n", cyan("●"), len(workflows))
	if len(workflows) == 0 {
		fmt.Printf("  %s No workflows in registry\n", yellow("ℹ"))
	} else {
		for _, workflow := range workflows {
			isActive := hasProject && activeWorkflows[workflow.Name]
			if isActive {
				fmt.Printf("  %s %s (active)\n", green("✓"), workflow.Name)
			} else {
				fmt.Printf("    %s\n", workflow.Name)
			}
			if workflow.Description != "" {
				fmt.Printf("      %s\n", workflow.Description)
			}
		}
	}
	fmt.Println()

	// List guidelines
	guidelines, err := reg.ListGuidelines()
	if err != nil {
		return fmt.Errorf("failed to list guidelines: %w", err)
	}

	fmt.Printf("%s Guidelines (%d):\n", cyan("●"), len(guidelines))
	if len(guidelines) == 0 {
		fmt.Printf("  %s No guidelines in registry\n", yellow("ℹ"))
	} else {
		for _, guideline := range guidelines {
			isActive := hasProject && activeGuidelines[guideline.Name]
			if isActive {
				fmt.Printf("  %s %s (active)\n", green("✓"), guideline.Name)
			} else {
				fmt.Printf("    %s\n", guideline.Name)
			}
			if guideline.Description != "" {
				fmt.Printf("      %s\n", guideline.Description)
			}
		}
	}
	fmt.Println()

	// List profiles
	profiles, err := reg.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	fmt.Printf("%s Profiles (%d):\n", cyan("●"), len(profiles))
	if len(profiles) == 0 {
		fmt.Printf("  %s No profiles found\n", yellow("ℹ"))
		fmt.Printf("  %s Create one with: agmd new profile:NAME\n", blue("ℹ"))
	} else {
		for _, profile := range profiles {
			fmt.Printf("    %s", profile.Name)
			if profile.Description != "" {
				fmt.Printf(" - %s", profile.Description)
			}
			fmt.Println()
		}
		fmt.Printf("\n  %s Use with: agmd init profile:NAME\n", blue("ℹ"))
	}
	fmt.Println()

	// List custom types (any directories in registry that aren't standard types)
	customTypes, err := listCustomTypes(reg.BasePath)
	if err == nil && len(customTypes) > 0 {
		fmt.Printf("%s Custom Types:\n", cyan("●"))
		for typeName, items := range customTypes {
			activeCount := 0
			if hasProject && activeCustomTypes[typeName] != nil {
				activeCount = len(activeCustomTypes[typeName])
			}

			if activeCount > 0 {
				fmt.Printf("  %s (%d items, %d active)\n", typeName, len(items), activeCount)
			} else {
				fmt.Printf("  %s (%d items)\n", typeName, len(items))
			}

			for _, item := range items {
				isActive := hasProject && activeCustomTypes[typeName] != nil && activeCustomTypes[typeName][item]
				if isActive {
					fmt.Printf("    %s %s (active)\n", green("✓"), item)
				} else {
					fmt.Printf("    %s\n", item)
				}
			}
		}
		fmt.Println()
	}

	if hasProject {
		summary := fmt.Sprintf("%d rules, %d workflows, %d guidelines",
			len(activeRules), len(activeWorkflows), len(activeGuidelines))

		// Add custom types to summary
		for typeName, items := range activeCustomTypes {
			if len(items) > 0 {
				summary += fmt.Sprintf(", %d %s", len(items), typeName)
			}
		}

		fmt.Printf("%s Active in current project: %s\n", blue("ℹ"), summary)
	} else {
		fmt.Printf("%s No project found in current directory (run 'agmd init' to create one)\n", yellow("ℹ"))
	}

	return nil
}

// extractActiveItems parses directives.md content and extracts active items
func extractActiveItems(content string, rules, workflows, guidelines map[string]bool, customTypes map[string]map[string]bool) {
	// Match :::include TYPE:NAME (for any type)
	includeRe := regexp.MustCompile(`(?m)^:::include\s+([a-z0-9-]+):([a-z0-9/_-]+)`)
	matches := includeRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			itemType := match[1]
			name := match[2]
			switch itemType {
			case "rule":
				rules[name] = true
			case "workflow":
				workflows[name] = true
			case "guideline":
				guidelines[name] = true
			default:
				// Custom type
				if customTypes[itemType] == nil {
					customTypes[itemType] = make(map[string]bool)
				}
				customTypes[itemType][name] = true
			}
		}
	}

	// Match :::list TYPE blocks (accept any type)
	listRe := regexp.MustCompile(`(?s):::list\s+([a-z0-9-]+)\s*\n(.*?)\n:::end`)
	listMatches := listRe.FindAllStringSubmatch(content, -1)
	for _, match := range listMatches {
		if len(match) >= 3 {
			listType := match[1]
			items := match[2]
			lines := strings.Split(items, "\n")
			for _, line := range lines {
				name := strings.TrimSpace(line)
				if name == "" {
					continue
				}
				switch listType {
				case "rule":
					rules[name] = true
				case "workflow":
					workflows[name] = true
				case "guideline":
					guidelines[name] = true
				default:
					// Handle custom types - we don't track them in separate maps for now
					// They'll be picked up by listCustomTypes() from the filesystem
				}
			}
		}
	}
}

// listCustomTypes scans registry for custom type directories (excluding profiles and shared)
func listCustomTypes(basePath string) (map[string][]string, error) {
	result := make(map[string][]string)
	excludedDirs := map[string]bool{
		"profile": true, // profiles are special - not content types
		"shared":  true, // shared is for common resources
	}

	// Read registry directory
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	// Find all type directories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		typeName := entry.Name()
		if excludedDirs[typeName] {
			continue
		}

		// Read items in custom type directory
		typePath := fmt.Sprintf("%s/%s", basePath, typeName)
		items, err := os.ReadDir(typePath)
		if err != nil {
			continue
		}

		var itemNames []string
		for _, item := range items {
			if !item.IsDir() && strings.HasSuffix(item.Name(), ".md") {
				// Remove .md extension
				name := strings.TrimSuffix(item.Name(), ".md")
				itemNames = append(itemNames, name)
			}
		}

		if len(itemNames) > 0 {
			result[typeName] = itemNames
		}
	}

	return result, nil
}
