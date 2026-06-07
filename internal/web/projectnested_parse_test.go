package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// postFormRequest builds a POST *http.Request whose PostForm is populated from
// the given url.Values (so PostFormValue resolves), for parser-level unit tests.
func postFormRequest(t *testing.T, form url.Values) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}
	return r
}

// TestParseProjectNestedForm covers SPEC-WEB-CONSOLE-007 M3 parsing (REQ-WC7-005/006):
// the 6 curated nested fields parse via explicit dot-path PostFormValue with *Set
// flags, empty=not-submitted (EC-1), bool companion disambiguation, and
// type-conversion guards (ParseErrs).
func TestParseProjectNestedForm(t *testing.T) {
	t.Parallel()

	t.Run("all fields submitted with set flags", func(t *testing.T) {
		t.Parallel()
		form := url.Values{}
		form.Set("quality.test_coverage_target", "85")
		form.Set("quality.enforce_quality__present", "1")
		form.Set("quality.enforce_quality", "1")
		form.Set("quality.tdd_settings.min_coverage_per_commit", "80")
		form.Set("git_convention.auto_detection.confidence_threshold", "0.75")
		form.Set("git_convention.auto_detection.enabled__present", "1")
		form.Set("git_convention.auto_detection.enabled", "1")
		form.Set("git_convention.custom.pattern", "^feat: .+")

		f := parseProjectNestedForm(postFormRequest(t, form))

		if !f.CoverageTargetSet || f.CoverageTarget != 85 {
			t.Errorf("CoverageTarget = %d set=%v, want 85 true", f.CoverageTarget, f.CoverageTargetSet)
		}
		if !f.EnforceQualitySet || !f.EnforceQuality {
			t.Errorf("EnforceQuality = %v set=%v, want true true", f.EnforceQuality, f.EnforceQualitySet)
		}
		if !f.MinCoverageSet || f.MinCoverage != 80 {
			t.Errorf("MinCoverage = %d set=%v, want 80 true", f.MinCoverage, f.MinCoverageSet)
		}
		if !f.ConfidenceSet || f.Confidence != 0.75 {
			t.Errorf("Confidence = %v set=%v, want 0.75 true", f.Confidence, f.ConfidenceSet)
		}
		if !f.AutoEnabledSet || !f.AutoEnabled {
			t.Errorf("AutoEnabled = %v set=%v, want true true", f.AutoEnabled, f.AutoEnabledSet)
		}
		if !f.CustomPatternSet || f.CustomPattern != "^feat: .+" {
			t.Errorf("CustomPattern = %q set=%v, want pattern true", f.CustomPattern, f.CustomPatternSet)
		}
		if len(f.ParseErrs) != 0 {
			t.Errorf("ParseErrs = %v, want empty", f.ParseErrs)
		}
	})

	t.Run("empty submission leaves *Set false (EC-1)", func(t *testing.T) {
		t.Parallel()
		f := parseProjectNestedForm(postFormRequest(t, url.Values{}))
		if f.CoverageTargetSet || f.MinCoverageSet || f.ConfidenceSet || f.CustomPatternSet {
			t.Errorf("empty form should leave all numeric/string *Set false, got %+v", f)
		}
		if f.EnforceQualitySet || f.AutoEnabledSet {
			t.Errorf("empty form (no companion) should leave bool *Set false, got %+v", f)
		}
		if f.touchesQuality() || f.touchesGitConvention() {
			t.Error("empty form should touch neither section")
		}
	})

	t.Run("bool companion present but checkbox absent → false set", func(t *testing.T) {
		t.Parallel()
		form := url.Values{}
		form.Set("quality.enforce_quality__present", "1") // companion only, no checkbox
		f := parseProjectNestedForm(postFormRequest(t, form))
		if !f.EnforceQualitySet {
			t.Error("companion present must set EnforceQualitySet = true")
		}
		if f.EnforceQuality {
			t.Error("companion present + checkbox absent must set EnforceQuality = false")
		}
		if !f.touchesQuality() {
			t.Error("bool companion present must mark quality as touched")
		}
	})

	t.Run("non-integer coverage records type-conversion guard", func(t *testing.T) {
		t.Parallel()
		form := url.Values{}
		form.Set("quality.test_coverage_target", "abc")
		f := parseProjectNestedForm(postFormRequest(t, form))
		if f.CoverageTargetSet {
			t.Error("non-integer must NOT mark CoverageTargetSet true")
		}
		if f.ParseErrs["quality.test_coverage_target"] == "" {
			t.Error("non-integer coverage must record a ParseErrs guard message")
		}
	})

	t.Run("non-numeric confidence records type-conversion guard", func(t *testing.T) {
		t.Parallel()
		form := url.Values{}
		form.Set("git_convention.auto_detection.confidence_threshold", "high")
		f := parseProjectNestedForm(postFormRequest(t, form))
		if f.ConfidenceSet {
			t.Error("non-numeric must NOT mark ConfidenceSet true")
		}
		if f.ParseErrs["git_convention.auto_detection.confidence_threshold"] == "" {
			t.Error("non-numeric confidence must record a ParseErrs guard message")
		}
	})
}

// TestProjectFieldsetRendersNestedWidgets covers SPEC-WEB-CONSOLE-007 M3 fieldset
// extension (REQ-WC7-004/010): the Project fieldset renders the 6 nested widgets —
// two numberFields (coverage targets), two toggles (enforce/auto-detect with hidden
// companions), one number (confidence), one text (custom pattern) — populated from
// the view-model current values.
func TestProjectFieldsetRendersNestedWidgets(t *testing.T) {
	t.Parallel()
	view := pageView{
		FieldErrors:             map[string]string{},
		DevelopmentModes:        developmentModeCanonical,
		Conventions:             conventionCanonical,
		CurTestCoverageTarget:   "85",
		CurEnforceQuality:       true,
		CurMinCoveragePerCommit: "80",
		CurConfidenceThreshold:  "0.75",
		CurAutoDetectionEnabled: false,
		CurCustomPattern:        "^feat: .+",
	}
	body := renderTempl(t, fieldsetProject(view))

	for _, want := range []string{
		// numberFields with persisted values + hints.
		`name="quality.test_coverage_target"`,
		`value="85"`,
		`name="quality.tdd_settings.min_coverage_per_commit"`,
		`value="80"`,
		`name="git_convention.auto_detection.confidence_threshold"`,
		`step="0.01"`,
		// toggles with hidden companions.
		`name="quality.enforce_quality__present"`,
		`name="quality.enforce_quality"`,
		`name="git_convention.auto_detection.enabled__present"`,
		`name="git_convention.auto_detection.enabled"`,
		// custom pattern text field.
		`name="git_convention.custom.pattern"`,
		`value="^feat: .+"`,
		// updated section count.
		`data-i18n="count.project"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("fieldsetProject render missing %q\n--- rendered ---\n%s", want, body)
		}
	}
	// enforce_quality is checked (CurEnforceQuality true), auto_detection.enabled unchecked.
	// Verify the enforce checkbox carries `checked` (its companion + checkbox both render).
	if !strings.Contains(body, `id="quality.enforce_quality" name="quality.enforce_quality" value="1" checked`) {
		t.Errorf("enforce_quality toggle should render checked when CurEnforceQuality=true:\n%s", body)
	}
	// auto_detection.enabled is unchecked → its checkbox must NOT carry `checked`.
	if strings.Contains(body, `id="git_convention.auto_detection.enabled" name="git_convention.auto_detection.enabled" value="1" checked`) {
		t.Errorf("auto_detection.enabled toggle must NOT render checked when CurAutoDetectionEnabled=false:\n%s", body)
	}
}
