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

var _ forge.CursorStore = Cursors{}

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
	if c.Kind == "" {
		return nil, false, errors.New("docstore: the cursor document kind is not configured")
	}
	d, err := c.Docs.GetDoc(c.Kind, docName(key))
	if errors.Is(err, core.ErrNotHeld) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var cd cursorDoc
	if err := json.Unmarshal(d.Data, &cd); err != nil {
		return nil, false, fmt.Errorf("docstore: the cursor document is not readable: %w", err)
	}
	if cd.Key != key {
		return nil, false, fmt.Errorf("docstore: the cursor document holds key %q, not %q", cd.Key, key)
	}
	if cd.Chunked == nil {
		return cd.Body, true, nil
	}
	for set := cd.Chunked; set != nil; set = set.Prev {
		body, ok, err := c.assemble(key, set)
		if err != nil {
			return nil, false, err
		}
		if ok {
			return body, true, nil
		}
	}
	return nil, false, fmt.Errorf("%w: key %q", ErrCursorIncomplete, key)
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

// stored is what is currently filed under a cursor's name.
type stored struct {
	rev int64
	set *chunkSet // the current manifest's generation, nil for a plain or absent cursor
}

func (c Cursors) current(key string) (stored, error) {
	d, err := c.Docs.GetDoc(c.Kind, docName(key))
	if errors.Is(err, core.ErrNotHeld) {
		return stored{}, nil
	}
	if err != nil {
		return stored{}, err
	}
	var cd cursorDoc
	if json.Unmarshal(d.Data, &cd) != nil { // unreadable: overwrite it, keep the revision
		return stored{rev: d.Rev}, nil
	}
	return stored{rev: d.Rev, set: cd.Chunked}, nil
}

func (c Cursors) SaveCursor(key string, body []byte) error {
	if c.Kind == "" {
		return errors.New("docstore: the cursor document kind is not configured")
	}
	single, err := json.Marshal(cursorDoc{Key: key, Body: body})
	if err != nil {
		return err
	}
	cur, err := c.current(key)
	if err != nil {
		return err
	}
	if len(single) <= c.limit() {
		if _, err := c.Docs.PutDoc(c.Kind, docName(key), cur.rev, single); err != nil {
			return err
		}
		c.drop(cur.set, nil) // the plain document is complete on its own; the old chunks are garbage
		if cur.set != nil {
			c.drop(cur.set.Prev, nil)
		}
		return nil
	}
	return c.saveChunked(key, body, len(single), cur)
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

func (c Cursors) saveChunked(key string, body []byte, encoded int, cur stored) error {
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
	if cur.set != nil {
		prev := *cur.set
		prev.Prev = nil
		set.Prev = &prev
	}
	manifest, err := json.Marshal(cursorDoc{Key: key, Chunked: set})
	if err != nil {
		return err
	}
	if len(manifest) > c.limit() { // checked before the first chunk is written
		return tooLarge("manifest", len(manifest), n)
	}
	for i, ref := range set.Chunks {
		if err := c.putSame(ref.Name, raws[i]); err != nil {
			return err
		}
	}
	// The manifest is written last: this one document is what adopts the new cursor.
	if _, err := c.Docs.PutDoc(c.Kind, docName(key), cur.rev, manifest); err != nil {
		return err
	}
	if cur.set != nil {
		c.drop(cur.set.Prev, set, cur.set) // the generation before the previous one
	}
	return nil
}

// putSame writes a content-addressed document; one already holding exactly these bytes is left alone.
func (c Cursors) putSame(name string, raw []byte) error {
	var rev int64
	switch d, err := c.Docs.GetDoc(c.Kind, name); {
	case err == nil:
		if bytes.Equal(d.Data, raw) {
			return nil
		}
		rev = d.Rev
	case !errors.Is(err, core.ErrNotHeld):
		return err
	}
	_, err := c.Docs.PutDoc(c.Kind, name, rev, raw)
	return err
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
