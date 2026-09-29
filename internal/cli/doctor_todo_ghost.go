// doctor_todo_ghost.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// REQ-TSP-042: the doctor ghost-inventory check.
//
// The divergence check beside it (t1307) judges the SQLite facts; this one
// inventories the non-SQLite ghost artifacts the SAME detector reports
// (StaleStoreFact.Ghosts — REQ-TSS-004's no-second-inspector rule): the
// path, class, and byte size of every leftover, judged in three states —
// absent OK, present WARN, an artifact that could not be read FAIL. It
// never deletes, moves, or rewrites anything: disposal stays an operator
// act (REQ-TSP-040's read-only discipline).
package cli

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoGhostInventoryCheckName is the doctor check identifier (also the
// value accepted by `moai doctor --check`).
const todoGhostInventoryCheckName = "Todo Ghost Inventory"

// checkTodoGhostInventory inventories the ghost artifacts under projectRoot.
func checkTodoGhostInventory(projectRoot string, verbose bool) DiagnosticCheck {
	fact := kanban.InspectStaleLocalStores(projectRoot)
	if len(fact.Ghosts) == 0 {
		return DiagnosticCheck{
			Name:    todoGhostInventoryCheckName,
			Status:  uikit.CheckOK,
			Message: "no ghost queue artifacts",
		}
	}

	var unreadable []string
	var lines []string
	for _, g := range fact.Ghosts {
		if g.Bytes < 0 {
			unreadable = append(unreadable, g.Path)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s (%s, %d bytes)", g.Path, g.Class, g.Bytes))
	}
	if len(unreadable) > 0 {
		return DiagnosticCheck{
			Name:    todoGhostInventoryCheckName,
			Status:  uikit.CheckFail,
			Message: fmt.Sprintf("%d ghost artifact(s) present but unreadable — the inventory is contradicted by the filesystem", len(unreadable)),
			Detail:  strings.Join(unreadable, ", "),
		}
	}

	return DiagnosticCheck{
		Name:    todoGhostInventoryCheckName,
		Status:  uikit.CheckWarn,
		Message: fmt.Sprintf("%d ghost queue artifact(s) present — leftovers of the storage cutover; removal is an operator act", len(fact.Ghosts)),
		Detail:  strings.Join(lines, "; "),
	}
}
