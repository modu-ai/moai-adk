// integration_codemaps_card.go — the codemaps-debt standing source (card t1018).
//
// WHAT THIS IS. A standing source in the sense the queue doctrine already
// defines: a workflow the operator authorized ONCE to issue a card when it
// finishes, rather than being asked for that card every time. The five
// properties that separate issuing from inventing are carried here — one card
// per threshold crossing, derived never invented, marked at the front, the
// issued id reported, and starting it a separate pick (this file appends a
// QUEUED card and never picks one).
//
// WHY `moai integration release` AND NOT THE SESSION-START HOOK. Both places
// can reach the queue and both cost about the same. They differ on one
// question: should the value be read in a batch that created no debt? The
// session-start hook runs on every session start and every `/clear`, so it
// reads the value whether or not anything landed. This verb runs exactly when
// a card's work lands in the integration branch — the event that PRODUCES the
// debt this card is about — so the check fires when, and only when, the thing
// it measures can have changed. That is the operator's decision (2026-09-20),
// recorded here rather than only in the card, because a wiring without its
// reason invites the next reader to move it back.
//
// WHY NOT THE PROCESS EXIT CODE. `moai graph check` exits non-zero when ANY
// layer is not fresh, and in a fresh worktree the `mx-index` and `edges`
// layers are `absent` (untracked derived artifacts) while codemaps is `fresh`.
// Gating on the exit code would therefore issue a codemaps-debt card from
// every new worktree, for reasons that have nothing to do with codemaps. The
// trigger reads the codemaps LAYER — its own value against its own threshold
// — and ignores every other layer's verdict.
package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// codemapsDebtCardPrefix marks the card as machine-issued, so a reader
// scanning the queue can tell it from one a person typed.
const codemapsDebtCardPrefix = "[GRAPH] "

// codemapsDebtCardText returns the card text for res, whether a card should be
// issued at all, and — when it should not — why not.
//
// The text is keyed on the CONTENT ANCHOR and carries no volatile figure. That
// is not a stylistic choice; it is what makes the queue's exact-duplicate
// refusal able to suppress the second lane's card. Measured on 2026-09-20
// against an isolated queue: two adds of the same text are refused with
// `todo add: t1 already holds this card`, while two texts differing only in
// the measured value (`... 45 >= 40 (anchor X, +120 commits)` and
// `... 46 >= 40 (anchor X, +121 commits)`) were BOTH admitted — t1 and t2. A
// card text carrying the value or the commit count is therefore a new text on
// every landing, and the suppression the design relies on never fires. The
// value still reaches the operator: it is printed on the issuing line, and
// `moai graph check` re-measures it on demand.
//
// The anchor is written in full rather than abbreviated. An abbreviation is a
// second decision (how many characters) that must then stay fixed forever for
// the key to stay stable, and it buys only line length in a listing that
// truncates anyway.
func codemapsDebtCardText(res graph.CheckResult) (text string, issue bool, reason string) {
	var layer *graph.LayerReport
	for i := range res.Layers {
		if res.Layers[i].Layer == graph.LayerCodemaps {
			layer = &res.Layers[i]
			break
		}
	}
	if layer == nil {
		return "", false, "no codemaps layer in the report"
	}
	if layer.Verdict == graph.VerdictAbsent {
		// Absent is UNJUDGEABLE, not zero. The layer reports Value 0 when it
		// could not measure — a missing codemaps directory, an unreadable
		// provenance block — and reading that 0 as "no debt" would silently
		// convert "we do not know" into "there is nothing to do". Declining
		// is right, because a `described-source-diff >= 40` card would be a
		// claim the report does not support; saying so out loud is also
		// right, because a permanently unjudgeable tree otherwise looks
		// exactly like a tree that keeps coming in clean.
		return "", false, fmt.Sprintf("codemaps layer unjudgeable (%s) — not issuing a card on an unmeasured value", layer.Reason)
	}
	if layer.Value < layer.Threshold {
		return "", false, fmt.Sprintf("%s (%d < %d)", codemapsBelowThresholdReason, layer.Value, layer.Threshold)
	}
	if layer.ContentAnchor == "" {
		// No anchor means no stable key, and a card with no stable key is a
		// card issued again on every landing. Declining is the conservative
		// direction: the next landing that does resolve an anchor issues it.
		return "", false, "codemaps over threshold but no content anchor resolved — no stable card key"
	}
	return fmt.Sprintf("%scodemaps %s >= %d (anchor %s)",
		codemapsDebtCardPrefix, layer.Metric, layer.Threshold, layer.ContentAnchor), true, ""
}

// codemapsBelowThresholdReason is the prefix of the ordinary decline, which
// is the common outcome and not worth a line on every landing. It is a
// constant rather than a repeated literal so the printer and the producer
// cannot drift into disagreeing about which declines are routine.
const codemapsBelowThresholdReason = "codemaps below threshold"

// errCodemapsCardAlreadyQueued aborts the queue write when the card is
// already present. Returning an error from inside Mutate is what leaves the
// queue file byte-identical — the same contract the `todo add` refusal path
// stands on, rather than a second way of not-writing invented here.
var errCodemapsCardAlreadyQueued = errors.New("codemaps debt card already queued")

// issueCodemapsDebtCard is the standing source's body: measure, and append a
// queued card when the codemaps layer is at or over its threshold.
//
// FAIL-OPEN, WITHOUT EXCEPTION. It returns no error and its caller ignores
// its outcome. By the time it runs the integration window has already been
// released, so failing the command afterwards would report a release that DID
// happen as a failure — worse than not issuing the card. Every declining path
// still says why on stderr, so a silent non-issue and a broken trigger stay
// distinguishable; that distinction is the reason the declines are reported
// rather than swallowed.
//
// It appends a QUEUED card and never picks one: what the machine creates is a
// queue entry, not started work. Choosing to start it stays the operator's.
func issueCodemapsDebtCard(stdout, stderr io.Writer, projectRoot string) {
	res, err := graph.CheckFreshness(projectRoot, graph.DefaultThresholds())
	if err != nil {
		// The report is partial but still worth reading: the codemaps row may
		// have been measured before the walk aborted on a later layer.
		_, _ = fmt.Fprintf(stderr, "codemaps debt check: %v\n", err)
	}
	text, issue, reason := codemapsDebtCardText(res)
	if !issue {
		if reason != "" && !strings.HasPrefix(reason, codemapsBelowThresholdReason) {
			_, _ = fmt.Fprintf(stderr, "codemaps debt check: %s\n", reason)
		}
		return
	}

	var issued kanban.BacklogItem
	var existingID string
	mutErr := newTodoStore().Mutate(func(rec *kanban.BacklogRecord) error {
		if match := kanban.ClassifyCardText(text, rec.Items); match.Kind == kanban.BacklogMatchExact {
			existingID = match.ID
			return errCodemapsCardAlreadyQueued
		}
		var appendErr error
		issued, _, appendErr = appendAnalyzedCard(rec, text, kanban.BacklogStateQueued, false)
		return appendErr
	})
	switch {
	case errors.Is(mutErr, errCodemapsCardAlreadyQueued):
		// Not a failure: this is the suppression doing its job when a second
		// lane lands inside the same debt period.
		_, _ = fmt.Fprintf(stderr, "codemaps debt already queued as %s\n", existingID)
	case mutErr != nil:
		_, _ = fmt.Fprintf(stderr, "codemaps debt card not issued: %v\n", mutErr)
	default:
		// The issued id is reported, so the queue cannot grow unobserved.
		_, _ = fmt.Fprintf(stdout, "codemaps debt card issued: %s %s\n", issued.ID, text)
	}
}
