# moai-lang-csharp - Working Examples

_Last updated: 2025-10-22_

## Example 1: Basic C# Project Setup with xUnit

### Project Structure
```
my-csharp-project/
├── MyProject.sln
├── src/
│   └── MyProject/
│       ├── MyProject.csproj
│       ├── Calculator.cs
│       └── Program.cs
└── tests/
    └── MyProject.Tests/
        ├── MyProject.Tests.csproj
        └── CalculatorTests.cs
```

### Calculator.cs
```csharp
namespace MyProject;

public class Calculator
{
    public int Add(int a, int b) => a + b;

    public int Subtract(int a, int b) => a - b;

    public double Divide(double a, double b)
    {
        if (b == 0)
            throw new DivideByZeroException("Cannot divide by zero");

        return a / b;
    }
}
```

### CalculatorTests.cs
```csharp
using Xunit;
using MyProject;

namespace MyProject.Tests;

public class CalculatorTests
{
    private readonly Calculator _calculator;

    public CalculatorTests()
    {
        _calculator = new Calculator();
    }

    [Fact]
    public void Add_TwoPositiveNumbers_ReturnsSum()
    {
        // Arrange
        int a = 2, b = 3;

        // Act
        var result = _calculator.Add(a, b);

        // Assert
        Assert.Equal(5, result);
    }

    [Fact]
    public void Add_TwoNegativeNumbers_ReturnsNegativeSum()
    {
        Assert.Equal(-5, _calculator.Add(-2, -3));
    }

    [Fact]
    public void Subtract_Numbers_ReturnsDifference()
    {
        Assert.Equal(1, _calculator.Subtract(3, 2));
    }

    [Fact]
    public void Divide_ValidNumbers_ReturnsQuotient()
    {
        Assert.Equal(2.5, _calculator.Divide(5.0, 2.0));
    }

    [Fact]
    public void Divide_ByZero_ThrowsDivideByZeroException()
    {
        Assert.Throws<DivideByZeroException>(() => _calculator.Divide(5.0, 0.0));
    }

    [Theory]
    [InlineData(1, 2, 3)]
    [InlineData(0, 0, 0)]
    [InlineData(-1, 1, 0)]
    [InlineData(100, 200, 300)]
    public void Add_TheoryTest_ReturnsExpectedSum(int a, int b, int expected)
    {
        var result = _calculator.Add(a, b);
        Assert.Equal(expected, result);
    }
}
```

### MyProject.csproj
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net9.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
  </PropertyGroup>
</Project>
```

### MyProject.Tests.csproj
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net9.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <IsPackable>false</IsPackable>
    <IsTestProject>true</IsTestProject>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="xunit" Version="2.9.3" />
    <PackageReference Include="xunit.runner.visualstudio" Version="2.8.2" />
    <PackageReference Include="coverlet.collector" Version="6.0.2" />
  </ItemGroup>

  <ItemGroup>
    <ProjectReference Include="..\..\src\MyProject\MyProject.csproj" />
  </ItemGroup>
</Project>
```

### Build and Run
```bash
# Create solution
dotnet new sln -n MyProject

# Create projects
dotnet new classlib -n MyProject -o src/MyProject
dotnet new xunit -n MyProject.Tests -o tests/MyProject.Tests

# Add projects to solution
dotnet sln add src/MyProject/MyProject.csproj
dotnet sln add tests/MyProject.Tests/MyProject.Tests.csproj

# Add project reference
cd tests/MyProject.Tests
dotnet add reference ../../src/MyProject/MyProject.csproj

# Run tests
dotnet test

# Run tests with coverage
dotnet test /p:CollectCoverage=true /p:CoverletOutputFormat=lcov
```

---

## Example 2: LINQ and Modern C# Features

### StringUtils.cs
```csharp
namespace MyProject.Utils;

public static class StringUtils
{
    public static string Trim(string? input) => input?.Trim() ?? string.Empty;

    public static IEnumerable<string> FilterNonEmpty(IEnumerable<string?> items)
    {
        return items
            .Where(s => !string.IsNullOrWhiteSpace(s))
            .Select(s => s!.Trim());
    }

    public static string Join(IEnumerable<string> items, string separator = ", ")
    {
        return string.Join(separator, items);
    }

    public static bool IsPalindrome(string input)
    {
        var normalized = input.ToLowerInvariant().Replace(" ", "");
        return normalized.SequenceEqual(normalized.Reverse());
    }
}
```

### StringUtilsTests.cs
```csharp
using Xunit;
using MyProject.Utils;

namespace MyProject.Tests.Utils;

public class StringUtilsTests
{
    [Theory]
    [InlineData("  hello  ", "hello")]
    [InlineData("world", "world")]
    [InlineData("", "")]
    [InlineData(null, "")]
    public void Trim_VariousInputs_ReturnsExpected(string? input, string expected)
    {
        var result = StringUtils.Trim(input);
        Assert.Equal(expected, result);
    }

    [Fact]
    public void FilterNonEmpty_MixedList_ReturnsOnlyNonEmpty()
    {
        var input = new List<string?> { "hello", "", "  ", null, "world" };
        var result = StringUtils.FilterNonEmpty(input).ToList();

        Assert.Equal(2, result.Count);
        Assert.Contains("hello", result);
        Assert.Contains("world", result);
    }

    [Theory]
    [InlineData("racecar", true)]
    [InlineData("A man a plan a canal Panama", true)]
    [InlineData("hello", false)]
    [InlineData("Able was I ere I saw Elba", true)]
    public void IsPalindrome_VariousStrings_ReturnsExpected(string input, bool expected)
    {
        var result = StringUtils.IsPalindrome(input);
        Assert.Equal(expected, result);
    }
}
```

---

## Example 3: Async/Await Testing

### ApiClient.cs
```csharp
namespace MyProject.Services;

public interface IApiClient
{
    Task<string> FetchDataAsync(string url);
}

public class ApiClient : IApiClient
{
    private readonly HttpClient _httpClient;

    public ApiClient(HttpClient httpClient)
    {
        _httpClient = httpClient;
    }

    public async Task<string> FetchDataAsync(string url)
    {
        var response = await _httpClient.GetAsync(url);
        response.EnsureSuccessStatusCode();
        return await response.Content.ReadAsStringAsync();
    }
}
```

### ApiClientTests.cs
```csharp
using Xunit;
using Moq;
using Moq.Protected;
using System.Net;
using MyProject.Services;

namespace MyProject.Tests.Services;

public class ApiClientTests
{
    [Fact]
    public async Task FetchDataAsync_SuccessfulRequest_ReturnsData()
    {
        // Arrange
        var mockHandler = new Mock<HttpMessageHandler>();
        mockHandler.Protected()
            .Setup<Task<HttpResponseMessage>>(
                "SendAsync",
                ItExpr.IsAny<HttpRequestMessage>(),
                ItExpr.IsAny<CancellationToken>())
            .ReturnsAsync(new HttpResponseMessage
            {
                StatusCode = HttpStatusCode.OK,
                Content = new StringContent("Test Data")
            });

        var httpClient = new HttpClient(mockHandler.Object);
        var apiClient = new ApiClient(httpClient);

        // Act
        var result = await apiClient.FetchDataAsync("https://api.example.com/data");

        // Assert
        Assert.Equal("Test Data", result);
    }

    [Fact]
    public async Task FetchDataAsync_FailedRequest_ThrowsHttpRequestException()
    {
        // Arrange
        var mockHandler = new Mock<HttpMessageHandler>();
        mockHandler.Protected()
            .Setup<Task<HttpResponseMessage>>(
                "SendAsync",
                ItExpr.IsAny<HttpRequestMessage>(),
                ItExpr.IsAny<CancellationToken>())
            .ReturnsAsync(new HttpResponseMessage
            {
                StatusCode = HttpStatusCode.InternalServerError
            });

        var httpClient = new HttpClient(mockHandler.Object);
        var apiClient = new ApiClient(httpClient);

        // Act & Assert
        await Assert.ThrowsAsync<HttpRequestException>(
            () => apiClient.FetchDataAsync("https://api.example.com/data"));
    }
}
```

---

## Example 4: TDD Workflow - User Service

### Step 1: RED - Write Failing Tests

```csharp
// UserService.cs (doesn't exist yet)
using Xunit;

namespace MyProject.Tests.Services;

public class UserServiceTests
{
    [Fact]
    public void CreateUser_ValidData_ReturnsUser()
    {
        var service = new UserService();
        var user = service.CreateUser("alice@example.com", "Alice");

        Assert.NotNull(user);
        Assert.Equal("alice@example.com", user.Email);
        Assert.Equal("Alice", user.Name);
    }

    [Fact]
    public void CreateUser_EmptyEmail_ThrowsArgumentException()
    {
        var service = new UserService();
        Assert.Throws<ArgumentException>(() => service.CreateUser("", "Alice"));
    }

    [Fact]
    public void CreateUser_InvalidEmail_ThrowsArgumentException()
    {
        var service = new UserService();
        Assert.Throws<ArgumentException>(() => service.CreateUser("invalid-email", "Alice"));
    }

    [Fact]
    public void GetUser_ExistingUser_ReturnsUser()
    {
        var service = new UserService();
        service.CreateUser("bob@example.com", "Bob");

        var user = service.GetUser("bob@example.com");

        Assert.NotNull(user);
        Assert.Equal("Bob", user.Name);
    }

    [Fact]
    public void GetUser_NonExistentUser_ReturnsNull()
    {
        var service = new UserService();
        var user = service.GetUser("nonexistent@example.com");

        Assert.Null(user);
    }
}
```

**Result**: Compilation fails (UserService doesn't exist)

### Step 2: GREEN - Minimal Implementation

```csharp
// src/MyProject/Services/UserService.cs
using System.Text.RegularExpressions;

namespace MyProject.Services;

public record User(string Email, string Name);

public class UserService
{
    private readonly Dictionary<string, User> _users = new();
    private static readonly Regex EmailRegex = new(@"^[\w\.-]+@[\w\.-]+\.\w+$");

    public User CreateUser(string email, string name)
    {
        if (string.IsNullOrWhiteSpace(email))
            throw new ArgumentException("Email cannot be empty", nameof(email));

        if (!EmailRegex.IsMatch(email))
            throw new ArgumentException("Invalid email format", nameof(email));

        if (string.IsNullOrWhiteSpace(name))
            throw new ArgumentException("Name cannot be empty", nameof(name));

        var user = new User(email, name);
        _users[email] = user;
        return user;
    }

    public User? GetUser(string email)
    {
        return _users.TryGetValue(email, out var user) ? user : null;
    }
}
```

**Result**: All tests pass ✓

### Step 3: REFACTOR - Improve Code Quality

```csharp
// Improved version with interface and better structure
namespace MyProject.Services;

public record User(string Email, string Name)
{
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
}

public interface IUserService
{
    User CreateUser(string email, string name);
    User? GetUser(string email);
    bool DeleteUser(string email);
}

public class UserService : IUserService
{
    private readonly Dictionary<string, User> _users = new();
    private static readonly Regex EmailRegex =
        new(@"^[\w\.-]+@[\w\.-]+\.\w+$", RegexOptions.Compiled);

    public User CreateUser(string email, string name)
    {
        ValidateEmail(email);
        ValidateName(name);

        var user = new User(email.ToLowerInvariant(), name);
        _users[user.Email] = user;
        return user;
    }

    public User? GetUser(string email)
    {
        if (string.IsNullOrWhiteSpace(email))
            return null;

        return _users.TryGetValue(email.ToLowerInvariant(), out var user) ? user : null;
    }

    public bool DeleteUser(string email)
    {
        if (string.IsNullOrWhiteSpace(email))
            return false;

        return _users.Remove(email.ToLowerInvariant());
    }

    private static void ValidateEmail(string email)
    {
        if (string.IsNullOrWhiteSpace(email))
            throw new ArgumentException("Email cannot be empty", nameof(email));

        if (!EmailRegex.IsMatch(email))
            throw new ArgumentException("Invalid email format", nameof(email));
    }

    private static void ValidateName(string name)
    {
        if (string.IsNullOrWhiteSpace(name))
            throw new ArgumentException("Name cannot be empty", nameof(name));
    }
}
```

**Improvements**:
- Added interface for dependency injection
- Extracted validation methods (single responsibility)
- Case-insensitive email lookup
- Added `DeleteUser` method
- Added `CreatedAt` timestamp to User record
- Used compiled regex for performance

**Result**: All tests still pass ✓

---

## Example 5: Dependency Injection with Mocking

### EmailService.cs (Interface)
```csharp
namespace MyProject.Services;

public interface IEmailService
{
    Task<bool> SendEmailAsync(string to, string subject, string body);
}

public class UserRegistration
{
    private readonly IEmailService _emailService;

    public UserRegistration(IEmailService emailService)
    {
        _emailService = emailService;
    }

    public async Task<bool> RegisterUserAsync(string email, string name)
    {
        if (string.IsNullOrWhiteSpace(email) || string.IsNullOrWhiteSpace(name))
            return false;

        var subject = $"Welcome {name}";
        var body = "Thank you for registering!";

        return await _emailService.SendEmailAsync(email, subject, body);
    }
}
```

### UserRegistrationTests.cs
```csharp
using Xunit;
using Moq;
using MyProject.Services;

namespace MyProject.Tests.Services;

public class UserRegistrationTests
{
    [Fact]
    public async Task RegisterUserAsync_ValidData_SendsEmail()
    {
        // Arrange
        var mockEmailService = new Mock<IEmailService>();
        mockEmailService
            .Setup(x => x.SendEmailAsync(
                "alice@example.com",
                "Welcome Alice",
                "Thank you for registering!"))
            .ReturnsAsync(true);

        var registration = new UserRegistration(mockEmailService.Object);

        // Act
        var result = await registration.RegisterUserAsync("alice@example.com", "Alice");

        // Assert
        Assert.True(result);
        mockEmailService.Verify(
            x => x.SendEmailAsync(
                "alice@example.com",
                "Welcome Alice",
                "Thank you for registering!"),
            Times.Once);
    }

    [Fact]
    public async Task RegisterUserAsync_EmptyEmail_ReturnsFalse()
    {
        // Arrange
        var mockEmailService = new Mock<IEmailService>();
        var registration = new UserRegistration(mockEmailService.Object);

        // Act
        var result = await registration.RegisterUserAsync("", "Alice");

        // Assert
        Assert.False(result);
        mockEmailService.Verify(
            x => x.SendEmailAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()),
            Times.Never);
    }

    [Fact]
    public async Task RegisterUserAsync_EmailServiceFails_ReturnsFalse()
    {
        // Arrange
        var mockEmailService = new Mock<IEmailService>();
        mockEmailService
            .Setup(x => x.SendEmailAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .ReturnsAsync(false);

        var registration = new UserRegistration(mockEmailService.Object);

        // Act
        var result = await registration.RegisterUserAsync("bob@example.com", "Bob");

        // Assert
        Assert.False(result);
    }
}
```

---

## Common Testing Patterns

### Collection Tests
```csharp
[Fact]
public void List_Operations_WorkCorrectly()
{
    var list = new List<int> { 1, 2, 3 };

    Assert.Contains(2, list);
    Assert.DoesNotContain(5, list);
    Assert.Equal(3, list.Count);
    Assert.All(list, item => Assert.True(item > 0));
}
```

### String Tests
```csharp
[Fact]
public void String_Operations_WorkCorrectly()
{
    var text = "Hello, World!";

    Assert.Contains("World", text);
    Assert.StartsWith("Hello", text);
    Assert.EndsWith("World!", text);
    Assert.Matches(@"^\w+,\s\w+!$", text);
}
```

### Record Equality Tests
```csharp
public record Person(string Name, int Age);

[Fact]
public void Records_EqualityWorks()
{
    var person1 = new Person("Alice", 30);
    var person2 = new Person("Alice", 30);
    var person3 = new Person("Bob", 25);

    Assert.Equal(person1, person2);  // Value equality
    Assert.NotEqual(person1, person3);
}
```

### Test Fixtures (Shared Context)
```csharp
public class DatabaseFixture : IDisposable
{
    public DbContext Context { get; private set; }

    public DatabaseFixture()
    {
        Context = new DbContext(/* in-memory options */);
        Context.Database.EnsureCreated();
    }

    public void Dispose()
    {
        Context.Database.EnsureDeleted();
        Context.Dispose();
    }
}

public class DatabaseTests : IClassFixture<DatabaseFixture>
{
    private readonly DatabaseFixture _fixture;

    public DatabaseTests(DatabaseFixture fixture)
    {
        _fixture = fixture;
    }

    [Fact]
    public void Database_CanInsertRecord()
    {
        // Use _fixture.Context
    }
}
```

---

_For detailed CLI commands and tool versions, see [reference.md](reference.md)_
