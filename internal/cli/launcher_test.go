package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCleanupMoaiWorktrees_GlobalPath verifies that cleanupMoaiWorktrees
// removes worker worktrees from both the local .claude/worktrees/ path and
// the global ~/.moai/worktrees/*/ path.
//
// NOTE: does not call t.Parallel() because it sets HOME via t.Setenv.
func TestCleanupMoaiWorktrees_GlobalPath(t *testing.T) {
	tests := []struct {
		name         string
		createLocal  bool // create a worktree under .claude/worktrees/
		createGlobal bool // create a worktree under ~/.moai/worktrees/myproject/
		wantSubstrs  []string
		wantEmpty    bool
	}{
		{
			name:         "global path only - no local worktrees dir",
			createLocal:  false,
			createGlobal: true,
			wantSubstrs:  []string{"worker-global-001"},
		},
		{
			name:         "both local and global paths",
			createLocal:  true,
			createGlobal: true,
			wantSubstrs:  []string{"worker-local-001", "worker-global-001"},
		},
		{
			name:         "neither path exists - returns empty",
			createLocal:  false,
			createGlobal: false,
			wantEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpHome := t.TempDir()
			t.Setenv("HOME", tmpHome)
			t.Setenv("USERPROFILE", tmpHome) // Windows: os.UserHomeDir() uses USERPROFILE, not HOME

			// Initialize a git repo with an initial commit.
			gitCmds := [][]string{
				{"init"},
				{"config", "user.email", "test@test.com"},
				{"config", "user.name", "Test"},
			}
			for _, args := range gitCmds {
				if _, err := runGitCommand(tmpDir, args...); err != nil {
					t.Skipf("git %v failed: %v", args, err)
				}
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{{"add", "."}, {"commit", "-m", "initial"}} {
				if _, err := runGitCommand(tmpDir, args...); err != nil {
					t.Skipf("git %v failed: %v", args, err)
				}
			}

			if tt.createLocal {
				localWorkerPath := filepath.Join(tmpDir, ".claude", "worktrees", "worker-local-001")
				if err := os.MkdirAll(filepath.Dir(localWorkerPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if _, err := runGitCommand(tmpDir, "worktree", "add", localWorkerPath, "-b", "worker-local-001"); err != nil {
					t.Skipf("git worktree add (local) failed: %v", err)
				}
			}

			if tt.createGlobal {
				globalWorktreeDir := filepath.Join(tmpHome, ".moai", "worktrees", "myproject")
				globalWorkerPath := filepath.Join(globalWorktreeDir, "worker-global-001")
				if err := os.MkdirAll(globalWorktreeDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if _, err := runGitCommand(tmpDir, "worktree", "add", globalWorkerPath, "-b", "worker-global-001"); err != nil {
					t.Skipf("git worktree add (global) failed: %v", err)
				}
			}

			result := cleanupMoaiWorktrees(tmpDir)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("expected empty result, got: %q", result)
				}
				return
			}

			for _, substr := range tt.wantSubstrs {
				if !strings.Contains(result, substr) {
					t.Errorf("result %q does not contain expected %q", result, substr)
				}
			}
		})
	}
}

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
		{
			name:    "--profile= empty value",
			args:    []string{"--profile="},
			wantErr: true,
		},
		{
			name:    "-p= empty value",
			args:    []string{"-p="},
			wantErr: true,
		},
		{
			name:    "-p with empty string value",
			args:    []string{"-p", ""},
			wantErr: true,
		},
		{
			name:    "--profile with empty string value",
			args:    []string{"--profile", ""},
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

func TestSyncPermissionModeToSettingsLocal(t *testing.T) {
	tests := []struct {
		name         string
		existing     string // existing settings.local.json content ("" = no file)
		mode         string // permission mode to sync
		wantMode     string // expected defaultMode ("" = should not exist)
		wantEnvKey   string // verify existing env is preserved
		wantEnvValue string
	}{
		{
			name:     "bypassPermissions creates permissions section",
			existing: "",
			mode:     "bypassPermissions",
			wantMode: "bypassPermissions",
		},
		{
			name:     "auto mode sets defaultMode",
			existing: "",
			mode:     "auto",
			wantMode: "auto",
		},
		{
			name:     "plan mode sets defaultMode",
			existing: "",
			mode:     "plan",
			wantMode: "plan",
		},
		{
			name:     "default mode sets defaultMode",
			existing: "",
			mode:     "default",
			wantMode: "default",
		},
		{
			name:     "dontAsk mode sets defaultMode",
			existing: "",
			mode:     "dontAsk",
			wantMode: "dontAsk",
		},
		{
			name:     "mode preserves existing env",
			existing: `{"env":{"SOME_EXISTING_VAR":"keep_me"}}`,
			mode:     "auto",
			wantMode:     "auto",
			wantEnvKey:   "SOME_EXISTING_VAR",
			wantEnvValue: "keep_me",
		},
		{
			name:     "empty mode removes defaultMode override",
			existing: `{"permissions":{"defaultMode":"bypassPermissions"},"env":{"FOO":"bar"}}`,
			mode:     "",
			wantMode:     "",
			wantEnvKey:   "FOO",
			wantEnvValue: "bar",
		},
		{
			name:     "acceptEdits removes defaultMode (matches project default)",
			existing: `{"permissions":{"defaultMode":"auto"}}`,
			mode:     "acceptEdits",
			wantMode: "",
		},
		{
			name:     "empty mode with no permissions is no-op",
			existing: `{"env":{"FOO":"bar"}}`,
			mode:     "",
			wantMode:     "",
			wantEnvKey:   "FOO",
			wantEnvValue: "bar",
		},
		{
			name:     "mode overwrites existing defaultMode",
			existing: `{"permissions":{"defaultMode":"acceptEdits","allow":["Read"]}}`,
			mode:     "auto",
			wantMode: "auto",
		},
		{
			name:     "empty mode preserves other permission keys",
			existing: `{"permissions":{"defaultMode":"bypassPermissions","allow":["Read"]}}`,
			mode:     "",
			wantMode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			claudeDir := filepath.Join(tmpDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0o755); err != nil {
				t.Fatal(err)
			}
			settingsPath := filepath.Join(claudeDir, "settings.local.json")

			if tt.existing != "" {
				if err := os.WriteFile(settingsPath, []byte(tt.existing), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			err := syncPermissionModeToSettingsLocal(settingsPath, tt.mode)
			if err != nil {
				t.Fatalf("syncPermissionModeToSettingsLocal() error: %v", err)
			}

			data, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read result: %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("parse result: %v", err)
			}

			perms, _ := result["permissions"].(map[string]any)
			if tt.wantMode != "" {
				if perms == nil {
					t.Fatal("expected permissions section, got nil")
				}
				got, _ := perms["defaultMode"].(string)
				if got != tt.wantMode {
					t.Errorf("defaultMode = %q, want %q", got, tt.wantMode)
				}
			} else {
				if perms != nil {
					if mode, ok := perms["defaultMode"]; ok {
						t.Errorf("defaultMode should not exist, got %v", mode)
					}
				}
			}

			if tt.wantEnvKey != "" {
				env, _ := result["env"].(map[string]any)
				if env == nil {
					t.Fatalf("expected env section, got nil")
				}
				got, _ := env[tt.wantEnvKey].(string)
				if got != tt.wantEnvValue {
					t.Errorf("env[%s] = %q, want %q", tt.wantEnvKey, got, tt.wantEnvValue)
				}
			}

			// Check other permission keys preserved
			if tt.name == "empty mode preserves other permission keys" {
				if perms == nil {
					t.Fatal("expected permissions section with allow key")
				}
				if _, ok := perms["allow"]; !ok {
					t.Error("permissions.allow should be preserved")
				}
			}
		})
	}
}

// TestSyncBypassToSettingsLocal_BackwardCompat verifies the legacy wrapper still works.
func TestSyncBypassToSettingsLocal_BackwardCompat(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// bypass=true should set bypassPermissions
	if err := syncBypassToSettingsLocal(settingsPath, true); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(settingsPath)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	perms, _ := result["permissions"].(map[string]any)
	if got, _ := perms["defaultMode"].(string); got != "bypassPermissions" {
		t.Errorf("bypass=true: defaultMode = %q, want %q", got, "bypassPermissions")
	}

	// bypass=false should remove it
	if err := syncBypassToSettingsLocal(settingsPath, false); err != nil {
		t.Fatal(err)
	}

	data, _ = os.ReadFile(settingsPath)
	result = nil
	_ = json.Unmarshal(data, &result)
	perms, _ = result["permissions"].(map[string]any)
	if perms != nil {
		if _, ok := perms["defaultMode"]; ok {
			t.Error("bypass=false: defaultMode should not exist")
		}
	}
}

func TestContainsPermissionMode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		mode string
		want bool
	}{
		{"flag with value", []string{"--permission-mode", "auto"}, "auto", true},
		{"flag=value form", []string{"--permission-mode=auto"}, "auto", true},
		{"different mode", []string{"--permission-mode", "plan"}, "auto", false},
		{"no flag", []string{"-b", "--bypass"}, "auto", false},
		{"flag without value at end", []string{"--permission-mode"}, "auto", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsPermissionMode(tt.args, tt.mode)
			if got != tt.want {
				t.Errorf("containsPermissionMode(%v, %q) = %v, want %v", tt.args, tt.mode, got, tt.want)
			}
		})
	}
}

func TestExpandModelString(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{"empty string", "", ""},
		{"standard opus", "claude-opus-4-6", "claude-opus-4-6"},
		{"opus 1m", "claude-opus-4-6[1m]", "claude-opus-4-6[1m]"},
		{"opus 4.7 direct", "claude-opus-4-7", "claude-opus-4-7"},
		{"standard sonnet", "claude-sonnet-4-6", "claude-sonnet-4-6"},
		{"sonnet 1m", "claude-sonnet-4-6[1m]", "claude-sonnet-4-6[1m]"},
		{"opus alias 1m", "opus[1m]", "opus[1m]"},
		{"sonnet alias 1m", "sonnet[1m]", "sonnet[1m]"},
		{"haiku", "claude-haiku-4-5-20251001", "claude-haiku-4-5-20251001"},
		{"opusplan", "opusplan", "opusplan"},
		{"arbitrary model", "some-model", "some-model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandModelString(tt.model)
			if got != tt.want {
				t.Errorf("expandModelString(%q) = %q, want %q", tt.model, got, tt.want)
			}
		})
	}
}

// TestBuildEnvForLaunch verifies that CLAUDE_CODE_EFFORT_LEVEL is injected
// when EffortLevel is set and absent when empty.
func TestBuildEnvForLaunch(t *testing.T) {
	const effortKey = "CLAUDE_CODE_EFFORT_LEVEL"

	t.Run("effort xhigh injected", func(t *testing.T) {
		env := buildEnvForLaunch("xhigh", os.Environ())
		found := ""
		for _, e := range env {
			if strings.HasPrefix(e, effortKey+"=") {
				found = strings.TrimPrefix(e, effortKey+"=")
				break
			}
		}
		if found != "xhigh" {
			t.Errorf("buildEnvForLaunch: %s not set to xhigh (got %q)", effortKey, found)
		}
	})

	t.Run("empty effort leaves env unchanged", func(t *testing.T) {
		base := []string{"PATH=/usr/bin", "HOME=/root"}
		env := buildEnvForLaunch("", base)
		for _, e := range env {
			if strings.HasPrefix(e, effortKey+"=") {
				t.Errorf("buildEnvForLaunch with empty effort injected %s", e)
			}
		}
		if len(env) != len(base) {
			t.Errorf("buildEnvForLaunch with empty effort changed env length: %d -> %d", len(base), len(env))
		}
	})

	t.Run("existing effort overridden", func(t *testing.T) {
		base := []string{"PATH=/usr/bin", effortKey + "=low"}
		env := buildEnvForLaunch("xhigh", base)
		count := 0
		val := ""
		for _, e := range env {
			if strings.HasPrefix(e, effortKey+"=") {
				count++
				val = strings.TrimPrefix(e, effortKey+"=")
			}
		}
		if count != 1 {
			t.Errorf("buildEnvForLaunch: expected 1 %s entry, got %d", effortKey, count)
		}
		if val != "xhigh" {
			t.Errorf("buildEnvForLaunch: %s = %q, want xhigh", effortKey, val)
		}
	})
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

// TestApplyGLMMode_NoSettingsLocalPollution verifies that applyGLMMode does NOT
// write GLM env vars to settings.local.json. Regression test for #676.
//
// Before the fix, injectGLMEnvForTeam() in applyGLMMode permanently wrote
// ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN, DISABLE_PROMPT_CACHING, etc. to
// settings.local.json, causing GLM mode to leak into all subsequent `claude`
// invocations after `moai glm` exited.
//
// After the fix, setGLMEnv() sets the current-process env (inherited by
// syscall.Exec into `claude`), and settings.local.json is left clean.
// NOTE: does not call t.Parallel() because it modifies process-level env via setGLMEnv.
func TestApplyGLMMode_NoSettingsLocalPollution(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GLM_API_KEY", "test-glm-key-676")

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(p string, args []string) error { return nil }

	// Call applyGLMMode directly — this is what unifiedLaunch calls in GLM mode.
	if err := applyGLMMode(tmpDir, ""); err != nil {
		t.Fatalf("applyGLMMode() error: %v", err)
	}

	// settings.local.json must NOT contain GLM env vars after applyGLMMode.
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not created — correct behavior.
			return
		}
		t.Fatalf("unexpected error reading settings.local.json: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "ANTHROPIC_BASE_URL") {
		t.Error("settings.local.json must NOT contain ANTHROPIC_BASE_URL (regression: #676)")
	}
	if strings.Contains(content, "DISABLE_PROMPT_CACHING") {
		t.Error("settings.local.json must NOT contain DISABLE_PROMPT_CACHING (regression: #676)")
	}
	if strings.Contains(content, "ANTHROPIC_AUTH_TOKEN") {
		t.Error("settings.local.json must NOT contain ANTHROPIC_AUTH_TOKEN (regression: #676)")
	}
}

// TestApplyGLMMode_ProcessEnvIsSet verifies that applyGLMMode sets GLM env
// vars in the current process via setGLMEnv(). The process env is what
// syscall.Exec inherits into the launched `claude` binary.
// NOTE: does not call t.Parallel() because it modifies process-level env via setGLMEnv.
func TestApplyGLMMode_ProcessEnvIsSet(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GLM_API_KEY", "test-glm-key-676-proc")
	// Ensure a clean baseline for the vars we check.
	t.Setenv("ANTHROPIC_BASE_URL", "")
	t.Setenv("DISABLE_PROMPT_CACHING", "")

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(p string, args []string) error { return nil }

	if err := applyGLMMode(tmpDir, ""); err != nil {
		t.Fatalf("applyGLMMode() error: %v", err)
	}

	// Process env must have GLM vars set (inherited by syscall.Exec into claude).
	if got := os.Getenv("ANTHROPIC_BASE_URL"); got == "" {
		t.Error("ANTHROPIC_BASE_URL must be set in process env after applyGLMMode")
	}
	if got := os.Getenv("DISABLE_PROMPT_CACHING"); got != "1" {
		t.Errorf("DISABLE_PROMPT_CACHING must be '1' after applyGLMMode, got %q", got)
	}
}
