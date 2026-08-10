package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness/delegationmap"
	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// delegationAnalyzeOutput is the --json stdout schema. It carries the analyzer
// result verbatim plus the two emission-side facts a caller cannot infer from
// it: whether the run wrote anything, and what it wrote.
type delegationAnalyzeOutput struct {
	delegationmap.Result

	// DryRun reports whether the write was suppressed.
	DryRun bool `json:"dry_run"`

	// Written lists the draft IDs written, empty on a dry run.
	Written []string `json:"written"`
}

// newHarnessDelegationCmd is the `moai harness delegation` parent factory
// (SPEC-HARNESS-LEARNING-EVO-002 REQ-HLA-015).
//
// It is a separate verb rather than a second source inside `moai harness
// propose` because its subject differs: propose reasons over the usage-log tier
// ladder, this reasons over the routing ledger. Folding them would require the
// ladder's mapper to accept a pattern key it is designed to reject.
func newHarnessDelegationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation",
		Short: "Delegation-map analysis from observed routing (analyze)",
		Long: `Loop 1 (analysis) over the routing observation ledger: read finalized rows,
aggregate which agents were actually delegated per /moai subcommand, and emit
proposals whose content is a concrete delegation-map amendment.

The delegation map is read only. Applying an amendment is a Tier-4 approval-gate
decision made by the orchestrator with the user; nothing here writes the map.`,
	}
	cmd.AddCommand(newHarnessDelegationAnalyzeCmd())
	return cmd
}

// newHarnessDelegationAnalyzeCmd is `moai harness delegation analyze`.
func newHarnessDelegationAnalyzeCmd() *cobra.Command {
	var (
		jsonOut    bool
		dryRun     bool
		limit      int
		ledgerPath string
		mapPath    string
	)
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze observed delegations and emit delegation-map proposals",
		Long: `Aggregate the routing ledger per subcommand and emit two kinds of proposal:
an undesignated-agent proposal, where a retained-catalog agent clears both
thresholds but the map does not designate it; and a designated-never-spawned
proposal, where a designation was observed zero times.

An absent, empty, oversized, or wholly malformed ledger is not an error: the
command exits 0 with an empty finding list and a machine-readable reason.

Use --dry-run to analyze without writing any draft.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRoot(cmd)
			if err != nil {
				return err
			}
			if ledgerPath == "" {
				ledgerPath = delegationmap.DefaultLedgerPath(root)
			}
			if mapPath == "" {
				mapPath = delegationmap.DefaultMapPath(root)
			}

			res, err := delegationmap.Analyze(delegationmap.Options{
				LedgerPath: ledgerPath,
				MapPath:    mapPath,
			})
			if err != nil {
				return fmt.Errorf("delegation analyze: %w", err)
			}
			if limit > 0 && len(res.Findings) > limit {
				res.Findings = res.Findings[:limit]
			}

			out := delegationAnalyzeOutput{Result: res, DryRun: dryRun, Written: []string{}}
			if !dryRun {
				candidates := delegationmap.BuildCandidates(res)
				written, werr := proposalgen.WriteProposals(proposalgen.ProposalDir(root), candidates)
				if werr != nil {
					return fmt.Errorf("delegation analyze: write proposals: %w", werr)
				}
				out.Written = written
				if out.Written == nil {
					out.Written = []string{}
				}
			}

			w := cmd.OutOrStdout()
			if jsonOut {
				body, merr := json.Marshal(out)
				if merr != nil {
					return fmt.Errorf("delegation analyze: marshal: %w", merr)
				}
				_, _ = fmt.Fprintln(w, string(body))
				return nil
			}

			_, _ = fmt.Fprintf(w, "reason: %s  (subcommands evaluated: %d, malformed lines: %d)\n",
				res.Reason, res.EvaluatedSubcommands, res.MalformedLines)
			for _, f := range res.Findings {
				_, _ = fmt.Fprintf(w, "%-26s %-10s %-18s support %.2f of %d rows (%d unattributed)\n",
					f.Kind, f.Subcommand, f.Agent, f.SupportRatio, f.QualifyingRows, f.UnattributedShare)
			}
			if dryRun {
				_, _ = fmt.Fprintln(w, "(dry run — no draft written)")
			} else {
				_, _ = fmt.Fprintf(w, "(%d drafts written; application requires the Tier-4 approval gate)\n", len(out.Written))
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&jsonOut, "json", false, "emit the structured result as JSON")
	f.BoolVar(&dryRun, "dry-run", false, "analyze without writing any draft")
	f.IntVar(&limit, "limit", 0, "cap the number of findings (0 = no cap)")
	f.StringVar(&ledgerPath, "ledger", "", "routing ledger path (default: the project's .moai/state ledger)")
	f.StringVar(&mapPath, "map", "", "delegation map path (default: the project's config section)")
	return cmd
}
