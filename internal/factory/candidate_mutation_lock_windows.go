//go:build windows

package factory

// clearWedgedCandidateMutationLock clears a candidate mutation artifact
// whose holder is positively observed absent (Windows: the artifact IS the
// lock, so a holder killed inside the push+record section wedges every
// later candidate push of this card). The generic recovery owns the
// liveness re-read; this wrapper only names the scope in its report.
func clearWedgedCandidateMutationLock(path string) (*ClearStaleReport, error) {
	return clearStaleLockAtPath(path, "candidate mutation lock")
}
