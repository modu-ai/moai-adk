# C# 13 Performance Optimization

**Focus**: Runtime performance, memory efficiency, compiler optimizations
**Standards**: .NET 9 with benchmarking tools
**Last Updated**: 2025-11-22

---

## Optimization Level 1: Allocation Reduction

```csharp
// Bad: Multiple allocations in loop
foreach (var item in items)
{
    var json = JsonSerializer.Serialize(item);  // Allocates each time
    Process(json);
}

// Good: Reuse serializer
var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
foreach (var item in items)
{
    var json = JsonSerializer.Serialize(item, options);  // Reuses options
}

// Better: Use Span<T> to avoid allocations
Span<byte> buffer = stackalloc byte[1024];
var length = JsonSerializer.GetBytes(item, buffer);
Process(buffer[..length]);
```

---

## Optimization Level 2: Method Inlining

```csharp
// Small hot-path methods should be inlined
[MethodImpl(MethodImplOptions.AggressiveInlining)]
public static int Add(int a, int b) => a + b;

// Prevent inlining for large methods
[MethodImpl(MethodImplOptions.NoInlining)]
public void LogException(Exception ex)
{
    // Complex logging logic
}

// Tiered compilation - optimize hot paths
[MethodImpl(MethodImplOptions.AggressiveOptimization)]
public decimal CalculateTotal(List<Order> orders)
{
    decimal total = 0m;
    foreach (var order in orders)
    {
        total += order.Amount;
    }
    return total;
}
```

---

## Optimization Level 3: Object Pooling

```csharp
using System.Buffers;

// Reuse arrays from ArrayPool
var buffer = ArrayPool<byte>.Shared.Rent(1024);
try
{
    // Use buffer
    Process(buffer);
}
finally
{
    ArrayPool<byte>.Shared.Return(buffer);
}

// Custom object pool
public class ObjectPool<T> where T : class, new()
{
    private readonly ConcurrentBag<T> _pool = new();
    private readonly Func<T> _factory;

    public ObjectPool(Func<T>? factory = null)
    {
        _factory = factory ?? (() => new T());
    }

    public T Rent()
    {
        return _pool.TryTake(out var item) ? item : _factory();
    }

    public void Return(T item)
    {
        _pool.Add(item);
    }
}
```

---

## Optimization Level 4: Collection Performance

```csharp
// Choose right collection type
// O(1) lookups - Dictionary
var userMap = new Dictionary<int, User>();

// Ordered with O(log n) lookups - SortedDictionary
var sortedUsers = new SortedDictionary<string, User>();

// Fast enumeration - List (not dictionary if order matters)
var userList = new List<User>();

// Concurrent access - ConcurrentDictionary
var concurrentCache = new ConcurrentDictionary<int, User>();

// Capacity pre-allocation prevents resizing
var list = new List<User>(capacity: 1000);  // Avoids internal resizing
var dict = new Dictionary<int, User>(capacity: 1000);
```

---

## Optimization Level 5: LINQ Query Optimization

```csharp
// Bad: Multiple enumerations
var hasActive = users.Any(u => u.IsActive);
var activeCount = users.Count(u => u.IsActive);
var activeUsers = users.Where(u => u.IsActive).ToList();

// Good: Single enumeration
var active = users.Where(u => u.IsActive).ToList();
var hasActive = active.Any();
var activeCount = active.Count;

// Bad: Filtering inefficiently
var expensive = users
    .Where(u => ExpensiveCheck(u))  // Applied to all
    .Where(u => u.Age > 18)         // Applied to filtered set
    .ToList();

// Good: Filter cheap conditions first
var optimized = users
    .Where(u => u.Age > 18)         // Cheap check first
    .Where(u => ExpensiveCheck(u))  // Applied to smaller set
    .ToList();

// Bad: ForEach for side effects with allocation
var tasks = users.Select(u => ProcessAsync(u)).ToList();
await Task.WhenAll(tasks);

// Good: Await directly
await Parallel.ForEachAsync(users, async (user, _) =>
{
    await ProcessAsync(user);
});
```

---

## Optimization Level 6: Async/Await Efficiency

```csharp
// Bad: Unnecessary Task creation
public async Task<int> GetCountAsync()
{
    return await Task.FromResult(42);  // Unnecessary
}

// Good: Synchronous return when possible
public int GetCount() => 42;

// Bad: ConfigureAwait(true) in library code
public async Task ProcessAsync()
{
    await someTask.ConfigureAwait(true);  // Captures context
}

// Good: Use ConfigureAwait(false)
public async Task ProcessAsync()
{
    await someTask.ConfigureAwait(false);  // Avoids context switch
}

// Bad: Task.Run for I/O
public async Task<Data> FetchAsync()
{
    return await Task.Run(() => httpClient.GetStringAsync(url));
}

// Good: Direct await for I/O
public async Task<Data> FetchAsync()
{
    return await httpClient.GetStringAsync(url);
}
```

---

## Optimization Level 7: Entity Framework Performance

```csharp
// Bad: Lazy loading with N+1 queries
var users = context.Users.ToList();
foreach (var user in users)
{
    var posts = user.Posts;  // Triggers query per user
}

// Good: Eager loading
var users = context.Users
    .Include(u => u.Posts)
    .ToList();

// Bad: Unnecessary change tracking
var users = context.Users.ToList();
var filtered = users.Where(u => u.IsActive).ToList();

// Good: AsNoTracking for read-only
var users = context.Users
    .AsNoTracking()
    .Where(u => u.IsActive)
    .ToList();

// Bad: Loading entire entities
var names = context.Users
    .Select(u => u.User)  // Loads entire entity
    .ToList();

// Good: Projecting only needed fields
var names = context.Users
    .Select(u => u.Name)
    .ToList();

// Bulk operations
context.Users
    .Where(u => !u.IsActive)
    .ExecuteDelete();  // Single DELETE statement

context.Users
    .Where(u => u.CreatedDate < cutoffDate)
    .ExecuteUpdate(s => s.SetProperty(u => u.IsArchived, true));
```

---

## Optimization Level 8: String Performance

```csharp
// Bad: String concatenation in loops
var result = "";
for (int i = 0; i < 1000; i++)
{
    result += i.ToString();  // O(n²) allocations
}

// Good: StringBuilder
var sb = new StringBuilder();
for (int i = 0; i < 1000; i++)
{
    sb.Append(i);
}
var result = sb.ToString();

// Better: String.Create for precision
var result = string.Create<int>(100, 42, (span, value) =>
{
    var written = value.TryFormat(span, out var chars);
});

// Bad: String.Format allocations
var text = string.Format("Hello {0} {1}", name, age);

// Good: String interpolation (usually inlines)
var text = $"Hello {name} {age}";
```

---

## Optimization Level 9: Benchmarking with BenchmarkDotNet

```csharp
using BenchmarkDotNet.Attributes;
using BenchmarkDotNet.Running;

[MemoryDiagnoser]
[SimpleJob(warmupCount: 3, targetCount: 5)]
public class MyBenchmark
{
    private int[] _data = new int[1000];

    [GlobalSetup]
    public void Setup()
    {
        _data = Enumerable.Range(0, 1000).ToArray();
    }

    [Benchmark]
    public int Linq()
    {
        return _data.Sum();
    }

    [Benchmark]
    public int ForLoop()
    {
        int sum = 0;
        for (int i = 0; i < _data.Length; i++)
        {
            sum += _data[i];
        }
        return sum;
    }
}

// Run: BenchmarkRunner.Run<MyBenchmark>();
```

---

## Optimization Level 10: Profile-Guided Optimization

**Using .NET tools**:
```bash
# Generate PGO data
dotnet publish -c Release -p:PublishReadyToRun=true

# Use with profiling
dotnet publish -c Release -p:PublishTieredGraph=true

# Analyze with dotnet-trace
dotnet trace collect dotnet MyApp.dll
dotnet trace analyze trace.nettrace
```

---

## Compiler Optimization Settings

```xml
<!-- .csproj optimization settings -->
<PropertyGroup>
    <!-- Enable Tier1 and Tier2 JIT -->
    <TieredCompilation>true</TieredCompilation>

    <!-- Profile-guided optimization -->
    <PublishReadyToRun>true</PublishReadyToRun>

    <!-- Enable code generation for Ready2Run -->
    <PublishTrimmed>true</PublishTrimmed>

    <!-- Disable debugging in Release -->
    <DebugType Condition="'$(Configuration)' == 'Release'">none</DebugType>

    <!-- Optimize for throughput -->
    <TieredCompilationQuickJitForLoops>true</TieredCompilationQuickJitForLoops>
</PropertyGroup>
```

---

**Last Updated**: 2025-11-22 | **Production Ready**
