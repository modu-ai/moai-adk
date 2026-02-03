---
paths:
  - "**/*.cs"
  - "**/*.csproj"
  - "**/*.sln"
---

# C# Rules

Version: C# 12 / .NET 8

## Tooling

- Build: dotnet CLI or Visual Studio
- Testing: xUnit with coverage >= 85%
- Linting: dotnet format

## Preferred Patterns

- Use nullable reference types
- Use records for DTOs
- Use ASP.NET Core minimal APIs

## MoAI Integration

- Use Skill("moai-lang-csharp") for detailed patterns
- Follow TRUST 5 quality gates
