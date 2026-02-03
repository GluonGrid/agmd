package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var newNoEditor bool
var newContent string

var newCmd = &cobra.Command{
	Use:   "new type:name",
	Short: "Create a new item in the registry",
	Long: `Create a new item in the registry and open it in your editor.

The type can be anything - agmd will create the folder if needed.

Examples:
  agmd new rule:typescript
  agmd new framework:react
  agmd new prompt:code-review
  agmd new profile:svelte-kit

For AI assistants (non-interactive):
  agmd new rule:test --no-editor
  agmd new rule:test --content "# My Rule\nContent here"
  echo "# My Rule" | agmd new rule:test --no-editor`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().BoolVar(&newNoEditor, "no-editor", false, "Don't open editor after creating")
	newCmd.Flags().StringVar(&newContent, "content", "", "Content to write (skips editor)")
}

func runNew(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Parse type:name
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use 'type:name' (e.g., 'rule:typescript')")
	}
	itemType := strings.ToLower(parts[0])
	name := parts[1]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Handle profile creation (special case)
	if itemType == "profile" {
		return createProfile(name, reg)
	}

	// Build path
	filePath := filepath.Join(reg.BasePath, itemType, name+".md")

	// Check if exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("%s:%s already exists", itemType, name)
	}

	// Create directory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Determine content source
	var content string

	if newContent != "" {
		// Use --content flag
		content = newContent
	} else if !isTerminal(os.Stdin) {
		// Read from stdin (piped input)
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = string(stdinContent)
	}

	// Build file content with frontmatter
	var fileContent string
	if content != "" {
		// Use provided content
		fileContent = fmt.Sprintf(`---
name: %s
description: ""
---

%s`, name, strings.TrimSpace(content))
	} else {
		// Use template
		fileContent = fmt.Sprintf(`---
name: %s
description: ""
---

# %s

`, name, strings.Title(strings.ReplaceAll(name, "-", " ")))
	}

	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	fmt.Printf("%s Created %s:%s\n", green("ok"), itemType, name)

	// Open editor unless --no-editor or content was provided
	if newNoEditor || content != "" {
		fmt.Printf("%s %s\n", blue("->"), filePath)
		return nil
	}

	fmt.Printf("%s Opening editor...\n", blue("->"))
	return openInEditor(filePath)
}

// isTerminal checks if a file is a terminal (not piped)
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return true
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// createProfile creates a new profile from current directives.md
func createProfile(name string, reg *registry.Registry) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Check directives.md exists
	if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("directives.md not found\nRun 'agmd init' first")
	}

	// Read content
	content, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Check for :::new blocks
	newBlocks := detectNewBlocks(string(content))
	if len(newBlocks.Items) > 0 {
		return fmt.Errorf("directives.md has unpromoted :::new blocks\nRun 'agmd promote' first")
	}

	// Check if profile exists
	profilePath := filepath.Join(reg.BasePath, "profile", name+".md")
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile:%s already exists", name)
	}

	// Save profile
	profile := registry.Profile{
		Name:    name,
		Content: string(content),
	}

	if err := reg.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Printf("%s Created profile:%s\n", green("ok"), name)
	fmt.Printf("\n%s Use in new project: agmd init profile:%s\n", blue("->"), name)

	return nil
}
