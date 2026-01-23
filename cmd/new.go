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
	Use:   "new [type:name]",
	Short: "Create a new rule, workflow, or guideline in the registry",
	Long: `Create a new rule, workflow, or guideline and save it to the ~/.agmd/ registry.

Types:
  rule       - Create a new rule
  workflow   - Create a new workflow
  guideline  - Create a new guideline
  [custom]   - Create new type (folder) with confirmation

Names can include subfolders:
  frontend/typescript   - Creates in a subfolder

Examples:
  agmd new rule:typescript            # Create a TypeScript rule
  agmd new rule:frontend/typescript   # Create in frontend/ subfolder
  agmd new workflow:release           # Create a release workflow
  agmd new guideline:code-style       # Create a code style guideline
  agmd new instruction:never          # Creates new 'instruction' type (with prompt)`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	// Parse type:name
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	name := parts[1]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Check if type exists, if not, prompt to create it
	paths := reg.Paths()
	var basePath string
	isCustomType := false

	switch itemType {
	case "rule":
		basePath = paths.Rules
	case "workflow":
		basePath = paths.Workflows
	case "guideline":
		basePath = paths.Guidelines
	default:
		// Custom type - check if folder exists
		customPath := filepath.Join(reg.BasePath, itemType)
		if _, err := os.Stat(customPath); os.IsNotExist(err) {
			// Prompt to create new type
			fmt.Printf("%s Type '%s' doesn't exist yet.\n", yellow("⚠"), itemType)
			fmt.Printf("\nCreate new type '%s' in the registry? This will create a new folder:\n", itemType)
			fmt.Printf("  %s\n\n", customPath)
			fmt.Print("Create? (y/N): ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				return fmt.Errorf("cancelled")
			}

			// Create the new type folder
			if err := os.MkdirAll(customPath, 0755); err != nil {
				return fmt.Errorf("failed to create type folder: %w", err)
			}
			fmt.Printf("%s Created new type: %s\n", green("✓"), itemType)
		}
		basePath = customPath
		isCustomType = true
	}

	fmt.Printf("%s Creating new %s: %s\n", blue("→"), itemType, name)

	// Handle subfolders in name (e.g., "frontend/typescript")
	targetPath := filepath.Join(basePath, name+".md")
	targetDir := filepath.Dir(targetPath)

	// Create subdirectories if needed
	if targetDir != basePath {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory: %w", err)
		}
		fmt.Printf("%s Created subdirectory: %s\n", green("✓"), filepath.Base(targetDir))
	}

	// Check if already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("%s '%s' already exists in registry", itemType, name)
	}

	// Extract base name for template (without path)
	baseName := filepath.Base(name)
	if ext := filepath.Ext(baseName); ext == ".md" {
		baseName = baseName[:len(baseName)-len(ext)]
	}

	// Create template content
	var template string
	if isCustomType {
		template = generateGenericTemplate(itemType, baseName)
	} else {
		template = generateTemplate(itemType, baseName)
	}

	// Write to registry
	if err := os.WriteFile(targetPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", itemType, err)
	}

	fmt.Printf("%s Created %s at: %s\n", green("✓"), itemType, targetPath)

	// Open in default editor
	if err := openInEditor(targetPath); err != nil {
		fmt.Printf("%s Could not open editor: %v\n", yellow("⚠"), err)
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. Edit the file to add your content: %s\n", targetPath)
		fmt.Printf("  2. Add to project: agmd add %s %s\n", itemType, name)
	} else {
		fmt.Println("\nNext step:")
		fmt.Printf("  • After editing, add to project: agmd add %s %s\n", itemType, name)
	}

	return nil
}

func generateGenericTemplate(itemType, name string) string {
	timestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf(`---
name: %s
description: Brief description of this %s
created_at: %s
---

# %s: %s

## Overview

Describe the purpose and content of this %s.

## Content

Add your content here.

## Notes

Additional context or information.
`, name, itemType, timestamp, strings.Title(itemType), strings.Title(strings.ReplaceAll(name, "-", " ")), itemType)
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
