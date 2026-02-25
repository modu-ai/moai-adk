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

func TestCleanupGLMSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		existingData   map[string]any
		wantRemoved    []string
		wantRemaining  []string
		wantEnvDeleted bool
	}{
		{
			name: "removes GLM env vars but keeps TEAMMATE_DISPLAY",
			existingData: map[string]any{
				"env": map[string]any{
					"ANTHROPIC_AUTH_TOKEN":         "glm-token",
					"ANTHROPIC_BASE_URL":           "https://glm.example.com",
					"ANTHROPIC_DEFAULT_OPUS_MODEL":  "glm-5",
					"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
					"ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.7-air",
					"CLAUDE_CODE_TEAMMATE_DISPLAY": "tmux",
					"OTHER_VAR":                    "keep-me",
				},
			},
			wantRemoved: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
			},
			wantRemaining:  []string{"OTHER_VAR", "CLAUDE_CODE_TEAMMATE_DISPLAY"},
			wantEnvDeleted: false, // env still has OTHER_VAR and TEAMMATE_DISPLAY
		},
		{
			name: "removes all GLM vars and deletes env section",
			existingData: map[string]any{
				"env": map[string]any{
					"ANTHROPIC_AUTH_TOKEN": "glm-token",
					"ANTHROPIC_BASE_URL":   "https://glm.example.com",
				},
			},
			wantRemoved:    []string{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL"},
			wantRemaining:   nil,
			wantEnvDeleted:  true, // env becomes empty, should be deleted
		},
		{
			name: "no GLM vars - no changes",
			existingData: map[string]any{
				"env": map[string]any{
					"OTHER_VAR": "keep-me",
				},
			},
			wantRemoved:    nil,
			wantRemaining:  []string{"OTHER_VAR"},
			wantEnvDeleted: false,
		},
		{
			name:          "no env section - no changes",
			existingData:  map[string]any{},
			wantRemoved:   nil,
			wantRemaining: nil,
			wantEnvDeleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			claudeDir := filepath.Join(tmpDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0o755); err != nil {
				t.Fatalf("setup .claude dir: %v", err)
			}

			settingsPath := filepath.Join(claudeDir, "settings.local.json")
			data, err := json.MarshalIndent(tt.existingData, "", "  ")
			if err != nil {
				t.Fatalf("marshal settings: %v", err)
			}
			if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
				t.Fatalf("write settings: %v", err)
			}

			cleanupGLMSettings(tmpDir)

			// Read back and verify
			resultData, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read result settings: %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(resultData, &result); err != nil {
				t.Fatalf("unmarshal result: %v", err)
			}

			// Check removed vars
			env, hasEnv := result["env"].(map[string]any)
			for _, key := range tt.wantRemoved {
				if hasEnv {
					if _, exists := env[key]; exists {
						t.Errorf("var %q should have been removed", key)
					}
				}
			}

			// Check remaining vars
			for _, key := range tt.wantRemaining {
				if !hasEnv {
					t.Errorf("env section was deleted but %q should remain", key)
					continue
				}
				if _, exists := env[key]; !exists {
					t.Errorf("var %q should still exist", key)
				}
			}

			// Check if env was deleted when empty
			if tt.wantEnvDeleted {
				if hasEnv {
					t.Error("env section should have been deleted when empty")
				}
			}
		})
	}
}

func TestCleanupGLMSettings_MissingFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Don't create settings.local.json - should not panic
	cleanupGLMSettings(tmpDir)
}
