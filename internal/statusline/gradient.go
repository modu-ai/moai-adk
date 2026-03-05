package statusline

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// interpolateGradientColor 는 블록 위치(0.0~1.0)에 대한 그라디언트 RGB 값을 반환한다.
// 경로: Green(0,255,0) → Yellow(255,255,0) → Red(255,0,0)
//
//	0.0~0.5: Green → Yellow
//	0.5~1.0: Yellow → Red
func interpolateGradientColor(blockPct float64) (r, g, b int) {
	// 범위 외 입력 클램핑
	if blockPct <= 0 {
		return 0, 255, 0
	}
	if blockPct >= 1 {
		return 255, 0, 0
	}

	if blockPct <= 0.5 {
		// Green → Yellow: R 증가, G 유지(255), B=0
		t := blockPct * 2 // 0.0 ~ 1.0
		r = int(float64(255) * t)
		g = 255
		b = 0
	} else {
		// Yellow → Red: R 유지(255), G 감소, B=0
		t := (blockPct - 0.5) * 2 // 0.0 ~ 1.0
		r = 255
		g = int(float64(255) * (1.0 - t))
		b = 0
	}
	return
}

// BuildGradientBar 는 RGB 연속 그라디언트를 적용한 프로그레스 바를 생성한다.
//
// pct: 사용률 (0~100)
// width: 바의 총 블록 수
// noColor: true이면 ANSI 이스케이프 없이 유니코드 블록 문자만 출력
//
// 각 채워진 블록마다 개별 RGB 색상이 적용된다 (REQ-V3-BAR-002).
// noColor 또는 width <= 0 인 경우 단순 블록 문자열 반환 (REQ-V3-BAR-003).
//
// @MX:ANCHOR: 모든 사용량 바(CW/5H/7D) 렌더링의 핵심 함수 - renderer.go에서 호출됨
// @MX:REASON: fan_in >= 3 (renderUsageBar → 3개 바 렌더링 경로에서 호출)
func BuildGradientBar(pct int, width int, noColor bool) string {
	if width <= 0 {
		return ""
	}

	// 채워진 블록 수 계산 (최대 width)
	filled := min((pct*width)/100, width)
	empty := width - filled

	filledChar := "█" // 사용 중 블록
	emptyChar := "░"  // 남은 블록

	// noColor 모드 또는 채워진 블록이 없으면 단순 문자열 반환
	if noColor || filled == 0 {
		return strings.Repeat(filledChar, filled) + strings.Repeat(emptyChar, empty)
	}

	// 블록마다 개별 그라디언트 색상 적용
	var sb strings.Builder
	for i := 0; i < filled; i++ {
		// blockPct: 0.0 (첫 번째 블록) ~ 1.0 (마지막 블록)
		var blockPct float64
		if filled > 1 {
			blockPct = float64(i) / float64(filled-1)
		}
		r, g, b := interpolateGradientColor(blockPct)
		hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(filledChar))
	}
	sb.WriteString(strings.Repeat(emptyChar, empty))

	return sb.String()
}

// BatteryIcon 은 사용률에 따른 배터리 아이콘을 반환한다.
// 70% 이하: 🔋, 71% 이상: 🪫
//
// @MX:NOTE: 70% 임계값은 AC-V3-13 기준 - 변경 시 usage_test.go의 BatteryIcon 테스트도 업데이트할 것
func BatteryIcon(pct int) string {
	if pct > 70 {
		return "🪫"
	}
	return "🔋"
}
