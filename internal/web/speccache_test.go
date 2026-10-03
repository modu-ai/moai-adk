package web

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Card t1460 — the SPEC rows + audit findings are computed once and reused until
// .moai/specs changes. Every request used to re-run spec.ListDocs + spec.Audit
// over the whole catalog, which made /, /kanban and /specs take seconds.

// countingLoader returns a loader that counts its calls and yields one row whose
// title is the call number, so a test can tell a cached result from a fresh one.
func countingLoader(calls *atomic.Int32) specLoader {
	return func(root string) ([]SpecRowVM, map[string][]FindingVM, error) {
		n := calls.Add(1)
		return []SpecRowVM{{ID: "SPEC-X-001", Title: itoa(int(n))}}, map[string][]FindingVM{}, nil
	}
}

// (a) Two consecutive loads run the underlying scan once.
func TestSpecCache_ConsecutiveLoadsScanOnce(t *testing.T) {
	var calls atomic.Int32
	c := newSpecCache(countingLoader(&calls), time.Hour)

	if _, _, err := c.get("/root"); err != nil {
		t.Fatalf("first get: %v", err)
	}
	rows, _, err := c.get("/root")
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("scan ran %d times for two loads, want 1", got)
	}
	if rows[0].Title != "1" {
		t.Fatalf("second load title = %q, want the cached %q", rows[0].Title, "1")
	}
}

// A max-age fallback re-scans even when no invalidation arrived, so a missed fs
// event cannot serve stale data forever.
func TestSpecCache_MaxAgeFallbackRescans(t *testing.T) {
	var calls atomic.Int32
	c := newSpecCache(countingLoader(&calls), time.Minute)
	now := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }

	_, _, _ = c.get("/root")
	now = now.Add(59 * time.Second)
	_, _, _ = c.get("/root")
	if got := calls.Load(); got != 1 {
		t.Fatalf("scan ran %d times inside max age, want 1", got)
	}
	now = now.Add(2 * time.Second)
	_, _, _ = c.get("/root")
	if got := calls.Load(); got != 2 {
		t.Fatalf("scan ran %d times after max age, want 2", got)
	}
}

// The max age counts from when a scan finished, not when it started: on a loaded
// machine a scan takes tens of seconds, and anchoring at the start let that
// scan consume its own age so the very next request rescanned (measured during
// card t1460: a 22.5 s cold scan, then a 19.7 s rescan 32 s later).
func TestSpecCache_SlowScanDoesNotConsumeItsOwnAge(t *testing.T) {
	var calls atomic.Int32
	now := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	c := newSpecCache(func(root string) ([]SpecRowVM, map[string][]FindingVM, error) {
		calls.Add(1)
		now = now.Add(50 * time.Second) // the scan itself takes 50 s
		return []SpecRowVM{{ID: "SPEC-X-001"}}, nil, nil
	}, time.Minute)
	c.now = func() time.Time { return now }

	_, _, _ = c.get("/root")
	now = now.Add(10 * time.Second)
	_, _, _ = c.get("/root")
	if got := calls.Load(); got != 1 {
		t.Fatalf("scan ran %d times 10 s after a 50 s scan with a 60 s max age, want 1", got)
	}
}

// (b) After a SPEC file changes and the hub's "spec" signal fires, the next load
// reflects the change. Exercised through the app wiring with the real loader.
func TestSpecCache_SpecEventInvalidatesApp(t *testing.T) {
	root := t.TempDir()
	writeBoardSpec(t, root, "SPEC-CACHE-001",
		"id: SPEC-CACHE-001\ntitle: \"Cache one\"\nstatus: draft\nupdated: 2026-10-01\n", "")
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})

	vm, err := a.buildSpecList("", "", "")
	if err != nil {
		t.Fatalf("buildSpecList: %v", err)
	}
	if len(vm.Rows) != 1 || vm.Rows[0].Status != "draft" {
		t.Fatalf("initial rows = %+v, want one draft row", vm.Rows)
	}

	writeBoardSpec(t, root, "SPEC-CACHE-001",
		"id: SPEC-CACHE-001\ntitle: \"Cache one\"\nstatus: completed\nupdated: 2026-10-02\n", "")

	// An unrelated signal leaves the cache in place.
	a.hub.Publish("verify")
	vm, _ = a.buildSpecList("", "", "")
	if vm.Rows[0].Status != "draft" {
		t.Fatalf("status after unrelated event = %q, want cached draft", vm.Rows[0].Status)
	}

	a.hub.Publish("spec")
	vm, err = a.buildSpecList("", "", "")
	if err != nil {
		t.Fatalf("buildSpecList after change: %v", err)
	}
	if vm.Rows[0].Status != "completed" {
		t.Fatalf("status after spec event = %q, want completed (stale data served)", vm.Rows[0].Status)
	}
}

// An invalidation that lands while a scan is in flight must not let that scan's
// (possibly pre-change) result be stored as fresh.
func TestSpecCache_InvalidateDuringScanIsNotLost(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	entered := make(chan struct{}, 1)
	first := true
	c := newSpecCache(func(root string) ([]SpecRowVM, map[string][]FindingVM, error) {
		n := calls.Add(1)
		if first {
			first = false
			entered <- struct{}{}
			<-release
		}
		return []SpecRowVM{{ID: "SPEC-X-001", Title: itoa(int(n))}}, nil, nil
	}, time.Hour)

	done := make(chan struct{})
	go func() { _, _, _ = c.get("/root"); close(done) }()
	<-entered
	c.invalidate()
	close(release)
	<-done

	rows, _, _ := c.get("/root")
	if got := calls.Load(); got != 2 {
		t.Fatalf("scan ran %d times, want 2 (invalidation during scan must force a rescan)", got)
	}
	if rows[0].Title != "2" {
		t.Fatalf("title = %q, want the rescanned %q", rows[0].Title, "2")
	}
}

// (c) N concurrent loads trigger one scan.
func TestSpecCache_ConcurrentLoadsScanOnce(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	c := newSpecCache(func(root string) ([]SpecRowVM, map[string][]FindingVM, error) {
		calls.Add(1)
		<-release
		return []SpecRowVM{{ID: "SPEC-X-001"}}, nil, nil
	}, time.Hour)

	const n = 16
	var started, finished sync.WaitGroup
	started.Add(n)
	finished.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer finished.Done()
			started.Done()
			if rows, _, err := c.get("/root"); err != nil || len(rows) != 1 {
				t.Errorf("get = %v, %v", rows, err)
			}
		}()
	}
	started.Wait()
	time.Sleep(20 * time.Millisecond) // let the callers queue behind the first scan
	close(release)
	finished.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("%d concurrent loads ran the scan %d times, want 1", n, got)
	}
}

// Callers get their own rows slice: an in-place sort by one request must not
// reorder the cached catalog another request reads.
func TestSpecCache_RowsAreCopied(t *testing.T) {
	var calls atomic.Int32
	c := newSpecCache(countingLoader(&calls), time.Hour)
	rows, _, _ := c.get(filepath.FromSlash("/root"))
	rows[0].Title = "mutated"
	again, _, _ := c.get("/root")
	if again[0].Title != "1" {
		t.Fatalf("cached title = %q, a caller's mutation leaked into the cache", again[0].Title)
	}
}
