package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSkill_ListShowsBundled(t *testing.T) {
	setup(t)

	out := run(t, "skill", "list")
	if !strings.Contains(out, "managing-agmd") {
		t.Errorf("skill list should show bundled managing-agmd skill, got:\n%s", out)
	}
}

func TestSkill_Show(t *testing.T) {
	setup(t)

	out := run(t, "skill", "show", "managing-agmd")
	if !strings.Contains(out, "managing-agmd") {
		t.Errorf("skill show should display SKILL.md content, got:\n%s", out)
	}
}

func TestSkill_LinkAndUnlink(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "skill", "link", "managing-agmd", "--target", "agents")

	if !fileExists(filepath.Join(dir, ".agents", "skills", "managing-agmd")) {
		t.Error("skill should be linked into .agents/skills/managing-agmd")
	}

	run(t, "skill", "unlink", "managing-agmd")

	if fileExists(filepath.Join(dir, ".agents", "skills", "managing-agmd")) {
		t.Error("skill symlink should be removed after unlink")
	}
}

func TestSkill_LinkToClaude(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "skill", "link", "managing-agmd", "--target", "claude")

	if !fileExists(filepath.Join(dir, ".claude", "skills", "managing-agmd")) {
		t.Error("skill should be linked into .claude/skills/managing-agmd")
	}
}

func TestSkill_LinkToBoth(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "skill", "link", "managing-agmd", "--target", "both")

	if !fileExists(filepath.Join(dir, ".claude", "skills", "managing-agmd")) {
		t.Error("skill should be linked into .claude/skills/ with --target both")
	}
	if !fileExists(filepath.Join(dir, ".agents", "skills", "managing-agmd")) {
		t.Error("skill should be linked into .agents/skills/ with --target both")
	}
}

func TestSkill_LinkAutoDetect(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// No .claude/ or .agents/ exists — should auto-create .agents/skills/
	run(t, "skill", "link", "managing-agmd")

	if !fileExists(filepath.Join(dir, ".agents", "skills", "managing-agmd")) {
		t.Error("skill should auto-link into .agents/skills/ when no agent dir exists")
	}
}

func TestSkill_Delete(t *testing.T) {
	setup(t)

	// The bundled managing-agmd skill is always present; delete it and verify
	run(t, "skill", "delete", "managing-agmd", "--force")

	out := run(t, "skill", "list")
	if strings.Contains(out, "managing-agmd") {
		t.Errorf("deleted skill should not appear in list, got:\n%s", out)
	}
}
