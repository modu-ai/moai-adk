---
id: SPEC-NO-HYBRID-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-NO-HYBRID-001

## 1. Overview

Anthropic의 SRP (Single Responsibility Principle) 권고를 본 프로젝트의 도구/워크플로우에 적용. 정책 문서 신설 + 1회성 audit. 실제 분리 작업은 audit 결과에 따라 후속 SPEC.

## 2. Approach Summary

**전략**: Policy-First, Audit-Then-Plan, Defer-Splits.

1. `.claude/rules/moai/development/single-responsibility.md` 정책 문서 신설
2. 5+ Anti-Pattern catalog (Anthropic verbatim + 본 프로젝트 변환)
3. 5-step audit 절차 정의
4. 1회성 audit 실행 → `.moai/reports/srp-audit-2026-04-30.md`
5. Cross-ref + Template-First sync

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] 본 프로젝트의 도구/워크플로우 인벤토리 작성:
  - `.claude/skills/*/SKILL.md` (~30+)
  - `.claude/agents/moai/*.md` (~25)
  - `internal/cli/*.go` 명령어 (~15)
  - hook handlers (~10)
  - 슬래시 커맨드 `/moai *` (~12)
- [ ] Anthropic blog "Seeing Like an Agent" verbatim 재확인 (2 인용)
- [ ] "3+ distinct execution modes" trigger 정의 명확화
- [ ] cohesion vs hybrid 구분 기준 마련

**Exit Criteria**: 인벤토리 + trigger 정의 + 구분 기준

### M1 — Policy Document 작성 (Priority: High)

- [ ] `.claude/rules/moai/development/single-responsibility.md` 신규 작성
  - §1 Overview + scope (도구 차원 SRP)
  - §2 Trigger Threshold (3+ distinct execution modes)
  - §3 Cohesion vs Hybrid (구분 기준)
  - §4 5+ Anti-Pattern Catalog (Anthropic verbatim 인용)
  - §5 Format-based Control vs Tool-based Control
  - §6 5-step Audit Procedure
  - §7 Remediation Sequencing (1+ violation → 후속 SPEC sequencing)
  - §8 Migration & Backward Break Considerations
- [ ] 정책 문서 3-4KB 범위

**Exit Criteria**: 8 절 모두 작성, verbatim 2+ 인용

### M2 — Anti-Pattern Catalog (5+) 작성 (Priority: High)

- [ ] AP1: **Hybrid tool confusion** (Anthropic verbatim)
  - 인용: "Adding parameters to serve multiple purposes simultaneously..."
  - 본 프로젝트 적용: 다목적 명령어 식별 → 후속 SPEC sequencing
- [ ] AP2: **Format-based control** (Anthropic verbatim)
  - 인용: "Attempting to constrain outputs through markdown formatting..."
  - 본 프로젝트 적용: "respond in JSON" prompt → structured tool support
- [ ] AP3: **Mode parameter explosion** (본 프로젝트 변형)
  - 도구가 `--mode init|analyze|generate|refresh` 같은 모드 파라미터로 다목적 처리
  - 권고: subcommand로 분리 (e.g., `init`, `refresh` 별도 명령)
- [ ] AP4: **Implicit subcommand routing** (본 프로젝트 변형)
  - 단일 명령어가 인자 패턴으로 다른 행동 분기
  - 권고: 명시 subcommand로 분기 명료화
- [ ] AP5: **Cross-domain handler** (본 프로젝트 변형)
  - 단일 hook이 multiple event type 처리
  - 권고: event type별 별도 handler

**Exit Criteria**: 5 AP 명시, 각 AP에 verbatim 또는 본 프로젝트 example

### M3 — Audit 절차 정의 (Priority: High)

- [ ] 5-step audit 절차:
  - **Step 1: Tool/Command Inventory**
    - `.claude/skills/*/SKILL.md`
    - `.claude/agents/moai/*.md`
    - `internal/cli/*.go`
    - hook handlers
    - 슬래시 커맨드
  - **Step 2: Multi-Mode Detection**
    - 각 항목당 mode/subcommand 수 측정
    - 3+ mode 항목 flag
  - **Step 3: SRP Violation Assessment**
    - Mode 간 책임 명확 분리되는가? (Y/N)
    - 통합 이유 (편의 / 진정한 응집성)?
    - 사용자 혼동 사례?
  - **Step 4: Remediation Recommendation**
    - Split / Subcommand / Keep 결정
    - 우선순위 (High / Medium / Low)
  - **Step 5: Anti-Pattern Catalog Update**
    - 본 프로젝트 발견된 패턴 → 카탈로그 추가

**Exit Criteria**: 5-step 절차 명시

### M4 — 1회성 Audit 실행 (Priority: High)

- [ ] `.moai/reports/srp-audit-2026-04-30.md` 작성
  - §1 Audit Scope (날짜, 범위, 인벤토리 수)
  - §2 Multi-Mode Detection 결과 (3+ mode 항목 list)
  - §3 SRP Violation Assessment (각 항목별)
  - §4 Remediation Recommendations (priority + 후속 SPEC 후보)
  - §5 Anti-Pattern 신규 발견 (본 프로젝트 example)
- [ ] 최소 10 항목 audit
- [ ] Audit 결과: 후속 SPEC 후보 list (실제 분리 작업)

**Exit Criteria**: audit report + 10 항목 + 후속 SPEC 후보 list

### M5 — Cross-Reference + Skill Integration (Priority: Medium)

- [ ] cross-ref 추가:
  - `.claude/rules/moai/development/agent-authoring.md`: SRP 정책 인용
  - `.claude/rules/moai/development/skill-authoring.md`: SRP 정책 인용
  - `.claude/skills/moai-foundation-cc/SKILL.md`: SRP 정책 인용
- [ ] `CLAUDE.md` §16 self-check 또는 §7 (Safe Development)에 SRP 인용 추가

**Exit Criteria**: 3+ cross-ref 추가

### M6 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `internal/template/templates/.claude/rules/moai/development/single-responsibility.md` 동기화
- [ ] `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` 동기화 (M5 cross-ref 반영)
- [ ] `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` 동기화 (M5 cross-ref 반영)
- [ ] `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md` 동기화
- [ ] `internal/template/templates/CLAUDE.md` 동기화
- [ ] `make build` 실행 → embedded.go 재생성
- [ ] CHANGELOG entry under Unreleased

**Exit Criteria**: Template-First sync clean

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 정책 문서 8 절 검증
- [ ] anti-pattern 5+ 검증
- [ ] audit report 10+ 항목 검증
- [ ] cross-ref 3+ 검증 (grep)
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 Trigger Threshold "3+ distinct execution modes"

정의:
- "Mode" = 사용자가 다른 결과를 기대하는 명령 분기
- "3+" = 3개 이상의 mode가 단일 도구에 응집됨
- "Distinct" = mode 간 로직 분리 가능 (즉, 응집성 부재)
- 예시 (본 프로젝트):
  - `/moai project`: init / analyze / generate / refresh = 4 mode → flag
  - `/moai db`: init / refresh / verify / list = 4 mode but cohesive (DB 단일 도메인) → keep with subcommand
  - `/moai sync`: PR 생성 단일 책임 → no flag

### 4.2 Cohesion vs Hybrid 구분 기준

| 기준 | Cohesion (keep) | Hybrid (split) |
|------|----------------|---------------|
| Mode 간 공유 state | YES (e.g., DB connection) | NO |
| Mode 간 의미 일관성 | YES (모두 DB 작업) | NO |
| 사용자 mental model | "단일 도메인" | "다른 책임" |
| Subcommand 명료성 | YES (`db init`, `db refresh`) | NO (parameters 충돌) |
| Code reuse | HIGH | LOW |

### 4.3 Audit Report 예시 구조

```markdown
# SRP Audit — 2026-04-30

## 1. Audit Scope
- Date: 2026-04-30
- Inventory: 30 skills + 25 agents + 15 CLI commands + 10 hooks + 12 slash commands

## 2. Multi-Mode Detection
- /moai project: 4 modes (init / analyze / generate / refresh) — FLAGGED
- /moai db: 4 modes (init / refresh / verify / list) — COHESIVE (subcommand 명료)
- ... (10+ entries)

## 3. SRP Violation Assessment
| Item | Modes | Cohesive? | Violation? | Severity |
|------|-------|-----------|-----------|----------|
| /moai project | 4 | NO | YES | High |
| ... | | | | |

## 4. Remediation Recommendations
1. /moai project → 후속 SPEC SPEC-PROJECT-SPLIT-001 (priority: High)
2. ... (3+ 후보)

## 5. Anti-Pattern New Findings
- AP6 (본 프로젝트 변형): ...
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| 기존 워크플로우 분리 시 backward break | High | High | 본 SPEC은 audit + 정책만, 분리는 후속 SPEC + migration guide |
| SRP 과도 적용 → 명령어 폭발 | Medium | Medium | "3+ distinct modes" trigger + cohesion 기준 |
| Anti-pattern 정의의 주관성 | High | Medium | Anthropic verbatim 우선, 본 프로젝트 example 보충 |
| Audit report 부정확 | Medium | Medium | 5-step 절차 + 10+ 항목 hard threshold |
| 후속 SPEC 후보 미실행 → audit dead doc | High | Medium | audit report 결과를 v2.x.0 minor release 계획에 포함 |

## 6. Dependencies

- 선행 SPEC: 없음 (standalone)
- 의존 입력: 본 프로젝트의 인벤토리 (manual)
- sibling SPEC: SPEC-CACHE-ORDER-001 (이번 wave) — 또 다른 Anthropic 권고
- 도구: `make build`, plan-auditor, grep

## 7. Open Questions Resolution

- **OQ1** (audit report 위치): `.moai/reports/srp-audit-2026-04-30.md` (CLAUDE.md §SPEC vs Report → reports/)
- **OQ2** (분리 작업 본 SPEC scope): 본 SPEC은 정책 + audit만, 분리는 후속 SPEC
- **OQ3** ("3+ distinct modes" 정확한 정의): M0/§4.1에서 명시 (mode = 사용자 다른 결과 기대)
- **OQ4** (cohesion vs hybrid 구분): §4.2 5-기준 명시
- **OQ5** (후속 SPEC 활용): audit report 결과를 v2.x.0 minor release 후속 SPEC 후보 list

## 8. Rollout Plan

1. M1-M6 구현 후 audit 결과 review
2. 후속 SPEC 후보 list → priority별 sequencing
3. CHANGELOG entry + v2.x.0 minor release
4. 후속 SPEC: high-priority 1-3개 후속 SPEC 생성 (별도 wave)

End of plan.md (SPEC-NO-HYBRID-001).
