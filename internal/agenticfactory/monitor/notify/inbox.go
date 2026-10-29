package notify

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
)

// Inbox is the channel to the agents and to the leader: a durable queue document per addressee in the
// claim store, written before Notify or Escalate return and removed only when the addressee acknowledges
// it. A message can be missed; the queue is the record of it. Nothing here sends anything anywhere: the
// addressee reads its own inbox (Notices, Escalations) and acknowledges (Ack*).
//
// It uses the same documents as the webhook queue (the claim store's collector kind, one document per
// addressee), so there is no new document kind.
type Inbox struct {
	Docs docstore.Docs
	Kind string
	// MaxDocBytes, SeenLimit and MaxIDBytes configure each addressee's queue (see docstore.Queue). They
	// come from configuration. MaxIDBytes must be at least the length of the ids below (32).
	MaxDocBytes int
	SeenLimit   int
	MaxIDBytes  int
}

var (
	_ Notifier  = Inbox{}
	_ Escalator = Inbox{}
)

func (in Inbox) queue(box, addressee string) (*docstore.Queue, error) {
	if addressee == "" {
		return nil, errors.New("notify: an inbox message needs an addressee")
	}
	q := &docstore.Queue{Docs: in.Docs, Kind: in.Kind, Key: "inbox/" + box + "/" + addressee,
		MaxDocBytes: in.MaxDocBytes, SeenLimit: in.SeenLimit, MaxIDBytes: in.MaxIDBytes}
	return q, q.Validate()
}

// messageID is a stable id for a message: the same message sent twice (a retry after a crash) is the
// same id, so the queue keeps it once.
func messageID(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:16])
}

// Notify queues the notice for its holder.
func (in Inbox) Notify(_ context.Context, n Notice) error {
	q, err := in.queue("notice", n.Holder)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(n)
	if err != nil {
		return err
	}
	return q.Put(messageID(raw), raw)
}

// Escalate queues the escalation for its target.
func (in Inbox) Escalate(_ context.Context, e Escalation) error {
	q, err := in.queue("escalation", e.To)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return q.Put(messageID(raw), raw)
}

// Received is a queued message with the id to acknowledge it by.
type Received[T any] struct {
	ID   string
	Item T
}

func receive[T any](q *docstore.Queue) ([]Received[T], error) {
	entries, err := q.Entries()
	if err != nil {
		return nil, err
	}
	out := make([]Received[T], 0, len(entries))
	for _, e := range entries {
		var item T
		if err := json.Unmarshal(e.Data, &item); err != nil {
			return nil, fmt.Errorf("notify: inbox message %s is not readable: %w", e.ID, err)
		}
		out = append(out, Received[T]{ID: e.ID, Item: item})
	}
	return out, nil
}

// Notices lists the notices waiting for holder, oldest first, without removing them.
func (in Inbox) Notices(holder string) ([]Received[Notice], error) {
	q, err := in.queue("notice", holder)
	if err != nil {
		return nil, err
	}
	return receive[Notice](q)
}

// AckNotices removes the notices holder has taken in.
func (in Inbox) AckNotices(holder string, ids []string) error {
	q, err := in.queue("notice", holder)
	if err != nil {
		return err
	}
	return q.Commit(ids)
}

// Escalations lists the escalations waiting for to, oldest first, without removing them.
func (in Inbox) Escalations(to string) ([]Received[Escalation], error) {
	q, err := in.queue("escalation", to)
	if err != nil {
		return nil, err
	}
	return receive[Escalation](q)
}

// AckEscalations removes the escalations the target has taken in.
func (in Inbox) AckEscalations(to string, ids []string) error {
	q, err := in.queue("escalation", to)
	if err != nil {
		return err
	}
	return q.Commit(ids)
}
