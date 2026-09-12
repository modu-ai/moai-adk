//go:build !windows

// slot_lease_mutation_unix.go — the Unix side of the slot-lease mutation
// lock's wedge recovery, which is deliberately nothing: flock(2) is released
// by the kernel when the holder exits, so an orphaned artifact blocks nothing.
package kanban

// clearWedgedSlotLeaseMutationLock on Unix removes nothing and reports why.
func clearWedgedSlotLeaseMutationLock(_ string) (*ClearStaleReport, error) {
	return &ClearStaleReport{
		Removed: false,
		Reason:  "slot-lease mutation-lock wedge clear is gated to windows; the unix substrate releases flock on process exit",
	}, nil
}
