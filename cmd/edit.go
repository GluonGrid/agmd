package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var editContent string

var editCmd = &cobra.Command{
	Use:   "edit [type:name]",
	Short: "Edit directives.md or a registry item",
	Long: `Open directives.md or a registry item in your default editor.

Without arguments, opens directives.md in the current directory.
With type:name argument, opens that item from the registry.

Examples:
  agmd edit                    # Edit directives.md
  agmd edit rule:typescript    # Edit rule from registry
  agmd edit framework:agmd     # Edit custom type from registry
  agmd edit profile:default    # Edit a profile template

For AI assistants (non-interactive):
  agmd edit rule:test --content "# Updated content"
  echo "New content" | agmd edit rule:test`,
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().StringVar(&editContent, "content", "", "Replace content (skips editor)")
}

func runEdit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Determine content source (for non-interactive editing)
	var newContent string
	if editContent != "" {
		newContent = editContent
	} else if !isTerminal(os.Stdin) {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		newContent = string(stdinContent)
	}

	// No arguments = edit directives.md
	if len(args) == 0 {
		if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
			return fmt.Errorf("directives.md not found\nRun 'agmd init' first")
		}

		if newContent != "" {
			if err := os.WriteFile(directivesMdFilename, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to write directives.md: %w", err)
			}
			fmt.Printf("%s Updated directives.md\n", green("ok"))
			return nil
		}

		return openInEditor(directivesMdFilename)
	}

	// Parse type:name
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	name := parts[1]

	// Load registry
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Build file path
	filePath := filepath.Join(reg.BasePath, itemType, name+".md")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("%s:%s not found at %s", itemType, name, filePath)
	}

	// Non-interactive edit
	if newContent != "" {
		// Preserve frontmatter, replace content
		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Extract frontmatter
		frontmatter := extractFrontmatterString(string(existingContent))

		var updatedContent string
		if frontmatter != "" {
			updatedContent = frontmatter + "\n" + strings.TrimSpace(newContent)
		} else {
			// Add default frontmatter
			updatedContent = fmt.Sprintf("---\nname: %s\ndescription: \"\"\n---\n\n%s", name, strings.TrimSpace(newContent))
		}

		if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("%s Updated %s:%s\n", green("ok"), itemType, name)
		return nil
	}

	fmt.Printf("%s Opening %s:%s...\n", blue("->"), itemType, name)
	return openInEditor(filePath)
}

// extractFrontmatterString extracts frontmatter (including delimiters) from content
func extractFrontmatterString(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}

	// Find closing ---
	endIdx := strings.Index(content[4:], "\n---")
	if endIdx == -1 {
		return ""
	}

	return content[:4+endIdx+4] + "\n"
}

// openInEditor opens a file in the user's editor
func openInEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vim"
	}

	editorCmd := exec.Command(editor, filePath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}
