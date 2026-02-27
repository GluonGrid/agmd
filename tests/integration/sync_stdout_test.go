package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSync_Stdout(t *testing.T) {
	reg := setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte(`# Project

## Rules

:::use rule:go-style
`), 0644)

	// Create a rule in the registry so it can be expanded
	os.MkdirAll(filepath.Join(reg, "rule"), 0755)
	os.WriteFile(filepath.Join(reg, "rule", "go-style.md"), []byte(`---
name: go-style
description: Go style guide
---
Use gofmt. Keep functions short.
`), 0644)

	out := run(t, "sync", "--stdout")

	// stdout should contain the expanded content
	if !strings.Contains(out, "gofmt") {
		t.Error("--stdout should print expanded content to stdout")
	}
	if !strings.Contains(out, "Keep functions short") {
		t.Error("--stdout should contain the full rule content")
	}

	// AGENTS.md should NOT exist on disk
	if fileExists(filepath.Join(dir, "AGENTS.md")) {
		t.Error("--stdout should not write AGENTS.md to disk")
	}

	// Cache should be written
	cachePath := filepath.Join(dir, ".agmd", "cache", "last-sync.md")
	if !fileExists(cachePath) {
		t.Error("--stdout should write sync cache")
	}
}

func TestSync_Stdout_NoAgentsFile(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nSome instructions.\n"), 0644)

	out := run(t, "sync", "--stdout")

	if !strings.Contains(out, "Some instructions") {
		t.Error("--stdout should output directives content")
	}

	if fileExists(filepath.Join(dir, "AGENTS.md")) {
		t.Error("--stdout should not create AGENTS.md")
	}
}

func TestSync_Diff_FirstRun(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nFirst run content.\n"), 0644)

	// First --diff with no cache should output full content
	out := run(t, "sync", "--diff")

	if !strings.Contains(out, "First run content") {
		t.Error("--diff with no cache should output full content")
	}

	// Cache should now exist
	cachePath := filepath.Join(dir, ".agmd", "cache", "last-sync.md")
	if !fileExists(cachePath) {
		t.Error("--diff should write sync cache")
	}
}

func TestSync_Diff_NoChanges(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nStable content.\n"), 0644)

	// First run to populate cache
	run(t, "sync", "--stdout")

	// Second run with --diff should produce no output (nothing changed)
	out := run(t, "sync", "--diff")

	if strings.TrimSpace(out) != "" {
		t.Errorf("--diff with no changes should produce no output, got:\n%s", out)
	}
}

func TestSync_Diff_WithChanges(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nOriginal content.\n"), 0644)

	// First run to populate cache
	run(t, "sync", "--stdout")

	// Modify directives
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nUpdated content.\n"), 0644)

	// --diff should show the change
	out := run(t, "sync", "--diff")

	if !strings.Contains(out, "---") || !strings.Contains(out, "+++") {
		t.Error("--diff should output unified diff format")
	}
	if !strings.Contains(out, "-Original content") {
		t.Error("--diff should show removed lines")
	}
	if !strings.Contains(out, "+Updated content") {
		t.Error("--diff should show added lines")
	}
}

func TestSync_Diff_CacheUpdatedAfterDiff(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nV1.\n"), 0644)
	run(t, "sync", "--stdout")

	// Change to V2
	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nV2.\n"), 0644)
	run(t, "sync", "--diff")

	// Running --diff again should produce no output (cache was updated)
	out := run(t, "sync", "--diff")
	if strings.TrimSpace(out) != "" {
		t.Errorf("second --diff should produce no output after cache update, got:\n%s", out)
	}
}

func TestSync_Stdout_And_Diff_MutuallyExclusive(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n"), 0644)

	_, err := runE(t, "sync", "--stdout", "--diff")
	if err == nil {
		t.Error("--stdout and --diff together should error")
	}
}

func TestSync_File_StillWorks(t *testing.T) {
	setup(t)
	dir := chdir(t)

	os.WriteFile(filepath.Join(dir, "directives.md"), []byte("# Project\n\nFile mode.\n"), 0644)

	run(t, "sync")

	// AGENTS.md should exist
	content := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(content, "File mode") {
		t.Error("plain sync should still write AGENTS.md")
	}
}
