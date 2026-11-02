package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

const resumeGuardPrefix = "resume-guard:"

// checkGuards lets a Guards entry be a guard key and nothing else: a resume guard, or the
// guard of one of the named in-place parks. A fact is the ledger's to show, never the
// caller's to claim.
func checkGuards(ev []Evidence) error {
	for _, e := range ev {
		for _, g := range e.Guards {
			switch {
			case strings.HasPrefix(g, resumeGuardPrefix) && len(g) > len(resumeGuardPrefix):
			case g == core.GuardOperatorHold || g == core.GuardFixLimit || g == core.GuardReclaimCap:
			default:
				return fmt.Errorf("%w: %q is not a guard key (facts come from the ledger's tables)", ErrInvalid, g)
			}
		}
	}
	return nil
}

func checkEvidence(e Evidence) error {
	switch {
	case !validKind(e.Kind):
		return fmt.Errorf("%w: evidence kind %q", ErrInvalid, e.Kind)
	case e.Result != ResultPass && e.Result != ResultFail:
		return fmt.Errorf("%w: evidence result %q", ErrInvalid, e.Result)
	case strings.TrimSpace(e.SubjectSHA) == "":
		return fmt.Errorf("%w: evidence is bound to the commit SHA it measured", ErrInvalid)
	}
	return nil
}

// Move changes the card's position; see the API for the epoch, the facts and the guards.
// A move without a lease is authorised by the card's public epoch number alone, which is
// not a secret: it is allowed only for the in-place resume of a ledger-made park and for
// settling a MERGING card from its landed or failed queue row, or from a waiting row whose
// candidate's newest green, gate or review run failed (N63), and
// the position table still has to accept the move (P3-7). A MERGING card whose holder died
// and whose queue row is verifying, landing, or waiting with a candidate that can land cannot
// be moved or cancelled by anyone until the queue's owner settles the row.
func (s *Store) Move(ctx context.Context, cardID string, epoch int64, to core.Position, actor, reason string, ev ...Evidence) error {
	if strings.TrimSpace(actor) == "" {
		return fmt.Errorf("%w: a move needs its actor", ErrInvalid)
	}
	if err := checkGuards(ev); err != nil {
		return err
	}
	for _, e := range ev {
		if err := checkEvidence(e); err != nil {
			return err
		}
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		card, err := cardTx(ctx, tx, cardID)
		if err != nil {
			return err
		}
		var held int64
		hasLease := true
		switch err := tx.QueryRowContext(ctx, `SELECT epoch FROM leases WHERE card_id=?`, cardID).Scan(&held); {
		case err == sql.ErrNoRows:
			hasLease = false
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		}
		if hasLease {
			if held != epoch {
				return fmt.Errorf("%w: %s is at epoch %d, not %d", ErrStaleEpoch, cardID, held, epoch)
			}
		} else {
			// A park the ledger made holds no lease. Only the resume of that park, in
			// place, at the card's own epoch, is a move without one; so is the settling
			// of a MERGING card whose holder died, once its queue row of the card's own
			// epoch is landed or failed: the row, not a lease, is the authority then.
			inPlace := card.Position.Park != nil && to.Park == nil && to.Stage == card.Position.Stage && to.Sub == card.Position.Sub
			settles := false
			if !inPlace && epoch == card.Epoch && card.Position.Stage == core.StageMerging && card.Position.Park == nil {
				if settles, err = queueSettledAt(ctx, tx, cardID, card.Epoch); err != nil {
					return err
				}
				// A waiting row whose candidate's newest run failed: with the holder gone nothing else would settle it,
				// so it is the same kind of authority (N63).
				if !settles {
					if settles, err = queueUnlandableAt(ctx, tx, cardID, card.Epoch); err != nil {
						return err
					}
				}
			}
			if (!inPlace && !settles) || epoch != card.Epoch {
				return fmt.Errorf("%w: %s holds no lease", ErrStaleEpoch, cardID)
			}
		}
		if to.Stage == core.StageMerging && card.Position.Stage != core.StageMerging {
			return fmt.Errorf("%w: MERGING is entered through Enqueue", ErrInvalid)
		}
		var guards []string
		for _, e := range ev {
			if err := insertEvidence(ctx, tx, cardID, epoch, e); err != nil {
				return err
			}
			guards = append(guards, e.Guards...)
		}
		facts, err := factsTx(ctx, tx, card)
		if err != nil {
			return err
		}
		if err := core.CheckMove(card.Position, to, append(facts, guards...)); err != nil {
			return err
		}
		if err := setPosition(ctx, tx, cardID, to); err != nil {
			return err
		}
		if err := insertTransition(ctx, tx, cardID, card.Position, to, epoch, actor, reason); err != nil {
			return err
		}
		if to.Stage == core.StageDone || to.Stage == core.StageCancelled {
			if _, err := tx.ExecContext(ctx, `DELETE FROM lease_keys WHERE card_id=?`, cardID); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM leases WHERE card_id=?`, cardID); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		if card.Position.Park != nil && card.Position.Park.Reason == ParkReasonReclaimCap && to.Park == nil {
			if _, err := tx.ExecContext(ctx, `UPDATE cards SET reclaim_count=0 WHERE id=?`, cardID); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		return nil
	})
}
