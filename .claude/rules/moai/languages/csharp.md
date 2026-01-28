---
paths:
  - "**/*.cs"
  - "**/*.csproj"
  - "**/*.sln"
---

# C# Rules

Version: C# 12 / .NET 8 LTS

## Tooling

- Build: dotnet CLI or Visual Studio 2026
- Testing: xUnit with coverage >= 85%
- Linting: dotnet format, Roslyn analyzers

## Best Practices (2026)

- Use Primary Constructors for cleaner class definitions
- Use Collection Expressions: `int[] nums = [1, 2, 3];`
- Enable nullable reference types project-wide
- Use records for immutable DTOs
- Dynamic PGO is enabled by default (~15% performance boost)

## C# 12 Features

```csharp
// Primary constructor
public class UserService(IUserRepository repo, ILogger<UserService> logger)
{
    public User GetUser(int id) => repo.FindById(id);
}

// Collection expressions
List<int> numbers = [1, 2, 3, 4, 5];
int[] combined = [..numbers, 6, 7, 8];

// Required members
public class Config
{
    public required string ConnectionString { get; init; }
}
```

## ASP.NET Core Patterns

- Use Minimal APIs for simple endpoints
- Use HybridCache for combined memory + distributed caching
- Implement rate limiting middleware for API protection
- Use Native AOT for serverless/container deployments

## Migration Guidance

- Incremental upgrades recommended: 6 → 7 → 8 → 9 → 10
- Review breaking changes docs before each upgrade
- Use static analyzers to detect deprecated patterns

## MoAI Integration

- Use Skill("moai-lang-csharp") for detailed patterns
- Follow TRUST 5 quality gates
