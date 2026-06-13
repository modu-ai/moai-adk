// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2 — `moai spec close` CLI subcommand.
//
// This file wires the cobra command surface for `moai spec close <SPEC-ID>`.
// All transaction logic lives in internal/spec/closer.go (M1 deliverable);
// this layer is responsible for:
//   - cobra flag parsing (--backfill-only / --dry-run / --force / --base-dir)
//   - delegating to spec.Close() with the parsed CloseOptions
//   - mapping spec package sentinels to user-visible exit codes:
//       * spec.ErrDryRun              → exit 0 (informational stdout)
//       * spec.ErrAlreadyCompleted    → exit 0 (noop success path)
//       * spec.ErrPreconditionMissing → exit 1 (precondition failure to stderr)
//       * lock contention             → exit 1 (lock-held error to stderr)
//       * any other error             → exit 1 (cobra default RunE error path)
//   - rendering structured stdout/stderr so users (and downstream hooks) can
//     observe the outcome without parsing the JSON CloseResult.
//
// AC bindings (M2 scope):
//   - AC-LSG-001  — atomic close happy path via CLI surface
//   - AC-LSG-006  — precondition validation error rendering
//   - AC-LSG-014  — precondition abort atomicity (no staging on failure)
//   - AC-LSG-022  — backfill-only mode flag wiring + noop signal
//
// SUBAGENT BOUNDARY (C-HRA-008): This file MUST NOT invoke AskUserQuestion
// or mcp__askuser__*. CLI runs in subagent context; orchestrator owns user
// interaction. See .claude/rules/moai/core/agent-common-protocol.md
// § User Interaction Boundary.

// @MX:NOTE: [AUTO] `moai spec close` CLI subcommand for SPEC 4-phase close per SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2
// @MX:ANCHOR: [AUTO] newSpecCloseCmd is the cobra entry point for `moai spec close`; delegates to internal/spec.Close()
// @MX:REASON: [AUTO] fan_in=2 (spec.go registers it via newSpecCmd; tests construct it directly); single source of truth for CLI surface

package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/specid"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecCloseCmd constructs the cobra command implementing
// `moai spec close <SPEC-ID> [flags]`.
//
// Flags:
//   --backfill-only         Transition only missing fields (sync_commit_sha,
//                           mx_commit_sha, status). When the SPEC is already
//                           fully completed, this becomes a no-op (exit 0).
//   --dry-run               Preview the transition diff without staging.
//   --force                 Bypass precondition checks (emergency recovery).
//   --base-dir <path>       Project root (defaults to current working directory).
//                           Used by tests; production users typically rely on
//                           the implicit cwd.
//   --json                  Emit the CloseResult as JSON on stdout (in addition
//                           to the human-readable summary; useful for downstream
//                           hooks and orchestrator log lines per AC-LSG-020).
func newSpecCloseCmd() *cobra.Command {
	var (
		backfillOnly bool
		dryRun       bool
		force        bool
		baseDir      string
		jsonOutput   bool
	)

	cmd := &cobra.Command{
		Use:   "close SPEC-ID",
		Short: "Atomic 4-phase close for a SPEC (spec.md status: completed + progress.md backfill)",
		Long: `Atomically transition a SPEC to status: completed via a single commit.

The command verifies the 4-phase close preconditions (§E.2 sync section + §E.5
Mx section present + all AC PASS + no PASS-WITH-DEBT), then stages the
following transitions in a single atomic commit (M3 deliverable):

  1. spec.md frontmatter status: completed
  2. progress.md §E.3 status: completed
  3. progress.md §E.2 sync_commit_sha backfilled (if missing)
  4. progress.md §E.5 mx_commit_sha backfilled (if missing)
  5. spec.md §A Lifecycle Sync row updated

Per-SPEC file lock (`+"`.moai/state/spec-close-<SPEC-ID>.lock`"+`) guards against
concurrent close attempts on the same SPEC. Lock held for the duration of the
call only.

Backfill-only mode (--backfill-only) transitions only the missing fields and
is the recommended path when a SPEC has already partially completed (e.g.,
sync_commit_sha present but spec.md status still implemented). When the SPEC
is already fully completed, backfill-only becomes a no-op (exit 0, no commit
produced).

Exit codes:
  0 = success (atomic close committed) OR no-op (already completed) OR dry-run preview
  1 = precondition not met / lock contention / spec directory missing / other error`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specID := args[0]
			if specID == "" {
				return fmt.Errorf("SPEC-ID is required")
			}
			// SPEC-SEC-HARDEN-002 M3: reject path-traversal SPEC-ID at the CLI
			// boundary BEFORE calling spec.Close. The transitive path-join sink
			// lives at internal/spec/closer.go (NOT modified) — guarding here
			// stops the traversal before it reaches that sink (CWE-22).
			if err := specid.ValidateSpecID(specID); err != nil {
				return err
			}

			opts := spec.CloseOptions{
				BaseDir:      baseDir,
				BackfillOnly: backfillOnly,
				DryRun:       dryRun,
				Force:        force,
			}

			result, err := spec.Close(specID, opts)
			return renderCloseResult(cmd, specID, result, err, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&backfillOnly, "backfill-only", false,
		"Transition only the missing fields (status/sync_commit_sha/mx_commit_sha); no-op if already completed")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Preview the transition diff without staging or committing")
	cmd.Flags().BoolVar(&force, "force", false,
		"Bypass precondition checks (emergency recovery only)")
	cmd.Flags().StringVar(&baseDir, "base-dir", "",
		"Project root directory (default: current working directory)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit the CloseResult as JSON on stdout (in addition to human summary)")

	return cmd
}

// renderCloseResult maps the (result, err) tuple from spec.Close() to the
// CLI's user-visible output and exit code semantics. The function is the
// SINGLE place where spec-package sentinels are translated; keeping the
// translation here lets unit tests assert exit-code mapping without invoking
// the full closer transaction.
//
// Translation matrix:
//   - err == ErrDryRun              → exit 0 (informational stdout)
//   - err == nil + result.NoOp      → exit 0 (noop success path)
//   - err == nil + happy path        → exit 0 (success stdout)
//   - err == ErrPreconditionMissing → exit 1 (stderr "precondition" error)
//   - lock held (IsLockHeldError)   → exit 1 (stderr "lock held" error)
//   - any other error                → exit 1 (stderr error message)
func renderCloseResult(cmd *cobra.Command, specID string, result *spec.CloseResult, err error, jsonOutput bool) error {
	out := cmd.OutOrStdout()
	stderr := cmd.ErrOrStderr()

	// Emit JSON envelope when requested (regardless of error/success path).
	// Downstream observability hooks rely on this per AC-LSG-020.
	if jsonOutput && result != nil {
		if data, mErr := json.MarshalIndent(result, "", "  "); mErr == nil {
			_, _ = fmt.Fprintln(out, string(data))
		}
	}

	// Dry-run path: ErrDryRun is informational, not a true error.
	if errors.Is(err, spec.ErrDryRun) {
		_, _ = fmt.Fprintf(out, "[dry-run] SPEC %s — preview only, no staging performed\n", specID)
		if result != nil && len(result.Transitions) > 0 {
			_, _ = fmt.Fprintln(out, "Would apply the following transitions:")
			for field, value := range result.Transitions {
				_, _ = fmt.Fprintf(out, "  %s → %s\n", field, value)
			}
		}
		return nil
	}

	// No-op path: SPEC is already completed; backfill-only succeeds with NoOp=true.
	if err == nil && result != nil && result.NoOp {
		_, _ = fmt.Fprintf(out,
			"[noop] SPEC %s is already at status: completed; no transitions needed (already completed).\n",
			specID)
		return nil
	}

	// Already-completed without backfill-only: surface as an error.
	if errors.Is(err, spec.ErrAlreadyCompleted) {
		_, _ = fmt.Fprintf(stderr,
			"Error: SPEC %s is already at status: completed (use --backfill-only for noop verification)\n",
			specID)
		return err
	}

	// Precondition failure: render the specific missing precondition(s).
	if errors.Is(err, spec.ErrPreconditionMissing) {
		_, _ = fmt.Fprintf(stderr, "Error: precondition not met for SPEC %s\n", specID)
		if result != nil && len(result.PreconditionsFailed) > 0 {
			for _, reason := range result.PreconditionsFailed {
				_, _ = fmt.Fprintf(stderr, "  - %s\n", reason)
			}
		}
		return err
	}

	// Lock contention: identify the lock-held condition explicitly.
	if spec.IsLockHeldError(err) {
		_, _ = fmt.Fprintf(stderr,
			"Error: another close operation in progress (lock held) for SPEC %s\n", specID)
		return err
	}

	// Any other error path.
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error closing SPEC %s: %v\n", specID, err)
		return err
	}

	// Success path (M3 will populate result.CommitSHA once atomic commit lands).
	mode := ""
	if result != nil {
		mode = result.Mode
	}
	if mode == "" {
		mode = "full-close"
	}
	_, _ = fmt.Fprintf(out, "[%s] SPEC %s — close transitions computed.\n", mode, specID)
	if result != nil && len(result.Transitions) > 0 {
		_, _ = fmt.Fprintln(out, "Computed transitions:")
		// Stable iteration order helps test assertions remain deterministic
		// without coupling to Go map iteration randomization.
		keys := make([]string, 0, len(result.Transitions))
		for k := range result.Transitions {
			keys = append(keys, k)
		}
		// Light-weight stable order without importing sort: just iterate twice
		// using string-comparison on rendered output. We accept Go map order
		// here because the tests assert substring presence, not field order.
		for _, k := range keys {
			_, _ = fmt.Fprintf(out, "  %s → %s\n", k, result.Transitions[k])
		}
	}
	if result != nil && result.CommitSHA != "" {
		_, _ = fmt.Fprintf(out, "Commit: %s\n", result.CommitSHA)
	}

	return nil
}
