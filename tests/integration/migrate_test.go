package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrate_CreatesDirectivesAndBackup(t *testing.T) {
	setup(t)
	dir := chdir(t)
	t.Setenv("EDITOR", "true") // prevent real editor from opening

	// Create a source file to migrate
	src := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(src, []byte("# My Rules\nDo the thing."), 0644)

	out := run(t, "migrate", "CLAUDE.md")

	if !strings.Contains(out, "Migration complete") {
		t.Errorf("expected 'Migration complete' in output, got:\n%s", out)
	}

	// backup should exist
	if !fileExists(filepath.Join(dir, "CLAUDE.md.backup")) {
		t.Error("backup CLAUDE.md.backup should exist")
	}

	// directives.md should be created
	if !fileExists(filepath.Join(dir, "directives.md")) {
		t.Error("directives.md should be created by migrate")
	}

	// directives.md should contain the original content
	directives := readFile(t, filepath.Join(dir, "directives.md"))
	if !strings.Contains(directives, "Do the thing.") {
		t.Error("directives.md should contain migrated content")
	}
}

func TestMigrate_FailsIfDirectivesExists(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Rules"), 0644)
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Existing"), 0644)

	_, err := runE(t, "migrate", "CLAUDE.md")
	if err == nil {
		t.Error("expected error when directives.md already exists")
	}
}

func TestMigrate_ForceOverwritesDirectives(t *testing.T) {
	setup(t)
	dir := chdir(t)
	t.Setenv("EDITOR", "true")

	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Rules\nNew content."), 0644)
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Old"), 0644)

	out := run(t, "migrate", "CLAUDE.md", "--force")

	if !strings.Contains(out, "Migration complete") {
		t.Errorf("expected 'Migration complete' in output, got:\n%s", out)
	}

	directives := readFile(t, filepath.Join(dir, "directives.md"))
	if !strings.Contains(directives, "New content.") {
		t.Error("directives.md should contain new content after force migrate")
	}
}
