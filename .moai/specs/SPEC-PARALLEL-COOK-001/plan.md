---
id: SPEC-PARALLEL-COOK-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-PARALLEL-COOK-001

## 1. Overview

Solo orchestrator + N sub-agent fan-out 패턴의 표준 cookbook 문서를 신설. 8 표준 페어 매트릭스, 3가지 fan-in 책임 모델, 5+ anti-pattern 카탈로그를 정의. 코드 변경 최소 (rules/development 디렉토리에 신규 markdown).

## 2. Approach Summary

**전략**: Documentation-First, Pattern-Catalog-Style.

1. 8 표준 페어 도출 — 본 프로젝트 24 agent catalog에서 자연스러운 동시 호출 조합
2. 페어별 fan-in 책임 명시 (orchestrator / reviewer / shared-state)
3. worktree isolation 적용 (CLAUDE.md §14 정책 매핑)
4. 실패 격리 코드 예시 + 5+ anti-pattern
5. CLAUDE.md §14 + Team cookbook cross-ref 갱신

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] 본 프로젝트 24 agent catalog 정확 확인 (CLAUDE.md §4)
- [ ] CLAUDE.md §14 worktree isolation 정책 verbatim 캡처
- [ ] Team cookbook (`.claude/rules/moai/workflow/team-pattern-cookbook.md`) 현재 내용 검토 → solo와의 boundary 명확화
- [ ] 본 프로젝트의 실제 fan-out 사례 5건 발굴 (SPEC PR 검색)

**Exit Criteria**: 8 페어 후보 도출 데이터 확보

### M1 — 8 표준 페어 매트릭스 작성 (Priority: High)

- [ ] 후보 페어 8개 정의 (예시):
  1. `expert-backend` + `expert-frontend` (independent feature dev)
  2. `manager-spec` + `manager-strategy` (planning + design parallel)
  3. `expert-testing` + `expert-debug` (test gen + repro analysis)
  4. `manager-docs` + `manager-quality` (sync + audit)
  5. `researcher` + `analyst` (research-only fan-out)
  6. `expert-security` + `expert-performance` (cross-cutting concerns)
  7. `expert-refactoring` + `manager-tdd` (refactor + test re-validation)
  8. `builder-skill` + `builder-agent` (builder fan-out)
- [ ] 각 페어의 input / output / fan-in agent / worktree 정책 명시
- [ ] 페어별 1-2 줄 use case 시나리오

**Exit Criteria**: 8 페어 매트릭스 표 완성

### M2 — 3가지 fan-in 책임 모델 정의 (Priority: High)

- [ ] **Model A — Orchestrator-aggregates**: orchestrator가 직접 결과 통합. 사용 시기 명시.
- [ ] **Model B — Reviewer-aggregates**: 별도 reviewer/aggregator agent 호출. 사용 시기 명시.
- [ ] **Model C — Shared-state-aggregates**: shared file (e.g., `.moai/specs/<ID>/progress.md`) 경유. 사용 시기 명시.
- [ ] 8 페어 각각이 어떤 모델 사용하는지 매핑

**Exit Criteria**: 3 모델 정의 + 페어별 매핑 표 완성

### M3 — 실패 격리 코드 예시 (Priority: Medium)

- [ ] Promise.allSettled 기반 코드 예시 (orchestrator pseudocode)
- [ ] 단일 페어 실패 시 다른 페어 보존 패턴
- [ ] partial-success aggregation 예시

**Exit Criteria**: 코드 블록 1+ 작성

### M4 — Anti-Pattern 카탈로그 (Priority: High)

- [ ] **AP-1** Aggregation 미정의 fan-out: aggregation strategy 없이 N개 spawn → reject
- [ ] **AP-2** Write 충돌 fan-out: 동일 파일에 N agent write → worktree isolation 미적용
- [ ] **AP-3** Sequential 작업 강제 병렬화: dependency 있는 작업을 병렬로 → race condition
- [ ] **AP-4** Read-only fan-out에 worktree 적용: research-only인데 worktree 격리 → over-engineering
- [ ] **AP-5** Team mode 잠재 작업을 solo로 강제: 3+ agent + cross-talk 필요인데 solo로 → Team mode 권장

**Exit Criteria**: 5 anti-pattern 명시 + 각각 reject 사유

### M5 — Cookbook 본문 작성 (Priority: High)

- [ ] `.claude/rules/moai/development/parallel-subagent-patterns.md` 신규 작성
  - Front matter (없음 — rule 파일은 frontmatter 미사용)
  - §1 Overview + scope
  - §2 8 표준 페어 매트릭스
  - §3 3 fan-in 책임 모델
  - §4 Worktree 적용 가이드
  - §5 실패 격리 코드 예시
  - §6 Anti-Pattern 카탈로그
  - §7 신규 페어 추가 절차
- [ ] living document 마킹 + 분기별 review 명시

**Exit Criteria**: cookbook 파일 작성 완료

### M6 — Cross-reference 갱신 (Priority: High)

- [ ] `CLAUDE.md` §14 "Parallel Execution Safeguards"에 추가:
  > "구체 페어 사례는 .claude/rules/moai/development/parallel-subagent-patterns.md 참조"
- [ ] `.claude/rules/moai/workflow/team-pattern-cookbook.md` 도입부에 추가:
  > "솔로 패턴은 .claude/rules/moai/development/parallel-subagent-patterns.md 참조"
- [ ] Template-First 동기화 (`internal/template/templates/CLAUDE.md`, `internal/template/templates/.claude/rules/...`)

**Exit Criteria**: 양방향 cross-ref 검증

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md Given-When-Then 시나리오 모두 PASS
- [ ] cookbook 파일이 `.claude/rules/moai/development/`에 정확 존재
- [ ] 8 페어 / 3 fan-in / 5+ anti-pattern 카운트 검증
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 페어 매트릭스 표 형식 (예시)

```markdown
| # | Pair | Input | Output | Fan-in Agent | Worktree | Use Case |
|---|------|-------|--------|--------------|----------|----------|
| 1 | expert-backend × expert-frontend | SPEC-XXX | code (api/, ui/) | orchestrator | both worktree | full-stack feature |
| 2 | manager-spec × manager-strategy | feature request | spec.md, design doc | orchestrator | neither | parallel planning |
... |
```

### 4.2 Anti-Pattern entry 형식 (예시)

```markdown
### AP-1: Aggregation undefined

**Symptom**: orchestrator spawns N sub-agents in parallel without specifying who aggregates the results.

**Why bad**: Each sub-agent returns independently; orchestrator ad-hoc merges, leading to inconsistent fan-in.

**Mitigation**: Before fan-out, orchestrator MUST select one of the 3 fan-in models (A/B/C). If none applies, reject the fan-out and re-plan.
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| 8 페어가 prescriptive하여 신규 페어 발굴 저해 | Medium | Medium | M5에 "신규 페어 추가 절차" 명시, living document |
| Team cookbook과 의미 충돌 | Low | High | M0에서 boundary 명확화, M6에서 양방향 cross-ref |
| 실제 사용 사례 부족 → 매트릭스가 가공 | Medium | Medium | M0에서 5건 실제 사례 수집 |
| anti-pattern이 모호하여 위반 판정 어려움 | Medium | Low | M4에서 각 AP에 구체 예시 + reject 사유 명시 |
| Template-First 동기화 누락 | Low | Medium | M6 checklist에 명시 |

## 6. Dependencies

- 선행 SPEC: 없음 (독립)
- 의존 정책: CLAUDE.md §14 worktree isolation rules
- 의존 cookbook: Team mode cookbook (cross-ref용)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (신규 페어 등재 절차): 본 SPEC plan-auditor 통과 후 PR로 cookbook 확장 — M5에 명시
- **OQ2** (fan-in default): orchestrator-aggregates를 default로 권장 (가장 단순, 가장 흔함)
- **OQ3** (위반 enforcement): 텍스트 가이드 only, hook 차단 별도 SPEC 후보로 분리
- **OQ4** (expert-debug 페어 포함 여부): debug는 본질적으로 sequential일 가능성 높음 → 표준 페어 후보에서 제외, anti-pattern AP-3에 사례

## 8. Rollout Plan

1. M5 cookbook 작성 후 본 프로젝트의 다음 fan-out PR에 적용 (dogfooding)
2. CHANGELOG 명시 후 v2.x.0 minor release
3. 분기별 review에서 페어 추가 / anti-pattern 보강

End of plan.md (SPEC-PARALLEL-COOK-001).
