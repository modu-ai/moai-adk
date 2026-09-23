// anchor_lock_holder.go — read-only accessors over the shared lock-aware
// anchor decision (anchor_lock.go), for callers that must act on WHO holds a
// lock rather than only on whether a tree is anchored: a launcher that writes
// its own lock needs to recognise its own pid, and may replace a lock only
// when the lock's holder is positively dead.
//
// Nothing here re-decides anything. Each accessor delegates to the same
// parseLockPID / lockAnchorVerdict pair AnchorDecision uses, so the three
// callers of AnchorDecision and the launcher cannot disagree about a lock.
package session

import "fmt"

// CodexAnchorLockReason renders the git worktree lock reason `moai codex -w`
// writes. The `pid <n>` token is what the anchor decision reads; the start
// token, when known, lets a reader tell a reused pid from the original holder.
func CodexAnchorLockReason(tree string, pid int, start string) string {
	if start == "" {
		return fmt.Sprintf("moai codex session %s (pid %d)", tree, pid)
	}
	return fmt.Sprintf("moai codex session %s (pid %d start %s)", tree, pid, start)
}

// LockReasonPID extracts the holder pid from a stored lock reason, by the same
// narrow rule the anchor decision applies.
func LockReasonPID(reason string) (int, bool) {
	return parseLockPID(reason)
}

// LockHolderConfirmedDead reports whether lock is present AND its holder was
// positively established dead — the only state in which a lock may be
// replaced. An unreadable reason or an undetermined liveness is not dead.
func LockHolderConfirmedDead(lock LockInfo) bool {
	if !lock.Locked {
		return false
	}
	anchored, _ := lockAnchorVerdict(lock)
	return !anchored
}
