package outbox

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// PurgeStores removes every user-scoped participation store: the queue, the
// capture spool, the dedupe ledger, and the outbox log (REQ-ANON-021's
// purge arm). Absent files are fine — purge is idempotent, and ClearSpool
// already reports nil for an absent spool. A removal that FAILS propagates:
// purge reporting success while a store survives would leave the user
// believing data was withdrawn that is still on disk (review-gate finding
// #7). No network request and no model call: this package has neither.
//
// The QUEUE's purge is serialized through the queue's OWN mutation lock
// (review gate finding, P2): a bare os.Remove deleted the file under a
// paused queue writer, and the writer's subsequent save RESURRECTED the
// purged report. The queue step now waits for any in-flight mutation to
// commit, empties the queue under the same lock, and only then removes the
// file — the writer has already finished, so nothing comes back.
func PurgeStores() error {
	// The generation bump comes FIRST (review gate finding, P2): a drain
	// that read its batch before this purge re-checks the generation before
	// each item, so the bump — not the file removals — is what stops it
	// from enqueueing a stale batch over the withdrawn store.
	if err := bugreport.BumpSpoolGeneration(); err != nil {
		return fmt.Errorf("outbox: purge generation: %w", err)
	}
	if err := bugreport.ClearSpool(); err != nil {
		return fmt.Errorf("outbox: purge spool: %w", err)
	}
	store := BugreportQueueStore()
	// Only an UNREADABLE queue (the mutation's load step failing to parse
	// or refusing an over-cap file) is ignored here — the user asked for
	// the store GONE, so the removal below proceeds on any content
	// failure. Every OTHER mutation failure propagates (review gate
	// finding, P2): a lock-acquire failure swallowed here would remove the
	// file under the lock's holder, and the holder's save would resurrect
	// the purged report — exactly the resurrection the lock-serialized
	// queue step exists to prevent.
	merr := store.MutateContext(context.Background(), func(rec *feedback.QueueRecord) error {
		rec.Items = []feedback.QueueItem{}
		return nil
	})
	if merr != nil && !errors.Is(merr, feedback.ErrQueueUnreadable) {
		return fmt.Errorf("outbox: purge queue: %w", merr)
	}
	queuePath, err := StorePath(QueueFileName)
	if err != nil {
		return err
	}
	if err := os.Remove(queuePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("outbox: purge queue: %w", err)
	}
	for _, name := range []string{LedgerFileName, OutboxFileName, ModelCallsFileName} {
		path, err := StorePath(name)
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
