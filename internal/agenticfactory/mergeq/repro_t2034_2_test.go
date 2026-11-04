//go:build repro_t2034

package mergeq

// repro_t2034_2_test.go — card t2034 claim 2 (reproduce only, never default-run).
//
// Claim: TrackedChanges runs `git diff HEAD`; if the gate commits a modification
// or checks out another commit inside the check worktree, that diff is clean
// but the landing uses the SHA pinned before the gate (cand.Result), so code
// other than what was checked lands. The tests assert the EXPECTED-CORRECT
// behavior (HEAD movement during the gate refuses the landing and the trunk
// stays), so a FAIL means REPRODUCED. They drive the real Merger and real
// GitCLI against temporary repositories built by the package's own helpers.

import (
	"context"
	"testing"
)

func TestRepro2GateMovesHead(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(t *testing.T, dir string)
		wantFail bool // expected-correct: the landing is refused
		control  bool
	}{
		{"claim: gate commits a modification", func(t *testing.T, d string) {
			write(t, d, "base.txt", "committed by the gate\n")
			git(t, d, "-c", "user.name=g", "-c", "user.email=g@e", "commit", "-aqm", "gate commit")
		}, true, false},
		{"claim: gate checks out another commit", func(t *testing.T, d string) {
			// the trunk tip before the merge (first parent of the merge result) is a commit that is not the merge result
			other := git(t, d, "rev-parse", "HEAD^1")
			git(t, d, "checkout", "-q", "--detach", other)
		}, true, false},
		{"control: gate modifies a tracked file without committing (existing guard)", func(t *testing.T, d string) {
			write(t, d, "base.txt", "dirty\n")
		}, true, true},
		{"control: gate does nothing (must land)", func(t *testing.T, d string) {}, false, true},
	}
	swept := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			swept++
			dir := repo(t)
			cand := candidate(t, dir, "c1", "a.txt", "a\n")
			before := git(t, dir, "rev-parse", "refs/heads/trunk")
			var pinnedHead, finalHead string
			q := &fakeQueue{}
			m := newMerger(t, dir, q, func(_ context.Context, d string) error {
				pinnedHead = git(t, d, "rev-parse", "HEAD")
				tc.mutate(t, d)
				finalHead = git(t, d, "rev-parse", "HEAD")
				return nil
			})
			m.Gate = "repro-gate"
			res, err := m.Land(context.Background(), "t1", cand)
			if err != nil {
				t.Fatal(err)
			}
			after := git(t, dir, "rev-parse", "refs/heads/trunk")
			t.Logf("outcome=%q reason=%q trunk before=%s after=%s gateStartHEAD=%s gateEndHEAD=%s trunkAfterField=%q",
				res.Outcome, res.Reason, before, after, pinnedHead, finalHead, res.TrunkAfter)
			failed := res.Outcome == OutcomeFailed
			if failed != tc.wantFail {
				t.Errorf("refused=%t, want %t (outcome %q)", failed, tc.wantFail, res.Outcome)
			}
			if res.Outcome == OutcomeLanded && finalHead != pinnedHead {
				t.Logf("MEASURED: landed trunk %s is the pre-gate merge result, gate ended on %s", after, finalHead)
			}
		})
	}
	if swept != len(cases) {
		t.Fatalf("swept %d of %d", swept, len(cases))
	}
}
