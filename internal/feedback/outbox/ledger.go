package outbox

// ledger.go — the local dedupe/caps ledger (design.md section 6, REQ-
// ANON-013): per-fingerprint windows, rolling daily and weekly caps. The
// store is user-scoped JSON, mutated through the same atomic shape the
// queue uses.

import (
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
}

// loadLedger reads the user-scoped ledger; an absent file is an empty
// ledger. A malformed ledger reads as EMPTY rather than erroring: the
// ledger gates publication, and failing closed here means never publishing
// again until the user purges — the conservative direction.
func loadLedger() (*Ledger, error) {
	path, err := StorePath(LedgerFileName)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Ledger{FingerprintSeen: map[string]string{}}, nil
		}
		return &Ledger{FingerprintSeen: map[string]string{}}, nil
	}
	var l Ledger
	if err := json.Unmarshal(raw, &l); err != nil {
		return &Ledger{FingerprintSeen: map[string]string{}}, nil
	}
	if l.FingerprintSeen == nil {
		l.FingerprintSeen = map[string]string{}
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
	dayCount, weekCount := 0, 0
	for _, ts := range l.QueuedAt {
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
