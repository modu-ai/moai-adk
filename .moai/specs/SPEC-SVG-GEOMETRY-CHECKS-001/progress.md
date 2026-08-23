# SPEC-SVG-GEOMETRY-CHECKS-001 — Progress

Card: t166 · Worktree: `.claude/worktrees/t166` · Branch: `WT-verify-geometry`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- SPEC ID regex check executed as Bash: `PASS`.
- Frontmatter: all 12 canonical fields present; `status: draft`; `phase: "v3.1.3 target"`.
- Out of Scope: four `### Out of Scope — …` sub-headings with bullets (spec.md §D).
- Requirements: 16 REQ entries in GEARS notation (spec.md §C) — at the Tier M ceiling of 16.
- Acceptance criteria: 16 `AC-SGC-xxx` Given-When-Then entries — at the Tier M ceiling of 16 — plus the §D.9 boundary-value table
  (acceptance.md §D).
- Premises of the delegation prompt were re-verified against the tree: script length 609 lines,
  fixture pair present and behaving 0/1, mirror `diff -rq` clean. The "§2.5 rule text settled"
  premise proved **partially false** and is now recorded as spec.md §B D7 rather than assumed.
- Audit iteration 1: FAIL 0.74 (Tier M threshold 0.80), nine blocking findings. Iteration 2 applied
  all nine plus the routed optional findings, entirely as clauses of the existing 16 requirements.
- Audit iteration 2: **PASS 0.81**, no dimension regressing, with four blocking-class defects
  recorded as mandatory pre-run-phase debt (N1-N4) plus four optional (N5-N8). All eight are cleared
  in v0.3.0, again as sentence-level edits inside existing entries — counts unchanged at 16 REQ /
  16 AC, ids contiguous 001-016, no dangling cross-reference, no stale `§ D.8`.
  Every measurement either auditor cited was independently reproduced before the fix was applied.
- Debt status at run-phase entry: **cleared**. Remaining accepted costs are stated, not latent —
  K3's bounded 11-16 unit exception, the departure-side and marker-less C4 blind spots (§A), and
  the C2 association window's reach (§A).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier M; scope ~5 files changed plus ~20 fixtures added; domain count 1 (one skill
subsystem: a Node lint script, its fixtures, its references, and the template mirror); file language
mix JavaScript plus markdown plus SVG; concurrency benefit LOW (coding-heavy, and M2 depends on M1's
fixtures existing).

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | Not trivial — a geometry engine, a reader, and ~20 fixtures |
| `serial` | **yes** | Coding-heavy, single domain, milestones strictly ordered (M1 RED gates M2 GREEN) |
| `fanout` | no | 1 domain, not research-heavy; Anthropic's coding-task parallelism caveat applies |
| `sweep` | no | Not ≥ ~30 files and not one uniform mechanical transform rule |

Decision: `serial`

Justification: the work is one coding-heavy change to one script inside one skill, and its
milestones are dependency-ordered rather than parallel — M1's fixtures must exist and fail before
M2's engine can be shown to turn them green. Fan-out would add reconciliation cost with no
independent work to reconcile, and the coding-task parallelism caveat argues against it directly.
Progression mode: semi-autonomous per operator selection at the Implementation Kickoff Approval
gate — the orchestrator reads evidence at each milestone boundary; no goal is armed.
