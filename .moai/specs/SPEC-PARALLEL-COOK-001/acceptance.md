---
id: SPEC-PARALLEL-COOK-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-PARALLEL-COOK-001

## Given-When-Then Scenarios

### Scenario 1: Cookbook 파일 존재

**Given** the SPEC-PARALLEL-COOK-001 implementation completes
**And** Template-First sync runs

**When** the user inspects the project's rules directory

**Then** the file `.claude/rules/moai/development/parallel-subagent-patterns.md` SHALL exist
**And** the corresponding template file at `internal/template/templates/.claude/rules/moai/development/parallel-subagent-patterns.md` SHALL exist
**And** both files SHALL be identical (same hash)

---

### Scenario 2: 8 표준 페어 매트릭스 정의

**Given** the cookbook file exists

**When** the user reads the "Standard Pair Matrix" section

**Then** the matrix SHALL contain at least 8 rows
**And** each row SHALL include columns: pair name, input, output, fan-in agent, worktree policy, use case
**And** each pair SHALL be composed of agents from the project's 24-agent catalog

---

### Scenario 3: 3가지 fan-in 책임 모델

**Given** the cookbook file exists

**When** the user reads the "Fan-in Responsibility Models" section

**Then** the section SHALL define exactly 3 models: Orchestrator-aggregates, Reviewer-aggregates, Shared-state-aggregates
**And** each model SHALL include "When to use" guidance
**And** each of the 8 standard pairs SHALL be mapped to one of the 3 models

---

### Scenario 4: 실패 격리 코드 예시

**Given** the cookbook file exists

**When** the user reads the "Failure Isolation" section

**Then** at least one code example using Promise.allSettled or equivalent SHALL be present
**And** the example SHALL demonstrate single-pair failure not aborting sibling pairs
**And** the example SHALL show partial-success aggregation

---

### Scenario 5: Anti-Patterns 카탈로그 (5+)

**Given** the cookbook file exists

**When** the user reads the "Anti-Patterns" section

**Then** the section SHALL contain at least 5 anti-pattern entries
**And** each entry SHALL have: Symptom, Why bad, Mitigation
**And** the entries SHALL include AP-1 (Aggregation undefined), AP-2 (Write conflict fan-out), and at least 3 more

---

### Scenario 6: Aggregation undefined → fan-out reject

**Given** the orchestrator identifies 4 independent work units
**And** the orchestrator has not selected a fan-in model

**When** the orchestrator attempts to spawn 4 parallel sub-agents

**Then** per the cookbook AP-1, the orchestrator SHALL halt fan-out
**And** the orchestrator SHALL emit a blocker report referencing the cookbook
**And** the orchestrator SHALL re-plan with explicit fan-in model selection

---

### Scenario 7: Worktree isolation per pair

**Given** the cookbook file exists

**When** the user reads each of the 8 pair definitions

**Then** each pair SHALL specify worktree isolation policy: "both worktree", "neither worktree", "left worktree", or "right worktree"
**And** read-only pairs (researcher, analyst, reviewer combinations) SHALL specify "neither worktree" per CLAUDE.md §14
**And** write-side pairs SHALL specify "worktree" for the writing agent

---

### Scenario 8: Cross-reference 양방향 검증

**Given** the cookbook file exists
**And** Template-First sync ran

**When** the user inspects CLAUDE.md §14

**Then** CLAUDE.md §14 SHALL contain a reference to `.claude/rules/moai/development/parallel-subagent-patterns.md`
**And** `team-pattern-cookbook.md` introduction SHALL contain a reference back to the new cookbook
**And** the new cookbook SHALL contain references to CLAUDE.md §14 and team-pattern-cookbook.md (mutual)

---

## Edge Cases

### EC-1: Pair not in matrix
If an orchestrator wishes to use a pair not in the standard 8, the cookbook §"New Pair Addition Procedure" SHALL describe the PR-based extension process. Ad-hoc pairs are permitted with documented rationale.

### EC-2: Team mode active
If Team mode (`workflow.team.enabled: true`) is active, the cookbook SHALL defer to `team-pattern-cookbook.md`. The new cookbook documents only solo orchestrator + N fan-out patterns.

### EC-3: Single sub-agent (not fan-out)
If the orchestrator delegates to a single sub-agent (no fan-out), this cookbook does not apply. The cookbook scope is fan-out (3+ agents) only.

### EC-4: Sequential dependency hidden in "parallel" call
If an orchestrator spawns N agents in parallel where some agents have implicit dependencies (e.g., agent B reads file written by agent A), this is AP-3 anti-pattern. The cookbook MUST flag this in code review.

### EC-5: Living document updates
The cookbook is marked as living document. New anti-patterns and pairs SHALL be added via PR with rationale. Quarterly review process documented.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Cookbook file exists | both local + template | file existence check |
| Standard pair count | >= 8 | matrix row count |
| Fan-in models | exactly 3 | section count |
| Pair-to-model mapping | 8/8 | each pair tagged |
| Anti-patterns | >= 5 | entry count |
| Code example | >= 1 failure isolation | code block exists |
| Cross-references | 3-way (CLAUDE.md + Team cookbook + new cookbook) | cross-ref scan |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented
- [ ] All 9 quality gate criteria meet threshold
- [ ] `.claude/rules/moai/development/parallel-subagent-patterns.md` exists with 8 pairs, 3 models, 5+ anti-patterns
- [ ] Template-First sync at `internal/template/templates/.claude/rules/moai/development/parallel-subagent-patterns.md`
- [ ] CLAUDE.md §14 cross-reference added
- [ ] `team-pattern-cookbook.md` cross-reference added
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] living document marker present (HISTORY section + quarterly review note)
- [ ] plan-auditor PASS
- [ ] No code change in Go source files (documentation-only SPEC verified by `git diff`)
- [ ] dogfooding: at least one PR after merge uses the cookbook for fan-out planning

End of acceptance.md (SPEC-PARALLEL-COOK-001).
