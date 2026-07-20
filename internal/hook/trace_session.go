// SPEC-HOOK-FAILURE-CLASSIFY-001 REQ-HFC-004: trace session_id resolution.
// PostToolUse / PostToolUseFailure payloads may omit session_id (Claude Code
// bug #541), in which case validateInput substitutes "unknown" and every
// session-id-less event collides into a shared trace-unknown.jsonl. The
// transcript_path field, however, is named after the session UUID
// (~/.claude/projects/<hash>/<session-uuid>.jsonl), so a resolvable session
// can be recovered from it. Fallback chain (documented last resort):
// explicit session_id → transcript_path UUID → "unknown" (unchanged).
package hook

import (
	"path/filepath"
	"regexp"
	"strings"
)

// sessionUUIDPattern matches a canonical RFC 4122 UUID string, the filename
// stem Claude Code uses for session transcripts.
var sessionUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// resolveTraceSessionID returns the session identifier to bind into the
// trace filename (trace-<sessionID>.jsonl). A present, non-placeholder
// session_id always wins (no regression for well-formed payloads); the
// "unknown" placeholder injected by validateInput and an empty value both
// trigger transcript_path derivation. When nothing resolves, the original
// value is kept so the documented "unknown" last-resort fallback applies.
func resolveTraceSessionID(input *HookInput) string {
	if input == nil {
		return ""
	}
	if input.SessionID != "" && input.SessionID != "unknown" {
		return input.SessionID
	}
	if uuid := sessionUUIDFromTranscriptPath(input.TranscriptPath); uuid != "" {
		return uuid
	}
	return input.SessionID
}

// sessionUUIDFromTranscriptPath extracts the session UUID from a transcript
// path of the form .../<session-uuid>.jsonl. Returns "" when the path is
// empty or its base name is not a UUID (fail-open — never guesses).
func sessionUUIDFromTranscriptPath(transcriptPath string) string {
	if transcriptPath == "" {
		return ""
	}
	base := strings.TrimSuffix(filepath.Base(transcriptPath), ".jsonl")
	if sessionUUIDPattern.MatchString(base) {
		return base
	}
	return ""
}
