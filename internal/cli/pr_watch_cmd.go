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
	Use:     "pr",
	Short:   "PR-related commands (watch, abort)",
	Long:    "Commands for monitoring and managing pull requests in CI/CD workflows.",
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
		Long: `Manage CI-watch state for a pull request.

This command does NOT poll CI itself. It provides two modes for orchestrator
consumption:

  --report   Emit a ready-to-merge markdown report for PR_NUMBER to stdout.
             Call this after CI has been confirmed green.
  --abort    Request that an active watch loop stop, by setting the abort flag
             in the CI-watch state file. Reports and exits cleanly when no
             active watch exists.

With neither flag, the command prints a short notice to stderr explaining that
no watch loop runs here, and exits 0.

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

			// Default: no watch loop runs here — state the available modes.
			fmt.Fprintf(os.Stderr, "[ci-watch] No watch loop runs in this command; nothing to do for PR %s on %s.\n",
				args[0], flags.branch)
			fmt.Fprintf(os.Stderr, "[ci-watch] Use --report once CI is green, or --abort to stop an active watch.\n")
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
	// The caller (the orchestrator, after confirming CI is green) supplies the
	// PR and branch context.
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
