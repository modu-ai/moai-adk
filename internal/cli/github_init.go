// Package cli provides GitHub init command.
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// initFlags holds flags for the init command.
type initFlags struct {
	// yesBranchProtection skips the interactive Confirmer and applies branch
	// protection automatically. Use this for orchestrator-driven invocations.
	// HARD: do NOT call AskUserQuestion from Go code — use this flag instead.
	yesBranchProtection bool
}

// newInitCmd creates the init command.
// T-23: Integrated bootstrap command
func newInitCmd() *cobra.Command {
	flags := &initFlags{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize GitHub Actions integration",
		Long:  "Set up self-hosted runner for Multi-LLM CI code review.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGitHubInitWithFlags(cmd, args, flags)
		},
	}

	cmd.Flags().BoolVar(&flags.yesBranchProtection, "yes-branch-protection", false,
		"Apply branch protection rules without interactive confirmation (for orchestrator use)")

	return cmd
}

// runGitHubInitWithFlags executes the initialization with the given flags.
func runGitHubInitWithFlags(cmd *cobra.Command, _ []string, flags *initFlags) error {
	out := cmd.OutOrStdout()

	// Run interactive Wizard
	wizard := NewWizard(out)
	state, err := wizard.Run()
	if err != nil {
		return fmt.Errorf("wizard execution failed: %w", err)
	}

	// Display success message
	displayInitSuccess(out, state)

	// Apply branch protection (REQ-CIAUT-025/026) — best-effort, non-fatal.
	// On gh-unavailable / permission failure, prints exact remediation command.
	if err := applyBranchProtectionStep(cmd, flags); err != nil {
		// Non-fatal: log + continue. The error message already includes the
		// manual `gh api` command for the user to run.
		_, _ = fmt.Fprintf(out, "\nBranch protection step deferred:\n%v\n", err)
	}

	return nil
}

// applyBranchProtectionStep optionally applies branch protection to main and
// release/* using the SSoT contexts in .github/required-checks.yml.
//
// Behaviour:
//   - flags.yesBranchProtection=true: skip prompt, apply directly
//   - else: ask via Confirmer (default: TTY y/N)
//   - On gh missing/unauthenticated: returns ErrGhUnavailable wrapped error
//     with the exact `gh api -X PUT` command for manual run
func applyBranchProtectionStep(cmd *cobra.Command, flags *initFlags) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Confirmer selection: --yes-branch-protection bypasses the prompt.
	var confirmer Confirmer = &ttyConfirmer{}
	if flags.yesBranchProtection {
		confirmer = &yesConfirmer{}
	}

	ok, err := confirmer.Confirm("Apply branch protection rules to main and release/* via gh?")
	if err != nil {
		return fmt.Errorf("confirm branch protection: %w", err)
	}
	if !ok {
		return nil
	}

	// Locate project root + load SSoT.
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	checks, err := config.LoadRequiredChecks(projectRoot)
	if err != nil {
		return fmt.Errorf("load required-checks SSoT: %w", err)
	}

	gh := NewRealGhClient()
	owner, repo, err := DiscoverOwnerRepo(ctx, gh)
	if err != nil {
		return err
	}

	for branch := range checks.Branches {
		if applyErr := ApplyBranchProtection(ctx, gh, owner, repo, branch, checks); applyErr != nil {
			return applyErr
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Applied branch protection to %s/%s @ %s\n", owner, repo, branch)
	}
	return nil
}

// displayInitSuccess displays the initialization success message.
func displayInitSuccess(out interface{}, state *WizardState) {
	messages := GetMessages(state.Language)
	_, _ = fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }),
		"\n%s\n"+
			"Repository: Current directory\n"+
			"Language: %s\n"+
			"Configured LLMs: %v\n"+
			"Models: %v\n"+
			"%s",
		messages.SuccessTitle,
		state.Language,
		state.SelectedLLMs,
		state.ModelChoices,
		messages.SuccessBody,
	)
}
