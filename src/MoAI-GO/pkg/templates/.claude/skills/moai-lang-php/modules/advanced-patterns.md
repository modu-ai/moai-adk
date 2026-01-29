# PHP Advanced Patterns

## Advanced Eloquent Patterns (Laravel)

Observers:
```php
<?php

namespace App\Observers;

use App\Models\User;

class UserObserver
{
    public function creating(User $user): void
    {
        $user->uuid = Str::uuid();
    }

    public function created(User $user): void
    {
        event(new UserCreated($user));
    }

    public function updating(User $user): void
    {
        $user->updated_by = auth()->id();
    }

    public function deleted(User $user): void
    {
        $user->posts()->delete();
    }
}
```

Accessors and Mutators:
```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Casts\Attribute;
use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    protected function fullName(): Attribute
    {
        return Attribute::make(
            get: fn () => "{$this->first_name} {$this->last_name}",
        );
    }

    protected function password(): Attribute
    {
        return Attribute::make(
            set: fn (string $value) => bcrypt($value),
        );
    }

    protected function settings(): Attribute
    {
        return Attribute::make(
            get: fn ($value) => json_decode($value, true),
            set: fn ($value) => json_encode($value),
        );
    }
}
```

Query Scopes:
```php
<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;

class Post extends Model
{
    public function scopePublished(Builder $query): Builder
    {
        return $query->where('status', 'published')
            ->whereNotNull('published_at');
    }

    public function scopeByAuthor(Builder $query, User $author): Builder
    {
        return $query->where('user_id', $author->id);
    }

    public function scopePopular(Builder $query, int $minViews = 1000): Builder
    {
        return $query->where('views', '>=', $minViews);
    }
}

// Usage
Post::published()->byAuthor($user)->popular()->get();
```

## Doctrine Advanced Mapping (Symfony)

Embeddables:
```php
<?php

namespace App\Entity;

use Doctrine\ORM\Mapping as ORM;

#[ORM\Embeddable]
class Adddess
{
    #[ORM\Column(length: 255)]
    private string $street;

    #[ORM\Column(length: 100)]
    private string $city;

    #[ORM\Column(length: 20)]
    private string $postalCode;

    #[ORM\Column(length: 2)]
    private string $country;

    public function __construct(string $street, string $city, string $postalCode, string $country)
    {
        $this->street = $street;
        $this->city = $city;
        $this->postalCode = $postalCode;
        $this->country = $country;
    }

    public function getFullAdddess(): string
    {
        return "{$this->street}, {$this->city} {$this->postalCode}, {$this->country}";
    }
}

#[ORM\Entity]
class Company
{
    #[ORM\Embedded(class: Adddess::class)]
    private Adddess $adddess;
}
```

Inheritance Mapping:
```php
<?php

namespace App\Entity;

use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\InheritanceType('SINGLE_TABLE')]
#[ORM\DiscriminatorColumn(name: 'type', type: 'string')]
#[ORM\DiscriminatorMap(['user' => User::class, 'admin' => Admin::class])]
abstract class Person
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    protected ?int $id = null;

    #[ORM\Column(length: 255)]
    protected string $name;
}

#[ORM\Entity]
class Admin extends Person
{
    #[ORM\Column(type: 'json')]
    private array $permissions = [];
}
```

## Queue and Job Processing

Laravel Queue Jobs:
```php
<?php

namespace App\Jobs;

use App\Models\User;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class ProcessUserData implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public int $tries = 3;
    public int $backoff = 60;
    public int $timeout = 120;

    public function __construct(
        public User $user,
        public array $options = []
    ) {}

    public function handle(): void
    {
        // Process user data
        $this->user->processData($this->options);
    }

    public function failed(\Throwable $exception): void
    {
        Log::error('Job failed', [
            'user_id' => $this->user->id,
            'error' => $exception->getMessage(),
        ]);
    }

    public function retryUntil(): \DateTime
    {
        return now()->addHours(24);
    }
}

// Dispatch
ProcessUserData::dispatch($user)->onQueue('high');
ProcessUserData::dispatch($user)->delay(now()->addMinutes(10));
```

Job Batching:
```php
<?php

use Illuminate\Bus\Batch;
use Illuminate\Support\Facades\Bus;

$batch = Bus::batch([
    new ProcessPodcast($podcast1),
    new ProcessPodcast($podcast2),
    new ProcessPodcast($podcast3),
])->then(function (Batch $batch) {
    // All jobs completed successfully
})->catch(function (Batch $batch, Throwable $e) {
    // First batch job failure detected
})->finally(function (Batch $batch) {
    // Batch has finished executing
})->dispatch();
```

## Event-Driven Architecture

Laravel Events:
```php
<?php

namespace App\Events;

use App\Models\Order;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class OrderPlaced
{
    use Dispatchable, SerializesModels;

    public function __construct(
        public Order $order
    ) {}
}

// Listener
namespace App\Listeners;

use App\Events\OrderPlaced;
use Illuminate\Contracts\Queue\ShouldQueue;

class SendOrderConfirmation implements ShouldQueue
{
    public function handle(OrderPlaced $event): void
    {
        Mail::to($event->order->user)->send(new OrderConfirmationMail($event->order));
    }

    public function shouldQueue(OrderPlaced $event): bool
    {
        return $event->order->total > 100;
    }
}
```

Symfony Event Subscribers:
```php
<?php

namespace App\EventSubscriber;

use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpKernel\Event\RequestEvent;
use Symfony\Component\HttpKernel\KernelEvents;

class RequestSubscriber implements EventSubscriberInterface
{
    public static function getSubscribedEvents(): array
    {
        return [
            KernelEvents::REQUEST => [
                ['onKernelRequest', 10],
            ],
        ];
    }

    public function onKernelRequest(RequestEvent $event): void
    {
        if (!$event->isMainRequest()) {
            return;
        }

        // Handle request
    }
}
```

## Caching Strategies

Redis Caching (Laravel):
```php
<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Redis;

class ProductService
{
    public function getProduct(int $id): ?Product
    {
        return Cache::tags(['products'])->remember(
            "product:{$id}",
            now()->addHours(24),
            fn () => Product::find($id)
        );
    }

    public function invalidateProduct(int $id): void
    {
        Cache::tags(['products'])->forget("product:{$id}");
    }

    public function clearAllProducts(): void
    {
        Cache::tags(['products'])->flush();
    }

    public function getWithRedisLock(int $id): ?Product
    {
        $lock = Cache::lock("product-lock:{$id}", 10);

        try {
            $lock->block(5);
            return $this->getProduct($id);
        } finally {
            $lock->release();
        }
    }
}
```

## Security Best Practices

OWASP Patterns:
```php
<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;

class SecurityHeaders
{
    public function handle(Request $request, Closure $next)
    {
        $response = $next($request);

        $response->headers->set('X-Content-Type-Options', 'nosniff');
        $response->headers->set('X-Frame-Options', 'DENY');
        $response->headers->set('X-XSS-Protection', '1; mode=block');
        $response->headers->set('Referrer-Policy', 'strict-origin-when-cross-origin');
        $response->headers->set(
            'Content-Security-Policy',
            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
        );

        return $response;
    }
}
```

Rate Limiting:
```php
<?php

namespace App\Providers;

use Illuminate\Cache\RateLimiting\Limit;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    public function boot(): void
    {
        RateLimiter::for('api', function ($request) {
            return Limit::perMinute(60)->by($request->user()?->id ?: $request->ip());
        });

        RateLimiter::for('uploads', function ($request) {
            return $request->user()->isPremium()
                ? Limit::none()
                : Limit::perMinute(10)->by($request->user()->id);
        });
    }
}
```

## CI/CD Integration

GitHub Actions for PHP:
```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: password
          MYSQL_DATABASE: testing
        ports:
          - 3306:3306

    steps:
      - uses: actions/checkout@v4

      - name: Setup PHP
        uses: shivammathur/setup-php@v2
        with:
          php-version: '8.3'
          extensions: mbstring, pdo, pdo_mysql

      - name: Install dependencies
        run: composer install --prefer-dist --no-progress

      - name: Run tests
        env:
          DB_CONNECTION: mysql
          DB_HOST: 127.0.0.1
          DB_DATABASE: testing
          DB_USERNAME: root
          DB_PASSWORD: password
        run: php artisan test

      - name: Run PHPStan
        run: vendor/bin/phpstan analyse

      - name: Run Pint
        run: vendor/bin/pint --test
```

## Docker Configuration

Production Dockerfile:
```dockerfile
FROM php:8.3-fpm-alpine

RUN apk add --no-cache \
    libpng-dev \
    libzip-dev \
    && docker-php-ext-install pdo pdo_mysql gd zip opcache

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /var/www

COPY composer.json composer.lock ./
RUN composer install --no-dev --optimize-autoloader --no-scripts

COPY . .
RUN composer dump-autoload --optimize

RUN chown -R www-data:www-data storage bootstrap/cache

EXPOSE 9000

CMD ["php-fpm"]
```

Docker Compose:
```yaml
version: '3.8'

services:
  app:
    build: .
    volumes:
      - .:/var/www
    depends_on:
      - db
      - redis

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - .:/var/www
      - ./nginx.conf:/etc/nginx/conf.d/default.conf

  db:
    image: mysql:8.0
    environment:
      MYSQL_DATABASE: laravel
      MYSQL_ROOT_PASSWORD: secret

  redis:
    image: redis:alpine
```
