# SPEC-V3R3-HYBRID-001 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                                  |
|---------|------------|-----------------------------------|------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B)     | 최초 작성 — 36 tasks (T-HYBRID-01..36) across M1-M5 (RED → GREEN x3 → REFACTOR) |

---

## Task ID Convention

- ID format: `T-HYBRID-NN` (zero-padded 2 자리)
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd` (TDD gate verification), `manager-docs` (documentation), `expert-backend` (Go 구현 + test), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M4 (GREEN) → M5 (REFACTOR + Trackable)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase) — Priority P0

Goal: Create 4 audit/integration test files that fail until M2-M4 implementation lands. Per `spec-workflow.md` TDD: write failing test first.

Reference template: `internal/template/lang_boundary_audit_test.go` (SPEC-V3R2-WF-005 M1 패턴), `internal/cli/glm_test.go` (existing test 패턴).

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-01 | Create `internal/template/provider_allowlist_audit_test.go` skeleton (package, imports, embedded FS access). Sentinel string definition (`PROVIDER_ALLOWLIST_VIOLATION`). | expert-backend | `internal/template/provider_allowlist_audit_test.go` (NEW, ~20 LOC scaffold) | none | 1 file (create) | RED setup |
| T-HYBRID-02 | Implement `TestProviderAllowlist` walking `internal/llm/providers/` directory in embedded FS, asserting exactly 4 `.go` files (excluding `_test.go`) named `glm.go`, `kimi.go`, `deepseek.go`, `qwen.go`. On 5th file: `t.Errorf("PROVIDER_ALLOWLIST_VIOLATION: %s exists; v3R3 allow-list closed at four", path)`. | expert-backend | `internal/template/provider_allowlist_audit_test.go` (test func, ~40 LOC) | T-HYBRID-01 | 1 file (extend) | RED — must compile and FAIL initially (no `internal/llm/providers/` directory yet) |
| T-HYBRID-03 | Create `internal/cli/cg_removal_test.go` with three sub-tests: `TestCGCommandReturnsBCError` (assert non-zero exit + `MOAI_CG_REMOVED` + `use 'moai hybrid glm' instead` in stderr), `TestCGCommandIsRegistered` (assert `cgCmd` still registered to rootCmd as BC stub), `TestCGCommandDoesNotLaunchClaudeCode` (assert `unifiedLaunchFunc` not called by runCG). | expert-backend | `internal/cli/cg_removal_test.go` (NEW, ~80 LOC) | T-HYBRID-01 | 1 file (create) | RED — must compile and FAIL initially (current `runCG` launches Claude Code, no BC stub yet) |
| T-HYBRID-04 | Create `internal/cli/hybrid_test.go` with seven integration sub-tests: `TestHybridHelpListsFourProviders`, `TestHybridUnknownProviderRejected`, `TestHybridRequiresTmux`, `TestHybridMissingAPIKey`, `TestHybridProxyOverride`, `TestHybridAuthFailureMasksKey`, `TestHybridBackwardCompatWithCG`. Each test scaffolded per acceptance.md §AC-HYBRID-01..02, 07..09, 13, 16, 17 G/W/T. | expert-backend | `internal/cli/hybrid_test.go` (NEW, ~250 LOC scaffold) | T-HYBRID-01 | 1 file (create) | RED — must compile and FAIL initially (no `hybridCmd`, no `internal/llm/` package yet) |
| T-HYBRID-05 | Create `internal/cli/migration_test.go` with two sub-tests: `TestMigrateCGTeamMode` (setup tmpdir with `team_mode: "cg"` config, invoke `migrateCGTeamMode(root)`, assert post-call `team_mode: "hybrid"` + `provider: "glm"` + stderr emits one-time notice), `TestMigrateCGTeamModeIdempotent` (run twice, assert second run is no-op). | expert-backend | `internal/cli/migration_test.go` (NEW, ~80 LOC) | T-HYBRID-01 | 1 file (create) | RED — must compile and FAIL initially (no `migrateCGTeamMode` helper yet) |
| T-HYBRID-06 | Run `go test ./internal/template/ -run TestProviderAllowlist` and `go test ./internal/cli/ -run "TestCG|TestHybrid|TestMigrate"` — confirm RED state for ALL new tests. Confirm existing tests (`glm_test.go`, `launcher_test.go`, `update_test.go`) GREEN unaffected. | manager-tdd | n/a (verification only) | T-HYBRID-02, T-HYBRID-03, T-HYBRID-04, T-HYBRID-05 | 0 files | RED gate verification |

[HARD] No implementation code in M1 outside of test files. Test failures must reference exact file/function names so subsequent milestones know where to implement.

[HARD] M1 must complete before M2 begins (TDD discipline).

**M1 priority: P0** — blocks all subsequent milestones.

---

## M2: Provider Abstraction Layer (GREEN, part 1) — Priority P0

Goal: Make `TestProviderAllowlist` GREEN. Establish `internal/llm/` package + 4 provider metadata + LLMConfig schema extension.

Reference: plan.md §2 M2, §3.2 to-be-created files.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-07 | Create `internal/llm/provider.go` defining `Provider` interface (`Name()`, `AnthropicBaseURL()`, `EnvKeyName()`, `DefaultModels()`, `BuildEnvVars(apiKey)`, `ProxyOverride(proxyURL)`) + `ModelSet` struct (`High`, `Medium`, `Low`). MX:NOTE tag per plan.md §6.2. | expert-backend | `internal/llm/provider.go` (NEW, ~60 LOC) | T-HYBRID-06 | 1 file (create) | GREEN |
| T-HYBRID-08 | Create `internal/llm/providers/glm.go` defining `GLMProvider` struct implementing `Provider` interface. base_url `https://api.z.ai/api/anthropic`, models `{glm-4.7, glm-4.7, glm-4.5-air}`, env_var `GLM_API_KEY`. `BuildEnvVars` includes Z.AI compat flags (`DISABLE_PROMPT_CACHING=1`) per research.md §5.2. | expert-backend | `internal/llm/providers/glm.go` (NEW, ~60 LOC) | T-HYBRID-07 | 1 file (create) | GREEN |
| T-HYBRID-09 | Create `internal/llm/providers/kimi.go` defining `KimiProvider`. base_url candidate `https://api.moonshot.ai/anthropic` (research.md §3.2 pending), models `{kimi-k2.6, kimi-k2.6, kimi-k2-flash}`, env_var `KIMI_API_KEY`. MX:WARN tag for pending endpoint per plan.md §6.3. | expert-backend | `internal/llm/providers/kimi.go` (NEW, ~60 LOC) | T-HYBRID-07 | 1 file (create) | GREEN |
| T-HYBRID-10 | Create `internal/llm/providers/deepseek.go` defining `DeepSeekProvider`. base_url `https://api.deepseek.com/anthropic` (verified), models `{deepseek-v4-pro, deepseek-v4-pro, deepseek-v4-flash}`, env_var `DEEPSEEK_API_KEY`. | expert-backend | `internal/llm/providers/deepseek.go` (NEW, ~60 LOC) | T-HYBRID-07 | 1 file (create) | GREEN |
| T-HYBRID-11 | Create `internal/llm/providers/qwen.go` defining `QwenProvider`. base_url candidate `https://dashscope.aliyuncs.com/compatible-mode/anthropic` (research.md §3.4 pending), models `{qwen3-coder-plus, qwen3-coder-plus, qwen3-coder-flash}`, env_var `DASHSCOPE_API_KEY`. MX:WARN tag for pending endpoint. | expert-backend | `internal/llm/providers/qwen.go` (NEW, ~60 LOC) | T-HYBRID-07 | 1 file (create) | GREEN |
| T-HYBRID-12 | Create `internal/llm/providers/registry.go` with `Registry` map (4 entries) + `List() []string` returning canonical order `["glm","kimi","deepseek","qwen"]` + `Lookup(name) (Provider, error)`. MX:ANCHOR tag per plan.md §6.1 — Registry is the closed allow-list SSOT. | expert-backend | `internal/llm/providers/registry.go` (NEW, ~50 LOC) | T-HYBRID-08, T-HYBRID-09, T-HYBRID-10, T-HYBRID-11 | 1 file (create) | GREEN |
| T-HYBRID-13 | Create `internal/llm/registry_test.go` verifying registry consistency: `TestRegistryHasFourProviders`, `TestRegistryLookupAllowlist`, `TestRegistryLookupRejectsUnknown`. | expert-backend | `internal/llm/registry_test.go` (NEW, ~80 LOC) | T-HYBRID-12 | 1 file (create) | GREEN — provides REQ-HYBRID-002 unit-level coverage |
| T-HYBRID-14 | Modify `internal/config/types.go` (line 52-100): extend `LLMConfig` struct with `Provider string yaml:"provider"` + `Providers map[string]ProviderYAML yaml:"providers,omitempty"`. Add `type ProviderYAML struct { BaseURL, EnvVar string; Models map[string]string }`. Preserve existing fields (Mode, TeamMode, GLMEnvVar, ClaudeModels, GLM, etc.). | expert-backend | `internal/config/types.go` (MODIFY, ~20 LOC added) | T-HYBRID-12 | 1 file (edit) | GREEN |
| T-HYBRID-15 | Modify `internal/config/defaults.go`: add `NewDefaultProvidersConfig()` returning the 4-provider default mapping (base_url, env_var, models per provider) per research.md §3 verified-stable baseline. Wire into `NewDefaultLLMConfig` so init paths get providers map populated. | expert-backend | `internal/config/defaults.go` (MODIFY, ~30 LOC added) | T-HYBRID-14 | 1 file (edit) | GREEN |
| T-HYBRID-16 | Modify `.moai/config/sections/llm.yaml` + mirror to `internal/template/templates/.moai/config/sections/llm.yaml`: add top-level `provider: ""` + `providers:` section per research.md §2.4. Preserve existing fields (mode, team_mode, claude_models, glm). | manager-docs | 2 files (project + template) | T-HYBRID-15 | 2 files (edit, parity) | GREEN — Embedded-template parity |
| T-HYBRID-17 | Run `cd /Users/goos/MoAI/moai-adk-go && make build` to regenerate `internal/template/embedded.go`. Verify diff is the expected llm.yaml schema addition. | manager-docs | `internal/template/embedded.go` (regenerated) | T-HYBRID-16 | 1 file (regenerated) | Build verification |
| T-HYBRID-18 | Run `go test ./internal/template/ -run TestProviderAllowlist` + `go test ./internal/llm/ -run TestRegistry` — confirm GREEN. | manager-tdd | n/a (verification only) | T-HYBRID-13, T-HYBRID-17 | 0 files | GREEN gate part 1 |

[HARD] T-HYBRID-08..11 may execute in parallel (independent files). T-HYBRID-12 must wait for all 4.

**M2 priority: P0** — blocks M3 (hybrid command depends on registry) and M4 (launcher depends on Provider interface).

---

## M3: `moai hybrid` Command + `moai cg` BC Stub (GREEN, part 2) — Priority P0

Goal: Make `TestCGCommand*` and basic `TestHybrid*` GREEN. Add user-facing commands.

Reference: plan.md §2 M3, §3.1 (cg.go REPLACE) + §3.2 (hybrid.go NEW).

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-19 | Replace `internal/cli/cg.go` (lines 1-58) with BC stub: `cgCmd` registered with `Hidden: false` + `Short: "[REMOVED in v3R3] Use 'moai hybrid glm' instead"`. `runCGRemoved` writes `MOAI_CG_REMOVED` + actionable migration suggestion to stderr + returns `errors.New("MOAI_CG_REMOVED")`. MX:NOTE tag per plan.md §6.2. | expert-backend | `internal/cli/cg.go` (REPLACE, ~30 LOC) | T-HYBRID-18 | 1 file (replace) | GREEN — turns `TestCGCommandReturnsBCError`, `TestCGCommandIsRegistered`, `TestCGCommandDoesNotLaunchClaudeCode` GREEN |
| T-HYBRID-20 | Create `internal/cli/hybrid.go` with `hybridCmd` (`Use: "hybrid <provider>"`) + `runHybrid` (parse `<provider>` arg → `providers.Lookup(name)` → tmux env injection → `unifiedLaunch(profileName, "claude_hybrid", filteredArgs)`). Includes `--proxy <url>` flag handling (REQ-HYBRID-015). MX:ANCHOR tag per plan.md §6.1. | expert-backend | `internal/cli/hybrid.go` (NEW, ~150 LOC) | T-HYBRID-19 | 1 file (create) | GREEN |
| T-HYBRID-21 | Add subcommands to `internal/cli/hybrid.go`: `hybridSetupCmd` (`Use: "setup <provider> [api-key]"` saves to `~/.moai/.env.<provider>`), `hybridStatusCmd` (`Use: "status [<provider>]"` masked key display), `hybridToolsCmd` (`Use: "tools <enable|disable>"` REQ-HYBRID-016 stub returning `not yet implemented (deferred to SPEC-GLM-MCP-001 follow-up)`). `init()` registers all to rootCmd. MX:TODO tag for tools subcommand. | expert-backend | `internal/cli/hybrid.go` (extend, ~80 LOC) | T-HYBRID-20 | 1 file (extend) | GREEN |
| T-HYBRID-22 | Run `go test ./internal/cli/ -run "TestCGCommand|TestHybridHelp|TestHybridUnknownProvider|TestHybridProxyOverride"` — confirm GREEN. | manager-tdd | n/a (verification only) | T-HYBRID-21 | 0 files | GREEN gate part 2 |

[HARD] T-HYBRID-19 and T-HYBRID-20 may NOT run in parallel — `hybridCmd` registration in T-HYBRID-21 conflicts with `cgCmd` registration in T-HYBRID-19 if both modify rootCmd init order. Sequence T-HYBRID-19 → T-HYBRID-20 → T-HYBRID-21.

**M3 priority: P0** — blocks M4 (launcher must route to hybrid mode + GLM helper functions must call new abstractions).

---

## M4: Generalize Injection + Launcher Routing + Migration (GREEN, part 3) — Priority P0

Goal: Make remaining `TestHybrid*` and `TestMigrate*` GREEN. Refactor `glm.go` injection to provider-agnostic helpers + update launcher routing + implement `migrateCGTeamMode`.

Reference: plan.md §2 M4, §3.1 modifications, §3.4 anchors.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-23 | Create `internal/llm/inject.go` extracting provider-agnostic helpers refactored from `internal/cli/glm.go`: `InjectTmuxSessionEnv(provider Provider, apiKey string)` ← refactor lines 356-385; `InjectProviderEnvForTeam(settingsPath, provider, apiKey)` ← refactor lines 520-581 with provider-namespaced backup keys (`MOAI_BACKUP_AUTH_TOKEN_<UPPER>`); `ClearTmuxSessionEnv(providers ...Provider)` ← refactor lines 392-421; `EnsureSettingsLocalJSONHybrid(settingsPath)` ← refactor lines 443-482 (teammateMode="tmux"). MX:ANCHOR + MX:WARN tags per plan.md §6.1, §6.3. | expert-backend | `internal/llm/inject.go` (NEW, ~250 LOC) | T-HYBRID-22 | 1 file (create) | GREEN |
| T-HYBRID-24 | Modify `internal/cli/glm.go`: (a) line 148-149 stderr message `moai cg` → `moai hybrid glm`; (b) `enableTeamMode` (line 218-348) generalized via `enableHybridTeamMode(cmd, providers.GLMProvider{}, isHybrid)` delegate; (c) line 266 `persistTeamMode(root, "cg")` → `persistTeamMode(root, "hybrid")` + `persistProvider(root, "glm")` atomic; (d) line 356-385 `injectTmuxSessionEnv` → delegate to `internal/llm.InjectTmuxSessionEnv`; (e) line 520-581 `injectGLMEnvForTeam` → delegate to `internal/llm.InjectProviderEnvForTeam`; (f) line 536-538 single backup slot → namespaced `MOAI_BACKUP_AUTH_TOKEN_GLM`. **Preserve** `glmCmd`, `glmSetupCmd`, `glmStatusCmd` (lines 21-89) + `runGLM` body for REQ-HYBRID-006 no-regression. | expert-backend | `internal/cli/glm.go` (MODIFY, ~50 LOC delta) | T-HYBRID-23 | 1 file (edit) | GREEN — must NOT break existing `TestRunCC*`, `TestRunGLM*`, `TestEnableTeamMode*` tests |
| T-HYBRID-25 | Modify `internal/cli/launcher.go`: (a) line 40 comment `claude_glm` → `claude_hybrid` rename; (b) line 57 `case "claude_glm":` → `case "claude_hybrid":` switching on `LLMConfig.Provider` field via `providers.Lookup(provider)`; (c) line 185 `persistTeamMode(root, "cg")` → `persistTeamMode(root, "hybrid")` + persist `provider` field; (d) line 267 `resetTeamModeForCC` rename comment to "switching to CC clears hybrid team mode". MX:WARN tag per plan.md §6.3 — case label rename requires synchronous update at multiple call sites. | expert-backend | `internal/cli/launcher.go` (MODIFY, ~10 LOC delta) | T-HYBRID-24 | 1 file (edit) | GREEN |
| T-HYBRID-26 | Modify `internal/cli/update.go`: add `migrateCGTeamMode(root string)` helper at end of `cleanMoaiManagedPaths` (line ~1411-1441). Reads `.moai/config/sections/llm.yaml`; if `team_mode: "cg"` AND `provider: ""`, atomic-rewrite to `team_mode: "hybrid"` + `provider: "glm"` (uses SPEC-V3R3-UPDATE-CLEANUP-001 atomic write infrastructure). Emits one-time stderr notice. Idempotent (second invocation is no-op). MX:NOTE tag per plan.md §6.2. | expert-backend | `internal/cli/update.go` (MODIFY, ~50 LOC added) | T-HYBRID-25 | 1 file (edit) | GREEN |
| T-HYBRID-27 | Run `cd /Users/goos/MoAI/moai-adk-go && make build` to regenerate `internal/template/embedded.go`. Then `go test ./internal/cli/ -run "TestHybrid|TestMigrate"` (full suite) → GREEN. Then `go test ./...` from repo root → 0 cascading failures (per CLAUDE.local.md §6 HARD). Verify existing `TestRunCC*`, `TestRunGLM*`, `TestEnableTeamMode*` tests still GREEN (REQ-HYBRID-006 no-regression). | manager-tdd | n/a (verification only) | T-HYBRID-26 | 1 file (regenerated) | GREEN gate part 3 — full suite |

[HARD] T-HYBRID-24 must NOT modify `glmCmd` registration, `runGLM` body (except line 148-149 message), or `moai glm setup` / `moai glm status` behavior — these are preserved per REQ-HYBRID-006.

[HARD] T-HYBRID-23 → T-HYBRID-24 → T-HYBRID-25 → T-HYBRID-26 sequential (each layer depends on the previous abstraction).

**M4 priority: P0** — blocks M5 (documentation can only substitute references after the actual mode rename lands).

---

## M5: Documentation Substitution + CHANGELOG + MX Tags (REFACTOR + Trackable) — Priority P1

Goal: Substitute `moai cg` references across 6+ documentation files + insert MX tags + add CHANGELOG `BC-V3R3-HYBRID-001` entry + final verification.

Reference: plan.md §2 M5 (M5a-i), §3.1 documentation modifications.

### M5a-d: `.claude/skills/` and `.claude/rules/` substitution

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-28 | Edit `.claude/skills/moai/team/glm.md` (lines 41, 50, 103, 127, 143): substitute `moai cg` → `moai hybrid glm` (5 mentions) per plan.md §2 M5a. Mirror to `internal/template/templates/.claude/skills/moai/team/glm.md`. | manager-docs | 2 files (project + template) | T-HYBRID-27 | 2 files (edit, parity) | REFACTOR |
| T-HYBRID-29 | Edit `.claude/skills/moai/team/run.md` (lines 73, 88, 92, 103, 104): substitute `moai cg` → `moai hybrid <provider>` (5 mentions) per plan.md §2 M5b. Update line 73 table row similarly. Mirror to template. | manager-docs | 2 files (project + template) | T-HYBRID-27 | 2 files (edit, parity) | REFACTOR |
| T-HYBRID-30 | Edit `.claude/skills/moai/workflows/run.md:904`: `active_mode: cc \| glm \| cg` → `active_mode: cc \| glm \| hybrid` per plan.md §2 M5c. Mirror to template. | manager-docs | 2 files (project + template) | T-HYBRID-27 | 2 files (edit, parity) | REFACTOR |
| T-HYBRID-31 | Edit `.claude/rules/moai/development/model-policy.md:44`: `Activation: moai cg (requires tmux)` → `Activation: moai hybrid <provider> (requires tmux). Example: moai hybrid glm` per plan.md §2 M5d. Mirror to template. | manager-docs | 2 files (project + template) | T-HYBRID-27 | 2 files (edit, parity) | REFACTOR |

### M5e-f: CLAUDE.md + CLAUDE.local.md substitution

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-32 | Edit `CLAUDE.md §15 Agent Teams (Experimental)` (~line 488-510, `### CG Mode (Claude + GLM Cost Optimization)` heading): rename heading to `### Hybrid Mode (Multi-LLM Cost Optimization)`; generalize body to "Teammates (GLM/Kimi/DeepSeek/Qwen, new tmux panes)"; activation example `moai hybrid <provider>`. CLAUDE.md is NOT in template (project root only). | manager-docs | `CLAUDE.md` (MODIFY) | T-HYBRID-27 | 1 file (edit) | REFACTOR |
| T-HYBRID-33 | Edit `CLAUDE.local.md §16 line 129`: `Modified by 'moai glm', 'moai cc', 'moai cg' commands at runtime` → `Modified by 'moai glm', 'moai cc', 'moai hybrid <provider>' commands at runtime`. CLAUDE.local.md is local-only (not in template). | manager-docs | `CLAUDE.local.md` (MODIFY) | T-HYBRID-27 | 1 file (edit) | REFACTOR |

### M5g: CHANGELOG entry

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-34 | Append `BC-V3R3-HYBRID-001` entry to `CHANGELOG.md` `## [Unreleased]` section per plan.md §2 M5g + spec.md §10.3. Include: 4-provider list, BC migration suggestion, auto-migration of `team_mode: "cg"` config, preservation of `moai glm` / `moai cc`, reference to SPEC §10. | manager-docs | `CHANGELOG.md` (MODIFY) | T-HYBRID-27 | 1 file (edit) | Trackable (TRUST 5) |

### M5h: MX tag insertion

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-35 | Insert 10 MX tags across 7 distinct files per plan.md §6 mx_plan: 3 ANCHOR (`registry.go Registry`, `hybrid.go hybridCmd`, `inject.go InjectProviderEnvForTeam`), 3 NOTE (`cg.go runCGRemoved`, `provider.go Provider interface`, `update.go migrateCGTeamMode`), 3 WARN (`inject.go backup namespacing`, `launcher.go:57 case`, `providers/{kimi,qwen}.go base_url`), 1 TODO (`hybrid.go hybridToolsCmd` deferral to SPEC-GLM-MCP-001). | expert-backend | 7 distinct files (Go source) | T-HYBRID-23, T-HYBRID-19, T-HYBRID-21, T-HYBRID-25, T-HYBRID-26, T-HYBRID-09, T-HYBRID-11 | 7 files (edit) | REFACTOR — MX protocol per `.claude/rules/moai/workflow/mx-tag-protocol.md` |

### M5i: Final verification

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|------------|-------------------|------------|--------------|---------------|
| T-HYBRID-36 | Run final verification battery: (a) `cd /Users/goos/MoAI/moai-adk-go && make build` regenerate embedded.go; (b) `go test ./...` from repo root — 0 cascading failures (CLAUDE.local.md §6 HARD); (c) `golangci-lint run` — 0 errors, ≤10 warnings (per quality.yaml); (d) `go vet ./...` — 0 issues; (e) `grep -rn "moai cg" .claude/ CLAUDE.md CLAUDE.local.md` returns only M5e/M5f-substituted content (no residual command references); (f) Update `progress.md` with `run_complete_at: <ISO-8601 UTC>` + `run_status: implementation-complete`. | manager-tdd | `progress.md` (MODIFY) + 1 file (regenerated) | T-HYBRID-28..35 | 2 files | REFACTOR / Trackable — gating point per CLAUDE.local.md §6 |

[HARD] T-HYBRID-36 is the gating point. No tests may regress. No new SPEC documents created in `.moai/specs/` or `.moai/reports/` during M5.

T-HYBRID-28..31 may execute in parallel (independent files). T-HYBRID-32, T-HYBRID-33, T-HYBRID-34 may execute in parallel after M4 completes.

**M5 priority: P1** — required for full spec.md §2.1 In Scope coverage and TRUST 5 Trackable.

---

## Aggregate Statistics

- **Total tasks**: 36 (T-HYBRID-01 through T-HYBRID-36)
- **Total milestones**: 5 (M1: 6 tasks, M2: 12 tasks, M3: 4 tasks, M4: 5 tasks, M5: 9 tasks)
- **Files created**: 13
  - 4 test scaffolds (M1): `provider_allowlist_audit_test.go`, `cg_removal_test.go`, `hybrid_test.go`, `migration_test.go`
  - 7 implementation files (M2): `provider.go`, `providers/{glm,kimi,deepseek,qwen,registry}.go`, `registry_test.go`
  - 1 implementation file (M3): `hybrid.go`
  - 1 implementation file (M4): `inject.go`
- **Files modified (source-of-truth)**: ~13 unique files
  - 1 `cg.go` REPLACE (M3)
  - 4 Go source files (`glm.go`, `launcher.go`, `update.go`, `config/types.go`, `config/defaults.go`) (M2/M4)
  - 1 project `llm.yaml` (M2)
  - 4 `.claude/skills/.claude/rules` files (M5a-d)
  - 1 `CLAUDE.md` (M5e), 1 `CLAUDE.local.md` (M5f), 1 `CHANGELOG.md` (M5g)
- **Files modified (embedded-template parity)**: ~5 mirror files (`.claude/skills/moai/team/{glm,run}.md`, `.claude/skills/moai/workflows/run.md`, `.claude/rules/moai/development/model-policy.md`, `.moai/config/sections/llm.yaml`)
- **MX tag insertions**: 10 across 7 distinct files (3 ANCHOR + 3 NOTE + 3 WARN + 1 TODO)
- **Owner-role distribution**:
  - `expert-backend`: 19 tasks (all Go test + implementation work — T-HYBRID-01..05, 07..15, 19..21, 23..26, 35)
  - `manager-tdd`: 4 tasks (T-HYBRID-06, 18, 22, 27, 36 — TDD gate verification)
  - `manager-docs`: 13 tasks (T-HYBRID-16, 17, 28..34 — content authoring + template parity + CHANGELOG)

---

## Owner-Role Rationale

- **Go test + implementation work** (`expert-backend`): all `internal/llm/` package work, `internal/cli/{hybrid,cg,glm,launcher,update}.go` modifications, MX tag insertion in Go source files. Requires Go expertise (cobra, embedded FS, atomic write, tmux env injection).
- **TDD gate verification** (`manager-tdd`): the project's `quality.yaml` declares `development_mode: tdd`. Each gate (RED M1, GREEN parts 1/2/3, final REFACTOR) is a manager-tdd checkpoint per CLAUDE.md §5 Run Phase methodology.
- **Content authoring** (`manager-docs`): all `.claude/skills/.claude/rules` substitutions (M5a-d), embedded template parity (`.moai/config/sections/llm.yaml`), CLAUDE.md/CLAUDE.local.md/CHANGELOG.md edits. Documentation per CLAUDE.md §4 Manager Agents catalog.

---

## Parallel Execution Opportunities

These task groups have no inter-dependencies and may run in parallel within their milestone:

- **M1 parallel**: T-HYBRID-02..05 (extend independent test files; true parallelism possible).
- **M2 parallel**: T-HYBRID-08, T-HYBRID-09, T-HYBRID-10, T-HYBRID-11 (4 provider metadata files independent; true parallelism possible). T-HYBRID-12 sequences after.
- **M3 sequential**: T-HYBRID-19 → T-HYBRID-20 → T-HYBRID-21 (rootCmd registration order matters).
- **M4 sequential**: T-HYBRID-23 → T-HYBRID-24 → T-HYBRID-25 → T-HYBRID-26 (each layer depends on the previous abstraction).
- **M5 parallel**: T-HYBRID-28..31 (independent skill/rule files); T-HYBRID-32..34 (independent project-root files) parallel; T-HYBRID-35 sequences after to avoid edit conflicts on Go source files; T-HYBRID-36 sequences last.

Per CLAUDE.md §1 HARD rule "Multi-File Decomposition: Split work when modifying 3+ files," every milestone with multi-file scope encodes per-file decomposition.

---

## Cross-Reference Map

Each task references which REQ(s) and which AC(s) it advances toward DoD:

| Task ID | REQ coverage | AC coverage |
|---------|--------------|-------------|
| T-HYBRID-01 | (scaffold) | (scaffold) |
| T-HYBRID-02 | REQ-HYBRID-002, 017 | AC-HYBRID-15, 18 |
| T-HYBRID-03 | REQ-HYBRID-003, 010, 014 | AC-HYBRID-03 |
| T-HYBRID-04 | REQ-HYBRID-001, 002, 007..009, 012, 013, 015, 018 | AC-HYBRID-01, 02, 07..09, 11, 12, 13, 16 |
| T-HYBRID-05 | REQ-HYBRID-011 | AC-HYBRID-10 |
| T-HYBRID-06 | (gate) | (gate) |
| T-HYBRID-07 | REQ-HYBRID-001, 002 | (interface foundation) |
| T-HYBRID-08 | REQ-HYBRID-002, 005 | AC-HYBRID-05 |
| T-HYBRID-09 | REQ-HYBRID-002, 005, 015 | AC-HYBRID-05, 13 |
| T-HYBRID-10 | REQ-HYBRID-002, 005 | AC-HYBRID-05 |
| T-HYBRID-11 | REQ-HYBRID-002, 005, 015 | AC-HYBRID-05, 13 |
| T-HYBRID-12 | REQ-HYBRID-002, 017 | AC-HYBRID-15, 18 |
| T-HYBRID-13 | REQ-HYBRID-002 | AC-HYBRID-02, 18 |
| T-HYBRID-14 | REQ-HYBRID-004 | AC-HYBRID-04 |
| T-HYBRID-15 | REQ-HYBRID-005 | AC-HYBRID-05, 17 |
| T-HYBRID-16 | REQ-HYBRID-004, 005 | AC-HYBRID-04, 05 |
| T-HYBRID-17 | (build) | (build) |
| T-HYBRID-18 | (gate) | AC-HYBRID-05, 15, 18 GREEN |
| T-HYBRID-19 | REQ-HYBRID-003, 010, 014 | AC-HYBRID-03 |
| T-HYBRID-20 | REQ-HYBRID-001, 002, 007, 015 | AC-HYBRID-01, 02, 13 |
| T-HYBRID-21 | REQ-HYBRID-008, 009, 016 | AC-HYBRID-08, 09, 14 |
| T-HYBRID-22 | (gate) | AC-HYBRID-01, 02, 03, 13, 14 GREEN |
| T-HYBRID-23 | REQ-HYBRID-007, 012, 018 | AC-HYBRID-07, 11, 16 |
| T-HYBRID-24 | REQ-HYBRID-006, 007 | AC-HYBRID-06, 07 |
| T-HYBRID-25 | REQ-HYBRID-007, 012, 013 | AC-HYBRID-07, 11, 12 |
| T-HYBRID-26 | REQ-HYBRID-011 | AC-HYBRID-10 |
| T-HYBRID-27 | (gate) | AC-HYBRID-06..12, 16, 17 GREEN |
| T-HYBRID-28..31 | REQ-HYBRID-001 (Trackable) | (documentation surface) |
| T-HYBRID-32 | REQ-HYBRID-001 (Trackable) | (CLAUDE.md surface) |
| T-HYBRID-33 | REQ-HYBRID-001 (Trackable) | (CLAUDE.local.md surface) |
| T-HYBRID-34 | (Trackable) | (CHANGELOG entry) |
| T-HYBRID-35 | (MX protocol) | (10 MX tags) |
| T-HYBRID-36 | (final gate) | (DoD steps) |

REQ coverage summary: 18 REQs, all referenced by at least one task (verified by transitive lookup against `spec.md` §5).

AC coverage summary: 18 ACs, all advanced toward DoD by tasks (verified against acceptance.md §1-18).

REQ Categories balance:
- Ubiquitous (REQ-001..006): T-HYBRID-04, 07, 08..14, 19, 20, 24
- Event-Driven (REQ-007..011): T-HYBRID-04, 05, 20, 21, 23..26
- State-Driven (REQ-012..014): T-HYBRID-03, 04, 19, 23, 25
- Optional (REQ-015..016): T-HYBRID-04, 09, 11, 20, 21
- Complex (REQ-017..018): T-HYBRID-02, 04, 12, 13, 23

Every category has at least one task. Task-to-REQ mapping covers all 18 REQs.

---

## Final Verification Pass (subset of T-HYBRID-36)

Before marking SPEC implementation complete, execute these checks in order:

1. **Audit + integration tests green**: `go test ./internal/template/ -run TestProviderAllowlist` + `go test ./internal/cli/ -run "TestCG|TestHybrid|TestMigrate"` + `go test ./internal/llm/ -run TestRegistry` — all PASS.
2. **Full repo green**: `go test ./...` — 0 failures (per CLAUDE.local.md §6 HARD).
3. **No regression on `moai cc` / `moai glm`**: `go test ./internal/cli/ -run "TestRunCC|TestRunGLM|TestEnableTeamMode"` — all PASS (REQ-HYBRID-006).
4. **Lint clean**: `golangci-lint run` — 0 errors, ≤10 warnings (per quality.yaml `lsp_quality_gates.sync`).
5. **Vet clean**: `go vet ./...` — 0 issues.
6. **Embedded parity**: diff `internal/template/embedded.go` after `make build` shows only the expected schema/file content changes (no spurious diffs).
7. **No residual `moai cg`**: `grep -rn "moai cg" .claude/ CLAUDE.md CLAUDE.local.md` returns only the substituted/migration-note lines (no command-invocation references).
8. **DoD checklist**: all items in `acceptance.md` §"Definition of Done" checked.

If any step fails, return to the appropriate milestone for remediation. Do NOT advance to `/moai sync` until all 8 checks pass.

---

End of tasks.md.
