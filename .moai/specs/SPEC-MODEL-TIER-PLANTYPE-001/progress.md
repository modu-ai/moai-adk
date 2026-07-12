---
id: SPEC-MODEL-TIER-PLANTYPE-001
title: "Progress — plan_type-aware model tier profiles"
version: "0.3.1"
status: completed
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

Run-phase landed as 5 per-milestone commits on `main` (Route A — Hybrid Trunk main-direct,
Tier L; 1-person-OSS policy). Mode 5 sub-agent sequential (each milestone depends on M1/M2
foundations; Template-First `make build` cadence between milestones is inherently sequential).
The `draft → in-progress` transition rode M1 (539aa3e3f).

### Milestone commit SHAs + scope

| Milestone | Commit | Files / lines | Summary |
|-----------|--------|---------------|---------|
| M1 | `539aa3e3f` | 10 files, +340/−9 | `llm.plan_type` config field {api, subscription} + `EffectivePlanType` resolution (absent → subscription) + `"fable"` added to `validRoutingModels` (both `ValidateModelRouting` + `ValidateModelRoutingProfiles` paths) + stale `inherit, haiku` literal fix + template llm.yaml/workflow.yaml twins |
| M2 | `374de51fd` | 4 files, +594/−846 | Unified plan_type × tier profile (60-cell matrix: Plan A rev2 §B.6 + Plan B §B.7 verbatim) + `ApplyTierProfile` (replace-both precedence, D1) replacing `agentModelMap`+`agentEffortMap`; old `ApplyModelPolicy`/`ApplyEffortPolicy` deleted; dangling `EFFORT-MAP-RETIREMENT` deferral pointer absorbed (net −252 lines) |
| M3 | `98db6a9d4` | 11 files, +610/−3 | `moai init --plan-type` flag + closed-set validation + wizard plan-type question (default subscription) + `moai update --plan-type` override flag (reads persisted default; override persists); tier-only legacy `ModelPolicy` mapping preserved |
| M4 | `2dd252ae7` | 8 files, +736/−74 | `moai web` `/model-policy` writable plan_type selector (D2 sanctioned single-field 013 read-only exception); `POST /model-policy/plan-type` registered (scoped write, closed-set validation, page-route 405 preserved); structure-derived preview (no matrix literal duplicate); 4-locale i18n (en/ko/ja/zh); `modelpolicy_templ.go` regenerated |
| M5 | `6e90acbd4` | 7 files, +564/−2 | GLM backend effort overlay: detection predicate `team_mode ∈ {cg, glm}` OR defensive `mode == glm` (cross-refs `IsCGMode`, no re-impl); 5→3 collapse (low→thinking-off; medium/high→reasoning-high; xhigh/max→reasoning-max) as named constants; coding-max override set {manager-develop, builder-harness}; effort-only overlay scope (model dimension untouched); **Branch-B explicit `ANTHROPIC_REASONING_EFFORT` injection** wired at both `setGLMEnv` + `injectGLMEnvForTeam` (shared `glmReasoningEnvVars()` helper) + inject↔clear parity in `buildTmuxClearVars()` |

### Per-AC PASS status (31 AC groups / 43 sub-criteria, acceptance.md SSOT)

| AC group | Milestone | Status | Verification basis |
|----------|-----------|--------|--------------------|
| AC-MTP-001..005b, 024 | M1 | PASS | `go test ./internal/config/...` green; fable-accept both paths; template twin + neutrality greps |
| AC-MTP-006..013 | M2 | PASS | 60-cell matrix-fidelity table test; apply-pass frontmatter assertions in `t.TempDir()`; `ApplyTierProfile` wired at both call sites; dual-map retirement greps |
| AC-MTP-014a..018b | M3 | PASS | init/update `--plan-type` help + out-of-set exit-nonzero; wizard case mapping; llm.yaml persistence round-trip in `t.TempDir()` |
| AC-MTP-019..023b, 027a, 027b | M4 | PASS | handler render + selector + both plan previews; router-level persistence round-trip; scoped-write negative invariants (405/4xx + byte-identity); i18n ×4; templ-sync clean |
| AC-MTP-028..032a | M5 | PASS | detection truth table (7 rows incl. corrected `team_mode="glm"→TRUE`); 5→3 collapse table; coding-max override membership; effort-only + non-GLM no-op; `glm.go` wiring reachability grep |
| AC-MTP-032b | M5 | **PASS-WITH-RESIDUAL** | Branch-B chosen + wired + inject↔clear parity verified in code; **wire-effectiveness EMPIRICALLY UNVERIFIED** — see §E.4 residual |
| AC-MTP-026a..026d | Global | PASS | touched-package suites green (config/template/web/core-project/cli); `moai spec lint` 0 SPEC-attributable errors (delta +0); coverage ≥85% on changed config/template |

Coverage note: `internal/config` and `internal/template` changed packages ≥85% per run-phase §E
capture; touched packages `internal/{config,template,web,cli,core/project}` all green.

## §E.3 Run-phase Audit-Ready Signal

- All 31 AC groups (43 sub-criteria) resolved: 30 clean PASS + 1 PASS-WITH-RESIDUAL
  (AC-MTP-032b — GLM overlay wire-effectiveness, see §E.4).
- Cross-platform build green (`go build ./...` exit 0; re-verified on `main` at sync).
- D1/D2/D3/D5/D7 kickoff resolutions reflected in delivered behavior: replace-both precedence,
  persisting scoped selector, `moai update --plan-type` override, no 36-cell derivation artifacts,
  GLM effort overlay (Branch B).
- Template-First honored: every changed file under `.moai/`/`.claude` template scope carries its
  `internal/template/templates/` twin; `make build` re-embeds; template neutrality greps clean
  (`grep -rn SPEC-MODEL-TIER-PLANTYPE internal/template/templates/` = 0).
- No-Haiku policy preserved (neither matrix contains `haiku`).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-12
sync_commit_sha: pending-backfill-sync
sync_status: PASS-WITH-DEBT
changelog_entry_position: "[Unreleased] › Added (top)"
frontmatter_status_transitions:
  spec_md: "draft → completed (merged in-progress → implemented → completed on sync commit)"
  progress_md: "draft → completed"
b12_self_test_a: "grep -c 'SPEC-MODEL-TIER-PLANTYPE' CHANGELOG.md == 0 before emission (PASS — no duplicate)"
b12_self_test_b: "acceptance.md SSOT = 31 AC groups (43 sub-criteria via a/b suffix); CHANGELOG references 31/43 (PASS — match)"
b12_self_test_c: "file paths in CHANGELOG verified via ls (internal/cli/glm.go, internal/template, internal/web, internal/config present) (PASS)"
canary_compliance_check:
  spec_lint_delta: "0 SPEC-attributable errors (only the pre-close StatusGitConsistency warning, which the completed transition resolves)"
  touched_pkg_tests: "green — internal/{config,template,web,cli,core/project}"
```

### KNOWN RESIDUAL / DEBT (verification-claim-integrity — do NOT overstate completion)

**GLM overlay wire-effectiveness is EMPIRICALLY UNVERIFIED (AC-MTP-032b PASS-WITH-RESIDUAL).**
The M5 GLM effort overlay CODE + WIRING + TESTS are complete: the collapse+override logic is
defined and unit-tested, the overlay is CALLED from both `setGLMEnv` and `injectGLMEnvForTeam`
(shared `glmReasoningEnvVars()` helper), Branch B is implemented as an explicit
`ANTHROPIC_REASONING_EFFORT` env write at the GLM launch path, and inject↔clear parity is present
in `buildTmuxClearVars()` (REQ-CGH-009). HOWEVER, whether z.ai actually CONSUMES that env var
through the Anthropic-compat shim (`base_url = https://api.z.ai/api/anthropic`) — i.e. whether the
resolved reasoning-control change actually reaches z.ai on the wire — is **NOT empirically
determined**. The Branch A (shim passthrough of `ANTHROPIC_REASONING_EFFORT`) vs Branch B
(request-body `reasoning_effort` field required) question needs a live GLM session with
outbound-request inspection to settle. Per AC-MTP-032b + verification-claim-integrity §1.1, **do
NOT claim the overlay is effective on the wire** — the code is wired and unit-tested; the wire
delivery is pending empirical z.ai determination.

**Delivery granularity limitation (research.md §D).** The reasoning-control delivery is
**session-global**, NOT per-agent: the GLM launch path sets one session-level
`ANTHROPIC_REASONING_EFFORT` env value, so the per-agent collapse LOGIC (REQ-MTP-027/028) is
fully defined and unit-tested but the delivery channel available through the shim carries only a
session-global value (the coding-max override implies session `reasoning_effort: max` when a
coding agent is the active spawn). Per-agent reasoning-control delivery, if a channel is later
built, is follow-up work (spec.md §C GLM effort-overlay boundaries).
