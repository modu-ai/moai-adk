package outbox

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// PurgeStores removes every user-scoped participation store: the queue, the
// capture spool, the dedupe ledger, and the outbox log (REQ-ANON-021's
// purge arm). Absent files are fine — purge is idempotent, and ClearSpool
// already reports nil for an absent spool. A removal that FAILS propagates:
// purge reporting success while a store survives would leave the user
// believing data was withdrawn that is still on disk (review-gate finding
// #7). No network request and no model call: this package has neither.
func PurgeStores() error {
	if err := bugreport.ClearSpool(); err != nil {
		return fmt.Errorf("outbox: purge spool: %w", err)
	}
	for _, name := range []string{QueueFileName, LedgerFileName, OutboxFileName} {
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
