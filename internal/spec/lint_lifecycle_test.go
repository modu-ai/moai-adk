package spec

import (
	"strings"
	"testing"
)

// lifecycleTestFrontmatter builds a complete, otherwise-valid 12-field frontmatter so
// that the only variable under test is the lifecycle value. Every other field is
// populated to keep the required-field branch of FrontmatterSchemaRule silent.
func lifecycleTestFrontmatter(lifecycle string) SPECFrontmatter {
	fm := phaseTestFrontmatter("v3.0.2")
	fm.Lifecycle = lifecycle
	return fm
}

// TestLifecycleEnum_NonCanonicalRejected verifies that a lifecycle value outside the
// schema SSOT's enum (spec-anchored | spec-lite | exploratory) emits exactly one
// FrontmatterLifecycleInvalid finding at error severity. The corpus carried five
// `spec-first` rows and one `design-only` row (card t1327) that the missing
// membership check never saw.
func TestLifecycleEnum_NonCanonicalRejected(t *testing.T) {
	cases := []struct {
		name      string
		lifecycle string
	}{
		{"spec-first", "spec-first"},
		{"design-only", "design-only"},
		{"case variant of a canonical value", "Spec-Anchored"},
	}

	rule := &FrontmatterSchemaRule{}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			doc := &SPECDoc{Path: "spec.md", Frontmatter: lifecycleTestFrontmatter(tt.lifecycle)}

			lifecycleFindings := phaseFindingsOf(rule.Check(doc, nil), "FrontmatterLifecycleInvalid")
			if len(lifecycleFindings) != 1 {
				t.Fatalf("lifecycle %q: expected exactly 1 FrontmatterLifecycleInvalid finding, got %d: %v",
					tt.lifecycle, len(lifecycleFindings), lifecycleFindings)
			}
			if lifecycleFindings[0].Severity != SeverityError {
				t.Errorf("lifecycle %q: expected severity %q, got %q",
					tt.lifecycle, SeverityError, lifecycleFindings[0].Severity)
			}
			if lifecycleFindings[0].Advisory {
				t.Errorf("lifecycle %q: finding must not be advisory at emission time", tt.lifecycle)
			}
			if !strings.Contains(lifecycleFindings[0].Message, tt.lifecycle) {
				t.Errorf("lifecycle %q: message must quote the offending value, got %q",
					tt.lifecycle, lifecycleFindings[0].Message)
			}
		})
	}
}

// TestLifecycleEnum_CanonicalAccepted verifies the three canonical enum values pass
// without a lifecycle finding.
func TestLifecycleEnum_CanonicalAccepted(t *testing.T) {
	cases := []string{"spec-anchored", "spec-lite", "exploratory"}

	rule := &FrontmatterSchemaRule{}
	for _, lifecycle := range cases {
		t.Run(lifecycle, func(t *testing.T) {
			doc := &SPECDoc{Path: "spec.md", Frontmatter: lifecycleTestFrontmatter(lifecycle)}

			lifecycleFindings := phaseFindingsOf(rule.Check(doc, nil), "FrontmatterLifecycleInvalid")
			if len(lifecycleFindings) != 0 {
				t.Fatalf("lifecycle %q is canonical: expected 0 FrontmatterLifecycleInvalid findings, got %d: %v",
					lifecycle, len(lifecycleFindings), lifecycleFindings)
			}
		})
	}
}

// TestLifecycleEnum_LegacyCompletedAccepted documents the legacy exception: `completed`
// mirrors the status field into lifecycle across 97 corpus rows from pre-consolidation
// practice, and closed-SPEC frontmatter is immutable by doctrine — so the lint accepts
// the spelling instead of demanding a 97-file sweep outside this card's scope. The
// residue is measured and reported (card t1327 verdict), not silently grandfathered.
func TestLifecycleEnum_LegacyCompletedAccepted(t *testing.T) {
	rule := &FrontmatterSchemaRule{}
	doc := &SPECDoc{Path: "spec.md", Frontmatter: lifecycleTestFrontmatter("completed")}

	lifecycleFindings := phaseFindingsOf(rule.Check(doc, nil), "FrontmatterLifecycleInvalid")
	if len(lifecycleFindings) != 0 {
		t.Fatalf("legacy lifecycle \"completed\": expected 0 FrontmatterLifecycleInvalid findings, got %d: %v",
			len(lifecycleFindings), lifecycleFindings)
	}
}

// TestLifecycleEnum_EmptyEmitsOnlyRequiredFieldFinding verifies the enum check runs
// after the required-field emptiness check, so an empty lifecycle produces the
// existing required-field finding once and no duplicate enum finding.
func TestLifecycleEnum_EmptyEmitsOnlyRequiredFieldFinding(t *testing.T) {
	rule := &FrontmatterSchemaRule{}
	for _, lifecycle := range []string{"", "   "} {
		doc := &SPECDoc{Path: "spec.md", Frontmatter: lifecycleTestFrontmatter(lifecycle)}
		findings := rule.Check(doc, nil)

		if got := len(phaseFindingsOf(findings, "FrontmatterLifecycleInvalid")); got != 0 {
			t.Errorf("empty lifecycle %q: expected 0 FrontmatterLifecycleInvalid findings, got %d", lifecycle, got)
		}

		required := phaseFindingsOf(findings, "FrontmatterInvalid")
		if len(required) != 1 {
			t.Fatalf("empty lifecycle %q: expected exactly 1 FrontmatterInvalid finding, got %d: %v",
				lifecycle, len(required), required)
		}
		if !strings.Contains(required[0].Message, "lifecycle") {
			t.Errorf("empty lifecycle %q: required-field finding should name lifecycle, got %q",
				lifecycle, required[0].Message)
		}
	}
}

// TestLifecycleEnum_NotEraDemotable verifies the new code is absent from the
// era-demotion set, mirroring the FrontmatterPhaseInvalid design decision: the
// membership guard exists to catch authoring mistakes at authoring time, and demoting
// it on grandfather-era SPECs would hide exactly those.
func TestLifecycleEnum_NotEraDemotable(t *testing.T) {
	if eraDemotableCodes["FrontmatterLifecycleInvalid"] {
		t.Error("FrontmatterLifecycleInvalid must NOT be registered in eraDemotableCodes")
	}

	in := []Finding{
		{Code: "FrontmatterLifecycleInvalid", Severity: SeverityError, Message: "m"},
	}
	out := applyEraDemotion(in, demotionCause{GrandfatheredEra: true})

	if out[0].Severity != SeverityError || out[0].Advisory {
		t.Errorf("FrontmatterLifecycleInvalid must survive era demotion as a non-advisory error, got severity=%q advisory=%v",
			out[0].Severity, out[0].Advisory)
	}
}
