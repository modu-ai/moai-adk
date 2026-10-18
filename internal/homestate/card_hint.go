package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// maxCardNoteBytes bounds one dispatch note: the note is a one-line hint for
// the next reader of the record, not a document.
const maxCardNoteBytes = 512

// hintChainLimit bounds the walk that refuses an order cycle. A chain this long
// is not a dispatch order anyone wrote by hand.
const hintChainLimit = 256

// AssignedHint is what RecordAssignedHint may change on an assigned card. A nil
// field is left as it is; a pointer to "" clears After, and a Note is only ever
// appended.
type AssignedHint struct {
	After *string
	Note  *string
}

// RecordAssignedHint sets the order hint (`after`) and appends a dispatch note
// on a card that is already assigned (card t1711). `assign --after` used to
// edit picked rows only, so an order decided after the assignment could not be
// recorded and a lane's bare `next` leased the wrong card first.
//
// Only an assigned row qualifies: past it the lease has taken the slot and the
// order no longer means anything. The write is version-checked like every other
// record write, never moves the state, owner or lease, and refuses an order
// that would wait on itself, directly or through the chain of recorded hints.
// The note rides the record's own event log (`card.note`) rather than a card
// column, so an older binary opening the shared database never meets a schema
// it does not know.
func (f *FactoryDB) RecordAssignedHint(ctx context.Context, runID, cardID string, hint AssignedHint, actor string, now time.Time) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	if hint.After == nil && hint.Note == nil {
		return Card{}, fmt.Errorf("%w: an order or a note is required", ErrInvalidCardInput)
	}
	if err := (CardFields{HintAfter: hint.After}).validate(); err != nil {
		return Card{}, err
	}
	var note string
	if hint.Note != nil {
		note = strings.TrimSpace(*hint.Note)
		if note == "" || len(note) > maxCardNoteBytes || strings.ContainsAny(note, "\r\n") {
			return Card{}, fmt.Errorf("%w: a dispatch note is one non-empty line of at most %d bytes", ErrInvalidCardInput, maxCardNoteBytes)
		}
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	var result Card
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		if cur.State != CardAssigned {
			return nil, fmt.Errorf("%w: card %s is %s; its order and note change only while assigned (or picked, through the picked record)", ErrIllegalTransition, cur.CardID, cur.State)
		}
		next := cur
		if hint.After != nil {
			after := strings.TrimSpace(*hint.After)
			if after != "" {
				if err := refuseHintCycle(ctx, tx, runID, cardID, after); err != nil {
					return nil, err
				}
			}
			next.HintAfter = after
		}
		if next != cur {
			next.Version = cur.Version + 1
			next.UpdatedAt = now.Format(time.RFC3339Nano)
			if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
				return nil, err
			}
			if err := appendEvent(ctx, tx, runID, "card.fields", map[string]any{"card_id": cardID, "version": next.Version, "actor": actor, "after": next.HintAfter}, now); err != nil {
				return nil, err
			}
		}
		if note != "" {
			if err := appendEvent(ctx, tx, runID, "card.note", map[string]any{"card_id": cardID, "version": next.Version, "actor": actor, "note": note}, now); err != nil {
				return nil, err
			}
		}
		result = next
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

// refuseHintCycle walks the recorded `after` chain from the proposed
// predecessor and refuses when it reaches the card itself.
func refuseHintCycle(ctx context.Context, tx *sql.Tx, runID, cardID, after string) error {
	for at, steps := after, 0; at != ""; steps++ {
		if at == cardID {
			return fmt.Errorf("%w: ordering %s after %s closes a cycle", ErrInvalidCardInput, cardID, after)
		}
		if steps >= hintChainLimit {
			return fmt.Errorf("%w: the order chain behind %s is longer than %d", ErrInvalidCardInput, after, hintChainLimit)
		}
		row, err := loadCard(ctx, tx, runID, at)
		if errors.Is(err, ErrCardNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		at = row.HintAfter
	}
	return nil
}

// CardNotes reads the dispatch notes recorded for a card, oldest first. It is
// the read half of the note `RecordAssignedHint` appends; the record's event
// log is its only home.
func (f *FactoryDB) CardNotes(ctx context.Context, runID, cardID string) ([]string, error) {
	// SQL: one constant statement; runID and cardID ride ? placeholders.
	rows, err := f.DB.QueryContext(ctx, `SELECT json_extract(payload_json,'$.note') FROM events
WHERE run_id=? AND kind='card.note' AND json_extract(payload_json,'$.card_id')=? ORDER BY seq`, runID, cardID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var notes []string
	for rows.Next() {
		var note string
		if err := rows.Scan(&note); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}
