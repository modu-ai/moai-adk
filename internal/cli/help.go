package cli

// @MX:NOTE: [AUTO] M6-S5 DDD: cobra SetHelpFunc wrapping — rootCmd help output migrated to
// tui.Section (4 groups) + tui.HelpBar. Replaces cobra's default string-interpolation template.
// Design source: screens.jsx::ScreenHelp, plan.md §7.2.

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
			title: "프로젝트",
			rows: [][2]string{
				{"moai init", "새 프로젝트 부트스트랩 (위저드)"},
				{"moai doctor", "환경 · 의존성 · 설정 진단"},
				{"moai status", "현재 프로젝트 요약"},
				{"moai update", "MoAI-ADK 자체 업데이트"},
				{"moai version", "버전 정보"},
			},
		},
		{
			title: "런처",
			rows: [][2]string{
				{"moai cc", "Claude Code 실행 (브릿지 포함)"},
				{"moai cg", "Claude + GLM 하이브리드 모드"},
				{"moai glm", "GLM 백엔드로 Claude Code 실행"},
				{"moai statusline", "tmux/vim용 상태줄 출력"},
			},
		},
		{
			title: "자율 개발",
			rows: [][2]string{
				{"moai loop", "Spec→Plan→Impl→Sync 루프"},
				{"moai spec", "스펙 카드 관리"},
				{"moai worktree", "워크트리 격리 작업"},
				{"moai brain", "아이데이션 워크플로우"},
			},
		},
		{
			title: "거버넌스",
			rows: [][2]string{
				{"moai constitution", "CX 7원칙 검사"},
				{"moai mx", "MX 앵커 · 그래프"},
				{"moai telemetry", "사용 통계 (옵트인)"},
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

	_, _ = fmt.Fprintln(out, accentStyle.Render("moai-adk")+" "+descStyle.Render("· 모두의 Agentic Development Kit · 모듈식 명령 모음"))
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
	_, _ = fmt.Fprintln(out, usageStyle.Render(`Use "moai <명령> --help" for more information about a command.`))
	_, _ = fmt.Fprintln(out)

	// HelpBar footer — keyboard hints matching ScreenHelp design.
	bar := tui.HelpBar([]tui.KeyHint{
		{Key: "↑↓", Label: "스크롤"},
		{Key: "/", Label: "검색"},
		{Key: "q", Label: "닫기"},
		{Key: "enter", Label: "선택"},
	}, &th)
	if bar != "" {
		_, _ = fmt.Fprintln(out, bar)
	}
}
