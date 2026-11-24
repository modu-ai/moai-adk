---
id: SPEC-20261124-card-select-batch-land
title: "READY checks only what the card touched; a batch of audited heads lands on one wide run"
status: draft
revision: 3
created: 2026-11-24
branch: spec/card-select-batch-land-followup-20261125
---

# READY checks only what the card touched; a batch of audited heads lands on one wide run

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-470 의 본문과 여섯 기준, 리더 결정 노트 8552(사용자 지시 2026-11-24 병목 개선안 ①, "같은 검증을 반복하지 않는다")와 리더 배차가 정했어요.

## Intent

Measured on the receipt of C-464 (2026-11-24): of 25 minutes, 14 were queue for the heavy slot and 10 were tests (3997 tests in 146 bundles); compiling took 2 seconds. One card runs the same wide set three times: in the lane's READY receipt, at landing, and again in the full run every five landings. Even `--lane` still picks 117 bundles for a change under `src/card/`, because the area rule selects every test whose first word is `card`. The aim is not to test less but to stop repeating the same check: the lane's receipt covers what the card touched, and the wide set runs once per batch, on the tree the batch will really land. Claims about CPU saved and the queue shrinking stay hypotheses until a second receipt is measured (decision note 8552).

## Non-goals

- No change to the full set (the bare command) or to the `--lane` set, and no change to the release-level full run.
- No change to the audit verdict rules, the scan receipt rule or the tier floor of `card land`; the batch calls the same landing for each head.
- The sentences of `leader-rules.md`, `lane-rules.md` and `audit.md` that name the new order belong to C-471; the `Verify` section of `CLAUDE.md` and `next-flow.md` are written here (revision 3, leader dispatch 2026-11-25).
- `card land` (one card) keeps its `--by leader` word: the session check of revision 3 is the batch's only.
- The wide check does not become a hidden default: a lane asks for the card set with `--card`.

## Requirements

- REQ-1: When `check affected` runs with `--card`, the command shall select the test targets that the range changes or adds, the targets whose name or source names a changed module, the targets of each scan one of whose rule files the range changes, and the tests that `skills/` and `docs/` changes name.
- REQ-2 (REQ-1): IF more than 10 test targets reach a changed module by its name word or by naming its path in their source, THEN the command shall ignore that module's word, list it in the answer and select the full set with that reason.
- REQ-3 (REQ-1): IF a changed path is `Cargo.toml`, `Cargo.lock`, a toolchain or `.cargo` file, `build.rs`, `.moai/project.yaml`, a permission or hook guard file, a ledger storage file (`schema`, `migrate`, `open`, `recover`), a path under the audit floor (`gate::TIER2_PREFIXES`), `install/git-hooks/`, `src/git_hooks/` or `src/card/land/`, THEN the command shall select the full set and name the reason and the paths.
- REQ-13 (REQ-1): IF a changed `src/**/*.rs` path reaches no integration test, THEN the command shall select the full set and name that reason and path, so that no path picks zero tests.
- REQ-4: When `--card` selects the card set, the summary line shall carry `selection=card`, so that the READY gate of `card note` keeps a note with that receipt.
- REQ-5 (REQ-4): IF a receipt carries `selection=lane`, THEN the READY gate shall refuse it as before.
- REQ-6: When `card land-batch` is given 2 to 4 `C-n=<head>` pairs and the check passes on their join, the command shall land each head by one compare-and-swap in the given order, through the same landing as `card land`.
- REQ-7 (REQ-6): IF the check fails on a set of more than one head, THEN the command shall run the first half of the set on its own and find the failing head by halving.
- REQ-8 (REQ-7): IF the check fails on a single head, THEN the command shall drop only that head and answer each head as landed, dropped with its reason, or not run.
- REQ-9 (REQ-6): IF the integration branch moved since a check ran, THEN the command shall run the check again on the join onto the new tip.
- REQ-10 (REQ-6): IF a head cannot be joined without a conflict, THEN the command shall drop it with that reason and go on with the others.
- REQ-11 (REQ-6): IF the check cannot start, THEN the command shall land nothing and say why.
- REQ-12 (REQ-6): IF the landing refuses a head (verdict, scans, stale tip), THEN the command shall stop at that head and say which heads landed and which did not run.
- REQ-14 (REQ-12): When a batch is about to land its first head, the command shall ask the landing for every head's refusals that are read only (verdict, scans, SPEC, tier floor, first parent).
- REQ-17 (REQ-14): IF one head would be refused, THEN the command shall drop that head with the reason, join the rest again and check them again, so that the integration branch moves only along the join that was checked.
- REQ-15 (REQ-6): IF the first parent of a join candidate is not the integration tip that the landing read (for a later head of a batch, the commit of the head before it), THEN the landing shall refuse it.
- REQ-16 (REQ-6): The command shall name each join commit for the SPEC of its card (`merge(next-flow): <SPEC-id> into <branch>, head <short>`) with a `Moai-Spec: <SPEC-id>` trailer and no card number.
- REQ-18 (REQ-16): IF a card names no SPEC, THEN the command shall drop that head with that reason.
- REQ-19 (REQ-6): When the wide `check affected` of a batch passes on a join and leaves a summary line of that join with `result=PASS` and no `selection=` word, the command shall count the scans that the merge tree needs as covered for each head of that run.
- REQ-20 (REQ-19): IF the check of a batch is another program or leaves no such summary line, THEN the command shall count no scan as covered, so that a landing in enforce mode asks the verdict receipt for it.
- REQ-21 (REQ-6): IF the session that calls `card land-batch` is not the session of the live registry entry named `leader` in mode `leader`, THEN the command shall refuse and change nothing.
- REQ-22 (REQ-6): The command shall take the check program only from the environment variable `MOAI_TEST_LAND_BATCH_CHECK` and from no command-line flag.
- REQ-23 (REQ-9): IF the integration branch moves 3 times after a check passed and before its heads landed, THEN the command shall stop with the code `tip_kept_moving` and list the heads that did not run.
- REQ-24: The `Verify` section of `CLAUDE.md` and `next-flow.md` shall say that a lane's READY receipt is `check affected --card`, that the wide set runs once per batch in `card land-batch`, and that the full run every five landings stays.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a range that changes only modules that are no hub (`src/ledger/sweep.rs`, `src/card/sweep.rs`, `src/card/work.rs`, documents) and adds `tests/card_sweep_cli.rs` gets a `--card --list` set of at most 20 targets and names no ignored word; the hub and shared paths of C-464 (`src/ledger/mod.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs`, `src/profile.rs`) run the full set and name each path with its reason · verify: `cargo test --locked --offline --test check_affected_card_select card_set_` and `cargo test --locked --offline --test check_affected_card_select wide_the_hub_paths_of_c464`
- AC-2 (REQ-1): the card set of a range holds a changed test file, a target whose source names the changed module, and the targets of a scan whose rule file changed, and holds no target that none of them reaches · verify: `cargo test --locked --offline --test check_affected_card_select card_set_`
- AC-3 (REQ-2, REQ-3, REQ-13): a range that changes `Cargo.lock`, a ledger `schema.rs` or `migrate.rs`, a hook guard file, `.moai/project.yaml`, `src/gate.rs`, `install/git-hooks/pre-push`, `src/git_hooks/mod.rs`, `tools/local_check.sh`, `REVIEW.md` or `src/card/land/check.rs` gets the full set with the reason named; so does a range whose word is shared by more than 10 targets (`src/ops/card.rs`) or whose `src` path reaches no test (`src/ledger/mod.rs`, `src/lonely.rs`); a range that changes only `src/ledger/sweep.rs` does not · verify: `cargo test --locked --offline --test check_affected_card_select wide_`
- AC-4 (REQ-4, REQ-5): a `--card` run writes a summary with `selection=card` and a card-pick PASS passes the READY gate, a `--lane` PASS still does not · verify: `cargo test --locked --offline --test check_affected_card_select receipt_`
- AC-5 (REQ-6): three heads pass one fake check, which runs once, and the three land in the given order on the integration branch · verify: `cargo test --locked --offline --test card_land_batch_cli batch_`
- AC-6 (REQ-7, REQ-8): with three heads where one fails the check, that head is dropped with its reason and the other two land · verify: `cargo test --locked --offline --test card_land_batch_cli split_`
- AC-7 (REQ-9): after the first half lands, the second half runs on the new tip and not on the earlier result · verify: `cargo test --locked --offline --test card_land_batch_cli split_`
- AC-8 (REQ-10, REQ-11, REQ-14, REQ-17): a conflicting head is dropped, a check that cannot start lands nothing, and a head with no verdict is dropped before any head lands while the rest is checked again and lands on that join · verify: `cargo test --locked --offline --test card_land_batch_cli refuse_`
- AC-9 (REQ-6, REQ-15): a join whose second parent is not the audited head, and a join whose first parent is beyond the tip, are each refused and only the head before it lands · verify: `cargo test --locked --offline --test card_land_batch_cli refuse_a_join`
- AC-10 (REQ-16, REQ-18): the join commits of a batch carry the SPEC in the subject and in `Moai-Spec`, and no card number · verify: `cargo test --locked --offline --test card_land_batch_cli a_join_commit`
- AC-11 (REQ-19, REQ-20): with `work_flow.land_verdict` set to `enforce` and a scan table that the heads need, a batch whose check left a full-selection PASS summary of the join lands every head, and a batch whose check left none drops the first head with `scans_missing` · verify: `cargo test --locked --offline --test card_land_batch_cli scans_`
- AC-12 (REQ-21): a batch called from a session that is not the registered leader session, with no leader entry, or with an ended lease, is refused and changes nothing; the registered leader session lands · verify: `cargo test --locked --offline --test card_land_batch_cli leader_`
- AC-13 (REQ-22, REQ-23): the help of `card land-batch` has no `--check-cmd`, and a tip that moves after each passing check stops the batch at the cap with `tip_kept_moving` · verify: `cargo test --locked --offline --test card_land_batch_cli seam_` and `cargo test --locked --offline --test card_land_batch_cli tip_`
- AC-14 (REQ-24): `CLAUDE.md` and `skills/moai/references/flows/next-flow.md` name `--card`, `card land-batch` and the full run every five landings · verify: `cargo test --locked --offline --test card_land_batch_cli docs_`

## Scope

- write: `.moai/specs/SPEC-20261124-card-select-batch-land/**`
- write: `src/check_affected.rs`
- write: `src/card/land.rs`, `src/card/land/**`
- write: `src/card/ready_receipt.rs`
- write: `tests/check_affected_card_select.rs`, `tests/card_land_batch_cli.rs`
- write (revision 3): `CLAUDE.md` (the `Verify` section), `skills/moai/references/flows/next-flow.md`; `CLAUDE.md` is a tier-2 path, so the change gets the full audit
- shared (line lease): `src/ops/card.rs` (the `land-batch` action and its dispatch), `src/ops/check.rs` (the `--card` flag), `docs/design-review/appendix/operations/card.md`, `tests/fixtures/state_folder_literals/allow.tsv`
- forbidden: `docs/leader-rules.md`, `docs/lane-rules.md`, `skills/moai/references/audit.md` (C-471)

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
| D-8 | The batch's refusals come from `land_preflight`, the same function as `land` stopped before its first write, with the expected first parent as an argument; no field is added to `LandRequest` | one source for the refusals, and `src/ops/card.rs` stays untouched | write the refusals twice, or add a field |
| D-9 | A full-set fallback names its reason: a wide class, a common word, or a `src` path that no test reaches (audit B1, 2026-11-24) | the first build chose zero tests on the hazard paths | go back to the card set for those paths |
| D-10 | The shared hub paths of C-464 (`src/ledger/mod.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs`, `src/profile.rs`) run the full set with `--card` and say why; the card set is for a range of modules that are no hub, and AC-1 and criterion 1 of C-470 read that way (leader decision, note 8636) | many tests rest on a hub, so a wide run is the safe side, and it agrees with the review feedback "ledger and common settings wide"; the real C-464 range measured 147 targets, one non-hub module 1 | go back to the 13-target reading of revision 1 |
| D-11 | The batch's scan cover is the receipt of its own wide run: the summary line `target/check-affected-<head12>/summary.txt` of the private worktree at the last join, read before the worktree goes, counts when it names that join, ends in `result=PASS` and has no `selection=` word (the full set runs every listed scan); the covered flag is an argument of `land_preflight` and of a new `land_in_batch` that reaches `scans::read` (D-8 stands: `LandRequest` gets no field), and the verdict receipt still covers a head the way it did | audit N3: a batch's wide run left no scan receipt, so enforce mode refused every head; the full run is a stronger receipt than the lane's, on the tree that lands | stop passing the flag |
| D-12 | The batch is confirmed by session: the entry `leader` of the lane registry must be live, in mode `leader`, and its `session` must equal the calling session (the rule of `skills/moai/references/internal/loop.md`, "The confirmed leader"); `--by leader` stays as the word of the call, `card land` is unchanged | audit N4: the word is a declaration any session can type; the registry already binds the leader to one host session id | drop the check, or move it to `card land` too |
| D-13 | The check program is set by `MOAI_TEST_LAND_BATCH_CHECK`, not by a flag; the flag `--check-cmd` is removed | audit N6: a product flag that swaps the wide check for any program lets a landing skip it; the seam stays for the tests, as `MOAI_FORGE_GH_BIN` does, and it counts no scan as covered unless it leaves the summary line | put the flag back |
| D-14 | A batch stops at 3 moved tips (`tip_kept_moving`) | audit N7: each re-run is a cold build, and a branch that moves after every check would loop; 3 is a chosen cap, not a measured one | raise the constant `MAX_TIP_MOVES` |
| D-7 | The name word of a changed module is `<folder>_<module>` below a folder and the module name for a top-level file; the cap of 10 counts the targets that reach a module by that word or by naming its path in their source, together | measured on the real C-464 range: the bare module word `sweep` or `work` reached the whole `card_*` family, and the path rule alone reached 14 targets for `src/ledger/reservation.rs`; the union cap gives 13 targets | go back to the bare word, or raise the cap |

## Result

Code head `7c8153bae1` on base `6559535dba`. Checks of 2026-11-24 on a Mac, load 21 to 35 from other lanes; every cargo run under `moai heavy-slot-acquire --name lane-470 --max 4`.

- Measured with the binary built from this lane (`target/debug/moai check affected --list`, 2026-11-25): (a) the real C-464 range (`dbc09e87a59f..98199c2d251f`) takes the full set with `--card`, 147 targets, because the range holds hub paths (`card_wide` names `src/card/mod.rs`, `src/card/sweep.rs`, `src/ledger/mod.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs`; reason first named `no test reached`, ignored words `card`, `ledger_reservation`); `src/card/sweep.rs` is named there because that module is not in this checkout, so no test names it. This replaces the 13 of revision 1, which was measured before the hub rule (D-9, decision note 8636). (b) A one-module range, `src/card/ready_leftovers.rs` only (commit `120f406769`): `--card` selects 1 target, the bare selection 94. The audit of note 8635 reports the same shape on its tree: `src/card/sweep.rs` alone 94 to 2, documents alone 79 to 2 (audit table, not re-run here). Logs: `target/agent-3/c464-card5.json`, `target/agent-3/leftovers-card.json`, `target/agent-3/leftovers-full.json`.
- The new tests ran with the whole neighbourhood in one nextest call (`--lib` and 10 test binaries, among them `check_affected_card_select`, `card_land_batch_cli`, `state_folder_literal_scan`, `hot_file_ops_merge`): 1991 passed, 0 failed (`target/run3.log`). The call before the last edits (`target/run2.log`) was 2014 of 2016, and the two failures were repaired: the two new test files were missing from `tests/fixtures/state_folder_literals/allow.tsv`, and the conflict test used a file that is tier 1 (a landing refused it for a missing receipt, not for the conflict).
- The verify commands of the AC lines, run as written (`target/verify1.out`): `check_affected_card_select` `card_set_` 2 passed, `wide_` 3 passed, `receipt_` 1 passed; `card_land_batch_cli` `batch_` 3 passed, `split_` 2 passed, `refuse_` 4 passed, each exit 0.
- Guard mutants (one slot call, `target/mutants1.py`, tree clean afterwards), every one killed with exit 100 by the named test: wide `Cargo.lock`, the cap of 10, scans by rule file only, the `selection=card` word, halving, the moved-tip re-read, two to four heads, leader only, the audited second parent, the audited clean join, the verdict read on the audited head. Not mutated: `conflicts_with` of `--card` and `--lane` (clap's own check, pinned by `card_and_lane_cannot_be_asked_together`) and the private-worktree removal (not a guard of its own; the tests list the leftover folders).
- AC-1: met — `check_affected_card_select card_set_` exit 0 (the non-hub fixture range selects at most 20 and names no ignored word) and `wide_the_hub_paths_of_c464` exit 0 (the four hub paths run the full set and each is named); the real one-module range selects 1 and the real C-464 range takes the full set (see the measurement above)
- AC-2: met — same command, 2 passed
- AC-3: met — `check_affected_card_select wide_` exit 0, 3 passed
- AC-4: met — `check_affected_card_select receipt_` exit 0, 1 passed; the unit test `a_card_summary_line_is_a_receipt_and_a_lane_one_is_not` of `ready_receipt.rs` passed in the nextest call
- AC-5: met — `card_land_batch_cli batch_` exit 0, 3 passed (one check, landed in the given order)
- AC-6: met — `card_land_batch_cli split_` exit 0, 2 passed (the failing head found by halving, the others landed)
- AC-7: met — same command, the passing first half lands before the rest is checked, and a moved tip is not reused (`batch_a_tip_that_moved_during_the_check_is_not_reused`)
- AC-8: met — `card_land_batch_cli refuse_` exit 0, 4 passed
- Revision 2 (repair after the audit FAIL, note 8610; base head `c7a81e91a8`, 2026-11-24T14:58Z): B1 the wide class now holds the tier-2 floor, `install/git-hooks/`, `src/git_hooks/`, `src/card/land/`, a common word and a `src` path that reaches no test (the auditor's 8-path table is pinned in `wide_classes_fall_back_to_the_full_set_and_say_why`); N1 `land` pins the first parent to the tip it read; N2 `land_preflight` judges every head before the first lands; N5 join commits name the SPEC and carry `Moai-Spec`; N8 the new text of `check.md` says 테스트; N9 the second parent is pinned by `refuse_a_join_whose_second_parent_is_not_the_audited_head`. Not repaired here (follow-up criteria of the leader): N3, N4, N6, N7.
- Revision 2 second repair (audit note 8635, decision note 8636): R2 AC-1, the Result lines and the fixture now read "a range of modules that are no hub is small, the hub paths run the full set and say why" (D-10); the fixture constant is `NON_HUB_RANGE`, the hub paths are `C464_HUBS`, and `wide_the_hub_paths_of_c464_run_the_full_set_and_name_each_path` pins them. R1 the two repair commits are one commit whose trailers (`Moai-Spec`, `Moai-Session`, `Moai-Verified`, `Co-Authored-By`) stand in the one last paragraph.
- Revision 2 checks: `cargo nextest run --test card_land_batch_cli --test card_land_cli --test check_affected_card_select` 89 passed, 0 failed (exit 0) under `moai heavy-slot-acquire --name agent-3 --max 4` (2026-11-25).
- AC-9: met — `card_land_batch_cli refuse_a_join` exit 0, 2 passed; AC-10: met — `card_land_batch_cli a_join_commit` exit 0, 1 passed
- Gaps: (1) the batch and the split ran through the real binary in temporary repositories built by the test fixture, not typed by hand in a separate throwaway repository. (2) The batch's default check runs `check affected` in-process on the joined tree and writes no scan receipts of its own. (3) The join commits carry a `Moai-Spec` trailer since revision 2. (4) D-1 (ledger means the storage layer) and D-5 (a verdict on the audited head covers a clean join) are readings and a policy extension the leader confirms at READY. (5) Criterion 3 of the card (CLAUDE.md and next-flow wording) belongs to C-471 and is not done here. (6) `src/ops/check.rs`, `docs/design-review/appendix/operations/check.md` and `tests/fixtures/state_folder_literals/allow.tsv` are outside the dispatched shared list and were edited under a line lease. (7) Windows and Linux were not run.
