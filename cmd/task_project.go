package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// projectMeta holds per-project metadata stored alongside the task files in
// <registry>/task/<project>/project.json. It is kept separate from the task
// markdown files so project-level settings (e.g. a display emoji) survive
// regardless of how many tasks exist.
type projectMeta struct {
	Emoji string `json:"emoji,omitempty"`
}

func getProjectMetaPath(reg *registry.Registry, projectName string) string {
	return filepath.Join(getTaskDir(reg, projectName), "project.json")
}

// loadProjectMeta reads a project's metadata, returning a zero value when no
// metadata file exists yet.
func loadProjectMeta(reg *registry.Registry, projectName string) (projectMeta, error) {
	metaPath := getProjectMetaPath(reg, projectName)
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return projectMeta{}, nil
		}
		return projectMeta{}, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return projectMeta{}, nil
	}
	var meta projectMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return projectMeta{}, fmt.Errorf("invalid project metadata file: %w", err)
	}
	return meta, nil
}

// saveProjectMeta writes a project's metadata, creating the task directory if
// it does not exist yet.
func saveProjectMeta(reg *registry.Registry, projectName string, meta projectMeta) error {
	taskDir := getTaskDir(reg, projectName)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return err
	}
	metaPath := getProjectMetaPath(reg, projectName)
	encoded, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(metaPath, encoded, 0o644)
}

// projectDisplayName prefixes a project name with its emoji when one is set.
func projectDisplayName(name, emoji string) string {
	if strings.TrimSpace(emoji) == "" {
		return name
	}
	return emoji + " " + name
}

// validateProjectEmoji normalizes and validates a user-supplied emoji. It allows
// multi-rune grapheme clusters (e.g. ZWJ family sequences) but rejects empty
// input, whitespace, and unreasonably long values.
func validateProjectEmoji(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", newTaskError("invalid_input", "emoji must not be empty (use --clear to remove it)")
	}
	if !utf8.ValidString(trimmed) {
		return "", newTaskError("invalid_input", "emoji must be valid UTF-8")
	}
	if strings.ContainsAny(trimmed, " \t\r\n") {
		return "", newTaskError("invalid_input", "emoji must be a single token without spaces")
	}
	// Generous cap so ZWJ sequences with skin-tone modifiers still fit while
	// preventing someone from stuffing a whole label in here.
	if utf8.RuneCountInString(trimmed) > 12 {
		return "", newTaskError("invalid_input", "emoji is too long (use a single emoji)")
	}
	return trimmed, nil
}

var taskEmojiClear bool

var taskEmojiCmd = &cobra.Command{
	Use:   "emoji [emoji]",
	Short: "Set, show, or clear the emoji shown before a project's name",
	Long: `Set, show, or clear the emoji shown before a project's name.

The emoji is stored as project metadata and displayed before the project name
in 'agmd task list' and the task web UI.

With no argument the current emoji is shown. Pass an emoji to set it, or use
--clear to remove it.

Examples:
  agmd task emoji 🚀                   # Set the current project's emoji
  agmd task emoji                      # Show the current project's emoji
  agmd task emoji --clear              # Remove the emoji
  agmd task emoji 🐍 --project myproj  # Set emoji for a specific project`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTaskEmoji,
}

func init() {
	taskCmd.AddCommand(taskEmojiCmd)
	taskEmojiCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskEmojiCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskEmojiCmd.Flags().BoolVar(&taskEmojiClear, "clear", false, "Remove the project's emoji")
}

func runTaskEmoji(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	if taskEmojiClear && len(args) > 0 {
		return newTaskError("invalid_input", "do not pass an emoji argument together with --clear")
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	projectName, err := getProjectName()
	if err != nil {
		return err
	}

	meta, err := loadProjectMeta(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load project metadata: %w", err)
	}

	action := "show"
	switch {
	case taskEmojiClear:
		action = "clear"
		meta.Emoji = ""
		if err := saveProjectMeta(reg, projectName, meta); err != nil {
			return fmt.Errorf("failed to save project metadata: %w", err)
		}
	case len(args) == 1:
		action = "set"
		emoji, vErr := validateProjectEmoji(args[0])
		if vErr != nil {
			return vErr
		}
		meta.Emoji = emoji
		if err := saveProjectMeta(reg, projectName, meta); err != nil {
			return fmt.Errorf("failed to save project metadata: %w", err)
		}
	}

	if taskJSON {
		return printJSON(struct {
			OK      bool   `json:"ok"`
			Action  string `json:"action"`
			Project string `json:"project"`
			Emoji   string `json:"emoji"`
		}{OK: true, Action: action, Project: projectName, Emoji: meta.Emoji})
	}

	switch action {
	case "clear":
		fmt.Printf("%s Cleared emoji for project '%s'\n", green("✓"), projectName)
	case "set":
		fmt.Printf("%s Set emoji for project '%s' to %s\n", green("✓"), projectName, meta.Emoji)
	default:
		if meta.Emoji == "" {
			fmt.Printf("No emoji set for project '%s'\n", projectName)
		} else {
			fmt.Printf("%s %s\n", meta.Emoji, projectName)
		}
	}
	return nil
}
