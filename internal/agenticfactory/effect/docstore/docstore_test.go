package docstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

func open(t *testing.T, dir string) Store {
	t.Helper()
	c, err := claim.OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	return Store{C: c}
}

func TestRoundTripListAndMissing(t *testing.T) {
	s := open(t, t.TempDir())
	if _, found, err := s.Get("deadbeef"); found || err != nil {
		t.Fatalf("missing: found=%v err=%v", found, err)
	}
	a := effect.Intent{Key: effect.Key(effect.IssueCreate, "repo:o/r", "d1"), Kind: effect.IssueCreate, Issue: 7, Status: effect.StatusSent, Attempts: 1}
	b := effect.Intent{Key: effect.Key(effect.LedgerComment, "issue-7", "d2"), Kind: effect.LedgerComment, Issue: 7, Status: effect.StatusIntended, Payload: "masked text", Seq: 2}
	for _, in := range []effect.Intent{a, b} {
		if err := s.Put(in); err != nil {
			t.Fatal(err)
		}
	}
	got, found, err := s.Get(a.Key)
	if got.Rev <= 0 {
		t.Errorf("Get must carry the document revision: %+v", got)
	}
	got.Rev = 0 // the revision is the store's, not part of what was written
	if err != nil || !found || got != a {
		t.Fatalf("get = %+v found=%v err=%v", got, found, err)
	}
	a.Status, a.ProviderID = effect.StatusConfirmed, "321"
	if err := s.Put(a); err != nil { // a second write moves the revision on
		t.Fatal(err)
	}
	if got, _, _ := s.Get(a.Key); got.Status != effect.StatusConfirmed || got.ProviderID != "321" {
		t.Errorf("update lost: %+v", got)
	}
	all, err := s.List()
	if err != nil || len(all) != 2 {
		t.Fatalf("list = %+v err=%v", all, err)
	}
}

// The reason this exists: after a restart the intent is still there, so the executor reconciles a sent
// intent from facts instead of sending it a second time.
func TestAnIntentSurvivesARestartAndIsNotSentAgain(t *testing.T) {
	dir := t.TempDir()
	key := effect.Key(effect.IssueCreate, "repo:o/r", "d1")
	first := open(t, dir)
	if err := first.Put(effect.Intent{Key: key, Kind: effect.IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1", Status: effect.StatusSent, Attempts: 1}); err != nil {
		t.Fatal(err)
	}

	restarted := open(t, dir) // a new process: nothing in memory
	s := &countingSender{}
	lk := &foundLookup{}
	ex := &effect.Executor{Store: restarted, Send: s, Lookup: lk, Timeouts: effect.Timeouts{IssueCreate: time.Second}}
	in, err := ex.Execute(context.Background(), effect.Request{Kind: effect.IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1"})
	if err != nil || in.Status != effect.StatusConfirmed || in.ProviderID != "321" || s.calls != 0 {
		t.Fatalf("in=%+v err=%v sends=%d", in, err, s.calls)
	}
	again := open(t, dir)
	if got, _, _ := again.Get(key); got.Status != effect.StatusConfirmed {
		t.Errorf("the reconciled state was not persisted: %+v", got)
	}
}

type countingSender struct{ calls int }

func (c *countingSender) Send(context.Context, effect.Intent) (string, error) {
	c.calls++
	return "x", nil
}

type foundLookup struct{}

func (foundLookup) Lookup(context.Context, effect.Intent) (effect.LookupResult, error) {
	return effect.LookupResult{Found: true, ProviderID: "321", Authoritative: true}, nil
}

func TestABadKeyOrCorruptDocumentIsAnErrorNotASilentMiss(t *testing.T) {
	s := open(t, t.TempDir())
	if err := s.Put(effect.Intent{Key: "Not A Valid Name!", Kind: effect.IssueCreate}); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("bad key: err = %v", err)
	}
	if _, _, err := s.Get("Not A Valid Name!"); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("bad key on get: err = %v", err)
	}
	// a document that holds a different intent than its name says
	if _, err := s.C.PutDoc("intent", "aaaa", 0, []byte(`{"key":"bbbb","kind":"issue-create"}`)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Get("aaaa"); err == nil {
		t.Error("a document under the wrong name was accepted")
	}
	if _, err := s.C.PutDoc("intent", "cccc", 0, []byte(`"not an object"`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.List(); err == nil {
		t.Error("an unreadable intent was skipped silently")
	}
}

func TestAMovedRevisionIsReportedNotOverwritten(t *testing.T) {
	s := open(t, t.TempDir())
	in := effect.Intent{Key: "abcd", Kind: effect.IssueCreate, Status: effect.StatusIntended}
	if err := s.Put(in); err != nil {
		t.Fatal(err)
	}
	// another writer moves the revision between our read and write: simulate with a wrapper that reads first
	cur, _ := s.C.GetDoc("intent", "abcd")
	if _, err := s.C.PutDoc("intent", "abcd", cur.Rev, []byte(`{"key":"abcd","kind":"issue-create","status":"confirmed"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.C.PutDoc("intent", "abcd", cur.Rev, []byte(`{"key":"abcd"}`)); !errors.Is(err, core.ErrRevisionConflict) {
		t.Fatalf("the claim store did not report the conflict: %v", err)
	}
	if got, _, _ := s.Get("abcd"); got.Status != effect.StatusConfirmed {
		t.Errorf("the other writer's state was overwritten: %+v", got)
	}
}
