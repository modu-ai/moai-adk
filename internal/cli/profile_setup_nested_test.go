package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// seedNestedProjectConfig writes a temp project with quality.yaml +
// git-convention.yaml carrying the 7 nested fields so the TUI nested persistence
// path can be round-tripped (SPEC-WEB-CONSOLE-010 M3).
func seedNestedProjectConfig(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"quality.yaml": "constitution:\n" +
			"  development_mode: tdd\n" +
			"  enforce_quality: true\n" +
			"  test_coverage_target: 80\n" +
			"  tdd_settings:\n" +
			"    min_coverage_per_commit: 70\n",
		"git-convention.yaml": "git_convention:\n" +
			"  convention: auto\n" +
			"  auto_detection:\n" +
			"    enabled: true\n" +
			"    confidence_threshold: 0.6\n" +
			"    sample_size: 100\n" +
			"  validation:\n" +
			"    enforce_on_push: false\n",
		"workflow.yaml": "workflow:\n  sentinel: DO_NOT_TOUCH\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

// TestTUINestedConfigRoundTrip covers AC-WC10-013: the TUI persistence path writes
// all 7 nested fields through the shared nested seam and reads them back.
func TestTUINestedConfigRoundTrip(t *testing.T) {
	root := seedNestedProjectConfig(t)

	in := nestedTUIInputs{
		CoverageTarget:    "92",
		MinCoverage:       "85",
		Confidence:        "0.8",
		SampleSize:        "250",
		EnforceQuality:    false,
		EnforceQualitySet: true,
		AutoDetectionOn:   false,
		AutoDetectionSet:  true,
		EnforceOnPush:     true,
		EnforceOnPushSet:  true,
	}
	if err := persistProjectNestedConfig(root, in); err != nil {
		t.Fatalf("persistProjectNestedConfig: %v", err)
	}

	got, err := readCurrentNestedConfig(root)
	if err != nil {
		t.Fatalf("readCurrentNestedConfig: %v", err)
	}
	checks := map[string]struct {
		got, want any
	}{
		"CoverageTarget":       {got.CoverageTarget, "92"},
		"MinCoverage":          {got.MinCoverage, "85"},
		"ConfidenceThreshold":  {got.ConfidenceThreshold, "0.8"},
		"SampleSize":           {got.SampleSize, "250"},
		"EnforceQuality":       {got.EnforceQuality, false},
		"AutoDetectionEnabled": {got.AutoDetectionEnabled, false},
		"EnforceOnPush":        {got.EnforceOnPush, true},
	}
	for field, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", field, c.got, c.want)
		}
	}
}

// TestTUINestedConfigEmptyPreserve covers AC-WC10-017: an empty numeric input
// (= preserve) leaves the on-disk value unchanged.
func TestTUINestedConfigEmptyPreserve(t *testing.T) {
	root := seedNestedProjectConfig(t)

	// Only coverage_target submitted; the rest empty (numeric) → preserve.
	// Bool flags are NOT set (simulating outside-project no-op) so they preserve too.
	in := nestedTUIInputs{
		CoverageTarget: "95",
		// MinCoverage / Confidence / SampleSize empty → preserve
		// bool *Set flags false → preserve
	}
	if err := persistProjectNestedConfig(root, in); err != nil {
		t.Fatalf("persistProjectNestedConfig: %v", err)
	}

	got, err := readCurrentNestedConfig(root)
	if err != nil {
		t.Fatalf("readCurrentNestedConfig: %v", err)
	}
	if got.CoverageTarget != "95" {
		t.Errorf("CoverageTarget = %q, want 95 (submitted)", got.CoverageTarget)
	}
	if got.MinCoverage != "70" {
		t.Errorf("MinCoverage = %q, want 70 (preserved)", got.MinCoverage)
	}
	if got.ConfidenceThreshold != "0.6" {
		t.Errorf("ConfidenceThreshold = %q, want 0.6 (preserved)", got.ConfidenceThreshold)
	}
	if got.SampleSize != "100" {
		t.Errorf("SampleSize = %q, want 100 (preserved)", got.SampleSize)
	}
	if got.EnforceQuality != true {
		t.Errorf("EnforceQuality = %v, want true (preserved, *Set false)", got.EnforceQuality)
	}
	if got.AutoDetectionEnabled != true {
		t.Errorf("AutoDetectionEnabled = %v, want true (preserved, *Set false)", got.AutoDetectionEnabled)
	}
}

// TestTUINestedConfigNoParallelWriter covers AC-WC10-013's grep side: profile_setup.go
// must NOT carry a parallel yaml.Marshal / os.WriteFile for nested config — it must
// route through the shared settings seam (AP-2). Comment lines are excluded so the
// doctrine prose ("no direct yaml.Marshal/os.WriteFile") does not false-positive.
func TestTUINestedConfigNoParallelWriter(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	codeLines := nonCommentLines(string(data))
	if strings.Contains(codeLines, "yaml.Marshal") {
		t.Error("profile_setup.go must NOT call yaml.Marshal directly (AP-2 — use the shared config-manager seam)")
	}
	if strings.Contains(codeLines, "os.WriteFile") {
		t.Error("profile_setup.go must NOT call os.WriteFile directly for config persistence (AP-2)")
	}
	// It MUST call the shared nested seam (the wrapper persistProjectNestedConfig
	// delegates to settings.WriteProjectNestedConfig).
	if !strings.Contains(codeLines, "persistProjectNestedConfig") {
		t.Error("profile_setup.go must drive the shared nested write seam (persistProjectNestedConfig)")
	}
}

// nonCommentLines returns src with whole-line `//` comments stripped, so a grep
// guard tests actual code rather than doctrine prose in comments.
func nonCommentLines(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// TestPermissionModeNormalizeAcceptEdits covers AC-WC10-015: acceptEdits normalizes
// to empty string so no redundant override is persisted. The TUI applies this in
// runProfileSetup; the schema's permission_mode Persist.Normalize encodes the same
// semantic for both surfaces.
func TestPermissionModeNormalizeAcceptEdits(t *testing.T) {
	// Schema-level normalization (shared by both surfaces).
	f, ok := settings.Field("permission_mode")
	if !ok {
		t.Fatal("permission_mode field not found in schema")
	}
	if f.Persist.Normalize == nil {
		t.Fatal("permission_mode Persist.Normalize is nil")
	}
	if got := f.Persist.Normalize("acceptEdits"); got != "" {
		t.Errorf("schema normalize(acceptEdits) = %q, want empty", got)
	}
	if got := f.Persist.Normalize("plan"); got != "plan" {
		t.Errorf("schema normalize(plan) = %q, want plan", got)
	}

	// TUI source-level guard: profile_setup.go must still apply the acceptEdits→""
	// normalization in its save path (the existing line 443 semantic).
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "permissionMode == defaultPermissionMode") {
		t.Error("profile_setup.go must normalize acceptEdits permission mode to empty (REQ-WC10-014)")
	}
}

// TestTUIEmptyLabelsSchemaSourced covers AC-WC10-014 (TUI side): the wizard sources
// its empty-option labels from the schema (settings.EmptyLabelFor), not inline
// literals — so both surfaces render the IDENTICAL canonical label per field.
func TestTUIEmptyLabelsSchemaSourced(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	src := string(data)
	for _, field := range []string{"model", "effort_level", "development_mode", "git_convention", "model_policy"} {
		marker := `settings.EmptyLabelFor("` + field + `")`
		if !strings.Contains(src, marker) {
			t.Errorf("profile_setup.go must source empty label from schema for %q (expected %s)", field, marker)
		}
	}
}

// TestConfigManagerStillUsed pins that the TUI nested write goes through the config
// manager (the shared seam delegates to it) — defense for the whole-section-copy
// invariant (B4).
func TestConfigManagerStillUsed(t *testing.T) {
	root := seedNestedProjectConfig(t)
	in := nestedTUIInputs{CoverageTarget: "99", AutoDetectionOn: true, AutoDetectionSet: true}
	if err := persistProjectNestedConfig(root, in); err != nil {
		t.Fatalf("persistProjectNestedConfig: %v", err)
	}
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	// Whole-section-copy: development_mode in the same quality section must survive.
	if string(cfg.Quality.DevelopmentMode) != "tdd" {
		t.Errorf("Quality.DevelopmentMode = %q, want tdd (whole-section-copy preserved)", cfg.Quality.DevelopmentMode)
	}
}
