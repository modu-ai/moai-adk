# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001

Revision 0.2.0 (2026-10-19). This child carries copies of the umbrella rows Q13 and Q17 with the umbrella's numbering unchanged, and owns the new rows Q-SWD-1 and Q-SWD-2. The umbrella is SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001; its decision-index.md holds stubs for Q13 and Q17. `interview.decision_gate` is `on` (`.moai/config/sections/interview.yaml`, line 6), so this file is part of the artifact set and every row routes with one of the four labels. The new rows use the prefix `Q-SWD-` so they cannot collide with the umbrella's Q1 to Q24 or with another child's new rows.

Labels: DECIDED and POLICY-COVERED need an authority anchor in the committed tree (product.md, a completed SPEC's HISTORY or Amendments row, `.moai/config/sections/*.yaml`, or the constitution). None of the rows below has one, so every row is FOUNDER. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the authority register; the rows cite it as a fact, not as authority. The two board records cited below are not committed artifacts, so they are recorded as the verdict or ruling of the row, not as an authority anchor. No row carries a recommendation. Every row's `Operator verdict:` was empty at authoring (revision 0.1.0). Q13 and Q17 were then filled from board record `d-20261019T015908Z-220a` (the operator's verdict), and Q13's reading, Q-SWD-1, and Q-SWD-2 from board record `d-20261019T021509Z-4604` (a leader ruling, which is not an operator verdict); see Board check.

**Gate (stated once; every other file points here by row id).** The Gate holds nothing open. All four rows carry a recorded answer: Q13 and Q17 carry the operator verdict of board record `d-20261019T015908Z-220a`; the reading of Q13 that applies to `mcp-server.md` and the rows Q-SWD-1 and Q-SWD-2 carry the leader ruling of board record `d-20261019T021509Z-4604`, which is the leader's interpretation of the operator ruling 220a and not a new operator verdict. This closes the decision part of M1 in plan.md. It does not satisfy the plan audit or the plan→run Kickoff gate, which are separate and still apply; until the Kickoff gate is met, no tracked file outside this SPEC directory is edited. A planner default is not a verdict (umbrella plan.md §B), and no row below rests on one.

Board check (2026-10-19, revision 0.2.0, acceptance.md E-SWD-17): `moai decision read --scope standing` and `moai decision read --scope card:t1724` were both read in this pass, exit 0 each, and the two records below were located by id in the saved output. At the read of revision 0.1.0 no record answered these rows; the two records were reported afterwards. The standing record, copied verbatim (header line, then body):

```
decision record: decided_by=operator (AskUserQuestion 10-19) evidence_refs=.moai/worktrees/t1723/.moai/reports/t1723/premise-check.md (lane-31), SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 decision-index ladder_path=operator verdict on umbrella decision-index rows Q13/Q17 id=d-20261019T015908Z-220a scope=standing kind=ruling
  body: Operator verdicts for the docs pre-rewrite family (umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001): Q13 (stale citations to removed commands/flags/pages) = replace with the current equivalent where one exists, delete the line where none exists. Q17 (locales) = all four locales (ko·en·ja·zh) change in the same change/card. Applies to every child depending on these rows (t1721 SWEEPA, t1723 SWEEPC, t1724 SWEEPD, t1725 SWEEPE, t1726 FLAG, t1727 SWEEPF); each child cites this record id in its decision-index Operator verdict field. Q-SWA-1 and other rows remain open.
```

The card-scoped record, copied verbatim:

```
decision record: decided_by=leader (session d1aa9eb6) under operator delegation; c26e precedent evidence_refs=lane-41 plan commit d64904ad69 ladder_path=⑥ leader judgment id=d-20261019T021509Z-4604 scope=card:t1724 kind=ruling
  body: t1724: the lane's Q13 reading is adopted — in guides/mcp-server.md the stale CLI-equivalent cell becomes the no-CLI marker and the table row stays (the row documents a registered MCP tool); this is the leader reading of operator verdict 220a, not a new operator verdict. Q-SWD-1 = retain the four prompt-caching retirement lines (accurate retirement notice, c26e). Q-SWD-2 = leave the adjacent Mermaid labels (outside the card's stale-citation scope).
```

The record 4604 names the `mcp-server.md` cell reading and the two Q-SWD rows. It does not name the Mermaid labels; for those the slash forms `/moai run SPEC-ID` and `/moai sync SPEC-ID` exist on the tree (plan.md §A.4), so the replace half of the operator verdict 220a applies, and the leader confirmed branch R (replace in place) in the revision brief of this pass. That confirmation is an instruction to the planner, not a board record.

Evidence references (E-SWD-n) point to acceptance.md §C, measured at tree commit 26834dd17d (plan.md §A.1) unless an entry names its own pin.

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: carried from the umbrella, whose count of distinct stale forms is 18 (umbrella revision 0.3.0, defect B12). Some forms have a current equivalent and some have none; replacing a citation changes what a reader is told to run. For this child Q13 covers the 36 Class C lines in eight files: in `mcp-server.md` the seven "CLI equivalent" cells (`moai spec list`, `moai audit cache`, `moai codex task`, `moai codex setup`, `moai codex job status|result|cancel`), and in `worktree/_index.md` the two Mermaid labels (`moai run SPEC-ID`, `moai sync SPEC-ID`), each in four locales (E-SWD-2, E-SWD-4). The per-form evidence, including the two registered commands that overlap only in part, is in plan.md §A.5 and is not repeated here.
Requirement: REQ-PRC-001 (page part): REQ-PRC-001-001, REQ-PRC-001-002
Class: implementation-level
Child: SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), decision row Q13; the umbrella row is shared with README and the other SWEEP children, each for its own files
Default: per form. For each of the seven `mcp-server.md` cells, delete the stale citation and keep the row by setting the cell to the page's no-CLI marker `—` alone (rule: smaller user-visible surface; the bare marker is already the whole cell of 18 rows per locale, E-SWD-6). For the two Mermaid labels, replace them with the registered slash forms `/moai run SPEC-ID` and `/moai sync SPEC-ID` (rule: preserves current behavior; the diagram keeps its four steps and the page body already uses these forms)
Alternate: delete the table row of a stale cell, which is the other reading of "delete the line"; or remove the two Mermaid label lines together. Neither is planned: a ruling for either returns the SPEC to the plan phase, because REQ-PRC-001-001 keeps every row and REQ-PRC-001-002 keeps the diagram's four steps (acceptance.md §A.3)
Scope: Q13 governs the 36 Class C lines only. It does not govern the four Class N lines in `prompt-caching.md`; their disposition is Q-SWD-1.
Operator verdict: "Q13 (stale citations to removed commands/flags/pages) = replace with the current equivalent where one exists, delete the line where none exists." Source: board record d-20261019T015908Z-220a, decided_by=operator via AskUserQuestion, 10-19, kind=ruling, scope=standing, umbrella rows Q13/Q17.
Leader ruling (not an operator verdict): board record d-20261019T021509Z-4604, decided_by=leader (session d1aa9eb6) under operator delegation, c26e precedent, kind=ruling, scope=card:t1724: "the lane's Q13 reading is adopted — in guides/mcp-server.md the stale CLI-equivalent cell becomes the no-CLI marker and the table row stays (the row documents a registered MCP tool); this is the leader reading of operator verdict 220a, not a new operator verdict."
Applied reading: this is the leader's interpretation of the operator ruling 220a, and it is not a new operator verdict. "Delete the line" is applied to the stale citation inside its `mcp-server.md` cell, which leaves the cell as the bare marker `—`; the table row is kept, because each row documents a registered MCP tool whose only stale part is the CLI-equivalent cell. The two registered commands that overlap in part (plan.md §A.5) are not "the current equivalent": the permitted replacement for each of the seven cells, in each locale, is the marker `—` alone. For `worktree/_index.md` the slash forms exist, so the replace branch applies. A ruling to delete the table rows instead returns the SPEC to the plan phase.

### Q17: Do the same edits land in en, ja, and zh in the same change? (cross-cutting; SWEEPD copy)
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, as the umbrella cites it; that document was not re-read in this pass and is outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: for this child the question has one live shape. `mcp-server.md` and `worktree/_index.md` exist in all four locales and carry the same lines in each (E-SWD-2, E-SWD-4), so the candidate change is a same-disposition edit in eight files. The parity baseline holds today: `bash scripts/docs-i18n-check.sh` on tree 74ad197380 exits 0 and prints `Errors:   0` and `Warnings: 0` (E-SWD-14). The card text states the four-locale rule ("all four locales (ko, en, ja, zh) are fixed in the same PR (Q17)"); whether the obligation binds this child's edits was the operator's call, since the document is not an authority, and the operator answered it (below).
Requirement: REQ-PRC-001-004
Class: implementation-level
Child: cross-cutting (README, REDIR, SWEEPA to SWEEPF, FLAG); SWEEPD copy
Default: edit all four locales in one change (rule: preserves current behavior; the four locales are in parity today)
Alternate: edit the ko copy in this change and the en, ja, and zh copies in a later change. This SPEC does not plan that branch: AC-PRC-001-001 to AC-PRC-001-004 read four locales, and a verdict for the alternate returns the SPEC to the plan phase
Operator verdict: "Q17 (locales) = all four locales (ko·en·ja·zh) change in the same change/card." Source: board record d-20261019T015908Z-220a, decided_by=operator via AskUserQuestion, 10-19, kind=ruling, scope=standing, umbrella rows Q13/Q17.

### Q-SWD-1: Do the four `prompt-caching.md` retirement statements count as stale citations?
Label: FOUNDER
Authority anchor: none in the committed tree. The retain rulings of the sibling children SWEEPA (Q-SWA-1) and SWEEPB (Q-SWB-1, board record `d-20261018T201957Z-c26e`) are read from their decision-indexes in other worktrees; they decide those children's lines, are not a completed SPEC's HISTORY or Amendments row, and were not queried on the board in this pass (acceptance.md G2), so they cannot support DECIDED.
Why unresolved: the four E-01 matches in `prompt-caching.md` (en 268, ja 171, ko 243, zh 171) are the sentence "`moai cg` has been retired ... It is not an alias for `moai cc`", which states that a verb is retired and does not tell the reader to run it. The installed CLI confirms it: `moai cg` exits 1 with `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, and `moai cc` is a listed verb (E-SWD-10). The file therefore has E-01 matches but no Class C line, and no E-02 or E-03 match (E-SWD-12). The umbrella's partial criterion for this child, zero E-01 matches on its file list, cannot hold while these four lines stay. The options differ in what a reader loses and in what the umbrella's closing check must admit.
Requirement: REQ-PRC-001-003
Class: implementation-level. Justification: the lines are documentation text; no command, default, or template changes. If the operator regards deleting a retirement notice as removing documented user-facing behavior, this row is reclassified product-level (fail-closed).
Child: SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), new row; the sibling children carry the corresponding rows Q-SWA-1 and Q-SWB-1
Options: (a) treat the four lines as stale citations and reword or delete them, which requires a new edit unit and a plan revision because this SPEC plans no edit to `prompt-caching.md`; (b) treat them as retained retirement notices: leave the four files unchanged and list the four lines by file and content as the only E-01 matches this child admits, which also requires the umbrella's closing check (E-01 exits 1 with empty stdout) and the GATE checker to admit the same four lines, an amendment outside this SPEC (plan.md R4).
Default: (b) retain (rule: preserves current behavior)
Alternate: (a) reword or delete; not planned, returns the SPEC to the plan phase
Operator verdict: none. Leader ruling (not an operator verdict): board record d-20261019T021509Z-4604, decided_by=leader (session d1aa9eb6) under operator delegation, c26e precedent, kind=ruling, scope=card:t1724: "Q-SWD-1 = retain the four prompt-caching retirement lines (accurate retirement notice, c26e)." This selects option (b).

### Q-SWD-2: Do the five adjacent Mermaid labels per locale join the sweep?
Label: FOUNDER
Authority anchor: none in the committed tree.
Why unresolved: besides the two labels the card names, each `worktree/_index.md` carries five Mermaid labels that name an unregistered form but match none of the umbrella families, because they have no `SPEC-ID`: `"moai plan<br/>..."` (one), `"moai plan"` (one), `"moai run<br/>..."` (two), and `"moai sync<br/>..."` (one), 20 lines in all (plan.md §A.2, E-SWD-13). `moai run` and `moai sync` are `Unknown command`; `moai plan` is a registered verb that renders a plan report and is described in its own help as distinct from the `/moai plan` skill (E-SWD-10). The card text names only the `SPEC-ID` pair. If the 20 lines stay, the swept page shows the corrected pair next to uncorrected labels of the same kind (plan.md R7), and the umbrella's closing measurement, which keys on the family patterns, will not see them.
Requirement: REQ-PRC-001-002, REQ-PRC-001-005
Class: implementation-level
Child: SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), new row
Default: leave the 20 labels unchanged (rule: preserves current behavior; the card's scope names only the two `SPEC-ID` lines)
Alternate: extend the edit unit of `worktree/_index.md` to the five labels, with the same registered-slash-form disposition as Q13; not planned, returns the SPEC to the plan phase
Operator verdict: none. Leader ruling (not an operator verdict): board record d-20261019T021509Z-4604, decided_by=leader (session d1aa9eb6) under operator delegation, c26e precedent, kind=ruling, scope=card:t1724: "Q-SWD-2 = leave the adjacent Mermaid labels (outside the card's stale-citation scope)." This selects the default.

## Verdict state of each row

| Row | Class | Default | Recorded answer |
|-----|-------|---------|-----------------|
| Q13 | implementation-level | per form: bare `—` for the seven CLI cells, slash form for the two Mermaid labels | operator verdict (board 220a, 10-19): replace where a current equivalent exists, else delete the line; leader reading (board 4604, not an operator verdict): the cell becomes the marker and the row stays |
| Q17 | implementation-level | all four locales in one change | operator verdict (board 220a, 10-19): all four locales in the same change/card |
| Q-SWD-1 | implementation-level | (b) retain | leader ruling (board 4604, not an operator verdict): retain the four lines |
| Q-SWD-2 | implementation-level | leave the 20 labels | leader ruling (board 4604, not an operator verdict): leave the adjacent labels |

All four rows are in the Gate set above, and each carries either the operator verdict or the recorded leader ruling. No row is open.
