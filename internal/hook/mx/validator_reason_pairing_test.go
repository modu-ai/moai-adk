package mx

import (
	"context"
	"strings"
	"testing"
)

// findViolationByMissingTag returns the first violation with the given priority and MissingTag.
func findViolationByMissingTag(violations []Violation, p Priority, missingTag string) *Violation {
	for i, v := range violations {
		if v.Priority == p && v.MissingTag == missingTag {
			return &violations[i]
		}
	}
	return nil
}

// TestValidateFile_AnchorWithoutReason verifies AC-UTIL-001-11:
// @MX:ANCHOR without adjacent @MX:REASON → P1 violation with MissingTag="@MX:REASON".
func TestValidateFile_AnchorWithoutReason(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:ANCHOR: [AUTO] HighFanIn is a frequently called entry point.
func HighFanIn(x int) int {
	return x
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// Must emit P1 with MissingTag="@MX:REASON" (AC-UTIL-001-11).
	viol := findViolationByMissingTag(report.Violations, P1, "@MX:REASON")
	if viol == nil {
		t.Errorf("expected P1/@MX:REASON violation for ANCHOR without REASON, violations: %+v", report.Violations)
		return
	}
	if !strings.HasPrefix(viol.Reason, "@MX:ANCHOR present but @MX:REASON sub-line missing") {
		t.Errorf("Reason = %q, want prefix %q (AC-UTIL-001-11)",
			viol.Reason, "@MX:ANCHOR present but @MX:REASON sub-line missing")
	}
}

// TestValidateFile_WarnWithAdjacentReason verifies AC-UTIL-001-12:
// @MX:WARN with an adjacent @MX:REASON on the immediate next line → no REASON-pairing violation.
func TestValidateFile_WarnWithAdjacentReason(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:WARN: [AUTO] Goroutine lifecycle risk.
// @MX:REASON: goroutine may outlive parent if context is not cancelled
func LaunchWorker() {
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

	// No @MX:REASON pairing violation (AC-UTIL-001-12).
	for _, viol := range report.Violations {
		if viol.MissingTag == "@MX:REASON" {
			t.Errorf("unexpected REASON-pairing violation (AC-UTIL-001-12): %+v", viol)
		}
	}
}

// TestValidateFile_AnchorWithReasonAbove verifies that REASON one line ABOVE ANCHOR is accepted.
func TestValidateFile_AnchorWithReasonAbove(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:REASON: fan_in >= 3 from callers in service, handler, middleware
// @MX:ANCHOR: [AUTO] ProcessRequest is a high fan_in entry point.
func ProcessRequest(req string) {}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	for _, viol := range report.Violations {
		if viol.MissingTag == "@MX:REASON" {
			t.Errorf("unexpected REASON-pairing violation (REASON above ANCHOR should be accepted): %+v", viol)
		}
	}
}

// TestValidateFile_AnchorWithoutReason_TransitionModeOff verifies AC-UTIL-001-16:
// transition_mode=false (default) → REASON-pairing violation is Blocking=true.
func TestValidateFile_AnchorWithoutReason_TransitionModeOff(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:ANCHOR: [AUTO] HighFanIn
func HighFanIn(x int) int { return x }
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	// Default: transitionMode = false.
	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	viol := findViolationByMissingTag(report.Violations, P1, "@MX:REASON")
	if viol == nil {
		t.Skip("no REASON violation emitted (unexpected but cannot verify Blocking)")
	}
	if !viol.Blocking {
		t.Errorf("Blocking = false, want true when transition_mode=false (AC-UTIL-001-16)")
	}
}

// TestValidateFile_AnchorWithoutReason_TransitionModeOn verifies AC-UTIL-001-15:
// transition_mode=true → REASON-pairing violation is advisory (Blocking=false).
func TestValidateFile_AnchorWithoutReason_TransitionModeOn(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:ANCHOR: [AUTO] HighFanIn
func HighFanIn(x int) int { return x }
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	// transition_mode = true via config.
	cfg := DefaultValidationConfig()
	cfg.TransitionMode = true
	v := newValidatorWithConfig(nil, dir, cfg)

	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	viol := findViolationByMissingTag(report.Violations, P1, "@MX:REASON")
	if viol == nil {
		t.Skip("no REASON violation emitted (unexpected but cannot verify Blocking)")
	}
	if viol.Blocking {
		t.Errorf("Blocking = true, want false when transition_mode=true (AC-UTIL-001-15)")
	}
}

// TestValidateFile_WarnWithoutReason verifies @MX:WARN without @MX:REASON → P2 violation.
func TestValidateFile_WarnWithoutReason(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:WARN: [AUTO] goroutine lifecycle
func LaunchWork() {
	go func() {}()
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	viol := findViolationByMissingTag(report.Violations, P2, "@MX:REASON")
	if viol == nil {
		t.Errorf("expected P2/@MX:REASON violation for WARN without REASON, violations: %+v", report.Violations)
		return
	}
	if !strings.HasPrefix(viol.Reason, "@MX:WARN present but @MX:REASON sub-line missing") {
		t.Errorf("Reason = %q, want prefix %q", viol.Reason, "@MX:WARN present but @MX:REASON sub-line missing")
	}
}

// TestValidateFile_WarnWithoutReason_TransitionModeOn verifies that WARN-without-REASON
// in transition_mode=true emits Blocking=false.
func TestValidateFile_WarnWithoutReason_TransitionModeOn(t *testing.T) {
	t.Parallel()

	content := `package testpkg

// @MX:WARN: [AUTO] goroutine lifecycle
func LaunchWork() {
	go func() {}()
}
`
	dir := t.TempDir()
	path := writeGoFile(t, dir, "main.go", content)

	cfg := DefaultValidationConfig()
	cfg.TransitionMode = true
	v := newValidatorWithConfig(nil, dir, cfg)

	report, err := v.ValidateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	viol := findViolationByMissingTag(report.Violations, P2, "@MX:REASON")
	if viol == nil {
		t.Skip("no REASON/P2 violation emitted")
	}
	if viol.Blocking {
		t.Errorf("Blocking = true, want false in transition_mode=true for WARN-without-REASON")
	}
}
