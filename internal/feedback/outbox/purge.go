package outbox

import (
	"context"
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
	if err := bugreport.ClearSpool(); err != nil {
		return fmt.Errorf("outbox: purge spool: %w", err)
	}
	store := BugreportQueueStore()
	err := store.MutateContext(context.Background(), func(rec *feedback.QueueRecord) error {
		rec.Items = []feedback.QueueItem{}
		return nil
	})
	if err != nil {
		return fmt.Errorf("outbox: purge queue: %w", err)
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
