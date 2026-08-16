package cli

// @MX:NOTE: [AUTO] cg command launches Claude + GLM hybrid mode for cost optimization
// @MX:NOTE: [AUTO] Requires tmux session for pane-level environment isolation
// @MX:NOTE: [AUTO] Sets teammateMode=tmux and injects GLM env for teammates

import (
	"github.com/spf13/cobra"
)

var cgCmd = &cobra.Command{
	Use:   "cg [-p profile]",
	Short: "Launch Claude Code with Claude + GLM hybrid mode",
	Long: `Launch Claude Code with hybrid mode.

CG stands for "Claude + GLM" - a cost-optimized team configuration:
  - Lead (current tmux pane): Uses Claude models (opus/sonnet)
  - Teammates (new tmux panes): Use GLM models via Z.AI proxy

This command:
  1. Validates tmux session (required for pane isolation)
  2. Removes GLM env from settings.local.json (lead = Claude)
  3. Injects GLM env into tmux session (teammates = GLM)
  4. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  5. Sets teammateMode=tmux in settings.local.json
  6. Saves team_mode: cg
  7. Launches Claude Code via exec (replaces current process)

Flags:
  -p, --profile <name>          Use a named Claude profile (~/.moai/claude-profiles/<name>/)
  --permission-mode <mode>      Set permission mode (default, acceptEdits, plan, auto, bypassPermissions, dontAsk)
  -b, --bypass                  Shorthand for --permission-mode bypassPermissions
  -w, --worktree [name]         Launch in an isolated git worktree (.claude/worktrees/<name>/);
                                name omitted = auto-generated (same as claude --worktree)
      --spawn                   Run this command in a new tmux window instead of
                                replacing the current session (requires tmux)

Prerequisites:
  1. A GLM API key configured via 'moai glm setup <api-key>'
  2. Running inside a tmux session for pane-level environment isolation

Examples:
  moai glm setup sk-xxx    # First: save API key (one-time)
  moai cg                  # Then: launch hybrid mode
  moai cg -p work          # Use 'work' profile with hybrid mode
  moai cg -w feat-auth --spawn   # GLM teammate in a new tmux window (session kept)

Use 'moai cc' to switch back to Claude-only mode.
Use 'moai glm' for all-GLM mode.`,
	GroupID:            "launch",
	DisableFlagParsing: true,
	RunE:               runCG,
}

func init() {
	rootCmd.AddCommand(cgCmd)
}

// runCG enables Claude + GLM hybrid mode and launches Claude Code.
func runCG(cmd *cobra.Command, args []string) error {
	// With DisableFlagParsing, cobra forwards --help/-h to RunE verbatim.
	// Serve help before the tmux precondition check (parity with runCC).
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	// SPEC-FACTORY-WORKER-FANOUT-001 v1.2.0: the retired -f/--factory token
	// errors on cg too — before the kanban rejection, so the operator who
	// reaches for the retired flag gets the retirement message, not a
	// backend-mismatch one that does not name their typo.
	if err := rejectRetiredFactoryFlag(args); err != nil {
		return err
	}
	// SPEC-FACTORY-MODE-001 REQ-FM-004: Kanban Mode is unavailable on cg. The
	// rejection precedes --spawn so `moai cg --kanban --spawn` fails here
	// rather than in the spawned window, where the operator would not see it.
	if err := rejectKanbanOnCG(args); err != nil {
		return err
	}
	// SPEC-FACTORY-WORKER-FANOUT-001: Factory Mode is unavailable on cg for
	// the same premise and at the same placement.
	if err := rejectFactoryOnCG(args); err != nil {
		return err
	}

	// --spawn: open a GLM teammate in a new tmux window and keep this session.
	// See cc.go for the ordering rationale.
	if spawnArgs, spawn := stripSpawnFlag(args); spawn {
		return spawnLaunch(cmd.OutOrStdout(), "cg", spawnArgs)
	}

	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		return err
	}
	// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a: see cc.go for the rationale.
	if err := resolveWorktreeL2Path(filteredArgs); err != nil {
		return err
	}
	filteredArgs = normalizeWorktreeFlag(filteredArgs)
	return unifiedLaunch(profileName, "claude_glm", filteredArgs)
}
