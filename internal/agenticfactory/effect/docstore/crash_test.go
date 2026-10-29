package docstore

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// rig boots a collector over a directory the way a restart does: new store, new queue, new collector, no
// memory of the previous process. Everything lives in bare temp directories; nothing is sent anywhere.
type rig struct {
	t    *testing.T
	dir  string
	now  time.Time
	body string // what the (fake) host answers when the collector polls
	chk  int    // MaxChunks of the cursor store
}

func newRig(t *testing.T) *rig {
	return &rig{t: t, dir: t.TempDir(), now: time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC), body: `[]`, chk: 8}
}

func (r *rig) boot(pushHealthy bool) (*forge.Collector, *Queue, *faultDocs) {
	fd := openFault(r.t, r.dir)
	q := newQueue(fd)
	q.Alive = func() bool { return pushHealthy }
	col := &forge.Collector{
		Client: &forge.Client{T: &oneShot{body: r.body}, B: &forge.Budget{}, Now: func() time.Time { return r.now },
			Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
		Path: "/p", Interval: forge.Adaptive{Min: time.Second, Max: time.Minute, Safety: time.Hour},
		Differ: forge.DiffByKey("number"), Durable: q,
		Cursors: Cursors{Docs: fd, Kind: CollectorKind, MaxChunks: r.chk}, CursorKey: "o/r/pulls",
	}
	return col, q, fd
}

func keys(rep forge.Report) string {
	var k []string
	for _, c := range rep.Changes {
		k = append(k, c.Key+":"+string(c.Kind))
	}
	slices.Sort(k)
	return strings.Join(k, ",")
}

// Crash 1 — right after the enqueue, before the acknowledgement reached the host. The host redelivers.
// Result: the event is reported once, then never again.
func TestCrashAfterEnqueueBeforeTheHostAckReportsTheEventExactlyOnce(t *testing.T) {
	r := newRig(t)
	_, q, fd := r.boot(true)
	fd.arm(0, 1)
	acked := 0
	if err := q.Accept("d1", hook, func() { acked++ }); !errors.Is(err, errCrash) || acked != 0 {
		t.Fatalf("err=%v acked=%d", err, acked)
	}
	colB, qB, _ := r.boot(true) // restart
	if err := qB.Accept("d1", hook, func() { acked++ }); err != nil || acked != 1 {
		t.Fatalf("redelivery: err=%v acked=%d", err, acked)
	}
	rep := colB.Tick(t.Context(), r.now)
	if got := keys(rep); got != "7:modified" {
		t.Fatalf("after the restart: %q, want the event once", got)
	}
	if err := colB.Ack(rep); err != nil {
		t.Fatal(err)
	}
	colC, _, _ := r.boot(true)
	if got := keys(colC.Tick(t.Context(), r.now)); got != "" {
		t.Errorf("an acknowledged event came back: %q", got)
	}
}

// Crash 2 — the event is queued and acknowledged to the host, the process dies before the collector looks.
func TestCrashAfterTheHostAckBeforeTheCollectorTickLosesNothing(t *testing.T) {
	r := newRig(t)
	_, q, _ := r.boot(true)
	acked := 0
	if err := q.Accept("d1", hook, func() { acked++ }); err != nil || acked != 1 {
		t.Fatal(err)
	}
	colB, _, _ := r.boot(true)
	rep := colB.Tick(t.Context(), r.now)
	if got := keys(rep); got != "7:modified" {
		t.Fatalf("the event was lost: %q", got)
	}
	_ = colB.Ack(rep)
	colC, _, _ := r.boot(true)
	if got := keys(colC.Tick(t.Context(), r.now)); got != "" {
		t.Errorf("duplicate after the acknowledgement: %q", got)
	}
}

// Crash 3 — the report was handed to the consumer, the process dies before Ack. At-least-once: the same
// change is reported again (a duplicate the consumer's idempotency absorbs), and never after the Ack.
func TestCrashBetweenTickAndAckRedeliversOnceMoreAndNeverLoses(t *testing.T) {
	r := newRig(t)
	_, q, _ := r.boot(true)
	_ = q.Enqueue("d1", hook)
	colB, _, _ := r.boot(true)
	if got := keys(colB.Tick(t.Context(), r.now)); got != "7:modified" {
		t.Fatalf("first delivery: %q", got)
	}
	colC, _, _ := r.boot(true) // crashed before Ack
	rep := colC.Tick(t.Context(), r.now)
	if got := keys(rep); got != "7:modified" {
		t.Fatalf("the unacknowledged event must come back: %q", got)
	}
	_ = colC.Ack(rep)
	colD, _, _ := r.boot(true)
	if got := keys(colD.Tick(t.Context(), r.now)); got != "" {
		t.Errorf("duplicate after the acknowledgement: %q", got)
	}
}

// Crash 4 — the cursor was saved, the process dies before the queue commit. The cursor stays advanced (no
// baseline, no repeat of polled changes) and the pushed event is delivered again until committed.
func TestCrashBetweenTheCursorSaveAndTheQueueCommitLosesNothing(t *testing.T) {
	r := newRig(t)
	r.body = `[{"number":1,"s":"open"}]`
	colA, q, fd := r.boot(false) // push not healthy: the collector polls
	_ = q.Enqueue("d1", hook)
	rep := colA.Tick(t.Context(), r.now)
	if !rep.Baseline || keys(rep) != "1:added,7:modified" {
		t.Fatalf("first report = %q baseline=%v", keys(rep), rep.Baseline)
	}
	fd.arm(2, 0) // put 1 = the cursor, put 2 = the queue commit
	if err := colA.Ack(rep); !errors.Is(err, errCrash) {
		t.Fatalf("ack err = %v, want the crash at the commit", err)
	}
	colB, _, _ := r.boot(false)
	repB := colB.Tick(t.Context(), r.now)
	if repB.Baseline || keys(repB) != "7:modified" {
		t.Fatalf("after the restart: %q baseline=%v, want only the uncommitted push event", keys(repB), repB.Baseline)
	}
	if err := colB.Ack(repB); err != nil {
		t.Fatal(err)
	}
	colC, _, _ := r.boot(false)
	if got := keys(colC.Tick(t.Context(), r.now)); got != "" {
		t.Errorf("after the commit: %q", got)
	}
}

// Crash 5 — the process dies with only some chunks of a big cursor written. The previous cursor stays; the
// changes come back; nothing is invented; the redo then completes.
func TestCrashWithSomeChunksWrittenKeepsThePreviousCursorAndLosesNoChange(t *testing.T) {
	r := newRig(t)
	small := `[{"number":1,"pad":"a"}]`
	r.body = small
	colA, _, _ := r.boot(false)
	repA := colA.Tick(t.Context(), r.now)
	if err := colA.Ack(repA); err != nil {
		t.Fatal(err)
	}
	big := string(bigBody(9, 600<<10))
	r.body = big
	r.now = r.now.Add(time.Hour)
	colB, _, fdB := r.boot(false)
	repB := colB.Tick(t.Context(), r.now)
	if repB.Baseline || keys(repB) != "1:modified" {
		t.Fatalf("big poll: %q baseline=%v", keys(repB), repB.Baseline)
	}
	fdB.arm(0, 2) // two chunks are written, then the crash
	if err := colB.Ack(repB); !errors.Is(err, errCrash) {
		t.Fatalf("ack err = %v", err)
	}
	colC, _, _ := r.boot(false) // restart: the cursor is still the small one
	repC := colC.Tick(t.Context(), r.now)
	if repC.Baseline || keys(repC) != "1:modified" {
		t.Fatalf("after the crash: %q baseline=%v, want the change again from the previous cursor", keys(repC), repC.Baseline)
	}
	if err := colC.Ack(repC); err != nil {
		t.Fatalf("the redo failed: %v", err)
	}
	colD, _, _ := r.boot(false)
	repD := colD.Tick(t.Context(), r.now.Add(time.Hour))
	if repD.Baseline || keys(repD) != "" {
		t.Errorf("after the redo: %q baseline=%v, want nothing", keys(repD), repD.Baseline)
	}
}

// A cursor too large for the configured bound is refused at the Ack; the collector keeps reporting the
// change and never claims the cursor advanced.
func TestACursorBeyondTheChunkBoundIsRefusedAtTheAckAndNothingAdvances(t *testing.T) {
	r := newRig(t)
	r.chk = 2
	r.body = string(bigBody(4, 600<<10))
	col, _, fd := r.boot(false)
	rep := col.Tick(t.Context(), r.now)
	fd.arm(0, 0)
	if err := col.Ack(rep); !errors.Is(err, ErrCursorTooLarge) {
		t.Fatalf("ack err = %v, want ErrCursorTooLarge", err)
	}
	if fd.puts != 0 {
		t.Errorf("%d writes for a cursor that cannot be stored", fd.puts)
	}
	again := col.Tick(t.Context(), rep.Next)
	if !again.Baseline || len(again.Changes) != 1 {
		t.Errorf("the refused ack must not advance: %+v", again)
	}
}
