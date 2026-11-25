---
id: SPEC-20261125-flow-receipt-revert-slot
title: "Reverting an earlier flow receipt after a later apply removed its file clears the slot, and the card check selects the slot tests"
status: draft
revision: 1
created: 2026-11-25
---

# Reverting an earlier flow receipt after a later apply removed its file clears the slot, and the card check selects the slot tests

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드의 본문과 두 인수 기준, 리더 배차(원인 가설과 "되돌리기의 바이트 단위 복원 단언은 약하게 하지 않는다")가 정했어요.

## Intent

`src/home_routing/rules.rs` keeps one flow ownership slot, `flow_sha256` of `~/.moai/home-routing/rules-<id>.json`, and `revert_entry` gives that slot back when it reverts a receipt entry. Since `3343c55bf3` (SPEC-20261125-update-conflict-baseline) a current apply copies the earlier layout's `rules/moai/flow.md` to a backup folder and removes it, and that removal has no receipt entry. Reverting the earlier receipt afterwards finds the file gone, compares `None` with the recorded digest, skips the entry as `user_changed_since_apply` and never reaches the slot write, so the slot keeps a digest of a file that no longer exists. Two tests of `tests/flow_receipt_slots_cli.rs` have failed since that landing (`an_earlier_receipt_revert_without_the_new_file_clears_the_slot`, `reverts_in_order_restore_the_state_file_byte_for_byte`; RED log in the Result), and the wide check of the batch that landed it did not pick the file, because the change to `src/home_routing/rules.rs` reaches no test that names that module path and the file belongs to another test area than the source. This SPEC makes the earlier entry's revert treat its already removed file as reverted and run the slot rule, and makes the card selection reach tests that name the folder module of a changed submodule.

## Non-goals

- No change to the apply, the plan, the backup or the removal of the earlier flow rule file (`3343c55bf3`), and no new receipt entry or receipt field for that removal.
- No change to the revert of the kinds `routing`, `moai_flow` and `old_block`, and no change to the skip of an entry whose existing file changed since the apply.
- The two failing tests and their byte-for-byte assertion are not weakened or rewritten to expect the stale slot.
- No change to the full set, the wide set or the batch landing check of `check affected`; only the card selection gains tests, never loses one.
- The real HOME is never touched; every test runs against a temporary HOME.

## Requirements

- REQ-1: When a revert undoes a receipt entry of the earlier kind `flow` that created its file (`before_sha256` null) and that file is absent from the config folder, the revert shall count the entry as reverted and write no file.
- REQ-2 (REQ-1): IF the entry is counted as reverted with its file absent, THEN the revert shall give the flow slot the same value as after a removal: `flow_sha256` stays while `rules/moai-flow.md` holds the recorded digest and is otherwise set to the entry's `baseline_before`.
- REQ-3 (REQ-1): IF the entry recorded earlier bytes (`before_sha256` set) and its file is absent, THEN the revert shall skip it as `user_changed_since_apply` and leave `flow_sha256` as it was.
- REQ-4: When every receipt of one config folder is reverted newest first after a current apply removed the earlier flow rule file, the revert shall restore the state file byte for byte. · failure: none — each failing step of REQ-1 to REQ-3 already has its own requirement
- REQ-5: When `check affected --card` selects tests for a changed source file below a folder `src/<dir>/`, it shall also select each test whose source names the module `moai_cli::<dir>` as a whole path while no more than `COMMON_WORD_CAP` tests name it, keeping every test it selects today.
- REQ-6 (REQ-5): IF more than `COMMON_WORD_CAP` tests name `moai_cli::<dir>`, THEN the selection shall add none of them and select as it does today.

## Acceptance

- AC-1 (REQ-1, REQ-2): with the earlier file absent, or removed by a current apply and the new file then edited, the earlier receipt's revert answers `reverted` and `flow_sha256` is null · verify: `cargo test --locked --offline --test flow_receipt_slots_cli an_earlier_receipt_revert_without_the_new_file_clears_the_slot`
- AC-2 (REQ-4): earlier receipt, current apply (which removes the earlier file), current receipt reverted, earlier receipt reverted: the state file equals the bytes a first apply and its revert leave, and the earlier bytes before the current apply are restored in between · verify: `cargo test --locked --offline --test flow_receipt_slots_cli reverts_in_order_restore_the_state_file_byte_for_byte`
- AC-3 (REQ-3): an earlier entry that recorded `before_sha256` with its file absent is skipped as `user_changed_since_apply` and the slot keeps its value · verify: `cargo test --locked --offline --test flow_receipt_slots_cli an_absent_earlier_file_with_recorded_bytes_is_still_skipped`
- AC-4 (REQ-5, REQ-6): a changed `src/home_routing/rules.rs` selects the test that names `moai_cli::home_routing` and keeps the area set it selected before; a folder module named by more than the cap selects none of them · verify: `cargo test --locked --offline --test check_affected_card_areas a_submodule_change_selects_the_tests_that_name_its_folder_module`
- AC-5 (REQ-5): `moai check affected --list --card <base>..<head>` over the head that changes only `src/home_routing/rules.rs` lists `flow_receipt_slots_cli` among the rust targets, and its target count is recorded beside the count before the change · verify: `moai check affected --list --card <base>..<head>` (the Result names the range)

## Scope

- write: `.moai/specs/SPEC-20261125-flow-receipt-revert-slot/**`
- write: `src/home_routing/rules.rs`
- write: `tests/flow_receipt_slots_cli.rs`
- write: `src/check_affected.rs`
- write: `tests/check_affected_card_areas.rs`
- read-only: `src/home_routing.rs`, `tests/areas.toml`, `tests/home_routing_cli.rs`, `tests/desktop_folder_rules_cli.rs`, `tests/update_conflict_baseline_cli.rs`
- forbidden: the real Claude config folder, the real MoAI home and the real HOME (tests pass a temporary HOME only), `tests/areas.toml` (a shared file; no line changes), and every other file under `src/` and `tests/`

## Plan

1. RED: the two failing tests (log `target/c545/red.log`: 4 passed, 2 failed, exit 101) plus the new AC-3 test and the AC-4 selection test, both failing before the code.
2. GREEN, revert: in `revert_entry` (`rules.rs`) an entry of kind `flow` with `before_sha256` null whose target is absent skips the file step and goes on to the slot rule that exists after the file step (`current_flow_holds_slot`, else `baseline_before`); an absent file of any other entry stays the skip it is today.
3. GREEN, selection: in `select_card_with_registry` (`check_affected.rs`) a changed `src/<dir>/<leaf>.rs` adds the tests naming `moai_cli::<dir>` as a whole path when at most `COMMON_WORD_CAP` do, whichever branch the file took, with the reason `module:<dir>` in `card_by`.
4. Guard mutants at tier 2: the `flow`/`before_sha256` condition of step 2 removed (the AC-3 test must fail), the step 3 addition removed (the AC-4 test must fail).
5. Result: commands, exit codes, target counts before and after, the guard-mutant lines.

## Risks

- A user who deleted the earlier file by hand gets the same outcome: `reverted` with a path that this revert did not remove, and the slot cleared. Nothing was written and nothing is lost, because the receipt created that file and its bytes live in the backup folder the removal made.
- An earlier entry with `before_sha256` set whose file a current apply removed still skips and leaves the slot stale (REQ-3). Restoring the older bytes would bring back the layout the update removed; the case needs its own decision and is left as a follow-up line in `docs/later.md`.
- The selection adds up to `COMMON_WORD_CAP` tests per changed folder module; the Result records the count for `src/home_routing/rules.rs` before and after. A folder that more tests name keeps today's selection (REQ-6).
- Not measured: Windows and Linux (the revert uses only paths the entry already resolves).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Recommended default: an earlier-kind (`flow`) entry that created its file (`before_sha256` null) and finds it absent counts as reverted and runs the slot rule; the alternatives are (b) every kind treats an absent file so, rejected because the other kinds' files are the user's to keep and their absence is not a product removal, (c) a receipt entry for the removal, rejected because it changes the receipt shape older builds read, (d) expect a stale slot in the tests, rejected by the card (the byte-for-byte assertion stays) | the card's cause hypothesis, confirmed by the RED log: the earlier file is gone before the revert, so the entry is skipped and the slot write is never reached | restore the unconditional `user_changed_since_apply` skip for an absent file; the two tests then fail again |
| D-2 | Recommended default: the card selection adds the tests that name the folder module of a changed submodule, capped at `COMMON_WORD_CAP`, on top of today's selection; the alternatives are (b) make `src/home_routing/**` a wide path, rejected because a full set per change of this folder is the cost the card selection exists to avoid, (c) move `flow_receipt_slots_cli` into the `home` area by renaming it, rejected because three SPECs pin its name and the miss would return for the next test of another area | the miss: the test names `moai_cli::home_routing`, no test names `home_routing::rules`, and the file's first word `flow` is in the `lane` area while the source is in `home` | remove the added match; the selection is then as before this change |
