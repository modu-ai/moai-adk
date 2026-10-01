// anchor_trace.go — SPEC-SESSION-ANCHOR-ATTR-001 W3: the MOAI_ANCHOR_TRACE
// switch (REQ-SAA-007..009).
//
// The t1337/t1339 misresolution class left three items honestly unknown (the
// runtime anchor's keying, which layer defended the wrong-tree write, the
// intermittency dynamics). This instrument is how the NEXT occurrence gets
// measured instead of guessed at: with the switch on, every anchor decision
// point in the repo-owned surface — the branch-guard's Seam A anchor read,
// registry relocations, disposal-side anchor decisions — appends one verbose
// JSONL row carrying session_id, pid, cwd, and a timestamp to the project's
// .moai/logs/anchor-trace.jsonl.
//
// Fail-open: a trace write failure is logged once and swallowed. The trace is
// an observation surface; it must never wedge an anchor decision.
package session

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// AnchorTraceLogName is the trace log file name under the project root's
// .moai/logs directory.
const AnchorTraceLogName = "anchor-trace.jsonl"

// anchorTraceUnknownSession is the session_id marker on trace rows whose
// decision point has no session identity — the disposal-side anchor decision
// is made by the sweep process, not by an anchored session.
const anchorTraceUnknownSession = "unknown"

// AnchorTraceEnabled reads the MOAI_ANCHOR_TRACE gate (REQ-SAA-008). This is
// the single environment lookup every decision point pays when the switch is
// off; no other work happens on that path.
func AnchorTraceEnabled() bool {
	return envTruthy(os.Getenv(config.EnvAnchorTrace))
}

// envTruthy accepts "1" and "true" case-insensitively; everything else,
// including the empty string, is falsy.
func envTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true":
		return true
	default:
		return false
	}
}

// AnchorTraceEvent is one verbose trace row (REQ-SAA-007).
type AnchorTraceEvent struct {
	// Timestamp is the row's RFC3339 UTC occurrence time.
	Timestamp string `json:"timestamp"`
	// SessionID is the deciding session, or the "unknown" marker when the
	// decision point has no session identity.
	SessionID string `json:"session_id"`
	// PID is the process that made the decision.
	PID int `json:"pid"`
	// Cwd is the working directory the decision was about.
	Cwd string `json:"cwd"`
	// Decision names the decision point ("relocate", "anchor_decision",
	// "branch_guard.anchor_read").
	Decision string `json:"decision"`
	// Detail is the outcome qualifier.
	Detail string `json:"detail,omitempty"`
}

// AnchorTracePath returns the trace log path for a project root.
func AnchorTracePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "logs", AnchorTraceLogName)
}

// TraceAnchorDecision appends one trace row when the switch is on, and does
// nothing beyond the single gate lookup when it is off (REQ-SAA-008). An
// empty projectRoot drops the row silently — never guessed at.
//
// @MX:NOTE: [AUTO] SPEC-SESSION-ANCHOR-ATTR-001 — the W3 measurement
// instrument for the runtime-anchor incident class; call sites must stay
// fail-open and cost one env lookup when the gate is off.
func TraceAnchorDecision(projectRoot, decision, sessionID, cwd, detail string) {
	if !AnchorTraceEnabled() {
		return
	}
	if projectRoot == "" {
		return
	}
	if sessionID == "" {
		sessionID = anchorTraceUnknownSession
	}
	evt := AnchorTraceEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID: sessionID,
		PID:       os.Getpid(),
		Cwd:       cwd,
		Decision:  decision,
		Detail:    detail,
	}
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Warn("anchor trace: marshal failed", "error", err)
		return
	}
	data = append(data, '\n')
	path := AnchorTracePath(projectRoot)
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Warn("anchor trace: open failed", "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(data); err != nil {
		slog.Warn("anchor trace: write failed", "error", err)
	}
}
