package outbox

// queue_wiring.go — the bugreport queue and its bound/attempt helpers
// (design.md section 6). The queue reuses feedback.QueueStore pointed at the
// user-scoped bugreport path; the manual flow's own queue
// (<project>/.moai/state/feedback/queue.json) is untouched.

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// BugreportQueueStore returns the QueueStore over the user-scoped bugreport
// queue. Exported: the CLI's participation commands and the publish package
// (M5) resolve the same store through it — one queue, one constructor.
//
// @MX:ANCHOR: [AUTO] BugreportQueueStore — every queue-touching surface resolves the store here
// @MX:REASON: a second constructor could point at a different path, splitting the queue the caps and the sender see (REQ-ANON-013)
func BugreportQueueStore() *feedback.QueueStore {
	path, err := StorePath(QueueFileName)
	if err != nil {
		// Unreachable in practice (MoaiHome resolution); a store over an
		// empty path fails at first use with a clear error.
		path = ""
	}
	return feedback.NewQueueStore(path)
}

// EnforceQueueBound drops the OLDEST items beyond the queue bound and
// returns how many were dropped, each recorded in the outbox log (design:
// "drop an item after the attempt limit, with every constant defined in
// internal/config/defaults.go" — the bound is the same enforcement family).
func EnforceQueueBound() int {
	store := BugreportQueueStore()
	dropped := 0
	_ = store.Mutate(func(rec *feedback.QueueRecord) error {
		for len(rec.Items) > config.DefaultBugreportQueueBound {
			oldest := rec.Items[0]
			rec.Items = rec.Items[1:]
			dropped++
			_ = AppendOutbox(OutboxRow{
				Outcome: "dropped",
				Reason:  fmt.Sprintf("queue bound of %d reached; oldest item %s dropped", config.DefaultBugreportQueueBound, oldest.ID),
			})
		}
		return nil
	})
	return dropped
}

// AttemptLimitReached reports whether a queue item has exhausted its send
// attempts; the sender drops it with a log row when true.
func AttemptLimitReached(item feedback.QueueItem) bool {
	return item.Attempts >= config.DefaultBugreportAttemptLimit
}
