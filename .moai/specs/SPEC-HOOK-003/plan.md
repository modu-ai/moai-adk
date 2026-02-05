---
spec_id: SPEC-HOOK-003
title: Security & Scanning - Implementation Plan
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, implementation-plan, security, ast-grep, scanning"
---

# SPEC-HOOK-003: Implementation Plan

## 1. Overview

### 1.1 Scope

Python 기반 AST-Grep 보안 스캔 훅(~284 LOC)을 Go로 포팅하여 실시간 보안 취약점 탐지 기능을 제공한다. `internal/hook/security/` 패키지에 4개 파일을 구현하며, Phase 2 Quality Integration의 P1 High 모듈이다.

### 1.2 Implementation Strategy

**Bottom-Up 접근**: AST-Grep 도구 통합 -> 규칙 관리 -> 스캐너 -> 보고서 생성 순서로 구현한다.

### 1.3 Dependencies

| Dependency       | Status  | Blocking | Impact                                     |
|------------------|---------|----------|--------------------------------------------|
| SPEC-HOOK-001    | Completed | Yes     | Registry, Protocol, Contract 재사용        |
| ast-grep (sg)    | External | No       | 스캔 엔진 (선택 사항, graceful degradation)  |
| Go 1.22+         | Available | No     | encoding/json, os/exec, context            |

---

## 2. Task Decomposition

### Milestone 1: Core Infrastructure (Primary Goal)

핵심 인프라 2개 파일 구현. 도구 통합과 규칙 관리.

#### Task 1.1: AST-Grep Integration (`ast_grep.go`)

**Priority**: High

**Description**: ast-grep 바이너리 통합 및 언어 지원.

**Implementation Details**:
- `IsAvailable()`: `exec.LookPath("sg")`로 가용성 확인
- `GetVersion()`: `sg --version` 실행
- 지원 언어 맵(14개 언어)
- 지원 파일 확장자

**Testing**:
- 도구 설치/미설치 상황 테스트
- 버전 확인 테스트
- 언어별 확장자 매칭 테스트

**Covered Requirements**: REQ-HOOK-100, REQ-HOOK-140, REQ-HOOK-141

#### Task 1.2: Rule Manager (`rules.go`)

**Priority**: High

**Description**: 보안 규칙 파일 탐지 및 로드.

**Implementation Details**:
- `FindRulesConfig()`: 규칙 파일 경로 탐지
  1. `.claude/skills/moai-tool-ast-grep/rules/sgconfig.yml`
  2. `sgconfig.yml`
  3. `.ast-grep/sgconfig.yml`
- `LoadRules()`: YAML/JSON 규칙 파일 파싱
- `GetDefaultRules()`: 내장 OWASP Top 10 규칙 반환

**Testing**:
- 규칙 파일 발견 테스트
- 규칙 로드 성공/실패 테스트
- 기본 규칙 반환 테스트

**Covered Requirements**: REQ-HOOK-110, REQ-HOOK-111, REQ-HOOK-112

---

### Milestone 2: Scanner and Reporter (Secondary Goal)

2개 핵심 기능 구현. 스캐너와 보고서 생성.

#### Task 2.1: Scanner (`scanner.go`)

**Priority**: High

**Description**: AST-Grep 스캔 실행 및 결과 파싱.

**Implementation Details**:
- `Scan()`: 단일 파일 스캔
  - `sg scan --json --config <config> <file>` 실행
  - JSON 출력 파싱
  - fallback 정규식 파싱
- `ScanMultiple()`: 다중 파일 병렬 스캔
- 타임아웃: 30초

**Testing**:
- 정상 스캔 테스트
- 취약점 발견 테스트
- 타임아웃 처리 테스트
- JSON 파싱 실패 시 fallback 테스트

**Covered Requirements**: REQ-HOOK-120, REQ-HOOK-121, REQ-HOOK-122, REQ-HOOK-123

#### Task 2.2: Finding Reporter (`reporter.go`)

**Priority**: Medium

**Description**: 스캔 결과를 Claude 피드백 형식으로 변환.

**Implementation Details**:
- `FormatResult()`: 단일 결과 포맷팅
- `FormatMultiple()`: 다중 결과 집계
- 심각도별 정렬 (error > warning > info)
- 최대 10개 결과 표시

**Testing**:
- 결과 포맷팅 테스트
- 심각도 정렬 테스트
- 10개 초과 결과 요약 테스트

**Covered Requirements**: REQ-HOOK-130, REQ-HOOK-131, REQ-HOOK-132

---

### Milestone 3: Integration (Final Goal)

PostToolUse 훅 통합.

#### Task 3.1: PostToolUse Security Handler

**Priority**: High

**Description**: PostToolUse 훅에서 스캐너 호출.

**Implementation Details**:
- `internal/hook/post_tool.go`에 스캐너 통합
- Write/Edit 작업 후 자동 스캔
- 환경 변수 `MOAI_DISABLE_AST_GREP_SCAN`로 비활성화 지원

---

## 3. Technology Specifications

### 3.1 Language and Runtime

| Component     | Specification          |
|---------------|------------------------|
| Language      | Go 1.22+               |
| Module        | `github.com/modu-ai/moai-adk-go` |
| Package       | `internal/hook/security` |
| Build         | `CGO_ENABLED=0`        |

### 3.2 Standard Library Dependencies

| Package           | Purpose                                |
|-------------------|----------------------------------------|
| `context`         | Cancellation, timeouts                  |
| `encoding/json`   | Scan result parsing                     |
| `os/exec`         | AST-Grep execution                     |
| `path/filepath`   | Path manipulation                       |
| `regexp`          | Fallback parsing                        |
| `time`            | Timeout duration                        |

### 3.3 External Dependencies

| Package       | Purpose                        |
|---------------|--------------------------------|
| ast-grep (sg) | Structural code search (optional) |

---

## 4. Risk Analysis

### 4.1 Technical Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **ast-grep 미설치**                     | High     | Medium     | Graceful degradation, skip 스캔                           |
| **JSON 출력 형식 변경**                  | Medium   | Low        | Fallback 정규식 파싱                                       |
| **규칙 파일 호환성**                     | Medium   | Medium     | 기본 OWASP 규칙 내장                                       |
| **대형 파일 타임아웃**                   | Medium   | Medium     | 타임아웃 설정, 파일 크기 제한                               |

### 4.2 Process Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **OWASP 규칙 완전성**                    | Medium   | Medium     | Python 훅 규칙 복사 + 검토                                  |

---

## 5. Migration Plan (Python -> Go)

### 5.1 Migration Strategy

**단계적 교체(Phased Replacement)**: settings.json의 hook command를 Python 스크립트에서 Go 서브커맨드로 변경한다.

### 5.2 Migration Steps

| Step | Action                                    | Verification                              |
|------|-------------------------------------------|-------------------------------------------|
| 1    | Go 보안 스캔 모듈 구현 및 단위 테스트 통과   | `go test ./internal/hook/security/...`    |
| 2    | PostToolUse 통합 테스트                     | 스캔 동작 확인                             |
| 3    | settings.json 생성기에서 hook command 변경   | `moai hook security-scan` 수동 실행 확인    |
| 4    | Python 훅 제거                              | `.claude/hooks/` 정리                      |

---

## 6. Architecture Design Direction

### 6.1 Package Structure

```
internal/hook/security/
    ast_grep.go              # AST-Grep tool integration
    ast_grep_test.go
    scanner.go               # Scan execution and parsing
    scanner_test.go
    rules.go                 # Rule configuration management
    rules_test.go
    reporter.go              # Finding reporting
    reporter_test.go
    types.go                 # Shared types
    owasp_rules.go           # Built-in OWASP rules
```

### 6.2 Dependency Flow

```
internal/cli/hook.go (PostToolUse)
    |
    v
internal/hook/security/scanner.go -- uses --> ast_grep.go
    |                                      |
    v                                      v
reporter.go                        exec.Command (sg scan)
                                    |
                                    v
                                rules.go (sgconfig.yml)
```

---

## 7. Quality Criteria

### 7.1 Coverage Target

| Scope                    | Target | Rationale                              |
|--------------------------|--------|----------------------------------------|
| `internal/hook/security/` 전체 | 85%    | 보안 모듈, 신뢰성 중요                   |
| Scanner                  | 90%    | 스캔 로직 완전 검증                      |
| Reporter                 | 85%    | 보고서 포맷팅 검증                      |

### 7.2 TRUST 5 Compliance

| Principle   | Security Module Application                                     |
|-------------|---------------------------------------------------------------|
| Tested      | 85%+ coverage, table-driven tests                               |
| Readable    | Go naming conventions, godoc comments                          |
| Unified     | gofumpt formatting, golangci-lint compliance                    |
| Secured     | Path validation, input sanitization                            |
| Trackable   | Conventional commits, SPEC-HOOK-003 reference in all commits  |
