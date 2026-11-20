---
id: SPEC-20261120-audit-request-notes
title: "audit requests, claims, verdicts and repairs are card notes of fixed shapes that a read command lists with ages and stage times"
status: draft
revision: 1
created: 2026-11-20
---

# audit requests, claims, verdicts and repairs are card notes of fixed shapes that a read command lists with ages and stage times

## Intent

A lane could not reach the auditor a leader had started for its card, and long waits for an audit were invisible. The audit request becomes a record any session reads: four card notes of fixed first lines (request, claim, verdict, repair), and `moai audit-queue`, a read-only command that lists the open requests with their age and state and computes the time of each stage from the notes' own times. A re-audit request then reaches the first auditor's owner without a live channel, and the stage times of a finished audit are known for measuring.

## Requirements

- REQ-1: When a card note's first line is `AUDIT-REQUEST <card> head=<40-hex> tier=<0|1|2> scope=<text>`, the audit-queue command shall read it as a request to audit that head.
- REQ-2: When a card note's first line is `AUDIT-CLAIM <card> head=<40-hex> by=<name>`, the audit-queue command shall read it as the claim of that head by that auditor.
- REQ-3: When a card note's first line is `AUDIT-VERDICT <card> head=<40-hex> verdict=<PASS|FAIL>`, the audit-queue command shall read it as the verdict on that head.
- REQ-4: When a card note's first line is `REPAIR <card> head=<40-hex> of=<40-hex>`, the audit-queue command shall read it as a repair head made for the head named by `of`.
- REQ-5: The audit-queue command shall list every request whose head has no verdict with its card, head, tier, scope, age in seconds and one state of waiting, running, repair or stale.
- REQ-6 (REQ-5): The audit-queue command shall give a request with no claim the state waiting, a claim with no verdict the state running, a FAIL verdict with no later repair the state repair, and a waiting or running request older than `--stale-secs` the state stale.
- REQ-7 (REQ-5): The audit-queue command shall name as the auditor of a card the `by` of that card's first claim, so that the request for a repair head goes to the same auditor.
- REQ-8 (REQ-5): The audit-queue command shall compute for each head, from the times of its notes, the seconds from request to claim, claim to verdict, verdict to repair and repair to the next claim, leaving a span out when one end has no note.
- REQ-9: IF a head already has a PASS verdict, THEN the audit-queue command shall list a later request for the same head as reuse and not as waiting.
- REQ-10 (REQ-1, REQ-2, REQ-3, REQ-4): IF a note's first line starts with one of the four words but does not have the shape of its words, THEN the audit-queue command shall list it as malformed with its card and note id, without counting it.
- REQ-11: The audit-queue command shall not write to the ledger.
- REQ-12: The help of the audit-queue command and its operations document shall state the four shapes, the four states and the reuse rule.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): a request posted as a note is listed with its card, head, tier, scope and age, and a head that has a verdict is no longer listed · verify: `cargo test --locked --offline --test cli_audit_queue lists`
- AC-2 (REQ-6, REQ-7): the four states are told apart at the edges (no claim, claim, FAIL, FAIL then repair, one second under and over `--stale-secs`) and the auditor of a card is its first claimer · verify: `cargo test --locked --offline --test cli_audit_queue state`
- AC-3 (REQ-8): the stage times of a finished audit and of an audit with a repair are computed from the note times, and a missing end leaves the span out · verify: `cargo test --locked --offline --test cli_audit_queue times`
- AC-4 (REQ-9, REQ-10, REQ-11): a second request for a passed head is reuse, a broken line is listed as malformed and not counted, and the ledger bytes are the same after the command · verify: `cargo test --locked --offline --test cli_audit_queue rules`
- AC-5 (REQ-12): the help and the operations document state the shapes, states and the reuse rule · verify: `cargo test --locked --offline --test cli_audit_queue help`

## Scope

- write: `.moai/specs/SPEC-20261120-audit-request-notes/spec.md`
- write: `tests/cli_audit_queue.rs`
- write: `src/ops/audit_queue.rs`
- write: `docs/design-review/appendix/operations/audit_queue.md`

## Decisions

- D-1: The command is a one-file operation under `src/ops/` with its logic and unit tests in the same file, so `src/lib.rs` is not touched (it is held for a long time by other cards). Reversible: the logic moves into the library crate later.
- D-2: Times come from the ledger's own note times (`card_note.created_at`), never from the text of a note, so a lane cannot move a stage time by writing it.
- D-3: A state word is a fact of the notes, not a judgement: a claim with no verdict is running even when its auditor is gone, and the `stale` state is what shows that.
- D-4: The rules about who audits are not enforced here: the command reads and lists; the workflow text says that an author never audits, that the exact head is audited, and that a repair goes to the same auditor.

## Plan

1. Tests first (RED): `tests/cli_audit_queue.rs` with the list, state, time, rule and help cases above, notes inserted with SQL where a case needs a chosen time; see them fail on the unchanged code.
2. `src/ops/audit_queue.rs`: the four shapes, the state and time judgement as functions of plain rows (unit tests in the file), the read-only query, the answer.
3. The operations document fragment; the help rows (`GROUPS` and `EXAMPLES` of `src/help.rs`) when that file is free.
4. Run the new file, the help and operations-document tests, `--bins`, clippy, rustfmt, then audit and write the Result.

## Result

- AC-1: met — `cargo test --locked --offline --test cli_audit_queue lists` passed (2 tests): a request written as a note is listed with its card, head, tier, scope and age, and a head with a verdict is not listed; an absent ledger answers `ledger: "absent"` and creates nothing.
- AC-2: met — `--test cli_audit_queue state` passed (2 tests): waiting, running, repair and the repair head's own request are told apart, the auditor of the repair head is the first claimer of the card, and `stale` starts with a limit of 900 over an age of about 1000 seconds and not with 1100 (a limit of 0 is refused by the argument parser); four unit tests in `src/ops/audit_queue.rs` read the edge at one second (`stale` strictly after the limit, a claim before its request is ignored, a span that ends before it starts is null).
- AC-3: met — `--test cli_audit_queue times` passed: request to claim 60 s, claim to verdict 240 s, verdict to repair 100 s and repair to the next claim 100 s from the inserted note times, and the repair head without a verdict has a null audit span.
- AC-4: met — `--test cli_audit_queue rules` passed (3 tests): a second request for a passed head is `reuse`, three broken first lines are listed as `malformed` and not counted, and the home snapshot is the same before and after the command.
- AC-5: met — `--test cli_audit_queue help` passed: the long help names the four shapes, the states, `--stale-secs` and the first-claim rule, and the operations fragment `audit_queue.md` states the shapes, `stale` and `reuse`.
- Also run: all nine tests of `cli_audit_queue` in one run: 9 passed, 0 failed; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0 without a warning; `rustfmt --edition 2021 --check` on the two new Rust files is clean.
- Help table: `src/help.rs` has the `GROUPS` row and the `EXAMPLES` row of `audit-queue`; the four `help --all` goldens (`tests/fixtures/help/all.txt`, `all-dumb.txt`, `all-60.txt`, `all-80.txt`) were regenerated with `MOAI_HELP_GOLDEN_WRITE=1` and differ from before by the one `audit-queue` line each (read with `git diff -U0`).
- Whole area run: `cargo test --locked --offline --no-fail-fast --lib --bins` and the readers of the help table and the operations documents (`cli_audit_queue`, `help_layout`, `product_cli`, `hot_file_ops_merge`, `index_cli`, `cli_project`, `check_affected_cli`, `orca_optional_3005`, `baseline_git_env`): 12 `test result: ok` lines, 0 failed (`--lib` 1638 passed).
- Python area modules, run with `python3 -m unittest` in an empty temporary HOME after the build: `repo_scans_test` (25), `self_check_test` (66), `local_check_test` (54), `ci_scope_test` (47), all OK.
- Not yet: the workflow text that tells lanes to post the four notes and the leader to read `moai audit-queue` is in files held by the card that cleans the workflow text and follows after it; the audit; no test here is specific to macOS and none was run for it.
