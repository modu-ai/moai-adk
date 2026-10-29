package docstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// ErrQueueFull marks an event that cannot be queued without breaking the document limit. The event is not
// queued and not acknowledged, so the host redelivers it later; nothing is dropped to make room.
var ErrQueueFull = errors.New("docstore: the webhook queue is full")

// ErrQueueConfig marks a queue whose configuration cannot work (for example a memory of consumed ids that
// could fill the document and leave no room for pending events).
var ErrQueueConfig = errors.New("docstore: the webhook queue is misconfigured")

// ErrBadDeliveryID marks a delivery id longer than the configured bound.
var ErrBadDeliveryID = errors.New("docstore: the webhook delivery id is longer than allowed")

// ErrNoDeliveryID marks an event without a delivery id. Without one a redelivery cannot be told from a new
// event, so it is refused (and therefore not acknowledged).
var ErrNoDeliveryID = errors.New("docstore: the webhook event carries no delivery id")

// Queue is the durable queue of pushed (webhook) events of one collector, kept in one claim document. An
// event is acknowledged to the host only after Enqueue returned nil, i.e. after the document write
// succeeded; the collector removes an event (Commit) only after the consumer acknowledged it. Both
// halves survive a restart, so a crash loses no event.
//
// Delivery ids make redelivery harmless: an id already pending, or among the most recently committed
// ones, is acknowledged again without queueing it twice.
type Queue struct {
	Docs Docs
	Kind string
	Key  string
	// MaxDocBytes is the largest stored document. Zero means DefaultDocLimit.
	MaxDocBytes int
	// SeenLimit is how many committed delivery ids are remembered for deduplication. Zero remembers none:
	// an id is then only deduplicated while it is pending. It is configuration, not a constant.
	SeenLimit int
	// MaxIDBytes is the longest delivery id accepted (GitHub's are 36-character UUIDs). It bounds what the
	// consumed-id memory can cost, so it is required when SeenLimit is set.
	MaxIDBytes int
	// Alive reports whether the push channel is healthy (the collector polls only as a safety net then).
	Alive func() bool
}

var _ forge.DurableFeed = (*Queue)(nil)

type queuedEvent struct {
	ID   string           `json:"id"`
	Key  string           `json:"key"`
	Kind forge.ChangeKind `json:"kind"`
}

type queueDoc struct {
	Key    string        `json:"key"`
	Events []queuedEvent `json:"events"`
	Seen   []string      `json:"seen,omitempty"`
}

func (q *Queue) limit() int {
	if q.MaxDocBytes > 0 {
		return q.MaxDocBytes
	}
	return DefaultDocLimit
}

// Validate refuses a configuration in which the remembered consumed ids alone could take more than half
// of the document: pending events must always have at least as much room as the memory of consumed ones,
// or a full memory would turn every delivery into ErrQueueFull.
func (q *Queue) Validate() error {
	if q.SeenLimit < 0 || q.MaxIDBytes < 0 {
		return fmt.Errorf("%w: negative limit", ErrQueueConfig)
	}
	if q.SeenLimit == 0 {
		return nil
	}
	if q.MaxIDBytes == 0 {
		return fmt.Errorf("%w: SeenLimit is set but MaxIDBytes is not", ErrQueueConfig)
	}
	if seen := q.SeenLimit * (q.MaxIDBytes + 3); seen*2 > q.limit() { // quotes and a comma per id
		return fmt.Errorf("%w: %d remembered ids of up to %d bytes can take %d of the %d bytes of the document",
			ErrQueueConfig, q.SeenLimit, q.MaxIDBytes, seen, q.limit())
	}
	return nil
}

func (q *Queue) name() string { return docName("queue\x00" + q.Key) }

func (q *Queue) read() (qd queueDoc, rev int64, err error) {
	if q.Kind == "" {
		return qd, 0, errors.New("docstore: the queue document kind is not configured")
	}
	d, err := q.Docs.GetDoc(q.Kind, q.name())
	if errors.Is(err, core.ErrNotHeld) {
		return queueDoc{Key: q.Key}, 0, nil
	}
	if err != nil {
		return qd, 0, err
	}
	if err := json.Unmarshal(d.Data, &qd); err != nil {
		return qd, 0, fmt.Errorf("docstore: the queue document is not readable: %w", err)
	}
	if qd.Key != q.Key {
		return qd, 0, fmt.Errorf("docstore: the queue document holds key %q, not %q", qd.Key, q.Key)
	}
	return qd, d.Rev, nil
}

// update runs fn on the current queue and stores the result by compare-and-swap; on a claim store a
// stale cache is repaired from the authority first (see modify). fn returns false to say nothing needs
// writing.
func (q *Queue) update(fn func(*queueDoc) (write bool, err error)) error {
	if q.Kind == "" {
		return errors.New("docstore: the queue document kind is not configured")
	}
	if err := q.Validate(); err != nil {
		return err
	}
	return modify(q.Docs, q.Kind, q.name(), func(cur []byte) ([]byte, error) {
		qd := queueDoc{Key: q.Key}
		if cur != nil {
			if err := json.Unmarshal(cur, &qd); err != nil {
				return nil, fmt.Errorf("docstore: the queue document is not readable: %w", err)
			}
			if qd.Key != q.Key {
				return nil, fmt.Errorf("docstore: the queue document holds key %q, not %q", qd.Key, q.Key)
			}
		}
		write, err := fn(&qd)
		if err != nil {
			return nil, err
		}
		if !write {
			return nil, errNoWrite
		}
		raw, err := json.Marshal(qd)
		if err != nil {
			return nil, err
		}
		if len(raw) > q.limit() {
			return nil, fmt.Errorf("%w: %d bytes against a limit of %d", ErrQueueFull, len(raw), q.limit())
		}
		return raw, nil
	})
}

// Enqueue stores one pushed event. It returns nil only once the event is durably in the queue (or was
// already there); any error means it is not, and the host must not be acknowledged.
func (q *Queue) Enqueue(id string, ch forge.Change) error {
	if id == "" {
		return ErrNoDeliveryID
	}
	if q.MaxIDBytes > 0 && len(id) > q.MaxIDBytes {
		return fmt.Errorf("%w: %d bytes against %d", ErrBadDeliveryID, len(id), q.MaxIDBytes)
	}
	return q.update(func(qd *queueDoc) (bool, error) {
		if slices.Contains(qd.Seen, id) || slices.ContainsFunc(qd.Events, func(e queuedEvent) bool { return e.ID == id }) {
			return false, nil
		}
		qd.Events = append(qd.Events, queuedEvent{ID: id, Key: ch.Key, Kind: ch.Kind})
		return true, nil
	})
}

// Accept is the receiver's entry: it queues the event and calls ack — the host acknowledgement — only
// after the queue write succeeded. On any error ack is not called.
func (q *Queue) Accept(id string, ch forge.Change, ack func()) error {
	if err := q.Enqueue(id, ch); err != nil {
		return err
	}
	ack()
	return nil
}

func (q *Queue) Healthy() bool { return q.Alive != nil && q.Alive() }

// Pending lists the queued events, oldest first, without removing them.
func (q *Queue) Pending() ([]forge.PendingChange, error) {
	qd, _, err := q.read()
	if err != nil {
		return nil, err
	}
	out := make([]forge.PendingChange, 0, len(qd.Events))
	for _, e := range qd.Events {
		out = append(out, forge.PendingChange{ID: e.ID, Change: forge.Change{Key: e.Key, Kind: e.Kind}})
	}
	return out, nil
}

// Commit removes the events the consumer acknowledged and remembers their ids for deduplication.
func (q *Queue) Commit(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return q.update(func(qd *queueDoc) (bool, error) {
		var removed []string
		qd.Events = slices.DeleteFunc(qd.Events, func(e queuedEvent) bool {
			if slices.Contains(ids, e.ID) {
				removed = append(removed, e.ID)
				return true
			}
			return false
		})
		if len(removed) == 0 {
			return false, nil
		}
		if q.SeenLimit > 0 {
			qd.Seen = append(qd.Seen, removed...)
			if over := len(qd.Seen) - q.SeenLimit; over > 0 {
				qd.Seen = qd.Seen[over:]
			}
		}
		return true, nil
	})
}
