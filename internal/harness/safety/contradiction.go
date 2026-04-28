// Package safety — Layer 3: Contradiction Detector (REQ-HL-008).
// trigger 키워드 중첩 및 chaining rules 모순을 탐지한다.
package safety

import (
	"fmt"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// SkillTriggers는 단일 skill의 trigger 목록이다.
// DetectOverlappingTriggers의 입력 타입으로 사용된다.
type SkillTriggers struct {
	// SkillPath는 skill 파일 경로이다.
	SkillPath string

	// Keywords는 이 skill의 trigger keyword 목록이다.
	Keywords []string
}

// DetectOverlappingTriggers는 여러 skill의 trigger keyword 중첩을 탐지한다.
// REQ-HL-008: 중첩된 keyword가 있으면 ContradictionReport에 기록한다.
//
// 같은 SkillPath 간의 중첩은 conflict로 간주하지 않는다.
//
// @MX:ANCHOR: [AUTO] DetectOverlappingTriggers는 trigger 중첩 탐지 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: contradiction_test.go, pipeline.go, Phase 4 coordinator
func DetectOverlappingTriggers(skillTriggers []SkillTriggers) harness.ContradictionReport {
	if len(skillTriggers) == 0 {
		return harness.ContradictionReport{}
	}

	// keyword → 해당 keyword를 사용하는 skill 경로 목록
	keywordToSkills := make(map[string][]string)

	for _, st := range skillTriggers {
		for _, kw := range st.Keywords {
			keywordToSkills[kw] = append(keywordToSkills[kw], st.SkillPath)
		}
	}

	var items []harness.ContradictionItem

	for kw, paths := range keywordToSkills {
		// 중복 경로 제거 (같은 skill이 여러 번 들어온 경우)
		uniquePaths := deduplicatePaths(paths)
		if len(uniquePaths) <= 1 {
			// 하나의 skill만 사용하면 conflict 없음
			continue
		}

		items = append(items, harness.ContradictionItem{
			Type: harness.ContradictionOverlappingTriggers,
			Description: fmt.Sprintf(
				"trigger keyword '%s'가 %d개 skill에서 중첩 사용됨: %v",
				kw, len(uniquePaths), uniquePaths,
			),
			ConflictingPaths:  uniquePaths,
			ConflictingValues: []string{kw},
		})
	}

	return harness.ContradictionReport{Items: items}
}

// deduplicatePaths는 경로 슬라이스에서 중복을 제거한다.
func deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	var result []string
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}

// DetectChainRuleContradictions는 기존 chaining rules와 제안된 rules 간의 모순을 탐지한다.
// REQ-HL-008: 같은 phase에 서로 다른 insert_before/insert_after가 있으면 모순으로 탐지한다.
//
// @MX:ANCHOR: [AUTO] DetectChainRuleContradictions는 chaining rule 모순 탐지 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: contradiction_test.go, pipeline.go, Phase 4 coordinator
func DetectChainRuleContradictions(existing, proposed harness.ChainingRules) harness.ContradictionReport {
	// 기존 rules를 phase → entry로 인덱싱
	existingByPhase := make(map[string]harness.ChainEntry, len(existing.Chains))
	for _, entry := range existing.Chains {
		existingByPhase[entry.Phase] = entry
	}

	var items []harness.ContradictionItem

	for _, proposedEntry := range proposed.Chains {
		existingEntry, ok := existingByPhase[proposedEntry.Phase]
		if !ok {
			// 기존에 없는 phase → conflict 없음
			continue
		}

		// 같은 phase에서 insert_before 충돌 검사
		if conflict := findChainConflict(existingEntry, proposedEntry); conflict != "" {
			items = append(items, harness.ContradictionItem{
				Type:              harness.ContradictionChainRules,
				Description:       conflict,
				ConflictingPaths:  []string{"chaining-rules.yaml"},
				ConflictingValues: []string{existingEntry.Phase},
			})
		}
	}

	return harness.ContradictionReport{Items: items}
}

// findChainConflict는 두 ChainEntry 간의 충돌을 탐지한다.
// 같은 phase에 서로 다른 agent가 insert_before에 있으면 충돌로 간주한다.
func findChainConflict(existing, proposed harness.ChainEntry) string {
	// insert_before 충돌: 기존과 제안이 모두 비어 있지 않고 다른 경우
	if len(existing.InsertBefore) > 0 && len(proposed.InsertBefore) > 0 {
		if !stringSlicesEqual(existing.InsertBefore, proposed.InsertBefore) {
			return fmt.Sprintf(
				"phase '%s'의 insert_before 충돌: 기존=%v, 제안=%v",
				existing.Phase, existing.InsertBefore, proposed.InsertBefore,
			)
		}
	}

	// insert_after 충돌: 기존과 제안이 모두 비어 있지 않고 다른 경우
	if len(existing.InsertAfter) > 0 && len(proposed.InsertAfter) > 0 {
		if !stringSlicesEqual(existing.InsertAfter, proposed.InsertAfter) {
			return fmt.Sprintf(
				"phase '%s'의 insert_after 충돌: 기존=%v, 제안=%v",
				existing.Phase, existing.InsertAfter, proposed.InsertAfter,
			)
		}
	}

	return ""
}

// stringSlicesEqual은 두 string 슬라이스가 순서까지 동일한지 확인한다.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
