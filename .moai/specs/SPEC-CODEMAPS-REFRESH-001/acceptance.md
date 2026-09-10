---
id: SPEC-CODEMAPS-REFRESH-001
title: "acceptance.md — codemaps 재생성 정확성 검증 인정 기준"
version: "0.1.1"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
tier: M
---

# acceptance.md — SPEC-CODEMAPS-REFRESH-001

## §A. 검증 원칙

이 SPEC의 합격 판정은 두 층이 **독립적으로** 성립해야 한다.

1. **게이트 층** — `moai graph check`의 codemaps verdict=fresh. 이것은 "문서가 최근 트리와 일치하는 시점에 생성됐다"는 최소 신호일 뿐이다.
2. **정확성 층** — 재생성 결과를 실제 트리와 대조한 3개 검증 항목의 증거 파일. 게이트가 advisory(`blocking: false`)이고 스탬프는 기계적으로 갱신 가능하므로, 정확성 층 없는 게이트 녹색은 합격이 아니다(REQ-CMR-008, 카드의 중심 요건).

모든 AC는 명령 + 기대 출력으로 이분 판정된다. 증거는 `.moai/reports/t432/`(세션 종료 후에도 경로가 살아 있는 tracked 위치)에 수출한다.

## §D. AC Matrix

### AC-CMR-001 — 재생성 완전성 (MUST)

- **Given** 트리가 기준선 상태(codemaps stale, provenance가 구 커밋 `7fc0af324cf4...`를 가리킴)이고
- **When** `/moai codemaps --force`가 완료되면
- **Then** `.moai/project/codemaps/`의 6개 문서(overview, modules, dependencies, data-flow, entry-points, docs-truth)가 전부 재생성되고, 재생성 리포트가 6파일 전부를 생성 대상으로 보고한다. 일부 파일만 갱신되면 FAIL.

증거: 재생성 리포트(6파일 목록) + `ls .moai/project/codemaps/` (7개 항목 — 문서 6 + provenance.json).

### AC-CMR-002 — 경로 실존 표 수출 (MUST, accuracy a)

- **Given** 재생성이 완료된 6개 문서가 있고
- **When** 문서에서 인용된 `(internal|pkg|cmd)/` 경로 전체를 추출해 각각 트리 존재 여부를 검사하면
- **Then** 증거 파일에 (경로 → exists/absent) 전수 표가 존재하고, absent 경로 전부가 t304 인계(known-6) 또는 new-findings 섹션에 분류돼 있다. 행 수 정의: 경로는 중복 제거한 유니크 기준으로 세며, 표의 행 수 = 유니크 인용 경로 수다(동일 경로의 반복 인용은 1행). 유니크 경로가 하나라도 누락되거나 미기록 absent가 남으면 FAIL.

증거: `.moai/reports/t432/codemaps-accuracy-verification.md` §경로 실존.

### AC-CMR-003 — 패키지 구조 대조 (MUST, accuracy b)

- **Given** `go list ./...` 출력(run 시점 실측, 기준선 137)과 재생성된 modules.md·dependencies.md가 있고
- **When** 문서의 패키지 열거를 `go list ./...`와 대조하면
- **Then** 모든 불일치(문서에만 있는 유령 패키지 / 트리에 있는데 문서가 누락한 패키지)가 증거 파일에 verbatim 기록됐고, known-6는 t304 인계 섹션으로 분류돼 있다. 대조를 실행했다는 기록 자체가 없으면 FAIL(불일치 0이어도 대조 명령+결과는 남아야 한다).

증거: 동 파일 §패키지 구조 대조 (명령 + 결과).

### AC-CMR-004 — 인용 식별자 실존 (MUST, accuracy c)

- **Given** 재생성된 entry-points.md·data-flow.md의 식별자 인용(예: `cli.Execute`, `InitDependencies`, `EmbeddedTemplates`, `Deployer.Deploy`)과 docs-truth.md §1 카탈로그가 있고
- **When** 각 식별자를 명명된 파일·패키지에서 grep하고, 카탈로그 §1의 모든 행을 `.claude/agents/` 트리 나열과 전수 대조하면
- **Then** hit/miss 전수 표가 증거 파일에 있고, miss 전부가 기록돼 있다(수정은 하지 않는다). 표 없이 "확인 완료" 선언만 있으면 FAIL.

증거: 동 파일 §식별자 실존.

### AC-CMR-005 — 스탬프 도달성 (MUST)

- **Given** 재생성·검증이 완료되고
- **When** `moai graph stamp codemaps --commit "$(git merge-base HEAD origin/develop)"`이 실행되면
- **Then** provenance.json의 commit_sha가 갱신됐고 `git merge-base --is-ancestor <새 sha> origin/develop`이 rc=0이며, 새 sha는 worktree 브랜치 HEAD가 아니다(merge-base와 HEAD가 일치하는 트렁크 상태는 예외적으로 허용).

증거: `jq -r .commit_sha .moai/project/codemaps/provenance.json` + merge-base 명령 rc. bare-HEAD 스탬프 흔적은 FAIL.

### AC-CMR-006 — 게이트 종결 (MUST)

- **Given** 재생성(AC-CMR-001)과 재스탬프(AC-CMR-005)가 완료되고
- **When** `moai graph check`를 실행하면
- **Then** codemaps 계층이 verdict=fresh(metric value < 40, 기대 0)이고, 다른 계층 어느 것도 verdict=stale이 아니다. 신규 worktree에서 mx-index/edges의 verdict=absent는 예상 상태로 FAIL 아니다.

증거: `moai graph check` verbatim 출력(progress.md §E.2 격납).

### AC-CMR-007 — 범위 위생 (MUST)

- **Given** SPEC 작업이 종결된 작업 트리가 있고 측정 창이 base `ad272be20`으로 고정돼 있으며
- **When** 레인 변경 집합 전체 — tracked 변경(`git diff --name-only ad272be20`, 커밋·미커밋 통합)과 untracked 신규(`git status --porcelain`의 `??` 행) — 의 변경 경로를 검사하면
- **Then** 변경이 `.moai/project/codemaps/**`, `.moai/reports/t432/**`, `.moai/state/**`, `.moai/specs/SPEC-CODEMAPS-REFRESH-001/**`(run-phase의 progress.md §E.2 갱신 포함) 에 한정된다. `internal/`, `pkg/`, `cmd/`, `gate.yaml` 변경 1건이라도 있으면 FAIL.

증거: 두 명령의 전체 출력(필터링 금지 — 목록 줄 보존). 측정 창을 base 대비로 고정하는 이유: 커밋 뒤 `git status --porcelain`만 보면 빈 집합이 돼 아무 것도 단정하지 못한 채 통과하는 공허 녹색이 생긴다. 두 명령 모두 단순 파이프라인으로 worktree 격리 가드와 호환(2026-09-02 스모크 실측: 양쪽 무출력 확인).

### AC-CMR-008 — 증거 독립성 (MUST, REQ-CMR-008의 전용 AC)

- **Given** 재생성·재스탬프·게이트 판독이 모두 완료되고
- **When** 런 리포트와 증거 파일을 함께 판독하면
- **Then** 정확성 증거(AC-CMR-002~004의 세 섹션)와 게이트 판정(AC-CMR-006 출력)이 서로 다른 근거로 각각 기록돼 있다. 정확성 섹션 없이 게이트 fresh만 존재하면 §D.6 종결 관문은 열리지 않는다(FAIL).

증거: `.moai/reports/t432/codemaps-accuracy-verification.md` 5섹션의 존재 + progress.md §E.2의 게이트 출력 — 양쪽 모두 확인.

## §D.5 엣지 케이스

| 케이스 | 기대 처리 |
|--------|-----------|
| 재생성이 known-6 팬텀을 재유도 | t304 인계 섹션 기록, 본문 무수정, 합격 판정 불변 |
| 재생성이 신규 부정확 인용을 만듦 | new-findings 기록 + 리드 보고, 본문 무수정 |
| run 중 origin/develop 이동 | merge-base 재계산 후 스탬프 — 그 시점 값이 판정 기준 |
| graph check value가 0이 아니라 1~3 | threshold 미만이면 fresh 유효(기대 0은 목표값이지 진위 기준은 < 40) |
| 다른 카드가 작업 중 described roots를 변경 | run 시작 재측정(§C)으로 흡수, 종결 시점 측정에 판정 귀속 |
| `/moai codemaps --force` 부분 실패 | fail-fast, M1 미종결, 재실행 또는 블로커 보고 |
| 레인이 검사 전에 커밋을 마침 | 측정 창이 base `ad272be20` 대비 합집합이라 공허 통과 없음 (AC-CMR-007) |

## §D.6 종결 관문

AC-CMR-001~008 전부 판정 완료 + 증거 파일 존재 + progress.md §E.2에 검증 명령·출력 기록 — 이 세 가지가 모일 때 본 SPEC은 run-phase를 종결할 수 있다. AC-CMR-002~004 증거 없는 AC-CMR-006 단독 충족은 AC-CMR-008 위반으로서 종결로 인정하지 않는다.

## §D.7 추적성

REQ-CMR-001↔AC-CMR-001 · REQ-CMR-002↔AC-CMR-002 · REQ-CMR-003↔AC-CMR-003 · REQ-CMR-004↔AC-CMR-004 · REQ-CMR-005↔AC-CMR-005 · REQ-CMR-006↔AC-CMR-006 · REQ-CMR-007↔AC-CMR-007 · REQ-CMR-008↔AC-CMR-008(§A 검증 원칙의 전용 AC).
