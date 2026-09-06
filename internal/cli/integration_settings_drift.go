package cli

// integration_settings_drift.go — the CLI half of the pre-merge
// `.claude/settings.json` drift assertion (card t488).
//
// Two surfaces, one predicate:
//
//   - `moai integration acquire` runs it as a PRECONDITION, before the window
//     is recorded. That is where it belongs because of cwd: the doctrine has a
//     lane run acquire from its own card worktree and only then enter the
//     release worktree, so the tree standing under acquire is the tree about
//     to be merged. Code does not enforce that ordering, which is why the
//     report always names the tree it measured — measuring a different tree is
//     acceptable, measuring one quietly is not.
//
//   - `moai integration preflight [path]` runs it alone, so a human can ask
//     the question outside a window and a lead can put the answer in a report.
//
// Neither surface can be dropped. Without the acquire precondition, running
// the check is a social protocol — the exact gap card t181 named when it wrote
// the announcement rule this lock exists to mechanize. Without the standalone
// verb, there is no way to ask without taking a window.
//
// The kill switch gates the REFUSAL and nothing else. Detection, preservation
// and the ledger row run on every acquire whatever its value: the reason the
// default-OFF posture was chosen is that the observed failure was nine days of
// nobody looking, not nine days of nothing being blocked, and an
// implementation that skipped detection while the flag was off would remove
// precisely the property that reasoning rests on.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

// The two surfaces get two sentinels, and the reason is this card's own rule.
//
// REQ-PSD-009 forbids stamping a bypass onto a lock record when there was no
// refusal to bypass, because a record must not describe an act that did not
// happen. `preflight` takes no window, so an error telling its caller that a
// window was REFUSED breaks the same rule one layer up — in the text a human
// reads rather than in a JSON field. Sharing one sentinel between a verb that
// refuses a window and a verb that never touches one is what made the output
// say it.
//
// The caller-facing detail is written to stdout before either is returned, so
// the report survives regardless of how the shell renders the error.
var (
	// errSettingsDriftRefused is returned by `acquire` when the refusal layer
	// is on and it declines to record the window.
	errSettingsDriftRefused = errors.New("release-integration window refused: the caller's tree has a modified tracked " + kanban.SettingsDriftWatchedPath)

	// errSettingsDriftDetected is returned by `preflight`, which reports a
	// verdict and takes nothing. It is the non-zero exit a script keys on —
	// a convenience signal, never the verdict, which is the status field.
	errSettingsDriftDetected = errors.New("settings drift detected: a modified tracked " + kanban.SettingsDriftWatchedPath + " (no integration window was taken; see the report above)")
)

// settingsDriftGateEnabled reports whether the REFUSAL layer is on for the
// project rooted at root (workflow.settings_drift_gate.enabled, distributed
// default false).
//
// Every failure path answers false, matching the four sibling guards: an
// unreadable config must not manufacture a refusal that no maintainer asked
// for. The observation layer does not consult this function at all, so a false
// here never suppresses detection.
func settingsDriftGateEnabled(root string) bool {
	cfg, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil || cfg == nil {
		return false
	}
	return cfg.Workflow.SettingsDriftGate.Enabled
}

// assessSettingsDriftForDir runs the gate against dir, preserving into root.
func assessSettingsDriftForDir(dir, root, card, branch string, bypassed bool) kanban.SettingsDriftResult {
	return kanban.AssessSettingsDrift(kanban.SettingsDriftParams{
		Dir:      dir,
		Root:     root,
		Card:     card,
		Branch:   branch,
		Bypassed: bypassed,
		Runner:   kanban.NewExecRunner(),
	})
}

// settingsDriftReportText renders the human-facing report as a string, so the
// caller performs exactly one write and can handle its error.
//
// The tree that was measured is named in every state, `clean` included: a
// verdict that does not say what it measured cannot be checked by the person
// reading it, and the acquire path does not enforce that the caller is
// standing in the tree about to be merged.
//
// The drifted file's CONTENT never appears here — only its path, hash and
// size. That file can hold tokens and machine-specific paths, which is the
// same reason the preserved copy stays untracked in a gitignored directory.
func settingsDriftReportText(r kanban.SettingsDriftResult) string {
	var b strings.Builder
	switch r.Status {
	case kanban.SettingsDriftClean:
		fmt.Fprintf(&b, "settings drift: clean (%s, 0 matches)\n", r.Worktree)
	case kanban.SettingsDriftUndetermined:
		// Not a pass, and said in those words: the absence of a signal is not
		// evidence of cleanliness, and a reader who skims must not be able to
		// take this line for one.
		fmt.Fprintf(&b, "settings drift: UNDETERMINED (%s) — not measured, which is not a pass\n  reason: %v\n", r.Worktree, r.Err)
	case kanban.SettingsDriftDetected:
		fmt.Fprintf(&b, "settings drift: DRIFT (%s, %d match)\n  file:      %s\n  sha256:    %s\n  size:      %d bytes\n",
			r.Worktree, r.MatchCount, r.Path, r.SHA256, r.SizeBytes)
		if r.PreservedPath != "" {
			fmt.Fprintf(&b, "  preserved: %s\n", r.PreservedPath)
		}
		if r.PreserveErr != nil {
			fmt.Fprintf(&b, "  preserve failed: %v (the verdict above stands — it is the match count, not the preservation)\n", r.PreserveErr)
		}
		if r.Bypassed {
			b.WriteString("  bypassed: --allow-settings-drift was given; the window was recorded anyway and the bypass is in the lock record\n")
		}
		b.WriteString("  report this to the lead with the preserved path and sha256. Nothing was restored, reverted or deleted; disposal is a human decision.\n")
	}
	return b.String()
}

// settingsDriftJSON renders the machine-readable verdict.
//
// `status` is the carrier and is ALWAYS present. It is a three-valued string
// rather than a boolean because a boolean cannot say "not measured": that
// state would collapse into false, or into an omitted field a consumer's
// default reads as a pass. `match_count` is omitted under `undetermined` for
// the same reason — a 0 there is itself a pass signal.
func settingsDriftJSON(r kanban.SettingsDriftResult) map[string]any {
	out := map[string]any{
		"status":   string(r.Status),
		"worktree": r.Worktree,
		"path":     r.Path,
	}
	if r.Status != kanban.SettingsDriftUndetermined {
		out["match_count"] = r.MatchCount
	}
	if r.SHA256 != "" {
		out["sha256"] = r.SHA256
	}
	if r.SizeBytes > 0 {
		out["size_bytes"] = r.SizeBytes
	}
	if r.PreservedPath != "" {
		out["preserved"] = r.PreservedPath
	}
	if r.Bypassed {
		out["bypassed"] = true
	}
	if r.Err != nil {
		out["error"] = r.Err.Error()
	}
	if r.PreserveErr != nil {
		out["preserve_error"] = r.PreserveErr.Error()
	}
	return out
}

// newIntegrationPreflightCmd is the standalone read surface (REQ-PSD-011).
//
// It is a new verb rather than a field on `status` because `status` answers
// "who holds the window", and folding a second question into that output makes
// its meaning depend on which question the reader had in mind.
//
// It is not purely read-only: a hit preserves and appends a ledger row, the
// same as the acquire path. Those writes leave for the PRIMARY checkout's
// state directory; the measured tree is never written to.
func newIntegrationPreflightCmd() *cobra.Command {
	var cardFlag string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "preflight [path]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Assert that a worktree has no modified tracked " + kanban.SettingsDriftWatchedPath,
		Long: `Assert that a worktree has no modified tracked ` + kanban.SettingsDriftWatchedPath + `.

This is the same check ` + "`moai integration acquire`" + ` runs before it records a
window, available on its own so a human can ask outside a window. With no
argument it measures the current directory's tree.

The verdict is the number of lines the predicate matched, never a process exit
code. Three states are reported and they are not interchangeable: clean
(measured, no drift), drift (measured, drift found), undetermined (not
measured — which is not a pass). On a hit the working copy is preserved under
the primary checkout's state directory and one ledger row is appended. Nothing
is ever restored, reverted or deleted: that file is written by the runtime and
can carry machine-specific values, so an automatic restore would itself destroy
data.

The exit code is non-zero on a hit as a convenience for scripts. It is a
signal, not the verdict — read the status field.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := ""
			if len(args) == 1 {
				dir = args[0]
			}
			if dir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot resolve the current directory: %w", err)
				}
				dir = cwd
			}
			root := integrationLockRoot()
			result := assessSettingsDriftForDir(dir, root, cardFlag, currentBranch(), false)

			if jsonOut {
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(settingsDriftJSON(result)); err != nil {
					return err
				}
			} else if err := writeSettingsDriftReport(cmd.OutOrStdout(), result); err != nil {
				return err
			}
			if result.Status == kanban.SettingsDriftDetected {
				return errSettingsDriftDetected
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&cardFlag, "card", "", "Card id recorded with the preserved copy")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

// writeSettingsDriftReport writes the report in one call so its error is
// handled rather than dropped at a dozen call sites.
func writeSettingsDriftReport(w io.Writer, r kanban.SettingsDriftResult) error {
	text := settingsDriftReportText(r)
	if text == "" {
		return nil
	}
	if _, err := io.WriteString(w, text); err != nil {
		return fmt.Errorf("write settings-drift report: %w", err)
	}
	return nil
}

// acquireSettingsDriftPrecondition runs the gate for `acquire` and reports
// whether the window may be recorded.
//
// The refusal is the ONLY thing the kill switch decides. Detection,
// preservation, the ledger row and the report all happen first and
// unconditionally, so a project that never opts in still accumulates the
// record that would have caught the two instances this card was written for.
//
// An `undetermined` verdict does not refuse. It is reported loudly and the
// window is recorded: the four sibling guards in this repository all fail open
// on uncertainty, and a lane blocked from integrating because git could not be
// run is a worse failure than an unmeasured tree that says so in its output.
func acquireSettingsDriftPrecondition(cmd *cobra.Command, root, card string, allowDrift bool) (kanban.SettingsDriftResult, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return kanban.SettingsDriftResult{
			Status:     kanban.SettingsDriftUndetermined,
			MatchCount: kanban.SettingsDriftMatchCountUnmeasured,
			Err:        err,
		}, nil
	}
	gateEnabled := settingsDriftGateEnabled(root)
	// The bypass is only a bypass when there is a refusal to bypass. With the
	// refusal layer off, recording the flag as one would make the lock record
	// state something that did not happen.
	bypassed := gateEnabled && allowDrift

	result := assessSettingsDriftForDir(cwd, root, card, currentBranch(), bypassed)
	if result.Status != kanban.SettingsDriftClean {
		if writeErr := writeSettingsDriftReport(cmd.OutOrStdout(), result); writeErr != nil {
			return result, writeErr
		}
	}
	if result.Status == kanban.SettingsDriftDetected && gateEnabled && !allowDrift {
		return result, errSettingsDriftRefused
	}
	return result, nil
}
