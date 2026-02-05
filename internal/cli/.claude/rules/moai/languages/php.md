---
paths:
  - "**/*.php"
  - "**/composer.json"
  - "**/composer.lock"
---

# PHP Rules

Version: PHP 8.3+

## Tooling

- Package management: Composer
- Linting: PHP_CodeSniffer, PHPStan level 9
- Formatting: PHP-CS-Fixer
- Testing: PHPUnit >= 85% coverage

## MUST

- Use strict types (declare(strict_types=1))
- Use typed properties and return types
- Use constructor property promotion
- Use named arguments for clarity
- Use readonly properties for immutability
- Handle exceptions with proper types

## MUST NOT

- Use @ error suppression operator
- Use global variables
- Mix HTML and PHP logic directly
- Use deprecated functions
- Ignore PHPStan errors
- Store credentials in code

## File Conventions

- *Test.php for test files
- PSR-4 autoloading structure
- Use PascalCase for classes
- Use camelCase for methods
- One class per file

## Testing

- Use PHPUnit with data providers
- Use Mockery or PHPUnit mocks
- Use Pest for expressive tests
- Use database transactions for isolation

## MoAI Integration

- Use Skill("moai-lang-php") for detailed patterns
- Follow TRUST 5 quality gates
