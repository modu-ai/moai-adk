package statusline

// goal_armed.go implements the SPEC-INFINITE-GOAL-001 REQ-3 goal-armed signal
// for the statusline. The builder resolves whether an armed goal exists for the
// current session and sets StatusData.GoalArmed; the renderer then suppresses
// the /clear directive markers while a goal is armed (auto-compact handles
// context pressure instead).
//
// The statusline package MUST NOT import internal/goal (that would couple a
// render-time hot path to the goal engine and pull in its state-file layout).
// Instead this file performs a minimal, best-effort read of the goal state
// file and inspects only the Status field. Fail-open: any read/parse error or
// a non-"armed" status → GoalArmed=false (markers shown, backward compat).

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// goalStateFileSnippet is the minimal subset of the goal state JSON the
// statusline inspects. Only Status is load-bearing for the suppression gate.
type goalStateFileSnippet struct {
	Status string `json:"status"`
}

// resolveGoalArmed reports whether an armed goal exists for sessionID under
// projectRoot. It reads .moai/state/goal/<sessionID>.json and returns true only
// when the file exists AND its status == "armed". Best-effort + fail-open:
// absent file, read error, parse error, or a non-armed status all return false.
// Constant-cost (one stat + one read of a small file) per statusline render.
func resolveGoalArmed(projectRoot, sessionID string) bool {
	if projectRoot == "" || sessionID == "" {
		return false
	}
	path := filepath.Join(projectRoot, ".moai", "state", "goal", sessionID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return false // absent or unreadable → not armed
	}
	var s goalStateFileSnippet
	if err := json.Unmarshal(data, &s); err != nil {
		return false // corrupt → not armed (fail-open)
	}
	return s.Status == "armed"
}
