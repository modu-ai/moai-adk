---
id: SPEC-PHASE-FIELD-VALIDATION-001
title: "phase 프론트매터 필드의 값-형태 검증과 오염 코퍼스 교정 — 진행"
version: "0.2.0"
status: in-progress
created: 2026-08-02
updated: 2026-08-02
author: Goos Kim
priority: P2
phase: "v3.0.2"
module: "internal/spec, .claude/agents/moai"
lifecycle: spec-anchored
tier: M
tags: "spec-lint, frontmatter, phase, drift, authoring-guard"
---

# 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (frontmatter `tier: M` 명시 — spec.md / plan.md / acceptance.md + progress.md)
- REQ **16**건 (Tier M 상한 16), AC **16**건 (상한 16), 미커버 REQ 0건
  (acceptance.md §H 대조표)
- 마일스톤 4개: M1 값-형태 검증 / M2 저작 지시 / M3 코퍼스 교정 14건 / M4 회귀 가드
- 신규 finding 코드: `FrontmatterPhaseInvalid` — 강등 대상 집합에 **미등록**
  (설계의 핵심; AC-PFV-003이 기계 판정)

### 저작 시점 실측 기준선

- `spec.md` 전수: 564개(이 SPEC 포함), 부정 토큰 정확 일치 위반 **9**건
- 전체 산출물 부정 토큰 위반 **31**건, 걸친 SPEC **20**개
  (spec.md 오염 9 + 형제만 오염 레거시 11)
- 부분 문자열 판정 시 추가 오탐 **8**건 → 정확 일치 채택 근거
- 엄격 semver 허용목록 시 오탐 **301**건 (310 − 9) → 허용목록 기각 근거
- `moai spec lint --json`: 62건, error **0**건, exit 0
  (`MissingExclusions` 24 / `StatusGitConsistency` 16 / `FrontmatterInvalid` 14 /
  `LegacyEARSKeyword` 7 / `OwnershipTransitionInvalid` 1)
- 이 SPEC의 `spec.md` 단독 린트: `[]` (finding 0건)
- 템플릿 미러 중립성 grep baseline: **9건**, 줄 `68 69 104 134 135 149 150 151 202`
  (전부 교육용 예시 식별자 — M2가 도입하는 것이 아님)

### 강등 경로 실측 (iter1 FAIL의 근원)

유산 분류 디렉터리 사본에 강등 대상 코드와 비대상 코드를 동시에 심고 한 번의
린트로 관측:

```
{"code":"CoverageIncomplete","severity":"error","advisory":null}      # 비대상 → 살아남음
{"code":"FrontmatterInvalid","severity":"warning","advisory":true}    # 대상  → 강등됨
```

전용 코드 방식이 작동한다는 직접 증거. 상세는 plan.md §A.5.

### 상태

- 미해결 결정: 없음. `[NEEDS CLARIFICATION]` 마커 0건.
- 열린 위험: plan.md §G 5건. 최상위는 "M1 단독 랜딩 시 error 9건 발생"
  (전용 코드 채택으로 v0.1.0의 1건에서 확대 — M3 선행 또는 동시 랜딩 필수).
- plan-audit iter1: FAIL 0.75 → 본 v0.2.0에서 D1~D10 반영.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
