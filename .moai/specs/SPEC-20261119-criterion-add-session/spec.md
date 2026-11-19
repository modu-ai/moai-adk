---
id: SPEC-20261119-criterion-add-session
title: "card criterion-add: record the session in the event and report the real criteria count"
status: draft
revision: 1
created: 2026-11-19
---

# card criterion-add: record the session in the event and report the real criteria count

## Intent

Card C-125, the follow-up of the C-108 tier 2 audit (card note on C-108, head 472b6bf2da, in PR 3230; findings N1 and N3). `card criterion-add` is the action C-108 adds (`SPEC-20261119-card-criterion-add`); its code is not on the tip before PR 3230 merges, so this SPEC is read against that head and the implementation starts on the tip after the merge.

N1. `workflow::add_criteria` (`src/ledger/workflow.rs`) writes the `card.criterion_add` event with the actor `--by` only, and `--by` is free text. `card.note` also records the session that wrote it (`session_holder("note", env).ok()` in `src/ops/card.rs`, stored in the note event as `session`). A criterion added by another session can keep `card done` refused (an open criterion blocks it, `card_done_guard`); the holder can withdraw it, but the ledger cannot say which session added it.

N3. `add_criteria` answers `total` as `highest + added.len()`, and `card criterion-add` prints it as `criteria`. The code comment says "how many criteria the card has now (withdrawn ones included)", which is the row count. The two numbers are equal while the numbers have no gap and differ when they do: the table accepts an insert with any `seq` (`card add` and the Issue import write numbers 1 to n, but nothing in the schema forces it), so a card with rows 1 and 5 answers 7 after two adds instead of 4.

This slice changes two things in `card criterion-add`: the event and the answer carry the caller's session, and `criteria` is the number of rows. It does not change how a new criterion is numbered (after the highest number, withdrawn included).

## Non-goals

- Authenticating the session: the id is read from the environment as `card note` reads it, so it says which session wrote the row, not who typed `--by`.
- Writing the session into rows that exist already: older `card.criterion_add` events keep the payload they have (no `session` key); a reader treats a missing key as unknown.
- Showing the event in `card show`: the event table is where an auditor looks; a display is a later card.
- Changing the numbering, the text rules or the refusals of `card criterion-add` (SPEC-20261119-card-criterion-add keeps them).
- `docs/design-review/appendix/operations/card.md` and `src/ops/card.rs` are reserved for card C-116 (leader inbox) while this SPEC is written; the implementation takes them after that release or merge (Decisions D-4).

## Requirements

- REQ-1: The `card.criterion_add` event payload shall record the id of the session that added the criteria as `session`, next to `first` and `count`, as the `card.note` event records its session.
- REQ-2 (REQ-1): IF the session has no id, THEN the event payload shall hold `session` as null while the criteria are added as usual.
- REQ-3 (REQ-1): The recorded `session` shall be the session's own id, whatever `--by` says.
- REQ-4: The answer of `card criterion-add` shall carry the same `session` as the event payload.
- REQ-5: The `criteria` of the answer and the `total` of `workflow::add_criteria` shall be the number of criterion rows the card has after the add, withdrawn ones included.
- REQ-6: The new criteria shall be numbered after the card's highest criterion number, as before this change.
- REQ-7: The help of `card criterion-add` shall state that the event records the adding session and that `criteria` counts the card's rows.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4): with `CLAUDE_CODE_SESSION_ID` set to a session id, `card criterion-add C-1 --ac "a" --by leader` stores one `card.criterion_add` event whose actor is `leader` and whose payload has `session` equal to that id, `first` and `count`; the answer has the same `session`; a different `--by` gives the same `session`.
  verify: `cargo test --locked --offline --test card_criterion_add_cli the_event_and_the_answer_record_the_adding_session`
- AC-2 (REQ-2): with no session variable, the criteria are added, the payload holds `session` as null and the answer holds `session` as null; a malformed session variable reads as no id.
  verify: `cargo test --locked --offline --test card_criterion_add_cli without_a_session_id_the_session_is_null_and_the_criteria_are_added`
- AC-3 (REQ-5, REQ-6): on a card whose criterion rows are 1 and 5, two adds answer `added [6, 7]` and `criteria 4`; the table holds four rows; `workflow::add_criteria` returns the same total.
  verify: `cargo test --locked --offline --test card_criterion_add_cli the_criteria_count_is_the_row_count_when_numbers_have_a_gap`; `cargo test --locked --offline --test ledger_workflow criteria_total_is_the_row_count`
- AC-4 (REQ-7): the help of `card criterion-add` names the recorded session and the row count.
  verify: `cargo test --locked --offline --test card_criterion_add_cli the_help_names_the_session_and_the_row_count`
- AC-5 (REQ-1…REQ-7): neighbours; every earlier `card_criterion_add_cli` and `ledger_workflow` test still passes.
  verify: `cargo test --locked --offline --test card_criterion_add_cli`; `cargo test --locked --offline --test ledger_workflow`; `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --lib`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ledger/workflow.rs` (the writer takes the session and counts rows), `src/ops/card.rs` (reads the session as `card note` does, the answer, the help; reserved for C-116 until it releases), `tests/card_criterion_add_cli.rs`, `tests/ledger_workflow.rs`, `.moai/specs/SPEC-20261119-criterion-add-session/spec.md`
- read-only: `src/card/dispatch.rs` (`session_holder`), `src/session_current.rs`, `src/ledger/schema.rs`, `tests/common/**`
- forbidden: the real HOME and ledger, the network, a change of table definitions or triggers, `docs/design-review/appendix/operations/card.md` until C-116 releases it (Decisions D-4)

## Plan

1. After PR 3230 merges: take the tip, add `src/ops/card.rs` and the tests to the reservation once C-116 releases them.
2. Tests first (RED): the four named tests, which fail before the code (no `session` key, `criteria` counted from the highest number).
3. `workflow::add_criteria` takes `session: Option<&str>` (one line, the author limit, as a note's session is checked), writes it into the payload and counts rows in the same transaction.
4. `card criterion-add` reads the session with `session_holder("criterion-add", env).ok()`, passes it and prints it.
5. Mutants, neighbours, SPEC Result with one line per AC.

## Risks

- The id is read from the environment, so a caller who sets the variable by hand can claim another session; the field says which session wrote the row, not who the person was. It is the same trust as a note's session.
- `add_criteria` gets one more parameter: every caller changes (`src/ops/card.rs`, `tests/ledger_workflow.rs`); a missed caller fails to compile, which is the check.
- The row count needs one more read in the write transaction; it runs in the same transaction as the inserts, so two adds at once cannot report a stale count.
- Before this change the two numbers are equal on every card made through `card add` and the import, so the answer changes only for a card with a gap in its numbers.

## Decisions

- D-1: the session goes in the event payload under the key `session`, the name the note event uses, so one query reads both kinds. Option not taken: a second actor column (a table change, forbidden).
- D-2: a missing session is `null`, not an absent key, so a reader tells "this session had no id" from "the row is older than this change" by `session` being present.
- D-3: the count is `SELECT COUNT(*)` after the inserts, in the same transaction. Option not taken: counting in memory from the highest number, which is the defect.
- D-4: this slice waits for card C-116 to release `src/ops/card.rs` and the operations document; the document row for `card criterion-add` is updated in the implementation if C-116 has released it by then, otherwise it stays a note on the card and the help is the contract.

## Result

Plan only. The implementation starts after PR 3230 merges and C-116 releases `src/ops/card.rs`; this section then gets one `- AC-n: met — evidence` line per criterion.
