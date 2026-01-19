package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"agmd/pkg/registry"
	"agmd/pkg/state"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate AGENTS.md for unregistered content",
	Long: `Validate AGENTS.md by checking for content in the managed section
that is not registered in the registry or listed in agents.toml.

This is useful after:
- Importing existing configuration (agmd import)
- Manually adding rules/workflows to the managed section
- Organizing content from custom to managed sections

The validator will detect patterns like:
- ### rule-name (unregistered rules)
- ### workflow-name (unregistered workflows)
- ### guideline-name (unregistered guidelines)

And suggest either promoting them to the registry or moving to custom section.

Examples:
  agmd validate           # Check for unregistered content`,
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

	fmt.Printf("%s Validating AGENTS.md...\n", blue("→"))

	// Check if AGENTS.md exists
	if _, err := os.Stat(agentsMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md not found. Run 'agmd init' or 'agmd import' first")
	}

	// Check if agents.toml exists
	if _, err := os.Stat(agentsTomlFilename); os.IsNotExist(err) {
		return fmt.Errorf("agents.toml not found. Run 'agmd init' first")
	}

	// Load state
	st, err := state.Load(agentsTomlFilename)
	if err != nil {
		return fmt.Errorf("failed to load agents.toml: %w", err)
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Read AGENTS.md
	content, err := os.ReadFile(agentsMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read AGENTS.md: %w", err)
	}

	// Extract managed section
	managedContent := extractManagedSection(string(content))
	if managedContent == "" {
		return fmt.Errorf("could not find managed section in AGENTS.md")
	}

	// Detect unregistered content
	unregistered := detectUnregistered(managedContent, st, reg)

	if len(unregistered.Rules) == 0 && len(unregistered.Workflows) == 0 && len(unregistered.Guidelines) == 0 {
		fmt.Printf("\n%s No unregistered content detected!\n", green("✓"))
		fmt.Println("\nAll content in managed section is properly registered.")
		return nil
	}

	// Report findings
	fmt.Printf("\n%s Found unregistered content in managed section:\n\n", yellow("⚠"))

	if len(unregistered.Rules) > 0 {
		fmt.Printf("%s Rules (%d):\n", cyan("●"), len(unregistered.Rules))
		for _, name := range unregistered.Rules {
			fmt.Printf("  • %s (not in registry)\n", name)
		}
		fmt.Println()
	}

	if len(unregistered.Workflows) > 0 {
		fmt.Printf("%s Workflows (%d):\n", cyan("●"), len(unregistered.Workflows))
		for _, name := range unregistered.Workflows {
			fmt.Printf("  • %s (not in registry)\n", name)
		}
		fmt.Println()
	}

	if len(unregistered.Guidelines) > 0 {
		fmt.Printf("%s Guidelines (%d):\n", cyan("●"), len(unregistered.Guidelines))
		for _, name := range unregistered.Guidelines {
			fmt.Printf("  • %s (not in registry)\n", name)
		}
		fmt.Println()
	}

	fmt.Println("Options:")
	fmt.Println("  1. Promote to registry:")
	fmt.Printf("     %s\n", blue("agmd promote"))
	fmt.Println()
	fmt.Println("  2. Move to custom section:")
	fmt.Printf("     Manually move content between markers:\n")
	fmt.Printf("     %s and %s\n", yellow("<!-- agmd:custom-start -->"), yellow("<!-- agmd:custom-end -->"))
	fmt.Println()
	fmt.Printf("%s Tip: Content in custom section doesn't require registration\n", blue("ℹ"))

	return fmt.Errorf("validation found unregistered content")
}

type UnregisteredContent struct {
	Rules      []string
	Workflows  []string
	Guidelines []string
}

// extractManagedSection extracts content between managed markers
func extractManagedSection(content string) string {
	startMarker := "<!-- agmd:managed-start -->"
	endMarker := "<!-- agmd:managed-end -->"

	// Find LAST occurrence (actual markers, not examples in rules)
	startIdx := strings.LastIndex(content, startMarker)
	if startIdx == -1 {
		return ""
	}

	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx == -1 {
		return ""
	}

	return content[startIdx : startIdx+endIdx]
}

// detectUnregistered scans managed section for ### headings and checks if they exist
func detectUnregistered(managedContent string, st *state.ProjectState, reg *registry.Registry) UnregisteredContent {
	result := UnregisteredContent{
		Rules:      []string{},
		Workflows:  []string{},
		Guidelines: []string{},
	}

	// Regex to match ### heading-name patterns
	re := regexp.MustCompile(`(?m)^### ([a-z0-9-]+)\s*$`)
	matches := re.FindAllStringSubmatch(managedContent, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		name := match[1]

		// Determine type based on section
		itemType := determineType(managedContent, match[0])

		switch itemType {
		case "rule":
			// Check if in registry
			if _, err := reg.GetRule(name); err != nil {
				// Not in registry - check if it's supposed to be (in agents.toml but missing)
				if !contains(result.Rules, name) {
					result.Rules = append(result.Rules, name)
				}
			}
		case "workflow":
			if _, err := reg.GetWorkflow(name); err != nil {
				if !contains(result.Workflows, name) {
					result.Workflows = append(result.Workflows, name)
				}
			}
		case "guideline":
			if _, err := reg.GetGuideline(name); err != nil {
				if !contains(result.Guidelines, name) {
					result.Guidelines = append(result.Guidelines, name)
				}
			}
		}
	}

	return result
}

// determineType determines if a ### heading is in Rules, Workflows, or Guidelines section
func determineType(content, heading string) string {
	headingPos := strings.Index(content, heading)
	if headingPos == -1 {
		return ""
	}

	beforeHeading := content[:headingPos]

	// Find the last ## section heading before this ### heading
	rulesPos := strings.LastIndex(beforeHeading, "## Rules")
	workflowsPos := strings.LastIndex(beforeHeading, "## Workflows")
	guidelinesPos := strings.LastIndex(beforeHeading, "## Guidelines")

	// Determine which section we're in based on most recent ## heading
	maxPos := rulesPos
	sectionType := "rule"

	if workflowsPos > maxPos {
		maxPos = workflowsPos
		sectionType = "workflow"
	}
	if guidelinesPos > maxPos {
		sectionType = "guideline"
	}

	return sectionType
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
