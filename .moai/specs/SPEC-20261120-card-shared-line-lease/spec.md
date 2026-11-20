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
- AC-5 (REQ-1, REQ-3, REQ-5): the existing card and ledger tests that read a reservation still pass · verify: `cargo test --locked --offline --test card_work_cli --test card_scope_widen_cli --test card_dispatch_cli --test ledger_reservation --test ledger_workflow --test prop_reservation`

## Scope

- write: `.moai/specs/SPEC-20261120-card-shared-line-lease/spec.md`
- write: `tests/card_line_lease_cli.rs`

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

Not yet written.
