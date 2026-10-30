package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func validNoteKind(k NoteKind) bool {
	switch k {
	case NoteLanded, NoteDecision, NoteGateWait, NoteStalled:
		return true
	}
	return false
}

// AddNote appends a note.
func (s *Store) AddNote(ctx context.Context, n Note) error {
	if !validNoteKind(n.Kind) {
		return fmt.Errorf("%w: note kind %q", ErrInvalid, n.Kind)
	}
	if strings.TrimSpace(n.Text) == "" {
		return fmt.Errorf("%w: a note needs some text", ErrInvalid)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES(?,?,?,`+nowSQL+`)`, string(n.Kind), nullIfEmpty(n.CardID), n.Text)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

// NotesSince lists the notes with a Seq above seq, oldest first.
func (s *Store) NotesSince(ctx context.Context, seq int64) ([]Note, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT seq,kind,card_id,text,at FROM notes WHERE seq>? ORDER BY seq`, seq)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var (
			n    Note
			kind string
			card sql.NullString
			at   string
		)
		if err := rows.Scan(&n.Seq, &kind, &card, &n.Text, &at); err != nil {
			return nil, err
		}
		n.Kind, n.CardID = NoteKind(kind), card.String
		if n.At, err = parseTime(at); err != nil {
			return nil, fmt.Errorf("ledger: note %d: %w", n.Seq, err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// Notes is NotesSince under the name of the Reader interface.
func (s *Store) Notes(ctx context.Context, sinceSeq int64) ([]Note, error) {
	return s.NotesSince(ctx, sinceSeq)
}
