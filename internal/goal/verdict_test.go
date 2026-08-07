package goal

import (
	"os"
	"path/filepath"
	"testing"
)

// TestVerdictPathSidecar verifies the sidecar lives beside the .json/.html pair
// (SPEC-GOAL-HTML-WIRING-001 REQ-WIRE-002 / AC-WIRE-001 precondition).
func TestVerdictPathSidecar(t *testing.T) {
	got := VerdictPath("/proj", "sess-1")
	want := filepath.Join("/proj", StateDir, "sess-1.verdict.json")
	if got != want {
		t.Errorf("VerdictPath = %q, want %q", got, want)
	}
	// Sidecar shares the directory with StatePath / HTMLPath.
	jsonPath := StatePath("/proj", "sess-1")
	if filepath.Dir(jsonPath) != filepath.Dir(got) {
		t.Errorf("dirs differ: json=%s verdict=%s", filepath.Dir(jsonPath), filepath.Dir(got))
	}
}

// TestSaveLoadVerdictRoundTrip verifies SaveVerdict → LoadVerdict round-trips the
// full Verdict (including the 5-section CeilingVerdict). This underpins AC-WIRE-001
// (the render path loads what the evaluator persisted).
func TestSaveLoadVerdictRoundTrip(t *testing.T) {
	root := t.TempDir()
	sid := "sess-rt"
	in := &Verdict{
		Decision: "block",
		Reason:   "ceiling reached",
		Turn:     7,
		Ceiling:  7,
		FailedConditions: []FailedCond{
			{Cmd: "go test ./...", Exit: 1, Tail: "FAIL: TestFoo"},
		},
		CeilingExit: true,
		Verdict: &CeilingVerdict{
			Claim:               "the goal did not converge",
			Evidence:            "turns_used=7",
			BaselineAttribution: "baseline A",
			Gaps:                "gap B",
			ResidualRisk:        "risk C",
		},
	}
	if err := SaveVerdict(root, sid, in); err != nil {
		t.Fatalf("SaveVerdict: %v", err)
	}

	// The sidecar file exists at VerdictPath.
	if _, err := os.Stat(VerdictPath(root, sid)); err != nil {
		t.Fatalf("sidecar missing after SaveVerdict: %v", err)
	}

	out, err := LoadVerdict(root, sid)
	if err != nil {
		t.Fatalf("LoadVerdict: %v", err)
	}
	if out == nil {
		t.Fatal("LoadVerdict returned nil after SaveVerdict")
	}
	if out.Turn != in.Turn || out.Ceiling != in.Ceiling {
		t.Errorf("Turn/Ceiling round-trip mismatch: got %+v want %+v", out, in)
	}
	if out.Verdict == nil {
		t.Fatal("round-trip dropped the *CeilingVerdict")
	}
	if out.Verdict.Claim != in.Verdict.Claim {
		t.Errorf("Claim mismatch: got %q want %q", out.Verdict.Claim, in.Verdict.Claim)
	}
	if out.Verdict.BaselineAttribution != in.Verdict.BaselineAttribution {
		t.Errorf("BaselineAttribution mismatch: got %q want %q",
			out.Verdict.BaselineAttribution, in.Verdict.BaselineAttribution)
	}
}

// TestLoadVerdictAbsentReturnsNil verifies the fail-open read path
// (REQ-WIRE-003 / AC-WIRE-002): no sidecar → (nil, nil), NOT an error.
func TestLoadVerdictAbsentReturnsNil(t *testing.T) {
	root := t.TempDir()
	v, err := LoadVerdict(root, "never-armed")
	if err != nil {
		t.Errorf("LoadVerdict on absent sidecar returned err: %v", err)
	}
	if v != nil {
		t.Errorf("LoadVerdict on absent sidecar returned non-nil: %+v", v)
	}
}

// TestLoadVerdictCorruptReturnsNil verifies the fail-open read path on a
// corrupt sidecar (REQ-WIRE-003): the render command degrades to nil rather
// than erroring out.
func TestLoadVerdictCorruptReturnsNil(t *testing.T) {
	root := t.TempDir()
	sid := "sess-bad"
	dir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(VerdictPath(root, sid), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, _ := LoadVerdict(root, sid)
	if v != nil {
		t.Errorf("LoadVerdict on corrupt sidecar returned non-nil: %+v", v)
	}
}
