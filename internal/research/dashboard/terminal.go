package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// 스타일 정의 (lipgloss 기반)
var (
	headerStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // 초록 (개선)
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // 빨강 (퇴보)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // 회색 (변화 없음)
	mustStyle   = lipgloss.NewStyle().Bold(true)                       // MUST 라벨
)

// RenderDashboard는 진행률 바, 실험 통계, 기준별 분석을 포함한
// 전체 터미널 대시보드를 렌더링한다.
func RenderDashboard(data *DashboardData) string {
	if data == nil {
		return ""
	}

	var b strings.Builder

	// 헤더
	b.WriteString(headerStyle.Render("Research Dashboard"))
	b.WriteString("\n\n")

	// 대상 및 점수 요약
	fmt.Fprintf(&b,"  Target: %s\n", data.Target)

	// 현재 점수 + 백분율
	scorePct := int(data.CurrentScore * 100)
	targetPct := int(data.TargetScore * 100)
	fmt.Fprintf(&b,"  Score:  %d%% / %d%% (target)\n", scorePct, targetPct)

	// 델타 표시 (현재 - 기준)
	delta := data.CurrentScore - data.Baseline
	deltaPct := int(delta * 100)
	var deltaStr string
	if delta > 0 {
		deltaStr = greenStyle.Render(fmt.Sprintf("+%d%%", deltaPct))
	} else if delta < 0 {
		deltaStr = redStyle.Render(fmt.Sprintf("%d%%", deltaPct))
	} else {
		deltaStr = dimStyle.Render("0%")
	}
	fmt.Fprintf(&b,"  Delta:  %s from baseline\n", deltaStr)

	// 전체 진행률 바
	scoreRatio := data.CurrentScore
	fmt.Fprintf(&b,"  Progress: %s %d%%\n", renderProgressBar(scoreRatio, 25), scorePct)
	b.WriteString("\n")

	// 실험 통계
	fmt.Fprintf(&b,"  Experiments: %d/%d", data.Experiments, data.MaxExperiments)
	fmt.Fprintf(&b,"  (Keep: %d, Discard: %d)\n", data.KeepCount, data.DiscardCount)

	// 기준별 분석
	if len(data.PerCriterion) > 0 {
		b.WriteString("\n")
		b.WriteString(headerStyle.Render("Per-Criterion Breakdown"))
		b.WriteString("\n\n")

		// 가장 긴 이름 길이 계산 (정렬용)
		maxNameLen := 0
		for _, cs := range data.PerCriterion {
			if len(cs.Name) > maxNameLen {
				maxNameLen = len(cs.Name)
			}
		}

		for _, cs := range data.PerCriterion {
			b.WriteString("  ")
			b.WriteString(renderCriterionLine(cs, maxNameLen))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// RenderCompact는 상태줄에 사용할 단일 라인 요약을 렌더링한다.
// 형식: "Research: {target} {score}% ({experiments}/{max}) {keep}K/{discard}D"
func RenderCompact(data *DashboardData) string {
	if data == nil {
		return ""
	}

	scorePct := int(data.CurrentScore * 100)
	return fmt.Sprintf("Research: %s %d%% (%d/%d) %dK/%dD",
		data.Target,
		scorePct,
		data.Experiments,
		data.MaxExperiments,
		data.KeepCount,
		data.DiscardCount,
	)
}

// renderProgressBar는 지정된 너비의 유니코드 블록 진행률 바를 생성한다.
// ratio는 0.0~1.0 범위로 클램프된다.
// █ (채움), ░ (비움) 문자를 사용한다.
func renderProgressBar(ratio float64, width int) string {
	// 범위 클램프
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// renderCriterionLine는 단일 기준의 이름, 바, 백분율, MUST 라벨을 렌더링한다.
// maxNameLen을 기준으로 이름을 좌측 패딩하여 정렬한다.
func renderCriterionLine(cs CriterionStatus, maxNameLen int) string {
	// 이름 패딩
	paddedName := fmt.Sprintf("%-*s", maxNameLen, cs.Name)

	// 백분율
	pct := int(cs.PassRate * 100)

	// 미니 진행률 바 (너비 15)
	bar := renderProgressBar(cs.PassRate, 15)

	// MUST 라벨
	var label string
	if cs.Weight == "MUST" {
		label = " " + mustStyle.Render("MUST")
	}

	return fmt.Sprintf("%s %s %3d%%%s", paddedName, bar, pct, label)
}
