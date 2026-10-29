package core

// Priority orders calls that share one budget, most important first. When the
// budget runs short the lowest priority waits first. The numeric order is the
// contract: a scheduler may compare values and never needs to know the names.
type Priority int

const (
	// PrioritySafety is state that keeps the factory safe: claim, handoff,
	// intake receipts and agent-number state. It is announced to the budget
	// but never held back by it, and a stale revision is refused by
	// compare-and-swap instead.
	PrioritySafety Priority = iota
	// PriorityNotice is a message to a human that something needs them: a
	// security failure, an approval request, or a real block. Hiding one of
	// these hides the block itself, so a Notice MAY spend the budget's reserve
	// (the calls kept unspent for safety writes), which a Report may not. An
	// implementer may still refuse a Notice when the budget is truly empty,
	// and the caller then keeps the notice queued; it is never dropped.
	PriorityNotice
	// PriorityReport is the Issue ledger comment and other progress summaries.
	// It never touches the reserve and is the first thing deferred.
	PriorityReport
)

func (p Priority) String() string {
	switch p {
	case PrioritySafety:
		return "safety"
	case PriorityNotice:
		return "notice"
	case PriorityReport:
		return "report"
	}
	return "unknown"
}

// Budget is the shared call budget. The amounts and the backoff belong to its
// implementer (the monitor lane); this package only names the priorities.
// Reserve returns false when a call of priority p should wait. A safety
// reservation is informational: callers proceed whatever it returns.
type Budget interface {
	Reserve(p Priority) bool
}
