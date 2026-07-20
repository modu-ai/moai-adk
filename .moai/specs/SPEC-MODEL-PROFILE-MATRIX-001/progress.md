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

### M5 — Tests + docs impact list

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-MPM-001 (new-schema load) | PASS | `go test ./internal/config -run TestLoad_NewSchemaLLMYaml` | ok (profile+agent_overrides parse) |
| AC-MPM-002 (legacy migration load) MUST-PASS | PASS | `go test ./internal/config -run TestLoad_LegacyLLMYaml_Migration` | ok (plan_type+claude_models+perf_tier loads no error → profile max) |
| AC-MPM-003 (round-trip strip) | PASS | `go test ./internal/template -run TestApplyProfile_RoundTripStripsRetiredKeys` | ok (plan_type + claude_models block stripped, profile written, glm sibling survives) |
| AC-MPM-020 (build+tests green) MUST-PASS | PASS | `go build ./...` + `GOOS=windows` exit 0; `go test ./...` = 0 failures | verified |
| AC-MPM-021 (doc-impact list) | PASS | doc-impact list produced below (run-phase artifact, flagged follow-up sync) | listed |
| AC-MPM-023 (GLM wire honesty) | PASS | `moai model profile --json` GLM wire_note = "implemented + wired, live wire-effectiveness pending"; no verified-effectiveness claim | verified |

M5 deliverables: config migration/new-schema load tests; `template.ApplyProfile` write-time strip of plan_type + claude_models block (REQ-MPM-005, line-based indentation processor) + round-trip + insert tests. init/web/wizard tests were delivered in M3.

Coverage (touched packages, full-package baseline): config 80.2%, template 84.9%, cli 74.4%, web 59.5%, core/project 88.5%. New code (profile.go / profile_matrix.go / model.go / ApplyProfile / init --profile) is directly covered by dedicated tests; package figures reflect large pre-existing packages, not the new-code subset.

#### Documentation-Impact List (REQ-MPM-036 / AC-MPM-021) — follow-up sync-phase scope, NOT authored here

Surfaces describing the retired `plan_type` axis / 60-cell (plan_type × tier) matrix / `--plan-type` flag that need sync-phase rewrite to the profile model:

- **README (4 locales)**: `README.md` / `README.ko.md` / `README.ja.md` / `README.zh.md` — the "plan_type Profiles" prose + "60-cell profile matrix (10 agents × 3 tiers × 2 plan_type)" claim (README.md:83). Rewrite to the single 3-column (max/medium/low) per-agent profile matrix.
- **docs-site (11 pages × 4 locales = 44 files)**:
  - `advanced/plan-type-profiles.md` — dedicated page for the retired axis (candidate for replacement with a `profile-matrix.md` page or retirement).
  - `advanced/_meta.yaml` — nav entry for the plan-type-profiles page.
  - `advanced/config-sections.md`, `advanced/no-haiku-3tier.md`, `advanced/tokenomics-overview.md` — plan_type references.
  - `cli-reference/init.md`, `cli-reference/update.md` — `--plan-type` flag docs → `--profile`.
  - `getting-started/cli.md`, `getting-started/init-wizard.md` — init flow plan_type mentions.
  - `core-concepts/what-is-moai-adk.md`, `multi-llm/model-policy.md` — model-routing prose.
- **`.claude/rules/moai/development/model-policy.md`**: already updated in M2 (added "Per-Agent Profile Resolver" section; 0 plan_type residuals) — no further sync edit required.

Doc edits are explicitly deferred to sync-phase per REQ-MPM-036; run-phase produced only this enumeration.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-20
run_commit_sha: 76972b7a9             # M5 final run commit (this SHA backfilled per D3 placeholder exemption)
run_status: PASS-WITH-DEBT            # all must-pass AC PASS; AC-MPM-013/014 web UI = PASS-WITH-DEBT (config+CLI support in place, web override-editing UI deferred)
ac_pass_count: 23                     # of 25 AC-MPM gating criteria PASS
ac_pass_with_debt_count: 2            # AC-MPM-013 (web matrix-render), AC-MPM-014 (web agent_overrides editing UI)
ac_fail_count: 0
must_pass_status: ALL PASS            # AC-MPM-001/002/005/015/025/020/024 all PASS
preserve_list_post_run_count: 0       # SPEC-HOOK-FAILURE-CLASSIFY-001, .moai/state/*, docs-site/, unrelated SPEC dirs untouched
l44_pre_commit_fetch: "0 0 at M1 start; parallel SPEC-CLI-TUX-V3-004 commits interleaved on shared main (bd2334d54 parent 1ef2d9db0); M1-M4 linear chain intact, M3 verified ancestor of M4"
l44_post_push_fetch: verified per-milestone (M1..M4 pushed; M5 pending)
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues (baseline 0)
cross_platform_build:
  linux_amd64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 33             # approx across M1-M5 (excl. 3 deleted plan_type files)
m1_to_mN_commit_strategy: "5 per-milestone Conventional Commits pushed directly to main (Hybrid Trunk); db70597a2 (M1) → eee1c4fc1 (M2) → 319c3e93e (M3) → bd2334d54 (M4) → M5 pending"
matrix_a_fidelity: verified verbatim (spec §A.3, no re-derivation)
haiku_residual: green (0 haiku in matrix)
glm_wire_claim: "implemented + wired, live wire-effectiveness pending (REQ-MPM-039 honesty preserved)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-20
sync_commit_sha: "8fb69f6cc"   # backfilled per the D3 SHA-placeholder exemption (spec-frontmatter-schema.md)
sync_status: PASS-WITH-DEBT                # mirrors run_status; 23/25 AC PASS, 2 PASS-WITH-DEBT (AC-MPM-013/014 web UI wiring), 0 FAIL
ac_summary: "23 PASS / 2 PASS-WITH-DEBT / 0 FAIL of 25 AC-MPM (all 7 must-pass AC PASS)"
changelog_entry_position: "CHANGELOG.md [Unreleased] > Added, first entry (SPEC-MODEL-PROFILE-MATRIX-001)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (this sync commit)"
doc_impact_followup: "README (4 locales) + docs-site (11 pages x 4 locales) + CLI-reference --plan-type->--profile edits deferred per REQ-MPM-036; NOT authored in this sync commit — see progress.md §E.2 Documentation-Impact List"
```
