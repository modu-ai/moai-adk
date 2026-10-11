// todo_auto_lane.go — card t1554: the unified `--auto` engine's lane surface.
//
// A lane session running `moai todo --auto` executes the serial cycle's
// foreman contract with three differences that preserve queue integrity by
// construction:
//
//   - SELECTION is the shared ranking stage (autoRank — Jev when the
//     capability answers, else the priority/readiness fallback), not the
//     lease's queue-order arms; the ranked top candidate is nominated through
//     the nominated lease edge, the same write `moai factory next --card <id>`
//     performs, whose keep-set predicates (hold, hold-marker, blocked, owned,
//     serial-slot, quota-hold, backend-skip) bound the selection mechanically.
//     A permanent refusal falls through to the next-ranked candidate; a
//     waitable one (raced / serial-slot / quota-hold) ends the pass —
//     retrying belongs to the caller's next invocation.
//   - QUOTA STEERING (SPEC-QUOTA-AWARE-SCHEDULING-001 REQ-QAS-009): while the
//     quota latch holds this lane, no NEW card is taken — the gated lease
//     leaves only the card already assigned to this lane leasable.
//   - COMPLETION writes no queue byte: the cycle records no done and performs
//     no unpick. The card stays leased; done-after-completion belongs to the
//     existing completion path (the leader's evidence read or `moai factory
//     complete`), and a missed deadline leaves the lease to the F1 expiry
//     machinery — the same semantic the lane boot loops carry for a failed
//     card session (REQ-SD-003's continue-with-next).
//
// One implementation serves all three backends' lanes (cc / glm / codex): the
// engine is backend-neutral, and every backend difference lives in the lease
// machinery the cycle calls. The cycle's preconditions are the lane verbs'
// own: lane admission (REQ-SD-015 — the cycle is never wider than `moai
// factory next`, so a label-only session is refused exactly as that verb
// refuses it) and the parent checkout (REQ-SD-010).
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// autoLaneRunResolveFn resolves the factory run the lane cycle leases from.
// A package var so tests pin the fixture run; production prefers the
// launcher-selected run (config.EnvFactoryRunID — the launcher stamps the
// chosen run into the lane environment) and falls back to single-active-run
// discovery only when the environment names none (card t1577). Both branches
// resolve through factorymsg.ResolveActiveRun, whose explicit branch checks
// the selection against the active-run table (safeID + active status) — a
// stale launcher selection fails closed instead of leasing blind (gate
// round 5: resolveFactoryCardRun's explicit path returns unvalidated).
var autoLaneRunResolveFn = func(ctx context.Context, root string) (string, error) {
	if run := strings.TrimSpace(os.Getenv(config.EnvFactoryRunID)); run != "" {
		return factorymsg.ResolveActiveRun(ctx, root, run)
	}
	return factorymsg.ResolveActiveRun(ctx, root, "")
}

// runAutoLaneCycle executes the lane `--auto` cycle against the queue at root,
// writing the narrated output (selection record → lease → directive →
// evidence → completion handoff, per card) to out and the refusal/fall-through
// notes to errOut. Exactly one card is in flight at any time; the loop stops
// when nothing leases (REQ-SD-003's stop condition).
func runAutoLaneCycle(ctx context.Context, root string, out, errOut io.Writer, opts autoOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// The loop drives the F1 lease machinery itself, so it inherits the `next`
	// verb's own preconditions: the lane boundary (REQ-SD-015 — the cycle is
	// never wider than the lane verbs it replaces, so a label-only session
	// that `factory next` refuses is refused here too) and the parent
	// checkout (REQ-SD-010).
	if !factoryLaneAdmission() {
		return factoryNotALaneError("todo --auto")
	}
	if err := factoryAssertParentCheckout(resolveProjectDir()); err != nil {
		return err
	}
	lane, err := factoryLaneLabelFromEnv("todo --auto")
	if err != nil {
		return err
	}
	runID, err := autoLaneRunResolveFn(ctx, root)
	if err != nil {
		return fmt.Errorf("todo --auto: %w", err)
	}
	store := todoStoreAt(root)
	if opts.sleep == nil {
		opts.sleep = time.Sleep
	}
	if opts.now == nil {
		opts.now = time.Now
	}
	latch := &factoryQuotaLatch{}
	for {
		// Quota steering evaluates once per pass, outside the selection arms
		// (the same shape the `next` verb's latch loop uses).
		held, holdLine := latch.evaluate(root)
		card, item, leased, err := autoLeaseRanked(ctx, store, root, runID, lane, held, holdLine, out, errOut, opts)
		if err != nil {
			return fmt.Errorf("todo --auto: %w", err)
		}
		if !leased {
			return nil // no card available: the loop's stop condition (REQ-SD-003)
		}
		wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, lane, errOut)
		if err != nil {
			return fmt.Errorf("todo --auto: %w", err)
		}
		_, _ = fmt.Fprintf(out, "%s stage=%s worktree=%s\n", card.CardID, dash(card.Stage), filepath.Base(wt))
		autoLaneWorkCard(out, item, root, opts)
	}
}

// autoLeaseRanked is the unified selection engine: under a quota hold it
// delegates to the gated lease (the lane's own assigned card alone); otherwise
// it ranks the queued candidates through the shared ranking stage (the serial
// cycle's autoRank — Jev when the capability answers, else the
// priority/readiness fallback) and nominates them in ranked order through the
// nominated lease edge. The item of the leased card rides along for the
// dispatch directive.
func autoLeaseRanked(ctx context.Context, store *factory.BacklogStore, root, runID, lane string, quotaHeld bool, holdLine string, out, errOut io.Writer, opts autoOptions) (homestate.Card, factory.BacklogItem, bool, error) {
	// Quota steering (REQ-QAS-009): a hold admits no new card — the gated
	// lease leaves only the card already assigned to this lane leasable.
	if quotaHeld {
		_, _ = fmt.Fprintln(errOut, holdLine)
		card, leased, err := factoryNextLeaseOnceGated(ctx, root, runID, lane, true)
		if err != nil || !leased {
			return card, factory.BacklogItem{}, leased, err
		}
		return card, autoAssignedItem(store, card.CardID), true, nil
	}
	// The lane's own assigned card precedes any new-candidate ranking (card
	// t1577): a quota-free pass still owes the leader-dispatched card, which
	// the queued-only candidate list below never sees. The same gated lease
	// the hold arm runs (noNewCards) leases only what is already this lane's,
	// assigned bundle heads included.
	card, leased, err := factoryNextLeaseOnceGated(ctx, root, runID, lane, true)
	if err != nil {
		return homestate.Card{}, factory.BacklogItem{}, false, err
	}
	if leased {
		return card, autoAssignedItem(store, card.CardID), true, nil
	}
	rec, err := store.LoadPure()
	if err != nil {
		return homestate.Card{}, factory.BacklogItem{}, false, err
	}
	queued, notes := autoQueuedCandidates(rec)
	for _, n := range notes {
		_, _ = fmt.Fprintln(out, n)
	}
	// The shared ranking stage (SPEC-TODO-AUTO-PRIORITY-001): the same engine
	// the operator's serial cycle selects with. It writes nothing to the
	// queue; the printed selection record is the decision record.
	res := autoRank(rec, queued, opts.landed, opts.jevRank)
	for _, line := range renderAutoSelectionRecord(res) {
		_, _ = fmt.Fprintln(out, line)
	}
	for _, it := range res.Ranked {
		card, err := factoryNextNominate(ctx, root, runID, lane, it.ID, "")
		var refusal *factoryNominateRefusal
		switch {
		case err == nil:
			return card, it, true, nil
		case errors.As(err, &refusal):
			if refusal.waitable() {
				// raced / serial-slot / quota-hold: re-ranking past the gate
				// would spin on the same ineligible candidate, so the pass
				// ends here; the caller's next invocation re-selects.
				_, _ = fmt.Fprintf(errOut, "todo --auto: %s refused (%s): %s — the pass ends here\n",
					it.ID, refusal.Token, refusal.Detail)
				return homestate.Card{}, factory.BacklogItem{}, false, nil
			}
			// A permanent keep-set refusal: the gate the old dedicated
			// --auto refusal became. Fall through to the next-ranked
			// candidate, narrating the skip (one line per fall-through).
			_, _ = fmt.Fprintf(errOut, "todo --auto: %s refused (%s): %s — next candidate\n",
				it.ID, refusal.Token, refusal.Detail)
			continue
		default:
			return homestate.Card{}, factory.BacklogItem{}, false, err
		}
	}
	return homestate.Card{}, factory.BacklogItem{}, false, nil
}

// autoAssignedItem loads the queue item of an already-assigned card for the
// dispatch directive (the directive's card field and text prefix); a read
// failure degrades to the id alone, never to an empty directive.
func autoAssignedItem(store *factory.BacklogStore, cardID string) factory.BacklogItem {
	if rec, err := store.LoadPure(); err == nil {
		for _, it := range rec.Items {
			if it.ID == cardID {
				return it
			}
		}
	}
	return factory.BacklogItem{ID: cardID}
}

// autoLaneWorkCard is the foreman contract for one leased card: the dispatch
// directive for ONE isolated in-session worker, the evidence wait, and the
// completion narration. It writes no queue byte — completion reports the
// handoff to the existing completion path, and a missed deadline leaves the
// lease to the F1 expiry machinery.
func autoLaneWorkCard(out io.Writer, card factory.BacklogItem, root string, opts autoOptions) {
	evidence := autoEvidencePath(root, card.ID)
	writeAutoDirective(out, card, evidence)
	if autoWaitForEvidence(evidence, opts) {
		_, _ = fmt.Fprintf(out, "evidence collected: %s\n", evidence)
		_, _ = fmt.Fprintf(out, "complete %s: no done written — the existing completion path (leader evidence read or moai factory complete) closes the card\n", card.ID)
		return
	}
	_, _ = fmt.Fprintf(out, "non-finding: %s worker evidence absent at deadline (%s) — no unpick; the lease stands until the F1 expiry machinery returns the card\n", card.ID, evidence)
}
