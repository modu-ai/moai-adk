---
id: SPEC-KANBAN-RENAME-001
title: "Progress — Factory Mode to Kanban Mode rename"
version: "0.3.0"
status: draft
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
| F4 | `AC-KR-012`'s `grep -viE 'kanban|factory'` filter discards **every** assertion line in these tests, so a deleted or weakened assertion was invisible — and `REQ-KR-013`'s other criterion (`AC-KR-011`) stays green on a weakened assertion by construction | added a filter-independent `+`/`-` count-invariance check on `t.Error`/`t.Fatal` lines, against a measured baseline of **226** across the six surface test files |
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

### Open question carried into run-phase

- **None remaining at plan-phase.** The `-k` collision question is answered (no collision — §A.6 of `spec.md`). M0 keeps the probe as a re-confirmation gate against the run-phase tree, because the `claude` CLI surface drifts between versions; a collision found there is still a blocker to surface, not a letter to silently re-pick.

### Decisions recorded

- **No deprecation alias** — verified: `internal/cli/factory.go` absent from `v3.0.1`, present only in `v3.1.0-rc.0`/`rc.1`; `CHANGELOG.md` has zero `factory` occurrences.
- **No state-record migration** — records are session-scoped and best-effort; a launch never depends on one (REQ-KR-010).
- **`AC-FM-*` identifiers preserved** — they cite a closed SPEC's acceptance criteria (REQ-KR-012).
- **`SPEC-FACTORY-MODE-001/` preserved verbatim** — closed historical record (AC-KR-024).
- **docs-site untouched** — its four `factory` hits are all `ExecutionFactory` in an unrelated example (AC-KR-023).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
