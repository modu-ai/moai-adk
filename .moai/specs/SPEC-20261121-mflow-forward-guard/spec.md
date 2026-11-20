---
id: SPEC-20261121-mflow-forward-guard
title: "The integration branch only moves forward: a reference-transaction hook and a tool guard with a recorded override"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# The integration branch only moves forward: a reference-transaction hook and a tool guard with a recorded override

## Intent

오늘 한 세션이 primary 에서 로컬 통합 가지를 뒤로 돌렸어요(reflog: 06:47 reset → 06:48 → 06:50 재병합) — 그 사이 열차 후보들이 그 tip 위에 쌓여 있어서 뒤로 이동은 쌓인 작업을 조용히 무효로 만들고 커밋을 잃을 수 있어요. 이 조각은 M-Flow 프로젝트(프로필의 work_flow 가 m-flow, 통합 가지 이름은 같은 프로필의 integration_branch)에서 통합 가지의 뒤로 이동을 두 겹으로 막아요: (1) git reference-transaction 훅 — 설치 시 생성되는 셸 스크립트로, prepared 단계에서 통합 가지의 비조상 갱신과 삭제를 한 줄 사유와 함께 거절해요(된다면 그 위에 되돌림 커밋); (2) 도구 가드 — 통합 가지를 겨냥한 reset --hard/--keep·push --force/--force-with-lease·branch -f/-D 를 명령이 도는 전에 거절해요. 예외는 기록되는 것뿐이에요: 가드 환경 변수에 결정 노트 참조를 담면 통과하고 한 줄을 기록 파일에 남겨요 — 묵음 없음. ff 와 병합 커밋은 그대로 통과해요. 배선(훅 설치 단계와 도구 가드 디스패치)은 점거 중인 파일의 카드 착지 뒤 조각이고, 이 조각은 판정 엔진·스크립트 생성기·규칙·임시 저장소 시험까지를 만들어요.

## Non-goals

- 훅 설치 단계의 hooks.rs 편집과 도구 가드의 디스패치 등록: 그 파일들은 활동 카드가 점거 중 — 착지 뒤 배선 조각.
- 원격 가지의 보호와 거절: 로컬 통합 가지가 이 조각의 대상이에요.
- 프로필의 work_flow 판독 변경: 기존 판독을 그대로 써요.
- 통합 가지가 아닌 참조의 어떤 동작도 바꾸지 않아요.

## Requirements

- REQ-1: The generated reference-transaction hook shall refuse, while the transaction is prepared, an update of the integration branch whose new value is not a descendant of its old value, and it prints a one-line reason that names the forward-only way of a revert commit on top. · failure: none — the refusal is one exit with one line, nothing else is written.
- REQ-2: The hook shall refuse the deletion of the integration branch and allow its creation from the zero value. · failure: none — the zero values are the git conventions for deletion and creation.
- REQ-3: The hook shall allow a fast-forward update and a merge commit on the integration branch, and it leaves every other reference unjudged. · failure: none — the judgement reads the transaction lines and nothing outside them.
- REQ-4: The hook shall treat the update as an exception only when the guard environment variable names a decision note, and it appends one record line with the note reference, the old and new values and the command stamp to the guard log. · failure: none — without the variable the refusal stands.
- REQ-5: The generated script shall carry the integration branch name, the override variable and the guard log path baked at install time, so the guard works in a project where the tool is not on PATH. · failure: none — the script reads no project profile at run time.
- REQ-6: The generator shall give the same script bytes whenever the inputs are the same, and the tests pin those bytes. · failure: none.
- REQ-7: The tool guard shall refuse reset --hard or reset --keep, push --force or --force-with-lease, and branch -f or branch -D while the command names the integration branch, and it leaves a command aimed at another branch alone. · failure: none — the guard judges the command words before the command runs.
- REQ-8: When the project's work flow is not m-flow, the project installs no guard, and the generated script shall judge nothing. · failure: none — the generator is only called for an m-flow project.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-5, REQ-6): in a temp repo with the generated hook installed, update-ref to a non-descendant and a deletion exit non-zero with the reason, a fast-forward and a merge commit and the branch creation succeed, another branch moves freely, and the bytes are pinned.
  verify: `cargo test --locked --offline --test mflow_guard_cli` — the hook scenes
- AC-2 (REQ-4): the override env with a note reference lets the refusal through and appends one record line; without it the refusal stands.
  verify: `cargo test --locked --offline --test mflow_guard_cli` — the override scenes
- AC-3 (REQ-7): the guard analysis refuses the listed commands aimed at the integration branch and passes the same commands aimed at another branch.
  verify: `cargo test --locked --offline --test mflow_guard_cli` — the guard scenes
- AC-4: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-mflow-forward-guard/**`, `src/mflow_guard.rs`, `src/hook_guards.rs`, `tests/mflow_guard_cli.rs`
- read-only: `src/hooks.rs`, `src/hook_engine.rs`, `src/profile.rs`
- forbidden: 훅 설치 단계 편집(점거 중), 원격 가지 보호, 프로필 판독 변경, 카드 번호의 출력 사용

## Plan

1. RED: 임시 저장소 시험을 먼저 써요 — 생성기 함수와 규칙 함수가 없으면 컴파일이 실패하는 것이 첫 실패예요.
2. GREEN: src/mflow_guard.rs 의 스크립트 생성기(베이크·조상 검사·삭제·override 기록)와 src/hook_guards.rs 의 명령 규칙을 채워요.
3. 배선(훅 설치 단계·도구 가드 등록)은 점거 파일의 카드 착지 뒤 별도 조각 — 이 조각의 시험은 훅 스크립트를 임시 저장소 .git/hooks 에 직접 놓고 git 명령으로 거절·통과·override 를 검증해요.
4. READY 전체 검사 → 감사는 리더에게 → READY. 뮤테이션은 조상 검사 제거·삭제 허용·override 묵음화 등 상한 안에서.

## Risks

- 훅은 prepared 단계에서만 거절해요: committed 와 aborted 단계는 그대로 통과시켜야 트랜잭션이 꼬이지 않아요 — 시험이 세 단계를 모두 잡아요.
- 생성 시점에 가지 이름을 베이크하므로 프로필의 통합 가지가 바뀌면 재설치가 필요해요: 스크립트 헤더가 생성 출처를 밝혀 다음 설치 계획이 편집·구버전을 알아차려요.
- override 는 환경 변수라 세션 간 새는 것을 못 막아요: 기록 줄이 유일한 흔적이므로 READY 노트와 문서가 변수 이름과 기록 자리를 명시해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 훅 본체를 자급 셸 스크립트로 생성한다 — 실행 때 도구가 PATH 에 없어도 가드가 동작하고, 설치 계획은 점거 중인 hooks.rs 를 만지지 않는다. | 카드 본문 (1), 노트 4610 파일 지도 | 카드 열림 |
| D-2 | 조상 검사는 git merge-base --is-ancestor 로 한다 — 훅이 도는 cwd 는 작업 트리 루트라 같은 저장소 판정이에요. | git reference-transaction 훅 규약 | 카드 열림 |

## Result

(status draft — written when the slice lands)
