//go:build !windows

package harness

// DRAFT mutant of the heal-lock prototype (t1432 amendment 0.4.1): the common path (a healthy or absent
// state entry) opens and locks a heal-lock file that already exists. It never creates one, so a test
// that only asks "no heal-lock entry exists afterwards" cannot see it; a test that pre-creates the file
// and holds it can.

func init() { protoLockCommonIfExists = true }
