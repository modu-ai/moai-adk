---
id: SPEC-HOOK-002
title: Code Quality Automation
version: "0.1.0"
status: Draft
created: 2026-02-04
updated: 2026-02-04
author: GOOS
priority: P1 High
phase: "Phase 2 - Quality Integration"
module: "internal/hook/quality/"
dependencies:
  - SPEC-HOOK-001
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: []
lifecycle: spec-anchored
tags: "hook, formatter, linter, tool-registry, code-quality, P1"
---

# SPEC-HOOK-002: Code Quality Automation

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 0.1.0   | 2026-02-04 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 Go 언어로 작성된 AI 개발 키트로, Claude Code와 통합하여 자동화된 코드 품질 관리 기능을 제공한다. 이 SPEC은 PostToolUse 훅 후크에서 자동으로 코드 포맷팅 및 린팅을 수행하는 Code Quality Automation 시스템을 정의한다.

### 1.2 Problem Statement

Python 기반 MoAI-ADK의 code formatter와 linter 훅은 16개 이상의 프로그래밍 언어를 지원하지만, 다음과 같은 문제가 있다:

- **도구 탐지 오버헤드**: 파일마다 도구 가용성을 확인하는 비용이 높음
- **중복 실행**: 동일 파일에 대해 여러 훅이 순차적으로 실행되어 성능 저하
- **부분적인 언어 지원**: 일부 언어의 도구 설정이 누락되거나 불완전
- **변경 감지 부재**: 파일이 실제로 변경되었는지 확인하지 않고 항상 실행

### 1.3 Target Module

- **경로**: `internal/hook/quality/`
- **파일 구성**: `formatter.go`, `linter.go`, `tool_registry.go`, `change_detector.go`
- **예상 LOC**: ~1,800

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-HOOK-001    | Internal | Compiled Hook System                          |
| SPEC-CONFIG-001  | Internal | Configuration Manager                         |
| Claude Code      | External | PostToolUse hook event                        |
| Go 1.22+         | Runtime  | context, encoding/json, os/exec, crypto/sha256 |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 `moai hook post-tool` 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | Claude Code는 PostToolUse 이벤트에 file_path를 포함하여 전송한다 | High       | 파일 경로 추적 불가로 훅 동작 불가       |
| A-002 | 도구 실행 시 타임아웃이 필요하다 (기본 30초)                     | High       | 무한 대기로 세션 중단 가능               |
| A-003 | 파일 해시(SHA-256) 기반 변경 감지가 신뢰할 수 있다              | High       | 오탐지/누탐지로 불필요한 실행 발생        |
| A-004 | SPEC-HOOK-001의 Registry와 Protocol을 재사용할 수 있다          | High       | 중복 구현으로 일관성 상실                |
| A-005 | 도구 바이너리는 시스템 PATH에 존재한다                         | Medium     | 도구 미발견으로 graceful degradation 필요 |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 사용자는 자동 포맷팅/린팅을 기대한다                            | High       | 수동 설정 부담 증가                     |
| B-002 | 16개 언어 지원이 충분하다                                      | Medium     | 추가 언어 요청 발생 가능                 |

---

## 3. Requirements (R)

### Module 1: Tool Registry (도구 등록 및 검색)

**REQ-HOOK-050** [Ubiquitous]
시스템은 **항상** 타입 안전한 도구 레지스트리를 통해 포매터와 린터를 관리해야 한다.

- `RegisterTool(tool ToolConfig)` 메서드로 도구 등록
- `GetToolsForLanguage(language string, toolType ToolType)`으로 언어별 도구 조회
- `GetToolsForFile(filePath string, toolType ToolType)`으로 파일별 도구 조회
- `IsToolAvailable(toolName string)`으로 도구 가용성 확인

**REQ-HOOK-051** [Event-Driven]
**WHEN** 레지스트리가 초기화되면 **THEN** 시스템은 16개 언어(Python, JavaScript, TypeScript, Go, Rust, Java, Kotlin, Swift, C/C++, Ruby, PHP, Elixir, Scala, R, Dart, C#, Markdown)에 대한 기본 도구를 자동 등록해야 한다.

**REQ-HOOK-052** [State-Driven]
**IF** 언어에 대해 다중 도구가 등록되어 있으면 **THEN** 레지스트리는 priority 필드를 기준으로 정렬하여 우선순위가 가장 높은 도구를 먼저 반환해야 한다.

**REQ-HOOK-053** [Unwanted]
시스템은 도구 실행을 위해 문자열 연결로 명령어를 생성**하지 않아야 한다**. 모든 명령어는 `exec.Command`와 함께 슬라이스 형태로 전달되어야 하여 쉘 인젝션을 방지한다.

### Module 2: Change Detection (파일 변경 감지)

**REQ-HOOK-060** [Ubiquitous]
시스템은 **항상** 포맷팅/린팅 전후 파일 해시를 비교하여 실제 변경 여부를 감지해야 한다.

- `ComputeFileHash(filePath string) ([]byte, error)`: SHA-256 해시 계산
- `HasFileChanged(filePath string, previousHash []byte) (bool, error)`: 변경 비교
- 해시 캐시로 중복 계산 방지

**REQ-HOOK-061** [Event-Driven]
**WHEN** 파일이 Write/Edit 도구로 수정되면 **THEN** 시스템은 수정 전 해시를 저장하고, 포맷팅 후 해시와 비교하여 변경 여부를 결정해야 한다.

**REQ-HOOK-062** [State-Driven]
**IF** 파일 해시가 변경되지 않았으면 **THEN** 시스템은 "no changes" 메시지를 출력하고 추가 컨텍스트를 제공하지 않아야 한다(suppressOutput: true).

### Module 3: Code Formatter (자동 포맷팅)

**REQ-HOOK-070** [Ubiquitous]
시스템은 **항상** PostToolUse 이벤트에서 Write/Edit 작업 후 자동으로 코드 포맷팅을 시도해야 한다.

**REQ-HOOK-071** [Event-Driven]
**WHEN** 지원되는 파일 확장자를 가진 파일이 작성/수정되면 **THEN** 시스템은 해당 언어의 포매터를 찾아 실행해야 한다.

**REQ-HOOK-072** [State-Driven]
**IF** 포매터가 파일을 실제로 수정했으면 **THEN** 시스템은 "Auto-formatted with {tool}" 메시지를 Claude에게 제공해야 한다.

**REQ-HOOK-073** [Unwanted]
시스템은 다음 확장자를 가진 파일을 포맷**하지 않아야 한다**: `.json`, `.lock`, `.min.js`, `.min.css`, `.map`, `.svg`, `.png`, `.jpg`, `.gif`, `.ico`, `.woff`, `.woff2`, `.ttf`, `.eot`.

**REQ-HOOK-074** [Unwanted]
시스템은 다음 디렉터리 내의 파일을 포맷**하지 않아야 한다**: `node_modules`, `.git`, `.venv`, `venv`, `__pycache__`, `.cache`, `dist`, `build`, `.next`, `.nuxt`, `target`, `vendor`.

### Module 4: Linter Integration (자동 린팅)

**REQ-HOOK-080** [Ubiquitous]
시스템은 **항상** 포맷팅 후에 린터를 실행하여 코드 품질 이슈를 감지해야 한다.

**REQ-HOOK-081** [Event-Driven]
**WHEN** 린터가 이슈를 발견하면 **THEN** 시스템은 이슈 요약(최대 5개)을 Claude에게 제공해야 한다.

**REQ-HOOK-082** [State-Driven]
**IF** 린터가 --fix 옵션을 지원하면 **THEN** 시스템은 자동 수정을 시도하고 수정된 이슈 수를 보고해야 한다.

**REQ-HOOK-083** [Optional]
**가능하면** 린터 실행 실패 시 graceful degradation으로 인해 포맷팅 결과는 유지하고 린팅 실패만 경고로 보고한다.

### Module 5: Cross-Platform Compatibility

**REQ-HOOK-090** [Ubiquitous]
시스템은 **항상** Windows, macOS, Linux 플랫폼에서 도구를 실행할 수 있어야 한다.

**REQ-HOOK-091** [Event-Driven]
**WHEN** Windows 플랫폼에서 실행 중이면 **THEN** 시스템은 SIGALRM 미지원 문제를 회피하기 위해 `context.WithTimeout`만 사용해야 한다.

**REQ-HOOK-092** [State-Driven]
**IF** 도구 실행이 타임아웃되면 **THEN** 시스템은 타임아웃 오류를 로깅하고 graceful하게 다음 도구로 넘어가거나 결과를 반환해야 한다.

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
// ToolType represents the type of code quality tool.
type ToolType string

const (
    ToolTypeFormatter     ToolType = "formatter"
    ToolTypeLinter         ToolType = "linter"
    ToolTypeTypeChecker    ToolType = "type_checker"
)

// ToolConfig represents configuration for a single tool.
type ToolConfig struct {
    Name             string
    Command          string
    Args             []string
    Extensions       []string
    ToolType         ToolType
    Priority         int           // Lower = higher priority
    TimeoutSeconds   int           // Default: 30
    CheckArgs        []string      // Args to check if tool exists
    FixArgs          []string      // Args to auto-fix issues
}

// ToolResult represents the result of tool execution.
type ToolResult struct {
    Success       bool
    ToolName      string
    Output        string
    Error         string
    ExitCode      int
    FileModified  bool
    IssuesFound   int
    IssuesFixed   int
    ExecutionTime time.Duration
}

// ToolRegistry manages code quality tool registration and execution.
type ToolRegistry interface {
    RegisterTool(tool ToolConfig)
    GetToolsForLanguage(language string, toolType ToolType) []ToolConfig
    GetToolsForFile(filePath string, toolType ToolType) []ToolConfig
    IsToolAvailable(toolName string) bool
    RunTool(tool ToolConfig, filePath string, cwd string) ToolResult
}

// ChangeDetector manages file hash-based change detection.
type ChangeDetector interface {
    ComputeHash(filePath string) ([]byte, error)
    HasChanged(filePath string, previousHash []byte) (bool, error)
    GetCachedHash(filePath string) ([]byte, bool)
    CacheHash(filePath string, hash []byte)
}

// Formatter handles automatic code formatting.
type Formatter interface {
    FormatFile(ctx context.Context, filePath string) (*ToolResult, error)
    ShouldFormat(filePath string) bool
}

// Linter handles automatic code linting.
type Linter interface {
    LintFile(ctx context.Context, filePath string) (*ToolResult, error)
    AutoFix(ctx context.Context, filePath string) (*ToolResult, error)
}
```

### 4.2 Language Support Matrix

| Language    | Extensions              | Formatters              | Linters                      |
|-------------|-------------------------|-------------------------|------------------------------|
| Python      | .py, .pyi               | ruff-format, black      | ruff, mypy                    |
| JavaScript  | .js, .jsx, .mjs, .cjs   | biome, prettier         | eslint, biome                 |
| TypeScript  | .ts, .tsx, .mts, .cts   | biome, prettier         | eslint, biome, tsc            |
| Go          | .go                     | gofmt, goimports        | golangci-lint                 |
| Rust        | .rs                     | rustfmt                 | clippy                        |
| Java        | .java                   | google-java-format      | checkstyle                    |
| Kotlin      | .kt, .kts               | ktlint                  | detekt                        |
| Swift       | .swift                  | swift-format            | swiftlint                     |
| C/C++       | .c, .cpp, .cc, .h, .hpp  | clang-format            | clang-tidy                    |
| Ruby        | .rb, .rake, .gemspec    | rubocop                 | rubocop                       |
| PHP         | .php                    | php-cs-fixer            | phpstan                       |
| Elixir      | .ex, .exs               | mix format              | credo                         |
| Scala       | .scala, .sc             | scalafmt                | scalafix                      |
| R           | .r, .R, .Rmd            | styler                  | lintr                         |
| Dart        | .dart                   | dart format             | dart analyze                  |
| C#          | .cs                     | dotnet format           | (none)                        |
| Markdown    | .md, .mdx               | prettier                | markdownlint                  |

### 4.3 Skip Patterns

**File Extensions to Skip**:
- Config files: `.json`, `.yaml`, `.yml`, `.toml`, `.lock`
- Minified files: `.min.js`, `.min.css`
- Source maps: `.map`
- Images: `.svg`, `.png`, `.jpg`, `.gif`, `.ico`, `.webp`
- Fonts: `.woff`, `.woff2`, `.ttf`, `.eot`, `.otf`
- Binaries: `.exe`, `.dll`, `.so`, `.dylib`, `.bin`

**Directories to Skip**:
- Dependencies: `node_modules`, `vendor`, `.venv`, `venv`
- Build outputs: `dist`, `build`, `target`, `.next`, `.nuxt`, `out`
- VCS: `.git`, `.svn`, `.hg`
- Cache: `__pycache__`, `.cache`, `.pytest_cache`
- IDE: `.idea`, `.vscode`, `.eclipse`

### 4.4 Data Structures

```go
// Supported languages mapping
var LanguageExtensions = map[string]string{
    ".py":    "python",
    ".pyi":   "python",
    ".go":    "go",
    ".rs":    "rust",
    ".js":    "javascript",
    ".ts":    "typescript",
    ".tsx":   "typescript",
    ".java":  "java",
    // ... etc
}

// Default tool configurations
var DefaultTools = map[string][]ToolConfig{
    "python": {
        {
            Name:           "ruff-format",
            Command:        "ruff",
            Args:           []string{"format"},
            Extensions:     []string{".py", ".pyi"},
            ToolType:       ToolTypeFormatter,
            Priority:       1,
            TimeoutSeconds: 30,
        },
        // ... more tools
    },
    // ... more languages
}
```

### 4.5 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| 단일 파일 포맷팅              | < 2s      | Benchmark test                         |
| 단일 파일 린팅                | < 5s      | Benchmark test                         |
| 도구 가용성 확인              | < 10ms    | Benchmark test (shutil.which)          |
| 파일 해시 계산                | < 5ms     | Benchmark test (SHA-256 1MB file)      |
| 도구 레지스트리 초기화        | < 50ms    | Benchmark test (16 languages)          |
| 메모리 사용량 (실행 중)       | < 50MB    | Runtime profiling                      |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-050~053 | `tool_registry.go`             |
| REQ-HOOK-060~062 | `change_detector.go`           |
| REQ-HOOK-070~074 | `formatter.go`                 |
| REQ-HOOK-080~083 | `linter.go`                    |
| REQ-HOOK-090~092 | All files (cross-platform)     |

### 5.2 Python Hook Mapping

| Python Script                   | Go Handler         | Status  |
|---------------------------------|--------------------|---------|
| `post_tool__code_formatter.py`  | `formatter.go`     | Planned |
| `post_tool__linter.py`          | `linter.go`        | Planned |
| `lib/tool_registry.py`          | `tool_registry.go` | Planned |

### 5.3 Integration Points

- **SPEC-HOOK-001**: Registry, Protocol, Contract 재사용
- **internal/cli/hook.go**: PostToolUse 서브커맨드에서 Formatter/Linter 호출
- **internal/config/**: 도구 설정, 타임아웃 값 로드

---

## Implementation Notes

**Status**: Draft
**Phase**: Phase 2 - Quality Integration

### Summary

Code Quality Automation system for automatic code formatting and linting after Write/Edit operations. Supports 16+ programming languages with automatic tool detection, file hash-based change detection, and cross-platform compatibility. Integrates with PostToolUse hook to provide real-time feedback to Claude Code.

### Python Reference

- `post_tool__code_formatter.py` (263 LOC)
- `post_tool__linter.py` (~400 LOC)
- `lib/tool_registry.py` (897 LOC)

### Estimated LOC

- `tool_registry.go`: ~500 LOC
- `formatter.go`: ~400 LOC
- `linter.go`: ~400 LOC
- `change_detector.go`: ~200 LOC
- `tool_registry_test.go`: ~400 LOC
- Total: ~1,900 LOC
