package status

import (
	"os"
	"path/filepath"
	"testing"
)

// --- QualityGateStatus struct ---

func TestQualityGateStatusStruct(t *testing.T) {
	status := &QualityGateStatus{
		Tested:    &QualityItem{Passed: true, Message: "Tests present"},
		Readable:  &QualityItem{Passed: true, Message: "No linting issues"},
		Unified:   &QualityItem{Passed: false, Message: "Needs formatting"},
		Secured:   &QualityItem{Passed: true, Message: "No vulnerabilities"},
		Trackable: &QualityItem{Passed: false, Message: "Non-conventional commits found"},
	}

	if !status.Tested.Passed {
		t.Error("Tested should be passed")
	}
	if status.Unified.Passed {
		t.Error("Unified should not be passed")
	}
	if status.Trackable.Passed {
		t.Error("Trackable should not be passed")
	}
}

// --- QualityItem struct ---

func TestQualityItemStruct(t *testing.T) {
	tests := []struct {
		name    string
		item    QualityItem
		passed  bool
		message string
	}{
		{
			name:    "passed item",
			item:    QualityItem{Passed: true, Message: "All good"},
			passed:  true,
			message: "All good",
		},
		{
			name:    "failed item",
			item:    QualityItem{Passed: false, Message: "Issues found"},
			passed:  false,
			message: "Issues found",
		},
		{
			name:    "empty message",
			item:    QualityItem{Passed: true, Message: ""},
			passed:  true,
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item.Passed != tt.passed {
				t.Errorf("Passed = %v, want %v", tt.item.Passed, tt.passed)
			}
			if tt.item.Message != tt.message {
				t.Errorf("Message = %q, want %q", tt.item.Message, tt.message)
			}
		})
	}
}

// --- isConventionalCommit ---

func TestIsConventionalCommit(t *testing.T) {
	tests := []struct {
		message string
		want    bool
	}{
		{"feat: add new feature", true},
		{"fix: resolve bug", true},
		{"docs: update README", true},
		{"style: format code", true},
		{"refactor: restructure module", true},
		{"test: add unit tests", true},
		{"chore: update deps", true},
		{"perf: optimize query", true},
		{"ci: update pipeline", true},
		{"build: update Makefile", true},
		{"revert: undo last change", true},
		{"random commit message", false},
		{"Add new feature", false},
		{"Fixed bug", false},
		{"", false},
		{"   feat: with leading spaces", true},
		{"FEAT: uppercase", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			got := isConventionalCommit(tt.message)
			if got != tt.want {
				t.Errorf("isConventionalCommit(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

// --- checkTestCoverage ---

func TestCheckTestCoverage_WithTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "example_test.go"), []byte("package example"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkTestCoverage(tmpDir)
	if result == nil {
		t.Fatal("checkTestCoverage returned nil")
	}
	if !result.Passed {
		t.Error("expected Passed = true when test files exist")
	}
	if result.Message != "Tests present" {
		t.Errorf("Message = %q, want %q", result.Message, "Tests present")
	}
}

func TestCheckTestCoverage_NoTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file (not a test file)
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkTestCoverage(tmpDir)
	if result == nil {
		t.Fatal("checkTestCoverage returned nil")
	}
	if result.Passed {
		t.Error("expected Passed = false when no test files exist")
	}
	if result.Message != "No tests found" {
		t.Errorf("Message = %q, want %q", result.Message, "No tests found")
	}
}

func TestCheckTestCoverage_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	result := checkTestCoverage(tmpDir)
	if result == nil {
		t.Fatal("checkTestCoverage returned nil")
	}
	if result.Passed {
		t.Error("expected Passed = false for empty dir")
	}
}

func TestCheckTestCoverage_NestedTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory with test file
	nestedDir := filepath.Join(tmpDir, "internal", "pkg")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "handler_test.go"), []byte("package pkg"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkTestCoverage(tmpDir)
	if result == nil {
		t.Fatal("checkTestCoverage returned nil")
	}
	if !result.Passed {
		t.Error("expected Passed = true when nested test files exist")
	}
}

// --- GetQualityGateStatus ---

func TestGetQualityGateStatus_ReturnsAllFields(t *testing.T) {
	tmpDir := t.TempDir()

	status, err := GetQualityGateStatus(tmpDir)
	if err != nil {
		t.Fatalf("GetQualityGateStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetQualityGateStatus returned nil")
	}
	if status.Tested == nil {
		t.Error("Tested is nil")
	}
	if status.Readable == nil {
		t.Error("Readable is nil")
	}
	if status.Unified == nil {
		t.Error("Unified is nil")
	}
	if status.Secured == nil {
		t.Error("Secured is nil")
	}
	if status.Trackable == nil {
		t.Error("Trackable is nil")
	}
}
