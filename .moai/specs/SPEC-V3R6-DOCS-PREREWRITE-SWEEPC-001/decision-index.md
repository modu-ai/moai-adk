# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001

Revision 0.1.0 (2026-10-19). This child carries copies of the umbrella rows Q13 and Q17 with the umbrella's numbering unchanged, and owns the new row Q-SWC-1. The umbrella is SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001; its decision-index.md (revision 0.3.1) holds stubs for Q13 and Q17, and the source text of both rows is at commit bdb7b764 of that file. `interview.decision_gate` is `on` (`.moai/config/sections/interview.yaml`), so this file is part of the artifact set and every row routes with one of the four labels. The new row uses the prefix `Q-SWC-` so it cannot collide with the umbrella's Q1 to Q24 or with another child's new rows. `interview.recommendation_mode` is `pull`: no row below carries a recommendation.

Labels: DECIDED and POLICY-COVERED need an authority anchor in the committed tree (product.md, a completed SPEC's HISTORY or Amendments row, `.moai/config/sections/*.yaml`, or the constitution). None of the rows below has one, so every row is FOUNDER. A `Default:` line is the option the published Default rule ranks first: a policy application and a proposal, never a verdict. No `DEFAULT-APPLIED` stamp is written anywhere in this SPEC.

**Gate (stated once; every other file points here by row id).**

- Q13 and Q17 carry an operator verdict: board record `d-20261019T015908Z-220a`, `decided_by=operator (AskUserQuestion 10-19)`, `scope=standing`, `kind=ruling`, timestamp 2026-10-19T01:59:08Z (the time the id carries). The record names t1723 SWEEPC among the children it applies to, and says each child cites it in its Operator verdict field. Both rows are met for this child.
- Q-SWC-1 is open (empty verdict). The record states "Q-SWA-1 and other rows remain open"; it does not answer Q-SWC-1, which asks the same question for this child's lines.
- Milestone map: plan.md M1 (close the gate) is done when Q-SWC-1 carries an operator verdict. M2 (the edit) needs Q-SWC-1 = (a) with Q13 and Q17 as recorded; under Q-SWC-1 = (b) M2 makes no edit and Q13 and Q17 have no operand. No tracked file outside this SPEC directory is edited before Q-SWC-1 carries a verdict.
- The umbrella start rule (umbrella plan.md §B: a child starts only when each decision row it depends on carries an operator verdict, attributed to the operator with a UTC timestamp) is therefore met for Q13 and Q17 and unmet for Q-SWC-1.
- Open item OI-1 (not a decision row). Q13's verdict is "replace with the current equivalent where one exists, delete the line where none exists". The record does not define "equivalent". spec.md Scope defines the per-line test T-eq, and plan.md §A.2 records its result for each of the twelve lines: for the form `moai cg` the equivalent is `moai migrate cg`, because the tree's own diagnostic names it and the tree registers it. Whether the operator accepts a migration-preview command as the equivalent of a retired launcher is not decidable from the committed tree. If the operator reads "equivalent" more narrowly, every line becomes deletable; only the result column of plan.md §A.2 and the added-column expectation of acceptance.md AC-PRC-001-004 change, and no requirement text does.

Board check (2026-10-19): `moai decision read` returned the board with exit 0; record `d-20261019T015908Z-220a` is present with the body quoted in the rows below. A search for `t1723`, `SWEEPC`, and `Q-SWC` found no other record, so no record answers Q-SWC-1 (acceptance.md E-SWC-11).

Evidence references (E-SWC-n) point to acceptance.md §C, measured at tree commit 74ad197380.

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: no longer unresolved for this child; the verdict is recorded below. Carried from the umbrella, whose count of distinct stale forms is 18 (umbrella revision 0.3.0, defect B12). For this child the row has an operand only when Q-SWC-1 treats the twelve notice lines as stale citations; the twelve lines and their T-eq result are in plan.md §A.2.
Requirement: REQ-PRC-001 (page part): REQ-PRC-001-003
Class: implementation-level
Child: SWEEPC (SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001), decision row Q13; the umbrella row is shared with README, SWEEPA, SWEEPB, SWEEPD to SWEEPF, each for its own files
Default: delete the stale citation (rule: smaller user-visible surface); superseded by the verdict, which decides per line
Alternate: replace the citation with the current equivalent
Operator verdict: Q13 (stale citations to removed commands/flags/pages) = replace with the current equivalent where one exists, delete the line where none exists. Operator, via AskUserQuestion 10-19; board record d-20261019T015908Z-220a, 2026-10-19T01:59:08Z (verbatim from the record body).

### Q17: Do the same edits land in en, ja, and zh in the same change? (cross-cutting; SWEEPC copy)
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, as the umbrella cites it; that document is outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: no longer unresolved for this child; the verdict is recorded below. All three pages exist in all four locales and carry one W01 match per file (E-SWC-3), so a same-disposition edit spans twelve files, or fewer where the README child has already made a Korean file compliant (spec.md Dependencies).
Requirement: REQ-PRC-001-004
Class: implementation-level
Child: cross-cutting (README, REDIR, SWEEPA to SWEEPF, FLAG); SWEEPC copy
Default: edit all four locales in one change (rule: preserves current behavior); matches the verdict
Alternate: edit the ko copy in this change and the en, ja, and zh copies in a later change
Operator verdict: Q17 (locales) = all four locales (ko·en·ja·zh) change in the same change/card. Operator, via AskUserQuestion 10-19; board record d-20261019T015908Z-220a, 2026-10-19T01:59:08Z (verbatim from the record body).

### Q-SWC-1: Do the twelve retirement statements count as stale citations?
Label: FOUNDER
Authority anchor: none in the committed tree. The umbrella retained the same sentence on the four `multi-llm/cg-mode.md` pages (its E-01a), but that is text in its own acceptance.md and the umbrella SPEC has status `draft`, so it cannot support DECIDED. The sibling rulings on the same question (board `d-20261018T201957Z-c26e`, scope `card:t1722`, option retain; the t1721 ceiling-exception record `d-20261018T202324Z-cb47`, scope `card:t1721`) name other cards; they are precedent, not a verdict for this row, and are not in the authority register.
Why unresolved: all twelve W01 matches in this child's files state that a verb is retired; none tells the reader to run it (plan.md §A.2). The installed CLI and the tree confirm the statement: `moai cg` exits 1 with `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, and `moai migrate cg` is registered (E-SWC-6). The umbrella counted the twelve lines in its E-01 total but examined only the `cg-mode.md` copy. The card's stated partial criterion (zero E-01 matches on this child's file list) cannot hold while the lines stay. Q13's verdict names "stale citations to removed commands"; whether an accurate retirement notice is one is this row's question, and the record that carries Q13 says other rows remain open. The options differ in what a reader loses, in what the umbrella's closing check and the GATE checker must admit (plan.md R4), and, for the three Korean lines, in the overlap with the README child.
Requirement: REQ-PRC-001-001, REQ-PRC-001-002
Class: implementation-level. Justification: the lines are documentation text; no command, default, or template changes. If the operator regards removing a retirement notice as removing documented user-facing behavior, this row is reclassified product-level (fail-closed).
Child: SWEEPC (SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001), new row; the umbrella and the other children carry no such row (SWEEPA carries Q-SWA-1 and SWEEPB carries Q-SWB-1 for their own files)
Options: (a) treat the twelve lines as stale citations: edit each Edit unit per Q13's verdict (replace where T-eq holds, delete where it does not), in all locales that are in S, in one change; (b) treat them as retained retirement notices on the umbrella's E-01a precedent: leave the twelve files unchanged by this card and list the twelve lines as the only W01 matches this child admits, which also requires the umbrella's closing check and the GATE checker to admit the same lines, an amendment outside this SPEC.
No Default is ranked: the published rule would keep the lines unchanged, which the card's stated criterion (zero E-01 matches) does not allow.
Operator verdict:

## Verdict state of each row

| Row | Class | Default | Operator verdict |
|-----|-------|---------|------------------|
| Q13 | implementation-level | superseded by the verdict | recorded: d-20261019T015908Z-220a, operator, 2026-10-19T01:59:08Z |
| Q17 | implementation-level | matches the verdict | recorded: d-20261019T015908Z-220a, operator, 2026-10-19T01:59:08Z |
| Q-SWC-1 | implementation-level | none ranked | empty (open) |

The meaning of the state is in the Gate above.
