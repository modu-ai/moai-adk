package hook

// handoff_inject.go implements the SessionStart consumer half of the reverse
// auto-resume handoff (SPEC-HANDOFF-AUTORESUME-001 M3). It is the 3rd registered
// SessionStart handler; the registry accumulate-all merge keeps its
// additionalContext alongside the sessionStartHandler / autoUpdateHandler outputs.
//
// The single INJECT+CONSUME cell is (source == "clear" ∧ mode == "auto" ∧ a live
// pending record). Every other (source, mode) combination preserves the row.
// All paths are best-effort fail-open: no path blocks the session and none invokes
// AskUserQuestion (subagent boundary, REQ-AUTORESUME-016).

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// handoffInjectHandler injects a saved handoff into SessionStart additionalContext
// on /clear when handoff.mode == "auto".
type handoffInjectHandler struct {
	cfg ConfigProvider
}

// NewHandoffInjectHandler creates the SessionStart handoff-inject handler. It
// reads cfg.Handoff.Mode/Guide (sessionStartHandler ConfigProvider pattern).
func NewHandoffInjectHandler(cfg ConfigProvider) Handler {
	return &handoffInjectHandler{cfg: cfg}
}

// EventType returns EventSessionStart.
func (h *handoffInjectHandler) EventType() EventType { return EventSessionStart }

// @MX:ANCHOR: [AUTO] SessionStart auto-resume 주입 단일 진입점 (4-source × mode branch table). 유일 소비 셀 = source==clear ∧ mode==auto ∧ live pending. 나머지 7셀은 pending 보존.
// @MX:REASON: registry에 3번째 SessionStart 핸들러로 등록 (deps.go). 중복 주입은 factory.db의 status='pending' 조건부 UPDATE가 막는다. claim-then-inject 순서(claim 성공 후 주입)는 유지한다. DB 오류는 fail-open이다. manual mode는 stale이어도 pure no-op (REQ-009 vs REQ-019 모순 해소). AskUserQuestion 미호출 (C-HRA-008).
// @MX:SPEC: SPEC-HANDOFF-AUTORESUME-001
//
// Handle processes a SessionStart event. Best-effort: every path returns allow
// (an empty HookOutput or an additionalContext injection); no path blocks.
func (h *handoffInjectHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	mode, guide := h.handoffConfig()

	// manual (or unknown) mode → pure no-op; never touch pending, even when stale
	// (REQ-AUTORESUME-009). This preserves the unchanged baseline UX.
	if mode != "auto" {
		return &HookOutput{}, nil
	}

	projectDir := resolveProjectDir(input)
	if projectDir == "" {
		return &HookOutput{}, nil
	}

	rec, present, err := handoff.ReadPending(projectDir)
	if err != nil {
		// Present but corrupt → warn + preserve (REQ-AUTORESUME-017). No rename.
		slog.Warn("session_start: handoff: pending record unreadable; preserving for inspection",
			"error", err,
		)
		return &HookOutput{}, nil
	}
	if !present {
		// Absent pending → silent no-op (edge case: no slog).
		return &HookOutput{}, nil
	}

	// Auto-mode stale TTL cleanup (REQ-AUTORESUME-019). Precedence N1: stale-cleanup
	// takes precedence over the notice-only hint — a stale record is removed and no
	// hint is emitted, regardless of source or guide (nothing live to resume).
	if handoffStale(rec, time.Now()) {
		if rmErr := handoff.ExpirePending(projectDir, rec.ID); rmErr != nil {
			slog.Warn("session_start: handoff: stale pending cleanup failed", "error", rmErr)
		}
		return &HookOutput{}, nil
	}

	// Non-clear source (startup / resume / compact) → notice-only. Never consume.
	// A stderr hint is emitted only when guide == true (REQ-AUTORESUME-010).
	if input.Source != "clear" {
		if guide {
			_, _ = fmt.Fprint(os.Stderr,
				"moai handoff: an auto-resume record is waiting; enter /clear to inject it, "+
					"or run `moai handoff clear` to discard it.\n")
		}
		return &HookOutput{}, nil
	}

	// The single INJECT+CONSUME cell: source == clear ∧ mode == auto ∧ live pending.
	newSessionID := ""
	if input != nil {
		newSessionID = input.SessionID
	}
	return h.claimAndInject(projectDir, newSessionID, rec), nil
}

// claimAndInject performs the claim-then-inject sequence using a transactional
// SQLite compare-and-swap, then renders and injects the additionalContext.
//
// newSessionID is the POST-/clear session id under which an embedded goal is
// re-armed (SPEC-INFINITE-GOAL-001 REQ-6, Option A session-id keying; D8: the
// embedded Ceiling is re-validated before writing).
func (h *handoffInjectHandler) claimAndInject(projectDir, newSessionID string, rec *handoff.PendingRecord) *HookOutput {
	_ = rec // the database claim below re-reads the winning row transactionally.
	token := fmt.Sprintf("%d-%s", time.Now().UnixNano(), consumeNonce(newSessionID))
	claimed, claimID, present, err := handoff.ClaimPending(projectDir, token)
	if err != nil || !present {
		if err != nil {
			slog.Warn("session_start: handoff: transactional claim failed; skipping injection (fail-open)", "error", err)
		}
		return &HookOutput{}
	}

	// Claim succeeded → render and inject (claim-then-inject order).
	out := &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventSessionStart),
			AdditionalContext: renderHandoffContext(claimed),
		},
	}

	// SPEC-INFINITE-GOAL-001 REQ-6 (Option A session-id keying): when the consumed
	// record carries an embedded goal, re-arm it under the NEW session-id. This is
	// a NET-NEW goal-state write surface (distinct from the AdditionalContext write
	// above — a separate file artifact). Best-effort + fail-open: a write error is
	// logged and never blocks the session.
	rearmEmbeddedGoal(projectDir, newSessionID, claimed)
	if err := handoff.FinishClaim(projectDir, claimID, true, "injected", token); err != nil {
		slog.Warn("session_start: handoff: injection succeeded but audit transition failed", "error", err)
	}

	return out
}

// rearmEmbeddedGoal writes a new goal state file under newSessionID when rec
// carries an embedded goal. D8 defense-in-depth: an unbounded embedded record
// (MaxTurns=0 with no real bound) is REJECTED — a corrupt pending.json must not
// re-open the unbounded hole. The new goal carries the embedded Ceiling and an
// armed status so the new session's stop-goal evaluator picks it up. Session-id
// keying (NOT SPEC-id) is preserved per Option A (Option B rejected: multi-
// session race).
func rearmEmbeddedGoal(projectDir, newSessionID string, rec *handoff.PendingRecord) {
	if rec == nil || rec.EmbeddedGoal == nil {
		return // nothing to re-arm
	}
	if newSessionID == "" {
		slog.Warn("session_start: handoff: embedded goal present but new session-id is empty; skipping rearm")
		return
	}
	eg := rec.EmbeddedGoal
	// D8 re-validation: reject an unbounded infinite arm.
	if eg.IsUnbounded() {
		slog.Warn("session_start: handoff: embedded goal is unbounded (MaxTurns=0 with no real bound); rejecting rearm (D8 defense-in-depth)",
			"condition", eg.Condition)
		return
	}
	// Reconstruct the goal's condition set from the embedded condition text. The
	// condition is stored verbatim; the parse (mechanical vs model) is the new
	// session's evaluator's job, so store a single mechanical condition carrying
	// the raw text as a best-effort reconstruction. The arm verb's parseCondition
	// is the canonical parser; here we mirror its mechanical-default shape.
	g := &goal.Goal{
		SessionID: newSessionID,
		Goal:      eg.Condition,
		Conditions: []goal.Condition{
			{Type: goal.ConditionMechanical, Cmd: eg.Condition, ExpectExit: 0},
		},
		Ceiling: goal.Ceiling{
			MaxTurns:    eg.MaxTurns,
			MaxDuration: eg.MaxDuration,
			CostCap:     eg.CostCap,
		},
		TurnsUsed:       0,
		Progress:        nil,
		ProgressionMode: goal.DefaultProgressionMode,
		CreatedAt:       time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
		Status:          goal.StatusArmed,
	}
	if err := goal.SaveGoal(projectDir, g); err != nil {
		slog.Warn("session_start: handoff: embedded goal rearm write failed; skipping (fail-open)",
			"error", err, "session_id", newSessionID)
	}
}

// handoffConfig resolves (mode, guide) from the ConfigProvider, defaulting to
// manual/no-guide when config is unavailable, and coercing an unknown mode value
// to manual (edge case: typo → safe no-op, D.1).
func (h *handoffInjectHandler) handoffConfig() (mode string, guide bool) {
	if h.cfg == nil {
		return config.DefaultHandoffMode, false
	}
	c := h.cfg.Get()
	if c == nil {
		return config.DefaultHandoffMode, false
	}
	mode = c.Handoff.Mode
	if mode != "auto" && mode != "manual" {
		slog.Warn("session_start: handoff: unknown mode; treating as manual", "mode", mode)
		mode = "manual"
	}
	return mode, c.Handoff.Guide
}

// handoffStale reports whether rec is past the auto-mode stale TTL. A record with
// no saved_at is treated as not-stale (conservative — never discard on a missing
// timestamp).
func handoffStale(rec *handoff.PendingRecord, now time.Time) bool {
	if rec.SavedAt.IsZero() {
		return false
	}
	return now.Sub(rec.SavedAt) > config.DefaultHandoffStaleTTL
}

// consumeNonce returns an 8-hex nonce for the consumed filename (REQ-AUTORESUME-014):
//   - a non-empty saved_by_session whose first 8 chars are clean hex → those chars
//     (attribution preserved);
//   - otherwise a crypto/rand 8-hex value;
//   - if crypto/rand fails (extremely rare) → a deterministic UnixNano low-32-bit
//     hex fallback.
//
// Cross-session collision is unreachable by the atomic-rename-as-claim argument
// (design.md §C.4): at most one session ever creates a consumed file, so the nonce
// only needs within-session uniqueness + a human-readable audit identifier.
func consumeNonce(session string) string {
	if len(session) >= 8 {
		if cand := strings.ToLower(session[:8]); isHex8(cand) {
			return cand
		}
	}
	var b [4]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("%08x", uint32(time.Now().UnixNano()))
}

// isHex8 reports whether s is exactly 8 lowercase hex digits.
func isHex8(s string) bool {
	if len(s) != 8 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
