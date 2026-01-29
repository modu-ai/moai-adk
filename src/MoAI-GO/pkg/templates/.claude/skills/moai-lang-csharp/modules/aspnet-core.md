# ASP.NET Core 8 Development

Comprehensive guide to ASP.NET Core 8 including Minimal APIs, Controllers, Middleware, and Authentication.

---

## Minimal APIs

### Basic Endpoint Setup

```csharp
var builder = WebApplication.CreateBuilder(args);

// Add services
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

// Configure pipeline
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseHttpsRedirection();

// Simple endpoints
app.MapGet("/", () => "Hello World!");

app.MapGet("/health", () => Results.Ok(new { status = "healthy", timestamp = DateTime.UtcNow }));

app.Run();
```

### Endpoint Groups

```csharp
public static class UserEndpoints
{
    public static void MapUserEndpoints(this WebApplication app)
    {
        var group = app.MapGroup("/api/users")
            .WithTags("Users")
            .WithOpenApi();

        group.MapGet("/", GetAllUsers);
        group.MapGet("/{id:guid}", GetUserById);
        group.MapPost("/", CreateUser);
        group.MapPut("/{id:guid}", UpdateUser);
        group.MapDelete("/{id:guid}", DeleteUser).RequireAuthorization("Admin");
    }

    private static async Task<IResult> GetAllUsers(
        IUserService service,
        CancellationToken ct)
    {
        var users = await service.GetAllAsync(ct);
        return Results.Ok(users);
    }

    private static async Task<IResult> GetUserById(
        Guid id,
        IUserService service,
        CancellationToken ct)
    {
        var user = await service.GetByIdAsync(id, ct);
        return user is not null ? Results.Ok(user) : Results.NotFound();
    }

    private static async Task<IResult> CreateUser(
        CreateUserRequest request,
        IUserService service,
        CancellationToken ct)
    {
        var user = await service.CreateAsync(request, ct);
        return Results.Created($"/api/users/{user.Id}", user);
    }

    private static async Task<IResult> UpdateUser(
        Guid id,
        UpdateUserRequest request,
        IUserService service,
        CancellationToken ct)
    {
        var success = await service.UpdateAsync(id, request, ct);
        return success ? Results.NoContent() : Results.NotFound();
    }

    private static async Task<IResult> DeleteUser(
        Guid id,
        IUserService service,
        CancellationToken ct)
    {
        var success = await service.DeleteAsync(id, ct);
        return success ? Results.NoContent() : Results.NotFound();
    }
}
```

### Typed Results

```csharp
app.MapGet("/api/users/{id:guid}", async Task<Results<Ok<UserDto>, NotFound>> (
    Guid id,
    IUserService service) =>
{
    var user = await service.GetByIdAsync(id);
    return user is not null
        ? TypedResults.Ok(user)
        : TypedResults.NotFound();
})
.WithName("GetUserById")
.WithOpenApi();

app.MapPost("/api/users", async Task<Results<Created<UserDto>, ValidationProblem>> (
    CreateUserRequest request,
    IValidator<CreateUserRequest> validator,
    IUserService service) =>
{
    var validation = await validator.ValidateAsync(request);
    if (!validation.IsValid)
        return TypedResults.ValidationProblem(validation.ToDictionary());
    
    var user = await service.CreateAsync(request);
    return TypedResults.Created($"/api/users/{user.Id}", user);
});
```

### Request Binding

```csharp
// Route parameters
app.MapGet("/users/{id:guid}", (Guid id) => $"User ID: {id}");

// Query parameters
app.MapGet("/search", (string? q, int page = 1, int size = 10) =>
    $"Query: {q}, Page: {page}, Size: {size}");

// Header values
app.MapGet("/headers", ([FromHeader(Name = "X-Request-Id")] string? requestId) =>
    $"Request ID: {requestId}");

// Body binding
app.MapPost("/users", (CreateUserRequest request) => Results.Ok(request));

// Complex binding
app.MapGet("/complex", (
    [FromQuery] string filter,
    [FromHeader(Name = "Authorization")] string? auth,
    [FromServices] IUserService service) =>
{
    // Use parameters
});
```

---

## Controllers

### Base Controller Setup

```csharp
[ApiController]
[Route("api/[controller]")]
public class UsersController(IUserService userService, ILogger<UsersController> logger)
    : ControllerBase
{
    [HttpGet]
    [ProducesResponseType<List<UserDto>>(StatusCodes.Status200OK)]
    public async Task<ActionResult<List<UserDto>>> GetAll(CancellationToken ct)
    {
        var users = await userService.GetAllAsync(ct);
        return Ok(users);
    }

    [HttpGet("{id:guid}")]
    [ProducesResponseType<UserDto>(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<ActionResult<UserDto>> GetById(Guid id, CancellationToken ct)
    {
        var user = await userService.GetByIdAsync(id, ct);
        if (user is null)
        {
            logger.LogWarning("User {UserId} not found", id);
            return NotFound();
        }
        return user;
    }

    [HttpPost]
    [ProducesResponseType<UserDto>(StatusCodes.Status201Created)]
    [ProducesResponseType<ValidationProblemDetails>(StatusCodes.Status400BadRequest)]
    public async Task<ActionResult<UserDto>> Create(
        [FromBody] CreateUserRequest request,
        CancellationToken ct)
    {
        var user = await userService.CreateAsync(request, ct);
        return CreatedAtAction(nameof(GetById), new { id = user.Id }, user);
    }

    [HttpPut("{id:guid}")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<ActionResult> Update(
        Guid id,
        [FromBody] UpdateUserRequest request,
        CancellationToken ct)
    {
        var success = await userService.UpdateAsync(id, request, ct);
        return success ? NoContent() : NotFound();
    }

    [HttpDelete("{id:guid}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<ActionResult> Delete(Guid id, CancellationToken ct)
    {
        var success = await userService.DeleteAsync(id, ct);
        return success ? NoContent() : NotFound();
    }
}
```

### Model Validation

```csharp
public class CreateUserRequest
{
    [Required]
    [StringLength(100, MinimumLength = 2)]
    public string Name { get; set; } = string.Empty;

    [Required]
    [EmailAdddess]
    public string Email { get; set; } = string.Empty;

    [Required]
    [MinLength(8)]
    [RegularExpression(@"^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$",
        ErrorMessage = "Password must contain uppercase, lowercase, and digit")]
    public string Password { get; set; } = string.Empty;
}

// Custom validation attribute
public class FutureDateAttribute : ValidationAttribute
{
    protected override ValidationResult? IsValid(object? value, ValidationContext context)
    {
        if (value is DateTime date && date <= DateTime.UtcNow)
            return new ValidationResult("Date must be in the future");
        return ValidationResult.Success;
    }
}
```

---

## Middleware

### Custom Middleware

```csharp
public class RequestLoggingMiddleware(RequestDelegate next, ILogger<RequestLoggingMiddleware> logger)
{
    public async Task InvokeAsync(HttpContext context)
    {
        var requestId = Guid.NewGuid().ToString("N")[..8];
        context.Items["RequestId"] = requestId;
        
        logger.LogInformation(
            "Request {RequestId}: {Method} {Path}",
            requestId, context.Request.Method, context.Request.Path);
        
        var sw = Stopwatch.StartNew();
        
        try
        {
            await next(context);
        }
        finally
        {
            sw.Stop();
            logger.LogInformation(
                "Response {RequestId}: {StatusCode} in {ElapsedMs}ms",
                requestId, context.Response.StatusCode, sw.ElapsedMilliseconds);
        }
    }
}

// Extension method for registration
public static class MiddlewareExtensions
{
    public static IApplicationBuilder UseRequestLogging(this IApplicationBuilder app)
        => app.UseMiddleware<RequestLoggingMiddleware>();
}

// Usage in Program.cs
app.UseRequestLogging();
```

### Exception Handling Middleware

```csharp
public class GlobalExceptionMiddleware(
    RequestDelegate next,
    ILogger<GlobalExceptionMiddleware> logger,
    IHostEnvironment env)
{
    public async Task InvokeAsync(HttpContext context)
    {
        try
        {
            await next(context);
        }
        catch (ValidationException ex)
        {
            await HandleValidationExceptionAsync(context, ex);
        }
        catch (NotFoundException ex)
        {
            await HandleNotFoundExceptionAsync(context, ex);
        }
        catch (Exception ex)
        {
            await HandleExceptionAsync(context, ex);
        }
    }

    private Task HandleValidationExceptionAsync(HttpContext context, ValidationException ex)
    {
        context.Response.StatusCode = StatusCodes.Status400BadRequest;
        context.Response.ContentType = "application/problem+json";
        
        var problem = new ValidationProblemDetails(
            ex.Errors.GroupBy(e => e.PropertyName)
                .ToDictionary(g => g.Key, g => g.Select(e => e.ErrorMessage).ToArray()))
        {
            Status = StatusCodes.Status400BadRequest,
            Title = "Validation failed"
        };
        
        return context.Response.WriteAsJsonAsync(problem);
    }

    private Task HandleNotFoundExceptionAsync(HttpContext context, NotFoundException ex)
    {
        context.Response.StatusCode = StatusCodes.Status404NotFound;
        context.Response.ContentType = "application/problem+json";
        
        var problem = new ProblemDetails
        {
            Status = StatusCodes.Status404NotFound,
            Title = "Not found",
            Detail = ex.Message
        };
        
        return context.Response.WriteAsJsonAsync(problem);
    }

    private Task HandleExceptionAsync(HttpContext context, Exception ex)
    {
        logger.LogError(ex, "Unhandled exception");
        
        context.Response.StatusCode = StatusCodes.Status500InternalServerError;
        context.Response.ContentType = "application/problem+json";
        
        var problem = new ProblemDetails
        {
            Status = StatusCodes.Status500InternalServerError,
            Title = "Internal server error",
            Detail = env.IsDevelopment() ? ex.Message : "An unexpected error occurred"
        };
        
        return context.Response.WriteAsJsonAsync(problem);
    }
}
```

---

## Authentication

### JWT Bearer Authentication

```csharp
// Program.cs configuration
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
                Encoding.UTF8.GetBytes(builder.Configuration["Jwt:Key"]!)),
            ClockSkew = TimeSpan.Zero
        };
    });

builder.Services.AddAuthorization();
```

### Token Generation Service

```csharp
public class JwtTokenService(IConfiguration config)
{
    public string GenerateToken(User user)
    {
        var securityKey = new SymmetricSecurityKey(
            Encoding.UTF8.GetBytes(config["Jwt:Key"]!));
        var credentials = new SigningCredentials(securityKey, SecurityAlgorithms.HmacSha256);

        var claims = new[]
        {
            new Claim(ClaimTypes.NameIdentifier, user.Id.ToString()),
            new Claim(ClaimTypes.Email, user.Email),
            new Claim(ClaimTypes.Name, user.Name),
            new Claim(ClaimTypes.Role, user.Role),
            new Claim("permissions", string.Join(",", user.Permissions))
        };

        var token = new JwtSecurityToken(
            issuer: config["Jwt:Issuer"],
            audience: config["Jwt:Audience"],
            claims: claims,
            expires: DateTime.UtcNow.AddMinutes(int.Parse(config["Jwt:ExpirationMinutes"]!)),
            signingCredentials: credentials);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }
    
    public string GenerateRefreshToken()
    {
        var randomBytes = new byte[64];
        using var rng = RandomNumberGenerator.Create();
        rng.GetBytes(randomBytes);
        return Convert.ToBase64String(randomBytes);
    }
}
```

### Authorization Policies

```csharp
builder.Services.AddAuthorization(options =>
{
    // Role-based policies
    options.AddPolicy("Admin", policy =>
        policy.RequireRole("Admin"));
    
    options.AddPolicy("AdminOrManager", policy =>
        policy.RequireRole("Admin", "Manager"));
    
    // Claim-based policies
    options.AddPolicy("CanEdit", policy =>
        policy.RequireClaim("permissions", "edit"));
    
    options.AddPolicy("CanDelete", policy =>
        policy.RequireClaim("permissions", "delete"));
    
    // Custom requirements
    options.AddPolicy("MinimumAge", policy =>
        policy.Requirements.Add(new MinimumAgeRequirement(18)));
    
    // Combined requirements
    options.AddPolicy("SeniorEditor", policy =>
        policy.RequireRole("Editor")
              .RequireClaim("experience_years")
              .RequireAssertion(ctx =>
              {
                  var years = int.Parse(ctx.User.FindFirstValue("experience_years") ?? "0");
                  return years >= 5;
              }));
});
```

### Custom Authorization Handler

```csharp
public class MinimumAgeRequirement(int age) : IAuthorizationRequirement
{
    public int MinimumAge { get; } = age;
}

public class MinimumAgeHandler : AuthorizationHandler<MinimumAgeRequirement>
{
    protected override Task HandleRequirementAsync(
        AuthorizationHandlerContext context,
        MinimumAgeRequirement requirement)
    {
        var birthDateClaim = context.User.FindFirst(c => c.Type == "birthdate");
        
        if (birthDateClaim is null)
            return Task.CompletedTask;
        
        if (DateTime.TryParse(birthDateClaim.Value, out var birthDate))
        {
            var age = DateTime.Today.Year - birthDate.Year;
            if (birthDate.Date > DateTime.Today.AddYears(-age))
                age--;
            
            if (age >= requirement.MinimumAge)
                context.Succeed(requirement);
        }
        
        return Task.CompletedTask;
    }
}

// Register handler
builder.Services.AddSingleton<IAuthorizationHandler, MinimumAgeHandler>();
```

---

## CORS Configuration

```csharp
builder.Services.AddCors(options =>
{
    options.AddPolicy("Development", policy =>
        policy.AllowAnyOrigin()
              .AllowAnyMethod()
              .AllowAnyHeader());
    
    options.AddPolicy("Production", policy =>
        policy.WithOrigins("https://myapp.com", "https://api.myapp.com")
              .WithMethods("GET", "POST", "PUT", "DELETE")
              .WithHeaders("Authorization", "Content-Type")
              .AllowCredentials());
});

// In pipeline
if (app.Environment.IsDevelopment())
    app.UseCors("Development");
else
    app.UseCors("Production");
```

---

## Rate Limiting

```csharp
builder.Services.AddRateLimiter(options =>
{
    options.GlobalLimiter = PartitionedRateLimiter.Create<HttpContext, string>(context =>
        RateLimitPartition.GetFixedWindowLimiter(
            partitionKey: context.User.Identity?.Name ?? context.Request.Headers.Host.ToString(),
            factory: _ => new FixedWindowRateLimiterOptions
            {
                AutoReplenishment = true,
                PermitLimit = 100,
                Window = TimeSpan.FromMinutes(1)
            }));

    options.AddPolicy("api", context =>
        RateLimitPartition.GetTokenBucketLimiter(
            partitionKey: context.User.Identity?.Name ?? context.Connection.RemoteIpAdddess?.ToString() ?? "anonymous",
            factory: _ => new TokenBucketRateLimiterOptions
            {
                TokenLimit = 100,
                ReplenishmentPeriod = TimeSpan.FromSeconds(10),
                TokensPerPeriod = 10
            }));

    options.OnRejected = async (context, token) =>
    {
        context.HttpContext.Response.StatusCode = StatusCodes.Status429TooManyRequests;
        await context.HttpContext.Response.WriteAsJsonAsync(new
        {
            error = "Too many requests",
            retryAfter = context.Lease.TryGetMetadata(MetadataName.RetryAfter, out var retryAfter)
                ? retryAfter.TotalSeconds
                : 60
        }, token);
    };
});

// Apply to endpoints
app.MapGet("/api/data", () => "Data")
   .RequireRateLimiting("api");
```

---

## Health Checks

```csharp
builder.Services.AddHealthChecks()
    .AddCheck("self", () => HealthCheckResult.Healthy())
    .AddDbContextCheck<AppDbContext>()
    .AddRedis(builder.Configuration.GetConnectionString("Redis")!)
    .AddUrlGroup(new Uri("https://api.external.com/health"), "external-api");

app.MapHealthChecks("/health", new HealthCheckOptions
{
    ResponseWriter = async (context, report) =>
    {
        context.Response.ContentType = "application/json";
        var result = new
        {
            status = report.Status.ToString(),
            checks = report.Entries.Select(e => new
            {
                name = e.Key,
                status = e.Value.Status.ToString(),
                duration = e.Value.Duration.TotalMilliseconds
            })
        };
        await context.Response.WriteAsJsonAsync(result);
    }
});
```

---

Version: 2.0.0
Last Updated: 2026-01-06
