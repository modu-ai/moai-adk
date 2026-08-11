---
id: SPEC-KANBAN-RENAME-001
title: "Progress — Factory Mode to Kanban Mode rename"
version: "0.4.0"
status: in-progress
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: cli
lifecycle: spec-anchored
tags: "rename, refactor, cli, template-mirror, behavior-preserving"
tier: L
---

# SPEC-KANBAN-RENAME-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: L (promoted from M at v0.2.0) — artifacts `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, plus `progress.md`.
- **Authored**: 2026-08-10, in the worktree `/Users/goos/.moai/worktrees/kanban` at HEAD `d39e3cdc6` (clean, at `origin/main`).
- **Requirements**: 25 (`REQ-KR-001` … `REQ-KR-025`), GEARS notation — **25 of the Tier L ceiling of 25, at the ceiling exactly with no headroom.** A twenty-sixth requirement forces a split of this SPEC; it cannot be absorbed, because Tier L is the top tier.
- **Acceptance criteria**: 28 (`AC-KR-001` … `AC-KR-028`), each naming its deciding command — **28 against the Tier L ceiling of 25, exceeding it by three.** The excess is carried as a **disclosed debt** for the plan auditor to rule on. The three cheapest merge candidates (`AC-KR-026`, `AC-KR-027`, `AC-KR-028`) are precisely the three the v0.1.1 audit added or separated for cause, so re-bundling them to fit the ceiling would trade a disclosed count for an undisclosed gap. Disposition is the orchestrator's decision, not this document's. Precedent: `SPEC-KANBAN-BOOTSTRAP-001` carries a four-criterion overage on the same terms; `SPEC-KANBAN-WORKTREE-001` was promoted in the same pass and lands within budget at 18 of 25.
- **SPEC ID check**: `[[ "SPEC-KANBAN-RENAME-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- **Revision**: v0.3.0 closed the Tier L plan-audit delta F1-F9 (§E.1.4); v0.2.0 promoted the tier and disclosed the budget (§E.1.3); v0.1.1 closed the plan-audit delta D1-D10 (§E.1.1).
- **Note**: Tier L raises the plan-auditor PASS threshold from 0.80 to 0.85. The v0.1.1 audit scored 0.92 against the Tier M threshold of 0.80; re-audited at Tier L the SPEC scored **0.848** — a marginal FAIL, with every blocking defect in the verification layer and every one delta-closable. The requirements themselves were audited at 0.92 and are unchanged.

### Measured, not assumed

Every figure in `spec.md` §A.3-§A.4 and `plan.md` §A was measured in this worktree at `d39e3cdc6`. Three measurements corrected the working assumptions the SPEC started from:

1. **Mirror parity is not uniform.** Three of six `.claude/` ↔ template pairs are *sanitized pairs*, not byte-identical. A byte-parity invariant would have instructed the implementer to copy §25-forbidden content into template source. Replaced by delta preservation (REQ-KR-018, AC-KR-017).
2. **Bare-word grep-zero is unworkable.** `grep -ri factory` matches ~110 files of unrelated vocabulary (measured 108 — `research.md` §H.3). Replaced by a token-scoped pattern, falsified against `internal/lsp`, `internal/tui`, `internal/hook`, `internal/core`, `docs-site/`, and `.claude/skills/moai/references/anti-patterns.md` — zero matches there, **28** on the real surface. Two figures were corrected at v0.3.0: the surface baseline is 28, not 26 — the 26 predates the v0.1.1 D4 widening to `.moai/project/` and contradicted this file's own D4 row two tables below ("baseline 26 → 28") — and `NOTICE.md` was dropped from the falsification list because no such file exists anywhere in this tree, so a run against it returned zero vacuously (`research.md` §H.2).
3. **`catalog.yaml` indexes directories, not files.** Grepping it for `factory` returns nothing, which could be misread as "unaffected". The `moai` skill's directory hash does change; `make build` output must be committed (AC-KR-020).

### E.1.1 v0.1.1 audit-delta closure (D1-D10)

| ID | Defect | Closure |
|---|---|---|
| D1 | `tags:` as a YAML sequence failed the decoder (`internal/spec/lint.go` declares `Tags string`), making every `moai spec` verb refuse the SPEC | `tags` is now a comma-separated string and `version` is quoted, on all four artifacts. `ParseFailure` is gone; `moai spec lint <spec.md>` reports **`✓ No findings`**, exit 0 |
| D2 | `exit=$?` read after a pipe reported `tail`'s status | AC-KR-011 and AC-KR-019 redirect to a log, read `$?` first, and assert a `^FAIL` count of 0 against the whole log |
| D3 | ref-less `git diff` is empty by construction after commit | AC-KR-012 / 023 / 024 anchored to `d39e3cdc6..HEAD` |
| D4 | two live surface files outside the enumerated 26 | `.moai/project/` added to the §D.1 scope (baseline 26 → 28); REQ-KR-024, AC-KR-028, milestone M2b |
| D5 | the token grep cannot see a leftover bare `-f` | REQ-KR-025 + AC-KR-026, file-scoped to the six contract documents; positive control **8**, measured |
| D6 | AC-KR-015 positive control stated 9 | corrected to **8** (re-measured); the nine *edit locations* in plan.md M2 step 3 are kept as a distinctly-labelled figure |
| D7 | `-k` collision unresolved | probed: no collision, short flags are `-c -d -h -n -p -r -v -w`; recorded in spec.md §A.6, M0 gate retained as re-confirmation |
| D8 | the two `$TOK` copies had drifted | the extra backtick alternative dropped; copies byte-identical; AC-KR-027 is the drift check |
| D9 | Go non-test count labelled 7 | corrected to **8** (the prose already enumerated eight paths) |
| D10 | BRE `\|` / `\b` without `-E` | AC-KR-003 and AC-KR-008 switched to `grep -E` with a plain `|` |

### E.1.2 Lint-gate finding: the glob invocation is structurally unsatisfiable

The audit prescribed `moai spec lint .moai/specs/SPEC-KANBAN-RENAME-001/*.md` exiting 0 as the D1 gate. That invocation **cannot exit 0 for any multi-artifact SPEC in this repository**, before or after this fix, and the reason is worth recording so it is not re-diagnosed later.

`DuplicateSPECIDRule.CheckAll` treats every path passed on the command line as a distinct SPEC and errors when two share an `id`. A non-`spec.md` artifact therefore has two possible states, and both fail the glob:

- **No frontmatter** → `ParseFailure`. Measured: `SPEC-ORCH-GIT-RELAX-001` glob → 5 errors, all `ParseFailure` on `plan.md` / `progress.md` / `research.md`.
- **Frontmatter carrying the same `id`** → `DuplicateSPECID`, since `id` is one of the 12 required fields. Measured: `SPEC-CLI-WIZARD-RESTRUCTURE-001` glob → 5 errors, all `DuplicateSPECID`.

There is no third state. The linter's own discovery unit confirms which invocation is canonical: `discoverSPECs` (`internal/spec/lint.go:308`) globs `SPEC-*/spec.md` and nothing else, so the no-argument run — the one CI performs — never passes a sibling artifact at all.

The two achievable gates, both measured in this tree after the fix:

```
$ go run ./cmd/moai spec lint .moai/specs/SPEC-KANBAN-RENAME-001/spec.md
✓ No findings — all SPEC documents are valid          # exit 0

$ for f in spec plan acceptance design research progress; do
    go run ./cmd/moai spec lint .moai/specs/SPEC-KANBAN-RENAME-001/$f.md; done
spec.md       → exit 0, no findings
plan.md       → exit 0, 1 warning  (MissingExclusions, grandfathered)
acceptance.md → exit 0, 1 warning  (MissingExclusions, grandfathered)
design.md     → exit 0, no findings
research.md   → exit 0, no findings
progress.md   → exit 0, 1 warning  (MissingExclusions, grandfathered)
```

Zero errors on every artifact individually — re-run at v0.3.0 over all six, the two Tier L artifacts included. `design.md` and `research.md` are clean because each carries its own `### Out of Scope —` sub-headings. The residual `MissingExclusions` warnings are grandfathered-era downgrades that fire because the rule looks for an Out-of-Scope section in every file it is handed; `spec.md` — the file the rule actually binds — carries three `### Out of Scope —` sub-headings and is clean.

### E.1.3 v0.2.0 tier promotion and budget disclosure

**What forced it.** Measured at v0.2.0 authoring time, in this worktree:

```
$ grep -cE '^\*\*REQ-KR-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KR-[0-9]{3}\*\*' acceptance.md
28
```

Against the Tier M ceilings of 16 and 16 the SPEC was **nine requirements and twelve criteria over budget**. The governing rule (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier) reads an overage as a signal to tier up or split, never to relax the budget, so the tier was raised. At Tier L the requirements fit exactly at 25 of 25 and the criteria remain 3 over at 28 of 25.

**The gap this closes is a disclosure gap, not a requirement defect.** At v0.1.1 no artifact stated a count against a ceiling anywhere: `spec.md` §B said "Requirement count: 25" and named no ceiling, `acceptance.md` said nothing about its count at all, and this file listed both figures as bare totals. That silence is the mechanism by which a nine-over/twelve-over condition survived an independent plan audit at **0.92** — an auditor reading the document was never invited to make the comparison, and the number that would have prompted it was never placed beside the number it violates. The requirements themselves were not in question at v0.2.0 and are not re-opened: nothing is added, removed, renumbered, or reworded.

**Sibling consistency.** All three siblings now carry a count-and-ceiling note in the same three places — `spec.md` §B, the acceptance traceability section, and `progress.md` §E.1. `SPEC-KANBAN-WORKTREE-001` lands within budget at 18 of 25; `SPEC-KANBAN-BOOTSTRAP-001` carries a four-criterion overage; this SPEC carries three. The overage disposition in each case is the plan auditor's ruling, not the document's claim.

**Artifacts added by the promotion.** `design.md` records the decisions this rename rests on with the rejected alternative and measured reason for each; `research.md` records every measurement as command → observed output → what it establishes. Both were authored at v0.2.0 by re-running the measurements rather than transcribing figures from the existing prose, and §H of `research.md` records the three places where the re-measurement disagreed with what the SPEC asserts.

### E.1.4 v0.3.0 audit-delta closure (F1-F9)

Scored **0.848** against the Tier L PASS threshold of **0.85**. The shortfall sat entirely in the verification layer — the requirements were audited at 0.92 and none is re-opened — and every defect closed with a delta. **No requirement and no criterion was added, removed, or renumbered**; the counts are re-measured below and unchanged.

| ID | Defect | Closure |
|---|---|---|
| F1 | `AC-KR-020` used a **ref-less** `git diff` on `catalog.yaml`, empty once M3 commits it — so the criterion FAILs at the moment it runs while asserting the file is committed | anchored to `d39e3cdc6..HEAD`; re-checked, no other criterion carries the ref-less form. `plan.md` M3 step 2's bare diff runs pre-commit and stays bare |
| F2 | `go test -run` with a pattern matching nothing **exits 0 and prints `PASS`** — and `AC-KR-001` / `AC-KR-005` are keyed on post-rename names, so `REQ-KR-011` traced only to criteria that pass when it is not done | each run now pairs with a name-existence `grep` and an absent-`[no tests to run]` assertion; `[no tests to run]` chosen over the `-v`-only warning line because it appears in both renderings. Rule added at acceptance.md §A.1, hazard at plan.md §B-6, anti-pattern AP-11 |
| F3 | `AC-KR-025` grepped `moai cc --help` for `factory` — **vacuous**: the help string never named the flag, so the check returned zero before the rename too, and the missing positive control concealed it | reduced to the exit-0 smoke it can decide, with the measured baseline stated. Discriminating it would need a new requirement (help text must document `-k`), which the ceiling forbids. Anti-pattern AP-12 |
| F4 | `AC-KR-012`'s `grep -viE 'kanban|factory'` filter discards ~~**every**~~ **[corrected at v0.4.0: 22 of 226, 9.7% of]** assertion line in these tests, so a deleted or weakened assertion was invisible — and `REQ-KR-013`'s other criterion (`AC-KR-011`) stays green on a weakened assertion by construction | added a filter-independent `+`/`-` count-invariance check on `t.Error`/`t.Fatal` lines, against a measured baseline of **226** across the six surface test files |
| F5 | `AC-KR-028`'s control said "2 files" but its command counted **lines** (measured **3**), and its mechanical half could not see a partial edit of `modules.md` | `-l` added, control restated as 2 files / 3 lines; the `research.md` §H.4 line-granularity gap **closed** with a bare-word grep bounded to the two named files (baseline **5**) |
| F6 | five references handed scope and cross-references to `SPEC-KANBAN-MULTISESSION-001`, which is not under `.moai/specs/` — superseded, preserved read-only — while the budget prose already named the three real siblings | each scope assigned to its actual owner across `spec.md` §A/§C/§E and `design.md` §C.2/§G; two mentions retained, both framing it as the retired predecessor |
| F7 | `NOTICE.md` does not exist anywhere in this tree: §C asserted an attribution living there and §D.1 listed it plus `references/anti-patterns.md` as falsification targets, two of seven returning zero **vacuously** | attribution corrected to `.moai/research/` (five files, measured); the falsification list corrected to five real targets. `research.md` §H.2's disposition revised from *report* to *correct* — a prior score does not immunise a false statement of fact |
| F8 | this file said "26 on the real surface", contradicting its own D4 row ("baseline 26 → 28") and the measured 28 | corrected to 28, with the stale `NOTICE.md` target dropped from the same sentence |
| F9 | the `.moai/state/kanban/` collision with `SPEC-KANBAN-BOARD-001` was resolved in the sibling, unrecorded here | the sibling moved its board state to `.moai/state/kanban-board/` (`REQ-KB-005`, its §A.3(e)); `REQ-KR-009` is deliberately unchanged, and the coexistence is now stated in `spec.md` §C and `design.md` §G so a later reader finds it recorded rather than re-inferring a resolved collision |

Counts re-measured after the edits, in this worktree:

```
$ grep -cE '^\*\*REQ-KR-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KR-[0-9]{3}\*\*' acceptance.md
28
```

**Declined, and why.** The auditor listed four optional findings not worth the budget, and they were left open deliberately: the unanalysed split remedy in the overage disclosure, the stale `related_specs:` frontmatter, the three criteria whose missing positive controls are defensible because their baseline is already zero, and the missing `test -f` guards on four criteria. Two figures were also left as they stand — "roughly 110" against a measured 108 (`research.md` §H.3 argues the approximation) and the Go non-test count of 8, which confirms the v0.1.1 correction and is correctly attributed there.

### E.1.5 v0.4.0 audit-delta closure (D1-D4)

Four defects, all in the verification layer; the requirements layer is re-opened at exactly one point (`REQ-KR-012` gains a clause) and no requirement or criterion is added, removed, or renumbered. Counts re-measured below and unchanged.

| ID | Defect | Closure |
|---|---|---|
| D1 | `REQ-KR-011` binds **16** test function names; its only criteria (`AC-KR-001`, `AC-KR-005`) reach **7** through their `-run` patterns, and the other **9** carry no `$TOK` token either — so `AC-KR-021` reads `0` with nine test names still announcing the old mode | `AC-KR-001` gains a bare-word grep bounded to the six surface test files (baseline **16** → target **0**), `plan.md` M1 step 6 enumerates the nine by name and line, `design.md` §F.3 records it as the fourth blind spot with the `$TOK`-extension alternative rejected |
| D2 | §A.1 declares two guards for **every** `-run`-keyed criterion; `AC-KR-002` and `AC-KR-009` carried neither. Worst on `AC-KR-009`, the **sole** criterion for both `REQ-KR-009` and `REQ-KR-010`, where `REQ-KR-010` is covered by an **absence** claim that a run which cannot fail does not decide | both criteria now carry the name-existence grep and the absent-`[no tests to run]` assertion; `AC-KR-009` additionally gains an old-path grep over `internal/kanban/` (baseline **3** lines → target **0**), which is what makes the absence claim falsifiable. §A.1 now names all four `-run`-keyed criteria; `plan.md` §B-6 updated to match |
| D3 | the v0.3.0 rationale for `AC-KR-012`'s count check claimed the filter discards **every** assertion line — measured, **22 of 226 (9.7%)** — and the check does not answer the hazard that rationale named: a `+`/`-` **count** comparison catches deletion and addition, not a **weakened predicate at constant line count** | figures corrected in `acceptance.md` and in the v0.3.0 F4 row above and item (4) of `spec.md` HISTORY, struck in place rather than silently restated; the predicate-weakening gap recorded as a **residual risk** in `design.md` §D, with the two mechanical alternatives (semantic assertion diff, hand-maintained 226-line inventory) and the reason each was declined |
| D4 | three names mix an `AC-FM-*` citation with a mode token (`TestACFM022a_Factory…` ×2, `TestACFM023c_Factory…`), so `REQ-KR-011` required renaming what `REQ-KR-012` protected and neither said which prevailed | `REQ-KR-012` gains a clause scoping its protection to the **citation substring** alone; the mode token in the same identifier is still renamed. The two bind different substrings and were never in conflict — `TestACFM022a_KanbanRaisesBlockCapUnconditionally` satisfies both. Restated in `plan.md` M1 step 6 where the implementer meets them |

Counts re-measured after the edits, in this worktree:

```
$ grep -cE '^\*\*REQ-KR-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KR-[0-9]{3}\*\*' acceptance.md
28
```

**Artifact versions.** `spec.md`, `plan.md`, `acceptance.md`, `design.md`, and this file move to `0.4.0`. `research.md` stays at `0.3.0` because it was not edited — every measurement this revision rests on was taken fresh in the worktree and is recorded here and at the criterion that uses it, and bumping an untouched artifact's version would assert a change that did not happen.

**One premise of the audit brief did not survive re-measurement, and it is recorded rather than quietly absorbed.** The brief stated that six test functions lie outside the `-run` patterns of `AC-KR-001` and `AC-KR-005`. Measured, it is **nine**: the brief's list omits `TestEnterFactoryMode_RestoresPriorValue` (`cc_test.go:546`), `TestEnterFactoryMode_WithoutSpecLeavesSpecVarUntouched` (`cc_test.go:570`), and `TestGLM_FactoryFlagParity` (`glm_test.go:646`) — none of which matches `ParseKanbanFlag|KanbanFlagStripped` or `CG_.*Kanban` post-rename. It also lists `TestRecordPathIsSessionKeyedUnderStateFactory` as uncovered, which is right on the brief's own terms (it is outside both patterns) though `AC-KR-009`'s rename-invariant `-run 'Path'` does select it. The fix is unaffected — a six-file bare-word grep with a baseline of 16 covers all sixteen names regardless of which are individually watched — but the enumeration in `plan.md` M1 step 6 carries nine, not six.

### Open question carried into run-phase

- **None remaining at plan-phase.** The `-k` collision question is answered (no collision — §A.6 of `spec.md`). M0 keeps the probe as a re-confirmation gate against the run-phase tree, because the `claude` CLI surface drifts between versions; a collision found there is still a blocker to surface, not a letter to silently re-pick.

### Decisions recorded

- **No deprecation alias** — verified: `internal/cli/factory.go` absent from `v3.0.1`, present only in `v3.1.0-rc.0`/`rc.1`; `CHANGELOG.md` has zero `factory` occurrences.
- **No state-record migration** — records are session-scoped and best-effort; a launch never depends on one (REQ-KR-010).
- **`AC-FM-*` identifiers preserved** — they cite a closed SPEC's acceptance criteria (REQ-KR-012).
- **`SPEC-FACTORY-MODE-001/` preserved verbatim** — closed historical record (AC-KR-024).
- **docs-site untouched** — its four `factory` hits are all `ExecutionFactory` in an unrelated example (AC-KR-023).

## §E.2 Run-phase Evidence

### Claim

The rename landed across all five surfaces — Go source, harness documentation, template mirror, generated catalog, and project documentation — with **behavior preserved and zero assertion drift**. All 28 acceptance criteria PASS. Five commits sit on `spec-kanban` in the worktree `/Users/goos/.moai/worktrees/kanban`, none pushed; each used `SKIP_MOAI_PRECOMMIT=1`, because `.git/hooks/pre-commit` runs `moai gate`, which exits non-zero on pre-existing ast-grep findings unrelated to this SPEC. Nothing under `/Users/goos/MoAI/moai-adk-go/` was touched.

| SHA | Milestone |
|---|---|
| `5cdc8b68d` | M1 — Go rename |
| `0d39be1d4` | M2 — harness documentation rename |
| `768024f30` | M2b — project documentation |
| `e8dd51918` | criterion amendment (`AC-KR-020`, `AC-KR-028`) |
| `f7e1ffdf2` | cross-reference correction in `AC-KR-028`'s rationale |

The substantive finding of this run is not in the implementation. **Three plan-phase criteria were defective, and running them is what exposed the defects** — none of the three required an implementation change:

1. **`plan.md` §F M2 step 3 misattributed which line the token pattern cannot see.** It named `quality-gates-quality.md` line ~112. Measured, line 112 *does* match — `Factory dedup gate` hits the `dedup` alternative. The invisible line is `mode-orchestration.md` line 84, where a backtick between `` `factory` `` and ` pipeline` breaks the `[Ff]actory pipeline` alternative. Sharper still: `$TOK` and `AC-KR-015`'s case-sensitive lowercase grep each match 8 lines, but **not the same 8** — `$TOK` sees `qgates:112` and misses `modeorch:84`; the lowercase grep sees `modeorch:84` and misses `qgates:112` (capitalized `Factory`). Their union is the 9 edit locations. Both were edited explicitly, and a case-insensitive sweep of the five sibling documents now returns `0` on both sides. The plan's 9-vs-8 arithmetic was never wrong — only its attribution of which pattern was blind to which line.
2. **`AC-KR-020`'s premise was false** (amended at v0.5.0). It asserted that editing template source necessarily changes `catalog.yaml`; `resolveHashSourcePath` hashes the skill directory's `SKILL.md` only, so an edit to a sibling workflow document under the same skill leaves the hash — and therefore the catalog — unchanged. The criterion was rewritten to require the catalog be committed *if and only if* the build changed it, which is what `make build` plus a clean porcelain actually decides.
3. **`AC-KR-028`'s third criterion was unsatisfiable** (amended at v0.5.0, its rationale's cross-reference corrected at v0.5.1). The unbounded bare-word grep counted the substring `FACTORY` inside the citation `SPEC-FACTORY-MODE-001`, which `REQ-KR-024` never asked to be renamed — driving it to `0` would require inventing a SPEC identifier or deleting the citation and orphaning the preserved record. The target was unreachable; the criterion now excludes the protected citation and nothing else.

### Evidence

All 28 criteria, each with the observation that decided it. Every figure below was read from a command run in this worktree; the HEAD at which each was taken is named in §Baseline-attribution.

| AC | Deciding observation | Result |
|---|---|---|
| `AC-KR-001` | name greps `TestParseKanbanFlag` → 4, `TestCC_KanbanFlagStripped` → 1; `go test ./internal/cli/ -run 'ParseKanbanFlag\|KanbanFlagStripped' -v` → `exit=0`; `[no tests to run]` → 0; actual `--- PASS` → 5 | PASS |
| `AC-KR-002` | name grep `TestParseKanbanFlag_PassThroughBoundary` → 1; `-run 'PassThroughBoundary'` → `exit=0`; `[no tests to run]` → 0; `--- PASS` → 1 | PASS |
| `AC-KR-003` | `grep -nE -- '"--factory"\|"-f"' internal/cli/kanban.go` → 0 matches | PASS |
| `AC-KR-004` | `claude --help` probe → no `-k` match; short-flag set `-c -d -h -n -p -r -v -w` | PASS |
| `AC-KR-005` | name grep `TestCG_.*Kanban` → 2; `-run 'CG_.*Kanban'` → `exit=0`; `[no tests to run]` → 0; `--- PASS` → 4 | PASS |
| `AC-KR-006` | `ls internal/kanban/` → 4 files (`record.go`, `record_test.go`, `revision.go`, `revision_test.go`); `internal/factory` absent → `OK` | PASS |
| `AC-KR-007` | `internal/cli/kanban.go` present, `internal/cli/factory.go` absent → `OK` | PASS |
| `AC-KR-008` | `envkeys.go:142` `EnvMoaiKanban = "MOAI_KANBAN"`, `:147` `EnvMoaiKanbanSpec = "MOAI_KANBAN_SPEC"`; call-site literal grep → 0 matches (`grep-exit=1`) | PASS |
| `AC-KR-009` | `record.go:43` `stateDirSegments = []string{".moai", "state", "kanban"}`; name grep → 1; `-run 'Path'` → `exit=0`, `[no tests to run]` 0, `--- PASS` 2; old-path grep over `internal/kanban/` → **0** (baseline 3) | PASS |
| `AC-KR-010` | `internal/cli/kanban.go:108` `func captureEnvState(key string)` — present, unrenamed | PASS |
| `AC-KR-011` | `go test ./... -count=1 -timeout 900s` → `test-exit=0`, `ok` 112, `^FAIL` 0 | PASS |
| `AC-KR-012` | `git diff c5d4a275d..5cdc8b68d -- '*.go'` → 203 insertions / 201 deletions over 14 files; changed lines filtered for non-rename content leave exactly one hunk, a doc comment | PASS |
| `AC-KR-013` | `AC-FM-` sum → **50** (baseline 50); `^func Test.*ACFM` → **3** (baseline 3) | PASS |
| `AC-KR-014` | `workflows/kanban.md` present on both sides; `workflows/factory.md` on neither | PASS |
| `AC-KR-015` | case-insensitive `factory` sweep over the five sibling documents, both sides → 0 | PASS |
| `AC-KR-016` | `kanban_chain` → 2; `factory_chain` → 0 | PASS |
| `AC-KR-017` | six-pair delta loop against the M0 baselines → six `OK`, zero `MISSING`, zero `DRIFT` | PASS |
| `AC-KR-018` | `grep -cE 'SPEC-[A-Z0-9-]+-[0-9]{3}'` on the renamed template contract document → 0 | PASS |
| `AC-KR-019` | `go test ./internal/template/... -count=1` → `tmpl-exit=0`, `^FAIL` 0, `ok … 19.842s` uncached | PASS |
| `AC-KR-020` | `make build` → `make-exit=0`, printed `catalog.yaml updated successfully (12403 bytes)`; `git status --porcelain` afterwards **empty** — the build changed nothing, so nothing was owed a commit | PASS |
| `AC-KR-021` | `grep -rlniIE "$TOK" internal/ .claude/ .moai/project/` → **0** (baseline 28) | PASS |
| `AC-KR-022` | same pattern over `internal/lsp internal/tui internal/hook internal/core docs-site/` → 0; control `grep -rli factory internal/lsp` → **9**, unrelated vocabulary untouched | PASS |
| `AC-KR-023` | `git diff --name-only d39e3cdc6..HEAD -- docs-site/` → 0 lines | PASS |
| `AC-KR-024` | same, `.moai/specs/SPEC-FACTORY-MODE-001/` → 0 lines | PASS |
| `AC-KR-025` | `./bin/moai --version` → `version-exit=0`, 1 line; `./bin/moai cc --help` → `help-exit=0`, 47 lines | PASS |
| `AC-KR-026` | bare `-f` over the six contract documents, both sides → **0** (baseline 8) | PASS |
| `AC-KR-027` | `TOK='` lines from `spec.md` and `acceptance.md` diffed → `OK`, byte-identical | PASS |
| `AC-KR-028` | renamed forms read back at `modules.md` 157/158/161/246 and `structure.md` 139; token grep over `.moai/project/` → 0; bounded bare-word grep → **0** (baseline 3) | PASS |

Three mixed identifiers landed correctly under `AC-KR-013`, preserving the closed SPEC's citation while renaming the mode token: `TestACFM022a_KanbanRaisesBlockCapUnconditionally`, `TestACFM022a_KanbanCapReplacesPreexistingEntry`, `TestACFM023c_KanbanEnvReachesChildEnvironment`.

The single non-rename hunk surfaced by `AC-KR-012` is a doc comment describing the flag, and its factual claim was verified independently rather than accepted: cobra shorthand `"k"` registrations across `internal/ cmd/ pkg/` → **0**; the only `"-k"` literals in the tree are the five this rename introduced; and the prior `-f` binding the comment references is real (`internal/cli/state.go:54` binds `-f` as the `--format` shorthand).

**Pre-flight (§C) — the baseline was RED, and it was characterized rather than resolved.** `go build ./... && go test ./... ` at `c5d4a275d` returned `base-exit=1` with `^FAIL` 3 and `ok` 111 over a 760-line log. The single failing test:

```
--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.10s)
    navigator_enrich_test.go:91: barrier file not created (goroutine did not reach barrier)
```

It is pre-existing and off-surface on three independent grounds: `grep -ciE 'factory|kanban' internal/cli/navigator_enrich_test.go` → **0**; `git diff --name-only d39e3cdc6..HEAD | grep -vc '^\.moai/specs/'` → **0**, so the branch had touched nothing outside the SPEC directory before M1; and it is racy rather than deterministic — the test waits on a 2-second wall-clock deadline for a goroutine to create a barrier file, and isolated re-runs failed **4 of 5**, with `-count=10` failing **8 of 10**. It subsequently passed in both post-M1 full-suite runs, which is consistent with the flakiness, not with a fix.

Recorded for the next reader: `internal/cli` takes ~227s and `TestGateCmd_RunE_Behavior` alone ~121s (`--- PASS: TestGateCmd_RunE_Behavior (121.09s)`). That package is slow, not hung — a short `-timeout` mis-diagnoses it as a hang.

### Baseline-attribution

- **M0 baselines** at `c5d4a275d`, before any edit, persisted at `.moai/state/verify/kr-rename/`: token grep → **28** files (2 under `.moai/project/`); six-pair `diff` line counts `contract` 0, `run` 0, `goal` 0, `moaidoc` 5, `modeorch` 7, `qgates` 9 (3 byte-identical + 3 sanitized, matching `spec.md` §A.4); `claude --help` probe → no `-k`; bare `-f` over the six contract documents, both sides → **8**; `AC-FM-` sum → **50** and `ACFM` function count → **3**; `^func Test.*[Ff]actory` over the six surface test files → **16**; case-insensitive `factory` over the two project documents → **5** lines.
- **M1 verification** at `5cdc8b68d`: `go build ./...` → 0; `go vet ./internal/{cli,kanban,config}/` → 0; factory test functions over the six files → **0** (from 16); `AC-FM-`/`ACFM` invariants held at 50/3; full suite → `test-exit=0`, 112 `ok`, 0 `FAIL`.
- **M2 / M2b / M3 verification** at `768024f30`: token grep 0; delta preservation six `OK`; bare `-f` 0; neutrality 0; exclusions all 0 lines against `d39e3cdc6..HEAD`; `make build` `make-exit=0` with empty porcelain; `go test ./internal/template/... -count=1` → `tmpl-exit=0`.
- **M4 final verification** at `f7e1ffdf2`: `go build ./...` → 0; `go test ./... -count=1 -timeout 900s` → `test-exit=0`, 112 `ok`, 0 `FAIL`, 115 log lines; `moai spec lint` → `✓ No findings`; token grep 0; delta preservation six `OK`; `AC-FM-` 50 / `ACFM` 3; old flags in `internal/cli/kanban.go` → 0; amended `AC-KR-028` → 0 and 0; amended `AC-KR-020` → empty porcelain.
- **Anchor**: `d39e3cdc6` is the merge-base with `origin/main` and an ancestor of HEAD (`git merge-base --is-ancestor` → YES), so every exclusion diff is anchored to a commit this branch genuinely descends from.
- **Filled at report time**, on the same tree at `f7e1ffdf2`, for the ten criteria the run's own verification batches had not individually recorded: `AC-KR-002`, `AC-KR-005`, `AC-KR-006`, `AC-KR-007`, `AC-KR-008`, `AC-KR-009`, `AC-KR-010`, `AC-KR-016`, `AC-KR-025`, `AC-KR-027`, plus the `AC-KR-022` post-rename falsification and its `internal/lsp` control and the `AC-KR-028` read-back. Each was decided by its own criterion's command as written in `acceptance.md` §B — greps, file-existence checks, the two-line `TOK` diff, the binary smoke, and three targeted `-run` invocations. The full suite was **not** re-run for this report; `AC-KR-011` and `AC-KR-019` transcribe the M4 and M3 measurements above.

### Gaps

- **The pre-flight baseline was characterized, not resolved.** `TestNavigatorEnrich_AtomicWriteBarrier` remains a racy pre-existing test. It was not fixed — out of scope for this SPEC.
- **The `-f` → `-k` substitution in M2 used a Perl lookaround** (`(?<![-\w])-f(?![\w])`), not the ERE that `AC-KR-026` names. Both landed on the same 8 sites and the criterion's own pattern now reads 0, but the two patterns are **not proven equivalent in general**.
- **Template neutrality was verified per-file only on the renamed contract document.** For the four other edited template files the evidence is the shipped guard passing, not a per-file grep.
- **Rendering was not checked.** No one confirmed the renamed documents still read coherently to a human; only that the tokens changed.
- **The six-pair delta baselines were originally written to `/tmp`.** They were copied to `.moai/state/verify/kr-rename/` (gitignored) mid-run, so `AC-KR-017` remains re-runnable, but the `/tmp` copies may be cleared by the OS.
- **No rebase was performed.** `origin/main` is **7 commits ahead** of this branch (`git rev-list --count --left-right origin/main...HEAD` → `7 5`). All 7 are documentation commits touching `docs-site/` and `.moai/specs/SPEC-CODEX-PHASE2-001/`; none touches `internal/` or `.claude/`. The rebase was deliberately deferred to PR time because it would move `docs-site/` into the `d39e3cdc6..HEAD` window (`git diff --stat d39e3cdc6..origin/main -- docs-site/` → 309 files changed, 4740 insertions, 1515 deletions) and spuriously fail `AC-KR-023`'s exclusion check.

### Residual-risk

- **`structure.md:139` now asserts that `internal/kanban` is absent from `origin/main`.** The assertion is true, but it is a rename-induced restatement of a measured claim that was not itself re-measured during this run.
- **M2's blanket substitution was safe only because the enumeration preceded it.** Every `factory` occurrence in the six documents was listed first and confirmed to be mode vocabulary, with zero unrelated matches. That property is a fact about this edit, not a guarantee about future edits to those files.
- **The pre-existing flaky test passed in both post-M1 suite runs.** A future run may show it red again, and a reader must not mistake that for a regression introduced by this SPEC.

## §E.3 Run-phase Audit-Ready Signal

- **Tier**: L, unchanged from plan-phase — the run touched no requirement and added none, so the 25-of-25 requirement ceiling and the disclosed three-criterion acceptance overage (§E.1) stand exactly as audited.
- **Implemented**: 2026-08-11, in the worktree `/Users/goos/.moai/worktrees/kanban` on branch `spec-kanban`, across five commits — `5cdc8b68d` (M1, Go rename), `0d39be1d4` (M2, harness documentation), `768024f30` (M2b, project documentation), `e8dd51918` (criterion amendment), `f7e1ffdf2` (cross-reference correction) — plus `630eb5c5f` recording this evidence. **None pushed.** Every commit used `SKIP_MOAI_PRECOMMIT=1`, because `.git/hooks/pre-commit` runs `moai gate`, which exits non-zero on pre-existing ast-grep findings unrelated to this SPEC; the bypass is recorded here rather than silently relied on, per `acceptance.md` §C.
- **Criteria disposition**: 28 of 28 PASS, each decided by its own command from `acceptance.md` §B. The per-criterion evidence, with the HEAD each figure was taken at, is §E.2 — not restated here.
- **Final verification** at `f7e1ffdf2`: `go build ./...` → 0; `go test ./... -count=1 -timeout 900s` → `test-exit=0`, 112 `ok`, `^FAIL` 0, 115 log lines; `moai spec lint` → `✓ No findings`.
- **Artifact revisions at run-phase exit**: `spec.md` v0.5.1, `plan.md` v0.5.0, `acceptance.md` v0.5.1. All four artifacts transitioned `draft → in-progress` on `630eb5c5f`; the close to `implemented` and `completed` rides the sync commit and is not this phase's to make.
- **Defect distribution**: **three plan-phase defects, zero implementation defects.** All three surfaced from *running* criteria that had only been reasoned about at plan time — a misattribution in `plan.md` §F M2 step 3, a false premise in `AC-KR-020`, and an unsatisfiable third criterion in `AC-KR-028` — and none of the three required a change to the implementation. The amendment trail is `e8dd51918` (v0.5.0, both criteria) and `f7e1ffdf2` (v0.5.1, the cross-reference in `AC-KR-028`'s rationale); §E.2 §Claim carries the substance.
- **Threshold note**: the Tier L plan-audit threshold of 0.85 governed entry to this phase, not its exit. What gates the run is `acceptance.md` §C, whose four conditions are met: all 28 criteria PASS with their commands run in this tree, the §E 5-section report is written into §E.2 with a non-empty Gaps section, no commit touches `/Users/goos/MoAI/moai-adk-go/`, and the `SKIP_MOAI_PRECOMMIT=1` use is recorded above.

### A plan-phase premise that run-phase falsified

**§E.1 "Measured, not assumed" item 3 asserts that "the `moai` skill's directory hash does change", and that premise is false.** `resolveHashSourcePath` in `internal/template/scripts/gen-catalog-hashes.go` resolves a skill directory's hash to its root `SKILL.md` alone, so editing a sibling workflow document under the same skill leaves the hash — and therefore `catalog.yaml` — untouched. Measured at `768024f30`: `make build` exits 0 and prints `catalog.yaml updated successfully (12403 bytes)`, while `git status --porcelain` immediately afterwards is **empty**. The authoritative correction is the v0.5.0 amendment to `AC-KR-020` in `acceptance.md`, which now requires the catalog be committed *if and only if* the build changed it.

§E.1 is deliberately left as written. It is the plan-phase record of what was believed and measured at plan time, `manager-spec` owns it, and overwriting the sentence would erase the fact that the belief was held — which is the part worth keeping. This note is the forward correction; the amended criterion is the authority.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
