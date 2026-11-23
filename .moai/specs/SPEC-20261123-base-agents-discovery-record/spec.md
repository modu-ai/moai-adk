---
id: SPEC-20261123-base-agents-discovery-record
title: "Recorded check that a real Claude Code finds the base agents and the project roles in a native config folder and in an Orca-profile-shaped one, with no name hidden"
status: draft
revision: 1
created: 2026-11-23
---

# Recorded check that a real Claude Code finds the base agents and the project roles in a native config folder and in an Orca-profile-shaped one, with no name hidden

## Intent

`SPEC-20261121-base-agents-user-folder` moved the five base agents (`moai-plan`, `moai-research`, `moai-impl`, `moai-audit`, `moai-docs`) to `~/.claude/agents/moai/` and the project roles to `.claude/agents/moai/`, and landed (its AC-4 check lists the agents of a native temporary config folder only). What the first card criterion still lacks is a recorded check that the real Claude Code finds and selects both kinds in the two shapes the user's machine has:

- native: `CLAUDE_CONFIG_DIR` unset or `~/.claude` itself;
- an Orca profile: `CLAUDE_CONFIG_DIR` names a profile folder whose `agents` entry is a link to `~/.claude/agents` (measured by the leader on 2026-11-23: the profile's agents, skills, output-styles, plugins and projects are links into `~/.claude`; the user's decision of the same day makes `~/.claude` the one source of truth and the profile only a reader).

The check runs only on temporary folders, with the environment cleared; the real home, `~/.claude` and the Orca profile are neither read nor written.

What was measured while writing this SPEC (Claude Code 2.1.330, temporary HOME, built binary of this tree):

- `claude --agent <unknown name> -p hi` prints `--agent '<name>' not found. Available agents: …` and exits before any model call; the list held the five base agents, the project role, a personal agent and Claude Code's own built-in ones, in both shapes.
- `claude --agent <known name> -p hi` passes the name check and stops at `Not logged in · Please run /login`, because a temporary config folder holds no credentials and credentials are never copied. So the check can show that a name is found and selected, and cannot show a model answer.
- The install honours the profile shape on its own: with `CLAUDE_CONFIG_DIR` naming the profile, a plan whose `agents` link points at `~/.claude/agents` targets `~/.claude/agents/moai`; a link that points at any other folder blocks the host with a message naming both folders.

## Non-goals

- A model-answered call of a base agent: it needs credentials, and credentials are never copied into a temporary folder; recorded as `not_run` with this reason.
- Any change of the install, the generation or the hooks; this change adds a test and this SPEC only.
- The real home, `~/.claude`, the Orca profile and Orca settings: never read or written.
- The common rules (`~/.claude/rules`) and the hooks of the user scope: other cards.

## Requirements

- REQ-1: When the base agents are installed into a temporary home with a native config folder, the installed Claude Code shall list the five base agents, a personal agent and a project role of a temporary project, each name once, in its available-agents list. · failure: none — the test fails and prints the list it saw
- REQ-2: When the config folder is an Orca-profile shape whose `agents` entry links to the user's agents folder, the installed Claude Code shall list the same names. · failure: none — the test fails and prints the list it saw
- REQ-3: When `claude --agent <name>` runs for each of the six names in either shape, the installed Claude Code shall pass the name check without a "not found" answer. · failure: none — the test fails and prints the answer
- REQ-4: IF the `agents` link of the profile points at another folder, THEN the install shall block that host so that the available-agents list holds no base agent.
- REQ-5: The check shall run on temporary folders with a cleared environment, so that no file outside them is read or written.
- REQ-6: IF `claude` is not on the path, THEN the check shall print that it skipped and pass. · failure: none — nothing runs
- REQ-7: When a known agent name runs through the child `claude`, the check shall end the run at the login step within a fixed deadline, with the background traffic switched off. · failure: none — the test fails and prints the answer

## Acceptance

- AC-1 (REQ-1, REQ-5): in a native temporary config folder the list of `claude --agent <unknown name>` holds the five base agents, the personal agent and the project role, each once · verify: `cargo test --test base_agents_discovery_record native_layout`
- AC-2 (REQ-2, REQ-5): the same list through a profile folder whose `agents` links to the canonical agents folder, installed with the profile as `CLAUDE_CONFIG_DIR`, with the profile's `agents` still a link and no file written into the profile · verify: `cargo test --test base_agents_discovery_record orca_profile_layout`
- AC-3 (REQ-3): all six names pass the name check in both shapes · verify: `cargo test --test base_agents_discovery_record`
- AC-4 (REQ-4): with a profile link to another folder the preview is blocked and the list holds no base agent · verify: `cargo test --test base_agents_discovery_record a_profile_link_to_another_folder`
- AC-5 (REQ-6): a `PATH` that is unset, empty or holds no `claude` finds no program, and each case then prints `skipped: \`claude\` is not on the PATH` and returns · verify: `cargo test --test base_agents_discovery_record no_claude_on_the_path_is_a_skip`
- AC-6 (REQ-7): the child `claude` has `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` and no variable outside the allowed set; a known name is accepted only when its answer holds `Not logged in` and no `not found`; a child that does not exit by the deadline is killed with its process group and reported · verify: `cargo test --test base_agents_discovery_record the_child_claude a_known_name_is_accepted a_child_is_killed`

## Scope

- write: `tests/base_agents_discovery_record.rs` and this SPEC folder
- read-only: `src/agent_install.rs`, `src/host_home.rs`, `skills/moai`, `agents`, `tests/common/mod.rs`, `tests/base_agents_nested.rs`
- forbidden: the real home folder, the user's Claude config folder, the Orca profile, any credential, Codex and OMP folders, any card number in a file name, SPEC id, branch or commit

## Plan

1. Write the SPEC and the test; no product code changes, so there is no red step for the install, and the control of REQ-4 and a name check that must fail for an unknown name carry the proof that the test can fail.
2. Run the test with the real `claude` 2.1.330 on the built binary, record the list lines and each name's answer in the Result.
3. Run the local checks of the change and hand the head over with the receipt.

## Risks

- `claude --agent <unknown name>` is an undocumented way to list definitions; if a later Claude Code prints another text, the test fails on the missing `Available agents:` line instead of passing silently.
- A logged-in environment variable (`ANTHROPIC_API_KEY`) could turn a name check into a model call; the environment is cleared, so none reaches the child.
- The built-in agent names of Claude Code may change; the test looks for the names this tool owns and for the names it created, never for the full list.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "Invokes" is recorded as the name check reaching the login step, not as a model answer | A temporary config folder has no credentials and credentials are never copied (HOLD `secret`) | Add a model-answered call when the user supplies a throwaway key for tests |
| D-2 | The Orca shape is a profile folder with an `agents` link only | The leader's read of 2026-11-23 shows the profile links agents to `~/.claude`; other entries do not change what is found here | Add the other links to the fixture when a card needs them |
| D-3 | No product change; the checked behaviour is the one already on the tip | Both shapes pass on the tip (measured), so a change would have nothing to fix | none |
| D-4 | A known name must end at `Not logged in`, the child gets the traffic switch and a deadline | Audit notes N1 and N2 of the card: a leaked credential would turn `-p hi` into a billed model call that still passed, and an unauthenticated run tried DNS lookups | Drop the three guards; the cost is a silent paid call |

## Result

Added `tests/base_agents_discovery_record.rs` and this SPEC; no product code changed (D-3). The test ran with the real `claude` 2.1.330 (`/Users/goos/.local/bin/claude`) and the binary built from this tree, on temporary folders with a cleared environment: `cargo nextest run --offline --locked -E 'binary(base_agents_discovery_record)'` — 4 run, 4 passed (`native_layout` 6.5 s, `orca_profile_layout` 7.1 s, the control 0.5 s, the skip helper 0.02 s). Seconds of the first two show the real `claude` ran; nothing was skipped.

- AC-1: met — in a native temporary config folder the list printed by `claude --agent no-such-agent-for-the-list -p hi` held `moai-audit`, `moai-docs`, `moai-impl`, `moai-plan`, `moai-research`, the personal `my-own` and the project role `prj-probe`, each once, next to Claude Code's own `claude`, `Explore`, `general-purpose`, `Plan`, `statusline-setup`; the five base files sit only under `agents/moai/` and none directly under `agents/`.
- AC-2: met — with `CLAUDE_CONFIG_DIR` naming a profile folder whose `agents` entry links to the canonical `agents` folder, the install (run with that variable) wrote through the link into `agents/moai/`, left the profile holding only its `agents` link, and the list read through the profile equals the list read natively.
- AC-3: met — in both shapes `claude --agent <name> -p hi` for each of the six names did not answer "not found"; it stopped at `Not logged in · Please run /login` (no credential in a temporary config folder, none copied). A model-answered call is `not_run` by design (D-1). The unknown name is answered "not found" in the same run, so the detector can fire.
- AC-4: met — with the profile's `agents` link pointing at another folder the preview of the claude host is `blocked` (`… is a link to a folder other than …`) and the list through that profile holds none of the five base agents, only the project role.
- AC-5: met — `no_claude_on_the_path_is_a_skip` passed for an unset, an empty and a `claude`-less `PATH`; each of the three other cases returns with the skip line in that situation.
- Mutant: renaming the installed `moai-plan.md`'s `name` to `moai-plan-x` after the install made `native_layout` fail with `moai-plan listed 0 times in [… "moai-plan-x" …]`; the edit was removed and the file re-run green.
- Not run: Windows and Linux; a model-answered call of any agent; the real home, `~/.claude` and the Orca profile were never read or written.
- Repair after the audit (tests only, D-4): the child `claude` is built by `claude_command` with `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`; `run_with_deadline` kills a child that is still running after 60 s with its process group; every known-name run must hold `Not logged in` (`login_stop_error`). The three new unit tests need no `claude`, so they also run where it is absent. The blocked-profile case now also applies the blocked plan's digest and requires `unchanged` with nothing written; the Orca case also requires that no flat file exists in the canonical folder.
- Mutants of the repair: dropping the traffic variable, the `Not logged in` check and the deadline check each fail their own test; planting a flat file after the Orca install fails `orca_profile_layout`. Two product mutants of `apply` for the blocked-profile case (the preflight filter, and the early return of the folder check) both survived: the plan returns blocked with no files at several steps, so that apply assertion is a regression pin and not shown to bite.
