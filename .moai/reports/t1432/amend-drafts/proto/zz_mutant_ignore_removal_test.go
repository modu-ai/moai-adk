//go:build !windows

package harness

// DRAFT mutant of the heal-lock prototype (t1432 amendment 0.4.1): the heal ignores a failed removal of
// the state-path entry (the mutant "a heal that ignores the removal failure" of AC-HRH-003).

func init() { protoIgnoreRemoval = true }
