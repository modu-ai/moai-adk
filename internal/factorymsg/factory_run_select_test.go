package factorymsg

import (
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// SPEC-LANE-JOIN-RUN-SELECT-001 REQ-LJR-002 (card t1853): run times render in
// KST (UTC+9) wherever the lane chooser and the ambiguity refusal show them.

// AC-LJR-002: UTC and offset inputs converge on one KST rendering, and a value
// that is not an RFC 3339 time comes back unchanged. The test pins time.Local
// to UTC for its duration, so an implementation that formats in the host zone
// instead of a fixed KST zone fails here on every host.
func TestFormatKST(t *testing.T) {
	prev := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = prev })

	cases := []struct{ name, in, want string }{
		{"utc-nanos", "2026-10-17T09:31:00.123456789Z", "2026-10-17 18:31:00 KST"},
		{"offset-input", "2026-10-17T18:31:00+09:00", "2026-10-17 18:31:00 KST"},
		{"utc-no-fraction", "2026-10-17T09:31:00Z", "2026-10-17 18:31:00 KST"},
		{"day-rollover", "2026-10-17T20:00:00Z", "2026-10-18 05:00:00 KST"},
		{"unparseable-unchanged", "not-a-time", "not-a-time"},
		{"empty-unchanged", "", ""},
	}
	for _, c := range cases {
		if got := FormatKST(c.in); got != c.want {
			t.Errorf("%s: FormatKST(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// AC-LJR-002: the AMBIGUOUS_FACTORY refusal text renders each start in KST
// and no longer carries the raw UTC clock reading.
func TestAmbiguousRunsErrorRendersStartInKST(t *testing.T) {
	err := &AmbiguousRunsError{Owners: []homestate.RunOwner{{
		RunID: "run-kst", LeadPID: 4242, Classification: homestate.OwnerLive,
		CreatedAt: "2026-10-17T09:31:00.123456789Z",
	}}}
	msg := err.Error()
	if !strings.Contains(msg, "started 2026-10-17 18:31:00 KST") {
		t.Errorf("refusal %q does not render the start as KST", msg)
	}
	if strings.Contains(msg, "09:31:00") {
		t.Errorf("refusal %q still carries the raw UTC clock reading", msg)
	}
}
