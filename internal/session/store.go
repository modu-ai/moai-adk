package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	// ErrCheckpointStale is returned when a checkpoint is older than the stale TTL.
	ErrCheckpointStale = errors.New("checkpoint is stale")
	// ErrBlockerOutstanding is returned when attempting to advance phase with unresolved blocker.
	ErrBlockerOutstanding = errors.New("blocker is outstanding, cannot advance phase")
	// ErrCheckpointInvalid is returned when checkpoint validation fails.
	ErrCheckpointInvalid = errors.New("checkpoint validation failed")
	// ErrCheckpointConcurrent is returned when concurrent checkpoint writes are detected.
	ErrCheckpointConcurrent = errors.New("concurrent checkpoint write conflict")
)

// SessionStore manages phase state persistence to .moai/state/.
type SessionStore interface {
	// Checkpoint persists the phase state to disk.
	Checkpoint(state PhaseState) error
	// Hydrate loads the phase state from disk.
	Hydrate(phase Phase, specID string) (*PhaseState, error)
	// AppendTaskLedger adds an entry to the task ledger.
	AppendTaskLedger(entry TaskLedgerEntry) error
	// WriteRunArtifact writes an artifact file for a run iteration.
	WriteRunArtifact(iterID, name string, body []byte) error
	// RecordBlocker persists a blocker report.
	RecordBlocker(report BlockerReport) error
	// ResolveBlocker marks a blocker as resolved.
	ResolveBlocker(phase Phase, specID string, resolution string) error
}

// FileSessionStore implements SessionStore using .moai/state/ directory.
type FileSessionStore struct {
	stateDir string        // .moai/state/
	staleTTL time.Duration // staleness threshold
}

// NewFileSessionStore creates a new FileSessionStore.
func NewFileSessionStore(stateDir string, staleTTL time.Duration) *FileSessionStore {
	return &FileSessionStore{
		stateDir: stateDir,
		staleTTL: staleTTL,
	}
}

// Checkpoint persists the phase state to disk with atomic write.
func (fs *FileSessionStore) Checkpoint(state PhaseState) error {
	if err := fs.ensureStateDir(); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// Check for outstanding blocker before checkpointing
	if state.BlockerRpt != nil && !state.BlockerRpt.Resolved {
		return ErrBlockerOutstanding
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	filename := fs.checkpointPath(state.Phase, state.SPECID)
	tmpFile := filename + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, filename); err != nil {
		_ = os.Remove(tmpFile) // Clean up on failure
		return fmt.Errorf("atomic rename: %w", err)
	}

	return nil
}

// Hydrate loads the phase state from disk and checks for staleness.
func (fs *FileSessionStore) Hydrate(phase Phase, specID string) (*PhaseState, error) {
	filename := fs.checkpointPath(phase, specID)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No checkpoint exists
		}
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}

	var state PhaseState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	// Check for staleness
	if time.Since(state.UpdatedAt) > fs.staleTTL {
		return nil, ErrCheckpointStale
	}

	return &state, nil
}

// AppendTaskLedger adds an entry to the task ledger markdown file.
func (fs *FileSessionStore) AppendTaskLedger(entry TaskLedgerEntry) error {
	if err := fs.ensureStateDir(); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	ledgerPath := filepath.Join(fs.stateDir, "task-ledger.md")

	// Open file in append mode, create if not exists
	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(entry.ToMarkdown()); err != nil {
		return fmt.Errorf("write entry: %w", err)
	}

	return nil
}

// WriteRunArtifact writes an artifact file for a run iteration.
func (fs *FileSessionStore) WriteRunArtifact(iterID, name string, body []byte) error {
	artifactDir := filepath.Join(fs.stateDir, "runs", iterID)
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("create artifact dir: %w", err)
	}

	artifactPath := filepath.Join(artifactDir, name)

	tmpFile := artifactPath + ".tmp"
	if err := os.WriteFile(tmpFile, body, 0644); err != nil {
		return fmt.Errorf("write artifact: %w", err)
	}

	if err := os.Rename(tmpFile, artifactPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename artifact: %w", err)
	}

	return nil
}

// RecordBlocker persists a blocker report to disk.
func (fs *FileSessionStore) RecordBlocker(report BlockerReport) error {
	if err := fs.ensureStateDir(); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	timestamp := report.Timestamp.Format("20060102-150405")
	filename := filepath.Join(fs.stateDir, fmt.Sprintf("blocker-%s-%s-%s.json", report.Kind, report.Provenance.Source, timestamp))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal blocker: %w", err)
	}

	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write tmp blocker: %w", err)
	}

	if err := os.Rename(tmpFile, filename); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename blocker: %w", err)
	}

	return nil
}

// ResolveBlocker finds and updates the most recent unresolved blocker for the phase/spec.
func (fs *FileSessionStore) ResolveBlocker(phase Phase, specID string, resolution string) error {
	pattern := filepath.Join(fs.stateDir, "blocker-*.json")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob blockers: %w", err)
	}

	// Find the most recent unresolved blocker
	var latestBlocker *BlockerReport
	var latestPath string
	var latestTime time.Time

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		var blocker BlockerReport
		if err := json.Unmarshal(data, &blocker); err != nil {
			continue
		}

		if !blocker.Resolved && blocker.Timestamp.After(latestTime) {
			latestBlocker = &blocker
			latestPath = match
			latestTime = blocker.Timestamp
		}
	}

	if latestBlocker == nil {
		return fmt.Errorf("no outstanding blocker found")
	}

	// Update the blocker
	latestBlocker.Resolve(resolution)
	data, err := json.MarshalIndent(latestBlocker, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal resolved blocker: %w", err)
	}

	tmpFile := latestPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write resolved blocker: %w", err)
	}

	if err := os.Rename(tmpFile, latestPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename resolved blocker: %w", err)
	}

	return nil
}

// ensureStateDir creates the state directory if it doesn't exist.
func (fs *FileSessionStore) ensureStateDir() error {
	return os.MkdirAll(fs.stateDir, 0755)
}

// checkpointPath returns the file path for a phase checkpoint.
func (fs *FileSessionStore) checkpointPath(phase Phase, specID string) string {
	return filepath.Join(fs.stateDir, fmt.Sprintf("checkpoint-%s-%s.json", phase, specID))
}
