---
id: SPEC-20261121-loop-leader-session-only
title: "The leader round arms the loop command only in a session confirmed as the leader"
status: draft
revision: 1
created: 2026-11-21
---

# The leader round arms the loop command only in a session confirmed as the leader

## Intent

`workflows/mode-leader.md` tells a session to run the leader round under `/loop 5m /moai --mode-leader`, and `references/adapters/orca.md` repeats it in the MODE-LEADER row. Nothing tells the session to check first that it is the leader. A session that reads the leader workflow by mistake (a lane that was handed the wrong text, or a session of unknown mode) can arm a timer that starts leader rounds in the wrong place, and a second leader round is the one failure the leader workflow names as worse than none. This SPEC adds one rule to the leader workflow and the Orca adapter: the loop command is armed only in a session that has confirmed it is the leader, meaning a session that was entered with `--mode-leader`. It changes skill text and one test only. How a lane is driven (the loop line that the launcher gives a lane, the lane workflow's wake rules, `src/lane_loop.rs`) is code and a user decision, and is outside this SPEC.

## Non-goals

- The lane's loop line (`/loop 5m /moai --mode-agent <duty>`), `src/lane_loop.rs`, `src/ops/lane_loop_line.rs`, `tests/lane_loop_cli.rs` and the wake rules of `workflows/mode-agent.md`.
- Removing `/loop` from the leader round: the leader still runs under it.
- The stage order and the report rules, which are other slices of the same card.

## Requirements

- REQ-1: The leader workflow shall tell a session to arm `/loop` only after it has confirmed that it is the leader session.
- REQ-2: The leader workflow shall define a confirmed leader session as one that was entered with `--mode-leader`.
- REQ-3: The leader workflow shall tell a session that is not confirmed as the leader not to arm the leader round's `/loop`.
- REQ-4: The Orca adapter shall say in its MODE-LEADER row that the loop command is armed only in a confirmed leader session.
- REQ-5: The lane workflow shall keep its wake rules as they are.
- REQ-6: The skill's source manifest shall match the files after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the leader workflow has the rule with the definition and the refusal for an unconfirmed session.
  verify: `cargo test --locked --offline --test loop_leader_session the_leader_workflow_arms_the_loop_only_in_a_confirmed_leader_session`
- AC-2 (REQ-4): the Orca adapter's MODE-LEADER row says the same.
  verify: `cargo test --locked --offline --test loop_leader_session the_orca_adapter_says_the_same_for_the_leader_row`
- AC-3 (REQ-5): the lane workflow's wake rules are unchanged.
  verify: `cargo test --locked --offline --test mode_workflow_steps`
- AC-4 (REQ-6): the manifest matches the package.
  verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261121-loop-leader-session-only/**`, `tests/loop_leader_session.rs`, `skills/moai/workflows/mode-leader.md` (after its other holder lands), `skills/moai/references/adapters/orca.md`, `skills/moai/references/source-manifest.json` (regenerated last, after its other holders release it)
- read-only: `skills/moai/workflows/mode-agent.md`, `src/lane_loop.rs`, `tests/mode_workflow_steps.rs`
- forbidden: a card, Issue or pull request number in the sentences this change adds, and any edit of a file another card holds before that card has released it or agreed in a pair line

## Plan

1. Write the test first: `tests/loop_leader_session.rs` reads the two files by path and checks the rule, the definition, the refusal and the Orca row; it fails on the tip because none of them says it.
2. Add one sentence to the `mode-channel` step of `workflows/mode-leader.md` and one clause to the MODE-LEADER row of `orca.md`.
3. Regenerate the manifest rows; run the new test, `mode_workflow_steps` and every test that reads these two files; ask for an independent audit.

## Risks

- `workflows/mode-leader.md` is held by another card at the time of writing: the work waits for its landing, then merges the tip before the edit.
- `tests/mode_workflow_steps.rs` pins lines of the leader workflow: the new sentence is appended inside the `mode-channel` step without rewording a pinned sentence, and the target runs without a name filter.
- The wider question (whether a lane should be driven by a loop command at all) is open and is not decided here.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only the leader round is limited; the lane loop line stays | The launcher gives lanes that line and two tests pin it, so removing it is a code change and a user decision | Add a lane SPEC that changes `src/lane_loop.rs` and the lane workflow together |
| D-2 | A confirmed leader session is one entered with `--mode-leader` | That flag is the only spelling that selects the leader route | Add another confirmation (a run of a leader command) later |

## Result

Filled in by the implementation lane on 2026-11-21 (`date -u`); the audit checks it again. The SPEC is `429a55e16e`; red `1e3e8ca66e`, green `379a15e31b`, the manifest rows `fc3507a86b`; the local tip `42f5af0800` is merged in as `e6c041ddb2` (no conflict). Evidence files are in this worktree's `target/notes/`.

Red (`red.out`, before the text): `tests/loop_leader_session.rs` 1 passed, 2 failed (the two tests of the new rule; the lane-line test passes on the tip text, as it should).

- AC-1: met — `cargo test --locked --offline --test loop_leader_session` (`green.out`): 3 passed, 0 failed; `the_leader_workflow_arms_the_loop_only_in_a_confirmed_leader_session` is the test of AC-1. The rule sits at the end of the `mode-channel` step of `skills/moai/workflows/mode-leader.md`.
- AC-2: met — same run, `the_orca_adapter_says_the_same_for_the_leader_row`.
- AC-3: met — same run, `the_lane_loop_line_is_left_as_it_is`.
- AC-4: met — `python3 -I tools/source_manifest.py check` prints `{"conflicted": false, "changed": [], "dropped": [], "ok": true}` on the final head.

### Checks

`moai check affected 42f5af0800..fc3507a86b` (the branch's own build of the binary, `--max 2`), summary line as printed: `check-affected base=42f5af0800fb head=fc3507a86b95 targets=43 py=8 scans=listed:42 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS`.

Targeted run before the manifest commit (`green.out`, 8 targets, no name filter, exit 0): `agent_defs_package` 19, `close_step_2903` 2, `flaky_followups` 3, `lane_rules_2879` 5, `loop_leader_session` 3, `mode_cards_53` 9, `mode_guidance_scan` 8, `mode_workflow_steps` 43 passed.

Mutants (five, on the committed text, each restored, `restored: clean = True`; `mutate.out`): the limit sentence losing its condition, the confirmation naming the other mode, an unconfirmed session arming the loop, the Orca row dropping the limit, the lane workflow losing its fixed loop line — each killed by the named test of the matching AC.

Shared files: `skills/moai/references/source-manifest.json` (two rows) was committed on this branch under the leader's standing pair approval for the card's own manifest keys (decision notes 5617, 5623, 5658); the card check refuses that path.

### Not run

- A leader session of a real project reading the new rule in a host: not run; the text is pinned by words only.
- The code that arms the lane's loop (`src/lane_loop.rs`) and the lane workflow: unchanged on purpose (decision D-1, leader note 5595).
- A second run of the same tests on the same tree: not repeated.
