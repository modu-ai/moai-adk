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

// ErrNoDeliveryID marks an event without a delivery id. Without one a redelivery cannot be told from a new
// event, so it is refused (and therefore not acknowledged).
var ErrNoDeliveryID = errors.New("docstore: the webhook event carries no delivery id")

// queueRetries bounds the compare-and-swap rounds of one queue write. It matches claim.UpdateDoc's bound.
const queueRetries = 10

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

// update runs fn on the current queue and stores the result by compare-and-swap, rereading on a conflict.
// fn returns false to say nothing needs writing.
func (q *Queue) update(fn func(*queueDoc) (write bool, err error)) error {
	for try := 0; try < queueRetries; try++ {
		qd, rev, err := q.read()
		if err != nil {
			return err
		}
		write, err := fn(&qd)
		if err != nil || !write {
			return err
		}
		raw, err := json.Marshal(qd)
		if err != nil {
			return err
		}
		if len(raw) > q.limit() {
			return fmt.Errorf("%w: %d bytes against a limit of %d", ErrQueueFull, len(raw), q.limit())
		}
		_, err = q.Docs.PutDoc(q.Kind, q.name(), rev, raw)
		if errors.Is(err, core.ErrRevisionConflict) {
			continue
		}
		return err
	}
	return fmt.Errorf("%w: the queue kept moving", core.ErrBusy)
}

// Enqueue stores one pushed event. It returns nil only once the event is durably in the queue (or was
// already there); any error means it is not, and the host must not be acknowledged.
func (q *Queue) Enqueue(id string, ch forge.Change) error {
	if id == "" {
		return ErrNoDeliveryID
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
