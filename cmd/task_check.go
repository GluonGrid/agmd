package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

// A human task can carry a machine check: a shell command whose exit 0 means
// the human step is done (for example "the merge request is merged"). `task
// check` runs the checks of open human tasks and completes those that pass,
// so a step done outside the board does not stay open. It never runs a check
// for a task that is not flagged human.

const defaultCheckAuthor = "check"

var taskCheckTimeout time.Duration
var taskCheckAuthor string
var taskCheckVerbose bool

var taskCheckRunCmd = &cobra.Command{
	Use:   "check",
	Short: "Run the checks of open human tasks and complete those that pass",
	Long: `Run the check command of every open human task in the project.

A check is set with 'agmd task new --human --check "<cmd>"' or
'agmd task edit --check "<cmd>"'. It runs with 'sh -c' in the current folder,
with AGMD_PROJECT and AGMD_TASK set, and is killed after --timeout. Exit 0
means the step is done: the task is completed with the comment
"verified automatically: <first output line>". Any other result leaves the
task as it is. Tasks that are not flagged human are never checked.

Output is one line per completed task, "<project>/<task>", and nothing else
(--verbose adds the checks that did not pass, on stderr). With --json, one
object lists completed and not-done checks.

Examples:
  agmd task check --project agents
  agmd task check --project agents --timeout 10s --json`,
	Args: cobra.NoArgs,
	RunE: runTaskCheck,
}

func init() {
	taskCmd.AddCommand(taskCheckRunCmd)
	taskCheckRunCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskCheckRunCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskCheckRunCmd.Flags().DurationVar(&taskCheckTimeout, "timeout", 30*time.Second, "Time limit per check")
	taskCheckRunCmd.Flags().StringVar(&taskCheckAuthor, "author", "", "Author of the completion comment (default $AGMD_AUTHOR or \"check\")")
	taskCheckRunCmd.Flags().BoolVarP(&taskCheckVerbose, "verbose", "v", false, "Also report checks that did not pass (stderr)")
}

// taskCheckResult is the outcome of one check.
type taskCheckResult struct {
	Task      string `json:"task"`
	ID        string `json:"id"`
	Passed    bool   `json:"-"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out,omitempty"`
	FirstLine string `json:"output,omitempty"`
	Note      string `json:"note,omitempty"`
}

// checkableTasks returns the open human tasks that have a check.
func checkableTasks(tasks []*Task) []*Task {
	var out []*Task
	for _, t := range tasks {
		if t.Human && t.Status != "completed" && strings.TrimSpace(t.Check) != "" {
			out = append(out, t)
		}
	}
	return out
}

// runHumanCheck runs one task's check. It refuses tasks that are not human.
func runHumanCheck(task *Task, timeout time.Duration) taskCheckResult {
	result := taskCheckResult{Task: task.Name, ID: task.ID, ExitCode: -1}
	if !task.Human || strings.TrimSpace(task.Check) == "" {
		return result
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := exec.CommandContext(ctx, "sh", "-c", task.Check)
	c.Env = append(os.Environ(), "AGMD_PROJECT="+task.ProjectName, "AGMD_TASK="+task.Name)
	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = &out
	c.WaitDelay = time.Second
	checkProcessGroup(c)

	err := c.Run()
	result.FirstLine = firstOutputLine(out.String())
	var exitErr *exec.ExitError
	switch {
	case ctx.Err() == context.DeadlineExceeded:
		result.TimedOut = true
	case err == nil:
		result.ExitCode = 0
		result.Passed = true
	case errors.As(err, &exitErr):
		result.ExitCode = exitErr.ExitCode()
	}
	return result
}

func firstOutputLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			if len(line) > 200 {
				line = line[:200]
			}
			return line
		}
	}
	return ""
}

// completeCheckedTask completes a task whose check passed. It takes the
// project lock and re-reads the task, so a task completed, edited or turned
// back into an agent task while its check ran is left alone.
func completeCheckedTask(reg *registry.Registry, projectName string, checked *Task, result *taskCheckResult, author string) (bool, error) {
	unlock, err := lockProject(reg, projectName)
	if err != nil {
		return false, err
	}
	defer unlock()

	task, err := resolveTaskRef(reg, projectName, checked.Name)
	if err != nil {
		return false, nil
	}
	if !task.Human || task.Status == "completed" || task.Check != checked.Check {
		return false, nil
	}
	if err := setTaskStatus(task, "completed"); err != nil {
		return false, err
	}
	if err := saveTask(task); err != nil {
		return false, fmt.Errorf("failed to save task: %w", err)
	}
	note := "verified automatically: " + result.FirstLine
	if result.FirstLine == "" {
		note = "verified automatically: check exited 0"
	}
	result.Note = note
	if _, err := addTaskComment(reg, projectName, task.ID, author, note); err != nil {
		return false, err
	}
	return true, nil
}

func runTaskCheck(cmd *cobra.Command, args []string) error {
	if taskCheckTimeout <= 0 {
		return newTaskError("invalid_input", "--timeout must be positive")
	}
	author := taskCheckAuthor
	if author == "" {
		author = os.Getenv("AGMD_AUTHOR")
	}
	if author == "" {
		author = defaultCheckAuthor
	}
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, err := getProjectName()
	if err != nil {
		return err
	}
	// Checks may be slow; they run without the project lock.
	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	completed := []taskCheckResult{}
	notDone := []taskCheckResult{}
	for _, task := range checkableTasks(tasks) {
		result := runHumanCheck(task, taskCheckTimeout)
		if !result.Passed {
			notDone = append(notDone, result)
			continue
		}
		done, err := completeCheckedTask(reg, projectName, task, &result, author)
		if err != nil {
			return err
		}
		if done {
			completed = append(completed, result)
			if !taskJSON {
				fmt.Printf("%s/%s\n", projectName, task.Name)
			}
		}
	}

	if taskJSON {
		return printJSON(struct {
			Project   string            `json:"project"`
			Completed []taskCheckResult `json:"completed"`
			NotDone   []taskCheckResult `json:"not_done"`
		}{projectName, completed, notDone})
	}
	if taskCheckVerbose {
		for _, r := range notDone {
			reason := fmt.Sprintf("exit %d", r.ExitCode)
			if r.TimedOut {
				reason = "timed out after " + taskCheckTimeout.String()
			}
			fmt.Fprintf(os.Stderr, "not done: %s/%s (%s)\n", projectName, r.Task, reason)
		}
	}
	return nil
}
