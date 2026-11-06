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
// card sits at a landing or post-landing state (factoryLandedState), so a lease expiry
// or any other transition never wakes anyone. The woken_at stamp is written in
// the same update that selects the wait, so a re-run prints nothing. A wait
// whose waker is the operator prints a leader-facing note, never a lane wake.
// Failures are fail-open: the landing already committed.
func factoryPrintWakes(ctx context.Context, out io.Writer, db *homestate.FactoryDB, card homestate.Card) {
	factoryPrintWakesSuffixed(ctx, out, db, card, "")
}

// factoryLandedState reports whether state is a landing state, or a state a
// card only reaches after landing (pushed, ci-green, done), that releases a
// wait on a predecessor. The later states matter because a wait recorded after
// the predecessor moved past the landing would otherwise stay open forever.
// abandoned and failed never landed and release nothing.
func factoryLandedState(state string) bool {
	switch state {
	case homestate.CardMergedLocal, homestate.CardMergedPR,
		homestate.CardPushed, homestate.CardCIGreen, homestate.CardDone:
		return true
	}
	return false
}

// factoryPrintWakesSuffixed is factoryPrintWakes with an extra suffix on each
// printed line; the record-time wake uses it to say the predecessor had
// already landed.
func factoryPrintWakesSuffixed(ctx context.Context, out io.Writer, db *homestate.FactoryDB, card homestate.Card, suffix string) {
	if !factoryLandedState(card.State) {
		return
	}
	waits, err := db.SettleLandedWaits(ctx, card.CardID, factoryCardNow())
	if err != nil {
		_, _ = fmt.Fprintf(out, "  note: waits on %s were not read (%v)\n", card.CardID, err)
		return
	}
	for _, w := range waits {
		reason, missing := factoryWaitDisplay(w.Reason)
		if w.Waker == homestate.WaitWakerOperator {
			_, _ = fmt.Fprintf(out, "  note: card=%s waits on the operator (predecessor %s landed); no lane wake - leader relays reason=%s%s%s\n", w.CardID, card.CardID, reason, missing, suffix)
			continue
		}
		_, _ = fmt.Fprintf(out, "wake: lane=%s card=%s reason=%s (predecessor %s landed)%s%s\n", w.Waker, w.CardID, reason, card.CardID, missing, suffix)
		_, _ = fmt.Fprintf(out, "  confirm the start from records, not from a reply: moai factory status (lease holder of %s) and the card progress record under .moai/reports/%s/\n", w.CardID, w.CardID)
	}
}

// factoryWaitTag validates a --scope / --next value: one trimmed non-empty line.
func factoryWaitTag(flag string, v *string, what string) (string, error) {
	s := strings.TrimSpace(*v)
	if s == "" || strings.ContainsAny(s, "\r\n") {
		return "", fmt.Errorf("%w: --%s must be one non-empty line naming %s", homestate.ErrInvalidCardInput, flag, what)
	}
	return s, nil
}

// factoryWaitReason folds the optional scope and next-check event into the
// reason text as leading `scope=<value> ` / `next=<value> ` tokens (in that
// order), so the stored reason, the record line, and the wake line all carry
// them without a schema change. An absent flag adds nothing: a call with
// neither stores the flattened reason byte-for-byte as before.
func factoryWaitReason(reason string, scope, next *string) (string, error) {
	parts := make([]string, 0, 3)
	if scope != nil {
		s, err := factoryWaitTag("scope", scope, "the refused operation, path, or permission")
		if err != nil {
			return "", err
		}
		parts = append(parts, "scope="+s)
	}
	if next != nil {
		n, err := factoryWaitTag("next", next, "the event that triggers the next check")
		if err != nil {
			return "", err
		}
		parts = append(parts, "next="+n)
	}
	parts = append(parts, factoryWaitLine(reason))
	return strings.TrimSpace(strings.Join(parts, " ")), nil
}

// factoryWaitDisplay renders a stored wait reason for the wake and note lines:
// an empty reason shows as `unknown`, and the tracking fields a record does
// not carry (it predates the tag, or the call omitted it) are appended as
// `scope=unknown` / `next=unknown`. Display only - stored data is never
// rewritten, and a missing field never blocks a read or a wake.
func factoryWaitDisplay(stored string) (reason, missing string) {
	reason = factoryWaitLine(stored)
	hasScope := strings.HasPrefix(reason, "scope=")
	hasNext := strings.HasPrefix(reason, "next=") || (hasScope && strings.Contains(reason, " next="))
	if reason == "" {
		reason = "unknown"
	}
	if !hasScope {
		missing += " scope=unknown"
	}
	if !hasNext {
		missing += " next=unknown"
	}
	return reason, missing
}

func newFactoryWaitCommand() *cobra.Command {
	var on, waker, reason, actor, scope, next, ack, run string
	var operator, resolve bool
	cmd := &cobra.Command{
		Use:   "wait <card>",
		Short: "Record why a card waits and who must wake it: --on <predecessor card> or --operator, --waker <lane|leader|operator>",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if resolve {
				for _, f := range []string{"on", "operator", "waker", "reason", "scope", "next", "ack"} {
					if cmd.Flags().Changed(f) {
						return fmt.Errorf("factory wait: --resolve closes the operator wait and cannot be combined with --%s", f)
					}
				}
				return factoryResolveOperatorWait(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0])
			}
			if cmd.Flags().Changed("ack") {
				for _, f := range []string{"on", "operator", "waker", "reason", "scope", "next"} {
					if cmd.Flags().Changed(f) {
						return fmt.Errorf("factory wait: --ack records a wake event and cannot be combined with --%s", f)
					}
				}
				return factoryAckWake(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0], ack, actor)
			}
			var scopeArg *string
			if cmd.Flags().Changed("scope") {
				scopeArg = &scope
			}
			var nextArg *string
			if cmd.Flags().Changed("next") {
				nextArg = &next
			}
			return factoryRecordWait(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0], on, operator, waker, reason, actor, run, scopeArg, nextArg)
		},
	}
	cmd.Flags().StringVar(&on, "on", "", "predecessor card whose landing releases the wait (default: the card's --after hint)")
	cmd.Flags().BoolVar(&operator, "operator", false, "the wait is released by an operator ruling (never woken by a landing)")
	cmd.Flags().BoolVar(&resolve, "resolve", false, "close the card's open --operator wait: the operator ruled, so the wait ends (idempotent; refused when the card has no operator wait)")
	cmd.Flags().StringVar(&waker, "waker", "", "who must wake the card: a lane name, leader, or operator")
	cmd.Flags().StringVar(&reason, "reason", "", "why the card waits")
	cmd.Flags().StringVar(&actor, "actor", "", "who records the wait")
	cmd.Flags().StringVar(&scope, "scope", "", "the refused operation, path, or permission that caused the wait (one line)")
	cmd.Flags().StringVar(&next, "next", "", "the event that triggers the next check of this wait (one line)")
	cmd.Flags().StringVar(&ack, "ack", "", "record the waker's wake event instead of a wait: received (acknowledged) or started (began the work); each is recorded once, in order")
	cmd.Flags().StringVar(&run, "run", "", "factory run id, used to read the card's --after hint when --on is omitted (default: the single active run; a run whose row is not active refuses stale)")
	return cmd
}

// factoryAckWake records the received or started event of the card's oldest
// delivered wait still awaiting that event. Delivery itself is written by the
// landing edge. An event recorded out of order is refused and writes nothing.
func factoryAckWake(ctx context.Context, out io.Writer, root, cardID, phase, actor string) error {
	if phase != homestate.WakeReceived && phase != homestate.WakeStarted {
		return fmt.Errorf("factory wait: --ack takes %s or %s, got %q (delivery is recorded by the landing edge)", homestate.WakeReceived, homestate.WakeStarted, phase)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	defer func() { _ = db.Close() }()
	waits, err := db.LoadCardWaits(ctx, cardID)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	if len(waits) == 0 {
		return fmt.Errorf("factory wait: card %s has no recorded wait to acknowledge", cardID)
	}
	// The wake the waker acknowledges is the card's oldest wait that holds the
	// phase's predecessor event and not the phase itself (received follows
	// delivered, started follows received). With no such wait the card's latest
	// wait is used, so a repeat is reported as already recorded and an event out
	// of order is refused by the ordering rule below, as before.
	w := waits[len(waits)-1]
	for _, candidate := range waits {
		st, err := db.LoadWakeState(ctx, candidate.Seq)
		if err != nil {
			return fmt.Errorf("factory wait: %w", err)
		}
		if (phase == homestate.WakeReceived && st.Delivered && !st.Received) ||
			(phase == homestate.WakeStarted && st.Received && !st.Started) {
			w = candidate
			break
		}
	}
	recorded, err := db.RecordWakeEvent(ctx, w.Seq, phase, actor, factoryCardNow())
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	if !recorded {
		_, _ = fmt.Fprintf(out, "wake %s already recorded: card=%s wait=%d\n", phase, cardID, w.Seq)
		return nil
	}
	_, _ = fmt.Fprintf(out, "wake %s recorded: card=%s wait=%d\n", phase, cardID, w.Seq)
	return nil
}

// factoryResolveOperatorWait closes the card's open operator-ruling waits. A
// landing never wakes such a wait, so without this path an open one would stay
// valid for the stop-goal hook forever. A card that never recorded an operator
// wait is refused; one whose operator waits are all closed already succeeds
// without writing.
func factoryResolveOperatorWait(ctx context.Context, out io.Writer, root, cardID string) error {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	defer func() { _ = db.Close() }()
	waits, err := db.LoadCardWaits(ctx, cardID)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	operatorWaits := 0
	for _, w := range waits {
		if w.CondKind == homestate.WaitCondOperatorRuling {
			operatorWaits++
		}
	}
	if operatorWaits == 0 {
		return fmt.Errorf("factory wait: card %s has no operator wait to resolve", cardID)
	}
	closed, err := db.ResolveOperatorWaits(ctx, cardID, factoryCardNow())
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	if closed == 0 {
		_, _ = fmt.Fprintf(out, "operator wait already resolved: card=%s\n", cardID)
		return nil
	}
	_, _ = fmt.Fprintf(out, "operator wait resolved: card=%s closed=%d\n", cardID, closed)
	return nil
}

// factoryRecordWait records one wait for cardID. With --on empty and no
// --operator the card's existing `after` hint is the predecessor, reusing the
// dependency already recorded rather than restating it.
func factoryRecordWait(ctx context.Context, out io.Writer, root, cardID, on string, operator bool, waker, reason, actor, run string, scope, next *string) error {
	waitReason, err := factoryWaitReason(reason, scope, next)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory wait: %w", err)
	}
	defer func() { _ = db.Close() }()
	w := homestate.CardWait{CardID: cardID, Waker: waker, Reason: waitReason, RecordedBy: actor}
	switch {
	case operator && on != "":
		return fmt.Errorf("factory wait: --on and --operator are exclusive")
	case operator:
		w.CondKind = homestate.WaitCondOperatorRuling
	default:
		w.CondKind, w.CondRef = homestate.WaitCondPredecessorLanded, on
		if w.CondRef == "" {
			runID, err := resolveFactoryCardRun(ctx, root, run)
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
	predecessorWait := w.CondKind == homestate.WaitCondPredecessorLanded && w.CondRef != ""
	// An identical wait already woken because its predecessor landed is
	// history: re-recording it would open a fresh wait that wakes again.
	if !predecessorWait || !factoryWaitAlreadyDelivered(ctx, db, w) {
		if err := db.RecordCardWait(ctx, w, factoryCardNow()); err != nil {
			return fmt.Errorf("factory wait: %w", err)
		}
	}
	line := fmt.Sprintf("wait recorded: card=%s condition=%s %s waker=%s", cardID, w.CondKind, w.CondRef, w.Waker)
	if scope != nil {
		line += " scope=" + strings.TrimSpace(*scope)
	}
	if next != nil {
		line += " next=" + strings.TrimSpace(*next)
	}
	_, _ = fmt.Fprintln(out, line)
	if predecessorWait {
		factoryWakeIfPredecessorLanded(ctx, out, db, w.CondRef)
	}
	return nil
}

// factoryLandedPredecessor returns the record of card ref when any run holds
// it at a landing state.
func factoryLandedPredecessor(ctx context.Context, db *homestate.FactoryDB, ref string) (homestate.Card, bool, error) {
	cards, err := db.ListCards(ctx, "")
	if err != nil {
		return homestate.Card{}, false, err
	}
	for _, c := range cards {
		if c.CardID == ref && factoryLandedState(c.State) {
			return c, true, nil
		}
	}
	return homestate.Card{}, false, nil
}

// factoryWakeIfPredecessorLanded closes the gap the landing edge leaves: a wait
// recorded after its predecessor already landed (including a landing that
// raced the record) would never see an edge. It reads the predecessor's
// current state after the record and, when landed, settles through the same
// SettleLandedWaits path the edge uses, so the woken_at guard keeps the wake
// to at most once. An unknown or not-yet-landed predecessor changes nothing.
// Fail-open: the wait is already recorded.
func factoryWakeIfPredecessorLanded(ctx context.Context, out io.Writer, db *homestate.FactoryDB, ref string) {
	card, landed, err := factoryLandedPredecessor(ctx, db, ref)
	if err != nil {
		_, _ = fmt.Fprintf(out, "  note: state of %s was not read (%v)\n", ref, err)
		return
	}
	if landed {
		factoryPrintWakesSuffixed(ctx, out, db, card, " - predecessor had already landed when the wait was recorded")
	}
}

// factoryWaitAlreadyDelivered reports whether an identical wait was already
// woken and its predecessor is still at a landing state.
func factoryWaitAlreadyDelivered(ctx context.Context, db *homestate.FactoryDB, w homestate.CardWait) bool {
	waits, err := db.LoadCardWaits(ctx, w.CardID)
	if err != nil {
		return false
	}
	for _, x := range waits {
		if x.WokenAt != "" && x.CondKind == w.CondKind && x.CondRef == w.CondRef &&
			x.Waker == w.Waker && x.Reason == w.Reason && x.RecordedBy == w.RecordedBy {
			_, landed, err := factoryLandedPredecessor(ctx, db, w.CondRef)
			return err == nil && landed
		}
	}
	return false
}
