---
id: SPEC-20261121-card-finish
title: "card finish closes a landed or reported card in one step, and card next and card list stop offering a card that already landed"
status: draft
revision: 1
created: 2026-11-21
---

# card finish closes a landed or reported card in one step, and card next and card list stop offering a card that already landed

## Intent

A card whose change already landed on the integration branch can come back to `ready` with no reservation and open criteria (its holder gave an ended reservation back, or the session restarted). The next dispatch then hands the same work out a second time. Today the only way to close such a card is take, review, landed, criterion, done, and a card that has no commit (research, audit) has no way to close at all. Two changes fix this: a `card finish` command closes a ready, active or review card in one transaction when the landing is proven, or when a research or audit card names its report or verdict note as the evidence, and `card next` and `card list` mark a card that already landed so it is not offered again.

## Non-goals

- A schema change or a change of the lease rules: both belong to other work.
- A change of `card landed`, `card done` or `card take`.
- Closing a card that has a code change but no landing (a card still in review waits for its merge as today).
- Reading the SPEC at the integration tip instead of the project folder (D-9, a follow-up).

## Requirements

- REQ-1: When a leader runs finish on a code card whose landed head is an ancestor of the integration tip, the card finish command shall close the card in one transaction with one landing event and one finish event.
- REQ-2 (REQ-1): IF the head is not an ancestor of the integration tip or the profile names no integration branch, THEN the card finish command shall refuse and write nothing.
- REQ-3: When a card has open criteria, the card finish command shall close the card only if one met option names each open criterion.
- REQ-4 (REQ-3): IF an open criterion has no met option or a met option names a number that is not an open criterion or a number is named twice, THEN the card finish command shall refuse and write nothing.
- REQ-5: When another session holds a reservation whose lease has not ended, the card finish command shall refuse and leave the ledger unchanged · failure: none — a lease of the caller or a lease that has ended passes.
- REQ-6: When the card kind is research or audit, the card finish command shall accept an existing report file or a verdict note of the card as the evidence in place of a landed head.
- REQ-7 (REQ-6): IF the report path is not a regular file under the reports folder of the user level home or under the project folder or the note is not a report or decision note of the card, THEN the card finish command shall refuse and write nothing.
- REQ-8: When the card kind changes the repository, the card finish command shall accept only landing evidence and refuse report evidence · failure: none — the refusal writes nothing.
- REQ-9: When the by option is missing or blank, the card finish command shall refuse and write nothing · failure: none — the by value is stored in both events when it is given.
- REQ-10: When a ready card has a landing proof, the card next command shall list the card under skipped with the reason already_landed and the proof that found it.
- REQ-11 (REQ-10): IF the profile is not m-flow or the integration branch cannot be read, THEN the card next command shall say in its landing key that git proofs were not checked and use only the recorded proofs.
- REQ-12: When the card list command lists a card, it shall add a landed key that is true when a landing event or a merged pull request link is recorded and leave every existing key as it is · failure: none — the key reads the ledger only and runs no git command.

## Acceptance

- AC-1 (REQ-1, REQ-2): a ready card with two open criteria and a head inside the integration tip closes to `done` with events `card.landed` and `card.finish` and no held reservation, and a head outside the tip, an unknown head and a profile without an integration branch each refuse with every table unchanged · verify: `cargo test --locked --offline --test card_finish_cli ac1_`
- AC-2 (REQ-3, REQ-4): two open criteria with one met option refuse, with two succeed, with the same number twice or a number that is not open refuse, and a met reference of 500 characters passes and 501 refuses (N and N+1) · verify: `cargo test --locked --offline --test card_finish_cli ac2_`
- AC-3 (REQ-5): the gate passes a lease that ends at the moment of the call and refuses one that ends a second later (a pure function with an explicit now), a live lease of another session refuses, an ended lease of another session and a live lease of the caller pass · verify: `cargo test --locked --offline --test card_finish_cli ac3_`
- AC-4 (REQ-6, REQ-7): a research card closes with a report file under the reports folder, an audit card closes with a verdict note, and a missing file, a directory, a path outside both folders and a note of another kind each refuse with every table unchanged · verify: `cargo test --locked --offline --test card_finish_cli ac4_`
- AC-5 (REQ-8): a code card refuses a report path and a note, and a research card refuses a head · verify: `cargo test --locked --offline --test card_finish_cli ac5_`
- AC-6 (REQ-9): finish without by, and with a blank by, refuses, and the by value is stored in both events · verify: `cargo test --locked --offline --test card_finish_cli ac6_`
- AC-7 (REQ-2, REQ-4, REQ-5, REQ-7, REQ-9): every refusal leaves every ledger table byte equal to its dump before the call, with no fence_reject event · verify: `cargo test --locked --offline --test card_finish_cli ac7_`
- AC-8 (REQ-10, REQ-11): `card next` skips a ready card with a landed head inside the tip, with a merged pull request link and with a trailer of its linked SPEC reachable from the tip, keeps a card whose head is outside the tip, and reports `landing.checked` false with a reason for a profile that is not m-flow · verify: `cargo test --locked --offline --test card_finish_cli ac8_`
- AC-9 (REQ-12): `card list` keeps its keys and adds landed true for a card with a landing event or a merged link and false for the others · verify: `cargo test --locked --offline --test card_finish_cli ac9_`
- AC-10 (REQ-1, REQ-10): the scenario of a landed card: it lands, its holder gives the reservation back or its lease ends, a new session reads `card next`, the landed card is skipped and the next real card is listed, `card finish` closes the landed card with one met option per criterion, and the new session takes the next real card · verify: `cargo test --locked --offline --test card_finish_cli ac10_`
- AC-11 (REQ-1, REQ-10, REQ-12): the existing card tests that read the answer of `card next` and `card list` still pass, run whole · verify: `cargo test --locked --offline --no-fail-fast --test card_work_cli --test card_dispatch_cli --test card_generated_scope --test card_spec_cli --test card_done_mflow_cli --test card_reclaim_cli --test card_line_lease_cli --test ledger_workflow`
- AC-12 (REQ-1, REQ-6, REQ-10): the card operations document names the finish command, its two evidence kinds and the already_landed reason, and the help of the command says that it writes the ledger · verify: `cargo test --locked --offline --test card_finish_cli ac12_`

## Scope

- write: `.moai/specs/SPEC-20261121-card-finish/spec.md`
- write: `src/card/finish.rs`, `src/card/landed.rs`, `src/card/mod.rs`, `src/card/command.rs`, `src/card/work.rs`, `src/card/spec_link.rs`, `src/ledger/workflow.rs`, `src/spec_close.rs`
- write: `tests/card_finish_cli.rs`, `docs/design-review/appendix/operations/card.md`
- last step: one clap row for the finish command in `src/ops/card.rs`, added with the card line-lease command because another card holds that file for a long time

## Decisions

- D-1: No schema change. The done trigger only looks for a `card.landed` event, so a report close writes a `card.landed` event with no head and with the report or note and `no_code` in its payload. The real git proof stays in the card layer. Reversible: a later trigger change can replace the event kind.
- D-2: A new `workflow::finish` runs in one write transaction and does not use the fenced path, because a fence refusal commits a `fence_reject` event and a finish refusal must write nothing. Its checks all come before its first write.
- D-3: A ready card has no reservation, so finish creates none and bumps no generation. It walks the card ready, active, review, done by the rows of the transition table and marks a held reservation `released` when one exists.
- D-4: The live-lease gate is a public pure function with an explicit now, so the edge at the moment the lease ends is tested without freezing the database clock. A caller without a session id passes only when the lease has ended.
- D-5: Research and audit cards accept `--evidence-report PATH` or `--evidence-note REF`, exactly one of them, and refuse `--head`. Other kinds accept a head (or the newest recorded landed head) or a merged pull request link and refuse both evidence options. This follows the leader decision of 2026-11-21 and does not weaken code cards.
- D-6: The report path must resolve to a regular file under the reports folder of the user level home or under the project folder. The reports folder comes from the home, not from the project folder name, because a worktree name does not match it. The path is one line of at most 500 characters.
- D-7: `card next` decides `already_landed` only from proofs that name the card: the recorded landing head inside the tip, a merged pull request link, or the `Moai-Spec` trailer of the card's linked SPEC reachable from the tip, read in one `git log` pass with a named commit limit and a `truncated` flag. A card number in a commit message is never a proof (a text match is a hypothesis). A false positive costs one explicit `card take`, and the skip entry names the proof.
- D-8: `card list` shows only what is recorded (`landed` is a ledger fact, with no git call).
- D-9: The SPEC check at finish runs for code kinds only and reads the SPEC from the project folder given by `--project`. Reading the SPEC at the integration tip instead is a follow-up.
- D-10: `card landed` keeps accepting a review card only; finish is the one path that closes a ready or active card.

## Plan

1. Tests first (RED): `tests/card_finish_cli.rs` with the cases of AC-1 to AC-10 and AC-12 on a copy ledger and a temporary repository with an integration branch; see them fail on the unchanged code.
2. `src/ledger/workflow.rs`: the gate function, the evidence type, `finish` and a read-only function that lists the recorded landing facts of the non-done cards; two new error variants with sentences that `contract::classify` sorts as conflict and invalid input.
3. `src/card/finish.rs` (argument checks, secret screen, landing proof, kind rules, report path rule) and `src/card/landed.rs` (the tip, one trailer pass, one ancestor check per distinct head), then the `next` hunk in `src/card/work.rs` and the `landed` column in `src/card/command.rs`.
4. The card operations document and the help layout; run the card and ledger tests whole, the repository scans, `--lib`, `--bins`, clippy, then audit and write the Result.
5. Last, with the line lease: the clap row in `src/ops/card.rs`.

## Risks

- `src/card/work.rs` and `src/ops/card.rs` are shared files; the `next` hunk stays small and the clap row goes in through the line lease.
- The shape of the `card next` answer is pinned by several test files; the new `already_landed` entries and the `landing` key add to the answer and change no existing key.
- A SPEC shared by several cards makes the trailer proof hide siblings; D-7 names the proof in the skip entry.
- Word-scan tests over the repository may trip on a new sentence; the sentences avoid the words of `contract::classify` that sort an error earlier.

## Result

Every line below was read from a run on the working tree of commit `5c4952ce55` plus the staged clap row of `src/ops/card.rs`. That row is not committed yet (the file is reserved by another card; the row exists as a patch of 65 added lines), so the committed head alone does not pass the tests that call the command. Nothing here was run on a committed head that has the row.

- AC-1: met — `cargo test --locked --offline --test card_finish_cli ac1_` (4 tests) passed: a ready card with two open criteria and a head inside the integration tip closes to `done` with the events `card.landed` and `card.finish` and no held reservation; a head outside the tip, an unknown head and a profile without an integration branch each refuse with the reason named and every table unchanged.
- AC-2: met — `ac2_` (2 tests) passed: one met option for two open criteria refuses, two succeed, a repeated number and a number that is not open refuse, a reference of 500 characters passes and 501 refuses.
- AC-3: met — `ac3_` (4 tests) passed: a live lease of another session refuses naming the holder, an ended lease of another session and a live lease of the caller pass, and the pure gate passes at `lease_until == now` and refuses at `now + 1`; a call without a session id passes only when the lease has ended.
- AC-4: met — `ac4_` (3 tests) passed: a research card closes with a report file under the reports folder, an audit card closes with a verdict note, a missing file, a directory, a path outside both folders, a `..` escape, a symlink out and a note of another kind each refuse; a typed path of 500 characters passes and 501 refuses.
- AC-5: met — `ac5_` (1 test) passed: a code card refuses a report path and a note, and a research card refuses a head.
- AC-6: met — `ac6_` (1 test) passed: finish without `--by` and with a blank `--by` refuses, and the by value is stored in both events.
- AC-7: met — `ac7_` (1 test) passed, and every refusal test of AC-1 to AC-6 compares the dump of all seven ledger tables before and after and checks that no `fence_reject` event was written.
- AC-8: met — `ac8_` (4 tests) passed: `card next` skips a ready card by a recorded head inside the tip, by a merged pull request link and by the `Moai-Spec` trailer of its linked SPEC, keeps a card whose head is outside the tip in the same test, and reports `landing.checked` false with a reason for a profile that is not m-flow; the trailer test also keeps a card whose number appears in a commit message but that has no trailer (a card number in a message is never a proof).
- AC-9: met — `ac9_` (1 test) passed: `card list` keeps its keys and adds `landed`.
- AC-10: met — `ac10_` (2 tests) passed: the scenario with the reservation given back and the one with the lease ended by SQL; in the second, the card is held and so is absent from `ready`, and the skip reason `already_landed` is shown only for the given-back card (a card that still holds a reservation is skipped as `held` before the landing check).
- AC-11: met — `cargo test --locked --offline --no-fail-fast --test card_work_cli --test card_dispatch_cli --test card_generated_scope --test card_spec_cli --test card_done_mflow_cli --test card_reclaim_cli --test card_line_lease_cli --test ledger_workflow --test card_finish_cli` exited 0: 26, 25, 8, 11, 4, 16, 16, 29 and 24 tests passed.
- AC-12: met — `ac12_` (1 test) passed: the help of `card finish` starts with `Writes the ledger.` and the card operations document names `card finish`, `already_landed`, `--evidence-report` and `--evidence-note`.

Mutants (all killed, the files restored byte for byte, `git diff --stat` empty afterwards):

| Mutant | Killed by |
| --- | --- |
| lease end read as `lease_until < now` | the gate edge test and the no-session test of AC-3 |
| a live lease of the caller no longer passes | the gate edge test and the caller test of AC-3 |
| every probed head counts as inside the tip | the AC-8 test that keeps a card with a head outside the tip |
| the trailer pass limit read as `>` instead of `>=` | the unit test of the limit boundary in `landed.rs` |
| the one-line limit 501 instead of 500 | the 500/501 tests of AC-2 and AC-4 |

Known and left as it is: `card done` reads the newest `card.landed` event and decodes `head` as text, so a newest event without a head (the report close writes one) makes `card done` refuse with "no local landing is linked". A report-closed card is already `done` and nothing reopens it, so no command reaches this today; the function is not changed here. The stored report path is the resolved path and may pass 500 characters (`/var` resolves to `/private/var` on macOS); the 500-character limit applies to the typed path. Not run on this tree: clippy, `--lib` and `--bins` as a whole, the repository scans, the python tests and a Windows build; the unit tests of `landed.rs` ran only inside the mutant run of the limit boundary.
