---
id: SPEC-20261120-shipped-workflows-commands
title: "Shipped workflows name moai commands only — the repo-only tools scripts leave the shipped text"
status: draft
revision: 1
created: 2026-11-21
---

## Intent

Every user installs moai and gets `skills/moai` and `agents`; a shipped workflow that says
"in the moai-adk repository run `tools/X`" gives other users nothing (user instruction
2026-11-20T20:1xZ: "로컬 개발 도구와 오픈소스 기능을 혼동하지 마라"). This card slices the fix
into three moves per row of the survey below: (a) a step that any user's project needs
becomes a moai command (Rust with tests) and the shipped text names only the command;
(b) a moai-adk-only tool leaves the shipped skills and agents and moves to this
repository's own places (`AGENTS.local.md`, `REVIEW.md`, `docs/`); (c) a step that is
already a moai command keeps the shipped text naming only the command. 애매한 행은 (b)로 두고
리더에게 노트한다(leader decision, card note 4548).

Survey (re-run at the integration tip `bcfb1bc5b7`, `git grep -nE 'tools/[^ ]+\.(sh|py)' --
skills agents` = 21 lines in 6 shipped files): `verify.md` ×7 (`tools/smoke/run.sh`:20,
`tools/source_manifest.py`:35·80, `tools/ci_scope.py`:57·78, `tools/self_check.sh`:17·49),
`source-manifest.json` ×5 (`tools/legacy_verdicts.py`:33, `tools/build_folder.py`:125,
`tools/self_check.sh`:269, `tools/gen_goal_block.py`:425·434, `tools/repo_scans.py`:434),
`build-folder.md` ×4 (`tools/build_folder.py` env:10·clean:13·remove:14·env:15),
`mode-agent.md` ×2 (`tools/self_check.sh`:63, `tools/local_check.sh`:64),
`self-check.md` ×2 (`tools/self_check.sh`:5, `tools/source_manifest.py`:23),
`rule-history.md` ×1 (`tools/local_check.sh`:118). Slice 1 re-verifies this survey, writes
the per-row plan (a/b/c) into this SPEC, and cleans the shipped text of the files this
card holds today; the rows whose files are another card's reservation are done in later
slices right after those cards return (the reservations are listed under Scope).

## Non-goals

- Writing the new Rust commands of move (a): each command is its own later slice behind
  this SPEC's per-row plan — slice 1 only rewrites the shipped text of the rows that need
  no new command and records the plan for the rest.
- Changing `tools/*.sh|*.py` themselves: the repo-only scripts stay in `tools/` for this
  repository's own development; only the shipped text stops pointing users at them.
- The scan test that pins criterion 1 (no shipped file names a `tools/` script): it lands
  in the last slice, after every shipped file has been cleaned — before that, the survey
  table in this SPEC is the working list.
- `tools/gen_goal_block.py` and `tools/repo_scans.py` mentions inside the manifest's
  historical `reason` prose: those sentences record what happened (repo_scans is already
  `moai-cli repo-scans` since #2708), and rewriting history lines is not this card's work.

## Requirements

- REQ-1: While slice 1 finishes, the shipped text of `build-folder.md` and `self-check.md` shall name no `tools/` script — the env and cleanup steps name `moai lane-env` (and the planned end-of-task cleanup option), the removal step names `moai optimize-plan` and `moai optimize-apply`, the manifest step moves to this repository's own docs, and the generic rule stays for every user. · failure: none — pinned by AC-1.
- REQ-2: While slice 1 finishes, this repository's own places shall carry what the shipped text dropped — `docs/moai-adk-tools.md` holds the moai-adk-only tool procedures (the repository's self-check runner, the skill-package manifest commands, the smoke harness, the ci_scope classifier, the local check set and the tool notes), so every moved row keeps its procedure in this repository. (`AGENTS.local.md` and the reserved workflow files follow in later slices.) · failure: none — pinned by AC-2.
- REQ-3: While the survey is written into this SPEC, every one of the 21 lines shall carry its move (a, b or c), its target command or destination, and its slice — with the rows whose files are another card's reservation marked for later slices, so no row is lost. · failure: none — pinned by AC-3.

## Acceptance

- AC-1 (REQ-1): `grep -E 'tools/[^ ]+\.(sh|py)'` over `build-folder.md` and `self-check.md` finds nothing, and the rewritten steps name `moai lane-env`, `moai optimize-plan` and `moai optimize-apply` where the tool was named before.
  verify: `grep -nE 'tools/[^ ]+\.(sh|py)' skills/moai/references/internal/build-folder.md skills/moai/references/internal/self-check.md` (no output)
- AC-2 (REQ-2): `docs/moai-adk-tools.md` exists with the moved procedures — the self-check runner, the manifest write/check commands, the smoke areas, the ci_scope classifier notes and the local check set — each named with the shipped file it moved from.
  verify: `grep -n 'self_check.sh\|source_manifest\|smoke/run.sh\|ci_scope.py\|local_check.sh' docs/moai-adk-tools.md` (the moved tools are named in this repository's own docs, not in the shipped text)
- AC-3 (REQ-3): this SPEC's survey table lists all 21 lines with move (a/b/c) and the later-slice marks; the file set matches `git grep` at the slice's tip.
  verify: `moai spec check SPEC-20261120-shipped-workflows-commands`

## Scope

- write: `skills/moai/references/internal/build-folder.md`, `skills/moai/references/internal/self-check.md`, `docs/moai-adk-tools.md`, this SPEC folder (`REVIEW.md` returned to scope-less state: the grep found no `tools/` reference in it to move)
- reserved by other cards (later slices take them when returned): `skills/moai/workflows/verify.md` and `skills/moai/references/source-manifest.json` (C-151, place 10), `skills/moai/workflows/mode-agent.md` and `AGENTS.local.md` (C-204), `skills/moai/references/rule-history.md` (C-214)
- read-only: `tools/*.sh`, `tools/*.py`, `src/ops/` (the commands the rows name or will name)
- forbidden: deleting or editing anything under `tools/`, changing the shipped text's Korean voice, a shipped file naming a `tools/` script after its slice lands

## Plan

1. Commit this SPEC with the survey table (REQ-3).
2. Slice 1 text cleanup, this card's own files: `build-folder.md` (env → `moai lane-env`; clean → the lane-env cleanup option named in the later slice, with the relative-path rule kept; remove → `moai optimize-plan`/`moai optimize-apply`) and `self-check.md` (the manifest step moves to `docs/moai-adk-tools.md`; the generic rule stays).
3. `docs/moai-adk-tools.md` collects the moved moai-adk-only procedures; `REVIEW.md` names it where the review flow used a tool script.
4. Later slices, in file order after the reservations return: `verify.md` (b+c rows), `mode-agent.md` (c + the moved rows), `source-manifest.json` (b rows — only where the manifest itself is edited by that slice's work), `rule-history.md` (b), `AGENTS.local.md` (b) — then the scan test of criterion 1, then the (a) commands one slice each.
5. Neighbours: the skills-related test modules and the manifest test keep passing (the manifest's sha256 entries change only when a shipped file's bytes change — slice 1 changes shipped bytes, so the manifest entries for the two edited files are regenerated in the same commit).

## Risks

- `source-manifest.json` is C-151's reservation while place 10 works: slice 1 changes the bytes of `build-folder.md` and `self-check.md`, so their manifest sha256 entries must be regenerated — that is a manifest edit inside C-151's reservation. The slice therefore lands its text cleanup and the manifest regeneration in one commit only after C-151 returns, or hands the regenerated entries to C-151 as a note (leader decides; this SPEC records the dependency here).
- The verify.md rows (7 of the 21) are the largest block and sit behind the same reservation.
- A shipped sentence that names both the generic rule and the repo-only tool must keep the generic half — the moved text must not take the rule away from users.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | Slice 1 cleans only the files this card holds today (build-folder.md, self-check.md, REVIEW.md, the new docs file) and records the rest as later slices | Five of the six shipped files are other cards' reservations this round (C-151, C-204, C-214) | Take the whole set at once when all return |
| D2 | `build-folder.md`'s steps name existing commands (`moai lane-env`, `moai optimize-plan`, `moai optimize-apply`) before any new command exists | The three steps are already covered by commands; no new Rust work is needed for this row | New commands where the mapping is missing |
| D3 | The manifest sha256 regeneration rides the same commit as the shipped-text change | The manifest test fails otherwise; C-151's reservation decides the timing, not the shape | — |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.
