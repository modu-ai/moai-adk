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
// authoring time. manager-develop turns them GREEN in M2 (Facet 1) and M3
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
// RED on the pre-implementation tree: the runtime prefix list carries only
// state surfaces, so a settings/config-only change reads reviewable.
func TestReviewableFromPorcelainRuntimeConfigOnlyFalse(t *testing.T) {
	for _, tc := range []struct{ name, path string }{
		{"local claude settings", ".claude/settings.json"},
		{"runtime-written settings variant", ".claude/settings.local.json"},
		{"managed config tree", ".moai/config/sections/workflow.yaml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if reviewableFromPorcelain(" M " + tc.path + "\n") {
				t.Errorf("runtime-managed config surface %q must not count as reviewable on the tree path", tc.path)
			}
		})
	}
	if !reviewableFromPorcelain(" M cmd/moai/main.go\n") {
		t.Fatalf("control failed: an ordinary source path must stay reviewable, so a false above would mean the predicate broke, not this criterion")
	}
}

// --- AC-CGSC-008: runtime-drift findings do not block (Facet 2, M3) --------

// runtimeDriftReviewText is a fail-verdict review whose findings pin ONLY
// runtime-managed configuration surfaces — the measured shape of disposition
// records #2 and #4 (settings PATH entry, auto_cleanup drift).
const runtimeDriftReviewText = "- [P1] `.claude/settings.json:13` personal PATH entry drifted\n" +
	"- [P2] `.moai/config/sections/workflow.yaml:65` auto_cleanup local drift"

// TestCodexReviewGateRuntimeDriftFindingsReclassified pins REQ-CGSC-008: when
// EVERY finding of a tree-scope review targets only the runtime-managed
// configuration surfaces, the turn is ALLOWED and the reclassification is
// recorded — those findings are known local drift, not review defects the
// session's lane owns.
//
// RED on the pre-implementation tree: the verdict is fail, so the gate BLOCKs.
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

	out, err := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("findings targeting only runtime-managed config surfaces must not block the turn (reclassify + record), got %+v", out)
	}
}
