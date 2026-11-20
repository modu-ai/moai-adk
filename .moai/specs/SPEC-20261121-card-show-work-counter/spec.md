---
id: SPEC-20261121-card-show-work-counter
title: "Card show's note-events work is counted, not clocked, and grows linearly"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# Card show's note-events work is counted, not clocked, and grows linearly

## Intent

오늘 리더 배차의 요구는 둘이에요 — (1) card show 의 노트 질의를 한 번의 조회로, (2) 노트 수를 늘려도 작업량이 선형인지 작업량 카운터 시험으로 증명. (1)은 이미 이뤄져 있어요: agent-4 의 head 1aa4b624b0 이 로컬 tip 에 병합돼 있고(git merge-base 로 확인), NOTE_EVENTS_SQL 이 서브쿼리 없는 단일 문장이에요(src/card/command.rs:191-193). 남은 것은 (2)예요. 기존 시험 the_note_events_are_read_once(src/card/command.rs:512)는 EXPLAIN QUERY PLAN 텍스트로 모양만 고정하지, SQLite 가 실제 스캔한 일은 재지 않아요 — 플랜 텍스트를 못 읽는 회귀(per-note 재스캔의 다른 모양)를 잡지 못해요. 이 조각은 시험 소유 연결에 SQLite progress handler 를 설치해 노트·이벤트 질의 하나가 소비한 가상 머신 진행 스텝(실제 스캔 일에 비례)을 세고, 노트 32·64 개 배증에서 work(2n) <= 3 x work(n) 을 단언해요. 옛 per-note 재스캔 모양은 배증에 약 4배라 이 바운드가 죽여요. 시계를 재지 않으니 부하에서 흔들리지 않아요(C-213 방식의 연산 횟수 계측 — cfg(test) 계측은 제품 빌드에 없음).

## Non-goals

- note_sessions·show 의 제품 동작 변경: 측정은 시험 쪽에서만 하고 제품 코드 본문은 그대로.
- tests/card_show_notes_cli.rs 변경: 세션 characterisation 시험은 이번 조각과 무관.
- 인덱스 추가 같은 스키마 변경: 원장 스키마 v1 유지(카드 공통 지킴 사항).

## Requirements

- REQ-1: The work meter of card show's note-events query shall count the virtual-machine progress steps of exactly one note_sessions call, installed on the test-owned connection before the call and removed after it returns. · failure: none — the handler exists in the test build only, and the removal runs on the line right after the call.
- REQ-2: The work test shall pass when doubling a card's note count at most triples the counted work, so a per-note rescan of the event table, which about quadruples the work on doubling, fails the test. · failure: none — both sizes run on the same in-memory schema, so the ratio does not move with machine load.

## Acceptance

- AC-1 (REQ-1, REQ-2): the work test passes on the landed single-read shape, and the per-note rescan mutation, one extra full scan of the note-events query per matched session, blows past the triple bound and fails the test.
  verify: `cargo test --lib the_note_events_work_grows_linearly` — the test name in the output
  verify: the mutation run recorded on the card's READY note
- AC-2: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-card-show-work-counter/spec.md`, `Cargo.toml`, `src/card/command.rs`
- read-only: `tests/card_show_notes_cli.rs`
- forbidden: note_sessions·show 제품 동작 변경, 스키마 변경, rusqlite 기능 추가 외 의존성 변경, 카드 번호의 출력 사용

## Plan

1. RED: 뮤테이션(note_sessions 가 맺은 세션 하나당 노트·이벤트 질의를 한 번 더 훑는 per-note 재스캔 복원)을 잠깐 얹어 배증 바운드가 실패하는 것을 기록해요.
2. GREEN: 현재 tip 코드에서 시험이 통과해요 — 계측은 시험이 소유한 메모리 연결의 progress handler 로, 제품 코드는 그대로.
3. 빌트 바이너리 spec check 0/0 → 예약 정리 → 감사는 리더에게 → READY.

## Risks

- progress_handler 는 &mut Connection 을 요구해(rusqlite 0.43 hooks/mod.rs:440) 제품 함수 안에는 설치할 수 없고 hooks 기능 뒤에 있어(lib.rs:135): 계측을 시험 소유 연결에 두고 Cargo.toml 의 rusqlite 에 hooks 기능 한 단어를 더하는 것으로 풀었어요. 컴파일되는 래퍼만 늘어나고 도구가 부르는 곳은 없어 제품 동작은 변하지 않고, Cargo.lock 도 그대로예요(--locked 안전).
- 대안인 Statement::get_status(VmStep) 는 문장 하나의 스텝만 세는 공개 API 가 없어(pub(crate), raw_statement.rs:266) 썼다면 제 함수에 질의가 하나 더 생기는 회귀를 못 봤을 거예요 — 연결 전체를 세는 progress handler 가 그 빈틈을 닫아요.
- 바운드 3배는 선형(약 2배)과 이차 회귀(약 4배) 사이 여유 값이에요: 질의 준비 같은 상수 비용이 두 크기에서 같아 비율을 밀지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 계측은 제품 코드가 아니라 시험 소유 연결의 progress handler 로 하고, 그 컴파일을 위해 rusqlite 의 hooks 기능을 켜요 — note_sessions 는 참조만 받아 제품 쪽 설치가 불가능하고, 연결 전체를 세는 측정이 재스캔 회귀의 어떤 모양도 잡아요. | rusqlite 0.43 hooks/mod.rs:440 시그니처, lib.rs:135 기능 게이트 | 카드 열림 |

## Result

- AC-1: met — `card::command::tests::the_note_events_work_grows_linearly_with_the_note_count` passes on the landed single-read shape in 0.01s: 노트 32·64 개 배증에서 카운터는 339 → 659(1.94배 — 바운드 3배 이내). 계측은 시험 소유 연결의 progress handler(1스텝마다 콜백)가 그 호출 사이에 센 가상 머신 진행 스텝이고, 호출 직후 떼어 한 호출의 일만 재요(REQ-1).
- AC-2: met — `cargo test --locked --offline --bins` 28 passed, `cargo test --locked --offline --lib` 1656 passed(0 failed, 2026-11-20 20:4xZ, c143 워크트리). fmt·clippy(lib+tests) 경고 0, `spec check --all` 오류 0(전체 경고 9는 기존 SPEC 것).
- 뮤테이션(per-note 재스캔 복원 — 맺힌 세션 하나마다 같은 노트·이벤트 질의를 다시 준비해 훑음): 179,331 → 1,371,379(7.6배)로 바운드를 넘어 시험 실패. 커밋 전 상태에서 확인하고 `git checkout` 으로 원복했어요 — 커밋 내용에는 뮤테이션이 없어요.
- Not changed: `note_sessions`·`show` 의 제품 본문, 스키마, 인덱스, `journal_mode`; `Cargo.toml` 은 rusqlite 의 `hooks` 기능 한 단어만(계측 컴파일용 — 부르는 곳 없음, Cargo.lock 불변).
