package cli

// @MX:NOTE: [AUTO] brain CLI is a thin wrapper to the /moai brain slash command
// The actual 7-phase ideation workflow runs via slash command inside Claude Code session.
// Terminal CLI serves only as informational hint guiding users to correct execution method.
// Reference: CLAUDE.local.md §1 "moai CLI vs /moai Slash Command"

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// @MX:WARN: [AUTO] User-facing guidance message — use caution when modifying
// @MX:REASON: [AUTO] Incorrect guidance can mislead users about workflow execution method
//
// brainInstructions is the 7-phase workflow summary (for --instructions-only flag).
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

// brainInstructionsOnly stores the --instructions-only flag value.
var brainInstructionsOnly bool

// brainCmd is an informational CLI command that guides users to the /moai brain workflow.
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

// runBrain is the actual execution logic of brainCmd.
// If --instructions-only flag is present, it prints 7-phase summary and exits.
// Otherwise, it prints guidance to run /moai brain in Claude Code.
func runBrain(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// --instructions-only: print 7-phase contract summary
	if brainInstructionsOnly {
		_, _ = fmt.Fprint(out, brainInstructions)
		return nil
	}

	// If idea argument is present — print slash command guidance
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
	// --instructions-only: print 7-phase workflow contract summary and exit
	brainCmd.Flags().BoolVar(
		&brainInstructionsOnly,
		"instructions-only",
		false,
		"Print the 7-phase workflow contract summary and exit",
	)

	// Register brain command to root command
	rootCmd.AddCommand(brainCmd)
}
