package web

// projectconfig_branches_test.go — refusal paths of projectconfig.go and
// validate.go that no test reached (Issue #1948, coverage of the hand-written
// code): a non-numeric int field is a per-field parse error that reaches the
// validation result. Tests only.
//
// Defect answer: these tests catch a non-numeric coverage minimum or sample size
// saved as a number, and a parse error that never reaches the rejection.
//
// Not tested, reported instead: readProjectConfig / readProjectNestedConfig error
// branches could not be reached — config.LoadRaw is lenient and reads an
// unreadable quality.yaml (a directory in its place) as defaults with no error, and
// an absent config dir yields the defaults ("tdd", "auto") although the function's
// comment says empty values. A broken section silently reading as defaults is a
// candidate defect, so it is not pinned here.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func nestedFormRequest(t *testing.T, form url.Values) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	return req
}

func TestParseProjectNestedFormRefusesNonNumericInts(t *testing.T) {
	bad := nestedFormRequest(t, url.Values{
		"quality.tdd_settings.min_coverage_per_commit": {"lots"},
		"git_convention.auto_detection.sample_size":    {"many"},
	})
	f := parseProjectNestedForm(bad)
	for _, key := range []string{"quality.tdd_settings.min_coverage_per_commit", "git_convention.auto_detection.sample_size"} {
		if f.ParseErrs[key] != "must be an integer" {
			t.Errorf("%s: parse error %q, want 'must be an integer'", key, f.ParseErrs[key])
		}
	}
	if f.MinCoverageSet || f.SampleSizeSet {
		t.Errorf("a non-numeric value must not be recorded as a set number: %+v", f)
	}
	// The parse errors reach the validation result (the atomic reject).
	errs := validateProjectNestedConfig("", f)
	if errs["quality.tdd_settings.min_coverage_per_commit"] == "" || errs["git_convention.auto_detection.sample_size"] == "" {
		t.Errorf("the parse errors must be part of the rejection: %v", errs)
	}
	// Control: numeric values are set and raise no parse error.
	ok := parseProjectNestedForm(nestedFormRequest(t, url.Values{
		"quality.tdd_settings.min_coverage_per_commit": {"80"},
		"git_convention.auto_detection.sample_size":    {"100"},
	}))
	if !ok.MinCoverageSet || ok.MinCoverage != 80 || !ok.SampleSizeSet || ok.SampleSize != 100 || len(ok.ParseErrs) != 0 {
		t.Errorf("numeric control: %+v", ok)
	}
}
