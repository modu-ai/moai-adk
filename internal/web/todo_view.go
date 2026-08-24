// todo_view.go — SPEC-WEB-TODO-QUEUE-001 M3: the read-only view model behind
// the /todo route.
//
// The view model does NOT touch the backlog store. It calls readTodoQueue
// (todo_queue_read.go), the console's single read seam onto the queue, so a
// change of storage lands in one function rather than here.
package web

// TodoVM is the backlog queue as the section renders it.
type TodoVM struct {
	// Root is the directory the queue resolved to — the primary checkout, the
	// home-based fallback, or (read-through, REQ-WTQ-005) the project-local
	// root. Rendered as provenance: a console served from a worktree shows the
	// primary's cards, and the operator can see which file that was.
	Root  string
	Items []TodoItemVM
}

// TodoItemVM is one backlog card. The five-field item contract the store holds
// is consumed as it stands; no field is added or renamed.
type TodoItemVM struct {
	ID     string
	Text   string
	State  string
	SpecID string
}

// buildTodo loads the backlog queue for the served project through the read
// seam. All three states are listed, none filtered out (resolved decision
// G-5): the audit view answers "where did card X go", which a queued-only
// working view cannot.
func (a *app) buildTodo() TodoVM {
	return readTodoQueue(a.cfg.ProjectRoot)
}
