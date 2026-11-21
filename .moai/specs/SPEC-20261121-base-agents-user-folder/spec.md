---
id: SPEC-20261121-base-agents-user-folder
title: "The base agents live only in the user's agents/moai folder, a project's agents/moai folder holds only its own harness roles, and flat copies from earlier versions are moved safely"
status: draft
revision: 1
created: 2026-11-21
---

# The base agents live only in the user's agents/moai folder, a project's agents/moai folder holds only its own harness roles, and flat copies from earlier versions are moved safely

## Intent

The user decided on 2026-11-21 (Claude Code only; GLM runs through the Claude Code launcher) that the five base agents (`moai-plan`, `moai-research`, `moai-impl`, `moai-audit`, `moai-docs`) are installed only under `~/.claude/agents/moai/`, and that a project's `.claude/agents/moai/` holds only that project's harness roles (`prj-*`). Today both kinds sit flat in an `agents/` folder, next to the user's own agents.

What the official sub-agent page says (https://code.claude.com/docs/en/sub-agents, read 2026-11-21 by the research agent and checked with the installed Claude Code 2.1.330):

- "Claude Code scans `.claude/agents/` and `~/.claude/agents/` recursively", so a nested `moai/` folder is found. A subfolder does not change the name: identity comes only from the `name` frontmatter field, and the page asks to keep `name` values unique across subfolders.
- Priority when two definitions share a name: managed settings, the `--agents` flag, `.claude/agents/` of the project, `~/.claude/agents/` of the user, then plugins. A project definition therefore hides a user definition of the same name; two files of one scope with the same name load only one of them.
- No command prints the definitions without a model run. A measured substitute: `claude --agent <unknown name> -p hi` exits 1 before any model call and prints `Available agents: …`. With a temporary config folder holding `agents/moai/moai-plan.md` and a project holding `.claude/agents/moai/prj-probe.md`, the list held `moai-plan` and `prj-probe`.

Consequences for the design:

- The agent names stay as they are, so nothing that calls `moai-plan` changes; only the folder changes.
- A flat file from an earlier version must not stay beside its nested copy: two user-scope files with one name make Claude Code load only one of them. So the nested copy and the removal of the flat file go together, and a flat file the user edited (or one that was never ours) is left alone and its nested copy is not written.
- Project role ids already start with `prj-`, so a project role can never shadow a base agent; a test pins that no `moai-*` file is planned under a project.

## Non-goals

- Codex and OMP agent folders (`~/.codex/agents`, `~/.omp/agent/agents`): unchanged here; their removal is another change.
- Personal agents (any other file under `agents/` or `agents/moai/`): never read, listed or touched.
- The real HOME: every test uses a temporary HOME and a temporary project.
- Plugin agents and managed settings.

## Requirements

- REQ-1: The home agent install shall plan and write the five base agents for Claude Code only inside the `moai` folder of the user agent folder, so that no base agent file is planned at the top level of `agents/`.
- REQ-2: The home agent install shall keep each base agent's `name` field as it is, so that the five names stay `moai-plan`, `moai-research`, `moai-impl`, `moai-audit` and `moai-docs`.
- REQ-3: When a flat base agent file from an earlier install still holds the bytes this tool recorded for it, the apply shall write its nested copy first and then remove the flat file with its bytes kept in the receipt.
- REQ-4: When a flat base agent file was edited or was never installed by this tool, the plan shall list the file as a conflict and leave the flat file and the nested path of that agent untouched. · failure: none — the file stays as the user wrote it
- REQ-5: When the apply of a move is reverted, the revert shall put the flat file back with its recorded bytes and remove the nested copy that the apply created.
- REQ-6: When the install runs again with nothing changed, the apply shall write nothing and report `unchanged`. · failure: none — nothing is written
- REQ-11 (REQ-3): IF a flat base agent file changed between the preview and the apply, THEN the apply shall stop with the re-preview error.
- REQ-12 (REQ-5): IF a file changed since the apply that is reverted, THEN the revert shall skip that file and report it.
- REQ-13: When the plan finds a flat base agent file it can move, the preview shall show the removal of the flat file. · failure: none — a preview only reports
- REQ-7: The home agent install shall read, list and touch only the file names the package renders, so that every other file under `agents/` and `agents/moai/` stays as it is.
- REQ-8: The project generation shall write a native role of Claude Code at `.claude/agents/moai/<role id>.md` and plan no file under `.claude/agents/` that carries a base agent name.
- REQ-9: When a project holds a flat role file that a generation recorded and nobody edited, the project plan shall move it to the nested path with its bytes kept in the receipt.
- REQ-14 (REQ-9): IF a flat role file was edited after the generation, THEN the project plan shall list it as a conflict without writing the nested copy.
- REQ-10: The tree shall carry a recorded check that Claude Code finds both kinds: with a temporary home and project, the available-agents list of `claude --agent <unknown name>` names the five base agents and the project roles. · failure: none — the check is skipped, and says so, when `claude` is not on the path

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-7): a plan and an apply on a temporary home put exactly `moai-audit.md`, `moai-docs.md`, `moai-impl.md`, `moai-plan.md` and `moai-research.md` under `.claude/agents/moai/`, none directly under `.claude/agents/`, each starting with `name: moai-<role>`, and a personal `agents/mine.md` and `agents/moai/mine.md` keep their bytes · verify: `cargo test --test base_agents_nested --test agents_home_cli`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6, REQ-11, REQ-12, REQ-13): a flat set written by the earlier layout moves (nested copy first, flat file removed, receipt holds the bytes), an edited flat file and a foreign flat file stay with no nested copy and a conflict line, the revert restores the flat file and removes the nested one, and the second apply reports `unchanged` · verify: `cargo test --test base_agents_nested`
- AC-3 (REQ-8, REQ-9, REQ-14): a project generation writes `prj-*` roles at `.claude/agents/moai/`, plans no base agent name, moves a recorded unedited flat role and keeps an edited one as a conflict · verify: `cargo test --test project_agents_nested`
- AC-4 (REQ-10): the available-agents list of Claude Code in a temporary home and project holds the five base agents and the project roles · verify: `cargo test --test base_agents_nested claude_code_finds_both_kinds -- --nocapture`
- AC-5 (REQ-1, REQ-3): the tests and fixtures that name the flat base-agent path (`agents_home_cli`, `update_user_layer_cli`, `update_tux_golden` with its fixtures, `install_script_test.py`) follow the new path and pass · verify: `moai-cli check affected`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11, REQ-12, REQ-13, REQ-14): the whole library and the binaries build and pass, the SPEC checks clean · verify: `cargo test --lib && cargo test --bins && moai-cli spec check SPEC-20261121-base-agents-user-folder`

## Scope

- write: `src/agent_install.rs`, `src/update.rs` (the agent summary and the write test only), `src/generation.rs` (the native role path and its move only), the tests named in the acceptance lines, `tests/fixtures/update_tux/`, new `tests/base_agents_nested.rs` and `tests/project_agents_nested.rs`, and this SPEC folder
- files held by other changes at the time of writing (one-line follows by pair notes): `tests/update_user_layer_cli.rs`, `tests/update_tux_golden.rs`, `tests/install_script_test.py`
- read-only: `src/agents.rs`, `src/agents_local_merge.rs`, `src/ops/agents_*`, `src/agent_defs.rs`, `src/omp_agents.rs`, `src/host_home.rs`
- forbidden: the Go repository, the real HOME, any file under the user's agents, Codex or OMP folders in the home, any card number in a file name, SPEC id, branch or commit

## Plan

Tests first (RED): `tests/base_agents_nested.rs` runs the install functions on a temporary home for the nested layout, the move, the kept edited and foreign flat files, the revert and the rerun, and runs the real `claude` for the discovery list; `tests/project_agents_nested.rs` does the same for the project roles. On the tip they fail for the right reason (files land flat).

Order, one commit group each with its full target list run before the next: (1) the home folder: `host_rel` for Claude Code becomes `agents/moai`, the plan reads the flat names for the move, the apply and the revert handle the legacy removal (a `legacy` field on a file plan, left out of the JSON when empty so unaffected fixtures stay as they are); (2) the follows of the tests and fixtures that name the flat path; (3) the project role path and its move in `src/generation.rs`; (4) the recorded discovery check.

## Risks

- A flat file that this tool wrote and the user edited must never be removed or overwritten; the move is decided by digest equality with the recorded baseline only.
- The nested copy and the flat removal must not be split by a failure: the nested file is written first and the receipt entry of the removal is written after the bytes are kept, so a stop leaves a duplicate that the next preview shows as a move again, never a missing agent.
- `host_folders_plain` checks every level from the host folder down; `agents/moai` joins the checked levels, so a link named `moai` blocks the host instead of being followed.
- The update summary and the TUX golden count file actions; a removal without a nested write must still count as a write (`agents_write_anything`).
- `tests/update_user_layer_cli.rs`, `tests/update_tux_golden.rs` and `tests/install_script_test.py` are held by other changes: the follow is one path line each, agreed by pair notes.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only the Claude Code folder is nested; Codex and OMP stay flat | Claude Code only decision of 2026-11-21; the other hosts leave in their own change | Nest them later in one `host_rel` line |
| D-2 | A flat file that was edited or is foreign blocks its nested copy | Two same-name files in one scope load only one (official page); keeping the user's file wins over installing ours | The user removes or renames the flat file; the next update installs the nested copy |
| D-3 | The discovery check uses `claude --agent <unknown name>` and its available-agents list | No documented command lists definitions without a model run; this one exits before the model call (measured on 2.1.330) | Replace with a documented listing when one exists |
| D-4 | A project role keeps its `prj-` prefix | The prefix already forbids a base name; the nested folder is for organizing, not naming | none |

## Result

(to be written at the end of the change)
