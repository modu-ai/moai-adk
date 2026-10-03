// todo_queue_read.go — the console's SINGLE read seam onto the backlog queue.
//
// This is the only file in internal/web that names a backlog-store symbol
// (`factory.ResolveTodoQueueRoot`, `factory.NewBacklogStore`,
// `factory.BacklogPathForRoot`, `factory.BacklogItem`); `todo_queue_read_test.go`
// asserts that mechanically. The view model calls readTodoQueue and never the
// store, so the queue's SQLite storage and legacy read-through are swapped
// by changing this function and nothing else.
//
// Deliberately ONE function, not a layer: no interface, no factory, no plugin
// point. A single concrete reader needs no seam beyond its own signature.
package web

import (
	"github.com/modu-ai/moai-adk/internal/factory"
)

// readTodoQueue resolves the backlog queue for the served project and reads it.
//
// BOTH halves are pure. Resolution goes through the PURE root resolver, and the
// read goes through LoadPure, which serves whichever storage layout it finds
// and migrates nothing — so rendering a page never moves the operator's backlog
// (REQ-WTQ-001, REQ-WTQ-004). Calling the adopting Load here would make a page
// render perform the one-time storage cutover, which is the `moai todo`
// command path's act. Adoption stays reachable only from there. The read itself
// takes no lock — lock-guarded writes and id issuance belong to the
// `moai todo` command, and the console is a consumer.
//
// An absent queue is empty; a failed read is unavailable, never a zero count.
// The view receives no raw error text. All three states are
// returned, none filtered out (resolved decision G-5); ordering is the store's.
func readTodoQueue(projectRoot string) TodoVM {
	root := factory.ResolveTodoQueueRoot(projectRoot)
	vm := TodoVM{Root: root}
	rec, err := factory.NewBacklogStore(factory.BacklogPathForRoot(root)).LoadPure()
	if err != nil {
		vm.Unavailable = true
		return vm
	}
	if rec == nil {
		return vm
	}
	for _, it := range rec.Items {
		row := TodoItemVM{ID: it.ID, Text: it.Text, State: string(it.State)}
		if it.SpecID != nil {
			row.SpecID = *it.SpecID
		}
		for _, f := range rec.Findings {
			if f.Names(it.ID) {
				row.Relations = append(row.Relations, todoRelationCell(f, it.ID))
			}
		}
		vm.Items = append(vm.Items, row)
	}
	return vm
}

// todoRelationCell renders one recorded finding as the display line the todo
// row shows. The addressed card is implied by its own row: on the subject's
// row the recorded direction reads forward ("blocks t2"), and on the
// counterpart's row the original direction is kept and marked as the other
// side of the record ("t1 blocks this").
func todoRelationCell(f factory.BacklogFinding, self string) string {
	if f.SubjectID == self {
		return f.Relation + " " + f.RelatedID + " (" + f.Source + ")"
	}
	return f.SubjectID + " " + f.Relation + " this (" + f.Source + ")"
}
