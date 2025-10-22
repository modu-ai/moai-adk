# moai-lang-php - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Composer & PHPUnit

```bash
# Initialize new project
composer init

# Install PHPUnit 11 as dev dependency
composer require --save-dev phpunit/phpunit:^11.0

# Install PHP_CodeSniffer and PHPStan
composer require --save-dev squizlabs/php_codesniffer:^3.10
composer require --save-dev phpstan/phpstan:^2.0

# Create phpunit.xml configuration
vendor/bin/phpunit --generate-configuration

# Add scripts to composer.json
composer config scripts.test "phpunit"
composer config scripts.lint "phpcs --standard=PSR12 src tests"
composer config scripts.analyze "phpstan analyse src tests --level=8"
```

**composer.json configuration**:
```json
{
  "name": "vendor/my-project",
  "require-dev": {
    "phpunit/phpunit": "^11.0",
    "squizlabs/php_codesniffer": "^3.10",
    "phpstan/phpstan": "^2.0"
  },
  "scripts": {
    "test": "phpunit",
    "test:coverage": "phpunit --coverage-html coverage",
    "lint": "phpcs --standard=PSR12 src tests",
    "lint:fix": "phpcbf --standard=PSR12 src tests",
    "analyze": "phpstan analyse src tests --level=8"
  },
  "autoload": {
    "psr-4": {
      "App\\": "src/"
    }
  },
  "autoload-dev": {
    "psr-4": {
      "App\\Tests\\": "tests/"
    }
  }
}
```

## Example 2: TDD Workflow with PHPUnit 11

**RED: Write failing test**
```php
<?php
// tests/CalculatorTest.php
namespace App\Tests;

use App\Calculator;
use PHPUnit\Framework\TestCase;
use PHPUnit\Framework\Attributes\Test;
use PHPUnit\Framework\Attributes\DataProvider;

class CalculatorTest extends TestCase
{
    #[Test]
    public function it_adds_two_positive_numbers(): void
    {
        $calculator = new Calculator();
        $this->assertSame(5, $calculator->add(2, 3));
    }

    #[Test]
    public function it_handles_negative_numbers(): void
    {
        $calculator = new Calculator();
        $this->assertSame(-3, $calculator->add(-1, -2));
    }

    #[Test]
    #[DataProvider('additionProvider')]
    public function it_adds_numbers_correctly(int $a, int $b, int $expected): void
    {
        $calculator = new Calculator();
        $this->assertSame($expected, $calculator->add($a, $b));
    }

    public static function additionProvider(): array
    {
        return [
            [0, 0, 0],
            [0, 5, 5],
            [-1, 1, 0],
            [100, 200, 300],
        ];
    }
}
```

**GREEN: Implement feature**
```php
<?php
// src/Calculator.php
namespace App;

class Calculator
{
    public function add(int $a, int $b): int
    {
        return $a + $b;
    }
}
```

**REFACTOR: Improve code quality**
```php
<?php
// src/Calculator.php
declare(strict_types=1);

namespace App;

/**
 * Calculator class providing basic arithmetic operations
 */
final readonly class Calculator
{
    /**
     * Adds two integers
     *
     * @param int $a First operand
     * @param int $b Second operand
     * @return int Sum of a and b
     */
    public function add(int $a, int $b): int
    {
        return $a + $b;
    }
}
```

## Example 3: PSR-12 Compliant Code with PHP_CodeSniffer

**Run linting**:
```bash
# Check all files
composer lint

# Fix auto-fixable issues
composer lint:fix
```

**Example PSR-12 compliant class**:
```php
<?php

declare(strict_types=1);

namespace App\Service;

use App\Repository\UserRepositoryInterface;
use App\Exception\UserNotFoundException;

final readonly class UserService
{
    public function __construct(
        private UserRepositoryInterface $userRepository
    ) {
    }

    public function findUserById(int $id): array
    {
        $user = $this->userRepository->find($id);

        if ($user === null) {
            throw new UserNotFoundException(
                sprintf('User with ID %d not found', $id)
            );
        }

        return $user;
    }

    public function createUser(string $email, string $name): int
    {
        $this->validateEmail($email);

        return $this->userRepository->create([
            'email' => $email,
            'name' => $name,
            'created_at' => date('Y-m-d H:i:s'),
        ]);
    }

    private function validateEmail(string $email): void
    {
        if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
            throw new \InvalidArgumentException(
                sprintf('Invalid email address: %s', $email)
            );
        }
    }
}
```

## Example 4: PHPStan Static Analysis

**phpstan.neon configuration**:
```neon
parameters:
    level: 8
    paths:
        - src
        - tests
    excludePaths:
        - src/Legacy
    ignoreErrors:
        - '#Call to an undefined method#'
```

**Run analysis**:
```bash
# Analyze with level 8 (strictest)
composer analyze

# Analyze specific directory
vendor/bin/phpstan analyse src/Service --level=8

# Generate baseline for existing errors
vendor/bin/phpstan analyse --generate-baseline
```

## Example 5: PHPUnit Coverage Configuration

**phpunit.xml**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<phpunit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:noNamespaceSchemaLocation="vendor/phpunit/phpunit/phpunit.xsd"
         bootstrap="vendor/autoload.php"
         colors="true"
         failOnRisky="true"
         failOnWarning="true">
    <testsuites>
        <testsuite name="Unit">
            <directory>tests/Unit</directory>
        </testsuite>
        <testsuite name="Integration">
            <directory>tests/Integration</directory>
        </testsuite>
    </testsuites>
    <source>
        <include>
            <directory>src</directory>
        </include>
    </source>
    <coverage>
        <report>
            <html outputDirectory="coverage/html"/>
            <text outputFile="coverage/coverage.txt"/>
        </report>
    </coverage>
</phpunit>
```

**Run tests with coverage**:
```bash
# Run all tests
composer test

# Generate coverage report
composer test:coverage

# Run specific test suite
vendor/bin/phpunit --testsuite=Unit
```

---

_For complete CLI reference and configuration options, see reference.md_
