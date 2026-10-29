package docstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

var errCrash = errors.New("simulated crash")

// faultDocs wraps a real claim store and breaks one PutDoc at a chosen count, so a crash can be placed
// between two document writes. failAt refuses the put before it writes; crashAt writes it and then
// reports the crash (the caller never learns the write landed). Counts are 1-based and restart at arm.
// A restart is a new claim.Store (and a new faultDocs) over the same directory.
type faultDocs struct {
	inner                 *claim.Store
	puts, failAt, crashAt int
}

func (f *faultDocs) arm(failAt, crashAt int) { f.puts, f.failAt, f.crashAt = 0, failAt, crashAt }

func (f *faultDocs) GetDoc(kind, name string) (claim.Doc, error) { return f.inner.GetDoc(kind, name) }

func (f *faultDocs) DeleteDoc(kind, name string, rev int64) error {
	return f.inner.DeleteDoc(kind, name, rev)
}

func (f *faultDocs) PutDoc(kind, name string, rev int64, data []byte) (claim.Doc, error) {
	f.puts++
	if f.puts == f.failAt {
		return claim.Doc{}, errCrash
	}
	d, err := f.inner.PutDoc(kind, name, rev, data)
	if err == nil && f.puts == f.crashAt {
		return claim.Doc{}, errCrash
	}
	return d, err
}

func openStore(t *testing.T, dir string) *claim.Store {
	t.Helper()
	st, err := claim.OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func openFault(t *testing.T, dir string) *faultDocs {
	t.Helper()
	return &faultDocs{inner: openStore(t, dir)}
}

func chunkedCursors(d Docs) Cursors {
	return Cursors{Docs: d, Kind: CollectorKind, MaxChunks: 8}
}

// noise is n deterministic printable bytes, so every chunk of a body has its own digest.
func noise(seed int64, n int) string {
	r := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a' + byte(r.Intn(26))
	}
	return string(b)
}

// bigBody is a JSON array of one issue carrying n bytes of padding.
func bigBody(seed int64, n int) []byte {
	return []byte(`[{"number":1,"pad":"` + noise(seed, n) + `"}]`)
}

func jsonString(n int) []byte { return []byte(`"` + strings.Repeat("a", n-2) + `"`) }

// The copy of the document cap is pinned to the real store: exactly the cap is accepted, one byte more is not.
func TestDocLimitMatchesTheClaimStore(t *testing.T) {
	st := openStore(t, t.TempDir())
	if _, err := st.PutDoc(CollectorKind, "limit-ok", 0, jsonString(DefaultDocLimit)); err != nil {
		t.Fatalf("a document of exactly DefaultDocLimit bytes was refused: %v", err)
	}
	if _, err := st.PutDoc(CollectorKind, "limit-over", 0, jsonString(DefaultDocLimit+1)); err == nil {
		t.Fatal("a document one byte over DefaultDocLimit was accepted: the copied cap drifted from claim/docs.go")
	}
}

// The size that counts is the stored form (base64 inflates the body by a third), measured before the write.
func TestAnOversizeCursorIsRefusedBeforeAnyWriteAndNeverCut(t *testing.T) {
	dir := t.TempDir()
	fd := openFault(t, dir)
	c := Cursors{Docs: fd, Kind: CollectorKind} // chunking off
	first := 0
	for n := DefaultDocLimit/4*3 - 200; ; n++ { // the smallest body whose stored form passes the limit
		raw, _ := json.Marshal(cursorDoc{Key: "k", Body: make([]byte, n)})
		if len(raw) > DefaultDocLimit {
			first = n
			break
		}
	}
	if err := c.SaveCursor("k", bytes.Repeat([]byte("x"), first-1)); err != nil {
		t.Fatalf("the largest body that fits was refused: %v", err)
	}
	fd.arm(0, 0)
	err := c.SaveCursor("k", bytes.Repeat([]byte("y"), first))
	var tl *TooLargeError
	if !errors.Is(err, ErrCursorTooLarge) || !errors.As(err, &tl) || tl.Encoded <= tl.Limit {
		t.Fatalf("err = %v, want a TooLargeError whose stored form exceeds the limit", err)
	}
	if fd.puts != 0 {
		t.Errorf("%d document writes were attempted for a cursor that cannot fit", fd.puts)
	}
	got, found, err := c.LoadCursor("k")
	if err != nil || !found || len(got) != first-1 || got[0] != 'x' {
		t.Errorf("the previous cursor must be untouched: len=%d found=%v err=%v", len(got), found, err)
	}
}

func TestMoreChunksThanAllowedIsRefusedBeforeAnyWrite(t *testing.T) {
	fd := openFault(t, t.TempDir())
	c := Cursors{Docs: fd, Kind: CollectorKind, MaxChunks: 2}
	err := c.SaveCursor("k", bytes.Repeat([]byte("z"), 600<<10)) // needs 4 chunks
	var tl *TooLargeError
	if !errors.As(err, &tl) || tl.What != "chunks" || tl.Chunks <= tl.MaxChunks {
		t.Fatalf("err = %v, want a chunk-count TooLargeError", err)
	}
	if fd.puts != 0 {
		t.Errorf("%d writes before the refusal", fd.puts)
	}
}

func TestAChunkedCursorRoundTripsAcrossARestart(t *testing.T) {
	dir := t.TempDir()
	body := bigBody(1, 600<<10)
	if err := chunkedCursors(openStore(t, dir)).SaveCursor("o/r/pulls", body); err != nil {
		t.Fatal(err)
	}
	st := openStore(t, dir)
	got, found, err := chunkedCursors(st).LoadCursor("o/r/pulls")
	if err != nil || !found || !bytes.Equal(got, body) {
		t.Fatalf("after a restart: len=%d found=%v err=%v", len(got), found, err)
	}
	docs, _ := st.ListDocs(CollectorKind)
	if len(docs) < 3 {
		t.Errorf("a 600KB cursor should be a manifest plus several chunks, got %d documents", len(docs))
	}
	for _, d := range docs {
		if len(d.Data) > DefaultDocLimit {
			t.Errorf("document %s is %d bytes, over the limit", d.Name, len(d.Data))
		}
	}
}

// totalPuts counts the document writes one save of body makes over a store that already holds prior.
func totalPuts(t *testing.T, prior, body []byte) int {
	t.Helper()
	fd := openFault(t, t.TempDir())
	c := chunkedCursors(fd)
	if prior != nil {
		if err := c.SaveCursor("k", prior); err != nil {
			t.Fatal(err)
		}
	}
	fd.arm(0, 0)
	if err := c.SaveCursor("k", body); err != nil {
		t.Fatal(err)
	}
	return fd.puts
}

// A crash at every point of a chunked save — before or after each document write — leaves either the whole
// previous cursor or the whole new one; a restart never sees a mixture and never an invented baseline.
func TestACrashAtEveryPointOfAChunkedSaveKeepsAWholeCursor(t *testing.T) {
	v1, v2 := bigBody(1, 600<<10), bigBody(2, 600<<10)
	total := totalPuts(t, v1, v2) // chunks, then the manifest last
	if total < 3 {
		t.Fatalf("expected several writes, got %d", total)
	}
	for _, mode := range []string{"refused-before-write", "crash-after-write"} {
		for k := 1; k <= total; k++ {
			dir := t.TempDir()
			fd := openFault(t, dir)
			c := chunkedCursors(fd)
			if err := c.SaveCursor("k", v1); err != nil {
				t.Fatal(err)
			}
			if mode == "refused-before-write" {
				fd.arm(k, 0)
			} else {
				fd.arm(0, k)
			}
			if err := c.SaveCursor("k", v2); !errors.Is(err, errCrash) {
				t.Fatalf("%s at write %d/%d: err = %v, want the crash", mode, k, total, err)
			}
			got, found, err := chunkedCursors(openStore(t, dir)).LoadCursor("k") // restart
			want := v1
			if mode == "crash-after-write" && k == total {
				want = v2 // the manifest landed: the new cursor is adopted whole
			}
			if err != nil || !found || !bytes.Equal(got, want) {
				t.Errorf("%s at write %d/%d: found=%v err=%v matches-expected=%v", mode, k, total, found, err, bytes.Equal(got, want))
			}
			// The interrupted save is simply redone.
			if err := chunkedCursors(openStore(t, dir)).SaveCursor("k", v2); err != nil {
				t.Errorf("%s at write %d: redo failed: %v", mode, k, err)
			}
			if got, _, _ := chunkedCursors(openStore(t, dir)).LoadCursor("k"); !bytes.Equal(got, v2) {
				t.Errorf("%s at write %d: after the redo the cursor is not the new one", mode, k)
			}
		}
	}
}

// With no previous cursor, a crash before the manifest leaves "no cursor": found=false without an error,
// which the collector reads as "nothing was ever acknowledged" — the true state.
func TestACrashBeforeTheFirstManifestLeavesNoCursor(t *testing.T) {
	v := bigBody(3, 600<<10)
	total := totalPuts(t, nil, v)
	for k := 1; k < total; k++ {
		dir := t.TempDir()
		fd := openFault(t, dir)
		fd.arm(0, k)
		if err := chunkedCursors(fd).SaveCursor("k", v); !errors.Is(err, errCrash) {
			t.Fatalf("write %d: %v", k, err)
		}
		if got, found, err := chunkedCursors(openStore(t, dir)).LoadCursor("k"); found || err != nil {
			t.Errorf("write %d/%d: found=%v err=%v (len %d), want no cursor", k, total, found, err, len(got))
		}
	}
}

// chunkNamesOf reads the manifest filed under key and returns the generation it describes.
func manifestOf(t *testing.T, st *claim.Store, key string) *chunkSet {
	t.Helper()
	d, err := st.GetDoc(CollectorKind, docName(key))
	if err != nil {
		t.Fatal(err)
	}
	var cd cursorDoc
	if err := json.Unmarshal(d.Data, &cd); err != nil || cd.Chunked == nil {
		t.Fatalf("not a manifest: %v", err)
	}
	return cd.Chunked
}

func removeDoc(t *testing.T, st *claim.Store, name string) {
	t.Helper()
	d, err := st.GetDoc(CollectorKind, name)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteDoc(CollectorKind, name, d.Rev); err != nil {
		t.Fatal(err)
	}
}

// A manifest whose chunks are not all there is not adopted: the previous generation is used; with none
// left the load fails and the collector pauses.
func TestAMissingChunkFallsBackToThePreviousGenerationThenFails(t *testing.T) {
	dir := t.TempDir()
	st := openStore(t, dir)
	c := chunkedCursors(st)
	v1, v2 := bigBody(1, 600<<10), bigBody(2, 600<<10)
	for _, v := range [][]byte{v1, v2} {
		if err := c.SaveCursor("k", v); err != nil {
			t.Fatal(err)
		}
	}
	set := manifestOf(t, st, "k")
	removeDoc(t, st, set.Chunks[0].Name) // a chunk only the newest generation uses
	got, found, err := c.LoadCursor("k")
	if err != nil || !found || !bytes.Equal(got, v1) {
		t.Fatalf("newest incomplete: found=%v err=%v isV1=%v", found, err, bytes.Equal(got, v1))
	}
	removeDoc(t, st, set.Prev.Chunks[0].Name)
	if _, found, err := c.LoadCursor("k"); found || !errors.Is(err, ErrCursorIncomplete) {
		t.Fatalf("no complete generation left: found=%v err=%v, want ErrCursorIncomplete", found, err)
	}
}

func TestACorruptChunkIsNotAdopted(t *testing.T) {
	dir := t.TempDir()
	st := openStore(t, dir)
	c := chunkedCursors(st)
	v1, v2 := bigBody(1, 600<<10), bigBody(2, 600<<10)
	_ = c.SaveCursor("k", v1)
	_ = c.SaveCursor("k", v2)
	set := manifestOf(t, st, "k")
	ref := set.Chunks[1]
	d, _ := st.GetDoc(CollectorKind, ref.Name)
	forged, _ := json.Marshal(chunkDoc{Key: "k", Digest: ref.Digest, Body: bytes.Repeat([]byte("q"), ref.Size)})
	if _, err := st.PutDoc(CollectorKind, ref.Name, d.Rev, forged); err != nil {
		t.Fatal(err)
	}
	got, found, err := c.LoadCursor("k")
	if err != nil || !found || !bytes.Equal(got, v1) {
		t.Errorf("a chunk with the right size and the wrong bytes must not be adopted: found=%v err=%v isV1=%v", found, err, bytes.Equal(got, v1))
	}
}

// Two generations are kept; the one before is removed, and a cursor that shrinks to one document takes
// its chunks with it.
func TestCleanupKeepsTwoGenerationsAndASmallCursorDropsTheChunks(t *testing.T) {
	st := openStore(t, t.TempDir())
	c := chunkedCursors(st)
	var firstChunks []string
	for i, seed := range []int64{1, 2, 3} {
		if err := c.SaveCursor("k", bigBody(seed, 600<<10)); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			for _, r := range manifestOf(t, st, "k").Chunks {
				firstChunks = append(firstChunks, r.Name)
			}
		}
	}
	for _, name := range firstChunks {
		if _, err := st.GetDoc(CollectorKind, name); !errors.Is(err, core.ErrNotHeld) {
			t.Errorf("a chunk of the generation before the previous one is still stored: %s (%v)", name, err)
		}
	}
	set := manifestOf(t, st, "k")
	docs, _ := st.ListDocs(CollectorKind)
	if want := 1 + len(set.Chunks) + len(set.Prev.Chunks); len(docs) != want {
		t.Errorf("documents = %d, want manifest plus two generations = %d", len(docs), want)
	}
	if err := c.SaveCursor("k", []byte(`[]`)); err != nil {
		t.Fatal(err)
	}
	if docs, _ := st.ListDocs(CollectorKind); len(docs) != 1 {
		t.Errorf("a cursor back to one document should leave exactly one, got %d", len(docs))
	}
	if got, found, err := c.LoadCursor("k"); err != nil || !found || string(got) != `[]` {
		t.Errorf("small cursor after chunked: %q found=%v err=%v", got, found, err)
	}
}

// End to end: a snapshot that does not fit one document is acknowledged, survives a restart, and the
// restarted collector reports only the real change and no baseline.
func TestABigSnapshotIsAcknowledgedAndResumedAfterARestart(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	body1 := string(bigBody(7, 300<<10))
	body2 := strings.TrimSuffix(body1, "]") + `,{"number":2,"pad":"x"}]`
	mk := func(body string) *forge.Collector {
		return &forge.Collector{
			Client: &forge.Client{T: &oneShot{body: body}, B: &forge.Budget{}, Now: func() time.Time { return now },
				Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
			Path: "/p", Interval: forge.Adaptive{Min: time.Second, Max: time.Minute, Safety: time.Hour},
			Differ: forge.DiffByKey("number"), Cursors: chunkedCursors(openStore(t, dir)), CursorKey: "o/r/big",
		}
	}
	first := mk(body1)
	rep := first.Tick(t.Context(), now)
	if err := first.Ack(rep); err != nil {
		t.Fatalf("a big snapshot could not be acknowledged: %v", err)
	}
	r2 := mk(body2).Tick(t.Context(), now)
	if r2.Baseline || len(r2.Changes) != 1 || r2.Changes[0].Key != "2" || r2.Changes[0].Kind != forge.ChangeAdded {
		t.Errorf("after the restart: %+v, want only issue 2 added and no baseline", r2)
	}
}
