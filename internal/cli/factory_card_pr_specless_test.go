package cli

// factory_card_pr_specless_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-B
// follow-up (card t1453, C5): a card with an empty spec_id reads
// `.moai/reports/<card-id>/verdict.md` as the sync-audit evidence of the
// merge-readiness check that precedes the pull request. The verdict file is the
// mechanised form of closing a SPEC-less card on a verdict; absent or failing,
// the card is refused — nothing is pushed, no pull request is opened. The
// triple's own condition table is TestMergeTripleSpecLessVerdict in
// internal/factorylane; this half pins the delivery edge that consumes it and,
// since GH #1913 (card t1959), the `merge ready` verb reaching the same
// SPEC-less condition (a) — an omitted --spec, or the card id as --spec,
// closes on the verdict file instead of being refused.

import (
	"os"
	"os/exec"
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
			[]string{verdictPath, "verdict FAIL"}},
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
			// card t1930 defect ②: the only gh traffic a refused delivery may
			// produce is the read-only already-merged probe (`pr view`) — never
			// a create or a merge request, and never a push.
			for _, c := range d.calls {
				if len(c) >= 2 && c[1] == "view" {
					continue
				}
				t.Errorf("a refused card reached gh beyond the probe (%v)", c)
			}
			if f.remoteBranchTip(t) != "" {
				t.Errorf("a refused card still reached origin (%q)", f.remoteBranchTip(t))
			}
			if c.State != homestate.CardMergeReady {
				t.Errorf("a refused card moved to %s", c.State)
			}
		})
	}
}

// gitHead reads the fixture repo's HEAD — the commit a bound audited_sha
// must name.
func gitHead(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// GH #1913 (card t1959): `merge ready` accepts a SPEC-less Class B card —
// `--spec` omitted, or the card id passed as `--spec`, keeps SpecDir empty
// and condition (a) closes on the card's verdict file (merge_specless.go).
// The gate itself does not move: a card WITH a SPEC keeps the progress.md
// record as its evidence, and the verdict file is read under the same
// admission rules (PASS label + audited_sha binding).
func TestFactoryMergeReadySpecLessOnVerdictFile(t *testing.T) {
	cases := []struct {
		name string
		spec []string // the --spec argv, if any, after "--card t9001"
	}{
		{"spec_omitted", nil},
		{"spec_is_card_id", []string{"--spec", "t9001"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, lockRoot, _ := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
			writeGitStrategyBody(t, lockRoot, "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n")
			dir := filepath.Join(repo, ".moai", "reports", "t9001")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			body := "verdict: PASS\naudited_sha: " + gitHead(t, repo) + "\n"
			if err := os.WriteFile(filepath.Join(dir, "verdict.md"), []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}

			args := append([]string{"merge", "ready", "--card", "t9001"}, tc.spec...)
			args = append(args, "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
			out, err := runFactoryMerge(t, args...)
			if err != nil || !strings.Contains(out, `"verdict":"cleared"`) {
				t.Fatalf("SPEC-less merge ready (%s): err=%v out=%s; want a cleared verdict", tc.name, err, out)
			}
			lock := readLockRoot(t, lockRoot)
			if !lock.Held() || lock.SessionName != "lane-9" || lock.Card != "t9001" {
				t.Fatalf("window not taken for the SPEC-less card: %+v", lock)
			}
		})
	}
}

// The SPEC-less acceptance is the verdict file, not an open gate: with no
// verdict file the card is refused by condition (a) — the refusal names the
// evidence path and its required machine lines, and no window is taken.
func TestFactoryMergeReadySpecLessWithoutVerdictRefused(t *testing.T) {
	_, lockRoot, _ := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	writeGitStrategyBody(t, lockRoot, "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n")

	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err != nil {
		t.Fatalf("a condition refusal is a verdict, not an error: %v", err)
	}
	for _, want := range []string{".moai/reports/t9001/verdict.md", "verdict: PASS", "sync-audit"} {
		if !strings.Contains(out, want) {
			t.Errorf("the SPEC-less refusal lacks %q: %s", want, out)
		}
	}
	if lock := readLockRoot(t, lockRoot); lock.Held() {
		t.Errorf("a refused SPEC-less card took the window: %+v", lock)
	}
}
