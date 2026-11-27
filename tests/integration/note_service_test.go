package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// These run the Agents app's own notes service (notes.py) against the same
// folder, when it is installed: set AGMD_NOTES_PY to its path, or it is looked
// for in /opt/agents/bin. Without it they are skipped.

func notesPy(t *testing.T) string {
	t.Helper()
	p := os.Getenv("AGMD_NOTES_PY")
	if p == "" {
		p = "/opt/agents/bin/notes.py"
	}
	if _, err := os.Stat(p); err != nil {
		t.Skip("notes.py not installed")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("no python3")
	}
	return p
}

// pyNotes runs code with s = notes.Store over registry/note.
func pyNotes(t *testing.T, py, registry, code string) string {
	t.Helper()
	script := fmt.Sprintf("import sys; sys.path.insert(0, %q); import notes\n"+
		"s = notes.Store(%q, %q)\n%s\n", filepath.Dir(py), filepath.Join(registry, "note"), filepath.Join(registry, "media"), code)
	out, err := exec.Command("python3", "-I", "-c", script).CombinedOutput()
	if err != nil {
		t.Fatalf("python: %v\n%s", err, out)
	}
	return string(out)
}

func TestNoteWithService(t *testing.T) {
	py := notesPy(t)
	registry := setup(t)

	// the service makes a note, agmd reads and appends; the service reads it back
	pyNotes(t, py, registry, `s.create("p", "From the app", "made in the app", ["t1"])`)
	if out, _ := runE(t, "note", "show", "from-the-app", "--project", "p"); out != "# From the app\nmade in the app\n" {
		t.Fatalf("show: %q", out)
	}
	if _, err := runE(t, "note", "append", "from-the-app", "added by agmd", "--project", "p"); err != nil {
		t.Fatal(err)
	}
	out := pyNotes(t, py, registry, `n = s.read("p", "from-the-app"); print(repr((n["title"], n["body"], n["tasks"])))`)
	if strings.TrimSpace(out) != `('From the app', 'made in the app\n\nadded by agmd\n', ['t1'])` {
		t.Fatalf("service read: %s", out)
	}

	// agmd makes a note with the slug the service would give
	runE(t, "note", "new", "--project", "p", "--title", "From the app")
	out = pyNotes(t, py, registry, `print(s.create("p", "From the app")["slug"]); print(s.read("p", "from-the-app-2")["title"])`)
	if out != "from-the-app-3\nFrom the app\n" {
		t.Fatalf("slugs: %q", out)
	}

	// byte-identical rendering: the same append through both, same file
	a := filepath.Join(registry, "note", "p", "from-the-app-2.md")
	b := filepath.Join(registry, "note", "p", "from-the-app-3.md")
	runE(t, "note", "append", "from-the-app-2", "x\r\ny", "--project", "p")
	pyNotes(t, py, registry, `s.append("p", "from-the-app-3", "x\r\ny")`)
	strip := func(s string) string {
		lines := strings.Split(s, "\n")
		var keep []string
		for _, l := range lines {
			if !strings.HasPrefix(l, "created: ") && !strings.HasPrefix(l, "updated: ") {
				keep = append(keep, l)
			}
		}
		return strings.Join(keep, "\n")
	}
	if strip(readFile(t, a)) != strip(readFile(t, b)) {
		t.Fatalf("renders differ:\n%q\n%q", readFile(t, a), readFile(t, b))
	}
}

func TestNoteConcurrency_WithServiceAppends(t *testing.T) {
	py := notesPy(t)
	registry := setup(t)
	pyNotes(t, py, registry, `s.create("p", "Shared")`)
	bin := agmdBinary(t)
	const n = 6
	var wg sync.WaitGroup
	errs := make([]error, 2*n)
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			c := exec.Command(bin, "note", "append", "shared", fmt.Sprintf("agmd %d", i), "--project", "p")
			c.Env = os.Environ()
			if out, err := c.CombinedOutput(); err != nil {
				errs[i] = fmt.Errorf("%v: %s", err, out)
			}
		}(i)
		go func(i int) {
			defer wg.Done()
			script := fmt.Sprintf("import sys; sys.path.insert(0, %q); import notes\n"+
				"notes.Store(%q, %q).append('p', 'shared', 'service %d')\n", filepath.Dir(py),
				filepath.Join(registry, "note"), filepath.Join(registry, "media"), i)
			if out, err := exec.Command("python3", "-I", "-c", script).CombinedOutput(); err != nil {
				errs[n+i] = fmt.Errorf("%v: %s", err, out)
			}
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	got := readFile(t, filepath.Join(registry, "note", "p", "shared.md"))
	for i := 0; i < n; i++ {
		for _, w := range []string{"agmd", "service"} {
			if !strings.Contains(got, fmt.Sprintf("\n%s %d\n", w, i)) {
				t.Errorf("%s %d lost:\n%s", w, i, got)
			}
		}
	}
}
