package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_CreatesDirectivesMd(t *testing.T) {
	setup(t)
	dir := chdir(t)

	out := run(t, "init")

	if !strings.Contains(out, "directives.md") {
		t.Errorf("expected 'directives.md' in output, got:\n%s", out)
	}
	if !fileExists(filepath.Join(dir, "directives.md")) {
		t.Error("directives.md should be created by init")
	}
}

func TestInit_FailsIfDirectivesExists(t *testing.T) {
	setup(t)
	chdir(t)

	run(t, "init")
	_, err := runE(t, "init")
	if err == nil {
		t.Error("expected error when directives.md already exists")
	}
}

func TestInit_WithDefaultProfile(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "init", "profile:default")

	if !fileExists(filepath.Join(dir, "directives.md")) {
		t.Error("directives.md should be created with default profile")
	}
}

func TestInit_Local(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "init", "--local")

	if !fileExists(filepath.Join(dir, "directives.md")) {
		t.Error("directives.md should be created with --local")
	}
	if !fileExists(filepath.Join(dir, ".agmd")) {
		t.Error(".agmd/ local registry should be created with --local")
	}
}
