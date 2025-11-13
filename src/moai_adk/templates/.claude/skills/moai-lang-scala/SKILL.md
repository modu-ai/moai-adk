---
name: "moai-lang-scala"
version: "2.1.0"
created: 2025-10-22
updated: 2025-11-02
status: stable
description: Scala 3.6+ best practices with ScalaTest 3.2, sbt 1.10, functional programming patterns, and Play Framework.
keywords: ['scala', 'scalatest', 'sbt', 'functional', 'play-framework']
allowed-tools: 
  - Read
  - Bash
---

# Lang Scala Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-scala |
| **Version** | 2.1.0 (2025-11-02) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

Scala 3.6+ best practices with ScalaTest 3.2, sbt 1.10, functional programming patterns, and Play Framework.

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-02)
- ✅ TDD workflow support
- ✅ Play Framework web application patterns

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
| **Scala** | 3.6.0 | Runtime | ✅ Current |
| **ScalaTest** | 3.2.19 | Testing | ✅ Current |
| **sbt** | 1.10.0 | Build tool | ✅ Current |
| **Play Framework** | 3.0.9 | Web framework | ✅ Current |

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
