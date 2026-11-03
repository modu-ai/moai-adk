package review

import (
	"errors"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// HandoffCurrent reports whether a writer or coordinator handoff names the
// latest HEAD. An earlier writer's receipt is kept but never reused once the
// HEAD changes (S46 HANDOFF08); a missing handoff does not count.
func HandoffCurrent(e *core.Evidence, latest core.HeadSHA) bool {
	return e != nil && e.Head == latest
}

var (
	// ErrBudgetExhausted: every configured audit round has been started.
	ErrBudgetExhausted = errors.New("review: audit round budget exhausted")
	// ErrBudgetUnset: no explicit round limit is configured; none is invented.
	ErrBudgetUnset = errors.New("review: audit round limit not configured")
	errRoundID     = errors.New("review: audit round id is empty")
)

// AuditBudget counts audit rounds for one Issue. The limit is the configured
// workflow.agentic_factory.audit.max_rounds value; zero means unset. Rounds
// are counted when started, a repeated id (lost response, same request)
// counts once, and a HEAD change never resets the count. Persisting Rounds is
// the claim store's job; this type holds only the decision.
type AuditBudget struct {
	Max    int
	Rounds []string
}

// Start records a round, or refuses it once the budget is spent.
func (b *AuditBudget) Start(roundID string) error {
	if b.Max <= 0 {
		return ErrBudgetUnset
	}
	if roundID == "" {
		return errRoundID
	}
	for _, id := range b.Rounds {
		if id == roundID {
			return nil
		}
	}
	if len(b.Rounds) >= b.Max {
		return ErrBudgetExhausted
	}
	b.Rounds = append(b.Rounds, roundID)
	return nil
}

// Used is the number of rounds started so far.
func (b *AuditBudget) Used() int { return len(b.Rounds) }

// BudgetBlock maps a budget error to the block reason it records.
func BudgetBlock(err error) (core.BlockReason, bool) {
	switch {
	case errors.Is(err, ErrBudgetExhausted):
		return core.ReasonBudgetExhausted, true
	case errors.Is(err, ErrBudgetUnset):
		return core.ReasonNotConfigured, true
	}
	return "", false
}
