package docstore

import (
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

// CollectorKind is the claim store's document kind for collector cursors (lane 18, safety priority).
const CollectorKind = "collector"

// CollectorCursors is Cursors on a claim store, using the collector document kind.
func CollectorCursors(c *claim.Store) Cursors { return Cursors{Docs: c, Kind: CollectorKind} }

// Cursors keeps a collector's acknowledged baseline (its cursor) in the claim store's documents, so it
// follows the same authority and revision rules as every other document. Kind is the document kind to use.
// It is required because the claim store, not this package, decides which kinds exist; with a kind it does
// not know, every call fails with the store's invalid-identifier error and the collector pauses (it never
// invents a baseline). Naming a kind for this purpose is lane 18's decision.
type Cursors struct {
	Docs Docs
	Kind string
}

var _ forge.CursorStore = Cursors{}

// docName turns a free-form cursor key (repository and path) into a document name.
func docName(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])[:32]
}

type cursorDoc struct {
	Key  string `json:"key"`
	Body []byte `json:"body"`
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
	return cd.Body, true, nil
}

func (c Cursors) SaveCursor(key string, body []byte) error {
	if c.Kind == "" {
		return errors.New("docstore: the cursor document kind is not configured")
	}
	raw, err := json.Marshal(cursorDoc{Key: key, Body: body})
	if err != nil {
		return err
	}
	var rev int64
	switch cur, err := c.Docs.GetDoc(c.Kind, docName(key)); {
	case err == nil:
		rev = cur.Rev
	case !errors.Is(err, core.ErrNotHeld):
		return err
	}
	_, err = c.Docs.PutDoc(c.Kind, docName(key), rev, raw)
	return err
}
