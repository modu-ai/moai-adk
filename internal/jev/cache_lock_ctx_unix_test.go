//go:build unix

package jev

// cache_lock_ctx_unix_test.go — iter16 F1 (card t1968, operator-approved
// bounded-retry repair): the alias-map flock was a BLOCKING LOCK_EX reached
// from Ask only AFTER the network answer was final, so a contending holder
// stalled an already-answered call unboundedly — a 30ms-deadline Ask with
// the answer in hand blocked past 250ms under a held lock. The lock acquire
// must honor ctx and a max wait: at either bound the alias-map update is
// SKIPPED and the call returns (never worse than the pre-lock code, which
// lost this append silently in the same contended case). Driven through the
// real Ask path — lookup miss -> network answer -> cacheStore -> the lock
// wait — with the contending lock held from the test's own fd, the same
// cross-process shape production contends on.

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// aliasLockHoldWindow is how long the test's own fd holds the alias-map
// lock: long enough that an unbounded wait cannot mistake it for a release
// (the expired-deadline fixture FAILS at ~this duration on the current
// blocking code), short enough that a RED run still terminates.
const aliasLockHoldWindow = time.Second

// holdAliasLock opens models.lock in dir and holds an exclusive flock on it
// from the test's own fd, releasing (best-effort) after the hold window so
// a RED run's blocked Ask still wakes and the failing measurement lands
// rather than a watchdog timeout.
func holdAliasLock(t *testing.T, dir string) *os.File {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, "models.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open lock file: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		t.Fatalf("hold alias-map lock: %v", err)
	}
	release := time.AfterFunc(aliasLockHoldWindow, func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) })
	t.Cleanup(func() {
		release.Stop()
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	})
	return f
}

// arrivedDoer answers immediately with a 200 naming its model and signals
// arrival, so the test knows the network answer is FINAL and the call has
// moved on to the post-send cache writes — the point where the lock wait
// begins.
type arrivedDoer struct {
	model   string
	arrived chan struct{}
	once    sync.Once
}

func (d *arrivedDoer) Do(r *http.Request) (*http.Response, error) {
	d.once.Do(func() { close(d.arrived) })
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(cacheOKBody(d.model))),
		Header:     http.Header{},
		Request:    r,
	}, nil
}

// An Ask whose ctx deadline is ALREADY expired — the answer already in
// hand, the lock held by another fd — must return the answer promptly
// instead of blocking on the contended alias-map lock. This is the iter16
// F1 fixture: at the current blocking LOCK_EX the call sits in the flock
// until the holder releases, so the ~200ms promptness bound fails at the
// full hold-window duration.
func TestAsk_ExpiredCtxDeadlineReturnsDespiteHeldAliasLock(t *testing.T) {
	dir := t.TempDir()
	holdAliasLock(t, dir)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	arrived := make(chan struct{})
	c := &Client{
		Enabled:        true,
		LoadCredential: func() string { return "test-key" },
		HTTP:           &arrivedDoer{model: "jev-1.20.0", arrived: arrived},
		CacheDir:       dir,
		MaxRetries:     1,
	}

	start := time.Now()
	res := c.Ask(ctx, cacheQuestion())
	elapsed := time.Since(start)
	if !res.OK() {
		t.Fatalf("an answered call must return OK; got %q (%s)", res.Availability, res.Condition)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Ask with an already-expired deadline and the answer in hand blocked %v on the held alias-map lock; want ≤ 200ms", elapsed)
	}
}

// An Ask whose ctx cancels WHILE WAITING on a held alias-map lock must
// return promptly with the answer (cancellation honored, update skipped):
// the lock wait selects on ctx.Done(). At the current blocking LOCK_EX the
// cancellation is ignored and the call sits until the holder releases.
func TestAsk_CtxCancelDuringAliasLockWaitSkipsUpdate(t *testing.T) {
	dir := t.TempDir()
	holdAliasLock(t, dir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	arrived := make(chan struct{})
	c := &Client{
		Enabled:        true,
		LoadCredential: func() string { return "test-key" },
		HTTP:           &arrivedDoer{model: "jev-1.20.0", arrived: arrived},
		CacheDir:       dir,
		MaxRetries:     1,
	}
	type outcome struct {
		res     Result
		sinceCx time.Duration
	}
	done := make(chan outcome, 1)
	go func() {
		res := c.Ask(ctx, cacheQuestion())
		done <- outcome{res: res}
	}()

	<-arrived // the answer is final; the call is at (or just before) the lock wait
	time.Sleep(20 * time.Millisecond)
	cxStart := time.Now()
	cancel()
	select {
	case o := <-done:
		if !o.res.OK() {
			t.Fatalf("an answered call must return OK; got %q (%s)", o.res.Availability, o.res.Condition)
		}
		if d := time.Since(cxStart); d > 500*time.Millisecond {
			t.Fatalf("Ask kept waiting %v after ctx cancellation on a held alias-map lock; want prompt return", d)
		}
		// The skipped update must not write the alias map: no models.json
		// in the dir (the cache ENTRY file is separate and may exist).
		if _, err := os.Stat(filepath.Join(dir, modelsMapName)); !os.IsNotExist(err) {
			t.Errorf("alias map was persisted despite the skipped lock wait: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Ask did not return after ctx cancellation — the alias-map lock wait ignores ctx")
	}
}

// updateAliasMap on a CacheDir that does not exist yet (a first-ever call
// on a fresh machine) must create the dir and record the model — the
// restart-offline hit depends on the map persisting. At the current tree
// the lock's OpenFile fails ENOENT before writeModelsMap's own MkdirAll
// precondition (inside atomicWrite) ever runs, so the update errors and
// the map is never written (iter16 F2).
func TestUpdateAliasMap_MissingCacheDirCreatedAndRecorded(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "jev-cache") // does not exist
	c := &Client{CacheDir: dir}
	if err := c.updateAliasMap(context.Background(), "jev-1.20.0"); err != nil {
		t.Fatalf("updateAliasMap on a missing cache dir errored: %v", err)
	}
	m, err := readModelsMap(dir)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got := m.Models[ModelID]; got != "jev-1.20.0" {
		t.Fatalf("alias map did not record the model: got %q, want jev-1.20.0", got)
	}
}
