---
name: "moai-foundation-tags"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: foundation
description: "Complete TAG system guide covering @SPEC, @TEST, @CODE, @DOC chains, TAG lifecycle management, orphan detection, validation rules, and enterprise traceability patterns. Enterprise v4.0 with November 2025 stable coverage."
allowed-tools: "Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: [tag-agent, qa-validator, doc-syncer]
keywords: [TAG-system, traceability, SPEC, TEST, CODE, DOC, tagging, metadata, governance]
tags: [foundation, tags, traceability, chains, validation, enterprise]
orchestration: 
can_resume: true
typical_chain_position: "foundational"
depends_on: []
---

# moai-foundation-tags

**The Complete @TAG System & Traceability Framework**

> **Version**: 4.0.0 Enterprise  
> **Tier**: Foundation  
> **Updated**: November 2025 Stable  
> **Keywords**: TAG-system, traceability, @SPEC, @TEST, @CODE, @DOC, governance

---

## Progressive Disclosure

### Level 1: Core Concepts (Quick Start)

### What It Does

This foundational Skill provides the **complete operational guide** for MoAI-ADK's @TAG system:

- **TAG System Architecture**: Four-chain model (@SPEC → @TEST → @CODE → @DOC)
- **TAG Lifecycle**: Creation, linking, validation, and deprecation
- **Orphan Detection & Cleanup**: Identify and remediate unreferenced TAGs
- **Validation Rules**: Enterprise-grade governance and compliance
- **Traceability Chains**: Link requirements to implementation to documentation
- **Enterprise Patterns**: November 2025 stable practices
- **50+ Official Documentation Links**: Official references and standards

**Core Principle**: TAGs are the **source of truth for traceability**, linking requirements through implementation to documentation with complete auditability.

---

## The Four-Chain TAG Model

### Chain Structure

Every feature in MoAI-ADK follows a complete chain from specification to documentation:

```
Requirement Layer (SPEC)
    ↓
    @SPEC-001: Define requirements
        ↓ (depends on)
    TEST Layer (@TEST)
        ↓
        @TEST-SPEC-001: Write failing tests
        @TEST-SPEC-001-001: Unit test
        @TEST-SPEC-001-002: Integration test
            ↓ (validates)
    Implementation Layer (@CODE)
        ↓
        @CODE-SPEC-001: Implement feature
        @CODE-SPEC-001-001: Main function
        @CODE-SPEC-001-002: Helper logic
        @CODE-SPEC-001-003: Error handling
            ↓ (fulfills)
    Documentation Layer (@DOC)
        ↓
        @DOC-SPEC-001: Document feature
        @DOC-SPEC-001-001: User guide
        @DOC-SPEC-001-002: API reference
        @DOC-SPEC-001-003: Migration guide
```

### Chain Semantics (November 2025)

Each link type defines a specific relationship:

| Chain Type | Direction | Meaning | Validation |
|-----------|-----------|---------|-----------|
| **@SPEC** | Requirement definition | "What must be done?" | Required in .moai/specs/ |
| **@TEST** | Validation logic | "How do we verify?" | ≥85% code coverage |
| **@CODE** | Implementation | "How is it built?" | Must reference @SPEC |
| **@DOC** | Knowledge capture | "How do users use it?" | Links to @SPEC/@CODE |

### Example: Complete Chain

```markdown
1. SPEC Created:
   File: .moai/specs/SPEC-001/spec.md
   Header: # SPEC-001: Authentication System
   TAG: @SPEC:AUTH-001
   Status: APPROVED

2. Tests Written (RED phase):
   File: tests/test_auth.py
   Content:
     """
     @TEST:SPEC:AUTH-001
     Validates SPEC-001 authentication requirements
     """
     def test_login_with_valid_credentials():
         # @TEST-SPEC-001-001: User login with correct password
         assert authenticate("user", "password") == True

3. Code Implemented (GREEN phase):
   File: src/auth.py
   Content:
     @CODE:SPEC:AUTH-001
     def authenticate(username, password):
         # @CODE-SPEC-001-001: Password validation
         if validate_password(password):
             return True

4. Documentation Written:
   File: docs/auth.md
   Header: # Authentication
   Content:
     @DOC:SPEC:AUTH-001
     ## User Login
     Users authenticate using username and password...

5. Traceability Chain:
   @SPEC:AUTH-001 --implements--> @TEST:SPEC:AUTH-001
   @TEST:SPEC:AUTH-001 --validates--> @CODE:SPEC:AUTH-001
   @CODE:SPEC:AUTH-001 --described-by--> @DOC:SPEC:AUTH-001
```

---

## TAG Lifecycle Management

### Phase 1: Birth (Creation)

**When**: SPEC document is created  
**Where**: `.moai/specs/SPEC-XXX/spec.md`  
**Format**: `@SPEC:DOMAIN-NNN` or `@SPEC:PROJECT-NAME-NNN`  
**Validation**: Must be unique in codebase

```markdown
---
title: SPEC-001 User Authentication
created_at: 2025-11-12
status: DRAFT
@SPEC:AUTH-001
---

# User Authentication System

## Requirements
- Support username/password login
- Hash passwords using bcrypt
- Rate limit login attempts
```

### Phase 2: Active Development (Growth)

**TEST TAGs Created**:
```python
"""
@TEST:SPEC:AUTH-001-001
Validates user login with valid credentials
Depends on: @SPEC:AUTH-001
Coverage: Happy path
"""
def test_login_valid():
    assert authenticate("user", "password") == True
```

**CODE TAGs Created**:
```python
def authenticate(username, password):
    """
    @CODE:SPEC:AUTH-001
    Implements authentication logic
    References: @TEST:SPEC:AUTH-001-001
    """
    return verify_password(username, password)
```

**Documentation TAGs Created**:
```markdown
## Authentication
@DOC:SPEC:AUTH-001
Authentication module handles user login...
```

### Phase 3: Maturity (Stability)

**All chain links validated**:
- @SPEC exists and approved
- @TEST coverage >= 85%
- @CODE references @SPEC
- @DOC links to @SPEC/@CODE
- Traceability chain complete

**Quality Gate Checks** (November 2025):
- No broken references
- No orphan TAGs
- Version tags current
- Deprecation warnings applied

### Phase 4: Sunset (Deprecation)

**When**: Feature is superseded or removed  
**Process**:
```markdown
@DEPRECATED:SPEC:AUTH-001
Superseded by: @SPEC:AUTH-002
Migration path: See @DOC:AUTH-MIGRATION-001
End-of-life: 2026-05-01
```

**Orphan Detection**:
- Scan for unreferenced @SPEC TAGs
- Find @TEST TAGs without @SPEC links
- Identify @CODE without @TEST
- Flag @DOC without source references

---

## TAG Validation Rules (Enterprise v4.0)

### Mandatory Validation (STRICT Mode)

All TAGs must satisfy:

```
Rule 1: Unique Identifier
├─ TAG must be unique across entire codebase
├─ Format: @[TYPE]:[DOMAIN]-[NNN]
└─ Example: @SPEC:AUTH-001, @TEST:SPEC:AUTH-001

Rule 2: Proper Hierarchy
├─ @SPEC must exist before @TEST/@CODE
├─ @TEST must exist before @CODE
├─ @DOC must reference @SPEC or @CODE
└─ Circular references forbidden

Rule 3: Complete Chains
├─ Every @SPEC should have @TEST
├─ Every @TEST should validate a @SPEC
├─ Every @CODE should implement a @SPEC
├─ Every @DOC should reference source
└─ No orphan TAGs allowed

Rule 4: Version Compliance
├─ TAG references must use current versions
├─ Deprecated features marked @DEPRECATED
├─ Migration paths documented
└─ November 2025 standards enforced

Rule 5: Documentation
├─ Every TAG needs description
├─ References documented
├─ Dependencies explicit
└─ Status clearly marked
```

### Validation Commands (November 2025)

```bash
# Scan for TAGs
rg '@SPEC|@TEST|@CODE|@DOC' -n

# Find orphans
rg '@SPEC:[A-Z]+-[0-9]+' -n | \
  while read spec; do
    grep -r "$spec" --exclude-dir=.git || echo "ORPHAN: $spec"
  done

# Verify chains
python .moai/scripts/validation/tag_chain_validator.py

# Generate TAG report
python .moai/scripts/analysis/tag_analyzer.py --report
```

---

## Orphan TAG Detection & Cleanup

### What Are Orphans?

Orphan TAGs are references that exist but lack proper linking:

**Type A: Unreferenced @SPEC**
```python
# Problem: Created but never used
@SPEC:UNUSED-001  # No @TEST links this

# Solution: Either implement or deprecate
@DEPRECATED:SPEC:UNUSED-001
Reason: Feature cancelled
Decision: 2025-11-12
```

**Type B: Floating @TEST**
```python
# Problem: Test exists but @SPEC missing
@TEST:ORPHAN-001  # No @SPEC to validate

# Solution: Link to existing @SPEC or create it
@TEST:SPEC:AUTH-001-001  # Now linked
```

**Type C: Stray @CODE**
```python
# Problem: Code has TAG but no @SPEC/@TEST
def feature():
    @CODE:FLOATING-001
    # Missing @SPEC and @TEST

# Solution: Trace requirements
@SPEC:FEATURE-001  # Create
@TEST:SPEC:FEATURE-001  # Create
@CODE:SPEC:FEATURE-001  # Link properly
```

**Type D: Orphan @DOC**
```markdown
# Documentation exists but no source
@DOC:UNDOCUMENTED-001  # Missing @SPEC/@CODE links

# Solution: Link to implementation
@DOC:SPEC:AUTH-001  # Now linked
```

### Detection Workflow (November 2025)

**Step 1: Scan Phase**
```bash
# Extract all TAGs
rg '@(SPEC|TEST|CODE|DOC):[A-Z]+-[0-9]+' -o --no-filename | sort | uniq > /tmp/all_tags.txt

# Create index
python << 'PYEOF'
import re
import subprocess

tags = {}
for line in open('/tmp/all_tags.txt'):
    tag = line.strip()
    tags[tag] = {'refs': 0, 'locations': []}

# Find references
result = subprocess.run(['rg', '@(SPEC|TEST|CODE|DOC):[A-Z]+-[0-9]+', '-n'], capture_output=True, text=True)
for match in result.stdout.split('\n'):
    for tag in tags:
        if tag in match:
            tags[tag]['refs'] += 1
            tags[tag]['locations'].append(match.split(':')[0])

# Report orphans
orphans = [tag for tag, data in tags.items() if data['refs'] == 1]
print(f"Found {len(orphans)} orphan TAGs")
for orphan in orphans:
    print(f"  - {orphan}: {tags[orphan]['locations'][0]}")
PYEOF
```

**Step 2: Analysis Phase**
```python
# Determine orphan type
def analyze_orphan(tag):
    if tag.startswith('@SPEC'):
        # Check if @TEST references it
        if search_references(f'{tag}', ['@TEST']):
            return 'OK'
        else:
            return 'ORPHAN_TYPE_A'
    elif tag.startswith('@TEST'):
        # Check if linked to @SPEC
        if has_spec_link(tag):
            return 'OK'
        else:
            return 'ORPHAN_TYPE_B'
    # ... continue for @CODE, @DOC
```

**Step 3: Remediation Phase**
```bash
# Option 1: Link existing orphan
# Change: @TEST:ORPHAN-001
# To:     @TEST:SPEC:AUTH-001-001

# Option 2: Mark as deprecated
# Add: @DEPRECATED:SPEC:UNUSED-001

# Option 3: Create missing link
# Create SPEC if orphan is @TEST/@CODE
```

---

## Traceability Matrix & Reports

### Enterprise Traceability Dashboard

**Monthly Report** (November 2025):

```
TAG System Health Report
Generated: 2025-11-12
Coverage: 1,247 features

Total TAGs: 4,892
├─ @SPEC: 1,000 (100% approved)
├─ @TEST: 3,000 (99.2% coverage)
├─ @CODE: 2,500 (100% implemented)
└─ @DOC: 892 (89.2% complete)

Quality Metrics:
├─ Orphan TAGs: 0 (0%)
├─ Broken chains: 0 (0%)
├─ Deprecated (active): 23 (1.8%)
├─ Test coverage: 96.4%
└─ Documentation complete: 89.2%

Chain Integrity:
├─ @SPEC → @TEST: 100%
├─ @TEST → @CODE: 99.8%
├─ @CODE → @DOC: 89.2%
└─ Circular refs: 0

Validation Status:
├─ Passed: 4,892 TAGs (100%)
├─ Warnings: 12 (0.24%)
├─ Errors: 0 (0%)
└─ Last scan: 2025-11-12 10:30 UTC
```

### TAG Cross-Reference Matrix

```
Feature     | SPEC    | TEST    | CODE    | DOC     | Status
------------|---------|---------|---------|---------|----------
Auth        | @SPEC:AUTH-001 | ✓ 12 | ✓ 45 | ✓ 8 | COMPLETE
Payment     | @SPEC:PAY-001  | ✓ 18 | ✓ 62 | ✓ 5 | COMPLETE
Reporting   | @SPEC:REP-001  | ✓ 25 | ✓ 88 | ⚠ 3 | MISSING_DOC
Admin       | @SPEC:ADM-001  | ✓ 8  | ✓ 28 | ✓ 4 | COMPLETE
Legacy API  | @DEPRECATED    | -    | -    | ✓ 2 | DEPRECATED
```

---

## Enterprise TAG Governance (November 2025)

### TAG Naming Convention

```
@[TYPE]:[NAMESPACE]-[NUMBER]
  ↑       ↑           ↑
  │       │           └─ Sequential identifier (001-999)
  │       └─ Feature domain (AUTH, PAY, REP, etc)
  └─ Type (SPEC, TEST, CODE, DOC)

Examples:
├─ @SPEC:AUTH-001           (Requirement)
├─ @TEST:SPEC:AUTH-001-001  (Unit test)
├─ @CODE:SPEC:AUTH-001-001  (Implementation)
└─ @DOC:SPEC:AUTH-001       (Documentation)
```

### TAG Assignment Matrix

| Component | Responsible | Timeline | Review |
|-----------|-------------|----------|--------|
| @SPEC | spec-builder | Day 1 | plan-agent |
| @TEST | test-engineer | Day 2 | qa-validator |
| @CODE | tdd-implementer | Days 3-5 | code-reviewer |
| @DOC | doc-syncer | Day 6 | technical-writer |

### Deprecation Policy

```
When removing features:

1. Mark with @DEPRECATED
   @DEPRECATED:SPEC:OLD-001
   Superseded by: @SPEC:NEW-001
   EOL Date: 2026-05-01

2. Create migration guide
   @DOC:MIGRATION:OLD-001-to-NEW-001
   Link: docs/migrations/old-to-new.md

3. Update @CODE references
   All @CODE:SPEC:OLD-001 become
   @CODE:SPEC:OLD-001 @DEPRECATED

4. Maintain 6-month runway
   Old feature supported during migration window
```

---

## Integration with MoAI-ADK Tools

### TAG Scanning

```bash
# Built-in tag-agent
Skill("moai-foundation-tags")  # This Skill
Task(subagent_type="tag-agent")  # Scan and validate TAGs

# Configuration
.moai/config.json:
{
  "tags": {
    "auto_sync": true,
    "storage_type": "code_scan",
    "policy": {
      "enforcement_mode": "strict",
      "enforce_chains": true
    }
  }
}
```

### TAG-Driven Workflow

```
1. /alfred:1-plan → Creates @SPEC
2. /alfred:2-run → Creates @TEST, @CODE (RED-GREEN-REFACTOR)
3. /alfred:3-sync → Creates @DOC
4. Validation → TAG chain verified
```

---

## November 2025 Stable Standards

### Version Requirements

This Skill covers:
- MoAI-ADK v0.22.5+
- Python 3.12+
- TAG system v4.0.0
- RFC 2119 (MUST/SHOULD/MAY keywords)
- November 2025 stable patterns

### Breaking Changes from v3.x

```markdown
v3.x Pattern:
  @TAG-OLD-001

v4.0.0 Pattern:
  @SPEC:AUTH-001          (Specification)
  @TEST:SPEC:AUTH-001     (Testing)
  @CODE:SPEC:AUTH-001     (Implementation)
  @DOC:SPEC:AUTH-001      (Documentation)

Migration:
  1. Audit existing @TAG-* references
  2. Map to v4.0.0 type system
  3. Create chain links
  4. Update references
  5. Validate chains
```

---

### Level 2: Practical Implementation

## Common TAG Patterns & Examples

### Pattern 1: New Feature Implementation

**Day 1: Spec Created**
```markdown
# SPEC-042: Email Notifications

@SPEC:EMAIL-NOTIFY-042

## Requirements
- Send email on user signup
- Support HTML and plain text
- Rate limit to 100/minute
```

**Day 2-3: Tests Written**
```python
# tests/test_email.py

@TEST:SPEC:EMAIL-NOTIFY-042-001
def test_send_email_on_signup():
    """Verify email sent after user registration"""
    user = create_user("test@example.com")
    assert email_sent_to("test@example.com")

@TEST:SPEC:EMAIL-NOTIFY-042-002
def test_rate_limiting():
    """Verify rate limit enforced"""
    for i in range(101):
        send_email()
    # 101st request should fail
    with pytest.raises(RateLimitError):
        send_email()
```

**Day 4-5: Code Implementation**
```python
# src/notifications.py

@CODE:SPEC:EMAIL-NOTIFY-042
def send_email(to, subject, body, html=False):
    """
    Send email notification
    
    References:
    - @TEST:SPEC:EMAIL-NOTIFY-042-001
    - @TEST:SPEC:EMAIL-NOTIFY-042-002
    """
    # @CODE:SPEC:EMAIL-NOTIFY-042-001: Rate limit check
    if rate_limiter.check(to):
        raise RateLimitError()
    
    # @CODE:SPEC:EMAIL-NOTIFY-042-002: Send via SMTP
    smtp = get_smtp_connection()
    smtp.send(to, subject, body, html=html)
```

**Day 6: Documentation**
```markdown
# Email Notifications

@DOC:SPEC:EMAIL-NOTIFY-042

## Overview
Send transactional emails to users...

## API Reference

### send_email(to, subject, body, html=False)
@DOC:SPEC:EMAIL-NOTIFY-042-001
Sends an email notification...

### Rate Limiting
@DOC:SPEC:EMAIL-NOTIFY-042-002
Rate limited to 100 emails per minute per recipient...
```

### Pattern 2: Refactoring with TAGs

**Before (Monolithic)**
```python
@CODE:AUTH-OLD-001
def authenticate(user, password, mfa=None):
    # 200 lines of validation logic
    pass
```

**After (Refactored)**
```python
@CODE:SPEC:AUTH-REFACTOR-001
def authenticate(user, password, mfa=None):
    """
    @CODE:SPEC:AUTH-REFACTOR-001
    Refactored from @CODE:AUTH-OLD-001
    """
    validate_password(user, password)  # @CODE:SPEC:AUTH-REFACTOR-001-001
    if mfa:
        validate_mfa(user, mfa)  # @CODE:SPEC:AUTH-REFACTOR-001-002
    return True

@CODE:SPEC:AUTH-REFACTOR-001-001
def validate_password(user, password):
    """Extracted password validation"""
    # 50 lines

@CODE:SPEC:AUTH-REFACTOR-001-002
def validate_mfa(user, code):
    """Extracted MFA validation"""
    # 50 lines
```

### Pattern 3: Bug Fix with TAGs

**Issue Reported**
```
BUG-101: Password reset token expires prematurely
Linked to: @SPEC:AUTH-001
```

**Fix Process**
```python
@SPEC:AUTH-BUG-FIX-101
title: Fix password reset token expiration

@TEST:SPEC:AUTH-BUG-FIX-101-001
def test_token_valid_24_hours():
    """Token should remain valid for 24 hours"""
    token = generate_reset_token("user@example.com")
    assert token_valid_after(token, hours=23)
    assert not token_valid_after(token, hours=25)

@CODE:SPEC:AUTH-BUG-FIX-101
def generate_reset_token(email):
    """Fixed: was using 12 hours, now 24 hours"""
    # @CODE:SPEC:AUTH-BUG-FIX-101-001: Corrected expiration
    return create_token(email, expires_in_hours=24)  # Was 12

@DOC:SPEC:AUTH-BUG-FIX-101
## Password Reset Token Fix
@DOC:SPEC:AUTH-BUG-FIX-101-001: Token valid for 24 hours
```

---

### Level 3: Advanced Governance

## Enterprise TAG Audit & Compliance

### Quarterly Audit Checklist

```bash
#!/bin/bash
# TAG System Quarterly Audit - November 2025

echo "=== TAG System Audit Report ==="
echo "Generated: $(date)"
echo ""

# 1. Count TAGs by type
echo "TAG Distribution:"
rg '@SPEC:' --count-matches
rg '@TEST:' --count-matches
rg '@CODE:' --count-matches
rg '@DOC:' --count-matches

# 2. Find orphans
echo ""
echo "Checking for orphan TAGs..."
python << 'PYEOF'
# Implementation: Tag chain validation
import re
import subprocess

specs = set(re.findall(r'@SPEC:[A-Z]+-\d+', 
    subprocess.run(['rg', '@SPEC:', '--no-filename', '-o'], capture_output=True, text=True).stdout))

orphans = []
for spec in specs:
    # Check if any TEST, CODE, or DOC references it
    refs = subprocess.run(['rg', f'{spec}.*(@TEST|@CODE|@DOC)', '--count-matches'],
        capture_output=True, text=True)
    if refs.stdout.strip() == '0':
        orphans.append(spec)

print(f"Found {len(orphans)} orphan SPECs")
for orphan in orphans:
    print(f"  {orphan}")
PYEOF

# 3. Validate chains
echo ""
echo "Validating chain integrity..."
python .moai/scripts/validation/tag_chain_validator.py

# 4. Coverage report
echo ""
echo "Test Coverage by TAG:"
python << 'PYEOF'
# Coverage analysis by @SPEC TAG
import json
coverage_data = json.load(open('.coverage'))
for spec, coverage in coverage_data.items():
    if coverage < 0.85:
        print(f"WARNING: {spec} coverage {coverage*100:.1f}% < 85%")
PYEOF

# 5. Deprecation review
echo ""
echo "Deprecated features review:"
rg '@DEPRECATED' -n
```

### Integration with CI/CD

```yaml
# .github/workflows/tag-validation.yml
name: TAG System Validation

on: [push, pull_request]

jobs:
  tag-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install ripgrep
        run: |
          sudo apt-get update
          sudo apt-get install -y ripgrep
      - name: Scan for TAGs
        run: |
          python .moai/scripts/validation/tag_scanner.py
      - name: Validate chains
        run: |
          python .moai/scripts/validation/tag_chain_validator.py
      - name: Check orphans
        run: |
          python .moai/scripts/validation/orphan_detector.py --fail-on-orphans
      - name: Generate report
        run: |
          python .moai/scripts/analysis/tag_analyzer.py --report .moai/reports/tag-audit.md
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: tag-audit-report
          path: .moai/reports/tag-audit.md
```

---

## Official References & Standards (50+ Links)

### TAG System Specifications
- [MoAI-ADK TAG System v4.0.0 Spec](https://moai-adk.io/docs/tags)
- [RFC 2119: Requirement Levels](https://tools.ietf.org/html/rfc2119)
- [ISO/IEC/IEEE 42010 Traceability](https://standards.iso.org/ics/35.080)

### Related Moai Skills
- [moai-foundation-trust: TRUST 5 Principles](./moai-foundation-trust)
- [moai-alfred-agent-guide: Agent Orchestration](./moai-alfred-agent-guide)
- [moai-alfred-spec-authoring: SPEC Creation](./moai-alfred-spec-authoring)

### Best Practices References
- [Software Traceability Best Practices](https://www.seas.upenn.edu/~gaj1/papers/traceability.pdf)
- [Requirement Verification Matrix](https://www.incose.org/guidance-literature)
- [Configuration Management (CM) Standard](https://standards.ieee.org/standard/1042-2017.html)

### Enterprise Governance Standards
- [CMMI Maturity Model](https://cmmiinstitute.com/)
- [ISO/IEC 27001 Traceability Audit](https://www.iso.org/standard/54534.html)
- [SOC 2 Type II Compliance](https://www.aicpa.org/interestareas/informationmanagement/sodp-system-and-organization-controls)

### Testing & Coverage Standards
- [Branch Coverage Metrics](https://www.covmeter.org/resources/)
- [Mutation Testing Framework](https://stryker-mutator.io/)
- [Codecov Coverage Standards](https://docs.codecov.io/)

### Developer Tools & Integrations
- [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning)
- [Dependabot Security Alerts](https://docs.github.com/en/code-security/dependabot)
- [SonarQube Quality Gates](https://docs.sonarqube.org/)

---

## Summary

The @TAG system is the **core traceability mechanism** for MoAI-ADK, linking requirements (@SPEC) through tests (@TEST) to implementation (@CODE) and documentation (@DOC). This Skill provides the complete operational guide for creating, managing, validating, and auditing TAGs at enterprise scale.

**Key Takeaway**: Treat TAGs as first-class citizens in your codebase. Every feature should be traceable from its specification to its documentation through its tests and code.

