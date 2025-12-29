---
name: moai-lang-php
description: PHP 8.3+ development specialist covering Laravel 11, Symfony 7, Eloquent ORM, and modern PHP patterns. Use when developing PHP APIs, web applications, or Laravel/Symfony projects.
version: 1.0.0
updated: 2025-12-07
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

PHP 8.3+ Development Specialist - Laravel 11, Symfony 7, Eloquent, Doctrine, and modern PHP patterns.

Auto-Triggers: `.php` files, `composer.json`, `artisan`, `symfony.yaml`, Laravel/Symfony discussions

Core Capabilities:
- PHP 8.3 Features: readonly classes, typed properties, attributes, enums, named arguments
- Laravel 11: Controllers, Models, Migrations, Form Requests, API Resources, Eloquent
- Symfony 7: Attribute-based routing, Doctrine ORM, Services, Dependency Injection
- ORMs: Eloquent (Laravel), Doctrine (Symfony)
- Testing: PHPUnit, Pest, feature/unit testing patterns
- Package Management: Composer with autoloading
- Coding Standards: PSR-12, Laravel Pint, PHP CS Fixer
- Docker: PHP-FPM, nginx, multi-stage builds

### Quick Patterns

Laravel Controller:
```php
<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreUserRequest;
use App\Http\Resources\UserResource;
use App\Models\User;
use Illuminate\Http\JsonResponse;

class UserController extends Controller
{
    public function store(StoreUserRequest $request): JsonResponse
    {
        $user = User::create($request->validated());
        return response()->json(new UserResource($user), 201);
    }
}
```

Laravel Form Request:
```php
<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class StoreUserRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    public function rules(): array
    {
        return [
            'name' => ['required', 'string', 'max:255'],
            'email' => ['required', 'email', 'unique:users,email'],
            'password' => ['required', 'min:8', 'confirmed'],
        ];
    }
}
```

Symfony Controller:
```php
<?php

namespace App\Controller;

use App\Entity\User;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\Routing\Attribute\Route;

#[Route('/api/users')]
class UserController extends AbstractController
{
    #[Route('', methods: ['POST'])]
    public function create(EntityManagerInterface $em): JsonResponse
    {
        $user = new User();
        $em->persist($user);
        $em->flush();
        return $this->json($user, 201);
    }
}
```

---

## Implementation Guide (5 minutes)

### PHP 8.3 Modern Features

Readonly Classes:
```php
<?php

readonly class UserDTO
{
    public function __construct(
        public int $id,
        public string $name,
        public string $email,
    ) {}
}
```

Enums with Methods:
```php
<?php

enum OrderStatus: string
{
    case Pending = 'pending';
    case Processing = 'processing';
    case Completed = 'completed';

    public function label(): string
    {
        return match($this) {
            self::Pending => 'Pending',
            self::Processing => 'Processing',
            self::Completed => 'Completed',
        };
    }
}
```

Attributes:
```php
<?php

#[Attribute(Attribute::TARGET_PROPERTY)]
class Validate
{
    public function __construct(
        public string $rule,
        public ?string $message = null,
    ) {}
}

class UserRequest
{
    #[Validate('required|email')]
    public string $email;
}
```

### Laravel 11 Patterns

Eloquent Model with Relationships:
```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class Post extends Model
{
    protected $fillable = ['title', 'content', 'user_id', 'status'];

    protected $casts = [
        'status' => PostStatus::class,
        'published_at' => 'datetime',
    ];

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function comments(): HasMany
    {
        return $this->hasMany(Comment::class);
    }

    public function scopePublished($query)
    {
        return $query->where('status', PostStatus::Published);
    }
}
```

API Resource:
```php
<?php

namespace App\Http\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

class PostResource extends JsonResource
{
    public function toArray(Request $request): array
    {
        return [
            'id' => $this->id,
            'title' => $this->title,
            'author' => new UserResource($this->whenLoaded('user')),
            'comments_count' => $this->whenCounted('comments'),
            'created_at' => $this->created_at->toIso8601String(),
        ];
    }
}
```

Migration:
```php
<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('posts', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('title');
            $table->text('content');
            $table->string('status')->default('draft');
            $table->timestamps();
            $table->softDeletes();
        });
    }
};
```

Service Layer Pattern:
```php
<?php

namespace App\Services;

use App\Models\User;
use App\DTOs\UserDTO;
use Illuminate\Support\Facades\DB;

class UserService
{
    public function create(UserDTO $dto): User
    {
        return DB::transaction(function () use ($dto) {
            $user = User::create([
                'name' => $dto->name,
                'email' => $dto->email,
                'password' => Hash::make($dto->password),
            ]);

            $user->profile()->create(['bio' => $dto->bio ?? '']);
            return $user->load('profile');
        });
    }
}
```

### Symfony 7 Patterns

Entity with Doctrine Attributes:
```php
<?php

namespace App\Entity;

use App\Repository\UserRepository;
use Doctrine\ORM\Mapping as ORM;
use Symfony\Component\Validator\Constraints as Assert;

#[ORM\Entity(repositoryClass: UserRepository::class)]
#[ORM\Table(name: 'users')]
class User
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\Column(length: 255)]
    #[Assert\NotBlank]
    private ?string $name = null;

    #[ORM\Column(length: 180, unique: true)]
    #[Assert\Email]
    private ?string $email = null;
}
```

Service with Dependency Injection:
```php
<?php

namespace App\Service;

use App\Entity\User;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;

class UserService
{
    public function __construct(
        private readonly EntityManagerInterface $entityManager,
        private readonly UserPasswordHasherInterface $passwordHasher,
    ) {}

    public function createUser(string $email, string $password): User
    {
        $user = new User();
        $user->setEmail($email);
        $user->setPassword($this->passwordHasher->hashPassword($user, $password));

        $this->entityManager->persist($user);
        $this->entityManager->flush();

        return $user;
    }
}
```

### Testing Patterns

PHPUnit Feature Test (Laravel):
```php
<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class UserApiTest extends TestCase
{
    use RefreshDatabase;

    public function test_can_create_user(): void
    {
        $response = $this->postJson('/api/users', [
            'name' => 'John Doe',
            'email' => 'john@example.com',
            'password' => 'password123',
            'password_confirmation' => 'password123',
        ]);

        $response->assertStatus(201)
            ->assertJsonStructure(['data' => ['id', 'name', 'email']]);

        $this->assertDatabaseHas('users', ['email' => 'john@example.com']);
    }
}
```

Pest Test (Laravel):
```php
<?php

use App\Models\User;
use App\Models\Post;

it('can create a post', function () {
    $user = User::factory()->create();

    $response = $this->actingAs($user)
        ->postJson('/api/posts', [
            'title' => 'My First Post',
            'content' => 'This is the content.',
        ]);

    $response->assertStatus(201);
    expect(Post::count())->toBe(1);
});

it('requires authentication to create post', function () {
    $this->postJson('/api/posts', [
        'title' => 'My Post',
        'content' => 'Content here.',
    ])->assertStatus(401);
});
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Production deployment patterns (Docker, Kubernetes)
- Advanced Eloquent patterns (observers, accessors, mutators)
- Doctrine advanced mapping (embeddables, inheritance)
- Queue and job processing
- Event-driven architecture
- Caching strategies (Redis, Memcached)
- Security best practices (OWASP patterns)
- CI/CD integration patterns

See:
- [Advanced Patterns](modules/advanced-patterns.md) - Complete advanced patterns guide

---

## Context7 Library Mappings

```
/laravel/framework - Laravel web framework
/symfony/symfony - Symfony components and framework
/doctrine/orm - Doctrine ORM for PHP
/phpunit/phpunit - PHP testing framework
/pestphp/pest - Elegant PHP testing framework
/laravel/sanctum - Laravel API authentication
/laravel/horizon - Laravel queue dashboard
```

---

## Works Well With

- `moai-domain-backend` - REST API and microservices architecture
- `moai-domain-database` - SQL patterns and ORM optimization
- `moai-workflow-testing` - TDD and testing strategies
- `moai-platform-deploy` - Docker and deployment patterns
- `moai-essentials-debug` - AI-powered debugging
- `moai-foundation-quality` - TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

PHP Version Check:
```bash
php --version  # Should be 8.3+
php -m | grep -E 'pdo|mbstring|openssl'
```

Composer Autoload Issues:
```bash
composer dump-autoload -o
composer clear-cache
```

Laravel Cache Issues:
```bash
php artisan config:clear
php artisan cache:clear
php artisan route:clear
php artisan view:clear
```

Symfony Cache Issues:
```bash
php bin/console cache:clear
php bin/console cache:warmup
```

Database Connection:
```php
try {
    DB::connection()->getPdo();
    echo "Connected successfully";
} catch (\Exception $e) {
    echo "Connection failed: " . $e->getMessage();
}
```

Migration Rollback:
```bash
php artisan migrate:rollback --step=1
php artisan migrate:fresh --seed  # Development only

php bin/console doctrine:migrations:migrate prev
```

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
