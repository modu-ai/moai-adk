---

name: moai-lang-go
description: Go best practices with go test, golint, gofmt, and standard library utilization. Use when writing or reviewing Go code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Go Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Go code discussions, framework guidance, or file extensions such as .go. |
| Tier | 3 |

## What it does

Provides Go-specific expertise for TDD development, including go test framework, golint/staticcheck, gofmt formatting, and effective standard library usage.

## When to use

- Engages when the conversation references Go work, frameworks, or files like .go.
- “Writing Go tests”, “How to use go tests”, “Go standard library”
- Automatically invoked when working with Go projects
- Go SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **go test**: Built-in testing framework
- **Table-driven tests**: Structured test cases
- **testify/assert**: Optional assertion library
- Test coverage ≥85% with `go test -cover`

**Code Quality**:
- **gofmt**: Automatic code formatting
- **golint**: Go linter (deprecated, use staticcheck)
- **staticcheck**: Advanced static analysis
- **go vet**: Built-in error detection

**Standard Library**:
- Use standard library first before external dependencies
- **net/http**: HTTP server/client
- **encoding/json**: JSON marshaling
- **context**: Context propagation

**Go Patterns**:
- Interfaces for abstraction (small interfaces)
- Error handling with explicit returns
- Defer for cleanup
- Goroutines and channels for concurrency

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Exported names start with capital letters
- Error handling: `if err != nil { return err }`
- Avoid naked returns in large functions

## Examples
```bash
go test ./... && golangci-lint run
```

## Inputs
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- The Go Authors. "Effective Go." https://go.dev/doc/effective_go (accessed 2025-03-29).
- GolangCI. "golangci-lint Documentation." https://golangci-lint.run/usage/quick-start/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Go-specific review)
- alfred-performance-optimizer (Go profiling)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
