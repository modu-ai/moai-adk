---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001
title: "Docs Pre-Rewrite Cleanup — stale CLI citations on cli, windows-guide, and what-is-moai-adk (child SWEEPC, card t1723)"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content/{ko,en,ja,zh}/{getting-started,core-concepts}"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, stale-citation, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001 — Stale CLI citations on three pages (child SWEEPC)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1723 (class C), child SWEEPC of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (card t1655, split 7 of 11, operator approval `d-20261014T060405Z-0fcd`). Every count below was measured on tree commit 74ad197380a0711cdb9816d4e37c94d29af730ea in this pass. The card names `child-cards.md` and `split-proposal.md` of the t1655 tree as its sources; neither is on disk (acceptance.md G1), so the file list and the E-01 pattern come from the umbrella's committed spec, acceptance, and decision-index. The operator verdicts on Q13 and Q17 (board record `d-20261019T015908Z-220a`, 2026-10-19T01:59:08Z) arrived during this pass and are cited in decision-index.md; the requirements below are written against them. The structure follows the siblings SWEEPA and SWEEPB after their plan audits: the gate and the branch classification are each stated once and cited by id. Status `draft`. |

**Tier note.** Tier M: the 12 files sit in the 5 to 15 file band of the tier table; each edit is one line, far under 300 changed lines; the umbrella assigned tier M to each child (umbrella spec.md, § Child SPECs). This SPEC carries 5 requirements and 6 acceptance criteria, under the Tier M ceilings of 16 and 16. No design.md or research.md is authored.

## Overview (WHY)

The umbrella requires that every page under `docs-site/content/` cite only `moai` verbs and `/moai` slash commands that resolve on the tree (REQ-PRC-001). The umbrella was split into eleven children; this child owns the page part of that requirement for three pages in four locales: `getting-started/cli.md`, `getting-started/windows-guide.md`, and `core-concepts/what-is-moai-adk.md`.

Measured state on tree 74ad197380 (evidence in acceptance.md §C; per-line inventory in plan.md §A.2):

- All 12 files carry exactly one match of the umbrella's E-01 pattern: 12 matches in 12 files. The other three umbrella families (E-02 absent sub-verbs, E-03 Mermaid labels, E-04 unregistered slash forms) match nothing in these files.
- **No match is a Class C line.** None of the 12 lines tells the reader to run an absent verb. All 12 are the one sentence "`moai cg` has been retired. It exits with a migration diagnostic ..." in the four locales, the same sentence the umbrella retained on the four `multi-llm/cg-mode.md` pages (its E-01a) but counted here. They are **Class N** lines. The installed CLI and the tree agree with the sentence: `moai cg` exits 1 with "moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration", and `moai migrate cg` is registered.
- Whether an accurate retirement statement is a stale citation is the new row Q-SWC-1, open. The sibling SWEEPB received a ruling on the same question for its own twelve lines (board `d-20261018T201957Z-c26e`, scope `card:t1722`); that record names t1722 only, so it answers nothing here.
- The card's relation claim ("shares ko files with child 4 (README) but different lines") does not hold for three lines: the three Korean E-01 lines are among the Korean copies of the CG sentence that the README child's REQ-PRC-004 reduces to links (umbrella E-08; acceptance.md E-SWC-8). The conflict is between obligations, handled under Dependencies.

## Scope (WHAT)

Files this child may change (12): `docs-site/content/{ko,en,ja,zh}/getting-started/{cli,windows-guide}.md` and `docs-site/content/{ko,en,ja,zh}/core-concepts/what-is-moai-adk.md`.

Terms, defined once here and used without restatement in the other artifacts:

- **E-01 pattern**: the umbrella's family F1 pattern, read by `git grep -E` as `(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)`.
- **W01 pattern**: the E-01 pattern with `[[:space:]]+` in place of the single space, so a respelling with two spaces or a tab still matches. Every match of E-01 is a match of W01. The criteria use W01.
- **Class C line**: a W01 match on a line that instructs the reader to run the matched form. **Class N line**: a W01 match on a line that names the matched form only to say it is retired, absent, or part of history.
- **Sweep-start set S**: the files among the twelve that match W01 on `main` (merge-base form, recomputed at read time). At the pin S is all twelve. A file that another card has already made compliant is not in S.
- **Edit unit**: the one physical line that holds the W01 match, and, when that line is deleted, one adjacent blank line. Nothing else in a file may change.
- **Equivalence test T-eq**, applied per matched line, for the cited form F: the line is *replaceable* when (1) the tree's own retirement diagnostic for F names a replacement R, and (2) R is registered on the tree (a `Use:` declaration under `internal/cli` that the build lists under its parent's `--help`). Otherwise the line is *deletable*. T-eq is the planner's operationalization of "a current equivalent exists" in the operator's verdict on Q13 (open item OI-1, decision-index.md Gate). The result per line is recorded once, in plan.md §A.2.

Dispositions are owned by decision rows in decision-index.md: Q13 and Q17 (verdicts recorded by board record `d-20261019T015908Z-220a`) and Q-SWC-1 (open). The state of the rows and the milestone each one gates are stated once, in decision-index.md (Gate).

## Requirements (GEARS)

Each requirement below is a part of the umbrella's REQ-PRC-001 (page part). The umbrella identifier is not renumbered; the four-segment spelling is the lint-collectible form of the same requirement. REQ-PRC-001-002 is the only requirement that applies under a retained verdict; the other four apply when the lines are treated as stale citations (decision-index.md, Gate). Which requirement applies under which answer of Q-SWC-1 is stated in acceptance.md §B (Applicability) and nowhere else.

- REQ-PRC-001-001: **Where** the operator's verdict on Q-SWC-1 treats the twelve notice lines as stale citations, the sweep **shall** edit the Edit unit of every file of the sweep-start set S, so that at the sweep tip no line of the twelve files matches W01.
- REQ-PRC-001-002: **Where** the operator's verdict on Q-SWC-1 treats the twelve notice lines as retained retirement notices, the sweep **shall** change none of the twelve files, measured on the card's own diff against the merge-base with `main`.
- REQ-PRC-001-003: **When** the sweep edits the line that matches W01 in one of the twelve files, the sweep **shall** replace that line with wording that drops the W01 form and cites the replacement R where the equivalence test T-eq holds for it, and **shall** delete that line where T-eq does not hold (operator verdict on Q13, record `d-20261019T015908Z-220a`).
- REQ-PRC-001-004: **Where** a page has a copy in S in more than one locale, the sweep **shall** edit every locale copy of that page that is in S in the one change that carries the card (operator verdict on Q17, record `d-20261019T015908Z-220a`), with the disposition T-eq gives for that line, so that a page edited in some of its locales only is a violation.
- REQ-PRC-001-005: The sweep **shall** modify no file outside the twelve listed files (apart from this SPEC's own records and the card's evidence), and within those files **shall** modify only the Edit unit, so that a page cannot be emptied or truncated.

These are contribution obligations except where REQ-PRC-001-001 states an end state: they bind this card's diff, not the content of the twelve files at any later tree. A change another card lands in those files and this branch absorbs is not this card's contribution and is not a regression of this SPEC. In particular, the README child's replacement of the three Korean E-01 lines by links (REQ-PRC-004 of the umbrella) is not one: it can only remove files from S.

## Dependencies and sequencing

- No `depends_on`. The umbrella states that the children land independently, in any order; its child map recommends landing after the README child (card t1720), not required.
- **Gate.** The state of Q13, Q17, and Q-SWC-1 and the milestone each gates are in decision-index.md (Gate); this file points there.
- **README child (card t1720, REQ-PRC-004).** The README child has no SPEC on disk, so its edit set was not read (acceptance.md G2). On the umbrella's evidence it targets the Korean copies of the CG sentence on nine Korean pages (umbrella E-08), among them `ko/getting-started/cli.md`, `ko/getting-started/windows-guide.md`, and `ko/core-concepts/what-is-moai-adk.md`. The three Korean E-01 lines here are those sentences (acceptance.md E-SWC-8). Both cards may therefore edit the same Korean line, which is a textual merge conflict for whichever lands second. This SPEC resolves it through S: a Korean file the README child has already made compliant is not in S and is not edited. The obligations of REQ-PRC-001-004 follow S, not the pin.
- **Other children.** The GATE, REDIR, SWEEPA, and SWEEPB specs on disk do not name any of the twelve files in their requirements (acceptance.md E-SWC-9). The GATE link-check observations about relative links on the `getting-started/cli` pages are not E-01 matches and are outside this child. SWEEPD to SWEEPF, FLAG, and README have no SPEC on disk; their file sets were not read.
- The closing measurement (`cited-verbs: 0`) is the umbrella's M6 and uses the GATE child's citation checker. GATE's exemption entry C-7 names the four `cg-mode.md` pages and the four `moai-todo.md` pages only. Under a retained verdict on Q-SWC-1 the checker would report these twelve lines; that is outside this card (plan.md R4).
- Line numbers on these pages are moving coordinates (plan.md R1). Every check identifies a hit by content.

## Out of Scope

Each item is named so that a later change cannot claim it silently.

### Out of Scope — other pages and other forms

- Pages outside the twelve files: the README child (card t1720), SWEEPA and SWEEPB (cards t1721, t1722), and SWEEPD to SWEEPF (cards t1724, t1725, t1727) own theirs.
- Any `moai` form on these three pages that the four umbrella families do not match. Resolution of the first-verb set was probed (plan.md §A.3); no edit is proposed for forms outside the families.

### Out of Scope — CG passage de-duplication and flags

- Choosing which Korean page keeps the CG retirement passage and reducing the other copies to links (umbrella Q1, README child). This child never decides that.
- The `--factory-run` and `--clear-policy` citations (umbrella Q18, card t1726).

### Out of Scope — CLI and Go code

- Any change under `internal/` or `cmd/`, including registering `moai cg` again or adding help text.

### Out of Scope — closing measurement and new prose

- The umbrella's closing measurement (M6) and the GATE citation checker, including its exemption list.
- New explanatory prose and new translation beyond the one-line replacement that the disposition of a line needs.
