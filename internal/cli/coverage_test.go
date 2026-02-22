package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/rank"
	"github.com/modu-ai/moai-adk/internal/update"
)

// --- Hook command coverage tests ---

func TestRunHookEvent_Success(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test-session"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, _ hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "pre-tool" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetContext(context.Background())

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runHookEvent error: %v", err)
			}
			return
		}
	}
	t.Error("pre-tool subcommand not found")
}

func TestRunHookEvent_ReadInputError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return nil, errors.New("invalid JSON")
			},
		},
		HookRegistry: &mockHookRegistry{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "post-tool" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("should error on ReadInput failure")
			}
			if !strings.Contains(err.Error(), "read hook input") {
				t.Errorf("error should mention read hook input, got %v", err)
			}
			return
		}
	}
	t.Error("post-tool subcommand not found")
}

func TestRunHookEvent_DispatchError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, _ hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				return nil, errors.New("dispatch failed")
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "session-end" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("should error on dispatch failure")
			}
			if !strings.Contains(err.Error(), "dispatch hook") {
				t.Errorf("error should mention dispatch hook, got %v", err)
			}
			return
		}
	}
	t.Error("session-end subcommand not found")
}

func TestRunHookEvent_WriteOutputError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			writeOutputFunc: func(_ io.Writer, _ *hook.HookOutput) error {
				return errors.New("write failed")
			},
		},
		HookRegistry: &mockHookRegistry{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "stop" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("should error on WriteOutput failure")
			}
			if !strings.Contains(err.Error(), "write hook output") {
				t.Errorf("error should mention write hook output, got %v", err)
			}
			return
		}
	}
	t.Error("stop subcommand not found")
}

func TestRunHookEvent_NilProtocol(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookRegistry: &mockHookRegistry{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "compact" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("should error with nil protocol")
			}
			if !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("error should mention not initialized, got %v", err)
			}
			return
		}
	}
	t.Error("compact subcommand not found")
}

func TestRunHookEvent_NilRegistry(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "session-start" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("should error with nil registry")
			}
			return
		}
	}
	t.Error("session-start subcommand not found")
}

// --- Hook list command coverage tests ---

func TestRunHookList_WithHandlers(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookRegistry: &mockHookRegistry{
			handlersFunc: func(event hook.EventType) []hook.Handler {
				if event == hook.EventSessionStart {
					return []hook.Handler{
						&mockHandler{eventType: hook.EventSessionStart},
					}
				}
				return nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runHookList error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Registered Hook Handlers") {
				t.Errorf("output should contain header, got %q", output)
			}
			if !strings.Contains(output, "SessionStart") {
				t.Errorf("output should contain SessionStart, got %q", output)
			}
			if !strings.Contains(output, "1 handler") {
				t.Errorf("output should contain handler count, got %q", output)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestRunHookList_NoHandlers(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookRegistry: &mockHookRegistry{
			handlersFunc: func(_ hook.EventType) []hook.Handler {
				return nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runHookList error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "No handlers registered") {
				t.Errorf("output should say no handlers, got %q", output)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestRunHookList_NilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runHookList with nil deps should not error, got %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "not initialized") {
				t.Errorf("output should mention not initialized, got %q", output)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

// --- Update command coverage tests ---

func TestRunUpdate_CheckOnlyWithChecker(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	t.Setenv("MOAI_UPDATE_SOURCE", "local")

	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{
			checkLatestFunc: func(_ context.Context) (*update.VersionInfo, error) {
				return &update.VersionInfo{Version: "2.0.0"}, nil
			},
		},
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background())

	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := updateCmd.Flags().Set("check", "false"); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Fatalf("runUpdate --check error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Latest version") {
		t.Errorf("output should contain 'Latest version', got %q", output)
	}
	if !strings.Contains(output, "2.0.0") {
		t.Errorf("output should contain version, got %q", output)
	}
}

func TestRunUpdate_CheckOnlyNilChecker(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	t.Setenv("MOAI_UPDATE_SOURCE", "local")

	// Set deps to nil to test the case where dependencies are not initialized.
	// With lazy initialization, EnsureUpdate would try to create a real checker
	// if deps is non-nil, so we test the nil deps path instead.
	deps = nil

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := updateCmd.Flags().Set("check", "false"); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Fatalf("runUpdate --check nil checker should not error, got %v", err)
	}

	if !strings.Contains(buf.String(), "not available") {
		t.Errorf("output should mention not available, got %q", buf.String())
	}
}

func TestRunUpdate_CheckLatestError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	t.Setenv("MOAI_UPDATE_SOURCE", "local")

	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{
			checkLatestFunc: func(_ context.Context) (*update.VersionInfo, error) {
				return nil, errors.New("network error")
			},
		},
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background())

	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := updateCmd.Flags().Set("check", "false"); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := updateCmd.RunE(updateCmd, []string{})
	if err == nil {
		t.Error("should error on CheckLatest failure")
	}
	if !strings.Contains(err.Error(), "check latest version") {
		t.Errorf("error should mention check latest, got %v", err)
	}
}

func TestRunUpdate_DefaultIsTemplateSync(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Run in a temp directory to avoid polluting the source tree with deployed templates.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Default moai update should run template sync, not binary update.
	// Even with nil orchestrator, the command should proceed to template sync.
	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{},
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background())

	if err := updateCmd.Flags().Set("check", "false"); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}

	// Default flow should attempt template sync, not binary update
	err = updateCmd.RunE(updateCmd, []string{})

	// Template sync may fail in test environment (no TTY, etc.) but
	// the error should NOT be about orchestrator or binary update.
	if err != nil {
		if strings.Contains(err.Error(), "orchestrator") {
			t.Errorf("default update should not require orchestrator, got %v", err)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Current version") {
		t.Errorf("output should contain version info, got %q", output)
	}
}

func TestRunUpdate_CheckModeShowsLatest(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{},
		Logger:        slog.Default(),
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background())

	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = updateCmd.Flags().Set("check", "false") }()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Fatalf("--check should not error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Latest version") {
		t.Errorf("output should contain 'Latest version', got %q", output)
	}
	if !strings.Contains(output, "Binary updates happen automatically") {
		t.Errorf("output should mention auto-update, got %q", output)
	}
}

func TestRunUpdate_CheckModeWithError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{
			checkLatestFunc: func(_ context.Context) (*update.VersionInfo, error) {
				return nil, errors.New("network timeout")
			},
		},
		Logger: slog.Default(),
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background())

	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = updateCmd.Flags().Set("check", "false") }()

	err := updateCmd.RunE(updateCmd, []string{})
	if err == nil {
		t.Error("--check should error on CheckLatest failure")
	}
	if !strings.Contains(err.Error(), "check latest version") {
		t.Errorf("error should mention check latest, got %v", err)
	}
}

// --- CC command coverage tests ---

func TestRunCC_WithConfig(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	tmpDir := t.TempDir()
	setupMinimalConfig(t, tmpDir)

	mgr := config.NewConfigManager()
	if _, err := mgr.Load(tmpDir); err != nil {
		t.Fatalf("Load config: %v", err)
	}

	deps = &Dependencies{Config: mgr}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{})
	if err != nil {
		t.Fatalf("runCC with config error: %v", err)
	}

	if !strings.Contains(buf.String(), "Switched to Claude backend") {
		t.Errorf("output should confirm switch, got %q", buf.String())
	}
}

func TestRunCC_NilConfig(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Create temp project
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

	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{})
	if err != nil {
		t.Fatalf("runCC nil deps should not error, got %v", err)
	}

	if !strings.Contains(buf.String(), "Claude") {
		t.Errorf("output should mention Claude, got %q", buf.String())
	}
}

// --- GLM command coverage tests ---

func TestRunGLM_WithConfig(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Enable test mode to prevent modifying actual settings files
	t.Setenv("MOAI_TEST_MODE", "1")
	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key")

	tmpDir := t.TempDir()
	setupMinimalConfig(t, tmpDir)

	mgr := config.NewConfigManager()
	if _, err := mgr.Load(tmpDir); err != nil {
		t.Fatalf("Load config: %v", err)
	}

	deps = &Dependencies{Config: mgr}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("runGLM with config error: %v", err)
	}

	// GLM Team mode should be enabled
	if !strings.Contains(buf.String(), "GLM Team mode enabled") {
		t.Errorf("output should mention GLM Team mode enabled, got %q", buf.String())
	}
}

func TestRunGLM_InjectsEnvToSettings(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key")

	tmpDir := t.TempDir()
	setupMinimalConfig(t, tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	mgr := config.NewConfigManager()
	if _, err := mgr.Load(tmpDir); err != nil {
		t.Fatalf("Load config: %v", err)
	}

	deps = &Dependencies{Config: mgr}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("runGLM error: %v", err)
	}

	// GLM Team mode should be enabled
	if !strings.Contains(buf.String(), "GLM Team mode enabled") {
		t.Errorf("output should mention GLM Team mode enabled, got %q", buf.String())
	}

	// Verify settings.local.json was created with GLM env
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.local.json should be created: %v", err)
	}
	if !strings.Contains(string(data), "ANTHROPIC_AUTH_TOKEN") {
		t.Error("settings.local.json should contain ANTHROPIC_AUTH_TOKEN")
	}
}

func TestRunGLM_NilConfig(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key")

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

	deps = nil

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("runGLM nil deps should not error, got %v", err)
	}

	// GLM Team mode should be enabled (uses defaults when deps is nil)
	if !strings.Contains(buf.String(), "GLM Team mode enabled") {
		t.Errorf("output should mention GLM Team mode enabled, got %q", buf.String())
	}
}

// --- Rank command coverage tests ---

func TestRankLogin_WithCredStore(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Create mock credentials for successful login
	mockCreds := &rank.Credentials{
		APIKey:   "test-api-key",
		Username: "testuser",
		UserID:   "test-user-id",
	}

	// Mock credential store with Save that succeeds
	// Mock browser to prevent actual browser opening during tests
	deps = &Dependencies{
		RankCredStore: &mockCredentialStore{
			saveFunc: func(_ *rank.Credentials) error {
				return nil
			},
			loadFunc: func() (*rank.Credentials, error) {
				return mockCreds, nil
			},
		},
		RankBrowser: &mockBrowser{}, // Prevents real browser from opening
	}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "login" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set a context with short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()
			cmd.SetContext(ctx)

			err := cmd.RunE(cmd, []string{})
			// OAuth flow will timeout, but we check the initial output
			_ = err // Expected to fail due to timeout

			output := buf.String()
			if !strings.Contains(output, "Opening browser") {
				t.Errorf("output should mention browser, got %q", output)
			}
			if !strings.Contains(output, "Complete authentication") {
				t.Errorf("output should mention authentication, got %q", output)
			}
			return
		}
	}
	t.Error("login subcommand not found")
}

func TestRankStatus_WithClient(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		RankClient: &mockRankClient{
			getUserRankFunc: func(_ context.Context) (*rank.UserRank, error) {
				return &rank.UserRank{
					Username: "testuser",
					Stats: &rank.UserRankStats{
						TotalTokens:   5000,
						TotalSessions: 10,
					},
				}, nil
			},
		},
	}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "status" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetContext(context.Background())

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("rank status with client error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "testuser") {
				t.Errorf("output should contain username, got %q", output)
			}
			if !strings.Contains(output, "5000") {
				t.Errorf("output should contain token count, got %q", output)
			}
			if !strings.Contains(output, "10") {
				t.Errorf("output should contain session count, got %q", output)
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestRankStatus_GetRankError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		RankClient: &mockRankClient{
			getUserRankFunc: func(_ context.Context) (*rank.UserRank, error) {
				return nil, errors.New("API error")
			},
		},
	}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "status" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("rank status should error on API failure")
			}
			if !strings.Contains(err.Error(), "get rank") {
				t.Errorf("error should mention get rank, got %v", err)
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestRankLogout_WithCredStore(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deleteCalled := false
	deps = &Dependencies{
		RankCredStore: &mockCredentialStore{
			deleteFunc: func() error {
				deleteCalled = true
				return nil
			},
		},
	}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "logout" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("rank logout error: %v", err)
			}

			if !deleteCalled {
				t.Error("Delete should have been called")
			}
			if !strings.Contains(buf.String(), "Logged out") {
				t.Errorf("output should mention logged out, got %q", buf.String())
			}
			return
		}
	}
	t.Error("logout subcommand not found")
}

func TestRankLogout_DeleteError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		RankCredStore: &mockCredentialStore{
			deleteFunc: func() error {
				return errors.New("permission denied")
			},
		},
	}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "logout" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("rank logout should error on Delete failure")
			}
			if !strings.Contains(err.Error(), "delete credentials") {
				t.Errorf("error should mention delete credentials, got %v", err)
			}
			return
		}
	}
	t.Error("logout subcommand not found")
}

// --- Doctor command coverage tests ---

func TestCheckGit_Verbose(t *testing.T) {
	check := checkGit(true)
	if check.Status == CheckOK && check.Detail == "" {
		t.Error("verbose git check should include detail")
	}
	if check.Status == CheckOK && !strings.Contains(check.Detail, "path:") {
		t.Errorf("verbose git detail should contain path, got %q", check.Detail)
	}
}

func TestCheckMoAIConfig_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("restore dir: %v", chErr)
		}
	}()

	check := checkMoAIConfig(true)
	if check.Status != CheckOK {
		t.Errorf("status = %q, want ok", check.Status)
	}
	if check.Detail == "" {
		t.Error("verbose check should include detail")
	}
	if !strings.Contains(check.Detail, "path:") {
		t.Errorf("detail should contain path, got %q", check.Detail)
	}
}

func TestCheckMoAIConfig_MissingSections(t *testing.T) {
	tmpDir := t.TempDir()
	// Create .moai/ but not .moai/config/sections/
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("restore dir: %v", chErr)
		}
	}()

	check := checkMoAIConfig(false)
	if check.Status != CheckWarn {
		t.Errorf("status = %q, want warn for missing sections", check.Status)
	}
	if !strings.Contains(check.Message, "sections") {
		t.Errorf("message should mention sections, got %q", check.Message)
	}
}

func TestCheckClaudeConfig_Present(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("restore dir: %v", chErr)
		}
	}()

	check := checkClaudeConfig(false)
	if check.Status != CheckOK {
		t.Errorf("status = %q, want ok for present .claude/", check.Status)
	}
	if !strings.Contains(check.Message, "found") {
		t.Errorf("message should mention found, got %q", check.Message)
	}
}

func TestCheckClaudeConfig_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("restore dir: %v", chErr)
		}
	}()

	check := checkClaudeConfig(true)
	if check.Status != CheckOK {
		t.Errorf("status = %q, want ok", check.Status)
	}
	if check.Detail == "" {
		t.Error("verbose should include detail")
	}
	if !strings.Contains(check.Detail, "path:") {
		t.Errorf("detail should contain path, got %q", check.Detail)
	}
}

func TestRunDoctor_FixFlag(t *testing.T) {
	// Run in a tmpDir with no .moai so MoAI Config check produces a warn
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("restore dir: %v", chErr)
		}
	}()

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	if err := doctorCmd.Flags().Set("fix", "true"); err != nil {
		t.Fatal(err)
	}
	if err := doctorCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatal(err)
	}
	if err := doctorCmd.Flags().Set("export", ""); err != nil {
		t.Fatal(err)
	}
	if err := doctorCmd.Flags().Set("check", ""); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := doctorCmd.Flags().Set("fix", "false"); err != nil {
			t.Logf("reset: %v", err)
		}
	}()

	err = doctorCmd.RunE(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("doctor --fix error: %v", err)
	}

	// Output may or may not contain "Suggested fixes" depending on whether any check fails.
	// The fix code path is still exercised either way.
	output := buf.String()
	if !strings.Contains(output, "passed") {
		t.Errorf("output should contain summary with 'passed', got %q", output)
	}
}

func TestRunDoctor_CheckFilter(t *testing.T) {
	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	if err := doctorCmd.Flags().Set("check", "Go Runtime"); err != nil {
		t.Fatal(err)
	}
	if err := doctorCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatal(err)
	}
	if err := doctorCmd.Flags().Set("fix", "false"); err != nil {
		t.Fatal(err)
	}
	if err := doctorCmd.Flags().Set("export", ""); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := doctorCmd.Flags().Set("check", ""); err != nil {
			t.Logf("reset: %v", err)
		}
	}()

	err := doctorCmd.RunE(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("doctor --check error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Go Runtime") {
		t.Errorf("output should contain filtered check name, got %q", output)
	}
	// Should NOT contain other checks
	if strings.Contains(output, "MoAI Config") {
		t.Errorf("output should not contain unfiltered checks, got %q", output)
	}
}

// --- Statusline command coverage tests ---

func TestRunStatusline_NilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	buf := new(bytes.Buffer)
	StatuslineCmd.SetOut(buf)
	StatuslineCmd.SetErr(buf)

	err := StatuslineCmd.RunE(StatuslineCmd, []string{})
	if err != nil {
		t.Fatalf("statusline nil deps error: %v", err)
	}

	output := buf.String()
	// Statusline should produce some output (git status, version, branch, or fallback)
	output = strings.TrimSpace(output)
	if output == "" {
		t.Errorf("output should not be empty")
	}
	// The new independent collection always shows git status or version when available
	// Check for any of the statusline indicators (emoji or content)
	if !strings.Contains(output, "📊") && !strings.Contains(output, "🔅") && !strings.Contains(output, "🔀") {
		// If no indicators, at least check for some content
		if len(output) < 3 {
			t.Errorf("output should have meaningful content, got %q", output)
		}
	}
}

// --- Init command coverage tests ---

func TestRunInit_WithRootFlag(t *testing.T) {
	tmpDir := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	if err := initCmd.Flags().Set("root", tmpDir); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := initCmd.Flags().Set("root", ""); err != nil {
			t.Logf("reset: %v", err)
		}
		if err := initCmd.Flags().Set("non-interactive", "false"); err != nil {
			t.Logf("reset: %v", err)
		}
	}()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("runInit error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("output should confirm init, got %q", output)
	}
}

func TestRunInit_WithNameAndLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	if err := initCmd.Flags().Set("root", tmpDir); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("name", "test-project"); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, flag := range []string{"root", "name", "language"} {
			if err := initCmd.Flags().Set(flag, ""); err != nil {
				t.Logf("reset %s: %v", flag, err)
			}
		}
		if err := initCmd.Flags().Set("non-interactive", "false"); err != nil {
			t.Logf("reset: %v", err)
		}
	}()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("runInit error: %v", err)
	}

	if !strings.Contains(buf.String(), "MoAI project initialized") {
		t.Errorf("output should confirm init, got %q", buf.String())
	}
}

// --- Helper functions ---

// setupMinimalConfig creates a minimal .moai config directory for testing.
func setupMinimalConfig(t *testing.T, dir string) {
	t.Helper()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(sectionsDir, "user.yaml"),
		[]byte("user:\n  name: test\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(sectionsDir, "language.yaml"),
		[]byte("language:\n  conversation_language: en\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(sectionsDir, "quality.yaml"),
		[]byte("constitution:\n  development_mode: ddd\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
}

// setupMinimalConfigWithMode creates a minimal .moai config with a specific mode.
// Currently unused but kept for future test expansions.
func setupMinimalConfigWithMode(t *testing.T, dir string, mode string) { //nolint:unused
	t.Helper()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(sectionsDir, "user.yaml"),
		[]byte("user:\n  name: test\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(sectionsDir, "language.yaml"),
		[]byte("language:\n  conversation_language: en\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(sectionsDir, "quality.yaml"),
		fmt.Appendf(nil, "constitution:\n  development_mode: %s\n", mode),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
}
