---
id: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
version: "0.1.0"
status: draft
created: 2026-05-15
updated: 2026-05-15
author: manager-spec
priority: P1
tags: "spec-lint, status-git-consistency, drift, chore-spec, sweep, lint-logic-fix, v3r4"
issue_number: 932
title: spec-lint StatusGitConsistencyRule chore(spec) sweep skip 로직 수정
phase: "v3.0.0 R4 — Foundation Cleanup"
module: "internal/spec/drift.go, internal/spec/transitions.go"
dependencies:
  - SPEC-V3R4-SPECLINT-DEBT-001
related_specs:
  - SPEC-V3R4-SPECLINT-DEBT-001
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "spec-lint 정확도 개선 + lint.skip 우회 자산화"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 — spec-lint StatusGitConsistencyRule chore(spec) sweep skip

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. Issue #932 기반. SPEC-V3R4-SPECLINT-DEBT-001 머지 후 `chore(spec): status drift 11건 sweep` (PR #930, commit `bdcb57f`) 머지에 의해 발생한 신규 7건 StatusGitConsistency WARN을 lint logic fix로 영구 해소. `internal/spec/drift.go::getGitImpliedStatus`가 `chore(spec):` prefix 의 sweep commit 을 latest mention 으로 인식하는 false-positive 를 제거한다. `transitions.go::ClassifyPRTitle`는 이미 `chore(spec): → ("skip-meta", "")` 매핑이 존재하나 drift.go 호출부가 empty status 를 "unknown prefix" 로 오인하여 `in-progress` 기본값으로 다운그레이드하는 결함 수정. BODP 평가: signals A=¬ B=¬ C=¬ → main @ origin/main (plan-in-main PR #822 준수). |

---

## 1. Goal

`internal/spec/drift.go::getGitImpliedStatus` 가 git log 에서 SPEC-ID 를 mention 하는 latest commit 1건만 보고 분류하던 휴리스틱을, **`chore(spec):` prefix sweep commit 을 건너뛰고 그 직전 의미 있는 commit 을 사용**하도록 개선한다. 이로써 SPEC-V3R4-SPECLINT-DEBT-001 Wave 2 (PR #930) 의 batch status sweep commit 이 무관한 11개 SPEC 의 git-implied status 를 `in-progress` 로 잘못 추론하던 신규 7건 StatusGitConsistency WARN 을 lint logic 차원에서 영구 해소한다.

### 1.1 배경

- 2026-05-15 SPEC-V3R4-SPECLINT-DEBT-001 의 Wave 2 결과물인 PR #930 (`chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)`) 이 origin/main (commit `bdcb57f`) 에 squash 머지되었다. 이 commit 메시지는 sweep 대상 11개 SPEC-ID 를 본문에서 모두 mention 한다.
- SPEC-V3R4-SPECLINT-DEBT-001 `acceptance.md` AC-SLD-007 v0.1.1 target 은 `≤ 55` WARNING 이었고, lint.skip 으로 51건을 suppress 함으로써 lint-final.md 기준 0건을 달성했다 (`lint-final.md` 보고).
- 그러나 PR #930 의 머지 직후 lint 재실행 결과 **신규 7건 StatusGitConsistency WARN** 이 등장했다. 분포: PR #930 sweep commit 이 mention 한 11개 SPEC 중 4개는 이미 lint.skip 등록되어 있어 silent, 나머지 7개는 lint.skip 미등록 상태로 신규 WARN 트리거.
- 근본 원인은 `internal/spec/drift.go:99-143` `getGitImpliedStatus` 의 동작 로직이다:
  1. `git log main --grep=<SPEC-ID> -1` 으로 latest 1건 commit 조회.
  2. 그 commit title 을 `ClassifyPRTitle` 로 분류.
  3. `transitions.go:22` 에 `chore(spec): → ("skip-meta", "")` 매핑이 정의되어 있으나, drift.go:134-137 의 `if status == ""` 분기가 이를 "unknown prefix" 와 동일하게 처리하여 `"in-progress"` 기본값을 반환.
  4. 결과: 7개 SPEC 의 frontmatter status (`completed`, `implemented` 등) 와 git-implied status (`in-progress`) 불일치 → WARN.
- 이 7건은 lint.skip 등록 (frontmatter 일괄 수정) 으로 우회 가능하나, 해당 접근법은 **lint logic 자체의 결함을 영구화** 한다. 향후 추가 sweep commit (status correction, batch frontmatter sync 등) 머지마다 동일한 false-positive 가 재발한다.
- 본 SPEC 은 lint.skip 우회 대신 **drift.go 의 휴리스틱 자체를 개선** 하여 `chore(spec):` (그리고 `revert:`) prefix commit 을 git log 스캔에서 건너뛰도록 한다. `ClassifyPRTitle` 의 `skip-meta` 카테고리 시그널을 drift.go 가 정직하게 해석하도록 호출부를 수정한다.

### 1.2 사용자 결정 사항

다음 결정은 본 SPEC plan-phase 에서 잠정 채택된다. run-phase 에서 코드 분석 결과에 따라 조정될 수 있다:

1. **스캔 범위**: latest 1건 → latest **N건** 으로 확장. N 의 기본값은 **10** (실용적 상한 — sweep commit 1~2건 + 직전 의미 commit 1건 + 여유분). 11번째 이후 commit 까지 모두 `chore(spec):` 인 경우는 극히 드물며, fallback (§1.2 §3) 로 처리.
2. **건너뛸 prefix 집합**: `chore(spec):`, `chore(specs):`, `revert:`. 이 세 prefix 는 `ClassifyPRTitle` 에서 각각 `skip-meta` / `skip-meta` / `no-op` 으로 분류되며 둘 다 lifecycle 변화를 의미하지 않음. (`revert:` 는 이미 `transitions.go:79` 에서 special-case 처리되나 status 가 `""` 라서 동일 결함 영향.)
3. **Fallback 동작**: 스캔 범위 내 모든 commit 이 skip 대상이면, 기존 동작 (마지막 commit 의 status 또는 `in-progress` 기본값) 을 유지하지 않고 `""` (empty) 반환 → `StatusGitConsistencyRule.Check` 는 `gitStatus == ""` 시 skip 하도록 한다 (해당 SPEC 은 의미 있는 git history 가 부재한 것으로 간주).
4. **단위 테스트**: `internal/spec/drift_test.go` 신규 작성. 테이블 기반 테스트로 (a) chore(spec) 단독, (b) chore(spec) + feat, (c) revert + feat, (d) 모두 chore(spec), (e) sync(): 우선 시나리오 커버. 통합 테스트는 lint 실행 결과 7→0 변화로 검증.
5. **SPEC-V3R4-SPECLINT-DEBT-001 의 lint.skip 자산 보존**: 본 SPEC 의 lint logic fix 이후 51개 lint.skip 항목은 그대로 유지한다. 47개 `completed → implemented` 케이스는 lint logic 개선만으로는 해소되지 않는 다른 원인 (latest commit 이 follow-up feat 인 케이스) 이며, 별도 follow-up SPEC (status-residuals.md §5 F1) 에서 해소 가능.

### 1.3 Non-Goals (Out of Scope)

본 SPEC 은 `getGitImpliedStatus` 의 **chore(spec)/revert sweep skip** 만 다룬다. 다음 작업은 명시적으로 범위 밖이다:

- **47개 `completed → implemented` ambiguous 케이스 자동 해소**: 이 케이스들의 latest commit 은 `feat: ... related to SPEC-XXX` 같은 follow-up bugfix/refactor 이다. chore(spec) skip 만으로는 해소되지 않으며 별도 SPEC 에서 `sync(` / `docs(sync)` prefix 우선 탐색 등 추가 휴리스틱이 필요하다 (status-residuals.md §5 F1).
- **lint.skip 51건 일괄 제거**: 본 SPEC 의 logic fix 이후에도 47+4 케이스는 여전히 WARN 을 트리거하므로 lint.skip 항목은 유지한다. 제거는 follow-up SPEC 에서 결정.
- **새 lint rule 추가**: 신규 검증 룰은 작성하지 않는다. 기존 StatusGitConsistencyRule 의 정확도만 개선한다.
- **`transitions.go::ClassifyPRTitle` API 변경**: 함수 시그니처는 그대로 유지한다. `("skip-meta", "")` 매핑도 유지한다. 호출부 (drift.go) 만 수정한다.
- **frontmatter status 수정**: 본 SPEC 은 코드 로직만 수정하며 어떤 SPEC 의 frontmatter 도 수정하지 않는다.
- **`scripts/spec-status-sync.go` 의 보강**: SPEC-V3R4-SPECLINT-DEBT-001 Wave 2 의 도구는 그대로 둔다. lint 정확도가 개선되면 해당 도구가 다시 호출될 필요 자체가 줄어든다.

---

## 2. Scope

### 2.1 In Scope

**A. Lint logic 개선 (P1, false-positive 영구 해소)**:

- A1. `internal/spec/drift.go::getGitImpliedStatus` 함수 수정:
  - `git log` 의 `-1` 플래그를 `-N` (기본 N=10) 으로 확장
  - 반환된 commit 들을 순회하며 `ClassifyPRTitle` 호출
  - `category == "skip-meta"` 또는 `category == "no-op"` 인 commit 은 건너뛴다
  - 첫 번째 non-skip commit 의 status 를 반환
  - 스캔 범위 내 모든 commit 이 skip 대상이면 `""` (empty) 반환
- A2. `StatusGitConsistencyRule.Check` (lint.go:886) 동작 확인:
  - `getGitImpliedStatus` 가 `""` 반환 시 skip (이미 `if err != nil` 분기에 의해 자연 skip). 다만 err 없이 `""` 반환되는 케이스 대비 명시적 가드 추가.

**B. 검증 + 회귀 방지 테스트**:

- B1. `internal/spec/drift_test.go` 신규 작성 (5+ 시나리오 테이블 기반 테스트, `t.TempDir()` 격리, git init + commit 시뮬레이션).
- B2. `internal/spec/transitions_test.go` 의 `TestClassifyPRTitle` 에 `chore(spec):`, `chore(specs):`, `revert:` 회귀 케이스 추가 (이미 있다면 보강).
- B3. `make test` + `go test -race ./internal/spec/...` 통과.

**C. End-to-end 검증**:

- C1. 로컬 `moai spec lint --strict` 실행 → 신규 7건 StatusGitConsistency WARN 이 0건으로 감소.
- C2. PR 생성 후 spec-lint CI job GREEN.
- C3. SPEC-V3R4-SPECLINT-DEBT-001 의 lint.skip 51건 효과 유지 (회귀 없음).

### 2.2 Out of Scope

§1.3 참조.

---

## 3. Requirements

### 3.1 EARS Requirements

#### REQ-LSCS-001 (chore(spec) sweep commit 스킵, Event-driven)

WHEN `internal/spec/drift.go::getGitImpliedStatus` 가 git log 를 통해 SPEC-ID 를 mention 하는 commit 들을 조회할 때, IF latest commit 의 title prefix 가 `chore(spec):` 또는 `chore(specs):` 인 경우, THEN 그 commit 은 status 분류 대상에서 제외되어야 한다. 함수는 그 직전 의미 있는 commit (skip 대상이 아닌 첫 commit) 의 status 를 반환해야 한다.

#### REQ-LSCS-002 (revert commit 스킵, Event-driven)

WHEN `getGitImpliedStatus` 가 git log scan 중 prefix 가 `revert:` 인 commit 을 만나면, THEN 그 commit 도 status 분류 대상에서 제외되어야 한다. (이미 `ClassifyPRTitle` 에서 `("no-op", "")` 반환하지만 drift.go 호출부가 정직하게 해석해야 한다.)

#### REQ-LSCS-003 (Skip-meta 카테고리의 정직한 해석, Ubiquitous)

`getGitImpliedStatus` SHALL treat `ClassifyPRTitle` 이 반환하는 `category == "skip-meta"` 또는 `category == "no-op"` 시그널을 "skip this commit" 으로 해석해야 한다. 현재의 "empty status → unknown prefix → in-progress 기본값" 변환 로직은 제거되거나 skip-meta/no-op 카테고리를 분리해서 처리해야 한다.

#### REQ-LSCS-004 (스캔 범위 상한, Ubiquitous)

`getGitImpliedStatus` SHALL scan at most **10 commits** (latest, in chronological order from newest to oldest) when searching for the first non-skip commit. WHERE all 10 commits are skip-meta/no-op, the function SHALL return an empty string `""` with `nil` error, signaling that no meaningful lifecycle commit exists.

#### REQ-LSCS-005 (Empty gitStatus 시 lint 게이트 우회, Event-driven)

WHEN `StatusGitConsistencyRule.Check` (lint.go:886) 가 `getGitImpliedStatus` 호출 결과로 `gitStatus == ""` 를 받으면, THEN 해당 SPEC 은 StatusGitConsistency 검사 대상에서 제외되어야 한다 (no Finding emitted). 이는 의미 있는 git history 가 없는 SPEC 에 대한 fail-safe 이다.

#### REQ-LSCS-006 (7건 신규 WARN 해소, Ubiquitous)

본 SPEC 의 run-phase 완료 후, origin/main HEAD 에서 `moai spec lint --strict` 를 실행하면 PR #930 (`bdcb57f`) 머지에 의해 발생한 신규 7건 StatusGitConsistency WARN 이 0건으로 감소해야 한다.

#### REQ-LSCS-007 (회귀 방지 단위 테스트, Ubiquitous)

`internal/spec/drift_test.go` SHALL 신규 작성되어 다음 5+ 시나리오를 테이블 기반 테스트로 커버해야 한다:

1. latest commit = `chore(spec): ...` AND 직전 commit = `feat(SPEC-XXX): ...` → status `implemented` 반환
2. latest commit = `chore(spec): ...` AND 직전 commit = `sync(SPEC-XXX): ...` → status `completed` 반환
3. latest commit = `revert: ...` AND 직전 commit = `feat(SPEC-XXX): ...` → status `implemented` 반환
4. latest 10 commits 전체가 `chore(spec):` → status `""` 반환 (skip)
5. latest commit = `feat(SPEC-XXX): ...` (기존 정상 동작) → status `implemented` 반환

`go test -race ./internal/spec/...` 가 통과해야 한다.

#### REQ-LSCS-008 (기존 lint.skip 자산 회귀 방지, Ubiquitous)

본 SPEC 의 lint logic 수정 이후, SPEC-V3R4-SPECLINT-DEBT-001 Wave 2 에서 등록된 51개 SPEC 의 `frontmatter.lint.skip: [StatusGitConsistency]` 가 정상 동작해야 한다. 해당 51개 SPEC 에서 StatusGitConsistency Finding 이 emit 되지 않아야 한다 (suppress 유지).

#### REQ-LSCS-009 (CI Green 게이트, Ubiquitous)

`.github/workflows/spec-lint.yml` CI job 이 본 SPEC 의 run-phase PR 에서 status 0 으로 종료되어야 한다. PR 은 spec-lint job 이 GREEN 이 될 때까지 `main` 으로 머지되어선 안 된다.

#### REQ-LSCS-010 (Self-coverage, Ubiquitous)

본 SPEC 자체가 자신의 제약을 만족해야 한다: §3 의 모든 REQ-LSCS-NNN 는 `acceptance.md` 의 최소 1개 AC 에서 참조되어야 하며, 본 SPEC frontmatter 는 7개 mandatory 필드 (`title`, `created`, `updated`, `phase`, `module`, `lifecycle`, `tags`) 를 모두 포함해야 하고, §1.3 Out of Scope 섹션은 최소 1개 명시적 항목을 포함해야 한다.

---

## 4. Approach Summary

### 4.1 Wave 분할

본 SPEC 의 run-phase 는 단일 Wave 로 구성된다 (변경 범위가 작고 단일 함수 + 단일 신규 테스트 파일):

- **Wave 1 (단일)**: A1 (drift.go fix) → B1 (drift_test.go) → A2 (lint.go 가드) → B2 (transitions_test.go 보강) → C1 (로컬 lint) → C2 (CI green) → C3 (회귀 없음).

### 4.2 도구

- **drift.go 수정 (A1)**: Edit tool. 함수 본체 ~30 LOC 추가/수정. for-loop 로 git log -N 결과 순회.
- **신규 테스트 작성 (B1)**: Write tool. `internal/spec/drift_test.go` 신규 파일. `t.TempDir()` + `exec.Command("git", "init")` 패턴으로 격리된 테스트 git repo 생성.
- **기존 테스트 보강 (B2)**: Edit tool. `internal/spec/transitions_test.go` 의 TestClassifyPRTitle 에 회귀 케이스 추가.
- **검증 (C1-C3)**: Bash tool. `make test`, `go test -race ./internal/spec/...`, `moai spec lint --strict`.

### 4.3 Branch + PR 전략

- **BODP signals (현재 시점)**: A=¬ (코드 의존성 없음 — 자체 함수만), B=¬ (작업 트리에 본 SPEC 미존재), C=¬ (현재 issue branch open PR 없음) → main @ origin/main.
- **Plan PR**: `claude/issue-932-20260515-1325` → main (squash). 본 PR.
- **Run PR**: `feat/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` → main (squash). plan PR 머지 후 신규 worktree 에서 작업.
- **Sync PR**: 작은 변경 (단일 함수 + 단일 신규 테스트 파일) 이므로 sync PR 은 run PR 에 흡수 가능. 별도 PR 시 `chore/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001-sync`.

### 4.4 Verification Pipeline

```
make test          # 전체 Go 테스트 통과
go test -race ./internal/spec/...  # race condition 없음
moai spec lint --strict  # exit 0, ERROR 0, WARNING 0 (51 lint.skip suppress 유지)
```

---

## 5. Success Criteria

§3 의 REQ 들이 `acceptance.md` AC 에서 검증된다. 본 섹션은 high-level 성과 지표:

- `moai spec lint --strict` exit code 0 on origin/main HEAD after run PR merge.
- spec-lint CI job GREEN on the run PR.
- StatusGitConsistency 신규 WARN: 7 → 0.
- 기존 51건 lint.skip suppress: 유지 (회귀 0건).
- `internal/spec/drift_test.go` 신규 테스트 파일 5+ 시나리오 모두 통과.
- `make test` + `go test -race ./internal/spec/...` 모두 통과.
- 본 SPEC 자체가 `moai spec lint --strict` 에서 ERROR/WARN 0건.

---

## 6. Risks

- **R1**: `getGitImpliedStatus` 의 git log scan 범위 확장 (1 → 10) 으로 인한 성능 영향. **미티게이션**: 188개 SPEC × 10 commits = 1880 git log 호출. 로컬 측정 시 1초 미만 예상. CI runner 에서도 측정 후 N 조정 가능.
- **R2**: `--grep=<SPEC-ID>` 로 fetched 된 commits 가 의도와 다르게 SPEC-ID 를 sub-string 으로 우연 매칭 (예: `SPEC-V3R2-RT-005` mention 시 `SPEC-V3R2-RT-005-foo` 도 매칭). **미티게이션**: 기존 코드의 동일 한계이며 본 SPEC 의 변경 범위 밖. 별도 SPEC 에서 word-boundary regex 도입 가능.
- **R3**: Test 격리 어려움 — `getGitImpliedStatus` 가 실제 git 명령을 호출하므로 단위 테스트에서 mocking 또는 fixture git repo 필요. **미티게이션**: `t.TempDir()` + `git init` + 시나리오별 commit 작성 패턴 사용. CLAUDE.local.md §6 Testing Guidelines 의 `t.TempDir()` 의무 준수.
- **R4**: 본 SPEC 의 fix 가 SPEC-V3R4-SPECLINT-DEBT-001 의 51개 lint.skip 항목 중 일부를 더 이상 필요 없게 만들 가능성. **미티게이션**: 본 SPEC 범위 밖 — lint.skip 51건은 그대로 유지. 향후 follow-up SPEC 에서 lint.skip 감축 가능 여부 평가.
- **R5**: `revert:` skip 으로 인해 의도된 revert lifecycle (예: SPEC 의 코드가 의도적으로 revert 되었을 때 status 가 `draft` 로 돌아가야 하는 케이스) 가 가려질 가능성. **미티게이션**: 현재 `transitions.go:79` 에서 `revert:` 는 이미 `("no-op", "")` 으로 분류되어 status 변화를 의도적으로 회피한다. 본 SPEC 은 drift.go 가 이 의도를 정직하게 해석하도록 만드는 수정이며 새 정책 도입이 아니다.

---

## 7. Out of Scope

§1.3 의 6개 항목이 명시적 Non-Goals 이다. 본 섹션은 reviewer 가 빠르게 확인할 수 있도록 핵심 3건을 재게재한다:

1. **47개 `completed → implemented` ambiguous 케이스 자동 해소**: 본 SPEC 의 chore(spec) skip 만으로는 해소되지 않는다. follow-up SPEC 에서 별도 처리 (status-residuals.md §5 F1).
2. **lint.skip 51건 일괄 제거**: 본 SPEC 은 코드 로직만 수정하며 SPEC frontmatter 는 손대지 않는다.
3. **새 lint rule 추가 또는 `ClassifyPRTitle` API 변경**: 기존 StatusGitConsistencyRule 의 정확도만 개선한다.

---

## 8. References

- Issue #932 (본 SPEC 의 원천 요청)
- Plan PR #931 (본 SPEC 의 plan-phase PR)
- PR #930 (`bdcb57f`): `chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)` — 본 SPEC 의 false-positive 발생 원인 commit
- SPEC-V3R4-SPECLINT-DEBT-001 (predecessor):
  - `spec.md` §3 REQ-SLD-007 (StatusGitConsistency 일괄 정정)
  - `status-residuals.md` §5 F1 (follow-up: getGitImpliedStatus 개선)
  - `lint-final.md` §1 (lint.skip 51건 suppress 효과 측정)
- `internal/spec/drift.go:99-143` `getGitImpliedStatus` — 수정 대상
- `internal/spec/transitions.go:22, 70-92` `ClassifyPRTitle` — `skip-meta` 매핑 소스
- `internal/spec/lint.go:879-914` `StatusGitConsistencyRule` — empty status 시 skip 처리 호출부
- `.claude/rules/moai/workflow/spec-workflow.md` — SPEC phase contract
- CLAUDE.local.md §6 Testing Guidelines (`t.TempDir()` 의무)
- CLAUDE.local.md §18 Enhanced GitHub Flow (branch/PR 전략)
- CLAUDE.local.md §18.12 BODP (branch origin decision)
