# Progress — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-audit-pending
plan_revision: 0.3.0
plan_commit: 6d335cd483 (feat(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan-phase artifacts (tier M, 3 artifacts), card t1725)
plan_followup_commits: b28da6f891 (progress.md and decision-index.md), bb7d928f68 (cite REQ in ACs, QE-1 relabel)
revision_1_commit: e85847fd33 (fix(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan revision 1 after audit FAIL 0.69 (D1-D12), card t1725)
revision_commit: c9ec672020 (revision 2, fix(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan revision 2 after audit FAIL 0.84 (D-N1..D-N7), card t1725; full c9ec672020afdb5300408fa58fe5ace443fe1f4e; backfilled by the follow-up commit because a commit cannot cite its own hash)
card: t1725 (Class C, docs-only), lane-26, run tn1p4e
parent: SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (child row SWEEPE; decisions Q13, Q17, Q24/N2)
measurement_pin: 74ad197380a0 (the tree the plan-phase counts were taken on; full 74ad197380a0711cdb9816d4e37c94d29af730ea; docs-site, .claude, internal, cmd, and scripts are identical to HEAD e85847fd33, acceptance.md S-00)
plan_audit: iteration 1 FAIL 0.69 (tier M threshold 0.80, MP-8 failed), verdict file `.moai/reports/t1725/plan-audit-verdict.md`, audited_sha bb7d928f682d34279d5d3584537fec0c3e68dd25, receipt rcpt-93c8df0b6680dfc7dac83fb9; iteration 2 FAIL 0.84 (above the 0.80 bar, but the required codex gate returned fail; blocking D-N1 to D-N3, optional D-N4 to D-N7), verdict file `.moai/reports/t1725/plan-audit-verdict-iter2.md`, audited_sha e85847fd331b14320d4825338d2923830f4d1bb0, receipt rcpt-90f562481259842169400426; revision 0.3.0 answers D-N1 to D-N5 and D-N7 and leaves D-N6 as recorded planner defaults; iteration 3 (delta on D-N1 to D-N3) not yet run
open_decisions: see decision-index.md. Q13, Q17, and Q24 carry operator verdicts (operator records d-20261019T015908Z-220a for Q13 and Q17, d-20261014T060405Z-0fcd for Q24). QE-1 to QE-6 carry DEFAULT-APPLIED planner verdicts (QE-1 to QE-5 at 2026-10-19T02:27:27Z, QE-6 at 2026-10-19T02:51:00Z, manager-spec); the leader records the default application of all six on the decision board before Kickoff (iteration-2 defect D-N6).

### Plan-phase measurements (source: .moai/reports/t1725/progress.md, manager-spec, 2026-10-19, pin 74ad197380a0; revision 0.3.0 re-measured every cell it touched, acceptance.md § D)

- Closing pattern (umbrella E-04) on the twelve page files: 10 rows, 18 lines (faq x4 `/moai init`; moai-feedback x4 `fb|bug|issue`; en foundations line 39 `relationship`; en verification-claim-integrity line 84 `brain`). The ja and zh foundations files print no row. Line 31 of each moai-feedback page (`# Short aliases` and translations) is not a match; it is the block label line (QE-6, REQ-SWEEPE-004).
- Complement census (`-P`, CJK-safe boundary): the same 18 lines; no unknown word beyond the six.
- Router Priority 1 (SKILL.md lines 69-70, repo and template copies): `init` is an alias of `project`, `fb` of `feedback`; `bug`, `issue`, `relationship`, `brain` are not. So the FAQ claim that `/moai init` "does nothing" is false, and the umbrella's word "unregistered" is looser than router resolution (QE-3).
- `moai brain` (CLI form, en verification-claim-integrity line 84) exits 1 with `Unknown command` on a tree build (rebuilt in revision 0.3.0: stdout empty, stderr 654 bytes, exit 1); umbrella families E-01..E-03 miss it. E-01..E-03 are 0 in the page files. Every other inline `moai <word>` token resolves (S-06b, S-06c parent-keyed, S-06d-1 to S-06d-5).
- Neither en faq line 24 nor en verification-claim-integrity line 84 carries an example marker; no HTML comment exists in either file.
- ja foundations line 40 and zh foundations line 40 translate the edited English clause; the ko row has no `/moai` clause (S-07, QE-4). The linked `commands.md` headings share no wording across the four locales (en 238 "Where the /moai Command Sits", ko 238 "/moai 명령은 어디에 얹혀 있나", ja 178 "...との関係", zh 178 "...的 /moai 命令的关系"); accepted residual risk (acceptance.md § G).
- Tree build VCS stamp (478285ab4c1f, modified) differs from HEAD; not used as provenance.
- Operator record d-20261019T015908Z-220a read through `moai decision read --scope card:t1725` (revision 0.1.0): Q13 replace-else-delete, Q17 four locales.
- Content fence (revision 0.3.0): fixture repositories `repo` and `repo2` (under `.moai/state/t1725-ctl/`, removed before the commit) hold the pin blobs, a `good` branch with exactly the pinned edits, and mutant branches. Observed: the iteration-2 same-count overwrite mutant keeps numstat `2 4` and fails S-18-en, S-16-en, and S-17 (control C-2); the English-in-every-locale mutants pass the earlier single-alternation S-12a and S-14 and fail every per-locale cell (control C-5); the parent-keyed S-06c kills `moai spec queue`, `moai feedback setup`, and `moai profile audit` (control C-6); the marker-anywhere mutant of the QE-2 alternate fails S-H1a (control C-10).
- Traceability (revision 0.3.0): the Group 4 awk verb was refused by the worktree-session guard and not retried; the two readings were taken with plain commands (plan.md § F): 7 requirement definitions, 11 mapping rows, no uncovered requirement, no orphan. Derived, not the verb's `COLLECTED:` line.
- `moai spec lint` on spec.md (tree build, revision 0.3.0): `0 error(s), 0 warning(s)`, exit 0, one INFO `OwnershipTransitionUnmeasured` for the first plan commit 6d335cd483 (no `Authored-By-Agent` trailer; a historical commit that is not rewritten).
- `spec_audit` MCP tool (`filter_spec` this SPEC, `project_root` this worktree), revision 0.3.0: era V3R6 (H-5), one INFO finding (`EraAutoDetected`), `modern_era_clean: 1`, `total_specs: 1`, no MUST-FIX finding.

### Plan-phase gaps carried forward

- The card's named source evidence (`.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md`) is absent; the file and line list is verified against the current tree only.
- The Q13 and Q17 authority anchors are operator records, not committed files; under the authority register they are not citable as DECIDED or POLICY-COVERED, so those rows stay FOUNDER with the verdict recorded.
- The green paths of the content cells (S-11, S-12a-xx, S-14-xx, S-16-xx, S-17, S-18-xx) were observed on fixtures, not on the real pages; the docs-site pages are run-phase work (acceptance.md G-11, G-13).
- The Group 4 traceability verb and the CN-4 ordering verb were not run (guard refusal of inline `awk`); plain-command readings stand in for them (acceptance.md G-15).
- The native register of the ja, ko, and zh pinned prose (including the new labels) was not reviewed by a native reader (acceptance.md G-8).

## §E.2 Run-phase Evidence

_pending run-phase (manager-develop)_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
