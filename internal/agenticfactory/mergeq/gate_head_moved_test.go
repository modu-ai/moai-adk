package mergeq

// gate_head_moved_test.go — card t2034 claim 2 regression.
//
// The landing uses the merge result pinned before the gate ran (Candidate.Result). A gate that commits a
// modification or checks out another commit inside the check worktree leaves `git diff HEAD` clean, so the
// tracked-changes guard alone passes while the code the gate judged is not the code that lands. The landing
// must be refused when HEAD is not the pinned merge result once the gate has finished.

import (
	"context"
	"strings"
	"testing"
)

func TestGateMovesHead(t *testing.T) {
	cases := []struct {
		name        string
		mutate      func(t *testing.T, dir string)
		wantOutcome Outcome
		wantReason  []string // substrings the reason must contain
		trunkMoves  bool
	}{
		{"gate commits a modification", func(t *testing.T, d string) {
			write(t, d, "base.txt", "committed by the gate\n")
			git(t, d, "-c", "user.name=g", "-c", "user.email=g@e", "commit", "-aqm", "gate commit")
		}, OutcomeFailed, []string{"moved HEAD off the pinned merge result", "pinned ", "now "}, false},
		{"gate checks out another commit", func(t *testing.T, d string) {
			// the trunk tip before the merge (first parent of the merge result) is a commit that is not the merge result
			git(t, d, "checkout", "-q", "--detach", git(t, d, "rev-parse", "HEAD^1"))
		}, OutcomeFailed, []string{"moved HEAD off the pinned merge result", "pinned ", "now "}, false},
		{"control: gate leaves a dirty tracked file (existing guard)", func(t *testing.T, d string) {
			write(t, d, "base.txt", "dirty\n")
		}, OutcomeFailed, []string{"left tracked changes", "base.txt"}, false},
		{"control: gate does nothing", func(t *testing.T, d string) {}, OutcomeLanded, nil, true},
		{"edge: gate moves HEAD away and back to the pinned result", func(t *testing.T, d string) {
			pinned := git(t, d, "rev-parse", "HEAD")
			git(t, d, "checkout", "-q", "--detach", git(t, d, "rev-parse", "HEAD^1"))
			git(t, d, "checkout", "-q", "--detach", pinned)
		}, OutcomeLanded, nil, true},
	}
	swept := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			swept++
			dir := repo(t)
			cand := candidate(t, dir, "c1", "a.txt", "a\n")
			before := git(t, dir, "rev-parse", "refs/heads/trunk")
			q := &fakeQueue{}
			m := newMerger(t, dir, q, func(_ context.Context, d string) error {
				tc.mutate(t, d)
				return nil
			})
			m.Gate = "head-gate"
			res, err := m.Land(context.Background(), "t1", cand)
			if err != nil {
				t.Fatal(err)
			}
			after := git(t, dir, "rev-parse", "refs/heads/trunk")
			t.Logf("outcome=%q reason=%q", res.Outcome, res.Reason)
			if res.Outcome != tc.wantOutcome {
				t.Fatalf("outcome %q, want %q (reason %q)", res.Outcome, tc.wantOutcome, res.Reason)
			}
			for _, s := range tc.wantReason {
				if !strings.Contains(res.Reason, s) {
					t.Errorf("reason %q lacks %q", res.Reason, s)
				}
			}
			if tc.trunkMoves == (after == before) {
				t.Errorf("trunk before=%s after=%s, trunkMoves want %t", before, after, tc.trunkMoves)
			}
		})
	}
	if swept != len(cases) {
		t.Fatalf("swept %d of %d", swept, len(cases))
	}
}
