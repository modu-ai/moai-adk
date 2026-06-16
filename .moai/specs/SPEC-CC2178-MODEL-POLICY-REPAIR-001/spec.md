---
id: SPEC-CC2178-MODEL-POLICY-REPAIR-001
title: "CC 2.1.178 Model-Policy Repair: 3-Axis Cost Routing Alignment + Phantom-Map Cleanup"
version: "0.1.0"
status: draft
created: 2026-06-16
updated: 2026-06-16
author: manager-spec
priority: P1
phase: "v3.0.0 (CC 2.1.178 alignment)"
module: "internal/template"
lifecycle: spec-anchored
era: V3R6
tags: "model-policy, cost-routing, cc2178, cycle-type, harness"
---

# SPEC-CC2178-MODEL-POLICY-REPAIR-001

## §A. Problem Statement

A prior research session (`.moai/research/cc-update-2.1.163-to-2.1.178.md`, 19871 bytes, 2026-06-16) analyzed Claude Code 2.1.163→2.1.178 changes and identified the **linchpin cost lever**: CC 2.1.175 introduced `availableModels` / `enforceAvailableModels` settings fields that constrain Default-model resolution at the settings level — the ONLY cost lever that does NOT touch the `#45847 [1m]` spawn-bug (per-agent model pinning is forbidden because it breaks `[1m]` entitlement inheritance; see `.claude/rules/moai/development/model-policy.md` § Inherit-by-Default).

The research doc's conclusion: route cost via the **Default model** (set to Sonnet) rather than per-agent pins, and align the full 3-axis cost model (model × effort × cycle_type).

Independent ground-truth verification against the current tree (2026-06-16) surfaced **4 defects** that this SPEC repairs:

- **D1 — `agentModelMap` severely stale** (`internal/template/model_policy.go:193-216`): 18 entries, only 3 match the current 8-agent retained catalog. 15 entries reference archived phantom agents (`manager-quality`, `manager-project`, `manager-strategy`, 7 `expert-*`, `builder-skill`, `builder-plugin`) that MUST NOT be spawned per SPEC-V3R6-AGENT-TEAM-REBUILD-001 (17→8 consolidation, 2026-05-25). The 2 current retained agents `manager-develop` and `builder-harness` are MISSING (exist only as legacy `manager-ddd`/`manager-tdd`/`builder-agent` names). `GetAgentModel` returns `""` for unknown agents (`model_policy.go:221-222`), so missing entries are silent no-ops — but the phantom entries are dead weight documenting a model-policy intent that no longer matches reality.
- **D2 — `agentEffortMap` stale** (`internal/template/model_policy.go:72-79`): 6 entries, 3 are archived phantoms (`manager-strategy`, `expert-security`, `expert-refactoring`). Missing current retained `manager-develop`, `builder-harness`.
- **D3 — CC 2.1.175 cost lever absent** (verified 0 occurrences): `availableModels` / `enforceAvailableModels` appear NOWHERE in `internal/template/templates/.claude/settings.json.tmpl` or `internal/config/`. The Default-model cost lever is completely unwired.
- **D4 — cycle_type globally fixed, not harness/Tier-linked** (`.moai/config/sections/quality.yaml`): `development_mode: tdd` + `enforce_quality: true` are flat globals. cycle_type is NOT routed by harness depth (minimal/standard/thorough) or SPEC Tier, despite the research doc's 3-axis model requiring it. The harness `skip_phases` array (`harness.yaml:37-45`) already encodes phase-skipping per level but does not feed `development_mode`.

**Constraint to re-verify (research task, not assumption)**: `model-policy.md` § Inherit-by-Default (L30-50) currently treats `#45847` as an active constraint with NO mention of CC≥2.1.174 normalization relaxation. The research doc's `last_analyzed=2.1.163`; this SPEC advances to 2.1.178, so the `[1m]` constraint MUST be re-checked against the newer runtime before finalizing whether per-agent pinning can be partially re-enabled.

## §B. Scope

### §B.1 In Scope

1. **Model axis (primary lever)** — wire `availableModels`/`enforceAvailableModels` (CC 2.1.175) in `settings.json.tmpl` to set **Default = Sonnet** (the `[1m]`-safe lever). Template-First change.
2. **Cycle_type axis (gap)** — route `quality.yaml development_mode` by harness/Tier instead of a flat global `tdd`. Backward-compatible (existing `tdd` projects must not break).
3. **Effort axis (already wired)** — tune the expensive-bias default only. Do NOT re-architect.
4. **Phantom-map cleanup** — remove 15 archived entries from `agentModelMap` (D1) and 3 from `agentEffortMap` (D2); add missing `manager-develop` + `builder-harness`.
5. **`agentEffortMap` redundancy decision** — determine whether the map is still needed (modern agents declare `effort:` in frontmatter → `ApplyEffortPolicy` may be redundant). If redundant, retire cleanly; if kept, prune to the retained set.
6. **`[1m]` re-verification** — research task to re-check `#45847` against CC 2.1.178 before finalizing the per-agent pinning policy.
7. **Task-triage decision** — decide whether the per-task triage signal (failure-cost × visual-verifiability) is in-scope or deferred; if in-scope, define the signal concretely.

### §B.2 Out of Scope (Exclusions — What NOT to Build)

- **EX-01**: Full per-agent model pinning matrix (forbidden by `[1m]` constraint regardless of re-verification outcome — Default-model routing is the only safe lever even if `#45847` is relaxed at 2.1.178, because the relaxation may be partial or Sonnet-specific).
- **EX-02**: vff (verbose-few-shot factuality) prose-discipline integration (separate follow-up SPEC per research doc; this SPEC is cost-routing-only).
- **EX-03**: docs-site 4-locale documentation of the new cost model (sync-phase work, not plan-phase; belongs to `/moai sync`).
- **EX-04**: Telemetry instrumentation to measure actual cost savings (deferred — the SPEC wires the levers; measurement is a follow-up observability SPEC).
- **EX-05**: Re-architecting the effort axis (`ApplyEffortPolicy` mechanism itself is out of scope; only the map contents and the redundancy decision are in scope).
- **EX-06**: Go-side `cmd/moai` CLI flag exposure for runtime model override (settings-level wiring only; no new CLI flags).

### §B.3 Anti-Goals

- **AG-01**: Do NOT break existing `tdd` projects when routing `development_mode` by harness — backward compatibility is MUST-PASS.
- **AG-02**: Do NOT introduce per-agent model pins — the `[1m]` constraint is the hard boundary even post-re-verification.
- **AG-03**: Do NOT absorb the 10 uncommitted working-tree files (internal/config/defaults.go, internal/template/deployer.go, etc.) — they belong to an unrelated parallel workstream. Scope discipline: touch ONLY the new SPEC directory.

## §C. Requirements (GEARS Format)

### §C.1 Model Axis — Default-Model Cost Lever (D3)

> **REQ-MPR-001** (Ubiquitous, MUST)
> The `settings.json.tmpl` template shall include an `availableModels` field listing the permitted model aliases (`sonnet`, `opus`, `haiku`) so that the Default-model resolution is constrained at the settings level per CC 2.1.175.

> **REQ-MPR-002** (Capability gate, MUST)
> Where the `availableModels` field is present, the `settings.json.tmpl` template shall include an `enforceAvailableModels: true` field so that the Default model is the authoritative cost lever and per-agent `model:` pins that escape the allowlist are rejected by the runtime.

> **REQ-MPR-003** (State-driven, MUST)
> While a `[1m]` parent session is active, the deployed `settings.json` shall set the Default model to `sonnet` (via the `model:` field or equivalent Default-model mechanism) so that cost routing flows through the `[1m]`-safe lever without per-agent pins.

### §C.2 Cycle_Type Axis — Harness/Tier Routing (D4)

> **REQ-MPR-004** (Capability gate, MUST)
> Where the harness Complexity Estimator resolves to `minimal`, the harness routing shall select a lightweight cycle_type (skip_phases-driven DDD-lite or equivalent) so that simple changes do not pay the full-TDD overhead.

> **REQ-MPR-005** (Capability gate, MUST)
> Where the harness Complexity Estimator resolves to `thorough`, the harness routing shall select full TDD (`development_mode: tdd`) so that critical features retain the RED-GREEN-REFACTOR discipline.

> **REQ-MPR-006** (State-driven, MUST)
> While a project's `quality.yaml` explicitly pins `development_mode: tdd`, the harness routing shall preserve that explicit pin and NOT override it with a harness-derived cycle_type so that existing TDD projects are backward-compatible (AG-01).

> **REQ-MPR-007** (Event-driven, MUST)
> When the harness level is `standard`, the harness routing shall default cycle_type to `tdd` (the current global behavior) so that the default case is unchanged.

### §C.3 Phantom-Map Cleanup (D1, D2)

> **REQ-MPR-008** (Unwanted, MUST)
> The `agentModelMap` (`internal/template/model_policy.go:193`) shall not contain any entry whose key is an archived phantom agent (`manager-quality`, `manager-project`, `manager-strategy`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-debug`, `expert-testing`, `expert-refactoring`, `builder-skill`, `builder-plugin`) so that the map documents only the retained 8-agent catalog.

> **REQ-MPR-009** (Event-driven, MUST)
> When `GetAgentModel` is called with `manager-develop` or `builder-harness`, the function shall return a non-empty model tuple so that the 2 missing retained agents are covered by the model policy.

> **REQ-MPR-010** (Unwanted, MUST)
> The `agentEffortMap` (`internal/template/model_policy.go:72`) shall not contain any entry whose key is an archived phantom agent (`manager-strategy`, `expert-security`, `expert-refactoring`) so that the effort map documents only the retained catalog.

### §C.4 Effort Map Redundancy Decision

> **REQ-MPR-011** (Ubiquitous, MUST)
> The SPEC shall, in its plan-phase research, determine whether `agentEffortMap` + `ApplyEffortPolicy` is redundant given that modern agents declare `effort:` directly in frontmatter, and record the decision (retire vs. prune-and-keep) with the test consequences documented.

> **REQ-MPR-012** (Capability gate, SHOULD)
> Where the research concludes `agentEffortMap` is redundant, the SPEC shall retire it cleanly (remove the map, `GetAgentEffort`, `ApplyEffortPolicy`, and their tests) rather than leaving dead code, provided no non-test caller depends on the public API.

### §C.5 [1m] Constraint Re-Verification

> **REQ-MPR-013** (Ubiquitous, MUST)
> The SPEC shall include a research task that re-verifies the `#45847` `[1m]` entitlement-inheritance constraint against the CC 2.1.178 runtime (fetching the actual upstream issue state and CHANGELOG at 2.1.178) before the plan-phase is marked audit-ready, and record the verdict (still-active / relaxed / partially-relaxed) in the plan-phase research notes.

> **REQ-MPR-014** (State-driven, MUST)
> While the `[1m]` re-verification verdict is "still-active" (constraint unchanged at 2.1.178), the SPEC shall scope per-agent pinning as permanently out-of-scope (EX-01) and rely exclusively on the Default-model lever (REQ-MPR-001..003).

> **REQ-MPR-015** (Event-driven, SHOULD)
> When the `[1m]` re-verification verdict is "relaxed" or "partially-relaxed", the SPEC shall record the relaxation scope as a follow-up SPEC candidate (not an in-scope expansion of this SPEC) so that this SPEC remains deliverable without scope creep.

### §C.6 Task-Triage Signal Decision

> **REQ-MPR-016** (Ubiquitous, MUST)
> The SPEC shall, in its plan-phase, decide whether the per-task triage signal (failure-cost × visual-verifiability, per research doc) is in-scope for THIS SPEC or deferred to a follow-up, and record the decision with rationale.

> **REQ-MPR-017** (Capability gate, SHOULD)
> Where the task-triage decision is "in-scope", the SPEC shall define the triage signal concretely (the two input dimensions, the mapping function to {opus/full-tdd, sonnet/ddd-lite}, and the harness integration point) as additional REQs.

### §C.7 Effort Axis Tuning (Light)

> **REQ-MPR-018** (Ubiquitous, SHOULD)
> The SPEC shall tune the effort-axis expensive-bias default only at the documentation level (a note in `model-policy.md` that effort is already harness-linked and this SPEC does not alter the mechanism), and shall not re-architect `ApplyEffortPolicy` (EX-05).

## §D. Acceptance Criteria (Summary — see acceptance.md for full Given-When-Then)

- **AC-MPR-001** (MUST): `settings.json.tmpl` contains `availableModels` with the 3 aliases.
- **AC-MPR-002** (MUST): `settings.json.tmpl` contains `enforceAvailableModels: true`.
- **AC-MPR-003** (MUST): Deployed settings Default model is `sonnet`.
- **AC-MPR-004** (MUST): Harness `minimal` resolves to a lightweight cycle_type (not full-TDD).
- **AC-MPR-005** (MUST): Harness `thorough` resolves to full TDD.
- **AC-MPR-006** (MUST): Explicit `quality.yaml development_mode: tdd` pin is preserved (backward-compat).
- **AC-MPR-007** (MUST): `agentModelMap` has 0 archived-phantom keys (grep-count = 0).
- **AC-MPR-008** (MUST): `GetAgentModel("manager-develop")` and `GetAgentModel("builder-harness")` return non-empty.
- **AC-MPR-009** (MUST): `agentEffortMap` has 0 archived-phantom keys.
- **AC-MPR-010** (MUST): Effort-map redundancy decision is recorded (retire vs. prune-and-keep) with test-consequence documentation.
- **AC-MPR-011** (MUST): `[1m]` re-verification verdict is recorded (still-active / relaxed / partially-relaxed) with upstream evidence.
- **AC-MPR-012** (SHOULD): Task-triage in-scope/deferred decision is recorded with rationale; if in-scope, the triage signal is defined concretely.
- **AC-MPR-013** (SHOULD): `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001` exits 0.

## §E. Constraints

1. **Template-First Rule** (CLAUDE.local.md §2 [HARD]): all template changes MUST land in `internal/template/templates/` FIRST, then `make build` regenerates `embedded.go`. Direct edits to `embedded.go` are forbidden.
2. **`[1m]` entitlement boundary** (model-policy.md § Inherit-by-Default): per-agent `model:` pins are forbidden regardless of re-verification outcome (EX-01); Default-model routing is the only safe lever.
3. **Backward compatibility** (AG-01): existing `tdd` projects must not break when cycle_type routing is introduced.
4. **8-agent catalog alignment** (CLAUDE.md §4): the retained catalog is exactly `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`, `Explore`. All maps must reconcile against this set.
5. **Scope discipline** (AG-03): the 10 uncommitted working-tree files are out-of-scope; only the new SPEC directory is touched in plan-phase.
6. **Language policy** (`.moai/config/sections/language.yaml`): `documentation: ko` — SPEC prose MAY be Korean; code identifiers / REQ tokens / Go symbols MUST stay English. This SPEC uses English REQ tokens with Korean-permitted prose; the author chose English prose for cross-locale readability of the REQ matrix.
7. **Mirror parity** (internal/template/CLAUDE.md): SSOT docs that exist in both `.claude/rules/` and `templates/.claude/rules/` must be edited in the same commit. `model-policy.md` is such a mirror.

## §F. Dependencies / Related SPECs

- **Depends on**: SPEC-V3R6-AGENT-TEAM-REBUILD-001 (established the 8-agent retained catalog that this SPEC reconciles the maps against; archived the 12 phantom agents).
- **Related**: `.moai/research/cc-update-2.1.163-to-2.1.178.md` (research doc, last_analyzed=2.1.163; this SPEC advances to 2.1.178).
- **Follow-up candidates** (deferred, recorded for future SPECs):
  - `SPEC-CC2178-VFF-PROSE-DISCIPLINE-001` (vff integration, EX-02).
  - `SPEC-CC2178-COST-TELEMETRY-001` (cost-savings measurement, EX-04).
  - `SPEC-CC2178-PER-AGENT-PIN-RELAXATION-001` (conditional, only if REQ-MPR-015 fires).

## §G. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `[1m]` re-verification reveals the constraint is partially relaxed, tempting scope expansion into per-agent pins | Medium | High (scope creep) | REQ-MPR-015 routes any relaxation to a follow-up SPEC; this SPEC stays Default-model-only. |
| cycle_type routing breaks existing `tdd` projects | Low | Critical (regression) | REQ-MPR-006 + AC-MPR-006 enforce explicit-pin preservation; backward-compat test is MUST-PASS. |
| `agentEffortMap` retirement removes a public API that a non-test caller depends on | Low | Medium | REQ-MPR-012 requires caller-grep evidence before retirement; if callers exist, prune-and-keep instead. |
| Template-neutrality CI guard rejects `settings.json.tmpl` change if it references internal SPEC IDs | Low | Medium | The template change references only CC model aliases (`sonnet`/`opus`/`haiku`) — no internal IDs. Pre-PR checklist (CLAUDE.local.md §2.1) applies. |
| `availableModels`/`enforceAvailableModels` field semantics differ from research doc's reading at 2.1.178 | Medium | High (wrong implementation) | REQ-MPR-013 re-verifies `#45847` AND the settings-field semantics at 2.1.178 before run-phase. |

## §H. History

- 2026-06-16: plan-phase artifacts authored by manager-spec (spec.md, plan.md, acceptance.md). Initial `status: draft`. Based on `.moai/research/cc-update-2.1.163-to-2.1.178.md` + independent ground-truth verification of D1-D4 against the current tree.
- 2026-06-16: SPEC ID canonicalized from orchestrator-proposed `SPEC-CC2178-MODEL-POLICY-REPAIR` (no numeric suffix) to `SPEC-CC2178-MODEL-POLICY-REPAIR-001` (appended `-001`) to satisfy the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`. See plan.md §B.7 for the decomposition trace.
