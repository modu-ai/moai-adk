package session

import "testing"

// TestLockHolderAccessorsAgreeWithAnchorDecision — the accessors are views
// over the anchor decision, never a second decision: a lock the decision
// treats as anchored is never "confirmed dead", and the pid the reason names
// round-trips through the Codex reason renderer.
func TestLockHolderAccessorsAgreeWithAnchorDecision(t *testing.T) {
	prev := sessionProcessLiveness
	t.Cleanup(func() { sessionProcessLiveness = prev })

	reason := CodexAnchorLockReason("card-tree", 4242, "Mon Sep 21")
	if pid, ok := LockReasonPID(reason); !ok || pid != 4242 {
		t.Fatalf("LockReasonPID(%q) = (%d, %v), want (4242, true)", reason, pid, ok)
	}
	if got := CodexAnchorLockReason("card-tree", 7, ""); got != "moai codex session card-tree (pid 7)" {
		t.Errorf("reason without start = %q", got)
	}

	cases := []struct {
		name       string
		lock       LockInfo
		alive, det bool
		wantDead   bool
	}{
		{"unlocked", LockInfo{}, false, true, false},
		{"dead holder", LockInfo{Locked: true, Reason: reason}, false, true, true},
		{"live holder", LockInfo{Locked: true, Reason: reason}, true, true, false},
		{"undetermined holder", LockInfo{Locked: true, Reason: reason}, false, false, false},
		{"unreadable reason", LockInfo{Locked: true, Reason: "held by an operator"}, false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionProcessLiveness = func(int) (bool, bool) { return tc.alive, tc.det }
			if got := LockHolderConfirmedDead(tc.lock); got != tc.wantDead {
				t.Errorf("LockHolderConfirmedDead = %v, want %v", got, tc.wantDead)
			}
			if anchored, _ := lockAnchorVerdict(tc.lock); tc.lock.Locked && anchored == tc.wantDead {
				t.Errorf("accessor disagrees with the anchor verdict (anchored=%v)", anchored)
			}
		})
	}
}
