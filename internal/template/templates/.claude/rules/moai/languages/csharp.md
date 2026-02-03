---
paths:
  - "**/*.cs"
  - "**/*.csproj"
  - "**/*.sln"
---

# C# Rules

Version: C# 12 / .NET 8

## Tooling

- Build: dotnet CLI or MSBuild
- Linting: .NET analyzers, StyleCop
- Testing: xUnit or NUnit
- Coverage: coverlet >= 85%

## MUST

- Use nullable reference types (enable in csproj)
- Use records for immutable data
- Use async/await for I/O operations
- Use primary constructors for simple classes
- Dispose resources with using statements
- Document public APIs with XML comments

## MUST NOT

- Catch Exception without filtering
- Use async void (except event handlers)
- Ignore nullable warnings
- Use magic strings for configuration
- Block async code with .Result or .Wait()
- Store secrets in appsettings.json

## File Conventions

- *Tests.cs for test files
- One type per file
- Use PascalCase for public members
- Use camelCase for private fields (with _prefix)
- Match namespace to folder structure

## Testing

- Use xUnit with [Theory] for data-driven tests
- Use NSubstitute or Moq for mocking
- Use FluentAssertions for readable assertions
- Use Testcontainers for integration tests

## MoAI Integration

- Use Skill("moai-lang-csharp") for detailed patterns
- Follow TRUST 5 quality gates
