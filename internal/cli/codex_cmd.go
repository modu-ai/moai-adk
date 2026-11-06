package cli

import (
	"github.com/spf13/cobra"
)

// codexCmd — the launcher-family sibling of cc/glm/cg (same group, same
// DisableFlagParsing discipline: --spawn and -- passthrough are handled by
// the launcher itself). SilenceErrors/SilenceUsage keep every diagnostic
// byte-identical to the named constants above (cobra would otherwise prefix
// "Error: " and append the usage block, breaking the exact-match cells).
var codexCmd = &cobra.Command{
	Use:   "codex [cli | status | app]",
	Short: "Codex launcher: launch the Codex CLI, or print the readiness readout",
	Long: "Launch Codex, or ask what is installed and wired.\n" +
		"\n" +
		"Called with no verb, this launches the Codex CLI at the project root.\n" +
		"The readiness readout moved to an explicit alias: moai codex status\n" +
		"reports six rows and starts nothing - the codex binary, CODEX_HOME,\n" +
		"the auth provider, the project wiring, the generated agent TOMLs, and\n" +
		"the harness entry. An incomplete wiring row is informational, not an\n" +
		"error: moai init --llm gpt generates the .codex wiring files.\n" +
		"Common local guidance shared with Claude, then Codex-specific local\n" +
		"guidance: both non-empty project-root files are injected as\n" +
		"developer instructions, in that order.\n" +
		"\n" +
		"  moai codex            launch the Codex CLI at the project root\n" +
		"  moai codex cli        the same launch, named explicitly\n" +
		"  moai codex status     print the readiness readout (starts nothing)\n" +
		"  moai codex app        launch the Codex desktop app (codex app)\n" +
		"  -w <worktree>         launch in an existing worktree (never creates one)\n" +
		"  -l, --lane            join the active factory as the next lane (CLI only)\n" +
		"  Factory sessions show launch_pending until their first prompt binds a session UUID.\n" +
		"  --spawn               open the launch in a new tmux window\n" +
		"  -d, --debug           launcher debug: trace the pre-exec path to stderr\n" +
		"                        (launcher-owned; the token never reaches codex)\n" +
		"  -- <codex-args...>    arguments after -- pass to codex verbatim",
	Example: "  # Launch the Codex CLI here\n" +
		"  moai codex\n" +
		"\n" +
		"  # Show what is installed, resolved, and wired (starts nothing)\n" +
		"  moai codex status\n" +
		"\n" +
		"  # Launch, passing arguments through verbatim\n" +
		"  moai codex -- --model o3\n" +
		"\n" +
		"  # Enter an existing worktree\n" +
		"  moai codex -w feat-login\n" +
		"\n" +
		"  # Launch the desktop app in a new tmux window\n" +
		"  moai codex app --spawn",
	GroupID:            "launch",
	DisableFlagParsing: true,
	SilenceErrors:      true,
	SilenceUsage:       true,
}

// init keeps the surviving codex parent registered next to its audit
// subcommands; the general RunE and the launch verbs were removed
// (card t1909 step ⑤, t1910 §3.3).
func init() {
	rootCmd.AddCommand(codexCmd)
}
