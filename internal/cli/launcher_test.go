package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"empty defaults to claude", "", "claude"},
		{"claude", "claude", "claude"},
		{"glm", "glm", "glm"},
		{"claude_glm", "claude_glm", "claude_glm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveMode(tt.mode)
			if got != tt.want {
				t.Errorf("resolveMode(%q) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestParseProfileFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantProfile string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name:        "no flags",
			args:        []string{},
			wantProfile: "",
			wantArgs:    []string{},
		},
		{
			name:        "-p with value",
			args:        []string{"-p", "work"},
			wantProfile: "work",
			wantArgs:    []string{},
		},
		{
			name:        "--profile with value",
			args:        []string{"--profile", "work"},
			wantProfile: "work",
			wantArgs:    []string{},
		},
		{
			name:        "--profile=value form",
			args:        []string{"--profile=work"},
			wantProfile: "work",
			wantArgs:    []string{},
		},
		{
			name:        "-p=value form",
			args:        []string{"-p=work"},
			wantProfile: "work",
			wantArgs:    []string{},
		},
		{
			name:        "-p with extra args",
			args:        []string{"-p", "work", "--bypass"},
			wantProfile: "work",
			wantArgs:    []string{"--bypass"},
		},
		{
			name:        "pass-through after --",
			args:        []string{"--", "-p", "work"},
			wantProfile: "",
			wantArgs:    []string{"--", "-p", "work"},
		},
		{
			name:    "-p without value at end",
			args:    []string{"-p"},
			wantErr: true,
		},
		{
			name:    "--profile without value at end",
			args:    []string{"--profile"},
			wantErr: true,
		},
		{
			name:    "-p followed by another flag",
			args:    []string{"-p", "--bypass"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, args, err := parseProfileFlag(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if profile != tt.wantProfile {
				t.Errorf("profile = %q, want %q", profile, tt.wantProfile)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, a, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestUnifiedLaunch_Claude(t *testing.T) {
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()

	var launchedProfile string
	var launchedArgs []string
	launchClaudeFunc = func(p string, args []string) error {
		launchedProfile = p
		launchedArgs = args
		return nil
	}

	err := unifiedLaunch("myprofile", "claude", []string{"--bypass"})
	if err != nil {
		t.Fatalf("unifiedLaunch error: %v", err)
	}

	if launchedProfile != "myprofile" {
		t.Errorf("profile = %q, want %q", launchedProfile, "myprofile")
	}
	if len(launchedArgs) != 1 || launchedArgs[0] != "--bypass" {
		t.Errorf("args = %v, want [--bypass]", launchedArgs)
	}
}

func TestUnifiedLaunch_GLM(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// GLM mode requires an API key
	t.Setenv("GLM_API_KEY", "test-key")

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(p string, args []string) error { return nil }

	err := unifiedLaunch("", "glm", nil)
	if err != nil {
		t.Fatalf("unifiedLaunch(glm) error: %v", err)
	}
}

func TestUnifiedLaunch_CG_NoTmux(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GLM_API_KEY", "test-key")
	t.Setenv("TMUX", "")
	t.Setenv("MOAI_TEST_MODE", "")

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(p string, args []string) error { return nil }

	err := unifiedLaunch("", "claude_glm", nil)
	if err == nil {
		t.Fatal("CG mode without tmux should error")
	}
	if !strings.Contains(err.Error(), "tmux session") {
		t.Errorf("error should mention tmux, got: %v", err)
	}
}

func TestUnifiedLaunch_CG_WithTestMode(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GLM_API_KEY", "test-key")
	t.Setenv("MOAI_TEST_MODE", "1")

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(p string, args []string) error { return nil }

	err := unifiedLaunch("", "claude_glm", nil)
	if err != nil {
		t.Fatalf("CG mode with MOAI_TEST_MODE=1 should not error, got: %v", err)
	}
}

func TestUnifiedLaunch_NotInProject(t *testing.T) {
	tmpDir := t.TempDir()
	// No .moai directory

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(p string, args []string) error { return nil }

	err := unifiedLaunch("", "claude", nil)
	if err == nil {
		t.Fatal("unifiedLaunch should error when not in a MoAI project")
	}
	if !strings.Contains(err.Error(), "not in a MoAI project") {
		t.Errorf("error should mention 'not in a MoAI project', got: %v", err)
	}
}

