package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4) D3: context-usage.json persistence.
//
// On every statusline render, builder.Build writes an authoritative snapshot of
// the raw context-window usage (+ handoff stage) to
// <projectDir>/.moai/state/context-usage.json. The Detection Heuristics doctrine
// (D4) reads this file FIRST and falls back to the byte / system-reminder
// heuristics only when it is absent, stale, or unparseable.
//
// The write is a pure function of context usage — it is NEVER gated by
// HandoffConfig (REQ-THRESHOLD-007), keeping the state-file authoritative
// regardless of auto-resume opt-in. It reuses model_cache.go's atomic
// temp+rename + silent-fail pattern (REQ-THRESHOLD-009); no new writer
// mechanism is introduced.

// contextUsageSchemaVersion is the on-disk schema version of context-usage.json.
const contextUsageSchemaVersion = 1

// contextUsageFreshWindow bounds how long an empty-session_id record stays valid
// for the freshness fallback (REQ-THRESHOLD-014). It is intentionally
// session-scoped (generous, not seconds) to absorb the write-if-changed
// throttle plateau, where captured_at is not refreshed while usage plateaus
// (design §D.3 caveat). A named constant keeps this reader threshold out of
// inline literals (§14 hardcoding-prevention).
const contextUsageFreshWindow = 12 * time.Hour

// templateSourceEmbedPath is the path-component marker for the moai-adk-go
// template embed source tree. The //go:embed all:templates directive in
// internal/template/embed.go compiles this tree (including dot-prefixed dirs
// like .moai/) into the distributed binary, so writeContextUsage MUST NOT
// write context-usage.json here — otherwise a runtime artifact leaks into the
// binary on make build. Built from filepath.Separator (not a bare "/" literal)
// so the guard is cross-platform (§14 hardcoding-prevention).
const templateSourceEmbedPath = "internal" + string(filepath.Separator) + "template" + string(filepath.Separator) + "templates"

// contextUsageRecord is the on-disk schema for
// <projectDir>/.moai/state/context-usage.json (REQ-THRESHOLD-010).
//
// writer_pid is the writing process identity — the discriminator that
// distinguishes concurrent empty-session_id writers sharing one file
// (REQ-THRESHOLD-018). It is deliberately excluded from the write-if-changed
// throttle payload (sameSemanticPayload) because it is render-ephemeral
// (statusline renders through a fresh process per invocation); including it
// would defeat the throttle by forcing a write on every render.
type contextUsageRecord struct {
	SchemaVersion     int     `json:"schema_version"`
	SessionID         string  `json:"session_id"`
	WriterPID         int     `json:"writer_pid"`
	CapturedAt        string  `json:"captured_at"`
	ContextWindowSize int     `json:"context_window_size"`
	TokensUsed        int     `json:"tokens_used"`
	RawPct            float64 `json:"raw_pct"`
	Stage             string  `json:"stage"`
	Band              string  `json:"band"`
}

// String renders the handoff stage as its on-disk / doctrine label.
func (s handoffStage) String() string {
	switch s {
	case handoffStageSoft:
		return "soft"
	case handoffStageHard:
		return "hard"
	default:
		return "none"
	}
}

// bandLabel classifies a raw context-window size into the on-disk band label.
// Uses the config large-window cutoff constant (no inline literal, §14).
func bandLabel(cwSize int) string {
	if cwSize >= config.HandoffLargeWindowCutoff {
		return "large"
	}
	return "standard"
}

// isTemplateSourceDir reports whether dir resolves into the moai-adk-go
// template embed source tree (internal/template/templates). The match is on a
// directory-component boundary, NOT a bare substring: .../internal/template/
// templates and .../internal/template/templates/.claude/hooks/moai match, but
// .../internalXtemplateXtemplates and .../internal/template/templates_bar do
// NOT. A user project dir never carries this path component, so the guard is
// inert for normal user projects. Empty dir returns false. Resolution to
// absolute via filepath.Abs falls back to dir on error (best-effort).
func isTemplateSourceDir(dir string) bool {
	if dir == "" {
		return false
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	abs = filepath.Clean(abs)
	sep := string(filepath.Separator)
	marker := sep + templateSourceEmbedPath
	// abs IS the templates dir (relative fallback), OR ends with the marker
	// (marker is the last path component), OR contains the marker bounded by a
	// trailing separator (marker is an interior component — a subdir of it).
	return abs == templateSourceEmbedPath ||
		strings.HasSuffix(abs, marker) ||
		strings.Contains(abs, marker+sep)
}

// writeContextUsage persists the current context-usage snapshot to
// <projDir>/.moai/state/context-usage.json using the atomic temp-file +
// rename pattern of WriteModelCache (MkdirAll + write-temp + rename). It is
// best-effort: any failure is silently ignored so the statusline render is
// never disrupted (REQ-THRESHOLD-009). The signature carries NO HandoffConfig —
// the write is unconditional with respect to Mode/Guide (REQ-THRESHOLD-007).
//
// It skips (no write, no panic) when the source signal is absent
// (mem.Available == false or non-positive window), the project dir cannot be
// resolved (projDir == ""), or projDir resolves into the template embed source
// tree (isTemplateSourceDir — prevents //go:embed all:templates from leaking a
// runtime artifact into the distributed binary).
func writeContextUsage(projDir, sessionID string, writerPID int, mem MemoryData, stage handoffStage) {
	if !mem.Available || mem.ContextWindowSize <= 0 || projDir == "" || isTemplateSourceDir(projDir) {
		return
	}

	stateDir := filepath.Join(projDir, ".moai", "state")
	path := filepath.Join(stateDir, "context-usage.json")

	next := buildContextUsageRecord(sessionID, writerPID, mem, stage)

	// Write-if-changed throttle (REQ-THRESHOLD-012): skip when the semantic
	// payload is byte-equal to the on-disk record, so render-rate invocations
	// do not churn the disk.
	if existing, err := readContextUsage(path); err == nil && existing != nil &&
		sameSemanticPayload(existing, next) {
		return
	}

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return // best-effort, silent
	}

	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath) //nolint:errcheck // cleanup best-effort
		return
	}
}

// buildContextUsageRecord assembles the on-disk record from the current usage
// snapshot. raw_pct is the raw context-window usage (tokens / window), NOT the
// auto-compact-scaled TokenBudget percentage.
func buildContextUsageRecord(sessionID string, writerPID int, mem MemoryData, stage handoffStage) *contextUsageRecord {
	rawPct := float64(mem.TokensUsed) * 100.0 / float64(mem.ContextWindowSize)
	return &contextUsageRecord{
		SchemaVersion:     contextUsageSchemaVersion,
		SessionID:         sessionID,
		WriterPID:         writerPID,
		CapturedAt:        time.Now().Format(time.RFC3339Nano),
		ContextWindowSize: mem.ContextWindowSize,
		TokensUsed:        mem.TokensUsed,
		RawPct:            rawPct,
		Stage:             stage.String(),
		Band:              bandLabel(mem.ContextWindowSize),
	}
}

// readContextUsage reads and parses context-usage.json. Returns (nil, err) on
// any failure (file missing, unparseable) so the caller falls back to a write
// (throttle) or to heuristics (reader).
func readContextUsage(path string) (*contextUsageRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec contextUsageRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// sameSemanticPayload reports whether two records carry the same throttle-
// relevant payload: session_id, stage, context_window_size, and the
// integer-rounded raw_pct. captured_at AND writer_pid are deliberately
// EXCLUDED — captured_at is expected to change, and writer_pid is
// render-ephemeral (including it would defeat the throttle, design §D.3).
func sameSemanticPayload(a, b *contextUsageRecord) bool {
	return a.SessionID == b.SessionID &&
		a.Stage == b.Stage &&
		a.ContextWindowSize == b.ContextWindowSize &&
		int(a.RawPct) == int(b.RawPct)
}

// isRealSessionID reports whether s is a real session UUID (as opposed to an
// empty value or an environment-fallback sentinel carrying no real UUID). The
// Claude Code stdin session_id is either a compact UUID token or empty; the
// environment-fallback sentinel is a prose "not-available" string containing
// whitespace, so any whitespace-bearing value is treated as "no real UUID"
// per REQ-THRESHOLD-014.
func isRealSessionID(s string) bool {
	if s == "" {
		return false
	}
	return !strings.ContainsAny(s, " \t")
}

// isFreshForSession reports whether the on-disk record is the CURRENT session's
// own authoritative snapshot (i.e., a reader may trust it rather than falling
// back to heuristics). It encodes the session_id guard (REQ-THRESHOLD-013), the
// fallback-UUID freshness path (REQ-THRESHOLD-014), and the empty-id writer_pid
// discriminator (REQ-THRESHOLD-018):
//
//   - real UUID on either side → valid iff BOTH sides carry the same UUID
//     (last-writer-wins guard; writer_pid is not consulted). A mixed
//     UUID-vs-empty pair is conservatively stale.
//   - empty session_id on BOTH sides → valid iff captured_at is within the
//     freshness window AND rec.writer_pid == curWriterID. The writer_pid match
//     stops a concurrent same-checkout empty-id session (session B) from reading
//     another session's (session A) fresh snapshot as its own.
func isFreshForSession(rec *contextUsageRecord, curSession string, curWriterID int) bool {
	if rec == nil {
		return false
	}
	recReal := isRealSessionID(rec.SessionID)
	curReal := isRealSessionID(curSession)
	if recReal || curReal {
		// UUID path: at least one side has a real UUID — both must match.
		return recReal && curReal && rec.SessionID == curSession
	}
	// Both sides lack a real UUID: freshness AND writer_pid discriminator.
	if rec.WriterPID != curWriterID {
		return false
	}
	return contextUsageFresh(rec.CapturedAt)
}

// contextUsageFresh reports whether captured_at is within the freshness window.
// An unparseable timestamp is treated as not fresh (conservative → heuristics).
func contextUsageFresh(capturedAt string) bool {
	t, err := time.Parse(time.RFC3339Nano, capturedAt)
	if err != nil {
		return false
	}
	return time.Since(t) < contextUsageFreshWindow
}

// resolveProjectDir resolves the project directory that anchors
// <projDir>/.moai/state/context-usage.json, following the stdin workspace
// chain and falling back to the process CWD (design §D.2). Returns "" only when
// no directory can be resolved (write is then skipped).
func resolveProjectDir(input *StdinData) string {
	if input != nil {
		if input.Workspace != nil && input.Workspace.CurrentDir != "" {
			return input.Workspace.CurrentDir
		}
		if input.CWD != "" {
			return input.CWD
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}
