//go:build unix

package cmd

import "syscall"

// oNoFollow makes an open fail on a symlink, as the notes service does.
const oNoFollow = syscall.O_NOFOLLOW
