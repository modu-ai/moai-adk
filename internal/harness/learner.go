// Package harness — 패턴 집계 및 tier 분류기.
// REQ-HL-002: usage-log.jsonl 이벤트를 읽어 패턴으로 집계하고 tier를 분류한다.
package harness

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// confidenceThreshold는 tier 승격을 허용하는 최소 신뢰도이다.
// REQ-HL-002: 0.70 미만이면 count에 관계없이 TierObservation으로 강제된다.
const confidenceThreshold = 0.70

// Learner는 usage-log.jsonl을 읽어 패턴을 집계하고 tier를 분류한다.
// WritePromotion은 tier-promotions.jsonl에 promotion 이벤트를 기록한다.
//
// @MX:ANCHOR: [AUTO] AggregatePatterns, ClassifyTier, WritePromotion은 학습 파이프라인 진입점.
// @MX:REASON: [AUTO] fan_in >= 3: learner_test.go, safety.go(Phase 3), applier.go
type Learner struct {
	// promotionPath는 tier-promotions.jsonl 파일 경로이다.
	promotionPath string

	// nowFn은 현재 시각을 반환하는 함수 (테스트에서 override 가능).
	nowFn func() time.Time
}

// NewLearner는 지정된 promotionPath를 사용하는 Learner를 생성한다.
func NewLearner(promotionPath string) *Learner {
	return &Learner{
		promotionPath: promotionPath,
		nowFn:         time.Now,
	}
}

// AggregatePatterns는 logPath의 JSONL 로그를 읽어 (event_type, subject, context_hash)
// 조합으로 그룹핑한 패턴 맵을 반환한다.
// REQ-HL-002: 파일이 없거나 비어 있으면 빈 맵을 반환한다.
//
// @MX:ANCHOR: [AUTO] 학습 파이프라인의 첫 번째 단계로 다수 호출자에서 사용.
// @MX:REASON: [AUTO] fan_in >= 3: learner_test.go, CLI harness status, Phase 5 IT-01
func AggregatePatterns(logPath string) (map[string]*Pattern, error) {
	patterns := make(map[string]*Pattern)

	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일 없음은 정상 상태 — 빈 맵 반환
			return patterns, nil
		}
		return nil, fmt.Errorf("learner: 로그 파일 열기 실패 %s: %w", logPath, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			// 파싱 실패 줄은 건너뜀 (데이터 손실 방지)
			continue
		}

		key := buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash)
		if p, ok := patterns[key]; ok {
			p.Count++
		} else {
			patterns[key] = &Pattern{
				Key:         key,
				EventType:   evt.EventType,
				Subject:     evt.Subject,
				ContextHash: evt.ContextHash,
				Count:       1,
				Confidence:  defaultConfidence,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("learner: 로그 파일 스캔 오류: %w", err)
	}

	return patterns, nil
}

// defaultConfidence는 초기 신뢰도 기본값이다.
// Phase 3 안전 레이어에서 canary check를 통해 갱신될 예정.
const defaultConfidence = 1.0

// buildPatternKey는 (event_type, subject, context_hash) 조합으로 고유 키를 생성한다.
func buildPatternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}

// ClassifyTier는 Pattern의 count와 confidence를 기준으로 Tier를 반환한다.
// REQ-HL-002:
//   - confidence < 0.70이면 count에 관계없이 TierObservation 반환.
//   - count가 thresholds[3](10 이상)이면 TierAutoUpdate.
//   - count가 thresholds[2](5 이상)이면 TierRule.
//   - count가 thresholds[1](3 이상)이면 TierHeuristic.
//   - count가 thresholds[0](1 이상)이면 TierObservation.
//
// thresholds는 [1, 3, 5, 10] 형식이어야 한다 (plan.md §4.3).
// thresholds가 비어 있으면 TierObservation을 반환한다.
func ClassifyTier(p *Pattern, thresholds []int) Tier {
	// 신뢰도 미달: count에 관계없이 Observation 강제
	if p.Confidence < confidenceThreshold {
		return TierObservation
	}

	if len(thresholds) == 0 {
		return TierObservation
	}

	// thresholds = [t0, t1, t2, t3] = [1, 3, 5, 10]
	// count >= t3 → AutoUpdate
	// count >= t2 → Rule
	// count >= t1 → Heuristic
	// count >= t0 → Observation
	// count < t0  → Observation (관찰 전)
	if len(thresholds) >= 4 && p.Count >= thresholds[3] {
		return TierAutoUpdate
	}
	if len(thresholds) >= 3 && p.Count >= thresholds[2] {
		return TierRule
	}
	if len(thresholds) >= 2 && p.Count >= thresholds[1] {
		return TierHeuristic
	}
	// count >= thresholds[0] 또는 count < thresholds[0] 모두 Observation
	return TierObservation
}

// WritePromotion은 Promotion 이벤트를 tier-promotions.jsonl에 append한다.
// REQ-HL-002: plan.md §4.2 스키마에 따라 기록한다.
// 부모 디렉토리가 없으면 자동으로 생성한다.
func (l *Learner) WritePromotion(p Promotion) error {
	// 부모 디렉토리 자동 생성
	dir := filepath.Dir(l.promotionPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("learner: promotion 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// Ts가 비어 있으면 현재 시각으로 설정
	if p.Ts.IsZero() {
		p.Ts = l.nowFn().UTC()
	}

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("learner: promotion 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	f, err := os.OpenFile(l.promotionPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("learner: promotion 파일 열기 실패 %s: %w", l.promotionPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("learner: promotion 파일 쓰기 실패: %w", err)
	}

	return nil
}
