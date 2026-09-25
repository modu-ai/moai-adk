// Package factorymsg implements the factory-only, run-scoped message broker.
// It intentionally does not read or migrate the legacy sessionmsg store.
package factorymsg

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	_ "modernc.org/sqlite"
)

const (
	SchemaVersion        = 2
	KindDispatchNotice   = "dispatch_notice"
	KindStatusRequest    = "status_request"
	KindStatusReport     = "status_report"
	KindBlocker          = "blocker"
	KindReceipt          = "receipt"
	DispositionAccepted  = "accepted"
	DispositionRejected  = "rejected"
	DispositionDuplicate = "duplicate"
	DispositionDeferred  = "deferred"
	MaxBatch             = 16
	MaxPayloadBytes      = 64 << 10
	MaxPending           = 1000
	MaxTTL               = 7 * 24 * time.Hour
	MaxDeadLetters       = 256
	BindingLaunchPending = "launch_pending"
	BindingBound         = "bound"
)

var safeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
var ErrEndpointLaunchPending = errors.New("factory endpoint is launch-pending")

// ErrStalePeer reports that a peer identity no longer matches its registered
// row (wrong generation, session, pid, or process start) or has no row at all.
// Test with errors.Is. A richer stale error defined elsewhere can join this
// class by implementing `Is(target error) bool` that returns true for
// ErrStalePeer. ErrEndpointLaunchPending is deliberately a separate class: a
// launch-pending endpoint is awaiting its first bind, not superseded.
var ErrStalePeer = errors.New("stale or unregistered peer")

// queryer is the single-row query surface shared by *sql.DB, *sql.Tx, and
// *sql.Conn, so a check can run on whichever handle the caller holds.
type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

const launchPendingSessionPrefix = "launch-pending:"

type Peer struct {
	ProjectKey, RunID, Backend, Role, Slot, SessionUUID string
	Generation                                          int64
	PID                                                 int
	ProcessStart                                        string
}
type SendRequest struct {
	From, To                                     Peer
	Kind, IdempotencyKey, TaskRef, CorrelationID string
	ExpectedTaskRevision, CurrentTaskRevision    int64
	TTL                                          time.Duration
	Payload                                      []byte
}
type Envelope struct {
	ID, ProjectKey, RunID, Kind, SenderSession, RecipientSession, TaskRef, CorrelationID string
	// Duplicate reports that Send returned an existing idempotent message.
	// It is transient and must not start a second recipient turn.
	Duplicate bool `json:"duplicate,omitempty"`
	// SenderSlot is the sender's stable lane slot, the idempotency scope.
	SenderSlot                            string
	SchemaVersion                         int
	SenderGeneration, RecipientGeneration int64
	CreatedAt, ExpiresAt                  time.Time
}
type Claim struct {
	Envelope
	ClaimToken string
}
type Status struct {
	Pending, Claimed, Acknowledged, DeadLetter int
	// Superseded counts pending or claimed messages whose recipient endpoint
	// is no longer current; they are not counted as Pending or Claimed.
	Superseded               int
	Capability, NextDelivery string
	Lanes                    []LaneStatus `json:"lanes"`
}

// LaneStatus keeps process liveness separate from unobserved model activity.
// All identity fields come from the current peers row, never the launcher UI.
type LaneStatus struct {
	Slot          string    `json:"slot"`
	Role          string    `json:"role"`
	Backend       string    `json:"backend"`
	SessionUUID   string    `json:"session_uuid"`
	Generation    int64     `json:"generation"`
	PID           int       `json:"pid"`
	ProcessStart  string    `json:"process_start"`
	UpdatedAt     time.Time `json:"updated_at"`
	ObservedAt    time.Time `json:"observed_at"`
	EndpointState string    `json:"endpoint_state"`
	BindingState  string    `json:"binding_state"`
	TaskState     string    `json:"task_state"`
}
type DeadLetter struct{ MessageID, Reason, CreatedAt string }

type Store struct {
	db                      *sql.DB
	root, runID, projectKey string
	now                     func() time.Time
	maxPending, maxDead     int
	ownerCurrent            func(int, string) bool
	recordReject            func(context.Context, string, string) error
	probeIdentity           func(int) (string, homestate.ProcessIdentityState)
	// bindStep, when set by a test, runs at each named boundary of the atomic
	// handoff rebind ("locked", then after each of its five writes); an error
	// aborts and rolls back the whole transaction.
	bindStep func(step string) error
}

func BrokerPath(projectRoot, runID string) (string, error) {
	if !safeID.MatchString(runID) {
		return "", fmt.Errorf("invalid factory run id %q", runID)
	}
	dir, err := homestate.FactoryDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "messages", runID, "broker.db"), nil
}

func projectKeyFromBrokerPath(path string) string {
	return filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path)))))
}

func Open(projectRoot, runID string) (*Store, error) {
	return OpenWithDeadline(projectRoot, runID, 5*time.Second)
}

// ValidateActiveRun rejects stale or invented run identifiers without
// initializing registry state.
func ValidateActiveRun(ctx context.Context, projectRoot, runID string) (err error) {
	if !safeID.MatchString(runID) {
		return errors.New("invalid factory run id")
	}
	path, err := homestate.FactoryDBPath(projectRoot)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		return errors.New("NO_ACTIVE_FACTORY")
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=busy_timeout(100)")
	if err != nil {
		return err
	}
	defer closeInto(&err, db, "factory state")
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status); err != nil || status != "active" {
		return errors.New("NO_ACTIVE_FACTORY")
	}
	return nil
}

// OpenExistingWithDeadline opens an already-initialized broker on the hook
// hot path without repeating WAL/schema setup on every turn boundary.
func OpenExistingWithDeadline(projectRoot, runID string, deadline time.Duration) (*Store, error) {
	path, err := BrokerPath(projectRoot, runID)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	if deadline <= 0 {
		return nil, errors.New("factory broker deadline must be positive")
	}
	busyMillis := deadline.Milliseconds() / 10
	if busyMillis < 1 {
		busyMillis = 1
	}
	v := url.Values{}
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busyMillis))
	v.Add("_txlock", "immediate")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, root: projectRoot, runID: runID, projectKey: projectKeyFromBrokerPath(path), now: time.Now, maxPending: MaxPending, maxDead: MaxDeadLetters}
	s.ownerCurrent = func(pid int, start string) bool {
		fp, state := homestate.ProbeProcessIdentity(pid)
		return state == homestate.ProcessIdentityLive && fp == start
	}
	s.recordReject = s.recordDead
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM peers`).Scan(&count); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := ensureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("upgrade factory message broker: %w", err)
	}
	return s, nil
}

// OpenWithDeadline bounds SQLite initialization and lock wait for hook paths.
func OpenWithDeadline(projectRoot, runID string, deadline time.Duration) (*Store, error) {
	path, err := BrokerPath(projectRoot, runID)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	v := url.Values{}
	if deadline <= 0 {
		return nil, errors.New("factory broker deadline must be positive")
	}
	// Leave half of the caller's budget for path setup, schema execution, and
	// cleanup. Some SQLite drivers do not interrupt a busy wait immediately
	// when the Go context expires, so using the full deadline here would make
	// the hook's end-to-end deadline untruthful.
	busyMillis := deadline.Milliseconds() / 2
	if busyMillis < 1 {
		busyMillis = 1
	}
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busyMillis))
	v.Add("_pragma", "journal_mode(WAL)")
	v.Add("_txlock", "immediate")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, root: projectRoot, runID: runID, projectKey: projectKeyFromBrokerPath(path), now: time.Now, maxPending: MaxPending, maxDead: MaxDeadLetters}
	s.ownerCurrent = func(pid int, start string) bool {
		fp, state := homestate.ProbeProcessIdentity(pid)
		return state == homestate.ProcessIdentityLive && fp == start
	}
	s.recordReject = s.recordDead
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	if _, err = db.ExecContext(ctx, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize factory message broker: %w", err)
	}
	if err = ensureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("upgrade factory message broker: %w", err)
	}
	if err = os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// ResolveActiveRun selects one active factory run or fails closed on 0/many.
func ResolveActiveRun(ctx context.Context, projectRoot, explicit string) (_ string, err error) {
	db, err := homestate.OpenFactory(projectRoot)
	if err != nil {
		return "", err
	}
	defer closeInto(&err, db, "factory state")
	if explicit != "" {
		if !safeID.MatchString(explicit) {
			return "", errors.New("invalid factory run id")
		}
		var status string
		if err := db.DB.QueryRowContext(ctx, `SELECT status FROM runs WHERE run_id=?`, explicit).Scan(&status); err != nil || status != "active" {
			return "", errors.New("NO_ACTIVE_FACTORY")
		}
		return explicit, nil
	}
	rows, err := db.DB.QueryContext(ctx, `SELECT run_id FROM runs WHERE status='active' ORDER BY run_id`)
	if err != nil {
		return "", err
	}
	defer closeInto(&err, rows, "active run rows")
	var runs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		runs = append(runs, id)
	}
	switch len(runs) {
	case 0:
		return "", errors.New("NO_ACTIVE_FACTORY")
	case 1:
		return runs[0], nil
	}
	// Reconciliation runs only on the branch that was going to fail anyway, so
	// a single-active-run join pays nothing. It only ever REMOVES provably-dead
	// owners from the active set — it never selects among survivors, so the
	// fail-closed behaviour below is preserved and ambiguity that survives it
	// is real ambiguity.
	rec, err := db.ReconcileActiveRuns(ctx, ReconcileOptionsFor(projectRoot))
	if err != nil {
		return "", err
	}
	switch len(rec.Remaining) {
	case 0:
		return "", errors.New("NO_ACTIVE_FACTORY")
	case 1:
		return rec.Remaining[0].RunID, nil
	default:
		return "", errors.New("AMBIGUOUS_FACTORY: " + describeRunOwners(rec.Remaining))
	}
}

// describeRunOwners renders the surviving runs with their owner
// classifications, so the operator reads why each survived rather than an
// opaque refusal.
func describeRunOwners(owners []homestate.RunOwner) string {
	parts := make([]string, 0, len(owners))
	for _, o := range owners {
		parts = append(parts, fmt.Sprintf("%s (owner %s)", o.RunID, o.Classification))
	}
	return strings.Join(parts, ", ")
}
func (s *Store) Close() error { return s.db.Close() }

// closeInto closes c as its caller returns and reports a close failure through
// *errp — but only when the caller is not already returning an error, so the
// first failure stays the one the caller sees, unwrapped and comparable.
func closeInto(errp *error, c io.Closer, what string) {
	if cerr := c.Close(); cerr != nil && *errp == nil {
		*errp = fmt.Errorf("close %s: %w", what, cerr)
	}
}

const schema = `
CREATE TABLE IF NOT EXISTS peers(slot TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, backend TEXT NOT NULL, role TEXT NOT NULL, session_uuid TEXT NOT NULL UNIQUE, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS messages(id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, project_key TEXT NOT NULL, run_id TEXT NOT NULL, sender_session TEXT NOT NULL, sender_generation INTEGER NOT NULL, recipient_session TEXT NOT NULL, recipient_generation INTEGER NOT NULL, kind TEXT NOT NULL, idem_key TEXT NOT NULL, task_ref TEXT NOT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL, expires_at TEXT NOT NULL, payload BLOB NOT NULL, state TEXT NOT NULL, claim_token TEXT NOT NULL DEFAULT '', claim_expires_at TEXT, disposition TEXT NOT NULL DEFAULT '', acknowledged_at TEXT, sender_slot TEXT NOT NULL DEFAULT '', UNIQUE(project_key,run_id,sender_slot,idem_key));
CREATE INDEX IF NOT EXISTS messages_recipient_state ON messages(recipient_session,recipient_generation,state,created_at);
CREATE TABLE IF NOT EXISTS dead_letters(id INTEGER PRIMARY KEY AUTOINCREMENT, message_id TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL);
` + handoffSchema + handoffBindSchema

func (p Peer) validate(projectKey, runID string) error {
	for name, v := range map[string]string{"project_key": p.ProjectKey, "run_id": p.RunID, "backend": p.Backend, "role": p.Role, "slot": p.Slot, "session_uuid": p.SessionUUID} {
		if !safeID.MatchString(v) {
			return fmt.Errorf("invalid %s", name)
		}
	}
	if strings.TrimSpace(p.ProcessStart) == "" || len(p.ProcessStart) > 512 {
		return errors.New("invalid process_start")
	}
	if p.ProjectKey != projectKey || p.RunID != runID {
		return errors.New("factory identity mismatch")
	}
	if p.Generation < 1 || p.PID < 1 {
		return errors.New("invalid generation or pid")
	}
	return nil
}

// RegisterPeer writes a lane's endpoint row for a turn hook (UserPromptSubmit)
// or, through RegisterLaunchPending, for a launcher provisional registration.
// Inside the same write transaction it reads the lane's handoff state: a turn
// registration is refused while a handoff is not final, and a replaced
// (tombstoned) session is refused for good (REQ-FLH-018); a launcher
// registration that commits finalizes an open handoff as NACK/STALE_GENERATION
// in the same commit (REQ-FLH-017).
//
// @MX:WARN: [AUTO] the handoff-state read and finalize must stay inside this write transaction, serialized by BEGIN IMMEDIATE against the handoff rebind
// @MX:REASON: REQ-FLH-017/018 — moving either before BEGIN lets a registration rotate the endpoint mid-handoff or commit a launch-pending row beside a non-final handoff (AC-FLH-019 (iv)/(v) mutants)
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func (s *Store) RegisterPeer(ctx context.Context, p Peer) (Peer, error) {
	if p.ProjectKey == "project" {
		p.ProjectKey = s.projectKey
	}
	if p.RunID != s.runID {
		return Peer{}, errors.New("factory identity mismatch")
	}
	if p.SessionUUID == "" || strings.TrimSpace(p.ProcessStart) == "" || !safeID.MatchString(p.Backend) || !safeID.MatchString(p.Role) {
		return Peer{}, errors.New("invalid peer identity")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Peer{}, err
	}
	defer func() { _ = tx.Rollback() }()
	// The bare `worker` sentinel (and its legacy spelling `agent`) takes the
	// next free `worker-<n>` slot. A number is taken when a row exists under
	// the canonical slot or either legacy spelling (`agent-<n>`, `lane-<n>`),
	// so rows an older launcher wrote into this run keep their number.
	if p.Slot == "worker" || p.Slot == "agent" {
		for n := 1; ; n++ {
			var count int
			e := tx.QueryRowContext(ctx, `SELECT count(*) FROM peers WHERE slot IN (?,?,?)`,
				fmt.Sprintf("worker-%d", n), fmt.Sprintf("agent-%d", n), fmt.Sprintf("lane-%d", n)).Scan(&count)
			if e != nil {
				return Peer{}, e
			}
			if count == 0 {
				p.Slot = fmt.Sprintf("worker-%d", n)
				break
			}
		}
	}
	if !safeID.MatchString(p.Slot) {
		return Peer{}, errors.New("invalid slot")
	}
	var oldSession, oldStart string
	var oldGen int64
	var oldPID int
	err = tx.QueryRowContext(ctx, `SELECT session_uuid,generation,pid,process_start FROM peers WHERE slot=?`, p.Slot).Scan(&oldSession, &oldGen, &oldPID, &oldStart)
	launcher := isLaunchPendingSession(p.SessionUUID)
	if !launcher {
		if refuse := s.refuseTurnRegistrationDuringHandoff(ctx, tx, p); refuse != nil {
			return Peer{}, refuse
		}
	}
	if err == nil {
		if oldSession == p.SessionUUID {
			if oldPID != p.PID || oldStart != p.ProcessStart {
				if s.ownerCurrent(oldPID, oldStart) {
					return Peer{}, errors.New("factory session UUID has a live owner")
				}
				if p.Generation <= oldGen {
					p.Generation = oldGen + 1
				}
			} else {
				p.Generation = oldGen
			}
		} else {
			if (oldPID != p.PID || oldStart != p.ProcessStart) && s.ownerCurrent(oldPID, oldStart) {
				return Peer{}, errors.New("factory logical lane has a live owner")
			}
			if p.Generation <= oldGen {
				p.Generation = oldGen + 1
			}
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Peer{}, err
	}
	if p.Generation < 1 {
		p.Generation = 1
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO peers(slot,project_key,run_id,backend,role,session_uuid,generation,pid,process_start,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?) ON CONFLICT(slot) DO UPDATE SET project_key=excluded.project_key,run_id=excluded.run_id,backend=excluded.backend,role=excluded.role,session_uuid=excluded.session_uuid,generation=excluded.generation,pid=excluded.pid,process_start=excluded.process_start,updated_at=excluded.updated_at`, p.Slot, s.projectKey, s.runID, p.Backend, p.Role, p.SessionUUID, p.Generation, p.PID, p.ProcessStart, s.now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return Peer{}, err
	}
	if launcher {
		if err := s.finalizeHandoffOnLauncherRegistration(ctx, tx, p.Slot); err != nil {
			return Peer{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Peer{}, err
	}
	p.ProjectKey = s.projectKey
	return p, nil
}

// RegisterLaunchPending records a launcher-observed process identity before a
// hook session UUID exists. The private collision-safe token is an internal
// row key only: status redacts it, and all delivery-facing lookups reject it.
func (s *Store) RegisterLaunchPending(ctx context.Context, p Peer) (Peer, error) {
	if p.SessionUUID != "" {
		return Peer{}, errors.New("launch-pending session UUID must be empty")
	}
	p.SessionUUID = launchPendingSessionPrefix + newID()
	return s.RegisterPeer(ctx, p)
}

// BindLaunchPending atomically replaces the exact launcher-owned provisional
// row. A bound row is never rotated here; authoritative turn hooks use
// RegisterPeer for that separate policy. A session UUID a handoff rebind
// tombstoned is refused with STALE_ENDPOINT on any row, inside this write
// transaction and before the row is read.
func (s *Store) BindLaunchPending(ctx context.Context, p Peer) (Peer, bool, error) {
	if p.ProjectKey == "project" {
		p.ProjectKey = s.projectKey
	}
	if isLaunchPendingSession(p.SessionUUID) {
		return Peer{}, false, errors.New("invalid bound session UUID")
	}
	if err := p.validate(s.projectKey, s.runID); err != nil {
		return Peer{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Peer{}, false, err
	}
	defer func() { _ = tx.Rollback() }()
	// t1082 REQ-FLH-010: a tombstoned session UUID is refused for good, before
	// any row is read or changed, whatever the generation.
	if tombstoned, err := sessionTombstoned(ctx, tx, p.SessionUUID); err != nil {
		return Peer{}, false, err
	} else if tombstoned {
		return Peer{}, false, staleError(ctx, tx, NackStaleEndpoint, p.Slot)
	}

	var current Peer
	err = tx.QueryRowContext(ctx, `SELECT project_key,run_id,backend,role,slot,session_uuid,generation,pid,process_start FROM peers WHERE slot=?`, p.Slot).Scan(
		&current.ProjectKey, &current.RunID, &current.Backend, &current.Role, &current.Slot,
		&current.SessionUUID, &current.Generation, &current.PID, &current.ProcessStart,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Peer{}, false, nil
	}
	if err != nil {
		return Peer{}, false, err
	}
	if !isLaunchPendingSession(current.SessionUUID) {
		return current, false, nil
	}
	if current.ProjectKey != s.projectKey || current.RunID != s.runID || current.Slot != p.Slot || current.PID != p.PID || current.ProcessStart != p.ProcessStart {
		return Peer{}, false, errors.New("launch-pending owner identity mismatch")
	}

	bound := current
	bound.SessionUUID = p.SessionUUID
	bound.Generation = current.Generation + 1
	result, err := tx.ExecContext(ctx, `UPDATE peers SET session_uuid=?,generation=?,updated_at=?
		WHERE slot=? AND project_key=? AND run_id=? AND session_uuid=?
		AND generation=? AND pid=? AND process_start=?`,
		bound.SessionUUID, bound.Generation, s.now().UTC().Format(time.RFC3339Nano),
		current.Slot, s.projectKey, s.runID, current.SessionUUID,
		current.Generation, current.PID, current.ProcessStart,
	)
	if err != nil {
		return Peer{}, false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Peer{}, false, err
	}
	if rows != 1 {
		return Peer{}, false, errors.New("launch-pending endpoint changed during bind")
	}
	if err := tx.Commit(); err != nil {
		return Peer{}, false, err
	}
	return bound, true, nil
}

// RollbackLaunchPending removes only the exact provisional endpoint returned
// by RegisterLaunchPending. A hook rebind or any newer launcher generation no
// longer matches this identity and is therefore preserved.
func (s *Store) RollbackLaunchPending(ctx context.Context, p Peer) (bool, error) {
	if p.ProjectKey == "project" {
		p.ProjectKey = s.projectKey
	}
	if p.ProjectKey != s.projectKey || p.RunID != s.runID ||
		!safeID.MatchString(p.Slot) || !isLaunchPendingSession(p.SessionUUID) ||
		p.Generation < 1 || p.PID < 1 || strings.TrimSpace(p.ProcessStart) == "" {
		return false, errors.New("invalid launch-pending rollback identity")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM peers
		WHERE slot=? AND project_key=? AND run_id=? AND session_uuid=?
		AND generation=? AND pid=? AND process_start=?`,
		p.Slot, s.projectKey, s.runID, p.SessionUUID,
		p.Generation, p.PID, p.ProcessStart)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows == 1, nil
}

func isLaunchPendingSession(sessionUUID string) bool {
	return strings.HasPrefix(sessionUUID, launchPendingSessionPrefix)
}

func (s *Store) Peer(ctx context.Context, sessionUUID string) (Peer, error) {
	if isLaunchPendingSession(sessionUUID) {
		return Peer{}, ErrEndpointLaunchPending
	}
	var p Peer
	err := s.db.QueryRowContext(ctx, `SELECT project_key,run_id,backend,role,slot,session_uuid,generation,pid,process_start FROM peers WHERE session_uuid=?`, sessionUUID).Scan(&p.ProjectKey, &p.RunID, &p.Backend, &p.Role, &p.Slot, &p.SessionUUID, &p.Generation, &p.PID, &p.ProcessStart)
	if err != nil {
		return Peer{}, err
	}
	return p, nil
}

// ResolveLane returns the sole current physical endpoint for a stable slot.
func (s *Store) ResolveLane(ctx context.Context, slot string) (Peer, error) {
	if !safeID.MatchString(slot) {
		return Peer{}, errors.New("invalid logical lane")
	}
	var p Peer
	err := s.db.QueryRowContext(ctx, `SELECT project_key,run_id,backend,role,slot,session_uuid,generation,pid,process_start FROM peers WHERE slot=?`, slot).Scan(&p.ProjectKey, &p.RunID, &p.Backend, &p.Role, &p.Slot, &p.SessionUUID, &p.Generation, &p.PID, &p.ProcessStart)
	if err != nil {
		return Peer{}, err
	}
	if isLaunchPendingSession(p.SessionUUID) {
		return Peer{}, ErrEndpointLaunchPending
	}
	return p, nil
}

func (s *Store) PeerByOwner(ctx context.Context, pid int, processStart string) (Peer, error) {
	if pid < 1 || strings.TrimSpace(processStart) == "" {
		return Peer{}, errors.New("invalid endpoint owner")
	}
	var p Peer
	err := s.db.QueryRowContext(ctx, `SELECT project_key,run_id,backend,role,slot,session_uuid,generation,pid,process_start FROM peers WHERE pid=? AND process_start=?`, pid, processStart).Scan(&p.ProjectKey, &p.RunID, &p.Backend, &p.Role, &p.Slot, &p.SessionUUID, &p.Generation, &p.PID, &p.ProcessStart)
	if err != nil {
		return Peer{}, err
	}
	if isLaunchPendingSession(p.SessionUUID) {
		return Peer{}, ErrEndpointLaunchPending
	}
	return p, nil
}

// CheckWritable performs a rollback-only reservation so hook inspection can
// report lock contention truthfully without mutating queue state.
func (s *Store) CheckWritable(ctx context.Context) (err error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer closeInto(&err, conn, "broker connection")
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `ROLLBACK`)
	return err
}

func validKind(k string) bool {
	switch k {
	case KindDispatchNotice, KindStatusRequest, KindStatusReport, KindBlocker, KindReceipt:
		return true
	}
	return false
}
func validDisposition(d string) bool {
	switch d {
	case DispositionAccepted, DispositionRejected, DispositionDuplicate, DispositionDeferred:
		return true
	}
	return false
}
func newID() string { var b [16]byte; _, _ = rand.Read(b[:]); return hex.EncodeToString(b[:]) }

// verifyPeer checks p against the registry on s.db. It must not be called
// while the caller holds an open transaction: the store has one pooled
// connection, so it would wait for that connection until ctx is done — and
// with a ctx that has no deadline or cancellation, it would wait forever. Use
// verifyPeerOn with the transaction instead.
func (s *Store) verifyPeer(ctx context.Context, p Peer) error {
	return s.verifyPeerOn(ctx, s.db, p)
}

// verifyPeerOn checks p against the registry using q, which may be s.db or a
// transaction/connection the caller already holds. A mismatch or missing row
// returns ErrStalePeer; a launch-pending session returns ErrEndpointLaunchPending.
// Any other error — a peer that fails validation, or a failed registry query
// (including ctx cancellation or deadline) — is returned unchanged and does
// not match ErrStalePeer.
//
// @MX:ANCHOR: [AUTO] stale-peer check shared by every Send/Poll/Ack entry and by in-transaction callers
// @MX:REASON: fan_in >= 7 via verifyPeer; the queryer seam lets tx-holding callers avoid the one-connection pool deadlock
func (s *Store) verifyPeerOn(ctx context.Context, q queryer, p Peer) error {
	if p.ProjectKey == "project" {
		p.ProjectKey = s.projectKey
	}
	if err := p.validate(s.projectKey, s.runID); err != nil {
		return err
	}
	if isLaunchPendingSession(p.SessionUUID) {
		return ErrEndpointLaunchPending
	}
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM peers WHERE slot=? AND session_uuid=? AND generation=? AND pid=? AND process_start=?`, p.Slot, p.SessionUUID, p.Generation, p.PID, p.ProcessStart).Scan(&n)
	if err != nil {
		return err
	}
	if n != 1 {
		return s.staleOrUnregistered(ctx, q, p)
	}
	return nil
}

func (s *Store) Send(ctx context.Context, r SendRequest) (Envelope, error) {
	reject := func(reason string, cause error) (Envelope, error) {
		if err := s.recordReject(ctx, "", reason); err != nil {
			return Envelope{}, fmt.Errorf("%w; record factory diagnostic: %v", cause, err)
		}
		return Envelope{}, cause
	}
	if err := s.verifyPeer(ctx, r.From); err != nil {
		return Envelope{}, err
	}
	if err := s.verifyPeer(ctx, r.To); err != nil {
		return Envelope{}, err
	}
	if !validKind(r.Kind) {
		return reject("poison:unknown-kind", errors.New("unknown factory message kind"))
	}
	if !safeID.MatchString(r.IdempotencyKey) {
		return reject("poison:invalid-idempotency", errors.New("invalid idempotency key"))
	}
	if !safeID.MatchString(r.TaskRef) {
		return reject("poison:invalid-task-ref", errors.New("invalid task reference"))
	}
	if !safeID.MatchString(r.CorrelationID) {
		return reject("poison:invalid-correlation", errors.New("invalid correlation id"))
	}
	if r.TTL <= 0 || r.TTL > MaxTTL {
		return reject("ttl:outside-policy", errors.New("ttl outside policy"))
	}
	if r.ExpectedTaskRevision != r.CurrentTaskRevision {
		return Envelope{}, errors.New("task revision mismatch")
	}
	if len(r.Payload) == 0 || len(r.Payload) > MaxPayloadBytes {
		return reject("poison:payload-bounds", errors.New("payload outside bounds"))
	}
	var pending int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM messages WHERE state IN ('pending','claimed')`).Scan(&pending); err != nil {
		return Envelope{}, err
	}
	if pending >= s.maxPending {
		return reject("overflow:backpressure", errors.New("factory broker backpressure"))
	}
	now := s.now().UTC()
	env := Envelope{ID: newID(), SchemaVersion: SchemaVersion, ProjectKey: s.projectKey, RunID: s.runID, Kind: r.Kind, SenderSession: r.From.SessionUUID, SenderSlot: r.From.Slot, RecipientSession: r.To.SessionUUID, SenderGeneration: r.From.Generation, RecipientGeneration: r.To.Generation, TaskRef: r.TaskRef, CorrelationID: r.CorrelationID, CreatedAt: now, ExpiresAt: now.Add(r.TTL)}
	_, err := s.db.ExecContext(ctx, `INSERT INTO messages(id,schema_version,project_key,run_id,sender_session,sender_slot,sender_generation,recipient_session,recipient_generation,kind,idem_key,task_ref,correlation_id,created_at,expires_at,payload,state) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'pending')`, env.ID, SchemaVersion, s.projectKey, s.runID, env.SenderSession, env.SenderSlot, env.SenderGeneration, env.RecipientSession, env.RecipientGeneration, env.Kind, r.IdempotencyKey, env.TaskRef, env.CorrelationID, env.CreatedAt.Format(time.RFC3339Nano), env.ExpiresAt.Format(time.RFC3339Nano), r.Payload)
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		var created, expires string
		var payload []byte
		// The idempotency scope is the sender's lane slot, not its session:
		// a retry after a sender restart returns the original message.
		err = s.db.QueryRowContext(ctx, `SELECT id,schema_version,project_key,run_id,kind,sender_session,sender_generation,recipient_session,recipient_generation,task_ref,correlation_id,created_at,expires_at,payload FROM messages WHERE project_key=? AND run_id=? AND sender_slot=? AND idem_key=?`, s.projectKey, s.runID, env.SenderSlot, r.IdempotencyKey).Scan(&env.ID, &env.SchemaVersion, &env.ProjectKey, &env.RunID, &env.Kind, &env.SenderSession, &env.SenderGeneration, &env.RecipientSession, &env.RecipientGeneration, &env.TaskRef, &env.CorrelationID, &created, &expires, &payload)
		// The recipient the request is compared against is the one the original
		// was sent to: a handoff release may since have moved the row's columns.
		origSession, origGen := env.RecipientSession, env.RecipientGeneration
		if err == nil {
			origSession, origGen = s.originalRecipient(ctx, env)
		}
		if err == nil && (env.Kind != r.Kind || origSession != r.To.SessionUUID || origGen != r.To.Generation || env.TaskRef != r.TaskRef || env.CorrelationID != r.CorrelationID || !bytes.Equal(payload, r.Payload)) {
			return Envelope{}, errors.New("idempotency key collision with different request")
		}
		env.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		env.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expires)
		env.Duplicate = err == nil
		return env, err
	}
	return env, err
}

// originalRecipient returns the recipient an existing envelope was sent to. A
// handoff release moves the envelope's recipient columns to the new generation;
// the request it was sent as still names the original one, recorded in
// lane_message_releases. With no release row it is the envelope's recipient.
func (s *Store) originalRecipient(ctx context.Context, env Envelope) (string, int64) {
	session, gen := env.RecipientSession, env.RecipientGeneration
	_ = s.db.QueryRowContext(ctx, `SELECT from_session,from_generation FROM lane_message_releases WHERE message_id=?`, env.ID).Scan(&session, &gen)
	return session, gen
}

func (s *Store) recordDead(ctx context.Context, messageID, reason string) error {
	if messageID == "" {
		messageID = newID()
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO dead_letters(message_id,reason,created_at) VALUES(?,?,?)`, messageID, reason, now); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM dead_letters WHERE id NOT IN (SELECT id FROM dead_letters ORDER BY id DESC LIMIT ?)`, s.maxDead)
	return err
}
func (s *Store) DeadLetters(ctx context.Context) (_ []DeadLetter, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT message_id,reason,created_at FROM dead_letters ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer closeInto(&err, rows, "broker rows")
	var out []DeadLetter
	for rows.Next() {
		var d DeadLetter
		if err := rows.Scan(&d.MessageID, &d.Reason, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) Claim(ctx context.Context, p Peer, limit int, lease time.Duration) (_ []Claim, err error) {
	if err := s.verifyPeer(ctx, p); err != nil {
		return nil, err
	}
	if limit < 1 || limit > MaxBatch {
		limit = MaxBatch
	}
	now := s.now().UTC()
	var eligible int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM messages WHERE recipient_session=? AND recipient_generation=? AND (state='pending' OR (state='claimed' AND (claim_expires_at<=? OR expires_at<=?)))`, p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)).Scan(&eligible); err != nil {
		return nil, err
	}
	if eligible == 0 {
		return nil, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.refuseWhileHandoffPending(ctx, tx, p.Slot); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO dead_letters(message_id,reason,created_at) SELECT id,'ttl:expired',? FROM messages WHERE recipient_session=? AND recipient_generation=? AND state IN ('pending','claimed') AND expires_at<=?`, now.Format(time.RFC3339Nano), p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM dead_letters WHERE id NOT IN (SELECT id FROM dead_letters ORDER BY id DESC LIMIT ?)`, s.maxDead); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE messages SET state='dead' WHERE recipient_session=? AND recipient_generation=? AND state IN ('pending','claimed') AND expires_at<=?`, p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,schema_version,project_key,run_id,kind,sender_session,sender_slot,sender_generation,recipient_session,recipient_generation,task_ref,correlation_id,created_at,expires_at FROM messages WHERE recipient_session=? AND recipient_generation=? AND (state='pending' OR (state='claimed' AND claim_expires_at<=?)) ORDER BY created_at LIMIT ?`, p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano), limit)
	if err != nil {
		return nil, err
	}
	defer closeInto(&err, rows, "broker rows")
	var out []Claim
	for rows.Next() {
		var c Claim
		var created, expires string
		if err := rows.Scan(&c.ID, &c.SchemaVersion, &c.ProjectKey, &c.RunID, &c.Kind, &c.SenderSession, &c.SenderSlot, &c.SenderGeneration, &c.RecipientSession, &c.RecipientGeneration, &c.TaskRef, &c.CorrelationID, &created, &expires); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		c.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expires)
		c.ClaimToken = newID()
		if _, err := tx.ExecContext(ctx, `UPDATE messages SET state='claimed',claim_token=?,claim_expires_at=?,disposition='' WHERE id=?`, c.ClaimToken, now.Add(lease).Format(time.RFC3339Nano), c.ID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) ReadBody(ctx context.Context, p Peer, id, token string) ([]byte, error) {
	if err := s.verifyPeer(ctx, p); err != nil {
		return nil, err
	}
	if err := s.refuseWhileHandoffPending(ctx, s.db, p.Slot); err != nil {
		return nil, err
	}
	var body []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM messages WHERE id=? AND recipient_session=? AND recipient_generation=? AND state='claimed' AND claim_token=? AND NOT EXISTS(SELECT 1 FROM lane_handoffs WHERE slot=? AND `+openHandoffStates+`)`, id, p.SessionUUID, p.Generation, token, p.Slot).Scan(&body)
	if err != nil {
		return nil, s.claimMismatch(ctx, p, id, token)
	}
	return body, nil
}
func (s *Store) RecordDisposition(ctx context.Context, p Peer, id, token, d string) error {
	if !validDisposition(d) {
		return errors.New("invalid disposition")
	}
	if err := s.verifyPeer(ctx, p); err != nil {
		return err
	}
	if err := s.refuseWhileHandoffPending(ctx, s.db, p.Slot); err != nil {
		return err
	}
	r, err := s.db.ExecContext(ctx, `UPDATE messages SET disposition=? WHERE id=? AND recipient_session=? AND recipient_generation=? AND state='claimed' AND claim_token=? AND NOT EXISTS(SELECT 1 FROM lane_handoffs WHERE slot=? AND `+openHandoffStates+`)`, d, id, p.SessionUUID, p.Generation, token, p.Slot)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return s.claimMismatch(ctx, p, id, token)
	}
	return nil
}
func (s *Store) Receipt(ctx context.Context, p Peer, id, token string) error {
	if err := s.verifyPeer(ctx, p); err != nil {
		return err
	}
	if err := s.refuseWhileHandoffPending(ctx, s.db, p.Slot); err != nil {
		return err
	}
	r, err := s.db.ExecContext(ctx, `UPDATE messages SET state='acknowledged',acknowledged_at=? WHERE id=? AND recipient_session=? AND recipient_generation=? AND state='claimed' AND claim_token=? AND disposition IN ('accepted','rejected','duplicate','deferred') AND NOT EXISTS(SELECT 1 FROM lane_handoffs WHERE slot=? AND `+openHandoffStates+`)`, s.now().UTC().Format(time.RFC3339Nano), id, p.SessionUUID, p.Generation, token, p.Slot)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		if err := s.claimMismatch(ctx, p, id, token); err != nil {
			if _, stale := StaleEndpoint(err); stale {
				return err
			}
			if reason, ok := HandoffNackReason(err); ok && reason == NackEndpointHandoffPending {
				return err
			}
		}
		return errors.New("receipt requires matching claim and persisted disposition")
	}
	return nil
}

// SettleReceiptControls consumes receipt control envelopes without a model
// turn. Receipts are terminal transport facts; requiring a receipt for a
// receipt would create an infinite acknowledgement chain.
func (s *Store) SettleReceiptControls(ctx context.Context, p Peer) (int, error) {
	if err := s.verifyPeer(ctx, p); err != nil {
		return 0, err
	}
	var pending int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM messages WHERE recipient_session=? AND recipient_generation=? AND kind=? AND state='pending'`, p.SessionUUID, p.Generation, KindReceipt).Scan(&pending); err != nil {
		return 0, err
	}
	if pending == 0 {
		return 0, nil
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	r, err := s.db.ExecContext(ctx, `UPDATE messages SET state='acknowledged',disposition='accepted',acknowledged_at=? WHERE recipient_session=? AND recipient_generation=? AND kind=? AND state='pending'`, now, p.SessionUUID, p.Generation, KindReceipt)
	if err != nil {
		return 0, err
	}
	n, _ := r.RowsAffected()
	return int(n), nil
}
func (s *Store) Status(ctx context.Context) (_ Status, err error) {
	st := Status{Capability: "native-claude", NextDelivery: "native-cross-session"}
	// A message is superseded when its recipient endpoint is no longer the
	// current one: Send only addresses current endpoints, so a missing match
	// means the lane re-registered after the message was queued.
	rows, err := s.db.QueryContext(ctx, `SELECT m.state, EXISTS(SELECT 1 FROM peers p WHERE p.session_uuid=m.recipient_session AND p.generation=m.recipient_generation), count(*) FROM messages m WHERE m.project_key=? AND m.run_id=? GROUP BY 1,2`, s.projectKey, s.runID)
	if err != nil {
		return st, err
	}
	defer closeInto(&err, rows, "broker rows")
	for rows.Next() {
		var state string
		var current bool
		var n int
		if err := rows.Scan(&state, &current, &n); err != nil {
			return st, err
		}
		switch {
		case (state == "pending" || state == "claimed") && !current:
			st.Superseded += n
		case state == "pending":
			st.Pending += n
		case state == "claimed":
			st.Claimed += n
		case state == "acknowledged":
			st.Acknowledged += n
		}
	}
	if err := rows.Err(); err != nil {
		return st, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM dead_letters`).Scan(&st.DeadLetter)
	if err != nil {
		return st, err
	}
	st.Lanes, err = s.laneRoster(ctx)
	for _, lane := range st.Lanes {
		if lane.Backend == "codex" {
			st.Capability = "managed-host-poll"
			st.NextDelivery = "pending-until-host-turn"
			break
		}
	}
	return st, err
}

func (s *Store) laneRoster(ctx context.Context) (_ []LaneStatus, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT slot,role,backend,session_uuid,generation,pid,process_start,updated_at FROM peers WHERE project_key=? AND run_id=? ORDER BY slot`, s.projectKey, s.runID)
	if err != nil {
		return nil, err
	}
	defer closeInto(&err, rows, "broker rows")
	lanes := make([]LaneStatus, 0)
	probe := s.probeIdentity
	if probe == nil {
		probe = homestate.ProbeProcessIdentity
	}
	for rows.Next() {
		var lane LaneStatus
		var updated string
		if err := rows.Scan(&lane.Slot, &lane.Role, &lane.Backend, &lane.SessionUUID, &lane.Generation, &lane.PID, &lane.ProcessStart, &updated); err != nil {
			return nil, err
		}
		lane.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
		if err != nil {
			return nil, fmt.Errorf("invalid peer updated_at: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fingerprint, state := probe(lane.PID)
		lane.ObservedAt = s.now().UTC()
		lane.BindingState = BindingBound
		if isLaunchPendingSession(lane.SessionUUID) {
			lane.BindingState = BindingLaunchPending
			lane.SessionUUID = ""
		}
		lane.EndpointState = "unknown"
		lane.TaskState = "unknown"
		switch state {
		case homestate.ProcessIdentityDead:
			lane.EndpointState = "dead"
		case homestate.ProcessIdentityLive:
			if fingerprint != "" && lane.ProcessStart != "" {
				lane.EndpointState = "stale"
				if fingerprint == lane.ProcessStart {
					lane.EndpointState = "live"
				}
			}
		}
		lanes = append(lanes, lane)
	}
	return lanes, rows.Err()
}
