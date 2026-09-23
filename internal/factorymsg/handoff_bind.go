package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Rebind and stale-traffic reasons (M3). STALE_* carry redirect metadata; see
// StaleEndpointError. ENDPOINT_HANDOFF_PENDING refuses body access and code
// writes on a lane whose handoff is not final yet.
const (
	NackStaleEndpoint          = "STALE_ENDPOINT"
	NackStaleGeneration        = "STALE_GENERATION"
	NackEndpointHandoffPending = "ENDPOINT_HANDOFF_PENDING"
	NackBindingEvidenceInvalid = "BINDING_EVIDENCE_INVALID"
)

// handoffBindSchema holds the three facts the atomic rebind writes besides the
// peer row, the tombstone, and the handoff state: the durable BOUND receipt,
// the per-handoff dispatch release marker, and one row per released message.
// A message row keeps its original recipient in from_*; its recipient columns
// move to the new generation.
const handoffBindSchema = `
CREATE TABLE IF NOT EXISTS lane_handoff_receipts(id TEXT PRIMARY KEY, handoff_id TEXT NOT NULL UNIQUE, slot TEXT NOT NULL, nonce TEXT NOT NULL, card_id TEXT NOT NULL, spec_id TEXT NOT NULL, old_session TEXT NOT NULL, old_generation INTEGER NOT NULL, session_uuid TEXT NOT NULL, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lane_dispatch_releases(handoff_id TEXT PRIMARY KEY, slot TEXT NOT NULL, to_session TEXT NOT NULL, to_generation INTEGER NOT NULL, message_count INTEGER NOT NULL, released_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lane_message_releases(message_id TEXT PRIMARY KEY, handoff_id TEXT NOT NULL, from_session TEXT NOT NULL, from_generation INTEGER NOT NULL, to_session TEXT NOT NULL, to_generation INTEGER NOT NULL, claim_token TEXT NOT NULL, released_at TEXT NOT NULL);
`

// EndpointRef is non-secret redirect metadata: a lane's current endpoint.
type EndpointRef struct {
	Slot, SessionUUID string
	Generation        int64
}

// StaleEndpointError refuses traffic from or to a replaced endpoint, a stale
// generation, a stale reservation, or a pre-handoff claim token. Current names
// the lane's current endpoint; it never carries a body, a token, or a secret.
type StaleEndpointError struct {
	Code    string
	Current EndpointRef
}

func (e *StaleEndpointError) Error() string {
	return fmt.Sprintf("stale or unregistered peer: %s; lane %s is current at %s generation %d", e.Code, e.Current.Slot, e.Current.SessionUUID, e.Current.Generation)
}

// StaleEndpoint reports the stale-endpoint refusal carried by err, if any.
func StaleEndpoint(err error) (*StaleEndpointError, bool) {
	var stale *StaleEndpointError
	if errors.As(err, &stale) {
		return stale, true
	}
	return nil, false
}

// HandoffBindEvidence is the mode-specific binding evidence the rebind
// validates. Interactive: the SessionStart of the user's next normal turn after
// /cd (its session UUID and owner) plus a Git readback of its cwd. Headless:
// the thread id the official app-server returned (recorded by M2) plus the
// controller's own readback of the target.
type HandoffBindEvidence struct {
	Mode, Nonce, CardID, SpecID string
	SessionUUID                 string
	PID                         int
	ProcessStart                string
	Cwd, WorktreeRoot           string
	Branch, Head                string
}

// HandoffBinding is the result of a committed rebind: the durable BOUND
// receipt, the replaced and the new endpoint, and how many dispatches were
// released to the new generation.
type HandoffBinding struct {
	HandoffID, ReceiptID, Slot string
	Old, New                   EndpointRef
	Released                   int
	BoundAt                    time.Time
}

type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func currentEndpoint(ctx context.Context, q rowQuerier, slot string) EndpointRef {
	ref := EndpointRef{Slot: slot}
	if err := q.QueryRowContext(ctx, `SELECT session_uuid,generation FROM peers WHERE slot=?`, slot).Scan(&ref.SessionUUID, &ref.Generation); err != nil {
		return EndpointRef{Slot: slot}
	}
	if isLaunchPendingSession(ref.SessionUUID) {
		ref.SessionUUID = "" // the provisional token is private
	}
	return ref
}

func staleError(ctx context.Context, q rowQuerier, code, slot string) error {
	return &StaleEndpointError{Code: code, Current: currentEndpoint(ctx, q, slot)}
}

// Named transaction steps a StepHook observes. Each fires while the calling
// write transaction holds the broker write lock.
const (
	// StepBindBegun fires in the rebind right after its write transaction began.
	StepBindBegun = "begun"
	// StepReserveInserted fires in a reservation after its RESERVED insert, before commit.
	StepReserveInserted = "reserve-inserted"
	// StepRegisterHandoffRead fires in a turn-hook peer registration after it
	// read the lane's handoff state, before it decides (REQ-FLH-018).
	StepRegisterHandoffRead = "register-handoff-read"
	// StepRegisterFinalize fires in a launcher provisional registration after
	// its slot-row write and handoff read-and-finalize, before commit (REQ-FLH-017).
	StepRegisterFinalize = "register-finalize"
)

type stepHookKey struct{}

// WithStepHook returns ctx carrying a transaction step observer. It is
// instrumentation for race tests of the handoff transactions: the hook runs at
// each named step while the write transaction holds the broker lock, and an
// error it returns aborts and rolls back that transaction. Production callers
// never set it.
func WithStepHook(ctx context.Context, hook func(step string) error) context.Context {
	return context.WithValue(ctx, stepHookKey{}, hook)
}

// HandleStats reports this broker handle's connection-pool statistics. A race
// test reads it to show a racer waits on the SQLite lock while holding its own
// connection (InUse==1) rather than queueing in a shared pool (WaitCount>0).
func (s *Store) HandleStats() sql.DBStats { return s.db.Stats() }

// SharesHandle reports whether a and b are one broker handle: the same *Store
// or the same underlying *sql.DB. Two racers sharing a handle serialize in the
// Go pool and never reach the SQLite write-lock boundary.
func SharesHandle(a, b *Store) bool { return a == b || a.db == b.db }

// ctxStep runs the context step observer, if any.
func ctxStep(ctx context.Context, name string) error {
	if hook, ok := ctx.Value(stepHookKey{}).(func(string) error); ok && hook != nil {
		return hook(name)
	}
	return nil
}

// step runs the rebind write-boundary seams: the package test seam bindStep,
// then the context observer.
func (s *Store) step(ctx context.Context, name string) error {
	if s.bindStep != nil {
		if err := s.bindStep(name); err != nil {
			return err
		}
	}
	return ctxStep(ctx, name)
}

// BindHandoff is the atomic mode-evidence rebind (REQ-FLH-008). In ONE broker
// write transaction it compares the lane row with the reserved source tuple
// (CAS), validates the evidence, and then writes the old-endpoint tombstone,
// the new peer generation, the BOUND state, the durable BOUND receipt, and the
// dispatch release. A failure at any write rolls every write back. A retried
// call on an already BOUND handoff with the same evidence returns the same
// receipt and writes nothing.
//
// Refusals: a stale nonce, a handoff no longer pending, or a lane row that no
// longer equals the reserved source tuple return STALE_GENERATION and write
// nothing. Evidence that does not match the reservation NACKs the handoff in
// the same transaction and writes nothing else.
//
// @MX:WARN: [AUTO] five writes that must commit or roll back together, serialized by BEGIN IMMEDIATE against every other writer of the slot row
// @MX:REASON: REQ-FLH-008/017 — splitting the CAS, tombstone, peer swap, BOUND, receipt, and release across transactions opens a window with two current endpoints or an early body release; M3b serializes RegisterPeer against this same transaction domain
// @MX:ANCHOR: [AUTO] single road to BOUND for interactive (hook SessionStart) and headless (cli controller) evidence
// @MX:REASON: fan_in >= 3 — hook interactive binder, cli headless binder, and the M3b race tests all enter here
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func (s *Store) BindHandoff(ctx context.Context, h Handoff, ev HandoffBindEvidence) (HandoffBinding, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return HandoffBinding{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := ctxStep(ctx, StepBindBegun); err != nil {
		return HandoffBinding{}, err
	}

	var stored Handoff
	if err := tx.QueryRowContext(ctx, `SELECT slot,card_id,spec_id,mode,nonce,state,reason,source_session,source_generation,source_pid,source_process_start,target_path,target_branch,develop_pin FROM lane_handoffs WHERE id=?`, h.ID).
		Scan(&stored.Slot, &stored.CardID, &stored.SpecID, &stored.Mode, &stored.Nonce, &stored.State, &stored.Reason,
			&stored.Source.SessionUUID, &stored.Source.Generation, &stored.Source.PID, &stored.Source.ProcessStart,
			&stored.TargetPath, &stored.TargetBranch, &stored.DevelopPin); err != nil {
		return HandoffBinding{}, fmt.Errorf("handoff %s: %w", h.ID, err)
	}
	stored.ID = h.ID
	if ev.Nonce != stored.Nonce || h.Nonce != stored.Nonce {
		return HandoffBinding{}, staleError(ctx, tx, NackStaleGeneration, stored.Slot)
	}
	if stored.State == HandoffBound {
		return boundReceipt(ctx, tx, stored, ev)
	}
	if stored.State != pendingStateFor(stored.Mode) {
		return HandoffBinding{}, staleError(ctx, tx, NackStaleGeneration, stored.Slot)
	}
	var row Peer
	if err := tx.QueryRowContext(ctx, `SELECT backend,role,session_uuid,generation,pid,process_start FROM peers WHERE slot=?`, stored.Slot).
		Scan(&row.Backend, &row.Role, &row.SessionUUID, &row.Generation, &row.PID, &row.ProcessStart); err != nil {
		return HandoffBinding{}, staleError(ctx, tx, NackStaleGeneration, stored.Slot)
	}
	if row.SessionUUID != stored.Source.SessionUUID || row.Generation != stored.Source.Generation || row.PID != stored.Source.PID || row.ProcessStart != stored.Source.ProcessStart {
		return HandoffBinding{}, staleError(ctx, tx, NackStaleGeneration, stored.Slot)
	}
	if err := s.step(ctx, "locked"); err != nil {
		return HandoffBinding{}, err
	}
	if reason, detail, err := s.validateBindEvidence(ctx, tx, stored, ev); err != nil {
		return HandoffBinding{}, err
	} else if reason != "" {
		return HandoffBinding{}, s.nackInTx(ctx, tx, stored, reason, detail)
	}

	now := s.now().UTC()
	stamp := now.Format(time.RFC3339Nano)
	b := HandoffBinding{
		HandoffID: stored.ID, ReceiptID: newID(), Slot: stored.Slot, BoundAt: now,
		Old: EndpointRef{Slot: stored.Slot, SessionUUID: row.SessionUUID, Generation: row.Generation},
		New: EndpointRef{Slot: stored.Slot, SessionUUID: ev.SessionUUID, Generation: row.Generation + 1},
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO lane_endpoint_tombstones(slot,session_uuid,generation,replaced_by_session,replaced_by_generation,handoff_id,bound_at) VALUES(?,?,?,?,?,?,?)`,
		stored.Slot, b.Old.SessionUUID, b.Old.Generation, b.New.SessionUUID, b.New.Generation, stored.ID, stamp); err != nil {
		return HandoffBinding{}, err
	}
	if err := s.step(ctx, "tombstone"); err != nil {
		return HandoffBinding{}, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE peers SET session_uuid=?,generation=?,pid=?,process_start=?,updated_at=? WHERE slot=? AND session_uuid=? AND generation=? AND pid=? AND process_start=?`,
		b.New.SessionUUID, b.New.Generation, ev.PID, ev.ProcessStart, stamp,
		stored.Slot, row.SessionUUID, row.Generation, row.PID, row.ProcessStart)
	if err != nil {
		return HandoffBinding{}, err
	}
	if n, err := res.RowsAffected(); err != nil || n != 1 {
		return HandoffBinding{}, errors.New("lane endpoint changed during rebind")
	}
	if err := s.step(ctx, "peer"); err != nil {
		return HandoffBinding{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE lane_handoffs SET state=?,reason='',updated_at=? WHERE id=? AND nonce=? AND state=?`, HandoffBound, stamp, stored.ID, stored.Nonce, stored.State); err != nil {
		return HandoffBinding{}, err
	}
	if err := insertHandoffEvent(ctx, tx, stored.ID, stored.State, HandoffBound, "", stamp); err != nil {
		return HandoffBinding{}, err
	}
	if err := s.step(ctx, "bound"); err != nil {
		return HandoffBinding{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO lane_handoff_receipts(id,handoff_id,slot,nonce,card_id,spec_id,old_session,old_generation,session_uuid,generation,pid,process_start,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		b.ReceiptID, stored.ID, stored.Slot, stored.Nonce, stored.CardID, stored.SpecID, b.Old.SessionUUID, b.Old.Generation,
		b.New.SessionUUID, b.New.Generation, ev.PID, ev.ProcessStart, stamp); err != nil {
		return HandoffBinding{}, err
	}
	if err := s.step(ctx, "receipt"); err != nil {
		return HandoffBinding{}, err
	}
	if b.Released, err = releaseDispatches(ctx, tx, stored.ID, b, stamp); err != nil {
		return HandoffBinding{}, err
	}
	if err := s.step(ctx, "release"); err != nil {
		return HandoffBinding{}, err
	}
	if err := tx.Commit(); err != nil {
		return HandoffBinding{}, err
	}
	return b, nil
}

func pendingStateFor(mode string) string {
	if mode == HandoffModeHeadless {
		return HandoffSwitchPendingHeadless
	}
	return HandoffSwitchPendingInteractive
}

// validateBindEvidence returns the NACK reason for evidence that does not bind
// this reservation, or "" when every check passes.
func (s *Store) validateBindEvidence(ctx context.Context, tx *sql.Tx, h Handoff, ev HandoffBindEvidence) (string, string, error) {
	switch {
	case ev.Mode != h.Mode:
		return NackBindingEvidenceInvalid, "mode", nil
	case ev.CardID != h.CardID || ev.SpecID != h.SpecID:
		return NackBindingEvidenceInvalid, "card_or_spec", nil
	case ev.Cwd != h.TargetPath || ev.WorktreeRoot != h.TargetPath:
		return NackTargetReadbackMismatch, "cwd", nil
	case ev.Branch != h.TargetBranch:
		return NackTargetReadbackMismatch, "branch", nil
	case ev.Head != h.DevelopPin:
		return NackTargetReadbackMismatch, "head", nil
	case !safeID.MatchString(ev.SessionUUID) || isLaunchPendingSession(ev.SessionUUID) || ev.SessionUUID == h.Source.SessionUUID:
		return NackBindingEvidenceInvalid, "session_not_new", nil
	case ev.PID < 1 || ev.ProcessStart == "" || !s.ownerCurrent(ev.PID, ev.ProcessStart):
		return NackBindingEvidenceInvalid, "owner_not_current", nil
	}
	var used int
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM peers WHERE session_uuid=?) + (SELECT count(*) FROM lane_endpoint_tombstones WHERE session_uuid=?)`, ev.SessionUUID, ev.SessionUUID).Scan(&used); err != nil {
		return "", "", err
	}
	if used > 0 {
		return NackBindingEvidenceInvalid, "session_in_use", nil
	}
	if h.Mode == HandoffModeHeadless {
		var thread string
		err := tx.QueryRowContext(ctx, `SELECT thread_id FROM lane_handoff_relocations WHERE handoff_id=? AND nonce=?`, h.ID, h.Nonce).Scan(&thread)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && thread != ev.SessionUUID) {
			return NackRelocationEvidenceInvalid, "thread_not_relocated", nil
		}
		if err != nil {
			return "", "", err
		}
	}
	return "", "", nil
}

// nackInTx terminally refuses the handoff inside the rebind transaction and
// commits only that refusal.
func (s *Store) nackInTx(ctx context.Context, tx *sql.Tx, h Handoff, reason, detail string) error {
	stamp := s.now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `UPDATE lane_handoffs SET state=?,reason=?,updated_at=? WHERE id=? AND nonce=? AND state=?`, HandoffNack, reason, stamp, h.ID, h.Nonce, h.State); err != nil {
		return err
	}
	if err := insertHandoffEvent(ctx, tx, h.ID, h.State, HandoffNack, reason, stamp); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return NewHandoffNack(reason, detail)
}

// boundReceipt answers a retried rebind of a BOUND handoff: the same evidence
// gets the durable receipt back; any other evidence is stale.
func boundReceipt(ctx context.Context, tx *sql.Tx, h Handoff, ev HandoffBindEvidence) (HandoffBinding, error) {
	b := HandoffBinding{HandoffID: h.ID, Slot: h.Slot, Old: EndpointRef{Slot: h.Slot}, New: EndpointRef{Slot: h.Slot}}
	var created string
	if err := tx.QueryRowContext(ctx, `SELECT id,old_session,old_generation,session_uuid,generation,created_at FROM lane_handoff_receipts WHERE handoff_id=?`, h.ID).
		Scan(&b.ReceiptID, &b.Old.SessionUUID, &b.Old.Generation, &b.New.SessionUUID, &b.New.Generation, &created); err != nil {
		return HandoffBinding{}, fmt.Errorf("BOUND handoff %s without receipt: %w", h.ID, err)
	}
	if b.New.SessionUUID != ev.SessionUUID {
		return HandoffBinding{}, staleError(ctx, tx, NackStaleGeneration, h.Slot)
	}
	if err := tx.QueryRowContext(ctx, `SELECT message_count FROM lane_dispatch_releases WHERE handoff_id=?`, h.ID).Scan(&b.Released); err != nil {
		return HandoffBinding{}, err
	}
	b.BoundAt, _ = time.Parse(time.RFC3339Nano, created)
	return b, nil
}

// releaseDispatches moves every undelivered message of the replaced endpoint
// to the new generation, exactly once: claims are cleared (the old token is
// kept to recognize it later as stale), the original recipient is recorded,
// and one release marker names the count.
func releaseDispatches(ctx context.Context, tx *sql.Tx, handoffID string, b HandoffBinding, stamp string) (int, error) {
	if _, err := tx.ExecContext(ctx, `INSERT INTO lane_message_releases(message_id,handoff_id,from_session,from_generation,to_session,to_generation,claim_token,released_at)
		SELECT id,?,recipient_session,recipient_generation,?,?,claim_token,? FROM messages WHERE recipient_session=? AND recipient_generation=? AND state IN ('pending','claimed')
		ON CONFLICT(message_id) DO UPDATE SET handoff_id=excluded.handoff_id,to_session=excluded.to_session,to_generation=excluded.to_generation,claim_token=excluded.claim_token,released_at=excluded.released_at`,
		handoffID, b.New.SessionUUID, b.New.Generation, stamp, b.Old.SessionUUID, b.Old.Generation); err != nil {
		return 0, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE messages SET recipient_session=?,recipient_generation=?,state='pending',claim_token='',claim_expires_at=NULL,disposition='' WHERE recipient_session=? AND recipient_generation=? AND state IN ('pending','claimed')`,
		b.New.SessionUUID, b.New.Generation, b.Old.SessionUUID, b.Old.Generation)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO lane_dispatch_releases(handoff_id,slot,to_session,to_generation,message_count,released_at) VALUES(?,?,?,?,?,?)`,
		handoffID, b.Slot, b.New.SessionUUID, b.New.Generation, n, stamp); err != nil {
		return 0, err
	}
	return int(n), nil
}

// openHandoffStates is the SQL predicate for a non-final handoff row.
const openHandoffStates = `state IN ('RESERVED','WT_READY','SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS')`

// handoffPending reports whether slot has an unfinished handoff. Body claim,
// read, disposition, ACK, and code-write authorization are refused while it
// does (REQ-FLH-009/013).
func handoffPending(ctx context.Context, q rowQuerier, slot string) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM lane_handoffs WHERE slot=? AND `+openHandoffStates, slot).Scan(&n)
	if err != nil && strings.Contains(err.Error(), "no such table") {
		// A broker created before handoffs existed, opened without schema
		// setup on the hook hot path, holds no handoff.
		return false, nil
	}
	return n > 0, err
}

func (s *Store) refuseWhileHandoffPending(ctx context.Context, q rowQuerier, slot string) error {
	pending, err := handoffPending(ctx, q, slot)
	if err != nil {
		return err
	}
	if pending {
		return NewHandoffNack(NackEndpointHandoffPending, slot)
	}
	return nil
}

// staleOrUnregistered classifies a peer identity that is not current: a
// tombstoned endpoint (STALE_ENDPOINT), a current session at an older
// generation (STALE_GENERATION), or neither.
func (s *Store) staleOrUnregistered(ctx context.Context, p Peer) error {
	var slot string
	err := s.db.QueryRowContext(ctx, `SELECT slot FROM lane_endpoint_tombstones WHERE session_uuid=? AND generation=? LIMIT 1`, p.SessionUUID, p.Generation).Scan(&slot)
	if err == nil {
		return staleError(ctx, s.db, NackStaleEndpoint, slot)
	}
	var gen int64
	if err := s.db.QueryRowContext(ctx, `SELECT slot,generation FROM peers WHERE session_uuid=?`, p.SessionUUID).Scan(&slot, &gen); err == nil && gen != p.Generation {
		return staleError(ctx, s.db, NackStaleGeneration, slot)
	}
	return errors.New("stale or unregistered peer")
}

// claimMismatch classifies a claim that matched no row: a handoff that became
// pending meanwhile, a claim token the handoff release retired (a pre-handoff
// token, STALE_GENERATION), or a plain mismatch.
func (s *Store) claimMismatch(ctx context.Context, p Peer, id, token string) error {
	if err := s.refuseWhileHandoffPending(ctx, s.db, p.Slot); err != nil {
		return err
	}
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM lane_message_releases WHERE message_id=? AND claim_token=? AND claim_token!=''`, id, token).Scan(&n); err == nil && n > 0 {
		return staleError(ctx, s.db, NackStaleGeneration, p.Slot)
	}
	return errors.New("claim identity mismatch")
}

// AuthorizeCardWrite admits code edits, commits, and card completion for a
// card only on the endpoint its BOUND handoff bound (REQ-FLH-013).
func (s *Store) AuthorizeCardWrite(ctx context.Context, p Peer, cardID string) error {
	if err := s.verifyPeer(ctx, p); err != nil {
		return err
	}
	var state, receiptSession string
	var receiptGen int64
	err := s.db.QueryRowContext(ctx, `SELECT h.state,coalesce(r.session_uuid,''),coalesce(r.generation,0) FROM lane_handoffs h LEFT JOIN lane_handoff_receipts r ON r.handoff_id=h.id WHERE h.slot=? AND h.card_id=? ORDER BY h.handoff_generation DESC LIMIT 1`, p.Slot, cardID).
		Scan(&state, &receiptSession, &receiptGen)
	if err != nil {
		return NewHandoffNack(NackEndpointHandoffPending, cardID)
	}
	if state != HandoffBound || receiptSession != p.SessionUUID || receiptGen != p.Generation {
		return NewHandoffNack(NackEndpointHandoffPending, cardID)
	}
	return nil
}

// refuseTurnRegistrationDuringHandoff is the REQ-FLH-018 check a turn-hook
// peer registration (UserPromptSubmit through RegisterPeer) makes inside its
// own write transaction, before it writes the slot row: a replaced endpoint's
// session UUID is STALE_ENDPOINT for good, and while the lane has a non-final
// handoff every registration is ENDPOINT_HANDOFF_PENDING — even one whose
// PID and process-start equal the current owner's, so the post-/cd session
// cannot move the endpoint without the rebind's tombstone and receipt.
//
// @MX:WARN: [AUTO] must read inside the caller's write transaction, after BEGIN IMMEDIATE and before the slot-row write
// @MX:REASON: REQ-FLH-018 / AC-FLH-019 (i)(iv) — a read before BEGIN misses an uncommitted RESERVED row and lets the registration rotate the endpoint mid-handoff
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func (s *Store) refuseTurnRegistrationDuringHandoff(ctx context.Context, tx *sql.Tx, p Peer) error {
	tombstoned, err := sessionTombstoned(ctx, tx, p.SessionUUID)
	if err != nil {
		return err
	}
	pending, err := handoffPending(ctx, tx, p.Slot)
	if err != nil {
		return err
	}
	if err := ctxStep(ctx, StepRegisterHandoffRead); err != nil {
		return err
	}
	if tombstoned {
		return staleError(ctx, tx, NackStaleEndpoint, p.Slot)
	}
	if pending {
		return NewHandoffNack(NackEndpointHandoffPending, p.Slot)
	}
	return nil
}

// sessionTombstoned reports whether a handoff rebind replaced this session.
func sessionTombstoned(ctx context.Context, q rowQuerier, session string) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM lane_endpoint_tombstones WHERE session_uuid=?`, session).Scan(&n)
	if err != nil && strings.Contains(err.Error(), "no such table") {
		return false, nil
	}
	return n > 0, err
}

// finalizeHandoffOnLauncherRegistration is the REQ-FLH-017 step a launcher
// provisional registration runs inside its own write transaction, after its
// slot-row write and before commit: the lane row no longer equals any open
// handoff's reserved source, so that handoff moves to NACK/STALE_GENERATION in
// the same commit. It writes no tombstone, BOUND receipt, or dispatch release.
//
// @MX:WARN: [AUTO] must run inside the launcher registration's write transaction, in the same commit as the slot-row write
// @MX:REASON: REQ-FLH-017 / AC-FLH-018, AC-FLH-019 (v) — a separate transaction leaves a committed launch-pending row beside a non-final handoff whose source tuple no longer matches
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func (s *Store) finalizeHandoffOnLauncherRegistration(ctx context.Context, tx *sql.Tx, slot string) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,state FROM lane_handoffs WHERE slot=? AND `+openHandoffStates, slot)
	if err != nil && strings.Contains(err.Error(), "no such table") {
		return ctxStep(ctx, StepRegisterFinalize)
	}
	if err != nil {
		return err
	}
	type open struct{ id, state string }
	var found []open
	for rows.Next() {
		var o open
		if err := rows.Scan(&o.id, &o.state); err != nil {
			_ = rows.Close()
			return err
		}
		found = append(found, o)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	stamp := s.now().UTC().Format(time.RFC3339Nano)
	for _, o := range found {
		if _, err := tx.ExecContext(ctx, `UPDATE lane_handoffs SET state=?,reason=?,updated_at=? WHERE id=? AND state=?`, HandoffNack, NackStaleGeneration, stamp, o.id, o.state); err != nil {
			return err
		}
		if err := insertHandoffEvent(ctx, tx, o.id, o.state, HandoffNack, NackStaleGeneration, stamp); err != nil {
			return err
		}
	}
	return ctxStep(ctx, StepRegisterFinalize)
}
