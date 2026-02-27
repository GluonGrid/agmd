package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList_ShowsItems(t *testing.T) {
	setup(t)

	run(t, "new", "rule:list-rule", "--content", "# List Rule")

	out := run(t, "list")
	if !strings.Contains(out, "list-rule") {
		t.Error("list should show list-rule")
	}
}

func TestShow_DisplaysContent(t *testing.T) {
	setup(t)

	run(t, "new", "rule:show-rule", "--content", "# Show Rule\nExpected content.")

	out := run(t, "show", "rule:show-rule")
	if !strings.Contains(out, "Expected content.") {
		t.Error("show should display item content")
	}
}

func TestDelete_RemovesItem(t *testing.T) {
	registry := setup(t)

	run(t, "new", "rule:del-rule", "--content", "# Del Rule")
	run(t, "delete", "rule:del-rule", "--force")

	if fileExists(filepath.Join(registry, "rule", "del-rule.md")) {
		t.Error("rule/del-rule.md should be removed after delete")
	}
}

func TestShow_Raw(t *testing.T) {
	setup(t)

	run(t, "new", "rule:raw-rule", "--content", "# Raw Rule\nContent here.")

	out := run(t, "show", "rule:raw-rule", "--raw")
	if !strings.Contains(out, "---") {
		t.Error("show --raw should include frontmatter delimiters")
	}
	if !strings.Contains(out, "Content here.") {
		t.Error("show --raw should include content")
	}
}

func TestEdit_UpdatesContent(t *testing.T) {
	registry := setup(t)

	run(t, "new", "rule:edit-rule", "--content", "# Original")
	run(t, "edit", "rule:edit-rule", "--content", "# Updated content")

	content := readFile(t, filepath.Join(registry, "rule", "edit-rule.md"))
	if !strings.Contains(content, "Updated content") {
		t.Error("edit should update the item content")
	}
}

func TestMv_MovesItem(t *testing.T) {
	registry := setup(t)

	run(t, "new", "rule:old-name", "--content", "# Rule")
	run(t, "mv", "rule:old-name", "rule:new-name")

	if fileExists(filepath.Join(registry, "rule", "old-name.md")) {
		t.Error("old-name.md should no longer exist")
	}
	if !fileExists(filepath.Join(registry, "rule", "new-name.md")) {
		t.Error("new-name.md should exist after mv")
	}
}

func TestPromote_PromotesNewBlock(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	// Write a directives.md with an inline :::new block
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(
		"# Project\n\n:::new rule:inline-rule\n# Inline Rule\nDo this.\n:::end\n",
	), 0644)

	run(t, "promote", "--all")

	if !fileExists(filepath.Join(registry, "rule", "inline-rule.md")) {
		t.Error("promote should create rule/inline-rule.md in registry")
	}
}

func TestNew_CreatesItem(t *testing.T) {
	registry := setup(t)

	run(t, "new", "rule:my-rule", "--content", "# My Rule\nDo this.")

	if !fileExists(filepath.Join(registry, "rule", "my-rule.md")) {
		t.Error("rule/my-rule.md should exist in registry")
	}
}

func TestNew_Local(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// Initialize a local registry first
	run(t, "init", "--local")
	run(t, "new", "rule:local-rule", "--local", "--content", "# Local Rule")

	// Should be in project-local .agmd/, not the global registry
	if !fileExists(filepath.Join(dir, ".agmd", "rule", "local-rule.md")) {
		t.Error("rule should be in project-local .agmd/rule/local-rule.md")
	}
}

func TestMv_ToLocal(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "init", "--local")
	run(t, "new", "rule:global-rule", "--content", "# Global Rule")
	run(t, "mv", "rule:global-rule", "--to-local")

	if !fileExists(filepath.Join(dir, ".agmd", "rule", "global-rule.md")) {
		t.Error("rule should be in .agmd/rule/ after --to-local")
	}
}

func TestMv_ToGlobal(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	run(t, "init", "--local")
	run(t, "new", "rule:local-rule", "--local", "--content", "# Local Rule")
	run(t, "mv", "rule:local-rule", "--to-global")

	if !fileExists(filepath.Join(registry, "rule", "local-rule.md")) {
		t.Error("rule should be in global registry after --to-global")
	}
	if fileExists(filepath.Join(dir, ".agmd", "rule", "local-rule.md")) {
		t.Error("rule should no longer be in local .agmd/ after --to-global")
	}
}
