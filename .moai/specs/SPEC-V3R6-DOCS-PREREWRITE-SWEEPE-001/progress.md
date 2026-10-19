# Progress — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-audit-pending
plan_revision: 0.2.0
plan_commit: 6d335cd483 (feat(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan-phase artifacts (tier M, 3 artifacts), card t1725)
plan_followup_commits: b28da6f891 (progress.md and decision-index.md), bb7d928f68 (cite REQ in ACs, QE-1 relabel)
revision_commit: pending-backfill (a commit cannot cite its own hash; the revision-1 commit SHA is reported to the orchestrator and backfilled by the leader's record)
card: t1725 (Class C, docs-only), lane-26, run tn1p4e
parent: SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (child row SWEEPE; decisions Q13, Q17, Q24/N2)
measurement_pin: 74ad197380a0 (the tree the plan-phase counts were taken on; full 74ad197380a0711cdb9816d4e37c94d29af730ea; docs-site, .claude, internal, cmd, and scripts are identical to HEAD bb7d928f68, acceptance.md S-00)
plan_audit: iteration 1 FAIL 0.69 (tier M threshold 0.80, MP-8 failed), verdict file `.moai/reports/t1725/plan-audit-verdict.md`, audited_sha bb7d928f682d34279d5d3584537fec0c3e68dd25, receipt rcpt-93c8df0b6680dfc7dac83fb9; revision 0.2.0 answers defects D1 to D12; iteration 2 (delta on D1 to D6) not yet run
open_decisions: see decision-index.md. Q13, Q17, and Q24 carry operator verdicts (operator records d-20261019T015908Z-220a for Q13 and Q17, d-20261014T060405Z-0fcd for Q24). QE-1 to QE-5 carry DEFAULT-APPLIED planner verdicts (2026-10-19T02:27:27Z, manager-spec); the leader records the default application on the decision board before Kickoff.

### Plan-phase measurements (source: .moai/reports/t1725/progress.md, manager-spec, 2026-10-19, pin 74ad197380a0; revision 0.2.0 re-measured every cell it touched, acceptance.md § D)

- Closing pattern (umbrella E-04) on the twelve page files: 10 rows, 18 lines (faq x4 `/moai init`; moai-feedback x4 `fb|bug|issue`; en foundations line 39 `relationship`; en verification-claim-integrity line 84 `brain`). The ja and zh foundations files print no row.
- Complement census (`-P`, CJK-safe boundary): the same 18 lines; no unknown word beyond the six.
- Router Priority 1 (SKILL.md lines 69-70, repo and template copies): `init` is an alias of `project`, `fb` of `feedback`; `bug`, `issue`, `relationship`, `brain` are not. So the FAQ claim that `/moai init` "does nothing" is false, and the umbrella's word "unregistered" is looser than router resolution (QE-3).
- `moai brain` (CLI form, en verification-claim-integrity line 84) exits 1 with `Unknown command` on a tree build; umbrella families E-01..E-03 miss it. E-01..E-03 are 0 in the page files. Every other inline `moai <word>` token resolves (S-06b, S-06c, S-06d).
- Neither en faq line 24 nor en verification-claim-integrity line 84 carries an example marker; no HTML comment exists in either file.
- ja foundations line 40 and zh foundations line 40 translate the edited English clause; the ko row has no `/moai` clause (S-07, QE-4).
- Tree build VCS stamp (478285ab4c1f, modified) differs from HEAD; not used as provenance.
- Operator record d-20261019T015908Z-220a read through `moai decision read --scope card:t1725` (revision 0.1.0): Q13 replace-else-delete, Q17 four locales.
- Traceability verb on this directory (revision 0.2.0): `COLLECTED: 7 REQ definitions (acceptance input: read)`, no UNCOVERED, no ORPHAN.

### Plan-phase gaps carried forward

- The card's named source evidence (`.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md`) is absent; the file and line list is verified against the current tree only.
- The Q13 and Q17 authority anchors are operator records, not committed files; under the authority register they are not citable as DECIDED or POLICY-COVERED, so those rows stay FOUNDER with the verdict recorded.
- The green paths of the content cells were observed on fixtures, not on the real pages; the docs-site pages are run-phase work (acceptance.md G-11).

## §E.2 Run-phase Evidence

_pending run-phase (manager-develop)_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
