---
id: SPEC-MX-002
version: 1.0.0
status: draft
created: 2026-03-11
updated: 2026-03-11
author: GOOS
priority: medium
---

# SPEC-MX-002 Acceptance Criteria

## AC-VAL-001: MX Validation Core - ANCHOR Detection

**Given** fan_in >= 3인 exported 함수가 있는 Go 파일이 존재하고
**And** 해당 함수에 @MX:ANCHOR 태그가 없을 때
**When** ValidateFile을 해당 파일에 대해 실행하면
**Then** P1 (Blocking) 위반이 감지되어야 한다
**And** 위반에 함수명, 파일 경로, 라인 번호, fan_in 값이 포함되어야 한다

---

**Given** fan_in >= 3인 exported 함수가 있는 Go 파일이 존재하고
**And** 해당 함수에 @MX:ANCHOR 태그가 이미 있을 때
**When** ValidateFile을 해당 파일에 대해 실행하면
**Then** 해당 함수에 대한 위반이 감지되지 않아야 한다

---

## AC-VAL-002: MX Validation Core - Priority Classification

**Given** 다음 조건의 함수들이 포함된 Go 파일이 있을 때:
  - fan_in=5 함수 (ANCHOR 누락)
  - goroutine 패턴 (WARN 누락)
  - 150줄 exported 함수 (NOTE 누락)
  - 테스트 없는 public 함수 (TODO 누락)
**When** ValidateFile을 실행하면
**Then** P1 위반 1건, P2 위반 1건, P3 위반 1건, P4 위반 1건이 분류되어야 한다
**And** P1, P2는 `blocking: true`로 표시되어야 한다
**And** P3, P4는 `blocking: false`로 표시되어야 한다

---

## AC-POST-001: PostToolUse MX Metrics

**Given** PostToolUse 핸들러에 MX Validator가 설정되어 있고
**And** Write 도구로 Go 파일이 수정되었을 때
**When** PostToolUse 훅이 실행되면
**Then** 반환된 메트릭에 `mx_validation` 키가 존재해야 한다
**And** `mx_validation.status`가 `pass`, `warn`, 또는 `fail` 중 하나여야 한다
**And** `mx_validation.violations` 배열이 존재해야 한다
**And** `mx_validation.duration_ms`가 0 이상이어야 한다

---

## AC-POST-002: PostToolUse Non-Blocking

**Given** MX 검증이 500ms 이상 소요되는 상황에서
**When** PostToolUse 훅이 실행되면
**Then** 훅은 500ms 이내에 응답을 반환해야 한다
**And** `mx_validation.status`가 `"skipped"`여야 한다
**And** 도구 실행이 차단되지 않아야 한다

---

## AC-SESSION-001: SessionEnd Batch Validation

**Given** 세션 중 3개의 Go 파일이 수정되었고
**And** 그 중 1개 파일에 fan_in >= 3 함수의 @MX:ANCHOR가 누락되어 있을 때
**When** 세션이 종료되면
**Then** SessionEnd 핸들러가 3개 파일에 대해 MX 검증을 실행해야 한다
**And** 검증 결과가 slog.Info로 로그에 출력되어야 한다
**And** P1 위반 1건이 리포트에 포함되어야 한다
**And** 세션 종료가 정상적으로 완료되어야 한다 (차단 없음)

---

## AC-SESSION-002: SessionEnd Non-Blocking on Failure

**Given** MX 검증 중 ast-grep 실행 실패가 발생했을 때
**When** 세션이 종료되면
**Then** 세션 종료가 정상적으로 완료되어야 한다
**And** 검증 오류가 slog.Warn으로 로그에 기록되어야 한다
**And** 반환되는 HookOutput이 비어 있어야 한다 (SessionEnd 프로토콜 준수)

---

## AC-SESSION-003: SessionEnd Performance Budget

**Given** 10개의 Go 파일이 수정된 세션에서
**When** 세션이 종료되면
**Then** MX 검증이 4초 이내에 완료되어야 한다
**And** 4초 초과 시 이미 완료된 파일의 결과만 리포트되어야 한다
**And** 타임아웃된 파일 목록이 로그에 표시되어야 한다

---

## AC-SYNC-001: Sync P1/P2 Blocking

**Given** SPEC 범위 파일에 P1 위반 (ANCHOR 누락)이 존재하고
**And** `--skip-mx` 플래그가 제공되지 않았을 때
**When** `/moai sync` Phase 0.6이 실행되면
**Then** sync가 차단되어야 한다
**And** 상세한 위반 리포트가 사용자에게 표시되어야 한다
**And** 리포트에 "Run /moai run to add missing tags, or use --skip-mx to bypass" 안내가 포함되어야 한다

---

## AC-SYNC-002: Sync P3/P4 Advisory Only

**Given** SPEC 범위 파일에 P3 위반만 존재하고
**And** P1/P2 위반이 없을 때
**When** `/moai sync` Phase 0.6이 실행되면
**Then** 경고 메시지가 출력되어야 한다
**And** sync가 계속 진행되어야 한다

---

## AC-SYNC-003: Sync Skip Flag

**Given** SPEC 범위 파일에 P1 위반이 존재하고
**And** `--skip-mx` 플래그가 제공되었을 때
**When** `/moai sync` Phase 0.6이 실행되면
**Then** sync가 계속 진행되어야 한다
**And** 리포트에 "MX validation skipped by user flag" 사실이 명시적으로 기록되어야 한다

---

## AC-CONFIG-001: Validation Config Parsing

**Given** mx.yaml에 다음 validation 섹션이 있을 때:
```yaml
validation:
  enabled: true
  post_tool_use:
    enabled: false
    timeout_ms: 300
  session_end:
    enabled: true
    timeout_ms: 3000
  sync:
    enforcement: advisory
  enforcement_levels:
    p1_anchor: blocking
    p2_warn: advisory
```
**When** ParseValidationConfig를 실행하면
**Then** post_tool_use.enabled가 false여야 한다
**And** post_tool_use.timeout_ms가 300이어야 한다
**And** session_end.timeout_ms가 3000이어야 한다
**And** sync.enforcement가 "advisory"여야 한다
**And** p2_warn enforcement_level이 "advisory"여야 한다

---

## AC-CONFIG-002: Validation Config Defaults

**Given** mx.yaml에 validation 섹션이 없을 때
**When** ParseValidationConfig를 실행하면
**Then** enabled가 true여야 한다 (기본값)
**And** post_tool_use.timeout_ms가 500이어야 한다 (기본값)
**And** session_end.timeout_ms가 4000이어야 한다 (기본값)
**And** sync.enforcement가 "strict"여야 한다 (기본값)
**And** p1_anchor가 "blocking"이어야 한다 (기본값)

---

## AC-EDGE-001: AST-grep Unavailable Fallback

**Given** ast-grep (`sg`) CLI가 설치되지 않은 환경에서
**When** ValidateFile을 실행하면
**Then** Grep 기반 fallback으로 검증이 수행되어야 한다
**And** fan_in 분석과 goroutine 패턴 감지가 동작해야 한다
**And** 검증 결과에 `fallback: true` 표시가 있어야 한다

---

## AC-EDGE-002: Existing Tag Preservation

**Given** @MX:ANCHOR, @MX:WARN 등 기존 태그가 있는 Go 파일에서
**When** ValidateFile 또는 ValidateFiles를 실행하면
**Then** 기존 @MX 태그가 수정되거나 삭제되지 않아야 한다
**And** 파일 내용이 검증 전후로 동일해야 한다 (SHA-256 비교)

---

## AC-EDGE-003: Thread Safety

**Given** 2개의 고루틴에서 동일한 Validator 인스턴스를 사용할 때
**When** 각각 다른 파일에 대해 ValidateFile을 동시 호출하면
**Then** race condition 없이 두 검증이 완료되어야 한다
**And** `go test -race`에서 경고가 발생하지 않아야 한다

---

## AC-EDGE-004: Empty Project

**Given** Go 파일이 없는 프로젝트 디렉토리에서
**When** ValidateFiles를 빈 파일 목록으로 호출하면
**Then** 빈 ValidationReport가 반환되어야 한다
**And** violations가 빈 배열이어야 한다
**And** 에러 없이 정상 완료되어야 한다

---

## AC-EDGE-005: Partial Validation on Timeout

**Given** 10개의 Go 파일이 검증 대상이고
**And** 타임아웃이 2초로 설정되어 있을 때
**When** 5번째 파일 검증 중 타임아웃이 발생하면
**Then** 1-4번째 파일의 검증 결과가 반환되어야 한다
**And** 5-10번째 파일이 "timed_out" 상태로 리포트되어야 한다
**And** 에러가 반환되지 않아야 한다 (부분 성공)

---

## AC-REPORT-001: Standard Report Format

**Given** 3개 파일에서 P1 1건, P2 2건, P3 1건 위반이 감지되었을 때
**When** FormatReport를 호출하면
**Then** Summary 섹션에 total files, P1-P4 위반 수, duration이 포함되어야 한다
**And** P1 Violations 섹션에 파일명:라인 형식의 위반 상세가 있어야 한다
**And** P2 Violations 섹션에 위반 상세가 있어야 한다
**And** P3 Violations 섹션에 위반 상세가 있어야 한다

---

## Quality Gate Criteria

### Definition of Done

- [ ] 모든 테스트 통과 (`go test -race ./internal/hook/mx/...`)
- [ ] 테스트 커버리지 85% 이상
- [ ] `go vet ./...` 경고 없음
- [ ] `golangci-lint run` 경고 없음
- [ ] PostToolUse MX 검증이 500ms 이내 완료
- [ ] SessionEnd MX 검증이 4초 이내 완료
- [ ] 기존 @MX 태그 불변성 보장 (읽기 전용 검증)
- [ ] ast-grep 미설치 시 graceful fallback 동작
- [ ] 동시 접근 시 race condition 없음 (`go test -race`)

### Verification Methods

| Criteria | Verification Method |
|----------|-------------------|
| P1/P2 detection accuracy | Table-driven 테스트 with 알려진 위반 파일 |
| Performance budget | Benchmark 테스트 (`b.N` loop) |
| Thread safety | `go test -race` 플래그 |
| Fallback behavior | ast-grep mock (nil Analyzer 주입) |
| Config parsing | YAML fixture 파일 기반 테스트 |
| Report format | Golden file comparison |
| Non-blocking guarantee | Context timeout 테스트 |
