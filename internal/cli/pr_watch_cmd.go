package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
	clipr "github.com/modu-ai/moai-adk/internal/cli/pr"
)

// prCmd is the top-level "moai pr" command group.
var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "PR-related commands (watch, abort)",
	Long:  "Commands for monitoring and managing pull requests in CI/CD workflows.",
	GroupID: "project",
}

// prWatchFlags holds flags for the "moai pr watch" subcommand.
type prWatchFlags struct {
	abort  bool
	report bool
	branch string
}

func newPRWatchCmd() *cobra.Command {
	flags := &prWatchFlags{}

	cmd := &cobra.Command{
		Use:   "watch <PR_NUMBER>",
		Short: "Watch CI checks for a PR (or --abort an active watch)",
		Long: `Monitor gh pr checks for the given PR number.

On all-required-pass: emits a ready-to-merge markdown report to stdout.
On required-failure:  emits a JSON handoff to stdout (exit 2).
On 30-min timeout:    exits with code 3.

The CI watch loop is invoked via scripts/ci-watch/run.sh.
This command handles --abort and --report modes for orchestrator consumption.

HARD: This command MUST NOT call AskUserQuestion.
The orchestrator presents the emitted report via AskUserQuestion.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if flags.abort {
				return nil // --abort needs no positional arg
			}
			if len(args) < 1 {
				return fmt.Errorf("PR_NUMBER is required (or use --abort)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			statePath := filepath.Join(cwd, ciwatch.StateFile)

			if flags.abort {
				return runPRWatchAbort(statePath)
			}

			if flags.report {
				return runPRWatchReport(args[0], flags.branch)
			}

			// Default: just print usage info directing user to the shell script.
			fmt.Fprintf(os.Stderr, "[ci-watch] Use scripts/ci-watch/run.sh to start the watch loop.\n")
			fmt.Fprintf(os.Stderr, "[ci-watch] Example: MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh %s %s\n",
				args[0], flags.branch)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.abort, "abort", false, "Abort the active CI watch loop")
	cmd.Flags().BoolVar(&flags.report, "report", false, "Emit ready-to-merge report for PR_NUMBER")
	cmd.Flags().StringVar(&flags.branch, "branch", "main", "Branch name for report context")

	return cmd
}

// runPRWatchAbort sets the abort flag in the active state file.
func runPRWatchAbort(statePath string) error {
	if err := ciwatch.SetAbortFlag(statePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "[ci-watch] No active watch found (state file missing)")
			return nil
		}
		return fmt.Errorf("set abort flag: %w", err)
	}
	fmt.Fprintln(os.Stderr, "[ci-watch] Abort requested — watch loop will stop within 30s")
	return nil
}

// runPRWatchReport emits a ready-to-merge report to stdout for orchestrator consumption.
func runPRWatchReport(prNumStr, branch string) error {
	var prNum int
	if _, err := fmt.Sscanf(prNumStr, "%d", &prNum); err != nil {
		return fmt.Errorf("invalid PR_NUMBER %q: %w", prNumStr, err)
	}

	// Build a minimal all-pass state for report generation.
	// Caller (scripts/ci-watch/run.sh exit 0) passes PR+branch context.
	state := ciwatch.CIState{
		PRNumber: prNum,
		Branch:   branch,
		// RequiredPassed will be 0 if not provided; report shows "0/0 pass" in that case.
		// Orchestrator should call this after confirming all-pass from run.sh exit 0.
	}

	return clipr.EmitReadyToMergeReport(os.Stdout, state)
}

func init() {
	prCmd.AddCommand(newPRWatchCmd())
	rootCmd.AddCommand(prCmd)
}
