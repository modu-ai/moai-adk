package codexbridge

import (
	"container/list"
	"sync"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

// eventQueue is a byte-bounded mailbox. The transport reader must never wait
// for a tool's HTTP consumer: that same transport carries the tool's RPC reply.
// Accounting includes conservative allocation overhead, not just JSON bytes.
type eventQueue struct {
	mu           sync.Mutex
	events       list.List
	bytes, limit int
	ready        chan struct{}
}

func newEventQueue(limit int) *eventQueue {
	// Small output limits still need room for protocol envelopes and tools.
	if limit < 64<<10 {
		limit = 64 << 10
	}
	return &eventQueue{limit: limit, ready: make(chan struct{}, 1)}
}

func eventBytes(event codexapp.Message) int {
	return len(event.ID) + len(event.Method) + len(event.Params) + 256
}

func (q *eventQueue) push(event codexapp.Message) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	cost := eventBytes(event)
	if cost > q.limit-q.bytes {
		return false
	}
	q.events.PushBack(event)
	q.bytes += cost
	select {
	case q.ready <- struct{}{}:
	default:
	}
	return true
}

func (q *eventQueue) pop() (codexapp.Message, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	first := q.events.Front()
	if first == nil {
		return codexapp.Message{}, false
	}
	event := q.events.Remove(first).(codexapp.Message)
	q.bytes -= eventBytes(event)
	if q.events.Len() > 0 {
		select {
		case q.ready <- struct{}{}:
		default:
		}
	}
	return event, true
}

type eventQueueLimitError struct{}

func (eventQueueLimitError) Error() string {
	return "app server bridge event queue byte limit exceeded"
}
func (eventQueueLimitError) Unwrap() error     { return ErrLimit }
func (eventQueueLimitError) CauseCode() string { return "appserver_event_queue_bytes_exceeded" }

// These notifications never contributed public text, tool results, terminal
// state or usage. Do not enqueue potentially unbounded reasoning/progress while
// Claude is executing a tool. Unknown notifications and every RPC request stay
// ordered and reach the protocol validator.
func informationalEvent(event codexapp.Message) bool {
	if len(event.ID) > 0 {
		return false
	}
	switch event.Method {
	case "item/reasoning/textDelta", "item/reasoning/summaryTextDelta", "item/reasoning/summaryPartAdded", "thread/status/changed", "item/started", "item/completed":
		return true
	}
	return false
}
