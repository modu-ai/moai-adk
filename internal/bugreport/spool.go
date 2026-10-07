package bugreport

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
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

// SpoolEntry is one spool line (design.md section 6): the kind, the verdict
// fixed at capture, the reason token, the moai-internal frames, and the
// optional closed-set detail. No error text, no paths, no timestamp — the
// payload the pipeline later builds derives only from these closed fields.
type SpoolEntry struct {
	Kind    Kind     `json:"kind"`
	Verdict Verdict  `json:"verdict"`
	Reason  string   `json:"reason,omitempty"`
	Frames  []string `json:"frames"`
	Detail  string   `json:"detail,omitempty"`
}

// marshalSpoolEntry renders the JSONL line with its trailing newline.
func marshalSpoolEntry(entry SpoolEntry) ([]byte, error) {
	line, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return append(line, '\n'), nil
}

// ReadSpool reads every entry from the user-scoped spool. A missing file is
// an empty spool; a malformed LINE is skipped (untrusted local input — the
// same distrust the read-back rule applies to the queue's detail), never a
// drain failure.
func ReadSpool() ([]SpoolEntry, error) {
	path, err := SpoolPath()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []SpoolEntry
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry SpoolEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if !entry.Kind.Valid() || !Verdict(entry.Verdict).Valid() {
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}

// ClearSpool empties the spool (the drain consumed every line). Removing
// the file rather than truncating keeps the store directory tidy; the next
// capture recreates it.
func ClearSpool() error {
	path, err := SpoolPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// appendSpoolLine appends one JSONL line to the user-scoped spool, bounded
// by the line and byte ceilings. The directory is created 0700 and the file
// 0600, so the store never depends on init having run and is never
// world-readable.
//
// The ceiling check and the append share ONE cross-process critical section
// — atomicfile.Claim on a sibling lock, the D37 discipline — because as two
// unsynchronized steps, 32 concurrent writers each saw room and the spool
// finished over its 200-line ceiling (review-gate finding #4). A writer
// that cannot take the section within its short budget drops the signal
// (fail-open): the ceiling is the point.
func appendSpoolLine(entry SpoolEntry) error {
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

	release, err := claimSpoolSection(path)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Inside the section: bounds are checked against the file exactly as it
	// will be appended to. A 64 KiB read at capture is well inside the
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

// spoolSectionClaim is the O_EXCL lock artifact's name next to the spool.
const spoolSectionClaim = "spool.lock"

// spoolSectionRetries / spoolSectionDelay bound a writer's wait for the
// section: short, because dropping the signal is always the acceptable
// outcome on the capture path.
const (
	spoolSectionRetries = 8
	spoolSectionDelay   = 5 * time.Millisecond
)

// claimSpoolSection takes the spool's critical section, returning its
// release func.
func claimSpoolSection(spoolPath string) (func() error, error) {
	lockPath := filepath.Join(filepath.Dir(spoolPath), spoolSectionClaim)
	var lastErr error
	for attempt := 0; attempt <= spoolSectionRetries; attempt++ {
		err := atomicfile.Claim(lockPath, 0o600)
		if err == nil {
			return func() error {
				if rmErr := os.Remove(lockPath); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
					return rmErr
				}
				return nil
			}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		lastErr = err
		time.Sleep(spoolSectionDelay)
	}
	return nil, lastErr
}
