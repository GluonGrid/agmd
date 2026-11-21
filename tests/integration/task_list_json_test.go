package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// listJSON runs 'task list' with the given extra flags and decodes it
// generically, so the tests see the exact keys a reader would see.
func listJSON(t *testing.T, extra ...string) map[string]interface{} {
	t.Helper()
	args := append([]string{"task", "list", "--project", "p", "--json"}, extra...)
	out, err := runE(t, args...)
	if err != nil {
		t.Fatalf("task list: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse %q: %v", out, err)
	}
	return payload
}

func keysOf(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func listedTask(t *testing.T, payload map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, raw := range payload["tasks"].([]interface{}) {
		task := raw.(map[string]interface{})
		if task["name"] == name {
			return task
		}
	}
	t.Fatalf("task %s not in list", name)
	return nil
}

func seedDependencyTasks(t *testing.T) {
	t.Helper()
	run(t, "task", "new", "setup-db", "--project", "p", "--content", "schema")
	run(t, "task", "new", "create-api", "--project", "p", "-p", "1", "-t", "feature", "--content", "api")
	run(t, "task", "new", "merge-pr", "--project", "p", "--human", "--content", "merge")
	run(t, "task", "dep", "add", "create-api", "setup-db", "--project", "p")
	run(t, "task", "dep", "add", "create-api", "merge-pr", "--project", "p")
}

// The flat list is read live by a harness: its keys and the types of the
// fields it reads must not change. "dependencies" is the only addition.
func TestTaskListJSON_FlatShapeUnchanged(t *testing.T) {
	setup(t)
	chdir(t)
	seedDependencyTasks(t)

	payload := listJSON(t)
	if got := keysOf(payload); got != "counts,project,tasks" {
		t.Errorf("top-level keys = %s, want counts,project,tasks (no edges)", got)
	}
	if got := keysOf(payload["counts"].(map[string]interface{})); got != "blocked,completed,in_progress,ready,total" {
		t.Errorf("counts keys = %s", got)
	}

	api := listedTask(t, payload, "create-api")
	if got, want := keysOf(api), "computed_status,created_at,dependencies,depends_on,id,name,pending_dependencies,priority,project,status,subject,type,updated_at"; got != want {
		t.Errorf("task keys = %s\nwant       %s", got, want)
	}
	if api["status"] != "pending" || api["computed_status"] != "blocked" || api["priority"] != float64(1) || api["type"] != "feature" {
		t.Errorf("harness fields changed: %v", api)
	}
	deps, ok := api["depends_on"].([]interface{})
	if !ok || len(deps) != 2 || deps[0] != "setup-db" || deps[1] != "merge-pr" {
		t.Errorf("depends_on = %v, want [setup-db merge-pr] as names", api["depends_on"])
	}
	if human := listedTask(t, payload, "merge-pr")["human"]; human != true {
		t.Errorf("merge-pr human = %v, want true", human)
	}
	if _, present := listedTask(t, payload, "setup-db")["human"]; present {
		t.Error("human should stay omitted for agent tasks")
	}
	if deps := listedTask(t, payload, "setup-db")["depends_on"]; deps == nil {
		t.Error("depends_on should be [] (not null or missing) without dependencies")
	}
}

func TestTaskListJSON_DependenciesCarryIDsAndNames(t *testing.T) {
	registry := setup(t)
	chdir(t)
	seedDependencyTasks(t)

	payload := listJSON(t)
	ids := map[string]string{}
	for _, raw := range payload["tasks"].([]interface{}) {
		task := raw.(map[string]interface{})
		ids[task["name"].(string)] = task["id"].(string)
	}
	refs := listedTask(t, payload, "create-api")["dependencies"].([]interface{})
	if len(refs) != 2 {
		t.Fatalf("dependencies = %v, want 2", refs)
	}
	for i, name := range []string{"setup-db", "merge-pr"} {
		ref := refs[i].(map[string]interface{})
		if ref["name"] != name || ref["id"] != ids[name] || ref["id"] == "" {
			t.Errorf("dependencies[%d] = %v, want {id:%s name:%s}", i, ref, ids[name], name)
		}
	}
	if refs := listedTask(t, payload, "setup-db")["dependencies"].([]interface{}); len(refs) != 0 {
		t.Errorf("setup-db dependencies = %v, want []", refs)
	}

	// A reference to a task that no longer exists is kept, marked missing.
	ghost := "---\nsubject: Ghost dep\nstatus: pending\npriority: 2\ndepends_on:\n  - gone\n---\n"
	if err := os.WriteFile(filepath.Join(registry, "task", "p", "orphan.md"), []byte(ghost), 0o644); err != nil {
		t.Fatal(err)
	}
	ref := listedTask(t, listJSON(t), "orphan")["dependencies"].([]interface{})[0].(map[string]interface{})
	if ref["name"] != "gone" || ref["missing"] != true || ref["id"] != nil {
		t.Errorf("missing dependency = %v, want {name:gone missing:true}", ref)
	}
}

func TestTaskListJSON_TreeAddsEdges(t *testing.T) {
	setup(t)
	chdir(t)
	seedDependencyTasks(t)
	run(t, "task", "status", "setup-db", "completed", "--project", "p")

	payload := listJSON(t, "--tree")
	if got := keysOf(payload); got != "counts,edges,project,tasks" {
		t.Errorf("top-level keys = %s, want counts,edges,project,tasks", got)
	}
	// setup-db is completed and hidden without --all, but its edge stays.
	if len(payload["tasks"].([]interface{})) != 2 {
		t.Errorf("tasks = %v, want the 2 open tasks", payload["tasks"])
	}
	edges := payload["edges"].([]interface{})
	if len(edges) != 2 {
		t.Fatalf("edges = %v, want 2", edges)
	}
	apiID := listedTask(t, payload, "create-api")["id"]
	want := map[string]bool{"setup-db": true, "merge-pr": false}
	for _, raw := range edges {
		edge := raw.(map[string]interface{})
		to := edge["to"].(string)
		satisfied, known := want[to]
		if !known || edge["from"] != "create-api" || edge["from_id"] != apiID || edge["to_id"] == "" {
			t.Errorf("unexpected edge %v", edge)
			continue
		}
		if edge["satisfied"] != satisfied {
			t.Errorf("edge to %s satisfied = %v, want %v", to, edge["satisfied"], satisfied)
		}
	}

	// No tasks: edges is still an empty list, not missing.
	empty := listJSON(t, "--tree", "--feature", "none")
	if edges, ok := empty["edges"].([]interface{}); !ok || len(edges) != 0 {
		t.Errorf("edges for empty list = %v, want []", empty["edges"])
	}
}
