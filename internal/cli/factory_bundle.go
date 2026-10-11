// factory_bundle.go — SPEC-TODO-CARD-ISSUANCE-001 M4 (REQ-TCI-018/-020):
// the bundle loader and the hub-chain hint. The bundle is the operator's
// loading act: it records the members in order with the bundle identity and
// position, chains each later member to the one before it through the after
// hint, and assigns ONLY the first member to the lane — the selection host
// serves the rest to that lane as the predecessors reach the local merge.
package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// newFactoryBundleCommand — `moai factory bundle <lane> <card>…` (REQ-TCI-018):
// load a bundle chain as one leader-side act. Every member's queue item must
// already be picked — the same precondition RecordPicked states
// (REQ-FR-022) — and one member missing it refuses the whole bundle: a
// half-loaded chain would lease its head and strand the rest.
func newFactoryBundleCommand() *cobra.Command {
	var run string
	cmd := &cobra.Command{
		Use:   "bundle <lane> <card>…",
		Short: "Load a bundle chain: record the members in order, assign the first to the lane",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			lane := strings.TrimSpace(args[0])
			cards := args[1:]
			if err := runFactoryBundle(cmd, lane, cards, run); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: factory bundle: %v\n", err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// runFactoryBundle is the bundle body: verify every member's queue
// precondition AND record the chain inside ONE critical section (card t1454
// card-review r2d finding D1). Releasing the lock between the picked check
// and the record left a window another session's unpick fit through — the
// check passed, the member left `picked`, and the load recorded a member
// whose precondition no longer held. The section's contract holds here: the
// queue is read only through the locked handle, and nothing below starts a
// git process or creates a worktree (REQ-FAL-007) — the record writes are
// sqlite only.
func runFactoryBundle(cmd *cobra.Command, lane string, cards []string, run string) error {
	if lane == "" {
		return fmt.Errorf("the lane label is required")
	}
	root := factoryCardRoot()
	store := newTodoStore()
	var loadErr error
	ran, err := factoryLeaseSection(store, func(l *factory.LockedBacklog) {
		loadErr = runFactoryBundleLocked(cmd, l, root, lane, cards, run)
	})
	if err != nil {
		return err
	}
	if !ran {
		return fmt.Errorf("the queue lock stayed held for the whole wait budget, so the bundle was not loaded; retry")
	}
	return loadErr
}

// runFactoryBundleLocked is the verify→record body with the queue lock held.
func runFactoryBundleLocked(cmd *cobra.Command, l *factory.LockedBacklog, root, lane string, cards []string, run string) error {
	rec, err := l.LoadPure()
	if err != nil {
		return fmt.Errorf("read the queue: %w", err)
	}
	byID := make(map[string]factory.BacklogItem, len(rec.Items))
	for _, it := range rec.Items {
		byID[it.ID] = it
	}
	for _, id := range cards {
		it, ok := byID[id]
		if !ok {
			return fmt.Errorf("no card %s in the queue", id)
		}
		// POSITIVE enumeration (REQ-THS-012): picked is the only state
		// the loader takes a member from; every other state — a state
		// added later included — falls to the refusal.
		switch it.State {
		case factory.BacklogStatePicked:
			// the only state the loader takes a member from
		default:
			return fmt.Errorf("%s is %s, not picked — pick every member first", id, it.State)
		}
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	runID, err := resolveFactoryCardRun(ctx, root, run)
	if err != nil {
		return err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	rows, err := db.ListCards(ctx, runID)
	if err != nil {
		return fmt.Errorf("read the records for the chain check: %w", err)
	}
	if err := factoryRefuseForeignChain(rows, lane, cards); err != nil {
		return err
	}
	now := factoryCardNow()
	bundleID := fmt.Sprintf("bundle-%s-%s", cards[0], now.UTC().Format("20060102T150405"))
	// The chain records and the head's assignment land inside ONE record
	// transaction (card t1454 card-review r2 finding 3): a member whose
	// record has already moved past `picked` aborts the whole load, leaving
	// no half-loaded chain behind.
	members := make([]homestate.BundleMemberSpec, len(cards))
	for i, id := range cards {
		members[i] = homestate.BundleMemberSpec{CardID: id, BundleID: bundleID, Order: i}
		if i > 0 {
			members[i].HintAfter = cards[i-1]
		}
	}
	// The head carries the hub-chain condition too (card t1533, review-gate
	// r2 finding a): only the second and later members carried an after, so
	// a head whose files cross a hub path was assigned with no conflict check
	// against another lane's same-hub work. The generated hint follows the
	// one record rule — a stored hint wins, a creation fills — and an
	// unmerged sharer refuses the load through the head assignment's T2
	// guard, so the conflict check runs before anything is recorded.
	var headRow *homestate.Card
	for i := range rows {
		if rows[i].CardID != cards[0] {
			continue
		}
		headRow = &rows[i]
		break
	}
	// The head's hub candidates exclude the bundle's own members (card
	// t1533, review-gate r8): a member is ordered by the bundle, and a head
	// waiting on its own follower refused the load with a dependency
	// opposite to the explicit order.
	memberSet := make(map[string]bool, len(cards))
	for _, id := range cards {
		memberSet[id] = true
	}
	if hf := factoryGeneratedHubFields(factoryHubChainFields(rec, rows, cards[0], memberSet), headRow); hf.HintAfter != nil {
		members[0].HintAfter = *hf.HintAfter
	}
	// The head's T2 re-points its dispatch binding onto this run inside the
	// chain's transaction, so the queue's current-dispatch record — written
	// under this held lock — follows inside it, right before the commit (turn-end
	// gate, card t1538): an unwritable record rolls the whole load back.
	head, err := factoryBundleRecord(ctx, db, runID, members, lane, now, func(head homestate.Card) error {
		return l.RefreshDispatchCurrent(head.CardID, runID, head.OwnerLabel)
	})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "bundle %s loaded: %s assigned to %s, %d member(s) chained\n", bundleID, head.CardID, lane, len(cards))
	return nil
}

// factoryBundleRecord is the record step of the bundle load — the seam the
// lock-hold test drives. The default is one RecordBundleChain transaction.
var factoryBundleRecord = func(ctx context.Context, db *homestate.FactoryDB, runID string, members []homestate.BundleMemberSpec, lane string, now time.Time, beforeCommit func(head homestate.Card) error) (homestate.Card, error) {
	return db.RecordBundleChain(ctx, runID, members, lane, "bundle", now, beforeCommit)
}

// factoryRefuseForeignChain refuses loading a card whose factory record
// already belongs to another lane's work (card t1533, card-review r2f
// finding 1): a bundle follow-up member sits at `picked` with no owner while
// it waits for its chain head, so the picked-state check alone admitted a
// re-bundle that re-chained the member under the loading lane and assigned
// it away from its own chain. A lane never mutates another lane's assignment
// or bundle (the lane-obligation axis): a member already carrying a bundle
// identity is refused unless the chain's recorded owner IS the loading lane,
// and a member owned outright by another lane is refused the same way. The
// read runs inside the load's lock-held section, so the check and the record
// share one exclusion.
func factoryRefuseForeignChain(rows []homestate.Card, lane string, cards []string) error {
	rowOf := make(map[string]homestate.Card, len(rows))
	chainOwner := make(map[string]string, len(rows))
	for _, c := range rows {
		rowOf[c.CardID] = c
		if owner := strings.TrimSpace(c.OwnerLabel); c.BundleID != "" && owner != "" {
			if _, seen := chainOwner[c.BundleID]; !seen {
				chainOwner[c.BundleID] = owner
			}
		}
	}
	for _, id := range cards {
		row, ok := rowOf[id]
		if !ok {
			continue // no record yet: the load creates it, nothing to steal
		}
		if row.BundleID != "" {
			owner := chainOwner[row.BundleID]
			if owner == lane {
				continue // the lane re-loads its own chain
			}
			if owner == "" {
				return fmt.Errorf("%s is already a member of bundle %s", id, row.BundleID)
			}
			return fmt.Errorf("%s is already %s's bundle member", id, owner)
		}
		if owner := strings.TrimSpace(row.OwnerLabel); owner != "" && owner != lane {
			return fmt.Errorf("%s is assigned to %s; a lane never takes another lane's card", id, owner)
		}
	}
	return nil
}

// factoryHubChainFields computes the hub-chain hint for a card about to be
// recorded (REQ-TCI-020): when the card's recorded files cross the embedded
// hub list and another open, RECORDED card's files share A HUB PATH WITH THE
// CANDIDATE — the intersection of the two hub crossings, not each card
// crossing some hub path of its own (card t1454 card-review r2 finding 6) —
// the new record's after hint names the first such card in queue order. The
// queue read supplies the files attributes and homestate the embedded list —
// this is the one place that sees both (design §7.2). A card with no files,
// no crossing, or no chainable predecessor carries no hint. Keep-set and
// selection read no file overlap: the hint is a RECORD-CREATION input only.
// exclude names cards that are never candidates — the bundle load passes its
// own member set, because a member is ordered by the bundle and a head made
// to wait on its follower refused the load with a dependency opposite to the
// explicit order (card t1533, review-gate r8) — and, by the same rule's
// general form (review-gate r16), a sharer the candidate's stored relations
// order BEHIND the candidate is never a candidate either.
func factoryHubChainFields(queueRec *factory.BacklogRecord, cards []homestate.Card, cardID string, exclude map[string]bool) homestate.CardFields {
	hub := make(map[string]bool)
	for _, p := range homestate.HubFiles() {
		hub[p] = true
	}
	var candidate *factory.BacklogItem
	for i := range queueRec.Items {
		if queueRec.Items[i].ID == cardID {
			candidate = &queueRec.Items[i]
			break
		}
	}
	if candidate == nil || candidate.Issuance == nil {
		return homestate.CardFields{}
	}
	candHub := make(map[string]bool)
	for _, f := range candidate.Issuance.Files {
		if hub[f] {
			candHub[f] = true
		}
	}
	if len(candHub) == 0 {
		return homestate.CardFields{}
	}
	recorded := make(map[string]bool, len(cards))
	for _, c := range cards {
		recorded[c.CardID] = true
	}
	deps := factoryHubDependencies(queueRec, cards, factoryMergedCards(cards))
	var tail *string
	for i := range queueRec.Items {
		it := &queueRec.Items[i]
		if it.ID == cardID || it.Issuance == nil || exclude[it.ID] {
			continue
		}
		// The generated edge never reverses or closes an existing after
		// relation (card t1533, review-gate r16 — the generation side of the
		// r14 rule): a sharer whose combined dependency path reaches the candidate is
		// ordered BEHIND it, so naming it as the candidate's predecessor
		// stored the exact reversal and the cycle went into the record with
		// the row.
		if factoryDependencyReaches(deps, it.ID, cardID) {
			continue
		}
		shares := false
		for _, f := range it.Issuance.Files {
			if candHub[f] {
				shares = true
				break
			}
		}
		if !shares {
			continue
		}
		// POSITIVE enumeration (REQ-THS-012): the open states a chain can
		// order behind; every other state — a state added later included —
		// falls through.
		switch it.State {
		case factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStateHold:
		default:
			continue
		}
		// The T2 after guard needs the predecessor's factory record; a card
		// never dispatched carries no record to reach the local merge with.
		if !recorded[it.ID] {
			continue
		}
		// The chain's tail is the predecessor: the candidate waits behind
		// the LAST open card sharing the hub path, not the first — naming
		// the first let a new candidate lease straight past the card still
		// in flight (card t1454 card-review r2b finding B1).
		after := it.ID
		tail = &after
	}
	if tail == nil {
		return homestate.CardFields{}
	}
	return homestate.CardFields{HintAfter: tail}
}

// factoryGeneratedHubFields merges the computed hub hint into the
// record-creation fields for a card that may already carry a factory row
// (card t1533, card-review r2f ledger): a GENERATED hint is a
// record-CREATION input only — written ONLY when the card has no row at
// all, NEVER into an existing row, whether that row holds a hint or an
// empty one (review-gate r6: filling an empty row recomputed the tail
// across the chain and minted a t1→t2→t1 cycle that outlived the failed
// lease). The stored hint survives every subsequent write; an explicit
// input (--after, a bundle member hint) lands through its own path and
// outranks the fill.
func factoryGeneratedHubFields(fields homestate.CardFields, row *homestate.Card) homestate.CardFields {
	if row != nil {
		fields.HintAfter = nil
	}
	return fields
}

// factoryMergedCards reduces recorded rows to the set of cards whose
// implementation pipeline has reached the merge — the selector's completion
// set and the nominated validation's wait check read the SAME set, so the
// two never drift (card t1533).
func factoryMergedCards(rows []homestate.Card) map[string]bool {
	merged := make(map[string]bool, len(rows))
	for _, c := range rows {
		switch c.State {
		// merged-pr belongs here beside merged-local: the T2 guard
		// (predecessorMerged) accepts it, and a selector that did not made a
		// github-flow predecessor release nothing — the follower answered no
		// card forever (card t1533, review-gate r2 finding b).
		case homestate.CardMergedLocal, homestate.CardMergedPR, homestate.CardPushed, homestate.CardCIGreen, homestate.CardDone:
			merged[c.CardID] = true
		}
	}
	return merged
}

// factoryHubDependencies gives stored relations priority over inferred hub waits.
// Inferred edges are added in queue order, with in-flight predecessors first;
// an edge that would close a combined after/hub cycle is never added.
// @MX:NOTE: Stored after and bundle order outrank inferred waits; actual
// in-flight hub predecessors outrank idle queue ordering.
func factoryHubDependencies(queueRec *factory.BacklogRecord, cards []homestate.Card, merged map[string]bool) map[string][]string {
	deps := make(map[string][]string, len(cards))
	for _, c := range cards {
		if c.HintAfter != "" {
			deps[c.CardID] = append(deps[c.CardID], c.HintAfter)
		}
		for _, pred := range cards {
			if c.BundleID != "" && c.BundleID == pred.BundleID && pred.BundleOrder < c.BundleOrder {
				deps[c.CardID] = append(deps[c.CardID], pred.CardID)
			}
		}
	}
	if queueRec == nil {
		return deps
	}
	inFlight := make(map[string]bool, len(cards))
	for _, c := range cards {
		inFlight[c.CardID] = homestate.IsLeaseHoldingState(c.State)
	}
	for _, flightFirst := range []bool{true, false} {
		for _, candidate := range queueRec.Items {
			for _, pred := range factoryHubWaitCandidates(queueRec, cards, merged, candidate.ID) {
				if inFlight[pred] != flightFirst || factoryDependencyReaches(deps, pred, candidate.ID) {
					continue
				}
				deps[candidate.ID] = append(deps[candidate.ID], pred)
			}
		}
	}
	return deps
}

// @MX:ANCHOR: [AUTO] Bounded traversal of combined after, bundle, and hub dependencies.
// @MX:REASON: Generation, inferred-edge insertion, and selection share the same cycle boundary.
func factoryDependencyReaches(deps map[string][]string, start, target string) bool {
	seen := make(map[string]bool, len(deps))
	pending := []string{start}
	for len(pending) > 0 {
		cur := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if cur == target {
			return true
		}
		if !seen[cur] {
			seen[cur] = true
			pending = append(pending, deps[cur]...)
		}
	}
	return false
}

// factoryHubWaitUnmerged reports whether cardID still waits behind an
// unmerged hub-chain predecessor, naming the blocking card: any recorded,
// open queue card whose files share a hub path with the candidate and whose
// record has not reached the merge — queue-earlier sharers in any open
// state, queue-LATER sharers while actually in flight (review-gate r10).
// The stored hint names one predecessor — the tail at the card's own record
// creation — but a candidate whose files cross several hub paths has a
// predecessor per hub path, and waiting on the named one alone leased the
// candidate beside the still-in-flight sharer of its other hub path. The
// queue read supplies the files attributes, the recorded rows the wait
// candidates, mergedLocal the merge states — the same inputs
// factoryHubChainFields reads, and like it a read-only predicate: selection
// consults it on every pass, it writes nothing.
func factoryHubWaitUnmerged(queueRec *factory.BacklogRecord, cards []homestate.Card, mergedLocal map[string]bool, cardID string) (string, bool) {
	deps := factoryHubDependencies(queueRec, cards, mergedLocal)
	for _, pred := range factoryHubWaitCandidates(queueRec, cards, mergedLocal, cardID) {
		if !factoryDependencyReaches(deps, pred, cardID) {
			return pred, true
		}
	}
	return "", false
}

// factoryHubWaitCandidates enumerates eligible hub predecessors before combined
// dependency ordering. It retains the merge, bundle, and in-flight boundaries.
func factoryHubWaitCandidates(queueRec *factory.BacklogRecord, cards []homestate.Card, mergedLocal map[string]bool, cardID string) []string {
	if queueRec == nil {
		return nil
	}
	hub := make(map[string]bool)
	for _, p := range homestate.HubFiles() {
		hub[p] = true
	}
	candIdx := -1
	var candHub map[string]bool
	for i := range queueRec.Items {
		if queueRec.Items[i].ID != cardID {
			continue
		}
		candIdx = i
		if it := &queueRec.Items[i]; it.Issuance != nil {
			candHub = make(map[string]bool)
			for _, f := range it.Issuance.Files {
				if hub[f] {
					candHub[f] = true
				}
			}
		}
		break
	}
	if candIdx < 0 || len(candHub) == 0 {
		return nil
	}
	recorded := make(map[string]bool, len(cards))
	rowOf := make(map[string]homestate.Card, len(cards))
	for _, c := range cards {
		recorded[c.CardID] = true
		rowOf[c.CardID] = c
	}
	candRow := rowOf[cardID]
	var predecessors []string
	for i := range queueRec.Items {
		if i == candIdx {
			continue
		}
		it := &queueRec.Items[i]
		if it.Issuance == nil || !recorded[it.ID] || mergedLocal[it.ID] {
			continue
		}
		// POSITIVE enumeration (REQ-THS-012): the open states a wait can
		// order behind; every other state — a state added later included —
		// falls through.
		switch it.State {
		case factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStateHold:
		default:
			continue
		}
		// Same-bundle members wait by the RECORDED BUNDLE ORDER, not the
		// queue order (card t1533, review-gate r9): a bundle loaded in
		// reverse queue order made the head wait on its own follower by the
		// queue-position rule while the follower waited on the head by the
		// bundle rule, and not even the bundle's first card leased. The
		// bundle orders its members; a member EARLIER in the bundle holds
		// the candidate, a later one never does — in either queue direction.
		if candRow.BundleID != "" {
			if predRow, ok := rowOf[it.ID]; ok && predRow.BundleID == candRow.BundleID && candRow.BundleOrder < predRow.BundleOrder {
				continue
			}
		}
		// A queue-LATER sharer holds the candidate only while it is actually
		// in flight (card t1533, review-gate r10): the wait no longer scans
		// queue-earlier entries alone, and a later card still waiting its
		// turn must not invert the queue's priority.
		if i > candIdx {
			if predRow, ok := rowOf[it.ID]; !ok || !homestate.IsLeaseHoldingState(predRow.State) {
				continue
			}
		}
		for _, f := range it.Issuance.Files {
			if candHub[f] {
				predecessors = append(predecessors, it.ID)
				break
			}
		}
	}
	return predecessors
}
