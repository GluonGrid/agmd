//go:build unix

package cmd

import (
	"os/exec"
	"syscall"
)

// checkProcessGroup runs a check in its own process group and kills the
// whole group on timeout, so a check's children do not outlive it.
func checkProcessGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error {
		return syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
	}
}
