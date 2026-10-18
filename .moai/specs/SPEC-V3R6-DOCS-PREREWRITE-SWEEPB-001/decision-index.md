# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001

Revision 0.2.0 (2026-10-19). The three `DEFAULT-APPLIED` stamps of revision 0.1.0 are withdrawn: a planner default is not an operator verdict (umbrella plan.md §B; the sibling SWEEPA withdrew its identical stamps after plan-audit iteration 1). Q13 and Q17 are open rows with an empty `Operator verdict:`; each keeps its rule-ranked `Default:` as a proposal and its options. Q-SWB-1 carries the ruling on board `d-20261018T201957Z-c26e`.

Revision 0.1.0 (2026-10-19). This child carries copies of the umbrella rows Q13 and Q17 with the umbrella's numbering unchanged, and owns the new row Q-SWB-1. The umbrella is SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001; its decision-index.md (revision 0.3.1) holds stubs for Q13 and Q17, and the source text of both rows is at commit bdb7b764 of that file. `interview.decision_gate` is `on`, so this file is part of the artifact set and every row routes with one of the four labels. The new row uses the prefix `Q-SWB-` so it cannot collide with the umbrella's Q1 to Q24 or with another child's new rows.

Labels: DECIDED and POLICY-COVERED need an authority anchor in the committed tree (product.md, a completed SPEC's HISTORY or Amendments row, `.moai/config/sections/*.yaml`, or the constitution). None of the rows below has one, so every row is FOUNDER. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the authority register; the rows cite it as a fact, not as authority. No row carries a recommendation. A `Default:` is the option the published Default rule ranks first, a policy application and a proposal, not a verdict.

**Gate (stated once; every other file points here by row id).** Q-SWB-1 is decided: option (b), retain, on board record `d-20261018T201957Z-c26e`. Consequently no milestone edits a page, and no tracked file outside this SPEC directory is edited. Q13 and Q17 stay open rows with empty verdicts and block nothing, because no milestone consumes them. Fact, not a verdict: Q13 governs Class C lines and there are none at the pin (plan.md §A.2); Q17 governs edits and there are none under (b). The run-phase work (plan.md M1 to M3) is evidence recording and regression-guard checks only.

Board check: Q-SWB-1 is answered by the record cited in the Gate. `moai decision read` was not queried in this pass for Q13 or Q17 (acceptance.md gap G7), so no claim is made about them.

Evidence references (E-SWB-n) point to acceptance.md §C, measured at tree commit 2bcce5d9dc.

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: carried from the umbrella, whose count of distinct stale forms is 18 (umbrella revision 0.3.0, defect B12). Some forms have a current equivalent and some have none; replacing a citation changes what a reader is told to run. For this child the row has no operand at the pin: the twelve E01 lines are Class N (E-SWB-5, plan.md §A.2). It applies only to a Class C line found at run start.
Requirement: REQ-PRC-001 (page part): REQ-PRC-001-001
Class: implementation-level
Child: SWEEPB (SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001), decision row Q13; the umbrella row is shared with README and SWEEPA, SWEEPC to SWEEPF, each for its own files
Default: delete the stale citation (rule: smaller user-visible surface)
Alternate: replace the citation with the current equivalent, where the operator names one for each form
Scope: Q13 governs Class C lines only. It does not govern the twelve Class N lines; their disposition is Q-SWB-1.
Operator verdict:

### Q17: Do the same edits land in en, ja, and zh in the same change? (cross-cutting; SWEEPB copy)
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, as the umbrella cites it (umbrella E-24); that document is outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: all three pages exist in all four locales and carry one E01 match per file (E-SWB-2), so a candidate change is a same-disposition edit in twelve files. Whether the obligation binds this child's edits is the operator's call, since the document is not an authority. The repository parity gate `bash scripts/docs-i18n-check.sh` was not run in this pass (gap G6).
Requirement: REQ-PRC-001-003
Class: implementation-level
Child: cross-cutting (README, REDIR, SWEEPA to SWEEPF, FLAG); SWEEPB copy
Default: edit all four locales in one change (rule: preserves current behavior; the four locales carry the same twelve logical citations today)
Alternate: edit the ko copy in this change and the en, ja, and zh copies in a later change
Operator verdict:

### Q-SWB-1: Do the twelve retirement statements count as stale citations?
Label: FOUNDER
Authority anchor: none in the committed tree. The umbrella retained the same notice on the four `cg-mode.md` pages (its E-01a), but that is text in its own acceptance.md and the umbrella SPEC has status `draft`, so it cannot support DECIDED. The umbrella's decided row Q24 (only narrative examples that carry an explicit example marker are exempt) points the other way for unmarked narrative; Q24 is scoped to the SWEEPE child's two slash-form lines and does not name notices.
Why unresolved: all twelve E01 matches in this child's files state that a verb is retired or part of history; none tells the reader to run it. `tokenomics-overview.md` and `multi-llm/_index.md` (four locales each) carry the `moai cg` retirement notice; `model-policy.md` (four locales) carries the clause that the `moai model profile` accessor existed through v3.1 and was retired. The installed CLI confirms each: `moai cg` exits 1 with `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, `moai model` exits 1 with `Unknown command` (E-SWB-6). The umbrella counted the twelve lines in its E-01 total but examined only the `cg-mode.md` copy, so the card's partial criterion (zero E01 matches on this child's files) cannot hold while the lines stay. The options differ in what a reader loses, in what the umbrella's closing check must admit (GATE exemption C-7 names none of these lines), and, for the Korean tokenomics and index lines, in the overlap with README REQ-PRC-004 (spec.md Dependencies).
Requirement: REQ-PRC-001-002
Class: implementation-level. Justification: the lines are documentation text; no command, default, or template changes. If the operator regards deleting the retirement notice as removing documented user-facing behavior, this row is reclassified product-level (fail-closed).
Child: SWEEPB (SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001), new row; the umbrella and the other children carry no such row (SWEEPA carries the analogous row Q-SWA-1 for its own files)
Options: (a) treat the twelve lines as stale citations, with one of two sub-dispositions for the edit units of plan.md §A.2: (a1) reword each unit so it drops the E01 form and keeps its information; (a2) delete each unit, whose cost is that deleting the tokenomics notice leaves the migration paragraphs that follow it under their heading without their lead sentence (plan.md R2); (b) treat them as retained retirement notices on the umbrella's E-01a precedent: leave the twelve files byte-unchanged and list the twelve lines as the only E01 matches this child admits, which also requires the umbrella's closing check and GATE's exemption list to admit them, an amendment outside this SPEC (plan.md R4).
Default: option (b), retain (rule: preserves current behavior)
Alternate: option (a1), reword
Operator verdict: (b) retain — ruled on board d-20261018T201957Z-c26e

## Rows

| Row | Class | Default | Alternate | Operator verdict |
|-----|-------|---------|-----------|------------------|
| Q13 | implementation-level | delete (smaller surface) | replace | empty (open) |
| Q17 | implementation-level | all four locales in one change | ko first, rest later | empty (open) |
| Q-SWB-1 | implementation-level | (b) retain | (a1) reword | (b) retain — ruled on board d-20261018T201957Z-c26e |

The state of the rows is in the Gate above.
