---
id: SPEC-20261119-flaky-policy
title: "A flaky check never passes by re-running: the merge train records it on a card"
status: draft
revision: 2
created: 2026-11-19
---

# A flaky check never passes by re-running: the merge train records it on a card

## Intent

Card C-99 (Verify D, from Addy Osmani's "How the Claude Code team thinks about verification": the agent cannot skip a flaky test, add sleeps or retry it). Today `mode-leader.md` step 9 says that a failure a re-run on the same SHA turned green "is flaky, not a stop" and stops there: the merge goes ahead and nothing is left behind. `verify.md` already classifies such a test and `flake-record` writes an Issue for it, but the merge train never says to record it, and Issues are being replaced by cards.

This slice (the part of C-99 whose files are free; see Non-goals) makes the merge-train step say what the leader does with a flaky result: the merge is not stopped, and one flaky card keeps the test, both run ids and the SHA, so the next lane repairs the cause or deletes the test with a written reason. It adds a test that pins the rule and its byte budget.

## Non-goals

- The `mode-agent.md` text that forbids sleeps, retries and skips as a repair and requires a recorded deletion (C-99 AC 2): `mode-agent.md` was reserved by card C-77 (now done), so it follows in a second SPEC.
- A `moai card list` filter for flaky cards (C-99 AC 3, first half): it needs `src/ops/card.rs` and `src/card/*`, held by C-43 and outside this card's scope. Until then a flaky card is found by its title prefix `flaky:`.
- Changing `flake-record` from Issues to cards (code, outside this card's scope).

## Requirements

- REQ-1: When the merge train meets a check that a re-run on the same SHA turned green, the leader shall let the merge through and record the test, the failed run id, the passing run id and the SHA on one flaky card.
- REQ-4 (REQ-1): IF the flaky card cannot be written, THEN the leader shall put the same four facts in the merge comment and say that the card is missing, and the merge still goes through.
- REQ-2: The `mode-leader.md` sentence that calls such a result not a stop shall point to the flaky card rule.
- REQ-3: The flaky card rule in `mode-leader.md` shall take at most 700 bytes while the file stays within the package file bound of 65536 bytes. Its placement inside the merge-train step and the 700/701 edge are pinned in SPEC-20261119-flaky-budget-pins (slice 2, audit #3225 N1, N10, N11).

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

- write: `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json` (the hash and origin of `mode-leader.md` only, written by `tools/source_manifest.py write`), `tests/workflow_size_budget.rs`, `.moai/specs/SPEC-20261119-flaky-policy/spec.md`
- read-only: `skills/moai/workflows/verify.md`, `skills/moai/workflows/mode-agent.md`, `docs/design-review/appendix/operations/flake_record.md`, `src/flake.rs`
- forbidden: `skills/moai/workflows/mode-agent.md` (second SPEC), `src/**` (card C-43 holds `src/ops/card.rs`)

## Plan

1. Tests first: the three checks in `tests/workflow_size_budget.rs` read `mode-leader.md`; they fail before the text is written.
2. Edit `mode-leader.md` step 9: point from the "not a stop" sentence to the rule and add the rule right after it.
3. Run the neighbour tests; the manifest hash of `mode-leader.md` is rewritten in this SPEC with the tool (Scope, D-2) — the AC-4 neighbour `skill_package` test failed before that rewrite and passes after it.
4. Slice 2: the placement and edge pins moved to SPEC-20261119-flaky-budget-pins (leader decision (a), card note 2685); the wording fixes for N2 and N3 need `mode-leader.md` and `verify.md`, which other cards hold.

## Risks

- `skills/moai/**` is hashed in `source-manifest.json`: the hash of `mode-leader.md` is rewritten here with the tool; another branch that edits the manifest merges through the manifest merge driver or `source_manifest.py write`.
- The existing Issue-based wording of `verify.md` (`[flake]` Issues, `flake-record`) still says Issues; the rule here says "card" for the merge train only. Aligning `verify.md` and `flake-record` is a follow-up (the Issue to card change is the leader's decision).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flaky record is a card titled `flaky: <test>` (kind `task`), one per test, with a `report` note per new failure | `card add --kind` exists, no code needed; "all cards" decision of 2026-11-19 | Back to a `[flake]` Issue |
| D-2 | This slice ships `mode-leader.md`, its manifest hash and its test; the `mode-agent.md` text follows in a second SPEC | `card take` refused `mode-agent.md` and the manifest while C-77 held them; C-77 is done | Add the `mode-agent.md` text here if the leader prefers one PR |
| D-3 | The byte budget of the rule is 700 bytes, the file bound stays 65536 | the file is about 44.6 KB; a rule over 700 bytes is a page of text in a step line | Raise the number in a later SPEC |

## Result

- AC-1: met — `workflow_size_budget` `flaky_rule_names_the_card_and_what_it_records` (RED first: `mode-leader.md has no Flaky rule…`); mutants G2 (merge stops), G3 (missing-card path gone), G4 (passing run id not recorded), G5 (re-run pass allowed), G6 (one card per test dropped) killed
- AC-2: met — `the_not_a_stop_sentence_points_to_the_flaky_rule` (RED first); mutant G1 (pointer removed) killed
- AC-3: met — `flaky_rule_stays_within_its_byte_budget`: the rule is 501 bytes (budget 700), `mode-leader.md` 45,128 bytes (bound 65,536); mutant G7 (rule grows past the budget) killed
- AC-4: met — `mode_cards_53` 9, `mode_workflow_steps` 43, `agent_defs_package` 19, `product_cli` 79 (including `skill_package::source_manifest_covers_every_package_file`, which failed before the manifest hash was rewritten with `python3 -I tools/source_manifest.py write`; `check` exits 0), `ac_command_lint` 5, `help_layout` 10, `home_isolation_scan` 18 and `cargo fmt --check` pass
