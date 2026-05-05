# SPEC-V3R3-HYBRID-001 Research (Phase 0.5)

> Deep codebase + external endpoint research for `moai hybrid` multi-LLM mode (cg deprecation + GLM/Kimi/DeepSeek/Qwen3-Coder team support).
> Companion to `spec.md` v0.1.0 and `plan.md` v0.1.0.
> Solo mode, no worktree. Working directory: `/Users/goos/MoAI/moai-adk-go`. Branch: `feature/SPEC-V3R3-HYBRID-001`.

## HISTORY

| Version | Date       | Author              | Description                                                                              |
|---------|------------|---------------------|------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | Phase 0.5 deep research: cg.go + glm.go 코드 매핑, 4 LLM provider endpoint 검증, GLM-MCP-001 dependency surface, env-injection pattern. |

---

## 1. Research Objectives

본 research.md는 plan-auditor PASS criterion #5 ("research.md grounds decisions in actual codebase scan AND verifies 4 LLM endpoints via WebFetch")를 충족하기 위해 다음을 수행한다:

1. **현행 `moai cg` implementation 매핑** — `cg.go`, `glm.go`, `launcher.go`의 정확한 LOC + 함수 의존성 그래프 작성.
2. **4 LLM API endpoint 검증** — GLM/Kimi/DeepSeek/Qwen3-Coder 공식 docs WebFetch 결과를 기록하고 unverified 항목은 "사용자 검증 대기" 명시.
3. **Anthropic-compat vs OpenAI-compat 차이** — 4 provider별 Anthropic-compat 가용성 확인 (provider-side 정식 endpoint 또는 사용자 운영 proxy 필요 여부).
4. **기존 GLM env injection 패턴 분석** — `injectGLMEnvForTeam` (`glm.go:520`), `injectTmuxSessionEnv` (`glm.go:356`)을 provider-agnostic으로 일반화하는 데 필요한 변경 surface.
5. **SPEC-GLM-MCP-001 dependency surface** — 본 SPEC의 REQ-HYBRID-016 `tools enable|disable` subcommand가 GLM-MCP-001 implementation에 어떻게 의존하는지 명시.

---

## 2. `moai cg` Codebase Surface Map

### 2.1 `internal/cli/cg.go` (57 LOC, 정확)

```
internal/cli/cg.go:1-58
  - Lines 1-5: package cli + imports (`github.com/spf13/cobra`).
  - Lines 7-44: var cgCmd = &cobra.Command{ Use: "cg [-p profile]", ... } — single command def.
  - Lines 46-48: init() registers cgCmd to rootCmd.
  - Lines 51-57: func runCG(cmd, args) — calls parseProfileFlag then unifiedLaunch(profileName, "claude_glm", filteredArgs).
```

**관찰**: cg.go는 매우 얇은 래퍼다. 실제 cg 로직은 (a) `unifiedLaunch` `internal/cli/launcher.go:25` `case "claude_glm"`, (b) `enableTeamMode(cmd, isHybrid: true)` `internal/cli/glm.go:218` 두 군데에 분산되어 있다.

### 2.2 `internal/cli/launcher.go` (713 LOC) — `claude_glm` 분기

```
internal/cli/launcher.go:21-280 (relevant slice)
  - Line 21: var unifiedLaunchFunc = unifiedLaunchDefault — testability indirection.
  - Line 24-26: func unifiedLaunch(profileName, modeOverride, extraArgs) — public entry.
  - Line 38-39: @MX:ANCHOR fan_in=3 (called from runCC, runCG, runGLM).
  - Line 40-41: func unifiedLaunchDefault — centralized launch.
  - Line 57: case "claude_glm": — cg mode-specific branch.
  - Line 185: persistTeamMode(root, "cg") — writes `team_mode: "cg"` to llm.yaml.
  - Line 267: resetTeamModeForCC — disables team_mode when switching to CC.
```

**관찰**: `claude_glm` 문자열 리터럴은 `launcher.go:57` 분기 + `glm.go:185` (`persistTeamMode(root, "cg")`) 등 **두 곳**에 등장한다. M3에서 두 곳을 동시에 `claude_hybrid` + `provider: "glm"`로 교체해야 한다 (단일-위치 변경 시 라우팅이 깨진다).

### 2.3 `internal/cli/glm.go` (943 LOC) — 변경 surface

```
internal/cli/glm.go:21-943 (relevant slices)
  - Line 21-60: var glmCmd — moai glm 명령. 본 SPEC에서 보존(REQ-HYBRID-006).
  - Line 62-67: var glmSetupCmd — moai glm setup. 보존 (alias로 유지 가능).
  - Line 69-82: var glmStatusCmd — moai glm status. 보존.
  - Line 84-89: init() — glmCmd.AddCommand + rootCmd.AddCommand. 보존.
  - Line 92-104: type SettingsLocal struct. 보존 (provider-agnostic).
  - Line 107-152: func runGLM. 보존; 단 line 148-149 messaging("for hybrid mode... use moai cg")는 "moai hybrid glm"으로 교체.
  - Line 154-165: func containsPermissionMode. 보존.
  - Line 167-179: func setGLMEnv. **provider-agnostic으로 generalize 후보** — `setProviderEnv(provider Provider, apiKey string)` 시그니처로 변경.
  - Line 182-205: func runGLMSetup. setup logic provider-agnostic으로 일반화 (env-key 이름만 provider별 차별).
  - Line 208-213: func maskAPIKey. 보존 (provider-agnostic 이미).
  - Line 218-348: func enableTeamMode(cmd, isHybrid bool). **핵심 generalize 대상** — `isHybrid bool` → `provider Provider` 파라미터 + `mode` 결정 로직 일반화.
  - Line 350-385: func injectTmuxSessionEnv(glmConfig, apiKey). **generalize 대상** — `glmConfig *GLMConfigFromYAML` → provider-agnostic config struct.
  - Line 387-421: func clearTmuxSessionEnv. 보존 (env-var 이름은 provider별 차별; generalize 시 list가 provider별 confg).
  - Line 423-438: func persistTeamMode(projectRoot, mode). 보존 + line 185 `"cg"` → `"hybrid"`.
  - Line 440-482: func ensureSettingsLocalJSON. 보존 (teammateMode: "tmux"는 provider 무관).
  - Line 485-505: func loadLLMSectionOnly. 보존 + 신규 `provider` 필드 처리.
  - Line 507-510: func disableTeamMode. 보존.
  - Line 512-581: func injectGLMEnvForTeam. **generalize 대상** — `injectProviderEnvForTeam(settingsPath, provider, apiKey, baseURL, models)`.
  - Line 583-644: func saveLLMSection. **provider 필드 추가**.
  - Line 646-655: type GLMConfigFromYAML. **provider-agnostic struct로 generalize** — `type ProviderConfig struct { BaseURL, Models { High, Medium, Low }, EnvVar string }`.
  - Line 657-686: func resolveGLMModels. **generalize** — `resolveProviderModels(provider, models) (high, medium, low)`.
  - Line 688-724: func loadGLMConfig. **generalize** — `loadProviderConfig(root, providerName) (*ProviderConfig, error)`.
  - Line 726-733: func getGLMEnvPath. **generalize** — `getProviderEnvPath(name) string` returns `~/.moai/.env.<name>`.
  - Line 735-753: func saveGLMKey. **generalize** — `saveProviderKey(provider, key)`.
  - Line 755-791: func loadGLMKey. **generalize** — `loadProviderKey(provider) string`.
  - Line 793-807: func escapeDotenvValue / unescapeDotenvValue. 보존.
  - Line 809-815: func getGLMAPIKey. **generalize** — `getProviderAPIKey(provider) string`.
  - Line 817-831: func buildGLMEnvVars. **generalize** — `buildProviderEnvVars(provider, apiKey) map[string]string`.
  - Line 833-890: func injectGLMEnv. **generalize** — `injectProviderEnv(settingsPath, provider, apiKey)`.
  - Line 893-905: func isTestEnvironment. 보존.
  - Line 911-943: func findProjectRoot. 보존.
```

**핵심 결론**: `glm.go`는 943 LOC이지만, 이 중 약 60% (~570 LOC)가 provider-agnostic으로 generalize 가능한 함수들이다. M2-M4에서 `internal/llm/provider.go` (provider abstraction interface) + `internal/llm/providers/{glm,kimi,deepseek,qwen}.go` (4 메타데이터 파일)을 신규 도입하고, glm.go의 GLM-specific 함수들을 위 신규 패키지로 이주(refactor + rename)한다.

### 2.4 `.moai/config/sections/llm.yaml` 현행 스키마

```yaml
# .moai/config/sections/llm.yaml (현행 v2.x)
llm:
  mode: ""                       # "" or "glm"
  team_mode: ""                  # "" or "claude" or "glm" or "cg" or "hybrid"
  glm_env_var: GLM_API_KEY        # GLM-specific
  performance_tier: ""
  claude_models: { high, medium, low }
  glm: { base_url, models: { high, medium, low, opus, sonnet, haiku } }
  default_model, quality_model, speed_model
```

**v3R3 확장 (REQ-HYBRID-004)**:

```yaml
llm:
  mode: ""
  team_mode: ""                              # 기존 "cg" → "hybrid"로 마이그레이션
  provider: ""                                # 신규: "" or "glm" or "kimi" or "deepseek" or "qwen"
  performance_tier: ""
  claude_models: { high, medium, low }
  providers:                                  # 신규: provider별 메타데이터 mapping
    glm:
      base_url: "https://api.z.ai/api/anthropic"
      env_var: "GLM_API_KEY"
      models: { high: "glm-4.7", medium: "glm-4.7", low: "glm-4.5-air" }
    kimi:
      base_url: "https://api.moonshot.ai/anthropic"  # provider-side 검증 대기 (§3.2)
      env_var: "KIMI_API_KEY"
      models: { high: "kimi-k2.6", medium: "kimi-k2.6", low: "kimi-k2-flash" }
    deepseek:
      base_url: "https://api.deepseek.com/anthropic"  # verified 2026-05-04
      env_var: "DEEPSEEK_API_KEY"
      models: { high: "deepseek-v4-pro", medium: "deepseek-v4-pro", low: "deepseek-v4-flash" }
    qwen:
      base_url: "https://dashscope.aliyuncs.com/compatible-mode/anthropic"  # 검증 대기 (§3.4)
      env_var: "DASHSCOPE_API_KEY"
      models: { high: "qwen3-coder-plus", medium: "qwen3-coder-plus", low: "qwen3-coder-flash" }
  glm:                                        # 기존 glm 섹션 deprecate (1년 alias 유지)
    base_url, models, ...
  default_model, quality_model, speed_model
```

호환성: 기존 `glm:` 섹션은 v3R3 동안 `providers.glm:`의 alias로 1년간 유지 (REQ-HYBRID-011 마이그레이션 + alias).

---

## 3. 4 LLM API Endpoint Verification (WebFetch 결과)

### 3.1 GLM (Z.AI / Zhipu AI) — VERIFIED

- **Source**: <https://docs.z.ai/scenario-example/develop-tools/claude>
- **WebFetch date**: 2026-05-04
- **Anthropic-compat base URL**: `"https://api.z.ai/api/anthropic"` ✅
- **Latest stable models** (2026-05-04 기준):
  - High tier (Opus equivalent): `"GLM-4.7"`
  - Medium tier (Sonnet equivalent): `"GLM-4.7"`
  - Low tier (Haiku equivalent): `"GLM-4.5-Air"`
- **Required env-vars**:
  - `ANTHROPIC_AUTH_TOKEN` (Z.AI API key)
  - `ANTHROPIC_BASE_URL = "https://api.z.ai/api/anthropic"`
  - `API_TIMEOUT_MS = "3000000"` (extended operations)
  - Optional: `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`
- **참고**: 사용자 요청에서 "GLM-4.6"으로 언급되었으나 Z.AI 공식 docs (2026-05-04 검증)는 GLM-4.7을 latest stable로 표시한다. SPEC-pinned baseline은 docs 기준 GLM-4.7로 채택; GLM-4.6 호환성은 사용자 override 경로(`provider.glm.models.high`)로 보장.
- **Compat 1급**: provider-side에서 Anthropic-compat 1급 지원 → `--proxy` 불필요.

### 3.2 Kimi K2 (Moonshot AI) — PENDING VERIFICATION

- **Source 1**: <https://platform.moonshot.ai/docs/api/chat> → 301 redirect to `https://platform.kimi.ai/docs/api/chat` (도메인 마이그레이션, 2026 Q1 추정)
- **Source 2**: <https://platform.moonshot.ai/docs/guide/use-anthropic-sdk-to-call-kimi-api> → 301 redirect to `https://platform.kimi.ai/docs/guide/use-anthropic-sdk-to-call-kimi-api` → 404
- **WebFetch date**: 2026-05-04
- **OpenAI-compat base URL**: `"https://api.moonshot.ai/v1"` (verified)
- **Anthropic-compat base URL**: **NOT FOUND in retrievable docs** ⚠️
  - 후보 1 (추정): `"https://api.moonshot.ai/anthropic"` — provider 패턴(`api.<host>/anthropic`)이 GLM/DeepSeek와 일치하나 docs 미확인.
  - 후보 2: 사용자 운영 proxy (예: <https://github.com/sigoden/aichat>를 proxy mode로 운영).
- **Latest stable model identifier**: `"kimi-k2.6"` (verified via api/chat docs)
- **결론**: Provider 등록은 1급으로 유지(REQ-HYBRID-002에 포함)하되, base_url은 (a) 사용자 검증 후 SPEC update, (b) `--proxy` 옵션 (REQ-HYBRID-015)으로 fallback. Run-phase에서 사용자 검증 단계 필수.

### 3.3 DeepSeek V4 — VERIFIED (with deprecation note)

- **Source**: <https://api-docs.deepseek.com/>
- **WebFetch date**: 2026-05-04
- **Anthropic-compat base URL**: `"https://api.deepseek.com/anthropic"` ✅
- **OpenAI-compat base URL**: `"https://api.deepseek.com"`
- **Latest stable models**:
  - High tier: `"deepseek-v4-pro"` (current)
  - Low tier: `"deepseek-v4-flash"` (current)
  - Deprecated (예정 2026-07-24): `"deepseek-chat"`, `"deepseek-reasoner"` — SPEC pinned baseline에서 제외.
- **참고**: 사용자 요청에서 "DeepSeek V3.x"로 언급되었으나 공식 docs (2026-05-04)는 V4 시리즈를 latest stable로 표시한다. SPEC pinned baseline은 V4. V3 호환성은 사용자 override(`provider.deepseek.models`).
- **Compat 1급**: provider-side에서 Anthropic-compat 1급 지원 → `--proxy` 불필요.

### 3.4 Qwen3-Coder (Alibaba DashScope / Bailian) — PENDING VERIFICATION

- **Source 1**: <https://help.aliyun.com/zh/model-studio/developer-reference/use-anthropic-sdk-to-call-qwen> → 404
- **Source 2**: <https://help.aliyun.com/zh/model-studio/anthropic-api> → 404
- **Source 3**: <https://bailian.console.aliyun.com/?tab=doc#/doc/?type=model&url=2840914> → loading shell (콘솔 SPA, 정적 추출 불가)
- **WebFetch date**: 2026-05-04
- **Anthropic-compat base URL**: **NOT FOUND in retrievable docs** ⚠️
  - 후보 1 (DashScope 패턴 추정): `"https://dashscope.aliyuncs.com/compatible-mode/anthropic"` 또는 `"https://dashscope.aliyuncs.com/api/v1/anthropic"` — provider 패턴 일관성 가설.
  - 후보 2: 사용자 운영 proxy.
- **Latest stable model identifier**: 사용자 보고 "Qwen3-Coder Latest" — pinned baseline 후보 `"qwen3-coder-plus"` 또는 `"qwen3-coder-30a3b-instruct"`. Docs 미확인.
- **Mainland China endpoint**: `dashscope.cn-hangzhou.aliyuncs.com` (별도 SPEC out-of-scope per spec.md §2.2).
- **결론**: Provider 등록은 1급으로 유지하되, base_url + 정확한 model identifier는 사용자 검증 후 SPEC update. `--proxy` 옵션으로 fallback.

### 3.5 Verification Summary Table

| Provider | Anthropic-compat Status | base_url | Latest Model | Verification Date |
|----------|------------------------|----------|--------------|-------------------|
| GLM (Z.AI) | ✅ VERIFIED 1급 | `https://api.z.ai/api/anthropic` | GLM-4.7 / GLM-4.5-Air | 2026-05-04 |
| Kimi K2 (Moonshot) | ⚠️ PENDING | (후보) `https://api.moonshot.ai/anthropic` | kimi-k2.6 | 사용자 검증 대기 |
| DeepSeek V4 | ✅ VERIFIED 1급 | `https://api.deepseek.com/anthropic` | deepseek-v4-pro / -flash | 2026-05-04 |
| Qwen3-Coder | ⚠️ PENDING | (후보) `https://dashscope.aliyuncs.com/compatible-mode/anthropic` | qwen3-coder-plus | 사용자 검증 대기 |

**Run-phase 필수 단계**: M2 milestone 진입 전 사용자에게 Kimi/Qwen3 정확한 endpoint 확인 → spec.md REQ-HYBRID-005 baseline 갱신 → embedded template `.moai/config/sections/llm.yaml` 미러.

---

## 4. Anthropic-compat vs OpenAI-compat 차이 분석

### 4.1 Anthropic-compat이 필수인 이유

`moai hybrid <provider>`는 Claude Code teammates의 subagent context에서 동작한다. Claude Code는 내부적으로 Anthropic Messages API (`POST /v1/messages`) 형식으로 prompt + tool use + thinking blocks를 송수신한다. 이 형식은 OpenAI Chat Completions API (`POST /v1/chat/completions`)와 다음 차이가 있다:

- **Tool use 형식**: Anthropic은 `{"type": "tool_use", "id": "...", "name": "...", "input": {...}}` content block 사용; OpenAI는 `{"function_call": {...}}` 또는 `{"tool_calls": [...]}` 사용.
- **System prompt 위치**: Anthropic은 top-level `system: "..."` 필드; OpenAI는 `messages: [{"role": "system", ...}]` 첫 메시지.
- **Streaming format**: Anthropic은 `event: message_start` / `event: content_block_delta` 등 SSE event types 차별화; OpenAI는 `data: {...}` 단일 chunk 형식.
- **Thinking blocks (Claude-specific)**: Anthropic만 `{"type": "thinking", "thinking": "..."}` content block 지원.

따라서 `moai hybrid <provider>`가 Claude Code와 호환되려면 provider가 Anthropic Messages API 형식을 1급으로 지원해야 한다. OpenAI-compat만 제공하는 provider는 사용자가 운영하는 adapter (예: claude-code-router, aichat proxy mode)를 거쳐야 한다.

### 4.2 4 Provider별 Anthropic-compat 가용성

| Provider | Anthropic-compat 1급 | adapter/proxy 필요? |
|----------|----------------------|---------------------|
| GLM (Z.AI) | ✅ | 불필요 |
| Kimi K2 (Moonshot) | ⚠️ 검증 대기 | 가능성 있음 (adapter 필요 시 `--proxy` 옵션) |
| DeepSeek V4 | ✅ | 불필요 |
| Qwen3-Coder | ⚠️ 검증 대기 | 가능성 있음 (`--proxy`) |

### 4.3 `--proxy` 옵션 설계 근거

REQ-HYBRID-015는 `moai hybrid <provider> --proxy <url>` 옵션으로 사용자가 ANTHROPIC_BASE_URL을 override 할 수 있게 한다. 사용 사례:

- Kimi/Qwen3가 provider-side Anthropic-compat을 정식 출시하기 전, 사용자가 self-hosted adapter 운영 (예: <https://github.com/sigoden/aichat> proxy mode, claude-code-router).
- 기업 내부 LLM gateway에 Anthropic-compat shim이 추가된 환경.
- Mainland China region의 별도 endpoint를 임시 사용 (별도 SPEC 정식 출시 전).

`--proxy` 사용 시 SPEC-pinned base_url은 무시되고 사용자 URL이 사용된다. 단, `Authorization: Bearer <api-key>` 헤더 형식은 그대로 유지(provider-side 인증 전제).

---

## 5. 기존 GLM Env-Injection 패턴 분석

### 5.1 두 주입 경로 (현행 v2.x)

`moai cg` 모드는 두 단계 env-var 주입을 수행한다:

**경로 A — tmux session-level injection** (`internal/cli/glm.go:356-385 injectTmuxSessionEnv`):

```go
vars := map[string]string{
    "ANTHROPIC_AUTH_TOKEN":           apiKey,
    "ANTHROPIC_BASE_URL":             glmConfig.BaseURL,
    "ANTHROPIC_DEFAULT_OPUS_MODEL":   glmConfig.Models.High,
    "ANTHROPIC_DEFAULT_SONNET_MODEL": glmConfig.Models.Medium,
    "ANTHROPIC_DEFAULT_HAIKU_MODEL":  glmConfig.Models.Low,
    "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":   "1",
    "DISABLE_PROMPT_CACHING":                   "1",
    "API_TIMEOUT_MS":                            "3000000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
}
mgr.InjectEnv(context.Background(), vars)
```

→ tmux `set-environment -g`로 session-scoped env-var 주입. 새 pane이 spawn될 때 inherit됨 (CG mode의 핵심 architectural pattern).

**경로 B — settings.local.json injection** (`internal/cli/glm.go:520-581 injectGLMEnvForTeam`):

```go
settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
settings.Env["ANTHROPIC_BASE_URL"] = glmConfig.BaseURL
// ... (위와 동일 키들)
settings.TeammateMode = "tmux"
delete(settings.Env, "CLAUDE_CODE_TEAMMATE_DISPLAY")  // legacy cleanup
```

→ `.claude/settings.local.json`에 평문 기록. Claude Code 시작 시 inherit. CG 모드는 leader가 Claude를 사용해야 하므로 settings.local.json에서는 이 env-var가 비어있어야 한다 (cg 분기에서 `removeGLMEnv` 호출). 반면 GLM 단일 모드 (`moai glm`)는 settings.local.json에도 GLM env를 기록한다.

### 5.2 Generalize 후 (v3R3)

provider-agnostic으로 generalize한 후의 env-var 키 일반화:

```go
// internal/llm/provider.go (신규)
type Provider interface {
    Name() string                          // "glm" | "kimi" | "deepseek" | "qwen"
    AnthropicBaseURL() string              // e.g., "https://api.z.ai/api/anthropic"
    EnvKeyName() string                    // e.g., "GLM_API_KEY"
    DefaultModels() ModelSet               // High, Medium, Low
    BuildEnvVars(apiKey string) map[string]string  // ANTHROPIC_AUTH_TOKEN + base_url + models + provider-specific flags
}

// internal/llm/providers/glm.go (예시)
func (p GLMProvider) BuildEnvVars(apiKey string) map[string]string {
    return map[string]string{
        "ANTHROPIC_AUTH_TOKEN":           apiKey,
        "ANTHROPIC_BASE_URL":             p.AnthropicBaseURL(),
        "ANTHROPIC_DEFAULT_OPUS_MODEL":   p.DefaultModels().High,
        "ANTHROPIC_DEFAULT_SONNET_MODEL": p.DefaultModels().Medium,
        "ANTHROPIC_DEFAULT_HAIKU_MODEL":  p.DefaultModels().Low,
        // GLM-specific (Z.AI proxy compatibility):
        "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":   "1",
        "DISABLE_PROMPT_CACHING":                   "1",
        "API_TIMEOUT_MS":                            "3000000",
        "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    }
}
```

각 provider는 BuildEnvVars에서 자신의 compat 요구사항을 표현한다 (DeepSeek/Kimi/Qwen은 다른 flag set일 수 있음 — 실측 후 확정). 공통 키 5개(`ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_*`)는 모든 provider에서 동일하다.

### 5.3 `MOAI_BACKUP_AUTH_TOKEN` 백업 슬롯

현행 `injectGLMEnvForTeam` (`glm.go:536-538`)은 단일 백업 슬롯 `MOAI_BACKUP_AUTH_TOKEN`을 사용한다:

```go
if existing := settings.Env["ANTHROPIC_AUTH_TOKEN"]; existing != "" && existing != apiKey {
    settings.Env["MOAI_BACKUP_AUTH_TOKEN"] = existing
}
```

다중 provider 전환 시(GLM → Kimi → DeepSeek) 단일 슬롯은 마지막 백업만 보존한다. v3R3에서는 provider-별 namespace 분리:

```go
backupKey := fmt.Sprintf("MOAI_BACKUP_AUTH_TOKEN_%s", strings.ToUpper(provider.Name()))
if existing := settings.Env["ANTHROPIC_AUTH_TOKEN"]; existing != "" && existing != apiKey {
    settings.Env[backupKey] = existing  // e.g., MOAI_BACKUP_AUTH_TOKEN_GLM
}
```

→ provider 전환 시 이전 토큰 보존. Claude OAuth 토큰은 `~/.claude/`에 별도 저장되어 영향받지 않는다.

---

## 6. SPEC-GLM-MCP-001 Dependency Surface

### 6.1 GLM-MCP-001 SPEC 개요 (PR #769 OPEN, plan-only)

PR #769 metadata 추출 결과 (`gh pr view 769 --json body`):

- **Goal**: Z.AI Vision/WebSearch/WebReader MCP 서버 (`@z_ai/mcp-server`)를 `moai glm` 모드에 통합.
- **Subcommand**: `moai glm tools enable|disable` (Vision/WebSearch/WebReader 토글).
- **Plan PASS**: plan-auditor PASS (score 0.91, MP 0 defects).
- **Status**: plan-only PR. Implementation은 별도 SPEC 후속.

### 6.2 본 SPEC과의 Surface Overlap

본 SPEC의 REQ-HYBRID-016은 forward-looking surface로 `moai hybrid <provider> tools enable|disable`을 정의한다. 즉:

- `moai glm tools enable` ← GLM-MCP-001 implementation 책임 (구체 MCP attach 로직).
- `moai hybrid glm tools enable` ← 본 SPEC implementation 책임 (CLI surface) + GLM-MCP-001 위임 (실제 attach).

### 6.3 의존성 분리 결정

본 SPEC은 GLM-MCP-001을 absorbing하지 않는다 (CLAUDE.local.md의 "이전 SPEC absorption" 패턴 회피):

- 본 SPEC은 **CLI surface + 4-provider abstraction**만 제공.
- MCP attach/detach 로직은 GLM-MCP-001에 남는다.
- Wire-up: `internal/cli/hybrid.go`가 `moai hybrid glm tools enable`을 받으면 GLM-MCP-001의 implementation 함수(예: `internal/mcp/zai_attach.go EnableTools()`)를 호출한다.

이 분리로 본 SPEC은 GLM-MCP-001의 implementation timeline에 block되지 않는다 (REQ-HYBRID-016의 surface는 stub로 deliver, MCP attach는 GLM-MCP-001 후속 PR에서 채움).

### 6.4 Forward-looking Surface Stub

```bash
# v3R3 첫 minor release: tools subcommand surface 등록 + stub
$ moai hybrid glm tools enable
Error: tool integration is implemented in SPEC-GLM-MCP-001 (PR #769 + follow-up). When that SPEC ships, this command attaches Z.AI Vision/WebSearch/WebReader MCP servers.

# GLM-MCP-001 implementation 후속 PR 머지 시: stub → 실제 동작
$ moai hybrid glm tools enable
✓ Z.AI Vision MCP server attached
✓ Z.AI WebSearch MCP server attached
✓ Z.AI WebReader MCP server attached
```

본 SPEC의 acceptance.md AC-HYBRID-14는 surface 등록만 검증한다 (help text + stub error message).

---

## 7. 마이그레이션 시나리오 분석 (REQ-HYBRID-011)

기존 사용자 프로젝트의 `.moai/config/sections/llm.yaml`은 다음 4가지 상태일 수 있다:

| 상태 | team_mode | 마이그레이션 동작 |
|------|-----------|-------------------|
| State A | `""` (비활성) | 변경 없음 (provider 필드 신규 추가만) |
| State B | `"claude"` | 변경 없음 (Claude 단일 모드 유지) |
| State C | `"glm"` | 변경 없음 (단일 GLM 모드 유지; `moai glm` 보존) |
| State D | `"cg"` | **자동 마이그레이션**: `team_mode: "hybrid"` + `provider: "glm"` |

State D 마이그레이션은 `moai update` cleanup phase에서 수행된다. SPEC-V3R3-UPDATE-CLEANUP-001 (PR #764 merged)이 도입한 atomic write + manifest provenance 인프라를 그대로 활용한다 (별도 신규 인프라 필요 없음).

마이그레이션 발생 시 stderr 메시지 (1회만, post-update notice):

```
NOTICE: SPEC-V3R3-HYBRID-001 BC migration applied to .moai/config/sections/llm.yaml
  - team_mode: "cg" → "hybrid"
  - provider: (added) "glm"
  See SPEC-V3R3-HYBRID-001 §10 BC Migration for details.
  Replace 'moai cg' invocations with 'moai hybrid glm' going forward.
```

---

## 8. 구현 surface 추정 (LOC)

| 변경 카테고리 | 파일 수 | LOC 추정 |
|---|---|---|
| 신규 provider abstraction (`internal/llm/provider.go` + 4 providers) | 5 | ~400 LOC |
| `internal/cli/hybrid.go` (신규 명령) | 1 | ~250 LOC |
| `internal/cli/cg.go` BC stub 변경 | 1 | -57 LOC + ~30 LOC stub |
| `internal/cli/glm.go` generalize (provider-agnostic refactor) | 1 | -200 LOC (move to providers/glm.go) + ~50 LOC adjustments |
| `internal/cli/launcher.go` `claude_glm` → `claude_hybrid` 분기 | 1 | ~30 LOC |
| `internal/config/types.go` LLMConfig 확장 (`Provider` 필드 + `Providers map`) | 1 | ~40 LOC |
| `internal/config/defaults.go` 4 provider 기본값 | 1 | ~80 LOC |
| `.moai/config/sections/llm.yaml` 스키마 확장 (project + template) | 2 | ~50 LOC |
| Test surface (provider_test, hybrid_test, cg_removal_test, allowlist_test) | 4-5 | ~500 LOC |
| 6 .claude/skills + .claude/rules 문서 substitution | 6-10 | ~30 LOC |
| CLAUDE.md, CLAUDE.local.md substitution | 2 | ~20 LOC |
| CHANGELOG entry | 1 | ~20 LOC |
| **Total** | **~26 files** | **~1500 LOC (net)** |

이 중 ~80%는 abstraction 도입 + 신규 코드, ~20%는 substitution.

---

## 9. Risk Mitigation Anchors (codebase-grounded)

spec.md §8 risks의 file-anchored mitigations:

| Risk (spec.md §8) | Mitigation Anchor |
|---|---|
| `moai cg` 사용자 워크플로 중단 | `internal/cli/cg.go` BC stub (REQ-HYBRID-014) — `cgCmd.RunE`에서 `MOAI_CG_REMOVED` + actionable suggestion 반환. Cobra 자동 "unknown command" fallback 차단. |
| Kimi/Qwen3 endpoint 미가용 | `--proxy <url>` 옵션 (REQ-HYBRID-015), `internal/cli/hybrid.go`에서 `--proxy` flag 파싱 후 `ANTHROPIC_BASE_URL` override. `internal/llm/providers/kimi.go`의 `AnthropicBaseURL()` 반환값을 `--proxy`로 대체. |
| Latest model 변경 | `provider.<name>.models` 사용자 override 경로 (REQ-HYBRID-005), `internal/config/types.go`의 `ProvidersConfig` struct + `loadProviderConfig` (`internal/llm/provider.go`)에서 사용자 값 우선 |
| `team_mode: "cg"` 마이그레이션 실패 | REQ-HYBRID-011 + REQ-HYBRID-014 양방향 안전망. `internal/cli/update.go cleanMoaiManagedPaths`에 `migrateCGTeamMode` 추가 + cg.go BC stub로 직접 호출 차단. |
| 단일 백업 슬롯 충돌 | provider-namespaced backup keys (`MOAI_BACKUP_AUTH_TOKEN_<UPPER>`) — `internal/llm/provider.go`의 `injectProviderEnvForTeam`에서 구현. |
| Template-First mirror 누락 | M2/M3/M4 종료 시 `make build` HARD rule. `internal/template/lang_boundary_audit_test.go` 패턴(SPEC-V3R2-WF-005 M1)을 차용해 `provider_allowlist_audit_test.go`에서 embedded FS 검증. |
| Anthropic-compat endpoint 응답 cryptic error | launch 직전 health check (`HEAD <base_url>` 또는 GET `/v1/messages` with empty body) — REQ-HYBRID-013 launch path에서 HTTP error propagate. |
| Provider tier 매핑 의미 상이 | research.md §3에서 provider별 model 등급 명시 + 사용자 override 권장 메시지 (init/update 시점). |
| BC stub fallback | REQ-HYBRID-014 명시: `cgCmd`는 binary에 stub로 보존. `internal/cli/cg.go`는 빈 파일이 아닌 active stub. |
| docs-site 4-locale 누락 | sync-phase manager-docs 의무 + CLAUDE.local.md §17 강제. |

---

## 10. File:line Citations (load-bearing anchors, ≥10)

1. `internal/cli/cg.go:1-58` — 현행 `moai cg` 명령 정의 (얇은 래퍼).
2. `internal/cli/cg.go:7-44` — `cgCmd` cobra 정의.
3. `internal/cli/cg.go:51-57` — `runCG` → `unifiedLaunch(profileName, "claude_glm", filteredArgs)`.
4. `internal/cli/launcher.go:21-26` — `unifiedLaunch` indirection (testability).
5. `internal/cli/launcher.go:38-39` — `@MX:ANCHOR fan_in=3` (runCC, runCG, runGLM).
6. `internal/cli/launcher.go:57` — `case "claude_glm":` 분기 (cg-specific).
7. `internal/cli/launcher.go:185` — `persistTeamMode(root, "cg")` (llm.yaml에 cg 기록).
8. `internal/cli/launcher.go:267` — `resetTeamModeForCC` (cc 전환 시 team_mode 초기화).
9. `internal/cli/glm.go:21-89` — `glmCmd` + setup/status subcommand 정의 (보존).
10. `internal/cli/glm.go:148-149` — runGLM의 "for hybrid mode... use moai cg" messaging (`moai hybrid glm`로 substitution 대상).
11. `internal/cli/glm.go:218-348` — `enableTeamMode(cmd, isHybrid bool)` (provider-agnostic generalize 핵심 대상).
12. `internal/cli/glm.go:251-253` — `if isHybrid && !inTmux` tmux 검증 (REQ-HYBRID-008 패턴).
13. `internal/cli/glm.go:266` — `persistTeamMode(root, "cg")` (M3 substitution 대상).
14. `internal/cli/glm.go:356-385` — `injectTmuxSessionEnv` (provider-agnostic generalize 대상).
15. `internal/cli/glm.go:366-374` — Anthropic-compat env-var keys (provider 공통).
16. `internal/cli/glm.go:520-581` — `injectGLMEnvForTeam` (provider-agnostic generalize 대상).
17. `internal/cli/glm.go:536-538` — 단일 백업 슬롯 `MOAI_BACKUP_AUTH_TOKEN` (provider-namespaced로 변경 대상).
18. `internal/cli/glm.go:646-655` — `type GLMConfigFromYAML` (provider-agnostic struct로 generalize 대상).
19. `internal/cli/glm.go:688-724` — `loadGLMConfig` (provider-agnostic generalize 대상).
20. `internal/cli/glm.go:726-733` — `getGLMEnvPath` → `~/.moai/.env.glm` (4-provider env path 패턴 확장).
21. `internal/config/types.go:52-70` — `LLMConfig` struct (`Mode`, `TeamMode`, `GLMEnvVar`, `GLM`); `Provider` + `Providers` 필드 추가 대상.
22. `internal/config/types.go:92-100` — `GLMModels` struct (High/Medium/Low + legacy Opus/Sonnet/Haiku).
23. `.moai/config/sections/llm.yaml:1-19` — 현행 v2.x 스키마 (project layer).
24. `internal/template/templates/.moai/config/sections/llm.yaml:1-43` — embedded template 스키마 (Template-First).
25. `.claude/skills/moai/team/glm.md:41,50,103,127,143` — `moai cg` 5개 mention (substitution 대상).
26. `.claude/skills/moai/team/run.md:73,88,92,103,104` — `moai cg` 5개 mention (substitution 대상).
27. `.claude/skills/moai/workflows/run.md:904` — `active_mode: cc | glm | cg` (substitution 대상).
28. `.claude/rules/moai/development/model-policy.md:44` — `Activation: moai cg` (substitution 대상).
29. `CLAUDE.md §15 line 498` — `Activation: moai cg (requires tmux)` (CG Mode 섹션 제목 + 본문 substitution).
30. `CLAUDE.local.md:129` — `Modified by 'moai glm', 'moai cc', 'moai cg' commands at runtime` (사용자 가이드 substitution).
31. SPEC-GLM-MCP-001 PR #769 — Z.AI MCP 서버 통합 plan-only PR (의존성).
32. SPEC-V3R3-UPDATE-CLEANUP-001 PR #764 — `moai update` atomic write + manifest provenance (REQ-HYBRID-011 인프라).
33. SPEC-V3R2-WF-005 PR #768 — 16-language neutrality 패턴 (4-provider neutrality 차용).

Total: **33 distinct file:line / SPEC anchors** (>10 minimum required by plan-auditor).

---

## 11. Open Questions for plan-auditor / Run Phase

### OQ-1: Kimi K2 Anthropic-compat endpoint 정확한 URL

- **Status**: 사용자 검증 대기
- **Action**: M2 milestone 진입 전 사용자에게 (a) <https://platform.kimi.ai/> 공식 docs에서 Anthropic-compat URL 확인, (b) 미존재 시 `--proxy` 패턴 채택 confirm.
- **Fallback**: spec.md §3.2 base_url을 `--proxy` 필수로 표기.

### OQ-2: Qwen3-Coder Anthropic-compat endpoint 정확한 URL

- **Status**: 사용자 검증 대기
- **Action**: M2 진입 전 (a) DashScope 공식 docs 검증, (b) Bailian console에서 Anthropic-compat 항목 확인.
- **Fallback**: spec.md §3.4 동일.

### OQ-3: GLM-4.7 vs GLM-4.6 사용자 의도

- **Status**: 사용자가 "GLM-4.6"으로 지칭, Z.AI 공식 docs는 GLM-4.7이 latest stable.
- **Resolution**: research.md §3.1에 따라 SPEC pinned baseline은 docs 기준 GLM-4.7. 사용자 override는 `provider.glm.models.high`로 가능. spec.md §1.1 "background"에 명시.

### OQ-4: DeepSeek V4 vs V3.x 사용자 의도

- **Status**: 사용자가 "DeepSeek V3.x"로 지칭, 공식 docs는 V4가 current (V3 deprecation 2026-07-24 예정).
- **Resolution**: SPEC pinned baseline은 V4. V3 호환은 사용자 override.

### OQ-5: `moai cg` BC stub vs 완전 삭제

- **Status**: 사용자 명시 "clean break (immediate removal)".
- **Resolution**: REQ-HYBRID-014에 따라 BC stub로 보존(완전 삭제 아님). Cobra "unknown command" fallback이 actionable 메시지를 제공하지 못하므로 stub 보존이 사용자에게 더 친절한 clean break.

### OQ-6: `moai glm tools` vs `moai hybrid glm tools` 우선순위

- **Status**: REQ-HYBRID-016은 `moai hybrid <provider> tools` surface를 정의하나 SPEC-GLM-MCP-001은 `moai glm tools`만 plan에 포함.
- **Resolution**: 두 surface 모두 deliver. `moai glm tools`는 GLM-MCP-001 implementation의 1차 surface; `moai hybrid glm tools`는 본 SPEC의 multi-provider 일관성 surface. Run-phase에서 동일 함수에 위임.

### OQ-7: `moai hybrid setup <provider> <key>` vs 기존 `moai glm setup <key>`

- **Status**: 본 SPEC은 `moai hybrid setup <provider>`를 권장하나 `moai glm setup <key>`도 alias로 1년간 보존.
- **Resolution**: spec.md §10.2 마이그레이션 표 명시. `moai glm setup`은 `~/.moai/.env.glm`에 저장 (기존 동작); `moai hybrid setup glm`은 동일 동작. 4 provider 모두 `moai hybrid setup <name>` 패턴 통일.

### OQ-8: provider별 cache_control / prompt_caching 정책 차별

- **Status**: 현행 GLM은 `DISABLE_PROMPT_CACHING=1`. 다른 provider도 동일?
- **Resolution**: spec.md §2.2 Out of Scope — 현행 정책 유지 (모든 provider `DISABLE_PROMPT_CACHING=1`). 변경은 SPEC-GLM-MCP-001 후속 또는 별도 SPEC.

---

## 12. Verification Checklist (research.md self-audit)

- [x] **현행 `moai cg` LOC 매핑**: cg.go (57 LOC), launcher.go relevant slice (line 21-280), glm.go (943 LOC) all enumerated in §2.
- [x] **4 provider endpoint WebFetch 결과 기록**: §3.1-3.4 + §3.5 summary table.
- [x] **Anthropic-compat 가용성 분류**: GLM/DeepSeek 1급 verified; Kimi/Qwen3 검증 대기 + `--proxy` fallback.
- [x] **Latest model name pinned baseline**: GLM-4.7, Kimi K2.6, DeepSeek-V4-pro/-flash, Qwen3-Coder Plus(추정).
- [x] **Env-injection 패턴 분석**: §5에 두 경로 (tmux session-level + settings.local.json) + provider-agnostic generalize 설계.
- [x] **GLM-MCP-001 dependency surface**: §6에 명시 — 본 SPEC은 CLI surface, GLM-MCP-001은 MCP attach 로직.
- [x] **마이그레이션 시나리오 4가지 상태 (A-D)**: §7.
- [x] **LOC 추정 ~1500 (net)**: §8.
- [x] **Risk mitigation file-anchored**: §9에서 spec.md §8 risks 모두 anchor 매핑.
- [x] **File:line citations ≥10**: §10에서 33개 anchor.
- [x] **Open questions 8건**: §11.
- [x] **사용자 의도 vs Z.AI/DeepSeek docs 차이 (GLM-4.6 vs 4.7, V3 vs V4)**: §11 OQ-3, OQ-4에서 명시 + spec.md §1.1 background 반영.

---

End of research.md.
