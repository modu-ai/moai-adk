package ciwatch

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// StateFile is the default relative path within the project root for the watch flag.
const StateFile = ".moai/state/ci-watch-active.flag"

// WatchState is the in-process representation of the state file.
// It is serialized as single-document YAML for atomic write via tempfile+rename.
type WatchState struct {
	// PRNumber is the PR being watched.
	PRNumber int `yaml:"pr_number"`
	// StartedAt is the ISO-8601 UTC timestamp when the watch loop started.
	StartedAt time.Time `yaml:"started_at"`
	// HeartbeatAt is the ISO-8601 UTC timestamp of the last heartbeat tick.
	// If now() - HeartbeatAt > threshold (90s), the state is considered stale.
	HeartbeatAt time.Time `yaml:"heartbeat_at"`
	// RequiredChecks is the ordered list of required check context names.
	RequiredChecks []string `yaml:"required_checks,omitempty"`
	// AbortRequested is set to true by `moai pr watch --abort`.
	// The watch loop polls this field and exits cleanly when true.
	AbortRequested bool `yaml:"abort_requested"`
}

// IsStale reports whether the heartbeat timestamp is older than threshold.
// A stale state means the watch loop crashed or was interrupted; a new
// invocation may safely take over.
func (s WatchState) IsStale(threshold time.Duration) bool {
	return time.Since(s.HeartbeatAt) > threshold
}

// WriteState atomically writes ws to path via tempfile+rename.
// Intermediate directories are created if missing.
func WriteState(path string, ws WatchState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	// Atomic write: write to temp file, then rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename state file: %w", err)
	}
	return nil
}

// ReadState reads and parses the YAML state file at path.
// Returns a wrapped os.ErrNotExist error if the file does not exist.
func ReadState(path string) (WatchState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WatchState{}, err
	}

	var ws WatchState
	if err := yaml.Unmarshal(data, &ws); err != nil {
		return WatchState{}, fmt.Errorf("parse state file: %w", err)
	}
	return ws, nil
}

// Touch updates the HeartbeatAt field of the state file to now, preserving
// all other fields. Used by the watch loop's 30-second heartbeat tick.
func Touch(path string) error {
	ws, err := ReadState(path)
	if err != nil {
		return fmt.Errorf("touch read: %w", err)
	}
	ws.HeartbeatAt = time.Now().UTC()
	return WriteState(path, ws)
}

// SetAbortFlag reads the state file at path and sets AbortRequested = true.
// Safe to call from `moai pr watch --abort` concurrently with the watch loop
// because WriteState uses atomic rename.
func SetAbortFlag(path string) error {
	ws, err := ReadState(path)
	if err != nil {
		return fmt.Errorf("set abort flag read: %w", err)
	}
	ws.AbortRequested = true
	return WriteState(path, ws)
}

// DeleteState removes the state file. Called when the watch loop reaches a
// terminal state (all-pass, required-fail, timeout, or abort).
func DeleteState(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
