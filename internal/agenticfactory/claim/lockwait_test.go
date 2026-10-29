package claim

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// A deterministic harness for the lock wait. The writers are real stores over one
// directory taking the real O_EXCL lock; only time is virtual. Every sleep is a
// stop: the scheduler moves the clock to the earliest waking writer and lets that
// one writer run until it sleeps or finishes, so the outcome does not depend on
// the machine's load, its clock or goroutine scheduling. A critical section of
// "hold" is a sleep taken while the lock is held (what an fsync under the lock
// costs on a loaded disk), and "gap" is the time a writer works between two
// operations.

type vsleeper struct {
	id   int
	wake time.Duration
	ch   chan struct{}
}

type vclock struct {
	mu       sync.Mutex
	cond     *sync.Cond
	now      time.Duration
	running  int
	sleepers []*vsleeper
}

func newVclock(writers int) *vclock {
	c := &vclock{running: writers}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *vclock) sleeper(id int) func(time.Duration) {
	return func(d time.Duration) {
		c.mu.Lock()
		s := &vsleeper{id: id, wake: c.now + d, ch: make(chan struct{})}
		c.sleepers = append(c.sleepers, s)
		c.running--
		c.cond.Broadcast()
		c.mu.Unlock()
		<-s.ch
	}
}

func (c *vclock) at() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *vclock) finished() {
	c.mu.Lock()
	c.running--
	c.cond.Broadcast()
	c.mu.Unlock()
}

// run lets the writers go one at a time, in virtual time order (ties by writer id).
func (c *vclock) run() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for {
		for c.running > 0 {
			c.cond.Wait()
		}
		if len(c.sleepers) == 0 {
			return
		}
		sort.Slice(c.sleepers, func(i, j int) bool {
			a, b := c.sleepers[i], c.sleepers[j]
			if a.wake != b.wake {
				return a.wake < b.wake
			}
			return a.id < b.id
		})
		s := c.sleepers[0]
		c.sleepers = c.sleepers[1:]
		if s.wake > c.now {
			c.now = s.wake
		}
		c.running++
		close(s.ch)
	}
}

type lockResult struct {
	ops, busy int
	waits     []time.Duration // virtual wait of every acquired lock
}

func (r lockResult) pct(p float64) time.Duration {
	if len(r.waits) == 0 {
		return 0
	}
	w := append([]time.Duration(nil), r.waits...)
	sort.Slice(w, func(i, j int) bool { return w[i] < w[j] })
	return w[int(float64(len(w)-1)*p)]
}

func (r lockResult) String() string {
	return fmt.Sprintf("ops=%d busy=%d (%.1f%%) wait p50=%v p95=%v max=%v", r.ops, r.busy, 100*float64(r.busy)/float64(r.ops), r.pct(.5), r.pct(.95), r.pct(1))
}

// contend runs writers writers, each doing ops operations that hold the lock for
// hold and work for gap between them, under policy pol (zero = the built-in).
func contend(t *testing.T, writers, ops int, hold, gap time.Duration, pol lockPolicy) lockResult {
	return contendStalled(t, writers, ops, hold, gap, 0, 0, pol)
}

// contendStalled is contend where every stallEvery-th operation of the whole run
// holds the lock for stall instead of hold (a holder descheduled, or an fsync that
// takes long on a busy disk), and where the writers' gaps differ by 3ms each so that
// their polling phases do not stay aligned.
func contendStalled(t *testing.T, writers, ops int, hold, gap time.Duration, stallEvery int, stall time.Duration, pol lockPolicy) lockResult {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "documents.lock")
	c := newVclock(writers)
	var mu sync.Mutex
	var res lockResult
	var held int
	var wg sync.WaitGroup
	for id := 0; id < writers; id++ {
		st, err := Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		sleep := c.sleeper(id)
		st.lockPol = pol
		st.lockPol.sleep = sleep
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer c.finished()
			sleep(0) // all writers start at time 0: take the first turn in id order, not in the order the OS starts the goroutines
			for n := 0; n < ops; n++ {
				start := c.at()
				var waited time.Duration
				err := st.withLockFile(path, "documents", func() error {
					waited = c.at() - start
					h := hold
					held++
					if stallEvery > 0 && held%stallEvery == 0 {
						h = stall
					}
					sleep(h)
					return nil
				})
				mu.Lock()
				res.ops++
				switch {
				case err == nil:
					res.waits = append(res.waits, waited)
				case errors.Is(err, core.ErrBusy):
					res.busy++
				default:
					t.Errorf("unexpected: %v", err)
				}
				mu.Unlock()
				sleep(gap + time.Duration(id)*3*time.Millisecond)
			}
		}()
	}
	c.run()
	wg.Wait()
	return res
}

// The measurement is the same on every run and every machine.
func TestTheLockWaitHarnessIsDeterministic(t *testing.T) {
	a := contend(t, 4, 60, 30*time.Millisecond, 10*time.Millisecond, lockPolicy{})
	b := contend(t, 4, 60, 30*time.Millisecond, 10*time.Millisecond, lockPolicy{})
	if a.String() != b.String() {
		t.Errorf("two runs differ:\n%s\n%s", a, b)
	}
}

// Baseline of the built-in wait, to be compared with any change under the same
// harness and the same parameters. Run with -v to read the table.
func TestLockWaitMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("matrix")
	}
	const ms = time.Millisecond
	for _, hold := range []time.Duration{5 * ms, 30 * ms, 100 * ms, 200 * ms} {
		for _, w := range []int{1, 2, 4} {
			t.Logf("steady  hold=%-6v writers=%d  %s", hold, w, contend(t, w, 200, hold, 10*ms, lockPolicy{}))
		}
	}
	for _, w := range []int{1, 2, 4} {
		t.Logf("stalled hold=5ms stall=450ms/every 50 writers=%d  %s", w, contendStalled(t, w, 200, 5*ms, 10*ms, 50, 450*ms, lockPolicy{}))
	}
}

// The default wait must carry a writer through the stalls that were measured: four
// resident writers (the main loop, notify, webhook, the collector) on one documents
// lock, where a holder is slow (a long fsync on a busy disk) or descheduled for up
// to the 450ms lane-20 saw, or a steady 200ms hold. In none of these may a
// well-behaved writer be told Busy. The numbers come from the harness above, not
// from the machine: they are the same on every run.
func TestTheDefaultLockWaitSurvivesTheMeasuredStalls(t *testing.T) {
	const ms = time.Millisecond
	for name, run := range map[string]func() lockResult{
		"steady 100ms hold, 4 writers": func() lockResult { return contend(t, 4, 200, 100*ms, 10*ms, lockPolicy{}) },
		"steady 200ms hold, 4 writers": func() lockResult { return contend(t, 4, 200, 200*ms, 10*ms, lockPolicy{}) },
		"450ms stalls, 2 writers":      func() lockResult { return contendStalled(t, 2, 200, 5*ms, 10*ms, 50, 450*ms, lockPolicy{}) },
		"450ms stalls, 4 writers":      func() lockResult { return contendStalled(t, 4, 200, 5*ms, 10*ms, 50, 450*ms, lockPolicy{}) },
		"1500ms stalls, 4 writers":     func() lockResult { return contendStalled(t, 4, 200, 5*ms, 10*ms, 50, 1500*ms, lockPolicy{}) },
	} {
		if r := run(); r.busy != 0 {
			t.Errorf("%s: %s", name, r)
		}
	}
}

// The wait is finite and its end is Busy, never Held: a lock whose holder stays
// away longer than the whole budget ends the wait with ErrBusy, which is the
// contract's retry signal, and not with the ownership error.
func TestAnExhaustedLockWaitIsBusyAndNeverHeld(t *testing.T) {
	const ms = time.Millisecond
	r := contendStalled(t, 2, 4, 5*ms, 10*ms, 2, time.Hour, lockPolicy{})
	if r.busy == 0 {
		t.Fatalf("a holder that never returns must exhaust the wait: %s", r)
	}
}
