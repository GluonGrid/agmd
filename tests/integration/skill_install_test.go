//go:build integration_network

package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestSkill_InstallFromGitHub verifies that a real skill can be installed from GitHub.
// This test requires network access and is excluded from the default test run.
//
// Run with: go test ./tests/integration/ -tags integration_network -run TestSkill_Install
//
func TestSkill_InstallFromGitHub(t *testing.T) {
	setup(t)

	out := run(t, "skill", "install", "millionco/react-doctor/skills/react-doctor")
	if !strings.Contains(out, "react-doctor") {
		t.Errorf("skill install should confirm react-doctor was installed, got:\n%s", out)
	}

	listOut := run(t, "skill", "list")
	if !strings.Contains(listOut, "react-doctor") {
		t.Errorf("skill list should show react-doctor after install, got:\n%s", listOut)
	}
}

func TestSkill_InstallAndLink(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "skill", "install", "millionco/react-doctor/skills/react-doctor")
	run(t, "skill", "link", "react-doctor")

	// Should be linked into .agents/skills/ or .claude/skills/ (auto-detect)
	agentsLink := filepath.Join(dir, ".agents", "skills", "react-doctor")
	claudeLink := filepath.Join(dir, ".claude", "skills", "react-doctor")
	if !fileExists(agentsLink) && !fileExists(claudeLink) {
		t.Error("react-doctor should be linked into .agents/skills/ or .claude/skills/")
	}
}
