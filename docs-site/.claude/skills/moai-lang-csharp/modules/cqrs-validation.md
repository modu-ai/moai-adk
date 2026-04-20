# CQRS and Validation Patterns

Comprehensive guide to MediatR CQRS, FluentValidation, and Handler patterns for clean architecture.

---

## MediatR Setup

### Installation and Configuration

```bash
dotnet add package MediatR
dotnet add package MediatR.Extensions.Microsoft.DependencyInjection
```

```csharp
// Program.cs
builder.Services.AddMediatR(cfg =>
{
    cfg.RegisterServicesFromAssembly(typeof(Program).Assembly);
    
    // Add behaviors (pipeline)
    cfg.AddBehavior<LoggingBehavior<,>>();
    cfg.AddBehavior<ValidationBehavior<,>>();
    cfg.AddBehavior<TransactionBehavior<,>>();
});
```

---

## Commands

### Command Definition

```csharp
// Command with response
public record CreateUserCommand(
    string Name,
    string Email,
    string Password) : IRequest<UserDto>;

// Command without response
public record DeleteUserCommand(Guid Id) : IRequest;

// Command with result type
public record UpdateUserCommand(
    Guid Id,
    string Name,
    string Email) : IRequest<Result<UserDto>>;
```

### Command Handler

```csharp
public class CreateUserCommandHandler(
    AppDbContext context,
    IPasswordHasher passwordHasher,
    IValidator<CreateUserCommand> validator,
    ILogger<CreateUserCommandHandler> logger)
    : IRequestHandler<CreateUserCommand, UserDto>
{
    public async Task<UserDto> Handle(CreateUserCommand request, CancellationToken ct)
    {
        // Validation
        var validationResult = await validator.ValidateAsync(request, ct);
        if (!validationResult.IsValid)
            throw new ValidationException(validationResult.Errors);

        // Check for duplicates
        var exists = await context.Users.AnyAsync(u => u.Email == request.Email, ct);
        if (exists)
            throw new ConflictException($"User with email {request.Email} already exists");

        // Create entity
        var user = User.Create(
            request.Name,
            request.Email,
            passwordHasher.Hash(request.Password));

        context.Users.Add(user);
        await context.SaveChangesAsync(ct);

        logger.LogInformation("Created user {UserId} with email {Email}", user.Id, user.Email);

        return new UserDto(user.Id, user.Name, user.Email, user.Role, user.CreatedAt);
    }
}
```

### Command with Domain Events

```csharp
public record CreateOrderCommand(
    Guid CustomerId,
    List<OrderItemDto> Items) : IRequest<OrderDto>;

public class CreateOrderCommandHandler(
    AppDbContext context,
    IMediator mediator)
    : IRequestHandler<CreateOrderCommand, OrderDto>
{
    public async Task<OrderDto> Handle(CreateOrderCommand request, CancellationToken ct)
    {
        var order = Order.Create(request.CustomerId, request.Items);
        
        context.Orders.Add(order);
        await context.SaveChangesAsync(ct);

        // Publish domain events
        foreach (var domainEvent in order.DomainEvents)
        {
            await mediator.Publish(domainEvent, ct);
        }
        order.ClearDomainEvents();

        return MapToDto(order);
    }
}
```

---

## Queries

### Query Definition

```csharp
// Simple query
public record GetUserByIdQuery(Guid Id) : IRequest<UserDto?>;

// Query with pagination
public record GetUsersQuery(
    int Page = 1,
    int PageSize = 10,
    string? SearchTerm = null,
    string? SortBy = null,
    bool SortDescending = false) : IRequest<PagedResult<UserDto>>;

// Query with includes
public record GetUserWithPostsQuery(Guid Id) : IRequest<UserWithPostsDto?>;
```

### Query Handler

```csharp
public class GetUserByIdQueryHandler(AppDbContext context)
    : IRequestHandler<GetUserByIdQuery, UserDto?>
{
    public async Task<UserDto?> Handle(GetUserByIdQuery request, CancellationToken ct)
    {
        return await context.Users
            .AsNoTracking()
            .Where(u => u.Id == request.Id)
            .Select(u => new UserDto(u.Id, u.Name, u.Email, u.Role, u.CreatedAt))
            .FirstOrDefaultAsync(ct);
    }
}

public class GetUsersQueryHandler(AppDbContext context)
    : IRequestHandler<GetUsersQuery, PagedResult<UserDto>>
{
    public async Task<PagedResult<UserDto>> Handle(GetUsersQuery request, CancellationToken ct)
    {
        var query = context.Users.AsNoTracking();

        // Apply search filter
        if (!string.IsNullOrEmpty(request.SearchTerm))
        {
            query = query.Where(u =>
                u.Name.Contains(request.SearchTerm) ||
                u.Email.Contains(request.SearchTerm));
        }

        // Apply sorting
        query = request.SortBy?.ToLower() switch
        {
            "name" => request.SortDescending
                ? query.OrderByDescending(u => u.Name)
                : query.OrderBy(u => u.Name),
            "email" => request.SortDescending
                ? query.OrderByDescending(u => u.Email)
                : query.OrderBy(u => u.Email),
            "createdat" => request.SortDescending
                ? query.OrderByDescending(u => u.CreatedAt)
                : query.OrderBy(u => u.CreatedAt),
            _ => query.OrderBy(u => u.Name)
        };

        var totalCount = await query.CountAsync(ct);

        var items = await query
            .Skip((request.Page - 1) * request.PageSize)
            .Take(request.PageSize)
            .Select(u => new UserDto(u.Id, u.Name, u.Email, u.Role, u.CreatedAt))
            .ToListAsync(ct);

        return new PagedResult<UserDto>(items, totalCount, request.Page, request.PageSize);
    }
}
```

---

## FluentValidation

### Installation

```bash
dotnet add package FluentValidation
dotnet add package FluentValidation.DependencyInjectionExtensions
```

```csharp
// Program.cs
builder.Services.AddValidatorsFromAssemblyContaining<Program>();
```

### Basic Validators

```csharp
public class CreateUserCommandValidator : AbstractValidator<CreateUserCommand>
{
    public CreateUserCommandValidator()
    {
        RuleFor(x => x.Name)
            .NotEmpty().WithMessage("Name is required")
            .MinimumLength(2).WithMessage("Name must be at least 2 characters")
            .MaximumLength(100).WithMessage("Name cannot exceed 100 characters");

        RuleFor(x => x.Email)
            .NotEmpty().WithMessage("Email is required")
            .EmailAdddess().WithMessage("Invalid email format")
            .MaximumLength(256).WithMessage("Email cannot exceed 256 characters");

        RuleFor(x => x.Password)
            .NotEmpty().WithMessage("Password is required")
            .MinimumLength(8).WithMessage("Password must be at least 8 characters")
            .Matches(@"[A-Z]").WithMessage("Password must contain at least one uppercase letter")
            .Matches(@"[a-z]").WithMessage("Password must contain at least one lowercase letter")
            .Matches(@"[0-9]").WithMessage("Password must contain at least one digit")
            .Matches(@"[^a-zA-Z0-9]").WithMessage("Password must contain at least one special character");
    }
}
```

### Async Validation with Dependencies

```csharp
public class CreateUserCommandValidator : AbstractValidator<CreateUserCommand>
{
    public CreateUserCommandValidator(IUserRepository userRepository)
    {
        RuleFor(x => x.Name)
            .NotEmpty()
            .MaximumLength(100);

        RuleFor(x => x.Email)
            .NotEmpty()
            .EmailAdddess()
            .MustAsync(async (email, ct) => !await userRepository.EmailExistsAsync(email, ct))
            .WithMessage("Email is already in use");

        RuleFor(x => x.Password)
            .SetValidator(new PasswordValidator());
    }
}

// Reusable validator
public class PasswordValidator : AbstractValidator<string>
{
    public PasswordValidator()
    {
        RuleFor(x => x)
            .NotEmpty().WithMessage("Password is required")
            .MinimumLength(8).WithMessage("Password must be at least 8 characters")
            .Matches(@"[A-Z]").WithMessage("Must contain uppercase letter")
            .Matches(@"[a-z]").WithMessage("Must contain lowercase letter")
            .Matches(@"[0-9]").WithMessage("Must contain digit");
    }
}
```

### Complex Validation Rules

```csharp
public class CreateOrderCommandValidator : AbstractValidator<CreateOrderCommand>
{
    public CreateOrderCommandValidator(IProductRepository productRepository)
    {
        RuleFor(x => x.CustomerId)
            .NotEmpty().WithMessage("Customer ID is required");

        RuleFor(x => x.Items)
            .NotEmpty().WithMessage("At least one item is required")
            .Must(items => items.Count <= 50).WithMessage("Maximum 50 items per order");

        RuleForEach(x => x.Items).ChildRules(item =>
        {
            item.RuleFor(i => i.ProductId)
                .NotEmpty()
                .MustAsync(async (id, ct) => await productRepository.ExistsAsync(id, ct))
                .WithMessage("Product not found");

            item.RuleFor(i => i.Quantity)
                .GreaterThan(0).WithMessage("Quantity must be positive")
                .LessThanOrEqualTo(100).WithMessage("Maximum quantity is 100");
        });

        // Cross-field validation
        RuleFor(x => x)
            .MustAsync(async (cmd, ct) =>
            {
                var totalItems = cmd.Items.Sum(i => i.Quantity);
                return totalItems <= 500;
            })
            .WithMessage("Total items cannot exceed 500");
    }
}
```

### Validation Groups

```csharp
public class UpdateUserCommandValidator : AbstractValidator<UpdateUserCommand>
{
    public UpdateUserCommandValidator()
    {
        // Always validate
        RuleFor(x => x.Id).NotEmpty();

        // Validate when updating profile
        RuleSet("Profile", () =>
        {
            RuleFor(x => x.Name).NotEmpty().MaximumLength(100);
            RuleFor(x => x.Bio).MaximumLength(500);
        });

        // Validate when updating email
        RuleSet("Email", () =>
        {
            RuleFor(x => x.Email).NotEmpty().EmailAdddess();
        });
    }
}

// Usage
var result = await validator.ValidateAsync(command, options =>
    options.IncludeRuleSets("Profile"));
```

---

## Pipeline Behaviors

### Validation Behavior

```csharp
public class ValidationBehavior<TRequest, TResponse>(
    IEnumerable<IValidator<TRequest>> validators)
    : IPipelineBehavior<TRequest, TResponse>
    where TRequest : notnull
{
    public async Task<TResponse> Handle(
        TRequest request,
        RequestHandlerDelegate<TResponse> next,
        CancellationToken cancellationToken)
    {
        if (!validators.Any())
            return await next();

        var context = new ValidationContext<TRequest>(request);

        var validationResults = await Task.WhenAll(
            validators.Select(v => v.ValidateAsync(context, cancellationToken)));

        var failures = validationResults
            .SelectMany(r => r.Errors)
            .Where(f => f is not null)
            .ToList();

        if (failures.Count > 0)
            throw new ValidationException(failures);

        return await next();
    }
}
```

### Logging Behavior

```csharp
public class LoggingBehavior<TRequest, TResponse>(
    ILogger<LoggingBehavior<TRequest, TResponse>> logger)
    : IPipelineBehavior<TRequest, TResponse>
    where TRequest : notnull
{
    public async Task<TResponse> Handle(
        TRequest request,
        RequestHandlerDelegate<TResponse> next,
        CancellationToken cancellationToken)
    {
        var requestName = typeof(TRequest).Name;
        var requestId = Guid.NewGuid().ToString("N")[..8];

        logger.LogInformation(
            "Handling {RequestName} ({RequestId})",
            requestName, requestId);

        var sw = Stopwatch.StartNew();

        try
        {
            var response = await next();
            sw.Stop();

            logger.LogInformation(
                "Handled {RequestName} ({RequestId}) in {ElapsedMs}ms",
                requestName, requestId, sw.ElapsedMilliseconds);

            return response;
        }
        catch (Exception ex)
        {
            sw.Stop();
            logger.LogError(
                ex,
                "Error handling {RequestName} ({RequestId}) after {ElapsedMs}ms",
                requestName, requestId, sw.ElapsedMilliseconds);
            throw;
        }
    }
}
```

### Transaction Behavior

```csharp
public class TransactionBehavior<TRequest, TResponse>(
    AppDbContext context,
    ILogger<TransactionBehavior<TRequest, TResponse>> logger)
    : IPipelineBehavior<TRequest, TResponse>
    where TRequest : notnull
{
    public async Task<TResponse> Handle(
        TRequest request,
        RequestHandlerDelegate<TResponse> next,
        CancellationToken cancellationToken)
    {
        // Skip for queries
        if (typeof(TRequest).Name.EndsWith("Query"))
            return await next();

        await using var transaction = await context.Database.BeginTransactionAsync(cancellationToken);

        try
        {
            var response = await next();
            await transaction.CommitAsync(cancellationToken);
            return response;
        }
        catch
        {
            await transaction.RollbackAsync(cancellationToken);
            throw;
        }
    }
}
```

### Caching Behavior

```csharp
public interface ICacheable
{
    string CacheKey { get; }
    TimeSpan? CacheDuration { get; }
}

public class CachingBehavior<TRequest, TResponse>(
    IDistributedCache cache,
    ILogger<CachingBehavior<TRequest, TResponse>> logger)
    : IPipelineBehavior<TRequest, TResponse>
    where TRequest : ICacheable
{
    public async Task<TResponse> Handle(
        TRequest request,
        RequestHandlerDelegate<TResponse> next,
        CancellationToken cancellationToken)
    {
        var cacheKey = request.CacheKey;

        var cachedValue = await cache.GetStringAsync(cacheKey, cancellationToken);
        if (cachedValue is not null)
        {
            logger.LogDebug("Cache hit for {CacheKey}", cacheKey);
            return JsonSerializer.Deserialize<TResponse>(cachedValue)!;
        }

        logger.LogDebug("Cache miss for {CacheKey}", cacheKey);
        var response = await next();

        var options = new DistributedCacheEntryOptions
        {
            AbsoluteExpirationRelativeToNow = request.CacheDuration ?? TimeSpan.FromMinutes(5)
        };

        await cache.SetStringAsync(
            cacheKey,
            JsonSerializer.Serialize(response),
            options,
            cancellationToken);

        return response;
    }
}
```

---

## Notifications (Domain Events)

### Notification Definition

```csharp
public record UserCreatedNotification(Guid UserId, string Email) : INotification;

public record OrderPlacedNotification(Guid OrderId, Guid CustomerId, decimal Total) : INotification;
```

### Notification Handlers

```csharp
public class SendWelcomeEmailHandler(IEmailService emailService)
    : INotificationHandler<UserCreatedNotification>
{
    public async Task Handle(UserCreatedNotification notification, CancellationToken ct)
    {
        await emailService.SendWelcomeEmailAsync(notification.Email, ct);
    }
}

public class UpdateAnalyticsHandler(IAnalyticsService analyticsService)
    : INotificationHandler<UserCreatedNotification>
{
    public async Task Handle(UserCreatedNotification notification, CancellationToken ct)
    {
        await analyticsService.TrackUserCreatedAsync(notification.UserId, ct);
    }
}
```

### Publishing Notifications

```csharp
public class CreateUserCommandHandler(
    AppDbContext context,
    IMediator mediator)
    : IRequestHandler<CreateUserCommand, UserDto>
{
    public async Task<UserDto> Handle(CreateUserCommand request, CancellationToken ct)
    {
        var user = User.Create(request.Name, request.Email, request.Password);
        
        context.Users.Add(user);
        await context.SaveChangesAsync(ct);

        // Publish notification
        await mediator.Publish(new UserCreatedNotification(user.Id, user.Email), ct);

        return MapToDto(user);
    }
}
```

---

## Result Pattern

```csharp
public class Result<T>
{
    public bool IsSuccess { get; }
    public T? Value { get; }
    public string? Error { get; }
    public List<string> ValidationErrors { get; } = [];

    private Result(T value) { IsSuccess = true; Value = value; }
    private Result(string error) { IsSuccess = false; Error = error; }
    private Result(List<string> errors) { IsSuccess = false; ValidationErrors = errors; }

    public static Result<T> Success(T value) => new(value);
    public static Result<T> Failure(string error) => new(error);
    public static Result<T> ValidationFailure(List<string> errors) => new(errors);

    public TResult Match<TResult>(
        Func<T, TResult> onSuccess,
        Func<string, TResult> onFailure)
        => IsSuccess ? onSuccess(Value!) : onFailure(Error ?? string.Join(", ", ValidationErrors));
}

// Usage in handler
public async Task<Result<UserDto>> Handle(CreateUserCommand request, CancellationToken ct)
{
    var validation = await validator.ValidateAsync(request, ct);
    if (!validation.IsValid)
        return Result<UserDto>.ValidationFailure(
            validation.Errors.Select(e => e.ErrorMessage).ToList());

    // ... create user
    return Result<UserDto>.Success(userDto);
}

// Usage in endpoint
app.MapPost("/api/users", async (CreateUserCommand cmd, IMediator mediator) =>
{
    var result = await mediator.Send(cmd);
    return result.Match(
        user => Results.Created($"/api/users/{user.Id}", user),
        error => Results.BadRequest(new { error }));
});
```

---

Version: 2.0.0
Last Updated: 2026-01-06
