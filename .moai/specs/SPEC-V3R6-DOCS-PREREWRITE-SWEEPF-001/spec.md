---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001
title: "Docs Pre-Rewrite Cleanup — workflow-commands index sweep: stale design-family slash citations in 3 pages x 4 locales (card t1727)"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content/{en,ko,ja,zh}/workflow-commands"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, slash-citations, workflow-commands, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, SPEC-V3R6-DOCS-PREREWRITE-GATE-001, SPEC-DOCS-SITE-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 — workflow-commands index sweep (child 11 of 11)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for factory card t1727 (Class C, docs only), the SWEEPF child of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (split ruling `d-20261014T060405Z-0fcd`; routing `d-20261014T114834Z-e039`). It owns the page part of REQ-PRC-001 and AC-PRC-001 for the twelve files named in § Scope, under the parent numbers, and adds REQ-SWF-002 to -007 and AC-SWF-002 to -006 for what the Q13, Q14, and Q17 answers need. Every measurement was taken on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-workflow-cmd-index`, equal to local `main`); nothing is carried over from the parent ledger. **Operator verdicts applied** (board record `d-20261019T015908Z-220a`, decided by the operator through the question channel on 2026-10-19, scope standing): Q13 = replace with the current equivalent where one exists, delete where none exists; Q17 = all four locales change in the same change. Both rows leave the open set. **Q14 stays open** (the `moai-design` page), and its marker option carries a conditional amendment of REQ-PRC-001 (REQ-SWF-006). **Found while measuring, not in the card text:** (1) `docs-site/data/menu/main.yaml` lines 204 to 208 carries the same `/moai design` name map in four locales plus a `ref:` to the page, and `docs-site/.locale-parity-baseline` line 51 lists the page; both are outside the twelve files (open items O-2, O-3); (2) the Korean `_index.md` carries two design rows and a `/moai goal` summary row that the other locales lack (23 `/moai` table rows against 21), so the locales are not line-for-line twins; (3) no slash command or CLI verb enters the design phase (premise of Q13's "current equivalent", measured in acceptance.md C-7). Status stays `draft`. |

**Tier note.** Tier M: twelve files edited, or eight edited and four deleted if Q14 removes the page (inside the 5 to 15 band); 7 requirements and 6 acceptance criteria (ceilings 16 and 16). Artifact set: spec.md, plan.md, acceptance.md, plus `decision-index.md` (the decision gate is on in this repository) and `progress.md`. No `design.md` or `research.md`.

## Overview (WHY)

The three `workflow-commands` index pages carry citations of a slash command that does not exist. Measured on tree `74ad197380` (evidence ids refer to acceptance.md §C):

- Twelve files exist: `_index.md`, `_meta.yaml`, and `moai-design.md` under `docs-site/content/{en,ko,ja,zh}/workflow-commands/` (C-3).
- 27 lines in the twelve files match `/moai design` or `/moai:design` (C-1). The same files hold no other unresolved `/moai <verb>` form: a negative-list search over the 17 registered commands (C-8) returns the same 27 lines and nothing else (C-2). Every other slash citation in the twelve files resolves.
- `/moai design` is not a registered slash command (C-8), and the workflow it names carries `user-invocable: false` (C-9). No file under `.claude/skills/moai/` that routes plan, run, or sync names the design workflow as an entry (C-7), so no current equivalent that a reader can type exists.
- The locales are not twins: `_index.md` has 1 design line in en, ja, and zh and 2 in ko; the page has 5 lines in en and ko and 4 in ja and zh (C-1); the Korean `_index.md` has 23 `/moai` table rows against 21 elsewhere (C-4).
- The page is linked from the four `_index.md` files and from the sidebar data file (C-5, which includes two paths outside the twelve files).

The two operator verdicts of record `d-20261019T015908Z-220a` fix what happens to the citations: where a current equivalent exists, the citation is replaced; where none exists, it is deleted; all four locales change together. What happens to the page itself (Q14) is not decided.

## Scope (WHAT)

The files this child changes. Paths are repository-relative; `<L>` is each of `en`, `ko`, `ja`, `zh`.

| # | Path | Change | Waits on |
|---|------|--------|----------|
| 1-4 | `docs-site/content/<L>/workflow-commands/_index.md` | delete the design table rows (en, ja, zh: 1 line each; ko: 2 lines), per the Q13 verdict | nothing |
| 5-8 | `docs-site/content/<L>/workflow-commands/_meta.yaml` | remove or retitle the `moai-design` entry (key plus `title: "/moai design"`) | Q14 |
| 9-12 | `docs-site/content/<L>/workflow-commands/moai-design.md` | delete, rewrite, or mark, per the Q14 answer | Q14 |

If Q14 removes the page, rows 9 to 12 are deleted and the remaining set is eight files. No other file is changed. The count sentences on `_index.md` ("All subcommands (16)" and the sentence below it) are not slash citations; their handling is open item O-4.

## Definitions

- **Design family.** The text `/moai design` or `/moai:design` followed by a character that is not alphanumeric or a hyphen (the pattern of acceptance.md C-1).
- **Current equivalent.** A slash command registered under `.claude/commands/moai/` or a `moai` CLI verb that does what the citation says it does. For the design family none exists (C-7, C-8, C-9), so the Q13 verdict resolves to deletion for these citations. The determination is measured, not assumed, and is re-run at M1.
- **Pin.** Unless a cell states another commit, every tree-attributed observation is on `74ad197380a0711cdb9816d4e37c94d29af730ea`.

## Requirements (GEARS)

REQ-PRC-001 keeps the parent's number (page part). The other requirements are this child's own.

### Citations resolve (parent REQ-PRC-001, page part)

- **REQ-PRC-001** (Ubiquitous): Each of the twelve files **shall** cite only `/moai` slash commands whose verb resolves to a file under `.claude/commands/moai/`, except where REQ-SWF-006 applies.

### Verdict-driven handling (Q13, Q17 — operator verdicts of record `d-20261019T015908Z-220a`)

- **REQ-SWF-002** (Ubiquitous): **Where** a stale citation in the twelve files has a current equivalent, the citation **shall** be replaced by that equivalent; **where** none exists, the line holding the citation **shall** be deleted. For the design family no equivalent exists (C-7), so each table row that cites it **shall** be deleted and **shall not** be replaced by a prose sentence that names the same command.
- **REQ-SWF-003** (Ubiquitous): The edits of REQ-PRC-001 and REQ-SWF-002 **shall** be made in all four locales (ko, en, ja, zh) in the same change, and after the change the four locales **shall** carry the same set of deletions: the per-locale count of `/moai` table rows in `_index.md` **shall** fall by the number of design rows that locale held.

### Page handling (Q14 — open; exactly one of REQ-SWF-004, -005, -006 applies once the row has a verdict)

- **REQ-SWF-004** (Ubiquitous, conditional on Q14 = remove the page): **Where** Q14 records that the `moai-design` page is removed, the four page files **shall** be deleted, the `moai-design` entry **shall** be removed from each `_meta.yaml`, and no file under `docs-site/content/` **shall** keep a link to the removed page.
- **REQ-SWF-005** (Ubiquitous, conditional on Q14 = rewrite the page): **Where** Q14 records that the page is rewritten, the four page files **shall** describe the design phase as the conditional route entered by the orchestrator after plan-audit PASS and Implementation Kickoff Approval, **shall** present no `/moai design` or `/moai:design` form, and **shall** carry a `title:` without a slash form; the `moai-design` entry of each `_meta.yaml` **shall** be retitled to match.
- **REQ-SWF-006** (Ubiquitous, conditional on Q14 = internal-only marker; **amendment of REQ-PRC-001**): **Where** Q14 records the internal-only marker, REQ-PRC-001 **shall** admit the design family on a `moai-design` page line only when that line carries the marker token recorded in progress.md §E.2, and the page **shall** state in its opening paragraph that the workflow cannot be typed as a slash command; every other file of the twelve **shall** stay free of the design family.

### Preservation

- **REQ-SWF-007** (Unwanted): The change **shall not** modify any file other than the twelve of § Scope, and in particular **shall not** modify the other pages of `workflow-commands/` (`moai-goal`, `moai-harness`, `moai-plan`, `moai-project`, `moai-run`, `moai-sync`) in any locale.

## Acceptance

The Given-When-Then criteria AC-PRC-001 and AC-SWF-002 to -006, their RED-now and green-path cells, and the evidence ledger are in `acceptance.md`.

## Out of Scope

Each item is named so that a later change cannot claim it silently. Scope of the other ten children lives in their own SPECs.

### Out of Scope — files outside the twelve

- `docs-site/data/menu/main.yaml` (lines 204 to 208: the `/moai design` name map in four locales and `ref: /workflow-commands/moai-design`). It is the structure curator's single-writer file and the card does not list it. It carries the same stale citation; whether to add it to this change is open item O-2.
- `docs-site/.locale-parity-baseline` line 51 (`workflow-commands/moai-design.md`). Under Q14 = remove, the entry names a page that no longer exists; its owner is not identified in the tree (open item O-3).
- The other six pages of `workflow-commands/` in any locale, and every other page of `docs-site/content/`.

### Out of Scope — other stale citations and flags

- Stale citations outside the twelve files (SWEEPA to SWEEPE, README), and flag citations (`--factory-run`, `--clear-policy`; FLAG child, Q18).
- Redirects in `docs-site/vercel.json` that mention `/design` (REDIR child, Q15).

### Out of Scope — checker, gates, and code

- The cited-verbs checker and the Hugo and link-check jobs (GATE child, SPEC-V3R6-DOCS-PREREWRITE-GATE-001, card t1717). This child does not depend on them and does not build them; they will measure these files later.
- Any change under `internal/`, `cmd/`, or `.claude/` (including the design workflow skill and its `user-invocable` flag), and any change to the count sentences of `_index.md` unless open item O-4 is answered.
- Pushing, opening a pull request, or any remote action.

## Dependencies and Parallel Work

- No `depends_on`. This child is implementable without the GATE checker, without the parent's M6, and without any sibling. The GATE checker (card t1717) and the parent's closing run M6 will later measure these files; they read this child's committed criteria by reference (parent plan §B.3) and cannot block it.
- **Non-overlapping siblings.** README (t1720), SWEEPA (t1721), SWEEPB (t1722), SWEEPC (t1723), SWEEPD (t1724), SWEEPE (t1725), FLAG (t1726), RESID (t1719), REDIR (t1718), and GATE (t1717) list no file under `docs-site/content/*/workflow-commands/` in the parent's child table. This was read from the parent table; the siblings' own file lists are not in this tree and were not measured.
- **Q24 marker.** The example marker belongs to SWEEPE. REQ-SWF-006 needs a marker token for the Q14 marker option; whether it reuses the Q24 token or names its own is part of the Q14 answer, and the GATE checker's exemption list must carry whatever is chosen (acceptance.md G-6).
- **Single writer of `_meta.yaml`.** The four `_meta.yaml` edits belong to the structure curator; see plan.md §A.
