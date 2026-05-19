---
id: SPEC-V3R5-CLAUDE-REFRESH-001
title: "MoAI-ADK Architecture Truth Reconciliation + Bundle A Settings Fix"
version: "0.2.0"
status: completed
created: 2026-05-18
updated: 2026-05-19
author: GOOS Kim
priority: P0
phase: "v3.5.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "core, architecture, refresh, mobile-retire, settings, w0"
---

# SPEC-V3R5-CLAUDE-REFRESH-001 — MoAI-ADK Architecture Truth Reconciliation + Bundle A Settings Fix

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-18 | GOOS Kim | Initial draft — W0 entry SPEC for Mega-Sprint v3.5.0. Combines Bundle A (settings.json.tmpl 2 LOC matcher fix) + Bundle B (CLAUDE.md architecture truth reconciliation + expert-mobile retire + AskUserQuestion compression + version bump). |
| 0.2.0 | 2026-05-19 | GOOS Kim | Sync-phase: lifecycle COMPLETE. Run-PR #1006 admin-squash-merged at fc31b30b4. 8/8 ACs binary PASS, NEW_COUNT=0 (delta-only D6 satisfied). W0 of Mega-Sprint v3.5.0 closed; unblocks W1 CONSTITUTION-DUAL-001. |

## 1. Goal

Reconcile MoAI-ADK template-level architecture documentation with runtime reality:

1. **Settings hook coverage** — Extend `settings.json.tmpl` matcher patterns to cover `/clear`, auto-compact, and `MultiEdit` events.
2. **Architecture truth** — Replace the false "6-phase pipeline" claim in CLAUDE.md §5 with the actual runtime behavior (`/moai run` spawns `manager-develop` as sole implementer; `expert-*` agents are utility-only).
3. **Dead code removal** — Retire `expert-mobile` agent (zero documented usage, dangling skill reference).
4. **SSOT integrity** — Correct `ToolSearch` syntax to match `askuser-protocol.md` canonical spec; compress `AskUserQuestion` documentation from 5 locations to ≤2 (SSOT + brief references).
5. **Version reconciliation** — Bump CLAUDE.md from v14.0.0 (43 days stale) to v14.2.0 with explicit W0 changelog entry.

This SPEC is the **W0 entry** of the Mega-Sprint v3.5.0 Harness Autonomy roadmap defined in `.moai/research/harness-autonomy-vision-2026-05-18.md` §5. It unblocks W1 (CONSTITUTION-DUAL-001) by establishing accurate architecture truth as the baseline for the FROZEN/EVOLVABLE zone separation.

## 2. Scope

### In Scope

- `internal/template/templates/.claude/settings.json.tmpl` — 2 matcher pattern fixes (lines 6 + 81)
- `internal/template/templates/CLAUDE.md` — §5 Agent Chain rewrite + §4 Agent Catalog update + §8 ToolSearch syntax + AskUserQuestion compression + footer version bump
- `internal/template/templates/.claude/agents/moai/expert-mobile.md` — hard delete
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — AskUserQuestion paraphrase reduction to 1-line SSOT reference
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` — AskUserQuestion summary trimmed with explicit SSOT citation

### Out of Scope (Deferred to Later Waves)

- W1: Frozen Guard runtime PreToolUse hook implementation
- W1: `zone-registry.md` 100% HARD-clause coverage (57 unmapped entries)
- W2: `expert-backend` / `expert-frontend` retire + domain knowledge migration to `moai-meta-harness` templates
- W2: `moai-domain-{backend,frontend,database}` skill retire
- W2: `manager-develop` redefinition as "universal implementer"
- W3: `harness-learner` autonomous evolution mechanism
- W3: 5-Layer Safety (Frozen Guard / Canary / Contradiction / Rate Limiter / Human Oversight)
- W4: `meta-harness` 7-Phase workflow + cold-start seed library
- W4: `/moai project --refresh` command
- Any change to `expert-{security, devops, performance, refactoring}` (cross-cutting universals stay in Core)
- Any change to `manager-develop.md` Delegation Protocol (W2 territory)
- Any change to actual runtime hook implementation (W1 territory)

## 3. Requirements (EARS)

### Ubiquitous (always-on system constraints)

**REQ-CLR-001** — The `settings.json.tmpl` SessionStart hook matcher SHALL include both `clear` and `compact` patterns in addition to existing `startup` and `resume`, so that the session restoration hook fires after `/clear` and after auto-compact events.

**REQ-CLR-002** — The `settings.json.tmpl` PostToolUse hook matcher SHALL include `MultiEdit` in addition to existing `Write` and `Edit`, so that post-tool processing covers all file-modification tools currently provided by Claude Code.

**REQ-CLR-003** — The CLAUDE.md §5 "Agent Chain for SPEC Execution" section SHALL NOT claim that `/moai run` automatically dispatches a multi-phase pipeline spanning `manager-spec` → `manager-strategy` → `expert-backend` → `expert-frontend` → `manager-quality` → `manager-docs`, and SHALL explicitly document the truth: `/moai run` spawns `manager-develop` as the sole implementer (per `quality.yaml` `development_mode` selecting `ddd` or `tdd` cycle), while `expert-{backend,frontend}` agents are dormant in the auto-workflow but active in utility commands (`/moai fix`, `/moai loop`, `/moai mx`, `/moai review`, `/moai design`, `/moai e2e`).

**REQ-CLR-004** — The file `internal/template/templates/.claude/agents/moai/expert-mobile.md` SHALL NOT exist after this SPEC completes, and a recursive grep for `moai-domain-mobile` across `internal/template/templates/` SHALL return zero matches. CLAUDE.md §4 Agent Catalog SHALL document this retirement.

**REQ-CLR-005** — The CLAUDE.md §8 documentation of `ToolSearch` SHALL use the exact syntax `ToolSearch(query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet")` (with no `max_results` parameter), matching the canonical form in `.claude/rules/moai/core/askuser-protocol.md` § ToolSearch Preload Procedure.

**REQ-CLR-006** — `AskUserQuestion` protocol documentation SHALL exist as a full specification only in `.claude/rules/moai/core/askuser-protocol.md` (SSOT). The four other locations that currently paraphrase the protocol (CLAUDE.md §1 HARD rules, CLAUDE.md §8 User Interaction Architecture, `moai-constitution.md` §MoAI Orchestrator, `agent-common-protocol.md` §User Interaction Boundary) SHALL contain only brief references citing the SSOT, with no duplicated procedural detail beyond a one-line summary plus a hyperlink-style reference.

**REQ-CLR-007** — The CLAUDE.md footer SHALL read `Version: 14.2.0 (Architecture Truth + W0 Bundle A+B)` and `Last Updated: 2026-05-18`, and the document SHALL include a W0 changelog entry summarizing the changes from v14.0.0 to v14.2.0 (settings matcher extensions, agent chain truth, expert-mobile retire, ToolSearch syntax fix, AskUserQuestion compression).

### Verification (binary observable post-condition)

**REQ-CLR-008** (Verification): The `moai agent lint --strict` after run-phase SHALL produce NO NEW findings versus the baseline captured at run-phase start (`/tmp/lint-baseline-w0.json`). Pre-existing 321 findings (237 ERROR + 84 WARN) are out-of-scope per Constraint C3 and dissolve in W2 (CORE-SLIM-001) which retires expert-{backend,frontend,mobile} causing the LR-08 preload drift findings.

## 4. Acceptance Criteria

See `design.md` §2 for the full hierarchical AC tree. AC placement justified in plan.md Decision Log D5.

## 5. Constraints

- **C1** — All file edits MUST stay inside `internal/template/templates/**` (template-first rule per CLAUDE.local.md §2). Live `.claude/` and live `CLAUDE.md` at project root are out of scope; they will be regenerated via `make build` and `moai update`.
- **C2** — No change to `manager-develop.md` Delegation Protocol (deferred to W2).
- **C3** — No change to `expert-backend.md` or `expert-frontend.md` (deferred to W2 retirement).
- **C4** — The retirement of `expert-mobile` is a hard delete with no stub agent. Rationale: zero documented user invocation, zero workflow dispatch — unlike `expert-backend/frontend` which DO have utility-command invocations and will need W2 stub redirects.
- **C5** — The CLAUDE.md version bump from v14.0.0 → v14.2.0 (skipping v14.1.0) reflects the cumulative refresh scope; this is intentional and documented in the changelog entry per REQ-CLR-007.
- **C6** — All commit messages and PR descriptions MUST follow Conventional Commits per CLAUDE.local.md §4.
- **C7** — Branch naming MUST follow `feat/SPEC-V3R5-CLAUDE-REFRESH-001-*` pattern per CLAUDE.local.md §18.2.

## 6. Dependencies

- **None** — This SPEC is the W0 entry point of the v3.5.0 Mega-Sprint; it has no upstream SPEC dependencies.
- **Downstream blockers** — W1 (SPEC-V3R5-CONSTITUTION-DUAL-001) consumes this SPEC's architecture truth output as its baseline. W2 (SPEC-V3R5-CORE-SLIM-001) consumes the `expert-mobile` retirement precedent and the §5 dormancy documentation.

## 7. References

- `.moai/research/harness-autonomy-vision-2026-05-18.md` §5 W0 — SPEC scope source
- `.moai/research/architecture-audit-2026-05-18.md` §3 F-001, F-008, F-101, F-102, F-103, F-104, F-105, F-011 — defect catalog this SPEC resolves
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field frontmatter schema (this SPEC complies)
- `.claude/rules/moai/workflow/spec-workflow.md` — plan-phase workflow contract
- `.claude/rules/moai/core/askuser-protocol.md` — AskUserQuestion SSOT (target consolidation point for REQ-CLR-006)
- `CLAUDE.local.md` §2 (template-first), §18.2 (branch naming), §22 (dev settings intent)
