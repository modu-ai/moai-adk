# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.2.4 — pre-kickoff application of optional findings N1-N3 from reduced plan-audit iteration 3 (`.moai/reports/t622/plan-audit-reduced-iter3.md`) per operator decision via lead: Implementation Kickoff approved on condition these are applied first; no re-audit follows; requirements, decisions, scope, judged REQ 12 and judged AC 16 unchanged. Revision 0.2.3 — targeted fix of reduced plan-audit iteration 2 findings D1-D8 (`.moai/reports/t622/plan-audit-reduced-iter2.md`, FAIL 0.75) per operator decision via lead: detector strings and wording only; requirements, decisions, scope, judged REQ 12 and judged AC 16 unchanged. Revision 0.2.2 — applies the lead's consumer-scope ruling X1-X4 (2026-09-11). 0.2.1 was committed as `caa601d7c`. An earlier 0.2.2 attempt stopped on a session limit and its partial edits were preserved as WIP commit `ea09ca650`; this revision completes it from that commit.
- **Revision note (0.2.4)** — acceptance.md only (plus this note and the spec.md HISTORY row):
  - N1: the command convention now prescribes literal values for `BASE`/`E`/`T` (`$BASE`/`$E`/`$T` in judge commands are notation to substitute) and forbids the `BASE=…; E=…; T=…; <command>` prefix form for git commands and commands naming `manager-git.md` (guard refusal observed in audit iteration 3: T2, T7, T4, L2). No judge command was rewritten — the audit names no judge unrunnable when its values are substituted literally (T3, T8, L1, V2-V5 ran).
  - N2: `test -e <file>` (expected exit 0; exit 1 = judge never ran = not PASS) placed before every expected-empty judge — 13 `test -s … exit 1` judges (AC-001 ×2, AC-002, AC-006, AC-015, AC-026, AC-027 ×2, AC-028, AC-029, AC-030 ×3), the AC-016 after-list ("빈 파일" → `test -s` exit 1), the AC-030 (A) expectations for `ac030-changed.txt` and `ac030-artifact-commits.txt`, and the AC-030 case selector `ac030-emit-status.txt`. Mutant (scratch): missing `ac028-b.txt` → `test -e` exit=1, `test -s` exit=1 — the old emptiness-only check would have read PASS.
  - N3: AC-028/029 reading-list selector gained `|| /--auto-merge/`. Mutant (scratch, 3 lines): old selector selects line 1 only; new selector selects lines 1 and 2 ("`--auto-merge` merges as soon as CI checks pass."); line 3 selected by neither.
- **Revision note (0.2.3)** — every fixture line, command and expected result is recorded in acceptance.md next to its criterion; the lines below are the observed outputs from this authoring run (scratch fixtures; base fragments extracted from the template copies on HEAD `0af445525`):
  - D1 AC-GDP-026: added (iii) `--merge` presence ≥ 1 in the SKILL Flags line, the `workflows/sync.md` `**Flags**:` line, the Supported Flags section and the delivery Step 3.4 section; (ii) made directional (the `--merge` line must carry "deprecated alias of/for `--auto-merge`" after the token, and must not say "not/no longer/never deprecated" or "un-deprecated"); reading record `$E/ac026-reading.md` added as a PASS precondition. Observed: 10-line fixture → (ii) hits 1, 2, 3, 5, 8, 9, 10 (auditor's reversed line "`--merge`: auto-merge the PR after sync (the older `--auto-merge` spelling is deprecated)" = line 2, hit), correct lines 4 and 6 not hit, `--merged-only` line 7 not hit; X1-X3 fixture → hits 1, 3, 5 (unchanged). Removal fixture (auditor's "Modes: auto, force, status, project. Flags: --auto-merge, --skip-mx" plus three fragments without `--merge`) → presence `0 0 0 0`. Positive control → presence `1 1 1 1`, (ii) file empty (`test -s` exit 1), (i) `1 1 1 1`. Base fragments → (ii) the same 10 lines as 0.2.2, presence skill 1 / sync 1 / qgc-flags 1 / dl 2.
  - D2 AC-GDP-028: (a) now requires "all" within two words before "approv"; (b) adds optional / not needed / unnecessary / at least one / a single / after·on·once·upon·with + any·one·an·a·some·the first approval / majority; reading record `$E/ac028-reading.md` (covers AC-028 and AC-029). Observed on the 15-line fixture: (a) hits 3, 15; (b) hits 2, 7, 8, 12, 13 — auditor lines 7 ("approvals are optional") and 8 ("at least one approval") hit (b) and miss (a). Base mg/dl sections: (a) 0, (b) 0.
  - D3 AC-GDP-029 (b): trigger widened to approval or review; exemptions widened (without requiring, no approval/review, not needed); a gating phrase (after/once/until/pending/require + approval/review) is a violation regardless of exemptions. Observed: (a) hits 4, 6, 9, 10, 11, 14; (b) hits 4, 11, 14 — auditor's correct lines 9 and 10 no longer hit, "after a code review" (11) hits. Base: (a) 0, (b) 0.
  - D4 AC-GDP-027 (ii): added disable / override / turns off / suppress / bypass / cancel. Observed on the 7-line fixture: (i) hits 1, 3; (ii) hits 1, 2, 3, 5, 6, 7; correct line 4 not hit. Base dl: (i) 8, 19; (ii) 8, 19.
  - D5: scope set recounted as nine path pairs (manager-git, agent-common-protocol, delivery, doc-execution, moai SKILL, reference, quality-gates-context, workflows/sync.md, command source) — "열 개" → "아홉 개", "나머지 아홉" (after excluding agent-common-protocol) → "나머지 여덟", §A.5 "위 아홉" → "위 여덟", §E.1 "뺀 아홉" → "뺀 여덟"; §C.4 and plan §A affected files 21 → 19 (21 in case B). Counting rule stated in spec §C.4 and acceptance conventions.
  - D6: spec §E.2 citation `caa601d7c` → `0af445525`; `git diff --stat b412f8a33 HEAD -- <scope L·T, published SKILL local, zone-registry local, internal/template, Makefile, docs-site/content>` at `0af445525` → no output (`test -s` exit 1).
  - D7: acceptance conventions now require `BASE`/`E`/`T` in the same invocation as each judge command.
  - D8 AC-GDP-016: anchor is the oldest commit touching either copy of `agent-common-protocol.md` (`git log --format=%H --reverse … -- <local> <template>`, first line), with an explicit literal SHA substitution step. Observed at `0af445525`: list empty, `test -s` exit 1 (expected at base). The ordering mutants are logic-only (no commits made in this run).
- **Revision note (0.2.2)**:
  - In scope, local and template, line-level: X1 `workflows/sync.md` usage line (local 95 / template 85), X2 `argument-hint` in `.claude/commands/moai/sync.md:3` and template `sync.md.tmpl:3`, X3 `delivery.md:404` "Auto-Merge PR (/moai sync --merge)" (same line number in both copies). All three verified by reading the lines on this tree.
  - X2 is a command source, so REQ-GDP-014 now also binds `make commands-emit` then `make commands-emit-check` (exit 0) with any regenerated published skill in the same commit as the source edit; new AC-GDP-030 judges both outcomes (A no change / B change) and carries a control showing the check can fail.
  - X4 (docs-site four locales) is out of scope; the follow-up docs card and its measured inputs are recorded in spec.md §D.
  - Kept from the WIP commit after re-measurement: scope set (nine path pairs; the 0.2.2 text called it "ten", corrected in 0.2.3), nine AC-GDP-026/027 fragments and their base controls, AC-GDP-013 parity rows for the command pair and published skill, AC-GDP-015/016/025 file lists, AC-GDP-030 body.
  - Corrected: AC-GDP-025 registry detector `manager-git\.md` → `manager-[g]it\.md` (the worktree guard refuses the literal word inside a pattern; the uncorrected command could not have run at run-phase). spec.md §C.5 content moved into §D (the ruling asks for it there). Base measurement HEAD updated to `ea09ca650`.
  - Added: empty-set guards on AC-GDP-015 (template diff must be non-empty), AC-GDP-016 (the always-loaded-rule commit must exist), AC-GDP-025 (scope diff must be non-empty; mutant for the "others" registry detector), AC-GDP-030 (tracked-path control, source-commit existence, mechanical same-commit subset check, published file non-empty).
- **Tier**: M kept per lead ruling (judged counts decide; file-count guidance is reference). Judged REQ 12 and AC 16 are within 16/16; affected files 19 (nine scope path pairs × local·template = 18, plus the generated `.toml`), +2 if AC-GDP-030 outcome B (21). 0.2.2 stated 21 with +2 by miscounting the scope list as ten. **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree re-measured for 0.2.3**: HEAD `0af445525` (parent `3512b9f75`; both change only SPEC/report files) — `git diff --stat b412f8a33 HEAD -- <scope files L·T, published moai-sync SKILL.md local, zone-registry.md local, internal/template, Makefile, docs-site/content>` → no output (`test -s` exit 1). **Tree measured (0.2.2)**: HEAD `ea09ca650` (parent `caa601d7c`; both change only SPEC files). `git diff --stat b412f8a33 ea09ca650 -- <scope files L·T, manager-git.toml, published moai-sync SKILL.md L·T, internal/template/commandemit, internal/template/agentemit, zone-registry.md L·T, docs-site/content, Makefile>` → no output; same for the eight guarding test files; `git diff --stat b412f8a33 -- internal/template/templates/` → no output. Plan-time counts on these copies are base counts.
- **Audit count**: reduced iteration 2 FAIL 0.75 (`.moai/reports/t622/plan-audit-reduced-iter2.md`, measured on `3512b9f75` = 0.2.2); this 0.2.3 revision is for a limited iteration 3 restricted to D1-D8 and regressions, declared the last one. Reduced iteration 1 FAIL 0.71 (`.moai/reports/t622/plan-audit-reduced-iter1.md`). No audit file for 0.2.1 exists in `.moai/reports/t622/` (listing at 0.2.2 authoring). This revision is for reduced iteration 2. Full-scope iterations before the split: FAIL 0.67 and FAIL 0.75.
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

## §F Phase 4 Mode Selection

Decision: serial

Recorded by the lane (card orchestrator) before the first run-phase `Agent()` spawn, after Implementation Kickoff Approval (operator via lead, 2026-09-11, conditioned on N1-N3, applied in `ccfe3005e`).

**Input parameters**

| Parameter | Value |
|---|---|
| tier | M (judged REQ 12, AC 16) |
| scope (files) | 19 affected (nine scope files as local/template pairs = 18, plus generated `manager-git.toml`); 21 if the published `moai-sync/SKILL.md` pair changes (AC-GDP-030 case B) |
| domain count | 2 — agent/rule/skill/command instruction text, and generated artifacts via `make agents-emit` / `make commands-emit` |
| file language mix | Markdown instruction files, one Go-template command source (`.md.tmpl`), generated `.toml` and published `SKILL.md`; no Go source edits |
| concurrency benefit | LOW — milestones share files (delivery.md in M1 and M2, manager-git.md in M1-M3) and carry ordering obligations: command-source regeneration in the same commit (M1), agent regeneration after all manager-git.md edits (M4), always-loaded rule edit last (M5, AC-GDP-016) |
| Agent Teams prerequisites | not requested (no `--team`) |
| tree at selection | HEAD `7f5a4420f` (SPEC 0.2.4 `ccfe3005e` + evidence), clean; `moai session list --json --filter-spec=SPEC-GIT-DELIVERY-PROCEDURE-001` → `[]` |

**Mode evaluation**

| Mode | Result | Rationale |
|---|---|---|
| direct | not selected | multi-file, generator runs and acceptance judges exceed a direct orchestrator edit |
| serial | **selected** | one write-capable `manager-develop` runs M1→M6 in order; shared files and commit-ordering obligations make sequential execution the correct shape |
| fanout | not selected | below the multi-domain threshold, and parallel writers would race on delivery.md / manager-git.md and break the same-commit and last-edit ordering |
| sweep | not selected | not ≥ ~30 files and not one uniform mechanical transform |
| agent-team | not selected | not requested by the operator |

Decision: serial

**Justification.** The run phase edits a small set of instruction files line by line where later milestones depend on earlier commits (M4 regenerates the `.toml` only after M1-M3 finish editing `manager-git.md`; M5 must be the final instruction-file commit). A single sequential `manager-develop` preserves those orderings and keeps one writer on the worktree. Compile scope needs no `internal/cli` slot: `go list -deps -test` of `./internal/template/`, `./internal/template/commandemit/` and `./internal/template/agentemit/` lists no `internal/cli` package (same-form control on `internal/template` hits).
