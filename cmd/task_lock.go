package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"agmd/pkg/registry"
)

// Several agents (and the web board) may run agmd against the same project at
// once. Two rules keep their writes from clobbering each other:
//
//  1. Every read-modify-write of a project's files (task files,
//     assignments.json, project.json) happens while holding the project lock.
//  2. Every file is written to a temp file and renamed into place, so readers
//     never see a half-written file, even without taking the lock.

const taskLockFilename = ".lock"

// lockProject blocks until it holds an exclusive, cross-process lock on the
// project's task directory, creating the directory if needed. Call the
// returned function to release it.
func lockProject(reg *registry.Registry, projectName string) (func(), error) {
	taskDir := getTaskDir(reg, projectName)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create task directory: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(taskDir, taskLockFilename), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open project lock: %w", err)
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to lock project '%s': %w", projectName, err)
	}
	return func() {
		unlockFile(f) //nolint:errcheck
		f.Close()
	}, nil
}

// writeFileAtomic replaces path with data via a temp file and rename.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := writeTempBeside(path, data, perm)
	if err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// writeFileExclusive creates path with data, failing with an error that
// satisfies os.IsExist when path already exists. The file appears complete or
// not at all.
func writeFileExclusive(path string, data []byte, perm os.FileMode) error {
	tmp, err := writeTempBeside(path, data, perm)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	return os.Link(tmp, path)
}

func writeTempBeside(path string, data []byte, perm os.FileMode) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(name)
		return "", err
	}
	if err := f.Chmod(perm); err != nil {
		f.Close()
		os.Remove(name)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}
