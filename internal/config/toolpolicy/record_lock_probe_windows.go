//go:build windows

package toolpolicy

// pidProvablyDead reports nothing on windows — there is no portable liveness
// probe across the signal interface (card t1630 sync-audit F2'). Windows
// takeover eligibility is therefore governed solely by the lockfile mtime age
// bound (recordLockStaleAge): a lockfile older than the stale bound is
// takeover-eligible regardless of the recorded pid.
func pidProvablyDead(int) bool { return false }
