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
	SchemaVersion        = 1
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
)

var safeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

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
	SchemaVersion                                                                        int
	SenderGeneration, RecipientGeneration                                                int64
	CreatedAt, ExpiresAt                                                                 time.Time
}
type Claim struct {
	Envelope
	ClaimToken string
}
type Status struct {
	Pending, Claimed, Acknowledged, DeadLetter int
	Capability, NextDelivery                   string
}
type DeadLetter struct{ MessageID, Reason, CreatedAt string }

type Store struct {
	db                      *sql.DB
	root, runID, projectKey string
	now                     func() time.Time
	maxPending, maxDead     int
	ownerCurrent            func(int, string) bool
	recordReject            func(context.Context, string, string) error
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
func ValidateActiveRun(ctx context.Context, projectRoot, runID string) error {
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
	defer db.Close()
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
	if err = os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// ResolveActiveRun selects one active factory run or fails closed on 0/many.
func ResolveActiveRun(ctx context.Context, projectRoot, explicit string) (string, error) {
	db, err := homestate.OpenFactory(projectRoot)
	if err != nil {
		return "", err
	}
	defer db.Close()
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
	defer rows.Close()
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
	default:
		return "", errors.New("AMBIGUOUS_FACTORY")
	}
}
func (s *Store) Close() error { return s.db.Close() }

const schema = `
CREATE TABLE IF NOT EXISTS peers(slot TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, backend TEXT NOT NULL, role TEXT NOT NULL, session_uuid TEXT NOT NULL UNIQUE, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS messages(id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, project_key TEXT NOT NULL, run_id TEXT NOT NULL, sender_session TEXT NOT NULL, sender_generation INTEGER NOT NULL, recipient_session TEXT NOT NULL, recipient_generation INTEGER NOT NULL, kind TEXT NOT NULL, idem_key TEXT NOT NULL, task_ref TEXT NOT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL, expires_at TEXT NOT NULL, payload BLOB NOT NULL, state TEXT NOT NULL, claim_token TEXT NOT NULL DEFAULT '', claim_expires_at TEXT, disposition TEXT NOT NULL DEFAULT '', acknowledged_at TEXT, UNIQUE(sender_session,idem_key));
CREATE INDEX IF NOT EXISTS messages_recipient_state ON messages(recipient_session,recipient_generation,state,created_at);
CREATE TABLE IF NOT EXISTS dead_letters(id INTEGER PRIMARY KEY AUTOINCREMENT, message_id TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL);
`

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
	if p.Slot == "agent" {
		for n := 1; ; n++ {
			slot := fmt.Sprintf("agent-%d", n)
			var x string
			e := tx.QueryRowContext(ctx, `SELECT session_uuid FROM peers WHERE slot=?`, slot).Scan(&x)
			if errors.Is(e, sql.ErrNoRows) {
				p.Slot = slot
				break
			}
			if e != nil {
				return Peer{}, e
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
	if err = tx.Commit(); err != nil {
		return Peer{}, err
	}
	p.ProjectKey = s.projectKey
	return p, nil
}

func (s *Store) Peer(ctx context.Context, sessionUUID string) (Peer, error) {
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
	return p, nil
}

// CheckWritable performs a rollback-only reservation so hook inspection can
// report lock contention truthfully without mutating queue state.
func (s *Store) CheckWritable(ctx context.Context) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
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
func (s *Store) verifyPeer(ctx context.Context, p Peer) error {
	if p.ProjectKey == "project" {
		p.ProjectKey = s.projectKey
	}
	if err := p.validate(s.projectKey, s.runID); err != nil {
		return err
	}
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM peers WHERE slot=? AND session_uuid=? AND generation=? AND pid=? AND process_start=?`, p.Slot, p.SessionUUID, p.Generation, p.PID, p.ProcessStart).Scan(&n)
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("stale or unregistered peer")
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
	env := Envelope{ID: newID(), SchemaVersion: SchemaVersion, ProjectKey: s.projectKey, RunID: s.runID, Kind: r.Kind, SenderSession: r.From.SessionUUID, RecipientSession: r.To.SessionUUID, SenderGeneration: r.From.Generation, RecipientGeneration: r.To.Generation, TaskRef: r.TaskRef, CorrelationID: r.CorrelationID, CreatedAt: now, ExpiresAt: now.Add(r.TTL)}
	_, err := s.db.ExecContext(ctx, `INSERT INTO messages(id,schema_version,project_key,run_id,sender_session,sender_generation,recipient_session,recipient_generation,kind,idem_key,task_ref,correlation_id,created_at,expires_at,payload,state) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'pending')`, env.ID, SchemaVersion, s.projectKey, s.runID, env.SenderSession, env.SenderGeneration, env.RecipientSession, env.RecipientGeneration, env.Kind, r.IdempotencyKey, env.TaskRef, env.CorrelationID, env.CreatedAt.Format(time.RFC3339Nano), env.ExpiresAt.Format(time.RFC3339Nano), r.Payload)
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		var created, expires string
		var payload []byte
		err = s.db.QueryRowContext(ctx, `SELECT id,schema_version,project_key,run_id,kind,sender_generation,recipient_session,recipient_generation,task_ref,correlation_id,created_at,expires_at,payload FROM messages WHERE sender_session=? AND idem_key=?`, env.SenderSession, r.IdempotencyKey).Scan(&env.ID, &env.SchemaVersion, &env.ProjectKey, &env.RunID, &env.Kind, &env.SenderGeneration, &env.RecipientSession, &env.RecipientGeneration, &env.TaskRef, &env.CorrelationID, &created, &expires, &payload)
		if err == nil && (env.Kind != r.Kind || env.RecipientSession != r.To.SessionUUID || env.RecipientGeneration != r.To.Generation || env.TaskRef != r.TaskRef || env.CorrelationID != r.CorrelationID || !bytes.Equal(payload, r.Payload)) {
			return Envelope{}, errors.New("idempotency key collision with different request")
		}
		env.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		env.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expires)
		return env, err
	}
	return env, err
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
func (s *Store) DeadLetters(ctx context.Context) ([]DeadLetter, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT message_id,reason,created_at FROM dead_letters ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func (s *Store) Claim(ctx context.Context, p Peer, limit int, lease time.Duration) ([]Claim, error) {
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
	if _, err := tx.ExecContext(ctx, `INSERT INTO dead_letters(message_id,reason,created_at) SELECT id,'ttl:expired',? FROM messages WHERE recipient_session=? AND recipient_generation=? AND state IN ('pending','claimed') AND expires_at<=?`, now.Format(time.RFC3339Nano), p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM dead_letters WHERE id NOT IN (SELECT id FROM dead_letters ORDER BY id DESC LIMIT ?)`, s.maxDead); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE messages SET state='dead' WHERE recipient_session=? AND recipient_generation=? AND state IN ('pending','claimed') AND expires_at<=?`, p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,schema_version,project_key,run_id,kind,sender_session,sender_generation,recipient_session,recipient_generation,task_ref,correlation_id,created_at,expires_at FROM messages WHERE recipient_session=? AND recipient_generation=? AND (state='pending' OR (state='claimed' AND claim_expires_at<=?)) ORDER BY created_at LIMIT ?`, p.SessionUUID, p.Generation, now.Format(time.RFC3339Nano), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Claim
	for rows.Next() {
		var c Claim
		var created, expires string
		if err := rows.Scan(&c.ID, &c.SchemaVersion, &c.ProjectKey, &c.RunID, &c.Kind, &c.SenderSession, &c.SenderGeneration, &c.RecipientSession, &c.RecipientGeneration, &c.TaskRef, &c.CorrelationID, &created, &expires); err != nil {
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
	var body []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM messages WHERE id=? AND recipient_session=? AND recipient_generation=? AND state='claimed' AND claim_token=?`, id, p.SessionUUID, p.Generation, token).Scan(&body)
	if err != nil {
		return nil, errors.New("claim identity mismatch")
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
	r, err := s.db.ExecContext(ctx, `UPDATE messages SET disposition=? WHERE id=? AND recipient_session=? AND recipient_generation=? AND state='claimed' AND claim_token=?`, d, id, p.SessionUUID, p.Generation, token)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return errors.New("claim identity mismatch")
	}
	return nil
}
func (s *Store) Receipt(ctx context.Context, p Peer, id, token string) error {
	if err := s.verifyPeer(ctx, p); err != nil {
		return err
	}
	r, err := s.db.ExecContext(ctx, `UPDATE messages SET state='acknowledged',acknowledged_at=? WHERE id=? AND recipient_session=? AND recipient_generation=? AND state='claimed' AND claim_token=? AND disposition IN ('accepted','rejected','duplicate','deferred')`, s.now().UTC().Format(time.RFC3339Nano), id, p.SessionUUID, p.Generation, token)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
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
func (s *Store) Status(ctx context.Context) (Status, error) {
	st := Status{Capability: "hook-boundary", NextDelivery: "pending-until-next-turn"}
	rows, err := s.db.QueryContext(ctx, `SELECT state,count(*) FROM messages GROUP BY state`)
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		var n int
		if err := rows.Scan(&state, &n); err != nil {
			return st, err
		}
		switch state {
		case "pending":
			st.Pending = n
		case "claimed":
			st.Claimed = n
		case "acknowledged":
			st.Acknowledged = n
		}
	}
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM dead_letters`).Scan(&st.DeadLetter)
	return st, err
}
