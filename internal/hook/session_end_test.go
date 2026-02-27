package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionEndHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()

	if got := h.EventType(); got != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", got, EventSessionEnd)
	}
}

func TestSessionEndHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *HookInput
		setupDir bool
	}{
		{
			name: "normal session end",
			input: &HookInput{
				SessionID:     "sess-end-1",
				CWD:           "", // will be set in test
				HookEventName: "SessionEnd",
			},
			setupDir: true,
		},
		{
			name: "session end without project dir",
			input: &HookInput{
				SessionID:     "sess-end-2",
				CWD:           "/tmp",
				HookEventName: "SessionEnd",
			},
			setupDir: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.setupDir {
				tmpDir := t.TempDir()
				memDir := filepath.Join(tmpDir, ".moai", "memory")
				if err := os.MkdirAll(memDir, 0o755); err != nil {
					t.Fatalf("setup memory dir: %v", err)
				}
				tt.input.CWD = tmpDir
				tt.input.ProjectDir = tmpDir
			}

			h := NewSessionEndHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// SessionEnd hooks return empty JSON {} per Claude Code protocol
			// They should NOT have hookSpecificOutput set
			if got.HookSpecificOutput != nil {
				t.Error("SessionEnd hook should not set hookSpecificOutput")
			}
		})
	}
}

func TestCleanupCurrentSessionTeam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sessionID string
		teams     map[string]string // teamName -> leadSessionId
		wantGone  []string          // team dirs that should be removed
		wantKept  []string          // team dirs that should remain
	}{
		{
			name:      "removes matching session team",
			sessionID: "sess-abc-123",
			teams: map[string]string{
				"my-team":    "sess-abc-123",
				"other-team": "sess-xyz-789",
			},
			wantGone: []string{"my-team"},
			wantKept: []string{"other-team"},
		},
		{
			name:      "no match leaves all teams",
			sessionID: "sess-no-match",
			teams: map[string]string{
				"team-a": "sess-111",
				"team-b": "sess-222",
			},
			wantGone: nil,
			wantKept: []string{"team-a", "team-b"},
		},
		{
			name:      "empty teams dir",
			sessionID: "sess-empty",
			teams:     map[string]string{},
			wantGone:  nil,
			wantKept:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			homeDir := t.TempDir()
			teamsDir := filepath.Join(homeDir, ".claude", "teams")
			if err := os.MkdirAll(teamsDir, 0o755); err != nil {
				t.Fatalf("setup teams dir: %v", err)
			}

			// Create team directories with config.json
			for name, leadSessionID := range tt.teams {
				teamDir := filepath.Join(teamsDir, name)
				if err := os.MkdirAll(teamDir, 0o755); err != nil {
					t.Fatalf("create team dir %s: %v", name, err)
				}
				cfg := teamConfig{LeadSessionID: leadSessionID}
				data, err := json.Marshal(cfg)
				if err != nil {
					t.Fatalf("marshal config for %s: %v", name, err)
				}
				if err := os.WriteFile(filepath.Join(teamDir, "config.json"), data, 0o644); err != nil {
					t.Fatalf("write config for %s: %v", name, err)
				}
			}

			cleanupCurrentSessionTeam(tt.sessionID, homeDir)

			for _, name := range tt.wantGone {
				if _, err := os.Stat(filepath.Join(teamsDir, name)); !os.IsNotExist(err) {
					t.Errorf("team dir %q should have been removed", name)
				}
			}
			for _, name := range tt.wantKept {
				if _, err := os.Stat(filepath.Join(teamsDir, name)); os.IsNotExist(err) {
					t.Errorf("team dir %q should still exist", name)
				}
			}
		})
	}
}

func TestCleanupCurrentSessionTeam_MissingTeamsDir(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	// Don't create .claude/teams/ — should not panic or error
	cleanupCurrentSessionTeam("any-session", homeDir)
}

func TestCleanupCurrentSessionTeam_InvalidConfigJSON(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	teamDir := filepath.Join(teamsDir, "bad-config")
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(teamDir, "config.json"), []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	// Should not panic; directory should remain (bad config is not cleaned up)
	cleanupCurrentSessionTeam("any-session", homeDir)

	if _, err := os.Stat(teamDir); os.IsNotExist(err) {
		t.Error("team dir with invalid config should not be removed")
	}
}

func TestGarbageCollectStaleTeams(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup teams dir: %v", err)
	}

	// Create a stale team dir (modtime > 24h ago)
	staleDir := filepath.Join(teamsDir, "stale-team")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatalf("create stale dir: %v", err)
	}
	staleTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(staleDir, staleTime, staleTime); err != nil {
		t.Fatalf("set stale time: %v", err)
	}

	// Create a fresh team dir (modtime < 24h)
	freshDir := filepath.Join(teamsDir, "fresh-team")
	if err := os.MkdirAll(freshDir, 0o755); err != nil {
		t.Fatalf("create fresh dir: %v", err)
	}

	garbageCollectStaleTeams(homeDir)

	// Stale should be gone
	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Error("stale team dir should have been removed")
	}

	// Fresh should remain
	if _, err := os.Stat(freshDir); os.IsNotExist(err) {
		t.Error("fresh team dir should still exist")
	}
}

func TestGarbageCollectStaleTeams_MissingDir(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	// Don't create .claude/teams/ — should not panic
	garbageCollectStaleTeams(homeDir)
}

func TestCleanupOrphanedTmuxSessions_GracefulWithContext(t *testing.T) {
	t.Parallel()

	// With a cancelled context, the function should return without panic.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Should not panic or hang.
	cleanupOrphanedTmuxSessions(ctx)
}

// TestGLMEnvVarsToClean_IncludesAuthToken verifies that ANTHROPIC_AUTH_TOKEN
// is in the GLM tmux cleanup list. This matches pre-v2.6 behavior which had no
// login issues. moai glm/cg sets ANTHROPIC_AUTH_TOKEN in the tmux session to the
// GLM API key; not clearing it causes the next Claude Code session to
// authenticate with the GLM key against Anthropic's default base URL, resulting
// in auth failure. The user's real Claude credential is stored in ~/.claude/
// (system credential storage), not in tmux env, so clearing the tmux var is
// always safe — it either removes a GLM key or is a no-op.
func TestGLMEnvVarsToClean_IncludesAuthToken(t *testing.T) {
	t.Parallel()

	found := false
	for _, v := range glmEnvVarsToClean {
		if v == "ANTHROPIC_AUTH_TOKEN" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("glmEnvVarsToClean must contain ANTHROPIC_AUTH_TOKEN: "+
			"not clearing it leaves the GLM key in tmux, causing auth failure "+
			"in the next session; got list: %v", glmEnvVarsToClean)
	}
}

// TestGLMEnvVarsToClean_ContainsExpectedVars verifies that all expected GLM
// model routing variables are present in the cleanup list.
func TestGLMEnvVarsToClean_ContainsExpectedVars(t *testing.T) {
	t.Parallel()

	expected := []string{
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	}

	varSet := make(map[string]bool, len(glmEnvVarsToClean))
	for _, v := range glmEnvVarsToClean {
		varSet[v] = true
	}

	for _, want := range expected {
		if !varSet[want] {
			t.Errorf("glmEnvVarsToClean missing expected GLM var %q; got list: %v", want, glmEnvVarsToClean)
		}
	}
}

// TestMoaiTmuxSessionPrefix verifies the naming convention constant used to
// filter tmux sessions during cleanup. Only sessions with this prefix are
// eligible for orphan cleanup — user-created sessions are never touched.
func TestMoaiTmuxSessionPrefix(t *testing.T) {
	t.Parallel()

	if moaiTmuxSessionPrefix == "" {
		t.Fatal("moaiTmuxSessionPrefix must not be empty")
	}
	if moaiTmuxSessionPrefix != "moai-" {
		t.Errorf("moaiTmuxSessionPrefix = %q, want %q", moaiTmuxSessionPrefix, "moai-")
	}
}

// TestCleanupGLMSettingsLocal verifies that SessionEnd removes GLM env vars
// from settings.local.json and restores the backed-up OAuth token.
func TestCleanupGLMSettingsLocal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		initialEnv       map[string]string
		wantAuthToken    string // "" means the key should be absent
		wantBaseURL      bool   // true means ANTHROPIC_BASE_URL should still be present
		wantHaiku        bool
		wantSonnet       bool
		wantOpus         bool
		wantBackupToken  bool // true means MOAI_BACKUP_AUTH_TOKEN should still be present
		wantOtherPresent bool // true means non-GLM key should still be present
	}{
		{
			name: "GLM active with backup OAuth token: restore OAuth token and remove GLM vars",
			initialEnv: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "glm-api-key-12345",
				"ANTHROPIC_BASE_URL":             "https://glm.example.com/v1",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-air",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
				"MOAI_BACKUP_AUTH_TOKEN":         "oauth-token-from-claude",
				"CLAUDE_CODE_TEAMMATE_DISPLAY":   "compact",
			},
			wantAuthToken:    "oauth-token-from-claude",
			wantBaseURL:      false,
			wantHaiku:        false,
			wantSonnet:       false,
			wantOpus:         false,
			wantBackupToken:  false,
			wantOtherPresent: true,
		},
		{
			name: "GLM active without backup OAuth token: remove GLM vars, delete auth token",
			initialEnv: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "glm-api-key-only",
				"ANTHROPIC_BASE_URL":             "https://glm.example.com/v1",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-air",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
			},
			wantAuthToken:    "",
			wantBaseURL:      false,
			wantHaiku:        false,
			wantSonnet:       false,
			wantOpus:         false,
			wantBackupToken:  false,
			wantOtherPresent: false,
		},
		{
			name: "no GLM vars present: file unchanged",
			initialEnv: map[string]string{
				"CLAUDE_CODE_TEAMMATE_DISPLAY": "compact",
			},
			wantAuthToken:    "",
			wantBaseURL:      false,
			wantHaiku:        false,
			wantSonnet:       false,
			wantOpus:         false,
			wantBackupToken:  false,
			wantOtherPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp dir with .claude/settings.local.json
			projectDir := t.TempDir()
			claudeDir := filepath.Join(projectDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0o755); err != nil {
				t.Fatalf("setup .claude dir: %v", err)
			}

			settingsPath := filepath.Join(claudeDir, "settings.local.json")
			initial := map[string]any{
				"env": tt.initialEnv,
			}
			data, err := json.Marshal(initial)
			if err != nil {
				t.Fatalf("marshal initial settings: %v", err)
			}
			if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
				t.Fatalf("write settings.local.json: %v", err)
			}

			// Call the function under test
			cleanupGLMSettingsLocal(projectDir)

			// Read back the file
			result, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read settings.local.json after cleanup: %v", err)
			}

			var out map[string]any
			if err := json.Unmarshal(result, &out); err != nil {
				t.Fatalf("parse settings.local.json after cleanup: %v", err)
			}

			env, _ := out["env"].(map[string]any)

			// Check ANTHROPIC_AUTH_TOKEN
			authToken, hasAuthToken := env["ANTHROPIC_AUTH_TOKEN"]
			if tt.wantAuthToken != "" {
				if !hasAuthToken {
					t.Errorf("ANTHROPIC_AUTH_TOKEN should be present with value %q, but is absent", tt.wantAuthToken)
				} else if authToken != tt.wantAuthToken {
					t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", authToken, tt.wantAuthToken)
				}
			} else if hasAuthToken {
				t.Errorf("ANTHROPIC_AUTH_TOKEN should be absent, but has value %q", authToken)
			}

			// Check removed GLM vars
			if _, ok := env["ANTHROPIC_BASE_URL"]; ok != tt.wantBaseURL {
				t.Errorf("ANTHROPIC_BASE_URL present=%v, want present=%v", ok, tt.wantBaseURL)
			}
			if _, ok := env["ANTHROPIC_DEFAULT_HAIKU_MODEL"]; ok != tt.wantHaiku {
				t.Errorf("ANTHROPIC_DEFAULT_HAIKU_MODEL present=%v, want present=%v", ok, tt.wantHaiku)
			}
			if _, ok := env["ANTHROPIC_DEFAULT_SONNET_MODEL"]; ok != tt.wantSonnet {
				t.Errorf("ANTHROPIC_DEFAULT_SONNET_MODEL present=%v, want present=%v", ok, tt.wantSonnet)
			}
			if _, ok := env["ANTHROPIC_DEFAULT_OPUS_MODEL"]; ok != tt.wantOpus {
				t.Errorf("ANTHROPIC_DEFAULT_OPUS_MODEL present=%v, want present=%v", ok, tt.wantOpus)
			}
			if _, ok := env["MOAI_BACKUP_AUTH_TOKEN"]; ok != tt.wantBackupToken {
				t.Errorf("MOAI_BACKUP_AUTH_TOKEN present=%v, want present=%v", ok, tt.wantBackupToken)
			}

			// Check non-GLM var preservation
			if _, ok := env["CLAUDE_CODE_TEAMMATE_DISPLAY"]; ok != tt.wantOtherPresent {
				t.Errorf("CLAUDE_CODE_TEAMMATE_DISPLAY present=%v, want present=%v", ok, tt.wantOtherPresent)
			}
		})
	}
}

// TestCleanupGLMSettingsLocal_NoFile verifies that missing settings.local.json
// is handled gracefully (no panic, no error).
func TestCleanupGLMSettingsLocal_NoFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// No .claude/settings.local.json created — should not panic
	cleanupGLMSettingsLocal(projectDir)
}

// TestCleanupGLMSettingsLocal_EmptyFile verifies that an empty
// settings.local.json is handled gracefully.
func TestCleanupGLMSettingsLocal_EmptyFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if err := os.WriteFile(settingsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	cleanupGLMSettingsLocal(projectDir)
}

// TestSessionEndHandler_Handle_CleansGLMFromSettingsLocal verifies that the
// Handle method triggers settings.local.json cleanup when ProjectDir is set.
func TestSessionEndHandler_Handle_CleansGLMFromSettingsLocal(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("setup .claude dir: %v", err)
	}

	// Write settings.local.json with GLM vars and a backup token
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	initial := map[string]any{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":           "glm-key",
			"ANTHROPIC_BASE_URL":             "https://glm.example.com/v1",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-air",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
			"MOAI_BACKUP_AUTH_TOKEN":         "real-oauth-token",
		},
	}
	data, err := json.Marshal(initial)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	h := NewSessionEndHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "test-cleanup-session",
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}

	// Verify GLM vars were cleaned and OAuth token was restored
	result, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after Handle: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("parse settings after Handle: %v", err)
	}

	env, _ := out["env"].(map[string]any)
	if token, ok := env["ANTHROPIC_AUTH_TOKEN"]; !ok || token != "real-oauth-token" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %v (present=%v), want %q", token, ok, "real-oauth-token")
	}
	if _, ok := env["ANTHROPIC_BASE_URL"]; ok {
		t.Error("ANTHROPIC_BASE_URL should have been removed")
	}
	if _, ok := env["MOAI_BACKUP_AUTH_TOKEN"]; ok {
		t.Error("MOAI_BACKUP_AUTH_TOKEN should have been removed")
	}
}

// TestSessionEndHandler_Handle_CWDFallbackToProjectDir verifies that Handle
// uses CWD for GLM settings cleanup, falling back to ProjectDir for legacy.
func TestSessionEndHandler_Handle_CWDFallbackToProjectDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cwd        string
		projectDir string
		wantClean  bool // true if GLM cleanup should have occurred
	}{
		{
			name:       "CWD set: uses CWD for cleanup",
			cwd:        "SET", // placeholder, replaced with real tmpDir
			projectDir: "",
			wantClean:  true,
		},
		{
			name:       "CWD empty, ProjectDir set: falls back to ProjectDir",
			cwd:        "",
			projectDir: "SET", // placeholder, replaced with real tmpDir
			wantClean:  true,
		},
		{
			name:       "both empty: no cleanup attempted",
			cwd:        "",
			projectDir: "",
			wantClean:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var targetDir string
			if tt.wantClean {
				targetDir = t.TempDir()
				claudeDir := filepath.Join(targetDir, ".claude")
				if err := os.MkdirAll(claudeDir, 0o755); err != nil {
					t.Fatalf("setup .claude dir: %v", err)
				}
				settingsPath := filepath.Join(claudeDir, "settings.local.json")
				initial := map[string]any{
					"env": map[string]string{
						"ANTHROPIC_AUTH_TOKEN":           "glm-key",
						"ANTHROPIC_BASE_URL":             "https://glm.example.com/v1",
						"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-air",
						"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
						"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
						"MOAI_BACKUP_AUTH_TOKEN":         "real-oauth",
					},
				}
				data, err := json.Marshal(initial)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
				if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
					t.Fatalf("write settings: %v", err)
				}
			}

			input := &HookInput{
				SessionID:     "test-cwd-fallback",
				HookEventName: "SessionEnd",
			}
			if tt.cwd == "SET" {
				input.CWD = targetDir
			} else {
				input.CWD = tt.cwd
			}
			if tt.projectDir == "SET" {
				input.ProjectDir = targetDir
			} else {
				input.ProjectDir = tt.projectDir
			}

			h := NewSessionEndHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}

			if tt.wantClean {
				// Verify GLM vars were cleaned
				settingsPath := filepath.Join(targetDir, ".claude", "settings.local.json")
				result, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("read settings after Handle: %v", err)
				}
				var out map[string]any
				if err := json.Unmarshal(result, &out); err != nil {
					t.Fatalf("parse settings after Handle: %v", err)
				}
				env, _ := out["env"].(map[string]any)
				if _, ok := env["ANTHROPIC_BASE_URL"]; ok {
					t.Error("ANTHROPIC_BASE_URL should have been removed")
				}
				if token, ok := env["ANTHROPIC_AUTH_TOKEN"]; !ok || token != "real-oauth" {
					t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want %q", token, "real-oauth")
				}
			}
		})
	}
}

func TestSessionEndHandler_AlwaysReturnsEmptyOutput(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "test-always-empty",
		CWD:           t.TempDir(),
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("output should never be nil")
	}
	if got.Decision != "" {
		t.Errorf("Decision should be empty, got %q", got.Decision)
	}
	if got.ExitCode != 0 {
		t.Errorf("ExitCode should be 0, got %d", got.ExitCode)
	}
}
