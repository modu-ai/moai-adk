# Progress — SPEC-BACKLOG-JSON-DISCLOSURE-001

Card: t395 · Worktree `.claude/worktrees/t395` · Branch `WT-stale-backlog-json`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (3-file artifact set: spec.md + plan.md + acceptance.md; plan-auditor PASS threshold 0.80)
- Requirements: 14 (ceiling 16) · Acceptance criteria: 16 (ceiling 16)
- Artifacts emitted: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- Status: `draft` — awaiting Implementation Kickoff Approval
- Evidence base: `.moai/reports/t395/{premise-verdict,reader-surfaces,r1-repro}.md`

### Plan-audit iterations

| Iter | Verdict | Score | Outcome |
|---|---|---|---|
| 1 | PASS-WITH-DEBT | 0.85 (Tier M threshold 0.80) | 5 blocking-class + 3 optional defects — `.moai/reports/t395/plan-audit.md` |
| 2 | PASS | 0.94 | iter-1's 8 defects all RESOLVED; 3 new optional raised — `.moai/reports/t395/plan-audit-iter2.md`. spec.md 0.1.0 → 0.2.0 |
| 3 | — | — | D9 closed (one clause). D10/D11/D12 accepted as residual risk by lead decision. spec.md 0.2.0 → 0.2.1 |

Iteration 2 closures: D5 (AC-BJD-015 single-regex blindness → two enumerated
greps), D2 (AC-BJD-007 → runnable commands), D4 (AC-BJD-010 Given → constructible
and self-evidencing, unbuilt Given = Gap), D3 (verb breadth unified at the
read-surface floor; open decision recorded in `plan.md` §D.1), D1 (rebinding →
directory-based), D6 (three files / four sites; REQ-BJD-003 `Where` → `While`),
D7 (§A.2 R5 citation replaced with the two greps that actually cover the sites),
D8 (AC-BJD-002 count scoped to this SPEC's line + archive-tables Given).

Carried forward, not closed here: **the disclosure breadth beyond the read
surface** is an open operator decision (`plan.md` §D.1). No clarification-gate
marker is placed — the run proceeds completely at the floor, and a later answer
widens rather than contradicts.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
