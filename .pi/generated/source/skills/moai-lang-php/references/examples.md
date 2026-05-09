# PHP Production-Ready Code Examples

## Complete Laravel 11 Application

### Project Structure

```
laravel_app/
├── app/
│   ├── Console/
│   │   └── Commands/
│   │       └── ProcessOrdersCommand.php
│   ├── Events/
│   │   └── UserRegistered.php
│   ├── Http/
│   │   ├── Controllers/
│   │   │   └── Api/
│   │   │       ├── UserController.php
│   │   │       └── PostController.php
│   │   ├── Middleware/
│   │   │   └── EnsureUserIsVerified.php
│   │   ├── Requests/
│   │   │   ├── StoreUserRequest.php
│   │   │   └── UpdateUserRequest.php
│   │   └── Resources/
│   │       ├── UserResource.php
│   │       └── PostResource.php
│   ├── Jobs/
│   │   └── SendWelcomeEmail.php
│   ├── Listeners/
│   │   └── SendWelcomeNotification.php
│   ├── Models/
│   │   ├── User.php
│   │   ├── Post.php
│   │   └── Comment.php
│   ├── Repositories/
│   │   ├── UserRepository.php
│   │   └── PostRepository.php
│   ├── Services/
│   │   ├── UserService.php
│   │   └── AuthService.php
│   └── DTOs/
│       ├── UserDTO.php
│       └── PostDTO.php
├── database/
│   ├── factories/
│   │   └── UserFactory.php
│   ├── migrations/
│   │   └── 2024_01_01_create_users_table.php
│   └── seeders/
│       └── UserSeeder.php
├── routes/
│   ├── api.php
│   └── web.php
├── tests/
│   ├── Feature/
│   │   └── UserApiTest.php
│   └── Unit/
│       └── UserServiceTest.php
├── composer.json
└── Dockerfile
```

### Eloquent Models with Modern Features

```php
<?php

namespace App\Models;

use App\Enums\UserRole;
use App\Enums\UserStatus;
use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\SoftDeletes;

class User extends Model
{
    use HasFactory, SoftDeletes, HasUuids;

    /**
     * The attributes that are mass assignable.
     */
    protected $fillable = [
        'name',
        'email',
        'password',
        'role',
        'status',
        'email_verified_at',
    ];

    /**
     * The attributes that should be hidden for serialization.
     */
    protected $hidden = [
        'password',
        'remember_token',
    ];

    /**
     * Get the attributes that should be cast.
     */
    protected function casts(): array
    {
        return [
            'email_verified_at' => 'datetime',
            'password' => 'hashed',
            'role' => UserRole::class,
            'status' => UserStatus::class,
            'created_at' => 'datetime',
            'updated_at' => 'datetime',
            'deleted_at' => 'datetime',
        ];
    }

    /**
     * Get the posts for the user.
     */
    public function posts(): HasMany
    {
        return $this->hasMany(Post::class);
    }

    /**
     * Get the comments for the user.
     */
    public function comments(): HasMany
    {
        return $this->hasMany(Comment::class);
    }

    /**
     * Scope a query to only include active users.
     */
    public function scopeActive($query)
    {
        return $query->where('status', UserStatus::Active);
    }

    /**
     * Scope a query to only include verified users.
     */
    public function scopeVerified($query)
    {
        return $query->whereNotNull('email_verified_at');
    }

    /**
     * Get the user's full name.
     */
    public function getFullNameAttribute(): string
    {
        return "{$this->first_name} {$this->last_name}";
    }

    /**
     * Check if user is admin.
     */
    public function isAdmin(): bool
    {
        return $this->role === UserRole::Admin;
    }
}
```

```php
<?php

namespace App\Models;

use App\Enums\PostStatus;
use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\SoftDeletes;

class Post extends Model
{
    use HasFactory, SoftDeletes, HasUuids;

    protected $fillable = [
        'user_id',
        'title',
        'slug',
        'content',
        'excerpt',
        'status',
        'published_at',
    ];

    protected function casts(): array
    {
        return [
            'status' => PostStatus::class,
            'published_at' => 'datetime',
            'created_at' => 'datetime',
            'updated_at' => 'datetime',
        ];
    }

    /**
     * Get the user that owns the post.
     */
    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    /**
     * Get the comments for the post.
     */
    public function comments(): HasMany
    {
        return $this->hasMany(Comment::class);
    }

    /**
     * Scope a query to only include published posts.
     */
    public function scopePublished($query)
    {
        return $query->where('status', PostStatus::Published)
            ->whereNotNull('published_at')
            ->where('published_at', '<=', now());
    }

    /**
     * Scope a query to search posts.
     */
    public function scopeSearch($query, string $search)
    {
        return $query->where(function ($q) use ($search) {
            $q->where('title', 'like', "%{$search}%")
              ->orWhere('content', 'like', "%{$search}%");
        });
    }

    /**
     * Get the route key for the model.
     */
    public function getRouteKeyName(): string
    {
        return 'slug';
    }
}
```

### Modern PHP 8.3 Enums

```php
<?php

namespace App\Enums;

enum UserRole: string
{
    case Admin = 'admin';
    case Editor = 'editor';
    case Author = 'author';
    case Subscriber = 'subscriber';

    /**
     * Get the label for the role.
     */
    public function label(): string
    {
        return match($this) {
            self::Admin => 'Administrator',
            self::Editor => 'Editor',
            self::Author => 'Author',
            self::Subscriber => 'Subscriber',
        };
    }

    /**
     * Get all permissions for the role.
     */
    public function permissions(): array
    {
        return match($this) {
            self::Admin => ['*'],
            self::Editor => ['posts.create', 'posts.edit', 'posts.delete', 'posts.publish'],
            self::Author => ['posts.create', 'posts.edit', 'posts.delete'],
            self::Subscriber => ['posts.read'],
        };
    }

    /**
     * Check if role can perform action.
     */
    public function can(string $permission): bool
    {
        $permissions = $this->permissions();

        if (in_array('*', $permissions)) {
            return true;
        }

        return in_array($permission, $permissions);
    }

    /**
     * Get all available roles.
     */
    public static function values(): array
    {
        return array_column(self::cases(), 'value');
    }
}
```

```php
<?php

namespace App\Enums;

enum PostStatus: string
{
    case Draft = 'draft';
    case Published = 'published';
    case Archived = 'archived';

    public function label(): string
    {
        return match($this) {
            self::Draft => 'Draft',
            self::Published => 'Published',
            self::Archived => 'Archived',
        };
    }

    public function color(): string
    {
        return match($this) {
            self::Draft => 'gray',
            self::Published => 'green',
            self::Archived => 'red',
        };
    }
}
```

### DTOs with Readonly Classes

```php
<?php

namespace App\DTOs;

readonly class UserDTO
{
    public function __construct(
        public int $id,
        public string $name,
        public string $email,
        public UserRole $role,
        public UserStatus $status,
        public ?\DateTimeImmutable $emailVerifiedAt = null,
    ) {}

    /**
     * Create from Eloquent model.
     */
    public static function fromModel(User $user): self
    {
        return new self(
            id: $user->id,
            name: $user->name,
            email: $user->email,
            role: $user->role,
            status: $user->status,
            emailVerifiedAt: $user->email_verified_at?->toDateTimeImmutable(),
        );
    }

    /**
     * Create from array.
     */
    public static function fromArray(array $data): self
    {
        return new self(
            id: $data['id'],
            name: $data['name'],
            email: $data['email'],
            role: UserRole::from($data['role']),
            status: UserStatus::from($data['status']),
            emailVerifiedAt: isset($data['email_verified_at'])
                ? new \DateTimeImmutable($data['email_verified_at'])
                : null,
        );
    }

    /**
     * Convert to array.
     */
    public function toArray(): array
    {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'email' => $this->email,
            'role' => $this->role->value,
            'status' => $this->status->value,
            'email_verified_at' => $this->emailVerifiedAt?->format('Y-m-d H:i:s'),
        ];
    }
}
```

### Repository Pattern

```php
<?php

namespace App\Repositories;

use App\Models\User;
use App\DTOs\UserDTO;
use Illuminate\Contracts\Pagination\LengthAwarePaginator;
use Illuminate\Database\Eloquent\Collection;

class UserRepository
{
    /**
     * Find user by ID.
     */
    public function find(int $id): ?User
    {
        return User::find($id);
    }

    /**
     * Find user by email.
     */
    public function findByEmail(string $email): ?User
    {
        return User::where('email', $email)->first();
    }

    /**
     * Get all users with pagination.
     */
    public function paginate(int $perPage = 15): LengthAwarePaginator
    {
        return User::with(['posts'])
            ->latest()
            ->paginate($perPage);
    }

    /**
     * Get active users.
     */
    public function getActive(): Collection
    {
        return User::active()
            ->verified()
            ->orderBy('name')
            ->get();
    }

    /**
     * Create a new user.
     */
    public function create(array $data): User
    {
        return User::create([
            'name' => $data['name'],
            'email' => $data['email'],
            'password' => $data['password'], // Will be hashed by cast
            'role' => $data['role'] ?? UserRole::Subscriber,
            'status' => UserStatus::Active,
        ]);
    }

    /**
     * Update user.
     */
    public function update(User $user, array $data): User
    {
        $user->update($data);
        return $user->fresh();
    }

    /**
     * Delete user (soft delete).
     */
    public function delete(User $user): bool
    {
        return $user->delete();
    }

    /**
     * Force delete user.
     */
    public function forceDelete(User $user): bool
    {
        return $user->forceDelete();
    }

    /**
     * Search users.
     */
    public function search(string $query, int $perPage = 15): LengthAwarePaginator
    {
        return User::where('name', 'like', "%{$query}%")
            ->orWhere('email', 'like', "%{$query}%")
            ->paginate($perPage);
    }

    /**
     * Get users by role.
     */
    public function getByRole(UserRole $role): Collection
    {
        return User::where('role', $role)->get();
    }
}
```

### Service Layer with Business Logic

```php
<?php

namespace App\Services;

use App\DTOs\UserDTO;
use App\Events\UserRegistered;
use App\Models\User;
use App\Repositories\UserRepository;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class UserService
{
    public function __construct(
        private readonly UserRepository $repository,
    ) {}

    /**
     * Register a new user.
     */
    public function register(array $data): UserDTO
    {
        return DB::transaction(function () use ($data) {
            $user = $this->repository->create([
                'name' => $data['name'],
                'email' => $data['email'],
                'password' => Hash::make($data['password']),
                'role' => UserRole::Subscriber,
            ]);

            event(new UserRegistered($user));

            return UserDTO::fromModel($user);
        });
    }

    /**
     * Update user profile.
     */
    public function updateProfile(User $user, array $data): UserDTO
    {
        $updateData = [];

        if (isset($data['name'])) {
            $updateData['name'] = $data['name'];
        }

        if (isset($data['email']) && $data['email'] !== $user->email) {
            $updateData['email'] = $data['email'];
            $updateData['email_verified_at'] = null;
        }

        if (isset($data['password'])) {
            $updateData['password'] = Hash::make($data['password']);
        }

        $updatedUser = $this->repository->update($user, $updateData);

        return UserDTO::fromModel($updatedUser);
    }

    /**
     * Verify user email.
     */
    public function verifyEmail(User $user): void
    {
        if ($user->email_verified_at !== null) {
            return;
        }

        $user->update(['email_verified_at' => now()]);
    }

    /**
     * Change user role.
     */
    public function changeRole(User $user, UserRole $newRole): UserDTO
    {
        $user->update(['role' => $newRole]);
        return UserDTO::fromModel($user->fresh());
    }

    /**
     * Get user statistics.
     */
    public function getUserStatistics(User $user): array
    {
        return [
            'total_posts' => $user->posts()->count(),
            'published_posts' => $user->posts()->published()->count(),
            'total_comments' => $user->comments()->count(),
            'member_since' => $user->created_at->diffForHumans(),
        ];
    }
}
```

### Form Requests with Validation

```php
<?php

namespace App\Http\Requests;

use App\Enums\UserRole;
use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rules\Enum;
use Illuminate\Validation\Rules\Password;

class StoreUserRequest extends FormRequest
{
    /**
     * Determine if the user is authorized to make this request.
     */
    public function authorize(): bool
    {
        return $this->user()?->isAdmin() ?? false;
    }

    /**
     * Get the validation rules that apply to the request.
     */
    public function rules(): array
    {
        return [
            'name' => ['required', 'string', 'max:255'],
            'email' => ['required', 'email', 'max:255', 'unique:users,email'],
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
            'role' => ['required', new Enum(UserRole::class)],
        ];
    }

    /**
     * Get custom messages for validator errors.
     */
    public function messages(): array
    {
        return [
            'name.required' => 'A name is required',
            'email.required' => 'An email adddess is required',
            'email.unique' => 'This email is already registered',
            'password.confirmed' => 'Password confirmation does not match',
        ];
    }

    /**
     * Get custom attributes for validator errors.
     */
    public function attributes(): array
    {
        return [
            'email' => 'email adddess',
        ];
    }

    /**
     * Prepare the data for validation.
     */
    protected function prepareForValidation(): void
    {
        $this->merge([
            'email' => strtolower($this->email),
        ]);
    }
}
```

```php
<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;
use Illuminate\Validation\Rules\Password;

class UpdateUserRequest extends FormRequest
{
    public function authorize(): bool
    {
        $user = $this->route('user');

        // Users can update their own profile, admins can update anyone
        return $this->user()?->id === $user->id
            || $this->user()?->isAdmin();
    }

    public function rules(): array
    {
        $userId = $this->route('user')->id;

        return [
            'name' => ['sometimes', 'string', 'max:255'],
            'email' => [
                'sometimes',
                'email',
                'max:255',
                Rule::unique('users', 'email')->ignore($userId),
            ],
            'password' => [
                'sometimes',
                'confirmed',
                Password::min(8)->letters()->numbers(),
            ],
        ];
    }
}
```

### API Resources for Response Transformation

```php
<?php

namespace App\Http\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

class UserResource extends JsonResource
{
    /**
     * Transform the resource into an array.
     */
    public function toArray(Request $request): array
    {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'email' => $this->email,
            'role' => [
                'value' => $this->role->value,
                'label' => $this->role->label(),
            ],
            'status' => [
                'value' => $this->status->value,
                'label' => $this->status->label(),
            ],
            'email_verified_at' => $this->email_verified_at?->toIso8601String(),
            'posts_count' => $this->whenCounted('posts'),
            'posts' => PostResource::collection($this->whenLoaded('posts')),
            'created_at' => $this->created_at->toIso8601String(),
            'updated_at' => $this->updated_at->toIso8601String(),
        ];
    }

    /**
     * Get additional data that should be returned with the resource array.
     */
    public function with(Request $request): array
    {
        return [
            'meta' => [
                'version' => '1.0.0',
            ],
        ];
    }
}
```

```php
<?php

namespace App\Http\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\ResourceCollection;

class UserCollection extends ResourceCollection
{
    /**
     * Transform the resource collection into an array.
     */
    public function toArray(Request $request): array
    {
        return [
            'data' => $this->collection,
            'links' => [
                'self' => url()->current(),
            ],
            'meta' => [
                'total' => $this->total(),
                'per_page' => $this->perPage(),
                'current_page' => $this->currentPage(),
                'last_page' => $this->lastPage(),
            ],
        ];
    }
}
```

### API Controllers

```php
<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreUserRequest;
use App\Http\Requests\UpdateUserRequest;
use App\Http\Resources\UserResource;
use App\Http\Resources\UserCollection;
use App\Models\User;
use App\Services\UserService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class UserController extends Controller
{
    public function __construct(
        private readonly UserService $userService,
    ) {}

    /**
     * Display a listing of the resource.
     */
    public function index(Request $request): UserCollection
    {
        $perPage = $request->input('per_page', 15);
        $users = User::with(['posts'])
            ->latest()
            ->paginate($perPage);

        return new UserCollection($users);
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(StoreUserRequest $request): JsonResponse
    {
        $userDTO = $this->userService->register($request->validated());

        return response()->json(
            new UserResource(User::find($userDTO->id)),
            201
        );
    }

    /**
     * Display the specified resource.
     */
    public function show(User $user): UserResource
    {
        $user->loadMissing(['posts', 'comments']);
        return new UserResource($user);
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(UpdateUserRequest $request, User $user): UserResource
    {
        $userDTO = $this->userService->updateProfile($user, $request->validated());
        return new UserResource(User::find($userDTO->id));
    }

    /**
     * Remove the specified resource from storage.
     */
    public function destroy(User $user): JsonResponse
    {
        $user->delete();

        return response()->json(null, 204);
    }

    /**
     * Get current authenticated user.
     */
    public function me(Request $request): UserResource
    {
        return new UserResource($request->user());
    }

    /**
     * Get user statistics.
     */
    public function statistics(User $user): JsonResponse
    {
        $stats = $this->userService->getUserStatistics($user);

        return response()->json([
            'data' => $stats,
        ]);
    }
}
```

---

## Queue Jobs and Events

### Job Classes

```php
<?php

namespace App\Jobs;

use App\Models\User;
use App\Notifications\WelcomeNotification;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class SendWelcomeEmail implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    /**
     * The number of times the job may be attempted.
     */
    public int $tries = 3;

    /**
     * The maximum number of unhandled exceptions to allow before failing.
     */
    public int $maxExceptions = 3;

    /**
     * The number of seconds the job can run before timing out.
     */
    public int $timeout = 60;

    /**
     * Create a new job instance.
     */
    public function __construct(
        public User $user,
    ) {}

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        $this->user->notify(new WelcomeNotification());
    }

    /**
     * Handle a job failure.
     */
    public function failed(\Throwable $exception): void
    {
        // Log failure or send notification
        logger()->error('Welcome email failed', [
            'user_id' => $this->user->id,
            'exception' => $exception->getMessage(),
        ]);
    }
}
```

### Event Classes

```php
<?php

namespace App\Events;

use App\Models\User;
use Illuminate\Broadcasting\Channel;
use Illuminate\Broadcasting\InteractsWithSockets;
use Illuminate\Broadcasting\PresenceChannel;
use Illuminate\Contracts\Broadcasting\ShouldBroadcast;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class UserRegistered implements ShouldBroadcast
{
    use Dispatchable, InteractsWithSockets, SerializesModels;

    /**
     * Create a new event instance.
     */
    public function __construct(
        public User $user,
    ) {}

    /**
     * Get the channels the event should broadcast on.
     */
    public function broadcastOn(): array
    {
        return [
            new PresenceChannel('admin'),
            new Channel('users'),
        ];
    }

    /**
     * The event's broadcast name.
     */
    public function broadcastAs(): string
    {
        return 'user.registered';
    }

    /**
     * Get the data to broadcast.
     */
    public function broadcastWith(): array
    {
        return [
            'id' => $this->user->id,
            'name' => $this->user->name,
            'email' => $this->user->email,
            'registered_at' => now()->toIso8601String(),
        ];
    }
}
```

### Event Listeners

```php
<?php

namespace App\Listeners;

use App\Events\UserRegistered;
use App\Jobs\SendWelcomeEmail;

class SendWelcomeNotification
{
    /**
     * Handle the event.
     */
    public function handle(UserRegistered $event): void
    {
        // Dispatch job to queue
        SendWelcomeEmail::dispatch($event->user)
            ->onQueue('emails')
            ->delay(now()->addMinutes(5));
    }
}
```

---

## Testing with PHPUnit and Pest

### PHPUnit Feature Tests

```php
<?php

namespace Tests\Feature;

use App\Models\User;
use App\Enums\UserRole;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class UserApiTest extends TestCase
{
    use RefreshDatabase;

    /**
     * Test user registration.
     */
    public function test_can_register_user(): void
    {
        $response = $this->postJson('/api/users', [
            'name' => 'John Doe',
            'email' => 'john@example.com',
            'password' => 'Password123!',
            'password_confirmation' => 'Password123!',
            'role' => UserRole::Subscriber->value,
        ]);

        $response->assertStatus(201)
            ->assertJsonStructure([
                'data' => [
                    'id',
                    'name',
                    'email',
                    'role',
                    'created_at',
                ],
            ]);

        $this->assertDatabaseHas('users', [
            'email' => 'john@example.com',
        ]);
    }

    /**
     * Test duplicate email validation.
     */
    public function test_cannot_register_with_duplicate_email(): void
    {
        User::factory()->create(['email' => 'john@example.com']);

        $response = $this->postJson('/api/users', [
            'name' => 'Jane Doe',
            'email' => 'john@example.com',
            'password' => 'Password123!',
            'password_confirmation' => 'Password123!',
            'role' => UserRole::Subscriber->value,
        ]);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['email']);
    }

    /**
     * Test authentication required.
     */
    public function test_authentication_required_for_user_list(): void
    {
        $response = $this->getJson('/api/users');

        $response->assertStatus(401);
    }

    /**
     * Test authenticated user can fetch users.
     */
    public function test_authenticated_user_can_fetch_users(): void
    {
        $user = User::factory()->create();
        User::factory()->count(5)->create();

        $response = $this->actingAs($user)
            ->getJson('/api/users');

        $response->assertStatus(200)
            ->assertJsonStructure([
                'data' => [
                    '*' => ['id', 'name', 'email'],
                ],
                'meta' => ['total', 'per_page'],
            ]);
    }

    /**
     * Test user can update own profile.
     */
    public function test_user_can_update_own_profile(): void
    {
        $user = User::factory()->create();

        $response = $this->actingAs($user)
            ->patchJson("/api/users/{$user->id}", [
                'name' => 'Updated Name',
            ]);

        $response->assertStatus(200);

        $this->assertDatabaseHas('users', [
            'id' => $user->id,
            'name' => 'Updated Name',
        ]);
    }

    /**
     * Test user cannot update other users.
     */
    public function test_user_cannot_update_other_users(): void
    {
        $user = User::factory()->create();
        $otherUser = User::factory()->create();

        $response = $this->actingAs($user)
            ->patchJson("/api/users/{$otherUser->id}", [
                'name' => 'Hacked Name',
            ]);

        $response->assertStatus(403);
    }
}
```

### Pest Tests

```php
<?php

use App\Models\User;
use App\Models\Post;
use App\Enums\PostStatus;

it('can create a post', function () {
    $user = User::factory()->create();

    $response = $this->actingAs($user)
        ->postJson('/api/posts', [
            'title' => 'My First Post',
            'content' => 'This is the content of my first post.',
            'status' => PostStatus::Draft->value,
        ]);

    $response->assertStatus(201);

    expect(Post::count())->toBe(1);
    expect(Post::first()->user_id)->toBe($user->id);
});

it('requires authentication to create post', function () {
    $response = $this->postJson('/api/posts', [
        'title' => 'My Post',
        'content' => 'Content here.',
        'status' => PostStatus::Draft->value,
    ]);

    $response->assertStatus(401);
});

it('validates post data', function () {
    $user = User::factory()->create();

    $response = $this->actingAs($user)
        ->postJson('/api/posts', [
            'title' => '', // Empty title
            'content' => 'Content',
        ]);

    $response->assertStatus(422)
        ->assertJsonValidationErrors(['title']);
});

it('can publish a post', function () {
    $user = User::factory()->create();
    $post = Post::factory()->for($user)->create([
        'status' => PostStatus::Draft,
    ]);

    $response = $this->actingAs($user)
        ->patchJson("/api/posts/{$post->id}", [
            'status' => PostStatus::Published->value,
        ]);

    $response->assertStatus(200);

    expect($post->fresh()->status)->toBe(PostStatus::Published);
});

it('filters published posts only', function () {
    $user = User::factory()->create();

    Post::factory()->for($user)->count(3)->create([
        'status' => PostStatus::Published,
        'published_at' => now()->subDay(),
    ]);

    Post::factory()->for($user)->count(2)->create([
        'status' => PostStatus::Draft,
    ]);

    $response = $this->getJson('/api/posts?status=published');

    $response->assertStatus(200)
        ->assertJsonCount(3, 'data');
});

test('post belongs to user', function () {
    $user = User::factory()->create();
    $post = Post::factory()->for($user)->create();

    expect($post->user->id)->toBe($user->id);
    expect($user->posts)->toHaveCount(1);
});
```

### Unit Tests

```php
<?php

namespace Tests\Unit;

use App\DTOs\UserDTO;
use App\Enums\UserRole;
use App\Enums\UserStatus;
use App\Models\User;
use App\Repositories\UserRepository;
use App\Services\UserService;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class UserServiceTest extends TestCase
{
    use RefreshDatabase;

    private UserService $userService;
    private UserRepository $userRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $this->userRepository = new UserRepository();
        $this->userService = new UserService($this->userRepository);
    }

    public function test_can_register_user(): void
    {
        $data = [
            'name' => 'John Doe',
            'email' => 'john@example.com',
            'password' => 'password123',
        ];

        $userDTO = $this->userService->register($data);

        $this->assertInstanceOf(UserDTO::class, $userDTO);
        $this->assertEquals('John Doe', $userDTO->name);
        $this->assertEquals('john@example.com', $userDTO->email);
        $this->assertDatabaseHas('users', ['email' => 'john@example.com']);
    }

    public function test_can_update_user_profile(): void
    {
        $user = User::factory()->create();

        $updateData = [
            'name' => 'Updated Name',
        ];

        $userDTO = $this->userService->updateProfile($user, $updateData);

        $this->assertEquals('Updated Name', $userDTO->name);
        $this->assertDatabaseHas('users', [
            'id' => $user->id,
            'name' => 'Updated Name',
        ]);
    }

    public function test_can_verify_email(): void
    {
        $user = User::factory()->create(['email_verified_at' => null]);

        $this->assertNull($user->email_verified_at);

        $this->userService->verifyEmail($user);

        $this->assertNotNull($user->fresh()->email_verified_at);
    }

    public function test_can_change_user_role(): void
    {
        $user = User::factory()->create(['role' => UserRole::Subscriber]);

        $userDTO = $this->userService->changeRole($user, UserRole::Admin);

        $this->assertEquals(UserRole::Admin, $userDTO->role);
        $this->assertDatabaseHas('users', [
            'id' => $user->id,
            'role' => UserRole::Admin->value,
        ]);
    }
}
```

---

## Artisan Commands

```php
<?php

namespace App\Console\Commands;

use App\Models\User;
use App\Services\UserService;
use Illuminate\Console\Command;

class ProcessInactiveUsersCommand extends Command
{
    /**
     * The name and signature of the console command.
     */
    protected $signature = 'users:process-inactive
                            {--days=30 : Number of days of inactivity}
                            {--notify : Send notification to users}
                            {--delete : Delete inactive users}';

    /**
     * The console command description.
     */
    protected $description = 'Process inactive users';

    /**
     * Execute the console command.
     */
    public function handle(UserService $userService): int
    {
        $days = $this->option('days');
        $notify = $this->option('notify');
        $delete = $this->option('delete');

        $this->info("Processing users inactive for {$days} days...");

        $inactiveUsers = User::where('last_login_at', '<', now()->subDays($days))
            ->whereNull('deleted_at')
            ->get();

        $this->info("Found {$inactiveUsers->count()} inactive users.");

        $bar = $this->output->createProgressBar($inactiveUsers->count());
        $bar->start();

        foreach ($inactiveUsers as $user) {
            if ($notify) {
                // Send notification
                $user->notify(new InactiveUserNotification());
            }

            if ($delete) {
                $user->delete();
            }

            $bar->advance();
        }

        $bar->finish();
        $this->newLine();

        $this->info('Processing complete!');

        return Command::SUCCESS;
    }
}
```

---

## Database Migrations

```php
<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('users', function (Blueprint $table) {
            $table->id();
            $table->uuid('uuid')->unique();
            $table->string('name');
            $table->string('email')->unique();
            $table->timestamp('email_verified_at')->nullable();
            $table->string('password');
            $table->string('role')->default('subscriber');
            $table->string('status')->default('active');
            $table->rememberToken();
            $table->timestamps();
            $table->softDeletes();

            $table->index(['email', 'status']);
            $table->index('created_at');
        });

        Schema::create('posts', function (Blueprint $table) {
            $table->id();
            $table->uuid('uuid')->unique();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('title');
            $table->string('slug')->unique();
            $table->text('excerpt')->nullable();
            $table->longText('content');
            $table->string('status')->default('draft');
            $table->timestamp('published_at')->nullable();
            $table->timestamps();
            $table->softDeletes();

            $table->index(['user_id', 'status']);
            $table->index('published_at');
            $table->fullText(['title', 'content']);
        });

        Schema::create('comments', function (Blueprint $table) {
            $table->id();
            $table->foreignId('post_id')->constrained()->cascadeOnDelete();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->text('content');
            $table->boolean('is_approved')->default(false);
            $table->timestamps();
            $table->softDeletes();

            $table->index(['post_id', 'is_approved']);
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('comments');
        Schema::dropIfExists('posts');
        Schema::dropIfExists('users');
    }
};
```

---

## Docker Production Setup

```dockerfile
# Dockerfile
FROM php:8.3-fpm-alpine AS base

# Install system dependencies
RUN apk add --no-cache \
    nginx \
    supervisor \
    postgresql-dev \
    zip \
    unzip \
    git \
    curl

# Install PHP extensions
RUN docker-php-ext-install \
    pdo \
    pdo_pgsql \
    pgsql \
    opcache \
    pcntl

# Install Composer
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

# Copy composer files
COPY composer.json composer.lock ./

# Install dependencies
RUN composer install --no-dev --optimize-autoloader --no-scripts

# Copy application code
COPY . .

# Set permissions
RUN chown -R www-data:www-data /var/www/html \
    && chmod -R 755 /var/www/html/storage

# Production optimizations
RUN php artisan config:cache \
    && php artisan route:cache \
    && php artisan view:cache

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD php artisan schedule:test || exit 1

CMD ["php-fpm"]
```

---

Last Updated: 2026-01-10
Version: 1.0.0
