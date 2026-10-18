# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001

Revision 0.1.0 (2026-10-19). This child owns the full row Q-SWA-1 and carries copies of the umbrella rows Q13 and Q17, with the umbrella's numbering unchanged. The umbrella is SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001; its decision-index.md (revision 0.3.1) holds stubs for Q13 and Q17, and the source text of both rows is at commit bdb7b764 of that file. `interview.decision_gate` is `on` (`.moai/config/sections/interview.yaml`), so this file is part of the artifact set and every row routes with one of the four labels. The new row uses the prefix `Q-SWA-` so it cannot collide with the umbrella's Q1 to Q24 or with another child's new rows.

Labels: DECIDED and POLICY-COVERED need an authority anchor in the committed tree (product.md, a completed SPEC's HISTORY or Amendments row, `.moai/config/sections/*.yaml`, or the constitution). None of the rows below has one, so every row is FOUNDER. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the authority register; the rows cite it as a fact, not as authority. No row carries a recommendation. The `Default:` lines on Q13 and Q17 are applications of the published Default rule, not recommendations.

Board check (2026-10-18): `moai decision read --scope card:t1655` returned 134 records and `moai decision read --scope card:t1721` returned 122 (standing rulings included). A search of both outputs for `Q13` and `Q17` found no match, and a search for a known record id in each file did match, so the search was live (acceptance.md E-SWA-10). No board record answers either row.

Evidence references (E-SWA-n) point to acceptance.md §C, measured at tree commit 2bcce5d9dc.

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: carried from the umbrella, whose count of distinct stale forms is 18 (umbrella revision 0.3.0, defect B12). Some forms have a current equivalent and some have none; replacing a citation changes what a reader is told to run, and the default and the alternate apply per form. For this child the forms are `moai codemaps` (4 lines, Class C) and, only if Q-SWA-1 brings them in, `moai cg`, `moai gpt`, and `moai model profile` (8 lines, Class N). The per-form evidence, including the one registered but hidden verb that does the cited job, is in plan.md §A.5 and is not repeated here.
Requirement: REQ-PRC-001 (page part): REQ-PRC-001-001 and REQ-PRC-001-002
Class: implementation-level
Child: SWEEPA (SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001), decision row Q13; the umbrella row is shared with README and SWEEPB to SWEEPF, each for its own files
Default: delete the stale citation (rule: smaller user-visible surface)
Alternate: replace the citation with the current equivalent, where the operator names one for each form
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:38:13Z claude-sonnet-5-5 manager-spec (plan close). Scope of this application: the four Class C lines. It does not decide whether the eight Class N lines are citations; that is Q-SWA-1.

### Q17: Do the same edits land in en, ja, and zh in the same change? (cross-cutting; SWEEPA copy)
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, as the umbrella cites it; that document was not re-read in this pass and is outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: for this child the question has one live shape. All three pages exist in all four locales and carry one E-01 match per file (E-SWA-2), so the candidate change is a same-disposition edit in twelve files. The parity baseline holds today: `bash scripts/docs-i18n-check.sh` on tree 2bcce5d9dc exits 0 and prints `Errors:   0` and `Warnings: 0` (E-SWA-9). The card text states the four-locale rule ("fix all four locales in the same PR per Q17"); that is the card's premise, not a recorded verdict, because no board record names Q17 (see the board check above). Whether the obligation binds this child's edits is the operator's call, since the document is not an authority.
Class: implementation-level
Child: cross-cutting (README, REDIR, SWEEPA to SWEEPF, FLAG); SWEEPA copy
Default: edit all four locales in one change (rule: preserves current behavior; the four locales are in parity today)
Alternate: edit the ko copy in this change and the en, ja, and zh copies in a later change
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:38:13Z claude-sonnet-5-5 manager-spec (plan close)

### Q-SWA-1: Do the eight retirement statements count as stale citations?
Label: FOUNDER
Authority anchor: none in the committed tree. The umbrella's handling of the same situation is text in its own acceptance.md (E-01 and E-01a) and the umbrella SPEC has status `draft`, so it is not a completed SPEC's HISTORY or Amendments row and cannot support DECIDED.
Why unresolved: eight of the twelve E-01 matches in this child's files are statements that a verb is retired or absent, not instructions to run it: `factory-mode.md` line 56 in all four locales (the `moai cg` and `moai gpt` row of the "Refused forms" table) and `profile-matrix.md` in all four locales (en 95, ko 88, ja 53, zh 53; the clause that the `moai model profile` accessor is gone). The installed CLI confirms each statement: `moai cg` exits 1 with `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, `moai gpt` and `moai model` exit 1 with `Unknown command`, and `moai migrate cg` is registered (E-SWA-6). The umbrella excluded the four `multi-llm/cg-mode.md` pages from E-01 as a retained retirement notice (its E-01a) but did not examine these eight lines, so its E-01 count includes them. The card's partial criterion, zero E-01 matches on this child's file list, cannot hold while they stay. The options differ in what a reader loses and in what the umbrella's closing check must admit.
Requirement: REQ-PRC-001-002
Class: implementation-level. Justification: the lines are documentation text; no command, default, or template changes. If the operator regards deleting the "Refused forms" row as removing documented user-facing behavior, this row is reclassified product-level (fail-closed).
Child: SWEEPA (SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001), new row; the umbrella and the other children carry no such row
Options: (a) treat the eight lines as stale citations: apply Q13 to each, deleting the statement or rewording it so that it no longer contains the E-01 form; (b) treat them as retained retirement notices on the umbrella's E-01a precedent: leave the eight lines unchanged and list them by file and content as the only E-01 matches this child admits, which also requires the umbrella's closing check (E-01 exits 1 with empty stdout) and the GATE checker to admit the same eight lines, an amendment outside this SPEC.
No Default is ranked: the published rule would keep the eight lines unchanged, which the card's stated criterion (zero E-01 matches) does not allow.
Operator verdict:

## Verdict state of each row

| Row | Class | Verdict state | Blocks the autonomous Kickoff |
|-----|-------|---------------|-------------------------------|
| Q13 | implementation-level | DEFAULT-APPLIED (Class C lines) | no |
| Q17 | implementation-level | DEFAULT-APPLIED | no |
| Q-SWA-1 | implementation-level | empty | yes; M3 and AC-PRC-001-002 wait on it |
