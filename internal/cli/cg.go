package cli

import (
	"github.com/spf13/cobra"
)

var cgCmd = &cobra.Command{
	Use:   "cg",
	Short: "Enable Claude + GLM hybrid mode",
	Long: `Enable hybrid mode where the team lead uses Claude and teammates use GLM models.

CG stands for "Claude + GLM" - a cost-optimized team configuration:
  - Lead (current tmux pane): Uses Claude models (opus/sonnet)
  - Teammates (new tmux panes): Use GLM models via Z.AI proxy

This mode requires:
  1. A GLM API key configured via 'moai glm <api-key>'
  2. Running inside a tmux session for pane-level environment isolation

If 'moai glm' was previously run, 'moai cg' automatically cleans up the
all-GLM settings and switches to hybrid mode. No need to run 'moai cc' first.

Examples:
  moai glm sk-xxx-your-key    # First: save API key (one-time setup)
  moai cg                     # Then: enable hybrid mode (direct switch)

Use 'moai cc' to switch back to Claude-only mode and disable team mode.
Use 'moai glm' for all-GLM mode (all agents use GLM).`,
	Args: cobra.NoArgs,
	RunE: runCG,
}

func init() {
	rootCmd.AddCommand(cgCmd)
}

// runCG enables Claude + GLM hybrid mode.
func runCG(cmd *cobra.Command, _ []string) error {
	// CG mode is always hybrid: lead = Claude, teammates = GLM
	return enableTeamMode(cmd, true)
}
