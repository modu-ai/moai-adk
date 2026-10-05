package spec

import (
	"strings"
	"testing"
)

// TestApplyEraDemotionNamesItsCause discharges AC-STV-011 at the unit seam:
// the annotation appended to a demoted finding must name the cause that
// actually fired, rather than naming grandfathered era in both cases.
//
// Before SPEC-STATUS-TRANSITION-VALIDITY-001, a document demoted solely because
// its frontmatter status is terminal was annotated `[grandfathered era —
// downgraded to warning]`, which misstates its own cause (REQ-STV-008).
func TestApplyEraDemotionNamesItsCause(t *testing.T) {
	demotable := []Finding{{Code: "FrontmatterInvalid", Severity: SeverityError, Message: "m"}}

	cases := []struct {
		name         string
		cause        demotionCause
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "grandfathered_era_only",
			cause:        demotionCause{GrandfatheredEra: true},
			wantContains: []string{"grandfathered era", "downgraded to warning"},
			wantAbsent:   []string{"terminal"},
		},
		{
			name:         "terminal_status_only",
			cause:        demotionCause{TerminalStatus: true},
			wantContains: []string{"terminal lifecycle status", "downgraded to warning"},
			wantAbsent:   []string{"grandfathered"},
		},
		{
			name:         "both_causes_are_named",
			cause:        demotionCause{GrandfatheredEra: true, TerminalStatus: true},
			wantContains: []string{"grandfathered era", "terminal lifecycle status"},
		},
	}

	var annotations []string
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := append([]Finding(nil), demotable...)
			out := applyEraDemotion(in, tc.cause)

			if len(out) != 1 {
				t.Fatalf("want 1 finding, got %d", len(out))
			}
			if out[0].Severity != SeverityWarning || !out[0].Advisory {
				t.Errorf("demotion behavior changed: severity=%q advisory=%v", out[0].Severity, out[0].Advisory)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(out[0].Message, want) {
					t.Errorf("annotation %q does not name %q", out[0].Message, want)
				}
			}
			for _, absent := range tc.wantAbsent {
				if strings.Contains(out[0].Message, absent) {
					t.Errorf("annotation %q claims %q, which is not the cause that fired", out[0].Message, absent)
				}
			}
			annotations = append(annotations, out[0].Message)
		})
	}

	// The load-bearing half of AC-STV-011: the two single-cause annotations
	// must actually DIFFER. Two messages that both pass their own substring
	// assertions but read identically would leave the defect in place.
	if len(annotations) >= 2 && annotations[0] == annotations[1] {
		t.Errorf("AC-STV-011: the two causes produce the same annotation %q", annotations[0])
	}
}

// TestApplyEraDemotionNotDemoted pins the no-op path: a document with no
// demotion cause passes through untouched.
func TestApplyEraDemotionNotDemoted(t *testing.T) {
	in := []Finding{{Code: "FrontmatterInvalid", Severity: SeverityError, Message: "m"}}
	out := applyEraDemotion(in, demotionCause{})

	if out[0].Severity != SeverityError || out[0].Advisory || out[0].Message != "m" {
		t.Errorf("undemoted finding was modified: %+v", out[0])
	}
	if (demotionCause{}).demoted() {
		t.Error("an empty demotionCause must not report as demoted")
	}
}

// TestDemotionCauseEndToEnd discharges AC-STV-011 through the linter, on two
// real fixtures: one demoted only by grandfathered era, one demoted only by
// terminal frontmatter status.
func TestDemotionCauseEndToEnd(t *testing.T) {
	// Both fixtures omit the conforming `### Out of Scope —` heading, so each
	// carries a MissingExclusions error — the finding class that receives the
	// annotation (eraDemotableCodes ∩ SeverityError).
	grandfathered := buildTransitionFixture(t, transitionFixture{
		from: "draft", to: "in-progress", trailer: "manager-develop",
		// no progress.md → era V2.x → grandfathered; `in-progress` is not terminal.
	})
	terminalOnly := buildTransitionFixture(t, transitionFixture{
		from: "implemented", to: "completed", trailer: "manager-docs",
		modernEra: true, // → era V3R6, so grandfathering is false; `completed` is terminal.
	})

	annotationFor := func(b builtFixture) string {
		t.Helper()
		report := lintFixture(t, b, false)
		fs := findingsWithCode(report, "MissingExclusions")
		if len(fs) != 1 {
			t.Fatalf("fixture did not produce exactly 1 MissingExclusions finding (got %d) — "+
				"the annotation vehicle this AC needs is absent: %+v", len(fs), fs)
		}
		if !fs[0].Advisory {
			t.Fatalf("fixture was not demoted, so no annotation exists to compare: %+v", fs[0])
		}
		return fs[0].Message
	}

	g := annotationFor(grandfathered)
	term := annotationFor(terminalOnly)

	if !strings.Contains(g, "grandfathered era") {
		t.Errorf("AC-STV-011: grandfathered fixture annotation %q does not name grandfathered era", g)
	}
	if strings.Contains(term, "grandfathered") {
		t.Errorf("AC-STV-011: terminal-status-only fixture claims grandfathered era: %q", term)
	}
	if !strings.Contains(term, "terminal lifecycle status") {
		t.Errorf("AC-STV-011: terminal-status-only fixture annotation %q does not name terminal status", term)
	}
}
