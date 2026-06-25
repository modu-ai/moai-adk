package preference

import (
	"strings"
	"testing"
)

// TestIsSensitiveDomain verifies the M5 sensitive-domain classifier (REQ-ADM-014,
// AC-ADM-014 [S2 Critical]). Sensitive domains (security review, one-off
// exploration queries, cold-start) MUST reduce recommendation strength to
// neutral and disclose the reduction.
func TestIsSensitiveDomain(t *testing.T) {
	t.Parallel()
	sensitive := []string{
		"security-review",
		"security_review",
		"securityreview",
		"one-off-query",
		"one_off_query",
		"cold-start",
		"cold_start",
		"vulnerability-scan",
		"cve-triage",
		"incident-response",
		"exploit-analysis",
	}
	for _, d := range sensitive {
		if !IsSensitiveDomain(d) {
			t.Errorf("IsSensitiveDomain(%q) = false, want true", d)
		}
	}

	nonSensitive := []string{
		"backend-language",
		"log_level",
		"tier-selection",
		"agent-delegation",
		"framework-choice",
		"editor-preference",
		"",
	}
	for _, d := range nonSensitive {
		if IsSensitiveDomain(d) {
			t.Errorf("IsSensitiveDomain(%q) = true, want false", d)
		}
	}
}

// TestIsSensitiveDomain_CaseInsensitive verifies the classifier folds case so
// "Security-Review" and "SECURITY-REVIEW" are caught.
func TestIsSensitiveDomain_CaseInsensitive(t *testing.T) {
	t.Parallel()
	for _, d := range []string{"Security-Review", "SECURITY-REVIEW", "Cold-Start", "One-Off-Query"} {
		if !IsSensitiveDomain(d) {
			t.Errorf("IsSensitiveDomain(%q) = false, want true (case-insensitive)", d)
		}
	}
}

// TestDecideStrength_Matrix verifies the full strength-decision matrix
// (REQ-ADM-014 + REQ-ADM-017 + design.md §A.4 cold-start gate):
//
//	sensitive domain             → Neutral + disclosure
//	non-sensitive + ColdStart    → Neutral (cold-start gate)
//	non-sensitive + General      → Strong
//	non-sensitive + Expert       → Weak
func TestDecideStrength_Matrix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		domain     string
		prof       Proficiency
		coldStart  bool
		wantStr    RecommendationStrength
		wantDisc   string
	}{
		{
			name:      "sensitive domain overrides everything",
			domain:    "security-review",
			prof:       ProficiencyExpert,
			coldStart:  false,
			wantStr:    StrengthNeutral,
			wantDisc:   "personalization reduced for sensitive domain",
		},
		{
			name:      "sensitive domain with cold-start too",
			domain:    "vulnerability-scan",
			prof:       ProficiencyColdStart,
			coldStart:  true,
			wantStr:    StrengthNeutral,
			wantDisc:   "personalization reduced for sensitive domain",
		},
		{
			name:      "non-sensitive general user strong",
			domain:    "backend-language",
			prof:       ProficiencyGeneral,
			coldStart:  false,
			wantStr:    StrengthStrong,
		},
		{
			name:      "non-sensitive expert weak (info-centric)",
			domain:    "backend-language",
			prof:       ProficiencyExpert,
			coldStart:  false,
			wantStr:    StrengthWeak,
		},
		{
			name:      "non-sensitive cold-start neutral (cold-start gate)",
			domain:    "backend-language",
			prof:       ProficiencyColdStart,
			coldStart:  true,
			wantStr:    StrengthNeutral,
		},
		{
			name:      "explicit cold-start flag forces neutral even if prof says general",
			domain:    "log_level",
			prof:       ProficiencyGeneral,
			coldStart:  true,
			wantStr:    StrengthNeutral,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			str, disc := DecideStrength(tc.domain, tc.prof, tc.coldStart)
			if str != tc.wantStr {
				t.Errorf("DecideStrength(%q,%v,%v) strength = %v, want %v",
					tc.domain, tc.prof, tc.coldStart, str, tc.wantStr)
			}
			if tc.wantDisc != "" && !strings.Contains(disc, tc.wantDisc) {
				t.Errorf("DecideStrength(%q,%v,%v) disclosure = %q, want substring %q",
					tc.domain, tc.prof, tc.coldStart, disc, tc.wantDisc)
			}
			// Non-sensitive non-coldStart paths do NOT carry the sensitive-domain
			// disclosure (they may carry a different disclosure or none).
			if tc.wantDisc == "" && strings.Contains(disc, "personalization reduced for sensitive domain") {
				t.Errorf("DecideStrength(%q,%v,%v) unexpectedly carried sensitive-domain disclosure: %q",
					tc.domain, tc.prof, tc.coldStart, disc)
			}
		})
	}
}

// TestRecommendationStrength_String verifies the enum string forms are stable
// for audit logging.
func TestRecommendationStrength_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    RecommendationStrength
		want string
	}{
		{StrengthStrong, "strong"},
		{StrengthWeak, "weak"},
		{StrengthNeutral, "neutral"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("%v.String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}

// TestDecideStrength_SensitiveDisclosureGuarantee verifies EVERY sensitive
// domain produces the disclosure — no false-negative where strength is Neutral
// but the disclosure is missing (which would hide the reduction from the user,
// violating AC-ADM-014's "저하 사실이 공개된다" requirement).
func TestDecideStrength_SensitiveDisclosureGuarantee(t *testing.T) {
	t.Parallel()
	for _, d := range []string{"security-review", "cold-start", "one-off-query", "vulnerability-scan"} {
		str, disc := DecideStrength(d, ProficiencyGeneral, false)
		if str != StrengthNeutral {
			t.Errorf("DecideStrength(%q) strength = %v, want Neutral", d, str)
		}
		if !strings.Contains(disc, "personalization reduced for sensitive domain") {
			t.Errorf("DecideStrength(%q) disclosure = %q, must disclose reduction", d, disc)
		}
	}
}

// TestDecideStrength_UnknownProficiencyIsNeutral verifies the defensive
// fallback: a proficiency value outside the closed enum (should not happen in
// production, but keeps DecideStrength total) collapses to Neutral with the
// cold-start disclosure.
func TestDecideStrength_UnknownProficiencyIsNeutral(t *testing.T) {
	t.Parallel()
	str, disc := DecideStrength("backend", Proficiency(999), false)
	if str != StrengthNeutral {
		t.Errorf("DecideStrength(unknown prof) strength = %v, want Neutral (defensive)", str)
	}
	if disc == "" {
		t.Errorf("DecideStrength(unknown prof) disclosure = empty, want cold-start disclosure")
	}
}

// TestIsSensitiveDomain_TokenBoundaryCases verifies a few boundary tokens so
// the over-inclusive set is exercised beyond the headline cases.
func TestIsSensitiveDomain_TokenBoundaryCases(t *testing.T) {
	t.Parallel()
	for _, d := range []string{"breach-investigation", "pen-test-scope", "PEN_TEST wed", "vulnerable-dependency"} {
		if !IsSensitiveDomain(d) {
			t.Errorf("IsSensitiveDomain(%q) = false, want true (token match)", d)
		}
	}
}
