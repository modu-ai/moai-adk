// context_model_source.go — connects the claim store's context documents to
// the monitor's ledger view (Issue #1948). ContextModelFrom turns a document
// reader into the Config.ContextModel source: it loads the Issue's stored
// ledger through ctxmodel.Load, so the document name and the "never written
// reads as empty" rule stay in one place, and encodes it with the ledger's own
// Marshal. The web only reads: the reader handed to ctxmodel.Load refuses every
// write. Nothing is cached; every refresh reads again.
package web

import (
	"context"
	"errors"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

// ContextDocs is the read side of claim.Store the ledger view needs.
type ContextDocs interface {
	GetDoc(kind, name string) (claim.Doc, error)
}

// readOnlyDocs completes ctxmodel.DocStore with a write that always fails.
type readOnlyDocs struct{ ContextDocs }

func (readOnlyDocs) UpdateDoc(string, string, func([]byte) ([]byte, error)) (claim.Doc, error) {
	return claim.Doc{}, errors.New("context model source is read-only")
}

// ContextModelFrom builds a ContextModel source. A nil reader yields a source
// that fails, which the panel renders as "model unreadable".
func ContextModelFrom(s ContextDocs) func(ctx context.Context, issue int) ([]byte, error) {
	return func(_ context.Context, issue int) ([]byte, error) {
		if s == nil {
			return nil, errors.New("context model source is not configured")
		}
		l, err := ctxmodel.Load(readOnlyDocs{s}, issue)
		if err != nil {
			return nil, err
		}
		return l.Marshal()
	}
}
