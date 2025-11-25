# Advanced C# 13 Patterns

**Focus**: Enterprise patterns, async/await advanced usage, LINQ optimization, concurrency
**Standards**: C# 13 (.NET 9)
**Last Updated**: 2025-11-22

---

## Pattern 1: Advanced Async/Await with Cancellation

```csharp
// Proper cancellation support
public async Task<T> FetchWithCancellationAsync<T>(
    string url,
    CancellationToken cancellationToken)
{
    using var client = new HttpClient();
    var response = await client.GetAsync(url, cancellationToken);
    response.EnsureSuccessStatusCode();
    return await response.Content.ReadAsAsync<T>(cancellationToken);
}

// Timeout wrapper
public async Task<T> WithTimeoutAsync<T>(
    Task<T> task,
    TimeSpan timeout)
{
    using var cts = new CancellationTokenSource(timeout);
    try
    {
        return await task;
    }
    catch (OperationCanceledException)
    {
        throw new TimeoutException($"Operation timed out after {timeout.TotalSeconds}s");
    }
}

// Retry pattern with exponential backoff
public async Task<T> RetryAsync<T>(
    Func<CancellationToken, Task<T>> operation,
    int maxRetries = 3,
    TimeSpan? initialDelay = null)
{
    var delay = initialDelay ?? TimeSpan.FromMilliseconds(100);

    for (int i = 0; i < maxRetries; i++)
    {
        try
        {
            using var cts = new CancellationTokenSource();
            return await operation(cts.Token);
        }
        catch when (i < maxRetries - 1)
        {
            await Task.Delay(delay);
            delay = TimeSpan.FromMilliseconds(delay.TotalMilliseconds * 2);
        }
    }
    throw new InvalidOperationException("Max retries exceeded");
}
```

---

## Pattern 2: LINQ Performance Optimization

```csharp
// Bad: N+1 queries
var users = context.Users.ToList();  // Query 1
foreach (var user in users)
{
    var posts = context.Posts.Where(p => p.UserId == user.Id).ToList();  // N queries
}

// Good: Single query with Include
var users = context.Users
    .Include(u => u.Posts)
    .ThenInclude(p => p.Comments)
    .ToList();

// Better: Projection to avoid loading unneeded data
var userPostCounts = context.Users
    .Select(u => new
    {
        u.Id,
        u.Name,
        PostCount = u.Posts.Count
    })
    .ToList();

// Best: AsNoTracking for read-only queries
var readOnlyUsers = context.Users
    .AsNoTracking()
    .Where(u => u.IsActive)
    .OrderBy(u => u.Name)
    .ToList();

// Parallel LINQ for CPU-bound operations
var result = data
    .AsParallel()
    .Where(x => ExpensiveOperation(x))
    .Select(x => TransformData(x))
    .ToList();
```

---

## Pattern 3: Entity Framework Core Best Practices

```csharp
// Batch operations for performance
public async Task BulkInsertAsync(List<User> users)
{
    context.Users.AddRange(users);
    await context.SaveChangesAsync();
}

// Update only changed fields
public async Task UpdateUserAsync(User user)
{
    var existing = await context.Users.FindAsync(user.Id);
    if (existing == null) throw new InvalidOperationException("User not found");

    // Only update changed fields
    context.Entry(existing).CurrentValues.SetValues(user);
    await context.SaveChangesAsync();
}

// Transaction support
public async Task<bool> TransferAsync(int fromUserId, int toUserId, decimal amount)
{
    using var transaction = await context.Database.BeginTransactionAsync();
    try
    {
        var from = await context.Users.FindAsync(fromUserId);
        var to = await context.Users.FindAsync(toUserId);

        if (from.Balance < amount) return false;

        from.Balance -= amount;
        to.Balance += amount;

        await context.SaveChangesAsync();
        await transaction.CommitAsync();
        return true;
    }
    catch
    {
        await transaction.RollbackAsync();
        throw;
    }
}

// Compiled queries for repeated operations
private static readonly Func<AppDbContext, int, Task<User>> GetUserQuery =
    EF.CompileAsyncQuery((AppDbContext context, int id) =>
        context.Users.First(u => u.Id == id));

public async Task<User> GetUserOptimizedAsync(int id)
{
    return await GetUserQuery(context, id);
}
```

---

## Pattern 4: Dependency Injection with Lifetime Management

```csharp
// Scoped - One instance per request
builder.Services.AddScoped<IRepository, Repository>();

// Transient - New instance every time
builder.Services.AddTransient<ILogger, ConsoleLogger>();

// Singleton - One instance for entire application
builder.Services.AddSingleton<ICache, MemoryCache>();

// Factory pattern for complex creation
builder.Services.AddScoped<IUserService>(sp =>
    new UserService(
        sp.GetRequiredService<IRepository>(),
        sp.GetRequiredService<ICache>(),
        new HttpClient { Timeout = TimeSpan.FromSeconds(30) }
    )
);

// Options pattern for configuration
builder.Services
    .Configure<AppSettings>(builder.Configuration.GetSection("AppSettings"))
    .AddSingleton<IValidateOptions<AppSettings>, AppSettingsValidator>();
```

---

## Pattern 5: Middleware Pipeline Customization

```csharp
// Custom middleware
public class RequestLoggingMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ILogger<RequestLoggingMiddleware> _logger;

    public RequestLoggingMiddleware(RequestDelegate next, ILogger<RequestLoggingMiddleware> logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var stopwatch = System.Diagnostics.Stopwatch.StartNew();

        try
        {
            await _next(context);
        }
        finally
        {
            stopwatch.Stop();
            _logger.LogInformation(
                "Request {Method} {Path} completed in {ElapsedMs}ms",
                context.Request.Method,
                context.Request.Path,
                stopwatch.ElapsedMilliseconds
            );
        }
    }
}

// Usage
app.UseMiddleware<RequestLoggingMiddleware>();
```

---

## Pattern 6: Record Types for Data Transfer Objects

```csharp
// Immutable DTO with value equality
public record UserDto(int Id, string Name, string Email);

// Record with init-only properties
public record OrderDto
{
    public int Id { get; init; }
    public string CustomerName { get; init; }
    public decimal Total { get; init; }
}

// With validation
public record CreateUserRequest(string Name, string Email)
{
    public CreateUserRequest() : this(string.Empty, string.Empty) { }

    public bool Validate(out string[] errors)
    {
        var errorList = new List<string>();
        if (string.IsNullOrEmpty(Name)) errorList.Add("Name required");
        if (!Email.Contains("@")) errorList.Add("Invalid email");
        errors = errorList.ToArray();
        return errors.Length == 0;
    }
}
```

---

## Pattern 7: Immutable Collections for Thread Safety

```csharp
using System.Collections.Immutable;

// Thread-safe collections
private readonly ImmutableList<int> _numbers = ImmutableList<int>.Empty;

public void AddNumber(int number)
{
    // Creates new collection, doesn't modify original
    var updated = _numbers.Add(number);
}

// Builder pattern for efficient creation
var builder = ImmutableList.CreateBuilder<string>();
for (int i = 0; i < 1000; i++)
{
    builder.Add($"Item {i}");
}
var immutableList = builder.ToImmutable();
```

---

## Pattern 8: Nullable Reference Types Safety

```csharp
#nullable enable

public class UserService
{
    private readonly IRepository _repository;

    public UserService(IRepository repository)
    {
        _repository = repository ?? throw new ArgumentNullException(nameof(repository));
    }

    public async Task<User?> GetUserAsync(int id)
    {
        return await _repository.GetUserAsync(id);
    }

    // Must handle null
    public async Task DisplayUserAsync(int id)
    {
        var user = await GetUserAsync(id);
        if (user is not null)
        {
            Console.WriteLine(user.Name);
        }
    }
}

#nullable restore
```

---

## Pattern 9: Expression Trees for Dynamic Queries

```csharp
// Build queries dynamically
public Expression<Func<User, bool>> BuildFilter(string name, int? minAge)
{
    var parameter = Expression.Parameter(typeof(User), "u");
    Expression predicate = null!;

    if (!string.IsNullOrEmpty(name))
    {
        var nameProperty = Expression.Property(parameter, nameof(User.Name));
        var constant = Expression.Constant(name);
        var call = Expression.Call(nameProperty,
            typeof(string).GetMethod("Contains", new[] { typeof(string) })!,
            constant);
        predicate = predicate == null ? call : Expression.AndAlso(predicate, call);
    }

    if (minAge.HasValue)
    {
        var ageProperty = Expression.Property(parameter, nameof(User.Age));
        var constant = Expression.Constant(minAge.Value);
        var greaterThan = Expression.GreaterThanOrEqual(ageProperty, constant);
        predicate = predicate == null ? greaterThan : Expression.AndAlso(predicate, greaterThan);
    }

    return Expression.Lambda<Func<User, bool>>(
        predicate ?? Expression.Constant(true),
        parameter);
}
```

---

## Pattern 10: Source Generators for Compile-Time Optimization

```csharp
// Client code (automatically generates mapper)
[AutoMapper]
public partial class UserMapper
{
    public partial UserDto Map(User user);
}

// Usage
var mapper = new UserMapper();
var dto = mapper.Map(user);  // No reflection at runtime
```

---

**Last Updated**: 2025-11-22 | **Production Ready**
