package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

// `task lint` finds tasks that are likely forgotten: old human- tasks from
// before the human flag (invisible to `list --human`), and open tasks that a
// comment says were replaced (they should be closed with `task replace`).

var replacedCommentPattern = regexp.MustCompile(`(?i)\b(replaced|superseded)\b`)

var taskLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "List tasks that are likely forgotten or mislabelled",
	Long: `List tasks that need attention:

  human-name-not-flagged  the name starts with "human-" but the task is not
                          flagged human (fix: agmd task edit <task> --human)
  replaced-but-open       the task is open but a comment says it was replaced
                          or superseded (fix: agmd task replace <task> --by <new>)

Output is one line per finding, "<project>/<task> <kind>: <detail>". The exit
status is 0 whether or not something is found; use --json to read findings.

Examples:
  agmd task lint --project agents
  agmd task lint --project agents --json`,
	Args: cobra.NoArgs,
	RunE: runTaskLint,
}

func init() {
	taskCmd.AddCommand(taskLintCmd)
	taskLintCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskLintCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
}

type taskLintFinding struct {
	Task   string `json:"task"`
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

func lintTasks(reg *registry.Registry, projectName string, tasks []*Task) ([]taskLintFinding, error) {
	findings := []taskLintFinding{}
	sorted := append([]*Task{}, tasks...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	for _, t := range sorted {
		if strings.HasPrefix(t.Name, "human-") && !t.Human {
			findings = append(findings, taskLintFinding{
				Task: t.Name, ID: t.ID, Kind: "human-name-not-flagged",
				Detail: "named human- but not flagged human; hidden from list --human",
			})
		}
		if t.Status == "completed" {
			continue
		}
		comments, err := loadTaskComments(reg, projectName, t.ID)
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			if replacedCommentPattern.MatchString(c.Body) {
				findings = append(findings, taskLintFinding{
					Task: t.Name, ID: t.ID, Kind: "replaced-but-open",
					Detail: fmt.Sprintf("open, but a comment says: %q", firstOutputLine(c.Body)),
				})
				break
			}
		}
	}
	return findings, nil
}

func runTaskLint(cmd *cobra.Command, args []string) error {
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	projectName, err := getProjectName()
	if err != nil {
		return err
	}
	tasks, err := loadProjectTasks(reg, projectName)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}
	findings, err := lintTasks(reg, projectName, tasks)
	if err != nil {
		return fmt.Errorf("failed to load comments: %w", err)
	}
	if taskJSON {
		return printJSON(struct {
			Project  string            `json:"project"`
			Findings []taskLintFinding `json:"findings"`
		}{projectName, findings})
	}
	for _, f := range findings {
		fmt.Printf("%s/%s %s: %s\n", projectName, f.Task, f.Kind, f.Detail)
	}
	return nil
}
