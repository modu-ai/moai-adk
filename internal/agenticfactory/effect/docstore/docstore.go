// Package docstore keeps effect intents in lane 18's claim store, as documents of kind "intent" named by
// the intent key (a hash, which fits the document name rule). The claim store is the durable home — and,
// when an authority is attached, the authority comes first — so an intent survives a restart and a
// resent-by-mistake send is prevented by reading it back. This package is only the IntentStore shape on
// top of the document calls; it adds no ledger of its own.
package docstore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

const kind = "intent"

// Store implements effect.IntentStore over a claim.Store.
type Store struct{ C *claim.Store }

var _ effect.IntentStore = Store{}

// Get returns the intent stored under key, or found=false when there is none.
func (s Store) Get(key string) (effect.Intent, bool, error) {
	d, err := s.C.GetDoc(kind, key)
	if errors.Is(err, core.ErrNotHeld) {
		return effect.Intent{}, false, nil
	}
	if err != nil {
		return effect.Intent{}, false, err
	}
	var in effect.Intent
	if err := json.Unmarshal(d.Data, &in); err != nil {
		return effect.Intent{}, false, fmt.Errorf("docstore: intent %s is not readable: %w", key, err)
	}
	if in.Key != key {
		return effect.Intent{}, false, fmt.Errorf("docstore: document %s holds intent %q", key, in.Key)
	}
	in.Rev = d.Rev
	return in, true, nil
}

// Put writes the intent with an optimistic revision check. If the stored revision moved between the read
// and the write — another writer got there first — the conflict is returned, never overwritten silently.
func (s Store) Put(in effect.Intent) error {
	raw, err := json.Marshal(in)
	if err != nil {
		return err
	}
	var rev int64
	cur, err := s.C.GetDoc(kind, in.Key)
	switch {
	case err == nil:
		rev = cur.Rev
	case !errors.Is(err, core.ErrNotHeld):
		return err
	}
	_, err = s.C.PutDoc(kind, in.Key, rev, raw)
	return err
}

// List returns every stored intent, ordered by key.
func (s Store) List() ([]effect.Intent, error) {
	docs, err := s.C.ListDocs(kind)
	if err != nil {
		return nil, err
	}
	out := make([]effect.Intent, 0, len(docs))
	for _, d := range docs {
		var in effect.Intent
		if err := json.Unmarshal(d.Data, &in); err != nil {
			return nil, fmt.Errorf("docstore: intent %s is not readable: %w", d.Name, err)
		}
		in.Rev = d.Rev
		out = append(out, in)
	}
	return out, nil
}

// PutIf writes the intent only if the stored revision is still in.Rev (0: not stored yet), with no re-read
// in between: the revision the caller read is the one the claim store compares.
func (s Store) PutIf(in effect.Intent) (effect.Intent, error) {
	raw, err := json.Marshal(in)
	if err != nil {
		return effect.Intent{}, err
	}
	d, err := s.C.PutDoc(kind, in.Key, in.Rev, raw)
	if err != nil {
		return effect.Intent{}, err
	}
	in.Rev = d.Rev
	return in, nil
}
