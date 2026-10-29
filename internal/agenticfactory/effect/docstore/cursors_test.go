package docstore

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
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

// Today's claim store lists the kinds intent, audit, context and checkpoint. A collector cursor needs a kind of
// its own, and adding one is lane 18's call: until then the real store refuses the write, which the collector
// treats as "cannot persist" (Ack returns the error, the same changes come back). This test pins that
// dependency so it cannot go unnoticed.
func TestTheRealClaimStoreHasNoCollectorKindYet(t *testing.T) {
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	c := Cursors{Docs: st, Kind: "collector"}
	if err := c.SaveCursor("o/r/pulls", []byte(`[]`)); !errors.Is(err, core.ErrIDInvalid) {
		t.Fatalf("save err = %v; if the store now lists this kind, wire it and replace this test with a round trip", err)
	}
}
