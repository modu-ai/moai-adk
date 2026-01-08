---
name: "moai-lang-csharp"
description: "C# 12 / .NET 8 development specialist covering ASP.NET Core, Entity Framework, Blazor, and modern C# patterns. Use when developing .NET APIs, web applications, or enterprise solutions."
version: 2.0.0
category: "language"
modularized: true
user-invocable: false
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
context7-libraries: ['/dotnet/aspnetcore', '/dotnet/efcore', '/dotnet/runtime']
updated: 2026-01-08
status: "active"
---

# C# 12 / .NET 8 Development Specialist

Modern C# development with ASP.NET Core, Entity Framework Core, Blazor, and enterprise patterns.

## Quick Reference

Auto-Triggers: `.cs`, `.csproj`, `.sln` files, C# projects, .NET solutions, ASP.NET Core applications

Core Stack:
- C# 12: Primary constructors, collection expressions, alias any type, default lambda parameters
- .NET 8: Minimal APIs, Native AOT, improved performance, WebSockets
- ASP.NET Core 8: Controllers, Endpoints, Middleware, Authentication
- Entity Framework Core 8: DbContext, migrations, LINQ, query optimization
- Blazor: Server/WASM components, InteractiveServer, InteractiveWebAssembly
- Testing: xUnit, NUnit, FluentAssertions, Moq

Quick Commands:
```bash
# Create .NET 8 Web API project
dotnet new webapi -n MyApi --framework net8.0

# Create Blazor Web App
dotnet new blazor -n MyBlazor --interactivity Auto

# Add Entity Framework Core
dotnet add package Microsoft.EntityFrameworkCore.SqlServer
dotnet add package Microsoft.EntityFrameworkCore.Design

# Add FluentValidation and MediatR
dotnet add package FluentValidation.AspNetCore
dotnet add package MediatR
```

---

## Module Index

This skill uses progressive disclosure with specialized modules for deep coverage:

### Language Features
- [C# 12 Features](modules/csharp12-features.md) - Primary constructors, collection expressions, type aliases, default lambdas

### Web Development
- [ASP.NET Core 8](modules/aspnet-core.md) - Minimal API, Controllers, Middleware, Authentication
- [Blazor Components](modules/blazor-components.md) - Server, WASM, InteractiveServer, Components

### Data Access
- [Entity Framework Core 8](modules/efcore-patterns.md) - DbContext, Repository pattern, Migrations, Query optimization

### Architecture Patterns
- [CQRS and Validation](modules/cqrs-validation.md) - MediatR CQRS, FluentValidation, Handler patterns

### Reference Materials
- [API Reference](reference.md) - Complete API reference, Context7 library mappings
- [Code Examples](examples.md) - Production-ready examples, testing templates

---

## Implementation Quick Start

### Project Structure (Clean Architecture)

```
src/
├── MyApp.Api/              # ASP.NET Core Web API
│   ├── Controllers/        # API Controllers
│   ├── Endpoints/          # Minimal API endpoints
│   └── Program.cs          # Application entry
├── MyApp.Application/      # Business logic layer
│   ├── Commands/           # CQRS Commands
│   ├── Queries/            # CQRS Queries
│   └── Validators/         # FluentValidation
├── MyApp.Domain/           # Domain entities
│   ├── Entities/           # Domain models
│   └── Interfaces/         # Repository interfaces
└── MyApp.Infrastructure/   # Data access layer
    ├── Data/               # DbContext
    └── Repositories/       # Repository implementations
```

### Essential Patterns

Primary Constructor with DI:
```csharp
public class UserService(IUserRepository repository, ILogger<UserService> logger)
{
    public async Task<User?> GetByIdAsync(Guid id)
    {
        logger.LogInformation("Fetching user {UserId}", id);
        return await repository.FindByIdAsync(id);
    }
}
```

Minimal API Endpoint:
```csharp
app.MapGet("/api/users/{id:guid}", async (Guid id, IUserService service) =>
{
    var user = await service.GetByIdAsync(id);
    return user is not null ? Results.Ok(user) : Results.NotFound();
})
.WithName("GetUser")
.WithOpenApi();
```

Entity Configuration:
```csharp
public class UserConfiguration : IEntityTypeConfiguration<User>
{
    public void Configure(EntityTypeBuilder<User> builder)
    {
        builder.HasKey(u => u.Id);
        builder.Property(u => u.Email).HasMaxLength(256).IsRequired();
        builder.HasIndex(u => u.Email).IsUnique();
    }
}
```

---

## Context7 Integration

For latest documentation, use Context7 MCP tools:

```
# ASP.NET Core documentation
mcp__context7__resolve-library-id("aspnetcore")
mcp__context7__get-library-docs("/dotnet/aspnetcore", "minimal-apis middleware")

# Entity Framework Core documentation
mcp__context7__resolve-library-id("efcore")
mcp__context7__get-library-docs("/dotnet/efcore", "dbcontext migrations")

# .NET Runtime documentation
mcp__context7__resolve-library-id("dotnet runtime")
mcp__context7__get-library-docs("/dotnet/runtime", "collections threading")
```

---

## Quick Troubleshooting

Build and Runtime:
```bash
dotnet build --verbosity detailed    # Detailed build output
dotnet run --launch-profile https    # Run with HTTPS profile
dotnet ef database update            # Apply EF migrations
dotnet ef migrations add Initial     # Create new migration
```

Common Patterns:
```csharp
// Null reference handling
var user = await context.Users.FindAsync(id);
ArgumentNullException.ThrowIfNull(user, nameof(user));

// Async enumerable for streaming
public async IAsyncEnumerable<User> StreamUsersAsync(
    [EnumeratorCancellation] CancellationToken ct = default)
{
    await foreach (var user in context.Users.AsAsyncEnumerable().WithCancellation(ct))
    {
        yield return user;
    }
}
```

---

## Works Well With

- `moai-domain-backend` - API design, database integration patterns
- `moai-platform-deploy` - Azure, Docker, Kubernetes deployment
- `moai-workflow-testing` - Testing strategies and patterns
- `moai-foundation-quality` - Code quality standards
- `moai-essentials-debug` - Debugging .NET applications
