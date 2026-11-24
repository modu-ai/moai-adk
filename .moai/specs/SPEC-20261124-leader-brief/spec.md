---
id: SPEC-20261124-leader-brief
title: "`moai leader brief` 가 4KB 이하 상태로 착지 대기열, 막힘, 결정 필요를 노트 번호와 함께 보여줘요"
status: draft
revision: 1
created: 2026-11-24
---

# `moai leader brief`: 리더가 한 번 깨어날 때 읽는 짧은 상태

## Intent

사용자 지시(2026-11-24T13:25:44Z, Fable 조언 "리더가 하나뿐인 병목")와 카드 C-477 의 측정이 출발점이에요. 12시간 동안 노트 484개(660KB) 가운데 245개를 리더가 썼고, 리더는 깨어날 때마다 원장 노트 본문과 레인 화면을 직접 읽어요. 지금 한 번 깨어날 때 읽는 길은 `moai card inbox`(2026-11-24 쟀을 때 7658바이트)와 `moai card orchestrate plan`(17613바이트)이고, 거기서 고른 노트의 본문을 따로 열어요.

이 SPEC 은 그 읽기를 `moai leader brief` 하나로 줄이는 조각이에요. brief 는 원장과 레인 등록부를 읽기만 하고, 착지 대기열·막힘·결정 필요 셋만 4KB 이하로 보여 줘요. 항목마다 노트 번호를 달아서 리더가 필요한 항목의 본문만 `moai card show` 로 열게 해요. 새 계산은 만들지 않고, 이미 읽기 전용으로 착지한 `card orchestrate plan` 의 답을 줄여 써요(`src/card/orchestrate.rs`). 그래서 대기열의 판정과 막힘 목록이 plan 과 어긋날 수 없어요.

카드 C-477 은 기준이 셋이고, 이 SPEC 은 두 개만 해요.

| 카드 기준 | 이 SPEC | 까닭 |
| --- | --- | --- |
| 1. brief 가 4KB 이하로 대기열·막힘·결정만 보여 주고 항목마다 노트 번호 | 한다(1조각) | 읽기 전용이라 바로 할 수 있어요 |
| 2. 착지 대기열 실행기가 PASS head 3개를 묶어 착지하고 실패 하나를 나눠 빼요 | 하지 않아요(2조각) | C-470 의 `card land-batch` 위에 올려야 하는데 C-470 은 수리 중이에요. C-470 이 착지한 뒤 2조각으로 해요 |
| 3. 리더 깨움 1회당 읽는 양을 전후로 잰 노트 | 한다(1조각) | brief 하나로 바꿨을 때의 바이트 수를 같은 원장에서 비교해요 |

## Non-goals

- 기준 2: 착지 대기열 실행기, 묶음, 실패 시 둘로 나누기, compare-and-swap. C-470 이 착지한 뒤 2조각의 몫이에요(Plan 3).
- 카드 본문의 (2) 감사 결과가 리더 대화로 돌아오지 않고 감사 세션이 판정 노트만 쓰는 길. 이 SPEC 은 건드리지 않아요.
- 노트에 답하기, 배차, `/clear`, 감사 실행, 착지. brief 는 아무것도 보내거나 시작하거나 쓰지 않아요.
- 레인 화면 읽기. brief 는 `orca` 도, `claude` 도 실행하지 않고 화면을 읽지 않아요. 화면 바이트는 측정에 넣지 않아요(Result 에 `not_run` 으로 적어요).
- `src/leader_round.rs`(C-469 의 몫)와 `card orchestrate plan`, `card inbox`, `card land` 의 출력 모양 바꾸기. brief 는 그 읽기 경로를 부르기만 해요.
- 스킬과 워크플로 글(`skills/moai/workflows/leader.md`). 그 글이 brief 를 부르게 하는 일은 그 글을 가진 카드의 몫이에요.

## Requirements

- REQ-1: When `moai leader brief` runs, the command shall print one JSON answer whose `result` holds `counts`, `queue`, `blocked`, `decisions` and `omitted` and whose whole envelope is at most 4096 bytes.
- REQ-2 (REQ-1): IF the three lists do not fit in 4096 bytes, THEN the command shall drop the last item of the longest list until the answer fits, keep the full totals in `counts` and print the dropped numbers in `omitted`.
- REQ-3 (REQ-1): The command shall carry on each `queue` item the card, the full 40-hex head and `ref` (`note:<id>` or `event:<id>` of the verdict that put it there), and on each `blocked` and `decisions` item the card, `note` (the note number) and the first line cut at 96 bytes.
- REQ-4 (REQ-1): When an unanswered note of a card that is not `done` starts with `DECISION-QUESTION`, or starts with `LANE-BLOCKED` and says `kind=decision`, the command shall list it under `decisions`, and every other unanswered `LANE-BLOCKED` note under `blocked`, newest note first · failure: none — an answered note, a note of a `done` card and a note that starts with another word are listed nowhere.
- REQ-5: When `moai leader brief` runs, the command shall write nothing and answer `writes` false · failure: none — the ledger file, the registry folder and the state folder stay byte for byte as they were.
- REQ-8 (REQ-5): When `moai leader brief` runs twice on an unchanged ledger and registry, the command shall print byte for byte equal answers · failure: none — the answer holds no clock value.
- REQ-6: IF the project has no ledger, THEN the command shall print empty lists with `ledger: "absent"` and exit 0 without creating anything.
- REQ-7 (REQ-1): When the integration branch or the lane registry cannot be read, the command shall list that fact in `unread` and still print the other parts · failure: none — an unreadable source is never shown as an empty queue.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5, REQ-6, REQ-8): on a temporary ledger with two audited heads, one blocked note and one decision note, `brief` prints the five parts, every item carries its number, `writes` is false, the home folder is byte for byte as before and two runs are equal; on a project with no ledger every list is empty with `ledger: "absent"` and exit 0 · verify: `cargo test --locked --offline --test leader_brief_cli shape_`
- AC-2 (REQ-2): on a ledger with 60 blocked notes and 40 audited heads, the whole stdout is at most 4096 bytes, `counts` holds the full totals, `omitted` plus the shown items equal them and the newest notes are the ones kept · verify: `cargo test --locked --offline --test leader_brief_cli size_`
- AC-3 (REQ-4): a `LANE-BLOCKED … kind=decision` note and a `DECISION-QUESTION` note are listed under `decisions`, a `LANE-BLOCKED … kind=claim` note under `blocked`, an answered note (a later note by the leader on the card) and a note of a `done` card are listed nowhere · verify: `cargo test --locked --offline --test leader_brief_cli split_`
- AC-4 (REQ-7): a project whose integration branch git cannot read lists `integration` in `unread` and still prints the blocked and decision parts · verify: `cargo test --locked --offline --test leader_brief_cli unread_`
- AC-5 (REQ-1): card criterion 3; on the same throwaway ledger the brief is smaller than `card inbox` plus `card orchestrate plan` together, and the measured bytes of both ways on the real ledger are written in the Result · verify: `cargo test --locked --offline --test leader_brief_cli measure_`
- AC-6: the neighbour checks pass: the help fixtures, the operation document test, the test area layout and the style checks · verify: `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings` and `moai check affected <base>..<head>` at the READY head

## Scope

- write: `.moai/specs/SPEC-20261124-leader-brief/**`, `src/leader_brief.rs` (new, the pure shrink of a plan and the one read), `src/ops/leader.rs` (new, the `leader` command word and its `brief` action), `tests/leader_brief_cli.rs` (new), `docs/design-review/appendix/operations/leader.md` (new)
- write by line lease (shared files, one line each, given back right after the commit): `src/lib.rs` (`pub mod leader_brief;`), `src/help.rs` (the `leader` row and its example), `tests/fixtures/help/*.txt` (the golden help), `tests/areas.toml` (the new test file's area, when the layout test asks)
- read-only: `src/card/orchestrate.rs` (`plan`), `src/card/inbox.rs`, `src/card/landed.rs`, `src/lane_registry.rs`, `src/leader_round.rs` (C-469 owns it)
- forbidden: a real `claude` or `orca` process, any model call, any network, the real HOME, any write of a card, a note, a reservation, a verdict, a registry entry or a ref by the brief, every skill and workflow text, the integration branch

## Plan

1. This SPEC, read by the leader. The leader's approval is a decision note that names the digest of this revision; the lane writes no approval.
2. Slice 1 (this SPEC), tests first: `tests/leader_brief_cli.rs` with the prefixes `shape_`, `size_`, `split_`, `unread_` and `measure_` on a temporary home and a temporary project; RED on the missing command; then `src/leader_brief.rs` and `src/ops/leader.rs`; the measurement of both ways on the real ledger (read only); at most three guard mutants; one bare `moai check affected` at the READY head; the READY note, which says that card criterion 2 stays open.
3. Slice 2 (a later revision or a follow-up card, after C-470 has landed `card land-batch`): the landing-queue runner, which takes the PASS heads of the brief's `queue`, lands them as one bundle, splits a failing bundle in two to find the head that breaks it and moves the integration branch by one compare-and-swap, with no leader turn. Card criterion 2 belongs to it. This slice writes nothing of it.

## Risks

- The brief is only as right as `card orchestrate plan`: a wrong judgement there shows here. The brief calls the plan's function and adds no judgement of its own, so a repair of the plan repairs the brief.
- A list that is cut hides items: `counts` keeps the full totals and `omitted` names what was dropped, so a cut list is never read as a complete one. The oldest items go first because a leader acts on the newest.
- An old unanswered note keeps a card in `blocked` for days (the real ledger holds notes older than five days). The brief shows them as they are, newest first; deciding that they are stale belongs to the leader, not to the command.
- A 4096-byte cap on an answer that holds Korean text counts bytes, not characters; the first line is cut on a character boundary at 96 bytes, so the cut never splits a character.
- The comparison in AC-5 leaves out the lane screens the leader reads today, because reading a screen needs a live host session; the Result says so and does not claim a total.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The command word is `moai leader brief`: a new operation file `src/ops/leader.rs` with the action `brief`. | The card's and the user's dispatch name (2026-11-24). The word `leader` is free (`moai leader` answered `unrecognized subcommand`, 2026-11-24T15:01Z); `leader-pulse`, `leader-round`, `leader-run` and `leader-tick` keep their names. | Rename the word in one file. |
| D-2 | brief shrinks the answer of `card orchestrate plan` and adds no judgement and no query of its own. | Rule 3 of the core rules (the cheapest source: existing code). The plan already holds `land_queue` and `attention` with their note numbers (`src/card/orchestrate.rs`, read 2026-11-24). | Write own queries. |
| D-3 | A `decisions` item is a `LANE-BLOCKED` note that says `kind=decision` or a `DECISION-QUESTION` note; the rest of the unanswered `LANE-BLOCKED` notes are `blocked`. | `skills/moai/workflows/work.md` ("Exceptions and Permissions") fixes the line `LANE-BLOCKED kind=decision …` for a decision. | Treat every block as a decision. |
| D-4 | The cap is 4096 bytes for the whole printed envelope; the cut drops the last item of the longest list, newest-first lists keep the newest. | The card's criterion 1 (4KB). A leader acts on the newest note; `counts` and `omitted` keep the rest visible. | Another cap or another cut order. |
| D-5 | The answer holds no clock value, so equal ledgers print equal bytes. | The plan's own rule (`REQ-9` of SPEC-20261124-orchestrate-plan). | Add ages. |
| D-6 | Card criterion 2 is not in this slice. | The dispatch of 2026-11-24: C-470 is under repair and `card land-batch` is not landed; criterion 2 is built on it. | Do it in this slice on a base that holds C-470. |

## Result

Status: draft (approval is the leader's, at READY); slice 1 done: `moai leader brief`, read only. Card criterion 2 is not done and not claimed (D-6, Plan 3).

Measured 2026-11-24 (UTC) on macOS in the worktree `.claude/worktrees/leader-brief`, branch `spec/leader-brief-20261124`, base `2e6873482e`. Code: `src/leader_brief.rs`, `src/ops/leader.rs`, the module line in `src/lib.rs`, the `leader` rows in `src/help.rs` and the four golden help files, and the operation document `docs/design-review/appendix/operations/leader.md`. RED first (`277d07bff9`): of 7 tests of `tests/leader_brief_cli.rs` 6 failed on `unrecognized subcommand 'leader'` (the seventh, `measure_`, passed on the error text and was made to check the exit status first). After the code (`e8bc758ea9`) 7 of 7 and the 3 unit tests of `src/leader_brief.rs` passed; the golden help files followed in `e399138173` (`MOAI_HELP_GOLDEN_WRITE=1`, `help_layout` 11 of 11 passed), and `test_area_layout`, `help_text_names_moai` and `leader_brief_cli` together ran 28 tests, 28 passed.

- AC-1 met (REQ-1, REQ-3, REQ-5, REQ-6, REQ-8): `shape_prints_the_parts_with_numbers_and_writes_nothing` (two audited heads, one by event and one by note, `ref` `event:<id>` and `note:<id>`; one block and one decision with their note numbers; the home folder equal before and after; two runs equal byte for byte) and `shape_a_project_without_a_ledger_is_empty_and_exits_zero`.
- AC-2 met (REQ-2): `size_the_whole_answer_stays_within_the_cap_and_keeps_the_newest` (60 blocked notes of long Korean text and 40 audited heads: the whole stdout is at most 4096 bytes, `counts` keeps 60 and 40, shown plus `omitted` equals `counts` for each list, the blocked list starts at the newest card `C-160` and its note numbers fall).
- AC-3 met (REQ-4): `split_a_decision_is_told_from_a_block_and_answered_or_done_notes_vanish` (`kind=decision` and `DECISION-QUESTION` are decisions; `kind=claim` and the look-alike `kind=decisions` are blocks; an answered note, a note of a `done` card and a `LANE-NOTE` line are listed nowhere).
- AC-4 met (REQ-7): `unread_an_unreadable_integration_branch_is_listed_and_the_rest_still_prints` and `unread_an_unreadable_lane_registry_is_listed` (a registry folder that is a file).
- AC-5 met (REQ-1, card criterion 3): `measure_the_brief_is_smaller_than_the_two_reads_it_replaces` (on the crowded ledger the brief is below a quarter of `card inbox` plus `card orchestrate plan`), and the real-ledger numbers below.
- AC-6 not_run here: the neighbour checks are the `check affected` receipt of the READY note.

Card criterion 3, read-only on the real ledger of this project, 2026-11-24T15:32Z, one wake, the same built binary for every line (`target/debug/moai`, project `.`):

| What a leader reads in one wake | Before | After |
| --- | --- | --- |
| `moai card inbox` | 7658 bytes | not read |
| `moai card orchestrate plan` | 17326 bytes | not read |
| the 11 notes those two list (5 blocks and decisions, 6 verdicts), read whole with `moai card show` | 14561 bytes | only those the leader opens |
| `moai leader brief` | not there | 1515 bytes (queue 6, blocked 4, decisions 1, nothing cut, `unread` empty) |
| total | 39545 bytes | 1515 bytes plus the bodies the leader chooses to open |

The brief is 26 times smaller than the three reads together and 16.5 times smaller than the two listings alone (24984 bytes). It does not remove the reading of a body the leader needs; it makes that reading a choice made from a one-line item with its note number. Not measured: the lane screens a leader reads today (they need live Orca sessions), so the real total before is larger than the 39545 bytes shown. The five blocked and decision items of the brief are as many as the five `blocked` items of `card inbox` of the same minute (counted by family, not compared one by one); the ten `ready_for_integration` items of the inbox are not part of the brief, and the brief lists the six audited heads of the landing queue instead, three of them marked `old` (a newer head of the same card was handed back).

- Run: `moai heavy-slot-acquire --name leader-brief --max 3 -- cargo nextest run --offline --locked --no-fail-fast --test leader_brief_cli` (with `RUSTC_WRAPPER=sccache`): 7 tests, 7 passed; the unit tests `leader_brief::` 3 of 3.
- Guard mutants (each applied once to `e399138173`, checked as applied, run on the one named test by `target/leader-brief/mutants.py` in one slot, restored; each kill is a FAIL of that test and no build error; the tree was clean after the last): A the cut loop never runs (`while false`), exit 100, `size_the_whole_answer_stays_within_the_cap_and_keeps_the_newest`; B `kind=decision` no longer makes a decision, exit 100, `split_a_decision_is_told_from_a_block_and_answered_or_done_notes_vanish`; C the blocked list reads oldest first, exit 100, `size_the_whole_answer_stays_within_the_cap_and_keeps_the_newest`.
- Deviation kept small: the answer holds `audit` in `counts` (a count of heads that wait for an audit), which the card did not ask for; it is one number and not a list.
- Known limit: a note's first line reaches the brief cut at 200 characters by `card inbox`'s own limit, so a `kind=decision` that stands after the 200th character reads as a block.
- Not done and not claimed: card criterion 2 and everything of Plan 3; the skill text that calls the brief; a measurement of the lane screens.
