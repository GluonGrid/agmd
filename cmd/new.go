package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [type] [name]",
	Short: "Create a new rule, workflow, or guideline in the registry",
	Long: `Create a new rule, workflow, or guideline and save it to the ~/.agmd/ registry.

Types:
  rule       - Create a new rule
  workflow   - Create a new workflow
  guideline  - Create a new guideline

Examples:
  agmd new rule typescript       # Create a TypeScript rule
  agmd new workflow release      # Create a release workflow
  agmd new guideline code-style  # Create a code style guideline`,
	Args: cobra.ExactArgs(2),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	itemType := strings.ToLower(args[0])
	name := args[1]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Validate type
	if itemType != "rule" && itemType != "workflow" && itemType != "guideline" {
		return fmt.Errorf("invalid type '%s'. Must be: rule, workflow, or guideline", itemType)
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	fmt.Printf("%s Creating new %s: %s\n", blue("→"), itemType, name)

	// Check if already exists
	paths := reg.Paths()
	var existingPath string
	switch itemType {
	case "rule":
		existingPath = filepath.Join(paths.Rules, name+".md")
	case "workflow":
		existingPath = filepath.Join(paths.Workflows, name+".md")
	case "guideline":
		existingPath = filepath.Join(paths.Guidelines, name+".md")
	}

	if _, err := os.Stat(existingPath); err == nil {
		return fmt.Errorf("%s '%s' already exists in registry", itemType, name)
	}

	// Create template content
	template := generateTemplate(itemType, name)

	// Write to registry
	if err := os.WriteFile(existingPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", itemType, err)
	}

	fmt.Printf("%s Created %s at: %s\n", green("✓"), itemType, existingPath)

	// Open in default editor
	if err := openInEditor(existingPath); err != nil {
		fmt.Printf("%s Could not open editor: %v\n", yellow("⚠"), err)
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. Edit the file to add your content: %s\n", existingPath)
		fmt.Printf("  2. Add to project: agmd add %s %s\n", itemType, name)
	} else {
		fmt.Println("\nNext step:")
		fmt.Printf("  • After editing, add to project: agmd add %s %s\n", itemType, name)
	}

	return nil
}

func generateTemplate(itemType, name string) string {
	timestamp := time.Now().Format(time.RFC3339)

	switch itemType {
	case "rule":
		return fmt.Sprintf(`---
name: %s
category: custom
description: Brief description of this rule
created_at: %s
---

# Rule: %s

## Purpose

Describe what this rule enforces and why it's important.

## Guidelines

- Guideline 1
- Guideline 2
- Guideline 3

## Examples

### Good Example

` + "```" + `
// Example of following the rule
` + "```" + `

### Bad Example

` + "```" + `
// Example of violating the rule
` + "```" + `

## Notes

Additional context, exceptions, or related information.
`, name, timestamp, strings.Title(strings.ReplaceAll(name, "-", " ")))

	case "workflow":
		return fmt.Sprintf(`---
name: %s
description: Brief description of this workflow
created_at: %s
---

# Workflow: %s

## Overview

Describe what this workflow accomplishes and when to use it.

## Prerequisites

- Prerequisite 1
- Prerequisite 2

## Steps

1. **Step 1**: Description
   - Detail about step 1
   - Commands or actions

2. **Step 2**: Description
   - Detail about step 2
   - Commands or actions

3. **Step 3**: Description
   - Detail about step 3
   - Commands or actions

## Verification

How to verify the workflow completed successfully:

` + "```bash" + `
# Verification commands
` + "```" + `

## Troubleshooting

Common issues and solutions:

- **Issue 1**: Solution
- **Issue 2**: Solution
`, name, timestamp, strings.Title(strings.ReplaceAll(name, "-", " ")))

	case "guideline":
		return fmt.Sprintf(`---
name: %s
description: Brief description of this guideline
created_at: %s
---

# Guideline: %s

## Overview

Describe the purpose and scope of this guideline.

## Best Practices

### Practice 1

Description of best practice 1.

` + "```" + `
// Example
` + "```" + `

### Practice 2

Description of best practice 2.

` + "```" + `
// Example
` + "```" + `

### Practice 3

Description of best practice 3.

` + "```" + `
// Example
` + "```" + `

## Anti-Patterns

What to avoid:

- Anti-pattern 1
- Anti-pattern 2

## References

- Link to documentation
- Related resources
`, name, timestamp, strings.Title(strings.ReplaceAll(name, "-", " ")))

	default:
		return ""
	}
}

// openInEditor opens a file in the user's default editor
func openInEditor(filePath string) error {
	// Check for VISUAL, then EDITOR environment variables
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		// Try common editors as fallback
		for _, e := range []string{"vim", "vi", "nano", "code", "subl"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found (set EDITOR or VISUAL environment variable)")
	}

	// Execute editor
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
