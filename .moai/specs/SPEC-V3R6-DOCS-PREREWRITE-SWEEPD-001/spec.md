---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001
title: "Docs Pre-Rewrite Cleanup — stale CLI citations on three pages (child SWEEPD, card t1724)"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content/{ko,en,ja,zh}/{cost-optimization,guides,worktree}"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, stale-citation, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001 — Stale CLI citations on three pages (child SWEEPD)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1724 (class C), child SWEEPD of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (card t1655, split 8 of 11, operator approval `d-20261014T060405Z-0fcd`). Every count below was measured on tree commit 74ad197380 (74ad197380a0711cdb9816d4e37c94d29af730ea) in this pass. The card names the t1655 `child-cards.md` as its source; that file is not on disk (acceptance.md G1), so the E-01, E-02, and E-03 hit sets were derived from the twelve files themselves and cross-checked against the umbrella's committed acceptance counts. Status `draft`. |

**Tier note.** Tier M: the 12 files sit in the 5 to 15 file band of the tier table; the sweep is expected to change 8 of them, one line per edit, 36 changed lines in all, far under 300; the umbrella assigned tier M to each child (umbrella spec.md, § Child SPECs). This SPEC carries 5 requirements and 6 acceptance criteria, under the Tier M ceilings of 16 and 16. No design.md or research.md is authored.

## Overview (WHY)

The umbrella requires that every page under `docs-site/content/` cite only `moai` verbs and `/moai` slash commands that resolve on the tree (REQ-PRC-001). The umbrella was split into eleven children; this child owns the page part of that requirement for three pages in four locales: `cost-optimization/prompt-caching.md`, `guides/mcp-server.md`, and `worktree/_index.md`. The umbrella's patterns are called E-01 (absent verbs), E-02 (absent sub-verbs), and E-03 (Mermaid labels `moai run` and `moai sync`); they are defined once under Scope.

Measured state on tree 74ad197380 (full evidence in acceptance.md §C; per-line inventory in plan.md §A.2):

- **`prompt-caching.md`**: one E-01 match per locale, 4 lines (en 268, ja 171, ko 243, zh 171), and nothing from E-02 or E-03. Each is the sentence "`moai cg` has been retired ... It is not an alias for `moai cc`". The installed CLI confirms it: `moai cg` exits 1 with `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, and `moai cc` is a listed verb. These four are **Class N** lines (they state that a verb is retired). The E-01 set for this file is therefore **not empty, and its Class C subset is empty**: nothing in the file tells a reader to run an absent command.
- **`mcp-server.md`**: seven lines per locale, 28 in all, all in the "CLI equivalent" column of the tool catalog, all in the same rows in every locale (lines 130, 163, 182 to 186). One is E-01 (`moai audit cache`, line 163) and six are E-02 (`moai spec list`, `moai codex task`, `moai codex setup`, `moai codex job status|result|cancel`). The installed CLI rejects or does not list each of them. These are **Class C** lines: the cell tells the reader which command to run.
- **`worktree/_index.md`**: two E-03 lines per locale, 8 in all: the Mermaid labels `B3["moai run SPEC-ID"]` and `B4["moai sync SPEC-ID"]` (en 91 to 92, ja 87 to 88, zh 96 to 97, ko 119 to 120). `moai run` and `moai sync` are not verbs (`Unknown command`); `/moai run` and `/moai sync` are registered slash commands. These are **Class C** lines. The same file carries five further Mermaid labels per locale of the same kind that no family pattern matches (`moai plan`, `moai run<br/>`, `moai sync<br/>`); whether they join the sweep is an open row (Q-SWD-2).

## Scope (WHAT)

Files this child may change (12): `docs-site/content/{ko,en,ja,zh}/{cost-optimization/prompt-caching,guides/mcp-server,worktree/_index}.md`. With the proposed defaults of decision-index.md, 8 of them change (`mcp-server.md` and `worktree/_index.md` in four locales) and the four `prompt-caching.md` files do not.

Terms, defined once here and used without restatement in the other artifacts:

- **E-01 pattern**: the umbrella's family F1, read by `git grep -E` as `(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)`.
- **E-02 pattern**: the umbrella's family F2, read as `(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)`.
- **E-03 pattern**: the umbrella's family F3, read as `(^|[^/])moai (run|sync) SPEC-ID`.
- **Class C line**: a match on a line that instructs the reader to run the matched form, or names it as the way to run something.
- **Class N line**: a match on a line that names the matched form only to say it is retired or does not exist.

Dispositions are owned by decision rows in decision-index.md: Q13 (delete or replace a Class C citation, per form), Q17 (all four locales in one change), Q-SWD-1 (whether the four Class N lines are in scope), and Q-SWD-2 (whether the five adjacent Mermaid labels per locale are in scope). **Q13 governs the 36 Class C lines only** (28 in `mcp-server.md`, 8 in `worktree/_index.md`); it does not govern the four Class N lines, whose disposition is Q-SWD-1. This SPEC fixes no disposition, and the gate on those rows is stated once, in decision-index.md (Gate).

**Edit units.** A hit is not edited alone; each class has a fixed edit unit, listed per page in plan.md §A.3:

- `mcp-server.md`, per locale: the last cell ("CLI equivalent") of each of seven rows, which is one line per row, 7 lines. The row's other cells and every other row stay.
- `worktree/_index.md`, per locale: the two Mermaid lines that hold `B3` and `B4`, replaced in place or removed together.
- `prompt-caching.md`: no edit unit under the default of Q-SWD-1.

## Requirements (GEARS)

Each requirement below is a part of the umbrella's REQ-PRC-001 (page part). The umbrella identifier is not renumbered; the four-segment spelling is the lint-collectible form of the same requirement.

- REQ-PRC-001-001: In each locale of `guides/mcp-server.md`, the "CLI equivalent" cell of each of the seven tool rows `spec_progress`, `audit_cache`, `codex_task`, `codex_setup`, `codex_job_status`, `codex_job_result`, and `codex_job_cancel` **shall** be replaced in place, by the no-CLI marker `—` or by the registered form that the operator's verdict on Q13 names for that row, so that no line of the four files matches the E-01 or E-02 pattern, every row stays in the table, and no other line changes.
- REQ-PRC-001-002: In each locale of `worktree/_index.md`, the two Mermaid labels `B3["moai run SPEC-ID"]` and `B4["moai sync SPEC-ID"]` **shall** be replaced in place by the registered slash forms or removed together, as the operator's verdict on Q13 directs, so that no line of the four files matches the E-03 pattern and no `B3` or `B4` reference is left dangling, and **shall** touch no other line.
- REQ-PRC-001-003: **Where** the operator's verdict on Q-SWD-1 retains the four retirement statements (the proposed default), the sweep **shall** leave the four `prompt-caching.md` files unchanged, so that the four Class N lines are the only E-01 matches the sweep admits; **where** the verdict treats them as stale citations, the SPEC returns to the plan phase before any edit.
- REQ-PRC-001-004: **Where** the operator's verdict on Q17 is that the edits land in all four locales in one change, each page that the sweep changes **shall** be changed in all four locales in that one change, with the same disposition for the corresponding edit unit in ko, en, ja, and zh.
- REQ-PRC-001-005: The sweep **shall** modify no file outside the eight files named in REQ-PRC-001-001 and REQ-PRC-001-002 (apart from this SPEC's own records and the card's evidence), and within those files **shall** modify only the edit units named under Scope, so that a page cannot be emptied, the five adjacent Mermaid labels per locale stay untouched, and lines owned by other cards on the same pages stay untouched.

## Dependencies and sequencing

- No `depends_on`. The umbrella states that the children land independently, in any order.
- **Start rule.** The umbrella (plan.md §B) says a child starts only when each decision row it depends on carries an operator verdict, and that a planner default is not a verdict. This child has an operand for Q13 (36 Class C lines) and for Q17 (the same edits), so both gate it. It also owns Q-SWD-1 and Q-SWD-2, both open. None of the four rows carries a verdict, so none counts as answered here. The gate set is stated once in decision-index.md (Gate).
- **Shared ko `prompt-caching.md`.** The umbrella (spec.md lines 122 and 126) records that the README child (t1720) shares ko files with this child on different lines and recommends landing this child after it. The card text calls that child "card 4"; this pass did not resolve that name to a card id (acceptance.md G5). Under the default of Q-SWD-1 this SPEC edits no `prompt-caching.md` line, so no line of the four files can conflict with another card's edit; the only duty it creates is to re-pin the base commit of AC-PRC-001-005 and AC-PRC-001-006 if the other card lands first and is absorbed (plan.md M4).
- The closing measurement (`cited-verbs: 0`) is the umbrella's M6 and uses the GATE child's citation checker (card t1717, SPEC-V3R6-DOCS-PREREWRITE-GATE-001). The lane's brief records that card as on plan-audit hold; this pass did not re-measure its state. This child's own criteria do not need the checker.
- Line numbers on these pages are moving coordinates (plan.md §C, R1). Every check identifies a hit by content.

## Out of Scope

Each item is named so that a later change cannot claim it silently.

### Out of Scope — other pages and other forms

- Pages outside the twelve files: the README child (card t1720) and the SWEEPA to SWEEPC, SWEEPE, and SWEEPF children (cards t1721 to t1723, t1725, t1727) own theirs.
- Any `moai` form on these three pages that the umbrella families E-01 to E-03 do not match. The registered forms on `mcp-server.md` were probed (plan.md §A.4); other forms on `worktree/_index.md` and `prompt-caching.md` were not (acceptance.md G3).

### Out of Scope — the retirement statements and the adjacent labels

- The four `prompt-caching.md` retirement statements, under the default of Q-SWD-1. A verdict that edits them returns the SPEC to the plan phase.
- The five adjacent Mermaid labels per locale on `worktree/_index.md` (plan.md §A.2), under the default of Q-SWD-2. A verdict that edits them returns the SPEC to the plan phase.

### Out of Scope — CLI and Go code

- Any change under `internal/` or `cmd/`, including registering `moai codex task`, `moai codex job`, `moai audit`, or `moai spec list`, or adding help text.

### Out of Scope — closing measurement and new prose

- The umbrella's closing measurement (M6) and the GATE citation checker. If Q-SWD-1 is answered "retain", the umbrella's closing check (E-01 exits 1 with empty stdout) must admit the four retained lines, an amendment outside this SPEC (plan.md R4).
- New explanatory prose and new translation beyond the mechanical edit that each disposition needs.
