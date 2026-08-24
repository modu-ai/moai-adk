// todo_view.go — SPEC-WEB-TODO-QUEUE-001 M3: the read-only view model behind
// the /todo route.
//
// The console is a CONSUMER of the backlog queue, never an owner:
// SPEC-KANBAN-TODO-CLI-001 owns lock-guarded writes and id issuance, and this
// file reads through kanban.BacklogStore.Load, which takes no lock. The queue
// root arrives from kanban.ResolveTodoQueueRoot — the PURE resolver, which
// performs no filesystem mutation on any branch, so rendering a page never
// migrates the operator's backlog (REQ-WTQ-001, REQ-WTQ-004).
package web

import (
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TodoVM is the backlog queue as the section renders it.
type TodoVM struct {
	// Root is the directory the queue resolved to — the primary checkout, the
	// home-based fallback, or (read-through, REQ-WTQ-005) the project-local
	// root. Rendered as provenance: a console served from a worktree shows the
	// primary's cards, and the operator can see which file that was.
	Root  string
	Items []TodoItemVM
}

// TodoItemVM is one backlog card. The five-field item contract
// (kanban.BacklogItem) is consumed as it stands; no field is added or renamed.
type TodoItemVM struct {
	ID     string
	Text   string
	State  string
	SpecID string
}

// buildTodo loads the backlog queue for the served project.
//
// Every failure mode — an absent file, an empty file, malformed JSON — yields
// an empty queue rather than an error: the section is a read-only status view,
// and failing a whole page over an unreadable queue would trade a small gap for
// a blank screen (REQ-WTQ-006).
//
// All three states are listed, none filtered out (resolved decision G-5): the
// audit view answers "where did card X go", which a queued-only working view
// cannot.
func (a *app) buildTodo() TodoVM {
	root := kanban.ResolveTodoQueueRoot(a.cfg.ProjectRoot)
	vm := TodoVM{Root: root}
	rec, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).Load()
	if err != nil || rec == nil {
		return vm
	}
	for _, it := range rec.Items {
		row := TodoItemVM{ID: it.ID, Text: it.Text, State: string(it.State)}
		if it.SpecID != nil {
			row.SpecID = *it.SpecID
		}
		vm.Items = append(vm.Items, row)
	}
	return vm
}
