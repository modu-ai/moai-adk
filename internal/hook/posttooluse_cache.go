package hook

// posttooluse_cache.go — SPEC-V3R6-PROMPT-CACHE-001 M3.
//
// PostToolUse cache telemetry: extract Anthropic Prompt Caching token counts
// from the API response usage block, append a JSONL entry to
// .moai/state/cache-usage.jsonl, and detect single-turn sessions that incur a
// cache-write penalty (REQ-PC-007 / AC-PC-010).
//
// This handler is observation-only: a write failure is reported in the result
// but never blocks the user's Claude Code session. It calls NO AskUserQuestion
// (C-HRA-008 subagent boundary).
//
// REQ-PC-004: extract cache_creation_input_tokens + cache_read_input_tokens and
//             append JSONL with timestamp / session_id / turn / both counts.
// REQ-PC-007: single-turn session (turn==1 only AND wall-time < 5min) → warn
//             with a concrete `session_ttl: "off"` recommendation.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/modu-ai/moai-adk/internal/state"
)

// singleTurnPenaltyWindow is the wall-time threshold below which a turn-1-only
// session is flagged as a cache-write penalty risk (plan.md M3 §3: < 5 min).
const singleTurnPenaltyWindow = 5 * time.Minute

// penaltyWarningTemplate is the REQ-PC-007 warning. It contains BOTH the
// "single-turn cache write penalty risk" string (AC-PC-010 §2) and the
// `session_ttl: "off"` recommendation (AC-PC-010 §3).
const penaltyWarningTemplate = `single-turn cache write penalty risk: session %q completed in %s with only 1 turn — ` +
	`the 1h cache write (+100%%) was never amortized by a cache read. ` +
	`Recommendation: set session_ttl: "off" in .moai/config/sections/cache.yaml for this workflow.`

// CacheTokenUsage is the extracted, normalized cache telemetry for a single
// turn. It is the recorder's input; ExtractCacheTokenUsage produces the token
// fields from a raw API response, while Turn / Elapsed are supplied by the
// caller (the hook tracks turn index + session elapsed time).
type CacheTokenUsage struct {
	// SessionID is the Claude Code session identifier.
	SessionID string
	// Turn is the 1-based turn index within the session.
	Turn int
	// CacheCreation is usage.cache_creation_input_tokens.
	CacheCreation int
	// CacheRead is usage.cache_read_input_tokens.
	CacheRead int
	// Model is the model identifier from the API response.
	Model string
	// Elapsed is the session wall-time at this turn (used for single-turn detection).
	Elapsed time.Duration
}

// CacheRecordResult is the outcome of CacheUsageRecorder.Record.
type CacheRecordResult struct {
	// Err is non-nil only when the JSONL append failed (observation-only — the
	// caller logs but does not block).
	Err error
	// PenaltyWarning is true when the single-turn penalty heuristic fired.
	PenaltyWarning bool
	// WarningMessage is the formatted REQ-PC-007 warning when PenaltyWarning is
	// true; empty otherwise.
	WarningMessage string
}

// CacheUsageRecorder appends cache telemetry and surfaces the single-turn
// penalty warning. It is stateless beyond the per-call inputs, so it is safe to
// construct per invocation.
type CacheUsageRecorder struct{}

// NewCacheUsageRecorder constructs a CacheUsageRecorder.
func NewCacheUsageRecorder() *CacheUsageRecorder {
	return &CacheUsageRecorder{}
}

// Record appends a cache-usage JSONL entry under the resolved project root and
// evaluates the single-turn cache-write penalty.
//
// Project-root resolution follows the B7 convention: input.CWD first, then
// CLAUDE_PROJECT_DIR, then os.Getwd() — via resolveProjectRootFromInputOrEnv.
// input may be nil (the recorder then falls back to the env/cwd resolver).
//
// @MX:ANCHOR: [AUTO] CacheUsageRecorder.Record — sole PostToolUse cache telemetry entry point (append + penalty warning)
// @MX:REASON: fan_in >= 3 expected — the PostToolUse handler wiring, AC-PC-005/006 JSONL tests, and AC-PC-010 penalty-warning test all route through this single method; the JSONL append + single-turn detection (turn==1 AND elapsed<5min) are the contractual telemetry behaviors (K2/K3/K5 inputs) that feed the moai doctor hit-rate metric.
func (r *CacheUsageRecorder) Record(input *HookInput, usage CacheTokenUsage) CacheRecordResult {
	var res CacheRecordResult

	root := resolveProjectRootFromInputOrEnv(input, "cache-usage-recorder")
	entry := state.CacheUsageEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     usage.SessionID,
		Turn:          usage.Turn,
		CacheCreation: usage.CacheCreation,
		CacheRead:     usage.CacheRead,
		Model:         usage.Model,
	}

	if err := state.AppendCacheUsage(root, entry); err != nil {
		// Observation-only: record the error, never block the session.
		res.Err = err
		slog.Debug("cache telemetry append failed (observation-only)",
			"session_id", usage.SessionID,
			"turn", usage.Turn,
			"error", err,
		)
	}

	if isSingleTurnPenalty(usage) {
		res.PenaltyWarning = true
		res.WarningMessage = fmt.Sprintf(penaltyWarningTemplate, usage.SessionID, usage.Elapsed.Round(time.Second))
		slog.Warn("cache: "+res.WarningMessage,
			"session_id", usage.SessionID,
			"elapsed", usage.Elapsed.String(),
		)
	}

	return res
}

// isSingleTurnPenalty reports whether the usage represents a single-turn session
// that incurs the 1h cache-write penalty: turn index 1 AND elapsed wall-time
// below the penalty window. Turn indices > 1 never trigger (false-positive
// guard for the AC-PC-010 negative case).
func isSingleTurnPenalty(usage CacheTokenUsage) bool {
	if usage.Turn != 1 {
		return false
	}
	return usage.Elapsed > 0 && usage.Elapsed < singleTurnPenaltyWindow
}

// apiResponseUsage is the minimal shape of an Anthropic API response needed to
// extract cache token counts.
type apiResponseUsage struct {
	Model string `json:"model"`
	Turn  int    `json:"turn"`
	Usage *struct {
		CacheCreation *int `json:"cache_creation_input_tokens"`
		CacheRead     *int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

// logCacheUsage is the live PostToolUse integration: it extracts a cache usage
// block from the tool response (when present) and appends a telemetry entry.
//
// This path is best-effort and observation-only — a missing usage block, a
// malformed response, or a write failure is silently tolerated so the user's
// session is never blocked. The single-turn penalty warning is NOT raised here
// (the live hook has no reliable per-session wall-time signal); the warning is
// surfaced by the recorder's Record path and by `moai doctor` (M4 K5 metric).
//
// Turn index is read from the response when present, defaulting to 0 ("unknown")
// rather than 1 so the live path never spuriously classifies a session as
// single-turn.
func logCacheUsage(input *HookInput) {
	if input == nil || len(input.ToolResponse) == 0 {
		return
	}
	var parsed apiResponseUsage
	if err := json.Unmarshal(input.ToolResponse, &parsed); err != nil {
		return
	}
	if parsed.Usage == nil {
		return
	}

	usage := CacheTokenUsage{
		SessionID: input.SessionID,
		Turn:      parsed.Turn,
		Model:     parsed.Model,
	}
	if parsed.Usage.CacheCreation != nil {
		usage.CacheCreation = *parsed.Usage.CacheCreation
	}
	if parsed.Usage.CacheRead != nil {
		usage.CacheRead = *parsed.Usage.CacheRead
	}

	root := resolveProjectRootFromInputOrEnv(input, "log-cache-usage")
	entry := state.CacheUsageEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     usage.SessionID,
		Turn:          usage.Turn,
		CacheCreation: usage.CacheCreation,
		CacheRead:     usage.CacheRead,
		Model:         usage.Model,
	}
	if err := state.AppendCacheUsage(root, entry); err != nil {
		slog.Debug("cache usage telemetry append failed (observation-only)",
			"session_id", usage.SessionID, "error", err)
	}
}

// ExtractCacheTokenUsage parses an Anthropic API response body and extracts the
// cache token counts plus model. It returns ok=false when the body is not JSON
// or carries no usage block (graceful — the caller skips telemetry for that
// turn). The Turn and Elapsed fields are NOT populated here; the caller fills
// them from session state.
//
// Pointer fields distinguish "absent" (no usage block at all → ok=false) from
// "present and zero" (cache_read_input_tokens: 0 → ok=true, value 0).
func ExtractCacheTokenUsage(responseBody []byte) (CacheTokenUsage, bool) {
	var parsed apiResponseUsage
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return CacheTokenUsage{}, false
	}
	if parsed.Usage == nil {
		return CacheTokenUsage{}, false
	}
	usage := CacheTokenUsage{Model: parsed.Model}
	if parsed.Usage.CacheCreation != nil {
		usage.CacheCreation = *parsed.Usage.CacheCreation
	}
	if parsed.Usage.CacheRead != nil {
		usage.CacheRead = *parsed.Usage.CacheRead
	}
	return usage, true
}
