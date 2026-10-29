package docstore

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

func small(d Docs) Cursors {
	return Cursors{Docs: d, Kind: CollectorKind, MaxDocBytes: 8 << 10, MaxChunks: 16}
}

// referenced lists every chunk document the manifest of key names, for both generations.
func referenced(t *testing.T, st *claim.Store, key string) []string {
	t.Helper()
	set := manifestOf(t, st, key)
	var names []string
	for ; set != nil; set = set.Prev {
		for _, r := range set.Chunks {
			names = append(names, r.Name)
		}
	}
	return names
}

// The cleanup step failing must not fail the save or damage the cursor; it only leaves unreferenced
// chunks behind (garbage no manifest names, never a wrong cursor).
func TestAFailingCleanupLeavesGarbageButNeverABadCursor(t *testing.T) {
	dir := t.TempDir()
	fd := openFault(t, dir)
	c := small(fd)
	bodies := [][]byte{bigBody(1, 24<<10), bigBody(2, 24<<10), bigBody(3, 24<<10), bigBody(4, 24<<10)}
	for i, b := range bodies {
		if i == 2 {
			fd.delErr = errors.New("cleanup refused") // the generation before the previous one cannot be removed
		}
		if err := c.SaveCursor("k", b); err != nil {
			t.Fatalf("save %d failed because its cleanup failed: %v", i, err)
		}
		if got, found, err := small(openStore(t, dir)).LoadCursor("k"); err != nil || !found || !bytes.Equal(got, b) {
			t.Fatalf("save %d: found=%v err=%v isNew=%v", i, found, err, bytes.Equal(got, b))
		}
	}
	st := openStore(t, dir)
	for _, name := range referenced(t, st, "k") {
		if _, err := st.GetDoc(CollectorKind, name); err != nil {
			t.Errorf("a chunk the manifest names is missing: %s (%v)", name, err)
		}
	}
	if docs, _ := st.ListDocs(CollectorKind); len(docs) <= 1+len(referenced(t, st, "k")) {
		t.Errorf("expected leftover garbage from the refused cleanup, got %d documents", len(docs))
	}
}

// saveRetryingBusy is a caller that honours the claim store's contract: a held document lock is Busy, which
// means "retry", and nothing else is retried. It returns how many times it was told Busy.
func saveRetryingBusy(c Cursors, key string, body []byte) (busy int, err error) {
	for deadline := time.Now().Add(30 * time.Second); ; {
		err = c.SaveCursor(key, body)
		if !errors.Is(err, core.ErrBusy) {
			return busy, err
		}
		busy++
		if time.Now().After(deadline) {
			return busy, err
		}
	}
}

// Two writers saving the same cursor concurrently: every save succeeds once Busy is retried as the claim
// store's contract says (a held lock is Busy, never a failure of the write), no other error appears, and
// what is left is one whole cursor that one of them wrote. Under load the global document lock can run its
// whole wait budget out for a writer (measured: 2 of 200 back-to-back puts at load ~60), so Busy here is
// expected and the test must not read it as a failure.
func TestConcurrentSavesOfOneCursorLeaveOneWholeCursor(t *testing.T) {
	dir := t.TempDir()
	const writers, rounds = 2, 5
	written := map[string]bool{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make(chan error, writers*rounds)
	var busyTotal int
	for w := range writers {
		wg.Go(func() {
			c := small(openStore(t, dir)) // each writer is its own process over the same documents
			for r := range rounds {
				b := bigBody(int64(100*w+r), 24<<10)
				mu.Lock()
				written[string(b)] = true
				mu.Unlock()
				busy, err := saveRetryingBusy(c, "k", b)
				mu.Lock()
				busyTotal += busy
				mu.Unlock()
				if err != nil {
					errs <- fmt.Errorf("writer %d round %d: %w", w, r, err)
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	t.Logf("saves told Busy and retried: %d", busyTotal)
	got, found, err := small(openStore(t, dir)).LoadCursor("k")
	if err != nil || !found || !written[string(got)] {
		t.Errorf("found=%v err=%v writtenByOne=%v: the result is not one whole cursor a writer saved", found, err, written[string(got)])
	}
}

// The Busy signal itself, without any load: while another holder keeps the documents lock past the wait
// budget, a save is told Busy (not "failed", not "unavailable"), the stored cursor is untouched, and the
// same save succeeds once the lock is free.
func TestASaveIsToldBusyWhileTheDocumentLockIsHeldAndSucceedsWhenItIsFree(t *testing.T) {
	dir := t.TempDir()
	c := small(openStore(t, dir))
	v1, v2 := bigBody(1, 24<<10), bigBody(2, 24<<10)
	if err := c.SaveCursor("k", v1); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, "docs.lock")
	f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	released := false
	t.Cleanup(func() {
		if !released {
			_ = os.Remove(lock)
		}
	})
	err = c.SaveCursor("k", v2)
	if !errors.Is(err, core.ErrBusy) || errors.Is(err, core.ErrStoreUnavailable) {
		t.Fatalf("err = %v, want ErrBusy while the lock is held", err)
	}
	if got, found, lerr := small(openStore(t, dir)).LoadCursor("k"); lerr != nil || !found || !bytes.Equal(got, v1) {
		t.Errorf("a Busy save changed the cursor: found=%v err=%v isV1=%v", found, lerr, bytes.Equal(got, v1))
	}
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	released = true
	if err := c.SaveCursor("k", v2); err != nil {
		t.Fatalf("the same save after the lock was freed: %v", err)
	}
	if got, _, _ := small(openStore(t, dir)).LoadCursor("k"); !bytes.Equal(got, v2) {
		t.Error("the retried save did not land whole")
	}
}

// Generations that share chunks (a body returning to earlier content, or changing only its tail) must
// keep the shared chunks: cleanup removes only what neither retained generation uses.
func TestChunksSharedBetweenGenerationsSurviveCleanup(t *testing.T) {
	dir := t.TempDir()
	st := openStore(t, dir)
	c := small(st)
	a := bigBody(1, 24<<10)
	b := append(append([]byte(nil), a[:len(a)-200]...), bytes.Repeat([]byte("z"), 200)...) // same head, new tail
	for i, body := range [][]byte{a, b, a, b, a, a} {
		if err := c.SaveCursor("k", body); err != nil {
			t.Fatal(err)
		}
		if got, found, err := c.LoadCursor("k"); err != nil || !found || !bytes.Equal(got, body) {
			t.Fatalf("save %d: found=%v err=%v", i, found, err)
		}
		for _, name := range referenced(t, st, "k") {
			if _, err := st.GetDoc(CollectorKind, name); err != nil {
				t.Fatalf("save %d: a chunk a retained generation names is gone: %s (%v)", i, name, err)
			}
		}
	}
}

// A remembered-id list that could take the document is refused when the queue is configured, instead of
// turning every later delivery into ErrQueueFull.
func TestASeenMemoryThatCouldFillTheQueueIsRefused(t *testing.T) {
	q := newQueue(&memDocs{})
	q.MaxDocBytes, q.SeenLimit, q.MaxIDBytes = 400, 10, 36 // 10 ids of up to 36 bytes need 390 of 400
	if err := q.Validate(); !errors.Is(err, ErrQueueConfig) {
		t.Fatalf("Validate = %v, want ErrQueueConfig", err)
	}
	acked := 0
	if err := q.Accept("d1", hook, func() { acked++ }); !errors.Is(err, ErrQueueConfig) || acked != 0 {
		t.Errorf("a misconfigured queue accepted a delivery: err=%v acked=%d", err, acked)
	}
	q.SeenLimit, q.MaxIDBytes = 3, 0
	if err := q.Validate(); !errors.Is(err, ErrQueueConfig) {
		t.Errorf("SeenLimit without MaxIDBytes = %v, want ErrQueueConfig", err)
	}
}

// Within the validated bound a full memory of consumed ids still leaves room for pending events.
func TestAFullSeenMemoryStillLeavesRoomForPendingEvents(t *testing.T) {
	q := newQueue(&memDocs{})
	q.MaxDocBytes, q.SeenLimit, q.MaxIDBytes = 1000, 8, 16
	if err := q.Validate(); err != nil {
		t.Fatal(err)
	}
	for i := range 20 { // churn well past the memory size
		id := fmt.Sprintf("delivery-%02d", i)
		if err := q.Enqueue(id, hook); err != nil {
			t.Fatalf("delivery %d: %v", i, err)
		}
		if err := q.Commit([]string{id}); err != nil {
			t.Fatal(err)
		}
	}
	for i := range 3 {
		if err := q.Enqueue(fmt.Sprintf("pending-%d", i), hook); err != nil {
			t.Fatalf("pending %d with a full memory: %v", i, err)
		}
	}
}

func TestAnOverlongDeliveryIdIsRefusedAndNotAcknowledged(t *testing.T) {
	q := newQueue(&memDocs{})
	acked := 0
	long := string(bytes.Repeat([]byte("x"), q.MaxIDBytes+1))
	if err := q.Accept(long, hook, func() { acked++ }); !errors.Is(err, ErrBadDeliveryID) || acked != 0 {
		t.Errorf("err=%v acked=%d", err, acked)
	}
	if ids := pendingIDs(t, q); len(ids) != 0 {
		t.Errorf("queued %v", ids)
	}
}

// Resuming from the previous generation is allowed but not silent: the report says so until an
// acknowledgement is saved, and the changes since that older cursor come back (never a baseline).
func TestFallingBackToThePreviousGenerationIsSaidInTheReport(t *testing.T) {
	dir := t.TempDir()
	st := openStore(t, dir)
	c := small(st)
	v1, v2 := bigBody(1, 24<<10), bigBody(2, 24<<10)
	_ = c.SaveCursor("o/r/big", v1)
	_ = c.SaveCursor("o/r/big", v2)
	removeDoc(t, st, manifestOf(t, st, "o/r/big").Chunks[0].Name)
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	col := &forge.Collector{
		Client: &forge.Client{T: &oneShot{body: string(v2)}, B: &forge.Budget{}, Now: func() time.Time { return now },
			Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
		Path: "/p", Interval: forge.Adaptive{Min: time.Second, Max: time.Minute, Safety: time.Hour},
		Differ: forge.DiffByKey("number"), Cursors: small(openStore(t, dir)), CursorKey: "o/r/big",
	}
	rep := col.Tick(t.Context(), now)
	if rep.Notice != FallbackNote || rep.Baseline || keys(rep) != "1:modified" {
		t.Fatalf("report = notice %q baseline=%v changes=%q", rep.Notice, rep.Baseline, keys(rep))
	}
	if err := col.Ack(rep); err != nil {
		t.Fatal(err)
	}
	if again := col.Tick(t.Context(), rep.Next); again.Notice != "" {
		t.Errorf("the notice outlived the acknowledgement: %q", again.Notice)
	}
}
