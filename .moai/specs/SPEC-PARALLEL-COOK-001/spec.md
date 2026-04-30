---
id: SPEC-PARALLEL-COOK-001
status: draft
version: "0.1.0"
priority: Medium
labels: [parallel, sub-agent, cookbook, fan-out, fan-in, orchestration, wave-3, tier-2]
issue_number: null
scope: [.claude/rules/moai/development, .claude/rules/moai/workflow/team-pattern-cookbook.md, CLAUDE.md]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-PARALLEL-COOK-001: Parallel Sub-agent Cookbook

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic "Three subagents working in parallel complete in roughly the time one would take." 권고를 본 프로젝트 솔로 sub-agent 모드에 대한 표준 cookbook으로 흡수.

---

## 1. Goal (목적)

Team mode 외 솔로 sub-agent fan-out 패턴의 표준화된 cookbook 문서를 신설한다. 8가지 표준 페어 매트릭스, fan-in 책임 모델, 실패 격리 패턴, aggregation 거부 룰을 정의하여 ad-hoc 병렬 호출의 일관성과 안전성을 확보한다.

### 1.1 배경

- Anthropic blog "Subagents in Claude Code": "Three subagents working in parallel complete in roughly the time one would take."
- Anthropic blog "Multi-Agent Systems": "Aggregation strategy is the bottleneck of parallel agents."
- 본 프로젝트의 Team mode cookbook은 multi-agent coordination 한정 → solo orchestrator + N sub-agent fan-out 패턴 부재
- CLAUDE.md §14 "Parallel Execution Safeguards"가 일반 원칙 제공하나 8 페어 매트릭스 부재

### 1.2 비목표 (Non-Goals)

- Team mode cookbook 변경 (cross-ref만 추가)
- 새 agent 종류 신설 (기존 24 agent 카탈로그 활용만)
- Agent() spawn 자동화 도구 구현 (cookbook은 텍스트 가이드)
- 강제 enforcement (hook 차단 등)
- Team / Solo 모드 자동 전환 로직

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/rules/moai/development/parallel-subagent-patterns.md` 신규 작성
- 8가지 표준 페어 매트릭스 (manager-/expert-/builder- 조합)
- 페어별 fan-in 책임 명시 (orchestrator / reviewer / shared-state)
- 페어별 worktree isolation 가이드 (CLAUDE.md §14 정책 매핑)
- 실패 격리 코드 예시 (Promise.allSettled 패턴)
- aggregation 거부 anti-pattern 카탈로그 (>=3건)
- Team cookbook과 CLAUDE.md §14에 cross-ref 추가
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 새 agent 신설
- Team mode cookbook 의미 변경
- Agent() spawn 자동화 helper Go 코드
- hook 기반 cookbook 위반 enforcement
- 매트릭스 외 ad-hoc 페어 차단
- 페어 priority 자동 추천 알고리즘

---

## 3. Environment (환경)

- 런타임: orchestrator (MoAI main session)
- Claude Code v2.1.111+, Opus 4.7
- 영향 디렉터리: `.claude/rules/moai/development/`, `.claude/rules/moai/workflow/`
- 의존: 본 프로젝트 24 agent catalog (CLAUDE.md §4)

---

## 4. Assumptions (가정)

- A1: orchestrator는 `Agent()` 도구로 동시 다중 호출 가능 (Promise.all 또는 parallel tool block)
- A2: 8 표준 페어는 본 프로젝트 catalog의 agent 조합으로 구성 가능
- A3: cookbook은 living document로 분기별 검토
- A4: fan-in 책임 모델 3가지 (orchestrator / reviewer / shared-state)는 모든 표준 페어를 커버
- A5: 실패 격리는 텍스트 가이드 + 코드 예시로 충분 (자동화 불필요)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-PC-001**: THE COOKBOOK FILE `.claude/rules/moai/development/parallel-subagent-patterns.md` SHALL exist and document at minimum 8 standard sub-agent pairs.
- **REQ-PC-002**: THE COOKBOOK SHALL define three fan-in responsibility models: orchestrator-aggregates, reviewer-aggregates, shared-state-aggregates.
- **REQ-PC-003**: THE COOKBOOK SHALL include at least one code example for failure isolation using Promise.allSettled or equivalent pattern.
- **REQ-PC-004**: THE COOKBOOK SHALL include at least 3 anti-patterns where parallel fan-out should be rejected.

### 5.2 Event-Driven Requirements

- **REQ-PC-005**: WHEN orchestrator identifies 3 or more independent work units, THE ORCHESTRATOR SHALL consult parallel-subagent-patterns.md before initiating fan-out.
- **REQ-PC-006**: WHEN a fan-out invocation lacks a defined aggregation strategy, THE FAN-OUT SHALL be rejected with a blocker report referencing this cookbook.
- **REQ-PC-007**: WHEN a sub-agent in a fan-out fails, THE ORCHESTRATOR SHALL apply the failure-isolation pattern documented in this cookbook (rather than aborting all sibling agents).

### 5.3 State-Driven Requirements

- **REQ-PC-008**: WHILE fan-out execution is in progress, THE ORCHESTRATOR SHALL track per-agent status (pending/running/succeeded/failed) and surface aggregated status to the user.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-PC-009**: WHERE the standard pair involves write operations to shared files, THE COOKBOOK SHALL mandate worktree isolation for the write-side agent.
- **REQ-PC-010**: WHERE the standard pair is read-only on both sides, THE COOKBOOK SHALL prohibit worktree isolation (CLAUDE.md §14 alignment).
- **REQ-PC-011**: IF a new pair is discovered outside the 8 standard pairs, THEN THE NEW PAIR SHALL be added via PR with documented rationale and fan-in model.
- **REQ-PC-012**: WHERE Team mode is active, THE COOKBOOK SHALL defer to `team-pattern-cookbook.md` and document only solo-mode patterns.

### 5.5 Unwanted (Negative) Requirements

- **REQ-PC-013**: THE COOKBOOK SHALL NOT mandate Team mode for solo fan-out scenarios.
- **REQ-PC-014**: THE COOKBOOK SHALL NOT prescribe time estimates for parallel speedup (priority-based language only per CLAUDE.md §6 agent-common-protocol).
- **REQ-PC-015**: THE COOKBOOK SHALL NOT enforce automated rejection of unknown pairs (text guide only).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 8 표준 페어 정의 | 문서 카운트 | >= 8 |
| 페어별 fan-in 모델 | 페어별 명시 | 8/8 |
| 페어별 worktree 가이드 | 페어별 명시 | 8/8 |
| 실패 격리 코드 예시 | 코드 블록 | >= 1 |
| Anti-pattern 카탈로그 | anti-pattern entries | >= 3 |
| Cross-ref 갱신 | Team cookbook + CLAUDE.md §14 | 2/2 |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Team mode cookbook과의 의미 분리 명확 (solo orchestrator + N fan-out only)
- C2: 24 agent catalog 변경 없음
- C3: 모든 페어는 본 프로젝트의 실제 agent 조합 사용 (가공 페어 금지)
- C4: living document 마킹 — 분기별 review 의무 명시
- C5: Template-First Rule 준수 (`internal/template/templates/` 동기화)

End of spec.md (SPEC-PARALLEL-COOK-001 v0.1.0).
