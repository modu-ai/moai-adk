package homestate

import (
	"context"
	"time"
)

// EventCardHold is the event kind of a card-level hold record.
const EventCardHold = "card.hold"

// CardHold is a card-level hold record (SPEC-FACTORY-COMPLETION-RECOVERY-001
// REQ-FCR-026 (d), decision-index Q10): the card did not advance and says why.
// It is an event, not a transition — the card keeps its state and version, so
// the same lane can fix the cause and complete again.
type CardHold struct {
	CardID        string   `json:"card"`
	Reason        string   `json:"reason"`                   // check-failed, or the gate / conflict that stopped the merge
	Detail        string   `json:"detail,omitempty"`         // the refusal text or a bounded tail of the failing output
	CardHead      string   `json:"card_head"`                // the card HEAD SHA the hold was judged on
	CheckKey      string   `json:"check_key,omitempty"`      // the moai verify run tree key (check failures)
	ConflictPaths []string `json:"conflict_paths,omitempty"` // paths both sides changed
	Branch        string   `json:"branch,omitempty"`         // the branch the card was to land in
	Actor         string   `json:"actor,omitempty"`
}

// RecordCardHold appends the hold as a card.hold event under runID. It writes
// no card column and bumps no version.
func (f *FactoryDB) RecordCardHold(ctx context.Context, runID string, h CardHold, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	return retryFactoryBusy(ctx, func() error {
		tx, err := f.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if err := appendEvent(ctx, tx, runID, EventCardHold, h, now); err != nil {
			return err
		}
		return tx.Commit()
	})
}
