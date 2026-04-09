package trace

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	// defaultMaxSize is the default maximum trace file size (10MB).
	defaultMaxSize = 10 * 1024 * 1024

	// channelCapacity is the buffer size for the async write channel.
	channelCapacity = 100
)

// TraceWriter writes TraceEntry values to JSONL files asynchronously.
// Writes are non-blocking: entries are enqueued to a buffered channel and
// flushed by a background goroutine. If the channel is full, the entry is
// dropped and a warning is logged (graceful degradation, REQ-OBS-003).
// Files are rotated when they exceed maxSize bytes (REQ-OBS-006).
type TraceWriter struct {
	logDir    string
	sessionID string
	maxSize   int64

	ch   chan TraceEntry
	done chan struct{}
	once sync.Once
}

// NewTraceWriter creates a new TraceWriter and starts its background goroutine.
// logDir is the directory where trace files are stored. sessionID is embedded in
// the filename and each entry. The writer must be closed via Close() to flush
// all pending entries.
func NewTraceWriter(logDir, sessionID string) *TraceWriter {
	w := &TraceWriter{
		logDir:    logDir,
		sessionID: sessionID,
		maxSize:   defaultMaxSize,
		ch:        make(chan TraceEntry, channelCapacity),
		done:      make(chan struct{}),
	}
	go w.run()
	return w
}

// Write enqueues a TraceEntry for async writing. Returns immediately without
// blocking the caller. If the internal channel is full, the entry is dropped
// and a warning is logged (graceful degradation).
func (w *TraceWriter) Write(entry TraceEntry) {
	select {
	case w.ch <- entry:
		// Enqueued successfully.
	default:
		slog.Warn("trace: channel full, dropping entry",
			"event", entry.Event,
			"session_id", entry.SessionID,
		)
	}
}

// Close flushes all pending trace entries and stops the background goroutine.
// It is safe to call Close multiple times; subsequent calls are no-ops.
func (w *TraceWriter) Close() error {
	w.once.Do(func() {
		close(w.ch)
	})
	<-w.done
	return nil
}

// filePath returns the primary trace file path for the session.
func (w *TraceWriter) filePath() string {
	return filepath.Join(w.logDir, fmt.Sprintf("trace-%s.jsonl", w.sessionID))
}

// rotatedFilePath returns the rotated backup file path.
func (w *TraceWriter) rotatedFilePath() string {
	return filepath.Join(w.logDir, fmt.Sprintf("trace-%s.1.jsonl", w.sessionID))
}

// run is the background goroutine that reads from the channel and writes to disk.
func (w *TraceWriter) run() {
	defer close(w.done)

	for entry := range w.ch {
		if err := w.writeEntry(entry); err != nil {
			slog.Warn("trace: failed to write entry",
				"event", entry.Event,
				"error", err,
			)
		}
	}
}

// writeEntry serializes entry as a JSON line and appends it to the trace file.
// Rotates the file if it exceeds maxSize before writing.
func (w *TraceWriter) writeEntry(entry TraceEntry) error {
	// Ensure the log directory exists.
	if err := os.MkdirAll(w.logDir, 0o755); err != nil {
		return fmt.Errorf("create log dir %q: %w", w.logDir, err)
	}

	path := w.filePath()

	// Check current file size and rotate if necessary (REQ-OBS-006).
	if info, err := os.Stat(path); err == nil {
		if info.Size() >= w.maxSize {
			if err := w.rotate(path); err != nil {
				slog.Warn("trace: rotation failed, continuing with existing file",
					"path", path,
					"error", err,
				)
			}
		}
	}

	// Serialize entry as JSON.
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal trace entry: %w", err)
	}

	// Append JSON line.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open trace file %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write trace entry: %w", err)
	}
	return nil
}

// rotate renames the current trace file to the .1.jsonl backup path.
func (w *TraceWriter) rotate(path string) error {
	rotated := w.rotatedFilePath()
	// Remove previous rotation if present.
	_ = os.Remove(rotated)
	if err := os.Rename(path, rotated); err != nil {
		return fmt.Errorf("rename %q to %q: %w", path, rotated, err)
	}
	slog.Info("trace: rotated file", "from", path, "to", rotated)
	return nil
}
