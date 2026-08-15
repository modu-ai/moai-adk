// backlog_store.go — the lock-guarded backlog queue store
// (SPEC-KANBAN-TODO-CLI-001 REQ-TODO-006..013, M1).
//
// The backlog is the OPERATOR's queue, not the board's: any session may
// append, pick, or complete a card, so this store deliberately applies NO
// sole-writer role guard (the explicit contrast with board_store.go — the
// board has exactly one writer, the lead; the backlog has every writer).
// What it does share with the board is the concurrency substrate: mutations
// serialize on a sibling advisory lock (backlog.lock, the same
// path-parameterized flock/atomic-create split the board lock uses) across
// the entire read-modify-write, and land through the same-directory temp +
// atomic rename. Reads are lock-free.
//
// Ids are issued INSIDE the locked mutation from the persisted high-water
// mark `last_seq`. The mark — never max-present-id — decides the next id,
// because `done` removes rows and a derived mark would reuse the removed
// card's id (the t4/t5/t6 collision this SPEC exists to kill).
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// backlogLockFileName names the lock artifact sibling to the backlog file.
const backlogLockFileName = "backlog.lock"

// backlogVersion is the record schema version. The schema is ADDITIVE within
// version 1: `last_seq` was appended as a top-level field with no version
// bump, and no per-item field may ever be added (spec.md §E out-of-scope).
const backlogVersion = 1

// BacklogState is a queue item's lifecycle state.
type BacklogState string

const (
	// BacklogStateQueued marks a card waiting in the queue.
	BacklogStateQueued BacklogState = "queued"
	// BacklogStatePicked marks a card chosen for a SPEC (spec_id attached).
	BacklogStatePicked BacklogState = "picked"
	// BacklogStateDropped marks a card the operator discarded.
	BacklogStateDropped BacklogState = "dropped"
)

// BacklogItem is one queued card. The five fields are the frozen per-item
// contract (REQ-TODO-013): SpecID is a pointer so an absent spec id
// round-trips as JSON null, not as an omitted key.
type BacklogItem struct {
	ID      string       `json:"id"`
	Text    string       `json:"text"`
	AddedAt string       `json:"added_at"`
	SpecID  *string      `json:"spec_id"`
	State   BacklogState `json:"state"`
}

// BacklogRecord is the backlog file's document shape. LastSeq is the
// persisted id high-water mark — additive top-level, absent in files that
// predate the field (derived from max present id on load).
type BacklogRecord struct {
	Version int           `json:"version"`
	LastSeq int           `json:"last_seq"`
	Items   []BacklogItem `json:"items"`
}

// BacklogStore guards one backlog file. Load is lock-free; every mutation
// goes through Mutate, which holds the sibling lock across the whole
// read-modify-write.
type BacklogStore struct {
	path string
}

// NewBacklogStore returns a store over the backlog file at path.
func NewBacklogStore(path string) *BacklogStore {
	return &BacklogStore{path: path}
}

// Path returns the backlog file's path.
func (s *BacklogStore) Path() string { return s.path }

// LockPath returns the sibling lock artifact's path (diagnostics and tests).
func (s *BacklogStore) LockPath() string {
	return filepath.Join(filepath.Dir(s.path), backlogLockFileName)
}

// @MX:ANCHOR: [AUTO] Load — the lock-free read every backlog verb renders from
// @MX:REASON: expected fan_in >= 3 (M2 list/next/done verbs + tests); the sole load path on the file
//
// Load reads the backlog file without taking the lock. A missing file is an
// empty queue, never an error (REQ-TODO-012). A malformed file surfaces as a
// parse error with the file untouched — there is no repair-on-load path; the
// operator's queued intent is the one thing that cannot be regenerated.
func (s *BacklogStore) Load() (*BacklogRecord, error) {
	rec, err := s.load()
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// load is Load without the record copy — the shared read both the lock-free
// verbs and the locked mutation path call.
func (s *BacklogStore) load() (*BacklogRecord, error) {
	raw, err := atomicfile.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &BacklogRecord{Version: backlogVersion, Items: []BacklogItem{}}, nil
		}
		return nil, fmt.Errorf("load backlog %s: %w", s.path, err)
	}
	var rec BacklogRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, fmt.Errorf("load backlog %s: parsing: %w", s.path, err)
	}
	normalizeBacklogRecord(&rec)
	return &rec, nil
}

// @MX:WARN: [AUTO] Mutate — cross-process lock held across the entire read-modify-write
// @MX:REASON: a mutation that loads or writes outside the lock reintroduces the lost-update race this SPEC kills (REQ-TODO-006/008)
//
// Mutate is THE guarded mutation entry point: it acquires the sibling lock,
// loads the record, applies mutate in place, and atomically writes the
// result. Returning an error from mutate aborts with the file unchanged.
// Ids must be issued from rec.LastSeq inside the callback — the high-water
// mark is normalized (max of persisted and max-present) before mutate runs,
// so a version-1 file predating last_seq or a hand-edited low value both
// resolve before issuance. Release errors are JOINED into the result rather
// than discarded: on Windows release removes the artifact, so a silent
// release failure would block every later writer.
func (s *BacklogStore) Mutate(mutate func(*BacklogRecord) error) (err error) {
	lock, err := s.acquireLock()
	if err != nil {
		return err
	}
	defer func() {
		err = joinBacklogReleaseErr(err, lock.Release(), s.path)
	}()

	rec, err := s.load()
	if err != nil {
		return err
	}
	if err := mutate(rec); err != nil {
		return fmt.Errorf("mutate backlog %s: mutation refused: %w", s.path, err)
	}
	// Re-normalize post-mutate: a callback may append or rewrite items, and
	// the written high-water mark must clear every present id.
	normalizeBacklogRecord(rec)
	return s.writeAtomic(rec)
}

// joinBacklogReleaseErr folds a lock-release failure into the mutation's own
// result. It JOINS rather than overwrites, and joins in both directions: a
// release failure that arrives alongside a mutation failure is the dangerous
// case, not the harmless one. Reporting only the mutation error there hides a
// wedged lock behind an unrelated message — on Windows release removes the
// artifact, so the survivor blocks every later writer while the operator
// retries against what looks like a transient mutation fault.
//
// Returns nil only when both are nil, so the caller's `err` stays clean on the
// success path.
func joinBacklogReleaseErr(mutErr, relErr error, path string) error {
	if relErr == nil {
		return mutErr
	}
	return errors.Join(mutErr, fmt.Errorf("mutate backlog %s: lock release failed: %w", path, relErr))
}

// @MX:ANCHOR: [AUTO] Add — the id-issuing append every add-path verb calls
// @MX:REASON: expected fan_in >= 3 (M2 add verb, tests, future importers); the only id issuer, and issuance outside the lock would mint duplicates
//
// Add appends text as a new queued card and returns the issued item with its
// 1-based position among the queued cards. The id comes from the persisted
// high-water mark inside the locked mutation, so a removed card's id is
// never reused (REQ-TODO-008).
func (s *BacklogStore) Add(text string) (*BacklogItem, int, error) {
	var item BacklogItem
	var pos int
	err := s.Mutate(func(rec *BacklogRecord) error {
		rec.LastSeq++
		item = BacklogItem{
			ID:      fmt.Sprintf("t%d", rec.LastSeq),
			Text:    text,
			AddedAt: time.Now().UTC().Format(time.RFC3339),
			State:   BacklogStateQueued,
		}
		rec.Items = append(rec.Items, item)
		pos = 0
		for _, it := range rec.Items {
			if it.State == BacklogStateQueued {
				pos++
			}
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return &item, pos, nil
}

// acquireBacklogLockSerialized acquires the backlog's sibling lock, retrying
// contention for the same bounded window as the board lock (25ms x 40 ~ 1s):
// a mutation racing a short-lived holder serializes behind it instead of
// failing, while a genuinely stuck holder surfaces as an error rather than a
// hang. The timeout error names the lock artifact so the operator can act on
// the right file.
func (s *BacklogStore) acquireLock() (*BoardLock, error) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return nil, fmt.Errorf("mutate backlog %s: creating dir: %w", s.path, err)
	}
	var lastErr error
	for attempt := 0; attempt <= boardLockRetries; attempt++ {
		impl, err := acquireBoardLockImpl(s.LockPath())
		if err == nil {
			return &BoardLock{path: s.LockPath(), impl: impl}, nil
		}
		if !IsBoardLockHeld(err) {
			return nil, fmt.Errorf("mutate backlog %s: lock %s: %w", s.path, s.LockPath(), err)
		}
		lastErr = err
		time.Sleep(boardLockRetryDelay)
	}
	return nil, fmt.Errorf("mutate backlog %s: lock %s: %w", s.path, s.LockPath(), lastErr)
}

// writeAtomic persists rec through the same-directory temp + atomic rename
// (REQ-TODO-010): the rename must stay inside one filesystem to be atomic,
// and no temp residue may survive the write.
func (s *BacklogStore) writeAtomic(rec *BacklogRecord) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("write backlog %s: creating dir: %w", s.path, err)
	}
	encoded, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("write backlog %s: encoding: %w", s.path, err)
	}
	encoded = append(encoded, '\n')

	tmp, err := os.CreateTemp(dir, ".backlog-*.tmp")
	if err != nil {
		return fmt.Errorf("write backlog %s: creating temp file: %w", s.path, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(encoded); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write backlog %s: writing temp file: %w", s.path, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write backlog %s: closing temp file: %w", s.path, err)
	}
	if err := atomicfile.Replace(tmpName, s.path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write backlog %s: replacing backlog file: %w", s.path, err)
	}
	return nil
}

// normalizeBacklogRecord establishes the invariants every in-memory record
// holds: version 1, a non-nil item slice, and a high-water mark that clears
// every present id (max of persisted and max-present — REQ-TODO-009's
// derive-on-absent and the hand-edited-low-value guard in one rule).
func normalizeBacklogRecord(rec *BacklogRecord) {
	if rec.Version == 0 {
		rec.Version = backlogVersion
	}
	if rec.Items == nil {
		rec.Items = []BacklogItem{}
	}
	if max := maxPresentBacklogSeq(rec.Items); max > rec.LastSeq {
		rec.LastSeq = max
	}
}

// maxPresentBacklogSeq returns the highest numeric suffix among t<N> ids.
func maxPresentBacklogSeq(items []BacklogItem) int {
	max := 0
	for _, it := range items {
		if n, ok := parseBacklogSeq(it.ID); ok && n > max {
			max = n
		}
	}
	return max
}

// parseBacklogSeq extracts N from an id of the form t<N> (N > 0).
func parseBacklogSeq(id string) (int, bool) {
	if !strings.HasPrefix(id, "t") {
		return 0, false
	}
	n, err := strconv.Atoi(id[1:])
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
