---
id: SPEC-V3R4-LINT-SKIP-CLEANUP-001
version: "0.1.3"
status: completed
created: 2026-05-15
updated: 2026-05-16
author: manager-spec
priority: P3
tags: "spec-lint, lint-skip, cleanup, metadata, status-git-consistency, walker-filter, v3r4, foundation"
issue_number: null
title: 55개 SPEC frontmatter의 lint.skip StatusGitConsistency 일괄 제거
phase: "v3.0.0 R4 — Foundation Cleanup"
module: ".moai/specs/*/spec.md (frontmatter only — 55 files)"
dependencies:
  - SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
related_specs:
  - SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
  - SPEC-V3R4-SPECLINT-DEBT-001
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "spec-lint Foundation Cleanup — workaround 영구 제거"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-LINT-SKIP-CLEANUP-001 — 55개 SPEC frontmatter의 lint.skip StatusGitConsistency 일괄 제거

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.2   | 2026-05-16 | manager-develop (run-phase amend) | run-phase 실측에서 lint.skip suppression이 walker filter 범위를 벗어나는 real status drift도 mask하고 있었음이 드러남 — AC-LSKC-002 wording 재정의 + §1 Goal amend + follow-up SPEC `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (가설) scope 명시. plan-audit D1 (stale date 2026-05-15 → 2026-05-16) + D5 (mid-run resume design §5.4) remediation 포함. status: draft → implemented. |
| 0.1.1   | 2026-05-16 | plan-audit remediation | plan-audit 0.904 PASS 후 P2 4건 remediation: (1) AC-LSKC-002 placeholder → plan.md §5.2 cross-ref, (2) HISTORY date harmonize 2026-05-16, (3) REQ-005↔AC-002 매핑 rationale 명시, (4) design.md §2.4 redundancy 정리. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (main `b395ec563`, PR #933 머지, `9e394e51b` sync 완료)이 `internal/spec/drift.go::getGitImpliedStatus`에 walker filter를 도입함으로써 `chore(spec):` sweep commit으로 인한 `StatusGitConsistency` false-positive WARN이 자동 skip된다. 결과적으로 55개 SPEC frontmatter에 누적된 `lint.skip: [StatusGitConsistency]` 회피책은 더 이상 필요하지 않다. 본 SPEC은 metadata-only 정리로 55 SPEC frontmatter에서 해당 엔트리만 제거한다. SPEC 본문(REQ/AC/HISTORY 등 H2/H3 섹션) 수정 0줄. BODP 평가: A=¬ B=¬ C=¬ → main @ origin/main (plan-in-main + worktree 미사용 정책 per feedback_worktree_never_use). 영향 SPEC 수는 frontmatter strict scan 기준 55개로 확정 (naive grep 기준 57개에서 본문 false-positive 2개 제외). |

---

## 1. Goal

`SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` 머지로 `moai spec lint --strict`가 `chore(spec):` sweep commit을 walker level에서 자동 skip하게 되었다. 따라서 PR #930 등을 통해 55개 SPEC frontmatter에 일괄 추가된 `lint.skip: [StatusGitConsistency]` 항목은 영구적인 metadata noise로 전락했다. 본 SPEC은 55개 SPEC의 frontmatter `lint.skip` 블록에서 `StatusGitConsistency` 엔트리만을 제거함으로써 회피책(workaround)을 영구 제거한다.

- 정리 후 `moai spec lint --strict`의 `StatusGitConsistency` WARN 동작: walker filter (chore-commit skip)로 인해 cleanup 대상 55 SPECs WARN을 0으로 유지할 것으로 예상되었으나 — run-phase 실측에서 lint.skip suppression이 walker filter 범위를 벗어나는 `feat:`/`feat(specs):` commit 기반 real status drift 54건도 mask하고 있었음이 드러남. 정리 후 노출되는 real drift WARN은 본 SPEC scope 외이며 follow-up SPEC `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (가설)에서 처리 권장.
- 정리 대상 SPEC 본문은 일체 수정하지 않음 — frontmatter `lint:` 블록만 제거
- 55 SPEC 외의 frontmatter 수정 0건 (영향 격리)

---

## 2. Background & Problem Statement

### 2.1 워크어라운드의 기원

- 2026-05-15 PR #921 (`SPEC-V3R4-SPECLINT-DEBT-001` sync) 직후 첫 sweep commit이 main에 진입하면서, `moai spec lint --strict`가 `StatusGitConsistency` WARN을 다수 SPEC에 대해 보고하기 시작했다.
- 당시 walker filter가 존재하지 않았기 때문에 일시적 회피책으로 PR #930 (`bdcb57f8d`)에서 11개 SPEC frontmatter에 `lint.skip: [StatusGitConsistency]`를 추가했다. 이후 동일 패턴을 PR #921 sweep 단계에서 추가 SPEC에 일괄 적용한 결과 총 55개 SPEC에 누적되었다.

### 2.2 회피책의 영구 noise 화

- 회피책은 lint 출력만 조용하게 만들 뿐 근본 원인(sweep commit이 git-implied status를 오염시키는 bootstrapping bug)을 해결하지 못한다.
- `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`이 walker filter로 근본 원인을 해소했으므로, frontmatter `lint.skip` 엔트리는 (a) 회피책의 흔적이며 (b) 다른 lint rule을 skip하는 정당한 이유와 시각적으로 구분이 어려운 영구 metadata noise가 된다.

### 2.3 정리 범위 격리 요구

- SPEC 본문(REQ/AC/HISTORY)은 lint 정리와 무관하다. 본 SPEC은 frontmatter 단일 필드만 손대고 본문은 0줄 수정한다.
- `lint.skip` 배열에 다른 lint rule이 함께 있는 경우는 해당 rule entries는 보존하고 `StatusGitConsistency`만 제거한다 (현재 55 SPEC 전수 조사 결과 모두 단일 엔트리이지만, 향후 확장성을 위해 정책으로 명시).

---

## 3. Requirements (EARS)

### 3.1 Ubiquitous Requirements

- **REQ-LSKC-001** (Ubiquitous): The system shall remove the `StatusGitConsistency` entry from the `lint.skip` array of every affected SPEC's `spec.md` frontmatter while preserving all other frontmatter fields and the entire body content.

- **REQ-LSKC-002** (Ubiquitous): The system shall preserve any `lint.skip` entries other than `StatusGitConsistency` when they coexist with `StatusGitConsistency` in the same array.

- **REQ-LSKC-003** (Ubiquitous): The system shall bump each affected SPEC's `version` field by a patch increment (e.g., `0.3.0` → `0.3.1`) and update `updated` to `2026-05-16`, and append a single new row to the HISTORY table that records the cleanup action.

### 3.2 Event-Driven Requirements

- **REQ-LSKC-004** (Event): When the cleanup edit removes `StatusGitConsistency` as the only entry in `lint.skip`, the system shall remove the entire `lint:` block (including the `skip:` key) from the frontmatter so that no empty `lint:` or `lint.skip: []` block remains.

- **REQ-LSKC-005** (Event): When `moai spec lint --strict` is invoked after the cleanup against the modified SPEC set (using the walker filter introduced in `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`), the system shall emit zero `StatusGitConsistency` warnings.

### 3.3 State-Driven Requirements

- **REQ-LSKC-006** (State): While the SPEC's `lint:` block contains entries other than the `skip` key (e.g., future `lint.warn`, `lint.error` extensions), the system shall preserve the `lint:` block and only remove the `StatusGitConsistency` entry from the `skip` sub-array. (Currently no such SPECs exist; this requirement is forward-compatibility.)

### 3.4 Unwanted Behavior Requirements

- **REQ-LSKC-007** (Unwanted): The system shall not modify any frontmatter field other than `version`, `updated`, and the `lint:` block on the 55 affected SPECs.

- **REQ-LSKC-008** (Unwanted): The system shall not modify the body (anything below the closing `---` of the YAML frontmatter) of any of the 55 affected SPECs.

- **REQ-LSKC-009** (Unwanted): The system shall not modify the frontmatter or body of any SPEC outside the 55-affected list.

### 3.5 Optional Requirements

- **REQ-LSKC-010** (Optional): Where the run-phase agent chooses to externalize the bulk edit as a reusable script (e.g., `.moai/scripts/lint-skip-cleanup.sh`), the system shall record the script's idempotency property — running it twice produces an identical result on the second invocation (no-op).

---

## 4. Acceptance Criteria

### AC-LSKC-001 — lint.skip 엔트리 제거

55개 영향 SPEC (research.md §2 list 참조) 의 frontmatter `lint.skip` 배열에서 `StatusGitConsistency` 엔트리가 모두 제거되어야 한다.

검증:
- `grep -rE "^\s+- StatusGitConsistency$" .moai/specs/*/spec.md` 결과 0건

### AC-LSKC-002 — lint.skip suppression 해제 (real drift exposure는 follow-up scope)

본 SPEC의 cleanup 행위는 lint.skip 회피책 metadata 제거 자체에 한정된다. run-phase 실측 결과 lint.skip 회피책이 walker filter 범위를 벗어나는 real status drift (`feat:`/`feat(specs):` commit 기반)도 mask하고 있었음이 드러났으며, 이는 본 SPEC scope 외 follow-up issue로 분리 처리한다.

검증:
- cleanup 대상 55 SPECs frontmatter에서 lint.skip 엔트리가 모두 제거되었음 (AC-LSKC-001과 동등하게 검증)
- cleanup 후 `moai spec lint --strict`의 `StatusGitConsistency` 카테고리 WARN이 노출되는 것은 expected outcome — 회피책 해제로 인한 real drift exposure는 회귀가 아닌 정상 동작
- real drift 해소는 follow-up SPEC `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (가설) scope (run-phase 실측: 55 cleanup population 중 54건 real drift, main-wide total 64 WARN)

### AC-LSKC-003 — SPEC 본문 미수정

55개 영향 SPEC의 본문 (frontmatter 종료 `---` 이후의 모든 H1/H2/H3 섹션 — `# title`, `## HISTORY` 새 row 1줄 제외, `## 1. Goal`, `## 2. ...`, REQ, AC 등 일체) 의 git diff 변경 라인은 HISTORY 신규 row 1줄을 제외하면 0줄이어야 한다.

검증:
- 각 SPEC에 대해 `git diff -- .moai/specs/<SPEC-ID>/spec.md` 의 `+` 라인이 frontmatter 영역 + HISTORY 1줄에 한정됨
- 자동 검증 스크립트: 본문 영역만 추출하여 sha256 비교 (run-phase에서 baseline 캡처 후 post-edit 비교)

### AC-LSKC-004 — 55개 외 SPEC 미수정

55개 외의 다른 모든 SPEC의 frontmatter + 본문 수정 0건이어야 한다.

검증:
- `git diff --name-only .moai/specs/` 결과가 55개 SPEC의 `spec.md` 파일만 포함 (다른 파일 0건)

### AC-LSKC-005 — frontmatter version/updated/HISTORY 갱신

55개 영향 SPEC 각각에 대해:
- `version` 필드가 patch 단위로 bump (예: `0.3.0` → `0.3.1`, `1.0.0` → `1.0.1`)
- `updated` 필드가 `2026-05-16`로 갱신
- HISTORY 표에 새 row 1개 추가 (`| <new-version> | 2026-05-16 | manager-develop (run-phase) | lint.skip StatusGitConsistency 회피책 제거 — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 walker filter 머지로 불필요해짐. |`)

검증:
- 각 SPEC `version` 필드가 baseline 대비 patch 증가
- 각 SPEC `updated: 2026-05-16`
- 각 SPEC HISTORY 표 line 수가 baseline + 1

---

## 5. REQ ↔ AC Coverage Matrix

| REQ | Mapped ACs | Notes |
|-----|-----------|-------|
| REQ-LSKC-001 | AC-LSKC-001, AC-LSKC-003 | 엔트리 제거 + 본문 보존 |
| REQ-LSKC-002 | AC-LSKC-001 | 현재 55 SPEC 모두 단일 엔트리이므로 동일 검증 |
| REQ-LSKC-003 | AC-LSKC-005 | version bump + updated + HISTORY row |
| REQ-LSKC-004 | AC-LSKC-001 | lint 블록 전체 제거 (현재 모든 케이스가 단일 엔트리이므로 항상 발동) |
| REQ-LSKC-005 | AC-LSKC-002 | lint --strict WARN 0. AC-LSKC-002 verification은 post-cleanup (run-phase M3 milestone)에서 수행 — REQ-LSKC-005 trigger는 cleanup 적용 시점에 발동. |
| REQ-LSKC-006 | AC-LSKC-001 | 정책상 forward-compat; 현재 검증 대상 없음 |
| REQ-LSKC-007 | AC-LSKC-004 | 55개 외 SPEC 격리 |
| REQ-LSKC-008 | AC-LSKC-003 | 본문 0줄 (HISTORY row 1줄 제외) |
| REQ-LSKC-009 | AC-LSKC-004 | 55개 외 SPEC 격리 |
| REQ-LSKC-010 | (Optional) | 검증 대상 아님 — run-phase 자유 선택 |

---

## 6. Scope

### In-Scope

- 55개 영향 SPEC (research.md §2 참조)의 frontmatter `lint:` 블록 내 `StatusGitConsistency` 엔트리 제거
- 55개 SPEC frontmatter `version` patch bump + `updated: 2026-05-16` + HISTORY 새 row 1줄 추가
- `moai spec lint --strict` 사후 검증 (`StatusGitConsistency` WARN 0건 확인)
- (Optional) 재사용 가능한 bulk edit 스크립트 작성 — idempotent

### Out of Scope

- 55개 영향 SPEC 외 다른 SPEC frontmatter 수정 (예: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 자체, SPEC-V3R4-SPECLINT-DEBT-001)
- SPEC 본문 (### 헤더 아래 REQ/AC/HISTORY 등 모든 H2/H3 섹션) 수정 — HISTORY 새 row 1줄 추가는 frontmatter-인접 메타데이터로 간주
- `internal/spec/` 코드 변경 (walker filter는 PR #933에서 이미 머지됨)
- CI workflow 수정 (`.github/workflows/spec-lint.yml` 등)
- docs-site / README / CHANGELOG 외 문서 수정
- 다른 lint rule (예: `OrphanBCID`, `MissingExclusions`, `RequirementCoverage`) 의 `lint.skip` 엔트리 정리
- 55개 SPEC의 `status`, `priority`, `dependencies`, `related_specs`, `tags`, `lifecycle` 등 frontmatter 다른 필드 수정
- 55개 SPEC의 `plan.md` / `acceptance.md` / `design.md` / `research.md` 수정

---

## 7. Constraints

- **C1 (HARD)**: predecessor SPEC `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (PR #933 `b395ec563`)이 main에 머지되어 있어야 한다. walker filter가 없으면 본 cleanup은 lint --strict 회귀를 초래한다.
- **C2 (HARD)**: 본 SPEC 자체의 frontmatter에는 `lint.skip` 회피책을 도입하지 않는다 — walker filter 기반 정리이므로 필요 없다.
- **C3 (HARD)**: 모든 수정은 main checkout에서 수행 (worktree 미사용 정책 per `feedback_worktree_never_use`).
- **C4 (HARD)**: `manager-develop` (run-phase) 가 55개 SPEC 모두를 sequential하게 수정 — 병렬 수정 금지 (lint 검증 baseline 충돌 회피).
- **C5 (HARD)**: `version` bump는 patch 단위 (semver Z 증가) — metadata cleanup이므로 minor/major bump 부적절.

---

## 8. Out of Scope (Explicit List)

본 SPEC은 다음 항목들을 명시적으로 out of scope 처리한다. 만약 run-phase 진행 중 아래 항목 수정 필요성이 발견되면 별도 SPEC을 발급하고 본 SPEC scope를 변경하지 않는다:

1. **SPEC 본문 (### 헤더 아래 REQ/AC/HISTORY 외 모든 sections)**: 수정 0줄. HISTORY 새 row 1줄 추가만 허용.
2. **55개 외 다른 SPEC (예: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001, SPEC-V3R4-SPECLINT-DEBT-001)**: frontmatter / body 수정 0건.
3. **`internal/spec/` Go 소스 코드**: walker filter는 PR #933에서 이미 머지됨. lint engine 추가 변경 없음.
4. **CI workflow 파일 (`.github/workflows/*`)**: lint job 정의 변경 없음.
5. **docs-site (`docs-site/content/**`)**: 4-locale reference 수정 없음.
6. **README.md / README.ko.md**: cleanup은 internal metadata 변경이므로 사용자 가시 변경 없음.
7. **CHANGELOG.md**: `[Unreleased]` 섹션에 1줄 정도 entry 추가는 sync-phase에서 처리 (run-phase가 아닌 sync-phase 범위).
8. **다른 lint rule 의 lint.skip 엔트리 (예: `OrphanBCID`, `MissingExclusions`)**: scope 격리. 별도 SPEC 발급.
9. **55개 SPEC 의 `plan.md` / `acceptance.md` / `design.md` / `research.md`**: cleanup 대상 아님.
10. **55개 SPEC의 `status`, `priority`, `dependencies`, `related_specs`, `tags`, `lifecycle` 등 frontmatter 다른 필드**: 유지.
11. **frontmatter ordering / formatting (key 순서, 들여쓰기 스타일)**: 가능한 한 baseline 보존. 자동 YAML re-serialize로 인한 key reordering은 허용하지 않음.

---

## 9. Glossary

- **lint.skip**: SPEC frontmatter `lint:` 블록 하위의 `skip:` 배열. 특정 lint rule을 해당 SPEC에 한해 비활성화하는 회피책.
- **StatusGitConsistency**: `internal/spec/lint.go`의 lint rule. frontmatter `status` 필드가 git-implied status (가장 최근 의미 있는 commit title의 분류 결과) 와 일치하는지 검증.
- **Walker Filter**: `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`이 `internal/spec/drift.go::getGitImpliedStatus`에 도입한 `chore(spec):` sweep commit skip 로직. `gitLogWindowSize=50` 만큼 commit history를 탐색하여 의미 있는 commit을 찾을 때까지 sweep commit을 skip한다.
- **Sweep Commit**: `chore(spec): ...` prefix를 사용하는 일괄 메타데이터 수정 commit. PR #930, #921 등이 예시.
- **Bootstrapping Bug**: sweep commit이 git-implied status를 오염시키는 회귀 — sweep commit 자체가 lint failure를 유발하고, 그 fix가 또 다른 sweep commit을 발생시켜 순환되는 문제.
- **Idempotent**: 동일 작업을 두 번 적용해도 결과가 동일한 성질. 본 cleanup은 두 번째 실행 시 no-op이어야 한다 (lint.skip에 StatusGitConsistency가 이미 없으므로).
- **BODP (Branch Origin Decision Protocol)**: §18.12 (CLAUDE.local.md) 참조. 3개 signal (A: dependency overlap, B: worktree co-location, C: open PR head) 기반 base branch 결정.
- **plan-in-main**: PR #822에서 정착된 컨벤션. plan-phase 산출물은 worktree 없이 main checkout에서 직접 작성하고 PR 경유로 머지.
