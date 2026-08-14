//go:build !windows

// board_lock_clear_unix.go — the Unix stale-lock clear is gated out
// (SPEC-KANBAN-BOARD-001 REQ-KB-023, review finding F5 containment).
//
// board_lock_unix.go's own header records that the Unix substrate holds
// flock(2) on an open descriptor which the kernel releases on process exit,
// so a killed holder leaves an inert artifact — the fourteen orphaned
// spec-close-*.lock files in this repository's .moai/state/ are the worked
// example. The clear therefore exists for the Windows substrate, where the
// artifact IS the lock and a killed holder blocks every subsequent mutation
// permanently. Gating the clear to Windows removes the Unix-side window the
// M1 implementation opened (acquire flock, THEN record owner identity) —
// that gap is unreachable when the kernel drops the lock on exit — and
// touches the documented AP-29 residual not at all (it is a Windows-substrate
// residual and the Windows clear keeps its re-read mitigation).
package kanban

// ClearStaleBoardLock on Unix is a no-op reporting the platform gate: the
// stale-clear window does not arise here. The lock artifact is inert and a
// subsequent AcquireBoardLock succeeds by taking a fresh flock on the same
// path.
func ClearStaleBoardLock(root string) (*ClearStaleReport, error) {
	return &ClearStaleReport{
		Removed: false,
		Reason:  "stale-lock clear is gated to windows; the unix substrate releases flock on process exit and an orphaned artifact blocks nothing here",
	}, nil
}
