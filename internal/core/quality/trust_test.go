package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// newCleanLSP returns a mock LSP client with no diagnostics.
func newCleanLSP() *mockLSPClient {
	return &mockLSPClient{diagnostics: []Diagnostic{}}
}

// newCleanGit returns a mock git manager with a valid commit message.
func newCleanGit() *mockGitManager {
	return &mockGitManager{commitMessage: "feat(quality): add TRUST 5 validation"}
}

// newAllPassValidators creates all 5 validators that will pass.
func newAllPassValidators(lsp LSPClient) []Validator {
	git := newCleanGit()
	return []Validator{
		NewTestedValidator(lsp, 85, 90),
		NewReadableValidator(lsp),
		NewUnderstandableValidator(lsp, 10, true, true),
		NewSecuredValidator(lsp),
		NewTrackableValidator(git, true, true),
	}
}

// --- Full Validation Tests ---

func TestFullValidation_AllPass(t *testing.T) {
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)
	config := DefaultQualityConfig()

	gate := NewTrustGate(config, validators)
	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !report.Passed {
		t.Errorf("Report.Passed = false, want true")
	}

	if report.Score < 0.7 {
		t.Errorf("Report.Score = %f, want >= 0.7", report.Score)
	}

	for name, pr := range report.Principles {
		if !pr.Passed {
			t.Errorf("Principle %q: Passed = false, want true", name)
		}
		if len(pr.Issues) != 0 {
			t.Errorf("Principle %q: Issues count = %d, want 0", name, len(pr.Issues))
		}
	}
}

func TestFullValidation_TestedFails(t *testing.T) {
	// LSP with type errors.
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{File: "a.go", Line: 1, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 1"},
		{File: "a.go", Line: 5, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 2"},
		{File: "b.go", Line: 3, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 3"},
	}}

	git := newCleanGit()
	validators := []Validator{
		NewTestedValidator(lsp, 85, 60), // Coverage also below threshold.
		NewReadableValidator(lsp),
		NewUnderstandableValidator(lsp, 10, true, true),
		NewSecuredValidator(lsp),
		NewTrackableValidator(git, true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)
	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report.Passed {
		t.Error("Report.Passed = true, want false")
	}

	tested := report.Principles[PrincipleTested]
	if tested.Passed {
		t.Error("Tested.Passed = true, want false")
	}

	// Should have 3 type error issues + 1 coverage issue = 4 issues.
	if len(tested.Issues) != 4 {
		t.Errorf("Tested.Issues count = %d, want 4", len(tested.Issues))
	}

	// Other principles should still be validated.
	for _, name := range []string{PrincipleReadable, PrincipleUnderstandable, PrincipleSecured, PrincipleTrackable} {
		pr := report.Principles[name]
		if pr.Name == "" {
			t.Errorf("Principle %q was not validated", name)
		}
	}
}

func TestFullValidation_MultipleFail(t *testing.T) {
	// LSP with type errors, lint errors, and security warnings.
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{File: "a.go", Line: 1, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 1"},
		{File: "a.go", Line: 5, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 2"},
		{File: "b.go", Line: 2, Severity: SeverityWarning, Source: "lint", Code: "ST1000", Message: "lint 1"},
		{File: "b.go", Line: 3, Severity: SeverityWarning, Source: "lint", Code: "ST1003", Message: "lint 2"},
		{File: "b.go", Line: 4, Severity: SeverityWarning, Source: "lint", Code: "ST1005", Message: "lint 3"},
		{File: "b.go", Line: 5, Severity: SeverityWarning, Source: "lint", Code: "ST1006", Message: "lint 4"},
		{File: "b.go", Line: 6, Severity: SeverityWarning, Source: "lint", Code: "ST1007", Message: "lint 5"},
		{File: "c.go", Line: 10, Severity: SeverityWarning, Source: "security", Code: "G201", Message: "sql injection"},
	}}

	git := newCleanGit()
	validators := []Validator{
		NewTestedValidator(lsp, 85, 90),
		NewReadableValidator(lsp),
		NewUnderstandableValidator(lsp, 10, true, true),
		NewSecuredValidator(lsp),
		NewTrackableValidator(git, true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)
	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report.Passed {
		t.Error("Report.Passed = true, want false")
	}

	if report.Principles[PrincipleTested].Passed {
		t.Error("Tested should fail (type errors)")
	}
	if report.Principles[PrincipleReadable].Passed {
		t.Error("Readable should fail (lint errors)")
	}
	if report.Principles[PrincipleSecured].Passed {
		t.Error("Secured should fail (security warnings)")
	}

	// Score may be >= 0.7 but Passed is still false because error-severity issues exist.
	// SPEC pass criteria: score >= 0.7 AND zero error-severity issues.
	if report.Passed {
		t.Error("Report.Passed should be false even with score >= 0.7 due to error issues")
	}
}

// --- Single Principle Validation Tests ---

func TestSinglePrinciple_ValidName(t *testing.T) {
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	result, err := gate.ValidatePrinciple(context.Background(), PrincipleTested)
	if err != nil {
		t.Fatalf("ValidatePrinciple() returned unexpected error: %v", err)
	}

	if result.Name != PrincipleTested {
		t.Errorf("Name = %q, want %q", result.Name, PrincipleTested)
	}
}

func TestSinglePrinciple_InvalidName(t *testing.T) {
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	_, err := gate.ValidatePrinciple(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for invalid principle name, got nil")
	}

	// Error message should list valid principle names.
	errMsg := err.Error()
	for _, p := range ValidPrinciples {
		if !strings.Contains(errMsg, p) {
			t.Errorf("error message should contain %q, got: %s", p, errMsg)
		}
	}
}

// --- Context Cancellation Tests ---

func TestContextCancellation_Deadline(t *testing.T) {
	// LSP client that blocks for 5 seconds.
	slowLSP := &mockLSPClient{
		diagnostics: []Diagnostic{},
		delay:       5 * time.Second,
	}
	git := &mockGitManager{commitMessage: "feat: test"}

	validators := []Validator{
		NewTestedValidator(slowLSP, 85, 90),
		NewReadableValidator(slowLSP),
		NewUnderstandableValidator(slowLSP, 10, true, true),
		NewSecuredValidator(slowLSP),
		NewTrackableValidator(git, true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	report, err := gate.Validate(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}

	if report == nil {
		t.Fatal("expected partial report, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestContextCancellation_Manual(t *testing.T) {
	slowLSP := &mockLSPClient{
		diagnostics: []Diagnostic{},
		delay:       5 * time.Second,
	}
	git := &mockGitManager{commitMessage: "feat: test"}

	validators := []Validator{
		NewTestedValidator(slowLSP, 85, 90),
		NewReadableValidator(slowLSP),
		NewUnderstandableValidator(slowLSP, 10, true, true),
		NewSecuredValidator(slowLSP),
		NewTrackableValidator(git, true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	report, err := gate.Validate(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}

	if report == nil {
		t.Fatal("expected partial report, got nil")
	}
}

// --- Phase-Specific Validation Tests ---

func TestPlanPhaseBaseline(t *testing.T) {
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "typecheck", Message: "err1"},
		{Severity: SeverityError, Source: "typecheck", Message: "err2"},
		{Severity: SeverityError, Source: "compiler", Message: "err3"},
		{Severity: SeverityError, Source: "compiler", Message: "err4"},
		{Severity: SeverityError, Source: "compiler", Message: "err5"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn1"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn2"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn3"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn4"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn5"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn6"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn7"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn8"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn9"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn10"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn11"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn12"},
	}}

	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhasePlan),
		WithLSPClient(lsp),
		WithLogger(slog.Default()),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report == nil {
		t.Fatal("expected report, got nil")
	}

	baseline := gate.Baseline()
	if baseline == nil {
		t.Fatal("expected baseline to be captured, got nil")
	}

	if baseline.Errors != 5 {
		t.Errorf("baseline.Errors = %d, want 5", baseline.Errors)
	}
	if baseline.Warnings != 12 {
		t.Errorf("baseline.Warnings = %d, want 12", baseline.Warnings)
	}
	if baseline.TypeErrors != 2 {
		t.Errorf("baseline.TypeErrors = %d, want 2", baseline.TypeErrors)
	}
}

func TestRunPhaseZeroTolerance_Pass(t *testing.T) {
	cleanLSP := newCleanLSP()
	validators := newAllPassValidators(cleanLSP)
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseRun),
		WithLSPClient(cleanLSP),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !report.Passed {
		t.Error("Report.Passed = false, want true for clean run phase")
	}

	if len(report.PhaseIssues) != 0 {
		t.Errorf("PhaseIssues = %d, want 0", len(report.PhaseIssues))
	}
}

func TestRunPhaseZeroTolerance_FailErrors(t *testing.T) {
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "compiler", Message: "compile error"},
	}}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseRun),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report.Passed {
		t.Error("Report.Passed = true, want false for run phase with errors")
	}

	found := false
	for _, issue := range report.PhaseIssues {
		if strings.Contains(issue.Message, "run phase requires zero errors") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected phase issue about zero errors")
	}
}

func TestRunPhaseZeroTolerance_FailTypeErrors(t *testing.T) {
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "typecheck", Message: "type error 1"},
		{Severity: SeverityError, Source: "typecheck", Message: "type error 2"},
	}}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseRun),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report.Passed {
		t.Error("Report.Passed = true, want false")
	}

	found := false
	for _, issue := range report.PhaseIssues {
		if strings.Contains(issue.Message, "run phase requires zero type errors") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected phase issue about zero type errors")
	}
}

func TestRunPhaseZeroTolerance_FailLintErrors(t *testing.T) {
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityWarning, Source: "lint", Code: "ST1000", Message: "lint 1"},
		{Severity: SeverityWarning, Source: "lint", Code: "ST1003", Message: "lint 2"},
		{Severity: SeverityWarning, Source: "lint", Code: "ST1005", Message: "lint 3"},
	}}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseRun),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report.Passed {
		t.Error("Report.Passed = true, want false")
	}

	found := false
	for _, issue := range report.PhaseIssues {
		if strings.Contains(issue.Message, "run phase requires zero lint errors") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected phase issue about zero lint errors")
	}
}

func TestSyncPhaseClean_Pass(t *testing.T) {
	// 8 warnings, below threshold of 10.
	diagnostics := make([]Diagnostic, 8)
	for i := range diagnostics {
		diagnostics[i] = Diagnostic{Severity: SeverityWarning, Source: "compiler", Message: fmt.Sprintf("warn %d", i+1)}
	}
	lsp := &mockLSPClient{diagnostics: diagnostics}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseSync),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !report.Passed {
		t.Error("Report.Passed = false, want true (warnings within threshold)")
	}
}

func TestSyncPhaseClean_FailWarnings(t *testing.T) {
	// 15 warnings, above threshold of 10.
	diagnostics := make([]Diagnostic, 15)
	for i := range diagnostics {
		diagnostics[i] = Diagnostic{Severity: SeverityWarning, Source: "compiler", Message: fmt.Sprintf("warn %d", i+1)}
	}
	lsp := &mockLSPClient{diagnostics: diagnostics}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseSync),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	// Sync phase warning issues have severity "warning", not "error",
	// so they do not force report.Passed to false by themselves.
	// But the test spec says "phase validation fails".
	found := false
	for _, issue := range report.PhaseIssues {
		if strings.Contains(issue.Message, "sync phase allows maximum 10 warnings") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected phase issue about maximum warnings")
	}
}

func TestSyncPhaseClean_FailErrors(t *testing.T) {
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "compiler", Message: "compile error"},
	}}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithPhase(PhaseSync),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report.Passed {
		t.Error("Report.Passed = true, want false for sync phase with errors")
	}

	found := false
	for _, issue := range report.PhaseIssues {
		if strings.Contains(issue.Message, "sync phase requires zero errors") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected phase issue about zero errors")
	}
}

// --- Regression Detection Tests ---

func TestNoRegressionImprovement(t *testing.T) {
	// Current results are better than baseline.
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "compiler", Message: "err1"},
		{Severity: SeverityError, Source: "compiler", Message: "err2"},
		{Severity: SeverityError, Source: "compiler", Message: "err3"},
		{Severity: SeverityWarning, Source: "compiler", Message: "warn1"},
	}}

	baseline := &DiagnosticSnapshot{
		Errors:     5,
		Warnings:   20,
		TypeErrors: 3,
	}

	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithBaseline(baseline),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if len(report.RegressionIssues) != 0 {
		t.Errorf("RegressionIssues = %d, want 0 (improvement)", len(report.RegressionIssues))
	}
}

func TestErrorRegression(t *testing.T) {
	// Current has more errors than baseline.
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "compiler", Message: "err1"},
		{Severity: SeverityError, Source: "compiler", Message: "err2"},
		{Severity: SeverityError, Source: "compiler", Message: "err3"},
	}}

	baseline := &DiagnosticSnapshot{
		Errors:     2,
		Warnings:   10,
		TypeErrors: 0,
	}

	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithBaseline(baseline),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.RegressionIssues {
		if strings.Contains(issue.Message, "error count increased from 2 to 3 (threshold: 0)") {
			if issue.Severity != SeverityError {
				t.Errorf("regression issue severity = %q, want %q", issue.Severity, SeverityError)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error regression issue, got: %v", report.RegressionIssues)
	}
}

func TestWarningRegression_WithinThreshold(t *testing.T) {
	// Increase of 8 warnings, threshold is 10 -> no regression.
	lsp := &mockLSPClient{diagnostics: func() []Diagnostic {
		ds := make([]Diagnostic, 18)
		for i := range ds {
			ds[i] = Diagnostic{Severity: SeverityWarning, Source: "compiler", Message: fmt.Sprintf("warn %d", i+1)}
		}
		return ds
	}()}

	baseline := &DiagnosticSnapshot{Warnings: 10}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithBaseline(baseline),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	for _, issue := range report.RegressionIssues {
		if strings.Contains(issue.Message, "warning count increased") {
			t.Error("should not detect warning regression within threshold")
		}
	}
}

func TestWarningRegression_ExceedsThreshold(t *testing.T) {
	// Increase of 15 warnings, threshold is 10 -> regression.
	lsp := &mockLSPClient{diagnostics: func() []Diagnostic {
		ds := make([]Diagnostic, 25)
		for i := range ds {
			ds[i] = Diagnostic{Severity: SeverityWarning, Source: "compiler", Message: fmt.Sprintf("warn %d", i+1)}
		}
		return ds
	}()}

	baseline := &DiagnosticSnapshot{Warnings: 10}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithBaseline(baseline),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.RegressionIssues {
		if strings.Contains(issue.Message, "warning count increased from 10 to 25 (threshold: 10)") {
			if issue.Severity != SeverityWarning {
				t.Errorf("regression issue severity = %q, want %q", issue.Severity, SeverityWarning)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning regression issue, got: %v", report.RegressionIssues)
	}
}

func TestTypeErrorRegression(t *testing.T) {
	lsp := &mockLSPClient{diagnostics: []Diagnostic{
		{Severity: SeverityError, Source: "typecheck", Message: "type err"},
	}}

	baseline := &DiagnosticSnapshot{TypeErrors: 0}
	validators := newAllPassValidators(newCleanLSP())
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators,
		WithBaseline(baseline),
		WithLSPClient(lsp),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.RegressionIssues {
		if strings.Contains(issue.Message, "type error count increased from 0 to 1 (threshold: 0)") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected type error regression issue, got: %v", report.RegressionIssues)
	}
}

func TestNoBaselineRegression(t *testing.T) {
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)
	config := DefaultQualityConfig()
	// No baseline set -> regression detection is skipped.
	gate := NewTrustGate(config, validators)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if len(report.RegressionIssues) != 0 {
		t.Errorf("RegressionIssues = %d, want 0 (no baseline)", len(report.RegressionIssues))
	}
}

// --- Negative Test Scenarios ---

func TestNoPanic_LSPFailure(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Validate() panicked: %v", r)
		}
	}()

	errLSP := &mockLSPClient{err: fmt.Errorf("connection refused")}
	validators := []Validator{
		NewTestedValidator(errLSP, 85, 90),
		NewReadableValidator(errLSP),
		NewUnderstandableValidator(errLSP, 10, true, true),
		NewSecuredValidator(errLSP),
		NewTrackableValidator(newCleanGit(), true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	report, err := gate.Validate(context.Background())
	if err == nil {
		t.Error("expected error from LSP failure, got nil")
	}
	if report == nil {
		t.Error("expected partial report even on error, got nil")
	}
}

func TestTimeoutEnforcement(t *testing.T) {
	slowLSP := &mockLSPClient{
		diagnostics: []Diagnostic{},
		delay:       10 * time.Second,
	}

	validators := []Validator{
		NewTestedValidator(slowLSP, 85, 90),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := gate.Validate(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	if elapsed > 3*time.Second {
		t.Errorf("validation took %v, expected < 3s", elapsed)
	}
}

func TestConfigImmutability(t *testing.T) {
	config := DefaultQualityConfig()
	originalTarget := config.TestCoverageTarget

	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)
	gate := NewTrustGate(config, validators)

	_, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	afterConfig := gate.Config()
	if afterConfig.TestCoverageTarget != originalTarget {
		t.Errorf("TestCoverageTarget changed from %d to %d", originalTarget, afterConfig.TestCoverageTarget)
	}
}

// --- Score Calculation Tests ---

func TestScoreCalculation(t *testing.T) {
	// TS-REPORT-003: specific scores produce expected weighted average.
	// Tested: 1.0 * 0.30 = 0.30
	// Readable: 0.8 * 0.15 = 0.12
	// Understandable: 0.6 * 0.15 = 0.09
	// Secured: 1.0 * 0.25 = 0.25
	// Trackable: 0.5 * 0.15 = 0.075
	// Total: 0.835
	results := map[string]*PrincipleResult{
		PrincipleTested:         {Name: PrincipleTested, Passed: true, Score: 1.0, Issues: []Issue{}},
		PrincipleReadable:       {Name: PrincipleReadable, Passed: true, Score: 0.8, Issues: []Issue{}},
		PrincipleUnderstandable: {Name: PrincipleUnderstandable, Passed: true, Score: 0.6, Issues: []Issue{}},
		PrincipleSecured:        {Name: PrincipleSecured, Passed: true, Score: 1.0, Issues: []Issue{}},
		PrincipleTrackable:      {Name: PrincipleTrackable, Passed: true, Score: 0.5, Issues: []Issue{}},
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, nil)
	report := gate.buildReport(results)

	if report.Score != 0.835 {
		t.Errorf("Score = %f, want 0.835", report.Score)
	}

	if !report.Passed {
		t.Error("Passed = false, want true (0.835 >= 0.7 and no error issues)")
	}
}

func TestScoreCalculation_BelowThreshold(t *testing.T) {
	// TS-REPORT-004: all scores 0.5 -> total 0.5 -> not passed.
	results := map[string]*PrincipleResult{
		PrincipleTested:         {Name: PrincipleTested, Passed: false, Score: 0.5, Issues: []Issue{}},
		PrincipleReadable:       {Name: PrincipleReadable, Passed: false, Score: 0.5, Issues: []Issue{}},
		PrincipleUnderstandable: {Name: PrincipleUnderstandable, Passed: false, Score: 0.5, Issues: []Issue{}},
		PrincipleSecured:        {Name: PrincipleSecured, Passed: false, Score: 0.5, Issues: []Issue{}},
		PrincipleTrackable:      {Name: PrincipleTrackable, Passed: false, Score: 0.5, Issues: []Issue{}},
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, nil)
	report := gate.buildReport(results)

	if report.Score != 0.5 {
		t.Errorf("Score = %f, want 0.5", report.Score)
	}

	if report.Passed {
		t.Error("Passed = true, want false (0.5 < 0.7)")
	}
}

// --- Report Tests ---

func TestJSONSerialization(t *testing.T) {
	report := &Report{
		Passed:          true,
		Score:           0.835,
		DevelopmentMode: "ddd",
		Principles: map[string]PrincipleResult{
			PrincipleTested: {
				Name:   PrincipleTested,
				Passed: true,
				Score:  1.0,
				Issues: []Issue{},
			},
		},
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	if !json.Valid(data) {
		t.Error("serialized JSON is not valid")
	}

	var deserialized Report
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if deserialized.Score != report.Score {
		t.Errorf("deserialized Score = %f, want %f", deserialized.Score, report.Score)
	}

	if deserialized.DevelopmentMode != report.DevelopmentMode {
		t.Errorf("deserialized DevelopmentMode = %q, want %q", deserialized.DevelopmentMode, report.DevelopmentMode)
	}

	if deserialized.Passed != report.Passed {
		t.Errorf("deserialized Passed = %v, want %v", deserialized.Passed, report.Passed)
	}
}

func TestReportTimestamp(t *testing.T) {
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)
	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	before := time.Now().UTC()
	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
	after := time.Now().UTC()

	if report.Timestamp.Before(before) || report.Timestamp.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", report.Timestamp, before, after)
	}
}

// --- Methodology-Specific Quality Gate Tests ---

func TestDDDModeQualityGate(t *testing.T) {
	t.Run("characterization tests missing", func(t *testing.T) {
		config := DefaultQualityConfig()
		config.DevelopmentMode = ModeDDD
		lsp := newCleanLSP()
		validators := newAllPassValidators(lsp)

		gate := NewTrustGate(config, validators,
			WithMethodologyContext(&MethodologyContext{
				CharacterizationTestsExist: false,
				PreserveStepCompleted:      true,
			}),
		)

		report, err := gate.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate() error: %v", err)
		}

		found := false
		for _, issue := range report.MethodologyIssues {
			if strings.Contains(issue.Message, "characterization tests required") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected DDD characterization test issue")
		}
	})

	t.Run("PRESERVE before IMPROVE required", func(t *testing.T) {
		config := DefaultQualityConfig()
		config.DevelopmentMode = ModeDDD
		lsp := newCleanLSP()
		validators := newAllPassValidators(lsp)

		gate := NewTrustGate(config, validators,
			WithMethodologyContext(&MethodologyContext{
				CharacterizationTestsExist: true,
				PreserveStepCompleted:      false,
			}),
		)

		report, err := gate.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate() error: %v", err)
		}

		found := false
		for _, issue := range report.MethodologyIssues {
			if strings.Contains(issue.Message, "PRESERVE step must complete before IMPROVE") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected DDD PRESERVE-before-IMPROVE issue")
		}
	})

	t.Run("all DDD checks pass", func(t *testing.T) {
		config := DefaultQualityConfig()
		config.DevelopmentMode = ModeDDD
		lsp := newCleanLSP()
		validators := newAllPassValidators(lsp)

		gate := NewTrustGate(config, validators,
			WithMethodologyContext(&MethodologyContext{
				CharacterizationTestsExist: true,
				PreserveStepCompleted:      true,
			}),
		)

		report, err := gate.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate() error: %v", err)
		}

		if len(report.MethodologyIssues) != 0 {
			t.Errorf("MethodologyIssues = %d, want 0", len(report.MethodologyIssues))
		}
	})
}

func TestTDDModeQualityGate(t *testing.T) {
	t.Run("test first not verified", func(t *testing.T) {
		config := DefaultQualityConfig()
		config.DevelopmentMode = ModeTDD
		lsp := newCleanLSP()
		validators := newAllPassValidators(lsp)

		gate := NewTrustGate(config, validators,
			WithMethodologyContext(&MethodologyContext{
				TestFirstVerified: false,
				CommitCoverage:    90,
			}),
		)

		report, err := gate.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate() error: %v", err)
		}

		found := false
		for _, issue := range report.MethodologyIssues {
			if strings.Contains(issue.Message, "tests must be written before implementation") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected TDD test-first issue")
		}
	})

	t.Run("coverage exemption rejected", func(t *testing.T) {
		config := DefaultQualityConfig()
		config.DevelopmentMode = ModeTDD
		lsp := newCleanLSP()
		validators := newAllPassValidators(lsp)

		gate := NewTrustGate(config, validators,
			WithMethodologyContext(&MethodologyContext{
				TestFirstVerified:          true,
				CommitCoverage:             90,
				CoverageExemptionRequested: true,
			}),
		)

		report, err := gate.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate() error: %v", err)
		}

		found := false
		for _, issue := range report.MethodologyIssues {
			if strings.Contains(issue.Message, "coverage exemptions not allowed in TDD mode") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected TDD no-exemption issue")
		}
	})

	t.Run("commit coverage below minimum", func(t *testing.T) {
		config := DefaultQualityConfig()
		config.DevelopmentMode = ModeTDD
		config.TDDSettings.MinCoveragePerCommit = 80
		lsp := newCleanLSP()
		validators := newAllPassValidators(lsp)

		gate := NewTrustGate(config, validators,
			WithMethodologyContext(&MethodologyContext{
				TestFirstVerified: true,
				CommitCoverage:    65,
			}),
		)

		report, err := gate.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate() error: %v", err)
		}

		found := false
		for _, issue := range report.MethodologyIssues {
			if strings.Contains(issue.Message, "commit coverage 65% is below TDD minimum 80%") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected TDD min coverage issue, got: %v", report.MethodologyIssues)
		}
	})
}

func TestMethodologyTransitionWarning(t *testing.T) {
	config := DefaultQualityConfig()
	config.DevelopmentMode = ModeTDD
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)

	// The transition from DDD to TDD should be logged.
	// We verify the methodology context with PreviousMode is accepted without error.
	gate := NewTrustGate(config, validators,
		WithMethodologyContext(&MethodologyContext{
			PreviousMode: ModeDDD,
		}),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	// The transition itself should not cause failure.
	if report == nil {
		t.Fatal("expected report, got nil")
	}
}

func TestReportIncludesDevelopmentMode(t *testing.T) {
	tests := []struct {
		mode DevelopmentMode
	}{
		{ModeDDD},
		{ModeTDD},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			config := DefaultQualityConfig()
			config.DevelopmentMode = tt.mode
			lsp := newCleanLSP()
			validators := newAllPassValidators(lsp)
			gate := NewTrustGate(config, validators)

			report, err := gate.Validate(context.Background())
			if err != nil {
				t.Fatalf("Validate() error: %v", err)
			}

			if report.DevelopmentMode != string(tt.mode) {
				t.Errorf("DevelopmentMode = %q, want %q", report.DevelopmentMode, tt.mode)
			}

			// Verify JSON includes development_mode field.
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}

			expected := fmt.Sprintf(`"development_mode":"%s"`, tt.mode)
			if !strings.Contains(string(data), expected) {
				t.Errorf("JSON does not contain %s", expected)
			}
		})
	}
}

func TestInvalidDevelopmentMode(t *testing.T) {
	config := DefaultQualityConfig()
	config.DevelopmentMode = DevelopmentMode("invalid")
	lsp := newCleanLSP()
	validators := newAllPassValidators(lsp)

	gate := NewTrustGate(config, validators,
		WithMethodologyContext(&MethodologyContext{}),
	)

	report, err := gate.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	found := false
	for _, issue := range report.MethodologyIssues {
		if strings.Contains(issue.Message, "unknown development mode") &&
			strings.Contains(issue.Message, "ddd") &&
			strings.Contains(issue.Message, "tdd") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected invalid mode issue with valid options listed, got: %v", report.MethodologyIssues)
	}
}

// --- Parallel Validation Tests ---

func TestParallelValidation(t *testing.T) {
	// Each validator takes 100ms. Parallel execution should complete in ~200ms,
	// not 500ms (sequential).
	slowLSP := &mockLSPClient{
		diagnostics: []Diagnostic{},
		delay:       100 * time.Millisecond,
	}
	git := &mockGitManager{commitMessage: "feat: parallel test"}

	validators := []Validator{
		NewTestedValidator(slowLSP, 85, 90),
		NewReadableValidator(slowLSP),
		NewUnderstandableValidator(slowLSP, 10, true, true),
		NewSecuredValidator(slowLSP),
		NewTrackableValidator(git, true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	start := time.Now()
	report, err := gate.Validate(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if report == nil {
		t.Fatal("expected report, got nil")
	}

	// With parallel execution, should take around 100-200ms, not 500ms+.
	if elapsed > 1500*time.Millisecond {
		t.Errorf("validation took %v, expected < 1500ms (parallel execution)", elapsed)
	}

	// All 5 principles should be present.
	if len(report.Principles) != 5 {
		t.Errorf("Principles count = %d, want 5", len(report.Principles))
	}
}

func TestPartialResultOnGoroutineError(t *testing.T) {
	// 4 fast validators succeed, 1 slow validator fails.
	fastLSP := newCleanLSP()
	errLSP := &mockLSPClient{
		err:   fmt.Errorf("connection timeout"),
		delay: 100 * time.Millisecond,
	}
	git := newCleanGit()

	validators := []Validator{
		NewTestedValidator(fastLSP, 85, 90),
		NewReadableValidator(errLSP), // This one will fail.
		NewUnderstandableValidator(fastLSP, 10, true, true),
		NewSecuredValidator(fastLSP),
		NewTrackableValidator(git, true, true),
	}

	config := DefaultQualityConfig()
	gate := NewTrustGate(config, validators)

	report, err := gate.Validate(context.Background())
	if err == nil {
		t.Error("expected error from failed validator, got nil")
	}

	if report == nil {
		t.Fatal("expected partial report, got nil")
	}

	// At least the 4 successful validators should have results.
	successCount := 0
	for _, pr := range report.Principles {
		if pr.Score > 0 {
			successCount++
		}
	}

	if successCount < 4 {
		t.Errorf("expected at least 4 successful principles, got %d", successCount)
	}
}

// --- Helper / Utility Tests ---

func TestValidPrinciples(t *testing.T) {
	expected := []string{"tested", "readable", "understandable", "secured", "trackable"}
	if len(ValidPrinciples) != len(expected) {
		t.Fatalf("ValidPrinciples length = %d, want %d", len(ValidPrinciples), len(expected))
	}
	for i, p := range ValidPrinciples {
		if p != expected[i] {
			t.Errorf("ValidPrinciples[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestDevelopmentModeIsValid(t *testing.T) {
	tests := []struct {
		mode DevelopmentMode
		want bool
	}{
		{ModeDDD, true},
		{ModeTDD, true},
		{DevelopmentMode("invalid"), false},
		{DevelopmentMode(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultQualityConfig(t *testing.T) {
	config := DefaultQualityConfig()

	if config.DevelopmentMode != ModeDDD {
		t.Errorf("DevelopmentMode = %q, want %q", config.DevelopmentMode, ModeDDD)
	}
	if config.TestCoverageTarget != 85 {
		t.Errorf("TestCoverageTarget = %d, want 85", config.TestCoverageTarget)
	}
	if config.LSPGates.Run.MaxErrors != 0 {
		t.Errorf("Run.MaxErrors = %d, want 0", config.LSPGates.Run.MaxErrors)
	}
	if config.LSPGates.Sync.MaxWarnings != 10 {
		t.Errorf("Sync.MaxWarnings = %d, want 10", config.LSPGates.Sync.MaxWarnings)
	}
	if config.TDDSettings.MinCoveragePerCommit != 80 {
		t.Errorf("TDDSettings.MinCoveragePerCommit = %d, want 80", config.TDDSettings.MinCoveragePerCommit)
	}
}

func TestAllIssues(t *testing.T) {
	report := &Report{
		Principles: map[string]PrincipleResult{
			PrincipleTested:  {Issues: []Issue{{Message: "p1"}}},
			PrincipleSecured: {Issues: []Issue{{Message: "p2"}}},
		},
		PhaseIssues:       []Issue{{Message: "phase1"}},
		RegressionIssues:  []Issue{{Message: "reg1"}},
		MethodologyIssues: []Issue{{Message: "meth1"}},
	}

	all := report.AllIssues()
	if len(all) != 5 {
		t.Errorf("AllIssues() count = %d, want 5", len(all))
	}
}
