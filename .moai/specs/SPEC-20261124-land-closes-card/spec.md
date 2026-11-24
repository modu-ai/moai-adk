---
id: SPEC-20261124-land-closes-card
title: "감사 PASS 가 기록된 head 를 착지하면 카드가 닫혀요: land 가 PASS 기록으로 열린 기준을 met 으로 만들고, 슬라이스는 --keep-open 일 때만 열린 채 남아요"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8023 (2026-11-24T00:05:26Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8019 에 대한 답): 판 1 승인, D-1~D-6 기본값 그대로예요; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요"
  digest: "sha256:6489768c92a4e73c2717cb1835ba6bf31037efc41499b823746d53460fc27493"
---

# 감사 PASS 가 기록된 head 를 착지하면 카드가 닫혀요: land 가 PASS 기록으로 열린 기준을 met 으로 만들고, 슬라이스는 --keep-open 일 때만 열린 채 남아요

## Intent

착지한 카드가 닫히지 않고 쌓여요. 카드 C-442(사용자 선택, 2026-11-24 카드 로직 전수 조사)가 원장에서 센 숫자: 감사 PASS 노트가 있는 `ready` 카드 54장, landing 노트가 있는 열린 카드 61장(카드 본문의 측정이고, 이 SPEC 은 다시 세지 않았어요). 까닭은 코드에 있어요(2026-11-24 읽음, NeXT `4acca3f4ec`):

- `card land` 는 열린 기준마다 `--met N=REF` 가 하나씩 없으면 "landing 만 기록하는 호출"이 되어 `closed: false` 로 답하고 카드를 그대로 둬요(`src/ledger/workflow.rs` `finish_plan` 의 `landing_only`, 1459-1463 와 2029-2041 줄; 카드 본문은 `src/card/workflow.rs` 라고 적었지만 파일은 `src/ledger/workflow.rs` 예요).
- 기준을 `met` 으로 만드는 `card criterion --met` 는 살아 있는 lease 가 필요해요(`src/ledger/workflow.rs` 582-588 줄, `reservation::fenced`). 리더는 레인의 lease 를 쥐지 않아서, 착지 뒤에 기준을 닫을 길이 `--met` 를 일일이 적는 것뿐이에요.
- 착지는 닫힌 카드의 워크트리와 브랜치만 정리해요(`src/card/land.rs` 367 줄). 열린 채 남은 카드는 정리도 받지 못해요.

`card land` 는 이미 독립 PASS 를 읽어요(SPEC-20261123-landing-reads-verdict: `card.verdict` 사건을 `src/card/land/check.rs` 가 읽고, 착지가 기대는 PASS 를 `relied` 로 이름 대요). 그러니 PASS 가 이 head 를 가리킨다는 사실은 이미 도구가 알고 있어요. 이 SPEC 은 그 사실을 닫는 데 써요:

- 독립 PASS 가 이 head 에 있으면 `card land` 가 열린 기준을 그 PASS 사건을 증거로 `met` 으로 만들고 카드를 닫아요.
- PASS 가 없으면 `--keep-open` 이 없는 한 거절해요.
- 슬라이스처럼 열린 채 남겨야 하는 카드는 `--keep-open` 으로 착지해요. 답이 열린 기준을 이름 대고, `card next` 는 그 카드를 계속 보여 줘요.

읽는 쪽은 착지를 돌리는 리더, 레인, 이 변경을 감사하는 세션이에요.

### 읽은 것 (2026-11-24, 팁 `4acca3f4ec`)

- 카드 C-442 본문과 기준 셋(`moai card show C-442`), 노트 없음, HOLD 표시 없음.
- `src/card/land.rs`: 호출 순서(카드 → git → SPEC → `finish_preflight` → 판정 읽기 → 시작 행 → 이동 → `finish_naming`), `landing_only` 의 `applied` 가 비는 것.
- `src/card/land/check.rs`: `read` 가 `relied`(없으면 observe 에서 `would_refuse`)를 돌려주는 것, 이미 가지에 든 head 는 시작 행이 이름 댄 판정으로 닫는 것.
- `src/ledger/workflow.rs`: `finish_plan`·`finish_naming`·`record_land_started`·`landing_facts`, `card.landed` payload.
- `src/card/landed.rs` `find`·`take_landed`: 기록된 landed head 가 가지 끝 안에 있으면 `ready` 카드를 `already_landed` 로 `skipped` 에 옮기는 것.
- `src/ops/card.rs` 의 `Land` 정의와 호출부, `docs/design-review/appendix/operations/card.md` 의 `card land`·`card next` 표 줄.

## Interview

- Q1: Should `no_pass` follow the observe mode of `work_flow.land_verdict`? -> A: No (D-1). The card's rule is a refusal without a PASS unless `--keep-open`; the mode keeps governing a `--keep-open` landing. Default accepted in leader note 8023.
- Q2: Is the evidence of a met criterion the AUDIT note or the verdict event? -> A: The `card.verdict` event (D-2), because no note text is read as a verdict (SPEC-20261123-landing-reads-verdict REQ-19). Accepted in note 8023.
- Q3: Where does the slice mark live? -> A: The `keep_open` key of the `card.landed` payload, no schema change (D-5). Accepted in note 8023.

## Non-goals

- `card finish` 의 변경: 그쪽의 "landing 만 기록하는 호출"과 `--met` 규칙은 그대로예요. 이 SPEC 은 `card land` 만 바꿔요.
- 독립 PASS 를 읽는 규칙(`check.rs` 의 `no_verdict`·`verdict_fail`·`self_verdict`·`lander_review`·`receipt_missing` 등)의 변경. 이 SPEC 은 `relied` 를 그대로 믿어요.
- `work_flow.land_verdict` 의 기본값을 `enforce` 로 바꾸는 일. 모드 전환은 따로 기록된 결정이에요.
- 노트 글을 판정으로 읽는 일: `AUDIT … verdict: PASS` 노트는 읽지 않고, 증거는 `card verdict` 가 쓴 `card.verdict` 사건이에요(SPEC-20261123-landing-reads-verdict REQ-19 그대로).
- `card criterion` 의 lease 규칙 변경과 `card release`·`card reopen` 의 변경.
- 이미 쌓인 54장·61장을 한꺼번에 닫는 일괄 작업. 같은 PASS 가 있으면 카드마다 `card land` 를 다시 부르면 닫혀요(REQ-9).
- 스킬·에이전트·리더 규칙 글의 수정(`skills/**`, `docs/leader-rules.md`): 코드가 들어간 뒤 글 조각이 따로 해요.

## Requirements

<!-- This SPEC adds a refusal that blocks (`no_pass`) at the landing. It does not use the observe mode of land_verdict, because the card asks for the refusal as the rule (D-1); the existing check in observe mode stays as it is for a --keep-open landing (REQ-6). -->
- REQ-1: When the card land command lands a head that the verdict check relies on a PASS for and no --keep-open is given, the card land command shall mark every open criterion of the card met with the verdict event as its evidence and close the card · failure: none — a refusal of the close is one of the finish checks and is checked by AC-1 and AC-10.
- REQ-2: When the card land command is given a --met option for an open criterion and relies on a PASS, the card land command shall keep the given reference for that criterion and use the verdict event for the other open criteria · failure: none — a --met that names a criterion that is not open stays refused as before.
- REQ-3: The card land command shall write the evidence of a criterion it marks met as one line that names the verdict event number and the verdict word PASS and the first twelve hex digits of the head · failure: none — the line is made from values the check already holds.
- REQ-4 (REQ-1): IF the card land command has no relied PASS for the head and no --keep-open is given, THEN the card land command shall refuse with the code `no_pass` and write nothing, whatever mode work_flow.land_verdict names.
- REQ-5: When --keep-open is given, the card land command shall record the landing and close nothing and answer closed false with the numbers of every open criterion · failure: none — the landing-only answer is the one the finish command already gives.
- REQ-6 (REQ-5): IF --keep-open is given and work_flow.land_verdict is enforce and the verdict check refuses the head, THEN the card land command shall refuse with that check's code and write nothing.
- REQ-7 (REQ-5): IF --keep-open is given together with a --met option or the card has no open criterion, THEN the card land command shall refuse as an invalid call and write nothing.
- REQ-8: When the card land command records a --keep-open landing, the card land command shall write keep_open true into the payload of the card.landed event · failure: none — a head that was recorded before writes no second event, and the first event stays as it was written.
- REQ-9: When the card land command is run again for a head that the integration branch already holds and whose start row names a verdict event, the card land command shall close the card on that verdict event and not read the verdicts again · failure: none — this is the rerun after a stopped landing of SPEC-20261123-landing-reads-verdict REQ-15, and a card whose head the branch holds by another route is judged by the verdict events as before.
- REQ-10: When the card next command reads a ready card whose newest landed head carries keep_open true, the card next command shall list the card in ready and not skip it as already_landed · failure: none — a card with no keep_open mark is skipped as before.
- REQ-11: When the card land command closes a card, the card land command shall give the answer the same keys as before and add the met list with the verdict evidence of every criterion it marked · failure: none — the keys `met`, `closed` and `open_criteria` already exist in the answer.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-11): a card with three open criteria and a head with an independent PASS on it lands in one call: the card is done, the three criteria are met with the line that names the event, PASS and the twelve hex digits, and the answer's `met` lists the three · verify: `cargo test --locked --offline --test card_land_cli ac_closes_`
- AC-2 (REQ-2): a call with `--met 2=<ref>` keeps that reference for criterion 2 and gives the verdict line to criteria 1 and 3 · verify: `cargo test --locked --offline --test card_land_cli ac_keeps_met_`
- AC-3 (REQ-4): a head with no verdict, a head with a PASS for another head and a head whose newest verdict is FAIL each refuse without --keep-open: the first two as `no_pass`, the FAIL with the check's own `verdict_fail`; the branch and the ledger are unchanged, also in observe mode, and a `--met` option for every criterion does not bypass it · verify: `cargo test --locked --offline --test card_land_cli ac_no_pass_`
- AC-4 (REQ-5, REQ-8): `--keep-open` on a head with no verdict lands in observe mode: the branch moves, one `card.landed` event carries `keep_open` true, the card keeps its state and its criteria stay open, and the answer names them · verify: `cargo test --locked --offline --test card_land_cli ac_keep_open_`
- AC-5 (REQ-6): the same call with the profile on `enforce` refuses with the check's code and writes nothing, and the same call with an independent PASS lands · verify: `cargo test --locked --offline --test card_land_cli ac_keep_enforce_`
- AC-6 (REQ-7): `--keep-open` with `--met` and `--keep-open` on a card with every criterion met each refuse and write nothing · verify: `cargo test --locked --offline --test card_land_cli ac_keep_invalid_`
- AC-7 (REQ-10): a slice landed with `--keep-open` and then released to `ready` is listed in `ready` of `card next`, and a card landed by `card finish` in landing-only form is still skipped as `already_landed` · verify: `cargo test --locked --offline --test card_finish_cli ac_keep_open_next_`
- AC-8 (REQ-8): the slice of AC-4 landed again with the same head and an independent PASS and no `--keep-open` closes the card with one `card.landed` event in total and the criteria met · verify: `cargo test --locked --offline --test card_land_cli ac_slice_then_close_`
- AC-9 (REQ-9): a landing stopped after the branch moved and before the close, run again with no `--keep-open`, closes on the verdict event the start row names, also after a later FAIL was recorded · verify: `cargo test --locked --offline --test card_land_cli ac_rerun_closes_`
- AC-10 (REQ-1): a card whose newest reservation another session holds with a live lease keeps today's refusal (the finish check), and a `review` card of the holder is closed as today · verify: `cargo test --locked --offline --test card_land_cli ac_held_`
- AC-11 (REQ-1, REQ-4): every test that landed a card before this SPEC passes after the fixture of that test records the verdict it needs or names `--keep-open`; none passes by a weaker assertion · verify: `cargo test --locked --offline --test card_land_cli --test card_land_scans_cli --test card_remote_copy_cli --test card_cleanup_outbox_cli --test branch_gc_cli --test e2e_temp_home_cli --test mode_cards_53 --test leader_route`
- AC-12 (REQ-4): the `no_pass` refusal comes before the start row and before the intent record: a refused call leaves no `card.land_started` event and no file under the common git folder's `moai/` · verify: `cargo test --locked --offline --test card_land_cli ac_no_pass_writes_nothing_`

## Scope

- write (this slice): `.moai/specs/SPEC-20261124-land-closes-card/**`
- write (code slice, after approval):
  - `src/card/land.rs` (the effective met list, `no_pass`, `--keep-open`, the answer);
  - `src/ledger/workflow.rs` (the `keep_open` mark in the `card.landed` payload, `LandingFacts.keep_open`);
  - `src/card/landed.rs` (`find` gives no head proof to a card whose newest landed head carries `keep_open`);
  - `src/ops/card.rs` (the `--keep-open` option of `Land` and its pass-through, the doc text of `Land`);
  - `tests/card_land_cli.rs`, `tests/card_finish_cli.rs`, and the fixtures of the other test files that land a card: `tests/card_land_scans_cli.rs`, `tests/card_remote_copy_cli.rs`, `tests/card_cleanup_outbox_cli.rs`, `tests/branch_gc_cli.rs`, `tests/e2e_temp_home_cli.rs`, `tests/mode_cards_53.rs`, `tests/leader_route.rs`.
- shared files by line lease, never in the scope: `docs/design-review/appendix/operations/card.md` (the `card land` and `card next` rows and one line for `--keep-open`), `src/help.rs` if its `card land` example needs the option, `src/lib.rs` (none expected), `tests/areas.toml` (none expected).
- read-only: `src/card/land/check.rs` (`relied` is used as it is; its signature changes only if the code slice proves it must), `src/card/finish.rs`, `src/ledger/reservation.rs`.
- forbidden: the pull request gates, `src/ledger/schema.rs` and `src/ledger/migrate.rs` (no new column: the mark is a key of the event payload), any skill or agent text, `CLAUDE.md`.

## Plan

1. The SPEC (landed as draft). Approval is the leader's note on the card, and no approval record is written by the lane.
2. After approval, the scope request for the paths of the code slice, then tests first (RED): the new cases of `tests/card_land_cli.rs` and `tests/card_finish_cli.rs`, run once to see them fail for the right reason.
3. `src/ledger/workflow.rs`: the `keep_open` key in the `card.landed` payload (one new argument path through `finish_naming`) and `LandingFacts.keep_open`.
4. `src/card/land.rs`: after the verdict is read, build the effective met list from the open criteria and `relied`, run `finish_preflight` again on it, refuse `no_pass`, handle `--keep-open`; `src/ops/card.rs` takes the option.
5. `src/card/landed.rs`: `find` skips the head proof when the newest landing is `keep_open`.
6. The fixtures of the other test files that land a card (AC-11), then the targeted tests, one `check affected` unwrapped, guard-mutant lines, READY with head, base and the receipt.
7. Text slice after the code lands: the `card land` row and the landing text of `skills/moai/references/flows/next-flow.md` and `docs/leader-rules.md` ("land closes the card on a recorded PASS; a slice lands with `--keep-open`"). A separate card.

## Risks

- R-1: `no_pass` is not tied to the observe mode, so every landing of a head with no `card verdict` refuses at once; the old `--met`-only landing is gone. The leader records the verdict first (the audit already does) or lands with `--keep-open`. This is the card's rule (D-1), and the fixture churn of AC-11 is its cost.
- R-2: `--keep-open` in observe mode lands a head that has no PASS, as the observe mode always allowed. It is a way to land without a verdict that the old `--met`-less call also had; it closes nothing, so the card stays visibly open.
- R-3: The evidence of a closed criterion is the verdict event, not the audit note the card body names. The two are the same audit, but only the event is read by tools (D-2).
- R-4: A card another session holds with a live lease in the `active` state is still refused by the finish check (REQ-1 does not change `finish_gate`); only a `review` card of the holder or an ended lease passes. If the ledger shows many such cards, a later card decides the lease question.
- R-5: Two writes cannot share one transaction (git and the ledger); a close refused after the move leaves the branch moved and says so, as in SPEC-20261121-card-land, and REQ-9 makes the rerun close it.
- R-6: `keep_open` is a payload key, so an older binary that reads the event ignores it and skips the card as `already_landed`; only a newer binary shows the slice.
- Dangerous path (CLAUDE.md): the landing moves a protected ref and closes a card, so the READY note carries a real run in a throwaway project: a slice landed with `--keep-open`, the same head closed later, and a refused `no_pass` call that leaves the branch unmoved.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A landing that closes needs a relied PASS in every mode; without it the call refuses `no_pass` unless `--keep-open`. | The card's change text (user pick 2026-11-24): "Without a PASS on that head, land refuses unless --keep-open". The mode `observe` of SPEC-20261123-landing-reads-verdict keeps governing a `--keep-open` landing. | Make `no_pass` follow the mode: report it as `would_refuse` in observe. |
| D-2 | The evidence of a met criterion is the `card.verdict` event the check relied on (event number, PASS, head prefix), not a note ref. | SPEC-20261123-landing-reads-verdict REQ-19: no note text is read as a verdict; an event has a number and a reviewer a tool can check. | Read the AUDIT note ref instead, which needs a parser the earlier SPEC refused. |
| D-3 | The close uses the finish transaction's own met list; `card criterion` and its lease rule are not touched. | `src/ledger/workflow.rs` `finish_naming` marks criteria met inside the one write transaction after the finish checks, so no lease is needed beyond `finish_gate`. | Add a leader-only met path to `card criterion` in another SPEC. |
| D-4 | `--keep-open` is an option of `card land` only; `card finish` keeps its implicit landing-only form. | The card names `card land`; the implicit form of `card finish` is how the 61 open cards came about but it is also the form the spec-less close-out uses. | Add the option to `card finish`. |
| D-5 | The slice mark is the payload key `keep_open` of `card.landed`; `next` reads it, and no schema change is made. | `src/ledger/workflow.rs` notes: events have free payloads and the `done` trigger reads only `card.landed`. | A column or a card state `slice` in another SPEC. |
| D-6 | An explicit `--met` for an open criterion wins over the verdict line for that criterion. | The lander may know a better reference; the verdict line only fills the rest, and REQ-1 never leaves a criterion open. | Ignore `--met` when a PASS is relied on. |

## Result

Measured 2026-11-24 (UTC) on macOS in the worktree `.claude/worktrees/land-closes-card`, code head `9d7eec15cbabb928e097163c4c151c47b36a1121`, base `4acca3f4ece0e438aadc107659f54f627e8bf6e6` (the `NeXT` tip when the worktree was made; the checks did not follow the tip). Commits: draft `6101db6494`, approval `a3f4dbba25`, Interview `9d8f9e5431`, RED `4afd38a7a6`, code and fixtures `9d7eec15cb`. The approval header copies the leader's decision note 8023; the lane recorded no approval of its own. The independent audit of the code head (note 8075, verdict FAIL for documents only) found no code defect: AC-1 to AC-12 met by tests and by throw-away dangerous-path runs, mutants M1 to M3 exit 100, `card finish` unchanged (46 of 46), trailer findings 0. Its blockers were this section (B1) and the `card land` row of `docs/design-review/appendix/operations/card.md` (B2); the repair commit after the code head adds this section, the row and the comment of N3, and changes no code.

- AC-1 met: `ac_closes_a_head_with_a_pass_marks_every_open_criterion_met` (`card_land_cli`) lands a three-criterion card on an independent PASS and checks the card, the three criteria and the answer's `met` as the AC says.
- AC-2 met: `ac_keeps_met_a_given_reference_and_fills_the_others_from_the_verdict` keeps the given reference for criterion 2 and fills the others from the verdict line.
- AC-3 met: `ac_no_pass_refuses_a_closing_landing_in_observe_mode` covers a head with no verdict, a PASS of another head and a newest FAIL, with the branch and the ledger unchanged.
- AC-4 met: `ac_keep_open_lands_a_slice_and_leaves_every_criterion_open`.
- AC-5 met: `ac_keep_enforce_applies_the_verdict_check_to_a_slice`.
- AC-6 met: `ac_keep_invalid_refuses_a_met_option_and_a_card_with_nothing_open`.
- AC-7 met: `ac_keep_open_next_lists_a_marked_slice_and_still_skips_an_unmarked_landing` (`card_finish_cli`). The test inserts the marked `card.landed` event itself; the writing side is AC-4 (audit note 8075, N4: the end-to-end path was shown by its dangerous-path runs, so the direct insert stays).
- AC-8 met: `ac_slice_then_close_lands_once_and_closes_on_the_later_pass`.
- AC-9 met: `ac_rerun_closes_on_the_verdict_the_start_row_names`.
- AC-10 met: `ac_held_keeps_the_refusal_for_a_live_lease_of_another_session` (the refusal kind is `conflict`).
- AC-11 met: `moai check affected` ran the whole suite on the code head, 3793 of 3793 passed (`e2e_temp_home_cli` among them); after the code the land-related files also ran alone: `card_land_cli` 70, `card_finish_cli` 46, `card_land_scans_cli` 9, `card_remote_copy_cli` 15, `card_cleanup_outbox_cli` 33, all passed. The old fixtures record an independent PASS first or name `--keep-open`. Five old tests changed meaning (`ac8_`, `ac36_a_card_that_stays_open_`, `ac52_`, `ac56_`, `ac59_`) and now check the same behaviour through `--keep-open` and the `no_pass` refusal.
- AC-12 met: `ac_no_pass_writes_nothing_before_the_start_row_or_the_intent_record`.
- Receipt of the code head: `check-affected base=4acca3f4ece0 head=9d7eec15cbab targets=133 scans=listed:79 test-threads=5 diff_check=0 fmt=0 ops_rustfmt=0 clippy=0 test=0 result=PASS`, run once unwrapped with `CLAUDE_CODE_SESSION_ID` unset, log folder `target/check-affected-9d7eec15cbab/`. The receipt of the repair head is in the REPAIRED note of the card.
- Guard mutants, each applied on the committed code head, run as one test filter and restored with `git checkout`: the `no_pass` return changed to `Ok` fails both `ac_no_pass_` tests (exit 101); the two `--keep-open` input checks switched off (`false &&`) fail `ac_keep_invalid_` (exit 101); the explicit-`--met` condition forced true fails `ac_keeps_met_` (exit 101); the `keep_open` filter removed from `landed::find` fails `ac_keep_open_next_` (exit 101).
- Dangerous path: the tests run the landing against real temporary git repositories and temporary homes: a slice landed with `--keep-open` and closed later (AC-8), a refused `no_pass` call that leaves the branch, the ledger and the intent record untouched (AC-3, AC-12), and a stopped landing closed by a rerun (AC-9).
- Repair, done after the audit: the `card land` row of `docs/design-review/appendix/operations/card.md` now names `--keep-open`, the `no_pass` refusal and the close on the PASS, and the `card next` row names the `keep_open` exception to `already_landed` (both through a line lease); the comment above `ac36_a_card_that_stays_open_is_not_cleaned` says the card stays open because of `--keep-open` (audit N3).
- Gap: the skill and leader-rule texts that tell lanes "land closes the card, a slice lands with `--keep-open`" are the separate text slice of Plan step 7.
- Gap: `src/ops/card.rs` was given back to the line-lease rule after its 14-line change was committed; any later edit goes through a line lease.
- Gap: nothing was run on Windows or Linux; the fixtures are macOS only.
- Gap: `keep_open` is a payload key, so a binary older than this change skips a slice as `already_landed` (R-6).
- Gap: `moai bundle check` reports the floor as tier 1 and the findings `not_on_origin` and `audit_missing`, both expected before the audit and a release push; the `result_missing` finding of the code head is closed by this section.
