//go:build windows

// slot_lease_mutation_windows.go — the Windows side of the slot-lease
// mutation lock's wedge recovery. On Windows the artifact IS the lock
// (atomic-create), so a holder killed inside the critical section would block
// every later mutation of that resource. The board clear's discipline is
// reused unchanged: remove ONLY when the recorded owner is positively observed
// absent, re-reading the identity immediately before the unlink.
package kanban

// clearWedgedSlotLeaseMutationLock clears a per-resource mutation artifact
// whose recorded owner is positively dead. Reached only after the whole wait
// budget was spent.
func clearWedgedSlotLeaseMutationLock(path string) (*ClearStaleReport, error) {
	return clearStaleLockAtPath(path, "slot lease mutation lock")
}
