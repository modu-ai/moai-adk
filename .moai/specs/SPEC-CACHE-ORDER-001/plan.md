---
id: SPEC-CACHE-ORDER-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-CACHE-ORDER-001

## 1. Overview

Anthropic prompt cache 활용 정책 표준화. 4-part prompt 구조 (static-prefix → dynamic-suffix → system-reminders → user-input) + `<system-reminder>` 가이드 + 모델 전환 회피 (advisor 패턴 cross-ref). 90% 비용 절감 목표.

## 2. Approach Summary

**전략**: Policy-First, Documentation-Driven, Gradual-Rollout.

1. `.claude/rules/moai/development/cache-friendly-prompts.md` 정책 문서 신설
2. 4-part 표준 구조 정의
3. `<system-reminder>` 사용 가이드 + 5+ 예시
4. 모델 전환 회피 (advisor 패턴 cross-ref)
5. `.moai/metrics/cache-hit-rate.jsonl` schema (자동 수집은 후속 SPEC)
6. agent-authoring / skill-authoring cross-ref

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] Anthropic blog "Harnessing Claude's Intelligence" verbatim 4 인용 재확인
- [ ] Anthropic SDK의 cache_hit_tokens 메타데이터 가용성 확인
- [ ] 본 프로젝트 agent body 5개 sample 분석 — static / dynamic 비율
- [ ] SPEC-ADVISOR-001 (Wave 1) advisor 패턴 cross-ref 가능성 확인
- [ ] `<system-reminder>` 사용 사례 본 프로젝트 0건 확인 (베이스라인)

**Exit Criteria**: 4 verbatim 확인 + baseline + advisor 패턴 인용 가능

### M1 — Policy Document 작성 (Priority: High)

- [ ] `.claude/rules/moai/development/cache-friendly-prompts.md` 신규 작성
  - §1 Overview + cache 경제학 (10% cached cost = 90% 절감)
  - §2 4-Part Prompt Structure (static-prefix → dynamic-suffix → system-reminders → user-input)
  - §3 `<system-reminder>` Mechanism + 5+ 예시
  - §4 Model Switch Avoidance (advisor 패턴 cross-ref)
  - §5 Cache Hit Rate Metric Schema (`.moai/metrics/cache-hit-rate.jsonl`)
  - §6 Static-Prefix Stability Guidelines
  - §7 Mutation Strategy (불가피 시 major version boundary)
  - §8 Anti-Patterns (mid-session 모델 전환, prompt mutation 남용)
- [ ] 정책 문서 3-4KB

**Exit Criteria**: 8 절 모두 작성, 4 verbatim 인용

### M2 — 4-Part Structure 명시 (Priority: High)

- [ ] **Part 1: Static Prefix** (변경 거의 없음 → cache hit)
  - Agent identity (e.g., "You are manager-spec subagent")
  - MoAI constitution refs
  - Tool descriptions (input schemas)
  - Standard protocols (askuser-protocol, agent-common-protocol)
- [ ] **Part 2: Dynamic Suffix** (per-invocation 변경)
  - Task-specific context
  - Active SPEC ID
  - Current file paths
  - Injected memory (SPEC-CONTEXT-INJ-001 markers)
- [ ] **Part 3: System Reminders** (message-level, prompt 미수정)
  - Per-turn hints (e.g., "/clear at 75%")
  - Constraint reminders (e.g., "use AskUserQuestion only")
- [ ] **Part 4: User Input** (가장 동적)
  - Latest user message verbatim

**Exit Criteria**: 4 part 모두 정의, 각 part에 example

### M3 — `<system-reminder>` 5+ Example (Priority: High)

- [ ] **Ex1: Threshold reminder**
  ```
  <system-reminder>Context usage: 72%. Approaching 75% threshold.</system-reminder>
  ```
- [ ] **Ex2: Constraint reminder**
  ```
  <system-reminder>Active SPEC: SPEC-METRICS-001. All file changes must align with this SPEC scope.</system-reminder>
  ```
- [ ] **Ex3: Tool availability**
  ```
  <system-reminder>AskUserQuestion has been preloaded via ToolSearch.</system-reminder>
  ```
- [ ] **Ex4: Memory hint**
  ```
  <system-reminder>Memory file lessons.md was updated since the last load. Consider re-loading.</system-reminder>
  ```
- [ ] **Ex5: Mode hint**
  ```
  <system-reminder>Current model: Sonnet (advisor mode). Refer back to Opus for final decision.</system-reminder>
  ```

**Exit Criteria**: 5 예시 명시, prompt mutation 대비

### M4 — Model Switch Avoidance + Advisor Cross-Ref (Priority: Medium)

- [ ] Model switch 회피 정책 명시:
  - Mid-session 모델 전환 = cache 무효화
  - Anthropic verbatim: "Avoid switching models (breaks cache); use subagents for cheaper alternatives"
- [ ] Advisor pattern cross-ref:
  - SPEC-ADVISOR-001 (Wave 1) 인용
  - 비싼 모델 (Opus) 호출 시 advisor (Haiku/Sonnet) 활용 → cache invalidation 회피
- [ ] 예외 케이스: 명시적 advisor 필요 시는 OK (cache loss 감수)

**Exit Criteria**: SPEC-ADVISOR-001 cross-ref + advisor 패턴 명시

### M5 — Cache Hit Rate Metric Schema (Priority: Medium)

- [ ] `.moai/metrics/cache-hit-rate.jsonl` schema 정의:
  ```jsonl
  {
    "timestamp": "ISO-8601",
    "agent_or_skill": "manager-spec",
    "input_tokens": 12000,
    "cached_tokens": 9600,
    "cache_hit_ratio": 0.80,
    "model": "claude-opus-4-7",
    "confidence": "high|low"
  }
  ```
- [ ] Confidence:
  - `high`: SDK 메타데이터에서 cache_hit_tokens 직접 획득
  - `low`: SDK 미지원 시 heuristic (prefix length 비교)
- [ ] 자동 수집은 후속 SPEC (본 SPEC은 schema만)

**Exit Criteria**: schema 정의 + confidence level 정의

### M6 — Cross-Reference + Skill Integration (Priority: Medium)

- [ ] cross-ref 추가:
  - `.claude/rules/moai/development/agent-authoring.md`: 4-part 구조 인용
  - `.claude/rules/moai/development/skill-authoring.md`: 4-part 구조 인용
  - `.claude/rules/moai/workflow/context-window-management.md`: cache invalidation 인용
- [ ] CLAUDE.md §6 (Quality Gates) 또는 §7 (Safe Development)에 cache 정책 인용
- [ ] SPEC-ADVISOR-001 cross-ref

**Exit Criteria**: 3+ cross-ref 추가

### M7 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `internal/template/templates/.claude/rules/moai/development/cache-friendly-prompts.md` 동기화
- [ ] `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` 동기화 (cross-ref 반영)
- [ ] `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` 동기화
- [ ] `internal/template/templates/CLAUDE.md` 동기화
- [ ] `make build` 실행 → embedded.go 재생성
- [ ] CHANGELOG entry under Unreleased

**Exit Criteria**: Template-First sync clean

### M8 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 8 절 검증
- [ ] 4 verbatim 인용 검증 (grep)
- [ ] 5 `<system-reminder>` 예시 검증
- [ ] cross-ref 3+ 검증 (grep)
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 4-Part Prompt 구조 (Markdown 마커 예시)

```markdown
<!-- static-prefix -->
You are manager-spec subagent.
Constitution: .claude/rules/moai/core/moai-constitution.md
Tool: AskUserQuestion (preloaded via ToolSearch)
<!-- /static-prefix -->

<!-- dynamic-suffix -->
Active SPEC: SPEC-XXX-001
Files in scope: spec.md, plan.md
<!-- /dynamic-suffix -->

<system-reminder>
Context: 72% usage.
</system-reminder>

User: Please review SPEC-XXX-001.
```

### 4.2 Cache Cost 경제학 (Anthropic 인용 기반)

```
Base input cost: 100% (예: 100 tokens × $3/1M = $0.30)
Cached tokens cost: 10% (예: 100 tokens × $0.30/1M = $0.03)

만약 prompt 80%가 static + cache hit:
- 80 tokens × 10% = 8 tokens 단가 × $3/1M = $0.024
- 20 tokens × 100% = 20 tokens 단가 × $3/1M = $0.060
- Total: $0.084 (vs base $0.30 = 72% 절감)
```

### 4.3 Static-Prefix Stability 규칙

- agent body 본문: stable (변경 시 major release)
- constitution refs: stable (constitution amendment 외 stable)
- tool descriptions: stable (tool 추가 시만 변경)
- 사용자 SPEC 진행 상태: dynamic (suffix에 배치)

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Cache hit rate metric SDK 미지원 | Medium | Medium | confidence: low fallback (heuristic) |
| 기존 agent body 재구성 비용 | High | Medium | 본 SPEC은 정책 + 가이드만, 재구성은 grad rollout 후속 |
| `<system-reminder>` 형식 oversimplification | Medium | Low | 5+ 예시 명시 |
| 모델 전환 회피가 quality 손해 | Medium | Medium | advisor 패턴 (SPEC-ADVISOR-001)로 보완 |
| Static-first 위반 자동 detection 부재 | High | Medium | manual review + lint 후속 SPEC |

## 6. Dependencies

- 선행 SPEC: 없음 (standalone)
- 의존 입력: SPEC-ADVISOR-001 (Wave 1) — advisor 패턴 cross-ref
- 의존 입력: SPEC-CONTEXT-INJ-001 (Wave 3) — dynamic-suffix 부분
- sibling SPEC: SPEC-NO-HYBRID-001 (이번 wave) — 또 다른 Anthropic 권고
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (cache_hit_tokens 헤더 가용성): M0에서 확인. 미지원 시 confidence: low fallback
- **OQ2** (`<system-reminder>` exact syntax): Anthropic docs 인용, M3에서 5 예시 명시
- **OQ3** (4-part boundary marker): Markdown comments (`<!-- static-prefix -->`) — M2 §4.1 명시
- **OQ4** (cache hit rate 자동 수집 hook): 본 SPEC은 schema만, 자동 수집은 후속 SPEC
- **OQ5** (모델 전환 detection): metric에 `model` 필드 포함, 변경 시 cache invalidation 추정

## 8. Rollout Plan

1. M1-M7 구현 후 정책 문서 review
2. 1 agent body sample 4-part 재구성 (dogfooding)
3. cache hit rate sample 측정 (manual)
4. CHANGELOG + v2.x.0 minor release
5. 후속 SPEC: 자동 cache hit rate 수집 + agent body 일괄 재구성

End of plan.md (SPEC-CACHE-ORDER-001).
