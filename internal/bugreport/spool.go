package bugreport

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// spoolFileName is the capture spool's name under <moai home>/state/bugreport.
const spoolFileName = "spool.jsonl"

// BugreportStoreDir is the pipeline's user-scoped store directory relative
// to the moai home (design.md section 6, D36): every bugreport store — the
// capture spool, the drain's queue, the dedupe ledger, and the outbox log —
// lives under <moai home>/state/bugreport/, never under the project tree.
// The spool carries the verdict the drain acts on, and in this SPEC's threat
// model a project file is hostile input: a repository that ships a
// well-formed verdict-moai spool at a project path must not be able to drive
// publication from a consenting user's account.
const BugreportStoreDir = "state/bugreport"

// ErrSpoolFull: the spool is bounded at DefaultBugreportSpoolMaxLines lines
// and DefaultBugreportSpoolMaxBytes bytes; capture drops a signal beyond
// either (fail-open).
var ErrSpoolFull = errors.New("bugreport: capture spool is full")

// SpoolPath returns the user-scoped spool's path.
func SpoolPath() (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, filepath.FromSlash(BugreportStoreDir), spoolFileName), nil
}

// spoolEntry is one spool line (design.md section 6): the kind, the verdict
// fixed at capture, the reason token, the moai-internal frames, and the
// optional closed-set detail. No error text, no paths, no timestamp — the
// payload the pipeline later builds derives only from these closed fields.
type spoolEntry struct {
	Kind    Kind     `json:"kind"`
	Verdict Verdict  `json:"verdict"`
	Reason  string   `json:"reason,omitempty"`
	Frames  []string `json:"frames"`
	Detail  string   `json:"detail,omitempty"`
}

// marshalSpoolEntry renders the JSONL line with its trailing newline.
func marshalSpoolEntry(entry spoolEntry) ([]byte, error) {
	line, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return append(line, '\n'), nil
}

// appendSpoolLine appends one JSONL line to the user-scoped spool, bounded
// by the line and byte ceilings. The directory is created 0700 and the file
// 0600, so the store never depends on init having run and is never
// world-readable.
func appendSpoolLine(entry spoolEntry) error {
	path, err := SpoolPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	line, err := marshalSpoolEntry(entry)
	if err != nil {
		return err
	}

	// The bounds are checked against the current file: over either ceiling
	// the signal is dropped. A 64 KiB read at capture is well inside the
	// 50 ms box.
	if info, statErr := os.Stat(path); statErr == nil {
		if info.Size()+int64(len(line)) > int64(config.DefaultBugreportSpoolMaxBytes) {
			return ErrSpoolFull
		}
		raw, readErr := os.ReadFile(path)
		if readErr == nil && bytes.Count(raw, []byte{'\n'})+1 > config.DefaultBugreportSpoolMaxLines {
			return ErrSpoolFull
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(line); err != nil {
		return err
	}
	return nil
}
