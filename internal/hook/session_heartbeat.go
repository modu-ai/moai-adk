package hook

import (
	"log/slog"
	"os"

	"github.com/modu-ai/moai-adk/internal/session"
)

// HeartbeatSeamUserPromptSubmit refreshes this session's registry heartbeat.
//
// Why this exists: session.Heartbeat had NO production caller. Measured on this
// repository (2026-09-20, GH #1711) the sole non-test caller was the manual
// `moai session heartbeat` verb, against 39 call sites for Register — so
// last_heartbeat froze at registration for the whole life of a session. The
// registry file was being rewritten continuously while every timestamp in it
// stayed hours old (file mtime 15:59:41Z against a newest last_heartbeat of
// 06:55:11Z), and 139 of 147 live entries still carried
// last_heartbeat == started_at.
//
// Why UserPromptSubmit: it fires once per user turn, which is the cheapest
// cadence that still means "this session is actively being driven". The
// registry write takes an exclusive advisory lock, so a per-tool-call seam
// (PreToolUse/PostToolUse) or a per-render seam (the statusline) would pay that
// lock orders of magnitude more often across every concurrent session for
// freshness no reader needs. The cost of the cadence is bounded and stated: a
// session driven entirely by an armed goal loop, with no user prompt between
// turns, does not beat — its liveness is carried by the PID probe, which is
// the direct evidence anyway (see internal/web sessionState, session.
// LiveAnchoredSessions).
//
// Fail-open by construction, like the routing seam beside it: every error path
// returns silently, and a session with no registry entry is a no-op because
// Heartbeat is idempotent on a missing entry (REQ-COORD-004). The existence
// check keeps the seam from creating an empty registry file in a project that
// never registered one.
func HeartbeatSeamUserPromptSubmit(input *HookInput) {
	if input == nil || input.SessionID == "" {
		return
	}
	root := resolveProjectRoot(input)
	if root == "" {
		return
	}
	path := session.RegistryPathFor(root)
	if _, err := os.Stat(path); err != nil {
		// No registry to beat against — SessionStart never registered here.
		return
	}
	if err := session.NewRegistry(path, nil).Heartbeat(input.SessionID); err != nil {
		slog.Debug("session heartbeat seam: refresh failed (non-blocking)",
			"session_id", input.SessionID,
			"registry", path,
			"error", err.Error(),
		)
	}
}
