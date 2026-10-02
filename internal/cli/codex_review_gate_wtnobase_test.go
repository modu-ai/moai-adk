package cli

// SPEC-CODEX-REVIEW-OWNERSHIP-001 M1 — the gate-level preservation line for a
// WT- branch whose merge base cannot be computed (REQ-CRO-004, AC-004 (b)).
//
// Why this file stands alone: the test below is authored WITHOUT the
// tree_scope key and must compile and pass on the pre-change tree, so the
// behaviour it pins is observed GREEN before any policy code exists
// (acceptance.md §C, plan.md §I M1). Before this test, only the RESOLVER half
// of the behaviour was pinned (TestCodexReviewScope_UnidentifiedFallsToTree
// calls reviewScopeResolver and never the gate); nothing asserted that such a
// session is still REVIEWED as a whole tree. The probe and the fixture here are
// shared with the policy tests in codex_review_ownership_test.go.

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// ownershipProbe counts what one gate evaluation touched: the reviewer lookup,
// the reviewable-change detector, the scope log rows and the wire requests of
// the faked codex session.
type ownershipProbe struct {
	sess    *fakeCodexSession
	lookups int
	detects int
	scopes  []reviewScope
}

// newOwnershipProbe installs a fake codex session (review text with a P1 bullet,
// so a review that runs BLOCKS) and wraps the lookup, detector and scope-log
// seams with counters. Every seam is restored on cleanup.
func newOwnershipProbe(t *testing.T) *ownershipProbe {
	t.Helper()
	p := &ownershipProbe{}
	p.sess = withCodexSession(t, codexSessionScript("- [P1] ownership probe finding"))

	look := codexLookPath
	codexLookPath = func(name string) (string, error) { p.lookups++; return look(name) }
	t.Cleanup(func() { codexLookPath = look })

	detect := reviewGateChangeDetector
	reviewGateChangeDetector = func(dir string) bool { p.detects++; return detect(dir) }
	t.Cleanup(func() { reviewGateChangeDetector = detect })

	logger := reviewGateScopeLogger
	reviewGateScopeLogger = func(s reviewScope, _ map[string]string) { p.scopes = append(p.scopes, s) }
	t.Cleanup(func() { reviewGateScopeLogger = logger })
	return p
}

// reviewed reports whether a review request reached the (faked) codex session.
func (p *ownershipProbe) reviewed() bool { return len(p.sess.sent) > 0 }

// request returns the thread/start cwd and the review/start target the session
// received — the gate's assembled review request (thread/start carries cwd,
// review/start carries the target; same wire reading as requireCardRequest).
func (p *ownershipProbe) request(t *testing.T) (cwd string, target map[string]any) {
	t.Helper()
	if len(p.sess.sent) < 3 {
		t.Fatalf("expected >=3 sent requests; got %d (%v)", len(p.sess.sent), p.sess.sent)
	}
	threadParams, _ := sentRequest(t, p.sess.sent[1])["params"].(map[string]any)
	cwd, _ = threadParams["cwd"].(string)
	reviewParams, _ := sentRequest(t, p.sess.sent[2])["params"].(map[string]any)
	target, _ = reviewParams["target"].(map[string]any)
	if target == nil {
		t.Fatalf("review/start target missing: %v", p.sess.sent[2])
	}
	return cwd, target
}

// newWTNoBaseTree builds the "WT- branch without a develop base" session tree:
// a repository whose only branches are main and WT-orphan (no develop ref), with
// one uncommitted reviewable file. The resolver classifies it TREE scope with
// Branch "WT-orphan" and a "merge base unavailable" basis (codex_review_scope.go).
func newWTNoBaseTree(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	cardScopeGit(t, dir, "init", "-q", "-b", "main")
	writeCardFile(t, dir, "x.go", "package x\n")
	cardScopeGit(t, dir, "add", "-A")
	cardScopeGit(t, dir, "commit", "-q", "-m", "init")
	cardScopeGit(t, dir, "checkout", "-q", "-b", "WT-orphan")
	writeCardFile(t, dir, "dirty.go", "package x\n\n// uncommitted reviewable work\n")
	return dir
}

// TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree pins, at the GATE
// level, that a WT- branch session whose merge base is unavailable is reviewed
// as a whole uncommitted tree: the reviewer is consulted once, the request is
// {target: uncommittedChanges, cwd: <tree>}, and the scope log row names the
// unavailable merge base. This is the behaviour the tree_scope policy must
// never swallow (REQ-CRO-004): an unidentifiable card base is not licence to
// stop reviewing.
func TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree(t *testing.T) {
	tree := newWTNoBaseTree(t)
	p := newOwnershipProbe(t)

	out, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "wt-nobase", CWD: tree}, true /* enabled */, tree)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("the probe review fails, so a reviewed turn must BLOCK; got %+v", out)
	}
	if p.lookups != 1 {
		t.Errorf("the reviewer must be consulted exactly once, got %d lookups", p.lookups)
	}
	cwd, target := p.request(t)
	if cwd != tree {
		t.Errorf("thread/start cwd = %q, want the session tree %q", cwd, tree)
	}
	if got, _ := target["type"].(string); got != codexTargetUncommitted {
		t.Errorf("target.type = %q, want %q (whole-tree request)", got, codexTargetUncommitted)
	}
	if len(p.scopes) != 1 {
		t.Fatalf("exactly one scope log row expected, got %d (%+v)", len(p.scopes), p.scopes)
	}
	row := p.scopes[0]
	if row.Class != reviewScopeTree || row.Branch != "WT-orphan" {
		t.Errorf("scope row = %+v, want tree class carrying branch WT-orphan", row)
	}
	if !strings.Contains(row.Basis, "merge base unavailable") {
		t.Errorf("scope basis %q must name the unavailable merge base", row.Basis)
	}
}
