package docstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Docs is the slice of claim.Store that Cursors needs.
type Docs interface {
	GetDoc(kind, name string) (claim.Doc, error)
	PutDoc(kind, name string, expectRev int64, data []byte) (claim.Doc, error)
}

// docRemover is implemented by the claim store. Cleaning up old chunks needs it; without it the old
// chunks stay behind as harmless garbage.
type docRemover interface {
	DeleteDoc(kind, name string, expectRev int64) error
}

// docUpdater is claim.Store's safe read-modify-write: when a revision conflict shows the cache is behind
// the authority, it reads the document from the authority again and runs fn on what is there.
type docUpdater interface {
	UpdateDoc(kind, name string, fn func(cur []byte) ([]byte, error)) (claim.Doc, error)
}

// errNoWrite is what a modify callback returns to say that nothing needs writing.
var errNoWrite = errors.New("docstore: nothing to write")

// modifyRetries bounds the compare-and-swap rounds of the fallback below. It matches claim.UpdateDoc's bound.
const modifyRetries = 10

// modify applies fn to a document's current data (nil when absent) and stores the result. On a claim
// store this is UpdateDoc, so a cache that fell behind the authority (a crash between the authority write
// and the cache write) is repaired by the next write instead of conflicting forever. Docs without
// UpdateDoc (test doubles) get a plain read-modify-write loop.
func modify(d Docs, kind, name string, fn func(cur []byte) ([]byte, error)) error {
	if u, ok := d.(docUpdater); ok {
		_, err := u.UpdateDoc(kind, name, fn)
		if errors.Is(err, errNoWrite) {
			return nil
		}
		return err
	}
	for try := 0; try < modifyRetries; try++ {
		var cur []byte
		var rev int64
		switch doc, err := d.GetDoc(kind, name); {
		case err == nil:
			cur, rev = doc.Data, doc.Rev
		case !errors.Is(err, core.ErrNotHeld):
			return err
		}
		data, err := fn(cur)
		if errors.Is(err, errNoWrite) {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = d.PutDoc(kind, name, rev, data)
		if errors.Is(err, core.ErrRevisionConflict) {
			continue
		}
		return err
	}
	return fmt.Errorf("%w: %s/%s kept moving", core.ErrBusy, kind, name)
}

// CollectorKind is the claim store's document kind for collector cursors (lane 18, safety priority).
const CollectorKind = "collector"

// DefaultDocLimit mirrors the claim store's per-document cap (claim/docs.go maxDocBytes). That constant is
// unexported and lane 18's, so it is copied here; TestDocLimitMatchesTheClaimStore pins the copy against
// the real store, so a drift fails a test instead of a write.
const DefaultDocLimit = 256 << 10

// ErrCursorTooLarge marks a cursor that cannot be stored within the document limit (and the chunk bound).
// It is returned before anything is written; the cursor is never cut to fit.
var ErrCursorTooLarge = errors.New("docstore: the cursor does not fit the document limit")

// ErrCursorIncomplete marks a chunked cursor whose chunks are not all present and intact, with no older
// complete generation to fall back to. The collector pauses on it rather than inventing a baseline.
var ErrCursorIncomplete = errors.New("docstore: the cursor's chunks are not all present and intact")

// TooLargeError says what did not fit.
type TooLargeError struct {
	Key       string
	What      string // "document", "chunks" or "manifest"
	Encoded   int    // bytes the stored form needs
	Limit     int
	Chunks    int
	MaxChunks int
}

func (e *TooLargeError) Error() string {
	return fmt.Sprintf("docstore: cursor %q: %s needs %d bytes (%d chunks) against a limit of %d bytes (%d chunks); nothing was written",
		e.Key, e.What, e.Encoded, e.Chunks, e.Limit, e.MaxChunks)
}

func (e *TooLargeError) Unwrap() error { return ErrCursorTooLarge }

// CollectorCursors is Cursors on a claim store, using the collector document kind.
func CollectorCursors(c *claim.Store) Cursors { return Cursors{Docs: c, Kind: CollectorKind} }

// Cursors keeps a collector's acknowledged baseline (its cursor) in the claim store's documents, so it
// follows the same authority and revision rules as every other document. Kind is the document kind to use.
// It is required because the claim store, not this package, decides which kinds exist; with a kind it does
// not know, every call fails with the store's invalid-identifier error and the collector pauses (it never
// invents a baseline). Naming a kind for this purpose is lane 18's decision.
//
// A cursor that fits one document is stored as one document. A larger one is split into content-addressed
// chunk documents plus a manifest written last: the manifest is the single document whose write adopts the
// new cursor, so a reader sees either the previous cursor or the whole new one, never a mixture. The
// manifest also names the previous generation, whose chunks are kept until the generation after next, so a
// reader that finds the newest generation incomplete falls back to the previous complete one.
type Cursors struct {
	Docs Docs
	Kind string
	// MaxDocBytes is the largest stored document (bytes of its JSON). Zero means DefaultDocLimit.
	MaxDocBytes int
	// MaxChunks bounds the chunks of one cursor. Zero turns chunking off: a cursor that does not fit one
	// document is refused (ErrCursorTooLarge). It is configuration, not a constant of this package.
	MaxChunks int
}

var (
	_ forge.CursorStore      = Cursors{}
	_ forge.NotedCursorStore = Cursors{}
)

// docName turns a free-form cursor key (repository and path) into a document name.
func docName(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])[:32]
}

// chunkName names a chunk by the key it belongs to and the digest of its bytes.
func chunkName(key, digest string) string { return docName(key + "\x00chunk\x00" + digest) }

type chunkRef struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Size   int    `json:"size"`
}

// chunkSet describes one generation of a chunked cursor.
type chunkSet struct {
	Digest string     `json:"digest"` // sha256 of the whole cursor
	Size   int        `json:"size"`
	Chunks []chunkRef `json:"chunks"`
	Prev   *chunkSet  `json:"prev,omitempty"` // the generation before this one; never nested further
}

// cursorDoc is the document at docName(key): the cursor itself, or (Chunked set) the manifest of it.
type cursorDoc struct {
	Key     string    `json:"key"`
	Body    []byte    `json:"body"`
	Chunked *chunkSet `json:"chunked,omitempty"`
}

type chunkDoc struct {
	Key    string `json:"key"`
	Digest string `json:"digest"`
	Body   []byte `json:"body"`
}

func (c Cursors) limit() int {
	if c.MaxDocBytes > 0 {
		return c.MaxDocBytes
	}
	return DefaultDocLimit
}

func (c Cursors) LoadCursor(key string) ([]byte, bool, error) {
	body, found, _, err := c.LoadCursorNoted(key)
	return body, found, err
}

// FallbackNote is what LoadCursorNoted says when it had to use the previous generation.
const FallbackNote = "the newest stored cursor generation is incomplete; resumed from the previous one, so changes since it may be reported again"

// LoadCursorNoted is LoadCursor that also says when the cursor it returned is not the newest one. The
// previous generation is a safe cursor (it can only make changes come back), but silently going backwards
// would hide that the newest write was lost.
func (c Cursors) LoadCursorNoted(key string) (body []byte, found bool, note string, err error) {
	if c.Kind == "" {
		return nil, false, "", errors.New("docstore: the cursor document kind is not configured")
	}
	d, err := c.Docs.GetDoc(c.Kind, docName(key))
	if errors.Is(err, core.ErrNotHeld) {
		return nil, false, "", nil
	}
	if err != nil {
		return nil, false, "", err
	}
	var cd cursorDoc
	if err := json.Unmarshal(d.Data, &cd); err != nil {
		return nil, false, "", fmt.Errorf("docstore: the cursor document is not readable: %w", err)
	}
	if cd.Key != key {
		return nil, false, "", fmt.Errorf("docstore: the cursor document holds key %q, not %q", cd.Key, key)
	}
	if cd.Chunked == nil {
		return cd.Body, true, "", nil
	}
	for set := cd.Chunked; set != nil; set = set.Prev {
		body, ok, err := c.assemble(key, set)
		if err != nil {
			return nil, false, "", err
		}
		if ok {
			if set != cd.Chunked {
				note = FallbackNote
			}
			return body, true, note, nil
		}
	}
	return nil, false, "", fmt.Errorf("%w: key %q", ErrCursorIncomplete, key)
}

// assemble joins one generation's chunks and checks every digest. ok=false means the generation is not
// complete or not intact; err is only for a store that could not be read.
func (c Cursors) assemble(key string, set *chunkSet) (body []byte, ok bool, err error) {
	var out bytes.Buffer
	for _, ref := range set.Chunks {
		d, err := c.Docs.GetDoc(c.Kind, ref.Name)
		if errors.Is(err, core.ErrNotHeld) || errors.Is(err, core.ErrIDInvalid) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		var cd chunkDoc
		if json.Unmarshal(d.Data, &cd) != nil || cd.Key != key || len(cd.Body) != ref.Size || hexSum(cd.Body) != ref.Digest {
			return nil, false, nil
		}
		out.Write(cd.Body)
	}
	if out.Len() != set.Size || hexSum(out.Bytes()) != set.Digest {
		return nil, false, nil
	}
	return out.Bytes(), true, nil
}

func hexSum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// setOf reads the chunk set a stored cursor document describes (nil for a plain, absent or unreadable one).
func setOf(cur []byte) *chunkSet {
	if cur == nil {
		return nil
	}
	var cd cursorDoc
	if json.Unmarshal(cur, &cd) != nil {
		return nil
	}
	return cd.Chunked
}

func (c Cursors) SaveCursor(key string, body []byte) error {
	if c.Kind == "" {
		return errors.New("docstore: the cursor document kind is not configured")
	}
	single, err := json.Marshal(cursorDoc{Key: key, Body: body})
	if err != nil {
		return err
	}
	if len(single) > c.limit() {
		return c.saveChunked(key, body, len(single))
	}
	var old *chunkSet
	if err := modify(c.Docs, c.Kind, docName(key), func(cur []byte) ([]byte, error) {
		old = setOf(cur)
		return single, nil
	}); err != nil {
		return err
	}
	// The plain document is complete on its own; the old chunks are garbage.
	c.drop(old, nil)
	if old != nil {
		c.drop(old.Prev, nil)
	}
	return nil
}

// chunkSize is the most raw bytes one chunk document can carry within the limit.
func (c Cursors) chunkSize(key string) (int, error) {
	overhead, err := json.Marshal(chunkDoc{Key: key, Digest: hexSum(nil), Body: []byte{}})
	if err != nil {
		return 0, err
	}
	avail := c.limit() - len(overhead)
	if avail < 4 {
		return 0, &TooLargeError{Key: key, What: "chunk", Encoded: len(overhead), Limit: c.limit()}
	}
	return avail / 4 * 3, nil // base64: 4 output bytes per 3 input bytes
}

func (c Cursors) saveChunked(key string, body []byte, encoded int) error {
	tooLarge := func(what string, n, chunks int) error {
		return &TooLargeError{Key: key, What: what, Encoded: n, Limit: c.limit(), Chunks: chunks, MaxChunks: c.MaxChunks}
	}
	if c.MaxChunks <= 0 {
		return tooLarge("document", encoded, 1)
	}
	per, err := c.chunkSize(key)
	if err != nil {
		return err
	}
	n := (len(body) + per - 1) / per
	if n > c.MaxChunks {
		return tooLarge("chunks", encoded, n)
	}
	set := &chunkSet{Digest: hexSum(body), Size: len(body)}
	raws := make([][]byte, 0, n)
	for i := range n {
		piece := body[i*per : min((i+1)*per, len(body))]
		ref := chunkRef{Digest: hexSum(piece), Size: len(piece)}
		ref.Name = chunkName(key, ref.Digest)
		raw, err := json.Marshal(chunkDoc{Key: key, Digest: ref.Digest, Body: piece})
		if err != nil {
			return err
		}
		if len(raw) > c.limit() {
			return tooLarge("chunks", len(raw), n)
		}
		set.Chunks = append(set.Chunks, ref)
		raws = append(raws, raw)
	}
	manifestFor := func(cur []byte) (*chunkSet, []byte, error) {
		s := *set
		old := setOf(cur)
		if old != nil {
			prev := *old
			prev.Prev = nil
			s.Prev = &prev
		}
		m, err := json.Marshal(cursorDoc{Key: key, Chunked: &s})
		if err != nil {
			return nil, nil, err
		}
		if len(m) > c.limit() {
			return nil, nil, tooLarge("manifest", len(m), n)
		}
		return old, m, nil
	}
	// Refuse before the first write when the manifest, as the cache sees it, would not fit.
	var seen []byte
	if d, err := c.Docs.GetDoc(c.Kind, docName(key)); err == nil {
		seen = d.Data
	} else if !errors.Is(err, core.ErrNotHeld) {
		return err
	}
	if _, _, err := manifestFor(seen); err != nil {
		return err
	}
	for i, ref := range set.Chunks {
		if err := c.putSame(ref.Name, raws[i]); err != nil {
			return err
		}
	}
	// The manifest is written last: this one document is what adopts the new cursor.
	var old *chunkSet
	if err := modify(c.Docs, c.Kind, docName(key), func(cur []byte) ([]byte, error) {
		var m []byte
		var err error
		old, m, err = manifestFor(cur)
		return m, err
	}); err != nil {
		return err
	}
	if old != nil {
		c.drop(old.Prev, set, old) // the generation before the previous one
	}
	return nil
}

// putSame writes a content-addressed document; one already holding exactly these bytes is left alone.
func (c Cursors) putSame(name string, raw []byte) error {
	return modify(c.Docs, c.Kind, name, func(cur []byte) ([]byte, error) {
		if bytes.Equal(cur, raw) {
			return nil, errNoWrite
		}
		return raw, nil
	})
}

// drop removes the chunks of old that none of keep use. It is best effort: a chunk left behind is garbage
// that no manifest names, never a wrong cursor.
func (c Cursors) drop(old *chunkSet, keep ...*chunkSet) {
	rm, ok := c.Docs.(docRemover)
	if !ok || old == nil {
		return
	}
	kept := map[string]bool{}
	for _, k := range keep {
		if k != nil {
			for _, ref := range k.Chunks {
				kept[ref.Name] = true
			}
		}
	}
	for _, ref := range old.Chunks {
		if kept[ref.Name] {
			continue
		}
		if d, err := c.Docs.GetDoc(c.Kind, ref.Name); err == nil {
			_ = rm.DeleteDoc(c.Kind, ref.Name, d.Rev)
		}
	}
}
