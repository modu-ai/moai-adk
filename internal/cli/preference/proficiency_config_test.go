package preference

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadProficiencyThresholds_FromConfig verifies the optional
// preference.yaml override path (design.md §A.4). A config file with
// expert_sessions + floor_sessions overrides the package defaults.
func TestLoadProficiencyThresholds_FromConfig(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	// Write a config with overridden thresholds.
	content := `# preference.yaml — optional proficiency threshold override (design.md §A.4)
preference:
  proficiency:
    expert_sessions: 30   # stricter expert band
    floor_sessions: 10    # general starts at 10 sessions
`
	path := filepath.Join(configDir, preferenceConfigFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got := LoadProficiencyThresholds(tmp)
	if got.Expert != 30 {
		t.Errorf("Expert = %d, want 30 (from config)", got.Expert)
	}
	if got.Floor != 10 {
		t.Errorf("Floor = %d, want 10 (from config)", got.Floor)
	}

	// EstimateProficiencyWithThresholds honors the override.
	if got := EstimateProficiencyWithThresholds(35, got.Expert, got.Floor); got != ProficiencyExpert {
		t.Errorf("with override, count=35 → %v, want Expert", got)
	}
	if got := EstimateProficiencyWithThresholds(15, got.Expert, got.Floor); got != ProficiencyGeneral {
		t.Errorf("with override, count=15 → %v, want General (10..29 band)", got)
	}
	if got := EstimateProficiencyWithThresholds(5, got.Expert, got.Floor); got != ProficiencyColdStart {
		t.Errorf("with override, count=5 → %v, want ColdStart (< 10 floor)", got)
	}
}

// TestLoadProficiencyThresholds_MissingFileFallsBack verifies the config
// fallback contract: a missing preference.yaml yields zero-value thresholds
// (caller passes the package defaults). A missing file MUST NOT break
// proficiency inference.
func TestLoadProficiencyThresholds_MissingFileFallsBack(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir() // no config written
	got := LoadProficiencyThresholds(tmp)
	if got.Expert != 0 || got.Floor != 0 {
		t.Errorf("missing config: Expert=%d Floor=%d, want 0/0 (zero → caller uses defaults)", got.Expert, got.Floor)
	}
}

// TestLoadProficiencyThresholds_UnparseableFallsBack verifies a corrupt or
// non-integer config value is ignored (zero-value retained) rather than
// erroring — the same config-fallback discipline.
func TestLoadProficiencyThresholds_UnparseableFallsBack(t *testing.T) {
	t.Parallel()
	raw := `preference:
  proficiency:
    expert_sessions: not-a-number
    floor_sessions: also-bad
`
	got := parseProficiencyThresholds(raw)
	if got.Expert != 0 || got.Floor != 0 {
		t.Errorf("unparseable config: Expert=%d Floor=%d, want 0/0 (bad values ignored)", got.Expert, got.Floor)
	}
}

// TestEstimateProficiencyWithThresholds_DegenerateBand verifies the defensive
// collapse when expert <= floor (misconfiguration): the General band empties
// and the function falls back to the ColdStart/Expert boundary.
func TestEstimateProficiencyWithThresholds_DegenerateBand(t *testing.T) {
	t.Parallel()
	// expert == floor (General band is empty).
	if got := EstimateProficiencyWithThresholds(15, 10, 10); got != ProficiencyExpert {
		t.Errorf("degenerate (expert=floor=10), count=15 → %v, want Expert (collapsed band)", got)
	}
	if got := EstimateProficiencyWithThresholds(5, 10, 10); got != ProficiencyColdStart {
		t.Errorf("degenerate (expert=floor=10), count=5 → %v, want ColdStart (collapsed band)", got)
	}
	// expert < floor (inverted — also degenerate).
	if got := EstimateProficiencyWithThresholds(8, 5, 10); got != ProficiencyExpert {
		t.Errorf("inverted (expert=5 < floor=10), count=8 → %v, want Expert (collapsed band)", got)
	}
}

// TestEstimateProficiencyWithThresholds_ZeroOrNegativeNormalized verifies
// non-positive thresholds are normalized to the package defaults so a bad
// config never produces undefined behavior.
func TestEstimateProficiencyWithThresholds_ZeroOrNegativeNormalized(t *testing.T) {
	t.Parallel()
	// Zero expert → normalize to default (20).
	if got := EstimateProficiencyWithThresholds(25, 0, 5); got != ProficiencyExpert {
		t.Errorf("expert=0, count=25 → %v, want Expert (normalized to default 20)", got)
	}
	// Negative floor → normalize to default (5).
	if got := EstimateProficiencyWithThresholds(10, 20, -3); got != ProficiencyGeneral {
		t.Errorf("floor=-3, count=10 → %v, want General (normalized to default 5..19)", got)
	}
}

// TestParseThresholdKV verifies the YAML-leaf extractor handles quoted values,
// inline comments, and rejects non-integers.
func TestParseThresholdKV(t *testing.T) {
	t.Parallel()
	cases := []struct {
		line  string
		key   string
		want  int
		found bool
	}{
		{"expert_sessions: 20", "expert_sessions:", 20, true},
		{"expert_sessions: 20   # note", "expert_sessions:", 20, true},
		{`expert_sessions: "20"`, "expert_sessions:", 20, true},
		{`floor_sessions: '5'`, "floor_sessions:", 5, true},
		{"expert_sessions: not-a-number", "expert_sessions:", 0, false},
		{"unrelated_key: 7", "expert_sessions:", 0, false},
		{"expert_sessions: -3", "expert_sessions:", 0, false}, // negative rejected
	}
	for _, tc := range cases {
		got, found := parseThresholdKV(tc.line, tc.key)
		if got != tc.want || found != tc.found {
			t.Errorf("parseThresholdKV(%q, %q) = (%d, %v), want (%d, %v)",
				tc.line, tc.key, got, found, tc.want, tc.found)
		}
	}
}
