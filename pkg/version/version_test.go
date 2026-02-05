package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Default value test (version is set at build time)
	got := GetVersion()
	if got == "" {
		t.Errorf("GetVersion() = %q, want non-empty string", got)
	}
}

func TestGetCommit(t *testing.T) {
	// Default value test
	got := GetCommit()
	if got != "none" {
		t.Errorf("GetCommit() = %q, want %q (default)", got, "none")
	}
}

func TestGetDate(t *testing.T) {
	// Default value test
	got := GetDate()
	if got != "unknown" {
		t.Errorf("GetDate() = %q, want %q (default)", got, "unknown")
	}
}

func TestGetFullVersion(t *testing.T) {
	got := GetFullVersion()

	// Verify format: "VERSION (commit: COMMIT, built: DATE)"
	if !strings.Contains(got, Version) {
		t.Errorf("GetFullVersion() = %q, should contain Version %q", got, Version)
	}
	if !strings.Contains(got, Commit) {
		t.Errorf("GetFullVersion() = %q, should contain Commit %q", got, Commit)
	}
	if !strings.Contains(got, Date) {
		t.Errorf("GetFullVersion() = %q, should contain Date %q", got, Date)
	}
	if !strings.Contains(got, "commit:") {
		t.Errorf("GetFullVersion() = %q, should contain 'commit:'", got)
	}
	if !strings.Contains(got, "built:") {
		t.Errorf("GetFullVersion() = %q, should contain 'built:'", got)
	}
}

func TestGetFullVersion_Format(t *testing.T) {
	// Test exact format with current values
	expected := Version + " (commit: " + Commit + ", built: " + Date + ")"
	got := GetFullVersion()
	if got != expected {
		t.Errorf("GetFullVersion() = %q, want %q", got, expected)
	}
}

func TestVersionVariables_Defaults(t *testing.T) {
	// Test that version variables are set (not empty)
	tests := []struct {
		name     string
		variable string
		notEmpty bool
	}{
		{"Version", Version, true},
		{"Commit", Commit, true},
		{"Date", Date, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.notEmpty && tt.variable == "" {
				t.Errorf("%s = %q, want non-empty string", tt.name, tt.variable)
			}
		})
	}
}

func TestVersionVariables_Modifiable(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origDate := Date

	// Cleanup after test
	defer func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	}()

	// Modify values (simulating ldflags injection)
	Version = "1.0.0"
	Commit = "abc123"
	Date = "2026-01-15"

	// Verify getters return modified values
	if GetVersion() != "1.0.0" {
		t.Errorf("GetVersion() after modification = %q, want %q", GetVersion(), "1.0.0")
	}
	if GetCommit() != "abc123" {
		t.Errorf("GetCommit() after modification = %q, want %q", GetCommit(), "abc123")
	}
	if GetDate() != "2026-01-15" {
		t.Errorf("GetDate() after modification = %q, want %q", GetDate(), "2026-01-15")
	}

	// Verify full version format with modified values
	expected := "1.0.0 (commit: abc123, built: 2026-01-15)"
	if got := GetFullVersion(); got != expected {
		t.Errorf("GetFullVersion() with modified values = %q, want %q", got, expected)
	}
}
