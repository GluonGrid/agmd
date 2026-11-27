//go:build windows

package cmd

// oNoFollow: Windows has no O_NOFOLLOW; note paths are still checked with Lstat.
const oNoFollow = 0
