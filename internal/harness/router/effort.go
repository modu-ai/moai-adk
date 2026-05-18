package router

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// defaultEffortMapping은 EffortForLevel의 기본값 맵입니다.
// cfg.EffortMapping이 비어있거나 해당 레벨 키가 없을 때 사용됩니다.
// REQ-HRN-001-005: minimal→medium, standard→high, thorough→xhigh.
var defaultEffortMapping = map[Level]string{
	LevelMinimal:  "medium",
	LevelStandard: "high",
	LevelThorough: "xhigh",
}

// EffortForLevel은 harness 레벨에 맞는 노력 수준 문자열을 반환합니다.
// cfg.EffortMapping에서 해당 레벨 키를 읽고, 없으면 기본값을 반환합니다.
// REQ-HRN-001-005, AC-HRN-001-10.
//
// @MX:ANCHOR: [AUTO] EffortForLevel — CLI route --json의 effort 필드 소스
// @MX:REASON: fan_in >= 3: CLI route 명령, effort_test.go, harness router integration에서 호출
func EffortForLevel(level Level, cfg *config.HarnessConfig) string {
	if cfg != nil && len(cfg.EffortMapping) > 0 {
		if v, ok := cfg.EffortMapping[string(level)]; ok && v != "" {
			return v
		}
	}
	// 기본값 폴백
	if v, ok := defaultEffortMapping[level]; ok {
		return v
	}
	return "medium" // 최종 폴백
}

// EffortForLevelFromProxy는 ConfigProxy를 통해 노력 수준을 반환합니다.
// 테스트에서 경량 config 래퍼를 사용할 때 사용됩니다.
func EffortForLevelFromProxy(level Level, proxy *ConfigProxy) string {
	if proxy != nil && len(proxy.EffortMapping) > 0 {
		if v, ok := proxy.EffortMapping[string(level)]; ok && v != "" {
			return v
		}
	}
	// 기본값 폴백
	if v, ok := defaultEffortMapping[level]; ok {
		return v
	}
	return "medium"
}

// ParseProfileFloor는 evaluator profile .md 파일에서 최소 pass_threshold를 파싱합니다.
// REQ-HRN-001-012: pass_threshold >= 0.60 FROZEN floor 검증에 사용됩니다.
// 파일에서 "pass_threshold" 또는 테이블의 PassThreshold 값을 찾아 최솟값을 반환합니다.
func ParseProfileFloor(profilePath string) (float64, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return 0, fmt.Errorf("ParseProfileFloor: read %q: %w", profilePath, err)
	}

	content := string(data)
	minThreshold := 1.0
	found := false

	// "Pass Threshold" 테이블 컬럼에서 값 추출
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}

		cells := parseTableRowSimple(line)
		if len(cells) < 3 {
			continue
		}

		// 헤더/구분선 건너뜀
		header := strings.ToLower(cells[0])
		if strings.Contains(header, "dimension") || strings.Contains(cells[0], "---") {
			continue
		}

		// 세 번째 컬럼 (Pass Threshold) 파싱
		thresholdStr := strings.TrimSpace(cells[2])
		if strings.HasSuffix(thresholdStr, "%") {
			thresholdStr = strings.TrimSuffix(thresholdStr, "%")
			v, err := strconv.ParseFloat(strings.TrimSpace(thresholdStr), 64)
			if err == nil {
				v = v / 100.0
				if v < minThreshold {
					minThreshold = v
					found = true
				}
			}
		} else {
			v, err := strconv.ParseFloat(thresholdStr, 64)
			if err == nil && v > 0 {
				if v < minThreshold {
					minThreshold = v
					found = true
				}
			}
		}
	}

	if !found {
		// pass_threshold 테이블이 없으면 기본값 (FROZEN floor)
		return 0.60, nil
	}
	return minThreshold, nil
}

// parseTableRowSimple는 Markdown 테이블 행을 단순 파싱합니다.
func parseTableRowSimple(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

// Ensure config import is used.
var _ *config.HarnessConfig
