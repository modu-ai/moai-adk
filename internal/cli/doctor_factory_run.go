package cli

// doctor_factory_run.go — the doctor factory section (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-001, AC-RNC-013). It reads the kanban session records — the same
// store the web view model reads — and reports the run's leader role in the
// persisted vocabulary:
//
//   - role `leader`  → ok, the displayed label is `leader` and the run is
//     recognized as present.
//   - role `lead`    → fail, the message carries the literal
//     "legacy run: relaunch required" (REQ-RNC-009: detection only, never a
//     role mapping — the remedy is a relaunch under the current vocabulary).
//   - no leader record of either kind → warn when other records exist
//     (lanes or companions without a leader), info when the project holds no
//     records at all.

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryRunCheckName is the `--check` selector for the doctor factory
// section.
const factoryRunCheckName = "Factory Run"

// legacyLeaderRelaunchMessage is the AC-RNC-013 literal the doctor factory
// section carries for a persisted legacy `lead` leader role.
const legacyLeaderRelaunchMessage = "legacy run: relaunch required"

// checkFactoryRun reports the leader role of the declared run, read from the
// kanban session records under the project root. Read-only: it never writes
// and never repairs.
func checkFactoryRun(cwd string, verbose bool) DiagnosticCheck {
	records, err := kanban.ReadAll(cwd)
	if err != nil {
		return DiagnosticCheck{
			Name:    factoryRunCheckName,
			Status:  uikit.CheckWarn,
			Message: "cannot read session records",
			Detail:  err.Error(),
		}
	}
	if len(records) == 0 {
		return DiagnosticCheck{
			Name:    factoryRunCheckName,
			Status:  uikit.CheckInfo,
			Message: "no session records — no factory or kanban run declared",
		}
	}

	var leader, legacy *kanban.Record
	laneCount := 0
	for i := range records {
		switch strings.ToLower(strings.TrimSpace(records[i].Role)) {
		case "leader":
			leader = &records[i]
		case "lead":
			// Legacy value, detection only: a pre-rename leader record is
			// reported as "legacy run: relaunch required", never as a present
			// leader (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009/-013).
			legacy = &records[i]
		case "lane":
			laneCount++
		}
	}

	switch {
	case leader != nil:
		return DiagnosticCheck{
			Name:    factoryRunCheckName,
			Status:  uikit.CheckOK,
			Message: "run present",
			Detail:  fmt.Sprintf("role label: leader (session %s); %d lane record(s)", leader.SessionID, laneCount),
		}
	case legacy != nil:
		return DiagnosticCheck{
			Name:    factoryRunCheckName,
			Status:  uikit.CheckFail,
			Message: legacyLeaderRelaunchMessage,
			Detail: fmt.Sprintf("session %s declares the legacy leader role %q — end its sessions and relaunch "+
				"under the leader/lane vocabulary; a new-vocabulary binary refuses to join this run",
				legacy.SessionID, legacy.Role),
		}
	default:
		return DiagnosticCheck{
			Name:    factoryRunCheckName,
			Status:  uikit.CheckWarn,
			Message: "no leader record found",
			Detail:  fmt.Sprintf("%d record(s) present, none carrying the leader role", len(records)),
		}
	}
}
