// Package preference — M5 session-level personalization toggle (toggle.go).
//
// This file implements REQ-ADM-013 / AC-ADM-013 [S2 Critical] + NFR-ADM-005 /
// AC-ADM-NFR-005 [S3 Major]: the session-level personalization recovery toggle.
//
// Design (design.md §A.5): the toggle state is stored as a sentinel file at
//
//	<projectRoot>/.moai/state/session-preference-disabled
//
// While the sentinel exists, the orchestrator suppresses BOTH:
//   - preference-based recommendation placement (the `(권장)` label override),
//   - uncertainty-based question-skipping (auto-handling of p≈0/1 decisions).
//
// The toggle is SESSION-SCOPED, NOT permanent:
//   - DisablePersonalization creates the sentinel.
//   - EnablePersonalization removes it (explicit re-enable).
//   - CleanupStaleSentinel removes it (called by the SessionStart hook branch
//     so a NEW session auto-reactivates personalization — NFR-ADM-005).
//
// The sentinel lives under .moai/state/ and MUST NOT leak into .moai/config/
// (NFR-ADM-005 evidence: `grep -r session-preference-disabled .moai/config/`
// must return 0 matches). This is the Loughrey autonomy buffer: the user can
// suppress personalization for one session without permanently disabling it.
//
// The CLI subcommand `moai preference toggle` is the user-facing entry point.
// It is advisory / non-blocking (constraint #7): sentinel write failures emit
// a stderr warning + non-zero exit but do NOT block workflows.

package preference

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// sessionDisabledSentinelName is the sentinel-file name (design.md §A.5). Its
// presence under <projectRoot>/.moai/state/ signals "personalization disabled
// for this session".
//
// @MX:NOTE: [AUTO] sentinel 이름 — design.md §A.5 정준 경로 (.moai/state/session-preference-disabled)
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-013, NFR-ADM-005, design.md §A.5
const sessionDisabledSentinelName = "session-preference-disabled"

// sentinelStateSubdir is the state subdirectory under the project root where
// the sentinel lives. It matches the M4 decay-cadence stamp location
// (.moai/state/) so both runtime artifacts share one state dir.
const sentinelStateSubdir = ".moai/state"

// sentinelPath returns the absolute path to the sentinel for the given project
// root. Pure function — no I/O.
func sentinelPath(projectRoot string) string {
	return filepath.Join(projectRoot, sentinelStateSubdir, sessionDisabledSentinelName)
}

// IsPersonalizationDisabled reports whether the sentinel exists for the given
// project root (REQ-ADM-013, AC-ADM-013). True → suppress preference-based
// recommendation placement + uncertainty-based question-skipping for this
// session.
//
// The check is advisory: an unreadable sentinel (permission denied) is treated
// as NOT disabled (personalization active) — fail-open to avoid stalling the
// orchestrator on an unreadable state file. A missing sentinel is the default
// active state.
//
// @MX:NOTE: [AUTO] IsPersonalizationDisabled — 세션 단위 토글 게이트 (sentinel 존재 = 비활성)
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-013, AC-ADM-013
func IsPersonalizationDisabled(projectRoot string) bool {
	if projectRoot == "" {
		return false
	}
	_, err := os.Stat(sentinelPath(projectRoot))
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	// Unreadable / permission denied → fail-open (treat as active). A broken
	// sentinel read should NOT silently disable personalization.
	return false
}

// DisablePersonalization creates the sentinel file, flipping
// IsPersonalizationDisabled to true for the session (REQ-ADM-013). The sentinel
// records the disable timestamp for audit. Idempotent: if the sentinel already
// exists, it is overwritten with the fresh timestamp (no error).
//
// `now` is injected (not time.Now()) so the audit record is reproducible in
// tests.
func DisablePersonalization(projectRoot string, now time.Time) error {
	if projectRoot == "" {
		return fmt.Errorf("preference: DisablePersonalization requires non-empty projectRoot")
	}
	stateDir := filepath.Join(projectRoot, sentinelStateSubdir)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("preference: DisablePersonalization mkdir state dir: %w", err)
	}
	content := fmt.Sprintf("disabled_at=%s\n", now.UTC().Format(time.RFC3339))
	if err := os.WriteFile(sentinelPath(projectRoot), []byte(content), 0o644); err != nil {
		return fmt.Errorf("preference: DisablePersonalization write sentinel: %w", err)
	}
	return nil
}

// EnablePersonalization removes the sentinel file, reactivating personalization
// (REQ-ADM-013). Idempotent: a missing sentinel is NOT an error — it is the
// default active state. Returns nil if the sentinel was absent or was removed
// cleanly.
func EnablePersonalization(projectRoot string) error {
	if projectRoot == "" {
		return fmt.Errorf("preference: EnablePersonalization requires non-empty projectRoot")
	}
	p := sentinelPath(projectRoot)
	err := os.Remove(p)
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		// Idempotent — already active.
		return nil
	}
	return fmt.Errorf("preference: EnablePersonalization remove sentinel: %w", err)
}

// CleanupStaleSentinel is the SessionStart-hook helper that removes a stale
// sentinel so a NEW session auto-reactivates personalization (NFR-ADM-005,
// AC-ADM-NFR-005: "신규 세션에서 자동 재활성화"). It is idempotent: a missing
// sentinel is a no-op (the common SessionStart path — most sessions do not
// have a stale sentinel).
//
// M5 delivers this helper. Wiring it into the SessionStart hook branch is
// acceptable if minimal; the helper existing is the M5 deliverable (per spawn
// prompt constraint #1). The hook caller invokes it at session start.
//
// @MX:NOTE: [AUTO] CleanupStaleSentinel — SessionStart 훅이 호출 (신규 세션 자동 재활성화)
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 NFR-ADM-005, AC-ADM-NFR-005
func CleanupStaleSentinel(projectRoot string) error {
	if projectRoot == "" {
		return nil // nothing to clean — no-op
	}
	p := sentinelPath(projectRoot)
	err := os.Remove(p)
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil // already clean — idempotent
	}
	return fmt.Errorf("preference: CleanupStaleSentinel remove: %w", err)
}

// ---- M4 toggle TOCTOU hardening (REQ-CONC-001-005 / AC-005) ----
//
// The prior runToggle did a read-then-flip (IsPersonalizationDisabled →
// Enable/Disable) with no serialization. Two concurrent toggles could both
// read "active" and both disable → net disabled instead of the correct "back
// to active" (lost flip). TogglePersonalization closes the TOCTOU by
// serializing the read-then-flip behind an O_EXCL sidecar lock.
//
// Lock primitive choice (SPEC option (a) "stdlib flock-equivalent directly in
// preference"): an O_EXCL sidecar lock file (os.OpenFile with O_CREATE|O_EXCL)
// is portable stdlib — no syscall.Flock (which would need //go:build build
// tags), no import of internal/cli's lockFile/unlockFile (which would create a
// cli→preference←cli import cycle), and no new build-tagged files. O_EXCL
// creation is atomic at the kernel level on local filesystems (the
// .moai/state/ case). A stale-lock timeout (toggleLockStaleTimeout) recovers
// if a holder crashes mid-flip: the next toggle removes a lock older than the
// timeout and retries.

const (
	// toggleLockStaleTimeout is the age beyond which an O_EXCL lock file is
	// considered abandoned (holder crashed). A real toggle completes in
	// milliseconds, so 10s is a very conservative cutoff.
	toggleLockStaleTimeout = 10 * time.Second
	// toggleLockBackoff is the retry interval when another toggle holds the
	// lock. Toggles are rare (human CLI), so 5ms is a light spin.
	toggleLockBackoff = 5 * time.Millisecond
	// toggleLockMaxAttempts bounds total wait (~2s at 5ms backoff) so a stuck
	// lock surfaces a diagnostic rather than hanging the CLI forever.
	toggleLockMaxAttempts = 400
)

// acquireToggleLock obtains an exclusive O_EXCL sidecar lock for the toggle
// read-then-flip critical section. Returns a release func the caller MUST
// defer. Stale locks (older than toggleLockStaleTimeout) are removed
// best-effort so a crashed holder does not permanently block toggles.
func acquireToggleLock(projectRoot string) (release func(), err error) {
	if err := os.MkdirAll(filepath.Join(projectRoot, sentinelStateSubdir), 0o755); err != nil {
		return nil, fmt.Errorf("preference: toggle lock mkdir: %w", err)
	}
	lockPath := sentinelPath(projectRoot) + ".toggle-lock"
	for attempt := 0; attempt < toggleLockMaxAttempts; attempt++ {
		// Stale-lock cleanup: if the lock is older than the timeout, a prior
		// holder crashed. Remove it best-effort; the O_EXCL below sorts out
		// any concurrent remover race (only one O_EXCL creator wins).
		if fi, statErr := os.Stat(lockPath); statErr == nil {
			if time.Since(fi.ModTime()) > toggleLockStaleTimeout {
				_ = os.Remove(lockPath)
			}
		}
		f, oErr := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if oErr == nil {
			_ = f.Close()
			return func() { _ = os.Remove(lockPath) }, nil
		}
		// ErrExist → another toggle holds the lock; retry.
		//
		// ErrPermission is ALSO retryable, because of Windows delete semantics:
		// os.Remove only marks a file for deletion, and the name stays in a
		// "delete pending" state until the last handle closes. Creating the same
		// name in that window fails with ERROR_ACCESS_DENIED, not ERROR_EXISTS —
		// so the previous holder's release() can make the next acquirer fail
		// outright. Treating it as transient lets the backoff ride out the
		// window; a genuinely unwritable lock path still surfaces an error after
		// the attempt budget rather than being swallowed.
		if !errors.Is(oErr, os.ErrExist) && !errors.Is(oErr, os.ErrPermission) {
			return nil, fmt.Errorf("preference: acquire toggle lock: %w", oErr)
		}
		time.Sleep(toggleLockBackoff)
	}
	return nil, fmt.Errorf("preference: toggle lock busy after %d attempts (remove %s if stale)", toggleLockMaxAttempts, lockPath)
}

// TogglePersonalization atomically flips the session-personalization sentinel
// state under an exclusive O_EXCL lock so concurrent toggles compose correctly
// (REQ-CONC-001-005 / AC-005). Returns the NEW disabled state (true = sentinel
// present = personalization disabled for the session).
//
// The lock serializes the read-then-flip: each toggle observes the effect of
// the previous one, so N concurrent toggles from a known start state produce
// the correct parity (even N → original state, odd N → flipped).
//
// @MX:NOTE: [AUTO] TogglePersonalization — locked read-then-flip (TOCTOU hardening)
// @MX:SPEC: SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-005, AC-CONC-001-005
func TogglePersonalization(projectRoot string, now time.Time) (bool, error) {
	if projectRoot == "" {
		return false, fmt.Errorf("preference: TogglePersonalization requires non-empty projectRoot")
	}
	release, err := acquireToggleLock(projectRoot)
	if err != nil {
		return false, err
	}
	defer release()

	wasDisabled := IsPersonalizationDisabled(projectRoot)
	if wasDisabled {
		if err := EnablePersonalization(projectRoot); err != nil {
			return false, err
		}
		return false, nil // now enabled
	}
	if err := DisablePersonalization(projectRoot, now); err != nil {
		return false, err
	}
	return true, nil // now disabled
}

// ---- CLI subcommand (newToggleCmd) ----

// toggleFlags holds the flags for the toggle subcommand. Value type so tests
// can construct an isolated cmd + flags pair without polluting the package-level
// default cmd (mirrors decayScanFlags in cmd.go).
type toggleFlags struct {
	memoryDir   string // --memory-dir override (empty = auto-resolve; accepted for parity, unused by the sentinel path)
	projectRoot string // --project-root override (empty = resolve from $CLAUDE_PROJECT_DIR)
	json        bool   // --json machine-readable output
}

// newToggleCmd builds the toggle subcommand. Factory form keeps tests isolated
// from the package-level PreferenceCmd.
func newToggleCmd() *cobra.Command {
	flags := &toggleFlags{}
	cmd := &cobra.Command{
		Use:   "toggle",
		Short: "Toggle session-level personalization on/off (non-permanent, resets each session)",
		Long: `Toggle session-level personalization.

While personalization is disabled for the current session, the orchestrator
suppresses both preference-based recommendation placement and uncertainty-based
question-skipping. The toggle is SESSION-SCOPED, NOT permanent: a new session
auto-reactivates personalization (the sentinel is cleaned up at session start).

The sentinel lives at <projectRoot>/.moai/state/session-preference-disabled
and MUST NOT leak into .moai/config/ (NFR-ADM-005).

Each invocation flips the state (disabled ↔ enabled). Exit codes: 0 success,
1 user error, 2 system error. The toggle is advisory — sentinel write
failures emit a stderr warning but do NOT block workflows.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToggle(cmd.OutOrStdout(), cmd.ErrOrStderr(), flags)
		},
	}
	cmd.Flags().StringVar(&flags.memoryDir, "memory-dir", "", "preference memory dir override (default: $CLAUDE_PROJECT_DIR-derived; accepted for parity with decay-scan, unused by the sentinel path)")
	cmd.Flags().StringVar(&flags.projectRoot, "project-root", "", "project root override (default: $CLAUDE_PROJECT_DIR)")
	cmd.Flags().BoolVar(&flags.json, "json", false, "emit the toggle result as JSON (stdout)")
	return cmd
}

// runToggle is the command body, extracted for testability. It resolves the
// project root, flips the sentinel state via the locked TogglePersonalization
// seam (REQ-CONC-001-005 — concurrent toggles compose correctly), and emits a
// machine-readable summary (JSON with --json, plain text otherwise).
func runToggle(stdout, stderr io.Writer, flags *toggleFlags) error {
	root, err := resolveToggleProjectRoot(flags.projectRoot)
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	nowDisabled, err := TogglePersonalization(root, now)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warn: could not toggle personalization: %v\n", err)
		return err
	}

	if flags.json {
		out := map[string]any{
			"disabled":     nowDisabled,
			"project_root": root,
			"toggled_at":   now.Format(time.RFC3339),
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return fmt.Errorf("preference: toggle encode JSON: %w", err)
		}
		return nil
	}
	state := "enabled"
	if nowDisabled {
		state = "disabled"
	}
	_, _ = fmt.Fprintf(stdout, "personalization %s for session (project_root=%s)\n", state, root)
	return nil
}

// resolveToggleProjectRoot applies the same resolution priority as
// resolveProjectRoot (cmd.go): --project-root flag if set, then
// $CLAUDE_PROJECT_DIR, then os.Getwd() fallback. Reuses the existing helper to
// avoid duplicating the resolution logic.
func resolveToggleProjectRoot(flagValue string) (string, error) {
	if flagValue != "" {
		abs, err := filepath.Abs(flagValue)
		if err != nil {
			return "", fmt.Errorf("preference: resolve --project-root %q: %w", flagValue, err)
		}
		return abs, nil
	}
	return resolveProjectRoot()
}
