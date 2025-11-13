---
name: "moai-lang-csharp"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-12
status: stable
description: "Enterprise C# 13 development with .NET 9, async/await, LINQ, Entity Framework Core, ASP.NET Core, and Context7 MCP integration for modern backend and enterprise applications."
keywords: [csharp, dotnet9, csharp13, aspnetcore, entityframework, backend, async, context7, enterprise]
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# C# - Enterprise v4.0.0

## Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-csharp |
| **Version** | 4.0.0 (2025-11-12) |
| **Allowed tools** | Read, Bash, Context7 MCP |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language Enterprise |
| **Context7 Integration** | ✅ C#/.NET/ASP.NET Core/EF Core |

---

## What It Does

Enterprise C# 13 development featuring async/await for modern concurrency, LINQ for powerful data queries, Entity Framework Core for ORM, ASP.NET Core for web applications, and production-ready patterns for scalable backend services. Context7 MCP integration provides real-time access to official C#/.NET documentation.

**Key capabilities**:
- ✅ C# 13 with async/await and modern syntax
- ✅ Advanced async/await patterns and Task management
- ✅ LINQ for powerful data transformation
- ✅ Entity Framework Core 9 for database operations
- ✅ ASP.NET Core 9 for web applications
- ✅ Dependency injection and middleware patterns
- ✅ Context7 MCP integration for real-time docs
- ✅ Unit testing with xUnit and Moq
- ✅ Performance optimization techniques
- ✅ Enterprise architecture patterns (SOLID, Clean Architecture)

---

## When to Use

**Automatic triggers**:
- C# backend development discussions
- .NET application development
- Async/await and concurrency patterns
- LINQ data transformations
- Entity Framework Core database operations
- ASP.NET Core web application development
- Enterprise service architecture

**Manual invocation**:
- Design backend application architecture
- Implement async/await patterns
- Optimize LINQ queries
- Design database schemas with EF Core
- Build REST APIs with ASP.NET Core
- Implement dependency injection
- Review enterprise C# code

---

## Technology Stack (2025-11-12)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **C#** | 13.0.0 | Core language | ✅ Current |
| **.NET** | 9.0.0 | Runtime & SDK | ✅ Current |
| **ASP.NET Core** | 9.0.0 | Web framework | ✅ Current |
| **Entity Framework Core** | 9.0.0 | ORM | ✅ Current |
| **xUnit** | 2.9.0 | Testing framework | ✅ Current |
| **LINQ** | Built-in | Data queries | ✅ Current |

---

## Quick Start: Hello Async/Await

```csharp
using System;
using System.Threading.Tasks;

public class HelloWorld
{
    public async Task<string> GreetAsync(string name)
    {
        await Task.Delay(100);
        return $"Hello, {name}!";
    }
}

// Usage
var greeter = new HelloWorld();
var greeting = await greeter.GreetAsync("C#");
Console.WriteLine(greeting);
```

---

## Level 1: Quick Reference

### Core Concepts

1. **Async/Await** - Modern asynchronous programming
   - Function marked with `async` - Returns Task
   - Caller uses `await` - Waits for result
   - Native error handling with `throws`
   - Replaces callbacks and completion handlers

2. **LINQ** - Language-Integrated Query
   - Unified syntax for data sources
   - Method chains for data transformation
   - Lazy evaluation (deferred execution)
   - Both in-memory and database queries

3. **Entity Framework Core** - ORM framework
   - DbContext for database operations
   - DbSet<T> for entity collections
   - LINQ to EF for database queries
   - Migrations for schema versioning

4. **ASP.NET Core** - Web application framework
   - Minimal APIs for simple endpoints
   - Dependency injection built-in
   - Middleware pipeline for request handling
   - Type-safe routing

5. **Dependency Injection** - Inversion of control
   - Services registered at startup
   - Automatic injection into constructors
   - Lifetime management (Scoped, Transient, Singleton)
   - Configuration of application behavior

### Project Structure

```
MyApp/
├── Program.cs                  # Application entry
├── Models/                     # Data models
├── Services/                   # Business logic
├── Controllers/                # API endpoints
├── DbContext/                  # Database context
├── Migrations/                 # Database migrations
├── Tests/                      # Unit tests
└── appsettings.json           # Configuration
```

---

## Level 2: Implementation Patterns

### Async/Await Pattern

```csharp
public async Task<User> GetUserAsync(int id)
{
    using var client = new HttpClient();
    var response = await client.GetAsync($"https://api.example.com/users/{id}");
    response.EnsureSuccessStatusCode();
    return await response.Content.ReadAsAsync<User>();
}

// Concurrent operations
public async Task<(Users, Posts)> GetUserDataAsync(int id)
{
    var users = GetUsersAsync(id);
    var posts = GetPostsAsync(id);
    await Task.WhenAll(users, posts);
    return (await users, await posts);
}
```

### LINQ Query Pattern

```csharp
// Fluent syntax
var results = _users
    .Where(u => u.Age > 18)
    .OrderBy(u => u.Name)
    .Select(u => new { u.Name, u.Email })
    .ToList();

// Lazy evaluation - query not executed until ToList()
var query = _users.Where(u => u.IsActive);
var count = query.Count(); // Executes here
```

### Entity Framework Core Pattern

```csharp
public class ApplicationDbContext : DbContext
{
    public DbSet<User> Users { get; set; }
    public DbSet<Post> Posts { get; set; }
}

// CRUD operations
var user = await context.Users.FirstOrDefaultAsync(u => u.Id == id);
context.Users.Add(newUser);
await context.SaveChangesAsync();
```

### ASP.NET Core Minimal API Pattern

```csharp
var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.MapGet("/users/{id}", async (int id, ApplicationDbContext db) =>
    await db.Users.FindAsync(id) is User user
        ? Results.Ok(user)
        : Results.NotFound());

app.Run();
```

---

## Level 3: Advanced Topics

### Async Best Practices

1. **Prefer async/await** over Task.Run for I/O
2. **Use ConfigureAwait(false)** in library code
3. **Handle exceptions** with try-catch around await
4. **Use CancellationToken** for cancellation support
5. **Avoid Task.Result** and Task.Wait() (deadlock risk)

### LINQ Optimization

- **Filter early**: Apply Where before Select
- **Project late**: Select only needed fields at end
- **Use AsNoTracking()**: For read-only queries
- **Avoid N+1 queries**: Use Include for relationships
- **Defer execution**: Chain queries before ToList()

### Performance Patterns

- **Connection pooling**: EF Core handles automatically
- **Batch operations**: Group saves before SaveChangesAsync()
- **Lazy loading**: Avoid loading unnecessary data
- **Caching**: Implement for frequently accessed data
- **Parallel processing**: Use Task.WhenAll for independent operations

### Security Patterns

- **Input validation**: Always validate user input
- **Parameterized queries**: EF Core prevents SQL injection
- **Authentication**: Implement JWT or OAuth2
- **Authorization**: Check permissions before operations
- **Encryption**: Protect sensitive data at rest and in transit

### Testing Strategy

- **Unit tests**: Test business logic with xUnit
- **Integration tests**: Test with real database
- **Mocking**: Use Moq for dependencies
- **AAA pattern**: Arrange-Act-Assert
- **Theory tests**: Multiple inputs with InlineData

---

## Context7 MCP Integration

**Get latest C#/.NET documentation on-demand:**

```python
# Access C# documentation via Context7
from context7 import resolve_library_id, get_library_docs

# C# Language Documentation
csharp_id = resolve_library_id("csharp")
docs = get_library_docs(
    context7_compatible_library_id=csharp_id,
    topic="async-await-patterns",
    tokens=5000
)

# Entity Framework Core Documentation
ef_id = resolve_library_id("entity-framework-core")
ef_docs = get_library_docs(
    context7_compatible_library_id=ef_id,
    topic="querying-data",
    tokens=4000
)

# ASP.NET Core Documentation
aspnetcore_id = resolve_library_id("aspnetcore")
aspnetcore_docs = get_library_docs(
    context7_compatible_library_id=aspnetcore_id,
    topic="dependency-injection",
    tokens=3000
)
```

---

## Related Skills & Resources

**Language Integration**:
- Skill("moai-context7-lang-integration"): Latest C#/.NET documentation

**Quality & Testing**:
- Skill("moai-foundation-testing"): C# testing best practices
- Skill("moai-foundation-trust"): TRUST 5 principles application

**Security & Performance**:
- Skill("moai-foundation-security"): Security patterns for .NET
- Skill("moai-essentials-debug"): C# debugging techniques

**Official Resources**:
- [C# Documentation](https://learn.microsoft.com/dotnet/csharp)
- [.NET Documentation](https://learn.microsoft.com/dotnet/)
- [Entity Framework Core](https://learn.microsoft.com/ef/core)
- [ASP.NET Core](https://learn.microsoft.com/aspnet/core)
- [xUnit Testing](https://xunit.net/docs/getting-started)

---

## Troubleshooting

**Problem**: NullReferenceException
**Solution**: Use null-coalescing operator `??` and null-conditional `?.`

**Problem**: Async deadlock
**Solution**: Use `.ConfigureAwait(false)` in library code or ensure all calls are awaited

**Problem**: N+1 query problem
**Solution**: Use `.Include()` for eager loading or projection with Select

**Problem**: Configuration not loading
**Solution**: Ensure appsettings.json is in correct location and properly configured

---

## Changelog

- **v4.0.0** (2025-11-12): Enterprise upgrade - Progressive Disclosure structure, 83% content reduction, Context7 integration
- **v3.0.0** (2025-03-15): C# 12 and .NET 8 patterns
- **v2.0.0** (2025-01-10): Basic C# async/await patterns
- **v1.0.0** (2024-12-01): Initial release

---

## Resources

**For working examples**: See `examples.md`

**For API reference**: See `reference.md`

**For advanced patterns**: See full SKILL.md in documentation archive

---

_Last updated: 2025-11-12 | Maintained by moai-adk team_
