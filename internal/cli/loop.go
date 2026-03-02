package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// loopCmd는 Ralph 피드백 루프의 생명주기를 제어하는 최상위 커맨드다.
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

// loopStartCmd는 지정한 SPEC-ID에 대한 피드백 루프를 시작한다.
var loopStartCmd = &cobra.Command{
	Use:   "start <SPEC-ID>",
	Short: "Start a feedback loop for a SPEC",
	Args:  cobra.ExactArgs(1),
	RunE:  runLoopStart,
}

// loopStatusCmd는 현재 피드백 루프 상태를 출력한다.
var loopStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current loop status",
	Args:  cobra.NoArgs,
	RunE:  runLoopStatus,
}

// loopPauseCmd는 실행 중인 루프를 일시정지한다.
var loopPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the running loop",
	Args:  cobra.NoArgs,
	RunE:  runLoopPause,
}

// loopResumeCmd는 일시정지된 루프를 재개한다.
var loopResumeCmd = &cobra.Command{
	Use:   "resume <SPEC-ID>",
	Short: "Resume a paused loop",
	Args:  cobra.ExactArgs(1),
	RunE:  runLoopResume,
}

// loopCancelCmd는 실행 중이거나 일시정지된 루프를 취소하고 상태를 삭제한다.
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

// runLoopStart는 SPEC-ID에 대한 새 피드백 루프를 시작한다.
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

// runLoopStatus는 현재 루프 상태 스냅샷을 출력한다.
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

// runLoopPause는 실행 중인 루프를 일시정지한다.
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

// runLoopResume는 일시정지된 루프를 스토리지에서 상태를 복원하여 재개한다.
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

// runLoopCancel은 루프를 취소하고 영속 상태를 삭제한다.
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
