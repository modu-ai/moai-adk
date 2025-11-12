# moai-lang-csharp - Working Examples

_Last updated: 2025-11-12_

## Quick Start (C# 13 Minimal Setup)

```csharp
using System;
using System.Threading.Tasks;

// Modern C# 13 with async/await
public class HelloWorld
{
    public async Task<string> GreetAsync(string name)
    {
        await Task.Delay(100); // Simulate async work
        return $"Hello, {name}!";
    }
}

// Usage
var greeting = new HelloWorld();
var result = await greeting.GreetAsync("C#");
Console.WriteLine(result);  // Output: Hello, C#!
```

## Basic Usage Examples

### Example 1: Async/Await Function - Basic

```csharp
using System;
using System.Net.Http;
using System.Threading.Tasks;

// Simple async function that fetches data
public class UserService
{
    private readonly HttpClient _httpClient = new();

    public async Task<string> FetchUserDataAsync(string userId)
    {
        try
        {
            var url = $"https://jsonplaceholder.typicode.com/users/{userId}";
            var response = await _httpClient.GetAsync(url);
            response.EnsureSuccessStatusCode();

            var jsonContent = await response.Content.ReadAsStringAsync();
            return jsonContent;
        }
        catch (HttpRequestException ex)
        {
            throw new InvalidOperationException($"Failed to fetch user {userId}", ex);
        }
    }
}

// Using the async function
class Program
{
    static async Task Main(string[] args)
    {
        var service = new UserService();
        try
        {
            var userData = await service.FetchUserDataAsync("1");
            Console.WriteLine($"User data fetched: {userData}");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error: {ex.Message}");
        }
    }
}
```

### Example 2: LINQ Query - Data Transformation

```csharp
using System;
using System.Collections.Generic;
using System.Linq;

// Sample data model
public class Product
{
    public int Id { get; set; }
    public string Name { get; set; }
    public decimal Price { get; set; }
    public string Category { get; set; }
}

// LINQ query examples
public class ProductService
{
    private List<Product> _products = new()
    {
        new { Id = 1, Name = "Laptop", Price = 999m, Category = "Electronics" },
        new { Id = 2, Name = "Mouse", Price = 29m, Category = "Electronics" },
        new { Id = 3, Name = "Keyboard", Price = 79m, Category = "Electronics" },
        new { Id = 4, Name = "Book", Price = 15m, Category = "Books" }
    };

    // Query: Filter expensive items
    public IEnumerable<Product> GetExpensiveProducts(decimal minPrice)
    {
        return _products
            .Where(p => p.Price >= minPrice)
            .OrderByDescending(p => p.Price)
            .Select(p => new { p.Name, p.Price, Discount = p.Price * 0.1m });
    }

    // Query: Group by category
    public IEnumerable<IGrouping<string, Product>> GetProductsByCategory()
    {
        return _products
            .GroupBy(p => p.Category)
            .OrderBy(g => g.Key);
    }

    // Query: Complex aggregation
    public decimal GetAveragePriceByCategory(string category)
    {
        return _products
            .Where(p => p.Category == category)
            .Average(p => p.Price);
    }
}

// Usage
var service = new ProductService();
var expensive = service.GetExpensiveProducts(50);
foreach (var item in expensive)
    Console.WriteLine($"{item.Name}: ${item.Price}");
```

### Example 3: Entity Framework Core - Database Access

```csharp
using Microsoft.EntityFrameworkCore;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

// Define DbContext
public class ApplicationDbContext : DbContext
{
    public DbSet<User> Users { get; set; }
    public DbSet<Post> Posts { get; set; }

    protected override void OnConfiguring(DbContextOptionsBuilder options)
    {
        options.UseSqlServer("Data Source=.;Initial Catalog=MyDb;Integrated Security=true");
    }
}

// Model classes
public class User
{
    public int Id { get; set; }
    public string Name { get; set; }
    public string Email { get; set; }
    public List<Post> Posts { get; set; } = new();
}

public class Post
{
    public int Id { get; set; }
    public string Title { get; set; }
    public string Content { get; set; }
    public int UserId { get; set; }
    public User User { get; set; }
}

// Database service
public class UserRepository
{
    private readonly ApplicationDbContext _context;

    public UserRepository(ApplicationDbContext context)
    {
        _context = context;
    }

    // Create
    public async Task<User> CreateUserAsync(string name, string email)
    {
        var user = new User { Name = name, Email = email };
        _context.Users.Add(user);
        await _context.SaveChangesAsync();
        return user;
    }

    // Read
    public async Task<User> GetUserByIdAsync(int id)
    {
        return await _context.Users
            .Include(u => u.Posts)
            .FirstOrDefaultAsync(u => u.Id == id);
    }

    // Read all with filtering
    public async Task<List<User>> GetUsersByNameAsync(string nameFilter)
    {
        return await _context.Users
            .Where(u => u.Name.Contains(nameFilter))
            .OrderBy(u => u.Name)
            .ToListAsync();
    }

    // Update
    public async Task UpdateUserAsync(int id, string newEmail)
    {
        var user = await _context.Users.FindAsync(id);
        if (user != null)
        {
            user.Email = newEmail;
            await _context.SaveChangesAsync();
        }
    }

    // Delete
    public async Task DeleteUserAsync(int id)
    {
        var user = await _context.Users.FindAsync(id);
        if (user != null)
        {
            _context.Users.Remove(user);
            await _context.SaveChangesAsync();
        }
    }
}
```

## Intermediate Patterns

### Example 4: Dependency Injection - Service Setup

```csharp
using Microsoft.Extensions.DependencyInjection;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

// Define service interfaces
public interface IEmailService
{
    Task SendEmailAsync(string to, string subject, string body);
}

public interface INotificationService
{
    Task NotifyUserAsync(int userId, string message);
}

// Implement services
public class EmailService : IEmailService
{
    public async Task SendEmailAsync(string to, string subject, string body)
    {
        // Simulate email sending
        await Task.Delay(100);
        Console.WriteLine($"Email sent to {to}: {subject}");
    }
}

public class NotificationService : INotificationService
{
    private readonly IEmailService _emailService;

    public NotificationService(IEmailService emailService)
    {
        _emailService = emailService;
    }

    public async Task NotifyUserAsync(int userId, string message)
    {
        await _emailService.SendEmailAsync(
            $"user{userId}@example.com",
            "Notification",
            message
        );
    }
}

// DI setup
public class ServiceConfiguration
{
    public static IServiceProvider ConfigureServices()
    {
        var services = new ServiceCollection();

        services.AddScoped<IEmailService, EmailService>();
        services.AddScoped<INotificationService, NotificationService>();

        return services.BuildServiceProvider();
    }
}

// Usage
var serviceProvider = ServiceConfiguration.ConfigureServices();
var notificationService = serviceProvider.GetRequiredService<INotificationService>();
await notificationService.NotifyUserAsync(1, "Welcome!");
```

### Example 5: ASP.NET Core API - REST Endpoint

```csharp
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.DependencyInjection;
using System.Collections.Generic;
using System.Threading.Tasks;

// Model
public class TodoItem
{
    public int Id { get; set; }
    public string Title { get; set; }
    public bool IsCompleted { get; set; }
}

// Controller
[ApiController]
[Route("api/[controller]")]
public class TodoController : ControllerBase
{
    private static List<TodoItem> _todos = new();
    private static int _nextId = 1;

    [HttpGet]
    public ActionResult<IEnumerable<TodoItem>> GetAll()
    {
        return Ok(_todos);
    }

    [HttpGet("{id}")]
    public ActionResult<TodoItem> GetById(int id)
    {
        var todo = _todos.Find(t => t.Id == id);
        if (todo == null)
            return NotFound();
        return Ok(todo);
    }

    [HttpPost]
    public ActionResult<TodoItem> Create([FromBody] TodoItem item)
    {
        item.Id = _nextId++;
        _todos.Add(item);
        return CreatedAtAction(nameof(GetById), new { id = item.Id }, item);
    }

    [HttpPut("{id}")]
    public IActionResult Update(int id, [FromBody] TodoItem item)
    {
        var existing = _todos.Find(t => t.Id == id);
        if (existing == null)
            return NotFound();

        existing.Title = item.Title;
        existing.IsCompleted = item.IsCompleted;
        return NoContent();
    }

    [HttpDelete("{id}")]
    public IActionResult Delete(int id)
    {
        var todo = _todos.Find(t => t.Id == id);
        if (todo == null)
            return NotFound();

        _todos.Remove(todo);
        return NoContent();
    }
}

// Startup configuration
public class Startup
{
    public void ConfigureServices(IServiceCollection services)
    {
        services.AddControllers();
    }

    public void Configure(IApplicationBuilder app)
    {
        app.UseRouting();
        app.UseEndpoints(endpoints =>
        {
            endpoints.MapControllers();
        });
    }
}
```

### Example 6: Reactive Programming with Reactive Extensions

```csharp
using System;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using System.Threading.Tasks;

// Reactive stream example
public class ReactivePipeline
{
    public static void Example()
    {
        // Create subject (publisher)
        var subject = new Subject<int>();

        // Define reactive pipeline
        subject
            .Where(x => x % 2 == 0)        // Filter even numbers
            .Select(x => x * 2)            // Double each number
            .Throttle(TimeSpan.FromMilliseconds(500))  // Debounce
            .Subscribe(
                onNext: x => Console.WriteLine($"Processed: {x}"),
                onError: ex => Console.WriteLine($"Error: {ex.Message}"),
                onCompleted: () => Console.WriteLine("Stream completed")
            );

        // Emit values
        for (int i = 1; i <= 10; i++)
        {
            subject.OnNext(i);
            Task.Delay(100).Wait();
        }

        subject.OnCompleted();
    }
}

// Practical example: User input filtering
public class SearchService
{
    private readonly Subject<string> _searchTerms = new();

    public IObservable<string> GetSearchResults()
    {
        return _searchTerms
            .DistinctUntilChanged()                    // Skip duplicates
            .Throttle(TimeSpan.FromMilliseconds(500))  // Debounce input
            .Select(term => SearchDatabase(term));    // Query database
    }

    public void OnSearchInput(string term)
    {
        _searchTerms.OnNext(term);
    }

    private string SearchDatabase(string term)
    {
        // Simulate database search
        return $"Results for '{term}'";
    }
}
```

## Advanced Patterns

### Example 7: Task Parallelization - Concurrent Operations

```csharp
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Threading.Tasks;

public class ParallelDataProcessor
{
    // Process multiple items concurrently
    public async Task<List<string>> ProcessItemsInParallelAsync(List<string> items)
    {
        var tasks = new List<Task<string>>();

        foreach (var item in items)
        {
            tasks.Add(ProcessItemAsync(item));
        }

        // Wait for all tasks
        var results = await Task.WhenAll(tasks);
        return new List<string>(results);
    }

    // Process with limited concurrency
    public async Task<List<string>> ProcessWithLimitAsync(List<string> items, int maxConcurrency)
    {
        var semaphore = new System.Threading.SemaphoreSlim(maxConcurrency);
        var tasks = new List<Task<string>>();

        foreach (var item in items)
        {
            tasks.Add(ProcessWithSemaphoreAsync(item, semaphore));
        }

        return new List<string>(await Task.WhenAll(tasks));
    }

    private async Task<string> ProcessItemAsync(string item)
    {
        await Task.Delay(100); // Simulate work
        return $"Processed: {item}";
    }

    private async Task<string> ProcessWithSemaphoreAsync(string item, System.Threading.SemaphoreSlim semaphore)
    {
        await semaphore.WaitAsync();
        try
        {
            return await ProcessItemAsync(item);
        }
        finally
        {
            semaphore.Release();
        }
    }

    // Benchmark concurrent vs sequential
    public async Task BenchmarkAsync(List<string> items)
    {
        var sw = Stopwatch.StartNew();

        // Sequential
        var sequential = new List<string>();
        foreach (var item in items)
        {
            sequential.Add(await ProcessItemAsync(item));
        }
        sw.Stop();
        Console.WriteLine($"Sequential: {sw.ElapsedMilliseconds}ms");

        sw.Restart();

        // Parallel
        var parallel = await ProcessItemsInParallelAsync(items);
        sw.Stop();
        Console.WriteLine($"Parallel: {sw.ElapsedMilliseconds}ms");
    }
}
```

### Example 8: Unit Testing with xUnit

```csharp
using Xunit;
using Moq;
using System.Threading.Tasks;

// Service to test
public class CalculatorService
{
    public int Add(int a, int b) => a + b;
    public int Subtract(int a, int b) => a - b;
    public int Multiply(int a, int b) => a * b;
    public int Divide(int a, int b)
    {
        if (b == 0)
            throw new DivideByZeroException();
        return a / b;
    }
}

// Test class
public class CalculatorServiceTests
{
    private readonly CalculatorService _calculator = new();

    [Fact]
    public void Add_TwoPositiveNumbers_ReturnsCorrectSum()
    {
        // Arrange
        int a = 5, b = 3;

        // Act
        var result = _calculator.Add(a, b);

        // Assert
        Assert.Equal(8, result);
    }

    [Theory]
    [InlineData(2, 3, 5)]
    [InlineData(0, 0, 0)]
    [InlineData(-1, 1, 0)]
    public void Add_VariousInputs_ReturnsCorrectSum(int a, int b, int expected)
    {
        var result = _calculator.Add(a, b);
        Assert.Equal(expected, result);
    }

    [Fact]
    public void Divide_ByZero_ThrowsException()
    {
        // Act & Assert
        Assert.Throws<DivideByZeroException>(() => _calculator.Divide(5, 0));
    }

    [Fact]
    public async Task AsyncMethod_ReturnsCorrectly()
    {
        // Arrange
        var mockService = new Mock<IDataService>();
        mockService.Setup(s => s.GetDataAsync()).ReturnsAsync("test data");

        // Act
        var result = await mockService.Object.GetDataAsync();

        // Assert
        Assert.Equal("test data", result);
        mockService.Verify(s => s.GetDataAsync(), Times.Once);
    }
}

// Interface for mocking
public interface IDataService
{
    Task<string> GetDataAsync();
}
```

### Example 9: Advanced LINQ - Complex Queries

```csharp
using System;
using System.Collections.Generic;
using System.Linq;

// Sample models
public class Order
{
    public int Id { get; set; }
    public int CustomerId { get; set; }
    public decimal TotalAmount { get; set; }
    public DateTime OrderDate { get; set; }
    public List<OrderItem> Items { get; set; } = new();
}

public class OrderItem
{
    public int Id { get; set; }
    public int ProductId { get; set; }
    public int Quantity { get; set; }
    public decimal UnitPrice { get; set; }
}

public class Customer
{
    public int Id { get; set; }
    public string Name { get; set; }
    public string Email { get; set; }
}

// Complex LINQ queries
public class OrderAnalytics
{
    private List<Order> _orders;
    private List<Customer> _customers;

    public OrderAnalytics(List<Order> orders, List<Customer> customers)
    {
        _orders = orders;
        _customers = customers;
    }

    // Query 1: Join and group
    public var GetCustomerOrderSummary()
    {
        return _customers
            .Join(
                _orders,
                c => c.Id,
                o => o.CustomerId,
                (c, o) => new { c.Name, o.TotalAmount, o.OrderDate }
            )
            .GroupBy(x => x.Name)
            .Select(g => new
            {
                Customer = g.Key,
                OrderCount = g.Count(),
                TotalSpent = g.Sum(x => x.TotalAmount),
                AverageOrder = g.Average(x => x.TotalAmount)
            })
            .OrderByDescending(x => x.TotalSpent)
            .ToList();
    }

    // Query 2: Nested grouping
    public var GetMonthlySalesReport()
    {
        return _orders
            .GroupBy(o => o.OrderDate.Year)
            .Select(yearGroup => new
            {
                Year = yearGroup.Key,
                Months = yearGroup
                    .GroupBy(o => o.OrderDate.Month)
                    .Select(monthGroup => new
                    {
                        Month = monthGroup.Key,
                        Sales = monthGroup.Sum(o => o.TotalAmount),
                        Count = monthGroup.Count()
                    })
                    .ToList()
            })
            .ToList();
    }

    // Query 3: Complex filtering with multiple conditions
    public List<Order> GetHighValueOrdersInLastQuarter()
    {
        var threeMonthsAgo = DateTime.Now.AddMonths(-3);

        return _orders
            .Where(o => o.OrderDate >= threeMonthsAgo)
            .Where(o => o.TotalAmount > 1000)
            .Where(o => o.Items.Count > 3)
            .OrderByDescending(o => o.TotalAmount)
            .ToList();
    }

    // Query 4: Projection with calculations
    public List<dynamic> GetOrderDetailsWithTax()
    {
        const decimal taxRate = 0.08m;

        return _orders
            .SelectMany(
                o => o.Items,
                (order, item) => new
                {
                    OrderId = order.Id,
                    OrderDate = order.OrderDate,
                    ItemPrice = item.UnitPrice * item.Quantity,
                    Tax = (item.UnitPrice * item.Quantity) * taxRate,
                    Total = (item.UnitPrice * item.Quantity) * (1 + taxRate)
                }
            )
            .Cast<dynamic>()
            .ToList();
    }
}
```

## Context7 Integration Examples

### Example 10: Getting Latest C# Documentation

Use Context7 to access real-time C# documentation:

```python
# Python example for accessing Context7
from context7_client import Context7

# Resolve library ID for C#/.NET
library_id = Context7.resolve_library_id("csharp")
print(f"C# library ID: {library_id}")

# Get latest documentation
docs = Context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="async-await-patterns",
    tokens=5000
)
print(f"Latest C# docs:\n{docs}")
```

### Example 11: Getting .NET Documentation

```python
# Get .NET specific documentation
from context7_client import Context7

library_id = Context7.resolve_library_id("dotnet")
docs = Context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="entity-framework-core",
    tokens=4000
)
print(f".NET patterns:\n{docs}")
```

### Example 12: Getting ASP.NET Core Documentation

```python
# Get ASP.NET Core framework documentation
from context7_client import Context7

library_id = Context7.resolve_library_id("aspnetcore")
docs = Context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="dependency-injection",
    tokens=4000
)
print(f"ASP.NET Core DI:\n{docs}")
```

## Migration Guides

### From C# 11 to C# 13

```csharp
// OLD: Required properties (C# 11)
public class UserOld
{
    public required string Name { get; init; }
    public required string Email { get; init; }
}

// NEW: Primary constructors (C# 12) + Partial properties (C# 13)
public partial class UserNew(string name, string email)
{
    partial string Name { get; set; }
    partial string Email { get; set; }
}

public partial class UserNew
{
    private string _name;
    private string _email;

    public partial string Name
    {
        get => _name;
        set => _name = value;
    }

    public partial string Email
    {
        get => _email;
        set => _email = value;
    }
}
```

### From Callbacks to Async/Await

```csharp
// OLD: Callback-based
public void FetchUserOld(int id, Action<User> callback)
{
    Task.Run(() =>
    {
        var user = new User { Id = id };
        callback(user);
    });
}

// NEW: Async/await
public async Task<User> FetchUserNewAsync(int id)
{
    await Task.Delay(100); // Simulate I/O
    return new User { Id = id };
}

// Usage
var user = await FetchUserNewAsync(1);
```

---

_For advanced patterns and comprehensive reference, see SKILL.md_
