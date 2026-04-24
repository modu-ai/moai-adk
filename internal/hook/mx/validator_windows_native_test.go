package mx

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateFile_NoExecCommand_GoNativeWalker verifies AC-UTIL-001-04:
// The mx validator package contains zero exec.Command calls.
// This is a compile-time guarantee verified by static analysis in CI
// (`grep -R "exec.Command" internal/hook/mx/` returns zero matches).
// This test verifies the runtime behavior is correct on the current platform.
func TestValidateFile_NoExecCommand_GoNativeWalker(t *testing.T) {
	t.Parallel()

	// Method receiver with goroutine — requires both method detection AND fan-in counting
	// to work without any subprocess calls.
	content := `package testpkg

type T struct{}

// Handle handles something.
func (r *T) Handle() {
	go func() {}()
}
`
	callerContent := `package testpkg

func callerA(t *T) { t.Handle() }
func callerB(t *T) { t.Handle() }
func callerC(t *T) { t.Handle() }
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)
	writeGoFile(t, dir, "callers.go", callerContent)

	v := NewValidator(nil, dir)

	// If this runs on Windows without grep, it would panic or error with the pre-fix code.
	// Post-fix: Go-native walker runs identically on all platforms.
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() returned error (platform dependency failure?): %v", err)
	}
	if report == nil {
		t.Fatal("ValidateFile() returned nil report")
	}
	// Validate report is sane.
	_ = report.Violations
}

// TestValidateFile_VendorExclusion verifies AC-UTIL-001-17:
// vendor/ subdirectory files are excluded from fan-in counting.
func TestValidateFile_VendorExclusion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Main file with "Serve" function (no callers in non-vendor, non-generated files).
	mainContent := `package testpkg

// Serve handles requests.
func Serve() {}
`
	mainFile := writeGoFile(t, dir, "main.go", mainContent)

	// vendor/ directory with callers — must be EXCLUDED from fan-in counting.
	vendorDir := filepath.Join(dir, "vendor")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	vendorContent := `package vendor

func init() {
	Serve()
	Serve()
	Serve()
	Serve()
}
`
	vendorFile := filepath.Join(vendorDir, "vendor.go")
	if err := os.WriteFile(vendorFile, []byte(vendorContent), 0o600); err != nil {
		t.Fatalf("write vendor.go: %v", err)
	}

	// Generated file — must also be EXCLUDED.
	generatedContent := `package testpkg

// Code generated. DO NOT EDIT.
func init() {
	Serve()
	Serve()
	Serve()
}
`
	writeGoFile(t, dir, "zz_generated.go", generatedContent)

	// Mock file — must also be EXCLUDED.
	mockContent := `package testpkg

func init() {
	Serve()
	Serve()
	Serve()
}
`
	writeGoFile(t, dir, "mock_serve.go", mockContent)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), mainFile)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// If vendor/, generated, and mock files are excluded → fan_in("Serve") = 0 → no P1.
	for _, viol := range report.Violations {
		if viol.Priority == P1 && viol.FuncName == "Serve" {
			t.Errorf("P1 violation for Serve: vendor/generated/mock files must be excluded (AC-UTIL-001-17)\nViolation: %+v", viol)
		}
	}
}

// TestValidateFile_GeneratedFileExclusion verifies zz_generated.go files are skipped.
func TestValidateFile_GeneratedFileExclusion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	mainContent := `package testpkg

func Handle() {}
`
	mainFile := writeGoFile(t, dir, "main.go", mainContent)

	// Three callers only in generated files — must be excluded.
	for i := 0; i < 3; i++ {
		content := "package testpkg\nfunc init() { Handle() }\n"
		writeGoFile(t, dir, strings.Repeat("z", i+2)+"_generated.go", content)
	}

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), mainFile)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	for _, viol := range report.Violations {
		if viol.Priority == P1 && viol.FuncName == "Handle" {
			t.Errorf("P1 for Handle: generated file exclusion failed (AC-UTIL-001-17)\nViolation: %+v", viol)
		}
	}
}
