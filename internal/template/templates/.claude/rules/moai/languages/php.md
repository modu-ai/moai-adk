---
paths: "**/*.php,**/composer.json,**/composer.lock"
---

# PHP Development Guide

Version: PHP 8.4+

Core Capabilities:

- Linting: PHP CS Fixer, Laravel Pint
- Formatting: PHP CS Fixer, Pint (PSR-12)
- Testing: PHPUnit, Pest with coverage >= 85%
- Package management: Composer with autoloading

### Quick Patterns

File Conventions:

- PascalCase for class files matching class name
- tests/ directory for test files with Test suffix
- src/ or app/ for source code
- composer.json for dependencies

Recommended Practices:

- Use type declarations for all function parameters and return types
- Use readonly classes and properties for immutable data
- Use enum types for fixed value sets
- Handle exceptions with proper context
- Follow PSR-12 coding standard
- Use Composer autoloading

Anti-Patterns to Avoid:

- Use global variables or functions for application state
- Suppress errors with @ operator
- Use eval() or variable variables
- Mix business logic in controllers
- Embed credentials in code

### PHP 8.4 Features

- Readonly classes for immutable DTOs
- Enums with backed values and methods
- Named arguments and match expressions
- Attributes for metadata (Route, Validate, etc.)
- Typed properties and constructor promotion
- Property hooks (PHP 8.4) for computed properties without boilerplate

### Laravel 12 Patterns

Controllers with Form Requests and API Resources. Eloquent models with relationships, scopes, casts. Migration with Schema builder. Service layer with DB transactions. Hotwire/Turbo integration for SPA-like experiences.

### Symfony 7 Patterns

Attribute-based routing and controllers. Doctrine ORM with entity attributes and repository pattern. Service dependency injection. Security with voters and access control.

### Testing

- PHPUnit: Feature and unit tests with RefreshDatabase trait
- Pest: Expressive syntax with it/expect pattern
- Factory patterns with Illuminate\Database\Factories\Factory (Laravel database factories) or Faker for test data generation
- Coverage: phpunit --coverage-html with 85%+ target

---

## Advanced Patterns

For comprehensive coverage including advanced domain modeling, queue/job pipeline patterns (Redis/SQS), Laravel Octane for high-throughput PHP, and production deployment configurations, apply the Quick Patterns and framework sections above together with the Documentation References; this guide is self-contained.

---

## Documentation References

- laravel/framework - Laravel framework
- symfony/symfony - Symfony framework
- doctrine/orm - Doctrine ORM
- phpunit/phpunit - PHPUnit testing
- pestphp/pest - Pest testing framework
- laravel/sanctum - Laravel Sanctum API auth
- laravel/horizon - Laravel Horizon queue dashboard

---

- `moai-domain-backend` - Backend service architecture
- `moai-domain-database` - Database patterns and ORM optimization
- `moai-workflow-testing` - DDD and testing strategies
- `moai-foundation-quality` - TRUST 5 quality principles

---

## Troubleshooting

- Version check: php --version, php -m | grep ext
- Composer: composer dump-autoload -o, composer clear-cache
- Laravel cache: php artisan config:clear, cache:clear, route:clear, view:clear
- Symfony cache: php bin/console cache:clear, cache:warmup
- Migration: php artisan migrate:rollback --step=1

---
