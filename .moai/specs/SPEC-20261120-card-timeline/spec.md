---
id: SPEC-20261120-card-timeline
title: "moai card timeline reads one card's events without writing and without making writers wait"
status: draft
revision: 1
created: 2026-11-20
---

# moai card timeline reads one card's events without writing and without making writers wait

## Intent

카드 한 장에서 무슨 일이 어떤 순서로 있었는지(누가 잡았고, 갱신하고, 인계하고, 거절당했고, 어떤 노트가 달렸는지)는 원장의 `event` 표에 전부 있어요. 그런데 지금 그것을 보는 길은 `card show` 뿐이고, `show` 는 카드의 현재 모습만 보여 주며 이력은 `reservation` 표를 따로 읽어야 해요. 협업 보드 설계(보고서 `collab-board-final-design-2026-11-20.md` 3판 5.1절, 6.1절 BOARD 01)가 첫 겹으로 정한 것은 읽기 전용 타임라인이에요: `moai card timeline C-n`.

이 명령에서 가장 조심할 것은 읽는 방식이에요. 같은 보고서 3.5절의 측정(임시 HOME, 노트 300→902개, 같은 날 20:48~20:49 UTC)에서 열려 있는 읽기 트랜잭션은 쓰기 커밋을 그 시간만큼 붙잡았고(8초를 쥐면 쓰기가 5.3초에 `ledger busy` 로 실패), 짧게 읽고 닫는 읽기는 쓰기에 영향이 없었어요(읽기 4,591번에 Busy 0건). 그래서 이 명령은 읽는 법이 한 가지예요: 짧은 대기로 읽기 연결을 열고, 읽기 트랜잭션 하나 안에서 정해진 개수만 읽고, 닫은 다음에 그려요. `event` 에는 `card_id` 색인이 없어서 카드 하나의 이벤트를 읽어도 표 전체를 한 번 훑어요 — 행마다 다시 훑는 질의(`card show` 의 노트 질의가 그래요, 별도 카드)는 쓰지 않고, 훑는 횟수를 질의 계획으로 시험해요. 스키마(표·색인·`journal_mode`)는 바꾸지 않아요(원장 v1 그대로).

## Non-goals

- 원장 스키마 변경과 `event(card_id)` 색인 추가(v2 결정이에요), 새 쓰기, 수집·관측 저장(BOARD 02), 생존 증명과 본인 재개(BOARD 03).
- 타임라인 상태의 우선순위 계산(보고서 5.6절)과 불변식 검사(5.7절): 이 조각은 이벤트 목록만 보여 줘요.
- `card show` 의 노트 질의 고치기: 별도 카드(그 질의의 읽기 방식에 이 명령을 맞춰요).
- 기다리는 조회(`--follow`): `--after` 는 기다리지 않아요.
- 카드 번호를 가지 이름·SPEC id·커밋에 쓰지 않아요(이 SPEC 의 id 도 그래요).

## Requirements

- REQ-1: When `card timeline C-n` is run for a card that exists, the command shall print that card's events oldest first, each with its time, kind, session, lane name and the scalar fields of its payload, in one `result` of the existing envelope.
- REQ-2: The command shall read the ledger over a connection that waits at most 250 milliseconds, inside one read transaction, with at most the requested number of events, and close both before it renders anything.
- REQ-3: The command shall read the card's events with one query and the lane names with one query, with no query run once per event row.
- REQ-4: The command shall write nothing, so that the ledger file and its folder are the same before and after and no folder or file is created.
- REQ-5: When `--after CURSOR` is given, the command shall return only the events after that cursor, without waiting for new ones.
- REQ-6: The result shall carry an opaque `next` cursor and an opaque `end` cursor, so that giving `next` back returns the following events with no gap and no overlap, and `end` names the newest event now.
- REQ-7: The command shall show every `fence_reject` event and leave out the `renew` bookkeeping events unless `--all` is given.
- REQ-8: The result shall carry a header with the project id, this session's id (null without one) and the time of the read.
- REQ-9: The result shall name the current owner of the card by lane name and session id, the lane name being display only.
- REQ-10 (REQ-1): IF the ledger is busy past the wait, or the card does not exist, or the project has no ledger, or `--after` or `--limit` is not valid, THEN the command shall fail in the existing envelope with the existing kind and exit code, writing nothing.
- REQ-11: The command shall document in its first help sentence and in the operations table that it reads the ledger only.
- REQ-12 (REQ-5): IF `--after` is not a cursor of this card, THEN the command shall refuse it with `card timeline refused` and return no event.

## Acceptance

- AC-1 (REQ-1, REQ-8, REQ-9): 한 카드가 만들기·준비·잡기·노트 둘·갱신·인계·인계 받기 뒤에 있으면 이벤트가 오래된 것부터 나오고, 각 줄에 시각·종류·세션·레인 이름·페이로드의 단순 값(긴 문자열은 200자로 잘림, 배열은 `<이름>_count`)이 있고, 노트 줄에는 작성자·첫 줄이 붙어요. 머리에 프로젝트 id·이 세션 id(세션 없으면 null)·읽은 시각이 있고, `owner` 에 현재 담당의 레인 이름(그 세션이 쓴 가장 새 노트의 작성자, 없으면 null)과 세션 id·세대·임대 끝이 있어요.
  verify: `cargo test --locked --offline --test card_timeline_cli the_timeline_lists_the_cards_events_oldest_first_with_the_contract_fields`
- AC-2 (REQ-4): 명령 전후로 원장 파일의 바이트와 폴더 목록·크기·수정 시각이 같고, 원장이 없는 프로젝트는 `ledger not found` 로 거절되며 아무것도 만들지 않아요.
  verify: `cargo test --locked --offline --test card_timeline_cli the_timeline_reads_and_writes_nothing`
- AC-3 (REQ-2, REQ-3): 노트 900개인 카드에서 이벤트 질의의 질의 계획에 상관 하위 질의(`CORRELATED`)가 없고 `event` 를 훑는 줄이 하나뿐이며, 레인 이름 질의도 `event` 를 한 번만 훑고, 명령이 쓰는 질의는 그 둘과 노트 조인뿐이에요.
  verify: `cargo test --locked --offline --test card_timeline_cli one_timeline_call_never_rescans_event_per_row`
- AC-4 (REQ-5, REQ-6, REQ-12): `--limit 3` 으로 이벤트 열 개를 이어 읽으면(`next` 를 `--after` 로 되돌려서) 열 개가 빠짐도 겹침도 없이 순서대로 나오고, 마지막 쪽 뒤에 새 이벤트가 생기면 `--after` 로 그것만 나오고, 읽는 동안 `end` 는 가장 새 이벤트를 가리켜요. 다른 카드의 커서나 모양이 틀린 커서는 거절돼요.
  verify: `cargo test --locked --offline --test card_timeline_cli the_cursor_continues_with_no_gap_and_no_overlap`
  verify: `cargo test --locked --offline --lib card::timeline`
- AC-5 (REQ-7): 갱신(`renew`) 이벤트는 기본 화면에 없고 `--all` 에 있으며, 거절 기록(`fence_reject`)은 둘 다에 있어요.
  verify: `cargo test --locked --offline --test card_timeline_cli bookkeeping_events_are_hidden_unless_all_and_a_fence_reject_always_shows`
- AC-6 (REQ-10): 다른 프로세스가 원장을 독점 잠금으로 쥐고 있으면 명령이 2초 안에 `kind: io`·`retryable: true`·종료 1 로 끝나고, 없는 카드는 `not_found`, 잘못된 `--limit`·`--after` 는 `invalid_input` 이며 기존 `card show` 와 같은 종료 코드를 써요.
  verify: `cargo test --locked --offline --test card_timeline_cli a_busy_ledger_a_missing_card_and_bad_input_use_the_existing_envelope`
- AC-7 (REQ-11): `card timeline --help` 의 첫 문장이 `Reads the ledger only.` 이고 운영 문서 표에 한 줄이 있어요.
  verify: `cargo test --locked --offline --test card_timeline_cli the_timeline_help_and_the_operations_table_say_it_reads_the_ledger_only`
- AC-8 (REQ-2, REQ-3): 증거 폴더의 측정 스크립트를 타임라인 반복으로 바꿔 세 번 이상 돌리면 쓰기 실패가 0건이고 쓰기 95번째 값이 대조군(다른 원장을 읽는 같은 수의 반복)의 2배를 넘지 않아요. · manual: `the run, the numbers and the script are written in the Result section of this SPEC; a timing measurement is not a deterministic test`
- AC-9 (REQ-1…REQ-12): 이웃 검사.
  verify: `cargo test --locked --offline --test card_cli`
  verify: `cargo test --locked --offline --test card_inbox_cli`
  verify: `cargo fmt --check`
  verify: `cargo clippy --offline --locked --all-targets -- -D warnings`

## Scope

- write: `src/card/timeline.rs`, `src/card/mod.rs`, `src/ops/card.rs`, `tests/card_timeline_cli.rs`, `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261120-card-timeline/**`
- read-only: `src/card/command.rs` (`show`, 다른 카드가 고치는 중), `src/card/inbox.rs`, `src/ledger/open.rs`, `src/ledger/schema.rs`, `src/ledger/reservation.rs`, `src/ledger/workflow.rs`
- forbidden: 실제 HOME 과 원장 쓰기, 스키마 변경, 네트워크, 키, 호스트 훅 설치. 시험은 임시 HOME 만 써요.

## Plan

1. 이 SPEC 을 쓰고 `spec check` 오류 0 을 확인해요.
2. 시험 먼저(RED): `src/card/timeline.rs` 의 단위 시험(커서 모양, 세부 값 줄이기)과 `tests/card_timeline_cli.rs` 의 AC-1~AC-7 을 임시 HOME 의 원장(`card add`·`card note` 와 SQL 삽입)으로 써서 명령이 없는 상태에서 실패시켜요.
3. `src/card/timeline.rs`: 짧은 대기(250 ms)로 읽기 연결 → 읽기 트랜잭션 하나(이벤트 질의 하나, 레인 이름 질의 하나, 현재 예약 질의 하나, 카드 한 줄) → 닫기 → JSON 만들기. 질의 문장은 상수로 두어 질의 계획 시험이 같은 문장을 써요.
4. `src/ops/card.rs` 에 `timeline C-n [--project] [--home] [--after] [--limit] [--all]` 동작을 연결하고 운영 문서 표에 한 줄을 더해요.
5. 변이(최대 5개): 커서 비교 `>` → `>=`, `renew` 숨기기 지우기, `fence_reject` 를 숨김 목록에 넣기, 읽기 트랜잭션 없이 질의, `--after` 가 가장 마지막 쪽을 건너뛰기.
6. 측정(AC-8): 증거 폴더의 `measure_ledger2.py` 를 복사해 읽는 쪽을 `card timeline` 으로 바꾸고 세 번 돌려 Result 에 적어요.
7. 검사(여섯 검사, `card*` 접두 시험 전부, 등재 검사)와 감사.

## Risks

- 레인 이름은 그 세션이 이 카드에 쓴 가장 새 노트의 `author` 예요. 노트를 한 번도 안 쓴 세션은 이름이 null 이에요(D-1). 이름은 표시일 뿐이고 판정(담당이 누구인지)은 세션 id 와 세대만 써요.
- `event` 에 카드 색인이 없어서 이벤트 질의는 표를 한 번 훑어요(오늘 약 3천 행, 하루 약 1천 행씩 늘어요). 이 명령은 그 시간을 측정 숫자로 남기고(Result), 색인은 v2 결정으로 남겨 둬요.
- 기본 화면에서 `renew` 만 숨겨요(D-2). 숨기는 이벤트가 늘어나면 한 목록(`BOOKKEEPING`)만 고치면 돼요; 숨긴 이벤트 앞에서 이어 읽으면 그 이벤트는 `--all` 로만 볼 수 있어요.
- 커서는 이벤트 번호에서 만들지만 계약은 불투명이에요: 모양을 바꿔도(D-3) 받은 그대로 돌려주는 호출자는 안 깨져요.

## Decisions

| id | 결정 | 근거 | 되돌리기 |
|---|---|---|---|
| D-1 | 레인 이름은 그 세션이 쓴 가장 새 노트의 작성자예요 | 원장에 레인 이름 칸이 없어요(스키마 v1 고정). 노트의 `author` 가 레인 이름이고 `card.note` 이벤트가 세션을 적어요(`src/ledger/workflow.rs:441`) | 레인 등록부에서 읽도록 바꿀 수 있어요(읽기 전용 파일 한 개 더) |
| D-2 | 기본 화면은 `renew` 만 숨겨요 | 설계 5.1절이 "renew 같은 장부 정리 이벤트"를 말하고, 코드에 있는 종류 가운데 `renew` 만 내용 없는 되풀이예요(`src/ledger/reservation.rs:773`) | 목록 한 줄 |
| D-3 | 커서는 `tl1:<카드 번호>:<이벤트 번호>` 예요 | 이벤트 번호가 로컬 순번이고 한 번 쓰이면 바뀌지 않아요(`event_no_update`, `event_no_delete` 트리거). 카드 번호를 넣어 다른 카드의 커서를 거절해요 | 모양을 바꿔도 계약이 불투명이에요 |
| D-4 | 원장이 없으면 `card show` 와 같은 `ledger not found` 로 거절해요 | 한 카드의 이력을 읽는 명령이라 카드가 없는 것과 같아요. `card inbox` 의 "빈 목록" 과 달라요 | 빈 결과로 바꿀 수 있어요 |
| D-5 | `--limit` 은 1~500, 기본 50 | 읽는 개수를 정해진 수로 묶는 것이 이 카드의 약속이에요. 노트 900개 카드도 쪽으로 읽어요 | 상수 하나 |
| D-6 | 결과에 `end`(지금 가장 새 이벤트의 커서)를 넣어요 | 리더가 라운드를 시작할 때 커서를 받아 적고 끝에 그 뒤를 읽는 법(설계 5.1절)에 필요해요 | 필드 하나 |

## Result

(구현 뒤에 적어요.)
