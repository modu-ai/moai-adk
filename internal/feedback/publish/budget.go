package publish

// budget.go — the rolling daily model-call cap (design.md section 8/10,
// DEC-3, AC-019): every call counts against DefaultBugreportModelCallsPerDay,
// counted at ATTEMPT time (a failing endpoint hammers no further than the
// cap), and the budget is consulted before the seam is entered. The window
// is in-process: one flush is the unit of spend in practice, and the
// per-item bound already caps a single queue at one call per item.

import (
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ModelCallBudget is the rolling daily cap on model calls.
type ModelCallBudget struct {
	mu   sync.Mutex
	at   []time.Time
	now  func() time.Time
}

// NewModelCallBudget returns the budget over the configured daily cap.
func NewModelCallBudget() *ModelCallBudget {
	return &ModelCallBudget{now: time.Now}
}

// Allow reports whether one more call may be made inside the rolling
// 24-hour window.
func (b *ModelCallBudget) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count() < config.DefaultBugreportModelCallsPerDay
}

// Record counts one call attempt.
func (b *ModelCallBudget) Record() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.at = append(b.at, b.now())
}

// count is the calls inside the window; the caller holds the lock.
func (b *ModelCallBudget) count() int {
	now := b.now()
	kept := b.at[:0]
	n := 0
	for _, at := range b.at {
		if now.Sub(at) < 24*time.Hour {
			kept = append(kept, at)
			n++
		}
	}
	b.at = kept
	return n
}
