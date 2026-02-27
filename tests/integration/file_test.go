package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFile_List(t *testing.T) {
	setup(t)

	run(t, "file", "new", "deploy.sh", "--content", "#!/bin/bash\ndeploy")

	out := run(t, "file", "list")
	if !strings.Contains(out, "deploy.sh") {
		t.Error("file list should show deploy.sh")
	}
}

func TestFile_NewAndShow(t *testing.T) {
	registry := setup(t)

	run(t, "file", "new", "check.sh", "--content", "#!/bin/bash\necho hello")

	if !fileExists(filepath.Join(registry, "file", "check.sh")) {
		t.Error("file should be stored in registry")
	}

	out := run(t, "file", "show", "check.sh")
	if !strings.Contains(out, "echo hello") {
		t.Error("file show should display file content")
	}
}

func TestFile_Add(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	// Create a file on disk then add it to the registry
	src := filepath.Join(dir, "build.sh")
	os.WriteFile(src, []byte("#!/bin/bash\necho build"), 0644)

	out := run(t, "file", "add", src)
	if !strings.Contains(out, "build.sh") {
		t.Errorf("expected 'build.sh' in output, got:\n%s", out)
	}
	if !fileExists(filepath.Join(registry, "file", "build.sh")) {
		t.Error("file should be stored in registry after file add")
	}
}

func TestFile_AddForce(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	src := filepath.Join(dir, "run.sh")
	os.WriteFile(src, []byte("#!/bin/bash\necho v1"), 0644)
	run(t, "file", "add", src)

	// Second add without --force should fail
	os.WriteFile(src, []byte("#!/bin/bash\necho v2"), 0644)
	_, err := runE(t, "file", "add", src)
	if err == nil {
		t.Error("expected error when adding duplicate file without --force")
	}

	// With --force it should succeed and update content
	run(t, "file", "add", src, "--force")

	stored := readFile(t, filepath.Join(registry, "file", "run.sh"))
	if !strings.Contains(stored, "echo v2") {
		t.Errorf("file should be updated with --force, got:\n%s", stored)
	}
}

func TestFile_Delete(t *testing.T) {
	setup(t)

	run(t, "file", "new", "tmp.sh", "--content", "#!/bin/bash")
	run(t, "file", "delete", "tmp.sh", "--force")

	out := run(t, "file", "list")
	if strings.Contains(out, "tmp.sh") {
		t.Error("deleted file should not appear in list")
	}
}
