---
id: SPEC-GRAPH-STAMP-ANCESTRY-001
title: "Graph codemaps 스탬프 조상성 선판정과 push 가드 종결"
version: "0.1.1"
status: draft
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph, internal/cli, .github/workflows/graph-freshness.yml, .moai/project/codemaps"
lifecycle: spec-anchored
tags: "graph, codemaps, provenance, stamp, ancestry, freshness, workflow, issue-1661"
era: V3R6
tier: M
issue_number: 1661
related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002, SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-STAMP-REACHABILITY-001, SPEC-GRAPH-GATE-RESTAMP-001, SPEC-CODEMAPS-REFRESH-002]
---

# SPEC-GRAPH-STAMP-ANCESTRY-001 — Graph codemaps 스탬프 조상성 선판정과 push 가드 종결

## HISTORY

| Version | Date | Change | Author |
|---|---|---|---|
| 0.1.1 | 2026-09-13 | 중단된 plan 단계를 재개해 실제 checker·CLI·workflow·provenance producer와 관련 완료 SPEC 6개를 대조했다. 선행 SPEC의 두 문구 충돌을 명시적으로 판정했다. 오류 경로의 `VerdictAbsent`는 substantive freshness verdict가 아닌 호환 운반체로 한정하고, push의 object-only 허용은 이 SPEC의 `HEAD` 조상성 검사로 대체한다. release PR의 merge-preview `HEAD` 분기는 현재 workflow의 t407 후속 결정을 보존한다. | manager-spec |
| 0.1.0 | 2026-09-13 | 카드 t688에서 최초 작성. Git 위상 비교 가능성과 freshness 값 초과를 서로 다른 실패 종류로 고정하고, checker 선판정·push 가드·진짜 codemaps 재생성을 한 실행 계약으로 묶었다. GH issue #1661은 발생 이력과 관계만 기록하며 회신·종료는 리드 소관으로 남겼다. | manager-spec |

## §A. 문제 진술

`graph-freshness`에는 서로 다른 두 적색 원인이 있다.

1. **스탬프 도달 불가**: 저장된 clean `commit_sha` 객체가 로컬에 존재하더라도 현재 checkout `HEAD`의 조상이 아니면 두 트리는 같은 이력 창에서 비교할 수 없다. squash는 파일을 보존하면서 원래 커밋을 부모로 남기지 않고, rebase는 커밋 식별자를 다시 쓰므로 이 상태를 만들 수 있다.
2. **freshness 값 초과**: 스탬프가 현재 checkout의 조상이라 비교 가능하더라도, `described-source-diff`가 임계값 40 이상이면 codemaps가 실제로 stale이다. 이 경우에는 재생성과 도달 가능한 스탬핑이 필요하다.

현재 checker는 첫 번째 조건을 freshness 계산보다 먼저 판별하지 않는다. 이 때문에 **객체는 있으나 조상이 아닌 스탬프**를 숫자 값이 있는 stale 결과처럼 표현할 수 있다. 이는 측정 실패를 측정값 초과로 바꾸며, 잘못된 복구인 맨손 재스탬프를 유도한다.

워크플로에도 같은 틈이 있다. `pull_request`에서는 ordinary PR의 `origin/<base_ref>`와 `release/*`의 merge-preview `HEAD`를 구분하지만, `push`는 객체 존재만 확인하고 통과한다. 따라서 `origin/develop`에 객체가 남아 있어도 `origin/main`이 그 커밋을 조상으로 갖지 않는 상태가 main push에서 초록으로 빠질 수 있다.

## §B. 목표와 범위

### §B.1 목표

- **M1**: clean 스탬프의 객체 존재와 checkout `HEAD` 조상성을 freshness 계산보다 먼저 판별한다. 객체가 존재하지만 조상이 아니면 freshness는 미측정이며, `VerdictAbsent` 운반체와 system error/exit 2로 보고한다.
- **M2**: `push` 이벤트가 clean 스탬프의 조상성을 checkout `HEAD`에 대해 검사하도록 워크플로 가드를 닫는다. 기존 ordinary PR와 `release/*` merge-preview 대상 선택은 유지한다.
- **M3**: 코드 변경 뒤 codemaps 생성 계약의 5개 문서를 실제로 재생성하고 해당 실행 경로의 판정 대상에서 도달 가능한 커밋에 스탬핑해, 위상 비교 가능 상태에서 독립 freshness 값이 40 미만으로 돌아오는 것을 증명한다. `docs-truth.md`는 생성기 산출물이 아니므로 에이전트 카탈로그 사실이 바뀐 경우에만 별도 수동 갱신하며, 본 SPEC의 graph 동작 변경만으로는 재생성됐다고 주장하지 않는다.

### §B.2 핵심 판정

도달 불가의 처분은 기존 계약과의 충돌을 최소화한다. `VerdictAbsent`를 오류 경로의 보고 운반체로 유지하되, reason과 CLI 출력은 **codemaps 본문이 없다는 뜻이 아니라 freshness를 측정하지 못했다는 뜻**을 명시한다. 숫자 `value`, contribution, content anchor, driving paths를 만들지 않으며 system error/exit 2로 fail closed 한다.

### §B.3 선행 계약 충돌의 판정

- `SPEC-V3R6-GRAPH-FRESHNESS-001` REQ-GF-004의 “fresh·stale·absent 중 어느 verdict도 없음”은 **측정된 freshness 의미를 부여하지 말라**는 제약으로 계승한다. 현재 오류 경로가 이미 운반하는 `VerdictAbsent`는 substantive verdict가 아니라 partial report의 호환 표식으로만 유지하며, non-nil system error와 exit 2가 일반 absent/exit 1과 구분한다.
- `SPEC-STAMP-REACHABILITY-001` §B.2와 그 acceptance의 push object-only 성공은 이 SPEC의 REQ-GSA-008이 **push 분기에 한해 대체**한다. anchorless provenance의 skip-with-reason과 missing-object failure는 그대로 보존한다.
- `release/*`의 merge-preview `HEAD` 판정은 선행 SPEC 원문 이후 t407에서 현재 workflow에 추가된 후속 계약이다. 이 SPEC은 그 분기를 재설계하지 않고 회귀 잠금한다.

## §C. 요구사항 (GEARS)

### REQ-GSA-001 — 조상성 선판정 (Ubiquitous)

Clean `commit_sha`가 있는 codemaps checker는 어떤 freshness 값이나 content anchor를 계산하기 전에 저장 스탬프가 현재 checkout `HEAD`의 조상인지 판별해야 한다(shall).

### REQ-GSA-002 — 객체 존재·비조상 처분 (When, event-detected)

**When** 저장 스탬프 객체가 존재하지만 현재 checkout `HEAD`의 조상이 아님이 감지되면, checker는 codemaps 층을 `VerdictAbsent` 운반체로 보고하고 freshness 미측정 reason과 non-nil system error를 반환해야 한다(shall). CLI는 exit 2를 반환해야 한다(shall).

### REQ-GSA-003 — `VerdictAbsent` 의미의 명시 (Ubiquitous)

비교 불가 경로의 `VerdictAbsent`는 backward-compatible report carrier일 뿐 codemaps 본문 부재의 증거로 해석되어서는 안 되며(shall not), reason은 `unreachable stamp`와 `freshness unmeasured` 의미를 구분 가능하게 담아야 한다(shall).

### REQ-GSA-004 — 미측정 값 비생성 (Unwanted)

조상성 판정이 실패한 checker와 CLI는 숫자 freshness `value` 또는 `threshold` 비교 결과, contribution, content anchor, driving paths를 출력하거나 JSON에 측정 결과로 직렬화해서는 안 된다(shall not).

### REQ-GSA-005 — 조치 가능한 복구 안내 (When, event-driven)

**When** 도달 불가 스탬프 오류가 보고되면, CLI 경계는 codemaps 본문을 실제로 재생성한 뒤 checkout에서 도달 가능한 커밋으로 스탬핑하라는 안내를 제공해야 하며(shall), 자동 재스탬프나 본문 없는 맨손 재스탬프를 수행하거나 권해서는 안 된다(shall not).

### REQ-GSA-006 — 도달 가능한 stale 유지 (While, state-driven)

**While** clean 스탬프가 현재 checkout `HEAD`의 조상이고 `described-source-diff`가 40 이상이면, codemaps 층은 숫자 값과 driving-path 귀속을 포함한 `stale` 결과를 유지하고 CLI는 exit 1을 반환해야 한다(shall).

### REQ-GSA-007 — 도달 가능한 fresh 유지 (While, state-driven)

**While** clean 스탬프가 현재 checkout `HEAD`의 조상이고 `described-source-diff`가 40 미만이면, codemaps 층은 `fresh` 결과를 유지하고 다른 층도 fresh인 경우 CLI는 exit 0을 반환해야 한다(shall).

### REQ-GSA-008 — push 조상성 가드 (When, event-driven)

**When** `graph-freshness`가 `push` 이벤트로 실행되고 provenance가 clean `commit_sha`를 담고 있으면, 워크플로는 객체 존재만이 아니라 해당 스탬프가 checkout `HEAD`의 조상인지 검사해야 하며(shall), 객체가 다른 ref에 남아 있어도 비조상 상태는 job failure로 끝나야 한다(shall).

### REQ-GSA-009 — PR 대상 선택 보존 (While, state-driven)

**While** `graph-freshness`가 `pull_request` 이벤트로 실행되면, `release/*`는 merge-preview `HEAD`, ordinary PR는 `origin/<base_ref>`를 비교 대상으로 계속 사용해야 한다(shall).

### REQ-GSA-010 — merge/squash/rebase 회귀 표 (Where, capability gate)

**Where** 조상성 가드의 회귀 검증이 실행되면, 제어된 로컬 이력 fixture는 merge-commit 계보, squash 계보, rebase-like rewritten 계보를 모두 포함해야 하며(shall), merge-commit에서만 원래 스탬프 조상성이 유지되고 나머지 두 계보는 객체가 존재해도 거부됨을 증명해야 한다(shall).

### REQ-GSA-011 — 진짜 재생성 종결 (When, event-driven)

**When** M1과 M2의 코드·워크플로 변경 뒤 codemaps가 stale이면, M3는 codemaps 본문을 실제 변경 내용에 맞게 재생성한 뒤 checkout에서 도달 가능한 커밋으로 스탬핑해야 하며(shall), 그 결과 clean 스탬프 조상성 검사가 통과하고 `described-source-diff`가 40 미만임을 함께 증명해야 한다(shall).

### REQ-GSA-012 — 기존 freshness 계약 보존 (Ubiquitous)

임계값 40, described-worthy 필터, dirty fingerprint 경로, content-anchor anti-false-green, contribution과 driving-path 귀속 계약은 유지되어야 한다(shall). 본문을 갱신하지 않은 맨손 재스탬프는 stale을 fresh로 바꿔서는 안 된다(shall not).

## §D. 판정 순서와 종료 의미

checker는 아래 표를 위에서 아래로 적용한다. freshness 값은 비교 가능한 행에서만 계산한다.

| 순서 | provenance/위상 상태 | freshness 계산 | report 의미 | CLI 종료 |
|---|---|---|---|---|
| 1 | clean 스탬프 객체를 해석할 수 없음 | 하지 않음 | `VerdictAbsent`; stamp unresolved, freshness unmeasured | 2 |
| 2 | clean 스탬프 객체는 존재하지만 checkout `HEAD`의 조상이 아님 | 하지 않음 | `VerdictAbsent`; stamp unreachable, freshness unmeasured | 2 |
| 3 | clean 스탬프가 checkout `HEAD`의 조상 | `described-source-diff` 계산 | 값 ≥40은 stale, 값 <40은 fresh | stale 1 / 전체 fresh 0 |
| 4 | 유효한 dirty fingerprint anchor | 기존 fingerprint 비교만 수행 | mismatch는 stale, match는 fresh | stale 1 / 전체 fresh 0 |

`VerdictAbsent`라는 문자열만으로 1·2행과 실제 본문 부재를 합치지 않는다. reason과 system-error 유무가 의미를 구분한다.

## §E. 제약

- 새 외부 의존성, provenance schema 변경, 새 verdict enum은 도입하지 않는다.
- `VerdictAbsent` 운반체와 0/1/2 종료코드 구조는 유지한다.
- 임계값 40과 `described-source-diff` 토큰은 바꾸지 않는다.
- workflow의 `release/*` merge-preview와 ordinary PR base 판정을 약화하지 않는다.
- controlled fixture는 로컬 임시 저장소에서 계보만 합성한다. 실제 release/main 병합이나 원격 ref 변경은 수행하지 않는다.
- 로컬 검증은 영향받는 package로 한정한다. `go test ./...`는 실행하지 않으며 전체 스위트 판정은 `origin/develop` CI가 맡는다.
- issue #1661의 comment/close, push, PR, merge는 이 SPEC plan/run 산출물의 소관이 아니다.

## §F. 성공 기준

1. 객체가 존재하지만 checkout `HEAD`의 조상이 아닌 fixture에서 checker와 CLI가 숫자 freshness 주장 없이 exit 2를 반환한다.
2. 같은 위상에서 push guard가 실패한다.
3. 도달 가능한 stale fixture는 숫자 값과 귀속을 유지한 채 exit 1이다.
4. 진짜 재생성 뒤 도달 가능한 스탬프와 `described-source-diff < 40`을 함께 관측한다.
5. 기존 bare-restamp, dirty fingerprint, content-anchor, described-worthy 회귀 검사가 통과한다.

## §G. 제외 범위

### Out of Scope — 자동 또는 맨손 재스탬프

- 도달 불가나 stale을 자동으로 재스탬핑하지 않는다.
- codemaps 본문을 재생성하지 않은 채 provenance만 다시 쓰는 복구를 추가하거나 권하지 않는다.

### Out of Scope — 병합 정책 변경

- release/hotfix의 merge-commit 원칙을 squash 또는 rebase로 바꾸지 않는다.
- GitHub의 실제 merge를 실행해 회귀를 검증하지 않는다. 로컬 fixture만 사용한다.

### Out of Scope — freshness 알고리즘 재설계

- 임계값 40 재산정, described-worthy predicate 변경, contribution 게이팅, 새 freshness metric을 다루지 않는다.
- mx-index, edges, citations 층의 판정 로직을 바꾸지 않는다.

### Out of Scope — provenance schema와 과거 이력 수선

- `schema_version`, `commit_sha`, `dirty`, `content_fingerprint`, `described_roots` 필드를 추가·제거·재해석하지 않는다.
- 과거 squash/rebase 이력을 다시 쓰거나 이미 고립된 커밋을 main 계보에 삽입하지 않는다.

### Out of Scope — GitHub issue 운영

- GH issue #1661에 comment를 게시하거나 issue를 close하는 행위는 리드 소관이다.

## §H. 교차 참조

- `SPEC-V3R6-GRAPH-FRESHNESS-001` — 0/1/2 freshness 기본 계약
- `SPEC-V3R6-GRAPH-FRESHNESS-002` — not-comparable system-error와 회귀 잠금
- `SPEC-GRAPH-FRESHNESS-CADENCE-001` — described-worthy와 임계값 40
- `SPEC-STAMP-REACHABILITY-001` — object/reachability guard와 명시적 commit 스탬핑
- `SPEC-GRAPH-GATE-RESTAMP-001` — content-anchor와 bare-restamp anti-false-green
- `SPEC-CODEMAPS-REFRESH-002` — genuine codemap regeneration 관계
