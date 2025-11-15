# moai-lang-csharp - CLI Reference & API Documentation

_Last updated: 2025-11-12_

## Quick Reference

### Installation

```bash
# Install .NET SDK (macOS with Homebrew)
brew install dotnet

# Install .NET (Linux/Ubuntu)
wget https://dot.net/v1/dotnet-install.sh -O dotnet-install.sh
chmod +x dotnet-install.sh
./dotnet-install.sh

# Verify installation
dotnet --version
dotnet --list-sdks
```

### Common Commands

```bash
# Create new project
dotnet new console -n MyApp
dotnet new webapi -n MyApi
dotnet new classlib -n MyLibrary

# Build project
dotnet build
dotnet build -c Release

# Run application
dotnet run
dotnet run --project src/MyApp

# Test project
dotnet test
dotnet test --verbosity detailed
dotnet test /p:CollectCoverage=true

# Format code
dotnet format
dotnet format --verify-no-changes

# Lint code (with analyzers)
dotnet build /p:EnforceCodeStyleInBuild=true

# Package library
dotnet pack
dotnet pack -c Release
```

## Tool Versions (2025-11-12)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **.NET** | 9.0.0 | Runtime & SDK | Current |
| **C#** | 13.0.0 | Language | Current |
| **xUnit** | 2.9.0 | Unit testing | Current |
| **NUnit** | 4.2.0 | Unit testing | Current |
| **Entity Framework** | 9.0.0 | ORM | Current |
| **ASP.NET Core** | 9.0.0 | Web framework | Current |
| **Newtonsoft.Json** | 13.0.3 | JSON serialization | Current |
| **Serilog** | 4.0.0 | Logging | Current |

---

## .NET Version Information

| Version | Release Date | Status | Key Features | EOL Date |
|---------|--------------|--------|--------------|----------|
| **.NET 9** | 2024-11 | Current | C# 13, minimal APIs, native AOT | 2026-11 |
| **.NET 8** | 2023-11 | LTS | C# 12, LINQ improvements | 2026-11 |
| **.NET 7** | 2022-11 | Maintenance | C# 11, generic math | 2024-05 |
| **.NET 6** | 2021-11 | LTS | C# 10, records | 2025-11 |

---

## Framework Compatibility Matrix

### C# Language Features by Version

| Feature | C# 11 | C# 12 | C# 13 | Recommended |
|---------|-------|-------|-------|-------------|
| Record types | ✅ | ✅ | ✅ | Yes |
| Required properties | ✅ | ✅ | ✅ | Yes |
| File-scoped types | ✅ | ✅ | ✅ | Yes |
| Primary constructors | ❌ | ✅ | ✅ | Yes |
| Collection expressions | ❌ | ✅ | ✅ | Yes |
| Partial properties | ❌ | ❌ | ✅ | New |
| Ref struct interfaces | ❌ | ❌ | ✅ | New |
| Params collections | ❌ | ❌ | ✅ | New |

### Entity Framework Core Versions

| Version | .NET | Release | Status | Features |
|---------|------|---------|--------|----------|
| **EF Core 9** | 9.0+ | 2024-11 | Current | Temporal tables, JSON arrays |
| **EF Core 8** | 8.0+ | 2023-11 | Current | Complex type support |
| **EF Core 7** | 7.0+ | 2022-11 | Maintenance | Model building improvements |
| **EF Core 6** | 6.0+ | 2021-11 | EOL | Lazy loading, computed columns |

### ASP.NET Core Minimal APIs

| Feature | ASP.NET 6+ | ASP.NET 7+ | ASP.NET 8+ | ASP.NET 9+ |
|---------|-----------|-----------|-----------|-----------|
| MapGet/Post/Put | ✅ | ✅ | ✅ | ✅ |
| Dependency injection | ✅ | ✅ | ✅ | ✅ |
| Filters | ✅ | ✅ | ✅ | ✅ |
| Swagger integration | ✅ | ✅ | ✅ | ✅ |
| OpenAPI support | ❌ | ✅ | ✅ | ✅ |
| Async streaming | ❌ | ❌ | ✅ | ✅ |

---

## API Reference

### Core C# Types

#### Async/Await APIs

```csharp
// Task management
Task                          // Fire-and-forget
Task<T>                       // Async operation returning T
Task.Delay(milliseconds)      // Async delay
Task.Run(action)              // Run on thread pool
Task.WhenAll(tasks)           // Wait for all tasks
Task.WhenAny(tasks)           // Wait for any task
CancellationToken             // Cancellation support
```

#### Collection APIs

```csharp
// LINQ query methods
Where(predicate)              // Filter items
Select(selector)              // Transform items
GroupBy(keySelector)          // Group by key
OrderBy/OrderByDescending()   // Sort items
Join(other, keySelector)      // Join collections
SelectMany(selector)          // Flatten collections
Aggregate(accumulator)        // Reduce to single value
```

#### String APIs

```csharp
// String manipulation
string.IsNullOrEmpty(s)       // Check for null/empty
string.IsNullOrWhiteSpace(s)  // Check for null/whitespace
string.Concat(strings)        // Concatenate strings
string.Join(separator, items) // Join with separator
s.Contains(value)             // Check substring
s.StartsWith/EndsWith(value)  // Check prefix/suffix
s.Split(separator)            // Split string
s.Trim/TrimStart/TrimEnd()    // Remove whitespace
$"interpolated {value}"       // String interpolation
```

#### Exception Handling

```csharp
try { }                       // Try block
catch (Exception ex) { }      // Catch specific exception
catch { }                     // Catch any exception
finally { }                   // Finally block (always runs)
throw new Exception()         // Throw exception
```

### .NET Framework APIs

#### HttpClient (Networking)

```csharp
HttpClient client             // HTTP client
await client.GetAsync(url)    // GET request
await client.PostAsync(url, content)  // POST request
await client.PutAsync(url, content)   // PUT request
await client.DeleteAsync(url) // DELETE request
await response.Content.ReadAsStringAsync()  // Read body
```

#### File I/O

```csharp
File.ReadAllText(path)        // Read entire file
File.WriteAllText(path, text) // Write text to file
File.ReadAllLines(path)       // Read lines from file
Directory.GetFiles(path)      // List files
Path.Combine(paths)           // Combine paths
```

#### JSON Serialization

```csharp
using System.Text.Json;

JsonSerializer.Serialize(obj) // Object to JSON
JsonSerializer.Deserialize<T>(json)  // JSON to object
```

#### Logging

```csharp
using Microsoft.Extensions.Logging;

ILogger logger                // Logging interface
logger.LogInformation(message)        // Info log
logger.LogWarning(message)    // Warning log
logger.LogError(ex, message)  // Error log
```

### Entity Framework Core APIs

#### DbContext Operations

```csharp
DbSet<T>                      // Entity set
context.Add(entity)           // Add entity
context.Update(entity)        // Update entity
context.Remove(entity)        // Delete entity
await context.SaveChangesAsync()      // Save changes
context.Database.Migrate()    // Run migrations
```

#### LINQ to EF Queries

```csharp
dbset.FirstOrDefaultAsync(predicate)  // Get first
dbset.SingleOrDefaultAsync(predicate) // Get single
dbset.ToListAsync()           // Get all
dbset.CountAsync(predicate)   // Count matching
dbset.Include(navigation)     // Eager load
dbset.Where(predicate)        // Filter
```

#### Migrations

```bash
# Create migration
dotnet ef migrations add MigrationName

# Apply migration
dotnet ef database update

# Remove migration
dotnet ef migrations remove

# List migrations
dotnet ef migrations list
```

### ASP.NET Core APIs

#### Minimal API Routing

```csharp
app.MapGet(pattern, handler)  // GET endpoint
app.MapPost(pattern, handler) // POST endpoint
app.MapPut(pattern, handler)  // PUT endpoint
app.MapDelete(pattern, handler)       // DELETE endpoint
app.MapPatch(pattern, handler)        // PATCH endpoint
app.Run()                     // Start server
```

#### Dependency Injection

```csharp
services.AddScoped<T, Implementation>()    // Scoped lifetime
services.AddTransient<T, Implementation>() // Transient lifetime
services.AddSingleton<T, Implementation>() // Singleton lifetime
app.Services.GetRequiredService<T>()       // Resolve service
```

#### Middleware

```csharp
app.UseRouting()              // Enable routing
app.UseAuthentication()       // Enable auth
app.UseAuthorization()        // Enable authz
app.UseStaticFiles()          // Serve static files
app.UseCors()                 // Enable CORS
```

---

## Best Practices

### Async/Await Patterns

1. **Always use async/await**: Prefer async methods for I/O operations
2. **Avoid sync-over-async**: Never use `.Result` or `.Wait()`
3. **Configure .ConfigureAwait(false)**: In library code for context switching
4. **Use CancellationToken**: Support cancellation in long-running operations
5. **Handle exceptions properly**: Use try-catch with async methods

### LINQ Best Practices

1. **Defer execution**: LINQ queries are lazy-evaluated
2. **Avoid LINQ in loops**: Materialize with `.ToList()` before looping
3. **Use efficient predicates**: Filter early in query chains
4. **Prefer EF-translatable queries**: Avoid client-side evaluation
5. **Use projection**: Select only needed fields from database

### Entity Framework Core

1. **Use async methods**: `ToListAsync()`, `FirstOrDefaultAsync()`
2. **Load related data carefully**: Use `.Include()` for eager loading
3. **Track entities appropriately**: Use `.AsNoTracking()` when read-only
4. **Batch operations**: Group multiple changes before `.SaveChangesAsync()`
5. **Handle concurrency**: Implement optimistic locking with timestamps

### ASP.NET Core

1. **Dependency injection**: Use DI for all dependencies
2. **Validate input**: Always validate user input
3. **Return proper status codes**: Use 200, 201, 400, 404, 500 appropriately
4. **Error handling**: Implement global exception middleware
5. **CORS properly**: Configure CORS for specific origins, not wildcard

### Testing with xUnit

1. **Arrange-Act-Assert**: Follow AAA pattern
2. **One assertion per test**: Focus on single behavior
3. **Use Theory for multiple inputs**: `[Theory]` with `[InlineData]`
4. **Mock external dependencies**: Use Moq for interfaces
5. **Test happy path and edge cases**: Include boundary conditions

---

## Environment Setup

### Development Environment

```bash
# Install development tools
dotnet tool install -g dotnet-format
dotnet tool install -g dotnet-ef
dotnet tool install -g dotnet-rimraf

# Create solution structure
dotnet new sln -n MySolution
dotnet new classlib -n MyProject
dotnet sln MySolution.sln add MyProject/MyProject.csproj

# Add test project
dotnet new xunit -n MyProject.Tests
dotnet sln MySolution.sln add MyProject.Tests/MyProject.Tests.csproj
```

### Docker Development

```dockerfile
# Dockerfile
FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src
COPY ["MyApp/MyApp.csproj", "MyApp/"]
RUN dotnet restore "MyApp/MyApp.csproj"
COPY . .
RUN dotnet build "MyApp/MyApp.csproj" -c Release -o /app/build

FROM mcr.microsoft.com/dotnet/runtime:9.0
WORKDIR /app
COPY --from=build /app/publish .
ENTRYPOINT ["dotnet", "MyApp.dll"]
```

---

## Troubleshooting

### Problem: NullReferenceException

**Symptom**: Object reference not set to an instance

**Solution**: Use null-coalescing operator `??` and null-conditional `?.`

```csharp
var value = obj?.Property ?? "default";
```

### Problem: Task not completed

**Symptom**: Test times out or hangs

**Solution**: Ensure all async methods are properly awaited

```csharp
public async Task TestAsync()
{
    var result = await asyncMethod();
    Assert.NotNull(result);
}
```

### Problem: Entity Framework lazy loading not working

**Symptom**: Navigation properties are null

**Solution**: Use `.Include()` for eager loading or enable lazy loading

```csharp
var user = await context.Users
    .Include(u => u.Posts)
    .FirstOrDefaultAsync(u => u.Id == 1);
```

### Problem: Async deadlock

**Symptom**: Application hangs on async method call

**Solution**: Use `.ConfigureAwait(false)` in library code

```csharp
public async Task<Data> GetDataAsync()
{
    var response = await httpClient.GetAsync(url)
        .ConfigureAwait(false);
    return await response.Content.ReadAsAsync<Data>()
        .ConfigureAwait(false);
}
```

### Problem: Configuration not loading

**Symptom**: appsettings.json values are null

**Solution**: Ensure configuration is properly bound

```csharp
var config = builder.Configuration;
var connectionString = config.GetConnectionString("DefaultConnection");
```

---

## Performance Benchmarks

| Operation | Performance | Latency | Memory | Notes |
|-----------|-------------|---------|--------|-------|
| **Async method call** | 1M ops/sec | <1µs | 50MB | Very fast overhead |
| **LINQ query (in-memory)** | 100K ops/sec | <10µs | 100MB | Depends on data size |
| **EF Core query (database)** | 1K-10K ops/sec | 1-100ms | 200MB | Network-bound |
| **JSON serialization** | 500K ops/sec | <5µs | 80MB | For small objects |
| **HTTP request** | 10K req/sec | 100ms | 150MB | Network-bound |
| **Parallel processing** | Varies | Reduced | Higher | Depends on operation |

---

## Related Skills & Resources

- Skill("moai-context7-lang-integration"): Get latest C#/.NET docs via Context7
- Skill("moai-foundation-testing"): C# testing best practices
- Skill("moai-foundation-security"): Security patterns for .NET
- Skill("moai-essentials-debug"): Debugging .NET applications

---

## Official Documentation

| Resource | Link | Coverage |
|----------|------|----------|
| **.NET Documentation** | https://learn.microsoft.com/dotnet/ | Core framework |
| **C# Language** | https://learn.microsoft.com/en-us/dotnet/csharp/ | Language features |
| **Entity Framework Core** | https://learn.microsoft.com/en-us/ef/core/ | ORM |
| **ASP.NET Core** | https://learn.microsoft.com/en-us/aspnet/core/ | Web framework |
| **xUnit.net** | https://xunit.net/docs/getting-started | Unit testing |
| **.NET Community** | https://dotnet.microsoft.com/community | Support forums |

---

## Code Examples by Category

### Async/Await Examples
- Basic async function
- Error handling and timeouts
- Concurrent operations with Task groups
- Cancellation support

### LINQ Examples
- Basic filtering and projection
- Grouping and aggregation
- Joins and complex queries
- LINQ to EF queries

### Entity Framework Examples
- CRUD operations
- Relationships and navigation
- Migrations
- Query optimization

### ASP.NET Core Examples
- Minimal API endpoints
- Dependency injection
- Middleware
- Error handling

### Testing Examples
- Unit tests with xUnit
- Mocking with Moq
- Theory-based tests
- Async test patterns

---

## Glossary

**Async/Await**: Syntax for writing asynchronous code that appears synchronous

**DbContext**: EF Core class representing database connection and operations

**LINQ**: Language-Integrated Query for data operations

**Middleware**: Components in the ASP.NET Core request pipeline

**Dependency Injection**: Pattern for providing object dependencies

**Async Task**: Non-blocking operation that completes asynchronously

**Entity**: Class representing a database table

**Migration**: Database schema change version

**Middleware Pipeline**: Sequence of middleware components

**Service Lifetime**: How long a service instance lives (Scoped, Transient, Singleton)

---

_For working examples, see examples.md_
