package docstore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

type onceSender struct{ n atomic.Int32 }

func (s *onceSender) Send(context.Context, effect.Intent) (string, error) {
	s.n.Add(1)
	return "issue-9", nil
}

type okLookup struct{}

func (okLookup) Lookup(context.Context, effect.Intent) (effect.LookupResult, error) {
	return effect.LookupResult{Found: true, ProviderID: "issue-9", Authoritative: true}, nil
}

func TestPutIfIsCreateOnlyThenCompareAndSwap(t *testing.T) {
	s := open(t, t.TempDir())
	in := effect.Intent{Key: "aa11", Kind: effect.IssueCreate, Status: effect.StatusIntended}
	created, err := s.PutIf(in)
	if err != nil || created.Rev <= 0 {
		t.Fatalf("create: %+v %v", created, err)
	}
	if _, err := s.PutIf(in); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("a second create-only write: %v, want a conflict", err)
	}
	got, ok, _ := s.Get("aa11")
	if !ok || got.Rev != created.Rev {
		t.Fatalf("Get must carry the revision: %+v want %d", got, created.Rev)
	}
	got.Status = effect.StatusSent
	next, err := s.PutIf(got)
	if err != nil || next.Rev <= got.Rev {
		t.Fatalf("update: %+v %v", next, err)
	}
	if _, err := s.PutIf(got); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("a write on a stale revision: %v, want a conflict", err)
	}
}

// Two executors share one claim store and both read the same Failed intent before either writes.
func TestTwoExecutorsOverTheClaimStoreSendOnce(t *testing.T) {
	s := open(t, t.TempDir())
	key := effect.Key(effect.IssueCreate, "repo:o/r", "d1")
	if err := s.Put(effect.Intent{Key: key, Kind: effect.IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1", Status: effect.StatusFailed, Attempts: 1}); err != nil {
		t.Fatal(err)
	}
	var arrived sync.WaitGroup
	arrived.Add(2)
	d := time.Second
	to := effect.Timeouts{IssueCreate: d, BranchPrepare: d, WorktreePrepare: d, CheckObserve: d, LedgerComment: d, LabelApply: d}
	sa, sb := &onceSender{}, &onceSender{}
	var done sync.WaitGroup
	for _, snd := range []*onceSender{sa, sb} {
		done.Add(1)
		go func() {
			defer done.Done()
			var first sync.Once
			st := barrierStore{s, func() { first.Do(func() { arrived.Done(); arrived.Wait() }) }}
			e := &effect.Executor{Store: st, Send: snd, Lookup: okLookup{}, Timeouts: to}
			if _, err := e.Execute(t.Context(), effect.Request{Kind: effect.IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1"}); err != nil {
				t.Errorf("execute: %v", err)
			}
		}()
	}
	done.Wait()
	if got := sa.n.Load() + sb.n.Load(); got != 1 {
		t.Fatalf("sends = %d, want exactly one", got)
	}
}

type barrierStore struct {
	Store
	wait func()
}

func (b barrierStore) Get(k string) (effect.Intent, bool, error) {
	in, ok, err := b.Store.Get(k)
	b.wait()
	return in, ok, err
}
