# moai-lang-php - CLI Reference

_Last updated: 2025-10-22_

## Official Documentation Links

- **PHP**: https://www.php.net/docs.php
- **PHPUnit**: https://docs.phpunit.de/en/11.0/
- **PHP_CodeSniffer**: https://github.com/squizlabs/PHP_CodeSniffer
- **PHPStan**: https://phpstan.org/user-guide/getting-started
- **Composer**: https://getcomposer.org/doc/
- **PSR Standards**: https://www.php-fig.org/psr/

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Status |
|------|---------|--------------|--------|
| **PHP** | 8.3.15 | 2024-12 | ✅ Latest |
| **PHPUnit** | 11.5.3 | 2024-12 | ✅ Latest |
| **PHP_CodeSniffer** | 3.10.3 | 2024-10 | ✅ Stable |
| **PHPStan** | 2.0.3 | 2024-12 | ✅ Latest |
| **Composer** | 2.8.4 | 2024-11 | ✅ Latest |

## Installation

### Prerequisites

```bash
# Check PHP installation
php --version  # Should be 8.3.x or higher

# Check Composer installation
composer --version  # Should be 2.8.x or higher
```

### Project Initialization

```bash
# Create new project
mkdir my-project && cd my-project
composer init

# Install testing dependencies
composer require --save-dev phpunit/phpunit:^11.0

# Install code quality tools
composer require --save-dev squizlabs/php_codesniffer:^3.10
composer require --save-dev phpstan/phpstan:^2.0
composer require --save-dev phpstan/phpstan-strict-rules:^2.0
```

## Common Commands

### Testing with PHPUnit

```bash
# Run all tests
composer test

# Run tests with coverage
vendor/bin/phpunit --coverage-html coverage

# Run specific test file
vendor/bin/phpunit tests/CalculatorTest.php

# Run tests matching pattern
vendor/bin/phpunit --filter=testAdd

# Run specific test suite
vendor/bin/phpunit --testsuite=Unit

# Run with debug output
vendor/bin/phpunit --debug

# Run in CI mode
vendor/bin/phpunit --coverage-text --colors=never
```

### Linting with PHP_CodeSniffer

```bash
# Check PSR-12 compliance
vendor/bin/phpcs --standard=PSR12 src tests

# Auto-fix issues
vendor/bin/phpcbf --standard=PSR12 src tests

# Check specific file
vendor/bin/phpcs src/Calculator.php

# Custom ruleset
vendor/bin/phpcs --standard=phpcs.xml src

# Generate report
vendor/bin/phpcs --report=summary src
vendor/bin/phpcs --report=json src > report.json
```

### Static Analysis with PHPStan

```bash
# Analyze with level 8
vendor/bin/phpstan analyse src tests --level=8

# Analyze with configuration
vendor/bin/phpstan analyse -c phpstan.neon

# Generate baseline
vendor/bin/phpstan analyse --generate-baseline

# Clear result cache
vendor/bin/phpstan clear-result-cache

# Memory limit for large projects
vendor/bin/phpstan analyse --memory-limit=1G
```

### Dependency Management

```bash
# Install all dependencies
composer install

# Install production dependencies only
composer install --no-dev

# Update dependencies
composer update

# Update specific package
composer update vendor/package

# Check for outdated packages
composer outdated

# Validate composer.json
composer validate

# Show dependency tree
composer show --tree

# Dump autoload
composer dump-autoload -o
```

## Configuration Files

### composer.json Scripts

```json
{
  "scripts": {
    "test": "phpunit",
    "test:unit": "phpunit --testsuite=Unit",
    "test:integration": "phpunit --testsuite=Integration",
    "test:coverage": "phpunit --coverage-html coverage",
    "lint": "phpcs --standard=PSR12 src tests",
    "lint:fix": "phpcbf --standard=PSR12 src tests",
    "analyze": "phpstan analyse src tests --level=8",
    "analyze:baseline": "phpstan analyse --generate-baseline",
    "quality": [
      "@lint",
      "@analyze",
      "@test"
    ]
  }
}
```

### phpcs.xml

```xml
<?xml version="1.0"?>
<ruleset name="Project Coding Standard">
    <description>PSR-12 with custom rules</description>

    <file>src</file>
    <file>tests</file>

    <arg name="colors"/>
    <arg value="sp"/>

    <rule ref="PSR12"/>

    <rule ref="Generic.Files.LineLength">
        <properties>
            <property name="lineLimit" value="120"/>
            <property name="absoluteLineLimit" value="150"/>
        </properties>
    </rule>

    <exclude-pattern>*/vendor/*</exclude-pattern>
    <exclude-pattern>*/cache/*</exclude-pattern>
</ruleset>
```

## Best Practices

### Code Organization
✅ Use strict types: `declare(strict_types=1);`
✅ Follow PSR-4 autoloading standard
✅ Use readonly properties (PHP 8.1+)
✅ Prefer final classes
✅ Keep methods small (<50 lines)

### Testing
✅ Write tests before implementation (TDD)
✅ Use PHPUnit attributes (#[Test], #[DataProvider])
✅ Follow AAA pattern (Arrange, Act, Assert)
✅ Maintain ≥85% code coverage
✅ Use type hints consistently

### Error Handling
✅ Use typed exceptions
✅ Provide meaningful error messages
✅ Validate inputs early
✅ Use strict comparisons (===)
✅ Enable error reporting in development

### Security
✅ Run `composer audit` regularly
✅ Keep dependencies updated
✅ Use prepared statements for SQL
✅ Validate and sanitize user inputs
✅ Use HTTPS for external requests

## Common Issues & Solutions

### Issue: PHPUnit not finding tests
**Solution**: Check PSR-4 autoloading in composer.json, run `composer dump-autoload`

### Issue: PHPCS conflicts with PSR-12
**Solution**: Update to latest PHP_CodeSniffer, check phpcs.xml configuration

### Issue: PHPStan memory exhausted
**Solution**: Increase memory limit: `phpstan analyse --memory-limit=1G`

### Issue: Coverage threshold not met
**Solution**: Check coverage report, add missing tests, configure phpunit.xml

---

_For working examples and use cases, see examples.md_
