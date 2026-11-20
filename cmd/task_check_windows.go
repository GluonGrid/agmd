//go:build windows

package cmd

import "os/exec"

// checkProcessGroup keeps the default: only the shell is killed on timeout.
func checkProcessGroup(c *exec.Cmd) {}
