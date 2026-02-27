package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/generator"
	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

var (
	syncStdout bool
	syncDiff   bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync and generate AGENTS.md from directives.md",
	Long: `Read directives.md and generate expanded AGENTS.md for AI agents.

The command:
1. Checks for unpromoted :::new blocks (errors if found)
2. Reads directives.md (committed, shared with team)
3. Merges directives.local.md if present (gitignored, personal overrides)
4. Expands all :::use and :::list directives with content from registry
5. Writes expanded output to AGENTS.md (or stdout with --stdout)

All non-directive content is preserved.

Flags:
  --stdout    Print generated content to stdout (for hooks). No file written.
  --diff      Print only changes since last sync to stdout. No file written.

Both --stdout and --diff update a local cache so --diff can detect changes.
Status messages go to stderr so stdout is clean content.

Personal overrides:
  Create directives.local.md alongside directives.md for machine-specific
  or personal directives that should not be committed. It is automatically
  appended to directives.md during sync. Add it to .gitignore.

Note: If you have :::new blocks, run 'agmd promote' first to add them to
your registry with proper metadata (name, description).

Examples:
  agmd sync               # Generate AGENTS.md from directives.md
  agmd sync --stdout      # Print expanded content to stdout (for hooks)
  agmd sync --diff        # Print only changes since last --stdout/--diff`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncStdout, "stdout", false, "Print generated content to stdout instead of writing AGENTS.md")
	syncCmd.Flags().BoolVar(&syncDiff, "diff", false, "Print only changes since last sync to stdout")
}

// generateSyncContent runs the full sync pipeline and returns the generated content.
// When quiet is true, no status messages are printed (for hook/pipe usage).
func generateSyncContent(quiet bool) (string, error) {
	logf := func(format string, args ...interface{}) {
		if !quiet {
			fmt.Printf(format+"\n", args...)
		}
	}

	blue := color.New(color.FgBlue).SprintFunc()

	// Check if directives.md exists
	if _, err := os.Stat(directivesMdFilename); err != nil {
		return "", fmt.Errorf("directives.md not found\nRun 'agmd init' first")
	}

	// Load registry
	logf("%s Loading registry...", blue("→"))
	reg, err := registry.New()
	if err != nil {
		return "", fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return "", fmt.Errorf("registry not found at %s\nRun 'agmd setup' first", tildeHome(reg.BasePath))
	}

	// Check for :::new blocks - they must be promoted first
	directivesBytes, err := os.ReadFile(directivesMdFilename)
	if err != nil {
		return "", fmt.Errorf("failed to read directives.md: %w", err)
	}
	newBlocks := detectNewBlocks(string(directivesBytes))
	if len(newBlocks.Items) > 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		logf("%s Found %d :::new blocks that need to be promoted first", yellow("⚠"), len(newBlocks.Items))
		return "", fmt.Errorf("cannot sync with unpromoted :::new blocks")
	}

	// Auto-sync filenames with frontmatter names
	autoSyncRegistryFilenames(reg)

	// Create generator
	gen := generator.New(reg, nil)

	// Check for directives.local.md (personal overrides, gitignored)
	localDirectivesPath := ""
	if _, err := os.Stat(localDirectivesMdFilename); err == nil {
		localDirectivesPath = localDirectivesMdFilename
		dim := color.New(color.Faint).SprintFunc()
		logf("%s Found %s %s", blue("→"), localDirectivesMdFilename, dim("(personal overrides)"))
		addToGitignore(localDirectivesMdFilename)
	}

	// Parse and expand directives from directives.md (+ optional .local)
	logf("%s Parsing and expanding directives...", blue("→"))
	content, err := gen.ParseAndExpand(directivesMdFilename, localDirectivesPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse and expand directives.md: %w", err)
	}

	return content, nil
}

func runSync(cmd *cobra.Command, args []string) error {
	if syncStdout && syncDiff {
		return fmt.Errorf("--stdout and --diff are mutually exclusive")
	}

	if syncStdout {
		return runSyncStdout()
	}
	if syncDiff {
		return runSyncDiff()
	}
	return runSyncFile()
}

// runSyncFile is the original behavior: generate and write AGENTS.md
func runSyncFile() error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("%s Generating AGENTS.md from directives.md...\n", blue("→"))

	content, err := generateSyncContent(false)
	if err != nil {
		return err
	}

	// Write expanded output to AGENTS.md
	if err := os.WriteFile(agentsMdFilename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("\n%s Generated AGENTS.md successfully!\n", green("✓"))
	return nil
}

// runSyncStdout prints generated content to stdout and updates cache
func runSyncStdout() error {
	content, err := generateSyncContent(true)
	if err != nil {
		return err
	}

	// Print content to stdout (for hook consumption)
	fmt.Print(content)

	// Update cache
	if err := writeSyncCache(content); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to update sync cache: %v\n", err)
	}

	return nil
}

// runSyncDiff diffs generated content against cache, prints diff, updates cache
func runSyncDiff() error {
	content, err := generateSyncContent(true)
	if err != nil {
		return err
	}

	// Read cached content
	cached, err := readSyncCache()
	if err != nil {
		// No cache — treat as first run, output everything
		fmt.Print(content)
		if writeErr := writeSyncCache(content); writeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to update sync cache: %v\n", writeErr)
		}
		return nil
	}

	// Compare
	if cached == content {
		// No changes — silent exit
		return nil
	}

	// Compute and print unified diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(cached),
		B:        difflib.SplitLines(content),
		FromFile: "previous",
		ToFile:   "current",
		Context:  3,
	}
	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		// Fallback: just print the full content
		fmt.Print(content)
	} else {
		fmt.Print(diffText)
	}

	// Update cache
	if err := writeSyncCache(content); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to update sync cache: %v\n", err)
	}

	return nil
}

// syncCachePath returns the path to the sync cache file.
// Uses .agmd/cache/ in the project directory.
func syncCachePath() string {
	return filepath.Join(".agmd", "cache", "last-sync.md")
}

// readSyncCache reads the cached sync content.
func readSyncCache() (string, error) {
	data, err := os.ReadFile(syncCachePath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeSyncCache writes content to the sync cache.
func writeSyncCache(content string) error {
	path := syncCachePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Ensure .agmd/cache is gitignored
	agmdDir := filepath.Dir(dir)
	gitignorePath := filepath.Join(agmdDir, ".gitignore")
	if data, err := os.ReadFile(gitignorePath); err != nil || !strings.Contains(string(data), "cache/") {
		existing := ""
		if err == nil {
			existing = string(data)
		}
		if !strings.HasSuffix(existing, "\n") && existing != "" {
			existing += "\n"
		}
		os.WriteFile(gitignorePath, []byte(existing+"cache/\n"), 0644)
	}
	return os.WriteFile(path, []byte(content), 0644)
}
