# PHP Performance Optimization

## JIT Compiler Configuration

PHP 8.0+ JIT (Just-In-Time) compiler can provide 2-3x performance improvements for CPU-bound code.

### php.ini Configuration

```ini
; Production JIT configuration
opcache.enable=1
opcache.enable_cli=1
opcache.jit=1255
opcache.jit_buffer_size=256M
opcache.jit_debug=0

; Development JIT configuration
opcache.enable=1
opcache.jit=1205
opcache.jit_buffer_size=128M
opcache.validate_timestamps=1
opcache.revalidate_freq=0

; General optimization
opcache.memory_consumption=512
opcache.interned_strings_buffer=16
opcache.max_accelerated_files=100000
opcache.revalidate_freq=2
opcache.fast_shutdown=1
opcache.preload=/var/www/html/preload.php
```

### JIT Flags Explanation

```
JIT=0    - Disabled
JIT=4    - Tracing (default, best for most apps)
JIT=1    - Function JIT
JIT=8    - Loop JIT
JIT=12   - Function + Loop
JIT=4095 - All (very aggressive)
JIT=1255 - Recommended (tracing + function + loop + preload)
```

### Measuring JIT Impact

```php
function fibonacci(int $n): int {
    if ($n <= 1) return $n;
    return fibonacci($n - 1) + fibonacci($n - 2);
}

// Benchmark with microtime
$start = microtime(true);
$result = fibonacci(35);
$time = microtime(true) - $start;

echo "Result: $result, Time: {$time}s\n";
// Without JIT: ~15s, With JIT: ~2s (7x faster!)
```

---

## Opcache Preloading

Preloading loads classes into shared memory at server startup.

### preload.php Example

```php
<?php
// /var/www/html/preload.php
opcache_compile_file(__DIR__ . '/vendor/autoload.php');
require_once __DIR__ . '/vendor/autoload.php';

// Preload frequently used classes
$classes = [
    \App\Models\User::class,
    \App\Models\Post::class,
    \App\Services\AuthService::class,
    \App\Services\UserService::class,
    \App\Repositories\UserRepository::class,
    \App\Http\Controllers\UserController::class,
    \Illuminate\Database\Eloquent\Model::class,
    \Illuminate\Http\Request::class,
];

foreach ($classes as $class) {
    if (class_exists($class)) {
        opcache_compile_file(
            (new ReflectionClass($class))->getFileName()
        );
    }
}

// Preload trait files
$traitFiles = glob(__DIR__ . '/app/Traits/*.php');
foreach ($traitFiles as $file) {
    opcache_compile_file($file);
}

// Preload interface files
$interfaceFiles = glob(__DIR__ . '/app/Contracts/*.php');
foreach ($interfaceFiles as $file) {
    opcache_compile_file($file);
}
```

### Configuration

```ini
opcache.preload=/var/www/html/preload.php
opcache.preload_user=www-data
```

### Preload Benefits

- **Faster startup**: Classes loaded in shared memory
- **Reduced memory per process**: Shared copy across all processes
- **Predictable performance**: No first-request delay

---

## Lazy Loading with WeakMap

WeakMap allows memory-efficient caching without preventing garbage collection.

```php
class UserCache {
    private WeakMap $cache;

    public function __construct() {
        $this->cache = new WeakMap();
    }

    public function get(User $user, string $key): mixed {
        if (!isset($this->cache[$user])) {
            $this->cache[$user] = [];
        }

        return $this->cache[$user][$key] ?? null;
    }

    public function set(User $user, string $key, mixed $value): void {
        if (!isset($this->cache[$user])) {
            $this->cache[$user] = [];
        }

        $this->cache[$user][$key] = $value;
    }

    public function has(User $user, string $key): bool {
        return isset($this->cache[$user][$key]);
    }

    public function clear(User $user): void {
        unset($this->cache[$user]);
    }
}

// Usage
$cache = new UserCache();

$user = User::find(1);
$cache->set($user, 'permissions', ['read', 'write', 'delete']);
$cache->set($user, 'preferences', ['theme' => 'dark']);

$perms = $cache->get($user, 'permissions');

// When $user is garbage collected, cache is automatically cleared
unset($user);
```

---

## Database Query Optimization

### Connection Pooling

```php
// Using persistent connections
$pdo = new PDO(
    'mysql:host=localhost;dbname=myapp;charset=utf8mb4',
    'user',
    'password',
    [
        PDO::ATTR_PERSISTENT => true,
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
    ]
);

// With connection parameters
$pdo = new PDO(
    'mysql:host=localhost;port=3306;dbname=myapp',
    'user',
    'password',
    [
        PDO::ATTR_PERSISTENT => true,
        PDO::ATTR_TIMEOUT => 30,
        PDO::MYSQL_ATTR_INIT_COMMAND => 'SET NAMES utf8mb4',
    ]
);
```

### Eager Loading with Relationships

```php
// Bad: N+1 queries
$users = User::all();
foreach ($users as $user) {
    echo $user->posts->count();  // Query per user!
}

// Good: Single query with eager loading
$users = User::with('posts')->get();
foreach ($users as $user) {
    echo $user->posts->count();  // No additional queries
}

// Multiple relationships
$users = User::with(['posts', 'comments', 'profile'])
    ->whereActive(true)
    ->get();

// Nested eager loading
$posts = Post::with('user', 'comments.author')->get();

// Conditional eager loading
$users = User::with([
    'posts' => fn($q) => $q->where('published', true)->limit(10),
    'profile' => fn($q) => $q->select('id', 'user_id', 'bio')
])->get();
```

### Query Optimization Patterns

```php
// Bad: Fetch all columns
$users = User::all();

// Good: Select only needed columns
$users = User::select('id', 'name', 'email')
    ->where('active', true)
    ->get();

// Pagination
$users = User::paginate(50);  // Better than LIMIT/OFFSET for large offsets

// Indexed columns
$user = User::whereEmail('john@example.com')->first();  // Use index on email

// Batch processing
foreach (User::chunk(1000) as $chunk) {
    foreach ($chunk as $user) {
        $user->update(['processed' => true]);
    }
}

// Bulk operations
User::whereStatus('inactive')
    ->delete();  // Single query

Post::insert([  // Bulk insert
    ['title' => 'Post 1', 'user_id' => 1],
    ['title' => 'Post 2', 'user_id' => 2],
]);
```

---

## Profiling with Xdebug 3

### Installation & Configuration

```bash
# Install Xdebug
pecl install xdebug

# php.ini configuration
zend_extension=xdebug.so
xdebug.mode=profile,trace
xdebug.output_dir=/var/xdebug
xdebug.start_with_request=trigger
xdebug.profiler_output_name=cachegrind.out.%H.%R
```

### Triggering Profiling

```bash
# From command line
XDEBUG_PROFILE=1 php script.php

# From URL
curl "http://localhost/api/users?XDEBUG_PROFILE=1"

# From CLI with specific trigger
XDEBUG_PROFILE=1 XDEBUG_SESSION=myide php artisan migrate
```

### Analyzing with KCachegrind

```bash
# Install KCachegrind (Linux/Mac)
brew install qcachegrind  # macOS
apt-get install kcachegrind  # Ubuntu

# Open generated profile
qcachegrind /var/xdebug/cachegrind.out.2025-11-22.001

# Or via webgrind (web interface)
# http://localhost:8080/webgrind/
```

### Code-Level Profiling

```php
class Profiler {
    private static array $timings = [];
    private static array $memoryUsage = [];

    public static function start(string $label): void {
        self::$timings[$label] = [
            'start' => microtime(true),
            'memory_start' => memory_get_usage(true)
        ];
    }

    public static function end(string $label): void {
        if (!isset(self::$timings[$label])) {
            return;
        }

        $time = microtime(true) - self::$timings[$label]['start'];
        $memory = (memory_get_usage(true) - self::$timings[$label]['memory_start']) / 1024 / 1024;

        self::$memoryUsage[$label] = [
            'time' => $time,
            'memory_mb' => $memory
        ];

        unset(self::$timings[$label]);
    }

    public static function report(): void {
        echo "Performance Report:\n";
        foreach (self::$memoryUsage as $label => $data) {
            printf(
                "%-30s: %8.4f sec, %8.2f MB\n",
                $label,
                $data['time'],
                $data['memory_mb']
            );
        }
    }
}

// Usage
Profiler::start('database_queries');
$users = User::with('posts')->get();
Profiler::end('database_queries');

Profiler::start('data_processing');
$processed = process_users($users);
Profiler::end('data_processing');

Profiler::report();
```

---

## Caching Strategies

### In-Memory Caching with Redis

```php
class CacheRepository {
    private Redis $cache;

    public function __construct(string $host = 'localhost', int $port = 6379) {
        $this->cache = new Redis();
        $this->cache->connect($host, $port);
    }

    public function remember(string $key, int $ttl, callable $callback): mixed {
        if ($this->cache->exists($key)) {
            return json_decode($this->cache->get($key), true);
        }

        $value = $callback();
        $this->cache->setex($key, $ttl, json_encode($value));

        return $value;
    }

    public function flush(): void {
        $this->cache->flushDB();
    }

    public function delete(string $key): void {
        $this->cache->del($key);
    }

    public function tags(string $tag, callable $callback): void {
        // Tag-based invalidation
        $key = "tag:$tag";
        if ($this->cache->exists($key)) {
            $keys = $this->cache->sMembers($key);
            foreach ($keys as $k) {
                $this->cache->del($k);
            }
            $this->cache->del($key);
        }

        $callback();
    }
}

// Usage
$cache = new CacheRepository();

$users = $cache->remember('users:all', 3600, function() {
    return User::select('id', 'name', 'email')->get();
});

// Tag-based invalidation
$cache->tags('users', function() {
    // When user changes, entire "users" tag is invalidated
    User::find(1)->update(['name' => 'Updated Name']);
});
```

### Query Result Caching

```php
class CachedRepository {
    private CacheRepository $cache;

    public function __construct(CacheRepository $cache) {
        $this->cache = $cache;
    }

    public function getAllUsers(): array {
        return $this->cache->remember(
            'repository:users:all',
            3600,
            fn() => User::all()->toArray()
        );
    }

    public function getUserById(int $id): ?User {
        return $this->cache->remember(
            "repository:users:$id",
            3600,
            fn() => User::find($id)
        );
    }

    public function invalidateUser(int $id): void {
        $this->cache->delete("repository:users:$id");
        $this->cache->delete('repository:users:all');
    }
}
```

---

## String Optimization

### String Operations Best Practices

```php
// Bad: String concatenation in loops
$result = '';
for ($i = 0; $i < 10000; $i++) {
    $result .= "Item $i\n";  // Creates new string each iteration
}

// Good: Use array and implode
$items = [];
for ($i = 0; $i < 10000; $i++) {
    $items[] = "Item $i";
}
$result = implode("\n", $items);

// Good: Use heredoc/nowdoc for large strings
$template = <<<SQL
    SELECT u.id, u.name, u.email,
           COUNT(p.id) as post_count
    FROM users u
    LEFT JOIN posts p ON u.id = p.user_id
    WHERE u.active = true
    GROUP BY u.id
SQL;

// Efficient string search
if (str_contains($email, '@example.com')) {  // PHP 8.0+
    // Only use if(@) avoids function call
}
```

---

## Memory Management

### Memory Profiling

```php
function profileMemory(callable $fn, string $label): mixed {
    $before = memory_get_usage(true);
    $result = $fn();
    $after = memory_get_usage(true);

    $used = ($after - $before) / 1024 / 1024;
    printf("%s: %.2f MB\n", $label, $used);

    return $result;
}

// Usage
$data = profileMemory(
    fn() => fetchLargeDataset(),
    'Fetching dataset'
);

// Peak memory
$peak = memory_get_peak_usage(true) / 1024 / 1024;
printf("Peak memory: %.2f MB\n", $peak);
```

### Garbage Collection Tuning

```ini
; Enable garbage collection
zend.enable_gc=1

; Tune collection behavior
zend.gc_root_buffer_size=100000
zend.gc_roots_buffer_size=100000
```

---

## Best Practices Summary

**DO**:
- Enable JIT compiler in production (opcache.jit=1255)
- Use preloading for frequently used classes
- Implement eager loading for relationships
- Use query pagination for large datasets
- Cache expensive computations
- Profile code regularly with Xdebug
- Use connection pooling for databases
- Select only needed columns

**DON'T**:
- Create new strings in tight loops (use implode)
- Make N+1 queries (use eager loading)
- Cache without TTL
- Use short tags (<?)
- Skip performance profiling
- Hardcode values (use constants)
- Forget to close resources

---

**Total Lines**: 422
