# progress.md — SPEC-GOAL-HTML-FLOW-001

> Lifecycle skeleton emitted at plan-phase. §E.1 (plan-phase audit-ready signal) is populated by manager-spec. §E.2 / §E.3 (run-phase evidence + audit-ready) are populated by manager-develop. §E.4 (sync-phase audit-ready) is populated by manager-docs. The literal `§E.2` / `§E.3` / `§E.4` heading tokens are parsed by `internal/spec/era.go` for V3R6 era classification — preserve them verbatim.

## §A. Plan-phase Artifacts

- spec.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/spec.md` (12-field frontmatter, GEARS REQ-GHF-001..010, §B Out of Scope with `### Out of Scope — <topic>` H3 sub-headings).
- plan.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/plan.md` (§A-§H, 6 milestones decision-reversibility-first).
- acceptance.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/acceptance.md` (§D matrix, 11 ACs GWT, §D.1 severity, §D.3 indirect, §D.4 closure, §D.5 forward-looking).
- progress.md: this file.

## §B. Plan-phase Tier

Tier M (4 artifacts: spec.md + plan.md + acceptance.md + this progress.md). plan-auditor PASS threshold: `0.80`.

## §C. Pre-plan Audit-ready Checklist (manager-spec self-verification)

- [x] SPEC ID regex `PASS` (verbatim): `ID="SPEC-GOAL-HTML-FLOW-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS` → `PASS`.
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags). No snake_case aliases.
- [x] `phase: "v3.1 target"` — release target label (NOT a lifecycle stage token; the prohibition on `plan`/`run`/`sync`/`mx` is respected).
- [x] `status: draft` — initial plan-phase emission.
- [x] Requirements in GEARS notation (Where / While / When / The <subject> shall).
- [x] Out of Scope section carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (5 sub-headings: Dashboard per-turn auto-refresh / Orchestrator AskUserQuestion simplification / Plan-auditor JSON sidecar / Mechanical re-arm pipeline / Web live dashboard v3.1 target).
- [x] Artifact set matches Tier M (spec.md + plan.md + acceptance.md + progress.md).
- [x] spec.md carries no implementation detail (functions named as consumer-facing API contracts only; no code).

## §D. ID Uniqueness

`SPEC-GOAL-HTML-FLOW-001` verified absent from `.moai/specs/` prior to artifact creation (no duplicates; no sibling SPEC claims the GOAL-HTML-FLOW domain).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-04
plan_tier: M
plan_artifact_count: 4
plan_artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
plan_auditor_verdict: pending
```

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_

sync_commit_sha: "pending-backfill-sync"

## §F Phase 4 Mode Selection

**Input parameters:**
- tier: M (spec.md frontmatter `tier: M`)
- scope (file count): ~8–10 files — NEW `internal/goal/dashboard.go`(+test), `internal/cli/goal.go`, `internal/goal/state.go`, `internal/goal/prune.go`, plan-HTML renderer(~2 files), `plan.html.mustache`, `spec-assembly.md`, `goal.md` workflow — within Tier M (5–15 files)
- domain count: 2 (internal/goal + internal/cli, plus skill/workflow markdown edits) — below Mode 4 ≥3 threshold
- file language mix: Go source (renderer, CLI, lifecycle) + markdown (skill/workflow) + mustache (template slot)
- concurrency benefit: LOW — coding-heavy TDD (sequential RED-GREEN-REFACTOR), per Anthropic's coding-task parallelism caveat

**Mode evaluation:**
- Mode 1 (trivial): not selected — multi-file feature, not a typo.
- Mode 2 (background): not selected — write-capable implementation, not read-only.
- Mode 3 (agent-team): RETIRED — never selected.
- Mode 4 (parallel): not selected — coding-heavy (concurrency benefit LOW); Mode 4 reserved for research-heavy multi-domain.
- Mode 5 (sub-agent): SELECTED — coding-heavy TDD, safe default per Anthropic's coding-task parallelism caveat; sequential per-milestone delegation.
- Mode 6 (workflow): not selected — not a high-volume mechanical transform; new-code work stays Mode 5.

**Decision:** sub-agent

**Justification:** Coding-heavy Go implementation (dashboard renderer, CLI verb, sibling lifecycle, plan-HTML report renderer) driven by TDD. Per Anthropic's finding that most coding tasks involve fewer truly parallelizable tasks than research, the sequential sub-agent path is the correct default. The 6 milestones are sequentially dependent (M5 re-arm UI consumes M1's renderer; M4 emission consumes M3 report) and cannot fan out. Implementation Kickoff Approval obtained in prior session (per resume); Mode 5 is strictly downstream of that gate.

## §H Recursive Self-Diagnosis Log

_<pending run-phase>_

## §I Token Accounting

_<pending sync-close>_
