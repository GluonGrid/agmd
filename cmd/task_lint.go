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

// replacedStatementPattern finds "replaced by <name>" or "superseded by
// <name>"; group 1 holds up to three words before it, group 2 the name.
var replacedStatementPattern = regexp.MustCompile("(?i)((?:[\\w']+\\s+){0,3})\\b(?:replaced|superseded)\\s+by\\s+[`'\"]?([\\w.-]+(?:/[\\w.-]+)?)")

// notYetWords before "replaced by" make the sentence a plan, a question or a
// negation, not a statement that it happened.
var notYetWords = map[string]bool{
	"not": true, "never": true, "isn't": true, "wasn't": true, "hasn't": true,
	"will": true, "would": true, "should": true, "could": true, "can": true,
	"may": true, "might": true, "must": true, "to": true, "if": true,
	"maybe": true, "perhaps": true, "whether": true,
}

// saysReplaced reports whether a comment states that the task was replaced
// or superseded by another task: "replaced by <name>" where <name> is a task
// of the project, a <project>/<task> reference, or a kebab-case task name,
// and the words before it do not make it a plan or a negation.
func saysReplaced(body string, taskMap map[string]*Task) bool {
	for _, m := range replacedStatementPattern.FindAllStringSubmatch(body, -1) {
		negated := false
		for _, w := range strings.Fields(strings.ToLower(m[1])) {
			if notYetWords[w] {
				negated = true
			}
		}
		if negated {
			continue
		}
		name := strings.TrimRight(m[2], ".")
		if taskMap[name] != nil || strings.ContainsAny(name, "/-") {
			return true
		}
	}
	return false
}

var taskLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "List tasks that are likely forgotten or mislabelled",
	Long: `List tasks that need attention:

  human-name-not-flagged  the name starts with "human-" but the task is not
                          flagged human (fix: agmd task update <task> --human)
  replaced-but-open       the task is open but a comment states it was
                          "replaced by <task>" or "superseded by <task>"
                          (fix: agmd task replace <task> --by <new>)

Only open tasks are checked; --all also checks completed ones.

Output is one line per finding, "<project>/<task> <kind>: <detail>". The exit
status is 0 whether or not something is found; use --json to read findings.

Examples:
  agmd task lint --project agents
  agmd task lint --project agents --all --json`,
	Args: cobra.NoArgs,
	RunE: runTaskLint,
}

func init() {
	taskCmd.AddCommand(taskLintCmd)
	taskLintCmd.Flags().StringVar(&taskProject, "project", "", "Project name (stable selector, overrides automatic resolution)")
	taskLintCmd.Flags().StringVar(&taskRepoPath, "repo-path", "", "Resolve project from this repository path (worktree-aware)")
	taskLintCmd.Flags().BoolVarP(&taskAll, "all", "a", false, "Also check completed tasks")
}

type taskLintFinding struct {
	Task   string `json:"task"`
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

// lintTasks checks open tasks, and completed ones too when includeCompleted.
func lintTasks(reg *registry.Registry, projectName string, tasks []*Task, includeCompleted bool) ([]taskLintFinding, error) {
	findings := []taskLintFinding{}
	taskMap := buildTaskMap(tasks)
	sorted := append([]*Task{}, tasks...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	for _, t := range sorted {
		if t.Status == "completed" && !includeCompleted {
			continue
		}
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
			if saysReplaced(c.Body, taskMap) {
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
	findings, err := lintTasks(reg, projectName, tasks, taskAll)
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
