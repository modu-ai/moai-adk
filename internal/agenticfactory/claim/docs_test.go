package claim

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestDocsRoundTripAndCompareAndSwap(t *testing.T) {
	s := open(t)
	if _, err := s.GetDoc("intent", "abc"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing doc: %v", err)
	}
	d, err := s.PutDoc("intent", "abc", 0, []byte(`{"key":"abc","status":"intended"}`))
	if err != nil || d.Rev != 1 || d.Kind != "intent" || d.Name != "abc" {
		t.Fatalf("create: %+v %v", d, err)
	}
	if _, err := s.PutDoc("intent", "abc", 0, []byte(`{}`)); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("second create: %v", err)
	}
	if _, err := s.PutDoc("intent", "abc", 7, []byte(`{}`)); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("wrong revision: %v", err)
	}
	d, err = s.PutDoc("intent", "abc", 1, []byte(`{"key":"abc","status":"sent"}`))
	if err != nil || d.Rev != 2 {
		t.Fatalf("update: %+v %v", d, err)
	}
	got, err := s.GetDoc("intent", "abc")
	if err != nil || got.Rev != 2 || string(got.Data) != `{"key":"abc","status":"sent"}` {
		t.Errorf("get: %+v %v", got, err)
	}
	if err := s.DeleteDoc("intent", "abc", 9); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("delete with a wrong revision: %v", err)
	}
	if err := s.DeleteDoc("intent", "abc", 2); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteDoc("intent", "abc", 2); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("delete of a missing doc: %v", err)
	}
}

func TestDocInputsAreValidated(t *testing.T) {
	s := open(t)
	for _, c := range []struct{ kind, name, data string }{
		{"unknown", "a", `{}`}, {"intent", "", `{}`}, {"intent", "Bad Name", `{}`}, {"intent", "../x", `{}`},
		{"intent", strings.Repeat("a", 65), `{}`}, {"intent", "a", `not json`}, {"intent", "a", ""},
		{"intent", "a", `"` + strings.Repeat("x", maxDocBytes) + `"`},
	} {
		if _, err := s.PutDoc(c.kind, c.name, 0, []byte(c.data)); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("%q/%q: %v", c.kind, c.name, err)
		}
	}
	if _, err := s.GetDoc("intent", "../x"); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("get with a bad name: %v", err)
	}
	if _, err := s.ListDocs("unknown"); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("list of an unknown kind: %v", err)
	}
}

func TestListDocsIsSortedAndPerKind(t *testing.T) {
	s := open(t)
	for _, n := range []string{"c", "a", "b"} {
		if _, err := s.PutDoc("intent", n, 0, []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = s.PutDoc("audit", "issue-1", 0, []byte(`{}`))
	ds, err := s.ListDocs("intent")
	if err != nil || len(ds) != 3 || ds[0].Name != "a" || ds[2].Name != "c" {
		t.Fatalf("list: %+v %v", ds, err)
	}
	if as, _ := s.ListDocs("audit"); len(as) != 1 {
		t.Errorf("kinds must not mix: %+v", as)
	}
	if none, err := s.ListDocs("context"); err != nil || len(none) != 0 {
		t.Errorf("empty kind: %+v %v", none, err)
	}
}

func TestDocsGoToTheAuthorityFirstAndRebuild(t *testing.T) {
	m := NewMemoryRef()
	s1 := attached(t, m)
	if _, err := s1.PutDoc("intent", "abc", 0, []byte(`{"n":1}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s1.PutDoc("audit", "issue-7", 0, []byte(`{"rounds":["r1"]}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s1.PutDoc("context", "issue-7", 0, []byte(`{"items":[]}`)); err != nil {
		t.Fatal(err)
	}
	if data, rev, err := m.Read(context.Background(), "intent/abc"); err != nil || rev != 1 || string(data) != `{"n":1}` {
		t.Fatalf("authority copy: %q %d %v", data, rev, err)
	}
	// The machine-local cache is lost; a second cache rebuilds all three kinds.
	s2 := attached(t, m)
	n, err := s2.Rebuild(context.Background())
	if err != nil || n != 3 {
		t.Fatalf("rebuild: %d %v", n, err)
	}
	for _, c := range []struct{ kind, name string }{{"intent", "abc"}, {"audit", "issue-7"}, {"context", "issue-7"}} {
		if d, err := s2.GetDoc(c.kind, c.name); err != nil || d.Rev != 1 {
			t.Errorf("%s/%s after rebuild: %+v %v", c.kind, c.name, d, err)
		}
	}
	// A write from the second cache continues the same revision line.
	if d, err := s2.PutDoc("intent", "abc", 1, []byte(`{"n":2}`)); err != nil || d.Rev != 2 {
		t.Errorf("write after rebuild: %+v %v", d, err)
	}
}

func TestDocWriteIsRefusedWhenTheAuthorityRefusesOrMoved(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	_, _ = s.PutDoc("intent", "abc", 0, []byte(`{"n":1}`))
	m.Fail = errors.New("401")
	if _, err := s.PutDoc("intent", "abc", 1, []byte(`{"n":2}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("failing authority: %v", err)
	}
	m.Fail = nil
	if d, _ := s.GetDoc("intent", "abc"); d.Rev != 1 || string(d.Data) != `{"n":1}` {
		t.Errorf("cache moved without the authority: %+v", d)
	}
	m.Bump("intent/abc")
	if _, err := s.PutDoc("intent", "abc", 1, []byte(`{"n":2}`)); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("stale cache: %v", err)
	}
}

// Safety documents (intents, audit rounds) are never held back by the
// budget; the context document is a report and waits when the budget is short.
func TestOnlyReportDocsAreDeferredByTheBudget(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	b := &refusingBudget{}
	s.SetBudget(b)
	for _, kind := range []string{"intent", "audit"} {
		if _, err := s.PutDoc(kind, "x", 0, []byte(`{}`)); err != nil {
			t.Errorf("%s must not be deferred: %v", kind, err)
		}
	}
	if _, err := s.PutDoc("context", "issue-1", 0, []byte(`{}`)); !errors.Is(err, core.ErrDeferred) {
		t.Fatalf("context under a short budget: %v", err)
	}
	if _, err := s.GetDoc("context", "issue-1"); !errors.Is(err, core.ErrNotHeld) {
		t.Error("a deferred write left a document")
	}
	if _, _, err := m.Read(context.Background(), "context/issue-1"); !errors.Is(err, core.ErrNotHeld) {
		t.Error("a deferred write reached the authority")
	}
	for _, p := range b.asked {
		if p == core.PriorityNotice {
			t.Error("documents are safety or report, never notice")
		}
	}
}

func TestCorruptDocFileFailsClosed(t *testing.T) {
	s := open(t)
	_, _ = s.PutDoc("intent", "abc", 0, []byte(`{}`))
	if err := os.WriteFile(s.docPath("intent", "abc"), []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetDoc("intent", "abc"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("get: %v", err)
	}
	if _, err := s.ListDocs("intent"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("list: %v", err)
	}
	if _, err := s.PutDoc("intent", "abc", 1, []byte(`{}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("a write over a corrupt doc must not guess its revision: %v", err)
	}
}

// Audit rounds persist in the claim store; the decision stays with the caller.
func TestAuditRoundsPersistAcrossAStoreRestart(t *testing.T) {
	dir := t.TempDir()
	s, _ := OpenWithLimit(dir, 5)
	if rounds, err := s.AuditRounds(1948); err != nil || len(rounds) != 0 {
		t.Fatalf("no rounds yet: %v %v", rounds, err)
	}
	add := func(id string) func([]string) ([]string, error) {
		return func(cur []string) ([]string, error) {
			for _, r := range cur {
				if r == id {
					return cur, nil
				}
			}
			return append(cur, id), nil
		}
	}
	for _, id := range []string{"r1", "r2", "r1"} {
		if _, err := s.UpdateAuditRounds(1948, add(id)); err != nil {
			t.Fatal(err)
		}
	}
	s2, _ := OpenWithLimit(dir, 5)
	rounds, err := s2.AuditRounds(1948)
	if err != nil || fmt.Sprint(rounds) != "[r1 r2]" {
		t.Errorf("after restart: %v %v", rounds, err)
	}
	if other, _ := s2.AuditRounds(2000); len(other) != 0 {
		t.Errorf("rounds are per issue: %v", other)
	}
	boom := errors.New("budget exhausted")
	if _, err := s2.UpdateAuditRounds(1948, func([]string) ([]string, error) { return nil, boom }); !errors.Is(err, boom) {
		t.Errorf("a refusing decision: %v", err)
	}
	if again, _ := s2.AuditRounds(1948); fmt.Sprint(again) != "[r1 r2]" {
		t.Errorf("a refused update changed the rounds: %v", again)
	}
	if _, err := s2.AuditRounds(0); !errors.Is(err, core.ErrIssueInvalid) {
		t.Errorf("issue 0: %v", err)
	}
}

// Concurrent audit starts lose no round and count none twice.
func TestConcurrentAuditRoundsLoseNothing(t *testing.T) {
	s := open(t)
	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("round-%d", i)
			for try := 0; try < 20; try++ {
				_, errs[i] = s.UpdateAuditRounds(1948, func(cur []string) ([]string, error) { return append(cur, id), nil })
				if !errors.Is(errs[i], core.ErrBusy) {
					return
				}
			}
		}(i)
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			t.Fatalf("writer %d: %v", i, e)
		}
	}
	rounds, _ := s.AuditRounds(1948)
	sort.Strings(rounds)
	if len(rounds) != n {
		t.Errorf("rounds = %v, want %d distinct entries", rounds, n)
	}
}

// A cache that is behind the authority reads it again and retries, so two
// machines adding audit rounds do not lose each other's round.
func TestAuditRoundsRefreshAStaleCacheFromTheAuthority(t *testing.T) {
	m := NewMemoryRef()
	s1, s2 := attached(t, m), attached(t, m)
	add := func(id string) func([]string) ([]string, error) {
		return func(cur []string) ([]string, error) { return append(cur, id), nil }
	}
	if _, err := s1.UpdateAuditRounds(1948, add("r1")); err != nil {
		t.Fatal(err)
	}
	// s2 never saw r1: its first attempt conflicts, it pulls, and retries.
	got, err := s2.UpdateAuditRounds(1948, add("r2"))
	if err != nil || fmt.Sprint(got) != "[r1 r2]" {
		t.Fatalf("second machine: %v %v", got, err)
	}
	if again, _ := s1.UpdateAuditRounds(1948, add("r3")); fmt.Sprint(again) != "[r1 r2 r3]" {
		t.Errorf("first machine after the other's write: %v", again)
	}
}

func TestDeleteGoesThroughTheAuthority(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	_, _ = s.PutDoc("intent", "issue-5", 0, []byte(`{"user_id":7}`))
	m.Fail = errors.New("down")
	if err := s.DeleteDoc("intent", "issue-5", 1); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Fatalf("delete with a failing authority: %v", err)
	}
	if _, err := s.GetDoc("intent", "issue-5"); err != nil {
		t.Errorf("the cache lost a document the authority still has: %v", err)
	}
	m.Fail = nil
	if err := s.DeleteDoc("intent", "issue-5", 1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.Read(context.Background(), "intent/issue-5"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("the authority still holds it: %v", err)
	}
}

func TestPullDropsADocumentTheAuthorityNoLongerHas(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	_, _ = s.PutDoc("intent", "abc", 0, []byte(`{}`))
	if err := m.Delete(context.Background(), "intent/abc", 1); err != nil {
		t.Fatal(err)
	}
	if err := s.pullDoc("intent", "abc"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetDoc("intent", "abc"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("stale cache entry survived: %v", err)
	}
	// With no authority a pull is a no-op.
	if err := open(t).pullDoc("intent", "abc"); err != nil {
		t.Errorf("detached pull: %v", err)
	}
}

func TestRebuildSkipsUnknownKindsAndRefusesBadDocs(t *testing.T) {
	m := NewMemoryRef()
	_, _ = m.CompareAndSwap(context.Background(), "unknownkind/x", 0, []byte(`{}`))
	s := attached(t, m)
	if n, err := s.Rebuild(context.Background()); err != nil || n != 0 {
		t.Errorf("unknown kinds are skipped: %d %v", n, err)
	}
	_, _ = m.CompareAndSwap(context.Background(), "intent/Bad Name", 0, []byte(`{}`))
	if _, err := s.Rebuild(context.Background()); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("an authority document with an invalid name: %v", err)
	}
}

func TestUpdateDocCreatesAppliesAndAborts(t *testing.T) {
	s := open(t)
	d, err := s.UpdateDoc("context", "issue-1", func(cur []byte) ([]byte, error) {
		if cur != nil {
			t.Errorf("a missing document is passed as nil, got %q", cur)
		}
		return []byte(`{"n":1}`), nil
	})
	if err != nil || d.Rev != 1 {
		t.Fatalf("create: %+v %v", d, err)
	}
	d, err = s.UpdateDoc("context", "issue-1", func(cur []byte) ([]byte, error) {
		if string(cur) != `{"n":1}` {
			t.Errorf("current data = %q", cur)
		}
		return []byte(`{"n":2}`), nil
	})
	if err != nil || d.Rev != 2 {
		t.Fatalf("update: %+v %v", d, err)
	}
	boom := errors.New("refused")
	if _, err := s.UpdateDoc("context", "issue-1", func([]byte) ([]byte, error) { return nil, boom }); !errors.Is(err, boom) {
		t.Errorf("fn error: %v", err)
	}
	if got, _ := s.GetDoc("context", "issue-1"); got.Rev != 2 {
		t.Errorf("an aborted update changed the document: %+v", got)
	}
	if _, err := s.UpdateDoc("nope", "x", func([]byte) ([]byte, error) { return []byte(`{}`), nil }); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("unknown kind: %v", err)
	}
}

func TestUpdateDocReadsTheAuthorityAgainOnAConflict(t *testing.T) {
	m := NewMemoryRef()
	s1, s2 := attached(t, m), attached(t, m)
	if _, err := s1.UpdateDoc("context", "issue-1", func([]byte) ([]byte, error) { return []byte(`{"by":"s1"}`), nil }); err != nil {
		t.Fatal(err)
	}
	var seen string
	d, err := s2.UpdateDoc("context", "issue-1", func(cur []byte) ([]byte, error) {
		seen = string(cur)
		return []byte(`{"by":"s2"}`), nil
	})
	if err != nil || d.Rev != 2 || seen != `{"by":"s1"}` {
		t.Fatalf("second machine: %+v %v seen=%q", d, err, seen)
	}
}
