package cli

// SPEC-CODEX-GATE-SCOPING-001 — plan-phase RED observations for the
// primary-checkout scoping policy (Facets 1-2).
//
// These tests are authored BEFORE the implementation exists, on the measured
// pre-implementation tree (HEAD 2de0a2cb613b04765a1554f86685a3b48e0be806, branch
// WT-codex-gate-scope), and are observed RED there as the acceptance.md two-cell
// adoption discipline requires (verification-completeness §2.1). Each AC's
// RED-now cell in .moai/specs/SPEC-CODEX-GATE-SCOPING-001/acceptance.md cites
// one of these tests by name and carries the verbatim failing output recorded at
// authoring time. Revised at HEAD 9ef1cbedc per plan-audit iteration-2 (D8:
// AC-001 asserts the skip row's content, AC-008 asserts the reclassification
// row on the diagnostic channel, AC-007 drops the settings.local.json surface
// as outside the REQ) and re-observed RED on the revised file.
// manager-develop turns them GREEN in M2 (Facet 1) and M3
// (Facet 2); a rename forced by the implementation is reported through the E1
// correction note, never applied silently.
//
// Every assertion reads what the gate DID (reviewer lookups, detector calls,
// skip rows, the decision) — never a reviewer verdict value (acceptance.md §A).
// The PRESERVE (regression-guard) half of this SPEC — card scope, linked
// worktrees, the shared prefix list, existing sync-gate collection — is
// intentionally NOT here: those criteria are green by construction on the
// pre-change tree and are adopted as regression-guards, not RED-first gates
// (acceptance.md §D.1).

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// --- AC-CGSC-001 / AC-CGSC-005: the primary-checkout skip (Facet 1, M2) ----

// primaryScopeFixture builds the shared premise for both Facet-1 reds: a real
// primary checkout (the cardScopeFixture's primary-role repo, git dir ==
// git common dir, branch develop) whose tree_scope is EXPLICITLY review, so
// the existing tree_scope skip policy is inert and the only missing behavior is
// the primary-scope axis under test.
func primaryScopeFixture(t *testing.T) string {
	t.Helper()
	root := newCardScopeFixture(t).primary
	writeOwnershipConfig(t, root, "review")
	return root
}

// TestCodexReviewGatePrimaryCheckoutSkip pins REQ-CGSC-002 on the Claude path:
// a tree-class session sitting in the PRIMARY checkout is allowed before the
// self-gate, the reviewer lookup and the review — one skip row whose basis is
// distinguishable from the tree_scope row, which stays silent.
//
// RED on the pre-implementation tree: no primary determination exists, so the
// gate falls through to the self-gate and reviews the whole tree — the exact
// measured harm of the primary-side gate blocks (card t1395 dispositions).
func TestCodexReviewGatePrimaryCheckoutSkip(t *testing.T) {
	root := primaryScopeFixture(t)

	p := newOwnershipProbe(t)
	skips := captureTreeScopeSkips(t)
	out := gatePath(t, root, root)

	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("a primary-checkout tree session must ALLOW without a review, got %+v", out)
	}
	if p.lookups != 0 || p.detects != 0 || p.reviewed() {
		t.Errorf("the primary skip must precede the self-gate, the lookup and the review; got lookups=%d detects=%d reviewed=%v",
			p.lookups, p.detects, p.reviewed())
	}
	if len(*skips) != 1 {
		t.Errorf("exactly one primary-skip row expected (basis distinguishable from tree_scope); got %d tree_scope rows", len(*skips))
		return
	}
	// Content (plan-audit iteration-2 D8-①): the row is more than an existence
	// count — REQ-CGSC-011 requires the row to carry the resolved tree and the
	// policy axis, and REQ-CGSC-002 requires a basis distinguishable from a
	// plain tree_scope row (whose basis reads "no card branch: …").
	row := (*skips)[0]
	if row.Dir != root {
		t.Errorf("the primary-skip row must carry the resolved tree, got dir %q want %q", row.Dir, root)
	}
	if !strings.Contains(row.Basis, "primary") || !strings.Contains(row.Basis, root) {
		t.Errorf("the primary-skip basis must name the primary reason and the tree path (%s), not a plain tree_scope basis; got %q", root, row.Basis)
	}
}

// TestCodexReviewGatePrimaryPolicySharedByBothPaths pins REQ-CGSC-003: for the
// SAME session state (the primary checkout, tree_scope review), BOTH automatic
// paths — the Claude Stop hook (HandleCodexReviewGate) and Codex Stop-chain
// member 6 (codexReviewMember) — apply the one primary-scope policy and skip.
//
// RED on the pre-implementation tree: neither path knows the primary axis, so
// the gate reviews and the chain takes the receipt route.
func TestCodexReviewGatePrimaryPolicySharedByBothPaths(t *testing.T) {
	root := primaryScopeFixture(t)

	p := newOwnershipProbe(t)
	skips := captureTreeScopeSkips(t)
	gatePath(t, root, root)
	gateSkipped := len(*skips) == 1 && !p.reviewed()

	got := chainPath(t, root)
	if !gateSkipped || !chainSkipped(got) {
		t.Errorf("both automatic paths must skip a primary-checkout tree session; gate skipped=%v (rows %d), chain skipped=%v (outcome %+v)",
			gateSkipped, len(*skips), chainSkipped(got), got)
	}
}

// --- AC-CGSC-007: the tree-only config-surface exclusion (Facet 2, M3) -----

// TestReviewableFromPorcelainRuntimeConfigOnlyFalse pins REQ-CGSC-007: a tree
// session whose only uncommitted changes are the runtime-managed CONFIGURATION
// surfaces (.claude/settings.json, .moai/config/) carries nothing reviewable —
// the self-gate must not fire. The ordinary-source control row kills the mutant
// where the predicate stops counting anything at all.
//
// The table names exactly the two surfaces REQ-CGSC-007 enumerates; the
// `.claude/settings.local.json` row the first draft carried was removed at
// plan-audit iteration-2 (D8-④ scope reduction — the file is outside the REQ's
// enumeration, so requiring it here demanded more than the REQ grants).
//
// RED on the pre-implementation tree: the runtime prefix list carries only
// state surfaces, so a settings/config-only change reads reviewable.
//
// Card-review repair R1 relocated WHERE the exclusion lives: out of the SHARED
// reviewableFromPorcelain (whose multi-review consumers must keep the baseline
// detector) into the tree-only config-only probe the scoped self-gate composes
// with. This AC's observable is unchanged — the assertions below target that
// probe — and the rename legs (repair R5) pin that only a rename on BOTH
// config sides reads config-only, so `.claude/settings.json -> main.go` keeps
// the self-gate firing.
func TestReviewableFromPorcelainRuntimeConfigOnlyFalse(t *testing.T) {
	for _, tc := range []struct{ name, path string }{
		{"local claude settings", ".claude/settings.json"},
		{"managed config tree", ".moai/config/sections/workflow.yaml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !treeConfigOnlyFromPorcelain(" M " + tc.path + "\n") {
				t.Errorf("runtime-managed config surface %q must read config-only, so the tree self-gate does not count it reviewable", tc.path)
			}
		})
	}
	if treeConfigOnlyFromPorcelain(" M cmd/moai/main.go\n") {
		t.Fatalf("control failed: an ordinary source path must keep the tree self-gate firing, so a true above would mean the probe broke, not this criterion")
	}
	if treeConfigOnlyFromPorcelain("R  .claude/settings.json -> cmd/moai/main.go\n") {
		t.Errorf("a rename whose destination is an ordinary source must keep the tree self-gate firing (repair R5)")
	}
	if !treeConfigOnlyFromPorcelain("R  .moai/config/a.yaml -> .moai/config/b.yaml\n") {
		t.Errorf("a rename on both config sides stays config-only")
	}
}

// --- AC-CGSC-008: runtime-drift findings do not block (Facet 2, M3) --------

// runtimeDriftReviewText is a fail-verdict review whose findings pin ONLY
// runtime-managed configuration surfaces — the measured shape of disposition
// records #2 and #4 (settings PATH entry, auto_cleanup drift).
const runtimeDriftReviewText = "- [P1] `.claude/settings.json:13` personal PATH entry drifted\n" +
	"- [P2] `.moai/config/sections/workflow.yaml:65` auto_cleanup local drift"

// captureGateDiagnostics swaps os.Stderr for a pipe for one gate call and
// returns the restore-and-read closure: call it after the gate returns to put
// stderr back and get everything the gate wrote to its diagnostic channel.
// REQ-CGSC-011's surface IS the channel — structured rows ride stderr while
// stdout stays the pure HookOutput contract — so the capture binds the row the
// REQ names rather than any emission seam the implementation may or may not
// add (plan-audit iteration-2 D8-②).
func captureGateDiagnostics(t *testing.T) func() string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = orig
		_ = w.Close()
		_ = r.Close()
	})
	return func() string {
		os.Stderr = orig
		if err := w.Close(); err != nil {
			t.Fatalf("close stderr capture: %v", err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("read stderr capture: %v", err)
		}
		return string(b)
	}
}

// TestCodexReviewGateRuntimeDriftFindingsReclassified pins REQ-CGSC-008: when
// EVERY finding of a tree-scope review targets only the runtime-managed
// configuration surfaces, the turn is ALLOWED and the reclassification is
// recorded — those findings are known local drift, not review defects the
// session's lane owns.
//
// RED on the pre-implementation tree: the verdict is fail, so the gate BLOCKs,
// and no reclassification row exists to find on the diagnostic channel.
func TestCodexReviewGateRuntimeDriftFindingsReclassified(t *testing.T) {
	// Premise: the fixture synthesizes a FAIL verdict whose two findings carry
	// exactly the config-surface files — a parser change that breaks this must
	// fail here, never as a vacuous green below.
	prem := synthesizeReviewOutput(runtimeDriftReviewText, codexMethodReviewStart)
	if prem.Verdict != "fail" || len(prem.Findings) != 2 ||
		prem.Findings[0].File != ".claude/settings.json" ||
		prem.Findings[1].File != ".moai/config/sections/workflow.yaml" {
		t.Fatalf("premise: the fixture must synthesize fail with two findings pinned to the config surfaces, got verdict %q findings %+v",
			prem.Verdict, prem.Findings)
	}

	withChangeDetector(t, true)
	withCodexSession(t, codexSessionScript(runtimeDriftReviewText))

	read := captureGateDiagnostics(t)
	out, err := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	diagnostics := read()
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("findings targeting only runtime-managed config surfaces must not block the turn (reclassify + record), got %+v", out)
	}
	// The row (REQ-CGSC-011): a silent allow is the mutant this kills — the
	// diagnostic channel must carry the reclassification reason in the SPEC's
	// own words ("runtime-managed drift") and name a targeted path.
	if !strings.Contains(diagnostics, "runtime-managed") || !strings.Contains(diagnostics, ".claude/settings.json") {
		t.Errorf("the reclassification must be recorded as a structured row on the diagnostic channel, naming the runtime-managed drift reason and a targeted path; diagnostics: %q", diagnostics)
	}
}
