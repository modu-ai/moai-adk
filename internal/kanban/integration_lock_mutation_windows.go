//go:build windows

// integration_lock_mutation_windows.go — the Windows side of the mutation
// lock's wedge recovery.
//
// On Windows the artifact IS the lock (atomic-create; board_lock_windows.go),
// so a process killed inside the critical section leaves a file that blocks
// every subsequent mutation of the record — permanently, and for a lock whose
// whole lifetime is meant to be one CLI invocation. That asymmetry is why this
// file exists and its Unix sibling is a no-op.
//
// The recovery is the board clear's discipline, reused rather than re-derived:
// remove ONLY when the recorded owner is positively observed absent, re-read
// the recorded identity immediately before the unlink, and abort on any
// mismatch. Uncertainty resolves toward TREAT-AS-LIVE — an unprobeable or
// unparseable owner is not cleared, the caller reports busy, and an operator
// retries. The inherited TOCTOU residual (AP-29) is narrowed by the re-read,
// not closed, and no new engineering is attempted here.
package kanban

// clearWedgedIntegrationMutationLock clears a mutation artifact whose recorded
// owner is positively dead. It is reached only after a contender has exhausted
// the whole wait budget, never on the ordinary acquire path.
func clearWedgedIntegrationMutationLock(path string) (*ClearStaleReport, error) {
	return clearStaleLockAtPath(path, "integration mutation lock")
}
