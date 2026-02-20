package lifecycle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWorkState_Save(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "last-session-state.json")

	config := WorkStateConfig{
		StoragePath: storagePath,
	}

	ws := NewWorkState(config)

	state := &WorkStateData{
		ActiveFiles: []string{
			"/path/to/file1.go",
			"/path/to/file2.go",
		},
		ContextSummary: "Working on feature implementation",
		Timestamp:      time.Now(),
	}

	err := ws.Save(state)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Fatal("state file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("failed to read state file: %v", err)
	}

	var savedState WorkStateData
	if err := json.Unmarshal(data, &savedState); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}

	if len(savedState.ActiveFiles) != 2 {
		t.Errorf("ActiveFiles length = %d, want 2", len(savedState.ActiveFiles))
	}
	if savedState.ContextSummary != "Working on feature implementation" {
		t.Errorf("ContextSummary = %q, want %q", savedState.ContextSummary, "Working on feature implementation")
	}
}

func TestWorkState_Load(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "last-session-state.json")

	// Create a state file
	state := WorkStateData{
		ActiveFiles: []string{
			"/path/to/active1.go",
			"/path/to/active2.go",
			"/path/to/active3.go",
		},
		ContextSummary: "Loaded state from file",
		Timestamp:      time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}

	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	config := WorkStateConfig{
		StoragePath: storagePath,
	}

	ws := NewWorkState(config)
	loaded, err := ws.Load()

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("Load() returned nil")
	}

	if len(loaded.ActiveFiles) != 3 {
		t.Errorf("ActiveFiles length = %d, want 3", len(loaded.ActiveFiles))
	}
	if loaded.ContextSummary != "Loaded state from file" {
		t.Errorf("ContextSummary = %q, want %q", loaded.ContextSummary, "Loaded state from file")
	}
}

func TestWorkState_LoadNonExistent(t *testing.T) {
	t.Parallel()

	config := WorkStateConfig{
		StoragePath: filepath.Join(t.TempDir(), "nonexistent.json"),
	}

	ws := NewWorkState(config)
	loaded, err := ws.Load()

	// Should return nil state without error for non-existent file
	if err != nil {
		t.Fatalf("Load() should not error on non-existent file: %v", err)
	}

	if loaded != nil {
		t.Error("Load() should return nil for non-existent file")
	}
}

func TestWorkState_LoadCorruptedFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "corrupted.json")

	// Write invalid JSON
	if err := os.WriteFile(storagePath, []byte("not valid json {"), 0644); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	config := WorkStateConfig{
		StoragePath: storagePath,
	}

	ws := NewWorkState(config)
	_, err := ws.Load()

	// Should return error for corrupted file
	if err == nil {
		t.Error("Load() should error on corrupted JSON")
	}
}

func TestWorkState_SaveCreatesDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "nested", "dir", "state.json")

	config := WorkStateConfig{
		StoragePath: storagePath,
	}

	ws := NewWorkState(config)
	state := &WorkStateData{
		ContextSummary: "Test",
		Timestamp:      time.Now(),
	}

	err := ws.Save(state)
	if err != nil {
		t.Fatalf("Save() should create directories: %v", err)
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Error("state file was not created in nested directory")
	}
}

func TestWorkState_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "concurrent.json")

	config := WorkStateConfig{
		StoragePath: storagePath,
	}

	ws := NewWorkState(config)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent Save calls
	for i := range numGoroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			state := &WorkStateData{
				ContextSummary: "concurrent test",
				Timestamp:      time.Now(),
			}
			_ = ws.Save(state)
		}(i)
	}

	// Concurrent Load calls
	for range numGoroutines {
		wg.Go(func() {
			_, _ = ws.Load()
		})
	}

	wg.Wait()

	// Should complete without panic
	loaded, err := ws.Load()
	if err != nil {
		t.Fatalf("Load() after concurrent access error = %v", err)
	}
	if loaded == nil {
		t.Fatal("Load() returned nil after concurrent access")
	}
}

func TestNewWorkState(t *testing.T) {
	t.Parallel()

	config := DefaultWorkStateConfig()
	ws := NewWorkState(config)

	if ws == nil {
		t.Fatal("NewWorkState() returned nil")
	}
}

func TestWorkState_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ WorkState = (*workStateImpl)(nil)
}

func TestWorkState_SaveNilState(t *testing.T) {
	t.Parallel()

	config := WorkStateConfig{
		StoragePath: filepath.Join(t.TempDir(), "state.json"),
	}

	ws := NewWorkState(config)
	err := ws.Save(nil)

	// Should not error on nil state
	if err != nil {
		t.Fatalf("Save(nil) error = %v", err)
	}
}

func TestWorkState_SaveInvalidPath(t *testing.T) {
	t.Parallel()

	config := WorkStateConfig{
		StoragePath: "/nonexistent/readonly/path/state.json",
	}

	ws := NewWorkState(config)
	state := &WorkStateData{
		ContextSummary: "test",
		Timestamp:      time.Now(),
	}

	err := ws.Save(state)

	// Should error on invalid path
	if err == nil {
		t.Skip("Could write to path (might be running as root)")
	}
}

func TestWorkState_LoadEmptyFile(t *testing.T) {
	t.Parallel()

	storagePath := filepath.Join(t.TempDir(), "empty.json")

	// Create empty file
	if err := os.WriteFile(storagePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	config := WorkStateConfig{
		StoragePath: storagePath,
	}

	ws := NewWorkState(config)
	_, err := ws.Load()

	// Should error on empty file (invalid JSON)
	if err == nil {
		t.Error("Load() should error on empty file")
	}
}
