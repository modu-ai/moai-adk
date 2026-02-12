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
