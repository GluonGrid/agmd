package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSync_MissingDirectives(t *testing.T) {
	setup(t)
	chdir(t) // empty dir, no directives.md

	out, err := runE(t, "sync")
	if err == nil && !strings.Contains(out, "directives.md") {
		t.Error("expected error mentioning directives.md")
	}
}

func TestSync_ListBlock(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "new", "rule:rule-a", "--content", "# Rule A\nDo A.")
	run(t, "new", "rule:rule-b", "--content", "# Rule B\nDo B.")

	os.WriteFile(filepath.Join(dir, "directives.md"),
		[]byte("# Project\n:::list\nrule:rule-a\nrule:rule-b\n:::end\n"), 0644)

	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "Rule A") {
		t.Error("AGENTS.md should contain Rule A")
	}
	if !strings.Contains(agents, "Rule B") {
		t.Error("AGENTS.md should contain Rule B")
	}
}

func TestSync_SkillsBlock(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// :::skills with no names → should render "No skills linked" message
	os.WriteFile(filepath.Join(dir, "directives.md"),
		[]byte("# Project\n## Skills available\n:::skills\n:::end\n"), 0644)

	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "No skills linked") {
		t.Errorf("AGENTS.md should contain 'No skills linked' message, got:\n%s", agents)
	}
}

func TestSync_DocsBlock(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "init")

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")
	run(t, "doc", "link", "my-docs")
	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "my-docs") {
		t.Error("AGENTS.md should reference my-docs after doc link + sync")
	}
}

func TestSync_UseDirective(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "new", "rule:my-rule", "--content", "# My Rule\nFollow this.")

	os.WriteFile(filepath.Join(dir, "directives.md"),
		[]byte("# Project\n:::use rule:my-rule\n"), 0644)

	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "Follow this.") {
		t.Error("AGENTS.md should contain :::use rule content")
	}
}

func TestSync_LocalOverlay(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "new", "rule:shared-rule", "--content", "# Shared\nShared content.")
	run(t, "new", "rule:local-rule", "--content", "# Local\nLocal override.")

	os.WriteFile(filepath.Join(dir, "directives.md"),
		[]byte("# Project\n:::use rule:shared-rule\n"), 0644)
	os.WriteFile(filepath.Join(dir, "directives.local.md"),
		[]byte(":::use rule:local-rule\n"), 0644)

	run(t, "sync")

	agents := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(agents, "Shared content.") {
		t.Error("AGENTS.md should contain shared rule content")
	}
	if !strings.Contains(agents, "Local override.") {
		t.Error("AGENTS.md should contain local overlay content")
	}
}

func TestSync_BasicUse(t *testing.T) {
	setup(t)
	dir := chdir(t)

	run(t, "init")
	out := run(t, "sync")

	if !strings.Contains(out, "Generated AGENTS.md") {
		t.Error("expected sync success message")
	}
	if !fileExists(filepath.Join(dir, "AGENTS.md")) {
		t.Error("AGENTS.md should exist")
	}
}
