package cmd

import (
	"fmt"
	"os"
	"strings"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

var showRaw bool

var showCmd = &cobra.Command{
	Use:   "show <type:name>",
	Short: "Show the content of a registry item",
	Long: `Display the content of a registry item (like cat).

Useful for AI assistants to read item content without opening an editor.

Examples:
  agmd show rule:typescript        # Show rule content
  agmd show workflow:commit        # Show workflow content
  agmd show guide:agmd             # Show guide content
  agmd show rule:typescript --raw  # Include frontmatter`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolVar(&showRaw, "raw", false, "Include frontmatter in output")
}

func runShow(cmd *cobra.Command, args []string) error {
	// Parse type:name
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}

	itemType := parts[0]
	name := parts[1]

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Get item
	item, err := reg.GetItem(itemType, name)
	if err != nil {
		return fmt.Errorf("%s:%s not found", itemType, name)
	}

	// Output content
	if showRaw {
		// Read raw file with frontmatter
		raw, err := os.ReadFile(item.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		fmt.Print(string(raw))
	} else {
		// Just the content without frontmatter
		fmt.Print(item.Content)
	}

	return nil
}
