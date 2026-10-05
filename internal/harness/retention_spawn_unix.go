//go:build !windows

package harness

import (
	"fmt"
	"os/exec"
	"syscall"
)

// spawnDetachedRetentionPruner launches the prune child fully detached: setsid
// puts it in its own session and process group, so the parent hook process can
// exit without waiting (REQ-DP-005 — no Wait call; the child outlives the
// parent, reparents, and leaves no zombie behind the short-lived hook). Stdin,
// stdout and stderr stay disconnected (exec's nil defaults), so the child can
// never hold the hook's console or pipes open.
//
// @MX:WARN: [AUTO] Detached-process spawn: a child process escapes the hook's lifetime and process group.
// @MX:REASON: [AUTO] The child runs outside any hook budget by design (REQ-DP-005); its work is bounded by one
// lock-protected, idempotent PruneStaleEntries and the attempt stamp written before the work keeps an
// orphaned child from repeating within the interval.
func SpawnDetachedRetentionPruner(executable string, args []string) error {
	cmd := exec.Command(executable, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("retention: detached prune child spawn failed: %w", err)
	}
	return nil
}
