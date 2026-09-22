// subagent_write_guard.go — PreToolUse guard refusing a subagent's Write that
// drastically shrinks an existing tracked file
// (SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001).
//
// The guard keys on the DESTRUCTIVENESS of the write, never on path scope
// (spec.md §B): a subagent may overwrite a file it was asked to change, but a
// blind full-file overwrite that replaces an existing tracked file with a
// fraction of what it held is the shape refused here. The discriminant is
// agent_id (NOT agent_type — spec.md §A.5: a `claude --agent <name>` main
// session carries agent_type but no agent_id).
//
// Family contract (settings_drift_gate / agent_model_guard / agent_stop_guard
// in internal/config/defaults.go): detection and observation run
// unconditionally; only the deny layer is gated on
// workflow.subagent_write_guard.enabled, shipped default-false.
package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// swgViolationPrefix begins every deny reason this guard emits (REQ-SWG-005).
const swgViolationPrefix = "SUBAGENT_DESTRUCTIVE_WRITE_VIOLATION:"

// SWG-T1 / SWG-T2 — the destructive-overwrite threshold pair (spec.md §D.2).
// Unit: BYTES (REQ-SWG-004) — lines are never the unit. Calibration: a
// PROPOSAL calibrated against a single measured incident (a subagent Write
// replaced a 417-line tracked file with a 6-line stub) — NOT a distribution
// measurement. M1 false-positive survey result over committed history: "22 survivors, 0 destructive accidents, 22 deliberate, 0 in-place SPEC amendments".
// Adopted by operator decision OD-1a (2026-09-22, spec.md §D.3); a later
// calibration card may re-measure via the withheld-log instrument
// (REQ-SWG-008a).
const (
	// swgSizeFloorBytes is SWG-T1: the existing file must be at or above this
	// many bytes for the floor condition to hold.
	swgSizeFloorBytes = 2000
	// swgShrinkRatioPercent is SWG-T2: the incoming byte length must be at or
	// below this percentage of the existing byte length.
	swgShrinkRatioPercent = 25
)

// Decision values for the audit row (REQ-SWG-008a's closed set). The four are
// mutually distinguishable by a reader of the log alone: `withheld` is NOT a
// variant of `allow` — it is how often the guard WOULD have fired with the
// deny layer on, the OD-1b calibration instrument.
const (
	swgDecisionDeny     = "deny"
	swgDecisionWithheld = "withheld"
	swgDecisionAllow    = "allow"
	swgDecisionFailOpen = "fail-open"
)

// swgWritePayload carries the structured fields REQ-SWG-001 allows the guard
// to read: file_path and content. Nothing else in the payload is interpreted,
// and no pattern is ever matched against any command text (spec.md §B.2).
type swgWritePayload struct {
	FilePath   string
	Content    string
	HasContent bool
}

// extractSubagentWritePayload parses the Write tool_input JSON into the
// structured fields REQ-SWG-001 allows.
func extractSubagentWritePayload(toolInput json.RawMessage) (swgWritePayload, error) {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return swgWritePayload{}, err
	}
	payload := swgWritePayload{}
	if v, ok := parsed["file_path"].(string); ok {
		payload.FilePath = v
	}
	if v, ok := parsed["content"].(string); ok {
		payload.Content = v
		payload.HasContent = true
	}
	return payload, nil
}

// swgEvaluation carries the outcome of one guard evaluation: the decision from
// REQ-SWG-008a's closed set, the deny or fail-open reason, and the context the
// audit row names.
type swgEvaluation struct {
	Decision  string // deny | withheld | allow | fail-open ("" when not evaluated)
	Reason    string // deny reason, or the fail-open cause
	FilePath  string
	PreBytes  int64 // existing file's byte length on disk
	PostBytes int64 // incoming content's byte length
}

// evaluateSubagentWrite runs the guard over one payload and returns the
// evaluation. denyEnabled selects the deny layer only (REQ-SWG-006): with it
// false a predicate-holding write is recorded `withheld`, never denied.
//
// @MX:NOTE: [AUTO] the predicate's evaluation order is load-bearing — the two
// payload-derived size conditions are decided BEFORE any git subprocess
// (REQ-SWG-004). @MX:REASON: one `git rev-parse` per kanban session moved
// SessionStart from under 500ms to 650-890ms against a 5s hook budget
// (session_start_record.go), and the Write path fires far more often.
func evaluateSubagentWrite(input *HookInput, denyEnabled bool) swgEvaluation {
	// REQ-SWG-003: a payload with an empty agent_id is not evaluated at all,
	// whatever agent_type holds — a main-session write is outside this
	// guard's subject (spec.md §A.5).
	if input == nil || input.AgentID == "" {
		return swgEvaluation{}
	}
	ev := swgEvaluation{}

	payload, err := extractSubagentWritePayload(input.ToolInput)
	if err != nil {
		return swgFailOpen(ev, "payload unparseable", err)
	}
	ev.FilePath = payload.FilePath
	if payload.FilePath == "" {
		return swgFailOpen(ev, "file_path absent", nil)
	}
	if !payload.HasContent {
		return swgFailOpen(ev, "content field absent", nil)
	}
	ev.PostBytes = int64(len(payload.Content))

	// Existing size is the file's byte length on disk (REQ-SWG-004). Any stat
	// failure is fail-open (REQ-SWG-007 "an unreadable existing file"): the
	// predicate needs the on-disk size as positive evidence.
	absPath := payload.FilePath
	if !filepath.IsAbs(absPath) {
		cwd := resolveProjectRootFromInputOrEnv(input, "subagent_write_guard")
		if cwd == "" {
			return swgFailOpen(ev, "relative file_path without a resolvable cwd", nil)
		}
		absPath = filepath.Join(cwd, absPath)
	}
	stat, err := os.Stat(absPath)
	if err != nil {
		return swgFailOpen(ev, "existing file unreadable", err)
	}
	ev.PreBytes = stat.Size()

	// Conditions 1+2 (REQ-SWG-004 order): payload-derived sizes first, so the
	// git subprocess below is reached only by writes already destructively
	// shaped. Integer math avoids float rounding at the boundary.
	if ev.PostBytes*100 > ev.PreBytes*swgShrinkRatioPercent {
		return swgAllow(ev) // condition 1 fails: not a shrink
	}
	if ev.PreBytes < swgSizeFloorBytes {
		return swgAllow(ev) // condition 2 fails: below the floor
	}

	// Condition 3: the target resolves to a path inside the repository being
	// worked in. The repository is resolved from the payload cwd the way the
	// guard family resolves it (input.CWD → CLAUDE_PROJECT_DIR → os.Getwd()).
	cwd := resolveProjectRootFromInputOrEnv(input, "subagent_write_guard")
	if cwd == "" {
		return swgFailOpen(ev, "cwd unresolvable", nil)
	}
	repoRoot, err := swgGitRepoRoot(cwd)
	if err != nil {
		return swgFailOpen(ev, "not a git repository or git unavailable", err)
	}
	rel, outside, err := swgRelInsideRepo(repoRoot, absPath)
	if err != nil {
		return swgFailOpen(ev, "path resolution failed", err)
	}
	if outside {
		return swgAllow(ev) // condition 3 fails: target outside the repository
	}

	// Condition 4: tracked by git at HEAD. A git failure here is an
	// uncertainty (fail-open); an empty ls-tree is a clean untracked answer
	// (REQ-SWG-010).
	tracked, err := swgTrackedAtHEAD(repoRoot, rel)
	if err != nil {
		return swgFailOpen(ev, "git query failed while answering tracked-at-HEAD", err)
	}
	if !tracked {
		return swgAllow(ev)
	}

	// All four conditions hold. Only the refusal is gated (REQ-SWG-006).
	if !denyEnabled {
		ev.Decision = swgDecisionWithheld
		return ev
	}
	ev.Decision = swgDecisionDeny
	ev.Reason = fmt.Sprintf(
		"%s: subagent Write would replace %s (existing %d bytes) with %d bytes. Re-issue the change as an Edit carrying the exact prior text, or perform the write from the main session.",
		swgViolationPrefix, ev.FilePath, ev.PreBytes, ev.PostBytes)
	return ev
}

func swgAllow(ev swgEvaluation) swgEvaluation {
	ev.Decision = swgDecisionAllow
	return ev
}

func swgFailOpen(ev swgEvaluation, what string, cause error) swgEvaluation {
	ev.Decision = swgDecisionFailOpen
	if cause != nil {
		ev.Reason = what + ": " + cause.Error()
	} else {
		ev.Reason = what
	}
	return ev
}

// swgGitRepoRoot resolves the repository root the cwd works in via
// `git rev-parse --show-toplevel`, through the gitcore.ExecCommand seam tests
// can observe and stub. Any failure is the caller's fail-open signal.
func swgGitRepoRoot(cwd string) (string, error) {
	return swgRunGit(cwd, "rev-parse", "--show-toplevel")
}

// swgTrackedAtHEAD answers whether relPath is tracked at HEAD in repoDir via
// `git ls-tree`, whose exit status separates "git broken" (error → the
// caller's fail-open) from "path absent at HEAD" (exit 0, empty output →
// untracked, REQ-SWG-010).
func swgTrackedAtHEAD(repoDir, relPath string) (bool, error) {
	out, err := swgRunGit(repoDir, "ls-tree", "--name-only", "HEAD", "--", relPath)
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// swgRelInsideRepo returns rel (repoRoot-relative) when absPath sits inside
// repoRoot, outside=true when it positively sits outside, and an error when
// containment cannot be decided. Both sides are symlink-normalized first: on
// macOS the temp dir hands out /var/... while `git rev-parse --show-toplevel`
// prints /private/var/..., and an un-normalized Rel would misreport
// containment (the path-handling lesson the git worktree test suite carries).
func swgRelInsideRepo(repoRoot, absPath string) (rel string, outside bool, err error) {
	rootReal, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", false, err
	}
	pathReal, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", false, err
	}
	rel, err = filepath.Rel(rootReal, pathReal)
	if err != nil {
		return "", false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", true, nil
	}
	return filepath.ToSlash(rel), false, nil
}

// swgRunGit runs `git -C dir <args...>` through the gitcore.ExecCommand
// indirection (the same seam the branch-guard fallback suite injects) and
// returns trimmed stdout.
func swgRunGit(dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir}, args...)
	cmd := gitcore.ExecCommand("git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v (%s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// subagentWriteGuardEnabled reports whether the deny layer is opted in, the
// family shape (nil provider and nil config both read as disabled).
func (h *preToolHandler) subagentWriteGuardEnabled() bool {
	if h.cfg == nil {
		return false
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return false
	}
	return cfg.Workflow.SubagentWriteGuard.Enabled
}

// checkSubagentDestructiveWrite evaluates the guard for one Write payload
// with the configured deny layer, appends the audit row, and returns the deny
// reason ("" when no deny).
func (h *preToolHandler) checkSubagentDestructiveWrite(input *HookInput) string {
	ev := evaluateSubagentWrite(input, h.subagentWriteGuardEnabled())
	if ev.Decision == "" {
		return "" // main-session payload: never evaluated (REQ-SWG-003)
	}
	appendSubagentWriteGuardAudit(h.projectRoot(), input, ev)
	if ev.Decision == swgDecisionDeny {
		slog.Warn("subagent destructive write denied",
			"tool_name", input.ToolName,
			"session_id", input.SessionID,
			"file_path", ev.FilePath,
			"reason", ev.Reason,
		)
		return ev.Reason
	}
	return ""
}

// swgAuditRelPath is the guard's audit log, relative to the audit project
// directory (the $CLAUDE_PROJECT_DIR → os.Getwd() resolution, matching the
// branch guard's central-logging pin).
const swgAuditRelPath = ".moai/logs/subagent-write-guard.log"

// appendSubagentWriteGuardAudit appends one structured row per decision to
// <auditDir>/.moai/logs/subagent-write-guard.log. The append runs on EVERY
// decision — deny, fail-open, allow, and withheld alike — whether or not the
// deny layer is enabled (REQ-SWG-008): the rows accumulate unconditionally,
// which is the guard's continued-firing signal and the source of the
// `withheld` count (the OD-1b calibration instrument, REQ-SWG-008a). The
// derived card id is audit CONTEXT only and gates nothing (REQ-SWG-013) —
// cardIDFromPath does not verify the card exists, so an unresolved id is
// recorded empty and changes no outcome. Logging errors are debug-level:
// a failed append must never turn into a deny or block the allow.
func appendSubagentWriteGuardAudit(auditDir string, input *HookInput, ev swgEvaluation) {
	if auditDir == "" {
		slog.Debug("subagent_write_guard: no audit dir resolved; row not written",
			"decision", ev.Decision)
		return
	}
	sessionID, agentID, agentType := "", "", ""
	cwd := ""
	if input != nil {
		sessionID = input.SessionID
		agentID = input.AgentID
		agentType = input.AgentType
		cwd = input.CWD
	}
	cardID := cardIDFromPath(cwd)
	entry := fmt.Sprintf("[%s] session=%s decision=%s agent_id=%s agent_type=%q path=%q pre=%d post=%d card=%s reason=%q\n",
		time.Now().UTC().Format(time.RFC3339),
		sessionID, ev.Decision, agentID, agentType,
		ev.FilePath, ev.PreBytes, ev.PostBytes, cardID, ev.Reason)
	logPath := filepath.Join(auditDir, swgAuditRelPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		slog.Debug("subagent_write_guard: could not create audit log dir", "path", logPath, "error", err)
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Debug("subagent_write_guard: could not open audit log", "path", logPath, "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(entry); err != nil {
		slog.Debug("subagent_write_guard: could not write audit log entry", "path", logPath, "error", err)
	}
}
