---
id: SPEC-MODEL-ROUTING-WIRE-001
title: "Wire the Tier×Phase Model Routing Matrix into Spawn Paths and Resolve Model-Policy Contradictions"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: M
related_specs: [SPEC-TOKEN-ROUTING-001]
tags: "model-routing, tier-phase-matrix, moai-route, spawn-wiring, model-policy, haiku-inherit, workflow-reflex"
---

# SPEC-MODEL-ROUTING-WIRE-001 — Wire the Tier×Phase Model Routing Matrix into Spawn Paths

## Epic Context

**Epic**: Workflow-Reflex (6-SPEC epic derived from the 3-lens workflow audit: model-tier routing / Loop Engineering / Harness Engineering). This SPEC is **2 of 6**.

- **Dependency notes**: SPEC 1 (SPEC-HARNESS-RATCHET-REWIRE-001) and this SPEC are independent of each other. SPEC 3 (SPEC-LOOP-VERDICT-CONTRACT-001) is independent. **Downstream SPEC-ADVISOR-RUNG-001 depends on THIS SPEC** — the `moai route` mechanical value surface and the pre-spawn consultation instruction are its prerequisites.
- **Tier**: M (standard) — see plan.md §A.4 for evidence.
- **era**: V3R6 (modern 3-phase close: plan→run→sync).
- **Predecessor debt**: this SPEC discharges the recorded D1 wiring debt of SPEC-TOKEN-ROUTING-001 (its progress.md AC-TR-005: "spawn wiring is prompt-layer (DD2)" — never done).

## Traceability (audit findings provenance)

| Finding ID | Severity | Summary |
|------------|----------|---------|
| R1 | HIGH | `model_routing` matrix + `RouteModelFor` have ZERO call sites outside `internal/config`; workflow.yaml comment claims spawn-time consultation that exists nowhere |
| R2 | HIGH | Default orchestration paths (Mode 4 fan-out, Mode 5 sequential run path) carry zero model/effort guidance — spawns inherit the orchestrator's expensive model; mode-orchestration.md "(inherit)" contradicts workflow.yaml role_profiles |
| R4 | MED | Uncommitted working-tree flip: manager-docs.md / manager-git.md (+ template mirrors) changed `model: haiku` → `inherit`, contradicting model-policy.md which still names them as the haiku exception |
| R5 | LOW | workflow.yaml `team.default_model: opus[1m]` is outside the documented enum and embeds the hazardous `[1m]` suffix |

---

## User Story

**As the** MoAI orchestrator spawning run/sync/mx-phase agents for a Tier-classified SPEC,
**I want** a pre-spawn instruction that consults the Tier×Phase `model_routing` matrix — backed by a mechanical `moai route <tier> <phase>` CLI so the value cannot be hallucinated — and self-consistent model-policy surfaces,
**so that** per-spawn model/effort routing (60-70% cost reduction potential already declared in the matrix) actually happens at spawn time, instead of every spawn silently inheriting the orchestrator's most expensive model while three documentation surfaces contradict each other.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Bash/Read.

### GAP-1 — Matrix and accessor are dead wiring (R1)

- **Measured source**: `.moai/config/sections/workflow.yaml` `model_routing:` block (matrix observed at lines 164-176; comment block at lines 154-163); `internal/config/model_routing.go` `RouteModelFor` (func at line 89; `@MX:ANCHOR` at line 87: "the spawn-time cost-routing accessor").
- **Observed pattern**: `grep -rn "RouteModelFor" --include='*.go'` (excluding tests) matches ONLY `internal/config/model_routing.go` + a comment in `internal/config/types.go` — zero external call sites. No skill/rule file references `model_routing`. The workflow.yaml comment states verbatim: *"the orchestrator consults RouteModelFor(tier, phase) at spawn time and applies the returned {model, effort} as a per-spawn override"* — behavior that exists nowhere. This is SPEC-TOKEN-ROUTING-001's recorded D1 wiring debt.

### GAP-2 — Default orchestration paths carry zero model guidance (R2)

- **Measured source**: `.claude/rules/moai/workflow/orchestration-mode-selection.md` §A Mode 4 row + §C.2 (concurrency ceiling 3-5 only — no model/effort); `.claude/skills/moai/workflows/run.md` + `run/phase-execution.md` (`grep -c model` → 0 in both); `run/mode-orchestration.md` "Team composition: implementer (inherit) + tester (inherit) + reviewer (inherit, read-only)" (observed near line 43).
- **Observed pattern**: Mode 4 guidance specifies concurrency only; Mode 5 sequential path has zero model references; mode-orchestration.md's "(inherit)" blanket contradicts workflow.yaml `role_profiles` (researcher=haiku/low, implementer=sonnet/xhigh, architect=opus/xhigh...) AND the repo's own `workflow_agents` taxonomy (lines 146-153, e.g. `read-only-extract: {model: haiku, effort: low}`).

### GAP-3 — haiku↔inherit contradiction across surfaces (R4)

- **Measured source**: `.claude/agents/moai/manager-docs.md:14` + `.claude/agents/moai/manager-git.md:13` → `model: inherit` (uncommitted working-tree flip, mirrored in `internal/template/templates/.claude/agents/moai/`); `.claude/rules/moai/development/model-policy.md` § Inherit-by-Default Convention → *"except `manager-docs` and `manager-git` which use `model: haiku`"* and *"Exceptions (do NOT migrate to inherit): `model: haiku` agents (`manager-docs`, `manager-git`) — Haiku has no `[1m]` variant, so the bug does NOT apply"*; `.claude/agents/moai/builder-harness.md` speed-critical exception table row (observed near line 165) still teaches `haiku` for mechanical agents.
- **Observed pattern**: Live agent frontmatter says `inherit`; policy prose says `haiku` is the mandated exception. Direction is undecided; surfaces actively contradict.

### GAP-4 — team.default_model out-of-enum with hazardous suffix (R5)

- **Measured source**: `.moai/config/sections/workflow.yaml:24` → `default_model: opus[1m]`.
- **Observed pattern**: The documented enum is `inherit/haiku/sonnet/opus` (settings-management.md team table; workflow_agents closed set). `opus[1m]` is outside the enum and embeds the `[1m]` suffix whose propagation hazard the model-policy doctrine exists to avoid.

### Aggregate defect claim

**The cost-routing matrix is authored but unwired, the spawn paths are guidance-free, and the model-policy surfaces contradict each other in two places.** This SPEC wires the matrix at the prompt layer, adds a mechanical CLI value surface, resolves both contradictions in one decided direction, and fixes the stale comment + out-of-enum config value.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the run workflow", "the CLI", "the mode-selection rule", "the model-policy surfaces"). No legacy `IF/THEN` modality.

### REQ-MRW-001 — Event-driven (When) — pre-spawn consultation instruction

**When** the orchestrator prepares a run/sync/mx-phase `Agent()` spawn for a Tier-classified SPEC, the run workflow (`.claude/skills/moai/workflows/run.md` Phase 0.95 / pre-spawn step) and `orchestration-mode-selection.md` §B.1 SHALL instruct the orchestrator to consult `model_routing[<TIER>-<phase>]` and pass the resolved `{model, effort}` as per-spawn runtime arguments (per-spawn args are `[1m]`-safe per model-policy.md — *"the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug"*), yielding to any explicit user override.

### REQ-MRW-002 — Ubiquitous — mechanical value surface (`moai route`)

The CLI SHALL provide a `moai route <tier> <phase>` subcommand that calls `RouteModelFor(tier, phase)` and prints the resolved entry (model, effort, fallback_applied) in plain text and, with `--json`, as structured JSON — giving the orchestrator a non-hallucinable value source and closing the zero-call-site defect.

### REQ-MRW-003 — Event-detected (When) — invalid-argument rejection

**When** `moai route` receives a tier outside `{S, M, L}` or a phase outside `{plan, run, sync, mx}`, the CLI SHALL exit non-zero with a stderr usage error naming the valid enum sets (subagent-boundary discipline: structured stderr, no interactive prompt).

### REQ-MRW-004 — Ubiquitous — Mode 4 worker-model guidance

The mode-selection rule (`orchestration-mode-selection.md` §A row 4 + §C.2) SHALL carry Mode 4 worker-model guidance: read-only research spawns SHOULD carry `model: sonnet` (`haiku` for pure extraction) plus purpose-appropriate effort per the `workflow_agents` taxonomy (workflow.yaml `workflow_agents:` block), so parallel fan-out stops silently inheriting the orchestrator's model.

### REQ-MRW-005 — Ubiquitous — haiku↔inherit contradiction resolved in ONE direction

The model-policy surfaces SHALL agree in a single decided direction on the manager-docs/manager-git model field. The surfaces to reconcile: `manager-docs.md` + `manager-git.md` frontmatter (live + template mirrors), `model-policy.md` § Inherit-by-Default exception prose, `builder-harness.md` speed-critical exception table. Direction is a plan-phase decision point (plan.md §D D1) with recommended default **revert to `model: haiku`** — haiku has no `[1m]` variant (immune to the entitlement-inheritance bug) and the matrix's S-sync/S-mx/M-mx rows expect haiku-class cheap routing. Both directions are documented in plan.md §D; the orchestrator confirms at Implementation Kickoff Approval.

### REQ-MRW-006 — Ubiquitous — team.default_model to documented enum

The workflow config SHALL set `team.default_model: inherit` (a documented enum value), removing the out-of-enum `opus[1m]` value and its embedded `[1m]` suffix hazard (live + template mirror).

### REQ-MRW-007 — Ubiquitous — truthful model_routing comment

The workflow.yaml `model_routing` comment block SHALL describe the actually-implemented consultation mechanism after this SPEC (prompt-layer pre-spawn instruction + `moai route` CLI value surface), replacing the aspirational "the orchestrator consults RouteModelFor(tier, phase) at spawn time" claim with a description that is true.

### REQ-MRW-008 — Ubiquitous — mode-orchestration role-model reconciliation

The team-mode orchestration doc (`run/mode-orchestration.md` team-composition line) SHALL be reconciled with the workflow.yaml `role_profiles` SSOT — either citing role_profiles as the model source or stating the per-role values — removing the blanket "(inherit)" contradiction.

### REQ-MRW-009 — Capability gate (Where) — template-first boundary

**Where** an edited surface has a template mirror under `internal/template/templates/` (verified present for: workflow.yaml, run.md, mode-orchestration.md, orchestration-mode-selection.md, model-policy.md, manager-docs.md, manager-git.md), the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees. New Go code (`moai route`) is NOT templated.

---

## Constraints

1. **`[1m]` workaround untouched (HARD)** — the inherit-by-default frontmatter convention and the per-spawn-runtime-arg safety rationale in model-policy.md are PRESERVED; this SPEC wires routing THROUGH the safe per-spawn channel, never via frontmatter model pins. Changing the `[1m]` bug workaround itself is out of scope.
2. **Matrix yields to explicit override (HARD)** — the pre-spawn instruction preserves workflow.yaml's declared precedence: an explicit user/caller override beats the matrix value. REQ-MRW-001 binds.
3. **Subagent boundary** — `moai route` MUST NOT prompt (positional args + flags + structured stderr only); a `TestRoute_NoAskUserQuestion`-style static guard test is required per internal/cli conventions.
4. **Naming**: `moai harness route` already exists (`internal/cli/harness_route.go`, SPEC-to-harness-level routing). The new top-level `moai route` has a different cobra parent (no registration collision), but the run-phase MUST keep help texts explicitly distinct; plan.md §D D2 records the naming decision.
5. **GEARS notation; era V3R6; 12 canonical frontmatter fields.**

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — [1m] bug workaround redesign

- Changing the inherit-by-default frontmatter convention, the Anthropic #45847/#51060/#36670 workaround mechanics, or the baseline-refill-breaker doctrine. This SPEC routes through the existing per-spawn-arg safe channel; it does not redesign the channel.

### Out of Scope — team.enabled default

- Flipping `workflow.team.enabled` (stays `false` per the Sonnet 5 / Opus 4.8 re-design). Only `team.default_model` is corrected.

### Out of Scope — CG mode

- `moai cg` / GLM teammate routing, GLM model tables, and `llm.yaml`. The `glm` value in the model_routing closed set is untouched.

### Out of Scope — RouteModelFor internal logic

- The Go accessor's validation, fallback semantics, and the matrix VALUES themselves (S/M/L × plan/run/sync/mx entries). Only call sites, documentation wiring, and the comment are in scope.

### Out of Scope — role_profiles redesign

- Changing workflow.yaml `role_profiles` values (researcher=haiku etc.). REQ-MRW-008 reconciles the DOC to the SSOT; it does not change the SSOT.

### Out of Scope — mechanical spawn-time enforcement

- A hook or Go runtime layer that mechanically injects the routed model into `Agent()` calls. This SPEC's wiring is prompt-layer (per the SPEC-TOKEN-ROUTING-001 DD2 decision) + a CLI value surface; runtime enforcement is a potential follow-up (candidate territory for downstream SPEC-ADVISOR-RUNG-001).

---

## Cross-References

- **EXTEND base (Go)**: `internal/config/model_routing.go` (`RouteModelFor`, line 89; `@MX:ANCHOR` line 87); new `internal/cli` subcommand file for `moai route`.
- **EXTEND base (doc)**: `.claude/skills/moai/workflows/run.md` Phase 0.95 area (§ Phase 0.95 Operational Entries observed near lines 39/88-92); `orchestration-mode-selection.md` §A/§B.1/§C.2; `run/mode-orchestration.md` team-composition line.
- **Reconcile surfaces (R4)**: `.claude/agents/moai/manager-docs.md:14`, `manager-git.md:13`, template mirrors, `model-policy.md` § Inherit-by-Default Convention, `builder-harness.md` model table.
- **Config**: `.moai/config/sections/workflow.yaml` — `model_routing` (154-176), `workflow_agents` (146-153), `team.default_model` (24) + template mirror.
- **Name-adjacent existing command**: `internal/cli/harness_route.go` (`moai harness route` — harness-level routing; distinct concern).
- **Predecessor**: SPEC-TOKEN-ROUTING-001 (matrix + accessor author; D1 wiring debt discharged here).
- **Downstream dependent**: SPEC-ADVISOR-RUNG-001 (unauthored; depends on this SPEC).
- **Epic**: Workflow-Reflex 2 of 6. Siblings: SPEC-HARNESS-RATCHET-REWIRE-001 (1 of 6), SPEC-LOOP-VERDICT-CONTRACT-001 (3 of 6).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Workflow-Reflex Epic 2 of 6. Prompt-layer spawn wiring + `moai route` CLI + haiku↔inherit reconciliation + config fixes. Tier M. |
