---
id: SPEC-20261122-bg-session-research
title: "Background sessions (claude --bg, agent view, claude agents --json) are measured in temporary repositories before any lane uses them"
status: draft
revision: 1
created: 2026-11-22
---

# Background sessions (claude --bg, agent view, claude agents --json) are measured in temporary repositories before any lane uses them

## Intent

User pick of 2026-11-22: Orca stays the lane host, and Claude Code's own background sessions are measured as a second host in temporary repositories. Remote Control server mode is not used. The result is a report that lists each measured behaviour with the Claude Code version it was measured on, and recommends keep, trial or drop with reasons. No lane is moved by this work.

Read whole on 2026-11-22 before this SPEC: the official `agent-view.md` (1,140 lines) and `worktrees.md` (442 lines) of code.claude.com/docs/en, and the whole-docs report section on lanes (`~/.moai/reports/moai-adk/claude-code-native-redesign-2026-11-22.md`, sections 2-4, 6-1 to 6-4 and the pick list). The installed Claude Code is 2.1.330, newer than every version the two pages name (the newest history row is v2.1.323), and `claude --help` lists `--bg`, `agents`, `attach`, `logs`, `stop`, `respawn`, `rm` and `daemon`.

What the pages say, which this work checks and does not trust:

- A background session moves into a worktree under `.claude/worktrees/` before its first edit, unless `worktree.bgIsolation` is `"none"`; the page names no worktree or branch pattern for it.
- By default such a session "commits without asking, and pushes the branch when the repository has a remote", may open a draft pull request, and leaves git alone when the task, `CLAUDE.md` or memory says the user handles it; rules files, settings and `includeGitInstructions` are not named.
- `claude agents --json` is "the supported way to read session state from outside Claude Code": fields `cwd`, `kind`, `startedAt`, and for background sessions `id`, `state` (`working`, `blocked`, `done`, `failed`, `stopped`), and while the process lives `pid`, `status` (`busy`, `waiting`, `idle`) and `waitingFor`; the files under `~/.claude/jobs/<id>/` are "not a stable interface"; the page gives no result or transcript field and no transcript path.
- An unattached idle session is stopped after about an hour and a pinned one is kept; `claude rm` keeps a worktree with uncommitted changes and refuses unpushed commits unless told otherwise.
- The page says agent view is in research preview and its interface may change.

The page leaves open, and the report must answer by measurement: the worktree and branch name of a background session, whether `.worktreeinclude` applies to it, what it does in a repository with no remote and under the flow rule of the project's `CLAUDE.md`, the exact `--json` output with no session and with one of each state, signal behaviour (interrupt versus terminate), the transcript location, the environment a session inherits, `CLAUDE_CODE_CHILD_SESSION`, the idle stop, and `claude rm` on a lane that was never pushed.

This SPEC is slice 1: the SPEC only. The measurement harness and the report follow in the slices of the Plan.

## Non-goals

- Moving a lane, a card or a worktree to a background session, or changing any moai command, skill or default.
- Remote Control server mode, cloud sessions, the desktop app and the agent view terminal screen (a person's interface, not a lane's).
- Any write to the real home folder, the real Claude config folder, a real repository or a real ledger.
- Measuring a model's answer quality: a session that needs a model reply is measured only with a provider the leader approves (D-4).

## Requirements

- REQ-1: When the measurement harness runs, the harness shall use one temporary repository, one temporary home and one temporary Claude config folder for each measurement and leave the real home, the real config folder and every real repository as they were · failure: none — a harness that cannot create the temporary places measures nothing and says so.
- REQ-2: When a measurement starts a supervisor or a background session, the harness shall stop them and remove their records before the next measurement and report anything it could not stop · failure: none — a process that survives is named with its pid and left for the person.
- REQ-3: When a behaviour is measured, the report shall name the measurement, the command, its exit status, its output and the Claude Code version printed by `claude --version` at that run · failure: none — a measurement that was not run is listed as not run with the reason, never as passed.
- REQ-4: When the placement of a background session is measured, the report shall record the working directory before the first edit, the worktree path, the branch name, the lock state and whether `.worktreeinclude` files arrive, each read from git or the file system after the session · failure: none — a session that edits nothing records that no worktree was made.
- REQ-5: When the git behaviour of a background session is measured, the report shall record, for a repository with no remote, for a repository with a remote and for each of the three instruction places (task text, `CLAUDE.md` line, memory entry), whether it committed, pushed or opened a pull request, read from `git log`, the remote and the session's final report · failure: none — a measurement that needs a model reply runs only under D-4.
- REQ-6: When `claude agents --json` is measured, the report shall record the exact output and exit status with no session, with each of the states that can be produced and with `--all` and `--cwd`, and name every field it saw · failure: none — a state that cannot be produced is listed as unobserved.
- REQ-7: When stopping and resuming are measured, the report shall record, for `claude stop`, `claude kill`, an interrupt signal and a terminate signal sent to the session process, the resulting `state`, whether the session restarts, and what `claude --resume` and `claude respawn` do afterwards · failure: none — a signal that cannot be sent safely is listed as not run.
- REQ-8: When the idle stop is measured, the report shall record the session's `status` and process before and after the documented idle time, with a pinned and an unpinned session · failure: none — the wait runs as a background job, and a run that was not waited out is listed as not run.
- REQ-9: When removal is measured, the report shall record what `claude rm` and the agent view delete do to a worktree with uncommitted changes, with commits that were never pushed and with a lock, and which flags change the answer · failure: none — nothing is removed outside the temporary places.
- REQ-10: When the environment of a background session is measured, the report shall record which of the supervisor's environment, the dispatching shell's environment, the project `env` settings and the temporary config folder the session saw · failure: none — a variable that holds a secret is recorded by name and presence only and never by value.
- REQ-11: When every measurement is done, the report shall recommend keep, trial or drop for background sessions as a lane host beside Orca, with the reasons taken from the measured rows and the versions they were measured on · failure: none — a row that was not run is named as the reason a recommendation is limited.
- REQ-12: When the report, the harness or any answer is printed or saved, the text shall hold no card number, no key and no path of the real home or real config folder · failure: none — paths are given relative to the temporary root.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-12): the harness test runs the harness against a stub `claude` in a temporary root, finds the stub's records only under the temporary root, and finds the real home and config folder unchanged (a listing before and after is equal); a stubbed leftover process is reported by pid · verify: `cargo test --locked --offline --test bg_session_probe_cli isolation_`
- AC-2 (REQ-3, REQ-6): the report builder test turns a recorded run into rows with the command, status, output and version, lists a measurement that was not run as not run, and prints the recorded `--json` output unchanged · verify: `cargo test --locked --offline --test bg_session_probe_cli report_`
- AC-3 (REQ-4, REQ-5, REQ-7, REQ-9, REQ-10): each of these measurements has a probe whose stub run records the rows it names, and a secret-shaped variable value never appears in the rows · verify: `cargo test --locked --offline --test bg_session_probe_cli rows_`
- AC-4 (REQ-8): the idle probe starts as a background job, records both readings and is listed as not run when the wait was not completed · verify: `cargo test --locked --offline --test bg_session_probe_cli idle_`
- AC-5 (REQ-11): the saved report holds one recommendation of keep, trial or drop, each reason cites a measured row, and no lane, card or default was changed by this work (the diff of the work names only the SPEC, the harness, its tests and the report) · verify: `moai check affected v2026.11.0..HEAD --dir . --max 2` and `git diff --stat v2026.11.0..HEAD`

## Scope

- write: `.moai/specs/SPEC-20261122-bg-session-research/spec.md`
- write (slices 2 and 3): `tests/bg_session_probe_cli.rs`, `src/ops/bg_session_probe.rs` (read-only probe plumbing, no change of any shipped command), and the report under `~/.moai/reports/moai-adk/`
- last step, by pair request: `src/main.rs` and `src/help.rs` only if the harness is exposed as a developer command; it need not be

## Decisions

- D-1: The harness is development tooling, not a product feature: nothing here changes what moai installs or does for a user, which keeps the product generic and free of dev tools.
- D-2: Every measurement runs with a temporary `HOME`, a temporary `CLAUDE_CONFIG_DIR` and a temporary repository, started from a scrubbed environment; a supervisor started this way is a separate instance (the page says so), so it cannot list or stop a session of the real config folder. The harness stops it with `claude daemon stop --any` and with a check of the process table.
- D-3: Measurements run in order of risk: first those that need no model and no credentials (empty `--json`, daemon status, placement with a stub, signals on a process that needs no model, delete rules on a hand-made worktree), then those that need a model reply.
- D-4: A measurement that needs a model reply (placement by a real edit, git behaviour, instruction places, idle stop with a real turn) spends the user's usage and sends a prompt to a provider, so it is a decision for the leader before it runs; the default if nobody answers is not to run it and to list those rows as not run. A local stand-in provider is preferred if the leader approves one. The first report may therefore be a report of the model-free rows plus a plan for the rest.
- D-5: The research-preview status makes every row a dated fact about one version, never a rule: the report carries the version per row, and the recommendation is `trial` at most unless every lane-critical row was measured and held.
- D-6: Idle stop and the lock sweep take long wall time; they run as background jobs and may be left to a later slice rather than made to finish in one turn.
- D-7: A claim of the page that a command exists (for example a shell-job flag) is checked against `claude --help` of the installed version before a probe uses it; a flag that the installed help does not list is not used.

## Plan

1. This SPEC, read by the leader and by the audit before any code or probe runs.
2. Slice 2, tests first (RED): `tests/bg_session_probe_cli.rs` `isolation_` and `report_` against a stub `claude`; then the probe plumbing and the report builder.
3. Slice 3: the model-free measurements on 2.1.330 into the report (REQ-4 to REQ-10 as far as no model is needed), with `not run` rows named.
4. Slice 4: the model-needing rows only after the leader's answer on D-4; then the recommendation.
5. Per slice: light checks, at most the mutant cap, one affected run, then the READY note with head and base only.

## Risks

- A supervisor or session left running keeps using the machine and, with real credentials, usage: the harness uses no credentials by default, stops what it starts and checks the process table afterwards.
- A probe that reaches the real config folder through an inherited variable would touch real sessions: the environment is rebuilt from a short list and the first test pins that the real folders are unchanged.
- The pages and the preview change quickly (the history table shows many fixes within weeks); rows carry the version and a re-run is cheap, so the report is dated and not a standing rule.
- Deleting or removing in a measurement could hit a real worktree: every removal takes a path that the harness itself created under the temporary root, and a path outside it refuses.
- A measurement on macOS says nothing about Linux or Windows, and the report says which platform each row was measured on.
- The agent's reading of a page can differ from the page: facts above carry the page's own words and line numbers in the research record, and the report re-checks each against the measured behaviour.
