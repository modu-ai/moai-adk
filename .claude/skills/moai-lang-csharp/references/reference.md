# C# 12 / .NET 8 API Reference

Comprehensive API reference and Context7 library mappings for C# 12 and .NET 8 development.

---

## Context7 Library Mappings

### Primary Libraries

ASP.NET Core:
- Context7 ID: `/dotnet/aspnetcore`
- Topics: minimal-apis, controllers, middleware, authentication, authorization, blazor, signalr
- Version: 8.0+

Entity Framework Core:
- Context7 ID: `/dotnet/efcore`
- Topics: dbcontext, migrations, linq, query-optimization, relationships, tracking
- Version: 8.0+

.NET Runtime:
- Context7 ID: `/dotnet/runtime`
- Topics: collections, threading, memory, span, async, linq
- Version: 8.0+

### Context7 Query Patterns

```
# Minimal APIs and Middleware
mcp__context7__get-library-docs("/dotnet/aspnetcore", "minimal-apis endpoint routing")
mcp__context7__get-library-docs("/dotnet/aspnetcore", "middleware pipeline")

# Authentication and Authorization
mcp__context7__get-library-docs("/dotnet/aspnetcore", "jwt authentication bearer")
mcp__context7__get-library-docs("/dotnet/aspnetcore", "authorization policies claims")

# Entity Framework Queries
mcp__context7__get-library-docs("/dotnet/efcore", "dbcontext configuration")
mcp__context7__get-library-docs("/dotnet/efcore", "linq queries optimization")
mcp__context7__get-library-docs("/dotnet/efcore", "migrations code-first")

# Blazor Components
mcp__context7__get-library-docs("/dotnet/aspnetcore", "blazor components lifecycle")
mcp__context7__get-library-docs("/dotnet/aspnetcore", "blazor interactiveserver wasm")

# Collections and Threading
mcp__context7__get-library-docs("/dotnet/runtime", "concurrent collections")
mcp__context7__get-library-docs("/dotnet/runtime", "async await patterns")
```

---

## .NET CLI Reference

### Project Creation

```bash
# Web API with .NET 8
dotnet new webapi -n MyApi --framework net8.0

# Blazor Web App with Auto interactivity
dotnet new blazor -n MyBlazor --interactivity Auto

# Class Library
dotnet new classlib -n MyLib --framework net8.0

# xUnit Test Project
dotnet new xunit -n MyApi.Tests --framework net8.0

# Solution file
dotnet new sln -n MySolution
dotnet sln add src/MyApi/MyApi.csproj
```

### Package Management

```bash
# Entity Framework Core
dotnet add package Microsoft.EntityFrameworkCore.SqlServer
dotnet add package Microsoft.EntityFrameworkCore.Design
dotnet add package Microsoft.EntityFrameworkCore.Tools

# Validation and CQRS
dotnet add package FluentValidation.AspNetCore
dotnet add package MediatR

# Authentication
dotnet add package Microsoft.AspNetCore.Authentication.JwtBearer

# Testing
dotnet add package xunit
dotnet add package Moq
dotnet add package FluentAssertions

# API Documentation
dotnet add package Swashbuckle.AspNetCore
```

### Entity Framework CLI

```bash
# Install EF tools
dotnet tool install --global dotnet-ef

# Migrations
dotnet ef migrations add InitialCreate
dotnet ef migrations add AddUserTable
dotnet ef database update
dotnet ef database update InitialCreate  # Rollback to specific migration

# Generate SQL script
dotnet ef migrations script

# Scaffold from existing database
dotnet ef dbcontext scaffold "Connection" Microsoft.EntityFrameworkCore.SqlServer
```

---

## NuGet Package Versions (.NET 8)

### Core Packages

```xml
<PackageReference Include="Microsoft.EntityFrameworkCore" Version="8.0.11" />
<PackageReference Include="Microsoft.EntityFrameworkCore.SqlServer" Version="8.0.11" />
<PackageReference Include="Microsoft.EntityFrameworkCore.Design" Version="8.0.11" />
<PackageReference Include="Microsoft.AspNetCore.Authentication.JwtBearer" Version="8.0.11" />
```

### Validation and CQRS

```xml
<PackageReference Include="FluentValidation" Version="11.11.0" />
<PackageReference Include="FluentValidation.AspNetCore" Version="11.3.0" />
<PackageReference Include="FluentValidation.DependencyInjectionExtensions" Version="11.11.0" />
<PackageReference Include="MediatR" Version="12.4.1" />
```

### Testing

```xml
<PackageReference Include="xunit" Version="2.9.2" />
<PackageReference Include="xunit.runner.visualstudio" Version="2.8.2" />
<PackageReference Include="Moq" Version="4.20.72" />
<PackageReference Include="FluentAssertions" Version="6.12.2" />
<PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.12.0" />
```

### Logging and Observability

```xml
<PackageReference Include="Serilog.AspNetCore" Version="8.0.3" />
<PackageReference Include="Serilog.Sinks.Console" Version="6.0.0" />
<PackageReference Include="OpenTelemetry.Extensions.Hosting" Version="1.10.0" />
```

---

## Configuration Patterns

### appsettings.json Structure

```json
{
  "ConnectionStrings": {
    "Default": "Server=localhost;Database=MyDb;Trusted_Connection=true;TrustServerCertificate=true"
  },
  "Jwt": {
    "Issuer": "https://myapp.com",
    "Audience": "https://myapp.com",
    "Key": "your-secret-key-min-32-chars",
    "ExpirationMinutes": 60
  },
  "Logging": {
    "LogLevel": {
      "Default": "Information",
      "Microsoft.AspNetCore": "Warning",
      "Microsoft.EntityFrameworkCore": "Warning"
    }
  }
}
```

### Environment-Specific Configuration

```
appsettings.json                    # Base configuration
appsettings.Development.json        # Development overrides
appsettings.Production.json         # Production overrides
```

Access configuration:
```csharp
builder.Configuration.GetConnectionString("Default")
builder.Configuration["Jwt:Issuer"]
builder.Configuration.GetSection("Jwt").Get<JwtSettings>()
```

---

## HTTP Status Codes Reference

### Success Responses

```csharp
Results.Ok(data)              // 200 OK
Results.Created(uri, data)    // 201 Created
Results.Accepted()            // 202 Accepted
Results.NoContent()           // 204 No Content
```

### Client Error Responses

```csharp
Results.BadRequest(errors)           // 400 Bad Request
Results.Unauthorized()               // 401 Unauthorized
Results.Forbid()                     // 403 Forbidden
Results.NotFound()                   // 404 Not Found
Results.Conflict()                   // 409 Conflict
Results.UnprocessableEntity(errors)  // 422 Unprocessable Entity
```

### Server Error Responses

```csharp
Results.Problem(detail: "Error message")  // 500 Internal Server Error
```

---

## Dependency Injection Lifetimes

### Service Lifetimes

```csharp
// Singleton: Single instance for application lifetime
builder.Services.AddSingleton<ICacheService, MemoryCacheService>();

// Scoped: Single instance per HTTP request
builder.Services.AddScoped<IUserService, UserService>();
builder.Services.AddScoped<IUserRepository, EfUserRepository>();

// Transient: New instance every time requested
builder.Services.AddTransient<IEmailService, SmtpEmailService>();
```

### Registration Patterns

```csharp
// Interface to implementation
builder.Services.AddScoped<IService, Service>();

// Self-registration
builder.Services.AddScoped<MyService>();

// Factory pattern
builder.Services.AddScoped<IService>(sp => 
    new Service(sp.GetRequiredService<IDependency>()));

// Options pattern
builder.Services.Configure<JwtSettings>(
    builder.Configuration.GetSection("Jwt"));
```

---

## Middleware Pipeline Order

### Recommended Order

```csharp
var app = builder.Build();

// 1. Exception handling (first)
app.UseExceptionHandler("/error");

// 2. HTTPS redirection
app.UseHttpsRedirection();

// 3. Static files (if needed)
app.UseStaticFiles();

// 4. Routing
app.UseRouting();

// 5. CORS (before auth)
app.UseCors();

// 6. Authentication
app.UseAuthentication();

// 7. Authorization
app.UseAuthorization();

// 8. Custom middleware
app.UseCustomMiddleware();

// 9. Endpoint mapping (last)
app.MapControllers();
app.MapBlazorHub();
```

---

## Performance Best Practices

### EF Core Query Optimization

```csharp
// Use AsNoTracking for read-only queries
var users = await context.Users.AsNoTracking().ToListAsync();

// Use projection to select only needed columns
var userDtos = await context.Users
    .Select(u => new UserDto(u.Id, u.Name))
    .ToListAsync();

// Use compiled queries for hot paths
private static readonly Func<AppDbContext, Guid, Task<User?>> GetUserById =
    EF.CompileAsyncQuery((AppDbContext ctx, Guid id) =>
        ctx.Users.FirstOrDefault(u => u.Id == id));

// Batch operations
await context.Users.Where(u => u.IsDeleted).ExecuteDeleteAsync();
await context.Users.Where(u => u.Role == "User").ExecuteUpdateAsync(
    s => s.SetProperty(u => u.Role, "Member"));
```

### Async Best Practices

```csharp
// Always use async/await for I/O operations
public async Task<User?> GetByIdAsync(Guid id, CancellationToken ct = default)
    => await context.Users.FindAsync([id], ct);

// Use ConfigureAwait(false) in library code
public async Task<string> GetDataAsync()
{
    var data = await httpClient.GetStringAsync(url).ConfigureAwait(false);
    return Process(data);
}

// Use ValueTask for frequently synchronous paths
public ValueTask<int> GetCachedValueAsync()
{
    if (_cache.TryGetValue(key, out var value))
        return ValueTask.FromResult(value);
    return new ValueTask<int>(FetchFromDatabaseAsync());
}
```

---

## Security Best Practices

### Input Validation

```csharp
// Always validate and sanitize input
[HttpPost]
public async Task<ActionResult> Create([FromBody] CreateRequest request)
{
    // Model validation via attributes
    if (!ModelState.IsValid)
        return BadRequest(ModelState);
    
    // Business validation
    var result = await validator.ValidateAsync(request);
    if (!result.IsValid)
        return BadRequest(result.Errors);
}
```

### SQL Injection Prevention

```csharp
// Use parameterized queries (EF Core does this automatically)
var user = await context.Users.FirstOrDefaultAsync(u => u.Email == email);

// For raw SQL, always use parameters
var users = await context.Users
    .FromSqlInterpolated($"SELECT * FROM Users WHERE Email = {email}")
    .ToListAsync();
```

### Secrets Management

```csharp
// Use User Secrets for development
dotnet user-secrets init
dotnet user-secrets set "Jwt:Key" "your-secret-key"

// Use Azure Key Vault for production
builder.Configuration.AddAzureKeyVault(
    new Uri("https://myvault.vault.azure.net/"),
    new DefaultAzureCredential());
```

---

Version: 2.0.0
Last Updated: 2026-01-06
