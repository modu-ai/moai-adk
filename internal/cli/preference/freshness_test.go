package preference

import (
	"strings"
	"testing"
)

// TestFreshnessLabel_Format verifies the M5 data-freshness disclosure string
// (REQ-ADM-015, AC-ADM-015 [S3 Major]). The disclosure MUST take the form
// "based on N-day-old data" so the acceptance grep
// `based on .*-day-old data` matches for every age value, including 0.
func TestFreshnessLabel_Format(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ageDays int
		wantSub string // the substring the label MUST contain
	}{
		{"fresh today", 0, "based on 0-day-old data"},
		{"one day", 1, "based on 1-day-old data"},
		{"one week", 7, "based on 7-day-old data"},
		{"two weeks", 14, "based on 14-day-old data"},
		{"ttl boundary", 28, "based on 28-day-old data"},
		{"stale", 56, "based on 56-day-old data"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FreshnessLabel(tc.ageDays)
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("FreshnessLabel(%d) = %q, want substring %q", tc.ageDays, got, tc.wantSub)
			}
		})
	}
}

// TestFreshnessLabel_NegativeAgeClamped verifies a negative age (clock skew,
// future last_used) is clamped to 0 so the disclosure never reports a negative
// day count, which would break the acceptance grep and read as nonsense.
func TestFreshnessLabel_NegativeAgeClamped(t *testing.T) {
	t.Parallel()
	got := FreshnessLabel(-5)
	if !strings.Contains(got, "based on 0-day-old data") {
		t.Errorf("FreshnessLabel(-5) = %q, want clamp to 0-day-old", got)
	}
}

// TestFreshnessLabel_MatchesAcceptanceGrep runs the literal AC-ADM-015 grep
// pattern against every label to prove the audit evidence would pass.
func TestFreshnessLabel_MatchesAcceptanceGrep(t *testing.T) {
	t.Parallel()
	pattern := "based on "
	pattern2 := "-day-old data"
	for _, age := range []int{0, 1, 7, 14, 28, 56, 100} {
		got := FreshnessLabel(age)
		if !strings.Contains(got, pattern) || !strings.Contains(got, pattern2) {
			t.Errorf("FreshnessLabel(%d) = %q, must contain %q and %q (AC-ADM-015 grep)", age, got, pattern, pattern2)
		}
	}
}
