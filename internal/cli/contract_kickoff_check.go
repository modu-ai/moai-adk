package cli

// contract_kickoff_check.go — `moai contract kickoff-check <SPEC-ID> --card <card>`,
// the contract-mode plan→run gate. It writes nothing. Exit codes:
//
//	0  the gate passes
//	1  the gate fails (every reason is listed; card-mismatch included)
//	2  usage, I/O, or contract store integrity error

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// contractDoctrineAmended is the doctrine flag the sign command passes to the
// signer: the kickoff package's compiled constant.
func contractDoctrineAmended() bool { return kickoff.JevDoctrineAmended }

// kickoffConfig maps the effective configuration onto the kickoff view.
func kickoffConfig(s config.AutonomySettings) kickoff.Config {
	return kickoff.Config{
		Mode:             s.Mode,
		Decider:          s.Decider,
		DeciderJevSole:   errors.Is(s.DeciderError, config.ErrKickoffDeciderJevSole),
		JevEnabled:       s.JevEnabled,
		JevMinConfidence: s.JevMinConfidence,
	}
}

func newContractKickoffCheckCmd() *cobra.Command {
	var card string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "kickoff-check <SPEC-ID> --card <card>",
		Short: "Check the contract-mode plan→run gate (exit 0 pass, 1 fail, 2 usage or I/O error)",
		Long: `Check whether a SPEC's signed contract opens the run phase in contract mode:
the card matches, verify reports signed-valid, the contract store records the
signature, no revocation covers it, and the signature is a human interactive
signature or an issued approve receipt while autonomous Kickoff is active.
Reads only; writes nothing.`,
		Args:          contractArgs(1, 1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractKickoffCheck(c, args[0], card, asJSON)
		},
	}
	cmd.Flags().StringVar(&card, "card", "", "Card id; must equal the contract's card field (required)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the verdict as JSON")
	return cmd
}

func runContractKickoffCheck(cmd *cobra.Command, id, card string, asJSON bool) error {
	if card == "" {
		return contractUsageError(cmd, "kickoff-check: --card is required")
	}
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}
	in := kickoff.CheckInput{
		Root: env.root, SpecID: id, Card: card,
		Config:              kickoffConfig(env.autonomy),
		Policy:              env.policy(),
		RegistryRuleIDs:     env.ruleIDs,
		RegistryFrozenFiles: env.frozenFiles,
		Enabled:             kickoff.AutonomousKickoffEnabled(),
		DoctrineAmended:     kickoff.JevDoctrineAmended,
	}
	if dir, derr := contract.ResolveSpecDir(env.root, id); derr == nil {
		if status, serr := spec.ParseStatus(dir); serr == nil {
			in.SpecStatus = status
		}
	}
	res, err := kickoff.Check(in)
	if err != nil {
		return contractUsageError(cmd, "kickoff-check: %v", err)
	}
	out := cmd.OutOrStdout()
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			return contractUsageError(cmd, "kickoff-check: %v", err)
		}
	} else {
		verdict := "pass"
		if !res.Pass {
			verdict = "fail (" + res.Reason + ")"
		}
		_, _ = fmt.Fprintf(out, "kickoff-check %s: %s\nreasons: %s\nmode: %s  decider: %s  verify: %s\n",
			id, verdict, orNone(strings.Join(res.Reasons, ", ")), res.Mode, res.Decider, res.VerifyState)
	}
	if !res.Pass {
		return &exitCodeError{code: contractExitInvalid, msg: "kickoff-check: " + res.Reason}
	}
	return nil
}
