---
id: SPEC-CONTEXT-INJ-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-CONTEXT-INJ-001

## 1. Overview

orchestrator의 sub-agent spawn prompt 작성 시 context 명시 주입 정책 표준화. 5KB cap + 3-tier 우선순위 (progress.md > MEMORY.md > domain lessons). 코드 변경 없이 정책 문서 + cross-ref만으로 일관성 확보.

## 2. Approach Summary

**전략**: Policy-First, Documentation-Driven, Marker-Based-Separation.

1. `.claude/rules/moai/development/context-injection.md` 정책 문서 신규 작성
2. progress.md 권장 schema 명시
3. orchestrator-callable agent (manager-*, expert-*) body에 cross-ref
4. moai-foundation-core SKILL.md Token Budget 절 보강
5. Template-First 동기화

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] 본 프로젝트의 현재 sub-agent 호출 사례 5건 분석 (orchestrator → manager-* / expert-*)
- [ ] 어떤 context가 주입되고 있는지 / 안 되고 있는지 매핑
- [ ] 5KB cap의 적정성 검토 (sub-agent 평균 초기 토큰 measure)
- [ ] progress.md 작성 사례 5건 분석 → 권장 schema 도출
- [ ] `~/.claude/projects/<hash>/memory/MEMORY.md` 형식 확인

**Exit Criteria**: 정책 도출 baseline 데이터 확보

### M1 — Policy Document 작성 (Priority: High)

- [ ] `.claude/rules/moai/development/context-injection.md` 신규 작성
  - §1 Overview + scope (orchestrator-only, sub-agent receivers)
  - §2 5KB Token Budget Cap 명시
  - §3 3-Tier Priority Order (progress > MEMORY > domain lessons)
  - §4 Marker Convention (`<!-- injected-context -->` ... `<!-- /injected-context -->`)
  - §5 progress.md Recommended Schema (권장 수준)
  - §6 Truncation Strategy (cap 초과 시)
  - §7 Anti-Patterns (raw secrets, cross-agent private, etc.)
  - §8 Exception Cases (research-only sub-agent 등)
- [ ] 정책 문서 1.5-2KB 범위로 압축

**Exit Criteria**: 정책 문서 완성 + 8 절 모두 작성

### M2 — Cross-reference 갱신 (Priority: High)

- [ ] orchestrator-callable agent body에 cross-ref 추가:
  - `.claude/agents/moai/manager-spec.md`
  - `.claude/agents/moai/manager-ddd.md`
  - `.claude/agents/moai/manager-tdd.md`
  - `.claude/agents/moai/manager-docs.md`
  - `.claude/agents/moai/manager-quality.md`
  - `.claude/agents/moai/manager-strategy.md`
  - `.claude/agents/moai/manager-git.md`
  - `.claude/agents/moai/manager-project.md`
  - `.claude/agents/moai/expert-*.md` (8개)
  - cross-ref 한 줄: "Spawn 시 context-injection 정책 준수: .claude/rules/moai/development/context-injection.md"
- [ ] `CLAUDE.md` §14 또는 §16에 정책 위치 안내 추가

**Exit Criteria**: 16+ agent body cross-ref 갱신

### M3 — moai-foundation-core SKILL.md Token Budget 절 보강 (Priority: Medium)

- [ ] `.claude/skills/moai-foundation-core/SKILL.md` Token Budget 절에 추가:
  - "sub-agent 호출 시 context 주입 정책: .claude/rules/moai/development/context-injection.md"
  - 5KB cap 명시
  - 3-tier 우선순위 요약
- [ ] modules/token-budget-allocation.md 보강 (있을 시)

**Exit Criteria**: SKILL.md cross-ref + 5KB cap 명시

### M4 — progress.md 권장 schema 정착 (Priority: Medium)

- [ ] `.claude/rules/moai/development/context-injection.md` §5에 권장 schema 작성:
  ```markdown
  # Progress — SPEC-XXX
  
  ## Last Action
  <one-line>
  
  ## State
  - Files touched: <list>
  - Next step: <one-line>
  
  ## Lessons (during this SPEC)
  - <bullet>
  
  ## References
  - @MX:NOTE locations
  - Related SPECs
  ```
- [ ] 1-3KB 권장 사이즈 명시
- [ ] 권장 강조: hard rule 아님

**Exit Criteria**: schema 권장 작성, 강도 명확화 (권장)

### M5 — Marker Convention 정착 (Priority: Medium)

- [ ] `<!-- injected-context -->` ... `<!-- /injected-context -->` 마커 표준 명시
- [ ] 마커가 sub-agent prompt 앞부분에 배치 (task description과 분리)
- [ ] 예시 prompt 5개 (정책 문서 §4에 포함)

**Exit Criteria**: marker 표준 + 예시 정착

### M6 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `internal/template/templates/.claude/rules/moai/development/context-injection.md` 동기화
- [ ] `internal/template/templates/.claude/agents/moai/*.md` 동기화 (M2 cross-ref 반영)
- [ ] `internal/template/templates/CLAUDE.md` 동기화 (M2 cross-ref)
- [ ] `make build` 실행 → embedded.go 재생성
- [ ] CHANGELOG entry

**Exit Criteria**: Template-First sync clean

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 5KB cap 준수 sample 5 sub-agent 호출 측정
- [ ] cross-ref 16+ agent body 검증
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 정책 문서 핵심 절 구조

```
# Context Injection Policy

## 1. Overview
This policy applies to the MoAI orchestrator when spawning sub-agents...

## 2. Token Budget Cap
5000 tokens per sub-agent invocation (default).
Override via .moai/config/sections/observability.yaml.context_injection.cap

## 3. Priority Order (3-Tier)
1. SPEC progress.md (highest)
2. Recent feedback (MEMORY.md excerpts, last 7 days)
3. Domain lessons (lessons.md, anti-patterns)

## 4. Marker Convention
<!-- injected-context -->
...injected text...
<!-- /injected-context -->

<task description follows>

## 5. Recommended progress.md Schema
(table)

## 6. Truncation Strategy
If sum > 5KB → truncate from lowest priority

## 7. Anti-Patterns
- Injecting raw secrets
- Cross-agent private memory without consent
- Injecting >5KB

## 8. Exception Cases
- research-only sub-agent: relaxed
- one-shot read-only: skip if irrelevant
```

### 4.2 cross-ref 텍스트 표준 (manager/expert agent body)

```markdown
## Context Injection

When spawned by the MoAI orchestrator, this agent receives context per
.claude/rules/moai/development/context-injection.md (5KB cap, 3-tier priority).
Injected context appears between `<!-- injected-context -->` markers.
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| 5KB cap이 부족하여 핵심 누락 | Medium | High | priority + truncation 전략 + 사용자 설정 가능 cap |
| orchestrator가 텍스트 정책 미준수 | High | Medium | 16+ agent body에 cross-ref 강제, 추후 helper 자동화 |
| progress.md 부재 SPEC | High | Low | silent skip 정책 |
| cross-ref 누락된 agent에서 정책 미적용 | Medium | Medium | M2 checklist 16+ entries 명시 |
| 정책 위반 자동 탐지 불가 | High | Medium | living document + 분기별 audit (별도 SPEC) |

## 6. Dependencies

- 선행 SPEC: 없음 (독립)
- 의존 입력: progress.md (SPEC-PERSIST-001), MEMORY.md (Claude Code memory)
- sibling SPEC: SPEC-MEM-SCOPE-001 (memory scope) — 본 SPEC이 인용 가능
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (5KB cap LSP 검증): 불가 (텍스트 정책), 자동화는 후속 SPEC
- **OQ2** (progress.md hard rule 격상 시점): 본 SPEC scope 외, 측정 후 결정
- **OQ3** (multiple memory 우선순위 algorithm): MEMORY.md > lessons.md > observations.md (명시)
- **OQ4** (예외 케이스 정의): research-only sub-agent (researcher, analyst) — §8에 명시

## 8. Rollout Plan

1. M1-M6 구현 후 본 프로젝트의 다음 SPEC run에 적용 (dogfooding)
2. 5 sub-agent 호출 측정 → cap 준수 확인
3. CHANGELOG + v2.x.0 minor release

End of plan.md (SPEC-CONTEXT-INJ-001).
