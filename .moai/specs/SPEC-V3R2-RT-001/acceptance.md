# SPEC-V3R2-RT-001 Acceptance Criteria — EARS Format

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.1.1, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.1   | 2026-05-14 | plan-auditor defect fix       | REQ sequential renumber (001-025), EARS AC conversion, frontmatter sync |
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 15 ACs (AC-V3R2-RT-001-01 through -15)     |

---

## Scope

`spec.md` §6 의 15개 AC 를 EARS 형식으로 변환. 각 AC 는 happy-path + 1-3개 edge case + Go 테스트 함수 매핑을 동반.

표기:
- **Test mapping**: AC 를 검증하는 Go 테스트 함수 (또는 수동 검증 단계).
- **Sentinel**: 부정 경로에서 테스트가 기대하는 정확한 에러 문자열.

---

## AC-V3R2-RT-001-01 — PreToolUse JSON 응답이 UpdatedInput + PermissionDecision 을 적용

Maps to: REQ-V3R2-RT-001-008, REQ-V3R2-RT-001-011.

### Happy path

- **WHEN** PreToolUse hook wrapper 가 stdout 으로 다음 JSON 을 출력한다:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"file_path":"/tmp/x"}}}
  ```
- **THE system SHALL**: ParseHookOutput 이 JSON parse 성공 → exit-code 무시
- **AND** `HookResponse.HookSpecificOutput` 에 nested PreToolUseOutput 정보 보존
- **AND** pending tool input 이 `{"file_path":"/tmp/x"}` 로 교체된 후 tool dispatch
- **AND** SPEC-V3R2-RT-002 의 permission stack resolver 가 `permissionDecision: allow` 수신

### Edge case — JSON parse 성공 시 exit code 무시

- **WHEN** 동일 JSON + wrapper exit code 7 (의미 있는 값) 이 ParseHookOutput 에 처리될 때
- **THE system SHALL**: exit code 7 을 완전히 무시; JSON 의 의미가 우선

### Test mapping

- `internal/hook/dual_parse_test.go::TestParseHookOutput_PreToolUseJSON` (existing baseline 확장)
- `internal/hook/registry_test.go::TestDispatch_PreToolUseUpdatedInput` (new, M1)

---

## AC-V3R2-RT-001-02 — SessionStart 이 AdditionalContext + WatchPaths 응답

Maps to: REQ-V3R2-RT-001-001, REQ-V3R2-RT-001-010, REQ-V3R2-RT-001-020.

### Happy path

- **WHEN** SessionStart hook 이 다음 JSON 을 출력:
  ```json
  {"additionalContext":"ctx","watchPaths":["/abs/.env"]}
  ```
- **THE system SHALL**: `HookResponse.AdditionalContext == "ctx"` 및 `HookResponse.WatchPaths == ["/abs/.env"]` 로 Populate
- **AND** `Validate()` 통과 (validator/v10 schema 위반 없음)
- **AND** AdditionalContext 가 다음 model turn 의 system-context block 에 append
- **AND** WatchPaths 가 `WatchPathsRegistrar.Register([]string)` 으로 forward

### Edge case — WatchPathsRegistrar 미등록 시 graceful

- **WHEN** 동일 JSON 출력 + registry 에 `WatchPathsRegistrar` 가 nil 인 상태에서 Dispatch 처리될 때
- **THE system SHALL**: WatchPaths 는 logged but not registered (no panic)
- **AND** AdditionalContext routing 은 정상 동작 (다른 필드는 영향 없음)

### Edge case — 빈 WatchPaths

- **WHEN** `{"additionalContext":"ctx","watchPaths":[]}` 출력이 Dispatch 처리될 때
- **THE system SHALL**: `WatchPathsRegistrar.Register` 를 호출하지 않음 (zero-length slice skip)

### Test mapping

- `internal/hook/dual_parse_test.go::TestParseHookOutput_SessionStartContextAndWatchPaths` (new, M1)
- `internal/hook/registry_test.go::TestDispatch_AdditionalContextRoutedToModelContext` (new, M1)

---

## AC-V3R2-RT-001-03 — Legacy hook exit code 2 → deny + stderr reason

Maps to: REQ-V3R2-RT-001-009.

### Happy path

- **WHEN** legacy hook wrapper 가 stdout 빈 출력 + exit code 2 + stderr `"blocked"` 로 종료할 때
- **THE system SHALL**: dual-parse fallback 진입 (JSON empty)
- **AND** `synthesizeFromExitCode(2, "blocked")` 호출
- **AND** `HookResponse.PermissionDecision == "deny"`
- **AND** `HookResponse.SystemMessage == "blocked"` (현재 `dual_parse.go:75-78` 동작)

### Edge case — exit code 1 (기타 non-zero)

- **WHEN** stdout 빈 출력 + exit code 1 + stderr `"transient error"` 로 종료할 때
- **THE system SHALL**: `HookResponse.SystemMessage == "transient error"` (PermissionDecision 미설정 — "no opinion")

### Edge case — exit code 0 + stderr (warning)

- **WHEN** stdout 빈 출력 + exit code 0 + stderr `"warning"` 으로 종료할 때
- **THE system SHALL**: `HookResponse{}` (empty 응답, allow + continue)
- **AND** stderr 는 무시되거나 별도 로그 (현재 동작 유지)

### Test mapping

- `internal/hook/dual_parse_test.go::TestSynthesizeFromExitCode_Code2WithStderr` (existing)
- `internal/hook/dual_parse_test.go::TestSynthesizeFromExitCode_Code1WithStderr` (existing)

---

## AC-V3R2-RT-001-04 — MOAI_HOOK_LEGACY=1 이 deprecation banner 억제

Maps to: REQ-V3R2-RT-001-015, REQ-V3R2-RT-001-007.

### Happy path

- **WHILE** 환경변수 `MOAI_HOOK_LEGACY=1` 설정 + legacy wrapper 가 exit-code-only 출력 시
- **THE system SHALL**: 본 세션에서 deprecation banner 가 emit 되지 않음
- **AND** dual-parse fallback 은 정상 동작 (exit code 의미 보존)
- **AND** subsequent legacy hook 도 banner 미발사

### Edge case — MOAI_HOOK_LEGACY=0 또는 미설정

- **WHEN** 환경변수 `MOAI_HOOK_LEGACY` 미설정 + 동일 legacy wrapper 가 ParseHookOutput 첫 호출될 때
- **THE system SHALL**: banner 가 1회 emit (SystemMessage 또는 별도 채널)
- **AND** subsequent legacy hook 은 banner 미반복 (REQ-024 once-per-session)

### Edge case — MOAI_HOOK_LEGACY=true 또는 yes (loose truthy)

- **WHEN** 환경변수 `MOAI_HOOK_LEGACY=true` 인 상태에서 banner 검사할 때
- **THE system SHALL**: banner suppressed (loose truthy 인정 — `IsLegacyEnvSet` 가 `"1"`, `"true"`, `"yes"` 모두 true 반환)

### Test mapping

- `internal/hook/banner_test.go::TestBanner_SuppressedByMoaiHookLegacy` (new, M1)
- `internal/hook/strict_mode_test.go::TestIsLegacyEnvSet_LooseTruthy` (new, M1)

---

## AC-V3R2-RT-001-05 — HookSpecificOutput.HookEventName 불일치 시 mismatch 에러

Maps to: REQ-V3R2-RT-001-021.

### Happy path

- **IF** PreToolUse hook 이 다음 JSON 출력:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PostToolUse","permissionDecision":"allow"}}
  ```
- **AND** registry 가 PreToolUse 이벤트 dispatch 중일 때
- **THEN THE system SHALL**: `HookSpecificOutputMismatch{Expected: "PreToolUse", Actual: "PostToolUse"}` 에러 반환
- **AND** `.moai/logs/hook.log` 에 mismatch entry append:
  ```
  HOOK_MISMATCH event=PreToolUse actual=PostToolUse session=<id> ts=<ts>
  ```
- **AND** hook 이 failed 처리 — output 무시 + default deny + reason="hookSpecificOutput mismatch"

### Edge case — HookEventName 빈 값 (legacy hook)

- **WHEN** `{"hookSpecificOutput":{"hookEventName":"","permissionDecision":"allow"}}` 이 Dispatch 검증될 때
- **THE system SHALL**: mismatch 처리 skip (빈 문자열은 "no opinion"; legacy compatibility)
- **AND** PermissionDecision: allow 는 정상 적용

### Edge case — HookSpecificOutput 자체 부재

- **WHEN** `{"continue":true}` (HookSpecificOutput 없음) 이 Dispatch 검증될 때
- **THE system SHALL**: mismatch 검사 skip; top-level fields 로 처리

### Test mapping

- `internal/hook/registry_test.go::TestDispatch_HookSpecificOutputMismatch` (new, M1) — sentinel `HookSpecificOutputMismatch`

---

## AC-V3R2-RT-001-06 — strict_mode true 시 legacy wrapper 즉시 reject

Maps to: REQ-V3R2-RT-001-016.

### Happy path

- **WHILE** `.moai/config/sections/system.yaml` 에 `hook.strict_mode: true` 설정 + legacy wrapper 가 stdout 빈 + exit code 0 인 상태에서
- **WHEN** ParseHookOutput 처리될 때
- **THE system SHALL**: `ErrHookProtocolLegacyRejected` 반환
- **AND** turn halts with user-visible SystemMessage `"hook protocol: legacy exit code rejected in strict mode"`
- **AND** banner 는 emit 되지 않음 (rejection 이 우선)

### Edge case — strict_mode + plugin source

- **WHILE** strict_mode=true + `HookInput.ConfigurationSource == "plugin"` 인 상태에서
- **WHEN** ParseHookOutput 처리될 때
- **THE system SHALL**: dual-parse fallback 적용 (NO error) — REQ-025 plugin bypass

### Edge case — strict_mode + JSON 출력 (정상 경로)

- **WHILE** strict_mode=true + valid JSON stdout 인 상태에서
- **WHEN** ParseHookOutput 처리될 때
- **THE system SHALL**: JSON parse 정상 + 응답 생성. strict_mode 는 fallback 만 차단.

### Test mapping

- `internal/hook/strict_mode_test.go::TestStrictMode_ErrorReturnedOnLegacyOutput` (new, M1)
- `internal/hook/strict_mode_test.go::TestStrictMode_PluginSourceBypass` (new, M1)

---

## AC-V3R2-RT-001-07 — PostToolUse AdditionalContext 의 @MX 마커 라우팅

Maps to: REQ-V3R2-RT-001-014.

### Happy path

- **WHEN** PostToolUse hook 이 다음 JSON 출력:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"@MX:WARN at line 42 — unbounded goroutine"}}
  ```
- **THE system SHALL**: AdditionalContext 가 `internal/hook/mx/ingest.go` 의 ingestion 함수에 forward
- **AND** SPEC-V3R2-SPC-002 가 정의한 라우팅 path 진입

### Edge case — 마커 5종 모두 인식

- **WHEN** AdditionalContext 가 `@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`, `@MX:TODO`, `@MX:LEGACY` 중 1개라도 포함될 때
- **THE system SHALL**: ingestion path 호출 (어느 마커든 fan-in 처리)

### Edge case — 마커 부재

- **WHEN** AdditionalContext 가 일반 텍스트 (마커 없음) 일 때
- **THE system SHALL**: ingestion path 호출되지 않음. 일반 context 처럼 model turn 에 append.

### Test mapping

- `internal/hook/post_tool_mx_test.go::TestPostTool_RoutesMXMarkers` (existing 확장, M5c)

---

## AC-V3R2-RT-001-08 — SubagentStop Continue:false → teammate idle blocker

Maps to: REQ-V3R2-RT-001-012.

### Happy path

- **WHEN** SubagentStop hook 이 다음 JSON 출력:
  ```json
  {"continue":false,"systemMessage":"coverage below 85%"}
  ```
- **THE system SHALL**: teammate 가 idle 상태 진입 차단
- **AND** orchestrator 가 BlockerReport 생성 (subagent 는 AskUserQuestion 직접 호출 금지)
- **AND** orchestrator 가 AskUserQuestion 으로 user 에게 surfacing

### Edge case — Continue 미설정 (default true)

- **WHEN** `{"systemMessage":"warning"}` (Continue 필드 없음) 이 Dispatch 처리될 때
- **THE system SHALL**: Continue 의 default 는 true; teammate idle 정상 진행

### Edge case — Continue:false 인 SessionEnd

- **WHEN** SessionEnd 이벤트 + `{"continue":false}` 이 Dispatch 처리될 때
- **THE system SHALL**: session 종료 차단; SystemMessage user 에게 노출 (`registry.go` Dispatch 의 generic Continue 처리)

### Test mapping

- `internal/hook/registry_test.go::TestDispatch_ContinueFalseBlocksTeammateIdle` (new, M1)
- `internal/hook/subagent_stop_test.go::TestSubagentStop_ContinueFalse` (existing 확장)

---

## AC-V3R2-RT-001-09 — 128 KiB AdditionalContext 가 64 KiB 로 절단

Maps to: REQ-V3R2-RT-001-017.

### Happy path

- **WHILE** hook 이 128 KiB 길이의 AdditionalContext 를 emit:
  ```json
  {"additionalContext":"<131072 bytes of text>"}
  ```
- **WHEN** ValidateHookResponse 호출될 때
- **THE system SHALL**: `HookResponse.AdditionalContext` 가 64 KiB (65536 bytes) 로 절단
- **AND** `HookResponse.SystemMessage` 에 truncation notice 포함:
  ```
  [Notice: additionalContext truncated from 131072 to 65536 bytes]
  ```

### Edge case — 정확히 64 KiB

- **WHEN** AdditionalContext 가 정확히 65536 bytes 일 때
- **THE system SHALL**: 절단 없음; SystemMessage 에 notice 추가되지 않음 (`> maxContextSize` 가 false)

### Edge case — 64 KiB + 1 byte

- **WHEN** AdditionalContext 가 65537 bytes (1 byte 초과) 일 때
- **THE system SHALL**: 65536 으로 절단 + notice 추가

### Test mapping

- `internal/hook/dual_parse_test.go::TestValidateHookResponse_TruncatesAt64KiB` (new, M1) — 128 KiB input
- `internal/hook/dual_parse_test.go::TestValidateHookResponse_ExactlyAtBoundary` (new, M1) — 65536 byte input

---

## AC-V3R2-RT-001-10 — PreToolUse 가 deny + UpdatedInput 둘 다 반환 시 ordering

Maps to: REQ-V3R2-RT-001-022.

### Happy path

- **IF** PreToolUse hook 이 다음 JSON 출력:
  ```json
  {"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","updatedInput":{"file_path":"/etc/passwd"}}}
  ```
- **THEN THE system SHALL**: UpdatedInput 이 pending tool input 에 먼저 merge 됨 → input = `{"file_path":"/etc/passwd"}`
- **AND** PermissionDecision: deny 가 평가되어 tool call 차단
- **AND** denial message 가 post-update input (`/etc/passwd`) 을 reference

### Edge case — UpdatedInput 만 (PermissionDecision 없음)

- **WHEN** `{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{"file_path":"/tmp"}}}` 이 Dispatch 될 때
- **THE system SHALL**: input 교체 + tool 진행 (PermissionDecision: "" → "no opinion" → 다른 hook 또는 default allow)

### Edge case — PermissionDecision 만 (UpdatedInput 없음)

- **WHEN** `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}` 이 Dispatch 될 때
- **THE system SHALL**: input 교체 없음 + allow 결정

### Test mapping

- `internal/hook/registry_test.go::TestDispatch_PreToolUseUpdatedInputThenPermissionDecision` (new, M1)
- `internal/hook/registry_test.go::TestDispatch_PreToolUseUpdatedInputOnly` (new, M1)

---

## AC-V3R2-RT-001-11 — api_version 2 wrapper 가 exit-code fallback skip

Maps to: REQ-V3R2-RT-001-018.

### Happy path

- **WHERE** wrapper file `.claude/hooks/moai/handle-future.sh` 가 첫 줄에 `# moai-hook-api-version: 2` 헤더 포함 + wrapper 가 stdout 빈 + exit code 0 으로 종료할 때
- **THE system SHALL**: exit-code synthesis 수행하지 않음
- **AND** explicit no-op `HookResponse{}` 반환 (allow + continue, but explicitly empty)

### Edge case — api_version 1 (default) wrapper 가 stdout 빈 + exit 0

- **WHEN** 헤더 없는 wrapper (api_version 1 으로 간주) 가 stdout 빈 + exit code 0 으로 종료할 때
- **THE system SHALL**: exit-code synthesis 수행 → `HookResponse{}` (의미상 동일하지만 경로가 다름)

### Edge case — api_version 2 wrapper 가 invalid JSON

- **WHEN** api_version 2 헤더 + stdout malformed JSON 이 ParseHookOutput 처리될 때
- **THE system SHALL**: JSON parse 실패; api_version 2 의 의미는 "fallback skip" 이므로 explicit error 반환 (NOT exit-code synthesis). 단, AC-15 의 deprecation banner 는 emit 되지 않음 (legacy 가 아니므로).

### Test mapping

- `internal/hook/api_version_test.go::TestApiVersion2_SkipsExitCodeFallback` (new, M1)
- `internal/hook/api_version_test.go::TestApiVersion2_ParsesShellHeader` (new, M1)

---

## AC-V3R2-RT-001-12 — validator/v10 가 invalid PermissionDecision 거절

Maps to: REQ-V3R2-RT-001-006.

### Happy path

- **WHEN** `HookResponse{PermissionDecision: "yes"}` (oneof 위반) 에 대해 `Validate()` 호출될 때
- **THE system SHALL**: non-nil error 반환
- **AND** error.Error() 에 `"PermissionDecision"` 부분 문자열 포함
- **AND** error 에 `"oneof"` validation tag 명시

### Edge case — valid value (`"allow"`)

- **WHEN** `HookResponse{PermissionDecision: "allow"}` 에 대해 `Validate()` 호출될 때
- **THE system SHALL**: nil error

### Edge case — empty string (no opinion)

- **WHEN** `HookResponse{PermissionDecision: ""}` 에 대해 `Validate()` 호출될 때
- **THE system SHALL**: nil error (omitempty 적용; oneof 검사 skip)

### Edge case — `"defer"` (v2.1.89+ headless)

- **WHEN** `HookResponse{PermissionDecision: "defer"}` 에 대해 `Validate()` 호출될 때
- **THE system SHALL**: nil error (defer 는 enum 에 포함; `response.go:46-61` PermissionDecisionDefer)

### Test mapping

- `internal/hook/response_test.go::TestHookResponse_ValidatorRejectsBadPermissionDecision` (new, M1)
- `internal/hook/response_test.go::TestHookResponse_ValidatorAcceptsAllowAskDenyDefer` (new, M1)

---

## AC-V3R2-RT-001-13 — 27 variant types round-trip serialization

Maps to: REQ-V3R2-RT-001-003, REQ-V3R2-RT-001-004.

### Happy path

- **WHEN** 27개 variant 타입 (`PreToolUseOutput`, `PostToolUseOutput`, ..., `StopFailureOutput`) 각 1개 instance 에 대해 `json.Marshal` → `json.Unmarshal` 왕복 수행할 때
- **THE system SHALL**: 모든 27 instance 가 marshal → unmarshal 후 동일한 값 (identity)
- **AND** 각 variant 의 `HookEventName()` 메서드가 정확한 이벤트 이름 반환

### Edge case — empty 변형 타입 (예: `PreCompactOutput{EventName: ""}`)

- **WHEN** `PreCompactOutput{}` 에 대해 marshal → unmarshal 수행할 때
- **THE system SHALL**: unmarshal 결과는 `PreCompactOutput{}` (zero value preserved)
- **AND** `HookEventName()` 는 여전히 `"PreCompact"` (메서드는 인스턴스 상태 무관)

### Edge case — variant 가 추가 필드 보유 (예: `PreToolUseOutput.UpdatedInput`)

- **WHEN** `PreToolUseOutput{UpdatedInput: json.RawMessage(`{"key":"value"}`)}` 에 대해 marshal → unmarshal 수행할 때
- **THE system SHALL**: UpdatedInput 의 RawMessage 가 byte-level 보존

### Test mapping

- `internal/hook/wire_format_freeze_test.go::TestWireFormatFreeze_AllVariants` (existing baseline 확장)
- `internal/hook/response_test.go::TestHookResponse_27VariantHookEventNameRoundTrip` (new, M1)

---

## AC-V3R2-RT-001-14 — Plugin-source bypass in strict mode

Maps to: REQ-V3R2-RT-001-025.

### Happy path

- **WHILE** `hook.strict_mode: true` 설정 + plugin-contributed hook (HookInput.ConfigurationSource == "plugin") + legacy exit-code-only output 인 상태에서
- **WHEN** ParseHookOutput 처리될 때
- **THE system SHALL**: `ErrHookProtocolLegacyRejected` 반환되지 않음
- **AND** dual-parse fallback 정상 적용
- **AND** `.moai/logs/hook.log` 에 plugin source 기록:
  ```
  HOOK_LEGACY_BYPASS source=plugin event=<event> session=<id>
  ```

### Edge case — non-plugin source + strict_mode

- **WHILE** strict_mode=true + ConfigurationSource="user_settings" + legacy output 인 상태에서
- **WHEN** ParseHookOutput 처리될 때
- **THE system SHALL**: `ErrHookProtocolLegacyRejected` 반환 (정상 strict 동작)

### Edge case — plugin source + JSON output

- **WHILE** ConfigurationSource="plugin" + valid JSON output 인 상태에서
- **WHEN** ParseHookOutput 처리될 때
- **THE system SHALL**: JSON parse 정상; bypass 분기 진입하지 않음 (fallback 이 발생하지 않으므로)

### Test mapping

- `internal/hook/strict_mode_test.go::TestStrictMode_PluginSourceBypass` (new, M1)
- `internal/hook/strict_mode_test.go::TestIsPluginSource_NonPluginSources` (new, M1)

---

## AC-V3R2-RT-001-15 — Deprecation banner once-per-session

Maps to: REQ-V3R2-RT-001-007, REQ-V3R2-RT-001-024.

### Happy path

- **WHILE** 새 세션 (sessionID = "session-A") + legacy hook 이 첫 호출된 상태에서
- **WHEN** ParseHookOutput 처리 + banner 검사될 때
- **THE system SHALL**: banner emit (1회) — SystemMessage 에 deprecation 텍스트 포함
- **WHEN** 동일 세션에서 두 번째 legacy hook 호출될 때
- **THE system SHALL**: banner emit 되지 않음 (already emitted in session)
- **AND** exit code 의미는 정상 처리 (allow/deny 합성)

### Edge case — 다른 세션은 독립

- **WHILE** "session-A" 에서 banner 이미 emit + 새 세션 "session-B" 시작 + legacy hook 인 상태에서
- **WHEN** banner 검사될 때
- **THE system SHALL**: "session-B" 에서 banner 첫 emit (per-session 격리)

### Edge case — MOAI_HOOK_LEGACY=1 + 새 세션

- **WHILE** "session-C" + MOAI_HOOK_LEGACY=1 + legacy hook 인 상태에서
- **WHEN** banner 검사될 때
- **THE system SHALL**: banner 미발사 (env opt-out 우선)

### Edge case — 동시성 (2 goroutine)

- **WHILE** "session-D" 에서 동시 2개 goroutine 이 ParseHookOutput 호출 중일 때
- **WHEN** banner LoadOrStore atomic check 수행될 때
- **THE system SHALL**: 정확히 1개 goroutine 만 banner emit (sync.Map atomic)

### Test mapping

- `internal/hook/banner_test.go::TestBanner_OncePerSession` (new, M1)
- `internal/hook/banner_test.go::TestBanner_PerSessionIsolation` (new, M1)
- `internal/hook/banner_test.go::TestBanner_ConcurrentSafety` (new, M1)

---

## Summary table — AC → REQ → Test

| AC | REQs covered | Test files |
|----|--------------|------------|
| AC-01 | REQ-008, REQ-011 | `dual_parse_test.go::TestParseHookOutput_PreToolUseJSON`, `registry_test.go::TestDispatch_PreToolUseUpdatedInput` |
| AC-02 | REQ-001, REQ-010, REQ-020 | `dual_parse_test.go::TestParseHookOutput_SessionStartContextAndWatchPaths`, `registry_test.go::TestDispatch_AdditionalContextRoutedToModelContext` |
| AC-03 | REQ-009 | `dual_parse_test.go::TestSynthesizeFromExitCode_*` (3 cases) |
| AC-04 | REQ-015, REQ-007 | `banner_test.go::TestBanner_SuppressedByMoaiHookLegacy`, `strict_mode_test.go::TestIsLegacyEnvSet_LooseTruthy` |
| AC-05 | REQ-021 | `registry_test.go::TestDispatch_HookSpecificOutputMismatch` |
| AC-06 | REQ-016 | `strict_mode_test.go::TestStrictMode_ErrorReturnedOnLegacyOutput`, `TestStrictMode_PluginSourceBypass` |
| AC-07 | REQ-014 | `post_tool_mx_test.go::TestPostTool_RoutesMXMarkers` |
| AC-08 | REQ-012 | `registry_test.go::TestDispatch_ContinueFalseBlocksTeammateIdle`, `subagent_stop_test.go::TestSubagentStop_ContinueFalse` |
| AC-09 | REQ-017 | `dual_parse_test.go::TestValidateHookResponse_TruncatesAt64KiB`, `TestValidateHookResponse_ExactlyAtBoundary` |
| AC-10 | REQ-022 | `registry_test.go::TestDispatch_PreToolUseUpdatedInputThenPermissionDecision` |
| AC-11 | REQ-018 | `api_version_test.go::TestApiVersion2_SkipsExitCodeFallback`, `TestApiVersion2_ParsesShellHeader` |
| AC-12 | REQ-006 | `response_test.go::TestHookResponse_ValidatorRejectsBadPermissionDecision`, `TestHookResponse_ValidatorAcceptsAllowAskDenyDefer` |
| AC-13 | REQ-003, REQ-004 | `wire_format_freeze_test.go::TestWireFormatFreeze_AllVariants`, `response_test.go::TestHookResponse_27VariantHookEventNameRoundTrip` |
| AC-14 | REQ-025 | `strict_mode_test.go::TestStrictMode_PluginSourceBypass`, `TestIsPluginSource_NonPluginSources` |
| AC-15 | REQ-007, REQ-024 | `banner_test.go::TestBanner_OncePerSession`, `TestBanner_PerSessionIsolation`, `TestBanner_ConcurrentSafety` |

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
