---
id: SPEC-20261124-ready-receipt-followup
title: "READY 영수증 관문 후속: 표식 판정 하나와 맨몸 실행 영수증만 받기"
status: approved
revision: 1
created: 2026-11-24
parent: SPEC-20261124-ready-receipt-gate
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7865 (2026-11-23T20:37:01Z, 사용자 위임 2026-11-23T13:57Z, 노트 7862 의 결정 1~4 를 기본값 B, b, a, a 로 승인, SPEC 커밋 a9f9e3080d); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:0f9f9e99c25b16e7cd43fa5ecffab9a04316aa27f0fed48ff6a05962a81a7122"
---

# READY 영수증 관문 후속: 표식 판정 하나와 맨몸 실행 영수증만 받기

## Intent

카드 C-420 은 C-414 감사(AUDIT #414)의 비차단 N1·N2·N3·N5 를 닫아요(카드 본문). 관문(`SPEC-20261124-ready-receipt-gate`)은 "돌리지 않고 올림"을 막는 장치인데 아직 두 군데로 빠져나갈 수 있어요.

읽어서 확인한 사실(코드, 2026-11-24, NeXT `6fd89cec16`):

- N1: 관문은 첫 낱말이 마커와 정확히 같아야 마커로 봐요(`src/card/ready_receipt.rs` 의 `marker_head`). 그런데 `card inbox`(`src/card/inbox.rs` 의 `family_of`)와 상태줄(`src/statusline/state.rs` 의 `marker_of`)은 `starts_with` 로 READY 를 세요. 그래서 `LANE-READY-INTEGRATION: C-1 ...` 처럼 마커 뒤에 글자가 붙은 줄은 장부에는 영수증 없이 들어가고 inbox 와 상태줄은 그것을 READY 로 세요. 지금 테스트가 `REPAIRED-X C-1 head=..` 를 "마커가 아님"으로 못박아 둬요.
- N2: `skills/moai/references/flows/next-flow.md:25` 는 `LANE-READY-INTEGRATION`, READY, REPAIRED, MERGE-READY 를 모두 hand-back 이라고 적어요. 관문은 앞의 하나와 REPAIRED 만 걸러서, 레인이 `READY (head abc)` 처럼 쓰면(상태줄이 `READY ` 를 Submitted 로 세요) 영수증 없이 받아져요. 같은 갈래의 우회 길이에요.
- N3: `moai check affected --lane`(C-406, 착지)도 `result=PASS` 요약 줄을 `summary.txt` 로 남기고, 줄에는 선택이 적히지 않아요(`src/check_affected.rs:1070-1084`; 답의 `selection` 필드에만 있어요). 그래서 부분 실행 영수증이 READY 로 받아져요. 그런데 `docs/design-review/appendix/operations/check.md:25` 는 "레인이 READY 를 쓰기 전에는 `--lane`" 이라고 적어요 — 카드의 N3 와 이 문장은 서로 반대예요(결정 3).
- N5: `card note --help` 는 관문을 말하지 않아요. 도움말은 `src/ops/card.rs`(`Note`, 579 줄 근처)에 있고, 그 파일은 지금 C-428 이 예약하고 있어요(결정 4).
- 요약 줄을 읽는 곳(`git grep` 로 찾음): `src/card/ready_receipt.rs`(`judge`), `tests/common/ready_receipt.rs`(`summary_line` 만드는 도우미), `tests/check_affected_cli.rs`·`tests/check_affected_one_slot.rs`(`ends_with(" result=…")`, 한 곳 `contains(" scans=listed:1 test-threads=4 ")`), `docs/design-review/appendix/operations/check.md:54`, `src/ops/check.rs` 도움말. `bundle` 도구와 스킬 문서는 요약 줄을 읽지 않아요.

## Interview

- Q1: 마커 판정은 어디에 두나요? → A: `src/card/ready_receipt.rs` 에 함수 하나(`marker_of_line`)를 두고 관문·`card inbox`·상태줄이 그것을 불러요(D-1).
- Q2: 마커 뒤에 글자가 붙은 줄은요? → A: 마커로 시작하면 마커예요(`starts_with`). 마커인데 `<마커> <카드> head=<40 hex>` 모양이 아니면 거절해요(관문의 기존 거절). `REPAIRED-X C-1 head=..` 도 이제 거절돼요(D-1).
- Q3: READY·MERGE-READY 도 관문에 넣나요? → A: 결정 1.
- Q4: 부분 실행 영수증은요? → A: 요약 줄에 `selection=lane` 낱말을 더하고 관문이 그 줄을 거절해요(결정 2, 결정 3).

## Non-goals

- 영수증 위조 방지(손으로 쓴 줄은 막지 못해요, 앞 SPEC 의 Q3 와 같아요).
- `MERGE-READY` 노트에 영수증 요구(리더의 노트이고 모양이 달라요. 결정 1 의 기본값).
- 상태줄이 `REPAIRED` 를 Submitted 로 세게 하는 일(이 SPEC 은 판정을 맞추기만 해요).
- `check affected` 의 단계 선택·실행 바꾸기(요약 줄에 낱말 하나만 더해요).

## Requirements

- REQ-1: The `card note` gate and `card inbox` and the status line shall call the one function `marker_of_line` of `src/card/ready_receipt.rs` to decide whether a first line starts with a hand-back marker.
- REQ-2 (REQ-1): WHEN a first line starts with a marker, the `card note` command shall count it as that marker whatever character follows the marker.
- REQ-3 (REQ-2): IF a line that starts with a marker does not read `<marker> <card> head=<40 lowercase hex>` for the card of the note, THEN the `card note` command shall refuse the note and write nothing to the ledger.
- REQ-4 (REQ-1): The hand-back markers shall be `LANE-READY-INTEGRATION`, `REPAIRED` and `READY ` (the word and one space).
- REQ-5 (REQ-4): IF a note whose first line starts with `READY ` names a head that has no PASS receipt, THEN the `card note` command shall refuse the note and write nothing to the ledger.
- REQ-6: WHEN `moai check affected` runs with `--lane`, it shall write the word `selection=lane` into its summary line right after the `scans=` word.
- REQ-7 (REQ-6): IF the summary line of a receipt holds the word `selection=lane`, THEN the `card note` command shall count it as no receipt and say that it is a `--lane` run.
- REQ-8: The help of `card note` shall name the receipt gate: the markers it judges, the receipt file and the command that makes it.
- REQ-9: A `moai check affected` run without `--lane` shall write its summary line without the word `selection=lane`, as before.

## Acceptance

- AC-1 (REQ-1, REQ-2): `LANE-READY-INTEGRATION: C-1 head=..` and `REPAIRED-X C-1 head=..` are refused by `card note` without a receipt, and the same first lines are counted as READY by `card inbox` and as Submitted by the status line, so the three judge a line alike for a table of first lines. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli --lib -E 'test(marker_of_line) | test(a_glued_marker_is_one_rule_for_gate_inbox_and_status_line)'`
- AC-2 (REQ-3): a line that starts with a marker and has the wrong shape (short head, other card, no `head=`, glued text) is refused with the shape in the sentence and the ledger holds no new note. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_marker_line_of_the_wrong_shape_is_refused)'`
- AC-3 (REQ-4, REQ-5): `READY C-1 head=<head>` is refused without a PASS receipt and kept with one, and `MERGE-READY ..` is kept without one. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_plain_ready_note_needs_the_receipt_and_merge_ready_does_not)'`
- AC-4 (REQ-6, REQ-9): `check affected --lane` leaves a `summary.txt` whose line holds `selection=lane` right after `scans=`, and a bare run leaves the line without it. · verify: `cargo nextest run --offline --locked --test check_affected_cli -E 'test(the_summary_line_says_a_lane_run)'`
- AC-5 (REQ-7): a PASS receipt line with `selection=lane` is refused and the sentence names the `--lane` run and the bare command to run. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_lane_receipt_is_refused_with_the_bare_command)'`
- AC-6 (REQ-8): `card note --help` names `summary.txt`, `LANE-READY-INTEGRATION`, `REPAIRED`, `READY` and `moai check affected`, and still holds no skip word. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(other_notes_are_kept_and_there_is_no_skip_flag) | test(the_note_help_names_the_receipt_gate)'`
- AC-7 (REQ-6, REQ-7): the operations page names `selection=lane`, the markers and the bare run for a READY receipt, and no longer says that a lane writes READY after `--lane`. · verify: `cargo nextest run --offline --locked --test check_affected_modes_cli -E 'test(the_operations_page_names_the_lane_word)'`

## Scope

- write: `src/card/ready_receipt.rs`, `src/card/inbox.rs`, `src/statusline/state.rs`, `src/check_affected.rs`, `src/ops/check.rs`, `tests/card_ready_receipt_cli.rs`, `tests/common/ready_receipt.rs`, `tests/check_affected_cli.rs`, `tests/check_affected_modes_cli.rs`, `docs/design-review/appendix/operations/check.md`, `.moai/specs/SPEC-20261124-ready-receipt-followup/spec.md`
- 범위 확장 요청(결정 4): `src/ops/card.rs`(C-428 예약 중)
- read-only: `src/card/work.rs`(관문 호출), `src/bundle.rs`, `skills/moai/references/flows/next-flow.md`
- forbidden: 실제 HOME 과 실제 장부, 네트워크, `CLAUDE.md`, `tools/repo_scans.json`(C-418), 영수증을 우회하는 플래그

## Plan

1. RED: AC 의 테스트를 먼저 써서 실패를 봐요(표식 판정 표, 풀어쓴 줄 몇 개, 부분 실행 요약 줄).
2. `marker_of_line` 을 `ready_receipt.rs` 에 두고 `marker_head` 가 그것을 써요; `inbox.rs` 의 `family_of` 와 `state.rs` 의 `marker_of` 가 그것을 불러요.
3. `check_affected.rs` 요약 줄에 `selection=lane` 낱말(`--lane` 일 때만)을 더하고 `judge` 가 거절해요; 요약 줄을 읽는 곳(Intent 의 목록)을 같이 맞춰요.
4. 도움말(`src/ops/card.rs`)은 결정 4 가 풀린 뒤에 마지막으로 해요.
5. `docs/design-review/appendix/operations/check.md` 의 25·54 줄과 `src/ops/check.rs` 도움말을 맞춰요.

## Risks

- `REPAIRED-X ...` 처럼 마커로 시작하는 다른 노트가 있으면 이제 거절돼요(모양을 안내하는 문장과 함께). 장부의 실제 첫 줄은 구현 전에 `moai card` 로 읽어 `REPAIRED` 와 `READY ` 로 시작하는 줄의 종류를 세겠어요(unverified).
- 이 SPEC 이 착지하면 `READY (head abc)` 같은 옛 첫 줄은 거절돼요. 레인은 `READY <카드> head=<40 hex>` 나 `LANE-READY-INTEGRATION` 으로 써야 해요.
- `--lane` 으로 READY 를 쓰던 흐름(`check.md:25` 의 문장)은 맨 명령으로 바뀌어요: 레인은 개발 중에는 `--lane`, READY head 에서는 맨 명령 한 번이에요. 한 번 더 도는 비용이 늘어요.
- 새 빌드의 `moai` 가 쓴 요약 줄만 낱말을 가져요; 맨 명령 줄은 그대로라서 옛 영수증은 받아져요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 마커 판정은 `starts_with` 하나(`ready_receipt::marker_of_line`)로 모아 관문·inbox·상태줄이 불러요 | 카드 N1; inbox 와 상태줄이 이미 `starts_with` 라서 둘을 바꾸지 않고 관문만 맞춰요 | 함수 호출을 되돌리기 |
| D-2 | 관문의 마커는 `LANE-READY-INTEGRATION`, `REPAIRED`, `READY `(기본값, 결정 1 의 B) | 카드 N2; `READY ` 가 상태줄에서 Submitted 로 세어지는 우회 길이에요 | 마커 목록에서 하나 빼기 |
| D-3 | `--lane` 실행은 요약 줄에 `selection=lane` 낱말을 더하고, 맨 명령 줄은 그대로예요(기본값, 결정 2 의 b) | 카드 N3; 기존 줄과 읽는 곳을 바꾸지 않아요 | 낱말 한 곳과 `judge` 한 줄을 지우기 |
| D-4 | 관문은 `--lane` 영수증을 거절해요; 레인은 READY head 에서 맨 명령을 한 번 돌려요(기본값, 결정 3) | 카드 N3 | `judge` 의 한 줄과 `check.md:25` 문장을 되돌리기 |

## Result

(구현 뒤에 채워요.)
