---

name: moai-lang-csharp
description: C# best practices with xUnit, .NET tooling, LINQ, and async/await patterns. Use when writing or reviewing C# code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# C# Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | C# code discussions, framework guidance, or file extensions such as .cs. |
| Tier | 3 |

## What it does

Provides C#-specific expertise for TDD development, including xUnit testing, .NET CLI tooling, LINQ query expressions, and async/await patterns.

## When to use

- Engages when the conversation references C# work, frameworks, or files like .cs.
- "Writing C# tests", "How to use xUnit", "LINQ queries"
- Automatically invoked when working with .NET projects
- C# SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **xUnit**: Modern .NET testing framework
- **Moq**: Mocking library for interfaces
- **FluentAssertions**: Expressive assertions
- Test coverage ≥85% with Coverlet

**Build Tools**:
- **.NET CLI**: dotnet build, test, run
- **NuGet**: Package management
- **MSBuild**: Build system

**Code Quality**:
- **StyleCop**: C# style checker
- **SonarAnalyzer**: Static code analysis
- **EditorConfig**: Code formatting rules

**C# Patterns**:
- **LINQ**: Query expressions for collections
- **Async/await**: Asynchronous programming
- **Properties**: Get/set accessors
- **Extension methods**: Add methods to existing types
- **Nullable reference types**: Null safety (C# 8+)

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Use PascalCase for public members
- Prefer `var` for local variables when type is obvious
- Async methods should end with "Async" suffix
- Use string interpolation ($"") over concatenation

## Examples
```bash
dotnet test && dotnet format --verify-no-changes
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
- Microsoft. "C# Programming Guide." https://learn.microsoft.com/dotnet/csharp/ (accessed 2025-03-29).
- Microsoft. ".NET Testing with dotnet test." https://learn.microsoft.com/dotnet/core/testing/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C#-specific review)
- web-api-expert (ASP.NET Core API development)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
