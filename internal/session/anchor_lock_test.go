package session

// anchor_lock_test.go — SPEC-WORKTREE-REAPER-001 M2 acceptance criteria
// AC-WR-009…013 for the shared lock-∪-registry anchor decision.
//
// The liveness probe is swapped through the sessionProcessLiveness seam so the
// UNDETERMINED case is reachable at all: no real platform produces it on
// demand, and it is the case the fail-closed direction exists for.

import (
	"os"
	"testing"
	"time"
)

// swapLiveness replaces the two-valued liveness probe and restores it.
func swapLiveness(t *testing.T, probe func(pid int) (bool, bool)) {
	t.Helper()
	orig := sessionProcessLiveness
	sessionProcessLiveness = probe
	t.Cleanup(func() { sessionProcessLiveness = orig })
}

// t207LockReason is the verbatim shape measured on this repository's locked
// worktrees; the pid is substituted per test.
func t207LockReason(pid int) string {
	return "claude session t207 (pid " + lockItoa(pid) + " start Sun Aug 23 07:26:09 2026)"
}

func lockItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// TestLockAnchor_LivePidAnchored is AC-WR-009: a lock reason naming a live
// process anchors the tree, and the verdict names the lock as the source.
func TestLockAnchor_LivePidAnchored(t *testing.T) {
	lock := LockInfo{Locked: true, Reason: t207LockReason(os.Getpid())}

	v := AnchorDecision(t.TempDir(), lock, time.Now())

	if !v.Anchored {
		t.Fatalf("live locked pid must anchor, got %+v", v)
	}
	if v.Source != AnchorSourceLock {
		t.Fatalf("expected source %q, got %q", AnchorSourceLock, v.Source)
	}
}

// TestLockAnchor_DeadPidNotAnchored is AC-WR-010: a positively-confirmed dead
// pid is the ONE negative the lock source may assert.
func TestLockAnchor_DeadPidNotAnchored(t *testing.T) {
	swapLiveness(t, func(int) (bool, bool) { return false, true })
	lock := LockInfo{Locked: true, Reason: t207LockReason(4242)}

	anchored, detail := lockAnchorVerdict(lock)

	if anchored {
		t.Fatalf("confirmed-dead pid must not anchor by the lock source (detail: %q)", detail)
	}
	// The tree is still not removable — see TestLockAnchor_DeadLockRemovalIsInert.
	v := AnchorDecision(t.TempDir(), lock, time.Now())
	if v.Anchored {
		t.Fatalf("empty registry + dead lock pid must not anchor, got %+v", v)
	}
}

// TestLockAnchor_IndeterminateIsAnchored is AC-WR-011: every shape the lock
// source cannot read positively falls closed to anchored. The fourth case is
// reachable only because the probe is two-valued (design.md §B.5).
func TestLockAnchor_IndeterminateIsAnchored(t *testing.T) {
	cases := []struct {
		name  string
		lock  LockInfo
		probe func(int) (bool, bool)
	}{
		{"no pid token", LockInfo{Locked: true, Reason: "claude session t210"}, nil},
		{"non-integer pid", LockInfo{Locked: true, Reason: "claude session (pid abc)"}, nil},
		{"empty reason", LockInfo{Locked: true}, nil},
		{"probe undetermined", LockInfo{Locked: true, Reason: t207LockReason(4242)},
			func(int) (bool, bool) { return false, false }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.probe != nil {
				swapLiveness(t, tc.probe)
			}
			anchored, detail := lockAnchorVerdict(tc.lock)
			if !anchored {
				t.Fatalf("%s: an unreadable lock must fail closed to anchored (detail: %q)", tc.name, detail)
			}
		})
	}
}

// TestLockAnchor_DeadLockRemovalIsInert is AC-WR-012: a dead lock is NOT an
// anchor, yet the tree is still not removable — git refuses a locked worktree
// regardless of pid liveness, so the sweep must pre-detect it rather than
// attempt a removal that fails on every sweep, forever.
func TestLockAnchor_DeadLockRemovalIsInert(t *testing.T) {
	swapLiveness(t, func(int) (bool, bool) { return false, true })
	lock := LockInfo{Locked: true, Reason: t207LockReason(4242)}

	if anchored, _ := lockAnchorVerdict(lock); anchored {
		t.Fatal("precondition: a confirmed-dead lock must not anchor")
	}
	if !LockRefusesRemoval(lock) {
		t.Fatal("a locked worktree must be pre-detected as refusal-class even with a dead pid")
	}
	if LockRefusesRemoval(LockInfo{}) {
		t.Fatal("an unlocked worktree must not be reported as refusal-class")
	}
}

// TestAnchorDecision_RegistryOnlyPathStillAnchors is AC-WR-013: the registry
// source is UNIONED, never replaced (REQ-WR-010). With no lock line the lock
// source has NO OPINION — not a negative — and the registry still decides.
func TestAnchorDecision_RegistryOnlyPathStillAnchors(t *testing.T) {
	tree := t.TempDir()
	host, _ := os.Hostname()
	writeAnchorRegistry(t, tree, []Entry{anchorTestEntry(host, tree, os.Getpid(), 2*time.Hour)})

	v := AnchorDecision(tree, LockInfo{}, time.Now())

	if !v.Anchored {
		t.Fatalf("a live registry entry must anchor even with no lock, got %+v", v)
	}
	if v.Source != AnchorSourceRegistry {
		t.Fatalf("expected source %q, got %q", AnchorSourceRegistry, v.Source)
	}
}

// TestParseWorktreeLocks_CapturesStoredReason pins design.md §B.3: `locked ` is
// git's own prefix, not part of the stored reason, and a bare `locked` line is
// a lock with an empty reason.
func TestParseWorktreeLocks_CapturesStoredReason(t *testing.T) {
	porcelain := "worktree /repo\nHEAD 0000\nbranch refs/heads/main\n\n" +
		"worktree /repo/.claude/worktrees/t207\nHEAD 1111\nbranch refs/heads/WT-web-live-todo\n" +
		"locked claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)\n\n" +
		"worktree /repo/.claude/worktrees/bare\nHEAD 2222\nbranch refs/heads/WT-bare\nlocked\n"

	locks := ParseWorktreeLocks(porcelain)

	got := locks["/repo/.claude/worktrees/t207"]
	want := "claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)"
	if !got.Locked || got.Reason != want {
		t.Fatalf("stored reason: got %+v, want Locked=true Reason=%q", got, want)
	}
	if bare := locks["/repo/.claude/worktrees/bare"]; !bare.Locked || bare.Reason != "" {
		t.Fatalf("bare locked line: got %+v, want Locked=true Reason=\"\"", bare)
	}
	if _, ok := locks["/repo"]; ok {
		t.Fatal("an unlocked worktree must be absent from the map (no opinion), not present-and-false")
	}
}
