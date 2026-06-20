# Progress — SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001

Lifecycle: plan → run → sync (3-phase). Tier L. cycle_type=ddd.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set authored by manager-spec (2026-06-20): spec.md + plan.md + acceptance.md + design.md + research.md + this progress.md (§E skeleton). Status: draft. era: V3R6.

- 5-artifact Tier L set complete; GEARS REQ-WADA-001..017 across 4 groups + cross-cutting template-mirror.
- AC matrix (acceptance.md §D) AC-WADA-001..017, each with a both-tree-verifiable grep/test command.
- Evidence baseline captured (research.md §B) with observed grep/test output; non-reproductions recorded honestly (OQ-1 release-update absent).
- Out of Scope section present (spec.md §F) with `### Out of Scope —` H3 sub-headings.
- SPEC ID self-check: `decomposition: SPEC ✓ | V3R6 ✓ | WORKFLOW ✓ | AGENT ✓ | DOC ✓ | ALIGN ✓ | 001 ✓ → PASS`.
- Awaiting: plan-auditor independent audit → Implementation Kickoff Approval (Tier L human gate) → run-phase.

## §E.2 Run-phase Evidence

### M1 — ANALYZE: occurrence inventory + replacement map

Pre-flight greps (re-run against working tree, sync 0 0 vs origin/main):
- `ls .claude/agents/` → `harness/ local/ moai/` (NO `builder/` — D1 confirmed).
- Raw total (ARCHIVED12 incl researcher): **139 lines** across **38 files** (matches plan-phase).
- `researcher` count: **14** (KEEP — live role_profile baseline).
- Archived-proper (ARCHIVED11) total: **125 lines** across **36 files** (matches plan-phase ~125).

Replacement map (migration table `archived-agent-rejection.md §C` drives all):

| Archived agent | Canonical replacement (this SPEC) |
|----------------|------------------------------------|
| `manager-strategy` | `manager-spec` (planning IS strategy — row 1) |
| `manager-quality` | sync-auditor / orchestrator verification batch / Stop hook / `gh issue create` (feedback.md, row 2) |
| `manager-brain` | `Explore` then `manager-spec` chain (row 3) |
| `manager-project` | `manager-docs` (row 4) |
| `expert-backend/frontend/devops/performance/refactoring` | per-spawn `Agent(general-purpose)` with domain whitelist (rows 7-12) |
| `expert-security` | Stop hook dependency-manifest audit OR per-spawn `Agent(general-purpose)` security reviewer (row 9) |
| `researcher` | **KEEP** — live role_profile (NOT replaced) |

Intent classification per occurrence: spawn target (replace) | documentation pointer to rejection rule (keep) | frontmatter `agents:` list (replace/remove) | role_profile (keep).

Mirror status per file group:
- MIRRORED (apply both trees): all `workflows/**`, `references/**`, `team/**` (except dev-only), all 6 `agents/moai/**`.
- DEV-ONLY no-mirror (live only): `workflows/release.md`/`github.md`/`release-update.md` — none carry archived-proper refs (release-update token absent).
- BASELINE-DIVERGED (changed-line mirror only): `manager-spec.md`/`manager-docs.md`/`manager-git.md`.

Group 2/3/4 baseline:
- `team-reader`/`team-validator` invalid subagent_type: glm.md (model table), debug.md, review.md (+frontmatter), plan.md (+frontmatter) — fixed M4 (REQ-WADA-009).
- `§ Terminology Glossary` absent in worktree-integration.md; 2 cross-refs (spec-workflow.md:23, worktree-state-guard.md:19) → M4 ADD section (REQ-WADA-010).
- GLM models `glm-5.1`/`glm-4.7`/`glm-4.5-air`/`glm-4.7-flashx` in glm.md:148-150,168-172 → M5 `glm-5.2[1m]` (REQ-WADA-011).
- Retired terms: `sprint contract` (workflows/run.md:80), `Round split` (team/run.md:69) → M5 (REQ-WADA-012).
- SKILL.md footer `Version: 2.6.0` / `Last Updated: 2026-02-25` (lines 342-343) → M5 bump (REQ-WADA-013).
- `release-update` token: **ABSENT** → AC-WADA-008 vacuously satisfied (OQ-1 confirmed not-reproduced).
- run.md: **246 LOC** both trees → M6 trim < 200 (REQ-WADA-014/015).

### M2 — workflow skill files archived-agent purge

26 files purged (workflows/** + references/{reference,anti-patterns,mx-tag}.md), both trees identical (changed-line mirror via edit-then-cp). Replacements per migration table §C:
- `manager-strategy` → `manager-spec` (phase-execution.md Phase 1/1.5, mode-orchestration.md).
- `manager-quality` → sync-auditor (review.md, moai.md, task-decomposition.md, quality-gates-quality.md, doc-execution.md) / manager-develop fixes (loop.md, fix.md, quality-gates-context.md, delivery.md) / Stop hook (gate.md, security.md, sync.md prose) / orchestrator `gh issue create` (feedback.md REQ-WADA-005).
- `manager-brain` → `Explore` + `manager-spec` (brain.md frontmatter).
- `manager-project` → `manager-docs` (project.md frontmatter, dedup).
- `expert-*` → per-spawn `Agent(general-purpose)` domain specialists (clean.md refactoring, review.md/clarity-interview.md/design.md/e2e.md frontend, security.md/quality-gates-quality.md security, doc-generation.md devops, loop.md/fix.md/mx.md/moai.md domain mix).
- design.md frontend role cross-references the FROZEN `design/constitution.md` carve-out note.
- `gh issue create` preserved in feedback.md (AC-WADA-005: count=2).
- residual grep (both trees, excl rejection-rule + dev-only trio): **0**. researcher KEEP: **14** (unchanged).

### M3 — team files + agent bodies + frontmatter purge

Team files (4): debug.md (frontmatter + body manager-quality → manager-develop), run.md (TRUST5 → sync-auditor ×2), glm.md (PHASE 3 → sync-auditor), review.md (fallback → sync-auditor). team-reader/team-validator deferred to M4.
Agent bodies (6 under .claude/agents/moai/): manager-develop (OUT OF SCOPE + Delegation Protocol → sync-auditor/per-spawn; @MX:ANCHOR comment reworded), builder-harness (OUT OF SCOPE + Delegation Protocol → per-spawn), sync-auditor (manager-quality supplement → orchestrator batch + Stop hook), manager-git (Input from → sync-auditor), manager-spec (line 5 absorption → §C row 1 pointer + Delegation/Step 6 → per-spawn), manager-docs (line 5 absorption → §C row 4 pointer + OUT OF SCOPE/Delegation → per-spawn/sync-auditor).
Mirror: 4 team files + manager-develop/builder-harness/sync-auditor.md fully mirrored (cp, baseline-identical); manager-spec/docs/git.md changed-line mirror only (Edit, baseline-diverged per §E.5/Mx lifecycle owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001).
Full archived-proper residual BOTH trees (skills+agents, excl rejection + dev-only): **0**. researcher KEEP: **14**.

### M4 — broken cross-refs + invalid dispatch

_<pending>_

### M5 — stale ground-truth

_<pending>_

### M6 — run.md LOC trim + verification + template parity

_<pending>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
