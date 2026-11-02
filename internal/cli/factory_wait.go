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
		reason, missing := factoryWaitDisplay(w.Reason)
		if w.Waker == homestate.WaitWakerOperator {
			_, _ = fmt.Fprintf(out, "  note: card=%s waits on the operator (predecessor %s landed); no lane wake - leader relays reason=%s%s\n", w.CardID, card.CardID, reason, missing)
			continue
		}
		_, _ = fmt.Fprintf(out, "wake: lane=%s card=%s reason=%s (predecessor %s landed)%s\n", w.Waker, w.CardID, reason, card.CardID, missing)
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
	var on, waker, reason, actor, scope, next string
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
			var scopeArg *string
			if cmd.Flags().Changed("scope") {
				scopeArg = &scope
			}
			var nextArg *string
			if cmd.Flags().Changed("next") {
				nextArg = &next
			}
			return factoryRecordWait(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0], on, operator, waker, reason, actor, scopeArg, nextArg)
		},
	}
	cmd.Flags().StringVar(&on, "on", "", "predecessor card whose landing releases the wait (default: the card's --after hint)")
	cmd.Flags().BoolVar(&operator, "operator", false, "the wait is released by an operator ruling (never woken by a landing)")
	cmd.Flags().StringVar(&waker, "waker", "", "who must wake the card: a lane name, leader, or operator")
	cmd.Flags().StringVar(&reason, "reason", "", "why the card waits")
	cmd.Flags().StringVar(&actor, "actor", "", "who records the wait")
	cmd.Flags().StringVar(&scope, "scope", "", "the refused operation, path, or permission that caused the wait (one line)")
	cmd.Flags().StringVar(&next, "next", "", "the event that triggers the next check of this wait (one line)")
	return cmd
}

// factoryRecordWait records one wait for cardID. With --on empty and no
// --operator the card's existing `after` hint is the predecessor, reusing the
// dependency already recorded rather than restating it.
func factoryRecordWait(ctx context.Context, out io.Writer, root, cardID, on string, operator bool, waker, reason, actor string, scope, next *string) error {
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
	line := fmt.Sprintf("wait recorded: card=%s condition=%s %s waker=%s", cardID, w.CondKind, w.CondRef, w.Waker)
	if scope != nil {
		line += " scope=" + strings.TrimSpace(*scope)
	}
	if next != nil {
		line += " next=" + strings.TrimSpace(*next)
	}
	_, _ = fmt.Fprintln(out, line)
	return nil
}
