// Package harness — M2-lite non-regression gate (SPEC-HARNESS-REGRESSION-GATE-001).
//
// This file holds the measurement-infrastructure scaffold for the in-Apply
// non-regression gate: the project-health metric triple, the new
// ApplyRegressionError type, the atomic baseline store, and the metric collector
// that assembles the triple from the internal/measure parsers.
//
// HONEST FRAMING (spec.md §A.2 / plan.md DD-7): for the current markdown-only
// harness write surface, the measured delta is typically Δ=0 (the gate is
// always-pass). The gate's genuine value is (1) a measurement-infrastructure
// scaffold that Phase5 reuses and (2) a dormant defense-in-depth safety net
// that fires only if the FROZEN allowlist is widened OR an applier defect writes
// outside the allowlist into tested Go/template code. It is NOT an active
// regression preventer for current harness operation.
package harness

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/measure"
	"gopkg.in/yaml.v3"
)

// MetricTriple is the project-health signal measured by the regression gate:
// test pass count, coverage percentage, and lint (go vet) line count.
//
// @MX:ANCHOR: [AUTO] MetricTriple is the shared measurement record of the gate.
// @MX:REASON: [AUTO] fan_in >= 3: regression_gate.go, applier.go (gate seam), regression_gate_test.go
type MetricTriple struct {
	TestsPassed int     `yaml:"tests_passed"`
	Coverage    float64 `yaml:"coverage"`
	LintCount   int     `yaml:"lint_count"`
}

// Regressions returns the list of dimensions in which `candidate` regressed
// relative to the receiver baseline. A dimension regresses when:
//   - tests_passed decreased, OR
//   - coverage decreased, OR
//   - lint_count increased.
//
// The returned slice is in canonical order (tests_passed, coverage, lint_count)
// and is nil when no dimension regressed (Δ ≥ 0 for all). No tolerance band is
// applied (tolerance is a Phase5 deferral per plan.md R2).
func (base MetricTriple) Regressions(candidate MetricTriple) []string {
	var regressed []string
	if candidate.TestsPassed < base.TestsPassed {
		regressed = append(regressed, "tests_passed")
	}
	if candidate.Coverage < base.Coverage {
		regressed = append(regressed, "coverage")
	}
	if candidate.LintCount > base.LintCount {
		regressed = append(regressed, "lint_count")
	}
	return regressed
}

// ApplyRegressionError is returned when the in-Apply non-regression gate detects
// a project-health regression and rolls back the applied change. It is a type
// DISTINCT from ApplyPendingError (applier.go) — the latter signals the L5
// human-gate pending case, this one signals a post-apply regression block.
//
// @MX:ANCHOR: [AUTO] ApplyRegressionError is the gate→orchestrator boundary type.
// @MX:REASON: [AUTO] fan_in >= 3: regression_gate.go, applier.go (Apply gate), regression_gate_test.go
type ApplyRegressionError struct {
	// Baseline is the metric triple measured BEFORE the apply.
	Baseline MetricTriple
	// Candidate is the metric triple measured AFTER the apply.
	Candidate MetricTriple
	// Regressed is the list of dimensions that regressed (e.g. ["coverage", "lint_count"]).
	Regressed []string
}

func (e *ApplyRegressionError) Error() string {
	regressed := "none"
	if len(e.Regressed) > 0 {
		regressed = ""
		for i, d := range e.Regressed {
			if i > 0 {
				regressed += ", "
			}
			regressed += d
		}
	}
	return fmt.Sprintf("apply: non-regression gate blocked (regressed: %s)", regressed)
}

// ─────────────────────────────────────────────
// Baseline store (atomic YAML read/write)
// ─────────────────────────────────────────────

// baselineFile is the schema persisted to .moai/harness/measurements-baseline.yaml.
// It is a separate, NEW file — the gate MUST NOT touch usage-log.jsonl, the
// lineage manifest.jsonl, observations.yaml, or tier-promotions.jsonl (REQ-RG-006/C11).
type baselineFile struct {
	TestsPassed int     `yaml:"tests_passed"`
	Coverage    float64 `yaml:"coverage"`
	LintCount   int     `yaml:"lint_count"`
	UpdatedAt   string  `yaml:"updated_at"`
}

// BaselineStore persists/loads the regression-gate baseline metric triple to a
// single YAML file using an atomic temp-file + rename write.
type BaselineStore struct {
	path string
}

// NewBaselineStore creates a store bound to the given baseline file path.
// The path is injectable so tests point it at t.TempDir(); the production caller
// passes <harness-dir>/measurements-baseline.yaml.
func NewBaselineStore(path string) *BaselineStore {
	return &BaselineStore{path: path}
}

// Load reads the baseline triple. When the file is absent (first run), it returns
// (zero-triple, present=false, nil) so the caller treats the candidate as the new
// baseline and does not block the Apply (REQ-RG-005).
func (s *BaselineStore) Load() (MetricTriple, bool, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return MetricTriple{}, false, nil
	}
	if err != nil {
		return MetricTriple{}, false, fmt.Errorf("baseline store: read %s: %w", s.path, err)
	}
	var bf baselineFile
	if err := yaml.Unmarshal(data, &bf); err != nil {
		return MetricTriple{}, false, fmt.Errorf("baseline store: parse %s: %w", s.path, err)
	}
	return MetricTriple{
		TestsPassed: bf.TestsPassed,
		Coverage:    bf.Coverage,
		LintCount:   bf.LintCount,
	}, true, nil
}

// Save writes the baseline triple atomically (temp-file + os.Rename, 0o644).
func (s *BaselineStore) Save(triple MetricTriple) error {
	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("baseline store: mkdir %s: %w", dir, err)
		}
	}

	bf := baselineFile{
		TestsPassed: triple.TestsPassed,
		Coverage:    triple.Coverage,
		LintCount:   triple.LintCount,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	data, err := yaml.Marshal(bf)
	if err != nil {
		return fmt.Errorf("baseline store: marshal: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "measurements-baseline-*.tmp")
	if err != nil {
		return fmt.Errorf("baseline store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("baseline store: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("baseline store: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("baseline store: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("baseline store: rename: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────
// Metric collector (Measurer)
// ─────────────────────────────────────────────

// Measurer computes a MetricTriple for a project root. It is an interface so the
// in-Apply gate can be tested without actually running `go test` inside a test —
// a stub Measurer injects fixed triples or a measurement-exec error (REQ-RG-008,
// REQ-RG-014, AC-RG-006/013).
type Measurer interface {
	Measure(projectRoot string) (MetricTriple, error)
}

// goMeasurer is the production Measurer: it runs `go test -json -coverprofile`
// and `go vet`, then assembles the triple via the internal/measure parsers.
//
// HONEST FRAMING: this runs the real toolchain and genuinely measures — for the
// current markdown-only write surface the result is identical pre/post apply
// (Δ=0), but the measurement is not faked or short-circuited.
type goMeasurer struct {
	// timeout bounds each go test / go vet invocation. Richer measurement-error
	// resilience (retries, partial-failure tolerance) is a Phase5 deferral.
	timeout time.Duration
}

// newGoMeasurer creates the production measurer with a default timeout.
func newGoMeasurer() *goMeasurer {
	return &goMeasurer{timeout: 5 * time.Minute}
}

// Measure runs the toolchain and returns the metric triple. It returns a wrapped
// error (fail-closed, REQ-RG-014) when the measurement step cannot execute at all
// — a build error that prevents `go test` from producing any JSON, or a timeout.
// A still-running-but-red suite (tests merely failing) yields a valid tests_passed
// count and is NOT an exec error; it is compared normally.
func (m *goMeasurer) Measure(projectRoot string) (MetricTriple, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	coverFile := filepath.Join(projectRoot, ".moai", "harness", "regression-coverage.out")
	if err := os.MkdirAll(filepath.Dir(coverFile), 0o755); err != nil {
		return MetricTriple{}, fmt.Errorf("measure: cover dir: %w", err)
	}

	// go test -json -coverprofile ./...
	testCmd := exec.CommandContext(ctx, "go", "test", "-count=1", "-json",
		"-coverprofile="+coverFile, "./...")
	testCmd.Dir = projectRoot
	testOut, testErr := testCmd.Output()
	// A non-zero exit from failing tests still produces valid JSON on stdout, so
	// we only treat it as a measurement-exec failure when the context expired
	// (timeout) or stdout is empty (could not run at all → build error).
	if ctx.Err() != nil {
		return MetricTriple{}, fmt.Errorf("measure: go test timed out: %w", ctx.Err())
	}
	if len(testOut) == 0 {
		// No JSON produced — the build did not even reach the test runner.
		return MetricTriple{}, fmt.Errorf("measure: go test produced no output (build error): %w", measurementExecErr(testErr))
	}

	// A build failure (go test -json emits "Action":"build-fail") means the suite
	// could not even compile — that is a measurement-exec failure, NOT a red suite
	// with a valid count. Fail closed per REQ-RG-014.
	if bytes.Contains(testOut, []byte(`"Action":"build-fail"`)) {
		return MetricTriple{}, fmt.Errorf("measure: go test build failure (fail-closed): %w", measurementExecErr(testErr))
	}

	passed, _ := measure.ParseGoTestJSON(testOut)
	coverage := measure.ParseCoverageFile(coverFile)

	// go vet ./... — lint line count from stderr.
	vetCmd := exec.CommandContext(ctx, "go", "vet", "./...")
	vetCmd.Dir = projectRoot
	vetStderr, _ := vetCmd.CombinedOutput()
	if ctx.Err() != nil {
		return MetricTriple{}, fmt.Errorf("measure: go vet timed out: %w", ctx.Err())
	}
	lintCount := measure.CountNonEmptyLines(vetStderr)

	return MetricTriple{
		TestsPassed: passed,
		Coverage:    coverage,
		LintCount:   lintCount,
	}, nil
}

// measurementExecErr normalizes a nil error into a sentinel so the wrapped
// fail-closed error is never "%!w(<nil>)".
func measurementExecErr(err error) error {
	if err == nil {
		return fmt.Errorf("no test output")
	}
	return err
}
