// Package cli provides GitHub init command.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newInitCmd creates the init command.
// T-23: Integrated bootstrap command
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize GitHub Actions integration",
		Long:  "Set up self-hosted runner for Multi-LLM CI code review.",
		Args:  cobra.NoArgs,
		RunE:  runGitHubInit,
	}
}

// runGitHubInit executes the initialization.
func runGitHubInit(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// Run interactive Wizard
	wizard := NewWizard(out)
	state, err := wizard.Run()
	if err != nil {
		return fmt.Errorf("wizard execution failed: %w", err)
	}

	// Display success message
	displayInitSuccess(out, state)

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
