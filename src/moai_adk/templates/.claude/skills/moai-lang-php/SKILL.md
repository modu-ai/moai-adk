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
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- PHP Manual. "PHP Documentation." https://www.php.net/manual/en/ (accessed 2025-03-29).
- PHPUnit. "PHPUnit Manual." https://phpunit.de/documentation.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (PHP-specific review)
- web-api-expert (Laravel/Symfony API development)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
