package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createProfileFromDirectives sets up a directives.md with known content and
// saves it as a named profile using `agmd new profile:name`.
func createProfileFromDirectives(t *testing.T, name, content string) {
	t.Helper()
	dir, _ := os.Getwd()
	_ = dir
	os.WriteFile("directives.md", []byte(content), 0644)
	run(t, "new", "profile:"+name)
}

func TestProfile_NewFromDirectives(t *testing.T) {
	setup(t)
	chdir(t)

	// Create a directives.md then save it as a profile
	os.WriteFile("directives.md", []byte("# My Profile\nDo the thing."), 0644)
	out := run(t, "new", "profile:myprofile")

	if !strings.Contains(out, "myprofile") {
		t.Errorf("expected profile name in output, got:\n%s", out)
	}
}

func TestProfile_ShowContent(t *testing.T) {
	setup(t)
	chdir(t)

	os.WriteFile("directives.md", []byte("# Team Rules\nFollow them."), 0644)
	run(t, "new", "profile:teamrules")

	out := run(t, "show", "profile:teamrules")
	if !strings.Contains(out, "Follow them.") {
		t.Errorf("profile content should be visible via show, got:\n%s", out)
	}
}

func TestProfile_ShowRawIncludesFrontmatter(t *testing.T) {
	setup(t)
	chdir(t)

	os.WriteFile("directives.md", []byte("# Rules\nBe consistent."), 0644)
	run(t, "new", "profile:rawtest")

	out := run(t, "show", "profile:rawtest", "--raw")
	// --raw should include the frontmatter block (name field)
	if !strings.Contains(out, "name:") {
		t.Errorf("--raw output should include frontmatter name:, got:\n%s", out)
	}
}

func TestProfile_ListsInRegistry(t *testing.T) {
	setup(t)
	chdir(t)

	os.WriteFile("directives.md", []byte("# Rules"), 0644)
	run(t, "new", "profile:listed-profile")

	out := run(t, "list", "profile")
	if !strings.Contains(out, "listed-profile") {
		t.Errorf("profile should appear in list, got:\n%s", out)
	}
}

// TestProfile_InitCopiesContent verifies that `agmd init profile:name` writes the
// profile body (without frontmatter) into directives.md.
func TestProfile_InitCopiesContent(t *testing.T) {
	setup(t)
	dir := chdir(t)

	// Save a profile from a directives.md with recognizable content
	os.WriteFile("directives.md", []byte("# MySetup Profile\n\nUse strict typing."), 0644)
	run(t, "new", "profile:mysetup")

	// Remove directives.md so init can create a fresh one
	os.Remove(filepath.Join(dir, "directives.md"))

	run(t, "init", "profile:mysetup")

	if !fileExists(filepath.Join(dir, "directives.md")) {
		t.Fatal("directives.md should be created from profile")
	}

	directives := readFile(t, filepath.Join(dir, "directives.md"))

	// Profile body should appear in directives.md
	if !strings.Contains(directives, "Use strict typing.") {
		t.Errorf("directives.md should contain profile body, got:\n%s", directives)
	}

	// Frontmatter (name:, description:) should NOT be written to directives.md
	if strings.Contains(directives, "name: mysetup") {
		t.Error("directives.md should not contain profile frontmatter")
	}
}

// TestProfile_InitCopiesFiles verifies that a profile with a `files:` frontmatter
// field causes the referenced registry files to be copied into the project on init.
func TestProfile_InitCopiesFiles(t *testing.T) {
	registry := setup(t)
	dir := chdir(t)

	// Add a file to the registry
	run(t, "file", "new", "setup.sh", "--content", "#!/bin/bash\necho setup")

	// Write a profile .md directly into the registry with a `files:` frontmatter field.
	// (agmd new profile:name reads directives.md and doesn't expose the files: field,
	// so we write the profile file manually.)
	profileContent := "---\nname: withfiles\ndescription: Profile with files\nfiles:\n  - file:setup.sh\n---\n# WithFiles Profile\nDo stuff.\n"
	profilePath := filepath.Join(registry, "profile", "withfiles.md")
	os.MkdirAll(filepath.Dir(profilePath), 0755)
	os.WriteFile(profilePath, []byte(profileContent), 0644)

	run(t, "init", "profile:withfiles")

	// setup.sh should have been copied next to directives.md
	if !fileExists(filepath.Join(dir, "setup.sh")) {
		t.Error("setup.sh should be copied into project dir by init profile:withfiles")
	}

	// The copied file should contain the original content
	copied := readFile(t, filepath.Join(dir, "setup.sh"))
	if !strings.Contains(copied, "echo setup") {
		t.Errorf("copied setup.sh should contain original content, got:\n%s", copied)
	}
}

func TestProfile_Delete(t *testing.T) {
	setup(t)
	chdir(t)

	os.WriteFile("directives.md", []byte("# Delete me"), 0644)
	run(t, "new", "profile:todelete")
	run(t, "delete", "profile:todelete", "--force")

	out := run(t, "list", "profile")
	if strings.Contains(out, "todelete") {
		t.Error("deleted profile should not appear in list")
	}
}
