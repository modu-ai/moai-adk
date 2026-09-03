# Progress — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Branch `WT-queue-upgrade-proof` · Base `4e4607abe`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts written: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- Tier: M (4-file plan-phase set; no `design.md` / `research.md`)
- Status: `draft`
- Requirements: 10 (`REQ-QUP-001`..`010`), one optional (`REQ-QUP-007`)
- Acceptance criteria: 11 rows (`AC-QUP-001a`..`010`), one optional
  (`AC-QUP-007`); every row now names a requirement (no orphan)
- Open clarifications at v0.3.0: 2 — `[NEEDS CLARIFICATION: G2 definition]` and
  `[NEEDS CLARIFICATION: downgrade intent vs quarantine rename]`, both in
  `plan.md §A`; both RESOLVED at v0.4.0 (see the closing entry below)
- Plan audit: iteration 1 returned FAIL (score 0.875 vs Tier M threshold 0.80;
  cause was the MP-3 frontmatter defects and the MP-7 clarification gate, not
  the score). Verdict: `.moai/reports/t470/plan-audit.md`
- Remediation landed at SPEC `v0.2.0`: D1 (`tags` sequence → string), D2
  (`lifecycle` enum), D3 (`AC-QUP-010` mutation replaced with one that produces
  RED, plus a positive precondition and relocation sentinel on `AC-QUP-002`),
  D4 (`AC-QUP-008`'s gitignored `git status` limb replaced with a file digest
  comparison), D6 (`REQ-QUP-010` added; `AC-QUP-010` no longer an orphan), and
  the optional D7/D8/D9/D10. No production file touched — `REQ-QUP-009` holds
- Plan audit: iteration 2 returned FAIL (score 0.9625, monotonic up from 0.875;
  above the Tier M threshold 0.80). Cause was MP-7 alone. Verdict:
  `.moai/reports/t470/plan-audit-iter2.md`. Tier M iteration ceiling (2) reached
- Remediation landed at SPEC `v0.3.0`: D11 (`AC-QUP-008` + its twin constraint
  `C-1` named the live queue repository-relative, which from a linked worktree
  resolves to an absent file — both now derive the PRIMARY checkout's path the
  way `todo_root.go:95-99` does, and a failed derivation FAILS rather than
  passing) and the optional D12 (`AC-QUP-002`'s "holds the queue" limb given a
  stated observation). No production file touched — `REQ-QUP-009` holds
- Clarification gate CLOSED at SPEC `v0.4.0` (D5 resolved). Both markers in
  `plan.md §A` are converted to RESOLVED records — question retained, answer
  stated, source named (the dispatcher's ruling on card `t470`), consequence
  stated; neither marker was edited out. G2 is ABSORBED into G1 (carried by
  `AC-QUP-001a`/`001b`/`002`/`003`/`004`/`006`; the "closes as unstarted"
  contingency is withdrawn). The downgrade marker's earlier mechanism was WRONG
  and is corrected — the `.migrated` rename never contradicted the downgrade
  intent (`export-json` re-creates `backlog.json`); the real hole is that the
  export lands in the NEW directory while a v3.1.2 binary reads the legacy one,
  ruled OUT OF SCOPE as a separate-card candidate. G4 was newly supplied and is
  likewise OUT OF SCOPE, filed in `spec.md §E` beside G3 and G5. `AC-QUP-008`
  gained a hand-verification note (worktree guard refuses the nested `$(...)`
  form) with a matching pointer on its twin constraint `C-1`. **MP-7's blocking
  condition is now cleared.** No production file touched — `REQ-QUP-009` holds
- Open clarifications: 0 (was 2)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
