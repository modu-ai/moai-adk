---
id: SPEC-20261124-card-select-batch-land
title: "READY checks only what the card touched; a batch of audited heads lands on one wide run"
status: draft
revision: 1
created: 2026-11-24
---

# READY checks only what the card touched; a batch of audited heads lands on one wide run

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-470 의 본문과 여섯 기준, 리더 결정 노트 8552(사용자 지시 2026-11-24 병목 개선안 ①, "같은 검증을 반복하지 않는다")와 리더 배차가 정했어요.

## Intent

Measured on the receipt of C-464 (2026-11-24): of 25 minutes, 14 were queue for the heavy slot and 10 were tests (3997 tests in 146 bundles); compiling took 2 seconds. One card runs the same wide set three times: in the lane's READY receipt, at landing, and again in the full run every five landings. Even `--lane` still picks 117 bundles for a change under `src/card/`, because the area rule selects every test whose first word is `card`. The aim is not to test less but to stop repeating the same check: the lane's receipt covers what the card touched, and the wide set runs once per batch, on the tree the batch will really land. Claims about CPU saved and the queue shrinking stay hypotheses until a second receipt is measured (decision note 8552).

## Non-goals

- No change to the full set (the bare command) or to the `--lane` set, and no change to the release-level full run.
- No change to the audit verdict rules, the scan receipt rule or the tier floor of `card land`; the batch calls the same landing for each head.
- The sentences of `CLAUDE.md`, `next-flow.md` and `leader-rules.md` that name the new order belong to C-471.
- The wide check does not become a hidden default: a lane asks for the card set with `--card`.

## Requirements

- REQ-1: When `check affected` runs with `--card`, the command shall select the test targets that the range changes or adds, the targets whose name or source names a changed module, the targets of each scan one of whose rule files the range changes, and the tests that `skills/` and `docs/` changes name.
- REQ-2 (REQ-1): IF more than 10 test targets reach a changed module by its name word or by naming its path in their source, THEN the command shall ignore that module and list its word in the answer.
- REQ-3 (REQ-1): IF a changed path is `Cargo.toml`, `Cargo.lock`, a toolchain or `.cargo` file, `build.rs`, `.moai/project.yaml`, a permission or hook guard file, or a ledger storage file (`schema`, `migrate`, `open`, `recover`), THEN the command shall select the full set and name the reason and the paths.
- REQ-4: When `--card` selects the card set, the summary line shall carry `selection=card`, so that the READY gate of `card note` keeps a note with that receipt.
- REQ-5 (REQ-4): IF a receipt carries `selection=lane`, THEN the READY gate shall refuse it as before.
- REQ-6: When `card land-batch` is given 2 to 4 `C-n=<head>` pairs and the check passes on their join, the command shall land each head by one compare-and-swap in the given order, through the same landing as `card land`.
- REQ-7 (REQ-6): IF the check fails on a set of more than one head, THEN the command shall run the first half of the set on its own and find the failing head by halving.
- REQ-8 (REQ-7): IF the check fails on a single head, THEN the command shall drop only that head and answer each head as landed, dropped with its reason, or not run.
- REQ-9 (REQ-6): IF the integration branch moved since a check ran, THEN the command shall run the check again on the join onto the new tip.
- REQ-10 (REQ-6): IF a head cannot be joined without a conflict, THEN the command shall drop it with that reason and go on with the others.
- REQ-11 (REQ-6): IF the check cannot start, THEN the command shall land nothing and say why.
- REQ-12 (REQ-6): IF the landing refuses a head (verdict, scans, stale tip), THEN the command shall stop at that head and say which heads landed and which did not run.

## Acceptance

- AC-1 (REQ-1, REQ-2): for a range that changes `src/ledger/sweep.rs`, `src/card/sweep.rs`, `src/card/work.rs`, `src/ops/card.rs` and adds `tests/card_sweep_cli.rs`, `--card --list` selects at most 20 targets and names the ignored word `card` · verify: `cargo test --locked --offline --test check_affected_card_select card_set_`
- AC-2 (REQ-1): the card set of a range holds a changed test file, a target whose source names the changed module, and the targets of a scan whose rule file changed, and holds no target that none of them reaches · verify: `cargo test --locked --offline --test check_affected_card_select card_set_`
- AC-3 (REQ-3): a range that changes `Cargo.lock`, a ledger `schema.rs` or `migrate.rs`, a hook guard file or `.moai/project.yaml` gets the full set with the reason named; a range that changes only `src/ledger/sweep.rs` does not · verify: `cargo test --locked --offline --test check_affected_card_select wide_`
- AC-4 (REQ-4, REQ-5): a `--card` run writes a summary with `selection=card` and a card-pick PASS passes the READY gate, a `--lane` PASS still does not · verify: `cargo test --locked --offline --test check_affected_card_select receipt_`
- AC-5 (REQ-6): three heads pass one fake check, which runs once, and the three land in the given order on the integration branch · verify: `cargo test --locked --offline --test card_land_batch_cli batch_`
- AC-6 (REQ-7, REQ-8): with three heads where one fails the check, that head is dropped with its reason and the other two land · verify: `cargo test --locked --offline --test card_land_batch_cli split_`
- AC-7 (REQ-9): after the first half lands, the second half runs on the new tip and not on the earlier result · verify: `cargo test --locked --offline --test card_land_batch_cli split_`
- AC-8 (REQ-10, REQ-11, REQ-12): a conflicting head is dropped, a check that cannot start lands nothing, and a head the landing refuses stops the rest · verify: `cargo test --locked --offline --test card_land_batch_cli refuse_`

## Scope

- write: `.moai/specs/SPEC-20261124-card-select-batch-land/**`
- write: `src/check_affected.rs`
- write: `src/card/land.rs`, `src/card/land/**`
- write: `src/card/ready_receipt.rs`
- write: `tests/check_affected_card_select.rs`, `tests/card_land_batch_cli.rs`
- shared (line lease): `src/ops/card.rs` (the `land-batch` action and its dispatch), `src/ops/check.rs` (the `--card` flag), `docs/design-review/appendix/operations/card.md`
- forbidden: `CLAUDE.md`, `skills/**/next-flow.md`, `docs/leader-rules.md` (C-471)

## Plan

1. RED: pin the card selection, the wide fallback and the receipt rule in `tests/check_affected_card_select.rs`; pin the batch landing, the split and the refusals in `tests/card_land_batch_cli.rs` with a fake check program.
2. GREEN, selection: `Pick::Card`, the module-name and wide-class rules, the `selection=card` word, and the receipt gate.
3. GREEN, landing: `card::land::batch` joins the heads with `git merge-tree --write-tree` and `git commit-tree`, checks the joined tree in a private worktree, and lands each join through `land`.
4. Measure the card set of the real C-464 range with `check affected --card --list`, and run both landing cases in a throwaway repository.
5. Guard mutants for each new guard, then the one bare `check affected` at READY.

## Risks

- A card selection that is too narrow lets a break through the lane receipt; the batch's wide run on the joined tree is the net, and the five-landing full run stays.
- A module name that no test names selects nothing for a behaviour reached only through the binary; the targets of the card's own criteria cover that, and a wide-class path falls back to the full set.
- The batch runs a long check inside one command; a failed or interrupted run lands nothing that was not checked, and the answer lists what landed before the stop.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Criteria 1 and 6 of C-470 read together: "ledger" in the wide classes means the storage layer (`schema`, `migrate`, `open`, `recover`) and migration paths; the rest of `src/ledger/` takes the card set, because criterion 1 names a `src/ledger` change (C-464) as a card-set case | lane reading of the two criteria; the leader's note 8552 says wide "when the effect is not clear" | widen `WIDE_CLASSES` in `src/check_affected.rs` to all of `src/ledger/` |
| D-2 | The card set is a new flag `--card`; the bare command and `--lane` keep their sets | the receipt gate and the existing tests rely on them; C-471 changes the documented order | stop passing `--card` |
| D-3 | The batch is a new action `card land-batch`; `card land` is unchanged | one card per `card land` is its rule and its tests | remove the action |
| D-4 | The batch's check is a program (`--check-cmd`, default this binary's `check affected <tip>..<join>`) that runs in the private worktree and exits 0 to pass | a program can be faked in a throwaway repository; the default is the real wide set | drop the option |
| D-5 | A verdict on the audited head covers a join candidate that is a clean join of it (`LandRequest.audited`: second parent, tree equal to the merge of the parents); the verdict, author and tier checks read the audited head and the tier floor reads the files from its merge base with the tip | the batch lands joins that no auditor saw; the card says "PASS head" for the heads it joins, and `authors_of(tip, audited)` is the lane's own range | stop passing `audited` from `land-batch` |
| D-6 | The card set takes a scan only when one of its rule files changed; a scan's scope is what it reads (often the whole tree) and its receipt is the landing's and the batch check's | measured on the real C-464 range: the scopes alone gave 63 scan targets and 89 in all, against the criterion of at most 20 | take `touched_by` again in `select_card` |
| D-7 | The name word of a changed module is `<folder>_<module>` below a folder and the module name for a top-level file; the cap of 10 counts the targets that reach a module by that word or by naming its path in their source, together | measured on the real C-464 range: the bare module word `sweep` or `work` reached the whole `card_*` family, and the path rule alone reached 14 targets for `src/ledger/reservation.rs`; the union cap gives 13 targets | go back to the bare word, or raise the cap |

## Result

Code head `7c8153bae1` on base `6559535dba`. Checks of 2026-11-24 on a Mac, load 21 to 35 from other lanes; every cargo run under `moai heavy-slot-acquire --name lane-470 --max 4`.

- Measured on the real C-464 range (`dbc09e87a59f..98199c2d251f`, `check affected --list`): the bare selection is 146 targets and 82 scans; `--card` is 13 targets and `scans=listed:0`, ignored words `card`, `ledger_reservation`, `profile`. Earlier builds of this lane gave 89 (scan scopes), 36 and 31 before D-6 and D-7. Logs: `target/c464-full.json`, `target/c464-card4.json`.
- The new tests ran with the whole neighbourhood in one nextest call (`--lib` and 10 test binaries, among them `check_affected_card_select`, `card_land_batch_cli`, `state_folder_literal_scan`, `hot_file_ops_merge`): 1991 passed, 0 failed (`target/run3.log`). The call before the last edits (`target/run2.log`) was 2014 of 2016, and the two failures were repaired: the two new test files were missing from `tests/fixtures/state_folder_literals/allow.tsv`, and the conflict test used a file that is tier 1 (a landing refused it for a missing receipt, not for the conflict).
- The verify commands of the AC lines, run as written (`target/verify1.out`): `check_affected_card_select` `card_set_` 2 passed, `wide_` 3 passed, `receipt_` 1 passed; `card_land_batch_cli` `batch_` 3 passed, `split_` 2 passed, `refuse_` 4 passed, each exit 0.
- Guard mutants (one slot call, `target/mutants1.py`, tree clean afterwards), every one killed with exit 100 by the named test: wide `Cargo.lock`, the cap of 10, scans by rule file only, the `selection=card` word, halving, the moved-tip re-read, two to four heads, leader only, the audited second parent, the audited clean join, the verdict read on the audited head. Not mutated: `conflicts_with` of `--card` and `--lane` (clap's own check, pinned by `card_and_lane_cannot_be_asked_together`) and the private-worktree removal (not a guard of its own; the tests list the leftover folders).
- AC-1: met — `check_affected_card_select card_set_` exit 0; the fixture range selects at most 20 and names `card`; the real C-464 range selects 13
- AC-2: met — same command, 2 passed
- AC-3: met — `check_affected_card_select wide_` exit 0, 3 passed
- AC-4: met — `check_affected_card_select receipt_` exit 0, 1 passed; the unit test `a_card_summary_line_is_a_receipt_and_a_lane_one_is_not` of `ready_receipt.rs` passed in the nextest call
- AC-5: met — `card_land_batch_cli batch_` exit 0, 3 passed (one check, landed in the given order)
- AC-6: met — `card_land_batch_cli split_` exit 0, 2 passed (the failing head found by halving, the others landed)
- AC-7: met — same command, the passing first half lands before the rest is checked, and a moved tip is not reused (`batch_a_tip_that_moved_during_the_check_is_not_reused`)
- AC-8: met — `card_land_batch_cli refuse_` exit 0, 4 passed
- Gaps: (1) the batch and the split ran through the real binary in temporary repositories built by the test fixture, not typed by hand in a separate throwaway repository. (2) The batch's default check runs `check affected` in-process on the joined tree and writes no scan receipts of its own. (3) The join commits carry no `Moai-Spec` trailer. (4) D-1 (ledger means the storage layer) and D-5 (a verdict on the audited head covers a clean join) are readings and a policy extension the leader confirms at READY. (5) Criterion 3 of the card (CLAUDE.md and next-flow wording) belongs to C-471 and is not done here. (6) `src/ops/check.rs`, `docs/design-review/appendix/operations/check.md` and `tests/fixtures/state_folder_literals/allow.tsv` are outside the dispatched shared list and were edited under a line lease. (7) Windows and Linux were not run.
