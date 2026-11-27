package cmd

import (
	"fmt"
	"io"
	"strings"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

var (
	noteTitle string
	noteTasks string
	noteAll   bool
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Read and add to the user's notes",
	Long: `The user's notes, one Markdown file per note:
  ~/.agmd/note/<project>/<slug>.md   (front matter: title, created, updated, tasks)
The same files as the Notes tab of the Agents app, under the same lock. Project = main
repo folder name, as for tasks; override with --project or --repo-path.

  list [--all]                 the notes of the project (slug, title, updated);
                               --all: the projects that have notes
  show SLUG                    the note's text
  new --title T [TEXT|-]       a new note (TEXT, or - for stdin); prints its slug
  append SLUG TEXT|-           adds TEXT as a new paragraph (a link alone on its
                               line shows as a card in the app)
  path SLUG                    the file: read it, write only through this command

The slug comes from the title (lowercase a-z 0-9 joined by -, at most 60
characters; -2, -3 when taken) and never changes. A TEXT that starts with -
(a list item) goes after --:  agmd note append SLUG -- "- [ ] step dots".`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the project's notes (or, with --all, the projects that have notes)",
	Args:  cobra.NoArgs,
	RunE:  runNoteList,
}

var noteShowCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Print a note",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteShow,
}

var notePathCmd = &cobra.Command{
	Use:   "path <slug>",
	Short: "Print a note's file path",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotePath,
}

var noteNewCmd = &cobra.Command{
	Use:   "new --title T [TEXT|-]",
	Short: "Make a new note and print its slug",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runNoteNew,
}

var noteAppendCmd = &cobra.Command{
	Use:   "append <slug> TEXT|-",
	Short: "Add a paragraph to the end of a note",
	Args:  cobra.ExactArgs(2),
	RunE:  runNoteAppend,
}

func init() {
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteListCmd, noteShowCmd, notePathCmd, noteNewCmd, noteAppendCmd)
	noteCmd.PersistentFlags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	noteCmd.PersistentFlags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	noteCmd.SetFlagErrorFunc(noteFlagError)
	noteListCmd.Flags().BoolVar(&noteAll, "all", false, "List the projects that have notes")
	noteNewCmd.Flags().StringVar(&noteTitle, "title", "", "The note's title (required)")
	noteNewCmd.Flags().StringVar(&noteTasks, "tasks", "", "Tasks the note belongs to, comma-separated")
}

// noteFlagError adds the way out for the usual cause: a TEXT starting with -.
func noteFlagError(cmd *cobra.Command, err error) error {
	return fmt.Errorf("%w (a TEXT that starts with - goes after --: agmd note append SLUG -- \"- item\")", err)
}

func noteContext() (*registry.Registry, string, error) {
	reg, err := registry.New()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load registry: %w", err)
	}
	project, err := getProjectName()
	if err != nil {
		return nil, "", err
	}
	return reg, project, checkNoteProject(project)
}

// noteText is the TEXT argument, or stdin for "-".
func noteText(cmd *cobra.Command, arg string) (string, error) {
	if arg != "-" {
		return arg, nil
	}
	raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), noteBodyMax+1))
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(raw) > noteBodyMax { // refused, never cut: the read stops one byte past the limit
		return "", newTaskError("invalid_input", "the text on stdin is too long (512 KB at most)")
	}
	return string(raw), nil
}

func runNoteList(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	if noteAll {
		reg, err := registry.New()
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
		projects, err := listNoteProjects(reg)
		if err != nil {
			return err
		}
		for _, p := range projects {
			fmt.Fprintf(out, "%s\t%d notes\n", p.Name, p.Notes)
		}
		return nil
	}
	reg, project, err := noteContext()
	if err != nil {
		return err
	}
	notes, err := listNotes(reg, project)
	if err != nil {
		return err
	}
	for _, n := range notes {
		when := ""
		if t := n.UpdatedTime(); !t.IsZero() {
			when = t.Local().Format("2006-01-02 15:04")
		}
		fmt.Fprintf(out, "%s\t%s\t%s\n", n.Slug, n.Title, when)
	}
	return nil
}

func runNoteShow(cmd *cobra.Command, args []string) error {
	reg, project, err := noteContext()
	if err != nil {
		return err
	}
	n, err := readNote(reg, project, args[0])
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if !strings.HasPrefix(strings.TrimLeft(n.Body, " \t\r\n"), "# ") {
		fmt.Fprintf(out, "# %s\n", n.Title)
	}
	fmt.Fprint(out, n.Body)
	return nil
}

func runNotePath(cmd *cobra.Command, args []string) error {
	reg, project, err := noteContext()
	if err != nil {
		return err
	}
	n, err := readNote(reg, project, args[0])
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), n.Path)
	return nil
}

func runNoteNew(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(noteTitle) == "" {
		return newTaskError("invalid_input", "a note needs a title: --title T")
	}
	reg, project, err := noteContext()
	if err != nil {
		return err
	}
	body := ""
	if len(args) == 1 {
		if body, err = noteText(cmd, args[0]); err != nil {
			return err
		}
	}
	n, err := createNote(reg, project, noteTitle, body, noteTasks)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), n.Slug)
	return nil
}

func runNoteAppend(cmd *cobra.Command, args []string) error {
	reg, project, err := noteContext()
	if err != nil {
		return err
	}
	text, err := noteText(cmd, args[1])
	if err != nil {
		return err
	}
	n, err := appendNote(reg, project, args[0], text)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "added to %s/%s\n", n.Project, n.Slug)
	return nil
}
