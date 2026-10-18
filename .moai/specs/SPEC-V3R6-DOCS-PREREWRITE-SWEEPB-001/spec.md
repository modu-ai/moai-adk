---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001
title: "Docs Pre-Rewrite Cleanup — stale CLI citations on tokenomics-overview and two multi-llm pages (child SWEEPB, card t1722)"
version: "0.4.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content/{ko,en,ja,zh}/{advanced,multi-llm}"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, stale-citation, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001 — Stale CLI citations on tokenomics-overview and two multi-llm pages (child SWEEPB)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1722 (class C), child SWEEPB of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (card t1655, split 6 of 11, operator approval `d-20261014T060405Z-0fcd`). Every count below was measured on tree commit 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 in this pass. The structure follows the sibling child SWEEPA (card t1721) after its plan-audit iteration 1 (FAIL 0.69): the five blocking defects of that report are designed out here (edit unit defined once in plan.md §A.2; gate state stated once in decision-index.md; Class C and Class N separated; no requirement defers to a decision a stamp already makes; a byte-unchanged and hunk-containment criterion closes the emptied-page mutant). Status `draft`. |
| 2026-10-19 | 0.2.0 | manager-spec | Revision before the plan audit, on the leader's instruction and the sibling's iteration-1 result (SWEEPA, FAIL 0.69, defects D2 and D3/D4). Withdrawn: the three `DEFAULT-APPLIED` stamps on Q13, Q17, and Q-SWB-1, because a planner default is not an operator verdict (umbrella plan.md §B). Q-SWB-1 is then decided (b) retain, on board record `d-20261018T201957Z-c26e`, so the requirements, milestones, and criteria are rewritten for zero page changes and the option (a) path is removed from them; Q13 and Q17 stay open with empty verdicts and no operand. The gate is stated once, in decision-index.md (Gate), with the milestone-to-row map; the revisability and stamp-effect paragraphs are removed, and the other files cite the Gate by row id. The measurement and the controls are unchanged; the criteria are reduced to regression-guards for zero page changes. Requirement count 4 and criterion count 5. Status stays `draft`. |
| 2026-10-19 | 0.3.0 | manager-spec | Revision after plan-audit iteration 1 (FAIL 0.78, `.moai/reports/t1722/plan-audit.md`, audited 2227269ae0). D1: REQ-PRC-001-002 and -003 are rewritten as contribution obligations (this card's own diff touches none of the twelve files); the "byte-identical to the pin" and "keep its inventoried citation" wording is deleted; the absolute-count checks are removed from the criteria; the README child's removal of the two Korean lines is stated not to be a regression. D2: every criterion's control now runs its own command form on a named historical input. D3: the umbrella start rule is reconciled in decision-index.md (Gate) by board record `d-20261018T223420Z-7f98`, a judgment that the rule applies only to rows with an operand; it is not a verdict on Q13 or Q17. O1 to O4 and O6: restatements turned into pointers, decider named on the Q-SWB-1 verdict, option (a) leftovers removed, AC-001 given a positive control over a HEAD grep, the fact clause taken out of REQ-PRC-001-001. Requirement count 4, criterion count 5. Status stays `draft`. |
| 2026-10-19 | 0.4.0 | manager-spec | Revision after plan-audit iteration 2 (FAIL 0.82, `.moai/reports/t1722/plan-audit-iter2.md`, audited 61e857c67f). D4 and D5, one reading (option A of the report, the lane's recorded judgment): REQ-PRC-001-001 is a contribution obligation, "this card adds no Class C line to any of the twelve files", and AC-PRC-001-002 decides it, because that criterion fails on any change to the twelve files whatever the changed line says. AC-PRC-001-001 is re-classed as a run-start inventory record on the HEAD tree: it decides no requirement and gates no change absorbed from `main`. Its substring exemption is removed, and the negative control moves to AC-PRC-001-002 on a scratch fixture that carries a Class C line containing an exempt substring (acceptance.md E-SWB-10a). D6: gap G7 is narrowed to the Q13 and Q17 verdict records, and its Q-SWB-1 wording now agrees with the Gate. O7 and O8: a revision note in the decision-index.md header, and "by this card" on the two sentences that keep the retained lines. Requirement count 4, criterion entries 5 (four deciding criteria and one inventory record). Status stays `draft`. |

**Tier note.** Tier M: the 12 files sit in the 5 to 15 file band; no page is edited under the ruling; the umbrella assigned tier M to each child (umbrella spec.md, § Child SPECs). This SPEC carries 4 requirements and 5 acceptance entries (four deciding criteria and one inventory record), under the Tier M ceilings of 16 and 16. No design.md or research.md is authored.

## Overview (WHY)

The umbrella requires that every page under `docs-site/content/` cite only `moai` verbs and `/moai` slash commands that resolve on the tree (REQ-PRC-001). The umbrella was split into eleven children; this child owns the page part of that requirement for three pages in four locales: `advanced/tokenomics-overview.md`, `multi-llm/_index.md`, and `multi-llm/model-policy.md`.

Measured state on tree 2bcce5d9dc (evidence in acceptance.md §C; per-line inventory in plan.md §A.2):

- All 12 files carry exactly one match of the umbrella's E-01 pattern (defined once under Scope): 12 matches in 12 files. The umbrella families E-02 (absent sub-verbs), E-03 (Mermaid labels), and E-04 (unregistered slash forms) match nothing in these files.
- **No match is a Class C line.** None of the 12 lines tells the reader to run an absent verb. Eight lines (`tokenomics-overview.md` and `multi-llm/_index.md`, four locales each) are the "`moai cg` has been retired" notice, a copy of the notice the umbrella retained on the four `multi-llm/cg-mode.md` pages (its E-01a). Four lines (`model-policy.md`, four locales) are a clause in the history section that says the `moai model profile` accessor existed through v3.1 and was retired. The installed CLI confirms each statement. These twelve are **Class N** lines.
- Every other `moai` form cited on the three pages resolves (plan.md §A.3).
- The umbrella counted these twelve lines inside its 48-file E-01 figure, so the umbrella's count of "stale" files over-counts for this child. Whether an accurate retirement statement is a stale citation is the new row Q-SWB-1, decided (b) retain (decision-index.md, Gate): the pages are not edited, and the card closes as rationale recorded plus zero page changes.
- Two of the eight notice lines (Korean `tokenomics-overview.md` line 112 and Korean `multi-llm/_index.md` line 164) are the same lines the README child (card t1720, REQ-PRC-004, umbrella E-08) targets for the CG passage de-duplication. The umbrella table says the shared Korean files differ in lines; for these two lines that is not so. The conflict is between obligations, so this SPEC deliberately holds no obligation on those lines (§ Dependencies).

## Scope (WHAT)

Files this child covers (12; this card edits none of them under the ruling on Q-SWB-1): `docs-site/content/{ko,en,ja,zh}/advanced/tokenomics-overview.md` and `docs-site/content/{ko,en,ja,zh}/multi-llm/{_index,model-policy}.md`.

Terms, defined once here and used without restatement in the other artifacts:

- **E01 pattern**: the umbrella's family F1 pattern, read by `git grep -E` as `(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)`.
- **Class C line**: an E01 match on a line that instructs the reader to run the matched form.
- **Class N line**: an E01 match on a line that names the matched form only to say it is retired, absent, or part of history.
- **Edit unit**: the lines of the superseded option (a), one per page and locale, listed in plan.md §A.2 for the record. Under the ruling no unit is edited.

Dispositions are owned by decision rows in decision-index.md: Q13 (delete or replace a Class C citation), Q17 (all four locales in one change), and Q-SWB-1 (whether Class N lines are stale citations). The state of those rows is stated once, in decision-index.md (Gate).

## Requirements (GEARS)

Each requirement below is a part of the umbrella's REQ-PRC-001 (page part). The umbrella identifier is not renumbered; the four-segment spelling is the lint-collectible form of the same requirement.

- REQ-PRC-001-001: This card **shall** add no Class C line to any of the twelve files, that is, no line that instructs the reader to run an E01 or W01 match.
- REQ-PRC-001-002: This card **shall** change none of the twelve files: the card's own diff against `main` (merge-base form, `main...HEAD`) touches none of them (Q-SWB-1, decision-index.md).
- REQ-PRC-001-003: This card **shall** change no locale copy (ko, en, ja, zh) of any of the three pages, individually per page.
- REQ-PRC-001-004: This card **shall** modify no file outside the twelve listed files, apart from this SPEC's own records and the card's evidence.

These are contribution obligations: they bind this card's diff, not the content of the twelve files at any later tree. A change to those files landed on `main` by another card and absorbed here is not this card's contribution and is not a regression of this SPEC; in particular the removal of Korean `tokenomics-overview.md:112` and Korean `multi-llm/_index.md:164` by the README child's REQ-PRC-004 is not one.

## Dependencies and sequencing

- No `depends_on`. The umbrella states that the children land independently, in any order.
- **Gate.** The state of Q13, Q17, and Q-SWB-1 and the reconciliation with the umbrella start rule are in decision-index.md (Gate); this file points there.
- **README child (card t1720, REQ-PRC-004).** Its RED-now cell (umbrella E-08) lists Korean `tokenomics-overview.md` and Korean `multi-llm/_index.md` among the pages whose CG retirement sentence is to be reduced to one canonical page. At the pin the sentence is on lines 112 and 164 (acceptance.md E-SWB-8). The umbrella's "different lines" statement does not hold for these two lines. The obligations conflict (README keeps the passage on one Korean page only), so this SPEC deliberately holds no obligation on those lines; their removal by the README child is not a regression of this SPEC. The README child has no SPEC on disk, so its edit set was not read.
- **Other children.** SWEEPA (card t1721, three `advanced` pages) shares no file with this child (acceptance.md E-SWB-8). The GATE and REDIR specs on disk name none of the twelve files. SWEEPC to SWEEPF and FLAG have no SPEC on disk; their file sets were not read (gap G4); plan.md M3 re-checks at run time.
- **Closing measurement.** The umbrella's M6 (`cited-verbs: 0`) uses the GATE child's citation checker. GATE's exemption entry C-7 admits the four `cg-mode.md` pages and the four `moai-todo.md` pages by name, and no entry names these twelve lines. The checker would report these lines. That is outside this card and was recorded as a follow-up of the umbrella closing card (plan.md R4); this child's own criteria do not need the checker.
- Line numbers on these pages are moving coordinates (plan.md §C, R1). Every check identifies a hit by content.

## Out of Scope

Each item is named so that a later change cannot claim it silently.

### Out of Scope — other pages and other forms

- Pages outside the twelve files: the README child (card t1720), SWEEPA (card t1721), and SWEEPC to SWEEPF (cards t1723 to t1725, t1727) own theirs.
- Any `moai` form on these three pages that the four umbrella families do not match. Resolution of such forms was probed (plan.md §A.3), but no edit is proposed for them.

### Out of Scope — CG passage de-duplication

- Choosing which Korean page keeps the CG retirement passage and reducing the other copies (umbrella Q1, REQ-PRC-004, README child). This child never decides that.

### Out of Scope — CLI and Go code

- Any change under `internal/` or `cmd/`, including registering `moai model`, `moai gpt`, or un-retiring `moai cg`.

### Out of Scope — closing measurement and new prose

- The umbrella's closing measurement (M6) and the GATE citation checker, including its exemption list.
- Any edit of the twelve pages, and any new prose or translation. Under the ruling on Q-SWB-1 this card leaves the retired-notice and history lines as written; that is not an obligation on a later change by another card, such as the README child's removal of the two Korean lines.
