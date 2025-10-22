# SPEC Documentation Standards

> YAML frontmatter validation and HISTORY section enforcement

_Last updated: 2025-10-22_

---

## SPEC File Structure

### Required YAML Frontmatter (7 Fields)

```yaml
---
id: SPEC-<DOMAIN>-<###>
version: 0.0.1
status: draft | active | completed | archived
created: YYYY-MM-DD
updated: YYYY-MM-DD
author: <name>
tags: [tag1, tag2, tag3]
---
```

**Field Validation Rules**:
- `id`: Pattern `SPEC-[A-Z]+-[0-9]{3}` (e.g., `SPEC-AUTH-001`)
- `version`: Semantic versioning `X.Y.Z`
- `status`: One of: `draft`, `active`, `completed`, `archived`
- `created`: ISO date format `YYYY-MM-DD`
- `updated`: ISO date format `YYYY-MM-DD`
- `author`: Non-empty string
- `tags`: Array of lowercase strings

---

## HISTORY Section Format

**Required at end of every SPEC**:

```markdown
## HISTORY

### v0.0.1 (YYYY-MM-DD)
- **INITIAL**: Brief description of initial SPEC creation

### v0.1.0 (YYYY-MM-DD)
- **FEATURE**: Description of feature addition
- **CHANGE**: Description of requirement modification

### v1.0.0 (YYYY-MM-DD)
- **RELEASE**: Production-ready release notes
- **BREAKING**: Description of breaking changes
```

**Change Types**:
- `INITIAL`: First version
- `FEATURE`: New requirement added
- `CHANGE`: Existing requirement modified
- `FIX`: Correction or clarification
- `BREAKING`: Breaking change
- `RELEASE`: Production release milestone

---

## Complete SPEC Template

```markdown
---
id: SPEC-AUTH-001
version: 0.1.0
status: active
created: 2025-10-22
updated: 2025-10-22
author: MoAI Team
tags: [authentication, security, jwt]
---

# @SPEC:AUTH-001 | JWT Authentication

## Overview
Brief description of the feature and its purpose.

## Requirements (EARS Format)

### Ubiquitous
- The authentication service shall support JWT tokens with HS256 signing.
- The API shall enforce HTTPS for all authentication endpoints.

### Event-Driven
- When the user submits valid credentials, the system shall return a JWT token.

### Unwanted Behavior
- If invalid credentials are provided, then the system shall return HTTP 401.

## Technical Details
- **Algorithm**: HS256
- **Token Expiration**: 24 hours
- **Storage**: HTTP-only cookie

## Test Cases
- ✅ Valid credentials return token
- ✅ Invalid credentials return 401
- ✅ Expired tokens return 401

## References
- Related SPECs: SPEC-AUTH-002
- External docs: JWT RFC 7519

## HISTORY

### v0.1.0 (2025-10-22)
- **FEATURE**: Add JWT token validation
- **CHANGE**: Update expiration from 1h to 24h

### v0.0.1 (2025-10-15)
- **INITIAL**: Draft JWT authentication SPEC
```

---

## Validation Script

```bash
#!/bin/bash
# Validate SPEC frontmatter

validate_spec() {
  local spec_file="$1"

  # Check required fields
  for field in id version status created updated author tags; do
    if ! grep -q "^$field:" "$spec_file"; then
      echo "❌ Missing required field: $field"
      return 1
    fi
  done

  # Validate ID format
  if ! grep -E "^id: SPEC-[A-Z]+-[0-9]{3}$" "$spec_file"; then
    echo "❌ Invalid ID format (must be SPEC-DOMAIN-###)"
    return 1
  fi

  # Check HISTORY section exists
  if ! grep -q "^## HISTORY$" "$spec_file"; then
    echo "❌ Missing HISTORY section"
    return 1
  fi

  echo "✅ SPEC validation passed"
  return 0
}

# Usage
validate_spec ".moai/specs/SPEC-AUTH-001/spec.md"
```

---

## Resources

**Semantic Versioning**: https://semver.org/
**YAML Spec**: https://yaml.org/spec/
**EARS Integration**: See `moai-foundation-ears` Skill

---

**Last Updated**: 2025-10-22
**Maintained by**: MoAI-ADK Foundation Team
