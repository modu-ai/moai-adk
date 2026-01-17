# C# 12 / .NET 8 Production Examples

Production-ready code examples, full-stack patterns, and testing templates.

---

## Complete Web API Example

### Program.cs (Entry Point)

```csharp
using Microsoft.EntityFrameworkCore;
using FluentValidation;
using MediatR;

var builder = WebApplication.CreateBuilder(args);

// Database
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("Default")));

// MediatR and FluentValidation
builder.Services.AddMediatR(cfg => cfg.RegisterServicesFromAssembly(typeof(Program).Assembly));
builder.Services.AddValidatorsFromAssemblyContaining<Program>();

// Services
builder.Services.AddScoped<IUserRepository, EfUserRepository>();
builder.Services.AddScoped<IPasswordHasher, PasswordHasher>();

// Authentication
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

builder.Services.AddAuthorization();
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseHttpsRedirection();
app.UseAuthentication();
app.UseAuthorization();

// Map endpoints
app.MapUserEndpoints();

app.Run();
```

### User Endpoints

```csharp
public static class UserEndpoints
{
    public static void MapUserEndpoints(this WebApplication app)
    {
        var group = app.MapGroup("/api/users")
            .WithTags("Users")
            .WithOpenApi();

        group.MapGet("/", async (IMediator mediator) =>
        {
            var users = await mediator.Send(new GetAllUsersQuery());
            return Results.Ok(users);
        })
        .WithName("GetAllUsers")
        .Produces<List<UserDto>>();

        group.MapGet("/{id:guid}", async (Guid id, IMediator mediator) =>
        {
            var user = await mediator.Send(new GetUserByIdQuery(id));
            return user is not null ? Results.Ok(user) : Results.NotFound();
        })
        .WithName("GetUserById")
        .Produces<UserDto>()
        .Produces(StatusCodes.Status404NotFound);

        group.MapPost("/", async (CreateUserCommand command, IMediator mediator) =>
        {
            var user = await mediator.Send(command);
            return Results.Created($"/api/users/{user.Id}", user);
        })
        .WithName("CreateUser")
        .Produces<UserDto>(StatusCodes.Status201Created)
        .ProducesValidationProblem();

        group.MapPut("/{id:guid}", async (Guid id, UpdateUserCommand command, IMediator mediator) =>
        {
            if (id != command.Id)
                return Results.BadRequest("ID mismatch");
            
            var result = await mediator.Send(command);
            return result ? Results.NoContent() : Results.NotFound();
        })
        .WithName("UpdateUser")
        .Produces(StatusCodes.Status204NoContent)
        .Produces(StatusCodes.Status404NotFound);

        group.MapDelete("/{id:guid}", async (Guid id, IMediator mediator) =>
        {
            var result = await mediator.Send(new DeleteUserCommand(id));
            return result ? Results.NoContent() : Results.NotFound();
        })
        .WithName("DeleteUser")
        .Produces(StatusCodes.Status204NoContent)
        .Produces(StatusCodes.Status404NotFound)
        .RequireAuthorization("Admin");
    }
}
```

---

## Domain Entities

### User Entity

```csharp
public class User
{
    public Guid Id { get; private set; }
    public string Name { get; private set; } = string.Empty;
    public string Email { get; private set; } = string.Empty;
    public string PasswordHash { get; private set; } = string.Empty;
    public string Role { get; private set; } = "User";
    public DateTime CreatedAt { get; private set; }
    public DateTime? UpdatedAt { get; private set; }
    public bool IsActive { get; private set; } = true;

    private readonly List<Post> _posts = [];
    public IReadOnlyCollection<Post> Posts => _posts.AsReadOnly();

    private User() { } // EF Core constructor

    public static User Create(string name, string email, string passwordHash)
    {
        return new User
        {
            Id = Guid.NewGuid(),
            Name = name,
            Email = email,
            PasswordHash = passwordHash,
            CreatedAt = DateTime.UtcNow
        };
    }

    public void Update(string name, string email)
    {
        Name = name;
        Email = email;
        UpdatedAt = DateTime.UtcNow;
    }

    public void Deactivate()
    {
        IsActive = false;
        UpdatedAt = DateTime.UtcNow;
    }

    public void AddPost(Post post)
    {
        _posts.Add(post);
    }
}
```

### Post Entity

```csharp
public class Post
{
    public Guid Id { get; private set; }
    public string Title { get; private set; } = string.Empty;
    public string Content { get; private set; } = string.Empty;
    public Guid AuthorId { get; private set; }
    public User Author { get; private set; } = null!;
    public DateTime CreatedAt { get; private set; }
    public DateTime? PublishedAt { get; private set; }
    public PostStatus Status { get; private set; } = PostStatus.Draft;

    private readonly List<Tag> _tags = [];
    public IReadOnlyCollection<Tag> Tags => _tags.AsReadOnly();

    private Post() { }

    public static Post Create(string title, string content, User author)
    {
        var post = new Post
        {
            Id = Guid.NewGuid(),
            Title = title,
            Content = content,
            AuthorId = author.Id,
            Author = author,
            CreatedAt = DateTime.UtcNow
        };
        author.AddPost(post);
        return post;
    }

    public void Publish()
    {
        if (Status == PostStatus.Draft)
        {
            Status = PostStatus.Published;
            PublishedAt = DateTime.UtcNow;
        }
    }

    public void AddTag(Tag tag)
    {
        if (!_tags.Contains(tag))
            _tags.Add(tag);
    }
}

public enum PostStatus
{
    Draft,
    Published,
    Archived
}
```

---

## CQRS Handlers

### Query Handler

```csharp
public record GetUserByIdQuery(Guid Id) : IRequest<UserDto?>;

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
```

### Command Handler with Validation

```csharp
public record CreateUserCommand(string Name, string Email, string Password) : IRequest<UserDto>;

public class CreateUserCommandValidator : AbstractValidator<CreateUserCommand>
{
    public CreateUserCommandValidator(IUserRepository userRepository)
    {
        RuleFor(x => x.Name)
            .NotEmpty().WithMessage("Name is required")
            .MaximumLength(100).WithMessage("Name cannot exceed 100 characters");

        RuleFor(x => x.Email)
            .NotEmpty().WithMessage("Email is required")
            .EmailAdddess().WithMessage("Invalid email format")
            .MustAsync(async (email, ct) => !await userRepository.EmailExistsAsync(email, ct))
            .WithMessage("Email already exists");

        RuleFor(x => x.Password)
            .NotEmpty().WithMessage("Password is required")
            .MinimumLength(8).WithMessage("Password must be at least 8 characters")
            .Matches(@"[A-Z]").WithMessage("Password must contain uppercase letter")
            .Matches(@"[a-z]").WithMessage("Password must contain lowercase letter")
            .Matches(@"[0-9]").WithMessage("Password must contain digit");
    }
}

public class CreateUserCommandHandler(
    AppDbContext context,
    IPasswordHasher passwordHasher,
    IValidator<CreateUserCommand> validator)
    : IRequestHandler<CreateUserCommand, UserDto>
{
    public async Task<UserDto> Handle(CreateUserCommand request, CancellationToken ct)
    {
        var validationResult = await validator.ValidateAsync(request, ct);
        if (!validationResult.IsValid)
            throw new ValidationException(validationResult.Errors);

        var user = User.Create(
            request.Name,
            request.Email,
            passwordHasher.Hash(request.Password));

        context.Users.Add(user);
        await context.SaveChangesAsync(ct);

        return new UserDto(user.Id, user.Name, user.Email, user.Role, user.CreatedAt);
    }
}
```

---

## Repository Pattern

### Repository Interface

```csharp
public interface IRepository<T> where T : class
{
    Task<T?> GetByIdAsync(Guid id, CancellationToken ct = default);
    Task<IReadOnlyList<T>> GetAllAsync(CancellationToken ct = default);
    Task<T> AddAsync(T entity, CancellationToken ct = default);
    Task UpdateAsync(T entity, CancellationToken ct = default);
    Task DeleteAsync(T entity, CancellationToken ct = default);
    Task<bool> ExistsAsync(Guid id, CancellationToken ct = default);
}

public interface IUserRepository : IRepository<User>
{
    Task<User?> GetByEmailAsync(string email, CancellationToken ct = default);
    Task<bool> EmailExistsAsync(string email, CancellationToken ct = default);
    Task<IReadOnlyList<User>> GetActiveUsersAsync(CancellationToken ct = default);
}
```

### Repository Implementation

```csharp
public class EfRepository<T>(AppDbContext context) : IRepository<T> where T : class
{
    protected readonly AppDbContext Context = context;
    protected readonly DbSet<T> DbSet = context.Set<T>();

    public virtual async Task<T?> GetByIdAsync(Guid id, CancellationToken ct = default)
        => await DbSet.FindAsync([id], ct);

    public virtual async Task<IReadOnlyList<T>> GetAllAsync(CancellationToken ct = default)
        => await DbSet.ToListAsync(ct);

    public virtual async Task<T> AddAsync(T entity, CancellationToken ct = default)
    {
        await DbSet.AddAsync(entity, ct);
        await Context.SaveChangesAsync(ct);
        return entity;
    }

    public virtual async Task UpdateAsync(T entity, CancellationToken ct = default)
    {
        DbSet.Update(entity);
        await Context.SaveChangesAsync(ct);
    }

    public virtual async Task DeleteAsync(T entity, CancellationToken ct = default)
    {
        DbSet.Remove(entity);
        await Context.SaveChangesAsync(ct);
    }

    public virtual async Task<bool> ExistsAsync(Guid id, CancellationToken ct = default)
        => await DbSet.FindAsync([id], ct) is not null;
}

public class EfUserRepository(AppDbContext context) 
    : EfRepository<User>(context), IUserRepository
{
    public async Task<User?> GetByEmailAsync(string email, CancellationToken ct = default)
        => await DbSet
            .Include(u => u.Posts)
            .FirstOrDefaultAsync(u => u.Email == email, ct);

    public async Task<bool> EmailExistsAsync(string email, CancellationToken ct = default)
        => await DbSet.AnyAsync(u => u.Email == email, ct);

    public async Task<IReadOnlyList<User>> GetActiveUsersAsync(CancellationToken ct = default)
        => await DbSet
            .AsNoTracking()
            .Where(u => u.IsActive)
            .OrderBy(u => u.Name)
            .ToListAsync(ct);
}
```

---

## Testing Examples

### Unit Test with xUnit and Moq

```csharp
public class CreateUserCommandHandlerTests
{
    private readonly Mock<AppDbContext> _contextMock;
    private readonly Mock<IPasswordHasher> _hasherMock;
    private readonly Mock<IValidator<CreateUserCommand>> _validatorMock;
    private readonly CreateUserCommandHandler _handler;

    public CreateUserCommandHandlerTests()
    {
        _contextMock = new Mock<AppDbContext>(new DbContextOptions<AppDbContext>());
        _hasherMock = new Mock<IPasswordHasher>();
        _validatorMock = new Mock<IValidator<CreateUserCommand>>();
        
        _handler = new CreateUserCommandHandler(
            _contextMock.Object,
            _hasherMock.Object,
            _validatorMock.Object);
    }

    [Fact]
    public async Task Handle_ValidCommand_ReturnsUserDto()
    {
        // Arrange
        var command = new CreateUserCommand("John Doe", "john@example.com", "Password123");
        
        _validatorMock.Setup(v => v.ValidateAsync(command, It.IsAny<CancellationToken>()))
            .ReturnsAsync(new ValidationResult());
        
        _hasherMock.Setup(h => h.Hash(command.Password))
            .Returns("hashed_password");

        var mockSet = new Mock<DbSet<User>>();
        _contextMock.Setup(c => c.Users).Returns(mockSet.Object);

        // Act
        var result = await _handler.Handle(command, CancellationToken.None);

        // Assert
        result.Should().NotBeNull();
        result.Name.Should().Be("John Doe");
        result.Email.Should().Be("john@example.com");
        
        mockSet.Verify(m => m.Add(It.IsAny<User>()), Times.Once);
        _contextMock.Verify(c => c.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task Handle_InvalidCommand_ThrowsValidationException()
    {
        // Arrange
        var command = new CreateUserCommand("", "invalid-email", "short");
        var failures = new List<ValidationFailure>
        {
            new("Name", "Name is required"),
            new("Email", "Invalid email format")
        };
        
        _validatorMock.Setup(v => v.ValidateAsync(command, It.IsAny<CancellationToken>()))
            .ReturnsAsync(new ValidationResult(failures));

        // Act & Assert
        await Assert.ThrowsAsync<ValidationException>(() =>
            _handler.Handle(command, CancellationToken.None));
    }
}
```

### Integration Test with TestServer

```csharp
public class UsersEndpointsTests : IClassFixture<WebApplicationFactory<Program>>
{
    private readonly HttpClient _client;
    private readonly WebApplicationFactory<Program> _factory;

    public UsersEndpointsTests(WebApplicationFactory<Program> factory)
    {
        _factory = factory.WithWebHostBuilder(builder =>
        {
            builder.ConfigureServices(services =>
            {
                // Replace real database with in-memory
                var descriptor = services.SingleOrDefault(
                    d => d.ServiceType == typeof(DbContextOptions<AppDbContext>));
                if (descriptor != null)
                    services.Remove(descriptor);

                services.AddDbContext<AppDbContext>(options =>
                    options.UseInMemoryDatabase("TestDb"));
            });
        });
        
        _client = _factory.CreateClient();
    }

    [Fact]
    public async Task CreateUser_ValidRequest_ReturnsCreated()
    {
        // Arrange
        var request = new { Name = "Test User", Email = "test@example.com", Password = "Password123" };
        var content = new StringContent(
            JsonSerializer.Serialize(request),
            Encoding.UTF8,
            "application/json");

        // Act
        var response = await _client.PostAsync("/api/users", content);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.Created);
        response.Headers.Location.Should().NotBeNull();
        
        var user = await response.Content.ReadFromJsonAsync<UserDto>();
        user.Should().NotBeNull();
        user!.Name.Should().Be("Test User");
    }

    [Fact]
    public async Task GetUserById_ExistingUser_ReturnsOk()
    {
        // Arrange
        using var scope = _factory.Services.CreateScope();
        var context = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        var user = User.Create("Existing User", "existing@example.com", "hash");
        context.Users.Add(user);
        await context.SaveChangesAsync();

        // Act
        var response = await _client.GetAsync($"/api/users/{user.Id}");

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
        var result = await response.Content.ReadFromJsonAsync<UserDto>();
        result!.Id.Should().Be(user.Id);
    }

    [Fact]
    public async Task GetUserById_NonExistingUser_ReturnsNotFound()
    {
        // Act
        var response = await _client.GetAsync($"/api/users/{Guid.NewGuid()}");

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.NotFound);
    }
}
```

---

## DTOs and Records

```csharp
// User DTOs
public record UserDto(Guid Id, string Name, string Email, string Role, DateTime CreatedAt);
public record CreateUserRequest(string Name, string Email, string Password);
public record UpdateUserRequest(string Name, string Email);
public record UserWithPostsDto(Guid Id, string Name, string Email, List<PostSummaryDto> Posts);

// Post DTOs
public record PostDto(Guid Id, string Title, string Content, DateTime CreatedAt, PostStatus Status);
public record PostSummaryDto(Guid Id, string Title, DateTime CreatedAt);
public record CreatePostRequest(string Title, string Content);

// Pagination
public record PagedResult<T>(List<T> Items, int TotalCount, int Page, int PageSize)
{
    public int TotalPages => (int)Math.Ceiling(TotalCount / (double)PageSize);
    public bool HasPreviousPage => Page > 1;
    public bool HasNextPage => Page < TotalPages;
}
```

---

Version: 2.0.0
Last Updated: 2026-01-06
