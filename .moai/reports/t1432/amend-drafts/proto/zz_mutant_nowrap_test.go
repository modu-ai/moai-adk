//go:build !windows

package harness

// DRAFT mutant of the heal-lock prototype (t1432 amendment 0.4.1): the heal-lock error formats its
// cause with %v instead of wrapping it, so errors.Is(err, fs.ErrPermission) no longer holds.

func init() { protoWrap = false }
