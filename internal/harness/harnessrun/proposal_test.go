package harnessrun

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// TestBuildHarnessRunCandidates_MapsFields is AC-HRR-003/004 (M2-b mapping).
//
// Each finding maps to a ProposalCandidate with: pattern_key in the reserved
// harness_run: namespace, confidence carried verbatim (NOT learner.go
// defaultConfidence), suggested_tier carried verbatim, and the producer-specific
// {surface, kind, summary, confidence, suggested_tier} fields carried via the
// Evidence map seam — the same sibling pattern delegationmap.BuildCandidates
// established.
func TestBuildHarnessRunCandidates_MapsFields(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			Surface:       ".claude/commands/harness/release-update.md",
			Kind:          KindFriction,
			Summary:       "release notes sweep requires manual re-entry each run",
			Confidence:    0.75,
			SuggestedTier: harness.TierAutoUpdate.String(),
		},
	}

	candidates := BuildHarnessRunCandidates(findings)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}

	c := candidates[0]

	// pattern_key shape: harness_run:<surface-sha256[:8]>:<kind> (plan §D-D2).
	if !strings.HasPrefix(c.PatternKey, PatternNamespace+":") {
		t.Errorf("pattern_key %q is outside the reserved %s namespace", c.PatternKey, PatternNamespace)
	}
	if !strings.HasSuffix(c.PatternKey, ":"+KindFriction) {
		t.Errorf("pattern_key %q does not end with the finding kind %q", c.PatternKey, KindFriction)
	}

	// Confidence carried verbatim — NOT learner.defaultConfidence (1.0).
	if c.Confidence != 0.75 {
		t.Errorf("confidence: got %v, want 0.75 (carried verbatim from finding)", c.Confidence)
	}

	// Suggested tier carried verbatim.
	if c.Tier != harness.TierAutoUpdate.String() {
		t.Errorf("tier: got %q, want %q", c.Tier, harness.TierAutoUpdate.String())
	}

	// Single run-time observation (harness-run is a single-shot signal).
	if c.ObservationCount != 1 {
		t.Errorf("observation_count: got %d, want 1", c.ObservationCount)
	}

	// DraftID follows the PROPOSAL-<sha256[:8]> convention.
	if !strings.HasPrefix(c.DraftID, "PROPOSAL-") {
		t.Errorf("draft_id %q lacks PROPOSAL- prefix", c.DraftID)
	}

	// Evidence seam carries the producer-specific fields.
	ev := c.Evidence
	for _, key := range []string{"surface", "kind", "summary", "confidence", "suggested_tier", "approval_gate"} {
		if _, ok := ev[key]; !ok {
			t.Errorf("evidence missing key %q", key)
		}
	}
	if got, _ := ev["surface"].(string); got != findings[0].Surface {
		t.Errorf("evidence surface: got %q, want %q", got, findings[0].Surface)
	}
	if got, _ := ev["approval_gate"].(string); !strings.Contains(got, "Tier-4") {
		t.Errorf("evidence approval_gate %q does not mention Tier-4", got)
	}
}

// TestBuildHarnessRunCandidates_EmptyInput is REQ-HRR-003 "no signal" path.
//
// An empty (or nil) findings slice MUST yield an empty (non-nil) candidate
// slice — distinct from a missing field. The orchestrator distinguishes
// "field absent" from "no signal" (REQ-HRR-003).
func TestBuildHarnessRunCandidates_EmptyInput(t *testing.T) {
	t.Parallel()

	for _, in := range [][]Finding{nil, {}} {
		got := BuildHarnessRunCandidates(in)
		if got == nil {
			t.Fatalf("nil returned for empty input; must be non-nil empty slice (field-present/no-signal distinction)")
		}
		if len(got) != 0 {
			t.Fatalf("expected 0 candidates for empty input, got %d", len(got))
		}
	}
}

// TestBuildHarnessRunCandidates_DeterministicIdempotent is the sibling of
// delegationmap.TestAnalyze_DeterministicIdempotent.
//
// Two calls over the same findings produce byte-identical candidate slices,
// including stable pattern_key + draft_id (no wall-clock drift).
func TestBuildHarnessRunCandidates_DeterministicIdempotent(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			Surface:       ".claude/agents/harness/hns-release-update-specialist.md",
			Kind:          KindGap,
			Summary:       "no structured improvement-findings emission step",
			Confidence:    ConservativeConfidenceFloor,
			SuggestedTier: harness.TierRule.String(),
		},
		{
			Surface:       ".claude/workflows/hns-release-update-run.js",
			Kind:          KindDrift,
			Summary:       "return object omits standard findings field",
			Confidence:    0.80,
			SuggestedTier: harness.TierRule.String(),
		},
	}

	first := BuildHarnessRunCandidates(findings)
	second := BuildHarnessRunCandidates(findings)

	if len(first) != len(second) {
		t.Fatalf("candidate count differs across runs: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].PatternKey != second[i].PatternKey {
			t.Errorf("candidate %d pattern_key drifted: %q vs %q", i, first[i].PatternKey, second[i].PatternKey)
		}
		if first[i].DraftID != second[i].DraftID {
			t.Errorf("candidate %d draft_id drifted: %q vs %q", i, first[i].DraftID, second[i].DraftID)
		}
		if first[i].SourceTs != second[i].SourceTs {
			t.Errorf("candidate %d source_ts drifted (non-deterministic clock read)", i)
		}
	}
}

// TestBuildHarnessRunCandidates_DistinctSurfacesDistinctKeys ensures two
// findings that differ only in surface produce distinct pattern_keys (the
// sha256 discriminator prevents collision).
func TestBuildHarnessRunCandidates_DistinctSurfacesDistinctKeys(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{Surface: "a.md", Kind: KindFriction, Summary: "x", Confidence: 0.7, SuggestedTier: harness.TierRule.String()},
		{Surface: "b.md", Kind: KindFriction, Summary: "x", Confidence: 0.7, SuggestedTier: harness.TierRule.String()},
	}
	out := BuildHarnessRunCandidates(findings)
	if out[0].PatternKey == out[1].PatternKey {
		t.Errorf("two distinct surfaces collapsed to one pattern_key %q — discriminator failed", out[0].PatternKey)
	}
}

// TestBuildHarnessRunCandidates_DistinctKindsDistinctKeys ensures two findings
// on the same surface but differing in kind produce distinct pattern_keys.
func TestBuildHarnessRunCandidates_DistinctKindsDistinctKeys(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{Surface: "a.md", Kind: KindFriction, Summary: "x", Confidence: 0.7, SuggestedTier: harness.TierRule.String()},
		{Surface: "a.md", Kind: KindDrift, Summary: "x", Confidence: 0.7, SuggestedTier: harness.TierRule.String()},
	}
	out := BuildHarnessRunCandidates(findings)
	if out[0].PatternKey == out[1].PatternKey {
		t.Errorf("two distinct kinds on one surface collapsed to one pattern_key %q", out[0].PatternKey)
	}
}

// TestPatternNamespace_IsolatedFromMapperSSOT is AC-HRR-010(b) namespace
// isolation live CI guard (residual-risk i closure, plan §I).
//
// harness_run MUST NOT be a member of harness.PatternBearingEventTypes() —
// that SSOT derives the proposalgen regex, so non-membership is what makes the
// existing mapper reject every harness_run: key by construction. This mirrors
// delegationmap/proposal_test.go:54-56 (delegation_map isolation guard).
//
// AP-11: widening the SSOT to make the mapper accept these keys is forbidden.
func TestPatternNamespace_IsolatedFromMapperSSOT(t *testing.T) {
	t.Parallel()

	for _, et := range harness.PatternBearingEventTypes() {
		if string(et) == PatternNamespace {
			t.Errorf("%q was added to PatternBearingEventTypes(); REQ-HRR-007 / AP-11 forbids widening the SSOT", PatternNamespace)
		}
	}
}

// TestPatternKey_RejectedByExistingMapper feeds every emitted harness_run:
// key to the REAL proposalgen.MapPromotions as a maximally-actionable
// promotion (actionable tier, confidence above threshold) and requires zero
// candidates. Running the mapper asserts against the gate that actually
// exists, not a reconstructed copy.
func TestPatternKey_RejectedByExistingMapper(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{Surface: ".claude/x.md", Kind: KindFriction, Summary: "s", Confidence: 0.9, SuggestedTier: harness.TierAutoUpdate.String()},
	}
	candidates := BuildHarnessRunCandidates(findings)
	if len(candidates) == 0 {
		t.Fatal("expected candidates to feed the mapper rejection test")
	}
	for _, c := range candidates {
		promoted := proposalgen.MapPromotions([]harness.Promotion{{
			PatternKey:       c.PatternKey,
			ObservationCount: 100,
			Confidence:       1.0,
			ToTier:           harness.TierAutoUpdate.String(),
			Ts:               time.Now().UTC(),
		}})
		if len(promoted) != 0 {
			t.Errorf("existing mapper accepted %q; the namespace is not isolated", c.PatternKey)
		}
	}
}

// TestNoLearnerDefaultConfidenceReference is REQ-HRR-004 mechanical guard.
//
// The harness-run producer MUST NOT reference learner.go's defaultConfidence
// (1.0) — findings.confidence is a run-time measured/estimated value, and
// reusing the learner's hardcoded 1.0 would disguise an unmeasured signal as
// a measured one (AP-2, verification-claim-integrity §1). This test reads the
// package's own source files and asserts no CODE REFERENCE (AST identifier)
// to the constant exists — prose mentions that explain the prohibition remain
// valuable and are explicitly exempted.
func TestNoLearnerDefaultConfidenceReference(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate package source dir")
	}
	pkgDir := filepath.Dir(thisFile)

	files := []string{"proposal.go", "types.go"}
	combined := strings.Builder{}
	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(pkgDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(body)
		combined.WriteString(src)
		// REQ-HRR-004 / AP-2: the precise guard is a CODE REFERENCE (an AST
		// identifier), not a documentation mention. A raw substring check
		// would false-flag the doc comment that explains the prohibition, and
		// naming the forbidden constant in prose is valuable for the reviewer.
		// AST walking keeps the guard falsifiable: it fires only on an actual
		// identifier use outside comments.
		fset := token.NewFileSet()
		astFile, perr := parser.ParseFile(fset, name, src, parser.ParseComments)
		if perr != nil {
			t.Fatalf("%s did not parse: %v", name, perr)
		}
		if id, pos := findIdent(astFile, "defaultConfidence"); id {
			p := fset.Position(pos)
			t.Errorf("%s:%d references learner.go defaultConfidence as a code identifier — "+
				"REQ-HRR-004 forbids reusing the hardcoded 1.0 (prose mentions are fine; this is an AST hit)",
				name, p.Line)
		}
	}
	// Sanity (package-wide): the constant must still be named in prose somewhere
	// in the package, otherwise it was silently renamed without updating this
	// guard's rationale.
	if !strings.Contains(combined.String(), "defaultConfidence") {
		t.Errorf("package no longer names defaultConfidence even in prose — the REQ-HRR-004 guard may be stale")
	}
}

// findIdent reports whether the file's AST contains a non-comment identifier
// with the given name, returning the first hit position.
func findIdent(f *ast.File, name string) (bool, token.Pos) {
	var hit bool
	var hitPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if hit {
			return false
		}
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if id.Name == name {
			hit = true
			hitPos = id.Pos()
			return false
		}
		return true
	})
	return hit, hitPos
}
