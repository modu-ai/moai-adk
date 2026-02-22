package template

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidatorValidateJSON(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errIs   error
	}{
		{"valid_object", []byte(`{"key":"value"}`), false, nil},
		{"valid_array", []byte(`[1,2,3]`), false, nil},
		{"valid_string", []byte(`"hello"`), false, nil},
		{"valid_number", []byte(`42`), false, nil},
		{"valid_bool", []byte(`true`), false, nil},
		{"valid_null", []byte(`null`), false, nil},
		{"invalid_unclosed_brace", []byte(`{"key":"value"`), true, ErrInvalidJSON},
		{"invalid_trailing_comma", []byte(`{"key":"value",}`), true, ErrInvalidJSON},
		{"invalid_random_text", []byte(`not json at all`), true, ErrInvalidJSON},
		{"empty_input", []byte{}, true, ErrInvalidJSON},
		{"nil_input", nil, true, ErrInvalidJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateJSON(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected %v, got: %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatorValidatePaths(t *testing.T) {
	v := NewValidator()
	root := t.TempDir() // Use platform-specific temp directory

	t.Run("valid_paths", func(t *testing.T) {
		files := []string{
			".claude/settings.json",
			"CLAUDE.md",
			".moai/config/sections/user.yaml",
		}

		errs := v.ValidatePaths(root, files)
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("path_traversal", func(t *testing.T) {
		files := []string{"../../../etc/passwd"}

		errs := v.ValidatePaths(root, files)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
		if errs[0].Path != "../../../etc/passwd" {
			t.Errorf("Path = %q, want %q", errs[0].Path, "../../../etc/passwd")
		}
	})

	t.Run("absolute_path", func(t *testing.T) {
		// Use a path that's absolute on the current platform
		absPath := filepath.Join(root, "malicious")
		files := []string{absPath}

		errs := v.ValidatePaths(root, files)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("mixed_valid_and_invalid", func(t *testing.T) {
		absPath := filepath.Join(root, "absolute")
		files := []string{
			"CLAUDE.md",
			"../escape",
			".claude/agents/file.md",
			absPath,
		}

		errs := v.ValidatePaths(root, files)
		if len(errs) != 2 {
			t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("empty_list", func(t *testing.T) {
		errs := v.ValidatePaths(root, []string{})
		if len(errs) != 0 {
			t.Errorf("expected 0 errors for empty list, got %d", len(errs))
		}
	})
}

func TestValidatorValidateDeployment(t *testing.T) {
	v := NewValidator()

	t.Run("all_files_present_and_valid", func(t *testing.T) {
		root := t.TempDir()

		// Create valid files
		writeFile(t, root, ".claude/settings.json", []byte(`{"hooks":{}}`))
		writeFile(t, root, "CLAUDE.md", []byte("# MoAI"))

		report := v.ValidateDeployment(root, []string{
			".claude/settings.json",
			"CLAUDE.md",
		})

		if !report.Valid {
			t.Errorf("report.Valid = false, want true; errors: %v", report.Errors)
		}
		if report.FilesChecked != 2 {
			t.Errorf("FilesChecked = %d, want 2", report.FilesChecked)
		}
		if len(report.Errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(report.Errors))
		}
	})

	t.Run("missing_file", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "CLAUDE.md", []byte("# MoAI"))

		report := v.ValidateDeployment(root, []string{
			".claude/settings.json", // does not exist
			"CLAUDE.md",
		})

		if report.Valid {
			t.Error("report.Valid = true, want false for missing file")
		}
		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(report.Errors))
		}
		if report.Errors[0].Path != ".claude/settings.json" {
			t.Errorf("error path = %q, want %q", report.Errors[0].Path, ".claude/settings.json")
		}
	})

	t.Run("invalid_json_file", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, ".claude/settings.json", []byte("{invalid json"))

		report := v.ValidateDeployment(root, []string{
			".claude/settings.json",
		})

		if report.Valid {
			t.Error("report.Valid = true, want false for invalid JSON")
		}
		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(report.Errors))
		}
	})

	t.Run("non_json_file_not_validated_as_json", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "CLAUDE.md", []byte("not json but that's fine"))

		report := v.ValidateDeployment(root, []string{"CLAUDE.md"})

		if !report.Valid {
			t.Errorf("report.Valid = false, want true for non-JSON file")
		}
	})

	t.Run("directory_in_expected_files", func(t *testing.T) {
		root := t.TempDir()

		// Create a directory at the expected path (not a file)
		dirPath := filepath.Join(root, ".claude", "agents")
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		report := v.ValidateDeployment(root, []string{
			".claude/agents",
		})

		// Should be valid (directory is a warning, not an error)
		if !report.Valid {
			t.Errorf("report.Valid = false, want true for directory warning")
		}
		if len(report.Warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(report.Warnings), report.Warnings)
		}
		if report.FilesChecked != 1 {
			t.Errorf("FilesChecked = %d, want 1", report.FilesChecked)
		}
	})

	t.Run("empty_file_list", func(t *testing.T) {
		root := t.TempDir()

		report := v.ValidateDeployment(root, []string{})

		if !report.Valid {
			t.Error("report.Valid = false, want true for empty list")
		}
		if report.FilesChecked != 0 {
			t.Errorf("FilesChecked = %d, want 0", report.FilesChecked)
		}
	})
}

// writeFile is a test helper that creates a file in the project root.
func writeFile(t *testing.T, root, relPath string, content []byte) {
	t.Helper()
	absPath := filepath.Join(root, relPath)
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	if err := os.WriteFile(absPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
}
