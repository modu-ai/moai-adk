---
id: SPEC-20261121-lane-goal-leader-codex-removed
title: "The lane loop, lane duty, goal block, lane idle and leader round serve Claude Code only; the Claude /loop line stays byte for byte"
status: draft
revision: 1
created: 2026-11-21
---

# The lane loop, lane duty, goal block, lane idle and leader round serve Claude Code only

## Intent

User decision of 2026-11-21: moai-adk stops supporting Codex (and the OMP host) and keeps Claude Code, with GLM as a model choice that runs through the Claude Code launcher. The cards and SPECs that describe Codex work stay as history.

The impact report of that removal left one area unassigned: the Codex and OMP sides of the lane loop (`src/lane_loop.rs`, `lane-loop-line`, the `lane-stop-hook` command), the lane duty line, the `/goal` paste block, the lane idle line and the leader round. This SPEC removes them. The lane loop module is shared (its `repo_ok`, `lane_ok`, `parse`, `render` and `read_bounded` serve the resume note, the stale-audit and expired-claim readers and the lane reset), so it stays and only its Codex and OMP parts go.

What must not change: the Claude wake line `/loop 5m /moai --mode-agent <duty>` that `lane-loop-line --host claude` prints, and the notes around it. Running lanes read that line today, so a test pins the whole Claude answer before any deletion.

The goal block stays, with its Codex part removed. The official Claude Code documentation has a built-in `/goal` command (D-1), so the leader's milestone condition, the agent's loop line and the card goal lines are still useful to Claude Code sessions.

## Non-goals

- The launcher, the model questions and the home update: another slice.
- The shipped skill text (`skills/moai/**`), which still names the `loop_line`, `goal_line` and `goal-lines` keys and the Codex wording: another slice. The `codex` key of the goal block answer is read by none of it.
- The Codex lane hook adapter in `src/hooks.rs` (`LaneLoop`, `codex_hooks`, the `lane` parameter) and the `--lane-loop` flags of `hooks-plan` and `hooks-apply`: the hooks slice. Until it lands the three Stop-hook items `stop_hook_command_with`, `STOP_HOOK_SUBCOMMAND` and `STOP_HOOK_TIMEOUT_SECS` stay in the lane loop module.
- The `lane-stop-hook` entry of the state-folder move list in `src/go_state.rs`: users' old state folders are still moved (D-6).
- Deleting or rewriting any card or SPEC that describes Codex work, including the SPEC of the leader round card-loop signals that names two Codex and OMP screens.
- The retired-host layer for stored files (`src/retired_hosts.rs`): command-line host words are refused, not skipped (D-2).

## Requirements

- REQ-1: The lane-loop-line command shall print for the host claude exactly the answer it prints today: the fields host, lane, writes, lines, verified and note in that order with the same values.
- REQ-2: The lane-loop-line command shall refuse the host codex or omp with one line that names claude as the only host and exit non-zero.
- REQ-3: The lane loop module shall carry no Codex or OMP host, no hook-file shape, no blocked line and no goal text and keep the helpers its other users call.
- REQ-4: The lane-duty-line command shall print the Claude line as it does today and refuse the host codex or omp as a usage error.
- REQ-5: The lane-stop-hook command shall answer with one line on standard error saying the host was removed, nothing on standard output, and exit zero.
- REQ-6: The lane loop module shall carry no Stop-hook state, no block cap and no snapshot reader once the lane-stop-hook command no longer calls them.
- REQ-7: The goal-block command shall keep the leader condition, the agent loop line, the goal lines and the counts and carry no codex key in its answer.
- REQ-8: The leader-round command shall refuse an observation whose host is codex or omp and recognise only the Claude bullet and prompt marks.
- REQ-9: The lane-idle line shall name only the Claude spelling of the duty line.
- REQ-10: Every help line and operation document of the commands in this area shall name no Codex or OMP host.

## Acceptance

- AC-1 (REQ-1): the whole Claude answer of `lane-loop-line` is pinned and unchanged by the removal · verify: `cargo test --locked --offline --test lane_loop_cli the_claude_answer_is_pinned_whole`
- AC-2 (REQ-2): `--host codex` and `--host omp` each give one line naming claude and a non-zero exit, and `--host claude` still works · verify: `cargo test --locked --offline --test lane_loop_cli a_retired_host_is_refused_with_one_line`
- AC-3 (REQ-3): the lane loop unit tests and every other user of its helpers pass · verify: `cargo test --locked --offline --lib lane_loop` and `cargo test --locked --offline --test lane_loop_cli resume_note_render_then_check`
- AC-4 (REQ-4): the Claude duty line is unchanged and a retired host is a usage error · verify: `cargo test --locked --offline --test lane_duty_line`
- AC-5 (REQ-5): `lane-stop-hook` prints one line on standard error, nothing on standard output, exits 0 and writes no file · verify: `cargo test --locked --offline --test lane_loop_cli the_lane_stop_hook_command_answers_one_line_and_exit_zero`
- AC-6 (REQ-6): the Stop-hook machinery is gone from the module and the crate compiles with all targets and no dead code · verify: `cargo clippy --locked --offline --all-targets -- -D warnings` and `cargo test --locked --offline --lib lane_loop`
- AC-7 (REQ-7): the goal block answer keeps its Claude fields and has no codex key · verify: `cargo test --locked --offline --test goal_block_cli`
- AC-8 (REQ-8): a leader round observation with the host codex is refused with a message naming claude · verify: `cargo test --locked --offline --test leader_round_cli a_removed_host_observation_is_refused` and `cargo test --locked --offline --lib leader_round`
- AC-9 (REQ-9): the lane-idle line has no Codex wording and its prefix is unchanged · verify: `cargo test --locked --offline --lib lane_idle` and `cargo test --locked --offline --test lane_next_cli`
- AC-10 (REQ-10): the help goldens and the operation documents carry no Codex or OMP host word of this area · verify: `cargo test --locked --offline --test help_layout` and `cargo test --locked --offline --test cli_text_citations`

## Scope

- write: `.moai/specs/SPEC-20261121-lane-goal-leader-codex-removed/**`
- write: `src/lane_loop.rs`, `src/ops/lane_loop_line.rs`, `src/lane_duty.rs`, `src/ops/lane_duty_line.rs`, `src/goal_block.rs`, `src/lane_idle.rs`, `src/leader_round.rs`, `src/leader_pulse.rs`, `src/ops/leader_pulse.rs`
- write: `tests/lane_loop_cli.rs`, `tests/lane_stop_hook_cli.rs`, `tests/lane_duty_line.rs`, `tests/goal_block_cli.rs`, `tests/leader_round_cli.rs`, `tests/fixtures/lane_loop/`, `tests/fixtures/leader_round/screens/synthetic-lane-compact-codex.json`, `tests/fixtures/leader_round/screens/synthetic-lane-compact-omp.json`
- write: `docs/design-review/appendix/operations/lane_loop_line.md`, `lane_duty_line.md`, `goal_block.md`, `leader_round.md`, `leader_pulse.md`, `lane_next.md`
- write, after the holder released them or paired by line range: `src/main.rs` and `src/help.rs` only at the lines of the `lane-stop-hook` command and the `--host` value lists of these commands, and `tests/fixtures/help/` only at the golden lines of those commands
- read-only: every other file, including `src/hooks.rs`, `src/lib.rs`, `src/go_state.rs`, all shipped skill text and `tests/lane_next_cli.rs` (reserved by another card; it pins only the prefix of the lane idle line, so it is run and not edited)
- forbidden: the launcher and home-update sources, the hook adapters of the hooks slice, `skills/moai/**`, any card or SPEC that describes Codex work

## Plan

1. SPEC first (this file), from the planning read of the tip and the official-documentation research for the goal block.
2. Baseline commit before any deletion: a test that pins the whole Claude answer of `lane-loop-line` (`host, lane, writes, lines, verified, note`), and one for the Claude duty line.
3. Free files in one commit so the crate compiles: the Codex and OMP host, arms, hook-file shape, blocked line, goal text and `END_CONDITION` of the lane loop module; the `--project` flag and the Codex block of `lane-loop-line`; the `lane_duty` host rows; the Codex mark characters and host list of the leader round; the Codex tail of the lane idle line; the `codex` key of the goal block; their tests, fixtures and operation documents.
4. When `src/main.rs` and `src/help.rs` are free or paired: the `lane-stop-hook` command becomes the one-line answer, the Stop-hook machinery of the lane loop module goes, `tests/lane_stop_hook_cli.rs` and its fixtures go, the `READ_BEFORE_DISPATCH` list and the help goldens follow, and the clap `--host` lists narrow.
5. When the hooks slice has removed the `LaneLoop` adapter: the three Stop-hook items of the lane loop module go with it.
6. Targeted tests, clippy with all targets, rustfmt, `spec check`, the affected set once, then the independent audit (tier 1).

## Risks

- A running lane breaks if the Claude line changes. Mitigation: AC-1 pins the whole answer first, and the Claude tests named in the Plan stay green at every commit.
- The Stop machinery cannot go while `src/main.rs` calls it. Mitigation: Plan step 4 follows the release or the pair of that file; step 3 compiles without it.
- A leftover Codex Stop hook in a user's `.codex/hooks.json` still calls `moai-cli lane-stop-hook`. Mitigation: the command answers one line and exits zero (REQ-5, D-4), so a stale hook neither blocks nor fails; the user's file is not touched.
- `tests/lane_stop_hook_cli.rs` is also on the hooks slice's list. Mitigation: whichever card lands first deletes it and says so in a card note (D-5).
- An unmerged branch of a withdrawn card rewrites `lane-stop-hook`. Mitigation: it is history and not part of this change.
- `tests/omp_host_survey_test.py` reads the OMP row of `src/lane_duty.rs`. Mitigation: that file belongs to the shipped-text slice; this change lands after its deletion or removes only that method and names it.
- The planning read ran no build; every compile claim is checked by the targeted runs of Plan step 6.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the goal block stays and loses only its Codex line and caveat | the official Claude Code documentation has a built-in `/goal [condition\|clear]` command that "keeps working across turns until the condition is met", set by the user in a session; source https://code.claude.com/docs/en/goal and https://code.claude.com/docs/en/commands, read 2026-11-21T14:32Z to 14:34Z, since v2.1.139 per https://code.claude.com/docs/en/whats-new/2026-w20. Not settled by the documents: whether the model can set a goal itself; the block keeps `model_can_input: false` | remove the goal block later |
| D-2 | a command-line host word codex or omp is refused with one line naming claude, not skipped and not answered with a Claude line | the retired-host layer is for stored files; a Claude line for another host would mislead | print the Claude line |
| D-3 | the leader round refuses an observation whose host is codex or omp, and drops the Codex bullet and prompt mark characters | nobody has such a file; a layer for it has no user | skip the lane with a reason |
| D-4 | the lane-stop-hook command stays as a one-line answer on standard error with exit zero, like the removed host answer of hook-dispatch | a stale Codex Stop hook must neither block nor fail | delete the command |
| D-5 | `tests/lane_stop_hook_cli.rs` is deleted by whichever card lands first, and that card says so on the other | two cards list it | none needed |
| D-6 | the `lane-stop-hook` entry of the state-folder move list in `src/go_state.rs` stays | it moves users' old state folders and a test pins it | remove with the Go migration |
| D-7 | `--host` of `lane-loop-line` stays a required string refused at run time until the hooks slice lands; the clap value lists narrow after | the help examples with `--host codex` are parsed by a test in `src/main.rs`; no existing call changes | make the flag optional |
| D-8 | `END_CONDITION` and the Codex goal text go in both outcomes of D-1 | only the Codex text and the Stop machinery use them | restore from history |

## Result

Not started.
