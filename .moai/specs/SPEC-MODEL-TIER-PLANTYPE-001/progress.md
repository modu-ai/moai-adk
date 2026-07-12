---
id: SPEC-MODEL-TIER-PLANTYPE-001
title: "Progress — plan_type-aware model tier profiles"
version: "0.3.1"
status: draft
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.1.x model-policy redesign"
module: "internal/config + internal/template + internal/cli + internal/web + internal/tmux"
lifecycle: spec-anchored
tier: L
tags: "progress, model-policy, plan-type, glm, effort-overlay"
---

# Progress — SPEC-MODEL-TIER-PLANTYPE-001

## §E.1 Plan-phase Audit-Ready Signal

- Current state (v0.3.1, 2026-07-12): **Tier L** (was M); 5-artifact set — spec.md,
  plan.md, acceptance.md, progress.md (this), research.md (GLM reasoning-control findings).
- **iter-3 (FAIL 0.83) bounded-defect fix applied (v0.3.1):** the REQ-MTP-026 GLM detection
  predicate was corrected — `moai glm` persists `llm.team_mode="glm"` (NOT `llm.mode="glm"`;
  `llm.mode` is dormant with zero non-test writer/reader, verified on the live tree), so the
  predicate is now `TeamMode ∈ {"cg", "glm"}` OR the defensive `Mode == "glm"`. AC-MTP-028
  truth-table row `(mode="", team_mode="glm")` flipped FALSE→TRUE; §D.2 edge case synced;
  false premise corrected on 3 surfaces (spec §A.5, plan §A.10, research §E). No REQ/AC/
  milestone count change.
- spec.md: **29 REQ** (REQ-MTP-001..024 + 026..030; 025 retired), GEARS. plan.md:
  **5 milestones** (M1..M5; the former M5 was descoped at v0.2.0, the new M5 is the GLM
  effort overlay), **7 decision points** (D1..D7, all RESOLVED — 0 residual NEEDS
  CLARIFICATION). acceptance.md: **31 AC groups** (AC-MTP-001..024, 026, 027, 028..032;
  025 retired) incl. a/b sub-criteria, **5 GWT scenarios**, **8 edge cases**.
- Design SSOT: `.moai/reports/model-tier-redesign-20260712.md` (rev2, user-approved).
  GLM reasoning-control research: `research.md` (this SPEC dir, cited z.ai sources).
- **Decisions resolved (kickoff 2026-07-12 + v0.3.0 GLM addition):** D1 replace-both
  precedence; D2 persisting plan_type selector (sanctioned 013 read-only exception,
  1 field); D3 `moai update --plan-type` override flag; D5 36-cell derivation descoped to
  SPEC-MODEL-TIER-ROUTING-PROFILES-001; **D7 GLM effort overlay** — backend-conditional,
  effort-only, 5→3 collapse + coding-max override, injection mechanism as a run-phase
  empirical gate (Branch A passthrough / Branch B explicit write).
- Baselines measured at plan time: plan.md §A.1–§A.10. v0.2.0 set (fable=0; 9/9 agent
  files carry model:+effort:; apply-pass call refs; spec-lint repo-global pre-existing
  errors/warnings in OTHER SPECs; go vet clean on touched trees) PLUS v0.3.0 GLM set:
  `reasoning_effort`=0 and GLM `thinking` toggle=0 in non-test Go; collapse-fn / override-set
  names=0; GLM detection fields (`llm.mode`/`llm.team_mode`) and effort constants
  (`EffortLevel*`) and injection points (`setGLMEnv`/`injectGLMEnvForTeam`/
  `buildTmuxClearVars`) all located.
- Status: `draft`. Reopens plan-audit (iteration 4, final) — iter-3 returned FAIL 0.83 on the
  one bounded predicate defect (D1/D2), now fixed in v0.3.1; all other M5 material (collapse /
  override / scope / D7-gate / wiring) PASSED iter-3.

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
