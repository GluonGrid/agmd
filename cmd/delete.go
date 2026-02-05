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

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:     "delete [type:name]",
	Aliases: []string{"del", "rm"},
	Short:   "Delete an item from the registry",
	Long: `Delete an item from the registry.

This command removes items from the registry (~/.agmd/) permanently.
A confirmation prompt will be shown unless --force is used.

For tasks, use the task subcommand:
  agmd task delete setup-db

Format:
  type:name   - Specify the type and name (e.g., rule:typescript)

Examples:
  agmd delete rule:typescript            # Delete a rule
  agmd rm workflow:old-workflow          # Delete a workflow (using alias)
  agmd del prompt:deprecated             # Delete a prompt (using alias)
  agmd delete rule:frontend/old --force  # Delete without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	itemSpec := args[0]

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Parse item spec (type:name)
	parts := strings.SplitN(itemSpec, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	name := parts[1]

	// Redirect task type
	if itemType == "task" {
		return fmt.Errorf("use 'agmd task delete %s' to delete tasks", name)
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", reg.BasePath)
	}

	// Get item path
	basePath := reg.TypePath(itemType)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("type '%s' does not exist in registry", itemType)
	}

	itemPath := filepath.Join(basePath, name+".md")

	// Check if item exists
	if _, err := os.Stat(itemPath); os.IsNotExist(err) {
		return fmt.Errorf("%s:%s not found at: %s", itemType, name, itemPath)
	}

	// Show what will be deleted
	fmt.Printf("%s Deleting %s:%s\n", blue("→"), itemType, name)
	fmt.Printf("  Path: %s\n", itemPath)

	// Confirmation prompt (unless --force)
	if !deleteForce {
		fmt.Printf("\n%s This will permanently delete this item from the registry.\n", yellow("⚠"))
		fmt.Print("\nAre you sure? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("\nCancelled.")
			return nil
		}
	}

	// Delete the file
	if err := os.Remove(itemPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fmt.Printf("%s Deleted %s:%s\n", green("✓"), itemType, name)

	// Clean up empty directories
	itemDir := filepath.Dir(itemPath)
	if itemDir != basePath {
		// Check if directory is empty
		entries, err := os.ReadDir(itemDir)
		if err == nil && len(entries) == 0 {
			if err := os.Remove(itemDir); err == nil {
				fmt.Printf("%s Removed empty directory: %s\n", green("✓"), filepath.Base(itemDir))
			}
		}
	}

	// Check if entire type folder is empty (for custom types)
	if itemType != "rule" && itemType != "workflow" && itemType != "guideline" {
		entries, err := os.ReadDir(basePath)
		if err == nil && len(entries) == 0 {
			fmt.Printf("\n%s The '%s' type folder is now empty.\n", blue("ℹ"), itemType)
			fmt.Print("Delete the entire type folder? (y/N): ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response == "y" || response == "yes" {
				if err := os.Remove(basePath); err == nil {
					fmt.Printf("%s Removed empty type folder: %s\n", green("✓"), itemType)
				}
			}
		}
	}

	// Warning about directives.md
	fmt.Printf("\n%s If this item is referenced in directives.md, remove or update the reference:\n", yellow("ℹ"))
	fmt.Printf("  :::include %s:%s\n", itemType, name)

	return nil
}
