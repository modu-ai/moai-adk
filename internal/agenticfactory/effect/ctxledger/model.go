package ctxledger

import (
	"context"
	"errors"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

// StepBlocked means the body must not be sent: lane 18's ctxmodel refused to render it (it would expose
// unmasked content or looks like a raw log), or the text failed this package's own secret and
// personal-data checks. It is reported by value, like StepFailed, and nothing was sent.
const StepBlocked StepStatus = "blocked"

// PublishModel posts the Issue's context-ledger model as the Issue's one context comment. The body is
// exactly what ctxmodel.Load followed by (*Ledger).Render returns; this package adds no content of its
// own. Render refuses rather than masks when a field would change under redaction, so a refusal is a
// block, never an edit: the body is not sent and the step reports StepBlocked.
//
// The model decides what is linkable: only REMOTE_VERIFIED items come out as AVAILABLE links, every
// other item shows its state without a link. Because the comment is the single ledger comment of the
// Issue, an unchanged model is a no-op (same text, same digest), a changed one updates the comment in
// place, and a hand edit of the comment is restored from the model on the next call.
func (l *Ledger) PublishModel(ctx context.Context, store ctxmodel.DocStore, issue int) StepResult {
	led, err := ctxmodel.Load(store, issue)
	if err != nil {
		return StepResult{Status: StepFailed, Reason: "the ledger model could not be read", Err: err}
	}
	if len(led.Items) == 0 {
		return StepResult{Status: StepSkipped, Reason: "the ledger model has no items yet"}
	}
	text, err := led.Render()
	if err != nil {
		if errors.Is(err, ctxmodel.ErrUnmasked) {
			return StepResult{Status: StepBlocked, Reason: "the model would expose unmasked content; nothing was sent", Err: err}
		}
		return StepResult{Status: StepFailed, Reason: "the ledger model could not be rendered", Err: err}
	}
	if _, err := l.Submit(ctx, Body{Issue: issue, Text: text, MaskPassed: true}); err != nil {
		if errors.Is(err, ErrNotMasked) {
			return StepResult{Status: StepBlocked, Reason: "the rendered text failed the secret or personal-data check; nothing was sent", Err: err}
		}
		return StepResult{Status: StepFailed, Reason: "the comment could not be recorded", Err: err}
	}
	if err := l.Flush(ctx); err != nil {
		return StepResult{Status: StepFailed, Reason: "the comment did not land", Err: err}
	}
	return StepResult{Status: StepPublished}
}
