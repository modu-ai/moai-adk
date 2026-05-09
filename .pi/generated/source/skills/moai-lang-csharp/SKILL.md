---
name: moai-lang-csharp
description: >
  C# 12 / .NET 8 development specialist covering ASP.NET Core, Entity Framework, Blazor, and modern C# patterns. Use when developing .NET APIs, web applications, or enterprise solutions.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "language, csharp, dotnet, aspnet-core, entity-framework, blazor"
  context7-libraries: "/dotnet/aspnetcore, /dotnet/efcore, /dotnet/runtime"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["C#", "Csharp", ".NET", "ASP.NET", "Entity Framework", "Blazor", ".cs", ".csproj", ".sln", "dotnet"]
  languages: ["csharp", "c#"]
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

To create a new .NET 8 Web API project, run dotnet new webapi with -n flag for project name and --framework net8.0.

To create a Blazor Web App, run dotnet new blazor with -n flag for project name and --interactivity Auto.

To add Entity Framework Core, run dotnet add package Microsoft.EntityFrameworkCore.SqlServer followed by Microsoft.EntityFrameworkCore.Design.

To add FluentValidation and MediatR, run dotnet add package FluentValidation.AspNetCore and dotnet add package MediatR.

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

Organize projects in a src folder with four main projects. MyApp.Api contains the ASP.NET Core Web API layer with Controllers folder for API Controllers, Endpoints folder for Minimal API endpoints, and Program.cs as the application entry point. MyApp.Application contains business logic including Commands folder for CQRS Commands, Queries folder for CQRS Queries, and Validators folder for FluentValidation. MyApp.Domain contains domain entities including Entities folder for domain models and Interfaces folder for repository interfaces. MyApp.Infrastructure contains data access including Data folder for DbContext and Repositories folder for repository implementations.

### Essential Patterns

Primary Constructor with DI: Define a public class UserService with constructor parameters for IUserRepository and ILogger of UserService. Create async methods like GetByIdAsync that take Guid id, log information using the logger with structured logging for UserId, and return the result from repository.FindByIdAsync.

Minimal API Endpoint: Use app.MapGet with route pattern like "/api/users/{id:guid}" and an async lambda taking Guid id and IUserService. Call the service method, check for null result, and return Results.Ok for found entities or Results.NotFound otherwise. Chain WithName for route naming and WithOpenApi for OpenAPI documentation.

Entity Configuration: Create a class implementing IEntityTypeConfiguration of your entity type. In the Configure method taking EntityTypeBuilder, call HasKey to set the primary key, use Property to configure fields with HasMaxLength and IsRequired, and use HasIndex with IsUnique for unique constraints.

---

## Context7 Integration

For latest documentation, use Context7 MCP tools.

For ASP.NET Core documentation, first resolve the library ID using mcp__context7__resolve-library-id with "aspnetcore", then fetch docs using mcp__context7__get-library-docs with the resolved library ID and topic like "minimal-apis middleware".

For Entity Framework Core documentation, resolve with "efcore" and fetch with topics like "dbcontext migrations".

For .NET Runtime documentation, resolve with "dotnet runtime" and fetch with topics like "collections threading".

---

## Quick Troubleshooting

Build and Runtime: Run dotnet build with --verbosity detailed for detailed output. Run dotnet run with --launch-profile https for HTTPS profile. Run dotnet ef database update to apply EF migrations. Run dotnet ef migrations add with migration name to create new migrations.

Common Patterns:

For null reference handling, use ArgumentNullException.ThrowIfNull with the variable and nameof expression after fetching from context.

For async enumerable streaming, create async methods returning IAsyncEnumerable of your type. Add EnumeratorCancellation attribute to the CancellationToken parameter. Use await foreach with AsAsyncEnumerable and WithCancellation to iterate, yielding each item.

---

## Works Well With

- `moai-domain-backend` - API design, database integration patterns
- `moai-platform-deploy` - Azure, Docker, Kubernetes deployment
- `moai-workflow-testing` - Testing strategies and patterns
- `moai-foundation-quality` - Code quality standards
- `moai-essentials-debug` - Debugging .NET applications
