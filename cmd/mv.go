package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv [type:name] [new-path]",
	Short: "Move a registry item to a different location or subfolder",
	Long: `Move a rule, workflow, or guideline to a different location in the registry.

Supports:
- Moving to subfolders: agmd mv rule:typescript frontend/typescript
- Renaming: agmd mv rule:old-name new-name
- Moving between types: agmd mv rule:typescript workflow:typescript

Examples:
  agmd mv rule:typescript frontend/typescript    # Move to subfolder
  agmd mv rule:old-name new-name                 # Rename
  agmd mv workflow:test frontend/test            # Move workflow to subfolder`,
	Args: cobra.ExactArgs(2),
	RunE: runMv,
}

func init() {
	rootCmd.AddCommand(mvCmd)
}

func runMv(cmd *cobra.Command, args []string) error {
	source := args[0]
	destination := args[1]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Parse source (type:name)
	sourceParts := strings.SplitN(source, ":", 2)
	if len(sourceParts) != 2 {
		return fmt.Errorf("source must be in format 'type:name' (e.g., 'rule:typescript')")
	}
	sourceType := strings.ToLower(sourceParts[0])
	sourceName := sourceParts[1]

	// Parse destination (can be type:name or just name)
	var destType, destName string
	if strings.Contains(destination, ":") {
		destParts := strings.SplitN(destination, ":", 2)
		destType = strings.ToLower(destParts[0])
		destName = destParts[1]
	} else {
		destType = sourceType // Same type
		destName = destination
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	paths := reg.Paths()

	// Get source path
	var sourceBasePath string
	switch sourceType {
	case "rule":
		sourceBasePath = paths.Rules
	case "workflow":
		sourceBasePath = paths.Workflows
	case "guideline":
		sourceBasePath = paths.Guidelines
	default:
		sourceBasePath = filepath.Join(reg.BasePath, sourceType)
	}

	sourcePath := filepath.Join(sourceBasePath, sourceName+".md")

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("%s:%s not found at: %s", sourceType, sourceName, sourcePath)
	}

	// Get destination path
	var destBasePath string
	switch destType {
	case "rule":
		destBasePath = paths.Rules
	case "workflow":
		destBasePath = paths.Workflows
	case "guideline":
		destBasePath = paths.Guidelines
	default:
		// Custom type - check if exists or create
		destBasePath = filepath.Join(reg.BasePath, destType)
		if _, err := os.Stat(destBasePath); os.IsNotExist(err) {
			fmt.Printf("%s Type '%s' doesn't exist yet.\n", yellow("⚠"), destType)
			fmt.Printf("\nCreate new type '%s'? (y/N): ", destType)

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				return fmt.Errorf("cancelled")
			}

			if err := os.MkdirAll(destBasePath, 0755); err != nil {
				return fmt.Errorf("failed to create type folder: %w", err)
			}
			fmt.Printf("%s Created new type: %s\n", green("✓"), destType)
		}
	}

	destPath := filepath.Join(destBasePath, destName+".md")
	destDir := filepath.Dir(destPath)

	// Create destination subdirectories if needed
	if destDir != destBasePath {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory: %w", err)
		}
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination already exists: %s", destPath)
	}

	fmt.Printf("%s Moving %s:%s → %s:%s\n", blue("→"), sourceType, sourceName, destType, destName)
	fmt.Printf("  From: %s\n", sourcePath)
	fmt.Printf("  To:   %s\n", destPath)

	// Move the file
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	fmt.Printf("%s Moved successfully\n", green("✓"))

	// Clean up empty source directories
	sourceDir := filepath.Dir(sourcePath)
	if sourceDir != sourceBasePath {
		// Check if directory is empty
		entries, err := os.ReadDir(sourceDir)
		if err == nil && len(entries) == 0 {
			if err := os.Remove(sourceDir); err == nil {
				fmt.Printf("%s Removed empty directory: %s\n", green("✓"), filepath.Base(sourceDir))
			}
		}
	}

	fmt.Println("\nNote: If this item is referenced in directives.md, update the reference:")
	fmt.Printf("  Old: @%s:%s\n", sourceType, sourceName)
	fmt.Printf("  New: @%s:%s\n", destType, destName)

	return nil
}
