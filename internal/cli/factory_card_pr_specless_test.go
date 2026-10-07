package cli

// factory_card_pr_specless_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-B
// follow-up (card t1453, C5): a card with an empty spec_id reads
// `.moai/reports/<card-id>/verdict.md` as the sync-audit evidence of the
// merge-readiness check that precedes the pull request. The verdict file is the
// mechanised form of closing a SPEC-less card on a verdict; absent or failing,
// the card is refused — nothing is pushed, no pull request is opened. The
// triple's own condition table is TestMergeTripleSpecLessVerdict in
// internal/factorylane; this half pins the delivery edge that consumes it and
// the git-flow surface that must not move.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// ghfSpecLess turns the fixture's card into a SPEC-less one: the card record
// carries no spec_id, and the verdict file (when body is non-empty) sits in the
// card's own worktree.
func ghfSpecLess(t *testing.T, f ghfFixture, verdictBody string) {
	t.Helper()
	db := fcOpen(t, f.root)
	if _, err := db.DB.Exec(`UPDATE cards SET spec_id='' WHERE card_id='t1'`); err != nil {
		t.Fatalf("clear spec id: %v", err)
	}
	_ = db.Close()
	if verdictBody == "" {
		return
	}
	dir := filepath.Join(f.wt, ".moai", "reports", "t1")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "verdict.md"), []byte(verdictBody), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMergeReadinessBeforePRSpecLess(t *testing.T) {
	const verdictPath = ".moai/reports/t1/verdict.md"
	cases := []struct {
		name    string
		verdict func(tip string) string // "" = no verdict file
		passes  bool
		wants   []string // substrings the refusal (err + output) must carry
	}{
		{"valid_pass_bound_to_head", func(tip string) string { return "verdict: PASS\naudited_sha: " + tip + "\n" }, true, nil},
		{"no_verdict_file", func(string) string { return "" }, false,
			[]string{verdictPath, "does not exist", "verdict: PASS", "audited_sha:"}},
		{"verdict_fail", func(tip string) string { return "verdict: FAIL\naudited_sha: " + tip + "\n" }, false,
			[]string{verdictPath, "verdict: FAIL"}},
		{"stale_audited_sha", func(string) string { return "verdict: PASS\naudited_sha: " + strings.Repeat("b", 40) + "\n" }, false,
			[]string{verdictPath, "audited_sha"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := ghfNew(t, ghfOpts{syncStatus: ""})
			ghfSpecLess(t, f, tc.verdict(f.tip))
			d := newGHDouble(t, f)
			out, err := ghfComplete(t)
			c := fcCard(t, f.root, "t1")
			if tc.passes {
				if err != nil {
					t.Fatalf("a SPEC-less card with a PASS verdict bound to HEAD was refused: %v\n%s", err, out)
				}
				if d.count("pr", "create") != 1 || c.State != homestate.CardPROpen {
					t.Fatalf("create calls=%d state=%s", d.count("pr", "create"), c.State)
				}
				return
			}
			if err == nil {
				t.Fatalf("complete passed a SPEC-less card with: %s", tc.name)
			}
			for _, want := range append(tc.wants, "sync-audit") {
				if !strings.Contains(err.Error()+out, want) {
					t.Errorf("the refusal lacks %q:\nerr: %v\nout: %s", want, err, out)
				}
			}
			if len(d.calls) != 0 || f.remoteBranchTip(t) != "" {
				t.Errorf("a refused card still reached gh (%v) or origin (%q)", d.calls, f.remoteBranchTip(t))
			}
			if c.State != homestate.CardMergeReady {
				t.Errorf("a refused card moved to %s", c.State)
			}
		})
	}
}

// git-flow does not move: `merge ready` still requires a SPEC id and never reads
// a verdict file in its place, so a SPEC-less card is refused there exactly as
// before this follow-up. (`factory complete` under git-flow never runs the
// triple at all — TestFactoryCompleteWindowGitFlowUnchanged.)
func TestFactoryMergeReadySpecLessGitFlowUnchanged(t *testing.T) {
	repo, lockRoot, _ := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	writeGitStrategyBody(t, lockRoot, "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n")
	dir := filepath.Join(repo, ".moai", "reports", "t9001")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "verdict.md"), []byte("verdict: PASS\naudited_sha: "+strings.Repeat("c", 40)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", "", "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err == nil || !strings.Contains(err.Error(), "--spec is required") {
		t.Fatalf("merge ready with an empty --spec: err=%v out=%s; want the --spec is required refusal", err, out)
	}
	if lock := readLockRoot(t, lockRoot); lock.Held() {
		t.Errorf("a refused merge ready took the window: %+v", lock)
	}
}
