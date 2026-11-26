---
id: SPEC-20261121-lane-goal-leader-codex-removed
title: "The lane loop, lane duty, goal block and leader round serve Claude Code only; the Claude /loop line stays byte for byte"
status: draft
revision: 2
created: 2026-11-21
---

# The lane loop, lane duty, goal block and leader round serve Claude Code only

## Intent

User decision of 2026-11-21: moai-adk stops supporting Codex (and the OMP host) and keeps Claude Code, with GLM as a model choice that runs through the Claude Code launcher.

This SPEC removes the Codex and OMP sides of the lane loop (`src/lane_loop.rs`, `lane-loop-line`, the `lane-stop-hook` command), the lane duty line, the `/goal` paste block and the leader round. The lane loop module is shared (`repo_ok`, `lane_ok`, `parse`, `render` and `read_bounded` serve the resume note, the stale-audit and expired-claim readers and the lane reset), so it stays and only its Codex and OMP parts go.

What must not change: the Claude wake line `/loop 5m 레인 <lane> — <duty>` that `lane-loop-line --host claude` prints and the notes around it. A test pins the whole answer.

The goal block stays without its Codex part: Claude Code has a `/goal` command (D-1).

Revision 2 rewrites revision 1 for the tree of 2026-11-26. Revision 1 and its first stage stayed on a branch that never landed; the Codex hook adapter it waited for is gone from the tree, so the whole removal is one change.

## Non-goals

- The launcher, the model questions, the home update, the project layer and the shipped skill text.
- The session id source `CODEX_THREAD_ID` (`src/session_current.rs` and the card, trailer and registration code that reads it): it decides who holds a card and belongs to the session identity code.
- The `lane-stop-hook` entries of the state folder lists in `src/go_state.rs` and `src/state_root.rs`: users' old state folders are still moved (D-6).
- Deleting or rewriting any card or SPEC that describes Codex work.
- Reading, changing or deleting a user's `.codex/` files or old block count files.

## Requirements

- REQ-1: The lane-loop-line command shall print for the host claude exactly the answer it printed before this change: the fields host, lane, writes, lines, verified and note in that order with the same values.
- REQ-2: The lane-loop-line and lane-duty-line commands shall refuse any other host word, codex and omp included, as a usage error that points at --help, with claude as the only host value their help lists.
- REQ-3: The lane loop module shall keep the helpers its other users call with their limits and carry no host other than Claude Code, no hook-file shape, no Stop decision and no goal text.
- REQ-4: The lane-duty-line command shall print the Claude line as it did before this change.
- REQ-5: The lane-stop-hook command shall answer with one line on standard error saying the hook was removed, nothing on standard output and exit zero, reading no standard input and writing nothing.
- REQ-6: The goal-block command shall keep the leader condition, the agent loop line, the goal lines and the counts and carry no codex key in its answer.
- REQ-7: The leader-round command shall refuse an observation whose host is not claude and recognise only the Claude bullet and prompt marks.
- REQ-8: The documentation of these commands shall offer no Codex or OMP use in its help lines, help goldens or operation documents.

## Acceptance

- AC-1 (REQ-1): the whole Claude answer of `lane-loop-line` is pinned · verify: `cargo test --locked --offline --test lane_loop_cli the_claude_answer_is_pinned_whole`
- AC-2 (REQ-2): a removed host word is a usage error that points at `--help`, and the help lists claude alone · verify: `cargo test --locked --offline --test lane_loop_cli a_removed_host_is_a_usage_error` and `cargo test --locked --offline --test lane_duty_line an_unknown_host_is_a_usage_error`
- AC-3 (REQ-3): the lane loop unit tests and the resume note commands pass · verify: `cargo test --locked --offline --lib lane_loop` and `cargo test --locked --offline --test lane_loop_cli resume_note_render_then_check`
- AC-4 (REQ-4): the Claude duty line is unchanged · verify: `cargo test --locked --offline --test lane_duty_line`
- AC-5 (REQ-5): old hook command lines get one line, exit 0 and no write, with standard input held open · verify: `cargo test --locked --offline --test lane_stop_hook_cli`
- AC-6 (REQ-6): the goal block answer keeps its Claude fields and has no codex key · verify: `cargo test --locked --offline --test card goal_block_cli`
- AC-7 (REQ-7): an observation with a removed host is refused and the removed marks are no marks · verify: `cargo test --locked --offline --test leader_round_cli a_removed_host_observation_is_refused` and `cargo test --locked --offline --lib leader_round`
- AC-8 (REQ-8): the help goldens and the help rules pass · verify: `cargo test --locked --offline --test help_layout` and `cargo test --locked --offline --bin moai-cli help_rules`

## Scope

- write: `.moai/specs/SPEC-20261121-lane-goal-leader-codex-removed/**`
- write: `src/lane_loop.rs`, `src/ops/lane_loop_line.rs`, `src/lane_duty.rs`, `src/ops/lane_duty_line.rs`, `src/goal_block.rs`, `src/ops/goal_block.rs`, `src/leader_round.rs`, `src/leader_pulse.rs`, `src/ops/leader_pulse.rs`
- write: `src/main.rs` at the lines of the `lane-stop-hook` command
- write, by line lease: `src/help.rs` and `tests/fixtures/help/*.txt` at the lines of these commands
- write: `tests/lane_loop_cli.rs`, `tests/lane_stop_hook_cli.rs`, `tests/lane_duty_line.rs`, `tests/card/goal_block_cli.rs`, `tests/leader_round_cli.rs`, `tests/home_writer_links.rs`, `tests/fixtures/lane_loop/**`, `tests/fixtures/leader_round/screens/**`
- write: `docs/design-review/appendix/operations/lane_loop_line.md`, `lane_duty_line.md`, `goal_block.md`, `leader_round.md`, `leader_pulse.md`, `docs/hooks-generation.md`

## Plan

1. The pin test of the whole Claude answer, holding the answer of a build made before this change.
2. The removal in one change: the lane loop module, the two line commands, the goal block, the leader round, the leader pulse wording and the lane-stop-hook answer, with their tests, fixtures, help lines and documents.
3. The named test targets, rustfmt, `spec check`, the affected card set once, then the independent audit (tier 1).

## Risks

- A running lane breaks if the Claude line changes. Mitigation: AC-1.
- A leftover Codex Stop hook in a user's `.codex/hooks.json` still calls `moai lane-stop-hook`. Mitigation: AC-5; the user's file is not touched.
- A caller that passed `--project` to `lane-loop-line` now gets a usage error. The flag only chose the folder of the Codex hook file; no shipped skill passes it.
- Four tests of `tests/home_writer_links.rs` about a link on the way to the Stop hook count are deleted with the writer they ran. The link refusal of the other writers stays tested there.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the goal block stays and loses only its Codex key and caveat | Claude Code documents `/goal`: "The `/goal` command sets a completion condition and Claude keeps working toward it without you prompting each step", condition up to 4,000 characters; https://code.claude.com/docs/en/goal and https://code.claude.com/docs/en/commands, read 2026-11-26. The block keeps `model_can_input: false` | remove the goal block later |
| D-2 | the clap value lists of `--host` hold claude only, so another word is a usage error | the hook adapter that parsed help examples with `--host codex` is gone; one kind of refusal in both commands closes the stage-1 audit line about the error kind | refuse at run time |
| D-3 | the leader round refuses an observation of another host and drops the `•` bullet and the `›` prompt mark | nobody has such a file | skip the lane with a reason |
| D-4 | the lane-stop-hook command stays as a one-line answer with exit zero; its flags are optional and ignored | a stale hook must neither block nor fail, like the removed host answer of hook-dispatch | delete the command |
| D-5 | `tests/lane_stop_hook_cli.rs` stays as the test of D-4 | the target name needs no registry change | delete with the command |
| D-6 | the state folder lists keep `lane-stop-hook` | they move users' old state folders | remove with the Go migration |
| D-7 | `--repo` of `lane-loop-line` stays and is still checked; `--project` goes | a call that works today keeps working where the flag still means something | remove `--repo` later |

## Result

Run on 2026-11-26 on macOS, base `c62f942eea`, code commit `ffdda7490a`. The verify lines were run as written, one after another, on the tree that became that commit; all exit 0.

- AC-1: met. 1 test passed.
- AC-2: met. Both lines passed, 1 test each.
- AC-3: met. 7 unit tests and 1 command test passed.
- AC-4: met. 8 tests passed.
- AC-5: met. 1 test passed.
- AC-6: met. 10 tests passed.
- AC-7: met. 1 command test and 33 unit tests passed.
- AC-8: met. 11 and 16 tests passed.

Further:

- On `ffdda7490a`, `moai check affected --card` over the range from the base answered PASS: 73 targets; the table step 88 of 88 tests, the test step 3019 of 3019; format, clippy and the diff check exit 0.
- Before the verify lines, two `cargo nextest run` runs of the wider named set (the lane loop, lane duty, stop hook, leader round, home writer link and help layout targets, the goal block module and the matching unit tests): 168 run, 166 passed, 2 failed each time. Both failures were the two new usage-error assertions: they expected the message to name `claude`, then the refused word, while the real message is `invalid argument value; see --help`. The assertions, REQ-2, AC-2 and the operation document now say what the command does; the two tests pass in the AC-2 lines. The other 166 tests were not changed after those runs.
- The pinned Claude answer of AC-1 equals what the installed `moai 4.2648.54-dev`, built before this change, printed for the same arguments.
- Left out on purpose: `CODEX_THREAD_ID` (Non-goals), the `lane-stop-hook` entries of the state folder lists (D-6), and one comment in `src/go_state.rs` that still says the lane loop writes a count file, because another open card works on that file. Four link tests of the removed Stop hook writer are gone with the writer.
- The affected card set and the independent audit are recorded on the card.
