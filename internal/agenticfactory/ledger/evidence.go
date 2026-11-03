package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
)

func validKind(k string) bool {
	switch k {
	case KindRed, KindGreen, KindGate, KindReview, KindMerge:
		return true
	}
	return false
}

// AddEvidence records evidence for the card at its lease epoch.
func (s *Store) AddEvidence(ctx context.Context, cardID string, epoch int64, e Evidence) error {
	switch {
	case !validKind(e.Kind):
		return fmt.Errorf("%w: evidence kind %q", ErrInvalid, e.Kind)
	case e.Result != ResultPass && e.Result != ResultFail:
		return fmt.Errorf("%w: evidence result %q", ErrInvalid, e.Result)
	case strings.TrimSpace(e.SubjectSHA) == "":
		return fmt.Errorf("%w: evidence is bound to the commit SHA it measured", ErrInvalid)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		if err := leaseTx(ctx, tx, cardID, epoch); err != nil {
			return err
		}
		return insertEvidence(ctx, tx, cardID, epoch, e)
	})
}

// refuseWhileLanding closes the window between a landing's last read of the newest runs (Merger.LastCheck, asked with
// the row in landing, when the merger has one) and the swap of the trunk ref. Both sides write through this database, and
// the row is set to landing before that read, so a FAILING run of a kind the read looks at (landingReadKinds) about the
// candidate of a row that is in landing is refused: either it committed before the row went to landing and the read sees
// it, or it commits after and is refused here. Passing runs, runs about another commit and the other kinds are not what the read looks at
// and are not refused. The refusal is retryable: the row leaves landing when the landing ends (landed or waiting), or
// when `moai af reconcile` settles a row a dead lander left behind.
func refuseWhileLanding(ctx context.Context, q querier, cardID string, e Evidence) error {
	if e.Result != ResultFail || !slices.Contains(landingReadKinds, e.Kind) {
		return nil
	}
	var n int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM merge_queue WHERE card_id=? AND candidate_sha=? AND state=?`,
		cardID, e.SubjectSHA, string(QueueLanding)).Scan(&n); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if n == 0 {
		return nil
	}
	return fmt.Errorf("%w: commit %s of card %s is in landing, so a failing %s run about it cannot be recorded while the trunk is being moved on it; "+
		"record the run again once `moai af land` has returned; if no lander is running (a process that is gone left the row in landing), "+
		"`moai af reconcile` settles the row first", ErrLanding, e.SubjectSHA, cardID, e.Kind)
}

func insertEvidence(ctx context.Context, q querier, cardID string, epoch int64, e Evidence) error {
	if err := refuseWhileLanding(ctx, q, cardID, e); err != nil {
		return err
	}
	_, err := q.ExecContext(ctx, `INSERT INTO evidence(card_id,kind,tree_sha,subject_sha,base_sha,command,result,path,epoch,at) VALUES(?,?,?,?,?,?,?,?,?,`+nowSQL+`)`,
		cardID, e.Kind, e.TreeSHA, nullIfEmpty(e.SubjectSHA), nullIfEmpty(e.BaseSHA), nullIfEmpty(e.Command), e.Result, nullIfEmpty(e.Path), epoch)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

// EvidenceFor lists the card's evidence, oldest first.
func (s *Store) EvidenceFor(ctx context.Context, cardID string) ([]EvidenceRecord, error) {
	return evidenceFor(ctx, s.db, cardID)
}

func evidenceFor(ctx context.Context, q querier, cardID string) ([]EvidenceRecord, error) {
	rows, err := q.QueryContext(ctx, `SELECT id,card_id,epoch,at,kind,tree_sha,subject_sha,base_sha,command,result,path FROM evidence WHERE card_id=? ORDER BY id`, cardID)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []EvidenceRecord
	for rows.Next() {
		var (
			r                          EvidenceRecord
			at                         string
			subject, base, command, pa sql.NullString
		)
		if err := rows.Scan(&r.ID, &r.CardID, &r.Epoch, &at, &r.Kind, &r.TreeSHA, &subject, &base, &command, &r.Result, &pa); err != nil {
			return nil, err
		}
		r.SubjectSHA, r.BaseSHA, r.Command, r.Path = subject.String, base.String, command.String, pa.String
		if r.At, err = parseTime(at); err != nil {
			return nil, fmt.Errorf("ledger: evidence %d: %w", r.ID, err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
