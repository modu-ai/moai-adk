# SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 — Implementation Plan (revision 0.2.0)

## A. Context

- Child of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (umbrella; its plan.md §B.3 import rule applies: the umbrella imports this child's criteria by SPEC ID, criterion ID, and commit SHA, never by copied text).
- Card t1725, Class C, docs-only. Tier M (3 artifacts; 12 page files edited, inside the 5 to 15 band). Requirement count 7 and acceptance-criterion count 11, both under the Tier M ceiling of 16.
- Pin 74ad197380a0. Rulings in force: Q13 and Q17 by d-20261019T015908Z-220a; Q24 by d-20261014T060405Z-0fcd.
- Revision 0.2.0 answers plan audit iteration 1 (FAIL 0.69, `.moai/reports/t1725/plan-audit-verdict.md`, defects D1 to D12). The unit of work is the line the umbrella's closing pattern flags with no example marker, not only a line whose citation fails to resolve (spec.md § 1, § 2).
- Planner defaults applied for QE-1 to QE-5 (spec.md § 6): QE-1 (replace the alias lines), QE-2 (no marker on line 84), QE-3 (umbrella pattern unchanged), QE-4 (ja and zh foundations rows change with the en row), QE-5 (the sibling `moai plan` FAQ bullet is out of scope). The leader records the default application on the decision board before Kickoff. The release-blocking criteria do not depend on the QE-2 alternate (acceptance.md § H).

## B. Milestones (ordered by likelihood of change: user-visible text first, mechanical verification last)

### M0 — pre-flight (before any edit)

- Confirm acceptance.md S-00 prints nothing and exits 0: the docs-site, `.claude`, `internal`, `cmd`, and `scripts` trees are unchanged since `74ad197380a0`. The scope fence base is that commit; no baseline variable is captured at run time.
- Re-run the RED-now cells with `HEAD` in place of the pin and confirm their values: S-01 ten rows, S-02 ten rows, S-06a one row, S-06b one row, S-11 empty, S-12a empty, S-12b four rows, S-13a empty, S-13b empty, S-14 empty, S-15 empty. If a value differs, stop and return a blocker report; the tree moved.
- Confirm the plan is audited and the Kickoff gate is met. Decision rows QE-1 to QE-5 carry `DEFAULT-APPLIED`; proceed on the defaults.

### M1 — replacement text in the eight user-visible pages (most likely to change after review)

1. `utility-commands/moai-feedback.md`, four locales. In the Usage fence replace the alias comment and the three alias lines (`# Short aliases`, `> /moai fb`, `> /moai bug`, `> /moai issue`; lines 31 to 34) with a comment in the page's language meaning "Direct slash command" (spec.md § 4) and the line `> /moai:feedback`. The standard form (`> /moai feedback`) stays. Numstat per file: exactly 2 insertions and 4 deletions.
2. `getting-started/faq.md`, four locales. Replace the `/moai init` bullet (en, ja, zh line 24; ko line 28) with the pinned line of spec.md § 4, byte for byte. The sibling bullet about `moai plan` stays untouched (QE-5). Numstat per file: exactly 1 insertion and 1 deletion.
3. The ja, ko, and zh text is the pinned native prose of spec.md § 4; identifiers (`moai init`, `/moai project`, `/moai:feedback`) stay verbatim. A wording change to a pinned line is a blocker report (spec.md § 4), not a local edit.
4. Commit, then flip check with `HEAD`: S-11 prints four rows ending `:1`; S-12a prints four rows ending `:1`; S-12b prints nothing and exits 1. At this point S-01 prints two rows (en foundations, en verification-claim-integrity).

### M2 — the foundations rows and the narrative line (four files)

1. `claude-code/foundations/_index.md`, en line 39, ja line 40, zh line 40: the three rows take the pinned wording of spec.md § 4 (en "and where `/moai` sits"; ja `/moai の位置づけ`; zh `/moai 的定位`). One insertion and one deletion each.
2. `core-concepts/verification-claim-integrity.md`, en line 84: delete the list items "the `/moai brain` command," and "the `moai brain` CLI," from the producer enumeration so the paragraph reads exactly as the pinned fragment of spec.md § 4, keeping the rest of the paragraph. No marker is added (QE-2 default). One insertion and one deletion.
3. Commit, then flip check with `HEAD`: S-14 prints three rows ending `:1`; S-15 prints one row ending `:1`; S-01, S-02, S-06a, S-06b print nothing and exit 1.

### M3 — verification and record (mechanical)

1. Run the closing cells S-01, S-02, S-03, S-05a to S-05c, S-06a to S-06c, S-11, S-12a, S-12b, S-14, S-15 with `HEAD` after the final commit; record verbatim stdout and exit code for each in the progress record (§E.2).
2. Run S-10 (`DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh`) and record Errors and Warnings.
3. Run S-13a and S-13b and compare them with the twelve paths and twelve numstat rows of AC-PRC-001f.
4. Report the commit SHA to the umbrella so M6 can import AC-PRC-001a to AC-PRC-001k at that SHA.

## C. Technical approach

- Edits are line-level replacements with the Edit tool, one pathspec per file; no file is rewritten. Stage by explicit pathspec only. Each milestone commits before its flip check, because `git grep <commit>` reads committed objects only.
- The replacement targets are registered forms: `/moai:feedback` (file `.claude/commands/moai/feedback.md`) and `/moai project` (file `.claude/commands/moai/project.md`). Resolution through a router alias does not exempt a flagged line; the closing pattern decides which lines are edited.
- The marker (spec.md § 2) is defined but not applied. If the operator picks the QE-2 alternate, add `<!-- slash-example: <reason> -->` on the line above line 84, apply the criteria of acceptance.md § H, and the GATE checker must read the marker.

## D. Risks

- **Operator reading of "explicit example marker".** If the operator treats the "Worked Example" heading as a marker, line 84 needs no edit. Mitigation: decision QE-2 carries its own criteria (acceptance.md § H); the default edit is harmless.
- **Closing pattern stricter than resolution.** The umbrella pattern also rejects the valid aliases `fb` and `init` (QE-3). Mitigation: the replacements use registered forms, which pass both tests.
- **Locale register drift.** Native prose in ja, ko, zh may differ in tone from neighbouring text. Mitigation: the lines are pinned and checked exactly (S-12a, S-14); the strict i18n check covers parity, not register.
- **Shared ko files.** The umbrella notes ko files shared with README, SWEEPB, SWEEPC, SWEEPD, and FLAG on different lines. Mitigation: this child edits only the lines above; the whole-tree fence (AC-PRC-001f) shows no collateral edit.
- **Counterpart rows read differently from their page headings.** The ja and zh foundations rows now say "position" while their linked pages' headings say "relationship" (QE-4). Mitigation: recorded as residual risk; the alternate (leave ja and zh unchanged) is one revert.
- **Tooling gap.** No executable resolver exists (G-1, G-3). The closing verdict by the GATE producer is deferred to the umbrella's M6.

## E. Anti-patterns to avoid

- Fixing the English page only (breaks Q17 and parity).
- Adding the example marker to make a line pass (the Q24 verdict says unmarked lines are fixed).
- Deleting whole sections to zero the pattern (caught by AC-PRC-001f and AC-PRC-001k).
- Editing a line outside the flagged set and the counterpart rows, including the sibling `moai plan` bullet (caught by AC-PRC-001f).
- Editing the docs-site pages in the plan phase (they are run-phase work).

## F. Cross-references

- Umbrella: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (REQ-PRC-001, AC-PRC-001, E-04, decision rows Q13, Q17, Q24).
- i18n obligation: `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 (a fact, not an authority); the operative authority is operator record d-20261019T015908Z-220a.
- Router: `.claude/skills/moai/SKILL.md` Priority 1 (lines 66 to 82).
- Traceability (revision 0.2.0, the Group 4 verb of `.claude/agents/moai/plan-auditor.md`, run on spec.md and acceptance.md of this directory): `COLLECTED: 7 REQ definitions (acceptance input: read)`, no `UNCOVERED` line, no `ORPHAN` line, exit 0. Revision 0.1.0 collected 1 of its 6 requirements because its IDs read as `REQ-PRC-001`.
