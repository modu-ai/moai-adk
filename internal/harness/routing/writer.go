package routing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Writer is the append-only routing-ledger writer (REQ-HEV-007). Each Append
// opens the ledger with O_APPEND|O_CREATE|O_WRONLY and writes exactly one JSONL
// line in a single write — there is NO read-modify-write of the ledger file, so
// concurrent sessions never race on ledger contents. An in-process mutex
// serializes goroutine appends within one Writer; cross-process safety relies on
// O_APPEND atomicity for the small (< PIPE_BUF) whole-line writes (rows are
// ~300-500 B, well under the atomic-append bound on all supported platforms).
//
// @MX:ANCHOR: [AUTO] Append is the single ledger write entry point.
// @MX:REASON: fan_in >= 3: Store reroute/sweep/finalize, CLI ledger path, Stop hook finalizer.
type Writer struct {
	path string
	mu   sync.Mutex
}

// NewWriter constructs a Writer targeting the ledger at path.
func NewWriter(path string) *Writer { return &Writer{path: path} }

// Append serializes row to a single JSONL line and appends it to the ledger,
// creating the parent directory and the file if needed. It is safe for
// concurrent use by multiple goroutines.
func (w *Writer) Append(row Row) error {
	line, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("routing: marshal row: %w", err)
	}
	line = append(line, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return fmt.Errorf("routing: ensure ledger dir: %w", err)
	}
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("routing: open ledger: %w", err)
	}
	if _, err := f.Write(line); err != nil {
		_ = f.Close()
		return fmt.Errorf("routing: append ledger: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("routing: close ledger: %w", err)
	}
	return nil
}
