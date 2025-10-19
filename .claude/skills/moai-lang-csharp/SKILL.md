---
name: moai-lang-csharp
description: C# best practices with xUnit, .NET tooling, LINQ, and async/await patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# C# Expert

## What it does

Provides C#-specific expertise for TDD development, including xUnit testing, .NET CLI tooling, LINQ query expressions, and async/await patterns.

## When to use

- "C# 테스트 작성", "xUnit 사용법", "LINQ 쿼리", ".NET 개발", "Azure", "게임 개발", "웹 API"
- "ASP.NET", ".NET Core", "Entity Framework", "Blazor", "Unity"
- ".NET 마이크로서비스", "WPF", "Windows Forms"
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

## Modern C# (12+)

**Recommended Version**: .NET 8 LTS, .NET 9+ for latest features

**Modern Features**:
- **Primary constructors** (12): Simplified constructor syntax
- **Collection expressions** (12): `[1, 2, 3]` syntax
- **Required properties** (11): `required` keyword
- **Record types** (9+): Immutable data classes with structural equality
- **Nullable reference types** (8+): Non-nullable by default
- **Records** (9): Simplified immutable classes
- **Async Main** (7.1+): Async program entry points
- **File-scoped types** (11): `file class MyClass`

**Version Check**:
```bash
dotnet --version  # Check .NET version
dotnet --info     # Detailed environment info
```

## Package Management Commands

### Using .NET CLI (Built-in - Recommended)
```bash
# Create project
dotnet new console -n MyApp
dotnet new classlib -n MyLib
dotnet new web -n MyWebApi

# Add dependencies
dotnet add package Newtonsoft.Json
dotnet add package xunit
dotnet add package -v 6.1.0 EntityFramework  # Specific version

# Remove dependencies
dotnet remove package Newtonsoft.Json

# Build & test
dotnet build
dotnet test
dotnet run

# Check dependencies
dotnet list package
dotnet list package --outdated
dotnet list package --vulnerable

# Update dependencies
dotnet package update  # Update all
dotnet package update Newtonsoft.Json

# Publish
dotnet publish -c Release

# Restore dependencies
dotnet restore
```

### NuGet Package.csproj Configuration
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <LangVersion>12</LangVersion>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="xunit" Version="2.6.0" />
    <PackageReference Include="Moq" Version="4.18.0" />
    <PackageReference Include="FluentAssertions" Version="6.11.0" />
  </ItemGroup>

  <ItemGroup>
    <ProjectReference Include="../MyLib/MyLib.csproj" />
  </ItemGroup>
</Project>
```

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
