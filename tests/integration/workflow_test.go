package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkflow_NewProject simulates a user setting up agmd in a fresh project:
// setup → init → new rule → add to directives.md → sync → verify AGENTS.md
func TestWorkflow_NewProject(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// Create a rule in the registry
	run(t, "new", "rule:typescript", "--content", "# TypeScript\nUse strict mode.")

	// Init project and add the rule to directives.md
	run(t, "init")
	directives := readFile(t, filepath.Join(dir, "directives.md"))
	directives += "\n:::use rule:typescript\n"
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(directives), 0644)

	// Sync and verify
	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "Use strict mode.") {
		t.Error("AGENTS.md should contain the typescript rule content")
	}
}

// TestWorkflow_Migration simulates migrating an existing CLAUDE.md into agmd:
// write CLAUDE.md → migrate → sync → AGENTS.md has the content
func TestWorkflow_Migration(t *testing.T) {
	setup(t)
	dir := chdir(t)
	t.Setenv("EDITOR", "true") // prevent editor opening

	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Instructions\n\nAlways write tests.\n"), 0644)

	run(t, "migrate", "CLAUDE.md")

	if !fileExists(filepath.Join(dir, "directives.md")) {
		t.Fatal("migrate should create directives.md")
	}
	if !fileExists(filepath.Join(dir, "CLAUDE.md.backup")) {
		t.Error("migrate should create CLAUDE.md.backup")
	}

	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "Always write tests.") {
		t.Error("AGENTS.md should contain migrated content after sync")
	}
}

// TestWorkflow_DocLinkAndSync simulates adding documentation and verifying it
// appears in AGENTS.md after sync:
// doc add → doc link → sync → AGENTS.md references the doc
func TestWorkflow_DocLinkAndSync(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// Create a docs folder
	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "api.md"), []byte("# API Reference"), 0644)

	// init must come before doc link so directives.md has :::docs block
	run(t, "init")
	run(t, "doc", "add", docsDir, "api-docs")
	run(t, "doc", "link", "api-docs")
	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "api-docs") {
		t.Error("AGENTS.md should reference api-docs after doc link + sync")
	}
}

// TestWorkflow_CollectAndPromote simulates extracting content from AGENTS.md
// into the registry, then promoting an inline :::new block:
// write source files → collect → promote → registry has items + directives updated
func TestWorkflow_CollectAndPromote(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	// Set up source files for collect
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

## Rules

:::list
rule:go-style
:::end
`), 0644)

	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(`# Agent Instructions

## Rules

### go-style

# Rule: Go Style

Always use gofmt.
`), 0644)

	// Collect extracts sections from AGENTS.md into registry
	run(t, "collect", "--all")

	if !fileExists(filepath.Join(registry, "rule", "go-style.md")) {
		t.Fatal("collect should store rule:go-style in registry")
	}

	// Now promote an inline :::new block
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

:::new rule:error-handling
# Rule: Error Handling

Always handle errors explicitly.
:::end
`), 0644)

	run(t, "promote", "--all")

	if !fileExists(filepath.Join(registry, "rule", "error-handling.md")) {
		t.Fatal("promote should store rule:error-handling in registry")
	}

	// directives.md should now use :::use instead of :::new
	updatedDirectives := readFile(t, filepath.Join(dir, "directives.md"))
	if strings.Contains(updatedDirectives, ":::new") {
		t.Error("promote should replace :::new block with :::use directive")
	}
	if !strings.Contains(updatedDirectives, "error-handling") {
		t.Error("promote should leave a reference to the promoted item")
	}
}

// TestWorkflow_SkillLinkAndSync simulates linking a skill and verifying it
// appears in AGENTS.md:
// link skill → init (with :::skills) → sync → AGENTS.md mentions skill
func TestWorkflow_SkillLinkAndSync(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// Write directives.md with empty :::skills block first
	// skill link will auto-update it to add the skill name
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\n## Skills\n\n:::skills\n:::end\n"), 0644)

	run(t, "skill", "link", "managing-agmd", "--target", "agents")
	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "managing-agmd") {
		t.Error("AGENTS.md should mention the linked managing-agmd skill after sync")
	}
}
