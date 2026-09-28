package cli

// contract_decide.go — `moai contract decide <card> --spec <SPEC-ID> --judgement <file|->`,
// the kickoff decision. Exit codes:
//
//	0  the decision was recorded (the outcome is in --json)
//	1  contract store integrity failure (nothing written)
//	2  usage, card mismatch, decider jev, malformed judgement, or I/O error
//	   (nothing written)

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
)

func newContractDecideCmd() *cobra.Command {
	var specID, judgement string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "decide <card> --spec <SPEC-ID> --judgement <file|->",
		Short: "Record a kickoff decision for a SPEC contract (exit 0 recorded, 1 store integrity, 2 usage or I/O error)",
		Long: `Evaluate the kickoff preconditions and outcome rules for a SPEC's contract and
record the decision in the contract store. The deciding LLM's judgement is
passed with --judgement (a JSON file, or - for standard input); decide calls no
LLM itself. A decision appends an issued receipt and writes
.moai/specs/<SPEC-ID>/kickoff-receipt.json; a path that does not decide records
only its event. decide never signs.`,
		Args:          contractCardArg,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractDecide(c, args[0], specID, judgement, asJSON)
		},
	}
	cmd.Flags().StringVar(&specID, "spec", "", "SPEC ID whose contract is decided (required)")
	cmd.Flags().StringVar(&judgement, "judgement", "", "Judgement JSON file, or - for standard input (required)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the decision as JSON")
	return cmd
}

func runContractDecide(cmd *cobra.Command, card, specID, judgementPath string, asJSON bool) error {
	if specID == "" || judgementPath == "" {
		return contractUsageError(cmd, "decide: --spec and --judgement are required")
	}
	var raw []byte
	var err error
	if judgementPath == "-" {
		raw, err = io.ReadAll(cmd.InOrStdin())
	} else {
		raw, err = os.ReadFile(judgementPath)
	}
	if err != nil {
		return contractUsageError(cmd, "decide: read judgement: %v", err)
	}
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}
	res, err := kickoff.Decide(kickoff.DecideInput{
		Root: env.root, SpecID: specID, Card: card, Judgement: raw,
		Config:              kickoffConfig(env.autonomy),
		Policy:              env.policy(),
		RegistryRuleIDs:     env.ruleIDs,
		RegistryFrozenFiles: env.frozenFiles,
		DoctrineAmended:     kickoff.JevDoctrineAmended,
	})
	switch {
	case errors.Is(err, receipt.ErrIntegrity):
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai contract: decide: %v\n", err)
		return &exitCodeError{code: contractExitInvalid, msg: "decide: contract store integrity"}
	case errors.Is(err, kickoff.ErrCardMismatch):
		return contractUsageError(cmd, "decide: card-mismatch: %v", err)
	case errors.Is(err, kickoff.ErrDeciderJevRefused):
		return contractUsageError(cmd, "decide: decider-jev-refused: %v", err)
	case err != nil:
		return contractUsageError(cmd, "decide: %v", err)
	}
	out := cmd.OutOrStdout()
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			return contractUsageError(cmd, "decide: %v", err)
		}
		return nil
	}
	_, _ = fmt.Fprintf(out, "decide %s (card %s): outcome %s", specID, card, res.Outcome)
	if res.Reason != "" {
		_, _ = fmt.Fprintf(out, " (%s)", res.Reason)
	}
	_, _ = fmt.Fprintln(out)
	if res.ReceiptPath != "" {
		_, _ = fmt.Fprintf(out, "receipt: %s\n", res.ReceiptPath)
	}
	return nil
}
