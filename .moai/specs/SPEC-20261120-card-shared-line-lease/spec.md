---
id: SPEC-20261120-card-shared-line-lease
title: "a card gives back the paths it no longer writes at review, and a line lease adds one shared line without a long reservation"
status: draft
revision: 1
created: 2026-11-20
---

# a card gives back the paths it no longer writes at review, and a line lease adds one shared line without a long reservation

## Intent

Lanes wait on each other for files in which each needs only one line (`pub mod` lines of `src/lib.rs`, rows of the help table, a shared test table), because a card keeps every path of its write scope until it is done, and a scope only grows. Two changes remove the wait: a card gives back the paths it no longer writes when it goes to review, and a short line lease takes a shared file, runs the one command that commits the line, and gives the file back in the same call, so the main reservation never holds it. Leader merges still resolve adjacent-line conflicts in these spots.

## Requirements

- REQ-1: When a holder sends a card to review, the card review command shall give back every path of the reservation that no `--keep` names.
- REQ-2 (REQ-1): IF a path named with `--keep` is not in the reservation, THEN the card review command shall refuse and write nothing.
- REQ-3: When a holder gives paths back, the card give-back command shall remove the named paths from the reservation of that generation and record them in one event.
- REQ-4 (REQ-3): IF a named path is not in the reservation or the caller does not hold the generation, THEN the card give-back command shall refuse and change no reservation, scope or card.
- REQ-5: When a holder runs the card line-lease command, the command shall add the named paths to the reservation, run the given command and give the added paths back in one call.
- REQ-6 (REQ-5): IF another card's reservation covers a named path, THEN the card line-lease command shall retry until the wait limit and then refuse with that card named and write nothing.
- REQ-7 (REQ-5): IF the given command exits non-zero or cannot start, THEN the card line-lease command shall still give the added paths back and refuse with that exit status named.
- REQ-8: The help of the three commands and the card operations document shall state that a path given back is free for any card at once and that a reopened card widens its scope again with a reason.

## Acceptance

- AC-1 (REQ-1, REQ-2): after `card review` the reservation lists only the kept paths, another card can take a given-back path at once, and a `--keep` path the card does not hold is refused · verify: `cargo test --locked --offline --test card_line_lease_cli review`
- AC-2 (REQ-3, REQ-4): `card give-back` removes the named paths, writes one `card.scope` event that names them and refuses an unknown path and a stale generation · verify: `cargo test --locked --offline --test card_line_lease_cli give_back`
- AC-3 (REQ-5, REQ-6, REQ-7): two cards each add one line to the same file with `card line-lease`, one after the other and while the first still runs, and neither reservation lists the file afterwards; a failing command gives the path back and its status is named in the refusal · verify: `cargo test --locked --offline --test card_line_lease_cli line_lease`
- AC-4 (REQ-8): the help of `card review`, `card give-back` and `card line-lease` and the card operations document state the rules · verify: `cargo test --locked --offline --test card_line_lease_cli help`
- AC-5 (REQ-1, REQ-3, REQ-5): the existing card and ledger tests that read a reservation still pass · verify: `cargo test --locked --offline --no-fail-fast --test card_work_cli --test card_scope_widen_cli --test card_dispatch_cli --test ledger_reservation --test ledger_workflow --test prop_reservation --test card_reclaim_cli --test claim_cli --test mode_cards_53 --test green_test_lock --test ledger_scope_return --test card_line_lease_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-card-shared-line-lease/spec.md`
- write: `src/ledger/scope_return.rs`, `src/ledger/mod.rs`, `src/card/work.rs`, `src/ops/card.rs`
- write: `tests/card_line_lease_cli.rs`, `tests/ledger_scope_return.rs`, and one helper line in `tests/card_reclaim_cli.rs` (its review helper names `--keep` because a review now gives paths back)
- write: `docs/design-review/appendix/operations/card.md`

## Decisions

- D-1: A give-back deletes the `reservation_scope` rows and writes the paths into an event. The rows of `reservation`, `card`, notes and events are never deleted (triggers); the scope rows carry no such trigger, and the history of every scope stays in `card.scope` events (`from`, `to`, the paths given back). A new column or table would need a schema migration, which this change does not make. Reversible: a later change may keep the rows with a `released_at`.
- D-2: Three separate verbs (`review --keep`, `give-back`, `line-lease`) instead of options on `card scope`, whose answer and refusals already say that a held scope only grows.
- D-3: `card review` gives back every path unless `--keep` names it. A card sent back with `card reopen` widens its scope again with `card scope --reason`, as it does today.
- D-4: A rejected fenced call (a stale generation, or a session that is not the holder) leaves the `fence_reject` event the ledger writes for every such call, `review` included; REQ-4 therefore says no reservation, scope or card changes, not that no event is written. An unknown path, an overlap and a failing command are not fence rejections and write nothing of that kind.

## Plan

1. Tests first (RED): `tests/card_line_lease_cli.rs` with the review, give-back, line-lease and help cases of AC-1 to AC-4; see them fail on the unchanged code.
2. A `narrow_in_tx` next to `widen_in_tx` (same fence, same event, rows removed, the paths in the event); `workflow::review` takes the paths to keep and gives back the rest in the same transaction.
3. `card give-back` and `card line-lease` in `src/card/` (the line lease retries `widen_in_tx` every quarter second up to `--wait-secs`, default 60, at most 600, runs the command with its standard output sent to standard error so the answer stays one JSON envelope, then gives its added paths back whatever the status), and their clap entries and help.
4. Update the tests that pinned the old review answer, the card operations document and the help layout; run the card and ledger tests, the repository scans, `--lib`, `--bins`, clippy, then audit and write the Result.

## Result

- AC-1: met — `cargo test --locked --offline --test card_line_lease_cli review` passed (3 tests): `card review` without `--keep` gives back the whole scope, with `--keep` only the kept paths stay (the answer lists `given_back` and `kept`), another card takes a given-back path at once, and a `--keep` path the card does not hold is refused with the ledger unchanged.
- AC-2: met — `cargo test --locked --offline --test card_line_lease_cli give_back` passed: the named paths leave the reservation with one `card.scope` event that names them; an unknown path is refused with nothing written, a mixed list of a held and an unknown path (in both orders, and for `review --keep`) removes nothing, and a stale generation or another session leaves only the ledger's usual `fence_reject` event (D-4) and no change to a reservation, scope or card.
- AC-3: met — `cargo test --locked --offline --test card_line_lease_cli line_lease` passed: two cards each add a line to one file one after the other and while the first still runs and neither reservation lists the file afterwards; a path the card already holds stays; a refused wait names the other card and writes nothing; `--wait-secs` 601 is refused naming 600, 0 refuses at once, 600 is accepted, a wait of 2 lasts at least 1.9 s and under 5 s; a failing or unstartable command gives the path back and its status is named; the command's output reaches standard error and standard output stays one JSON envelope.
- AC-4: met — `cargo test --locked --offline --test card_line_lease_cli help` passed: the help of the three commands carries the rules (the test joins wrapped lines), the line-lease help says the paths stay in the reservation if the process is killed and that a reopened card widens its scope with `card scope`, and `docs/design-review/appendix/operations/card.md` names the three commands.
- AC-5: met — the 12 files of the acceptance list ran with `--no-fail-fast` and passed (202 tests at the audited head); `card_reclaim_cli` needed one change, its review helper passes `--keep` because `review` no longer leaves the reservation whole.
- Audit repair: the independent audit of head c657a46d77 found two blocking test gaps (a mixed list; the `--wait-secs` edges and the wait bound) and listed notes; this revision adds those tests and the ones for a live lease (review and line-lease need one, give-back does not) and the output stream, rewords the refusal of a failing command from "exited" to "ended" so the error classifier no longer calls it retryable I/O, adds to the error of a failed second ledger open that the command already ran and that the lease may remain, and states the `fence_reject` event and the kill limit in the help and the operations document.
- Not done here, and needed in the same release: the shipped MODE-AGENT text of the hand-off step still says `card review <card> --generation <g>` without `--keep`, and after a failed audit says `card reopen` and repair; since a review now gives every path back, a reopened card has an empty reservation and `card check` refuses its repair writes until `card scope` widens it again (the card layer is right, the text is behind; that file is reserved by another card and the text change must land with this one); a process killed by a signal while the command runs leaves the leased paths in the reservation until `card give-back` frees them (stated in the help); the older race and property tests still call `workflow::review`, which the command no longer uses, and there is no race test of `scope_return::review_keeping` itself; a stale generation after a hand-off (as opposed to a generation ahead) is not tested separately; the leased command has no time limit and its output is shown when it ends; `lease_take` matches paths without regard to ASCII case while `give_back` and `--keep` match exactly (a refusal, so the safe side).
- Mutants of the repair (seven, the ones the audit found alive, each restored): give-back removes the known paths before refusing a mixed list (killed by the mixed-list test, at the CLI and at the ledger), line-lease accepts an expired lease, give-back needs a live lease and review accepts an expired lease (all three killed by the live-lease test), the command output goes to standard output (killed by the stream test), the wait limit of 600 is not enforced (killed by the wait-edge test) and the wait lasts three times its limit (killed by the wait-bound test).
- Second audit round (head cfb8f24bed): B-1 and B-2 closed by the tests (seven mutants killed); one blocking selector finding, the mixed-list tests were not picked by the AC filters, closed by renaming them `give_back_with_a_mixed_list_removes_nothing` and by writing the twelve files of the AC-5 list into its verify line; the refusal of a failing command is now asserted not `io` and not retryable (it is `invalid_input` through the word "ended" of the classifier, which is imprecise and pinned by that assertion); the error of a failed second ledger open still has kind `io` and `retryable: true` next to the sentence that the command must not be run again (not changed here: the kind comes from the error text classifier).
- closure_status: the whole-area run on the final head, the python area modules and the re-audit are still to come.
