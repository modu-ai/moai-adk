package observe

import "sort"

// PatternThresholds는 패턴 분류를 위한 관찰 횟수 임계값이다.
type PatternThresholds struct {
	// Heuristic은 휴리스틱 분류에 필요한 최소 관찰 횟수이다.
	Heuristic int
	// Rule은 규칙 분류에 필요한 최소 관찰 횟수이다.
	Rule int
	// HighConfidence는 높은 신뢰도 분류에 필요한 최소 관찰 횟수이다.
	HighConfidence int
}

// DefaultThresholds는 기본 패턴 분류 임계값을 반환한다.
func DefaultThresholds() PatternThresholds {
	return PatternThresholds{
		Heuristic:      3,
		Rule:           5,
		HighConfidence: 10,
	}
}

// PatternDetector는 관찰 목록에서 반복 패턴을 탐지한다.
type PatternDetector struct {
	thresholds PatternThresholds
}

// NewPatternDetector는 지정된 임계값으로 패턴 탐지기를 생성한다.
func NewPatternDetector(thresholds PatternThresholds) *PatternDetector {
	return &PatternDetector{thresholds: thresholds}
}

// Detect는 관찰 목록을 분석하여 패턴을 탐지하고 count 내림차순으로 반환한다.
// 관찰은 Agent:Target 키로 그룹화되며, 각 그룹의 관찰 횟수에 따라 분류된다.
func (d *PatternDetector) Detect(observations []*Observation) []*Pattern {
	if len(observations) == 0 {
		return nil
	}

	// Agent:Target 키로 그룹화
	groups := make(map[string][]*Observation)
	order := make([]string, 0) // 삽입 순서 유지 (안정 정렬용)
	for _, obs := range observations {
		key := obs.Agent + ":" + obs.Target
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], obs)
	}

	// 그룹별 패턴 생성
	patterns := make([]*Pattern, 0, len(groups))
	for _, key := range order {
		obs := groups[key]
		p := &Pattern{
			Key:          key,
			Count:        len(obs),
			Observations: obs,
		}

		// FirstSeen, LastSeen 계산
		p.FirstSeen = obs[0].Timestamp
		p.LastSeen = obs[0].Timestamp
		for _, o := range obs[1:] {
			if o.Timestamp.Before(p.FirstSeen) {
				p.FirstSeen = o.Timestamp
			}
			if o.Timestamp.After(p.LastSeen) {
				p.LastSeen = o.Timestamp
			}
		}

		// 임계값 기반 분류
		p.Classification = d.classify(p.Count)

		patterns = append(patterns, p)
	}

	// count 내림차순 정렬 (안정 정렬로 동일 count 시 삽입 순서 유지)
	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}

// classify는 관찰 횟수에 따른 패턴 분류를 반환한다.
func (d *PatternDetector) classify(count int) PatternClassification {
	switch {
	case count >= d.thresholds.HighConfidence:
		return ClassHighConfidence
	case count >= d.thresholds.Rule:
		return ClassRule
	case count >= d.thresholds.Heuristic:
		return ClassHeuristic
	default:
		return ClassObservation
	}
}
