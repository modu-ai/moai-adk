---
name: "moai-lang-php"
version: "2.1.0"
created: 2025-10-22
updated: 2025-11-02
status: stable
description: PHP 8.4+ best practices with PHPUnit 11, Composer, PSR-12 standards, and web frameworks (Laravel, Symfony).
keywords: ['php', 'phpunit', 'composer', 'psr', 'laravel', 'symfony']
allowed-tools: 
  - Read
  - Bash
---

# Lang Php Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-php |
| **Version** | 2.1.0 (2025-11-02) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

PHP 8.4+ best practices with PHPUnit 11, Composer, PSR-12 standards, and web frameworks (Laravel, Symfony).

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-02)
- ✅ TDD workflow support
- ✅ Web framework patterns (Laravel, Symfony)

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-11-02)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **PHP** | 8.4.0 | Runtime | ✅ Current |
| **PHPUnit** | 11.5.0 | Testing | ✅ Current |
| **Composer** | 2.8.0 | Package manager | ✅ Current |
| **Laravel** | 12.0.0 | Web framework | ✅ Current |
| **Symfony** | 7.3.5 | Web framework | ✅ Current |

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## References (Latest Documentation)

_Documentation links updated 2025-10-22_

---

## Changelog

- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)

---

## Best Practices

✅ **DO**:
- Follow language best practices
- Use latest stable tool versions
- Maintain test coverage ≥85%
- Document all public APIs

❌ **DON'T**:
- Skip quality gates
- Use deprecated tools
- Ignore security warnings
- Mix testing frameworks
