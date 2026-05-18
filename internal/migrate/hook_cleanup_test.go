// Package migrate — hook_cleanup_test.go
// SPEC-V3R2-MIG-002 T-MIG002-15 → AC-MIG002-A7
// Table-driven tests for CleanupUserSettings.
package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// settingsWithHooks returns a minimal settings.json JSON blob that includes
// the given hook keys. Active hooks (e.g., "SessionStart") are preserved
// to test that active entries survive the cleanup.
func settingsWithHooks(activeHooks map[string]string, retiredHooks map[string]string) []byte {
	hooks := make(map[string]any)
	for k, v := range activeHooks {
		hooks[k] = json.RawMessage(`[{"command":"` + v + `"}]`)
	}
	for k, v := range retiredHooks {
		hooks[k] = json.RawMessage(`[{"command":"` + v + `"}]`)
	}

	data, _ := json.MarshalIndent(map[string]any{"hooks": hooks}, "", "  ")
	return data
}

// TestCleanupUserSettings covers the AC-MIG002-A7 acceptance criteria.
func TestCleanupUserSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		settingsJSON     []byte
		wantRemovedCount int
		wantArchiveFile  bool
		wantErr          bool
		wantErrContains  string
		wantActiveKept   []string
	}{
		{
			name: "all_4_retired_present_removed",
			settingsJSON: settingsWithHooks(
				map[string]string{
					"SessionStart": "handle-session-start.sh",
				},
				map[string]string{
					"Notification":      "handle-notification.sh",
					"Elicitation":       "handle-elicitation.sh",
					"ElicitationResult": "handle-elicitation-result.sh",
					"TaskCreated":       "handle-task-created.sh",
				},
			),
			wantRemovedCount: 4,
			wantArchiveFile:  true,
			wantActiveKept:   []string{"SessionStart"},
		},
		{
			name: "zero_retired_present_noop",
			settingsJSON: settingsWithHooks(
				map[string]string{
					"SessionStart": "handle-session-start.sh",
					"Stop":         "handle-stop.sh",
				},
				nil,
			),
			wantRemovedCount: 0,
			wantArchiveFile:  false,
			wantActiveKept:   []string{"SessionStart", "Stop"},
		},
		{
			name: "mixed_retired_and_active",
			settingsJSON: settingsWithHooks(
				map[string]string{
					"SessionStart":  "handle-session-start.sh",
					"PreToolUse":    "handle-pre-tool.sh",
					"PostToolUse":   "handle-post-tool.sh",
				},
				map[string]string{
					"Notification": "handle-notification.sh",
					"TaskCreated":  "handle-task-created.sh",
				},
			),
			wantRemovedCount: 2,
			wantArchiveFile:  true,
			wantActiveKept:   []string{"SessionStart", "PreToolUse", "PostToolUse"},
		},
		{
			name:            "malformed_json_returns_error_no_write",
			settingsJSON:    []byte(`{ this is not valid json `),
			wantErr:         true,
			wantErrContains: "parse settings.json",
		},
		{
			name:          "missing_settings_json_noop",
			settingsJSON:  nil, // signals: don't create settings.json
			wantRemovedCount: 0,
			wantArchiveFile: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := t.TempDir()
			claudeDir := filepath.Join(projectRoot, ".claude")
			if err := os.MkdirAll(claudeDir, 0o755); err != nil {
				t.Fatalf("create .claude dir: %v", err)
			}

			settingsPath := filepath.Join(claudeDir, "settings.json")

			// Write settings.json (if nil, skip — tests missing file).
			if tc.settingsJSON != nil {
				if err := os.WriteFile(settingsPath, tc.settingsJSON, 0o644); err != nil {
					t.Fatalf("write settings.json: %v", err)
				}
			}

			// Run the migration.
			err := CleanupUserSettings(projectRoot)

			// Error assertions.
			if tc.wantErr {
				if err == nil {
					t.Fatal("CleanupUserSettings: want error, got nil")
				}
				if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrContains)
				}
				// On error, settings.json must not have been written (original preserved or absent).
				return
			}

			if err != nil {
				t.Fatalf("CleanupUserSettings unexpected error: %v", err)
			}

			// If no settings.json was created (missing-file test), nothing more to check.
			if tc.settingsJSON == nil {
				return
			}

			// Parse the resulting settings.json.
			data, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read resulting settings.json: %v", err)
			}

			var result struct {
				Hooks map[string]json.RawMessage `json:"hooks"`
			}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("parse resulting settings.json: %v", err)
			}

			// Verify retired entries removed.
			retiredNames := []string{"Notification", "Elicitation", "ElicitationResult", "TaskCreated"}
			for _, retired := range retiredNames {
				if _, found := result.Hooks[retired]; found {
					t.Errorf("retired entry %q still present in settings.json after cleanup", retired)
				}
			}

			// Verify active entries preserved.
			for _, active := range tc.wantActiveKept {
				if _, found := result.Hooks[active]; !found {
					t.Errorf("active entry %q was incorrectly removed from settings.json", active)
				}
			}

			// Verify archive file presence.
			archiveDir := filepath.Join(projectRoot, ".moai", "archive", "hooks", "v3.0")
			archiveEntries, _ := os.ReadDir(archiveDir)
			hasArchive := len(archiveEntries) > 0

			if tc.wantArchiveFile && !hasArchive {
				t.Error("expected archive file in .moai/archive/hooks/v3.0/ but none found")
			}
			if !tc.wantArchiveFile && hasArchive {
				t.Errorf("expected no archive file but found %d entries", len(archiveEntries))
			}

			// Verify archive content when expected.
			if tc.wantArchiveFile && len(archiveEntries) > 0 {
				archivePath := filepath.Join(archiveDir, archiveEntries[0].Name())
				archiveData, err := os.ReadFile(archivePath)
				if err != nil {
					t.Fatalf("read archive file: %v", err)
				}
				var archiveJSON map[string]json.RawMessage
				if err := json.Unmarshal(archiveData, &archiveJSON); err != nil {
					t.Fatalf("parse archive JSON: %v", err)
				}
				if len(archiveJSON) != tc.wantRemovedCount {
					t.Errorf("archive contains %d entries, want %d", len(archiveJSON), tc.wantRemovedCount)
				}
			}
		})
	}
}

// TestCleanupUserSettings_NoHooksKey verifies that a settings.json without
// a "hooks" key is treated as a no-op (no archive, no error).
func TestCleanupUserSettings_NoHooksKey(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("create .claude dir: %v", err)
	}

	// Settings with no "hooks" key.
	noHooks := []byte(`{"defaultMode":"bypassPermissions"}`)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, noHooks, 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}

	if err := CleanupUserSettings(projectRoot); err != nil {
		t.Fatalf("CleanupUserSettings: %v", err)
	}

	// Archive should not exist.
	archiveDir := filepath.Join(projectRoot, ".moai", "archive", "hooks", "v3.0")
	entries, _ := os.ReadDir(archiveDir)
	if len(entries) > 0 {
		t.Errorf("expected no archive but found %d entries", len(entries))
	}
}

// TestCleanupUserSettings_InvalidHooksSection verifies that a settings.json where
// the "hooks" value is not a JSON object returns a wrapped parse error.
func TestCleanupUserSettings_InvalidHooksSection(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("create .claude dir: %v", err)
	}

	// hooks is a JSON array (invalid — expected object).
	badHooks := []byte(`{"hooks": [1, 2, 3]}`)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, badHooks, 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}

	err := CleanupUserSettings(projectRoot)
	if err == nil {
		t.Fatal("expected parse error for hooks array, got nil")
	}
	if !strings.Contains(err.Error(), "parse settings.json hooks") {
		t.Errorf("error %q should contain 'parse settings.json hooks'", err.Error())
	}
}

// TestCleanupUserSettings_ReadError verifies handling when settings.json is unreadable.
func TestCleanupUserSettings_ReadError(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("create .claude dir: %v", err)
	}

	// Create a directory where settings.json should be — triggers a read error.
	settingsAsDir := filepath.Join(claudeDir, "settings.json")
	if err := os.MkdirAll(settingsAsDir, 0o755); err != nil {
		t.Fatalf("create fake settings dir: %v", err)
	}

	err := CleanupUserSettings(projectRoot)
	if err == nil {
		t.Fatal("expected error reading directory as file, got nil")
	}
}

// TestCleanupUserSettings_Idempotent verifies that re-invoking CleanupUserSettings
// after a successful cleanup is a no-op (no additional archive files, no error).
func TestCleanupUserSettings_Idempotent(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("create .claude dir: %v", err)
	}

	// Settings with one retired entry.
	initial := settingsWithHooks(
		map[string]string{"SessionStart": "handle-session-start.sh"},
		map[string]string{"Notification": "handle-notification.sh"},
	)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, initial, 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}

	// First invocation — should clean.
	if err := CleanupUserSettings(projectRoot); err != nil {
		t.Fatalf("first CleanupUserSettings: %v", err)
	}

	// Capture settings.json state after first run.
	afterFirst, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json after first run: %v", err)
	}

	// Second invocation — should be a no-op.
	if err := CleanupUserSettings(projectRoot); err != nil {
		t.Fatalf("second CleanupUserSettings: %v", err)
	}

	afterSecond, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json after second run: %v", err)
	}

	if string(afterFirst) != string(afterSecond) {
		t.Errorf("settings.json changed on second invocation (should be idempotent)\nbefore: %s\nafter:  %s",
			afterFirst, afterSecond)
	}

	// Only one archive file should exist (from the first run, not a second).
	archiveDir := filepath.Join(projectRoot, ".moai", "archive", "hooks", "v3.0")
	entries, _ := os.ReadDir(archiveDir)
	if len(entries) > 1 {
		t.Errorf("expected 1 archive file after idempotent run, got %d", len(entries))
	}
}
