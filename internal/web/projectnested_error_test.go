package web

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// TestProjectNestedGitConventionRoundTrip exercises the git_convention nested write
// branch end-to-end (confidence + enabled toggle + sample_size + enforce_on_push) so
// writeProjectNestedConfig's git_convention path is covered, alongside the quality
// path. Complements TestProjectNestedSiblingPreserved (quality branch). The custom
// engine is removed (REQ-WC9-010), so the live levers are exercised instead.
func TestProjectNestedGitConventionRoundTrip(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	form := nestedSaveForm(map[string]string{
		"git_convention.auto_detection.confidence_threshold": "0.88",
		"git_convention.auto_detection.enabled__present":     "1", // companion + checkbox absent → false
		"git_convention.auto_detection.sample_size":          "200",
		"git_convention.validation.enforce_on_push__present":  "1",
		"git_convention.validation.enforce_on_push":           "1",
	})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("git_convention nested save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	if cfg.GitConvention.AutoDetection.ConfidenceThreshold != 0.88 {
		t.Errorf("confidence_threshold = %v, want 0.88", cfg.GitConvention.AutoDetection.ConfidenceThreshold)
	}
	if cfg.GitConvention.AutoDetection.Enabled {
		t.Error("auto_detection.enabled = true, want false (companion + unchecked)")
	}
	if cfg.GitConvention.AutoDetection.SampleSize != 200 {
		t.Errorf("sample_size = %d, want 200", cfg.GitConvention.AutoDetection.SampleSize)
	}
	if !cfg.GitConvention.Validation.EnforceOnPush {
		t.Error("validation.enforce_on_push = false, want true (companion + checkbox)")
	}
}

// TestProjectNestedRejectStillRejects covers the rejected-POST server contract after
// the project render surface was retired (SPEC-DESIGN-MOAIWEBV2-001 M1): a POST with
// an invalid nested field is still validated and rejected atomically (400, no disk
// write). The value-echo re-render assertions were removed with the project widgets —
// those fields no longer render (development_mode / git_convention / quality.* are now
// editable via yaml config / CLI), but the preserved parse+validate seam
// (parseProjectNestedForm, REQ-MWV2-031) still rejects invalid submissions.
func TestProjectNestedRejectStillRejects(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	// Submit the nested fields, with min_coverage_per_commit invalid (200) to force reject.
	form := nestedSaveForm(map[string]string{
		"quality.test_coverage_target":                       "95",
		"quality.enforce_quality__present":                   "1",
		"quality.enforce_quality":                            "1",
		"quality.tdd_settings.min_coverage_per_commit":       "200", // invalid → reject
		"git_convention.auto_detection.confidence_threshold": "0.33",
		"git_convention.auto_detection.enabled__present":     "1",
		"git_convention.auto_detection.enabled":              "1",
		"git_convention.auto_detection.sample_size":          "175",
	})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("reject status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	// Atomic reject — no value persisted (originals survive).
	cfg := loadRawCfg(t, root)
	if cfg.Quality.TestCoverageTarget != 70 {
		t.Errorf("atomic reject leaked test_coverage_target = %d, want 70", cfg.Quality.TestCoverageTarget)
	}
	if cfg.Quality.TDDSettings.MinCoveragePerCommit != 60 {
		t.Errorf("atomic reject leaked min_coverage_per_commit = %d, want 60", cfg.Quality.TDDSettings.MinCoveragePerCommit)
	}
}

// TestProjectNestedWriteFailureSurfacesError covers the nested write-seam error path
// in handleSave: a writeProjectNestedConfig failure (after a successful profile +
// scalar write) surfaces a readable inline error, never blank.
func TestProjectNestedWriteFailureSurfacesError(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	a.writeProjectNestedConfig = func(string, projectNestedForm) error {
		return errors.New("nested config save boom")
	}
	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": "85"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code < 400 {
		t.Errorf("nested write failure status = %d, want >= 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "nested config save boom") {
		t.Errorf("nested write failure should surface the error, got: %s", rec.Body.String())
	}
}

// TestProjectNestedReadFailureRendersInlineError covers the nested read-seam error
// path on GET: a readProjectNestedConfig failure surfaces a readable inline error,
// never blank, never panic.
func TestProjectNestedReadFailureRendersInlineError(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	a.readProjectNestedConfig = func(string) (projectNestedCurrent, error) {
		return projectNestedCurrent{}, errors.New("nested disk read boom")
	}
	rec := serveGet(t, a.routes(), "/")
	if rec.Code < 400 {
		t.Errorf("nested read failure status = %d, want >= 400", rec.Code)
	}
	body := rec.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatal("nested read failure produced a blank page")
	}
	if !strings.Contains(body, "nested disk read boom") {
		t.Errorf("nested read failure should surface the error, got: %s", body)
	}
}

// TestReadProjectNestedConfigDefaults covers readProjectNestedConfig on an absent
// config dir (EC-5): the LoadRaw compiled-in defaults are returned, no panic, no
// error.
func TestReadProjectNestedConfigDefaults(t *testing.T) {
	t.Parallel()
	root := t.TempDir() // no .moai/config/sections
	cur, err := readProjectNestedConfig(root)
	if err != nil {
		t.Fatalf("readProjectNestedConfig on absent config should not error: %v", err)
	}
	// The numeric fields must be formatted strings (never empty) so the numberField
	// value= attribute is always populated.
	if cur.CoverageTarget == "" || cur.MinCoverage == "" || cur.ConfidenceThreshold == "" {
		t.Errorf("absent-config defaults left a numeric field empty: %+v", cur)
	}
}
