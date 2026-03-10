package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetActiveMode_Parse tests parsing llm.yaml for team_mode.
func TestGetActiveMode_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantMode string
	}{
		{
			name: "empty team_mode returns cc",
			content: `
llm:
    team_mode: ""
`,
			wantMode: "cc",
		},
		{
			name: "glm mode",
			content: `
llm:
    team_mode: "glm"
`,
			wantMode: "glm",
		},
		{
			name: "cg mode",
			content: `
llm:
    team_mode: "cg"
`,
			wantMode: "cg",
		},
		{
			name: "no quotes",
			content: `
llm:
    team_mode: glm
`,
			wantMode: "glm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			// Create the full path structure: .moai/config/sections/llm.yaml
			configDir := filepath.Join(tempDir, ".moai", "config", "sections")
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				t.Fatal(err)
			}
			llmPath := filepath.Join(configDir, "llm.yaml")
			if err := os.WriteFile(llmPath, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			got, err := GetActiveMode(tempDir)
			if err != nil {
				t.Fatalf("GetActiveMode() error = %v", err)
			}
			if got != tt.wantMode {
				t.Errorf("GetActiveMode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// TestGetActiveMode_Default tests that missing file returns cc.
func TestGetActiveMode_Default(t *testing.T) {
	tempDir := t.TempDir()
	// No llm.yaml file

	got, err := GetActiveMode(tempDir)
	if err != nil {
		t.Fatalf("GetActiveMode() error = %v", err)
	}
	if got != "cc" {
		t.Errorf("GetActiveMode() = %v, want cc", got)
	}
}

// TestBuildTmuxSessionConfig_GLMMode tests GLM env var loading.
func TestBuildTmuxSessionConfig_GLMMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create llm.yaml with glm mode
	llmContent := `
llm:
    team_mode: "glm"
`
	configDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	llmPath := filepath.Join(configDir, "llm.yaml")
	if err := os.WriteFile(llmPath, []byte(llmContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := BuildTmuxSessionConfig("test-project", "SPEC-TEST-001", "/worktree", tempDir)
	if err != nil {
		t.Fatalf("BuildTmuxSessionConfig() error = %v", err)
	}

	if cfg.ActiveMode != "glm" {
		t.Errorf("ActiveMode = %v, want glm", cfg.ActiveMode)
	}

	// Check that GLMEnvVars map is created (may be empty if no .env.glm exists)
	if cfg.GLMEnvVars == nil {
		t.Error("GLMEnvVars map not created")
	}
}

// TestBuildTmuxSessionConfig_CCMode tests that CC mode has no GLM env vars.
func TestBuildTmuxSessionConfig_CCMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create llm.yaml with empty team_mode (cc)
	llmContent := `
llm:
    team_mode: ""
`
	configDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	llmPath := filepath.Join(configDir, "llm.yaml")
	if err := os.WriteFile(llmPath, []byte(llmContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := BuildTmuxSessionConfig("test-project", "SPEC-TEST-001", "/worktree", tempDir)
	if err != nil {
		t.Fatalf("BuildTmuxSessionConfig() error = %v", err)
	}

	if cfg.ActiveMode != "cc" {
		t.Errorf("ActiveMode = %v, want cc", cfg.ActiveMode)
	}

	if len(cfg.GLMEnvVars) != 0 {
		t.Errorf("GLMEnvVars should be empty in CC mode, got %d vars", len(cfg.GLMEnvVars))
	}
}
