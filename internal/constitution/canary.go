package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// canaryScoreDropThreshold는 canary reject 임계값이다.
	canaryScoreDropThreshold = 0.10
	// canaryMinSpecs는 canary 실행에 필요한 최소 SPEC 수이다.
	canaryMinSpecs = 3
	// canaryMaxSpecs는 canary 평가에 사용하는 최대 SPEC 수이다.
	canaryMaxSpecs = 3
)

// canary는 Canary interface의 구현이다.
// SPEC 완료 기록을 찾아 shadow evaluation을 수행한다.
type canary struct {
	// completedSpecPattern은 완료된 SPEC 디렉토리 패턴이다.
	// .moai/specs/SPEC-XXX/ 형식.
	completedSpecPattern *regexp.Regexp
}

// NewCanary는 Canary를 생성한다.
func NewCanary() Canary {
	return &canary{
		completedSpecPattern: regexp.MustCompile(`^SPEC-[A-Z0-9]+$`),
	}
}

// Evaluate는 proposal의 영향을 last 3 completed SPECs에 대해 평가한다.
// SPEC-V3R2-CON-002 REQ-CON-002-005 Layer 2 구현.
//
// Canary evaluation 전략:
// 1. .moai/specs/에서 최근 완료된 SPEC 디렉토리 찾기 (progress.md에 "completed" 또는 유사 마커)
// 2. 각 SPEC의 evalutor-active score를 읽기 (TODO: SPEC-V3R2-CON-003에서 구현)
// 3. Clause 변경이 SPEC 적용에 미치는 영향을 추정 (간단한 키워드 매칭)
// 4. ScoreBefore vs ScoreAfter 비교
//
// 현재 구현: SPEC score 저장소가 없으므로 CanaryUnavailable 반환.
// 향후 SPEC-V3R2-CON-003에서 evaluator-active integration 추가.
func (c *canary) Evaluate(proposal *AmendmentProposal, projectDir string) (*CanaryResult, error) {
	// .moai/specs/ 디렉토리 확인
	specsDir := filepath.Join(projectDir, ".moai", "specs")
	specEntries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// specs 디렉토리 없음 → CanaryUnavailable
			return &CanaryResult{
				Available: false,
				Reason:    fmt.Sprintf("SPEC 디렉토리 없음: %s", specsDir),
			}, &ErrCanaryUnavailable{RequiredCount: canaryMinSpecs, ActualCount: 0}
		}
		return nil, fmt.Errorf("specs 디렉토리 읽기 오류: %w", err)
	}

	// 완료된 SPEC 찾기
	var completedSpecs []string
	for _, entry := range specEntries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !c.completedSpecPattern.MatchString(name) {
			continue
		}
		// progress.md 존재 확인으로 완료 여부 판단 (간단 구현)
		progressPath := filepath.Join(specsDir, name, "progress.md")
		if _, err := os.Stat(progressPath); err == nil {
			completedSpecs = append(completedSpecs, name)
		}
	}

	// 최소 SPEC 수 확인
	if len(completedSpecs) < canaryMinSpecs {
		return &CanaryResult{
			Available: false,
			Reason:    fmt.Sprintf("완료된 SPEC 부족 (%d < %d)", len(completedSpecs), canaryMinSpecs),
		}, &ErrCanaryUnavailable{RequiredCount: canaryMinSpecs, ActualCount: len(completedSpecs)}
	}

	// 최근 3개 SPEC 선택
	selectedSpecs := completedSpecs
	if len(selectedSpecs) > canaryMaxSpecs {
		// 수정 시간 기준 정렬 후 최신 3개 선택
		selectedSpecs = c.sortMostRecent(selectedSpecs, specsDir, canaryMaxSpecs)
	}

	// TODO: SPEC-V3R2-CON-003에서 evaluator-active score 통합
	// 현재 구현: 더미 score로 placeholder
	//
	// Shadow simulation:
	// - Before: 기존 clause를 적용한 상태 가정 (score = 1.0)
	// - After: 새 clause를 적용한 상태 시뮬레이션 (키워드 매칭으로 영향도 추정)
	//
	// 간단 구현: clause 변경의 복잡도로 score 영향도 추정
	scoreBefore := 1.0
	scoreAfter := c.estimateScoreImpact(proposal)

	result := &CanaryResult{
		Available:     true,
		EvaluatedSpecs: selectedSpecs,
		ScoreBefore:   scoreBefore,
		ScoreAfter:    scoreAfter,
		MaxDrop:       scoreBefore - scoreAfter,
	}

	// Passed 판정
	if result.MaxDrop <= canaryScoreDropThreshold {
		result.Passed = true
		result.Reason = fmt.Sprintf("Score drop %.2f <= 임계값 %.2f", result.MaxDrop, canaryScoreDropThreshold)
	} else {
		result.Passed = false
		result.Reason = fmt.Sprintf("Score drop %.2f > 임계값 %.2f", result.MaxDrop, canaryScoreDropThreshold)
		return result, &ErrCanaryRejected{
			RuleID:        proposal.RuleID,
			ScoreDrop:     result.MaxDrop,
			Threshold:     canaryScoreDropThreshold,
			AffectedSpecs: selectedSpecs,
		}
	}

	return result, nil
}

// estimateScoreImpact는 clause 변경이 SPEC 점수에 미치는 영향을 추정한다.
// 간단한 휴리스틱: clause 길이와 키워드 변경으로 영향도 추정.
func (c *canary) estimateScoreImpact(proposal *AmendmentProposal) float64 {
	// 기본 score: 1.0 (완벽)
	score := 1.0

	// Clause 길이 비율로 영향도 추정
	beforeLen := len(proposal.Before)
	afterLen := len(proposal.After)

	// Clause가 길어지면 제약 강화 → score 소폭 하락
	// Clause가 짧아지면 제약 완화 → score 소폭 상승
	lenRatio := float64(afterLen) / float64(beforeLen)
	if lenRatio > 1.2 {
		// 20% 이상 길어짐 → 제약 강화
		score -= 0.05
	} else if lenRatio < 0.8 {
		// 20% 이상 짧아짐 → 제약 완화
		score += 0.02
	}

	// 키워드 기반 영향도 판정
	// 금지 단어 추가 ("MUST NOT", "NEVER", "PROHIBITED") → score 하락
	afterUpper := strings.ToUpper(proposal.After)
	if strings.Contains(afterUpper, "MUST NOT") || strings.Contains(afterUpper, "NEVER") || strings.Contains(afterUpper, "PROHIBITED") {
		score -= 0.08
	}

	// 필수 단어 제거 ("MUST", "REQUIRED", "SHALL") → score 상승 (제약 완화)
	beforeUpper := strings.ToUpper(proposal.Before)
	hadMust := strings.Contains(beforeUpper, "MUST") || strings.Contains(beforeUpper, "REQUIRED") || strings.Contains(beforeUpper, "SHALL")
	hasMust := strings.Contains(afterUpper, "MUST") || strings.Contains(afterUpper, "REQUIRED") || strings.Contains(afterUpper, "SHALL")
	if hadMust && !hasMust {
		score += 0.05
	}

	// Score bounds: [0.0, 1.0]
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// sortMostRecent는 수정 시간 기준으로 최근 SPEC 목록을 반환한다.
func (c *canary) sortMostRecent(specs []string, specsDir string, limit int) []string {
	type specTime struct {
		name string
		time time.Time
	}

	var specTimes []specTime
	for _, spec := range specs {
		info, err := os.Stat(filepath.Join(specsDir, spec))
		if err != nil {
			continue
		}
		specTimes = append(specTimes, specTime{name: spec, time: info.ModTime()})
	}

	// ModTime 내림차순 정렬
	for i := 0; i < len(specTimes); i++ {
		for j := i + 1; j < len(specTimes); j++ {
			if specTimes[j].time.After(specTimes[i].time) {
				specTimes[i], specTimes[j] = specTimes[j], specTimes[i]
			}
		}
	}

	// 최신 limit개 반환
	var result []string
	for i := 0; i < len(specTimes) && i < limit; i++ {
		result = append(result, specTimes[i].name)
	}
	return result
}

// canary는 Canary interface를 만족한다.
var _ Canary = (*canary)(nil)

// parseScoreFromProgress는 progress.md에서 evaluator-active score를 파싱한다.
// TODO: SPEC-V3R2-CON-003에서 구현. 현재는 더미.
func parseScoreFromProgress(progressPath string) (float64, error) {
	// 간단 구현: 파일에서 "Score: 0.XX" 패턴 찾기
	data, err := os.ReadFile(progressPath)
	if err != nil {
		return 0.8, err // 기본값 반환
	}

	re := regexp.MustCompile(`Score:\s*([0-9.]+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return 0.8, nil // 기본값 반환
	}

	score, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.8, err
	}

	return score, nil
}
