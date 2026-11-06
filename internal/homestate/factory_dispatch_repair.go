package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// factory_dispatch_repair.go — the atomic factory half of a dispatch repair
// (SPEC-FACTORY-COMPLETION-RECOVERY-001 REQ-FCR-024, card t1963).
//
// The repair used to be four separate transactions — the retired-run check, the
// card row (RecordPicked), the picked → assigned edge, and the engagement
// (RecordDispatchEngagement). A consolidation committing between the check and
// the first write retired the run and moved its cards, and the writes then
// recreated the card's row and binding in the run it had retired. The
// consolidation takes the factory write lock and the queue lock does not stop
// it, so the only fix is to hold the write lock across the check AND the writes:
// the factory DSN opens transactions IMMEDIATE, so this one takes the lock at
// BEGIN, and a consolidation waits for the repair's commit.

// RepairDispatchAtomic completes the factory-record half of a dispatch for
// cardID in runID under owner, in ONE write transaction:
//
//  1. refuse with ErrRunRetiredByConsolidation when a consolidation retired
//     runID (the same marker RunRetiredByConsolidation reads);
//  2. load the card's row in runID, or record it at picked (the mirror's T1);
//  3. move a picked row to assigned under owner (the mirror's T2);
//  4. the engagement: ensure the run row, move a row that carries a DIFFERENT
//     owner to owner (refused while the row holds a valid lease), and point the
//     card's dispatch binding at runID — exactly RecordDispatchEngagement's
//     semantics.
//
// All or nothing: any refusal or failure rolls back every write, so a failed
// repair no longer leaves a picked row behind. canon maps a label into the
// caller's canonical owner vocabulary (nil = identity). afterCheck (nil = none) is
// a test-only seam that runs after the check and before the first write, inside
// the transaction.
//
// @MX:WARN: [AUTO] the retired-run check and the repair's writes must stay in this one transaction
// @MX:REASON: moving the check out lets a consolidation commit in between and the writes recreate work in the retired run (REQ-FCR-024 atomicity)
func (f *FactoryDB) RepairDispatchAtomic(ctx context.Context, cardID, runID, owner string, canon func(string) string, now time.Time, afterCheck func()) error {
	cardID, runID, owner = strings.TrimSpace(cardID), strings.TrimSpace(runID), strings.TrimSpace(owner)
	if cardID == "" || runID == "" {
		return fmt.Errorf("%w: card id and run id are required", ErrInvalidCardInput)
	}
	if canon == nil {
		canon = func(label string) string { return label }
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	nowText := now.Format(time.RFC3339Nano)
	return f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		if err := refuseIfRetiredByConsolidation(ctx, tx, runID); err != nil {
			return nil, err
		}
		if afterCheck != nil {
			afterCheck()
		}
		card, err := loadCard(ctx, tx, runID, cardID)
		switch {
		case errors.Is(err, ErrCardNotFound):
			if card, err = recordPickedTx(ctx, tx, runID, cardID, CardFields{}, "dispatch", now); err != nil {
				return nil, err
			}
		case err != nil:
			return nil, err
		}
		if card.State == CardPicked {
			if _, err := transitionTx(ctx, f, tx, TransitionRequest{
				RunID: runID, CardID: cardID, To: CardAssigned, ExpectedVersion: card.Version,
				Actor: "dispatch", Owner: owner, Now: now,
			}, now); err != nil {
				return nil, err
			}
		}
		if err := ensureRunRowTx(ctx, tx, runID, nowText); err != nil {
			return nil, err
		}
		if owner != "" {
			cur, lerr := loadCard(ctx, tx, runID, cardID)
			switch {
			case errors.Is(lerr, ErrCardNotFound):
				// No row in this run: nothing carries a prior owner.
			case lerr != nil:
				return nil, lerr
			case strings.TrimSpace(cur.OwnerLabel) != "" && canon(cur.OwnerLabel) != canon(owner):
				if err := restampOwnerTx(ctx, tx, cur, canon(owner), now); err != nil {
					return nil, err
				}
			}
		}
		if err := upsertDispatchBindingTx(ctx, tx, cardID, runID, nowText); err != nil {
			return nil, err
		}
		return nil, nil
	})
}
