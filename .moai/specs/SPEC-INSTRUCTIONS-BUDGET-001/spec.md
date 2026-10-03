---
id: SPEC-INSTRUCTIONS-BUDGET-001
title: "InstructionsLoaded aggregate character budget: mechanically-derived instruction-file set with a session-level character constant"
version: "0.1.0"
status: completed
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.2.1"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook,instructions-loaded,budget,aggregate,always-loaded"
tier: M
related_specs: []
---

# SPEC-INSTRUCTIONS-BUDGET-001

## §A — History

- **2026-10-01** — plan-phase v0.1.0 authored from card t1318 (Tier M, Class C, lead-issued 09-29). **Operator ruling (2026-10-01, via AskUserQuestion): `sessionCharBudget = 210000`.** Card-recorded context: t1303 (09-29) measured an aggregate floor of 205,977 chars over an 18-file instruction set; the backing report `.moai/reports/t1303/` is absent in the current checkout, so 205,977 is carried as a card-recorded figure, not a re-citable baseline. Plan-phase re-measurement on tree `2868afc0b` (this run, this tree): the same 18 files now aggregate **214,155 chars** (227,076 bytes) — the always-loaded surface has grown ~8,178 chars since t1303 and **already exceeds the 210,000 ruling at arrival**. Consequence recorded in §B and flagged as the top open design question for the auditor.

## §B — Problem

`internal/hook/instructions_loaded.go` `checkCharacterBudget` (line ~91-110) enforces only the per-file 40,000-character limit (`const charBudget = 40000`, line 103). No aggregate metric exists: the instruction files loaded together at session start (the `CLAUDE.md` import chain plus the always-loaded rules) can each pass the per-file check while their sum grows without bound. t1303 (09-29) measured the aggregate floor by hand at 205,977 chars; the figure is card-recorded only and nothing in the tree mechanizes it.

Two structural gaps:

1. **No aggregate metric.** The hook checks one file per InstructionsLoaded event; nothing sums the set, so per-file compliance masks session-level growth. `rule-loading-budget.md` already obligates recording "the measured count and byte total of always-loaded files … and the runtime `InstructionsLoaded`/token observation" for every change to `.claude/rules/moai/` — with no metric in the tree, that obligation has no mechanical backing.
2. **No session-level constant.** The only budget in the file is per-file. The operator ruled 210,000 chars (2026-10-01) as the session-level figure; it exists nowhere in code or doctrine.

**Arrival-state fact (measured, this plan run):** the derived set on tree `2868afc0b` aggregates 214,155 chars > 210,000. Once implemented, the advisory fires from the first InstructionsLoaded event. This SPEC mechanizes the metric and records the ruling; it does NOT diet the surface (out of scope, §F) — the arriving advisory is the metric doing its job and is the pressure that produces a diet card.

## §C — Goal

The InstructionsLoaded handler computes an aggregate character metric over a **mechanically derived** instruction-file set (never a hand-written path list), compares it against a named `sessionCharBudget` constant carrying the operator's 210,000 ruling, and surfaces an advisory SystemMessage on breach — with regression tests that catch hand-list mutants, wrong-set mutants, and boundary mutants.

## §D — Requirements (GEARS)

### D.1 — The derived instruction-file set (normative definition)

Given the write-side project root R (`resolveProjectRoot(input)`, the existing helper), the **instruction-file set** S(R) is derived mechanically at metric time:

1. `R/CLAUDE.md` — the anchor;
2. the transitive `@`-import closure of the anchor: every line matching `^@` followed by a repo-relative path resolves under R and is itself scanned for imports (missing/unreadable members are skipped, §D.3);
3. every `*.md` under `R/.claude/rules/moai/` whose frontmatter — when the file starts with a `---` block — lacks a top-level `paths:` key, i.e. the always-loaded rule set (mechanical selector per `rule-authoring.md` § The surface and its four slots, slot 1).

At authoring time on tree `2868afc0b` this derivation yields **18 files** (5 import-closure members + 13 always-loaded rules) and is the same set t1303 measured.

### D.2 — Requirements

- **REQ-INSTRBUDGET-001 (aggregate metric).** The handler SHALL expose `aggregateInstructionChars(R)` = Σ `utf8.RuneCount(file)` over S(R). Rune count is the metric because the existing per-file check at line 100 uses `utf8.RuneCount` — the two budgets measure the same unit.
- **REQ-INSTRBUDGET-002 (session constant + advisory breach message).** The package SHALL define `sessionCharBudget = 210000` — the operator ruling of 2026-10-01 (card t1318, AskUserQuestion). WHEN `aggregateInstructionChars(R) > sessionCharBudget`, `Handle` SHALL return a `HookOutput` whose `SystemMessage` names the aggregate figure, the budget, and the derived file count. The message SHALL be advisory — it SHALL NOT block the event, set a non-zero exit path, or suppress the per-file checks.
- **REQ-INSTRBUDGET-003 (tolerance).** Unreadable or missing set members SHALL be skipped silently-and-continue (the per-file precedent, `instructions_loaded.go:94`). An empty derivation SHALL skip the aggregate comparison entirely.
- **REQ-INSTRBUDGET-004 (mechanical derivation — no hand list).** The set SHALL NOT be a hand-written enumeration of the 18 paths. A mutant that hardcodes the list MUST be caught: adding a new always-loaded rule file (or a new `@`-import) to a fixture MUST change the metric; adding a `paths:`-scoped rule MUST NOT.
- **REQ-INSTRBUDGET-005 (regression tests).** The package SHALL carry the test trio of plan.md §F M3: (a) derivation semantics on fixtures, (b) aggregate movement — growing one fixture member by N runes moves the aggregate by exactly N, (c) budget boundary — a fixture set grown past 210,000 produces the SystemMessage, one under it produces none.

## §E — Constraints

- Hook handlers run under the ≤5s event budget (`internal/hook/CLAUDE.md`); the aggregate adds ≤ ~20 small-file reads (~227 KB total on this tree) — constant cost, inside budget. Advisory-check discipline (`coding-standards.md` § Advisory-Check Discipline) is satisfied by constant cost + the skip-on-error tolerance of REQ-INSTRBUDGET-003.
- The existing per-file check and the CLAUDE.md fallback path are preserved byte-for-byte in behavior — this SPEC is additive.
- Code comments and identifiers in English (language.yaml `code_comments: en`).
- `sessionCharBudget` lives file-locally adjacent to `charBudget` (decision D2 in plan.md §D); re-homing it to `internal/config/defaults.go` is an auditor-callable change, not a plan-phase act.

## §F — Out of Scope

### Out of Scope — surface diet

- Dieting the always-loaded surface to fit under 210,000 (the arrival state already exceeds it — §B). That is a separate card; this SPEC only makes the overshoot observable.

### Out of Scope — gate semantics

- Making the aggregate check blocking (exit-2 or gate semantics).
- Config-file plumbing of the budget (`.moai/config/sections/*.yaml`) — the card scopes a Go constant.

### Out of Scope — metric-surface extension and doctrine

- Extending the metric to the other always-loaded slots of `rule-authoring.md` (output styles, `MEMORY.md` head) — the InstructionsLoaded hook measures instruction files; the 18-file set is the instruction-file subset. Boundary recorded in §G.
- Changes to `rule-loading-budget.md` or `rule-authoring.md` doctrine.

## §G — Known Relations and Non-Goals

- `coding-standards.md` § File Size Limits states the 40k per-file budget applies to "every instruction file the InstructionsLoaded hook measures — always-loaded and `paths:`-scoped alike". This SPEC does not contradict that: the per-file check keeps measuring whatever file arrives per event; the aggregate's derived set is the session-start set (t1303's 18). A `paths:`-scoped rule that loads mid-session counts toward the per-file check and not toward the session-start aggregate — the two scopes are different questions and both stay measured.
- `rule-loading-budget.md` (the measurement obligation for `.claude/rules/moai/` changes) gains its mechanical backing from REQ-INSTRBUDGET-001; neither file is edited by this SPEC.
- `verification-completeness.md` §2 governs the two-cell AC discipline in acceptance.md — three release-blocking criteria with measured RED-now cells, one regression-guard criterion for the arrival figure.

## §H — Cross-References

- Card t1318 (lead-issued 09-29, Tier M, Class C) — dispatch record and operator ruling.
- t1303 (09-29) — original 18-file/205,977-char hand measurement; report absent from checkout (card-recorded figure).
- `internal/hook/instructions_loaded.go` — the measured implementation site.
- `.claude/rules/moai/workflow/rule-loading-budget.md` — the measurement obligation this SPEC mechanizes.
- `.claude/rules/moai/development/rule-authoring.md` — the always-loaded surface definition (mechanical `paths:` selector).
