---
id: SPEC-20261123-lane-completion-handshake
title: "A lane that hands a card back sends the leader one cross-session notice, the leader reads the ledger first and then clears and dispatches"
status: draft
revision: 2
created: 2026-11-23
---

# A lane that hands a card back sends the leader one cross-session notice, the leader reads the ledger first and then clears and dispatches

## Intent

User instruction 2026-11-23 (leader session, direct): when a lane reaches `LANE-DONE` and a leader window is registered, the lane tells that window with one cross-session message, and the leader then sends `/clear` and the next card's dispatch. Today the shipped text says only that a hand-back is not done (`flows/next-flow`) and that the leader clears after the ledger shows done or released; it has no sentence for the message, its order, its address or its failure path. A lane would either say nothing and idle until the leader's next watch round, or message by a name that another project's session may carry.

This SPEC writes the handshake into the shipped skill text, in `flows/next-flow` ("A lane" and "Transition") and in `workflows/work.md`. The ledger stays the record: the message is only a notice. The text adds no command and no behaviour of `moai`; what is not implemented is written as not implemented.

Why the managed `CLAUDE.md` block, the output style and the agent definitions are not in this slice: the block is "facts only, within 15 lines" (`BLOCK_MAX_LINES`, `src/agents.rs`), its flow rules live in the flow rule file of the config folder; the output style shapes what the user reads; `agents/*.md` define what a subagent returns. The flow rule file source and the output style source sit under `skills/moai/references/home/**`, which another card holds (home routing). The same sentences go there after that card lands, as a follow-up.

### Read before writing (2026-11-23, this worktree at `4266b0586c`)

- `skills/moai/references/flows/next-flow.md` (whole): "A lane" says a `LANE-READY-INTEGRATION`, READY, REPAIRED or MERGE-READY note and a `LANE-DONE` line hand a card back and none is done in the ledger; "Transition" says the leader sends `/clear` only after the ledger shows done or released with nothing open, and reads an ACK once.
- `skills/moai/workflows/work.md` (whole): step 7 says a READY note or `LANE-DONE` is not the start signal.
- `src/lane_registry.rs` `show` and `.moai/specs/SPEC-20261123-leader-registration-lease/spec.md` REQ-1, REQ-5: `session-register --mode leader` records the host session id and `session-register --show --name leader` prints it (`free` when no leader is registered or the lease ended).
- `src/host_wake.rs` and `moai host-wake --help`: the wake command addresses a session by name; the registered id is not an argument (SPEC-20261122-session-messages).
- `src/agents.rs` `render_agents`: the managed block holds facts only.
- This repository's `CLAUDE.md` "Writing language": instructions to a lane are in the user's conversation language, only subagent prompts are English; the leader text added here says the same.

## Non-goals

- A `moai` command that sends the notice, or a change of `moai host-wake` to take a registered id. The text names the gap.
- The managed `CLAUDE.md` block, `src/agents.rs`, `src/home_style.rs`, `references/home/**` and `agents/*.md`: see the Intent; the same text follows into the flow rule file once its owner lands.
- A leader workflow file and the `/moai leader` entry: another card. The leader's side is written in `flows/next-flow` only.
- Changing when a lane may take its next card without a registered leader: the existing step 7 of `work.md` stays.

## Requirements

- REQ-1: When a lane hands a card back, the `flows/next-flow` text shall tell the lane to write its READY or REPAIRED note (the `LANE-READY-INTEGRATION` note included) and give the card back before it sends any message · failure: none — the order is a text rule. Revision 2 replaces the term HANDOFF, which no other text defines.
- REQ-2: When a lane has handed its card back and a leader is registered with a live lease, the `flows/next-flow` text shall tell the lane to send one cross-session message that names the card and the head, to the leader's registered session.
- REQ-3 (REQ-2): IF no leader is registered, the lease has ended, the host offers no cross-session channel or the send is refused or held, THEN the `flows/next-flow` text shall tell the lane to write the ledger note only, and to send nothing more for that card.
- REQ-4: The `flows/next-flow` text shall tell the lane to read the leader's session id with `moai session-register --show --name leader` and never to address a row picked by the name `leader`.
- REQ-5: When the leader gets the message, the `flows/next-flow` text shall tell the leader to read the ledger first and to send `/clear` and then the next card's dispatch only when the ledger shows the card done or released and nothing is open.
- REQ-6 (REQ-5): IF the ledger does not show the card done or released, THEN the `flows/next-flow` text shall tell the leader to send no `/clear` and no dispatch and to wait for the ledger.
- REQ-7: The `flows/next-flow` text shall tell a lane never to send `/clear` to itself and never to take its next card on its own while a leader is registered.
- REQ-8: The `flows/next-flow` text shall say that no `moai` command sends the message yet, pointing to SPEC-20261122-session-messages for the wake that exists.
- REQ-9: The `flows/next-flow` text shall tell the lane to write its message, and the leader to write a dispatch, an answer and the line after `/clear`, in the user's conversation language, and only a subagent prompt in English.
- REQ-10: The `workflows/work.md` text shall carry the lane's sentences of REQ-1 to REQ-4 and REQ-7 in short form and point to `flows/next-flow`.
- REQ-11: The flow reference's byte budget in `tests/next_flow_skill_scan.rs` shall be 8192 bytes, its own constant, while the audit contract keeps 6144 bytes · failure: none — the scan test reads both. Revision 2 raises 7168 to 8192 (decision D-2).
- REQ-12: The source manifest of the skill package shall carry the new digests of the two changed files · failure: none — the existing package test reads it.
- REQ-13: IF a lane's card is blocked on a decision, THEN the `flows/next-flow` and `workflows/work.md` text shall tell the lane to take the next card only when no leader is registered, and otherwise to leave the blocked card as its note and wait for the leader's dispatch; the Final flow sentence carries the same condition.
- REQ-14: The `flows/next-flow` text shall say one message per finished card with no retry loop, and that `LANE-DONE` is the lane's last line, after the message when one is sent.
- REQ-15: The `flows/next-flow` text shall tell the lane to match the registered session id to a row of the host's session list and to send nothing when no row matches (R-2).

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-10): the lane's two sentences (the order, and the one message to the registered leader with the ledger-only fallback) are in `flows/next-flow.md` and `workflows/work.md` · verify: `cargo test --offline --locked --test lane_completion_handshake ac1_`
- AC-2 (REQ-4, REQ-7): the address rule (registered id read with `session-register --show`, never a row picked by name) and the lane's two refusals (no `/clear` to itself, no next card on its own while a leader is registered) are in both files · verify: `cargo test --offline --locked --test lane_completion_handshake ac2_`
- AC-3 (REQ-5, REQ-6): the leader's sentence reads the ledger first and the notice without a ledger change starts nothing · verify: `cargo test --offline --locked --test lane_completion_handshake ac3_`
- AC-4 (REQ-8, REQ-9): the text names the missing sending command with the SPEC it points to, and the language rule for the leader's lines · verify: `cargo test --offline --locked --test lane_completion_handshake ac4_`
- AC-5 (REQ-11, REQ-12): the package's manifest and word scans pass with the changed files · verify: `cargo test --offline --locked --test next_flow_skill_scan` and `cargo test --offline --locked --test product_cli skill_package::`

- AC-6 (REQ-13, REQ-14, REQ-15): a blocked card moves the lane on only with no registered leader in both files and in the Final flow sentence; the retry and last-line sentences and the id-to-row match are pinned · verify: `cargo nextest run --test lane_completion_handshake` (ac1_, ac2_, ac6_)

## Decisions

- D-1 (revision 2, 2026-11-23, audit #373 FAIL, leader decision, reversible, category none): the blocked-card sentences gain the "no registered leader" condition (B2); HANDOFF becomes "READY or REPAIRED note" (N3); the retry and last-line sentences are pinned (N1, N4); N2, N5 and N7 are done in the same slice. N6 stays: this SPEC remains in English.
- D-2 (revision 2, 2026-11-23, leader decision "shorten first, raise the budget only if that fails", reversible, category none): after the one merge of the integration tip, `next-flow.md` was 7699 bytes against 7168, because the guard-mutant line of another card (about 600 bytes) joined the file. Shortening this slice's own sentences and one duplicate in Transition left 7686 bytes before the revision-2 sentences and 7712 after them (`wc -c`); the rest could only come from rules and pins of the flow's base text and of the other card, which this slice does not own. The budget is now 8192 bytes (8 KiB). The audit contract keeps 6144.

## Scope

- write: `.moai/specs/SPEC-20261123-lane-completion-handshake/**`, `skills/moai/references/flows/next-flow.md`, `skills/moai/workflows/work.md`, `tests/lane_completion_handshake.rs`, `tests/next_flow_skill_scan.rs` (only the flow reference's byte budget)
- write (one line, leased): `skills/moai/references/source-manifest.json` through `moai card line-lease`
- read-only: `src/lane_registry.rs`, `src/host_wake.rs`, `src/agents.rs`
- forbidden: the home folder of the skill package (another card holds it), the real HOME, the network, every Rust source file

## Plan

1. RED: `tests/lane_completion_handshake.rs` pins the sentences (AC-1 to AC-4) and fails on the current text.
2. Write the lane's and the leader's text in `flows/next-flow.md` and the short form in `workflows/work.md`; GREEN.
3. Recompute the manifest entries for the two files under a line lease; run the scan and package tests.
4. Targeted checks, one `check affected`, READY with head, base and the receipt; the follow-up for the flow rule file is named in the card note.

## Risks

- R-1: The scan tests forbid words (`resident leader`, `leader mode`, `mode-agent`, `mode-leader`); the text uses `leader` and `registered leader` only.
- R-2: The cross-session tool addresses a session by listing name; "the exact id" is read from the registry and matched against the listing. When the listing has no row that matches that id, REQ-3 applies and nothing is sent.
- R-3: The managed block, the output style and the flow rule file stay without the text until the home routing card lands; a user's global files lag the skill until then.

## Result

Status: SPEC draft (revision 2, audit #373 repair); the user approves it. One slice: the skill text and its tests; the flow rule file and the output style follow after the home routing card lands (Non-goals).

- `skills/moai/references/flows/next-flow.md`: a "Completion message" section (the order, the one message, the address, the ledger-only fallback, the notice rule, the lane's two refusals, the missing command) and the leader's three sentences at the top of "Transition". `skills/moai/workflows/work.md`: the short form before the lane-commands paragraph.
- `tests/next_flow_skill_scan.rs`: `FLOW_REF_LIMIT` 7168 for the flow reference (it was 6144 and the file is 7100 bytes now); the audit contract keeps 6144.
- `skills/moai/references/source-manifest.json`: two digest lines, written under a line lease by `tools/source_manifest.py write`; `check` exits 0.
- AC-1: met — `cargo nextest run --test lane_completion_handshake` `ac1_` (1 passed).
- AC-2: met — `lane_completion_handshake` `ac2_` (1 passed).
- AC-3: met — `lane_completion_handshake` `ac3_` (1 passed); the leader's sentences stand between the Transition and Final flow headings.
- AC-4: met — `lane_completion_handshake` `ac4_` (1 passed).
- AC-5: met — `cargo nextest run --offline --locked -E 'binary(lane_completion_handshake) | binary(next_flow_skill_scan) | binary(core_rules_2876) | (binary(product_cli) & test(/skill_package::/))'` at the repair head: 80 passed, 0 failed (the budget is now 8192, D-2).
- AC-6: met — `lane_completion_handshake` `ac6_` and the extended `ac1_`/`ac2_` pass in the same run; each was RED before the text change (`target/handshake/red2.log`: ac1, ac2, ac4 and ac6 failed).
- Revision 2 (audit #373 FAIL): B1 the merged file is 7712 bytes and the budget 8192 (D-2); B2 the blocked-card sentences in `next-flow.md` (A lane, Final flow) and `work.md` name the no-leader condition; N1 N3 N4 pinned, N2 N5 N7 done (see D-1).
- Not done here: the flow rule file and output style sources under the home folder of the skill package (held by another card); a `moai` command that sends the message by registered id (REQ-8 names the gap); the follow-up of the Plan step 4 is the card note.
