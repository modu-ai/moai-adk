---
id: SPEC-MX-002
version: 1.0.0
status: draft
created: 2026-03-11
updated: 2026-03-11
author: GOOS
priority: medium
lifecycle: spec-anchored
depends_on: SPEC-MX-001
---

# SPEC-MX-002: MX Tag Auto-Validation System

| Field       | Value                                                |
|-------------|------------------------------------------------------|
| SPEC ID     | SPEC-MX-002                                          |
| Title       | MX Tag Auto-Validation System                        |
| Status      | Draft                                                |
| Created     | 2026-03-11                                           |
| Author      | GOOS                                                 |
| Priority    | Medium                                               |
| Lifecycle   | spec-anchored                                        |
| Depends On  | SPEC-MX-001 (@MX TAG System)                         |

---

## 1. Overview

### 1.1 Problem Statement

SPEC-MX-001은 @MX 태그의 유형, 구문, 수명 주기 규칙, 3-Pass 스캔 알고리즘을 정의한다. 그러나 태그가 **있어야 할 곳에 실제로 존재하는지** 자동으로 검증하는 메커니즘은 없다. 현재 상태:

- **PostToolUse 훅**: AST-grep 통합이 있지만 @MX 인식이 없음 (관찰 전용)
- **SessionEnd 훅**: 세션 종료 시 정리 작업만 수행, MX 검증 없음
- **Sync 워크플로우 Phase 0.6**: MX 검증이 존재하지만 선택적(optional)이며 강제되지 않음

이로 인해 다음 문제가 발생한다:

- fan_in >= 3인 함수에 @MX:ANCHOR가 누락되어도 감지되지 않음
- goroutine 패턴에 @MX:WARN이 없어도 세션이 정상 종료됨
- sync 단계에서 MX 위반이 있어도 PR 생성이 차단되지 않음

### 1.2 Solution

3개 통합 지점에서 @MX 태그 자동 검증을 수행하는 시스템:

1. **PostToolUse MX 인식** -- Write/Edit 후 MX 검증 결과를 메트릭에 포함 (관찰 전용)
2. **SessionEnd 배치 검증** -- 세션 종료 시 수정된 모든 파일에 대해 MX 태그 누락 검사 (관찰 전용, 세션 차단 없음)
3. **Sync Phase 0.6 강제** -- P1/P2 MX 위반 시 sync를 차단하여 PR 품질 보장

### 1.3 Scope

**In Scope:**
- MX 검증 모듈 (`internal/hook/mx/validator.go`)
- PostToolUse 핸들러에 MX 인식 확장
- SessionEnd 핸들러에 MX 배치 검증 추가
- Sync 워크플로우 Phase 0.6 강제화
- mx.yaml에 `validation` 섹션 추가
- 표준 MX 검증 리포트 포맷

**Out of Scope:**
- @MX 태그 자동 삽입 (SPEC-MX-001 범위)
- IDE/에디터 통합
- 실시간 LSP 기반 MX 진단
- 크로스 리포지토리 MX 검증

---

## 2. Environment

### 2.1 Platform

- MoAI-ADK (Go Edition) v2.6.0+
- Go 1.26+
- ast-grep (`sg`) CLI (기존 설치 필수)

### 2.2 Integration Points

| System                          | Integration Type     | Purpose                                      |
|---------------------------------|---------------------|----------------------------------------------|
| `internal/hook/post_tool.go`    | PostToolUse 훅 확장  | Write/Edit 후 MX 검증 메트릭 추가             |
| `internal/hook/session_end.go`  | SessionEnd 훅 확장   | 세션 종료 시 배치 MX 검증                      |
| `internal/astgrep/analyzer.go`  | AST 분석 도구        | 구조적 패턴 매칭으로 MX 태그 대상 식별          |
| `internal/hook/quality/change_detector.go` | 변경 감지    | SHA-256 기반 수정 파일 추적                    |
| `.moai/config/sections/mx.yaml` | 설정 파일            | 임계값, 제한, 제외 패턴, 검증 수준              |
| `.claude/skills/moai/workflows/sync.md` | Sync 워크플로우 | Phase 0.6 MX 검증 강제화                      |

### 2.3 Dependencies

- `internal/astgrep` 패키지 (기존): AST-grep 래퍼
- `internal/hook/quality` 패키지 (기존): 변경 감지
- `.moai/config/sections/mx.yaml` (기존): MX 설정

---

## 3. Assumptions

- A1: ast-grep (`sg`) CLI가 개발 환경에 설치되어 있다. 미설치 시 MX 검증은 graceful하게 건너뛴다.
- A2: PostToolUse 훅의 30초 타임아웃 내에서 MX 검증이 완료된다 (현재 사용량 <1초, MX 검증 예상 <500ms).
- A3: SessionEnd 훅에서 MX 검증에 약 4초의 예산이 남아 있다 (tmux 정리 후).
- A4: 검증 대상 파일은 세션 중 수정된 파일로 제한된다 (전체 코드베이스 스캔 아님).
- A5: mx.yaml의 기존 thresholds, limits, exclude 설정이 검증에도 동일하게 적용된다.
- A6: Sync Phase 0.6 강제화는 `--skip-mx` 플래그로 명시적 우회가 가능하다.

---

## 4. Requirements

### 4.1 MX Validation Module

#### REQ-VAL-001: MX Validation Core Module

시스템은 `internal/hook/mx/validator.go`에 MX 검증 코어 모듈을 **항상** 제공해야 한다.

검증 모듈은 다음을 수행한다:
- 주어진 Go 파일에서 fan_in >= `thresholds.fan_in_anchor`인 함수에 @MX:ANCHOR 존재 여부 확인
- 주어진 Go 파일에서 goroutine 패턴(`go func`, `go `)에 @MX:WARN 존재 여부 확인
- 주어진 Go 파일에서 cyclomatic complexity >= `thresholds.complexity_warn`인 함수에 @MX:WARN 존재 여부 확인
- mx.yaml의 `exclude` 패턴에 매칭되는 파일은 검증에서 제외
- 검증 결과를 표준 `ValidationReport` 구조체로 반환

#### REQ-VAL-002: Validation Priority Classification

시스템은 MX 검증 위반을 다음 우선순위로 분류해야 한다:

- **P1 (Blocking)**: fan_in >= `thresholds.fan_in_anchor`인 함수에 @MX:ANCHOR 누락
- **P2 (Blocking)**: goroutine 또는 complexity >= `thresholds.complexity_warn`에 @MX:WARN 누락
- **P3 (Advisory)**: 100줄 초과 exported 함수에 @MX:NOTE 누락
- **P4 (Advisory)**: 테스트 없는 public 함수에 @MX:TODO 누락

**WHEN** P1 또는 P2 위반이 감지되면, **THEN** 해당 위반은 sync 단계에서 차단 사유가 된다.
**WHEN** P3 또는 P4 위반이 감지되면, **THEN** 해당 위반은 리포트에만 포함되고 차단하지 않는다.

### 4.2 PostToolUse MX Awareness

#### REQ-POST-001: PostToolUse MX Metrics Extension

**WHEN** Write 또는 Edit 도구 사용 후, **THEN** PostToolUse 핸들러는 기존 AST-grep 스캔 결과에 MX 검증 결과를 추가해야 한다.

메트릭에 포함할 필드:
- `mx_validation.violations`: 위반 목록 (priority, type, file, line, description)
- `mx_validation.status`: `pass` | `warn` | `fail`
- `mx_validation.duration_ms`: 검증 소요 시간

**IF** MX 검증에 500ms 이상 소요되면, **THEN** 시스템은 검증을 건너뛰고 `mx_validation.status: "skipped"` 및 사유를 리포트해야 한다.

#### REQ-POST-002: PostToolUse Non-Blocking

시스템은 PostToolUse MX 검증에서 도구 실행을 차단**하지 않아야 한다**. MX 검증 결과는 관찰 전용 메트릭으로만 제공된다.

### 4.3 SessionEnd MX Batch Validation

#### REQ-SESSION-001: SessionEnd Batch MX Validation

**WHEN** 세션이 종료되면, **THEN** SessionEnd 핸들러는 세션 중 수정된 모든 .go 파일에 대해 MX 배치 검증을 수행해야 한다.

검증 프로세스:
1. `git diff --name-only HEAD` 또는 change_detector 캐시에서 수정된 파일 목록 획득
2. mx.yaml `exclude` 패턴에 매칭되는 파일 제외
3. 각 파일에 대해 MX 검증 실행
4. 검증 리포트를 로그로 출력 (slog.Info)

#### REQ-SESSION-002: SessionEnd Non-Blocking

시스템은 SessionEnd MX 검증에서 세션 종료를 차단**하지 않아야 한다**. MX 검증은 관찰 전용이며, 검증 실패가 세션 종료를 방해해서는 안 된다.

#### REQ-SESSION-003: SessionEnd Performance Budget

**WHILE** SessionEnd 훅의 전체 타임아웃이 5초이고 기존 정리 작업(tmux, GLM)이 약 1초를 사용하는 상태에서, **THEN** MX 배치 검증은 최대 4초 내에 완료되어야 한다.

**IF** 4초 내에 검증이 완료되지 않으면, **THEN** 시스템은 검증을 중단하고 부분 결과만 리포트해야 한다.

### 4.4 Sync Phase 0.6 Enforcement

#### REQ-SYNC-001: Sync MX Validation Enforcement

**WHEN** `/moai sync` 실행 시 Phase 0.6에 도달하면, **THEN** MX 검증을 필수(mandatory)로 실행해야 한다.

**WHEN** P1 (ANCHOR 누락) 또는 P2 (WARN 누락) 위반이 감지되면, **THEN** sync를 차단하고 위반 사항을 사용자에게 리포트해야 한다.

**WHEN** P3 또는 P4 위반만 감지되면, **THEN** 경고를 출력하되 sync를 계속 진행해야 한다.

#### REQ-SYNC-002: Sync Skip Flag

**WHERE** 사용자가 `--skip-mx` 플래그를 명시적으로 제공하면, **THEN** P1/P2 위반이 있어도 sync를 계속 진행할 수 있어야 한다. 단, 스킵 사실이 리포트에 명시적으로 기록된다.

#### REQ-SYNC-003: Sync Scope Limitation

Sync Phase 0.6 MX 검증은 현재 SPEC에서 수정된 파일로만 범위를 제한해야 한다. 전체 코드베이스를 검증하지 않는다.

### 4.5 Validation Report Format

#### REQ-REPORT-001: Standard Validation Report Format

시스템은 모든 3개 통합 지점(PostToolUse, SessionEnd, Sync)에서 동일한 리포트 포맷을 사용해야 한다:

```
## MX Validation Report -- [Context] -- [Timestamp]

### Summary
- Total files scanned: N
- P1 violations: N (blocking)
- P2 violations: N (blocking)
- P3 violations: N (advisory)
- P4 violations: N (advisory)
- Duration: Nms

### P1 Violations (ANCHOR Missing)
- file.go:42 func ProcessOrder (fan_in=5, @MX:ANCHOR missing)

### P2 Violations (WARN Missing)
- file.go:88 goroutine without @MX:WARN

### P3 Violations (NOTE Missing)
- file.go:120 exported func (150 lines, no @MX:NOTE)

### P4 Violations (TODO Missing)
- file.go:200 public func with no test file
```

#### REQ-REPORT-002: Report Context Adaptation

**WHEN** PostToolUse 컨텍스트에서 리포트를 생성하면, **THEN** 단일 파일에 대한 간략 리포트를 반환한다.
**WHEN** SessionEnd 컨텍스트에서 리포트를 생성하면, **THEN** 전체 수정 파일에 대한 요약 리포트를 로그에 출력한다.
**WHEN** Sync 컨텍스트에서 리포트를 생성하면, **THEN** 상세 리포트를 사용자에게 표시하고 차단 여부를 명시한다.

### 4.6 Configuration Integration

#### REQ-CONFIG-001: mx.yaml Validation Section

시스템은 기존 mx.yaml에 `validation` 섹션을 추가하여 검증 동작을 제어해야 한다:

```yaml
mx:
  # ... 기존 설정 ...
  validation:
    enabled: true
    post_tool_use:
      enabled: true
      timeout_ms: 500
    session_end:
      enabled: true
      timeout_ms: 4000
    sync:
      enforcement: strict    # strict | advisory | disabled
      skip_flag: "--skip-mx"
    enforcement_levels:
      p1_anchor: blocking     # blocking | advisory | disabled
      p2_warn: blocking        # blocking | advisory | disabled
      p3_note: advisory        # blocking | advisory | disabled
      p4_todo: advisory        # blocking | advisory | disabled
```

**WHEN** `validation.enabled`가 `false`이면, **THEN** 모든 통합 지점에서 MX 검증을 건너뛴다.
**WHEN** 개별 통합 지점의 `enabled`가 `false`이면, **THEN** 해당 지점에서만 검증을 건너뛴다.

#### REQ-CONFIG-002: Configuration Defaults

**WHEN** mx.yaml에 `validation` 섹션이 없으면, **THEN** 시스템은 위 기본값을 사용한다.

### 4.7 Edge Cases

#### REQ-EDGE-001: AST-grep Unavailable

**IF** ast-grep (`sg`) CLI가 설치되지 않았으면, **THEN** MX 검증은 Grep 기반 fallback으로 전환해야 한다. Grep fallback은 정확도가 낮지만 기본적인 fan_in 분석과 goroutine 패턴 감지가 가능하다.

#### REQ-EDGE-002: Existing Tag Preservation

시스템은 검증 과정에서 기존 @MX 태그를 절대 수정하거나 삭제**하지 않아야 한다**. 검증은 읽기 전용(read-only) 작업이다.

#### REQ-EDGE-003: Concurrent File Access

**WHILE** PostToolUse와 SessionEnd가 동시에 실행될 수 있는 상태에서, **THEN** MX 검증 모듈은 스레드 안전(thread-safe)해야 한다.

#### REQ-EDGE-004: Empty Project

**WHEN** 프로젝트에 Go 파일이 없으면, **THEN** MX 검증은 빈 리포트를 반환하고 성공으로 처리한다.

#### REQ-EDGE-005: Partial Validation on Timeout

**IF** SessionEnd 또는 PostToolUse에서 타임아웃이 발생하면, **THEN** 시스템은 이미 완료된 파일의 검증 결과만 리포트하고, 타임아웃 파일 목록을 함께 표시한다.

---

## 5. Specifications

### 5.1 Implementation Architecture

| Component                    | File Location                              | Purpose                                      |
|------------------------------|--------------------------------------------|----------------------------------------------|
| MX Validator Core            | `internal/hook/mx/validator.go`            | MX 검증 로직, 리포트 생성                     |
| MX Validator Types           | `internal/hook/mx/types.go`                | ValidationReport, Violation 등 구조체         |
| MX Config Reader             | `internal/hook/mx/config.go`               | mx.yaml validation 섹션 파싱                  |
| PostToolUse Extension        | `internal/hook/post_tool.go`               | runAstScan 이후 runMxValidation 호출 추가     |
| SessionEnd Extension         | `internal/hook/session_end.go`             | Handle 함수에 MX 배치 검증 추가               |
| Sync Workflow Update         | `.claude/skills/moai/workflows/sync.md`    | Phase 0.6 강제화 규칙 업데이트                |
| Config Template Update       | `.moai/config/sections/mx.yaml`            | validation 섹션 추가                          |
| MX Validator Tests           | `internal/hook/mx/validator_test.go`       | TDD RED-GREEN-REFACTOR                       |

### 5.2 Validator Interface

```go
// Validator defines the interface for MX tag validation.
type Validator interface {
    // ValidateFile checks a single file for MX tag compliance.
    ValidateFile(ctx context.Context, filePath string) (*FileReport, error)

    // ValidateFiles checks multiple files and returns an aggregated report.
    ValidateFiles(ctx context.Context, filePaths []string) (*ValidationReport, error)
}
```

### 5.3 Fan-In Analysis Integration

fan_in 분석은 기존 `internal/astgrep/analyzer.go`의 `FindPattern` 메서드를 활용하거나, SPEC-MX-001의 REQ-SCAN-001에서 정의한 Grep 기반 참조 카운팅을 사용한다:

1. 파일에서 exported 함수 선언 추출
2. `Grep(pattern="<function_name>", type="go", output_mode="count")` 실행
3. 선언 1건 제외 = fan_in 근사치
4. fan_in >= `thresholds.fan_in_anchor`이면 @MX:ANCHOR 존재 여부 확인

### 5.4 Change Detection Integration

수정된 파일 목록 획득 방법 (우선순위 순):

1. **PostToolUse**: `input.ToolInput` 에서 파일 경로 직접 추출
2. **SessionEnd**: `git diff --name-only HEAD` 실행 또는 change_detector 캐시 활용
3. **Sync**: SPEC 범위 파일 목록 사용

---

## 6. Out of Scope

- **OS-001**: @MX 태그 자동 삽입/수정 (SPEC-MX-001 범위)
- **OS-002**: IDE/에디터 통합 (VS Code extension 등)
- **OS-003**: 실시간 LSP 기반 MX 진단
- **OS-004**: 크로스 리포지토리 MX 검증
- **OS-005**: Go 이외 언어에 대한 MX 검증 (향후 확장 가능)

---

## 7. Open Questions

- **OQ-001**: SessionEnd에서 수정된 파일 목록을 `git diff`로 가져올지, change_detector 캐시를 활용할지? **현재 결정**: `git diff --name-only HEAD`를 우선 사용하고, 실패 시 change_detector fallback.
- **OQ-002**: Sync Phase 0.6에서 P1/P2 위반 시 자동 태그 삽입을 제안할지? **현재 결정**: 검증만 수행하고 삽입은 제안하지 않음 (SPEC-MX-001의 agent 자율 태깅과 분리).

---

## 8. Traceability

| Requirement    | Acceptance Criteria | Plan Reference |
|----------------|---------------------|----------------|
| REQ-VAL-001    | AC-VAL-001          | TASK-1         |
| REQ-VAL-002    | AC-VAL-002          | TASK-1         |
| REQ-POST-001   | AC-POST-001         | TASK-2         |
| REQ-POST-002   | AC-POST-002         | TASK-2         |
| REQ-SESSION-001| AC-SESSION-001      | TASK-3         |
| REQ-SESSION-002| AC-SESSION-002      | TASK-3         |
| REQ-SESSION-003| AC-SESSION-003      | TASK-3         |
| REQ-SYNC-001   | AC-SYNC-001         | TASK-4         |
| REQ-SYNC-002   | AC-SYNC-002         | TASK-4         |
| REQ-SYNC-003   | AC-SYNC-003         | TASK-4         |
| REQ-REPORT-001 | AC-REPORT-001       | TASK-1         |
| REQ-REPORT-002 | AC-REPORT-002       | TASK-1         |
| REQ-CONFIG-001 | AC-CONFIG-001       | TASK-5         |
| REQ-CONFIG-002 | AC-CONFIG-002       | TASK-5         |
| REQ-EDGE-001   | AC-EDGE-001         | TASK-1         |
| REQ-EDGE-002   | AC-EDGE-002         | TASK-1         |
| REQ-EDGE-003   | AC-EDGE-003         | TASK-1         |
| REQ-EDGE-004   | AC-EDGE-004         | TASK-1         |
| REQ-EDGE-005   | AC-EDGE-005         | TASK-3         |
