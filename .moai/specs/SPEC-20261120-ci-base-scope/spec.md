---
id: SPEC-20261120-ci-base-scope
title: "CI reads the scope tool from the base commit"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# CI reads the scope tool from the base commit

## Intent

ci.yml 의 `targets` 작업이 검사 범위(macos 필요 여부)를 정할 때 그 체크아웃(=PR 병합 커밋)의 `tools/ci_scope.py` 를 실행해요. 그래서 PR 이 자기 사본을 고쳐 자기 검사를 줄일 수 있어요 — 스스로를 심판하는 사본이에요. 사용자 결정(2026-11-20, 보호 설정과 같은 방향): CI 는 check-guard 와 같은 방식으로 기준 커밋의 사본을 읽게 해요. PR 의 바뀐 경로는 그대로 PR 에서 나오고, `tools/ci_scope.py` 를 고치는 PR 은 기준 규칙으로 판정돼요(check-guard 가 이미 그 파일에 audit-tier 2 를 요구해요).

이 조각은 `targets` 작업을 기준 커밋 체크아웃으로 바꾸고(check-guard.yml 43-45 행과 같은 모양), 바뀐 경로를 files API로 얻으며(filename + previous_filename, check-guard 가 이름·삭제도 판정하는 것과 같음), 잘림 방어는 기존 실패 안전(전체 매트릭스)을 유지해요. push·dispatch 는 애초 전체 매트릭스라 불변이고, 집계 `ci` 이름과 두 다리 이름도 불변이에요. `tools/ci_scope.py` 는 표준입력만 읽는 자족 도구라 기준 사본 한 파일이면 충분해요.

## Non-goals

- `tools/ci_scope.py` 자체의 분류 로직 변경: 이 조각은 CI 의 읽는 위치만 바꿔요.
- local_check·다른 워크플로: 대상은 ci.yml 의 `targets` 작업 하나예요.
- ruleset·필수 검사 이름 변경: 없음(집계 `ci` 유지).
- Out of scope: 네트워크(러너의 gh 호출은 런타임), 카드 번호의 출력 사용.

## Requirements

- REQ-1: When the workflow runs for a pull request, the `targets` job shall check out the pull request's base commit and run the scope tool from that checkout, so the pull request's own copy of the tool is never executed. · failure: none — the fail-safe of REQ-2 keeps the full matrix while the checkout or the tool cannot run.
- REQ-2: The changed paths the scope decision judges shall come from the pull-request files API, including each file's previous name, with a truncated listing keeping the fail-safe full matrix instead of judging a partial list. · failure: none — the truncated case is the stated outcome.
- REQ-3: The push and dispatch branches of the decision shall keep the full matrix without reading any pull-request data.
- REQ-4: The aggregate `ci` check name and the two check-set leg names shall stay exactly as they are.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a pinned workflow test parses ci.yml and asserts the base checkout, the files-API path listing with the previous names and the truncation guard, and that the push and dispatch branches keep the full matrix.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-2 (REQ-4): the aggregate and leg names are unchanged in the workflow and its tests keep passing.
  verify: `python3 -m unittest tests.github_assets_test tests.ci_scope_test`
- AC-3: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261120-ci-base-scope/spec.md`, `.github/workflows/ci.yml`, `tests/github_assets_test.py`
- read-only: `tools/ci_scope.py`, `.github/workflows/check-guard.yml`
- forbidden: ruleset 변경, 실제 HOME, 카드 번호의 출력 사용

## Plan

1. RED: github_assets_test 에 기준 체크아웃·files API·잘림 방어·불변 이름의 단언을 더해 지금 ci.yml 로 깨져요.
2. GREEN: targets 작업을 기준 커밋 체크아웃 + files API 경로 + 기준 사본 실행으로 고쳐요.
3. 변이(체크아웃 ref 제거, previous_filename 빠짐, 잘림 방어 빠짐) → 비축소 READY 검사 → tier 2 감사는 리더에게(READY 노트가 요청) → READY.

## Risks

- files API 가 3000 파일을 넘으면 목록이 잘려요: 잘림을 감지해 실패 안전(전체 매트릭스)으로 떨어뜨려요 — 현재 포스트처럼.
- `targets` 작업이 gh 를 쓰게 돼요: 권한은 check-guard 와 같은 최소(읽기)로 유지해요.
