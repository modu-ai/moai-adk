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

Prerequisites:
  1. A GLM API key configured via 'moai glm setup <api-key>'
  2. Running inside a tmux session for pane-level environment isolation

Examples:
  moai glm setup sk-xxx    # First: save API key (one-time)
  moai cg                  # Then: launch hybrid mode
  moai cg -p work          # Use 'work' profile with hybrid mode

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
	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		return err
	}
	return unifiedLaunch(profileName, "claude_glm", filteredArgs)
}
