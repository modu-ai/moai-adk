//go:build !windows

package factory

// clearWedgedCandidateMutationLock is a no-op off Windows: the unix
// substrate releases flock on process exit, so an orphaned candidate
// mutation artifact blocks nothing (the same shape the integration
// mutation lock's unix file records).
func clearWedgedCandidateMutationLock(_ string) (*ClearStaleReport, error) {
	return &ClearStaleReport{
		Removed: false,
		Reason:  "candidate mutation-lock wedge clear is gated to windows; the unix substrate releases flock on process exit and an orphaned artifact blocks nothing here",
	}, nil
}
