---
id: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
version: "0.1.1"
status: draft
created: 2026-05-15
updated: 2026-05-15
author: manager-spec
priority: P1
tags: "spec-lint, drift, chore-skip, status-git-consistency, bootstrapping-bug, v3r4, foundation"
issue_number: null
title: spec-lint git-implied status가 chore(spec) sweep commit을 건너뛰도록 수정
phase: "v3.0.0 R4 — Foundation Cleanup"
module: "internal/spec/drift.go, internal/spec/lint.go (StatusGitConsistencyRule)"
dependencies: []
related_specs:
  - SPEC-V3R4-SPECLINT-DEBT-001
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "spec-lint Foundation Cleanup — bootstrapping bug 해소"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 — spec-lint git-implied status가 chore(spec) sweep commit을 건너뛰도록 수정

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.1   | 2026-05-15 | plan-audit remediation | plan-auditor 0.924 P1+P2 cosmetic 결함 4건 반영. P1: design.md §3.2의 검증 안 된 `^build:` 인용을 `^refactor:` (transitions.go:32) 로 교체하고 §6.5에 미등록 prefix fall-through ack 추가. P2-1: spec.md §3.1 Ubiquitous Requirements에 있던 REQ-LSCSK-002 ("shall not modify") 를 modality 정합성에 따라 §3.4 Unwanted Behavior Requirements로 재배치 (ID 보존). P2-2: REQ-LSCSK-010 의 "may be externalized" 표현을 EARS Optional 표준형 (`Where ..., the system shall ...`) 으로 재작성. P2-3: research.md §7.1/§2.3의 transitions.go 인용 라인을 실측값으로 보정 (60-92 → 70-92, 75-92 → 84-88). |
| 0.1.0   | 2026-05-15 | manager-spec  | 초기 draft. PR #930 (`bdcb57f8d`) 머지 직후 main에서 `moai spec lint --strict`가 `StatusGitConsistency` WARNING을 7건 반복 보고한다. 원인은 `internal/spec/drift.go::getGitImpliedStatus`가 `git log --grep=<specID> -1`로 **가장 최근** commit을 가져온 뒤 `chore(spec):` sweep commit의 title을 `ClassifyPRTitle`에 넘기는 구조에 있다. `ClassifyPRTitle("chore(spec): ...")`은 의도적으로 `(category="skip-meta", status="")`을 반환하지만, drift.go line 134-136이 빈 status를 `"in-progress"`로 강제 fallback하여 frontmatter status (`implemented`/`completed`)와 불일치를 일으킨다. 본 SPEC은 walker filter를 drift.go에 도입해 chore(spec) commit을 skip하고 의미 있는 분류가 나올 때까지 더 오래된 commit을 탐색하도록 만든다. `ClassifyPRTitle`의 표준 의미는 보존한다 (AC-LSCSK-003 regression guard). BODP 평가: signals A=¬ B=¬ C=¬ → main @ origin/main (plan-in-main 원칙 PR #822 준수). |

---

## 1. Goal

`moai spec lint --strict` 명령이 origin/main (commit `bdcb57f8d`) 기준 `StatusGitConsistency` WARNING을 7건 보고하는 현재 상태를, **WARNING 0건**으로 정리한다. 동시에 51개 SPEC에 누적된 `lint.skip: [StatusGitConsistency]` 회피책의 추가 확산을 막는다. 본 SPEC은 **lint 엔진 (`internal/spec/drift.go`) 만을 수정하며 `.moai/specs/` 어떤 frontmatter도 변경하지 않는다** — 7개 영향 SPEC의 status enum은 그대로 보존된다.

### 1.1 7건 WARNING 대상

PR #930 (`bdcb57f8d`) 머지 직후 `moai spec lint --strict` 출력에 나타나는 SPEC 목록:

| SPEC ID | frontmatter status | git-implied status (현재) | 진단 |
|---------|--------------------|---------------------------|------|
| SPEC-UTIL-001 | implemented | in-progress | sweep commit이 latest로 매칭됨 |
| SPEC-V3R2-CON-001 | implemented | in-progress | 동일 |
| SPEC-V3R2-CON-002 | implemented | in-progress | 동일 |
| SPEC-V3R2-CON-003 | implemented | in-progress | 동일 |
| SPEC-V3R2-RT-001 | implemented | in-progress | 동일 |
| SPEC-V3R2-SPC-003 | implemented | in-progress | 동일 |
| SPEC-V3R4-HARNESS-003 | completed | in-progress | 동일 |

7건 모두 `chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)` commit (PR #930) body의 sweep 대상 목록에서 매칭되었다. 이 commit의 title은 `chore(spec):` prefix이므로 `ClassifyPRTitle`이 `(skip-meta, "")`를 반환하는데, drift.go line 134-136에서 빈 status가 `"in-progress"`로 fallback되어 false-positive WARNING이 발생한다.

### 1.2 근본 원인 (Bootstrapping Bug)

원인 체인 (verified, see research.md §3):

1. `internal/spec/drift.go::getGitImpliedStatus` line 107: `git log <branch> --oneline --no-merges --grep=<specID> -1` 실행 → **가장 최근** commit 한 건만 가져옴
2. 가져온 commit title을 line 129에서 `ClassifyPRTitle`에 전달
3. PR #930 sweep commit의 title은 `chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)`이며, body가 7개 SPEC을 모두 언급 → `--grep=SPEC-UTIL-001` 등 모든 검색에서 이 commit이 latest로 매칭됨
4. `ClassifyPRTitle("chore(spec): ...")`은 transitions.go line 22의 규칙에 의해 `(category="skip-meta", status="")`을 반환 — 이는 **loop prevention 의도로 설계된 정상 동작**임
5. drift.go line 134-136이 빈 status를 받아 `"in-progress"`로 fallback → frontmatter (`implemented` / `completed`)와 불일치 → WARNING 발생

이는 **bootstrapping bug**다 — status drift를 sweep으로 해소하는 행위 자체가 새로운 drift WARNING을 만들어낸다. 본 SPEC 이전까지의 우회책은 SPEC frontmatter에 `lint.skip: [StatusGitConsistency]`를 추가하는 것이며, 현재 51개 SPEC이 이 회피책을 사용한다 (research.md §4 참조). 본 SPEC은 우회책 추가 확산을 멈추고 lint 엔진 자체에서 문제를 해결한다.

### 1.3 SPEC-V3R4-SPECLINT-DEBT-001 와의 관계

- 선행 SPEC `SPEC-V3R4-SPECLINT-DEBT-001` (PR #913/#917/#921 lifecycle, completed)은 ERROR 66건 + WARNING 140건을 0건으로 reset한 일괄 sweep SPEC이었다.
- 그 sweep 행위 자체가 sweep commit을 만들고, 그 commit이 본 SPEC이 해소할 7건 WARNING의 직접 원인이 된다.
- 즉, 선행 SPEC이 해소한 1차 부채를 본 SPEC이 마무리하여 sweep loop를 닫는다.

---

## 2. Background & Problem Statement

### 2.1 spec-lint 시스템 개요

`moai spec lint` 명령은 `.moai/specs/` 하위 모든 SPEC을 순회하면서 frontmatter 무결성, EARS modality 형식, REQ↔AC 커버리지, 의존성 순환, 그리고 frontmatter status와 git 이력 일관성을 검사한다. `--strict` 플래그가 켜진 경우 모든 WARNING도 ERROR로 승격된다. spec-lint CI workflow가 `--strict`로 실행되므로 WARNING 1건이라도 남으면 다른 SPEC의 정상 머지가 차단될 위험이 존재한다.

### 2.2 git-implied status 의 본래 의도

`StatusGitConsistencyRule` (lint.go:879-914)은 frontmatter에 적힌 status가 git 이력에서 추론한 lifecycle 상태와 일치하는지 확인한다. 예를 들어 frontmatter가 `status: draft`인데 git에 이미 `feat(SPEC-XXX):` 머지 commit이 있다면 frontmatter 갱신 누락(`implemented`로 미수정)을 잡아내는 안전망이다. 이 검사는 SPEC-driven 워크플로우의 무결성 핵심이며, 단순히 비활성화하는 것은 잘못된 선택이다.

### 2.3 sweep commit이 만드는 거짓 음성

`chore(spec):` prefix를 가진 commit은 status 변경/lint-skip 등록 같은 메타데이터 정리만 수행한다. `ClassifyPRTitle`은 이를 `skip-meta` 카테고리로 분류하여 빈 status를 반환함으로써 의도적으로 lifecycle 추론에서 제외시킨다. 그러나 `getGitImpliedStatus`는 이런 chore commit을 단순히 "가장 최근"이라는 이유로 채택하고, 그 빈 status를 `"in-progress"`로 fallback한다. 이 fallback이 원래 머지된 `impl(spec):` 또는 `feat(...)` commit의 정보를 가린다.

### 2.4 회피책의 한계

현재 51개 SPEC이 `lint.skip: [StatusGitConsistency]` 를 frontmatter에 명시하여 본 검사를 통과시키고 있다. 그러나 새로운 sweep을 수행할 때마다 추가 SPEC들이 본 회피책에 합류해야 하므로 결국 모든 SPEC이 lint.skip을 보유하는 상태로 수렴한다. 이는 검사 자체를 무력화하는 것과 같다.

### 2.5 해결 방향

본 SPEC은 **lint 엔진의 git 이력 walker가 chore(spec) commit을 건너뛰고 다음 의미 있는 commit으로 이동하도록** 변경한다. 즉:

- `getGitImpliedStatus`는 `git log --grep=<specID> -<N>` (다중 commit 조회)로 확장한다.
- Go-side 필터로 commit title이 chore(spec) skip pattern에 매칭되면 다음 commit으로 이동한다.
- 의미 있는 분류 (즉 `ClassifyPRTitle`이 비어 있지 않은 status를 반환) 를 가진 첫 commit을 채택한다.
- 모든 commit이 skip pattern에 해당하면 unknown 반환 → `StatusGitConsistencyRule`은 이를 skip 처리한다 (false-positive 방지).

`ClassifyPRTitle` 자체의 의미는 보존된다 (regression guard AC-LSCSK-003).

---

## 3. Requirements (EARS)

### 3.1 Ubiquitous Requirements

**REQ-LSCSK-001**: The system shall preserve the existing semantics of `ClassifyPRTitle("chore(spec): ...")` as returning `(category="skip-meta", status="", error=nil)` without modification.

### 3.2 Event-Driven Requirements

**REQ-LSCSK-003**: When `getGitImpliedStatus(specID)` is invoked, the system shall retrieve up to N commits matching the SPEC-ID via `git log --grep=<specID>` and walk them from newest to oldest.

**REQ-LSCSK-004**: When walking commits, if a commit title matches the skip pattern set, the system shall advance to the next older commit instead of returning the fallback `"in-progress"` status.

**REQ-LSCSK-005**: When walking commits, if a commit's `ClassifyPRTitle` result returns a non-empty status, the system shall return that status as the git-implied status and terminate the walk.

### 3.3 State-Driven Requirements

**REQ-LSCSK-006**: While the walk window of N commits is being scanned and only skip-pattern commits are found, the system shall return an `unknown` signal (error or sentinel) when the window is exhausted.

**REQ-LSCSK-007**: While `getGitImpliedStatus` returns an unknown signal, the `StatusGitConsistencyRule` shall treat the result as a skip condition and emit no finding (neither WARNING nor ERROR).

### 3.4 Unwanted Behavior Requirements

**REQ-LSCSK-002**: The system shall not modify any frontmatter content in the 7 affected SPEC documents (SPEC-UTIL-001, SPEC-V3R2-CON-001, SPEC-V3R2-CON-002, SPEC-V3R2-CON-003, SPEC-V3R2-RT-001, SPEC-V3R2-SPC-003, SPEC-V3R4-HARNESS-003) as part of this SPEC.

**REQ-LSCSK-008**: The system shall not classify `chore(spec):` commits or any other configured skip-pattern commits as `"in-progress"` via the fallback path.

**REQ-LSCSK-009**: The system shall not require frontmatter edits to the 7 affected SPECs in order to reach 0 WARNING. The fix shall live exclusively in `internal/spec/drift.go` (and optionally accompanying tests).

### 3.5 Optional Feature

**REQ-LSCSK-010**: Where the operations team requires externalized configuration of skip patterns, the system shall expose the skip pattern list via `.moai/config/sections/spec-lint.yaml` `git_status_skip_patterns:` array. For v3.0.0-rc1 the pattern set is hard-coded in Go source; externalization is deferred (see plan.md §7 OQ2).

---

## 4. Acceptance Criteria

**AC-LSCSK-001**: After the fix is merged into main, `moai spec lint --strict` reports **0 errors and 0 warnings** for the 7 specific SPECs listed in §1.1, without those SPECs requiring `lint.skip` entries.

**AC-LSCSK-002**: A unit test that replays a git fixture (sweep commit on top of an earlier `impl(spec): SPEC-UTIL-001` commit) shall verify `getGitImpliedStatus("SPEC-UTIL-001")` returns the status implied by the **earlier** `impl(spec):` commit, not the sweep commit.

**AC-LSCSK-003**: A regression test shall verify that `ClassifyPRTitle("chore(spec): ...")` continues to return `(category="skip-meta", status="", error=nil)`. This guards against accidental semantic changes in `transitions.go`.

**AC-LSCSK-004**: The walker shall examine at most N commits (N value decided in plan.md §7 OQ1; recommended N=50) before terminating with an unknown signal. The `StatusGitConsistencyRule` shall treat unknown as a skip (no finding emitted).

**AC-LSCSK-005**: All existing tests in `internal/spec/transitions_test.go` and any drift-related tests shall continue to pass. The new test suite shall add at least 4 cases:
- Case (a): A sweep commit hides a real `impl(spec):` commit — walker returns the real status.
- Case (b): A `chore(spec):` commit precedes a real `feat(SPEC-XXX):` commit — walker skips chore and returns implemented.
- Case (c): Only chore commits exist in the SPEC's git history — walker returns unknown/error within the N-commit budget.
- Case (d) [control case]: Latest commit is a real `impl(spec):` commit — walker returns its status without invoking the skip filter.

---

## 5. REQ ↔ AC Coverage Matrix

| REQ | Mapped ACs |
|-----|------------|
| REQ-LSCSK-001 | AC-LSCSK-003 |
| REQ-LSCSK-002 | AC-LSCSK-001 (verifies SPECs unchanged), AC-LSCSK-005 (test suite does not modify SPECs) |
| REQ-LSCSK-003 | AC-LSCSK-002, AC-LSCSK-005 case (a) |
| REQ-LSCSK-004 | AC-LSCSK-002, AC-LSCSK-005 case (b) |
| REQ-LSCSK-005 | AC-LSCSK-002, AC-LSCSK-005 case (a), case (b), case (d) |
| REQ-LSCSK-006 | AC-LSCSK-004, AC-LSCSK-005 case (c) |
| REQ-LSCSK-007 | AC-LSCSK-001, AC-LSCSK-004 |
| REQ-LSCSK-008 | AC-LSCSK-001, AC-LSCSK-002, AC-LSCSK-005 case (a), case (b) |
| REQ-LSCSK-009 | AC-LSCSK-001 (no frontmatter edits required) |
| REQ-LSCSK-010 | (deferred — verified by plan.md §7 OQ2 documentation, not by automated AC) |

각 REQ는 최소 1개 AC와 매핑되고, 각 AC는 최소 1개 REQ와 매핑된다 (REQ-LSCSK-010 제외 — optional feature는 v3.0.0-rc1 범위 밖).

---

## 6. Scope

### In-Scope

- `internal/spec/drift.go::getGitImpliedStatus` 함수 수정 — git log 다중 commit 조회 + Go-side skip pattern 필터 + walker 종료 조건
- `internal/spec/drift.go` 내 새 helper 함수 (예: `shouldSkipCommitTitle(title string) bool`)
- 새 unit test 파일 또는 기존 test 파일 확장 — 최소 4건의 새 test case (AC-LSCSK-005 a/b/c/d)
- `internal/spec/lint.go::StatusGitConsistencyRule::Check` 검증 — unknown 신호를 받았을 때 finding을 emit하지 않는지 확인 (코드 변경 없을 수도 있음; 현재 line 897-900가 이미 error를 skip 처리하므로)
- `getGitImpliedStatus` godoc 문서화 — Korean (per code_comments=ko)

### Out-of-Scope (HARD — reject if proposed)

- **SPEC 콘텐츠 재작성**: 7개 영향 SPEC의 REQ/AC 본문 수정 금지
- **Frontmatter 변경**: 7개 영향 SPEC의 frontmatter status/version/tags 등 모든 필드 수정 금지 (fix는 lint 엔진에 위치하며 데이터에 위치하지 않는다)
- **CI workflow 변경**: `.github/workflows/spec-lint.yml` 수정 금지
- **기존 lint.skip 제거**: 51개 SPEC의 `lint.skip: [StatusGitConsistency]` 엔트리 제거 작업 (별도 cleanup task, 본 SPEC 범위 밖)
- **`ClassifyPRTitle` 의미 변경**: `chore(spec):` 분류가 빈 status를 반환하는 의도적 설계는 보존된다 (AC-LSCSK-003 regression guard)
- **새 severity level 추가**: 현행 SeverityWarning / SeverityError 외 새 등급 도입 금지
- **새 lint rule 추가**: `StatusGitConsistencyRule` 외 신규 rule 도입 금지
- **DB / SPEC-ID convention 변경**: SPEC-ID 명명 규칙 / 분류 체계 변경 금지

---

## 7. Constraints

- Go 1.23+ (per `.claude/rules/moai/languages/go.md`)
- 모든 코드 수정은 `internal/spec/` 패키지에 한정
- 외부 CLI interface (예: `moai spec lint` 명령 플래그) 변경 금지
- `.moai/specs/` 콘텐츠 변경 금지
- 테스트는 `t.TempDir()` 안에서 git 저장소 fixture를 만들어 실행 (CLAUDE.local.md §6 Test Isolation 준수)
- 코드 주석은 Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`)
- Identifier (변수/함수/타입 이름) 는 English
- `go test ./...` 전체 통과 + `go test -race ./internal/spec/` 통과 필수
- `golangci-lint run` 0 warning 유지

---

## 8. Out-of-Scope (Explicit List)

본 섹션은 §6에서 이미 명시한 out-of-scope 항목의 명시 목록이며, run-phase에서 scope creep을 방지한다:

1. 7개 영향 SPEC frontmatter 수정 (status / version / created / updated / tags / 모든 필드)
2. 7개 영향 SPEC 본문 (REQ / AC / HISTORY 등) 수정
3. 51개 SPEC에 등록된 `lint.skip: [StatusGitConsistency]` 엔트리 제거 또는 수정
4. `.github/workflows/spec-lint.yml` 또는 다른 CI workflow 수정
5. `internal/spec/transitions.go::ClassifyPRTitle` 의 `chore(spec):` 분류 규칙 변경
6. `internal/spec/lint.go` 의 신규 rule 추가 또는 기존 rule severity 변경
7. `.moai/config/sections/spec-lint.yaml` 의 skip pattern externalization (deferred to future SPEC per plan.md §7 OQ2)
8. SPEC-ID 명명 규칙, lifecycle status enum, transitions.go의 transition rule 추가/삭제
9. `moai spec lint` CLI flag 추가 (예: `--no-status-check`)
10. Documentation site (docs-site/) 문서 갱신 (별도 sync-phase 책임)

---

## 9. Glossary

- **sweep commit**: SPEC frontmatter의 status / lint.skip / 메타데이터를 일괄 수정하는 commit. Title prefix는 `chore(spec):` 이며 본 SPEC에서는 PR #930 (`bdcb57f8d`) 이 대표 사례다.
- **status-sync commit**: sync-phase 결과로 생성되는 commit. Title prefix는 `docs(sync):` 또는 `sync:` 이며 frontmatter status를 `implemented → completed` 로 전이시킨다. 본 SPEC의 skip pattern 적용 여부는 plan.md §7 OQ3에서 결정된다.
- **bootstrapping bug**: 어떤 시스템이 자기 자신을 정리/sweep할 때, 그 정리 행위 자체가 새로운 정리 대상을 만들어내는 자기 참조형 결함. 본 SPEC의 7건 WARNING이 대표 사례다.
- **git-implied status**: `getGitImpliedStatus(specID)` 함수가 git log를 분석하여 추론하는 lifecycle status (draft / planned / in-progress / implemented / completed / superseded / archived / rejected). frontmatter status와의 일치성을 `StatusGitConsistencyRule` 이 검증한다.
- **lint.skip**: SPEC frontmatter의 array 필드. `[RuleName1, RuleName2, ...]` 형식으로 특정 lint rule을 해당 SPEC에 대해 비활성화한다. 회피책의 일종이며, 본 SPEC은 이 메커니즘에 의존하지 않도록 lint 엔진을 직접 개선하는 것이 목적이다.
- **walker filter**: 본 SPEC에서 도입하는 새 메커니즘. `git log --grep=<specID> -<N>` 결과를 newest-to-oldest 순회하면서 skip pattern에 매칭되는 commit은 건너뛰고 의미 있는 분류가 나오는 첫 commit을 채택한다.

---

## REQ Coverage

REQ-LSCSK-001, REQ-LSCSK-002, REQ-LSCSK-003, REQ-LSCSK-004, REQ-LSCSK-005, REQ-LSCSK-006, REQ-LSCSK-007, REQ-LSCSK-008, REQ-LSCSK-009, REQ-LSCSK-010
