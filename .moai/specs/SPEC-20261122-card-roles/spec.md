---
id: SPEC-20261122-card-roles
title: "Card roles: each card names its authors, auditor and lander from ledger rows, and a verdict from an author is refused"
status: draft
revision: 5
created: 2026-11-22
approval: {by: leader, at: 2026-11-23, basis: "leader decision note 7720 on card C-415 (2026-11-23T17:22:40Z), answering the request of note 7719, option (a); leader judgment under the user's delegation of 2026-11-23T13:57Z, not a direct user pick", digest: "sha256:66dd2a069505880dbf961586ffcce76ea897aa085875235b36f81282fd6ceb29"}
---

# Card roles: each card names its authors, auditor and lander from ledger rows, and a verdict from an author is refused

## Intent

User decision of 2026-11-21 (stage 3 of the flow redesign, option (a)): there is no resident leader session; the author, the auditor and the lander of a card are recorded per card by the ledger. Today the ledger knows who held a card (the reservation rows) and who landed it (the `card.landed` row), but nothing says who audits, and nothing stops the session that implemented a card from writing the verdict on it. The rule "the agent that implemented a change never audits it" lives only in instructions.

This SPEC is slice 1 of the roles: the read side and the two refusals. A card shows its authors, its auditor and its lander, taken from ledger rows; an audit assignment never names an author; a verdict from a session that held the card is refused with the reason, and so is a verdict from a session that is not the assigned auditor. The typed submit and block rows and the inbox that lists unanswered rows are the next slices (the plan of the flow skeleton, steps 4a to 4f), and are not here.

Revision 3 (leader decision, note 7563 of card C-351, 2026-11-23, after the merge of the integration tip): SPEC-20261123-landing-reads-verdict landed the `card verdict` command and the `card.verdict` event (head, base, verdict, tier, evidence, receipt, reviewer) and `card land` reads them. Revision 1 of this SPEC defined a second `card verdict` with other options and other fields, so two writers of one event kind and two clap rows of one name could not stand together. This revision deletes the verdict command and the verdict event of revision 1 (REQ-3 and REQ-7; the record and its repeat rule are REQ-1 and REQ-4 of that SPEC) and keeps REQ-2 and REQ-6 as two refusals of that one command, made in the write transaction so that the ledger never holds a verdict the roles forbid. The landing keeps its own `self_verdict` check, which also reads commit trailers.

Revision 4 (card C-411, from the non-blocking findings N1 to N3 of the audit of C-351): `card assign` compared the named auditor with the recorded holders as exact strings, so a spelling of an author that differs in ASCII case or by a `:<12hex>` worktree suffix passed the check and then every verdict was refused (the verdict writer compares through `trailers::same_identity`). This revision makes the assignment use the same comparison (REQ-4) and refuses an auditor value that is not a session identity, a lane name for one (REQ-8), because no verdict writer can ever match it. The weak `>=` check of the AC-5 test becomes an exact count.

Revision 5 (card C-415, from the non-blocking findings N1 to N3 of the audit of C-411): the two refusals of revision 4 are real but the help of `card assign` and its row in the operations table do not say them (REQ-9), and the refusal of an empty, spaced or over-long value does not tell the reader the shape a value needs (REQ-10). Only text changes: the help text, one document row, one refusal message and a blank line of this file (N1); no check, no comparison and no event changes.

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
- REQ-4: When an audit assignment names an identity that `trailers::same_identity` finds equal to the holder of any generation of the card, the card assign command shall refuse the assignment · failure: none — the answer names the identity and the card, and the ledger is unchanged.
- REQ-5: When the card assign command accepts an assignment, the card assign command shall write one `card.audit_assigned` event that holds the auditor and the assigning caller · failure: none — only the leader word assigns, and a card that is not in review refuses before any write.
- REQ-6: When an auditor is assigned to a card, the card verdict command shall refuse a verdict from any other session · failure: none — the answer names the assigned auditor, and a card with no assignment accepts a verdict from any session that never held it.
- REQ-8: When an audit assignment names a value that is not a session identity (a session id with an optional `:<12hex>` worktree key, ASCII case ignored), the card assign command shall refuse the assignment · failure: none — the answer names the value and the shape it needs, and the ledger is unchanged.
- REQ-9: When the help of the card assign command is asked for, the card assign command shall state that another spelling of an author is refused and that a value which is not a session identity, a lane name for one, is refused · failure: none — the operations table row of the command says the same.
- REQ-10: When an audit assignment names an empty value, a value with whitespace or a value over 100 characters, the card assign command shall refuse the assignment with an answer that states the session identity shape it needs · failure: none — the ledger is unchanged.

## Acceptance

- AC-1 (REQ-1): a card with two reservation generations held by two sessions, one `card.audit_assigned` row and one `card.landed` row shows both holders as authors, the assigned identity as auditor and the landing's caller and head as lander; a card with none of these rows shows each role as absent · verify: `cargo test --locked --offline --test card_roles_cli roles_`
- AC-2 (REQ-2): the session that took a card and sent it to review writes a verdict, and the command refuses with a message that names the card and says the writer held it; a session that held an earlier generation is refused the same way; the event rows of the ledger are unchanged, and a session that never held the card still records its verdict · verify: `cargo test --locked --offline --test card_verdict_cli verdict_author_`
- AC-4 (REQ-4): an assignment that names the holder of any generation, in the same spelling or in another (upper case, or with a `:<12hex>` worktree suffix), refuses and writes nothing · verify: `cargo test --locked --offline --test card_roles_cli assign_author_`
- AC-7 (REQ-8): an assignment that names a lane name or another value that is not a session identity refuses with the value and the needed shape and writes nothing, and a session identity with a worktree suffix is accepted · verify: `cargo test --locked --offline --test card_roles_cli assign_identity_`
- AC-8 (REQ-9): `moai card assign --help` names the refusal of another spelling of an author (ASCII case and the `:<12hex>` suffix), the refusal of a value that is not a session identity with a lane name as the example, and the shape a session identity has; the first sentence still starts with `Writes the ledger.` · verify: `cargo test --locked --offline --test card_roles_cli assign_help_`
- AC-9 (REQ-10): an assignment that names an empty value, a value with a space or a tab, or a value of 101 characters refuses with an answer that says `session id` and the optional `:` and 12 hex digits, and writes nothing · verify: `cargo test --locked --offline --test card_roles_cli assign_value_`
- AC-5 (REQ-5): the leader's assignment of a session that never held the card writes one row; another caller word and a card that is not in review each refuse before any write · verify: `cargo test --locked --offline --test card_roles_cli assign_rows_`
- AC-6 (REQ-6): after an assignment, a verdict from another session that never held the card refuses naming the assigned auditor, and the assigned auditor's verdict is accepted · verify: `cargo test --locked --offline --test card_verdict_cli verdict_assigned_`

## Scope

- write: `.moai/specs/SPEC-20261122-card-roles/spec.md`
- write (slices 2 and 3, on this card's own widening): `src/card/roles.rs` (the role reader and the assign action), `src/card/mod.rs` (one `pub mod` line), `src/ledger/workflow.rs` (the `AUDIT_ASSIGNED_KIND` constant and the two refusals of `record_verdict`, as their own function), `tests/card_roles_cli.rs`, `tests/card_verdict_cli.rs` (the two refusals), `docs/design-review/appendix/operations/card.md` (the assign row and the show row)
- write by line-lease or the leader's word: `src/ops/card.rs` (the clap row of `assign`, and the roles in the show answer), because another card reserves that file
- write (revision 5, card C-415): the `assign` help in `src/ops/card.rs`, the first refusal of `assign` in `src/card/roles.rs`, `tests/card_roles_cli.rs`, the assign row of `docs/design-review/appendix/operations/card.md`
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
| D-7 | `card assign` compares the named auditor with the holders through `trailers::same_identity` and refuses a value whose shape is not a session identity (a lane name is refused, not warned) | Audit of C-351 (AUDIT #351, N1 and N2): an exact-string comparison lets another spelling of an author through, and a value that the verdict writer can never match leaves the card unauditable; a refusal at assignment is the only point where nothing is written yet. Card C-411 asks for a refusal or a warning, and the refusal is the one that cannot leave a card blocked | Warn instead of refuse for the shape check: the row is written and the answer carries the warning |
| D-8 | The refusal of a malformed value and the help of `card assign` share one sentence for the shape (a session id of hex digits and dashes, 8 to 64 characters, optionally followed by `:` and 12 hex digits); no check changes | AUDIT #411 N2 and N3 (non-blocking): the refusals exist but the help, the operations row and the first refusal do not teach the shape, so the leader meets the refusal before the rule | Restore the earlier texts; no data depends on them |

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

Revision 4 (the card of the non-blocking audit findings N1 to N3). Approval basis: the leader's decision note 7632 on that card (2026-11-23T15:28:55Z), which approves this revision with the refusal of option (a) of note 7631. Numbers come from runs made for this result in the worktree of the card.

- AC-4 met (REQ-4, now through `trailers::same_identity`): `assign_author_identities_are_refused` and the new `assign_author_spellings_are_refused` passed (the upper-case spelling, the `:<12hex>` suffix and both together are refused, nothing is written, a session that never held the card is still accepted).
- AC-7 met (REQ-8): the new `assign_identity_shapes_are_checked_with_a_reason` passed: `agent-4`, `lane-3`, `abc`, a bad suffix and a short suffix are refused naming the value and the word `session`, and a session id with a worktree suffix is accepted and stored as given; the lib unit test `session_identities_are_session_ids_with_an_optional_worktree_key` pins the shapes.
- AC-5 stays met with an exact count: the `>=` check of `assign_rows_…` is now `assert_eq!` on the event count before and after each refusal (finding N3).
- Red first: before the code change both new tests failed because the assignment was written (exit 0 where 1 was expected).
- Guard mutants (each undone by the inverse edit, `git status` clean after): M1 exact-string holder comparison back in `assign` killed by `assign_author_spellings_are_refused` (exit 100); M2 shape check off (`if false && …`) killed by `assign_identity_shapes_are_checked_with_a_reason` (exit 100).
- closure_status: open until the audit.

Revision 5 (the card of the non-blocking audit findings N1 to N3 of the audit of C-411). Approval basis: the leader's decision note 7720 on card C-415 (2026-11-23T17:22:40Z), option (a) of the request in note 7719; recorded in the front matter `approval`. Numbers come from runs made for this result in the worktree of the card.

- AC-8 met (REQ-9): `assign_help_states_the_refusals_and_the_shape` passed: the help starts with `Writes the ledger.` and names ASCII case, `:<12hex>`, a lane name, `agent-4` and the shape `hex digits and dashes, 8 to 64 characters`. The words are compared after joining whitespace, because clap wraps the help at 100 columns. The operations row of `card assign` says the same (`operations_doc_matches_code` passed).
- AC-9 met (REQ-10): `assign_value_refusals_state_the_needed_shape` passed: an empty value, a space, two words, a tab and 101 characters are refused with `session id` and `12 hex digits`, and the event rows are unchanged.
- N1 done: the blank line before the revision 4 paragraph of this Result. The new tests read no document, so `tools/repo_scans.json` needs no entry; `operations_doc_matches_code`, which reads the operations folder, is already registered there.
- Red first: both new tests failed before the change (exit 100): the help lacked the words and the empty-value refusal lacked `session id`.
- Guard mutants (each undone by the inverse edit, `git status` clean after): M1 the lane-name sentence removed from the help of `assign` killed by `assign_help_states_the_refusals_and_the_shape` (exit 100); M2 the empty-value refusal back to its earlier text killed by `assign_value_refusals_state_the_needed_shape` (exit 100).
- closure_status: open until the audit.
