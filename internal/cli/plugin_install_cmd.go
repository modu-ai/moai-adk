package cli

// plugin_install_cmd.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, the `moai plugin
// install` verb. The install scripts call it by the installed path after the
// binary lands (OD-6 option a). It has no harness: it acts on every tool it
// finds, prints on stderr, honors the opt-out and exits 0 on every outcome the
// step treats as fail-open; only a usage error exits non-zero (REQ-019).

import (
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// newPluginCmd builds the `plugin` noun group (one leaf, in the tools help
// group).
func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Short:   "Manage the moai plugin for Claude Code and Codex",
		GroupID: "tools",
	}
	cmd.AddCommand(newPluginInstallCmd())
	return cmd
}

func newPluginInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the moai plugin into every Claude Code and Codex found on PATH",
		Long: `Add the moai-adk marketplace and install the moai plugin into every
Claude Code and Codex found on PATH (two commands per tool, the second only
after the first succeeds). The tools' own default scope applies and the child
processes inherit this environment unchanged, so the profile that moves is the
one CLAUDE_CONFIG_DIR / CODEX_HOME (else ~/.claude / ~/.codex) names.

A tool that is absent is skipped with one line; a tool command that fails or
times out prints the two commands to run yourself. The command exits 0 in
every such case. Set ` + config.EnvSkipPluginInstall + `=1 to run no tool command at all.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// The verb has no init target: the llm.claude_bin pin comes from the
			// project the working directory sits in, as it does for the launcher.
			root, _ := findProjectRoot()
			opts := newPluginInstallOptions([]pluginTool{pluginToolClaude, pluginToolCodex}, root, false)
			return runPluginInstallStep(cmd.ErrOrStderr(), opts)
		},
	}
}
