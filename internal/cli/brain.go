package cli

// @MX:NOTE: [AUTO] brain CLI는 /moai brain 슬래시 커맨드로의 thin 래퍼임
// 실제 7-phase 아이디에이션 워크플로우는 Claude Code 세션 내 슬래시 커맨드로 실행된다.
// 터미널 CLI는 사용자에게 올바른 실행 방법을 안내하는 informational hint 역할만 한다.
// 참조: CLAUDE.local.md §1 "moai CLI vs /moai Slash Command"

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// @MX:WARN: [AUTO] 사용자에게 표시되는 안내 메시지 — 변경 시 주의
// @MX:REASON: brain 워크플로우는 Claude Code 세션 외부에서 실행 불가.
// 메시지 변경 시 CLAUDE.local.md §1의 CLI/슬래시 커맨드 경계 설명과 일관성 유지 필요.

// brainInstructions는 7-phase 워크플로우 요약 (--instructions-only 플래그용).
const brainInstructions = `
/moai brain — 7-Phase Ideation Workflow Summary
================================================

Phase 1: Discovery
  Clarifies the idea via Socratic interview (AskUserQuestion, up to 5 rounds).
  Scores clarity 0-5; proceeds to Phase 2 when score >= 4.

Phase 2: Diverge
  Generates 5-15 divergent concept angles to prevent premature convergence.
  Uses moai-foundation-thinking diverge-converge framework.

Phase 3: Research
  Runs parallel WebSearch + Context7 in a single message.
  Produces .moai/brain/IDEA-NNN/research.md with cited sources.

Phase 4: Converge
  Reduces diverged angles to the strongest product concept.
  Produces .moai/brain/IDEA-NNN/ideation.md with Lean Canvas (9 blocks).

Phase 5: Critical Evaluation
  Challenges the converged concept with First Principles analysis.
  Appends Evaluation Report section to ideation.md.

Phase 6: Proposal
  Translates evaluated concept into SPEC decomposition candidates.
  Produces .moai/brain/IDEA-NNN/proposal.md
  Format: - SPEC-{DOMAIN}-{NUM}: {scope}

Phase 7: Handoff
  Produces paste-ready Claude Design handoff bundle (5 files):
    prompt.md, context.md, references.md, acceptance.md, checklist.md
  Under: .moai/brain/IDEA-NNN/claude-design-handoff/

To run: /moai brain "<your idea>" in Claude Code chat.
`

// brainInstructionsOnly는 --instructions-only 플래그 값을 저장한다.
var brainInstructionsOnly bool

// brainCmd는 /moai brain 워크플로우로 사용자를 안내하는 informational CLI 커맨드이다.
var brainCmd = &cobra.Command{
	Use:     `brain "<idea>"`,
	Short:   "Run the /moai brain ideation workflow",
	GroupID: "project",
	Long: `The /moai brain workflow converts vague ideas into validated product proposals
with Claude Design handoff packages.

IMPORTANT: This CLI command is an informational hint only.
The actual 7-phase workflow runs inside Claude Code — not in the terminal.

To run the brain workflow:
  1. Open Claude Code in your project directory
  2. Type: /moai brain "<your idea>"
  3. The workflow will guide you through 7 phases automatically

Workflow position:
  brain → [Claude Design] → /moai design --path A → /moai project → /moai plan → run → sync

Use --instructions-only to see the 7-phase contract summary.`,
	RunE: runBrain,
}

// runBrain은 brainCmd의 실제 실행 로직이다.
// --instructions-only 플래그가 있으면 7-phase 요약을 출력하고 종료한다.
// 그 외에는 사용자에게 Claude Code에서 /moai brain을 실행하라는 안내를 출력한다.
func runBrain(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// --instructions-only: 7-phase 계약 요약 출력
	if brainInstructionsOnly {
		_, _ = fmt.Fprint(out, brainInstructions)
		return nil
	}

	// 아이디어 인자가 있는 경우 — 슬래시 커맨드 안내 출력
	ideaPart := ""
	if len(args) > 0 {
		ideaPart = " \"" + strings.Join(args, " ") + "\""
	}

	_, _ = fmt.Fprintf(out, `moai brain — Claude Code slash command hint
==========================================

The /moai brain workflow runs inside Claude Code, not in the terminal.

To run the brain workflow:
  1. Open Claude Code in your project directory
  2. Type: /moai brain%s
  3. The workflow will guide you through 7 phases automatically

Tip: Use --instructions-only to see the 7-phase workflow summary.
`, ideaPart)

	return nil
}

func init() {
	// --instructions-only: 7-phase 워크플로우 계약 요약 출력 후 종료
	brainCmd.Flags().BoolVar(
		&brainInstructionsOnly,
		"instructions-only",
		false,
		"Print the 7-phase workflow contract summary and exit",
	)

	// brain 커맨드를 root 커맨드에 등록
	rootCmd.AddCommand(brainCmd)
}
