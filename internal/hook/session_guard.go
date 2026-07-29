package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// preEditAdvisoryLogRelPath is the fail-open advisory log path, relative to the
// project root, for the Pre-Edit Sync Check advisory hook
// (SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001 M4, REQ-PES-004).
const preEditAdvisoryLogRelPath = ".moai/logs/preedit-session-guard.log"

// checkForeignSessionAdvisory is the read-only ADVISORY PreToolUse-on-Edit check
// (SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001 M4, REQ-PES-004). It reads the
// active-sessions registry in-process, counts FOREIGN active sessions (those
// whose SessionID differs from the current session), and — when ≥1 foreign
// session is live — writes an advisory to stderr and appends a structured line
// to .moai/logs/preedit-session-guard.log.
//
// This check is FAIL-OPEN by construction: it ALWAYS returns ("", "") (no
// decision, no reason) so the edit proceeds regardless. It never blocks, never
// calls AskUserQuestion, and tolerates every error (missing registry file, parse
// failure, registry read failure) by returning silently. It is the mechanical
// companion to the procedural Pre-Edit Sync Check doctrine
// (agent-common-protocol.md § Pre-Edit Sync Check); the doctrine is the
// load-bearing rule and this hook is a non-blocking nudge.
//
// Stale-registry caveat: the registry uses heartbeat-age-only staleness
// (session.DefaultStaleMinutes). A dead-PID entry with a fresh-looking
// LastHeartbeat is counted as foreign — accepted as conservative over-count
// (a false positive is cheap; this hook does not block). No kill -0 liveness
// probe is performed, matching the registry's own Purge semantics.
//
// @MX:NOTE this advisory is read-only and fail-open (REQ-PES-004).
func checkForeignSessionAdvisory(input *HookInput, projectDir string) (decision string, reason string) {
	// Fail-open by construction: every error path returns the empty allow.
	if input == nil || projectDir == "" {
		return "", ""
	}
	regPath := filepath.Join(projectDir, session.DefaultRegistryPath)
	reg := session.NewRegistry(regPath, nil)
	entries, err := reg.Query("")
	if err != nil {
		// Missing/unreadable/empty registry — fail-open silently.
		return "", ""
	}
	ownID := input.SessionID
	others := make([]session.Entry, 0, len(entries))
	for _, e := range entries {
		if e.SessionID != ownID {
			others = append(others, e)
		}
	}
	if len(others) == 0 {
		return "", ""
	}
	appendPreEditAdvisory(input, projectDir, others)
	return "", ""
}

// appendPreEditAdvisory writes a fail-open advisory to stderr and appends a
// structured entry to <projectDir>/.moai/logs/preedit-session-guard.log, modeled
// on appendBranchGuardAdvisory (REQ-WBG-012). Logging errors are swallowed —
// fail-open must never block the hook's allow decision.
func appendPreEditAdvisory(input *HookInput, projectDir string, others []session.Entry) {
	sessionID := ""
	toolName := ""
	if input != nil {
		sessionID = input.SessionID
		toolName = input.ToolName
	}
	msg := fmt.Sprintf("preedit_session_guard: %d foreign active session(s) detected during %s edit (this session %s) — Pre-Edit Sync Check applies; see agent-common-protocol.md § Pre-Edit Sync Check",
		len(others), toolName, shortGuardSessionID(sessionID))
	fmt.Fprintln(os.Stderr, msg)

	entry := fmt.Sprintf("[%s] session=%s tool=%s foreign_count=%d\n",
		time.Now().UTC().Format(time.RFC3339), sessionID, toolName, len(others))
	logPath := filepath.Join(projectDir, preEditAdvisoryLogRelPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(entry)
}

// shortGuardSessionID returns an 8-char prefix of a session id for compact log
// output, mirroring the internal session.shortID used by FormatStderrReminder.
func shortGuardSessionID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}
