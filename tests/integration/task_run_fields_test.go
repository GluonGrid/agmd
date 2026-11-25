package integration

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runFieldsIn(t *testing.T, project, name string) map[string]string {
	t.Helper()
	out, err := runE(t, "task", "show", name, "--project", project, "--json")
	if err != nil {
		t.Fatalf("show %s/%s: %v\n%s", project, name, err, out)
	}
	var item struct {
		Run map[string]string `json:"run"`
	}
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("show output is not JSON: %v\n%s", err, out)
	}
	return item.Run
}

func TestTaskRunFields_SetOnNewAndUpdate(t *testing.T) {
	registry := setup(t)
	chdir(t)
	if out, err := runE(t, "task", "new", "ui", "--project", "p", "--content", "x",
		"--set", "run.profile=frontend", "--set", "run.model=opus"); err != nil {
		t.Fatalf("new: %v\n%s", err, out)
	}
	if got := runFieldsIn(t, "p", "ui"); !reflect.DeepEqual(got, map[string]string{"profile": "frontend", "model": "opus"}) {
		t.Fatalf("run after new = %v", got)
	}
	file := readFile(t, filepath.Join(registry, "task", "p", "ui.md"))
	if !strings.Contains(file, "run:\n    model: opus\n    profile: frontend\n") {
		t.Errorf("front matter lacks run map:\n%s", file)
	}

	// Set and unset in one step; an empty value clears too.
	if out, err := runE(t, "task", "update", "ui", "--project", "p",
		"--set", "run.effort=high", "--set", "run.skills=impeccable,wt", "--unset", "run.model", "--set", "run.profile="); err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if got := runFieldsIn(t, "p", "ui"); !reflect.DeepEqual(got, map[string]string{"effort": "high", "skills": "impeccable,wt"}) {
		t.Fatalf("run after update = %v", got)
	}
	// The flag values do not leak into the next command.
	run(t, "task", "update", "ui", "--project", "p", "--subject", "UI")
	if got := runFieldsIn(t, "p", "ui"); len(got) != 2 {
		t.Errorf("run after subject update = %v", got)
	}

	// list --json carries it as well.
	out, err := runE(t, "task", "list", "--project", "p", "--json")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var list struct {
		Tasks []struct {
			Name string            `json:"name"`
			Run  map[string]string `json:"run"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		t.Fatalf("list output is not JSON: %v\n%s", err, out)
	}
	if len(list.Tasks) != 1 || list.Tasks[0].Run["effort"] != "high" {
		t.Errorf("list = %+v", list.Tasks)
	}

	// Clearing every field removes the map from the file and the JSON.
	run(t, "task", "update", "ui", "--project", "p", "--unset", "run.effort", "--unset", "run.skills")
	if got := runFieldsIn(t, "p", "ui"); got != nil {
		t.Errorf("run after clearing = %v", got)
	}
	if strings.Contains(readFile(t, filepath.Join(registry, "task", "p", "ui.md")), "run:") {
		t.Error("empty run map still written")
	}
	if out := run(t, "task", "show", "ui", "--project", "p"); strings.Contains(out, "run:") {
		t.Errorf("show prints empty run:\n%s", out)
	}
	run(t, "task", "update", "ui", "--project", "p", "--set", "run.repo=agmd")
	if out := run(t, "task", "show", "ui", "--project", "p"); !strings.Contains(out, "run: repo=agmd") {
		t.Errorf("show lacks run line:\n%s", out)
	}
}

func TestTaskRunFields_Refusals(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "x")
	for _, args := range [][]string{
		{"--set", "profile=frontend"}, // no run. prefix
		{"--set", "run.profile"},      // no value
		{"--set", "run.Bad=x"},        // key case
		{"--set", "run.=x"},           // empty key
		{"--unset", "model"},          // no run. prefix
		{"--set", "run.a=1", "--unset", "run.a"},
	} {
		full := append([]string{"task", "update", "a", "--project", "p"}, args...)
		if out, err := runE(t, full...); err == nil {
			t.Errorf("%v accepted:\n%s", args, out)
		}
	}
	if out, err := runE(t, "task", "new", "b", "--project", "p", "--content", "x", "--set", "bad"); err == nil {
		t.Errorf("new with bad --set accepted:\n%s", out)
	}
	if _, err := runE(t, "task", "show", "b", "--project", "p"); err == nil {
		t.Error("task created despite a bad --set")
	}
}

func TestTaskRunFields_KeptAcrossMoveReplaceReopen(t *testing.T) {
	setup(t)
	chdir(t)
	want := map[string]string{"profile": "frontend", "effort": "high"}
	run(t, "task", "new", "old", "--project", "p", "--content", "x", "--set", "run.profile=frontend", "--set", "run.effort=high")
	run(t, "task", "new", "new", "--project", "p", "--content", "y", "--set", "run.model=sonnet")

	if out, err := runE(t, "task", "replace", "old", "--by", "new", "--project", "p"); err != nil {
		t.Fatalf("replace: %v\n%s", err, out)
	}
	if got := runFieldsIn(t, "p", "old"); !reflect.DeepEqual(got, want) {
		t.Errorf("old after replace = %v", got)
	}
	if got := runFieldsIn(t, "p", "new"); !reflect.DeepEqual(got, map[string]string{"model": "sonnet"}) {
		t.Errorf("new after replace = %v", got)
	}

	if out, err := runE(t, "task", "reopen", "old", "--project", "p"); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	if got := runFieldsIn(t, "p", "old"); !reflect.DeepEqual(got, want) {
		t.Errorf("old after reopen = %v", got)
	}

	if out, err := runE(t, "task", "move", "old", "--project", "p", "--to-project", "q"); err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	if got := runFieldsIn(t, "q", "old"); !reflect.DeepEqual(got, want) {
		t.Errorf("old after move = %v", got)
	}
}
