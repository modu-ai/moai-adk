---
id: SPEC-20261121-ref-guard
title: "a reference-transaction guard keeps the integration branch moving only through the landing command"
status: draft
revision: 1
created: 2026-11-21
---

# a reference-transaction guard keeps the integration branch moving only through the landing command

## Intent

`card land` (SPEC-20261121-card-land) moves the integration branch with one compare-and-swap, but nothing stops any other git command from moving it: a lane that runs `git branch -f`, a commit made in a worktree that has the branch checked out, a `reset --hard` or a `fetch` into it all move the one branch every landing is measured against. The audit of the landing command also left one open item for this slice: a stop between the move and the close leaves the branch moved and no ledger row of the move.

This SPEC is slice 1b of SPEC-20261121-flow-skeleton: a git `reference-transaction` hook that refuses every create, move and delete of the integration branch unless a short record written by the lander names exactly that move, plus the case and alias rules of the integration branch's name, plus the ledger row that `card land` writes before it moves the branch. The merge of a head onto the tip in a scratch worktree and the land evidence are the slice after this one; the master SPEC's plan text still calls the land evidence slice 1b and is revised with that slice.

Measured on this machine in scratch repositories (git 2.54.0, `reference-transaction` hook that only logs): the old value the hook reads is zero whenever the caller did not name one, so the guard resolves the current value itself; `git pack-refs --all` and `git gc` send two transactions, first `zero → value` and then the prune of the loose copy as `value → zero` with the loose file present, the packed entry equal to the current value and the old value named; `git branch -D` sends `zero → zero` for a loose-only and a packed-only ref; `git symbolic-ref` sends the new value as `ref:<target>`; the hook runs twice per state when the loose and the packed store both change; and a hook at the `committed` state can run `update-ref` and recreate a deleted ref. The leader measured that the hook, placed in the shared hooks folder, fires for the primary worktree and every linked worktree and refuses all eleven ways of moving the ref that were tried, that an environment variable is no pass, and that a lower-case alias of a mixed-case branch overwrote the branch's own file on a case-insensitive file system and lost the branch.

## Non-goals

- The merge of a head onto the tip in a scratch worktree, the land evidence and any check of an audit verdict: the slice after this one.
- Every moai command that takes a branch name accepting any spelling of the integration branch's name, the refusal on the pre-push side and every remote ref: they belong to the remote copy step, because no remote copy of the integration branch exists yet.
- Installing the hook file. The hook text is provided here as bytes; the managed-hook installer of the go-path-removal work adds it as a third managed name by a pair request on that module, and until then tests and the leader install the text by hand.
- Repairing a case variant that already exists. The guard refuses to touch it and says so; removing it is a manual step with the hook disabled for that one command.
- A defence against a caller that edits `.git` or removes the hook on purpose: the guard stops accidents, not sabotage (the hook-guard rules already refuse `core.hooksPath` on a command line).
- Closing the one measured gap of the prune rule (D-3): a later slice may restore the ref from the `committed` state.

## Requirements

- REQ-1: When git prepares a reference transaction that creates, moves or deletes the integration branch, the reference guard shall refuse it unless the intent record names exactly that ref, the value the ref has now and the new value, and the current value is an ancestor of the new value · failure: none — a refusal exits non-zero before any ref changes and names the missing or mismatching part, a creation needs the record but no ancestor test, no record ever allows a deletion, and a replayed or rewound move is refused.
- REQ-2: When a reference transaction leaves the integration branch at the value it already has, including the prune of its loose copy after packing, the reference guard shall let the transaction pass · failure: none — a delete that names no old value, a loose ref that is not packed and a packed ref with no loose copy are deletions and are refused.
- REQ-3: IF a reference transaction creates, moves or deletes a branch under refs/heads whose name differs from the integration branch's name only in letter case, or begins with such a spelling and a slash, THEN the reference guard shall refuse it with a message that names the right spelling and the command to use · failure: none — nothing is written and no alias or second branch is ever created.
- REQ-4: IF a reference transaction makes a ref under refs/heads symbolic to the integration branch or to a case variant of its name, or makes the integration branch itself symbolic, THEN the reference guard shall refuse it · failure: none — a worktree's HEAD that points at the integration branch is not a ref under refs/heads and stays allowed.
- REQ-5: When git runs the hook, the hook text shall start the guard only at the prepared and committed states and refuse the transaction naming the cause when the moai program cannot be found or cannot run · failure: none — the hook never lets a branch transaction pass unchecked because the guard is missing, and the other states exit without starting any program.
- REQ-6: When the card land command moves the integration branch, the card land command shall write the intent record before the reference update and remove it when the call ends · failure: none — a record left by a stopped call is replaced by the next call, and a move the guard refuses is reported with the guard's own message and not as a moved tip.
- REQ-7: When the card land command has made every check of the close, the card land command shall write one ledger row that names the branch, the tip read, the candidate and the caller before it moves the branch · failure: none — a call that refuses writes no row, a rerun for the same card, candidate and tip writes no second row, and a close refused after the move leaves the row as the ledger record of the move.
- REQ-8: When the reference guard needs the integration branch's name, the reference guard shall read it through the one function the landing command uses · failure: none — a profile that names no integration branch leaves nothing to protect and the guard lets every transaction pass without a word.
- REQ-9: When a transaction that the intent record allowed has committed, the reference guard shall remove the record · failure: none — the same record cannot allow a second move, and a transaction that aborted leaves the record to the lander that wrote it.
- REQ-10: When the installer asks for the hook, the reference guard module shall provide the hook text as bytes of a POSIX shell script · failure: none — the text holds no spelling of any branch name and finds moai on the path and then in the user's moai folder.

## Acceptance

- AC-1 (REQ-1): in a temporary repository with the hook installed and a profile that names the integration branch, each of these leaves the branch where it was and exits non-zero: `update-ref`, `update-ref --stdin`, `branch -f`, `branch -D`, `branch -m`, `push . HEAD:<ref>`, `fetch . <ref>:<ref>`, a commit and a `reset --hard` in a worktree that has the branch checked out, and the same call with an environment variable typed by hand · verify: `cargo test --locked --offline --test card_guard_cli ac1_`
- AC-2 (REQ-1, REQ-9): the move that the intent record names passes and removes the record; a record with another new value, another old value, another ref, or a new value that is not a descendant of the current one refuses; the record of a finished move cannot allow it again, and no record allows a deletion; a creation with a record naming zero as the old value passes · verify: `cargo test --locked --offline --test card_guard_cli ac2_`
- AC-3 (REQ-2): `pack-refs --all`, `gc`, `pack-refs --all --no-prune` and a second `pack-refs` pass in the loose and in the packed state and from a linked worktree, the branch keeps its value, and `branch -D` and `update-ref -d` with no old value refuse in the loose-only and in the packed-only state · verify: `cargo test --locked --offline --test card_guard_cli ac3_`
- AC-4 (REQ-3): creating a branch whose name differs only in case, one whose first path part differs only in case, and moving or deleting such a branch that already exists, refuse in the packed state, with a message that holds the profile's spelling and the landing command, and a branch of any other name passes · verify: `cargo test --locked --offline --test card_guard_cli ac4_`
- AC-5 (REQ-4): `symbolic-ref` of a branch to the integration branch and to a case variant refuse, the integration branch made symbolic refuses, and a worktree that checks the branch out keeps working · verify: `cargo test --locked --offline --test card_guard_cli ac5_`
- AC-6 (REQ-5): a stand-in moai on the path records its calls: none at the preparing and aborted states and one each at prepared and committed; with no moai on the path a branch transaction refuses and the message names the cause; an ordinary commit on a lane branch passes with the guard running · verify: `cargo test --locked --offline --test card_guard_cli ac6_`
- AC-7 (REQ-6): `card land` under the guard moves the branch, the record exists before the update and is gone after, a record left by a stopped call is replaced, and a move the guard refuses (the record removed by a wrapper just before the update) answers with the guard's message and the branch and the ledger as they were · verify: `cargo test --locked --offline --test card_land_cli ac19_`
- AC-8 (REQ-7): after a landing the ledger holds the start row before the landing event, naming branch, tip, candidate and caller; a card in a state that cannot finish refuses with no row; a rerun after a stop between the move and the close writes no second row; and a close refused after the move leaves the row · verify: `cargo test --locked --offline --test card_land_cli ac20_`
- AC-9 (REQ-8): a profile that names another integration branch moves what the guard protects, and a profile with no integration branch lets every transaction pass · verify: `cargo test --locked --offline --test card_guard_cli ac9_`
- AC-10 (REQ-9, REQ-10): the hook text starts with a shell line, holds the guard's command and no branch name, is installed executable by the test and works as the installed hook in AC-1 to AC-6, and an aborted transaction leaves the record in place · verify: `cargo test --locked --offline --test card_guard_cli ac10_`
- AC-11 (REQ-7): when a lane takes the ready card between the checks and the close while the call carries a generation, the refusal after the move names the cause and the moved branch and does not say that nothing was changed, so the suffix strip of the landing command's message is pinned (the open non-blocking item of the first audit) · verify: `cargo test --locked --offline --test card_land_cli ac21_`
- AC-12 (REQ-1): the card operations document names the `card guard` action, its states, the intent record and the guard's refusals, and the help of `card guard` says that git runs it and that nobody types it · verify: `cargo test --locked --offline --test card_guard_cli ac12_`

## Scope

- write: `.moai/specs/SPEC-20261121-ref-guard/spec.md`
- write: `src/card/guard.rs`, `src/card/land.rs`, `src/card/landed.rs`, `src/card/mod.rs`, `src/ledger/workflow.rs`
- write: `tests/card_guard_cli.rs`, `tests/card_land_cli.rs`, `docs/design-review/appendix/operations/card.md`
- last step: the clap lines of the `card guard` action in `src/ops/card.rs` only (the subcommand and its dispatch arm), by the pair approval the other landing lines had, and the pair request that adds the hook as a managed name in the go-path-removal module when that module is on the tip.

## Decisions

- D-1: The guard's entry is an action of the card command, `card guard <state>`, read from the hook's arguments and standard input (decision D-1 of the flow skeleton: no new top-level word). It is run by git and by nothing else; the hook text is provided as bytes by the module.
- D-2: The guard decides on the value the ref has now, resolved by the guard from git, and never on the old value the hook reads: that value is zero when the caller named none.
- D-3: A transaction that deletes the integration branch passes as a prune only when the loose file exists, the packed store holds the same value as the resolved current value and the transaction names that value as its old value; every other delete is refused. Measured gap: after `git pack-refs --no-prune`, a delete that names the old value looks the same as the prune. It is listed under Risks and left to a later slice.
- D-4: The intent record is one JSON file in the common git folder, `moai/land-intent`, with `ref`, `old`, `new`, `by` and `at`, written by a temporary file and a rename, mode 0600, and single use. The `old` value is zero for a creation. A record carries no expiry: a stale record can allow only the one fast-forward the lander had checked, and the next call replaces it.
- D-5: Case variants are refused for create, move and delete, deletion included: on a case-insensitive file system deleting a variant's file deletes the branch's own. The names are compared by lower-casing both; a name that begins with a variant and a slash is refused for the same reason (a folder where the branch's file is). The message names the profile's spelling and the landing command.
- D-6: A symbolic update is read from the new value of the transaction line (`ref:<target>`, measured with git 2.54.0).
- D-7: When the hook cannot find or start moai it refuses the transaction: a guard that fails open is a guard that is off without a word. The cost is that a git client with no moai on its path cannot update a branch in this repository; the hook text therefore looks in the path and then in the user's moai folder, and the leader may reverse this choice.
- D-8: The start row is the free event kind `card.land_started` with branch, tip, candidate and caller, written in the one transaction that runs the checks of the close, so it refuses for the same reasons as the close and no schema or table changes. One row exists per card, candidate and tip; a rerun finds it and writes none.
- D-9: The order of a landing call becomes: checks, the start row, the intent record, the reference update, removal of the record, the close. A stop at any point leaves one of: nothing moved, a row without a move (the next call proceeds), or a move with its row (the next call closes).

## Plan

1. Tests first (RED): `tests/card_guard_cli.rs` with a real git, a hook installed from the module's text and a stand-in moai on the path, AC-1 to AC-6, AC-9, AC-10 and AC-12; `tests/card_land_cli.rs` gets AC-7, AC-8 and AC-11 (the next acceptance numbers of that file).
2. `src/card/guard.rs`: the pure judgement over the transaction lines, the resolved values and the record, the case and alias rules, the record file, the entry and the hook text; `src/card/landed.rs` gets the common-folder lookup through the git environment scrub.
3. `src/ledger/workflow.rs`: the start row written by the checks of the close; `src/card/land.rs`: the new order of D-9.
4. The card operations document and the help of the action; light checks, at most five mutants after the commit, one affected run.
5. Last: the clap lines of the action by pair approval, then the pair request on the managed-hook module, then the READY note with head and base.

## Risks

- A transaction that is a prune and a delete at once. After `git pack-refs --no-prune`, `git update-ref -d <ref> <value>` looks like the prune and deletes the branch (measured: the loose and the packed copy both go). Mitigations in this slice: the plain `branch -D`, `branch -m` and `update-ref -d` with no old value are refused; a restore from the `committed` state works (measured) and is the later slice.
- The hook runs twice per state when the loose and the packed store both change, and the guard runs on every branch transaction at two states. The hook text exits early at the other states and the cost of the two process starts per transaction is measured in the slice, not assumed.
- The profile is read from the worktree the hook runs in, so a branch whose profile names no integration branch leaves it unprotected while that branch is checked out. The slice reads the profile through the one function and tests the case, and does not claim a stronger rule.
- A test that creates a case variant cannot rely on the file system: a case-sensitive file system creates it, a case-insensitive one refuses a loose collision. The tests use the packed state, where the guard alone decides, and assert the message and the absence of the second branch.
- git before 2.28 has no `reference-transaction` hook, and the symbolic form was measured only with git 2.54.0; the installer's check of the git version is part of its own work.
- After the hook is installed the leader's own fast-forward merge of the integration branch in the primary worktree is refused. The leader lands with `card land`; the card operations document says so.

## Result

Not started.
