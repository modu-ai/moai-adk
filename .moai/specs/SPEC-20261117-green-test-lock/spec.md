---
id: SPEC-20261117-green-test-lock
title: "GREEN 동안 시험 파일 잠금 — 설계"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

카드 C-86(사용자 지시 2026-11-17 "권고 말고 강제")의 세 번째 규칙 "GREEN 동안 시험 파일 잠금"의 설계예요(코드 없음). 조사
제안 B(같은 날 사용자 승인: "RED 기록, GREEN 동안 시험 읽기 전용, 시험 약화 감지")와 조사 근거(에이전트는 실패하는 시험을
지우거나 고치고, "속이지 마라"는 지시는 효과가 거의 없음 — ImpossibleBench, METR 2025)를 따라요.

지금 RED/GREEN이 남는 곳(조사, 기준 `ae6512a691`):

- `src/run.rs`: 변경 전략 `tdd`의 증거 단계 `red`·`green`·`regression`, `red_confirmed`(실패가 모두 assertion 실패일 때만
  RED 인정), `oracle_changes`(`tests/`, `fixtures`, `golden`, `*_test.py` 경로). 내부 `run-verify`가 영수증을 stdout으로만
  내고 저장하지 않아요. 단계가 지금 어디인지 아는 기록은 없어요.
- `src/self_audit.rs`: PR 본문 `## Self-Audit`의 `red:` 줄(처음 실패한 시험 또는 없는 이유) — 사람이 쓰는 글이에요.
- 시험 약화: `internal/quality`가 "assertion을 빼거나 skip을 더하거나 지운 시험은 지적 대상"이라고 적었고, 감사가 손으로 봐요.
- hook: `hook_guards::write_guard`(tool-preflight 기본 규칙 A1–A4, 1조각 A5는 PR #3173에서 감사 중)가 파일 쓰기를 판단해요.
  RED/GREEN을 아는 규칙은 없어요.
- 카드: 상태 `active`·`review`는 단계가 아니에요. SPEC `## Result`의 RED 문장도 사람이 쓴 글이에요.

그래서 hook이 "지금 GREEN"을 알 방법이 없어요. 이 SPEC은 그 기록과, 그 기록으로 hook이 시험 파일 쓰기를 막는 규칙, hook이 못 보는
것을 감사가 잡는 규칙을 정해요.

## Non-goals

- 구현(다음 조각들, 아래 Plan).
- Bash 리디렉션이나 hook을 부르지 않는 호스트의 쓰기를 막는 일: 감사(REQ-9, REQ-10)가 그물이에요.
- 시험이 맞는 것을 시험하는지 판단(mutation 하한은 제안 B의 다른 항목).
- 언어마다 시험 실행 명령을 정하는 일(`internal/quality`와 언어 가이드가 이미 정함).

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When `moai tdd red` is run on a SPEC work branch, the tool shall record the SPEC id, the RED commit and the locked test paths in the worktree's own git folder.
- REQ-2 (REQ-1): IF the head commit has no `Moai-Phase: red` trailer, changes no test path, or its named tests do not fail with an assertion failure, THEN `moai tdd red` shall refuse and record nothing.
- REQ-3: While the phase record is GREEN or REFACTOR, tool-preflight shall refuse a file write to a locked test path or to a new path that matches a test pattern.
- REQ-4: While the phase record is GREEN or REFACTOR, tool-preflight shall allow every other file write.
- REQ-5: When `moai tdd green` finds the recorded RED tests passing at the head, the tool shall move the record to REFACTOR and keep the lock.
- REQ-6 (REQ-5): IF any recorded RED test does not pass at the head, THEN `moai tdd green` shall refuse and keep the record at GREEN.
- REQ-7: When `moai tdd reopen --reason <text>` is run, the tool shall end the lock and append the reason to the phase log that the audit reads. · failure: none
- REQ-8: When `moai tdd done` is run, the tool shall remove the phase record and keep the phase log. · failure: none
- REQ-9: The audit shall list every test-pattern path that changed between the RED commit and the audited head, with the reasons the phase log holds.
- REQ-10: The audit shall report each added skip marker and each removed assertion line in any file changed between the RED commit and the audited head.
- REQ-11: The test patterns shall be one table per detected language plus the test paths the project declares.
- REQ-12: Tool-preflight shall read the phase record from the file system without starting a program.

실패 경로: REQ-1 → REQ-2, REQ-5 → REQ-6(괄호로 연결). REQ-7, REQ-8 → 없음(줄 끝 `failure: none`).

## Acceptance

시험 파일 `tests/green_test_lock.rs`와 `tests/tdd_cli.rs`는 구현 조각에서 만들어요(지금은 없어요).

- AC-1 (REQ-1, REQ-2): RED 기록이 trailer·시험 경로·assertion 실패를 모두 갖출 때만 생기고, 하나라도 빠지면(N+1) 아무것도
  쓰지 않아요.
  verify: `cargo test --locked --offline --test tdd_cli red_needs_the_trailer_a_test_path_and_an_assertion_failure`
- AC-2 (REQ-3, REQ-4, REQ-12): GREEN에서 잠긴 시험 파일과 새 시험 패턴 파일 쓰기는 거부, 그 밖의 쓰기는 허용, 기록이 없으면 이
  규칙은 아무것도 막지 않아요.
  verify: `cargo test --locked --offline --test green_test_lock locked_tests_refuse_other_writes_pass`
- AC-3 (REQ-5, REQ-6): RED 시험이 하나라도 실패하면 GREEN에 머물고, 모두 통과하면 REFACTOR로 가며 잠금은 그대로예요.
  verify: `cargo test --locked --offline --test tdd_cli green_needs_every_red_test_passing`
- AC-4 (REQ-7, REQ-8): reopen은 이유를 남기고 잠금을 풀고, done은 기록을 지우고 로그를 남겨요.
  verify: `cargo test --locked --offline --test tdd_cli reopen_and_done_keep_the_log`
- AC-5 (REQ-9, REQ-10): RED 커밋과 head 사이의 시험 경로 변경, 새 skip, 지운 assertion을 감사 보고가 이름으로 적어요.
  verify: `cargo test --locked --offline --test tdd_cli audit_lists_test_changes_skips_and_dropped_assertions`
- AC-6 (REQ-11): 감지기 언어마다 패턴 행이 있고, 프로젝트가 선언한 시험 경로가 더해져요.
  verify: `cargo test --locked --offline --test green_test_lock every_detected_language_has_test_patterns`

## Scope

- write(구현 조각): `src/tdd.rs`(새), `src/hook_guards.rs`, `src/main.rs`(명령 `tdd`), `src/run.rs`(시험 패턴 공유),
  `tests/green_test_lock.rs`, `tests/tdd_cli.rs`, `skills/moai/workflows/run.md`·`audit.md` 한 줄씩,
  `skills/moai/references/internal/quality.md`, `.moai/specs/SPEC-20261117-green-test-lock/spec.md`
- write(이번 조각): `.moai/specs/SPEC-20261117-green-test-lock/spec.md`
- read-only: `src/run.rs`, `src/self_audit.rs`, `src/project_detect.rs`, `skills/moai/references/languages/**`
- forbidden: 사용자 HOME, 호스트 설정 파일, 공유 원격 기록 고쳐 쓰기(force-push)

## Plan

1. 설계(이 SPEC). 아래 질문 3개를 사용자 선택으로 받아요.
2. 시험 패턴 표: `run::oracle_changes`를 언어별 표(D-4)로 넓혀 `src/tdd.rs`로 옮기고 run.rs가 같은 표를 써요.
3. `moai tdd red|green|reopen|done|status`: 기록과 로그(D-2), RED 확인은 기존 `run::red_confirmed` 판단을 써요.
4. hook: `write_guard`의 A6 규칙(REQ-3, REQ-4, REQ-12).
5. 감사: `moai tdd audit --red <commit> --head <commit>`이 REQ-9, REQ-10 보고를 내고, `audit.md`가 그 보고를 읽게 해요.
6. 관문(뒤 조각): 변경 전략이 `tdd`인 SPEC의 PR에 RED 기록(trailer 커밋)이 없으면 관문이 멈춤.

## Risks

- 기록이 git 폴더 안이라 다른 worktree나 다른 기계에서는 보이지 않아요 → 공유 기록은 RED 커밋의 trailer이고, 감사는 그것으로
  판단해요(hook 기록은 그 세션의 강제용).
- 에이전트가 `moai tdd red`를 부르지 않으면 잠금도 없어요 → 6단계 관문과 감사가 RED 기록 없는 TDD 변경을 잡아요.
- Rust처럼 시험이 원본 파일 안(`#[cfg(test)] mod tests`)에 있으면 파일 단위 잠금이 안 돼요 → REQ-10의 감사(지운 assertion,
  새 `#[ignore]`)가 맡아요.
- 패턴이 너무 넓으면 시험이 아닌 파일(`fixtures`라는 이름의 기능 코드)을 막아요 → 프로젝트 선언(D-4)으로 빼거나 더해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | GREEN을 아는 방법: (A) `moai tdd red`가 쓰는 기록 파일 — 추천 · (B) 가지의 RED trailer 커밋을 hook이 git으로 읽음 · (C) 카드 상태 | A는 hook이 프로그램 없이 읽음(REQ-12, A5와 같은 비용); B는 쓰기마다 git 실행; C는 단계가 아님 | 기록 대신 trailer만 읽게 개정 |
| D-2 | 기록 위치: (A) worktree 자기 git 폴더 `<git-dir>/moai/tdd.json`과 로그 `tdd-log.jsonl` — 추천 · (B) 작업 트리 `.moai/state/` · (C) 커밋에만 | A는 커밋되지 않고 worktree마다 따로, 다른 세션과 섞이지 않음; 공유 증거는 RED 커밋의 `Moai-Phase: red` trailer | 위치 개정 |
| D-3 | 잠그는 범위: (A) RED 커밋이 바꾼 시험 경로 + 새로 만드는 시험 패턴 파일 — 추천 · (B) 모든 시험 패턴 파일 · (C) RED 커밋이 바꾼 경로만 | A는 "실패하는 시험을 고치지 않음"과 "GREEN에 새 시험 없음(새 시험은 새 RED)"을 함께 지킴; B는 관계없는 시험 수리까지 막음 | 범위 개정 |
| D-4 | 시험 패턴: 언어 가이드와 같은 감지 언어별 표(Rust `tests/**`, `benches/**`; Python `tests/**`, `test_*.py`, `*_test.py`, `conftest.py`; JS/TS `*.test.*`, `*.spec.*`, `__tests__/**`, `test/**`, `tests/**`; Go `*_test.go`, `testdata/**`; Java/Kotlin `src/test/**`; Ruby `spec/**`, `test/**`; PHP `tests/**`, `*Test.php`; C/C++ `test/**`, `tests/**`; C# `*.Tests/**`, `*Tests/**`; Swift `Tests/**`; 공통 `fixtures/**`, `golden/**`) + 프로젝트 선언(프로필의 `test_paths`) | 기존 `run::oracle_changes`를 넓힘; 언어 가이드에 한 줄씩 더함 | 표 개정 |
| D-5 | REFACTOR에도 잠금 유지, `done`에서 해제 | Beck: 구조 변경 전후에 같은 시험을 돌림 | REFACTOR에서 해제로 개정 |
| D-6 | [NEEDS CLARIFICATION: GREEN 중 시험을 고쳐야 할 때: (A) `moai tdd reopen --reason`으로 GREEN을 끝내고 새 RED부터 — 추천 · (B) `moai tdd unlock <path> --reason`으로 파일 하나만 풀기] | 사용자 선택 대기. A는 이유가 늘 남고 RED가 다시 확인됨; B는 빠르지만 잠금이 약해짐 | — |
| D-7 | [NEEDS CLARIFICATION: GREEN 중 새 시험 파일: (A) 거부(새 시험 = 새 RED) — 추천 · (B) 허용(기존 시험만 잠금)] | 사용자 선택 대기 | — |
| D-8 | [NEEDS CLARIFICATION: 강제 범위: (A) 변경 전략이 `tdd`인 SPEC의 PR은 RED 기록이 있어야 관문 통과(6단계) — 추천 · (B) 에이전트가 `moai tdd red`를 부를 때만 잠금, 관문은 보지 않음] | 사용자 선택 대기. A가 "권고 말고 강제"에 맞지만 모든 TDD 변경에 RED 커밋을 요구함 | — |
| D-9 | RED 확인은 `run::red_confirmed`와 같은 판단(실패가 모두 assertion 실패, 컴파일·네트워크·0개 선택은 RED 아님) | `internal/tdd-results`의 `EXPECTED_RED`와 같음 | — |
