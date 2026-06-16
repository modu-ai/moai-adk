package cli

// @MX:NOTE: [AUTO] M6-S5 DDD: cobra SetHelpFunc wrapping — rootCmd help output migrated to
// tui.Section (4 groups) + tui.HelpBar. Replaces cobra's default string-interpolation template.
// Design source: screens.jsx::ScreenHelp, plan.md §7.2.
// @MX:NOTE: Help strings are English-only per language-neutrality policy (CLAUDE.local.md §15/§14).
// The 16-language neutrality applies to the distributed templates; the CLI itself ships a single
// English help surface. screens.jsx design source (Korean) is pending sync to English — tracked separately.

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// helpGroup describes a named group of commands shown in moai --help / moai help.
type helpGroup struct {
	title string
	// rows holds (commandName, description) pairs.
	rows [][2]string
}

// rootHelpGroups returns the 4 ScreenHelp groups in order.
// These must match screens.jsx::ScreenHelp group definitions (M6-S5 design contract).
func rootHelpGroups() []helpGroup {
	return []helpGroup{
		{
			title: "Project",
			rows: [][2]string{
				{"moai init", "Bootstrap a new project (wizard)"},
				{"moai doctor", "Diagnose environment · dependencies · settings"},
				{"moai status", "Summarize the current project"},
				{"moai update", "Update MoAI-ADK itself"},
				{"moai version", "Version info"},
			},
		},
		{
			title: "Launchers",
			rows: [][2]string{
				{"moai cc", "Run Claude Code (with bridge)"},
				{"moai cg", "Claude + GLM hybrid mode"},
				{"moai glm", "Run Claude Code with GLM backend"},
				{"moai web", "Launch browser-based settings console"},
				{"moai statusline", "Emit status line for tmux/vim"},
			},
		},
		{
			title: "Autonomous Development",
			rows: [][2]string{
				{"moai loop", "Spec→Plan→Impl→Sync loop"},
				{"moai spec", "Manage spec cards"},
				{"moai worktree", "Isolated worktree work"},
				{"moai brain", "Ideation workflow"},
			},
		},
		{
			title: "Governance",
			rows: [][2]string{
				{"moai constitution", "Check the CX 7 principles"},
				{"moai mx", "MX anchors · graph"},
				{"moai telemetry", "Usage stats (opt-in)"},
			},
		},
	}
}

// renderRootHelp is the cobra SetHelpFunc handler for rootCmd.
// It renders the help output using tui.Section (4 groups) + tui.HelpBar
// only when cmd is rootCmd (Use == "moai"). For subcommands, it falls back
// to cobra's default help template so "moai doctor --help" remains standard.
//
// The function writes to cmd.OutOrStdout() as required by cobra's contract,
// so that tests using SetOut(buf) receive output via the buffer.
//
// @MX:NOTE: [AUTO] Cobra SetHelpFunc inherits to subcommands; guard with cmd.Use == "moai"
// to restrict tui help to rootCmd only. Subcommand --help uses cobra default.
func renderRootHelp(cmd *cobra.Command, args []string) {
	// Only apply tui help layout to the root command.
	if cmd.Use != "moai" {
		// Restore cobra's default behavior by calling the default help function.
		// cobra's default help func writes Long + usage; replicate via cmd.UsageString().
		out := cmd.OutOrStdout()
		if cmd.Long != "" {
			_, _ = fmt.Fprintln(out, cmd.Long)
			_, _ = fmt.Fprintln(out)
		} else if cmd.Short != "" {
			_, _ = fmt.Fprintln(out, cmd.Short)
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprint(out, cmd.UsageString())
		return
	}
	renderRootHelpTUI(cmd)
}

// renderRootHelpTUI renders the tui-based help layout for rootCmd.
func renderRootHelpTUI(cmd *cobra.Command) {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	// Header: brief description line
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Fg)).Bold(true)

	_, _ = fmt.Fprintln(out, accentStyle.Render("moai-adk")+" "+descStyle.Render("· Modu-AI's Agentic Development Kit · modular command collection"))
	_, _ = fmt.Fprintln(out)

	// Determine max command column width for alignment.
	groups := rootHelpGroups()
	maxCmdWidth := 0
	for _, g := range groups {
		for _, row := range g.rows {
			if len(row[0]) > maxCmdWidth {
				maxCmdWidth = len(row[0])
			}
		}
	}

	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Bold(true)
	descRowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Body))

	for _, g := range groups {
		_, _ = fmt.Fprintln(out, tui.Section(g.title, tui.SectionOpts{Theme: &th}))
		for _, row := range g.rows {
			padded := row[0] + strings.Repeat(" ", maxCmdWidth-len(row[0]))
			line := "  " + cmdStyle.Render(padded) + "  " + descRowStyle.Render(row[1])
			_, _ = fmt.Fprintln(out, line)
		}
		_, _ = fmt.Fprintln(out)
	}

	// Usage hint
	usageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim))
	_, _ = fmt.Fprintln(out, usageStyle.Render(`Use "moai <command> --help" for more information about a command.`))
	_, _ = fmt.Fprintln(out)

	// HelpBar footer — keyboard hints matching ScreenHelp design.
	bar := tui.HelpBar([]tui.KeyHint{
		{Key: "↑↓", Label: "scroll"},
		{Key: "/", Label: "search"},
		{Key: "q", Label: "close"},
		{Key: "enter", Label: "select"},
	}, &th)
	if bar != "" {
		_, _ = fmt.Fprintln(out, bar)
	}
}
