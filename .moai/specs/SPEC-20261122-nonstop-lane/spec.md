---
id: SPEC-20261122-nonstop-lane
title: "Non-stop lane running is a built-in choice: one selection, one stored pick, one generated loop prompt and one ledger watch"
status: draft
revision: 2
created: 2026-11-22
---

# Non-stop lane running is a built-in choice: one selection, one stored pick, one generated loop prompt and one ledger watch

## Intent

User request of 2026-11-22: offer running a lane without stopping as a choice that ships with moai. When `/moai work` starts a lane (and once at `/moai project`) moai asks the user with a selection, "run this lane non-stop" (recommended when there are several lanes) or "one card at a time". Non-stop means the lane runs Claude Code's own `/loop` in self-paced mode with a prompt that moai generates: keep the current card; after a READY note take the next card the ledger names for this lane; while waiting for an audit verdict or an answer, watch the ledger notes of its own cards with Monitor and work the next card meanwhile; never end a turn with neither a card in progress nor a watch; a needed decision becomes a blocked note. The pick is stored in the profile, changed by a command without asking again, and can also be asked per lane. On the user's pick, Claude Code's `/goal` may also be used per card. Only Claude Code's native `/loop`, Monitor and `/goal` are used; the ledger stays the record and nothing polls outside Monitor.

Read whole on 2026-11-22 before this SPEC (the pages of code.claude.com/docs/en: scheduled-tasks, goal, tools-reference) and the facts this SPEC rests on:

- A `/loop` with no interval is self-paced: after each iteration Claude picks the next delay between one minute and one hour with `ScheduleWakeup`, may use Monitor directly, and ends the loop by calling `ScheduleWakeup` with `stop: true`. An iteration that ends without scheduling the next one stops the loop (one extra iteration about 20 minutes later only with `CLAUDE_CODE_LOOP_KEEPALIVE`).
- A recurring loop expires seven days after its creation and the self-paced loop is not restored by `--resume` or `--continue`; tasks fire only while Claude Code runs and is idle, and each fire goes through the session's permission checks. `CLAUDE_CODE_DISABLE_CRON=1` removes `/loop` and `CronCreate`.
- A scheduled fire runs only skills Claude may invoke on its own; built-in commands, skills with `disable-model-invocation` and MCP prompts reach it as plain text.
- Monitor runs a background script and gives Claude each output line; a watch ends at its deadline (5 minutes by default, at most 30, at most 10 in a `-p` run) and Claude gets one notice, so it can start the watch again. It uses Bash's permission rules; on Bedrock, Claude Platform on AWS, Google Cloud's Agent Platform, Microsoft Foundry, or with telemetry or non-essential traffic off, it needs Claude Code v2.1.310 or later (the dynamic loop v2.1.248 or later); on Windows it needs Git Bash.
- `/goal` is a session-scoped prompt-based Stop hook: a small fast model checks the condition after every turn (extra billed tokens), the condition is at most 4,000 characters and may carry a turn clause, one goal is active per session, evaluation waits while a subagent or a background command runs (check-ins after 30 minutes), and it is unavailable when `disableAllHooks` is true or `allowManagedHooksOnly` is set. Claude cannot start `/goal` itself: the user types it.

The code already holds the older parts this SPEC replaces for Claude Code: `lane-loop-line` prints a fixed five-minute `/loop` for the Issue era, and the MODE-AGENT workflow names the wake and keep-going rules in text. This SPEC adds the stored pick, the generated card-ledger prompt and the ledger watch; the Codex and OMP arms of `lane-loop-line` belong to the removal cards and are not touched here.

This SPEC is slice 1 of the work: the SPEC only. The code, the skill text and the tests follow in the slices of the Plan.

## Docs

What a user reads once this lands (written in the operation document and, in slice 4, the help text and the skill):

- `/moai work` asks once: run this lane non-stop (recommended with several lanes) or one card at a time. The answer is saved in the project profile; `moai lane-run set nonstop|single [--lane <name>]` changes it without asking again, and `moai lane-run show [--lane <name>]` tells the pick in force and why.
- With the non-stop pick, `moai lane-run prompt --lane <name>` prints the text the lane gives to Claude Code's own `/loop`; it keeps working cards from the ledger and watches its own cards' notes with Monitor. A loop spends tokens and runs tools while nobody watches, needs auto mode or allow rules for unattended turns, ends after seven days and is not restored by a resume.
- `CLAUDE_CODE_DISABLE_CRON=1` turns the pick into one card at a time, and `show` says so.
- `moai lane-run goal --card <name>` prints a `/goal` condition for one card; the user types `/goal` themselves, and a goal adds a second model's check and its tokens.

## Non-goals

- Starting `/loop` or `/goal` from moai: both are Claude Code commands the lane or the user types, and moai only prints their text.
- Any scheduler, daemon or resident process of moai, and any polling outside the Monitor tool.
- A cloud schedule or a Desktop scheduled task: the loop is session-scoped by choice and the seven-day expiry is told, not worked around.
- Changing the permission mode: a non-stop lane inherits the session's permissions, and the question tells the user that unattended turns need auto mode or allow rules.
- The Codex and OMP hosts, and the older Issue-era text of `lane-loop-line`.
- Orchestration of several lanes by one agent and the background-session research: other work.
- Choosing which card a lane takes: the ledger's `card next` does that.

## Requirements

- REQ-1: When a lane is started with `/moai work` or a project is set up with `/moai project`, the skill shall ask once, as a selection with the recommended option first, whether to run the lane non-stop or one card at a time and save the answer in the profile · failure: none — a pick that is already saved is not asked again, and a run in the agent mode, where nobody can answer, asks nothing and uses the saved pick or `single`.
- REQ-2: When the saved pick is changed with the `lane-run set` command, the lane-run command shall write the new value for the project or for the named lane without asking again and leave every other profile key as it was · failure: none — a value other than `nonstop` or `single` refuses before any write.
- REQ-3: When the effective pick of a lane is read, the lane-run command shall give the lane's own value before the project's value and report `unset` when neither exists · failure: none — an unreadable profile reports the reason and no pick.
- REQ-4: When the pick is `nonstop`, the lane-run prompt command shall print one plain-text prompt for the self-paced `/loop` that names the lane, keeps the current card, takes the next card of `moai card next` after a READY note, watches the ledger notes of the lane's own cards with the watch command while it waits, handles a verdict or a repair request first, turns a needed decision into a blocked note and never ends a turn with neither a card in progress nor a watch · failure: none — a lane name that is not accepted refuses before printing.
- REQ-5: When the pick is `single`, the lane-run prompt command shall print no loop prompt and say that no loop is started · failure: none — the answer names the pick it read.
- REQ-6: When the generated prompt is followed and `moai card next` names no card and no pull request of the lane is open, the prompt shall make the lane print one idle line, stop the watch and end the loop with `ScheduleWakeup` and `stop: true` · failure: none — the idle line is printed once for the same state.
- REQ-7: When the watch command runs, the watch command shall read the named cards read-only through the card timeline cursor and print one line for each note that is new since it started, and nothing for no change · failure: none — a card that cannot be read prints one `unreadable` line and the watch goes on with the others.
- REQ-8: When the watch command reaches its own deadline, the watch command shall exit with status 0 before the Monitor deadline so that the lane gets the expiry notice and starts it again · failure: none — a deadline outside 60 to 1,700 seconds refuses before reading.
- REQ-9: When a host fact makes the non-stop pick impossible, the lane-run show command shall name the fact and report the effective pick as `single` · failure: none — the facts are `CLAUDE_CODE_DISABLE_CRON` set to a true value and a loop that Claude Code cannot be asked for, and a fact that cannot be read is reported as unread and never as met.
- REQ-10: When the user picks `/goal` for a card, the lane-run goal command shall print one condition of at most 4,000 characters that names the card, one measurable end state, the check that proves it and a turn limit · failure: none — the command prints the text only and never starts a goal.
- REQ-11: When a lane that holds the non-stop pick starts a turn with no loop pending, the skill shall print the prompt for the user to paste or follow it itself and say that the loop is gone after a resume and after seven days · failure: none — a lane in the single pick starts nothing.
- REQ-12: When the generated prompt, the help text or any answer of the lane-run commands is printed, the text shall hold no card number, no key and no path outside the project · failure: none — the card is a placeholder and the lane name comes from the caller.

## Acceptance

- AC-1 (REQ-1, REQ-11): the `mode-agent` and project workflows of the skill carry the selection with the recommended option first and the rule that an unanswerable run asks nothing; a test reads the skill files from the package and finds both texts, and the source manifest lists the changed files · verify: `cargo test --locked --offline --test lane_run_cli skill_text_` and the skill package manifest test
- AC-2 (REQ-2, REQ-3): in a temporary project, `lane-run set nonstop` and `lane-run set single --lane <name>` write the project value and the lane value and no other key changes (the file is compared before and after outside the `lane_run` key), the effective pick is the lane's before the project's, `unset` is reported when neither exists, a wrong value or a project without a profile refuses with nothing written, and the written profile still reads through the profile reader (an unknown field under `lane_run` refuses) · verify: `cargo test --locked --offline --test lane_run_cli pick_`
- AC-3 (REQ-4, REQ-6, REQ-12): with `nonstop`, the prompt names the lane, the current card rule, `moai card next`, the watch command, the blocked-note rule, the never-end-a-turn rule and the idle stop with `stop: true`; it holds no card number, no key-shaped text and no absolute path outside the project; with `single`, no prompt is printed · verify: `cargo test --locked --offline --test lane_run_cli prompt_`
- AC-4 (REQ-7, REQ-8): in a temporary ledger with two cards, the watch prints one line for each new note of either card and nothing for a quiet minute, an unreadable card prints one line and the other is still watched, and the watch exits 0 at its deadline; a deadline of 59 or 1,701 seconds refuses before reading · verify: `cargo test --locked --offline --test lane_run_cli watch_`
- AC-5 (REQ-9): with `CLAUDE_CODE_DISABLE_CRON=1` the show answer names that fact and the effective pick reads `single` although the saved pick is `nonstop`; with it unset the saved pick stands, and a fact that cannot be read is reported unread · verify: `cargo test --locked --offline --test lane_run_cli facts_`
- AC-6 (REQ-10): the goal text is at most 4,000 characters for the longest accepted card name, carries the end state, the check and the turn limit, and the command starts nothing · verify: `cargo test --locked --offline --test lane_run_cli goal_`
- AC-7 (REQ-12): the help text and the operations document of the four commands carry no card number and examples that run (the example rules of the help test hold) · verify: `cargo test --locked --offline --test card_cli` and `--test lane_run_cli help_`

## Scope

- write: `.moai/specs/SPEC-20261122-nonstop-lane/spec.md`
- write (slice 2, revision 2, leader decision on the ledger): `src/profile.rs` (the optional top-level `lane_run` key in the reader, its table row, its check and the profile's own test literals; the profile also gains another lane's field, so the later landing side matches the literals)
- write (slices 2 to 4): `src/ops/lane_run.rs` (the operation and its logic in one file, so `src/lib.rs` is not touched), `tests/lane_run_cli.rs`, `docs/design-review/appendix/operations/lane-run.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/project.md` and the skill source manifest
- last step, by pair request: `src/main.rs` and `src/help.rs` (the four actions and their examples)

## Decisions

- D-1: One operation `lane-run` with the actions `show`, `set`, `prompt`, `watch` and `goal`, in the flat style of `lane-loop-line` and `goal-block`. `show` and `prompt` and `goal` read only; `set` is the only write; `watch` reads only and runs until its deadline.
- D-2 (revision 2): The pick lives in one optional top-level profile key `lane_run` with `default` (`nonstop` or `single`, the project's pick) and `lanes` (lane name to `nonstop` or `single`). Revision 1 put it under `work_flow`, but that key requires `flow` and refuses unknown fields, so a project without a recorded flow could not hold the pick and a written pick would make every later profile read fail; the leader picked the top-level key. The profile reader learns the key in the same slice that first writes it (an absent key reads as before, an unknown field under it refuses). `set` rewrites the profile file by a temporary file and a rename, keeping every other key and its order; a project without a profile refuses.
- D-3: The pick is asked where a person can answer, as a selection with `nonstop` first and marked recommended when the project has several lanes; the question says in one line that unattended turns need auto mode or allow rules for the moai commands and that the loop ends after seven days or a resume. In the agent mode, where the lane may ask no user, an unset pick is `single`: a started loop spends tokens and permission, so it needs a person's yes; the leader decides whether the leader's own lanes get an unset default of `nonstop` (open decision for the leader, reversible).
- D-4: The generated prompt is plain text, never a slash command, because a fire runs only skills Claude may invoke on its own and `/moai` may not be one; it tells the lane to use the host's `/loop` self-pacing and to read the ledger through moai commands. It names no card, no row of cards and no leader: the next card is whatever `moai card next` names.
- D-5: The watch is one command that Monitor runs: it keeps one cursor per card from `card timeline` (`--after` never waits and has no gap or overlap), reads every 30 seconds with the ledger opened read-only, prints a line `<card> note <id> <kind> <first line>` for each new note only, and ends at its own deadline (default 1,500 seconds, at most 1,700) so the Monitor deadline of 30 minutes is never the one that ends it. It starts after the notes that already exist, so a restart prints no old note.
- D-6: Facts are read, not guessed: the environment variable, and the settings files read-only for `disableAllHooks` and `allowManagedHooksOnly` when the goal is asked. The Claude Code version floors (v2.1.248 for the dynamic loop and v2.1.310 for Monitor on the providers that fetch no flags) are printed as text in `show`, and the version itself is reported `unread` because moai reads no host version.
- D-7: `/goal` is optional and per card, on the user's pick, and the text is the lane's end state for one card (the READY note exists with its head and base read from git, and the card is in review), with a turn clause. It adds a separate model's judgement and extra tokens, so the question names that; it is never the default.
- D-8: The Issue-era `lane-loop-line` stays for the removal cards to retire; the skill's text stops pointing at its Claude arm for new lanes once this lands.
- D-9: No lane identity is invented: the lane name is the one the lane registered with `session-register`, passed by the caller.

## Plan

1. This SPEC, read by the leader and by the audit before any code.
2. Slice 2, tests first (RED): `tests/lane_run_cli.rs` `pick_`, `facts_`, `prompt_` and `goal_` on a temporary project; then the profile reader key, the operation `src/ops/lane_run.rs` and its document.
3. Slice 3, tests first: `watch_` on a temporary ledger with a card writer between reads; the watch command and its deadline.
4. Slice 4, tests first: `skill_text_` and `help_`; the selection text in the project and agent workflows, the source manifest, the help examples and the pair request for `src/main.rs`, `src/help.rs` and the profile reader.
5. Per slice: light checks, at most the mutant cap, one affected run, then the READY note with head and base only.

## Risks

- A loop that a person does not watch spends tokens and runs tools: the selection says so, the pick is never made by moai, and an unanswerable run uses `single`.
- The self-paced loop is not restored on resume and ends after seven days; the skill re-offers the prompt at a turn that finds none pending, and the SPEC does not claim a resume.
- `ScheduleWakeup`, Monitor and `/goal` are host features whose behaviour changed within weeks in the pages read; the tests pin only moai's own texts and commands, and the prompt names the tools by their documented names so a renamed tool shows as a host change and not as a silent failure.
- A lane that uses `/goal` and a loop together can have the goal evaluation deferred while the Monitor runs as background work; the goal text carries a turn clause and the question names that the pair is an advanced pick.
- A third-party provider may not fetch feature flags, which raises the minimum versions; `show` prints the floors and a wrong guess shows as a Monitor or `/loop` error that the lane reports as a blocked note.
- The profile file is the project's and is shared by moai commands: the write keeps unknown keys and the temporary-file rename, and a concurrent writer loses at worst one of two picks of the same key.
