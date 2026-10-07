package outbox

// ledger.go — the local dedupe/caps ledger (design.md section 6, REQ-
// ANON-013): per-fingerprint windows, rolling daily and weekly caps. The
// store is user-scoped JSON, mutated through the same atomic shape the
// queue uses.

import (
	"io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Ledger is the on-disk dedupe and caps state.
type Ledger struct {
	// FingerprintSeen maps a fingerprint to the last time it was queued or
	// sent (RFC3339).
	FingerprintSeen map[string]string `json:"fingerprint_seen"`
	// QueuedAt records the timestamps of every queued report inside the
	// rolling windows; entries older than a week are pruned on load.
	QueuedAt []string `json:"queued_at"`
	// Discarded maps a fingerprint to the time its reservation ended
	// TERMINALLY (review-gate finding 8: a send-attempt-exhausted discard).
	// The orphan recovery re-queues only genuinely unfinished reservations;
	// a discarded fingerprint inside the dedupe window is finished work, not
	// an orphan.
	Discarded map[string]string `json:"discarded,omitempty"`
}

// loadLedger reads the user-scoped ledger. An ABSENT file is an empty
// ledger. Every other failure is an ERROR, never silent amnesia (card-review
// finding, P2 — reversing the earlier empty-on-corrupt reading): a
// corrupted ledger used to read as empty history, resetting the dedupe
// window and the send caps so extra reports enqueued past an exhausted cap.
// The drain treats an unreadable ledger as a RETRYABLE failure — the item
// and the spool tail stay for the next drain — so the error preserves the
// spool where the fake-empty read consumed it. The read itself is bounded
// like the consent and spool reads: non-regular files are refused without
// opening, and the open+read runs under DefaultBugreportLedgerReadTimeBox
// with the DefaultBugreportLedgerMaxBytes size cap. On deadline the helper
// goroutine is left parked on the blocked handle; it exits when the
// blocking writer closes, and the caller never waits for it.
func loadLedger() (*Ledger, error) {
	path, err := StorePath(LedgerFileName)
	if err != nil {
		return nil, err
	}
	if info, serr := os.Stat(path); serr == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("outbox: ledger is not a regular file: %s", path)
	}
	type readResult struct {
		raw []byte
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		f, oerr := os.Open(path)
		if oerr != nil {
			done <- readResult{err: oerr}
			return
		}
		defer func() { _ = f.Close() }()
		// Read at most cap+1 bytes (review gate finding, P2): the size
		// verdict must be made BEFORE the whole file is allocated — a
		// full os.ReadFile on a huge file charged its entire size to
		// memory before the cap check could reject it.
		raw, rerr := io.ReadAll(io.LimitReader(f, config.DefaultBugreportLedgerMaxBytes+1))
		done <- readResult{raw: raw, err: rerr}
	}()
	var raw []byte
	select {
	case r := <-done:
		if r.err != nil {
			if os.IsNotExist(r.err) {
				return &Ledger{FingerprintSeen: map[string]string{}, Discarded: map[string]string{}}, nil
			}
			return nil, r.err
		}
		raw = r.raw
	case <-time.After(config.DefaultBugreportLedgerReadTimeBox):
		return nil, fmt.Errorf("outbox: ledger read exceeded its %s time box", config.DefaultBugreportLedgerReadTimeBox)
	}
	if len(raw) > config.DefaultBugreportLedgerMaxBytes {
		return nil, fmt.Errorf("outbox: ledger is %d bytes, over the %d cap — purge required", len(raw), config.DefaultBugreportLedgerMaxBytes)
	}
	var l Ledger
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, fmt.Errorf("outbox: ledger is corrupt: %w", err)
	}
	if l.FingerprintSeen == nil {
		l.FingerprintSeen = map[string]string{}
	}
	if l.Discarded == nil {
		l.Discarded = map[string]string{}
	}
	return &l, nil
}

// saveLedger persists the ledger atomically (temp + rename in the same
// directory).
func saveLedger(l *Ledger) error {
	path, err := StorePath(LedgerFileName)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ledger-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(encoded, '\n')); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(outboxFilePerm); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// FingerprintAllowed reports whether the fingerprint may be queued: no
// queue-or-send record inside the window days.
func (l *Ledger) FingerprintAllowed(fp string, now time.Time, windowDays int) bool {
	last, ok := l.FingerprintSeen[fp]
	if !ok {
		return true
	}
	lastAt, err := time.Parse(time.RFC3339, last)
	if err != nil {
		return true // an unreadable stamp gates nothing it can be read from
	}
	return now.Sub(lastAt) >= time.Duration(windowDays)*24*time.Hour
}

// GlobalCapsAllowed reports whether one more report may be queued under the
// rolling daily and weekly caps, with the cap name when not.
func (l *Ledger) GlobalCapsAllowed(now time.Time) (bool, string) {
	return l.globalCapsAllowed(now, "")
}

// GlobalCapsAllowedExcluding is GlobalCapsAllowed with one QueuedAt entry
// (matched by its exact timestamp) excluded from the count — the orphan-
// recovery adoption (review-gate findings 5 and 9): the orphaned
// reservation's slot belongs to the report being re-queued, so the caps
// must not judge the report against its own slot.
func (l *Ledger) GlobalCapsAllowedExcluding(now time.Time, exceptStamp string) (bool, string) {
	return l.globalCapsAllowed(now, exceptStamp)
}

func (l *Ledger) globalCapsAllowed(now time.Time, exceptStamp string) (bool, string) {
	dayCount, weekCount := 0, 0
	for _, ts := range l.QueuedAt {
		if exceptStamp != "" && ts == exceptStamp {
			continue
		}
		at, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}
		if now.Sub(at) < 24*time.Hour {
			dayCount++
		}
		if now.Sub(at) < 7*24*time.Hour {
			weekCount++
		}
	}
	if dayCount >= config.DefaultBugreportDailyCap {
		return false, fmt.Sprintf("daily cap of %d reports reached", config.DefaultBugreportDailyCap)
	}
	if weekCount >= config.DefaultBugreportWeeklyCap {
		return false, fmt.Sprintf("weekly cap of %d reports reached", config.DefaultBugreportWeeklyCap)
	}
	return true, ""
}

// RecordQueued stamps the fingerprint and the timestamp.
func (l *Ledger) RecordQueued(fp string, now time.Time) {
	l.FingerprintSeen[fp] = now.UTC().Format(time.RFC3339)
	l.QueuedAt = append(l.QueuedAt, now.UTC().Format(time.RFC3339))
}

// RecordQueuedAdopting stamps the fingerprint and REPLACES the old stamp's
// QueuedAt entry with the new one instead of appending — the orphan-
// recovery adoption: one report, one reservation (review-gate findings 5
// and 9). When no entry matches the old stamp (a malformed ledger), it
// falls back to appending so the reservation is never silently lost. (A
// same-second collision with another fingerprint's entry is below RFC3339's
// resolution and is accepted, matching the rollback's exact-timestamp
// match.)
func (l *Ledger) RecordQueuedAdopting(fp string, now time.Time, oldStamp string) {
	stamp := now.UTC().Format(time.RFC3339)
	l.FingerprintSeen[fp] = stamp
	for i, ts := range l.QueuedAt {
		if ts == oldStamp {
			l.QueuedAt[i] = stamp
			return
		}
	}
	l.QueuedAt = append(l.QueuedAt, stamp)
}

// MarkDiscarded records that the fingerprint's reservation ended TERMINALLY
// — the send-attempt-exhausted discard (review-gate finding 8). Inside the
// dedupe window the recovery path treats a discarded fingerprint as decided
// work; past the window a recurring capture is a new report, so the marker
// is consulted only inside the recovery branch and never gates a fresh,
// window-expired queue.
func (l *Ledger) MarkDiscarded(fp string, now time.Time) {
	if l.Discarded == nil {
		l.Discarded = map[string]string{}
	}
	l.Discarded[fp] = now.UTC().Format(time.RFC3339)
}

// TerminallyDiscardedWithin reports whether the fingerprint carries a
// terminal discard marker INSIDE the window. Existence alone is not
// authority: a marker can outlive its window (the queue-bound drop of a
// re-queued item removes the item without re-stamping the marker), and an
// expired marker must not suppress a genuinely unfinished reservation —
// scoped exactly like the sent-record reconcile, an unreadable stamp
// suppresses nothing.
func (l *Ledger) TerminallyDiscardedWithin(fp string, now time.Time, windowDays int) bool {
	stamp, ok := l.Discarded[fp]
	if !ok {
		return false
	}
	at, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		return false
	}
	return now.Sub(at) < time.Duration(windowDays)*24*time.Hour
}
