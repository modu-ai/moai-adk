package homestate

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// recordUnavailableFile is the log of factory-record writes a dispatch path
// could not make (REQ-FR-025). It sits beside factory.db, outside the
// database, so it survives a database write failure.
const recordUnavailableFile = "record-unavailable.jsonl"

// RecordUnavailableEntry is one failed factory-record write: which run, card,
// and lane the dispatch recorded, when, and the write error. Reconciled turns
// true once a later successful write for the run has appended its
// `record.drift` event.
type RecordUnavailableEntry struct {
	ID         string `json:"id"`
	At         string `json:"at"`
	RunID      string `json:"run_id"`
	CardID     string `json:"card_id"`
	Lane       string `json:"lane"`
	Error      string `json:"error"`
	Reconciled bool   `json:"reconciled"`
}

// RecordUnavailablePath returns the log path for the project at root.
func RecordUnavailablePath(root string) (string, error) {
	dir, err := FactoryDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, recordUnavailableFile), nil
}

// AppendRecordUnavailable appends one entry to the project's log.
func AppendRecordUnavailable(root string, e RecordUnavailableEntry) error {
	path, err := RecordUnavailablePath(root)
	if err != nil {
		return err
	}
	if e.ID == "" {
		var b [8]byte
		if _, err := rand.Read(b[:]); err != nil {
			return err
		}
		e.ID = hex.EncodeToString(b[:])
	}
	if e.At == "" {
		e.At = time.Now().UTC().Format(time.RFC3339Nano)
	}
	e.Reconciled = false
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	unlock, err := lockRecordUnavailable(path)
	if err != nil {
		return err
	}
	defer unlock()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// readRecordUnavailableFile parses the log line by line. A torn line is a
// warning, not a failure (F2): it is skipped and counted, so one corrupt
// append never blinds the report or the reconciliation (REQ-FR-025).
func readRecordUnavailableFile(path string) ([]RecordUnavailableEntry, int, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	var out []RecordUnavailableEntry
	skipped := 0
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e RecordUnavailableEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			skipped++
			continue
		}
		out = append(out, e)
	}
	return out, skipped, sc.Err()
}

// ReadRecordUnavailable returns the unreconciled entries of runID (every run
// when runID is empty) plus the number of unparseable lines it skipped. It
// never writes, and a torn line is reported through the skipped count rather
// than failing the read (REQ-FR-025).
func ReadRecordUnavailable(root, runID string) ([]RecordUnavailableEntry, int, error) {
	path, err := RecordUnavailablePath(root)
	if err != nil {
		return nil, 0, err
	}
	all, skipped, err := readRecordUnavailableFile(path)
	if err != nil {
		return nil, 0, err
	}
	var out []RecordUnavailableEntry
	for _, e := range all {
		if !e.Reconciled && (runID == "" || e.RunID == runID) {
			out = append(out, e)
		}
	}
	return out, skipped, nil
}

// reconcileUnavailable appends one `record.drift` event per unreconciled log
// entry of runID inside the caller's transaction, and returns the settle step
// withCardTx runs once the transaction is over: settle(true) after the commit
// marks those entries reconciled, settle(false) on any other outcome releases
// whatever the reconciliation holds (nothing, in the waiting flow). The drift
// event names the dispatched card and lane and the factory record's state for
// that card as this write found it, or "absent". A context carrying
// WithBoundedReconcile takes the claim's non-waiting flow instead.
func (f *FactoryDB) reconcileUnavailable(ctx context.Context, tx *sql.Tx, runID string, now time.Time) (func(committed bool), error) {
	path := filepath.Join(filepath.Dir(f.Path), recordUnavailableFile)
	if boundedReconcile(ctx) {
		return f.reconcileUnavailableBounded(ctx, tx, path, runID, now)
	}
	entries, _, err := readRecordUnavailableFile(path)
	if err != nil || len(entries) == 0 {
		// An unreadable log never blocks a factory-record write. Torn lines
		// are skipped by the reader, so valid entries still reconcile.
		return nil, nil
	}
	ids, err := f.appendDriftEvents(ctx, tx, runID, entries, now)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return func(committed bool) {
		if committed {
			_ = markRecordUnavailableReconciled(path, ids)
		}
	}, nil
}

// appendDriftEvents appends one `record.drift` event per unreconciled entry of
// runID inside tx and returns the ids of the entries it covered.
func (f *FactoryDB) appendDriftEvents(ctx context.Context, tx *sql.Tx, runID string, entries []RecordUnavailableEntry, now time.Time) (map[string]bool, error) {
	ids := map[string]bool{}
	for _, e := range entries {
		if e.Reconciled || e.RunID != runID {
			continue
		}
		state := "absent"
		if c, err := loadCard(ctx, tx, runID, e.CardID); err == nil {
			state = c.State
		} else if !errors.Is(err, ErrCardNotFound) {
			return nil, err
		}
		if err := appendEvent(ctx, tx, runID, "record.drift", map[string]any{
			"card_id": e.CardID, "lane": e.Lane, "dispatched_at": e.At, "factory_state": state, "error": e.Error,
		}, now); err != nil {
			return nil, err
		}
		ids[e.ID] = true
	}
	return ids, nil
}

// lockRecordUnavailable takes the exclusive lock (the admission-lock
// primitive: flock on unix, LockFileEx on windows) that serializes appends to
// the unavailable-record log with its reconciliation rewrite.
func lockRecordUnavailable(path string) (func(), error) {
	impl, err := acquireAdmissionLock(path + ".lock")
	if err != nil {
		return nil, fmt.Errorf("lock %s: %w", filepath.Base(path), err)
	}
	return func() { _ = impl.release() }, nil
}

// tryLockRecordUnavailable is lockRecordUnavailable without the wait: held
// reports whether another holder has the lock, in which case no lock is taken.
func tryLockRecordUnavailable(path string) (unlock func(), held bool, err error) {
	impl, ok, err := tryAcquireAdmissionLock(path + ".lock")
	if err != nil {
		return nil, false, fmt.Errorf("lock %s: %w", filepath.Base(path), err)
	}
	if !ok {
		return nil, true, nil
	}
	return func() { _ = impl.release() }, false, nil
}

type boundedReconcileKey struct{}

// WithBoundedReconcile marks ctx as the context of a lease claim's record
// write, the one write path whose drift-log reconciliation never waits for the
// log's lock (SPEC-FACTORY-ATOMIC-LEASE-001 REQ-FAL-014, plan D2). The marker
// rides the context of one call, never the connection, so every write that
// does not receive it keeps the waiting flow.
func WithBoundedReconcile(ctx context.Context) context.Context {
	return context.WithValue(ctx, boundedReconcileKey{}, true)
}

func boundedReconcile(ctx context.Context) bool {
	v, _ := ctx.Value(boundedReconcileKey{}).(bool)
	return v
}

// reconcileUnavailableBounded is the claim's flow (plan D2): (1) the unlocked
// read finding no unreconciled entry for the run means no lock, no event and no
// mark; (2) otherwise the log's lock is tried without waiting, and a holder
// means the whole reconciliation is skipped for this write — no events, no
// mark; (3) with the lock held the log is re-read under it, an entry is
// reconciled only if it is still unreconciled there, and the lock is kept
// through the commit; (4) settle(true) marks the entries under the held lock
// and releases it, settle(false) releases it without marking. The skip is
// decided before any event is appended, so a skipped write never leaves events
// that the next write would append again.
//
// @MX:WARN: [AUTO] holds the drift log's lock across one record write's transaction while the record's write lock is held
// @MX:REASON: lock order is queue lock, then record write lock, then drift-log lock, and no path holds the log's lock and then waits for the record; a settle that never runs would leak the lock until the process ends
// @MX:SPEC: SPEC-FACTORY-ATOMIC-LEASE-001
func (f *FactoryDB) reconcileUnavailableBounded(ctx context.Context, tx *sql.Tx, path, runID string, now time.Time) (func(committed bool), error) {
	pending := func(entries []RecordUnavailableEntry) bool {
		for _, e := range entries {
			if !e.Reconciled && e.RunID == runID {
				return true
			}
		}
		return false
	}
	entries, _, err := readRecordUnavailableFile(path)
	if err != nil || !pending(entries) {
		return nil, nil
	}
	if recordUnavailableAfterReadHook != nil {
		recordUnavailableAfterReadHook()
	}
	unlock, held, err := tryLockRecordUnavailable(path)
	if err != nil || held {
		// A lock that cannot be tried never blocks the claim, and a holder
		// leaves the entries for a later write.
		return nil, nil
	}
	entries, _, err = readRecordUnavailableFile(path)
	if err != nil || !pending(entries) {
		unlock()
		return nil, nil
	}
	ids, err := f.appendDriftEvents(ctx, tx, runID, entries, now)
	if err != nil {
		unlock()
		return nil, err
	}
	return func(committed bool) {
		if committed {
			_ = rewriteRecordUnavailableReconciledLocked(path, ids)
		}
		unlock()
	}, nil
}

// recordUnavailableAfterReadHook is a test seam for the claim-scoped
// reconciliation: it is called between the unlocked read of the log and the
// claim's try for the log's lock, and only on the claim-scoped path. It is nil
// in production.
var recordUnavailableAfterReadHook func()

// recordUnavailableRewriteHook is a test seam called between the rewrite's
// read of the log and its replacement of the file. It is nil in production.
var recordUnavailableRewriteHook func()

// markRecordUnavailableReconciled rewrites the log with the given entries
// marked reconciled (write-to-temp, then rename).
func markRecordUnavailableReconciled(path string, ids map[string]bool) error {
	// The read and the rename run under the same lock as every append, so an
	// entry appended mid-rewrite cannot land in the file the rename replaces.
	unlock, err := lockRecordUnavailable(path)
	if err != nil {
		return err
	}
	defer unlock()
	return rewriteRecordUnavailableReconciledLocked(path, ids)
}

// rewriteRecordUnavailableReconciledLocked is the rewrite of
// markRecordUnavailableReconciled for a caller that already holds the log's
// lock (the lock is a non-reentrant flock, so taking it again would wait on the
// caller's own descriptor).
func rewriteRecordUnavailableReconciledLocked(path string, ids map[string]bool) error {
	// Torn lines the reader skipped drop out here: the rewrite is the one
	// place the log self-heals.
	entries, _, err := readRecordUnavailableFile(path)
	if err != nil {
		return err
	}
	if recordUnavailableRewriteHook != nil {
		recordUnavailableRewriteHook()
	}
	var buf bytes.Buffer
	for _, e := range entries {
		if ids[e.ID] {
			e.Reconciled = true
		}
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		buf.Write(append(line, '\n'))
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), recordUnavailableFile+".*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}
