# SPEC-V3R3-HYBRID-001 Implementation Plan (Phase 1B)

> Implementation plan for `moai hybrid` Multi-LLM Mode (cg deprecation + GLM/Kimi/DeepSeek/Qwen3-Coder team support).
> Companion to `spec.md` v0.1.0 and `research.md` v0.1.0.
> Authored against branch `feature/SPEC-V3R3-HYBRID-001` at `/Users/goos/MoAI/moai-adk-go` (solo mode, no worktree).

## HISTORY

| Version | Date       | Author                        | Description                                                              |
|---------|------------|-------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B) | 최초 작성 — `moai cg` 제거 + `moai hybrid` 도입 5-milestone plan + REQ↔AC traceability |

---

## 1. Plan Overview

### 1.1 Goal restatement

본 plan은 spec.md REQ-HYBRID-001..018을 실행 가능한 5-milestone 작업 분해로 변환한다. 핵심 deliverable:

- **신규 abstraction layer**: `internal/llm/provider.go` (interface) + `internal/llm/providers/{glm,kimi,deepseek,qwen}.go` (4 메타데이터 파일).
- **신규 명령**: `internal/cli/hybrid.go` — `moai hybrid <provider> [-p profile] [-- claude-args...]`.
- **BC stub**: `internal/cli/cg.go` — `MOAI_CG_REMOVED` actionable 에러 반환.
- **Generalize**: `internal/cli/glm.go`의 GLM-specific 함수들을 provider-agnostic helper로 refactor (LOC ~200 이주, ~50 LOC adjustments).
- **Routing 일반화**: `internal/cli/launcher.go:57 case "claude_glm"` → `case "claude_hybrid"` + provider 라우팅.
- **Schema 확장**: `.moai/config/sections/llm.yaml` + embedded template — `provider` + `providers.<name>.{base_url, models, env_var}` 추가.
- **Test surface**: `provider_allowlist_audit_test.go` (REQ-HYBRID-002, 017), `cg_removal_test.go` (REQ-HYBRID-003, 010, 014), `hybrid_test.go` (REQ-HYBRID-007..009, 012, 013, 015), `migration_test.go` (REQ-HYBRID-011).
- **Documentation substitution**: ~10 mention across 6+ files (CLAUDE.md §15, CLAUDE.local.md §16, .claude/skills/moai/team/glm.md, .claude/skills/moai/team/run.md, .claude/skills/moai/workflows/run.md, .claude/rules/moai/development/model-policy.md).
- **CHANGELOG**: `BC-V3R3-HYBRID-001` entry with migration table.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED (M1)**: 4 신규 audit/unit tests를 먼저 작성. 모두 실패 상태 확인 (`provider_allowlist_audit_test.go`, `cg_removal_test.go`, `hybrid_test.go`, `migration_test.go`).
- **GREEN part 1 (M2)**: `internal/llm/provider.go` interface + 4 provider 메타데이터 파일 + `internal/config/types.go` LLMConfig 확장. provider_allowlist_audit_test → GREEN.
- **GREEN part 2 (M3)**: `internal/cli/hybrid.go` 신규 명령 + `internal/cli/cg.go` BC stub. cg_removal_test + hybrid_test (basic) → GREEN.
- **GREEN part 3 (M4)**: `internal/cli/glm.go` generalize + `injectProviderEnvForTeam` + `internal/cli/launcher.go:57` 분기 변경 + `internal/cli/update.go` migrateCGTeamMode. hybrid_test (full) + migration_test → GREEN.
- **REFACTOR (M5)**: 6+ 문서 substitution + CLAUDE.md §15 / CLAUDE.local.md §16 + CHANGELOG `BC-V3R3-HYBRID-001` entry + MX tag 삽입 + final `make build` + full `go test ./...`.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|---|---|---|
| Provider abstraction interface | `internal/llm/provider.go` (NEW) | REQ-HYBRID-001, 002, 005, 017 |
| GLM provider metadata | `internal/llm/providers/glm.go` (NEW) | REQ-HYBRID-002, 005 |
| Kimi K2 provider metadata | `internal/llm/providers/kimi.go` (NEW) | REQ-HYBRID-002, 005, 015 |
| DeepSeek V4 provider metadata | `internal/llm/providers/deepseek.go` (NEW) | REQ-HYBRID-002, 005 |
| Qwen3-Coder provider metadata | `internal/llm/providers/qwen.go` (NEW) | REQ-HYBRID-002, 005, 015 |
| Provider registry + allow-list test | `internal/llm/providers/registry.go` + `internal/llm/registry_test.go` (NEW) | REQ-HYBRID-002, 017 |
| `moai hybrid` 명령 정의 | `internal/cli/hybrid.go` (NEW) | REQ-HYBRID-001, 002, 007, 008, 009, 015, 016 |
| `moai cg` BC stub | `internal/cli/cg.go` (REPLACE) | REQ-HYBRID-003, 010, 014 |
| LLMConfig 확장 (`Provider` + `Providers map`) | `internal/config/types.go` (MODIFY) | REQ-HYBRID-004 |
| Provider 기본값 | `internal/config/defaults.go` (MODIFY) | REQ-HYBRID-005 |
| `provider-agnostic` injection helpers | `internal/llm/inject.go` (NEW) — refactor of `glm.go`'s `injectGLMEnvForTeam`, `injectTmuxSessionEnv`, etc. | REQ-HYBRID-007, 012, 018 |
| `internal/cli/glm.go` generalize callsites | `internal/cli/glm.go` (MODIFY) | REQ-HYBRID-006, 007 (보존 + GLM provider로 위임) |
| Launcher 라우팅 일반화 | `internal/cli/launcher.go` (MODIFY) | REQ-HYBRID-007 |
| `moai update` migration helper | `internal/cli/update.go` (MODIFY) | REQ-HYBRID-011 |
| Project llm.yaml 스키마 확장 | `.moai/config/sections/llm.yaml` (MODIFY) | REQ-HYBRID-004 |
| Embedded template llm.yaml 미러 | `internal/template/templates/.moai/config/sections/llm.yaml` (MODIFY) | REQ-HYBRID-004 |
| Provider allow-list audit test | `internal/template/provider_allowlist_audit_test.go` (NEW) | REQ-HYBRID-002, 017 |
| `moai cg` removal test | `internal/cli/cg_removal_test.go` (NEW) | REQ-HYBRID-003, 010, 014 |
| `moai hybrid` integration tests | `internal/cli/hybrid_test.go` (NEW) | REQ-HYBRID-001, 002, 007..009, 012, 013, 015, 018 |
| Migration test (`team_mode: cg` → `hybrid`) | `internal/cli/migration_test.go` (NEW) | REQ-HYBRID-011 |
| CLAUDE.md §15 substitution | `CLAUDE.md` (MODIFY) | Trackable |
| CLAUDE.local.md §16 substitution | `CLAUDE.local.md` (MODIFY) | Trackable |
| `.claude/skills/moai/team/glm.md` substitution (5 lines) | `.claude/skills/moai/team/glm.md` (MODIFY) | REQ-HYBRID-001 |
| `.claude/skills/moai/team/run.md` substitution (5 lines) | `.claude/skills/moai/team/run.md` (MODIFY) | REQ-HYBRID-001 |
| `.claude/skills/moai/workflows/run.md:904` | `.claude/skills/moai/workflows/run.md` (MODIFY) | REQ-HYBRID-001 |
| `.claude/rules/moai/development/model-policy.md:44` | `.claude/rules/moai/development/model-policy.md` (MODIFY) | REQ-HYBRID-001 |
| CHANGELOG `BC-V3R3-HYBRID-001` entry | `CHANGELOG.md` (MODIFY) | Trackable (TRUST 5) |

[HARD] Embedded-template parity: 모든 `.claude/...` 변경은 `internal/template/templates/.claude/...` mirror + `make build` 필수 (CLAUDE.local.md §2 Template-First HARD).

### 1.4 Traceability Matrix (REQ → AC mapping)

Plan-auditor PASS criterion #2: every REQ maps to at least one AC.

| REQ ID | Category | Mapped AC(s) |
|---|---|---|
| REQ-HYBRID-001 | Ubiquitous | AC-HYBRID-01, AC-HYBRID-14 |
| REQ-HYBRID-002 | Ubiquitous | AC-HYBRID-01, AC-HYBRID-02, AC-HYBRID-15, AC-HYBRID-18 |
| REQ-HYBRID-003 | Ubiquitous | AC-HYBRID-03 |
| REQ-HYBRID-004 | Ubiquitous | AC-HYBRID-04 |
| REQ-HYBRID-005 | Ubiquitous | AC-HYBRID-05, AC-HYBRID-17 |
| REQ-HYBRID-006 | Ubiquitous | AC-HYBRID-06 |
| REQ-HYBRID-007 | Event-Driven | AC-HYBRID-07 |
| REQ-HYBRID-008 | Event-Driven | AC-HYBRID-08 |
| REQ-HYBRID-009 | Event-Driven | AC-HYBRID-09 |
| REQ-HYBRID-010 | Event-Driven | AC-HYBRID-03 |
| REQ-HYBRID-011 | Event-Driven | AC-HYBRID-10 |
| REQ-HYBRID-012 | State-Driven | AC-HYBRID-11 |
| REQ-HYBRID-013 | State-Driven | AC-HYBRID-12 |
| REQ-HYBRID-014 | State-Driven | AC-HYBRID-03, AC-HYBRID-17 |
| REQ-HYBRID-015 | Optional | AC-HYBRID-13 |
| REQ-HYBRID-016 | Optional | AC-HYBRID-14 |
| REQ-HYBRID-017 | Complex (Unwanted) | AC-HYBRID-15, AC-HYBRID-18 |
| REQ-HYBRID-018 | Complex (Auth Failure) | AC-HYBRID-16 |

Coverage: **18/18 REQs mapped, 18/18 ACs validated** (some ACs map to multiple REQs; see acceptance.md for full G/W/T per AC).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

Reference: `internal/template/lang_boundary_audit_test.go` (SPEC-V3R2-WF-005 M1 패턴), `internal/cli/glm_test.go` (existing test 패턴).

Owner role: `expert-backend` (Go test) or direct `manager-tdd` execution.

Scope:

1. **`internal/template/provider_allowlist_audit_test.go`** (REQ-HYBRID-002, 017):
   - `TestProviderAllowlist`: walk `internal/llm/providers/` directory in embedded FS, assert exactly 4 `.go` files (excluding `_test.go`) named `glm.go`, `kimi.go`, `deepseek.go`, `qwen.go`. On 5th file: `t.Errorf("PROVIDER_ALLOWLIST_VIOLATION: %s exists; v3R3 allow-list is closed at four (glm, kimi, deepseek, qwen)", path)`.
   - `TestProviderRegistryConsistency`: load `internal/llm/providers/registry.go` provider list, assert 4 entries.
2. **`internal/cli/cg_removal_test.go`** (REQ-HYBRID-003, 010, 014):
   - `TestCGCommandReturnsBCError`: invoke `moai cg` via cobra subcommand, assert non-zero exit + stderr contains `MOAI_CG_REMOVED` + `use 'moai hybrid glm' instead`.
   - `TestCGCommandIsRegistered`: assert `cgCmd` is still registered to rootCmd (BC stub, not removed entirely).
   - `TestCGCommandDoesNotLaunchClaudeCode`: assert `unifiedLaunchFunc` not called by runCG (BC path short-circuits).
3. **`internal/cli/hybrid_test.go`** (REQ-HYBRID-001, 002, 007..009, 012, 013, 015, 018):
   - `TestHybridHelpListsFourProviders`: invoke `moai hybrid --help`, assert output contains "glm", "kimi", "deepseek", "qwen".
   - `TestHybridUnknownProviderRejected`: invoke `moai hybrid mistral`, assert exit code non-zero + stderr contains `UNKNOWN_PROVIDER`.
   - `TestHybridRequiresTmux`: simulate non-tmux env (`MOAI_TEST_MODE=0`, no `TMUX` env), invoke `moai hybrid glm`, assert `HYBRID_REQUIRES_TMUX` error.
   - `TestHybridMissingAPIKey`: simulate tmux env + missing `~/.moai/.env.glm`, assert `HYBRID_MISSING_API_KEY`.
   - `TestHybridProxyOverride`: invoke `moai hybrid kimi --proxy https://my-proxy/anthropic`, assert ANTHROPIC_BASE_URL set to proxy URL.
   - `TestHybridAuthFailureMasksKey`: simulate HTTP 401 from provider, assert error message contains `HYBRID_AUTH_FAILED` + masked key (no plain-text key).
   - `TestHybridBackwardCompatWithCG`: load fixture `team_mode: "cg"` config, run `moai update`, assert `team_mode: "hybrid"` + `provider: "glm"` post-update + tmux env-vars match v2.x bit-for-bit.
4. **`internal/cli/migration_test.go`** (REQ-HYBRID-011):
   - `TestMigrateCGTeamMode`: setup tmpdir with `.moai/config/sections/llm.yaml` containing `team_mode: "cg"`, invoke `migrateCGTeamMode(root)`, assert post-call file contains `team_mode: "hybrid"` + `provider: "glm"` + stderr emits one-time notice.
   - `TestMigrateCGTeamModeIdempotent`: run twice, assert second run is no-op (notice emitted only once).

Verification gate before advancing to M2:
- `go test ./internal/template/ -run TestProviderAllowlist` → RED (no `internal/llm/providers/` directory yet).
- `go test ./internal/cli/ -run "TestCG|TestHybrid|TestMigrate"` → RED (functions/files missing).

[HARD] No implementation code in M1 outside of test files. The test failures must reference exact file/function names so subsequent milestones know where to implement.

### M2: Provider abstraction layer (GREEN, part 1) — Priority P0

Owner role: `expert-backend`.

Scope:

1. **`internal/llm/provider.go`** (NEW): interface definition.
   ```go
   type Provider interface {
       Name() string                          // "glm" | "kimi" | "deepseek" | "qwen"
       AnthropicBaseURL() string
       EnvKeyName() string                    // e.g., "GLM_API_KEY"
       DefaultModels() ModelSet               // High, Medium, Low
       BuildEnvVars(apiKey string) map[string]string
       ProxyOverride(proxyURL string) Provider  // returns a copy with overridden base_url
   }
   type ModelSet struct { High, Medium, Low string }
   ```
2. **`internal/llm/providers/glm.go`** (NEW): GLM provider metadata. base_url `https://api.z.ai/api/anthropic`, models `{glm-4.7, glm-4.7, glm-4.5-air}`, env_var `GLM_API_KEY`. BuildEnvVars includes Z.AI compat flags (DISABLE_PROMPT_CACHING=1, etc.) per research.md §5.2.
3. **`internal/llm/providers/kimi.go`** (NEW): Kimi K2 provider metadata. base_url candidate `https://api.moonshot.ai/anthropic` (research.md §3.2 pending), models `{kimi-k2.6, kimi-k2.6, kimi-k2-flash}`, env_var `KIMI_API_KEY`. BuildEnvVars TBD per provider docs (likely DISABLE_PROMPT_CACHING=1 by default until Kimi docs confirm).
4. **`internal/llm/providers/deepseek.go`** (NEW): DeepSeek V4 provider metadata. base_url `https://api.deepseek.com/anthropic`, models `{deepseek-v4-pro, deepseek-v4-pro, deepseek-v4-flash}`, env_var `DEEPSEEK_API_KEY`. BuildEnvVars TBD.
5. **`internal/llm/providers/qwen.go`** (NEW): Qwen3-Coder provider metadata. base_url candidate `https://dashscope.aliyuncs.com/compatible-mode/anthropic` (research.md §3.4 pending), models `{qwen3-coder-plus, qwen3-coder-plus, qwen3-coder-flash}`, env_var `DASHSCOPE_API_KEY`. BuildEnvVars TBD.
6. **`internal/llm/providers/registry.go`** (NEW): provider registry map.
   ```go
   var Registry = map[string]Provider{
       "glm":      GLMProvider{},
       "kimi":     KimiProvider{},
       "deepseek": DeepSeekProvider{},
       "qwen":     QwenProvider{},
   }
   func List() []string { return []string{"glm", "kimi", "deepseek", "qwen"} }
   func Lookup(name string) (Provider, error) { ... }
   ```
7. **`internal/config/types.go`** (MODIFY): extend `LLMConfig`.
   ```go
   type LLMConfig struct {
       // ... existing fields preserved (Mode, TeamMode, GLMEnvVar, ClaudeModels, GLM, ...)
       Provider  string                   `yaml:"provider"`              // NEW: "" or one of allow-list
       Providers map[string]ProviderYAML `yaml:"providers,omitempty"`   // NEW: per-provider config
   }
   type ProviderYAML struct {
       BaseURL string                 `yaml:"base_url"`
       EnvVar  string                 `yaml:"env_var"`
       Models  map[string]string     `yaml:"models"`  // {high, medium, low}
   }
   ```
8. **`internal/config/defaults.go`** (MODIFY): add `NewDefaultProvidersConfig()` returning the 4-provider default mapping.
9. **`.moai/config/sections/llm.yaml`** + **`internal/template/templates/.moai/config/sections/llm.yaml`** (MODIFY): add `provider: ""` + `providers:` section per research.md §2.4.
10. Run `make build` to regenerate `internal/template/embedded.go`.

Verification:
- `go test ./internal/llm/... -run TestProviderRegistry` → GREEN (4 entries).
- `go test ./internal/template/ -run TestProviderAllowlist` → GREEN (exactly 4 .go files in providers/).
- `internal/llm/providers/registry_test.go` covers REQ-HYBRID-002, REQ-HYBRID-005 baseline.

### M3: `moai hybrid` 명령 + `moai cg` BC stub (GREEN, part 2) — Priority P0

Owner role: `expert-backend`.

Scope:

1. **`internal/cli/hybrid.go`** (NEW):
   - `var hybridCmd = &cobra.Command{ Use: "hybrid <provider>", ... }` with provider subcommand routing.
   - `func runHybrid(cmd, args)`: parse `<provider>` arg → `providers.Lookup(name)` → `provider.BuildEnvVars(apiKey)` → tmux env injection + `unifiedLaunch(profileName, "claude_hybrid", filteredArgs)` (post-M4 launcher.go change).
   - `--proxy <url>` flag handling (REQ-HYBRID-015).
   - `var hybridSetupCmd = ... { Use: "setup <provider> [api-key]" }` — saves to `~/.moai/.env.<provider>`.
   - `var hybridStatusCmd = ... { Use: "status [<provider>]" }` — masked key display per provider.
   - `var hybridToolsCmd = ... { Use: "tools <enable|disable>" }` (REQ-HYBRID-016 stub) — returns `not yet implemented (deferred to SPEC-GLM-MCP-001 follow-up)` message.
   - `init()` registers `hybridCmd` to rootCmd.
2. **`internal/cli/cg.go`** (REPLACE): BC stub.
   ```go
   var cgCmd = &cobra.Command{
       Use:    "cg",
       Hidden: false,  // visible in help so users know it exists but is removed
       Short:  "[REMOVED in v3R3] Use 'moai hybrid glm' instead",
       Long:   "...detailed BC notice referencing SPEC-V3R3-HYBRID-001 §10...",
       RunE:   runCGRemoved,
   }
   func runCGRemoved(cmd, args) error {
       fmt.Fprintln(cmd.ErrOrStderr(), "Error: MOAI_CG_REMOVED")
       fmt.Fprintln(cmd.ErrOrStderr(), "  'moai cg' was removed in v3R3 per SPEC-V3R3-HYBRID-001.")
       fmt.Fprintln(cmd.ErrOrStderr(), "  Replacement: 'moai hybrid glm' (multi-LLM hybrid mode).")
       fmt.Fprintln(cmd.ErrOrStderr(), "  See SPEC-V3R3-HYBRID-001 §10 BC Migration for details.")
       return errors.New("MOAI_CG_REMOVED")
   }
   ```
3. Mirror to `internal/template/templates/internal/cli/...` if applicable (cg.go and hybrid.go are Go source files, not template files; no embedded mirror needed).

Verification:
- `go test ./internal/cli/ -run TestCGCommand` → GREEN.
- `go test ./internal/cli/ -run "TestHybridHelp|TestHybridUnknownProvider|TestHybridProxyOverride"` → GREEN.

### M4: Generalize injection + launcher routing + migration (GREEN, part 3) — Priority P0

Owner role: `expert-backend`.

Scope:

1. **`internal/llm/inject.go`** (NEW): provider-agnostic helpers refactored from `internal/cli/glm.go`.
   - `InjectTmuxSessionEnv(provider Provider, apiKey string)` ← refactor `glm.go:356-385`.
   - `InjectProviderEnvForTeam(settingsPath string, provider Provider, apiKey string)` ← refactor `glm.go:520-581`. Provider-namespaced backup keys (`MOAI_BACKUP_AUTH_TOKEN_<UPPER>`).
   - `ClearTmuxSessionEnv(providers ...Provider)` ← refactor `glm.go:392-421`. Accepts list of providers to clear (typically just the active one).
   - `EnsureSettingsLocalJSONHybrid(settingsPath)` ← refactor `glm.go:443-482`. teammateMode="tmux" + clean legacy CLAUDE_CODE_TEAMMATE_DISPLAY.
2. **`internal/cli/glm.go`** (MODIFY): swap GLM-specific calls to use `internal/llm` helpers.
   - `runGLM`: replace inline GLM env build with `providers.GLMProvider{}.BuildEnvVars(apiKey)`.
   - `runGLM` line 148-149 messaging: substitute `moai cg` → `moai hybrid glm`.
   - `enableTeamMode(cmd, isHybrid bool)`: keep public API but delegate body to `enableHybridTeamMode(cmd, providers.GLMProvider{}, isHybrid)`.
   - **Preserve** `moai glm`, `moai glm setup`, `moai glm status` for REQ-HYBRID-006 (no behavioral change).
3. **`internal/cli/launcher.go:57`** (MODIFY): replace `case "claude_glm":` branch with `case "claude_hybrid":` — switch on `LLMConfig.Provider` field and call `providers.Lookup(provider)` to get current Provider; route accordingly.
4. **`internal/cli/launcher.go:185`** (MODIFY): `persistTeamMode(root, "cg")` → `persistTeamMode(root, "hybrid")` + separately persist `provider: "<name>"`.
5. **`internal/cli/update.go`** (MODIFY): add `migrateCGTeamMode(root)` helper.
   - Read `.moai/config/sections/llm.yaml`.
   - If `team_mode: "cg"` and `provider: ""`: rewrite to `team_mode: "hybrid"` + `provider: "glm"`. Atomic write (Same pattern as SPEC-V3R3-UPDATE-CLEANUP-001 atomic write).
   - Emit one-time stderr notice (research.md §7).
   - Invoke from `cleanMoaiManagedPaths` (`internal/cli/update.go:1411`) or equivalent post-cleanup hook.
6. Mirror project `.moai/config/sections/llm.yaml` to embedded template; run `make build`.

Verification:
- `go test ./internal/cli/ -run "TestHybrid"` (full suite) → GREEN.
- `go test ./internal/cli/ -run TestMigrateCGTeamMode` → GREEN.
- `go test ./...` → 0 cascading failures.

[HARD] M4 must NOT break existing `TestRunCC*`, `TestRunGLM*`, `TestEnableTeamMode*` tests in `internal/cli/glm_test.go` and `internal/cli/launcher_test.go`. The generalize refactor preserves observable behavior of `moai cc` / `moai glm` (REQ-HYBRID-006).

### M5: Documentation substitution + CHANGELOG + MX tags (REFACTOR + Trackable) — Priority P1

Owner role: `manager-docs` for documentation substitution, `expert-backend` for MX tag insertion in Go code.

Scope:

#### M5a: `.claude/skills/moai/team/glm.md` substitution

Targets:
- Line 41: `User runs: moai cg` → `User runs: moai hybrid glm`
- Line 50: `Saves team_mode: cg to llm.yaml` → `Saves team_mode: hybrid + provider: glm to llm.yaml`
- Line 103: `| cg | CG Mode | Claude (this pane) | GLM (new tmux panes) |` → `| hybrid (provider=glm) | Hybrid Mode | Claude (this pane) | GLM (new tmux panes) |`
- Line 127: `moai cg` → `moai hybrid glm`
- Line 143: `moai cg injects these into the tmux session:` → `moai hybrid <provider> injects these into the tmux session (example: provider=glm):`

Mirror to `internal/template/templates/.claude/skills/moai/team/glm.md`.

#### M5b: `.claude/skills/moai/team/run.md` substitution

Targets (lines 73, 88, 92, 103, 104): replace each `moai cg` mention with `moai hybrid <provider>` (example provider=glm). Update line 73 table row similarly.

Mirror to `internal/template/templates/`.

#### M5c: `.claude/skills/moai/workflows/run.md:904` substitution

Target: `active_mode: cc | glm | cg` → `active_mode: cc | glm | hybrid`.

Mirror to `internal/template/templates/`.

#### M5d: `.claude/rules/moai/development/model-policy.md:44` substitution

Target: `Activation: moai cg (requires tmux)` → `Activation: moai hybrid <provider> (requires tmux). Example: moai hybrid glm`.

Mirror to `internal/template/templates/`.

#### M5e: CLAUDE.md §15 substitution

Target: `## 15. Agent Teams (Experimental)` 섹션 내 `### CG Mode (Claude + GLM Cost Optimization)` 헤딩 (line ~496-510).
- 헤딩 변경: `### Hybrid Mode (Multi-LLM Cost Optimization)`.
- 본문 다중 provider 일반화 — diagram에 "Teammates (GLM/Kimi/DeepSeek/Qwen, new tmux panes)" 표기.
- Activation: `moai hybrid <provider>` 예시.

Note: CLAUDE.md is not embedded in template; project root only.

#### M5f: CLAUDE.local.md §16 substitution

Target line 129: `Modified by 'moai glm', 'moai cc', 'moai cg' commands at runtime` → `Modified by 'moai glm', 'moai cc', 'moai hybrid <provider>' commands at runtime`.

CLAUDE.local.md is local-only (not in template per CLAUDE.local.md §2 Local-Only Files).

#### M5g: CHANGELOG entry

Append to `## [Unreleased]`:

```markdown
### Breaking Changes

- **BC-V3R3-HYBRID-001 (SPEC-V3R3-HYBRID-001)**: Removed `moai cg` (Claude + GLM hybrid)
  and replaced it with `moai hybrid <provider>`, which now supports four team-LLM
  providers as first-class citizens: GLM (Z.AI), Kimi K2 (Moonshot AI), DeepSeek V4
  (DeepSeek), and Qwen3-Coder (Alibaba DashScope).

  - `moai cg` invocations after upgrade return `MOAI_CG_REMOVED` with an actionable
    migration suggestion (`use 'moai hybrid glm' instead`).
  - Existing projects with `team_mode: "cg"` in `.moai/config/sections/llm.yaml` are
    auto-migrated to `team_mode: "hybrid"` + `provider: "glm"` on the next `moai update`
    run.
  - `moai glm` (single-LLM all-GLM mode) and `moai cc` (Claude-only mode) remain
    unchanged.
  - See SPEC-V3R3-HYBRID-001 §10 BC Migration for full migration table.
```

#### M5h: MX tag insertion

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and §6 below.

#### M5i: Final verification

1. `cd /Users/goos/MoAI/moai-adk-go && make build` — regenerate embedded.go.
2. `go test ./...` — full suite, 0 cascading failures (CLAUDE.local.md §6 HARD).
3. Verify `grep -rn "moai cg" .claude/ CLAUDE.md CLAUDE.local.md` returns only M5e/M5f-substituted content (no residual).
4. Update `progress.md` with `run_complete_at` + `run_status: implementation-complete`.

[HARD] No new SPEC documents created in `.moai/specs/` or `.moai/reports/` during M5.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Milestone |
|---|---|---|---|
| `internal/cli/cg.go` | line 7-44 (cgCmd) + line 51-57 (runCG) | Replace with BC stub returning `MOAI_CG_REMOVED` | M3 |
| `internal/cli/glm.go` | line 148-149 | Substitute `moai cg` → `moai hybrid glm` in stderr message | M4 |
| `internal/cli/glm.go` | line 218-348 (enableTeamMode) | Generalize via `enableHybridTeamMode(cmd, provider, isHybrid)` delegate | M4 |
| `internal/cli/glm.go` | line 266 | `persistTeamMode(root, "cg")` → `persistTeamMode(root, "hybrid")` + `persistProvider(root, "glm")` | M4 |
| `internal/cli/glm.go` | line 356-385 (injectTmuxSessionEnv) | Move to `internal/llm/inject.go InjectTmuxSessionEnv(provider, apiKey)` | M4 |
| `internal/cli/glm.go` | line 520-581 (injectGLMEnvForTeam) | Move to `internal/llm/inject.go InjectProviderEnvForTeam(...)` with namespaced backup | M4 |
| `internal/cli/glm.go` | line 536-538 | Single backup slot → provider-namespaced (`MOAI_BACKUP_AUTH_TOKEN_<UPPER>`) | M4 |
| `internal/cli/glm.go` | line 646-655 (GLMConfigFromYAML) | Generalize to `ProviderConfig` in `internal/llm/provider.go` | M2 |
| `internal/cli/launcher.go` | line 40 | Comment: `claude_glm` → `claude_hybrid` rename | M4 |
| `internal/cli/launcher.go` | line 57 | `case "claude_glm":` → `case "claude_hybrid":` | M4 |
| `internal/cli/launcher.go` | line 185 | `persistTeamMode(root, "cg")` → `persistTeamMode(root, "hybrid")` | M4 |
| `internal/cli/launcher.go` | line 267 | `resetTeamModeForCC` ← rename or generalize comment to "switching to CC clears hybrid team mode" | M4 |
| `internal/cli/update.go` | `cleanMoaiManagedPaths` end (line ~1441) | Add `migrateCGTeamMode(root)` post-cleanup hook | M4 |
| `internal/config/types.go` | line 52-70 (LLMConfig) | Add `Provider string` + `Providers map[string]ProviderYAML` fields | M2 |
| `internal/config/types.go` | (after line 100) | Add `type ProviderYAML struct` | M2 |
| `internal/config/defaults.go` | NewDefaultLLMConfig | Add `NewDefaultProvidersConfig()` returning 4-provider map | M2 |
| `.moai/config/sections/llm.yaml` | top-level | Add `provider: ""` + `providers:` section | M2 |
| `internal/template/templates/.moai/config/sections/llm.yaml` | top-level | Mirror M2 changes | M2 |
| `.claude/skills/moai/team/glm.md` | lines 41, 50, 103, 127, 143 | Substitute `moai cg` → `moai hybrid glm` (5 mentions) | M5a |
| `.claude/skills/moai/team/run.md` | lines 73, 88, 92, 103, 104 | Substitute `moai cg` → `moai hybrid <provider>` (5 mentions) | M5b |
| `.claude/skills/moai/workflows/run.md` | line 904 | `active_mode: cc \| glm \| cg` → `active_mode: cc \| glm \| hybrid` | M5c |
| `.claude/rules/moai/development/model-policy.md` | line 44 | `Activation: moai cg` → `Activation: moai hybrid <provider>` | M5d |
| `CLAUDE.md` | §15 ~line 488-510 (CG Mode section) | Section title + content multi-provider generalization | M5e |
| `CLAUDE.local.md` | §16 line 129 | `'moai cg'` → `'moai hybrid <provider>'` | M5f |
| `CHANGELOG.md` | `## [Unreleased]` section | Add `BC-V3R3-HYBRID-001` entry | M5g |

### 3.2 To-be-created (new files)

| File | Reason | Milestone |
|---|---|---|
| `internal/llm/provider.go` | Provider abstraction interface | M2 |
| `internal/llm/inject.go` | Provider-agnostic env-injection helpers | M4 |
| `internal/llm/providers/glm.go` | GLM provider metadata | M2 |
| `internal/llm/providers/kimi.go` | Kimi K2 provider metadata | M2 |
| `internal/llm/providers/deepseek.go` | DeepSeek V4 provider metadata | M2 |
| `internal/llm/providers/qwen.go` | Qwen3-Coder provider metadata | M2 |
| `internal/llm/providers/registry.go` | Provider registry + List() + Lookup() | M2 |
| `internal/llm/registry_test.go` | Provider registry consistency unit tests | M2 |
| `internal/cli/hybrid.go` | `moai hybrid` command + subcommands (setup/status/tools) | M3 |
| `internal/template/provider_allowlist_audit_test.go` | REQ-HYBRID-002, 017 audit | M1 |
| `internal/cli/cg_removal_test.go` | REQ-HYBRID-003, 010, 014 BC verification | M1 |
| `internal/cli/hybrid_test.go` | REQ-HYBRID-001..009 + 012, 013, 015, 018 | M1 |
| `internal/cli/migration_test.go` | REQ-HYBRID-011 cg→hybrid migration | M1 |

### 3.3 NOT to be touched (preserved by reference)

- `internal/cli/glm.go:21-89` (glmCmd, glmSetupCmd, glmStatusCmd) — `moai glm` 단일 모드 보존 per REQ-HYBRID-006.
- `internal/cli/glm.go:107-152` `runGLM` — single-LLM 모드 본문 보존 (line 148-149 message만 substitution).
- `internal/cli/cc.go` (확인 필요; 존재 시 보존) — `moai cc` Claude-only 모드 보존.
- `~/.claude/` 인증 흐름 — Claude OAuth 토큰 영향 없음.
- 16-language neutrality contract (CLAUDE.local.md §15) — 4-provider neutrality 차용 적용은 본 SPEC.

### 3.4 Reference citations (file:line)

본 plan에서 인용된 load-bearing anchors (research.md §10 + 추가):

1. `internal/cli/cg.go:1-58` — 현행 cg 명령 (M3 replace).
2. `internal/cli/glm.go:21-89` — glmCmd 보존 영역.
3. `internal/cli/glm.go:107-152` — runGLM 본문 (line 148-149 substitution).
4. `internal/cli/glm.go:218-348` — enableTeamMode generalize 대상 (M4).
5. `internal/cli/glm.go:356-385` — injectTmuxSessionEnv (M4 move).
6. `internal/cli/glm.go:520-581` — injectGLMEnvForTeam (M4 move).
7. `internal/cli/glm.go:646-724` — GLMConfigFromYAML + loadGLMConfig (M2 generalize).
8. `internal/cli/launcher.go:21-26` — unifiedLaunch indirection.
9. `internal/cli/launcher.go:38-39` — @MX:ANCHOR fan_in=3.
10. `internal/cli/launcher.go:57` — `case "claude_glm":` (M4 rename).
11. `internal/cli/launcher.go:185` — persistTeamMode("cg") (M4).
12. `internal/config/types.go:52-100` — LLMConfig + GLMModels (M2 extend).
13. `.moai/config/sections/llm.yaml` — current schema baseline.
14. `internal/template/templates/.moai/config/sections/llm.yaml` — embedded template baseline.
15. `.claude/skills/moai/team/glm.md:41,50,103,127,143` — M5a 5 substitution sites.
16. `.claude/skills/moai/team/run.md:73,88,92,103,104` — M5b 5 substitution sites.
17. `.claude/skills/moai/workflows/run.md:904` — M5c.
18. `.claude/rules/moai/development/model-policy.md:44` — M5d.
19. `CLAUDE.md §15 line 488-510 (CG Mode)` — M5e.
20. `CLAUDE.local.md §16 line 129` — M5f.
21. `internal/template/lang_boundary_audit_test.go:1-60` (SPEC-V3R2-WF-005 M1 scaffold) — provider_allowlist_audit_test.go 모델.
22. SPEC-GLM-MCP-001 PR #769 — REQ-HYBRID-016 dependency.
23. SPEC-V3R3-UPDATE-CLEANUP-001 PR #764 — atomic write 인프라 (REQ-HYBRID-011).

Total: **23 distinct file:line / SPEC anchors** (>10 minimum).

---

## 4. Technology Stack Constraints

본 SPEC은 **신규 외부 의존성을 추가하지 않는다**:

- 신규 Go modules: 없음. Anthropic-compat HTTP는 `net/http` 표준 라이브러리만 사용.
- LLM provider별 native SDK (e.g., `github.com/sashabaranov/go-openai`): **금지** (spec.md §1.2 Non-Goals).
- 신규 directory 구조: `internal/llm/` + `internal/llm/providers/` (단일 신규 패키지 트리).
- 신규 binary 의존: 없음.

추가 surface (additive):
- `internal/llm/` 패키지 (~5 신규 파일).
- `internal/cli/hybrid.go` (1 신규 파일).
- `~/.moai/.env.{kimi,deepseek,qwen}` 파일 패턴 (기존 `.env.glm`과 동일 dotenv 형식).
- 4 신규 environment variable name (provider별): `KIMI_API_KEY`, `DEEPSEEK_API_KEY`, `DASHSCOPE_API_KEY` (`GLM_API_KEY`는 기존).

---

## 5. Risk Analysis & Mitigations (file-anchored)

spec.md §8 risks의 file-anchored mitigations + 추가 plan-level risks.

| Risk | Probability | Impact | Mitigation Anchor (file:line) |
|---|---|---|---|
| `moai cg` 사용자 워크플로 중단 | M | M | `internal/cli/cg.go` BC stub (M3) — `runCGRemoved` 함수에서 `MOAI_CG_REMOVED` + actionable message. Cobra "unknown command" fallback 차단. CHANGELOG `BC-V3R3-HYBRID-001` 명시. |
| Kimi/Qwen3 endpoint 미가용 | M | M | M2 `internal/llm/providers/{kimi,qwen}.go` 후보 base_url 등록 + M3 `internal/cli/hybrid.go` `--proxy <url>` flag 처리 (REQ-HYBRID-015). Run-phase에서 사용자 검증 후 base_url update. |
| Latest model 변경 (GLM-4.8, Kimi K3 출시 등) | M | H | M2 `internal/config/types.go ProvidersConfig` + `.moai/config/sections/llm.yaml providers.<name>.models` user override 경로. SPEC-pinned baseline은 init/update 시 적용. |
| `team_mode: "cg"` 마이그레이션 실패 | L | M | M4 `internal/cli/update.go migrateCGTeamMode` (atomic write per SPEC-V3R3-UPDATE-CLEANUP-001 인프라) + M3 `internal/cli/cg.go` BC stub로 직접 호출 차단 — 양방향 안전망. M1 `migration_test.go`에서 idempotency 검증. |
| 단일 백업 슬롯 충돌 (provider 전환 시) | M | M | M4 `internal/llm/inject.go InjectProviderEnvForTeam` provider-namespaced backup keys (`MOAI_BACKUP_AUTH_TOKEN_<UPPER>`). 단일 슬롯 패턴 deprecate. |
| Template-First mirror 누락 | H | L | M2/M4 종료 시 `make build` HARD rule. CI에서 embedded FS 동기화 검증 (existing pattern). M5e/M5f는 template-out-of-scope (CLAUDE.md/CLAUDE.local.md). |
| Anthropic-compat endpoint cryptic error | M | M | M3 `internal/cli/hybrid.go` launch 직전 health check 옵션 (REQ-HYBRID-013) — `HEAD <base_url>` 실패 시 underlying HTTP error propagate. |
| Provider tier 매핑 의미 상이 | L | H | M2 `.moai/config/sections/llm.yaml providers.<name>.models` 사용자 직접 override 권장. init/update 시 provider docs 링크 포함 stderr notice. |
| BC stub fallback (cobra "unknown command") | L | M | M3 `internal/cli/cg.go` cgCmd registration 보존 (binary stub) — REQ-HYBRID-014 명시. M1 `cg_removal_test.go TestCGCommandIsRegistered` 검증. |
| docs-site 4-locale 누락 | M | M | sync-phase manager-docs 의무 (CLAUDE.local.md §17). 본 SPEC plan은 docs-site 변경을 sync-phase로 위임 (run-phase는 코드 + .claude/.md 한정). |
| `moai cc`/`moai glm` 회귀 (M4 generalize 부작용) | M | H | M4 `internal/cli/glm.go` 보존 영역 (line 21-89, line 107-152) 명시. existing tests `glm_test.go`, `launcher_test.go` 통과 확인. M1 RED gate에서 기존 테스트 GREEN 유지 검증. |
| `provider`+`team_mode` 두 필드의 일관성 | L | M | M4 `persistTeamMode` + `persistProvider` atomic 호출 — 중간 상태 노출 방지. config 검증 `validateLLMConfig` 추가 검토 (`team_mode: "hybrid"` 시 `provider != ""` 강제). |
| GLM-MCP-001 implementation 미완료 시 `moai hybrid glm tools` UX | L | L | M3 `internal/cli/hybrid.go hybridToolsCmd` stub은 명시적 "deferred to SPEC-GLM-MCP-001" 메시지. 사용자 혼동 방지. |
| 4 provider model name 변경 시 SPEC 갱신 부담 | L | M | research.md §3 verification table 매년 review. 사용자 override 경로가 stable한 escape hatch. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|---|---|---|
| `internal/llm/providers/registry.go Registry` | `@MX:ANCHOR fan_in=N — SPEC-V3R3-HYBRID-001 REQ-HYBRID-002,017 enforcer; closed allow-list of 4 providers (glm, kimi, deepseek, qwen). Adding a 5th entry requires atomic-reversal SPEC.` | Registry는 4-provider allow-list의 single source of truth. 모든 hybrid 명령 라우팅이 이 map을 lookup하므로 fan_in 매우 높음. |
| `internal/cli/hybrid.go hybridCmd` | `@MX:ANCHOR fan_in=N — SPEC-V3R3-HYBRID-001 REQ-HYBRID-001 canonical entry point for multi-LLM hybrid mode. Cobra command def; subcommand routing for setup/status/tools.` | hybridCmd는 사용자가 마주하는 surface; 변경 시 모든 hybrid 워크플로 영향. |
| `internal/llm/inject.go InjectProviderEnvForTeam` | `@MX:ANCHOR fan_in=N — SPEC-V3R3-HYBRID-001 REQ-HYBRID-007,012 provider-agnostic env-injection contract. Touching this signature affects all 4 providers.` | env-injection은 모든 provider에서 사용; signature 변경 시 4 provider 모두 영향. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|---|---|---|
| `internal/cli/cg.go runCGRemoved` | `@MX:NOTE — BC stub for SPEC-V3R3-HYBRID-001. Returns MOAI_CG_REMOVED with actionable migration path. Do not silently delete cgCmd registration; cobra fallback is insufficient.` | BC stub의 의도 명시 — 미래 PR이 cg.go를 완전 삭제하려는 시도 차단. |
| `internal/llm/provider.go Provider interface` | `@MX:NOTE — Provider abstraction per SPEC-V3R3-HYBRID-001. Anthropic-compat HTTP only (no native SDKs). Each provider supplies BuildEnvVars implementing its own compat flags (e.g., DISABLE_PROMPT_CACHING=1 for Z.AI).` | interface design 의도 명시 — 미래 PR이 native SDK 의존성 추가 시도 차단. |
| `internal/cli/update.go migrateCGTeamMode` | `@MX:NOTE — One-time migration for SPEC-V3R3-HYBRID-001. Idempotent; emits stderr notice only on actual rewrite. Uses SPEC-V3R3-UPDATE-CLEANUP-001 atomic write infrastructure.` | 마이그레이션 의도 + idempotency 보장 명시. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|---|---|---|
| `internal/llm/inject.go InjectProviderEnvForTeam (backup namespacing)` | `@MX:WARN @MX:REASON — Provider transition (e.g., GLM→Kimi) MUST use namespaced backup keys (MOAI_BACKUP_AUTH_TOKEN_<UPPER>). Single MOAI_BACKUP_AUTH_TOKEN slot causes data loss across multi-provider transitions. Reverting to single slot fails TestHybridBackwardCompatWithCG.` | 단일 슬롯 회귀 방지. |
| `internal/cli/launcher.go:57 case "claude_hybrid"` | `@MX:WARN @MX:REASON — Renaming this case label or the underlying mode string requires updating both unifiedLaunch and persistTeamMode call sites synchronously. Single-site rename leaves routing inconsistent — TestHybridProviderRouting fails.` | 두 곳 동시 변경 필수성 명시. |
| `internal/llm/providers/{kimi,qwen}.go base_url` | `@MX:WARN @MX:REASON — Anthropic-compat base URL pending provider-side verification (research.md §3.2, §3.4). Users should prefer --proxy <url> until provider docs confirm. Hardcoding without verification risks runtime auth failure.` | 미검증 endpoint 사용자 인지. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC's run phase)

본 SPEC은 plan-phase 산출물 + multi-milestone 구현이지만 모든 work는 M1-M5 GREEN 내에서 수렴한다. 임시 `@MX:TODO`는 GREEN 단계에서 모두 해결 (spec-workflow.md GREEN-phase resolution rule).

예외: M3 `internal/cli/hybrid.go hybridToolsCmd`의 stub은 `@MX:TODO` 표기 가능 (`// @MX:TODO: implement Z.AI MCP attach via SPEC-GLM-MCP-001 follow-up`) — 이는 의도적 deferral이며 SPEC-GLM-MCP-001 implementation PR 머지 시 해결.

### 6.5 MX tag count summary

- @MX:ANCHOR: 3 targets
- @MX:NOTE: 3 targets
- @MX:WARN: 3 targets
- @MX:TODO: 1 target (intentional deferral, linked to SPEC-GLM-MCP-001)
- **Total**: 10 MX tag insertions across ~7 distinct files

---

## 7. Solo Mode Discipline

[HARD] All run-phase work for SPEC-V3R3-HYBRID-001 executes in:

```
/Users/goos/MoAI/moai-adk-go
```

Branch: `feature/SPEC-V3R3-HYBRID-001` (already checked out per session context).

[HARD] No worktree is used (per user directive: solo mode, no worktree). All Read/Write/Edit tool invocations use absolute paths under the main project root.

[HARD] `make build` and `go test ./...` execute from the repo root: `cd /Users/goos/MoAI/moai-adk-go && make build && go test ./...`.

[HARD] CHANGELOG entry MUST cite `BC-V3R3-HYBRID-001` ID exactly (per CLAUDE.local.md §18 BC tracking convention) and reference SPEC-V3R3-HYBRID-001 §10 for migration table.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md` Phase 0.5 description). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.2.0 schema** — `spec.md` frontmatter has all 9 required fields (`id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`). Plus optional fields (`phase`, `module`, `dependencies`, `related_theme`, `breaking`, `bc_id`, `lifecycle`, `tags`).
- [x] **C2: HISTORY entry for v0.1.0** — `spec.md:30-34` HISTORY table has v0.1.0 row with description.
- [x] **C3: 18 EARS REQs across 5 categories** — `spec.md` §5 (Ubiquitous 6, Event-Driven 5, State-Driven 3, Optional 2, Complex 2).
- [x] **C4: 18 ACs all map to REQs (100% coverage)** — `spec.md` §6, plan §1.4 traceability matrix confirms 18/18 REQ → AC mapping.
- [x] **C5: BC scope clarity** — `spec.md` §10 BC Migration section + `breaking: true` + `bc_id: [BC-V3R3-HYBRID-001]` frontmatter.
- [x] **C6: File:line anchors ≥10** — research.md §10 (33 anchors), plan.md §3.4 (23 anchors).
- [x] **C7: Exclusions section present** — `spec.md` §1.2 Non-Goals + `spec.md` §2.2 Out of Scope (10 entries each).
- [x] **C8: TDD methodology declared** — plan §1.2 + `.moai/config/sections/quality.yaml development_mode: tdd`.
- [x] **C9: mx_plan section** — plan §6 (10 MX tag insertions across 4 categories).
- [x] **C10: Risk table with mitigations** — `spec.md` §8 (10 risks) + plan §5 (14 risks, file-anchored mitigations).
- [x] **C11: Solo mode path discipline** — plan §7 (4 HARD rules, no worktree per user directive).
- [x] **C12: No implementation code in plan documents** — verified self-check: spec.md, plan.md, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, type signatures, YAML/Markdown templates. No Go function bodies (only signatures + docstrings as illustration).
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md §1-18.
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md (M1-M5 with expert-backend / manager-tdd / manager-docs assignments).
- [x] **C15: Cross-SPEC consistency** — SPEC-GLM-MCP-001 (PR #769 OPEN, plan-only) declared in `dependencies:` (forward-looking surface); SPEC-V3R3-UPDATE-CLEANUP-001 (PR #764 merged) infrastructure reused (research.md §7); SPEC-V3R2-WF-005 (PR #768 merged) neutrality pattern applied.
- [x] **C16: BC migration completeness** — spec.md §10 covers (a) BC ID, (b) migration table v2.x→v3R3, (c) CHANGELOG wording, (d) docs touch points, (e) removal timeline.
- [x] **C17: 4-LLM allow-list documented** — spec.md §1, §2.2, §5 REQ-HYBRID-002, plan §1.3 deliverable list, plan §6.1 ANCHOR. 5번째 provider 차단 확인 (REQ-HYBRID-017 + AC-HYBRID-15).
- [x] **C18: Endpoint verification status documented** — research.md §3 + §3.5 summary table. Verified (GLM, DeepSeek) vs Pending (Kimi, Qwen3) 명시.

All 18 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: Create 4 test files (`provider_allowlist_audit_test.go`, `cg_removal_test.go`, `hybrid_test.go`, `migration_test.go`). Confirm RED for all NEW tests; existing tests (`glm_test.go`, `launcher_test.go`) GREEN unaffected.
2. **M2 (P0)**: Create `internal/llm/{provider.go, providers/{glm,kimi,deepseek,qwen,registry}.go, registry_test.go}` + extend `internal/config/types.go` LLMConfig + `.moai/config/sections/llm.yaml` schema + embedded template mirror + `make build`. Confirm `TestProviderAllowlist` GREEN.
3. **M3 (P0)**: Create `internal/cli/hybrid.go` + replace `internal/cli/cg.go` with BC stub. Confirm `TestCGCommand*` + basic `TestHybrid*` GREEN.
4. **M4 (P0)**: Create `internal/llm/inject.go` + generalize `internal/cli/glm.go` callsites + update `internal/cli/launcher.go:57` + add `internal/cli/update.go migrateCGTeamMode` + final `make build`. Confirm full `TestHybrid*` + `TestMigrate*` GREEN. Confirm `go test ./...` 0 cascading failures.
5. **M5 (P1)**: Substitute 6 documentation files (M5a-d) + CLAUDE.md §15 (M5e) + CLAUDE.local.md §16 (M5f) + CHANGELOG `BC-V3R3-HYBRID-001` (M5g) + insert 10 MX tags (M5h) + final `make build` + full `go test ./...` (M5i). Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`.

Total milestones: 5. Total file edits (existing): ~25. Total file creations (new): ~13. Total CHANGELOG entries: 1.

---

End of plan.md.
