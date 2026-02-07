package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T, root string)
		wantValid    bool
		wantErrors   []string
		wantWarnings []string
	}{
		{
			name:      "new project is valid",
			setup:     func(t *testing.T, root string) { t.Helper() },
			wantValid: true,
		},
		{
			name: "existing moai project is invalid",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, ".moai/config/sections")
			},
			wantValid:  false,
			wantErrors: []string{"project already initialized"},
		},
		{
			name: "existing claude directory triggers warning",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, ".claude")
			},
			wantValid:    true,
			wantWarnings: []string{".claude/ directory already exists"},
		},
		{
			name: "existing CLAUDE.md triggers warning",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "CLAUDE.md", "# test\n")
			},
			wantValid:    true,
			wantWarnings: []string{"CLAUDE.md already exists"},
		},
		{
			name: "no git directory triggers warning",
			setup: func(t *testing.T, root string) {
				t.Helper()
				// just an empty dir, no .git
			},
			wantValid:    true,
			wantWarnings: []string{"Git repository not detected"},
		},
		{
			name: "with git directory no git warning",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, ".git")
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			v := NewValidator(nil)
			result, err := v.Validate(root)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}

			for _, want := range tt.wantErrors {
				if !containsSubstring(result.Errors, want) {
					t.Errorf("expected error containing %q, got %v", want, result.Errors)
				}
			}

			for _, want := range tt.wantWarnings {
				if !containsSubstring(result.Warnings, want) {
					t.Errorf("expected warning containing %q, got %v", want, result.Warnings)
				}
			}
		})
	}
}

func TestValidate_InvalidRoot(t *testing.T) {
	v := NewValidator(nil)
	_, err := v.Validate("/nonexistent/path/xyz")
	if err == nil {
		t.Fatal("expected error for invalid root")
	}
}

func TestValidateMoAI(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, root string)
		wantValid  bool
		wantErrors []string
	}{
		{
			name: "no moai directory is invalid",
			setup: func(t *testing.T, root string) {
				t.Helper()
			},
			wantValid:  false,
			wantErrors: []string{".moai/ directory not found"},
		},
		{
			name: "complete moai structure is valid",
			setup: func(t *testing.T, root string) {
				t.Helper()
				for _, dir := range requiredMoAIDirs {
					mkDir(t, root, filepath.Join(".moai", dir))
				}
				writeFile(t, root, ".moai/manifest.json", `{"version":"1.0.0","deployed_at":"","files":{}}`)
				writeFile(t, root, ".moai/config/sections/user.yaml", "user:\n  name: test\n")
				for _, dir := range requiredClaudeDirs {
					mkDir(t, root, filepath.Join(".claude", dir))
				}
				writeFile(t, root, "CLAUDE.md", "# test\n")
			},
			wantValid: true,
		},
		{
			name: "missing subdirectory reports error",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, ".moai/config/sections")
				// missing specs, reports, memory, logs
			},
			wantValid:  false,
			wantErrors: []string{"missing required directory"},
		},
		{
			name: "invalid yaml reports error",
			setup: func(t *testing.T, root string) {
				t.Helper()
				for _, dir := range requiredMoAIDirs {
					mkDir(t, root, filepath.Join(".moai", dir))
				}
				writeFile(t, root, ".moai/config/sections/broken.yaml", "invalid: yaml: ][")
			},
			wantValid:  false,
			wantErrors: []string{"invalid YAML"},
		},
		{
			name: "invalid manifest json reports error",
			setup: func(t *testing.T, root string) {
				t.Helper()
				for _, dir := range requiredMoAIDirs {
					mkDir(t, root, filepath.Join(".moai", dir))
				}
				writeFile(t, root, ".moai/manifest.json", "not json{{{")
			},
			wantValid:  false,
			wantErrors: []string{"invalid JSON"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			v := NewValidator(nil)
			result, err := v.ValidateMoAI(root)
			if err != nil {
				t.Fatalf("ValidateMoAI() error = %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v (errors: %v)", result.Valid, tt.wantValid, result.Errors)
			}

			for _, want := range tt.wantErrors {
				if !containsSubstring(result.Errors, want) {
					t.Errorf("expected error containing %q, got %v", want, result.Errors)
				}
			}
		})
	}
}

func TestBackupExistingProject(t *testing.T) {
	t.Run("backs up existing moai directory", func(t *testing.T) {
		root := t.TempDir()
		mkDir(t, root, ".moai/config/sections")
		writeFile(t, root, ".moai/config/sections/user.yaml", "user:\n  name: test\n")

		backupPath, err := BackupExistingProject(root)
		if err != nil {
			t.Fatalf("BackupExistingProject() error = %v", err)
		}

		if backupPath == "" {
			t.Fatal("expected non-empty backup path")
		}

		if !strings.Contains(backupPath, ".moai-backups") {
			t.Errorf("backup path %q does not contain .moai-backups", backupPath)
		}

		// Verify original .moai/ no longer exists
		if dirExists(filepath.Join(root, ".moai")) {
			t.Error("original .moai/ should have been moved")
		}

		// Verify backup exists
		if !dirExists(backupPath) {
			t.Error("backup directory should exist")
		}

		// Verify backup contains the original file
		backupFile := filepath.Join(backupPath, "config", "sections", "user.yaml")
		if !fileExists(backupFile) {
			t.Error("backup should contain config/sections/user.yaml")
		}
	})

	t.Run("returns empty for non-existing moai directory", func(t *testing.T) {
		root := t.TempDir()
		backupPath, err := BackupExistingProject(root)
		if err != nil {
			t.Fatalf("BackupExistingProject() error = %v", err)
		}
		if backupPath != "" {
			t.Errorf("expected empty backup path, got %q", backupPath)
		}
	})
}

func TestValidateMoAI_MissingClaudeDir(t *testing.T) {
	root := t.TempDir()

	// Create complete .moai structure but no .claude
	for _, dir := range requiredMoAIDirs {
		mkDir(t, root, filepath.Join(".moai", dir))
	}
	writeFile(t, root, ".moai/manifest.json", `{"version":"1.0.0","deployed_at":"","files":{}}`)
	writeFile(t, root, ".moai/config/sections/user.yaml", "user:\n  name: test\n")
	writeFile(t, root, "CLAUDE.md", "# test\n")

	v := NewValidator(nil)
	result, err := v.ValidateMoAI(root)
	if err != nil {
		t.Fatalf("ValidateMoAI() error = %v", err)
	}

	// Should be valid but with warnings about missing .claude
	if !containsSubstring(result.Warnings, ".claude/ directory not found") {
		t.Errorf("expected warning about missing .claude, got %v", result.Warnings)
	}
}

func TestValidateMoAI_MissingCLAUDEMD(t *testing.T) {
	root := t.TempDir()

	for _, dir := range requiredMoAIDirs {
		mkDir(t, root, filepath.Join(".moai", dir))
	}
	writeFile(t, root, ".moai/manifest.json", `{"version":"1.0.0","deployed_at":"","files":{}}`)
	// No CLAUDE.md

	v := NewValidator(nil)
	result, err := v.ValidateMoAI(root)
	if err != nil {
		t.Fatalf("ValidateMoAI() error = %v", err)
	}

	if !containsSubstring(result.Warnings, "CLAUDE.md not found") {
		t.Errorf("expected warning about missing CLAUDE.md, got %v", result.Warnings)
	}
}

func TestValidateMoAI_MissingManifest(t *testing.T) {
	root := t.TempDir()

	for _, dir := range requiredMoAIDirs {
		mkDir(t, root, filepath.Join(".moai", dir))
	}
	// No manifest.json

	v := NewValidator(nil)
	result, err := v.ValidateMoAI(root)
	if err != nil {
		t.Fatalf("ValidateMoAI() error = %v", err)
	}

	if !containsSubstring(result.Warnings, "manifest.json not found") {
		t.Errorf("expected warning about missing manifest, got %v", result.Warnings)
	}
}

func TestValidateMoAI_NonYAMLFilesSkipped(t *testing.T) {
	root := t.TempDir()

	for _, dir := range requiredMoAIDirs {
		mkDir(t, root, filepath.Join(".moai", dir))
	}
	writeFile(t, root, ".moai/manifest.json", `{"version":"1.0.0","deployed_at":"","files":{}}`)
	// Valid YAML file
	writeFile(t, root, ".moai/config/sections/user.yaml", "user:\n  name: test\n")
	// Non-YAML file should be skipped
	writeFile(t, root, ".moai/config/sections/readme.txt", "not yaml\n")
	writeFile(t, root, "CLAUDE.md", "# test\n")

	for _, dir := range requiredClaudeDirs {
		mkDir(t, root, filepath.Join(".claude", dir))
	}

	v := NewValidator(nil)
	result, err := v.ValidateMoAI(root)
	if err != nil {
		t.Fatalf("ValidateMoAI() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateMoAI_ClaudeDirMissingSubdir(t *testing.T) {
	root := t.TempDir()

	for _, dir := range requiredMoAIDirs {
		mkDir(t, root, filepath.Join(".moai", dir))
	}
	writeFile(t, root, ".moai/manifest.json", `{"version":"1.0.0","deployed_at":"","files":{}}`)
	writeFile(t, root, "CLAUDE.md", "# test\n")

	// Create .claude but only some subdirs
	mkDir(t, root, ".claude/skills")
	// Missing agents/moai, commands/moai, rules/moai

	v := NewValidator(nil)
	result, err := v.ValidateMoAI(root)
	if err != nil {
		t.Fatalf("ValidateMoAI() error = %v", err)
	}

	if !containsSubstring(result.Warnings, "missing directory: .claude/") {
		t.Errorf("expected warning about missing .claude subdirectory, got %v", result.Warnings)
	}
}

func TestValidateRoot_NotADirectory(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "notadir.txt")
	writeFile(t, root, "notadir.txt", "hello\n")

	err := validateRoot(filePath)
	if err == nil {
		t.Fatal("expected error for file path as root")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error = %q, want 'not a directory'", err)
	}
}

// containsSubstring checks if any element in the slice contains the substring.
func containsSubstring(ss []string, sub string) bool {
	for _, s := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ensureFileContent reads a file and checks it contains a substring.
// Currently unused but kept for future test expansions.
func ensureFileContent(t *testing.T, path, want string) { //nolint:unused
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("file %s does not contain %q", filepath.Base(path), want)
	}
}
