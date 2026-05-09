# SPEC-V3R2-RT-001 Deep Research (Phase 0.5)

> Research artifact for **Hook JSON-OR-ExitCode Dual Protocol**.
> Companion to `spec.md` (v0.1.0). Authored against branch `plan/SPEC-V3R2-RT-001` from `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001` (worktree mode).

## HISTORY

| Version | Date       | Author                                  | Description                                                              |
|---------|------------|-----------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5 |

---

## 1. Goal of Research

`spec.md` §1 (Goal), §2 (Scope), §3 (Environment), §4 (Assumptions), §7 (Constraints), §8 (Risks) 의 주장을 구체적인 file:line 증거와 외부 라이브러리 평가로 입증한다. Run phase 가 REQ-V3R2-RT-001-001..051 을 known-good baseline 위에서 구현할 수 있도록 한다.

본 연구는 다음 7가지 질문에 답한다:

1. **기존 hook 인프라 인벤토리**: `internal/hook/` 에 이미 존재하는 dual-parse 관련 코드는 무엇이며, 25개 REQ 를 100% 만족시키기 위한 갭은 무엇인가?
2. **Claude Code 2026.x 의 HookJSONOutput 스키마**: 정확한 wire format 은? `additionalContext`, `permissionDecision`, `updatedInput`, `systemMessage`, `continue`, `watchPaths`, `retry` 7개 필드의 의미는?
3. **`go-playground/validator/v10` 통합**: SCH-001 미머지 시 직접 추가 위험은?
4. **27개 hook 이벤트 vs 26개 셸 wrapper**: 1개 차이의 정체는? wrappers-unchanged 정책의 운영 의미는?
5. **Breaking-change 분석**: BC-V3R2-001 의 wrappers-unchanged + handlers-rewritten 호환성 shim 의 구체적 작동 방식은?
6. **Plugin-source bypass (REQ-051)**: SPEC-V3R2-RT-005 의 Source enum 미머지 시에도 본 SPEC 가 독립적으로 동작 가능한가?
7. **api_version 2 opt-in (REQ-030)**: 현재 0개 wrapper 가 api_version 2 인 상황에서 의미 있는 future-proofing 인가?

---

## 2. Inventory of `internal/hook/` 인프라 (existing)

`ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001/internal/hook/` 결과: 116 파일 (소스 + 테스트 + 서브패키지 디렉토리).

### 2.1 핵심 파일 (본 SPEC 직접 관련)

| # | File | Size | Purpose | Implements (partial) |
|---|------|------|---------|----------------------|
| 1 | `types.go` | 21,631 bytes (523 줄) | EventType (27개 const), HookInput, HookOutput, HookSpecificOutput, factory functions (NewAllowOutput, NewDenyOutput, ...), Handler/Registry/Protocol/Contract/ConfigProvider interfaces | REQ-003 (27 events 열거), REQ-002 (PermissionDecision constants), legacy bridge |
| 2 | `response.go` | 11,114 bytes (288 줄) | `HookResponse` struct, `PermissionDecision` enum (4값: allow/ask/deny/defer), `RetryHint` struct, **27개 variant types** with `HookEventName() string` 메서드 | REQ-001 (HookResponse fields — 7개 모두 존재), REQ-003/004 (27 variants — 모두 존재) |
| 3 | `dual_parse.go` | 6,297 bytes (190 줄) | `ParseHookOutput()`, `synthesizeFromExitCode()`, `ValidateHookResponse()` (64 KiB 절단 포함), `ToHookOutput()` / `ToHookResponse()` 호환성 shim, error sentinels (`ErrHookProtocolLegacyRejected` 등) | REQ-005 (dual-parse), REQ-011 (exit-code synthesis), REQ-022 (64 KiB truncation 부분), REQ-021 (error sentinel) |
| 4 | `protocol.go` | 3,313 bytes (99 줄) | `jsonProtocol` 구조체, `ReadInput()` (camelCase ↔ snake_case 정규화), `WriteOutput()` (json.Encoder), `validateInput()` | wire-format I/O |
| 5 | `registry.go` | 11,032 bytes (324 줄) | `registry` 구조체, `Dispatch()` (27 events 핸들러 디스패치), `ensureTraceWriter()`, `EnableObservability()`, AdditionalContext merge logic | REQ-014/015 (Continue/SystemMessage 부분), REQ-040 미구현 |
| 6 | `errors.go` | 718 bytes | `ErrHookInvalidInput` 등 sentinel errors | error class |
| 7 | `wire_format_freeze_test.go` | 7,777 bytes (245 줄) | wire format byte-level 안정성 테스트 (27 variant round-trip) | regression baseline for REQ-001 |
| 8 | `dual_parse_test.go` | 10,307 bytes (304 줄) | dual-parse happy path tests | regression baseline for REQ-005, REQ-010, REQ-011 |
| 9 | `response_test.go` | 6,671 bytes (240 줄) | HookResponse round-trip tests | regression baseline for REQ-001 |

### 2.2 보조 파일 (본 SPEC 통합 지점)

| File | Purpose | Touch in this SPEC? |
|------|---------|---------------------|
| `session_start.go` (19,913 bytes) | SessionStart handler | M5b: WatchPaths forward 추가 |
| `session_end.go` (21,660 bytes) | SessionEnd handler | M4: Continue:false 의미론 통합 |
| `pre_tool.go` (20,457 bytes) | PreToolUse handler | M4: UpdatedInput-then-Decision ordering 명시 |
| `post_tool.go` (18,509 bytes) | PostToolUse handler | M5c: @MX 마커 라우팅 추가 |
| `subagent_stop.go` (5,154 bytes) | SubagentStop handler | M4: Continue:false → teammate idle blocker |
| `teammate_idle.go` (6,400 bytes) | TeammateIdle handler | M4: Continue:false 처리 |
| `mx/` 디렉토리 (12 파일) | @MX 태그 ingestion | M5c integration point |
| `lifecycle/` (10 파일) | session lifecycle helpers | (no change) |

### 2.3 셸 wrapper 인벤토리

`ls .claude/hooks/moai/` 결과: **27개 핸들러 .sh 파일** 발견 — `handle-{event-kebab-case}.sh` 패턴.

전체 목록 (verified):
```
handle-agent-hook.sh
handle-compact.sh
handle-config-change.sh
handle-cwd-changed.sh
handle-elicitation-result.sh
handle-elicitation.sh
handle-file-changed.sh
handle-harness-observe.sh
handle-instructions-loaded.sh
handle-notification.sh
handle-permission-request.sh
handle-post-compact.sh
handle-post-tool-failure.sh
handle-post-tool.sh
handle-pre-tool.sh
handle-session-end.sh
handle-session-start.sh
handle-spec-status.sh    # 154 bytes — minimal (delegates to internal status)
handle-stop-failure.sh
handle-stop.sh
handle-subagent-start.sh
handle-subagent-stop.sh
handle-task-completed.sh
handle-task-created.sh
handle-teammate-idle.sh
handle-user-prompt-submit.sh
handle-worktree-create.sh
```

`spec.md` §3 의 "26개 셸 wrappers" 표현은 spec-status (154 bytes minimal delegate) 를 통상 hook event 가 아닌 internal helper 로 분류한 것으로 추정. 본 SPEC 입장에서는 27개 모두 wrappers-unchanged 적용 — 어느 것도 수정하지 않는다.

### 2.4 Delta Analysis (skeleton → spec compliance)

| Skeleton state | Spec requirement | Gap |
|----------------|------------------|-----|
| `HookResponse` struct in `response.go:11-44` 가 7개 필드 모두 정의 (`AdditionalContext`, `PermissionDecision`, `UpdatedInput`, `SystemMessage`, `Continue`, `WatchPaths`, `Retry`) | REQ-001 (7개 필드 byte-for-byte) | ✅ 충족. `omitempty` 태그도 적용됨. |
| `PermissionDecision` enum in `response.go:46-61` 4개 값 (allow/ask/deny/defer) | REQ-002 (allow/ask/deny + zero "no opinion") | ✅ 충족 + `defer` 추가 (v2.1.89+ 기능). zero-value `""` 도 정확히 "no opinion" 의미. |
| 27개 variant types in `response.go:76-287` with `HookEventName() string` | REQ-003, REQ-004 (27 variants + HookEventName 메서드) | ✅ 충족. `wire_format_freeze_test.go` 가 round-trip 검증 중. |
| `ParseHookOutput()` in `dual_parse.go:47-62` JSON-first then exit-code | REQ-005 (dual-parse), REQ-010 (JSON success), REQ-011 (exit-code fallback) | ✅ 충족. JSON parse 성공 시 exit-code 무시 (line 53-55). 실패 시 `synthesizeFromExitCode` (line 61). |
| `synthesizeFromExitCode()` in `dual_parse.go:64-86` exit code 0/2/non-zero 분기 | REQ-011 (0→allow, 2→deny+stderr, non-zero→user msg) | ⚠️ 부분 충족. 0 → allow (`HookResponse{}`), 2 → deny + SystemMessage, non-zero → SystemMessage only. spec REQ-011 의 "deny + stderr as reason" 와 "non-zero → user-visible systemMessage" 와 일치. |
| `ValidateHookResponse()` in `dual_parse.go:90-115` truncates AdditionalContext at 64 KiB | REQ-022 (64 KiB 절단 + systemMessage notice) | ✅ 충족. line 111 의 메시지 텍스트 표준화 필요 (`"AdditionalContext truncated to 64 KiB budget"` 와 정확히 일치 시키기). |
| `HookSpecificOutputMismatch` struct in `dual_parse.go:23-30` | REQ-040 (mismatch detection + log) | ⚠️ 타입은 정의됨. **Detection wiring 부재** — `registry.go` Dispatch 에서 mismatch 검사가 없음. `.moai/logs/hook.log` append 도 없음. |
| `ErrHookProtocolLegacyRejected` in `dual_parse.go:11-13` | REQ-021 (strict_mode legacy rejection) | ⚠️ 에러 타입 정의됨. **strict_mode 분기 부재** — `ParseHookOutput()` 가 strict_mode 검사를 하지 않음. config 로딩도 부재. |
| Deprecation banner | REQ-007, REQ-050 (once-per-session banner) | ❌ 미구현. banner 함수 부재. |
| MOAI_HOOK_LEGACY env | REQ-020 | ❌ 미구현. 환경변수 검사 부재. |
| `validator/v10` | REQ-006 | ❌ go.mod 에 부재. SCH-001 의존성. |
| `internal/config/types.go` `HookConfig` | REQ-021 (strict_mode config 키) | ❌ 미구현. `HookConfig` 구조체 부재. |
| `system.yaml` `hook.strict_mode` 키 | REQ-021 | ❌ 미구현. 현재 `system.yaml` 27 줄에 hook 섹션 없음. |
| AdditionalContext routing to model context | REQ-012 (SessionStart/UserPromptSubmit/PreToolUse/PostToolUse → next turn) | ⚠️ 부분. `registry.go:155-162` 가 UserPromptSubmit 의 AdditionalContext 만 merge. SessionStart/PreToolUse/PostToolUse routing 없음. |
| UpdatedInput-then-PermissionDecision ordering | REQ-041 (deterministic order) | ⚠️ 부분. `registry.go:179-193` 에 PermissionDecision/Reason wiring 있으나 UpdatedInput 우선 적용은 명시되지 않음. |
| api_version 2 opt-in | REQ-030 | ❌ 미구현. `# moai-hook-api-version: 2` 파싱 부재. |
| WatchPaths registrar interface | REQ-032 | ❌ 미구현. `WatchPathsRegistrar` 인터페이스 부재. |
| Plugin-source bypass | REQ-051 | ❌ 미구현. `HookInput.ConfigurationSource` 는 정의됨 (types.go:235-237), 하지만 strict_mode 분기에서 사용되지 않음. |

**요약**: skeleton 은 wire format (response.go) + happy-path dual-parse (dual_parse.go) 는 완성. 그러나 **strict_mode/banner/MOAI_HOOK_LEGACY/api_version/mismatch detection/AdditionalContext routing/plugin bypass** 는 모두 미구현. M1-M5 가 이 7개 영역을 채운다.

---

## 3. Claude Code 2026.x HookJSONOutput 스키마

### 3.1 정확한 wire format

Claude Code 2.1.111+ 의 hook output 은 다음 7개 top-level 필드를 가진 JSON 객체:

```json
{
  "additionalContext": "string (text appended to model turn system-context block)",
  "permissionDecision": "allow | ask | deny | defer (PreToolUse/PermissionRequest)",
  "updatedInput": { /* arbitrary JSON; PreToolUse rewrites tool input */ },
  "systemMessage": "string (user-visible status message)",
  "continue": true | false (omit = no opinion = continue),
  "watchPaths": ["/abs/path/1", "/abs/path/2"],
  "retry": { "attempts": 3, "backoff": "exponential" },
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse | PostToolUse | ...",
    /* event-specific fields */
  }
}
```

Discriminator: `hookSpecificOutput.hookEventName` 가 27개 이벤트 중 어느 것의 응답인지 식별. mismatch 시 hook 이 다른 이벤트의 PermissionDecision 을 위조할 위험이 있으므로 REQ-040 의 mismatch detection 이 security boundary.

### 3.2 본 코드베이스의 응답

`internal/hook/response.go:11-44` 의 `HookResponse` 가 위 7개 필드 모두 정확히 정의. `omitempty` 태그가 zero-value field 를 marshal 에서 제외. JSON tag 이름 (`additionalContext`, `permissionDecision`, ...) 도 Claude Code 와 byte-for-byte 일치.

`HookSpecificOutput` (types.go:269-281) 은 `HookEventName`, `PermissionDecision`, `PermissionDecisionReason`, `AdditionalContext`, `SessionTitle`, `UpdatedInput`, `UpdatedMCPToolOutput`, `UpdatedToolOutput` 8개 필드 — Claude Code 의 nested struct 와 정합.

본 SPEC 의 GREEN 작업은 이 wire format 자체에는 변경을 가하지 않는다 — 이미 byte-for-byte 정합. 변경은 dispatch routing + strict_mode/banner 분기.

### 3.3 27 이벤트 vs HookEventName 문자열

`response.go:85-287` 의 27개 variant 의 `HookEventName()` 반환값:
- `PreToolUse`, `PostToolUse`, `SessionStart`, `SessionEnd`, `Stop`, `SubagentStop`, `PreCompact`, `PostCompact`, `PostToolUseFailure`, `Notification`, `UserPromptSubmit`, `PermissionRequest`, `PermissionDenied`, `ConfigChange`, `InstructionsLoaded`, `FileChanged`, `TeammateIdle`, `TaskCompleted`, `SubagentStart`, `WorktreeCreate`, `WorktreeRemove`, `CwdChanged`, `Setup`, `Elicitation`, `ElicitationResult`, `TaskCreated`, `StopFailure`

Total = 27, matches `types.go:117-147` `ValidEventTypes()` 정확히 1:1.

---

## 4. validator/v10 통합 분석

### 4.1 의존성 상태

`grep -n "validator" /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001/go.mod` → **부재** (verified). `go.sum` 에도 없음.

SPEC-V3R2-SCH-001 (Constitution phase, 별도 SPEC) 가 `validator/v10` 을 도입할 예정이나 현재 미머지. 본 SPEC 의 M2 첫 작업이 `go get github.com/go-playground/validator/v10` 직접 추가.

### 4.2 충돌 위험 평가

- **충돌 가능성**: 낮음. Go module 시스템은 동일 의존성을 자동 dedup. SCH-001 PR 이 본 SPEC 머지 후 진입해도 `go.mod` 의 `require` 라인은 단일.
- **버전 차이**: 본 SPEC 은 `latest` (v10.X.Y) 를 호출. SCH-001 가 다른 버전을 명시한다면 Go module resolver 가 higher MVS 선택. 차이가 minor 수준이면 무해.
- **롤백 가능성**: SCH-001 가 유의미한 버전 정책을 적용하면 본 SPEC 의 `go.mod` entry 는 SCH-001 PR 에서 자연스럽게 정렬됨. 별도 cleanup 불필요.

### 4.3 Schema tag 적용

본 SPEC 의 validator 사용은 매우 제한적:

```go
type HookResponse struct {
    AdditionalContext  string             `json:"additionalContext,omitempty"`
    PermissionDecision PermissionDecision `json:"permissionDecision,omitempty" validate:"omitempty,oneof=allow ask deny defer"`
    UpdatedInput       json.RawMessage    `json:"updatedInput,omitempty"`
    SystemMessage      string             `json:"systemMessage,omitempty"`
    Continue           *bool              `json:"continue,omitempty"`
    WatchPaths         []string           `json:"watchPaths,omitempty"`
    Retry              *RetryHint         `json:"retry,omitempty" validate:"omitempty"`
    HookSpecificOutput json.RawMessage    `json:"hookSpecificOutput,omitempty"`
}

type RetryHint struct {
    Attempts int    `json:"attempts,omitempty" validate:"omitempty,gte=0,lte=10"`
    Backoff  string `json:"backoff,omitempty" validate:"omitempty,oneofci=linear exponential constant"`
}
```

`Validate() error` 메서드는 `validator.New().Struct(r)` 호출. 패키지 단위 cached `*validator.Validate` (one-time init ~50µs, struct call ~1-5µs). p99 5ms 제약 내.

### 4.4 ValidationErrors 메시지

`validator.ValidationErrors` 의 `Error()` 출력 예시:
```
Key: 'HookResponse.PermissionDecision' Error:Field validation for 'PermissionDecision' failed on the 'oneof' tag
```

AC-12 의 sentinel: "Validate() returns a non-nil error naming the offending field". `"PermissionDecision"` substring 검사로 충족.

---

## 5. 27 hook events vs 26-27 셸 wrappers

### 5.1 격차의 정체

`spec.md` §3 의 "26개 wrappers" vs §2 in-scope 의 "27 events". 실제 `ls .claude/hooks/moai/` 결과 27개 .sh 파일.

차이 해석:
- `handle-spec-status.sh` (154 bytes, minimal delegate) 는 일반 Claude Code event 가 아닌 internal helper 일 가능성 높음. `wc -l` 이 154 bytes ≈ 6-10 줄 — 다른 wrapper (~750-1600 bytes) 보다 훨씬 작음.
- spec.md 가 작성된 시점에는 26개였고 이후 spec-status 가 추가되었거나 카운트 오차일 수 있음.
- 본 SPEC 의 입장에서는 **27개 모두 wrappers-unchanged 정책 적용** — 어느 것도 수정하지 않는다.

### 5.2 wrapper 의 통상 형태

`handle-pre-tool.sh` (989 bytes, 가장 큰 단일 wrapper) 의 일반 패턴:
```bash
#!/bin/bash
INPUT=$(cat)
echo "$INPUT" | "$HOMEBREW_PREFIX/bin/moai" hook pre-tool
```

stdin 을 그대로 forward + stdout 을 그대로 forward + exit code 보존. 본 SPEC 가 dual-parse 를 도입해도 wrapper 자체는 변경 없음 — Go 핸들러가 JSON 을 emit 하면 `moai hook pre-tool` 의 stdout 이 JSON 이 되고, stdin/stdout pipeline 이 transparent forwarding.

### 5.3 BC-V3R2-001 호환성 shim 구성

```
wrapper (unchanged) ──── stdin ─────> moai binary (changed: HookResponse JSON)
                  └──── stdout ──── ParseHookOutput (changed: dual-parse)
                                          │
                                          ▼
                                   Claude Code runtime
```

핵심:
- wrapper 의 행동 = stdin pipe + binary 호출 + stdout pipe + exit code propagate. 변경 없음.
- moai binary 의 hook command 가 이전에는 exit code (deny=2) 를 사용했고 이제는 stdout 으로 JSON 을 emit. 이것이 "handlers-rewritten".
- ParseHookOutput 은 양쪽 모두 수용 (dual-parse). 외부 plugin author 가 작성한 wrapper (moai binary 가 아닌 임의 스크립트를 호출) 는 v3.x 동안 exit-code-only 로 작동 가능.

---

## 6. Plugin-source bypass (REQ-051) 분석

### 6.1 SPEC-V3R2-RT-005 의존성

REQ-051 은 "WHILE BC-V3R2-001 deprecation window, WHEN plugin-contributed hook emits legacy exit-code, THEN apply dual-parse fallback regardless of strict-mode" 를 명시.

Plugin 식별 방법:
- `HookInput.ConfigurationSource` 필드가 이미 `types.go:235-237` 에 정의 (`v2.1.49+`). 값: `"user_settings" | "project_settings" | "local_settings" | "policy_settings" | "skills"`.
- Claude Code 가 hook 을 dispatch 할 때 어느 settings 출처에서 등록된 hook 인지 표시.
- "plugin" 값은 spec.md 와 master §RT-005 에서 언급되지만 현재 Claude Code 2.1.111+ schema 에 명시적 enum value 로 존재할 수도 있고 RT-005 에서 새로 추가될 수도 있다.

### 6.2 RT-005 미머지 시 본 SPEC 의 독립성

본 SPEC 의 strict_mode bypass 로직은 string literal 비교로 충분:
```go
func IsPluginSource(input *HookInput) bool {
    return input.ConfigurationSource == "plugin"
}
```

RT-005 가 후속 머지 시 string literal 을 enum constant 로 대체하면 됨. 본 SPEC 은 enum 의 의미론에 의존하지 않으며, 단지 ConfigurationSource 가 "plugin" 인지만 판단.

만약 RT-005 가 ConfigurationSource 의 valid enum 에 "plugin" 을 추가하지 않는 결정을 내리면, 본 SPEC 의 plugin-bypass 는 dead code 가 되지만 무해. CI 가 fail 하지 않음.

---

## 7. api_version 2 opt-in (REQ-030) 분석

### 7.1 현재 상태

`grep -rn "moai-hook-api-version" /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001/.claude/hooks/` → 0개 매치. **현재 어떤 wrapper 도 api_version 2 를 선언하지 않음.**

### 7.2 의미 있는 future-proofing 인가?

YES. 이유:
1. **v3.1+ 의 wrapper 작성자**: 새로운 wrapper (특히 외부 plugin) 가 explicit no-op 응답을 표현해야 하는 경우 (`{}` JSON) 에 exit-code synthesis 가 잘못된 의미를 주입하는 것을 방지.
2. **Wrapper migration path**: 기존 wrapper 가 api_version 2 로 업그레이드될 때, exit code 0 + stdout 빈 응답이 "explicit no-op" 으로 해석되어야 함. 현재 default 는 exit code 0 → allow 라서 의미상 유사하지만 명시성 차이.
3. **ParseHookOutput 시그니처 안정성**: 현재 시그니처는 `func(stdout, exitCode, stderr) (*HookResponse, error)`. `wrapperPath string` 인자 추가 또는 `HookInput` 에 `WrapperPath string` 필드 추가 (이미 wrapper 은 `$0` 으로 자기 경로 알 수 있음). 본 SPEC 의 M5a 는 인자 추가 (signature change) 보다는 `HookInput` 필드 확장 권장 — non-breaking.

### 7.3 구현 방식

```go
// internal/hook/api_version.go
func ReadHookApiVersion(wrapperPath string) (int, error) {
    f, err := os.Open(wrapperPath)
    if err != nil { return 0, err }
    defer f.Close()
    // Read first 30 lines, look for "# moai-hook-api-version: N"
    scanner := bufio.NewScanner(f)
    re := regexp.MustCompile(`^#\s*moai-hook-api-version:\s*(\d+)\s*$`)
    for i := 0; i < 30 && scanner.Scan(); i++ {
        if m := re.FindStringSubmatch(scanner.Text()); m != nil {
            return strconv.Atoi(m[1])
        }
    }
    return 1, nil  // default: api_version 1
}
```

테스트는 mock wrapper file (with header) + mock wrapper file (without header) 양쪽 검증.

---

## 8. Breaking-change 분석

### 8.1 BC-V3R2-001 의 약속

`spec.md:39` 에서 명시: "wrappers-unchanged + handlers-rewritten 호환성 shim 으로 26개 셸 래퍼는 alpha.2 → rc.2 deprecation window 동안 작동. exit-code-only 제거는 v4.0 으로 이연."

### 8.2 사용자 영향 매트릭스

| 사용자 카테고리 | v3.0 동작 | 영향 |
|----------------|----------|------|
| 일반 사용자 (moai binary 만 사용) | JSON 출력으로 자동 마이그레이션. 동일 기능. | None — handler 변경은 transparent. |
| Plugin author (외부 hook 스크립트 작성) | exit-code-only 가 deprecation window 동안 작동. session-당 1회 banner 노출. | Cosmetic (banner). 실제 기능 동일. |
| CI / air-gapped 환경 | banner 가 stdout 에 fluf 추가 가능. `MOAI_HOOK_LEGACY=1` 으로 suppress. | Suppressible. |
| Strict-mode 채택 팀 | `hook.strict_mode: true` 설정 시 exit-code-only hook 즉시 reject. early failure mode. | Opt-in. Default false. |

### 8.3 v4.0 에서 제거될 surfaces

본 SPEC 가 도입하는 surface 중 v4.0 에서 제거 예정:
- `synthesizeFromExitCode()` (`dual_parse.go:64-86`). exit-code 의미론 자체가 사라짐.
- `MOAI_HOOK_LEGACY=1` 환경변수 분기. 더이상 legacy mode 가 없음.
- Deprecation banner. legacy 자체가 없으므로 banner 불필요.
- `IsLegacyEnvSet()`. 동일 이유.

본 SPEC 가 도입하는 surface 중 v4.0 이후에도 남는 것:
- `HookResponse` struct + 27 variants — wire format 자체는 영구.
- `Validate()` 메서드 — schema 검증.
- strict_mode (이름은 default 가 될 가능성).
- HookSpecificOutputMismatch detection — security boundary.
- AdditionalContext routing.
- WatchPaths registrar interface.

### 8.4 v4.0 마이그레이션 가이드 (proactive 작성)

본 SPEC 의 CHANGELOG entry 가 v4.0 sunset 명시. v4.0 SPEC 가 작성될 때 본 SPEC 의 §8 가 baseline.

---

## 9. 외부 라이브러리 평가 요약

| Library / Source | Purpose | Decision |
|------------------|---------|----------|
| `github.com/go-playground/validator/v10` | Schema 검증 (`oneof`, `gte/lte`) | **ADOPT** (M2 에서 직접 추가; SCH-001 후속 머지 시 dedup) |
| `encoding/json` 표준 | wire format I/O + `json.RawMessage` discriminator | **ADOPT** (이미 사용 중) |
| `sync.Map` 표준 | banner once-per-session thread-safe storage | **ADOPT** (atomic LoadOrStore) |
| `regexp` 표준 | api_version header 파싱 | **ADOPT** (이미 normalize.go 등에서 사용) |
| `github.com/golang/mock` (mock 라이브러리) | hook unit tests | **REJECT** (현재 코드베이스는 stdlib `testing` + 인터페이스 mock 사용 — 일관성) |
| `github.com/cespare/xxhash` 등 hash lib | banner session_id hashing | **REJECT** (불필요; sessionID 자체가 unique key) |
| `golang.org/x/exp/slog` 구조화 로깅 | mismatch logging to `.moai/logs/hook.log` | **PROPOSED** — 현재 코드베이스가 이미 `log/slog` (Go 1.21+) 사용 중인지 확인 필요. 미사용 시 `os.OpenFile` + `fmt.Fprintf` 단순 append. |

---

## 10. Cross-SPEC dependency status

### 10.1 Blocked by

- **SPEC-V3R2-SCH-001** (validator/v10 통합): 현재 미머지. **mitigation**: M2 에서 직접 의존성 추가. 후속 SCH-001 머지 시 dedup 자동.
- **SPEC-V3R2-CON-001** (FROZEN-zone codification): Wave 6 history 기준 완료. 본 SPEC 의 protocol-is-structural 선언이 의존하지만 별도 작업 불필요.
- **SPEC-V3R2-RT-005** (settings provenance Source enum): 미머지. **mitigation**: 본 SPEC 은 string literal `"plugin"` 으로 ConfigurationSource 검사 (RT-005 enum constants 의존 없음).

### 10.2 Blocks

- **SPEC-V3R2-RT-002** (permission stack): PreToolUse 의 PermissionDecision 필드 consumer. 본 SPEC 머지 후 진입.
- **SPEC-V3R2-RT-003** (sandbox): UpdatedInput 으로 명령 인자 mid-turn mutation. 본 SPEC 머지 후 진입.
- **SPEC-V3R2-RT-006** (handler completeness): 27개 이벤트 비즈니스 로직 완성. 본 SPEC 의 wire format 위에서 진행.
- **SPEC-V3R2-SPC-002** (@MX 자율 add/update/remove): PostToolUse `additionalContext` 라우팅. 본 SPEC 의 M5c integration point 사용.
- **SPEC-V3R2-HRN-002** (evaluator fresh-memory injection): PreToolUse `additionalContext` 사용.

### 10.3 Related (non-blocking)

- **SPEC-V3R2-MIG-001** (v2→v3 migrator): hook types 에 validator/v10 import 추가. 본 SPEC 와 무관.
- **SPEC-V3R2-CON-003** (constitution consolidation): hook-protocol 텍스트를 CLAUDE.md §10 에서 hooks-system.md 로 이동. 본 SPEC 머지와 독립.

---

## 11. File:line evidence anchors

다음 anchors 는 run phase 에서 load-bearing. plan.md §3.4 에서 verbatim 인용.

1. `spec.md:43-55` — In-scope items 1-9 (typed HookResponse + 7 fields + dual-parse + validator + context-injection + input-mutation + continue:false + deprecation banner + strict_mode).
2. `spec.md:57-66` — Out-of-scope items 1-8 (RT-006/RT-002/RT-003/RT-004/RT-005/RT-007/v3.1+/SPC-002).
3. `spec.md:106-145` — 25 EARS REQs (Ubiquitous 7, Event 7, State 3, Optional 3, Unwanted 3, Complex 2).
4. `spec.md:147-163` — 15 ACs (AC-01..AC-15).
5. `spec.md:167-173` — 7 constraints.
6. `internal/hook/types.go:115-147` — `ValidEventTypes()` 27 events 열거.
7. `internal/hook/types.go:155-165` — PermissionDecision constants (allow/deny/ask/defer + block).
8. `internal/hook/types.go:167-268` — `HookInput` struct (29 필드 — 27 events 의 union of payload fields).
9. `internal/hook/types.go:269-281` — `HookSpecificOutput` struct (8 필드 + hookEventName discriminator).
10. `internal/hook/types.go:283-312` — `HookOutput` struct (legacy bridge).
11. `internal/hook/types.go:314-325` — `NewAllowOutput()` factory + `@MX:ANCHOR fan_in=19` 표시 (이미 존재).
12. `internal/hook/response.go:11-44` — `HookResponse` struct + 7 필드.
13. `internal/hook/response.go:46-61` — `PermissionDecision` enum (4 값).
14. `internal/hook/response.go:64-70` — `RetryHint` struct.
15. `internal/hook/response.go:76-287` — 27 variant types + `HookEventName()` 메서드.
16. `internal/hook/dual_parse.go:11-19` — error sentinels (3개: LegacyRejected, InvalidPermissionDecision, SpecificOutputMismatch).
17. `internal/hook/dual_parse.go:23-30` — `HookSpecificOutputMismatch` struct.
18. `internal/hook/dual_parse.go:47-62` — `ParseHookOutput()` 진입점.
19. `internal/hook/dual_parse.go:64-86` — `synthesizeFromExitCode()`.
20. `internal/hook/dual_parse.go:88-115` — `ValidateHookResponse()` (64 KiB truncation).
21. `internal/hook/dual_parse.go:117-189` — `ToHookOutput`/`ToHookResponse` 호환성 shim.
22. `internal/hook/protocol.go:26-47` — `ReadInput()` (camelCase ↔ snake_case 정규화).
23. `internal/hook/protocol.go:52-63` — `WriteOutput()` json.Encoder.
24. `internal/hook/registry.go:65-200` — `Dispatch()` 27 events 핸들러 디스패치.
25. `internal/hook/registry.go:155-162` — UserPromptSubmit AdditionalContext merge.
26. `internal/hook/registry.go:179-193` — PreToolUse PermissionDecision/Reason wiring.
27. `.claude/hooks/moai/*.sh` — 27 wrappers (verified count via `ls`).
28. `.moai/config/sections/system.yaml:1-27` — 본 SPEC 가 hook.strict_mode 키 추가 대상.
29. `.claude/rules/moai/core/agent-common-protocol.md:#user-interaction-boundary` — subagent prohibition.
30. `.claude/rules/moai/workflow/spec-workflow.md:#phase-0-5-plan-audit-gate` — plan-auditor entry condition.
31. `CLAUDE.local.md:§6` — test isolation (`t.TempDir()` + `filepath.Abs`).
32. `CLAUDE.local.md:§14` — no hardcoded paths in `internal/`.
33. master `.moai/v3-redesign/r3-cc-architecture-reread.md:§2 Decision 5` — JSON-OR-exitcode adoption candidate.
34. master `.moai/v3-redesign/r6-commands-hooks-style-rules.md:§2.2` — handler audit.
35. master `.moai/v3-redesign/r6-commands-hooks-style-rules.md:§A` — Hook Coverage Matrix.

Total: **35 distinct file:line anchors** (exceeds plan-auditor minimum of 10).

---

## 12. Open questions resolved during research

| Question | Resolution |
|----------|-----------|
| Q1: `HookSpecificOutput.HookEventName` 검증 위치는 ParseHookOutput 또는 Dispatch 어디인가? | **Dispatch** — ParseHookOutput 은 wire format unmarshal 만 책임; mismatch 는 dispatched event 와의 비교이므로 registry 가 owner. |
| Q2: deprecation banner 의 SystemMessage 경로는 stdout JSON 인가 stderr 인가? | **JSON systemMessage** — Claude Code 가 user 에게 노출. stderr 는 hook process 내부 로깅 용. |
| Q3: 64 KiB 절단 시 `AdditionalContext` 의 어떤 부분을 보존? | **앞 64 KiB** (현재 `dual_parse.go:108-109` 의 `[:maxContextSize]` 동작 유지). 마지막 부분 보존이 의미상 중요한 use case 는 보고된 바 없음. |
| Q4: api_version 2 의 wrapperPath 는 어떻게 ParseHookOutput 까지 전달? | **HookInput.WrapperPath** 신규 필드 (또는 환경변수 `MOAI_HOOK_WRAPPER_PATH`). 본 SPEC M5a 에서 결정 — 환경변수 방식이 wrapper code change 없이 동작하므로 우선. |
| Q5: WatchPaths 를 SessionStart 외 다른 이벤트가 emit 가능? | **InstructionsLoaded, FileChanged 도 가능** (`response.go:198, 207` variant 가 WatchPaths 필드 보유). 본 SPEC REQ-032 는 SessionStart 만 다루지만 future expansion 은 자연스럽게 가능. |

---

End of research.md.
