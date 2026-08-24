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
// <projectDir>/.moai/state/context-usage/<session-id>.json — one record per
// session (SPEC-SESSION-TELEMETRY-001 REQ-ST-001). The Detection Heuristics
// doctrine (D4) reads the current session's record FIRST and falls back to the
// byte / system-reminder heuristics only when it is absent or unparseable.
//
// The write is a pure function of context usage — it is NEVER gated by
// HandoffConfig (REQ-THRESHOLD-007), keeping the state-file authoritative
// regardless of auto-resume opt-in. It reuses model_cache.go's atomic
// temp+rename + silent-fail pattern (REQ-THRESHOLD-009); no new writer
// mechanism is introduced.

// contextUsageSchemaVersion is the on-disk schema version of a session
// telemetry record. Bumped to 2 by SPEC-SESSION-TELEMETRY-001: the payload
// gained the session's model and effort. A reader tolerates records at the
// previous version — the two fields are simply absent (REQ-ST-003).
const contextUsageSchemaVersion = 2

// templateSourceEmbedPath is the path-component marker for the moai-adk-go
// template embed source tree. The //go:embed all:templates directive in
// internal/template/embed.go compiles this tree (including dot-prefixed dirs
// like .moai/) into the distributed binary, so writeContextUsage MUST NOT
// write a telemetry record here — otherwise a runtime artifact leaks into the
// binary on make build. Built from filepath.Separator (not a bare "/" literal)
// so the guard is cross-platform (§14 hardcoding-prevention).
const templateSourceEmbedPath = "internal" + string(filepath.Separator) + "template" + string(filepath.Separator) + "templates"

// contextUsageDirName is the on-disk directory that holds the per-session
// telemetry records. Kept as a named constant so the writer, the reader, and
// the path helper share one spelling (§14 hardcoding-prevention).
const contextUsageDirName = "context-usage"

// SessionTelemetryPath returns the on-disk path of one session's telemetry
// record: <stateDir>/context-usage/<sessionID>.json. It is the one place the
// record's location is spelled, so a cross-package consumer never reconstructs
// it. stateDir is the project's .moai/state directory.
//
// The record is keyed by the identifier the session runtime delivered to the
// render (REQ-ST-002) — never by the project-wide
// .moai/state/current-session-id.txt sidecar, which carries the same
// last-writer-wins shape the per-session split exists to remove.
// The key is a path component arriving from outside the process, so a value
// that would resolve outside the per-session directory is REFUSED rather than
// sanitised: rewriting "../escape" into "escape" produces a file that looks
// legitimate and belongs to no session (REQ-ST-007). A refused key yields "",
// and the caller writes nothing while its render still completes.
func SessionTelemetryPath(stateDir, sessionID string) string {
	if !usableSessionKey(sessionID) {
		return ""
	}
	return filepath.Join(stateDir, contextUsageDirName, sessionID+".json")
}

// usableSessionKey reports whether sessionID is a single, ordinary filename
// component. Both separators are rejected on every platform: a backslash is a
// separator on Windows, and a key carrying one is hostile wherever it is read.
func usableSessionKey(sessionID string) bool {
	switch sessionID {
	case "", ".", "..":
		return false
	}
	if strings.ContainsAny(sessionID, `/\`) || strings.ContainsRune(sessionID, filepath.Separator) {
		return false
	}
	return !filepath.IsAbs(sessionID)
}

// SessionTelemetryRecord is the on-disk schema for one session's record at
// <projectDir>/.moai/state/context-usage/<session-id>.json (REQ-THRESHOLD-010,
// SPEC-SESSION-TELEMETRY-001 REQ-ST-001).
//
// writer_pid is the writing process identity — the discriminator that
// distinguishes concurrent empty-session_id writers sharing one file
// (REQ-THRESHOLD-018). It is deliberately excluded from the write-if-changed
// throttle payload (sameSemanticPayload) because it is render-ephemeral
// (statusline renders through a fresh process per invocation); including it
// would defeat the throttle by forcing a write on every render.
type SessionTelemetryRecord struct {
	SchemaVersion     int     `json:"schema_version"`
	SessionID         string  `json:"session_id"`
	WriterPID         int     `json:"writer_pid"`
	CapturedAt        string  `json:"captured_at"`
	ContextWindowSize int     `json:"context_window_size"`
	TokensUsed        int     `json:"tokens_used"`
	RawPct            float64 `json:"raw_pct"`
	Stage             string  `json:"stage"`
	Band              string  `json:"band"`

	// Model is the model this session actually runs — the backend-resolved
	// display name (D-5), so a GLM-backed session records the z.ai model
	// rather than the Claude display name the runtime supplied. Effort is the
	// session's effort level as the runtime delivered it.
	//
	// Both are omitted when the render payload did not supply them, or when
	// the record predates schema version 2. A reader presents an empty value
	// as "not recorded" and never infers a substitute (REQ-ST-003).
	Model  string `json:"model,omitempty"`
	Effort string `json:"effort,omitempty"`
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

// writeContextUsage persists the current session's telemetry snapshot to
// <projDir>/.moai/state/context-usage/<sessionID>.json using the atomic temp-file +
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
func writeContextUsage(projDir, sessionID string, writerPID int, mem MemoryData, stage handoffStage, model, effort string) {
	if !mem.Available || mem.ContextWindowSize <= 0 || projDir == "" || isTemplateSourceDir(projDir) {
		return
	}

	stateDir := filepath.Join(projDir, ".moai", "state")
	path := SessionTelemetryPath(stateDir, sessionID)
	if path == "" {
		return // key refused (REQ-ST-007); the render still completes
	}

	next := buildContextUsageRecord(sessionID, writerPID, mem, stage, model, effort)

	// Write-if-changed throttle (REQ-THRESHOLD-012): skip when the semantic
	// payload is byte-equal to the on-disk record, so render-rate invocations
	// do not churn the disk.
	if existing, err := ReadSessionTelemetry(path); err == nil && existing != nil &&
		sameSemanticPayload(existing, next) {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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
func buildContextUsageRecord(sessionID string, writerPID int, mem MemoryData, stage handoffStage, model, effort string) *SessionTelemetryRecord {
	rawPct := float64(mem.TokensUsed) * 100.0 / float64(mem.ContextWindowSize)
	return &SessionTelemetryRecord{
		SchemaVersion:     contextUsageSchemaVersion,
		SessionID:         sessionID,
		WriterPID:         writerPID,
		CapturedAt:        time.Now().Format(time.RFC3339Nano),
		ContextWindowSize: mem.ContextWindowSize,
		TokensUsed:        mem.TokensUsed,
		RawPct:            rawPct,
		Stage:             stage.String(),
		Band:              bandLabel(mem.ContextWindowSize),
		Model:             model,
		Effort:            effort,
	}
}

// ReadSessionTelemetry reads and parses one session telemetry record. Returns (nil, err) on
// any failure (file missing, unparseable) so the caller falls back to a write
// (throttle) or to heuristics (reader).
func ReadSessionTelemetry(path string) (*SessionTelemetryRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec SessionTelemetryRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// sameSemanticPayload reports whether two records carry the same throttle-
// relevant payload: session_id, stage, context_window_size, the
// integer-rounded raw_pct, and the session's model and effort. captured_at AND
// writer_pid are deliberately EXCLUDED — captured_at is expected to change, and
// writer_pid is render-ephemeral (including it would defeat the throttle,
// design §D.3).
//
// Model and effort are INCLUDED deliberately (plan.md §F item 1, measured):
// they change rarely within a session, so they do not defeat the throttle, and
// including them is what makes a mid-session model or effort switch reach disk
// on the very next render. Excluding them would leave the record holding a
// value that is present and wrong until an unrelated context value moved —
// a state a reader cannot distinguish from a current one, and which
// REQ-ST-003's "not recorded" path does not cover.
func sameSemanticPayload(a, b *SessionTelemetryRecord) bool {
	return a.SessionID == b.SessionID &&
		a.Stage == b.Stage &&
		a.ContextWindowSize == b.ContextWindowSize &&
		int(a.RawPct) == int(b.RawPct) &&
		a.Model == b.Model &&
		a.Effort == b.Effort
}

// resolveProjectDir resolves the project directory that anchors
// the per-session telemetry record, following the stdin workspace
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
