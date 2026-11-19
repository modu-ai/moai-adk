---
id: SPEC-20261119-flaky-policy
title: "A flaky check never passes by re-running: the merge train records it on a card"
status: draft
revision: 1
created: 2026-11-19
---

# A flaky check never passes by re-running: the merge train records it on a card

## Intent

Card C-99 (Verify D, from Addy Osmani's "How the Claude Code team thinks about verification": the agent cannot skip a flaky test, add sleeps or retry it). Today `mode-leader.md` step 9 says that a failure a re-run on the same SHA turned green "is flaky, not a stop" and stops there: the merge goes ahead and nothing is left behind. `verify.md` already classifies such a test and `flake-record` writes an Issue for it, but the merge train never says to record it, and Issues are being replaced by cards.

This slice (the part of C-99 whose files are free; see Non-goals) makes the merge-train step say what the leader does with a flaky result: the merge is not stopped, and one flaky card keeps the test, both run ids and the SHA, so the next lane repairs the cause or deletes the test with a written reason. It adds a test that pins the rule and its byte budget.

## Non-goals

- The `mode-agent.md` text that forbids sleeps, retries and skips as a repair and requires a recorded deletion (C-99 AC 2), and the `source-manifest.json` hash of the edited workflow files: both paths are reserved by card C-77, so they follow in a second SPEC when C-77 releases them (LANE-BLOCKED note on C-99).
- A `moai card list` filter for flaky cards (C-99 AC 3, first half): it needs `src/ops/card.rs` and `src/card/*`, held by C-43 and outside this card's scope. Until then a flaky card is found by its title prefix `flaky:`.
- Changing `flake-record` from Issues to cards (code, outside this card's scope).

## Requirements

- REQ-1: When the merge train meets a check that a re-run on the same SHA turned green, the leader shall let the merge through and record the test, the failed run id, the passing run id and the SHA on one flaky card.
- REQ-4 (REQ-1): IF the flaky card cannot be written, THEN the leader shall put the same four facts in the merge comment and say that the card is missing, and the merge still goes through.
- REQ-2: The `mode-leader.md` sentence that calls such a result not a stop shall point to the flaky card rule.
- REQ-3: The flaky card rule in `mode-leader.md` shall take at most 700 bytes while the file stays within the package file bound of 65536 bytes.

## Acceptance

- AC-1 (REQ-1, REQ-4): the merge-train step names the flaky card (title `flaky: <test>`, one card per test, a note on the open one), what it records (test, failed run id, passing run id, SHA), what is done when the card cannot be written (the four facts go in the merge comment), and says the merge is not stopped and that a re-run's pass is never the verification of a change that touches the test.
  verify: `cargo test --locked --offline --test workflow_size_budget flaky_rule_names_the_card_and_what_it_records`
- AC-2 (REQ-2): the "not a stop" sentence contains the pointer to the rule.
  verify: `cargo test --locked --offline --test workflow_size_budget the_not_a_stop_sentence_points_to_the_flaky_rule`
- AC-3 (REQ-3): the rule's text is at most 700 bytes and `mode-leader.md` is within the package file bound.
  verify: `cargo test --locked --offline --test workflow_size_budget flaky_rule_stays_within_its_byte_budget`
- AC-4 (REQ-1…REQ-4): neighbours.
  verify: `cargo test --locked --offline --test mode_cards_53`; `cargo test --locked --offline --test mode_workflow_steps`; `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test ac_command_lint`; `cargo fmt --check`

## Scope

- write: `skills/moai/workflows/mode-leader.md`, `tests/workflow_size_budget.rs`, `.moai/specs/SPEC-20261119-flaky-policy/spec.md`
- read-only: `skills/moai/workflows/verify.md`, `skills/moai/workflows/mode-agent.md`, `docs/design-review/appendix/operations/flake_record.md`, `src/flake.rs`
- forbidden: `skills/moai/workflows/mode-agent.md` and `skills/moai/references/source-manifest.json` (card C-77), `src/**` (card C-43 holds `src/ops/card.rs`)

## Plan

1. Tests first: the three checks in `tests/workflow_size_budget.rs` read `mode-leader.md`; they fail before the text is written.
2. Edit `mode-leader.md` step 9: point from the "not a stop" sentence to the rule and add the rule right after it.
3. Run the neighbour tests; record which test would fail because the manifest is stale (the manifest follows in the second SPEC).

## Risks

- `skills/moai/**` is hashed in `source-manifest.json`; edited without it, the package tests that compare the manifest fail until the second SPEC lands. This PR waits for C-77 to release the manifest, or the leader moves that edit.
- The existing Issue-based wording of `verify.md` (`[flake]` Issues, `flake-record`) still says Issues; the rule here says "card" for the merge train only. Aligning `verify.md` and `flake-record` is a follow-up (the Issue to card change is the leader's decision).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flaky record is a card titled `flaky: <test>` (kind `task`), one per test, with a `report` note per new failure | `card add --kind` exists, no code needed; "all cards" decision of 2026-11-19 | Back to a `[flake]` Issue |
| D-2 | This slice ships only `mode-leader.md` and its test; the `mode-agent.md` text and the manifest follow when C-77 releases them | `card take` refused those paths (reserved by C-77) | Take the paths now if the leader moves C-77's claim |
| D-3 | The byte budget of the rule is 700 bytes, the file bound stays 65536 | the file is about 44.6 KB; a rule over 700 bytes is a page of text in a step line | Raise the number in a later SPEC |

## Result

(RED and the checks come first; one line per AC is written after them.)
