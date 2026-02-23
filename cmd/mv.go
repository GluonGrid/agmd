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

var mvToGlobal bool
var mvToLocal bool

var mvCmd = &cobra.Command{
	Use:   "mv [type:name] [new-path]",
	Short: "Move a registry item to a different location or subfolder",
	Long: `Move an item to a different location in the registry.

Supports:
- Moving to subfolders: agmd mv rule:typescript frontend/typescript
- Renaming: agmd mv rule:old-name new-name
- Moving between types: agmd mv rule:typescript prompt:typescript
- Moving between registries: agmd mv rule:my-rule --to-global / --to-local

Examples:
  agmd mv rule:typescript frontend/typescript    # Move to subfolder
  agmd mv rule:old-name new-name                 # Rename
  agmd mv workflow:test frontend/test            # Move to subfolder
  agmd mv rule:my-rule --to-global               # Promote local item to global ~/.agmd
  agmd mv rule:my-rule --to-local                # Move global item to local .agmd/`,
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: completeTypeName,
	RunE:              runMv,
}

func init() {
	rootCmd.AddCommand(mvCmd)
	mvCmd.Flags().BoolVar(&mvToGlobal, "to-global", false, "Move item from local .agmd/ to global ~/.agmd")
	mvCmd.Flags().BoolVar(&mvToLocal, "to-local", false, "Move item from global ~/.agmd to local .agmd/")
}

func runMv(cmd *cobra.Command, args []string) error {
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

	// Handle --to-global / --to-local (registry migration)
	if mvToGlobal || mvToLocal {
		if mvToGlobal && mvToLocal {
			return fmt.Errorf("cannot use --to-global and --to-local together")
		}
		if len(args) != 1 {
			return fmt.Errorf("usage: agmd mv type:name --to-global|--to-local")
		}
		return runMvBetweenRegistries(args[0], mvToGlobal, reg, green, blue)
	}

	if len(args) != 2 {
		return fmt.Errorf("usage: agmd mv type:name new-name  (or use --to-global/--to-local)")
	}

	source := args[0]
	destination := args[1]

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

	// Reject reserved types that have their own subcommands
	if sourceType == "task" || sourceType == "doc" || sourceType == "file" {
		return fmt.Errorf("cannot move %s type items with 'agmd mv'", sourceType)
	}
	if destType == "task" || destType == "doc" || destType == "file" {
		return fmt.Errorf("cannot move items to %s type with 'agmd mv'", destType)
	}

	// Get source path (check local first, then global)
	var sourceBasePath string
	var sourcePath string
	if reg.LocalPath != "" {
		localCandidate := filepath.Join(reg.LocalPath, sourceType, sourceName+".md")
		if _, err := os.Stat(localCandidate); err == nil {
			sourceBasePath = filepath.Join(reg.LocalPath, sourceType)
			sourcePath = localCandidate
		}
	}
	if sourcePath == "" {
		sourceBasePath = reg.TypePath(sourceType)
		sourcePath = filepath.Join(sourceBasePath, sourceName+".md")
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("%s:%s not found", sourceType, sourceName)
	}

	// Get destination path (always same registry as source)
	destBasePath := filepath.Join(filepath.Dir(sourceBasePath), destType)
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

// runMvBetweenRegistries moves an item between local and global registries
func runMvBetweenRegistries(typeAndName string, toGlobal bool, reg *registry.Registry, green, blue func(a ...interface{}) string) error {
	parts := strings.SplitN(typeAndName, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("must be in format 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	itemName := parts[1]

	var srcPath, dstPath string

	if toGlobal {
		// local → global
		if reg.LocalPath == "" {
			return fmt.Errorf("no local registry found in current directory")
		}
		srcPath = filepath.Join(reg.LocalPath, itemType, itemName+".md")
		dstPath = filepath.Join(reg.BasePath, itemType, itemName+".md")
	} else {
		// global → local
		if reg.LocalPath == "" {
			return fmt.Errorf("no local registry found — run 'agmd init --local' first")
		}
		srcPath = filepath.Join(reg.BasePath, itemType, itemName+".md")
		dstPath = filepath.Join(reg.LocalPath, itemType, itemName+".md")
	}

	// Check source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		if toGlobal {
			return fmt.Errorf("%s:%s not found in local registry (.agmd/)", itemType, itemName)
		}
		return fmt.Errorf("%s:%s not found in global registry (~/.agmd/)", itemType, itemName)
	}

	// Check destination doesn't already exist
	if _, err := os.Stat(dstPath); err == nil {
		if toGlobal {
			return fmt.Errorf("%s:%s already exists in global registry — delete it first", itemType, itemName)
		}
		return fmt.Errorf("%s:%s already exists in local registry — delete it first", itemType, itemName)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	direction := "local → global"
	if !toGlobal {
		direction = "global → local"
	}
	fmt.Printf("%s Moving %s:%s (%s)\n", blue("→"), itemType, itemName, direction)
	fmt.Printf("  From: %s\n", srcPath)
	fmt.Printf("  To:   %s\n", dstPath)

	// Read and write (os.Rename won't work across filesystems)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination: %w", err)
	}
	if err := os.Remove(srcPath); err != nil {
		return fmt.Errorf("failed to remove source: %w", err)
	}

	fmt.Printf("%s Moved successfully\n", green("ok"))
	return nil
}
