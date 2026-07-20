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
		form.Set("git_convention.auto_detection.sample_size", "150")
		form.Set("git_convention.validation.enforce_on_push__present", "1")
		form.Set("git_convention.validation.enforce_on_push", "1")

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
		if !f.SampleSizeSet || f.SampleSize != 150 {
			t.Errorf("SampleSize = %d set=%v, want 150 true", f.SampleSize, f.SampleSizeSet)
		}
		if !f.EnforceOnPushSet || !f.EnforceOnPush {
			t.Errorf("EnforceOnPush = %v set=%v, want true true", f.EnforceOnPush, f.EnforceOnPushSet)
		}
		if len(f.ParseErrs) != 0 {
			t.Errorf("ParseErrs = %v, want empty", f.ParseErrs)
		}
	})

	t.Run("empty submission leaves *Set false (EC-1)", func(t *testing.T) {
		t.Parallel()
		f := parseProjectNestedForm(postFormRequest(t, url.Values{}))
		if f.CoverageTargetSet || f.MinCoverageSet || f.ConfidenceSet || f.SampleSizeSet {
			t.Errorf("empty form should leave all numeric *Set false, got %+v", f)
		}
		if f.EnforceQualitySet || f.AutoEnabledSet || f.EnforceOnPushSet {
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

// SPEC-DESIGN-MOAIWEBV2-001 M1: TestProjectFieldsetRendersNestedWidgets was removed
// along with the orphan `project` render surface (fieldsetProject). The server-side
// parse path remains covered by TestParseProjectNestedForm above (parseProjectNestedForm
// is PRESERVED per REQ-MWV2-031 — development_mode / git_convention / quality.* stay
// editable via yaml config / CLI, only the unreachable render surface was dropped).
