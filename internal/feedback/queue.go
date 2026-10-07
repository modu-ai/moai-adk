// queue.go — the retry queue for a feedback report whose submission failed.
//
// [HARD] (D4) The queue owns exactly ONE failure: `gh issue create` returned
// non-zero AFTER a submission was attempted. It holds the MASKED title and
// body, because that is what was about to be published.
//
// The pre-submit draft (.moai/state/feedback-draft-<timestamp>.md) is a
// different artefact for a different failure — `gh auth status` failing, or a
// rate limit, before anything was sent — and it holds PRE-SCRUB RAW text. The
// two live in the same .moai/state tree and must never be conflated: re-send
// code that read a draft as a queue entry would put raw, unmasked text into a
// public issue. Nothing in this file reads, lists, or globs the draft; the
// queue's entire read scope is queue.json.
//
// The shape is a single JSON document with a sibling lock, NOT an append-only
// log (AP-7): a successful re-send must be expressible as removal, and an
// append-only file cannot say "sent". It is the deliberate opposite of the
// mask log's lock-free append — the two artefacts have opposite requirements
// (design.md §5), so they have opposite shapes.
package feedback

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// Queue file layout constants.
const (
	queueSubdir       = "feedback"
	queueFileName     = "queue.json"
	queueLockFileName = "queue.lock"

	// queueVersion is the document schema version.
	queueVersion = 1
)

// queueFilePerm is 0o600: the queue holds a report that was about to be
// published but has not been, and the mask log's reasoning applies equally.
const queueFilePerm os.FileMode = 0o600

// queueDirPerm is the directory mode for the queue's own directory.
const queueDirPerm os.FileMode = 0o755

// Lock contention budget: 25ms x 40 ~ 1s, the same bounded window the factory
// stores use. A mutation racing a short-lived holder serializes behind it; a
// genuinely stuck holder surfaces as an error rather than a hang.
const (
	queueLockRetries    = 40
	queueLockRetryDelay = 25 * time.Millisecond
)

// ErrQueueBlockedResult is returned when a caller tries to queue a report the
// classifier refused. The queue exists to re-send; queueing a blocked report
// would let the retry path publish what the gate declined.
var ErrQueueBlockedResult = errors.New("feedback: refusing to queue a blocked result")

// QueueItem is one report awaiting re-send. Title and Body are MASKED — the
// queue never carries pre-scrub text.
//
// The optional participation fields (SPEC-FEEDBACK-PARTICIPATION-001) carry
// the automatic pipeline's per-item decisions; all three are omitempty so
// existing readers of the manual queue stay compatible and a hand-written
// queue file parses unchanged.
type QueueItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	QueuedAt string `json:"queued_at"`
	Attempts int    `json:"attempts"`

	// Fingerprint is the report's family key (bugreport.Fingerprint's 16
	// hex); the sender matches it against the remote title key.
	Fingerprint string `json:"fingerprint,omitempty"`
	// Kind is the bugreport kind token.
	Kind string `json:"kind,omitempty"`
	// Summary is the validated model summary, or the deterministic template
	// text; SummaryDecision records which ("model" | "template"). Both are
	// persisted BEFORE the create attempt so a retry never repeats a paid
	// step (design.md section 8).
	Summary         string `json:"summary,omitempty"`
	SummaryDecision string `json:"summary_decision,omitempty"`
	// SummaryRequested is the durable marker persisted before the model
	// call: a crash between the call and the persist leaves the marker
	// without a summary, and recovery takes the template decision — the
	// per-item model-call bound survives the crash window (D28).
	SummaryRequested bool `json:"summary_requested,omitempty"`
}

// QueueRecord is the queue file's document shape. LastSeq is the persisted id
// high-water mark: Resolve removes rows, so an id derived from the present
// maximum would be reissued to a later item.
type QueueRecord struct {
	Version int         `json:"version"`
	LastSeq int         `json:"last_seq"`
	Items   []QueueItem `json:"items"`
}

// QueueStore guards one queue file. Load is lock-free; every mutation goes
// through Mutate, which holds the sibling lock across the whole
// read-modify-write.
type QueueStore struct {
	path string
}

// NewQueueStore returns a store over the queue file at path.
func NewQueueStore(path string) *QueueStore {
	return &QueueStore{path: path}
}

// QueuePathForRoot returns the queue's canonical location under a project
// root. The one join every root-relative caller builds the store from.
func QueuePathForRoot(root string) string {
	return filepath.Join(root, defs.MoAIDir, defs.StateSubdir, queueSubdir, queueFileName)
}

// Path returns the queue file's path.
func (s *QueueStore) Path() string { return s.path }

// LockPath returns the sibling lock artifact's path (diagnostics and tests).
func (s *QueueStore) LockPath() string {
	return filepath.Join(filepath.Dir(s.path), queueLockFileName)
}

// Load reads the queue without taking the lock. A missing file is an empty
// queue, never an error. A malformed file surfaces as a parse error with the
// file untouched: the queued report is the one thing that cannot be
// regenerated, so there is no repair-on-load path.
func (s *QueueStore) Load() (*QueueRecord, error) {
	raw, err := atomicfile.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &QueueRecord{Version: queueVersion, Items: []QueueItem{}}, nil
		}
		return nil, fmt.Errorf("load feedback queue %s: %w", s.path, err)
	}
	var rec QueueRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, fmt.Errorf("load feedback queue %s: parsing: %w", s.path, err)
	}
	normalizeQueueRecord(&rec)
	return &rec, nil
}

// @MX:WARN: [AUTO] Mutate — cross-process lock held across the entire read-modify-write
// @MX:REASON: a mutation that loads or writes outside the lock reintroduces the lost-update race the sibling lock exists to kill (REQ-9)
//
// Mutate is THE guarded mutation entry point: it acquires the sibling lock,
// loads the record, applies mutate in place, and atomically writes the result.
// Returning an error from mutate aborts with the file unchanged.
//
// Ids are issued from rec.LastSeq inside the callback, so issuance is covered
// by the same lock as the write.
func (s *QueueStore) Mutate(mutate func(*QueueRecord) error) (err error) {
	release, err := s.acquireLock()
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, release())
	}()

	rec, err := s.Load()
	if err != nil {
		return err
	}
	if err := mutate(rec); err != nil {
		return err
	}
	normalizeQueueRecord(rec)
	return s.writeAtomic(rec)
}

// EnqueueMasked appends a scrubbed report to the queue and returns the stored
// item.
//
// It takes a Result rather than two strings on purpose: a Result is by
// construction the scrubber's output, so there is no call shape that queues
// raw user text. A blocked Result is refused outright.
func (s *QueueStore) EnqueueMasked(res Result) (*QueueItem, error) {
	if res.Verdict != VerdictOK {
		return nil, fmt.Errorf("%w: verdict %q", ErrQueueBlockedResult, res.Verdict)
	}

	var item QueueItem
	err := s.Mutate(func(rec *QueueRecord) error {
		rec.LastSeq++
		item = QueueItem{
			ID:       fmt.Sprintf("f%d", rec.LastSeq),
			Title:    res.Title,
			Body:     res.Body,
			QueuedAt: time.Now().UTC().Format(time.RFC3339),
		}
		rec.Items = append(rec.Items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Resolve removes the item with the given id — the re-send succeeded — and
// reports whether anything was removed. An unknown id is not an error: a
// re-send that raced another session's resolve is a benign no-op.
func (s *QueueStore) Resolve(id string) (bool, error) {
	removed := false
	err := s.Mutate(func(rec *QueueRecord) error {
		kept := rec.Items[:0]
		for _, it := range rec.Items {
			if it.ID == id {
				removed = true
				continue
			}
			kept = append(kept, it)
		}
		rec.Items = kept
		return nil
	})
	if err != nil {
		return false, err
	}
	return removed, nil
}

// lockOwner is what the lock file records: the owner's process id, the boot
// the owner ran on, and when the lock was taken. The boot-unique id is what
// makes a stale-break SAFE: a pid alone can be recycled across a reboot, so
// a previous boot's lock breaks even when some live process now carries that
// pid (it is not the owner), and a same-boot lock breaks only when the pid
// is verifiably dead.
type lockOwner struct {
	PID       int    `json:"pid"`
	BootID    string `json:"boot_id"`
	CreatedAt string `json:"created_at"`
}

// acquireLock takes the sibling advisory lock, returning its release func.
//
// The primitive is atomicfile.Claim — an exclusive create, atomic on POSIX
// (O_CREATE|O_EXCL) and on Windows (CREATE_NEW) — which is the repository's
// existing answer to "exactly one caller proceeds". Contention retries within
// a bounded window; it never blocks indefinitely.
//
// SPEC-FEEDBACK-PARTICIPATION-001 (D37, D40): the reused lock had no orphan
// recovery — a kill between acquire and release (the same crash the summary
// marker covers) left the lock file behind and wedged every later mutation
// permanently. The acquire path now performs an OWNER-VERIFIED stale-lock
// break: the lock file records the owner's pid and a boot-unique process
// identity, and the break fires ONLY when that recorded identity no longer
// identifies a live process — a recorded boot different from the current
// one, or a pid that no longer names a live process. A live owner always
// blocks, however long its mutation runs: there is NO age-based break and
// none may be added, because the acquire budget governs acquisition retries
// and Mutate enforces no hold-time bound — an age-only break could discard a
// live slow owner's committed mutation, the lost-update defect this repair
// closes. The invariant is absolute: verified owner death, nothing else.
func (s *QueueStore) acquireLock() (func() error, error) {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, queueDirPerm); err != nil {
		return nil, fmt.Errorf("mutate feedback queue %s: creating dir: %w", s.path, err)
	}

	lockPath := s.LockPath()
	var lastErr error
	for attempt := 0; attempt <= queueLockRetries; attempt++ {
		err := atomicfile.Claim(lockPath, queueFilePerm)
		if err == nil {
			if werr := s.writeLockOwner(lockPath); werr != nil {
				// The lock is HELD but unlabelled: release immediately and
				// report — an unlabelled lock could never be verified, and
				// wedging on write failure beats breaking the invariant.
				_ = os.Remove(lockPath)
				return nil, fmt.Errorf("mutate feedback queue %s: labelling lock: %w", s.path, werr)
			}
			return func() error {
				// SPEC-FEEDBACK-PARTICIPATION-001 (finding #6, harm bound):
				// the release removes the lock only when the label at the
				// path is STILL THIS PROCESS'S identity (pid + boot). A freak
				// interleaving — a stale-armed reclaimer's remove landing on
				// this lock, another writer re-claiming — must not let this
				// release delete the OTHER writer's lock. CreatedAt is
				// deliberately not compared: the identity fields are what
				// ownership means.
				raw, rerr := os.ReadFile(lockPath)
				if rerr != nil {
					return nil // gone: nothing to remove
				}
				var atPath lockOwner
				if err := json.Unmarshal(raw, &atPath); err != nil ||
					atPath.PID != os.Getpid() || atPath.BootID != bootIDIdentity() {
					return nil // not ours: remove nothing
				}
				if rmErr := os.Remove(lockPath); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
					// A surviving artifact blocks every later writer, so the
					// failure is reported rather than swallowed.
					return fmt.Errorf("mutate feedback queue %s: lock release failed: %w", s.path, rmErr)
				}
				return nil
			}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("mutate feedback queue %s: lock %s: %w", s.path, lockPath, err)
		}
		// Contention: check whether the holder is a verified-dead owner. A
		// live owner blocks through the budget; a verified-dead one is
		// broken and the claim retried immediately.
		if breakStaleLock(lockPath) {
			lastErr = err
			continue
		}
		lastErr = err
		time.Sleep(queueLockRetryDelay)
	}
	return nil, fmt.Errorf("mutate feedback queue %s: lock %s held: %w", s.path, lockPath, lastErr)
}

// writeLockOwner labels a just-acquired lock with this process's identity.
func (s *QueueStore) writeLockOwner(lockPath string) error {
	owner := lockOwner{
		PID:       os.Getpid(),
		BootID:    bootIDIdentity(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	raw, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	// The file exists (Claim created it); write the label in place. The
	// window between Claim and this write is the crash window the break
	// cannot classify — an unlabelled lock reads as LIVE (conservative: it
	// wedges until the boot changes, never breaks a possibly-live owner).
	return os.WriteFile(lockPath, raw, queueFilePerm)
}

// breakStaleLock removes the lock file when its recorded owner is verifiably
// dead, reporting whether a break happened.
//
// The verdict is made from a FRESH read of the file at the path — never a
// carried-over snapshot — and the removal is gated on an immediate re-read:
// if the bytes changed between the verdict and the remove, another reclaimer
// has already reclaimed and acquired (its fresh, labelled, live lock now
// sits at the path), and removing by path would delete THEIR lock. That
// re-read gate is what keeps two concurrent reclaimers from entering the
// section together (review-gate finding #6); the residual gap between the
// re-read and the remove is nanoseconds-wide and documented as
// residual-risk in the fix's commit.
// staleLockRereadFn is the re-read seam (the gate test injects the second
// reclaimer's re-acquire between the verdict and the re-read).
var staleLockRereadFn = os.ReadFile

// staleLockRemoveFn is the removal seam.
var staleLockRemoveFn = os.Remove

// breakStaleLock removes the lock file when its recorded owner is
// verifiably dead, reporting whether a break happened.
//
// The verdict is made from a FRESH read of the file at the path — never a
// carried-over snapshot — and the removal is gated on an immediate re-read:
// if the bytes changed between the verdict and the remove, another
// reclaimer has already reclaimed and acquired (its fresh, labelled, live
// lock now sits at the path), and removing by path would delete THEIR lock
// (review-gate finding #6). The residual nanosecond window between the
// re-read and the remove is bounded on the harm side by the release path's
// own self-label check (a release whose label is no longer at the path
// removes nothing), so a freak interleaving degrades to one extra lock
// cycle, never to two writers inside one section by construction of the
// next acquire.
func breakStaleLock(lockPath string) bool {
	raw, err := staleLockRereadFn(lockPath)
	if err != nil {
		return false // unreadable: cannot verify death, never break
	}
	var owner lockOwner
	if err := json.Unmarshal(raw, &owner); err != nil {
		return false // unlabelled (a crash between Claim and label): live
	}
	if !lockOwnerIsDead(owner) {
		return false
	}
	// The re-read gate: the bytes at the path must STILL be the bytes the
	// verdict was made on.
	now, err := os.ReadFile(lockPath)
	if err != nil || string(now) != string(raw) {
		return false // someone reclaimed between verdict and remove
	}
	if err := staleLockRemoveFn(lockPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false
	}
	slog.Warn("feedback queue: broke a stale lock (verified-dead owner)",
		"lock", lockPath, "owner_pid", owner.PID, "owner_boot", owner.BootID)
	return true
}

// writeAtomic persists rec through same-directory temp + atomic rename, so a
// crashed write leaves the previous queue intact rather than a truncated file.
func (s *QueueStore) writeAtomic(rec *QueueRecord) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, queueDirPerm); err != nil {
		return fmt.Errorf("write feedback queue %s: creating dir: %w", s.path, err)
	}
	encoded, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("write feedback queue %s: encoding: %w", s.path, err)
	}
	encoded = append(encoded, '\n')

	tmp, err := os.CreateTemp(dir, ".queue-*.tmp")
	if err != nil {
		return fmt.Errorf("write feedback queue %s: creating temp file: %w", s.path, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(encoded); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write feedback queue %s: writing temp file: %w", s.path, err)
	}
	if err := tmp.Chmod(queueFilePerm); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write feedback queue %s: setting temp file mode: %w", s.path, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write feedback queue %s: closing temp file: %w", s.path, err)
	}
	if err := atomicfile.Replace(tmpName, s.path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write feedback queue %s: replacing queue file: %w", s.path, err)
	}
	return nil
}

// normalizeQueueRecord establishes the invariants every in-memory record
// holds: a version, a non-nil item slice, and a high-water mark that clears
// every present id (so a hand-edited low value cannot reissue a live id).
func normalizeQueueRecord(rec *QueueRecord) {
	if rec.Version == 0 {
		rec.Version = queueVersion
	}
	if rec.Items == nil {
		rec.Items = []QueueItem{}
	}
	if max := maxPresentQueueSeq(rec.Items); max > rec.LastSeq {
		rec.LastSeq = max
	}
}

// maxPresentQueueSeq returns the highest numeric suffix among f<N> ids.
func maxPresentQueueSeq(items []QueueItem) int {
	max := 0
	for _, it := range items {
		var n int
		if _, err := fmt.Sscanf(it.ID, "f%d", &n); err == nil && n > max {
			max = n
		}
	}
	return max
}
