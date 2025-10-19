---
name: moai-lang-csharp
description: C# best practices with xUnit, .NET tooling, LINQ, and async/await patterns
allowed-tools:
- Read
- Bash
---

# C# Expert

## What it does

Provides C#-specific expertise for TDD development, including xUnit testing, .NET CLI tooling, LINQ query expressions, and async/await patterns.

## When to use

- "C# 테스트 작성", "xUnit 사용법", "LINQ 쿼리"
- Automatically invoked when working with .NET projects
- C# SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with xUnit
User: "/alfred:2-build SERVICE-001"
Claude: (creates RED test with xUnit, GREEN implementation with async/await, REFACTOR)

### Example 2: LINQ query optimization
User: "LINQ 쿼리 최적화"
Claude: (analyzes LINQ queries and suggests IEnumerable vs IQueryable optimizations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C#-specific review)
- web-api-expert (ASP.NET Core API development)
