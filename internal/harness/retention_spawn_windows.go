//go:build windows

package harness

import (
	"fmt"
	"os/exec"
	"syscall"
)

// detachedProcessFlag is the Win32 DETACHED_PROCESS creation flag (0x00000008,
// documented PROCESS_CREATION_FLAGS value). The stdlib syscall package exports
// CREATE_NEW_PROCESS_GROUP but not the console-suppression modes; this local
// constant supplies the ONE console-suppression mode REQ-DP-006 requires (the
// alternatives DETACHED_PROCESS / CREATE_NO_WINDOW are exclusive — exactly one
// is set, never both).
const detachedProcessFlag = 0x00000008

// spawnDetachedRetentionPruner launches the prune child detached on Windows
// (REQ-DP-006): CREATE_NEW_PROCESS_GROUP gives it its own process group,
// DETACHED_PROCESS is the single console-suppression mode (no console flash),
// and HideWindow keeps a window from appearing where the combination would
// show one. The parent exits without waiting (REQ-DP-005 — no Wait call).
// The child's runtime behavior on Windows stays documented-unobserved (spec F3):
// build+vet parity is the enforced half of REQ-DP-006, and the state-file lock
// remains the documented in-process-only mutex (retention.go, F5 limitation).
//
// @MX:WARN: [AUTO] Detached-process spawn: a child process escapes the hook's lifetime and process group.
// @MX:REASON: [AUTO] The child runs outside any hook budget by design (REQ-DP-005); its work is bounded by one
// lock-protected, idempotent PruneStaleEntries and the attempt stamp written before the work keeps an
// orphaned child from repeating within the interval.
func SpawnDetachedRetentionPruner(executable string, args []string) error {
	cmd := exec.Command(executable, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcessFlag,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("retention: detached prune child spawn failed: %w", err)
	}
	return nil
}
