# Progress — SPEC-MODEL-PROFILE-MATRIX-001

- **plan_complete_at**: 2026-07-20
- **plan_status**: audit-ready

## §E.1 Plan-phase Audit-Ready Signal

**Plan iteration:** v0.1.1 — plan-audit iter-1 FAIL (0.69, Tier L threshold 0.85) revised for iter-2. Tier L 5-file set now complete: spec.md (40 REQ-MPM), plan.md (M1–M5, D1–D7, 2 DECISIONs), acceptance.md (25 AC-MPM), design.md (architecture §A–§G), research.md (verified investigation §A–§F). Status: draft.

**iter-1 defects addressed:**
- **D1 (effort-injection over-claim):** rewrote REQ-MPM-025/026/027 + AC-MPM-016/017 — the Agent tool accepts per-spawn `model` only, NOT per-spawn `effort` for named subagents; profile injects MODEL at spawn, effort = documented intent (frontmatter default + GLM overlay + Workflow/`Agent(general-purpose)` prompt). Removed the "effort overrides frontmatter at spawn" claim. → DECISION-001.
- **D2 (ApplyTierProfile enumeration + web save-path):** corrected spec §A.4 + plan §B to enumerate all 4 production call sites (`initializer.go:195`, `update.go:486`, `update.go:1447`, `web/agentfm.go:108`); corrected the wrong "web is display-only" premise (the web settings-save path mutates frontmatter today via `applyPerfTierEdits`). Added REQ-MPM-040 + AC-MPM-025 for the web save-path mutation retirement.
- **D3 (Tier L artifacts):** authored design.md + research.md, grounded in the verified investigation (call-site grep, Agent-tool capability, current schema state, GLM overlay mechanics).
- **D4 (default profile discrepancy):** default profile = `medium` (confirmed); reconciled `DefaultModelPolicy = high (→max)` vs config/template `medium` in REQ-MPM-002. Both `[NEEDS CLARIFICATION]` markers resolved into DECISION-001/002 (plan.md §E, dated 2026-07-20) — no open clarification markers remain.

Investigation findings (no agentlint LR-12; statusline not a reader; plan_type UI-selector removed but web save-path still mutates; ApplyTierProfile 4 call sites; template source `model: inherit` vs mutated local `model: opus`; `moai model profile` accessor does not exist yet) are recorded in research.md and plan.md §B and shape scope.

## §F Phase 4 Mode Selection

- Inputs: tier=L, scope ~20-25 files (Go config/template/cli/web + template llm.yaml + rules docs), domains=2 (Go source + template/docs), language mix Go-dominant, concurrency benefit LOW (coding-heavy).
- Mode evaluation: trivial NO (semantic change) / background NO (write work) / agent-team RETIRED / parallel NO (coding-heavy, 2 domains) / workflow NO (not mechanical-uniform) / sub-agent YES.
- Decision: sub-agent (Mode 5, sequential manager-develop, cycle_type=tdd)
- Justification: coding-heavy multi-milestone implementation per Anthropic coding-task parallelism caveat; Plan Audit Gate skipped per 4-condition contract (PASS 0.93 ≥ 0.90, plan artifacts unchanged since c703a203c, verdict < 24h, verdict PASS). Implementation Kickoff Approval: user-pasted resume with explicit `실행: /moai run` directive.

## §E.2 Run-phase Evidence

Run-phase cycle_type=tdd, Mode 5 sequential. Milestones M1–M5, per-M Conventional Commits pushed to main (Hybrid Trunk).

### M1 — Config schema + Matrix A data model

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-MPM-001 (EffectiveProfile) | PASS | `go test ./internal/config -run TestEffectiveProfile` | ok (profile→alias→medium default) |
| AC-MPM-004/007 (agent_overrides validate) | PASS | `go test ./internal/config -run TestValidateAgentOverrides` | ok (non-catalog / out-of-enum model / effort each error) |
| AC-MPM-005 (Matrix A max fidelity) | PASS | `go test ./internal/template -run TestResolveAgentModelEffort_MatrixAFidelity` | ok (all 10 grouped agents = max column) |
| AC-MPM-006 (override precedence) | PASS | `go test ./internal/template -run TestResolveAgentModelEffort_OverridePrecedence` | ok |
| AC-MPM-007 (inherit) | PASS | `TestResolveAgentModelEffort_Inherit` | ok (Explore + user agent → inherit, hasGroup=false) |
| AC-MPM-008 (template+local profiles schema) | PASS | template + local llm.yaml carry profile+profiles+agent_overrides, no plan_type/claude_models | verified by Read |
| AC-MPM-011 (init no plan_type) | PASS | `go test ./internal/cli -run TestInitCmd_PlanTypeRetired` | ok (deployed llm.yaml has no plan_type) |
| AC-MPM-024 (no haiku) | PASS | `TestDefaultProfileMatrix_NoHaiku` | ok |

M1 build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Full `go test ./...` = 0 failures. golangci-lint baseline 0 issues (pre-M1).

New files: internal/config/profile.go, internal/config/profile_test.go, internal/template/profile_matrix.go, internal/template/profile_matrix_test.go. Edited: types.go (+Profile/Profiles/AgentOverrides), validation.go (+validateProfile/validateAgentOverrides), template+local llm.yaml, init_test.go (plan_type test inverted per AC-MPM-011).

### M2 — Runtime resolver + frontmatter reconciliation

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-MPM-016 (resolver max, model-as-arg) | PASS | `go test ./internal/cli -run TestResolveModelProfileReport_MaxClaude` | ok (Matrix A max, Explore→inherit, 11 agents) |
| AC-MPM-019 (GLM overlay) | PASS | `go test ./internal/cli -run TestResolveModelProfileReport_GLMOverlay` | ok (fable→glm-5.2; manager-develop coding-max→reasoning-max; git→thinking-off) |
| AC-MPM-023 (frontmatter inherit) | PASS | 10 local `.claude/agents/moai/*.md` frontmatter == template source (model: inherit + doc-canonical effort) | parity check OK all 10 |
| AC-MPM-025 (resolver reads profile, not frontmatter) | PASS (partial — resolver side) | `go run ./cmd/moai model profile` reads llm.profile matrix | verified smoke (web save-path retirement is M3) |

M2 deliverables: `moai model profile [--json]` read-only resolver (`internal/cli/model.go`); model-policy.md + template mirror gain a "Per-Agent Profile Resolver" section (model = per-spawn runtime arg; effort = documented intent); 8 local agent files restored `model: opus → inherit` (+ manager-develop/manager-design `effort: high → xhigh`) to match template lint-canonical source. REQ-MPM-039 honesty: GLM wire-note "implemented + wired, live wire-effectiveness pending". Full `go test ./...` = 0 failures. catalog.yaml unchanged (rules/local-agents not hashed).

### M3 — Selection surfaces: init wizard + web console

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-MPM-009 (one wizard question, no plan_type) | PASS | wizard has one `model_policy` select; no plan_type question (`grep plan_type internal/cli/wizard/`) | verified |
| AC-MPM-010 (--profile flag persist) | PASS | `go test ./internal/cli -run TestInitCmd_ProfilePersistence` | ok (init --profile max → llm.yaml profile: max, no plan_type) |
| AC-MPM-011 (--plan-type retired) | PASS | `go test ./internal/cli -run TestInitCmd_PlanTypeFlagRetired` | ok (flag not registered; --profile registered) |
| AC-MPM-012 (web profile selector, no plan_type surface) | PASS | ActivePlanType/PlanTypeIsEmpty removed; selector persists llm.profile | verified (schemaform.go seeds EffectiveProfile) |
| AC-MPM-025 (web save no frontmatter mutation) MUST-PASS | PASS | `go test ./internal/web -run TestAgentFMPolicy_ProfilePersistsWithoutFrontmatterMutation` + `grep -rn ApplyTierProfile internal/web --include=*.go \| grep -v _test.go` = empty | ok (frontmatter untouched; 0 production ApplyTierProfile calls in internal/web) |
| AC-MPM-013 (web per-agent resolved render from matrix) | PASS-WITH-DEBT | resolved render available via `moai model profile`; the web agentfm rows still render per-agent frontmatter (a pre-existing SPEC-WEB-CONSOLE-011 surface), not the matrix | debt: web matrix-render swap deferred |
| AC-MPM-014 (web agent_overrides editing) | PASS-WITH-DEBT | config-layer agent_overrides parse/validate/persist implemented (M1 validateAgentOverrides); the existing web agentfm frontmatter-editing surface remains (persists frontmatter, not llm.agent_overrides) | debt: web override-editing → llm.agent_overrides UI not wired |

M3 deliverables: init `--profile <max\|medium\|low>` flag + `template.ApplyProfile` writer (inserts profile: when absent, migration-safe); `--plan-type` flag removed; init persists llm.profile (wizard answer normalized high→max). web: retire the `applyPerfTierEdits` → tier-profile frontmatter re-application (REQ-MPM-040 MUST-PASS); profile selector persists llm.profile + performance_tier alias; ActivePlanType display removed.

**Web debt (honest gap per verification-claim-integrity §3.4)**: the SPEC-WEB-CONSOLE-011 agentfm surface (per-agent frontmatter model/effort editing) is a large pre-existing subsystem. AC-MPM-013 (matrix-render) and AC-MPM-014 (agent_overrides editing UI) require a templ redesign of that surface beyond safe run-phase scope; the config-layer agent_overrides support (M1) + the `moai model profile` matrix render (M2) are in place, so the debt is UI-wiring only. Neither AC is in the §D.1 must-pass set. Full `go test ./...` = 0 failures.

### M4 — Retire plan_type + ApplyTierProfile

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-MPM-018 (no plan_type model/effort resolution) | PASS | `grep -rn "tierProfiles\|ApplyTierProfile\|EffectivePlanType\|GetTierProfileEntry" internal/ --include=*.go \| grep -v _test.go \| grep -v //` | CLEAN (0 production refs) |
| AC-MPM-022 (delegation/GLM env/model_routing_profiles unmodified; no frontmatter effort re-authored) | PASS | delegation.yaml, GLM env code paths, workflow.yaml model_routing_profiles untouched; agent frontmatter restored to template doc-canonical (M2), not re-authored | verified |
| AC-MPM-024 (existing guards green) | PASS | `go test ./...` = 0 failures; HaikuResidualRule + config guards green | verified |

M4 deletions: `internal/config/plan_type.go` (PlanType* constants, IsValidPlanType, ValidPlanTypes, EffectivePlanType, validatePlanType); LLMConfig.PlanType field; template `tierProfiles` (66-cell), `TierProfileEntry`, `tierProfileRow`, `tierColumnIndex`, `GetTierProfileEntry`, `tierProfileAgentOrder`, `TierProfileAgents`, `ApplyTierProfile` (+ modelLineRegex/effortLineRegex/insertEffortInFrontmatter/frontmatterOpenPrefix), `ApplyPlanType`, `ResolveProjectPlanType`, `ApplyGLMEffortOverlay`. All 4 ApplyTierProfile call sites retired (initializer.go, update.go ×2, web/agentfm.go — last was M3). update.go `--plan-type` flag → `--profile`; `applyUpdateTierProfile` → `applyUpdateProfile` (llm.profile persistence, no frontmatter mutation). wizard PlanType field removed. Retired test files deleted: config/plan_type_test.go, cli/update_plantype_test.go. model_policy_test.go rewritten (kept model-alias/deployer/MapModelPolicy/NormalizeToTier/perf-tier tests; dropped tier-profile/plan-type tests). KEPT (design §F): MapModelPolicyToTier/Effort, NormalizeToTier, ResolveProjectPerformanceTier (perf_tier alias axis). Cross-platform build exit 0; lint 0 issues.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
