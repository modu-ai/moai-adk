package cli

// verify_receipts.go — the out-of-hook receipt producers of the Codex Stop
// chain (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d, design §D3.3): each runs a
// check that cannot finish inside the Codex Stop handler and records the
// outcome as a verify-snapshot receipt bound to the current tree. The Stop
// chain only compares these receipts; it never runs the checks.

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	verifyExtraCommands = append(verifyExtraCommands, newVerifySyncGateCmd)
}

// verifyReceiptContext returns the command's context, or a background one when
// the command runs outside Execute (tests call RunE directly).
func verifyReceiptContext(c *cobra.Command) context.Context {
	if ctx := c.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func newVerifySyncGateCmd(projectRoot *string) *cobra.Command {
	return &cobra.Command{
		Use:   "sync-gate",
		Short: "Run the sync-phase quality gate checks and record a receipt (Codex Stop chain)",
		Long: `Run the sync-phase quality gate's fast checks (compile/vet for the detected
language) for the current tree and record the outcome as a receipt. The Codex
Stop chain reads this receipt when HEAD is a sync-phase commit; it never runs
the checks itself. Run it after the sync-phase commit, and again whenever a
Stop continuation names it.

Prints the receipt as JSON. The verdict is "pass" or "fail"; a failing check's
output is written to stderr. Exit 0 whenever a receipt was recorded.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			root, err := verifyResolveRoot(*projectRoot)
			if err != nil {
				return err
			}
			r, err := produceSyncGateReceipt(verifyReceiptContext(c), root)
			if err != nil {
				return err
			}
			return verifyEmitJSON(c, map[string]any{
				"check_id":    r.CheckID,
				"verdict":     r.Verdict,
				"exit_code":   r.ExitCode,
				"head":        r.Head,
				"tree_digest": r.TreeDigest,
				"command":     r.Command,
			})
		},
	}
}
