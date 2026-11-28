package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Comments let agents and humans leave a progress log on a task without
// rewriting its description. They live beside the task files, one JSON file
// per task: <project>/comments/<task-id>.json.

type taskComment struct {
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

const defaultCommentAuthor = "user"

var taskCommentAuthor string
var taskCommentDelete int

var taskCommentCmd = &cobra.Command{
	Use:   "comment <task-id-or-name> [body]",
	Short: "Add a comment to a task, or delete one",
	Long: `Add a comment to a task. The body comes from the argument, --content, or stdin.

The author defaults to $AGMD_AUTHOR, or "user". Agents should pass their
workflow id so the board shows who did what.

Comments are shown, numbered from #1 (oldest), by 'agmd task show' and in
the web board. --delete <n> removes comment n; a comment by another author
is refused with "conflict" (status 6) unless you pass --force.

A comment on a completed task is stored, but a warning on stderr says it is
easily missed: reopen the task ('agmd task reopen') or file a new task.

Examples:
  agmd task comment setup-db "Schema migrated, seeding next" --author agent-1
  git log -1 --format=%B | agmd task comment setup-db --author agent-1
  agmd task comment setup-db --delete 2 --author agent-1`,
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: completeTaskName,
	RunE:              runTaskComment,
}

func init() {
	taskCmd.AddCommand(taskCommentCmd)
	taskCommentCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskCommentCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskCommentCmd.Flags().StringVar(&taskCommentAuthor, "author", "", "Comment author (default $AGMD_AUTHOR or \"user\")")
	taskCommentCmd.Flags().StringVar(&taskContent, "content", "", "Comment body")
	taskCommentCmd.Flags().IntVar(&taskCommentDelete, "delete", 0, "Delete comment number n (1 = oldest, as numbered by 'task show')")
	taskCommentCmd.Flags().BoolVar(&taskForce, "force", false, "With --delete: delete a comment by another author")
}

func runTaskComment(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("delete") {
		return runTaskCommentDelete(args)
	}
	body := taskContent
	if len(args) == 2 {
		body = args[1]
	}
	if body == "" && !isTerminal(os.Stdin) {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		body = string(stdinContent)
	}
	author := taskCommentAuthor
	if author == "" {
		author = os.Getenv("AGMD_AUTHOR")
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, taskRef, err := resolveTaskProjectAndRef(args[0])
	if err != nil {
		return err
	}
	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return err
	}
	defer unlock()

	task, err := resolveTaskRef(reg, projectName, taskRef)
	if err != nil {
		return err
	}
	comment, err := addTaskComment(reg, projectName, task.ID, author, body)
	if err != nil {
		return err
	}
	// A comment from the claiming workflow is a sign of life for its claim.
	assignments, err := loadTaskAssignments(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}
	if a, held := assignments[task.ID]; held && a.WorkflowID == comment.Author {
		refreshClaim(assignments, task.ID, time.Now())
		if err := saveTaskAssignments(reg, projectName, assignments); err != nil {
			return fmt.Errorf("failed to save assignments: %w", err)
		}
	}

	if task.Status == "completed" {
		fmt.Fprintf(os.Stderr, "warning: task '%s' is completed; this comment is easily missed. Run 'agmd task reopen %s' to continue the work, or file a new task\n", task.Name, task.Name)
	}

	if taskJSON {
		return printJSON(struct {
			OK      bool        `json:"ok"`
			Action  string      `json:"action"`
			Project string      `json:"project"`
			Task    string      `json:"task"`
			ID      string      `json:"id"`
			Comment taskComment `json:"comment"`
		}{
			OK:      true,
			Action:  "comment",
			Project: projectName,
			Task:    task.Name,
			ID:      task.ID,
			Comment: comment,
		})
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Commented on task '%s'\n", green("✓"), task.Name)
	return nil
}

func getTaskCommentsPath(reg *registry.Registry, projectName, taskID string) string {
	return filepath.Join(getTaskDir(reg, projectName), "comments", taskID+".json")
}

// loadTaskComments returns a task's comments, oldest first.
func loadTaskComments(reg *registry.Registry, projectName, taskID string) ([]taskComment, error) {
	raw, err := os.ReadFile(getTaskCommentsPath(reg, projectName, taskID))
	if err != nil {
		if os.IsNotExist(err) {
			return []taskComment{}, nil
		}
		return nil, err
	}
	comments := []taskComment{}
	if err := json.Unmarshal(raw, &comments); err != nil {
		return nil, fmt.Errorf("invalid comments file for task %s: %w", taskID, err)
	}
	return comments, nil
}

// addTaskComment appends a comment. The caller must hold the project lock.
func addTaskComment(reg *registry.Registry, projectName, taskID, author, body string) (taskComment, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return taskComment{}, newTaskError("invalid_input", "comment body is empty")
	}
	author = strings.TrimSpace(author)
	if author == "" {
		author = defaultCommentAuthor
	}
	comments, err := loadTaskComments(reg, projectName, taskID)
	if err != nil {
		return taskComment{}, err
	}
	comment := taskComment{
		Author:    author,
		Body:      body,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	comments = append(comments, comment)
	if err := saveTaskComments(reg, projectName, taskID, comments); err != nil {
		return taskComment{}, err
	}
	return comment, nil
}

// saveTaskComments replaces a task's comments. The caller must hold the
// project lock.
func saveTaskComments(reg *registry.Registry, projectName, taskID string, comments []taskComment) error {
	path := getTaskCommentsPath(reg, projectName, taskID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(path, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to save comments: %w", err)
	}
	return nil
}

// deleteTaskComment removes comment number n (1 = oldest). Only its author
// may delete it, unless force is set. The caller must hold the project lock.
func deleteTaskComment(reg *registry.Registry, projectName, taskID string, n int, author string, force bool) (taskComment, error) {
	comments, err := loadTaskComments(reg, projectName, taskID)
	if err != nil {
		return taskComment{}, err
	}
	if n < 1 || n > len(comments) {
		return taskComment{}, newTaskError("not_found", "comment %d not found: the task has %d comment(s), numbered from 1 as in 'task show'", n, len(comments))
	}
	removed := comments[n-1]
	if removed.Author != author && !force {
		return taskComment{}, newTaskError("conflict", "comment %d is by '%s', not '%s'; pass --force to delete it anyway", n, removed.Author, author)
	}
	comments = append(comments[:n-1], comments[n:]...)
	if len(comments) == 0 {
		return removed, removeTaskComments(reg, projectName, taskID)
	}
	return removed, saveTaskComments(reg, projectName, taskID, comments)
}

// removeTaskComments deletes the comments of deleted tasks. The caller must
// hold the project lock.
func removeTaskComments(reg *registry.Registry, projectName string, taskIDs ...string) error {
	for _, id := range taskIDs {
		if err := os.Remove(getTaskCommentsPath(reg, projectName, id)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// attachComments fills item.Comments from disk.
func attachComments(reg *registry.Registry, item *taskJSONItem) error {
	comments, err := loadTaskComments(reg, item.Project, item.ID)
	if err != nil {
		return err
	}
	if len(comments) > 0 {
		item.Comments = comments
	}
	return nil
}

func printTaskComments(comments []taskComment) {
	if len(comments) == 0 {
		return
	}
	dim := color.New(color.Faint).SprintFunc()
	fmt.Printf("\n%s\n", dim(fmt.Sprintf("comments (%d):", len(comments))))
	for i, c := range comments {
		fmt.Printf("%s %s %s\n", dim(fmt.Sprintf("#%d", i+1)), dim(c.CreatedAt), c.Author)
		for _, line := range strings.Split(c.Body, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
}

func runTaskCommentDelete(args []string) error {
	if len(args) > 1 || taskContent != "" {
		return newTaskError("invalid_input", "--delete takes no comment body")
	}
	author := taskCommentAuthor
	if author == "" {
		author = os.Getenv("AGMD_AUTHOR")
	}
	if author = strings.TrimSpace(author); author == "" {
		author = defaultCommentAuthor
	}
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, taskRef, err := resolveTaskProjectAndRef(args[0])
	if err != nil {
		return err
	}
	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return err
	}
	defer unlock()

	task, err := resolveTaskRef(reg, projectName, taskRef)
	if err != nil {
		return err
	}
	removed, err := deleteTaskComment(reg, projectName, task.ID, taskCommentDelete, author, taskForce)
	if err != nil {
		return err
	}
	if taskJSON {
		return printJSON(struct {
			OK      bool        `json:"ok"`
			Action  string      `json:"action"`
			Project string      `json:"project"`
			Task    string      `json:"task"`
			ID      string      `json:"id"`
			Index   int         `json:"index"`
			Comment taskComment `json:"comment"`
		}{
			OK:      true,
			Action:  "comment_delete",
			Project: projectName,
			Task:    task.Name,
			ID:      task.ID,
			Index:   taskCommentDelete,
			Comment: removed,
		})
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Deleted comment %d by '%s' on task '%s'\n", green("✓"), taskCommentDelete, removed.Author, task.Name)
	return nil
}
