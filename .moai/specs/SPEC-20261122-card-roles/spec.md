---
id: SPEC-20261122-card-roles
title: "Card roles: each card names its authors, auditor and lander from ledger rows, and a verdict from an author is refused"
status: draft
revision: 2
created: 2026-11-22
---

# Card roles: each card names its authors, auditor and lander from ledger rows, and a verdict from an author is refused

## Intent

User decision of 2026-11-21 (stage 3 of the flow redesign, option (a)): there is no resident leader session; the author, the auditor and the lander of a card are recorded per card by the ledger. Today the ledger knows who held a card (the reservation rows) and who landed it (the `card.landed` row), but nothing says who audits, and nothing stops the session that implemented a card from writing the verdict on it. The rule "the agent that implemented a change never audits it" lives only in instructions.

This SPEC is slice 1 of the roles: the read side and the two refusals. A card shows its authors, its auditor and its lander, taken from ledger rows; an audit assignment never names an author; a typed verdict row from a session that held the card is refused with the reason. The typed submit and block rows, the landing gate that reads a verdict, and the inbox that lists unanswered rows are the next slices (the plan of the flow skeleton, steps 4a to 4f), and are not here.

Grounded in the code at the integration tip `fd56877af5` before this SPEC:

- Holders are recorded per generation in the `reservation` table (`holder`, `status`, never deleted: `src/ledger/schema.rs:112`), and the holder is the session that took the card (`src/card/dispatch.rs`, `session_holder`).
- Events are free-text kinds with a JSON payload and the writing session as `actor` (`src/ledger/schema.rs:133`), and `reservation::add_event` is the one writer; a typed signal fits an event row and needs no schema change (the plan of 2026-11-21, constraints).
- The `card.review` row records the session that sent a card to review (`src/ledger/workflow.rs`, `review`), and the `card.landed` row records the head, the new tip and the caller word.

## Non-goals

- A landing gate that reads a verdict, tier floors, and holders landing their own cards: the next slice (the plan's decision D3).
- The typed submit row, the typed block and answer rows, `card inbox` and `card wait` on typed rows.
- Choosing an auditor automatically, or launching an audit session: this SPEC records an assignment someone made.
- Changing the schema, the version gate or any existing event kind.
- Treating a lane name as an identity: until the lane name becomes the holder (a later slice), the identity compared is the holder string recorded in the reservation.

## Requirements

- REQ-1: When the card show command is asked for a card, the card show command shall list the authors, the auditor and the lander of that card from ledger rows · failure: none — a role with no row is listed as absent, never guessed.
- REQ-2: When a session writes a verdict for a card, the card verdict command shall refuse the verdict if that session holds or ever held a reservation of the card · failure: none — the answer names the card and the reason, the ledger is unchanged, and the exit status is non-zero.
- REQ-3: When the card verdict command accepts a verdict, the card verdict command shall write one `card.verdict` event that holds the head, the result and the evidence reference, with the writing session as its actor, in one transaction · failure: none — a card that is not in review, a head that is not a full commit id, or a result other than pass or fail refuses before any write.
- REQ-4: When an audit assignment names an identity that holds or ever held a reservation of the card, the card assign command shall refuse the assignment · failure: none — the answer names the identity and the card, and the ledger is unchanged.
- REQ-5: When the card assign command accepts an assignment, the card assign command shall write one `card.audit_assigned` event that holds the auditor and the assigning caller · failure: none — only the leader word assigns, and a card that is not in review refuses before any write.
- REQ-6: When an auditor is assigned to a card, the card verdict command shall refuse a verdict from any other session · failure: none — the answer names the assigned auditor, and a card with no assignment accepts a verdict from any session that never held it.
- REQ-7: When a session repeats a verdict with the same head and result, the card verdict command shall write no second row · failure: none — the answer is the first row with a note that nothing was written.

## Acceptance

- AC-1 (REQ-1): a card with two reservation generations held by two sessions, one `card.audit_assigned` row and one `card.landed` row shows both holders as authors, the assigned identity as auditor and the landing's caller and head as lander; a card with none of these rows shows each role as absent · verify: `cargo test --locked --offline --test card_roles_cli roles_`
- AC-2 (REQ-2): the session that took a card and sent it to review writes a verdict, and the command refuses with a message that names the card and says the writer held it; a session that held an earlier generation is refused the same way; the event count of the card is unchanged · verify: `cargo test --locked --offline --test card_roles_cli verdict_author_`
- AC-3 (REQ-3): a session that never held a card in review writes a verdict with a full head, a result and an evidence reference and the ledger holds exactly one `card.verdict` row with those fields and that session as actor; a short head, a result of maybe, and a card that is active each refuse before any write · verify: `cargo test --locked --offline --test card_roles_cli verdict_rows_`
- AC-4 (REQ-4): an assignment that names the holder of any generation refuses and writes nothing · verify: `cargo test --locked --offline --test card_roles_cli assign_author_`
- AC-5 (REQ-5): the leader's assignment of a session that never held the card writes one row; another caller word and a card that is not in review each refuse before any write · verify: `cargo test --locked --offline --test card_roles_cli assign_rows_`
- AC-6 (REQ-6): after an assignment, a verdict from another session that never held the card refuses naming the assigned auditor, and the assigned auditor's verdict is accepted · verify: `cargo test --locked --offline --test card_roles_cli verdict_assigned_`
- AC-7 (REQ-7): the same session repeating the same head and result writes no second row and says so; a different result for the same head writes a new row · verify: `cargo test --locked --offline --test card_roles_cli verdict_repeat_`

## Scope

- write: `.moai/specs/SPEC-20261122-card-roles/spec.md`
- write (slices 2 and 3, on this card's own widening): `src/card/roles.rs` (the role reader, the verdict and assign actions), `src/card/mod.rs` (one `pub mod` line), `src/ledger/workflow.rs` (the two writers `record_verdict` and `record_audit_assigned`, as their own functions), `tests/card_roles_cli.rs`, `docs/design-review/appendix/operations/card.md` (two rows and the show row)
- write by line-lease or the leader's word: `src/ops/card.rs` (the clap rows of `verdict` and `assign`, and the roles in the show answer), because another card reserves that file
- read-only: `src/ledger/schema.rs`, `src/card/dispatch.rs`, `src/card/command.rs`
- forbidden: the schema, the version gate, `src/main.rs`, `src/help.rs`, any existing event kind

## Plan

1. This SPEC, read by the leader and by the audit before any code.
2. Slice 2, tests first (RED): `tests/card_roles_cli.rs` with `roles_`, `verdict_` and `assign_` against a temporary home and a real ledger; the sessions are made with the host's session variable, one per role.
3. Slice 3: `src/card/roles.rs` and the two ledger writers; make the tests pass; the show answer gains a `roles` object.
4. Slice 4: the clap rows through the line-lease or the leader's word, the operations table, the SPEC Result.
5. Per slice: light checks, at most the mutant cap, one affected run, then the READY note with head and base only.

## Risks

- A session id is not a person: two sessions of one person count as two identities, so the rule stops a session from judging its own card and does not stop a person who opens a second session. That is the rule the instructions already state, now written in the ledger.
- The holder of a card can be a lane name once a later slice makes the lane the holder; the comparison reads the recorded holder string, so old and new reservations both count. A card whose holder changed name between generations lists both.
- A verdict row that nothing reads yet is inert: the landing gate comes in the next slice. Until then the leader still reads the verdict note by hand, and the typed row is a second record, so a disagreement between the two is possible and the note wins.
- `src/ops/card.rs` is reserved by another card; the clap rows are small additions, and the earlier slices of this lane were done through the leader's word for the same reason.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | An author is any session that holds or ever held a reservation of the card, every generation, released or not | The rule says the agent that implemented a change never audits it; every holder may have written part of it | Narrow the query to the generation that sent the card to review |
| D-2 | The roles are read from existing rows, and only the assignment and the verdict are new event kinds, with no schema change | The plan of 2026-11-21 puts typed signals in event rows, and the version gate refuses a schema change | Stop writing the two kinds; old readers ignore unknown kinds |
| D-3 | Only the leader word assigns an auditor | Until a typed verdict gates landing, the leader is the one who dispatches audits (plan decision D3) | Allow the holder-independent caller list to grow |
| D-4 | A verdict is accepted only for a card in review | A card in review is the one that was submitted; a verdict on an active card judges unfinished work | Allow `done` cards for late findings |
| D-5 | The head of a verdict must be a full 40-character commit id | A short id can name two commits and a verdict must name one | Accept a unique prefix after resolving it |

## Result

Slices 2 and 3, revision 2. Code head 60895b3320 over base b68fce6b5c (the SPEC commit, landed). Numbers come from runs made for this result.

- AC-1 met: `roles_are_read_from_reservations_the_assignment_and_the_landing` passed; `card show` gains a `roles` object (`authors`, `auditor`, `lander`) read from the reservation, `card.audit_assigned` and `card.landed` rows.
- AC-2 met: `verdict_author_sessions_are_refused_with_the_reason` passed (the current holder and an earlier holder are refused, the card is named, nothing is written).
- AC-3 met: `verdict_rows_hold_the_head_the_result_the_evidence_and_the_writer` passed (a short head, a result of maybe and an active card refuse before any write).
- AC-4 met: `assign_author_identities_are_refused` passed.
- AC-5 met: `assign_rows_hold_the_auditor_and_the_caller_and_only_the_leader_writes_them` passed.
- AC-6 met: `verdict_assigned_auditor_is_the_only_one_accepted` passed.
- AC-7 met: `verdict_repeat_writes_no_second_row` passed.
- Red first: the seven tests failed on the code before slice 3 (unknown subcommands `verdict` and `assign`, and no `roles` in `card show`).
- Mutants (byte-copy restore, tree clean after): M1 author check of verdict off, M2 assigned-auditor check off, M3 repeat detection off, M4 author check of assign off, M5 review-state check of verdict off; all five killed, each by the test named for it.
- `moai check affected b68fce6b5c..60895b3320`: PASS, 69 targets, 900 test-target tests passed and none failed, lib 1630 passed, bins 28 and 28 passed; diff check, fmt, ops rustfmt, clippy, lib, bins, build and python steps exit 0.
- By the leader's word (decision on the lane's blocked note): the line `pub mod roles;` in `src/card/mod.rs` and one new row in the operations table, both files reserved by another card of this lane; new lines only. The roles answer of `card show` is described in that new row, so the existing `card show` row stays as it was.
- Not run: a full test run outside the affected set, the Windows type check, a mutant on the evidence-length check and on the identity shape check (the unit test pins both shapes).
- closure_status: open until the audit.
