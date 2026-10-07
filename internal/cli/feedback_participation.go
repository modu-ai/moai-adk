package cli

// feedback_participation.go — the `moai feedback participation` subcommands
// (SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-014/021, plan.md M4): preview,
// purge, flush. Preview prints the queued payload render through the ONE
// render function the publication create path reuses; purge removes every
// user-scoped store; flush drains the spool into the queue (the send half
// of flush is M5's publication package; DEC-2's trigger set gains it there).

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
	"github.com/spf13/cobra"
)

var feedbackParticipationCmd = &cobra.Command{
	Use:   "participation",
	Short: "Inspect and control the opt-in improvement participation pipeline",
}

var feedbackParticipationPreviewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Print exactly what would be filed for each queued report (no network)",
	RunE:  runParticipationPreview,
}

var feedbackParticipationPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Remove every local participation store (queue, spool, ledger, log)",
	RunE:  runParticipationPurge,
}

var feedbackParticipationFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Drain the capture spool into the queue (time-boxed, network-free)",
	RunE:  runParticipationFlush,
}

func init() {
	feedbackParticipationCmd.AddCommand(feedbackParticipationPreviewCmd)
	feedbackParticipationCmd.AddCommand(feedbackParticipationPurgeCmd)
	feedbackParticipationCmd.AddCommand(feedbackParticipationFlushCmd)
	// The `moai feedback` root is built per-invocation (newFeedbackCmd), so
	// the participation subtree joins it there, not through a package-level
	// command variable.
}

// runParticipationPreview prints, per queued item, exactly the rendered
// title and body bytes the publication create path hands to gh — the one
// render function, reused (REQ-ANON-014). No network request exists on this
// path (the outbox package cannot import the network).
func runParticipationPreview(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	store := outbox.BugreportQueueStore()
	rec, err := store.Load()
	if err != nil {
		return fmt.Errorf("participation preview: %w", err)
	}
	for _, item := range rec.Items {
		_, _ = fmt.Fprintln(out, item.Title)
		_, _ = fmt.Fprintln(out, item.Body)
		_, _ = fmt.Fprintln(out)
	}
	_, _ = fmt.Fprintf(out, "%d queued report(s)\n", len(rec.Items))
	return nil
}

// runParticipationPurge removes the queue, the spool, the ledger, and the
// outbox log (REQ-ANON-021). No network request and no model call — the
// removal is local file work only.
func runParticipationPurge(cmd *cobra.Command, _ []string) error {
	if err := outbox.PurgeStores(); err != nil {
		return fmt.Errorf("participation purge: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Participation stores removed: queue, spool, ledger, outbox log.")
	return nil
}

// runParticipationFlush drains the user-scoped spool into the queue
// (DEC-2's flush trigger set; time-boxed inside the drain). Never fatal: a
// flush failure warns and returns nil — the command's own work is done.
// M5 wires the publication send after the drain at this call site.
func runParticipationFlush(cmd *cobra.Command, _ []string) error {
	if err := outbox.Drain(); err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "warn: participation drain failed:", err.Error())
	}
	return nil
}
