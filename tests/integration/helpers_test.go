package integration

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"agmd/cmd"
)

// run executes an agmd command in-process, returns combined stdout+stderr.
func run(t *testing.T, args ...string) string {
	t.Helper()
	out, _ := runE(t, args...)
	return out
}

// runE executes an agmd command and returns combined output + error.
func runE(t *testing.T, args ...string) (string, error) {
	t.Helper()

	r, w, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w

	err := cmd.ExecuteArgs(args)

	w.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	var buf bytes.Buffer
	io.Copy(&buf, r)

	t.Logf("agmd %v\n%s", args, buf.String())
	return buf.String(), err
}

// setup runs agmd setup into a fresh registry and sets AGMD_HOME for this test.
// Returns the registry path.
func setup(t *testing.T) string {
	t.Helper()
	registry := filepath.Join(t.TempDir(), "registry")
	t.Setenv("AGMD_HOME", registry)
	run(t, "setup")
	return registry
}

// chdir changes CWD to a new temp dir for this test and restores it on cleanup.
// Returns the new directory path.
func chdir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	return dir
}

// fileExists reports whether path exists (or is a symlink, even a dangling one).
func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// readFile reads a file and returns its content as string.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile %s: %v", path, err)
	}
	return string(b)
}
