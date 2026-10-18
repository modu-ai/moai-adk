---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001
title: "Docs Pre-Rewrite Cleanup — stale CLI citations on three advanced pages (child SWEEPA, card t1721)"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content/{ko,en,ja,zh}/advanced"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, stale-citation, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001 — Stale CLI citations on three advanced pages (child SWEEPA)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1721 (class C), child SWEEPA of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (card t1655, split 5 of 11, operator approval `d-20261014T060405Z-0fcd`). Every count below was measured on tree commit 2bcce5d9dc (2bcce5d9dc4c5f57bb98bcfa553e0061fa373681) in this pass. The card names `split-proposal.md` and `child-cards.md` of the t1655 tree as its sources; neither is on disk (acceptance.md G1), so the file list and the E-01 pattern come from the umbrella's committed spec, acceptance, and decision-index instead. Status `draft`. |

**Tier note.** Tier M: the 12 files sit in the 5 to 15 file band of the tier table; each edit is one line or one clause, far under 300 changed lines; the umbrella assigned tier M to each child (umbrella spec.md, § Child SPECs). This SPEC carries 4 requirements and 5 acceptance criteria, under the Tier M ceilings of 16 and 16. No design.md or research.md is authored.

## Overview (WHY)

The umbrella requires that every page under `docs-site/content/` cite only `moai` verbs and `/moai` slash commands that resolve on the tree (REQ-PRC-001). The umbrella was split into eleven children; this child owns the page part of that requirement for three advanced pages in four locales: `bas-navigator.md`, `factory-mode.md`, and `profile-matrix.md`. The umbrella's pattern for absent CLI verbs is called E-01 (defined once under Scope).

Measured state on tree 2bcce5d9dc (full evidence in acceptance.md §C; per-line inventory in plan.md §A.2):

- All 12 files carry exactly one E-01 match: 12 matches in 12 files. The other three umbrella families (E-02 absent sub-verbs, E-03 Mermaid labels, E-04 unregistered slash forms) match nothing in these files.
- Four matches tell the reader to run an absent command: the `moai codemaps` line in the `bas-navigator.md` command block, line 75 in all four locales. The installed CLI rejects it. These are **Class C** lines.
- Eight matches state that a verb is retired or absent: the `moai cg` / `moai gpt` row of the "Refused forms" table in `factory-mode.md` (line 56 in all four locales) and the `moai model profile` accessor clause in `profile-matrix.md` (en 95, ko 88, ja 53, zh 53). The installed CLI confirms each statement. These are **Class N** lines. The umbrella excluded the equivalent notice in the four `multi-llm/cg-mode.md` pages from E-01 (its E-01a) but counted these eight lines, so whether a Class N line is a stale citation is an open decision row (Q-SWA-1).
- Two other cards edit the same `factory-mode.md`; § Dependencies and sequencing records both.

## Scope (WHAT)

Files this child may change (12): `docs-site/content/{ko,en,ja,zh}/advanced/{bas-navigator,factory-mode,profile-matrix}.md`.

Terms, defined once here and used without restatement in the other artifacts:

- **E-01 pattern**: the umbrella's family F1 pattern, read by `git grep -E` as `(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)`.
- **Class C line**: an E-01 match on a line that instructs the reader to run the matched form.
- **Class N line**: an E-01 match on a line that names the matched form only to say it is retired or does not exist.

Dispositions are owned by decision rows in decision-index.md: Q13 (delete or replace a citation), Q17 (all four locales in one change), and Q-SWA-1 (whether Class N lines are in scope). This SPEC fixes no disposition.

## Requirements (GEARS)

Each requirement below is a part of the umbrella's REQ-PRC-001 (page part). The umbrella identifier is not renumbered; the four-segment spelling is the lint-collectible form of the same requirement.

- REQ-PRC-001-001: The four Class C lines (the `moai codemaps` command in the `bas-navigator.md` command block of each locale) **shall** each be deleted or replaced as the operator's verdict on Q13 directs, so that no line of the four `bas-navigator.md` files matches the E-01 pattern.
- REQ-PRC-001-002: **Where** the operator's verdict on Q-SWA-1 treats the eight Class N lines as stale citations, the sweep **shall** delete or reword each of them as Q13 directs, so that no line of the eight `factory-mode.md` and `profile-matrix.md` files matches the E-01 pattern; **where** the verdict treats them as retained retirement notices, the sweep **shall** leave the eight lines unchanged, and they **shall** be the only E-01 matches in the twelve files.
- REQ-PRC-001-003: **Where** the operator's verdict on Q17 is that the edits land in all four locales in one change, the sweep **shall** change each of the three pages in all four locales in that one change and **shall** apply the same disposition to the corresponding line in ko, en, ja, and zh.
- REQ-PRC-001-004: The sweep **shall** modify no file outside the twelve listed files (apart from this SPEC's own records and the card's evidence), and **shall** modify no line of those files other than the lines inventoried in plan.md §A.2, so that lines owned by other cards on the same pages stay untouched.

## Dependencies and sequencing

- No `depends_on`. The umbrella states that the children land independently, in any order.
- The closing measurement (`cited-verbs: 0`) is the umbrella's M6 and uses the GATE child's citation checker (card t1717, SPEC-V3R6-DOCS-PREREWRITE-GATE-001). The lane's brief records that card as on plan-audit hold; this pass did not re-measure its state. This child's own criteria do not need the checker.
- **Shared file with card t1863** (commit 84f9387535, SPEC-FACTORY-LEADER-ADOPT-CORE-001 M4, branch WT-10-18-t1849 per the lane's brief). It adds a six-line section to `factory-mode.md` in all four locales. The en hunk header is `@@ -57,0 +58,6 @@`: a pure insertion after the blank line that follows the "Refused forms" table, so it removes nothing and does not touch the row this child may edit. Observed on 2026-10-18: 84f9387535 is an ancestor of neither local `main` nor this tree's HEAD (acceptance.md E-SWA-8). The ranges are disjoint but adjacent, and the claim is re-verified at run time (AC-PRC-001-005).
- **Shared file with card t1726** (child FLAG). It owns the `--factory-run` and `--clear-policy` citations in `factory-mode.md`; this child changes none of those lines (AC-PRC-001-005).
- The README child (card t1720) shares no file with this child.
- Line numbers on these pages are moving coordinates (plan.md §C, R1). Every check identifies a hit by content.

## Out of Scope

Each item is named so that a later change cannot claim it silently.

### Out of Scope — other pages and other forms

- Pages outside the twelve files: the README child (card t1720) and the SWEEPB to SWEEPF children (cards t1722 to t1725, t1727) own theirs.
- Any `moai` form on these three pages that the four umbrella families do not match. This pass did not verify such forms (acceptance.md G3).

### Out of Scope — flags and other cards' text

- The `--factory-run` and `--clear-policy` citations (umbrella Q18, card t1726) and the `factory-mode.md` lines that carry them.
- The six-line section that card t1863 adds to `factory-mode.md`.

### Out of Scope — CLI and Go code

- Any change under `internal/` or `cmd/`, including registering `moai codemaps`, un-hiding `moai navigator-sync`, or adding help text.

### Out of Scope — closing measurement and new prose

- The umbrella's closing measurement (M6) and the GATE citation checker.
- New explanatory prose and new translation beyond the mechanical edit that each disposition needs.
