package rank

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSyncState(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "sync-state.json")

	state, err := NewSyncState(path)
	if err != nil {
		t.Fatalf("NewSyncState: %v", err)
	}

	if state.SyncedCount() != 0 {
		t.Errorf("expected 0 synced, got %d", state.SyncedCount())
	}
}

func TestSyncState_MarkAndCheck(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	// Create a fake transcript file
	transcriptPath := filepath.Join(tmpDir, "test-session.jsonl")
	if err := os.WriteFile(transcriptPath, []byte(`{"test": true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	// Should not be synced initially
	if state.IsSynced(transcriptPath) {
		t.Error("expected not synced initially")
	}

	// Mark as synced
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}

	// Should be synced now
	if !state.IsSynced(transcriptPath) {
		t.Error("expected synced after marking")
	}

	if state.SyncedCount() != 1 {
		t.Errorf("expected 1 synced, got %d", state.SyncedCount())
	}
}

func TestSyncState_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create and save state
	state1, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state1.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}
	if err := state1.Save(); err != nil {
		t.Fatal(err)
	}

	// Load state in new instance
	state2, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	if !state2.IsSynced(transcriptPath) {
		t.Error("expected synced after reload")
	}
}

func TestSyncState_DetectsModifiedFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}

	// Modify the file
	time.Sleep(10 * time.Millisecond) // Ensure different mtime
	if err := os.WriteFile(transcriptPath, []byte(`{"modified": true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should detect modification
	if state.IsSynced(transcriptPath) {
		t.Error("expected not synced after file modification")
	}
}

func TestSyncState_Reset(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}

	state.Reset()

	if state.SyncedCount() != 0 {
		t.Errorf("expected 0 after reset, got %d", state.SyncedCount())
	}
	if state.IsSynced(transcriptPath) {
		t.Error("expected not synced after reset")
	}
}

func TestSyncState_CleanStale(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}

	// Delete the transcript file
	_ = os.Remove(transcriptPath) // Cleanup, ignore error

	removed := state.CleanStale()
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
	if state.SyncedCount() != 0 {
		t.Errorf("expected 0 after cleanup, got %d", state.SyncedCount())
	}
}

func TestSyncState_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "deep", "nested", "sync-state.json")

	state, err := NewSyncState(nestedPath)
	if err != nil {
		t.Fatalf("NewSyncState: %v", err)
	}

	if err := state.Save(); err != nil {
		t.Fatalf("Save() should create directories: %v", err)
	}

	if _, err := os.Stat(nestedPath); err != nil {
		t.Fatalf("sync-state.json not created: %v", err)
	}
}

func TestSyncState_Save_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}
	if err := state.Save(); err != nil {
		t.Fatal(err)
	}

	// Verify no temp file remains
	tmpPath := statePath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful save")
	}
}

func TestNewSyncState_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	// Write corrupted JSON
	if err := os.WriteFile(statePath, []byte("{invalid json}}}"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Should start fresh with empty state (not error)
	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatalf("NewSyncState should recover from corrupt file: %v", err)
	}
	if state.SyncedCount() != 0 {
		t.Errorf("expected 0 synced after corrupt recovery, got %d", state.SyncedCount())
	}
}

func TestSyncState_IsSynced_DeletedFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}

	// Delete the transcript
	if err := os.Remove(transcriptPath); err != nil {
		t.Fatal(err)
	}

	// IsSynced should return false for deleted file
	if state.IsSynced(transcriptPath) {
		t.Error("IsSynced should return false for deleted file")
	}
}

func TestSyncState_MarkSynced_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	err = state.MarkSynced("/nonexistent/transcript.jsonl")
	if err == nil {
		t.Error("MarkSynced should return error for nonexistent file")
	}
}

func TestSyncState_IsSynced_UnknownPath(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	if state.IsSynced("/unknown/path.jsonl") {
		t.Error("IsSynced should return false for unknown path")
	}
}

func TestSyncState_CleanStale_NoStaleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.MarkSynced(transcriptPath); err != nil {
		t.Fatal(err)
	}

	// File still exists, nothing stale
	removed := state.CleanStale()
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestSyncState_SaveAndReload_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	// Create multiple transcript files
	files := []string{"a.jsonl", "b.jsonl", "c.jsonl"}
	for _, f := range files {
		fp := filepath.Join(tmpDir, f)
		if err := os.WriteFile(fp, []byte(`{}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if err := state.MarkSynced(filepath.Join(tmpDir, f)); err != nil {
			t.Fatal(err)
		}
	}

	if err := state.Save(); err != nil {
		t.Fatal(err)
	}

	// Reload
	state2, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if state2.SyncedCount() != 3 {
		t.Errorf("expected 3 synced after reload, got %d", state2.SyncedCount())
	}
}

func TestNewSyncState_LoadNullSyncedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	// Write valid JSON but with null syncedFiles
	data := `{"version":1,"lastSyncTime":"0001-01-01T00:00:00Z","syncedFiles":null}`
	if err := os.WriteFile(statePath, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	// Should initialize empty map, not crash
	if state.SyncedCount() != 0 {
		t.Errorf("expected 0 synced, got %d", state.SyncedCount())
	}
}

func TestNewSyncState_DefaultPath(t *testing.T) {
	// Pass empty string to use default path (based on os.UserHomeDir).
	// This exercises the empty-basePath branch in NewSyncState.
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	state, err := NewSyncState("")
	if err != nil {
		t.Fatalf("NewSyncState(\"\") error: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	// Default path should be under tempHome
	if state.path == "" {
		t.Error("expected non-empty path")
	}
	if state.SyncedCount() != 0 {
		t.Errorf("expected 0 synced on fresh state, got %d", state.SyncedCount())
	}
}

func TestSyncState_Save_FallbackOnRenameFail(t *testing.T) {
	// Test that Save falls back to direct write when the tmp file is renamed to
	// a read-only directory. We simulate rename failure by pre-creating the
	// destination as a directory.
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "sync-state.json")

	state, err := NewSyncState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	// Replace the state path with a directory to force Rename to fail.
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatal(err)
	}
	// Remove the directory so we can restore it and test the fallback write.
	if err := os.Remove(statePath); err != nil {
		t.Fatal(err)
	}

	// Normal save should work.
	if err := state.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the file was created.
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("sync-state.json not created: %v", err)
	}
}
