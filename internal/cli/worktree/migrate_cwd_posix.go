//go:build !windows

// migrate_cwd_posix.go — the migration's process-cwd scan on POSIX hosts. A
// rename succeeds while a process stands inside the tree, so this scan is the
// only guard; any lsof failure skips every tree (card t2075 keeps that).

package worktree

import (
	"bytes"
	"fmt"
	"strings"
)

// defaultActiveProcessCWDs scans lsof for the working directories of every
// active process (the conservative process-level guard for worktree moves).
func defaultActiveProcessCWDs() ([]string, error) {
	out, err := boundedOutput(lsofCommandTimeout, "", "lsof", "-a", "-d", "cwd", "-F0n")
	if err != nil {
		return nil, fmt.Errorf("lsof cwd scan: %w", err)
	}
	var directories []string
	for _, field := range bytes.Split(out, []byte{0}) {
		value := strings.TrimLeft(string(field), "\r\n")
		if strings.HasPrefix(value, "n/") {
			directories = append(directories, strings.TrimSuffix(value[1:], " (deleted)"))
		}
	}
	return directories, nil
}
