// factory_wait.go — card t2013: record a lane's wait in the factory record and
// print, once, who to wake when the landing edge satisfies it. The verb that
// changes the condition only PRINTS the wake; the session that ran it does the
// send, and start is confirmed from the lease holder and the progress record.
package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryWaitLine flattens text to one line so a wait reason cannot break the
// one-line wake format.
func factoryWaitLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// factoryPrintWakes prints, once per wait, the lane and card to wake for every
// recorded wait the just-landed card satisfies. It prints nothing unless the
// card sits at a landing state (merged-local or merged-pr), so a lease expiry
// or any other transition never wakes anyone. The woken_at stamp is written in
// the same update that selects the wait, so a re-run prints nothing. A wait
// whose waker is the operator prints a leader-facing note, never a lane wake.
// Failures are fail-open: the landing already committed.
func factoryPrintWakes(ctx context.Context, out io.Writer, db *homestate.FactoryDB, card homestate.Card) {
	if card.State != homestate.CardMergedLocal && card.State != homestate.CardMergedPR {
		return
	}
	waits, err := db.SettleLandedWaits(ctx, card.CardID, factoryCardNow())
	if err != nil {
		_, _ = fmt.Fprintf(out, "  note: waits on %s were not read (%v)\n", card.CardID, err)
		return
	}
	for _, w := range waits {
		reason := factoryWaitLine(w.Reason)
		if w.Waker == homestate.WaitWakerOperator {
			_, _ = fmt.Fprintf(out, "  note: card=%s waits on the operator (predecessor %s landed); no lane wake - leader relays reason=%s\n", w.CardID, card.CardID, reason)
			continue
		}
		_, _ = fmt.Fprintf(out, "wake: lane=%s card=%s reason=%s (predecessor %s landed)\n", w.Waker, w.CardID, reason, card.CardID)
		_, _ = fmt.Fprintf(out, "  confirm the start from records, not from a reply: moai factory status (lease holder of %s) and the card progress record under .moai/reports/%s/\n", w.CardID, w.CardID)
	}
}

func newFactoryWaitCommand() *cobra.Command {
	var on, waker, reason, actor string
	var operator bool
	cmd := &cobra.Command{
		Use:   "wait <card>",
		Short: "Record why a card waits and who must wake it: --on <predecessor card> or --operator, --waker <lane|leader|operator>",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return factoryRecordWait(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0], on, operator, waker, reason, actor)
		},
	}
	cmd.Flags().StringVar(&on, "on", "", "predecessor card whose landing releases the wait (default: the card's --after hint)")
	cmd.Flags().BoolVar(&operator, "operator", false, "the wait is released by an operator ruling (never woken by a landing)")
	cmd.Flags().StringVar(&waker, "waker", "", "who must wake the card: a lane name, leader, or operator")
	cmd.Flags().StringVar(&reason, "reason", "", "why the card waits")
	cmd.Flags().StringVar(&actor, "actor", "", "who records the wait")
	return cmd
}

// factoryRecordWait records one wait for cardID. With --on empty and no
// --operator the card's existing `after` hint is the predecessor, reusing the
// dependency already recorded rather than restating it.
func factoryRecordWait(ctx context.Context, out io.Writer, root, cardID, on string, operator bool, waker, reason, actor string) error {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	defer func() { _ = db.Close() }()
	w := homestate.CardWait{CardID: cardID, Waker: waker, Reason: factoryWaitLine(reason), RecordedBy: actor}
	switch {
	case operator && on != "":
		return fmt.Errorf("factory wait: --on and --operator are exclusive")
	case operator:
		w.CondKind = homestate.WaitCondOperatorRuling
	default:
		w.CondKind, w.CondRef = homestate.WaitCondPredecessorLanded, on
		if w.CondRef == "" {
			runID, err := resolveFactoryCardRun(ctx, root, "")
			if err != nil {
				return fmt.Errorf("factory wait: no --on given and the run is unresolved: %w", err)
			}
			card, err := db.LoadCard(ctx, runID, cardID)
			if err != nil {
				return fmt.Errorf("factory wait: no --on given and the card record is unreadable: %w", err)
			}
			w.CondRef = card.HintAfter
		}
	}
	if err := db.RecordCardWait(ctx, w, factoryCardNow()); err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	_, _ = fmt.Fprintf(out, "wait recorded: card=%s condition=%s %s waker=%s\n", cardID, w.CondKind, w.CondRef, w.Waker)
	return nil
}
