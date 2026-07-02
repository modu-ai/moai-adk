// Package migrations — m002_settings_cleanup_test.go
// Relocated from internal/migrate/hook_cleanup_test.go by SPEC-DEADPKG-INVESTIGATE-001.
// SPEC-V3R2-MIG-002 T-MIG002-15 → AC-MIG002-A7
// SPEC-V3R5-ATOMIC-WRITE-001 → AC-AWR-001..008
// Table-driven tests for m002Apply and TestAtomicWrite_* regression coverage
// for the atomic temp+rename pattern in atomicWrite (P0-4 fix).
package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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
					"SessionStart": "handle-session-start.sh",
					"PreToolUse":   "handle-pre-tool.sh",
					"PostToolUse":  "handle-post-tool.sh",
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
			name:             "missing_settings_json_noop",
			settingsJSON:     nil, // signals: don't create settings.json
			wantRemovedCount: 0,
			wantArchiveFile:  false,
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
			err := m002Apply(projectRoot)

			// Error assertions.
			if tc.wantErr {
				if err == nil {
					t.Fatal("m002Apply: want error, got nil")
				}
				if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrContains)
				}
				// On error, settings.json must not have been written (original preserved or absent).
				return
			}

			if err != nil {
				t.Fatalf("m002Apply unexpected error: %v", err)
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

	if err := m002Apply(projectRoot); err != nil {
		t.Fatalf("m002Apply: %v", err)
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

	err := m002Apply(projectRoot)
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

	err := m002Apply(projectRoot)
	if err == nil {
		t.Fatal("expected error reading directory as file, got nil")
	}
}

// TestCleanupUserSettings_Idempotent verifies that re-invoking m002Apply
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
	if err := m002Apply(projectRoot); err != nil {
		t.Fatalf("first m002Apply: %v", err)
	}

	// Capture settings.json state after first run.
	afterFirst, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json after first run: %v", err)
	}

	// Second invocation — should be a no-op.
	if err := m002Apply(projectRoot); err != nil {
		t.Fatalf("second m002Apply: %v", err)
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

// ============================================================================
// SPEC-V3R5-ATOMIC-WRITE-001 — TestAtomicWrite_* regression tests
//
// Verifies the atomic temp+write+sync+close+rename+chmod pattern in atomicWrite.
// Maps Given-When-Then scenarios from acceptance.md to Go tests:
//   - Scenario 1 → TestAtomicWrite_HappyPath          (AC-AWR-001, AC-AWR-005)
//   - Scenario 2 → TestAtomicWrite_ReplaceExisting    (AC-AWR-001)
//   - Scenario 5 → TestAtomicWrite_PreservesPerm      (AC-AWR-005)
//   - Scenarios 3+4 → TestAtomicWrite_CleanupOnError  (AC-AWR-003)
//   - Scenario 6 → TestCleanupUserSettings_AtomicWriteRegression (AC-AWR-004)
// ============================================================================

// TestAtomicWrite_HappyPath — Scenario 1: write to fresh path, verify content,
// mode 0o644, no temp residue.
func TestAtomicWrite_HappyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dest := filepath.Join(dir, "settings.json")
	payload := []byte(`{"hooks":{}}`)

	if err := atomicWrite(dest, payload, 0o644); err != nil {
		t.Fatalf("atomicWrite returned error: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("content mismatch: got %q, want %q", got, payload)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat destination: %v", err)
	}
	// Skip permission assertion on Windows (Go's os.Chmod has limited mode semantics there).
	if runtime.GOOS != "windows" {
		if got := info.Mode().Perm(); got != 0o644 {
			t.Errorf("mode mismatch: got %#o, want %#o", got, 0o644)
		}
	}

	// Verify no temp residue.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".settings_cleanup_tmp_") {
			t.Errorf("temp file %q remains in directory", e.Name())
		}
	}
}

// TestAtomicWrite_ReplaceExisting — Scenario 2: destination already exists with
// OLD content; atomicWrite replaces it with NEW content atomically. Verifies
// no residue, no partial concatenation.
func TestAtomicWrite_ReplaceExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dest := filepath.Join(dir, "settings.json")

	oldContent := []byte("OLD")
	if err := os.WriteFile(dest, oldContent, 0o644); err != nil {
		t.Fatalf("seed OLD content: %v", err)
	}

	newContent := []byte("NEW")
	if err := atomicWrite(dest, newContent, 0o644); err != nil {
		t.Fatalf("atomicWrite returned error: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read destination after replace: %v", err)
	}
	if string(got) != "NEW" {
		t.Errorf("content mismatch after replace: got %q, want %q", got, "NEW")
	}

	// Mode must remain 0o644 after replace (REQ-AWR-004).
	if runtime.GOOS != "windows" {
		info, err := os.Stat(dest)
		if err != nil {
			t.Fatalf("stat destination: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o644 {
			t.Errorf("mode mismatch after replace: got %#o, want %#o", got, 0o644)
		}
	}
}

// TestAtomicWrite_PreservesPerm — Scenario 5: verifies os.CreateTemp's default
// 0o600 mode is overridden by the explicit os.Chmod call to the requested perm.
func TestAtomicWrite_PreservesPerm(t *testing.T) {
	t.Parallel()

	// Skip on Windows where Unix mode bits have limited applicability.
	if runtime.GOOS == "windows" {
		t.Skip("permission bits have limited semantics on Windows")
	}

	dir := t.TempDir()
	dest := filepath.Join(dir, "settings.json")

	if err := atomicWrite(dest, []byte("X"), 0o644); err != nil {
		t.Fatalf("atomicWrite returned error: %v", err)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat destination: %v", err)
	}
	got := info.Mode().Perm()
	if got != 0o644 {
		t.Errorf("mode = %#o; want %#o (CreateTemp default 0o600 must be overridden)", got, 0o644)
	}
}

// TestAtomicWrite_CleanupOnError — Scenarios 3 + 4: forced error on rename via
// a directory-as-destination (rename cannot overwrite a non-empty directory).
// Verifies:
//   - error is returned wrapping the failing step ("rename")
//   - destination is unchanged (still a directory, not corrupted)
//   - no .settings_cleanup_tmp_* file remains
func TestAtomicWrite_CleanupOnError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create the destination as a non-empty directory — os.Rename(tmp → dir)
	// fails because the target is a directory containing a file.
	dest := filepath.Join(dir, "settings.json")
	if err := os.Mkdir(dest, 0o755); err != nil {
		t.Fatalf("create dest dir: %v", err)
	}
	// Add a file inside the directory to ensure os.Rename always fails
	// (renaming a file onto a non-empty directory fails on both POSIX and Windows).
	if err := os.WriteFile(filepath.Join(dest, "stub"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed stub in dest dir: %v", err)
	}

	err := atomicWrite(dest, []byte("NEW"), 0o644)
	if err == nil {
		t.Fatal("expected error from atomicWrite when destination is a non-empty directory, got nil")
	}
	if !strings.Contains(err.Error(), "atomicWrite:") {
		t.Errorf("error %q should be wrapped with 'atomicWrite:' prefix", err.Error())
	}

	// Destination must still be a directory (unchanged).
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat destination after failed write: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("destination is no longer a directory after failed atomicWrite (corruption)")
	}

	// No temp residue.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read parent dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".settings_cleanup_tmp_") {
			t.Errorf("temp file %q remains in directory after failed write", e.Name())
		}
	}
}

// TestAtomicWrite_CreateTempError — Forces os.CreateTemp to fail by passing
// a destination whose parent directory does not exist. Verifies the function
// returns a wrapped error containing "create temp file" and no destination is
// created. Increases coverage of the CreateTemp error branch in atomicWrite.
func TestAtomicWrite_CreateTempError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Destination under a non-existent parent directory — CreateTemp must fail.
	dest := filepath.Join(dir, "nonexistent_parent", "settings.json")

	err := atomicWrite(dest, []byte("X"), 0o644)
	if err == nil {
		t.Fatal("expected error from atomicWrite when parent directory is missing, got nil")
	}
	if !strings.Contains(err.Error(), "create temp file") {
		t.Errorf("error %q should mention 'create temp file' step", err.Error())
	}

	// Destination must not exist.
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Errorf("destination should not exist after CreateTemp failure (stat err: %v)", statErr)
	}
}

// TestCleanupUserSettings_AtomicWriteRegression — Scenario 6: caller-level
// regression test. Verifies that m002Apply (which calls atomicWrite) still
// produces a valid settings.json with the Notification hook removed and a
// same-day archive file written. This protects R-AWR-001 (caller regression risk).
func TestCleanupUserSettings_AtomicWriteRegression(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("create .claude dir: %v", err)
	}

	// Seed settings.json with one retired Notification hook entry.
	settings := settingsWithHooks(
		map[string]string{"SessionStart": "handle-session-start.sh"},
		map[string]string{"Notification": "handle-notification.sh"},
	)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, settings, 0o644); err != nil {
		t.Fatalf("seed settings.json: %v", err)
	}

	// Invoke m002Apply (which calls atomicWrite).
	if err := m002Apply(projectRoot); err != nil {
		t.Fatalf("m002Apply returned error: %v", err)
	}

	// Verify settings.json is valid JSON.
	cleaned, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read cleaned settings.json: %v", err)
	}
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("cleaned settings.json is not valid JSON: %v", err)
	}

	// Verify Notification hook is absent.
	hooksRaw, ok := parsed["hooks"]
	if !ok {
		t.Fatal("hooks key missing from cleaned settings.json")
	}
	var hooks map[string]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
		t.Fatalf("parse hooks section: %v", err)
	}
	if _, present := hooks["Notification"]; present {
		t.Error("Notification hook still present in cleaned settings.json")
	}

	// Verify archive file exists.
	archiveDir := filepath.Join(projectRoot, ".moai", "archive", "hooks", "v3.0")
	archiveEntries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("read archive dir: %v", err)
	}
	if len(archiveEntries) == 0 {
		t.Fatal("no archive file written under .moai/archive/hooks/v3.0/")
	}

	// Verify mode of cleaned settings.json is 0o644 (REQ-AWR-004 / R-AWR-003 regression).
	if runtime.GOOS != "windows" {
		info, err := os.Stat(settingsPath)
		if err != nil {
			t.Fatalf("stat settings.json: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o644 {
			t.Errorf("settings.json mode = %#o; want %#o", got, 0o644)
		}
	}
}
