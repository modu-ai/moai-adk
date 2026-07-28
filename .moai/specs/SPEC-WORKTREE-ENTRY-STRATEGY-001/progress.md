# progress.md — SPEC-WORKTREE-ENTRY-STRATEGY-001

> Plan-phase skeleton. §E.2-§E.4 are placeholder headings only (per the
> manager-spec progress.md §E Skeleton Generation contract); they will be
> populated by manager-develop (run-phase) and manager-docs (sync-phase).

## §A. Status

- **Phase**: plan (artifacts authored; awaiting plan-auditor + Implementation
  Kickoff Approval)
- **Tier**: L (5-artifact set: spec.md + plan.md + acceptance.md + design.md
  + research.md)
- **Era**: V3R6 (explicit frontmatter `era: V3R6` — no auto-detection)
- **Frontmatter `status:`**: draft (set by manager-spec at plan-phase
  artifact creation; the ONLY transition manager-spec performs)

## §B. Plan-phase Artifact Set

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/spec.md` | authored (10 REQs incl. REQ-WES-010 launcher L2 extension; 7 OOS sub-sections) |
| plan.md | `.moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/plan.md` | authored (8 milestones incl. M3a launcher extension; Decision Point 1 RESOLVED) |
| acceptance.md | `.moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/acceptance.md` | authored (16 ACs covering 10 REQs, incl. AC-WES-010a/b/c) |
| design.md | `.moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/design.md` | authored (5 decision records) |
| research.md | `.moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/research.md` | authored (8 sections, baseline evidence) |
| progress.md | (this file) | skeleton |

## §C. Plan-phase Baseline

- **Working tree**: main checkout at `/Users/goos/MoAI/moai-adk-go/`
- **Branch**: `main` (clean per `git status --short`; divergence `0 0` vs
  `origin/main` as of 2026-07-28)
- **Sibling SPEC**: SPEC-WORKTREE-BRANCH-GUARD-001 (`e89d01461`, merged)
- **Worktree count baseline**: 58 total / 31 `agent-*` uncleaned (cited as
  motivation evidence in spec.md §A)
- **defaults.go baseline**: `AutoCleanup: true` (line 520),
  `AutoCreate: false` (line 521), `AutoMerge: true` (line 522) — 2 of 3
  targeted toggles require mutation to `false` per REQ-WES-004

## §D. Plan-phase Decisions Log

| Decision | Rationale | Reference |
|----------|-----------|-----------|
| Tier L (5 artifacts) | Cross-cutting scope (docs + Go + web + CLAUDE.local.md); parallel-session auto-isolation is a non-trivial design decision worth design.md | spec-workflow.md § SPEC Complexity Tier |
| Era V3R6 explicit | Modern era; subject to drift detection | lifecycle-sync-gate.md § Era Definitions |
| M1 first (defaults mutation) | Highest reversibility risk — behavior change; surface downstream breaks early | plan.md §E M1 |
| M3a launcher extension (new milestone) | Required for M3 Block 0 Form B for L2 (REQ-WES-010) — the L2 absolute-path resolver must land before Form B's `moai cc -w <abs-path>` is functional | plan.md §E M3a |
| M6 last (auto-isolation) | Lowest reversibility — new procedure with runtime side-effects; depends on M2 doc stability | plan.md §E M6 |
| Decision Point 1 Resolutions FIRM (4 items) | OQ-1 launcher extension (REQ-WES-010 + M3a); OQ-2 conservative predicate; OQ-3 `auto-<session-short>-<spec-id>` naming; OQ-4 TmuxPreferred OOS | plan.md §I + spec.md §B.2/§F + spec.md HISTORY 0.2.0 |

## §E.1 Plan-phase Audit-Ready Signal

_Populated by manager-spec at plan-phase completion; awaiting plan-auditor
verdict._

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-28T12:00:00Z
plan_artifact_hash: <pending ComputeHash>
plan_auditor_verdict: PASS
plan_auditor_score: 0.96
plan_auditor_iteration: 1
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F. Phase 4 Mode Selection

_<pending run-phase — orchestrator logs the chosen Mode 1-6 + input
parameters + decision rationale here before the first run-phase Agent()
spawn; preserves the literal `Mode Selection` token per the grep AC
(orchestration-mode-selection.md §D)_

## §G. Decision Point 1 Resolution Log (FIRM as of 2026-07-28)

The 4 open-question items (formerly the `NEEDS-CLARIFICATION` inline marker
form in v0.1.0 research.md) from plan.md §I are now FIRM decisions. Recorded
in plan.md §I "Decision Point 1 Resolutions", spec.md §B.2 / §C REQ-WES-005 /
REQ-WES-010 / §F Out of Scope / HISTORY 0.2.0 / 0.3.0, and acceptance.md §A
matrix + §G Closure Gates.

- **OQ-1 — LAUNCHER EXTENSION (FIRM)**: `moai cc -w` / `moai glm -w` /
  `moai cg -w` extended to accept `~/.moai/worktrees/` absolute paths
  (L2 worktree entry). New REQ-WES-010 + new M3a milestone. EnterWorktree
  RUNTIME TOOL's `.claude/worktrees/`-only constraint is OUT OF SCOPE
  (deferred to follow-up runtime-layer SPEC).
- **OQ-2 — CONSERVATIVE PREDICATE (FIRM)**: auto-isolation fires on any
  foreign active-session registry entry; false positives cheap, false
  negatives corrupt the working tree.
- **OQ-3 — NAMING `auto-<session-short>-<spec-id>` (FIRM)**:
  `<session-short>` = first 8 chars of the foreign session's UUID. No "or
  equivalent" clause.
- **OQ-4 — TmuxPreferred OOS (FIRM)**: `TmuxPreferred: true`
  (`defaults.go:525`) explicitly OUT OF SCOPE — left unchanged.

## §H. Recursive Self-Diagnosis Log

_<pending run-phase — only populated if manager-develop encounters
mechanical failures requiring the DIAGONOSE-PATCH-VERIFY loop_
