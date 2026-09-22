package session

// anchor_disposal_guard_ac_rar_008_test.go — t1058 /
// SPEC-SESSION-REGISTRY-READ-ANCHOR-001 AC-RAR-008.
//
// The fourth consumer coordinate of the disposal guard:
// internal/session/anchor_lock.go:111 — AnchorDecision's registry branch,
// which is the union partner of the lock source and the one a change to
// LiveAnchoredSessions (S2) would alter. The three CLI coordinates are
// covered by the sibling fixture in internal/cli/worktree.
//
// [HARD] Both directions. The converse control is load-bearing: a fixture
// that only measures "a live anchored session yields Anchored=true" is
// satisfied by an implementation that anchors everything, which is the
// empty-green shape this SPEC exists to prevent.
//
// Grade, stated plainly: this measures the live-anchored-worktree judgement
// as unchanged at S1 landing and pre-places the regression guard for S2. It
// does NOT verify the criterion under S2 — that behaviour is unverified here.

import (
	"os"
	"testing"
	"time"
)

// TestACRAR008_AnchorDecisionRegistryCoordinate covers
// internal/session/anchor_lock.go:111 in both directions, with the lock
// source deliberately silent (LockInfo{} carries NO opinion) so the verdict
// is attributable to the registry branch alone.
func TestACRAR008_AnchorDecisionRegistryCoordinate(t *testing.T) {
	host, _ := os.Hostname()

	t.Run("live anchored session anchors the tree", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir()) // isolate the caller-registry root
		tree := t.TempDir()
		writeAnchorRegistry(t, tree, []Entry{
			// Live pid, heartbeat two hours old: the real lane shape, where
			// the PID probe and not the timestamp carries the verdict.
			anchorTestEntry(host, tree, os.Getpid(), 2*time.Hour),
		})

		v := AnchorDecision(tree, LockInfo{}, time.Now())

		if !v.Anchored {
			t.Fatalf("a live registry entry must anchor the tree, got %+v", v)
		}
		if v.Source != AnchorSourceRegistry {
			t.Errorf("verdict source = %q, want %q", v.Source, AnchorSourceRegistry)
		}
	})

	t.Run("converse control: a dead entry leaves the tree free", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
		tree := t.TempDir()
		// Present-but-dead: pid 0 is rejected at LiveAnchoredSessions'
		// `e.PID > 0` guard without probing the OS, and the heartbeat is well
		// past DefaultStaleMinutes. A present row is a stronger converse than
		// an absent registry — it shows liveness is evaluated, not presence.
		writeAnchorRegistry(t, tree, []Entry{
			anchorTestEntry(host, tree, 0, 2*time.Hour),
		})

		v := AnchorDecision(tree, LockInfo{}, time.Now())

		if v.Anchored {
			t.Fatalf("no live session is anchored — the tree must be reported free, got %+v", v)
		}
		if v.Source != AnchorSourceNone {
			t.Errorf("verdict source = %q, want %q", v.Source, AnchorSourceNone)
		}
	})
}
