package docstore

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

func newQueue(d Docs) *Queue {
	return &Queue{Docs: d, Kind: CollectorKind, Key: "o/r/hooks", SeenLimit: 4, MaxIDBytes: 36, Alive: func() bool { return true }}
}

var hook = forge.Change{Key: "7", Kind: forge.ChangeModified}

func pendingIDs(t *testing.T, q *Queue) []string {
	t.Helper()
	p, err := q.Pending()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, e := range p {
		ids = append(ids, e.ID)
	}
	return ids
}

// The host acknowledgement is sent only after the queue write succeeded.
func TestTheHostIsAcknowledgedOnlyAfterTheQueueWriteSucceeded(t *testing.T) {
	dir := t.TempDir()
	fd := openFault(t, dir)
	q := newQueue(fd)
	acked := 0
	fd.arm(1, 0) // the queue write is refused
	if err := q.Accept("d1", hook, func() { acked++ }); !errors.Is(err, errCrash) || acked != 0 {
		t.Fatalf("err=%v acked=%d, want the write error and no acknowledgement", err, acked)
	}
	if ids := pendingIDs(t, newQueue(openStore(t, dir))); len(ids) != 0 {
		t.Errorf("a refused write left %v queued", ids)
	}
	fd.arm(0, 0)
	if err := q.Accept("d1", hook, func() { acked++ }); err != nil || acked != 1 {
		t.Fatalf("err=%v acked=%d, want one acknowledgement after a good write", err, acked)
	}
	if ids := pendingIDs(t, q); len(ids) != 1 {
		t.Errorf("pending = %v", ids)
	}
}

// The write lands, then the process dies before the acknowledgement: the host redelivers, and the
// redelivery is acknowledged without queueing the event a second time.
func TestACrashAfterTheWriteBeforeTheAckIsAbsorbedByTheRedelivery(t *testing.T) {
	dir := t.TempDir()
	fd := openFault(t, dir)
	acked := 0
	fd.arm(0, 1)
	if err := newQueue(fd).Accept("d1", hook, func() { acked++ }); !errors.Is(err, errCrash) || acked != 0 {
		t.Fatalf("err=%v acked=%d", err, acked)
	}
	q := newQueue(openStore(t, dir)) // restart
	if ids := pendingIDs(t, q); len(ids) != 1 {
		t.Fatalf("the durable write was lost: %v", ids)
	}
	if err := q.Accept("d1", hook, func() { acked++ }); err != nil || acked != 1 {
		t.Fatalf("redelivery: err=%v acked=%d", err, acked)
	}
	if ids := pendingIDs(t, q); len(ids) != 1 {
		t.Errorf("the redelivery duplicated the event: %v", ids)
	}
}

func TestARedeliveryAfterTheEventWasConsumedIsNotQueuedAgain(t *testing.T) {
	q := newQueue(&memDocs{})
	if err := q.Enqueue("d1", hook); err != nil {
		t.Fatal(err)
	}
	if err := q.Commit([]string{"d1"}); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue("d1", hook); err != nil {
		t.Fatal(err)
	}
	if ids := pendingIDs(t, q); len(ids) != 0 {
		t.Errorf("a consumed event came back: %v", ids)
	}
}

// The memory of consumed ids is bounded by configuration: the oldest is forgotten first.
func TestTheConsumedIdMemoryIsBounded(t *testing.T) {
	q := newQueue(&memDocs{})
	q.SeenLimit = 2
	for _, id := range []string{"a", "b", "c"} {
		_ = q.Enqueue(id, hook)
		_ = q.Commit([]string{id})
	}
	_ = q.Enqueue("c", hook) // still remembered
	_ = q.Enqueue("a", hook) // forgotten
	if ids := pendingIDs(t, q); len(ids) != 1 || ids[0] != "a" {
		t.Errorf("pending = %v, want only the forgotten id a", ids)
	}
}

// A full queue refuses and does not acknowledge; nothing already queued is dropped to make room.
func TestAFullQueueRefusesWithoutDroppingOrAcknowledging(t *testing.T) {
	q := newQueue(&memDocs{})
	q.MaxDocBytes = 400
	q.SeenLimit, q.MaxIDBytes = 1, 16
	var queued []string
	acked := 0
	var full error
	for i := range 40 {
		id := fmt.Sprintf("delivery-%02d", i)
		if err := q.Accept(id, forge.Change{Key: fmt.Sprint(i), Kind: forge.ChangeAdded}, func() { acked++ }); err != nil {
			full = err
			break
		}
		queued = append(queued, id)
	}
	if !errors.Is(full, ErrQueueFull) {
		t.Fatalf("err = %v, want ErrQueueFull", full)
	}
	if acked != len(queued) {
		t.Errorf("acknowledged %d, queued %d: the refused event was acknowledged", acked, len(queued))
	}
	got := pendingIDs(t, q)
	if strings.Join(got, ",") != strings.Join(queued, ",") {
		t.Errorf("queued events changed: %v, want %v", got, queued)
	}
}

func TestAnEventWithoutADeliveryIdIsRefused(t *testing.T) {
	q := newQueue(&memDocs{})
	acked := 0
	if err := q.Accept("", hook, func() { acked++ }); !errors.Is(err, ErrNoDeliveryID) || acked != 0 {
		t.Errorf("err=%v acked=%d", err, acked)
	}
}

func TestAnUnreadableQueueIsAnErrorNotAnEmptyQueue(t *testing.T) {
	m := &memDocs{}
	q := newQueue(m)
	if _, err := m.PutDoc(CollectorKind, q.name(), 0, []byte(`"garbage"`)); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Pending(); err == nil {
		t.Error("an unreadable queue read as empty")
	}
	if err := q.Enqueue("d1", hook); err == nil {
		t.Error("an event was queued over an unreadable queue document")
	}
}

func TestConcurrentDeliveriesAreAllQueued(t *testing.T) {
	st := openStore(t, t.TempDir())
	const n = 6
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range n {
		wg.Go(func() {
			errs[i] = newQueue(st).Enqueue(fmt.Sprintf("d%d", i), forge.Change{Key: fmt.Sprint(i), Kind: forge.ChangeAdded})
		})
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("delivery %d: %v", i, err)
		}
	}
	if ids := pendingIDs(t, newQueue(st)); len(ids) != n {
		t.Errorf("queued %d of %d concurrent deliveries: %v", len(ids), n, ids)
	}
}
