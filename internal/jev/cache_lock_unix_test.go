//go:build unix

package jev

// cache_lock_unix_test.go — P2 concurrency (card t1968 codex review round 9,
// operator-approved repair): updateAliasMap's read-modify-write ran unlocked,
// so two concurrent requests during a model switch both read the SAME prior
// mapping and the last writer's map erased the other's Prior append — the
// displaced id was never recorded, and the stale probe later missed an
// actually-saved previous model (offline lookups then fail). The whole update
// must serialize under the cache dir's alias-map lock, so BOTH displaced ids
// survive a concurrent switch. (Unix-only: the lock is an flock, and the
// barrier doer exercises the real file layer; the behavior it pins — the
// alias map's move history — is platform-independent.)

import (
	"context"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// barrierDoer answers every Do with a 200 response naming its model, but only
// after every doer of the round has arrived: Do signals arrival and then
// blocks until the round closes the release channel. Both requests therefore
// resume from the network at the SAME instant and run their alias-map
// read-modify-write concurrently — the exact overlap a model switch races on,
// driven through the real Ask path (lookup miss -> network -> cacheStore ->
// updateAliasMap).
type barrierDoer struct {
	model   string
	release <-chan struct{}
	arrived *sync.WaitGroup
}

func (d *barrierDoer) Do(r *http.Request) (*http.Response, error) {
	d.arrived.Done()
	<-d.release
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(cacheOKBody(d.model))),
		Header:     http.Header{},
		Request:    r,
	}, nil
}

// Two concurrent alias moves from one primed model must both leave their
// displacement in the persisted Prior history: whichever move lands second
// displaces the first move's model, so the final history names BOTH the
// primed id and the first move's id. Without the lock, both updates read the
// primed map and the last writer's map drops the first move's Prior append —
// the history loses an actually-saved previous model. Each round primes a
// fresh cache dir, races one jev-1.20.0 answer against one jev-1.21.0 answer
// under jev-1.19.0, and checks the persisted history; the race is repeated
// because a single unserialized overlap can, by scheduling luck, serialize —
// but a lock cannot lose ANY round.
func TestUpdateAliasMap_ConcurrentMovesKeepBothDisplacedIDs(t *testing.T) {
	const (
		primed = "jev-1.19.0"
		moveB  = "jev-1.20.0"
		moveC  = "jev-1.21.0"
	)
	const rounds = 20

	done := make(chan struct{})
	go func() {
		defer close(done)
		for round := 0; round < rounds; round++ {
			dir := t.TempDir()
			if err := writeModelsMap(dir, modelsMap{Models: map[string]string{ModelID: primed}}); err != nil {
				t.Errorf("round %d: prime: %v", round, err)
				return
			}
			release := make(chan struct{})
			var arrived sync.WaitGroup
			arrived.Add(2)
			mk := func(model string) *Client {
				return &Client{
					Enabled:        true,
					LoadCredential: func() string { return "test-key" },
					HTTP:           &barrierDoer{model: model, release: release, arrived: &arrived},
					CacheDir:       dir,
					MaxRetries:     1,
				}
			}
			results := make([]Result, 2)
			var wg sync.WaitGroup
			wg.Add(2)
			go func() { defer wg.Done(); results[0] = mk(moveB).Ask(context.Background(), cacheQuestion()) }()
			go func() { defer wg.Done(); results[1] = mk(moveC).Ask(context.Background(), cacheQuestion()) }()
			arrived.Wait() // both requests sit inside their network call; resume them together
			close(release)
			wg.Wait()
			for i, res := range results {
				if !res.OK() {
					t.Errorf("round %d: request %d did not complete: %q (%s) — the history check below would be vacuous", round, i, res.Availability, res.Condition)
					return
				}
			}
			m, err := readModelsMap(dir)
			if err != nil {
				t.Errorf("round %d: read back: %v", round, err)
				return
			}
			prior := m.Prior[ModelID]
			hasB, hasC := slices.Contains(prior, moveB), slices.Contains(prior, moveC)
			if len(prior) != 2 || !slices.Contains(prior, primed) || hasB == hasC {
				t.Errorf("round %d: concurrent alias moves lost a displaced id — prior=%v (final model %q); both the primed id and the first move's id must survive in the move history",
					round, prior, m.Models[ModelID])
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the concurrent alias-map updates did not settle — an update holds the alias-map lock without releasing it")
	}
}
