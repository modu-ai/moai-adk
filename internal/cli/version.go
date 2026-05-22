package cli

// @MX:NOTE: [AUTO] Version command displays moai-adk version, commit hash, build date
// @MX:NOTE: [AUTO] Version injected at build time via ldflags -X github.com/modu-ai/moai-adk/pkg/version.Version

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// @MX:NOTE: [AUTO] version command output — version card composed of tui.Box + 3 tui.Pills.
// renderVersion composes the version command output using tui primitives.
// Design source: screens.jsx ScreenVersion section (mirrors banner.go IMPROVE pattern).
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show version information",
	GroupID: "tools",
	RunE: func(cmd *cobra.Command, _ []string) error {
		out := cmd.OutOrStdout()
		th := resolveTheme()

		title := tui.Box(tui.BoxOpts{
			Title:  "moai-adk " + version.GetVersion(),
			Accent: true,
			Theme:  &th,
		})

		p1 := tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: version.GetVersion(), Theme: &th})
		p2 := tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: shortCommit(version.GetCommit()), Theme: &th})
		p3 := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "built " + version.GetDate(), Theme: &th})
		row := lipgloss.JoinHorizontal(lipgloss.Top, p1, " ", p2, " ", p3)

		_, _ = fmt.Fprintln(out, title)
		_, _ = fmt.Fprintln(out, row)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
