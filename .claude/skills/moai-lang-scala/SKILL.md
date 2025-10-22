---
name: moai-lang-scala
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Scala 3.4+ best practices with ScalaTest 3.2, sbt 1.9, functional programming patterns.
keywords: ['scala', 'scalatest', 'sbt', 'functional']
allowed-tools:
  - Read
  - Bash
---

# Lang Scala Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-scala |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

Scala 3.4+ best practices with ScalaTest 3.2, sbt 1.9, functional programming patterns.

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ TDD workflow support

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

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Scala** | 3.4.0 | Compiler | ✅ Current |
| **ScalaTest** | 3.2.0 | Testing | ✅ Current |
| **sbt** | 1.9.0 | Build Tool | ✅ Current |
| **Scalafmt** | 3.8.0 | Formatting | ✅ Current |

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
- Prefer immutability and pure functions
- Use ScalaTest for testing
- Maintain test coverage ≥85%
- Leverage Scala 3 features

❌ **DON'T**:
- Skip quality gates
- Use null values
- Ignore compiler warnings
- Mix functional and imperative styles
