---
name: moai-lang-php
description: PHP 8.4+ best practices with PHPUnit 11, Composer, PSR-12 standards, and web frameworks (Laravel, Symfony).
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, moai, php  


## Quick Reference (30 seconds)

# PHP Web Development — Enterprise

**Primary Focus**: PHP 8.4+ with Laravel 12, Symfony 7, async patterns, and production deployment
**Best For**: REST APIs, web applications, enterprise systems, microservices
**Key Libraries**: Laravel 12, Symfony 7, PHPUnit 11, Composer 2.8, PHPStan 2.0
**Auto-triggers**: PHP, Laravel, Symfony, Composer, PHPUnit, web frameworks

| Version | Release | Support |
|---------|---------|---------|
| PHP 8.4.0 | Nov 2024 | Nov 2027 |
| Laravel 12.0 | 2025-02 | 2027-08 |
| Symfony 7.3 | 2025-09 | 2027-09 |
| PHPUnit 11.5 | 2024-11 | Active |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core PHP 8.4 concepts with practical examples:
- **PHP 8.4 Core**: Fibers, Attributes, JIT compiler, nullsafe operator
- **Strong Type System**: Type declarations, union types, enums
- **Laravel Basics**: Controllers, models, routing, Eloquent ORM
- **Symfony Setup**: Service container, routing, dependency injection
- **Examples**: See `examples.md` for full code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready enterprise patterns:
- **OOP Patterns**: Dependency injection, repository pattern, service container
- **Static Analysis**: PHPStan, Psalm for type safety
- **Performance**: JIT compilation, Opcache, lazy loading
- **Testing**: PHPUnit fixtures, data providers, mocking
- **Pattern Reference**: See `reference.md` for API details and best practices

### Level 3: Production Deployment (Consult backend/performance skills)

Enterprise deployment and optimization:
- **Docker & Compose**: Container packaging with multi-stage builds
- **Performance**: Query optimization, caching strategies
- **Monitoring**: Error tracking, metrics, observability
- **Scaling**: Worker management, connection pooling
- **Details**: Skill("moai-domain-backend"), Skill("moai-essentials-perf")

---

## Learn More

- **Examples**: See `examples.md` for Laravel, Symfony, PHPUnit, and async patterns
- **Reference**: See `reference.md` for API details, configuration, and troubleshooting
- **Advanced Patterns**: See `modules/advanced-patterns.md` for metaprogramming and concurrency
- **Optimization**: See `modules/optimization.md` for performance tuning and profiling
- **PHP 8.4**: https://www.php.net/releases/8.4/
- **Laravel Docs**: https://laravel.com/docs
- **Symfony Docs**: https://symfony.com/doc/

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-essentials-perf"), Skill("moai-domain-backend")
**Auto-loads**: PHP projects mentioning Laravel, Symfony, PHPUnit, Composer, web frameworks

---

## Installation Commands

```bash
# Core PHP development stack
composer require laravel/framework:^12.0
composer require symfony/framework-bundle:^7.0
composer require phpunit/phpunit:^11.5
composer require phpstan/phpstan:^2.0

# Development tools
composer require --dev symfony/var-dumper
composer require --dev mockery/mockery

# Production addons
composer require redis
composer require monolog/monolog
composer require guzzlehttp/guzzle

# Async support
composer require amphp/amp
composer require react/event-loop
```

---

## Best Practices

1. **Always use strict types**: `declare(strict_types=1)` in all files
2. **Leverage type system**: Use union types, enums, named arguments
3. **Apply DI pattern**: Use constructor injection consistently
4. **Test-first approach**: Write tests before implementation
5. **Use static analysis**: Run PHPStan at max level
6. **Optimize with JIT**: Enable JIT compiler in production
7. **Monitor performance**: Profile with Xdebug, optimize hot paths
8. **Security-first**: Input validation, prepared statements, secure headers

---

## What It Does

PHP 8.4+ with Laravel, Symfony, modern async support, and production-grade best practices.

**Key capabilities**:
- Modern syntax (Fibers, Attributes, match expressions, property hooks)
- Strong typing with Generics (via PHPStan/Psalm)
- Framework expertise (Laravel Eloquent, Symfony DI container)
- Testing framework integration (PHPUnit with data providers)
- Static analysis at max level

---

## Works Well With

- `moai-domain-backend` (backend architecture patterns)
- `moai-essentials-perf` (performance profiling with Xdebug)
- `moai-essentials-debug` (debugging support)
- `moai-foundation-trust` (quality gates)

---

## Changelog

- **v3.1.0** (2025-11-22): Modularized structure with advanced patterns and optimization modules
- **v3.0.0** (2025-11-21): PHP 8.4 features, strong typing, framework expansion
- **v2.0.0** (2025-10-15): Latest tool versions, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial release

---

## Advanced Patterns

See `modules/advanced-patterns.md` for:
- Fibers and cooperative multitasking
- Attributes and reflection
- Advanced OOP patterns (Factory, Strategy, Observer)
- Generics and type-safe collections
- Metaprogramming with magic methods

See `modules/optimization.md` for:
- JIT compiler configuration and tuning
- Opcache preloading strategies
- Lazy loading with WeakMap
- Profiling with Xdebug 3
- Database query optimization

---

## Context7 Integration

### Related Libraries & Tools
- [Laravel](/laravel/laravel): Modern PHP framework for rapid development
- [Symfony](/symfony/symfony): Enterprise PHP framework with components
- [Composer](/composer/composer): Dependency manager for PHP
- [PHPUnit](/sebastianbergmann/phpunit): Testing framework for PHP
- [PHPStan](/phpstan/phpstan): Static analysis tool for PHP

### Official Documentation
- [PHP 8.4](https://www.php.net/releases/8.4/)
- [Laravel Documentation](https://laravel.com/docs)
- [Symfony Documentation](https://symfony.com/doc/)
- [PHPUnit Guide](https://phpunit.de/documentation.html)
- [Composer Guide](https://getcomposer.org/doc/)

### Version-Specific Guides
Latest stable version: PHP 8.4.0, Laravel 12.0, Symfony 7.3
- [PHP 8.4 Release Notes](https://www.php.net/releases/8.4/)
- [Laravel 12 Upgrade](https://laravel.com/docs/12.x/upgrade)
- [Symfony 7 Migration](https://symfony.com/doc/7.0/setup.html)
- [PHPUnit 11 Guide](https://phpunit.de/getting-started-with-phpunit.html)
