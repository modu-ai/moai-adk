package core

// Priority separates writes that keep the factory safe from writes that only
// report on it. When the shared call budget runs short, reports wait first.
type Priority int

const (
	// PrioritySafety is claim, handoff, ACK and agent-number state: never held
	// back by a budget, and a stale revision is refused by compare-and-swap.
	PrioritySafety Priority = iota
	// PriorityReport is the Issue ledger comment and other summaries.
	PriorityReport
)

func (p Priority) String() string {
	switch p {
	case PrioritySafety:
		return "safety"
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
