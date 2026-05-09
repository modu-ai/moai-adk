package cli

// @MX:NOTE: [AUTO] Ralph feedback loop for autonomous code quality improvement
// @MX:NOTE: [AUTO] Iterates through Spec -> Plan -> Implement -> Sync phases (4-phase Stepper)
// @MX:NOTE: [AUTO] Uses LSP diagnostics and test results to drive loop decisions
// @MX:NOTE: [AUTO] M6-S3 DDD: status output migrated to tui.Box + tui.Stepper + tui.Section + tui.KV

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/spf13/cobra"
)

// loopPhaseLabels is the canonical 4-phase label set for the Ralph feedback loop.
// Order matches ScreenLoop in screens.jsx: Spec → Plan → Implement → Sync.
//
// @MX:ANCHOR: [AUTO] loopPhaseLabels defines the canonical phase order for loop status display
// @MX:REASON: loopPhaseLabels is referenced by runLoopStatus (status render) and renderLoopPhaseStrip
var loopPhaseLabels = [4]string{"Spec", "Plan", "Implement", "Sync"}

// phaseIndex returns the 1-based index of phase within loopPhaseLabels.
// Returns 0 if phase is not found (e.g., empty string or custom value).
func phaseIndex(phase string) int {
	for i, label := range loopPhaseLabels {
		if strings.EqualFold(label, phase) {
			return i + 1
		}
	}
	return 0
}

// renderLoopPhaseStrip renders the 4-phase horizontal strip (Spec/Plan/Implement/Sync).
// Each phase is rendered as a tui.Box panel; the active phase uses Accent border.
// statusIcon selects the icon for each phase: "ok", "run", "pending".
func renderLoopPhaseStrip(activePhase string, th *tui.Theme) string {
	var parts []string
	for i, label := range loopPhaseLabels {
		idx := i + 1
		activeIdx := phaseIndex(activePhase)

		var status string
		switch {
		case idx < activeIdx:
			status = "ok"
		case idx == activeIdx:
			status = "run"
		default:
			status = "skip"
		}

		icon := tui.StatusIcon(status)
		accent := idx == activeIdx
		body := icon + " " + label

		parts = append(parts, tui.Box(tui.BoxOpts{
			Title:  fmt.Sprintf("%d", idx),
			Body:   body,
			Theme:  th,
			Accent: accent,
		}))
	}
	return strings.Join(parts, "  ")
}

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

	th := tui.LightTheme()
	pill := tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: "실행 중", Theme: &th})
	stepper := tui.Stepper(1, 4, &th)
	header := tui.Section("자율 개발 루프", tui.SectionOpts{Theme: &th})
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), header)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("SPEC", specID, tui.KVOpts{Theme: &th, KeyWidth: 10}))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("상태", pill, tui.KVOpts{Theme: &th, KeyWidth: 10}))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("단계", stepper, tui.KVOpts{Theme: &th, KeyWidth: 10}))
	return nil
}

// runLoopStatus prints a snapshot of the current loop status.
// Output uses tui.Box (4-phase strip), tui.Stepper, tui.Section, and tui.KV
// to match the ScreenLoop design (screens.jsx §12).
func runLoopStatus(cmd *cobra.Command, _ []string) error {
	if deps == nil || deps.LoopController == nil {
		return fmt.Errorf("loop controller not initialized")
	}
	status := deps.LoopController.Status()
	if status.SpecID == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active loop.")
		return nil
	}

	th := tui.LightTheme()
	out := cmd.OutOrStdout()

	// Header: section title
	_, _ = fmt.Fprintln(out, tui.Section("자율 개발 루프", tui.SectionOpts{Theme: &th}))

	// 4-phase strip (Spec → Plan → Implement → Sync)
	_, _ = fmt.Fprintln(out, renderLoopPhaseStrip(string(status.Phase), &th))

	// Meta-data: phase stepper + KV rows
	phaseIdx := phaseIndex(string(status.Phase))
	if phaseIdx == 0 {
		phaseIdx = 1 // default to Spec if phase is unrecognised
	}
	stepper := tui.Stepper(phaseIdx, 4, &th)
	_, _ = fmt.Fprintln(out, tui.Section("현재 상태", tui.SectionOpts{Theme: &th}))
	_, _ = fmt.Fprintln(out, tui.KV("SPEC", status.SpecID, tui.KVOpts{Theme: &th, KeyWidth: 12}))
	_, _ = fmt.Fprintln(out, tui.KV("단계", stepper, tui.KVOpts{Theme: &th, KeyWidth: 12}))
	_, _ = fmt.Fprintln(out, tui.KV("반복", fmt.Sprintf("%d / %d", status.Iteration, status.MaxIter), tui.KVOpts{Theme: &th, KeyWidth: 12}))

	var runPill string
	if status.Running {
		runPill = tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: "실행 중", Theme: &th})
	} else {
		runPill = tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Label: "정지", Theme: &th})
	}
	_, _ = fmt.Fprintln(out, tui.KV("상태", runPill, tui.KVOpts{Theme: &th, KeyWidth: 12}))

	if status.Converged {
		convergePill := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Label: "수렴 완료", Theme: &th})
		_, _ = fmt.Fprintln(out, tui.KV("수렴", convergePill, tui.KVOpts{Theme: &th, KeyWidth: 12}))
	}
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
	th := tui.LightTheme()
	pill := tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Label: "일시정지", Theme: &th})
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("루프", pill, tui.KVOpts{Theme: &th, KeyWidth: 8}))
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
	th := tui.LightTheme()
	stepper := tui.Stepper(1, 4, &th)
	pill := tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: "재시작", Theme: &th})
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("SPEC", specID, tui.KVOpts{Theme: &th, KeyWidth: 10}))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("상태", pill, tui.KVOpts{Theme: &th, KeyWidth: 10}))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("단계", stepper, tui.KVOpts{Theme: &th, KeyWidth: 10}))
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
	th := tui.LightTheme()
	pill := tui.Pill(tui.PillOpts{Kind: tui.PillErr, Label: "취소됨", Theme: &th})
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), tui.KV("루프", pill, tui.KVOpts{Theme: &th, KeyWidth: 8}))
	return nil
}
