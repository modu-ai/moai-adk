---
paths:
  - "**/*.php"
  - "**/composer.json"
  - "**/composer.lock"
---

# PHP Rules

Version: PHP 8.3+ (8.4 recommended)

## Tooling

- Linting: PHP CS Fixer, PHPStan (level 9)
- Testing: PHPUnit with coverage >= 85%
- Package management: Composer

## Best Practices (2026)

- Use Property Hooks (PHP 8.4) for getter/setter logic
- Use Eager Loading to prevent N+1 query problems
- Use Redis for session and cache storage
- Use typed properties and return types everywhere
- Follow PSR-12 coding standard

## PHP 8.4 Features

```php
// Property Hooks
class User {
    public string $fullName {
        get => $this->firstName . ' ' . $this->lastName;
        set => [$this->firstName, $this->lastName] = explode(' ', $value);
    }
}

// Asymmetric visibility
class Config {
    public private(set) string $value;
}
```

## Laravel 11 Patterns

- Use Eloquent eager loading: `with(['relation'])`
- Use query scopes for reusable filters
- Use Form Requests for validation
- Use Resources for API responses

## Performance

- Use OPcache in production
- Use Redis for caching and sessions
- Use database indexing on frequently queried columns
- Use Laravel Octane for long-running processes

## MoAI Integration

- Use Skill("moai-lang-php") for detailed patterns
- Follow TRUST 5 quality gates
