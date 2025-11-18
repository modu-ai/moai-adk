# moai-spec-authoring Skill

**Version**: 1.2.0
**Created**: 2025-10-23
**Updated**: 2025-10-29
**Status**: Active
**Tier**: Foundation

## Overview

Comprehensive guide for authoring SPEC documents in MoAI-ADK projects. Provides complete YAML metadata structure, EARS syntax patterns, version management strategies, and validation tools.

## Key Features

- **7 Required + 9 Optional Metadata Fields**: Complete reference with lifecycle examples
- **5 Official EARS Patterns**: Ubiquitous, Event-driven, State-driven, Optional, Unwanted Behaviors
- **Version Lifecycle**: Semantic versioning from draft to production
- **Validation Tools**: Pre-submission checklist and automation scripts
- **Common Pitfalls**: Prevention strategies for 7 major issues

## File Structure (Progressive Disclosure)

```
.claude/skills/moai-spec-authoring/
├── SKILL.md          # Core overview + Quick Start (~500 words)
├── reference.md      # Complete metadata reference + EARS syntax
├── examples.md       # Real-world examples + patterns + troubleshooting
├── examples/
│   └── validate-spec.sh  # Automated SPEC validation script
└── README.md         # This file
```

## Quick Links

- **Quick Start**: [SKILL.md](./SKILL.md#quick-start-5-step-spec-creation)
- **Metadata Reference**: [reference.md](./reference.md#complete-metadata-field-reference)
- **EARS Syntax**: [reference.md](./reference.md#ears-requirement-syntax)
- **Examples**: [examples.md](./examples.md#real-world-ears-examples)
- **Troubleshooting**: [examples.md](./examples.md#troubleshooting)

## Usage

### Automatic Activation

This Skill automatically loads when:
- `/alfred:1-plan` command is executed
- SPEC document creation is requested
- Requirements clarification is discussed

### Manual Reference

Consult detailed sections for:
- SPEC authoring best practices
- Existing SPEC document validation
- Metadata issue troubleshooting
- EARS syntax pattern reference

## Validation Command

```bash
# Validate SPEC metadata
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-AUTH-001/spec.md

# Check for duplicate IDs

# Scan entire TAG chain
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# Use automated script
./examples/validate-spec.sh .moai/specs/SPEC-AUTH-001
```

## Example SPEC Structure

```markdown
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-29
updated: 2025-10-29
priority: high
---


## HISTORY
### v0.0.1 (2025-10-29)
- **INITIAL**: JWT authentication SPEC draft

## Environment
**Runtime**: Node.js 20.x

## Assumptions
1. User storage: PostgreSQL
2. Secret management: Environment variables

## Requirements

### Ubiquitous Requirements
**UR-001**: The system shall provide JWT-based authentication.

### Event-driven Requirements
**ER-001**: WHEN the user submits valid credentials, the system shall issue a JWT token.

### State-driven Requirements
**SR-001**: WHILE the user is in an authenticated state, the system shall permit access to protected resources.

### Optional Features
**OF-001**: WHERE multi-factor authentication is enabled, the system can require OTP verification.

### Unwanted Behaviors
**UB-001**: IF a token has expired, THEN the system shall deny access and return HTTP 401.
```

## Integration

Works seamlessly with:
- `spec-builder` agent - SPEC creation
- `moai-foundation-ears` - EARS syntax patterns
- `moai-foundation-specs` - Metadata validation
- `moai-foundation-tags` - TAG system integration

## Support

For questions or issues:
1. Refer to comprehensive documentation: `SKILL.md`, `reference.md`, `examples.md`
2. Use `/alfred:1-plan` for guided SPEC creation
3. Review existing SPECs in `.moai/specs/` for examples

---

**Maintained By**: MoAI-ADK Team
**Last Updated**: 2025-10-29
