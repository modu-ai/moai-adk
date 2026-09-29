// doctor_owner_label.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// REQ-TSP-051: the doctor owner-label drift check.
//
// The todo_runtime_assignments table once accumulated the pre-rename
// spellings; the migration (REQ-TSP-050) relabels them, and this check is
// the reproducible verdict that it happened: it counts the rows still
// carrying a legacy spelling, using the SAME detectors the refusal and
// stale-record paths use (kanban.IsLegacyLeaderSpelling /
// kanban.IsLegacyFactoryRoleValue — no second detector), and reports zero
// rows as OK. Read-only: it judges, it never repairs.
package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// ownerLabelDriftCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const ownerLabelDriftCheckName = "Owner Label Drift"

// checkOwnerLabelDrift counts the legacy owner_label rows in the queue's
// todo_runtime_assignments table.
func checkOwnerLabelDrift(projectRoot string, verbose bool) DiagnosticCheck {
	// The PURE path: the doctor judges the queue it would read, it never
	// adopts or relocates one.
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(projectRoot))
	rec, err := store.LoadPure()
	if err != nil {
		return DiagnosticCheck{
			Name:    ownerLabelDriftCheckName,
			Status:  uikit.CheckWarn,
			Message: "cannot read the queue database",
			Detail:  err.Error(),
		}
	}

	counts := map[string]int{}
	for i := range rec.Runtime.Assignments {
		label := rec.Runtime.Assignments[i].OwnerLabel
		if kanban.IsLegacyLeaderSpelling(label) || kanban.IsLegacyFactoryRoleValue(label) {
			counts[label]++
		}
	}
	if len(counts) == 0 {
		return DiagnosticCheck{
			Name:    ownerLabelDriftCheckName,
			Status:  uikit.CheckOK,
			Message: "no legacy owner_label rows",
		}
	}

	labels := make([]string, 0, len(counts))
	total := 0
	for label, n := range counts {
		labels = append(labels, label)
		total += n
	}
	sort.Strings(labels)
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, fmt.Sprintf("%s x%d", label, counts[label]))
	}
	return DiagnosticCheck{
		Name:    ownerLabelDriftCheckName,
		Status:  uikit.CheckWarn,
		Message: fmt.Sprintf("%d owner_label row(s) still carry a legacy spelling", total),
		Detail:  "The migration (MigrateOwnerLabelVocabulary) relabels lead/worker-N/agent-N to leader/lane-N in one locked write. Rows: " + strings.Join(parts, ", "),
	}
}
