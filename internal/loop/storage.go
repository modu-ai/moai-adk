package loop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileStorage implements the Storage interface using JSON files on disk.
// State files are stored at {baseDir}/{specID}.json.
// Writes are atomic: data is written to a temp file and renamed.
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new FileStorage with the given base directory.
func NewFileStorage(baseDir string) *FileStorage {
	return &FileStorage{baseDir: baseDir}
}

// SaveState persists the loop state to a JSON file using atomic write.
// The directory is created automatically if it does not exist.
func (fs *FileStorage) SaveState(state *LoopState) error {
	if state == nil {
		return fmt.Errorf("loop: storage: cannot save nil state")
	}

	if err := os.MkdirAll(fs.baseDir, 0o755); err != nil {
		return fmt.Errorf("loop: storage: create directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("loop: storage: marshal state: %w", err)
	}

	path := fs.statePath(state.SpecID)
	if err := atomicWriteFile(path, data); err != nil {
		return fmt.Errorf("loop: storage: write state: %w", err)
	}

	return nil
}

// LoadState reads and deserializes loop state from a JSON file.
// Returns ErrCorruptedState if the file contains invalid JSON.
// Returns an os.ErrNotExist-wrapping error if the file does not exist.
func (fs *FileStorage) LoadState(specID string) (*LoopState, error) {
	path := fs.statePath(specID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("loop: storage: state file not found: %w", err)
		}
		return nil, fmt.Errorf("loop: storage: read state: %w", err)
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("loop: storage: %w: invalid JSON in %s", ErrCorruptedState, path)
	}

	var state LoopState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("loop: storage: %w: %v", ErrCorruptedState, err)
	}

	// Validate essential fields.
	if state.SpecID == "" {
		return nil, fmt.Errorf("loop: storage: %w: missing spec_id", ErrCorruptedState)
	}
	if !IsValidPhase(state.Phase) {
		return nil, fmt.Errorf("loop: storage: %w: invalid phase %q", ErrCorruptedState, state.Phase)
	}

	return &state, nil
}

// DeleteState removes the state file for the given spec ID.
// Returns nil if the file does not exist (idempotent).
func (fs *FileStorage) DeleteState(specID string) error {
	path := fs.statePath(specID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loop: storage: delete state: %w", err)
	}
	return nil
}

// statePath returns the file path for a given spec ID.
func (fs *FileStorage) statePath(specID string) string {
	return filepath.Join(fs.baseDir, specID+".json")
}

// atomicWriteFile writes data to path atomically using temp file + rename.
func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".loop-state-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // cleanup on error path

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	return os.Rename(tmpName, path)
}

// Compile-time interface compliance check.
var _ Storage = (*FileStorage)(nil)
