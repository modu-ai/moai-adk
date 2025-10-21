---

name: moai-lang-php
description: PHP best practices with PHPUnit, Composer, and PSR standards. Use when writing or reviewing PHP code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# PHP Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | PHP code discussions, framework guidance, or file extensions such as .php. |
| Tier | 3 |

## What it does

Provides PHP-specific expertise for TDD development, including PHPUnit testing, Composer package management, and PSR (PHP Standards Recommendations) compliance.

## When to use

- Engages when the conversation references PHP work, frameworks, or files like .php.
- "Writing PHP tests", "How to use PHPUnit", "PSR standard"
- Automatically invoked when working with PHP projects
- PHP SPEC implementation (`/alfred:2-run`)

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
```bash
vendor/bin/phpunit && vendor/bin/phpstan analyse
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- PHP Manual. "PHP Documentation." https://www.php.net/manual/en/ (accessed 2025-03-29).
- PHPUnit. "PHPUnit Manual." https://phpunit.de/documentation.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (PHP-specific review)
- web-api-expert (Laravel/Symfony API development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
