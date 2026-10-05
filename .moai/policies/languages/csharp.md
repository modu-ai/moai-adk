---
paths: "**/*.cs,**/*.csproj,**/*.sln"
---

# C# Development Guide

# C# 14 / .NET 10 (LTS) Development Specialist

Modern C# development with ASP.NET Core, Entity Framework Core, Blazor, and enterprise patterns.


Core Stack:

- C# 14: Primary constructors, collection expressions, alias any type, default lambda parameters
- .NET 10: Minimal APIs, Native AOT, improved performance, WebSockets
- ASP.NET Core 10: Controllers, Endpoints, Middleware, Authentication
- Entity Framework Core 10: DbContext, migrations, LINQ, query optimization
- Blazor: Server/WASM components, InteractiveServer, InteractiveWebAssembly
- Testing: xUnit, NUnit, FluentAssertions, Moq

Quick Commands:

To create a new .NET 10 Web API project, run dotnet new webapi with -n flag for project name and --framework net10.0.

To create a Blazor Web App, run dotnet new blazor with -n flag for project name and --interactivity Auto.

To add Entity Framework Core, run dotnet add package Microsoft.EntityFrameworkCore.SqlServer followed by Microsoft.EntityFrameworkCore.Design.

To add FluentValidation and MediatR, run dotnet add package FluentValidation.AspNetCore and dotnet add package MediatR.

---

## Coverage Areas

This guide is self-contained. Use the sections below as the primary reference for:

- C# 14 language features such as primary constructors, collection expressions, type aliases, and default lambda parameters
- ASP.NET Core 10 application structure, Minimal APIs, controllers, middleware, and authentication
- Blazor Server / WebAssembly component patterns
- Entity Framework Core 10 data-access and migration workflows
- CQRS + FluentValidation application patterns

---

## Implementation Quick Start

### Project Structure (Clean Architecture)

Organize projects in a src folder with four main projects. MyApp.Api contains the ASP.NET Core Web API layer with Controllers folder for API Controllers, Endpoints folder for Minimal API endpoints, and Program.cs as the application entry point. MyApp.Application contains business logic including Commands folder for CQRS Commands, Queries folder for CQRS Queries, and Validators folder for FluentValidation. MyApp.Domain contains domain entities including Entities folder for domain models and Interfaces folder for repository interfaces. MyApp.Infrastructure contains data access including Data folder for DbContext and Repositories folder for repository implementations.

### Essential Patterns

Primary Constructor with DI: Define a public class UserService with constructor parameters for IUserRepository and ILogger of UserService. Create async methods like GetByIdAsync that take Guid id, log information using the logger with structured logging for UserId, and return the result from repository.FindByIdAsync.

Minimal API Endpoint: Use app.MapGet with route pattern like "/api/users/{id:guid}" and an async lambda taking Guid id and IUserService. Call the service method, check for null result, and return Results.Ok for found entities or Results.NotFound otherwise. Chain WithName for route naming and WithOpenApi for OpenAPI documentation.

Entity Configuration: Create a class implementing IEntityTypeConfiguration of your entity type. In the Configure method taking EntityTypeBuilder, call HasKey to set the primary key, use Property to configure fields with HasMaxLength and IsRequired, and use HasIndex with IsUnique for unique constraints.

---

## Documentation References

For latest documentation, use WebSearch / WebFetch against the official Microsoft / .NET docs.

For ASP.NET Core documentation, WebSearch "ASP.NET Core minimal APIs middleware" and WebFetch the relevant learn.microsoft.com page.

For Entity Framework Core documentation, WebSearch "EF Core dbcontext migrations" and fetch the official docs.

For .NET Runtime documentation, WebSearch ".NET runtime collections threading" and fetch the official docs.

---

## Testing

xUnit Fixture Pattern: Define a public class implementing ICollectionFixture of DatabaseFixture. Add Collection attribute referencing the collection definition. Tests in the collection share the fixture instance across test runs, ensuring database state setup only once.

NUnit TestCase Pattern: Use TestCase attribute with expected return values. Define test method with input parameters. NUnit runs the test once per TestCase attribute, comparing actual to expected via Assert.AreEqual.

Moq Setup Pattern: Create Mock of IRepository. Call Setup with lambda matching method call and argument constraint (It.IsAny of T). Configure Returns with expected value. Verify the mock was called with specific arguments using Verify and Times.Once.

FluentAssertions Style: Use Should assertions. Call result.Should().BeEquivalentTo expecting matching object structure. Use Should().Throw of ExceptionType for exception assertions in async methods. Note: FluentAssertions changed its license in 2025 — consider plain xUnit asserts or Shouldly as alternatives for new projects.

---

## Troubleshooting

Build and Runtime: Run dotnet build with --verbosity detailed for detailed output. Run dotnet run with --launch-profile https for HTTPS profile. Run dotnet ef database update to apply EF migrations. Run dotnet ef migrations add with migration name to create new migrations.

Common Patterns:

For null reference handling, use ArgumentNullException.ThrowIfNull with the variable and nameof expression after fetching from context.

For async enumerable streaming, create async methods returning IAsyncEnumerable of your type. Add EnumeratorCancellation attribute to the CancellationToken parameter. Use await foreach with AsAsyncEnumerable and WithCancellation to iterate, yielding each item.

---


- `moai-domain-backend` - API design, database integration patterns
- `moai-workflow-testing` - Testing strategies and patterns
- `moai-foundation-quality` - Code quality standards
