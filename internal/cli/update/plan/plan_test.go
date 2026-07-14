package plan

import (
	"os"
	"path/filepath"
	"testing"

	mrg "github.com/modu-ai/moai-adk/internal/merge"
)

// TestIsUserOwnedNamespace verifies the authoritative user-owned namespace predicate
// (REQ-UNP-001/002/003/009, REQ-HPR-005, REQ-HNS-001/005, SPEC-V3R6-HARNESS-V4-001 M1).
func TestIsUserOwnedNamespace(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Canonical hns-* skills (REQ-HPR-005)
		{".claude/skills/hns-my-tool/SKILL.md", true},
		{".claude/skills/hns-test/skills/module.md", true},
		{".claude/skills/hns-", true}, // Prefix match is still user-owned namespace

		// Legacy harness-* skills (REQ-HNS-001, §24.1)
		{".claude/skills/harness-my-tool/SKILL.md", true},
		{".claude/skills/harness-", true}, // Prefix match

		// Legacy my-harness-* (REQ-UNP-001, REQ-HNS-005)
		{".claude/skills/my-harness-old/SKILL.md", true},

		// Harness agents (REQ-UNP-002)
		{".claude/agents/harness", true},
		{".claude/agents/harness/my-specialist.md", true},
		{".claude/agents/harness/subdir/file.md", true},

		// .moai/harness extension (REQ-UNP-003)
		{".moai/harness", true},
		{".moai/harness/manifest.json", true},

		// Harness commands (SPEC-V3R6-HARNESS-V4-001 M1 AC-HV4-010a)
		{".claude/commands/harness", true},
		{".claude/commands/harness/my-command.js", true},

		// Canonical hns-* Runner Workflows (SPEC-HNS-PREFIX-RENAME-001 M2)
		{".claude/workflows/hns-my-runner.js", true},
		{".claude/workflows/hns-", true}, // Prefix match

		// Legacy harness-* Runner Workflows (SPEC-V3R6-HARNESS-V4-001 M1 AC-HV4-010b)
		{".claude/workflows/harness-legacy.js", true},

		// User direct-added skills (REQ-UNP-009)
		{".claude/skills/custom/SKILL.md", true},
		{".claude/skills/my-custom/skills/module.md", true},
		{".claude/skills/user-tool/README.md", true},

		// MoAI-managed skills (should be false)
		{".claude/skills/moai/anything", false},
		{".claude/skills/moai-core/SKILL.md", false},
		{".claude/skills/moai-workflow/skills/module.md", false},

		// User agents (REQ-UNP-009)
		{".claude/agents/custom-agent.md", true},
		{".claude/agents/my-team/custom.md", true},

		// System agent domain subfolders (should be false)
		{".claude/agents/core/anything", false},
		{".claude/agents/expert/test", false},
		{".claude/agents/meta/file.md", false},

		// System agent prefixes (should be false)
		{".claude/agents/moai-system.md", false},
		{".claude/agents/moai-custom.md", false},
		// Note: manager-*, expert-*, builder-*, evaluator-* are NOT managed
		// They are user-owned because they're not "moai-*" or "moai" prefix
		// IsMoaiManaged only checks for moai-*/moai in agents directory
		{".claude/agents/manager-test.md", false}, // Correct: user-owned, not MoAI-managed
		{".claude/agents/expert-backend.md", false}, // Correct: user-owned

		// Empty and edge cases
		{"", false},
		{".claude/skills/", false},
		{".claude/agents/", false},

		// Windows path normalization (NFR-UNP-003)
		{".claude\\skills\\hns-test\\SKILL.md", true},
		{".claude\\agents\\harness\\custom.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsUserOwnedNamespace(tt.path)
			if got != tt.want {
				t.Errorf("IsUserOwnedNamespace(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsUserOwnedNamespace_SupersetGuarantee verifies NFR-UNP-005: IsUserOwnedNamespace
// is a STRICT SUPERSET of IsUserAreaPath for user-owned surfaces.
func TestIsUserOwnedNamespace_SupersetGuarantee(t *testing.T) {
	userOwnedPaths := []string{
		".claude/skills/hns-my-tool/SKILL.md",
		".claude/skills/harness-legacy/SKILL.md",
		".claude/agents/harness/my-specialist.md",
		".claude/commands/harness/my-command.js",
		".claude/workflows/hns-runner.js",
		".claude/workflows/harness-legacy.js",
	}

	for _, rel := range userOwnedPaths {
		areaPath := IsUserAreaPath(rel)
		ownedNs := IsUserOwnedNamespace(rel)

		if areaPath && !ownedNs {
			t.Errorf("NFR-UNP-005 violation: IsUserAreaPath(%q)=true but IsUserOwnedNamespace=false (must be superset)", rel)
		}
	}
}

// TestIsUserAreaPath verifies user area path detection for directory-based paths
// (not colon-separated command names, which are not recognized).
func TestIsUserAreaPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Canonical hns-* skills
		{".claude/skills/hns-my-tool/SKILL.md", true},
		{".claude/skills/hns-test/skills/module.md", true},

		// Legacy harness-* skills
		{".claude/skills/harness-my-tool/SKILL.md", true},

		// Legacy my-harness-*
		{".claude/skills/my-harness-old/SKILL.md", true},

		// Harness agents
		{".claude/agents/harness", true},
		{".claude/agents/harness/my-specialist.md", true},

		// Legacy my-harness agents
		{".claude/agents/my-harness", true},
		{".claude/agents/my-harness/custom.md", true},

		// Harness commands
		{".claude/commands/harness", true},
		{".claude/commands/harness/my-command.js", true},

		// Canonical hns-* workflows
		{".claude/workflows/hns-runner.js", true},

		// Legacy harness-* workflows
		{".claude/workflows/harness-legacy.js", true},

		// MoAI-managed paths (should be false)
		{".claude/skills/moai-core/SKILL.md", false},
		{".claude/agents/core/anything", false},
		{".claude/rules/moai-core/test.md", false},
		{".claude/commands/moai-test.js", false},

		// Windows path normalization
		{".claude\\skills\\hns-test\\SKILL.md", true},
		{".claude\\agents\\harness\\custom.md", true},

		// Empty and edge cases
		{"", false},
		{".claude/skills/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsUserAreaPath(tt.path)
			if got != tt.want {
				t.Errorf("IsUserAreaPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsMoaiManaged verifies MoAI-managed file detection.
func TestIsMoaiManaged(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// .moai/config/ paths
		{".moai/config/config.yaml", true},
		{".moai/config/sections/quality.yaml", true},

		// .moai/evolution/ protection
		{".moai/evolution/learnings.json", true},
		{".moai/evolution/new-skills/", true},

		// MoAI-managed skills
		{".claude/skills/moai-core/SKILL.md", true},
		{".claude/skills/moai-workflow/skills/module.md", true},
		{".claude/skills/moai-foundation/SKILL.md", true},

		// MoAI-managed rules
		{".claude/rules/moai-core/test.md", true},
		{".claude/rules/moai-workflow/spec.md", true},

		// MoAI-managed commands
		{".claude/commands/moai-test.js", true},
		{".claude/commands/moai-run.js", true},

		// MoAI-managed output-styles
		{".claude/output-styles/moai-default.md", true},

		// MoAI-managed hooks
		{".claude/hooks/moai-handler.sh", true},

		// System agent domain subfolders (post SPEC-V3R6-AGENT-FOLDER-SPLIT-001)
		{".claude/agents/core/manager-develop.md", true},
		{".claude/agents/expert/backend.md", true},
		{".claude/agents/meta/super-advisor.md", true},

		// System agent prefixes
		{".claude/agents/moai-system.md", true},
		// manager-* and expert-* are user-owned in .claude/agents/ root
		// Only moai-* and moai prefixes are MoAI-managed
		{".claude/agents/manager-test.md", false}, // User-owned, not managed
		{".claude/agents/expert-backend.md", false}, // User-owned, not managed

		// User-owned surfaces (should be false)
		{".claude/skills/hns-my-tool/SKILL.md", false},
		{".claude/agents/harness/custom.md", false},
		{".claude/agents/my-agent.md", false},

		// Non-.claude paths
		{".moai/specs/SPEC-001/spec.md", false},
		{"README.md", false},
		{"internal/file.go", false},

		// Windows path handling
		{".moai\\config\\config.yaml", true},
		{".claude\\skills\\moai-core\\SKILL.md", true},

		// Empty and edge cases
		{"", false},
		{".claude", false},
		{".claude/skills", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsMoaiManaged(tt.path)
			if got != tt.want {
				t.Errorf("IsMoaiManaged(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestClassifyFileRisk verifies file risk classification based on filename and existence.
func TestClassifyFileRisk(t *testing.T) {
	tests := []struct {
		filename string
		exists   bool
		want     string
	}{
		// High-risk files (critical config, golden values from M3b characterization)
		{"CLAUDE.md", true, "high"},
		{"CLAUDE.md", false, "high"},
		{"settings.json", true, "high"},
		{"settings.json", false, "high"},

		// New files are low risk
		{"new-file.md", false, "low"},
		{"test.go", false, "low"},
		{"config.yaml", false, "low"},

		// Existing files are medium risk
		{"existing-file.md", true, "medium"},
		{"internal/test.go", true, "medium"},
		{".moai/config/quality.yaml", true, "medium"},

		// Edge cases
		{"", false, "low"},
		{".gitignore", true, "medium"},
		{".gitignore", false, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := ClassifyFileRisk(tt.filename, tt.exists)
			if got != tt.want {
				t.Errorf("ClassifyFileRisk(%q, exists=%v) = %q, want %q", tt.filename, tt.exists, got, tt.want)
			}
		})
	}
}

// TestDetermineStrategy verifies merge strategy selection based on file extension and name.
func TestDetermineStrategy(t *testing.T) {
	tests := []struct {
		filename string
		want     mrg.MergeStrategy
	}{
		// CLAUDE.md section merge
		{"CLAUDE.md", mrg.SectionMerge},
		{"path/to/CLAUDE.md", mrg.SectionMerge},

		// .gitignore entry merge
		{".gitignore", mrg.EntryMerge},
		{"path/.gitignore", mrg.EntryMerge},

		// JSON files
		{"settings.json", mrg.JSONMerge},
		{"package.json", mrg.JSONMerge},
		{"manifest.json", mrg.JSONMerge},

		// YAML files
		{"config.yaml", mrg.YAMLDeep},
		{"config.yml", mrg.YAMLDeep},
		{".moai/config/quality.yaml", mrg.YAMLDeep},

		// Default: line merge for everything else
		{"README.md", mrg.LineMerge},
		{"test.go", mrg.LineMerge},
		{"script.sh", mrg.LineMerge},
		{"file.txt", mrg.LineMerge},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetermineStrategy(tt.filename)
			if got != tt.want {
				t.Errorf("DetermineStrategy(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// TestDetermineChangeType verifies change type detection based on file existence.
func TestDetermineChangeType(t *testing.T) {
	tests := []struct {
		exists bool
		want   string
	}{
		{true, "update existing"},
		{false, "new file"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := DetermineChangeType(tt.exists)
			if got != tt.want {
				t.Errorf("DetermineChangeType(%v) = %q, want %q", tt.exists, got, tt.want)
			}
		})
	}
}

// TestAnalyzeFiles verifies the file analysis pipeline.
func TestAnalyzeFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		templates    []string
		createFiles  map[string]string // files to create in tempDir
		wantCount    int              // expected number of analyzed files
		excludedPath string           // path that should be excluded (MoAI-managed)
	}{
		{
			name:      "new files only",
			templates: []string{"new-file.md", "config.yaml"},
			wantCount: 2,
		},
		{
			name:      "mix of new and existing files",
			templates: []string{"existing.md", "new.md"},
			createFiles: map[string]string{
				"existing.md": "# Existing Content",
			},
			wantCount: 2,
		},
		{
			name:        "excludes MoAI-managed files",
			templates:   []string{".claude/skills/moai-core/SKILL.md", "user-file.md"},
			wantCount:   1, // Only user-file.md should be analyzed
			excludedPath: ".claude/skills/moai-core/SKILL.md",
		},
		{
			name:      "handles .tmpl suffix removal",
			templates: []string{"settings.json.tmpl", "config.yaml.tmpl"},
			wantCount: 2,
		},
		{
			name:      "high-risk files classified correctly",
			templates: []string{"CLAUDE.md", "settings.json"},
			wantCount: 2,
		},
		{
			name:      "strategy selection based on extension",
			templates: []string{".gitignore", "data.json", "config.yaml", "readme.md"},
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test files
			for filename, content := range tt.createFiles {
				path := filepath.Join(tempDir, filename)
				err := os.MkdirAll(filepath.Dir(path), 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				err = os.WriteFile(path, []byte(content), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// Run analysis
			got := AnalyzeFiles(tt.templates, tempDir)

			// Verify count
			if len(got) != tt.wantCount {
				t.Errorf("AnalyzeFiles() count = %d, want %d", len(got), tt.wantCount)
			}

			// Verify excluded path is not in results
			if tt.excludedPath != "" {
				for _, analysis := range got {
					if analysis.Path == tt.excludedPath {
						t.Errorf("MoAI-managed file %q should be excluded from analysis", tt.excludedPath)
					}
				}
			}

			// Verify basic properties for non-empty results
			for _, analysis := range got {
				if analysis.Path == "" {
					t.Error("analysis Path should not be empty")
				}
				if analysis.Changes != "new file" && analysis.Changes != "update existing" {
					t.Errorf("analysis Changes = %q, want 'new file' or 'update existing'", analysis.Changes)
				}
				if analysis.RiskLevel != "high" && analysis.RiskLevel != "medium" && analysis.RiskLevel != "low" {
					t.Errorf("analysis RiskLevel = %q, want high/medium/low", analysis.RiskLevel)
				}
			}
		})
	}
}

// TestGetProjectConfigVersion verifies config version retrieval.
func TestGetProjectConfigVersion(t *testing.T) {
	// Create .moai/config/sections directory (the correct path for system.yaml)
	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	err := os.MkdirAll(sectionsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create sections directory: %v", err)
	}

	// Test with missing config.yaml
	// Actual behavior: returns "0.0.0" when config is missing (to force update)
	version, err := GetProjectConfigVersion(tempDir)
	if err != nil {
		t.Errorf("GetProjectConfigVersion() unexpected error when missing: %v", err)
	}
	if version != "0.0.0" {
		t.Errorf("GetProjectConfigVersion() when missing = %q, want %q", version, "0.0.0")
	}

	// Create system.yaml with template_version
	configPath := filepath.Join(sectionsDir, "system.yaml")
	configContent := `moai:
  template_version: "2.5.0"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Test successful version retrieval with template_version set
	version, err = GetProjectConfigVersion(tempDir)
	if err != nil {
		t.Errorf("GetProjectConfigVersion() unexpected error: %v", err)
	}
	if version != "2.5.0" {
		t.Errorf("GetProjectConfigVersion() = %q, want %q", version, "2.5.0")
	}

	// Test with template_version set
	configContent = `moai:
  template_version: "3.0.0"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to update system.yaml: %v", err)
	}

	// Version should now be extracted correctly
	version, err = GetProjectConfigVersion(tempDir)
	if err != nil {
		t.Errorf("GetProjectConfigVersion() with template_version error: %v", err)
	}
	// Check if we can successfully parse the template_version
	if version == "" {
		t.Error("GetProjectConfigVersion() returned empty version when template_version is set")
	}
	// The exact value depends on the YAML parsing - just verify it's not empty or 0.0.0
	if version == "0.0.0" {
		t.Errorf("GetProjectConfigVersion() = %q, want non-zero when template_version is set", version)
	}
}

// TestIsUserOwnedNamespace_EdgeCases verifies edge cases and boundary conditions.
func TestIsUserOwnedNamespace_EdgeCases(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Subdirectory traversal
		{".claude/skills/hns-tool/subdir/file.md", true},
		{".claude/agents/harness/nested/deep/file.md", true},

		// Mixed separators (Windows/Unix)
		{".claude\\skills\\hns-tool\\SKILL.md", true},
		{".claude/agents\\harness\\file.md", true},

		// Prefix-only paths (prefix match is user-owned)
		{".claude/skills/hns-", true},
		{".claude/agents/harness", true}, // Exact directory match

		// Empty components
		{".claude/skills//file.md", false},

		// Case sensitivity in hns- prefix
		// ".claude/skills/HNS-tool" matches the ".claude/skills/" prefix pattern
		// and "HNS-tool" is not "moai" or "moai-*", so it's user-owned
		{".claude/skills/HNS-tool/SKILL.md", true},
		{".claude/skills/hns-HNS/SKILL.md", true},      // Mixed case in name is fine
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsUserOwnedNamespace(tt.path)
			if got != tt.want {
				t.Errorf("IsUserOwnedNamespace(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsUserAreaPath_EdgeCases verifies edge cases for user area path detection.
func TestIsUserAreaPath_EdgeCases(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Subdirectory traversal
		{".claude/skills/hns-tool/deep/nest/file.md", true},

		// Windows paths
		{".claude\\skills\\hns-tool\\SKILL.md", true},
		{".claude\\agents\\harness\\custom.md", true},

		// Exact directory matches
		{".claude/agents/harness", true},
		{".claude/commands/harness", true},

		// Non-matching prefixes
		{".claude/skills/moai-tool/SKILL.md", false},
		{".claude/agents/core/test.md", false},

		// Empty and malformed paths
		{"", false},
		{".claude/skills/", false},
		{"skills/hns-tool/SKILL.md", false}, // Missing .claude prefix
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsUserAreaPath(tt.path)
			if got != tt.want {
				t.Errorf("IsUserAreaPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestClassifyFileRisk_EdgeCases verifies edge cases for risk classification.
func TestClassifyFileRisk_EdgeCases(t *testing.T) {
	tests := []struct {
		filename string
		exists   bool
		want     string
	}{
		// Case sensitivity in basename (filepath.Base preserves case)
		{"claude.md", false, "low"},     // Not CLAUDE.md
		{"SETTINGS.JSON", true, "medium"},  // Case-sensitive: not exact match

		// Path with directory
		{".claude/settings.json", true, "high"},
		{".moai/CLAUDE.md", true, "high"},

		// Files without extension
		{"README", true, "medium"},
		{"Makefile", false, "low"},

		// Empty filename
		{"", false, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := ClassifyFileRisk(tt.filename, tt.exists)
			if got != tt.want {
				t.Errorf("ClassifyFileRisk(%q, exists=%v) = %q, want %q", tt.filename, tt.exists, got, tt.want)
			}
		})
	}
}
