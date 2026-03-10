---
id: SPEC-MX-002
version: 1.0.0
status: draft
created: 2026-03-11
updated: 2026-03-11
author: GOOS
priority: medium
---

# SPEC-MX-002 Implementation Plan

## Overview

MX Tag Auto-Validation System을 TDD (RED-GREEN-REFACTOR) 방식으로 구현한다. 6개의 태스크로 구분하며, 우선순위 기반 마일스톤으로 진행한다.

---

## Primary Goal: MX Validation Core + Configuration

### TASK-1: MX Validation Module (`internal/hook/mx/`)

**목적**: MX 검증의 핵심 로직을 독립 패키지로 구현한다.

**생성 파일**:
- `internal/hook/mx/types.go` -- 구조체 정의 (Violation, FileReport, ValidationReport, Config)
- `internal/hook/mx/config.go` -- mx.yaml의 `validation` 섹션 파싱
- `internal/hook/mx/validator.go` -- Validator 인터페이스 및 구현체

**참조 구현**:
- Fan-in 분석: SPEC-MX-001 REQ-SCAN-001의 Grep 기반 참조 카운팅
- AST 분석: `internal/astgrep/analyzer.go`의 Analyzer 인터페이스
- 변경 감지: `internal/hook/quality/change_detector.go`의 SHA-256 패턴

**핵심 기능**:
1. `ValidateFile(ctx, filePath) (*FileReport, error)` -- 단일 파일 MX 검증
   - exported 함수 추출 (AST-grep 또는 regex fallback)
   - fan_in 분석 (Grep 기반 참조 카운팅)
   - goroutine 패턴 감지 (`go func`, `go ` 패턴)
   - 기존 @MX 태그 파싱 (해당 함수에 태그 존재 여부)
   - P1-P4 위반 분류
2. `ValidateFiles(ctx, filePaths) (*ValidationReport, error)` -- 배치 검증
   - 파일별 병렬 검증 (errgroup 사용)
   - 타임아웃 지원 (context 기반)
   - 부분 결과 반환 (타임아웃 시)
3. `FormatReport(report, context) string` -- 표준 리포트 포맷 생성

**설계 원칙**:
- Validator는 읽기 전용 -- 태그를 수정하거나 삽입하지 않음
- 스레드 안전 (sync.RWMutex 또는 channel 기반)
- mx.yaml exclude 패턴 존중
- ast-grep 미설치 시 Grep fallback

### TASK-5: Configuration Update (`mx.yaml`)

**목적**: mx.yaml에 `validation` 섹션을 추가한다.

**수정 파일**:
- `.moai/config/sections/mx.yaml` -- validation 섹션 추가

**추가 내용**:
```yaml
mx:
  # ... 기존 설정 유지 ...
  validation:
    enabled: true
    post_tool_use:
      enabled: true
      timeout_ms: 500
    session_end:
      enabled: true
      timeout_ms: 4000
    sync:
      enforcement: strict
      skip_flag: "--skip-mx"
    enforcement_levels:
      p1_anchor: blocking
      p2_warn: blocking
      p3_note: advisory
      p4_todo: advisory
```

---

## Secondary Goal: Hook Integration

### TASK-2: PostToolUse MX Awareness

**목적**: PostToolUse 훅의 기존 AST-grep 스캔 이후에 MX 검증 결과를 메트릭에 추가한다.

**수정 파일**:
- `internal/hook/post_tool.go` -- `runMxValidation` 메서드 추가

**참조 구현**:
- PostToolUse AST 통합: `internal/hook/post_tool.go:84-91` (기존 `runAstScan` 패턴)

**구현 방법**:
1. `postToolHandler` 구조체에 `mxValidator` 필드 추가
2. `runAstScan` 호출 이후에 `runMxValidation` 호출
3. `runMxValidation`은 Write/Edit된 파일 경로를 추출하여 `ValidateFile` 호출
4. 결과를 `metrics["mx_validation"]`에 추가
5. 500ms 타임아웃 초과 시 `status: "skipped"` 반환
6. 관찰 전용 -- 절대 도구 실행을 차단하지 않음

**타임아웃 전략**:
- `context.WithTimeout(ctx, 500*time.Millisecond)` 사용
- 타임아웃 시 부분 결과 + `"skipped"` 상태 반환

### TASK-3: SessionEnd MX Batch Validation

**목적**: 세션 종료 시 수정된 모든 Go 파일에 대해 MX 배치 검증을 수행한다.

**수정 파일**:
- `internal/hook/session_end.go` -- `validateMxTags` 함수 추가

**구현 방법**:
1. `Handle` 함수에서 기존 정리 작업 이후 `validateMxTags` 호출
2. `git diff --name-only HEAD` 실행으로 수정된 파일 목록 획득
3. `.go` 파일만 필터링
4. mx.yaml `exclude` 패턴 적용
5. `ValidateFiles` 호출 (4초 타임아웃)
6. 결과를 `slog.Info`로 로그 출력
7. 검증 실패가 세션 종료를 차단하지 않음 (best-effort)

**수정 파일 감지**:
- 우선: `git diff --name-only HEAD` (staged + unstaged)
- Fallback: `git status --porcelain` 파싱

**성능 예산**:
- 최대 4초 (context.WithTimeout)
- 파일당 <500ms 목표
- 최대 약 8개 파일 검증 가능 (4초 / 500ms)
- 초과 시 부분 결과만 리포트

---

## Final Goal: Sync Enforcement + Workflow Update

### TASK-4: Sync Phase 0.6 Enforcement

**목적**: Sync 워크플로우의 Phase 0.6 MX 검증을 필수로 전환한다.

**수정 파일**:
- `.claude/skills/moai/workflows/sync.md` -- Phase 0.6 규칙 업데이트
- `internal/template/templates/.claude/skills/moai/workflows/sync.md` -- 템플릿 소스 동기화

**핵심 변경**:
1. Phase 0.6의 MX 검증을 `[HARD]` 규칙으로 변경
2. P1/P2 위반 시 sync 차단 로직 추가
3. `--skip-mx` 플래그 우회 메커니즘 문서화
4. P3/P4는 경고만 출력하고 sync 계속 진행

**Sync 차단 로직**:
```
Phase 0.6: MX Tag Validation
1. SPEC 범위 수정 파일 목록 획득
2. ValidateFiles 실행
3. IF P1 or P2 violations AND NOT --skip-mx:
   -> Block sync, show detailed report
   -> Suggest: "Run /moai run to add missing tags, or use --skip-mx to bypass"
4. IF P3 or P4 only:
   -> Show advisory warnings
   -> Continue sync
5. IF --skip-mx:
   -> Log skip in report
   -> Continue sync
```

### TASK-6: Tests (TDD: RED-GREEN-REFACTOR)

**목적**: 전체 구현에 대한 테스트를 TDD 방식으로 작성한다.

**생성 파일**:
- `internal/hook/mx/validator_test.go` -- Validator 유닛 테스트
- `internal/hook/mx/config_test.go` -- Config 파싱 테스트
- `internal/hook/mx/types_test.go` -- 타입 테스트 (필요시)

**테스트 시나리오**:
1. **ValidateFile**:
   - P1: fan_in >= 3 함수에 ANCHOR 누락 감지
   - P2: goroutine 패턴에 WARN 누락 감지
   - P3: 100줄 초과 exported 함수에 NOTE 누락 감지
   - P4: 테스트 없는 public 함수에 TODO 누락 감지
   - 기존 @MX 태그가 있는 함수는 위반 없음
   - exclude 패턴 매칭 파일은 건너뜀
   - 빈 파일 / Go가 아닌 파일 처리
2. **ValidateFiles**:
   - 여러 파일 배치 검증
   - 타임아웃 시 부분 결과 반환
   - 빈 파일 목록 처리
3. **Config**:
   - validation 섹션 파싱
   - 기본값 적용 (섹션 없을 때)
   - enforcement_levels 매핑
4. **FormatReport**:
   - PostToolUse 컨텍스트 (간략)
   - SessionEnd 컨텍스트 (요약)
   - Sync 컨텍스트 (상세)

**테스트 격리**:
- 모든 테스트에서 `t.TempDir()` 사용
- 실제 프로젝트 파일 수정 없음
- ast-grep 의존성: 인터페이스 모킹으로 격리

---

## Task Dependencies

```
TASK-1 (Validator Core) ─────────┐
                                  ├─> TASK-2 (PostToolUse)
TASK-5 (Config)  ────────────────┤
                                  ├─> TASK-3 (SessionEnd)
TASK-6 (Tests) ──── 전 단계 병렬 ─┤
                                  └─> TASK-4 (Sync Enforcement)
```

- TASK-1과 TASK-5는 독립적으로 병렬 진행 가능
- TASK-2, TASK-3은 TASK-1 완료 후 진행
- TASK-4는 TASK-1 완료 후 진행
- TASK-6은 각 TASK와 동시에 TDD로 진행 (RED 먼저)

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| ast-grep 미설치 환경 | MX 검증 불가 | Grep 기반 fallback 구현 (REQ-EDGE-001) |
| SessionEnd 타임아웃 초과 | 부분 검증 | 4초 context timeout + 부분 결과 반환 |
| PostToolUse 성능 저하 | 개발 체감 지연 | 500ms 타임아웃 + skip 메커니즘 |
| fan_in 분석 오탐 | 잘못된 P1 위반 | Grep 기반 근사치 허용 (SPEC-MX-001 A4와 동일) |
| 동시 파일 접근 | Race condition | sync.RWMutex 또는 errgroup 사용 |

---

## Architecture Design Direction

### Module Structure

```
internal/hook/mx/
  ├── types.go          # Violation, FileReport, ValidationReport, Priority
  ├── config.go         # ParseValidationConfig, default values
  ├── validator.go      # MXValidator struct implementing Validator interface
  ├── validator_test.go # Table-driven tests
  └── config_test.go    # Config parsing tests
```

### Key Design Decisions

1. **독립 패키지**: `internal/hook/mx/`로 분리하여 post_tool과 session_end 모두에서 import
2. **인터페이스 기반**: `Validator` 인터페이스로 테스트 모킹 용이
3. **DI (Dependency Injection)**: ast-grep Analyzer를 생성자에서 주입 (nil 허용 = Grep fallback)
4. **Context 기반 타임아웃**: 모든 public 메서드가 context.Context를 첫 번째 인자로 받음
5. **읽기 전용 보장**: Validator는 파일을 읽기만 하고, 수정/삽입하지 않음
