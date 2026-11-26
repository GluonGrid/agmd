package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Run fields are free string settings for whoever starts a worker on the
// task (profile, model, effort, ...). agmd stores them under "run:" in the
// front matter and does not interpret them. They are set with
// `--set run.<key>=<value>` and cleared with `--unset run.<key>` on
// `task new` and `task update`.

var taskSetFields []string
var taskUnsetFields []string

const runFieldPrefix = "run."

var runFieldKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// runFieldChanges is what --set/--unset ask for; an empty value in Set
// removes the key.
type runFieldChanges struct {
	Set   map[string]string
	Unset []string
}

func (c runFieldChanges) empty() bool {
	return len(c.Set) == 0 && len(c.Unset) == 0
}

// parseRunFieldKey checks "run.<key>" and returns <key>.
func parseRunFieldKey(flag, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if !strings.HasPrefix(ref, runFieldPrefix) {
		return "", newTaskError("invalid_input", "%s '%s': only run fields are supported, use run.<key> (for example run.profile)", flag, ref)
	}
	key := strings.TrimPrefix(ref, runFieldPrefix)
	if !runFieldKeyPattern.MatchString(key) {
		return "", newTaskError("invalid_input", "%s '%s': key must be lowercase letters, digits, '-' or '_'", flag, ref)
	}
	return key, nil
}

// parseRunFieldFlags validates --set run.k=v and --unset run.k values.
func parseRunFieldFlags(sets, unsets []string) (runFieldChanges, error) {
	c := runFieldChanges{Set: map[string]string{}}
	for _, s := range sets {
		ref, value, ok := strings.Cut(s, "=")
		if !ok {
			return c, newTaskError("invalid_input", "--set '%s': use run.<key>=<value>", s)
		}
		key, err := parseRunFieldKey("--set", ref)
		if err != nil {
			return c, err
		}
		if strings.ContainsAny(value, "\n\r") {
			return c, newTaskError("invalid_input", "--set '%s': value must be one line", ref)
		}
		c.Set[key] = strings.TrimSpace(value)
	}
	for _, u := range unsets {
		key, err := parseRunFieldKey("--unset", u)
		if err != nil {
			return c, err
		}
		if _, both := c.Set[key]; both {
			return c, newTaskError("invalid_input", "run.%s is both set and unset", key)
		}
		c.Unset = append(c.Unset, key)
	}
	return c, nil
}

// applyRunFields applies c to task.Run and describes each change.
func applyRunFields(task *Task, c runFieldChanges) []string {
	var changes []string
	keys := make([]string, 0, len(c.Set))
	for k := range c.Set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.Set[k]
		if v == "" {
			if _, ok := task.Run[k]; ok {
				delete(task.Run, k)
				changes = append(changes, "run."+k+" cleared")
			}
			continue
		}
		if task.Run == nil {
			task.Run = map[string]string{}
		}
		task.Run[k] = v
		changes = append(changes, fmt.Sprintf("run.%s=%s", k, v))
	}
	for _, k := range c.Unset {
		if _, ok := task.Run[k]; ok {
			delete(task.Run, k)
			changes = append(changes, "run."+k+" cleared")
		}
	}
	if len(task.Run) == 0 {
		task.Run = nil
	}
	return changes
}

// formatRunFields renders run fields as "k=v, k=v" in key order.
func formatRunFields(run map[string]string) string {
	keys := make([]string, 0, len(run))
	for k := range run {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+run[k])
	}
	return strings.Join(parts, ", ")
}
