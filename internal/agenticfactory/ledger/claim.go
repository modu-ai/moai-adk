package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Actors written into transitions by the ledger itself.
const (
	actorClaim = "claim"
	actorReap  = "reap"
	actorAdd   = "add"
)

// probed is a lease a liveness probe looked at: what it saw, so the write transaction
// can tell whether the lease is still that one.
type probed struct {
	cardID    string
	epoch     int64
	heartbeat string
	alive     bool // the stale lease's holder is alive but has not renewed it
}

// probeLeases asks alive about every lease, outside any transaction (I3). dead are the
// leases whose holder is Dead; stale are the expired ones whose holder is Unknown. A
// nil alive, or a lease with no process recorded, is Unknown.
func (s *Store) probeLeases(ctx context.Context, alive func(pid int, start string) Liveness) (dead, stale []probed, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT card_id,epoch,pid,pid_start,heartbeat_at,expires_at < `+nowSQL+` FROM leases ORDER BY card_id`)
	if err != nil {
		return nil, nil, fmt.Errorf("ledger: %w", err)
	}
	type row struct {
		p       probed
		pid     int
		start   string
		expired bool
	}
	var all []row
	for rows.Next() {
		var (
			r     row
			pid   sql.NullInt64
			start sql.NullString
		)
		if err := rows.Scan(&r.p.cardID, &r.p.epoch, &pid, &start, &r.p.heartbeat, &r.expired); err != nil {
			rows.Close()
			return nil, nil, err
		}
		r.pid, r.start = int(pid.Int64), start.String
		all = append(all, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	for _, r := range all {
		verdict := LivenessUnknown
		if alive != nil && r.pid > 0 {
			verdict = alive(r.pid, r.start)
		}
		switch {
		case verdict == LivenessDead:
			dead = append(dead, r.p)
		case verdict == LivenessUnknown && r.expired:
			stale = append(stale, r.p)
		case verdict == LivenessAlive && r.expired:
			r.p.alive = true
			stale = append(stale, r.p)
		}
	}
	return dead, stale, nil
}

// reapTx reclaims the probed dead leases that are still the ones the probe saw, and
// notes the stale ones once.
func (s *Store) reapTx(ctx context.Context, tx *sql.Tx, dead, stale []probed) ([]Reclaimed, error) {
	same := func(p probed) (bool, error) {
		var epoch int64
		var hb string
		switch err := tx.QueryRowContext(ctx, `SELECT epoch,heartbeat_at FROM leases WHERE card_id=?`, p.cardID).Scan(&epoch, &hb); {
		case err == sql.ErrNoRows:
			return false, nil
		case err != nil:
			return false, fmt.Errorf("ledger: %w", err)
		}
		return epoch == p.epoch && hb == p.heartbeat, nil
	}
	var out []Reclaimed
	for _, d := range dead {
		if ok, err := same(d); err != nil {
			return nil, err
		} else if !ok {
			continue // the holder renewed, or the card was claimed again, since the probe
		}
		// A card whose reclaim fails is skipped alone: its savepoint rolls back what the
		// attempt did, a note says so once, and the other leases are still reclaimed.
		if _, err := tx.ExecContext(ctx, `SAVEPOINT reclaim`); err != nil {
			return nil, fmt.Errorf("ledger: %w", err)
		}
		r, err := s.reclaimTx(ctx, tx, d)
		if err != nil {
			var ce *cardError
			if !errors.As(err, &ce) {
				return nil, err
			}
			if _, rerr := tx.ExecContext(ctx, `ROLLBACK TO reclaim`); rerr != nil {
				return nil, fmt.Errorf("ledger: %w", rerr)
			}
			if _, rerr := tx.ExecContext(ctx, `RELEASE reclaim`); rerr != nil {
				return nil, fmt.Errorf("ledger: %w", rerr)
			}
			if nerr := noteOnce(ctx, tx, d.cardID, fmt.Sprintf("card %s epoch %d: %v; skipped", d.cardID, d.epoch, ce.err)); nerr != nil {
				return nil, nerr
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `RELEASE reclaim`); err != nil {
			return nil, fmt.Errorf("ledger: %w", err)
		}
		out = append(out, r)
	}
	for _, st := range stale {
		if ok, err := same(st); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		text := fmt.Sprintf("lease epoch %d expired; the holder's liveness is unknown, so it was not reclaimed", st.epoch)
		if st.alive {
			text = fmt.Sprintf("lease epoch %d expired; the holder is alive but has not renewed it, so it was not reclaimed", st.epoch)
		}
		if err := noteOnce(ctx, tx, st.cardID, text); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// noteOnce writes a stalled note unless this card already has one with this text. The
// text carries the epoch, so a note is once per card and epoch.
func noteOnce(ctx context.Context, q querier, cardID, text string) error {
	var n int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM notes WHERE kind='stalled' AND card_id=? AND text=?`, cardID, text).Scan(&n); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if n > 0 {
		return nil
	}
	if _, err := q.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES('stalled',?,?,`+nowSQL+`)`, cardID, text); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

// reclaimTx takes a card back from its dead holder: the lease and its keys go, the
// reclaim count rises, and the ReclaimCap-th reclaim parks the card for the operator.
func (s *Store) reclaimTx(ctx context.Context, tx *sql.Tx, p probed) (Reclaimed, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM lease_keys WHERE card_id=?`, p.cardID); err != nil {
		return Reclaimed{}, fmt.Errorf("ledger: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM leases WHERE card_id=?`, p.cardID); err != nil {
		return Reclaimed{}, fmt.Errorf("ledger: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE cards SET reclaim_count=reclaim_count+1, updated_at=`+nowSQL+` WHERE id=?`, p.cardID); err != nil {
		return Reclaimed{}, fmt.Errorf("ledger: %w", err)
	}
	c, err := cardTx(ctx, tx, p.cardID)
	if err != nil {
		return Reclaimed{}, err
	}
	if err := insertTransition(ctx, tx, p.cardID, c.Position, c.Position, p.epoch, actorReap, "reclaimed: the holder is dead"); err != nil {
		return Reclaimed{}, err
	}
	r := Reclaimed{CardID: p.cardID, Epoch: p.epoch, ReclaimCount: c.ReclaimCount}
	if c.ReclaimCount >= ReclaimCap && c.Position.Park == nil {
		to := c.Position
		to.Park = &core.Park{Reason: ParkReasonReclaimCap, Evidence: fmt.Sprintf("%d reclaims: its holders kept dying", c.ReclaimCount), Resume: "the operator decides"}
		if err := core.CheckMove(c.Position, to, []string{core.GuardReclaimCap}); err != nil {
			return Reclaimed{}, err
		}
		if err := setPosition(ctx, tx, p.cardID, to); err != nil {
			return Reclaimed{}, err
		}
		if err := insertTransition(ctx, tx, p.cardID, c.Position, to, p.epoch, actorReap, string(ParkReasonReclaimCap)); err != nil {
			return Reclaimed{}, err
		}
		r.Parked = true
	}
	return r, nil
}

// Reap takes back the leases whose holder alive judges Dead, expired or not. A holder
// judged Alive is left alone; one judged Unknown is never reclaimed either, and gets a
// stalled note once its lease has expired (I5). The ReclaimCap-th reclaim of a card
// parks it for the operator (I7). alive is required (a nil one is ErrInvalid) and is
// called before the write transaction opens, never inside it.
func (s *Store) Reap(ctx context.Context, alive func(pid int, start string) Liveness) ([]Reclaimed, error) {
	if alive == nil {
		return nil, fmt.Errorf("%w: Reap needs a liveness judge", ErrInvalid)
	}
	dead, stale, err := s.probeLeases(ctx, alive)
	if err != nil {
		return nil, err
	}
	if err := s.reapLocks(ctx, alive); err != nil {
		return nil, err
	}
	if len(dead) == 0 && len(stale) == 0 {
		return nil, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // nothing to undo after a commit
	out, err := s.reapTx(ctx, tx, dead, stale)
	if err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func insertTransition(ctx context.Context, q querier, cardID string, from, to core.Position, epoch int64, actor, reason string) error {
	_, err := q.ExecContext(ctx, `INSERT INTO transitions(card_id,from_stage,from_sub,to_stage,to_sub,epoch,actor,reason,at) VALUES(?,?,?,?,?,?,?,?,`+nowSQL+`)`,
		cardID, nullIfEmpty(string(from.Stage)), nullIfEmpty(string(from.Sub)), string(to.Stage), nullIfEmpty(string(to.Sub)), epoch, actor, nullIfEmpty(reason))
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

func setPosition(ctx context.Context, q querier, cardID string, pos core.Position) error {
	var reason, ev, res any
	if p := pos.Park; p != nil {
		reason, ev, res = string(p.Reason), p.Evidence, p.Resume
	}
	_, err := q.ExecContext(ctx, `UPDATE cards SET stage=?, sub_stage=?, integrated=?, parked=?, park_reason=?, park_evidence=?, park_resume=?, updated_at=`+nowSQL+` WHERE id=?`,
		string(pos.Stage), nullIfEmpty(string(pos.Sub)), boolInt(pos.RemoteIntegrated), boolInt(pos.Park != nil), reason, ev, res, cardID)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

// Claim is ONE write transaction after a probe outside it. It reclaims the leases whose
// holder ClaimRequest.Alive judged Dead (the probe ran first, so the callback is never
// called inside the transaction; a lease that was renewed or claimed again since is
// left alone), picks the highest-priority card that nobody holds and that is READY, or
// BUILDING or REVIEWING with its holder gone, not parked, with every dependency DONE and
// every conflict key free, inserts its lease and keys, bumps cards.epoch (the epoch
// of the latest claim, plus one) and, for a READY card, moves it to BUILDING/SPEC
// (class C, or none given) or through SPEC to BUILDING/IMPLEMENT (class A or B), each
// move recorded and checked by core.CheckMove. A card that is skipped is changed in no
// way: a conflict parks nothing. ErrNoWork when nothing qualifies.
func (s *Store) Claim(ctx context.Context, r ClaimRequest) (Claimed, error) {
	if strings.TrimSpace(r.SessionID) == "" || strings.TrimSpace(r.Worktree) == "" {
		return Claimed{}, fmt.Errorf("%w: a claim needs a session id and a worktree", ErrInvalid)
	}
	lease := r.Lease
	if lease <= 0 {
		lease = LeaseDuration
	}
	dead, stale, err := s.probeLeases(ctx, r.Alive)
	if err != nil {
		return Claimed{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Claimed{}, fmt.Errorf("ledger: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // nothing to undo after a commit
	if _, err := s.reapTx(ctx, tx, dead, stale); err != nil {
		return Claimed{}, err
	}
	card, keys, err := s.pickTx(ctx, tx)
	if err == ErrNoWork {
		// What was reclaimed stays reclaimed even when nothing is left to claim.
		if cerr := tx.Commit(); cerr != nil {
			return Claimed{}, fmt.Errorf("ledger: %w", cerr)
		}
		return Claimed{}, ErrNoWork
	}
	if err != nil {
		return Claimed{}, err
	}
	epoch := card.Epoch + 1
	if _, err := tx.ExecContext(ctx, `UPDATE cards SET epoch=? WHERE id=? AND epoch=?`, epoch, card.ID, card.Epoch); err != nil {
		return Claimed{}, fmt.Errorf("ledger: %w", err)
	}
	pos := card.Position
	if pos.Stage == core.StageReady {
		// A and B skip the spec audit as two recorded moves, so each is one table edge
		// (E01, then E23); C stops after the first.
		for _, to := range startMoves(card.Spec.Class) {
			if err := core.CheckMove(pos, to, nil); err != nil {
				return Claimed{}, err
			}
			if err := insertTransition(ctx, tx, card.ID, pos, to, epoch, actorClaim, "claimed"); err != nil {
				return Claimed{}, err
			}
			pos = to
		}
		if err := setPosition(ctx, tx, card.ID, pos); err != nil {
			return Claimed{}, err
		}
	} else if err := insertTransition(ctx, tx, card.ID, pos, pos, epoch, actorClaim, "claimed again"); err != nil {
		return Claimed{}, err
	}
	var pid, start any
	if r.PID > 0 {
		pid = r.PID
	}
	start = nullIfEmpty(r.PIDStart)
	if _, err := tx.ExecContext(ctx, `INSERT INTO leases(card_id,session_id,epoch,pid,pid_start,worktree,heartbeat_at,expires_at) VALUES(?,?,?,?,?,?,`+nowSQL+`,`+offsetSQL(lease)+`)`,
		card.ID, r.SessionID, epoch, pid, start, r.Worktree); err != nil {
		return Claimed{}, fmt.Errorf("ledger: %w", err)
	}
	for _, k := range keys {
		if _, err := tx.ExecContext(ctx, `INSERT INTO lease_keys(key,card_id) VALUES(?,?)`, k, card.ID); err != nil {
			return Claimed{}, fmt.Errorf("ledger: %w", err)
		}
	}
	var exp string
	if err := tx.QueryRowContext(ctx, `SELECT expires_at FROM leases WHERE card_id=?`, card.ID).Scan(&exp); err != nil {
		return Claimed{}, fmt.Errorf("ledger: %w", err)
	}
	expires, err := parseTime(exp)
	if err != nil {
		return Claimed{}, fmt.Errorf("ledger: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Claimed{}, fmt.Errorf("ledger: %w", err)
	}
	return Claimed{CardID: card.ID, Epoch: epoch, Position: pos, Spec: card.Spec, Keys: keys, ExpiresAt: expires}, nil
}

// startMoves are the positions a READY card passes through when claimed.
func startMoves(class string) []core.Position {
	spec := core.Position{Stage: core.StageBuilding, Sub: core.SubSpec}
	if class == "A" || class == "B" {
		return []core.Position{spec, {Stage: core.StageBuilding, Sub: core.SubImplement}}
	}
	return []core.Position{spec}
}

// pickTx is the highest-priority card nobody holds that can be claimed: READY, or
// BUILDING or REVIEWING with its holder gone; not parked; every dependency DONE; every
// conflict key free. A card that fails a test is skipped and changed in no way.
func (s *Store) pickTx(ctx context.Context, tx *sql.Tx) (Card, []string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT `+cardCols+` FROM cards c
		WHERE parked=0 AND stage IN ('READY','BUILDING','REVIEWING')
		  AND NOT EXISTS (SELECT 1 FROM leases l WHERE l.card_id=c.id)
		ORDER BY priority DESC, created_at ASC, id ASC`)
	if err != nil {
		return Card{}, nil, fmt.Errorf("ledger: %w", err)
	}
	var cands []Card
	var broken []*cardError
	for rows.Next() {
		c, err := scanCard(rows)
		var ce *cardError
		switch {
		case errors.As(err, &ce):
			broken = append(broken, ce) // this card alone; it is skipped below
		case err != nil:
			rows.Close()
			return Card{}, nil, err
		default:
			cands = append(cands, c)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Card{}, nil, err
	}
	for _, ce := range broken {
		if err := noteOnce(ctx, tx, ce.id, fmt.Sprintf("card %s epoch %d: %v; skipped", ce.id, ce.epoch, ce.err)); err != nil {
			return Card{}, nil, err
		}
	}
	for _, c := range cands {
		ok, cancelled, err := depsDone(ctx, tx, c.Spec.Depends)
		if err != nil {
			return Card{}, nil, err
		}
		if cancelled != "" {
			text := fmt.Sprintf("card %s waits on %s, which is cancelled, so it can never run; skipped", c.ID, cancelled)
			if err := noteOnce(ctx, tx, c.ID, text); err != nil {
				return Card{}, nil, err
			}
			continue
		}
		if !ok {
			continue
		}
		keys, err := KeysFor(c.Spec.Paths)
		if err != nil {
			if nerr := noteOnce(ctx, tx, c.ID, fmt.Sprintf("card %s epoch %d: %v; skipped", c.ID, c.Epoch, err)); nerr != nil {
				return Card{}, nil, nerr
			}
			continue
		}
		free, err := keysFree(ctx, tx, keys)
		if err != nil {
			return Card{}, nil, err
		}
		if free {
			return c, keys, nil
		}
	}
	return Card{}, nil, ErrNoWork
}

// depsDone says whether every dependency is DONE, and names the first one that is
// CANCELLED (a card waiting on it can never run).
func depsDone(ctx context.Context, q querier, deps []string) (done bool, cancelled string, err error) {
	done = true
	for _, d := range deps {
		var stage string
		switch err := q.QueryRowContext(ctx, `SELECT stage FROM cards WHERE id=?`, d).Scan(&stage); {
		case err == sql.ErrNoRows:
			done = false
		case err != nil:
			return false, "", fmt.Errorf("ledger: %w", err)
		case stage == string(core.StageCancelled):
			return false, d, nil
		case stage != string(core.StageDone):
			done = false
		}
	}
	return done, "", nil
}

func keysFree(ctx context.Context, q querier, keys []string) (bool, error) {
	for _, k := range keys {
		var holder string
		switch err := q.QueryRowContext(ctx, `SELECT card_id FROM lease_keys WHERE key=?`, k).Scan(&holder); {
		case err == sql.ErrNoRows:
		case err != nil:
			return false, fmt.Errorf("ledger: %w", err)
		default:
			return false, nil
		}
	}
	return true, nil
}
