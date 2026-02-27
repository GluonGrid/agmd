package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func initAndSync(t *testing.T) {
	t.Helper()
	run(t, "init")
	run(t, "sync")
}

func TestSymlink_Remove(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--claude")
	run(t, "symlink", "remove", "CLAUDE.md")

	if fileExists(filepath.Join(dir, "CLAUDE.md")) {
		t.Error("CLAUDE.md symlink should be removed")
	}
}

func TestSymlink_Claude(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--claude")

	if !fileExists(filepath.Join(dir, "CLAUDE.md")) {
		t.Error("CLAUDE.md symlink should exist after symlink add --claude")
	}
}

func TestSymlink_Cursor(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--cursor")

	if !fileExists(filepath.Join(dir, ".cursorrules")) {
		t.Error(".cursorrules symlink should exist after symlink add --cursor")
	}
}

func TestSymlink_Windsurf(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--windsurf")

	if !fileExists(filepath.Join(dir, ".windsurfrules")) {
		t.Error(".windsurfrules symlink should exist after symlink add --windsurf")
	}
}

func TestSymlink_Aider(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--aider")

	if !fileExists(filepath.Join(dir, ".aider.conf.yml")) {
		t.Error(".aider.conf.yml symlink should exist after symlink add --aider")
	}
}

func TestSymlink_Copilot(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--copilot")

	if !fileExists(filepath.Join(dir, ".github", "copilot-instructions.md")) {
		t.Error(".github/copilot-instructions.md symlink should exist after symlink add --copilot")
	}
}

func TestSymlink_All(t *testing.T) {
	setup(t)
	dir := chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--all")

	for _, path := range []string{
		"CLAUDE.md",
		".cursorrules",
		".windsurfrules",
		".aider.conf.yml",
		filepath.Join(".github", "copilot-instructions.md"),
	} {
		if !fileExists(filepath.Join(dir, path)) {
			t.Errorf("symlink %s should exist after symlink add --all", path)
		}
	}
}

func TestSymlink_List(t *testing.T) {
	setup(t)
	chdir(t)

	initAndSync(t)
	run(t, "symlink", "add", "--claude")

	out := run(t, "symlink", "list")
	if !strings.Contains(out, "CLAUDE.md") {
		t.Errorf("symlink list should show CLAUDE.md, got:\n%s", out)
	}
}
