package integration

import (
	"strings"
	"testing"
)

// 'agmd task --help' is the one-shot reference agents read once per session.
func TestTaskHelp_AgentReferenceSections(t *testing.T) {
	setup(t)
	chdir(t)

	out := run(t, "task", "--help")
	for _, want := range []string{
		// command matrix
		"Commands:", "read ", "create ", "change ", "dependencies ", "lifecycle ", "human tasks ", "checks ",
		// patterns
		"update <t>", "dep add <t> <dep>", "dep remove <t> <dep>",
		"next --claim --workflow-id W", "heartbeat", "--human", "--check", "check",
		"reopen <t>", "replace <old> --by <new>", "move <t> --to-project", "--delete N", "lint [--all]",
		// JSON and exit codes
		"--json", "Exit codes:", "2 invalid_input", "3 not_found", "6 conflict",
		"Examples:", "list --json",
		// aliases listed once, at the end
		"Aliases (compatibility",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("task --help is missing %q", want)
		}
	}
	if strings.Contains(out, "Available Commands:") {
		t.Error("task --help should not repeat cobra's command list")
	}
	aliases := strings.Index(out, "Aliases (compatibility")
	if aliases < strings.Index(out, "Examples:") {
		t.Error("aliases should come after the examples, at the end")
	}
	for _, alias := range []string{"edit = update", "blocked-by = dep add", "unblock = dep remove"} {
		if strings.Count(out, alias) != 1 {
			t.Errorf("alias %q should be listed exactly once", alias)
		}
	}
	if lines := strings.Count(out, "\n"); lines > 60 {
		t.Errorf("task --help has %d lines; keep it to one screen", lines)
	}
}

func TestTaskHelp_SubcommandHelpUnchanged(t *testing.T) {
	setup(t)
	chdir(t)

	out := run(t, "task", "dep", "--help")
	if !strings.Contains(out, "Available Commands:") || !strings.Contains(out, "add") {
		t.Errorf("subcommand help should keep cobra's default layout, got:\n%s", out)
	}
}

func TestRootHelp_PointsAgentsToTaskHelp(t *testing.T) {
	setup(t)
	chdir(t)

	out := run(t, "--help")
	if !strings.Contains(out, "agmd task --help") || !strings.Contains(out, "once per session") {
		t.Errorf("agmd --help should point agents to 'agmd task --help', got:\n%s", out)
	}
}
