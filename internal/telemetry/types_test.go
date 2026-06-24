package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestUsageRecordBackwardCompat_LegacyLineDecodes verifies that a legacy JSONL
// line written before SPEC-STOP-EVIDENCE-GATE-001 (without the new
// is_test_pass / is_test_fail / path_kind fields) still decodes successfully
// into UsageRecord, with the new fields landing at zero values (AC-SEG-010).
//
// REQ-SEG-010: absent new fields are "evidence not observable", NOT an error.
func TestUsageRecordBackwardCompat_LegacyLineDecodes(t *testing.T) {
	t.Parallel()

	// A legacy line: the exact 9-field shape produced before this SPEC.
	legacy := `{"ts":"2026-06-15T10:30:00Z","session_id":"sess-legacy","skill_id":"moai-workflow-tdd","trigger":"explicit","context_hash":"a1b2c3d4","agent_type":"manager-develop","phase":"run","duration_ms":45000,"outcome":"success"}`

	var got UsageRecord
	if err := json.Unmarshal([]byte(legacy), &got); err != nil {
		t.Fatalf("legacy line must decode without error, got: %v", err)
	}

	// Pre-existing fields must survive.
	if got.SessionID != "sess-legacy" {
		t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-legacy")
	}
	if got.Outcome != OutcomeSuccess {
		t.Errorf("Outcome = %q, want %q", got.Outcome, OutcomeSuccess)
	}

	// New fields must be zero values (absent in legacy JSONL).
	if got.IsTestPass {
		t.Errorf("IsTestPass = true, want false (absent in legacy line)")
	}
	if got.IsTestFail {
		t.Errorf("IsTestFail = true, want false (absent in legacy line)")
	}
	if got.PathKind != "" {
		t.Errorf("PathKind = %q, want empty (absent in legacy line)", got.PathKind)
	}
}

// TestUsageRecordOmitempty_ZeroValueOmitsNewKeys verifies that marshaling a
// UsageRecord with zero-value new fields OMITS the three new JSON keys entirely
// (omitempty), so existing JSONL parsers observe no schema change (AC-SEG-010).
func TestUsageRecordOmitempty_ZeroValueOmitsNewKeys(t *testing.T) {
	t.Parallel()

	r := UsageRecord{
		Timestamp: time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC),
		SessionID: "sess-zero",
		SkillID:   "moai-workflow-tdd",
		Trigger:   TriggerExplicit,
		Outcome:   OutcomeUnknown,
		// new fields left at zero value
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	out := string(data)

	for _, key := range []string{"is_test_pass", "is_test_fail", "path_kind"} {
		if strings.Contains(out, key) {
			t.Errorf("zero-value marshal must OMIT %q, got: %s", key, out)
		}
	}
}

// TestUsageRecordOmitempty_NonZeroValuePersists verifies that when the new
// fields carry non-zero values they ARE serialized (round-trip), so a
// record-time writer successor SPEC can populate them and the gate can read
// them back.
func TestUsageRecordOmitempty_NonZeroValuePersists(t *testing.T) {
	t.Parallel()

	r := UsageRecord{
		Timestamp:  time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC),
		SessionID:  "sess-full",
		SkillID:    "moai-workflow-tdd",
		Trigger:    TriggerExplicit,
		Outcome:    OutcomeSuccess,
		IsTestPass: true,
		IsTestFail: false,
		PathKind:   PathKindCodeChange,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var back UsageRecord
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("round-trip unmarshal error: %v", err)
	}

	if !back.IsTestPass {
		t.Errorf("IsTestPass did not round-trip; got false")
	}
	if back.PathKind != PathKindCodeChange {
		t.Errorf("PathKind = %q, want %q", back.PathKind, PathKindCodeChange)
	}
	// is_test_pass=true must appear in the serialized form.
	if !strings.Contains(string(data), "is_test_pass") {
		t.Errorf("is_test_pass must be present when true, got: %s", string(data))
	}
	// is_test_fail=false must be omitted (omitempty).
	if strings.Contains(string(data), "is_test_fail") {
		t.Errorf("is_test_fail must be omitted when false, got: %s", string(data))
	}
}

// TestPathKindConstants pins the three path-kind bucket constant values so they
// match the design.md §2.2 taxonomy + the JSON values readers infer against.
func TestPathKindConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		got  string
		want string
	}{
		{PathKindDocsOnly, "docs-only"},
		{PathKindCodeChange, "code-change"},
		{PathKindUnknown, "unknown"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("path-kind constant = %q, want %q", c.got, c.want)
		}
	}
}
