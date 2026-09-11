# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.2.2 — applies the lead's consumer-scope ruling X1-X4 (2026-09-11). 0.2.1 was committed as `caa601d7c`. An earlier 0.2.2 attempt stopped on a session limit and its partial edits were preserved as WIP commit `ea09ca650`; this revision completes it from that commit.
- **Revision note (0.2.2)**:
  - In scope, local and template, line-level: X1 `workflows/sync.md` usage line (local 95 / template 85), X2 `argument-hint` in `.claude/commands/moai/sync.md:3` and template `sync.md.tmpl:3`, X3 `delivery.md:404` "Auto-Merge PR (/moai sync --merge)" (same line number in both copies). All three verified by reading the lines on this tree.
  - X2 is a command source, so REQ-GDP-014 now also binds `make commands-emit` then `make commands-emit-check` (exit 0) with any regenerated published skill in the same commit as the source edit; new AC-GDP-030 judges both outcomes (A no change / B change) and carries a control showing the check can fail.
  - X4 (docs-site four locales) is out of scope; the follow-up docs card and its measured inputs are recorded in spec.md §D.
  - Kept from the WIP commit after re-measurement: ten-file scope set, nine AC-GDP-026/027 fragments and their base controls, AC-GDP-013 parity rows for the command pair and published skill, AC-GDP-015/016/025 file lists, AC-GDP-030 body.
  - Corrected: AC-GDP-025 registry detector `manager-git\.md` → `manager-[g]it\.md` (the worktree guard refuses the literal word inside a pattern; the uncorrected command could not have run at run-phase). spec.md §C.5 content moved into §D (the ruling asks for it there). Base measurement HEAD updated to `ea09ca650`.
  - Added: empty-set guards on AC-GDP-015 (template diff must be non-empty), AC-GDP-016 (the always-loaded-rule commit must exist), AC-GDP-025 (scope diff must be non-empty; mutant for the "others" registry detector), AC-GDP-030 (tracked-path control, source-commit existence, mechanical same-commit subset check, published file non-empty).
- **Tier**: M kept per lead ruling (judged counts decide; file-count guidance is reference). Judged REQ 12 and AC 16 are within 16/16; affected files 21 (+2 if AC-GDP-030 outcome B). **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: HEAD `ea09ca650` (parent `caa601d7c`; both change only SPEC files). `git diff --stat b412f8a33 ea09ca650 -- <scope files L·T, manager-git.toml, published moai-sync SKILL.md L·T, internal/template/commandemit, internal/template/agentemit, zone-registry.md L·T, docs-site/content, Makefile>` → no output; same for the eight guarding test files; `git diff --stat b412f8a33 -- internal/template/templates/` → no output. Plan-time counts on these copies are base counts.
- **Audit count**: reduced iteration 1 FAIL 0.71 (`.moai/reports/t622/plan-audit-reduced-iter1.md`). No audit file for 0.2.1 exists in `.moai/reports/t622/` (listing at 0.2.2 authoring). This revision is for reduced iteration 2. Full-scope iterations before the split: FAIL 0.67 and FAIL 0.75.
- **Pre-write self-checks**: SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` on `SPEC-GIT-DELIVERY-PROCEDURE-001` → `PASS` (Bash, 0.2.2). Frontmatter 12 canonical fields present; `phase: "v3.2.0 target"`.
- **Counts**: REQ 26 numbers — judged 12 (001-006, 013-015, 024-026), moved 12, withdrawn 2. AC 30 numbers — judged 16 (001-006, 013-016, 025-030), moved 12, withdrawn 2.
- **commands-emit grounding (read, not executed — no make/go test at plan time)**: `Makefile:51-52` `commands-emit` = `COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission`; `Makefile:58-60` `commands-emit-check` = same test with the env var scrubbed, read-only. `golden_test.go:28` `templatesDir = "../templates"`; `:57-74` update branch writes only template-tree published paths; `:94-106` `TestCommandSourcesUnmodified` compares hashes before and after emission within one run, so a source edit alone does not trip it. `commandemit.go:49-53` roots `.claude/commands/moai` → `.agents/skills`. `loader.go:4-6` `argument-hint` and `allowed-tools` are "NOT carried into the published skill"; `emit.go:122-130` `renderSkill` writes header, `name`, `description`, body. Published copies: local and template both 7 lines, `diff` exit 0, case-insensitive `merge` count 0 and 0, both tracked (`git ls-files` lists all four of the two published copies and two command sources). Expected outcome A; judged by command in run-phase.
- **Re-measured controls for new or changed criteria (this tree = base)**:
  - AC-GDP-026 fragments (7)(8)(9), template and local copies each: lines 1 / 1 / 24; `--auto-merge` count 0 / 0 / 0; (ii) violations at fragment lines 1 (usage), 1 (hint), 9 (next-steps = file line 404). Local and template next-steps fragments `diff` exit 0. X1-X3 mutant fixture (6 lines) → (ii) hits lines 1, 3, 5; `--auto-merge` lines 1, 2, 4, 6.
  - AC-GDP-027 on fragments (7)(8)(9): `--no-merge` count 0 / 0 / 0 (both copies).
  - AC-GDP-013 command source `argument-hint` local vs template `diff` exit 0; command source copies `diff` → `2c2` only.
  - AC-GDP-005 command sources: `gh pr merge[^|]*--squash` count 0 and 0.
  - AC-GDP-025: `[ZONE:` lines in command sources 0 and 0; registry entries naming agent-common-protocol.md 13 and 13 (positive control); registry "others" detector (with `manager-[g]it`) 0 and 0; mutant fixture (3 lines: commands path, workflows/sync.md path, agent-common-protocol path) → 2.
  - AC-GDP-030 same-commit subset check (`/usr/bin/grep -v -x -F -f <src> <artifact>`): artifact SHA absent from src → exit 0 (FAIL detected); artifact SHA present → exit 1; empty src file → exit 0 with 1 line (FAIL — an empty source list cannot pass).
  - AC-GDP-015/016/025 new empty-set guards: `git diff --stat b412f8a33 -- internal/template/templates/` has no output on this tree, so `test -s` on the template or scope diff is exit 1 at base — the guard blocks a pass over an empty diff.
- **Docs-site X4 inputs (re-measured)**: `git grep -c -e --merge -e --no-merge b412f8a33 -- <moai-sync.md en/ko/ja/zh, what-is-moai-adk.md en/ko/ja/zh>` → moai-sync.md en 5, ko 4, ja 6, zh 9; what-is-moai-adk.md 1 each. Line numbers from `git grep -n` with the same paths plus `-e --auto-merge` (no `--auto-merge` hits) are recorded in spec.md §D.
- **Frozen check for the command files**: `.claude/commands/moai/sync.md` and `internal/template/templates/.claude/commands/moai/sync.md.tmpl` carry 0 `[ZONE:` lines and 0 zone-registry entries (local and template registry); edit site line 3 is outside any Frozen block.
- **Carried from 0.2.1 (unchanged base)**: AC-GDP-001 paragraphs 1 / autofail 1 / listgroup 0; AC-GDP-002 block 9 lines, fetch-without-rev-list 1, standalone rev-list at block line 5, joined 0, rev-list 1, session command 1, table rows 8; AC-GDP-004 squash 2, merge_method 0; AC-GDP-005 six `--squash` lines (manager-git 32/114, delivery 343/355, toml 26/108); AC-GDP-006 delivery Step 3.4 hits at section lines 8 and 20, doc-execution subsection hit at line 8; AC-GDP-014 `.toml` sections 11 and 9 lines diff exit 0; AC-GDP-026 fragments (1)-(6) lines 1/3/4/6/1/52 and violations 7; AC-GDP-027 delivery `--no-merge` at 8 and 19; AC-GDP-028/029 base (a) 0 and (b) 0 with their mutant fixtures; AC-GDP-015 fixture 5 letter / 2 digit tokens; AC-GDP-025 four clause counts 1 per copy.
- **plan_status**: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
