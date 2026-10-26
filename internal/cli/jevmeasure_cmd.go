package cli

// jevmeasure_cmd.go — SPEC-FETCH-WEB-001 M6 (REQ-FW-014, plan B6). The
// `moai jevmeasure bench` command: bench-case loader + run recorder over the
// shared fetchweb pipeline, with --baseline compare mode. Thin command layer
// — the bench semantics live in internal/jevmeasure.
//
// Run mode (no --run) executes the pipeline and is GATED behind
// workflow.jev.enabled exactly like `moai fetch web`: gate off means zero
// network requests and exit 2. Compare mode (--run) is pure data over
// recorded results and needs no gate.

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/jevmeasure"
)

func init() {
	rootCmd.AddCommand(newJevmeasureCmd())
}

// newJevmeasureCmd builds the `jevmeasure` parent with its `bench` leaf.
func newJevmeasureCmd() *cobra.Command {
	parent := &cobra.Command{
		Use:     "jevmeasure",
		Short:   "Jev measurement harness (bench cases, baselines)",
		GroupID: "tools",
	}
	bench := &cobra.Command{
		Use:   "bench --cases <cases.json> [--run <run.json>] [--baseline <path>] [--run-out <path>]",
		Short: "Run jevfetch-shaped bench cases and compare containment + token savings",
		Long: `Bench cases in the jevfetch shape: a question, an expected quote, and a
page snapshot. The runner executes the fetchweb pipeline over each snapshot
and records the output; the comparison reports per case whether the expected
quote is contained in the output and the approximate token-savings rate
versus the snapshot's token count, plus a set-level summary.

Modes:
  --run <path>          Compare a RECORDED run result (pure data, no network,
                        no gate).
  (no --run)            Execute the pipeline now, behind the
                        workflow.jev.enabled gate: gate off means zero
                        network requests and exit 2.

Baseline:
  --baseline <path>     Absent file: write it from this comparison. Present
                        file: diff deterministically — exit 0 when every
                        case's containment verdict matches the baseline,
                        exit 1 on any flip (a contained quote lost is a
                        regression).

Upstream's numeric bench claims are not asserted anywhere; this harness
reports what it measures and promises no figure.`,
		RunE:          runJevmeasureBench,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}
	bench.Flags().String("cases", "", "bench case file (JSON list of {id, question, expected_quote, page_snapshot})")
	bench.Flags().String("run", "", "recorded run result to compare (disables live execution)")
	bench.Flags().String("baseline", "", "baseline.json to write (absent) or diff (present)")
	bench.Flags().String("run-out", "", "write the executed run result JSON to this path")
	parent.AddCommand(bench)
	return parent
}

// runJevmeasureBench executes the bench command.
func runJevmeasureBench(cmd *cobra.Command, _ []string) error {
	casesPath, _ := cmd.Flags().GetString("cases")
	runPath, _ := cmd.Flags().GetString("run")
	baselinePath, _ := cmd.Flags().GetString("baseline")
	runOutPath, _ := cmd.Flags().GetString("run-out")

	if casesPath == "" {
		return fmt.Errorf("jevmeasure bench: --cases is required")
	}
	cases, err := jevmeasure.LoadBenchCases(casesPath)
	if err != nil {
		return err
	}

	var runs []jevmeasure.CaseRun
	if runPath != "" {
		raw, err := os.ReadFile(runPath)
		if err != nil {
			return fmt.Errorf("jevmeasure bench: run result unreadable: %w", err)
		}
		if err := json.Unmarshal(raw, &runs); err != nil {
			return fmt.Errorf("jevmeasure bench: run result not a JSON run list: %w", err)
		}
	} else {
		// Live execution: the gate is read FIRST and short-circuits — a
		// gated-off bench run makes zero network requests.
		root, rootErr := findProjectRootFn()
		enabled := false
		if rootErr == nil {
			enabled, _ = jevEnabled(root)
		}
		if !enabled {
			fmt.Fprintln(cmd.ErrOrStderr(), "jevmeasure bench: workflow.jev.enabled is false — no bench execution, no network request. Record a run (--run) to compare offline, or enable Jev.")
			return &exitCodeError{code: 2, msg: "jevmeasure bench: gated off (workflow.jev.enabled is false)"}
		}
		runs = jevmeasure.RunBench(cmd.Context(), cases, fetchWebAskerFn())
		if runOutPath != "" {
			raw, err := json.MarshalIndent(runs, "", "  ")
			if err == nil {
				err = os.WriteFile(runOutPath, raw, 0o600)
			}
			if err != nil {
				return fmt.Errorf("jevmeasure bench: run result write failed: %w", err)
			}
		}
	}

	if baselinePath == "" {
		report := jevmeasure.CompareRun(cases, runs)
		raw, _ := json.MarshalIndent(report, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(raw))
		return nil
	}

	if _, statErr := os.Stat(baselinePath); os.IsNotExist(statErr) {
		report, err := jevmeasure.WriteBaseline(baselinePath, cases, runs)
		if err != nil {
			return err
		}
		raw, _ := json.MarshalIndent(report, "", "  ")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n(baseline written to %s)\n", raw, baselinePath)
		return nil
	}

	diff, err := jevmeasure.DiffAgainstBaseline(baselinePath, cases, runs)
	if err != nil {
		return err
	}
	for _, c := range diff.Cases {
		verdict := "PASS"
		if !c.Pass {
			verdict = "FAIL"
		}
		line := fmt.Sprintf("%s %s", verdict, c.CaseID)
		if c.Reason != "" {
			line += ": " + c.Reason
		}
		fmt.Fprintln(cmd.OutOrStdout(), line)
	}
	if diff.AnyRegression {
		return &exitCodeError{code: 1, msg: "jevmeasure bench: baseline regression"}
	}
	return nil
}
