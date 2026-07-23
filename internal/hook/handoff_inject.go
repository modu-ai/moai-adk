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
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
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
// @MX:REASON: registry에 3번째 SessionStart 핸들러로 등록 (deps.go). 중복 주입을 막는 불변식은 exclusive-create claim gate 단독 — 소비 경로에서 gate를 계속 보유하므로 rename이 loser에게 실패를 돌려주지 않아도(Windows sharing violation) 뒤 세션이 같은 레코드를 다시 이길 수 없다. gate 해제는 SavePending(다음 레코드 기록) 담당. claim-then-inject 순서(claim 성공 후 주입)는 유지 — 순서 반전 시 AC-013 회귀. rename 실패는 errno 무관 fail-open (Windows MoveFileEx). manual mode는 stale이어도 pure no-op (REQ-009 vs REQ-019 모순 해소). AskUserQuestion 미호출 (C-HRA-008).
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

	// Elect the winner BEFORE renaming. Exclusivity must come from a primitive
	// that guarantees it (O_CREATE|O_EXCL on a path both racers compute
	// identically), not from the rename failing for the loser: POSIX rename(2)
	// gives the loser ENOENT once the source is gone, but that is a property of
	// POSIX, not a documented cross-platform contract, and on Windows two
	// concurrent claims were observed to both succeed (AC-013a: injected=2 with
	// a single consumed file — one file object moved twice).
	if !claimPendingGate(projectDir) {
		return &HookOutput{}
	}

	// The gate is a DURABLE claim marker, not a mutex, so it is NOT released on
	// every path. It is dropped only on the two early returns below, where
	// nothing was consumed and the record is still live — holding it there would
	// strand a record no caller ever claimed. From the rename onwards the gate is
	// retained, because this caller may have consumed the record: retaining it is
	// the only thing that stops a later contender from winning the same record
	// when the rename reports success without the record actually moving (the
	// Windows fault above). Exclusivity therefore rests solely on the
	// exclusive-create claim, never on the rename failing for the loser and never
	// on the stat re-check below. The retained gate is cleared by
	// handoff.SavePending when the next record is written.

	// Re-check inside the gate. The record was read before the gate was taken,
	// so a contender that lost an earlier round still holds a stale "present"
	// observation. This narrows the read-then-consume window; it is a freshness
	// check, not the exclusivity mechanism.
	if _, err := os.Stat(handoff.PendingPath(projectDir)); err != nil {
		releasePendingGate(projectDir)
		return &HookOutput{}
	}

	// Consumed filename: <UnixNano ts>-<nonce8>.json. ts is the integer consume
	// timestamp (not the RFC3339 saved_at) so AC-014's `^\d+-` regex holds.
	ts := time.Now().UnixNano()
	consumedPath := filepath.Join(consumedDir, fmt.Sprintf("%d-%s.json", ts, consumeNonce(rec.SavedBySession)))

	// Archive the claimed record. This rename is no longer load-bearing for
	// exclusivity — the gate above already decided the winner — so it is a plain
	// move whose only job is preserving the pending bytes verbatim (AC-013b:
	// any rename error still skips injection, errno-agnostic). A rename error
	// consumed nothing and left the record live, so the gate is released here to
	// keep it claimable by a later session.
	if err := handoffRenameFunc(handoff.PendingPath(projectDir), consumedPath); err != nil {
		slog.Warn("session_start: handoff: claim rename failed; skipping injection (fail-open)",
			"error", err,
		)
		releasePendingGate(projectDir)
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

// claimPendingGate reports whether this caller won the right to consume
// pending.json. Exactly one concurrent caller gets true.
//
// The gate is never force-reclaimed here. Timing out and stealing it would
// reintroduce the very defect this replaced: two contenders that both decide a
// gate is stale each remove it and each then create it, so both win. A gate that
// outlives its consumer — retained by a caller that claimed the record, or
// leaked by a killed process — is instead cleared by handoff.SavePending when
// the next record is written: a record that has just been saved has by
// definition never been claimed, so clearing it there is unambiguous and
// race-free.
func claimPendingGate(projectDir string) bool {
	err := atomicfile.Claim(handoff.ClaimGatePath(projectDir), 0o600)
	if err == nil {
		return true
	}
	if !errors.Is(err, fs.ErrExist) {
		slog.Warn("session_start: handoff: claim gate unavailable; skipping injection (fail-open)",
			"error", err,
		)
	}
	return false
}

// releasePendingGate drops the gate. It is called ONLY from the paths that
// consumed nothing and left pending.json live (the in-gate stat re-check finding
// the record already gone, and a failed claim rename), so that the record stays
// claimable by a later session. A caller that reached the rename keeps the gate:
// see the release policy in claimAndInject.
func releasePendingGate(projectDir string) {
	if err := os.Remove(handoff.ClaimGatePath(projectDir)); err != nil && !os.IsNotExist(err) {
		slog.Warn("session_start: handoff: could not release claim gate", "error", err)
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
