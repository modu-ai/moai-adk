---
id: SPEC-20261120-project-harness-namespace
title: "project generation puts its agents and skills in a moai namespace, apart from the user's own, and moves earlier output into it through a preview"
status: draft
revision: 1
created: 2026-11-20
---

# project generation puts its agents and skills in a moai namespace, apart from the user's own, and moves earlier output into it through a preview

## Intent

`/moai project` writes the agents and skills it generates next to the files a user makes by hand: `.claude/agents/<role>.md`, `.codex/agents/<role>.toml`, `.claude/skills/<id>/SKILL.md` and `.agents/skills/<id>/SKILL.md`. A user who adds their own agent or skill cannot tell from the folder which files MoAI owns, and a later regenerate or delete has no folder to stay inside. The user asked for a split: generated agents in a `moai` folder, generated skills in `moai-*` folders, so that the user's own files and ours are managed apart.

This change moves the four generated paths into that namespace, once the official documents of both hosts were read for whether they load the new places (`research.md`), and moves what an earlier run wrote through the existing plan and apply preview. The generation record already lists the files MoAI wrote with the digest of their bytes; that record is the proof that a file is ours. A file without a record entry is the user's and is never touched.

## Requirements

- REQ-1: The `generation-plan` operation shall plan each Claude Code agent file of a request as `.claude/agents/moai/<role>.md` with the front matter `name` equal to the role id.
- REQ-2: The `generation-plan` operation shall plan each Codex CLI agent file of a request as `.codex/agents/moai/<role>.toml` with the `name` key equal to the role id.
- REQ-3: The `generation-plan` operation shall plan each project skill and each skill projection of a request as `.claude/skills/moai-<id>/SKILL.md` for Claude Code and `.agents/skills/moai-<id>/SKILL.md` for Codex CLI with the front matter `name` equal to the folder name.
- REQ-4 (REQ-3): IF a skill id would make the folder name longer than sixty-four characters, THEN the `generation-plan` operation shall refuse the request with the name of the field and the limit.
- REQ-5: WHEN a file that the generation record owns lies at the former path of a requested agent or skill and its bytes equal the recorded ones, the `generation-plan` operation shall plan the removal of the former path together with the creation of the new path. · failure: none
- REQ-6 (REQ-5): IF a file that the generation record owns at a former path was edited after it was generated, THEN the `generation-plan` operation shall plan the former path and the new path as conflict and leave the edited file where it is.
- REQ-7: The `generation-plan` operation shall plan no change to a file at a former path that the generation record does not own.
- REQ-8 (REQ-7): IF a file the generation record does not own lies at the former path of a requested agent or skill, THEN the `generation-plan` operation shall plan the new path as conflict and say why, because Claude Code loads only one of two agents that share a name.
- REQ-9: The `generation-apply` operation shall refuse to remove a former path whose bytes differ from the plan.
- REQ-10: The project route shall tell the agent that the agents and skills it generates are written inside the `moai` folder of the agents folder and the `moai-` skill folders, apart from the files the user made.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a plan for one agent per host and one project skill and one skill projection names exactly the namespaced paths, writes the front matter and the TOML keys named in the requirements, and plans nothing at the former paths · verify: `cargo test --locked --offline --test project_harness_namespace fresh_`
- AC-2 (REQ-4): an id of sixty and of fifty-nine characters, the first refused with the field and the limit and the second planned · verify: `cargo test --locked --offline --test project_harness_namespace folder_name_`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): in temporary projects, an earlier run's owned and untouched files move through plan and apply with their bytes kept; an edited one stays and both paths are conflict; a file of the user's at a former path is never changed and the new path is conflict with the reason; a former path edited after the plan is not removed · verify: `cargo test --locked --offline --test project_harness_namespace moved_`
- AC-4 (REQ-1, REQ-2, REQ-3): `research.md` of this folder holds the official quotes and the read times that the paths rest on, and the test reads them · verify: `cargo test --locked --offline --test project_harness_namespace research_`
- AC-5 (REQ-10): step 6 of `skills/moai/workflows/project.md` names the namespace folders, and the scan and reader tests of the project text pass · verify: `cargo test --locked --offline --test project_harness_namespace text_`

## Scope

- write: `.moai/specs/SPEC-20261120-project-harness-namespace/**`
- write: `src/generation.rs`, `tests/project_harness_namespace.rs`
- write (widened when the file is free and its other holder has merged): `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`
- write (widened with the tests that pin the former paths, listed when found): `tests/project_skills_generation.rs` and the readers of `generation-plan` named in the Result

## Decisions

- D-1: Claude Code agents go in `.claude/agents/moai/` because the official page says the folder is scanned recursively and that the subfolder does not change the identity (`research.md`). Claude Code skills stay one folder deep and take the prefix `moai-`, because the official skills page documents `.claude/skills/<skill-name>/SKILL.md` only and says nothing of a grouping folder.
- D-2: Codex CLI agents go in `.codex/agents/moai/`. The Codex documents list files directly in `.codex/agents/` and are silent on subfolders; the official source reads the folder recursively (`agent-roles/src/discovery.rs`, commit `a1afa122738382adabe2dd1ccfc510a70ed381f0` of main, `research.md`). The ground is the source, not a document. The leader may choose the fallback, a `moai-` file-name prefix in `.codex/agents/` (one function changes); the test pins the choice in one place. The earlier Go generation also wrote `.codex/agents/moai/` (`src/go_state.rs`).
- D-3: Codex CLI skills take the same prefix as Claude Code skills (`.agents/skills/moai-<id>/`), one folder deep, although the source reads nested folders: the document lists one level, and one rule for both hosts is simpler to explain and to test.
- D-4: The skill front matter `name` is the folder name (`moai-<id>`), so a skill of the user's with the plain id never shares a name with ours. An agent keeps `name: prj-<name>`: the prefix `prj-` already keeps it apart from the user-layer `moai-*` agents, and the folder is the namespace.
- D-5: Two agent files that share a `name` under one `.claude/agents/` are loaded by read order, with one picked and no documented rule, so a move never leaves a copy of ours beside the old path and a file of the user's that sits at the old path of a requested agent turns the new path into a conflict. A user file that has another file name but declares the same `name` is not found by path and is not detected (not done here).
- D-6: Only files that a request names move, and the record is the proof of ownership. An owned agent or skill projection at a former path that the new request no longer names stays where it is and is reported in a warning; owned project skills no longer requested are removed at the former path by the rule that already exists.
- D-7: `skills/moai/workflows/project.md` step 6 is also rewritten by the project-harness research card on the same physical line. The text edit of this card waits for that card's text to be merged and for the project-text card to return the file, and then changes only the two sentences about the agent and skill folders.
- D-8: The user-level agents (`~/.claude/agents`, `~/.codex/agents`) and the project copy of the `moai` skill are other layers and stay as they are.

## Plan

1. This card, now: this SPEC and `research.md` (Claude Code and Codex CLI, official quotes with read times).
2. Tests first: `tests/project_harness_namespace.rs` (red: the paths are the old ones); then `src/generation.rs` (path functions, the skill name, the folder length bound, the move plan); then the tests that pin the former paths; then the text of step 6 when `project.md` is free.
3. Whole-area run (every reader of `generation-plan` and `project_skill_*`), clippy, rustfmt, spec check, audit, merge.

## Result

- AC-1: not yet — nothing written.
- AC-2: not yet — nothing written.
- AC-3: not yet — nothing written.
- AC-4: not yet — `research.md` is written and the test that reads it is not.
- AC-5: not yet — `project.md` is held by another card.
