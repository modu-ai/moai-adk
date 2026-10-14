# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001

Revision 0.3.0 (2026-10-14), route 2 (operator decision `d-20261014T114834Z-e039`). This index keeps the full rows Q12, Q20, Q22, and Q24. Q22 keeps its full row with a proposed owner (GATE), pending leader confirmation; it is not decided here. Every other row is a stub. A stub keeps the label, the class, and the fixed row shape, and names the child SPEC that owns the decision; the child's decision index carries the full row and its Operator verdict. The source text of every moved row is at commit bdb7b764, path `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/decision-index.md`. Stub verdict lines are empty here by design: an answer written in this index is not a verdict for the owning child. No Operator verdict was written or changed in this revision.

Revision 0.2.2 (2026-10-14). Each row states one decision the operator does not settle in the interview. No row carries a recommendation.

Revision 0.2.2 re-pins the RED-now cells to 8816de4dae under leader ruling `d-20261014T072226Z-9b67` and the leader's scope choice B+. Supporting, verification, and protocol cells stay at f35103480f and are labelled; the drift table in acceptance section A lists the four items that differ. E-09b2 and E-28 exclude this SPEC's directory. The E-12 header count, the E-19b reading note, and the Q2 wording are corrected. No label, class, default, or verdict changed.

Revision 0.2.1 corrects revision 0.2.0 under leader ruling `d-20261014T070143Z-75ac`: the G21 gap sentence in acceptance.md, the E-04 count (22 files, 45 matching lines), and the Q2 reference scope (E-09b and E-09b2). No label, class, default, or verdict changed.

Revision 0.2.0 changes five things. It follows the umbrella split (decision Q20 = (c), `d-20261014T055241Z-1bc6`) and two later operator records (`d-20261014T060405Z-0fcd` and `d-20261014T061812Z-f5ca`):

1. Every row carries a `Child:` line that names the child SPEC owning the decision. Q12 and Q22 are parent rows (operator-held). Q20 is the umbrella's own tier row.
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
Child: parent (operator-held) in child-cards.md §5; owner GATE, leader-confirmed (card:t1655, 2026-10-14) (see the owner note below)
Owner (revision 0.3.0; leader-confirmed for card:t1655 on 2026-10-14; this assignment does not decide the Q22 answer): GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001). Reasoning: (1) REQ-PRC-015 requires the three jobs of REQ-PRC-012 to REQ-PRC-014 to be required checks, and GATE writes all three. (2) child-cards.md (GATE entry) and split-proposal §6 close REQ-PRC-015 after GATE's job names exist. (3) child-cards.md §5 names the parent for Q22, which does not satisfy decision e039 (open questions Q22, Q21, and Q7 are decided inside the owning child); so the child-cards.md row gives no child owner. Q12, whether the contexts are added, stays at the parent, as decision 0fcd and child-cards.md put it. If the leader does not confirm GATE, Q22 stays in this index, and the e039 route for Q22 needs a new operator record.
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
