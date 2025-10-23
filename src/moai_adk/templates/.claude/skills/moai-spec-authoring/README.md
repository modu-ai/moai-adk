# moai-spec-authoring Skill

**Version**: 1.0.0  
**Created**: 2025-10-23  
**Status**: Active  
**Tier**: Foundation

## Overview

Comprehensive guide for authoring high-quality MoAI-ADK SPEC documents with proper metadata, EARS syntax, and best practices.

## Key Features

- **7 Required + 9 Optional Metadata Fields**: Complete reference with examples
- **5 EARS Patterns**: Ubiquitous, Event-driven, State-driven, Optional, Constraints
- **Version Lifecycle**: Semantic versioning from draft to production
- **TAG Integration**: @SPEC, @TEST, @CODE, @DOC chain management
- **Validation Tools**: Pre-submission checklist and automated scripts
- **Common Pitfalls**: Prevention strategies for 7 major issues

## Quick Links

- [5-Step Quick Start](#quick-start-5-step-spec-creation)
- [Metadata Reference](#metadata-reference)
- [EARS Syntax Guide](#ears-requirement-syntax)
- [Validation Checklist](#pre-submission-validation-checklist)
- [Common Pitfalls](#common-pitfalls--prevention)

## Usage

### Automatic Activation

This Skill automatically loads during:
- `/alfred:1-plan` command execution
- SPEC document creation requests
- Requirement clarification discussions

### Manual Reference

Access detailed sections for:
- Learning SPEC authoring best practices
- Validating existing SPEC documents
- Troubleshooting metadata issues
- Understanding EARS syntax patterns

## File Structure

```
.claude/skills/moai-spec-authoring/
├── SKILL.md       # Complete Skill documentation (1,300 lines)
└── README.md      # This file
```

## Integration

Works seamlessly with:
- `spec-builder` agent - SPEC generation
- `moai-foundation-ears` - EARS syntax patterns
- `moai-foundation-specs` - Metadata validation
- `moai-foundation-tags` - TAG system integration

## Validation Command

```bash
# Validate SPEC metadata
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-AUTH-001/spec.md

# Check duplicate IDs
rg "@SPEC:AUTH-001" -n .moai/specs/

# Full TAG chain scan
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n
```

## Example SPEC Structure

```markdown
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @YourHandle
priority: high
---

# @SPEC:AUTH-001: JWT Authentication System

## HISTORY
### v0.0.1 (2025-10-23)
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
**ER-001**: WHEN a user submits valid credentials, the system shall issue a JWT token.

### State-driven Requirements
**SR-001**: WHILE a user is authenticated, the system shall allow access to protected resources.

### Optional Features
**OF-001**: WHERE MFA is enabled, the system can require OTP verification.

### Constraints
**C-001**: IF a token is expired, the system shall deny access.
```

## Support

For questions or issues:
1. Consult `SKILL.md` for comprehensive documentation
2. Invoke `/alfred:1-plan` for guided SPEC creation
3. Reference existing SPECs in `.moai/specs/` for examples

---

**Maintained By**: MoAI-ADK Team  
**Last Updated**: 2025-10-23
