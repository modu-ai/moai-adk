package escalation_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

func TestParseRecordRejectsMissingFrontmatter(t *testing.T) {
	for name, body := range map[string]string{
		"no-leading-fence":  "# just markdown\n",
		"unterminated":      "---\nschema_version: 1\n",
		"invalid-yaml-type": "---\noccurrences: [1, 2]\n---\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := escalation.ParseRecord([]byte(body)); err == nil {
				t.Error("ParseRecord accepted a record without valid frontmatter")
			}
		})
	}
}

// CRLF line endings do not break frontmatter parsing.
func TestParseRecordAcceptsCRLF(t *testing.T) {
	data, err := sampleRecord("t9001").Marshal()
	if err != nil {
		t.Fatal(err)
	}
	r, err := escalation.ParseRecord([]byte(strings.ReplaceAll(string(data), "\n", "\r\n")))
	if err != nil || r.Card != "t9001" || r.Occurrences != 1 {
		t.Errorf("ParseRecord(CRLF) = %+v, %v", r, err)
	}
}

// A record that cannot be parsed is an error, never "no decision needed".
func TestNeedsDecisionMalformedRecordIsError(t *testing.T) {
	w := escalationtest.NewWorktree(t, "t9001")
	w.Write(".moai/reports/t9001/escalation/ownership-move-0000000000000000.md", "not a record\n")
	if _, err := escalation.NeedsDecision(w.Root, "t9001"); err == nil {
		t.Error("NeedsDecision accepted a malformed record")
	}
}

// The fingerprint depends on the class and every part, and is stable.
func TestFingerprintStableAndDistinct(t *testing.T) {
	a := escalation.Fingerprint(escalation.ClassOwnershipMove, "x", "y")
	if a != escalation.Fingerprint(escalation.ClassOwnershipMove, "x", "y") {
		t.Error("fingerprint not stable")
	}
	for _, other := range []string{
		escalation.Fingerprint(escalation.ClassInvariantViolation, "x", "y"),
		escalation.Fingerprint(escalation.ClassOwnershipMove, "xy"),
		escalation.Fingerprint(escalation.ClassOwnershipMove, "x", "z"),
	} {
		if other == a {
			t.Errorf("fingerprint collision: %s", a)
		}
	}
}

// A record with no not-observed entries still renders the section and parses
// back to an empty list.
func TestMarshalEmptyNotObserved(t *testing.T) {
	r := sampleRecord("t9001")
	r.NotObserved = nil
	data, err := r.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "## Not observed\n\n- none\n") {
		t.Errorf("empty not-observed section not rendered:\n%s", data)
	}
	back, err := escalation.ParseRecord(data)
	if err != nil || back.NotObserved == nil || len(back.NotObserved) != 0 {
		t.Errorf("not_observed round-trip = %#v, %v", back.NotObserved, err)
	}
}
