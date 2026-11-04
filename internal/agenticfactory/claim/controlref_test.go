package claim

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func attached(t *testing.T, ref ControlRef) *Store {
	t.Helper()
	s := open(t)
	s.Attach(ref)
	return s
}

func TestWriteGoesToTheAuthorityFirstAndRevisionAdvances(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	rec, err := s.Acquire(ref(t, 7), "w", "d", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	data, rev, err := m.Read(context.Background(), "issue/7")
	if err != nil || rev != 1 {
		t.Fatalf("authority copy: rev=%d %v", rev, err)
	}
	var onRef Record
	if err := json.Unmarshal(data, &onRef); err != nil || onRef.Holder != "w" || onRef.Rev != 0 {
		t.Errorf("authority payload: %+v %v", onRef, err)
	}
	if rec.Rev != 1 {
		t.Errorf("returned record must carry the revision, got %d", rec.Rev)
	}
	out, err := s.Update(ref(t, 7), "w", rec.Epoch, func(r *Record) error { r.State = core.StateImplementing; return nil })
	if err != nil || out.Rev != 2 {
		t.Fatalf("update: %+v %v", out, err)
	}
	if cached, _ := s.Load(ref(t, 7)); cached.Rev != 2 {
		t.Errorf("cache rev = %d", cached.Rev)
	}
}

func TestAuthorityFailureLeavesTheCacheUntouched(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	rec, _ := s.Acquire(ref(t, 7), "w", "d", "run-1")
	m.Fail = errors.New("401 bad credentials")
	if _, err := s.Update(ref(t, 7), "w", rec.Epoch, func(r *Record) error { r.State = core.StateImplementing; return nil }); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Fatalf("update with failing authority: %v", err)
	}
	if _, err := s.Acquire(ref(t, 8), "w", "d8", "run-1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("acquire with failing authority: %v", err)
	}
	m.Fail = nil
	if cur, _ := s.Load(ref(t, 7)); cur.State != core.StateClaimed || cur.Rev != 1 {
		t.Errorf("cache moved without the authority: %+v", cur)
	}
	if _, err := s.Load(ref(t, 8)); !errors.Is(err, core.ErrNotHeld) {
		t.Error("a claim the authority refused left a cache record")
	}
}

func TestStaleCacheCannotOverwriteAnotherWritersRevision(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	rec, _ := s.Acquire(ref(t, 7), "w", "d", "run-1")
	m.Bump("issue/7") // someone else advanced the authoritative copy
	_, err := s.Update(ref(t, 7), "w", rec.Epoch, func(r *Record) error { r.State = core.StateImplementing; return nil })
	if !errors.Is(err, core.ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
	if cur, _ := s.Load(ref(t, 7)); cur.State != core.StateClaimed {
		t.Errorf("cache changed on a conflict: %+v", cur)
	}
}

func TestRebuildRestoresTheCacheFromTheAuthority(t *testing.T) {
	m := NewMemoryRef()
	s1 := attached(t, m)
	a, _ := s1.RegisterAgent(sid(1))
	rec, err := s1.AcquireFor(ref(t, 7), a, "d", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = s1.PutRequest("req-1", "h", 4, pol)
	_, _ = s1.Update(ref(t, 7), a.SessionID, rec.Epoch, func(r *Record) error { r.State = core.StateImplementing; r.AttemptSeq = 1; return nil })

	// The machine-local cache is lost; a new Store points at the same authority.
	s2 := attached(t, m)
	if _, err := s2.Load(ref(t, 7)); !errors.Is(err, core.ErrNotHeld) {
		t.Fatal("test setup: the second cache must start empty")
	}
	n, err := s2.Rebuild(context.Background())
	if err != nil || n != 3 {
		t.Fatalf("rebuild restored %d: %v", n, err)
	}
	got, err := s2.Load(ref(t, 7))
	if err != nil || got.State != core.StateImplementing || got.Holder != a.SessionID || got.AgentName != "agent-1" || got.Rev != 2 {
		t.Fatalf("restored record: %+v %v", got, err)
	}
	if _, err := s2.Verify(ref(t, 7), a.SessionID, rec.Epoch); err != nil {
		t.Errorf("ownership must verify after rebuild: %v", err)
	}
	if r, err := s2.GetRequest("req-1"); err != nil || r.Priority != 4 {
		t.Errorf("restored request: %+v %v", r, err)
	}
	if ag, _, err := s2.ResumeAgent(a.SessionID, a.Number); err != nil || ag != a {
		t.Errorf("restored agent: %+v %v", ag, err)
	}
	// Rebuild is repeatable, and a later write continues from the same revision.
	if _, err := s2.Rebuild(context.Background()); err != nil {
		t.Fatal(err)
	}
	if out, err := s2.Update(ref(t, 7), a.SessionID, rec.Epoch, func(r *Record) error { r.State = core.StateVerifying; return nil }); err != nil || out.Rev != 3 {
		t.Errorf("write after rebuild: %+v %v", out, err)
	}
}

func TestRebuildFailsClosedOnUnreadableAuthority(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	_, _ = s.Acquire(ref(t, 7), "w", "d", "run-1")
	m.Fail = errors.New("429")
	if _, err := s.Rebuild(context.Background()); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("rebuild with failing authority: %v", err)
	}
	m.Fail = nil
	_, _ = m.CompareAndSwap(context.Background(), "issue/9", 0, []byte("{not json"))
	if _, err := s.Rebuild(context.Background()); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("rebuild over corrupt authority data: %v", err)
	}
	if _, err := s.Load(ref(t, 9)); !errors.Is(err, core.ErrNotHeld) {
		t.Error("corrupt authority data reached the cache")
	}
}

// Two caches on two machines share one authority: agent numbers stay unique.
func TestAgentNumbersAreUniqueAcrossCaches(t *testing.T) {
	m := NewMemoryRef()
	stores := []*Store{attached(t, m), attached(t, m), attached(t, m)}
	const n = 18
	var wg sync.WaitGroup
	got := make([]core.Agent, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i], errs[i] = stores[i%3].RegisterAgent(sid(i + 1))
		}(i)
	}
	wg.Wait()
	seen := map[int]string{}
	for i, a := range got {
		if errs[i] != nil {
			t.Fatalf("register %d: %v", i, errs[i])
		}
		if prev, dup := seen[a.Number]; dup {
			t.Fatalf("number %d given to %s and %s", a.Number, prev, a.SessionID)
		}
		seen[a.Number] = a.SessionID
	}
}

func TestReleasedAgentNumberIsFreedInTheAuthorityToo(t *testing.T) {
	m := NewMemoryRef()
	s1, s2 := attached(t, m), attached(t, m)
	a, _ := s1.RegisterAgent(sid(1))
	if err := s1.ReleaseAgent(a.SessionID, a.Number); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.Read(context.Background(), "agent/1"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("released number still in the authority: %v", err)
	}
	b, err := s2.RegisterAgent(sid(2))
	if err != nil || b.Number != 1 {
		t.Errorf("reuse through the other cache: %+v %v", b, err)
	}
}

func TestMemoryRefContract(t *testing.T) {
	m := NewMemoryRef()
	ctx := context.Background()
	if _, _, err := m.Read(ctx, "k"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing key: %v", err)
	}
	if _, err := m.CompareAndSwap(ctx, "k", 5, []byte("x")); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("create with a wrong expectation: %v", err)
	}
	rev, err := m.CompareAndSwap(ctx, "k", 0, []byte("x"))
	if err != nil || rev != 1 {
		t.Fatalf("create: %d %v", rev, err)
	}
	if _, err := m.CompareAndSwap(ctx, "k", 0, []byte("y")); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("second create: %v", err)
	}
	keys, _ := m.Keys(ctx, "k")
	if len(keys) != 1 {
		t.Errorf("keys: %v", keys)
	}
	if err := m.Delete(ctx, "k", 9); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("delete with wrong revision: %v", err)
	}
	if err := m.Delete(ctx, "k", 1); err != nil {
		t.Errorf("delete: %v", err)
	}
	m.Fail = errors.New("down")
	if _, err := m.Keys(ctx, ""); err == nil {
		t.Error("failure injection did not apply to Keys")
	}
}

// A claim held through another cache is visible only in the authority; it must
// still stop the agent's number from being released.
func TestReleaseSeesClaimsHeldThroughAnotherCache(t *testing.T) {
	m := NewMemoryRef()
	s1, s2 := attached(t, m), attached(t, m)
	a, _ := s1.RegisterAgent(sid(1))
	if _, err := s2.AcquireFor(ref(t, 7), a, "d", "run-1"); err != nil {
		t.Fatalf("claim through the second cache: %v", err)
	}
	if err := s1.ReleaseAgent(a.SessionID, a.Number); !errors.Is(err, core.ErrHeld) {
		t.Fatalf("release must see the other cache's claim: %v", err)
	}
	if _, _, err := s1.ResumeAgent(a.SessionID, a.Number); err != nil {
		t.Errorf("identity damaged by a refused release: %v", err)
	}
}

// An identity released through one cache cannot claim through another, even
// though that cache's own file still lists it.
func TestReleasedIdentityIsRefusedByEveryCache(t *testing.T) {
	m := NewMemoryRef()
	s1, s2 := attached(t, m), attached(t, m)
	a, _ := s1.RegisterAgent(sid(1))
	if _, err := s2.Rebuild(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s1.ReleaseAgent(a.SessionID, a.Number); err != nil {
		t.Fatal(err)
	}
	if _, err := s2.AcquireFor(ref(t, 9), a, "d", "run-1"); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("stale cache accepted a released identity: %v", err)
	}
	if _, err := s2.Load(ref(t, 9)); !errors.Is(err, core.ErrNotHeld) {
		t.Error("a refused claim left a record")
	}
}

func TestDetachedStoreHasNoAuthority(t *testing.T) {
	s := open(t)
	if _, err := s.Rebuild(context.Background()); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("rebuild without an authority: %v", err)
	}
	s.Attach(NewMemoryRef())
	s.Attach(nil)
	if _, err := s.Acquire(ref(t, 3), "w", "d", "run-1"); err != nil {
		t.Errorf("a detached store works from its cache: %v", err)
	}
}
