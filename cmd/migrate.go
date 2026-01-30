package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	migrateForce       bool
	migrateInteractive bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate <file>",
	Short: "Migrate a raw CLAUDE.md or AGENTS.md to agmd format",
	Long: `Migrate an unstructured AI agent configuration file to the agmd format.

Use this for projects that don't use agmd yet - raw CLAUDE.md or AGENTS.md files
with freeform content that needs to be organized into rules, workflows, and guidelines.

The migration process:
1. Creates backup of original file (with .backup suffix)
2. Initializes project if needed (creates directives.md)
3. Walks through each section interactively (with -i flag)
4. Wraps content with :::new markers based on your choices
5. Run 'agmd promote' then 'agmd sync' to save to registry and generate AGENTS.md

For collecting rules from a project that already uses agmd (has directives.md),
use 'agmd collect' instead.

Examples:
  agmd migrate CLAUDE.md              # Migrate (opens editor to organize)
  agmd migrate CLAUDE.md -i           # Interactive walkthrough of sections
  agmd migrate .cursor/rules.md       # Migrate from any location
  agmd migrate existing.md --force    # Re-migrate with force`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVarP(&migrateForce, "force", "f", false, "Allow migration even if directives.md exists")
	migrateCmd.Flags().BoolVarP(&migrateInteractive, "interactive", "i", false, "Interactively walk through each section")
}

// MigrateSection represents a detected section in the source file
type MigrateSection struct {
	Header     string // The header text (without ##)
	Content    string // The content of the section
	StartLine  int    // Line number where section starts
	EndLine    int    // Line number where section ends
	ItemType   string // "rule", "workflow", "guideline", or "" for skip
	ItemName   string // Slugified name for the item
}

func runMigrate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	sourceFile := args[0]

	// Check if source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s not found", sourceFile)
	}

	fmt.Printf("%s Migrating %s...\n", blue("→"), sourceFile)

	// Read source file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sourceFile, err)
	}

	// Create backup IN PLACE (same directory)
	backupPath := sourceFile + ".backup"
	fmt.Printf("%s Creating backup...\n", blue("→"))
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("%s Created backup: %s\n", green("✓"), backupPath)

	// Check if registry exists
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		fmt.Printf("%s Registry not found. Run 'agmd setup' first.\n", yellow("⚠"))
		return fmt.Errorf("registry not initialized")
	}

	// Initialize project if needed
	if _, err := os.Stat(directivesMdFilename); os.IsNotExist(err) {
		fmt.Printf("%s Project not initialized. Initializing...\n", blue("→"))
		if err := runInitCommand(); err != nil {
			return fmt.Errorf("failed to initialize project: %w", err)
		}
		fmt.Printf("%s Project initialized\n", green("✓"))
	}

	// Choose migration mode
	if migrateInteractive {
		return runInteractiveMigrate(string(content))
	}

	return runSimpleMigrate(string(content))
}

// runSimpleMigrate appends content with instructions for manual organization
func runSimpleMigrate(content string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Appending migrated content to directives.md...\n", blue("→"))

	migrateSection := fmt.Sprintf(`

---

## Migrated Content (Organize Below)

Below is your migrated configuration. Use :::new markers to create new rules/workflows:

` + "```" + `
:::new rule:my-rule-name
# Rule: My Rule Name
Content here...
:::end

:::new workflow:my-workflow
# Workflow: My Workflow
Steps here...
:::end
` + "```" + `

---

%s

`, content)

	// Read existing directives.md
	existingDirectives, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Append migrate section
	newDirectives := string(existingDirectives) + migrateSection
	if err := os.WriteFile(directivesMdFilename, []byte(newDirectives), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}
	fmt.Printf("%s Migrated content appended to directives.md\n", green("✓"))

	fmt.Printf("\n%s Migration complete!\n", green("✓"))
	fmt.Println("\n" + blue("→") + " Opening directives.md for you to organize content...")
	fmt.Println()
	fmt.Println("Tips:")
	fmt.Println("  • Migrated content is at the bottom of directives.md")
	fmt.Println("  • Wrap sections with :::new markers to create new registry items")
	fmt.Println("  • Run 'agmd promote' to save items to registry")
	fmt.Println("  • Run 'agmd sync' to generate AGENTS.md")
	fmt.Println()
	fmt.Println("Or use 'agmd migrate -i' for interactive walkthrough.")
	fmt.Println()

	// Open in editor
	if err := openInEditor(directivesMdFilename); err != nil {
		fmt.Printf("%s Could not open editor: %v\n", yellow("⚠"), err)
		fmt.Println("\nManually edit directives.md to organize your content.")
	}

	return nil
}

// runInteractiveMigrate walks through sections interactively using TUI
func runInteractiveMigrate(content string) error {
	yellow := color.New(color.FgYellow).SprintFunc()

	// Detect sections
	sections := detectMigrateSections(content)

	if len(sections) == 0 {
		fmt.Printf("%s No sections (## headers) detected. Falling back to simple mode.\n", yellow("⚠"))
		return runSimpleMigrate(content)
	}

	// Run TUI
	processedSections, err := runMigrateTUI(sections)
	if err != nil {
		return fmt.Errorf("migrate TUI failed: %w", err)
	}

	// Count stats
	rules := 0
	workflows := 0
	guidelines := 0
	skipped := 0

	for _, section := range processedSections {
		switch section.ItemType {
		case "rule":
			rules++
		case "workflow":
			workflows++
		case "guideline":
			guidelines++
		case "":
			skipped++
		}
	}

	return finishInteractiveMigrate(processedSections, rules, workflows, guidelines, skipped)
}

// finishInteractiveMigrate writes the organized content to directives.md
func finishInteractiveMigrate(sections []MigrateSection, rules, workflows, guidelines, skipped int) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Build the directives content
	var builder strings.Builder
	builder.WriteString("\n\n---\n\n## Migrated Content\n\n")

	for _, section := range sections {
		if section.ItemType == "" {
			// Skipped - keep as raw content
			builder.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", section.Header, section.Content))
		} else {
			// Wrapped with :::new
			builder.WriteString(fmt.Sprintf(":::new %s:%s\n", section.ItemType, section.ItemName))
			builder.WriteString(fmt.Sprintf("## %s\n\n%s\n", section.Header, section.Content))
			builder.WriteString(":::end\n\n")
		}
	}

	// Read existing directives.md
	existingDirectives, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return fmt.Errorf("failed to read directives.md: %w", err)
	}

	// Append
	newDirectives := string(existingDirectives) + builder.String()
	if err := os.WriteFile(directivesMdFilename, []byte(newDirectives), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s Migration complete!\n", green("✓"))
	fmt.Printf("  • %d rules\n", rules)
	fmt.Printf("  • %d workflows\n", workflows)
	fmt.Printf("  • %d guidelines\n", guidelines)
	if skipped > 0 {
		fmt.Printf("  • %d skipped\n", skipped)
	}
	fmt.Println()
	fmt.Printf("%s Run 'agmd sync' to promote items and generate AGENTS.md\n", blue("ℹ"))

	return nil
}

// detectMigrateSections parses content and returns sections based on ## headers
func detectMigrateSections(content string) []MigrateSection {
	var sections []MigrateSection

	lines := strings.Split(content, "\n")
	headerRe := regexp.MustCompile(`^##\s+(.+)$`)

	var currentSection *MigrateSection
	var contentBuilder strings.Builder

	for i, line := range lines {
		if match := headerRe.FindStringSubmatch(line); match != nil {
			// Save previous section
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(contentBuilder.String())
				currentSection.EndLine = i - 1
				sections = append(sections, *currentSection)
			}

			// Start new section
			currentSection = &MigrateSection{
				Header:    match[1],
				StartLine: i,
			}
			contentBuilder.Reset()
		} else if currentSection != nil {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	// Save last section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(contentBuilder.String())
		currentSection.EndLine = len(lines) - 1
		sections = append(sections, *currentSection)
	}

	return sections
}

// slugify converts a header to a valid item name
func slugify(s string) string {
	// Lowercase
	s = strings.ToLower(s)
	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove non-alphanumeric characters (except hyphens)
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	s = reg.ReplaceAllString(s, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")
	// Trim hyphens from ends
	s = strings.Trim(s, "-")
	return s
}

// editSectionContent opens section content in editor for editing
func editSectionContent(section MigrateSection) (string, error) {
	// Create temp file with section content
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("agmd-section-%s.md", slugify(section.Header)))

	fullContent := fmt.Sprintf("## %s\n\n%s", section.Header, section.Content)
	if err := os.WriteFile(tmpFile, []byte(fullContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Open in editor
	if err := openInEditor(tmpFile); err != nil {
		return "", err
	}

	// Read back
	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	// Strip the ## header if present (we'll add it back)
	content := string(edited)
	headerRe := regexp.MustCompile(`^##\s+.+\n+`)
	content = headerRe.ReplaceAllString(content, "")

	return strings.TrimSpace(content), nil
}

// runInitCommand is a helper to run project initialization
func runInitCommand() error {
	// Call the init command logic
	return runInit(nil, nil)
}
