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
- A session that is not a string (`json_extract` returns a number or an object) made the old query fail the whole read at that note; `add_note` only writes a string or `null`, so no such event exists in a ledger made by the tool, and the new code reads the value the same way. The edge is not pinned by a test: a hand-edited ledger is not a supported input.
- The new query still scans `event` once: with no index on `card_id` that cost grows with the event table, not with the notes of the card. That is the intended cost (one scan instead of one per note) and is measured in AC-5.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | One query over the card's `card.note` events and a map in memory, not an index | the user's instruction relayed by the leader: schema v1 unchanged | A later schema version may add an index on `event(card_id, kind)`; then the query stays and gets cheaper |
| D-2 | The first event (lowest event id) of the card wins, as the old `ORDER BY e.event_id LIMIT 1` did | keeps the answer byte for byte | none needed |
| D-3 | AC-2 and AC-4 are `manual:` lines with the numbers recorded in Result | a byte comparison between two builds and a p95 over a script are not tests of this repository | A repository test needs two builds; none is planned |
