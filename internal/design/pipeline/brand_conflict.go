// Package pipeline: /moai design 워크플로우 브랜드 충돌 검사기.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-05) 구현.
//
// BrandConflict는 `.moai/project/brand/visual-identity.md`에서 색상을 추출하고
// tokens.json의 color 토큰과 비교하여 충돌하는 경우 ValidationWarning을 반환한다.
//
// 서브에이전트 경계 원칙: 이 패키지는 경고 데이터를 반환할 뿐이며,
// AskUserQuestion을 직접 호출하지 않는다 (오케스트레이터의 책임).
package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// hexColorPattern: 마크다운에서 hex 색상 코드(#RGB, #RRGGBB) 추출용 정규식.
var hexColorPattern = regexp.MustCompile(`#(?:[0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})\b`)

// BrandColor: visual-identity.md에서 추출한 브랜드 색상 항목.
type BrandColor struct {
	// HexValue: 정규화된 소문자 hex 코드 (예: "#1d4ed8")
	HexValue string
	// SourceLine: 원본 마크다운 라인 (디버깅용)
	SourceLine string
}

// ExtractBrandColors: visual-identity.md 파일에서 hex 색상 코드를 추출한다.
//
// 파일이 없거나 _TBD_ 플레이스홀더만 있으면 빈 슬라이스를 반환한다.
// 파일 읽기 오류는 에러로 반환한다.
func ExtractBrandColors(visualIdentityPath string) ([]BrandColor, error) {
	data, err := os.ReadFile(visualIdentityPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일 없음: 브랜드 컨텍스트 미초기화 — 빈 결과 반환
			return nil, nil
		}
		return nil, fmt.Errorf("visual-identity.md 읽기 실패 (%s): %w", visualIdentityPath, err)
	}

	content := string(data)

	// _TBD_ 플레이스홀더만 있는 경우 (브랜드 인터뷰 미완료)
	if strings.Contains(content, "_TBD_") && !hexColorPattern.MatchString(content) {
		return nil, nil
	}

	var colors []BrandColor
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		matches := hexColorPattern.FindAllString(line, -1)
		for _, hex := range matches {
			colors = append(colors, BrandColor{
				HexValue:   strings.ToLower(hex),
				SourceLine: strings.TrimSpace(line),
			})
		}
	}

	return colors, scanner.Err()
}

// CheckBrandConflicts: tokens.json color 토큰과 브랜드 색상을 비교하여 충돌을 반환한다.
//
// tokens: Validate()에 전달하는 것과 동일한 map[string]any (tokens.json 파싱 결과)
// brandColors: ExtractBrandColors()로 추출한 브랜드 색상 목록
//
// 반환값: 충돌이 있는 경우 ValidationWarning 슬라이스 (없으면 빈 슬라이스)
//
// [HARD] 서브에이전트 경계 보존: 이 함수는 경고 데이터를 반환할 뿐이며,
// AskUserQuestion이나 사용자 입력을 요청하지 않는다.
//
// @MX:ANCHOR: [AUTO] 브랜드 충돌 검사 핵심 진입점 — orchestrator에서 호출
// @MX:REASON: REQ-DPL-009 brand context priority 구현; fan_in 증가 예상
func CheckBrandConflicts(tokens map[string]any, brandColors []BrandColor) []*dtcg.ValidationWarning {
	if len(brandColors) == 0 {
		return nil
	}

	// 브랜드 hex 값을 Set으로 구성 (O(1) 조회)
	brandHexSet := make(map[string]struct{}, len(brandColors))
	for _, bc := range brandColors {
		brandHexSet[bc.HexValue] = struct{}{}
	}

	var warnings []*dtcg.ValidationWarning

	for key, rawToken := range tokens {
		tokenDef, ok := rawToken.(map[string]any)
		if !ok {
			continue
		}

		typeVal, _ := tokenDef["$type"].(string)
		if typeVal != "color" {
			continue
		}

		rawValue, hasValue := tokenDef["$value"]
		if !hasValue {
			continue
		}

		tokenHex, ok := rawValue.(string)
		if !ok {
			continue
		}

		normalizedTokenHex := strings.ToLower(strings.TrimSpace(tokenHex))

		// 토큰 color 값이 브랜드 색상 셋에 없으면 충돌
		if _, inBrand := brandHexSet[normalizedTokenHex]; !inBrand {
			warnings = append(warnings, &dtcg.ValidationWarning{
				TokenPath: key,
				Category:  "brand-conflict",
				Message: fmt.Sprintf(
					"color 토큰 값 '%s'이 brand visual-identity.md에 없음 — 브랜드 색상을 사용하거나 인터뷰를 통해 추가하세요",
					normalizedTokenHex,
				),
			})
		}
	}

	return warnings
}

// RunBrandConflictCheck: visual-identity.md 경로와 tokens를 받아 원스텝으로 충돌을 검사한다.
//
// visualIdentityPath가 빈 문자열이거나 파일이 없으면 빈 결과를 반환한다.
// 이 함수는 ExtractBrandColors + CheckBrandConflicts를 합성하는 편의 함수이다.
func RunBrandConflictCheck(visualIdentityPath string, tokens map[string]any) ([]*dtcg.ValidationWarning, error) {
	brandColors, err := ExtractBrandColors(visualIdentityPath)
	if err != nil {
		return nil, err
	}
	return CheckBrandConflicts(tokens, brandColors), nil
}
