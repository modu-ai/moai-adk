package cli

// @MX:NOTE: [AUTO] Ralph feedback loop for autonomous code quality improvement
// @MX:NOTE: [AUTO] Iterates through Analyze -> Implement -> Test -> Review phases
// @MX:NOTE: [AUTO] Uses LSP diagnostics and test results to drive loop decisions

import (
	"fmt"

	"github.com/spf13/cobra"
)

// loopCmd is the top-level command that controls the Ralph feedback loop lifecycle.
var loopCmd = &cobra.Command{
	Use:     "loop",
	Short:   "Manage the Ralph feedback loop lifecycle",
	GroupID: "tools",
	Long: `Control the Ralph autonomous feedback loop.

The loop iterates through Analyze -> Implement -> Test -> Review phases
for a given SPEC, using LSP diagnostics and test results to drive decisions.

Subcommands:
  start <SPEC-ID>    Start a new feedback loop
  status             Show current loop status
  pause              Pause the running loop
  resume <SPEC-ID>   Resume a paused loop
  cancel             Cancel and clear the running loop

Examples:
  moai loop start SPEC-001
  moai loop status
  moai loop pause
  moai loop resume SPEC-001
  moai loop cancel`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

// loopStartCmd starts a feedback loop for the specified SPEC-ID.
var loopStartCmd = &cobra.Command{
	Use:   "start <SPEC-ID>",
	Short: "Start a feedback loop for a SPEC",
	Args:  cobra.ExactArgs(1),
	RunE:  runLoopStart,
}

// loopStatusCmd prints the current feedback loop status.
var loopStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current loop status",
	Args:  cobra.NoArgs,
	RunE:  runLoopStatus,
}

// loopPauseCmd pauses the running loop.
var loopPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the running loop",
	Args:  cobra.NoArgs,
	RunE:  runLoopPause,
}

// loopResumeCmd resumes a paused loop.
var loopResumeCmd = &cobra.Command{
	Use:   "resume <SPEC-ID>",
	Short: "Resume a paused loop",
	Args:  cobra.ExactArgs(1),
	RunE:  runLoopResume,
}

// loopCancelCmd cancels the running or paused loop and removes its state.
var loopCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel the running loop",
	Args:  cobra.NoArgs,
	RunE:  runLoopCancel,
}

func init() {
	loopCmd.AddCommand(loopStartCmd)
	loopCmd.AddCommand(loopStatusCmd)
	loopCmd.AddCommand(loopPauseCmd)
	loopCmd.AddCommand(loopResumeCmd)
	loopCmd.AddCommand(loopCancelCmd)
	rootCmd.AddCommand(loopCmd)
}

// runLoopStart starts a new feedback loop for the given SPEC-ID.
func runLoopStart(cmd *cobra.Command, args []string) error {
	if deps == nil || deps.LoopController == nil {
		return fmt.Errorf("loop controller not initialized")
	}
	specID := args[0]
	if err := deps.LoopController.Start(cmd.Context(), specID); err != nil {
		return fmt.Errorf("loop start: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Loop started for SPEC %s\n", specID)
	return nil
}

// runLoopStatus prints a snapshot of the current loop status.
func runLoopStatus(cmd *cobra.Command, _ []string) error {
	if deps == nil || deps.LoopController == nil {
		return fmt.Errorf("loop controller not initialized")
	}
	status := deps.LoopController.Status()
	if status.SpecID == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active loop.")
		return nil
	}
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "SPEC:      %s\n", status.SpecID)
	_, _ = fmt.Fprintf(out, "Phase:     %s\n", status.Phase)
	_, _ = fmt.Fprintf(out, "Iteration: %d / %d\n", status.Iteration, status.MaxIter)
	_, _ = fmt.Fprintf(out, "Running:   %v\n", status.Running)
	_, _ = fmt.Fprintf(out, "Converged: %v\n", status.Converged)
	return nil
}

// runLoopPause pauses the running loop.
func runLoopPause(cmd *cobra.Command, _ []string) error {
	if deps == nil || deps.LoopController == nil {
		return fmt.Errorf("loop controller not initialized")
	}
	if err := deps.LoopController.Pause(); err != nil {
		return fmt.Errorf("loop pause: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Loop paused.")
	return nil
}

// runLoopResume resumes a paused loop by restoring state from storage.
func runLoopResume(cmd *cobra.Command, args []string) error {
	if deps == nil || deps.LoopController == nil {
		return fmt.Errorf("loop controller not initialized")
	}
	specID := args[0]
	if err := deps.LoopController.ResumeFromStorage(cmd.Context(), specID); err != nil {
		return fmt.Errorf("loop resume: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Loop resumed for SPEC %s\n", specID)
	return nil
}

// runLoopCancel cancels the loop and removes its persistent state.
func runLoopCancel(cmd *cobra.Command, _ []string) error {
	if deps == nil || deps.LoopController == nil {
		return fmt.Errorf("loop controller not initialized")
	}
	if err := deps.LoopController.Cancel(); err != nil {
		return fmt.Errorf("loop cancel: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Loop cancelled.")
	return nil
}
