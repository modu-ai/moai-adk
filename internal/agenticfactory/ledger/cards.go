package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// A querier is a *sql.DB or a *sql.Tx.
type querier interface {
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}

func validClass(c string) bool { return c == "" || c == "A" || c == "B" || c == "C" }

// AddCard inserts a card in its Position (READY for a new one). ErrCardExists when the
// id is taken. Every dependency in Spec.Depends must already exist and not be CANCELLED,
// and none may be the card itself (ErrDependency); a card is only ever added after what
// it depends on, so a cycle cannot form. Spec.Class must be A, B, C or empty (taken as C).
func (s *Store) AddCard(ctx context.Context, c Card) error {
	if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.Title) == "" {
		return fmt.Errorf("%w: a card needs an id and a title", ErrInvalid)
	}
	if err := c.Position.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if c.Position.Park != nil {
		if err := c.Position.Park.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}
	if !validClass(c.Spec.Class) {
		return fmt.Errorf("%w: class %q is not A, B or C", ErrInvalid, c.Spec.Class)
	}
	if _, err := KeysFor(c.Spec.Paths); err != nil {
		return err
	}
	spec, err := json.Marshal(c.Spec)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // nothing to undo after a commit
	for _, d := range c.Spec.Depends {
		if d == c.ID {
			return fmt.Errorf("%w: %s depends on itself", ErrDependency, c.ID)
		}
		var stage string
		switch err := tx.QueryRowContext(ctx, `SELECT stage FROM cards WHERE id=?`, d).Scan(&stage); {
		case err == sql.ErrNoRows:
			return fmt.Errorf("%w: %s depends on %s, which does not exist", ErrDependency, c.ID, d)
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case stage == string(core.StageCancelled):
			return fmt.Errorf("%w: %s depends on %s, which is cancelled", ErrDependency, c.ID, d)
		}
	}
	var ev, res, reason any
	if p := c.Position.Park; p != nil {
		reason, ev, res = string(p.Reason), p.Evidence, p.Resume
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO cards(id,title,stage,sub_stage,parked,park_reason,park_evidence,park_resume,priority,spec_json,reclaim_count,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,0,`+nowSQL+`,`+nowSQL+`)`,
		c.ID, c.Title, string(c.Position.Stage), nullIfEmpty(string(c.Position.Sub)), boolInt(c.Position.Park != nil), reason, ev, res, c.Priority, string(spec))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("%w: %s", ErrCardExists, c.ID)
		}
		return fmt.Errorf("ledger: %w", err)
	}
	return tx.Commit()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

const cardCols = `id,title,stage,sub_stage,parked,park_reason,park_evidence,park_resume,priority,spec_json,epoch,reclaim_count,created_at,updated_at`

func scanCard(r interface{ Scan(...any) error }) (Card, error) {
	var (
		c                       Card
		stage                   string
		sub, reason, evid, resu sql.NullString
		parked                  int
		spec, created, updated  string
	)
	if err := r.Scan(&c.ID, &c.Title, &stage, &sub, &parked, &reason, &evid, &resu, &c.Priority, &spec, &c.Epoch, &c.ReclaimCount, &created, &updated); err != nil {
		return Card{}, err
	}
	c.Position = core.Position{Stage: core.Stage(stage), Sub: core.SubStage(sub.String), RemoteIntegrated: false}
	if parked == 1 {
		c.Position.Park = &core.Park{Reason: core.BlockReason(reason.String), Evidence: evid.String, Resume: resu.String}
	}
	if err := json.Unmarshal([]byte(spec), &c.Spec); err != nil {
		return Card{}, fmt.Errorf("ledger: card %s: unreadable spec: %w", c.ID, err)
	}
	var err error
	if c.CreatedAt, err = parseTime(created); err != nil {
		return Card{}, fmt.Errorf("ledger: card %s: %w", c.ID, err)
	}
	if c.UpdatedAt, err = parseTime(updated); err != nil {
		return Card{}, fmt.Errorf("ledger: card %s: %w", c.ID, err)
	}
	return c, nil
}

// Cards lists every card, oldest first.
func (s *Store) Cards(ctx context.Context) ([]Card, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+cardCols+` FROM cards ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer rows.Close()
	var out []Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// cardTx reads one card inside a transaction.
func cardTx(ctx context.Context, q querier, id string) (Card, error) {
	c, err := scanCard(q.QueryRowContext(ctx, `SELECT `+cardCols+` FROM cards WHERE id=?`, id))
	if err == sql.ErrNoRows {
		return Card{}, fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	return c, err
}

// Leases lists every lease with the keys its card holds.
func (s *Store) Leases(ctx context.Context) ([]Lease, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT card_id,session_id,epoch,pid,pid_start,worktree,heartbeat_at,next_wake_at,expires_at FROM leases ORDER BY card_id`)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer rows.Close()
	var out []Lease
	for rows.Next() {
		var (
			l           Lease
			pid         sql.NullInt64
			start, wake sql.NullString
			hb, exp     string
		)
		if err := rows.Scan(&l.CardID, &l.SessionID, &l.Epoch, &pid, &start, &l.Worktree, &hb, &wake, &exp); err != nil {
			return nil, err
		}
		l.PID, l.PIDStart = int(pid.Int64), start.String
		if l.HeartbeatAt, err = parseTime(hb); err != nil {
			return nil, fmt.Errorf("ledger: lease %s: %w", l.CardID, err)
		}
		if l.ExpiresAt, err = parseTime(exp); err != nil {
			return nil, fmt.Errorf("ledger: lease %s: %w", l.CardID, err)
		}
		if wake.Valid {
			t, err := parseTime(wake.String)
			if err != nil {
				return nil, fmt.Errorf("ledger: lease %s: %w", l.CardID, err)
			}
			l.NextWakeAt = &t
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		keys, err := keysOf(ctx, s.db, out[i].CardID)
		if err != nil {
			return nil, err
		}
		out[i].Keys = keys
	}
	return out, nil
}

func keysOf(ctx context.Context, q querier, cardID string) ([]string, error) {
	rows, err := q.QueryContext(ctx, `SELECT key FROM lease_keys WHERE card_id=? ORDER BY key`, cardID)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
