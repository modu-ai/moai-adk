package cli

// contract_revoke.go — `moai contract revoke <card> --spec <SPEC-ID>`.
//
// A thin adapter over internal/contract/revoke. Exit codes:
//
//	0  revoked, or already revoked (idempotent; nothing written)
//	1  no contract, or the contract is not signed (nothing written)
//	2  usage, card mismatch, I/O, or contract store integrity error

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/contract/revoke"
)

func newContractRevokeCmd() *cobra.Command {
	var specID string
	cmd := &cobra.Command{
		Use:   "revoke <card> --spec <SPEC-ID>",
		Short: "Revoke a signed SPEC contract (exit 0 revoked, 1 not signed, 2 usage or I/O error)",
		Long: `Withdraw the signature of a SPEC contract. Appends a revoke event to the
contract store and writes one escalation record of kind revoke; a run in
progress stops at the next stage boundary. Revocation never deletes a worktree
or branch, pushes, changes the queue, or edits the contract or SPEC documents.`,
		Args:          contractCardArg,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractRevoke(c, args[0], specID)
		},
	}
	cmd.Flags().StringVar(&specID, "spec", "", "SPEC ID whose contract is revoked (required)")
	return cmd
}

// contractCardArg requires exactly one positional card id (exit 2 otherwise).
func contractCardArg(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return contractUsageError(cmd, "%s: takes exactly one card id (%d given)", cmd.CommandPath(), len(args))
	}
	return nil
}

func runContractRevoke(cmd *cobra.Command, card, specID string) error {
	if specID == "" {
		return contractUsageError(cmd, "revoke: --spec is required")
	}
	root, err := findProjectRootFn()
	if err != nil {
		return contractUsageError(cmd, "%v", err)
	}
	res, err := revoke.Revoke(revoke.Options{Root: root, SpecID: specID, Card: card}, revoke.Seams{})
	if err != nil {
		return contractUsageError(cmd, "revoke: %v", err)
	}
	out := cmd.OutOrStdout()
	switch res.Status {
	case revoke.StatusRevoked:
		_, _ = fmt.Fprintf(out, "revoked %s (card %s, seal %s)\nrecord: %s\n", specID, card, shortSeal(res.Seal), res.RecordPath)
	case revoke.StatusAlreadyRevoked:
		_, _ = fmt.Fprintf(out, "already revoked %s (card %s, seal %s); nothing written\n", specID, card, shortSeal(res.Seal))
	case revoke.StatusNotSigned:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai contract: revoke: %s has no signed contract; nothing written\n", specID)
		return &exitCodeError{code: contractExitInvalid, msg: "revoke: " + specID + " has no signed contract"}
	default:
		return contractUsageError(cmd, "revoke: unexpected status %q", res.Status)
	}
	return nil
}

func shortSeal(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
