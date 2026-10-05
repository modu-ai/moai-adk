package graph

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// SPEC-GRAPH-GATE-RESTAMP-001 (card t478) — the codemaps layer measures its
// described-source diff from a CONTENT ANCHOR (the point the codemaps body
// last actually changed), not from the stamped commit. A bare re-stamp — new
// provenance sha, body untouched — therefore no longer resets the measurement
// window to zero.
//
// Every assertion below fixes its expected signal before the measurement runs:
// a specific verdict AND a specific relation on the value, never a bare
// "not fresh" or "err == nil".

// writeCodemapsBody writes the codemaps body docs (NOT provenance.json) into
// the fixture, leaving them for the caller to commit or not. Separate from
// writeCodemapsProvenance, which deliberately leaves the body UNTRACKED (the
// pre-existing fixture shape AC-7 locks).
func writeCodemapsBody(t *testing.T, root, content string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "modules.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// commitDescribedChurn adds n described-worthy files under internal/alpha and
// commits them, so the described-source diff from any earlier anchor is >= n.
func commitDescribedChurn(t *testing.T, root string, n int, tag string) string {
	t.Helper()
	for i := 0; i < n; i++ {
		p := filepath.Join(root, "internal", "alpha", fmt.Sprintf("%s%03d.go", tag, i))
		if err := os.WriteFile(p, []byte("package alpha\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "churn "+tag)
	return gitFix(t, root, "rev-parse", "HEAD")
}

// stampAt writes a clean codemaps provenance sidecar pinned at commit, WITHOUT
// touching the body — the bare re-stamp `moai graph stamp codemaps` performs.
func stampAt(t *testing.T, root, commit string) {
	t.Helper()
	writeCodemapsProvenanceBlock(t, root, &mx.Provenance{
		SchemaVersion: mx.ProvenanceSchemaVersion,
		TreeRoot:      root,
		CommitSHA:     commit,
		GeneratedBy:   "codemaps-gen",
	})
}

// isAncestor reports whether a is an ancestor-or-self of b in root's history.
func isAncestor(t *testing.T, root, a, b string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "merge-base", "--is-ancestor", a, b)
	return cmd.Run() == nil
}

// AC-1 — the mutant direction, and the point of this SPEC: a bare re-stamp
// over an untouched body must stay STALE. Before the repair the same fixture
// reports fresh (value 0), because the window restarts at the new stamp.
func TestCheckCodemaps_BareRestampStaysStale(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)

	// The body is committed and never touched again after this point.
	writeCodemapsBody(t, root, "# modules (generated at base)\n")
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "codemaps body")
	bodyCommit := gitFix(t, root, "rev-parse", "HEAD")

	// Described sources move well past the threshold, and land as commits.
	head := commitDescribedChurn(t, root, th.CodemapsChangedFiles, "gen")

	// The bare re-stamp: provenance re-pinned at HEAD, body untouched.
	stampAt(t, root, head)

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("bare re-stamp must remain judgeable, got system error: %v", err)
	}
	if rep.Verdict != VerdictStale {
		t.Errorf("verdict = %q, want %q — a bare re-stamp over an untouched body must not buy a green gate", rep.Verdict, VerdictStale)
	}
	if rep.Value < rep.Threshold {
		t.Errorf("value = %d, want >= threshold %d — the window must start at the body's last change, not at the stamp", rep.Value, rep.Threshold)
	}
	if rep.ContentAnchorSource != AnchorSourceLastBodyChange {
		t.Errorf("content_anchor_source = %q, want %q", rep.ContentAnchorSource, AnchorSourceLastBodyChange)
	}
	if rep.ContentAnchor != bodyCommit {
		t.Errorf("content_anchor = %q, want the body commit %q", rep.ContentAnchor, bodyCommit)
	}
	if rep.Metric != MetricDescribedSourceDiff {
		t.Errorf("metric = %q, want %q — the token is cited by .github/workflows/graph-freshness.yml and must not change", rep.Metric, MetricDescribedSourceDiff)
	}
}

// AC-2a — genuine regeneration left uncommitted resolves via rule A: the
// anchor is the stamped sha itself, and the layer is fresh.
func TestCheckCodemaps_UncommittedRegenerationIsFresh(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)

	writeCodemapsBody(t, root, "# modules (generated at base)\n")
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "codemaps body")

	head := commitDescribedChurn(t, root, th.CodemapsChangedFiles, "gen")

	// The body IS regenerated — in the working tree only, not committed.
	writeCodemapsBody(t, root, "# modules (regenerated at head)\n")
	stampAt(t, root, head)

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("checkCodemaps: %v", err)
	}
	if rep.Verdict != VerdictFresh {
		t.Errorf("verdict = %q, want %q — a genuine regeneration must still pass", rep.Verdict, VerdictFresh)
	}
	if rep.Value != 0 {
		t.Errorf("value = %d, want 0 — nothing described changed since the stamped head", rep.Value)
	}
	if rep.ContentAnchor != head {
		t.Errorf("content_anchor = %q, want the stamped sha %q", rep.ContentAnchor, head)
	}
	if rep.ContentAnchorSource != AnchorSourceWorkingTreeDiffers {
		t.Errorf("content_anchor_source = %q, want %q", rep.ContentAnchorSource, AnchorSourceWorkingTreeDiffers)
	}
}

// AC-2b — genuine regeneration that was committed resolves via rule B: the
// anchor is the regeneration commit, and the layer is fresh.
func TestCheckCodemaps_CommittedRegenerationIsFresh(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)

	writeCodemapsBody(t, root, "# modules (generated at base)\n")
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "codemaps body")

	commitDescribedChurn(t, root, th.CodemapsChangedFiles, "gen")

	// Regenerate the body AND commit it, then stamp at that point.
	writeCodemapsBody(t, root, "# modules (regenerated after the churn)\n")
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "codemaps body regenerated")
	regen := gitFix(t, root, "rev-parse", "HEAD")
	stampAt(t, root, regen)

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("checkCodemaps: %v", err)
	}
	if rep.Verdict != VerdictFresh {
		t.Errorf("verdict = %q, want %q — a committed regeneration must still pass", rep.Verdict, VerdictFresh)
	}
	if rep.Value != 0 {
		t.Errorf("value = %d, want 0 — nothing described changed since the regeneration commit", rep.Value)
	}
	if rep.ContentAnchor != regen {
		t.Errorf("content_anchor = %q, want the regeneration commit %q", rep.ContentAnchor, regen)
	}
	if rep.ContentAnchorSource != AnchorSourceLastBodyChange {
		t.Errorf("content_anchor_source = %q, want %q", rep.ContentAnchorSource, AnchorSourceLastBodyChange)
	}
}

// AC-3 — a rule-B anchor is an ancestor-or-self of the stamped sha. That
// property is the mechanical ground of the conservatism guarantee (spec.md
// §B.2): restricting the walk to the stamp's own history is what makes the
// anchor move the value redder only, never greener.
func TestCheckCodemaps_RuleBAnchorIsAncestorOfStamp(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)

	writeCodemapsBody(t, root, "# modules (generated at base)\n")
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "codemaps body")

	head := commitDescribedChurn(t, root, th.CodemapsChangedFiles, "gen")
	stampAt(t, root, head)

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("checkCodemaps: %v", err)
	}
	if rep.ContentAnchorSource != AnchorSourceLastBodyChange {
		t.Fatalf("fixture did not resolve via rule B: source = %q", rep.ContentAnchorSource)
	}
	if !isAncestor(t, root, rep.ContentAnchor, head) {
		t.Errorf("git merge-base --is-ancestor %s %s failed — a rule-B anchor outside the stamp's history breaks the conservatism guarantee", rep.ContentAnchor, head)
	}
	// Conservatism, measured rather than asserted: the anchor's diff is at
	// least as large as the stamp's own.
	fromStamp, err := gitDiffNameCount(root, head, mx.DefaultDescribedRoots)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Value < fromStamp {
		t.Errorf("value from anchor = %d < value from stamp = %d — moving the anchor must never make the gate greener", rep.Value, fromStamp)
	}
}

// AC-4 — no anchor-resolution outcome ever yields `fresh` (REQ-GGR-007).
// C1 is the reachable witness: the codemaps directory holds only
// provenance.json, so the body set is empty in the working tree AND at the
// stamp. That is a DETERMINATE OBSERVATION, not a failed measurement — there
// is simply nothing to describe — so it reports absent WITHOUT a system
// error (exit 1), matching how the sibling citations layer already disposes
// of its doc-less state (checkCitations, `docs == 0`).
//
// C2 (bodies exist, no anchor resolves) is deliberately NOT covered by a
// fixture: a shallow boundary commit and a root commit both report every
// file as ADDED, so `git log -1 <S> -- <body>` is non-empty whenever bodies
// exist, and no git state reaches C2. It stays in the code fail-closed. A
// fixture pretending to reach it would be a vacuous green (plan.md §G).
func TestCheckCodemaps_BodyAbsentIsAbsentWithoutError(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)

	head := commitDescribedChurn(t, root, th.CodemapsChangedFiles, "gen")
	stampAt(t, root, head) // creates the dir with provenance.json only

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("an absent body is a determinate observation, not a failed measurement — want a nil error, got: %v", err)
	}
	if rep.Verdict != VerdictAbsent {
		t.Errorf("verdict = %q, want %q", rep.Verdict, VerdictAbsent)
	}
	if rep.Verdict == VerdictFresh {
		t.Error("an unanchorable layer reported fresh — the exact failure REQ-GGR-007 forbids")
	}
	if !strings.Contains(rep.Reason, "codemaps document") {
		t.Errorf("reason = %q, want it to name the body absence", rep.Reason)
	}
	if rep.ContentAnchor != "" || rep.ContentAnchorSource != "" {
		t.Errorf("C1 reported anchor %q/%q, want both empty — no anchor was resolved",
			rep.ContentAnchor, rep.ContentAnchorSource)
	}
}

// AC-7 — regression lock on the pre-existing fixture shape: the codemaps body
// exists in the working tree but was NEVER committed (what
// writeCodemapsProvenance builds). Rule A must fire on the UNTRACKED half of
// its union probe, so the layer stays judgeable. Mutant direction: drop the
// `ls-files --others` term from rule A and this test goes red — without it
// every pre-existing codemaps fixture falls through to rule C (absent).
func TestCheckCodemaps_UntrackedBodyResolvesViaRuleA(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")

	// The pre-existing helper: body written to disk after the base commit,
	// never committed.
	writeCodemapsProvenance(t, root, head)

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("an untracked body must stay judgeable, got system error: %v", err)
	}
	if rep.Verdict == VerdictAbsent {
		t.Errorf("verdict = %q — an untracked body must NOT fall through to rule C; %s", rep.Verdict, rep.Reason)
	}
	if rep.Verdict != VerdictFresh {
		t.Errorf("verdict = %q, want %q — nothing described changed since the stamped head", rep.Verdict, VerdictFresh)
	}
	if rep.ContentAnchor != head {
		t.Errorf("content_anchor = %q, want the stamped sha %q (rule A)", rep.ContentAnchor, head)
	}
	if rep.ContentAnchorSource != AnchorSourceWorkingTreeDiffers {
		t.Errorf("content_anchor_source = %q, want %q", rep.ContentAnchorSource, AnchorSourceWorkingTreeDiffers)
	}
}

// AC-5 boundary — the dirty-fingerprint path resolves NO anchor: it compares
// a content hash against another, so reporting a commit anchor there would
// name a window it never measured. Both fields stay empty (and drop out of
// the JSON via omitempty).
func TestCheckCodemaps_DirtyPathCarriesNoAnchor(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)
	roots := mx.DefaultDescribedRoots
	fp, err := mx.AggregateDescribedFingerprintFiltered(root, roots)
	if err != nil {
		t.Fatal(err)
	}
	writeCodemapsProvenanceBlock(t, root, &mx.Provenance{
		SchemaVersion:      mx.ProvenanceSchemaVersion,
		TreeRoot:           root,
		Dirty:              true,
		ContentFingerprint: fp,
		GeneratedBy:        "codemaps-gen",
	})

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("checkCodemaps: %v", err)
	}
	if rep.Verdict != VerdictFresh {
		t.Fatalf("dirty-path fixture verdict = %q, want %q (unchanged behaviour)", rep.Verdict, VerdictFresh)
	}
	if rep.Metric != MetricGenerationFP {
		t.Errorf("metric = %q, want %q — the dirty path is untouched by this SPEC", rep.Metric, MetricGenerationFP)
	}
	if rep.ContentAnchor != "" || rep.ContentAnchorSource != "" {
		t.Errorf("dirty path reported anchor %q/%q, want both empty — it measures a fingerprint, not a commit window",
			rep.ContentAnchor, rep.ContentAnchorSource)
	}
}
