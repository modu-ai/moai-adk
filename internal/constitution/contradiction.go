package constitution

import (
	"fmt"
	"regexp"
	"strings"
)

// contradictionDetector는 ContradictionDetector interface의 구현이다.
// Rule 간 모순을 탐지한다.
type contradictionDetector struct{}

// NewContradictionDetector는 ContradictionDetector를 생성한다.
func NewContradictionDetector() ContradictionDetector {
	return &contradictionDetector{}
}

// Scan은 proposal이 registry의 다른 rule과 모순하는지 확인한다.
// SPEC-V3R2-CON-002 REQ-CON-002-006 Layer 3 구현.
//
// 모순 탐지 전략:
// 1. ID 충돌: 동일한 ID 재사용 → 차단
// 2. Zone 모순: Frozen→Evolvable demotion 없이 Frozen 제약 완화 → 차단
// 3. Clause 모순: "MUST X"와 "MUST NOT X" 동시 존재 → 차단
// 4. 의미 모순: 반대 의미 키워드 쌍 → 경고
func (d *contradictionDetector) Scan(proposal *AmendmentProposal, registry *Registry) (*ContradictionResult, error) {
	result := &ContradictionResult{
		Conflicts: []ConflictDetail{},
	}

	// Registry의 모든 rule과 비교
	for _, rule := range registry.Entries {
		// 자기 자신은 스킵
		if rule.ID == proposal.RuleID {
			continue
		}

		// ID 충돌 (다른 rule이 같은 ID 사용 - 중복 불가)
		// 이는 loader에서 이미 차단되므로 여기서는 추가 체크 안 함

		// Zone 모순 탐지
		if conflict := d.detectZoneContradiction(proposal, &rule); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			if conflict.IsBlocking {
				result.HasBlockingContradiction = true
			}
		}

		// Clause 모순 탐지
		if conflict := d.detectClauseContradiction(proposal, &rule); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			if conflict.IsBlocking {
				result.HasBlockingContradiction = true
			}
		}
	}

	// 모순이 있으면 에러 반환
	if result.HasBlockingContradiction {
		conflictingIDs := make([]string, 0, len(result.Conflicts))
		descriptions := make([]string, 0, len(result.Conflicts))
		for _, c := range result.Conflicts {
			conflictingIDs = append(conflictingIDs, c.ConflictingRuleID)
			descriptions = append(descriptions, c.Description)
		}
		return result, &ErrContradictionDetected{
			NewRuleID:      proposal.RuleID,
			ConflictingIDs: conflictingIDs,
			Conflicts:      descriptions,
		}
	}

	return result, nil
}

// detectZoneContradiction은 zone 변경 모순을 탐지한다.
func (d *contradictionDetector) detectZoneContradiction(proposal *AmendmentProposal, rule *Rule) *ConflictDetail {
	// proposal의 After clause에서 zone 힌트 추정
	afterUpper := strings.ToUpper(proposal.After)

	// Frozen 제약 완화 감지: "MUST" 제거, "MAY" 또는 "SHOULD" 추가
	beforeUpper := strings.ToUpper(proposal.Before)
	hadMust := strings.Contains(beforeUpper, "MUST") || strings.Contains(beforeUpper, "SHALL") || strings.Contains(beforeUpper, "REQUIRED")
	lostMust := hadMust && !strings.Contains(afterUpper, "MUST") && !strings.Contains(afterUpper, "SHALL") && !strings.Contains(afterUpper, "REQUIRED")
	addedMay := strings.Contains(afterUpper, "MAY") || strings.Contains(afterUpper, "SHOULD") || strings.Contains(afterUpper, "OPTIONAL")

	if lostMust && addedMay {
		return &ConflictDetail{
			ConflictingRuleID: rule.ID,
			Description:       fmt.Sprintf("제약 완화: %s의 강제 조항이 권장 사항으로 변경됨", rule.ID),
			IsBlocking:        true, // Frozen zone은 제약 완화 차단
		}
	}

	return nil
}

// detectClauseContradiction은 clause 내용 모순을 탐지한다.
func (d *contradictionDetector) detectClauseContradiction(proposal *AmendmentProposal, rule *Rule) *ConflictDetail {
	// Proposal의 After clause와 기존 rule의 clause 비교

	// 1. 반대 의미 키워드 쌍 탐지
	afterWords := strings.Fields(strings.ToUpper(proposal.After))
	ruleWords := strings.Fields(strings.ToUpper(rule.Clause))

	// "MUST NOT" + "MUST" 조합 탐지
	proposalMustNot := containsSequence(afterWords, "MUST", "NOT")
	ruleMust := containsWord(ruleWords, "MUST") || containsWord(ruleWords, "SHALL") || containsWord(ruleWords, "REQUIRED")

	if proposalMustNot && ruleMust {
		return &ConflictDetail{
			ConflictingRuleID: rule.ID,
			Description:       fmt.Sprintf("의미 모순: %s은(는) 요구사항인데 proposal은 'MUST NOT' 금지를 추가", rule.ID),
			IsBlocking:        true,
		}
	}

	// 2. 대립되는 행동 패턴 탐지
	// 예: "always X" vs "never X"
	if extractAction(proposal.After) != "" && extractAction(proposal.After) == extractAction(rule.Clause) {
		proposalMod := extractModifier(proposal.After)
		ruleMod := extractModifier(rule.Clause)

		if isOppositeModifier(proposalMod, ruleMod) {
			return &ConflictDetail{
				ConflictingRuleID: rule.ID,
				Description:       fmt.Sprintf("행동 모순: %s('%s') vs proposal('%s')", rule.ID, ruleMod, proposalMod),
				IsBlocking:        true,
			}
		}
	}

	return nil
}

// containsSequence는 슬라이스에서 연속 단어 시퀀스를 찾는다.
func containsSequence(words []string, seq ...string) bool {
	for i := 0; i <= len(words)-len(seq); i++ {
		match := true
		for j, word := range seq {
			if i+j >= len(words) || words[i+j] != word {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// containsWord는 슬라이스에서 단어를 찾는다.
func containsWord(words []string, target string) bool {
	for _, w := range words {
		if w == target {
			return true
		}
	}
	return false
}

// extractAction은 clause에서 동작/대상을 추출한다.
// 간단 구현: 따옴표 안의 내용 또는 첫 명사구.
func extractAction(clause string) string {
	// 따옴표 안의 내용 추출
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindStringSubmatch(clause)
	if len(matches) >= 2 {
		return matches[1]
	}
	// TODO: 더 정교한 NLP 기반 추정 (SPEC-V3R2-CON-003)
	return ""
}

// extractModifier는 clause에서 수식어(MUST/SHOULD/MAY/NOT 등)를 추출한다.
func extractModifier(clause string) string {
	upper := strings.ToUpper(clause)
	modifiers := []string{"MUST NOT", "MUST", "SHALL NOT", "SHALL", "REQUIRED", "SHOULD", "MAY", "OPTIONAL"}
	for _, mod := range modifiers {
		if strings.Contains(upper, mod) {
			return mod
		}
	}
	return ""
}

// isOppositeModifier는 두 수식어가 대립 관계인지 판단한다.
func isOppositeModifier(mod1, mod2 string) bool {
	opposites := map[string][]string{
		"MUST NOT": {"MUST", "SHALL", "REQUIRED"},
		"MUST":     {"MUST NOT", "NEVER"},
		"SHALL":    {"SHALL NOT", "MUST NOT"},
		"NEVER":    {"MUST", "SHALL", "REQUIRED", "ALWAYS"},
		"ALWAYS":   {"NEVER", "MUST NOT"},
	}

	for _, opposite := range opposites[mod1] {
		if opposite == mod2 {
			return true
		}
	}
	return false
}

// contradictionDetector는 ContradictionDetector interface를 만족한다.
var _ ContradictionDetector = (*contradictionDetector)(nil)
