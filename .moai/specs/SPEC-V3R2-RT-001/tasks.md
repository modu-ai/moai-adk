# SPEC-V3R2-RT-001 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)      | Initial task breakdown — 31 tasks (T-RT001-01..31) across M1-M5         |

---

## Task ID Convention

- ID format: `T-RT001-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go), `expert-testing` (테스트 케이스 설계), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- DDD/TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M5 (GREEN/REFACTOR)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase) — Priority P0

Goal: 18개 새로운 실패 테스트 + 4개 신규 테스트 파일 추가. 기존 happy-path 테스트 (`dual_parse_test.go`, `response_test.go`, `wire_format_freeze_test.go`, `protocol_test.go`, `registry_test.go`, `types_test.go`) 의 GREEN baseline 보존.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT001-01 | `internal/hook/dual_parse_test.go` 확장: `TestParseHookOutput_StrictModeRejectsLegacy` 추가 (AC-06, REQ-021). strict_mode 활성 + legacy 출력 → ErrHookProtocolLegacyRejected 검증. | expert-testing | `internal/hook/dual_parse_test.go` (extend, ~30 LOC) | none | 1 file (extend) | RED — strict_mode 분기 부재 |
| T-RT001-02 | `dual_parse_test.go` 확장: `TestParseHookOutput_MoaiHookLegacySuppressesBanner` (AC-04, REQ-020). 환경변수 `MOAI_HOOK_LEGACY=1` 시 banner 미발사 검증. `t.Setenv` 사용. | expert-testing | same file (extend, ~25 LOC) | T-RT001-01 | 1 file (extend) | RED — banner 함수 부재 |
| T-RT001-03 | `dual_parse_test.go` 확장: `TestParseHookOutput_AdditionalContextTruncatedAt64KiB` (AC-09, REQ-022) + `TestValidateHookResponse_ExactlyAtBoundary` (boundary 65536). 128 KiB / 65536 / 65537 byte 입력. | expert-testing | same file (extend, ~50 LOC) | T-RT001-01 | 1 file (extend) | RED — truncation message 표준화 부족 |
| T-RT001-04 | `dual_parse_test.go` 확장: `TestParseHookOutput_PluginSourceBypassesStrict` (AC-14, REQ-051). strict_mode=true + ConfigurationSource="plugin" + legacy → bypass 검증. | expert-testing | same file (extend, ~30 LOC) | T-RT001-01 | 1 file (extend) | RED — plugin bypass 분기 부재 |
| T-RT001-05 | `dual_parse_test.go` 확장: `TestParseHookOutput_ApiVersion2SkipsExitCodeFallback` (AC-11, REQ-030). mock wrapper file + api_version 2 헤더 + 빈 stdout → explicit no-op. | expert-testing | same file (extend, ~30 LOC) | T-RT001-01 | 1 file (extend) | RED — api_version 검사 부재 |
| T-RT001-06 | `internal/hook/response_test.go` 확장: `TestHookResponse_ValidatorRejectsBadPermissionDecision` + `TestHookResponse_ValidatorAcceptsAllowAskDenyDefer` (AC-12, REQ-002, REQ-006). `r.Validate()` non-nil error / nil error 분기. | expert-testing | `internal/hook/response_test.go` (extend, ~40 LOC) | T-RT001-01 | 1 file (extend) | RED — Validate() 메서드 부재 |
| T-RT001-07 | `response_test.go` 확장: `TestHookResponse_27VariantHookEventNameRoundTrip` (AC-13, REQ-003, REQ-004). 27개 variant 의 marshal→unmarshal identity + HookEventName() 반환값 정확성. | expert-testing | same file (extend, ~80 LOC) | T-RT001-01 | 1 file (extend) | GREEN at baseline (variants 이미 존재) — regression sentinel |
| T-RT001-08 | `internal/hook/registry_test.go` 확장: `TestDispatch_HookSpecificOutputMismatch` (AC-05, REQ-040). PreToolUse dispatch 중 hookEventName="PostToolUse" 응답 → mismatch 에러. | expert-testing | `internal/hook/registry_test.go` (extend, ~40 LOC) | T-RT001-01 | 1 file (extend) | RED — mismatch detection wiring 부재 |
| T-RT001-09 | `registry_test.go` 확장: `TestDispatch_PreToolUseUpdatedInputThenPermissionDecision` (AC-10, REQ-041) + `TestDispatch_PreToolUseUpdatedInputOnly`. UpdatedInput 우선 적용 + deny 결정 ordering 검증. | expert-testing | same file (extend, ~50 LOC) | T-RT001-01 | 1 file (extend) | RED — ordering 명시 부재 |
| T-RT001-10 | `registry_test.go` 확장: `TestDispatch_AdditionalContextRoutedToModelContext` (AC-02, REQ-012). SessionStart/UserPromptSubmit/PreToolUse/PostToolUse AdditionalContext 가 model context queue 로 전달 검증. | expert-testing | same file (extend, ~40 LOC) | T-RT001-01 | 1 file (extend) | RED — routing 부분 구현 |
| T-RT001-11 | `registry_test.go` 확장: `TestDispatch_ContinueFalseBlocksTeammateIdle` (AC-08, REQ-014). SubagentStop + Continue:false → teammate idle blocker 검증. | expert-testing | same file (extend, ~30 LOC) | T-RT001-01 | 1 file (extend) | RED — Continue:false escalation 명시 부재 |
| T-RT001-12 | 신규 파일 `internal/hook/strict_mode_test.go` 생성: `TestStrictMode_ConfigLoad` (system.yaml `hook.strict_mode: true` 로딩) + `TestStrictMode_ErrorReturnedOnLegacyOutput` (AC-06) + `TestIsLegacyEnvSet_LooseTruthy` (AC-04). | expert-testing | new file (~80 LOC) | T-RT001-01 | 1 file (create) | RED — strict_mode 모듈 부재 |
| T-RT001-13 | `strict_mode_test.go` 확장: `TestStrictMode_PluginSourceBypass` (AC-14, REQ-051) + `TestIsPluginSource_NonPluginSources`. ConfigurationSource enum 5개 값 분기 검증. | expert-testing | same file (extend, ~40 LOC) | T-RT001-12 | 1 file (extend) | RED |
| T-RT001-14 | 신규 파일 `internal/hook/banner_test.go` 생성: `TestBanner_OncePerSession` (REQ-007, REQ-050) + `TestBanner_PerSessionIsolation` (sessionID-A vs sessionID-B 격리) + `TestBanner_ConcurrentSafety` (sync.Map atomic). | expert-testing | new file (~70 LOC) | T-RT001-01 | 1 file (create) | RED — banner 모듈 부재 |
| T-RT001-15 | `banner_test.go` 확장: `TestBanner_SuppressedByMoaiHookLegacy` (AC-04). 환경변수 + new session 시 banner 미발사 검증. | expert-testing | same file (extend, ~20 LOC) | T-RT001-14 | 1 file (extend) | RED |
| T-RT001-16 | 신규 파일 `internal/hook/api_version_test.go` 생성: `TestApiVersion2_ParsesShellHeader` (regex match `# moai-hook-api-version: N`) + `TestApiVersion2_SkipsExitCodeFallback` (AC-11) + `TestApiVersion1_DefaultBehavior` (헤더 없는 wrapper → api_version 1). | expert-testing | new file (~60 LOC) | T-RT001-01 | 1 file (create) | RED — api_version 모듈 부재 |
| T-RT001-17 | 신규 파일 `internal/hook/post_tool_mx_test.go` 확장 (이미 존재): `TestPostTool_RoutesMXMarkers` 5개 마커 (NOTE/WARN/ANCHOR/TODO/LEGACY) + 마커 부재 시 routing skip 검증. | expert-testing | `internal/hook/post_tool_mx_test.go` (extend, ~40 LOC) | T-RT001-01 | 1 file (extend) | RED — @MX routing point 부재 |
| T-RT001-18 | 전체 RED 검증: `go test ./internal/hook/` 실행하여 위 18개 새 테스트 모두 실패 확인. 기존 테스트 GREEN 회귀 0 (regression sentinel). 결과를 `progress.md` 의 "M1 RED gate" 항목에 기록. | manager-tdd | n/a (verification only) | T-RT001-01..17 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. TDD discipline.

T-RT001-01..17 는 file region 이 독립이면 parallel 가능. 동일 파일 (`dual_parse_test.go`, `response_test.go`, `registry_test.go`) 확장 작업은 sequential.

---

## M2: validator/v10 + HookResponse Validate() (GREEN, part 1) — Priority P0

Goal: validator/v10 의존성 추가 + HookResponse 에 schema 태그 + Validate() 메서드. AC-12 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT001-19 | `go.mod` 에 `github.com/go-playground/validator/v10` 직접 의존성 추가. `cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001 && go get github.com/go-playground/validator/v10`. `go mod tidy`. | expert-backend | `go.mod`, `go.sum` (regenerated) | T-RT001-18 | 2 files (edit) | Dependency setup |
| T-RT001-20 | `internal/hook/response.go:11-44` 에 validator 태그 추가: `PermissionDecision` → `validate:"omitempty,oneof=allow ask deny defer"`. line 64-70 `RetryHint` 에 `Attempts` → `validate:"omitempty,gte=0,lte=10"`, `Backoff` → `validate:"omitempty,oneofci=linear exponential constant"`. | expert-backend | `internal/hook/response.go` (extend, ~10 LOC) | T-RT001-19 | 1 file (edit) | GREEN — validator surface 노출 |
| T-RT001-21 | `internal/hook/response.go` 패키지 끝에 `var validate = validator.New()` + `func (r *HookResponse) Validate() error { if r == nil { return nil }; return validate.Struct(r) }` 추가. | expert-backend | same file (extend, ~10 LOC) | T-RT001-20 | 1 file (edit) | GREEN — Validate() 메서드 |
| T-RT001-22 | `internal/hook/dual_parse.go:88-115` `ValidateHookResponse()` 에 `r.Validate()` 호출 추가 (기존 PermissionDecision string 비교 외 추가 layer). validator error 는 offending field 명을 포함하여 wrap. | expert-backend | `internal/hook/dual_parse.go` (extend, ~10 LOC) | T-RT001-21 | 1 file (edit) | GREEN — AC-12 wiring |

**M2 priority: P0** — blocks M3-M5. AC-12 GREEN. T-RT001-19 → 20 → 21 → 22 sequential.

---

## M3: Strict mode + MOAI_HOOK_LEGACY env + Deprecation Banner (GREEN, part 2) — Priority P0

Goal: strict_mode + env opt-out + once-per-session banner. AC-04, AC-06, AC-15 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT001-23 | 신규 파일 `internal/hook/strict_mode.go` 생성. 4개 함수: `IsStrictModeEnabled(cfg ConfigProvider) bool` / `IsLegacyEnvSet() bool` (`MOAI_HOOK_LEGACY=1\|true\|yes` loose-truthy) / `EnforceStrictMode(stdout []byte, exitCode int, input *HookInput) error` / `IsPluginSource(input *HookInput) bool` (HookInput.ConfigurationSource == "plugin"). | expert-backend | new file (~70 LOC) | T-RT001-22 | 1 file (create) | GREEN — strict_mode 모듈 |
| T-RT001-24 | 신규 파일 `internal/hook/banner.go` 생성. `var emittedBanners sync.Map` (sessionID → bool) + `func MaybeEmitDeprecationBanner(sessionID string) (banner string, emitted bool)` (LoadOrStore atomic, MOAI_HOOK_LEGACY=1 suppress, 1회 emit). banner 텍스트: `"DEPRECATED: Hook produced exit-code-only output. Migrate to JSON output before v4.0. See https://moai.ai.kr/docs/hook-migration."`. | expert-backend | new file (~50 LOC) | T-RT001-23 | 1 file (create) | GREEN — banner 모듈 |
| T-RT001-25 | `internal/hook/dual_parse.go:47-62` `ParseHookOutput()` 확장: 인자에 `*HookInput` 추가 (또는 `sessionID, configSource string` 분리). JSON parse 실패 분기에서 strict_mode 검사 + plugin bypass + banner emission 통합. signature 변경은 호환성을 위해 `ParseHookOutputWithContext(stdout, exitCode, stderr, input)` 새 함수로 추가하고 기존 `ParseHookOutput` 은 wrapper 로 유지. | expert-backend | `internal/hook/dual_parse.go` (extend, ~30 LOC) | T-RT001-24 | 1 file (edit) | GREEN — REQ-005, REQ-021, REQ-007, REQ-050, REQ-051 wiring |
| T-RT001-26 | `internal/config/types.go` 에 `HookConfig{StrictMode bool}` struct 추가 (`mapstructure:"strict_mode"`, `yaml:"strict_mode"` 태그). `Config` 구조체에 `Hook HookConfig` 필드 추가. | expert-backend | `internal/config/types.go` (extend, ~10 LOC) | T-RT001-23 | 1 file (edit) | GREEN — config struct |
| T-RT001-27 | `internal/config/loader.go` (또는 동등 sections 로더) 가 `system.yaml` 의 `hook.strict_mode` 키를 `Config.Hook.StrictMode` 에 매핑하도록 확장. 기본값 `false` (zero value). | expert-backend | `internal/config/loader.go` (extend, ~15 LOC) | T-RT001-26 | 1 file (edit) | GREEN — config 로딩 |
| T-RT001-28 | `.moai/config/sections/system.yaml` 끝에 다음 추가:<br>```yaml\nhook:\n  strict_mode: false  # SPEC-V3R2-RT-001: opt into JSON-only enforcement; default off\n```<br>그리고 동일 내용을 `internal/template/templates/.moai/config/sections/system.yaml` 에 적용 (Template-First per `CLAUDE.local.md` §2). | expert-backend | 2 files (edit) | T-RT001-27 | 2 files (edit) | GREEN — Template-First 동기화 |

**M3 priority: P0** — blocks M4. AC-04, AC-06, AC-15 GREEN. T-RT001-23, 24, 26 parallel; 25 deps on 23+24; 27 deps on 26; 28 deps on 27.

[HARD] T-RT001-28 은 template 변경이므로 후속 `make build` 시 `internal/template/embedded.go` 재생성 필요 (M5e 에서).

---

## M4: HookSpecificOutputMismatch + UpdatedInput-then-Decision + AdditionalContext routing + Continue:false escalation (GREEN, part 3) — Priority P0

Goal: registry.go Dispatch 의 4가지 의미 강화. AC-02, AC-05, AC-08, AC-09, AC-10 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT001-29 | `internal/hook/registry.go:65-200` `Dispatch()` 에 mismatch detection 삽입. PreToolUse/PostToolUse/UserPromptSubmit/PermissionRequest 응답 후 `HookSpecificOutput.HookEventName != string(event)` 검사 (빈 값은 skip — legacy compat). 불일치 시 `&HookSpecificOutputMismatch{Expected: string(event), Actual: hookEventName}` 에러 반환 + `.moai/logs/hook.log` append (lazy file open + JSON-line format). hook 자체는 failed 처리. | expert-backend | `internal/hook/registry.go` (extend, ~30 LOC) | T-RT001-28 | 1 file (edit) | GREEN — REQ-040 |
| T-RT001-30 | `registry.go:179-193` PreToolUse 경로 정리. 응답에 `UpdatedInput` 와 `PermissionDecision` 둘 다 있을 때 ordering: (1) UpdatedInput 을 pending tool input 에 merge → (2) PermissionDecision 평가 against updated input. 명시적 주석:<br>```go\n// SPEC-V3R2-RT-001 REQ-041: deterministic order — UpdatedInput first, then PermissionDecision\n```<br>denial message 가 post-update input 을 reference. | expert-backend | same file (extend, ~15 LOC) | T-RT001-29 | 1 file (edit) | GREEN — REQ-041, AC-10 |
| T-RT001-31 | `registry.go:155-162` AdditionalContext merge 강화. SessionStart, UserPromptSubmit, PreToolUse, PostToolUse 응답의 AdditionalContext 를 firing order 로 누적. 누적된 텍스트는 별도 함수 `AccumulateAdditionalContext(events []*HookOutput) string` 으로 추출. model context queue 와의 interface 만 노출 (실제 큐 소비는 SPEC-V3R2-SPC-002). | expert-backend | same file + 신규 helper function (~25 LOC) | T-RT001-29 | 1 file (edit) | GREEN — REQ-012 |
| T-RT001-32 | `registry.go` Stop / SubagentStop 경로 강화. hook 가 `Continue: false` 반환 시 (1) teammate idle 차단 (`subagent_stop.go::Handle` 와 정합), (2) orchestrator 에 BlockerReport 발행, (3) AskUserQuestion 라우팅은 orchestrator 책임 (subagent 는 직접 호출 금지). 본 작업은 registry 의 mapping 만 정확히 표현; 실제 BlockerReport 발행은 별도 SPEC (RT-004 SessionStore 통합) 에서 진행. | expert-backend | `internal/hook/registry.go` (extend, ~15 LOC) | T-RT001-30 | 1 file (edit) | GREEN — REQ-014, AC-08 |
| T-RT001-33 | `internal/hook/dual_parse.go:88-115` `ValidateHookResponse()` 의 64 KiB 절단 메시지 정규화. 현재 `\n[Notice: additionalContext truncated from %d to %d bytes]` 를 유지하되 SystemMessage prefix 가 아닌 suffix 로 명시 (이미 그렇게 동작 중; line 110-112 의 string concat 검증). REQ-022 sentinel 정확성 보장. | expert-backend | `internal/hook/dual_parse.go` (refactor, ~5 LOC) | T-RT001-29 | 1 file (edit) | REFACTOR — REQ-022, AC-09 |

**M4 priority: P0** — blocks M5. AC-02, AC-05, AC-08, AC-09, AC-10 GREEN. T-RT001-29 → 30, 31, 32 parallel; 33 independent.

---

## M5: api_version 2 + WatchPaths + @MX routing + plugin bypass + CHANGELOG + MX tags (GREEN, part 4 + Trackable) — Priority P1

Goal: User-facing surface, future-proofing, trackability. AC-01, AC-03, AC-07, AC-11, AC-13, AC-14 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT001-34 | 신규 파일 `internal/hook/api_version.go` 생성. 2개 함수:<br>- `ReadHookApiVersion(wrapperPath string) (int, error)`: 첫 30 줄 내 정규식 `^#\s*moai-hook-api-version:\s*(\d+)\s*$` 매칭, 미매칭 시 default 1.<br>- `ShouldSkipExitCodeFallback(wrapperPath string) bool`: api_version >= 2 시 true. | expert-backend | new file (~50 LOC) | T-RT001-32 | 1 file (create) | GREEN — REQ-030, AC-11 |
| T-RT001-35 | `internal/hook/dual_parse.go` `ParseHookOutputWithContext` 확장. wrapperPath 인자 (또는 환경변수 `MOAI_HOOK_WRAPPER_PATH`) 가 제공되면 `ShouldSkipExitCodeFallback` 호출. true + 빈 stdout 이면 explicit no-op `&HookResponse{}` 반환 (NO synthesizeFromExitCode). | expert-backend | `internal/hook/dual_parse.go` (extend, ~15 LOC) | T-RT001-34 | 1 file (edit) | GREEN — AC-11 |
| T-RT001-36 | 신규 파일 `internal/hook/watch_paths.go` 생성. `type WatchPathsRegistrar interface { Register(paths []string) error }`. nil-safe wrapping (registrar 가 nil 이면 graceful no-op + warn log). Registry 에 `WithWatchPathsRegistrar(r WatchPathsRegistrar)` 옵션 함수 노출. | expert-backend | new file (~30 LOC) | T-RT001-32 | 1 file (create) | GREEN — REQ-032 |
| T-RT001-37 | `internal/hook/session_start.go` 의 SessionStart handler 가 `HookResponse.WatchPaths` 를 emit 하면 `WatchPathsRegistrar.Register(paths)` 호출 (nil 인 경우 graceful skip). 본 SPEC 는 contract 만 노출; 실제 file-system watcher 구현은 별도 SPEC. | expert-backend | `internal/hook/session_start.go` (extend, ~15 LOC) | T-RT001-36 | 1 file (edit) | GREEN — REQ-032 wiring |
| T-RT001-38 | `internal/hook/post_tool.go` PostToolUse handler 확장. `HookResponse.AdditionalContext` 가 5개 마커 (`@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`, `@MX:TODO`, `@MX:LEGACY`) 중 하나라도 포함하면 `internal/hook/mx/ingest.go` (이미 존재) 의 ingestion 함수에 forward. 본 SPEC 는 routing point 만; 의미론은 SPEC-V3R2-SPC-002. | expert-backend | `internal/hook/post_tool.go` (extend, ~20 LOC) | T-RT001-32 | 1 file (edit) | GREEN — REQ-016, AC-07 |
| T-RT001-39 | CHANGELOG entry 추가. `## [Unreleased] / ### Changed (BREAKING — BC-V3R2-001)` 섹션 (없으면 신규 추가) 아래:<br>```markdown\n- SPEC-V3R2-RT-001: Hook output protocol now follows Claude Code 2026.x JSON-OR-ExitCode dual contract. JSON output is parsed first; exit-code semantics remain as v3.x deprecation-window fallback. New `hook.strict_mode` config (default `false`) opts into JSON-only enforcement. New `MOAI_HOOK_LEGACY=1` opt-out for CI/air-gapped installs. Deprecation banner emits once per session per project. Wrappers unchanged; handlers rewritten. Removal of exit-code-only path deferred to v4.0.\n``` | manager-docs | `CHANGELOG.md` (extend, ~7 LOC) | T-RT001-38 | 1 file (edit) | Trackable |
| T-RT001-40 | MX 태그 7개 삽입 per `plan.md` §6: 3 ANCHOR (`dual_parse.go:47` ParseHookOutput, `response.go:11` HookResponse, `registry.go:65` Dispatch) + 2 NOTE (`strict_mode.go` EnforceStrictMode, `banner.go` MaybeEmitDeprecationBanner) + 2 WARN (`dual_parse.go:64` synthesizeFromExitCode, `registry.go` mismatch detection insertion). 태그 본문은 `plan.md` §6 verbatim. | manager-docs | 5 files (edit, MX 태그 삽입) | T-RT001-39 | 5 files (edit) | Trackable — plan §6 |
| T-RT001-41 | `make build` 을 worktree root 에서 실행. `internal/template/embedded.go` 재생성 검증 — `system.yaml` template 변경 (T-RT001-28) 만 diff scope (다른 .claude/ 변경 없음). | manager-docs | `internal/template/embedded.go` (regenerated) | T-RT001-40 | 1 file (regenerated) | Build verification |
| T-RT001-42 | 전체 `go test ./...` 을 worktree root 에서 실행. 모든 audit / wire-format / dual-parse / strict-mode / banner / api-version / registry 테스트 PASS + 0 cascading regression (per `CLAUDE.local.md` §6 HARD). `go vet ./...` + `golangci-lint run` 경고 0. | manager-tdd | n/a (verification only) | T-RT001-41 | 0 files | GREEN gate (final) |
| T-RT001-43 | `progress.md` 업데이트: `run_complete_at: <timestamp>` + `run_status: implementation-complete`. ACs 통과 수 (15/15) + new tests (~28) + MX tags (7) + PR number 채움. | manager-docs | `progress.md` (extend) | T-RT001-42 | 1 file (edit) | Trackable closure |

**M5 priority: P1** — completes the SPEC. AC-01, AC-03, AC-07, AC-11, AC-13, AC-14 GREEN.

T-RT001-34, 36, 38 parallel. T-RT001-35 deps 34. T-RT001-37 deps 36. T-RT001-39 → 40 → 41 → 42 → 43 sequential.

---

## Task summary by milestone

| Milestone | Task IDs | Total tasks | Priority | Owner role mix |
|-----------|----------|-------------|----------|----------------|
| M1 (RED) | T-RT001-01..18 | 18 | P0 | expert-testing (17) + manager-tdd (1 verification) |
| M2 (GREEN part 1) | T-RT001-19..22 | 4 | P0 | expert-backend |
| M3 (GREEN part 2) | T-RT001-23..28 | 6 | P0 | expert-backend |
| M4 (GREEN part 3) | T-RT001-29..33 | 5 | P0 | expert-backend |
| M5 (GREEN part 4 + REFACTOR + Trackable) | T-RT001-34..43 | 10 | P1 | expert-backend (5) + manager-docs (4) + manager-tdd (1 verification) |
| **TOTAL** | T-RT001-01..43 | **43 tasks** | — | — |

> NOTE: 43 tasks span the 5 milestones. M1 (RED) opens with 18 test-only tasks; M2-M4 (GREEN core) deliver the dual-protocol subsystem in 15 tasks; M5 (final GREEN + REFACTOR + Trackable) closes with 10 tasks of CLI/CHANGELOG/MX/build/verify.

---

## Dependency graph (critical path)

```
T-RT001-01..17 (M1 tests, parallel within file regions)
   ↓
T-RT001-18 (M1 verification gate — 18 RED + 0 regression)
   ↓
T-RT001-19 (go.mod validator/v10) → T-RT001-20 (response.go 태그) → T-RT001-21 (Validate 메서드) → T-RT001-22 (dual_parse 통합)
   ↓
T-RT001-23, 24, 26 (M3 parallel)
   ↓
T-RT001-25 (ParseHookOutputWithContext)
T-RT001-27 (loader 확장) → T-RT001-28 (system.yaml + template)
   ↓
T-RT001-29 (mismatch detection) → T-RT001-30, 31, 32 (parallel)
T-RT001-33 (truncation refactor, parallel from 29)
   ↓
T-RT001-34, 36, 38 (M5 core parallel)
T-RT001-35 (api_version wiring) deps 34
T-RT001-37 (WatchPaths wiring) deps 36
   ↓
T-RT001-39 (CHANGELOG) → T-RT001-40 (7 MX tags) → T-RT001-41 (make build) → T-RT001-42 (go test full + vet + lint) → T-RT001-43 (progress.md closure)
```

Critical path: T-RT001-18 → 19 → 20 → 21 → 22 → 25 → 28 → 29 → 30 → 35 → 39 → 40 → 41 → 42 → 43 (15 sequential gates).

---

## Cross-task constraints

[HARD] All file edits use absolute paths under the worktree root `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001` per CLAUDE.md §Worktree Isolation Rules.

[HARD] All tests use `t.TempDir()` per `CLAUDE.local.md` §6 — no test creates files in the project root. environment 검사 테스트는 `t.Setenv` 사용.

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; no `filepath.Join(cwd, absPath)` patterns per `CLAUDE.local.md` §6.

[HARD] No new direct module dependencies beyond `github.com/go-playground/validator/v10` (added by T-RT001-19; SCH-001 후속 머지 시 dedup).

[HARD] No `.claude/hooks/moai/*.sh` 파일 수정 — wrappers-unchanged 정책. master §8 BC-V3R2-001 의 핵심 약속.

[HARD] Code comments in Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`). Godoc / 외부 노출 식별자 docstring 은 영어 유지 (industry standard).

[HARD] Commit messages in Korean (per `.moai/config/sections/language.yaml` `git_commit_messages: ko`).

[HARD] Subagent (manager-tdd, expert-backend, expert-testing, manager-docs) 는 `AskUserQuestion` 직접 호출 금지. user 결정 필요 시 BlockerReport 형태로 orchestrator 에 반환 (per agent-common-protocol.md §User Interaction Boundary).

[HARD] `CHANGELOG.md` entry 는 BREAKING 라벨 + BC-V3R2-001 명시 + v4.0 sunset note 포함.

---

End of tasks.md.
