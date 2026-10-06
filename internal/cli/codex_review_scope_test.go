package cli

// SPEC-CODEX-GATE-SCOPE-001 — the codex review gate's review-target scoping.
//
// File layout (test-first, per acceptance.md §C):
//   - M1 regression line (observed GREEN on the pre-change tree):
//     TestCodexReviewGate_TreeScopeRequestShapeUnchanged pins the non-card
//     session's review request shape-identical to its form before this SPEC
//     (REQ-CGS-003, the SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 line).
//   - M2 RED set (observed FAILING before the implementation): the scope
//     discriminator (AC-CGS-003/004/005/008), the card-scope request
//     (AC-CGS-001), the frozen-projectDir bypass (AC-CGS-006), the scoped
//     self-gate (AC-CGS-007), the scope-bound receipt key (AC-CGS-009/012),
//     the two-path parity (AC-CGS-010), and the scope log (AC-CGS-013).

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// cardScopeGit runs one git subcommand in dir with a pinned identity, so the
// fixture repositories are reproducible and never consult the operator's git
// configuration (same shape as stopFixture.git, but not bound to a fixture).
func cardScopeGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir,
		"-c", "user.name=fx", "-c", "user.email=fx@example.com",
		"-c", "commit.gpgsign=false", "-c", "maintenance.auto=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// --- M1: the tree-scope regression line (AC-CGS-002, GREEN pre-change) ---

// TestCodexReviewGate_TreeScopeRequestShapeUnchanged pins the non-card
// session's assembled review request: a session on a plain branch keeps the
// exact pre-SPEC shape — target uncommittedChanges (the REQ-CRT-006
// serialization) with cwd = the resolved tree. The card-scope work added by
// this SPEC sits BESIDE this shape, never over it (REQ-CGS-003, spec.md §F).
func TestCodexReviewGate_TreeScopeRequestShapeUnchanged(t *testing.T) {
	f := newStopFixture(t) // branch "main" — not a card branch
	f.dirty(t, "tree scope regression")
	// The whole-tree review this shape pin rides now requires the explicit
	// primary_scope restore: the distributed default skips a primary-checkout
	// tree session (REQ-CGSC-002 / REQ-CGSC-004).
	f.write(t, filepath.Join(".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n  codex:\n    review_gate:\n      primary_scope: review\n")

	sess := withCodexSession(t, codexSessionScript("- [P1] tree scope findings"))

	input := &hook.HookInput{SessionID: "sess-tree", CWD: f.root}
	out, err := HandleCodexReviewGate(input, true /* enabled */, f.root)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a fail verdict must BLOCK; got %+v", out)
	}

	if len(sess.sent) < 3 {
		t.Fatalf("expected >=3 sent requests; got %d (%v)", len(sess.sent), sess.sent)
	}
	// cwd rides the thread/start request (openCodexSessionResolved); the target
	// rides review/start. Both are the gate's assembled review-request fields.
	threadReq := sentRequest(t, sess.sent[1])
	if threadReq["method"] != codexMethodThreadStart {
		t.Fatalf("2nd request = %v, want %s", threadReq["method"], codexMethodThreadStart)
	}
	threadParams, _ := threadReq["params"].(map[string]any)
	if got, _ := threadParams["cwd"].(string); got != f.root {
		t.Errorf("thread/start cwd = %q, want the resolved tree %q", got, f.root)
	}
	req := sentRequest(t, sess.sent[2])
	if req["method"] != codexMethodReviewStart {
		t.Fatalf("3rd request = %v, want %s", req["method"], codexMethodReviewStart)
	}
	params, _ := req["params"].(map[string]any)
	if params == nil {
		t.Fatalf("review/start params missing: %v", req)
	}
	target, _ := params["target"].(map[string]any)
	if target == nil {
		t.Fatalf("review/start target must be the coerced target object, got %v", params["target"])
	}
	if got, _ := target["type"].(string); got != codexTargetUncommitted {
		t.Errorf("target.type = %q, want %q (REQ-CRT-006 shape)", got, codexTargetUncommitted)
	}
}

// --- fixtures (acceptance.md §B) ---

// writeCardFile writes rel under dir with parent mkdir, failing the test on
// error.
func writeCardFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// cardScopeFixture is the acceptance.md §B fixture: a primary-role repo on
// develop carrying foreign WIP (file F — no lane owns it), plus a linked card
// worktree on a WT- branch holding one card commit (file A + file B v1) and
// one uncommitted modification (file B v2).
type cardScopeFixture struct {
	primary string
	card    string
	base    string // git merge-base develop HEAD at fixture time
	head    string // the card branch HEAD at fixture time
}

func newCardScopeFixture(t *testing.T) *cardScopeFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	primary := t.TempDir()
	cardScopeGit(t, primary, "init", "-q", "-b", "develop")
	writeCardFile(t, primary, "go.mod", "module example.com/card\n\ngo 1.22\n")
	cardScopeGit(t, primary, "add", "-A")
	cardScopeGit(t, primary, "commit", "-q", "-m", "base")
	// file F: foreign WIP sitting in the primary-role tree (uncommitted).
	writeCardFile(t, primary, "foreign_wip.go", "package main\n\n// foreign WIP: no lane owns this\n")

	card := t.TempDir()
	cardScopeGit(t, primary, "worktree", "add", "-q", "-b", "WT-fixture-card", card)
	writeCardFile(t, card, "card_a.go", "package main\n\n// card commit file A\n")
	writeCardFile(t, card, "card_b.go", "package main\n\n// card file B v1\n")
	cardScopeGit(t, card, "add", "-A")
	cardScopeGit(t, card, "commit", "-q", "-m", "feat(card): A")
	writeCardFile(t, card, "card_b.go", "package main\n\n// card file B v2 (uncommitted)\n")

	f := &cardScopeFixture{primary: primary, card: card}
	f.base = cardScopeGit(t, card, "merge-base", "develop", "HEAD")
	f.head = cardScopeGit(t, card, "rev-parse", "HEAD")
	return f
}

// newEmptyCardWorktree adds a second card worktree whose card diff is EMPTY:
// zero commits beyond develop and a clean tree.
func (f *cardScopeFixture) newEmptyCardWorktree(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	cardScopeGit(t, f.primary, "worktree", "add", "-q", "-b", name, dir)
	return dir
}

func TestCardMergeBaseConfiguredMain(t *testing.T) {
	for _, remoteAhead := range []bool{true, false} {
		t.Run(fmt.Sprintf("remoteAhead=%v", remoteAhead), func(t *testing.T) {
			f := newCardScopeFixture(t)
			local, remote := f.head, f.base
			if remoteAhead {
				local, remote = remote, local
			}
			cardScopeGit(t, f.card, "branch", "main", local)
			cardScopeGit(t, f.card, "update-ref", "refs/remotes/origin/main", remote)
			writeWorktreeBaseBranchConfig(t, f.card, "main")
			got, err := cardMergeBase(f.card)
			if err != nil || got != f.head {
				t.Fatalf("configured main base = %q, %v; want newest ancestor %s", got, err, f.head)
			}
		})
	}
}

// requireCardRequest asserts the wire requests a card-session turn assembled:
// thread/start carries cwd = the card worktree, review/start carries the card
// diff target, and the primary-role tree is referenced NOWHERE (foreign WIP
// stays out of the target, REQ-CGS-005).
func requireCardRequest(t *testing.T, sent []string, f *cardScopeFixture) {
	t.Helper()
	if len(sent) < 3 {
		t.Fatalf("expected >=3 sent requests; got %d (%v)", len(sent), sent)
	}
	threadReq := sentRequest(t, sent[1])
	threadParams, _ := threadReq["params"].(map[string]any)
	if got, _ := threadParams["cwd"].(string); got != f.card {
		t.Errorf("thread/start cwd = %q, want the card worktree %q", got, f.card)
	}
	reviewReq := sentRequest(t, sent[2])
	params, _ := reviewReq["params"].(map[string]any)
	target, _ := params["target"].(map[string]any)
	if target == nil {
		t.Fatalf("review/start target missing: %v", reviewReq)
	}
	if got, _ := target["type"].(string); got != codexTargetBaseBranch {
		t.Errorf("target.type = %q, want %q (the card diff target)", got, codexTargetBaseBranch)
	}
	if got, _ := target["branch"].(string); got != f.base {
		t.Errorf("target.branch = %q, want the recomputed merge base %q", got, f.base)
	}
	raw := sent[1] + sent[2]
	if strings.Contains(raw, f.primary) {
		t.Errorf("the review request must not reference the primary-role tree %q (file F stays out): %s", f.primary, raw)
	}
}

// --- AC-CGS-001: a card session's review goes to the card diff ---

// TestCodexReviewGate_CardScopeRequestIsCardDiff proves the gate scoping for
// a card session: the assembled review request names the card diff
// (baseBranch pinned to the recomputed merge base, cwd = the card worktree),
// and the primary role tree's foreign WIP (file F) is not in the target.
func TestCodexReviewGate_CardScopeRequestIsCardDiff(t *testing.T) {
	f := newCardScopeFixture(t)
	sess := withCodexSession(t, codexSessionScript("- [P1] card findings"))

	input := &hook.HookInput{SessionID: "sess-card", CWD: f.card, ProjectDir: f.primary}
	out, err := HandleCodexReviewGate(input, true /* enabled */, f.primary)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("the fixture review fails, so the turn must BLOCK; got %+v", out)
	}
	requireCardRequest(t, sess.sent, f)
}

// --- AC-CGS-003: stale env cannot change the scope ---

// TestCodexReviewGate_StaleEnvKeepsTreeScope proves a tree-scope session
// carrying a factory env label stays tree-scope: the label value is never a
// decision input (REQ-CGS-004, decision-index Q2 CONFIRMED) — observed both
// at the discriminator and at the assembled request.
func TestCodexReviewGate_StaleEnvKeepsTreeScope(t *testing.T) {
	f := newCardScopeFixture(t)
	// The primary role tree is reviewed here, so the explicit primary_scope
	// restore rides along (the distributed default skips it, REQ-CGSC-002/004).
	writeOwnershipConfigWithPrimary(t, f.primary, "", "review")

	t.Setenv(config.EnvMoaiFactoryWorker, "worker-9-stale")
	s1 := reviewScopeResolver(f.primary)
	if s1.Class != reviewScopeTree {
		t.Fatalf("develop session with a stale env label must be tree-scope, got %+v", s1)
	}
	t.Setenv(config.EnvMoaiFactoryWorker, "orca-lane-77")
	s2 := reviewScopeResolver(f.primary)
	if !reflect.DeepEqual(s1, s2) {
		t.Fatalf("a different label value must not move the scope: %+v vs %+v", s1, s2)
	}

	sess := withCodexSession(t, codexSessionScript("- [P1] tree findings"))
	out, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "s", CWD: f.primary}, true, f.primary)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("the fixture review fails, so the turn must BLOCK; got %+v", out)
	}
	if len(sess.sent) < 3 {
		t.Fatalf("expected >=3 sent requests; got %d", len(sess.sent))
	}
	threadParams, _ := sentRequest(t, sess.sent[1])["params"].(map[string]any)
	if got, _ := threadParams["cwd"].(string); got != f.primary {
		t.Errorf("tree-scope cwd = %q, want the resolved tree %q", got, f.primary)
	}
	params, _ := sentRequest(t, sess.sent[2])["params"].(map[string]any)
	target, _ := params["target"].(map[string]any)
	if got, _ := target["type"].(string); got != codexTargetUncommitted {
		t.Errorf("tree-scope target.type = %q, want %q (REQ-CRT-006 shape)", got, codexTargetUncommitted)
	}
}

// --- AC-CGS-004: the branch alone decides card scope ---

// TestCodexReviewGate_BranchAloneDecidesCardScope proves a WT- branch session
// with NO env at all is card-scope: env absence must not block card detection
// (REQ-CGS-004).
func TestCodexReviewGate_BranchAloneDecidesCardScope(t *testing.T) {
	f := newCardScopeFixture(t)
	t.Setenv(config.EnvMoaiFactoryWorker, "") // env absent

	scope := reviewScopeResolver(f.card)
	if scope.Class != reviewScopeCard {
		t.Fatalf("WT- branch with no env must be card-scope, got %+v", scope)
	}
	if scope.Branch != "WT-fixture-card" {
		t.Errorf("scope.Branch = %q, want WT-fixture-card", scope.Branch)
	}
	if scope.MergeBase != f.base {
		t.Errorf("scope.MergeBase = %q, want the fixture merge base %q", scope.MergeBase, f.base)
	}
}

// --- AC-CGS-005: no coupling to label formats (positive observation) ---

// TestCodexReviewScope_LabelValuesDoNotAffectDecision proves the discriminator
// does not consume label text: the SAME WT- session under two different
// arbitrary env labels resolves to the identical scope. Per acceptance.md §E
// this is a POSITIVE observation (same result for two values), not an absence
// claim over the source.
func TestCodexReviewScope_LabelValuesDoNotAffectDecision(t *testing.T) {
	f := newCardScopeFixture(t)

	t.Setenv(config.EnvMoaiFactoryWorker, "worker-3")
	first := reviewScopeResolver(f.card)

	t.Setenv(config.EnvMoaiFactoryWorker, "lane-orca-77-different-format")
	second := reviewScopeResolver(f.card)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("scope must not depend on the label value: %+v vs %+v", first, second)
	}
	if first.Class != reviewScopeCard {
		t.Fatalf("premise: the fixture session must be card-scope, got %+v", first)
	}
}

// --- AC-CGS-006: the frozen project dir is bypassed ---

// TestCodexReviewGate_FrozenProjectDirBypassed proves the §A.3(b) shape: the
// session sits in its card worktree while both the payload project_dir and the
// frozen CLAUDE_PROJECT_DIR name the primary. The scope resolves from the
// SESSION's working directory tree, and the primary-only change (file F) is
// not reviewed (REQ-CGS-005).
func TestCodexReviewGate_FrozenProjectDirBypassed(t *testing.T) {
	f := newCardScopeFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", f.primary) // the spawn-frozen env arm
	sess := withCodexSession(t, codexSessionScript("- [P1] card findings"))

	input := &hook.HookInput{SessionID: "s", CWD: f.card, ProjectDir: f.primary}
	out, err := HandleCodexReviewGate(input, true, f.primary)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("the fixture review fails, so the turn must BLOCK; got %+v", out)
	}
	requireCardRequest(t, sess.sent, f)
}

// --- AC-CGS-007: an empty card diff never reaches the reviewer ---

// TestCodexReviewGate_EmptyCardDiffSkipsReviewer proves the scoped self-gate:
// a card-scope session whose card diff is empty (0 commits, 0 uncommitted)
// ALLOWs without the reviewer being consulted — even though the frozen
// primary role tree it would have measured pre-SPEC carries foreign WIP
// (REQ-CGS-006: the self-gate measures the same scope the review would).
func TestCodexReviewGate_EmptyCardDiffSkipsReviewer(t *testing.T) {
	f := newCardScopeFixture(t)
	empty := f.newEmptyCardWorktree(t, "WT-fixture-empty")

	withCodexLookPath(t, func(string) (string, error) {
		t.Fatal("an empty card diff must not reach the reviewer")
		return "", nil
	})
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "s", CWD: empty}, true, f.primary)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("an empty card diff must ALLOW, got %+v", out)
	}
	if runner.calls != 0 {
		t.Errorf("the reviewer must not be invoked; got %d calls", runner.calls)
	}
}

// --- AC-CGS-008: unidentified sessions fall to tree scope (Q5 fail-open) ---

// TestCodexReviewScope_UnidentifiedFallsToTree proves the discriminator's
// fail-open direction: a detached HEAD, a plain non-WT branch, a non-git
// directory, and a WT- branch whose merge base cannot be computed all resolve
// to TREE scope — positive-evidence activation: the new narrower behavior
// only turns on with card evidence (decision-index Q5 CONFIRMED).
func TestCodexReviewScope_UnidentifiedFallsToTree(t *testing.T) {
	f := newCardScopeFixture(t)

	detached := f.newEmptyCardWorktree(t, "WT-fixture-detached")
	cardScopeGit(t, detached, "checkout", "-q", "--detach")

	solo := t.TempDir() // WT- branch, but no develop branch exists
	cardScopeGit(t, solo, "init", "-q", "-b", "main")
	writeCardFile(t, solo, "x.go", "package x\n")
	cardScopeGit(t, solo, "add", "-A")
	cardScopeGit(t, solo, "commit", "-q", "-m", "init")
	cardScopeGit(t, solo, "checkout", "-q", "-b", "WT-orphan")

	cases := []struct {
		name      string
		dir       string
		wantBasis string // substring the tree-scope basis must carry
	}{
		{"detached HEAD", detached, ""},
		{"plain develop branch", f.primary, "no card branch"},
		{"WT- branch without a develop base", solo, "merge base"},
		{"non-git directory", t.TempDir(), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := reviewScopeResolver(tc.dir)
			if s.Class != reviewScopeTree {
				t.Fatalf("%s: want tree scope (the documented fail-open), got %+v", tc.name, s)
			}
			if tc.wantBasis != "" && !strings.Contains(s.Basis, tc.wantBasis) {
				t.Errorf("%s: basis %q must name %q", tc.name, s.Basis, tc.wantBasis)
			}
		})
	}
}

// --- AC-CGS-009: a verdict is bound to its own scope's state ---

// TestCodexReviewScope_ReceiptBoundToScopeState proves the receipt binding:
// a card-scope fail receipt (a) re-blocks the SAME card diff state, (b) never
// matches a tree-scope state on the same tree, and (c) is invisible to a
// leader session whose own tree has no recorded receipt (REQ-CGS-007).
func TestCodexReviewScope_ReceiptBoundToScopeState(t *testing.T) {
	ctx := context.Background()
	f := newCardScopeFixture(t)
	fakeCodexVersion(t, "codex-cli 0.0.0-scope")
	withCodexSession(t, codexSessionScript("- [P1] card findings"))
	if _, err := runVerifyCodexReview(t, f.card); err != nil {
		t.Fatalf("produce the card fail receipt: %v", err)
	}

	cardScope := reviewScopeResolver(f.card)
	if cardScope.Class != reviewScopeCard {
		t.Fatalf("premise: fixture must resolve card-scope, got %+v", cardScope)
	}
	cardState, err := cardReviewReceiptState(ctx, cardScope, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}

	// (a) same scope, same diff state → the recorded fail still re-blocks.
	chk := verify.CheckReceipt(verify.LoadReceipt(f.card, cardState), cardState, time.Now(), 0)
	if !chk.Run || chk.Receipt == nil || chk.Receipt.Verdict != codexReviewVerdictFail {
		t.Fatalf("the same card diff state must reuse the recorded fail; got %+v", chk)
	}

	// (b) a tree-scope state on the same tree must NOT match the card receipt.
	treeKey, err := verify.Key(ctx, f.card)
	if err != nil {
		t.Fatal(err)
	}
	treeState := codexReviewReceiptState(ctx, treeKey, "/fake/codex")
	if chk := verify.CheckReceipt(verify.LoadReceipt(f.card, treeState), treeState, time.Now(), 0); chk.Run {
		t.Fatalf("a tree-scope state must never be blocked by a card-scope receipt: %+v", chk)
	}

	// (c) the leader's own tree (primary) sees no receipt at all.
	leaderKey, err := verify.Key(ctx, f.primary)
	if err != nil {
		t.Fatal(err)
	}
	leaderState := codexReviewReceiptState(ctx, leaderKey, "/fake/codex")
	if r := verify.LoadReceipt(f.primary, leaderState); r != nil {
		t.Fatalf("the leader tree must not see the card's receipt: %+v", r)
	}
}

// --- AC-CGS-012: the scope base is never pinned ---

// TestCodexReviewScope_AbsorbedDevelopRecomputesBase proves the merge-base
// discipline inside the gate (gitflow-lane-protocol §8): after the card
// absorbs develop, the scope resolves to the NEW merge base at evaluation
// time, and a receipt recorded before the absorption no longer matches
// (REQ-CGS-002: recomputed at each gate evaluation, never a pinned SHA).
func TestCodexReviewScope_AbsorbedDevelopRecomputesBase(t *testing.T) {
	ctx := context.Background()
	f := newCardScopeFixture(t)
	fakeCodexVersion(t, "codex-cli 0.0.0-scope")

	s1 := reviewScopeResolver(f.card)
	if s1.Class != reviewScopeCard || s1.MergeBase != f.base {
		t.Fatalf("premise: pre-absorption scope must be card at the fixture base, got %+v", s1)
	}
	// the binding a receipt RECORDED before the absorption carries.
	st1, err := cardReviewReceiptState(ctx, s1, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}

	// absorb: develop advances, then the card branch merges it (the §8 shape —
	// a merge INTO develop does not move the base; the card's absorption does).
	writeCardFile(t, f.primary, "absorbed.go", "package main\n\n// absorbed from develop\n")
	cardScopeGit(t, f.primary, "add", "absorbed.go")
	cardScopeGit(t, f.primary, "commit", "-q", "-m", "feat(develop): advance")
	cardScopeGit(t, f.card, "add", "-A")
	cardScopeGit(t, f.card, "commit", "-q", "-m", "feat(card): B")
	cardScopeGit(t, f.card, "merge", "-q", "--no-edit", "develop")

	s2 := reviewScopeResolver(f.card)
	wantBase := cardScopeGit(t, f.card, "merge-base", "develop", "HEAD")
	if s2.MergeBase != wantBase {
		t.Errorf("post-absorption base = %q, want the recomputed %q", s2.MergeBase, wantBase)
	}
	if s2.MergeBase == s1.MergeBase {
		t.Errorf("absorption must move the base; %q stayed pinned", s1.MergeBase)
	}

	// the receipt binding follows: the pre-absorption binding no longer
	// matches the absorbed state — neither its head nor its diff digest.
	st2, err := cardReviewReceiptState(ctx, s2, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	if st1.Head == st2.Head || st1.TreeDigest == st2.TreeDigest {
		t.Errorf("the absorbed state must not satisfy a pre-absorption receipt binding: %+v vs %+v", st1, st2)
	}
}

// --- AC-CGS-010: the two execution paths see the same scope ---

// TestCodexReviewGate_AndProducerSeeSameScope proves REQ-CGS-009: for the same
// session state, the turn-end path and `moai verify codex-review` assemble the
// SAME scope — identical cwd on thread/start and an identical target object on
// review/start. One discriminator serves both (plan §C [HARD]).
func TestCodexReviewGate_AndProducerSeeSameScope(t *testing.T) {
	f := newCardScopeFixture(t)
	fakeCodexVersion(t, "codex-cli 0.0.0-scope")
	sess := withCodexSession(t, codexSessionScript("- [P1] card findings"))

	if _, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "s", CWD: f.card}, true, f.primary); err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if _, err := runVerifyCodexReview(t, f.card); err != nil {
		t.Fatalf("producer error: %v", err)
	}
	if len(sess.sent) < 6 {
		t.Fatalf("expected 6 sent requests (gate 3 + producer 3); got %d (%v)", len(sess.sent), sess.sent)
	}
	gateThread, _ := sentRequest(t, sess.sent[1])["params"].(map[string]any)
	prodThread, _ := sentRequest(t, sess.sent[4])["params"].(map[string]any)
	gateReview, _ := sentRequest(t, sess.sent[2])["params"].(map[string]any)
	prodReview, _ := sentRequest(t, sess.sent[5])["params"].(map[string]any)

	if got, want := gateThread["cwd"], prodThread["cwd"]; got != want || want != f.card {
		t.Errorf("both paths must review the card tree: gate cwd=%v, producer cwd=%v, want %q", got, want, f.card)
	}
	gateTarget, _ := gateReview["target"].(map[string]any)
	prodTarget, _ := prodReview["target"].(map[string]any)
	if !reflect.DeepEqual(gateTarget, prodTarget) {
		t.Errorf("both paths must assemble the same target: gate=%v, producer=%v", gateTarget, prodTarget)
	}
	if got, _ := gateTarget["type"].(string); got != codexTargetBaseBranch {
		t.Errorf("card session target.type = %q on both paths, want %q", got, codexTargetBaseBranch)
	}
	if got, _ := gateTarget["branch"].(string); got != f.base {
		t.Errorf("card session target.branch = %q, want the merge base %q", got, f.base)
	}
}

// --- AC-CGS-013: the scope determination is observable ---

// TestCodexReviewGate_ScopeLogObservability proves REQ-CGS-010: each gate turn
// records its scope class and basis (branch match / none), distinguishable per
// class, with the factory env label carried as context only.
func TestCodexReviewGate_ScopeLogObservability(t *testing.T) {
	f := newCardScopeFixture(t)

	type scopeRow struct {
		class string
		basis string
		env   map[string]string
	}
	var rows []scopeRow
	prev := reviewGateScopeLogger
	reviewGateScopeLogger = func(scope reviewScope, env map[string]string) {
		rows = append(rows, scopeRow{class: scope.Class, basis: scope.Basis, env: env})
	}
	t.Cleanup(func() { reviewGateScopeLogger = prev })

	withCodexSession(t, codexSessionScript("- [P1] findings"))
	t.Setenv(config.EnvMoaiFactoryWorker, "worker-3-obs")

	if _, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "card-turn", CWD: f.card}, true, f.primary); err != nil {
		t.Fatalf("gate error (card turn): %v", err)
	}
	if _, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "tree-turn", CWD: f.primary}, true, f.primary); err != nil {
		t.Fatalf("gate error (tree turn): %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("each gate turn must record its scope; got %d rows (%+v)", len(rows), rows)
	}
	card, tree := rows[0], rows[1]
	if card.class != reviewScopeCard {
		t.Errorf("card turn: class = %q, want %q", card.class, reviewScopeCard)
	}
	if !strings.Contains(card.basis, "branch match") || !strings.Contains(card.basis, "WT-fixture-card") {
		t.Errorf("card turn: basis %q must name the branch match", card.basis)
	}
	if card.env[config.EnvMoaiFactoryWorker] != "worker-3-obs" {
		t.Errorf("card turn: env context must carry the label as context, got %v", card.env)
	}
	if tree.class != reviewScopeTree {
		t.Errorf("tree turn: class = %q, want %q", tree.class, reviewScopeTree)
	}
	if tree.basis == card.basis || strings.Contains(tree.basis, "branch match") {
		t.Errorf("tree turn: basis %q must be distinguishable from the card basis %q", tree.basis, card.basis)
	}
}

// --- AC-CGS-011 (card leg): fail-open survives scoping ---

// TestCodexReviewGate_CardScopeFailOpenOnMissingReviewer proves a missing
// reviewer ALLOWs in card scope too (REQ-CGS-008 / REQ-MCP-012, unchanged).
// Regression line: observed GREEN before the implementation (the missing
// reviewer ALLOW predates this SPEC).
func TestCodexReviewGate_CardScopeFailOpenOnMissingReviewer(t *testing.T) {
	f := newCardScopeFixture(t)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "s", CWD: f.card}, true, f.primary)
	if err != nil {
		t.Fatalf("a missing reviewer must not error the gate; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("missing reviewer in card scope must ALLOW (fail-open), got %+v", out)
	}
	if runner.calls != 0 {
		t.Errorf("no reviewer means no review call; got %d calls", runner.calls)
	}
}

// --- producer card leg (AC-CGS-001 receipt surface) ---

// TestProduceCodexReviewReceipt_CardScopeRecordsCardState proves the receipt
// producer resolves the card scope for a card session state: the receipt binds
// to the card HEAD with the card-scope config, and the tree-scope state of the
// same tree reads nothing (cross-class separation, REQ-CGS-007/009).
func TestProduceCodexReviewReceipt_CardScopeRecordsCardState(t *testing.T) {
	ctx := context.Background()
	f := newCardScopeFixture(t)
	fakeCodexVersion(t, "codex-cli 0.0.0-scope")
	withCodexSession(t, codexSessionScript("- [P1] card findings"))

	got, err := runVerifyCodexReview(t, f.card)
	if err != nil {
		t.Fatalf("producer: %v", err)
	}
	if got["verdict"] != codexReviewVerdictFail {
		t.Fatalf("verdict = %v, want %q", got["verdict"], codexReviewVerdictFail)
	}

	head := cardScopeGit(t, f.card, "rev-parse", "HEAD")
	state, err := cardReviewReceiptState(ctx, reviewScopeResolver(f.card), "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	if state.Head != head {
		t.Errorf("card receipt binds to the card HEAD %q, got %q", head, state.Head)
	}
	r := verify.LoadReceipt(f.card, state)
	if r == nil || r.Verdict != codexReviewVerdictFail {
		t.Fatalf("the card receipt must be recorded and readable: %+v", r)
	}

	treeKey, err := verify.Key(ctx, f.card)
	if err != nil {
		t.Fatal(err)
	}
	if rr := verify.LoadReceipt(f.card, codexReviewReceiptState(ctx, treeKey, "/fake/codex")); rr != nil {
		t.Errorf("the card receipt must not be readable under the tree-scope key: %+v", rr)
	}
}

// --- card self-gate unit: the runtime-prefix exclusion (REQ-CGS-002) ---

// TestCodexReviewScope_CardDetectorRuntimePathsExcluded proves the card-scope
// change detector excludes runtime-managed paths on the uncommitted legs, so
// hook/session state churn in the card worktree never trips the self-gate.
func TestCodexReviewScope_CardDetectorRuntimePathsExcluded(t *testing.T) {
	f := newCardScopeFixture(t)
	scope := reviewScopeResolver(f.card)
	if scope.Class != reviewScopeCard {
		t.Fatalf("premise: fixture must resolve card-scope, got %+v", scope)
	}
	if !hasReviewableCardChanges(scope) {
		t.Fatal("premise: the card diff (A committed + B modified) is reviewable")
	}

	// runtime-managed churn alone must not make an otherwise-clean card
	// diff reviewable.
	empty := f.newEmptyCardWorktree(t, "WT-fixture-runtime")
	writeCardFile(t, empty, filepath.Join(".moai", "state", "drift.json"), "{}\n")
	writeCardFile(t, empty, filepath.Join(".claude", "agent-memory", "note.md"), "x\n")
	runtimeScope := reviewScopeResolver(empty)
	if hasReviewableCardChanges(runtimeScope) {
		t.Errorf("runtime-managed churn must not count as a reviewable card change")
	}
}

// TestCodexReviewScope_CardUntrackedLegCountsNonRuntimeFiles proves the
// uncommitted union's UNTRACKED leg (REQ-CGS-002): a new untracked source file
// in the card worktree is reviewable card work even with zero uncommitted
// tracked changes, and the receipt digest over the same scope moves with it —
// a re-edit of an untracked card file must not read as a stale-free binding.
func TestCodexReviewScope_CardUntrackedLegCountsNonRuntimeFiles(t *testing.T) {
	f := newCardScopeFixture(t)
	wt := f.newEmptyCardWorktree(t, "WT-fixture-untracked")

	scope := reviewScopeResolver(wt)
	if scope.Class != reviewScopeCard {
		t.Fatalf("premise: the new worktree must resolve card-scope, got %+v", scope)
	}
	if hasReviewableCardChanges(scope) {
		t.Fatal("premise: the fresh worktree is clean")
	}

	writeCardFile(t, wt, filepath.Join("newdir", "new_file.go"), "package main\n\n// untracked card work\n")
	if err := os.Symlink(filepath.Join("newdir", "new_file.go"), filepath.Join(wt, "link.go")); err != nil {
		t.Logf("symlink unavailable on this platform, testing the file leg only: %v", err)
	}

	if !hasReviewableCardChanges(scope) {
		t.Fatal("a non-runtime untracked file must count as reviewable card work")
	}

	// the binding moves with the untracked content: re-editing the file
	// changes the digest, so a receipt recorded before the edit goes stale.
	ctx := context.Background()
	fakeCodexVersion(t, "codex-cli 0.0.0-scope")
	before, err := cardReviewReceiptState(ctx, scope, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	writeCardFile(t, wt, filepath.Join("newdir", "new_file.go"), "package main\n\n// untracked card work v2\n")
	after, err := cardReviewReceiptState(ctx, scope, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	if before.TreeDigest == after.TreeDigest {
		t.Fatal("a re-edit of an untracked card file must move the card binding digest")
	}
}
