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

// @MX:NOTE: [AUTO] version 명령어 출력 — tui.Box + tui.Pill 3개로 구성된 version card.
// renderVersion은 version command의 출력을 tui primitives로 구성합니다.
// 디자인 소스: screens.jsx ScreenVersion 섹션 (banner.go IMPROVE 패턴 mirror).
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
