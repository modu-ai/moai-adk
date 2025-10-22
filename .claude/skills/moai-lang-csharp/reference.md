# moai-lang-csharp - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Installation

```bash
# Install .NET SDK
# Download from: https://dotnet.microsoft.com/download

# macOS (Homebrew)
brew install --cask dotnet-sdk

# Verify installation
dotnet --version  # Should show 9.0.100+

# Install .NET tools globally
dotnet tool install --global dotnet-ef         # Entity Framework
dotnet tool install --global dotnet-format     # Code formatter
dotnet tool install --global dotnet-reportgenerator-globaltool  # Coverage reports
```

### Common Commands

```bash
# Create new project
dotnet new console -n MyApp           # Console app
dotnet new classlib -n MyLib          # Class library
dotnet new xunit -n MyTests           # xUnit test project
dotnet new sln -n MySolution          # Solution file

# Build and run
dotnet build                          # Build project
dotnet run                            # Run project
dotnet watch run                      # Run with hot reload

# Testing
dotnet test                           # Run all tests
dotnet test --filter "FullyQualifiedName~Calculator"  # Filter tests
dotnet test --logger "console;verbosity=detailed"     # Verbose output

# Code coverage
dotnet test /p:CollectCoverage=true
dotnet test /p:CollectCoverage=true /p:CoverletOutputFormat=lcov
dotnet test /p:CollectCoverage=true /p:Threshold=85

# Package management
dotnet add package Newtonsoft.Json    # Add NuGet package
dotnet remove package Newtonsoft.Json # Remove package
dotnet list package                   # List packages
dotnet restore                        # Restore dependencies

# Format code
dotnet format                         # Format all files
dotnet format --verify-no-changes     # Check formatting (CI)

# Clean
dotnet clean                          # Clean build artifacts
```

## Tool Versions (2025-10-22)

- **C#**: 13.0 - Latest language version
- **.NET**: 9.0.100 - LTS release
- **xUnit**: 2.9.3 - Modern testing framework
- **Moq**: 4.20.72 - Mocking library
- **Coverlet**: 6.0.2 - Code coverage tool

## Official Documentation Links

- **C# Language**: https://learn.microsoft.com/en-us/dotnet/csharp/
- **.NET Documentation**: https://learn.microsoft.com/en-us/dotnet/
- **xUnit**: https://xunit.net/
- **Moq**: https://github.com/moq/moq4
- **NuGet**: https://www.nuget.org/

## C# 13 / .NET 9 Features

### C# 13 Key Features
- **Params collections**: `params ReadOnlySpan<T>`
- **Ref struct improvements**: More flexible ref struct usage
- **Lock object**: New `System.Threading.Lock` type
- **Escape character improvements**: `\e` escape sequence
- **Implicit indexer access**: Simpler collection expressions

### .NET 9 Highlights
- **Performance improvements**: Up to 30% faster for some workloads
- **Parallel testing by default**: `dotnet test` runs parallel
- **Native AOT enhancements**: Better ahead-of-time compilation
- **AI/ML integration**: Enhanced ML.NET support
- **LINQ performance**: Optimized query execution

## xUnit Framework

### Core Components
- **xunit**: Testing framework
- **xunit.runner.visualstudio**: Visual Studio integration
- **xunit.assert**: Assertion library (included)

### Test Attributes

```csharp
[Fact]                     // Single test
[Theory]                   // Parameterized test
[InlineData(1, 2)]        // Test data inline
[MemberData(nameof(Data))] // Test data from member
[ClassData(typeof(Data))]  // Test data from class
[Trait("Category", "Unit")] // Test categorization
[Skip("Reason")]           // Skip test
```

### Common Assertions

```csharp
// Equality
Assert.Equal(expected, actual);
Assert.NotEqual(expected, actual);
Assert.Same(expected, actual);         // Reference equality
Assert.NotSame(expected, actual);

// Boolean
Assert.True(condition);
Assert.False(condition);

// Null checks
Assert.Null(obj);
Assert.NotNull(obj);

// Collections
Assert.Contains(item, collection);
Assert.DoesNotContain(item, collection);
Assert.Empty(collection);
Assert.NotEmpty(collection);
Assert.All(collection, item => Assert.True(item > 0));
Assert.Single(collection);             // Exactly one item

// Strings
Assert.StartsWith("Hello", text);
Assert.EndsWith("World", text);
Assert.Contains("test", text);
Assert.DoesNotContain("bad", text);
Assert.Matches(@"\d+", text);          // Regex match

// Numeric (with tolerance)
Assert.Equal(expected, actual, precision: 2);

// Ranges
Assert.InRange(value, low, high);
Assert.NotInRange(value, low, high);

// Types
Assert.IsType<MyType>(obj);
Assert.IsAssignableFrom<IEnumerable>(obj);

// Exceptions
Assert.Throws<ArgumentException>(() => method());
Assert.ThrowsAsync<ArgumentException>(async () => await methodAsync());
```

### Test Lifecycle

```csharp
public class MyTests : IDisposable
{
    // Constructor runs before each test
    public MyTests()
    {
        // Setup code
    }

    // Dispose runs after each test
    public void Dispose()
    {
        // Cleanup code
    }

    [Fact]
    public void Test_Something()
    {
        // Test code
    }
}
```

### Test Fixtures (Shared Context)

```csharp
public class DatabaseFixture : IDisposable
{
    public DbContext Context { get; }

    public DatabaseFixture()
    {
        // Expensive setup (once per test class)
        Context = CreateDbContext();
    }

    public void Dispose()
    {
        Context.Dispose();
    }
}

public class MyTests : IClassFixture<DatabaseFixture>
{
    private readonly DatabaseFixture _fixture;

    public MyTests(DatabaseFixture fixture)
    {
        _fixture = fixture;
    }
}
```

### Collection Fixtures (Shared across multiple classes)

```csharp
[CollectionDefinition("Database collection")]
public class DatabaseCollection : ICollectionFixture<DatabaseFixture> { }

[Collection("Database collection")]
public class Tests1 { }

[Collection("Database collection")]
public class Tests2 { }
```

## Moq Framework

### Basic Mocking

```csharp
var mock = new Mock<IMyInterface>();

// Setup method return value
mock.Setup(x => x.GetValue()).Returns(42);
mock.Setup(x => x.GetValueAsync()).ReturnsAsync(42);

// Setup with parameters
mock.Setup(x => x.Add(It.IsAny<int>(), It.IsAny<int>()))
    .Returns((int a, int b) => a + b);

// Setup property
mock.Setup(x => x.Count).Returns(10);
mock.SetupProperty(x => x.Name);  // Auto-implemented

// Verify method was called
mock.Verify(x => x.GetValue(), Times.Once);
mock.Verify(x => x.GetValue(), Times.Exactly(2));
mock.Verify(x => x.GetValue(), Times.Never);
mock.Verify(x => x.GetValue(), Times.AtLeastOnce);

// Verify with specific arguments
mock.Verify(x => x.Add(1, 2), Times.Once);
mock.Verify(x => x.Add(It.IsAny<int>(), It.Is<int>(y => y > 0)), Times.Once);
```

### Argument Matchers

```csharp
It.IsAny<T>()                    // Any value of type T
It.Is<T>(x => x > 5)             // Custom predicate
It.IsIn(1, 2, 3)                 // Value in set
It.IsNotIn(4, 5, 6)              // Value not in set
It.IsInRange(1, 10, Range.Inclusive)
It.IsRegex(@"\d+")               // String matches regex
```

### Sequential Returns

```csharp
mock.SetupSequence(x => x.GetValue())
    .Returns(1)
    .Returns(2)
    .Throws<Exception>();
```

### Callback Actions

```csharp
var called = false;
mock.Setup(x => x.DoSomething())
    .Callback(() => called = true)
    .Returns(true);
```

## .NET CLI Project Templates

### Available Templates

```bash
dotnet new list                  # List all templates

# Common templates
dotnet new console               # Console application
dotnet new classlib              # Class library
dotnet new web                   # ASP.NET Core empty web
dotnet new webapi                # ASP.NET Core Web API
dotnet new mvc                   # ASP.NET Core MVC
dotnet new razor                 # Razor Pages
dotnet new blazor                # Blazor app
dotnet new xunit                 # xUnit test project
dotnet new nunit                 # NUnit test project
dotnet new mstest                # MSTest project
```

## Code Coverage with Coverlet

### Configuration (.csproj)

```xml
<ItemGroup>
  <PackageReference Include="coverlet.collector" Version="6.0.2">
    <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
    <PrivateAssets>all</PrivateAssets>
  </PackageReference>
</ItemGroup>
```

### Generate Coverage Report

```bash
# Basic coverage
dotnet test /p:CollectCoverage=true

# Generate lcov format
dotnet test /p:CollectCoverage=true /p:CoverletOutputFormat=lcov

# Multiple formats
dotnet test /p:CollectCoverage=true /p:CoverletOutputFormat=\"json,lcov,cobertura\"

# Exclude from coverage
dotnet test /p:CollectCoverage=true /p:Exclude="[xunit.*]*"

# Set coverage threshold (fail if below)
dotnet test /p:CollectCoverage=true /p:Threshold=85
```

### Generate HTML Report

```bash
# Install ReportGenerator
dotnet tool install --global dotnet-reportgenerator-globaltool

# Generate report
reportgenerator \
  "-reports:coverage.info" \
  "-targetdir:coverage-report" \
  "-reporttypes:Html"
```

## C# Best Practices

### Naming Conventions
- **Classes/Interfaces/Records**: `PascalCase`
- **Methods**: `PascalCase`
- **Properties**: `PascalCase`
- **Local variables**: `camelCase`
- **Private fields**: `_camelCase` (with underscore)
- **Constants**: `PascalCase` or `UPPER_SNAKE_CASE`
- **Interfaces**: `IPascalCase` (I prefix)

### Modern C# Patterns

```csharp
// Nullable reference types
#nullable enable
string? nullable = null;
string nonNullable = "hello";

// Pattern matching
var result = value switch
{
    0 => "zero",
    > 0 => "positive",
    < 0 => "negative",
    _ => "unknown"
};

// Records for immutable data
public record User(string Name, int Age);

// Init-only properties
public class Person
{
    public string Name { get; init; }
    public int Age { get; init; }
}

// File-scoped namespaces (C# 10+)
namespace MyApp;  // No braces needed

// Global using (C# 10+)
// Add to any .cs file or GlobalUsings.cs
global using System.Linq;

// Required properties (C# 11+)
public required string Name { get; init; }

// Raw string literals (C# 11+)
var json = """
{
    "name": "Alice",
    "age": 30
}
""";

// Primary constructors (C# 12+)
public class Logger(ILoggerService service)
{
    public void Log(string message) => service.Log(message);
}
```

### Async/Await Best Practices

```csharp
// Always use async/await for I/O operations
public async Task<string> FetchDataAsync()
{
    using var client = new HttpClient();
    return await client.GetStringAsync("https://api.example.com");
}

// ConfigureAwait(false) for library code
await Task.Delay(100).ConfigureAwait(false);

// Avoid async void (except event handlers)
public async Task MyMethodAsync()  // ✓ GOOD
public async void MyMethodAsync()  // ✗ BAD

// Use Task.WhenAll for parallel operations
var tasks = urls.Select(url => FetchDataAsync(url));
var results = await Task.WhenAll(tasks);
```

### Dependency Injection

```csharp
// Program.cs (.NET 6+)
var builder = WebApplication.CreateBuilder(args);

// Register services
builder.Services.AddScoped<IUserService, UserService>();
builder.Services.AddSingleton<IConfiguration>(builder.Configuration);
builder.Services.AddTransient<IEmailService, EmailService>();

var app = builder.Build();
app.Run();

// Constructor injection
public class UserController
{
    private readonly IUserService _userService;

    public UserController(IUserService userService)
    {
        _userService = userService;
    }
}
```

## Testing Guidelines

### Arrange-Act-Assert Pattern

```csharp
[Fact]
public void Add_TwoNumbers_ReturnsSum()
{
    // Arrange
    var calculator = new Calculator();
    int a = 2, b = 3;

    // Act
    var result = calculator.Add(a, b);

    // Assert
    Assert.Equal(5, result);
}
```

### Test Naming Convention

```
MethodName_StateUnderTest_ExpectedBehavior
```

Examples:
- `Add_TwoPositiveNumbers_ReturnsSum`
- `Divide_ByZero_ThrowsException`
- `GetUser_NonExistentId_ReturnsNull`

### Test Coverage Target

- **Minimum**: 85% line coverage
- **Ideal**: 90%+ line coverage
- Focus on business logic, not trivial getters/setters

## Common Pitfalls

### Pitfall 1: Ignoring Null Reference Warnings

```csharp
// ❌ BAD
string GetName() => null;  // Warning CS8603

// ✅ GOOD
string? GetName() => null;  // Explicitly nullable
```

### Pitfall 2: Not Disposing Resources

```csharp
// ❌ BAD
var stream = File.OpenRead("file.txt");
// ... forget to dispose

// ✅ GOOD
using var stream = File.OpenRead("file.txt");
// Automatically disposed
```

### Pitfall 3: Blocking Async Code

```csharp
// ❌ BAD
var result = MyMethodAsync().Result;  // Potential deadlock

// ✅ GOOD
var result = await MyMethodAsync();
```

### Pitfall 4: Catching Generic Exceptions

```csharp
// ❌ BAD
try {
    DoSomething();
} catch (Exception) {  // Too broad
    // ...
}

// ✅ GOOD
try {
    DoSomething();
} catch (InvalidOperationException ex) {  // Specific
    // ...
}
```

---

_For working examples, see [examples.md](examples.md)_
