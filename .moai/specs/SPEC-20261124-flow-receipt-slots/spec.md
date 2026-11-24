---
id: SPEC-20261124-flow-receipt-slots
title: "Reverting an earlier flow receipt keeps the new flow file's ownership hash"
status: approved
revision: 1
created: 2026-11-24
approval: {by: leader, at: 2026-11-24, basis: "leader decision note 8142 of card C-441 (answers note 8141, 2026-11-24T03:09:15Z, under the user's delegation of 2026-11-23T13:57Z, not a direct user pick), approving revision 1 at head 98b17fdda3 with choice (a); recorded by the lane, not approved by it", digest: "sha256:344041586d53bf222574dc0d21846c06efcd9aca1ddd34f26298bf9091619c22"}
---

# Reverting an earlier flow receipt keeps the new flow file's ownership hash

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-441 의 본문과 두 인수 기준(C-435 감사 N1), 리더 배차가 정했어요.

## Intent

`src/home_routing/rules.rs` records which flow rule file MoAI last wrote in one slot, `flow_sha256` of `~/.moai/home-routing/rules-<id>.json`. Two receipt kinds write that slot: the current `moai_flow` (`rules/moai-flow.md`) and the earlier layout's `flow` (`rules/moai/flow.md`, no longer written, still revertible). After a new apply wrote `rules/moai-flow.md`, reverting the earlier `flow` receipt removes `rules/moai/flow.md` and then sets the shared slot to that entry's `baseline_before` (null for a first apply), so the slot no longer holds the new file's hash. A later package with newer flow text then plans `rules/moai-flow.md` as `conflict` instead of `merge`. No file bytes are lost, and reverts in order (newest first) are correct. This SPEC makes the earlier receipt's revert leave a slot that the current flow file still owns, and pins the out-of-order revert in a temporary HOME.

## Non-goals

- No new key in the state file and no schema change: the one `flow_sha256` slot stays, so existing state files and older builds keep reading it (the state file denies unknown keys).
- No change to the apply, the plan or the revert of any other kind (`routing`, `old_block`, `moai_flow`).
- A `moai_flow` revert after the earlier receipt was already reverted out of order sets the slot to that receipt's `baseline_before`, which can name the removed earlier file's hash. It owns no file, a plan treats an absent file as `create`, and it is left as it is (follow-up if it ever matters).
- The real HOME is never touched; every test runs against a temporary HOME.

## Requirements

- REQ-1: When a revert undoes a receipt entry of the earlier kind `flow`, and the file `rules/moai-flow.md` of the same config folder holds the bytes whose digest `flow_sha256` records, the revert shall keep `flow_sha256` as it is.
- REQ-2 (REQ-1): IF `rules/moai-flow.md` is absent, not readable as a plain file, or does not hold the recorded digest, THEN the revert shall set `flow_sha256` to the entry's `baseline_before` as it does today.
- REQ-3: When every receipt of one config folder is reverted newest first, the revert shall restore the state file byte for byte.
- REQ-4 (REQ-1): IF a package with newer flow text is planned after that out-of-order revert, THEN the flow row shall be `merge`.
- REQ-5 (REQ-3): IF a revert skips an entry because its file changed since the apply, THEN the revert shall leave `flow_sha256` as the skipped entry found it.

## Acceptance

- AC-1 (REQ-1, REQ-4): a temporary HOME holds an earlier `flow` receipt entry whose file `rules/moai/flow.md` is untouched, then a current apply writes `rules/moai-flow.md`; reverting the earlier receipt first keeps `flow_sha256` equal to the digest of `rules/moai-flow.md`, and a plan of a package with newer flow text shows the flow row as `merge` · verify: `cargo test --locked --offline --test flow_receipt_slots_cli an_earlier_receipt_revert_keeps_the_new_files_hash`
- AC-2 (REQ-2): with `rules/moai-flow.md` absent, or edited so its digest differs from the slot, the earlier receipt's revert sets `flow_sha256` to the entry's `baseline_before` · verify: `cargo test --locked --offline --test flow_receipt_slots_cli an_earlier_receipt_revert_without_the_new_file_clears_the_slot`
- AC-3 (REQ-3): the earlier receipt applied, then the current apply, then the current receipt reverted, then the earlier receipt reverted: the state file equals its bytes before the first apply (absent or the same JSON bytes) · verify: `cargo test --locked --offline --test flow_receipt_slots_cli reverts_in_order_restore_the_state_file_byte_for_byte`

- AC-4 (REQ-5): an earlier receipt entry whose `rules/moai/flow.md` was edited after the apply is skipped as `user_changed_since_apply` and `flow_sha256` keeps its value · verify: `cargo test --locked --offline --test flow_receipt_slots_cli a_skipped_earlier_entry_leaves_the_slot`

## Scope

- write: `.moai/specs/SPEC-20261124-flow-receipt-slots/**`
- write: `src/home_routing/rules.rs`
- write: `tests/flow_receipt_slots_cli.rs`
- read-only: `src/home_routing.rs`, `tests/desktop_folder_rules_cli.rs`, `tests/home_routing_cli.rs`
- forbidden: the real Claude config folder, the real MoAI home and the real HOME (a test passes a temporary HOME only), and every other file under `src/` and `tests/`

## Plan

1. RED: `tests/flow_receipt_slots_cli.rs` pins AC-1 to AC-3 in a temporary HOME with a hand-written earlier `flow` receipt (the fixture of `desktop_folder_rules_cli.rs`) and a package with newer flow text; AC-1 fails before the code.
2. GREEN: in `revert_entry` (`rules.rs`), for kind `flow` only, read the state's `flow_sha256` and the digest of `rules/moai-flow.md` (same base and plain-file checks as the entry's own target); equal digests skip the slot write, every other case writes `baseline_before` as before.
3. Guard mutant: remove the new skip and run the AC-1 test; it must exit non-zero.
4. Result: the commands, their exit codes and the guard-mutant line.

## Risks

- Equal digests of the earlier and the current file (same text in both) are the case the digest comparison alone cannot tell apart: the slot is kept, which is right because `rules/moai-flow.md` exists and holds it. AC-1 uses the equal-text case and the newer-text case.
- The state file is read twice in one revert; the second read happens after the file removal and before the slot write, so a failing read must stop the revert as it does today (`?`), not skip the write silently.
- Not measured: Windows and Linux (the revert uses only the paths the entry already resolves).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Keep the one `flow_sha256` slot and skip the slot write of an earlier-kind revert while the current flow file holds that digest (card option "clear the slot only when the reverted kind owns it"); the alternative, a second slot per kind, is rejected because it adds a key to a state file that denies unknown keys and cannot tell an existing slot's owner | card C-441 options; `rules.rs` state struct reading | restore `update_rules_state` call of the `flow` kind unconditionally |

## Result

Code head `1bf7dc9037` on base `f1862176de` (changed files: `src/home_routing/rules.rs`, `tests/flow_receipt_slots_cli.rs`, this SPEC); the product change itself is `f26190f687`, `bf023942bc` adds the repair tests below, `1bf7dc9037` only re-breaks their lines the way rustfmt wants, and the commit holding this Result follows it and changes this SPEC only. Checks of 2026-11-24 on a Mac, a temporary HOME only.

- RED (`16a1df4e46`, old code): `cargo test --locked --offline --test flow_receipt_slots_cli` exit 101, 3 passed and 1 failed: `an_earlier_receipt_revert_keeps_the_new_files_hash` (`same_text=true: the slot still holds the new file's digest`, left `Null`, right the digest). The other three tests pin behavior that already held (AC-2 to AC-4).
- GREEN (`f26190f687`): the same command, 4 passed, 0 failed. The change is `current_flow_holds_slot` and one condition in `revert_entry`: for kind `flow` only, the slot write is skipped while `rules/moai-flow.md` is a plain file holding the digest the slot records; a link, an unreadable or absent file, a different digest or an empty slot writes `baseline_before` as before.
- Guard mutant on the committed file: `current_flow_holds_slot` made to return `false` (the guarded code removed), same command, exit 101, `an_earlier_receipt_revert_keeps_the_new_files_hash` FAILED, the other three passed. The mutant was reverted by hand and the tree read clean afterwards.
- Dangerous path through the built binary (`moai-cli` of this tree, `HOME` of the child set to a throwaway folder): an earlier `flow` receipt and `rules/moai/flow.md` written by hand, `home-routing-apply` of `skills/moai`, then `home-routing-revert` of the earlier receipt. The slot after the revert was `sha256:43c9b1a6…101f`, equal to the sha256 of `rules/moai-flow.md`; the earlier file was gone; the user's `CLAUDE.md` in the same folder had the same sha256 (`d3612e1b…6345`) before and after. The same run on the mutant binary printed `None` for the slot (the defect).
- Repair after the audit of `bca773f817` (two guards the first tests did not pin), tests added at `bf023942bc` (line breaks fixed at `1bf7dc9037`, the mutants below ran at `bf023942bc`): `a_routing_receipt_revert_restores_its_own_slot_while_the_flow_slot_is_held` (the kind restriction in `revert_entry`: while `rules/moai-flow.md` holds the flow slot, reverting a receipt that changed only the routing text gives `routing_sha256` back and leaves the flow slot) and `a_linked_new_file_with_the_slots_bytes_does_not_hold_the_slot` (unix only: `rules/moai-flow.md` is a link to a file with the slot's exact bytes, so the earlier receipt's revert sets the slot to null and leaves the link and its target). Guard mutants on the committed test file, `cargo nextest run --locked --offline --test flow_receipt_slots_cli`: the `kind == "flow"` restriction removed, exit 100, the routing test FAILED; `current_file` in `current_flow_holds_slot` replaced by a link-following `fs::read`, exit 100, the link test FAILED. Both mutants were reverted with `git checkout` and the tree read clean.
- AC-1: met — `cargo test --locked --offline --test flow_receipt_slots_cli an_earlier_receipt_revert_keeps_the_new_files_hash` exit 0; equal-text and newer-text cases keep the slot, and a package with newer flow text plans `merge`.
- AC-2: met — same file, `an_earlier_receipt_revert_without_the_new_file_clears_the_slot` passed: an absent file and a user-edited file (its bytes unchanged) both leave the slot at `baseline_before` (null).
- AC-3: met — same file, `reverts_in_order_restore_the_state_file_byte_for_byte` passed: reverting the current receipt restores the bytes the earlier layout left, and then the earlier receipt gives the bytes a first apply and its revert leave.
- AC-4: met — same file, `a_skipped_earlier_entry_leaves_the_slot` passed: `user_changed_since_apply` is reported and the slot keeps its value.
- Gaps: the out-of-order case where the current receipt is reverted after the earlier one leaves the slot at that receipt's `baseline_before`, which can name the removed earlier file's hash (non-goal, harmless: a plan of an absent file is `create`). Windows and Linux not run. `moai check affected` is the READY note's receipt, not this Result's.
