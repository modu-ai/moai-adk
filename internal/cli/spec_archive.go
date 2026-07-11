package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/spf13/cobra"
)

// spec_archive.go — `moai spec archive` (SPEC-SESSIONSTART-PERF-001 M2).
//
// Relocates finished SPECs out of .moai/specs/ into .moai/archive/specs/<year>/,
// bounding the working set that every lifecycle scan has to walk. Archiving is a
// MOVE, never a delete: archived SPECs stay git-tracked and grep-discoverable.
//
// Confirmation model: --dry-run previews, --yes applies. A bare invocation with
// eligible SPECs reports the plan and REFUSES to move anything. That refusal is
// deliberate — this command relocates directories in bulk, and the
// grandfather-false-positive incident in
// .claude/rules/moai/core/verification-claim-integrity.md §5 is precisely what an
// unreviewed bulk relocation looks like when it goes wrong.

func newSpecArchiveCmd() *cobra.Command {
	var (
		dryRun     bool
		assumeYes  bool
		graceDays  int
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive finished SPECs out of .moai/specs/",
		Long: `Archive terminal SPECs whose last activity is older than the grace window.

A SPEC is archive-eligible only when BOTH hold:
  1. its status is terminal (completed / superseded / archived / rejected), AND
  2. its last activity is older than the grace window (default 90 days).

Age alone never archives an active SPEC. Grandfather-era SPECs are neither forced
into nor protected from archival — they are archived only when they independently
satisfy both criteria, exactly like any other SPEC.

Eligible SPECs move to .moai/archive/specs/<year>/ and stay git-tracked and
grep-discoverable. Nothing is ever deleted.

Preview first, then apply:
  moai spec archive --dry-run              # report the eligible set; move nothing
  moai spec archive --dry-run --json       # same, machine-readable
  moai spec archive --grace-days 180 --dry-run
  moai spec archive --yes                  # apply the plan`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if graceDays < 0 {
				return fmt.Errorf("--grace-days must not be negative (got %d)", graceDays)
			}

			projectRoot, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("failed to find project root: %w", err)
			}

			opts := spec.ArchiveOptions{GraceDays: graceDays}

			// Always plan first. Planning is observation-only, so both the preview and
			// the apply path see exactly the same eligible set.
			plan, err := spec.PlanArchive(projectRoot, opts)
			if err != nil {
				return fmt.Errorf("failed to plan archive: %w", err)
			}

			if jsonOutput {
				return printArchiveJSON(cmd.OutOrStdout(), plan)
			}

			printArchivePlan(cmd.OutOrStdout(), plan, dryRun)

			if dryRun || len(plan.Candidates) == 0 {
				return nil
			}

			if !assumeYes {
				// Report-then-refuse: the operator has now seen the plan above.
				return fmt.Errorf("refusing to move %d SPEC(s) without confirmation: re-run with --yes to apply, or --dry-run to preview",
					len(plan.Candidates))
			}

			applied, err := spec.ExecuteArchive(projectRoot, opts)
			if err != nil {
				return fmt.Errorf("failed to archive: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nArchived %d SPEC(s).\n", len(applied.Candidates))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Report the eligible set without moving anything")
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "Confirm the move (required to apply the plan)")
	cmd.Flags().IntVar(&graceDays, "grace-days", 0, "Grace window in days (0 = configured default, 90)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output the plan in JSON format")

	return cmd
}

// printArchiveJSON emits the plan as the sole stdout content, so the output stays
// pipeable into jq.
func printArchiveJSON(out io.Writer, plan *spec.ArchivePlan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	_, _ = fmt.Fprintln(out, string(data))
	return nil
}

// printArchivePlan renders the human-readable plan.
//
// The era column is surfaced deliberately: it lets an operator confirm at a glance
// that grandfather-protected SPECs in the list are there on their own merit
// (terminal + past grace), not because a heuristic swept them in.
func printArchivePlan(out io.Writer, plan *spec.ArchivePlan, dryRun bool) {
	header := "Archive plan"
	if dryRun {
		header = "Archive plan (dry-run — nothing will move)"
	}

	_, _ = fmt.Fprintf(out, "%s\n", header)
	_, _ = fmt.Fprintf(out, "Scanned %d SPEC(s); grace window %d days (cutoff %s).\n\n",
		plan.Scanned, plan.GraceDays, plan.Cutoff.Format("2006-01-02"))

	if len(plan.Candidates) == 0 {
		_, _ = fmt.Fprintln(out, "No SPEC is archive-eligible.")
		return
	}

	_, _ = fmt.Fprintf(out, "%-42s %-12s %-12s %-8s %s\n", "SPEC-ID", "STATUS", "LAST-ACTIVITY", "ERA", "DESTINATION")
	_, _ = fmt.Fprintln(out, strings.Repeat("-", 120))

	for _, c := range plan.Candidates {
		era := "modern"
		if c.EraFinal {
			era = "legacy"
		}
		_, _ = fmt.Fprintf(out, "%-42s %-12s %-12s %-8s %s\n",
			c.SPECID,
			c.Status,
			c.LastActivity.Format("2006-01-02"),
			era,
			c.DestDir,
		)
	}

	_, _ = fmt.Fprintln(out, strings.Repeat("-", 120))
	_, _ = fmt.Fprintf(out, "%d of %d SPEC(s) archive-eligible.\n", len(plan.Candidates), plan.Scanned)
}
