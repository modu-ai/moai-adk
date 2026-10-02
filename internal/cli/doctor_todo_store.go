// doctor_todo_store.go — SPEC-TODO-STALE-STORE-001 (card t1307) M2: the
// doctor divergence check over a stale project-local queue store.
//
// After the home-database cutover a rollback snapshot can remain under
// .moai/state/todo/ or .moai/state/kanban/ while every read is answered by
// ~/.moai/db/<project-key>/todo/backlog.db. The read surface discloses the
// ghost store (M1); this check makes the SAME divergence reproducible with
// one doctor line (REQ-TSS-010), from the SAME detector the disclosure
// uses (REQ-TSS-004) — no second inspector.
//
// Read-only on every branch (REQ-TSS-013): the detector opens neither
// store's engine and takes no lock, so the check can never migrate, DDL,
// or lock the queue it is diagnosing.
package cli

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// todoStoreDivergenceCheckName is the doctor check identifier (also the
// value accepted by `moai doctor --check`).
const todoStoreDivergenceCheckName = "Todo Store Divergence"

// checkTodoStoreDivergence judges the stale-local-store fact for
// projectRoot. Verdict order: no legacy store is OK (the steady state);
// divergence is FAIL and names the store path with BOTH last_seq values; a
// missing home database is FAIL; an unreadable legacy store is WARN — its
// own state, never folded into divergence (acceptance §D.2); matching
// last_seq values are OK.
func checkTodoStoreDivergence(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: todoStoreDivergenceCheckName, Status: uikit.CheckOK}

	fact := factory.InspectStaleLocalStores(projectRoot)
	if len(fact.Stores) == 0 {
		check.Message = "no stale project-local queue store"
		return check
	}

	if fact.Divergent {
		var lines []string
		for _, st := range fact.Stores {
			if !st.Readable || st.LastSeq == fact.HomeLastSeq {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: local last_seq %d vs home last_seq %d",
				st.Path, st.LastSeq, fact.HomeLastSeq))
		}
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("stale project-local queue store diverges from the home database (%d store(s))", len(lines))
		check.Detail = "The store is a rollback snapshot, NOT the queue; reads are answered by " + fact.HomePath + ". " + strings.Join(lines, "; ")
		return check
	}

	if !fact.HomePresent {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("home queue database missing at %s while stale project-local store(s) exist", fact.HomePath)
		return check
	}

	var unreadable []string
	for _, st := range fact.Stores {
		if !st.Readable {
			unreadable = append(unreadable, st.Path)
		}
	}
	if len(unreadable) > 0 {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("%d stale project-local store(s) present but unreadable (zero-byte or not a queue database)", len(unreadable))
		check.Detail = "Reported as its own state, not divergence: an unreadable store carries no comparable last_seq. Path(s): " + strings.Join(unreadable, ", ")
		return check
	}

	check.Message = fmt.Sprintf("stale project-local store(s) present with last_seq matching the home database (%d)", fact.HomeLastSeq)
	return check
}
