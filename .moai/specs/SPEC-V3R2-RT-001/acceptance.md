# SPEC-V3R2-RT-001 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 15 ACs (AC-V3R2-RT-001-01 through -15)     |

---

## Scope

`spec.md` §6 의 15개 AC 를 Given/When/Then 형식으로 변환. 각 AC 는 happy-path + 1-3개 edge case + Go 테스트 함수 매핑을 동반.

표기:
- **Test mapping**: AC 를 검증하는 Go 테스트 함수 (또는 수동 검증 단계).
- **Sentinel**: 부정 경로에서 테스트가 기대하는 정확한 에러 문자열.

---

## AC-V3R2-RT-001-01 — PreToolUse JSON 응답이 UpdatedInput + PermissionDecision 을 적용

Maps to: REQ-V3R2-RT-001-010, REQ-V3R2-RT-001-013.

### Happy path

- **Given** PreToolUse hook wrapper 가 stdout 으로 다음 JSON 을 출력한다:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"file_path":"/tmp/x"}}}
  ```
- **When** registry 의 Dispatch 가 PreToolUse 이벤트를 라우팅
- **Then** ParseHookOutput 이 JSON parse 성공 → exit-code 무시
- **And** `HookResponse.HookSpecificOutput` 에 nested PreToolUseOutput 정보 보존
- **And** pending tool input 이 `{"file_path":"/tmp/x"}` 로 교체된 후 tool dispatch
- **And** SPEC-V3R2-RT-002 의 permission stack resolver 가 `permissionDecision: allow` 수신

### Edge case — JSON parse 성공 시 exit code 무시

- **Given** 동일 JSON + wrapper exit code 7 (의미 있는 값)
- **When** ParseHookOutput 처리
- **Then** exit code 7 은 완전히 무시됨; JSON 의 의미가 우선

### Test mapping

- `internal/hook/dual_parse_test.go::TestParseHookOutput_PreToolUseJSON` (existing baseline 확장)
- `internal/hook/registry_test.go::TestDispatch_PreToolUseUpdatedInput` (new, M1)

---

## AC-V3R2-RT-001-02 — SessionStart 이 AdditionalContext + WatchPaths 응답

Maps to: REQ-V3R2-RT-001-001, REQ-V3R2-RT-001-012, REQ-V3R2-RT-001-032.

### Happy path

- **Given** SessionStart hook 이 다음 JSON 을 출력:
  ```json
  {"additionalContext":"ctx","watchPaths":["/abs/.env"]}
  ```
- **When** ParseHookOutput 이 파싱
- **Then** `HookResponse.AdditionalContext == "ctx"`
- **And** `HookResponse.WatchPaths == ["/abs/.env"]`
- **And** `Validate()` 통과 (validator/v10 schema 위반 없음)
- **And** AdditionalContext 가 다음 model turn 의 system-context block 에 append
- **And** WatchPaths 가 `WatchPathsRegistrar.Register([]string)` 으로 forward

### Edge case — WatchPathsRegistrar 미등록 시 graceful

- **Given** 동일 JSON 출력
- **And** registry 에 `WatchPathsRegistrar` 가 nil
- **When** Dispatch 처리
- **Then** WatchPaths 는 logged but not registered (no panic)
- **And** AdditionalContext routing 은 정상 동작 (다른 필드는 영향 없음)

### Edge case — 빈 WatchPaths

- **Given** `{"additionalContext":"ctx","watchPaths":[]}`
- **When** Dispatch 처리
- **Then** `WatchPathsRegistrar.Register` 는 호출되지 않음 (zero-length slice skip)

### Test mapping

- `internal/hook/dual_parse_test.go::TestParseHookOutput_SessionStartContextAndWatchPaths` (new, M1)
- `internal/hook/registry_test.go::TestDispatch_AdditionalContextRoutedToModelContext` (new, M1)

---

## AC-V3R2-RT-001-03 — Legacy hook exit code 2 → deny + stderr reason

Maps to: REQ-V3R2-RT-001-011.

### Happy path

- **Given** legacy hook wrapper 가 stdout 빈 출력 + exit code 2 + stderr `"blocked"`
- **When** ParseHookOutput 처리
- **Then** dual-parse fallback 진입 (JSON empty)
- **And** `synthesizeFromExitCode(2, "blocked")` 호출
- **And** `HookResponse.PermissionDecision == "deny"`
- **And** `HookResponse.SystemMessage == "blocked"` (현재 `dual_parse.go:75-78` 동작)

### Edge case — exit code 1 (기타 non-zero)

- **Given** stdout 빈 출력 + exit code 1 + stderr `"transient error"`
- **When** ParseHookOutput 처리
- **Then** `HookResponse.SystemMessage == "transient error"` (PermissionDecision 미설정 — "no opinion")

### Edge case — exit code 0 + stderr (warning)

- **Given** stdout 빈 출력 + exit code 0 + stderr `"warning"` (의미상 무해)
- **When** ParseHookOutput 처리
- **Then** `HookResponse{}` (empty 응답, allow + continue)
- **And** stderr 는 무시되거나 별도 로그 (현재 동작 유지)

### Test mapping

- `internal/hook/dual_parse_test.go::TestSynthesizeFromExitCode_Code2WithStderr` (existing)
- `internal/hook/dual_parse_test.go::TestSynthesizeFromExitCode_Code1WithStderr` (existing)

---

## AC-V3R2-RT-001-04 — MOAI_HOOK_LEGACY=1 이 deprecation banner 억제

Maps to: REQ-V3R2-RT-001-020, REQ-V3R2-RT-001-007.

### Happy path

- **Given** 환경변수 `MOAI_HOOK_LEGACY=1` 설정
- **And** legacy wrapper 가 exit-code-only 출력
- **When** ParseHookOutput 첫 호출 (세션 시작 후 첫 legacy)
- **Then** 본 세션에서 deprecation banner 가 emit 되지 않음
- **And** dual-parse fallback 은 정상 동작 (exit code 의미 보존)
- **And** subsequent legacy hook 도 banner 미발사

### Edge case — MOAI_HOOK_LEGACY=0 또는 미설정

- **Given** 환경변수 `MOAI_HOOK_LEGACY` 미설정
- **And** 동일 legacy wrapper
- **When** ParseHookOutput 첫 호출
- **Then** banner 가 1회 emit (SystemMessage 또는 별도 채널)
- **And** subsequent legacy hook 은 banner 미반복 (REQ-050 once-per-session)

### Edge case — MOAI_HOOK_LEGACY=true 또는 yes (loose truthy)

- **Given** 환경변수 `MOAI_HOOK_LEGACY=true`
- **When** banner 검사
- **Then** banner suppressed (loose truthy 인정 — `IsLegacyEnvSet` 가 `"1"`, `"true"`, `"yes"` 모두 true 반환).

### Test mapping

- `internal/hook/banner_test.go::TestBanner_SuppressedByMoaiHookLegacy` (new, M1)
- `internal/hook/strict_mode_test.go::TestIsLegacyEnvSet_LooseTruthy` (new, M1)

---

## AC-V3R2-RT-001-05 — HookSpecificOutput.HookEventName 불일치 시 mismatch 에러

Maps to: REQ-V3R2-RT-001-040.

### Happy path

- **Given** PreToolUse hook 이 다음 JSON 출력:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PostToolUse","permissionDecision":"allow"}}
  ```
- **And** registry 가 PreToolUse 이벤트 dispatch 중
- **When** Dispatch 가 응답 검증
- **Then** `HookSpecificOutputMismatch{Expected: "PreToolUse", Actual: "PostToolUse"}` 에러 반환
- **And** `.moai/logs/hook.log` 에 mismatch entry append:
  ```
  HOOK_MISMATCH event=PreToolUse actual=PostToolUse session=<id> ts=<ts>
  ```
- **And** hook 이 failed 처리 — output 무시 + default deny + reason="hookSpecificOutput mismatch"

### Edge case — HookEventName 빈 값 (legacy hook)

- **Given** `{"hookSpecificOutput":{"hookEventName":"","permissionDecision":"allow"}}`
- **When** Dispatch 검증
- **Then** mismatch 처리 skip (빈 문자열은 "no opinion"; legacy compatibility)
- **And** PermissionDecision: allow 는 정상 적용

### Edge case — HookSpecificOutput 자체 부재

- **Given** `{"continue":true}` (HookSpecificOutput 없음)
- **When** Dispatch 검증
- **Then** mismatch 검사 skip; top-level fields 로 처리

### Test mapping

- `internal/hook/registry_test.go::TestDispatch_HookSpecificOutputMismatch` (new, M1) — sentinel `HookSpecificOutputMismatch`

---

## AC-V3R2-RT-001-06 — strict_mode true 시 legacy wrapper 즉시 reject

Maps to: REQ-V3R2-RT-001-021.

### Happy path

- **Given** `.moai/config/sections/system.yaml` 에 `hook.strict_mode: true` 설정
- **And** legacy wrapper 가 stdout 빈 + exit code 0
- **When** ParseHookOutput 처리
- **Then** `ErrHookProtocolLegacyRejected` 반환
- **And** turn halts with user-visible SystemMessage `"hook protocol: legacy exit code rejected in strict mode"`
- **And** banner 는 emit 되지 않음 (rejection 이 우선)

### Edge case — strict_mode + plugin source

- **Given** strict_mode=true + `HookInput.ConfigurationSource == "plugin"`
- **When** ParseHookOutput 처리
- **Then** dual-parse fallback 적용 (NO error) — REQ-051 plugin bypass

### Edge case — strict_mode + JSON 출력 (정상 경로)

- **Given** strict_mode=true + valid JSON stdout
- **When** ParseHookOutput 처리
- **Then** JSON parse 정상 + 응답 생성. strict_mode 는 fallback 만 차단.

### Test mapping

- `internal/hook/strict_mode_test.go::TestStrictMode_ErrorReturnedOnLegacyOutput` (new, M1)
- `internal/hook/strict_mode_test.go::TestStrictMode_PluginSourceBypass` (new, M1)

---

## AC-V3R2-RT-001-07 — PostToolUse AdditionalContext 의 @MX 마커 라우팅

Maps to: REQ-V3R2-RT-001-016.

### Happy path

- **Given** PostToolUse hook 이 다음 JSON 출력:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"@MX:WARN at line 42 — unbounded goroutine"}}
  ```
- **When** Dispatch 처리 후 PostToolUse handler 가 응답 검사
- **Then** AdditionalContext 가 `internal/hook/mx/ingest.go` 의 ingestion 함수에 forward
- **And** SPEC-V3R2-SPC-002 가 정의한 라우팅 path 진입

### Edge case — 마커 5종 모두 인식

- **Given** AdditionalContext 가 `@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`, `@MX:TODO`, `@MX:LEGACY` 중 1개라도 포함
- **When** PostToolUse handler 검사
- **Then** ingestion path 호출 (어느 마커든 fan-in 처리)

### Edge case — 마커 부재

- **Given** AdditionalContext 가 일반 텍스트 (마커 없음)
- **When** handler 검사
- **Then** ingestion path 호출되지 않음. 일반 context 처럼 model turn 에 append.

### Test mapping

- `internal/hook/post_tool_mx_test.go::TestPostTool_RoutesMXMarkers` (existing 확장, M5c)

---

## AC-V3R2-RT-001-08 — SubagentStop Continue:false → teammate idle blocker

Maps to: REQ-V3R2-RT-001-014.

### Happy path

- **Given** SubagentStop hook 이 다음 JSON 출력:
  ```json
  {"continue":false,"systemMessage":"coverage below 85%"}
  ```
- **When** Dispatch 처리
- **Then** teammate 가 idle 상태 진입 차단
- **And** orchestrator 가 BlockerReport 생성 (subagent 는 AskUserQuestion 직접 호출 금지)
- **And** orchestrator 가 AskUserQuestion 으로 user 에게 surfacing

### Edge case — Continue 미설정 (default true)

- **Given** `{"systemMessage":"warning"}` (Continue 필드 없음)
- **When** Dispatch
- **Then** Continue 의 default 는 true; teammate idle 정상 진행

### Edge case — Continue:false 인 SessionEnd

- **Given** SessionEnd 이벤트 + `{"continue":false}`
- **When** Dispatch
- **Then** session 종료 차단; SystemMessage user 에게 노출 (`registry.go` Dispatch 의 generic Continue 처리)

### Test mapping

- `internal/hook/registry_test.go::TestDispatch_ContinueFalseBlocksTeammateIdle` (new, M1)
- `internal/hook/subagent_stop_test.go::TestSubagentStop_ContinueFalse` (existing 확장)

---

## AC-V3R2-RT-001-09 — 128 KiB AdditionalContext 가 64 KiB 로 절단

Maps to: REQ-V3R2-RT-001-022.

### Happy path

- **Given** hook 이 128 KiB 길이의 AdditionalContext 를 emit:
  ```json
  {"additionalContext":"<131072 bytes of text>"}
  ```
- **When** ValidateHookResponse 호출
- **Then** `HookResponse.AdditionalContext` 가 64 KiB (65536 bytes) 로 절단
- **And** `HookResponse.SystemMessage` 에 truncation notice 포함:
  ```
  [Notice: additionalContext truncated from 131072 to 65536 bytes]
  ```

### Edge case — 정확히 64 KiB

- **Given** AdditionalContext 가 정확히 65536 bytes
- **When** ValidateHookResponse
- **Then** 절단 없음; SystemMessage 에 notice 추가되지 않음 (`> maxContextSize` 가 false)

### Edge case — 64 KiB + 1 byte

- **Given** AdditionalContext 가 65537 bytes (1 byte 초과)
- **When** ValidateHookResponse
- **Then** 65536 으로 절단 + notice 추가

### Test mapping

- `internal/hook/dual_parse_test.go::TestValidateHookResponse_TruncatesAt64KiB` (new, M1) — 128 KiB input
- `internal/hook/dual_parse_test.go::TestValidateHookResponse_ExactlyAtBoundary` (new, M1) — 65536 byte input

---

## AC-V3R2-RT-001-10 — PreToolUse 가 deny + UpdatedInput 둘 다 반환 시 ordering

Maps to: REQ-V3R2-RT-001-041.

### Happy path

- **Given** PreToolUse hook 이 다음 JSON 출력:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","updatedInput":{"file_path":"/etc/passwd"}}}
  ```
- **When** registry 처리
- **Then** UpdatedInput 이 pending tool input 에 먼저 merge 됨 → input = `{"file_path":"/etc/passwd"}`
- **And** PermissionDecision: deny 가 평가되어 tool call 차단
- **And** denial message 가 post-update input (`/etc/passwd`) 을 reference

### Edge case — UpdatedInput 만 (PermissionDecision 없음)

- **Given** `{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{"file_path":"/tmp"}}}`
- **When** Dispatch
- **Then** input 교체 + tool 진행 (PermissionDecision: "" → "no opinion" → 다른 hook 또는 default allow)

### Edge case — PermissionDecision 만 (UpdatedInput 없음)

- **Given** `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}`
- **When** Dispatch
- **Then** input 교체 없음 + allow 결정

### Test mapping

- `internal/hook/registry_test.go::TestDispatch_PreToolUseUpdatedInputThenPermissionDecision` (new, M1)
- `internal/hook/registry_test.go::TestDispatch_PreToolUseUpdatedInputOnly` (new, M1)

---

## AC-V3R2-RT-001-11 — api_version 2 wrapper 가 exit-code fallback skip

Maps to: REQ-V3R2-RT-001-030.

### Happy path

- **Given** wrapper file `.claude/hooks/moai/handle-future.sh` 가 첫 줄에 `# moai-hook-api-version: 2` 헤더 포함
- **And** wrapper 가 stdout 빈 + exit code 0 으로 종료
- **When** ParseHookOutput 처리 (wrapperPath 또는 env var 인지)
- **Then** exit-code synthesis 수행하지 않음
- **And** explicit no-op `HookResponse{}` 반환 (allow + continue, but explicitly empty)

### Edge case — api_version 1 (default) wrapper 가 stdout 빈 + exit 0

- **Given** 헤더 없는 wrapper (api_version 1 으로 간주)
- **And** stdout 빈 + exit code 0
- **When** ParseHookOutput 처리
- **Then** exit-code synthesis 수행 → `HookResponse{}` (의미상 동일하지만 경로가 다름)

### Edge case — api_version 2 wrapper 가 invalid JSON

- **Given** api_version 2 헤더 + stdout malformed JSON
- **When** ParseHookOutput
- **Then** JSON parse 실패; api_version 2 의 의미는 "fallback skip" 이므로 explicit error 반환 (NOT exit-code synthesis). 단, AC-15 의 deprecation banner 는 emit 되지 않음 (legacy 가 아니므로).

### Test mapping

- `internal/hook/api_version_test.go::TestApiVersion2_SkipsExitCodeFallback` (new, M1)
- `internal/hook/api_version_test.go::TestApiVersion2_ParsesShellHeader` (new, M1)

---

## AC-V3R2-RT-001-12 — validator/v10 가 invalid PermissionDecision 거절

Maps to: REQ-V3R2-RT-001-006.

### Happy path

- **Given** `HookResponse{PermissionDecision: "yes"}` (oneof 위반)
- **When** `Validate()` 호출
- **Then** non-nil error 반환
- **And** error.Error() 에 `"PermissionDecision"` 부분 문자열 포함
- **And** error 에 `"oneof"` validation tag 명시

### Edge case — valid value (`"allow"`)

- **Given** `HookResponse{PermissionDecision: "allow"}`
- **When** `Validate()`
- **Then** nil error

### Edge case — empty string (no opinion)

- **Given** `HookResponse{PermissionDecision: ""}`
- **When** `Validate()`
- **Then** nil error (omitempty 적용; oneof 검사 skip)

### Edge case — `"defer"` (v2.1.89+ headless)

- **Given** `HookResponse{PermissionDecision: "defer"}`
- **When** `Validate()`
- **Then** nil error (defer 는 enum 에 포함; `response.go:46-61` PermissionDecisionDefer)

### Test mapping

- `internal/hook/response_test.go::TestHookResponse_ValidatorRejectsBadPermissionDecision` (new, M1)
- `internal/hook/response_test.go::TestHookResponse_ValidatorAcceptsAllowAskDenyDefer` (new, M1)

---

## AC-V3R2-RT-001-13 — 27 variant types round-trip serialization

Maps to: REQ-V3R2-RT-001-003, REQ-V3R2-RT-001-004.

### Happy path

- **Given** 27개 variant 타입 (`PreToolUseOutput`, `PostToolUseOutput`, ..., `StopFailureOutput`) 각 1개 instance
- **When** `json.Marshal` → `json.Unmarshal` 왕복
- **Then** 모든 27 instance 가 marshal → unmarshal 후 동일한 값 (identity)
- **And** 각 variant 의 `HookEventName()` 메서드가 정확한 이벤트 이름 반환

### Edge case — empty 변형 타입 (예: `PreCompactOutput{EventName: ""}`)

- **Given** `PreCompactOutput{}`
- **When** marshal → unmarshal
- **Then** unmarshal 결과는 `PreCompactOutput{}` (zero value preserved)
- **And** `HookEventName()` 는 여전히 `"PreCompact"` (메서드는 인스턴스 상태 무관)

### Edge case — variant 가 추가 필드 보유 (예: `PreToolUseOutput.UpdatedInput`)

- **Given** `PreToolUseOutput{UpdatedInput: json.RawMessage(`{"key":"value"}`)}`
- **When** marshal → unmarshal
- **Then** UpdatedInput 의 RawMessage 가 byte-level 보존

### Test mapping

- `internal/hook/wire_format_freeze_test.go::TestWireFormatFreeze_AllVariants` (existing baseline 확장)
- `internal/hook/response_test.go::TestHookResponse_27VariantHookEventNameRoundTrip` (new, M1)

---

## AC-V3R2-RT-001-14 — Plugin-source bypass in strict mode

Maps to: REQ-V3R2-RT-001-051.

### Happy path

- **Given** `hook.strict_mode: true` 설정
- **And** plugin-contributed hook (HookInput.ConfigurationSource == "plugin")
- **And** legacy exit-code-only output
- **When** ParseHookOutput 처리
- **Then** `ErrHookProtocolLegacyRejected` 반환되지 않음
- **And** dual-parse fallback 정상 적용
- **And** `.moai/logs/hook.log` 에 plugin source 기록:
  ```
  HOOK_LEGACY_BYPASS source=plugin event=<event> session=<id>
  ```

### Edge case — non-plugin source + strict_mode

- **Given** strict_mode=true + ConfigurationSource="user_settings" + legacy output
- **When** ParseHookOutput
- **Then** `ErrHookProtocolLegacyRejected` 반환 (정상 strict 동작)

### Edge case — plugin source + JSON output

- **Given** ConfigurationSource="plugin" + valid JSON output
- **When** ParseHookOutput
- **Then** JSON parse 정상; bypass 분기 진입하지 않음 (fallback 이 발생하지 않으므로)

### Test mapping

- `internal/hook/strict_mode_test.go::TestStrictMode_PluginSourceBypass` (new, M1)
- `internal/hook/strict_mode_test.go::TestIsPluginSource_NonPluginSources` (new, M1)

---

## AC-V3R2-RT-001-15 — Deprecation banner once-per-session

Maps to: REQ-V3R2-RT-001-007, REQ-V3R2-RT-001-050.

### Happy path

- **Given** 새 세션 (sessionID = "session-A")
- **And** legacy hook 이 첫 호출
- **When** ParseHookOutput 처리 + banner 검사
- **Then** banner emit (1회) — SystemMessage 에 deprecation 텍스트 포함
- **When** 동일 세션에서 두 번째 legacy hook 호출
- **Then** banner emit 되지 않음 (already emitted in session)
- **And** exit code 의미는 정상 처리 (allow/deny 합성)

### Edge case — 다른 세션은 독립

- **Given** "session-A" 에서 banner 이미 emit
- **And** 새 세션 "session-B" 시작 + legacy hook
- **When** banner 검사
- **Then** "session-B" 에서 banner 첫 emit (per-session 격리)

### Edge case — MOAI_HOOK_LEGACY=1 + 새 세션

- **Given** "session-C" + MOAI_HOOK_LEGACY=1 + legacy hook
- **When** banner 검사
- **Then** banner 미발사 (env opt-out 우선)

### Edge case — 동시성 (2 goroutine)

- **Given** "session-D" 에서 동시 2개 goroutine 이 ParseHookOutput 호출
- **When** banner LoadOrStore atomic check
- **Then** 정확히 1개 goroutine 만 banner emit (sync.Map atomic)

### Test mapping

- `internal/hook/banner_test.go::TestBanner_OncePerSession` (new, M1)
- `internal/hook/banner_test.go::TestBanner_PerSessionIsolation` (new, M1)
- `internal/hook/banner_test.go::TestBanner_ConcurrentSafety` (new, M1)

---

## Summary table — AC → REQ → Test

| AC | REQs covered | Test files |
|----|--------------|------------|
| AC-01 | REQ-010, REQ-013 | `dual_parse_test.go::TestParseHookOutput_PreToolUseJSON`, `registry_test.go::TestDispatch_PreToolUseUpdatedInput` |
| AC-02 | REQ-001, REQ-012, REQ-032 | `dual_parse_test.go::TestParseHookOutput_SessionStartContextAndWatchPaths`, `registry_test.go::TestDispatch_AdditionalContextRoutedToModelContext` |
| AC-03 | REQ-011 | `dual_parse_test.go::TestSynthesizeFromExitCode_*` (3 cases) |
| AC-04 | REQ-020, REQ-007 | `banner_test.go::TestBanner_SuppressedByMoaiHookLegacy`, `strict_mode_test.go::TestIsLegacyEnvSet_LooseTruthy` |
| AC-05 | REQ-040 | `registry_test.go::TestDispatch_HookSpecificOutputMismatch` |
| AC-06 | REQ-021 | `strict_mode_test.go::TestStrictMode_ErrorReturnedOnLegacyOutput`, `TestStrictMode_PluginSourceBypass` |
| AC-07 | REQ-016 | `post_tool_mx_test.go::TestPostTool_RoutesMXMarkers` |
| AC-08 | REQ-014 | `registry_test.go::TestDispatch_ContinueFalseBlocksTeammateIdle`, `subagent_stop_test.go::TestSubagentStop_ContinueFalse` |
| AC-09 | REQ-022 | `dual_parse_test.go::TestValidateHookResponse_TruncatesAt64KiB`, `TestValidateHookResponse_ExactlyAtBoundary` |
| AC-10 | REQ-041 | `registry_test.go::TestDispatch_PreToolUseUpdatedInputThenPermissionDecision` |
| AC-11 | REQ-030 | `api_version_test.go::TestApiVersion2_SkipsExitCodeFallback`, `TestApiVersion2_ParsesShellHeader` |
| AC-12 | REQ-006, REQ-002 | `response_test.go::TestHookResponse_ValidatorRejectsBadPermissionDecision`, `TestHookResponse_ValidatorAcceptsAllowAskDenyDefer` |
| AC-13 | REQ-003, REQ-004 | `wire_format_freeze_test.go::TestWireFormatFreeze_AllVariants`, `response_test.go::TestHookResponse_27VariantHookEventNameRoundTrip` |
| AC-14 | REQ-051 | `strict_mode_test.go::TestStrictMode_PluginSourceBypass`, `TestIsPluginSource_NonPluginSources` |
| AC-15 | REQ-007, REQ-050 | `banner_test.go::TestBanner_OncePerSession`, `TestBanner_PerSessionIsolation`, `TestBanner_ConcurrentSafety` |

Total new test functions: **~28 across 5 new test files + 3 existing files extended**.

---

## Definition of Done

본 SPEC 는 다음 모든 조건이 참일 때 완료:

1. 위 15개 AC 가 `go test ./internal/hook/` 으로 모두 PASS.
2. 전체 `go test ./...` 이 worktree root 에서 zero failures + zero cascading regressions.
3. `make build` 성공 + `internal/template/embedded.go` 정상 재생성.
4. `go vet ./...` + `golangci-lint run` 경고 0.
5. `progress.md` 가 `run_complete_at: <timestamp>` + `run_status: implementation-complete` 로 업데이트.
6. CHANGELOG entry 가 `## [Unreleased] / ### Changed (BREAKING — BC-V3R2-001)` 아래 존재.
7. 7개 MX tag 가 `plan.md` §6 verbatim 으로 삽입 (3 ANCHOR + 2 NOTE + 2 WARN).
8. `manager-git` 이 연 PR 의 모든 required CI 통과 (Lint, Test ubuntu/macos/windows, Build 5 platforms, CodeQL).

---

End of acceptance.md.
