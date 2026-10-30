package ledger

import (
	"context"
	"database/sql"
	"fmt"
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
	case strings.TrimSpace(e.TreeSHA) == "":
		return fmt.Errorf("%w: evidence is bound to a tree SHA", ErrInvalid)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		if err := leaseTx(ctx, tx, cardID, epoch); err != nil {
			return err
		}
		return insertEvidence(ctx, tx, cardID, epoch, e)
	})
}

func insertEvidence(ctx context.Context, q querier, cardID string, epoch int64, e Evidence) error {
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
	defer rows.Close()
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
