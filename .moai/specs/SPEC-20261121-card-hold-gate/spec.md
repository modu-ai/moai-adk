---
id: SPEC-20261121-card-hold-gate
title: "A structured HOLD and prerequisite on a card are read by card next and card take before any reservation or worktree, and nothing lifts a HOLD by itself"
status: approved
revision: 2
approval: {by: leader, at: 2026-11-24, basis: "leader decision note 8280 of card C-261 (answers lane note 8279, 2026-11-24T06:22:35Z, under the user's delegation of 2026-11-23T13:57Z, not a direct user pick), approving revision 2 at head f3cdb5537f", digest: "sha256:2ba1296e4d1e48a1f1547676f2bb2e1a99ceae0cc77f6969907ac0b0b31116c4"}
created: 2026-11-21
---

# 카드의 구조화된 HOLD·선행 조건을 next 와 take 가 워크트리·예약 전에 읽는다

## Interview

- Q1: 무엇이 문제였나요? → A: 사용자가 확인한 현장 점검(2026-11-21)에서 Codex CLI 화면 13개(리더 1 + 레인 12)를 다른 프로젝트에서 보니, 레인 7개가 준비된 후보를 반복해서 조회했지만 HOLD·선행 조건·범위 때문에 여럿을 시작하지 못했어요. 레인 둘은 HOLD 를 읽기 전에 워크트리를 만들었고, 한 레인은 카드를 잡은 뒤 HOLD 를 읽고 돌려줬어요. 그 프로젝트와 살아 있는 레인 설정·보호 설정은 건드리지 않아요.
- Q2: 지금은 어떻게 되어 있나요? → A: 원장에 `hold` 상태는 있지만 `card withdraw` 로만 가고(그 카드는 `card.withdraw` 이벤트가 이유를 남기고 `card take` 가 이유를 말하며 거절해요), HOLD 의 종류(돈·법무·외부 계약·비밀·실제 데이터)를 구조로 남기는 곳이 없어요. `card next` 는 `ready` 가 아닌 카드를 이유 없이 건너뛰고, `blocks` 링크(`card_link`)를 보지 않으며, 그 링크를 만드는 명령도 없어요. 그래서 HOLD 는 자유 글(카드 본문·노트)로만 적혀서 레인이 읽기 전에 워크트리부터 만들 수 있어요.
- Q3: 무엇을 만들까요? → A: (1) HOLD 를 구조로 기록하고 푸는 명령 `card hold`·`card unhold`(이벤트 `card.hold`·`card.unhold`, 새 상태값 없음), (2) 선행 조건을 구조로 기록하는 명령 `card prereq`(기존 `blocks` 링크), (3) `card next` 가 HOLD·선행 조건이 있는 카드를 이유와 함께 건너뛰고, (4) `card take` 가 예약을 쓰기 전에 거절하고, (5) `card show` 가 둘을 보여 주고, (6) 레인 시작 안내가 워크트리 전에 그걸 읽도록 하는 것까지예요.
- Q4: 자유 글 HOLD 는요? → A: 읽지도 풀지도 추정하지도 않아요. 노트에 "HOLD" 라는 말이 있어도 코드는 구조화된 이벤트만 읽고, HOLD 를 푸는 것은 사람이 `card unhold` 를 부르는 일뿐이에요(리더 HOLD 해제 규칙 그대로).
- Q5: 이미 자유 글로 적힌 HOLD 는요? (리더 결정 노트 5170, 2026-11-21) → A: 막힌 카드가 `ready` 로 보이면 안 돼요. 그래서 코드가 자유 글을 대신 풀거나 구조로 바꾸지는 않지만, 기존 카드의 본문·노트에서 HOLD 말이 든 것을 읽기 전용 후보 목록으로 보여 주고, 사람이 하나 골라 `card hold` 를 부르면 그때 구조화된 HOLD 가 돼요(REQ-11).
- Q6: 이번에 안 하는 것은요? → A: 같은 전체 조회를 반복하지 않게 하는 상태 지문(fingerprint)과 상태가 풀리면 다시 시작하는 기능은 다음 조각이에요. 이 SPEC 은 거절과 안내까지만 해요.

## Intent

HOLD 와 선행 조건이 자유 글에만 있으면 레인은 카드를 잡고 워크트리를 만든 뒤에야 알아요. 이 SPEC 은 그 둘을 원장의 구조화된 사실로 만들어 `card next`(건너뜀 이유)와 `card take`(예약 전 거절)가 읽게 하고, 워크트리를 만드는 단계의 안내가 먼저 읽도록 해서 빈 워크트리와 잡았다 돌려주기를 없애요. 풀기는 항상 사람이 해요.

## Non-goals

- 상태값 추가(스키마의 `state` CHECK 는 그대로), 스키마 변경. HOLD 는 이벤트, 선행 조건은 기존 `blocks` 링크예요.
- 자유 글 HOLD 를 자동으로 풀거나 추정하는 것, 리더·사용자 HOLD 결정을 코드가 대신하는 것.
- 같은 전체 조회를 줄이는 상태 지문과 재개 알림(Q6, 다음 조각), `card withdraw` 의 동작 변경.
- 그 현장 프로젝트와 살아 있는 레인의 설정을 바꾸는 것.

## Requirements

- REQ-1: When `card hold C-n --category <category> --reason <text> --by <name>` runs on a card in the draft or ready state that nobody holds, the command shall append one `card.hold` event that records the category, the reason and the by · failure: none
- REQ-2 (REQ-1): IF the category is not one of money, legal, external_contract, secret, real_data or other, THEN `card hold` shall refuse and change nothing.
- REQ-3: When `card unhold C-n --by <name> --reason <text>` runs on a card whose newest hold event is a `card.hold`, the command shall append one `card.unhold` event · failure: none
- REQ-4: Where no `card unhold` command is run, a `card.hold` event shall stay in force whatever the notes of the card say.
- REQ-5: When `card prereq C-n --needs C-m` runs, the command shall add one active `blocks` link from C-m to C-n and refuse a card that needs itself or a card that already needs C-n · failure: none
- REQ-6: When `card next` lists a ready card whose newest hold event is a `card.hold`, the card shall appear in `skipped` with the reason `hold`, the category and the reason, and not in `ready` and not as the next card · failure: none
- REQ-7: When `card next` lists a ready card with an active `blocks` link from a card that is not done, the card shall appear in `skipped` with the reason `prerequisite` that names those cards · failure: none
- REQ-8: IF the card has a hold or an unfinished prerequisite, THEN `card take` shall refuse before it writes any reservation, name the category and reason or the unfinished cards, and say that nothing was taken.
- REQ-9: When `card show` prints a card, the output shall carry its current hold (category, reason, by, time, or null) and its unfinished prerequisites · failure: none
- REQ-10: Where the lane start guidance tells a lane to create a worktree for a card, it shall first tell the lane to read the hold and the prerequisites of the card from `card next` or `card show` and to create no worktree for a card that is skipped for them · failure: none

- REQ-11: When `card hold-candidates` runs, the command shall list each card that is not done, has the word HOLD in free text in its body or notes and has no `card.hold` in force, with the card, the note id and a short quote of the line, and write nothing · failure: none

## Acceptance

- AC-1 (REQ-1…REQ-4): temp-repo and ledger tests: a ready card gets a hold with each of the six categories; a bad category and a held (reserved) or done card are refused with nothing written; `card unhold` appends the unhold event and the card is offered again; only the `card unhold` command ever appends an unhold event (no other command writes one); a note that says HOLD in free text changes nothing in `card next` or `card take`; the event rows hold the category, reason and by.
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-2 (REQ-5, REQ-7): tests: `card prereq` adds the link (a second call changes nothing), refuses a self need and a two-card cycle; a card whose prerequisite is not done is skipped with `prerequisite` and the card names; when the prerequisite card is done the card is in `ready`.
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-3 (REQ-6, REQ-8): tests: `card next` skips a held card with the reason `hold`, the category and the reason, not in `ready` and not as `next`; `card take` of a held card or of a card with an unfinished prerequisite exits 1 naming them, says nothing was taken, and leaves no reservation row, no state change and no event (ledger snapshot equal).
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-4 (REQ-9): test: `card show` carries `hold` (null without one, the fields with one) and `prerequisites`.
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-5 (REQ-10): the named guidance sentences are in the shipped agent-mode text and the skill source manifest matches it; the card commands have operations documents, help rows, examples and goldens.
  verify: `cargo test --offline --locked --test card_hold_gate_cli guidance`; `cargo test --offline --locked --bin moai-cli help_table_names_every_command_once`; `cargo test --offline --locked --test help_layout`
- AC-5b (REQ-11): tests: a ledger with one card that has HOLD in its body, one with HOLD in a note and one without lists exactly the first two with their note ids and quotes; the ledger snapshot is equal before and after; `card next` and `card take` treat the two as before until a person calls `card hold` on one of them, which then moves out of the list and out of `ready`.
  verify: `cargo test --offline --locked --test card_hold_gate_cli candidates`
- AC-6 (REQ-1…REQ-11): at most six mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result.

## Scope

- write: `.moai/specs/SPEC-20261121-card-hold-gate/**`, `src/ledger/workflow.rs` (`next` and the hold, unhold and prerequisite functions), `src/ledger/reservation.rs` (the refusal before the reserve write), `src/card/work.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/input.rs`, `src/card/hold_candidates.rs` (new), `src/card/mod.rs` (one `pub mod` line, by line lease), `src/ops/card.rs`, `src/help.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`, `docs/design-review/appendix/operations/card.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/references/source-manifest.json`, `tests/card_hold_gate_cli.rs`
- read-only: `src/ledger/schema.rs`, `src/card/spec_link.rs`, `tests/common/home.rs`
- forbidden: schema changes, the real HOME and ledger, the project the field check came from
- 읽은 사실(2026-11-24, tip `c99df246a1`; 판 1 은 2026-11-21 tip `64faee764c`): `src/ledger/workflow.rs` 의 `next_filled` 는 `state <> 'done'` 카드 중 `ready` 만 후보로 올리고 건너뛴 이유는 held, scope_overlap, empty_scope, review_no_holder, no_criteria 다섯이에요(`SkipReason`, `workflow.rs:216`). `no_criteria` 는 `counted_criteria` 하나를 `next` 와 `take` 가 같이 불러 둘이 어긋나지 않게 한 선례(SPEC-20261124-card-next-skips-no-criteria)라서, 이 SPEC 의 HOLD·선행 조건 검사도 함수 하나를 둘이 같이 불러요. `card_link` 의 `blocks` 는 스키마에 있고(`src/ledger/schema.rs:77`, 같은 쌍은 한 줄뿐인 유일 인덱스) 읽거나 쓰는 코드와 명령이 아직 없어요(`git grep` 2026-11-24: `src/card/export_format.rs` 의 가져오기 말고 없음). `withdraw`(`src/ledger/workflow.rs:805`)는 `hold` 상태와 `card.withdraw` 이벤트 선례예요. `take` 는 `src/card/dispatch.rs:437` 에서 `reserve`(`src/ledger/reservation.rs:683`)를 불러요. 쓰기 파일 대부분(`work.rs`, `dispatch.rs`, `help.rs`, `mode-agent.md`)은 다른 카드가 예약 중일 수 있어, 필요한 파일을 쓰기 전에 보유 카드가 반환하는 순서를 리더에게 물어요.

## Plan

1. **S1 원장**: RED — `tests/card_hold_gate_cli.rs` 를 임시 HOME 과 원장으로 먼저 써요(AC-1~AC-4). GREEN — `src/ledger/workflow.rs` 에 hold·unhold·prerequisite 함수와 `next` 의 건너뜀 이유, `card take` 의 예약 전 거절.
2. **S2 명령·표시·도움말**: `card hold`·`card unhold`·`card prereq` ·`card hold-candidates` 를 `src/ops/card.rs`·`src/card/*` 에 잇고 `card show` 를 넓혀요. 운영 문서, 도움말 행·예시·골든.
3. **S3 안내**: `mode-agent.md` 의 시작 단계에 "워크트리 전에 HOLD·선행 조건을 읽는다" 문장을 넣고 매니페스트를 다시 생성해요.
4. 변형 6개 이하, 이웃 검사를 돌려 Result 에 적어요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. 모두 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요.

- D1 (제안, 되돌릴 수 있음, 범주 none): HOLD 는 새 상태값이 아니라 `card.hold`·`card.unhold` 이벤트예요. 대안: `hold` 상태 재사용. 기본값: 이벤트. 이유: `hold` 상태는 이미 철회(withdraw)를 뜻하고 스키마 CHECK 는 값을 고정해요.
- D2 (제안, 되돌릴 수 있음, 범주 none): HOLD 종류는 money, legal, external_contract, secret, real_data, other 여섯이에요(MoAI 의 HOLD 다섯 범주 + 그 밖). 대안: 자유 단어. 기본값: 여섯 고정.
- D3 (제안, 되돌릴 수 있음, 범주 none): 선행 조건은 기존 `blocks` 링크(`from_card` 가 `to_card` 를 막음)를 쓰고 새 링크 종류를 만들지 않아요. 대안: 새 종류. 기본값: 기존.
- D4 (제안, 되돌릴 수 있음, 범주 none): 사람이 HOLD 를 거는 권한은 제한하지 않아요(카드를 만들고 고치는 쪽과 같음). 풀기도 같지만 사람이 부를 때만 일어나요. 대안: 리더만. 기본값: 제한 없음, 이벤트가 by 를 남겨요.
- D5 (제안, 되돌릴 수 있음, 범주 none): `card hold-candidates` 는 대문자 낱말 `HOLD` 가 든 카드 본문·노트를 후보로 보여 주기만 해요(읽기 전용, 자동 이동 없음). 대안: 대소문자 무시, 또는 사용자가 고른 카드를 한 번에 옮기는 `--apply`. 기본값: 대문자 낱말 일치, 옮기기는 사람이 `card hold` 를 부르는 일뿐.

## Risks

- 쓰기 파일 대부분을 다른 카드가 예약 중일 수 있어 순서가 어긋나면 S2·S3 가 밀려요. S1(`workflow.rs`, 새 시험 파일)만으로도 `next`·`take` 가 읽기 시작해요.
- 기존 시험이 `card next` 의 건너뜀 목록이나 `card show` 의 키를 정확히 맞춘다면 새 키(`hold`·`prerequisites`) 때문에 빨개질 수 있어요. 이웃 검사에서 드러나요.
- `blocks` 링크가 이미 다른 도구로 만들어진 원장이 있으면(가져오기) 그 의미(누가 누구를 막는지)가 이 SPEC 과 달라질 수 있어요. 방향은 스키마 주석과 같게 `from_card` 가 `to_card` 를 막는 것으로 읽고 시험으로 고정해요.
- 이 SPEC 은 같은 전체 조회를 줄이는 기능을 하지 않아 현장의 반복 조회는 남아요(Q6, 다음 조각).

## Result

판 2(승인 노트 8280)를 구현했어요. 가지 `spec/card-hold-gate-20261121`, 이 가지를 자른 NeXT 팁 `c99df246a1`(그 뒤로 팁이 움직였지만 쫓지 않았어요). 2026-11-24, 다른 레인의 부하가 큰 Mac(load 60~100)에서 `moai heavy-slot-acquire --name card-hold-gate --max 3` 안에서 돌렸어요. 같은 head 의 `moai check affected` 영수증은 이 글이 아니라 READY 노트에 있어요.

- 구현: `src/ledger/workflow.rs`(`current_hold`·`unfinished_prerequisites`·`hold`·`unhold`·`add_prerequisite`, `SkipReason::Hold`·`Prerequisite`, `next_filled` 가 둘을 `empty_scope` 보다 먼저 봐요), `src/card/dispatch.rs`(`hold_gate_refusal`: `take` 가 예약 앞에서 거절), `src/card/work.rs`(`hold`·`unhold`·`prereq` 와 `next` 의 `category`·`hold_reason`·`needs`), `src/card/command.rs`(`card show` 의 `hold`·`prerequisites`), `src/card/hold_candidates.rs`(`card hold-candidates`), `src/ops/card.rs`(네 명령), 운영 문서 `card.md`(네 행과 `next` 행), `skills/moai/workflows/work.md` 와 매니페스트.
- 판 1 과 다른 점(숨기지 않아요): (1) 레인 시작 안내는 `mode-agent.md` 가 없어져 `skills/moai/workflows/work.md` 의 take 단계에 넣었어요. (2) 카드 하위 명령은 도움말 표(`src/help.rs`)에 없어요(`fold`·`withdraw` 도 없어요) — `help_table_names_every_command_once` 는 최상위 명령만 보므로 도움말 행·골든은 바꿀 것이 없었고 `src/help.rs`·`tests/fixtures/help` 는 건드리지 않았어요(C-242 와 겹치지 않아요). (3) `take` 의 거절은 `reservation.rs` 가 아니라 `dispatch.rs` 에 있어요: `ReserveError` 에 변형을 더하면 `tests/ledger_reservation.rs` 의 빠짐없는 `match` 가 깨지는데 그 파일은 범위 밖이에요. 그래서 검사와 예약 사이(두 트랜잭션)에 누가 HOLD 를 쓰면 그 한 번은 지나가요 — `no_criteria` 검사와 같은 틈이에요. (4) 선행 조건의 순환은 두 카드 짝만 막아요(SPEC REQ-5 그대로; 세 카드 이상의 고리는 잡지 않아요).
- 새 시험 `tests/card_hold_gate_cli.rs` 14개와 `src/card/hold_candidates.rs` 단위 시험 2개가 통과했어요(`cargo nextest run --locked --offline --test card_hold_gate_cli --bin moai-cli ...` 슬롯 안, exit 0, 75 passed 에 이웃 `source_manifest_facts`·`card_cli`·`card_next_skip_reason_cli` 포함). 이웃 148개(`card_next_skip_reason_cli`·`card_needs_criteria_cli`·`card_show_notes_cli`·`card_cli`·`card_dispatch_cli`·`card_fold_cli`·`card_withdraw_cli`·`ledger_workflow` + 단위) 통과, `help_layout` 11개 통과.
- AC-6 가드 뮤턴트(S1 코드 head `6ca0858053` 에서, 하나씩 적용 확인 → 시험 → `git checkout --` 로 복원, 복원 뒤 `git diff --quiet`): 모두 `cargo nextest run --test card_hold_gate_cli` exit 100.

  | 뮤턴트 | 지운 것 | 빨개진 시험 |
  |---|---|---|
  | A | `next` 가 HOLD 를 안 봄(`current_hold` 자리에 `None`) | `next_skips_a_held_card_with_its_category_and_reason` 외 둘(6개 실패) |
  | B | `next` 가 선행 조건을 안 봄 | `next_skips_a_card_whose_prerequisite_is_not_done_and_offers_it_when_it_is`(2개) |
  | C | 끝난(`done`) 선행 카드도 계속 막음(`c.state <> 'nothing'`) | 같은 시험, `show_carries_the_hold_and_the_unfinished_prerequisites`, `take_refuses_a_held_card_and_a_card_with_an_open_prerequisite_and_writes_nothing`(6개) |
  | D | `hold_gate_refusal` 이 항상 `None` | `take_refuses_…_and_writes_nothing`, `nothing_but_unhold_lifts_a_hold_and_free_text_changes_nothing`(4개) |
  | E | 두 카드 순환 거절을 끔 | `a_prerequisite_is_added_once_and_refuses_a_self_need_and_a_two_card_cycle`(2개) |
  | F | `card.unhold` 를 무시함(`kind != "card.hold"` 를 끔) | `take_refuses_…`, `unhold_appends_its_event_and_the_card_is_offered_again`(4개) |

- AC-1: met — `a_ready_card_takes_a_hold_of_each_category_and_the_event_keeps_the_fields`, `a_hold_is_refused_for_a_bad_category_a_taken_card_and_a_card_already_on_hold`, `unhold_appends_its_event_and_the_card_is_offered_again`, `nothing_but_unhold_lifts_a_hold_and_free_text_changes_nothing`, `card_hold_takes_each_category_and_refuses_a_bad_one_writing_nothing`, `card_unhold_lifts_only_a_hold_in_force_and_the_card_is_offered_again` 통과. 한 가지 더 정했어요: 이미 HOLD 인 카드에 `card hold` 를 또 부르면 거절해요(사건을 늘리지 않아요).
- AC-2: met — `a_prerequisite_is_added_once_and_refuses_a_self_need_and_a_two_card_cycle`, `card_prereq_adds_one_link_once_and_refuses_a_self_need_and_a_cycle`, `next_skips_a_card_whose_prerequisite_is_not_done_and_offers_it_when_it_is` 통과.
- AC-3: met — `next_skips_a_held_card_with_its_category_and_reason`, `take_refuses_a_held_card_and_a_card_with_an_open_prerequisite_and_writes_nothing`(HOME 폴더 스냅샷이 같고 예약 행 0) 통과.
- AC-4: met — `show_carries_the_hold_and_the_unfinished_prerequisites` 통과.
- AC-5: met — `guidance_tells_a_lane_to_read_hold_and_prerequisites_before_it_creates_a_worktree` 통과, 매니페스트는 `tools/source_manifest.py write` 로 다시 만들었어요(`check` 는 `moai check affected` 의 단계로 돌아요), `help_layout` 11개와 `help_table_names_every_command_once`(bin 시험 안) 통과. 도움말 행·골든은 위 (2) 때문에 바꿀 것이 없었어요.
- AC-5b: met — `hold_candidates_lists_free_text_holds_read_only_until_a_person_moves_one` 통과(HOME 스냅샷이 같고, 사람이 `card hold` 를 부른 뒤에만 목록과 `ready` 에서 빠져요).
- AC-6: met — 위 표(뮤턴트 여섯, 모두 복원).
- 틈(후속으로 남겨요): 뮤턴트는 S1 코드(원장·`next`·`take`)에만 돌렸고 `hold_candidates.rs` 의 낱말 일치와 명령 포장 함수(`hold`·`unhold`·`prereq`)는 뮤턴트를 돌리지 않았어요(AC-6 은 여섯 이하). 세 카드 이상의 선행 고리, 틈 (3). `card scope` 사건의 `--by` 에 `leader` 를 적었어요(리더 노트 8280 의 위임이고, 사건은 이 세션 id 도 남겨요).
- 뒤의 SPEC(`SPEC-20261127-card-prereq-remove`) 이 바꾼 것 1: REQ-5 는 이제 두 카드 짝만이 아니라 어떤 길이의 고리도 거절하고 고리의 카드를 차례로 말해요. 앞의 틈 (3)(세 카드 이상의 선행 고리)은 이것으로 닫혀요. `card prereq --remove` 로 링크를 지울 수 있어요.
- 뒤의 SPEC 이 바꾼 것 2: `card hold-candidates`(REQ-11)는 코드대로 `draft`·`ready`·`active`·`review` 의 열린 카드를 나열해요. 이 문서의 옛 문구(끝나지 않은 카드)는 코드와 어긋났고 코드가 맞아요.
