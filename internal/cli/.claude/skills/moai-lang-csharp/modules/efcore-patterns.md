# Entity Framework Core 8 Patterns

Comprehensive guide to Entity Framework Core 8 including DbContext, Repository pattern, Migrations, and Query optimization.

---

## DbContext Configuration

### Basic DbContext

```csharp
public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<User> Users => Set<User>();
    public DbSet<Post> Posts => Set<Post>();
    public DbSet<Tag> Tags => Set<Tag>();
    public DbSet<Comment> Comments => Set<Comment>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        // Apply all configurations from assembly
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(AppDbContext).Assembly);
        
        // Global query filters
        modelBuilder.Entity<User>().HasQueryFilter(u => !u.IsDeleted);
        modelBuilder.Entity<Post>().HasQueryFilter(p => !p.IsDeleted);
    }

    public override Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        // Automatic audit tracking
        foreach (var entry in ChangeTracker.Entries<IAuditable>())
        {
            switch (entry.State)
            {
                case EntityState.Added:
                    entry.Entity.CreatedAt = DateTime.UtcNow;
                    break;
                case EntityState.Modified:
                    entry.Entity.UpdatedAt = DateTime.UtcNow;
                    break;
            }
        }
        
        return base.SaveChangesAsync(cancellationToken);
    }
}
```

### Service Registration

```csharp
// Program.cs
builder.Services.AddDbContext<AppDbContext>(options =>
{
    options.UseSqlServer(
        builder.Configuration.GetConnectionString("Default"),
        sqlOptions =>
        {
            sqlOptions.EnableRetryOnFailure(
                maxRetryCount: 5,
                maxRetryDelay: TimeSpan.FromSeconds(30),
                errorNumbersToAdd: null);
            sqlOptions.CommandTimeout(30);
            sqlOptions.MigrationsAssembly("MyApp.Infrastructure");
        });
    
    if (builder.Environment.IsDevelopment())
    {
        options.EnableSensitiveDataLogging();
        options.EnableDetailedErrors();
    }
});

// For read replicas
builder.Services.AddDbContext<ReadOnlyDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("ReadReplica"))
           .UseQueryTrackingBehavior(QueryTrackingBehavior.NoTracking));
```

---

## Entity Configuration

### Fluent API Configuration

```csharp
public class UserConfiguration : IEntityTypeConfiguration<User>
{
    public void Configure(EntityTypeBuilder<User> builder)
    {
        builder.ToTable("Users");
        
        // Primary key
        builder.HasKey(u => u.Id);
        builder.Property(u => u.Id)
            .HasDefaultValueSql("NEWSEQUENTIALID()");
        
        // Properties
        builder.Property(u => u.Name)
            .HasMaxLength(100)
            .IsRequired();
        
        builder.Property(u => u.Email)
            .HasMaxLength(256)
            .IsRequired();
        
        builder.Property(u => u.PasswordHash)
            .HasMaxLength(256)
            .IsRequired();
        
        // Indexes
        builder.HasIndex(u => u.Email)
            .IsUnique()
            .HasDatabaseName("IX_Users_Email");
        
        builder.HasIndex(u => u.Name)
            .HasDatabaseName("IX_Users_Name");
        
        // Relationships
        builder.HasMany(u => u.Posts)
            .WithOne(p => p.Author)
            .HasForeignKey(p => p.AuthorId)
            .OnDelete(DeleteBehavior.Cascade);
        
        builder.HasMany(u => u.Comments)
            .WithOne(c => c.Author)
            .HasForeignKey(c => c.AuthorId)
            .OnDelete(DeleteBehavior.Restrict);
    }
}

public class PostConfiguration : IEntityTypeConfiguration<Post>
{
    public void Configure(EntityTypeBuilder<Post> builder)
    {
        builder.ToTable("Posts");
        
        builder.HasKey(p => p.Id);
        
        builder.Property(p => p.Title)
            .HasMaxLength(200)
            .IsRequired();
        
        builder.Property(p => p.Content)
            .HasColumnType("nvarchar(max)")
            .IsRequired();
        
        builder.Property(p => p.Status)
            .HasConversion<string>()
            .HasMaxLength(20);
        
        // Many-to-many relationship
        builder.HasMany(p => p.Tags)
            .WithMany(t => t.Posts)
            .UsingEntity<PostTag>(
                j => j.HasOne(pt => pt.Tag).WithMany().HasForeignKey(pt => pt.TagId),
                j => j.HasOne(pt => pt.Post).WithMany().HasForeignKey(pt => pt.PostId),
                j =>
                {
                    j.HasKey(pt => new { pt.PostId, pt.TagId });
                    j.ToTable("PostTags");
                });
    }
}
```

### Value Objects and Owned Types

```csharp
public class Adddess
{
    public string Street { get; private set; } = string.Empty;
    public string City { get; private set; } = string.Empty;
    public string PostalCode { get; private set; } = string.Empty;
    public string Country { get; private set; } = string.Empty;
}

public class AdddessConfiguration : IEntityTypeConfiguration<User>
{
    public void Configure(EntityTypeBuilder<User> builder)
    {
        builder.OwnsOne(u => u.Adddess, adddess =>
        {
            adddess.Property(a => a.Street).HasMaxLength(200);
            adddess.Property(a => a.City).HasMaxLength(100);
            adddess.Property(a => a.PostalCode).HasMaxLength(20);
            adddess.Property(a => a.Country).HasMaxLength(100);
        });
    }
}
```

---

## Repository Pattern

### Generic Repository Interface

```csharp
public interface IRepository<T> where T : class
{
    Task<T?> GetByIdAsync(Guid id, CancellationToken ct = default);
    Task<IReadOnlyList<T>> GetAllAsync(CancellationToken ct = default);
    Task<IReadOnlyList<T>> FindAsync(
        Expression<Func<T, bool>> predicate, 
        CancellationToken ct = default);
    Task<T> AddAsync(T entity, CancellationToken ct = default);
    Task AddRangeAsync(IEnumerable<T> entities, CancellationToken ct = default);
    Task UpdateAsync(T entity, CancellationToken ct = default);
    Task DeleteAsync(T entity, CancellationToken ct = default);
    Task<bool> ExistsAsync(Guid id, CancellationToken ct = default);
    Task<int> CountAsync(CancellationToken ct = default);
    Task<int> CountAsync(Expression<Func<T, bool>> predicate, CancellationToken ct = default);
}

public interface IUnitOfWork : IDisposable
{
    IRepository<User> Users { get; }
    IRepository<Post> Posts { get; }
    Task<int> SaveChangesAsync(CancellationToken ct = default);
    Task BeginTransactionAsync(CancellationToken ct = default);
    Task CommitAsync(CancellationToken ct = default);
    Task RollbackAsync(CancellationToken ct = default);
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

    public virtual async Task<IReadOnlyList<T>> FindAsync(
        Expression<Func<T, bool>> predicate, 
        CancellationToken ct = default)
        => await DbSet.Where(predicate).ToListAsync(ct);

    public virtual async Task<T> AddAsync(T entity, CancellationToken ct = default)
    {
        await DbSet.AddAsync(entity, ct);
        await Context.SaveChangesAsync(ct);
        return entity;
    }

    public virtual async Task AddRangeAsync(IEnumerable<T> entities, CancellationToken ct = default)
    {
        await DbSet.AddRangeAsync(entities, ct);
        await Context.SaveChangesAsync(ct);
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

    public virtual async Task<int> CountAsync(CancellationToken ct = default)
        => await DbSet.CountAsync(ct);

    public virtual async Task<int> CountAsync(
        Expression<Func<T, bool>> predicate, 
        CancellationToken ct = default)
        => await DbSet.CountAsync(predicate, ct);
}
```

### Unit of Work Implementation

```csharp
public class UnitOfWork(AppDbContext context) : IUnitOfWork
{
    private IDbContextTransaction? _transaction;
    private IRepository<User>? _users;
    private IRepository<Post>? _posts;

    public IRepository<User> Users => _users ??= new EfRepository<User>(context);
    public IRepository<Post> Posts => _posts ??= new EfRepository<Post>(context);

    public async Task<int> SaveChangesAsync(CancellationToken ct = default)
        => await context.SaveChangesAsync(ct);

    public async Task BeginTransactionAsync(CancellationToken ct = default)
    {
        _transaction = await context.Database.BeginTransactionAsync(ct);
    }

    public async Task CommitAsync(CancellationToken ct = default)
    {
        if (_transaction is not null)
        {
            await _transaction.CommitAsync(ct);
            await _transaction.DisposeAsync();
            _transaction = null;
        }
    }

    public async Task RollbackAsync(CancellationToken ct = default)
    {
        if (_transaction is not null)
        {
            await _transaction.RollbackAsync(ct);
            await _transaction.DisposeAsync();
            _transaction = null;
        }
    }

    public void Dispose()
    {
        _transaction?.Dispose();
        context.Dispose();
    }
}
```

---

## Migrations

### Creating Migrations

```bash
# Add migration
dotnet ef migrations add InitialCreate

# Add migration with specific context
dotnet ef migrations add AddUserTable --context AppDbContext

# Add migration to specific project
dotnet ef migrations add AddUserTable --project src/MyApp.Infrastructure --startup-project src/MyApp.Api

# Apply migrations
dotnet ef database update

# Update to specific migration
dotnet ef database update InitialCreate

# Rollback (update to previous migration)
dotnet ef database update PreviousMigration

# Generate SQL script
dotnet ef migrations script

# Generate idempotent script (safe to run multiple times)
dotnet ef migrations script --idempotent

# Remove last migration
dotnet ef migrations remove
```

### Migration Best Practices

```csharp
public partial class AddUserTable : Migration
{
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.CreateTable(
            name: "Users",
            columns: table => new
            {
                Id = table.Column<Guid>(nullable: false, defaultValueSql: "NEWSEQUENTIALID()"),
                Name = table.Column<string>(maxLength: 100, nullable: false),
                Email = table.Column<string>(maxLength: 256, nullable: false),
                CreatedAt = table.Column<DateTime>(nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_Users", x => x.Id);
            });

        migrationBuilder.CreateIndex(
            name: "IX_Users_Email",
            table: "Users",
            column: "Email",
            unique: true);
    }

    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropTable(name: "Users");
    }
}

// Data migration
public partial class SeedDefaultRoles : Migration
{
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.InsertData(
            table: "Roles",
            columns: new[] { "Id", "Name" },
            values: new object[,]
            {
                { Guid.NewGuid(), "Admin" },
                { Guid.NewGuid(), "User" },
                { Guid.NewGuid(), "Guest" }
            });
    }

    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.Sql("DELETE FROM Roles WHERE Name IN ('Admin', 'User', 'Guest')");
    }
}
```

---

## Query Optimization

### Efficient Querying

```csharp
// Use AsNoTracking for read-only queries
public async Task<IReadOnlyList<UserDto>> GetAllUsersAsync(CancellationToken ct)
    => await context.Users
        .AsNoTracking()
        .Select(u => new UserDto(u.Id, u.Name, u.Email))
        .ToListAsync(ct);

// Use projection to load only needed data
public async Task<UserDto?> GetUserByIdAsync(Guid id, CancellationToken ct)
    => await context.Users
        .AsNoTracking()
        .Where(u => u.Id == id)
        .Select(u => new UserDto(u.Id, u.Name, u.Email))
        .FirstOrDefaultAsync(ct);

// Eager loading with Include
public async Task<User?> GetUserWithPostsAsync(Guid id, CancellationToken ct)
    => await context.Users
        .Include(u => u.Posts)
        .ThenInclude(p => p.Tags)
        .FirstOrDefaultAsync(u => u.Id == id, ct);

// Filtered include
public async Task<User?> GetUserWithRecentPostsAsync(Guid id, CancellationToken ct)
    => await context.Users
        .Include(u => u.Posts.Where(p => p.CreatedAt > DateTime.UtcNow.AddDays(-30)))
        .FirstOrDefaultAsync(u => u.Id == id, ct);
```

### Pagination

```csharp
public async Task<PagedResult<UserDto>> GetPagedUsersAsync(
    int page, 
    int pageSize, 
    string? searchTerm,
    CancellationToken ct)
{
    var query = context.Users.AsNoTracking();
    
    if (!string.IsNullOrEmpty(searchTerm))
    {
        query = query.Where(u => 
            u.Name.Contains(searchTerm) || 
            u.Email.Contains(searchTerm));
    }
    
    var totalCount = await query.CountAsync(ct);
    
    var items = await query
        .OrderBy(u => u.Name)
        .Skip((page - 1) * pageSize)
        .Take(pageSize)
        .Select(u => new UserDto(u.Id, u.Name, u.Email))
        .ToListAsync(ct);
    
    return new PagedResult<UserDto>(items, totalCount, page, pageSize);
}
```

### Compiled Queries

```csharp
public class UserRepository(AppDbContext context) : IUserRepository
{
    // Compiled query for hot paths
    private static readonly Func<AppDbContext, Guid, Task<User?>> GetUserByIdCompiled =
        EF.CompileAsyncQuery((AppDbContext ctx, Guid id) =>
            ctx.Users.FirstOrDefault(u => u.Id == id));

    private static readonly Func<AppDbContext, string, Task<User?>> GetUserByEmailCompiled =
        EF.CompileAsyncQuery((AppDbContext ctx, string email) =>
            ctx.Users.FirstOrDefault(u => u.Email == email));

    public Task<User?> GetByIdAsync(Guid id)
        => GetUserByIdCompiled(context, id);

    public Task<User?> GetByEmailAsync(string email)
        => GetUserByEmailCompiled(context, email);
}
```

### Bulk Operations

```csharp
// EF Core 7+ bulk operations
public async Task DeleteInactiveUsersAsync(CancellationToken ct)
{
    await context.Users
        .Where(u => u.LastLoginAt < DateTime.UtcNow.AddYears(-1))
        .ExecuteDeleteAsync(ct);
}

public async Task UpdateUserRolesAsync(string oldRole, string newRole, CancellationToken ct)
{
    await context.Users
        .Where(u => u.Role == oldRole)
        .ExecuteUpdateAsync(s => s
            .SetProperty(u => u.Role, newRole)
            .SetProperty(u => u.UpdatedAt, DateTime.UtcNow), ct);
}
```

### Raw SQL Queries

```csharp
// FromSqlInterpolated for parameterized queries
public async Task<List<User>> SearchUsersAsync(string searchTerm, CancellationToken ct)
    => await context.Users
        .FromSqlInterpolated($@"
            SELECT * FROM Users 
            WHERE Name LIKE {'%' + searchTerm + '%'} 
            OR Email LIKE {'%' + searchTerm + '%'}")
        .ToListAsync(ct);

// For non-entity results
public async Task<List<UserStatistics>> GetUserStatisticsAsync(CancellationToken ct)
{
    var connection = context.Database.GetDbConnection();
    await connection.OpenAsync(ct);
    
    await using var command = connection.CreateCommand();
    command.CommandText = @"
        SELECT u.Id, u.Name, COUNT(p.Id) as PostCount
        FROM Users u
        LEFT JOIN Posts p ON u.Id = p.AuthorId
        GROUP BY u.Id, u.Name";
    
    var results = new List<UserStatistics>();
    await using var reader = await command.ExecuteReaderAsync(ct);
    while (await reader.ReadAsync(ct))
    {
        results.Add(new UserStatistics(
            reader.GetGuid(0),
            reader.GetString(1),
            reader.GetInt32(2)));
    }
    
    return results;
}
```

---

## Performance Tips

### Query Best Practices

1. Always use `AsNoTracking()` for read-only queries
2. Use projections (`Select`) instead of loading full entities
3. Use `Include` sparingly - prefer explicit joins for complex queries
4. Use compiled queries for frequently executed queries
5. Use pagination for large result sets
6. Use `AsSplitQuery()` for queries with multiple collections

### Index Recommendations

```csharp
// Create indexes for frequently queried columns
builder.HasIndex(u => u.Email).IsUnique();
builder.HasIndex(u => u.CreatedAt);

// Composite indexes for multi-column queries
builder.HasIndex(u => new { u.Status, u.CreatedAt });

// Filtered indexes for soft delete
builder.HasIndex(u => u.Email)
    .HasFilter("[IsDeleted] = 0");
```

### Connection Pooling

```csharp
builder.Services.AddDbContext<AppDbContext>(options =>
{
    options.UseSqlServer(connectionString, sql =>
    {
        sql.MinBatchSize(1);
        sql.MaxBatchSize(100);
    });
});

// Or use DbContextPooling for high-throughput
builder.Services.AddDbContextPool<AppDbContext>(options =>
    options.UseSqlServer(connectionString), poolSize: 128);
```

---

Version: 2.0.0
Last Updated: 2026-01-06
