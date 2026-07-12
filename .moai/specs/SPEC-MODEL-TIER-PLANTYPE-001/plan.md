---
id: SPEC-MODEL-TIER-PLANTYPE-001
title: "Implementation plan — plan_type-aware model tier profiles"
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
tags: "plan, model-policy, plan-type, tier-profiles, glm, effort-overlay"
---

# Plan — SPEC-MODEL-TIER-PLANTYPE-001

## §A Context & Measured Baselines (2026-07-12)

All baselines below were measured on the working tree at plan time. Run-phase MUST
re-verify each anchor before editing (content-token anchors, not line numbers).

### A.1 internal/config/model_routing.go

- `validRoutingModels = {inherit, sonnet, opus, glm}` — `grep -c '"fable"'
  internal/config/model_routing.go` → **0**.
- Two independent validation paths consult the set: `ValidateModelRoutingProfiles`
  (profiles map) and `ValidateModelRouting` (legacy flat map). Both must accept `fable`.
- Stale message literal: the flat-path error text names
  `{inherit, haiku, sonnet, opus, glm}` although `haiku` is NOT in the set (removed under
  No-Haiku). Fix the literal while adding `fable` (REQ-MTP-005).

### A.2 internal/template/model_policy.go

- `agentModelMap` = 5 entries, `[3]string` (high/medium/low): manager-spec,
  manager-develop, manager-docs, manager-git, builder-harness. Auditors + manager-design +
  super-advisor + Explore absent (implicit inherit).
- `agentEffortMap` = 5 entries: manager-spec/plan-auditor/sync-auditor/manager-develop
  xhigh, builder-harness high.
- Divergent precedence today: `ApplyModelPolicy` REPLACES `model:` lines
  (`modelLineRegex.ReplaceAll`); `ApplyEffortPolicy` PRESERVES an existing `effort:` line
  (skips when `effortLineRegex.Match`), inserting only when absent.
- `ModelAliasTable` ALREADY maps `"fable" → "claude-fable-5"` — no alias work needed.
- `ApplyPerformanceTier` is the llm.yaml persistence precedent (regex line-replace) to
  mirror for `plan_type` persistence (REQ-MTP-016).

### A.3 Shipped agent template files (critical evidence for D1)

- `internal/template/templates/.claude/agents/moai/*.md` = **9 files**; `grep -l
  "^model:"` → **9/9**; `grep -l "^effort:"` → **9/9** (manager-develop's fields sit
  below its long description block — they exist).
- Consequence: under the historical preserve-existing-effort rule, the new effort matrix
  would NEVER land on any shipped agent → 100% inert headline feature. See D1.

### A.4 Call sites

- `grep -c 'ApplyModelPolicy\|ApplyEffortPolicy'` measured (2026-07-12, iter-1 audit
  correction): `internal/core/project/initializer.go` → **2** (the two call lines);
  `internal/cli/update.go` → **3** (two call lines + one comment line "Runs after
  ApplyModelPolicy..."). Total **5 grep-matching lines across 2 files**; the actual
  call sites are 4 (2 per file). All 5 lines go to 0 at M2 (calls replaced by
  `ApplyTierProfile`; the stale comment is rewritten in the same edit).
- The `agentEffortMap` godoc carries a dangling deferral pointer to
  "SPEC-CC2178-EFFORT-MAP-RETIREMENT-001" (never authored; `grep -c
  'EFFORT-MAP-RETIREMENT' internal/template/model_policy.go` → **1**). M2 absorbs the
  retirement and removes/updates the pointer (spec.md §E absorption note).

### A.5 internal/cli — flag + wizard

- `--model-policy` canonical enum is ALREADY `{max, medium, low}` (init.go flag help +
  `validateInitFlags`); deprecated boolean aliases `--high`/`--medium-alias`/`--low` map
  to max/medium/low via `resolveModelPolicy`.
- Wizard: `internal/cli/wizard/wizard.go` has `case "model_policy"` → `result.ModelPolicy`
  (`types.go` comments the legacy high/medium/low vocabulary); init.go copies
  `result.ModelPolicy` → `opts.ModelPolicy` → `template.ApplyModelPolicy`.
- NOTE the pre-existing dual vocabulary: CLI/tier = {max,medium,low} vs
  `template.ModelPolicy` = {high,medium,low}. REQ-MTP-012's mapping function is the
  bridge; REQ-MTP-015 keeps the CLI contract as-is.

### A.6 internal/config/types.go — LLMConfig

- No plan-type field exists. `PerformanceTier` carries
  `validate:"omitempty,oneof=high medium low"` while the CLI persists max/medium/low —
  a pre-existing tag/value drift (see D4, adjacent one-line fix).

### A.7 internal/web — /model-policy view

- Route registered: `internal/web/app.go` → `mux.HandleFunc("/model-policy",
  a.handleModelPolicy)` (baseline 1 occurrence — reachability anchor). The new persist
  endpoint path `"/model-policy/plan-type"` has baseline **0** occurrences in app.go
  (registration reachability anchor for AC-MTP-027a).
- View is READ-ONLY today: non-GET → 405; no FieldDef/persist path
  (SPEC-WEB-CONSOLE-013 REQ-WC13-020/021). NOTE (iter-1 audit correction): the literal
  token `PersistTarget` DOES appear once in `internal/web/modelpolicy.go` — inside the
  file-header comment ("NO PersistTarget") — so the token-absence grep baseline is
  **1**, not 0; the scoped-write invariant is therefore pinned by behavior tests
  (AC-MTP-021), not by a token-absence grep. Current page renders llm.performance_tier
  + the model_routing_profiles 3×12 table.
- i18n: `mp.*` keys present ×4 locales (`grep -c '"mp.title"'
  internal/web/assets/i18n.js` → 4).
- `.templ` → `_templ.go` generation must stay in sync (`modelpolicy.templ` /
  `modelpolicy_templ.go` both committed).

### A.8 YAML surfaces

- Local + template `llm.yaml`: no `plan_type` key (template default
  `performance_tier: "medium"`).
- `workflow.yaml` closed-set comments name `{inherit, sonnet, opus, glm}` — add `fable`
  mention (local + template twins, REQ-MTP-024).

### A.9 Quality baselines (plan-phase capture, per quality gates)

- `go vet ./internal/config/... ./internal/template/... ./internal/web/...
  ./internal/cli/... ./internal/core/project/...` → exit **0** (clean).
- `moai spec lint` repo-global: **29 errors / 28 warnings**, all pre-existing in OTHER
  SPECs (lint exit is repo-global; the delta attributable to this SPEC must be 0 errors).
- Run-phase pre-flight MUST additionally capture `golangci-lint run` on the touched
  packages before the first edit (LSP/lint baseline per run-phase gate).

### A.10 GLM backend + effort-injection surface (M5 evidence, measured 2026-07-12)

- **Detection fields (CORRECTED iter-3)** — the ACTUAL persisted GLM signal is
  `internal/config/types.go` `LLMConfig.TeamMode` (yaml `llm.team_mode`), NOT `LLMConfig.Mode`.
  `persistTeamMode` (glm.go:508, body `llmCfg.TeamMode = mode` at glm.go:519) is the sole
  writer: `moai glm` calls `persistTeamMode(root, "glm")` (glm.go:334, launcher.go:138) →
  `team_mode = "glm"`; `moai cg` calls `persistTeamMode(root, "cg")` (glm.go:300,
  launcher.go:229) → `team_mode = "cg"`; `moai cc` → `""`. The `LLMConfig.Mode` field (yaml
  `llm.mode`) is **DORMANT** — verified 2026-07-12: `grep` for a non-test writer of
  `llmCfg.Mode` / `LLM.Mode =` → 0; non-test reader of `LLM.Mode == "glm"` → 0. So `moai glm`
  does NOT set `mode="glm"` (an earlier draft of this plan claimed it did — that was the iter-3
  bounded defect). The `LLMConfig.TeamMode` godoc comment (`"", "claude", "glm", "hybrid"`) is
  also STALE — the values `persistTeamMode` actually writes are `"cg"` / `"glm"` / `""`.
  Adjacent M5 notes: fix the stale godoc AND add a "dormant field" note to `LLMConfig.Mode`
  (both non-blocking; surfaced here).
- **Runtime detector** — `internal/tmux/cg_detect.go` `IsCGMode` reads `team_mode == "cg"`
  (layered OR: llm.yaml `team_mode` + tmux SESSION GLM markers, with a preserved
  `teammateMode=tmux` + process-env fallback). The M5 predicate (REQ-MTP-026) is the
  CONFIG-level intent signal — `team_mode ∈ {"cg", "glm"}` (the real signal) OR the defensive
  `mode == "glm"` (dormant field); it does NOT re-implement `IsCGMode`, it cross-references it.
  CRITICAL: the predicate MUST include `team_mode == "glm"` — omitting it (keying only off
  `mode=="glm"` OR `team_mode=="cg"`) leaves the primary all-GLM `moai glm` session undetected
  and the overlay inert (the iter-3 defect).
- **Effort constants** — `internal/template/model_policy.go` already defines
  `EffortLevelLow/Medium/High/XHigh/Max` (`low`/`medium`/`high`/`xhigh`/`max`). Collapse INPUT.
- **Injection points** — `internal/cli/glm.go` `setGLMEnv(glmConfig, apiKey)` (direct
  `moai glm`, `os.Setenv` at ~L193) and `injectGLMEnvForTeam(settingsPath, glmConfig, apiKey)`
  (`moai cg`, `.claude/settings.local.json` `env` block at ~L584). Both write
  `ANTHROPIC_BASE_URL` (= `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`) +
  `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` (the tier→GLM model mapping). Neither
  writes any reasoning/effort control: `grep -c 'reasoning_effort' internal/cli/glm.go` → 0;
  `grep -c 'thinking' internal/cli/glm.go` → 0.
- **Teardown parity** — `buildTmuxClearVars()` (~L485) is the `moai cc` cleanup list
  (REQ-CGH-009 inject↔clear key-parity); any Branch-B reasoning key must be added there.
- **Repo-wide** — `reasoning_effort` (any casing) and the GLM `thinking`-toggle concept
  appear in ZERO non-test Go files (baseline 0 for both — reachability anchors for M5 ACs).
- **Collapse function / override set** — proposed home `internal/template` (near the effort
  constants; `internal/cli` → `internal/template` import is already established by
  `ApplyModelPolicy`). Grep for any collapse-fn / override-set name → baseline 0.

## §B Decision Points & Clarifications

All four original NEEDS CLARIFICATION markers were resolved by user decisions at
Implementation Kickoff (2026-07-12, plan-audit iteration 1). The v0.3.0 GLM effort overlay
adds decision point D7 (injection mechanism), also RESOLVED as a deliberate run-phase
empirical gate (the architecture is confirmed; only the A/B delivery branch is left for
run-phase measurement — this is an intentional run-phase gate, NOT an open plan-phase
clarification). **0 residual NEEDS CLARIFICATION markers.**

### D1 — ApplyTierProfile precedence (effort dimension reachability) — RESOLVED

**RESOLUTION (user, kickoff 2026-07-12): replace-both confirmed** — the recommended
default. REQ-MTP-011 stands as written.

- The delegation brief provisionally proposed keeping the legacy split "unless spec
  review finds otherwise". Spec review FOUND OTHERWISE: §A.3 measures 9/9 shipped agent
  files carrying `effort:` frontmatter, so preserve-existing-effort makes the entire
  plan_type × tier effort matrix unreachable on every shipped agent (inert-headline
  hazard — same failure class as the reachability lesson from the project-harness bridge
  work).
- Encoded in REQ-MTP-011: replace BOTH. The tier profile becomes the SSOT for
  shipped-agent model/effort; users who want to pin a value re-pin after init/update
  (identical to today's `model:` semantics, which already replace).
- Alternative considered: replace-both-only-when-current-value-matches-template-baseline
  (preserves genuine user edits) — rejected as disproportionate complexity for Tier M;
  may be revisited as follow-up.

### D2 — Web plan_type selector interactivity — RESOLVED (scope change)

**RESOLUTION (user, kickoff 2026-07-12): PERSISTING selector** — the user chose the
write option over the recommended read-only preview toggle. This is a deliberate,
user-approved scope change and a sanctioned exception to the SPEC-WEB-CONSOLE-013
REQ-WC13-021 read-only doctrine, bounded to exactly one field.

- Encoding: REQ-MTP-020 (selector persists via the scoped endpoint) + REQ-MTP-021
  (exactly ONE write path — `POST /model-policy/plan-type` — closed-set validation,
  reject leaves llm.yaml byte-identical, no other field writable, page route keeps 405
  on non-GET). WHY note recorded in spec.md §B.4; HISTORY 0.2.0 row documents the
  013-doctrine reconciliation.
- New/updated ACs: AC-MTP-027a (endpoint registration reachability, baseline 0 → 1 in
  app.go), AC-MTP-027b (router-level persistence round-trip), AC-MTP-021 (redesigned
  negative invariant: no write path other than plan_type).

### D3 — moai update plan-type override flag — RESOLVED

**RESOLUTION (user, kickoff 2026-07-12): add the `--plan-type` flag to `moai update`**
— the user chose the flag over the recommended persisted-value-only reading.

- Encoding: REQ-MTP-018 amended — update reads persisted `llm.plan_type` by default AND
  accepts `--plan-type` override which takes precedence for the run and persists the new
  value. AC-MTP-018b added (flag parse + persist + profile application).

### D4 — Adjacent drift: PerformanceTier validate tag (non-marker decision note)

- `LLMConfig.PerformanceTier` tag says `oneof=high medium low`; persisted values are
  max/medium/low. One-line adjacent fix while M1 touches `LLMConfig`. Not a blocking
  clarification — included in M1 (surfaced at kickoff; no objection raised).
- NOT to be confused with D6 (below) — spec.md §B.6/§B.7 matrix notes cite D6 for the
  default-tier decision (iter-1 SHOULD-fix: citation corrected from D4 to D6).

### D5 — M5 retention — RESOLVED (descoped)

**RESOLUTION (user, kickoff 2026-07-12): descope confirmed** — the recommended option.

- M5 and REQ-MTP-025/AC-MTP-025 are removed from this SPEC; the 36-cell derivation
  transfers to follow-up **SPEC-MODEL-TIER-ROUTING-PROFILES-001** (to be authored).
  spec.md §B.5 records the retirement; §C gains the corresponding Out of Scope entry.
  REQ count 25 → 24; milestone count 5 → 4.

### D6 — Default-when-absent tier stays `medium` (resolved design note)

- The default tier when `performance_tier` is absent remains `medium` for BOTH plans
  (preserves the template `performance_tier: "medium"` default; BC-safe). The design's
  ★recommended markers (api→medium, subscription→max) affect only the wizard's
  recommended-option labeling, never the absent-value default. Cited by the spec.md
  §B.6/§B.7 matrix notes.

### D7 — GLM effort-injection mechanism (v0.3.0; RESOLVED as a run-phase empirical gate)

**Architecture RESOLVED (user, 2026-07-12): backend overlay, effort-only.** The GLM effort
overlay is an overlay on the active plan_type profile (NOT a third plan_type), remapping the
per-agent EFFORT value only; the model dimension is already carried under GLM by
`llm.glm.models`. The collapse mapping (REQ-MTP-027) and coding-max override
(REQ-MTP-028) are fixed. What is NOT yet known — and is deliberately left as a **run-phase
empirical gate** — is the DELIVERY mechanism by which the resolved GLM reasoning state
reaches z.ai:

- **UNVERIFIED question**: `moai glm` sets `base_url = https://api.z.ai/api/anthropic` (the
  Anthropic-compat shim). It is unverified whether Claude Code's sub-agent `effort:`
  frontmatter reaches z.ai as `reasoning_effort` / `thinking` through that shim.
- **Branch A (passthrough works)**: the overlay adjusts the effort representation pre-launch
  (the per-agent effort the profile wrote), relying on the shim to translate it. Lowest-cost.
- **Branch B (passthrough does NOT translate)**: MoAI writes `reasoning_effort` and/or the
  `thinking` toggle EXPLICITLY at the GLM launch path — mirroring the existing
  `ANTHROPIC_DEFAULT_*` injection in `setGLMEnv` (env) / `injectGLMEnvForTeam`
  (settings.local.json `env`). If Branch B, the new key(s) MUST be added to
  `buildTmuxClearVars()` (inject↔clear parity, REQ-CGH-009).
- **Delivery-granularity contingency**: if only a session-global reasoning channel is
  available through the shim (no per-agent granularity), the run-phase records that as a
  documented limitation and derives the session `reasoning_effort` from the overlay (a coding
  agent active ⇒ session `reasoning_effort: max` via the override). The per-agent collapse
  LOGIC is still defined and unit-tested regardless of the delivery branch.
- **Why a run-phase gate, not a plan-phase clarification**: determining A vs B requires
  running a GLM session and observing the outbound request to z.ai — an empirical
  measurement, not a user preference. The SPEC states both branches; the run-phase MUST
  determine which holds and MUST NOT claim the overlay is effective (a real reasoning-control
  change reaches z.ai) without that evidence (verification-claim integrity). Recorded in the
  M5 self-verification (§E / AC-MTP-032b).
- **Proposed injection point**: a single collapse+override helper in `internal/template`
  (near the effort constants), CALLED from `internal/cli/glm.go` `setGLMEnv` AND
  `injectGLMEnvForTeam` (or a shared inner helper both invoke). Reachability is pinned by
  AC-MTP-032a (grep the call site in glm.go, baseline 0 → target ≥ 1).

## §C Pre-flight (run-phase entry checklist)

1. Re-verify every §A anchor by content token (grep), not line number.
2. Capture `golangci-lint run` + `go vet` baseline on touched packages (must be clean or
   findings recorded pre-edit).
3. `git diff --cached --stat` empty / no parallel-session overlap on
   `internal/{config,template,cli,web}` + `internal/core/project` (pre-spawn sync check).
4. D1/D2/D3/D5 clarifications: RESOLVED at kickoff 2026-07-12 (see §B) — verify the
   resolutions are reflected in spec.md v0.2.0 before the first edit.
5. Template-First ordering: every template-twin file edited under
   `internal/template/templates/` FIRST, then `make build`, then local sync where
   applicable.

## §D Constraints

- **Template-First + make build**: config/template changes land in
  `internal/template/templates/` first; `make build` re-embeds (no generated
  `embedded.go` — `//go:embed all:templates`).
- **Template neutrality (§25)**: zero internal SPEC IDs / REQ tokens in any file under
  `internal/template/templates/` (CI guard `template-neutrality-check.yaml` +
  `internal_content_leak_test.go`).
- **Hardcoding (§14)**: new closed-set literals become named constants/vars; no inline
  env names (none expected); matrix values live in exactly ONE Go structure
  (REQ-MTP-023 forbids a web-layer duplicate).
- **Language neutrality (§15)**: llm.yaml/workflow.yaml template comments stay
  language-agnostic (they describe the tool, not a project language — low risk).
- **Test isolation**: all apply-pass tests use `t.TempDir()`; no OTEL env vars via
  `t.Setenv` in parallel tests.
- **Matrix fidelity**: §B.6/§B.7 of spec.md are VERBATIM design copies — run-phase must
  not "improve" values; any discrepancy discovered is a blocker report, not an inline
  edit.
- **Existing tests to extend**: `internal/template/model_policy_test.go`,
  `internal/config/model_routing_test.go`, `internal/web/modelpolicy_test.go`,
  `internal/cli` init-flag tests, wizard tests.
- **No time estimates**; milestones are priority-ordered.

## §E Self-Verification (run-phase §E evidence contract)

Each milestone closes with: (E1) per-AC PASS/FAIL matrix rows for its ACs; (E2)
`go build ./...` + `go test ./...` verbatim tails; (E3) coverage for changed packages
(≥85%); (E5) `golangci-lint run` + `go vet` clean on touched packages; templ-generation
sync check (M4); template-neutrality grep (M1/M3); evidence persisted per the
file-redirect contract with citable paths.

## §F Milestones (priority-ordered)

### M1 — Config field + fable enum (Priority: High; blocks all)

- `LLMConfig.PlanType` + closed-set validation + effective-default accessor
  (REQ-MTP-001..003), D4 adjacent tag fix (pending kickoff), template llm.yaml key
  (REQ-MTP-004), `fable` in `validRoutingModels` + both validation paths + message
  literals (REQ-MTP-005), workflow.yaml comment mentions (REQ-MTP-024).
- Tests: `internal/config` loader/validator table tests; `model_routing_test.go` fable
  acceptance both paths.

### M2 — Profile structure + unified apply pass (Priority: High; depends M1)

- Tier-profile structure (REQ-MTP-006) + verbatim matrices (REQ-MTP-007..009) +
  `ApplyTierProfile` with D1 replace-both precedence (REQ-MTP-010..011) + legacy
  ModelPolicy tier-only mapping (REQ-MTP-012) + unknown-agent skip (REQ-MTP-013).
  Replace both call sites; delete the dual maps + `ApplyModelPolicy`/`ApplyEffortPolicy`
  (or reduce to thin deprecated shims ONLY if an unforeseen external consumer surfaces —
  record decision in §E); rewrite the stale "Runs after ApplyModelPolicy" comment in
  update.go; remove/update the dangling "SPEC-CC2178-EFFORT-MAP-RETIREMENT-001" deferral
  pointer in the agentEffortMap godoc (absorption per spec.md §E).
- Tests: 60-cell table-driven matrix fidelity; apply-pass frontmatter assertions in
  `t.TempDir()`; byte-identity for unknown agents; tier-only mapping assertions.

### M3 — CLI flags + wizard + persistence (Priority: High; depends M2)

- init `--plan-type` flag + validation (REQ-MTP-014), `--model-policy` contract
  unchanged (REQ-MTP-015), llm.yaml persistence (REQ-MTP-016), wizard question
  (REQ-MTP-017), update path: persisted-value default + `--plan-type` override flag that
  also persists (REQ-MTP-018, D3).
- Tests: flag validation tables (init + update); wizard case mapping; persistence +
  update-path integration (default read + override persist) in `t.TempDir()`.

### M4 — Web preview + persisting selector (Priority: Medium; depends M2)

- Active plan_type display + default label (REQ-MTP-019), dual-plan preview + persisting
  selector per D2 (REQ-MTP-020), scoped persist endpoint `POST /model-policy/plan-type`
  with closed-set validation + no-other-field-writable invariant + page-route 405
  preserved (REQ-MTP-021), 4-locale i18n keys (REQ-MTP-022), derivation from the Go
  structure (REQ-MTP-023). Register the endpoint in app.go (reachability anchor,
  baseline 0 → 1); regenerate `modelpolicy_templ.go`.
- Tests: handler render assertions (body contains selector + both plan tables +
  structure-derived cell); router-level persistence round-trip (through the real mux);
  out-of-set POST rejected 4xx + llm.yaml byte-identical; persist changes ONLY the
  plan_type line; page-route 405 regression; i18n key-count ×4.

(The FORMER M5 — the 36-cell routing-profile derivation — was removed at v0.2.0, descoped
per D5 to SPEC-MODEL-TIER-ROUTING-PROFILES-001. The M5 below is a NEW, unrelated milestone
added at v0.3.0 for the GLM effort overlay.)

### M5 — GLM backend effort overlay (Priority: Medium; depends M2)

- GLM backend-detection predicate (REQ-MTP-026): `team_mode ∈ {"cg", "glm"}` (the real
  persisted signal) OR the defensive `mode == "glm"` (dormant field) — MUST include
  `team_mode == "glm"` (the primary `moai glm` signal), cross-referencing
  `internal/tmux/cg_detect.go` `IsCGMode` (no re-implementation); fix the stale
  `LLMConfig.TeamMode` godoc + add a dormant-field note to `LLMConfig.Mode` (adjacent, §A.10).
- Collapse function + named constants for the 5→3 mapping (REQ-MTP-027) and the coding-max
  override set constant ({manager-develop, builder-harness}, REQ-MTP-028), in
  `internal/template` near the effort constants. No magic literals (§14).
- Overlay scope: effort-only, GLM-only, on top of the plan_type profile output
  (REQ-MTP-029); non-GLM backend ⇒ identity no-op.
- WIRE the collapse+override helper into `internal/cli/glm.go` `setGLMEnv` AND
  `injectGLMEnvForTeam` (or a shared helper) — reachability, not mere definition
  (REQ-MTP-030). Resolve the injection MECHANISM empirically (D7): determine Branch A
  (passthrough) vs Branch B (explicit `reasoning_effort`/`thinking` write); if Branch B, add
  the key(s) to `buildTmuxClearVars()`. Record the branch decision + the observed evidence
  (outbound z.ai request inspection) in §E (AC-MTP-032b) — do NOT claim overlay effectiveness
  without that evidence.
- Tests: table-driven collapse test (5 Claude levels → GLM state); override test
  (develop/harness → reasoning-max regardless of collapse input, non-override agent →
  collapse result); detection-predicate table test (mode=glm→true, team_mode=cg→true, both
  empty→false); effort-only invariant (model unchanged); non-GLM no-op; reachability grep in
  glm.go. All in `t.TempDir()` where files are touched.

## §G Anti-Patterns (run-phase guardrails)

- Do NOT preserve existing `effort:` "to be safe" — that silently reverts D1 and makes
  the effort matrix inert (the exact hazard this plan documents).
- Do NOT duplicate the matrix as a second literal in the web layer (REQ-MTP-023) or in
  tests (tests may re-declare EXPECTED values — that is the point of a fidelity test —
  but production code has exactly one structure).
- Do NOT edit local `.moai/config/sections/*.yaml` as the primary change — template
  first; local dev config is dogfood and may legitimately differ.
- Do NOT touch `llm.glm`, settings-schema Launch `model_policy`, or agent body prose
  (spec.md §C exclusions).
- Do NOT widen the web write path beyond `llm.plan_type` — the D2 exception is bounded
  to exactly one field + one endpoint; enrolling any other field in a persist path
  violates REQ-MTP-021 and the 013-doctrine reconciliation.
- Do NOT bind plan selection into the legacy ModelPolicy mapping — the mapping is
  tier-only (REQ-MTP-012); plan selection belongs solely to plan_type resolution.
- Do NOT rely on `haiku` anywhere in the new matrices (No-Haiku policy).
- Do NOT claim templ sync without running the generator (verification-claim integrity).
- Do NOT let the GLM overlay touch the per-agent `model:` value or the `llm.glm.models`
  mapping — the overlay is effort-only (REQ-MTP-029); the model→GLM mapping is
  `llm.glm.models`' job.
- Do NOT re-implement `IsCGMode` inside the M5 predicate — read the two config fields and
  cross-reference the SSOT detector (REQ-MTP-026).
- Do NOT inline magic effort/reasoning literals — the 5→3 collapse and the coding-max
  override set are named constants (§14, REQ-MTP-027/028).
- Do NOT DEFINE the collapse function without WIRING it into the glm.go launch path — a
  defined-but-uncalled overlay is inert (the reachability lesson; AC-MTP-032a pins the call
  site, baseline 0 → ≥ 1).
- Do NOT claim the overlay actually changes z.ai reasoning behavior without empirically
  observing the outbound request (D7 / AC-MTP-032b); "the function is called" ≠ "the effort
  reached z.ai as reasoning_effort".
- If Branch B is chosen, do NOT forget the `buildTmuxClearVars()` symmetry — an injected key
  with no teardown leaks GLM reasoning state into a subsequent `moai cc` session.

## §H Cross-References

- spec.md (this SPEC) §B.6/§B.7 — verbatim matrices; §B.8 — GLM overlay REQs;
  acceptance.md — AC gate set.
- Design SSOT: `.moai/reports/model-tier-redesign-20260712.md`.
- GLM reasoning-control research: `research.md` (this SPEC dir) — z.ai thinking/reasoning
  facts + cited sources + collapse-design rationale + shim-injection risk.
- GLM code surface: `internal/cli/glm.go` (`setGLMEnv`, `injectGLMEnvForTeam`,
  `buildTmuxClearVars`), `internal/tmux/cg_detect.go` (`IsCGMode`),
  `internal/config/types.go` (`LLMConfig.Mode` / `LLMConfig.TeamMode`),
  `internal/config/defaults.go` (`DefaultGLMBaseURL`),
  `internal/template/model_policy.go` (`EffortLevel*` constants).
- `.claude/rules/moai/core/glm-web-tooling.md` § CG Mode — `team_mode` detection SSOT +
  field disambiguation.
- `internal/config/CLAUDE.md`, `internal/template/CLAUDE.md` — module conventions.
- CLAUDE.local.md §2 (Template-First), §14 (hardcoding), §15 (neutrality), §25
  (template internal-content isolation).
