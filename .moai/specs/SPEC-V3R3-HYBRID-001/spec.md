---
id: SPEC-V3R3-HYBRID-001
title: moai hybrid Multi-LLM Mode (cg deprecation + GLM/Kimi/DeepSeek/Qwen team support)
version: "0.1.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: MoAI Plan Workflow
priority: P1
labels: [hybrid, multi-llm, glm, kimi, deepseek, qwen, breaking-change, v3r3]
issue_number: null
phase: "v3.0.0 — Phase 7 — Multi-LLM Hybrid Mode"
module: "internal/cli/, internal/llm/, .claude/skills/moai/workflows/, .moai/config/sections/llm.yaml"
dependencies:
  - SPEC-GLM-MCP-001
related_theme: "Theme 7 — Cost Optimization + LLM Abstraction"
breaking: true
bc_id: [BC-V3R3-HYBRID-001]
lifecycle: spec-anchored
tags: "hybrid, multi-llm, cost-optimization, llm-abstraction, breaking, v3r3"
---

# SPEC-V3R3-HYBRID-001: `moai hybrid` Multi-LLM Mode

## HISTORY

| Version | Date       | Author              | Description                                                                                              |
|---------|------------|---------------------|----------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | 최초 작성. `moai cg` 제거 + `moai hybrid` 도입 + 4 LLM provider (GLM/Kimi/DeepSeek/Qwen3-Coder) 1급 지원. BREAKING. |

---

## 1. Goal (목적)

`moai cg` 명령(Claude leader + GLM teammates 코스트-최적화 모드)을 **clean-break 방식으로 제거**하고, 그 자리를 다중 LLM provider를 1급 시민으로 지원하는 **`moai hybrid`** 명령으로 대체한다. 사용자 의도 그대로:

- **제거 대상**: `moai cg` (다음 major release에서 `BC-V3R3-HYBRID-001`로 즉시 폐기. deprecation alias 유지하지 않음.)
- **신규 도입**: `moai hybrid [provider] [-p profile]` — Claude를 leader로 두고 teammates를 사용자가 선택한 team-LLM provider로 실행.
- **1급 지원 4 LLM (allow-list, 닫힌 집합)**: GLM (Z.AI), Kimi K2 (Moonshot AI), DeepSeek V3+, Qwen3-Coder (Alibaba DashScope).
- **유지**: `moai glm` (단일 LLM 모드), `moai cc` (Claude 전용 모드)는 그대로 남는다. `moai hybrid`는 이들의 **다중-LLM 슈퍼셋**이지, 대체가 아니다.

### 1.1 배경

R2 (2026-04 회고): `moai cg`는 v2.x 시리즈에서 "Claude + GLM 코스트 60-70% 절감 모드"로 도입되어 1년간 운영되었다. 그러나 (a) `cg`라는 이름이 GLM에 종속되어 다른 team-LLM provider (Kimi/DeepSeek/Qwen3) 등장 시 확장 불가, (b) `moai glm`과 `moai cg`의 역할 분리(`glm` = all-LLM, `cg` = hybrid)는 사용자에게 혼동, (c) provider별 환경변수 주입 로직(`injectGLMEnvForTeam` `internal/cli/glm.go:520`)이 GLM 전용으로 하드코딩되어 다중 provider 지원 시 대규모 분기를 유발한다. v3R3에서 multi-provider abstraction을 도입하는 시점에 `moai cg`를 제거하고 `moai hybrid` 단일 명령으로 통일하는 것이 가장 깨끗한 경로다.

R3 (Z.AI/Anthropic-compat 검증 결과, research.md §3): GLM은 이미 Anthropic-compat endpoint (`https://api.z.ai/api/anthropic`)를 1급 제공한다. DeepSeek 또한 Anthropic-compat endpoint (`https://api.deepseek.com/anthropic`)를 v4 기준 정식 제공한다. Kimi와 Qwen3-Coder는 OpenAI-compat이 1차이며, Anthropic-compat은 (a) provider-side 정식 endpoint 또는 (b) 사용자가 운영하는 proxy/adapter (예: `moai hybrid kimi --proxy https://my-proxy/anthropic`)로 대응 가능하다. 본 SPEC은 4 provider를 모두 1급 시민으로 등록하되, **Anthropic-compat 가용성 검증 책임을 launch path에 이양**하여 endpoint 미가용 시 명확한 에러를 반환하도록 한다.

### 1.2 비목표 (Non-Goals)

- 5번째 LLM provider 추가 (allow-list는 4종 고정; v3R3 기간 동안 닫힌 집합)
- `moai glm`, `moai cc` 제거 또는 동작 변경 (단일-LLM 모드는 그대로)
- Claude OAuth 토큰 / `~/.claude/` 인증 흐름 변경
- LLM provider별 SDK 신규 의존성 추가 (Anthropic-compat HTTP만 사용; `moai` Go 바이너리는 endpoint URL + env-var 주입만 담당)
- Web LLM API integration (chat UI 내장 등) — `moai hybrid`는 Claude Code teammates용 CLI subagent 모드 한정
- Mainland China 전용 endpoint (예: api.deepseek.com.cn, dashscope.cn-hangzhou.aliyuncs.com)는 본 SPEC에서 다루지 않음 (사용자 별도 SPEC에서 처리)
- 동시 다중 provider 사용 (예: Claude leader + GLM analyst + Kimi tester를 한 세션에 혼합) — 한 세션은 한 provider만 선택; 멀티-provider 라우팅은 별도 SPEC
- LLM provider별 cache_control / prompt_caching 활성화 정책 변경 (현행 `DISABLE_PROMPT_CACHING=1` 정책 유지; 변경은 SPEC-GLM-MCP-001 후속)
- `moai cg` deprecation alias 유지 (clean break — 사용자 의도)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**:
  - `internal/cli/cg.go`: 파일 삭제 또는 BC 에러 stub로 축소 (M3 결정).
  - `internal/cli/hybrid.go` (신규): `moai hybrid` 명령 정의 + provider 라우팅.
  - `internal/llm/provider.go` (신규): provider abstraction (`type Provider interface { Name() string; AnthropicBaseURL() string; DefaultModels() ModelSet; EnvSchema() []EnvKey }`).
  - `internal/llm/providers/{glm,kimi,deepseek,qwen}.go` (신규 4 파일): provider별 메타데이터 정의 (base URL, latest stable model name, env var 이름).
  - `internal/cli/launcher.go`: `case "claude_glm"` (line 57) → `case "claude_hybrid"`로 교체 + provider 인자 라우팅.
  - `internal/cli/glm.go:218 enableTeamMode` → `enableHybridTeamMode(provider Provider)` 일반화 (GLM 하드코딩 분리).
  - `.moai/config/sections/llm.yaml`: `provider: <glm|kimi|deepseek|qwen>` + `provider.<name>.{base_url, models.{high,medium,low}, env_var}` 섹션 추가.
  - `internal/template/templates/.moai/config/sections/llm.yaml`: 위 스키마 미러 (Template-First HARD rule).
  - 4 provider env-key registry: `~/.moai/.env.{glm,kimi,deepseek,qwen}` 파일 (기존 `.env.glm` 패턴 확장).
  - `cmd/moai/main.go` 또는 `internal/cli/root.go`: `cgCmd` 등록 제거 + `hybridCmd` 등록 추가.
- **Modifies (read-only references)**:
  - `.claude/skills/moai/team/glm.md`: `moai cg` 언급(lines 41, 50, 103, 127, 143)을 `moai hybrid glm`으로 교체.
  - `.claude/skills/moai/team/run.md`: line 73, 88, 92, 103, 104 — 동일 substitution.
  - `.claude/skills/moai/workflows/run.md:904`: `active_mode: cc | glm | cg` → `cc | glm | hybrid`.
  - `.claude/rules/moai/development/model-policy.md:44`: `Activation: moai cg` → `moai hybrid <provider>`.
  - `CLAUDE.md §15 CG Mode`: 섹션 제목 `CG Mode` → `Hybrid Mode`로 변경 + 본문 4-provider 일반화.
  - `CLAUDE.local.md §16` 등 `moai cg` 언급: 사용자 로컬 가이드는 보존하되 `moai hybrid glm`로 substitution 권장 주석 추가.
- **CHANGELOG**: `BC-V3R3-HYBRID-001` 항목 (BREAKING CHANGE) + migration guide.
- **Test surface**: `internal/cli/hybrid_test.go` (신규) + `internal/llm/provider_test.go` (신규) + `internal/cli/cg_removal_test.go` (`moai cg` 호출 시 BC 에러 메시지 반환 검증).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 구현 코드 (Go function body, type method body) — 본 SPEC은 plan-phase 산출물; 실제 구현은 `/moai run SPEC-V3R3-HYBRID-001` 단계.
- Mainland China 전용 endpoint (예: `api.deepseek.com.cn`, `dashscope.cn-hangzhou.aliyuncs.com`) — 별도 SPEC에서 region-aware routing으로 처리.
- Web LLM API integration (Claude Code 외부에서 web UI로 LLM 호출) — `moai hybrid`는 Claude Code teammates의 CLI subagent context 한정. 웹 chat 통합은 본 SPEC 범위 밖.
- 4-LLM allow-list 외 provider (예: Anthropic 자체, OpenAI gpt-X, Mistral, Cohere, xAI Grok 등) — v3R3 기간 동안 닫힌 집합 유지. 5번째 provider 추가는 atomic-reversal SPEC (REQ-WF005-012 패턴 차용) 필요.
- LLM provider별 SDK 의존성 (e.g., `github.com/sashabaranov/go-openai`) — Anthropic-compat HTTP endpoint만 사용; provider별 native SDK 추가 금지.
- Multi-provider single-session routing (예: Claude leader + GLM analyst + Kimi tester 혼합) — 한 세션은 한 provider만 선택; 멀티-provider routing은 별도 SPEC.
- `moai cg` deprecation alias / soft warning 기간 — clean break (사용자 명시적 선택). `moai cg` 호출 시 즉시 BC 에러.
- Claude OAuth 토큰 / `~/.claude/` 인증 흐름 변경.
- `injectGLMEnvForTeam`의 `MOAI_BACKUP_AUTH_TOKEN` 백업 로직 변경 (provider 일반화 시 보존).
- Z.AI cache_control 정책 변경 (`DISABLE_PROMPT_CACHING=1` 등) — SPEC-GLM-MCP-001 후속.
- LLM provider별 MCP 서버 통합 (Z.AI Vision/WebSearch 등) — SPEC-GLM-MCP-001 의존.
- Provider별 동적 model alias (`kimi-k2-latest`, `deepseek-v3-latest`) — base SPEC date의 pinned baseline + `.moai/config/sections/llm.yaml` override 경로만 제공. Provider-side alias 자동 추적은 향후 SPEC.

---

## 3. Environment (환경)

- 런타임: Go 1.23+, Cobra CLI, Anthropic-compat HTTP endpoint.
- 영향 디렉터리:
  - 신규: `internal/cli/hybrid.go`, `internal/llm/provider.go`, `internal/llm/providers/{glm,kimi,deepseek,qwen}.go`.
  - 수정: `internal/cli/launcher.go`, `internal/cli/glm.go`, `internal/cli/root.go` (or cobra root), `.moai/config/sections/llm.yaml`, `internal/template/templates/.moai/config/sections/llm.yaml`, CLAUDE.md, `.claude/skills/moai/{team/glm.md, team/run.md, workflows/run.md}`, `.claude/rules/moai/development/model-policy.md`.
  - 삭제: `internal/cli/cg.go` (M3 결정에 따라 BC stub로 축소 가능).
- 외부 endpoints (research.md §3 검증 결과):
  - GLM: `https://api.z.ai/api/anthropic` (verified 2026-05-04, Anthropic-compat 1급).
  - Kimi: Anthropic-compat endpoint **사용자 검증 대기** — 후보 `https://api.moonshot.ai/anthropic` (확인 안 됨); fallback `--proxy <user-supplied>` 옵션 제공.
  - DeepSeek: `https://api.deepseek.com/anthropic` (verified 2026-05-04, Anthropic-compat 1급).
  - Qwen3-Coder: Anthropic-compat endpoint **사용자 검증 대기** — DashScope 공식 docs (Aliyun Bailian) WebFetch 실패 (404 / loading shell); fallback `--proxy` 옵션 제공.
- 외부 레퍼런스: `internal/cli/cg.go:1-58`, `internal/cli/glm.go:21-943`, `internal/cli/launcher.go:21-280` (모드 라우팅), `internal/config/types.go:52-100` (LLMConfig struct), CLAUDE.md §15, CLAUDE.local.md §16.

---

## 4. Assumptions (가정)

- `moai cg`는 v2.x 시리즈에서 1년간 안정 운영되었으나 사용자 베이스가 작아 (CG 모드 활성 사용자 ≤ 50명 추정) clean-break의 마이그레이션 부담이 제한적이다.
- GLM/DeepSeek Anthropic-compat endpoint는 향후 1년 지속 (provider commitment 관찰 결과).
- Kimi와 Qwen3-Coder의 Anthropic-compat은 (a) provider-side 공식 endpoint 또는 (b) `moai hybrid <provider> --proxy <url>` 옵션으로 대응. (a)가 미가용일 때 명확한 에러 메시지 반환.
- 사용자가 `~/.moai/.env.{glm,kimi,deepseek,qwen}` 파일에 provider별 API 키를 분리 저장한다 (`moai hybrid setup <provider>` subcommand로 wrapper 제공).
- `injectGLMEnvForTeam`의 tmux session-level env-var 주입 패턴은 provider-agnostic으로 일반화 가능하다 (env-key 이름만 provider별 차별화: `ANTHROPIC_AUTH_TOKEN`은 모든 provider 공통; `ANTHROPIC_BASE_URL`은 provider별 차별).
- `.moai/config/sections/llm.yaml`의 `team_mode: "cg"` 기록을 가진 기존 사용자 프로젝트는 `moai update` cleanup phase에서 `team_mode: "hybrid"` + `provider: "glm"`으로 자동 마이그레이션된다 (REQ-HYBRID-014).
- `moai cg` 호출 시 반환되는 BC 에러는 사용자가 즉시 새 명령(`moai hybrid glm`)을 인지할 수 있도록 actionable suggestion을 포함한다.
- 4 provider의 latest stable model 이름은 SPEC date(2026-05-04) 기준으로 pinned: GLM-4.7, Kimi K2.6, DeepSeek-V4, Qwen3-Coder Plus (research.md §3 검증; pinned baseline은 v3R3 기간 동안 유효; `.moai/config/sections/llm.yaml`로 사용자 override 가능).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-HYBRID-001**
The MoAI CLI **shall** expose `moai hybrid <provider> [-p profile] [-- claude-args...]` as the canonical command for Claude-leader + team-LLM-teammates hybrid execution.

**REQ-HYBRID-002**
The MoAI CLI **shall** accept exactly four `<provider>` values: `glm`, `kimi`, `deepseek`, `qwen`. Any other value **shall** be rejected with `UNKNOWN_PROVIDER` and the error message **shall** list the four allowed values.

**REQ-HYBRID-003**
The MoAI CLI **shall** delete or reduce `internal/cli/cg.go` such that invoking `moai cg [args...]` returns a BC error with sentinel `MOAI_CG_REMOVED` and an actionable migration suggestion (`use 'moai hybrid glm' instead`).

**REQ-HYBRID-004**
The MoAI configuration schema (`.moai/config/sections/llm.yaml`) **shall** include a top-level `provider` key (string, one of the four allow-list values, scalar active provider) and a `providers.<name>.{base_url, models.{high,medium,low}, env_var}` sub-section per allow-list provider (plural map container).

**REQ-HYBRID-005**
The 4 LLM providers **shall** be pinned at SPEC creation date (2026-05-04) to verified-stable model names: GLM-4.7, Kimi K2.6, DeepSeek-V4, Qwen3-Coder Plus (per research.md §3). User overrides via `.moai/config/sections/llm.yaml` `provider.<name>.models` sub-fields are permitted.

**REQ-HYBRID-006**
The MoAI CLI **shall** preserve `moai glm` (single-LLM all-GLM mode) and `moai cc` (Claude-only mode) without behavioral change. `moai hybrid` **shall** be the multi-LLM superset, not a replacement of these.

### 5.2 Event-Driven Requirements

**REQ-HYBRID-007**
**When** the user invokes `moai hybrid <provider>`, the CLI **shall**: (a) load `~/.moai/.env.<provider>` for the API key, (b) inject `ANTHROPIC_AUTH_TOKEN` and `ANTHROPIC_BASE_URL` into the active tmux session, (c) write `provider: "<name>"` and `team_mode: "hybrid"` to `.moai/config/sections/llm.yaml`, (d) launch Claude Code via `unifiedLaunch(profileName, "claude_hybrid", filteredArgs)`.

**REQ-HYBRID-008**
**When** the user invokes `moai hybrid <provider>` outside a tmux session, the CLI **shall** reject with `HYBRID_REQUIRES_TMUX` and provide the recovery path: `tmux new -s moai && moai hybrid <provider>`.

**REQ-HYBRID-009**
**When** the user invokes `moai hybrid <provider>` and `~/.moai/.env.<provider>` does not exist or is empty, the CLI **shall** reject with `HYBRID_MISSING_API_KEY` and instruct: `moai hybrid setup <provider> <api-key>`.

**REQ-HYBRID-010**
**When** the user invokes `moai cg [args...]`, the CLI **shall** print the BC error and exit with non-zero status; **shall not** attempt to launch Claude Code.

**REQ-HYBRID-011**
**When** the user invokes `moai update` against a project whose `.moai/config/sections/llm.yaml` contains `team_mode: "cg"`, the CLI cleanup phase **shall** rewrite to `team_mode: "hybrid"` + `provider: "glm"` and emit a one-time migration notice.

### 5.3 State-Driven Requirements

**REQ-HYBRID-012**
**While** `team_mode: "hybrid"` is active and `provider: "<name>"` is set, the active LLM provider **shall** be `<name>` and Claude Code teammates **shall** observe `ANTHROPIC_BASE_URL` pointing to that provider's Anthropic-compat endpoint.

**REQ-HYBRID-013**
**While** the user is on `moai hybrid <provider>` and the provider's Anthropic-compat endpoint is unreachable, the launch path **shall** propagate the underlying HTTP error and **shall not** silently fall back to a different provider.

**REQ-HYBRID-014**
**While** `BC-V3R3-HYBRID-001` is active in CHANGELOG, `moai cg` **shall** remain present in the binary as a BC stub returning `MOAI_CG_REMOVED`; `moai cg` **shall not** be silently absent (cobra "unknown command" error is insufficient).

### 5.4 Optional Requirements

**REQ-HYBRID-015**
**Where** the provider's Anthropic-compat endpoint is not yet officially available (Kimi, Qwen3-Coder per research.md §3 pending verification), the user **shall** be able to supply `--proxy <url>` to override `ANTHROPIC_BASE_URL` at launch time.

**REQ-HYBRID-016**
**Where** the user wants per-provider tools toggle (e.g., enabling Z.AI Vision MCP for `glm` only — depends on SPEC-GLM-MCP-001), the CLI **shall** support `moai hybrid <provider> tools enable|disable` as a forward-looking subcommand surface.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-HYBRID-017 (Unwanted Behavior)**
**If** a future PR proposes a 5th provider (e.g., `mistral`, `cohere`, `xai-grok`), **then** CI **shall** reject with `PROVIDER_ALLOWLIST_VIOLATION` per the closed-set rule (REQ-HYBRID-002). Reversal requires a new SPEC with atomic-migration semantics.

**REQ-HYBRID-018 (Complex: Auth Failure)**
**While** `moai hybrid <provider>` is launching, **when** the provider returns HTTP 401/403 (auth failure), the CLI **shall** mask the API key in any error message (per `maskAPIKey` `internal/cli/glm.go:208`) and emit `HYBRID_AUTH_FAILED` with a hint to verify the key via `moai hybrid status <provider>`.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-HYBRID-01** (REQ-HYBRID-001): Given a user runs `moai hybrid --help` When the help output is rendered Then the synopsis line `moai hybrid <provider> [-p profile]` appears and the description names the four allow-list providers.
- **AC-HYBRID-02** (REQ-HYBRID-002): Given a user runs `moai hybrid mistral` When the CLI parses arguments Then exit code is non-zero and stderr contains `UNKNOWN_PROVIDER` plus the four allowed values.
- **AC-HYBRID-03** (REQ-HYBRID-003, REQ-HYBRID-010, REQ-HYBRID-014): Given a user runs `moai cg` after upgrade When the CLI executes Then exit code is non-zero and stderr contains `MOAI_CG_REMOVED` and `use 'moai hybrid glm' instead`.
- **AC-HYBRID-04** (REQ-HYBRID-004): Given the embedded template `.moai/config/sections/llm.yaml` When parsed Then it contains a top-level `provider:` key (default empty string) and `provider.<name>.{base_url, models.high, models.medium, models.low, env_var}` for all four providers.
- **AC-HYBRID-05** (REQ-HYBRID-005): Given the embedded template When `provider.glm.models.high`, `provider.kimi.models.high`, `provider.deepseek.models.high`, `provider.qwen.models.high` are read Then the values match the SPEC-pinned baselines (GLM-4.7, Kimi K2.6, DeepSeek-V4, Qwen3-Coder Plus or equivalent — see research.md §3).
- **AC-HYBRID-06** (REQ-HYBRID-006): Given a user runs `moai glm` and `moai cc` after upgrade When the binaries execute Then both produce the v2.x baseline behavior (no functional regression).
- **AC-HYBRID-07** (REQ-HYBRID-007): Given a user inside a tmux session runs `moai hybrid glm` with `~/.moai/.env.glm` populated When the launch completes Then `tmux show-environment | grep ANTHROPIC_BASE_URL` returns the GLM endpoint and `.moai/config/sections/llm.yaml` shows `team_mode: "hybrid"` + `provider: "glm"`.
- **AC-HYBRID-08** (REQ-HYBRID-008): Given a user outside tmux runs `moai hybrid kimi` When the CLI executes Then exit code is non-zero and stderr contains `HYBRID_REQUIRES_TMUX` plus the recovery path.
- **AC-HYBRID-09** (REQ-HYBRID-009): Given a user inside tmux runs `moai hybrid deepseek` and `~/.moai/.env.deepseek` is missing When the CLI executes Then exit code is non-zero and stderr contains `HYBRID_MISSING_API_KEY` plus `moai hybrid setup deepseek`.
- **AC-HYBRID-10** (REQ-HYBRID-011): Given a project with `team_mode: "cg"` in `.moai/config/sections/llm.yaml` When `moai update` runs Then post-update the file contains `team_mode: "hybrid"` and `provider: "glm"` and stderr emits a one-time migration notice.
- **AC-HYBRID-11** (REQ-HYBRID-012): Given `moai hybrid qwen` succeeds When a Claude Code teammate launches in a new tmux pane Then `env | grep ANTHROPIC_BASE_URL` returns the Qwen endpoint (or proxy URL if --proxy was used).
- **AC-HYBRID-12** (REQ-HYBRID-013): Given the GLM endpoint is unreachable (simulated via DNS block) When `moai hybrid glm` launches Then the underlying HTTP error is propagated to stderr and the CLI does NOT silently fall back to Kimi/DeepSeek/Qwen.
- **AC-HYBRID-13** (REQ-HYBRID-015): Given a user runs `moai hybrid kimi --proxy https://my-proxy/anthropic` When the launch completes Then `ANTHROPIC_BASE_URL` is set to `https://my-proxy/anthropic` (overriding the provider default).
- **AC-HYBRID-14** (REQ-HYBRID-016): Given a user runs `moai hybrid glm tools --help` When the help output is rendered Then it lists `enable` and `disable` subcommand stubs (forward-looking surface; full implementation deferred to SPEC-GLM-MCP-001).
- **AC-HYBRID-15** (REQ-HYBRID-017): Given a PR adding `internal/llm/providers/mistral.go` to the codebase When CI runs Then the test `TestProviderAllowlist` fails with `PROVIDER_ALLOWLIST_VIOLATION` referencing the four allowed providers.
- **AC-HYBRID-16** (REQ-HYBRID-018): Given `moai hybrid glm` is launching and the API key is invalid When the provider returns HTTP 401 Then stderr contains `HYBRID_AUTH_FAILED` and the API key in the error message is masked (e.g., `sk-x****wxyz`, never the full key).
- **AC-HYBRID-17** (REQ-HYBRID-005, REQ-HYBRID-014): Given an upgraded user re-runs the same workflow that previously used `moai cg` When `moai hybrid glm` is invoked Then the resulting tmux env-vars and settings.local.json `teammateMode: "tmux"` match the v2.x `moai cg` behavior bit-for-bit (no regression).
- **AC-HYBRID-18** (REQ-HYBRID-002, REQ-HYBRID-017): Given the `internal/llm/providers/registry.go` (or equivalent) When inspected by static analysis or unit test Then exactly four provider entries are present (`glm`, `kimi`, `deepseek`, `qwen`) and no five-provider configuration loads at runtime.

---

## 7. Constraints (제약)

- 16-language neutrality (CLAUDE.local.md §15): provider 이름은 모두 lowercase ASCII, 다국어 번역 이름은 `.claude/skills/moai/team/glm.md` 등 사용자 문서에서만 가능.
- Template-First HARD rule (CLAUDE.local.md §2): `.moai/config/sections/llm.yaml` 변경은 `internal/template/templates/.moai/config/sections/llm.yaml` mirror + `make build` 필수.
- v3R3 시리즈 BC 가시성: 모든 BREAKING CHANGE는 `bc_id` (`BC-V3R3-HYBRID-001`)로 추적되며 CHANGELOG 명시 + migration guide 필수.
- 9-direct-dep 정책: SPEC-GLM-MCP-001을 직접 dependency로 선언; provider별 native SDK 의존성 추가 금지 (Anthropic-compat HTTP만).
- tmux-first: hybrid 모드는 tmux session-level env isolation에 의존 (현행 `cg` 모드와 동일한 architectural choice).
- API-key 보안: provider 키는 절대 stdout/stderr에 plain-text로 출력 금지 (현행 `maskAPIKey` `internal/cli/glm.go:208` 패턴 일반화).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 확률 | 완화 |
|---|---|---|---|
| `moai cg` 사용자(추정 ≤ 50명)가 upgrade 후 명령 부재로 워크플로 중단 | M | M | clean-break + BC stub로 즉시 actionable 에러 메시지 (`use 'moai hybrid glm' instead`); CHANGELOG `BC-V3R3-HYBRID-001` 명시 + docs-site 업데이트 |
| Kimi, Qwen3 Anthropic-compat endpoint가 provider-side에서 정식 출시되지 않은 상태 | M | M | `--proxy <url>` 옵션 (REQ-HYBRID-015)으로 사용자 운영 proxy 허용; provider 등록은 1급으로 유지하되 launch 시 endpoint 미가용 → 명확한 에러 |
| 4 provider의 latest model 이름이 SPEC date 이후 변경 (e.g., GLM-4.8, Kimi K3 출시) | M | H | `.moai/config/sections/llm.yaml`로 사용자 override 경로 제공 (REQ-HYBRID-005); SPEC-pinned baseline은 v3R3 기간 동안 유효 마감 |
| 기존 `team_mode: "cg"` config를 가진 사용자 프로젝트가 `moai update` 미실행 상태에서 `moai hybrid` 호출 | L | M | REQ-HYBRID-011 자동 마이그레이션 + REQ-HYBRID-014 BC stub로 `moai cg` 호출 또한 차단 — 양방향 안전망 |
| provider-별 env-var 주입 시 `MOAI_BACKUP_AUTH_TOKEN` 백업 로직이 다중 provider 전환 시 깨짐 (e.g., GLM → Kimi 전환 시 GLM 키가 deepseek 백업으로 잘못 저장) | M | M | provider-별로 백업 키 namespace 분리 (`MOAI_BACKUP_AUTH_TOKEN_<provider_uppercase>`); 단일 백업 슬롯 패턴 deprecate |
| Template-First mirror 누락 → `internal/template/embedded.go` regenerate 실패 | H | L | M2/M3/M4 각 milestone 종료 시 `make build` 강제 + CI에서 embedded FS 동기화 검증 |
| Anthropic-compat endpoint 미가용 시 Claude Code가 cryptic error 반환 (e.g., "context window exceeded" misreport) | M | M | launch 직전 health check (`HEAD <base_url>/v1/messages` 또는 GET 인증 endpoint) 실행 옵션 (REQ-HYBRID-013); 미구현 시 사용자 의식적 trade-off |
| 4 provider별 model-tier 매핑(high/medium/low) 의미가 provider마다 상이 → cost-optimization 효과 불확실 | L | H | `.moai/config/sections/llm.yaml`의 `models.{high,medium,low}` 사용자 직접 override 권장; SPEC-pinned baseline은 provider docs 기준 best-fit |
| `moai cg` BC stub이 단순 cobra "unknown command" 으로 fallback 시 actionable 메시지 사라짐 | L | M | REQ-HYBRID-014 명시: `cgCmd`는 binary에 stub로 보존 (RunE에서 즉시 BC 에러 + 비-zero exit). 단순 삭제 금지 |
| docs-site 4개국어 동기화 누락 (CLAUDE.local.md §17) | M | M | sync-phase에서 manager-docs가 ko/en/zh/ja 4 locale 모두 업데이트; CI script `docs-i18n-check.sh` 검증 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-GLM-MCP-001** (PR #769 OPEN): Z.AI MCP 서버 통합. 본 SPEC의 REQ-HYBRID-016 `tools enable|disable` subcommand surface는 SPEC-GLM-MCP-001 implementation에 위임. 본 SPEC은 forward-looking surface만 정의; 실제 MCP attach 로직은 GLM-MCP-001 후속 PR에서.

### 9.2 Blocks

- 향후 Mainland China region SPEC (region-aware routing for `api.deepseek.com.cn`, `dashscope.cn-hangzhou.aliyuncs.com`) — 본 SPEC의 provider abstraction을 base로 사용.
- 향후 Multi-provider single-session SPEC (Claude leader + GLM analyst + Kimi tester 혼합) — 본 SPEC의 provider registry를 base로.

### 9.3 Related

- SPEC-V3R3-UPDATE-CLEANUP-001 (PR #764 merged): `moai update` cleanup phase가 본 SPEC의 REQ-HYBRID-011 자동 마이그레이션을 실행할 인프라를 이미 제공한다.
- SPEC-V3R2-WF-005 (PR #768 merged): 16-language neutrality 패턴을 provider neutrality에도 차용 (provider 4종 동등 취급, 어느 provider도 "primary" 라벨 금지).

---

## 10. BC Migration (Breaking Change 처리)

### 10.1 BC ID

- **BC-V3R3-HYBRID-001**: `moai cg` 명령 제거 + `moai hybrid <provider>` 도입.

### 10.2 Migration Path

| Before (v2.x) | After (v3R3) |
|---|---|
| `moai glm setup sk-xxx` | `moai hybrid setup glm sk-xxx` (구 `moai glm setup`도 alias로 1년간 유지) |
| `moai cg` | `moai hybrid glm` |
| `moai cg -p work` | `moai hybrid glm -p work` |
| `.moai/config/sections/llm.yaml: team_mode: "cg"` | `.moai/config/sections/llm.yaml: team_mode: "hybrid"` + `provider: "glm"` (자동 마이그레이션 by REQ-HYBRID-011) |

### 10.3 CHANGELOG Wording (proposed)

```markdown
## [Unreleased]

### Breaking Changes

- **BC-V3R3-HYBRID-001 (SPEC-V3R3-HYBRID-001)**: Removed `moai cg` (Claude + GLM hybrid) and replaced it with `moai hybrid <provider>`, which now supports four team-LLM providers as first-class citizens: GLM (Z.AI), Kimi K2 (Moonshot AI), DeepSeek V4 (DeepSeek), and Qwen3-Coder (Alibaba DashScope).
  - `moai cg` invocations after upgrade return `MOAI_CG_REMOVED` with an actionable migration suggestion (`use 'moai hybrid glm' instead`).
  - Existing projects with `team_mode: "cg"` in `.moai/config/sections/llm.yaml` are auto-migrated to `team_mode: "hybrid"` + `provider: "glm"` on the next `moai update` run.
  - `moai glm` (single-LLM all-GLM mode) and `moai cc` (Claude-only mode) remain unchanged.
  - See SPEC-V3R3-HYBRID-001 §10 BC Migration for full migration table.
```

### 10.4 Documentation Touch Points

- `CLAUDE.md §15`: 섹션 제목 `CG Mode` → `Hybrid Mode` + 본문 4-provider 일반화.
- `CLAUDE.local.md §16` (사용자 로컬 가이드): 보존하되 `moai cg` 언급 옆에 `→ moai hybrid glm` substitution 주석.
- `docs-site` 4개국어 (ko/en/zh/ja): hybrid 모드 사용 가이드 + BC 마이그레이션 페이지 신설 (CLAUDE.local.md §17 4-locale 동기화 의무).
- `.claude/skills/moai/team/glm.md`: file 자체는 보존 (GLM 전용 사용 사례 가이드 역할 유지)하되 `moai cg` 명령은 `moai hybrid glm`으로 substitution.

### 10.5 Removal Timeline

- v3R3 시리즈 첫 minor release: `moai cg` 즉시 BC stub로 축소 (clean break, deprecation alias 없음).
- BC stub은 v3R4 시리즈 종료까지 유지 (`MOAI_CG_REMOVED` 에러 메시지 제공) → v3R5에서 cobra root에서 완전 제거 검토 (별도 cleanup SPEC).

---

## 11. Traceability (추적성)

- REQ 총 18개: Ubiquitous 6, Event-Driven 5, State-Driven 3, Optional 2, Complex 2.
- AC 총 18개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지) — 매트릭스는 plan.md §1.4 참조.
- 4 LLM provider 검증: research.md §3 (GLM/DeepSeek 검증 완료, Kimi/Qwen3 사용자 검증 대기).
- BC 영향: `BC-V3R3-HYBRID-001` 1건 (clean break, deprecation alias 없음).
- 의존성: SPEC-GLM-MCP-001 (PR #769 OPEN, plan-only) 1건.
- 구현 경로 예상:
  - 신규 4 provider 메타데이터 파일 (`internal/llm/providers/{glm,kimi,deepseek,qwen}.go`)
  - `internal/cli/cg.go` BC stub 변경
  - `internal/cli/launcher.go` `claude_glm` → `claude_hybrid` 라우팅 일반화
  - `.moai/config/sections/llm.yaml` 스키마 확장
  - 6 .claude/skills/.claude/rules 문서 substitution (~10 mention)
  - CHANGELOG `BC-V3R3-HYBRID-001` 항목

---

End of SPEC.
