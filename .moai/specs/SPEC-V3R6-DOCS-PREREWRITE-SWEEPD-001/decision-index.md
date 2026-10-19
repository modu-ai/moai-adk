# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001

Revision 0.1.0 (2026-10-19). This child carries copies of the umbrella rows Q13 and Q17 with the umbrella's numbering unchanged, and owns the new rows Q-SWD-1 and Q-SWD-2. The umbrella is SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001; its decision-index.md holds stubs for Q13 and Q17. `interview.decision_gate` is `on` (`.moai/config/sections/interview.yaml`, line 6), so this file is part of the artifact set and every row routes with one of the four labels. The new rows use the prefix `Q-SWD-` so they cannot collide with the umbrella's Q1 to Q24 or with another child's new rows.

Labels: DECIDED and POLICY-COVERED need an authority anchor in the committed tree (product.md, a completed SPEC's HISTORY or Amendments row, `.moai/config/sections/*.yaml`, or the constitution). None of the rows below has one, so every row is FOUNDER. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the authority register; the rows cite it as a fact, not as authority. No row carries a recommendation. The `Default:` lines are proposals ranked by the published Default rule, not recommendations and not verdicts; every row's `Operator verdict:` was empty at authoring, and Q13 and Q17 were filled afterwards from board record `d-20261019T015908Z-220a` (see Board check).

**Gate (stated once; every other file points here by row id).** Q13 and Q17 now carry the operator verdict recorded below (board record `d-20261019T015908Z-220a`); Q-SWD-1 and Q-SWD-2 are open. Until each of the four carries an operator verdict or an explicit leader ruling, no tracked file outside this SPEC directory is edited, which holds M2 and M3 of plan.md together. Q13 is needed by M2 and M3, Q17 by M2 and M3, and Q-SWD-1 and Q-SWD-2 by the boundary of both; the four are gated as one set so that no file states a partial order of its own. A planner default is not a verdict (umbrella plan.md §B).

Board check (2026-10-19): `moai decision read --scope card:t1724` was read and searched for `Q13`, `Q17`, and the ruling ids cited by the sibling SPECs; the one match is a record about card t1721, not an answer to these rows (acceptance.md E-SWD-17). That read found no record for these rows. Afterwards board record `d-20261019T015908Z-220a` (decided_by=operator via AskUserQuestion, 10-19, kind=ruling, scope=standing, umbrella rows Q13 and Q17) was reported by the leader and is recorded in the Q13 and Q17 rows; this pass did not re-read it from the board. No board record answers Q-SWD-1 or Q-SWD-2.

Evidence references (E-SWD-n) point to acceptance.md §C, measured at tree commit 74ad197380.

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: carried from the umbrella, whose count of distinct stale forms is 18 (umbrella revision 0.3.0, defect B12). Some forms have a current equivalent and some have none; replacing a citation changes what a reader is told to run, and the default and the alternate apply per form. For this child Q13 covers the 36 Class C lines in eight files: in `mcp-server.md` the seven "CLI equivalent" cells (`moai spec list`, `moai audit cache`, `moai codex task`, `moai codex setup`, `moai codex job status|result|cancel`), and in `worktree/_index.md` the two Mermaid labels (`moai run SPEC-ID`, `moai sync SPEC-ID`), each in four locales (E-SWD-2, E-SWD-4). The per-form evidence, including the two registered but only partly equivalent candidates (`moai spec status --list`, `moai codex status`), is in plan.md §A.5 and is not repeated here.
Requirement: REQ-PRC-001 (page part): REQ-PRC-001-001, REQ-PRC-001-002
Class: implementation-level
Child: SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), decision row Q13; the umbrella row is shared with README and the other SWEEP children, each for its own files
Default: per form. For each of the seven `mcp-server.md` cells, delete the stale citation and keep the row by setting the cell to the page's no-CLI marker `—` (rule: smaller user-visible surface; the marker is already used on the same page, E-SWD-9a). For the two Mermaid labels, replace them with the registered slash forms `/moai run SPEC-ID` and `/moai sync SPEC-ID` (rule: preserves current behavior; the diagram keeps its four steps and the page body already uses these forms)
Alternate: replace a cell with the current equivalent where the operator names one for that row (candidates: `moai spec status --list` for `spec_progress`, `moai codex status` for `codex_setup`; no candidate was found for the other five), or remove the two Mermaid label lines together (branch D of AC-PRC-001-002 and AC-PRC-001-004)
Scope: Q13 governs the 36 Class C lines only. It does not govern the four Class N lines in `prompt-caching.md`; their disposition is Q-SWD-1.
Operator verdict: where a current document or command exists to replace the stale citation, replace it; where none exists, delete that line. Source: board record d-20261019T015908Z-220a, decided_by=operator via AskUserQuestion, operator ruling 10-19, kind=ruling, scope=standing, umbrella rows Q13/Q17.
Lane reading (not an operator verdict): "delete that line" is applied to the stale citation inside its `mcp-server.md` cell, leaving the cell as the no-CLI marker `—`; the table row is kept, because each row documents a registered MCP tool whose only stale part is the CLI-equivalent cell. The two partial-overlap candidates (`moai spec status --list`, `moai codex status`) do not count as "a current command exists" unless the operator names them. For `worktree/_index.md` the slash forms `/moai run SPEC-ID` and `/moai sync SPEC-ID` exist, so the replace branch applies. If the operator meant deleting the table row, the SPEC needs a plan-phase revision of REQ-PRC-001-001 and the A.5 table.

### Q17: Do the same edits land in en, ja, and zh in the same change? (cross-cutting; SWEEPD copy)
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, as the umbrella cites it; that document was not re-read in this pass and is outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: for this child the question has one live shape. `mcp-server.md` and `worktree/_index.md` exist in all four locales and carry the same lines in each (E-SWD-2, E-SWD-4), so the candidate change is a same-disposition edit in eight files. The parity baseline holds today: `bash scripts/docs-i18n-check.sh` on tree 74ad197380 exits 0 and prints `Errors:   0` and `Warnings: 0` (E-SWD-14). The card text states the four-locale rule ("all four locales (ko, en, ja, zh) are fixed in the same PR (Q17)"); at the E-SWD-17 read no board record named Q17 for this card, and the operator verdict recorded below came afterwards. Whether the obligation binds this child's edits was the operator's call, since the document is not an authority.
Requirement: REQ-PRC-001-004
Class: implementation-level
Child: cross-cutting (README, REDIR, SWEEPA to SWEEPF, FLAG); SWEEPD copy
Default: edit all four locales in one change (rule: preserves current behavior; the four locales are in parity today)
Alternate: edit the ko copy in this change and the en, ja, and zh copies in a later change. This SPEC does not plan that branch: AC-PRC-001-001 to AC-PRC-001-004 read four locales, and a verdict for the alternate returns the SPEC to the plan phase
Operator verdict: fix all four locales (ko, en, ja, zh) in the same card and the same change. Source: board record d-20261019T015908Z-220a, decided_by=operator via AskUserQuestion, operator ruling 10-19, kind=ruling, scope=standing, umbrella rows Q13/Q17.

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
Operator verdict:

### Q-SWD-2: Do the five adjacent Mermaid labels per locale join the sweep?
Label: FOUNDER
Authority anchor: none in the committed tree.
Why unresolved: besides the two labels the card names, each `worktree/_index.md` carries five Mermaid labels that name an unregistered form but match none of the umbrella families, because they have no `SPEC-ID`: `"moai plan<br/>..."` (one), `"moai plan"` (one), `"moai run<br/>..."` (two), and `"moai sync<br/>..."` (one), 20 lines in all (plan.md §A.2, E-SWD-13). `moai run` and `moai sync` are `Unknown command`; `moai plan` is a registered verb that renders a plan report and is described in its own help as distinct from the `/moai plan` skill (E-SWD-10). The card text names only the `SPEC-ID` pair. If the 20 lines stay, the swept page shows the corrected pair next to uncorrected labels of the same kind (plan.md R7), and the umbrella's closing measurement, which keys on the family patterns, will not see them.
Requirement: REQ-PRC-001-002, REQ-PRC-001-005
Class: implementation-level
Child: SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), new row
Default: leave the 20 labels unchanged (rule: preserves current behavior; the card's scope names only the two `SPEC-ID` lines)
Alternate: extend the edit unit of `worktree/_index.md` to the five labels, with the same registered-slash-form disposition as Q13; not planned, returns the SPEC to the plan phase
Operator verdict:

## Verdict state of each row

| Row | Class | Default | Operator verdict |
|-----|-------|---------|------------------|
| Q13 | implementation-level | proposed, per form: `—` for the seven CLI cells, slash form for the two Mermaid labels | recorded: replace where a current command or document exists, else delete that line (board 220a, operator ruling 10-19) |
| Q17 | implementation-level | proposed: all four locales in one change | recorded: all four locales in the same card and change (board 220a, operator ruling 10-19) |
| Q-SWD-1 | implementation-level | proposed: (b) retain | empty (open) |
| Q-SWD-2 | implementation-level | proposed: leave the 20 labels | empty (open) |

All four rows are in the Gate set above; the Gate still waits for Q-SWD-1 and Q-SWD-2.
