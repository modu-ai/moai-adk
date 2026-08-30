//go:build !windows

// integration_lock_mutation_unix.go — the Unix side of the mutation lock's
// wedge recovery, which is deliberately nothing at all.
//
// The Unix substrate holds flock(2) on an open descriptor and the kernel
// releases it when the descriptor closes, which it does on process exit
// (board_lock_unix.go's own header). A holder killed mid-mutation therefore
// leaves an INERT artifact: the next acquirer takes a fresh flock on the same
// path and proceeds. There is no wedge here to clear, so no clear is added —
// mirroring board_lock_clear_unix.go's gate and its stated reason.
package kanban

// clearWedgedIntegrationMutationLock on Unix reports the platform gate and
// removes nothing. Returning a report rather than an error keeps the caller's
// shape identical on both platforms: the caller inspects Removed, which is
// false here, and falls through to the busy sentinel.
func clearWedgedIntegrationMutationLock(_ string) (*ClearStaleReport, error) {
	return &ClearStaleReport{
		Removed: false,
		Reason:  "mutation-lock wedge clear is gated to windows; the unix substrate releases flock on process exit and an orphaned artifact blocks nothing here",
	}, nil
}
