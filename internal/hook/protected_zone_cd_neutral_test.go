package hook

// protected_zone_cd_neutral_test.go — card t1677: the cd-chain generation-
// duplication boundary. Within ONE cd command the walker runs world 0's move
// then world 1's move over the same accumulated set, and the union keeps
// every entry — so a directory reached by both generations stayed twice
// (once tagged gen 0, once gen 1), and a chain of 4 literal cds already
// outgrew zoneCwdsBound (card t1574 K8, REQ-ZSP-006), denying a legitimate
// script fail-closed as loop-unbounded. The repair folds a directory
// present in BOTH generations into the one generation-neutral entry the
// zoneCwd type already defines — a DEDUPLICATION of reachable states, never
// a truncation: the cd-may-fail conservatism (round 8 P1) keeps the
// pre-move set in the union, and the folded set still denies fail-closed
// once a longer chain genuinely outgrows the bound.

import (
	"fmt"
	"testing"
)

// zoneCdNeutralChain builds a chain of n literal relative cds, each into a
// distinct subdirectory (single-operand form), followed by one read-only
// command. The read-only tail keeps mutating false, so a deny on these
// cells is the unbounded verdict alone — the fast path returns before the
// manifest for a completed walk, and an out-of-bound walk denies regardless
// of the tail being read-only (card t1574 K7).
func zoneCdNeutralChain(n int) string {
	cmd := ""
	for i := 1; i <= n; i++ {
		cmd += fmt.Sprintf("cd d%02d; ", i)
	}
	return cmd + "ls zone_dir/a.log"
}

// TestProtectedZoneShellCdChainNeutralBoundary pins both sides of the
// boundary: a chain of 4 literal cds fits the bound once the
// per-generation duplication is folded (16 entries exactly at
// zoneCwdsBound — a legitimate script must pass), a chain of 5 does not and
// still answers the fail-closed verdict, and the n=3 control stays allow
// under either shape (no over-tightening).
func TestProtectedZoneShellCdChainNeutralBoundary(t *testing.T) {
	cases := []struct {
		n        int
		wantDeny bool
	}{
		{n: 3, wantDeny: false},
		{n: 4, wantDeny: false},
		{n: 5, wantDeny: true},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("n=%d", tc.n), func(t *testing.T) {
			root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
			h := zoneTestHandler(t, root)
			cmd := zoneCdNeutralChain(tc.n)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
			if tc.wantDeny {
				wantZoneDeny(t, cmd, d, r, harnessLearnerIdentity, "category", "loop-unbounded")
				return
			}
			// the family's allow convention: the control fails only on a
			// deny — the permission layer may answer the passing call in
			// its own vocabulary (protected_zone_guard_test.go)
			if d == DecisionDeny {
				t.Errorf("%s: decision=%q reason=%q, want allow", cmd, d, r)
			}
		})
	}
}
