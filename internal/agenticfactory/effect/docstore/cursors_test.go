package docstore

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// memDocs is a stand-in for the claim store's documents so the test can use a kind the real store lists.
type memDocs struct {
	docs map[string]claim.Doc
}

func (m *memDocs) GetDoc(kind, name string) (claim.Doc, error) {
	d, ok := m.docs[kind+"/"+name]
	if !ok {
		return claim.Doc{}, core.ErrNotHeld
	}
	return d, nil
}

func (m *memDocs) PutDoc(kind, name string, expect int64, data []byte) (claim.Doc, error) {
	if m.docs == nil {
		m.docs = map[string]claim.Doc{}
	}
	cur := m.docs[kind+"/"+name]
	if cur.Rev != expect {
		return claim.Doc{}, core.ErrRevisionConflict
	}
	d := claim.Doc{Kind: kind, Name: name, Rev: cur.Rev + 1, Data: append([]byte(nil), data...)}
	m.docs[kind+"/"+name] = d
	return d, nil
}

func TestACursorRoundTripsAndUpdatesInPlace(t *testing.T) {
	c := Cursors{Docs: &memDocs{}, Kind: "collector"}
	if _, found, err := c.LoadCursor("o/r/pulls"); found || err != nil {
		t.Fatalf("empty: found=%v err=%v", found, err)
	}
	if err := c.SaveCursor("o/r/pulls", []byte(`[{"number":1}]`)); err != nil {
		t.Fatal(err)
	}
	if err := c.SaveCursor("o/r/pulls", []byte(`[{"number":1},{"number":2}]`)); err != nil {
		t.Fatal(err)
	}
	got, found, err := c.LoadCursor("o/r/pulls")
	if err != nil || !found || string(got) != `[{"number":1},{"number":2}]` {
		t.Errorf("got %q found=%v err=%v", got, found, err)
	}
	if _, found, _ := c.LoadCursor("o/r/other"); found {
		t.Error("another key found this cursor")
	}
}

func TestNoConfiguredKindIsAnErrorNotAnEmptyCursor(t *testing.T) {
	c := Cursors{Docs: &memDocs{}}
	if _, _, err := c.LoadCursor("k"); err == nil {
		t.Error("load without a kind returned no error")
	}
	if err := c.SaveCursor("k", []byte(`1`)); err == nil {
		t.Error("save without a kind returned no error")
	}
}

func TestADocumentUnderTheWrongKeyIsRefused(t *testing.T) {
	m := &memDocs{}
	c := Cursors{Docs: m, Kind: "collector"}
	if err := c.SaveCursor("a", []byte(`1`)); err != nil {
		t.Fatal(err)
	}
	d := m.docs["collector/"+docName("a")]
	m.docs["collector/"+docName("b")] = d // the same bytes filed under another key's name
	if _, _, err := c.LoadCursor("b"); err == nil {
		t.Error("a cursor filed under another key was accepted")
	}
}

// The real claim store lists the collector kind, so the cursor round-trips through real documents.
func TestTheRealClaimStoreKeepsTheCursorAcrossAReopen(t *testing.T) {
	dir := t.TempDir()
	open := func() Cursors {
		st, err := claim.OpenWithLimit(dir, 100)
		if err != nil {
			t.Fatal(err)
		}
		return CollectorCursors(st)
	}
	first := open()
	if _, found, err := first.LoadCursor("o/r/pulls"); found || err != nil {
		t.Fatalf("empty store: found=%v err=%v", found, err)
	}
	if err := first.SaveCursor("o/r/pulls", []byte(`[{"number":1},{"number":2}]`)); err != nil {
		t.Fatalf("a real save must succeed now that the collector kind exists: %v", err)
	}
	if err := first.SaveCursor("o/r/pulls", []byte(`[{"number":1},{"number":2},{"number":3}]`)); err != nil {
		t.Fatalf("the second save (a new revision) failed: %v", err)
	}
	got, found, err := open().LoadCursor("o/r/pulls") // a new Store over the same directory: a restart
	if err != nil || !found || string(got) != `[{"number":1},{"number":2},{"number":3}]` {
		t.Errorf("after a reopen: %q found=%v err=%v", got, found, err)
	}
}

// End to end on real documents: ack, restart, only the real change is reported.
func TestACollectorResumesFromARealStoredCursorAfterARestart(t *testing.T) {
	dir := t.TempDir()
	open := func() Cursors {
		st, err := claim.OpenWithLimit(dir, 100)
		if err != nil {
			t.Fatal(err)
		}
		return CollectorCursors(st)
	}
	body1 := `[{"number":1,"s":"open"},{"number":2,"s":"open"}]`
	body2 := `[{"number":1,"s":"merged"},{"number":2,"s":"open"}]`
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	mk := func(body string) *forge.Collector {
		tr := &oneShot{body: body}
		return &forge.Collector{
			Client: &forge.Client{T: tr, B: &forge.Budget{}, Now: func() time.Time { return now },
				Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
			Path: "/repos/o/r/pulls", Interval: forge.Adaptive{Min: 10 * time.Second, Max: 5 * time.Minute, Safety: 30 * time.Minute},
			Differ: forge.DiffByKey("number"), Cursors: open(), CursorKey: "o/r/pulls",
		}
	}
	first := mk(body1)
	rep := first.Tick(t.Context(), now)
	if !rep.Baseline || len(rep.Changes) != 2 {
		t.Fatalf("first report = %+v", rep)
	}
	if err := first.Ack(rep); err != nil {
		t.Fatalf("a real ack failed: %v", err)
	}
	restarted := mk(body2)
	r2 := restarted.Tick(t.Context(), now)
	if r2.Baseline || len(r2.Changes) != 1 || r2.Changes[0].Key != "1" || r2.Changes[0].Kind != forge.ChangeModified {
		t.Errorf("after the restart: %+v, want only issue 1 modified and no baseline", r2)
	}
}

type oneShot struct{ body string }

func (o *oneShot) Do(context.Context, forge.Request) (forge.Response, error) {
	return forge.Response{Status: 200, Body: []byte(o.body)}, nil
}

// A snapshot too large for one document is refused at the ack; nothing advances and the changes come back.
func TestAnOversizeSnapshotIsRefusedAtTheAckAndNothingAdvances(t *testing.T) {
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	big := `[{"number":1,"pad":"` + strings.Repeat("x", 300<<10) + `"}]`
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	col := &forge.Collector{
		Client: &forge.Client{T: &oneShot{body: big}, B: &forge.Budget{}, Now: func() time.Time { return now },
			Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
		Path: "/p", Interval: forge.Adaptive{Min: time.Second, Max: time.Minute, Safety: time.Hour},
		Differ: forge.DiffByKey("number"), Cursors: CollectorCursors(st), CursorKey: "o/r/big",
	}
	rep := col.Tick(t.Context(), now)
	if err := col.Ack(rep); !errors.Is(err, ErrCursorTooLarge) {
		t.Fatalf("ack err = %v, want ErrCursorTooLarge for an oversize snapshot with chunking off", err)
	}
	again := col.Tick(t.Context(), rep.Next)
	if len(again.Changes) != 1 || !again.Baseline {
		t.Errorf("the unsaved ack must not advance: %+v", again)
	}
}
