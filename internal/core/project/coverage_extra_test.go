package project

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests targeting uncovered functions in root.go, reporter.go, and validator.go
// to push core/project coverage above 85%.

// resolveSymlinks resolves symlinks for cross-platform comparison (macOS /var -> /private/var).
func resolveSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve symlinks for %q: %v", path, err)
	}
	return resolved
}

func TestFindProjectRoot_WithMoAI(t *testing.T) {
	// This test creates a temp dir with .moai, changes to it, and verifies FindProjectRoot works
	root := resolveSymlinks(t, t.TempDir())
	mkDir(t, root, ".moai")

	// Save and restore current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	foundRoot, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error: %v", err)
	}

	if foundRoot != root {
		t.Errorf("FindProjectRoot() = %q, want %q", foundRoot, root)
	}
}

func TestFindProjectRoot_NoMoAI(t *testing.T) {
	// Create a temp dir without .moai at any level
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	_, err = FindProjectRoot()
	if err == nil {
		t.Error("expected error when not in a MoAI project")
	}
}

func TestFindProjectRootOrCurrent_WithMoAI(t *testing.T) {
	root := resolveSymlinks(t, t.TempDir())
	mkDir(t, root, ".moai")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	foundRoot, err := FindProjectRootOrCurrent()
	if err != nil {
		t.Fatalf("FindProjectRootOrCurrent() error: %v", err)
	}

	if foundRoot != root {
		t.Errorf("FindProjectRootOrCurrent() = %q, want %q", foundRoot, root)
	}
}

func TestFindProjectRootOrCurrent_NoMoAI(t *testing.T) {
	tmpDir := resolveSymlinks(t, t.TempDir())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	foundRoot, err := FindProjectRootOrCurrent()
	if err != nil {
		t.Fatalf("FindProjectRootOrCurrent() error: %v", err)
	}

	if foundRoot != tmpDir {
		t.Errorf("FindProjectRootOrCurrent() = %q, want %q", foundRoot, tmpDir)
	}
}

func TestMustFindProjectRoot_Panics(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic from MustFindProjectRoot when not in project")
		}
	}()

	MustFindProjectRoot()
}

func TestMustFindProjectRoot_Success(t *testing.T) {
	root := resolveSymlinks(t, t.TempDir())
	mkDir(t, root, ".moai")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	// Should not panic
	result := MustFindProjectRoot()
	if result == "" {
		t.Error("MustFindProjectRoot should return non-empty path")
	}
}

func TestNoOpReporter(t *testing.T) {
	r := &NoOpReporter{}

	// All methods should be no-ops (no panic)
	r.StepStart("test", "message")
	r.StepUpdate("update")
	r.StepComplete("complete")
	r.StepError(nil)
}

func TestConsoleReporter(t *testing.T) {
	r := NewConsoleReporter()

	if r == nil {
		t.Fatal("NewConsoleReporter() returned nil")
	}

	// Redirect stdout is complex, so just verify no panic
	r.StepStart("Detection", "analyzing project")
	r.StepStart("Detection", "")
	r.StepUpdate("progress")
	r.StepComplete("done")
	r.StepComplete("")
	r.StepError(os.ErrNotExist)
}

func TestPhaseExecutor_SetReporter(t *testing.T) {
	pe := newTestPhaseExecutor()
	reporter := &NoOpReporter{}

	pe.SetReporter(reporter)

	if pe.reporter != reporter {
		t.Error("SetReporter did not set the reporter")
	}
}

func TestValidateYAMLFiles_UnreadableFile(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	mkDir(t, root, ".moai/config/sections")

	// Create a YAML file then make it unreadable
	yamlPath := filepath.Join(sectionsDir, "test.yaml")
	writeFile(t, root, ".moai/config/sections/test.yaml", "valid: yaml\n")
	if err := os.Chmod(yamlPath, 0000); err != nil {
		t.Skip("cannot change permissions")
	}
	t.Cleanup(func() {
		_ = os.Chmod(yamlPath, 0644)
	})

	v := NewValidator(nil)
	result := &ValidationResult{Valid: true}
	v.(*projectValidator).validateYAMLFiles(sectionsDir, result)

	if result.Valid {
		t.Error("expected invalid result for unreadable YAML file")
	}
}

func TestValidateJSONFile_UnreadableFile(t *testing.T) {
	root := t.TempDir()
	jsonPath := filepath.Join(root, "test.json")
	if err := os.WriteFile(jsonPath, []byte(`{"valid": true}`), 0644); err != nil {
		t.Fatalf("failed to write json: %v", err)
	}
	if err := os.Chmod(jsonPath, 0000); err != nil {
		t.Skip("cannot change permissions")
	}
	t.Cleanup(func() {
		_ = os.Chmod(jsonPath, 0644)
	})

	v := NewValidator(nil)
	result := &ValidationResult{Valid: true}
	v.(*projectValidator).validateJSONFile(jsonPath, result)

	// Should add a warning, not error
	if len(result.Warnings) == 0 {
		t.Error("expected warning for unreadable JSON file")
	}
}

func TestValidateYAMLFiles_UnreadableDir(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, "unreadable")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sectionsDir, 0000); err != nil {
		t.Skip("cannot change permissions")
	}
	t.Cleanup(func() {
		_ = os.Chmod(sectionsDir, 0755)
	})

	v := NewValidator(nil)
	result := &ValidationResult{Valid: true}
	v.(*projectValidator).validateYAMLFiles(sectionsDir, result)

	if len(result.Warnings) == 0 {
		t.Error("expected warning for unreadable directory")
	}
}

func TestFindProjectRoot_ChildDir(t *testing.T) {
	// Test that FindProjectRoot works from a child directory
	root := resolveSymlinks(t, t.TempDir())
	mkDir(t, root, ".moai")
	childDir := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.Chdir(childDir); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}

	foundRoot, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() from child dir error: %v", err)
	}

	if foundRoot != root {
		t.Errorf("FindProjectRoot() from child = %q, want %q", foundRoot, root)
	}
}
