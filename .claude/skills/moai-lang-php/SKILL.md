---
name: moai-lang-php
description: PHP best practices with PHPUnit, Composer, and PSR standards
allowed-tools:
- Read
- Bash
---

# PHP Expert

## What it does

Provides PHP-specific expertise for TDD development, including PHPUnit testing, Composer package management, and PSR (PHP Standards Recommendations) compliance.

## When to use

- "PHP 테스트 작성", "PHPUnit 사용법", "PSR 표준"
- Automatically invoked when working with PHP projects
- PHP SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **PHPUnit**: PHP testing framework
- **Mockery**: Mocking library
- **PHPSpec**: BDD-style testing (alternative)
- Test coverage with `phpunit --coverage-html`

**Code Quality**:
- **PHP_CodeSniffer**: PSR compliance checker
- **PHPStan**: Static analysis tool
- **PHP CS Fixer**: Code formatting

**Package Management**:
- **Composer**: Dependency management
- **composer.json**: Package configuration
- **Packagist**: Public package registry

**PSR Standards**:
- **PSR-1**: Basic coding standard
- **PSR-2/PSR-12**: Coding style guide
- **PSR-4**: Autoloading standard
- **PSR-7**: HTTP message interfaces

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Type declarations (PHP 7.4+)
- Namespaces for organization
- Dependency injection over global state

## Examples

### Example 1: TDD with PHPUnit
User: "/alfred:2-build SERVICE-001"
Claude: (creates RED test with PHPUnit, GREEN implementation, REFACTOR with types)

### Example 2: PSR compliance
User: "PSR 표준 확인"
Claude: (runs phpcs --standard=PSR12 and reports violations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (PHP-specific review)
- web-api-expert (Laravel/Symfony API development)
