---
spec_id: SPEC-HOOK-002
title: Code Quality Automation - Implementation Plan
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, implementation-plan, formatter, linter, tool-registry"
---

# SPEC-HOOK-002: Implementation Plan

## 1. Overview

### 1.1 Scope

Python 기반 Code Quality Automation 훅(~1,560 LOC)을 Go로 포팅하여 16개 언어에 대한 자동 포맷팅 및 린팅 기능을 제공한다. `internal/hook/quality/` 패키지에 4개 파일을 구현하며, Phase 2 Quality Integration의 P1 High 모듈이다.

### 1.2 Implementation Strategy

**Bottom-Up 접근**: Tool Registry -> Change Detector -> Formatter -> Linter 순서로 구현한다. 각 단계는 독립적으로 테스트 가능하며, 이전 단계의 인터페이스에 의존한다.

### 1.3 Dependencies

| Dependency       | Status  | Blocking | Impact                                     |
|------------------|---------|----------|--------------------------------------------|
| SPEC-HOOK-001    | Completed | Yes     | Registry, Protocol, Contract 재사용        |
| SPEC-CONFIG-001  | Planned | Yes      | 설정 로드, 타임아웃 값 조회                 |
| Go 1.22+         | Available | No     | crypto/sha256, context, os/exec            |

---

## 2. Task Decomposition

### Milestone 1: Core Infrastructure (Primary Goal)

핵심 인프라 2개 파일 구현. 모든 품질 도구의 기반이 되는 레지스트리와 변경 감지 시스템.

#### Task 1.1: Tool Registry (`tool_registry.go`)

**Priority**: High

**Description**: 16개 언어에 대한 포매터와 린터 도구 레지스트리 구현.

**Implementation Details**:
- `ToolConfig` 구조체: 도구 명령어, 인자, 확장자, 우선순위
- `ToolRegistry` 인터페이스 및 구현
- `_registerDefaultTools()`: 16개 언어 도구 자동 등록
- `IsToolAvailable()`: `exec.LookPath`로 도구 가용성 확인
- `RunTool()`: `exec.Command`로 도구 실행, 파일 해시 기반 변경 감지
- 도구 캐시로 가용성 확인 결과 재사용

**Testing**:
- 도구 등록/조회 단위 테스트
- 언어별 도구 조회 테스트
- 우선순위 정렬 테스트
- 도구 실행 성공/실패 테스트
- 병렬 안전성(`t.Parallel()`) 테스트

**Covered Requirements**: REQ-HOOK-050, REQ-HOOK-051, REQ-HOOK-052, REQ-HOOK-053

#### Task 1.2: Change Detector (`change_detector.go`)

**Priority**: High

**Description**: 파일 해시 기반 변경 감지 시스템 구현.

**Implementation Details**:
- `ComputeHash()`: SHA-256 해시 계산
- `HasChanged()`: 이전 해시와 비교
- 해시 캐시(map[string][]byte)로 중복 계산 방지
- 캐시 만료 정책(TTL: 5분)

**Testing**:
- 해시 계산 정확성 테스트
- 캐시 히트/미스 테스트
- 빈 파일 처리 테스트
- 대용량 파일(10MB+) 처리 테스트

**Covered Requirements**: REQ-HOOK-060, REQ-HOOK-061, REQ-HOOK-062

---

### Milestone 2: Quality Tools (Secondary Goal)

2개 품질 도구 구현. 포매터와 린터 핸들러.

#### Task 2.1: Formatter (`formatter.go`)

**Priority**: High

**Description**: 자동 코드 포맷팅 핸들러 구현.

**Implementation Details**:
- `FormatFile()`: ToolRegistry를 통해 포매터 실행
- `ShouldFormat()`: 파일 확장자, 디렉터리, 바이너리 체크
- SKIP_EXTENSIONS, SKIP_DIRECTORIES 상수 정의
- 포맷팅 후 Claude 피드백(`additionalContext`) 생성

**Testing**:
- 지원 언어 포맷팅 테스트
- 건너뛰기 파일/디렉터리 테스트
- 포매터 없을 때 graceful degradation 테스트
- 타임아웃 처리 테스트

**Covered Requirements**: REQ-HOOK-070, REQ-HOOK-071, REQ-HOOK-072, REQ-HOOK-073, REQ-HOOK-074

#### Task 2.2: Linter (`linter.go`)

**Priority**: Medium

**Description**: 자동 코드 린팅 핸들러 구현.

**Implementation Details**:
- `LintFile()`: ToolRegistry를 통해 린터 실행
- `AutoFix()`: --fix 옵션 지원 시 자동 수정
- 이슈 요약 (최대 5개, severity별 정렬)
- JSON 출력 파싱 (ruff, eslint 등)

**Testing**:
- 지원 언어 린팅 테스트
- 자동 수정 기능 테스트
- JSON 파싱 테스트
- 이슈 요약 포맷 테스트

**Covered Requirements**: REQ-HOOK-080, REQ-HOOK-081, REQ-HOOK-082, REQ-HOOK-083

---

### Milestone 3: Integration and Cross-Platform (Final Goal)

CLI 통합, 크로스 플랫폼 검증.

#### Task 3.1: PostToolUse Quality Handler

**Priority**: High

**Description**: PostToolUse 훅에서 Formatter와 Linter 호출.

**Implementation Details**:
- `internal/hook/post_tool.go`에 품질 검사 통합
- 포맷팅 -> 린팅 순차 실행
- 결과 집계 및 Claude 피드백

#### Task 3.2: Cross-Platform Tests

**Priority**: High

**Description**: Windows, macOS, Linux 검증.

**Test Cases**:
- Windows: PATH 검색, cmd.exe 실행
- macOS: Homebrew 도구 탐지
- Linux: 시스템 도구 탐지
- 타임아웃 처리 (Windows: context, Unix: signal)

---

## 3. Technology Specifications

### 3.1 Language and Runtime

| Component     | Specification          |
|---------------|------------------------|
| Language      | Go 1.22+               |
| Module        | `github.com/modu-ai/moai-adk-go` |
| Package       | `internal/hook/quality` |
| Build         | `CGO_ENABLED=0`        |

### 3.2 Standard Library Dependencies

| Package           | Purpose                                |
|-------------------|----------------------------------------|
| `context`         | Cancellation, timeouts                  |
| `crypto/sha256`   | File hash computation                   |
| `encoding/json`   | Linter output parsing                   |
| `os/exec`         | Tool execution                         |
| `path/filepath`   | Path manipulation                       |
| `sync`            | Thread-safe cache                      |
| `time`            | Timeout duration, TTL                   |

### 3.3 Internal Dependencies

| Package            | Interface Used        | Purpose                        |
|--------------------|----------------------|--------------------------------|
| `internal/hook`    | `Registry`, `Handler` | Hook system integration        |
| `internal/config`  | `ConfigManager`      | Configuration loading          |

---

## 4. Risk Analysis

### 4.1 Technical Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **도구 실행 타임아웃**                    | High     | Medium     | context.WithTimeout + graceful degradation                 |
| **파일 해시 충돌**                      | Low      | Low        | SHA-256 사용 (충돌 확률 무시할 수 있음)                      |
| **JSON 파싱 실패**                      | Medium   | Medium     | fallback 처리, 정규식 기반 파싱                             |
| **Windows 경로 구분자**                  | Medium   | High       | filepath.Clean(), filepath.Join() 일관 사용                 |
| **도구 설치 경로 다양성**                | Medium   | High       | exec.LookPath으로 시스템 PATH 검색                          |

### 4.2 Process Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **16개 언어 도구 설정 복잡성**           | Medium   | High       | Python 훅 설정 복사 + 테스트 커버리지                        |
| **린터 JSON 형식 호환성**                 | Medium   | Medium     | 주요 린터(ruff, eslint)만 우선 지원                         |

---

## 5. Migration Plan (Python -> Go)

### 5.1 Migration Strategy

**단계적 교체(Phased Replacement)**: settings.json의 hook command를 Python 스크립트에서 Go 서브커맨드로 변경한다.

### 5.2 Migration Steps

| Step | Action                                    | Verification                              |
|------|-------------------------------------------|-------------------------------------------|
| 1    | Go 품질 도구 모듈 구현 및 단위 테스트 통과   | `go test ./internal/hook/quality/...`      |
| 2    | PostToolUse 통합 테스트                     | Formatter/Linter 동작 확인                 |
| 3    | settings.json 생성기에서 hook command 변경   | `moai hook post-tool` 수동 실행 확인       |
| 4    | Python 훅 제거                              | `.claude/hooks/` 정리                      |

---

## 6. Architecture Design Direction

### 6.1 Package Structure

```
internal/hook/quality/
    tool_registry.go          # Tool registration and execution
    tool_registry_test.go
    formatter.go              # Code formatting handler
    formatter_test.go
    linter.go                 # Linting handler
    linter_test.go
    change_detector.go        # File hash-based change detection
    change_detector_test.go
    types.go                  # Shared types and constants
    skip_patterns.go          # Skip patterns (extensions, directories)
```

### 6.2 Dependency Flow

```
internal/cli/hook.go (PostToolUse)
    |
    v
internal/hook/quality/formatter.go -- uses --> tool_registry.go
    |                                         |
    v                                         v
change_detector.go                    exec.Command (tool execution)
```

### 6.3 Constructor Pattern

```go
// NewToolRegistry creates a registry with default tools.
func NewToolRegistry() *ToolRegistry {
    r := &ToolRegistry{
        tools:   make(map[string][]ToolConfig),
        cache:   make(map[string]bool),
        hashCache: make(map[string][]byte),
    }
    r._registerDefaultTools()
    return r
}

// NewFormatter creates a formatter with registry.
func NewFormatter(registry *ToolRegistry) *Formatter {
    return &Formatter{
        registry: registry,
        detector:  NewChangeDetector(),
    }
}
```

---

## 7. Quality Criteria

### 7.1 Coverage Target

| Scope                    | Target | Rationale                              |
|--------------------------|--------|----------------------------------------|
| `internal/hook/quality/` 전체 | 90%    | 품질 도구, 신뢰성 중요                   |
| Tool Registry            | 95%    | 도구 실행 로직 완전 검증                  |
| Formatter                | 90%    | 핵심 경로 + 오류 경로 검증               |
| Linter                   | 85%    | 핵심 경로 + 일부 도구별 테스트            |

### 7.2 TRUST 5 Compliance

| Principle   | Quality Module Application                                     |
|-------------|---------------------------------------------------------------|
| Tested      | 90%+ coverage, table-driven tests, cross-platform matrix      |
| Readable    | Go naming conventions, godoc comments                         |
| Unified     | gofumpt formatting, golangci-lint compliance                   |
| Secured     | Path validation, shell injection prevention (exec.Command)     |
| Trackable   | Conventional commits, SPEC-HOOK-002 reference in all commits  |
