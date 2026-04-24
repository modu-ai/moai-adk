package mx

import (
	"context"
	"testing"
)

// TestCountFanIn_SubstringFalsePositive_PreFix is a characterization test that pins
// the pre-fix behavior for IMP-V3U-005 (substring fan-in false positives).
// CHARACTERIZATION: pins pre-fix behavior
// Pre-fix: "New" matches "NewContext","NewValidator","RenewToken" via strings.Count → inflated fan_in → spurious P1.
// Post-fix: word-boundary regex counts only direct calls to "New" → fan_in=0 → no P1.
func TestCountFanIn_SubstringFalsePositive_PreFix(t *testing.T) {
	// CHARACTERIZATION: pins pre-fix behavior
	t.Parallel()

	// mainFile declares "New" — but it is never directly called.
	mainContent := `package testpkg

// New creates a new instance.
func New() string {
	return ""
}
`
	// callerContent has NewContext, NewValidator, RenewToken but ZERO direct calls to New().
	callerContent := `package testpkg

func NewContext() string    { return "" }
func NewValidator() string  { return "" }
func RenewToken() string    { return "" }

func callerA() string { return NewContext() }
func callerB() string { return NewValidator() }
func callerC() string { return RenewToken() }
`
	dir := t.TempDir()
	mainFile := writeGoFile(t, dir, "main.go", mainContent)
	writeGoFile(t, dir, "callers.go", callerContent)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), mainFile)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// Post-fix: no P1 for "New" because word-boundary match finds zero direct callers.
	// Pre-fix: this test FAILS (P1Count > 0 due to substring false positives from
	// NewContext, NewValidator, RenewToken, RenewSession appearing as "New" matches).
	for _, viol := range report.Violations {
		if viol.Priority == P1 && viol.FuncName == "New" {
			t.Errorf("spurious P1 for 'New': word-boundary not enforced (IMP-V3U-005)\nViolation: %+v", viol)
		}
	}
}

// TestCountFanIn_WordBoundaryOnly verifies AC-UTIL-001-02:
// funcName="New" with one direct caller returns fan_in=1 (below threshold=3), no P1.
func TestCountFanIn_WordBoundaryOnly(t *testing.T) {
	t.Parallel()

	mainContent := `package testpkg

// New creates a new instance.
func New() string {
	return ""
}
`
	// callerContent has exactly ONE direct call to New(), plus substring-only matches.
	callerContent := `package testpkg

func callerA() string { return New() }  // one direct call — word-boundary match

func callerB() string { return NewContext() }   // substring only — must NOT count
func callerC() string { return NewValidator() } // substring only — must NOT count
func callerD() string { return RenewToken() }  // substring only — must NOT count

func NewContext() string    { return "" }
func NewValidator() string  { return "" }
func RenewToken() string    { return "" }
`
	dir := t.TempDir()
	mainFile := writeGoFile(t, dir, "main.go", mainContent)
	writeGoFile(t, dir, "callers.go", callerContent)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), mainFile)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// With exactly 1 direct caller, fan_in=1 < threshold=3 → no P1.
	for _, viol := range report.Violations {
		if viol.Priority == P1 && viol.FuncName == "New" {
			t.Errorf("P1 for 'New' with fan_in=1: expected no violation (AC-UTIL-001-02)\nViolation: %+v", viol)
		}
	}
}

// TestCountFanIn_ThreeDirectCallers verifies that exactly 3 direct callers triggers P1.
func TestCountFanIn_ThreeDirectCallers(t *testing.T) {
	t.Parallel()

	mainContent := `package testpkg

// Serve handles requests.
func Serve() {}
`
	// Three explicit direct callers (word-boundary matches).
	callerContent := `package testpkg

func handlerA() { Serve() }
func handlerB() { Serve() }
func handlerC() { Serve() }
`
	dir := t.TempDir()
	mainFile := writeGoFile(t, dir, "main.go", mainContent)
	writeGoFile(t, dir, "callers.go", callerContent)

	v := NewValidator(nil, dir)
	report, err := v.ValidateFile(context.Background(), mainFile)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// fan_in=3 >= threshold=3 → P1 violation expected.
	found := false
	for _, viol := range report.Violations {
		if viol.Priority == P1 && viol.FuncName == "Serve" {
			found = true
			if viol.FanIn < 3 {
				t.Errorf("FanIn = %d, want >= 3", viol.FanIn)
			}
		}
	}
	if !found {
		t.Errorf("expected P1 for Serve with 3 direct callers, violations: %+v", report.Violations)
	}
}
