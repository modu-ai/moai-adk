---
name: moai-lang-csharp
version: 3.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: C# 12 enterprise development with .NET 8, ASP.NET Core, Entity Framework Core, async patterns, and Context7 MCP integration.
keywords: ['csharp', 'dotnet8', 'aspnetcore', 'entityframework', 'async', 'enterprise', 'context7']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# 🚀 C# 12 Enterprise Development Premium Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-csharp |
| **Version** | 3.0.0 (2025-11-06) - Premium Edition |
| **Allowed tools** | Read, Bash, Context7 MCP Integration |
| **Auto-load** | C# projects, .NET applications, enterprise development |
| **Tier** | Premium Language |
| **Context7 Integration** | C# .NET 8 + ASP.NET Core Official Docs |

---

## 🎯 What It Does

**Enterprise-grade C# 12 development** with .NET 8 ecosystem, ASP.NET Core, Entity Framework Core, modern async patterns, and production-ready architecture.

### Core Capabilities

**🏗️ Modern C# 12 Features**:
- **Primary Constructors**: Concise class initialization with constructor parameters
- **Collection Expressions**: Natural syntax for creating collections `[1, 2, 3]`
- **Inline Arrays**: Efficient stack-allocated arrays `Span<int> array = [1, 2, 3];`
- **Lambda Attributes**: `[Attribute] lambda => expression` syntax
- **Ref Structs**: Improved performance with ref readonly struct

**⚡ .NET 8 Enterprise Stack**:
- **ASP.NET Core 8**: Minimal APIs, gRPC, SignalR, Blazor Server/WebAssembly
- **Entity Framework Core 8**: Performance optimizations, raw SQL, compiled queries
- **.NET MAUI**: Cross-platform mobile and desktop development
- **Azure Integration**: App Services, Functions, Blob Storage, Cosmos DB
- **Performance**: ReadyToRun, AOT compilation, tiered compilation

**🔧 Enterprise Development Tools**:
- **IDE**: Visual Studio 2022, Visual Studio Code, Rider
- **Testing**: xUnit 2.6+, FluentAssertions, Moq, testcontainers-dotnet
- **Code Quality**: StyleCop, SonarQube, Code Analysis, Roslyn Analyzers
- **CI/CD**: GitHub Actions, Azure DevOps, Docker, Kubernetes

---

## 🌟 Enterprise Patterns & Best Practices

### Modern C# 12 Primary Constructors

```csharp
// Primary constructor with dependency injection
public class OrderService(
    IOrderRepository orderRepository,
    IPaymentService paymentService,
    ILogger<OrderService> logger) : IOrderService
{
    public async Task<OrderResult> ProcessOrderAsync(OrderRequest request)
    {
        try
        {
            // Repository pattern with async/await
            var order = await orderRepository.CreateAsync(request);

            // Payment processing with structured concurrency
            var paymentTask = paymentService.ProcessPaymentAsync(order.Id, request.Amount);
            var inventoryTask = orderRepository.ReserveInventoryAsync(order.Items);

            await Task.WhenAll(paymentTask, inventoryTask);

            return new OrderResult(order.Id, OrderStatus.Completed);
        }
        catch (Exception ex)
        {
            logger.LogError(ex, "Order processing failed");
            throw new OrderProcessingException("Failed to process order", ex);
        }
    }
}
```

### Collection Expressions and Inline Arrays

```csharp
public class ProductCatalog
{
    private readonly Dictionary<string, Product> _products;
    private readonly Span<Product> _featuredProducts;

    public ProductCatalog()
    {
        // Collection expressions
        _products = new()
        {
            ["laptop"] = new Product("Laptop", 999.99m),
            ["phone"] = new Product("Phone", 699.99m),
            ["tablet"] = new Product("Tablet", 399.99m)
        };

        // Inline array for stack allocation
        _featuredProducts = [_products["laptop"], _products["phone"]];
    }

    public Product[] GetFeaturedProducts() => [.. _featuredProducts];

    public IReadOnlyCollection<string> GetCategories() => ["Electronics", "Computers", "Mobile"];
}
```

### ASP.NET Core 8 Minimal APIs with Advanced Patterns

```csharp
// Program.cs with modern minimal APIs
var builder = WebApplication.CreateBuilder(args);

// Enhanced service configuration
builder.Services
    .AddDbContext<AppDbContext>(options =>
        options.UseNpgsql(builder.Configuration.GetConnectionString("DefaultConnection")))
    .AddScoped<IOrderService, OrderService>()
    .AddScoped<IPaymentService, PaymentService>()
    .AddValidatorsFromAssemblyContaining<OrderRequestValidator>()
    .AddProblemDetails()
    .AddEndpointsApiExplorer()
    .AddSwaggerGen();

// Performance optimizations
builder.Services.Configure<JsonOptions>(options =>
{
    options.JsonSerializerOptions.PropertyNamingPolicy = JsonNamingPolicy.CamelCase;
    options.JsonSerializerOptions.WriteIndented = false;
});

var app = builder.Build();

// Middleware pipeline
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseExceptionHandler();
app.UseHttpsRedirection();

// Advanced minimal API endpoints
app.MapGroup("/api/orders")
    .MapOrdersApi()
    .RequireAuthorization()
    .WithTags("Orders");

app.MapGroup("/api/products")
    .MapProductsApi()
    .WithTags("Products");

app.Run();
```

### Advanced Minimal API Endpoint Groups

```csharp
public static class OrderEndpoints
{
    public static IEndpointRouteBuilder MapOrdersApi(this IEndpointRouteBuilder group)
    {
        // Create order with validation and error handling
        group.MapPost("/", async (
            OrderRequest request,
            IOrderService orderService,
            IValidator<OrderRequest> validator,
            HttpContext httpContext,
            CancellationToken cancellationToken) =>
        {
            var validationResult = await validator.ValidateAsync(request, cancellationToken);

            if (!validationResult.IsValid)
            {
                return Results.ValidationProblem(validationResult.ToDictionary());
            }

            try
            {
                var result = await orderService.ProcessOrderAsync(request, cancellationToken);

                return Results.Created($"/api/orders/{result.OrderId}", new
                {
                    OrderId = result.OrderId,
                    Status = result.Status,
                    CreatedAt = DateTime.UtcNow,
                    Links = new
                    {
                        Self = $"{httpContext.Request.Scheme}://{httpContext.Request.Host}/api/orders/{result.OrderId}",
                        Payment = $"{httpContext.Request.Scheme}://{httpContext.Request.Host}/api/payments/{result.OrderId}"
                    }
                });
            }
            catch (OrderProcessingException ex)
            {
                return Results.Problem(
                    detail: ex.Message,
                    instance: httpContext.Request.Path,
                    statusCode: StatusCodes.Status400BadRequest,
                    title: "Order Processing Failed"
                );
            }
        })
        .WithName("CreateOrder")
        .Accepts<OrderRequest>("application/json")
        .Produces<OrderResponse>(StatusCodes.Status201Created)
        .ProducesProblem(StatusCodes.Status400BadRequest)
        .ProducesProblem(StatusCodes.Status500InternalServerError);

        // Get order with projection
        group.MapGet("/{orderId:guid}", async (
            Guid orderId,
            IOrderService orderService,
            IMapper mapper,
            CancellationToken cancellationToken) =>
        {
            var order = await orderService.GetOrderAsync(orderId, cancellationToken);

            if (order == null)
            {
                return Results.NotFound();
            }

            var orderDto = mapper.Map<OrderDto>(order);
            return Results.Ok(orderDto);
        })
        .WithName("GetOrder")
        .Produces<OrderDto>(StatusCodes.Status200OK)
        .Produces(StatusCodes.Status404NotFound);

        // Advanced search with filtering and pagination
        group.MapGet("/", async (
            [AsParameters] OrderQueryParameters parameters,
            IOrderService orderService,
            CancellationToken cancellationToken) =>
        {
            var (orders, totalCount) = await orderService.SearchOrdersAsync(parameters, cancellationToken);

            var orderDtos = orders.Select(o => new OrderDto
            {
                Id = o.Id,
                CustomerId = o.CustomerId,
                TotalAmount = o.TotalAmount,
                Status = o.Status.ToString(),
                CreatedAt = o.CreatedAt
            }).ToList();

            var pagination = new PaginationMetadata(
                totalCount,
                parameters.Page,
                parameters.PageSize);

            return Results.Ok(new PaginatedResponse<OrderDto>(orderDtos, pagination));
        })
        .WithName("SearchOrders")
        .Produces<PaginatedResponse<OrderDto>>(StatusCodes.Status200OK);

        return group;
    }
}
```

### Entity Framework Core 8 Advanced Patterns

```csharp
// Advanced entity configuration with performance optimizations
public class OrderConfiguration : IEntityTypeConfiguration<Order>
{
    public void Configure(EntityTypeBuilder<Order> builder)
    {
        // Primary key configuration
        builder.HasKey(o => o.Id);

        // Optimistic concurrency
        builder.Property(o => o.RowVersion)
            .IsRowVersion()
            .IsConcurrencyToken();

        // Indexes for performance
        builder.HasIndex(o => o.CustomerId)
            .HasDatabaseName("IX_Orders_CustomerId");

        builder.HasIndex(o => new { o.Status, o.CreatedAt })
            .HasDatabaseName("IX_Orders_Status_CreatedAt");

        // Value conversions
        builder.Property(o => o.TotalAmount)
            .HasPrecision(18, 2)
            .HasConversion<decimal>();

        builder.Property(o => o.Status)
            .HasConversion<string>();

        // Complex type configuration
        builder.ComplexProperty(o => o.ShippingAddress, address =>
        {
            address.Property(a => a.Street).HasMaxLength(200);
            address.Property(a => a.City).HasMaxLength(100);
            address.Property(a => a.PostalCode).HasMaxLength(20);
        });

        // Query filters for soft delete
        builder.HasQueryFilter(o => !o.IsDeleted);

        // Table configuration
        builder.ToTable("Orders", tb => tb.HasTrigger("Orders_AuditTrigger"));
    }
}

// Repository pattern with compiled queries for performance
public class OrderRepository : IOrderRepository
{
    private readonly AppDbContext _context;

    // Pre-compiled query for performance
    private static readonly Func<AppDbContext, Guid, Task<Order?>> GetOrderByIdQuery =
        EF.CompileAsyncQuery((AppDbContext context, Guid id) =>
            context.Orders
                .Include(o => o.Customer)
                .Include(o => o.Items)
                    .ThenInclude(i => i.Product)
                .FirstOrDefaultAsync(o => o.Id == id));

    public OrderRepository(AppDbContext context)
    {
        _context = context;
    }

    public async Task<Order?> GetOrderAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await GetOrderByIdQuery(_context, id);
    }

    public async Task<(IReadOnlyList<Order> Orders, int TotalCount)> SearchOrdersAsync(
        OrderQueryParameters parameters,
        CancellationToken cancellationToken = default)
    {
        var query = _context.Orders
            .Include(o => o.Customer)
            .Include(o => o.Items)
                .ThenInclude(i => i.Product)
            .AsQueryable();

        // Apply filters
        if (parameters.CustomerId.HasValue)
        {
            query = query.Where(o => o.CustomerId == parameters.CustomerId);
        }

        if (parameters.Status.HasValue)
        {
            query = query.Where(o => o.Status == parameters.Status.Value);
        }

        if (parameters.DateFrom.HasValue)
        {
            query = query.Where(o => o.CreatedAt >= parameters.DateFrom.Value);
        }

        if (parameters.DateTo.HasValue)
        {
            query = query.Where(o => o.CreatedAt <= parameters.DateTo.Value);
        }

        // Get total count
        var totalCount = await query.CountAsync(cancellationToken);

        // Apply pagination and ordering
        var orders = await query
            .OrderByDescending(o => o.CreatedAt)
            .Skip((parameters.Page - 1) * parameters.PageSize)
            .Take(parameters.PageSize)
            .ToListAsync(cancellationToken);

        return (orders, totalCount);
    }

    public async Task<Order> CreateAsync(Order order, CancellationToken cancellationToken = default)
    {
        _context.Orders.Add(order);
        await _context.SaveChangesAsync(cancellationToken);
        return order;
    }

    public async Task UpdateAsync(Order order, CancellationToken cancellationToken = default)
    {
        _context.Entry(order).State = EntityState.Modified;
        await _context.SaveChangesAsync(cancellationToken);
    }

    public async Task DeleteAsync(Guid id, CancellationToken cancellationToken = default)
    {
        var order = await _context.Orders.FindAsync(new object[] { id }, cancellationToken);
        if (order != null)
        {
            order.IsDeleted = true;
            await _context.SaveChangesAsync(cancellationToken);
        }
    }
}
```

---

## 🔧 Modern Development Workflow

### .csproj with Latest .NET 8 Features

```xml
<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <PublishReadyToRun>true</PublishReadyToRun>
    <PublishTrimmed>false</PublishTrimmed>
  </PropertyGroup>

  <ItemGroup>
    <!-- ASP.NET Core packages -->
    <PackageReference Include="Microsoft.AspNetCore.OpenApi" Version="8.0.8" />
    <PackageReference Include="Microsoft.EntityFrameworkCore.Design" Version="8.0.8">
      <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
      <PrivateAssets>all</PrivateAssets>
    </PackageReference>
    <PackageReference Include="Microsoft.EntityFrameworkCore.PostgreSQL" Version="8.0.8" />
    <PackageReference Include="Microsoft.EntityFrameworkCore.Tools" Version="8.0.8">
      <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
      <PrivateAssets>all</PrivateAssets>
    </PackageReference>

    <!-- Testing packages -->
    <PackageReference Include="Microsoft.AspNetCore.Mvc.Testing" Version="8.0.8" />
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.11.1" />
    <PackageReference Include="xunit" Version="2.9.2" />
    <PackageReference Include="xunit.runner.visualstudio" Version="2.8.2">
      <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
      <PrivateAssets>all</PrivateAssets>
    </PackageReference>
    <PackageReference Include="FluentAssertions" Version="6.12.1" />
    <PackageReference Include="Moq" Version="4.20.72" />
    <PackageReference Include="Testcontainers.PostgreSql" Version="3.10.0" />

    <!-- Additional packages -->
    <PackageReference Include="Serilog.AspNetCore" Version="8.0.2" />
    <PackageReference Include="Serilog.Sinks.Postgresql" Version="2.3.0" />
    <PackageReference Include="FluentValidation.AspNetCore" Version="11.3.0" />
    <PackageReference Include="AutoMapper.Extensions.Microsoft.DependencyInjection" Version="12.0.1" />
    <PackageReference Include="Swashbuckle.AspNetCore" Version="6.6.2" />
  </ItemGroup>

</Project>
```

### Advanced Testing with xUnit and Testcontainers

```csharp
public class OrderServiceIntegrationTests : IClassFixture<PostgreSqlContainer>, IDisposable
{
    private readonly PostgreSqlContainer _postgresContainer;
    private readonly AppDbContext _dbContext;
    private readonly IOrderService _orderService;
    private readonly ServiceProvider _serviceProvider;

    public OrderServiceIntegrationTests(PostgreSqlContainer postgresContainer)
    {
        _postgresContainer = postgresContainer;
        _postgresContainer.StartAsync().Wait();

        var services = new ServiceCollection();

        services.AddDbContext<AppDbContext>(options =>
            options.UseNpgsql(_postgresContainer.GetConnectionString()));

        services.AddScoped<IOrderService, OrderService>();
        services.AddAutoMapper(typeof(MappingProfile));

        _serviceProvider = services.BuildServiceProvider();
        _dbContext = _serviceProvider.GetRequiredService<AppDbContext>();
        _orderService = _serviceProvider.GetRequiredService<IOrderService>();

        _dbContext.Database.EnsureCreated();
    }

    [Fact]
    public async Task ProcessOrderAsync_ShouldCreateOrderAndProcessPayment_WhenValidRequest()
    {
        // Arrange
        var request = new OrderRequest
        {
            CustomerId = Guid.NewGuid(),
            Items = new[]
            {
                new OrderItemRequest { ProductId = Guid.NewGuid(), Quantity = 2, Price = 99.99m }
            }
        };

        // Act
        var result = await _orderService.ProcessOrderAsync(request);

        // Assert
        result.Should().NotBeNull();
        result.OrderId.Should().NotBeEmpty();
        result.Status.Should().Be(OrderStatus.Completed);

        // Verify in database
        var order = await _dbContext.Orders
            .Include(o => o.Items)
            .FirstOrDefaultAsync(o => o.Id == result.OrderId);

        order.Should().NotBeNull();
        order!.Items.Should().HaveCount(1);
        order.TotalAmount.Should().Be(199.98m);
    }

    [Fact]
    public async Task ProcessOrderAsync_ShouldThrowException_WhenInsufficientInventory()
    {
        // Arrange
        var request = new OrderRequest
        {
            CustomerId = Guid.NewGuid(),
            Items = new[]
            {
                new OrderItemRequest { ProductId = Guid.NewGuid(), Quantity = 1000, Price = 99.99m }
            }
        };

        // Act & Assert
        await _orderService
            .Invoking(x => x.ProcessOrderAsync(request))
            .Should()
            .ThrowAsync<InsufficientInventoryException>()
            .WithMessage("*Insufficient inventory*");
    }

    public void Dispose()
    {
        _dbContext?.Dispose();
        _postgresContainer?.DisposeAsync().AsTask().Wait();
    }
}
```

---

## 📊 Performance Optimization Strategies

### Compiled Queries and Caching

```csharp
public class OrderQueryService
{
    private readonly AppDbContext _context;
    private readonly IMemoryCache _cache;

    // Pre-compiled query for frequent operations
    private static readonly Func<AppDbContext, Guid, Task<CustomerOrderSummary>>
        GetCustomerOrderSummaryQuery = EF.CompileAsyncQuery(
            (AppDbContext context, Guid customerId) =>
                context.Orders
                    .Where(o => o.CustomerId == customerId && !o.IsDeleted)
                    .GroupBy(o => o.CustomerId)
                    .Select(g => new CustomerOrderSummary
                    {
                        CustomerId = g.Key,
                        TotalOrders = g.Count(),
                        TotalAmount = g.Sum(o => o.TotalAmount),
                        LastOrderDate = g.Max(o => o.CreatedAt)
                    })
                    .FirstOrDefaultAsync());

    public OrderQueryService(AppDbContext context, IMemoryCache cache)
    {
        _context = context;
        _cache = cache;
    }

    public async Task<CustomerOrderSummary?> GetCustomerOrderSummaryAsync(
        Guid customerId,
        CancellationToken cancellationToken = default)
    {
        var cacheKey = $"CustomerOrderSummary_{customerId}";

        if (_cache.TryGetValue(cacheKey, out CustomerOrderSummary? cached))
        {
            return cached;
        }

        var summary = await GetCustomerOrderSummaryQuery(_context, customerId);

        if (summary != null)
        {
            _cache.Set(cacheKey, summary, TimeSpan.FromMinutes(5));
        }

        return summary;
    }
}
```

### Memory Optimization with Span<T> and Array Pooling

```csharp
public class DataProcessor
{
    public async Task<ProcessedData> ProcessLargeDataAsync(
        IAsyncEnumerable<RawData> rawData,
        CancellationToken cancellationToken = default)
    {
        var buffer = ArrayPool<byte>.Shared.Rent(8192);
        var results = new List<ProcessedData>();

        try
        {
            await foreach (var data in rawData.WithCancellation(cancellationToken))
            {
                // Use Span<T> for zero-allocation operations
                var dataSpan = data.RawBytes.AsSpan();

                // Process data without allocations
                var processed = ProcessDataInPlace(dataSpan, buffer.AsSpan());
                results.Add(processed);
            }

            return new ProcessedData(results);
        }
        finally
        {
            ArrayPool<byte>.Shared.Return(buffer);
        }
    }

    private ProcessedItem ProcessDataInPlace(ReadOnlySpan<byte> input, Span<byte> output)
    {
        // Zero-allocation data processing
        // ... processing logic
        return new ProcessedItem(/* processed data */);
    }
}
```

---

## 🔒 Security Best Practices

### ASP.NET Core Security Configuration

```csharp
public static class SecurityConfiguration
{
    public static IServiceCollection AddApplicationSecurity(this IServiceCollection services, IConfiguration configuration)
    {
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
                ValidIssuer = configuration["Jwt:Issuer"],
                ValidAudience = configuration["Jwt:Audience"],
                IssuerSigningKey = new SymmetricSecurityKey(
                    Encoding.UTF8.GetBytes(configuration["Jwt:SecretKey"]!))
            };
        });

        services.AddAuthorization(options =>
        {
            options.AddPolicy("RequireAdminRole", policy =>
                policy.RequireRole("Admin"));

            options.AddPolicy("RequireCustomerAccess", policy =>
                policy.RequireClaim("permission", "customer:read"));
        });

        // Rate limiting
        services.AddRateLimiter(options =>
        {
            options.GlobalLimiter = PartitionedRateLimiter.Create<HttpContext, string>(
                httpContext =>
                    RateLimitPartition.GetSlidingWindowLimiter(
                        partitionKey: httpContext.User?.Identity?.Name ?? httpContext.Request.RemoteIpAddress?.ToString() ?? "anonymous",
                        factory: partition => new SlidingWindowRateLimiterOptions
                        {
                            PermitLimit = 100,
                            Window = TimeSpan.FromMinutes(1),
                            SegmentsPerWindow = 10
                        }));
        });

        // CORS configuration
        services.AddCors(options =>
        {
            options.AddPolicy("DefaultPolicy", builder =>
            {
                builder.WithOrigins(configuration.GetSection("Cors:AllowedOrigins").Get<string[]>())
                       .WithMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
                       .WithHeaders("Content-Type", "Authorization")
                       .AllowCredentials();
            });
        });

        return services;
    }
}
```

---

## 📈 Monitoring & Observability

### Advanced Logging with Serilog

```csharp
public static class LoggingConfiguration
{
    public static IHostBuilder ConfigureLogging(this IHostBuilder hostBuilder, IConfiguration configuration)
    {
        return hostBuilder.UseSerilog((context, services, loggerConfiguration) =>
        {
            loggerConfiguration
                .ReadFrom.Configuration(configuration)
                .ReadFrom.Services(services)
                .Enrich.FromLogContext()
                .Enrich.WithMachineName()
                .Enrich.WithEnvironmentName()
                .WriteTo.Console()
                .WriteTo.PostgreSQL(
                    connectionString: configuration.GetConnectionString("DefaultConnection"),
                    tableName: "Logs",
                    needAutoCreateTable: true,
                    columnOptions: new ColumnOptions
                    {
                        Timestamp = { ConvertToUtc = true },
                        Level = { EnumAs存储 = true }
                    })
                .WriteTo.Seq("http://localhost:5341");

            if (context.HostingEnvironment.IsProduction())
            {
                loggerConfiguration.MinimumLevel.Warning();
            }
            else
            {
                loggerConfiguration.MinimumLevel.Debug();
            }
        });
    }
}

// Custom enricher for request tracking
public class RequestTrackingEnricher : ILogEventEnricher
{
    private readonly IHttpContextAccessor _httpContextAccessor;

    public RequestTrackingEnricher(IHttpContextAccessor httpContextAccessor)
    {
        _httpContextAccessor = httpContextAccessor;
    }

    public void Enrich(LogEvent logEvent, ILogEventPropertyFactory propertyFactory)
    {
        var httpContext = _httpContextAccessor.HttpContext;

        if (httpContext != null)
        {
            var requestId = httpContext.TraceIdentifier;
            var userId = httpContext.User?.FindFirst("sub")?.Value;

            logEvent.AddPropertyIfAbsent(
                propertyFactory.CreateProperty("RequestId", requestId));

            if (!string.IsNullOrEmpty(userId))
            {
                logEvent.AddPropertyIfAbsent(
                    propertyFactory.CreateProperty("UserId", userId));
            }
        }
    }
}
```

---

## 🔄 Context7 MCP Integration

### Real-time Documentation Access

```csharp
public class CSharpDocumentationService
{
    private static readonly string CSHARP_API = "/websites/learn_microsoft-en-us-dotnet-csharp";
    private readonly IContext7Client _context7Client;

    public CSharpDocumentationService(IContext7Client context7Client)
    {
        _context7Client = context7Client;
    }

    public async Task<string> GetAsyncPatternsAsync(string pattern = "async await Task")
    {
        return await _context7Client.GetLibraryDocsAsync(
            CSHARP_API,
            $"asynchronous programming {pattern}"
        );
    }

    public async Task<string> GetEntityFrameworkPatternsAsync()
    {
        return await _context7Client.GetLibraryDocsAsync(
            CSHARP_API,
            "Entity Framework Core patterns compiled queries"
        );
    }

    public async Task<string> GetAspnetCoreMinimalApisAsync()
    {
        return await _context7Client.GetLibraryDocsAsync(
            CSHARP_API,
            "ASP.NET Core minimal APIs MapGet MapPost"
        );
    }

    public async Task<string> GetLatestBestPracticeAsync(string feature)
    {
        // Always get the most current documentation
        return await _context7Client.GetLibraryDocsAsync(CSHARP_API, feature);
    }
}
```

---

## 📚 Progressive Disclosure Examples

### High Freedom (Quick Answer - 15 tokens)
"Use C# 12 with .NET 8 and ASP.NET Core 8 minimal APIs for high-performance web applications."

### Medium Freedom (Detailed Guidance - 35 tokens)
"Implement primary constructors for dependency injection, use collection expressions `[1, 2, 3]`, enable nullable reference types, and configure EF Core with compiled queries for performance."

### Low Freedom (Comprehensive Implementation - 80 tokens)
"Configure ASP.NET Core 8 with minimal API endpoints, implement CQRS with MediatR, use EF Core 8 compiled queries with `EF.CompileAsyncQuery()`, enable ready-to-run compilation, add Serilog for structured logging, and implement comprehensive exception handling with Problem Details."

---

## 🎯 Works Well With

### Core MoAI Skills
- `Skill("moai-domain-backend")` - Backend architecture patterns
- `Skill("moai-domain-security")` - Enterprise security patterns
- `Skill("moai-foundation-trust")` - TRUST 5 compliance

### Microsoft Ecosystem
- **Azure**: App Services, Functions, Cosmos DB, Blob Storage
- **Databases**: SQL Server, PostgreSQL, Cosmos DB, Redis
- **DevOps**: Azure DevOps, GitHub Actions, Docker
- **Monitoring**: Application Insights, Prometheus, Grafana

---

## 🚀 Production Deployment

### Docker Configuration

```dockerfile
# Multi-stage build for .NET 8
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

COPY ["src/Api/Api.csproj", "src/Api/"]
RUN dotnet restore "src/Api/Api.csproj"

COPY . .
WORKDIR "/src/src/Api"
RUN dotnet publish "Api.csproj" -c Release -o /app/publish

# Production image
FROM mcr.microsoft.com/dotnet/aspnet:8.0
WORKDIR /app
COPY --from=build /app/publish .
ENTRYPOINT ["dotnet", "Api.dll"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: csharp-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: csharp-api
  template:
    metadata:
      labels:
        app: csharp-api
    spec:
      containers:
      - name: csharp-api
        image: csharp-api:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: ASPNETCORE_ENVIRONMENT
          value: "Production"
        - name: ConnectionStrings__DefaultConnection
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-connection
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

## ✅ Quality Assurance Checklist

- [ ] **C# 12 Modern Features**: Primary constructors, collection expressions, inline arrays
- [ ] **.NET 8 Ecosystem**: ASP.NET Core 8, EF Core 8, minimal APIs
- [ ] **Context7 Integration**: Real-time documentation from official Microsoft sources
- [ ] **Performance Optimization**: Compiled queries, ready-to-run, tiered compilation
- [ ] **Security Configuration**: JWT authentication, authorization, rate limiting
- [ ] **Testing Coverage**: xUnit, FluentAssertions, testcontainers, integration tests
- [ ] **Observability**: Serilog, Application Insights, structured logging
- [ ] **Production Readiness**: Docker, Kubernetes, CI/CD pipeline
- [ ] **Code Quality**: StyleCop, SonarQube, nullable reference types

---

**Last Updated**: 2025-11-06
**Version**: 3.0.0 (Premium Edition - C# 12 + .NET 8)
**Context7 Integration**: Fully integrated with .NET 8 official APIs
**Status**: Production Ready - Enterprise Grade