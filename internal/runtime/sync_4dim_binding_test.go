package runtime

import (
	"math"
	"testing"
)

// judgeFinding builds a single judge-shaped finding for the verdict fixtures.
func judgeFinding(dimension, severity string) map[string]any {
	return map[string]any{"dimension": dimension, "severity": severity, "summary": "x", "file": "f.go", "evidence": "cmd+out"}
}

// judgeScore builds a per-dimension score entry for the verdict fixtures.
func judgeScore(dimension string, score float64) map[string]any {
	return map[string]any{"dimension": dimension, "score": score}
}

// cleanPassVerdict returns a happy-path 4-dim verdict: all four dims above floor,
// harmonic mean PASS, not INCOMPLETE, no contested finding.
func cleanPassVerdict() *FourDimVerdict {
	return &FourDimVerdict{
		Verdict: "PASS",
		Scores: []map[string]any{
			judgeScore("Functionality", 0.9),
			judgeScore("Security", 0.9),
			judgeScore("Craft", 0.9),
			judgeScore("Consistency", 0.9),
		},
		ZeroScored: nil,
		Missing:    nil,
		Findings:   nil,
	}
}

// TestBindingOnCleanPath verifies AC-AUDIT-SNAPSHOT-003 (A3 happy path):
// a clean 4-dim PASS verdict (all dims above floor, not INCOMPLETE, no
// contested finding) is BINDING — the orchestrator does NOT spawn the cold
// sync-auditor.
func TestBindingOnCleanPath(t *testing.T) {
	t.Parallel()
	v := cleanPassVerdict()
	binding, reason := v.IsBinding()
	if !binding {
		t.Fatalf("clean PASS verdict MUST be binding (no cold sync-auditor spawn); fallback reason: %q", reason)
	}
	if reason != "" {
		t.Errorf("binding verdict must return empty fallback reason, got %q", reason)
	}
}

// TestFallbackOnIncomplete verifies AC-AUDIT-SNAPSHOT-003b failure mode (a):
// an INCOMPLETE verdict triggers the cold sync-auditor fallback.
func TestFallbackOnIncomplete(t *testing.T) {
	t.Parallel()
	v := cleanPassVerdict()
	v.Verdict = "INCOMPLETE"
	v.Missing = []string{"Security"}
	binding, reason := v.IsBinding()
	if binding {
		t.Fatal("INCOMPLETE verdict MUST NOT be binding — cold sync-auditor fallback required")
	}
	if reason == "" {
		t.Error("fallback reason must explain the INCOMPLETE trigger")
	}
}

// TestFallbackOnZeroScoredDimension verifies AC-AUDIT-SNAPSHOT-003b failure
// mode (b): any must-pass dimension scoring 0 triggers the fallback.
func TestFallbackOnZeroScoredDimension(t *testing.T) {
	t.Parallel()
	v := cleanPassVerdict()
	v.Scores = []map[string]any{
		judgeScore("Functionality", 0.0), // dim-0
		judgeScore("Security", 0.9),
		judgeScore("Craft", 0.9),
		judgeScore("Consistency", 0.9),
	}
	v.ZeroScored = []string{"Functionality"}
	binding, reason := v.IsBinding()
	if binding {
		t.Fatal("a must-pass dimension scoring 0 MUST NOT be binding — fallback required")
	}
	if reason == "" {
		t.Error("fallback reason must explain the dim-0 trigger")
	}
}

// TestFallbackOnContestedCritical verifies AC-AUDIT-SNAPSHOT-003b failure mode
// (c-i): any one judge reporting a finding at `critical` severity triggers the
// fallback (contested-finding predicate i).
func TestFallbackOnContestedCritical(t *testing.T) {
	t.Parallel()
	v := cleanPassVerdict()
	// A single critical finding from any judge on any dimension is enough.
	v.Findings = []map[string]any{
		judgeFinding("Security", "critical"),
	}
	binding, reason := v.IsBinding()
	if binding {
		t.Fatal("a critical-severity finding MUST NOT be binding — contested-finding fallback required (predicate i)")
	}
	if reason == "" {
		t.Error("fallback reason must explain the contested-critical trigger")
	}
}

// TestFallbackOnContestedSeverityConflict verifies AC-AUDIT-SNAPSHOT-003b
// failure mode (c-ii): two or more judges returning conflicting severity
// classifications for the same dimension (e.g. one marks Functionality
// `critical`, another marks it `minor`) triggers the fallback (predicate ii).
func TestFallbackOnContestedSeverityConflict(t *testing.T) {
	t.Parallel()
	v := cleanPassVerdict()
	// Two findings on the SAME dimension with conflicting severities — neither
	// needs to be `critical` on its own; the conflict itself is the trigger.
	v.Findings = []map[string]any{
		judgeFinding("Functionality", "major"),
		judgeFinding("Functionality", "minor"),
	}
	binding, reason := v.IsBinding()
	if binding {
		t.Fatal("conflicting severities on the same dimension MUST NOT be binding — contested-finding fallback required (predicate ii)")
	}
	if reason == "" {
		t.Error("fallback reason must explain the severity-conflict trigger")
	}
}

// TestBindingSurvivesNonContestedFindings verifies that non-contested findings
// (minor-only, no conflict) do NOT trigger the fallback — the happy-path
// binding is preserved when findings are benign.
func TestBindingSurvivesNonContestedFindings(t *testing.T) {
	t.Parallel()
	v := cleanPassVerdict()
	// Two `minor` findings on DIFFERENT dimensions — no critical, no
	// same-dimension conflict. Binding survives.
	v.Findings = []map[string]any{
		judgeFinding("Craft", "minor"),
		judgeFinding("Consistency", "minor"),
	}
	binding, reason := v.IsBinding()
	if !binding {
		t.Fatalf("non-contested minor findings on different dimensions MUST remain binding; fallback reason: %q", reason)
	}
}

// TestBindingRequiresCompleteFourDimensions (card t2009, S10): a top-level PASS
// is binding only when Scores carries exactly the four distinct dimensions the
// workflow judges, each with a finite score in 0..1. A PASS with empty,
// partial, duplicated, unknown-dimension, or non-numeric scores falls back to
// the cold sync-auditor.
func TestBindingRequiresCompleteFourDimensions(t *testing.T) {
	t.Parallel()

	four := func(mut func(s []map[string]any) []map[string]any) *FourDimVerdict {
		v := cleanPassVerdict()
		v.Scores = mut(v.Scores)
		return v
	}
	cases := map[string]*FourDimVerdict{
		"pass-with-empty-scores": four(func([]map[string]any) []map[string]any { return nil }),
		"three-dimensions":       four(func(s []map[string]any) []map[string]any { return s[:3] }),
		"duplicate-dimension": four(func(s []map[string]any) []map[string]any {
			return append(s[:3:3], judgeScore("Craft", 0.9))
		}),
		"five-entries": four(func(s []map[string]any) []map[string]any {
			return append(s[:4:4], judgeScore("Craft", 0.9))
		}),
		"unknown-dimension": four(func(s []map[string]any) []map[string]any {
			s[2] = judgeScore("Style", 0.9)
			return s
		}),
		"string-score": four(func(s []map[string]any) []map[string]any {
			s[1] = map[string]any{"dimension": "Security", "score": "0.9"}
			return s
		}),
		"null-score": four(func(s []map[string]any) []map[string]any {
			s[1] = map[string]any{"dimension": "Security", "score": nil}
			return s
		}),
		"nan-score": four(func(s []map[string]any) []map[string]any {
			s[1] = judgeScore("Security", math.NaN())
			return s
		}),
		"inf-score": four(func(s []map[string]any) []map[string]any {
			s[1] = judgeScore("Security", math.Inf(1))
			return s
		}),
		"score-above-one": four(func(s []map[string]any) []map[string]any {
			s[1] = judgeScore("Security", 1.5)
			return s
		}),
	}
	for name, v := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			binding, reason := v.IsBinding()
			if binding {
				t.Fatalf("an incomplete or malformed PASS verdict MUST NOT be binding (%s)", name)
			}
			if reason == "" {
				t.Errorf("fallback reason must explain the %s trigger", name)
			}
		})
	}

	// Control: the same builder with nothing damaged stays binding.
	if binding, reason := four(func(s []map[string]any) []map[string]any { return s }).IsBinding(); !binding {
		t.Fatalf("an undamaged four-dimension PASS must stay binding; reason: %q", reason)
	}
}
