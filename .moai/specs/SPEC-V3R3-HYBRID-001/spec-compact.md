# SPEC-V3R3-HYBRID-001 Compact Reference

> Auto-extract from `spec.md` v0.1.0 — REQs + ACs + Files + Exclusions only.
> Use this file for fast plan-auditor scans and cross-SPEC referencing.

---

## Requirements (EARS)

### Ubiquitous Requirements

- **REQ-HYBRID-001**: The MoAI CLI **shall** expose `moai hybrid <provider> [-p profile] [-- claude-args...]` as the canonical command for Claude-leader + team-LLM-teammates hybrid execution.
- **REQ-HYBRID-002**: The MoAI CLI **shall** accept exactly four `<provider>` values: `glm`, `kimi`, `deepseek`, `qwen`. Any other value **shall** be rejected with `UNKNOWN_PROVIDER` and the error message **shall** list the four allowed values.
- **REQ-HYBRID-003**: The MoAI CLI **shall** delete or reduce `internal/cli/cg.go` such that invoking `moai cg [args...]` returns a BC error with sentinel `MOAI_CG_REMOVED` and an actionable migration suggestion (`use 'moai hybrid glm' instead`).
- **REQ-HYBRID-004**: The MoAI configuration schema (`.moai/config/sections/llm.yaml`) **shall** include a top-level `provider` key (string, one of the four allow-list values) and a `provider.<name>.{base_url, models.{high,medium,low}, env_var}` sub-section per allow-list provider.
- **REQ-HYBRID-005**: The 4 LLM providers **shall** be pinned at SPEC creation date (2026-05-04) to verified-stable model names: GLM-4.7, Kimi K2.6, DeepSeek-V4, Qwen3-Coder Plus (per research.md §3). User overrides via `.moai/config/sections/llm.yaml` `provider.<name>.models` sub-fields are permitted.
- **REQ-HYBRID-006**: The MoAI CLI **shall** preserve `moai glm` (single-LLM all-GLM mode) and `moai cc` (Claude-only mode) without behavioral change. `moai hybrid` **shall** be the multi-LLM superset, not a replacement of these.

### Event-Driven Requirements

- **REQ-HYBRID-007**: **When** the user invokes `moai hybrid <provider>`, the CLI **shall**: (a) load `~/.moai/.env.<provider>` for the API key, (b) inject `ANTHROPIC_AUTH_TOKEN` and `ANTHROPIC_BASE_URL` into the active tmux session, (c) write `provider: "<name>"` and `team_mode: "hybrid"` to `.moai/config/sections/llm.yaml`, (d) launch Claude Code via `unifiedLaunch(profileName, "claude_hybrid", filteredArgs)`.
- **REQ-HYBRID-008**: **When** the user invokes `moai hybrid <provider>` outside a tmux session, the CLI **shall** reject with `HYBRID_REQUIRES_TMUX` and provide the recovery path: `tmux new -s moai && moai hybrid <provider>`.
- **REQ-HYBRID-009**: **When** the user invokes `moai hybrid <provider>` and `~/.moai/.env.<provider>` does not exist or is empty, the CLI **shall** reject with `HYBRID_MISSING_API_KEY` and instruct: `moai hybrid setup <provider> <api-key>`.
- **REQ-HYBRID-010**: **When** the user invokes `moai cg [args...]`, the CLI **shall** print the BC error and exit with non-zero status; **shall not** attempt to launch Claude Code.
- **REQ-HYBRID-011**: **When** the user invokes `moai update` against a project whose `.moai/config/sections/llm.yaml` contains `team_mode: "cg"`, the CLI cleanup phase **shall** rewrite to `team_mode: "hybrid"` + `provider: "glm"` and emit a one-time migration notice.

### State-Driven Requirements

- **REQ-HYBRID-012**: **While** `team_mode: "hybrid"` is active and `provider: "<name>"` is set, the active LLM provider **shall** be `<name>` and Claude Code teammates **shall** observe `ANTHROPIC_BASE_URL` pointing to that provider's Anthropic-compat endpoint.
- **REQ-HYBRID-013**: **While** the user is on `moai hybrid <provider>` and the provider's Anthropic-compat endpoint is unreachable, the launch path **shall** propagate the underlying HTTP error and **shall not** silently fall back to a different provider.
- **REQ-HYBRID-014**: **While** `BC-V3R3-HYBRID-001` is active in CHANGELOG, `moai cg` **shall** remain present in the binary as a BC stub returning `MOAI_CG_REMOVED`; `moai cg` **shall not** be silently absent (cobra "unknown command" error is insufficient).

### Optional Requirements

- **REQ-HYBRID-015**: **Where** the provider's Anthropic-compat endpoint is not yet officially available (Kimi, Qwen3-Coder per research.md §3 pending verification), the user **shall** be able to supply `--proxy <url>` to override `ANTHROPIC_BASE_URL` at launch time.
- **REQ-HYBRID-016**: **Where** the user wants per-provider tools toggle (e.g., enabling Z.AI Vision MCP for `glm` only — depends on SPEC-GLM-MCP-001), the CLI **shall** support `moai hybrid <provider> tools enable|disable` as a forward-looking subcommand surface.

### Complex Requirements (Unwanted Behavior / Composite)

- **REQ-HYBRID-017** (Unwanted Behavior): **If** a future PR proposes a 5th provider (e.g., `mistral`, `cohere`, `xai-grok`), **then** CI **shall** reject with `PROVIDER_ALLOWLIST_VIOLATION` per the closed-set rule (REQ-HYBRID-002). Reversal requires a new SPEC with atomic-migration semantics.
- **REQ-HYBRID-018** (Complex: Auth Failure): **While** `moai hybrid <provider>` is launching, **when** the provider returns HTTP 401/403 (auth failure), the CLI **shall** mask the API key in any error message (per `maskAPIKey` `internal/cli/glm.go:208`) and emit `HYBRID_AUTH_FAILED` with a hint to verify the key via `moai hybrid status <provider>`.

---

## Acceptance Criteria

- **AC-HYBRID-01** (REQ-HYBRID-001): Given a user runs `moai hybrid --help` When the help output is rendered Then the synopsis line `moai hybrid <provider> [-p profile]` appears and the description names the four allow-list providers.
- **AC-HYBRID-02** (REQ-HYBRID-002): Given a user runs `moai hybrid mistral` When the CLI parses arguments Then exit code is non-zero and stderr contains `UNKNOWN_PROVIDER` plus the four allowed values.
- **AC-HYBRID-03** (REQ-HYBRID-003, 010, 014): Given a user runs `moai cg` after upgrade When the CLI executes Then exit code is non-zero and stderr contains `MOAI_CG_REMOVED` and `use 'moai hybrid glm' instead`.
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
- **AC-HYBRID-17** (REQ-HYBRID-005, 014): Given an upgraded user re-runs the same workflow that previously used `moai cg` When `moai hybrid glm` is invoked Then the resulting tmux env-vars and settings.local.json `teammateMode: "tmux"` match the v2.x `moai cg` behavior bit-for-bit (no regression).
- **AC-HYBRID-18** (REQ-HYBRID-002, 017): Given the `internal/llm/providers/registry.go` (or equivalent) When inspected by static analysis or unit test Then exactly four provider entries are present (`glm`, `kimi`, `deepseek`, `qwen`) and no five-provider configuration loads at runtime.

---

## Files to Modify

### To-be-created (13 files)

#### Test scaffolds (M1, 4 files)

- `internal/template/provider_allowlist_audit_test.go` (REQ-HYBRID-002, 017 audit)
- `internal/cli/cg_removal_test.go` (REQ-HYBRID-003, 010, 014 BC verification)
- `internal/cli/hybrid_test.go` (REQ-HYBRID-001..009 + 012, 013, 015, 018)
- `internal/cli/migration_test.go` (REQ-HYBRID-011 cg→hybrid migration)

#### Provider abstraction layer (M2, 7 files)

- `internal/llm/provider.go` (Provider interface + ModelSet struct)
- `internal/llm/providers/glm.go` (GLM provider metadata + Z.AI compat flags)
- `internal/llm/providers/kimi.go` (Kimi K2 provider metadata, base_url pending)
- `internal/llm/providers/deepseek.go` (DeepSeek V4 provider metadata)
- `internal/llm/providers/qwen.go` (Qwen3-Coder provider metadata, base_url pending)
- `internal/llm/providers/registry.go` (4-provider Registry + List() + Lookup())
- `internal/llm/registry_test.go` (registry consistency unit tests)

#### Implementation files (M3-M4, 2 files)

- `internal/cli/hybrid.go` (`moai hybrid` command + setup/status/tools subcommands)
- `internal/llm/inject.go` (provider-agnostic env-injection helpers refactored from glm.go)

### To-be-modified (~13 source-of-truth files + ~5 template mirrors)

#### Go source (M2-M4, 5 files)

- `internal/cli/cg.go` (REPLACE with BC stub returning `MOAI_CG_REMOVED`)
- `internal/cli/glm.go` (lines 148-149 message + 218-348 generalize + 266 persistTeamMode + 356-385/520-581 delegate to internal/llm + 536-538 namespaced backup; preserve glmCmd, runGLM body)
- `internal/cli/launcher.go` (line 40 comment + 57 case label + 185 persistTeamMode + 267 reset comment)
- `internal/cli/update.go` (line ~1411 add migrateCGTeamMode helper)
- `internal/config/types.go` (line 52-100 extend LLMConfig with Provider + Providers fields)
- `internal/config/defaults.go` (add NewDefaultProvidersConfig)

#### Project config (M2, 1 file + template mirror)

- `.moai/config/sections/llm.yaml` (add `provider: ""` + `providers:` section)
- `internal/template/templates/.moai/config/sections/llm.yaml` (mirror)

#### Documentation skill/rule files (M5a-d, 4 files + template mirrors)

- `.claude/skills/moai/team/glm.md` (lines 41, 50, 103, 127, 143: `moai cg` → `moai hybrid glm` 5 mentions)
- `.claude/skills/moai/team/run.md` (lines 73, 88, 92, 103, 104: `moai cg` → `moai hybrid <provider>` 5 mentions)
- `.claude/skills/moai/workflows/run.md:904` (`active_mode: cc | glm | cg` → `active_mode: cc | glm | hybrid`)
- `.claude/rules/moai/development/model-policy.md:44` (`Activation: moai cg` → `Activation: moai hybrid <provider>`)
- All four mirrored to `internal/template/templates/.claude/...`

#### Project-root documentation (M5e-g, 3 files; not in template)

- `CLAUDE.md §15` (rename `### CG Mode` heading to `### Hybrid Mode` + 4-provider generalization)
- `CLAUDE.local.md §16 line 129` (`'moai cg'` → `'moai hybrid <provider>'` in runtime-managed list)
- `CHANGELOG.md` (`## [Unreleased]` add `BC-V3R3-HYBRID-001` entry per spec.md §10.3)

#### Embedded template parity

- All edits to `.claude/skills/.claude/rules/.moai/config/` mirrored to `internal/template/templates/...` (per `CLAUDE.local.md §2 Template-First Rule HARD constraint`)

---

## Exclusions (Out of Scope — What NOT to Build)

Per `spec.md` §1.2 Non-Goals + §2.2 Out of Scope:

1. 구현 코드 (Go function body, type method body) — 본 SPEC은 plan-phase 산출물; 실제 구현은 `/moai run SPEC-V3R3-HYBRID-001` 단계.
2. 5번째 LLM provider 추가 (allow-list는 4종 고정; v3R3 기간 동안 닫힌 집합) — REQ-HYBRID-002, REQ-HYBRID-017로 차단.
3. Mainland China 전용 endpoint (예: `api.deepseek.com.cn`, `dashscope.cn-hangzhou.aliyuncs.com`) — 별도 SPEC에서 region-aware routing으로 처리.
4. Web LLM API integration (Claude Code 외부에서 web UI로 LLM 호출) — `moai hybrid`는 Claude Code teammates의 CLI subagent context 한정.
5. LLM provider별 native SDK 의존성 (e.g., `github.com/sashabaranov/go-openai`) — Anthropic-compat HTTP endpoint만 사용; provider별 native SDK 추가 금지.
6. Multi-provider single-session routing (예: Claude leader + GLM analyst + Kimi tester 혼합) — 한 세션은 한 provider만 선택; 멀티-provider routing은 별도 SPEC.
7. `moai cg` deprecation alias / soft warning 기간 — clean break (사용자 명시적 선택). `moai cg` 호출 시 즉시 BC 에러.
8. Claude OAuth 토큰 / `~/.claude/` 인증 흐름 변경.
9. `moai glm` (단일 LLM 모드), `moai cc` (Claude 전용 모드) 동작 변경 — 단일-LLM 모드는 그대로 유지 (REQ-HYBRID-006).
10. `injectGLMEnvForTeam`의 `MOAI_BACKUP_AUTH_TOKEN` 백업 로직 변경 (provider 일반화 시 보존; namespaced 분리만 추가).
11. Z.AI cache_control / prompt_caching 정책 변경 (`DISABLE_PROMPT_CACHING=1` 등) — SPEC-GLM-MCP-001 후속.
12. LLM provider별 MCP 서버 통합 (Z.AI Vision/WebSearch 등) — SPEC-GLM-MCP-001 의존; REQ-HYBRID-016은 forward-looking surface stub만 정의.
13. Provider별 동적 model alias (`kimi-k2-latest`, `deepseek-v3-latest`) — base SPEC date의 pinned baseline + `.moai/config/sections/llm.yaml` override 경로만 제공. Provider-side alias 자동 추적은 향후 SPEC.

---

End of spec-compact.md.
