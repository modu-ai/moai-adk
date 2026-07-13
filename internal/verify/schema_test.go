package verify

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestSnapshotSchema asserts the snapshot JSON carries: check id, exact
// command, exit code, parsed result counts, capture timestamp, execution
// duration, and the snapshot key — and that the per-check conditions block
// decodes the loop-verdict conditions shape (read-compat).
func TestSnapshotSchema(t *testing.T) {
	t.Parallel()

	errCount := 0
	testsPass := true
	covThreshold := 85.0
	covActual := 87.0
	zeroErrors := true
	zeroWarnings := false

	snap := &Snapshot{
		Key:        "abc123:deadbeefdeadbeef",
		RecordedAt: time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC),
		Checks: []CheckEntry{
			{
				CheckID:    "test",
				Command:    "go test ./...",
				ExitCode:   0,
				RecordedAt: time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC),
				DurationMS: 41200,
				Conditions: &Conditions{
					ZeroErrors:        &zeroErrors,
					ErrorCount:        &errCount,
					TestsPass:         &testsPass,
					CoverageThreshold: &covThreshold,
					CoverageActual:    &covActual,
					ZeroWarnings:      &zeroWarnings,
				},
			},
		},
	}

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	js := string(data)
	for _, field := range []string{
		`"key"`, `"recorded_at"`, `"checks"`,
		`"check_id"`, `"command"`, `"exit_code"`, `"duration_ms"`,
		`"conditions"`,
		// loop-verdict-compatible conditions field names (REQ-SNAP-001)
		`"zero_errors"`, `"error_count"`, `"tests_pass"`,
		`"coverage_threshold"`, `"coverage_actual"`, `"zero_warnings"`,
	} {
		if !strings.Contains(js, field) {
			t.Errorf("snapshot JSON missing field %s: %s", field, js)
		}
	}
}

// TestSnapshotSchemaLoopVerdictDecode asserts a loop-verdict-shaped conditions
// JSON block decodes into the Conditions type without loss (read-compat leg).
func TestSnapshotSchemaLoopVerdictDecode(t *testing.T) {
	t.Parallel()

	loopVerdictConditions := `{
		"zero_errors": true,
		"error_count": 0,
		"tests_pass": true,
		"coverage_threshold": 85,
		"coverage_actual": 87.0,
		"zero_warnings": false
	}`
	var c Conditions
	if err := json.Unmarshal([]byte(loopVerdictConditions), &c); err != nil {
		t.Fatalf("loop-verdict conditions must decode: %v", err)
	}
	if c.ZeroErrors == nil || !*c.ZeroErrors {
		t.Error("zero_errors: want true")
	}
	if c.ErrorCount == nil || *c.ErrorCount != 0 {
		t.Error("error_count: want 0")
	}
	if c.TestsPass == nil || !*c.TestsPass {
		t.Error("tests_pass: want true")
	}
	if c.CoverageThreshold == nil || *c.CoverageThreshold != 85 {
		t.Error("coverage_threshold: want 85")
	}
	if c.CoverageActual == nil || *c.CoverageActual != 87.0 {
		t.Error("coverage_actual: want 87.0")
	}
	if c.ZeroWarnings == nil || *c.ZeroWarnings {
		t.Error("zero_warnings: want false")
	}
}

// TestSnapshotFindCommand asserts exact byte-string command matching: a
// near-miss variant (added flag) never matches (REQ-SNAP-010 M1 granularity).
func TestSnapshotFindCommand(t *testing.T) {
	t.Parallel()

	snap := &Snapshot{
		Key: "k",
		Checks: []CheckEntry{
			{CheckID: "test", Command: "go test ./...", ExitCode: 0},
		},
	}
	if e := snap.FindCommand("go test ./..."); e == nil {
		t.Fatal("exact byte-string match must hit")
	}
	if e := snap.FindCommand("go test -count=1 ./..."); e != nil {
		t.Error("near-miss command variant must NOT match (no normalization in M1)")
	}
	if e := snap.FindCommand("go test ./... "); e != nil {
		t.Error("trailing-space variant must NOT match (exact byte-string only)")
	}
}
