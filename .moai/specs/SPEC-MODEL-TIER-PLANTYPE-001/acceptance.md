---
id: SPEC-MODEL-TIER-PLANTYPE-001
title: "Acceptance criteria — plan_type-aware model tier profiles"
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
tags: "acceptance, model-policy, plan-type, tier-profiles, glm, effort-overlay"
---

# Acceptance — SPEC-MODEL-TIER-PLANTYPE-001

All criteria are mechanically verifiable (go test / grep with baseline→target counts /
CLI smoke). Baselines were measured 2026-07-12 and re-verified after plan-audit iter-1
(plan.md §A). Compound multi-token greps are avoided; each grep discriminates one fact.
Reachability criteria pin the registration/consumption site, not just token presence.

v0.2.0 amendment map: AC-MTP-004b/010a/012/021/023a re-pinned per auditor findings;
AC-MTP-018 split into 018a/018b (D3 update flag); AC-MTP-025 retired (D5 descope);
AC-MTP-027a/b added (D2 persisting selector). AC group count: 26 (001–024, 026, 027;
025 retired).

v0.3.0 amendment map: AC-MTP-028..032 added (M5 GLM backend effort overlay — detection
predicate, 5→3 collapse table test, coding-max override table test, effort-only overlay
scope, reachability + run-phase empirical gate). AC group count: 31 (001–024, 026, 027,
028–032; 025 retired). All M5 baselines measured 2026-07-12 (plan.md §A.10): `reasoning_effort`
and the GLM `thinking` toggle appear in ZERO non-test Go files; the collapse-fn / override-set
names appear in ZERO Go files — every M5 grep AC has a discriminating 0 → ≥1 (or 0 → exact)
delta.

v0.3.1 amendment map (plan-audit iter-3 D1/D2 fix): AC-MTP-028 truth table corrected — row
`(mode="", team_mode="glm")` flipped FALSE → **TRUE** (the primary `moai glm` signal;
`persistTeamMode` writes `team_mode="glm"`, and `llm.mode` is dormant — verified on the live
tree), and the genuinely-FALSE cases replaced with `team_mode ∈ {"claude", "hybrid", ""}`;
§D.2 edge case synced. No AC group count change (31 groups unchanged).

## §D AC Matrix

### M1 — Config field + fable enum

- **AC-MTP-001** (REQ-MTP-001) — Go test: the LLM section loader parses
  `plan_type: api` and `plan_type: subscription` into `LLMConfig.PlanType`.
  Check: `go test ./internal/config/ -run 'PlanType'` → PASS (new table test).
- **AC-MTP-002** (REQ-MTP-002) — Go test: absent AND empty `plan_type` both resolve to
  effective plan type `subscription` via the exported accessor. → PASS.
- **AC-MTP-003** (REQ-MTP-003) — Go test: `plan_type: enterprise` yields a validation
  error whose message contains both the offending value and the tokens `api` and
  `subscription`. → PASS.
- **AC-MTP-004a** (REQ-MTP-004) — Template twin:
  `grep -c '^\s*plan_type:' internal/template/templates/.moai/config/sections/llm.yaml`
  — baseline 0 → target 1.
- **AC-MTP-004b** (REQ-MTP-004, neutrality) —
  `grep -rl 'SPEC-MODEL-TIER-PLANTYPE' internal/template/templates/ | wc -l` → **0**
  (template internal-content isolation holds; `-rl | wc -l` form yields a single count —
  measured 0 at plan time and re-verified post-iter-1).
- **AC-MTP-005a** (REQ-MTP-005) —
  `grep -c '"fable"' internal/config/model_routing.go` — baseline 0 → target ≥ 1; AND
  Go tests assert `model: fable` is accepted by BOTH `ValidateModelRouting` (flat) and
  `ValidateModelRoutingProfiles` (profiles):
  `go test ./internal/config/ -run 'ModelRouting'` → PASS with new fable cases.
- **AC-MTP-005b** (REQ-MTP-005, stale literal) —
  `grep -c 'inherit, haiku' internal/config/model_routing.go` — baseline 1 → target 0;
  AND a Go test asserts the invalid-model error message names `fable`.
- **AC-MTP-024** (REQ-MTP-024) —
  `grep -c 'fable' .moai/config/sections/workflow.yaml` — baseline 0 → target ≥ 1; AND
  `grep -c 'fable' internal/template/templates/.moai/config/sections/workflow.yaml`
  — baseline 0 → target ≥ 1 (closed-set comment mentions, both twins).

### M2 — Profile structure + unified apply pass

- **AC-MTP-006** (REQ-MTP-006..009) — Matrix fidelity: a table-driven Go test asserts,
  for EVERY (plan ∈ {api, subscription}) × (tier ∈ {max, medium, low}) × (agent ∈ 10),
  that the profile lookup returns exactly the {model, effort} pair of spec.md §B.6/§B.7
  — 60 asserted cells, including the rev2 cells (api/max manager-develop = fable/high;
  api/medium manager-develop = opus/high; api/low manager-develop = opus/medium).
  → PASS.
- **AC-MTP-007** (REQ-MTP-006 retirement) —
  `grep -c 'agentModelMap = map\[string\]\[3\]string' internal/template/model_policy.go`
  — baseline 1 → target 0 (the `[3]string` dual-map representation is gone).
- **AC-MTP-008** (REQ-MTP-010/011, api branch) — Apply-pass test in `t.TempDir()` with
  the 9 shipped agent fixtures: after applying (api, medium),
  `manager-develop.md` frontmatter contains `model: opus` AND `effort: high`;
  `plan-auditor.md` contains `model: fable` AND `effort: high` (moved off inherit).
  → PASS.
- **AC-MTP-009** (REQ-MTP-010/011, subscription branch) — Same harness, (subscription,
  max): `manager-develop.md` → `model: sonnet` + `effort: high`; `manager-spec.md` →
  `model: opus` + `effort: high`; `manager-docs.md` → `model: sonnet` + `effort: low`.
  → PASS.
- **AC-MTP-010a** (REQ-MTP-010, call-site replacement) —
  `grep -c 'ApplyModelPolicy\|ApplyEffortPolicy' internal/core/project/initializer.go`
  — baseline 2 → target 0; same grep on `internal/cli/update.go` — baseline **3**
  (two call lines + the "Runs after ApplyModelPolicy" comment line; iter-1 audit
  correction) → target 0 (the stale comment is rewritten at M2 alongside the call
  replacement).
- **AC-MTP-010b** (REQ-MTP-010, reachability) —
  `grep -c 'ApplyTierProfile' internal/core/project/initializer.go` → ≥ 1 AND
  `grep -c 'ApplyTierProfile' internal/cli/update.go` → ≥ 1 (the new pass is WIRED at
  both consumers, not merely defined).
- **AC-MTP-011** (REQ-MTP-011, precedence) — Go test: an agent fixture pre-seeded with
  `model: sonnet` + `effort: xhigh` is rewritten to the profile's pair under the D1
  precedence (both lines replaced); AND the precedence rule is stated in godoc:
  `grep -c 'precedence' internal/template/model_policy.go` → ≥ 1. → PASS.
  (D1 RESOLVED at kickoff 2026-07-12: replace-both confirmed — the fixture outcome
  above is final.)
- **AC-MTP-012** (REQ-MTP-012, tier-only) — Go test: the legacy mapping function
  translates the TIER ONLY — high→`max`, medium→`medium`, low→`low` — and its return
  signature carries NO plan value (plan selection is asserted separately via the
  effective plan_type resolution tests, AC-MTP-002/016/018); `template.IsValidModelPolicy`
  remains exported and passing its existing tests. → PASS.
- **AC-MTP-013** (REQ-MTP-013) — Go test: a file `custom-agent.md` (not in the profile)
  is byte-identical after the apply pass. → PASS.

### M3 — CLI flag + wizard + persistence

- **AC-MTP-014a** (REQ-MTP-014) — `go run ./cmd/moai init --help 2>&1 | grep -c
  'plan-type'` → ≥ 1.
- **AC-MTP-014b** (REQ-MTP-014) — Go test (or CLI smoke in `t.TempDir()`):
  `moai init --plan-type bogus` exits non-zero and stderr names `api` and
  `subscription`. → PASS.
- **AC-MTP-015** (REQ-MTP-015) — Existing `--model-policy` validation tests still PASS
  unmodified in their assertions (max/medium/low accepted; legacy bool aliases resolve);
  `go test ./internal/cli/ -run 'InitFlags|ModelPolicy'` → PASS.
- **AC-MTP-016** (REQ-MTP-016) — Go test: init with `--plan-type api` in `t.TempDir()`
  leaves `.moai/config/sections/llm.yaml` containing `plan_type: api` (line-grep in
  test). → PASS.
- **AC-MTP-017** (REQ-MTP-017) —
  `grep -c 'case "plan_type"' internal/cli/wizard/wizard.go` — baseline 0 → target 1;
  AND a wizard unit test maps the answer into the result struct; AND the wizard default
  option is subscription. → PASS.
- **AC-MTP-018a** (REQ-MTP-018, persisted default) — Go test: a project whose `llm.yaml`
  carries `plan_type: api` gets the api-branch profile re-applied by the update path
  with NO flag given (agent frontmatter assertion equals an api cell, not a
  subscription cell). → PASS.
- **AC-MTP-018b** (REQ-MTP-018, D3 override flag) — (parse) `go run ./cmd/moai update
  --help 2>&1 | grep -c 'plan-type'` → ≥ 1, and an out-of-set value exits non-zero
  naming `api` and `subscription`; (persist + apply) Go test: a project persisted with
  `plan_type: subscription` updated via `--plan-type api` ends with `llm.yaml`
  containing `plan_type: api` AND agent frontmatter matching an api cell. → PASS.

### M4 — Web preview

- **AC-MTP-019** (REQ-MTP-019) — Handler test: GET `/model-policy` → 200; with
  `plan_type: api` configured the body contains the active value `api`; with the field
  absent the body contains the default-subscription label. → PASS.
- **AC-MTP-020** (REQ-MTP-020) — Handler test: the body contains the plan selector
  markup AND both plan previews (api + subscription), each rendering 10 agent rows × 3
  tiers; at least one asserted cell value is READ FROM the Go profile structure in the
  test (cross-check, not a hardcoded duplicate of the matrix — this cross-check also
  carries the single-source guarantee alongside AC-MTP-023a). → PASS.
- **AC-MTP-021** (REQ-MTP-021, scoped-write negative invariant — redesigned per D2 +
  iter-1 baseline correction; the former token-absence grep is DROPPED: `PersistTarget`
  already appears once in the modelpolicy.go header comment, baseline 1, and D2 now
  legitimately adds a write path) — behavior tests:
  (a) page route: non-GET on `/model-policy` → 405 (existing test still PASS);
  (b) out-of-set POST to `/model-policy/plan-type` (e.g., `enterprise`) → 4xx AND
  `llm.yaml` byte-identical (pre/post file-content comparison in test);
  (c) a successful persist changes ONLY the `plan_type` line — the test asserts the
  rest of `llm.yaml` is byte-identical pre/post;
  (d) no other endpoint under `/model-policy` accepts a write — POST to
  `/model-policy` itself and to any non-registered sub-path returns 405/404 with
  `llm.yaml` unchanged. → PASS.
- **AC-MTP-022** (REQ-MTP-022) — For EACH new i18n key introduced (key list finalized in
  run-phase, prefix `mp.plan`): `grep -c '"<key>"' internal/web/assets/i18n.js` → exactly
  4 (one per locale, en/ko/ja/zh). Spot-anchor: `grep -c '"mp.plan.title"'
  internal/web/assets/i18n.js` — baseline 0 → target 4.
- **AC-MTP-023a** (REQ-MTP-023, single source — command rewritten per iter-1 SHOULD-fix;
  the original bare-directory `grep -c` form was broken: it emitted 57 per-file `:0`
  lines and scanned `_test.go`) —
  `grep -rE 'fable ?/ ?high' internal/web --include='*.go' --include='*.templ' | grep -v '_test\.go:' | wc -l`
  → **0** (baseline 0, stays 0) — production web code carries NO literal matrix cells.
  The positive single-source guarantee is carried by AC-MTP-020's structure-derived
  cross-check; this grep is the complementary literal-absence spot-check.
- **AC-MTP-023b** (reachability + generation sync) —
  `grep -c '"/model-policy"' internal/web/app.go` → 1 (page-route registration
  unchanged; the persist endpoint has its own anchor in AC-MTP-027a);
  AND `templ generate` (or the repo's make target) followed by `git status --porcelain
  internal/web/` shows no drift between `modelpolicy.templ` and `modelpolicy_templ.go`.
- **AC-MTP-027a** (REQ-MTP-021, endpoint registration reachability — new per D2) —
  `grep -c '"/model-policy/plan-type"' internal/web/app.go` — baseline **0** → target
  **1** (the persist endpoint is REGISTERED on the mux, not merely defined in a
  handler file).
- **AC-MTP-027b** (REQ-MTP-020/021, persistence round-trip — new per D2) — Router-level
  Go test (through the real app mux, not a direct handler call): POST a valid plan-type
  change (`subscription` → `api`) to `/model-policy/plan-type` → 2xx/3xx; `llm.yaml`
  now contains `plan_type: api`; a subsequent GET `/model-policy` renders `api` as the
  active plan. → PASS.

### M5 — GLM backend effort overlay (v0.3.0)

- **AC-MTP-028** (REQ-MTP-026, backend detection — truth table corrected iter-3) —
  Table-driven Go test on the detection predicate, using the ACTUAL persisted values
  (`persistTeamMode` writes `team_mode` = `"glm"` for `moai glm`, `"cg"` for `moai cg`;
  `mode` is dormant):
  - `(mode: "", team_mode: "glm")` → **TRUE** — the PRIMARY all-GLM `moai glm` signal (this is
    the row the iter-3 defect had wrong; keying only off `mode=="glm"` OR `team_mode=="cg"`
    would leave this inert);
  - `(mode: "", team_mode: "cg")` → TRUE — the `moai cg` signal;
  - `(mode: "glm", team_mode: "")` → TRUE — the defensive dormant-`mode` OR;
  - `(mode: "glm", team_mode: "cg")` → TRUE;
  - `(mode: "", team_mode: "claude")` → FALSE — legacy non-GLM `team_mode` value;
  - `(mode: "", team_mode: "hybrid")` → FALSE — legacy non-GLM `team_mode` value;
  - `(mode: "", team_mode: "")` → FALSE — no signal;
  AND plan_type value (`api` vs `subscription`) has NO effect on the predicate result. → PASS.
  Reachability: `grep -c 'cg_detect\|IsCGMode\|internal/tmux' internal/template/*.go` on the
  new predicate file → ≥ 1 (the predicate cross-references the `IsCGMode` SSOT rather than
  re-implementing it).
- **AC-MTP-029** (REQ-MTP-027, collapse mapping) — Table-driven Go test asserting the collapse
  of all 5 Claude effort levels to the GLM canonical state:
  `low` → thinking-disabled; `medium` → reasoning-high; `high` → reasoning-high;
  `xhigh` → reasoning-max; `max` → reasoning-max. → PASS. Reachability + no-magic-literal:
  the collapse function + its GLM-state / `reasoning_effort` / `thinking` value tokens are
  named constants — `grep -rc '<collapseFnName>' internal/template/` → baseline 0 → ≥ 1;
  `grep -c 'reasoning_effort\|thinking' internal/template/<glm-overlay-file>.go` → ≥ 1
  (the value tokens are declared as constants, not inline in the mapping switch).
- **AC-MTP-030** (REQ-MTP-028, coding-max override) — Table-driven Go test: for the override
  set, `manager-develop` and `builder-harness` resolve to `reasoning-max` under a GLM backend
  REGARDLESS of the input effort (e.g. an input of `low` — which would collapse to
  thinking-disabled — still yields `reasoning-max` for these two agents); a NON-override agent
  with input `low` yields thinking-disabled (collapse result, un-overridden). → PASS. The
  override set is a named constant of EXACTLY {`manager-develop`, `builder-harness`}:
  `grep -c '<overrideSetName>' internal/template/<glm-overlay-file>.go` → baseline 0 → ≥ 1,
  and a test asserts the set membership is exactly those two agents (no third member).
- **AC-MTP-031** (REQ-MTP-029, effort-only + GLM-only scope) — Go test: (a) applying the
  overlay to a profile-produced {model, effort} pair under a GLM backend changes ONLY the
  effort representation — the `model` value is byte-identical to the plan_type profile output
  (the overlay never rewrites `model:`); (b) under a NON-GLM backend (predicate FALSE) the
  overlay is an identity no-op — the {model, effort} pair equals the plan_type profile output
  unchanged. → PASS.
- **AC-MTP-032a** (REQ-MTP-030, wiring reachability) — The collapse+override helper is CALLED
  from the GLM launch path, not merely defined:
  `grep -c '<collapseFnName>\|<overlayHelperName>' internal/cli/glm.go` — baseline **0** →
  target ≥ 1, with the call reachable from BOTH `setGLMEnv` and `injectGLMEnvForTeam` (a
  shared inner helper invoked by both satisfies this). A defined-but-uncalled overlay FAILS
  this AC (the reachability lesson — inert-headline hazard).
- **AC-MTP-032b** (REQ-MTP-030, injection mechanism run-phase empirical gate) — The run-phase
  MUST record, in progress.md §E.2/§E.3, the empirical determination of the delivery branch:
  Branch A (shim passthrough forwards effort→`reasoning_effort`) vs Branch B (explicit
  `reasoning_effort`/`thinking` write), WITH observed evidence (the outbound z.ai request or
  an equivalent instrumentation showing the reasoning control that actually reached the shim).
  A claim that "the overlay is effective" (a real reasoning-control change reaches z.ai)
  without that evidence FAILS this AC (verification-claim integrity). **When** Branch B is
  chosen: `grep -c 'reasoning_effort\|thinking' internal/cli/glm.go` → ≥ 1 AND the injected
  key(s) appear in `buildTmuxClearVars()` (`grep` the clear list, ≥ 1 — inject↔clear parity).
  **When** Branch A is chosen: the pre-launch effort adjustment is asserted by an
  `injectGLMEnv`/`setGLMEnv`-path test, and the Branch-B grep targets are documented as
  intentionally not-applicable (recorded, not silently absent).

### (Former M5) — RETIRED (descoped per plan.md D5, kickoff 2026-07-12)

> This is the ORIGINAL M5 (36-cell routing-profile derivation), retired at v0.2.0. It is
> distinct from the NEW M5 (GLM backend effort overlay) above, added at v0.3.0.

- **AC-MTP-025** — RETIRED. The 36-cell derivation AC transfers with REQ-MTP-025 to
  follow-up SPEC-MODEL-TIER-ROUTING-PROFILES-001. This ID is struck from the gate set
  and MUST NOT be reused within this SPEC. (Recorded per the original descope
  contingency; progress.md §E.1 carries the descope note.)

### Global gates

- **AC-MTP-026a** — Full suite: `go test ./...` → PASS (exit 0).
- **AC-MTP-026b** — `go vet` on the 5 touched package trees → exit 0 (baseline 0 → 0);
  `golangci-lint run` on changed packages → 0 new findings vs the run-phase pre-flight
  baseline.
- **AC-MTP-026c** — Coverage: `go test -cover` for `internal/config` and
  `internal/template` changed packages ≥ 85% (verbatim output cited).
- **AC-MTP-026d** — `moai spec lint`: 0 errors attributable to
  SPEC-MODEL-TIER-PLANTYPE-001 files (repo-global baseline 29 pre-existing errors /
  28 warnings in other SPECs; delta from this SPEC = +0 errors).

## §D.1 Given-When-Then scenarios

### Scenario 1 — API-metered project init (happy path, rev2 encoding)

- **Given** a fresh directory and no existing `.moai/` config
- **When** the user runs `moai init --plan-type api --model-policy medium`
- **Then** `.moai/config/sections/llm.yaml` contains `plan_type: api`, and the deployed
  `manager-develop.md` frontmatter reads `model: opus` + `effort: high` (A-medium rev2),
  and `manager-spec.md` reads `model: fable` + `effort: high`, and `manager-git.md`
  reads `model: sonnet` + `effort: low`.

### Scenario 2 — Legacy project untouched-config upgrade (backward compat)

- **Given** an existing project whose `llm.yaml` has NO `plan_type` key
- **When** `moai update` re-applies the tier profile with persisted tier `max`
- **Then** the subscription branch is selected (B-max): `manager-develop.md` reads
  `model: sonnet` + `effort: high` and `super-advisor.md` reads `model: opus` +
  `effort: xhigh`; no error or warning about the missing key is raised.

### Scenario 3 — Invalid plan type rejected

- **Given** any directory
- **When** the user runs `moai init --plan-type enterprise`
- **Then** the command exits non-zero before any file is written, and stderr names the
  closed set {api, subscription}.

### Scenario 4 — Web selector persists plan type (D2)

- **Given** a project with `plan_type: subscription` and the web console running
- **When** the user opens GET `/model-policy` and submits `api` through the plan_type
  selector
- **Then** `llm.yaml` now contains `plan_type: api` with every other line byte-identical,
  the re-rendered page shows `api` as the active plan (subscription preview still
  offered), a POST of `enterprise` to the persist endpoint is rejected 4xx with
  `llm.yaml` unchanged, and a POST to `/model-policy` (the page route) still returns 405.

### Scenario 5 — GLM backend collapses effort (v0.3.0 overlay)

- **Given** a project on the `api` plan_type whose profile assigns `super-advisor` = fable /
  xhigh and `manager-develop` = opus / high (A-medium), and the session is GLM-backed
  (`llm.team_mode: cg` written by `moai cg`)
- **When** the GLM effort overlay resolves each agent's GLM reasoning state
- **Then** `super-advisor` collapses to `reasoning_effort: max` (xhigh → reasoning-max),
  `manager-develop` resolves to `reasoning_effort: max` (the coding-max override, NOT its
  collapse result of reasoning-high), a `manager-git` agent at effort `low` collapses to
  `thinking: disabled`, and NONE of the three agents' `model:` value is altered by the overlay
  (the model→GLM mapping stays with `llm.glm.models`); AND on a non-GLM (Claude) session of the
  same project the overlay is an identity no-op (effort values equal the plan_type profile
  output unchanged).

## §D.2 Edge cases

- `plan_type: API` (wrong case) → validation error (closed set is lowercase-exact) —
  covered by AC-MTP-003 variant row.
- Agent file with malformed frontmatter (no closing `---`) → apply pass leaves it
  untouched and does not error the whole run (existing `insertEffortInFrontmatter`
  guard semantics carried over).
- `--plan-type` given while wizard also answers → flag wins (REQ-MTP-017 precedence) —
  asserted in the M3 wizard test.
- Explore: no `Explore.md` file exists — apply pass emits no error and writes nothing
  for the Explore row.
- Empty `plan_type: ""` → identical to absent (subscription) — covered by AC-MTP-002.
- GLM overlay under `llm.team_mode: "glm"` → predicate **TRUE** — this IS the primary `moai glm`
  all-GLM signal (`persistTeamMode` writes `team_mode="glm"`); the genuinely-FALSE cases are
  `team_mode ∈ {"claude", "hybrid", ""}` with `mode ≠ "glm"` — covered by AC-MTP-028.
- GLM overlay applied to an agent whose effort string is unrecognized (not one of the 5 Claude
  levels) → collapse maps it to the documented GLM default state, no panic — covered by the
  collapse totality clause in REQ-MTP-027 (asserted in the AC-MTP-029 test table).
- `Explore` under GLM: its profile row is `inherit / <effort>`; the overlay collapses the
  effort like any other agent, but there is no `Explore.md` file to write and `inherit` is not
  in the override set — no error (consistent with the §D.2 Explore no-file edge case above).

## §D.3 Definition of Done

- All ACs in the v0.3.0 gate set (AC-MTP-001..024, 026..032, sub-criteria included) PASS with
  verbatim evidence cited (per-AC matrix in progress.md §E.2/§E.3). AC-MTP-025 is retired and
  outside the gate set.
- D1/D2/D3/D5 resolutions (kickoff 2026-07-12, plan.md §B) reflected in the delivered
  behavior: replace-both precedence, persisting selector with scoped write path, update
  `--plan-type` override, no 36-cell derivation artifacts.
- D7 (GLM effort overlay, v0.3.0) delivered: backend-conditional effort-only overlay with the
  5→3 collapse + coding-max override, WIRED into the glm.go launch path (AC-MTP-032a), and the
  injection MECHANISM empirically determined (Branch A vs B) WITH observed z.ai evidence
  recorded — no overlay-effectiveness claim without that evidence (AC-MTP-032b).
- Template-First verified: every changed file under `.claude/`/`.moai/` template scope
  has its `internal/template/templates/` twin edited in the same milestone commit, and
  `make build` succeeds.
- No new lint/vet findings; full `go test ./...` green; coverage gate met.
