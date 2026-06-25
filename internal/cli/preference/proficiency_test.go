package preference

import (
	"testing"
)

// TestEstimateProficiency_Thresholds verifies the M5 proficiency inference
// (REQ-ADM-017, AC-ADM-017 [S2 Critical], design.md §A.4). Thresholds:
//   - sessionCount >= 20  → Expert   (weak recommendation, info-centric)
//   - 5 <= sessionCount <= 19 → General (strong recommendation)
//   - sessionCount < 5    → ColdStart (neutral — REQ-ADM-014 cold-start gate)
func TestEstimateProficiency_Thresholds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		count   int
		want    Proficiency
	}{
		{"zero sessions", 0, ProficiencyColdStart},
		{"one session", 1, ProficiencyColdStart},
		{"four sessions just below general", 4, ProficiencyColdStart},
		{"five sessions general entry", 5, ProficiencyGeneral},
		{"ten sessions general mid", 10, ProficiencyGeneral},
		{"nineteen sessions general exit", 19, ProficiencyGeneral},
		{"twenty sessions expert entry", 20, ProficiencyExpert},
		{"fifty sessions expert deep", 50, ProficiencyExpert},
		{"one hundred sessions expert saturated", 100, ProficiencyExpert},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := EstimateProficiency(tc.count)
			if got != tc.want {
				t.Errorf("EstimateProficiency(%d) = %v, want %v", tc.count, got, tc.want)
			}
		})
	}
}

// TestEstimateProficiency_NegativeCountIsColdStart verifies a negative session
// count (defensive — a fresh install or corrupted counter) is treated as
// ColdStart rather than panicking or producing an invalid enum value.
func TestEstimateProficiency_NegativeCountIsColdStart(t *testing.T) {
	t.Parallel()
	got := EstimateProficiency(-1)
	if got != ProficiencyColdStart {
		t.Errorf("EstimateProficiency(-1) = %v, want ColdStart (defensive clamp)", got)
	}
}

// TestProficiency_String verifies the string forms are stable for logging /
// audit output. These are NOT user-facing labels; they appear in decision logs
// per AC-ADM-017 ("숙련도 추정 로그").
func TestProficiency_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		p    Proficiency
		want string
	}{
		{ProficiencyColdStart, "cold_start"},
		{ProficiencyGeneral, "general"},
		{ProficiencyExpert, "expert"},
	}
	for _, tc := range cases {
		if got := tc.p.String(); got != tc.want {
			t.Errorf("%v.String() = %q, want %q", tc.p, got, tc.want)
		}
	}
}
