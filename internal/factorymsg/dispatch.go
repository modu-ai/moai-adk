package factorymsg

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Dispatch lifecycle states. Each is recorded as a distinct value; a message
// arrival or receipt never moves a dispatch past DispatchDelivered.
const (
	DispatchAssigned       = "assigned"
	DispatchDelivered      = "delivered"
	DispatchStarted        = "started"
	DispatchResultRecorded = "result_recorded"
	DispatchIntegrated     = "integrated"
	DispatchAbandoned      = "abandoned"
)

// Result-application outcomes. Every outcome other than ApplyAccepted leaves
// the dispatch record unchanged.
const (
	ApplyAccepted     = "accepted"
	ApplyUnknown      = "unknown"
	ApplyStale        = "stale"
	ApplyDuplicate    = "duplicate"
	ApplyCollision    = "collision"
	ApplyInvalidState = "invalid-state"
)

var (
	ErrDispatchNotFound  = errors.New("dispatch not found")
	ErrDispatchFenced    = errors.New("dispatch fencing token mismatch")
	ErrDispatchState     = errors.New("dispatch state does not allow this transition")
	ErrDispatchOwnerLive = errors.New("previous dispatch owner is live; explicit revoke required")
)

const dispatchSchema = `
CREATE TABLE IF NOT EXISTS dispatches(project_key TEXT NOT NULL, run_id TEXT NOT NULL, dispatch_id TEXT NOT NULL, card_id TEXT NOT NULL, lane_slot TEXT NOT NULL, attempt INTEGER NOT NULL, assignee_generation INTEGER NOT NULL, state TEXT NOT NULL, result_attempt INTEGER NOT NULL DEFAULT 0, result_digest TEXT NOT NULL DEFAULT '', result_ref TEXT NOT NULL DEFAULT '', updated_at TEXT NOT NULL, PRIMARY KEY(project_key,run_id,dispatch_id));
`

// Dispatch is one run-scoped dispatch record. AssigneeGeneration is the
// fencing token: a copy of peers.generation taken when the attempt was
// assigned or regranted.
type Dispatch struct {
	DispatchID, CardID, LaneSlot, State string
	Attempt, AssigneeGeneration         int64
	ResultAttempt                       int64
	ResultDigest, ResultRef             string
	UpdatedAt                           string
}

// ResultReport is a result as the receiver observed it: the reporter's lane
// slot and generation come from the carrying envelope, the digest from its body.
type ResultReport struct {
	DispatchID          string
	Attempt, Generation int64
	Slot, Digest, Ref   string
}

// AssignmentKey and ResultKey scope message idempotency to one dispatch
// attempt, so a reassignment never collides with the previous attempt's key.
func AssignmentKey(dispatchID string, attempt int64) string {
	return fmt.Sprintf("dispatch:%s:%d", dispatchID, attempt)
}
func ResultKey(dispatchID string, attempt int64) string {
	return fmt.Sprintf("result:%s:%d", dispatchID, attempt)
}

// ResultDigest is the digest ApplyResult compares for duplicate vs collision.
func ResultDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

const dispatchColumns = `dispatch_id,card_id,lane_slot,state,attempt,assignee_generation,result_attempt,result_digest,result_ref,updated_at`

func scanDispatch(row *sql.Row) (Dispatch, error) {
	var d Dispatch
	err := row.Scan(&d.DispatchID, &d.CardID, &d.LaneSlot, &d.State, &d.Attempt, &d.AssigneeGeneration, &d.ResultAttempt, &d.ResultDigest, &d.ResultRef, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Dispatch{}, ErrDispatchNotFound
	}
	return d, err
}

func (s *Store) loadDispatch(ctx context.Context, q queryer, dispatchID string) (Dispatch, error) {
	if !safeID.MatchString(dispatchID) {
		return Dispatch{}, errors.New("invalid dispatch id")
	}
	return scanDispatch(q.QueryRowContext(ctx, `SELECT `+dispatchColumns+` FROM dispatches WHERE project_key=? AND run_id=? AND dispatch_id=?`, s.projectKey, s.runID, dispatchID))
}

// Dispatch reads one dispatch record.
func (s *Store) Dispatch(ctx context.Context, dispatchID string) (Dispatch, error) {
	return s.loadDispatch(ctx, s.db, dispatchID)
}

// CreateDispatch records attempt 1 for a current assignee endpoint.
func (s *Store) CreateDispatch(ctx context.Context, dispatchID, cardID string, assignee Peer) (Dispatch, error) {
	if !safeID.MatchString(dispatchID) || !safeID.MatchString(cardID) {
		return Dispatch{}, errors.New("invalid dispatch or card id")
	}
	return s.dispatchTx(ctx, dispatchID, func(tx *sql.Tx, _ Dispatch, err error) (string, []any, error) {
		if err == nil {
			return "", nil, errors.New("dispatch already exists")
		}
		if !errors.Is(err, ErrDispatchNotFound) {
			return "", nil, err
		}
		if err := s.verifyPeerOn(ctx, tx, assignee); err != nil {
			return "", nil, err
		}
		return `INSERT INTO dispatches(project_key,run_id,dispatch_id,card_id,lane_slot,attempt,assignee_generation,state,updated_at) VALUES(?,?,?,?,?,1,?,?,?)`,
			[]any{s.projectKey, s.runID, dispatchID, cardID, assignee.Slot, assignee.Generation, DispatchAssigned}, nil
	})
}

// MarkDispatchDelivered and StartDispatch are the assignee's fenced,
// explicit transitions; a message receipt does not perform them.
func (s *Store) MarkDispatchDelivered(ctx context.Context, assignee Peer, dispatchID string, attempt int64) (Dispatch, error) {
	return s.assigneeTransition(ctx, assignee, dispatchID, attempt, DispatchAssigned, DispatchDelivered)
}
func (s *Store) StartDispatch(ctx context.Context, assignee Peer, dispatchID string, attempt int64) (Dispatch, error) {
	return s.assigneeTransition(ctx, assignee, dispatchID, attempt, DispatchDelivered, DispatchStarted)
}

func (s *Store) assigneeTransition(ctx context.Context, assignee Peer, dispatchID string, attempt int64, from, to string) (Dispatch, error) {
	return s.dispatchTx(ctx, dispatchID, func(tx *sql.Tx, d Dispatch, err error) (string, []any, error) {
		if err != nil {
			return "", nil, err
		}
		if err := s.verifyPeerOn(ctx, tx, assignee); err != nil {
			return "", nil, err
		}
		if attempt != d.Attempt || assignee.Slot != d.LaneSlot || assignee.Generation != d.AssigneeGeneration {
			return "", nil, ErrDispatchFenced
		}
		if d.State != from {
			return "", nil, ErrDispatchState
		}
		return `UPDATE dispatches SET state=?,updated_at=? WHERE project_key=? AND run_id=? AND dispatch_id=?`, []any{to}, nil
	})
}

// IntegrateDispatch is the lead's result_recorded -> integrated transition.
func (s *Store) IntegrateDispatch(ctx context.Context, dispatchID string) (Dispatch, error) {
	return s.leadTransition(ctx, dispatchID, DispatchIntegrated, DispatchResultRecorded)
}

// AbandonDispatch is the lead's give-up transition from any non-terminal state.
func (s *Store) AbandonDispatch(ctx context.Context, dispatchID string) (Dispatch, error) {
	return s.leadTransition(ctx, dispatchID, DispatchAbandoned, DispatchAssigned, DispatchDelivered, DispatchStarted)
}

func (s *Store) leadTransition(ctx context.Context, dispatchID, to string, from ...string) (Dispatch, error) {
	return s.dispatchTx(ctx, dispatchID, func(_ *sql.Tx, d Dispatch, err error) (string, []any, error) {
		if err != nil {
			return "", nil, err
		}
		if !oneOf(d.State, from...) {
			return "", nil, ErrDispatchState
		}
		return `UPDATE dispatches SET state=?,updated_at=? WHERE project_key=? AND run_id=? AND dispatch_id=?`, []any{to}, nil
	})
}

// ReassignDispatch starts a new attempt on another (or the same) lane. The
// previous owner must be confirmed not live, or the lead must revoke it
// explicitly. Results reported for the previous attempt are stale afterwards.
func (s *Store) ReassignDispatch(ctx context.Context, dispatchID string, attempt int64, to Peer, revoke bool) (Dispatch, error) {
	return s.dispatchTx(ctx, dispatchID, func(tx *sql.Tx, d Dispatch, err error) (string, []any, error) {
		if err != nil {
			return "", nil, err
		}
		if !oneOf(d.State, DispatchAssigned, DispatchDelivered, DispatchStarted) {
			return "", nil, ErrDispatchState
		}
		if attempt != d.Attempt {
			return "", nil, ErrDispatchFenced
		}
		if err := s.verifyPeerOn(ctx, tx, to); err != nil {
			return "", nil, err
		}
		if !revoke {
			var pid int
			var start string
			e := tx.QueryRowContext(ctx, `SELECT pid,process_start FROM peers WHERE slot=?`, d.LaneSlot).Scan(&pid, &start)
			if e == nil && s.ownerCurrent(pid, start) {
				return "", nil, ErrDispatchOwnerLive
			}
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return "", nil, e
			}
		}
		return `UPDATE dispatches SET state=?,attempt=attempt+1,lane_slot=?,assignee_generation=?,updated_at=? WHERE project_key=? AND run_id=? AND dispatch_id=?`,
			[]any{DispatchAssigned, to.Slot, to.Generation}, nil
	})
}

// RegrantDispatch moves authority for the SAME attempt to a newer generation
// of the same lane, in one transaction with the caller's own state change.
// The admission conditions belong to the calling SPEC; change may be nil.
func (s *Store) RegrantDispatch(ctx context.Context, dispatchID string, attempt int64, to Peer, change func(*sql.Tx) error) (Dispatch, error) {
	return s.dispatchTx(ctx, dispatchID, func(tx *sql.Tx, d Dispatch, err error) (string, []any, error) {
		if err != nil {
			return "", nil, err
		}
		if !oneOf(d.State, DispatchAssigned, DispatchDelivered, DispatchStarted) {
			return "", nil, ErrDispatchState
		}
		if attempt != d.Attempt || to.Slot != d.LaneSlot || to.Generation <= d.AssigneeGeneration {
			return "", nil, ErrDispatchFenced
		}
		if err := s.verifyPeerOn(ctx, tx, to); err != nil {
			return "", nil, err
		}
		if change != nil {
			if err := change(tx); err != nil {
				return "", nil, err
			}
		}
		return `UPDATE dispatches SET assignee_generation=?,updated_at=? WHERE project_key=? AND run_id=? AND dispatch_id=?`, []any{to.Generation}, nil
	})
}

// ApplyResult decides a reported result in one transaction, in the fixed
// order of the result-application table. Only ApplyAccepted writes.
//
// @MX:ANCHOR: [AUTO] exactly-once result application; the check order is the contract
// @MX:REASON: reordering steps 4/5/6 changes which retries are duplicate vs stale (AC-DHR-014 mutants)
func (s *Store) ApplyResult(ctx context.Context, r ResultReport) (string, error) {
	if !safeID.MatchString(r.DispatchID) || !safeID.MatchString(r.Slot) {
		return "", errors.New("invalid dispatch id or reporter slot")
	}
	if r.Digest == "" || strings.TrimSpace(r.Ref) == "" || len(r.Ref) > 512 {
		return "", errors.New("invalid result digest or reference")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	d, err := s.loadDispatch(ctx, tx, r.DispatchID)
	if errors.Is(err, ErrDispatchNotFound) {
		return ApplyUnknown, nil // (1)
	}
	if err != nil {
		return "", err
	}
	if r.Attempt != d.Attempt || r.Slot != d.LaneSlot { // (2), (3)
		return ApplyStale, nil
	}
	var laneGeneration int64
	err = tx.QueryRowContext(ctx, `SELECT generation FROM peers WHERE slot=?`, d.LaneSlot).Scan(&laneGeneration)
	if errors.Is(err, sql.ErrNoRows) {
		return ApplyStale, nil // no current lane endpoint to be fresh against
	}
	if err != nil {
		return "", err
	}
	if r.Generation < laneGeneration { // (4)
		return ApplyStale, nil
	}
	if d.ResultAttempt == d.Attempt { // (5)
		if d.ResultDigest == r.Digest {
			return ApplyDuplicate, nil
		}
		return ApplyCollision, nil
	}
	if r.Generation != d.AssigneeGeneration { // (6)
		return ApplyStale, nil
	}
	if d.State != DispatchStarted { // (7)
		return ApplyInvalidState, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE dispatches SET state=?,result_attempt=?,result_digest=?,result_ref=?,updated_at=? WHERE project_key=? AND run_id=? AND dispatch_id=?`,
		DispatchResultRecorded, d.Attempt, r.Digest, r.Ref, s.now().UTC().Format(time.RFC3339Nano), s.projectKey, s.runID, d.DispatchID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil { // (8)
		return "", err
	}
	return ApplyAccepted, nil
}

// dispatchTx loads the record inside one immediate transaction, lets decide
// return the statement to run, then re-reads the record. For UPDATE
// statements the caller supplies only the SET values; updated_at and the key
// columns are appended here.
func (s *Store) dispatchTx(ctx context.Context, dispatchID string, decide func(*sql.Tx, Dispatch, error) (string, []any, error)) (Dispatch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Dispatch{}, err
	}
	defer func() { _ = tx.Rollback() }()
	d, loadErr := s.loadDispatch(ctx, tx, dispatchID)
	stmt, args, err := decide(tx, d, loadErr)
	if err != nil {
		return Dispatch{}, err
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	if strings.HasPrefix(stmt, "INSERT") {
		args = append(args, now)
	} else {
		args = append(args, now, s.projectKey, s.runID, dispatchID)
	}
	if _, err := tx.ExecContext(ctx, stmt, args...); err != nil {
		return Dispatch{}, err
	}
	d, err = s.loadDispatch(ctx, tx, dispatchID)
	if err != nil {
		return Dispatch{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dispatch{}, err
	}
	return d, nil
}

func oneOf(v string, set ...string) bool {
	for _, x := range set {
		if v == x {
			return true
		}
	}
	return false
}
