package memo

// Priority levels for session memo content.
// Lower values are higher priority and are loaded first.
type Priority int

const (
	P1Required Priority = iota + 1 // Active SPEC-ID, workflow phase, execution mode
	P2High                          // TaskList summary, active agents
	P3Medium                        // Last 3 agent execution result summaries
	P4Low                           // User decision history, errors/warnings
)

// Section represents a single section of the session memo.
type Section struct {
	Priority Priority
	Title    string
	Content  string
	Budget   int // max token budget for this section
}

// priorityLabel returns the markdown header prefix for a given priority level.
func priorityLabel(p Priority) string {
	switch p {
	case P1Required:
		return "P1"
	case P2High:
		return "P2"
	case P3Medium:
		return "P3"
	case P4Low:
		return "P4"
	default:
		return "P?"
	}
}
