package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollect_ItemsFromAgentsMd(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	// Create a directives.md with :::list blocks (type:name format inside :::list)
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

## Rules

:::list
rule:typescript
:::end
`), 0644)

	// Create the matching AGENTS.md with the rule content
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(`# Agent Instructions

## Rules

### typescript

# Rule: TypeScript Best Practices

Use strict mode and type safety.
`), 0644)

	out := run(t, "collect", "--all")

	if !strings.Contains(out, "typescript") {
		t.Errorf("expected 'typescript' in collect output, got:\n%s", out)
	}

	// Item should be in registry
	if !fileExists(filepath.Join(registry, "rule", "typescript.md")) {
		t.Error("rule:typescript should be stored in registry after collect")
	}
}

// helpers shared by collect tests
var collectDirectives = `# Project

## Rules

:::list
rule:typescript
:::end
`

var collectAgents = `# Agent Instructions

## Rules

### typescript

# Rule: TypeScript Best Practices

Use strict mode and type safety.
`

func TestCollect_Overwrite(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(collectDirectives), 0644)
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(collectAgents), 0644)

	// First collect — creates the item
	run(t, "collect", "--all")

	// Overwrite the registry item with different content so we can detect the change
	os.WriteFile(filepath.Join(registry, "rule", "typescript.md"), []byte("old content"), 0644)

	// Collect again with --overwrite — should replace
	out := run(t, "collect", "--overwrite")
	if !strings.Contains(out, "Overwritten") {
		t.Errorf("expected 'Overwritten' in output, got:\n%s", out)
	}

	stored := readFile(t, filepath.Join(registry, "rule", "typescript.md"))
	if strings.Contains(stored, "old content") {
		t.Error("--overwrite should replace existing registry item")
	}
}

func TestCollect_FromAlternativeFile(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	// directives.md points at CLAUDE.md sections
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(collectDirectives), 0644)

	// Use CLAUDE.md as the source instead of AGENTS.md
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(collectAgents), 0644)

	run(t, "collect", "--all", "--file", "CLAUDE.md")

	if !fileExists(filepath.Join(registry, "rule", "typescript.md")) {
		t.Error("rule:typescript should be collected from CLAUDE.md")
	}
}

func TestCollect_MissingDirectives(t *testing.T) {
	setup(t)
	chdir(t)

	_, err := runE(t, "collect", "--all")
	if err == nil {
		t.Error("expected error when directives.md is missing")
	}
}

func TestCollect_MissingAgentsMd(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n"), 0644)

	_, err := runE(t, "collect", "--all")
	if err == nil {
		t.Error("expected error when AGENTS.md is missing")
	}
}

func TestCollect_MultipleTypes(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

## Rules

:::list
rule:typescript
rule:eslint
:::end

## Workflows

:::list
workflow:deploy
:::end
`), 0644)

	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(`# Agent Instructions

## Rules

### typescript

Use strict mode.

### eslint

Enable recommended rules.

## Workflows

### deploy

1. Run tests
2. Build
3. Deploy
`), 0644)

	run(t, "collect", "--all")

	for _, item := range []struct{ typ, name string }{
		{"rule", "typescript"},
		{"rule", "eslint"},
		{"workflow", "deploy"},
	} {
		if !fileExists(filepath.Join(registry, item.typ, item.name+".md")) {
			t.Errorf("%s:%s should be in registry after collect", item.typ, item.name)
		}
	}
}

func TestCollect_UseDirective(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

## Rules

:::use rule:typescript

:::list
rule:eslint
:::end
`), 0644)

	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(`# Agent Instructions

## Rules

### typescript

Use strict mode.

### eslint

Enable recommended rules.
`), 0644)

	run(t, "collect", "--all")

	if !fileExists(filepath.Join(registry, "rule", "typescript.md")) {
		t.Error("rule:typescript from :::use should be collected")
	}
	if !fileExists(filepath.Join(registry, "rule", "eslint.md")) {
		t.Error("rule:eslint from :::list should be collected")
	}
}

func TestCollect_NestedHeadersInContent(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

## Rules

:::list
rule:typescript
:::end

## Workflows

:::list
workflow:deploy
:::end
`), 0644)

	// The AGENTS.md has ## Purpose and ## Guidelines inside the typescript
	// rule — these should be collected as part of the item content, not
	// treated as section boundaries (only ## headers matching directives
	// sections are boundaries).
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(`# Agent Instructions

## Rules

### typescript

# Rule: TypeScript

## Purpose

Enforce strict typing.

## Guidelines

- Use strict mode
- Prefer interfaces

## Workflows

### deploy

Deploy steps here.
`), 0644)

	run(t, "collect", "--all")

	stored := readFile(t, filepath.Join(registry, "rule", "typescript.md"))
	if !strings.Contains(stored, "Purpose") {
		t.Error("collected content should include nested ## Purpose header")
	}
	if !strings.Contains(stored, "Guidelines") {
		t.Error("collected content should include nested ## Guidelines header")
	}
	if strings.Contains(stored, "deploy") {
		t.Error("collected content should not bleed into the next section")
	}
}
