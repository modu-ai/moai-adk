---
id: SPEC-20261119-card-scope-widen
title: "card scope widens a held card's write scope under the same generation"
status: draft
revision: 1
created: 2026-11-19
---

# card scope widens a held card's write scope under the same generation

## Intent

Card C-94 (workflow W1). A lane that is told its write scope is too small today has to `card release` and `card take` again with a longer `--scope`. The leader's evidence (card note 2318): eight such release-and-retake asks (notes 2143, 2144, 2180, 2183, 2256, 2110, 2258, 2278), each approved by a note. A retake bumps the generation, so the old fence is gone, and it can lose a reservation to another lane in between.

Measured from the code at tip bdb4d3db8c: `card scope` (`src/card/dispatch.rs`, `scope`) replaces the scope of a `draft` or `ready` card nobody holds and refuses every other card (`conflict`, "release or hand it over first"). A held card's write scope is the rows of table `reservation_scope` of its active reservation (`src/ledger/reservation.rs`, inserted once in `reserve_in_tx`), and `card check` reads those rows, not `card.write_scope`. The tables forbid deleting a row (`reservation_no_delete`) and events are append-only, so widening is inserting rows and one event.

This slice lets `card scope` take a card that has an active reservation: the new list must keep every pattern the reservation has, the added ones are checked against every other card's active reservation in the same write transaction, and one `card.scope` event records who widened it and why.

## Non-goals

- Narrowing a held card, or removing a pattern: refused, and the rows are never deleted.
- Changing `card.write_scope` of a held card (the card's own scope, used by a later `card take` without `--scope`): only the reservation grows; the help says so.
- Enforcing who may widen (the leader, by policy): `--by` is free text recorded in the event, as for `card scope` today.
- Renewing or reclaiming a lease: widening leaves lease, worktree and branch as they were.
- A new event kind: the kind stays `card.scope`; `card export` and `ledger recover` copy events as written and no reader switches on the kind (checked: `git grep "card.scope"` finds only `dispatch.rs`).

## Requirements

- REQ-1: When `card scope` is given a card that has an active reservation, `card scope` shall insert the added patterns into that reservation's write scope under the same generation, with its holder, lease, worktree and branch as they were.
- REQ-2 (REQ-1): IF the new list leaves out a pattern the reservation already has, THEN `card scope` shall refuse it as `conflict`, name the patterns left out, and write nothing.
- REQ-3 (REQ-1): IF an added pattern could be written under a pattern of another card's active reservation, THEN `card scope` shall refuse it as `conflict`, name that card, and write nothing.
- REQ-4 (REQ-1): IF `--reason` is missing, blank, over one line or over 500 characters, or has a word shaped like a secret, or no entry of the list is new, or the widened scope would have more than 64 patterns, or an entry fails the check `card add --scope` applies, THEN `card scope` shall refuse the request as `invalid_input` and write nothing.
- REQ-5: When `card scope` widens a held card, `card scope` shall write one `card.scope` event in the same transaction as the new rows, with who widened it (`--by`), why (`--reason`), the generation, the holder, the old and the new scope. · failure: none — the rows and the event are one transaction, so a refused widening writes neither.
- REQ-6: The widening shall leave `card.write_scope` of the card as it was.
- REQ-7: While the reservation holds an added pattern, `card check` for the holder at the same generation shall answer allowed for a path under it.
- REQ-8: While a card is `draft` or `ready` and nobody holds it, `card scope` shall replace its scope as it did before this change.
- REQ-9: The help of `card scope` and `docs/design-review/appendix/operations/card.md` shall state that a held card's scope only grows, that `--reason` is required for it, that the card's own scope is not changed, and that the generation stays.

## Acceptance

- AC-1 (REQ-1, REQ-5, REQ-6, REQ-7): on a card taken for a holder, `card scope --scope <old…> --scope <new> --by leader --reason …` answers the same generation, lease, worktree and branch, the reservation lists the new pattern, `card.write_scope` is as before, one `card.scope` event carries by, reason, generation, holder, from and to, and `card check` for the new path answers allowed (and refused before).
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_held_card_gets_wider_under_the_same_generation`
- AC-2 (REQ-3): a pattern another live reservation holds is refused as `conflict` with that card's name and nothing is written; it is allowed after that card is released.
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_pattern_another_card_holds_is_refused_by_name`
- AC-3 (REQ-2): a list that leaves out a current pattern is refused as `conflict` naming it; the reservation rows and events are unchanged.
  verify: `cargo test --locked --offline --test card_scope_widen_cli narrowing_a_held_card_is_refused`
- AC-4 (REQ-4): missing, blank, two-line, 501-character and secret-shaped `--reason`, a list with nothing new, 64 patterns accepted and 65 refused, an absolute path: `invalid_input`, nothing written.
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_widening_that_cannot_be_recorded_is_invalid_input`
- AC-5 (REQ-8): a `ready` card nobody holds still gets its scope replaced, with the same answer and event as before (no `held` key; a `--reason` is recorded when given), and a card that was held and then given back is replaceable again; a card in any other state still gets the refusal it got before.
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_card_nobody_holds_keeps_the_old_scope_rules`
- AC-6 (REQ-1, REQ-3): two lanes widening toward the same path at once: exactly one wins.
  verify: `cargo test --locked --offline --lib ledger::reservation::tests::widen`
- AC-7 (REQ-9): the help of `card scope` and the operations document name the four statements of REQ-9.
  verify: `cargo test --locked --offline --test card_scope_widen_cli the_help_and_the_ops_doc_name_the_rule`
- AC-8 (REQ-1…REQ-9): neighbours.
  verify: `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_reclaim_cli`; `cargo test --locked --offline --test ledger_reservation`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs` (the new `--reason` option and the help only), `src/card/command.rs` (only if the option needs a parser change), `tests/card_scope_widen_cli.rs`, `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261119-card-scope-widen/spec.md`
- read-only: `src/ledger/schema.rs`, `src/ledger/recover.rs`, `src/card/export_db.rs`, `src/scope.rs`, `tests/common/**`, `tests/card_reclaim_cli.rs` (as a pattern)
- forbidden: the real HOME and ledger, the network, a change of table definitions or triggers, `skills/moai/**`

## Plan

1. Tests first: the CLI tests of AC-1 to AC-5 and AC-7 fail before the code (the help test, the held-card tests: today `card scope` refuses a held card).
2. `reservation::widen_in_tx` (reads the active reservation, checks the superset, `find_overlap` for the added patterns only, the 64-pattern total, inserts the rows) and the `card.scope` event with the extra payload fields; `dispatch::scope` calls it for a card with an active reservation, keeps its old path for `draft` and `ready`.
3. `--reason` on `card scope` (`src/ops/card.rs`), the help, the operations document.
4. A unit test of two interleaved widenings (`reservation.rs`), mutants, SPEC Result.

## Risks

- A widening is not a handover: the holder of a card whose lease has ended still cannot write (the fence answers `lease_expired`); the new rows stay for whoever reclaims the card. Widening an expired reservation is allowed on purpose, so the leader can prepare a card a lane left.
- The payload of a `card.scope` event of a held card has more keys than before (`held`, `generation`, `holder`, `reason`): a reader that insists on exactly `from` and `to` would break; none exists at tip.
- A pattern that overlaps one of the card's own patterns is added as written (the superset rule is textual): `src/**` next to `src/a.rs` is legal and redundant.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only the reservation grows; `card.write_scope` stays | research note 2632 point 1: a later `take` without `--scope` reads the card's own scope, and for C-59 that is empty | Also append to `card.write_scope` |
| D-2 | The actor is not checked; `--by` is recorded | `card scope` today takes free text; the leader rule is policy | Require a leader marker later |
| D-3 | The event kind stays `card.scope`, with extra payload keys on a held card | export and recover copy events verbatim; no reader switches on the kind | A new kind `card.widen` |
| D-4 | The new list must contain every current pattern (textual match); the added ones are the difference | the command already means "this is the scope"; a narrowing attempt then has a natural refusal, and rows are never deleted anyway | An `--add` option that only adds |
| D-5 | `--reason` is required for a held card and optional (recorded when given) for an unheld one | the card says the event names who and why; an unheld card keeps its old call unchanged | Always required |

## Result

(RED and the checks come first; one line per AC is written after them.)
