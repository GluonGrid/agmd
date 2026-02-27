package integration

import (
	"strings"
	"testing"
)

func TestTask_List(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "task-alpha", "--content", "Alpha.")
	run(t, "task", "new", "task-beta", "--content", "Beta.")

	out := run(t, "task", "list")
	if !strings.Contains(out, "task-alpha") {
		t.Error("task list should show task-alpha")
	}
	if !strings.Contains(out, "task-beta") {
		t.Error("task list should show task-beta")
	}
}

func TestTask_Status(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "my-task", "--content", "Do something.")
	run(t, "task", "status", "my-task", "completed")

	out := run(t, "task", "list", "--all")
	if !strings.Contains(out, "completed") {
		t.Error("task should show as completed")
	}
}

func TestTask_Show(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "show-task", "--content", "Show this content.")

	out := run(t, "task", "show", "show-task")
	if !strings.Contains(out, "Show this content.") {
		t.Error("task show should display task content")
	}
}

func TestTask_New(t *testing.T) {
	setup(t)
	chdir(t)

	out := run(t, "task", "new", "my-task", "--content", "Do the thing.")
	if !strings.Contains(out, "my-task") {
		t.Error("expected task name in output")
	}
}

func TestTask_NewWithTypeAndPriority(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "critical-bug", "-t", "bug", "-p", "0", "--content", "Critical bug.")

	out := run(t, "task", "list")
	if !strings.Contains(out, "critical-bug") {
		t.Error("task list should show critical-bug")
	}
	if !strings.Contains(out, "P0") {
		t.Error("task list should show P0 priority")
	}
}

func TestTask_Delete(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "del-task", "--content", "To be deleted.")
	run(t, "task", "delete", "del-task", "--force")

	out := run(t, "task", "list")
	if strings.Contains(out, "del-task") {
		t.Error("deleted task should not appear in list")
	}
}

func TestTask_Clean(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "done-task", "--content", "Completed.")
	run(t, "task", "status", "done-task", "completed")
	run(t, "task", "new", "active-task", "--content", "Still active.")

	run(t, "task", "clean", "--force")

	out := run(t, "task", "list", "--all")
	if strings.Contains(out, "done-task") {
		t.Error("clean should remove completed tasks")
	}
	if !strings.Contains(out, "active-task") {
		t.Error("clean should keep active tasks")
	}
}

func TestTask_Edit(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "my-task", "--content", "Do this.")
	run(t, "task", "edit", "my-task", "-p", "0", "-t", "bug")

	out := run(t, "task", "list")
	if !strings.Contains(out, "P0") {
		t.Error("task list should show P0 after edit")
	}
	if !strings.Contains(out, "bug") {
		t.Error("task list should show bug type after edit")
	}
}

func TestTask_ListStatusFilter(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "ready-task", "--content", "Ready.")
	run(t, "task", "new", "done-task", "--content", "Done.")
	run(t, "task", "status", "done-task", "completed")

	outReady := run(t, "task", "list", "--status", "ready")
	if !strings.Contains(outReady, "ready-task") {
		t.Error("--status ready should show ready-task")
	}
	if strings.Contains(outReady, "done-task") {
		t.Error("--status ready should not show done-task")
	}
}

func TestTask_BlockedByAndUnblock(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "dep-task", "--content", "Dependency.")
	run(t, "task", "new", "main-task", "--content", "Needs dep.", "--blocked-by", "dep-task")

	out := run(t, "task", "list")
	if !strings.Contains(out, "blocked") {
		t.Errorf("main-task should show as blocked, got:\n%s", out)
	}

	run(t, "task", "unblock", "main-task", "dep-task")

	out = run(t, "task", "list")
	if strings.Contains(out, "blocked") {
		t.Errorf("main-task should not be blocked after unblock, got:\n%s", out)
	}
}

func TestTask_ListFilters(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "bug-one", "-t", "bug", "-p", "0", "--content", "Bug.")
	run(t, "task", "new", "feat-one", "-t", "feature", "-p", "2", "--content", "Feature.")

	outBug := run(t, "task", "list", "--type", "bug")
	if !strings.Contains(outBug, "bug-one") {
		t.Error("--type bug should show bug-one")
	}
	if strings.Contains(outBug, "feat-one") {
		t.Error("--type bug should not show feat-one")
	}

	outP0 := run(t, "task", "list", "--priority", "0")
	if !strings.Contains(outP0, "bug-one") {
		t.Error("--priority 0 should show bug-one")
	}
	if strings.Contains(outP0, "feat-one") {
		t.Error("--priority 0 should not show feat-one (P2)")
	}
}

func TestTask_ShowAll(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "task", "new", "show-a", "--content", "Content A.")
	run(t, "task", "new", "show-b", "--content", "Content B.")

	out := run(t, "task", "show", "--all")
	if !strings.Contains(out, "Content A.") {
		t.Error("show --all should include Content A.")
	}
	if !strings.Contains(out, "Content B.") {
		t.Error("show --all should include Content B.")
	}
}
