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

// ReplayWriter buffers agent action entries and flushes them periodically.
// It writes to a separate replay.jsonl file alongside the journal.
type ReplayWriter struct {
	mu     sync.Mutex
	path   string
	buf    []ActionEntry
	bufCap int
}

// NewReplayWriter creates a replay writer for the given SPEC directory.
// Entries are buffered (up to bufSize) and flushed on Flush or when the buffer is full.
func NewReplayWriter(specDir string, bufSize int) *ReplayWriter {
	if bufSize <= 0 {
		bufSize = 10
	}
	return &ReplayWriter{
		path:   filepath.Join(specDir, "replay.jsonl"),
		buf:    make([]ActionEntry, 0, bufSize),
		bufCap: bufSize,
	}
}

// LogAction records an agent action. The entry is buffered and flushed
// when the buffer reaches capacity. Call Flush at session boundaries to
// ensure no entries are lost.
func (rw *ReplayWriter) LogAction(entry ActionEntry) error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	rw.buf = append(rw.buf, entry)

	if len(rw.buf) >= rw.bufCap {
		return rw.flushLocked()
	}
	return nil
}

// Flush writes all buffered entries to disk. Call this at session end,
// pre-compact, or periodically to prevent data loss on crash.
func (rw *ReplayWriter) Flush() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.flushLocked()
}

// flushLocked writes buffered entries to disk. Caller must hold rw.mu.
func (rw *ReplayWriter) flushLocked() error {
	if len(rw.buf) == 0 {
		return nil
	}

	dir := filepath.Dir(rw.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create replay directory: %w", err)
	}

	f, err := os.OpenFile(rw.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open replay file: %w", err)
	}
	defer f.Close()

	for _, entry := range rw.buf {
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("marshal replay entry: %w", err)
		}
		data = append(data, '\n')
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write replay entry: %w", err)
		}
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync replay file: %w", err)
	}

	rw.buf = rw.buf[:0]
	return nil
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
