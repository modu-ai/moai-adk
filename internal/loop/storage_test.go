package loop

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStorage_SaveAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	now := time.Now().Truncate(time.Second)
	state := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseTest,
		Iteration: 3,
		MaxIter:   5,
		Feedback: []Feedback{
			{
				Phase:        PhaseAnalyze,
				Iteration:    1,
				TestsPassed:  10,
				TestsFailed:  2,
				LintErrors:   1,
				BuildSuccess: true,
				Coverage:     78.5,
			},
		},
		StartedAt: now,
		UpdatedAt: now,
	}

	if err := storage.SaveState(state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := storage.LoadState("SPEC-TEST-001")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.SpecID != state.SpecID {
		t.Errorf("SpecID = %q, want %q", loaded.SpecID, state.SpecID)
	}
	if loaded.Phase != state.Phase {
		t.Errorf("Phase = %q, want %q", loaded.Phase, state.Phase)
	}
	if loaded.Iteration != state.Iteration {
		t.Errorf("Iteration = %d, want %d", loaded.Iteration, state.Iteration)
	}
	if loaded.MaxIter != state.MaxIter {
		t.Errorf("MaxIter = %d, want %d", loaded.MaxIter, state.MaxIter)
	}
	if len(loaded.Feedback) != 1 {
		t.Fatalf("Feedback length = %d, want 1", len(loaded.Feedback))
	}
	if loaded.Feedback[0].TestsPassed != 10 {
		t.Errorf("Feedback[0].TestsPassed = %d, want 10", loaded.Feedback[0].TestsPassed)
	}
	if loaded.Feedback[0].Coverage != 78.5 {
		t.Errorf("Feedback[0].Coverage = %f, want 78.5", loaded.Feedback[0].Coverage)
	}
}

func TestFileStorage_SaveCreatesValidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	state := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseAnalyze,
		Iteration: 1,
		MaxIter:   5,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storage.SaveState(state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	path := filepath.Join(dir, "SPEC-TEST-001.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !json.Valid(data) {
		t.Error("saved file is not valid JSON")
	}
}

func TestFileStorage_LoadNonExistent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	state, err := storage.LoadState("SPEC-NONEXIST-001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if state != nil {
		t.Error("expected nil state")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("error should wrap os.ErrNotExist, got: %v", err)
	}
}

func TestFileStorage_LoadCorruptedJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	// Write invalid JSON.
	path := filepath.Join(dir, "SPEC-TEST-001.json")
	if err := os.WriteFile(path, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	state, err := storage.LoadState("SPEC-TEST-001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if state != nil {
		t.Error("expected nil state")
	}
	if !errors.Is(err, ErrCorruptedState) {
		t.Errorf("error should wrap ErrCorruptedState, got: %v", err)
	}
}

func TestFileStorage_LoadMissingSpecID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	// Write JSON with empty spec_id.
	data := `{"spec_id":"","phase":"analyze","iteration":1,"max_iterations":5}`
	path := filepath.Join(dir, "SPEC-EMPTY.json")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	state, err := storage.LoadState("SPEC-EMPTY")
	if err == nil {
		t.Fatal("expected error for missing spec_id")
	}
	if state != nil {
		t.Error("expected nil state")
	}
	if !errors.Is(err, ErrCorruptedState) {
		t.Errorf("error should wrap ErrCorruptedState, got: %v", err)
	}
}

func TestFileStorage_LoadInvalidPhase(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	// Write JSON with invalid phase.
	data := `{"spec_id":"SPEC-TEST","phase":"deploy","iteration":1,"max_iterations":5}`
	path := filepath.Join(dir, "SPEC-TEST.json")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	state, err := storage.LoadState("SPEC-TEST")
	if err == nil {
		t.Fatal("expected error for invalid phase")
	}
	if state != nil {
		t.Error("expected nil state")
	}
	if !errors.Is(err, ErrCorruptedState) {
		t.Errorf("error should wrap ErrCorruptedState, got: %v", err)
	}
}

func TestFileStorage_Delete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	state := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseAnalyze,
		Iteration: 1,
		MaxIter:   5,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storage.SaveState(state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	if err := storage.DeleteState("SPEC-TEST-001"); err != nil {
		t.Fatalf("DeleteState failed: %v", err)
	}

	// Verify file is gone.
	_, err := storage.LoadState("SPEC-TEST-001")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestFileStorage_DeleteNonExistent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	// Deleting non-existent file should be idempotent (no error).
	if err := storage.DeleteState("SPEC-NONEXIST"); err != nil {
		t.Errorf("DeleteState on non-existent file should not error, got: %v", err)
	}
}

func TestFileStorage_AutoCreateDirectory(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "nested", "loop")
	storage := NewFileStorage(dir)

	state := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseAnalyze,
		Iteration: 1,
		MaxIter:   5,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storage.SaveState(state); err != nil {
		t.Fatalf("SaveState should auto-create directory, got: %v", err)
	}

	// Verify directory was created.
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("path should be a directory")
	}

	// Verify file was created.
	loaded, err := storage.LoadState("SPEC-TEST-001")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if loaded.SpecID != "SPEC-TEST-001" {
		t.Errorf("SpecID = %q, want %q", loaded.SpecID, "SPEC-TEST-001")
	}
}

func TestFileStorage_SaveNilState(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	err := storage.SaveState(nil)
	if err == nil {
		t.Fatal("expected error for nil state")
	}
}

func TestFileStorage_OverwriteExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	storage := NewFileStorage(dir)

	now := time.Now()
	state1 := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseAnalyze,
		Iteration: 1,
		MaxIter:   5,
		StartedAt: now,
		UpdatedAt: now,
	}

	if err := storage.SaveState(state1); err != nil {
		t.Fatalf("SaveState (1) failed: %v", err)
	}

	state2 := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseTest,
		Iteration: 3,
		MaxIter:   5,
		StartedAt: now,
		UpdatedAt: time.Now(),
	}

	if err := storage.SaveState(state2); err != nil {
		t.Fatalf("SaveState (2) failed: %v", err)
	}

	loaded, err := storage.LoadState("SPEC-TEST-001")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.Phase != PhaseTest {
		t.Errorf("Phase = %q, want %q (should be overwritten)", loaded.Phase, PhaseTest)
	}
	if loaded.Iteration != 3 {
		t.Errorf("Iteration = %d, want 3", loaded.Iteration)
	}
}
