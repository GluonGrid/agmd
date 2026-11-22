package cmd

// JSON contract for task dependencies.
//
// Every task item carries "depends_on" (the references as stored, normally
// task names; unchanged for existing readers) and "dependencies", the same
// list resolved to canonical {id, name}. 'task list --json --tree' returns
// the flat list plus "edges", one per dependency, for rendering a graph.

// taskDependencyRef is one dependency resolved to its canonical id and name.
// Missing is set when the reference matches no task in the project; the
// name is then the stored reference and there is no id.
type taskDependencyRef struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Missing bool   `json:"missing,omitempty"`
}

// taskDependencyEdge says that task From depends on task To. Satisfied
// reports whether To is done (completed and past any hold), so a renderer
// knows the edge's state even when To is hidden from the list.
type taskDependencyEdge struct {
	From      string `json:"from"`
	FromID    string `json:"from_id"`
	To        string `json:"to"`
	ToID      string `json:"to_id,omitempty"`
	Missing   bool   `json:"missing,omitempty"`
	Satisfied bool   `json:"satisfied"`
}

type taskGraphJSONPayload struct {
	taskListJSONPayload
	Edges []taskDependencyEdge `json:"edges"`
}

// findDependency resolves a stored dependency reference (a name, or an id
// in older files) against the project's tasks, keyed by name.
func findDependency(ref string, taskMap map[string]*Task) *Task {
	if dep, ok := taskMap[ref]; ok {
		return dep
	}
	for _, t := range taskMap {
		if t.ID != "" && t.ID == ref {
			return t
		}
	}
	return nil
}

func resolveTaskDependencies(task *Task, taskMap map[string]*Task) []taskDependencyRef {
	refs := make([]taskDependencyRef, 0, len(task.DependsOn))
	for _, ref := range task.DependsOn {
		if dep := findDependency(ref, taskMap); dep != nil {
			refs = append(refs, taskDependencyRef{ID: dep.ID, Name: dep.Name})
		} else {
			refs = append(refs, taskDependencyRef{Name: ref, Missing: true})
		}
	}
	return refs
}

// buildTaskGraphJSON is the flat list payload plus one edge per dependency
// of each listed task. An edge's target may be absent from tasks (completed
// and hidden without --all, filtered out, or missing).
func buildTaskGraphJSON(projectName, feature string, tasks []*Task, taskMap map[string]*Task, assignments map[string]taskAssignment) taskGraphJSONPayload {
	payload := taskGraphJSONPayload{
		taskListJSONPayload: buildTaskListJSON(projectName, feature, tasks, taskMap, assignments, false),
		Edges:               []taskDependencyEdge{},
	}
	now := taskNow()
	for _, task := range tasks {
		for _, ref := range resolveTaskDependencies(task, taskMap) {
			edge := taskDependencyEdge{
				From:    task.Name,
				FromID:  task.ID,
				To:      ref.Name,
				ToID:    ref.ID,
				Missing: ref.Missing,
			}
			if !ref.Missing {
				edge.Satisfied = dependencySatisfied(taskMap[ref.Name], now)
			}
			payload.Edges = append(payload.Edges, edge)
		}
	}
	return payload
}

// printTaskListJSON prints the list as the flat payload, or with --tree as
// the graph payload.
func printTaskListJSON(projectName string, tasks []*Task, taskMap map[string]*Task, assignments map[string]taskAssignment) error {
	if taskTree {
		return printJSON(buildTaskGraphJSON(projectName, taskFeature, tasks, taskMap, assignments))
	}
	return printJSON(buildTaskListJSON(projectName, taskFeature, tasks, taskMap, assignments, false))
}
