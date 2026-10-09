package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryAuditDecideCards applies a kickoff approval by the audit decider
// (edge T8a) to each named card. In a lane session the lane may approve only
// a card whose record owner it is. The queue item's hold state is read here,
// inside the transition's transaction, right before the commit; an unreadable
// queue fails closed.
func factoryAuditDecideCards(ctx context.Context, root string, out io.Writer, cards []string, run string) error {
	runID, err := resolveFactoryCardRun(ctx, root, run)
	if err != nil {
		return fmt.Errorf("factory decide: %w", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory decide: %w", err)
	}
	defer func() { _ = db.Close() }()
	lane := strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker))
	inLane := factoryLaneRefusal()
	actor := "operator"
	if inLane {
		actor = lane
	}
	refused := 0
	for _, cardID := range cards {
		cur, err := db.LoadCard(ctx, runID, cardID)
		if err == nil && inLane && (lane == "" || cur.OwnerLabel != lane) {
			err = fmt.Errorf("refused — %s: a lane may approve by audit only a card it owns (owner %q, lane %q)", factoryLaneBoundarySentinel, cur.OwnerLabel, lane)
		}
		if err == nil && cur.State != homestate.CardKickoff {
			err = fmt.Errorf("card is %s, not %s", cur.State, homestate.CardKickoff)
		}
		if err == nil {
			// T8a leases from this run and re-points the binding onto it in the
			// same transaction, so the queue's current-dispatch record has to
			// follow (turn-end gate, card t1538): otherwise a retried older
			// dispatch reads the stale record as current and drags the binding
			// back, and the older run's approval verifies again beside the new
			// run's valid lease. The record is written under the queue lock,
			// inside the factory transaction right before its commit — an
			// unwritable record rolls the approval back, and a refused approval
			// never reaches the hook.
			err = todoStoreAt(root).WithLock(func(l *factory.LockedBacklog) error {
				var terr error
				cur, terr = db.Transition(ctx, homestate.TransitionRequest{
					RunID: runID, CardID: cardID, To: homestate.CardRun, ExpectedVersion: cur.Version,
					// The hold is read inside the transition, right before the commit.
					Actor: actor, Decider: homestate.DeciderAudit, QueueHoldRead: func() string { return factoryQueueHold(cardID) }, Now: factoryCardNow(),
					BeforeCommit: func(next homestate.Card) error { return l.RefreshDispatchCurrent(cardID, runID, next.OwnerLabel) },
				})
				return terr
			})
		}
		if err != nil {
			refused++
			_, _ = fmt.Fprintf(out, "%s: refused: %v\n", cardID, err)
			continue
		}
		_, _ = fmt.Fprintf(out, "%s: %s (v%d)\n", cardID, cur.State, cur.Version)
	}
	if refused > 0 {
		return fmt.Errorf("factory decide: %d of %d cards refused", refused, len(cards))
	}
	return nil
}

// factoryQueueHold reads the queue item's hold state for a card. A queue
// that cannot be read, or that does not hold the card, reads as unreadable,
// which the audit decider refuses (fail closed).
func factoryQueueHold(cardID string) string {
	rec, err := newTodoReadStore().LoadPure()
	if err != nil || rec == nil {
		return homestate.QueueHoldUnreadable
	}
	for _, it := range rec.Items {
		if it.ID == cardID {
			if string(it.State) == string(homestate.QueueHoldHeld) {
				return homestate.QueueHoldHeld
			}
			return homestate.QueueHoldClear
		}
	}
	return homestate.QueueHoldUnreadable
}
