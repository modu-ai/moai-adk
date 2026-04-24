package mx

import (
	"context"
	"testing"
)

// TestValidateFile_MethodReceiverBlindspot_PreFix is a characterization test that pins
// the pre-fix behavior for IMP-V3U-001 (method receiver blindspot).
// CHARACTERIZATION: pins pre-fix behavior
// Pre-fix: exported method receivers are invisible to exportedFuncRe → no P2 for goroutine.
// Post-fix: exportedMethodRe detects method receivers → P2 emitted for Handle().
func TestValidateFile_MethodReceiverBlindspot_PreFix(t *testing.T) {
	// CHARACTERIZATION: pins pre-fix behavior
	t.Parallel()

	content := `package testpkg

type FileReport struct{}

// Handle processes a file report.
func (r *FileReport) Handle() {
	go func() {
		// background work
	}()
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// Post-fix: exported method Handle() contains a goroutine → must emit P2 violation.
	// Pre-fix: this test FAILS (P2Count = 0 because method receiver is not detected).
	if report.P2Count() == 0 {
		t.Errorf("P2Count() = 0: exported method receiver goroutine NOT detected (IMP-V3U-001); exportedMethodRe fix required")
	}
	found := false
	for _, viol := range report.Violations {
		if viol.FuncName == "Handle" && viol.Priority == P2 && viol.MissingTag == "@MX:WARN" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected P2/@MX:WARN violation for method receiver Handle(), violations: %+v", report.Violations)
	}
}

// TestValidateFile_MethodReceiver_DetectsGoroutineWithoutWarn verifies AC-UTIL-001-01:
// Given a Go source file containing `func (r *FileReport) Handle() { go func(){}() }` with
// no @MX:WARN tag, When ValidateFile is called, Then the returned FileReport.Violations
// contains exactly one entry with Priority == P2 and FuncName == "Handle".
func TestValidateFile_MethodReceiver_DetectsGoroutineWithoutWarn(t *testing.T) {
	t.Parallel()

	content := `package testpkg

type FileReport struct{}

func (r *FileReport) Handle() {
	go func() {
		// background work
	}()
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// Must detect exactly one P2 violation for Handle (AC-UTIL-001-01).
	p2Warn := 0
	for _, viol := range report.Violations {
		if viol.Priority == P2 && viol.MissingTag == "@MX:WARN" && viol.FuncName == "Handle" {
			p2Warn++
		}
	}
	if p2Warn != 1 {
		t.Errorf("P2/@MX:WARN violations for Handle = %d, want 1 (AC-UTIL-001-01)\nAll violations: %+v",
			p2Warn, report.Violations)
	}
}

// TestValidateFile_MethodReceiver_ValueReceiver verifies value receiver methods are also detected.
func TestValidateFile_MethodReceiver_ValueReceiver(t *testing.T) {
	t.Parallel()

	content := `package testpkg

type Processor struct{}

// Process runs in a goroutine.
func (p Processor) Process() {
	go func() {
		// do work
	}()
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	found := false
	for _, viol := range report.Violations {
		if viol.FuncName == "Process" && viol.Priority == P2 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected P2 violation for value receiver method Process(), violations: %+v", report.Violations)
	}
}

// TestValidateFile_MethodReceiver_WithWarnTag verifies that a method receiver
// with @MX:WARN does NOT emit a P2 violation.
func TestValidateFile_MethodReceiver_WithWarnTag(t *testing.T) {
	t.Parallel()

	content := `package testpkg

type Worker struct{}

// @MX:WARN: [AUTO] goroutine lifecycle risk
// @MX:REASON: goroutine may outlive context if not cancelled
func (w *Worker) Run() {
	go func() {
		// background work
	}()
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	for _, viol := range report.Violations {
		if viol.FuncName == "Run" && viol.Priority == P2 && viol.MissingTag == "@MX:WARN" {
			t.Errorf("unexpected P2/@MX:WARN for Run() which has @MX:WARN tag: %+v", viol)
		}
	}
}
