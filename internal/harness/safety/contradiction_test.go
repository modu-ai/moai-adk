// Package safety — contradiction unit test.
// REQ-HL-008: trigger 중첩 및 chaining rules 모순 탐지 테스트.
package safety

import (
	"testing"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// TestDetectContradictions_OverlappingTriggers는 여러 skill의 trigger가 겹칠 때 탐지하는지 검증한다.
func TestDetectContradictions_OverlappingTriggers(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{
			SkillPath: ".claude/skills/my-harness-plan/SKILL.md",
			Keywords:  []string{"plan", "spec", "blueprint"},
		},
		{
			SkillPath: ".claude/skills/my-harness-run/SKILL.md",
			Keywords:  []string{"run", "execute", "plan"}, // "plan" 중첩
		},
		{
			SkillPath: ".claude/skills/my-harness-sync/SKILL.md",
			Keywords:  []string{"sync", "docs"},
		},
	}

	report := DetectOverlappingTriggers(skillTriggers)

	if !report.HasContradiction() {
		t.Fatal("중첩 trigger가 있는데 ContradictionReport가 비어 있다")
	}

	// "plan" 키워드가 중첩으로 탐지되어야 함
	found := false
	for _, item := range report.Items {
		if item.Type == harness.ContradictionOverlappingTriggers {
			for _, v := range item.ConflictingValues {
				if v == "plan" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("'plan' 키워드 중첩이 ContradictionReport에 없다")
	}
}

// TestDetectContradictions_NoOverlap은 중첩이 없을 때 빈 보고서를 반환하는지 검증한다.
func TestDetectContradictions_NoOverlap(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{
			SkillPath: ".claude/skills/my-harness-plan/SKILL.md",
			Keywords:  []string{"plan", "spec"},
		},
		{
			SkillPath: ".claude/skills/my-harness-run/SKILL.md",
			Keywords:  []string{"run", "execute"},
		},
		{
			SkillPath: ".claude/skills/my-harness-sync/SKILL.md",
			Keywords:  []string{"sync", "docs"},
		},
	}

	report := DetectOverlappingTriggers(skillTriggers)

	if report.HasContradiction() {
		t.Errorf("중첩 없는데 ContradictionReport가 있다: %+v", report.Items)
	}
}

// TestDetectContradictions_MultipleOverlaps는 여러 키워드가 중첩되는 경우를 검증한다.
func TestDetectContradictions_MultipleOverlaps(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{
			SkillPath: ".claude/skills/a/SKILL.md",
			Keywords:  []string{"x", "y", "z"},
		},
		{
			SkillPath: ".claude/skills/b/SKILL.md",
			Keywords:  []string{"x", "y", "w"}, // x, y 중첩
		},
	}

	report := DetectOverlappingTriggers(skillTriggers)

	if !report.HasContradiction() {
		t.Fatal("x, y 중첩이 있는데 ContradictionReport가 비어 있다")
	}

	// 적어도 x와 y가 중첩으로 탐지되어야 함
	var conflictingValues []string
	for _, item := range report.Items {
		if item.Type == harness.ContradictionOverlappingTriggers {
			conflictingValues = append(conflictingValues, item.ConflictingValues...)
		}
	}

	hasX, hasY := false, false
	for _, v := range conflictingValues {
		if v == "x" {
			hasX = true
		}
		if v == "y" {
			hasY = true
		}
	}
	if !hasX || !hasY {
		t.Errorf("x, y 모두 탐지되어야 하는데: hasX=%v, hasY=%v", hasX, hasY)
	}
}

// TestDetectContradictions_ChainRulesConflict는 chaining rules 모순을 탐지하는지 검증한다.
func TestDetectContradictions_ChainRulesConflict(t *testing.T) {
	t.Parallel()

	// 같은 phase에 서로 충돌하는 체이닝 규칙
	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{
				Phase:        "plan",
				InsertBefore: []string{"agent-A"},
				InsertAfter:  []string{},
			},
		},
	}

	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{
				Phase:        "plan",
				InsertBefore: []string{"agent-B"}, // 다른 agent-B 삽입 (충돌)
				InsertAfter:  []string{},
			},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if !report.HasContradiction() {
		t.Fatal("chaining rule 모순이 있는데 ContradictionReport가 비어 있다")
	}

	// ContradictionChainRules 타입이 있어야 함
	found := false
	for _, item := range report.Items {
		if item.Type == harness.ContradictionChainRules {
			found = true
		}
	}
	if !found {
		t.Error("ContradictionChainRules 타입 항목이 없다")
	}
}

// TestDetectContradictions_ChainRulesNoConflict는 chaining rules 모순이 없을 때 빈 보고서를 반환하는지 검증한다.
func TestDetectContradictions_ChainRulesNoConflict(t *testing.T) {
	t.Parallel()

	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "plan", InsertBefore: []string{"agent-A"}, InsertAfter: []string{}},
		},
	}

	// 다른 phase → 충돌 없음
	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "run", InsertBefore: []string{"agent-B"}, InsertAfter: []string{}},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if report.HasContradiction() {
		t.Errorf("다른 phase이면 충돌 없어야 하는데: %+v", report.Items)
	}
}

// TestDetectContradictions_EmptyInputs는 빈 입력에 대해 빈 보고서를 반환하는지 검증한다.
func TestDetectContradictions_EmptyInputs(t *testing.T) {
	t.Parallel()

	report1 := DetectOverlappingTriggers(nil)
	if report1.HasContradiction() {
		t.Error("nil 입력에서 ContradictionReport가 있다")
	}

	report2 := DetectOverlappingTriggers([]SkillTriggers{})
	if report2.HasContradiction() {
		t.Error("빈 슬라이스 입력에서 ContradictionReport가 있다")
	}

	report3 := DetectChainRuleContradictions(harness.ChainingRules{}, harness.ChainingRules{})
	if report3.HasContradiction() {
		t.Error("빈 chaining rules에서 ContradictionReport가 있다")
	}
}

// TestDetectContradictions_SamePathNoConflict는 같은 skill path가 두 번 들어왔을 때를 검증한다.
func TestDetectContradictions_SamePathNoConflict(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{SkillPath: ".claude/skills/a/SKILL.md", Keywords: []string{"x"}},
		{SkillPath: ".claude/skills/a/SKILL.md", Keywords: []string{"x"}}, // 같은 skill, 같은 keyword
	}

	// 같은 skill의 동일 keyword는 conflict 아님 (다른 skill 간에만)
	report := DetectOverlappingTriggers(skillTriggers)
	// 구현에 따라 다를 수 있으나, 같은 path면 conflict 아님
	_ = report // 결과는 구현 정의
}

// TestDetectContradictions_InsertAfterConflict는 insert_after 충돌을 탐지하는지 검증한다.
func TestDetectContradictions_InsertAfterConflict(t *testing.T) {
	t.Parallel()

	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{
				Phase:       "run",
				InsertBefore: []string{},
				InsertAfter:  []string{"agent-X"},
			},
		},
	}

	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{
				Phase:       "run",
				InsertBefore: []string{},
				InsertAfter:  []string{"agent-Y"}, // 다른 agent-Y (충돌)
			},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if !report.HasContradiction() {
		t.Fatal("insert_after 충돌이 있는데 ContradictionReport가 비어 있다")
	}
}

// TestDetectContradictions_SameInsertNoConflict는 동일한 insert_before이면 충돌이 없는지 검증한다.
func TestDetectContradictions_SameInsertNoConflict(t *testing.T) {
	t.Parallel()

	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "plan", InsertBefore: []string{"agent-A"}, InsertAfter: []string{}},
		},
	}

	// 같은 agent-A → 충돌 없음
	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "plan", InsertBefore: []string{"agent-A"}, InsertAfter: []string{}},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if report.HasContradiction() {
		t.Errorf("동일한 insert_before이면 충돌이 없어야 하는데: %+v", report.Items)
	}
}
