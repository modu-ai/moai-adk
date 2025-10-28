---
name: moai-foundation-ears
version: 2.1.0
created: 2025-10-22
updated: 2025-10-29
status: active
description: EARS requirement authoring guide (Ubiquitous/Event-driven/State-driven/Optional/Unwanted Behaviors) with 5 official patterns.
keywords: ['ears', 'requirements', 'authoring', 'syntax', 'unwanted-behaviors']
allowed-tools:
  - Read
  - Bash
---

# Foundation Ears Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-ears |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Foundation |

---

## What It Does

Official EARS (Easy Approach to Requirements Syntax) requirement authoring guide with 5 patterns: Ubiquitous, Event-driven, State-driven, Optional, and Unwanted Behaviors.

**Key capabilities**:
- ✅ Five official EARS patterns with real-world examples
- ✅ Best practices enforcement for foundation domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-29)
- ✅ TDD workflow support
- ✅ Unwanted Behaviors pattern for error handling & quality gates

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

- **v2.1.0** (2025-10-29): Standardized Unwanted Behaviors as 5th official EARS pattern, replacing Constraints terminology
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
- Follow foundation best practices
- Use latest stable tool versions
- Maintain test coverage ≥85%
- Document all public APIs

❌ **DON'T**:
- Skip quality gates
- Use deprecated tools
- Ignore security warnings
- Mix testing frameworks
