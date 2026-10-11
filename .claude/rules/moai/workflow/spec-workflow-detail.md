---
description: "Detail companion for spec-workflow.md — the Route A/B step tables with the merge_method footnote, the [SHOULD] SPEC-phase anti-patterns, the DDD/TDD methodology mode bodies with self-review/drift/delegation, the Phase 1 Depends_on Pre-flight Check, and the Phase 1 Report Persistence two-stream contract"
paths: "**/spec-workflow.md,**/.claude/rules/moai/workflow/spec-workflow-detail.md"
---

# SPEC Workflow — Detail Companion

> Detail companion of `spec-workflow.md`. The stub keeps every [HARD] rule and pointer; this file owns the moved detail bodies: the Route A/B step tables and the `merge_method` footnote, the [SHOULD] SPEC-phase anti-patterns, the DDD/TDD methodology bodies (DDD / TDD modes, Pre-submission Self-Review, Drift Guard, Methodology delegation), the Phase 1 Depends_on Pre-flight Check, and the Phase 1 Report Persistence contract. Load when executing a route's step table, choosing or running a development methodology, running the Phase 1 pre-flight, or resolving a plan-audit report-stream question.

## Route A/B Step Tables

Step tables for the two SPEC lifecycle routes defined in `spec-workflow.md` § SPEC Phase Discipline; the stub keeps the [ZONE:Frozen] [HARD] route definitions and step-ordering rules.

**Route A — Hybrid Trunk main-direct (default, Tier S / M):**

| Step | Location | Command | Branch | Merge strategy | Lifecycle event (trigger) |
|------|----------|---------|--------|----------------|---------------------------|
| 1 (plan) | main checkout | `/moai plan SPEC-XXX` | `main` (direct) | n/a (no PR) | plan-phase artifacts committed + pushed to `main` |
| 2 (run)  | main checkout | `/moai run SPEC-XXX` | `main` (direct) | n/a (no PR) | run-phase commits pushed to `main` + tests green |
| 3 (sync) | main checkout | `/moai sync SPEC-XXX` | `main` (direct) | n/a (no PR) | single sync commit pushed to `main` (carries `implemented → completed`) |

**Route B — PR route (Tier L OR explicit `--pr`):**

| Step | Location | Command | Branch | PR strategy | Lifecycle event (trigger) |
|------|----------|---------|--------|-------------|---------------------------|
| 1 (plan) | main checkout | `/moai plan SPEC-XXX` | `plan/SPEC-XXX` | configured* | plan PR merged into main |
| 2 (run)  | main checkout (default) OR L2 SPEC worktree (opt-in) | (opt-in) `moai cc -w SPEC-XXX` then `/moai run SPEC-XXX` inside it; OR `/moai run SPEC-XXX` on `feat/SPEC-XXX` branch in main checkout | `feat/SPEC-XXX` | configured* | run PR merged into main |
| 3 (sync) | same as Step 2 | `/moai sync SPEC-XXX` (same L2 worktree as Step 2 if L2 was used; otherwise same feature branch) | `sync/SPEC-XXX` (or `chore/SPEC-XXX-sync`) | configured* | sync PR merged into main |
| 4 (cleanup) | host checkout (only if L2 was created) | `moai worktree done SPEC-XXX` | n/a | n/a | L2 worktree disposed |

\* Route B PR strategy is the configured `merge_method` (`git_strategy.<mode>.merge_method`; one of `squash` | `merge` | `rebase`), **default `squash`**. Squash remains the documented recommendation — one squash commit per phase yields clean, revertable SPEC history — and is the value applied when `merge_method` is absent or unset. The method is configurable (per the per-mode `merge_method` field) so that workflows such as gitflow `release/*` may opt into a merge commit; the FROZEN default and its rationale are unchanged. Route A has no PR and therefore no `merge_method` — it pushes directly to `main`. Step 4 (worktree cleanup) applies to Route B only when an L2 worktree was created.

## Anti-patterns — SPEC Phase Discipline

[SHOULD] Anti-patterns (advisory):
- Creating an L2/L3 worktree for plan (Step 1). Plan-in-worktree forces a base rebase after plan PR merge and prevents parallel SPEC plan visibility.
- Stacking plan + run in the same L2 worktree. Once the plan PR merges, the worktree base becomes stale; subsequent run work either rebases (extra cost) or proceeds against a stale tree (correctness risk).
- Disposing the L2 worktree after run merge but before sync merge. Sync re-enters the tree with codemap / MX / docs writes; the host checkout cannot stand in for a disposed worktree.

Cross-reference: see `.claude/rules/moai/workflow/worktree-integration.md` § SPEC-to-Worktree Mapping for per-step L2 worktree applicability and decision tree.

## Run-phase Methodology

Methodology mode bodies of `spec-workflow.md` § Run Phase (Development Methodology); the stub keeps the Methodology Auto-Detection table, MX Tag Integration, and the Re-planning Gate.

### DDD Mode — ANALYZE-PRESERVE-IMPROVE

Best for existing projects with < 10% test coverage. Uses manager-develop agent with cycle_type=ddd.

**ANALYZE**: Read existing code, map domain boundaries, identify side effects and implicit contracts.
**PRESERVE**: Write characterization tests capturing current behavior. Create behavior snapshots for regression detection.
**IMPROVE**: Make small incremental changes. Run characterization tests after each change. Refactor with test validation.

### TDD Mode — RED-GREEN-REFACTOR (default)

Best for all development work, new projects, and brownfield with 10%+ coverage. Uses manager-develop agent with cycle_type=tdd.

**RED**: Write a failing test describing desired behavior. Verify it fails. One test at a time.
**GREEN**: Write simplest implementation that passes. No premature optimization.
**REFACTOR**: Clean up while keeping tests green. Extract patterns, remove duplication.

Brownfield enhancement: Pre-RED step reads existing code to understand current behavior before writing the failing test.

### Pre-submission Self-Review

Before marking implementation complete: review full diff against SPEC acceptance criteria. Ask "Is there a simpler approach?" and "Would removing any changes still satisfy the SPEC?" Skip for single-file changes under 50 lines, bug fixes with reproduction test, or user-approved annotation cycle changes.

### Drift Guard

After each methodology cycle, compare planned files against actual modifications. Warns at <= 30% drift. Triggers re-planning (Phase 14) above 30%.

### Methodology delegation (team mode experimental)

The run-phase methodology (DDD/TDD) is applied by a single `manager-develop` sub-agent (serial), with multi-domain research fanned out via fanout (parallel read-only `Agent()`) where warranted; the Agent Teams layer is an explicit-request experimental alternative (see § Agent Teams Variant). Native Agent Teams remain experimental; retired CG routing does not establish mixed-provider teammate capability.

## Depends_on Pre-flight Check

The Depends_on Pre-flight Check is the first sub-step of Phase 1, executed BEFORE the plan-auditor subagent invocation — sub-step 0, not a separate phase (no phase inflation).

**Procedure:**
1. Load the SPEC's frontmatter `depends_on:` list (Optional field per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields).
2. Where `depends_on` is absent or empty, the pre-flight trivially PASSes and proceeds to the plan-auditor step.
3. Where `depends_on` lists one or more SPEC IDs, resolve each dependency's current `status:` frontmatter field by reading `.moai/specs/<dep-ID>/spec.md`.

**Fulfillment definition (strict):** dependency fulfillment is defined as the dependency SPEC's `status: completed` — all other 7 status values (draft, planned, in-progress, implemented, superseded, archived, rejected) are considered unfulfilled. The evaluation is deterministic per-status: no partial credit, no "near-completed" interpretation, no score-based bypass.

**Blocker on unfulfilled dependency (3-option):** When one or more `depends_on` entries are unfulfilled, the pre-flight SHALL NOT proceed to the plan-auditor step. The orchestrator SHALL surface a structured blocker via `AskUserQuestion` with three options:
- **wait** — abort run; re-invoke after deps complete
- **override** — proceed with `--ignore-deps` flag; logged to `.moai/logs/depends-on-override.log` (the override path MUST record the unfulfilled dependency IDs + override rationale in the log; a bare `--ignore-deps` without the logged rationale is prohibited)
- **abort** — cancel run

The `--ignore-deps` flag and `.moai/logs/depends-on-override.log` path are literal tokens. The pre-flight is orchestrator-side doctrine; Go implementation is deferred to a follow-up SPEC.

## Report Persistence

Two report streams exist for plan audits; they are distinct by design, cross-referenced with `.claude/agents/moai/plan-auditor.md` § Output Format, and do NOT share a directory:

- **plan-phase review stream** — `plan-audit.md` (or `plan-audit-iter<N>.md`, one file per iteration), exported by the plan-auditor to the card evidence path `.moai/reports/<card-id>/` (or `.moai/reports/<SPEC-ID>/` for a SPEC-scoped audit produced without a card) per the audit-artifact convention (`.moai/docs/audit-artifact-convention.md`). Iteration `N` follows the plan-auditor Retry Loop Contract (max 3). Consumed by the plan workflow's assembly/annotation cycle.
- **run-gate stream** — `<SPEC-ID>-<YYYY-MM-DD>.md`, date-based, under the gitignored runtime record directory `.moai/reports/plan-audit/`. Written by the Phase 1 Plan Audit Gate (`internal/runtime/audit_report.go`). Every gate call persists a record here; multiple calls on the same day append to the same file. This date-file is the verdict **record surface** only — it is never the hash subject for skip-eligibility (see below).

Skip-eligibility inputs (normative, matching the Go implementation): (a) the "most recent plan-auditor verdict" the run-gate consults is the plan-phase review stream's **final-iteration verdict**, resolved by `runtime.ResolveLatestPlanAudit`; (b) the artifact-hash check recomputes and compares the **plan-artifact hash** — `internal/runtime/audit_cache.go` `ComputeHash` hashes the SPEC directory's plan artifacts (the union subject set below) as exact bytes, with cache key = (specID, planArtifactHash); (c) the run-gate stream's date-file records the verdict but is not hashed and never supplies cache identity. A review file without hash/score/version metadata is a cache miss and requires a fresh audit.

**Plan-artifact hash subject list (Go verbatim):** the hash subject set is the union `{acceptance.md, decision-index.md, design.md, plan.md, research.md, spec.md, tasks.md}` — matching `internal/runtime/audit_cache.go` `planArtifactNames` verbatim. The set is tier-conditional by construction via the "skip if missing" rule in `ComputeHash`: a Tier S directory (spec.md, plan.md) hashes only those present; a Tier M directory adds acceptance.md; a Tier L directory contributes design.md AND research.md as mechanical subjects (SPEC-AUDIT-SNAPSHOT-001 A1 — design.md/research.md changes now mechanically invalidate a cached skip verdict); a grandfathered V3R4 directory carrying tasks.md retains it as a subject (K-2 backward compat); a directory carrying `decision-index.md` hashes it too, so a decision-index edit after the audited SHA invalidates the audit — this intentionally widens the skip-cache key, and an existing SPEC with a decision-index sees one cache miss and is re-audited.

**Amendment as cache-invalidating event:** when a SPEC is amended in-place per the `completed → in-progress (amendment)` transition (completed → in-progress, `## Amendments` HISTORY row added), the plan-artifact hash changes because `spec.md` is modified — this is a cache-invalidating event that invalidates any cached plan-auditor PASS verdict for the SPEC, forcing Phase 1 plan-audit re-execution on the next `/moai run`. During the amendment transition the SPEC stays V3R6 modern era: frontmatter status `in-progress` (not `completed`) keeps the `internal/spec/audit.go` completed-no-drift predicate from firing.

Reports in both streams are local artifacts (gitignored).

---

Version: 1.0.0 (split from spec-workflow.md — core stub + detail companion)
Classification: Detail companion — paths-scoped, never part of the always-loaded surface.
