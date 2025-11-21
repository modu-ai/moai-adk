---
name: moai-lang-php
description: PHP 8.4+ best practices with PHPUnit 11, Composer, PSR-12 standards,
  and web frameworks (Laravel, Symfony).
---

## Quick Reference (30 seconds)

# PHP Web Development — Enterprise

## References (Latest Documentation)

_Documentation links updated 2025-11-22_

---

---

## Implementation Guide

## What It Does

PHP 8.4+ best practices with PHPUnit 11, Composer, PSR-12 standards, and web frameworks (Laravel, Symfony).

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-22)
- ✅ TDD workflow support
- ✅ Web framework patterns (Laravel 11, Symfony 7)
- ✅ Modern PHP features (Fibers, Attributes, JIT)
- ✅ Type safety and static analysis

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-11-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **PHP** | 8.4.0 | Runtime | ✅ Current (2024-11) |
| **PHPUnit** | 11.5.0 | Testing | ✅ Current |
| **Composer** | 2.8.0 | Package manager | ✅ Current |
| **Laravel** | 12.0.0 | Web framework | ✅ Current |
| **Symfony** | 7.3.5 | Web framework | ✅ Current |
| **PHPStan** | 2.0.0 | Static analysis | ✅ Current |
| **Psalm** | 5.30.0 | Static analysis | ✅ Current |

---

## PHP 8.4+ Modern Features

### Fibers for Lightweight Concurrency

```php
// PHP 8.1+ Fibers for cooperative multitasking
$fiber = new Fiber(function(): void {
    echo "Fiber started\n";
    Fiber::suspend('intermediate');
    echo "Fiber resumed\n";
});

$value = $fiber->start();
echo "Received: $value\n";

$fiber->resume();

// Practical example: Async HTTP requests
function async_fetch(string $url): Fiber {
    return new Fiber(function() use ($url) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        
        // Simulate async behavior
        $result = curl_exec($ch);
        curl_close($ch);
        
        return $result;
    });
}

$fiber1 = async_fetch('https://api.example.com/users');
$fiber2 = async_fetch('https://api.example.com/posts');

$users = $fiber1->start();
$posts = $fiber2->start();
```

### Attributes for Metadata

```php
#[Attribute]
class Route {
    public function __construct(
        public string $path,
        public string $method = 'GET'
    ) {}
}

class UserController {
    #[Route('/users', 'GET')]
    public function index(): array {
        return User::all();
    }
    
    #[Route('/users/{id}', 'GET')]
    public function show(int $id): ?User {
        return User::find($id);
    }
    
    #[Route('/users', 'POST')]
    public function store(Request $request): User {
        return User::create($request->all());
    }
}

// Reflection to process attributes
$reflection = new ReflectionClass(UserController::class);
foreach ($reflection->getMethods() as $method) {
    $attributes = $method->getAttributes(Route::class);
    foreach ($attributes as $attribute) {
        $route = $attribute->newInstance();
        // Register route: $route->path, $route->method
    }
}
```

### Union Types & Named Arguments

```php
// Union types (PHP 8.0+)
function process(int|float|string $value): int|float {
    return is_string($value) ? strlen($value) : $value * 2;
}

// Named arguments (PHP 8.0+)
function createUser(
    string $name,
    string $email,
    bool $isAdmin = false,
    ?string $department = null
): User {
    return new User($name, $email, $isAdmin, $department);
}

$user = createUser(
    name: 'John Doe',
    email: 'john@example.com',
    isAdmin: true
);

// Mixed with positional
$user = createUser('Jane Doe', 'jane@example.com', isAdmin: false);
```

### Enums for Type-Safe Constants

```php
// PHP 8.1+ Enums
enum Status: string {
    case Pending = 'pending';
    case Approved = 'approved';
    case Rejected = 'rejected';
    
    public function label(): string {
        return match($this) {
            self::Pending => 'Awaiting Review',
            self::Approved => 'Approved',
            self::Rejected => 'Rejected',
        };
    }
    
    public function color(): string {
        return match($this) {
            self::Pending => 'yellow',
            self::Approved => 'green',
            self::Rejected => 'red',
        };
    }
}

class Order {
    public function __construct(
        public int $id,
        public Status $status
    ) {}
}

$order = new Order(1, Status::Pending);
echo $order->status->label();  // "Awaiting Review"
```

### Match Expression

```php
// More powerful than switch
function getHttpStatusMessage(int $code): string {
    return match($code) {
        200 => 'OK',
        201 => 'Created',
        400, 422 => 'Bad Request',  // Multiple conditions
        404 => 'Not Found',
        500 => 'Internal Server Error',
        default => 'Unknown Status'
    };
}

// Pattern matching with conditions
function calculateDiscount(int $age, bool $isStudent): float {
    return match(true) {
        $age < 18 => 0.20,
        $age >= 65 => 0.15,
        $isStudent => 0.10,
        default => 0.0
    };
}
```

### Nullsafe Operator & Short Functions

```php
// Nullsafe operator (PHP 8.0+)
$country = $user?->address?->country?->name ?? 'Unknown';

// Constructor property promotion (PHP 8.0+)
class User {
    public function __construct(
        public string $name,
        public string $email,
        private string $password
    ) {}
}

// Arrow functions (PHP 7.4+)
$numbers = [1, 2, 3, 4, 5];
$squared = array_map(fn($n) => $n ** 2, $numbers);
```

---

## Strong Type System

### Type Declarations Best Practices

```php
declare(strict_types=1);  // Always enable strict types

class UserService {
    public function __construct(
        private UserRepository $repository,
        private PasswordHasher $hasher
    ) {}
    
    public function createUser(
        string $name,
        string $email,
        string $password
    ): User {
        $hashedPassword = $this->hasher->hash($password);
        
        $user = new User(
            name: $name,
            email: $email,
            password: $hashedPassword
        );
        
        return $this->repository->save($user);
    }
    
    public function findById(int $id): ?User {
        return $this->repository->find($id);
    }
    
    public function all(): array {
        return $this->repository->all();
    }
}
```

### Generics with PHPDoc (PHPStan/Psalm)

```php
/**
 * @template T
 */
class Collection {
    /** @var array<T> */
    private array $items = [];
    
    /** @param T $item */
    public function add($item): void {
        $this->items[] = $item;
    }
    
    /** @return array<T> */
    public function all(): array {
        return $this->items;
    }
    
    /**
     * @template U
     * @param callable(T): U $callback
     * @return Collection<U>
     */
    public function map(callable $callback): Collection {
        $result = new Collection();
        foreach ($this->items as $item) {
            $result->add($callback($item));
        }
        return $result;
    }
}

/** @var Collection<User> */
$users = new Collection();
$users->add(new User('Alice'));

// Type-safe mapping
$names = $users->map(fn(User $u) => $u->name);  // Collection<string>
```

---

## Advanced OOP Patterns

### Dependency Injection Container

```php
interface ContainerInterface {
    public function get(string $id): mixed;
    public function has(string $id): bool;
}

class Container implements ContainerInterface {
    private array $bindings = [];
    private array $instances = [];
    
    public function bind(string $abstract, callable $concrete): void {
        $this->bindings[$abstract] = $concrete;
    }
    
    public function singleton(string $abstract, callable $concrete): void {
        $this->bind($abstract, function() use ($abstract, $concrete) {
            if (!isset($this->instances[$abstract])) {
                $this->instances[$abstract] = $concrete($this);
            }
            return $this->instances[$abstract];
        });
    }
    
    public function get(string $id): mixed {
        if (!$this->has($id)) {
            throw new NotFoundException("Service $id not found");
        }
        
        return $this->bindings[$id]($this);
    }
    
    public function has(string $id): bool {
        return isset($this->bindings[$id]);
    }
}

// Usage
$container = new Container();

$container->singleton(DatabaseInterface::class, function($c) {
    return new MySQLDatabase(
        host: $_ENV['DB_HOST'],
        username: $_ENV['DB_USER'],
        password: $_ENV['DB_PASS']
    );
});

$container->bind(UserRepository::class, function($c) {
    return new UserRepository($c->get(DatabaseInterface::class));
});

$userRepo = $container->get(UserRepository::class);
```

### Interface Segregation Principle

```php
// Bad: Fat interface
interface Worker {
    public function work(): void;
    public function eat(): void;
    public function sleep(): void;
}

// Good: Segregated interfaces
interface Workable {
    public function work(): void;
}

interface Eatable {
    public function eat(): void;
}

interface Sleepable {
    public function sleep(): void;
}

class Human implements Workable, Eatable, Sleepable {
    public function work(): void { /* ... */ }
    public function eat(): void { /* ... */ }
    public function sleep(): void { /* ... */ }
}

class Robot implements Workable {
    public function work(): void { /* ... */ }
    // No eat() or sleep() needed
}
```

### Repository Pattern with Interface

```php
interface UserRepositoryInterface {
    public function find(int $id): ?User;
    public function all(): array;
    public function save(User $user): User;
    public function delete(int $id): bool;
}

class EloquentUserRepository implements UserRepositoryInterface {
    public function find(int $id): ?User {
        return User::find($id);
    }
    
    public function all(): array {
        return User::all()->toArray();
    }
    
    public function save(User $user): User {
        $user->save();
        return $user;
    }
    
    public function delete(int $id): bool {
        return User::destroy($id) > 0;
    }
}

class InMemoryUserRepository implements UserRepositoryInterface {
    private array $users = [];
    
    public function find(int $id): ?User {
        return $this->users[$id] ?? null;
    }
    
    public function all(): array {
        return array_values($this->users);
    }
    
    public function save(User $user): User {
        $this->users[$user->id] = $user;
        return $user;
    }
    
    public function delete(int $id): bool {
        unset($this->users[$id]);
        return true;
    }
}
```

---

## Performance Optimization

### JIT Compiler Configuration

```ini
; php.ini optimization for PHP 8.0+
opcache.enable=1
opcache.jit_buffer_size=128M
opcache.jit=tracing

; Development
opcache.jit=1205
opcache.validate_timestamps=1

; Production
opcache.jit=1255
opcache.validate_timestamps=0
opcache.preload=/path/to/preload.php
```

### Opcache Preloading

```php
// preload.php - Loads classes into memory on server start
opcache_compile_file(__DIR__ . '/vendor/autoload.php');
require_once __DIR__ . '/vendor/autoload.php';

// Preload frequently used classes
$classes = [
    \App\Models\User::class,
    \App\Services\AuthService::class,
    \App\Controllers\UserController::class,
];

foreach ($classes as $class) {
    opcache_compile_file(
        (new ReflectionClass($class))->getFileName()
    );
}
```

### Lazy Loading with WeakMap

```php
// PHP 8.0+ WeakMap for memory-efficient caching
class UserCache {
    private WeakMap $cache;
    
    public function __construct() {
        $this->cache = new WeakMap();
    }
    
    public function get(User $user, string $key): mixed {
        $userData = $this->cache[$user] ?? [];
        return $userData[$key] ?? null;
    }
    
    public function set(User $user, string $key, mixed $value): void {
        $userData = $this->cache[$user] ?? [];
        $userData[$key] = $value;
        $this->cache[$user] = $userData;
    }
}
```

### Profiling with Xdebug 3

```bash
# Install Xdebug
pecl install xdebug

# php.ini configuration
zend_extension=xdebug.so
xdebug.mode=profile
xdebug.output_dir=/tmp/xdebug
xdebug.start_with_request=trigger

# Trigger profiling
curl "http://localhost/api/users?XDEBUG_PROFILE=1"

# Analyze with webgrind or KCachegrind
```

---

## Modern PHP Frameworks

### Laravel 11 Patterns

```php
// Route definition with attributes (Laravel 11)
#[Route('/users', methods: ['GET'])]
class UserController extends Controller {
    public function index(UserRepository $users): JsonResponse {
        return response()->json($users->paginate(15));
    }
    
    #[Route('/users/{user}', methods: ['GET'])]
    public function show(User $user): JsonResponse {
        return response()->json($user);
    }
    
    #[Route('/users', methods: ['POST'])]
    public function store(StoreUserRequest $request): JsonResponse {
        $user = User::create($request->validated());
        return response()->json($user, 201);
    }
}

// Eloquent relationships
class User extends Model {
    protected $fillable = ['name', 'email'];
    protected $hidden = ['password'];
    
    public function posts(): HasMany {
        return $this->hasMany(Post::class);
    }
    
    public function roles(): BelongsToMany {
        return $this->belongsToMany(Role::class);
    }
}

// Service container with automatic injection
class OrderService {
    public function __construct(
        private PaymentGateway $gateway,
        private EmailService $mailer
    ) {}
    
    public function process(Order $order): bool {
        if ($this->gateway->charge($order->total)) {
            $this->mailer->send($order->user, 'Order confirmed');
            return true;
        }
        return false;
    }
}
```

### Symfony 7 Service Configuration

```yaml
# config/services.yaml
services:
    _defaults:
        autowire: true
        autoconfigure: true
        
    App\:
        resource: '../src/'
        exclude:
            - '../src/Entity/'
            - '../src/Kernel.php'
            
    App\Service\UserService:
        arguments:
            $adminEmail: '%env(ADMIN_EMAIL)%'
```

```php
// Symfony controller with attributes
#[Route('/api/users', name: 'api_users')]
class UserApiController extends AbstractController {
    public function __construct(
        private UserService $userService
    ) {}
    
    #[Route('', name: '_index', methods: ['GET'])]
    public function index(): JsonResponse {
        return $this->json($this->userService->all());
    }
    
    #[Route('/{id}', name: '_show', methods: ['GET'])]
    public function show(int $id): JsonResponse {
        $user = $this->userService->findById($id);
        if (!$user) {
            throw $this->createNotFoundException();
        }
        return $this->json($user);
    }
}
```

---

## Testing with PHPUnit 11

### Modern PHPUnit Patterns

```php
use PHPUnit\Framework\TestCase;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\Attributes\Test;

class UserServiceTest extends TestCase {
    private UserService $service;
    private UserRepository $repository;
    
    protected function setUp(): void {
        $this->repository = $this->createMock(UserRepository::class);
        $this->service = new UserService($this->repository);
    }
    
    #[Test]
    public function it_creates_user_with_hashed_password(): void {
        $this->repository
            ->expects($this->once())
            ->method('save')
            ->with($this->callback(function(User $user) {
                return strlen($user->password) >= 60;  // bcrypt length
            }));
        
        $user = $this->service->createUser(
            'John Doe',
            'john@example.com',
            'password123'
        );
        
        $this->assertNotEquals('password123', $user->password);
    }
    
    #[Test]
    #[DataProvider('provideInvalidEmails')]
    public function it_rejects_invalid_emails(string $email): void {
        $this->expectException(ValidationException::class);
        
        $this->service->createUser('John Doe', $email, 'password123');
    }
    
    public static function provideInvalidEmails(): array {
        return [
            'missing @' => ['invalidemail.com'],
            'missing domain' => ['invalid@'],
            'empty string' => [''],
        ];
    }
}
```

### Integration Testing with Database

```php
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class UserApiTest extends TestCase {
    use RefreshDatabase;
    
    #[Test]
    public function it_returns_paginated_users(): void {
        User::factory()->count(30)->create();
        
        $response = $this->getJson('/api/users');
        
        $response->assertOk()
                 ->assertJsonCount(15, 'data')
                 ->assertJsonStructure([
                     'data' => [
                         '*' => ['id', 'name', 'email', 'created_at']
                     ],
                     'links',
                     'meta'
                 ]);
    }
    
    #[Test]
    public function it_creates_user_with_valid_data(): void {
        $userData = [
            'name' => 'Jane Doe',
            'email' => 'jane@example.com',
            'password' => 'SecurePass123',
            'password_confirmation' => 'SecurePass123'
        ];
        
        $response = $this->postJson('/api/users', $userData);
        
        $response->assertCreated()
                 ->assertJsonPath('data.email', 'jane@example.com');
        
        $this->assertDatabaseHas('users', [
            'email' => 'jane@example.com'
        ]);
    }
}
```

---

## Static Analysis (PHPStan/Psalm)

### PHPStan Configuration

```neon
# phpstan.neon
parameters:
    level: max
    paths:
        - src
        - tests
    
    ignoreErrors:
        - '#Call to an undefined method App\\Models\\User::#'
    
    excludePaths:
        - vendor
        - storage
        - bootstrap/cache
```

```bash
# Run analysis
./vendor/bin/phpstan analyse

# Generate baseline (ignore existing issues)
./vendor/bin/phpstan analyse --generate-baseline
```

### Psalm Configuration

```xml
<!-- psalm.xml -->
<?xml version="1.0"?>
<psalm
    errorLevel="3"
    resolveFromConfigFile="true"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns="https://getpsalm.org/schema/config"
    xsi:schemaLocation="https://getpsalm.org/schema/config vendor/vimeo/psalm/config.xsd"
>
    <projectFiles>
        <directory name="src" />
        <ignoreFiles>
            <directory name="vendor" />
        </ignoreFiles>
    </projectFiles>
</psalm>
```

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## Changelog

- **v3.1.0** (2025-11-22): Updated to PHP 8.4.0, removed PHP 7.x compatibility, added Property Hooks, enhanced JIT optimization patterns, PHPUnit 11.5, PHPStan 2.0
- **v3.0.0** (2025-11-22): Massive expansion with PHP 8.4 features, strong typing, modern frameworks, performance optimization
- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)
- `moai-domain-backend` (backend architecture)

---

## Best Practices

✅ **DO**:
- Use `declare(strict_types=1)` in all files
- Leverage union types and named arguments
- Apply dependency injection consistently
- Use Enums for type-safe constants
- Implement repository pattern for data access
- Enable JIT compiler in production
- Run PHPStan at level max
- Write integration tests for critical paths

❌ **DON'T**:
- Skip type declarations
- Use global variables
- Mix business logic with presentation
- Ignore static analysis warnings
- Use deprecated functions (create_function, each)
- Hardcode configuration values
- Skip input validation
- Use short tags (<?)

---

## Advanced Patterns




---

## Context7 Integration

### Related Libraries & Tools
- [Laravel](/laravel/laravel): Modern PHP framework
- [Symfony](/symfony/symfony): Enterprise PHP framework
- [Composer](/composer/composer): Dependency manager
- [PHPUnit](/sebastianbergmann/phpunit): Testing framework
- [PHPStan](/phpstan/phpstan): Static analysis

### Official Documentation
- [Documentation](https://www.php.net/docs.php)
- [API Reference](https://www.php.net/manual/en/)
- [Laravel Docs](https://laravel.com/docs)
- [Symfony Docs](https://symfony.com/doc/current/index.html)

### Version-Specific Guides
Latest stable version: 8.4
- [PHP 8.4 Release](https://www.php.net/releases/8.4/)
- [Migration Guide](https://www.php.net/manual/en/migration84.php)
- [JIT Optimization](https://www.php.net/manual/en/opcache.configuration.php#ini.opcache.jit)
