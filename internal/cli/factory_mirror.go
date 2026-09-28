package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryRecordUnavailableTag prefixes the stderr line and the status line of
// a dispatch whose factory-record mirror write failed.
const factoryRecordUnavailableTag = "FACTORY_RECORD_UNAVAILABLE"

// factoryAssignmentWriter records a dispatched card in the factory record as
// assigned to the lane. It is a seam: tests replace it to inject a write
// failure without touching file permissions (the store re-applies them).
var factoryAssignmentWriter = writeFactoryAssignment

// writeFactoryAssignment is the dispatch mirror (REQ-FR-025): T1 (when the
// card has no record) and T2 through the transition API. A card already
// assigned to the same lane is left as it is.
func writeFactoryAssignment(ctx context.Context, root string, store *kanban.BacklogStore, runID, cardID, lane string) error {
	record, err := store.LoadPure()
	if err != nil {
		return fmt.Errorf("read queue: %w", err)
	}
	picked := false
	for _, item := range record.Items {
		if item.ID == cardID {
			picked = item.State == kanban.BacklogStatePicked
		}
	}
	if !picked {
		return fmt.Errorf("queue item %s is not picked", cardID)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	now := factoryCardNow()
	card, err := db.LoadCard(ctx, runID, cardID)
	if errors.Is(err, homestate.ErrCardNotFound) {
		card, err = db.RecordPicked(ctx, runID, cardID, homestate.CardFields{}, "dispatch", now)
	}
	if err != nil {
		return err
	}
	switch {
	case card.State == homestate.CardPicked:
		_, err = db.Transition(ctx, homestate.TransitionRequest{RunID: runID, CardID: cardID, To: homestate.CardAssigned, ExpectedVersion: card.Version, Actor: "dispatch", Owner: lane, Now: now})
		return err
	case card.State == homestate.CardAssigned && card.OwnerLabel == lane:
		return nil
	default:
		return fmt.Errorf("factory record for %s is %s (owner %q), not assignable to %s", cardID, card.State, card.OwnerLabel, lane)
	}
}

// mirrorFactoryAssignment writes the factory record for a dispatch that has
// already been recorded in the queue runtime report. It never fails the
// dispatch: a failed write leaves a FACTORY_RECORD_UNAVAILABLE line on stderr
// and an entry in the unavailable-record log, which `moai factory status`
// reports and the next successful write for the run reconciles.
//
// @MX:NOTE: [AUTO] fail-open by lead decision — the dispatch and its queue runtime row stand even when factory.db cannot be written
func mirrorFactoryAssignment(ctx context.Context, stderr io.Writer, root string, store *kanban.BacklogStore, runID, cardID, lane string) {
	err := factoryAssignmentWriter(ctx, root, store, runID, cardID, lane)
	if err == nil {
		return
	}
	_, _ = fmt.Fprintf(stderr, "%s run=%s card=%s lane=%s: %v\n", factoryRecordUnavailableTag, runID, cardID, lane, err)
	entry := homestate.RecordUnavailableEntry{RunID: runID, CardID: cardID, Lane: lane, Error: err.Error()}
	if logErr := homestate.AppendRecordUnavailable(root, entry); logErr != nil {
		_, _ = fmt.Fprintf(stderr, "%s: the unavailable-record log could not be written either: %v\n", factoryRecordUnavailableTag, logErr)
	}
}
