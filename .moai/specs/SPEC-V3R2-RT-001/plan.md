# SPEC-V3R2-RT-001 Implementation Plan

> Implementation plan for **Hook JSON-OR-ExitCode Dual Protocol**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored against branch `plan/SPEC-V3R2-RT-001` at worktree `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001`. See §7 for cwd resolution rule.

## HISTORY

| Version | Date       | Author                        | Description                                                              |
|---------|------------|-------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B. Hooks 부분 구현(`dual_parse.go`, `response.go`)이 이미 존재하므로 본 plan은 27개 EARS REQ + 15개 AC를 100% 만족시키도록 갭을 메우는 작업으로 구성된다. |

---

## 1. Plan Overview

### 1.1 Goal restatement

Claude Code 2026.x 가 정의한 `HookJSONOutput` (JSON-OR-ExitCode) 프로토콜을 moai 의 27개 Claude Code 후크 이벤트 전체에 도입한다. `internal/hook/*.go` 의 핸들러는 stdout 으로 typed `HookResponse` JSON 페이로드를 emit 하며, stdout 이 비어있거나 JSON 파싱이 실패하면 런타임이 legacy exit-code 의미론으로 fallback 한다.

본 plan 은 다음을 달성한다:

- `internal/hook/response.go`(이미 존재, 288 줄)의 `HookResponse` 구조체를 master §4.3 Layer 3 type block 에 정렬되도록 강화 (`Continue *bool`, `WatchPaths []string`, `Retry *RetryHint` 모두 이미 정의됨; `validator/v10` 태그 + `Validate()` 메서드 추가).
- `internal/hook/dual_parse.go`(이미 존재, 190 줄)의 `ParseHookOutput` dual-parse 의미론을 27개 이벤트 전부에 통일적으로 적용하고 strict_mode + MOAI_HOOK_LEGACY 환경 분기를 도입.
- 27개 이벤트 variant types (PreToolUseOutput, PostToolUseOutput, ... StopFailureOutput) 의 `HookEventName() string` 메서드는 이미 `response.go:85-287` 에 정의됨; 본 plan 은 누락된 mismatch detection (`HookSpecificOutputMismatch`) wiring 을 `registry.go:65-200` Dispatch 경로에 추가.
- `internal/config/types.go` 에 `HookConfig{StrictMode bool}` 필드 추가 + `.moai/config/sections/system.yaml` `hook.strict_mode: true|false` 키 로딩.
- `MOAI_HOOK_LEGACY=1` 환경변수 분기 + 세션 단위 deprecation banner 1회 emit 로직.
- Wrappers-unchanged 정책: `.claude/hooks/moai/*.sh` 26개(현재 27개 중 1개는 spec-status 단축형) 셸 래퍼는 stdin/stdout 을 그대로 forward 하며 본 SPEC 에서는 절대 수정하지 않는다 (per spec.md §2 Out-of-scope, master §8 BC-V3R2-001).
- BC-V3R2-001 마이그레이션 윈도(v3.0.0 → v3.x)에 걸쳐 dual-parse fallback 은 항상 작동; v4.0 에서 exit-code-only 경로 제거 예정.

본 plan 은 **breaking change** (`spec.md` frontmatter `breaking: true`) 이지만 wrappers-unchanged + handlers-rewritten 호환성 shim 을 통해 사용자-체감 회귀 0 을 목표로 한다 (§8 참고).

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED**: 18개 새로운 실패 테스트를 작성 — 27개 variant 의 `HookEventName()` round-trip, validator/v10 거절 사례, strict_mode 거절, MOAI_HOOK_LEGACY suppress, deprecation banner once-per-session, HookSpecificOutputMismatch detection, 64 KiB AdditionalContext 절단, UpdatedInput-then-PermissionDecision 순서, api_version 2 opt-in, plugin-source bypass. 기존 `internal/hook/dual_parse_test.go`, `response_test.go`, `wire_format_freeze_test.go` 의 happy-path 는 보존 (regression baseline).
- **GREEN**: `response.go` 에 validator 태그 + `Validate()` 메서드 추가. `dual_parse.go` 에 strict_mode/legacy-env 분기 + banner once-per-session 로직 추가. `registry.go` Dispatch 에 HookSpecificOutput.HookEventName mismatch detection 삽입. `types.go` 에 HookConfig 필드 추가. 새 파일 `internal/hook/strict_mode.go`, `internal/hook/banner.go`, `internal/hook/api_version.go` 생성.
- **REFACTOR**: 5개 hook 핸들러 (`session_start.go`, `session_end.go`, `pre_tool.go`, `post_tool.go`, `user_prompt_submit.go`) 의 inline JSON 응답 생성 코드를 typed `HookResponse` builder 로 통일. `ToHookOutput()`/`ToHookResponse()` (이미 `dual_parse.go:117-189` 존재) 의 호환성 shim 을 모든 핸들러에서 사용하도록 정리.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| `validator/v10` 의존성 추가 | `go.mod` (직접 의존성 추가; SCH-001 미머지 시 본 SPEC 가 책임) | REQ-006 |
| `HookResponse` validator 태그 + `Validate()` | `internal/hook/response.go` (extend) | REQ-001, REQ-002, REQ-006 |
| 27 variant type `HookEventName()` 검증 | `internal/hook/response.go` (이미 line 85-287; 미처리 누락 검증) | REQ-003, REQ-004 |
| Dual-parse strict_mode 분기 | `internal/hook/dual_parse.go` (extend) + `internal/hook/strict_mode.go` (new) | REQ-005, REQ-021 |
| MOAI_HOOK_LEGACY env 분기 | `internal/hook/strict_mode.go` (new) | REQ-020 |
| Deprecation banner once-per-session | `internal/hook/banner.go` (new) | REQ-007, REQ-050 |
| HookSpecificOutputMismatch detection | `internal/hook/registry.go` (extend Dispatch) | REQ-040 |
| 64 KiB AdditionalContext 절단 + truncation systemMessage | `internal/hook/dual_parse.go` `ValidateHookResponse` (이미 90-115; 메시지 표준화) | REQ-022 |
| AdditionalContext 라우팅 (SessionStart/UserPromptSubmit/PreToolUse/PostToolUse) | `internal/hook/registry.go` Dispatch 후 effects 큐 | REQ-012, REQ-016 |
| UpdatedInput then PermissionDecision 순서 | `internal/hook/registry.go` (PreToolUse 경로 정리) | REQ-041 |
| WatchPaths 등록 hook (SessionStart) | `internal/hook/session_start.go` (extend) + `internal/watcher/` (out-of-scope, integration point만 본 SPEC) | REQ-032 |
| api_version 2 opt-in (`# moai-hook-api-version: 2`) | `internal/hook/api_version.go` (new) + `internal/hook/dual_parse.go` (참조) | REQ-030 |
| Plugin-source bypass (Source = "plugin" → strict-mode 우회) | `internal/hook/strict_mode.go` (new; SPEC-V3R2-RT-005 Source enum 사용) | REQ-051 |
| `HookConfig{StrictMode}` 필드 | `internal/config/types.go` (extend) + `internal/config/loader.go` 확장 | REQ-021 |
| `system.yaml` `hook.strict_mode` 키 로딩 | `internal/config/loader.go` (extend) + `.moai/config/sections/system.yaml` (template extend) | REQ-021 |
| AdditionalContext + @MX 마커 라우팅 hook (PostToolUse) | `internal/hook/post_tool.go` (extend) — 본 SPEC 은 라우팅 포인트만 노출; 의미론은 SPEC-V3R2-SPC-002 | REQ-016 |
| `ErrHookProtocolLegacyRejected` (이미 `dual_parse.go:11-13`), 에러 wiring | `internal/hook/dual_parse.go` (extend) | REQ-021 |
| `.moai/logs/hook.log` mismatch logging | `internal/hook/registry.go` (extend) | REQ-040, REQ-042 |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 6 files (per §6 below) | mx_plan |

`.claude/hooks/moai/*.sh` 26개 셸 래퍼는 본 SPEC 에서 **수정하지 않는다** — wrappers-unchanged 원칙. `make build` 로 `internal/template/embedded.go` 재생성은 필요할 수 있다 (template 변경 사항이 `.moai/config/sections/system.yaml` 에 추가되는 경우).

### 1.4 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task):

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-001-001 | Ubiquitous | AC-02 | T-RT001-04, T-RT001-12 |
| REQ-V3R2-RT-001-002 | Ubiquitous | AC-12 | T-RT001-04, T-RT001-13 |
| REQ-V3R2-RT-001-003 | Ubiquitous | AC-13 | T-RT001-05 (existing 27 variants) |
| REQ-V3R2-RT-001-004 | Ubiquitous | AC-13 | T-RT001-05 |
| REQ-V3R2-RT-001-005 | Ubiquitous | AC-01 | T-RT001-06, T-RT001-15 |
| REQ-V3R2-RT-001-006 | Ubiquitous | AC-12 | T-RT001-04, T-RT001-13 |
| REQ-V3R2-RT-001-007 | Ubiquitous | AC-04, AC-15 | T-RT001-08, T-RT001-17 |
| REQ-V3R2-RT-001-010 | Event-Driven | AC-01 | T-RT001-15 |
| REQ-V3R2-RT-001-011 | Event-Driven | AC-03 | T-RT001-15 |
| REQ-V3R2-RT-001-012 | Event-Driven | AC-02 | T-RT001-19 |
| REQ-V3R2-RT-001-013 | Event-Driven | AC-01 | T-RT001-19, T-RT001-20 |
| REQ-V3R2-RT-001-014 | Event-Driven | AC-08 | T-RT001-19 |
| REQ-V3R2-RT-001-015 | Event-Driven | (orchestrator-side) | T-RT001-19 |
| REQ-V3R2-RT-001-016 | Event-Driven | AC-07 | T-RT001-21 |
| REQ-V3R2-RT-001-020 | State-Driven | AC-04 | T-RT001-08 |
| REQ-V3R2-RT-001-021 | State-Driven | AC-06 | T-RT001-09, T-RT001-16 |
| REQ-V3R2-RT-001-022 | State-Driven | AC-09 | T-RT001-07, T-RT001-18 |
| REQ-V3R2-RT-001-030 | Optional | AC-11 | T-RT001-22 |
| REQ-V3R2-RT-001-031 | Optional | (orchestrator MAY) | T-RT001-23 |
| REQ-V3R2-RT-001-032 | Optional | AC-02 | T-RT001-24 |
| REQ-V3R2-RT-001-040 | Unwanted | AC-05 | T-RT001-10, T-RT001-20 |
| REQ-V3R2-RT-001-041 | Unwanted | AC-10 | T-RT001-20 |
| REQ-V3R2-RT-001-042 | Unwanted | (warn log) | T-RT001-20 |
| REQ-V3R2-RT-001-050 | Complex | AC-15 | T-RT001-08, T-RT001-17 |
| REQ-V3R2-RT-001-051 | Complex | AC-14 | T-RT001-25 |

Coverage: **25 REQs** (REQ-001..051 with the 5.x.0 numbering) → **15 ACs** → **25 tasks T-RT001-01..25** (some tasks cover multiple REQs and ACs).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone 은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

기존 테스트 baseline: `internal/hook/dual_parse_test.go` (304줄), `internal/hook/response_test.go` (240줄), `internal/hook/wire_format_freeze_test.go` (245줄), `internal/hook/types_test.go` (244줄), `internal/hook/protocol_test.go` (467줄), `internal/hook/registry_test.go` (388줄). 본 milestone 은 18개 새로운 실패 테스트만 추가.

Owner role: `expert-backend` (Go test) or direct `manager-tdd` execution.

Scope:
1. `internal/hook/dual_parse_test.go` 확장:
   - `TestParseHookOutput_StrictModeRejectsLegacy` (AC-06, REQ-021).
   - `TestParseHookOutput_MoaiHookLegacySuppressesBanner` (AC-04, REQ-020, REQ-007).
   - `TestParseHookOutput_DeprecationBannerOnce` (AC-15, REQ-007, REQ-050).
   - `TestParseHookOutput_AdditionalContextTruncatedAt64KiB` (AC-09, REQ-022).
   - `TestParseHookOutput_PluginSourceBypassesStrict` (AC-14, REQ-051).
   - `TestParseHookOutput_ApiVersion2SkipsExitCodeFallback` (AC-11, REQ-030).
2. `internal/hook/response_test.go` 확장:
   - `TestHookResponse_ValidatorRejectsBadPermissionDecision` (AC-12, REQ-006).
   - `TestHookResponse_ValidatorAcceptsAllowAskDeny` (AC-12, REQ-002).
   - `TestHookResponse_27VariantHookEventNameRoundTrip` (AC-13, REQ-003, REQ-004) — 27개 variant 의 marshal→unmarshal identity.
3. `internal/hook/registry_test.go` 확장:
   - `TestDispatch_HookSpecificOutputMismatch` (AC-05, REQ-040).
   - `TestDispatch_PreToolUseUpdatedInputThenPermissionDecision` (AC-10, REQ-041).
   - `TestDispatch_AdditionalContextRoutedToModelContext` (AC-02, REQ-012).
   - `TestDispatch_ContinueFalseBlocksTeammateIdle` (AC-08, REQ-014).
4. 신규 테스트 파일 `internal/hook/strict_mode_test.go` 생성:
   - `TestStrictMode_ConfigLoad` — `.moai/config/sections/system.yaml` `hook.strict_mode: true` 로딩 검증.
   - `TestStrictMode_ErrorReturnedOnLegacyOutput` (AC-06, REQ-021).
5. 신규 테스트 파일 `internal/hook/banner_test.go` 생성:
   - `TestBanner_OncePerSession` (REQ-007, REQ-050).
   - `TestBanner_SuppressedByMoaiHookLegacy` (REQ-020).
6. 신규 테스트 파일 `internal/hook/api_version_test.go` 생성:
   - `TestApiVersion2_SkipsExitCodeFallback` (REQ-030, AC-11).
   - `TestApiVersion2_ParsesShellHeader` — `# moai-hook-api-version: 2` line 추출.
7. `go test ./internal/hook/` 실행 — 본 milestone 에서 추가한 18개 테스트가 모두 RED 임을 확인. 기존 테스트 100% GREEN 유지 (regression baseline).

Verification gate before advancing to M2: 18개 새 테스트가 sentinel 메시지와 함께 실패. 기존 테스트 회귀 0.

[HARD] No implementation code in M1 outside of test files. (validator/v10 import 추가도 M2 의 일이다 — M1 테스트는 build error 가 발생할 수 있으며, 그것이 정확히 RED 의 정의다.)

### M2: Validator/v10 + HookResponse Validate() (GREEN, part 1) — Priority P0

Owner role: `expert-backend`.

Scope:
1. `go.mod` 에 `github.com/go-playground/validator/v10` 직접 의존성 추가 (`go get github.com/go-playground/validator/v10`). SPEC-V3R2-SCH-001 이 머지 전이므로 본 SPEC 가 직접 책임. SCH-001 이 후속 머지 시 충돌 없음 (동일 의존성).
2. `internal/hook/response.go` 의 `HookResponse` 구조체 (line 11-44) 에 validator/v10 태그 추가:
   - `PermissionDecision` → `validate:"omitempty,oneof=allow ask deny defer"`.
   - `Retry.Attempts` → `validate:"omitempty,gte=0,lte=10"`.
   - `Retry.Backoff` → `validate:"omitempty,oneofci=linear exponential constant"`.
3. `internal/hook/response.go` 에 패키지 단위 `var validate = validator.New()` 추가 + `func (r *HookResponse) Validate() error { return validate.Struct(r) }` 메서드 추가.
4. `internal/hook/dual_parse.go` 의 `ValidateHookResponse` (line 90-115) 를 강화: 기존 logic 보존 + 새로 추가된 `Validate()` 호출 wrapping. validator 에러는 offending field 명을 포함하여 wrap 한다 (AC-12).
5. 27개 variant 타입 (`response.go:76-287`) 에는 validator 태그가 필요 없다 (PreToolUseOutput 등은 외부 hook 가 emit 하는 nested struct 이며 본 layer 에서는 round-trip 만 검증 — `wire_format_freeze_test.go` 와 정합).

Verification: `TestHookResponse_ValidatorRejectsBadPermissionDecision` (AC-12) GREEN 전환. 기존 테스트 100% GREEN 유지.

### M3: Strict mode + MOAI_HOOK_LEGACY env + Deprecation Banner (GREEN, part 2) — Priority P0

Owner role: `expert-backend`.

Scope:
1. 신규 파일 `internal/hook/strict_mode.go`:
   - `IsStrictModeEnabled(cfg ConfigProvider) bool` — `.moai/config/sections/system.yaml` `hook.strict_mode` 값 조회.
   - `IsLegacyEnvSet() bool` — `MOAI_HOOK_LEGACY=1` 환경변수 검사.
   - `EnforceStrictMode(stdout []byte, exitCode int) error` — strict_mode 활성화 시 stdout JSON 파싱 실패면 `ErrHookProtocolLegacyRejected` (이미 `dual_parse.go:11-13`) 반환.
   - `IsPluginSource(input *HookInput) bool` — `HookInput.ConfigurationSource == "plugin"` 검사 (REQ-051; SPEC-V3R2-RT-005 Source provenance 와 정합).
2. 신규 파일 `internal/hook/banner.go`:
   - `package var emittedBanners sync.Map` — session_id key, bool value.
   - `func MaybeEmitDeprecationBanner(sessionID string) (banner string, emitted bool)` — 세션당 1회만 emit, MOAI_HOOK_LEGACY=1 시 suppress.
   - banner 텍스트: `"DEPRECATED: Hook produced exit-code-only output. Migrate to JSON output before v4.0. See https://moai.ai.kr/docs/hook-migration."` (영문 — sysmsg 는 international 사용자 노출).
3. `internal/hook/dual_parse.go` 의 `ParseHookOutput` (line 47-62) 를 확장:
   - JSON parse 실패 + strict_mode 활성화 + plugin-source 가 아닌 경우 → `ErrHookProtocolLegacyRejected` 반환.
   - JSON parse 실패 + 정상 fallback 경로 진입 시 banner emit 1회 (sessionID 기반).
   - `MOAI_HOOK_LEGACY=1` 시 banner suppress.
4. `internal/config/types.go` 에 `HookConfig` struct 추가 (`StrictMode bool` 필드, `mapstructure:"strict_mode"` 태그). `Config` 구조체에 `Hook HookConfig` 필드 추가.
5. `internal/config/loader.go` (또는 동등한 sections 로더) 가 `system.yaml` 의 `hook.strict_mode` 키를 `Config.Hook.StrictMode` 에 매핑하도록 확장. 기본값 `false`.
6. `.moai/config/sections/system.yaml` (현재 27 줄) 끝에 `hook.strict_mode: false` (default off) 키 추가. 템플릿도 동일 추가 (`internal/template/templates/.moai/config/sections/system.yaml`).

Verification: `TestStrictMode_*`, `TestBanner_*` 4개 테스트 GREEN 전환. AC-04, AC-06, AC-15 충족.

### M4: HookSpecificOutputMismatch + UpdatedInput-then-Decision + AdditionalContext routing + Continue:false escalation (GREEN, part 3) — Priority P0

Owner role: `expert-backend`.

Scope:
1. `internal/hook/registry.go` Dispatch (line 65-200) 에 mismatch detection 삽입:
   - PreToolUse / PostToolUse / UserPromptSubmit / PermissionRequest 핸들러 응답 후 `HookSpecificOutput.HookEventName != event` 검사.
   - 불일치 시 `HookSpecificOutputMismatch{Expected: event, Actual: hookEventName}` 에러 반환 + `.moai/logs/hook.log` append.
   - hook 자체는 failed 처리 (output 무시, default deny + reason="hookSpecificOutput mismatch"). REQ-040.
2. `internal/hook/registry.go` PreToolUse 경로 정리:
   - hook 가 `UpdatedInput` + `PermissionDecision: deny` 둘 다 emit 한 경우, `UpdatedInput` 을 먼저 적용한 후 `PermissionDecision` 평가 (REQ-041, AC-10).
   - 이미 `registry.go:179-193` 에 부분 핸들링이 있으나 ordering 명시 안 됨; 명시적 주석 + 단위 테스트 보강.
3. `internal/hook/registry.go` 에서 `AdditionalContext` 라우팅 강화:
   - SessionStart, UserPromptSubmit, PreToolUse, PostToolUse 응답의 `AdditionalContext` 를 model turn 의 system-context block 에 firing order 로 append.
   - 구현은 dispatch 결과 출력 시 `HookSpecificOutput.AdditionalContext` 필드를 보존 (이미 `registry.go:155-162` 에 일부 존재) + 외부 컨텍스트 큐(`internal/runtime/context.go`)와의 인터페이스만 노출 — 본 SPEC 은 노출까지, 큐 소비는 SPEC-V3R2-SPC-002.
4. `internal/hook/registry.go` Stop / SubagentStop 경로:
   - hook 가 `Continue: false` 반환 시 teammate idle 차단 + orchestrator 에 blocker 신호 전달 (REQ-014).
   - 이미 `subagent_stop.go` + `teammate_idle.go` 에 일부 처리 — 본 plan 은 mapping 명시 + 단위 테스트.
5. `internal/hook/dual_parse.go` `ValidateHookResponse` 의 64 KiB 절단 로직 (line 105-113) 표준화:
   - `SystemMessage` 에 prepend 하지 않고 별도 `\n[Notice: ...]` suffix 로 변경 (현재 동작 유지하되 메시지 텍스트를 `AdditionalContext truncated to 64 KiB budget` 로 정규화 — REQ-022, AC-09).

Verification: `TestDispatch_HookSpecificOutputMismatch`, `TestDispatch_PreToolUseUpdatedInputThenPermissionDecision`, `TestDispatch_AdditionalContextRoutedToModelContext`, `TestDispatch_ContinueFalseBlocksTeammateIdle` 4개 GREEN 전환. AC-02, AC-05, AC-08, AC-09, AC-10 충족.

### M5: api_version 2 opt-in + WatchPaths + @MX routing + CHANGELOG + MX tags (GREEN, part 4 + Trackable) — Priority P1

Owner role: `expert-backend` (구현), `manager-docs` (CHANGELOG, MX tags).

Scope:

#### M5a: api_version 2 opt-in (REQ-030, AC-11)

1. 신규 파일 `internal/hook/api_version.go`:
   - `func ReadHookApiVersion(wrapperPath string) (int, error)` — 셸 래퍼 첫 30 줄 내에서 `# moai-hook-api-version: 2` 패턴 추출 (정규식 `^#\s*moai-hook-api-version:\s*(\d+)\s*$`).
   - `func ShouldSkipExitCodeFallback(wrapperPath string) bool` — api_version >= 2 시 true.
2. `internal/hook/dual_parse.go` `ParseHookOutput` 를 확장하여 wrapperPath 인자 (또는 `HookInput` 의 새 필드 `WrapperPath`) 를 받아, `ShouldSkipExitCodeFallback` 가 true 이면 stdout 이 비었어도 exit-code synthesis 를 수행하지 않고 explicit no-op `HookResponse{}` 를 반환.
3. shell wrapper 자체는 수정하지 않는다 (wrappers-unchanged 정책). api_version 2 는 future-proofing — 현재 26개 wrapper 모두 api_version 1 로 간주된다.

#### M5b: WatchPaths SessionStart wiring (REQ-032, AC-02 partial)

1. `internal/hook/session_start.go` 가 `HookResponse.WatchPaths` 를 emit 하면, `internal/watcher/` (out-of-scope; integration point만 제공) 가 file-system watch 를 등록하고 변경 시 `EventFileChanged` 발화. 본 SPEC 은 인터페이스 (`WatchPathsRegistrar`) 만 정의:
   - 신규 파일 `internal/hook/watch_paths.go` 에서 `type WatchPathsRegistrar interface { Register(paths []string) error }` 선언.
   - `registry.go` Dispatch 의 SessionStart 후처리에서 `WatchPaths` 를 `WatchPathsRegistrar` 에 forward.
2. 실제 `WatchPathsRegistrar` 의 구현 (file-system watcher with debouncing) 은 별도 SPEC (RT-006 또는 새 RT-008) 에서 다룬다 — 본 SPEC 은 contract만 노출.

#### M5c: @MX 마커 라우팅 hook (REQ-016)

1. `internal/hook/post_tool.go` 의 PostToolUse 핸들러에서 `HookResponse.AdditionalContext` 를 검사하여 `@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`, `@MX:TODO`, `@MX:LEGACY` 마커가 포함된 경우 `internal/hook/mx/ingest.go` (이미 존재 — `mx/` 디렉토리에 12 파일) 의 ingestion 함수에 forward.
2. 본 SPEC 은 라우팅 포인트(integration hook)만 노출; 마커 의미론은 SPEC-V3R2-SPC-002.

#### M5d: ConfigurationSource = "plugin" bypass (REQ-051, AC-14)

1. `internal/hook/strict_mode.go` 의 `EnforceStrictMode` 에서 input.ConfigurationSource 가 `"plugin"` 인 경우 strict_mode 가 true 라도 dual-parse fallback 적용 (no error). REQ-051 BC 윈도 제공.
2. 로그에 `source: plugin` provenance 기록 (SPEC-V3R2-RT-005 와 정합).

#### M5e: CHANGELOG + MX tags + 최종 검증

1. `CHANGELOG.md` 의 `## [Unreleased] / ### Changed (BREAKING)` 섹션에 entry 추가:
   ```
   ### Changed (BREAKING — BC-V3R2-001)
   - SPEC-V3R2-RT-001: Hook output protocol now follows Claude Code 2026.x JSON-OR-ExitCode dual contract. JSON output is parsed first; exit-code semantics remain as v3.x deprecation-window fallback. New `hook.strict_mode` config (default `false`) opts into JSON-only enforcement. New `MOAI_HOOK_LEGACY=1` opt-out for CI/air-gapped installs. Deprecation banner emits once per session per project. Wrappers unchanged; handlers rewritten. Removal of exit-code-only path deferred to v4.0.
   ```
2. §6 의 7개 MX tag 삽입.
3. `make build` 를 worktree root 에서 실행하여 `internal/template/embedded.go` 재생성. system.yaml template 1줄 추가만 expected.
4. `go test ./internal/hook/...` + 전체 `go test ./...` 실행 — 모든 테스트 PASS, 회귀 0.
5. `go vet ./...` + `golangci-lint run` 실행 — 경고 0.
6. `progress.md` 업데이트: `run_complete_at` + `run_status: implementation-complete`.

[HARD] No new SPEC documents are created in `.moai/specs/` or `.moai/reports/` during M5 — this is a SPEC implementation phase, not a planning phase.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `internal/hook/response.go:11-44` | `HookResponse` struct | validator/v10 태그 추가 + `Validate()` 메서드 | M2 / REQ-006 |
| `internal/hook/response.go:64-70` | `RetryHint` struct | validator 태그 추가 (Attempts gte/lte, Backoff oneofci) | M2 / REQ-031 |
| `internal/hook/dual_parse.go:11-19` | error sentinel block | `ErrHookProtocolLegacyRejected` 보존, mismatch 에러는 그대로 사용 | M3 / REQ-021 |
| `internal/hook/dual_parse.go:47-62` | `ParseHookOutput()` | strict_mode 분기 + plugin-source bypass + banner emission | M3 / REQ-005, REQ-021, REQ-007, REQ-050, REQ-051 |
| `internal/hook/dual_parse.go:90-115` | `ValidateHookResponse()` | 64 KiB truncation 메시지 표준화 + validator 통합 | M2/M4 / REQ-022, REQ-006 |
| `internal/hook/dual_parse.go:117-189` | `ToHookOutput`/`ToHookResponse` | 호환성 shim 보존; PostToolUse output mapping 보강 | M4 / REQ-016 |
| `internal/hook/registry.go:65-200` | `Dispatch()` | HookSpecificOutputMismatch detection + AdditionalContext routing + Continue:false escalation | M4 / REQ-040, REQ-012, REQ-014 |
| `internal/hook/registry.go:155-162` | `AdditionalContext` merge logic | order preservation, append semantics | M4 / REQ-012 |
| `internal/hook/registry.go:179-193` | PreToolUse PermissionDecision/Reason wiring | UpdatedInput-first ordering 명시 + 주석 | M4 / REQ-041 |
| `internal/hook/types.go:283-312` | `HookOutput` struct | 보존 (legacy bridge); 변경 없음 — 단지 docstring 에 BC-V3R2-001 reference 추가 | doc-only / BC |
| `internal/hook/post_tool.go` (existing PostToolUse handler) | `Handle()` 내부 | `AdditionalContext` 의 @MX 마커 라우팅 hook 추가 | M5c / REQ-016 |
| `internal/hook/session_start.go` (existing SessionStart handler) | `Handle()` 내부 | `WatchPaths` forward to `WatchPathsRegistrar` | M5b / REQ-032 |
| `internal/config/types.go` (top-level Config struct) | new field | `Hook HookConfig{StrictMode bool}` 필드 추가 | M3 / REQ-021 |
| `internal/config/loader.go` (sections loader) | section dispatch | `system.yaml` `hook.strict_mode` 키 매핑 | M3 / REQ-021 |
| `.moai/config/sections/system.yaml` | end of file | `hook.strict_mode: false` 키 추가 | M3 / REQ-021 |
| `internal/template/templates/.moai/config/sections/system.yaml` | end of file | 동일 키 추가 (template 동기화) | M3 / Template-First |
| `CHANGELOG.md` | `## [Unreleased] / ### Changed (BREAKING)` | BC-V3R2-001 entry 추가 | M5e / Trackable |

### 3.2 To-be-created (new files)

| File | Reason | LOC estimate |
|------|--------|--------------|
| `internal/hook/strict_mode.go` | strict_mode + MOAI_HOOK_LEGACY + plugin-source 검사 함수군 | ~70 |
| `internal/hook/strict_mode_test.go` | strict_mode 테스트 | ~80 |
| `internal/hook/banner.go` | deprecation banner once-per-session | ~50 |
| `internal/hook/banner_test.go` | banner 테스트 | ~70 |
| `internal/hook/api_version.go` | api_version 2 opt-in 파싱 | ~50 |
| `internal/hook/api_version_test.go` | api_version 테스트 | ~60 |
| `internal/hook/watch_paths.go` | `WatchPathsRegistrar` interface + nil-safe wiring | ~30 |

Total new: ~410 LOC. Total modified: ~150 LOC. Net additions: ~560 LOC across 7 new files + 8 modified Go files + 2 modified template files + 1 CHANGELOG.

### 3.3 NOT to be touched (preserved by reference)

본 SPEC 의 run phase 가 절대 수정하지 않을 파일들. 다른 SPEC 의 소유 영역이거나 wrappers-unchanged 원칙에 따라 보존.

- `.claude/hooks/moai/*.sh` — 26개 셸 래퍼 모두. wrappers-unchanged. master §8 BC-V3R2-001 의 핵심 호환성 약속.
- `internal/hook/types.go` 의 `HookInput`, `HookOutput`, `HookSpecificOutput` 외부 visible API — legacy bridge 로 보존; docstring 외 변경 없음.
- `.claude/rules/moai/core/agent-common-protocol.md` — load-bearing rule.
- 27개 hook handler 의 외부 인터페이스 (`Handler.Handle(ctx, input) (*HookOutput, error)`) — variant types 추가로도 Handler 인터페이스는 보존.
- `.moai/specs/SPEC-V3R2-RT-002/` (permission stack) — 본 SPEC 은 protocol 만, RT-002 가 PermissionDecision 의미론 처리.
- `.moai/specs/SPEC-V3R2-RT-005/` (settings provenance) — Source = "plugin" 검사는 RT-005 의 enum 을 import 만; 본 SPEC 은 RT-005 머지 전이라도 string literal `"plugin"` 으로 동작.
- `.moai/specs/SPEC-V3R2-RT-006/` (handler completeness) — 10 logging-only stubs 의 비즈니스 로직은 본 SPEC 범위 밖.
- `.moai/specs/SPEC-V3R2-SPC-002/` (@MX 의미론) — 본 SPEC 은 라우팅 포인트만 노출.

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability and research.md §10, the following anchors are load-bearing and cited verbatim throughout this plan:

1. `spec.md:43-55` (in-scope items 1-9: typed HookResponse, PermissionDecision, HookSpecificOutput discriminator, dual-parse shim, validator 통합, context-injection wiring, input-mutation wiring, continue:false escalation, deprecation banner)
2. `spec.md:57-66` (out-of-scope items 1-8)
3. `spec.md:106-145` (24 EARS REQs — REQ-001 through REQ-051)
4. `spec.md:147-163` (15 ACs — AC-01 through AC-15)
5. `spec.md:167-173` (constraints — Go 1.22+, validator/v10, p99 5ms, binary +250 KiB, wrappers unchanged)
6. `internal/hook/types.go:1-14` (package import 영역 — 본 SPEC 은 import 변경 없음)
7. `internal/hook/types.go:115-147` (`ValidEventTypes()` 27 events 열거 — REQ-003 의 27 variant 와 1:1 대응)
8. `internal/hook/types.go:155-165` (PermissionDecision constants — REQ-002 와 정합)
9. `internal/hook/types.go:269-281` (`HookSpecificOutput` struct — discriminator field)
10. `internal/hook/types.go:283-312` (`HookOutput` struct — legacy bridge, 보존)
11. `internal/hook/response.go:11-44` (`HookResponse` struct — validator 태그 추가 대상)
12. `internal/hook/response.go:46-61` (`PermissionDecision` constants — `defer` 포함 4개 — Defer 는 v2.1.89+ headless 전용)
13. `internal/hook/response.go:64-70` (`RetryHint` struct — Attempts/Backoff)
14. `internal/hook/response.go:76-287` (27 variant types 의 `HookEventName()` 메서드 — REQ-004)
15. `internal/hook/dual_parse.go:11-19` (error sentinels: `ErrHookProtocolLegacyRejected`, `ErrHookInvalidPermissionDecision`, `ErrHookSpecificOutputMismatch`)
16. `internal/hook/dual_parse.go:47-62` (`ParseHookOutput()` dual-parse 진입점)
17. `internal/hook/dual_parse.go:64-86` (`synthesizeFromExitCode()` legacy fallback)
18. `internal/hook/dual_parse.go:88-115` (`ValidateHookResponse()` 64 KiB 절단)
19. `internal/hook/dual_parse.go:117-189` (`ToHookOutput`/`ToHookResponse` 호환성 shim)
20. `internal/hook/protocol.go:26-47` (`ReadInput()` JSON 정규화 진입점)
21. `internal/hook/protocol.go:52-63` (`WriteOutput()` JSON serialization)
22. `internal/hook/registry.go:65-200` (`Dispatch()` 27 events handler 디스패치 — mismatch detection 삽입 위치)
23. `internal/hook/registry.go:155-162` (`AdditionalContext` merge logic in UserPromptSubmit)
24. `internal/hook/registry.go:179-193` (PreToolUse `PermissionDecision`/`Reason` wiring)
25. `.claude/hooks/moai/*.sh` — 26 wrappers (verified via `ls`); wrappers-unchanged
26. `.moai/config/sections/system.yaml:1-27` — 본 SPEC 이 `hook.strict_mode: false` 키 추가
27. `.claude/rules/moai/core/agent-common-protocol.md:#user-interaction-boundary` — subagent prohibition rule
28. `.claude/rules/moai/workflow/spec-workflow.md:#phase-0-5-plan-audit-gate` — plan-auditor entry condition
29. `CLAUDE.local.md:§6` — test isolation (`t.TempDir()` + `filepath.Abs`)
30. `CLAUDE.local.md:§14` — no hardcoded paths in `internal/`

Total: **30 distinct file:line anchors** (exceeds the §Hard-Constraints minimum of 10 for plan.md).

---

## 4. Technology Stack Constraints

Per `spec.md` §7 Constraints, **새로운 의존성은 validator/v10 1개만**:

- Go 1.22+ (already required by `go.mod`).
- `github.com/go-playground/validator/v10` — SPEC-V3R2-SCH-001 미머지 시 본 SPEC 가 직접 추가. 현재 `go.mod` 부재 확인 (위 `grep -n validator go.mod` 결과). M2 의 첫 작업이 `go get github.com/go-playground/validator/v10`.
- `encoding/json` 표준 라이브러리 — `json.RawMessage` 가 `HookSpecificOutput` discriminator 의 two-pass decode 를 처리.
- `sync.Map` 표준 라이브러리 — banner once-per-session 의 thread-safe 저장소.
- 기존 `golang.org/x/sys/{unix,windows}` — 본 SPEC 은 사용하지 않음.

추가 surface:

- 7개 신규 Go 파일 (per §3.2).
- 8개 수정 Go 파일 (per §3.1).
- 2개 수정 YAML template (system.yaml + 동일 template).
- 1개 CHANGELOG entry.
- 7개 MX tag (per §6).

**No new directory structures** — `internal/hook/` 는 이미 116 파일.

**No new shell wrapper** — wrappers-unchanged 정책.

---

## 5. Risk Analysis & Mitigations

`spec.md` §8 의 6개 risk 를 file-anchored mitigation 으로 확장.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| `validator/v10` SPEC-V3R2-SCH-001 머지 전 — 직접 의존성 추가 시 충돌 가능성 | M | M | M2 첫 단계에서 `go get` 직접 추가. 후속 SCH-001 머지 시 `go.mod` 의 동일 의존성은 자동 통합 (Go module dedup). 충돌 발생 시 SCH-001 PR 의 위치를 우선시한다. |
| 외부 plugin-hook 작성자가 invalid JSON 을 emit 하고 exit-code fallback 에 무한 의존 | M | M | dual-parse 는 v3.x 전체 minor cycle 동안 유지; `MOAI_HOOK_LEGACY=1` opt-out for CI/air-gapped; `hook.strict_mode: true` opt-in for early rejection. spec §8 risk row 1 + REQ-051 plugin-source bypass. |
| Discriminated-union mismatch 에러 메시지 가독성 부족 | M | M | REQ-040 + AC-05: `HookSpecificOutputMismatch{Expected, Actual}` struct (이미 `dual_parse.go:23-30`); error 메시지는 `expected X, got Y` 포맷. mismatch 시 `.moai/logs/hook.log` append. `moai doctor hook --validate` 향후 surface. |
| 64 KiB AdditionalContext 절단이 hook 작성자에게 silent | L | M | REQ-022 + AC-09: 절단 시 `SystemMessage` 에 `[Notice: additionalContext truncated from N to 65536 bytes]` suffix 자동 추가. 사용자 가시. |
| PreToolUse 의 `UpdatedInput` 과 `PermissionDecision` 의 ordering 모호성 | L | M | REQ-041 + AC-10: deterministic order — UpdatedInput first, then PermissionDecision against the updated input. `registry.go` 주석 + 단위 테스트 명시. |
| 27 logging-only handler 가 protocol upgrade 후에도 빈 응답 | M | L | SPEC-V3R2-RT-006 explicitly enumerates 27-event business-logic 처리; 본 SPEC 은 wire format 만 owner. logging-only handler 는 빈 `HookResponse{}` (allow + continue) 반환 — 정상 동작. |
| Shell wrapper JSON forwarding 이 non-UTF-8 바이트에서 깨짐 | L | L | wrappers 는 `cat` 의미론 사용; validator/v10 는 non-UTF-8 string 을 명시적 에러로 reject. 본 SPEC 의 ValidateHookResponse 가 fail-fast. |
| Banner emission race condition (multi-threaded dispatch) | L | L | `sync.Map` 사용; `LoadOrStore` atomic check. M3 banner_test.go 가 verify. |
| SPEC-V3R2-RT-002 (permission stack) 의 PermissionDecision consumer 가 본 SPEC 과 동시 개발 | L | L | `PermissionDecision` 은 string enum (`response.go:46-61`); RT-002 는 consumer 일 뿐 본 SPEC 의 type 정의에 의존. 동시 개발 가능. |
| api_version 2 wrapper 가 현재 0개이므로 코드가 dead path | L | L | future-proofing; v3.1+ 에서 첫 wrapper 적용 시 활성화. M5a 테스트는 mock wrapper 로 검증. |
| Plugin-source bypass 가 RT-005 미머지 시 ConfigurationSource 필드 부재 | L | M | `HookInput.ConfigurationSource` 는 이미 `types.go:235-237` 에 정의됨 (v2.1.49+). RT-005 의존성은 enum 값 의미론만이며 string literal `"plugin"` 검사로 충족. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/hook/dual_parse.go:ParseHookOutput()` (line 47) | `@MX:ANCHOR fan_in=N - SPEC-V3R2-RT-001 BC-V3R2-001 dual-parse contract; every hook handler routes through this function. Modifying the JSON-first then exit-code-fallback ordering breaks the v3.x deprecation window. Touching this affects all 27 events.` | dual-parse 는 본 SPEC 의 핵심 진입점. 27개 이벤트 전부의 wire format 이 여기를 통과. |
| `internal/hook/response.go:HookResponse` (line 11) | `@MX:ANCHOR fan_in=27 - SPEC-V3R2-RT-001 typed response struct; matches Claude Code 2026.x HookJSONOutput byte-for-byte. Field name/JSON tag changes break protocol compatibility with Claude Code runtime.` | HookResponse 는 27개 variant 가 공유하는 top-level type. JSON tag 변경은 wire format break. |
| `internal/hook/registry.go:Dispatch()` (line 65) | `@MX:ANCHOR fan_in=N - SPEC-V3R2-RT-001 27-event dispatch contract; HookSpecificOutputMismatch detection + AdditionalContext routing + Continue:false escalation 모두 본 함수에서 enforce. Ordering of these checks is contract.` | Dispatch 는 27개 이벤트가 통과하는 단일 게이트. 검사 순서가 contract. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/hook/strict_mode.go:EnforceStrictMode()` (new file) | `@MX:NOTE - SPEC-V3R2-RT-001 strict_mode 는 .moai/config/sections/system.yaml hook.strict_mode 키를 참조. plugin-source 는 BC-V3R2-001 마이그레이션 윈도 내내 우회 (REQ-051). MOAI_HOOK_LEGACY=1 은 CI/air-gapped 용 opt-out.` | 미래 contributors 가 strict_mode 의도 를 이해할 수 있도록 의도 명시. |
| `internal/hook/banner.go:MaybeEmitDeprecationBanner()` (new file) | `@MX:NOTE - SPEC-V3R2-RT-001 banner 는 session_id 기준 once-per-session. v4.0 에서 exit-code fallback 제거 시 본 banner 도 동시 제거 예정. v3.x 전체 cycle 동안만 존재.` | banner 의 lifecycle (v3.x sunset) 명시. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/hook/dual_parse.go:synthesizeFromExitCode()` (line 64) | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-001 BC-V3R2-001 legacy fallback. v3.x 전체 deprecation window 동안 작동해야 함; v4.0 에서 함수 자체 제거 예정. exit code 0/2/non-zero 의미론 변경은 26개 wrapper 의 동작을 모두 깨뜨림.` | legacy fallback 은 v4.0 까지 보존되어야 하는 호환성 약속. 변경 시 BC 위반. |
| `internal/hook/registry.go:HookSpecificOutputMismatch detection` (line ~150 — to be inserted in M4) | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-001 REQ-040 mismatch detection. mismatch 시 hook 은 failed 처리 (output 무시 + default deny). 본 검사를 비활성화하면 attacker-controlled hook 이 다른 이벤트의 PermissionDecision 을 위조할 수 있음.` | security boundary. mismatch detection 이 위조 방지의 1차 방어선. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

본 SPEC 은 audit-ready 한 dual-protocol subsystem 을 완성한다. M1-M5 사이클이 모두 GREEN 으로 수렴; `@MX:TODO` 는 0개. 구현 중 도입된 `@MX:TODO` 는 GREEN-phase 종료 전 모두 해소 (per `.claude/rules/moai/workflow/mx-tag-protocol.md`).

### 6.5 MX tag count summary

- @MX:ANCHOR: 3 targets (`dual_parse.go:47`, `response.go:11`, `registry.go:65`)
- @MX:NOTE: 2 targets (`strict_mode.go`, `banner.go`)
- @MX:WARN: 2 targets (`dual_parse.go:64`, `registry.go` mismatch insertion)
- @MX:TODO: 0 targets
- **Total**: 7 MX tag insertions planned across 6 distinct files (dual_parse.go 가 2회 — line 47 + line 64)

---

## 7. Worktree Mode Discipline

[HARD] All run-phase work for SPEC-V3R2-RT-001 executes in:

```
/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001
```

Branch: `plan/SPEC-V3R2-RT-001` (already checked out per session context). Run-phase agent 는 동일 브랜치에서 계속하거나 sibling `feat/SPEC-V3R2-RT-001-dual-protocol` 으로 분기 (per `CLAUDE.local.md` §18.2 branch naming).

[HARD] Worktree is used for this SPEC (per session context). All Read/Write/Edit tool invocations use absolute paths under the worktree root.

[HARD] `make build` 및 `go test ./...` 은 worktree root 에서 실행: `cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001 && make build && go test ./...`.

> Note: Run-phase agent 는 실제 worktree cwd 에서 동작; 절대 경로는 reference 용. worktree-root 는 run time `git -C <worktree> rev-parse --show-toplevel` 으로 resolve.

Base branch HEAD: `496595c3f` (verified via `git -C <worktree> rev-parse HEAD`).

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:#phase-0-5-plan-audit-gate`). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.1.0 schema** — `spec.md:1-23` frontmatter has all required fields (`id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `dependencies`, `bc_id`, `related_principle`, `related_pattern`, `related_problem`, `related_theme`, `breaking`, `lifecycle`, `tags`).
- [x] **C2: HISTORY entry for v0.1.0** — `spec.md:29-31` HISTORY table has v0.1.0 row with description.
- [x] **C3: 24+ EARS REQs across 6 categories** — `spec.md:104-145` (Ubiquitous 7, Event-Driven 7, State-Driven 3, Optional 3, Unwanted 3, Complex 2 = **25 REQs**).
- [x] **C4: 15 ACs all map to REQs (100% coverage)** — `spec.md:147-163`. Each AC explicitly cites the REQ(s) it maps to. Plan §1.4 traceability matrix confirms 25 REQ → 15 AC → 25 task mapping. Bidirectional coverage: every REQ has at least 1 AC; every AC has at least 1 task.
- [x] **C5: BC scope clarity** — `spec.md:20` (`breaking: true`) + `spec.md:15` (`bc_id: [BC-V3R2-001]`) + spec.md §1 의 wrappers-unchanged + handlers-rewritten 호환성 shim 명시. 본 plan §1.1 + §3.3 에서 추가 강조.
- [x] **C6: File:line anchors ≥10** — research.md §11 (29 anchors), this plan.md §3.4 (30 anchors). Both exceed minimum.
- [x] **C7: Exclusions section present** — `spec.md:57-66` Out-of-scope (8 entries explicitly mapped to other SPECs: RT-006, RT-002, RT-003, RT-004, RT-005, RT-007, v3.1+ deferred, SPC-002).
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml` `development_mode: tdd`.
- [x] **C9: mx_plan section** — this plan §6 (7 MX tag insertions across 4 categories: 3 ANCHOR + 2 NOTE + 2 WARN + 0 TODO).
- [x] **C10: Risk table with mitigations** — `spec.md:175-184` (6 risks) + this plan §5 (11 risks, file-anchored mitigations).
- [x] **C11: Worktree mode path discipline** — this plan §7 (3 HARD rules, worktree-mode per session context, base HEAD `496595c3f`).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, code-block templates, and pseudo-Go for interface declarations. No executable Go function bodies.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md §AC-01..AC-15 (15 ACs, each with happy path + 1-3 edge cases + test mapping).
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md §M1-M5 (manager-tdd / expert-backend / manager-docs assignments). M1 RED phase = expert-backend test-only; M2-M5 GREEN/REFACTOR = expert-backend impl + manager-docs trackability.
- [x] **C15: Cross-SPEC consistency** — blocked-by dependencies verified: SPEC-V3R2-SCH-001 (validator/v10 — at-risk per §5; mitigation in M2 첫 작업), SPEC-V3R2-CON-001 (FROZEN zone — completed per Wave 6 history), SPEC-V3R2-RT-005 (Source provenance — referenced not blocking; string literal `"plugin"` 으로 충족). RT-001 blocks RT-002 (PermissionDecision consumer), RT-003 (sandbox via UpdatedInput), RT-006 (handler completeness), SPC-002 (@MX routing), HRN-002 (evaluator fresh-memory injection) per `spec.md` §9.2.

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent 가 순서대로 실행 (P0 first, dependencies resolved):

1. **M1 (P0)**: 18개 새 테스트 추가 — `dual_parse_test.go` 6개 확장 + `response_test.go` 3개 확장 + `registry_test.go` 4개 확장 + 신규 `strict_mode_test.go`/`banner_test.go`/`api_version_test.go` 5개. 모두 RED 확인 + 기존 테스트 GREEN 회귀 0.
2. **M2 (P0)**: `go get github.com/go-playground/validator/v10` + `response.go` validator 태그 + `Validate()` 메서드 + `dual_parse.go` ValidateHookResponse 강화. AC-12 GREEN.
3. **M3 (P0)**: 신규 `strict_mode.go`/`banner.go` + `dual_parse.go` strict_mode 분기 + banner emission + `internal/config/types.go` HookConfig + `loader.go` 확장 + `system.yaml` template 동기화. AC-04, AC-06, AC-15 GREEN.
4. **M4 (P0)**: `registry.go` mismatch detection + UpdatedInput-then-Decision ordering + AdditionalContext routing + Continue:false escalation. AC-02, AC-05, AC-08, AC-09, AC-10 GREEN.
5. **M5 (P1)**: 신규 `api_version.go`/`watch_paths.go` + `post_tool.go` @MX 라우팅 + `session_start.go` WatchPaths forward + `strict_mode.go` plugin-source bypass + CHANGELOG entry + 7 MX tags + final `make build` + `go test ./...` + `go vet` + `golangci-lint`. AC-01, AC-03, AC-07, AC-11, AC-13, AC-14 GREEN. `progress.md` `run_complete_at` 기입.

Total milestones: 5. Total file edits (existing): ~8 Go + 2 YAML + 1 CHANGELOG. Total file creations (new): 7. Total CHANGELOG entries: 1 (BREAKING under BC-V3R2-001). Total MX tag insertions: 7.

---

End of plan.md.
