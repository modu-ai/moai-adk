package outbox

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// PurgeStores removes every user-scoped participation store: the queue, the
// capture spool, the dedupe ledger, and the outbox log (REQ-ANON-021's
// purge arm). Absent files are fine — purge is idempotent. No network
// request and no model call: this package has neither.
func PurgeStores() error {
	bugreport.ClearSpool() // absent is fine
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
