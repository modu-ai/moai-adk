---
id: SPEC-20261120-orca-1-4-267-alignment
title: "the Orca rules of the lane texts follow what Orca's pages and CLI say: no resend by terminal send on a null delivered_at, peek for watching, tracked work by Run and Task, notices by mail"
status: draft
revision: 1
created: 2026-11-20
---

# the Orca rules of the lane texts follow what Orca's pages and CLI say: no resend by terminal send on a null delivered_at, peek for watching, tracked work by Run and Task, notices by mail

## Intent

The MODE-AGENT and MODE-LEADER texts carry a rule written when Orca's mail was measured on an older build: when a mail's `delivered_at` is still `null` and the target looks idle, send the same text again with `orca terminal send`. Orca's four official pages (`research.md`) never mention `delivered_at`, never describe mail states, and never tell a sender to resend; the installed CLI (1.4.267) says that a default `check` consumes the mail of the agent that owns the inbox, that a script which only watches must use `--peek` or `--all`, and that a lost answer is replayed by its request id. The old rule can type the same instruction twice into a lane and, through a watching script's plain `check`, take mail the lane never sees.

This change removes the resend rule from the adapter text and the two mode texts, writes the rules the pages and the local help do support (watching with `--peek` and `--all`, tracked work as Run, Task and Dispatch, one-off prompts and notices on the channel the page names, an unclear send followed by its request id), marks what the pages leave out as unverified, and cites the page or the local help and its read time in each rule. The Orca step stays optional, the independent audit, the required checks and the HOLD categories stay, and nothing here changes a setting or deploys.

## Requirements

- REQ-1: The MODE-AGENT and MODE-LEADER texts shall not tell a lane or the leader to send the text of a mail again with `orca terminal send` because the `delivered_at` of the mail is `null`.
- REQ-2: The MODE-AGENT and MODE-LEADER texts shall tell the sender to read a `null` `delivered_at` as not yet known and never to resend the mail as a new mail.
- REQ-3: The MODE-AGENT and MODE-LEADER texts shall tell the sender to write which of stored, delivered and work done a command showed and to write any other mail state as unconfirmed.
- REQ-4: The MODE-AGENT and MODE-LEADER texts shall tell a reader that only watches a mailbox to use `orca orchestration check --peek` or `--all` and to leave the default check to the agent that owns the inbox.
- REQ-5: The MODE-AGENT and MODE-LEADER texts shall tell the sender to carry supervised work as a Run, a Task and a Dispatch and to send a plain notice as mail.
- REQ-6 (REQ-2): IF the state of a send or a mail is unclear, THEN the MODE-AGENT and MODE-LEADER texts shall tell the sender to follow it by a card note and the request id the command printed instead of sending again.
- REQ-7: The Orca adapter text shall name the four pages and the local CLI version it rests on with their read times and list what the pages do not document.
- REQ-8: The Orca adapter text shall say that a missing `hostScope`, native chat and sleep or restore are unverified and that a lane keeps its card and its reservation while a fact is unverified.
- REQ-9: The Orca adapter text shall say that a screen read is the frame that is rendered now and that a cursor of `terminal read` is a position in the output stream.
- REQ-10: Every rule of the Orca texts that this change writes shall name the page or the local help it rests on, without a card number, an Issue number or a decision number.

## Acceptance

- AC-1 (REQ-1, REQ-2): the two mode texts no longer carry the resend rule and say what a `null` `delivered_at` means · verify: `cargo test --locked --offline --test orca_alignment resend_`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): the two mode texts carry the mail states, the peek rule, the Run Task Dispatch and mail split and the request-id rule · verify: `cargo test --locked --offline --test orca_alignment rules_`
- AC-3 (REQ-7, REQ-8, REQ-9): the adapter text carries the pages with their read times, the local version, the undocumented list, the unverified facts with ownership kept and the screen and cursor meaning · verify: `cargo test --locked --offline --test orca_alignment adapter_`
- AC-4 (REQ-10): the changed rules cite a page or the local help and carry no card, Issue or decision number, and the scan and reader tests of the three files pass · verify: `cargo test --locked --offline --test orca_alignment cites_`
- AC-5 (REQ-1…REQ-10): the code and tests that touch lane sends follow the new texts: `src/lane_send.rs` plans, writes and refuses as before and its tests and the tests that pin the old rule pass · verify: `cargo test --locked --offline --test mode_workflow_steps`

## Scope

- write: `.moai/specs/SPEC-20261120-orca-1-4-267-alignment/**`, `tests/orca_alignment.rs`
- write (when the files are free): `skills/moai/references/adapters/orca.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json`, `tests/mode_workflow_steps.rs`
- write (only if a refusal text turns out to contradict the new rules): `src/lane_send.rs`

## Decisions

- D-1: The pages name no mail state, so the three states of REQ-3 are not a document's states: stored is a send whose answer says ok, delivered is a set `delivered_at` of the local send answer, work done is a `worker_done` mail or a Task `completed`. A state a command did not show is written `unconfirmed`. The local help does not mention `delivered_at`; the field is read from the answer as the earlier measurements saw it, and the text says it is not documented.
- D-2: The page says "For one-off prompts, use orca terminal send" and "Use orca terminal send for a lightweight prompt to an agent you are watching", so a terminal send stays the channel for a typed line to a lane that is read first (the read-before-send items stay). What goes away is sending again because a mail's field is `null`, which no page supports.
- D-3: The web page's `check --ack <deliveryId>` example is the older form; the installed CLI says no `--ack` is needed and that a script must never `--ack`. The texts copy the local rule, not the web example.
- D-4: Facts the pages leave out (mail states, `delivered_at`, what happens to mail after a restart, `hostScope` on the orchestration page, sleep and restore) are written `unverified`. An unverified fact never ends the ownership of a card or a reservation and never authorises a recovery; it is read again or settled by the card note.
- D-5: The older drafts on the orchestration redesign overlap this card (the Run, Task and Dispatch split). This SPEC adds no new orchestration behaviour; it only removes the resend rule and states the split the page states. Whatever those drafts still add stays theirs.
- D-6: `src/lane_send.rs` sends a text once after two equal reads and refuses otherwise; it has no mail code and no resend path, so the card's wish to update it with the text is met by a test that it still refuses a changed screen and never loops, not by a code change. If a refusal text names a resend, it is reworded.
- D-7: The base of the M-Flow stays the local integration branch at an exact commit; nothing here names Orca's default branch or a new integration branch.

## Plan

1. This card, now: this SPEC and `research.md` (the four pages, the local help, their read times).
2. Tests first: `tests/orca_alignment.rs` reading the three shipped files (red against the resend rule), then the text of the adapter and of the two mode files when the files are free, then the tests that pinned the old wording (`tests/mode_workflow_steps.rs`), the manifest hashes, the scan tests.
3. Whole-area run, clippy, rustfmt, spec check, audit, merge-ready.

## Result

- AC-1: met — `cargo test --test orca_alignment resend_`: the two mode texts no longer carry the resend rule and say a `null` `delivered_at` is read as not yet known (2 passed).
- AC-2: met — `cargo test --test orca_alignment rules_`: mail states, the peek rule, the Run Task Dispatch and mail split and the request-id rule are in both mode texts (4 passed).
- AC-3: met — `cargo test --test orca_alignment adapter_`: the adapter text names the four pages with their read times, the local version, the undocumented list, the unverified facts with ownership kept, and the screen and cursor meaning (3 passed).
- AC-4: met — `cargo test --test orca_alignment cites_` passed, and the scan and reader tests of the three files passed: the 19 test files that read the mode texts (`mode_workflow_steps`, `orca_optional_3005`, `lane_rules_*`, `workflow_size_budget`, `core_rules_2876`, `cli_text_citations` and the rest) and 8 python files (`ci_scope_test`, `source_manifest_test`, `repo_scans_test` and others), all ok.
- AC-5: met — `cargo test --test mode_workflow_steps`: 43 passed with the four pinned phrases changed to the new rule; `src/lane_send.rs` needed no change (D-6), and `cargo test --lib --bins` passed (1541, 28, 28).
- Mutants: five rule texts broken one at a time (resend rule back, peek rule removed, request id removed, unverified ownership removed, a card number cited); each was caught by its test, and the tree was restored.
- Not run here: a live Orca session; every rule rests on the pages and the local help named in the texts, and the facts the pages leave out are written as unverified.
