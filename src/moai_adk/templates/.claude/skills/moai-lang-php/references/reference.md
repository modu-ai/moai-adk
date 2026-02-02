# PHP 8.3+ and Laravel 11 Complete Reference

## PHP 8.3 Feature Matrix

| Feature               | Status | Production Ready | Description                              |
| --------------------- | ------ | ---------------- | ---------------------------------------- |
| Readonly Classes      | Stable | Yes              | Immutable classes with readonly modifier |
| Typed Properties      | Stable | Yes              | Strong typing for class properties       |
| Constructor Promotion | Stable | Yes              | Shorthand property declaration           |
| Attributes            | Stable | Yes              | Native metadata annotations              |
| Enums                 | Stable | Yes              | First-class enumeration support          |
| Named Arguments       | Stable | Yes              | Function calls with named parameters     |
| Match Expression      | Stable | Yes              | Enhanced switch-case statements          |
| Nullsafe Operator     | Stable | Yes              | Safe property/method access with ?->     |
| Union Types           | Stable | Yes              | Multiple type declarations (string\|int) |
| Intersection Types    | Stable | Yes              | Combined type requirements (A&B)         |
| Never Type            | Stable | Yes              | Functions that never return              |
| Fibers                | Stable | Yes              | Lightweight concurrency primitives       |

---

## PHP 8.3 Modern Features

### Readonly Classes

```php
<?php

// Readonly class - all properties are readonly
readonly class Point
{
    public function __construct(
        public float $x,
        public float $y,
    ) {}
}

// Usage
$point = new Point(10.5, 20.3);
// $point->x = 15; // Error: Cannot modify readonly property
```

### Typed Properties

```php
<?php

class User
{
    public int $id;
    public string $name;
    public ?string $email = null;
    public array $roles = [];
    private \DateTimeImmutable $createdAt;

    public function __construct(int $id, string $name)
    {
        $this->id = $id;
        $this->name = $name;
        $this->createdAt = new \DateTimeImmutable();
    }
}
```

### Constructor Property Promotion

```php
<?php

// Traditional approach
class OldUser
{
    private string $name;
    private string $email;

    public function __construct(string $name, string $email)
    {
        $this->name = $name;
        $this->email = $email;
    }
}

// Modern approach with property promotion
class ModernUser
{
    public function __construct(
        private string $name,
        private string $email,
    ) {}

    public function getName(): string
    {
        return $this->name;
    }
}
```

### Attributes (Annotations)

```php
<?php

#[Attribute(Attribute::TARGET_CLASS)]
class Table
{
    public function __construct(
        public string $name,
    ) {}
}

#[Attribute(Attribute::TARGET_PROPERTY)]
class Column
{
    public function __construct(
        public string $name,
        public string $type = 'string',
        public bool $nullable = false,
    ) {}
}

#[Table('users')]
class User
{
    #[Column('id', 'integer')]
    public int $id;

    #[Column('email', 'string')]
    public string $email;

    #[Column('created_at', 'datetime', nullable: true)]
    public ?\DateTimeImmutable $createdAt;
}

// Reading attributes
$reflection = new ReflectionClass(User::class);
$attributes = $reflection->getAttributes(Table::class);
foreach ($attributes as $attribute) {
    $table = $attribute->newInstance();
    echo $table->name; // 'users'
}
```

### Advanced Enums

```php
<?php

// Basic enum
enum Status: string
{
    case Pending = 'pending';
    case Approved = 'approved';
    case Rejected = 'rejected';
}

// Enum with methods and static methods
enum OrderStatus: string
{
    case Pending = 'pending';
    case Processing = 'processing';
    case Shipped = 'shipped';
    case Delivered = 'delivered';
    case Cancelled = 'cancelled';

    public function label(): string
    {
        return match($this) {
            self::Pending => 'Pending',
            self::Processing => 'Processing',
            self::Shipped => 'Shipped',
            self::Delivered => 'Delivered',
            self::Cancelled => 'Cancelled',
        };
    }

    public function color(): string
    {
        return match($this) {
            self::Pending => 'yellow',
            self::Processing => 'blue',
            self::Shipped => 'purple',
            self::Delivered => 'green',
            self::Cancelled => 'red',
        };
    }

    public function canTransitionTo(self $newStatus): bool
    {
        return match($this) {
            self::Pending => in_array($newStatus, [self::Processing, self::Cancelled]),
            self::Processing => in_array($newStatus, [self::Shipped, self::Cancelled]),
            self::Shipped => $newStatus === self::Delivered,
            self::Delivered, self::Cancelled => false,
        };
    }

    public static function activeStatuses(): array
    {
        return [self::Pending, self::Processing, self::Shipped];
    }

    public static function fromString(string $value): self
    {
        return self::from(strtolower($value));
    }
}
```

### Match Expression

```php
<?php

// Traditional switch
function getLegacyPrice(string $type): int
{
    switch ($type) {
        case 'basic':
            return 100;
        case 'premium':
            return 200;
        case 'enterprise':
            return 500;
        default:
            throw new InvalidArgumentException('Invalid type');
    }
}

// Modern match expression
function getPrice(string $type): int
{
    return match($type) {
        'basic' => 100,
        'premium' => 200,
        'enterprise' => 500,
        default => throw new InvalidArgumentException('Invalid type'),
    };
}

// Match with multiple conditions
function getDiscount(string $type, int $quantity): float
{
    return match(true) {
        $quantity >= 100 => 0.3,
        $quantity >= 50 => 0.2,
        $quantity >= 10 => 0.1,
        $type === 'premium' => 0.05,
        default => 0.0,
    };
}
```

### Nullsafe Operator

```php
<?php

class Adddess
{
    public function __construct(
        public ?string $street = null,
    ) {}
}

class User
{
    public function __construct(
        public ?Adddess $adddess = null,
    ) {}
}

// Traditional approach
$street = null;
if ($user !== null) {
    if ($user->adddess !== null) {
        $street = $user->adddess->street;
    }
}

// Nullsafe operator
$street = $user?->adddess?->street;
```

### Union and Intersection Types

```php
<?php

// Union types
function processId(int|string $id): string
{
    return is_int($id) ? "ID-{$id}" : $id;
}

// Multiple union types
function formatValue(int|float|string|null $value): string
{
    return match(true) {
        is_null($value) => 'N/A',
        is_int($value) || is_float($value) => number_format($value, 2),
        default => (string) $value,
    };
}

// Intersection types
interface Loggable
{
    public function log(): void;
}

interface Serializable
{
    public function serialize(): string;
}

function process(Loggable&Serializable $object): void
{
    $object->log();
    $data = $object->serialize();
}
```

### Never Type

```php
<?php

function redirect(string $url): never
{
    header("Location: {$url}");
    exit;
}

function abort(int $code, string $message): never
{
    http_response_code($code);
    echo json_encode(['error' => $message]);
    exit;
}

function handleError(\Throwable $e): never
{
    logger()->error($e->getMessage());
    throw $e;
}
```

---

## Laravel 11 Complete Reference

### Application Structure

```
laravel-11-app/
├── app/
│   ├── Console/
│   │   ├── Commands/          # Custom Artisan commands
│   │   └── Kernel.php          # Console kernel
│   ├── Events/                 # Event classes
│   ├── Exceptions/             # Exception handlers
│   │   └── Handler.php
│   ├── Http/
│   │   ├── Controllers/        # HTTP controllers
│   │   ├── Middleware/         # HTTP middleware
│   │   ├── Requests/           # Form request validation
│   │   └── Resources/          # API resource transformers
│   ├── Jobs/                   # Queue jobs
│   ├── Listeners/              # Event listeners
│   ├── Mail/                   # Mail classes
│   ├── Models/                 # Eloquent models
│   ├── Notifications/          # Notification classes
│   ├── Policies/               # Authorization policies
│   ├── Providers/              # Service providers
│   │   ├── AppServiceProvider.php
│   │   └── RouteServiceProvider.php
│   └── Services/               # Business logic services
├── bootstrap/
│   ├── app.php                 # Application bootstrap
│   └── cache/                  # Framework cache
├── config/                     # Configuration files
│   ├── app.php
│   ├── database.php
│   ├── queue.php
│   └── ...
├── database/
│   ├── factories/              # Model factories
│   ├── migrations/             # Database migrations
│   └── seeders/                # Database seeders
├── public/                     # Public web root
│   └── index.php
├── resources/
│   ├── css/                    # CSS assets
│   ├── js/                     # JavaScript assets
│   └── views/                  # Blade templates
├── routes/
│   ├── api.php                 # API routes
│   ├── channels.php            # Broadcast channels
│   ├── console.php             # Console commands
│   └── web.php                 # Web routes
├── storage/
│   ├── app/                    # Application storage
│   ├── framework/              # Framework files
│   └── logs/                   # Application logs
├── tests/
│   ├── Feature/                # Feature tests
│   └── Unit/                   # Unit tests
├── artisan                     # Artisan CLI
├── composer.json               # PHP dependencies
└── phpunit.xml                 # PHPUnit configuration
```

### Laravel 11 Eloquent ORM Reference

#### Model Definitions

```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Database\Eloquent\Factories\HasFactory;

class Product extends Model
{
    use HasFactory, SoftDeletes;

    // Table name (optional if following convention)
    protected $table = 'products';

    // Primary key (optional if 'id')
    protected $primaryKey = 'id';

    // Key type
    protected $keyType = 'int';

    // Auto-incrementing
    public $incrementing = true;

    // Timestamps
    public $timestamps = true;

    // Date format
    protected $dateFormat = 'Y-m-d H:i:s';

    // Connection name
    protected $connection = 'mysql';

    // Mass assignable attributes
    protected $fillable = [
        'name',
        'description',
        'price',
        'category_id',
    ];

    // Guarded attributes (opposite of fillable)
    protected $guarded = [
        'id',
        'created_at',
        'updated_at',
    ];

    // Hidden attributes (for JSON serialization)
    protected $hidden = [
        'deleted_at',
    ];

    // Visible attributes (for JSON serialization)
    protected $visible = [
        'id',
        'name',
        'price',
    ];

    // Append accessors to JSON
    protected $appends = [
        'formatted_price',
    ];

    // Attribute casting
    protected function casts(): array
    {
        return [
            'price' => 'decimal:2',
            'is_available' => 'boolean',
            'metadata' => 'array',
            'published_at' => 'datetime',
        ];
    }
}
```

#### Eloquent Relationships

```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;
use Illuminate\Database\Eloquent\Relations\HasOne;
use Illuminate\Database\Eloquent\Relations\HasOneThrough;
use Illuminate\Database\Eloquent\Relations\HasManyThrough;
use Illuminate\Database\Eloquent\Relations\MorphTo;
use Illuminate\Database\Eloquent\Relations\MorphMany;
use Illuminate\Database\Eloquent\Relations\MorphToMany;

class User extends Model
{
    // One-to-One
    public function profile(): HasOne
    {
        return $this->hasOne(Profile::class);
    }

    // One-to-Many
    public function posts(): HasMany
    {
        return $this->hasMany(Post::class);
    }

    // Inverse of One-to-Many (Many-to-One)
    public function company(): BelongsTo
    {
        return $this->belongsTo(Company::class);
    }

    // Many-to-Many
    public function roles(): BelongsToMany
    {
        return $this->belongsToMany(Role::class)
            ->withPivot('assigned_at', 'assigned_by')
            ->withTimestamps()
            ->using(RoleUser::class); // Custom pivot model
    }

    // Has One Through
    public function latestPost(): HasOneThrough
    {
        return $this->hasOneThrough(
            Post::class,
            Author::class,
            'user_id',      // Foreign key on authors table
            'author_id',    // Foreign key on posts table
            'id',           // Local key on users table
            'id'            // Local key on authors table
        )->latest();
    }

    // Has Many Through
    public function comments(): HasManyThrough
    {
        return $this->hasManyThrough(
            Comment::class,
            Post::class,
            'user_id',      // Foreign key on posts table
            'post_id',      // Foreign key on comments table
            'id',           // Local key on users table
            'id'            // Local key on posts table
        );
    }

    // Polymorphic Relationship
    public function images(): MorphMany
    {
        return $this->morphMany(Image::class, 'imageable');
    }

    // Polymorphic Many-to-Many
    public function tags(): MorphToMany
    {
        return $this->morphToMany(Tag::class, 'taggable');
    }
}

class Image extends Model
{
    // Inverse of Polymorphic
    public function imageable(): MorphTo
    {
        return $this->morphTo();
    }
}
```

#### Query Scopes

```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Builder;

class Post extends Model
{
    // Local scope
    public function scopePublished(Builder $query): void
    {
        $query->where('status', 'published')
            ->whereNotNull('published_at')
            ->where('published_at', '<=', now());
    }

    // Local scope with parameters
    public function scopeOfType(Builder $query, string $type): void
    {
        $query->where('type', $type);
    }

    // Chainable scopes
    public function scopeRecent(Builder $query, int $days = 7): void
    {
        $query->where('created_at', '>=', now()->subDays($days));
    }

    // Global scope (in model boot method)
    protected static function booted(): void
    {
        static::addGlobalScope('active', function (Builder $builder) {
            $builder->where('is_active', true);
        });
    }
}

// Usage
Post::published()->recent(30)->get();
Post::ofType('article')->published()->get();
Post::withoutGlobalScope('active')->get();
```

#### Eloquent Accessors & Mutators

```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Casts\Attribute;
use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    // Modern accessor (Laravel 9+)
    protected function firstName(): Attribute
    {
        return Attribute::make(
            get: fn (string $value) => ucfirst($value),
            set: fn (string $value) => strtolower($value),
        );
    }

    // Computed attribute
    protected function fullName(): Attribute
    {
        return Attribute::make(
            get: fn () => "{$this->first_name} {$this->last_name}",
        );
    }

    // Legacy accessor
    public function getEmailAttribute($value): string
    {
        return strtolower($value);
    }

    // Legacy mutator
    public function setPasswordAttribute($value): void
    {
        $this->attributes['password'] = bcrypt($value);
    }
}
```

#### Eloquent Events

```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    protected static function booted(): void
    {
        // Creating event (before save)
        static::creating(function (User $user) {
            $user->uuid = \Illuminate\Support\Str::uuid();
        });

        // Created event (after save)
        static::created(function (User $user) {
            $user->profile()->create(['bio' => '']);
        });

        // Updating event
        static::updating(function (User $user) {
            if ($user->isDirty('email')) {
                $user->email_verified_at = null;
            }
        });

        // Updated event
        static::updated(function (User $user) {
            cache()->forget("user.{$user->id}");
        });

        // Saving event (before creating or updating)
        static::saving(function (User $user) {
            $user->email = strtolower($user->email);
        });

        // Saved event (after creating or updating)
        static::saved(function (User $user) {
            logger()->info("User {$user->id} saved");
        });

        // Deleting event
        static::deleting(function (User $user) {
            $user->posts()->delete();
        });

        // Deleted event
        static::deleted(function (User $user) {
            logger()->info("User {$user->id} deleted");
        });

        // Restoring event (soft deletes)
        static::restoring(function (User $user) {
            $user->posts()->restore();
        });

        // Restored event
        static::restored(function (User $user) {
            logger()->info("User {$user->id} restored");
        });
    }
}
```

### Laravel Validation Rules Reference

```php
<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;
use Illuminate\Validation\Rules\Password;

class ValidationExamplesRequest extends FormRequest
{
    public function rules(): array
    {
        return [
            // Basic validation
            'name' => 'required|string|max:255',
            'email' => 'required|email|unique:users,email',
            'age' => 'required|integer|min:18|max:100',

            // Array validation
            'tags' => 'array|min:1|max:5',
            'tags.*' => 'string|max:50',

            // Nested validation
            'user' => 'required|array',
            'user.name' => 'required|string',
            'user.email' => 'required|email',
            'user.adddess' => 'array',
            'user.adddess.street' => 'required_with:user.adddess|string',

            // Conditional validation
            'billing_adddess' => 'required_if:same_as_shipping,false',
            'shipping_method' => 'required_unless:is_digital,true',
            'tracking_number' => 'required_with:shipping_method',

            // File validation
            'avatar' => 'nullable|image|mimes:jpeg,png,jpg|max:2048',
            'document' => 'required|file|mimes:pdf,doc,docx|max:10240',

            // Date validation
            'birth_date' => 'required|date|before:today',
            'appointment' => 'required|date|after:tomorrow',
            'expired_at' => 'nullable|date|after_or_equal:start_date',

            // Password validation
            'password' => [
                'required',
                'confirmed',
                Password::min(8)
                    ->letters()
                    ->mixedCase()
                    ->numbers()
                    ->symbols()
                    ->uncompromised(),
            ],

            // Enum validation
            'status' => ['required', Rule::enum(OrderStatus::class)],

            // Unique with exceptions
            'slug' => [
                'required',
                'string',
                Rule::unique('posts')->ignore($this->post),
            ],

            // In validation
            'role' => ['required', Rule::in(['admin', 'editor', 'viewer'])],

            // Exists validation
            'category_id' => 'required|exists:categories,id',
            'user_id' => [
                'required',
                Rule::exists('users', 'id')->where('is_active', true),
            ],

            // Custom validation
            'custom_field' => [
                'required',
                function ($attribute, $value, $fail) {
                    if (strtoupper($value) !== $value) {
                        $fail("The {$attribute} must be uppercase.");
                    }
                },
            ],

            // Sometimes validation
            'optional_field' => 'sometimes|required|string',

            // Nullable vs Required
            'middle_name' => 'nullable|string|max:100',

            // Boolean validation
            'is_active' => 'required|boolean',
            'accepted_terms' => 'accepted',

            // Numeric validation
            'price' => 'required|numeric|min:0|max:99999.99',
            'quantity' => 'required|integer|between:1,1000',

            // String validation
            'username' => 'required|string|alpha_dash|min:3|max:20',
            'code' => 'required|string|alpha_num|size:6',

            // URL validation
            'website' => 'nullable|url|active_url',

            // JSON validation
            'metadata' => 'nullable|json',

            // IP validation
            'ip_adddess' => 'nullable|ip',
            'ipv4_adddess' => 'nullable|ipv4',

            // Regex validation
            'phone' => 'required|regex:/^([0-9\s\-\+\(\)]*)$/',

            // Multiple of validation
            'quantity' => 'required|integer|multiple_of:5',
        ];
    }
}
```

### Artisan Commands Reference

```bash
# Application management
php artisan serve                    # Start development server
php artisan tinker                   # REPL console
php artisan optimize                 # Optimize framework
php artisan optimize:clear           # Clear optimization cache
php artisan about                    # Display application info

# Cache management
php artisan cache:clear              # Clear application cache
php artisan config:clear             # Clear configuration cache
php artisan route:clear              # Clear route cache
php artisan view:clear               # Clear compiled views
php artisan event:clear              # Clear cached events

# Cache building
php artisan config:cache             # Cache configuration
php artisan route:cache              # Cache routes
php artisan view:cache               # Compile views
php artisan event:cache              # Cache events

# Database
php artisan migrate                  # Run migrations
php artisan migrate:fresh            # Drop all tables and re-migrate
php artisan migrate:refresh          # Reset and re-run migrations
php artisan migrate:reset            # Rollback all migrations
php artisan migrate:rollback         # Rollback last migration
php artisan migrate:status           # Show migration status
php artisan db:seed                  # Seed database
php artisan db:wipe                  # Drop all tables

# Model generation
php artisan make:model User -mfsc    # Model with migration, factory, seeder, controller
php artisan make:model Post --all    # Model with all resources

# Resource generation
php artisan make:controller UserController --resource --api
php artisan make:request StoreUserRequest
php artisan make:resource UserResource
php artisan make:migration create_users_table
php artisan make:seeder UserSeeder
php artisan make:factory UserFactory
php artisan make:middleware CheckAge
php artisan make:policy UserPolicy
php artisan make:event UserRegistered
php artisan make:listener SendWelcomeEmail
php artisan make:job ProcessOrder
php artisan make:mail WelcomeMail
php artisan make:notification OrderShipped
php artisan make:command SendEmails
php artisan make:rule Uppercase
php artisan make:provider CustomServiceProvider
php artisan make:test UserTest --unit
php artisan make:test UserFeatureTest

# Queue management
php artisan queue:work               # Process queue jobs
php artisan queue:work --queue=high,default
php artisan queue:listen             # Listen for jobs
php artisan queue:restart            # Restart queue workers
php artisan queue:retry all          # Retry failed jobs
php artisan queue:failed             # List failed jobs
php artisan queue:flush              # Delete all failed jobs

# Schedule
php artisan schedule:run             # Run scheduled commands
php artisan schedule:work            # Run scheduler in foreground
php artisan schedule:list            # List scheduled commands

# Storage
php artisan storage:link             # Create symbolic link to storage

# Vendor publishing
php artisan vendor:publish           # Publish vendor assets

# Route information
php artisan route:list               # List all routes
php artisan route:list --except-vendor
php artisan route:list --path=api

# Custom commands
php artisan inspire                  # Display inspiring quote
```

### Composer Configuration

```json
{
  "name": "your-org/laravel-app",
  "type": "project",
  "description": "Laravel 11 Application",
  "keywords": ["laravel", "framework"],
  "license": "MIT",
  "require": {
    "php": "^8.3",
    "laravel/framework": "^11.0",
    "laravel/sanctum": "^4.0",
    "laravel/tinker": "^2.9"
  },
  "require-dev": {
    "fakerphp/faker": "^1.23",
    "laravel/pint": "^1.13",
    "laravel/sail": "^1.26",
    "mockery/mockery": "^1.6",
    "nunomaduro/collision": "^8.0",
    "pestphp/pest": "^2.34",
    "pestphp/pest-plugin-laravel": "^2.3",
    "phpstan/phpstan": "^1.10",
    "phpunit/phpunit": "^11.0",
    "spatie/laravel-ignition": "^2.4"
  },
  "autoload": {
    "psr-4": {
      "App\\": "app/",
      "Database\\Factories\\": "database/factories/",
      "Database\\Seeders\\": "database/seeders/"
    }
  },
  "autoload-dev": {
    "psr-4": {
      "Tests\\": "tests/"
    }
  },
  "scripts": {
    "post-autoload-dump": [
      "Illuminate\\Foundation\\ComposerScripts::postAutoloadDump",
      "@php artisan package:discover --ansi"
    ],
    "post-update-cmd": [
      "@php artisan vendor:publish --tag=laravel-assets --ansi --force"
    ],
    "post-root-package-install": [
      "@php -r \"file_exists('.env') || copy('.env.example', '.env');\""
    ],
    "post-create-project-cmd": [
      "@php artisan key:generate --ansi",
      "@php -r \"file_exists('database/database.sqlite') || touch('database/database.sqlite');\"",
      "@php artisan migrate --graceful --ansi"
    ],
    "test": "pest",
    "test:coverage": "pest --coverage",
    "pint": "pint",
    "stan": "phpstan analyse"
  },
  "extra": {
    "laravel": {
      "dont-discover": []
    }
  },
  "config": {
    "optimize-autoloader": true,
    "preferred-install": "dist",
    "sort-packages": true,
    "allow-plugins": {
      "pestphp/pest-plugin": true,
      "php-http/discovery": true
    }
  },
  "minimum-stability": "stable",
  "prefer-stable": true
}
```

### PHPStan Configuration

```neon
# phpstan.neon
includes:
    - ./vendor/larastan/larastan/extension.neon

parameters:
    paths:
        - app
        - config
        - database
        - routes
        - tests

    level: 6

    ignoreErrors:
        - '#Unsafe usage of new static#'

    excludePaths:
        - ./*/*/FileToBeExcluded.php

    checkMissingIterableValueType: false
```

### Laravel Pint Configuration

```json
{
  "preset": "laravel",
  "rules": {
    "array_syntax": {
      "syntax": "short"
    },
    "binary_operator_spaces": {
      "default": "single_space"
    },
    "blank_line_after_namespace": true,
    "blank_line_after_opening_tag": true,
    "blank_line_before_statement": {
      "statements": ["return"]
    },
    "braces": true,
    "cast_spaces": true,
    "class_attributes_separation": {
      "elements": {
        "method": "one"
      }
    },
    "concat_space": {
      "spacing": "none"
    },
    "declare_equal_normalize": true,
    "elseif": true,
    "encoding": true,
    "full_opening_tag": true,
    "function_declaration": true,
    "indentation_type": true,
    "line_ending": true,
    "lowercase_cast": true,
    "lowercase_keywords": true,
    "method_argument_space": {
      "on_multiline": "ensure_fully_multiline"
    },
    "native_function_casing": true,
    "no_blank_lines_after_class_opening": true,
    "no_closing_tag": true,
    "no_spaces_after_function_name": true,
    "no_spaces_inside_parenthesis": true,
    "no_trailing_whitespace": true,
    "no_trailing_whitespace_in_comment": true,
    "single_blank_line_at_eof": true,
    "single_class_element_per_statement": {
      "elements": ["property"]
    },
    "single_import_per_statement": true,
    "single_line_after_imports": true,
    "switch_case_semicolon_to_colon": true,
    "switch_case_space": true,
    "visibility_required": true,
    "ordered_imports": {
      "sort_algorithm": "alpha"
    }
  }
}
```

### Performance Optimization Tips

#### Eloquent Performance

```php
<?php

// Eager loading (N+1 query prevention)
$users = User::with('posts', 'profile')->get();
$users = User::with(['posts' => function ($query) {
    $query->where('published', true)->orderBy('created_at', 'desc');
}])->get();

// Lazy eager loading
$users = User::all();
$users->load('posts');

// Preventing lazy loading in production
Model::preventLazyLoading(!app()->isProduction());

// Chunk processing for large datasets
User::chunk(100, function ($users) {
    foreach ($users as $user) {
        // Process user
    }
});

// Cursor for memory efficiency
foreach (User::cursor() as $user) {
    // Process user
}

// Select specific columns
User::select('id', 'name', 'email')->get();

// Count without loading
$count = User::count();
$max = User::max('age');
$avg = User::avg('score');

// Exists check
if (User::where('email', $email)->exists()) {
    // User exists
}
```

#### Caching Strategies

```php
<?php

use Illuminate\Support\Facades\Cache;

// Cache::get with default
$value = Cache::get('key', 'default');

// Cache::remember
$users = Cache::remember('users.active', 3600, function () {
    return User::active()->get();
});

// Cache::rememberForever
$settings = Cache::rememberForever('settings', function () {
    return Setting::all();
});

// Cache::put
Cache::put('key', 'value', 3600);
Cache::put('key', 'value', now()->addHours(1));

// Cache::forever
Cache::forever('key', 'value');

// Cache::forget
Cache::forget('key');

// Cache::flush (clear all)
Cache::flush();

// Cache tags (Redis, Memcached only)
Cache::tags(['users', 'active'])->put('key', 'value', 3600);
$value = Cache::tags(['users'])->get('key');
Cache::tags(['users'])->flush();

// Cache::increment / decrement
Cache::increment('views', 1);
Cache::decrement('stock', 5);
```

---

## Context7 Library Integration

### Available PHP Libraries

| Library           | Context7 ID               | Topics                                    |
| ----------------- | ------------------------- | ----------------------------------------- |
| Laravel Framework | /laravel/framework        | eloquent, routing, middleware, validation |
| Laravel Sanctum   | /laravel/sanctum          | API authentication, SPA authentication    |
| Laravel Horizon   | /laravel/horizon          | queue monitoring, job metrics             |
| Symfony           | /symfony/symfony          | components, bundles, services             |
| Doctrine ORM      | /doctrine/orm             | entities, repositories, DQL               |
| PHPUnit           | /phpunit/phpunit          | testing, assertions, mocking              |
| Pest PHP          | /pestphp/pest             | elegant testing, expectations             |
| Composer          | /composer/composer        | dependency management, autoloading        |
| Guzzle HTTP       | /guzzlehttp/guzzle        | HTTP client, requests, middleware         |
| Monolog           | /Seldaek/monolog          | logging, handlers, formatters             |
| Carbon            | /briannesbitt/carbon      | date/time manipulation                    |
| PhpSpreadsheet    | /PHPOffice/PhpSpreadsheet | Excel, CSV processing                     |

### Usage Example

```php
// Step 1: Resolve library ID
$libraryId = mcp__context7__resolve_library_id("laravel/framework");
// Returns: /laravel/framework

// Step 2: Get documentation
$docs = mcp__context7__get_library_docs(
    context7CompatibleLibraryID: "/laravel/framework",
    topic: "eloquent relationships eager loading",
    tokens: 5000
);
```

---

## Security Best Practices

### OWASP Top 10 Prevention

```php
<?php

// 1. SQL Injection Prevention
// ✅ Use query builder or Eloquent
User::where('email', $email)->first();

// ❌ Avoid raw queries with user input
// DB::select("SELECT * FROM users WHERE email = '$email'");

// ✅ If raw query needed, use bindings
DB::select('SELECT * FROM users WHERE email = ?', [$email]);

// 2. XSS Prevention
// ✅ Blade auto-escapes
{{ $userInput }}

// ✅ Raw output only for trusted content
{!! $trustedHtml !!}

// 3. CSRF Protection (enabled by default)
// ✅ Forms include CSRF token
@csrf

// 4. Mass Assignment Protection
// ✅ Use $fillable or $guarded
protected $fillable = ['name', 'email'];

// 5. Authentication & Authorization
// ✅ Use Laravel's built-in auth
Auth::check()
Gate::allows('update', $post)
$this->authorize('update', $post)

// 6. Password Hashing
// ✅ Use Hash facade or password cast
Hash::make($password)
protected $casts = ['password' => 'hashed'];

// 7. Rate Limiting
// ✅ Apply rate limiting middleware
Route::middleware('throttle:60,1')->group(function () {
    Route::get('/api/users', [UserController::class, 'index']);
});

// 8. File Upload Validation
// ✅ Validate file types and sizes
$request->validate([
    'file' => 'required|file|mimes:pdf,jpg,png|max:2048',
]);

// 9. Environment Variables
// ✅ Never commit .env file
// ✅ Use config() helper
config('app.key')

// 10. Secure Headers
// ✅ Add security headers middleware
```

---

Last Updated: 2026-01-10
Version: 1.0.0
