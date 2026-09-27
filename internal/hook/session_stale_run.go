// session_stale_run.go — the stale-run notice (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-022 hook clause, REQ-RNC-025).
//
// A session whose LAUNCH LABEL or whose existing SESSION RECORD carries a
// legacy role value (`lead`, `worker`, `agent` — pre-rename vocabulary) must
// never be silently adopted into the new vocabulary: SessionStart emits the
// stale-run notice instead of a leader or lane notice, and the session-record
// writer leaves the record absent or byte-identical. The remedy is a relaunch
// (factory runs first via `moai factory runs --retire`).
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// legacyLaunchLabelValue returns the launch-label value carrying a legacy
// role, or "" when every label in the environment is current-vocabulary (or
// absent). The leader label (MOAI_KANBAN_LEAD_NAME — kanban and factory
// leaders alike) and the lane label (MOAI_FACTORY_WORKER) are the two
// role-bearing launch labels.
func legacyLaunchLabelValue() string {
	if label := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLeadName)); kanban.IsLegacyLeadLabel(label) {
		return label
	}
	if label := strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker)); kanban.IsLegacyFactoryRoleValue(label) {
		return label
	}
	return ""
}

// isLegacyRecordRole reports whether a session record's role value is legacy
// vocabulary — detection only, never a role mapping (REQ-RNC-009).
func isLegacyRecordRole(role string) bool {
	return role == "lead" || role == "worker" || role == "agent"
}

// staleRunNoticeFor returns the stale-run notice for this session when its
// launch label OR its existing session record carries a legacy role value,
// and "" when the session is current-vocabulary (or not a run member).
func staleRunNoticeFor(root, sessionID string) string {
	if label := legacyLaunchLabelValue(); label != "" {
		return staleRunNotice(label)
	}
	if sessionID != "" && root != "" {
		if rec, err := kanban.Read(root, sessionID); err == nil && isLegacyRecordRole(rec.Role) {
			return staleRunNotice(rec.Role)
		}
	}
	return ""
}

// staleRunNotice renders the stale-run message. The factory retire step is
// named only for factory sessions (the run id + MOAI_FACTORY_WORKERS
// discriminator): a kanban run has no factory run to retire — its remedy is
// ending the session and relaunching. M1 wording is minimal English; M3
// polishes the locale text (plan.md §F M3).
func staleRunNotice(value string) string {
	runID := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanID))
	if runID != "" && os.Getenv(config.EnvMoaiFactoryWorkers) != "" {
		return fmt.Sprintf("stale run: this session carries legacy role value %q from a binary before the leader/lane rename — "+
			"end this session, retire the run with 'moai factory runs --retire %s', then relaunch", value, runID)
	}
	return fmt.Sprintf("stale run: this session carries legacy role value %q from a binary before the leader/lane rename — "+
		"end this session and relaunch it under the current vocabulary", value)
}

// legacyFactoryHookNotice is the factory message hook's stale-run answer: a
// lane whose label is legacy never registers a broker peer — it reports the
// stale-run condition instead.
func legacyFactoryHookNotice(label, runID string) string {
	if !kanban.IsLegacyFactoryRoleValue(strings.TrimSpace(label)) {
		return ""
	}
	if runID != "" {
		return fmt.Sprintf("stale run: lane label %q is legacy vocabulary from a binary before the leader/lane rename — "+
			"end this session, retire the run with 'moai factory runs --retire %s', then relaunch", label, runID)
	}
	return staleRunNotice(label)
}
