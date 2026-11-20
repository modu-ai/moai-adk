---
id: SPEC-20261120-card-show-notes-once
title: "card show reads each card's note sessions once, not once per note"
status: draft
revision: 1
created: 2026-11-20
---

# card show reads each card's note sessions once, not once per note

## Intent

`moai card show` gives every note the session that wrote it, taken from the note's `card.note` event. It does that with a subquery that runs once per note and scans the whole `event` table each time, because `event` has no index on `card_id`. On a card with 900 notes and 903 events, measured in the collaboration-board design report (sections 3.5 and 6.2, evidence folder `collab-board-evidence-2026-11-20`), the note query takes 177.6 ms and 0.8 ms without the subquery; while two `card show` loops read such a card, the p95 of another lane's `card note` write goes from 46.4 ms to 152.7 ms, and with two writers from 90.9 ms to 1,788.7 ms (control: the same loops on another ledger). Every lane writes to this ledger, so a slow read of a big card slows all of them. The change reads the card's `card.note` events once and matches them to the notes in memory. The ledger schema, the printed answer and the rules for a read stay as they are.

## Non-goals

- A new index on `event`, a new table, a change of `journal_mode` or any other schema change: the ledger schema stays v1.
- A change of the output of `card show`, of any other command, or of the order of the notes.
- Making `card show` faster in any other part (the criteria, the links, the reservation query), or a read cache.
- The wait of a write behind a long read transaction (design report 3.5, finding 1): reads stay as short as they are.

## Requirements

- REQ-1: When `card show` reads the notes of a card, `card show` shall read the `card.note` events of that card once for the whole card and give each note the session of the first event (lowest event id) of that card that names the note, and `null` when no event names it or that event has no session. · failure: none — a read that fails is reported as the other queries of `show` report theirs (`LedgerError` text, exit 1) and prints no answer.
- REQ-2: When `card show` reads a card, `card show` shall print the same bytes as it printed before this change for the same ledger. · failure: none
- REQ-3: When `card show` reads the notes of a card, the query that reads the `card.note` events shall read the `event` table once and with no subquery, whatever the number of notes. · failure: none
- REQ-4: When `card show` reads a card, `card show` shall write no row. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): notes written by two sessions and by none, a note without any event (imported), a second event that names the same note and an event of another card that names the same note id all show the same `session` they showed before, in note order, and a card with 300 notes is shown whole with each note's own session and body.
  verify: `cargo test --locked --offline --test card_show_notes_cli`
- AC-2 (REQ-2, REQ-4): on a ledger with a 900-note card, the answer of the binary built before this change and of the binary built after it are the same bytes (`cmp`), and no row of any table differs after the read.
  verify: `manual: the lane runs a script with both binaries and records the cmp result and the table dump comparison in Result`
- AC-3 (REQ-3): the plan of the query that reads the events has no correlated subquery and names `event` once.
  verify: `cargo test --locked --offline --lib card::command::tests::the_note_events_are_read_once`
- AC-4 (REQ-3): with two `card show` loops on a 900-note card, the p95 of a `card note` write is within twice the p95 of the same measurement on another ledger, measured with `measure_ledger2.py` of the evidence folder against a binary built from this change.
  verify: `manual: the lane runs the script and records the table in Result`
- AC-5 (REQ-1…REQ-4): neighbours.
  verify: `cargo test --locked --offline --test card_cli`; `cargo test --locked --offline --test card_inbox_cli`; `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test card_export_cli`; `cargo test --locked --offline --lib card::`

## Scope

- write: `src/card/command.rs` (the note query of `show` and its unit test only), `tests/card_show_notes_cli.rs`, `.moai/specs/SPEC-20261120-card-show-notes-once/**`
- read-only: `src/ledger/schema.rs`, `src/ledger/workflow.rs`, `src/ledger/open.rs`, `tests/common/**`
- forbidden: the real HOME and ledger, the network, `src/ledger/schema.rs`, a new index or table, `journal_mode`, `skills/moai/**`, any change of what `card show` prints

## Plan

1. Tests first: `tests/card_show_notes_cli.rs` pins what `session` is today (it passes before the change, as a characterisation); the unit test of the query plan is added with the code and fails to compile without it.
2. In `show`, replace the per-note subquery with one query, `SELECT json_extract(payload, '$.note_id'), json_extract(payload, '$.session') FROM event WHERE card_id = ?1 AND kind = 'card.note' ORDER BY event_id`, keep the first session per note id in a map and read each note's session from it.
3. Compare the bytes of the answer before and after on a 900-note card, run the neighbouring tests, mutate the new code (first wins, card filter, `null` when absent), and measure the writer p95.

## Risks

- A second event that names one note, or an event of another card with the same note id, would change a session if the map kept the last event or ignored the card: AC-1 pins both.
- A session that is not text or `null` (`json_extract` returns a number or an object) made the old query fail the whole read at that note; `add_note` only writes text or `null`, so no such event exists in a ledger made by the tool. The new code skips such an event, so the note shows `null` instead of the read failing; a note id that is not an integer is skipped the same way. Both are pinned in the unit test as skipped; a hand-edited ledger is not a supported input.
- The new query still scans `event` once: with no index on `card_id` that cost grows with the event table, not with the notes of the card. That is the intended cost (one scan instead of one per note) and is measured in AC-5.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | One query over the card's `card.note` events and a map in memory, not an index | the user's instruction relayed by the leader: schema v1 unchanged | A later schema version may add an index on `event(card_id, kind)`; then the query stays and gets cheaper |
| D-2 | The first event (lowest event id) of the card wins, as the old `ORDER BY e.event_id LIMIT 1` did | keeps the answer byte for byte | none needed |
| D-3 | AC-2 and AC-4 are `manual:` lines with the numbers recorded in Result | a byte comparison between two builds and a p95 over a script are not tests of this repository | A repository test needs two builds; none is planned |

## Result

Built from the local integration tip f9523ae236 (merged into the branch, no rebase); the tip is now 70707144ba and `git merge-tree --write-tree` of the head with it exits 0. Characterisation first: commits f0bd29ded3 and 0eba585cdb added `tests/card_show_notes_cli.rs` and it passed on the unchanged code (4 passed); then 7c7d2e1f10 replaced the per-note subquery.

- AC-1: met — `tests/card_show_notes_cli.rs` 4 passed on the old code and on the new (two sessions and none, an imported note without an event, a second event for one note and another card's event with the same note id, 300 notes with every note's own session and body, no row written; two reads give the same bytes).
- AC-2: met — a ledger made with the installed 4.2647.9 CLI: one card with 900 notes written by four sessions and none, plus 5 imported notes, 2 second events for the first note and 1 event of another card. `card show` of the build before this change and of the build after it give 209,335 bytes each, the same sha256 (07341b94…9741), `cmp` reports identical, and the ledger file hash is the same before and after each read.
- AC-3: met — `card::command::tests::the_note_events_are_read_once`: `EXPLAIN QUERY PLAN` of the events query has no `SUBQUERY` and names `event` once; the notes query names neither `event` nor a subquery; the first event decides (also with a `null` session), another card's event and an event of another kind add nothing, a text note id and a numeric session are skipped.
- AC-4: met (one run, load average about 40, so read the ratios and not the milliseconds) — measure_ledger2.py of the design report, build before and after, p95 of a write in ms: two `card show` loops on the same ledger 221.8 → 83.5 (control on another ledger 20.0 → 100.3 in the second run, so the ratio to the control went 11.1 → 0.8); two writers and two `card show` loops 782.7 → 154.6 (control 20.3 → 339.2, ratio 38.6 → 0.5). Both ratios are under 2; against the no-reader case of the same run the first one is 83.5 / 40.6 = 2.1 (that case was slower than in the first run too, 20.8 → 40.6). Every write succeeded (0 failed of 120 and 60). Neither run is a release build. The machine was much busier during the second run (the control itself got slower), so the absolute numbers do not compare; the ratio to a control measured in the same run is the evidence.
- AC-5: met — `cargo test --lib card::` 96 passed, `--lib ledger::` 25, `--bins` 28, all 31 targets of `tests/card*.rs`, `tests/ledger*.rs` and `help_layout` exit 0 (including card_cli, card_inbox_cli, card_work_cli, card_export_cli), fmt, ops rustfmt, clippy `--all-targets` and `git diff --check` exit 0, 39 of 39 registered `repo_scans` checks exit 0, head-built trailer-check exit 0.
- Mutants, 5 of 5 killed by the unit test: last event wins (`insert` for `or_insert`), every card's events (`?1 = ?1`), any event kind (the kind filter dropped), a `null` session skipped, a non-text session turned into `null`.
- Not changed: the schema (no index, no table, `journal_mode`), the answer of any command, `docs/design-review/appendix/operations/card.md` (the `session` field is described there as before and the doc needs no change; the file was held by another card).
