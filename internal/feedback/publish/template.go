package publish

// template.go — the deterministic create body (DEC-6's contract): the
// queued render IS the template text. The queued body was produced by the
// one render function (outbox.RenderReport) from closed fixed fields only,
// and the preview prints exactly those bytes — the create path hands the
// same bytes to gh, so a user can preview precisely what will be filed.

import (
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// CreateBody returns the body bytes the create path hands to gh. M6's
// model step augments this with a validated summary section when (and only
// when) the model succeeds; whenever the model is unavailable,
// unauthenticated, or fails — and in M5, always — this deterministic text
// is the body (DEC-6).
func CreateBody(item feedback.QueueItem) string {
	return item.Body
}
