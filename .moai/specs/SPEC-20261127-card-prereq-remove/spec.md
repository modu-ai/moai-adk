---
id: SPEC-20261127-card-prereq-remove
title: "카드 선행 조건을 지울 수 있고, 어떤 길이의 순환도 거절하고, HOLD·선행 조건 검사는 예약 트랜잭션 안에서 해요"
status: approved
revision: 1
approval: {by: user, at: 2026-11-27, basis: "사용자가 직접 골랐어요(리더 세션의 질문 도구, 레저 노트 10699) 선택지 '기본값으로 승인 (Recommended)' — 개정 1을 글 그대로 승인하고 D-1 은 (a) 철회만 풀기, D-3 은 끝난 카드를 지나는 고리까지 모두 거절, D-4 는 앞의 검사를 그대로 두고 안쪽 검사를 더함, D-6 은 HOLD 후보 목록 문구를 코드에 맞춤. 승인 대상 커밋 41453a1b7d", digest: "sha256:d544bb164300ab23629902f6f257ef1d260cf94e73ca3b7210d2372c9fb37865"}
created: 2026-11-27
---

# 카드 선행 조건을 지울 수 있고, 어떤 길이의 순환도 거절하고, HOLD·선행 조건 검사는 예약 트랜잭션 안에서 해요

## Intent

`SPEC-20261121-card-hold-gate` 가 카드의 HOLD 와 선행 조건(`blocks` 링크)을 `card next` 와 `card take` 가 읽게 했어요. 그 SPEC 을 리뷰하면서 남은 틈이 넷이에요. (1) `card take` 는 HOLD·선행 조건을 예약 트랜잭션 **앞**의 따로 연 읽기로 검사해서, 그 사이에 쓰인 HOLD 를 지나가요. (2) 선행 조건을 지우는 명령이 없어서, 잘못 건 선행 조건이나 철회된 카드에 건 선행 조건이 카드를 영영 막아요. (3) 순환은 두 카드 짝만 거절하고 세 카드 이상의 고리는 그대로 받아요. (4) HOLD 후보 목록의 상태 범위가 SPEC 문구와 코드에서 서로 달라요.

이 SPEC 은 (1)~(3) 을 고치고 (4) 는 문구를 코드에 맞춰요. 게이트를 약하게 하는 변경이 아니에요: 검사를 트랜잭션 안으로 옮기는 것은 더 엄격하고, 선행 조건을 풀어 주는 길은 사람이 `--by` 와 함께 부르는 명령과 철회된 카드 한 가지뿐이에요. 게이트를 건드리므로 tier 2 로 보고 승인 전에는 코드를 쓰지 않아요.

## Interview

- Q1: 철회된 카드에 건 선행 조건은 어떻게 하나요? → A: (기본값, 승인에서 확정) 철회된(`card.withdraw`) 카드는 더 이상 선행 조건을 막지 않아요. 철회는 그 일을 하지 않기로 한 것이라 기다릴 이유가 없어서예요. 합쳐진(`card.fold`) 카드는 계속 막아요: 일이 다른 카드로 옮겨 갔을 뿐 끝난 것이 아니라서, 사람이 새 카드에 선행 조건을 걸고 옛 것을 지워요. 아무 이유 없이 `hold` 상태인 카드도 계속 막아요(D-1).
- Q2: 선행 조건을 지우면 링크 행은 사라지나요? → A: 아니요. 행은 지워지지 않아요(스키마 트리거가 막아요). 링크의 `state` 를 `retired` 로 바꾸고 `card.prereq_remove` 이벤트 하나를 남겨요. 다시 걸면 같은 행이 `active` 로 돌아와요(D-2).
- Q3: 순환 검사는 끝난 카드도 보나요? → A: (기본값) 네. 모든 `active` 링크를 카드 상태와 상관없이 따라가요. 끝난 카드를 지나는 고리까지 거절하면 조금 엄격하지만, 단순하고 예측할 수 있고 `--remove` 로 풀 수 있어요(D-3).
- Q4: 트랜잭션 안의 검사는 어느 예약 길에 두나요? → A: `card take` 가 쓰는 두 길(`reserve_sweeping`, `reserve_own_scope_sweeping`)에 두어요. 입양(`adopt`)과 시험이 쓰는 `reserve`·`reserve_own_scope` 는 그대로예요. 앞의 검사는 그대로 두어서 거절 문구와 순서(HOLD → 기준 없음 → 진행 한도)가 바뀌지 않고, 트랜잭션 안 검사가 그 사이의 틈을 닫아요(D-4).
- Q5: 이벤트 `event` 표의 인덱스는요? → A: 이미 있어요. `event_card_kind(card_id, kind, event_id)` 가 새 장부와 옛 장부(열 때 만들어요) 모두에 있고, `current_hold` 질의가 그 인덱스를 타는 것을 확인했어요(`SEARCH event USING INDEX event_card_kind (card_id=? AND kind=?)`). 카드가 말한 "표 전체를 읽는다" 는 사실이 아니라서 인덱스는 더하지 않고, 인덱스를 타는 것을 시험 하나로 못 박아요(D-5).

## Non-goals

- 스키마 변경, 새 상태값, 새 링크 종류. 행은 지우지 않고 `retired` 로 바꿔요.
- `card hold`·`card unhold`·`card withdraw`·`card fold` 의 동작, `card next` 의 건너뜀 이유 이름.
- 세션 id 나 권한을 보는 새 검사. `--remove` 도 `card prereq` 처럼 이름(`--by`)만 받아요.
- 합쳐진 카드의 선행 조건을 새 카드로 자동으로 옮기는 것.
- `adopt` 와 시험용 `reserve` 길에 게이트를 거는 것.

## Requirements

- REQ-1: When `card prereq C-n --needs C-m` would add a link that closes a loop of any length of active `blocks` links, the command shall refuse, name every card of the loop in order, and write nothing · failure: none
- REQ-2 (REQ-1): Where the loop is two cards, the refusal shall keep its present words, `C-m already needs C-n, so they would wait for each other`.
- REQ-3: When `card prereq C-n --needs C-m --remove --by <name>` runs and an active `blocks` link from C-m to C-n exists, the command shall set that link's state to `retired`, append one `card.prereq_remove` event that records `needs` and `by`, and answer `removed: true` · failure: none
- REQ-4 (REQ-3): IF there is no active link, THEN the command shall answer `removed: false` and write nothing.
- REQ-5 (REQ-3): IF a card is not found or the arguments are invalid, THEN the command shall refuse and write nothing.
- REQ-6: When a retired link is added again with `card prereq`, the command shall make the same row active after the loop check of REQ-1 · failure: none
- REQ-7: When the last unfinished prerequisite of a ready card is removed, `card next` shall offer the card in `ready` · failure: none
- REQ-8 (REQ-7): While a ready card has no unfinished prerequisite left, `card take` shall not refuse it for a prerequisite.
- REQ-9: While a prerequisite card is in the `hold` state and its newest `card.withdraw` or `card.fold` event is a `card.withdraw`, it shall not count as an unfinished prerequisite.
- REQ-10 (REQ-9): While a prerequisite card is folded, in the `hold` state with neither event, or in any state other than `done` and `hold`, it shall keep counting as an unfinished prerequisite.
- REQ-11: When `card take` reserves a card, the reservation transaction shall read the HOLD and the unfinished prerequisites inside the same write transaction and, IF either is found, refuse with the words of the early check, `card take refused: …; nothing was taken`, and write no reservation, state change or event · failure: none
- REQ-12: When the HOLD of a card is read, the query shall search the `event` table through the `event_card_kind` index · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): a temporary ledger with `a needs b`, `b needs c` refuses `c needs a` and names the three cards in order; a four-card loop is refused the same way; a card that is not in the loop is not named; a link that was retired does not close a loop; the two-card refusal keeps its words; the link rows and events are equal before and after each refusal · verify: `cargo test --locked --offline --test card card_hold_gate_cli::prereq_ac1_`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): `card prereq --remove` retires the link (the row stays, state `retired`), writes one `card.prereq_remove` event with `needs` and `by`, and answers `removed: true`; a second call answers `removed: false` and writes nothing; an unknown card and a missing `--by` are refused with nothing written; adding the same pair again makes the same row active and the loop check still runs · verify: `cargo test --locked --offline --test card card_hold_gate_cli::prereq_ac2_`
- AC-3 (REQ-7, REQ-8, REQ-9, REQ-10): after `--remove`, `card next` lists the card in `ready`, `card show` lists no prerequisite and `card take` succeeds; a withdrawn prerequisite does not block, and a folded one, a plain `hold` card with no withdraw or fold event, an `active` one and a `review` one still block; a `done` one does not · verify: `cargo test --locked --offline --test card card_hold_gate_cli::prereq_ac3_`
- AC-4 (REQ-11): called directly (so the early check does not run), `reserve_sweeping` and `reserve_own_scope_sweeping` refuse a card with a `card.hold` in force and a card with an unfinished prerequisite with the gate error, and leave the whole ledger equal; `reserve` and `reserve_own_scope` keep their behaviour; `card take` through the binary refuses with the words it has now · verify: `cargo test --locked --offline --test card ledger_reservation::prereq_ac4_`; `cargo test --locked --offline --test card card_hold_gate_cli::take_refuses`
- AC-5 (REQ-12): `EXPLAIN QUERY PLAN` of the HOLD read names `event_card_kind` · verify: `cargo test --locked --offline --test card card_hold_gate_cli::prereq_ac5_`
- AC-6 (REQ-3): the operations document names `--remove`, the retired state and the any-length loop refusal, and the old SPEC says in its Result that REQ-5 now refuses loops of any length and that `card hold-candidates` lists open cards in `draft`, `ready`, `active` and `review` · verify: `git diff --check`; `moai spec check SPEC-20261127-card-prereq-remove`
- AC-7: the neighbours still pass · verify: `cargo test --locked --offline --test card card_hold_gate_cli::`; `cargo test --locked --offline --test card ledger_reservation::`; `cargo test --locked --offline --test card card_next_skip_reason_cli::`; `cargo test --locked --offline --test card card_dispatch_cli::`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `moai trailer-check NeXT..HEAD`

## Scope

- write: `src/ledger/workflow.rs`, `src/ledger/reservation.rs`, `src/card/dispatch.rs`, `src/card/work.rs`, `src/ops/card.rs`(공유 파일, 줄 리스로), `docs/design-review/appendix/operations/card.md`(공유 파일, 줄 리스로), `tests/card/card_hold_gate_cli.rs`, `tests/card/ledger_reservation.rs`, `.moai/specs/SPEC-20261127-card-prereq-remove/**`, `.moai/specs/SPEC-20261121-card-hold-gate/spec.md`(Result 에 두 줄만)
- read-only: `src/ledger/schema.rs`, `src/card/hold_candidates.rs`, `src/card/command.rs`, `tests/common/home.rs`
- forbidden: `src/ledger/schema.rs`(스키마는 바꾸지 않아요), 진짜 HOME 의 장부(시험은 임시 HOME 만 써요)

## Plan

1. 승인을 기다려요. 승인 뒤 쓰기 범위를 SPEC Scope 로 넓히고 겹치는 점유를 읽어요(`src/ops/card.rs`·`card.md` 는 줄 리스, 쓴 뒤 바로 반환).
2. RED: `tests/card/card_hold_gate_cli.rs` 에 `prereq_ac1_`~`prereq_ac3_`, `prereq_ac5_`, `tests/card/ledger_reservation.rs` 에 `prereq_ac4_` 를 먼저 써서 실패를 봐요(임시 HOME·임시 장부).
3. `workflow.rs`: `add_prerequisite` 에 순환 검사(`from_card → to_card` 의 `active` 링크를 너비 우선으로 따라가 `needs` 에 닿으면 거절, 닿은 길을 담은 `PrereqError::Cycle { card_id, needs, path }`, 두 카드 순환은 지금 문구 그대로)와 `retire_prerequisite`(링크를 `retired` 로, 이벤트 `card.prereq_remove`), `unfinished_prerequisites` 의 철회 규칙(`NOT (c.state = 'hold' AND newest(card.withdraw, card.fold) = card.withdraw)`), `hold_gate_refusal` 을 `dispatch.rs` 에서 옮겨 와 두 곳이 같은 함수를 불러요.
4. `reservation.rs`: `ReserveError::HoldGate(String)`(`code` 는 `hold_gate`, 글은 그대로 보여 줘요)와 `reserve_in_tx` 의 `gate` 인자. 두 sweeping 길만 `true` 를 넘겨요. 검사는 `state != "ready"` 거절 뒤, 쓰기 앞이에요.
5. `work.rs`·`ops/card.rs`: `card prereq` 에 `--remove`, 답에 `removed`. 운영 문서와 옛 SPEC 의 Result 두 줄.
6. 검증: AC 의 명령. 가드 뮤턴트(새 가드 테스트마다 한 줄: 순환 검사 끔, 철회 규칙 끔, 접힌 카드까지 풀어 줌, 트랜잭션 안 검사 끔, 링크를 지우지 않고 `active` 로 둠). 위험한 길: 행이 지워지지 않는 장부(트리거)와 임시 HOME 에서 진짜 `moai` 로 take·remove 를 돌려요. `check affected --card` 뒤 READY.

읽은 사실(2026-11-27, NeXT 끝점 7ec80a8155): `add_prerequisite`(`src/ledger/workflow.rs`)는 같은 쌍의 거꾸로 링크 하나만 봐서 두 카드 순환만 거절해요. `card_link` 는 `kind IN ('blocks',…)` 와 `state IN ('active','retired')` 를 갖고 `card_link_no_delete` 트리거가 행 삭제를 막아요. `event` 에는 `kind` 검사가 없어서 새 이벤트 이름은 스키마를 안 바꿔요. `hold_gate_refusal` 은 `src/card/dispatch.rs` 에서 `card take` 가 `is_ready` 일 때 예약 앞에서 불러요; `reserve_in_tx`(`src/ledger/reservation.rs`)는 그 검사를 하지 않아요. `ReserveError` 는 시험 둘(`tests/card/ledger_reservation.rs`, `ledger_recover.rs`)이 이름으로 맞춰 보는데 `other =>` 갈래와 `matches!` 를 써서 변형을 더해도 깨지지 않아요(`ledger_reservation.rs` 의 `match` 를 읽었어요).

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. 모두 제안이고 승인에서 고르는 값이에요.

- D-1 (되돌릴 수 있음, 범주 none, 승인 때, 리더): 철회된 카드 한 가지만 막기를 풀어요. 선택지: (a) 철회만 풀기, (b) 철회와 합침 모두 풀기, (c) 아무것도 안 풀고 `--remove` 만. 기본값: (a). 이유: 철회는 일이 없어진 것이고 합침은 일이 옮겨 간 것이라, 합침을 풀면 아직 안 끝난 일을 기다리지 않게 돼요.
- D-2 (되돌릴 수 있음, 범주 none): 지우기는 행 삭제가 아니라 `retired` 로 바꾸기예요(트리거가 삭제를 막아요). 선택지: 새 링크 종류. 기본값: 같은 행의 상태.
- D-3 (되돌릴 수 있음, 범주 none): 순환 검사는 카드 상태와 상관없이 모든 `active` 링크를 따라가요. 선택지: 끝난 카드를 지나는 고리는 허용. 기본값: 모두 거절.
- D-4 (되돌릴 수 있음, 범주 none): 트랜잭션 안 검사는 `card take` 가 쓰는 두 길에만, 앞의 검사는 그대로 둬요. 선택지: 앞의 검사를 없애고 트랜잭션 안에만. 기본값: 둘 다(문구와 순서가 안 바뀌어요).
- D-5 (되돌릴 수 있음, 범주 none): `event` 인덱스는 이미 있어서 더하지 않고 시험으로 못 박아요. 선택지: 인덱스를 새로 정의. 기본값: 시험만.
- D-6 (되돌릴 수 있음, 범주 none): HOLD 후보 목록의 범위는 코드(`draft`·`ready`·`active`·`review`)에 문구를 맞춰요. 선택지: 코드를 "끝나지 않은 모든 카드" 로. 기본값: 문구를 코드에 맞춤(`hold` 상태 카드는 이미 철회된 카드라 후보가 아니에요).

## Risks

- **순환 거절이 엄격해져요.** 지금 받아 주던 세 카드 이상의 고리를 거절해요. 이미 그런 고리가 있는 장부에서는 `--remove` 로 풀어야 하고, 그 장부의 카드는 지금도 영영 막혀 있어요.
- **철회 규칙은 카드의 `card.withdraw`/`card.fold` 이벤트를 읽어요.** `hold` 상태로 만들어진 카드(가져오기)는 이벤트가 없어 계속 막아요. 철회한 카드를 다시 `ready` 로 되살리면 `hold` 상태가 아니라서 다시 막아요.
- **`ReserveError` 변형이 하나 늘어요.** 지금 이름으로 맞추는 시험은 `other =>` 갈래나 `matches!` 라서 깨지지 않지만, 빠짐없는 `match` 가 어딘가에 있으면 컴파일이 깨져요. 구현 때 `git grep` 으로 다시 확인해요.
- **앞의 검사와 안의 검사가 같은 함수를 불러야 해요.** 둘의 글이 어긋나지 않게 함수 하나로 모아요(`no_criteria` 가 선례).
- **두 SPEC 의 관계:** `SPEC-20261121-card-hold-gate` 의 REQ-5(두 카드 순환)는 이 SPEC 의 REQ-1 이 넓혀요. 옛 SPEC 의 Result 에 한 줄 적어요.

## Result

구현(2026-11-27, 브랜치 `spec/card-prereq-remove-20261127`): RED 커밋 `61976c908c`(`moai tdd red`), 구현 커밋 `99acba7ed6`(`moai tdd green` 통과), NeXT 끝점 `8b6ab17634` 머지 뒤 `check affected --card` 가 `result=PASS`(head `b3ff33c97a4e`, 테스트 3076개 통과, `clippy=0`).

바뀐 곳: `src/ledger/workflow.rs`(`prerequisite_loop` 너비 우선 순환 검사와 `PrereqError::Cycle { path }`, `retire_prerequisite`, 철회 규칙이 든 `unfinished_prerequisites`, `dispatch.rs` 에서 옮겨 온 `hold_gate_refusal`, 상수 `HOLD_READ_SQL`), `src/ledger/reservation.rs`(`ReserveError::HoldGate`, `reserve_in_tx` 의 `gate` 인자, 두 sweeping 길만 `true`), `src/card/dispatch.rs`, `src/card/work.rs`·`src/ops/card.rs`(`card prereq --remove`), 운영 문서 `card.md` 와 옛 SPEC 의 Result 두 줄. 승인된 결정(D-1 (a), D-3, D-4, D-6)은 글 그대로예요.

- AC-1: met — `prereq_ac1_a_loop_of_any_length_is_refused_naming_every_card_and_writing_nothing`, `prereq_ac1_a_retired_link_does_not_close_a_loop` 통과(세 카드·네 카드 고리, 고리 밖 카드는 이름에 없음, 두 카드 문구 그대로, 링크 행과 이벤트 수 불변).
- AC-2: met — `prereq_ac2_remove_retires_the_link_and_writes_one_event`, `prereq_ac2_a_retired_link_is_added_again_as_the_same_row_after_the_loop_check` 통과.
- AC-3: met — `prereq_ac3_removing_the_last_prerequisite_offers_the_card_and_take_goes_through`, `prereq_ac3_only_a_withdrawn_prerequisite_stops_blocking` 통과(철회와 `done` 은 풀리고 합쳐진·사건 없는 `hold`·`active`·`review` 는 계속 막아요).
- AC-4: met — `ledger_reservation::prereq_ac4_` 세 개 통과, `card_hold_gate_cli::take_refuses_…` 가 이웃 모듈 실행(81개 통과)에 들어 있어요.
- AC-5: met — `prereq_ac5_the_hold_read_goes_through_the_event_card_kind_index` 통과(계획이 `event_card_kind` 를 이름으로 대고 `SCAN event` 가 없어요). 인덱스는 이미 있어서 스키마는 안 바꿨어요.
- AC-6: met — 운영 문서 `card.md` 의 `card prereq` 행에 `--remove`, `retired`, 길이에 상관없는 고리 거절, 철회 규칙을 적었고 옛 SPEC 의 Result 에 두 줄을 더했어요. `git diff --check` 와 `moai spec check` 는 오류 0 이에요.
- AC-7: met — 이웃 모듈 `card_hold_gate_cli::`·`ledger_reservation::`·`card_next_skip_reason_cli::`·`card_dispatch_cli::` 81개 통과(구현 커밋에서), 합친 head 의 `check affected --card` 가 3076개 통과와 `clippy=0`, `moai trailer-check NeXT..HEAD` 지적 0.

가드 뮤턴트(다섯, 모두 복원): 순환 검사의 도달 조건을 거짓으로(`prereq_ac1_`), 철회 조건을 없는 사건 이름과 비교(`prereq_ac3_only_a_withdrawn_…`), 합쳐진 카드까지 풀어 줌(같은 테스트), 트랜잭션 안 검사를 끔(`prereq_ac4_the_sweeping_…refuse…`), 지우기가 `active` 를 씀(`prereq_ac2_remove_…`) — 모두 테스트가 붉어졌어요(exit 101).

틈: 이미 세 카드 이상의 고리가 든 장부는 이번 변경이 풀지 않아요(`card prereq --remove` 로 풀어요). `adopt` 와 시험용 `reserve`·`reserve_own_scope` 에는 안쪽 검사를 넣지 않았어요(D-4). Windows·Linux 는 돌리지 않았어요.
