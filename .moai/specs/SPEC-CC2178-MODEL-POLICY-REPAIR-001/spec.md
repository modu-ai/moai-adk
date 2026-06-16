---
id: SPEC-CC2178-MODEL-POLICY-REPAIR-001
title: "CC 2.1.178 Model-Policy Repair: 3-Axis Cost Routing Alignment + Phantom-Map Cleanup"
version: "0.3.0"
status: implemented
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
- **D4 — cycle_type globally fixed, not harness/Tier-linked** (`.moai/config/sections/quality.yaml`): the globals `development_mode: tdd` and `enforce_quality: true` are nested under the top-level `constitution:` key in `quality.yaml` (verified 2026-06-16 — NOT flat top-level globals as an earlier draft of this SPEC asserted; the nesting matters because M3 must read the struct field `cfg.Quality.DevelopmentMode` via the existing reader at `internal/config/manager.go:394` which reads `MOAI_DEVELOPMENT_MODE` env, and via the `constitution` section in the parsed config). cycle_type is NOT routed by harness depth (minimal/standard/thorough) or SPEC Tier, despite the research doc's 3-axis model requiring it. The harness Complexity Estimator (`internal/harness/router/router.go:104-108`) resolves a `Level` from SPEC frontmatter `harness_level` but does NOT emit a cycle_type — there is currently no symbol anywhere in `internal/` that maps harness level → cycle_type (verified: `grep -rn "resolveCycleType\|ResolveCycleType" internal/` returns 0 matches). The harness `skip_phases` array (`harness.yaml:37-45`) encodes phase-skipping per level but does not feed `development_mode`.

**Constraint to re-verify (research task, not assumption)**: `model-policy.md` § Inherit-by-Default (L30-50) currently treats `#45847` as an active constraint with NO mention of CC≥2.1.174 normalization relaxation. The research doc's `last_analyzed=2.1.163`; this SPEC advances to 2.1.178, so the `[1m]` constraint MUST be re-checked against the newer runtime before finalizing whether per-agent pinning can be partially re-enabled.

## §B. Scope

### §B.1 In Scope

1. **Model axis (primary lever)** — wire `availableModels`/`enforceAvailableModels` (CC 2.1.175) in `settings.json.tmpl` to set **Default = Sonnet** (the `[1m]`-safe lever). Template-First change.
2. **Cycle_type axis (gap)** — route `quality.yaml development_mode` by harness/Tier instead of a flat global `tdd`. Backward-compatible (existing `tdd` projects must not break).
3. **Effort axis (already wired)** — tune the expensive-bias default only. Do NOT re-architect.
4. **Phantom-map cleanup** — remove 15 archived entries from `agentModelMap` (D1) and 3 from `agentEffortMap` (D2); add missing `manager-develop` + `builder-harness`.
5. **`agentEffortMap` phantom pruning + deferral-record** (D5 iter-2 split) — prune the 3 archived phantom keys, sync map values to match agent files (plan-auditor/sync-auditor → `xhigh`), add missing retained agents (manager-develop/builder-harness). Full retirement of the map + `ApplyEffortPolicy` is UNSAFE (2 production callers at `initializer.go:181` + `update.go:2661`) and is DEFERRED to a follow-up SPEC; this SPEC records the deferral rationale only.
6. **`[1m]` re-verification** — research task to re-check `#45847` against CC 2.1.178 before finalizing the per-agent pinning policy.
7. **Task-triage decision** — decide whether the per-task triage signal (failure-cost × visual-verifiability) is in-scope or deferred; if in-scope, define the signal concretely.

### §B.2 Out of Scope (Exclusions — What NOT to Build)

- **EX-01**: Full per-agent model pinning matrix (forbidden by `[1m]` constraint regardless of re-verification outcome — Default-model routing is the only safe lever even if `#45847` is relaxed at 2.1.178, because the relaxation may be partial or Sonnet-specific).
- **EX-02**: vff (verbose-few-shot factuality) prose-discipline integration (separate follow-up SPEC per research doc; this SPEC is cost-routing-only).
- **EX-03**: docs-site 4-locale documentation of the new cost model (sync-phase work, not plan-phase; belongs to `/moai sync`).
- **EX-04**: Telemetry instrumentation to measure actual cost savings (deferred — the SPEC wires the levers; measurement is a follow-up observability SPEC).
- **EX-05** (iter-3 reworded — D11 remediation): **fixing the dead domain-walker to target `.claude/agents/moai/` IS in-scope (REQ-MPR-019, MUST)** — this is a surgical 1-line fix to `domains := []string{"moai"}` at both `model_policy.go:137` (ApplyEffortPolicy) and `model_policy.go:246` (ApplyModelPolicy), restoring live injection. What remains out-of-scope is the **full retirement** of `agentEffortMap` + `ApplyEffortPolicy` — migrating the 2 production callers (`initializer.go:181`, `update.go:2661`) to an alternative injection path. That full retirement stays deferred to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001`. Walker-target fix = in-scope; full retirement = out-of-scope.
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

> **Canonical phantom-key enumeration (single source of truth — referenced by REQ-MPR-008, AC-MPR-007, acceptance §D.3, and plan.md §B.1)**: `agentModelMap` (`internal/template/model_policy.go:193-216`) currently has **19 entries** (verified by `awk` extraction on 2026-06-16; both the iter-1 SPEC and the plan-auditor iter-1 report undercounted as 18 — this iter-2 correction is the authoritative count). Exactly 3 are retained-correct (`manager-spec`, `manager-docs`, `manager-git`). The remaining **16 keys MUST ALL be removed** — they comprise 13 archived phantom agents, 2 legacy rename-aliases of `manager-develop` (`manager-ddd`, `manager-tdd`), and 1 legacy rename-alias of `builder-harness` (`builder-agent`) plus its 2 archived builder siblings (`builder-skill`, `builder-plugin`). The legacy aliases MUST be removed (not retained as aliases) because the retained canonical names `manager-develop` and `builder-harness` are added back separately by REQ-MPR-009; keeping the aliases alongside the canonical names would duplicate the entries.
>
> The 16 keys to remove (verbatim, cross-checked against `model_policy.go:193-216` on 2026-06-16):
>
> | # | Key | Category | Reason |
> |---|-----|----------|--------|
> | 1 | `manager-ddd` | legacy alias | pre-rename name of `manager-develop`; superseded by REQ-MPR-009 |
> | 2 | `manager-tdd` | legacy alias | pre-rename name of `manager-develop`; superseded by REQ-MPR-009 |
> | 3 | `manager-quality` | archived phantom | archived by SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
> | 4 | `manager-project` | archived phantom | archived by SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
> | 5 | `manager-strategy` | archived phantom | archived by SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
> | 6 | `expert-backend` | archived phantom | archived (8 `expert-*` agents consolidated) |
> | 7 | `expert-frontend` | archived phantom | archived |
> | 8 | `expert-security` | archived phantom | archived |
> | 9 | `expert-devops` | archived phantom | archived |
> | 10 | `expert-performance` | archived phantom | archived |
> | 11 | `expert-debug` | archived phantom | archived |
> | 12 | `expert-testing` | archived phantom | archived |
> | 13 | `expert-refactoring` | archived phantom | archived |
> | 14 | `builder-agent` | legacy alias | pre-rename name of `builder-harness`; superseded by REQ-MPR-009 |
> | 15 | `builder-skill` | archived phantom | archived builder variant |
> | 16 | `builder-plugin` | archived phantom | archived builder variant |
>
> Post-cleanup target: 5 entries — `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `builder-harness`.

> **REQ-MPR-008** (Unwanted, MUST)
> The `agentModelMap` (`internal/template/model_policy.go:193`) shall not contain ANY of the 16 canonical phantom keys enumerated in the table above (13 archived phantoms + 2 `manager-develop` legacy aliases + 1 `builder-harness` legacy alias + 2 archived builder siblings) so that the map documents only the retained 8-agent catalog — ALL 16 keys must be removed, not a subset.

> **REQ-MPR-009** (Event-driven, MUST)
> When `GetAgentModel` is called with `manager-develop` or `builder-harness`, the function shall return a non-empty model tuple so that the 2 missing retained agents are covered by the model policy. The model tuples are chosen to be consistent with this SPEC's central thesis (route cost to Default=Sonnet): **`manager-develop` = `{sonnet, sonnet, haiku}`** and **`builder-harness` = `{sonnet, sonnet, haiku}`** — i.e. the most-frequently-spawned run-phase agents follow Default-Sonnet at the High policy, NOT Opus (see plan.md §F.2 for the design rationale and the rejected Option-C alternative of `{opus, sonnet, haiku}` which would contradict the cost-routing thesis). The iter-1 SPEC derived these from retired legacy aliases (`manager-ddd`/`manager-tdd` = `{opus, sonnet, sonnet}`) — that derivation is REJECTED in iter-2 because it would pin the busiest agent to Opus, undermining the 3-axis cost-routing goal (D6 remediation).

> **REQ-MPR-010** (Unwanted, MUST)
> The `agentEffortMap` (`internal/template/model_policy.go:72`) shall not contain any entry whose key is an archived phantom agent (`manager-strategy`, `expert-security`, `expert-refactoring` — the same 3 phantom keys as REQ-MPR-011a) so that the effort map documents only the retained catalog. (Note: this REQ is the pruning mandate; REQ-MPR-011a adds the broader reconciliation including value-sync and missing-agent addition.)

### §C.4 Effort-Map Phantom Pruning (SAFE, in-scope) vs. Full Retirement (UNSAFE, deferred)

> **Factual correction to iter-1 premise (D5 remediation) + iter-3 dead-walker correction (D11 remediation)**: the iter-1 SPEC premised REQ-MPR-011/012 on "modern agents declare `effort:` directly → `agentEffortMap` + `ApplyEffortPolicy` is redundant → retire it". The iter-2 SPEC corrected part of this: `ApplyEffortPolicy` (`model_policy.go:134-180`) HAS **2 production callers** (`internal/core/project/initializer.go:181`, `internal/cli/update.go:2661`) whose job is to INJECT `effort:` into freshly-deployed agent files that lack it. BUT the iter-2 rationale — "retiring the map means new `moai init` deployments get NO effort injection = behavior regression" — rested on a further premise that the injection actually fires today. **iter-3 ground-truth verification (2026-06-16) shows that further premise is FACTUALLY FALSE in the current tree**: both `ApplyEffortPolicy` (L137) and `ApplyModelPolicy` (L246) hardcode `domains := []string{"core", "expert", "meta", "harness"}` — subdirectories created by SPEC-V3R6-AGENT-FOLDER-SPLIT-001 but since **re-consolidated back to `.claude/agents/moai/`** (verified: `find .claude/agents -type d` returns only `moai/` and `local/`; the four domains do not exist). The walker therefore calls `os.ReadDir(agentsDir)` → `os.IsNotExist(err)` → `continue` for ALL four domains and **injects NOTHING today**. It is a complete no-op against the real on-disk layout.
>
> This means REQ-MPR-008/009 (model-map cleanup) and REQ-MPR-011a/011b (effort-map prune + reconcile) currently have **zero production effect** — the map contents are corrected but the dead walker never reads `.claude/agents/moai/` to apply them. The **iter-3 SPEC therefore ABSORBS the walker-target fix (REQ-MPR-019)** as in-scope: correcting `domains` to target the real `.claude/agents/moai/` layout is the actual highest-value defect in this SPEC, because without it none of the map cleanup takes live effect. With REQ-MPR-019 applied, the walker reads `moai/`, reaches the retained 7 agent files, and REQ-MPR-008/009/011a/011b take live effect on the next `moai init`/`moai update`. The iter-2 prune-vs-retire SPLIT is retained but its rationale is now honest:
>
> - **(a) Prune phantom keys from `agentEffortMap` + reconcile map↔file divergence** — SAFE, in-scope (REQ-MPR-011a/011b). Remove `manager-strategy`, `expert-security`, `expert-refactoring` (archived phantoms). Sync `plan-auditor`/`sync-auditor` map values from `high` to `xhigh` (matching hand-authored files). Add missing retained `manager-develop` (`xhigh`) and `builder-harness` (`high`).
> - **(b) Fix the dead domain-walker to target `.claude/agents/moai/`** — SAFE, in-scope, REQ-MPR-019 (iter-3 absorbed). Both `ApplyEffortPolicy` and `ApplyModelPolicy` share the same walker bug and MUST be fixed together (same `domains` slice). This restores live injection so (a) actually takes effect.
> - **(c) Retire `agentEffortMap` + `ApplyEffortPolicy` entirely** — UNSAFE given the 2 production callers. Downgraded to **analysis-only observation; full retirement deferred to a follow-up SPEC** (`SPEC-CC2178-EFFORT-MAP-RETIREMENT-001`) that first migrates the 2 callers (`initializer.go:181`, `update.go:2661`) to an alternative effort-injection path. REQ-MPR-012 records this deferral rationale only.

> **REQ-MPR-011** (Ubiquitous, MUST) — renamed **REQ-MPR-011a (prune phantoms, SAFE)** in iter-2
> The SPEC shall prune the 3 archived phantom keys (`manager-strategy`, `expert-security`, `expert-refactoring`) from `agentEffortMap`, sync the map values for `plan-auditor` and `sync-auditor` from `high` to `xhigh` (matching the hand-authored agent files), and add the missing retained agents `manager-develop` (effort `xhigh`) and `builder-harness` (effort `high`) to the map so that the map is internally consistent with the deployed agent files and the retained 8-agent catalog.

> **REQ-MPR-011b** (Unwanted, MUST) — **map↔file divergence reconciliation**
> The `agentEffortMap` shall not contain entries whose value contradicts the hand-authored `effort:` field in the corresponding agent file for retained agents (e.g., the map MUST NOT say `plan-auditor: high` while `.claude/agents/moai/plan-auditor.md` says `effort: xhigh`) so that a fresh `moai init` deployment injects the same effort value the maintainer hand-authored. The reconciliation direction is map-←-file (the file is the authoritative human intent; the map is the deployment injector).

> **REQ-MPR-012** (Ubiquitous, SHOULD) — **downgraded in iter-2 from "retire the map" to "record deferral"; iter-3 rationale re-grounded**
> The SPEC shall, in its plan-phase research, record the deferral rationale for NOT retiring `agentEffortMap` + `ApplyEffortPolicy`: (1) the 2 production callers (`initializer.go:181`, `update.go:2661`) that would need migration, (2) the follow-up SPEC candidate name (`SPEC-CC2178-EFFORT-MAP-RETIREMENT-001`). The iter-2 "regression risk" rationale is re-grounded in iter-3: the walker is currently a no-op (D11), so the deferral is about caller-migration effort, NOT about preserving live injection. Full retirement is out-of-scope for THIS SPEC.

> **REQ-MPR-019** (Capability gate, MUST) — **NEW in iter-3 (D11 remediation): dead domain-walker fix**
> Where the agent-definition directory layout is consolidated under `.claude/agents/moai/` (the real on-disk layout post-AGENT-FOLDER-SPLIT revert, verified 2026-06-16), the `ApplyEffortPolicy` function (`model_policy.go:137`) and the `ApplyModelPolicy` function (`model_policy.go:246`) SHALL iterate the real `moai/` directory instead of the four dead subdirectories (`core`, `expert`, `meta`, `harness`) that no longer exist on disk. The exact change is `domains := []string{"moai"}` (both sites) — or an equivalent mechanism robust to the current consolidated layout. The walker MUST reach the retained 7 agent files under `.claude/agents/moai/` so that REQ-MPR-008/009 (model-map cleanup) and REQ-MPR-011a/011b (effort-map prune + reconcile) take live effect on the next `moai init`/`moai update`, restoring the live-injection behavior the iter-2 SPEC assumed but that the dead walker silently defeated.

### §C.5 [1m] Constraint Re-Verification

> **REQ-MPR-013** (Ubiquitous, MUST) — **iter-3 timing reworded to match plan.md §F.1 M1 schedule**
> The SPEC shall include a research task that re-verifies the `#45847` `[1m]` entitlement-inheritance constraint against the CC 2.1.178 runtime (fetching the actual upstream issue state and CHANGELOG at 2.1.178) **before run-phase M4 GREEN** (the Default-model cost-lever milestone that depends on the verdict), and record the verdict (still-active / relaxed / partially-relaxed) in the M1 research notes. The re-verification is scheduled as run-phase M1 (per plan.md §F.1) because it requires fetching live upstream GitHub issue state — a network/runtime task genuinely belonging in run-phase, not plan-phase doc authoring. The conservative default (EC-01: fetch failure → assume "still-active" → EX-01 preserved) means the SPEC is safe regardless of verdict.

> **REQ-MPR-014** (State-driven, MUST) — **rewritten in iter-2 to add new information beyond EX-01 (D10 remediation); the iter-1 version was tautological with EX-01**
> While the `[1m]` re-verification verdict is "still-active" (constraint unchanged at 2.1.178), the SPEC shall scope per-agent pinning as out-of-scope AND record the specific relaxation conditions that WOULD re-enable per-agent pinning in a follow-up SPEC (REQ-MPR-015): namely (a) the `availableModels`/`enforceAvailableModels` settings fields are confirmed to allow per-agent `model:` overrides that escape the allowlist without breaking `[1m]` inheritance at CC 2.1.178+, AND (b) a caller-migration path exists for any agent whose pinned model differs from Default. This adds testable substance beyond the bare EX-01 exclusion: a reviewer can verify the recorded relaxation conditions are concrete and falsifiable, not a restatement of "per-agent pins are forbidden".

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
- **AC-MPR-003** (MUST): Deployed settings Default model is `sonnet`. **Verification command is CONDITIONAL — pinned at M1** once REQ-MPR-013 confirms the exact CC 2.1.178 Default-resolution JSON key (post-2.1.175 the settings file carries `model`, `availableModels`, and possibly other model-named keys, so a naive `grep '"model"'` matches multiple lines). The M1 research note MUST record the confirmed key and the exact grep command.
- **AC-MPR-004** (MUST): Harness `minimal` resolves to a lightweight cycle_type (not full-TDD), verified via the `ResolveCycleType` symbol contract defined in plan.md §F.3 M3.
- **AC-MPR-005** (MUST): Harness `thorough` resolves to full TDD, verified via `ResolveCycleType`.
- **AC-MPR-006** (MUST): Explicit `quality.yaml constitution.development_mode: tdd` pin is preserved (backward-compat).
- **AC-MPR-007** (MUST): `agentModelMap` has 0 of the 16 canonical phantom keys (grep-count = 0) — see the canonical enumeration table in §C.3.
- **AC-MPR-008** (MUST): `GetAgentModel("manager-develop")` and `GetAgentModel("builder-harness")` return non-empty (tuples `{sonnet, sonnet, haiku}` per REQ-MPR-009 iter-2 decision).
- **AC-MPR-009** (MUST): `agentEffortMap` has 0 archived-phantom keys.
- **AC-MPR-010** (MUST): Effort-map phantom pruning + map↔file reconciliation is applied (REQ-MPR-011a/011b); full-retirement deferral rationale is recorded (REQ-MPR-012 downgraded).
- **AC-MPR-011** (MUST): `[1m]` re-verification verdict is recorded (still-active / relaxed / partially-relaxed) with upstream evidence.
- **AC-MPR-012** (SHOULD): Task-triage in-scope/deferred decision is recorded with rationale; if in-scope, the triage signal is defined concretely.
- **AC-MPR-013** (SHOULD): `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` exits 0 with no MUST-FIX findings (path-prefixed invocation — the bare-ID form `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001` ParseFails per D3 verification on 2026-06-16).
- **AC-MPR-014** (MUST, NEW in iter-2 — D7 remediation): Harness `standard` resolves to `tdd` (the current default), verified via `ResolveCycleType("standard") == "tdd"`. Binds the previously-orphaned REQ-MPR-007.
- **AC-MPR-015** (MUST, NEW in iter-3 — D11 remediation): The dead domain-walker is fixed — both `ApplyEffortPolicy` (`model_policy.go:137`) and `ApplyModelPolicy` (`model_policy.go:246`) iterate `.claude/agents/moai/` (NOT the dead `{core,expert,meta,harness}`). Verified by grepping the `domains` slice in `model_policy.go` asserting `"moai"` is present and the four dead dirs are gone, AND a test asserting the walker reads `.claude/agents/moai/`. Binds REQ-MPR-019.

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
- 2026-06-16: **iter-2 plan-audit remediation (D1-D10)**. plan-auditor iter-1 returned FAIL 0.66 (threshold 0.80). This iter-2 revision resolves all 3 BLOCKING defects (D1 `resolveCycleType` non-existent symbol → `ResolveCycleType` contract defined in plan.md M3; D2 phantom-key enumeration unified to 16 canonical keys across REQ/AC/§D.3/plan §B.1; D3 lint command pinned to path-prefixed form) and all 4 SHOULD-FIX defects (D4 AC-MPR-003 made conditional pending M1 key-confirmation; D5 effort-map retirement premise FACTUALLY INVERTED — `ApplyEffortPolicy` has 2 production callers, full retirement deferred, only phantom-pruning is in-scope; D6 `manager-develop` model tuple resolved to `{sonnet, sonnet, haiku}` to align with Default-Sonnet thesis; D7 AC-MPR-014 added for orphaned REQ-MPR-007). Plus 3 MINOR defects (D8 file count 10→15, D9 estimator file named, D10 REQ-MPR-014 rewritten to add relaxation-condition substance). Auditor prose-precision note applied: §A D4 corrected to reflect `development_mode`/`enforce_quality` are nested under `constitution:` not flat globals. No scope expansion (Tier M retained; auditor confirmed no forced split). Commit `1f6b59a47` left unamended; this iter-2 is a NEW commit.
- 2026-06-16: **iter-3 plan-audit remediation (D11 walker-absorb + D12 + D13 + REQ-MPR-013 timing)**. plan-auditor iter-2 returned FAIL 0.71 (threshold 0.80) — full report at `.moai/reports/plan-audit/SPEC-CC2178-MODEL-POLICY-REPAIR-001-iter2.md`. This iter-3 revision resolves 1 CRITICAL + 1 MAJOR + 1 MINOR + 1 internal contradiction: **D11 (CRITICAL)** — the iter-2 §C.4 D5 reframe's load-bearing premise ("retiring `agentEffortMap`/`ApplyEffortPolicy` = new deployments lose effort injection = regression") was FACTUALLY FALSE in the current tree: `ApplyEffortPolicy` (L137) and `ApplyModelPolicy` (L246) both hardcode `domains := []string{"core", "expert", "meta", "harness"}` — subdirectories that DO NOT EXIST post-AGENT-FOLDER-SPLIT revert (agents live in `.claude/agents/moai/`). The walker is a complete no-op, defeating REQ-MPR-008/009/011a/011b. Per user decision (OPTION A — absorb walker fix), this SPEC adds **REQ-MPR-019** (MUST, M2) + **AC-MPR-015** fixing `domains` to `[]string{"moai"}` at both L137 and L246, restoring live injection so the map cleanup actually takes effect. §C.4 rationale rewritten honest; EX-05 rewritten to clarify walker-target fix IS in-scope, full retirement stays out-of-scope. **D12 (MAJOR)** — plan.md §B.2 cited 3 stale agent-file paths (`.claude/agents/meta/plan-auditor.md`, `.claude/agents/meta/sync-auditor.md`, `.claude/agents/builder/builder-harness.md`) → corrected to `.claude/agents/moai/` (the `meta/` and `builder/` subdirs do not exist). The stale glob `.claude/agents/{meta,moai,builder}/*.md` → `.claude/agents/moai/*.md`. Effort values cited were already correct. **D13 (MINOR)** — REQ-MPR-018 (effort-axis doc tuning, SHOULD) had NO acceptance criterion; bound via **AC-MPR-016** (SHOULD, M5) + §D.2 traceability matrix row added. **Internal contradiction** — REQ-MPR-013 said the `#45847` re-verification happens "before plan-phase is marked audit-ready" but plan.md §F.1 schedules it as M1 (run-phase); reworded to "before run-phase M4 GREEN" for logical consistency. Version 0.2.0 → 0.3.0. Tier M retained (REQ-MPR-019 + 2 ACs is a modest scope addition; no Tier-L inflation). frontmatter `status: draft` unchanged; `updated: 2026-06-16`.
