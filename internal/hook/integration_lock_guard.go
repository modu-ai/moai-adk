// integration_lock_guard.go — the PreToolUse layer of the release-integration
// holder lock (card t194).
//
// Card t181 made release-tree integration serialize by announcement and named
// its own gap: announcement is a social protocol, so a lane that skips it is
// stopped by nothing, and the check that lane still runs (MERGE_HEAD printing
// nothing) is exactly the insufficient probe t181 exists to correct. This
// guard is the mechanical layer that closes it — the FIRST check, ahead of
// both MERGE_HEAD and the pre-commit HEAD re-read, which t181 keeps as the
// later ones.
//
// It is modelled on the branch-state guard next door and inherits its three
// load-bearing properties, for the same reasons:
//
//   - Opt-in, inert by default. Workflow.IntegrationLock.Enabled gates the
//     call site; on the disabled path no record is read at all.
//   - A deny sentinel the orchestrator can match without parsing prose.
//   - FAIL OPEN. A deny fires only on positive evidence that someone else
//     holds the window. Uncertainty — an unreadable record, an unknown
//     session, a missing project root — allows and says why on stderr. A
//     guard that blocked on uncertainty would wedge the batch it protects.
//
// One asymmetry with the branch guard is deliberate: an UNREADABLE record
// allows here, while `kanban.ReadIntegrationLock` treats the same record as a
// hard error for its CLI callers. The CLI is a lane asking "may I enter?",
// where refusing to answer is the safe reply; the guard is on a hot tool path,
// where the same refusal would deny every git merge in the repository until
// someone hand-repaired a JSON file. Same fact, different blast radius.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// integrationLockViolationPrefix is the deny sentinel. The orchestrator
// matches the prefix rather than the reason text, so the remediation wording
// can change without breaking recognition.
const integrationLockViolationPrefix = "INTEGRATION_LOCK_VIOLATION"

// integrationMergePattern matches the act the doctrine serializes: merging a
// lane branch into the release branch. Only `git merge` is matched — not
// commit, not push. A lane that has already merged is past the window's
// contended step, and widening the pattern would deny ordinary work in every
// worktree for as long as any lane holds the window.
var integrationMergePattern = regexp.MustCompile(`\bgit\s+merge\b`)

// checkIntegrationLock returns DecisionDeny plus an
// "INTEGRATION_LOCK_VIOLATION: ..." reason when ALL of the following hold:
// the command is a git merge, a record names a holder, that holder is not this
// session, and the holder is live. Anything else returns ("", "") — the allow
// fall-through.
func checkIntegrationLock(input *HookInput, projectRoot string) (decision string, reason string) {
	if input == nil || len(input.ToolInput) == 0 || projectRoot == "" {
		return "", ""
	}
	command := extractIntegrationCommand(input.ToolInput)
	if command == "" {
		return "", ""
	}
	// Scrub quoted spans before matching, reusing the branch guard's helper:
	// `echo 'git merge ...'` names the command without running it, and denying
	// it would be a false positive on a string. Measured — the first cut of
	// this guard did deny exactly that.
	if !integrationMergePattern.MatchString(substituteQuotedArguments(command)) {
		return "", ""
	}

	lock, err := kanban.ReadIntegrationLock(projectRoot)
	if err != nil {
		// Fail open, loudly. See the package comment: refusing every merge in
		// the repository because one JSON file is malformed is a worse
		// failure than the one this guard prevents.
		fmt.Fprintf(os.Stderr, "[moai:integration-lock] advisory: cannot read the integration lock (%v); allowing\n", err)
		return "", ""
	}
	if !lock.Held() {
		return "", ""
	}
	if input.SessionID != "" && lock.SessionID == input.SessionID {
		return "", "" // the holder is this session
	}
	if input.SessionID == "" {
		// An unidentified caller cannot be shown to be a non-holder. Allow,
		// and say so: silently allowing here would make the guard look
		// enforcing while it is not.
		fmt.Fprintf(os.Stderr, "[moai:integration-lock] advisory: this session has no id; cannot tell it from the holder (%s) — allowing\n", lock.SessionID)
		return "", ""
	}
	if lock.Stale() {
		fmt.Fprintf(os.Stderr, "[moai:integration-lock] advisory: holder %s (pid %d) is gone; allowing — reclaim with `moai integration acquire`\n", lock.SessionID, lock.PID)
		return "", ""
	}

	reason = fmt.Sprintf("%s: the release integration window is held by %s (pid %d) since %s on %s. Wait for its completion report, or take it over deliberately with `moai integration acquire --force`.",
		integrationLockViolationPrefix, holderLabelOf(lock), lock.PID, lock.AcquiredAt, lock.Branch)
	return DecisionDeny, reason
}

// holderLabelOf prefers the human-facing session name over the id, so a deny
// names the lane an operator would address in a dispatch.
func holderLabelOf(lock *kanban.IntegrationLock) string {
	if lock == nil {
		return "unknown"
	}
	if strings.TrimSpace(lock.SessionName) != "" {
		return lock.SessionName
	}
	if lock.SessionID != "" {
		return lock.SessionID
	}
	return "unknown"
}

// extractIntegrationCommand pulls the command string out of Bash tool input.
// Returns "" when the payload is not parseable or carries no command.
func extractIntegrationCommand(toolInput json.RawMessage) string {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return ""
	}
	command, _ := parsed["command"].(string)
	return command
}
