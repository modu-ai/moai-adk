---
id: SPEC-20261125-flow-measure
title: "moai flow-measure prints the flow stage times from the ledger without writing, as a Rust command"
status: draft
revision: 1
created: 2026-11-25
parent: null
---

# moai flow-measure prints the flow stage times from the ledger without writing, as a Rust command

## Interview

- Q1: 구간 시간(제출→판정, PASS→done 따위)을 무엇이 내나요? -> A: "원장을 읽기 전용으로 읽는 moai 하위 명령 하나가 내요. Python 스크립트는 정의를 읽는 참고일 뿐 돌리지 않아요. 카드 이름을 주면 카드별 한 줄과 PASS→done 의 P50/P90(nearest rank, 원값 함께), `--hours N` 을 주면 그 창의 집계를 내요. JSON 답과 사람이 읽는 표 둘 다요."(사용자 지시 2026-11-24T22:18Z 를 전한 리더 배차, 흐름 뼈대 카드의 기준 2)
- Q2: 시간을 어떻게 정의하나요? -> A: "제출 = PASS 앞의 가장 새 READY/REPAIRED/LANE-READY-INTEGRATION 노트, 판정 = verdict=pass 인 card.verdict 이벤트(행위자가 감사자 세션), done = card.finish 이벤트, 반납 = 예약이 held 를 떠난 시각."(같은 배차, 감사 보고서와 같은 정의)

## Intent

개선이 먹혔는지는 구간 시간으로 봐요. 지금 그 숫자는 Python 스크립트(`measure.py`)가 내는데, 이 저장소는 Python 을 쓰지도 돌리지도 않아요(CLAUDE.local.md). 이 SPEC 은 같은 숫자를 `moai flow-measure` 가 원장에서 읽기 전용으로 내게 해서, 감사 보고서의 정의와 같은 값을 누구나 같은 명령으로 다시 구하게 해요. 읽는 사람: 리더, 감사자, 개선 카드의 작성자.

## Non-goals

- 원장에 쓰지 않아요: 숫자를 파일이나 `metrics.jsonl` 에 덧붙이는 일은 이 명령에 없어요(출력은 호출한 쪽이 보관해요).
- 영수증 폴더(`check-affected-*`)의 슬롯 대기와 테스트 시간은 읽지 않아요. 이 SPEC 은 원장과 git 이 아닌 원장만 읽어요(Decisions D-6). 영수증 숫자와 리더 노트 길이는 뒤 카드의 몫이에요.
- 카드 상태 판정과 감사 판정 규칙은 바꾸지 않아요. 이 명령은 `card.verdict` 와 `card.finish` 이벤트를 읽기만 해요.
- 이 일을 맡긴 카드의 기준 3 과 그 밖의 뼈대 일(`card close`, 열차, 신호, 모드)은 하지 않아요.

## Requirements

<!-- This SPEC adds no blocking check: the command is a report. -->
- REQ-1: When `flow-measure` runs, the command shall open the ledger read-only, write nothing and start no host program, and answer a project without a ledger with `ledger: "absent"`, empty lists and exit 0. · failure: none — the command is a report.
- REQ-2: When a card is measured, the command shall take the submit time from the newest hand-back note not later than the newest `pass` verdict, the verdict time from that `card.verdict` event, the done time from the first `card.finish` event at or after it, and the return time from the moment its reservation left `held`. · failure: none — a missing end is `null`.
- REQ-3: When card names are given, the command shall print for each card the seconds from submit to verdict and from pass to done, the number of `fail` verdicts, the auditor session, the 40-hex head and the reservation state, and `null` for a span whose two ends are not both observed. · failure: none — `null` is the answer.
- REQ-4: When a set of spans is summarised, the command shall give its count, its P50 and its P90 by nearest rank (the value at rank ceil(p × n) of the ascending list), the raw values too for the named cards, and `null` and never zero for no span. · failure: none — a small sample is shown as it is.
- REQ-5: When `--hours <N>` is given, the command shall count the `card.add`, `card.landed` and `card.finish` events of the window separately, give the count, P50 and P90 of each stage whose end event falls in the window, give the share of cards whose first verdict in the window is `fail`, and count the `LANE-BLOCKED` notes of the window by their `kind`. · failure: none — a window with no events answers zero counts and `null` summaries.
- REQ-6: When the command answers, it shall give the JSON answer and, in the same answer, a `table` of lines a person reads, whose lines carry the same numbers. · failure: none — the table is derived from the JSON values.
- REQ-7 (REQ-1): IF neither a card name nor `--hours` is given, or a card name or the hours value is not valid, or a named card does not exist, THEN the command shall refuse with a message that names the argument, exit non-zero and write nothing.

## Acceptance

- AC-1 (REQ-1): 원장이 없는 프로젝트에서 `flow-measure --hours 12` 는 `ledger: "absent"` 와 빈 목록으로 종료 0 이고 원장 파일이나 폴더를 만들지 않아요; 원장이 있을 때 실행 앞뒤로 원장 파일의 바이트가 같아요 · verify: `cargo test --locked --offline --test flow_measure_cli a_project_without_a_ledger_answers_absent_and_a_ledger_is_not_written`
- AC-2 (REQ-2, REQ-3): 첫 READY, FAIL 판정, REPAIRED, PASS 판정, finish 가 있는 카드의 한 줄이 정의대로 초 단위로 나와요: 제출은 PASS 앞의 가장 새 노트(REPAIRED), 판정은 pass 이벤트, done 은 finish 이벤트이고, `fail_count` 1, 감사자 세션, head, 예약 상태가 들어 있어요. 리더가 쓴 노트와 결정 노트는 첫 줄이 표지여도 제출이 아니에요 · verify: `cargo test --locked --offline --test flow_measure_cli a_card_line_reads_submit_verdict_and_done_by_the_definitions`
- AC-3 (REQ-3): 판정이 없는 카드와 finish 가 없는 카드는 끝이 없는 구간이 `null` 이고 0 이 아니에요 · verify: `cargo test --locked --offline --test flow_measure_cli a_span_with_a_missing_end_is_null_and_never_zero`
- AC-4 (REQ-4): 여러 카드의 PASS→done 은 nearest rank 의 P50/P90 과 개수와 원값을 내고(5개면 3번째와 5번째 값), 표본 하나는 둘 다 그 값이며, 표본이 없으면 `null` 이에요 · verify: `cargo test --locked --offline --test flow_measure_cli the_percentiles_are_nearest_rank_and_carry_the_raw_values`
- AC-5 (REQ-5): `--hours` 의 창 안과 밖에 걸친 `card.add`, `card.landed`, `card.finish` 이벤트가 따로 세어지고 창 밖 이벤트는 세지 않아요 · verify: `cargo test --locked --offline --test flow_measure_cli a_window_counts_issued_landed_and_finished_apart`
- AC-6 (REQ-5): 창의 단계별 P50/P90 과 표본 수, 첫 판정 FAIL 비율, `LANE-BLOCKED` 의 `kind` 별 수가 고정 픽스처에서 정한 값이에요(`kind` 가 없는 노트는 `unspecified`) · verify: `cargo test --locked --offline --test flow_measure_cli a_window_gives_stage_percentiles_first_verdict_fail_share_and_block_kinds`
- AC-7 (REQ-6): 두 모드의 답에 `table` 줄이 있고 줄이 카드 이름과 같은 숫자를 말해요 · verify: `cargo test --locked --offline --test flow_measure_cli the_table_lines_carry_the_numbers_of_the_json`
- AC-8 (REQ-7): 인자가 없거나, 카드 이름이 틀리거나, 없는 카드이거나, `--hours` 가 0 이거나 숫자가 아니면 어느 인자인지 말하고 종료가 0 이 아니며 원장이 바뀌지 않아요 · verify: `cargo test --locked --offline --test flow_measure_cli a_bad_argument_is_refused_by_name_and_writes_nothing`
- AC-9: 이웃 테스트가 그대로 통과해요(도움말 배치와 명령 등록) · verify: `cargo test --locked --offline --test help_layout` and `cargo test --locked --offline --lib flow_measure`

## Scope

- write: `.moai/specs/SPEC-20261125-flow-measure/**`, `src/flow_measure.rs`, `src/ops/flow_measure.rs`, `docs/design-review/appendix/operations/flow_measure.md`, `tests/flow_measure_cli.rs`
- shared (line lease, 한 줄씩): `src/lib.rs`(`pub mod flow_measure;`), `src/help.rs`(명령 묶음 한 줄, 예제 한 줄)
- read-only: `src/ledger/**`, `src/card/ready_receipt.rs`(`marker_of_line`), `src/card/input.rs`(`parse_ref`), `src/leader_brief.rs`(열기 방식의 본보기)
- forbidden: 원장에 쓰는 모든 코드, `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME, 보고서 폴더의 Python 스크립트(정의를 읽기만 하고 돌리지 않아요)

## Plan

1. SPEC 초안을 쓰고 커밋해요. 승인은 레인이 기록하지 않아요.
2. RED: `tests/flow_measure_cli.rs` 에 AC-1 ~ AC-8 을 써요. 고정 시각의 SQL 픽스처 원장으로 `flow-measure` 가 아직 없어서 실패하는 것을 봐요(명령이 없다는 이유가 옳은 이유).
3. GREEN: `src/flow_measure.rs` 에 읽기 전용 열기, 한 읽기 트랜잭션 안의 세 조회(이벤트, 노트, 예약), 순수 계산(`measure`), 표 그리기를 만들고, `src/ops/flow_measure.rs` 와 문서 조각, `lib.rs`·`help.rs` 한 줄을 더해요.
4. 확인: 새 테스트, AC-9 의 이웃 테스트, `cargo fmt --all`, 새 가드 테스트마다 가드 뮤턴트 한 줄(Python 을 쓰지 않아요: 뮤턴트는 편집 도구로 한 줄씩 고쳐 해당 테스트만 돌리고 `git checkout --` 로 되돌려요), 위험 경로는 읽기 전용 명령이라 해당 없음, `moai check affected --card` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적어요. 감사는 이 변경을 쓰지 않은 세션이 맡아요.

## Risks

- 이벤트 시각의 한계: `card.verdict` 이벤트가 없는 옛 카드(노트 `AUDIT … verdict:` 로만 판정한 카드)는 판정이 `null` 이에요. 정의가 이벤트를 요구하므로 노트 판정은 세지 않아요(알려진 한계). 읽지 못한 판정 이벤트(`verdict` 가 `pass`·`fail` 이 아니거나 payload 가 JSON 이 아닌 것)는 답의 `unreadable_verdicts` 가 세요.
- 재개된 카드: 한 카드가 `done → ready` 로 되돌아가 finish 가 둘이면 PASS 뒤 첫 finish 를 쓰고, 그 앞의 PASS 는 세지 않아요(가장 새 pass 판정이 기준).
- 원장이 클수록 `event` 테이블 읽기(`card_id` 인덱스가 없음)가 한 번 전체를 훑어요: 종류로 좁혀 한 번만 읽고 읽기 트랜잭션을 계산 전에 닫아요.
- 시계: 창의 끝은 `--now`(없으면 SQLite 의 `unixepoch()`)예요. 같은 원장과 같은 `--now` 는 같은 바이트를 내요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 최상위 명령 `flow-measure`(ops 파일 하나). 대안: `card` 아래 하위 명령 | 리더 배차가 두 이름을 허용했고 ops 파일은 공유 파일(`src/card/command.rs`)을 건드리지 않아요. 기본: 최상위 | 하위 명령으로 옮겨요 |
| D-2 | 구간은 초 정수로 JSON 에 내고 표는 분(소수 한 자리)으로 보인다 | 부동소수 비교 없이 테스트가 정확해요. 기본: 초 | 표 단위만 바꿔요 |
| D-3 | 카드는 위치 인자(`C-<번호>` 만 받아요), 프로젝트는 `--project`(기본 `.`) | 배차가 "카드 이름을 주면"이라 했어요. 기본: 이대로 | 인자 모양만 바꿔요 |
| D-4 | `--now <unix 초>` 로 창의 끝을 고정할 수 있다 | 고정 픽스처 테스트와 같은 숫자 재현. 기본: 없으면 SQLite 시각 | 옵션을 지워요 |
| D-5 | 창의 단계는 끝 이벤트의 시각이 창 안에 있으면 그 창의 표본이다(`take_to_submit`, `submit_to_verdict`, `pass_to_landed`, `pass_to_done`) | 기존 스크립트가 착지 시각으로 창을 잡았어요; 끝 시각이 한 가지 기준이에요. 기본: 끝 시각 | 시작 시각 기준으로 바꿔요 |
| D-6 | 영수증 폴더(슬롯 대기, 테스트 시간)와 리더 노트 길이는 이 SPEC 에서 빼고 원장만 읽는다 | 기준 2 의 정의가 원장 사건이에요; 영수증 폴더는 워크트리마다 흩어져 읽기 전용 보장이 어려워요. 기본: 뺌 | 뒤 카드로 더해요 |

## Result

`status: draft`; the approval is the user's and is not recorded here. Implemented in `src/flow_measure.rs` (read, compute, table), `src/ops/flow_measure.rs` (arguments, help), one line each in `src/lib.rs` and `src/help.rs` (line lease, given back at once) and the four `all*` help goldens (one line each, written with `MOAI_HELP_GOLDEN_WRITE` under a line lease), the document fragment `docs/design-review/appendix/operations/flow_measure.md`. No Python was written or run; the Python scripts of the reports folder were read for their definitions only.

- RED (`target/fm/red1.log`): the 8 new tests failed with `error: unrecognized subcommand 'flow-measure'` before the command existed.
- GREEN: `cargo nextest run --locked --offline --test flow_measure_cli --lib -E "binary(flow_measure_cli) | test(/^flow_measure::/)"` → `11 tests run: 11 passed` (8 integration, 3 unit; `target/fm/green4.log`). Neighbors `help_layout`, `help_english_terminal_cli`, `help_text_names_moai`, `ops_table_gen_cli`, `hot_file_ops_merge` and the 8 new tests → `33 tests run: 33 passed` (`target/fm/neighbors2.log`); the first neighbor run failed only on the four goldens that now hold the new line.
- AC-1 met: `a_project_without_a_ledger_answers_absent_and_a_ledger_is_not_written` passes in the GREEN run.
- AC-2 met: `a_card_line_reads_submit_verdict_and_done_by_the_definitions` passes in the GREEN run.
- AC-3 met: `a_span_with_a_missing_end_is_null_and_never_zero` passes in the GREEN run.
- AC-4 met: `the_percentiles_are_nearest_rank_and_carry_the_raw_values` passes in the GREEN run.
- AC-5 met: `a_window_counts_issued_landed_and_finished_apart` passes in the GREEN run.
- AC-6 met: `a_window_gives_stage_percentiles_first_verdict_fail_share_and_block_kinds` passes in the GREEN run.
- AC-7 met: `the_table_lines_carry_the_numbers_of_the_json` passes in the GREEN run.
- AC-8 met: `a_bad_argument_is_refused_by_name_and_writes_nothing` passes in the GREEN run.
- AC-9 met: `help_layout` and the other neighbors pass in the 33-test run, and the 3 unit tests of `--lib flow_measure` pass in the GREEN run.
- Guard mutants: 21 of 21 killed, one exact-once fixed-string edit each, restored with `git checkout --`, `git status --short` empty after the batch (shell and awk, `target/fm/mutants.sh`, output `target/fm/mut.out`). M21 survived the first batch because the fixture landed the old card before its pass; the fixture was fixed and M21 was killed on the next run (`target/fm/mut21.out`). The read-only open (`open_read_only_within`) has no mutant: a write open is another type and does not compile.
- Dangerous path: the command only reads. One run against the real ledger, read-only (`flow-measure <two cards> --hours 12 --project .`, `writes: false`), gave a card line for the landed card (submit to verdict 3.2 min, pass to done 15.8 min, 3 failed verdicts, auditor session, head `dfed56351e40`, reservation released) and a window of issued 24, landed 32, finished 23, 12 blocks (11 decision, 1 environment); the ledger files were not opened for writing.
- Not run: the whole `cargo test`, Linux and Windows, a ledger with `card.verdict` events of an older payload shape, the receipt folders (Decision D-6).
