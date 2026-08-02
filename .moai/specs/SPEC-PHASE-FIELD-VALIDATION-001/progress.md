---
id: SPEC-PHASE-FIELD-VALIDATION-001
title: "phase 프론트매터 필드의 값-형태 검증과 오염 코퍼스 교정 — 진행"
version: "0.2.0"
status: completed
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

```yaml
sync_complete_at: 2026-08-02
sync_commit_sha: 460c529b2
sync_status: audit-ready

b12_self_test_a: "grep -c 'SPEC-PHASE-FIELD-VALIDATION-001' CHANGELOG.md → 0 (방출 전 실행, 사전 중복 없음)"
b12_self_test_b: "AC 개수: grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 16 (AC-PFV-001..016, 0이 아님을 직접 확인). CHANGELOG 항목은 16/16 판정으로 서술"
b12_self_test_c: "CHANGELOG 항목이 지목한 파일 5개 전부 ls 로 실재 확인 (internal/spec/lint.go / internal/spec/lint_phase_test.go / .claude/agents/moai/manager-spec.md / internal/template/templates/.claude/agents/moai/manager-spec.md / internal/template/catalog.yaml)"

changelog_entry_position: "[Unreleased] > ### Added 의 첫 항목 (SPEC-REF-SEO-ABSORB-001 항목 바로 앞)"

frontmatter_status_transitions:
  spec.md: "in-progress → implemented → completed (단일 sync 커밋 병합 전이; updated: 2026-08-02 — 이미 해당 날짜라 값 변화 없음)"
  plan.md: "in-progress → implemented → completed (동일)"
  acceptance.md: "in-progress → implemented → completed (동일)"
  progress.md: "in-progress → implemented → completed (동일)"

canary_compliance_check:
  applicable: true
  reason: "이 SPEC은 전방위 정책을 정의한다 — 모든 SPEC 산출물의 `phase:` 값은 릴리스 대상 버전 문자열이어야 하며 워크플로 단계 토큰이 아니다. 그 정책을 자기 자신에게 적용한 결과가 아래 self-check 이다."
  self_check: "이 SPEC 4개 산출물의 phase 값 = \"v3.0.2\" (4/4). 새 가드가 자기 자신을 통과한다."
  observed: "moai spec lint --json 에서 FrontmatterPhaseInvalid 0건"

user_facing_surface_judgment:
  readme: "변경 없음"
  docs_site: "변경 없음"
  reason: "변경은 SPEC 저작 파이프라인 내부에 한정된다 — 린트 finding 코드 하나와 저작 에이전트 지시문. 사용자가 쓰는 CLI 플래그·출력·설정 키가 하나도 바뀌지 않았다. 다만 `moai spec lint` 를 돌리는 사용자에게는 새 error 코드가 보일 수 있으며, 그 사실은 위 CHANGELOG 항목이 담당한다."

residual_risk:
  - "AC-PFV-006 은 저장소가 error 0건 상태인 동안 실패할 수 없다 — 형제 AC 보다 약한 검사다 (run-phase 에서 이미 부채로 기록)."
  - "판정은 정확 일치 denylist 다. `plan` / `run` / `sync` 이외의 새로운 오염 토큰이 등장하면 잡지 못한다. 허용목록으로 뒤집으면 301건 오탐이 되므로 의도적 선택이다."
```

**sync 시점 회귀 관측**

| 검사 | 명령 | 관측 |
|---|---|---|
| 저장소 spec lint | `./bin/moai spec lint --json` | exit 0, 62건 / error 0건 (`MissingExclusions` 24 / `StatusGitConsistency` 16 / `FrontmatterInvalid` 14 / `LegacyEARSKeyword` 7 / `OwnershipTransitionInvalid` 1) — run-phase 기준선과 동일 |
| 4개 산출물 status | `grep -n '^status:' {spec,plan,acceptance,progress}.md` | 4/4 `completed` |

**미검증 (Gaps)**

- `§E.2 Run-phase Evidence` / `§E.3 Run-phase Audit-Ready Signal` 은 `_<pending run-phase>_` 로 남아 있다. 두 절의 소유자는 manager-develop 이며 sync-phase 소유 범위 밖이라 채우지 않았다. run-phase 증거는 커밋 `d320795bc` 의 메시지 본문에 있다.
- `sync_commit_sha` 는 sync 커밋 시점에 자기 해시를 참조할 수 없어 placeholder 로 기록한 뒤 이 후속 커밋에서 백필했다. 백필 값 `460c529b2` 는 `git rev-parse --short 460c529b2` 로 해소를 확인하고 `git show --stat` 으로 대상 파일 5개(CHANGELOG.md + 산출물 4개)를 관측한 뒤 기록했다.
