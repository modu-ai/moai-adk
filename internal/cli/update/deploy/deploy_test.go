package deploy

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestCleanMoaiManagedPaths verifies that all MoAI-managed paths are removed
// and that the function handles various edge cases gracefully.
func TestCleanMoaiManagedPaths(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc      func(string) error
		verifyFunc     func(string) error
		expectError   bool
		errorContains string
	}{
		{
			name: "remove all managed paths",
			setupFunc: func(root string) error {
				// Create all managed paths
				paths := []string{
					filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON),
					filepath.Join(root, defs.ClaudeDir, defs.CommandsMoaiSubdir, "test.txt"),
					filepath.Join(root, defs.ClaudeDir, defs.AgentsMoaiSubdir, "agent.md"),
					filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "moai-test"),
					filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "moai-another"),
					filepath.Join(root, defs.ClaudeDir, defs.RulesMoaiSubdir, "rule.md"),
					filepath.Join(root, defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
					filepath.Join(root, defs.ClaudeDir, defs.HooksMoaiSubdir, "hook.sh"),
					filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "config.yaml"),
				}
				for _, p := range paths {
					if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
						return err
					}
					if err := os.WriteFile(p, []byte("test"), 0o644); err != nil {
						return err
					}
				}
				return nil
			},
			verifyFunc: func(root string) error {
				// Verify all managed paths are removed
				paths := []string{
					filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON),
					filepath.Join(root, defs.ClaudeDir, defs.CommandsMoaiSubdir),
					filepath.Join(root, defs.ClaudeDir, defs.AgentsMoaiSubdir),
					filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "moai-test"),
					filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "moai-another"),
					filepath.Join(root, defs.ClaudeDir, defs.RulesMoaiSubdir),
					filepath.Join(root, defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
					filepath.Join(root, defs.ClaudeDir, defs.HooksMoaiSubdir),
					filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir),
				}
				for _, p := range paths {
					if _, err := os.Stat(p); err == nil || !os.IsNotExist(err) {
						t.Errorf("path %s should not exist after cleanup", p)
					}
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "skip non-existent paths gracefully",
			setupFunc: func(root string) error {
				// Only create a few paths, leaving most non-existent
				paths := []string{
					filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON),
					filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir),
				}
				for _, p := range paths {
					if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
						return err
					}
					if err := os.WriteFile(p, []byte("test"), 0o644); err != nil {
						return err
					}
				}
				return nil
			},
			verifyFunc: func(root string) error {
				// Verify created paths are removed
				if _, err := os.Stat(filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON)); err == nil {
					t.Error("settings.json should be removed")
				}
				if _, err := os.Stat(filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)); err == nil {
					t.Error("config dir should be removed")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "glob pattern matches multiple moai skills",
			setupFunc: func(root string) error {
				skillsDir := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir)
				if err := os.MkdirAll(skillsDir, 0o755); err != nil {
					return err
				}
				// Create moai-* files
				for _, name := range []string{"moai-workflow-spec", "moai-foundation-core", "other-skill"} {
					path := filepath.Join(skillsDir, name)
					if err := os.WriteFile(path, []byte("skill"), 0o644); err != nil {
						return err
					}
				}
				return nil
			},
			verifyFunc: func(root string) error {
				skillsDir := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir)
				// moai-* should be removed
				for _, name := range []string{"moai-workflow-spec", "moai-foundation-core"} {
					path := filepath.Join(skillsDir, name)
					if _, err := os.Stat(path); err == nil {
						t.Errorf("moai skill %s should be removed", name)
					}
				}
				// non-moai file should remain
				otherPath := filepath.Join(skillsDir, "other-skill")
				if _, err := os.Stat(otherPath); os.IsNotExist(err) {
					t.Error("non-moai skill should remain")
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			var buf bytes.Buffer

			if err := tt.setupFunc(root); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := CleanMoaiManagedPaths(root, &buf)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("error should contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verifyFunc != nil {
				if err := tt.verifyFunc(root); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// TestCleanMoaiManagedPaths_Output verifies that progress messages are written
// to the output writer.
func TestCleanMoaiManagedPaths_Output(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer

	// Create some managed paths
	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CleanMoaiManagedPaths(root, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should contain progress messages
	if !strings.Contains(output, "Removing") && !strings.Contains(output, "Removed") {
		t.Error("output should contain progress messages")
	}
}

// TestMigrateLegacyMemoryDir verifies the migration from .moai/memory/ to .moai/state/.
func TestMigrateLegacyMemoryDir(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc      func(string) error
		verifyFunc     func(string) error
		expectError   bool
		errorContains string
	}{
		{
			name: "rename legacy to state when state does not exist",
			setupFunc: func(root string) error {
				legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
				if err := os.MkdirAll(legacyDir, 0o755); err != nil {
					return err
				}
				// Create a file in legacy
				testFile := filepath.Join(legacyDir, "test.txt")
				return os.WriteFile(testFile, []byte("test"), 0o644)
			},
			verifyFunc: func(root string) error {
				legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
				stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)

				// Legacy should not exist
				if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
					t.Error("legacy directory should not exist after migration")
				}

				// State should exist with the migrated file
				if _, err := os.Stat(stateDir); os.IsNotExist(err) {
					t.Error("state directory should exist after migration")
				}
				testFile := filepath.Join(stateDir, "test.txt")
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Error("migrated file should exist in state directory")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "remove legacy when both exist",
			setupFunc: func(root string) error {
				legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
				stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)

				// Create both directories
				if err := os.MkdirAll(legacyDir, 0o755); err != nil {
					return err
				}
				if err := os.MkdirAll(stateDir, 0o755); err != nil {
					return err
				}

				// Create files in both
				if err := os.WriteFile(filepath.Join(legacyDir, "legacy.txt"), []byte("legacy"), 0o644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(stateDir, "state.txt"), []byte("state"), 0o644)
			},
			verifyFunc: func(root string) error {
				legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
				stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)

				// Legacy should be removed
				if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
					t.Error("legacy directory should be removed when state exists")
				}

				// State should exist with its original file
				if _, err := os.Stat(stateDir); os.IsNotExist(err) {
					t.Error("state directory should still exist")
				}
				if _, err := os.Stat(filepath.Join(stateDir, "state.txt")); os.IsNotExist(err) {
					t.Error("state file should still exist")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "no-op when neither directory exists",
			setupFunc: func(root string) error {
				// Don't create any directories
				return nil
			},
			verifyFunc: func(root string) error {
				// Both should not exist (no migration occurred)
				legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
				stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)

				if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
					t.Error("legacy directory should not exist")
				}
				if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
					t.Error("state directory should not exist")
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			var buf bytes.Buffer

			if err := tt.setupFunc(root); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := MigrateLegacyMemoryDir(root, &buf)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("error should contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verifyFunc != nil {
				if err := tt.verifyFunc(root); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// TestMigrateLegacyMemoryDir_Output verifies that progress messages are written.
func TestMigrateLegacyMemoryDir_Output(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer

	// Create legacy directory
	legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyMemoryDir(root, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Migrating") && !strings.Contains(output, "Migrated") {
		t.Error("output should contain migration progress messages")
	}
}

// TestScaffoldEvolutionDir verifies that the evolution directory structure is created.
func TestScaffoldEvolutionDir(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc    func(string) error
		verifyFunc   func(string) error
		expectError bool
	}{
		{
			name: "create complete evolution directory structure",
			setupFunc: func(root string) error {
				// Start with empty directory
				return nil
			},
			verifyFunc: func(root string) error {
				evolutionDir := filepath.Join(root, ".moai", "evolution")

				// Verify subdirectories exist
				subdirs := []string{"telemetry", "learnings", "new-skills"}
				for _, subdir := range subdirs {
					dirPath := filepath.Join(evolutionDir, subdir)
					if info, err := os.Stat(dirPath); os.IsNotExist(err) {
						t.Errorf("subdirectory %s should exist", subdir)
					} else if !info.IsDir() {
						t.Errorf("subdirectory %s should be a directory", subdir)
					}

					// Verify .gitkeep exists
					gitkeep := filepath.Join(dirPath, ".gitkeep")
					if info, err := os.Stat(gitkeep); os.IsNotExist(err) {
						t.Errorf(".gitkeep should exist in %s", subdir)
					} else if info.Mode()&0o644 != 0o644 {
						t.Errorf(".gitkeep should have 0644 permissions")
					}
				}

				// Verify manifest.yaml exists with default content
				manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
				content, err := os.ReadFile(manifestPath)
				if os.IsNotExist(err) {
					t.Error("manifest.yaml should exist")
					return nil
				}
				if err != nil {
					return err
				}
				if !strings.Contains(string(content), "schema_version: 1") {
					t.Error("manifest.yaml should contain default schema_version")
				}

				// Verify changelog.md exists with default content
				changelogPath := filepath.Join(evolutionDir, "changelog.md")
				content, err = os.ReadFile(changelogPath)
				if os.IsNotExist(err) {
					t.Error("changelog.md should exist")
					return nil
				}
				if !strings.Contains(string(content), "MoAI Evolution Changelog") {
					t.Error("changelog.md should contain default title")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "preserve existing files without overwriting",
			setupFunc: func(root string) error {
				evolutionDir := filepath.Join(root, ".moai", "evolution")

				// Create subdirectories with custom content
				subdirs := []string{"telemetry", "learnings", "new-skills"}
				for _, subdir := range subdirs {
					dirPath := filepath.Join(evolutionDir, subdir)
					if err := os.MkdirAll(dirPath, 0o755); err != nil {
						return err
					}
					gitkeep := filepath.Join(dirPath, ".gitkeep")
					if err := os.WriteFile(gitkeep, []byte("custom"), 0o644); err != nil {
						return err
					}
				}

				// Create custom manifest.yaml
				manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
				customManifest := "schema_version: 2\ncustom: true\n"
				if err := os.WriteFile(manifestPath, []byte(customManifest), 0o644); err != nil {
					return err
				}

				// Create custom changelog.md
				changelogPath := filepath.Join(evolutionDir, "changelog.md")
				customChangelog := "# Custom Changelog\n\nExisting content.\n"
				if err := os.WriteFile(changelogPath, []byte(customChangelog), 0o644); err != nil {
					return err
				}

				return nil
			},
			verifyFunc: func(root string) error {
				evolutionDir := filepath.Join(root, ".moai", "evolution")

				// Verify .gitkeep files were NOT overwritten
				subdirs := []string{"telemetry", "learnings", "new-skills"}
				for _, subdir := range subdirs {
					gitkeep := filepath.Join(evolutionDir, subdir, ".gitkeep")
					content, err := os.ReadFile(gitkeep)
					if err != nil {
						return err
					}
					if string(content) != "custom" {
						t.Errorf(".gitkeep in %s should not be overwritten", subdir)
					}
				}

				// Verify manifest.yaml was NOT overwritten
				manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
				content, err := os.ReadFile(manifestPath)
				if err != nil {
					return err
				}
				if !strings.Contains(string(content), "custom: true") {
					t.Error("manifest.yaml should not be overwritten")
				}

				// Verify changelog.md was NOT overwritten
				changelogPath := filepath.Join(evolutionDir, "changelog.md")
				content, err = os.ReadFile(changelogPath)
				if err != nil {
					return err
				}
				if !strings.Contains(string(content), "Custom Changelog") {
					t.Error("changelog.md should not be overwritten")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "partial directory structure already exists",
			setupFunc: func(root string) error {
				// Create only telemetry directory
				evolutionDir := filepath.Join(root, ".moai", "evolution")
				telemetryDir := filepath.Join(evolutionDir, "telemetry")
				if err := os.MkdirAll(telemetryDir, 0o755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(telemetryDir, ".gitkeep"), []byte("exists"), 0o644)
			},
			verifyFunc: func(root string) error {
				evolutionDir := filepath.Join(root, ".moai", "evolution")

				// All subdirectories should exist
				subdirs := []string{"telemetry", "learnings", "new-skills"}
				for _, subdir := range subdirs {
					dirPath := filepath.Join(evolutionDir, subdir)
					if _, err := os.Stat(dirPath); os.IsNotExist(err) {
						t.Errorf("subdirectory %s should exist", subdir)
					}
				}

				// telemetry .gitkeep should have original content
				telemetryGitkeep := filepath.Join(evolutionDir, "telemetry", ".gitkeep")
				content, err := os.ReadFile(telemetryGitkeep)
				if err != nil {
					return err
				}
				if string(content) != "exists" {
					t.Error("existing .gitkeep should not be overwritten")
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()

			if err := tt.setupFunc(root); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := ScaffoldEvolutionDir(root)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verifyFunc != nil {
				if err := tt.verifyFunc(root); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// TestCleanMoaiManagedPaths_CallsMigrate verifies that CleanMoaiManagedPaths
// calls MigrateLegacyMemoryDir as part of its cleanup process.
func TestCleanMoaiManagedPaths_CallsMigrate(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer

	// Create legacy memory directory
	legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "test.txt"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run CleanMoaiManagedPaths (which should call MigrateLegacyMemoryDir)
	if err := CleanMoaiManagedPaths(root, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify migration occurred
	stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		t.Error("state directory should exist after CleanMoaiManagedPaths calls MigrateLegacyMemoryDir")
	}

	// Legacy should not exist
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Error("legacy directory should not exist after migration")
	}
}

// TestCleanMoaiManagedPaths_EmptyProject verifies cleanup on empty project.
func TestCleanMoaiManagedPaths_EmptyProject(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer

	// No paths exist - should skip gracefully
	if err := CleanMoaiManagedPaths(root, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should contain skipped/removed messages for non-existent paths
	if !strings.Contains(output, "Skipped") && !strings.Contains(output, "Removed") {
		t.Error("output should contain skipped or removed messages")
	}
}

// TestMigrateLegacyMemoryDir_EmptyProject verifies migration behavior when
// projectRoot is empty (no legacy or state directories).
func TestMigrateLegacyMemoryDir_EmptyProject(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer

	// Neither legacy nor state exist - should be no-op
	if err := MigrateLegacyMemoryDir(root, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestScaffoldEvolutionDir_MultipleRuns verifies that running ScaffoldEvolutionDir
// multiple times is idempotent (doesn't overwrite existing files).
func TestScaffoldEvolutionDir_MultipleRuns(t *testing.T) {
	root := t.TempDir()

	// First run - create everything
	if err := ScaffoldEvolutionDir(root); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	// Second run - should not overwrite
	if err := ScaffoldEvolutionDir(root); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	// Verify custom content is preserved
	evolutionDir := filepath.Join(root, ".moai", "evolution")
	manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	// Should still have the original default content
	if !strings.Contains(string(content), "schema_version: 1") {
		t.Error("manifest content should be preserved")
	}
}
