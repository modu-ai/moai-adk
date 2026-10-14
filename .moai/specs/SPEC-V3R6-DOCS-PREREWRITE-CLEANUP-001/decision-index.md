# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001

Revision 0.3.1 (2026-10-14), route 2 (operator decision `d-20261014T114834Z-e039`). This index keeps the full rows Q12, Q20, and Q24. The full Q22 row also stays here for reference; Q22 has owner GATE (leader-confirmed), recorded for card:t1655 on 2026-10-14; the answer is not decided here. Every other row is a stub. A stub keeps the label, the class, and the fixed row shape, and names the child SPEC that owns the decision; the child's decision index carries the full row and its Operator verdict. The source text of every moved row is at commit bdb7b764, path `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/decision-index.md`. Stub verdict lines are empty here by design: an answer written in this index is not a verdict for the owning child. No Operator verdict was written or changed in this revision.

Revision 0.2.2 (2026-10-14). Each row states one decision the operator does not settle in the interview. No row carries a recommendation.

Revision 0.2.2 re-pins the RED-now cells to 8816de4dae under leader ruling `d-20261014T072226Z-9b67` and the leader's scope choice B+. Supporting, verification, and protocol cells stay at f35103480f and are labelled; the drift table in acceptance section A lists the four items that differ. E-09b2 and E-28 exclude this SPEC's directory. The E-12 header count, the E-19b reading note, and the Q2 wording are corrected. No label, class, default, or verdict changed.

Revision 0.2.1 corrects revision 0.2.0 under leader ruling `d-20261014T070143Z-75ac`: the G21 gap sentence in acceptance.md, the E-04 count (22 files, 45 matching lines), and the Q2 reference scope (E-09b and E-09b2). No label, class, default, or verdict changed.

Revision 0.2.0 changes five things. It follows the umbrella split (decision Q20 = (c), `d-20261014T055241Z-1bc6`) and two later operator records (`d-20261014T060405Z-0fcd` and `d-20261014T061812Z-f5ca`):

1. Every row carries a `Child:` line that names the child SPEC owning the decision. Q12 is a parent row (operator-held). Q22 has owner GATE (leader-confirmed), so its row names GATE. Q20 is the umbrella's own tier row.
2. Q20 carries its operator verdict: (c), split into child SPECs with tier M kept.
3. Q23 is new. It covers the moai-mx redirect pair in `docs-site/vercel.json` (lines 163 and 168), raised by the leader's B04 ruling and owned by the REDIR child.
4. Q24 is new. It records the operator's verdict on N2: only narrative examples that carry an explicit example marker are exempt from the stale-citation check.
5. Every other `Operator verdict:` stays empty. Every `Planner default applied:` line is kept unchanged; those values are provisional, not verdicts.

Labels: DECIDED and POLICY-COVERED need a committed anchor in the authority register (committed artifacts only). None of the rows below has one, so every row routes to FOUNDER or EVIDENCE-NEEDED. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the authority register; the rows cite it as a fact, not as an authority.

Evidence references (E-nn) point to acceptance.md §C, measured at tree 8816de4dae for the RED-now cells; the supporting cells are labelled at f35103480f and listed in the drift table in acceptance.md section A.

### Q1: Which single Korean page keeps the CG retirement passage?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to the owning child in revision 0.3.0; the full row, its default, and its planner default are in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: README (SPEC-V3R6-DOCS-PREREWRITE-README-001), decision row Q1
Operator verdict:

### Q2: Do the five unreferenced images go, or do they gain a reference?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764. The child's basis must name the tracked historical report (E-09b2) as a non-live reference.
Class: implementation-level
Child: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), decision row Q2
Operator verdict:

### Q3: What byte budget does the README OG image meet after re-encoding?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), decision row Q3
Operator verdict:

### Q4: Are the two Nextra-named scripts deleted?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), decision row Q4
Operator verdict:

### Q5: Are `translate.mjs` and `migrate-frontmatter.sh` kept?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: moved in revision 0.3.0. The row is a producer input only: it does not gate AC-PRC-007. The full row is in the owning child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), decision row Q5, producer input only
Operator verdict:

### Q6: Does the Python-era `docs-site/README.md` go or get replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), decision row Q6
Operator verdict:

### Q7: Is `docs-site/static/llms.txt` removed or regenerated?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to the owning child in revision 0.3.0, and decided inside that child (operator decision `d-20261014T114834Z-e039`). Cross-SPEC effect: AC-DVR-019b of the completed SPEC-V3R6-DOCS-V3-REBUILD-001 requires the canonical domain in `llms.txt`. The RESID row names the amendment path under each branch (defect B10). Source: commit bdb7b764.
Class: product-level
Child: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), decision row Q7
Operator verdict:

### Q8: Which release does the README badge and `hugo.toml` name until v3.2.0 GA?
Label: FOUNDER
Authority anchor: none in the committed tree (no badge policy in the committed docs or rules, E-22)
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764.
Class: product-level
Child: README (SPEC-V3R6-DOCS-PREREWRITE-README-001), decision row Q8
Operator verdict:

### Q9: Does the i18n parity job block pull requests from this change onward?
Label: FOUNDER
Authority anchor: none in the committed tree (SPEC-DOCS-SITE-001 is `implemented` and silent on strict mode, E-23)
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), decision row Q9
Operator verdict:

### Q10: Which Hugo version does the new CI build job use?
Label: FOUNDER
Authority anchor: `docs-site/vercel.json` `HUGO_VERSION` (committed; E-17b)
Why unresolved: moved to the owning child in revision 0.3.0; the full row, its default, and its planner default are in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), decision row Q10
Operator verdict:

### Q11: Which link-check tool runs in the new job?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: moved to the owning child in revision 0.3.0; the full row is in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), decision row Q11
Operator verdict:

### Q12: Are the three new jobs added to the branch protection's required checks?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the current required contexts are five, and none of the three jobs is among them (E-21, external). Changing branch protection is an operator-held action on an external shared system (keep-set). The operator form is needed, and the change cannot be observed from the tree. If REQ-PRC-002's cited-verbs step becomes a required context, it is a fourth candidate for the same answer.
Class: implementation-level
Child: parent (operator-held)
Operator verdict:

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to the owning children in revision 0.3.0 (README and SWEEPA to SWEEPF, each for its own files). The full row, its default, and its planner default are in those children's decision indexes. Source: commit bdb7b764. The row's count of stale forms is corrected to 18 in the children (defect B12).
Class: implementation-level
Child: README (SPEC-V3R6-DOCS-PREREWRITE-README-001) and SWEEPA to SWEEPF, decision row Q13
Operator verdict:

### Q14: Is the `moai-design` page kept, rewritten, or removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to SWEEPF in revision 0.3.0; the full row is in that child's decision index. Its marker option needs an amendment of REQ-PRC-001, which SWEEPF writes (defect B03). Source: commit bdb7b764.
Class: product-level
Child: SWEEPF (SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001), decision row Q14
Operator verdict:

### Q15: Are the three redirect rules for `/agency` and `/design` removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to REDIR in revision 0.3.0; the full row, the class justification, and the fail-closed reclassification clause are in that child's decision index. Source: commit bdb7b764.
Class: implementation-level (the justification and the fail-closed clause move with the row)
Child: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), decision row Q15
Operator verdict:

### Q16: Are the four shadowed changelog index pages removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to REDIR in revision 0.3.0; the full row, the class justification, and the fail-closed reclassification clause are in that child's decision index. Source: commit bdb7b764.
Class: implementation-level (the justification and the fail-closed clause move with the row)
Child: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), decision row Q16
Operator verdict:

### Q17: Do the same edits land in en, ja, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: moved in revision 0.3.0 as a cross-cutting row. Each owning child applies the answer to its own files, and each child's decision index carries a copy of the row. Source: commit bdb7b764.
Class: implementation-level
Child: cross-cutting (REDIR, README, SWEEPA to SWEEPF, FLAG), decision row Q17
Operator verdict:

### Q18: Is each `--factory-run` and `--clear-policy` citation kept, with the flag added to help, or removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to FLAG in revision 0.3.0; GATE takes the answer as an input. The help-text option is outside this SPEC's scope. The full row is in that child's decision index. Source: commit bdb7b764.
Class: product-level
Child: FLAG (SPEC-V3R6-DOCS-PREREWRITE-FLAG-001), decision row Q18
Operator verdict:

### Q19: Is the maintainer absolute path in `docs-site/scripts/README_TRANSLATE.md` removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to RESID in revision 0.3.0. The row has no REQ or AC yet; RESID adds one or excludes the row (verdict O7). The full row, its default, and its planner default are in that child's decision index. Source: commit bdb7b764.
Class: implementation-level
Child: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), decision row Q19
Operator verdict:

### Q20: Is this SPEC's tier M, or L with the Tier L artifact set?
Label: FOUNDER
Authority anchor: none in the committed tree that decides this SPEC's tier. The tier table is `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.
Why unresolved: the frontmatter says `tier: M` (three artifacts, PASS threshold 0.80). The planned edit set spans more than the 15 files that Tier M allows. The stale families alone touch 48 files for CLI forms (E-01) and 22 files for slash forms (E-04; 45 matching lines), and the image, script, workflow, redirect, and skill edits add to that. Tier L requires `design.md` and `research.md` and a 0.85 threshold. The operator decides between the two options. This revision keeps `tier: M`, authors neither file, and does not decide the question.
Options: (a) keep Tier M, reduce the scope to 15 files or fewer, and state the reasoning; (b) move to Tier L, author `design.md` and `research.md`, and re-audit at 0.85; (c) split into child SPECs, each tier M, with the umbrella keeping tier M.
Class: implementation-level. This row changes the artifact set and the audit gate, not shipped behavior. No Default is ranked.
Child: parent (umbrella tier)
Operator verdict: (c) split into child SPECs, tier M kept — operator, d-20261014T055241Z-1bc6 (2026-10-14T05:52:41Z)

### Q21: How are the two external changelog redirect rules handled?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to REDIR in revision 0.3.0, and decided inside that child (operator decision `d-20261014T114834Z-e039`). The full row, with options (a) and (b), is in that child's decision index. Source: commit bdb7b764.
Class: product-level (the public changelog URLs change what they serve under either option)
Child: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), decision row Q21
Operator verdict:

### Q22: Is REQ-PRC-015 moved to an operator-held section, or is the branch-protection exclusion amended?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: REQ-PRC-015 requires the three jobs to be required status checks on `main` (external), while the Out of Scope section excludes changing the branch protection of `main` (spec.md). The requirement and the exclusion conflict, and neither is decided here. Revision 0.2.0 places REQ-PRC-015 under an operator-held heading in spec.md; that placement is not the verdict, and the exclusion stays as written until Q22 is answered. The current required contexts are five (E-21). The options: (a) move the closure out of the requirements into an operator-held section, keeping the exclusion; (b) amend the exclusion to name this change as an operator-held action.
Class: implementation-level (an external setting; keep-set)
Child: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), owner GATE (leader-confirmed), recorded for card:t1655 on 2026-10-14. The earlier split note named the parent before this confirmation (Appendix A, note 1; see the owner note below)
Owner (revision 0.3.0; this assignment does not decide the Q22 answer): Q22 has owner GATE (leader-confirmed), recorded for card:t1655 on 2026-10-14 (SPEC-V3R6-DOCS-PREREWRITE-GATE-001). Reasoning: (1) REQ-PRC-015 requires the three jobs of REQ-PRC-012 to REQ-PRC-014 to be required checks, and GATE writes all three. (2) Appendix A (GATE row; note 2) and split-proposal §6 close REQ-PRC-015 after GATE's job names exist. (3) The earlier split note placed Q22 at the parent (Appendix A, note 1), which does not satisfy decision e039: open questions Q22, Q21, and Q7 are decided inside the owning child, so the split note gives no child owner for Q22. Q12, whether the contexts are added, stays at the parent, as decision 0fcd and Appendix A, note 3, put it. The answer to Q22 is decided inside GATE (decision e039).
Operator verdict:

### Q23: Are the two moai-mx redirect rules removed, or is REQ-PRC-011 amended to exempt the pair?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: moved to REDIR in revision 0.3.0 by the leader's B04 ruling (`d-20261014T060405Z-0fcd`). The full row, with options (a) and (b), is in that child's decision index. Source: commit bdb7b764. Requirement: REQ-PRC-011.
Class: product-level (option (a) changes what the public moai-mx URLs serve; option (b) amends a requirement)
Child: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), decision row Q23
Operator verdict:

### Q24: Which narrative examples are exempt from the stale-citation check (N2)?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the E-04 slash family also counts narrative and error examples: `docs-site/content/en/getting-started/faq.md` line 24 (`/moai init` described as a misconception) and `docs-site/content/en/core-concepts/verification-claim-integrity.md` line 84 (a retired `/moai brain` narrative). The tree does not decide whether lines of that kind are exempt from the REQ-PRC-001 check (E-04, N2).
Class: implementation-level
Child: SWEEPE (applies the exemption; the marker definition is SWEEPE's, defect B03). The row stays here in full because its operator verdict is recorded here (decision 0fcd).
Operator verdict: exempt only narrative examples that carry an explicit example marker; lines without a marker are fixed — operator, d-20261014T060405Z-0fcd (2026-10-14T06:04:05Z)

## Appendix A: child SPEC list

This is the child list of this index. The proposed child IDs and card IDs follow spec.md § Child SPECs, and the leader confirms the final child IDs. The precondition column gives each child's before-start condition in one line; where the split note states none, the cell says so.

| TAG | Proposed child ID | Card | REQ | AC | Decision rows owned | Precondition |
|-----|-------------------|------|-----|----|---------------------|--------------|
| GATE | SPEC-V3R6-DOCS-PREREWRITE-GATE-001 | t1717 | REQ-PRC-002, REQ-PRC-012, REQ-PRC-013, REQ-PRC-014 | AC-PRC-002, AC-PRC-012, AC-PRC-013, AC-PRC-014 | Q9, Q10, Q11; Q22 with owner GATE (leader-confirmed); inputs Q5 and Q18 | Before starting, correct the RED cell of AC-PRC-012 (audit B01); the current cell cites E-16d, which is green. |
| REDIR | SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 | t1718 | REQ-PRC-010, REQ-PRC-011 | AC-PRC-010, AC-PRC-011 (extended to the moai-mx pair under Q23) | Q15, Q16, Q21, Q23 | Record the moai-mx pair decision row and its AC first; per audit B04, lines 163 and 168 of docs-site/vercel.json hide the existing moai-mx.md pages in the four locales. |
| RESID | SPEC-V3R6-DOCS-PREREWRITE-RESID-001 | t1719 | REQ-PRC-005, REQ-PRC-006, REQ-PRC-007, REQ-PRC-008, REQ-PRC-009 | AC-PRC-005, AC-PRC-006, AC-PRC-007, AC-PRC-008, AC-PRC-009 | Q2, Q3, Q4, Q6, Q7, Q19 | Correct the Q2 basis to match the tracked historical report that references cg-mode-infographic (audit B05); add a REQ and AC for Q19 or exclude it; split into 3a (Q2, Q4, Q6, Q19) and 3b (Q3, Q7) if the Q3 or Q7 decision is late. |
| README | SPEC-V3R6-DOCS-PREREWRITE-README-001 | t1720 | REQ-PRC-001 (README part), REQ-PRC-003, REQ-PRC-004 | AC-PRC-001 (README part), AC-PRC-003, AC-PRC-004 | Q1, Q8, Q13, Q17 | Set the badge values after decision Q8; the README flag lines are handled by FLAG. |
| SWEEPA | SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001 | t1721 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | None stated; its closing check runs after GATE (child 1). |
| SWEEPB | SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001 | t1722 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | None stated; landing after README (child 4) is recommended. |
| SWEEPC | SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001 | t1723 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | None stated; landing after README (child 4) is recommended. |
| SWEEPD | SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001 | t1724 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | None stated; landing after README (child 4) is recommended. |
| SWEEPE | SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 | t1725 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17; Q24 (N2, decision 0fcd) | At start, check whether docs-site/content/en/getting-started/faq.md line 24 and docs-site/content/en/core-concepts/verification-claim-integrity.md line 84 carry an explicit example marker (Q24); lines without one are fixed. |
| FLAG | SPEC-V3R6-DOCS-PREREWRITE-FLAG-001 | t1726 | new REQ, written by the child | new AC, written by the child | Q18 | Write the REQ and AC first; the add-to-help option of Q18 changes internal/ and cmd/, so choosing it needs a separate SPEC. |
| SWEEPF | SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 | t1727 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q14, Q17 | Confirm who edits _meta.yaml before starting; the structure curator is its single author. |

Notes:

1. The earlier split note placed Q22 at the parent, together with REQ-PRC-015 and Q12. Decision e039 moved the owner of Q22 to GATE: Q22 has owner GATE (leader-confirmed), and its answer is decided inside GATE.
2. REQ-PRC-015 is closed by the parent after the three job names created by GATE exist.
3. Q12 and Q20 have no child and stay at the parent. Q12 stays at the parent under decision 0fcd.
4. Q13 is applied by README and by SWEEPA to SWEEPF, each to its own files. Q17 is cross-cutting: each child that changes locale files applies it to its own files.
