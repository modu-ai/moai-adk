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
	head, err := factoryBundleRecord(ctx, db, runID, members, lane, now)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "bundle %s loaded: %s assigned to %s, %d member(s) chained\n", bundleID, head.CardID, lane, len(cards))
	return nil
}

// factoryBundleRecord is the record step of the bundle load — the seam the
// lock-hold test drives. The default is one RecordBundleChain transaction.
var factoryBundleRecord = func(ctx context.Context, db *homestate.FactoryDB, runID string, members []homestate.BundleMemberSpec, lane string, now time.Time) (homestate.Card, error) {
	return db.RecordBundleChain(ctx, runID, members, lane, "bundle", now)
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
func factoryHubChainFields(queueRec *factory.BacklogRecord, cards []homestate.Card, cardID string) homestate.CardFields {
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
	var tail *string
	for i := range queueRec.Items {
		it := &queueRec.Items[i]
		if it.ID == cardID || it.Issuance == nil {
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
