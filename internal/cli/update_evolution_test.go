package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScaffoldEvolutionDir_CreatesDirectoryTree verifies that scaffoldEvolutionDir
// creates the required subdirectories and files in a fresh project root.
func TestScaffoldEvolutionDir_CreatesDirectoryTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := scaffoldEvolutionDir(root); err != nil {
		t.Fatalf("scaffoldEvolutionDir: %v", err)
	}

	// Expected directories.
	expectedDirs := []string{
		filepath.Join(root, ".moai", "evolution"),
		filepath.Join(root, ".moai", "evolution", "telemetry"),
		filepath.Join(root, ".moai", "evolution", "learnings"),
		filepath.Join(root, ".moai", "evolution", "new-skills"),
	}
	for _, dir := range expectedDirs {
		fi, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", dir, err)
			continue
		}
		if !fi.IsDir() {
			t.Errorf("%s should be a directory", dir)
		}
	}

	// Expected files.
	expectedFiles := []string{
		filepath.Join(root, ".moai", "evolution", "manifest.yaml"),
		filepath.Join(root, ".moai", "evolution", "changelog.md"),
		filepath.Join(root, ".moai", "evolution", "telemetry", ".gitkeep"),
		filepath.Join(root, ".moai", "evolution", "learnings", ".gitkeep"),
		filepath.Join(root, ".moai", "evolution", "new-skills", ".gitkeep"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}
}

// TestScaffoldEvolutionDir_ManifestContent verifies the default manifest.yaml schema.
func TestScaffoldEvolutionDir_ManifestContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := scaffoldEvolutionDir(root); err != nil {
		t.Fatalf("scaffoldEvolutionDir: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".moai", "evolution", "manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	content := string(data)
	expectedKeys := []string{
		"schema_version: 1",
		"evolved_skills: []",
		"new_skills: []",
		"learnings_count: 0",
		"last_evolution_date:",
		"rate_limit:",
	}
	for _, key := range expectedKeys {
		if !strings.Contains(content, key) {
			t.Errorf("manifest.yaml missing expected key %q\ncontent:\n%s", key, content)
		}
	}
}

// TestScaffoldEvolutionDir_Idempotent verifies that running scaffoldEvolutionDir
// twice on the same root does not overwrite existing content.
func TestScaffoldEvolutionDir_Idempotent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// First call.
	if err := scaffoldEvolutionDir(root); err != nil {
		t.Fatalf("first scaffoldEvolutionDir: %v", err)
	}

	// Write custom content to manifest.yaml to detect overwrites.
	manifestPath := filepath.Join(root, ".moai", "evolution", "manifest.yaml")
	customContent := "schema_version: 99\ncustom: data\n"
	if err := os.WriteFile(manifestPath, []byte(customContent), 0o644); err != nil {
		t.Fatalf("write custom manifest: %v", err)
	}

	// Second call — should not overwrite existing manifest.
	if err := scaffoldEvolutionDir(root); err != nil {
		t.Fatalf("second scaffoldEvolutionDir: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest after second call: %v", err)
	}
	if string(data) != customContent {
		t.Errorf("manifest overwritten by second scaffoldEvolutionDir call\ngot:\n%s", string(data))
	}
}

// TestIsMoaiManaged_EvolutionPaths verifies that .moai/evolution/ paths are protected.
func TestIsMoaiManaged_EvolutionPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{".moai/evolution/manifest.yaml", true},
		{".moai/evolution/learnings/LEARN-001.md", true},
		{".moai/evolution/new-skills/moai-evolved-auth/SKILL.md", true},
		{".moai/evolution/telemetry/usage.jsonl", true},
		{".moai/config/sections/quality.yaml", true},
		{".moai/other/file.yaml", false},
		{".claude/skills/moai-foo/SKILL.md", true},
		{"CLAUDE.md", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got := isMoaiManaged(tt.path)
			if got != tt.want {
				t.Errorf("isMoaiManaged(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestScaffoldEvolutionDir_PreservesLearning verifies that a pre-existing
// LEARN-001.md file under .moai/evolution/learnings/ is not modified.
func TestScaffoldEvolutionDir_PreservesLearning(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Pre-create learnings directory and a learning file.
	learningsDir := filepath.Join(root, ".moai", "evolution", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatalf("mkdir learnings: %v", err)
	}
	learningContent := "# LEARN-001\n\nDo not rationalize failures.\n"
	learningPath := filepath.Join(learningsDir, "LEARN-001.md")
	if err := os.WriteFile(learningPath, []byte(learningContent), 0o644); err != nil {
		t.Fatalf("write learning: %v", err)
	}

	// Run scaffolding — should not touch the learning file.
	if err := scaffoldEvolutionDir(root); err != nil {
		t.Fatalf("scaffoldEvolutionDir: %v", err)
	}

	data, err := os.ReadFile(learningPath)
	if err != nil {
		t.Fatalf("read learning after scaffold: %v", err)
	}
	if string(data) != learningContent {
		t.Errorf("learning file modified by scaffoldEvolutionDir\ngot:\n%s", string(data))
	}
}
