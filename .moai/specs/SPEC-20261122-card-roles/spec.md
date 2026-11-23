---
id: SPEC-20261122-card-roles
title: "Card roles: each card names its authors, auditor and lander from ledger rows, and a verdict from an author is refused"
status: draft
revision: 3
created: 2026-11-22
---

# Card roles: each card names its authors, auditor and lander from ledger rows, and a verdict from an author is refused

## Intent

User decision of 2026-11-21 (stage 3 of the flow redesign, option (a)): there is no resident leader session; the author, the auditor and the lander of a card are recorded per card by the ledger. Today the ledger knows who held a card (the reservation rows) and who landed it (the `card.landed` row), but nothing says who audits, and nothing stops the session that implemented a card from writing the verdict on it. The rule "the agent that implemented a change never audits it" lives only in instructions.

This SPEC is slice 1 of the roles: the read side and the two refusals. A card shows its authors, its auditor and its lander, taken from ledger rows; an audit assignment never names an author; a verdict from a session that held the card is refused with the reason, and so is a verdict from a session that is not the assigned auditor. The typed submit and block rows and the inbox that lists unanswered rows are the next slices (the plan of the flow skeleton, steps 4a to 4f), and are not here.

Revision 3 (leader decision, note 7563 of card C-351, 2026-11-23, after the merge of the integration tip): SPEC-20261123-landing-reads-verdict landed the `card verdict` command and the `card.verdict` event (head, base, verdict, tier, evidence, receipt, reviewer) and `card land` reads them. Revision 1 of this SPEC defined a second `card verdict` with other options and other fields, so two writers of one event kind and two clap rows of one name could not stand together. This revision deletes the verdict command and the verdict event of revision 1 (REQ-3 and REQ-7; the record and its repeat rule are REQ-1 and REQ-4 of that SPEC) and keeps REQ-2 and REQ-6 as two refusals of that one command, made in the write transaction so that the ledger never holds a verdict the roles forbid. The landing keeps its own `self_verdict` check, which also reads commit trailers.

Grounded in the code at the integration tip `fd56877af5` before revision 1, and at `12701c4497` for revision 3:

- Holders are recorded per generation in the `reservation` table (`holder`, `status`, never deleted: `src/ledger/schema.rs:112`), and the holder is the session that took the card (`src/card/dispatch.rs`, `session_holder`).
- Events are free-text kinds with a JSON payload and the writing session as `actor` (`src/ledger/schema.rs:133`), and `reservation::add_event` is the one writer; a typed signal fits an event row and needs no schema change (the plan of 2026-11-21, constraints).
- The `card.review` row records the session that sent a card to review (`src/ledger/workflow.rs`, `review`), and the `card.landed` row records the head, the new tip and the caller word.
- `workflow::record_verdict` is the one writer of `card.verdict` and `workflow::reservation_holders` lists the holders of every generation (both of SPEC-20261123-landing-reads-verdict).

## Non-goals

- Tier floors, and holders landing their own cards: the landing's own checks (`self_verdict`, `lander_review`) of SPEC-20261123-landing-reads-verdict.
- The verdict command, its options and its event fields: that SPEC owns them.
- The typed submit row, the typed block and answer rows, `card inbox` and `card wait` on typed rows.
- Choosing an auditor automatically, or launching an audit session: this SPEC records an assignment someone made.
- Changing the schema, the version gate or any existing event kind.
- Treating a lane name as an identity: until the lane name becomes the holder (a later slice), the identity compared is the holder string recorded in the reservation.

## Requirements

- REQ-1: When the card show command is asked for a card, the card show command shall list the authors, the auditor and the lander of that card from ledger rows · failure: none — a role with no row is listed as absent, never guessed.
- REQ-2: When a session writes a verdict for a card, the card verdict command shall refuse the verdict if that session holds or ever held a reservation of the card · failure: none — the answer names the card and the reason, the ledger is unchanged, and the exit status is non-zero.
- REQ-4: When an audit assignment names an identity that holds or ever held a reservation of the card, the card assign command shall refuse the assignment · failure: none — the answer names the identity and the card, and the ledger is unchanged.
- REQ-5: When the card assign command accepts an assignment, the card assign command shall write one `card.audit_assigned` event that holds the auditor and the assigning caller · failure: none — only the leader word assigns, and a card that is not in review refuses before any write.
- REQ-6: When an auditor is assigned to a card, the card verdict command shall refuse a verdict from any other session · failure: none — the answer names the assigned auditor, and a card with no assignment accepts a verdict from any session that never held it.

## Acceptance

- AC-1 (REQ-1): a card with two reservation generations held by two sessions, one `card.audit_assigned` row and one `card.landed` row shows both holders as authors, the assigned identity as auditor and the landing's caller and head as lander; a card with none of these rows shows each role as absent · verify: `cargo test --locked --offline --test card_roles_cli roles_`
- AC-2 (REQ-2): the session that took a card and sent it to review writes a verdict, and the command refuses with a message that names the card and says the writer held it; a session that held an earlier generation is refused the same way; the event rows of the ledger are unchanged, and a session that never held the card still records its verdict · verify: `cargo test --locked --offline --test card_verdict_cli verdict_author_`
- AC-4 (REQ-4): an assignment that names the holder of any generation refuses and writes nothing · verify: `cargo test --locked --offline --test card_roles_cli assign_author_`
- AC-5 (REQ-5): the leader's assignment of a session that never held the card writes one row; another caller word and a card that is not in review each refuse before any write · verify: `cargo test --locked --offline --test card_roles_cli assign_rows_`
- AC-6 (REQ-6): after an assignment, a verdict from another session that never held the card refuses naming the assigned auditor, and the assigned auditor's verdict is accepted · verify: `cargo test --locked --offline --test card_verdict_cli verdict_assigned_`

## Scope

- write: `.moai/specs/SPEC-20261122-card-roles/spec.md`
- write (slices 2 and 3, on this card's own widening): `src/card/roles.rs` (the role reader and the assign action), `src/card/mod.rs` (one `pub mod` line), `src/ledger/workflow.rs` (the `AUDIT_ASSIGNED_KIND` constant and the two refusals of `record_verdict`, as their own function), `tests/card_roles_cli.rs`, `tests/card_verdict_cli.rs` (the two refusals), `docs/design-review/appendix/operations/card.md` (the assign row and the show row)
- write by line-lease or the leader's word: `src/ops/card.rs` (the clap row of `assign`, and the roles in the show answer), because another card reserves that file
- widened by the leader (note 7563) and not needed in the end: `src/card/verdict.rs`
- read-only: `src/ledger/schema.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/verdict.rs`
- forbidden: the schema, the version gate, `src/main.rs`, `src/help.rs`, any existing event kind

## Plan

1. This SPEC, read by the leader and by the audit before any code.
2. Slice 2, tests first (RED): `tests/card_roles_cli.rs` with `roles_` and `assign_` against a temporary home and a real ledger; the sessions are made with the host's session variable, one per role.
3. Slice 3: `src/card/roles.rs`; make the tests pass; the show answer gains a `roles` object.
4. Slice 4: the clap row of assign through the line-lease or the leader's word, the operations table, the SPEC Result.
5. Revision 3, after the merge of the integration tip: delete the verdict writer of revision 1, RED tests `verdict_author_` and `verdict_assigned_` in `tests/card_verdict_cli.rs`, then the two refusals in `record_verdict`.
6. Per slice: light checks, at most the mutant cap, one affected run, then the READY note with head and base only.

## Risks

- A session id is not a person: two sessions of one person count as two identities, so the rule stops a session from judging its own card and does not stop a person who opens a second session. That is the rule the instructions already state, now written in the ledger.
- The holder of a card can be a lane name once a later slice makes the lane the holder; the comparison reads the recorded holder string through `trailers::same_identity`, so old and new reservations both count and a legacy bare session id equals its full value. A card whose holder changed name between generations lists both.
- A verdict row written before the refusals existed is still read by the landing; the refusals stop new rows only, and the landing's `self_verdict` check stays the last guard.
- `src/ops/card.rs` is reserved by another card; the clap rows are small additions, and the earlier slices of this lane were done through the leader's word for the same reason.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | An author is any session that holds or ever held a reservation of the card, every generation, released or not | The rule says the agent that implemented a change never audits it; every holder may have written part of it | Narrow the query to the generation that sent the card to review |
| D-2 | The roles are read from existing rows, and only the assignment is a new event kind, with no schema change | The plan of 2026-11-21 puts typed signals in event rows, and the version gate refuses a schema change | Stop writing the kind; old readers ignore unknown kinds |
| D-3 | Only the leader word assigns an auditor | The leader is the one who dispatches audits (plan decision D3) | Allow the holder-independent caller list to grow |
| D-6 | One verdict command: the one of SPEC-20261123-landing-reads-verdict, with the two refusals of this SPEC added in its write transaction; the verdict command, event fields, review-state rule (old D-4) and full-head rule (old D-5) of revision 1 are deleted | Leader decision, note 7563 of card C-351, 2026-11-23: two writers of one event kind cannot stand together, and a refusal made at write time keeps a forbidden verdict out of the ledger | Option (b) of note 7561: drop the write-time refusals and keep only the landing's `self_verdict` |

## Result

Slices 2 and 3, revision 3, on the branch after the merge of the integration tip `12701c4497`. Numbers come from runs made for this result in the worktree of the card.

- AC-1 met: `roles_are_read_from_reservations_the_assignment_and_the_landing` passed (`card_roles_cli`, 3 of 3 passed with the two assign tests).
- AC-2 met: `verdict_author_sessions_are_refused_with_the_reason` passed (`card_verdict_cli`): the current holder and an earlier holder are refused, the card and `held` are named, the ledger is unchanged, and a session that never held the card still records one verdict.
- AC-4 met: `assign_author_identities_are_refused` passed.
- AC-5 met: `assign_rows_hold_the_auditor_and_the_caller_and_only_the_leader_writes_them` passed.
- AC-6 met: `verdict_assigned_auditor_is_the_only_one_accepted` passed: another session is refused with the assigned auditor named, the assigned auditor is accepted.
- Red first: before the refusals, both new tests of `card_verdict_cli` failed because the verdict was written (exit 0 where 1 was expected).
- Guard mutants (tree restored by the inverse edit, `git status` clean after): M1 holder refusal off (`if false && …`) killed by `verdict_author_` (exit 101); M2 assigned-auditor refusal off killed by `verdict_assigned_` (exit 101). The assign-side mutants of revision 1 were not run again: that code is unchanged by the merge.
- Regression runs on the merged tree: `card_land_cli` 60 of 60 and `card_finish_cli` 45 of 45 passed, `card_verdict_cli` 5 of 5, `card_roles_cli` 3 of 3.
- The `moai check affected` receipt of the head is in the READY note of the card.
- By the leader's word (decision of 2026-11-22 on the lane's blocked note, and note 7563): the line `pub mod roles;` in `src/card/mod.rs` and the assign row of the operations table; revision 3 also drops the verdict row of revision 1 from that table, where `card verdict` already has the row of SPEC-20261123-landing-reads-verdict.
- Not run: a full test run outside the affected set, the Windows type check, a mutant on the identity shape check (the unit test pins the shapes).
- closure_status: open until the audit.
