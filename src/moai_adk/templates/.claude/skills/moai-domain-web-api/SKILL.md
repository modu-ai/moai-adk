---
name: moai-domain-web-api
version: 4.0.0
created: 2025-10-22
updated: '2025-11-18'
status: stable
description: REST API and GraphQL design with OpenAPI 3.1, authentication, versioning,
  and rate limiting.
keywords:
- rest
- graphql
- openapi
- api-design
- authentication
allowed-tools:
- Read
- Bash
stability: stable
---


# Domain Web Api Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-domain-web-api |
| **Version** | 2.0.0 (2025-11-18) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Domain |

---

## What It Does

REST API and GraphQL design with OpenAPI 3.1, authentication, versioning, and rate limiting.

**Key capabilities**:
- ✅ Best practices enforcement for domain domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-18)
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

## Tool Version Matrix (2025-11-18)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **OpenAPI** | 3.1.0 | Primary | ✅ Current |
| **Postman** | 11.21.0 | Primary | ✅ Current |
| **Swagger UI** | 5.18.2 | Primary | ✅ Current |

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

- ** .0** (2025-11-18): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- ** .0** (2025-11-18): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-core-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)

---

## Best Practices

✅ **DO**:
- Follow domain best practices
- Use latest stable tool versions
- Maintain test coverage ≥85%
- Document all public APIs

❌ **DON'T**:
- Skip quality gates
- Use deprecated tools
- Ignore security warnings
- Mix testing frameworks
