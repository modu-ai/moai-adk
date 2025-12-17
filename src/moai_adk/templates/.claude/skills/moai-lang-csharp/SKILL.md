---
name: moai-lang-csharp
description: C# 12 / .NET 8 development specialist covering ASP.NET Core, Entity Framework, Blazor, and modern C# patterns. Use when developing .NET APIs, web applications, or enterprise solutions.
version: 1.0.0
category: language
tags:
  - csharp
  - dotnet
  - aspnetcore
  - efcore
  - blazor
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

C# 12 / .NET 8 Development Specialist - Modern C# with ASP.NET Core, Entity Framework Core, Blazor, and enterprise patterns.

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

## Implementation Guide (5 minutes)

### C# 12 Key Features

Primary Constructors - Class-level constructor parameters:
```csharp
// Primary constructor with dependency injection
public class UserService(IUserRepository repository, ILogger<UserService> logger)
{
    public async Task<User?> GetByIdAsync(Guid id)
    {
        logger.LogInformation("Fetching user {UserId}", id);
        return await repository.FindByIdAsync(id);
    }
}

// Record with primary constructor
public record CreateUserCommand(string Name, string Email);
```

Collection Expressions - Unified collection syntax:
```csharp
// Array, List, Span with unified syntax
int[] numbers = [1, 2, 3, 4, 5];
List<string> names = ["Alice", "Bob", "Charlie"];
Span<int> span = [10, 20, 30];

// Spread operator
int[] combined = [..numbers, 6, 7, 8];
List<string> allNames = [..names, "David", "Eve"];
```

Alias Any Type - Type aliases for complex types:
```csharp
// Alias for tuple
using Point = (int X, int Y);

// Alias for complex generic
using UserCache = System.Collections.Generic.Dictionary<Guid, User>;

public class LocationService
{
    public Point GetLocation() => (10, 20);
    private readonly UserCache _cache = [];
}
```

Default Lambda Parameters:
```csharp
// Lambda with default parameter
var greet = (string name, string greeting = "Hello") => $"{greeting}, {name}!";

Console.WriteLine(greet("Alice"));           // "Hello, Alice!"
Console.WriteLine(greet("Bob", "Hi"));       // "Hi, Bob!"
```

### ASP.NET Core 8 Patterns

Minimal API with Endpoints:
```csharp
var builder = WebApplication.CreateBuilder(args);

// Service registration
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("Default")));
builder.Services.AddScoped<IUserService, UserService>();

var app = builder.Build();

// Endpoint routing with typed results
app.MapGet("/api/users/{id:guid}", async (Guid id, IUserService service) =>
{
    var user = await service.GetByIdAsync(id);
    return user is not null ? Results.Ok(user) : Results.NotFound();
})
.WithName("GetUser")
.WithOpenApi()
.Produces<User>(200)
.Produces(404);

app.MapPost("/api/users", async (CreateUserRequest request, IUserService service) =>
{
    var user = await service.CreateAsync(request);
    return Results.Created($"/api/users/{user.Id}", user);
})
.WithValidation<CreateUserRequest>();

app.Run();
```

Controller-Based API:
```csharp
[ApiController]
[Route("api/[controller]")]
public class UsersController(IUserService userService, ILogger<UsersController> logger)
    : ControllerBase
{
    [HttpGet("{id:guid}")]
    [ProducesResponseType<User>(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<ActionResult<User>> GetById(Guid id)
    {
        var user = await userService.GetByIdAsync(id);
        if (user is null)
        {
            logger.LogWarning("User {UserId} not found", id);
            return NotFound();
        }
        return user;
    }

    [HttpPost]
    [ProducesResponseType<User>(StatusCodes.Status201Created)]
    [ProducesResponseType<ValidationProblemDetails>(StatusCodes.Status400BadRequest)]
    public async Task<ActionResult<User>> Create([FromBody] CreateUserRequest request)
    {
        var user = await userService.CreateAsync(request);
        return CreatedAtAction(nameof(GetById), new { id = user.Id }, user);
    }
}
```

### Entity Framework Core 8 Patterns

DbContext Configuration:
```csharp
public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<User> Users => Set<User>();
    public DbSet<Post> Posts => Set<Post>();
    public DbSet<Tag> Tags => Set<Tag>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(AppDbContext).Assembly);
    }
}

// Entity configuration
public class UserConfiguration : IEntityTypeConfiguration<User>
{
    public void Configure(EntityTypeBuilder<User> builder)
    {
        builder.HasKey(u => u.Id);
        builder.Property(u => u.Email).HasMaxLength(256).IsRequired();
        builder.HasIndex(u => u.Email).IsUnique();
        builder.HasMany(u => u.Posts).WithOne(p => p.Author).HasForeignKey(p => p.AuthorId);
    }
}
```

Repository Pattern with Specification:
```csharp
public interface IRepository<T> where T : class
{
    Task<T?> GetByIdAsync(Guid id, CancellationToken ct = default);
    Task<IReadOnlyList<T>> ListAsync(CancellationToken ct = default);
    Task<T> AddAsync(T entity, CancellationToken ct = default);
    Task UpdateAsync(T entity, CancellationToken ct = default);
    Task DeleteAsync(T entity, CancellationToken ct = default);
}

public class EfRepository<T>(AppDbContext context) : IRepository<T> where T : class
{
    private readonly DbSet<T> _dbSet = context.Set<T>();

    public async Task<T?> GetByIdAsync(Guid id, CancellationToken ct = default)
        => await _dbSet.FindAsync([id], ct);

    public async Task<IReadOnlyList<T>> ListAsync(CancellationToken ct = default)
        => await _dbSet.ToListAsync(ct);

    public async Task<T> AddAsync(T entity, CancellationToken ct = default)
    {
        await _dbSet.AddAsync(entity, ct);
        await context.SaveChangesAsync(ct);
        return entity;
    }

    public async Task UpdateAsync(T entity, CancellationToken ct = default)
    {
        _dbSet.Update(entity);
        await context.SaveChangesAsync(ct);
    }

    public async Task DeleteAsync(T entity, CancellationToken ct = default)
    {
        _dbSet.Remove(entity);
        await context.SaveChangesAsync(ct);
    }
}
```

### FluentValidation Patterns

Request Validation:
```csharp
public record CreateUserRequest(string Name, string Email, string Password);

public class CreateUserRequestValidator : AbstractValidator<CreateUserRequest>
{
    public CreateUserRequestValidator(IUserRepository userRepository)
    {
        RuleFor(x => x.Name)
            .NotEmpty().WithMessage("Name is required")
            .MaximumLength(100).WithMessage("Name cannot exceed 100 characters");

        RuleFor(x => x.Email)
            .NotEmpty().WithMessage("Email is required")
            .EmailAddress().WithMessage("Invalid email format")
            .MustAsync(async (email, ct) => !await userRepository.EmailExistsAsync(email, ct))
            .WithMessage("Email already exists");

        RuleFor(x => x.Password)
            .NotEmpty().WithMessage("Password is required")
            .MinimumLength(8).WithMessage("Password must be at least 8 characters")
            .Matches(@"[A-Z]").WithMessage("Password must contain uppercase letter")
            .Matches(@"[a-z]").WithMessage("Password must contain lowercase letter")
            .Matches(@"[0-9]").WithMessage("Password must contain digit");
    }
}

// Registration in Program.cs
builder.Services.AddValidatorsFromAssemblyContaining<CreateUserRequestValidator>();
```

### MediatR CQRS Pattern

Command and Query Separation:
```csharp
// Query
public record GetUserByIdQuery(Guid Id) : IRequest<User?>;

public class GetUserByIdQueryHandler(AppDbContext context)
    : IRequestHandler<GetUserByIdQuery, User?>
{
    public async Task<User?> Handle(GetUserByIdQuery request, CancellationToken ct)
        => await context.Users
            .AsNoTracking()
            .Include(u => u.Posts)
            .FirstOrDefaultAsync(u => u.Id == request.Id, ct);
}

// Command
public record CreateUserCommand(string Name, string Email, string Password) : IRequest<User>;

public class CreateUserCommandHandler(
    AppDbContext context,
    IPasswordHasher passwordHasher,
    IValidator<CreateUserCommand> validator)
    : IRequestHandler<CreateUserCommand, User>
{
    public async Task<User> Handle(CreateUserCommand request, CancellationToken ct)
    {
        await validator.ValidateAndThrowAsync(request, ct);

        var user = new User
        {
            Id = Guid.NewGuid(),
            Name = request.Name,
            Email = request.Email,
            PasswordHash = passwordHasher.Hash(request.Password),
            CreatedAt = DateTime.UtcNow
        };

        context.Users.Add(user);
        await context.SaveChangesAsync(ct);
        return user;
    }
}

// Usage in controller
[HttpPost]
public async Task<ActionResult<User>> Create(
    [FromBody] CreateUserCommand command,
    [FromServices] IMediator mediator)
{
    var user = await mediator.Send(command);
    return CreatedAtAction(nameof(GetById), new { id = user.Id }, user);
}
```

### Blazor Patterns

Interactive Server Component:
```csharp
@page "/users"
@rendermode InteractiveServer
@inject IUserService UserService

<h1>Users</h1>

@if (_loading)
{
    <p>Loading...</p>
}
else if (_users is null)
{
    <p>No users found.</p>
}
else
{
    <table class="table">
        <thead>
            <tr><th>Name</th><th>Email</th><th>Actions</th></tr>
        </thead>
        <tbody>
            @foreach (var user in _users)
            {
                <tr>
                    <td>@user.Name</td>
                    <td>@user.Email</td>
                    <td>
                        <button @onclick="() => DeleteUser(user.Id)" class="btn btn-danger btn-sm">
                            Delete
                        </button>
                    </td>
                </tr>
            }
        </tbody>
    </table>
}

@code {
    private List<User>? _users;
    private bool _loading = true;

    protected override async Task OnInitializedAsync()
    {
        _users = await UserService.GetAllAsync();
        _loading = false;
    }

    private async Task DeleteUser(Guid id)
    {
        await UserService.DeleteAsync(id);
        _users = await UserService.GetAllAsync();
    }
}
```

### Authentication and Authorization

JWT Authentication Setup:
```csharp
builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(options =>
    {
        options.TokenValidationParameters = new TokenValidationParameters
        {
            ValidateIssuer = true,
            ValidateAudience = true,
            ValidateLifetime = true,
            ValidateIssuerSigningKey = true,
            ValidIssuer = builder.Configuration["Jwt:Issuer"],
            ValidAudience = builder.Configuration["Jwt:Audience"],
            IssuerSigningKey = new SymmetricSecurityKey(
                Encoding.UTF8.GetBytes(builder.Configuration["Jwt:Key"]!))
        };
    });

builder.Services.AddAuthorization(options =>
{
    options.AddPolicy("Admin", policy => policy.RequireRole("Admin"));
    options.AddPolicy("CanEdit", policy => policy.RequireClaim("permission", "edit"));
});
```

---

## Advanced Patterns

For comprehensive documentation including advanced patterns, performance optimization, testing strategies, and deployment configurations, see:

- [reference.md](reference.md) - Complete API reference, Context7 library mappings, advanced patterns
- [examples.md](examples.md) - Production-ready code examples, full-stack patterns, testing templates

### Context7 Integration

```csharp
// ASP.NET Core - mcp__context7__get_library_docs("/dotnet/aspnetcore", "minimal-apis middleware", 1)
// EF Core - mcp__context7__get_library_docs("/dotnet/efcore", "dbcontext migrations", 1)
// .NET Runtime - mcp__context7__get_library_docs("/dotnet/runtime", "collections threading", 1)
// Blazor - mcp__context7__get_library_docs("/dotnet/aspnetcore", "blazor components", 1)
```

---

## Works Well With

- `moai-domain-backend` - API design, database integration patterns
- `moai-platform-deploy` - Azure, Docker, Kubernetes deployment
- `moai-workflow-testing` - Testing strategies and patterns
- `moai-foundation-quality` - Code quality standards
- `moai-essentials-debug` - Debugging .NET applications

---

## Quick Troubleshooting

Build and Runtime:
```bash
dotnet build --verbosity detailed    # Detailed build output
dotnet run --launch-profile https    # Run with HTTPS profile
dotnet ef database update            # Apply EF migrations
dotnet ef migrations add Initial     # Create new migration
```

Common Issues:
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

Version: 1.0.0
Last Updated: 2025-12-07
Status: Production Ready
