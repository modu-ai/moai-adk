package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrRecordDropped is returned by AsyncRecorder.Record when the internal buffer
// is full and the record cannot be enqueued without blocking.
var ErrRecordDropped = errors.New("telemetry: record dropped (buffer full)")

// AsyncRecorder is a non-blocking telemetry writer.
// Delivers records to a single writer goroutine via an internal channel
// so that the caller path is never blocked.
//
// @MX:WARN: [AUTO] goroutine + channel pattern: goroutine lifetime is explicitly terminated via Close()
// @MX:REASON: [AUTO] without Close(), the writer goroutine waits indefinitely when GC'd, causing a goroutine leak
type AsyncRecorder struct {
	projectRoot string
	ch          chan UsageRecord
	done        chan struct{}
	wg          sync.WaitGroup
}

// NewAsyncRecorder creates and starts a new AsyncRecorder.
// bufSize is the internal channel buffer size. When the buffer is full, Record() returns ErrRecordDropped.
// The returned AsyncRecorder must be terminated with Close().
//
// @MX:ANCHOR: [AUTO] NewAsyncRecorder — async telemetry entry point
// @MX:REASON: [AUTO] public API used by many callers including hook/post_tool_metrics.go
func NewAsyncRecorder(projectRoot string, bufSize int) *AsyncRecorder {
	if bufSize < 1 {
		bufSize = 256
	}
	r := &AsyncRecorder{
		projectRoot: projectRoot,
		ch:          make(chan UsageRecord, bufSize),
		done:        make(chan struct{}),
	}
	r.wg.Add(1)
	go r.run()
	return r
}

// Record enqueues a UsageRecord for asynchronous writing.
// Returns ErrRecordDropped without blocking when the buffer is full.
func (r *AsyncRecorder) Record(rec UsageRecord) error {
	select {
	case r.ch <- rec:
		return nil
	default:
		slog.Warn("telemetry: async recorder buffer full, record dropped",
			"skill_id", rec.SkillID,
			"session_id", rec.SessionID,
		)
		return ErrRecordDropped
	}
}

// Close signals the writer goroutine to flush remaining records and exit.
// ctx controls the maximum wait duration.
// Record() must not be called after Close().
func (r *AsyncRecorder) Close(ctx context.Context) error {
	close(r.ch)

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("telemetry: async recorder close timeout: %w", ctx.Err())
	}
}

// run is the single writer goroutine. It consumes records from ch and writes
// them to daily JSONL files. It reuses the file handle for the same day to
// minimize syscall overhead (CRITICAL 6 fix).
func (r *AsyncRecorder) run() {
	defer r.wg.Done()

	dir := filepath.Join(r.projectRoot, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("telemetry: async recorder cannot create dir", "err", err)
		// Drain all records without writing and exit when directory creation fails
		for range r.ch {
		}
		return
	}

	// Per-day file handle cache (CRITICAL 6: open each day's file only once)
	var (
		currentDay string
		currentFile *os.File
		currentBuf  *bufio.Writer
	)

	flushAndClose := func() {
		if currentBuf != nil {
			_ = currentBuf.Flush()
		}
		if currentFile != nil {
			_ = currentFile.Close()
			currentFile = nil
			currentBuf = nil
		}
	}
	defer flushAndClose()

	// Flush after a certain number of records (prevent buffer accumulation)
	const flushEvery = 16
	count := 0

	for rec := range r.ch {
		// Detect day rollover using the UTC date key
		dayKey := rec.Timestamp.UTC().Format("2006-01-02")
		if dayKey != currentDay {
			flushAndClose()
			currentDay = dayKey
			path := filepath.Join(dir, "usage-"+dayKey+".jsonl")
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				slog.Error("telemetry: async recorder open file", "path", path, "err", err)
				continue
			}
			currentFile = f
			currentBuf = bufio.NewWriterSize(f, 4096)
		}

		line, err := json.Marshal(rec)
		if err != nil {
			slog.Error("telemetry: async recorder marshal", "err", err)
			continue
		}

		if _, err := currentBuf.Write(append(line, '\n')); err != nil {
			slog.Error("telemetry: async recorder write", "err", err)
			continue
		}

		count++
		if count%flushEvery == 0 {
			if err := currentBuf.Flush(); err != nil {
				slog.Error("telemetry: async recorder flush", "err", err)
			}
		}
	}
	// When the channel closes, flush remaining buffer (handled by defer flushAndClose)
}

// --- Package-level singleton ---

var (
	globalRecorder     *AsyncRecorder
	globalRecorderMu   sync.Mutex
	globalRecorderRoot string
)

// GetRecorder returns the package-level singleton AsyncRecorder for projectRoot.
// Starts the recorder on first call. When projectRoot changes, closes the existing recorder and starts a new one.
//
// @MX:NOTE: [AUTO] Singleton pattern: concentrates file I/O into a single writer goroutine per process
func GetRecorder(projectRoot string) *AsyncRecorder {
	globalRecorderMu.Lock()
	defer globalRecorderMu.Unlock()

	if globalRecorder != nil && globalRecorderRoot == projectRoot {
		return globalRecorder
	}

	// projectRoot has changed or not yet initialized
	if globalRecorder != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = globalRecorder.Close(ctx)
	}

	globalRecorder = NewAsyncRecorder(projectRoot, 256)
	globalRecorderRoot = projectRoot
	return globalRecorder
}
