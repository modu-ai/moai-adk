# SPEC Metadata Extended - Reference

## YAML Frontmatter Fields

### Field: `id`

**Type**: `string`
**Required**: ✅ Yes
**Format**: `{DOMAIN}-{NUMBER}` (e.g., `AUTH-001`, `CACHE-FIX-002`)
**Rules**:
- Domain: 2-8 characters, UPPERCASE
- Number: 3 digits, zero-padded (001, 002, 100)
- No duplicates allowed across project

**Validation**:
```bash
# Check for duplicates
rg "@SPEC:AUTH-001" .moai/specs/  # Should return exactly 1 hit

# Check format
rg "^id: [A-Z]+-\d{3}$" .moai/specs/SPEC-*/spec.md
```

### Field: `title`

**Type**: `string`
**Required**: ✅ Yes
**Max Length**: 128 characters
**Style**: User-facing, descriptive noun phrase

**Examples**:
- ✅ "User Authentication System"
- ✅ "Redis Cache Optimization"
- ❌ "Implement auth" (too vague)
- ❌ "User Authentication System for Multi-tenant Applications with Role-Based Access Control" (too long)

### Field: `version`

**Type**: `semver` (X.Y.Z)
**Required**: ✅ Yes
**Format**: `major.minor.patch`
**Rules**:
- Draft: `0.x.x` (no implementation yet)
- Alpha: `1.0.0-alpha` (partial implementation)
- Beta: `1.0.0-beta` (mostly done, testing)
- Release: `1.0.0+` (fully implemented)

**Examples**:
```yaml
version: 0.1.0  # Initial draft
version: 0.2.0  # Requirements refined
version: 1.0.0  # First implementation
version: 1.1.0  # Bug fixes + features
version: 2.0.0  # Major refactor
```

### Field: `status`

**Type**: `enum`
**Required**: ✅ Yes
**Allowed Values**:
- `draft` – Not ready for implementation
- `active` – Current work item
- `in-progress` – Being implemented
- `completed` – Implemented + tested
- `deprecated` – No longer used
- `archived` – Historical reference

**State Transitions**:
```
draft → active → in-progress → completed
                             ↓
                         deprecated
                             ↓
                          archived
```

### Field: `created`

**Type**: `ISO8601` (YYYY-MM-DD)
**Required**: ✅ Yes
**Example**: `2025-11-03`
**Rule**: Never change after creation

### Field: `updated`

**Type**: `ISO8601` (YYYY-MM-DD)
**Required**: ✅ Yes
**Example**: `2025-11-03`
**Rule**: Update whenever spec changes

### Field: `author`

**Type**: `string`
**Required**: ✅ Yes
**Format**: `@USERNAME` or `@NICKNAME`
**Examples**:
- `@GOOS🪿엉아`
- `@alice.smith`
- `@team-backend`

### Field: `priority`

**Type**: `enum`
**Required**: ✅ Yes
**Allowed Values**:
- `critical` – Blocking other work (P0)
- `high` – Important feature (P1)
- `medium` – Nice to have (P2)
- `low` – Backlog (P3)

## HISTORY Section Format

### Structure

```markdown
## HISTORY

### v1.1.0 (2025-11-05)
- Fixed race condition in verification
- Added timeout handling
- Updated error messages

### v1.0.0 (2025-11-03)
- Initial implementation complete
- All tests passing
- Documentation updated

### v0.2.0 (2025-11-02)
- Refined email requirements
- Added rate limiting constraints
- Extended EARS patterns

### v0.1.0 (2025-11-01)
- Initial draft with basic auth
- Ubiquitous requirements defined
```

### Rules

- Start with most recent version
- One entry per version change
- Bullet points, not paragraphs
- Link to commits if available
- Update `updated` field when modifying

## EARS Requirements Format

### Pattern 1: Ubiquitous

```markdown
- The system shall provide [capability].
- The system shall validate [input] before [action].
- The system shall [behavior] within [time] of [trigger].

Examples:
- The system shall provide user authentication via email.
- The system shall validate email format (RFC 5322) before storage.
- The system shall send verification email within 10 seconds of signup.
```

### Pattern 2: Event-driven

```markdown
- WHEN [condition], the system shall [behavior].
- Upon [event], the system shall [action] and [action].

Examples:
- WHEN a user clicks 'Sign Up', the system shall display signup form.
- Upon verification link click, the system shall activate user account.
- WHEN 3 failed attempts occur, the system shall lock the account.
```

### Pattern 3: State-driven

```markdown
- WHILE [state], the system shall [behavior].
- WHILE not [condition], the system shall [action].

Examples:
- WHILE the user is unauthenticated, the system shall deny access.
- WHILE session is active, the system shall refresh token automatically.
- WHILE rate limit not exceeded, the system shall process requests.
```

### Pattern 4: Optional Features

```markdown
- WHERE [condition], the system may [behavior].
- If [feature] enabled, the system may [action].

Examples:
- WHERE 2FA is enabled, the system may require additional verification.
- If API quota available, the system may allow batch operations.
- WHERE user preference set, the system may send notifications.
```

### Pattern 5: Constraints (Unwanted Behaviors)

```markdown
- IF [condition], the system shall [constraint].
- The system shall NOT [unwanted behavior].

Examples:
- IF password invalid 3x, the system shall lock account.
- The system shall NOT store plaintext passwords.
- IF token expired, the system shall return 401 error.
- The system shall NOT process requests exceeding rate limit.
```

## TAG Placement

### Location

Place `@SPEC:ID` in first 50 lines of spec.md:

```markdown
# User Authentication SPEC

@SPEC:AUTH-001

## Overview
...
```

### Usage in Code

Link implementation to spec:

```python
# src/auth.py
# @CODE:AUTH-001

def verify_email(email: str, token: str) -> bool:
    """Verify email with token. Requirement: AUTH-001"""
    # Implementation here
    pass
```

## Duplicate Detection

### Before Creating SPEC

```bash
# Check if ID already exists
rg "@SPEC:AUTH-001" .moai/specs/

# Expected: 0 hits (if new) or 1 hit (if existing)
# If >1 hit: ERROR – duplicate detected
```

### Duplicate Resolution

If duplicate found:
1. Use different number: AUTH-001 → AUTH-002
2. Or use different domain: AUTH-001 → SECURITY-001
3. Recheck with `rg "@SPEC:{NEW-ID}" .moai/specs/`

## Validation Checklist

```bash
#!/bin/bash
# validate-spec.sh - Verify SPEC completeness

SPEC_DIR=".moai/specs/SPEC-$1"

if [ ! -d "$SPEC_DIR" ]; then
  echo "❌ Directory not found: $SPEC_DIR"
  exit 1
fi

# Check metadata fields
for field in "id" "title" "version" "status" "created" "updated" "author" "priority"; do
  if ! grep "^$field:" "$SPEC_DIR/spec.md" > /dev/null; then
    echo "❌ Missing field: $field"
  fi
done

# Check SPEC ID in file
SPEC_ID=$(grep "^id:" "$SPEC_DIR/spec.md" | cut -d' ' -f2)
if ! grep "@SPEC:$SPEC_ID" "$SPEC_DIR/spec.md" > /dev/null; then
  echo "❌ Missing @SPEC:$SPEC_ID tag in spec.md"
fi

# Check for EARS patterns (at least 3)
EARS_COUNT=$(grep -E "^- (The system shall|WHEN|WHILE|WHERE|IF)" "$SPEC_DIR/spec.md" | wc -l)
if [ "$EARS_COUNT" -lt 5 ]; then
  echo "⚠️  Only $EARS_COUNT EARS requirements (recommend ≥5)"
fi

# Check HISTORY section
if ! grep "^## HISTORY" "$SPEC_DIR/spec.md" > /dev/null; then
  echo "❌ Missing HISTORY section"
fi

echo "✅ SPEC validation complete"
```

## File Structure

```
.moai/specs/SPEC-AUTH-001/
├── spec.md           # Requirements + metadata
├── plan.md           # Implementation plan (phase 2)
└── acceptance.md     # Test scenarios (phase 2)
```

### spec.md Contents (Phase 1)

```markdown
---
id: AUTH-001
title: "User Authentication System"
version: 0.1.0
status: active
created: 2025-11-03
updated: 2025-11-03
author: @USERNAME
priority: high
---

# User Authentication SPEC

@SPEC:AUTH-001

## Overview
[System description]

## Ubiquitous Requirements
[The system shall...]

## Event-driven Requirements
[WHEN... the system shall...]

## State-driven Requirements
[WHILE... the system shall...]

## Optional Features
[WHERE... the system may...]

## Constraints
[IF... the system shall...]

## HISTORY

### v0.1.0 (2025-11-03)
- Initial draft
```
