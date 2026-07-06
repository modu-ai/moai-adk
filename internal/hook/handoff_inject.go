package hook

// handoff_inject.go implements the SessionStart consumer half of the reverse
// auto-resume handoff (SPEC-HANDOFF-AUTORESUME-001 M3). It is the 3rd registered
// SessionStart handler; the registry accumulate-all merge keeps its
// additionalContext alongside the sessionStartHandler / autoUpdateHandler outputs.
//
// The single INJECT+CONSUME cell is (source == "clear" ∧ mode == "auto" ∧ a live
// pending record). Every other (source, mode) combination preserves pending.json.
// All paths are best-effort fail-open: no path blocks the session and none invokes
// AskUserQuestion (subagent boundary, REQ-AUTORESUME-016).

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// handoffRenameFunc is the claim-rename seam. Production uses os.Rename; tests
// override it to force an arbitrary errno (AC-AUTORESUME-013 cross-platform
// fail-open — the handler must skip injection on ANY rename error, not only
// os.IsNotExist, for Windows MoveFileEx compatibility).
var handoffRenameFunc = os.Rename

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
// @MX:REASON: registry에 3번째 SessionStart 핸들러로 등록 (deps.go). claim-then-inject 순서(rename 성공 후 주입)가 2세션 race에서 중복 주입을 방지하는 불변식 — 순서 반전 시 AC-013 회귀. rename 실패는 errno 무관 fail-open (Windows MoveFileEx). manual mode는 stale이어도 pure no-op (REQ-009 vs REQ-019 모순 해소). AskUserQuestion 미호출 (C-HRA-008).
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
		if rmErr := os.Remove(handoff.PendingPath(projectDir)); rmErr != nil && !os.IsNotExist(rmErr) {
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
	return h.claimAndInject(projectDir, rec), nil
}

// claimAndInject performs the claim-then-inject sequence: atomically rename
// pending.json → consumed/<ts>-<nonce>.json, then (only on rename success) render
// and inject the additionalContext (REQ-AUTORESUME-012). Any rename failure skips
// injection and returns allow (REQ-AUTORESUME-013 fail-open).
func (h *handoffInjectHandler) claimAndInject(projectDir string, rec *handoff.PendingRecord) *HookOutput {
	consumedDir := handoff.ConsumedDir(projectDir)
	if err := os.MkdirAll(consumedDir, 0o700); err != nil {
		slog.Warn("session_start: handoff: cannot create consumed dir; skipping injection",
			"error", err,
		)
		return &HookOutput{}
	}

	// Consumed filename: <UnixNano ts>-<nonce8>.json. ts is the integer consume
	// timestamp (not the RFC3339 saved_at) so AC-014's `^\d+-` regex holds.
	ts := time.Now().UnixNano()
	consumedPath := filepath.Join(consumedDir, fmt.Sprintf("%d-%s.json", ts, consumeNonce(rec.SavedBySession)))

	// Claim via atomic rename. The rename-as-claim guarantee means at most one
	// concurrent session wins; the loser observes a rename error (errno-agnostic).
	if err := handoffRenameFunc(handoff.PendingPath(projectDir), consumedPath); err != nil {
		slog.Warn("session_start: handoff: claim rename failed; skipping injection (fail-open)",
			"error", err,
		)
		return &HookOutput{}
	}

	// Rename succeeded → render and inject (claim-then-inject order).
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventSessionStart),
			AdditionalContext: renderHandoffContext(rec),
		},
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
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
