---
id: SPEC-20261121-card-finish
title: "card finish closes a landed or reported card in one step, and card next and card list stop offering a card that already landed"
status: draft
revision: 5
created: 2026-11-21
---

# card finish closes a landed or reported card in one step, and card next and card list stop offering a card that already landed

## Intent

A card whose change already landed on the integration branch can come back to `ready` with no reservation and open criteria (its holder gave an ended reservation back, or the session restarted). The next dispatch then hands the same work out a second time. Today the only way to close such a card is take, review, landed, criterion, done, and a card that has no commit (research, audit) has no way to close at all. Two changes fix this: a `card finish` command closes a ready, active or review card in one transaction when the landing is proven, or when a research or audit card names its report or verdict note as the evidence, and `card next` and `card list` mark a card that already landed so it is not offered again.

Revision 2 (leader decision of 2026-11-21, user order to repair card closing): git and the ledger are not one transaction, so a call has to be safe to repeat and to race. A second call on a closed card is refused and writes nothing, two callers close it once, a stale generation is refused, and a landing that cannot close the card yet (an open criterion without evidence) is still recorded and the answer says what is open. A landed card that is given back is answered as landed and names the finish command.

## Non-goals

- A schema change or a change of the lease rules: both belong to other work.
- A change of `card landed`, `card done` or `card take`; `card release` only adds the landed answer of REQ-12 and keeps its effect.
- Writing the real ledger from a test: every test runs with a temporary home.
- Closing a card that has a code change but no landing (a card still in review waits for its merge as today).
- Reading the SPEC of a card whose evidence is a merged pull request link off m-flow at any place but the project folder (D-9).

## Requirements

- REQ-1: When a leader runs finish on a code card whose landed head is an ancestor of the integration tip, the card finish command shall close the card in one transaction and record the head, the tip, the generation and the caller in one landing event and one finish event · failure: none — a head outside the tip, an unknown head, a profile without an integration branch, or on an m-flow profile a merged pull request link without a proven landing refuses and writes nothing.
- REQ-2: When every open criterion of a card has one met option, the card finish command shall mark each of them met and close the card · failure: none — a met option that names a number that is not open or names a number twice refuses and writes nothing.
- REQ-3: When an open criterion of a code card has no met option, the card finish command shall record the proven landing once for the leader or the holder and answer which criteria are open without any other change · failure: none — a head that git does not prove, a landing that is not recorded, or a caller that is neither the leader nor the holder refuses and writes nothing.
- REQ-4: When another session holds a reservation whose lease has not ended and the card is not in review with a landing that git proves for the leader word or the holder session, the card finish command shall refuse and leave the ledger unchanged · failure: none — a lease of the caller, a lease that has ended, or a review card whose head git proves inside the tip passes when `--by leader` is given or the caller is the holder session (the holder's reservation is released in the same transaction).
- REQ-5: When the card kind is research or audit, the card finish command shall accept an existing report file or a verdict note of the card as the evidence in place of a landed head · failure: none — a path that is not a regular file under the reports folder or the project folder, a note that is not a report or decision note, or an open criterion without a met option refuses and writes nothing.
- REQ-6: When the card kind changes the repository, the card finish command shall accept only landing evidence and refuse report evidence · failure: none — the refusal writes nothing.
- REQ-7: When the by option is missing or blank, the card finish command shall refuse and write nothing · failure: none — the by value is stored in both events when it is given.
- REQ-8: When finish is called for a card that is already done, the card finish command shall refuse naming that state and write no second landing, finish or release event · failure: none — a retry after a call that wrote nothing closes the card once.
- REQ-9: When two finish calls reach one card at the same time, the card finish command shall close it exactly once and refuse the other call · failure: none — the checks and the writes of one call share one write transaction.
- REQ-10: When the generation option is given and differs from the generation of the card's reservation, the card finish command shall refuse and write nothing, not even a fence_reject event · failure: none — a call without the option is not checked.
- REQ-11: When a ready card has a landing proof, the card next command and the card list command shall mark it as already landed, skipped with the reason already_landed and landed true · failure: none — a profile that is not m-flow or an integration branch that cannot be read says in the landing key that git proofs were not checked and uses only the recorded proofs.
- REQ-12: When a card with a recorded landing is released, the card release command shall answer that the card is landed and name the finish command · failure: none — the answer reads the ledger only and runs no git command.

## Acceptance

- AC-1 (REQ-1): a ready card with two open criteria and a head inside the integration tip closes to `done` with events `card.landed` and `card.finish` and no held reservation, the landing event carries the head, the tip, the generation and the caller, a head outside the tip, an unknown head and a profile without an integration branch each refuse with every table unchanged, and a head that was fast-forwarded before any ledger write closes from the same evidence · verify: `cargo test --locked --offline --test card_finish_cli ac1_`
- AC-2 (REQ-2): two open criteria with two met options succeed, the same number twice or a number that is not open refuses, and a met reference of 500 characters passes and 501 refuses (N and N+1) · verify: `cargo test --locked --offline --test card_finish_cli ac2_`
- AC-3 (REQ-4): the gate passes a lease that ends at the moment of the call and refuses one that ends a second later (a pure function with an explicit now), a live lease of another session refuses, an ended lease of another session and a live lease of the caller pass · verify: `cargo test --locked --offline --test card_finish_cli ac3_`
- AC-4 (REQ-5): a research card closes with a report file under the reports folder, an audit card closes with a verdict note, and a missing file, a directory, a path outside both folders, a note of another kind and an open criterion without a met option each refuse with every table unchanged · verify: `cargo test --locked --offline --test card_finish_cli ac4_`
- AC-5 (REQ-6): a code card refuses a report path and a note, and a research card refuses a head · verify: `cargo test --locked --offline --test card_finish_cli ac5_`
- AC-6 (REQ-7): finish without by, and with a blank by, refuses, and the by value is stored in both events · verify: `cargo test --locked --offline --test card_finish_cli ac6_`
- AC-7 (REQ-1, REQ-2, REQ-4, REQ-5, REQ-7, REQ-8, REQ-10): every refusal leaves every ledger table byte equal to its dump before the call, with no fence_reject event · verify: `cargo test --locked --offline --test card_finish_cli ac7_`
- AC-8 (REQ-11): `card next` skips a ready card with a landed head inside the tip, and with a merged pull request link, keeps a card whose head is outside the tip and a card whose linked SPEC has a trailer commit on the tip (the trailer is a hint field `spec_trailer_hints` of the landing key, never the skip reason), and reports `landing.checked` false with a reason for a profile that is not m-flow · verify: `cargo test --locked --offline --test card_finish_cli ac8_`
- AC-9 (REQ-11): `card list` keeps its keys and adds landed true for a card with a landing event or a merged link and false for the others · verify: `cargo test --locked --offline --test card_finish_cli ac9_`
- AC-10 (REQ-1, REQ-11): the scenario of a landed card: it lands, its holder gives the reservation back or its lease ends, a new session reads `card next`, the landed card is skipped and the next real card is listed, `card finish` closes the landed card with one met option per criterion, and the new session takes the next real card · verify: `cargo test --locked --offline --test card_finish_cli ac10_`
- AC-11 (REQ-1, REQ-11, REQ-12): the existing card tests that read the answer of `card next`, `card list` and `card release` still pass, run whole · verify: `cargo test --locked --offline --no-fail-fast --test card_work_cli --test card_dispatch_cli --test card_generated_scope --test card_spec_cli --test card_done_mflow_cli --test card_reclaim_cli --test card_line_lease_cli --test ledger_workflow`
- AC-12 (REQ-1, REQ-5, REQ-11): the card operations document names the finish command, its two evidence kinds, the landing-only answer and the already_landed reason, and the help of the command says that it writes the ledger · verify: `cargo test --locked --offline --test card_finish_cli ac12_`
- AC-13 (REQ-3): a code card with two open criteria and one met option records one landing event, keeps its state, its criteria, its reservation and its generation, answers closed false and the open number, and a second identical call records no second landing event; a card with no proven head refuses with every table unchanged · verify: `cargo test --locked --offline --test card_finish_cli ac13_`
- AC-14 (REQ-8): a second finish on a done card refuses naming `done` with every table unchanged and no second done or release event, and a call that failed before any write is followed by one call that closes the card once · verify: `cargo test --locked --offline --test card_finish_cli ac14_`
- AC-15 (REQ-9): two processes that call finish on one card at once give exactly one success and one refusal, one done event and one release event · verify: `cargo test --locked --offline --test card_finish_cli ac15_`
- AC-16 (REQ-10): a generation lower or higher than the card's refuses with every table unchanged and no fence_reject event, the current generation passes, and a call without the option passes · verify: `cargo test --locked --offline --test card_finish_cli ac16_`
- AC-17 (REQ-12): releasing a card that has a landing event answers landed true and names `card finish`, releasing one without a landing keeps the old answer, and no git command runs · verify: `cargo test --locked --offline --test card_finish_cli ac17_`
- AC-18 (REQ-1): on an m-flow profile a merged pull request link whose merge commit is not in the repository refuses with every table unchanged and the reason that a merged pull request link alone is not enough, the same card closes with a proven head, and on a github-flow profile the link alone still closes · verify: `cargo test --locked --offline --test card_finish_cli ac18_`
- AC-19 (REQ-1, REQ-2): a code card whose SPEC fails the check refuses with every table unchanged, a card with no met criterion and no met option refuses, and a card with one criterion met earlier closes with no met option · verify: `cargo test --locked --offline --test card_finish_cli ac19_`
- AC-20 (REQ-3, REQ-11): a landing-only call by a session that is neither the leader nor the holder refuses with every table unchanged, the leader (`--by leader`) and the holder pass, and `card next` does not hide a card whose recorded head is not on the tip · verify: `cargo test --locked --offline --test card_finish_cli ac20_`
- AC-21 (REQ-1, REQ-11): a tag named like the integration branch and pointing elsewhere changes no proof: a head on the branch closes the card, a head only on the tag refuses, and `card next` reads the branch · verify: `cargo test --locked --offline --test card_finish_cli ac21_`
- AC-22 (REQ-4): a review card held by another session with a live lease closes for the leader (`--by leader`) or the holder session on a head that git proves inside the tip and the holder's reservation is released in the same call; an active card held by another live session still refuses with every table unchanged; a review card whose head git does not prove still refuses · verify: `cargo test --locked --offline --test card_finish_cli ac22_`
- AC-23 (REQ-11): a two-slice card whose first slice carries its SPEC trailer on the tip stays in `ready` of `card next` (the hint names the trailer) and its release answer does not name `card finish` · verify: `cargo test --locked --offline --test card_finish_cli ac23_`
- AC-24 (REQ-1): the SPEC of a card is read from the proven tip, not from the working tree: a SPEC that passes on the tip closes the card while the working-tree copy is damaged, and a SPEC that fails on the tip refuses while the working-tree copy passes · verify: `cargo test --locked --offline --test card_finish_cli ac24_`
- AC-25 (REQ-4): a review card held with a live lease refuses non-landing evidence (a report file or a note of a research card), refuses a third session that gives neither `--by leader` nor the holder session even with a head git proves, and closes for either of those two with the reservation released; every refusal leaves every table unchanged · verify: `cargo test --locked --offline --test card_finish_cli ac25_`
- AC-26 (REQ-3): the owner of a landing-only call is the leader word or the holder session of the card's reservation: a session that is not the holder refuses on a review card held live and on an active card whose lease ended, and the holder passes · verify: `cargo test --locked --offline --test card_finish_cli ac26_`
- AC-27 (REQ-1): the integration branch of the profile is accepted as a branch name only: a name that starts with `refs/heads/` or has no `refs/` prefix is read as a branch, any other `refs/` name (a tag) refuses with every table unchanged · verify: `cargo test --locked --offline --test card_finish_cli ac27_`

## Scope

- write: `.moai/specs/SPEC-20261121-card-finish/spec.md`
- write: `src/card/finish.rs`, `src/card/landed.rs`, `src/card/mod.rs`, `src/card/command.rs`, `src/card/work.rs`, `src/card/dispatch.rs`, `src/card/spec_link.rs`, `src/ledger/workflow.rs`, `src/spec_close.rs`
- write: `tests/card_finish_cli.rs`, `docs/design-review/appendix/operations/card.md`
- last step: the clap lines of the finish command in `src/ops/card.rs` only (the subcommand and its dispatch arm, with the generation option of REQ-10); another card holds that file and the leader approved these lines as a pair

## Decisions

- D-1: No schema change. The done trigger only looks for a `card.landed` event, so a report close writes a `card.landed` event with no head and with the report or note and `no_code` in its payload. The real git proof stays in the card layer. Reversible: a later trigger change can replace the event kind.
- D-2: A new `workflow::finish` runs in one write transaction and does not use the fenced path, because a fence refusal commits a `fence_reject` event and a finish refusal must write nothing. Its checks all come before its first write.
- D-3: A ready card has no reservation, so finish creates none and bumps no generation. It walks the card ready, active, review, done by the rows of the transition table and marks a held reservation `released` when one exists.
- D-4: The live-lease gate is a public pure function with an explicit now, so the edge at the moment the lease ends is tested without freezing the database clock. A caller without a session id passes only when the lease has ended.
- D-5: Research and audit cards accept `--evidence-report PATH` or `--evidence-note REF`, exactly one of them, and refuse `--head`. Other kinds accept a head (or the newest recorded landed head) or a merged pull request link and refuse both evidence options. This follows the leader decision of 2026-11-21 and does not weaken code cards.
- D-6: The report path must resolve to a regular file under the reports folder of the user level home or under the project folder. The reports folder comes from the home, not from the project folder name, because a worktree name does not match it. The path is one line of at most 500 characters.
- D-7: `card next` decides `already_landed` only from proofs that name the card: the recorded landing head inside the tip or a merged pull request link. The `Moai-Spec` trailer of the card's linked SPEC reachable from the tip is only a hint (`landing.spec_trailer_hints`, read in one `git log` pass with a named commit limit and a `truncated` flag), because one SPEC with several slices has a trailer after its first slice. A card number in a commit message is never a proof (a text match is a hypothesis). The skip entry names the proof.
- D-8: `card list` shows only what is recorded (`landed` is a ledger fact, with no git call).
- D-9: The SPEC check at finish runs for code kinds only. When git proved the landing it reads the SPEC from the integration branch tip it proved the landing against (`git show`, AC-24), so a checkout that was not moved after the fast-forward cannot make it read an old copy; a card whose evidence is a merged pull request link off m-flow has no tip and reads the project folder given by `--project`.
- D-10: `card landed` keeps accepting a review card only; finish is the one path that closes a ready or active card.
- D-11: Open criteria without a met option are a landing-only call for a code card (leader decision): the landing event is recorded once, nothing else changes, and the answer is `ok` with `closed` false and the open numbers. A research or audit card keeps the refusal, because its evidence is a report or a verdict and a landing event without a head is not a landing. Reversible: the call can refuse again.
- D-12: `--generation` is compared with the generation of the card's newest reservation (a ready card keeps the number of the reservation it had); a mismatch refuses before any write and never goes through the fenced path, so no fence_reject event is written. A card that never had a reservation has generation 0.
- D-13: The landing event records `head`, `tip`, `generation` and `by` (the tip is the one that contained the head when the call ran); the finish event keeps the criteria. The evidence is checked in git before the transaction, and the transaction re-reads the card state, so git and the ledger never need one transaction: a landing proven in git and not yet written is simply proven again by the next call.
- D-14: The release answer adds `landed` and `next_step` only when the card has a `card.landed` event or a merged pull request link, read from the ledger; every existing key and the effect of release stay as they are.

## Plan

1. Revision 1 is built and committed on this branch (the clap lines included, merged with the tip); revision 2 starts from it.
2. Tests first (RED): AC-13 to AC-17 and the additions to AC-1, AC-4 and AC-7 in `tests/card_finish_cli.rs`, on a temporary home and a temporary repository; the old tests that expected a refusal for an open criterion change to the new answer. See them fail on the unchanged code.
3. `src/ledger/workflow.rs` (landing event payload, the landing-only outcome, the generation check, the done refusal kept) and `src/card/finish.rs` (the `--generation` argument, the answer), the clap lines, the release answer in the card layer, then the card operations document and the help.
4. Light checks first, mutants (at most five, after the commit), one affected run, then the READY note; the canonical text of the closing path is slice 2 after the READY.

## Risks

- `src/card/work.rs` and `src/ops/card.rs` are shared files; the `next` hunk stays small and the clap row goes in through the line lease.
- The shape of the `card next` answer is pinned by several test files; the new `already_landed` entries and the `landing` key add to the answer and change no existing key.
- A SPEC with several slices, or shared by several cards, has a trailer on the tip before its last card is done; D-7 therefore makes the trailer a hint and never a proof.
- Word-scan tests over the repository may trip on a new sentence; the sentences avoid the words of `contract::classify` that sort an error earlier.

## Result

Slice 1 of revision 5 (the code, the clap lines and the tests, with the repairs of the first audit, of the failure simulation and of the re-audit follow-ups; the first audit and the re-audit of `f566ebf268` both ended and that head is on the local integration branch). The canonical text of the closing path is slice 2 and is not part of this result.

The numbers of AC-1 to AC-17 are from one run of `moai check affected 6e1bedb4f3a4..74bf6302cfda` (81 integration targets, the whole lib and both bins). The numbers of AC-18 to AC-24 are from `moai check affected 6e1bedb4f3a4..9155159f137e` (the same 81 targets; logs in the build folder `check-affected-9155159f137e`) and the runs named. The numbers of AC-25 to AC-27 and the final whole-file count are from the runs named below. The tests use a temporary home and never the real ledger.

- AC-1 met: `ac1_` (6 tests) passed: the landing event carries `head`, `tip`, `generation` and `by`; a head already on the integration branch with no landing event in the ledger (the fast-forward done, the ledger write not) closes the card from the same evidence; a head outside the tip, an unknown head and a profile without an integration branch refuse with every table unchanged.
- AC-2 met: `ac2_` (2 tests) passed: two met options close two open criteria, a number twice and a number that is not open refuse, a reference of 500 characters passes and 501 refuses.
- AC-3 met: `ac3_` (4 tests) passed, unchanged from the first revision.
- AC-4 met: `ac4_` (4 tests) passed: the report and verdict closes, the refusals, and a research card with an open criterion and no met option refuses with every table unchanged.
- AC-5 met: `ac5_` (1 test) passed. AC-6 met: `ac6_` (1 test) passed.
- AC-7 met: `ac7_` (1 test) passed: every refusal, including the closed card and the stale generation, leaves all seven tables byte equal and writes no `fence_reject` event.
- AC-8 met: `ac8_` (4 tests) passed; the trailer test now shows the trailer only as `landing.spec_trailer_hints`. AC-9 met: `ac9_` (1 test) passed. AC-10 met: `ac10_` (2 tests) passed.
- AC-11 met: in the same affected run `card_work_cli` 26, `card_dispatch_cli` 25, `card_generated_scope` 8, `card_spec_cli` 11, `card_done_mflow_cli` 24, `card_reclaim_cli` 21, `card_line_lease_cli` 16 and `ledger_workflow` 29 tests passed, none failed.
- AC-12 met: `ac12_` (1 test) passed: the help names `--generation` and `open_criteria`, and the card operations document names the finish command, both evidence kinds, the landing-only answer and `already_landed`.
- AC-13 met: `ac13_` (2 tests) passed: one landing event, state, criteria, reservation and generation unchanged, the answer is `closed` false with the open numbers, and a repeat records no second landing event.
- AC-14 met: `ac14_` (2 tests) passed: a second finish on a done card refuses naming `done` with every table unchanged, and a refused call followed by a correct one closes the card with one finish event.
- AC-15 met: `ac15_` (2 tests) passed: two writers on one ledger file behind a barrier and two processes, 20 rounds each, give one success, one refusal and one finish event. These two tests already passed before the change, because the write transaction takes the write lock first; they pin it.
- AC-16 met: `ac16_` (1 test) passed: a lower and a higher generation refuse with no `fence_reject` event, the current one and an absent option pass.
- AC-18 met: `ac18_` (1 test) passed: on an m-flow profile a merged link alone refuses with the reason that the link is not enough and every table unchanged, the same card closes on a proven head, and off m-flow the link alone closes via `merged_pr` with the finish event only. RED before the repair.
- AC-19 met: `ac19_` (1 test) passed: a SPEC that the tip holds with errors refuses with its error count (the SPEC is on the tip, not absent, and the checkout copy is good), a card with its one criterion withdrawn and nothing named refuses with the words of the finish guard (`card finish refused: …`, not the schema trigger's), and a card with a criterion met earlier closes with no met option.
- AC-20 met: `ac20_` (1 test) passed: a landing-only call by a third session refuses with every table unchanged and hides nothing from `card next`; `--by leader` and the holder pass; a recorded head outside the tip keeps its card offered. RED before the repair.
- AC-21 met: `ac21_` (1 test) passed: with a tag named like the integration branch pointing at another commit, a head only on the tag refuses, the branch tip closes the card, and `card next` proves a landed card against the branch. RED before the repair.
- AC-22 met: `ac22_` (1 test) passed: a review card whose lease another session holds closes for the leader (`--by leader`) on a proven head and its reservation is `released` in the same call; an active card of another live session refuses with `is held by`; a review card with an unproven head refuses. RED before the repair.
- AC-23 met: `ac23_` (1 test) and the rewritten trailer test of `ac8_` passed: a two-slice card stays offered with its trailer reported as `landing.spec_trailer_hints`, and its release answer names no finish step. RED before the repair.
- AC-24 met: `ac24_` (1 test) passed: the SPEC is read from the proven tip: the good one closes while the checkout holds a damaged copy, the damaged one refuses while the checkout holds a good copy. RED before the repair.
- AC-25 met: `ac25_` (1 test) passed: a review card held live refuses a report file (a research card) and a third session with `--by agent-b` (reason `is held by`), every table unchanged; the holder session and `--by leader` close their cards and both reservations are `released`. RED before the repair (the third session closed it).
- AC-26 met: `ac26_` (1 test) passed: a landing-only call by a session that is not the holder refuses on a review card held live (`is held by`) and on an active card whose lease ended (the leader-or-holder reason), and the holder passes on the one whose lease ended. RED before the repair.
- AC-27 met: `ac27_` (1 test) passed: a profile naming `refs/tags/integ` refuses (`integration branch`, every table unchanged) while `refs/heads/integ` closes the card with the tip of the branch. RED before the repair.
- AC-17 met: `ac17_` (1 test) passed: releasing a card with a landing event or a merged link answers `landed` true and names `card finish`, a plain card keeps its old keys, and a stand-in for git that fails when called is never called.

Mutants of the first run (all killed, the files restored byte for byte, the tree clean afterwards):

| Mutant | Killed by |
| --- | --- |
| a lower generation passes the check | `ac16_` |
| an open criterion refuses again for a landing | `ac13_` |
| a repeated head records a second landing event | `ac13_` |
| the release answer misses the landing event | `ac17_` |
| the landing event records generation 0 | `ac1_` and `ac13_` |

Mutants of the repairs (restored byte for byte, tree clean afterwards): R1 the merged link is never taken, R2 the m-flow guard of the link is removed and R5 the owner of a landing-only call is not checked were killed by `ac18_` and `ac20_`; R3 the SPEC check gets no SPEC was killed by `ac19_`; S1 an active card also passes the lease gate was killed by `ac22_` and `ac3_`; S2 the trailer hint is not recorded was killed by `ac8_`; S3 the SPEC is read from the checkout again was killed by `ac24_`. R4, the code guard of nothing met (`== 0` to `< 0`), survived: it is an equivalent mutant, because the schema trigger of the done transition (`src/ledger/schema.rs:183`) refuses the same card with the same words inside the same transaction, which `ac19_` pins. The tag fix is shown by `ac21_` failing before and passing after. Mutants of the re-audit follow-ups, all killed (restored byte for byte, tree clean): F1 the review exception without the landing condition (`ac25_`), F2 any session counts as the holder of a landing-only call (`ac26_`), N2 any caller passes the review exception (`ac25_`, `ac26_`), N3 any `refs/` name is read as it stands (`ac27_`) and R4 again, which `ac19_` now kills because it asserts the finish wording of the guard, so R4 is no longer an equivalent mutant of the test.

Other results of this stage: `clippy` with all targets, `fmt`, the operations `rustfmt` and `diff --check` exited 0 in the affected run. The Python step of the first run failed once (five fake card scopes under `docs/…`, renamed to `study/…`); the Python step of the second run ran 185 tests, all OK. The second run had two failures of the new tests, both repaired in the last test commit: clippy `needless_borrows_for_generic_args` and the fixture lint `git_fixture_calls_live_in_the_shared_helper` (a direct git spawn). After that commit `cargo clippy --all-targets -- -D warnings` exited 0 and `card_finish_cli` (42 tests) and `ac_command_lint` (5 tests) passed. The whole affected run was not repeated after that commit; its other 80 targets, the lib and the bins had passed. The follow-up commits were checked by `moai check affected f566ebf268..<head>` (named in the READY note).

Choices a reviewer may reverse: off m-flow, a merged pull request link as the evidence with an open criterion and no met option answers `ok` with `closed` false and no event written; the generation check runs before the lease gate; the lease gate is passed only by a review card with a landing git proved (not by a merged link); the SPEC is read from the tip only when git proved the landing, and a merged link off m-flow still reads the checkout.

Scope notes: the release answer is built in `src/card/dispatch.rs`, which was added to this card's scope; `src/ops/card.rs` changed only in the lines of the finish subcommand and its dispatch arm, under the leader's pair approval.

Not run: the independent audit of the follow-up commits, the Windows type check, `tools/self_check.sh`, a whole affected run after the last test commit, and any test against the real ledger. Left to the second SPEC: a landing newer than the last move into review, any recorded head as the proof, `card done`'s own branch ref, a review card whose lease ended in `card next`, a finish with `--head` and a merged link off m-flow, and the query of cards whose head is on the integration branch with no landing record.
