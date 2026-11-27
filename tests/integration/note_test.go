package integration

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The two files in testdata/notes/demo-app were written by the Agents app's
// notes service (notes.py); agmd note must read and extend them unchanged.

// noteFixtures is resolved before chdir moves the working directory.
var noteFixtures = func() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testdata", "notes", "demo-app")
}()

func setupNoteFixtures(t *testing.T) string {
	t.Helper()
	reg := setup(t)
	chdir(t)
	dst := filepath.Join(reg, "note", "demo-app")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"login-screen-inspiration.md", "release-checklist-1-4.md"} {
		writeFile(t, filepath.Join(dst, name), readFile(t, filepath.Join(noteFixtures, name)))
	}
	return reg
}

func TestNoteReadsServiceFiles(t *testing.T) {
	reg := setupNoteFixtures(t)

	out, err := runE(t, "note", "list", "--project", "demo-app")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "login-screen-inspiration\tLogin screen inspiration\t2026-11-2") ||
		!strings.HasPrefix(lines[1], "release-checklist-1-4\tRelease checklist 1.4\t") {
		t.Fatalf("list, last updated first:\n%s", out)
	}

	out, _ = runE(t, "note", "show", "release-checklist-1-4", "--project", "demo-app")
	if !strings.HasPrefix(out, "# Release checklist 1.4\n- [x] bump the version\n") || strings.Contains(out, "created:") {
		t.Fatalf("show: %q", out)
	}

	out, _ = runE(t, "note", "path", "login-screen-inspiration", "--project", "demo-app")
	if strings.TrimSpace(out) != filepath.Join(reg, "note", "demo-app", "login-screen-inspiration.md") {
		t.Fatalf("path: %q", out)
	}

	out, _ = runE(t, "note", "list", "--all")
	if strings.TrimSpace(out) != "demo-app\t2 notes" {
		t.Fatalf("list --all: %q", out)
	}
}

func TestNoteAppendKeepsServiceFormat(t *testing.T) {
	reg := setupNoteFixtures(t)
	path := filepath.Join(reg, "note", "demo-app", "release-checklist-1-4.md")
	before := readFile(t, path)

	if out, err := runE(t, "note", "append", "release-checklist-1-4", "--project", "demo-app", "--", "- [ ] tell the testers"); err != nil ||
		strings.TrimSpace(out) != "added to demo-app/release-checklist-1-4" {
		t.Fatalf("append: %v %q", err, out)
	}
	after := readFile(t, path)
	updated := regexp.MustCompile(`(?m)^updated: \d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$`)
	if !updated.MatchString(after) || strings.Contains(after, "updated: 2026-11-25T13:54:48Z") {
		t.Fatalf("updated not stamped:\n%s", after)
	}
	want := updated.ReplaceAllString(before, "U") + "\n- [ ] tell the testers\n"
	if got := updated.ReplaceAllString(after, "U"); got != want {
		t.Fatalf("append changed more than the end:\n%q\nwant\n%q", got, want)
	}
	if !fileExists(filepath.Join(reg, "note", ".lock")) || readFile(t, filepath.Join(reg, "note", ".gitignore")) != ".lock\n.trash/\n.*.tmp\n" {
		t.Fatal("lock or .gitignore missing")
	}

	// stdin, and the tasks line of the other note survives
	r, w, _ := os.Pipe()
	w.WriteString("https://example.com/a\n\n")
	w.Close()
	old := os.Stdin
	os.Stdin = r
	_, err := runE(t, "note", "append", "login-screen-inspiration", "-", "--project", "demo-app")
	os.Stdin = old
	if err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(reg, "note", "demo-app", "login-screen-inspiration.md"))
	if !strings.Contains(got, "\ntasks: login-redesign\n---\n") || !strings.HasSuffix(got, "- [ ] typed in the app (more)\n\nhttps://example.com/a\n") {
		t.Fatalf("stdin append:\n%s", got)
	}
}

func TestNoteNewSlugRule(t *testing.T) {
	reg := setup(t)
	chdir(t)
	dir := filepath.Join(reg, "note", "p")

	out, err := runE(t, "note", "new", "--project", "p", "--title", "  Login  Screen: ideas! ", "--tasks", "a, b,a,bad name", "first line")
	if err != nil || strings.TrimSpace(out) != "login-screen-ideas" {
		t.Fatalf("new: %v %q", err, out)
	}
	got := readFile(t, filepath.Join(dir, "login-screen-ideas.md"))
	if !regexp.MustCompile(`^---\ntitle: Login Screen: ideas!\ncreated: (\S+Z)\nupdated: (\S+Z)\ntasks: a, b\n---\nfirst line\n$`).MatchString(got) {
		t.Fatalf("file:\n%q", got)
	}
	if fi, _ := os.Stat(filepath.Join(dir, "login-screen-ideas.md")); fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", fi.Mode())
	}

	for i, want := range []string{"login-screen-ideas-2", "login-screen-ideas-3"} {
		out, _ = runE(t, "note", "new", "--project", "p", "--title", "Login screen ideas")
		if strings.TrimSpace(out) != want {
			t.Fatalf("taken %d: %q", i, out)
		}
	}
	if got := readFile(t, filepath.Join(dir, "login-screen-ideas-2.md")); !strings.HasSuffix(got, "\n---\n") {
		t.Fatalf("empty note: %q", got)
	}

	long := strings.Repeat("abcdefghi ", 8) // 80 characters
	out, _ = runE(t, "note", "new", "--project", "p", "--title", long)
	slug := strings.TrimSpace(out)
	if len(slug) != 59 || strings.HasSuffix(slug, "-") { // 60 characters, the trailing - cut
		t.Fatalf("long slug %q", slug)
	}
	out, _ = runE(t, "note", "new", "--project", "p", "--title", long)
	if strings.TrimSpace(out) != slug[:56]+"-2" {
		t.Fatalf("long slug taken: %q", out)
	}
	out, _ = runE(t, "note", "new", "--project", "p", "--title", "!!!")
	if strings.TrimSpace(out) != "note" {
		t.Fatalf("no letters: %q", out)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left: %s", e.Name())
		}
	}
}

func TestNoteErrors(t *testing.T) {
	reg := setupNoteFixtures(t)
	cases := [][]string{
		{"note", "show", "missing", "--project", "demo-app"},
		{"note", "show", "../x", "--project", "demo-app"},
		{"note", "show", "x", "--project", ".hidden"},
		{"note", "new", "--project", "demo-app"},
		{"note", "new", "--project", "demo-app", "--title", "   "},
		{"note", "append", "release-checklist-1-4", " \n\n", "--project", "demo-app"},
		{"note", "append", "missing", "text", "--project", "demo-app"},
	}
	for _, args := range cases {
		if _, err := runE(t, args...); err == nil {
			t.Errorf("%v: expected an error", args)
		}
	}
	if fileExists(filepath.Join(reg, "note", "demo-app", "missing.md")) {
		t.Fatal("append made a note")
	}

	// a symlinked note is refused, also for writing
	target := filepath.Join(t.TempDir(), "outside.md")
	writeFile(t, target, "---\ntitle: x\n---\nsecret\n")
	if err := os.Symlink(target, filepath.Join(reg, "note", "demo-app", "link.md")); err != nil {
		t.Fatal(err)
	}
	if out, err := runE(t, "note", "show", "link", "--project", "demo-app"); err == nil || strings.Contains(out, "secret") {
		t.Fatalf("symlink read: %v %q", err, out)
	}
	if _, err := runE(t, "note", "append", "link", "more", "--project", "demo-app"); err == nil {
		t.Fatal("symlink append accepted")
	}
	if readFile(t, target) != "---\ntitle: x\n---\nsecret\n" {
		t.Fatal("symlink target written")
	}
	if out, _ := runE(t, "note", "list", "--project", "demo-app"); strings.Contains(out, "link") {
		t.Fatalf("symlink listed: %q", out)
	}
}
