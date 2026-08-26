# progress.md — SPEC-STAMP-REACHABILITY-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M canonical set (spec.md, plan.md, acceptance.md) + research.md + progress.md, written in worktree `.claude/worktrees/t291` (branch `WT-stamp-orphan-guard`, base `da791eb0a`), 2026-08-27.
- Requirements: 10 REQ (GEARS, pattern-annotated: REQ-SR-001..010); 10 AC (Given-When-Then) each carrying a RED-now baseline measured at base; REQ↔AC traceability both directions complete (acceptance.md §D.1). Counts below Tier M ceilings.
- Premise re-measurement (this run, verbatim outcomes): `git merge-base --is-ancestor 0d15864ae90b origin/main` rc≠0 (`NOT_ANCESTOR` — orphan fixture live); provenance `.commit_sha` read as `410da655f39d…` via jq; `git merge-base --is-ancestor 410da655f39d1d3731abf2f247e8e5353a9d0de5 origin/main` rc=0 (`CURRENT_STAMP_ANCESTOR` — guard-green baseline live). SPEC ID regex self-check executed as Bash: `PASS`; catalog uniqueness probe count 0.
- Countermeasure adjudication per lead dispatch: option 1 (CI pre-merge reachability guard) and option 2 (explicit stamp-commit mode) BOTH selected as in-scope milestones M2/M1; option 3 alternatives rejected-and-recorded only. Validation split recorded as a boundary decision: resolvability at stamp time, survivability enforced by the CI gate where it is decidable.
- Scope fences: #1666 recurrence classification stays with t278; provenance schema evolution; historical-orphan retrofit; mx/edges layers — all fenced in spec.md §E with H3 Out-of-Scope headings.
- Audit cycle: plan-audit iteration 1 verdict PASS-WITH-DEBT 0.81 (Clarity 0.85 / Completeness 1.00 / Testability 0.70 / Traceability 0.75) → iteration-2 remediation v0.2.0 applied D1-D7 (push-event anatomy AC-SP-011 + pinned single-mechanism decision, honest threshold-crossing recovery restatement, ≥ threshold-direction correction, REQ-SR-004 instrument AC-SP-012, measured-paste-executable D5 fixture, D6/D7 adoptions); all six guard-branch cells plus mutant cell executed and quoted (research.md §8), no unobserved cell remaining at plan phase except the post-M2 static anatomy claims recorded RED-now by construction.
- Constraint inheritance: predecessor REQ-GFR-014 (never restamp vs branch-local HEAD) binds this SPEC's own delivery; described-source churn below threshold means the committed stamp carries unchanged — at-or-above threshold accepts the pre-merge red until a POST-MERGE main restamp outside delivery scope (spec.md §D v0.2.0).
- Evidence paths cited across artifacts resolve within the worktree tree (triage-table.md tracked on base; code anchors at graph_stamp.go/provenance.go/check.go/graph_check.go line-cited in plan.md §H).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
