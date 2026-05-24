// Package proposalgen reader unit tests.
// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-001..003.
package proposalgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReader_LiveFixture verifies AC-PGN-001: the reader parses the frozen
// baseline snapshot of tier-promotions.jsonl (8 records, 4 unique pattern_keys
// as of 2026-05-24) without error.
func TestReader_LiveFixture(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "tier-promotions-current-baseline.jsonl")
	promotions, malformed, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions(%q) error = %v, want nil", path, err)
	}
	if malformed != 0 {
		t.Errorf("malformed = %d, want 0 on clean baseline fixture", malformed)
	}
	if got, want := len(promotions), 8; got != want {
		t.Errorf("len(promotions) = %d, want %d", got, want)
	}
	unique := uniquePatternKeys(promotions)
	if got, want := len(unique), 4; got != want {
		t.Errorf("len(uniquePatternKeys) = %d, want %d (keys=%v)", got, want, unique)
	}

	expected := map[string]bool{
		"agent_invocation:Bash:": true,
		"subagent_stop:unknown:": true,
		"session_stop::":         true,
		"user_prompt::":          true,
	}
	for _, k := range unique {
		if !expected[k] {
			t.Errorf("unexpected pattern_key in baseline fixture: %q", k)
		}
	}
}

// TestReader_MissingFile verifies REQ-PGN-003: a missing input file returns
// an empty slice without error (graceful no-op).
func TestReader_MissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "does-not-exist.jsonl")
	promotions, malformed, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions(missing) error = %v, want nil per REQ-PGN-003", err)
	}
	if len(promotions) != 0 {
		t.Errorf("len(promotions) = %d, want 0 on missing file", len(promotions))
	}
	if malformed != 0 {
		t.Errorf("malformed = %d, want 0 on missing file", malformed)
	}
}

// TestReader_EmptyFile verifies REQ-PGN-003: a zero-byte file returns an
// empty slice without error.
func TestReader_EmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "empty.jsonl")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	promotions, malformed, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions(empty) error = %v, want nil", err)
	}
	if len(promotions) != 0 {
		t.Errorf("len(promotions) = %d, want 0 on empty file", len(promotions))
	}
	if malformed != 0 {
		t.Errorf("malformed = %d, want 0 on empty file", malformed)
	}
}

// TestReader_MalformedLineTolerance verifies REQ-PGN-002: malformed lines are
// skipped, the malformed counter accumulates, and valid lines continue to be
// parsed.
func TestReader_MalformedLineTolerance(t *testing.T) {
	t.Parallel()

	const data = `{"ts":"2026-05-24T10:08:37.191814Z","pattern_key":"agent_invocation:Bash:","from_tier":"","to_tier":"observation","observation_count":1,"confidence":1}
this is not json at all
{"ts":"2026-05-24T10:08:37.191953Z","pattern_key":"subagent_stop:unknown:","from_tier":"","to_tier":"auto_update","observation_count":41,"confidence":1}
`
	path := filepath.Join(t.TempDir(), "mixed.jsonl")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	promotions, malformed, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions(mixed) error = %v, want nil", err)
	}
	if got, want := len(promotions), 2; got != want {
		t.Errorf("len(promotions) = %d, want %d (malformed line should be skipped)", got, want)
	}
	if malformed != 1 {
		t.Errorf("malformed = %d, want 1", malformed)
	}
}

// TestReader_BlankLinesIgnored verifies that blank lines between records do
// not contribute to malformed_lines and do not produce zero-value Promotion
// entries.
func TestReader_BlankLinesIgnored(t *testing.T) {
	t.Parallel()

	const data = `{"ts":"2026-05-24T10:08:37.191814Z","pattern_key":"agent_invocation:Bash:","from_tier":"","to_tier":"observation","observation_count":1,"confidence":1}

{"ts":"2026-05-24T10:08:37.191953Z","pattern_key":"subagent_stop:unknown:","from_tier":"","to_tier":"auto_update","observation_count":41,"confidence":1}

`
	path := filepath.Join(t.TempDir(), "blanks.jsonl")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	promotions, malformed, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions(blanks) error = %v, want nil", err)
	}
	if got, want := len(promotions), 2; got != want {
		t.Errorf("len(promotions) = %d, want %d", got, want)
	}
	if malformed != 0 {
		t.Errorf("malformed = %d, want 0 (blank lines should not count as malformed)", malformed)
	}
}

// TestReader_AllMalformed asserts that a file containing only malformed lines
// returns an empty slice plus the full malformed count, without producing an
// error.
func TestReader_AllMalformed(t *testing.T) {
	t.Parallel()

	const data = "not json line 1\nnot json line 2\nnot json line 3\n"
	path := filepath.Join(t.TempDir(), "garbage.jsonl")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	promotions, malformed, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions(garbage) error = %v, want nil", err)
	}
	if len(promotions) != 0 {
		t.Errorf("len(promotions) = %d, want 0", len(promotions))
	}
	if malformed != 3 {
		t.Errorf("malformed = %d, want 3", malformed)
	}
}

// TestReader_FieldsParsedCorrectly verifies that all six Promotion fields are
// unmarshaled into the canonical internal/harness.Promotion struct.
func TestReader_FieldsParsedCorrectly(t *testing.T) {
	t.Parallel()

	const data = `{"ts":"2026-05-24T10:08:37.191814Z","pattern_key":"code_change:func_extract:auth_module","from_tier":"observation","to_tier":"recommendation","observation_count":7,"confidence":0.85}
`
	path := filepath.Join(t.TempDir(), "one.jsonl")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	promotions, _, err := ReadPromotions(path)
	if err != nil || len(promotions) != 1 {
		t.Fatalf("ReadPromotions: err=%v len=%d, want err=nil len=1", err, len(promotions))
	}
	p := promotions[0]
	if p.PatternKey != "code_change:func_extract:auth_module" {
		t.Errorf("PatternKey = %q, want code_change:func_extract:auth_module", p.PatternKey)
	}
	if p.FromTier != "observation" {
		t.Errorf("FromTier = %q, want observation", p.FromTier)
	}
	if p.ToTier != "recommendation" {
		t.Errorf("ToTier = %q, want recommendation", p.ToTier)
	}
	if p.ObservationCount != 7 {
		t.Errorf("ObservationCount = %d, want 7", p.ObservationCount)
	}
	if p.Confidence != 0.85 {
		t.Errorf("Confidence = %v, want 0.85", p.Confidence)
	}
	if p.Ts.IsZero() {
		t.Errorf("Ts is zero; expected parsed timestamp")
	}
}

// TestUniquePatternKeys_Ordering asserts deterministic alphabetical ordering
// of the unique pattern key slice (required for stable test assertions and
// JSON output stability downstream).
func TestUniquePatternKeys_Ordering(t *testing.T) {
	t.Parallel()

	const data = `{"ts":"2026-05-24T10:00:00Z","pattern_key":"z_pattern:x:y","from_tier":"","to_tier":"observation","observation_count":1,"confidence":1}
{"ts":"2026-05-24T10:01:00Z","pattern_key":"a_pattern:x:y","from_tier":"","to_tier":"observation","observation_count":1,"confidence":1}
{"ts":"2026-05-24T10:02:00Z","pattern_key":"m_pattern:x:y","from_tier":"","to_tier":"observation","observation_count":1,"confidence":1}
{"ts":"2026-05-24T10:03:00Z","pattern_key":"a_pattern:x:y","from_tier":"","to_tier":"observation","observation_count":2,"confidence":1}
`
	path := filepath.Join(t.TempDir(), "ordering.jsonl")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	promotions, _, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions: %v", err)
	}
	got := uniquePatternKeys(promotions)
	want := []string{"a_pattern:x:y", "m_pattern:x:y", "z_pattern:x:y"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("uniquePatternKeys ordering = %v, want %v", got, want)
	}
}
