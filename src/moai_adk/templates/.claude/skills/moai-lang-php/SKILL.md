---
name: moai-lang-php
description: >
  PHP 8.3+ development specialist covering Laravel 11, Symfony 7, Eloquent
  ORM, and modern PHP patterns. Use when developing PHP APIs, web
  applications, or Laravel/Symfony projects.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(php:*) Bash(composer:*) Bash(phpunit:*) Bash(phpstan:*) Bash(phpcs:*) Bash(artisan:*) Bash(laravel:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "language, php, laravel, symfony, eloquent, doctrine, phpunit"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["PHP", "Laravel", "Symfony", "Eloquent", "Doctrine", "PHPUnit", "Pest", ".php", "composer.json", "artisan"]
  languages: ["php"]
---

## Quick Reference (30 seconds)

PHP 8.3+ Development Specialist - Laravel 11, Symfony 7, Eloquent, Doctrine, and modern PHP patterns.

Auto-Triggers: PHP files with .php extension, composer.json, artisan command, symfony.yaml, Laravel or Symfony discussions

Core Capabilities:

- PHP 8.3 Features: readonly classes, typed properties, attributes, enums, named arguments
- Laravel 11: Controllers, Models, Migrations, Form Requests, API Resources, Eloquent
- Symfony 7: Attribute-based routing, Doctrine ORM, Services, Dependency Injection
- ORMs: Eloquent for Laravel, Doctrine for Symfony
- Testing: PHPUnit, Pest, feature and unit testing patterns
- Package Management: Composer with autoloading
- Coding Standards: PSR-12, Laravel Pint, PHP CS Fixer
- Docker: PHP-FPM, nginx, multi-stage builds

### Quick Patterns

Laravel Controller Pattern:

In the App\Http\Controllers\Api namespace, create a UserController extending Controller. Import StoreUserRequest, UserResource, User, and JsonResponse. Define a store method accepting StoreUserRequest that creates a User using validated data and returns a JsonResponse with UserResource wrapping the user and status 201.

Laravel Form Request Pattern:

In the App\Http\Requests namespace, create a StoreUserRequest extending FormRequest. The authorize method returns true. The rules method returns an array with name requiring required, string, and max 255 validation, email requiring required, email, and unique on users table, and password requiring required, min 8, and confirmed validation.

Symfony Controller Pattern:

In the App\Controller namespace, create a UserController extending AbstractController. Import User, EntityManagerInterface, JsonResponse, and Route attribute. Apply Route attribute at class level for api/users path. Create a create method with Route attribute for empty path and POST method. Inject EntityManagerInterface, create new User, persist and flush, then return json response with user and status 201.

---

## Implementation Guide (5 minutes)

### PHP 8.3 Modern Features

Readonly Classes:

Declare a readonly class UserDTO with a constructor promoting public properties for int id, string name, and string email.

Enums with Methods:

Create an OrderStatus enum backed by string. Define cases Pending with pending value, Processing with processing value, and Completed with completed value. Add a label method that uses match expression on $this to return appropriate display labels for each case.

Attributes:

Create a Validate attribute class targeting properties with Attribute attribute. The constructor accepts a string rule and optional string message. Create a UserRequest class with email property decorated with Validate attribute specifying required and email rules.

### Laravel 11 Patterns

Eloquent Model with Relationships:

In the App\Models namespace, create a Post model extending Model. Set protected fillable array with title, content, user_id, and status. Set protected casts array with status casting to PostStatus enum and published_at casting to datetime. Define a user method returning BelongsTo relationship. Define a comments method returning HasMany relationship. Add a scopePublished method that filters by published status.

API Resource Pattern:

In the App\Http\Resources namespace, create a PostResource extending JsonResource. The toArray method takes a Request parameter. Return an array with id, title, author using UserResource with whenLoaded for user relationship, comments_count using whenCounted, and created_at formatted as ISO 8601 string.

Migration Pattern:

Create an anonymous migration class extending Migration. The up method calls Schema create on posts table. Define id, foreignId for user_id with constrained and cascadeOnDelete, string for title, text for content, string for status defaulting to draft, timestamps, and softDeletes.

Service Layer Pattern:

In the App\Services namespace, create a UserService class. Define a create method accepting UserDTO. Use DB transaction wrapping User create with data from DTO properties, profile creation with default bio, and returning user with loaded profile relationship. Catch ActiveRecord\RecordInvalid exceptions to handle validation failures.

### Symfony 7 Patterns

Entity with Doctrine Attributes:

In the App\Entity namespace, create a User class. Apply ORM\Entity attribute with repositoryClass pointing to UserRepository. Apply ORM\Table attribute with name users. Add private nullable int id with ORM\Id, ORM\GeneratedValue, and ORM\Column attributes. Add private nullable string name with ORM\Column length 255 and Assert\NotBlank. Add private nullable string email with ORM\Column length 180 unique and Assert\Email.

Service with Dependency Injection:

In the App\Service namespace, create a UserService class. The constructor accepts readonly EntityManagerInterface and readonly UserPasswordHasherInterface via property promotion. Define createUser method taking email and password strings. Create new User, set email, hash password using the password hasher, persist with entity manager, flush, and return user.

### Testing Patterns

PHPUnit Feature Test for Laravel:

In Tests\Feature namespace, create UserApiTest extending TestCase with RefreshDatabase trait. The test_can_create_user method posts JSON to api/users with name, email, password, and password_confirmation. Assert status 201 and JSON structure with data containing id, name, and email. Assert database has users table with the email.

Pest Test for Laravel:

Use App\Models\User and Post. Create a test using it function for can create a post. Create user with factory. Call actingAs with user, post JSON to api/posts with title and content. Assert status 201 and expect Post count to be 1. Create second test for requires authentication that posts without authentication and asserts status 401.

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:

- Production deployment patterns for Docker and Kubernetes
- Advanced Eloquent patterns including observers, accessors, and mutators
- Doctrine advanced mapping with embeddables and inheritance
- Queue and job processing
- Event-driven architecture
- Caching strategies with Redis and Memcached
- Security best practices following OWASP patterns
- CI/CD integration patterns

See:

- modules/advanced-patterns.md for complete advanced patterns guide

---

## Context7 Library Mappings

- laravel/framework for Laravel web framework
- symfony/symfony for Symfony components and framework
- doctrine/orm for Doctrine ORM for PHP
- phpunit/phpunit for PHP testing framework
- pestphp/pest for elegant PHP testing framework
- laravel/sanctum for Laravel API authentication
- laravel/horizon for Laravel queue dashboard

---

## Works Well With

- moai-domain-backend for REST API and microservices architecture
- moai-domain-database for SQL patterns and ORM optimization
- moai-workflow-testing for DDD and testing strategies
- moai-platform-deploy for Docker and deployment patterns
- moai-essentials-debug for AI-powered debugging
- moai-foundation-quality for TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

PHP Version Check:

Run php with version flag to verify 8.3 or later. Use php with -m flag piped to grep for checking pdo, mbstring, and openssl extensions.

Composer Autoload Issues:

Run composer dump-autoload with -o flag for optimized autoloader. Run composer clear-cache to clear the package cache.

Laravel Cache Issues:

Run php artisan config:clear to clear configuration cache. Run php artisan cache:clear to clear application cache. Run php artisan route:clear to clear route cache. Run php artisan view:clear to clear compiled views.

Symfony Cache Issues:

Run php bin/console cache:clear to clear cache. Run php bin/console cache:warmup to warm up the cache.

Database Connection:

Use try-catch block around DB::connection()->getPdo() call. Output success message on connection or exception message on failure.

Migration Rollback:

Use php artisan migrate:rollback with step 1 to rollback last migration. Use php artisan migrate:fresh with seed flag for development reset only. For Symfony, use php bin/console doctrine:migrations:migrate prev to rollback.

---

Version: 1.1.0 | Updated: 2026-01-11 | Status: Active
