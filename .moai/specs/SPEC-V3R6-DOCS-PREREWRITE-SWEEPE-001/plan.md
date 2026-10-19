# SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 — Implementation Plan (revision 0.3.0)

## A. Context

- Child of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (umbrella; its plan.md §B.3 import rule applies: the umbrella imports this child's criteria by SPEC ID, criterion ID, and commit SHA, never by copied text).
- Card t1725, Class C, docs-only. Tier M (3 artifacts; 12 page files edited, inside the 5 to 15 band). Requirement count 7 and acceptance-criterion count 11, both under the Tier M ceiling of 16.
- Pin 74ad197380a0. Rulings in force: Q13 and Q17 by d-20261019T015908Z-220a; Q24 by d-20261014T060405Z-0fcd.
- Revision 0.2.0 answers plan audit iteration 1 (FAIL 0.69, `.moai/reports/t1725/plan-audit-verdict.md`, defects D1 to D12). The unit of work is the line the umbrella's closing pattern flags with no example marker, not only a line whose citation fails to resolve (spec.md § 1, § 2).
- Revision 0.3.0 answers plan audit iteration 2 (FAIL 0.84, `.moai/reports/t1725/plan-audit-verdict-iter2.md`, blocking defects D-N1 to D-N3 and optional D-N4 to D-N7): the block label line joins the edit set (D-N1, decision QE-6), the pinned label and the whole-line content fence are checked per locale (D-N2), the FAQ and foundations content cells are one per locale (D-N3), S-06c is parent-keyed and S-06d is rewritten as single-invocation cells (D-N4, D-N5), and the ja and zh heading divergence is an accepted residual (D-N7).
- Planner defaults applied for QE-1 to QE-6 (spec.md § 6): QE-1 (replace the alias lines), QE-2 (no marker on line 84), QE-3 (umbrella pattern unchanged), QE-4 (ja and zh foundations rows change with the en row), QE-5 (the sibling `moai plan` FAQ bullet is out of scope), QE-6 (the label line above the alias lines changes with them). The leader records the default application on the decision board before Kickoff. The release-blocking criteria do not depend on the QE-2 alternate (acceptance.md § H).

## B. Milestones (ordered by likelihood of change: user-visible text first, mechanical verification last)

### M0 — pre-flight (before any edit)

- Confirm acceptance.md S-00 prints nothing and exits 0: the docs-site, `.claude`, `internal`, `cmd`, and `scripts` trees are unchanged since `74ad197380a0`. The scope fence base is that commit; no baseline variable is captured at run time.
- Re-run the RED-now cells with `HEAD` in place of the pin and confirm their values: S-01 ten rows, S-02 ten rows, S-06a one row, S-06b one row, S-11 empty (exit 1), the four S-12a-xx empty (exit 1), S-12b four rows, S-13a empty, S-13b empty, the four S-16-xx empty (exit 0), S-17 the twelve unedited blob ids, the three S-14-xx empty (exit 1), S-15 empty (exit 1), the four S-18-xx empty (exit 1). If a value differs, stop and return a blocker report; the tree moved.
- Confirm the plan is audited and the Kickoff gate is met. Decision rows QE-1 to QE-6 were adopted as written by the leader ruling recorded on the decision board as d-20261019T033214Z-032b (scope card:t1725, 2026-10-19T03:32:14Z).

### M1 — replacement text in the eight user-visible pages (most likely to change after review)

1. `utility-commands/moai-feedback.md`, four locales. In the Usage fence replace lines 31 to 34 — the block label line (`# Short aliases` and its translations; not a closing-pattern match, but in the edit set by REQ-SWEEPE-004 and decision QE-6) and the three flagged alias lines (`> /moai fb`, `> /moai bug`, `> /moai issue`) — with exactly the two-line pinned block of spec.md § 4: the label in the page's language meaning "Direct slash command" at line 31 and `> /moai:feedback` at line 32. The standard form (`> /moai feedback`) and its `# Standard form` comment stay. Numstat per file: exactly 2 insertions and 4 deletions.
2. `getting-started/faq.md`, four locales. Replace the `/moai init` bullet (en, ja, zh line 24; ko line 28) with the pinned line of spec.md § 4, byte for byte. The sibling bullet about `moai plan` stays untouched (QE-5). Numstat per file: exactly 1 insertion and 1 deletion.
3. The ja, ko, and zh text is the pinned native prose of spec.md § 4; identifiers (`moai init`, `/moai project`, `/moai:feedback`) stay verbatim. A wording change to a pinned line is a blocker report (spec.md § 4), not a local edit.
4. Commit, then flip check with `HEAD`: S-11 prints four rows at line 32; each S-18-xx prints one row at line 31; each S-16-xx prints its pinned diff block; each S-12a-xx prints one row at its line; S-12b prints nothing and exits 1. At this point S-01 prints two rows (en foundations, en verification-claim-integrity).

### M2 — the foundations rows and the narrative line (four files)

1. `claude-code/foundations/_index.md`, en line 39, ja line 40, zh line 40: the three rows take the pinned wording of spec.md § 4 (en "and where `/moai` sits"; ja `/moai の位置づけ`; zh `/moai 的定位`). One insertion and one deletion each.
2. `core-concepts/verification-claim-integrity.md`, en line 84: delete the list items "the `/moai brain` command," and "the `moai brain` CLI," from the producer enumeration so the paragraph reads exactly as the pinned fragment of spec.md § 4, keeping the rest of the paragraph. No marker is added (QE-2 default). One insertion and one deletion.
3. Commit, then flip check with `HEAD`: each S-14-xx prints one row at its line; S-15 prints one row ending `:1`; S-17 prints the twelve pinned blob ids; S-01, S-02, S-06a, S-06b print nothing and exit 1.

### M3 — verification and record (mechanical)

1. Run the closing cells S-01, S-02, S-03, S-05a to S-05c, S-06a to S-06c, S-11, S-12a-en, S-12a-ja, S-12a-ko, S-12a-zh, S-12b, S-14-en, S-14-ja, S-14-zh, S-15, S-18-en, S-18-ja, S-18-ko, S-18-zh with `HEAD` after the final commit; record verbatim stdout and exit code for each in the progress record (§E.2).
2. Run S-10 (`DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh`) and record Errors and Warnings.
3. Run S-13a and S-13b and compare them with the twelve paths and twelve numstat rows of AC-PRC-001f; run S-16-en, S-16-ja, S-16-ko, S-16-zh and compare each with its pinned diff block; run S-17 and compare it with the twelve pinned blob ids. A mismatch in S-16 or S-17 is a blocker report (the pinned text is the contract), not a local rewording.
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
- **Counterpart rows read differently from their page headings.** The ja and zh foundations rows now say "position" while their linked pages' headings say "relationship" (QE-4). The four linked headings share no wording (measured in acceptance.md § G), and aligning them would edit unflagged pages outside the twelve page files. Mitigation: accepted residual risk; the alternate (leave ja and zh unchanged) is one revert.
- **Block label line outside the closing pattern.** Line 31 of each feedback page is edited although the closing pattern does not match it (QE-6). Mitigation: REQ-SWEEPE-002 and -004 and spec.md § 2 name it, the replacement is pinned, and S-16, S-17, S-18 check it; the alternate (keep the label) re-derives them.
- **The content fence is strict by design.** S-17 pins a blob id for every page file, so a reworded pinned sentence, a stray space, or a changed line ending fails it. Mitigation: that is the purpose (a same-count overwrite is otherwise invisible); the run phase returns a blocker report and the orchestrator re-delegates an amendment to manager-spec, which re-derives S-12a, S-14, S-16, S-17, and S-18 together.
- **Tooling gap.** No executable resolver exists (G-1, G-3). The closing verdict by the GATE producer is deferred to the umbrella's M6.

## E. Anti-patterns to avoid

- Fixing the English page only (breaks Q17 and parity).
- Adding the example marker to make a line pass (the Q24 verdict says unmarked lines are fixed).
- Deleting whole sections to zero the pattern (caught by AC-PRC-001f and AC-PRC-001k).
- Editing a line outside the edit set (the flagged set, the counterpart rows, and the feedback block label lines), including the sibling `moai plan` bullet (caught by AC-PRC-001f).
- Keeping the old `# Short aliases` comment above `> /moai:feedback`, or writing the English label into the ja, ko, and zh pages (caught by S-18-xx).
- Editing the docs-site pages in the plan phase (they are run-phase work).

## F. Cross-references

- Umbrella: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (REQ-PRC-001, AC-PRC-001, E-04, decision rows Q13, Q17, Q24).
- i18n obligation: `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 (a fact, not an authority); the operative authority is operator record d-20261019T015908Z-220a.
- Router: `.claude/skills/moai/SKILL.md` Priority 1 (lines 66 to 82).
- Traceability (revision 0.2.0, the Group 4 verb of `.claude/agents/moai/plan-auditor.md`, run on spec.md and acceptance.md of this directory): `COLLECTED: 7 REQ definitions (acceptance input: read)`, no `UNCOVERED` line, no `ORPHAN` line, exit 0. Revision 0.1.0 collected 1 of its 6 requirements because its IDs read as `REQ-PRC-001`.
- Traceability (revision 0.3.0): the Group 4 verb itself was refused by the worktree-session guard ("cannot be shown not to be git" for an inline `awk` program), so it was not run and not retried; the same two readings were taken with plain commands. `grep -n -o -E '^- \*\*REQ-[A-Z0-9-]+'` on spec.md printed seven definitions (lines 62 to 68: `REQ-PRC-001`, `REQ-SWEEPE-001` to `REQ-SWEEPE-006`; exit 0); `grep -n -E '^\| AC-'` on acceptance.md printed eleven mapping rows (lines 17 to 27; exit 0) naming: REQ-PRC-001 in a, b, c, h; REQ-SWEEPE-001 in g; REQ-SWEEPE-002 in a, b, d, k; REQ-SWEEPE-003 in d, i, j; REQ-SWEEPE-004 in f; REQ-SWEEPE-005 in e; REQ-SWEEPE-006 in i. Read by hand: no requirement without an acceptance row, no acceptance row naming an undefined requirement. This is a derived reading, not the verb's `COLLECTED:` line (acceptance.md gap G-15).
- Ordering (CN-4): `grep -n -E '^[ \t]*Exit:'` on this plan printed nothing (exit 1), so no milestone binds an `Exit:` criterion; the CN-4 verb is also an inline `awk` program and was not attempted after the Group 4 refusal, and M1 step 4, M2 step 3, and M3 step 1 to 3 read consistently with the criteria's flip order by hand.
