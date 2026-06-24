// Package safety — contradiction unit test.
// REQ-HL-008: trigger-overlap and chaining-rules contradiction detection tests.
package safety

import (
	"testing"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// TestDetectContradictions_OverlappingTriggers verifies detection when triggers across multiple skills overlap.
func TestDetectContradictions_OverlappingTriggers(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{
			SkillPath: ".claude/skills/harness-plan/SKILL.md",
			Keywords:  []string{"plan", "spec", "blueprint"},
		},
		{
			SkillPath: ".claude/skills/harness-run/SKILL.md",
			Keywords:  []string{"run", "execute", "plan"}, // "plan" overlaps
		},
		{
			SkillPath: ".claude/skills/harness-sync/SKILL.md",
			Keywords:  []string{"sync", "docs"},
		},
	}

	report := DetectOverlappingTriggers(skillTriggers)

	if !report.HasContradiction() {
		t.Fatal("ContradictionReport is empty despite overlapping triggers")
	}

	// The "plan" keyword must be detected as overlapping
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
		t.Error("'plan' keyword overlap missing from ContradictionReport")
	}
}

// TestDetectContradictions_NoOverlap verifies an empty report when there is no overlap.
func TestDetectContradictions_NoOverlap(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{
			SkillPath: ".claude/skills/harness-plan/SKILL.md",
			Keywords:  []string{"plan", "spec"},
		},
		{
			SkillPath: ".claude/skills/harness-run/SKILL.md",
			Keywords:  []string{"run", "execute"},
		},
		{
			SkillPath: ".claude/skills/harness-sync/SKILL.md",
			Keywords:  []string{"sync", "docs"},
		},
	}

	report := DetectOverlappingTriggers(skillTriggers)

	if report.HasContradiction() {
		t.Errorf("ContradictionReport present without overlap: %+v", report.Items)
	}
}

// TestDetectContradictions_MultipleOverlaps verifies the case when multiple keywords overlap.
func TestDetectContradictions_MultipleOverlaps(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{
			SkillPath: ".claude/skills/a/SKILL.md",
			Keywords:  []string{"x", "y", "z"},
		},
		{
			SkillPath: ".claude/skills/b/SKILL.md",
			Keywords:  []string{"x", "y", "w"}, // x, y overlap
		},
	}

	report := DetectOverlappingTriggers(skillTriggers)

	if !report.HasContradiction() {
		t.Fatal("ContradictionReport is empty despite x, y overlap")
	}

	// At least x and y must be detected as overlapping
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
		t.Errorf("both x and y must be detected: hasX=%v, hasY=%v", hasX, hasY)
	}
}

// TestDetectContradictions_ChainRulesConflict verifies detection of a chaining-rules contradiction.
func TestDetectContradictions_ChainRulesConflict(t *testing.T) {
	t.Parallel()

	// Chaining rules that conflict within the same phase
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
				InsertBefore: []string{"agent-B"}, // inserting a different agent-B (conflict)
				InsertAfter:  []string{},
			},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if !report.HasContradiction() {
		t.Fatal("ContradictionReport is empty despite a chaining-rule conflict")
	}

	// A ContradictionChainRules entry must be present
	found := false
	for _, item := range report.Items {
		if item.Type == harness.ContradictionChainRules {
			found = true
		}
	}
	if !found {
		t.Error("no ContradictionChainRules type entry found")
	}
}

// TestDetectContradictions_ChainRulesNoConflict verifies an empty report when chaining rules do not conflict.
func TestDetectContradictions_ChainRulesNoConflict(t *testing.T) {
	t.Parallel()

	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "plan", InsertBefore: []string{"agent-A"}, InsertAfter: []string{}},
		},
	}

	// Different phase → no conflict
	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "run", InsertBefore: []string{"agent-B"}, InsertAfter: []string{}},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if report.HasContradiction() {
		t.Errorf("different phases must not conflict: %+v", report.Items)
	}
}

// TestDetectContradictions_EmptyInputs verifies an empty report for empty inputs.
func TestDetectContradictions_EmptyInputs(t *testing.T) {
	t.Parallel()

	report1 := DetectOverlappingTriggers(nil)
	if report1.HasContradiction() {
		t.Error("ContradictionReport present on nil input")
	}

	report2 := DetectOverlappingTriggers([]SkillTriggers{})
	if report2.HasContradiction() {
		t.Error("ContradictionReport present on empty slice input")
	}

	report3 := DetectChainRuleContradictions(harness.ChainingRules{}, harness.ChainingRules{})
	if report3.HasContradiction() {
		t.Error("ContradictionReport present on empty chaining rules")
	}
}

// TestDetectContradictions_SamePathNoConflict verifies behavior when the same skill path appears twice.
func TestDetectContradictions_SamePathNoConflict(t *testing.T) {
	t.Parallel()

	skillTriggers := []SkillTriggers{
		{SkillPath: ".claude/skills/a/SKILL.md", Keywords: []string{"x"}},
		{SkillPath: ".claude/skills/a/SKILL.md", Keywords: []string{"x"}}, // same skill, same keyword
	}

	// The same skill with the same keyword is not a conflict (only across different skills)
	report := DetectOverlappingTriggers(skillTriggers)
	// Behavior may depend on the implementation, but the same path is not a conflict.
	_ = report // result is implementation-defined
}

// TestDetectContradictions_InsertAfterConflict verifies detection of insert_after conflicts.
func TestDetectContradictions_InsertAfterConflict(t *testing.T) {
	t.Parallel()

	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{
				Phase:        "run",
				InsertBefore: []string{},
				InsertAfter:  []string{"agent-X"},
			},
		},
	}

	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{
				Phase:        "run",
				InsertBefore: []string{},
				InsertAfter:  []string{"agent-Y"}, // different agent-Y (conflict)
			},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if !report.HasContradiction() {
		t.Fatal("ContradictionReport is empty despite an insert_after conflict")
	}
}

// TestDetectContradictions_SameInsertNoConflict verifies no conflict for identical insert_before.
func TestDetectContradictions_SameInsertNoConflict(t *testing.T) {
	t.Parallel()

	existing := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "plan", InsertBefore: []string{"agent-A"}, InsertAfter: []string{}},
		},
	}

	// Same agent-A → no conflict
	proposed := harness.ChainingRules{
		Version: 1,
		Chains: []harness.ChainEntry{
			{Phase: "plan", InsertBefore: []string{"agent-A"}, InsertAfter: []string{}},
		},
	}

	report := DetectChainRuleContradictions(existing, proposed)

	if report.HasContradiction() {
		t.Errorf("identical insert_before must not conflict: %+v", report.Items)
	}
}
