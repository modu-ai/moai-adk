---
name: moai-lang-csharp
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "C# best practices with .NET 8, ASP.NET Core, Entity Framework, and modern async programming for 2025"
keywords: [csharp, programming, dotnet, aspnetcore, entityframework, backend, async]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# C# Development Mastery

**Modern C# Development with 2025 Best Practices**

> Comprehensive C# development guidance covering .NET 8 applications, ASP.NET Core APIs, Entity Framework Core with modern async/await patterns, and cross-platform development using the latest tools and frameworks.

## What It Does

### Backend Development
- **Web API Development**: ASP.NET Core with minimal APIs, controllers, and modern routing
- **Database Integration**: Entity Framework Core with LINQ, migrations, and performance optimization
- **Microservices**: gRPC, message queuing, distributed systems patterns
- **Real-time Communication**: SignalR, WebSockets with async/await
- **Testing**: xUnit, Moq, FluentAssertions with integration testing

### Cross-Platform Development
- **Desktop Applications**: WPF, MAUI, WinUI 3 for Windows and cross-platform
- **Mobile Applications**: .NET MAUI for iOS, Android, Windows
- **Console Applications**: Modern CLI apps with DI and configuration
- **Background Services**: .NET Core Workers, Hosted Services

### Cloud Integration
- **Azure Integration**: Azure Functions, App Service, Blob Storage
- **Docker & Kubernetes**: Containerization and orchestration
- **DevOps**: CI/CD pipelines, health checks, monitoring
- **Performance**: Profiling, optimization, memory management

## When to Use

### Perfect Scenarios
- **Building REST APIs and microservices with ASP.NET Core**
- **Developing enterprise applications with Entity Framework Core**
- **Creating cross-platform mobile apps with .NET MAUI**
- **Implementing real-time applications with SignalR**
- **Building cloud-native applications with Azure integration**
- **Developing high-performance backend services**
- **Creating modern desktop applications with WPF/MAUI**

### Common Triggers
- "Create C# web API"
- "Build ASP.NET Core application"
- "Set up Entity Framework Core"
- "Implement async/await patterns"
- "Optimize C# performance"
- "Test C# application"
- "C# best practices"

## Tool Version Matrix (2025-11-06)

### Core .NET
- **.NET**: 8.0 (current LTS) / 9.0 Preview
- **C#**: 12.0 (current) / 13.0 Preview
- **Package Managers**: NuGet, .NET CLI
- **Runtime**: .NET 8.0 LTS

### Web Frameworks
- **ASP.NET Core**: 8.0 - Web framework
- **Entity Framework Core**: 8.0 - ORM framework
- **Blazor**: 8.0 - Web UI framework
- **SignalR**: 8.0 - Real-time communication
- **gRPC**: 2.57.x - High-performance RPC

### Testing Tools
- **xUnit**: 2.6.x - Testing framework
- **Moq**: 4.20.x - Mocking framework
- **FluentAssertions**: 6.12.x - Assertion library
- **Bogus**: 35.5.x - Test data generation
- **Microsoft.AspNetCore.Mvc.Testing**: 8.0 - Integration testing

### Development Tools
- **Visual Studio 2022**: 17.10+
- **Visual Studio Code**: C# Dev Kit extension
- **Rider**: 2024.2+
- **.NET CLI**: 8.0.400+

### Database Tools
- **SQL Server**: 2022 / Azure SQL
- **PostgreSQL**: 16.x
- **SQLite**: 3.45.x
- **MongoDB**: 7.0.x (with MongoDB.Driver)

## Ecosystem Overview

### Package Management

```bash
# Create new projects
dotnet new webapi -n MyApi
dotnet new mvc -n MyMvcApp
dotnet new maui -n MyMauiApp
dotnet new worker -n MyWorkerService
dotnet new classlib -n MyLibrary

# Add packages
dotnet add package Microsoft.EntityFrameworkCore.SqlServer
dotnet add package Microsoft.Extensions.Caching.StackExchangeRedis
dotnet add package Swashbuckle.AspNetCore
dotnet add package xunit

# Build and run
dotnet build
dotnet run --project MyApi.csproj
dotnet test

# Global tools
dotnet tool install --global dotnet-ef
dotnet tool install --global dotnet-aspnet-codegenerator
```

### Project Structure (2025 Best Practice)

```
MyDotNetSolution/
├── src/
│   ├── MyApi/                    # Web API project
│   │   ├── Controllers/          # API controllers
│   │   ├── Endpoints/            # Minimal API endpoints
│   │   ├── Services/             # Business logic services
│   │   ├── Models/               # Data models and DTOs
│   │   ├── Data/                 # Data access layer
│   │   ├── Configuration/        # Configuration classes
│   │   ├── Filters/              # Action filters and middleware
│   │   └── Program.cs           # Application entry point
│   ├── MyCore/                  # Core business logic
│   │   ├── Entities/            # Domain entities
│   │   ├── Interfaces/          # Service interfaces
│   │   ├── ValueObjects/        # Value objects
│   │   └── Enums/               # Enumerations
│   ├── MyInfrastructure/        # Infrastructure concerns
│   │   ├── Persistence/         # Database implementations
│   │   ├── ExternalServices/    # External API clients
│   │   ├── Messaging/           # Message queue implementations
│   │   └── Caching/             # Cache implementations
│   └── MyTests/                 # Test projects
│       ├── Unit/                # Unit tests
│       ├── Integration/         # Integration tests
│       └── Functional/          # Functional tests
├── tests/                       # Additional test projects
├── docs/                        # Documentation
├── docker/                      # Docker configurations
├── .github/workflows/           # GitHub Actions
├── Directory.Build.props        # Solution-wide MSBuild properties
└── MySolution.sln               # Solution file
```

## Modern Development Patterns

### C# 12.0 Language Features

```csharp
// Primary constructors for classes
public class UserService(
    IUserRepository userRepository,
    IEmailService emailService,
    ILogger<UserService> logger) : IUserService
{
    // Fields are automatically created from constructor parameters
    public async Task<User> CreateUserAsync(CreateUserRequest request, CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Creating user with username: {Username}", request.Username);
        
        var user = new User(request.Username, request.Email);
        
        await userRepository.AddAsync(user, cancellationToken);
        await emailService.SendWelcomeEmailAsync(user.Email, cancellationToken);
        
        logger.LogInformation("User created successfully with ID: {UserId}", user.Id);
        return user;
    }
}

// Collection expressions
public class DataProcessor
{
    public int[] ProcessNumbers(IEnumerable<int> numbers)
    {
        return numbers
            .Where(n => n > 0)
            .OrderByDescending(n => n)
            .Take(10)
            .ToArray();
    }
    
    public List<string> GetDefaultPermissions()
    {
        return ["read", "write", "delete"]; // Collection expression
    }
}

// Required members
public class CreateUserRequest
{
    public required string Username { get; init; }
    public required string Email { get; init; }
    public string? FirstName { get; init; }
    public string? LastName { get; init; }
}

// Raw string literals
public class EmailService
{
    private readonly string _welcomeEmailTemplate = """
        Welcome to our platform!

        Hello {FirstName} {LastName},

        Thank you for registering with us. Your account has been created successfully.

        Best regards,
        The Team
        """;
    
    public async Task SendWelcomeEmailAsync(string email, string? firstName = null, string? lastName = null)
    {
        var emailBody = _welcomeEmailTemplate
            .Replace("{FirstName}", firstName ?? "User")
            .Replace("{LastName}", lastName ?? "");
        
        // Send email logic
    }
}

// Using aliases for numeric types
using Age = int;
using UserId = System.Guid;
using Price = decimal;

public class User
{
    public UserId Id { get; set; }
    public required string Username { get; set; }
    public Age Age { get; set; }
    public Price AccountBalance { get; set; }
}

// List patterns
public class NotificationService
{
    public string ProcessMessage(string[] messageParts)
    {
        return messageParts switch
        {
            [] => "Empty message",
            [var single] => $"Single part: {single}",
            [var first, var second] => $"Two parts: {first}, {second}",
            [var first, .. var middle, var last] => $"Multiple parts: {first} ... {last}",
            _ => "Complex message"
        };
    }
}
```

### ASP.NET Core 8.0 Minimal APIs

```csharp
var builder = WebApplication.CreateBuilder(args);

// Add services
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("DefaultConnection")));
builder.Services.AddScoped<IUserRepository, UserRepository>();
builder.Services.AddScoped<IUserService, UserService>();
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

builder.Services.AddRateLimiter(options =>
{
    options.AddPolicy("Default", context =>
        RateLimitPartition.GetSlidingWindowLimiter(
            partitionKey: context.Connection.RemoteIpAddress?.ToString() ?? "anonymous",
            factory: _ => new SlidingWindowRateLimiterOptions
            {
                PermitLimit = 100,
                Window = TimeSpan.FromMinutes(1),
                SegmentsPerWindow = 2
            }));
});

var app = builder.Build();

// Middleware pipeline
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseHttpsRedirection();
app.UseAuthentication();
app.UseAuthorization();
app.UseRateLimiter();

// API Endpoint Groups
var userGroup = app.MapGroup("/api/users")
    .RequireAuthorization()
    .AddEndpointFilter<ValidationFilter>();

// Minimal API endpoints
userGroup.MapGet("/", async (IUserService userService, CancellationToken cancellationToken) =>
{
    var users = await userService.GetAllUsersAsync(cancellationToken);
    return Results.Ok(users);
})
    .WithName("GetAllUsers")
    .WithOpenApi();

userGroup.MapGet("/{id:guid}", async (Guid id, IUserService userService, CancellationToken cancellationToken) =>
{
    var user = await userService.GetUserByIdAsync(id, cancellationToken);
    return user is not null ? Results.Ok(user) : Results.NotFound();
})
    .WithName("GetUserById")
    .WithOpenApi();

userGroup.MapPost("/", async (CreateUserRequest request, IUserService userService, 
    IValidator<CreateUserRequest> validator, CancellationToken cancellationToken) =>
{
    var validationResult = await validator.ValidateAsync(request, cancellationToken);
    if (!validationResult.IsValid)
    {
        return Results.ValidationProblem(validationResult.ToDictionary());
    }

    var user = await userService.CreateUserAsync(request, cancellationToken);
    return Results.Created($"/api/users/{user.Id}", user);
})
    .WithName("CreateUser")
    .WithOpenApi()
    .RequireRateLimiting("Default");

userGroup.MapPut("/{id:guid}", async (Guid id, UpdateUserRequest request, 
    IUserService userService, CancellationToken cancellationToken) =>
{
    var user = await userService.UpdateUserAsync(id, request, cancellationToken);
    return user is not null ? Results.Ok(user) : Results.NotFound();
})
    .WithName("UpdateUser")
    .WithOpenApi();

userGroup.MapDelete("/{id:guid}", async (Guid id, IUserService userService, 
    CancellationToken cancellationToken) =>
{
    var success = await userService.DeleteUserAsync(id, cancellationToken);
    return success ? Results.NoContent() : Results.NotFound();
})
    .WithName("DeleteUser")
    .WithOpenApi();

// Real-time endpoints
var chatGroup = app.MapGroup("/api/chat")
    .RequireAuthorization();

app.MapHub<ChatHub>("/chatHub");

app.Run();

// Custom endpoint filters
public class ValidationFilter : IEndpointFilter
{
    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        // Add global validation logic here
        return await next(context);
    }
}
```

### Entity Framework Core 8.0 Patterns

```csharp
// Domain entities with value objects
public class User
{
    public Guid Id { get; private set; }
    public string Username { get; private set; } = string.Empty;
    public Email Email { get; private set; } = null!;
    public UserProfile Profile { get; private set; } = null!;
    public IReadOnlyCollection<UserRole> Roles => _roles.AsReadOnly();
    
    private readonly List<UserRole> _roles = new();
    
    // Private constructor for EF Core
    private User() { }
    
    public User(string username, Email email, UserProfile profile)
    {
        Id = Guid.NewGuid();
        Username = username;
        Email = email;
        Profile = profile;
        CreatedAt = DateTime.UtcNow;
    }
    
    public void UpdateProfile(UserProfile newProfile)
    {
        Profile = newProfile;
        UpdatedAt = DateTime.UtcNow;
    }
    
    public void AddRole(UserRole role)
    {
        if (!_roles.Contains(role))
        {
            _roles.Add(role);
            UpdatedAt = DateTime.UtcNow;
        }
    }
    
    public DateTime CreatedAt { get; private set; }
    public DateTime? UpdatedAt { get; private set; }
}

// Value object
public readonly record struct Email(string Value)
{
    public static Email Create(string email)
    {
        if (!Regex.IsMatch(email, @"^[^@\s]+@[^@\s]+\.[^@\s]+$"))
            throw new ArgumentException("Invalid email format", nameof(email));
        
        return new Email(email.ToLowerInvariant());
    }
}

// EF Core configuration
public class AppDbContext : DbContext
{
    public AppDbContext(DbContextOptions<AppDbContext> options) : base(options) { }
    
    public DbSet<User> Users => Set<User>();
    public DbSet<UserRole> UserRoles => Set<UserRole>();
    public DbSet<Role> Roles => Set<Role>();
    
    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<User>(builder =>
        {
            builder.HasKey(u => u.Id);
            builder.Property(u => u.Username).IsRequired().HasMaxLength(50);
            builder.Property(u => u.Email)
                .HasConversion(
                    email => email.Value,
                    value => Email.Create(value))
                .IsRequired()
                .HasMaxLength(100);
            
            builder.OwnsOne(u => u.Profile, profile =>
            {
                profile.Property(p => p.FirstName).HasMaxLength(50);
                profile.Property(p => p.LastName).HasMaxLength(50);
                profile.Property(p => p.Bio).HasMaxLength(500);
            });
            
            builder.Property(u => u.CreatedAt).HasDefaultValueSql("GETUTCDATE()");
            
            builder.HasMany(u => u.Roles)
                .WithMany(r => r.Users)
                .UsingEntity<UserRole>(
                    join => join.HasOne<Role>().WithMany(),
                    join => join.HasOne<User>().WithMany());
        });
        
        modelBuilder.Entity<Role>(builder =>
        {
            builder.HasKey(r => r.Id);
            builder.Property(r => r.Name).IsRequired().HasMaxLength(50);
            
            builder.HasData(
                new Role { Id = 1, Name = "Admin" },
                new Role { Id = 2, Name = "User" },
                new Role { Id = 3, Name = "Moderator" });
        });
    }
    
    public override async Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        UpdateTimestamps();
        return await base.SaveChangesAsync(cancellationToken);
    }
    
    private void UpdateTimestamps()
    {
        var entries = ChangeTracker
            .Entries()
            .Where(e => e.Entity is User && (e.State == EntityState.Added || e.State == EntityState.Modified));
        
        foreach (var entry in entries)
        {
            var user = (User)entry.Entity;
            
            if (entry.State == EntityState.Added)
            {
                user.CreatedAt = DateTime.UtcNow;
            }
            else
            {
                user.UpdatedAt = DateTime.UtcNow;
            }
        }
    }
}

// Repository pattern with async/await
public interface IUserRepository
{
    Task<User?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task<User?> GetByUsernameAsync(string username, CancellationToken cancellationToken = default);
    Task<User?> GetByEmailAsync(Email email, CancellationToken cancellationToken = default);
    Task<IReadOnlyList<User>> GetAllAsync(int page = 1, int pageSize = 20, 
        CancellationToken cancellationToken = default);
    Task<User> AddAsync(User user, CancellationToken cancellationToken = default);
    Task<User> UpdateAsync(User user, CancellationToken cancellationToken = default);
    Task<bool> DeleteAsync(Guid id, CancellationToken cancellationToken = default);
    Task<bool> ExistsAsync(Guid id, CancellationToken cancellationToken = default);
}

public class UserRepository : IUserRepository
{
    private readonly AppDbContext _context;
    
    public UserRepository(AppDbContext context)
    {
        _context = context;
    }
    
    public async Task<User?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.Users
            .Include(u => u.Roles)
            .FirstOrDefaultAsync(u => u.Id == id, cancellationToken);
    }
    
    public async Task<User?> GetByUsernameAsync(string username, CancellationToken cancellationToken = default)
    {
        return await _context.Users
            .Include(u => u.Roles)
            .FirstOrDefaultAsync(u => u.Username == username, cancellationToken);
    }
    
    public async Task<User?> GetByEmailAsync(Email email, CancellationToken cancellationToken = default)
    {
        return await _context.Users
            .Include(u => u.Roles)
            .FirstOrDefaultAsync(u => u.Email == email, cancellationToken);
    }
    
    public async Task<IReadOnlyList<User>> GetAllAsync(int page = 1, int pageSize = 20, 
        CancellationToken cancellationToken = default)
    {
        return await _context.Users
            .Include(u => u.Roles)
            .OrderBy(u => u.Username)
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .AsNoTracking()
            .ToListAsync(cancellationToken);
    }
    
    public async Task<User> AddAsync(User user, CancellationToken cancellationToken = default)
    {
        await _context.Users.AddAsync(user, cancellationToken);
        await _context.SaveChangesAsync(cancellationToken);
        return user;
    }
    
    public async Task<User> UpdateAsync(User user, CancellationToken cancellationToken = default)
    {
        _context.Users.Update(user);
        await _context.SaveChangesAsync(cancellationToken);
        return user;
    }
    
    public async Task<bool> DeleteAsync(Guid id, CancellationToken cancellationToken = default)
    {
        var user = await GetByIdAsync(id, cancellationToken);
        if (user is null) return false;
        
        _context.Users.Remove(user);
        await _context.SaveChangesAsync(cancellationToken);
        return true;
    }
    
    public async Task<bool> ExistsAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.Users.AnyAsync(u => u.Id == id, cancellationToken);
    }
}
```

### Modern Async Patterns

```csharp
// Service layer with proper async handling
public class UserService : IUserService
{
    private readonly IUserRepository _userRepository;
    private readonly IEmailService _emailService;
    private readonly ILogger<UserService> _logger;
    private readonly IMemoryCache _cache;
    
    public UserService(
        IUserRepository userRepository,
        IEmailService emailService,
        ILogger<UserService> logger,
        IMemoryCache cache)
    {
        _userRepository = userRepository;
        _emailService = emailService;
        _logger = logger;
        _cache = cache;
    }
    
    public async Task<User> CreateUserAsync(CreateUserRequest request, 
        CancellationToken cancellationToken = default)
    {
        // Validate input
        var email = Email.Create(request.Email);
        
        // Check for existing user
        var existingUser = await _userRepository.GetByEmailAsync(email, cancellationToken);
        if (existingUser is not null)
        {
            throw new UserAlreadyExistsException($"User with email {email.Value} already exists");
        }
        
        // Create new user
        var profile = new UserProfile(request.FirstName, request.LastName, request.Bio);
        var user = new User(request.Username, email, profile);
        
        // Save to database
        var createdUser = await _userRepository.AddAsync(user, cancellationToken);
        
        // Send welcome email (fire and forget)
        _ = Task.Run(async () =>
        {
            try
            {
                await _emailService.SendWelcomeEmailAsync(email, profile.FirstName, profile.LastName);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to send welcome email to {Email}", email.Value);
            }
        });
        
        // Cache the user
        _cache.Set($"user_{createdUser.Id}", createdUser, TimeSpan.FromMinutes(30));
        
        return createdUser;
    }
    
    public async Task<User?> GetUserByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        // Try cache first
        if (_cache.TryGetValue($"user_{id}", out User? cachedUser))
        {
            return cachedUser;
        }
        
        // Fetch from database
        var user = await _userRepository.GetByIdAsync(id, cancellationToken);
        
        // Cache if found
        if (user is not null)
        {
            _cache.Set($"user_{id}", user, TimeSpan.FromMinutes(30));
        }
        
        return user;
    }
    
    public async Task<IReadOnlyList<User>> GetAllUsersAsync(int page = 1, int pageSize = 20,
        CancellationToken cancellationToken = default)
    {
        return await _userRepository.GetAllAsync(page, pageSize, cancellationToken);
    }
    
    public async Task<User?> UpdateUserAsync(Guid id, UpdateUserRequest request,
        CancellationToken cancellationToken = default)
    {
        var user = await _userRepository.GetByIdAsync(id, cancellationToken);
        if (user is null) return null;
        
        var updatedProfile = new UserProfile(
            request.FirstName ?? user.Profile.FirstName,
            request.LastName ?? user.Profile.LastName,
            request.Bio ?? user.Profile.Bio);
        
        user.UpdateProfile(updatedProfile);
        
        var updatedUser = await _userRepository.UpdateAsync(user, cancellationToken);
        
        // Update cache
        _cache.Set($"user_{id}", updatedUser, TimeSpan.FromMinutes(30));
        
        return updatedUser;
    }
    
    public async Task<bool> DeleteUserAsync(Guid id, CancellationToken cancellationToken = default)
    {
        var success = await _userRepository.DeleteAsync(id, cancellationToken);
        
        if (success)
        {
            // Remove from cache
            _cache.Remove($"user_{id}");
        }
        
        return success;
    }
    
    // Parallel operations example
    public async Task<UserProfileSummary> GetUserProfileSummaryAsync(Guid userId,
        CancellationToken cancellationToken = default)
    {
        var userTask = GetUserByIdAsync(userId, cancellationToken);
        var postsTask = GetUserPostsCountAsync(userId, cancellationToken);
        var followersTask = GetUserFollowersCountAsync(userId, cancellationToken);
        
        await Task.WhenAll(userTask, postsTask, followersTask);
        
        var user = await userTask;
        if (user is null) throw new UserNotFoundException(userId);
        
        var postsCount = await postsTask;
        var followersCount = await followersTask;
        
        return new UserProfileSummary
        {
            User = user,
            PostsCount = postsCount,
            FollowersCount = followersCount
        };
    }
    
    private async Task<int> GetUserPostsCountAsync(Guid userId, CancellationToken cancellationToken = default)
    {
        // Simulate external API call
        await Task.Delay(100, cancellationToken);
        return Random.Shared.Next(0, 100);
    }
    
    private async Task<int> GetUserFollowersCountAsync(Guid userId, CancellationToken cancellationToken = default)
    {
        // Simulate external API call
        await Task.Delay(150, cancellationToken);
        return Random.Shared.Next(0, 1000);
    }
}

// Background service with async patterns
public class UserCleanupService : BackgroundService
{
    private readonly ILogger<UserCleanupService> _logger;
    private readonly IUserRepository _userRepository;
    private readonly TimeSpan _cleanupInterval = TimeSpan.FromHours(1);
    
    public UserCleanupService(
        ILogger<UserCleanupService> logger,
        IUserRepository userRepository)
    {
        _logger = logger;
        _userRepository = userRepository;
    }
    
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                _logger.LogInformation("Starting user cleanup process");
                
                await CleanupInactiveUsersAsync(stoppingToken);
                
                _logger.LogInformation("User cleanup process completed");
            }
            catch (OperationCanceledException)
            {
                _logger.LogInformation("User cleanup service was cancelled");
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error during user cleanup process");
            }
            
            await Task.Delay(_cleanupInterval, stoppingToken);
        }
    }
    
    private async Task CleanupInactiveUsersAsync(CancellationToken cancellationToken)
    {
        const int batchSize = 100;
        var cutoffDate = DateTime.UtcNow.AddDays(-365);
        
        var inactiveUsers = await _userRepository.GetInactiveUsersBeforeAsync(cutoffDate, cancellationToken);
        
        foreach (var batch in inactiveUsers.Chunk(batchSize))
        {
            await Parallel.ForEachAsync(batch, cancellationToken, async (user, ct) =>
            {
                try
                {
                    await _userRepository.DeleteAsync(user.Id, ct);
                    _logger.LogInformation("Deleted inactive user {UserId} ({Username})", user.Id, user.Username);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "Failed to delete user {UserId}", user.Id);
                }
            });
        }
    }
}
```

## Performance Considerations

### Memory Management

```csharp
// Efficient data structures and memory usage
public class DataProcessor
{
    // Use Span<T> for zero-allocation string processing
    public static bool IsValidEmail(ReadOnlySpan<char> email)
    {
        var atIdx = email.IndexOf('@');
        if (atIdx <= 0 || atIdx == email.Length - 1) return false;
        
        var dotIdx = email.LastIndexOf('.');
        if (dotIdx <= atIdx + 1 || dotIdx == email.Length - 1) return false;
        
        return true;
    }
    
    // Use ArrayPool for temporary arrays
    public static byte[] ProcessLargeData(ReadOnlySpan<byte> data)
    {
        var buffer = ArrayPool<byte>.Shared.Rent(data.Length * 2);
        
        try
        {
            // Process data into buffer
            var processedLength = ProcessDataInternal(data, buffer);
            
            var result = new byte[processedLength];
            buffer.AsSpan(0, processedLength).CopyTo(result);
            
            return result;
        }
        finally
        {
            ArrayPool<byte>.Shared.Return(buffer);
        }
    }
    
    private static int ProcessDataInternal(ReadOnlySpan<byte> source, Span<byte> destination)
    {
        // Processing logic
        return Math.Min(source.Length * 2, destination.Length);
    }
}

// Efficient LINQ usage with immediate execution when needed
public class UserRepository
{
    private readonly AppDbContext _context;
    
    public UserRepository(AppDbContext context)
    {
        _context = context;
    }
    
    // Use AsNoTracking for read-only queries
    public async Task<IReadOnlyList<User>> GetActiveUsersAsync(int page = 1, int pageSize = 20,
        CancellationToken cancellationToken = default)
    {
        return await _context.Users
            .AsNoTracking() // No change tracking for read-only
            .Where(u => u.IsActive)
            .OrderBy(u => u.Username)
            .Select(u => new UserDto // Project only needed fields
            {
                Id = u.Id,
                Username = u.Username,
                Email = u.Email.Value,
                CreatedAt = u.CreatedAt
            })
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync(cancellationToken);
    }
    
    // Use compiled queries for frequently executed queries
    private static readonly Func<AppDbContext, Guid, Task<User?>> GetUserByIdCompiled =
        EF.CompileAsyncQuery((AppDbContext context, Guid id) =>
            context.Users
                .Include(u => u.Roles)
                .FirstOrDefaultAsync(u => u.Id == id));
    
    public async Task<User?> GetUserByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await GetUserByIdCompiled(_context, id);
    }
    
    // Batch operations for better performance
    public async Task UpdateUsersLastLoginAsync(IEnumerable<Guid> userIds,
        CancellationToken cancellationToken = default)
    {
        const int batchSize = 1000;
        
        foreach (var batch in userIds.Chunk(batchSize))
        {
            await _context.Users
                .Where(u => batch.Contains(u.Id))
                .ExecuteUpdateAsync(setters => setters
                    .SetProperty(u => u.LastLoginAt, DateTime.UtcNow), cancellationToken);
        }
    }
}

// Memory-efficient caching
public class CacheService
{
    private readonly IMemoryCache _cache;
    private readonly Timer _cleanupTimer;
    
    public CacheService(IMemoryCache cache)
    {
        _cache = cache;
        
        // Set up periodic cleanup
        _cleanupTimer = new Timer(CleanupExpiredEntries, null, 
            TimeSpan.FromMinutes(5), TimeSpan.FromMinutes(5));
    }
    
    public async Task<T?> GetOrCreateAsync<T>(string key, Func<Task<T>> factory,
        TimeSpan? absoluteExpiration = null, TimeSpan? slidingExpiration = null)
    {
        if (_cache.TryGetValue(key, out T? cachedValue))
        {
            return cachedValue;
        }
        
        var value = await factory();
        
        var options = new MemoryCacheEntryOptions
        {
            AbsoluteExpirationRelativeToNow = absoluteExpiration,
            SlidingExpiration = slidingExpiration,
            Size = 1 // Enable size-based eviction
        };
        
        _cache.Set(key, value, options);
        return value;
    }
    
    private void CleanupExpiredEntries(object? state)
    {
        // MemoryCache automatically removes expired entries, but we can
        // implement additional cleanup logic here if needed
        _cache.Compact(0.95); // Remove 5% of least recently used items
    }
    
    public void Remove(string key)
    {
        _cache.Remove(key);
    }
    
    public void Clear()
    {
        _cache.Compact(1.0); // Remove all entries
    }
}
```

### Database Performance

```csharp
// Optimized database operations
public class OptimizedDbContext : DbContext
{
    private readonly ILogger<OptimizedDbContext> _logger;
    
    public OptimizedDbContext(DbContextOptions<OptimizedDbContext> options,
        ILogger<OptimizedDbContext> logger) : base(options)
    {
        _logger = logger;
    }
    
    // Use connection resiliency
    protected override void OnConfiguring(DbContextOptionsBuilder optionsBuilder)
    {
        optionsBuilder.EnableRetryOnFailure(
            maxRetryCount: 3,
            maxRetryDelay: TimeSpan.FromSeconds(30),
            errorNumbersToAdd: null);
        
        // Configure sensitive data logging for development
        optionsBuilder.EnableSensitiveDataLogging(false);
        optionsBuilder.EnableDetailedErrors(false);
        
        // Set command timeout
        optionsBuilder.UseSqlServer(sqlOptions =>
        {
            sqlOptions.CommandTimeout(30);
            sqlOptions.EnableRetryOnFailure();
        });
    }
    
    // Batch operations
    public async Task BulkInsertAsync<T>(IEnumerable<T> entities, 
        CancellationToken cancellationToken = default) where T : class
    {
        const int batchSize = 1000;
        
        foreach (var batch in entities.Chunk(batchSize))
        {
            await Set<T>().AddRangeAsync(batch, cancellationToken);
            await SaveChangesAsync(cancellationToken);
            
            // Clear change tracker for memory efficiency
            ChangeTracker.Clear();
        }
    }
    
    // Optimized queries with proper indexing hints
    public async Task<IReadOnlyList<User>> SearchUsersAsync(string searchTerm,
        int page = 1, int pageSize = 20, CancellationToken cancellationToken = default)
    {
        var normalizedSearch = searchTerm.Trim().ToUpperInvariant();
        
        return await Users
            .AsNoTracking()
            .Where(u => EF.Functions.ILike(u.Username, $"%{normalizedSearch}%") ||
                       EF.Functions.ILike(u.Profile.FirstName, $"%{normalizedSearch}%") ||
                       EF.Functions.ILike(u.Profile.LastName, $"%{normalizedSearch}%"))
            .OrderBy(u => u.Username)
            .Select(u => new UserDto // Project to DTO for less data transfer
            {
                Id = u.Id,
                Username = u.Username,
                Email = u.Email.Value,
                FullName = $"{u.Profile.FirstName} {u.Profile.LastName}",
                CreatedAt = u.CreatedAt
            })
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync(cancellationToken);
    }
    
    // Use raw SQL for performance-critical operations
    public async Task<int> DeleteOldRecordsAsync(DateTime cutoffDate,
        CancellationToken cancellationToken = default)
    {
        var sql = """
            DELETE FROM UserLogs 
            WHERE CreatedAt < @cutoffDate
            """;
        
        return await Database.ExecuteSqlRawAsync(sql,
            new SqlParameter("@cutoffDate", cutoffDate));
    }
    
    public override async Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            return await base.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateConcurrencyException ex)
        {
            _logger.LogWarning(ex, "Concurrency conflict detected");
            throw new ConcurrencyException("Data was modified by another user", ex);
        }
        catch (DbUpdateException ex)
        {
            _logger.LogError(ex, "Database update failed");
            throw new DataUpdateException("Failed to save changes to database", ex);
        }
    }
}

// Connection pooling and transaction management
public class UnitOfWork : IUnitOfWork, IDisposable
{
    private readonly OptimizedDbContext _context;
    private IDbContextTransaction? _transaction;
    
    public UnitOfWork(OptimizedDbContext context)
    {
        _context = context;
    }
    
    public async Task BeginTransactionAsync(IsolationLevel isolationLevel = IsolationLevel.ReadCommitted,
        CancellationToken cancellationToken = default)
    {
        _transaction = await _context.Database.BeginTransactionAsync(isolationLevel, cancellationToken);
    }
    
    public async Task CommitAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            await _context.SaveChangesAsync(cancellationToken);
            await _transaction?.CommitAsync(cancellationToken)!;
        }
        catch
        {
            await RollbackAsync(cancellationToken);
            throw;
        }
    }
    
    public async Task RollbackAsync(CancellationToken cancellationToken = default)
    {
        if (_transaction != null)
        {
            await _transaction.RollbackAsync(cancellationToken);
        }
    }
    
    public void Dispose()
    {
        _transaction?.Dispose();
        _context.Dispose();
    }
}
```

### HTTP Client Performance

```csharp
// Optimized HTTP client with connection pooling and circuit breaker
public class OptimizedHttpClient
{
    private readonly HttpClient _httpClient;
    private readonly ILogger<OptimizedHttpClient> _logger;
    
    public OptimizedHttpClient(IHttpClientFactory httpClientFactory,
        ILogger<OptimizedHttpClient> logger)
    {
        _httpClient = httpClientFactory.CreateClient("OptimizedClient");
        _logger = logger;
    }
    
    public async Task<T?> GetAsync<T>(string uri, CancellationToken cancellationToken = default)
    {
        try
        {
            var response = await _httpClient.GetAsync(uri, cancellationToken);
            response.EnsureSuccessStatusCode();
            
            return await response.Content.ReadFromJsonAsync<T>(cancellationToken: cancellationToken);
        }
        catch (HttpRequestException ex)
        {
            _logger.LogError(ex, "HTTP request failed for URI: {Uri}", uri);
            throw;
        }
    }
    
    public async Task<T?> PostAsync<T>(string uri, T data, CancellationToken cancellationToken = default)
    {
        try
        {
            var response = await _httpClient.PostAsJsonAsync(uri, data, cancellationToken);
            response.EnsureSuccessStatusCode();
            
            return await response.Content.ReadFromJsonAsync<T>(cancellationToken: cancellationToken);
        }
        catch (HttpRequestException ex)
        {
            _logger.LogError(ex, "HTTP POST request failed for URI: {Uri}", uri);
            throw;
        }
    }
}

// HTTP client configuration in Program.cs
builder.Services.AddHttpClient("OptimizedClient", client =>
{
    client.BaseAddress = new Uri("https://api.example.com/");
    client.Timeout = TimeSpan.FromSeconds(30);
    client.DefaultRequestHeaders.Add("User-Agent", "MyApp/1.0");
})
.ConfigurePrimaryHttpMessageHandler(() => new SocketsHttpHandler
{
    PooledConnectionLifetime = TimeSpan.FromMinutes(5),
    PooledConnectionIdleTimeout = TimeSpan.FromMinutes(1),
    MaxConnectionsPerServer = 100
})
.AddPolicyHandler(GetRetryPolicy())
.AddPolicyHandler(GetCircuitBreakerPolicy());

static IAsyncPolicy<HttpResponseMessage> GetRetryPolicy()
{
    return HttpPolicyExtensions
        .HandleTransientHttpError()
        .OrResult(msg => msg.StatusCode == System.Net.HttpStatusCode.TooManyRequests)
        .WaitAndRetryAsync(3, retryAttempt => TimeSpan.FromSeconds(Math.Pow(2, retryAttempt)));
}

static IAsyncPolicy<HttpResponseMessage> GetCircuitBreakerPolicy()
{
    return HttpPolicyExtensions
        .HandleTransientHttpError()
        .CircuitBreakerAsync(5, TimeSpan.FromSeconds(30));
}
```

## Testing Strategy

### xUnit Configuration

```csharp
// Test project setup with dependency injection
public class TestFixture : IDisposable
{
    public ServiceProvider ServiceProvider { get; }
    public AppDbContext DbContext { get; }
    
    public TestFixture()
    {
        var services = new ServiceCollection();
        
        // Configure in-memory database
        services.AddDbContext<AppDbContext>(options =>
            options.UseInMemoryDatabase(Guid.NewGuid().ToString()));
        
        // Register services
        services.AddScoped<IUserRepository, UserRepository>();
        services.AddScoped<IUserService, UserService>();
        services.AddScoped<IEmailService, MockEmailService>();
        
        // Configure logging
        services.AddLogging(builder =>
            builder.SetMinimumLevel(LogLevel.Warning));
        
        ServiceProvider = services.BuildServiceProvider();
        DbContext = ServiceProvider.GetRequiredService<AppDbContext>();
        
        // Seed test data
        SeedTestData();
    }
    
    private void SeedTestData()
    {
        var users = new List<User>
        {
            new User("admin", Email.Create("admin@example.com"),
                new UserProfile("Admin", "User", "System administrator")),
            new User("user1", Email.Create("user1@example.com"),
                new UserProfile("Regular", "User", "Regular user"))
        };
        
        DbContext.Users.AddRange(users);
        DbContext.SaveChanges();
    }
    
    public void Dispose()
    {
        DbContext.Dispose();
        ServiceProvider.Dispose();
    }
}

// Integration tests with WebApplicationFactory
public class UserIntegrationTests : IClassFixture<WebApplicationFactory<Program>>
{
    private readonly WebApplicationFactory<Program> _factory;
    private readonly HttpClient _client;
    
    public UserIntegrationTests(WebApplicationFactory<Program> factory)
    {
        _factory = factory.WithWebHostBuilder(builder =>
        {
            builder.ConfigureServices(services =>
            {
                // Configure test services
                services.Remove(services.SingleOrDefault(
                    d => d.ServiceType == typeof(DbContextOptions<AppDbContext>)));
                
                services.AddDbContext<AppDbContext>(options =>
                    options.UseInMemoryDatabase("TestDb"));
                
                // Replace email service with mock
                services.AddScoped<IEmailService, MockEmailService>();
            });
        });
        
        _client = _factory.CreateClient();
    }
    
    [Fact]
    public async Task CreateUser_ValidInput_ReturnsCreated()
    {
        // Arrange
        var request = new CreateUserRequest
        {
            Username = "newuser",
            Email = "newuser@example.com",
            FirstName = "New",
            LastName = "User"
        };
        
        // Act
        var response = await _client.PostAsJsonAsync("/api/users", request);
        
        // Assert
        response.EnsureSuccessStatusCode();
        
        var user = await response.Content.ReadFromJsonAsync<UserDto>();
        Assert.NotNull(user);
        Assert.Equal("newuser", user!.Username);
        Assert.Equal("newuser@example.com", user.Email);
    }
    
    [Fact]
    public async Task GetUsers_ReturnsUserList()
    {
        // Act
        var response = await _client.GetAsync("/api/users");
        
        // Assert
        response.EnsureSuccessStatusCode();
        
        var users = await response.Content.ReadFromJsonAsync<List<UserDto>>();
        Assert.NotNull(users);
        Assert.True(users!.Count >= 2); // At least the seeded users
    }
    
    [Fact]
    public async Task GetUser_ExistingId_ReturnsUser()
    {
        // Arrange
        var createResponse = await _client.PostAsJsonAsync("/api/users", new CreateUserRequest
        {
            Username = "testuser",
            Email = "test@example.com",
            FirstName = "Test",
            LastName = "User"
        });
        
        var createdUser = await createResponse.Content.ReadFromJsonAsync<UserDto>();
        Assert.NotNull(createdUser);
        
        // Act
        var response = await _client.GetAsync($"/api/users/{createdUser!.Id}");
        
        // Assert
        response.EnsureSuccessStatusCode();
        
        var user = await response.Content.ReadFromJsonAsync<UserDto>();
        Assert.NotNull(user);
        Assert.Equal(createdUser!.Id, user!.Id);
    }
}
```

### Unit Testing Patterns

```csharp
public class UserServiceTests : IClassFixture<TestFixture>
{
    private readonly TestFixture _fixture;
    private readonly IUserService _userService;
    private readonly Mock<IEmailService> _mockEmailService;
    
    public UserServiceTests(TestFixture fixture)
    {
        _fixture = fixture;
        _userService = fixture.ServiceProvider.GetRequiredService<IUserService>();
        _mockEmailService = new Mock<IEmailService>();
        _userService = new UserService(
            _fixture.ServiceProvider.GetRequiredService<IUserRepository>(),
            _mockEmailService.Object,
            Mock.Of<ILogger<UserService>>(),
            Mock.Of<IMemoryCache>());
    }
    
    [Fact]
    public async Task CreateUser_ValidRequest_ReturnsUser()
    {
        // Arrange
        var request = new CreateUserRequest
        {
            Username = "newuser",
            Email = "newuser@example.com",
            FirstName = "New",
            LastName = "User"
        };
        
        _mockEmailService
            .Setup(s => s.SendWelcomeEmailAsync(It.IsAny<Email>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.CompletedTask);
        
        // Act
        var result = await _userService.CreateUserAsync(request);
        
        // Assert
        Assert.NotNull(result);
        Assert.Equal("newuser", result.Username);
        Assert.Equal("newuser@example.com", result.Email.Value);
        
        _mockEmailService.Verify(
            s => s.SendWelcomeEmailAsync(It.Is<Email>(e => e.Value == "newuser@example.com"), "New", "User"),
            Times.Once);
    }
    
    [Fact]
    public async Task CreateUser_DuplicateEmail_ThrowsException()
    {
        // Arrange
        var request = new CreateUserRequest
        {
            Username = "duplicate",
            Email = "admin@example.com", // Already exists in seed data
            FirstName = "Duplicate",
            LastName = "User"
        };
        
        // Act & Assert
        var exception = await Assert.ThrowsAsync<UserAlreadyExistsException>(
            () => _userService.CreateUserAsync(request));
        
        Assert.Contains("already exists", exception.Message);
        
        _mockEmailService.Verify(
            s => s.SendWelcomeEmailAsync(It.IsAny<Email>(), It.IsAny<string>(), It.IsAny<string>()),
            Times.Never);
    }
    
    [Theory]
    [InlineData("test@example.com", true)]
    [InlineData("invalid-email", false)]
    [InlineData("test@.com", false)]
    [InlineData("@example.com", false)]
    public void EmailValidation_ValidatesEmailCorrectly(string email, bool expectedValid)
    {
        // Act
        var isValid = EmailValidator.IsValid(email);
        
        // Assert
        Assert.Equal(expectedValid, isValid);
    }
    
    [Fact]
    public async Task UpdateUser_ValidRequest_UpdatesUser()
    {
        // Arrange
        var existingUser = await _userService.GetUserByIdAsync(Guid.Parse("00000000-0000-0000-0000-000000000001"));
        Assert.NotNull(existingUser);
        
        var updateRequest = new UpdateUserRequest
        {
            FirstName = "Updated",
            LastName = "Name",
            Bio = "Updated bio"
        };
        
        // Act
        var result = await _userService.UpdateUserAsync(existingUser!.Id, updateRequest);
        
        // Assert
        Assert.NotNull(result);
        Assert.Equal("Updated", result!.Profile.FirstName);
        Assert.Equal("Name", result.Profile.LastName);
        Assert.Equal("Updated bio", result.Profile.Bio);
    }
}

// Performance testing
public class PerformanceTests : IClassFixture<TestFixture>
{
    private readonly TestFixture _fixture;
    
    public PerformanceTests(TestFixture fixture)
    {
        _fixture = fixture;
    }
    
    [Fact]
    public async Task BulkInsert_PerformanceTest()
    {
        // Arrange
        const int userCount = 1000;
        var users = Enumerable.Range(1, userCount)
            .Select(i => new User($"user{i}", Email.Create($"user{i}@example.com"),
                new UserProfile($"User{i}", $"Test{i}", $"Bio for user {i}")))
            .ToList();
        
        var stopwatch = Stopwatch.StartNew();
        
        // Act
        await _fixture.DbContext.BulkInsertAsync(users);
        
        stopwatch.Stop();
        
        // Assert
        Assert.True(stopwatch.ElapsedMilliseconds < 5000, // Should complete within 5 seconds
            $"Bulk insert took {stopwatch.ElapsedMilliseconds}ms, expected less than 5000ms");
        
        var dbUsers = await _fixture.DbContext.Users.CountAsync();
        Assert.True(dbUsers >= userCount, $"Expected at least {userCount} users, got {dbUsers}");
    }
}
```

### API Testing

```csharp
// API contract testing
public class ApiContractTests : IClassFixture<WebApplicationFactory<Program>>
{
    private readonly WebApplicationFactory<Program> _factory;
    private readonly HttpClient _client;
    
    public ApiContractTests(WebApplicationFactory<Program> factory)
    {
        _factory = factory;
        _client = _factory.CreateClient();
    }
    
    [Fact]
    public async Task GetUsers_ReturnsCorrectContract()
    {
        // Act
        var response = await _client.GetAsync("/api/users");
        
        // Assert
        response.EnsureSuccessStatusCode();
        Assert.Equal("application/json; charset=utf-8", response.Content.Headers.ContentType?.ToString());
        
        var content = await response.Content.ReadAsStringAsync();
        Assert.NotEmpty(content);
        
        // Validate JSON schema
        var users = JsonSerializer.Deserialize<List<UserDto>>(content);
        Assert.NotNull(users);
    }
    
    [Fact]
    public async Task CreateUser_ValidInput_ReturnsCorrectContract()
    {
        // Arrange
        var request = new
        {
            Username = "newuser",
            Email = "newuser@example.com",
            FirstName = "New",
            LastName = "User"
        };
        
        // Act
        var response = await _client.PostAsJsonAsync("/api/users", request);
        
        // Assert
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        Assert.Equal("application/json; charset=utf-8", response.Content.Headers.ContentType?.ToString());
        
        var content = await response.Content.ReadAsStringAsync();
        var user = JsonSerializer.Deserialize<UserDto>(content, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
        
        Assert.NotNull(user);
        Assert.Equal("newuser", user!.Username);
        Assert.Equal("newuser@example.com", user.Email);
        Assert.NotNull(user.Id);
    }
    
    [Fact]
    public async Task GetUsers_WithoutAuthentication_ReturnsUnauthorized()
    {
        // Act
        var response = await _client.GetAsync("/api/users/protected-endpoint");
        
        // Assert
        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }
}
```

## Security Best Practices

### Input Validation and Sanitization

```csharp
// Data annotations for validation
public class CreateUserRequest
{
    [Required(ErrorMessage = "Username is required")]
    [StringLength(50, MinimumLength = 3, ErrorMessage = "Username must be between 3 and 50 characters")]
    [RegularExpression(@"^[a-zA-Z0-9_]+$", ErrorMessage = "Username can only contain letters, numbers, and underscores")]
    public string Username { get; init; } = string.Empty;
    
    [Required(ErrorMessage = "Email is required")]
    [EmailAddress(ErrorMessage = "Invalid email format")]
    [StringLength(100, ErrorMessage = "Email must be less than 100 characters")]
    public string Email { get; init; } = string.Empty;
    
    [StringLength(50, ErrorMessage = "First name must be less than 50 characters")]
    [RegularExpression(@"^[a-zA-Z\s]+$", ErrorMessage = "First name can only contain letters and spaces")]
    public string? FirstName { get; init; }
    
    [StringLength(50, ErrorMessage = "Last name must be less than 50 characters")]
    [RegularExpression(@"^[a-zA-Z\s]+$", ErrorMessage = "Last name can only contain letters and spaces")]
    public string? LastName { get; init; }
    
    [StringLength(500, ErrorMessage = "Bio must be less than 500 characters")]
    [RegularExpression(@"^[^<>]*$", ErrorMessage = "Bio cannot contain HTML tags")]
    public string? Bio { get; init; }
}

// Custom validation attributes
public class StrongPasswordAttribute : ValidationAttribute
{
    public StrongPasswordAttribute() : base("Password must be at least 8 characters and contain uppercase, lowercase, digit, and special character")
    {
    }
    
    protected override ValidationResult? IsValid(object? value, ValidationContext validationContext)
    {
        if (value is string password)
        {
            if (password.Length < 8)
            {
                return new ValidationResult("Password must be at least 8 characters long");
            }
            
            if (!Regex.IsMatch(password, @"[A-Z]"))
            {
                return new ValidationResult("Password must contain at least one uppercase letter");
            }
            
            if (!Regex.IsMatch(password, @"[a-z]"))
            {
                return new ValidationResult("Password must contain at least one lowercase letter");
            }
            
            if (!Regex.IsMatch(password, @"\d"))
            {
                return new ValidationResult("Password must contain at least one digit");
            }
            
            if (!Regex.IsMatch(password, @"[!@#$%^&*(),.?""{}|<>]"))
            {
                return new ValidationResult("Password must contain at least one special character");
            }
            
            return ValidationResult.Success;
        }
        
        return new ValidationResult("Password is required");
    }
}

// HTML sanitization
public static class HtmlSanitizer
{
    private static readonly HtmlSanitizer _sanitizer = new HtmlSanitizer()
        .AllowTags("p", "br", "strong", "em", "u", "ol", "ul", "li")
        .AllowAttributes("class").OnAllTags()
        .RemoveAttributes("style", "onclick", "onload");
    
    public static string Sanitize(string? html)
    {
        if (string.IsNullOrEmpty(html))
        {
            return string.Empty;
        }
        
        return _sanitizer.Sanitize(html);
    }
}
```

### Authentication and Authorization

```csharp
// JWT authentication configuration
public static class AuthenticationConfiguration
{
    public static IServiceCollection AddJwtAuthentication(this IServiceCollection services,
        IConfiguration configuration)
    {
        var jwtSettings = configuration.GetSection("JwtSettings");
        var key = Encoding.UTF8.GetBytes(jwtSettings["Secret"]!);
        
        services.AddAuthentication(options =>
        {
            options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
            options.DefaultChallengeScheme = JwtBearerDefaults.AuthenticationScheme;
        })
        .AddJwtBearer(options =>
        {
            options.TokenValidationParameters = new TokenValidationParameters
            {
                ValidateIssuer = true,
                ValidateAudience = true,
                ValidateLifetime = true,
                ValidateIssuerSigningKey = true,
                ValidIssuer = jwtSettings["Issuer"],
                ValidAudience = jwtSettings["Audience"],
                IssuerSigningKey = new SymmetricSecurityKey(key),
                ClockSkew = TimeSpan.Zero
            };
        });
        
        return services;
    }
}

// JWT service
public interface IJwtService
{
    string GenerateToken(User user);
    ClaimsPrincipal? GetPrincipalFromToken(string token);
}

public class JwtService : IJwtService
{
    private readonly IConfiguration _configuration;
    private readonly string _secret;
    private readonly string _issuer;
    private readonly string _audience;
    
    public JwtService(IConfiguration configuration)
    {
        _configuration = configuration;
        _secret = _configuration["JwtSettings:Secret"]!;
        _issuer = _configuration["JwtSettings:Issuer"]!;
        _audience = _configuration["JwtSettings:Audience"]!;
    }
    
    public string GenerateToken(User user)
    {
        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_secret));
        var credentials = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);
        
        var claims = new[]
        {
            new Claim(JwtRegisteredClaimNames.Sub, user.Id.ToString()),
            new Claim(JwtRegisteredClaimNames.Email, user.Email.Value),
            new Claim(JwtRegisteredClaimNames.Name, user.Username),
            new Claim(ClaimTypes.Name, user.Username),
            new Claim(ClaimTypes.Email, user.Email.Value),
            new Claim(ClaimTypes.NameIdentifier, user.Id.ToString()),
            new Claim("role", string.Join(",", user.Roles.Select(r => r.Name)))
        };
        
        var token = new JwtSecurityToken(
            issuer: _issuer,
            audience: _audience,
            claims: claims,
            expires: DateTime.UtcNow.AddHours(24),
            signingCredentials: credentials);
        
        return new JwtSecurityTokenHandler().WriteToken(token);
    }
    
    public ClaimsPrincipal? GetPrincipalFromToken(string token)
    {
        try
        {
            var tokenHandler = new JwtSecurityTokenHandler();
            var key = Encoding.UTF8.GetBytes(_secret);
            
            var validationParameters = new TokenValidationParameters
            {
                ValidateIssuer = true,
                ValidateAudience = true,
                ValidateLifetime = true,
                ValidateIssuerSigningKey = true,
                ValidIssuer = _issuer,
                ValidAudience = _audience,
                IssuerSigningKey = new SymmetricSecurityKey(key),
                ClockSkew = TimeSpan.Zero
            };
            
            var principal = tokenHandler.ValidateToken(token, validationParameters, out var validatedToken);
            
            return principal;
        }
        catch
        {
            return null;
        }
    }
}

// Authorization policies
public static class AuthorizationConfiguration
{
    public static IServiceCollection AddAuthorizationPolicies(this IServiceCollection services)
    {
        services.AddAuthorization(options =>
        {
            options.AddPolicy("AdminOnly", policy =>
                policy.RequireRole("Admin"));
            
            options.AddPolicy("UserOrAdmin", policy =>
                policy.RequireRole("User", "Admin"));
            
            options.AddPolicy("CanManageUsers", policy =>
                policy.Requirements.Add(new CanManageUsersRequirement()));
            
            options.AddPolicy("MinimumAge", policy =>
                policy.Requirements.Add(new MinimumAgeRequirement(18)));
        });
        
        services.AddScoped<IAuthorizationHandler, CanManageUsersHandler>();
        services.AddScoped<IAuthorizationHandler, MinimumAgeHandler>();
        
        return services;
    }
}

// Custom authorization requirements
public class CanManageUsersRequirement : IAuthorizationRequirement { }

public class MinimumAgeRequirement : IAuthorizationRequirement
{
    public MinimumAgeRequirement(int minimumAge)
    {
        MinimumAge = minimumAge;
    }
    
    public int MinimumAge { get; }
}

// Authorization handlers
public class CanManageUsersHandler : AuthorizationHandler<CanManageUsersRequirement>
{
    protected override Task HandleRequirementAsync(AuthorizationHandlerContext context,
        CanManageUsersRequirement requirement)
    {
        if (context.User.IsInRole("Admin"))
        {
            context.Succeed(requirement);
        }
        
        return Task.CompletedTask;
    }
}

// Using authorization in controllers
[ApiController]
[Route("api/[controller]")]
[Authorize]
public class UsersController : ControllerBase
{
    [HttpGet]
    [AllowAnonymous]
    public async Task<IActionResult> GetUsers()
    {
        // Public endpoint
    }
    
    [HttpPost]
    [Authorize(Roles = "Admin")]
    public async Task<IActionResult> CreateUser()
    {
        // Admin only endpoint
    }
    
    [HttpDelete("{id}")]
    [Authorize(Policy = "CanManageUsers")]
    public async Task<IActionResult> DeleteUser(Guid id)
    {
        // Users who can manage users
    }
}
```

### Security Headers and Middleware

```csharp
// Security headers middleware
public class SecurityHeadersMiddleware
{
    private readonly RequestDelegate _next;
    
    public SecurityHeadersMiddleware(RequestDelegate next)
    {
        _next = next;
    }
    
    public async Task InvokeAsync(HttpContext context)
    {
        // Add security headers
        context.Response.Headers.Add("X-Content-Type-Options", "nosniff");
        context.Response.Headers.Add("X-Frame-Options", "DENY");
        context.Response.Headers.Add("X-XSS-Protection", "1; mode=block");
        context.Response.Headers.Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains");
        context.Response.Headers.Add("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'");
        context.Response.Headers.Add("Referrer-Policy", "strict-origin-when-cross-origin");
        
        await _next(context);
    }
}

// Rate limiting middleware
public class RateLimitingMiddleware
{
    private readonly RequestDelegate _next;
    private readonly IMemoryCache _cache;
    private readonly RateLimitOptions _options;
    
    public RateLimitingMiddleware(RequestDelegate next, IMemoryCache cache, IOptions<RateLimitOptions> options)
    {
        _next = next;
        _cache = cache;
        _options = options.Value;
    }
    
    public async Task InvokeAsync(HttpContext context)
    {
        var clientId = GetClientId(context);
        var cacheKey = $"rate_limit_{clientId}";
        
        var requestCount = await _cache.GetOrCreateAsync(cacheKey, async () =>
        {
            await Task.CompletedTask;
            return 1;
        }, new MemoryCacheEntryOptions
        {
            AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(1),
            SlidingExpiration = TimeSpan.FromSeconds(30)
        });
        
        if (requestCount > _options.MaxRequestsPerMinute)
        {
            context.Response.StatusCode = StatusCodes.Status429TooManyRequests;
            await context.Response.WriteAsync("Rate limit exceeded");
            return;
        }
        
        await _next(context);
    }
    
    private string GetClientId(HttpContext context)
    {
        // Try to get user ID from claims, otherwise use IP address
        var userId = context.User.FindFirst(ClaimTypes.NameIdentifier)?.Value;
        return userId ?? context.Connection.RemoteIpAddress?.ToString() ?? "anonymous";
    }
}

public class RateLimitOptions
{
    public int MaxRequestsPerMinute { get; set; } = 100;
}

// Configure middleware in Program.cs
app.UseMiddleware<SecurityHeadersMiddleware>();
app.UseMiddleware<RateLimitingMiddleware>();
```

## Integration Patterns

### Entity Framework Core Integration

```csharp
// Repository pattern with Unit of Work
public interface IUnitOfWork : IDisposable
{
    IUserRepository Users { get; }
    Task<int> SaveChangesAsync(CancellationToken cancellationToken = default);
    Task BeginTransactionAsync(CancellationToken cancellationToken = default);
    Task CommitAsync(CancellationToken cancellationToken = default);
    Task RollbackAsync(CancellationToken cancellationToken = default);
}

public class UnitOfWork : IUnitOfWork
{
    private readonly AppDbContext _context;
    private IDbContextTransaction? _transaction;
    
    public UnitOfWork(AppDbContext context)
    {
        _context = context;
        Users = new UserRepository(_context);
    }
    
    public IUserRepository Users { get; }
    
    public async Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        return await _context.SaveChangesAsync(cancellationToken);
    }
    
    public async Task BeginTransactionAsync(CancellationToken cancellationToken = default)
    {
        _transaction = await _context.Database.BeginTransactionAsync(cancellationToken);
    }
    
    public async Task CommitAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            await _context.SaveChangesAsync(cancellationToken);
            await _transaction?.CommitAsync(cancellationToken)!;
        }
        catch
        {
            await RollbackAsync(cancellationToken);
            throw;
        }
    }
    
    public async Task RollbackAsync(CancellationToken cancellationToken = default)
    {
        if (_transaction != null)
        {
            await _transaction.RollbackAsync(cancellationToken);
        }
    }
    
    public void Dispose()
    {
        _transaction?.Dispose();
        _context.Dispose();
    }
}

// Service with Unit of Work
public class UserService
{
    private readonly IUnitOfWork _unitOfWork;
    private readonly ILogger<UserService> _logger;
    
    public UserService(IUnitOfWork unitOfWork, ILogger<UserService> logger)
    {
        _unitOfWork = unitOfWork;
        _logger = logger;
    }
    
    public async Task<User> CreateUserAsync(CreateUserRequest request,
        CancellationToken cancellationToken = default)
    {
        await _unitOfWork.BeginTransactionAsync(cancellationToken);
        
        try
        {
            var user = new User(request.Username, Email.Create(request.Email),
                new UserProfile(request.FirstName, request.LastName, request.Bio));
            
            await _unitOfWork.Users.AddAsync(user, cancellationToken);
            await _unitOfWork.CommitAsync(cancellationToken);
            
            _logger.LogInformation("User created successfully with ID: {UserId}", user.Id);
            return user;
        }
        catch
        {
            await _unitOfWork.RollbackAsync(cancellationToken);
            throw;
        }
    }
}
```

### Message Queue Integration

```csharp
// RabbitMQ integration
public interface IMessagePublisher
{
    Task PublishAsync<T>(T message, string routingKey, CancellationToken cancellationToken = default);
}

public class RabbitMqPublisher : IMessagePublisher
{
    private readonly IConnection _connection;
    private readonly ILogger<RabbitMqPublisher> _logger;
    
    public RabbitMqPublisher(IConnection connection, ILogger<RabbitMqPublisher> logger)
    {
        _connection = connection;
        _logger = logger;
    }
    
    public async Task PublishAsync<T>(T message, string routingKey,
        CancellationToken cancellationToken = default)
    {
        using var channel = await _connection.CreateChannelAsync(cancellationToken);
        
        await channel.ExchangeDeclareAsync("app_exchange", ExchangeType.Direct, durable: true, 
            cancellationToken: cancellationToken);
        
        var messageBody = JsonSerializer.SerializeToUtf8Bytes(message);
        
        await channel.BasicPublishAsync("app_exchange", routingKey, true, new BasicProperties
        {
            ContentType = "application/json",
            DeliveryMode = DeliveryModes.Persistent
        }, messageBody, cancellationToken);
        
        _logger.LogInformation("Message published to routing key: {RoutingKey}", routingKey);
    }
}

// Message consumer
public interface IMessageConsumer<T>
{
    Task ConsumeAsync(T message, CancellationToken cancellationToken = default);
}

public class UserCreatedEventConsumer : BackgroundService
{
    private readonly IConnection _connection;
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly ILogger<UserCreatedEventConsumer> _logger;
    
    public UserCreatedEventConsumer(IConnection connection, IServiceScopeFactory scopeFactory,
        ILogger<UserCreatedEventConsumer> logger)
    {
        _connection = connection;
        _scopeFactory = scopeFactory;
        _logger = logger;
    }
    
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        using var channel = await _connection.CreateChannelAsync(stoppingToken);
        
        await channel.QueueDeclareAsync("user_created_queue", durable: true, exclusive: false,
            autoDelete: false, cancellationToken: stoppingToken);
        
        await channel.QueueBindAsync("user_created_queue", "app_exchange", "user.created",
            cancellationToken: stoppingToken);
        
        var consumer = new AsyncEventingBasicConsumer(channel);
        consumer.ReceivedAsync += async (sender, args) =>
        {
            try
            {
                var message = JsonSerializer.Deserialize<UserCreatedEvent>(args.Body.ToArray());
                
                if (message != null)
                {
                    await HandleUserCreatedEvent(message, stoppingToken);
                }
                
                await channel.BasicAckAsync(args.DeliveryTag, false, stoppingToken);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error processing user created event");
                await channel.BasicNackAsync(args.DeliveryTag, false, true, stoppingToken);
            }
        };
        
        await channel.BasicConsumeAsync("user_created_queue", false, consumer, stoppingToken);
        
        _logger.LogInformation("User created event consumer started");
        
        // Keep the consumer running
        while (!stoppingToken.IsCancellationRequested)
        {
            await Task.Delay(TimeSpan.FromSeconds(1), stoppingToken);
        }
    }
    
    private async Task HandleUserCreatedEvent(UserCreatedEvent @event,
        CancellationToken cancellationToken = default)
    {
        using var scope = _scopeFactory.CreateScope();
        var emailService = scope.ServiceProvider.GetRequiredService<IEmailService>();
        
        await emailService.SendWelcomeEmailAsync(@event.Email, @event.FirstName, @event.LastName, cancellationToken);
        
        _logger.LogInformation("Processed user created event for user: {UserId}", @event.UserId);
    }
}

public record UserCreatedEvent(
    Guid UserId,
    string Email,
    string FirstName,
    string LastName,
    DateTime CreatedAt);
```

## Modern Development Workflow

### Project Configuration

```xml
<!-- Directory.Build.props -->
<Project>
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
    <LangVersion>12.0</LangVersion>
    <GenerateDocumentationFile>true</GenerateDocumentationFile>
    <NoWarn>$(NoWarn);1591</NoWarn>
  </PropertyGroup>
  
  <ItemGroup>
    <PackageReference Include="Microsoft.Extensions.Logging" Version="8.0.0" />
    <PackageReference Include="Microsoft.Extensions.DependencyInjection" Version="8.0.0" />
    <PackageReference Include="Microsoft.Extensions.Configuration" Version="8.0.0" />
    <PackageReference Include="Microsoft.Extensions.Hosting" Version="8.0.0" />
  </ItemGroup>
</Project>

<!-- MyApi.csproj -->
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <AssemblyName>MyApi</AssemblyName>
    <RootNamespace>MyApi</RootNamespace>
  </PropertyGroup>
  
  <ItemGroup>
    <PackageReference Include="Microsoft.AspNetCore.OpenApi" Version="8.0.0" />
    <PackageReference Include="Microsoft.EntityFrameworkCore.SqlServer" Version="8.0.0" />
    <PackageReference Include="Microsoft.EntityFrameworkCore.Design" Version="8.0.0" />
    <PackageReference Include="Microsoft.EntityFrameworkCore.Tools" Version="8.0.0" />
    <PackageReference Include="Microsoft.AspNetCore.Authentication.JwtBearer" Version="8.0.0" />
    <PackageReference Include="Microsoft.AspNetCore.Authorization" Version="8.0.0" />
    <PackageReference Include="Swashbuckle.AspNetCore" Version="6.5.0" />
    <PackageReference Include="FluentValidation.AspNetCore" Version="11.3.0" />
    <PackageReference Include="Serilog.AspNetCore" Version="8.0.0" />
    <PackageReference Include="Serilog.Sinks.Console" Version="5.0.0" />
    <PackageReference Include="Serilog.Sinks.File" Version="5.0.0" />
  </ItemGroup>
</Project>
```

### Docker Configuration

```dockerfile
# Multi-stage Dockerfile for .NET 8.0
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build-env
WORKDIR /app

# Copy csproj and restore dependencies
COPY *.csproj ./
RUN dotnet restore

# Copy everything else and build
COPY . ./
RUN dotnet publish -c Release -o out

# Build runtime image
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app
COPY --from=build-env /app/out .

# Create non-root user
RUN adduser --disabled-password --gecos '' appuser && chown -R appuser /app
USER appuser

EXPOSE 8080
ENV ASPNETCORE_URLS=http://+:8080
ENV ASPNETCORE_ENVIRONMENT=Production

ENTRYPOINT ["dotnet", "MyApi.dll"]
```

### CI/CD Configuration

```yaml
# .github/workflows/dotnet.yml
name: .NET CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  DOTNET_VERSION: '8.0.x'
  SOLUTION_FILE: 'MySolution.sln'
  
jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: ${{ env.DOTNET_VERSION }}
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/.nuget/packages
        key: ${{ runner.os }}-nuget-${{ hashFiles('**/*.csproj') }}
        restore-keys: |
          ${{ runner.os }}-nuget-
    
    - name: Restore dependencies
      run: dotnet restore ${{ env.SOLUTION_FILE }}
    
    - name: Build
      run: dotnet build ${{ env.SOLUTION_FILE }} --no-restore --configuration Release
    
    - name: Run tests
      run: dotnet test ${{ env.SOLUTION_FILE }} --no-build --configuration Release --verbosity normal --collect:"XPlat Code Coverage"
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: '**/coverage.cobertura.xml'
    
    - name: Run security scan
      run: |
        dotnet tool install --global dotnet-ssh --version 0.2.0
        dotnet-ssh scan MyApi/bin/Release/net8.0/MyApi.dll
  
  build-and-deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: ${{ env.DOTNET_VERSION }}
    
    - name: Build and publish
      run: dotnet publish ${{ env.SOLUTION_FILE }} --configuration Release --output ./publish
    
    - name: Build Docker image
      run: |
        docker build -t myregistry.com/myapi:${{ github.sha }} .
        docker tag myregistry.com/myapi:${{ github.sha }} myregistry.com/myapi:latest
    
    - name: Push Docker image
      if: github.ref == 'refs/heads/main'
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push myregistry.com/myapi:${{ github.sha }}
        docker push myregistry.com/myapi:latest
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**C# Target**: .NET 8.0 LTS with C# 12.0 and modern ASP.NET Core patterns  

This skill provides comprehensive C# development guidance with 2025 best practices, covering everything from modern ASP.NET Core APIs to Entity Framework Core optimization and cross-platform development with .NET MAUI.
