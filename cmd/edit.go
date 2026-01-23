package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [type:name]",
	Short: "Edit a rule, workflow, or guideline in the registry",
	Long: `Open a registry item in your default editor (specified by $EDITOR environment variable).
Falls back to vim if $EDITOR is not set.

Examples:
  agmd edit rule:typescript
  agmd edit workflow:release
  agmd edit guideline:code-style
  EDITOR=nano agmd edit rule:custom-auth`,
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Require type:name argument
	if len(args) != 1 {
		return fmt.Errorf("usage: agmd edit TYPE:NAME (e.g., 'agmd edit rule:typescript')")
	}

	// Parse type:name
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	name := parts[1]

	// Validate type
	if itemType != "rule" && itemType != "workflow" && itemType != "guideline" {
		return fmt.Errorf("invalid type '%s'. Must be 'rule', 'workflow', or 'guideline'", itemType)
	}

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Get registry paths
	paths := reg.Paths()
	var basePath string
	switch itemType {
	case "rule":
		basePath = paths.Rules
	case "workflow":
		basePath = paths.Workflows
	case "guideline":
		basePath = paths.Guidelines
	}

	filePath := fmt.Sprintf("%s/%s.md", basePath, name)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("%s:%s does not exist at %s", itemType, name, filePath)
	}

	// Get editor from environment, default to vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	fmt.Printf("%s Opening %s:%s (%s) in %s...\n", blue("→"), itemType, name, filePath, editor)

	// Create command to open editor
	editorCmd := exec.Command(editor, filePath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	// Run editor
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Printf("%s Editing complete\n", green("✓"))
	return nil
}
