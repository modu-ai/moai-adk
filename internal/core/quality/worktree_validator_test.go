package quality

import (
	"context"
	"errors"
	"os"
	"testing"
)

// mockGate implements Gate for testing the worktree validator.
type mockGate struct {
	report *Report
	err    error
}

func (m *mockGate) Validate(_ context.Context) (*Report, error) {
	return m.report, m.err
}

func (m *mockGate) ValidatePrinciple(_ context.Context, _ string) (*PrincipleResult, error) {
	return nil, nil
}

// mustNewWorktreeValidator is a test helper that calls NewWorktreeValidator and fails the test on error.
func mustNewWorktreeValidator(t *testing.T, factory GateFactory, config QualityConfig) *worktreeValidator {
	t.Helper()
	v, err := NewWorktreeValidator(factory, config, nil)
	if err != nil {
		t.Fatalf("NewWorktreeValidator() error = %v", err)
	}
	return v
}

// mockGateFactory returns a GateFactory that always returns the given Gate.
func mockGateFactory(gate Gate) GateFactory {
	return func(_ QualityConfig) Gate {
		return gate
	}
}

func TestNewWorktreeValidator(t *testing.T) {
	t.Parallel()

	gate := &mockGate{}
	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		TestCoverageTarget: 85,
	}

	v, err := NewWorktreeValidator(mockGateFactory(gate), config, nil)
	if err != nil {
		t.Fatalf("NewWorktreeValidator() error = %v", err)
	}
	if v == nil {
		t.Fatal("NewWorktreeValidator returned nil")
	}
}

func TestNewWorktreeValidator_NilGateFactory(t *testing.T) {
	t.Parallel()

	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		TestCoverageTarget: 85,
	}

	_, err := NewWorktreeValidator(nil, config, nil)
	if err == nil {
		t.Fatal("NewWorktreeValidator(nil factory) should return error")
	}
}

func TestWorktreeValidator_Validate_InvalidPath(t *testing.T) {
	t.Parallel()

	gate := &mockGate{}
	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		TestCoverageTarget: 85,
	}
	v := mustNewWorktreeValidator(t, mockGateFactory(gate), config)

	_, err := v.Validate(context.Background(), "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestWorktreeValidator_Validate_NotADirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := dir + "/afile.txt"
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	gate := &mockGate{}
	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		TestCoverageTarget: 85,
	}
	v := mustNewWorktreeValidator(t, mockGateFactory(gate), config)

	_, err := v.Validate(context.Background(), filePath)
	if err == nil {
		t.Fatal("expected error for file path, got nil")
	}
}

func TestWorktreeValidator_Validate_PassingQuality(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gate := &mockGate{
		report: &Report{
			Passed: true,
			Score:  1.0,
			Principles: map[string]PrincipleResult{
				PrincipleTested:   {Name: PrincipleTested, Passed: true, Score: 1.0},
				PrincipleReadable: {Name: PrincipleReadable, Passed: true, Score: 1.0},
			},
		},
	}
	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		EnforceQuality:     true,
		TestCoverageTarget: 85,
	}
	v := mustNewWorktreeValidator(t, mockGateFactory(gate), config)

	report, err := v.Validate(context.Background(), dir)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report == nil {
		t.Fatal("Validate() returned nil report")
	}
	if !report.Passed {
		t.Errorf("Validate() report.Passed = false, want true")
	}
}

func TestWorktreeValidator_Validate_FailingQuality(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gate := &mockGate{
		report: &Report{
			Passed: false,
			Score:  0.5,
			Principles: map[string]PrincipleResult{
				PrincipleTested: {
					Name:   PrincipleTested,
					Passed: false,
					Score:  0.5,
					Issues: []Issue{
						{
							File:     "main.go",
							Line:     10,
							Severity: SeverityError,
							Message:  "undefined: foo",
							Rule:     "type-error",
						},
					},
				},
			},
		},
	}
	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		EnforceQuality:     true,
		TestCoverageTarget: 85,
	}
	v := mustNewWorktreeValidator(t, mockGateFactory(gate), config)

	report, err := v.Validate(context.Background(), dir)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report == nil {
		t.Fatal("Validate() returned nil report")
	}
	if report.Passed {
		t.Errorf("Validate() report.Passed = true, want false (type error present)")
	}
}

func TestWorktreeValidator_ValidateWithConfig_Override(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gate := &mockGate{
		report: &Report{
			Passed: true,
			Score:  1.0,
		},
	}
	defaultConfig := QualityConfig{
		DevelopmentMode:    ModeDDD,
		TestCoverageTarget: 90,
	}
	v := mustNewWorktreeValidator(t, mockGateFactory(gate), defaultConfig)

	overrideConfig := QualityConfig{
		DevelopmentMode:    ModeTDD,
		TestCoverageTarget: 0,
	}

	report, err := v.ValidateWithConfig(context.Background(), dir, overrideConfig)
	if err != nil {
		t.Fatalf("ValidateWithConfig() error = %v", err)
	}
	if report == nil {
		t.Fatal("ValidateWithConfig() returned nil report")
	}
}

func TestWorktreeValidator_Validate_GateError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gate := &mockGate{
		err: context.Canceled,
	}
	config := QualityConfig{
		DevelopmentMode:    ModeTDD,
		TestCoverageTarget: 85,
	}
	v := mustNewWorktreeValidator(t, mockGateFactory(gate), config)

	_, err := v.Validate(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error for gate failure, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("error = %v (gate error propagated through wrapper)", err)
	}
}

func TestValidateWorktreePath_ValidDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := validateWorktreePath(dir); err != nil {
		t.Errorf("validateWorktreePath(%q) = %v, want nil", dir, err)
	}
}

func TestValidateWorktreePath_Nonexistent(t *testing.T) {
	t.Parallel()

	err := validateWorktreePath("/nonexistent/path")
	if err == nil {
		t.Error("validateWorktreePath() = nil, want error")
	}
}

func TestValidateWorktreePath_File(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := dir + "/afile.txt"
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	err := validateWorktreePath(filePath)
	if err == nil {
		t.Error("validateWorktreePath() = nil for file, want error")
	}
}
