package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetup_CreatesRegistry(t *testing.T) {
	registry := filepath.Join(t.TempDir(), "registry")
	t.Setenv("AGMD_HOME", registry)

	run(t, "setup")

	for _, path := range []string{
		filepath.Join(registry, "profile", "default.md"),
		filepath.Join(registry, "guide", "agmd.md"),
		filepath.Join(registry, "skill", "managing-agmd", "SKILL.md"),
	} {
		if !fileExists(path) {
			t.Errorf("missing: %s", path)
		}
	}
}

func TestSetup_AlreadyExists(t *testing.T) {
	setup(t) // first setup

	out := run(t, "setup") // second setup — should warn
	if !strings.Contains(out, "already exists") {
		t.Error("expected 'already exists' message")
	}
}

func TestSetup_Force(t *testing.T) {
	setup(t)

	// Delete a file to simulate corruption
	registry := os.Getenv("AGMD_HOME")
	os.Remove(filepath.Join(registry, "guide", "agmd.md"))

	run(t, "setup", "--force")

	if !fileExists(filepath.Join(registry, "guide", "agmd.md")) {
		t.Error("--force should recreate guide/agmd.md")
	}
}

func TestSetup_SkillHasRequiredFiles(t *testing.T) {
	registry := filepath.Join(t.TempDir(), "registry")
	t.Setenv("AGMD_HOME", registry)

	run(t, "setup")

	skillDir := filepath.Join(registry, "skill", "managing-agmd")
	for _, name := range []string{
		"SKILL.md",
		"references/migration.md",
		"references/tasks.md",
		"references/directives.md",
		"scripts/check-setup.sh",
	} {
		if !fileExists(filepath.Join(skillDir, name)) {
			t.Errorf("missing skill file: %s", name)
		}
	}
}
