# SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 — Implementation Plan (revision 0.1.0)

## A. Context

- Child of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (umbrella; its plan.md §B.3 import rule applies: the umbrella imports this child's criteria by SPEC ID, criterion ID, and commit SHA, never by copied text).
- Card t1725, Class C, docs-only. Tier M (3 artifacts; 10 files edited, inside the 5 to 15 band). Requirement count 6 and acceptance-criterion count 7, both under the Tier M ceiling of 16.
- Pin 74ad197380a0. Rulings in force: Q13 and Q17 by d-20261019T015908Z-220a; Q24 by d-20261014T060405Z-0fcd.
- Planner defaults only for the residual rows QE-1 (decided by Q13), QE-2, QE-3 (spec.md § 6). QE-2 and QE-3 carry no operator verdict; the run phase proceeds on the defaults, which add no marker and keep the umbrella pattern.

## B. Milestones (ordered by likelihood of change: user-visible text first, mechanical verification last)

### M0 — pre-flight (before any edit)

- Capture `BASELINE_SHA=$(git rev-parse HEAD)` and record it in the progress record (umbrella remedy R2).
- Re-run cells S-01, S-02, S-06a, S-11, S-12 on the run-phase tree and confirm they still show the RED-now values (ten rows, ten rows, one row, no output, no output). If a count differs, stop and return a blocker report; the tree moved.
- Confirm the plan is audited and the Kickoff gate is met. Decision rows QE-2 and QE-3 are open; proceed on the defaults.

### M1 — replacement text in the user-visible pages (eight files; most likely to change after review)

1. `utility-commands/moai-feedback.md`, four locales. In the Usage fence replace the alias comment and the three alias lines (`# Short aliases`, `> /moai fb`, `> /moai bug`, `> /moai issue`; lines 31 to 34) with a comment in the page's language meaning "Direct slash command" and the line `> /moai:feedback`. The standard form (`> /moai feedback`) stays. Net per file: at most 2 insertions and 4 deletions.
2. `getting-started/faq.md`, four locales. Replace the `/moai init` bullet (en, ja, zh line 24; ko line 28) with one bullet saying `moai init` is a terminal command that sets up a project and that inside Claude Code project documentation is generated with `/moai project`. The sibling bullet about `moai plan` stays untouched. Net per file: 1 insertion and 1 deletion.
3. Write the ja, ko, and zh text as native prose, not word-for-word translation; keep identifiers (`moai init`, `/moai project`, `/moai:feedback`) verbatim.
4. Flip check: S-11 and S-12 (with the working HEAD) each print four rows ending `:1`.

### M2 — narrative lines in the two English-only pages (two files)

1. `claude-code/foundations/_index.md` line 39: change "and the /moai relationship" to "and where `/moai` sits". One insertion, one deletion.
2. `core-concepts/verification-claim-integrity.md` line 84: delete the list items "the `/moai brain` command," and "the `moai brain` CLI," from the producer enumeration, keeping the grammar of the remaining list intact. No marker is added (QE-2 default). One insertion, one deletion.
3. Flip check: S-01, S-02, and S-06a (working HEAD) print nothing and exit 1.

### M3 — verification and record (mechanical)

1. Run the closing cells S-01, S-02, S-03, S-05a to S-05c, S-06a, S-11, S-12 with the working HEAD; record verbatim stdout and exit code for each in the progress record.
2. Run S-10 (`DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh`) and record Errors and Warnings.
3. Run S-13 (`git diff --numstat BASELINE_SHA HEAD -- docs-site`) and compare each file against the AC-PRC-001f limits.
4. Report the commit SHA to the umbrella so M6 can import AC-PRC-001a to AC-PRC-001g at that SHA.

## C. Technical approach

- Edits are line-level replacements with the Edit tool, one pathspec per file; no file is rewritten. Stage by explicit pathspec only.
- The replacement targets are registered forms: `/moai:feedback` (file `.claude/commands/moai/feedback.md`) and `/moai project` (file `.claude/commands/moai/project.md`).
- The marker (spec.md § 2) is defined but not applied. If the operator picks the QE-2 alternate, add `<!-- slash-example: <reason> -->` on the line above line 84, and the GATE checker must read it.

## D. Risks

- **Operator reading of "explicit example marker".** If the operator treats the "Worked Example" heading as a marker, line 84 needs no edit. Mitigation: decision QE-2 stays open; the default edit is harmless.
- **Closing pattern stricter than resolution.** The umbrella pattern also rejects the valid aliases `fb` and `init` (QE-3). Mitigation: the replacements use registered forms, which pass both tests.
- **Locale register drift.** Native prose in ja, ko, zh may differ in tone from neighbouring text. Mitigation: match the surrounding page; the strict i18n check covers parity, not register.
- **Shared ko files.** The umbrella notes ko files shared with README, SWEEPB, SWEEPC, SWEEPD, and FLAG on different lines. Mitigation: this child edits only the lines above; the bounded-diff check (AC-PRC-001f) shows no collateral edit.
- **Tooling gap.** No executable resolver exists (G-1, G-3). The closing verdict by the GATE producer is deferred to the umbrella's M6.

## E. Anti-patterns to avoid

- Fixing the English page only (breaks Q17 and parity).
- Adding the example marker to make a line pass (the Q24 verdict says unmarked lines are fixed).
- Deleting whole sections to zero the pattern (caught by AC-PRC-001f).
- Editing the docs-site pages in the plan phase (they are run-phase work).

## F. Cross-references

- Umbrella: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (REQ-PRC-001, AC-PRC-001, E-04, decision rows Q13, Q17, Q24).
- i18n obligation: `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 (a fact, not an authority); the operative authority is operator record d-20261019T015908Z-220a.
- Router: `.claude/skills/moai/SKILL.md` Priority 1 (lines 66 to 82).
