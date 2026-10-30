// Package ledger is the one local authority of the local-only AgenticFactory: a
// single SQLite file per project, <moaiHome>/db/<projectKey>/af/ledger.db. Cards,
// leases, conflict keys, evidence, the merge queue, locks and notes live there and
// nowhere else; nothing in this package reads or writes a forge Issue.
//
// This file is the contract other lanes code against (design contract
// .moai/reports/issue-1948/local-adk-design-20261030.md §3): types, constants, errors,
// the Reader interface, and Store with its method set. Every Store method here is a
// stub that returns ErrNotImplemented; the implementations replace the bodies and
// change no signature. Two additions to the contract's schema, both additive:
// cards.park_evidence and cards.park_resume (a core.Park needs them and the contract
// had no column), and Spec.Depends (Claim needs the dependencies a card waits on).
//
// Rules that bind every method once implemented:
//   - I2 every write after Claim carries card id AND epoch in its WHERE clause; a stale
//     epoch changes no row and returns ErrStaleEpoch.
//   - I3 writes run in BEGIN IMMEDIATE, short, with no external command and no agent
//     turn inside; busy_timeout is BusyTimeout; the SQLite version is checked at Open.
//   - I4 time is the database clock, never the caller's.
//   - I5 a live process is never reclaimed; unknown liveness counts as alive.
//   - I6 every stage change goes through core.CheckMove.
//   - I7 rows are never deleted.
package ledger

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Defaults (contract §6): named, not inline numbers.
const (
	HeartbeatThrottle = 30 * time.Second // a session writes its heartbeat at most this often
	LeaseDuration     = 15 * time.Minute // a lease is good this long after its last heartbeat
	RenewEvery        = 5 * time.Minute  // the lane renews at this period
	ReclaimCap        = 3                // reclaims of one card before it parks for the operator
	FixLoopCap        = 3                // GREEN fix attempts before the card parks as fix-limit
	ReviewLines       = 400              // changed lines above which a card goes to REVIEWING
	LandingLockTTL    = 10 * time.Minute // merge-queue:landing
	SnapshotEvery     = 30 * time.Minute // while the store is open
	BusyTimeout       = 5 * time.Second  // SQLite busy_timeout
	MinSQLite         = "3.35"           // Open refuses an older engine
)

// ParkReasonFixLimit and ParkReasonReclaimCap are the reasons the ledger itself parks
// a card with; both are operator-facing and resolve through E20.
const (
	ParkReasonFixLimit   core.BlockReason = "fix-limit"
	ParkReasonReclaimCap core.BlockReason = "reclaim-cap"
)

// LockMerge is the name of the landing lock (contract §5).
const LockMerge = "merge-queue:landing"

var (
	ErrNotImplemented = errors.New("ledger: not implemented yet")
	ErrNoWork         = errors.New("ledger: no card can be claimed")
	ErrStaleEpoch     = errors.New("ledger: stale epoch; the card was claimed again")
	ErrKeyConflict    = errors.New("ledger: a conflict key is held by another card")
	ErrNotFound       = errors.New("ledger: no such card")
	ErrCardExists     = errors.New("ledger: the card already exists")
	ErrLockHeld       = errors.New("ledger: the lock is held")
	ErrSQLiteTooOld   = errors.New("ledger: the SQLite engine is older than " + MinSQLite)
)

// Spec is spec_json: what the card must show and where it may write.
type Spec struct {
	Acceptance []string `json:"acceptance"`        // test names or commands
	Paths      []string `json:"paths"`             // the paths the card edits; KeysFor turns them into keys
	Stop       string   `json:"stop"`              // when to stop
	Class      string   `json:"class"`             // "A" | "B" | "C"
	Risk       []string `json:"risk,omitempty"`    // "security", "permission", "data", ...
	Depends    []string `json:"depends,omitempty"` // card ids that must be DONE before this one is claimed
}

// Card is one row of cards. Position is read from stage, sub_stage, parked and the
// park columns; a card that is not parked has a nil Position.Park.
type Card struct {
	ID           string
	Title        string
	Position     core.Position
	Priority     int
	Spec         Spec
	ReclaimCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ClaimRequest names who claims. The session id, the worktree and the process are what
// a later Reap judges liveness by.
type ClaimRequest struct {
	SessionID string
	Worktree  string
	PID       int
	PIDStart  string        // the process start fingerprint
	Lease     time.Duration // 0 means LeaseDuration
	// Alive judges a holder for the lazy reap at the start of Claim. nil judges every
	// holder alive: nothing is reclaimed (I5).
	Alive func(pid int, start string) Liveness
}

// Claimed is what Claim returns: the card, its fencing epoch, the position it moved to
// and the keys it now holds. Every later write carries Epoch.
type Claimed struct {
	CardID    string
	Epoch     int64
	Position  core.Position
	Spec      Spec
	Keys      []string
	ExpiresAt time.Time
}

// Lease is one row of leases with the keys its card holds.
type Lease struct {
	CardID      string
	SessionID   string
	Epoch       int64
	PID         int
	PIDStart    string
	Worktree    string
	HeartbeatAt time.Time
	NextWakeAt  *time.Time
	ExpiresAt   time.Time
	Keys        []string
}

// Evidence kinds. Evidence is bound to a tree SHA (I8): one for another tree does not
// stand for this one.
const (
	KindRed    = "red"    // acceptance tests fail before the change
	KindGreen  = "green"  // acceptance tests pass after it
	KindGate   = "gate"   // build, vet, lint, affected tests
	KindReview = "review" // cross-model review of the diff
	KindMerge  = "merge"  // the merge result verified in a candidate worktree
)

// Evidence results.
const (
	ResultPass = "pass"
	ResultFail = "fail"
)

// Evidence is one row of evidence. The epoch and the time are the store's. Guards is
// not stored: it carries the guard keys a Move hands to core.CheckMove that the stored
// evidence cannot show, such as "resume-guard:<reason>" for a write that resolves a
// park, or core.GuardOperatorHold.
type Evidence struct {
	Kind    string
	TreeSHA string
	BaseSHA string
	Command string
	Result  string
	Path    string
	Guards  []string
}

// EvidenceRecord is a stored evidence row.
type EvidenceRecord struct {
	ID     int64
	CardID string
	Epoch  int64
	At     time.Time
	Evidence
}

// Liveness is the answer of the liveness adapter about a recorded process.
type Liveness int

const (
	LivenessUnknown Liveness = iota // treated as alive (I5)
	LivenessAlive
	LivenessDead
)

// Reclaimed is one card a Reap took back. Parked is true when this was the ReclaimCap-th.
type Reclaimed struct {
	CardID       string
	Epoch        int64 // the epoch that was reclaimed
	ReclaimCount int
	Parked       bool
}

// NoteKind is the closed set of notes.
type NoteKind string

const (
	NoteLanded   NoteKind = "landed"
	NoteDecision NoteKind = "decision"
	NoteGateWait NoteKind = "gate_wait"
	NoteStalled  NoteKind = "stalled"
)

// Note is one row of the append-only notes; Seq orders them and NotesSince reads from it.
type Note struct {
	Seq    int64
	Kind   NoteKind
	CardID string
	Text   string
	At     time.Time
}

// LockHolder identifies who holds a lock, and by which process.
type LockHolder struct {
	SessionID string
	PID       int
	PIDStart  string
}

// QueueState is the closed set of merge-queue states.
type QueueState string

const (
	QueueWaiting   QueueState = "waiting"
	QueueVerifying QueueState = "verifying"
	QueueLanding   QueueState = "landing"
	QueueLanded    QueueState = "landed"
	QueueFailed    QueueState = "failed"
)

// QueueItem is one row of merge_queue.
type QueueItem struct {
	CardID       string
	CandidateSHA string
	TrunkBefore  string
	State        QueueState
	FailReason   string
	EnqueuedAt   time.Time
	UpdatedAt    time.Time
}

// Reader is the read side the board and the flow model code against; a fake of it is
// enough for their tests.
type Reader interface {
	Cards(ctx context.Context) ([]Card, error)
	Leases(ctx context.Context) ([]Lease, error)
	Queue(ctx context.Context) ([]QueueItem, error)
	Notes(ctx context.Context, sinceSeq int64) ([]Note, error)
}

// Store is the open ledger. Its methods are safe for concurrent use and for use by
// several processes on the same file: SQLite serializes the writers.
type Store struct {
	db *sql.DB
}

var _ Reader = (*Store)(nil)

// Open opens or creates the ledger at path, applies the schema and checks the engine
// version (ErrSQLiteTooOld).
func Open(ctx context.Context, path string) (*Store, error) { return nil, ErrNotImplemented }

// Close closes the ledger.
func (s *Store) Close() error { return ErrNotImplemented }

// AddCard inserts a card in its Position (READY for a new one). ErrCardExists when the
// id is taken.
func (s *Store) AddCard(ctx context.Context, c Card) error { return ErrNotImplemented }

// Claim is one transaction: it reaps the dead holders (lazily, ClaimRequest.Alive),
// picks the highest-priority READY, unparked card whose dependencies are DONE and
// whose conflict keys are all free, inserts its lease and keys, bumps the epoch and
// moves it READY to BUILDING/SPEC (class C) or BUILDING/IMPLEMENT (A, B). ErrNoWork
// when nothing qualifies.
func (s *Store) Claim(ctx context.Context, r ClaimRequest) (Claimed, error) {
	return Claimed{}, ErrNotImplemented
}

// Heartbeat renews the lease of the card at epoch and records when the lane plans to
// wake next. ErrStaleEpoch when the epoch is not the lease's.
func (s *Store) Heartbeat(ctx context.Context, cardID string, epoch int64, nextWake *time.Time) error {
	return ErrNotImplemented
}

// Move changes the card's position. The epoch must be the lease's (ErrStaleEpoch) and
// core.CheckMove must accept the move: the facts it needs are derived from the stored
// evidence plus ev, and ev's Guards are the guard keys the stored rows cannot show.
// The transition is recorded with the actor and the reason.
func (s *Store) Move(ctx context.Context, cardID string, epoch int64, to core.Position, actor, reason string, ev ...Evidence) error {
	return ErrNotImplemented
}

// ExpandKeys adds conflict keys to the lease before an out-of-scope edit. ErrKeyConflict
// when another card holds one; nothing is added then, and nothing is parked.
func (s *Store) ExpandKeys(ctx context.Context, cardID string, epoch int64, keys []string) error {
	return ErrNotImplemented
}

// AddEvidence records evidence for the card at epoch.
func (s *Store) AddEvidence(ctx context.Context, cardID string, epoch int64, e Evidence) error {
	return ErrNotImplemented
}

// EvidenceFor lists the card's evidence, oldest first. It is on Store only, not on
// Reader.
func (s *Store) EvidenceFor(ctx context.Context, cardID string) ([]EvidenceRecord, error) {
	return nil, ErrNotImplemented
}

// Release drops the card's lease and keys; the card row stays (I7).
func (s *Store) Release(ctx context.Context, cardID string, epoch int64) error {
	return ErrNotImplemented
}

// Reap takes back the leases whose holder alive judges Dead, expired or not. A holder
// judged Alive is left alone; one judged Unknown is never reclaimed either, and gets a
// stalled note once its lease has expired (I5). The ReclaimCap-th reclaim of a card
// parks it for the operator (I7).
func (s *Store) Reap(ctx context.Context, alive func(pid int, start string) Liveness) ([]Reclaimed, error) {
	return nil, ErrNotImplemented
}

// AddNote appends a note.
func (s *Store) AddNote(ctx context.Context, n Note) error { return ErrNotImplemented }

// NotesSince lists the notes with a Seq above seq, oldest first.
func (s *Store) NotesSince(ctx context.Context, seq int64) ([]Note, error) {
	return nil, ErrNotImplemented
}

// AcquireLock takes the named lock for ttl and returns its epoch. ErrLockHeld when
// another live holder has it; a lock whose holder is Dead or whose ttl ran out is taken
// over with a higher epoch.
func (s *Store) AcquireLock(ctx context.Context, name string, holder LockHolder, ttl time.Duration) (int64, error) {
	return 0, ErrNotImplemented
}

// ReleaseLock gives the lock up; ErrStaleEpoch when the epoch is not the lock's.
func (s *Store) ReleaseLock(ctx context.Context, name string, epoch int64) error {
	return ErrNotImplemented
}

// Enqueue puts the card's candidate on the merge queue in the waiting state.
func (s *Store) Enqueue(ctx context.Context, cardID string, epoch int64, candidateSHA string) error {
	return ErrNotImplemented
}

// QueueUpdate moves a queue item to state under the landing lock; lockEpoch is the
// epoch AcquireLock returned. trunkBefore and failReason are recorded as given.
func (s *Store) QueueUpdate(ctx context.Context, cardID string, lockEpoch int64, state QueueState, trunkBefore, failReason string) error {
	return ErrNotImplemented
}

// Snapshot writes a consistent copy of the ledger to dst (VACUUM INTO a temporary file,
// then rename).
func (s *Store) Snapshot(ctx context.Context, dst string) error { return ErrNotImplemented }

// Cards, Leases, Queue and Notes are the Reader methods.
func (s *Store) Cards(ctx context.Context) ([]Card, error)      { return nil, ErrNotImplemented }
func (s *Store) Leases(ctx context.Context) ([]Lease, error)    { return nil, ErrNotImplemented }
func (s *Store) Queue(ctx context.Context) ([]QueueItem, error) { return nil, ErrNotImplemented }
func (s *Store) Notes(ctx context.Context, sinceSeq int64) ([]Note, error) {
	return nil, ErrNotImplemented
}
