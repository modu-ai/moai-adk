//go:build !windows

// sweep_cwd_posix.go — the sweep's process-cwd probe for POSIX hosts
// (linux/darwin), REQ-WS-006.
//
// The lsof pattern is duplicated (small, deliberate) from
// internal/cli/update_worktree_processes.go `activeProcessCWDs`: that helper
// lives in the PARENT cli package, which imports this worktree package for
// DI wiring — a child→parent import would cycle (plan.md §B import-cycle
// hazard). The original is cited here as the pattern source; keep the two
// parsers in sync.

package worktree

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// platformProcessCWDs lists the working directories of every process on the
// host via one whole-system `lsof -a -d cwd -F0n` invocation (NUL-field
// parse). One invocation per sweep run — not per tree — keeps large runs off
// lsof's per-fd slow path (design.md §C). lsof absent or failing returns an
// error, which the sweep renders as an unanswerable predicate: the tree is
// preserved, never reported unoccupied.
var platformProcessCWDs = func() ([]string, error) {
	command := exec.Command("lsof", "-a", "-d", "cwd", "-F0n")
	out, err := command.Output()
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
