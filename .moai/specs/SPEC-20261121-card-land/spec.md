---
id: SPEC-20261121-card-land
title: "card land moves the integration branch to a candidate commit and closes the card in one call"
status: draft
revision: 6
created: 2026-11-21
---

# card land moves the integration branch to a candidate commit and closes the card in one call

## Intent

Landing is two hand steps today: the leader moves the integration branch (a compare-and-swap in a shell) and then calls `card finish` with the head. Between them the branch has moved and the card is still open, and nothing but a note says so. This is slice 1a of SPEC-20261121-flow-skeleton: `card land` does both in one call, checks everything it can before it moves the branch, and is safe to run again after a stop, because git is the record of the move and the ledger is the record of the close.

Slice 1b (the merge of a head onto the tip in a scratch worktree) follows directly and is not in this SPEC: here the candidate is a commit that already has the tip on its first-parent line. The three non-blocking findings of the audit of SPEC-20261121-card-finish ride with this slice: the tag refusal test asserts the wording of the refusal, the wording of the AC-19 line of that SPEC matches what the code says, and the words "same as `card done`" are replaced where the SPEC is read from the proven tip.

The command is an action of `moai card` (decision D-1 of the flow skeleton). Until the typed verdict exists only the leader lands (D-3); the audit evidence is not checked here and the leader checks the verdict note by hand (D-2).

## Non-goals

- The merge of a head onto the tip, the land evidence (parents, clean merge tree, manifest match, tier floor) and any check of an audit verdict: slice 1b and step 4 of SPEC-20261121-flow-skeleton.
- A land of a card that has no integration branch (a profile that is not m-flow): it refuses, as `card finish` does for a landing it cannot prove.
- A new ledger table or a new column, and any new event kind but the start row `card.land_started` of SPEC-20261121-ref-guard: the close is `card finish`'s own transaction and writes its own events.
- The clap lines of the command in `src/ops/card.rs` are the last step, by the same pair approval as the finish command had.

## Requirements

- REQ-1: When the leader runs the card land command with a candidate commit that has the integration tip on its first-parent line, the card land command shall move the integration branch to the candidate by one compare-and-swap reference update and then close the card as the card finish command does · failure: none — a branch that moved since the tip was read, a refused update, a refused close or any refusal below leaves the ledger as it was apart from the start row that SPEC-20261121-ref-guard writes before the update, and a refused close after the move leaves the branch moved and says so without saying that nothing was changed, and a reference update that git refuses while the branch has not moved is reported with git's reason and not as tip_moved.
- REQ-2: When the candidate is already inside the integration tip, the card land command shall skip the reference update and close the card with the candidate as its landed head · failure: none — a card already done with the same head answers done and writes nothing, and a card already done with another head refuses.
- REQ-3: IF the integration tip is not on the first-parent line of the candidate, THEN the card land command shall refuse with the reason stale_base that names the tip and the candidate · failure: none — nothing is written and the branch stays where it is.
- REQ-4: When the card land command is called, the card land command shall make every check that the card finish command would make before it moves the branch · failure: none — a card in a state that cannot be finished, a generation that does not match, a met option that names no open criterion or names one twice, a SPEC that fails at the candidate, or a card that closes on a report (a research or audit card is `card finish`'s) refuses before the branch moves.
- REQ-5: When the integration branch is checked out in any worktree of the repository, the card land command shall refuse before the reference update · failure: none — a move under a checked-out branch would leave that worktree behind the reference, so nothing is written and the answer names the worktree.
- REQ-6: When the candidate names no commit of the repository, the card land command shall refuse before the reference update · failure: none — nothing is written.
- REQ-7: When the caller is not the leader, the card land command shall refuse until the typed verdict exists · failure: none — the refusal names the leader word and nothing is written.
- REQ-8: When an open criterion has no met option, the card land command shall move the branch and record the landing as the card finish command does and leave the card open · failure: none — the answer says closed false and names the open criteria, and a rerun with the met options closes the card.
- REQ-9: When the card land command is run again after a stop between the reference update and the close, the card land command shall answer as one uninterrupted run would · failure: none — the rerun finds the candidate inside the tip, skips the update and closes the card, with one landing event and one finish event.
- REQ-10: When the profile names an integration branch that is not under refs/heads, the card land command shall refuse as the card finish command does · failure: none — nothing is written.
- REQ-11: When the card land command runs git, the card land command shall run it through the git environment scrub and with a reference-log message that names the candidate and no card number · failure: none — a git that cannot run refuses before any write.
- REQ-12: When the integration branch points at a commit that is not the result commit of the newest successful landing of that branch and not the tip or the target of an interrupted landing after it, the card land command shall refuse before any write and name the drift and the `moai card drift --restore` command that puts the branch back · failure: none — the answer is a conflict that holds both commits, the ledger is unchanged, a commit added by hand on top of the last landing is drift too, and a branch with no start row has nothing to compare and lands.

## Acceptance

- AC-1 (REQ-1): a card in review whose candidate is a merge commit with the tip as its first parent lands in one call: the branch points at the candidate, the card is done, its reservation is released, the answer names the old tip, the new tip and the moved flag, and the reference log names the candidate · verify: `cargo test --locked --offline --test card_land_cli ac1_`
- AC-2 (REQ-1): a branch moved by another process between the read of the tip and the update leaves the branch where that process put it and every table as it was apart from the start row of SPEC-20261121-ref-guard, and the answer names tip_moved · verify: `cargo test --locked --offline --test card_land_cli ac2_`
- AC-3 (REQ-2): a candidate already inside the tip closes the card with no reference update, and a second call on the done card answers done with nothing written while another head refuses · verify: `cargo test --locked --offline --test card_land_cli ac3_`
- AC-4 (REQ-3): a candidate that branched from an older tip refuses with stale_base naming both commits, and so does a candidate that holds the tip only on a second parent, with the branch and the ledger unchanged · verify: `cargo test --locked --offline --test card_land_cli ac4_`
- AC-5 (REQ-4): a card in a state that cannot be finished, a wrong generation, a met option for a criterion that is not open, a met option named twice and a SPEC with errors at the candidate each refuse with the branch and the ledger unchanged · verify: `cargo test --locked --offline --test card_land_cli ac5_`
- AC-6 (REQ-5, REQ-6): the integration branch checked out in a second worktree refuses and names it, and a candidate that is not a commit refuses, both before any write · verify: `cargo test --locked --offline --test card_land_cli ac6_`
- AC-7 (REQ-7): a call with any caller word other than the leader word refuses with the branch and the ledger unchanged · verify: `cargo test --locked --offline --test card_land_cli ac7_`
- AC-8 (REQ-8): an open criterion with no met option moves the branch, records one landing event, leaves the card open and answers its open criteria, and the same call with the met options closes it with no second event and no second move · verify: `cargo test --locked --offline --test card_land_cli ac8_`
- AC-9 (REQ-9): a branch moved to the candidate by hand and a card still open land in one call that closes it with one landing event and one finish event · verify: `cargo test --locked --offline --test card_land_cli ac9_`
- AC-10 (REQ-10, REQ-11): a profile whose integration branch names a tag path refuses with the words "is not a branch name", a git that cannot run refuses before any write, and the git calls of the command take their program from the same variable as the other card commands · verify: `cargo test --locked --offline --test card_land_cli ac10_`
- AC-11 (REQ-10): the finish command refuses a profile whose integration branch names a tag path with the words "is not a branch name", so the arm in `src/card/landed.rs` cannot be removed unseen · verify: `cargo test --locked --offline --test card_finish_cli ac27_`
- AC-12 (REQ-4): the AC-19 line of SPEC-20261121-card-finish says what the code and its test say, and no text of the finish path says the SPEC is read "as `card done` does" where it is read from the proven tip · verify: `cargo test --locked --offline --test card_finish_cli ac19_`
- AC-13 (REQ-1): a close that is refused after the branch moved (a criterion marked met by another session between the checks and the close) answers that the branch already moved to the candidate and never says that nothing was changed, leaves the branch at the candidate and the ledger without a landing event, and the same call run again closes the card without a second move · verify: `cargo test --locked --offline --test card_land_cli ac13_`
- AC-14 (REQ-4): the SPEC is read at the candidate and not at the old tip: a good SPEC on the tip with a broken one at the candidate refuses, and a broken one on the tip with a good one at the candidate lands · verify: `cargo test --locked --offline --test card_land_cli ac14_`
- AC-15 (REQ-1, REQ-2): the landing event records the candidate as head and the new tip (the candidate after a move, the tip read when the head was already inside) and the caller word · verify: `cargo test --locked --offline --test card_land_cli ac15_`
- AC-16 (REQ-4): a research card is refused with the words that a card closing on a report is the finish command's, and the branch and the ledger stay as they were · verify: `cargo test --locked --offline --test card_land_cli ac16_`
- AC-17 (REQ-11): a caller whose environment names another repository (the git directory, work tree, index file and namespace variables) lands in the project's own repository · verify: `cargo test --locked --offline --test card_land_cli ac17_`
- AC-18 (REQ-1): a reference update that git refuses because a lock file is in the way is reported with git's reason and not as tip_moved, and the branch stays and the ledger holds at most the start row of SPEC-20261121-ref-guard · verify: `cargo test --locked --offline --test card_land_cli ac18_`
- AC-19 (REQ-12): `git branch -C` of an unrelated commit over the integration branch makes the next landing refuse with the word drift, both commits and the `moai card drift --restore` command, nothing written; after the recovery by that command the same call lands · verify: `cargo test --locked --offline --test card_land_cli ac23_`
- AC-20 (REQ-12): a commit added by hand on top of the last landing is drift, and so is a rewind to the tip the previous landing started from; an interrupted landing is accepted on its from-tip and on its target, and an interrupted landing after a landing does not change the commit a restore names · verify: `cargo test --locked --offline --test card_land_cli ac25_`, `ac26_`, `ac27_` and `ac31_`
- AC-21 (REQ-9): the lander killed after the move and before the close, then run again, moves nothing, writes no second start row and closes · verify: `cargo test --locked --offline --test card_land_cli ac24_`
- AC-22 (REQ-12): `moai card drift` reads and writes nothing and says `ok`, `drift` or `no_landing` with the recorded commit · verify: `cargo test --locked --offline --test card_land_cli ac28_`
- AC-23 (REQ-12): `moai card drift --restore` previews the move and the commits it drops and writes only a one-use nonce into the git folder; with `--apply` it needs `--by leader`, the digest of the latest preview of the same state and a permission reference, moves the branch with the hook installed (a rewind, a descendant or an unrelated copy) under a restore intent and writes one `card.tip_restored` row · verify: `cargo test --locked --offline --test card_land_cli ac29_`, `ac30_`, `ac32_`, `ac33_` and `ac34_`

## Scope

- write: `.moai/specs/SPEC-20261121-card-land/spec.md`, `.moai/specs/SPEC-20261121-card-finish/spec.md`
- write: `src/card/land.rs`, `src/card/landed.rs`, `src/card/finish.rs`, `src/card/mod.rs`, `src/ledger/workflow.rs`
- write: `tests/card_land_cli.rs`, `tests/card_finish_cli.rs`, `docs/design-review/appendix/operations/card.md`
- last step: the clap lines of the land command in `src/ops/card.rs` only (the subcommand and its dispatch arm); another card holds that file and the leader approves the lines as for the finish command.

## Decisions

- D-1: The check part of `workflow::finish` becomes one function that a read-only caller and the write transaction both call, so that the checks before the move and the checks at the close are the same code and cannot drift. The write transaction repeats them, because a race can change the ledger between the two.
- D-2: The reference update is `update-ref refs/heads/<branch> <candidate> <tip read before>`: git refuses when the branch is not at that tip, which is the compare-and-swap. The branch is named under `refs/heads/` as `tip_of` names it.
- D-3: First-parent line: with the tip an ancestor of the candidate, the oldest commit of `git rev-list --first-parent <tip>..<candidate>` has the tip as its first parent. A candidate with the tip only on a second parent is stale_base: its first-parent line is another history.
- D-4: The landing event records the candidate as head and the new tip (the candidate after the move, the tip read when the head was already inside), so the card that landed by `card land` and the one that landed by `card finish --head` read the same.
- D-5: The git worktree list is read with `git worktree list --porcelain`; a `branch refs/heads/<name>` line for the integration branch refuses, and a detached or bare worktree does not.
- D-6: Only the leader word is accepted as the caller in this slice (D-3 of the flow skeleton); the holder lands with a typed verdict in step 4.

## Plan

1. Tests first (RED): `tests/card_land_cli.rs` with AC-1 to AC-10 on a temporary home and a temporary repository with an integration branch, and the AC-11 and AC-12 changes in `tests/card_finish_cli.rs`.
2. The check function of `workflow::finish` split out and the finish transaction calling it (a refactor with the existing finish tests green), then `src/card/land.rs` and the one write helper of `src/card/landed.rs` that runs `update-ref` through the same scrub as the read helper.
3. The SPEC and documentation wording of AC-12 and the card table row of `docs/design-review/appendix/operations/card.md`.
4. Light checks, the mutants (at most five, after the commit), one affected run, then the READY note with head and base.
5. Last: the clap lines in `src/ops/card.rs` by pair approval.

## Risks

- The drift check compares with the ledger's newest successful landing of the branch (a start row followed by a landed row of the same card and head). A branch the ledger never moved has nothing to compare, so a copy before the first landing is not seen; and a first landing that was interrupted uses its own from-tip as the recorded base. A hand-made fast-forward is drift now (the revision-5 decision of the leader on the second re-check): a commit nobody landed is not a landing, and a rewind accepted as the previous start would make the next landing drop a landed change. A branch fast-forwarded by hand after its last landing is refused until the user picks: `moai card drift --restore` previews the move back and lists the commits it drops, so nothing the hand merge added is lost without a look. The hook of SPEC-20261121-ref-guard cannot see `git branch -c` and `-C`; this check is the layer that notices them.
- A move under a race is refused by git itself, but the close can still be refused after the move when the ledger changed between the checks and the write; the answer says the branch moved, and the rerun closes the card or repeats the refusal. That is the cost of git and the ledger being two stores.
- `MOAI_GIT_BIN` lets a test put a wrapper around git; the wrapper that makes the branch move at the wrong moment (AC-2) is written by the test and is not a spawn of git by the test.

## Result

Slice 1a with the repairs of the first audit (the first audit of `4b43a97984` ended FAIL, small: a false `nothing was changed` after the move, five guards with no test, and three non-blocking items). The numbers are from one run of `moai check affected dcdaeaa98a73..94df32e4529b` (82 targets: the whole lib and both bins, `diff_check`, `fmt`, `ops_rustfmt`, `clippy`, `test_lib`, `test_bins`, `test_targets`, `build_cli` and `python` all exit 0, result PASS). The SPEC text of this Result and the AC lines of the repairs were written after that run and change no code. The re-audit goes to the first auditor and covers the repairs only.

- AC-1 met: `ac1_` (1 test) passed: the candidate is now a real merge commit with the tip as its first parent; the branch points at it, the card is done, the reservation is released and the reference log names the candidate. AC-2 met: `ac2_` (1 test) passed.
- AC-3 met: `ac3_` (1 test) passed. AC-4 met: `ac4_` (1 test) passed: an older base and a tip held only on a second parent both refuse with `stale_base`.
- AC-5 met: `ac5_` (1 test) passed, and the generation refusal now says `nothing was changed` once (RED before the repair: it said it twice).
- AC-6 met: `ac6_` (1 test) passed. AC-7 met: `ac7_` (1 test) passed. AC-8 met: `ac8_` (1 test) passed. AC-9 met: `ac9_` (1 test) passed. AC-10 met: `ac10_` (1 test) passed.
- AC-11 met: `ac27_` of `card_finish_cli` passed. AC-12 met: `ac19_` of `card_finish_cli` passed and now also refuses a tip that holds no copy of the SPEC (`is absent`), which closes the audit's gap between the AC-19 line and its test.
- AC-13 met: `ac13_` (1 test) passed: a criterion marked met by another session between the checks and the close makes the close refuse; the answer says the branch already moved to the candidate and never `nothing was changed`, the branch stays at the candidate, no landing event exists, and the same call run again closes the card with `moved` false. RED before the repair.
- AC-14 met: `ac14_` (1 test) passed: a good SPEC on the tip with a broken one at the candidate refuses and the reverse lands. AC-15 met: `ac15_` (1 test) passed: the landing event holds the head, the new tip and the caller word, for a move and for a head already inside.
- AC-16 met: `ac16_` (1 test) passed: a research card is refused with the finish-command words and nothing moves. AC-17 met: `ac17_` (1 test) passed: `GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE` and `GIT_NAMESPACE` of the caller do not reach the git of the command.
- AC-18 met: `ac18_` (1 test) passed: a stale lock file on the branch reference is reported with git's reason and not as `tip_moved`, with the branch and the ledger unchanged. RED before the repair.
- Whole targets in the run: `card_land_cli` 16 and `card_finish_cli` 45 passed, none failed (run alone before the affected run: 16 of 16 and 45 of 45).
- Mutants on the five guards of the audit (byte-copy restore, tree clean after each): the SPEC read at the old tip (killed by `ac14_`), the new tip not recorded (`ac15_` and `ac1_`), the already-moved sentence removed (`ac13_`), the report-card refusal removed (`ac16_`) and the git environment scrub removed (`ac17_`): all five killed.
- Not run: the re-audit (the first auditor's), the Windows type check, `self_check.sh`, a full test run outside the affected sets.
- Left as the SPEC states them: who may land is the leader word only (D-3, D-6) and the ledger row of the move before the close comes with the next slice; the annotated tag object id that passes the commit check is not repaired here.
- Revision 5 (the second re-check's B-3 and N-18, N-19): the drift rule is strict (REQ-12, AC-19 to AC-23), `card drift` is the read-only view and the preview-first restore, and the refusal says `nothing was written` once. Red first: `ac23_`, `ac25_`, `ac26_`, `ac28_` and `ac29_` failed on the earlier code. Whole `card_land_cli` target: 29 passed, none failed. Five mutants of the new module (byte-copy restore, tree clean after): M17 (the newest start row read as the landing) first survived and is now killed by `ac31_`; M18 (target of an interrupted first landing not accepted) killed by `ac13_`, `ac24_`, `ac27_`; M19 (the from-tip of a finished landing accepted) killed by `ac26_`, `ac28_`, `ac29_`; M20 (digest not compared) killed by `ac29_`, `ac30_`; M21 (leader word not checked) killed by `ac29_`.
- Revision 6 (SPEC-20261121-ref-guard revision 8, D-17): the restore runs under the installed hook through a restore intent, the preview digest is bound to a one-use nonce, and the plain view writes no file. `card_land_cli` 32 passed, none failed.
