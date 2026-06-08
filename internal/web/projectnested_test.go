package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// seedNestedProject writes a temp project root with quality.yaml + git-convention.yaml
// pre-populated with the curated editable nested fields PLUS non-editable sibling
// nested fields, so the HARD-4 nested-isolation tests can prove the siblings survive
// a single-field write. Out-of-scope sections carry the 006 DO_NOT_TOUCH sentinel.
func seedNestedProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		// quality: editable nested (test_coverage_target / enforce_quality /
		// tdd_settings.min_coverage_per_commit) + NON-editable siblings that must survive
		// a single-field write (coverage_exemptions.max_exempt_percentage,
		// tdd_settings.test_first_required, lsp_quality_gates.enabled).
		"quality.yaml": "constitution:\n" +
			"  development_mode: tdd\n" +
			"  enforce_quality: true\n" +
			"  test_coverage_target: 70\n" +
			"  tdd_settings:\n" +
			"    test_first_required: true\n" +
			"    min_coverage_per_commit: 60\n" +
			"  coverage_exemptions:\n" +
			"    max_exempt_percentage: 42\n" +
			"  lsp_quality_gates:\n" +
			"    enabled: true\n",
		// git_convention: editable nested (convention / auto_detection.confidence_threshold /
		// auto_detection.enabled / auto_detection.sample_size / validation.enforce_on_push)
		// + NON-editable sibling (validation.max_length) that must survive a single-field write.
		"git-convention.yaml": "git_convention:\n" +
			"  convention: angular\n" +
			"  auto_detection:\n" +
			"    enabled: true\n" +
			"    confidence_threshold: 0.5\n" +
			"    sample_size: 100\n" +
			"  validation:\n" +
			"    enforce_on_push: false\n" +
			"    max_length: 80\n",
		// Out-of-scope sentinels (HARD-1/HARD-2 — must round-trip unchanged).
		"workflow.yaml":     "workflow:\n  sentinel: DO_NOT_TOUCH\n",
		"harness.yaml":      "harness:\n  sentinel: DO_NOT_TOUCH\n",
		"git-strategy.yaml": "git_strategy:\n  sentinel: DO_NOT_TOUCH\n",
		"llm.yaml":          "llm:\n  mode: \"\"\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

// loadRawCfg loads the project config via the config manager (the same path the
// write seam uses) for post-write assertions.
func loadRawCfg(t *testing.T, root string) *config.Config {
	t.Helper()
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	return cfg
}

// nestedSaveForm builds a POST form carrying the baseline profile fields plus the
// supplied nested key/value pairs. The two scalars (development_mode/git_convention)
// are left empty (preserve) unless overridden via extra.
func nestedSaveForm(extra map[string]string) url.Values {
	form := url.Values{}
	form.Set("__profile", "default")
	for k, v := range extra {
		form.Set(k, v)
	}
	return form
}

// realApp builds an app rooted at a real seeded project with REAL seams (no stubs),
// so the write+read round-trip exercises the actual config-manager path.
func realApp(t *testing.T, root string) *app {
	t.Helper()
	return newApp(Config{ProjectRoot: root, ProfileName: "default"})
}

// TestProjectNestedRoundTrip covers AC-WC7-007 (REQ-WC7-005): a single nested field
// (quality.test_coverage_target=85) submitted via POST /save persists to disk.
func TestProjectNestedRoundTrip(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": "85"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	if cfg.Quality.TestCoverageTarget != 85 {
		t.Errorf("test_coverage_target = %d, want 85 (round-trip)", cfg.Quality.TestCoverageTarget)
	}
}

// TestProjectNestedSiblingPreserved covers AC-WC7-008 (REQ-WC7-005/012, HARD-4):
// writing ONLY quality.test_coverage_target leaves coverage_exemptions.max_exempt_percentage
// + tdd_settings.test_first_required + lsp_quality_gates.enabled byte-identical.
func TestProjectNestedSiblingPreserved(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": "85"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	// Target changed.
	if cfg.Quality.TestCoverageTarget != 85 {
		t.Errorf("target test_coverage_target = %d, want 85", cfg.Quality.TestCoverageTarget)
	}
	// 3+ non-edited siblings preserved.
	if cfg.Quality.CoverageExemptions.MaxExemptPercentage != 42 {
		t.Errorf("SIBLING coverage_exemptions.max_exempt_percentage = %d, want 42 (preserved)", cfg.Quality.CoverageExemptions.MaxExemptPercentage)
	}
	if !cfg.Quality.TDDSettings.TestFirstRequired {
		t.Error("SIBLING tdd_settings.test_first_required = false, want true (preserved)")
	}
	if !cfg.Quality.LSPQualityGates.Enabled {
		t.Error("SIBLING lsp_quality_gates.enabled = false, want true (preserved)")
	}
	// The unedited min_coverage_per_commit (nested-of-nested sibling within TDDSettings) preserved.
	if cfg.Quality.TDDSettings.MinCoveragePerCommit != 60 {
		t.Errorf("SIBLING tdd_settings.min_coverage_per_commit = %d, want 60 (preserved)", cfg.Quality.TDDSettings.MinCoveragePerCommit)
	}
}

// TestProjectNestedGitConventionSiblingPreserved covers AC-WC7-009 + AC-WC9-015
// (REQ-WC7-005, HARD-4): writing ONLY git_convention.auto_detection.confidence_threshold
// leaves the retained siblings (validation.max_length, auto_detection.enabled/sample_size)
// byte-identical.
func TestProjectNestedGitConventionSiblingPreserved(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	form := nestedSaveForm(map[string]string{"git_convention.auto_detection.confidence_threshold": "0.9"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	if cfg.GitConvention.AutoDetection.ConfidenceThreshold != 0.9 {
		t.Errorf("target confidence_threshold = %v, want 0.9", cfg.GitConvention.AutoDetection.ConfidenceThreshold)
	}
	if cfg.GitConvention.Validation.MaxLength != 80 {
		t.Errorf("SIBLING validation.max_length = %d, want 80 (preserved)", cfg.GitConvention.Validation.MaxLength)
	}
	// The unedited auto_detection siblings (enabled, sample_size) preserved.
	if !cfg.GitConvention.AutoDetection.Enabled {
		t.Error("SIBLING auto_detection.enabled = false, want true (preserved)")
	}
	if cfg.GitConvention.AutoDetection.SampleSize != 100 {
		t.Errorf("SIBLING auto_detection.sample_size = %d, want 100 (preserved)", cfg.GitConvention.AutoDetection.SampleSize)
	}
}

// TestProjectNestedEmptyPreserves covers AC-WC7-010 (REQ-WC7-006, EC-1): an empty
// quality.test_coverage_target submission leaves the persisted value unchanged.
func TestProjectNestedEmptyPreserves(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	// Submit empty coverage (and no other nested field) → nothing should change.
	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": ""})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	if cfg.Quality.TestCoverageTarget != 70 {
		t.Errorf("empty submission clobbered test_coverage_target = %d, want 70 (preserved)", cfg.Quality.TestCoverageTarget)
	}
}

// TestProjectNestedToggleEC1 covers AC-WC7-011a (REQ-WC7-006, EC-1 bool): when the
// enforce_quality toggle is NOT submitted (no companion), the persisted bool is
// preserved.
func TestProjectNestedToggleEC1(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	// No companion for enforce_quality → preserve. Submit an unrelated field so the
	// quality section is otherwise touched (proving enforce_quality still rides through).
	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": "88"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	if !cfg.Quality.EnforceQuality {
		t.Error("enforce_quality = false, want true (no companion submitted → EC-1 preserve)")
	}
}

// TestProjectNestedToggleUnchecked covers AC-WC7-011b (REQ-WC7-005, bool change):
// companion present + checkbox unchecked → the bool is persisted as false.
func TestProjectNestedToggleUnchecked(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	// Companion present, checkbox absent → enforce_quality becomes false.
	form := nestedSaveForm(map[string]string{"quality.enforce_quality__present": "1"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	cfg := loadRawCfg(t, root)
	if cfg.Quality.EnforceQuality {
		t.Error("enforce_quality = true, want false (companion + unchecked → set false)")
	}
}

// TestProjectNestedAtomicReject covers AC-WC7-012 (REQ-WC7-007, HARD-5, EC-2): one
// valid nested field + one invalid nested field → 400, no section written (all
// persisted values unchanged), re-render with the per-field error.
func TestProjectNestedAtomicReject(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	// Valid coverage 90 + INVALID min_coverage_per_commit 150 (out of 0-100) → reject all.
	form := nestedSaveForm(map[string]string{
		"quality.test_coverage_target":                 "90",
		"quality.tdd_settings.min_coverage_per_commit": "150",
	})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atomic reject status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "quality.tdd_settings.min_coverage_per_commit") {
		t.Error("response missing the per-field error for min_coverage_per_commit")
	}
	cfg := loadRawCfg(t, root)
	// NEITHER value persisted — original 70 must survive (atomic reject).
	if cfg.Quality.TestCoverageTarget != 70 {
		t.Errorf("atomic reject leaked test_coverage_target = %d, want 70 (no write on any invalid field)", cfg.Quality.TestCoverageTarget)
	}
	if cfg.Quality.TDDSettings.MinCoveragePerCommit != 60 {
		t.Errorf("atomic reject leaked min_coverage_per_commit = %d, want 60", cfg.Quality.TDDSettings.MinCoveragePerCommit)
	}
}

// TestProjectNestedOutOfRangeReject covers AC-WC7-013 (REQ-WC7-014, HARD-3):
// quality.test_coverage_target=150 → 400, a FieldErrors entry for the field, write 0.
func TestProjectNestedOutOfRangeReject(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": "150"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("out-of-range status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "quality.test_coverage_target") {
		t.Error("response missing the per-field error for test_coverage_target")
	}
	if !strings.Contains(body, "must be between 0 and 100") {
		t.Error("response missing the reused range message")
	}
	cfg := loadRawCfg(t, root)
	if cfg.Quality.TestCoverageTarget != 70 {
		t.Errorf("out-of-range write leaked test_coverage_target = %d, want 70 (no write)", cfg.Quality.TestCoverageTarget)
	}
}

// TestProjectNestedCustomConventionRejected covers AC-WC9-009 (REQ-WC9-003): the
// `custom` engine is removed, so submitting git_convention=custom is rejected at the
// 4-value enum validator (no custom.pattern concept) → 400, FieldErrors for
// git_convention, write 0 (atomic reject leaves the persisted convention angular).
func TestProjectNestedCustomConventionRejected(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	// Set convention=custom (scalar) → enum-rejected (custom engine removed).
	form := nestedSaveForm(map[string]string{"git_convention": "custom"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("custom-rejected status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "unrecognized git convention") {
		t.Error("response missing the enum-reject message for convention=custom")
	}
	cfg := loadRawCfg(t, root)
	// convention scalar must NOT be persisted (atomic reject leaves angular).
	if cfg.GitConvention.Convention != "angular" {
		t.Errorf("custom-rejected leaked convention = %q, want angular (no write)", cfg.GitConvention.Convention)
	}
}

// TestSaveNestedFullPage covers AC-WC7-019 (REQ-WC7-009): the /save response is the
// SAME full HTML page (hx-boost full-page swap), NOT a section-scoped partial
// fragment.
func TestSaveNestedFullPage(t *testing.T) {
	t.Parallel()
	root := seedNestedProject(t)
	a := realApp(t, root)

	form := nestedSaveForm(map[string]string{"quality.test_coverage_target": "85"})
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// Templ normalizes the doctype to lowercase (<!doctype html>); match
	// case-insensitively (coverage_test.go convention).
	if !strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Error("/save response must be a full HTML page (<!doctype html>), not a partial fragment")
	}
	if !strings.Contains(body, `hx-boost="true"`) {
		t.Error("/save response must preserve the hx-boost full-page form")
	}
}
