# Progress — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Branch `WT-queue-upgrade-proof` · Base `4e4607abe`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts written: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- Tier: M (4-file plan-phase set; no `design.md` / `research.md`)
- Status: `draft`
- Requirements: 10 (`REQ-QUP-001`..`010`), one optional (`REQ-QUP-007`)
- Acceptance criteria: 11 rows (`AC-QUP-001a`..`010`), one optional
  (`AC-QUP-007`); every row now names a requirement (no orphan)
- Open clarifications: 2 — `[NEEDS CLARIFICATION: G2 definition]` and
  `[NEEDS CLARIFICATION: downgrade intent vs quarantine rename]`, both in
  `plan.md §A`
- Plan audit: iteration 1 returned FAIL (score 0.875 vs Tier M threshold 0.80;
  cause was the MP-3 frontmatter defects and the MP-7 clarification gate, not
  the score). Verdict: `.moai/reports/t470/plan-audit.md`
- Remediation landed at SPEC `v0.2.0`: D1 (`tags` sequence → string), D2
  (`lifecycle` enum), D3 (`AC-QUP-010` mutation replaced with one that produces
  RED, plus a positive precondition and relocation sentinel on `AC-QUP-002`),
  D4 (`AC-QUP-008`'s gitignored `git status` limb replaced with a file digest
  comparison), D6 (`REQ-QUP-010` added; `AC-QUP-010` no longer an orphan), and
  the optional D7/D8/D9/D10. No production file touched — `REQ-QUP-009` holds
- D5 is deliberately NOT remediated: both clarification markers stay open in
  `plan.md §A` for the dispatcher. MP-7 remains failed until those answers
  arrive, which is the expected state

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
