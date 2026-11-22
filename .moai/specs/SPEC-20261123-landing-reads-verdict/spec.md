---
id: SPEC-20261123-landing-reads-verdict
title: "Landing reads the audit verdict: one command lands a head only on an independent PASS of that exact head and closes the card"
status: draft
revision: 2
created: 2026-11-23
---

# Landing reads the audit verdict: one command lands a head only on an independent PASS of that exact head and closes the card

## Intent

`moai card land` already moves the integration branch to a candidate head by one compare-and-swap and closes the card in the same call (SPEC-20261121-card-land). It reads no audit verdict. The audit contract says the lander reads the verdict note "before it runs `moai card land`", and that reading is done by eye on free note text. Observed in the flow survey of build 19 (the card body; not measured again here):

- the leader landed by five git commands and a separate `card finish`, and landed cards stayed open;
- `card land` accepts a head whose only PASS was written by the author of the change;
- 38 percent of verdicts are FAIL, and a READY note waits a median of 12.7 minutes for its first answer.

This SPEC makes the verdict a record the tool reads:

- a new command `moai card verdict` appends a structured verdict record (card, head, base, verdict, tier, evidence, receipt, reviewer) to the ledger;
- `moai card land` refuses unless the exact candidate head has an independent PASS in that record, and then lands, records the landing with the verdict it relied on, closes the card when every criterion is met and releases the reservation.

Who may land, and whether the lander may also be the reviewer, is the user's decision (U-1, pick A of revision 2): only the leader lands, and the leader reviews only at tier 0. The mechanism records both identities, so another answer is a small change.

Tiers are the four classes of SPEC-20261123-audit-rule-text (docs-only 0, test-only 0 or 1, code 1, high-risk 2). This SPEC reads the tier the way `src/gate.rs` `tier_of` does and adds no tier.

### Sources (read whole, 2026-11-22T22:2xZ, at tip 124a31c2b7)

- The card body and its four criteria (`moai card show`).
- `src/card/land.rs`, `src/card/finish.rs`: the call order of the landing, `finish_preflight`, `record_land_started`, the `already_done` answer.
- `src/ledger/schema.rs` and `src/ledger/workflow.rs`: notes and events are append-only; events have a free `kind` and a unique `event_key`; the events in use are `card.landed`, `card.land_started`, `card.finish` and `card.tip_restored`.
- `src/audit_gate.rs`: the author rule (`Moai-Session:` trailers of the commits, fail closed when none), the FAIL rule (a posted FAIL stops every tier until the same auditor posts a later PASS), `sessions_of`.
- `src/trailers.rs`: the `Moai-Session` value is `<session id>:<12 hex worktree key>` or a legacy bare id; `same_identity`.
- `skills/moai/references/audit.md`: "Who records the verdict" and "Audit tiers"; SPEC-20261123-audit-rule-text: the four tier lines.
- `CLAUDE.md` workflow steps 5 and 7: tier 0 needs `ci` alone, tier 1 or 2 also needs a non-author PASS, a posted FAIL stops every tier (all for pull requests).

## Non-goals

- No change to the pull request path: `audit-gate`, `gate-check` and the merge gate stay as they are. This SPEC covers the local landing on the integration branch only.
- No change of who may call `card land` in this SPEC: the caller word stays `leader` (U-1, pick A).
- No new ledger table or column and no migration: the record is a free event kind (D-1).
- No machine-made check receipt: `check affected` keeps printing its receipt as text; the reviewer cites it in the verdict record (D-6). A receipt written by the check itself is another card.
- No audit of the change itself: the command records and reads a verdict and judges nothing about the code.
- No edit of the audit contract text, the AUDIT workflow or the audit agent. They say "add one `moai card note`"; that sentence changes in a later text slice after this SPEC is approved (Plan step 6).

## Requirements

<!-- This SPEC adds a check that blocks: it starts in observe (REQ-14). Record the would-block count and the false positives in the Result, and leave the switch to enforce to a recorded decision (U-2). -->
- REQ-1: When a session runs the card verdict command with a card and a head and a base and a verdict word and a tier and one line of evidence, the card verdict command shall append one `card.verdict` event to the card that holds those values and the reviewer identity.
- REQ-2 (REQ-1): IF the head or the base is not 40 lowercase hex or names no commit or the base is not an ancestor of the head or the evidence is blank or holds a secret word or the session has no host session id, THEN the card verdict command shall refuse and write nothing.
- REQ-3: When the card verdict command is given the receipt of a check run, the card verdict command shall store the head the receipt names and each step name with its exit code in the event · failure: none — a receipt with no head or a step with no exit code is one of the refusals of the record call and is checked by AC-3.
- REQ-4: When the card verdict command is run again with the same card and head and base and verdict and tier and evidence and reviewer, the card verdict command shall write nothing and answer the event it found · failure: none — a repeat has no step that can fail; a ledger that cannot be opened is the ledger's own refusal.
- REQ-5: When the card verdict command takes the reviewer identity, the card verdict command shall read it from the host session id of the running session and never from an option · failure: none — a session with no host session id is refused by REQ-2.
- REQ-6: When the card land command is called, the card land command shall read the verdict events of the card for the candidate head and decide the landing from them before it writes anything.
- REQ-7 (REQ-6): IF the card has no verdict event at all, THEN the card land command shall refuse with the code `no_verdict` and write nothing.
- REQ-8 (REQ-6): IF the newest verdict of any one reviewer on the candidate head is FAIL, THEN the card land command shall refuse with the code `verdict_fail` and write nothing.
- REQ-9 (REQ-6): IF the card has verdict events but none names the candidate head, THEN the card land command shall refuse with the code `verdict_other_head` that names the newest verdict head and write nothing.
- REQ-10 (REQ-6): IF every PASS on the candidate head comes from a reviewer that authored a commit of the range from the integration tip to the head that is not a clean join or that holds or held a reservation of the card, THEN the card land command shall refuse with the code `self_verdict` and write nothing.
- REQ-11 (REQ-10): IF no commit of that range that counts as authored carries a `Moai-Session` trailer, THEN the card land command shall refuse with the code `authors_unknown` and write nothing.
- REQ-12 (REQ-6): IF the tier is 1 or 2 and the PASS carries no receipt for the candidate head with at least one step and every step at exit 0, THEN the card land command shall refuse with the code `receipt_missing` and write nothing.
- REQ-13: When the card land command lands a head, the card land command shall name the verdict event it relied on and its reviewer and its tier and the lander in the start row and in the landing event · failure: none — the refusals that come before the start row are REQ-7 to REQ-12 and a refused start row writes no landing.
- REQ-14: While the verdict check is in observe mode, the card land command shall answer the refusal it would have made as `would_refuse` with its code and then continue as it did before this SPEC.
- REQ-20 (REQ-6): IF the only PASS on the candidate head is by the session that runs the landing and the tier is 1 or 2 or the landing has no session, THEN the card land command shall refuse with the code `lander_review` and write nothing.
- REQ-21: When a commit of the range from the integration tip to the head is a merge commit whose tree equals the clean merge of its parents, the card land command shall not count that commit as authored · failure: none — a merge commit with another tree and a merge that git cannot compute count as authored.
- REQ-22: When the project profile names `work_flow.land_verdict` as `enforce` or `observe`, the card land command shall apply that mode · failure: none — a profile that names none means observe and another word is REQ-23.
- REQ-23 (REQ-22): IF `work_flow.land_verdict` names another word, THEN the profile reader shall refuse the profile as an invalid work_flow setting.
- REQ-15: When the card land command is run again for a head whose start row names a verdict event, the card land command shall use that event and not read the verdicts again · failure: none — the checks that do not depend on the verdicts (the SPEC at the head and the criteria and the generation) are made again as before.
- REQ-16: When the card is done with the same landed head, the card land command shall answer done and write nothing whatever the verdict events say now · failure: none — another landed head refuses as in SPEC-20261121-card-land.
- REQ-17: When a FAIL is recorded for a head after the integration branch moved to it, the card land command shall still close the card and name the late FAIL in its answer · failure: none — the close has the refusals of the finish command and none of them reads a verdict.
- REQ-18: When the card land command is stopped at any point of a landing, the card land command shall leave only a state that the same call run again completes with one landing and one finish event · failure: none — a close refused after the move leaves the branch moved and says so as in SPEC-20261121-card-land.
- REQ-19: When the card land command decides a landing, the card land command shall read no card note text as a verdict · failure: none — a head whose only verdict is a note has no verdict event and is REQ-7.

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-5): a PASS recorded by a session with a host session id holds the card, head, base, verdict, tier, evidence and that session as reviewer; the same call again writes no second event and answers the first; an option that names another reviewer does not exist and is refused by the command line · verify: `cargo test --locked --offline --test card_verdict_cli ac1_`
- AC-2 (REQ-2): a short head, a head that is no commit, a base that is not an ancestor of the head, blank evidence, evidence with a secret word and a call with no host session id each refuse and the ledger is as it was · verify: `cargo test --locked --offline --test card_verdict_cli ac2_`
- AC-3 (REQ-3): a receipt given with a head and two steps is stored with the head and each step name and exit code, and a receipt given with a step that has no exit code is refused · verify: `cargo test --locked --offline --test card_verdict_cli ac3_`
- AC-4 (REQ-7, REQ-8, REQ-9, REQ-10): the four refusals of the card each answer their own code with the branch and the ledger unchanged and not even a start row: a head with no verdict (`no_verdict`), a head with a FAIL (`verdict_fail`), a PASS recorded for an older head only (`verdict_other_head`) and a PASS by a session that wrote a commit of the head (`self_verdict`) · verify: `cargo test --locked --offline --test card_land_cli ac40_`
- AC-5 (REQ-8): a FAIL by one reviewer and a later PASS by another reviewer still refuses with `verdict_fail`, and the same reviewer's later PASS lands · verify: `cargo test --locked --offline --test card_land_cli ac41_`
- AC-6 (REQ-10): a PASS by the holder of any reservation of the card refuses with `self_verdict` even when no commit carries the holder's trailer, and a legacy bare session id in a trailer is the same session as its full value · verify: `cargo test --locked --offline --test card_land_cli ac42_`
- AC-7 (REQ-11): a range of commits with no `Moai-Session` trailer refuses with `authors_unknown` · verify: `cargo test --locked --offline --test card_land_cli ac43_`
- AC-8 (REQ-12): at tier 1 a PASS with no receipt, a receipt for another head, a receipt with no step and a receipt with a step at a non-zero exit each refuse with `receipt_missing`, and a tier 0 PASS with no receipt lands; the tier is the higher of the stated tier and the floor of the changed paths, so a head that changes a tier 2 path with a PASS stating tier 0 is judged at tier 2 · verify: `cargo test --locked --offline --test card_land_cli ac44_`
- AC-9 (REQ-6, REQ-13): a head with an independent PASS lands in one call: the branch moves by compare-and-swap, the start row and the landing event name the verdict event and the reviewer and the tier and the lander, a card with every criterion met closes and its reservation is released · verify: `cargo test --locked --offline --test card_land_cli ac45_`
- AC-10 (REQ-16): a second call on the landed card with the same head answers done and writes nothing, also after a FAIL was recorded for that head, and another head refuses as before · verify: `cargo test --locked --offline --test card_land_cli ac46_`
- AC-11 (REQ-15, REQ-17): a landing that stopped after the move is closed by the same call run again from the verdict the start row names, and a FAIL recorded between the two calls does not stop the close and is named as `late_fail` in the answer · verify: `cargo test --locked --offline --test card_land_cli ac47_`
- AC-12 (REQ-18): the call killed after the start row and before the move, and killed after the move and before the close, leaves one of the states of SPEC-20261121-card-land, and a fresh process run again ends with one start row and one landing event and one finish event (the close is one write transaction, so it has no stop point of its own) · verify: `cargo test --locked --offline --test card_land_cli ac48_`
- AC-13 (REQ-18): a lock file on the branch reference refuses with git's reason and leaves the start row, and the call run again after the lock is gone lands once · verify: `cargo test --locked --offline --test card_land_cli ac49_`
- AC-14 (REQ-6, REQ-18): a moved integration branch gives three outcomes: the branch moved between the tip read and the update refuses with `tip_moved` and keeps the verdict valid for a later call on the same head; the branch advanced past the base of the candidate refuses with `stale_base` and a PASS of the old head does not carry to the head that merged the new tip; the branch moved to contain the head closes the card with no second move · verify: `cargo test --locked --offline --test card_land_cli ac50_`
- AC-15 (REQ-19): a card note whose first line is a well-formed `AUDIT … verdict: PASS` line for the head, with no verdict event, refuses with `no_verdict` · verify: `cargo test --locked --offline --test card_land_cli ac51_`
- AC-16 (REQ-14): in observe mode a head with no verdict lands as it did before and the answer carries `would_refuse` with the code `no_verdict`, and in enforce mode the same call refuses · verify: `cargo test --locked --offline --test card_land_cli ac52_`
- AC-19 (REQ-20): the landing session's own PASS lands a tier 0 head, refuses with `lander_review` at tier 1 and at tier 2 (also when the floor of the paths raises a tier 0 PASS), and a landing with no session refuses at tier 1 · verify: `cargo test --locked --offline --test card_land_cli ac54_`
- AC-20 (REQ-10, REQ-21): the lander's clean merge commit does not make the lander an author, a merge commit with another tree and a merge that git cannot compute do, and a lane commit still does · verify: `cargo test --locked --offline --test card_land_cli ac55_`
- AC-21 (REQ-22, REQ-23): the profile with `land_verdict` `enforce` refuses a head with no verdict, the profile with `observe` or with none answers `would_refuse` and lands, and another word makes every card command that reads the profile refuse it · verify: `cargo test --locked --offline --test card_land_cli ac56_` and `cargo test --locked --offline --lib profile::tests::land_verdict`
- AC-17 (REQ-9, REQ-13): the landing event and the start row hold the reviewer and the lander as two values, so a rule that bars the lander from reviewing can be read from the record without a new field · verify: `cargo test --locked --offline --test card_land_cli ac53_`
- AC-18 (REQ-6): every test of `card_land_cli` and `card_finish_cli` that landed a card before this SPEC keeps passing with a verdict recorded in its fixture and the observe or enforce mode named · verify: `cargo test --locked --offline --test card_land_cli --test card_finish_cli`

## Scope

- write (this slice): `.moai/specs/SPEC-20261123-landing-reads-verdict/**`
- write (code slice, after approval):
  - `src/card/verdict.rs` (new), `src/card/land.rs`, `src/card/land/check.rs` (new, the verdict check of the landing), `src/card/mod.rs` (one `pub mod` line);
  - `src/ledger/workflow.rs` (the verdict record, the start row and the landing event carry the verdict values; a read helper for verdict events);
  - `src/profile.rs` (the `work_flow.land_verdict` key);
  - `src/ops/card.rs` (the clap lines of `card verdict` only, last step, as the land command had; `card land` gets no new option);
  - `tests/card_verdict_cli.rs` (new), `tests/card_land_cli.rs`, `docs/design-review/appendix/operations/card.md` (one table row).
- `src/card/mod.rs`, `src/ops/card.rs` and the docs file are reserved for another card while this slice starts; they are written after the leader frees them (the leader's answer on the card).
- read-only: `src/audit_gate.rs` (`sessions_of` is reused as it is; it changes only if its signature must be widened), `src/gate.rs` (`tier_of`), `src/trailers.rs` (`same_identity`, `value_shape`), `src/card/finish.rs`.
- The card's own write scope lists `src/card`, `src/audit_gate.rs`, `tests`, `docs` and `.moai/specs` and not `src/ledger/workflow.rs` or `src/ops/card.rs`; the code slice needs both added by the leader (`card scope`).
- forbidden: the pull request gates (`src/gate.rs` and `src/ops/gate*`), any skill or agent text, the ledger schema (`src/ledger/schema.rs`, `src/ledger/migrate.rs`).

## Plan

1. The SPEC (landed as draft); revision 2 records the user's picks U-1 to U-3 as decisions and no approval record.
2. After approval, tests first (RED): `tests/card_verdict_cli.rs` AC-1 to AC-3, then the new cases of `tests/card_land_cli.rs` AC-4 to AC-17 on a temporary home and a temporary repository, with the git wrapper the existing land tests use for the stop points.
3. `src/card/verdict.rs` and the verdict read helper of `src/ledger/workflow.rs`; the event key is a digest of the card and head and base and verdict and tier and evidence and reviewer, so a repeat finds its own event (REQ-4).
4. `src/card/land.rs`: the verdict check as a step of the order after the SPEC check and before the first write, the same step in `finish_preflight` so the start row transaction checks again (a race can add a FAIL in between), and the start row and the landing event carry the verdict values.
5. The mode of the check (`work_flow.land_verdict`, `observe` or `enforce`) read from the project profile; the first lands run in observe and the would-block count and false positives go into the Result (U-2).
6. Text slice, after the code lands: `skills/moai/references/audit.md` ("Who records the verdict"), `skills/moai/workflows/audit.md` and `run.md` say "run `moai card verdict`" in place of "add one `moai card note`", in the same rule text as SPEC-20261123-audit-rule-text. Tier 2 paths: audit by the leader's assignment.
7. Targeted tests, then one `check affected`, then READY with head, base and the receipt.

## Risks

- R-1: A lander-made merge commit carries the lander's `Moai-Session`. By U-1 a clean join does not count as authorship (REQ-21, D-10); a merge with another tree does. `git merge-tree --write-tree` needs git 2.38 or newer; an older git makes every merge count as authored, so the lander cannot review even tier 0 there.
- R-2: A lane that forgot its trailer is caught only through the reservation holders (REQ-10). A head written by a session that never held the card and wrote no trailer reads as `authors_unknown` and refuses (REQ-11).
- R-3: The verdict event is written by a local command into the local ledger; anyone who can run `moai` as the user can record one. The rule does not stop a forged verdict of the user's own account. It stops the mistakes seen: a missing, stale, failed or self-written verdict, and a missing receipt.
- R-4: Two stores (git and the ledger) cannot change in one transaction. REQ-15 and REQ-18 keep the cost the same as before this SPEC: a landing stopped after the move is closed by the same call run again.
- R-5: Every existing land test needs a verdict in its fixture (AC-18); a missing fixture shows as `no_verdict` and is the first thing the code slice meets.
- R-6: A late FAIL after the move does not undo it (REQ-17). The answer names it so the leader can open a repair; undoing a landing is `card drift --restore`, the user's choice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The verdict record is a ledger event of the free kind `card.verdict`, with no new table and no migration. | Events are append-only with a unique `event_key`, so a repeat is found by key; notes are free text and the card says the reader must not parse note text. | A table of its own and a migration in a later SPEC. |
| D-2 | The reviewer is the session that records the verdict, taken from the host session id, never from an option. | `audit.md` "Who records the verdict": the recording session is the reviewer of record; an option could be typed to look independent. | Add an option with a check against the host id. |
| D-3 | Independence compares the session-id part of `Moai-Session` (before the colon) with every author of the commits from the tip to the head, and with every holder of any reservation of the card. | `trailers::same_identity` treats a legacy bare id as every value of its session; a session that reviews from another worktree is still the author. | Compare the full value only. |
| D-4 | One reviewer's newest FAIL on the head refuses, whoever else passed it. | `audit_gate` rule 2 and `CLAUDE.md` workflow 7: a posted FAIL stops the merge until the same auditor posts a later PASS. The card text says "the newest verdict for the exact head"; the two read the same except when two reviewers disagree on one head, and then this SPEC follows the gate. | Read only the newest verdict of the head. |
| D-5 | The tier is the higher of the stated tier and the floor of the changed paths (`tier_of`). | The audit rule text SPEC: the stated tier never goes below the floor. | None needed; it only raises. |
| D-6 | The receipt is cited in the verdict record by the reviewer (head, steps with exit codes); the tool checks its head and its exit codes, not that a run happened. | No machine receipt exists today: `check affected` prints text and the ready note quotes it. | A receipt written by `check affected` itself, another card. |
| D-7 | The verdicts are judged once per landing and the verdict event is named in the start row; a rerun uses it. | A rerun after the move must not depend on verdicts recorded after the move (REQ-15, REQ-17). | Read the verdicts at each call. |
| D-8 | The pull request gates are unchanged. | The merge gate, `audit-gate` and tier 0's "no open FAIL" rule stay for pull requests; the local landing has its own rule. | None. |

### Decisions of the user (revision 2)

The user's picks of 2026-11-22T22:34:35Z, selections that name this SPEC and were relayed in a leader note on the card. They are decisions and carry no approval record in this file. Category: none; reversible; decided by the user.

| Id | Pick | What the code does |
| --- | --- | --- |
| U-1 | A: only the leader lands; the leader may be the reviewer of a tier 0 change it wrote none of, never at tier 1 or 2. A lander's merge commit that only joins the tip does not count as authorship (the leader's reading of A, because the leader lands every head). | REQ-20 (`lander_review`) and REQ-21 (a clean join is not authorship); AC-19 and AC-20 test both. |
| U-2 | Observe first; the user records the switch to enforce later; the would-block count and false positives go into the Result. | REQ-14 and REQ-22; the mode is the profile key `work_flow.land_verdict` (D-9), so the switch is a tracked change of the profile. |
| U-3 | A tier 0 landing needs a recorded PASS too; no receipt at tier 0. | REQ-12 asks a receipt at tier 1 and 2 only; REQ-7 holds at every tier. |

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-9 | The mode lives in the profile as `work_flow.land_verdict` with the words `observe` and `enforce`, and no key means observe. | A tracked profile change is a recorded decision with a diff; the integration branch is already read from the same key group. A flag on the call would not be recorded. | Another place for the key; the reader is one function. |
| D-10 | A merge commit is a clean join when its tree equals the tree `git merge-tree --write-tree` makes of its two parents; any other merge, an octopus merge and a git that cannot compute the tree count as authored. | The user's reading of U-1 needs a test a tool can run; an edited merge (a conflict resolution or an added change) is authorship. | Count every merge commit as authored (stricter). |
| D-11 | The reviewer may equal the lander only at the effective tier 0 (the higher of the stated tier and the floor of the changed paths), and a landing with no session cannot be told from the reviewer, so it counts as the reviewer at tier 1 and 2. | REQ-20 and fail closed. | Allow a landing with no session. |
| D-12 | The authors of a head that is already inside the tip with no start row of this card are read over the range from the newest PASS's base to the head. | There is no tip to start from; the recorded base is the only range left. A wrong base can only add authors or leave none, and none refuses (`authors_unknown`). | Refuse such a head. |

## Result

Status: SPEC draft (revision 3). The user's picks U-1 to U-3 are recorded above as decisions; the SPEC carries no approval record and no acceptance line is marked met: an independent audit of the code slice has not run. Observed on the code slice (own run, `cargo test --locked --offline --test card_land_cli --test card_verdict_cli`, exit 0): 49 tests of `card_land_cli` (the 17 new ones `ac40_`..`ac56_` first failed 17 of 17 before the code) and the 3 of `card_verdict_cli` pass; the profile test `land_verdict_key_is_validated` passes. The base is tip 64322298ff. Observe-mode numbers (would-block count and false positives) are empty until the first landings run in observe.
