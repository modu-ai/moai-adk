---
id: SPEC-20261119-card-criterion-add
title: "card criterion-add: add an acceptance criterion to a card that exists"
status: draft
revision: 1
created: 2026-11-19
---

# card criterion-add: add an acceptance criterion to a card that exists

## Intent

Card C-108. `card done` closes a card only with one criterion met, none open and a merged pull request (`src/ledger/workflow.rs`, `done`; trigger `card_done_guard`). The only way to give a card a criterion is `card add --ac` at creation (and the Issue import). A card created without one can never be closed: `card done` answers "the card has no acceptance criterion that is met" and no command adds a criterion (C-40, a migrated Issue card verified fixed at tip by PRs #3091 and #3073, card note 2576; audit of #3225, finding N2: the flaky rule in `skills/moai/workflows/mode-leader.md` tells the leader to run `card add --kind task --title "flaky: <test>"`, which has no `--ac`).

Measured at tip 7d744a28f3 plus the C-94 head 15b333fe73: table `card_ac` accepts inserts (`card add` and the import do it); triggers forbid delete and a change of `card_id` or `seq`; `card_touch_ac_insert` moves the card's `rev` and `updated_at`; table `event` takes any `kind` text. `card criterion C-n N --met|--withdraw` is a fenced writer (needs the holder and the generation) and only changes a criterion that exists.

This slice adds one action, `card criterion-add C-n --ac TEXT [--ac TEXT …] [--by W]`, that appends criteria to an existing card in one transaction with one event. It is a separate action, not `card criterion add`: `card criterion` already takes the card and the criterion number as positionals, and a subcommand word in the same place would make `add` a possible card reference.

## Non-goals

- Removing, editing or renumbering a criterion: the table forbids delete and a change of `seq`; a criterion that no longer applies is withdrawn (`card criterion --withdraw`).
- Marking the new criterion met: `card criterion` does that, as the holder, at the generation.
- A fence: adding a criterion changes nothing a lane may write, so it needs no generation and no holder (the same as `card note` and `card scope`). A card nobody holds, such as a migrated Issue card, can then be given a criterion by whoever finds it.
- A count or length limit beyond what `card add --ac` has (one line, not empty, no control character): both commands get the same text rules; limits for both are a later card (Risks).
- The flaky-rule text in `skills/moai/workflows/mode-leader.md` (card C-108 criterion 3): it needs `skills/moai/references/source-manifest.json`, which card C-59 holds; it is slice 2 after that release.

## Requirements

- REQ-1: When `card criterion-add` is given a card and one or more `--ac` texts, `card criterion-add` shall append each text as an `open` criterion numbered after the card's highest criterion number, in the order given, in one transaction with one `card.criterion_add` event.
- REQ-2 (REQ-1): IF the card is `done`, THEN `card criterion-add` shall refuse the request as `conflict`, name the state, and write nothing.
- REQ-3 (REQ-1): IF a text is empty or has a control character, or `--by` is not one line of 1 to 80 characters, THEN `card criterion-add` shall refuse the request as `invalid_input`, name the text by its place in the list, and write nothing.
- REQ-4 (REQ-1): IF the request has no `--ac`, THEN `card criterion-add` shall refuse it as `usage` (exit 2) before it reads the ledger.
- REQ-5 (REQ-1): IF the card is not named by a C-number, or does not exist, THEN `card criterion-add` shall refuse the request as `invalid_input` or `not_found` and write nothing.
- REQ-6: The `card.criterion_add` event shall record who added the criteria (`--by`), the number of the first one and how many were added.
- REQ-7: While a card is held, `card criterion-add` shall leave its reservation, generation and lease as they were.
- REQ-8: While a card has an open criterion, `card done` shall refuse it, as it did before this change.
- REQ-9: The help of `card criterion-add` and `docs/design-review/appendix/operations/card.md` shall state that the action writes the ledger, that criteria are only added, that no generation is needed, and that a `done` card is refused.

## Acceptance

- AC-1 (REQ-1, REQ-6): on a card with two criteria, one of them withdrawn, `card criterion-add C-1 --ac "a" --ac "b" --by leader` answers criteria 3 and 4 (numbered after the highest, withdrawn included), status `open`, in the given order; `card show` lists all four; one `card.criterion_add` event has actor `leader`, `first` 3 and `count` 2; the card's `rev` moved; a card with no criterion gets number 1.
  verify: `cargo test --locked --offline --test card_criterion_add_cli criteria_are_appended_after_the_highest_number_with_one_event`
- AC-2 (REQ-1, REQ-8): a card made without a criterion is closed after `card criterion-add`, `card criterion --met`, a merged pull request and `card review`; `card done` still refuses a card with zero met criteria and a card with an open one.
  verify: `cargo test --locked --offline --test card_criterion_add_cli a_card_made_without_a_criterion_can_be_closed_once_one_is_added_and_met`
- AC-3 (REQ-2, REQ-5): a `done` card is refused as `conflict` naming `done`; a card that does not exist is `not_found`, a name that is not a C-number `invalid_input`; every table is as before.
  verify: `cargo test --locked --offline --test card_criterion_add_cli a_done_card_and_a_missing_card_are_refused_and_nothing_is_written`
- AC-4 (REQ-3, REQ-4): no `--ac` is `usage` (exit 2); an empty text, a text with a line break, a blank `--by`, an 81-character `--by`, and the second of three texts being the bad one: `invalid_input` naming the place, nothing written, the first good text not stored.
  verify: `cargo test --locked --offline --test card_criterion_add_cli a_text_that_cannot_be_stored_refuses_the_whole_request`
- AC-5 (REQ-7): on a card held by a session in `active` and one in `review`, the criterion is added; the reservation row (generation, holder, lease, status) and its scope rows are as before; `card check` for the holder still answers as before.
  verify: `cargo test --locked --offline --test card_criterion_add_cli adding_to_a_held_card_leaves_its_reservation_alone`
- AC-6 (REQ-1): two `card criterion-add` processes at once give every criterion its own number, none lost, none twice.
  verify: `cargo test --locked --offline --test card_criterion_add_cli two_adds_at_once_number_every_criterion_once`
- AC-7 (REQ-9): the help names the four statements of REQ-9, the operations document has a `card criterion-add` row naming them, and the first line of the help says it writes the ledger.
  verify: `cargo test --locked --offline --test card_criterion_add_cli the_help_and_the_ops_doc_name_the_rule`
- AC-8 (REQ-1…REQ-9): neighbours.
  verify: `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test ledger_workflow`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --lib`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ledger/workflow.rs` (the writer), `src/card/input.rs` (the text rules, shared with `card add`), `src/ops/card.rs` (the action, its help and its answer), `tests/card_criterion_add_cli.rs`, `tests/ledger_workflow.rs`, `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261119-card-criterion-add/spec.md`
- read-only: `src/ledger/schema.rs`, `src/card/work.rs`, `src/card/command.rs`, `src/card/dispatch.rs` (held by card C-43), `tests/card_reclaim_cli.rs` (as a pattern), `tests/common/**`
- forbidden: the real HOME and ledger, the network, a change of table definitions or triggers, `skills/moai/**` (slice 2)

## Plan

1. Tests first: `tests/card_criterion_add_cli.rs` for AC-1 to AC-7 (all fail before the code: there is no such action).
2. `input::criteria_from(start, entries)`: the rule of `card add --ac`, numbered from `start`; `card add` keeps calling it with 1.
3. `workflow::add_criteria`: one write transaction (state check, highest number, inserts, event).
4. The action in `src/ops/card.rs`, its help, the operations document row.
5. Mutants, neighbours, SPEC Result.

## Risks

- The text rules are the ones `card add --ac` has: no length limit and no secret screening (a note has both). A criterion is quoted in cards, exports and audits; limits for both commands are one later card.
- Any caller can add a criterion to any card, held or not, and a card in `review` with every criterion met then needs the new one met before `card done`. That is the point (the leader finds the missing criterion), and the event records who did it.
- `done` cards are refused; a closed card whose criteria were wrong needs a new card.

## Result

RED first: commit a3802ca715 (seven tests, all failing before the code with "unrecognized subcommand 'criterion-add'"). Built on the C-94 head 15b333fe73 (tip 7d744a28f3 at the time); the change shares no file with C-94 except `src/ops/card.rs` and `docs/design-review/appendix/operations/card.md`, which C-94 released first.

- AC-1: met — `criteria_are_appended_after_the_highest_number_with_one_event`: on a card with criteria 1 (open) and 2 (withdrawn), `criterion-add --ac "  a  " --ac b --by leader` answers `added [3, 4]`, `criteria 4`, `by leader`, `events 1`; `card show` lists the four with the text trimmed (`a`), status `open`; `rev` grew; one `card.criterion_add` event has actor `leader` and payload `first 3`, `count 2`; a card with no criterion gets number 1 and the actor defaults to `user`. Mutants A1 (number collides), A4, A5 (payload), A6 (actor), A9 (total), I1 (no trim), O3 (answer `by`) killed. The test of the writer itself is `ledger_workflow::criteria_are_added_to_any_card_but_a_closed_one_and_each_gets_its_own_number` (rows, event JSON, no write on a refusal).
- AC-2: met — `a_card_made_without_a_criterion_can_be_closed_once_one_is_added_and_met`: with zero criteria `card done` is refused ("the card has no acceptance criterion that is met"); with one open criterion added it is refused with the same sentence (nothing is met yet); with criterion 1 met and criterion 2 added (open) it is refused ("an acceptance criterion is still open"); with both met it closes (`state done`). `card done` itself is unchanged.
- AC-3: met — `a_done_card_and_a_missing_card_are_refused_and_nothing_is_written`: a `done` card is `conflict` naming "C-2 is `done`", C-99 is `not_found`, `card-1` and `#12` are `invalid_input` (`#12`: "name the card by its C-number"); the dump of seven tables is byte-equal before and after. Mutants A2 (done accepted), A3 (missing card), O2 (host number accepted) killed.
- AC-4: met — `a_text_that_cannot_be_stored_refuses_the_whole_request`: empty, blank, a line break, the second of three texts bad ("criterion 2 must be one line"), blank `--by` and an 81-character `--by` are `invalid_input`; the dump is byte-equal (the good texts are not stored); an 80-character `--by` is accepted; no `--ac` is `usage`, exit 2. The limit is 80 characters (`ACTOR_MAX_CHARS`), not 128 as the first draft said. Mutants I2 (line break), O1 (`--by` unchecked), A8 (author unchecked in the ledger) killed; A7 (empty list) is killed by the ledger-level test only, because the parser requires one `--ac`.
- AC-5: met — `adding_to_a_held_card_leaves_its_reservation_alone`: on an `active` and a `review` card held by a session, the reservation rows (generation, holder, status, acquired_at, lease_until, worktree, branch), the scope rows and `card check` for the holder are equal before and after; the states are unchanged; the review card then refuses `card done` for the open criterion.
- AC-6: met — `two_adds_at_once_number_every_criterion_once`: four processes at once, numbers 1 to 5 with no gap and no repeat, all four texts present. (The numbering runs in one write transaction; mutant A1 shows the numbering is what the test pins.)
- AC-7: met — `the_help_and_the_ops_doc_name_the_rule`: the help starts "Writes the ledger." and contains "criteria are only added", "needs no generation", "a `done` card is refused"; the operations document has a `card criterion-add` row with `workflow::add_criteria`, 더하기만, 세대가 필요 없, `done` 카드는 거절.
- AC-8: met — `card_criterion_add_cli` 7, `card_work_cli` 24, `card_dispatch_cli` 25, `card_reclaim_cli` 9, `card_scope_widen_cli` 9, `ledger_workflow` 24, `ledger_reservation` 27, `ledger_recover` 49, `help_layout` 10, `product_cli` 79, `home_isolation_scan` 18, `git_env_scrub` 3, `ac_command_lint` 5, `--lib` 1451, `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0 (the `ledger_workflow` count and the clippy run of the final commit are in the report below).

Mutants: 14 of 14 killed (A1 to A9 in `src/ledger/workflow.rs`, I1 and I2 in `src/card/input.rs`, O1 to O3 in `src/ops/card.rs`).

Not done here: card C-108 criterion 3 (the flaky rule's `card add` example in `skills/moai/workflows/mode-leader.md` carries `--ac`): slice 2, after card C-59 releases `skills/moai/references/source-manifest.json`. The help and the operations document say "criteria are only added": a `card criterion-add` text has no length limit and no secret screening, as `card add --ac` has none (Risks).
