---
id: SPEC-MODEL-TIER-PLANTYPE-001
title: "plan_type-aware model tier profiles (API vs subscription) for moai init / moai web"
version: "0.3.1"
status: completed
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.1.x model-policy redesign"
module: "internal/config + internal/template + internal/cli + internal/web + internal/tmux"
lifecycle: spec-anchored
era: V3R6
tier: L
related_specs: [SPEC-CC2178-MODEL-POLICY-REPAIR-001, SPEC-AGENT-ARCH-V2-001, SPEC-TOKEN-ROUTING-001, SPEC-WEB-CONSOLE-013, SPEC-GLM-MODEL-ALLOWLIST-001]
tags: "model-policy, plan-type, tier-profiles, fable, config, cli-init, wizard, web-console, glm, effort-overlay, reasoning-effort"
---

# SPEC-MODEL-TIER-PLANTYPE-001 — plan_type-aware model tier profiles (API vs subscription)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-12 | manager-spec | Initial draft — encodes the user-approved model tier redesign (rev2) from `.moai/reports/model-tier-redesign-20260712.md` |
| 0.2.0 | 2026-07-12 | manager-spec | Kickoff clarifications + plan-audit iter-1 (FAIL 0.88) fixes. D1 resolved: replace-both precedence confirmed. **D2 scope change (user decision)**: `/model-policy` gains a single, narrowly-scoped write path persisting exactly `llm.plan_type` — a sanctioned, user-approved exception to the SPEC-WEB-CONSOLE-013 REQ-WC13-021 read-only doctrine; all other read-only invariants of that view are preserved (WHY: recorded in §B.4). D3 resolved: `moai update` gains a `--plan-type` override flag (REQ-MTP-018 amended). D5 resolved: M5/REQ-MTP-025 descoped to follow-up SPEC-MODEL-TIER-ROUTING-PROFILES-001 (§C exclusion added). Auditor MUST-fixes: REQ-MTP-012 legacy mapping is tier-only (plan selection owned by plan_type resolution); AC baselines re-measured (update.go grep = 3; modelpolicy.go `PersistTarget` = 1). SHOULD-fixes: AC-MTP-023a grep rewritten; §B.6/§B.7 note citation corrected D4→D6; AC-MTP-004b grep form fixed; M2 absorbs the dangling `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001` deferral pointer. |
| 0.3.0 | 2026-07-12 | manager-spec | **New dimension (user decision): GLM backend effort overlay.** Adds a backend-conditional overlay (NOT a third plan_type — plan_type stays {api, subscription}) that remaps the per-agent EFFORT value the active plan_type profile produced, applied ONLY when the session backend is GLM. New §B.8 REQ-MTP-026..030: GLM backend detection (via the `llm.team_mode` / `llm.mode` signals); the 5-level Claude effort → GLM collapse mapping as named constants (low→thinking-off; medium/high→reasoning-high; xhigh/max→reasoning-max); the coding-max override set ({manager-develop, builder-harness} → reasoning-max) as a named constant; overlay scope (effort-only — the MODEL column is already handled by `llm.glm.models` tier→GLM mapping); and the effort-injection mechanism marked as a **run-phase empirical gate** (two branches: A shim passthrough / B explicit `reasoning_effort`/`thinking` write at the GLM launch path). New milestone M5 (final; the former M5 was descoped at 0.2.0). REQ count 24→29 (026–030; 025 remains retired). AC groups 26→31 (028–032 added). **Tier M → L** (adds `research.md` documenting the z.ai reasoning-control findings + collapse-design rationale + shim-injection risk). related_specs += SPEC-GLM-MODEL-ALLOWLIST-001. Reopens plan-audit (iteration 3). |
| 0.3.1 | 2026-07-12 | manager-spec | **plan-audit iter-3 (FAIL 0.83) bounded-defect fix (D1/D2).** Corrected the REQ-MTP-026 GLM detection predicate: the `moai glm` case was keyed off `LLMConfig.Mode == "glm"`, but that field is DORMANT (zero non-test writer/reader — verified on the live tree). `moai glm` actually persists `llm.team_mode = "glm"` via `persistTeamMode` (glm.go:334 → llmCfg.TeamMode), so the old predicate resolved FALSE for the PRIMARY all-GLM `moai glm` session (`mode="" / team_mode="glm"`) — an inert overlay in the headline mode (the exact inert-headline hazard §A.3 warns of). Corrected predicate: `TeamMode ∈ {"cg", "glm"}` (the real persisted signals — glm→"glm", cg→"cg") OR `Mode == "glm"` (defensive OR for the dormant/future field). AC-MTP-028 truth table row 5 flipped `(mode="", team_mode="glm") → FALSE` to `→ TRUE`; genuinely-FALSE cases now use `team_mode ∈ {"claude", "hybrid", ""}`. §D.2 edge case synced. False-premise ("moai glm writes llm.mode") corrected on 3 surfaces (§A.5, plan.md §A.10, research.md §E). No REQ/AC/milestone COUNT change (29 REQ / 31 AC groups / 5 milestones unchanged). Reopens plan-audit (iteration 4, final). |

## §A Context & Goal

### A.1 Problem

The current model policy pipeline assumes a single billing context. It splits the per-agent
model decision and the per-agent effort decision into two disconnected maps
(`agentModelMap` `[3]string` high/medium/low + `agentEffortMap`) applied by two separate
passes (`ApplyModelPolicy` / `ApplyEffortPolicy`) with divergent precedence semantics.
DeepSWE leaderboard evidence (113 tasks, 2026-07-09) shows the optimal {model, effort}
assignment differs fundamentally between API metered billing (dollars are the constraint —
Opus 4.8 wins $/solved-task) and subscription billing (weekly token quota with Opus-weighted
deduction is the constraint — the opusplan pattern wins). One matrix cannot serve both.

### A.2 Approved design (SSOT)

`.moai/reports/model-tier-redesign-20260712.md` (rev2, user-approved 2026-07-12) defines:

- **Plan A (api)** — 3 tiers (A-max / A-medium ★recommended / A-low). rev2 change: execution
  agents moved from sonnet to opus (per-task cost inversion: opus $13.22 < sonnet $26.40
  despite lower unit price).
- **Plan B (subscription)** — 3 tiers (B-max ★recommended / B-medium / B-low). Execution
  stays sonnet high (Opus quota is the scarce resource; opusplan structure).
- Both matrices assign explicit {model, effort} to all 10 retained agents (auditors and
  manager-design move from implicit `inherit` to explicit assignment; Explore remains
  `inherit`).

This SPEC encodes the approved matrices verbatim (§B.6/§B.7). The matrix VALUES are settled
design input — not re-litigated here.

### A.3 Scope surface (WHAT changes, observable)

1. New config field `llm.plan_type` ∈ {api, subscription}; absent → subscription.
2. `"fable"` joins the model routing closed set (`validRoutingModels`).
3. A single plan_type × tier profile structure replaces the dual
   `agentModelMap`/`agentEffortMap`; one unified apply pass yields {model, effort}
   per agent atomically.
4. `moai init` gains `--plan-type` flag + a wizard plan-type question; `--model-policy`
   (max/medium/low) keeps its current contract.
5. `moai web` `/model-policy` view shows the active plan_type and a per-tier
   {model, effort} preview for both plans, derived from the Go profile structure; the
   plan_type selector PERSISTS the chosen value through exactly one scoped write path
   (D2 user decision at kickoff — see §B.4 WHY note).
6. `moai update` accepts a `--plan-type` override that persists (D3 user decision).
   The formerly-optional plan_type-specific `model_routing_profiles` 36-cell derivation
   is DESCOPED (D5) to follow-up SPEC-MODEL-TIER-ROUTING-PROFILES-001 — see §C.
7. **GLM backend effort overlay (v0.3.0, §B.8).** A backend-conditional overlay remaps
   the per-agent EFFORT value the active plan_type profile produced, applied ONLY when the
   session backend is GLM (`llm.team_mode ∈ {"glm", "cg"}` — the actual persisted signals;
   `llm.mode == "glm"` retained only as a defensive OR for the currently-dormant field). It collapses
   the 5-level Claude effort vocabulary to GLM's binary thinking toggle + 2-level
   `reasoning_effort` (z.ai supports only `{high, max}`), with a coding-max override for
   `manager-develop` and `builder-harness`. It touches EFFORT only — the model dimension is
   already carried under GLM by the `llm.glm.models` tier→GLM env mapping
   (`ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`). The effort-injection MECHANISM
   (does the effort reach z.ai as `reasoning_effort`?) is an unverified run-phase gate with
   two branches (§B.8 / plan.md D7).

### A.4 Baseline invariants (measured 2026-07-12, see plan.md §A for evidence)

- `validRoutingModels` = {inherit, sonnet, opus, glm}; zero `"fable"` occurrences in
  `internal/config/model_routing.go`.
- `LLMConfig` has no plan-type field; template `llm.yaml` has no `plan_type` key.
- All 9 shipped agent template files carry BOTH `model:` and `effort:` frontmatter (9/9).
- `--model-policy` canonical CLI enum is already {max, medium, low} (performance tier);
  the legacy {high, medium, low} vocabulary survives only in `template.ModelPolicy`
  (wizard/prefs/web-launch surfaces).
- `/model-policy` web view is READ-ONLY (405 non-GET) per SPEC-WEB-CONSOLE-013 REQ-WC13-021.

### A.5 GLM-overlay baseline invariants (measured 2026-07-12, see plan.md §A.10)

- GLM backend detection fields already exist, but the ACTUAL persisted signal is
  `LLMConfig.TeamMode` (yaml `llm.team_mode`), NOT `LLMConfig.Mode`: both `moai glm` and
  `moai cg` write through `persistTeamMode`, which assigns `llmCfg.TeamMode` — `moai glm`
  writes `llm.team_mode = "glm"` (glm.go `persistTeamMode(root, "glm")`) and `moai cg`
  writes `llm.team_mode = "cg"`. The `LLMConfig.Mode` field (yaml `llm.mode`) is currently
  **DORMANT**: it has ZERO non-test writer and ZERO non-test reader (verified 2026-07-12 —
  no `moai glm` path sets it). `internal/tmux/cg_detect.go` `IsCGMode` reads
  `team_mode == "cg"` (corroborated by tmux session GLM markers). NOTE the `LLMConfig.TeamMode`
  godoc comment is stale (`"", "claude", "glm", "hybrid"`); the values actually written by
  `persistTeamMode` are `"cg"` / `"glm"` / `""` (see plan.md §A.10 — godoc + dormant-field
  comment fixes are adjacent M5 notes, not blocking clarifications).
- The 5-level Claude effort vocabulary already exists as named constants in
  `internal/template/model_policy.go`: `EffortLevelLow/Medium/High/XHigh/Max` = `low`/`medium`/
  `high`/`xhigh`/`max`. These are the collapse-function INPUT domain.
- GLM env/settings are injected at the `internal/cli/glm.go` launch path — `setGLMEnv`
  (direct `moai glm`, process `os.Setenv`) and `injectGLMEnvForTeam` (`moai cg`,
  `.claude/settings.local.json` `env` block). NEITHER injects any effort/reasoning control
  today: `grep -c 'reasoning_effort' internal/cli/glm.go` → 0; `grep -c 'thinking'
  internal/cli/glm.go` → 0. `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`
  (`internal/config/defaults.go`) — the Anthropic-compat shim.
- `moai cc` teardown clears GLM env via `buildTmuxClearVars()` (the REQ-CGH-009 inject↔clear
  key-parity invariant); any Branch-B reasoning key MUST be added there for symmetry.
- `reasoning_effort` / GLM `thinking`-toggle appear in ZERO non-test Go files repo-wide
  (baseline 0 for both concepts).

## §B Requirements (GEARS)

### B.1 Config field `llm.plan_type` (Milestone M1)

- **REQ-MTP-001** — The `internal/config` LLM section loader shall expose a `plan_type`
  field (Go: `LLMConfig.PlanType`, YAML: `llm.plan_type`) whose valid value set is the
  closed set {`api`, `subscription`}.
- **REQ-MTP-002** — **While** `llm.plan_type` is absent or empty, the config layer shall
  resolve the effective plan type to `subscription` (backward compatibility: existing
  projects select the subscription branch without any config edit).
- **REQ-MTP-003** — **When** an out-of-set `plan_type` value is detected during config
  validation, the validator shall return an error naming the offending value and the
  closed set {api, subscription}.
- **REQ-MTP-004** — Where the template distribution is concerned, the template source
  `internal/template/templates/.moai/config/sections/llm.yaml` shall carry the
  `plan_type` key with a neutral explanatory comment (Template-First rule; no internal
  SPEC IDs in template content).
- **REQ-MTP-005** — The model routing closed set (`validRoutingModels` in
  `internal/config/model_routing.go`) shall include `"fable"`, and BOTH validation paths
  (`ValidateModelRouting` flat map + `ValidateModelRoutingProfiles`) shall accept
  `model: fable`; the closed-set error-message literals in both paths shall name the
  updated set (the flat-path message currently names a stale `haiku` member that is not
  in the set — it shall be corrected in the same edit).
- **REQ-MTP-024** — The `workflow.yaml` closed-set documentation comments (local
  `.moai/config/sections/workflow.yaml` AND its template twin) shall name `fable` as a
  valid `model_routing` / `model_routing_profiles` model value.

### B.2 Unified plan_type × tier profile structure (Milestone M2)

- **REQ-MTP-006** — `internal/template` shall define a single tier-profile structure keyed
  by (plan_type ∈ {api, subscription}) × (tier ∈ {max, medium, low}) that yields one
  atomic {model, effort} pair per agent, replacing the dual `agentModelMap`
  (`[3]string`) + `agentEffortMap` representation.
- **REQ-MTP-007** — The `api` profile values shall equal the Plan A rev2 matrix in §B.6
  verbatim (all 30 {model, effort} cells).
- **REQ-MTP-008** — The `subscription` profile values shall equal the Plan B matrix in
  §B.7 verbatim (all 30 {model, effort} cells).
- **REQ-MTP-009** — The profile structure shall carry explicit rows for all 10 retained
  agents; `plan-auditor`, `sync-auditor`, `manager-design`, and `super-advisor` move from
  map-absence (implicit inherit) to explicit assignment, and `Explore` shall remain
  `model: inherit` (Explore has no agent file — its row informs display/derivation
  surfaces only, and the apply pass shall skip it).
- **REQ-MTP-010** — A single `ApplyTierProfile` pass shall replace the two-pass
  `ApplyModelPolicy` + `ApplyEffortPolicy` sequence at BOTH production call sites
  (`internal/core/project/initializer.go` and `internal/cli/update.go`), patching each
  known agent file's `model:` and `effort:` frontmatter in one traversal and updating
  manifest hashes for every written file.
- **REQ-MTP-011** — The unified pass shall define and document one explicit precedence
  rule for pre-existing frontmatter values: it shall REPLACE both existing `model:` and
  existing `effort:` lines with the profile values (the tier profile is the SSOT for
  shipped-agent model/effort). Rationale: 9/9 shipped agent files carry `effort:`
  frontmatter, so the historical preserve-existing-effort rule would render the entire
  effort matrix unreachable (inert-headline hazard). The precedence rule shall be stated
  in the function's godoc. (Plan.md D1 — RESOLVED at kickoff 2026-07-12: replace-both
  confirmed by user.)
- **REQ-MTP-012** — **When** a legacy `template.ModelPolicy` value ({high, medium, low})
  reaches the tier-profile pass (wizard/prefs/web-launch surfaces), a single mapping
  function shall translate it to the TIER value ONLY (high→max, medium→medium,
  low→low). The mapping function shall NOT select a plan — plan selection is owned
  solely by the effective plan_type resolution (REQ-MTP-002/015/018).
  `template.ModelPolicy` and `IsValidModelPolicy` shall remain exported for the
  untouched legacy consumer surfaces.
- **REQ-MTP-013** — **When** the apply pass encounters an agent file whose name is not in
  the profile (unknown/user-added agent), it shall leave the file byte-identical
  (preserve current skip semantics).

### B.3 CLI flag, wizard, persistence (Milestone M3)

- **REQ-MTP-014** — `moai init` shall accept a new `--plan-type` flag with the closed set
  {api, subscription}; **When** an out-of-set value is detected at flag validation, the
  command shall exit non-zero with a usage error naming the closed set.
- **REQ-MTP-015** — The existing `--model-policy` flag contract shall remain unchanged:
  canonical values {max, medium, low} plus the deprecated one-cycle boolean aliases
  (`--high`→max, `--medium-alias`→medium, `--low`→low); the resolved tier shall select
  the row within the plan chosen by the effective plan_type.
- **REQ-MTP-016** — **When** `moai init` resolves an explicit plan type (flag or wizard),
  the CLI shall persist it to `.moai/config/sections/llm.yaml` as `plan_type: <value>`
  (same persistence pattern as `ApplyPerformanceTier`).
- **REQ-MTP-017** — Where the interactive init wizard runs, it shall ask one plan-type
  question (api vs subscription) whose default/recommended option is `subscription`,
  and the answer shall flow into the same resolution path as the `--plan-type` flag
  (flag takes precedence over wizard).
- **REQ-MTP-018** — **When** `moai update` re-applies the tier profile, it shall read the
  effective plan type from the persisted `llm.plan_type` (absent → subscription) by
  default; AND `moai update` shall accept a `--plan-type` override flag (closed set
  {api, subscription}) which, when provided, takes precedence for that run AND persists
  the new value to `llm.plan_type`. **When** an out-of-set override value is detected at
  flag validation, the command shall exit non-zero naming the closed set. (Plan.md D3 —
  RESOLVED at kickoff 2026-07-12: user chose the override flag.)

### B.4 moai web model-policy board (Milestone M4)

> **WHY — sanctioned exception to the SPEC-WEB-CONSOLE-013 read-only doctrine.** The
> `/model-policy` view was pinned READ-ONLY by SPEC-WEB-CONSOLE-013 REQ-WC13-021 (no
> write path, no PersistTarget, no form control). At this SPEC's Implementation Kickoff
> (2026-07-12), the user explicitly chose a PERSISTING plan_type selector over the
> recommended read-only preview toggle (plan.md D2 — RESOLVED). The exception is
> deliberately narrow: exactly ONE field (`llm.plan_type`) becomes writable through
> exactly ONE dedicated endpoint; every other read-only invariant of the view —
> read-only routing tables, no FieldDef enrollment for other fields, 405 on unsupported
> methods for the page route — is preserved. This paragraph is the documented
> reconciliation with the 013 doctrine; the 013 SPEC itself is NOT amended (its
> invariant remains authoritative for every surface except this sanctioned endpoint).

- **REQ-MTP-019** — The `/model-policy` view shall display the active plan type read from
  `llm.plan_type`; **While** the field is absent or empty, the view shall render a
  "(default: subscription)" style label instead of an empty value.
- **REQ-MTP-020** — The `/model-policy` view shall render a plan_type selector and a
  per-tier {model, effort} preview covering BOTH plans (each plan: 10 agents × 3 tiers),
  with the active plan type pre-selected; **When** the user submits a plan-type change
  through the selector, the view shall persist the new value to
  `.moai/config/sections/llm.yaml` `plan_type` via the scoped persist endpoint
  (REQ-MTP-021) and re-render with the new active plan. (Plan.md D2 — RESOLVED at
  kickoff 2026-07-12: user chose the persisting selector.)
- **REQ-MTP-021** — The `/model-policy` surface shall expose exactly ONE write path: a
  dedicated persist endpoint `POST /model-policy/plan-type` scoped to the `llm.plan_type`
  field alone (the sanctioned exception above). The endpoint shall accept only values
  from the closed set {api, subscription}; **When** an out-of-set value is submitted,
  the endpoint shall reject with a 4xx status and leave `llm.yaml` byte-identical.
  **When** a persist succeeds, no `llm.yaml` content other than the `plan_type` line
  shall change. No OTHER field shall become writable through this view, and the page
  route `/model-policy` itself shall continue to reject non-GET methods with 405.
- **REQ-MTP-022** — Every new user-facing label introduced on the view shall carry a
  `data-i18n` key with translations present in all 4 locales (en/ko/ja/zh) in
  `internal/web/assets/i18n.js`.
- **REQ-MTP-023** — The preview cell values shall be derived from the Go tier-profile
  structure at render time (single source); the web layer shall NOT re-declare the
  matrix as a second literal.

### B.5 Routing-profile derivation — DESCOPED (former M5 / REQ-MTP-025)

REQ-MTP-025 (plan_type-specific `workflow.model_routing_profiles` 36-cell derivation
per design §6) was REMOVED at kickoff (plan.md D5 — RESOLVED 2026-07-12: descope
confirmed). The requirement transfers to follow-up SPEC-MODEL-TIER-ROUTING-PROFILES-001
and no artifact in this SPEC claims it. See §C "Out of Scope — plan_type-specific
routing-profile derivation". The REQ ID `REQ-MTP-025` is retired and MUST NOT be
reused within this SPEC.

### B.6 Plan A (api) profile matrix — rev2 (verbatim from design §3)

| Agent | A-max | A-medium ★recommended | A-low |
|---|---|---|---|
| manager-spec | fable / high | fable / high | opus / high |
| plan-auditor | fable / high | fable / high | opus / high |
| sync-auditor | fable / high | opus / high | opus / medium |
| manager-design | fable / high | fable / high | opus / high |
| super-advisor | fable / xhigh | fable / high | opus / high |
| manager-develop | fable / high | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / medium |
| manager-docs | sonnet / medium | sonnet / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | inherit / medium | inherit / low | inherit / low |

### B.7 Plan B (subscription) profile matrix (verbatim from design §4)

| Agent | B-max ★recommended | B-medium | B-low |
|---|---|---|---|
| manager-spec | opus / high | opus / high | opus / medium |
| plan-auditor | opus / high | opus / medium | sonnet / high |
| sync-auditor | opus / high | opus / medium | sonnet / high |
| manager-design | opus / high | opus / medium | sonnet / high |
| super-advisor | opus / xhigh | opus / high | opus / medium |
| manager-develop | sonnet / high | sonnet / high | sonnet / high |
| builder-harness | sonnet / high | sonnet / medium | sonnet / medium |
| manager-docs | sonnet / low | sonnet / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | inherit / medium | inherit / low | inherit / low |

Notes binding both matrices:

- The tier axis {max, medium, low} is the existing performance-tier vocabulary
  (`--model-policy` canonical enum); "A-max" ≡ (plan_type=api, tier=max), etc.
- ★recommended marks the wizard's recommended option per plan (api→medium,
  subscription→max). The default-when-absent tier remains `medium` for both plans
  (preserves the template `performance_tier: "medium"` default — plan.md D6 design
  note; distinct from plan.md D4, which is the PerformanceTier validate-tag fix).
- Model values are Claude Code short aliases (resolution to canonical IDs stays with the
  existing `ModelAliasTable`, which already maps `fable` → `claude-fable-5`).
- The subscription matrices intentionally SUPERSEDE the legacy `agentModelMap` tuples;
  "backward compatibility" (REQ-MTP-002) means absent `plan_type` selects the
  subscription BRANCH — it does not mean byte-identical apply-pass output. The redesign
  of the values is the purpose of this SPEC.
- Neither matrix contains `haiku` — consistent with the No-Haiku 3-tier policy
  (SPEC-AGENT-ARCH-V2-001).

### B.8 GLM backend effort overlay (Milestone M5)

> **WHY — GLM does not support Claude's 5-level effort.** Under a GLM-backed session
> (`moai glm` / `moai cg`), sub-agent model calls route to z.ai via the Anthropic-compat
> shim (`base_url = https://api.z.ai/api/anthropic`). z.ai's reasoning control is NOT the
> Claude 5-level effort vocabulary — it is a **binary thinking toggle** plus a **2-level
> `reasoning_effort`** (`{high, max}`; `max` is the omit-default; `high` requires an explicit
> value; `reasoning_effort` has no effect when thinking is disabled). z.ai recommends `max`
> for coding tasks and `high` for fast, economical agent loops (sources: research.md §B —
> docs.z.ai/guides/capabilities/thinking-mode + the GLM-5.2 OpenAI-compat `reasoning_effort`
> guide). The plan_type profiles (§B.6/§B.7) emit Claude effort values ({low, medium, high,
> xhigh, max} — actual matrix cells use low/medium/high/xhigh); these cannot be sent verbatim
> to z.ai. This overlay is the collapse. It is an OVERLAY, NOT a third plan_type — plan_type
> stays {api, subscription}, and the overlay sits on top of whichever plan_type profile is
> active, remapping only the effort dimension. The MODEL dimension is already carried under
> GLM by the `llm.glm.models` tier→GLM env mapping (`ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,
> FABLE}_MODEL`, set by `setGLMEnv` / `injectGLMEnvForTeam`), so the overlay does NOT touch
> `model:`.

- **REQ-MTP-026** — `internal/template` shall define a GLM backend-detection predicate that
  resolves TRUE when the effective session backend is GLM, defined as
  `LLMConfig.TeamMode ∈ {"cg", "glm"}` (the ACTUAL persisted GLM signals — `moai glm` writes
  `team_mode = "glm"` and `moai cg` writes `team_mode = "cg"`, both via `persistTeamMode`)
  OR `LLMConfig.Mode == "glm"` (a defensive OR for the currently-dormant `llm.mode` field —
  it has no writer/reader today but is reserved so a future `moai glm` that populates `mode`
  is covered). `team_mode` is the real signal; `mode` is the defensive fallback — the
  predicate MUST NOT rely on `mode` alone (that would leave the primary all-GLM `moai glm`
  session, `mode="" / team_mode="glm"`, undetected — the inert-headline hazard). The predicate
  shall read the two `llm.yaml`-persisted fields only (the config-level intent signals); it
  shall NOT re-implement the stricter runtime `internal/tmux/cg_detect.go` `IsCGMode`
  (tmux-session + GLM-marker) check, and shall cross-reference it. **While** neither field
  indicates GLM (`TeamMode ∉ {"cg", "glm"}` AND `Mode ≠ "glm"`), the predicate shall resolve
  FALSE and the overlay shall be an identity no-op (effort values pass through unchanged).
- **REQ-MTP-027** — `internal/template` shall define the Claude-effort → GLM collapse as a
  single pure function whose mapping is expressed through NAMED CONSTANTS (no magic literals,
  per CLAUDE.local.md §14), covering all 5 Claude effort levels
  (`EffortLevelLow/Medium/High/XHigh/Max`) onto the GLM canonical reasoning state
  {`thinking-off`, `reasoning-high`, `reasoning-max`}:
  - `low` → thinking **disabled** (`thinking: {type: disabled}`)
  - `medium` → `reasoning_effort: high`
  - `high` → `reasoning_effort: high`
  - `xhigh` → `reasoning_effort: max`
  - `max` → `reasoning_effort: max`
  The GLM canonical state and the `thinking` / `reasoning_effort` value tokens shall be
  named constants. The collapse function shall be total over the 5-level input domain (an
  unrecognized effort string maps to the GLM default state, documented in godoc).
- **REQ-MTP-028** — `internal/template` shall define the **coding-max override set** as a
  named constant collection containing exactly `manager-develop` AND `builder-harness`.
  **When** the overlay resolves the GLM effort for an agent in this set under a GLM backend,
  it shall force `reasoning_effort: max` (z.ai coding-task recommendation), OVERRIDING that
  agent's collapse result. The override shall apply to no other agent.
- **REQ-MTP-029** — The overlay shall remap the per-agent EFFORT value produced by the active
  plan_type tier profile (§B.6/§B.7) and shall NOT alter the per-agent MODEL value (the model
  → GLM mapping is owned by `llm.glm.models`). The overlay shall be applied ONLY when the
  REQ-MTP-026 predicate is TRUE; under a non-GLM (Claude) backend the per-agent {model,
  effort} pair is exactly the plan_type profile output (§B.6/§B.7), unchanged.
- **REQ-MTP-030** — The collapse+override overlay shall be WIRED INTO the GLM launch path
  (`internal/cli/glm.go` — the `setGLMEnv` direct-launch path AND the `injectGLMEnvForTeam`
  team path, or a shared helper both call), not merely defined. The MECHANISM by which the
  resolved GLM reasoning state reaches z.ai is an **UNVERIFIED run-phase empirical gate** with
  two branches, and the run-phase MUST empirically determine which holds before claiming the
  overlay is effective (plan.md D7):
  - **Branch A (shim passthrough)** — if Claude Code already forwards the sub-agent `effort:`
    frontmatter to the z.ai shim as `reasoning_effort`/`thinking`, the overlay adjusts the
    effort representation pre-launch so the passthrough yields the collapsed GLM state.
  - **Branch B (explicit write)** — if passthrough does NOT translate effort→reasoning
    control, MoAI shall write the reasoning control (`reasoning_effort` and/or the `thinking`
    toggle) EXPLICITLY at the GLM launch path (env var and/or `.claude/settings.local.json`
    `env` block, mirroring the existing `ANTHROPIC_DEFAULT_*` injection). **When** Branch B
    is chosen, the new reasoning key(s) SHALL be added to `buildTmuxClearVars()` so `moai cc`
    teardown clears them (the REQ-CGH-009 inject↔clear key-parity invariant). **Where** only
    a session-global reasoning-control channel is available (no per-agent granularity through
    the shim), the run-phase shall record this as a documented delivery-granularity limitation
    and derive the session-level `reasoning_effort` from the overlay (the coding-max override
    implies session `reasoning_effort: max` whenever a coding agent is the active spawn); the
    per-agent collapse LOGIC (REQ-MTP-027/028) remains defined and unit-tested regardless of
    the delivery branch.

## §C Exclusions (Out of Scope)

### Out of Scope — Runtime spawn-time routing semantics

- `RouteModelFor` accessor behavior, fallback semantics (`inherit`/`medium` default), and
  the Tier×Phase key structure are unchanged; this SPEC only widens the model closed set
  (`fable`).
- Orchestrator spawn-shape selection (Phase 0.95 Modes) is untouched.

### Out of Scope — plan_type-specific routing-profile derivation

- The plan_type-specific `workflow.model_routing_profiles` 36-cell derivation per design
  §6 (phase→representative-agent mapping; SPEC-Tier ±1-step effort adjustment) is
  DESCOPED from this SPEC (kickoff decision D5, 2026-07-12) and transfers to follow-up
  **SPEC-MODEL-TIER-ROUTING-PROFILES-001** (to be authored). The former REQ-MTP-025 and
  AC-MTP-025 are retired; this SPEC only updates the workflow.yaml closed-set
  documentation comments (REQ-MTP-024).

### Out of Scope — GLM backend model mappings

- The `llm.glm` block (base_url, models, context_windows) and the existing
  `ANTHROPIC_DEFAULT_*` model env injection are untouched. plan_type describes the CLAUDE
  billing context only; the v0.3.0 GLM effort overlay (§B.8) READS the GLM backend signals
  (`llm.mode` / `llm.team_mode`) and remaps effort, but does not redefine any `llm.glm` model
  mapping.

### Out of Scope — GLM effort-overlay boundaries

- The overlay remaps EFFORT only; it does NOT change the per-agent `model:` value, does NOT
  alter the `llm.glm.models` tier→GLM env mapping, and does NOT introduce a third plan_type
  (plan_type stays {api, subscription}).
- Empirically verifying and hardening the delivery MECHANISM beyond selecting Branch A vs
  Branch B (§B.8 REQ-MTP-030) — e.g. a per-agent reasoning-control channel if only
  session-global delivery proves available — is a run-phase determination and, where a new
  channel must be built, follow-up work, not part of this SPEC's collapse-logic scope.
- Tuning the collapse thresholds against measured GLM output quality (e.g. whether `medium`
  should map to `reasoning-max` instead of `reasoning-high`) is post-deployment calibration,
  out of scope (mirrors the existing "Empirical effort calibration" exclusion below).

### Out of Scope — Web Launch section model_policy surface

- The settings-schema `model_policy` select (`internal/settings/schema.go`,
  `internal/web/validate.go` Launch section) keeps the legacy {high, medium, low}
  vocabulary; only the mapping function (REQ-MTP-012) bridges it. Redesigning that
  surface belongs to the web-console SPEC line.

### Out of Scope — Empirical effort calibration

- The design's stated limitation (no leaderboard data for effort variants) is
  acknowledged; post-deployment measurement via `.moai/state/verify/` and any resulting
  matrix revision is follow-up work, not part of this SPEC.

### Out of Scope — Agent body/prompt content changes

- Only `model:` / `effort:` frontmatter lines of agent files are patched; agent body
  prose, tools lists, and skills lists are untouched.

### Out of Scope — docs-site documentation

- 4-locale user documentation for plan_type belongs to the docs-site sync line, not this
  SPEC.

## §D Acceptance Criteria

The canonical AC enumeration lives in `acceptance.md` (AC-MTP-001 .. AC-MTP-032, with
paired sub-criteria using the `a`/`b` suffix convention; AC-MTP-025 retired with the D5
descope). Summary of gate structure:

- M1: config field + validation + fable enum (AC-MTP-001..005b, 024)
- M2: profile structure + matrices fidelity + unified apply pass (AC-MTP-006..013)
- M3: CLI flags (init + update) + wizard + persistence (AC-MTP-014..018b)
- M4: web preview + persisting selector + scoped-write invariant + i18n + reachability
  (AC-MTP-019..023b, 027)
- M5: GLM backend detection + collapse mapping + coding-max override + overlay scope +
  injection reachability/run-gate (AC-MTP-028..032b)
- Global: full test suite, vet/lint, coverage, spec-lint delta (AC-MTP-026)

## §E Cross-references

- Design SSOT: `.moai/reports/model-tier-redesign-20260712.md` (rev2)
- GLM reasoning-control research (Tier L artifact): `research.md` (this SPEC dir) —
  z.ai binary thinking + 2-level `reasoning_effort`, the shim-injection risk, and the
  collapse-design rationale, with cited sources (docs.z.ai/guides/capabilities/thinking-mode;
  GLM-5.2 OpenAI-compat `reasoning_effort` guide).
- GLM backend detection SSOT: `internal/tmux/cg_detect.go` `IsCGMode` (reads
  `team_mode == "cg"`); `.claude/rules/moai/core/glm-web-tooling.md` § CG Mode (the
  `llm.mode` / `llm.team_mode` field disambiguation).
- GLM env injection surface: `internal/cli/glm.go` `setGLMEnv` / `injectGLMEnvForTeam` /
  `buildTmuxClearVars`; `internal/config/defaults.go` `DefaultGLMBaseURL`.
- Adjacent GLM SPEC: SPEC-GLM-MODEL-ALLOWLIST-001 (completed — GLM model allowlist; the
  overlay leaves that allowlist and the `llm.glm.models` mapping untouched).
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Template-First rule + neutrality: CLAUDE.local.md §2/§25, `internal/template/CLAUDE.md`
- Hardcoding rules: CLAUDE.local.md §14 (constants; env names via `envkeys.go`)
- Predecessors: SPEC-CC2178-MODEL-POLICY-REPAIR-001 (map↔file reconciliation),
  SPEC-AGENT-ARCH-V2-001 M3c (No-Haiku 3-tier + `--model-policy` flag),
  SPEC-TOKEN-ROUTING-001 (model_routing closed sets),
  SPEC-WEB-CONSOLE-013 M3 (read-only /model-policy view; §B.4 documents this SPEC's
  sanctioned single-field write-path exception)
- Absorbed deferral: the `agentEffortMap` godoc in `internal/template/model_policy.go`
  defers full effort-map retirement to "SPEC-CC2178-EFFORT-MAP-RETIREMENT-001" (never
  authored). This SPEC's M2 unified apply pass ABSORBS that retirement — M2 removes the
  dual maps and updates/removes the dangling deferral pointer (plan.md §F M2).
- Follow-up: SPEC-MODEL-TIER-ROUTING-PROFILES-001 (descoped 36-cell derivation, D5).
