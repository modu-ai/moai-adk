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
	case strings.TrimSpace(e.TreeSHA) == "":
		return fmt.Errorf("%w: evidence is bound to a tree SHA", ErrInvalid)
	}
	return nil
}

// Move changes the card's position; see the API for the epoch, the facts and the guards.
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
			// place, at the card's own epoch, is a move without one.
			inPlace := card.Position.Park != nil && to.Park == nil && to.Stage == card.Position.Stage && to.Sub == card.Position.Sub
			if !inPlace || epoch != card.Epoch {
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
