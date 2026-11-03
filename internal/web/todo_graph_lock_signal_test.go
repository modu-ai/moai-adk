package web

// todo_graph_lock_signal_test.go — tells "waits on the queue lock" apart from
// "answers late" (Issue #1948, B9). TestTodoGraphViewDoesNotWaitOnQueueLock keeps
// its two-second condition untouched; under -race and a starved CPU it has
// failed once (load 98→33, 468.4s) without saying which of the two happened.
// This test asserts the contract by signal and order instead of by clock:
//
//   - the lock is held by a holder that has not been released;
//   - the page request is answered while that holder still holds the lock
//     (an order, not a duration) — however late;
//   - at no point while it is pending is any goroutine inside the queue's lock
//     acquisition (a signal read from the goroutine stacks), so a handler that
//     waits on the lock fails at once with that cause, and one that is merely
//     slow just keeps being waited for.
//
// A control contender proves the detector can see a real lock wait; without it a
// blind detector would make the contract assertion vacuous.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// lockWaiters counts goroutines currently inside the queue's lock acquisition.
func lockWaiters() int {
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	count := 0
	for _, g := range strings.Split(string(buf[:n]), "\n\n") {
		if strings.Contains(g, "factory.(*BacklogStore).acquireLock") {
			count++
		}
	}
	return count
}

func TestTodoGraphViewAnswersWhileLockHeldWithoutWaitingOnIt(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	path := writeBacklog(t, root, graphQueue)
	store := factory.NewBacklogStore(path)

	release := make(chan struct{})
	acquired := make(chan struct{})
	holderDone := make(chan error, 1)
	var rest sync.WaitGroup
	t.Cleanup(func() {
		close(release)
		if err := <-holderDone; err != nil {
			t.Errorf("queue lock holder: %v", err)
		}
		rest.Wait()
	})
	go func() {
		holderDone <- store.Mutate(func(*factory.BacklogRecord) error {
			close(acquired)
			<-release
			return nil
		})
	}()
	<-acquired // the holder is inside the lock; no clock decides this

	if got := lockWaiters(); got != 0 {
		t.Fatalf("before any request %d goroutines already wait on the queue lock", got)
	}

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	type result struct {
		code int
		body string
	}
	answered := make(chan result, 1)
	rest.Add(1)
	go func() {
		defer rest.Done()
		rec := serveGet(t, a.routes(), "/todo?view=graph")
		answered <- result{rec.Code, rec.Body.String()}
	}()

	// Wait for the answer. The only clock here is a poll interval for reading the
	// signal; nothing bounds how late the answer may come.
	for {
		select {
		case r := <-answered:
			select {
			case <-release:
				t.Fatal("the holder was released before the page answered: the order was not observed")
			default:
			}
			if r.code != http.StatusOK || !strings.Contains(r.body, "t9") {
				t.Fatalf("answer under a held lock: status %d, archived node present: %v", r.code, strings.Contains(r.body, "t9"))
			}
			goto control
		case <-time.After(10 * time.Millisecond):
			if got := lockWaiters(); got != 0 {
				t.Fatalf("the page request waits on the queue lock (%d goroutines inside acquireLock)", got)
			}
		}
	}

control:
	// Control: a real contender for the same lock must be visible to the detector.
	contenderDone := make(chan error, 1)
	rest.Add(1)
	go func() {
		defer rest.Done()
		contenderDone <- store.Mutate(func(*factory.BacklogRecord) error { return nil })
	}()
	for lockWaiters() == 0 {
		select {
		case err := <-contenderDone:
			t.Fatalf("the control contender returned (%v) before the detector saw it wait: the detector is blind", err)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// BenchmarkTodoGraphView measures the page request on its own, with the queue
// lock free and with it held, so the time a limit would have to allow is read
// from data rather than guessed: ns/op is the mean, max-ns the slowest request.
// Run it deliberately: go test -race -run '^$' -bench TodoGraphView -benchtime=300x ./internal/web/
func BenchmarkTodoGraphView(b *testing.B) {
	for _, held := range []bool{false, true} {
		name := "lock-free"
		if held {
			name = "lock-held"
		}
		b.Run(name, func(b *testing.B) {
			home := b.TempDir()
			orig := factory.HomeDirFn
			factory.HomeDirFn = func() (string, error) { return home, nil }
			b.Cleanup(func() { factory.HomeDirFn = orig })
			root := b.TempDir()
			path := factory.BacklogPathForRoot(root)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				b.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(graphQueue), 0o600); err != nil {
				b.Fatal(err)
			}
			store := factory.NewBacklogStore(path)
			if held {
				release := make(chan struct{})
				acquired := make(chan struct{})
				done := make(chan error, 1)
				go func() {
					done <- store.Mutate(func(*factory.BacklogRecord) error {
						close(acquired)
						<-release
						return nil
					})
				}()
				<-acquired
				b.Cleanup(func() {
					close(release)
					if err := <-done; err != nil {
						b.Error(err)
					}
				})
			}
			a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
			a.recordLastProfile = func(string) error { return nil }
			var maxNs int64
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := time.Now()
				req := httptest.NewRequest(http.MethodGet, "/todo?view=graph", nil)
				req.Host = "127.0.0.1:8080"
				rec := httptest.NewRecorder()
				a.routes().ServeHTTP(rec, req)
				if d := time.Since(start).Nanoseconds(); d > maxNs {
					maxNs = d
				}
				if rec.Code != http.StatusOK {
					b.Fatalf("status %d", rec.Code)
				}
			}
			b.ReportMetric(float64(maxNs), "max-ns")
		})
	}
}
