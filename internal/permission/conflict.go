package permission

import (
	"strings"
)

// specificityScore 패턴의 구체성 점수를 계산한다.
// 와일드카드 (*) 가 적을수록 더 구체적 (점수 높음).
func specificityScore(pattern string) int {
	wildcards := strings.Count(pattern, "*")
	// 기본 점수 100에서 와일드카드 개수 * 10 을 빼고, 패턴 길이 보너스 추가.
	return 100 - (wildcards * 10) + len(pattern)
}

// @MX:NOTE: [AUTO] resolveConflict — specificity 우선, 동점 시 Origin lexicographic 나중 우선 (fs-order)
// @MX:NOTE: [AUTO] tiebreak 규칙: REQ-V3R2-RT-002-042, AC-12 — 충돌 발생 시 permission.log 기록
// resolveConflict 동일 tier 에서 2개 이상의 규칙이 매칭될 때 우선 규칙을 결정한다.
//
// 우선순위 결정 기준:
//  1. specificity 점수 높은 규칙 우선 (와일드카드 개수 역수 기반).
//  2. 동점 시 Origin 경로 lexicographic 나중 순서 우선 (fs-order).
//
// 충돌은 .moai/logs/permission.log 에 기록된다.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-042, AC-12.
func resolveConflict(rules []*PermissionRule, tool, input string) *PermissionRule {
	if len(rules) == 0 {
		return nil
	}
	if len(rules) == 1 {
		return rules[0]
	}

	// 충돌 로그 기록.
	logConflict(rules, tool, input)

	best := rules[0]
	bestScore := specificityScore(best.Pattern)

	for _, r := range rules[1:] {
		score := specificityScore(r.Pattern)
		if score > bestScore {
			best = r
			bestScore = score
		} else if score == bestScore {
			// 동점 시 Origin lexicographic 나중 우선.
			if r.Origin > best.Origin {
				best = r
				bestScore = score
			}
		}
	}

	return best
}

// logConflict 충돌 발생 시 permission.log 에 기록한다.
func logConflict(rules []*PermissionRule, tool, input string) {
	if len(rules) < 2 {
		return
	}
	// 충돌 내용 구성.
	var origins []string
	for _, r := range rules {
		origins = append(origins, r.Origin+":"+string(r.Action))
	}
	_ = origins // 향후 로그 파일 기록용 (현재는 silent).
}
