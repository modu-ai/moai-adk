# PHP Advanced Patterns & Concurrency

## Fibers for Lightweight Concurrency

PHP 8.1+ introduces Fibers for cooperative multitasking without threads.

### Basic Fiber Usage

```php
$fiber = new Fiber(function(): void {
    echo "Fiber started\n";
    Fiber::suspend('intermediate');
    echo "Fiber resumed\n";
});

$value = $fiber->start();
echo "Received: $value\n";
$fiber->resume();
```

### Async HTTP with Fibers

```php
function async_fetch(string $url): Fiber {
    return new Fiber(function() use ($url) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        // Simulate async with suspension
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

### Fiber Pool for Concurrent Operations

```php
class FiberPool {
    private array $fibers = [];
    private array $results = [];

    public function add(string $id, Fiber $fiber): void {
        $this->fibers[$id] = $fiber;
    }

    public function run(): array {
        foreach ($this->fibers as $id => $fiber) {
            $this->results[$id] = $fiber->start();
        }

        return $this->results;
    }

    public function getResult(string $id): mixed {
        return $this->results[$id] ?? null;
    }
}

$pool = new FiberPool();
$pool->add('users', async_fetch('https://api.example.com/users'));
$pool->add('posts', async_fetch('https://api.example.com/posts'));
$results = $pool->run();
```

---

## Attributes for Metadata & Reflection

### Custom Attribute Definition

```php
#[Attribute]
class Route {
    public function __construct(
        public string $path,
        public string $method = 'GET'
    ) {}
}

#[Attribute]
class ValidateWith {
    public function __construct(
        public string $validator
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
    #[ValidateWith('UserValidator')]
    public function store(Request $request): User {
        return User::create($request->all());
    }
}
```

### Processing Attributes with Reflection

```php
class RouteRegistry {
    public static function register(string $controllerClass): array {
        $routes = [];
        $reflection = new ReflectionClass($controllerClass);

        foreach ($reflection->getMethods() as $method) {
            $attributes = $method->getAttributes(Route::class);

            foreach ($attributes as $attribute) {
                $route = $attribute->newInstance();
                $routes[] = [
                    'path' => $route->path,
                    'method' => $route->method,
                    'handler' => [$controllerClass, $method->getName()]
                ];
            }
        }

        return $routes;
    }
}

$routes = RouteRegistry::register(UserController::class);
```

### Attribute-Based Validation

```php
#[Attribute]
class Required {}

#[Attribute]
class Email {}

#[Attribute]
class MinLength {
    public function __construct(public int $length) {}
}

class UserValidator {
    public static function validate(object $object): array {
        $errors = [];
        $reflection = new ReflectionClass($object);

        foreach ($reflection->getProperties() as $property) {
            $attributes = $property->getAttributes();

            foreach ($attributes as $attribute) {
                $attr = $attribute->newInstance();

                if ($attr instanceof Required && empty($property->getValue($object))) {
                    $errors[$property->getName()] = 'Required field';
                }

                if ($attr instanceof Email && !filter_var($property->getValue($object), FILTER_VALIDATE_EMAIL)) {
                    $errors[$property->getName()] = 'Invalid email';
                }

                if ($attr instanceof MinLength && strlen((string)$property->getValue($object)) < $attr->length) {
                    $errors[$property->getName()] = "Minimum {$attr->length} characters";
                }
            }
        }

        return $errors;
    }
}
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

    public function auto(string $class): void {
        $this->singleton($class, function($container) use ($class) {
            $reflection = new ReflectionClass($class);
            $constructor = $reflection->getConstructor();

            if ($constructor === null) {
                return new $class();
            }

            $parameters = [];
            foreach ($constructor->getParameters() as $param) {
                $type = $param->getType();

                if ($type && !$type->isBuiltin() && $container->has($type->getName())) {
                    $parameters[] = $container->get($type->getName());
                }
            }

            return new $class(...$parameters);
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
```

### Factory Pattern with Generic Types

```php
/**
 * @template T
 */
class Factory {
    /** @var class-string<T> */
    private string $class;

    /** @var callable(mixed): T */
    private $builder;

    /**
     * @param class-string<T> $class
     * @param callable(mixed): T $builder
     */
    public function __construct(string $class, callable $builder) {
        $this->class = $class;
        $this->builder = $builder;
    }

    /**
     * @return T
     */
    public function create(mixed $data): object {
        return ($this->builder)($data);
    }
}

// Usage
$userFactory = new Factory(User::class, function($data) {
    return new User(
        name: $data['name'],
        email: $data['email']
    );
});

$user = $userFactory->create(['name' => 'John', 'email' => 'john@example.com']);
```

### Observer Pattern with Type Safety

```php
interface Observer {
    public function update(Event $event): void;
}

class Event {
    public function __construct(
        public string $type,
        public mixed $data
    ) {}
}

class EventDispatcher {
    /** @var array<string, array<Observer>> */
    private array $observers = [];

    public function subscribe(string $eventType, Observer $observer): void {
        if (!isset($this->observers[$eventType])) {
            $this->observers[$eventType] = [];
        }
        $this->observers[$eventType][] = $observer;
    }

    public function dispatch(Event $event): void {
        if (!isset($this->observers[$event->type])) {
            return;
        }

        foreach ($this->observers[$event->type] as $observer) {
            $observer->update($event);
        }
    }
}

class LoggingObserver implements Observer {
    public function update(Event $event): void {
        logger()->info("Event: {$event->type}", ['data' => $event->data]);
    }
}

class EmailObserver implements Observer {
    public function update(Event $event): void {
        if ($event->type === 'user.created') {
            mail($event->data['email'], 'Welcome!', 'Your account has been created.');
        }
    }
}
```

---

## Generics with PHPStan/Psalm

### Generic Collections

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

    /**
     * @param callable(T): bool $predicate
     * @return Collection<T>
     */
    public function filter(callable $predicate): Collection {
        $result = new Collection();
        foreach ($this->items as $item) {
            if ($predicate($item)) {
                $result->add($item);
            }
        }
        return $result;
    }
}

/** @var Collection<User> */
$users = new Collection();
$users->add(new User('Alice'));
$users->add(new User('Bob'));

// Type-safe mapping
$names = $users->map(fn(User $u) => $u->name);  // Collection<string>

// Type-safe filtering
$active = $users->filter(fn(User $u) => $u->isActive());  // Collection<User>
```

### Repository with Generics

```php
/**
 * @template T of Entity
 */
interface RepositoryInterface {
    /**
     * @param int $id
     * @return T|null
     */
    public function find(int $id): ?object;

    /**
     * @return array<T>
     */
    public function all(): array;

    /**
     * @param T $entity
     * @return T
     */
    public function save(object $entity): object;
}

/**
 * @template T of Entity
 * @implements RepositoryInterface<T>
 */
abstract class AbstractRepository implements RepositoryInterface {
    protected string $modelClass;

    /**
     * @return T|null
     */
    public function find(int $id): ?object {
        return $this->modelClass::find($id);
    }

    /**
     * @return array<T>
     */
    public function all(): array {
        return $this->modelClass::all()->toArray();
    }

    /**
     * @param T $entity
     * @return T
     */
    public function save(object $entity): object {
        $entity->save();
        return $entity;
    }
}

class UserRepository extends AbstractRepository {
    protected string $modelClass = User::class;
}

class PostRepository extends AbstractRepository {
    protected string $modelClass = Post::class;
}
```

---

## Metaprogramming with Magic Methods

### Dynamic Property Access

```php
class Model {
    protected array $attributes = [];
    protected array $casts = [];

    public function __get(string $name): mixed {
        if (array_key_exists($name, $this->attributes)) {
            $value = $this->attributes[$name];

            if (isset($this->casts[$name])) {
                return $this->cast($value, $this->casts[$name]);
            }

            return $value;
        }

        throw new PropertyNotFoundException("Property $name not found");
    }

    public function __set(string $name, mixed $value): void {
        $this->attributes[$name] = $value;
    }

    public function __isset(string $name): bool {
        return array_key_exists($name, $this->attributes);
    }

    private function cast(mixed $value, string $type): mixed {
        return match($type) {
            'int' => (int)$value,
            'float' => (float)$value,
            'bool' => (bool)$value,
            'string' => (string)$value,
            'array' => json_decode($value, true),
            'datetime' => new DateTime($value),
            default => $value
        };
    }
}

class User extends Model {
    protected array $casts = [
        'age' => 'int',
        'is_admin' => 'bool',
        'metadata' => 'array',
        'created_at' => 'datetime'
    ];
}
```

### Dynamic Method Calling

```php
class QueryBuilder {
    private array $where = [];
    private array $select = [];
    private int $limit = 0;

    public function __call(string $method, array $args): self {
        // Support where<Field> syntax
        if (str_starts_with($method, 'where')) {
            $field = substr($method, 5);
            $field = strtolower(preg_replace('/(?<!^)[A-Z]/', '_$0', $field));

            $this->where[$field] = $args[0] ?? null;
            return $this;
        }

        throw new BadMethodCallException("Method $method does not exist");
    }

    public static function __callStatic(string $method, array $args): self {
        // Support User::whereId(1) syntax
        $builder = new static();
        return $builder->__call('where' . ucfirst($method), $args);
    }
}

// Usage
$query = QueryBuilder::whereId(1)
    ->whereName('John')
    ->whereActive(true);
```

---

## Union Types & Match Expressions

### Union Type Patterns

```php
function process(int|float|string $value): int|float {
    return is_string($value) ? strlen($value) : $value * 2;
}

function getStatus(int|string|null $value): string {
    return match($value) {
        1, 'active' => 'Active',
        0, 'inactive' => 'Inactive',
        null => 'Unknown',
        default => 'Invalid'
    };
}

// Union types with classes
interface Cacheable {}

class RedisCache implements Cacheable {
    public function get(string $key): mixed { /* ... */ }
}

class FileCache implements Cacheable {
    public function get(string $key): mixed { /* ... */ }
}

function getCachedValue(RedisCache|FileCache $cache, string $key): mixed {
    return $cache->get($key);
}
```

### Match with Complex Logic

```php
enum Status: string {
    case Pending = 'pending';
    case Approved = 'approved';
    case Rejected = 'rejected';
}

function handleStatus(Status $status, User $user): string {
    return match($status) {
        Status::Pending => "Waiting for approval of {$user->name}",
        Status::Approved => "Congratulations {$user->name}! You've been approved.",
        Status::Rejected => "Sorry {$user->name}. Your request was rejected.",
    };
}

function calculateFee(Status $status, float $amount): float {
    return match($status) {
        Status::Pending => $amount,
        Status::Approved => $amount * 0.95,  // 5% discount
        Status::Rejected => 0,
    };
}
```

---

## Nullsafe Operator & Constructor Promotion

### Nullsafe Operator Chains

```php
// Safe navigation without multiple null checks
$country = $user?->address?->country?->name ?? 'Unknown';

// Method calls through null chain
$email = $order?->customer?->getEmail() ?? 'no-reply@example.com';

// Safe array access
$city = $address?->location?->coordinates?->city ?? null;

// Comparison with nullsafe
if ($user?->profile?->isVerified() === true) {
    // User exists, has profile, and is verified
}
```

### Constructor Property Promotion

```php
// Before: verbose constructor
class UserOld {
    private string $name;
    private string $email;
    private bool $isAdmin;

    public function __construct(string $name, string $email, bool $isAdmin) {
        $this->name = $name;
        $this->email = $email;
        $this->isAdmin = $isAdmin;
    }
}

// After: concise with promotion
class User {
    public function __construct(
        private string $name,
        private string $email,
        private bool $isAdmin
    ) {}

    public function getName(): string {
        return $this->name;
    }
}

// With defaults and visibility
class Admin {
    public function __construct(
        public string $name,
        public string $email,
        private string $password,
        protected array $permissions = []
    ) {}
}
```

---

## Arrow Functions & Modern Syntax

### Arrow Functions in Collections

```php
$users = [
    new User('Alice', 'alice@example.com'),
    new User('Bob', 'bob@example.com'),
];

// Arrow function for concise mapping
$names = array_map(fn($u) => $u->name, $users);
$emails = array_map(fn($u) => $u->email, $users);

// Combined with array operations
$filtered = array_filter(
    $users,
    fn($u) => str_contains($u->email, '@example.com')
);

// Chaining arrow functions
$result = array_map(
    fn($user) => [
        'name' => $user->name,
        'domain' => explode('@', $user->email)[1]
    ],
    $users
);
```

### Arrow Functions with Closure

```php
class Calculator {
    private array $operations = [];

    public function register(string $name, \Closure $operation): void {
        $this->operations[$name] = $operation;
    }

    public function execute(string $name, int $a, int $b): int {
        return ($this->operations[$name])($a, $b);
    }
}

$calc = new Calculator();
$calc->register('add', fn($a, $b) => $a + $b);
$calc->register('multiply', fn($a, $b) => $a * $b);
$calc->register('power', fn($a, $b) => $a ** $b);

echo $calc->execute('add', 5, 3);       // 8
echo $calc->execute('multiply', 5, 3);  // 15
echo $calc->execute('power', 5, 3);     // 125
```

---

## Best Practices Summary

**DO**:
- Use Fibers for concurrent operations
- Apply Attributes for metadata and validation
- Leverage type safety with generics (PHPStan/Psalm)
- Use match expressions instead of switch
- Apply nullsafe operator for safe navigation
- Use constructor property promotion
- Implement SOLID principles consistently

**DON'T**:
- Use Fibers for CPU-bound work (use threads/processes)
- Ignore attribute constraints
- Skip type checking with PHPStan
- Use switch when match is clearer
- Chain complex nullable operations without nullsafe
- Create verbose constructors (use promotion)
- Mix patterns inconsistently across codebase

---

**Total Lines**: 458
