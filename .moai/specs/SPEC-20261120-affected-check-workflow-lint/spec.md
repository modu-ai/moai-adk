---
id: SPEC-20261120-affected-check-workflow-lint
title: "The affected check lints a changed workflow file and always selects the help layout test for a help change"
status: draft
revision: 1
created: 2026-11-20
parent: SPEC-20261120-change-claims-and-test-guard
---

# The affected check lints a changed workflow file and always selects the help layout test for a help change

## Intent

오늘 감사에서 워크플로 파일(`release.yml`)의 무효 트리거를 로컬에서 잡을 길이 없었어요. 저장소 어디에도 `actionlint` 호출이 없어서, 워크플로를 바꾼 변경도 READY 전에 문법 검사를 받지 않았어요. 또 도움말 글을 바꾼 변경이 골든 파일 검사(`help_layout`)를 놓친 일이 있었어요. 이 SPEC 은 `moai check affected` 가 (1) 범위가 `.github/workflows/` 아래 파일을 바꾸면 `actionlint` 단계를 더하고 도구가 없으면 통과가 아니라 `not_run` 으로 보고하게 하고, (2) 도움말 글 소스나 도움말 골든 파일을 바꾼 범위에서 도움말 배치 시험을 고르는 것을 시험으로 고정해요.

## Non-goals

- `tools/local_check.sh` 와 CI 단계 목록: 바꾸지 않아요. 그 목록은 CI 두 갈래(leg)와 시험이 그대로 고정하고 있어서, 넣으면 CI 에도 도구 설치가 필요해요.
- `actionlint` 설치: 하지 않아요. 도구가 없으면 `not_run` 으로 보고만 해요.
- `shellcheck`: 부르지 않아요(`-shellcheck=` 로 끄고 돌려요).

## Requirements

- REQ-1: When the range given to `moai check affected` changes a workflow file under `.github/workflows/`, `moai check affected` shall add an `actionlint` step for those files. · failure: none
- REQ-2 (REQ-1): IF `actionlint` is not installed, THEN `moai check affected` shall report that step as `not_run`, never as a pass.
- REQ-3: When the range given to `moai check affected` changes the help text source or a help golden file, `moai check affected` shall select the help layout test. · failure: none

## Acceptance

- AC-1 (REQ-1): 범위가 워크플로 파일을 바꾸면 `--list` 의 단계에 `actionlint` 가 그 파일들과 함께 들어가고, 바꾸지 않으면 들어가지 않아요 · verify: `cargo test --locked --offline --test check_affected_cli actionlint_step`
- AC-2 (REQ-2): `actionlint` 가 PATH 에 없는 기계에서는 그 단계의 결과 단어가 `not_run` 이고 요약의 `result=` 는 `PASS` 가 아니에요 · verify: `cargo test --locked --offline --test check_affected_cli actionlint_missing`
- AC-3 (REQ-3): 골든 파일 하나, 또는 도움말 글 소스만 바꾼 범위에서도 선택에 `help_layout` 이 들어가요(2026-11-20 읽은 `--list`: 골든 4개만 바꾼 범위가 `help_layout` 을 골랐어요) · verify: `cargo test --locked --offline --test check_affected_cli help_layout`

## Scope

- write: 이 SPEC, `src/check_affected.rs`, `tests/check_affected_cli.rs`
- 쥔 카드가 풀 때까지 기다려요(2026-11-20 예약 읽음): `src/check_affected.rs` 와 `src/ops/check.rs` 는 다른 카드가 쥐고 있어요. 리더 결정(노트 4518): 그 카드가 반환하면 이 일을 시작해요
- read-only: `tools/local_check.sh`, `.github/workflows/`
- forbidden: 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 1(이번): SPEC 만. 구현 파일이 풀리기를 기다려요.
2. 조각 2: `tests/check_affected_cli.rs` 에 AC-1~3 의 RED 시험을 먼저 쓰고, `steps()` 에 `actionlint` 단계를 넣어 GREEN. 도구 없음은 프로그램 시작 실패(종료 코드 127)와 구분해 `not_run` 으로 보고해요.
3. 조각 3: 결과 기록(SPEC Result), 변이 시험(단계를 빼면 AC-1 이 실패하는지), READY.

## Risks

- 요약 줄의 `key=value` 단어 순서는 고정이에요(`check affected --help`). 새 단어 `actionlint` 는 단계 순서에 맞는 자리에 넣고, 이 순서를 읽는 시험(`check_affected_cli`)을 함께 고쳐요.
- `not_run` 은 지금 숫자 종료 코드만 나오는 요약에 새로 생기는 값이에요. 이 값을 읽는 쪽(READY 노트를 읽는 도구, `result=` 판정)이 어떻게 세는지 구현 전에 읽어요. `result=PASS` 는 모든 단계가 0 일 때만이라는 말이 있어서 `not_run` 은 PASS 로 세지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 린트는 `tools/local_check.sh` 가 아니라 `moai check affected` 의 단계로 넣어요 | 변경 범위만 보는 이 검사가 "워크플로를 바꾼 카드만"이라는 요청에 맞고, CI 단계 목록은 시험이 고정해요 | `local_check.sh` 에 단계 추가 |
| D-2 | 명령은 `actionlint -shellcheck= <파일들>` 이에요 | 리더 노트(2026-11-20): `shellcheck` 없이 가능해야 해요. `shellcheck` 는 이 기계에 없어요(2026-11-20 `which` 로 읽음) | 옵션을 빼서 `shellcheck` 연동 |

## Result

(아직 구현 전이에요.)
