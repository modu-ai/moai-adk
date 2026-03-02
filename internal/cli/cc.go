package cli

import (
	"github.com/spf13/cobra"
)

// findProjectRootFn is the function used to locate the project root.
// Tests may override this to return a temporary directory, preventing
// any file mutations from reaching the real project or home directory.
var findProjectRootFn = findProjectRoot

var ccCmd = &cobra.Command{
	Use:   "cc [-p profile] [-- claude-args...]",
	Short: "Launch Claude Code with Claude backend",
	Long: `Launch Claude Code with Claude backend.

This command:
  1. Removes GLM-specific environment variables from .claude/settings.local.json
  2. Resets team mode if it was enabled (glm or cg)
  3. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  4. Reads DO_CLAUDE_* settings and converts them to CLI flags
  5. Launches Claude Code via exec (replaces current process)

Flags:
  -p, --profile <name>   Use a named Claude profile (~/.moai/claude-profiles/<name>/)
  -b, --bypass           Enable --dangerously-skip-permissions
  -c, --continue         Continue previous session
  -m, --model <model>    Override model selection
  --chrome / --no-chrome Toggle Chrome MCP

Examples:
  moai cc                     # Default profile, launch Claude
  moai cc -p work             # Use 'work' profile
  moai cc -p work -- --print  # Profile + pass-through args to Claude`,
	GroupID:            "launch",
	DisableFlagParsing: true,
	RunE:               runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude, then launches Claude Code.
func runCC(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	profileName, filteredArgs := parseProfileFlag(args)
	return unifiedLaunch(profileName, "claude", filteredArgs)
}
