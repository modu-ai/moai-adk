package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Writer appends journal entries to a JSONL file.
// It is crash-safe: each entry is flushed immediately after write.
type Writer struct {
	mu   sync.Mutex
	path string
}

// NewWriter creates a journal writer for the given SPEC directory.
// If specDir is empty, uses the fallback session directory.
func NewWriter(specDir string) *Writer {
	journalPath := filepath.Join(specDir, "journal.jsonl")
	return &Writer{path: journalPath}
}

// NewSessionWriter creates a journal writer for non-SPEC sessions.
func NewSessionWriter(stateDir string) *Writer {
	journalPath := filepath.Join(stateDir, "sessions", "journal.jsonl")
	return &Writer{path: journalPath}
}

// Write appends a single entry to the journal file.
// The write is atomic: data is flushed and synced to disk immediately.
func (w *Writer) Write(entry Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	dir := filepath.Dir(w.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create journal directory: %w", err)
	}

	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open journal file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal journal entry: %w", err)
	}
	data = append(data, '\n')

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write journal entry: %w", err)
	}

	// Flush to disk immediately for crash safety
	return f.Sync()
}

// LogSessionStart records a session start event.
func (w *Writer) LogSessionStart(sessionID, specID, phase string) error {
	return w.Write(Entry{
		SessionID: sessionID,
		Type:      "session_start",
		SpecID:    specID,
		Phase:     phase,
		Status:    "in_progress",
	})
}

// LogPhaseBegin records the start of a workflow phase.
func (w *Writer) LogPhaseBegin(sessionID, specID, phase, description string) error {
	return w.Write(Entry{
		SessionID: sessionID,
		Type:      "phase_begin",
		SpecID:    specID,
		Phase:     phase,
		Status:    "in_progress",
		Context:   map[string]string{"description": description},
	})
}

// LogPhaseEnd records the completion of a workflow phase.
func (w *Writer) LogPhaseEnd(sessionID, specID, phase, status string) error {
	return w.Write(Entry{
		SessionID: sessionID,
		Type:      "phase_end",
		SpecID:    specID,
		Phase:     phase,
		Status:    status,
	})
}

// LogCheckpoint records a mid-phase checkpoint for crash recovery.
func (w *Writer) LogCheckpoint(sessionID, specID, phase string, ctx map[string]string) error {
	return w.Write(Entry{
		SessionID: sessionID,
		Type:      "checkpoint",
		SpecID:    specID,
		Phase:     phase,
		Status:    "in_progress",
		Context:   ctx,
	})
}

// LogSessionEnd records a session end event with reason and token usage.
func (w *Writer) LogSessionEnd(sessionID, specID, phase, reason string, tokensUsed int) error {
	status := "completed"
	if reason != "completed" {
		status = "interrupted"
	}
	return w.Write(Entry{
		SessionID:  sessionID,
		Type:       "session_end",
		SpecID:     specID,
		Phase:      phase,
		Status:     status,
		TokensUsed: tokensUsed,
		Context:    map[string]string{"reason": reason},
	})
}
