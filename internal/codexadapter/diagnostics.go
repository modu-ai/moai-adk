package codexadapter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DiagnosticSinkRel is the adapter's own log, relative to the project root.
//
// It is deliberately NOT stderr. On an exit-2 path stderr carries the hook's
// blocking reason or continuation prompt, which REQ-4 requires passing through
// unmodified — appending a diagnostic there would change what the model reads.
const DiagnosticSinkRel = ".moai/logs/codex-adapter.jsonl"

// RecordDiscards persists undeliverable messages and, when safe, mirrors them
// to stderr.
//
// hookBlocked reports whether the underlying hook exited 2. When it did, the
// stderr mirror is suppressed (AC-REQ-3c) but the sink record is still written:
// suppressing the operator-visible copy must never suppress the record itself.
func RecordDiscards(projectRoot string, discards []Discard, hookBlocked bool, stderr io.Writer) error {
	if len(discards) == 0 {
		return nil
	}

	path := filepath.Join(projectRoot, DiagnosticSinkRel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create diagnostic sink directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open diagnostic sink: %w", err)
	}
	defer func() { _ = f.Close() }()

	for _, d := range discards {
		line, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("marshal discard record: %w", err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("write discard record: %w", err)
		}

		if !hookBlocked && stderr != nil {
			// A failed stderr mirror must not fail the call: the sink record
			// above is the durable one, and losing the console copy is not
			// worth discarding a write that already succeeded.
			_, _ = fmt.Fprintf(stderr, "codex-adapter: dropped %q on %s (%d bytes): %s\n",
				d.Key, d.Event, d.ContentLength, d.Reason)
		}
	}

	// Close explicitly so a write error on flush is reported rather than
	// swallowed by the deferred close.
	if err := f.Close(); err != nil {
		return fmt.Errorf("close diagnostic sink: %w", err)
	}
	return nil
}
