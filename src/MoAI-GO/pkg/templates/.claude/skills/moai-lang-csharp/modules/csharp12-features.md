# C# 12 Language Features

Comprehensive guide to C# 12 features for .NET 8+ development.

---

## Primary Constructors

Primary constructors allow defining constructor parameters directly on the class declaration, reducing boilerplate code.

### Basic Usage

```csharp
// Traditional approach
public class UserService
{
    private readonly IUserRepository _repository;
    private readonly ILogger<UserService> _logger;

    public UserService(IUserRepository repository, ILogger<UserService> logger)
    {
        _repository = repository;
        _logger = logger;
    }

    public async Task<User?> GetByIdAsync(Guid id)
    {
        _logger.LogInformation("Fetching user {UserId}", id);
        return await _repository.FindByIdAsync(id);
    }
}

// C# 12 Primary Constructor
public class UserService(IUserRepository repository, ILogger<UserService> logger)
{
    public async Task<User?> GetByIdAsync(Guid id)
    {
        logger.LogInformation("Fetching user {UserId}", id);
        return await repository.FindByIdAsync(id);
    }
}
```

### With Dependency Injection

```csharp
// Service with multiple dependencies
public class OrderService(
    IOrderRepository orderRepository,
    IPaymentService paymentService,
    INotificationService notificationService,
    ILogger<OrderService> logger)
{
    public async Task<Order> CreateOrderAsync(CreateOrderRequest request)
    {
        logger.LogInformation("Creating order for customer {CustomerId}", request.CustomerId);
        
        var order = Order.Create(request.CustomerId, request.Items);
        await orderRepository.AddAsync(order);
        
        await paymentService.ProcessPaymentAsync(order.Id, request.PaymentDetails);
        await notificationService.SendOrderConfirmationAsync(order);
        
        return order;
    }
}
```

### Capturing Parameters as Fields

```csharp
// When you need to store a parameter
public class CacheService(IMemoryCache cache, TimeSpan defaultExpiration)
{
    // Explicitly capture as field when needed
    private readonly TimeSpan _defaultExpiration = defaultExpiration;

    public T GetOrCreate<T>(string key, Func<T> factory)
    {
        return cache.GetOrCreate(key, entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = _defaultExpiration;
            return factory();
        })!;
    }
}
```

### With Record Types

```csharp
// Records with primary constructors (existing feature, enhanced in C# 12)
public record UserDto(Guid Id, string Name, string Email);

public record CreateUserCommand(string Name, string Email, string Password);

public record OrderItem(Guid ProductId, int Quantity, decimal UnitPrice)
{
    public decimal Total => Quantity * UnitPrice;
}

// Immutable record with validation
public record Email
{
    public string Value { get; }
    
    public Email(string value)
    {
        if (!value.Contains('@'))
            throw new ArgumentException("Invalid email format", nameof(value));
        Value = value;
    }
}
```

### Inheritance with Primary Constructors

```csharp
// Base class with primary constructor
public class Entity(Guid id)
{
    public Guid Id { get; } = id;
}

// Derived class calling base primary constructor
public class User(Guid id, string name, string email) : Entity(id)
{
    public string Name { get; } = name;
    public string Email { get; } = email;
}

// Controller with base class
public class BaseController(ILogger logger)
{
    protected void LogInfo(string message) => logger.LogInformation(message);
}

public class UsersController(IUserService userService, ILogger<UsersController> logger)
    : BaseController(logger)
{
    public async Task<User?> GetById(Guid id)
    {
        LogInfo($"Getting user {id}");
        return await userService.GetByIdAsync(id);
    }
}
```

---

## Collection Expressions

Unified syntax for creating arrays, lists, spans, and other collection types.

### Basic Collections

```csharp
// Arrays
int[] numbers = [1, 2, 3, 4, 5];
string[] names = ["Alice", "Bob", "Charlie"];

// Lists
List<int> numberList = [1, 2, 3, 4, 5];
List<string> nameList = ["Alice", "Bob", "Charlie"];

// Spans
Span<int> spanNumbers = [10, 20, 30];
ReadOnlySpan<char> chars = ['a', 'b', 'c'];

// Empty collections
int[] empty = [];
List<string> emptyList = [];
```

### Spread Operator

```csharp
int[] first = [1, 2, 3];
int[] second = [4, 5, 6];

// Combine arrays
int[] combined = [..first, ..second];  // [1, 2, 3, 4, 5, 6]

// Add elements with spread
int[] withExtras = [0, ..first, ..second, 7, 8];  // [0, 1, 2, 3, 4, 5, 6, 7, 8]

// Conditional spread
bool includeDefaults = true;
int[] defaults = [100, 200];
int[] result = [..first, ..(includeDefaults ? defaults : [])];
```

### Collection Builders

```csharp
// Custom collection support
public class CustomCollection<T> : IEnumerable<T>
{
    private readonly List<T> _items = [];
    
    public void Add(T item) => _items.Add(item);
    public IEnumerator<T> GetEnumerator() => _items.GetEnumerator();
    IEnumerator IEnumerable.GetEnumerator() => GetEnumerator();
}

// Works with collection expressions
CustomCollection<int> custom = [1, 2, 3];
```

### Practical Examples

```csharp
// Method parameters
public void ProcessItems(params int[] items) { }
ProcessItems([1, 2, 3]);

// LINQ with collection expressions
var filtered = numbers.Where(n => [1, 3, 5].Contains(n));

// Dictionary initialization (still uses traditional syntax)
Dictionary<string, int> scores = new()
{
    ["Alice"] = 100,
    ["Bob"] = 95
};

// Combining collections in expressions
List<User> GetUsers(List<User> admins, List<User> regular, bool includeAdmins)
    => [..(includeAdmins ? admins : []), ..regular];
```

---

## Alias Any Type

Type aliases for any type using the `using` directive, not just namespaces.

### Basic Type Aliases

```csharp
// Alias for tuples
using Point = (int X, int Y);
using Rectangle = (Point TopLeft, Point BottomRight);

public class GeometryService
{
    public Point GetCenter(Rectangle rect)
    {
        return (
            (rect.TopLeft.X + rect.BottomRight.X) / 2,
            (rect.TopLeft.Y + rect.BottomRight.Y) / 2
        );
    }
    
    public double CalculateDistance(Point a, Point b)
    {
        var dx = b.X - a.X;
        var dy = b.Y - a.Y;
        return Math.Sqrt(dx * dx + dy * dy);
    }
}
```

### Generic Type Aliases

```csharp
// Alias for complex generics
using UserCache = System.Collections.Generic.Dictionary<Guid, User>;
using StringList = System.Collections.Generic.List<string>;
using AsyncResult<T> = System.Threading.Tasks.Task<T>;

public class CacheManager
{
    private readonly UserCache _cache = [];
    
    public async AsyncResult<User?> GetOrFetchAsync(Guid id, Func<Guid, AsyncResult<User?>> fetch)
    {
        if (_cache.TryGetValue(id, out var user))
            return user;
            
        user = await fetch(id);
        if (user is not null)
            _cache[id] = user;
            
        return user;
    }
}
```

### Nullable Type Aliases

```csharp
// Alias for nullable types
using NullableInt = int?;
using OptionalString = string?;

public class ConfigService
{
    public NullableInt GetTimeout() => 30;
    public OptionalString GetApiKey() => Environment.GetEnvironmentVariable("API_KEY");
}
```

### Function Type Aliases

```csharp
// Alias for delegate types
using Predicate<T> = System.Func<T, bool>;
using AsyncPredicate<T> = System.Func<T, System.Threading.Tasks.Task<bool>>;
using Handler<T> = System.Action<T>;

public class FilterService
{
    public List<User> Filter(List<User> users, Predicate<User> predicate)
        => users.Where(predicate).ToList();
    
    public async Task<List<User>> FilterAsync(List<User> users, AsyncPredicate<User> predicate)
    {
        var result = new List<User>();
        foreach (var user in users)
        {
            if (await predicate(user))
                result.Add(user);
        }
        return result;
    }
}
```

---

## Default Lambda Parameters

Lambda expressions can now have default parameter values.

### Basic Default Parameters

```csharp
// Lambda with defaults
var greet = (string name, string greeting = "Hello") => $"{greeting}, {name}!";

Console.WriteLine(greet("Alice"));           // "Hello, Alice!"
Console.WriteLine(greet("Bob", "Hi"));       // "Hi, Bob!"

// Multiple defaults
var formatDate = (DateTime date, string format = "yyyy-MM-dd", bool utc = false) =>
    (utc ? date.ToUniversalTime() : date).ToString(format);

Console.WriteLine(formatDate(DateTime.Now));                        // "2024-01-15"
Console.WriteLine(formatDate(DateTime.Now, "dd/MM/yyyy"));          // "15/01/2024"
Console.WriteLine(formatDate(DateTime.Now, "yyyy-MM-dd", true));    // UTC date
```

### In LINQ Expressions

```csharp
// Filter function with default
Func<User, string, bool> matchesSearch = (user, term = "") =>
    string.IsNullOrEmpty(term) || 
    user.Name.Contains(term, StringComparison.OrdinalIgnoreCase);

var users = new List<User> { /* ... */ };

// Use without search term
var allUsers = users.Where(u => matchesSearch(u)).ToList();

// Use with search term
var filtered = users.Where(u => matchesSearch(u, "john")).ToList();
```

### Callback Patterns

```csharp
// Event handler with optional context
Action<string, int, string?> logEvent = (message, level = 1, context = null) =>
{
    var prefix = level switch
    {
        1 => "INFO",
        2 => "WARN",
        3 => "ERROR",
        _ => "DEBUG"
    };
    var ctx = context is not null ? $" [{context}]" : "";
    Console.WriteLine($"[{prefix}]{ctx} {message}");
};

logEvent("Application started");                    // [INFO] Application started
logEvent("High memory usage", 2);                   // [WARN] High memory usage
logEvent("Connection failed", 3, "Database");       // [ERROR] [Database] Connection failed
```

### Configuration Builders

```csharp
// Builder pattern with defaults
var createConfig = (
    string host = "localhost",
    int port = 5432,
    string database = "mydb",
    bool ssl = false) => new ConnectionConfig(host, port, database, ssl);

var defaultConfig = createConfig();
var productionConfig = createConfig("prod-db.example.com", ssl: true);
var customConfig = createConfig(port: 5433, database: "testdb");
```

---

## Inline Arrays

Fixed-size arrays allocated inline with the containing type for performance optimization.

### Basic Inline Arrays

```csharp
[System.Runtime.CompilerServices.InlineArray(10)]
public struct Buffer10<T>
{
    private T _element0;
}

public class BufferExample
{
    public void UseBuffer()
    {
        Buffer10<int> buffer = default;
        
        // Access like a span
        Span<int> span = buffer;
        span[0] = 1;
        span[1] = 2;
        
        // Iterate
        foreach (var item in span)
        {
            Console.WriteLine(item);
        }
    }
}
```

### Performance-Critical Scenarios

```csharp
[InlineArray(16)]
public struct Vector16
{
    private float _element0;
}

public class VectorMath
{
    public static float DotProduct(in Vector16 a, in Vector16 b)
    {
        ReadOnlySpan<float> spanA = a;
        ReadOnlySpan<float> spanB = b;
        
        float sum = 0;
        for (int i = 0; i < 16; i++)
        {
            sum += spanA[i] * spanB[i];
        }
        return sum;
    }
}
```

---

## Interceptors (Experimental)

Compile-time method interception for source generators and AOT optimization.

### Basic Concept

```csharp
// Note: Interceptors are experimental and require opt-in
// <Features>InterceptorsPreviewNamespaces=...</Features>

// Original method
public class Calculator
{
    public int Add(int a, int b) => a + b;
}

// Interceptor (generated by source generator)
static class GeneratedInterceptors
{
    [InterceptsLocation("path/to/file.cs", line: 10, character: 20)]
    public static int Add_Intercepted(this Calculator calc, int a, int b)
    {
        Console.WriteLine($"Intercepted: {a} + {b}");
        return a + b;
    }
}
```

---

## Best Practices

### When to Use Primary Constructors

Good Use Cases:
- Dependency injection in services
- Simple data-holding classes
- Classes with few dependencies

Avoid When:
- Complex initialization logic needed
- Multiple constructors with different signatures
- Parameters need validation or transformation

### When to Use Collection Expressions

Good Use Cases:
- Creating small fixed collections
- Combining existing collections
- Method parameters expecting collections

Avoid When:
- Building collections dynamically in loops
- Performance-critical large collection operations

### Type Aliases Guidelines

Good Use Cases:
- Complex generic types used frequently
- Tuple types with semantic meaning
- Improving code readability

Avoid When:
- Simple types that are already clear
- Types used only once or twice
- Would confuse rather than clarify

---

## Migration Guide

### From Traditional to Primary Constructors

```csharp
// Before
public class Service
{
    private readonly IDependency _dep;
    
    public Service(IDependency dep)
    {
        _dep = dep;
    }
    
    public void DoWork() => _dep.Execute();
}

// After
public class Service(IDependency dep)
{
    public void DoWork() => dep.Execute();
}
```

### From Object Initializers to Collection Expressions

```csharp
// Before
var list = new List<int> { 1, 2, 3 };
var array = new int[] { 1, 2, 3 };

// After
List<int> list = [1, 2, 3];
int[] array = [1, 2, 3];
```

---

Version: 2.0.0
Last Updated: 2026-01-06
