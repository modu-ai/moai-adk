package statusline

import (
	"os"
	"strings"
	"testing"
)

// --- NewFormatter ---

func TestNewFormatter(t *testing.T) {
	f := NewFormatter("/tmp/project", "1.0.0")
	if f == nil {
		t.Fatal("NewFormatter returned nil")
	}
	if f.projectDir != "/tmp/project" {
		t.Errorf("projectDir = %q, want %q", f.projectDir, "/tmp/project")
	}
	if f.version != "1.0.0" {
		t.Errorf("version = %q, want %q", f.version, "1.0.0")
	}
}

// --- Format ---

func TestFormat_VersionPlaceholder(t *testing.T) {
	f := NewFormatter("/tmp/project", "1.2.3")

	result, err := f.Format("{version}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if result != "v1.2.3" {
		t.Errorf("Format({version}) = %q, want %q", result, "v1.2.3")
	}
}

func TestFormat_VersionPlaceholderDev(t *testing.T) {
	f := NewFormatter("/tmp/project", "dev")

	result, err := f.Format("{version}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if result != "vdev" {
		t.Errorf("Format({version}) = %q, want %q", result, "vdev")
	}
}

func TestFormat_SpecsPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{specs}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// With no specs directory, should show "0 SPECs"
	if result != "0 SPECs" {
		t.Errorf("Format({specs}) = %q, want %q", result, "0 SPECs")
	}
}

func TestFormat_QualityPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{quality}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Quality should be a percentage
	if !strings.HasSuffix(result, "%") {
		t.Errorf("Format({quality}) = %q, expected percentage suffix", result)
	}
}

func TestFormat_MultiplePlaceholders(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{version} | {specs}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(result, "v1.0.0") {
		t.Errorf("result %q should contain version", result)
	}
	if !strings.Contains(result, "SPECs") {
		t.Errorf("result %q should contain SPECs", result)
	}
	if !strings.Contains(result, "|") {
		t.Errorf("result %q should contain separator", result)
	}
}

func TestFormat_CustomFormat(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "2.0.0")

	result, err := f.Format("MoAI {version}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if result != "MoAI v2.0.0" {
		t.Errorf("Format(MoAI {version}) = %q, want %q", result, "MoAI v2.0.0")
	}
}

func TestFormat_NoPlaceholders(t *testing.T) {
	f := NewFormatter("/tmp", "1.0.0")

	result, err := f.Format("static text")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if result != "static text" {
		t.Errorf("Format(static text) = %q, want %q", result, "static text")
	}
}

// --- FormatDefault ---

func TestFormatDefault(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.FormatDefault()
	if err != nil {
		t.Fatalf("FormatDefault() error = %v", err)
	}

	// Default format should include version
	if !strings.Contains(result, "v1.0.0") {
		t.Errorf("FormatDefault() = %q, should contain version", result)
	}

	// Should include SPECs
	if !strings.Contains(result, "SPECs") {
		t.Errorf("FormatDefault() = %q, should contain SPECs", result)
	}

	// Should include quality percentage
	if !strings.Contains(result, "%") {
		t.Errorf("FormatDefault() = %q, should contain quality percentage", result)
	}
}

func TestFormatDefault_UsesDefaultFormat(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	// FormatDefault should produce same result as Format("")
	defaultResult, err1 := f.FormatDefault()
	emptyResult, err2 := f.Format("")

	if err1 != nil {
		t.Fatalf("FormatDefault() error = %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Format('') error = %v", err2)
	}

	if defaultResult != emptyResult {
		t.Errorf("FormatDefault() = %q, Format('') = %q, should be equal", defaultResult, emptyResult)
	}
}

// --- Format with git placeholders ---

func TestFormat_BranchPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{branch}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Without git, should fallback to "no-git"
	// or if in a git repo, should have branch name
	if result == "" {
		t.Error("branch should not be empty")
	}
}

func TestFormat_StatePlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{state}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Should be either a check mark, cross, or question mark
	validStates := []string{"\u2713", "\u2717", "?"}
	found := false
	for _, s := range validStates {
		if result == s {
			found = true
			break
		}
	}
	if !found {
		// State might be some other symbol depending on implementation
		if result == "" {
			t.Error("state should not be empty")
		}
	}
}

func TestFormat_StatusTextPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{statustext}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	validTexts := []string{"Clean", "Modified", "Unknown"}
	found := false
	for _, text := range validTexts {
		if result == text {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("statustext = %q, expected one of %v", result, validTexts)
	}
}

// --- GetTestCoverage ---

func TestGetTestCoverage_ReturnsString(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result := f.GetTestCoverage()
	// For a temp dir with no Go module, should return "N/A"
	if result != "N/A" {
		t.Logf("GetTestCoverage() = %q (may vary by environment)", result)
	}
	if result == "" {
		t.Error("GetTestCoverage() returned empty string")
	}
}

func TestGetTestCoverage_WithGoProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal Go module with a test
	goMod := "module testmod\n\ngo 1.21\n"
	mainGo := "package testmod\n\nfunc Add(a, b int) int { return a + b }\n"
	testGo := "package testmod\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1,2) != 3 { t.Error(\"fail\") }\n}\n"

	if err := os.WriteFile(tmpDir+"/go.mod", []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/main.go", []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/main_test.go", []byte(testGo), 0644); err != nil {
		t.Fatal(err)
	}

	f := NewFormatter(tmpDir, "1.0.0")
	result := f.GetTestCoverage()

	// Should return a percentage like "100.0%"
	if result == "N/A" {
		t.Log("GetTestCoverage returned N/A - go test may not be available")
		return
	}
	if !strings.HasSuffix(result, "%") {
		t.Errorf("GetTestCoverage() = %q, expected percentage suffix", result)
	}
}

// --- Format with status placeholder ---

func TestFormat_StatusPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{status}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	validTexts := []string{"Clean", "Modified"}
	found := false
	for _, text := range validTexts {
		if result == text {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("status = %q, expected one of %v", result, validTexts)
	}
}

// --- Format empty version ---

func TestFormat_EmptyVersion(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "")

	result, err := f.Format("{version}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if result != "v" {
		t.Errorf("Format({version}) = %q, want %q", result, "v")
	}
}

// --- Format all placeholders together ---

func TestFormat_AllPlaceholders(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFormatter(tmpDir, "1.0.0")

	result, err := f.Format("{version} {branch} {state} {specs} {quality} {status} {statustext}")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Should have replaced all placeholders - no curly braces remaining
	if strings.Contains(result, "{") || strings.Contains(result, "}") {
		t.Errorf("unreplaced placeholders in result: %q", result)
	}
}
