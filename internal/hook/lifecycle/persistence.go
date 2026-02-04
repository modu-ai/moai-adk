package lifecycle

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// workStateImpl implements WorkState.
type workStateImpl struct {
	config WorkStateConfig
	mu     sync.RWMutex
}

// Compile-time interface compliance check.
var _ WorkState = (*workStateImpl)(nil)

// NewWorkState creates a new WorkState instance.
func NewWorkState(config WorkStateConfig) *workStateImpl {
	return &workStateImpl{
		config: config,
	}
}

// Save persists the work state to storage.
// REQ-HOOK-380: Save work state to .moai/memory/last-session-state.json.
func (w *workStateImpl) Save(state *WorkStateData) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if state == nil {
		return nil
	}

	// Update timestamp
	state.Timestamp = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(w.config.StoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Warn("failed to create state directory",
			"dir", dir,
			"error", err.Error(),
		)
		return err
	}

	// Marshal state
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		slog.Warn("failed to marshal work state",
			"error", err.Error(),
		)
		return err
	}

	// Write to file
	if err := os.WriteFile(w.config.StoragePath, data, 0644); err != nil {
		slog.Warn("failed to write work state file",
			"path", w.config.StoragePath,
			"error", err.Error(),
		)
		return err
	}

	slog.Debug("work state saved",
		"path", w.config.StoragePath,
	)

	return nil
}

// Load retrieves the work state from storage.
// REQ-HOOK-382: Restore state on SessionStart.
func (w *workStateImpl) Load() (*WorkStateData, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(w.config.StoragePath); os.IsNotExist(err) {
		return nil, nil // No state file, return nil without error
	}

	// Read file
	data, err := os.ReadFile(w.config.StoragePath)
	if err != nil {
		slog.Warn("failed to read work state file",
			"path", w.config.StoragePath,
			"error", err.Error(),
		)
		return nil, err
	}

	// Unmarshal state
	var state WorkStateData
	if err := json.Unmarshal(data, &state); err != nil {
		slog.Warn("failed to unmarshal work state",
			"path", w.config.StoragePath,
			"error", err.Error(),
		)
		return nil, err
	}

	slog.Debug("work state loaded",
		"path", w.config.StoragePath,
		"timestamp", state.Timestamp.Format(time.RFC3339),
	)

	return &state, nil
}
