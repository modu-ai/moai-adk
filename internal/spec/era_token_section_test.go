// era_token_section_test.go — AC-TA-009 parser-safety regression
// (SPEC-TOKEN-ACCOUNTING-001 M3). Proves that adding a populated
// `## §I Token Accounting` section to a V3R6 progress.md does NOT change
// spec.ClassifyEra() output, because era.go matches only §E.{2,3,4,5} headings
// and the sync_commit_sha / mx_commit_sha fields — never §I.
package spec

import (
	"strings"
	"testing"
)

// v3r6ProgressMD is a minimal V3R6 progress.md fixture (§E.2 + §E.4 + a
// non-empty sync_commit_sha) — the H-4 (new 3-phase predicate) layout.
const v3r6ProgressMD = `# Progress — SPEC-EXAMPLE

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready

## §E.2 Run-phase Evidence

- run evidence line

## §E.3 Run-phase Audit-Ready Signal

_<pending>_

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: a1b2c3d4e5f6789012345678
`

// tokenSectionI is a fully populated `## §I Token Accounting` block as the
// M3 §I writer would emit it. It carries the nine token-accounting fields and
// MUST NOT contain any `§E.N` heading token or the `sync_commit_sha` /
// `mx_commit_sha` field names (those are the only strings era.go greps).
const tokenSectionI = `
## §I Token Accounting

- tokens_spent: 1860
- tokens_input: 300
- tokens_output: 60
- tokens_cache_creation: 0
- tokens_cache_read: 1500
- cache_hit_ratio: 0.8095
- token_attribution: session-set
- token_attribution_confidence: high
- token_session_count: 2
`

// TestEraUnchangedByTokenSection is AC-TA-009 / acceptance.md §D.1 Scenario 4.
// It asserts that ClassifyEra returns the SAME era (V3R6) and a rationale
// still anchored on H-4 before and after a populated §I section is appended.
func TestEraUnchangedByTokenSection(t *testing.T) {
	t.Parallel()

	before := EraSignals{
		ProgressMDExists:  true,
		ProgressMDContent: v3r6ProgressMD,
	}
	after := EraSignals{
		ProgressMDExists:  true,
		ProgressMDContent: v3r6ProgressMD + tokenSectionI,
	}

	// Sanity: the §I block actually landed in the "after" content and does
	// not accidentally carry a parser token.
	if !strings.Contains(after.ProgressMDContent, "## §I Token Accounting") {
		t.Fatal("test fixture invariant: §I heading absent from after-content")
	}

	beforeEra, beforeRule := ClassifyEra(before)
	afterEra, afterRule := ClassifyEra(after)

	// Primary AC-TA-009 assertion: era classification is invariant.
	if beforeEra != afterEra {
		t.Fatalf("AC-TA-009 violation: era changed across §I addition\n  before: %q (%s)\n  after:  %q (%s)",
			beforeEra, beforeRule, afterEra, afterRule)
	}

	// Both classifications must be V3R6 (the modern era) — otherwise the
	// fixture itself is wrong and the invariance assertion above is vacuous.
	if beforeEra != EraV3R6 {
		t.Fatalf("fixture invariant: before-era want V3R6, got %q (%s)", beforeEra, beforeRule)
	}
	if afterEra != EraV3R6 {
		t.Fatalf("fixture invariant: after-era want V3R6, got %q (%s)", afterEra, afterRule)
	}

	// Rationale must still reference H-4 (the §E.2 + §E.4 + sync_commit_sha
	// predicate) — §I must NOT introduce a new classification path.
	if !strings.Contains(beforeRule, "H-4") {
		t.Errorf("before rationale should reference H-4, got %q", beforeRule)
	}
	if !strings.Contains(afterRule, "H-4") {
		t.Errorf("after rationale should reference H-4, got %q", afterRule)
	}
}
