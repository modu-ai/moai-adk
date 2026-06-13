// Package harness — regression gate tests (SPEC-HARNESS-REGRESSION-GATE-001 M3).
// Covers MetricTriple, ApplyRegressionError, the atomic baseline store, and the
// metric collector that assembles the triple from the internal/measure parsers.
package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestApplyRegressionError_Error verifies the Error() string names the regressed
// dimensions (REQ-RG-008, C8, DD-3).
func TestApplyRegressionError_Error(t *testing.T) {
	t.Parallel()

	err := &ApplyRegressionError{
		Baseline:  MetricTriple{TestsPassed: 100, Coverage: 87.7, LintCount: 0},
		Candidate: MetricTriple{TestsPassed: 98, Coverage: 80.0, LintCount: 2},
		Regressed: []string{"tests_passed", "coverage", "lint_count"},
	}
	msg := err.Error()
	if !strings.Contains(msg, "tests_passed") {
		t.Errorf("Error() = %q, want it to name regressed dimension tests_passed", msg)
	}
	if !strings.Contains(msg, "coverage") {
		t.Errorf("Error() = %q, want it to name regressed dimension coverage", msg)
	}
	if !strings.Contains(msg, "lint_count") {
		t.Errorf("Error() = %q, want it to name regressed dimension lint_count", msg)
	}
}

// TestApplyRegressionError_DistinctFromPending verifies the new type is distinct
// from ApplyPendingError — errors.As must discriminate them (C8).
func TestApplyRegressionError_DistinctFromPending(t *testing.T) {
	t.Parallel()

	var err error = &ApplyRegressionError{Regressed: []string{"coverage"}}

	var regErr *ApplyRegressionError
	if !errors.As(err, &regErr) {
		t.Fatal("errors.As must match *ApplyRegressionError")
	}
	var pendErr *ApplyPendingError
	if errors.As(err, &pendErr) {
		t.Error("ApplyRegressionError must NOT be assignable to *ApplyPendingError")
	}
}

// TestMetricTriple_Regressions verifies the delta comparison semantics:
// tests_passed decrease OR coverage decrease OR lint_count increase = regression.
func TestMetricTriple_Regressions(t *testing.T) {
	t.Parallel()

	base := MetricTriple{TestsPassed: 100, Coverage: 85.0, LintCount: 1}

	tests := []struct {
		name      string
		candidate MetricTriple
		want      []string
	}{
		{"identical -> no regression", base, nil},
		{"all improved -> no regression", MetricTriple{TestsPassed: 110, Coverage: 90.0, LintCount: 0}, nil},
		{"tests dropped", MetricTriple{TestsPassed: 99, Coverage: 85.0, LintCount: 1}, []string{"tests_passed"}},
		{"coverage dropped", MetricTriple{TestsPassed: 100, Coverage: 84.9, LintCount: 1}, []string{"coverage"}},
		{"lint increased", MetricTriple{TestsPassed: 100, Coverage: 85.0, LintCount: 2}, []string{"lint_count"}},
		{"all regressed", MetricTriple{TestsPassed: 90, Coverage: 80.0, LintCount: 5}, []string{"tests_passed", "coverage", "lint_count"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := base.Regressions(tt.candidate)
			if len(got) != len(tt.want) {
				t.Fatalf("Regressions() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Regressions()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestBaselineStore_AbsentFile verifies first-run behavior: an absent baseline
// file yields (absent=true) so the gate treats the candidate as baseline and
// does not block (REQ-RG-005).
func TestBaselineStore_AbsentFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "measurements-baseline.yaml")
	store := NewBaselineStore(path)

	triple, present, err := store.Load()
	if err != nil {
		t.Fatalf("Load() on absent file must not error: %v", err)
	}
	if present {
		t.Error("Load() on absent file must report present=false")
	}
	if (triple != MetricTriple{}) {
		t.Errorf("Load() on absent file must return zero triple, got %+v", triple)
	}
}

// TestBaselineStore_AtomicRoundTrip verifies Save writes atomically (temp+rename)
// and Load round-trips the metric triple (REQ-RG-004).
func TestBaselineStore_AtomicRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "measurements-baseline.yaml")
	store := NewBaselineStore(path)

	want := MetricTriple{TestsPassed: 1234, Coverage: 87.7, LintCount: 0}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// File must exist and contain the YAML keys.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("baseline file not written: %v", err)
	}
	if !strings.Contains(string(data), "tests_passed") {
		t.Errorf("baseline YAML missing tests_passed key: %s", data)
	}

	// No leftover temp file (atomic rename consumed it).
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file after atomic save: %s", e.Name())
		}
	}

	got, present, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !present {
		t.Error("Load() after Save must report present=true")
	}
	if got != want {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

// TestCollector_AssemblesTriple verifies the metric collector assembles the
// {tests_passed, coverage, lint_count} triple from the internal/measure parsers
// via a stub command runner (REQ-RG-001, AC-RG-011).
func TestCollector_AssemblesTriple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Stub measurer returns a fixed triple without running go test.
	stub := stubMeasurer{triple: MetricTriple{TestsPassed: 42, Coverage: 73.5, LintCount: 3}}

	got, err := stub.Measure(dir)
	if err != nil {
		t.Fatalf("Measure() error: %v", err)
	}
	want := MetricTriple{TestsPassed: 42, Coverage: 73.5, LintCount: 3}
	if got != want {
		t.Errorf("Measure() = %+v, want %+v", got, want)
	}
}

// stubMeasurer is a test Measurer returning a fixed triple (or error).
type stubMeasurer struct {
	triple MetricTriple
	err    error
}

func (s stubMeasurer) Measure(projectRoot string) (MetricTriple, error) {
	return s.triple, s.err
}
