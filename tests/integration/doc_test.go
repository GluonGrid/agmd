package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoc_Link(t *testing.T) {
	setup(t)
	dir := chdir(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")

	out := run(t, "doc", "link", "my-docs")
	if !strings.Contains(out, "Linked") {
		t.Error("expected link confirmation")
	}
	if !fileExists(filepath.Join(dir, "docs", "my-docs")) {
		t.Error("symlink docs/my-docs should exist in project")
	}
}

func TestDoc_Unlink(t *testing.T) {
	setup(t)
	dir := chdir(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")
	run(t, "doc", "link", "my-docs")
	run(t, "doc", "unlink", "my-docs")

	if fileExists(filepath.Join(dir, "docs", "my-docs")) {
		t.Error("symlink should be removed after unlink")
	}
}

func TestDoc_List(t *testing.T) {
	setup(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "readme.md"), []byte("# Readme"), 0644)
	run(t, "doc", "add", docsDir, "project-docs")

	out := run(t, "doc", "list")
	if !strings.Contains(out, "project-docs") {
		t.Error("doc list should show project-docs")
	}
}

func TestDoc_Add(t *testing.T) {
	setup(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)

	out := run(t, "doc", "add", docsDir, "my-docs")
	if !strings.Contains(out, "my-docs") {
		t.Error("expected confirmation of added doc")
	}

	registry := os.Getenv("AGMD_HOME")
	if !fileExists(filepath.Join(registry, "doc", "my-docs")) {
		t.Error("doc should be stored in registry")
	}
}

func TestDoc_Show(t *testing.T) {
	setup(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "guide.md"), []byte("# Guide"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")

	out := run(t, "doc", "show", "my-docs")
	if !strings.Contains(out, "guide.md") {
		t.Errorf("doc show should list files in the doc folder, got:\n%s", out)
	}
}

func TestDoc_Delete(t *testing.T) {
	setup(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)
	run(t, "doc", "add", docsDir, "del-docs")
	run(t, "doc", "delete", "del-docs", "--force")

	out := run(t, "doc", "list")
	if strings.Contains(out, "del-docs") {
		t.Error("deleted doc should not appear in list")
	}
}

func TestDoc_AddForce(t *testing.T) {
	setup(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# v1"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")

	// Add again with different dir — should fail without --force
	docsDir2 := t.TempDir()
	os.WriteFile(filepath.Join(docsDir2, "index.md"), []byte("# v2"), 0644)
	_, err := runE(t, "doc", "add", docsDir2, "my-docs")
	if err == nil {
		t.Error("expected error when adding duplicate doc without --force")
	}

	// With --force it should succeed
	out := run(t, "doc", "add", docsDir2, "my-docs", "--force")
	if !strings.Contains(out, "my-docs") {
		t.Errorf("expected confirmation after force add, got:\n%s", out)
	}
}

func TestDoc_LinkCustomDest(t *testing.T) {
	setup(t)
	dir := chdir(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")

	// Link to a custom destination instead of docs/my-docs
	run(t, "doc", "link", "my-docs", "./reference")

	if !fileExists(filepath.Join(dir, "reference")) {
		t.Error("symlink should be created at custom destination ./reference")
	}
}

func TestDoc_LinkGitignore(t *testing.T) {
	setup(t)
	dir := chdir(t)

	docsDir := t.TempDir()
	os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs"), 0644)
	run(t, "doc", "add", docsDir, "my-docs")
	run(t, "doc", "link", "my-docs", "--gitignore")

	gitignore := readFile(t, filepath.Join(dir, ".gitignore"))
	if !strings.Contains(gitignore, "my-docs") {
		t.Errorf(".gitignore should contain the symlink path, got:\n%s", gitignore)
	}
}
