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
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- Microsoft. "C# Programming Guide." https://learn.microsoft.com/dotnet/csharp/ (accessed 2025-03-29).
- Microsoft. ".NET Testing with dotnet test." https://learn.microsoft.com/dotnet/core/testing/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C#-specific review)
- web-api-expert (ASP.NET Core API development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
