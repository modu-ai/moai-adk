// todo_queue_read.go — the console's SINGLE read seam onto the backlog queue.
//
// This is the only file in internal/web that names a backlog-store symbol
// (`kanban.ResolveTodoQueueRoot`, `kanban.NewBacklogStore`,
// `kanban.BacklogPathForRoot`, `kanban.BacklogItem`); `todo_queue_read_test.go`
// asserts that mechanically. The view model calls readTodoQueue and never the
// store, so the queue's storage — a JSON file today, under review — is swapped
// by changing this function and nothing else.
//
// Deliberately ONE function, not a layer: no interface, no factory, no plugin
// point. A single concrete reader needs no seam beyond its own signature.
package web

import (
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// readTodoQueue resolves the backlog queue for the served project and reads it.
//
// BOTH halves are pure. Resolution goes through the PURE root resolver, and the
// read goes through LoadPure, which serves whichever storage layout it finds
// and migrates nothing — so rendering a page never moves the operator's backlog
// (REQ-WTQ-001, REQ-WTQ-004). Calling the adopting Load here would make a page
// render perform the one-time storage cutover, which is the `moai todo`
// command path's act. Adoption stays reachable only from there. The read itself
// takes no lock — lock-guarded writes and id issuance belong to
// SPEC-KANBAN-TODO-CLI-001, and the console is a consumer.
//
// Every failure mode — an absent file, an empty file, malformed JSON — yields
// an empty queue rather than an error (REQ-WTQ-006). All three states are
// returned, none filtered out (resolved decision G-5); ordering is the store's.
func readTodoQueue(projectRoot string) TodoVM {
	root := kanban.ResolveTodoQueueRoot(projectRoot)
	vm := TodoVM{Root: root}
	rec, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).LoadPure()
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
