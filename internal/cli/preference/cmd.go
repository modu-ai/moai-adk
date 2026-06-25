// Package preference — CLI subcommand surface (cmd.go).
//
// This file wires the `moai preference` parent command and its two children:
//
//   - `moai preference decay-scan` (M4, REQ-ADM-011/012, NFR-ADM-004): the
//     daily background decay-policy pass (power-law weight refresh + 28-day
//     TTL eviction of transient entries).
//   - `moai preference toggle` (M5, REQ-ADM-013, NFR-ADM-005): the session-
//     level personalization recovery toggle (non-permanent; auto-reactivates
//     each new session).
//
// The CLI is a maintenance entry point: it resolves the Claude Code memory
// dir, opens the preference fileStore, runs one maintenance pass, prints a
// machine-readable summary, and writes the cadence/sentinel state files so
// the SessionStart-hook branch (or a cron) can skip or reactivate as needed.
// The subcommands are advisory — errors go to structured stderr and the
// command exits non-zero; they do NOT block any workflow.

package preference

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// PreferenceCmd is the parent for the preference subtree. It owns no behavior
// of its own; subcommands register under it. Exported so internal/cli/root.go
// can attach it via rootCmd.AddCommand(PreferenceCmd).
//
// Group "tools" matches the existing root.go grouping convention for utility
// commands (astgrep, telemetry, etc.).
var PreferenceCmd = &cobra.Command{
	Use:   "preference",
	Short: "Manage the user-decision memory layer captured from orchestrator questions",
	Long: `Manage the user-decision memory layer captured from orchestrator
question tool_results (SPEC-V3R6-ASKUSER-DECISION-MEMORY-001).

The preference memory persists user decisions captured from the orchestrator's
user-facing question channel under ~/.claude/projects/{slug}/memory/user_decisions/.
It has three tiers (core / recall / archival) and a power-law decay policy
with a 28-day TTL for transient entries.

Subcommands:
  decay-scan   Run one background decay pass (≤1/day, timestamp-gated)
  toggle       Toggle session-level personalization on/off (non-permanent)`,
	GroupID: "tools",
}

// decayScanFlags holds the flags for the decay-scan subcommand. It is a value
// type so tests can construct an isolated cmd + flags pair without polluting
// the package-level default cmd.
type decayScanFlags struct {
	memoryDir string // --memory-dir override (empty = auto-resolve)
	json      bool   // --json machine-readable output
	nowStr    string // --now RFC3339 override (empty = time.Now(); test injection)
	force     bool   // --force bypass the 24h timestamp gate
}

// newDecayScanCmd builds the decay-scan subcommand. Factory form keeps tests
// isolated from the package-level PreferenceCmd.
func newDecayScanCmd() *cobra.Command {
	flags := &decayScanFlags{}
	cmd := &cobra.Command{
		Use:   "decay-scan",
		Short: "Run one preference decay scan (power-law weight refresh + 28-day TTL eviction)",
		Long: `Run one preference decay scan.

The scan walks the recall tier:
  - STABLE entries: weight floored at 0.5 (pure time-decay exempt).
  - TRANSIENT entries: weight refreshed to power-law(age); entries whose
    last_used is older than 28 days are soft-deleted (moved to archival).

Cadence: at most once per 24h, gated by a timestamp file at
.moai/state/preference-decay-last-run. Use --force to bypass the gate (e.g.
for a manual maintenance run or a cron that wants to re-run after a data fix).

Exit codes: 0 success (or scan skipped due to cadence gate), 1 user error,
2 system error. The scan is advisory — it never blocks a workflow.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDecayScan(cmd.OutOrStdout(), cmd.ErrOrStderr(), flags)
		},
	}
	cmd.Flags().StringVar(&flags.memoryDir, "memory-dir", "", "preference memory dir override (default: $CLAUDE_PROJECT_DIR-derived ~/.claude/projects/{slug}/memory/)")
	cmd.Flags().BoolVar(&flags.json, "json", false, "emit the scan report as JSON (stdout)")
	cmd.Flags().StringVar(&flags.nowStr, "now", "", "override wall-clock now (RFC3339; test/debug only)")
	cmd.Flags().BoolVar(&flags.force, "force", false, "bypass the 24h cadence gate and run unconditionally")
	return cmd
}

// stateDirFromProjectRoot resolves the .moai/state directory used for the
// decay-cadence timestamp file. It mirrors resolveMemoryDir's resolution
// priority ($CLAUDE_PROJECT_DIR first, cwd fallback) but anchors to the
// project root's .moai/state/ rather than the memory dir.
func stateDirFromProjectRoot(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state")
}

// resolveProjectRoot returns the project root via the same priority order as
// resolveCaptureMemoryDir (internal/hook/user_decision_capture.go):
// $CLAUDE_PROJECT_DIR first, os.Getwd() fallback. The decay-scan subcommand
// reuses this resolution so the CLI finds the same project the hooks see.
func resolveProjectRoot() (string, error) {
	if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {
		return dir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("preference: resolve project root (cwd fallback): %w", err)
	}
	return cwd, nil
}

// resolveMemoryDirOverride applies the same resolution priority when the user
// passes --memory-dir: accept the value verbatim if set, otherwise fall back
// to the canonical ~/.claude/projects/{slug}/memory/ derivation.
func resolveMemoryDirOverride(flagValue string) (string, error) {
	if flagValue != "" {
		abs, err := filepath.Abs(flagValue)
		if err != nil {
			return "", fmt.Errorf("preference: resolve --memory-dir %q: %w", flagValue, err)
		}
		return abs, nil
	}
	root, err := resolveProjectRoot()
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("preference: user home dir: %w", err)
	}
	// Reuse the hook-layer slug derivation by computing the same way
	// resolveMemoryDir does. We inline a slug computation here so the CLI
	// does not depend on internal/hook (avoiding an import cycle with a
	// package that may import internal/cli in the future).
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("preference: absolutize project root: %w", err)
	}
	return filepath.Join(home, ".claude", "projects", memorySlug(absRoot), "memory"), nil
}

// memorySlug mirrors internal/hook/session_end.go projectSlug: encode an
// absolute path into Claude Code's project-directory naming convention
// (/ and \ and . → -). Duplicated here to avoid an internal/hook import.
func memorySlug(absPath string) string {
	clean := filepath.Clean(absPath)
	var out []rune
	for _, r := range clean {
		switch r {
		case '/', '\\', '.':
			out = append(out, '-')
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

// runDecayScan is the command body, extracted for testability. It resolves
// the memory dir, checks the cadence gate, runs the scan, prints the report,
// and marks the timestamp. All errors are wrapped and returned; the cobra
// RunE wrapper maps them to exit codes via the caller.
func runDecayScan(stdout, stderr io.Writer, flags *decayScanFlags) error {
	now := time.Now()
	if flags.nowStr != "" {
		parsed, err := time.Parse(time.RFC3339, flags.nowStr)
		if err != nil {
			return fmt.Errorf("preference: parse --now %q (want RFC3339): %w", flags.nowStr, err)
		}
		now = parsed
	}

	memDir, err := resolveMemoryDirOverride(flags.memoryDir)
	if err != nil {
		return err
	}

	// Cadence gate (unless --force). The state dir lives under the project
	// root's .moai/state/, NOT under the memory dir (which is per-user).
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}
	stateDir := stateDirFromProjectRoot(projectRoot)
	if !flags.force {
		due, dErr := ScanDue(stateDir, now)
		if dErr != nil {
			return fmt.Errorf("preference: cadence gate check: %w", dErr)
		}
		if !due {
			msg := "decay scan skipped: last run was within the 24h cadence window\n"
			if flags.json {
				_, _ = fmt.Fprintln(stdout, `{"status":"skipped","reason":"within_cadence_window"}`)
			} else {
				_, _ = fmt.Fprint(stderr, msg)
			}
			return nil
		}
	}

	storeIface, err := NewFileStore(memDir)
	if err != nil {
		return fmt.Errorf("preference: open store: %w", err)
	}
	fs, ok := storeIface.(*fileStore)
	if !ok {
		return fmt.Errorf("preference: internal error: store is not *fileStore (got %T)", storeIface)
	}

	report, err := fs.DecayScan(now)
	if err != nil {
		return fmt.Errorf("preference: decay scan: %w", err)
	}

	if err := MarkScanned(stateDir, now); err != nil {
		// Non-fatal: the scan ran, we just couldn't persist the stamp. Log to
		// stderr but do not fail the command — the advisory posture means a
		// stamp-write failure should not block the next legitimate scan.
		_, _ = fmt.Fprintf(stderr, "warn: could not write cadence stamp: %v\n", err)
	}

	if flags.json {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("preference: encode report JSON: %w", err)
		}
		return nil
	}
	_, _ = fmt.Fprintln(stdout, report.String())
	return nil
}

// init registers the children under the PreferenceCmd parent. The parent
// itself is registered under rootCmd by internal/cli/root.go via
// rootCmd.AddCommand(preference.PreferenceCmd).
func init() {
	PreferenceCmd.AddCommand(newDecayScanCmd())
	PreferenceCmd.AddCommand(newToggleCmd())
}
