---
name: "moai-foundation-tags"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: foundation
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


> **Version**: 4.0.0 Enterprise  
> **Tier**: Foundation  
> **Updated**: November 2025 Stable  

---

## Progressive Disclosure

### Level 1: Core Concepts (Quick Start)

### What It Does


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
        ↓ (depends on)
        ↓
            ↓ (validates)
        ↓
            ↓ (fulfills)
        ↓
```

### Chain Semantics (November 2025)

Each link type defines a specific relationship:

| Chain Type | Direction | Meaning | Validation |
|-----------|-----------|---------|-----------|

### Example: Complete Chain

```markdown
1. SPEC Created:
   File: .moai/specs/SPEC-001/spec.md
   Header: # SPEC-001: Authentication System
   Status: APPROVED

2. Tests Written (RED phase):
   File: tests/test_auth.py
   Content:
     """
     Validates SPEC-001 authentication requirements
     """
     def test_login_with_valid_credentials():
         assert authenticate("user", "password") == True

3. Code Implemented (GREEN phase):
   File: src/auth.py
   Content:
     def authenticate(username, password):
         if validate_password(password):
             return True

4. Documentation Written:
   File: docs/auth.md
   Header: # Authentication
   Content:
     ## User Login
     Users authenticate using username and password...

5. Traceability Chain:
```

---

## TAG Lifecycle Management

### Phase 1: Birth (Creation)

**When**: SPEC document is created  
**Where**: `.moai/specs/SPEC-XXX/spec.md`  
**Validation**: Must be unique in codebase

```markdown
---
title: SPEC-001 User Authentication
created_at: 2025-11-12
status: DRAFT
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
Validates user login with valid credentials
Coverage: Happy path
"""
def test_login_valid():
    assert authenticate("user", "password") == True
```

**CODE TAGs Created**:
```python
def authenticate(username, password):
    """
    Implements authentication logic
    """
    return verify_password(username, password)
```

**Documentation TAGs Created**:
```markdown
## Authentication
Authentication module handles user login...
```

### Phase 3: Maturity (Stability)

**All chain links validated**:
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
End-of-life: 2026-05-01
```

**Orphan Detection**:

---

## TAG Validation Rules (Enterprise v4.0)

### Mandatory Validation (STRICT Mode)

All TAGs must satisfy:

```
Rule 1: Unique Identifier
├─ TAG must be unique across entire codebase
├─ Format: @[TYPE]:[DOMAIN]-[NNN]

Rule 2: Proper Hierarchy
└─ Circular references forbidden

Rule 3: Complete Chains
└─ No orphan TAGs allowed

Rule 4: Version Compliance
├─ TAG references must use current versions
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

# Find orphans
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

```python
# Problem: Created but never used

# Solution: Either implement or deprecate
Reason: Feature cancelled
Decision: 2025-11-12
```

```python

```

```python
def feature():

# Solution: Trace requirements
```

```markdown
# Documentation exists but no source

# Solution: Link to implementation
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
            return 'OK'
        else:
            return 'ORPHAN_TYPE_A'
        if has_spec_link(tag):
            return 'OK'
        else:
            return 'ORPHAN_TYPE_B'
```

**Step 3: Remediation Phase**
```bash
# Option 1: Link existing orphan

# Option 2: Mark as deprecated

# Option 3: Create missing link
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

Quality Metrics:
├─ Orphan TAGs: 0 (0%)
├─ Broken chains: 0 (0%)
├─ Deprecated (active): 23 (1.8%)
├─ Test coverage: 96.4%
└─ Documentation complete: 89.2%

Chain Integrity:
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
```

### TAG Assignment Matrix

| Component | Responsible | Timeline | Review |
|-----------|-------------|----------|--------|

### Deprecation Policy

```
When removing features:

   EOL Date: 2026-05-01

2. Create migration guide
   Link: docs/migrations/old-to-new.md


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

v4.0.0 Pattern:

Migration:
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


## Requirements
- Send email on user signup
- Support HTML and plain text
- Rate limit to 100/minute
```

**Day 2-3: Tests Written**
```python
# tests/test_email.py

def test_send_email_on_signup():
    """Verify email sent after user registration"""
    user = create_user("test@example.com")
    assert email_sent_to("test@example.com")

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

def send_email(to, subject, body, html=False):
    """
    Send email notification
    
    References:
    """
    if rate_limiter.check(to):
        raise RateLimitError()
    
    smtp = get_smtp_connection()
    smtp.send(to, subject, body, html=html)
```

**Day 6: Documentation**
```markdown
# Email Notifications


## Overview
Send transactional emails to users...

## API Reference

### send_email(to, subject, body, html=False)
Sends an email notification...

### Rate Limiting
Rate limited to 100 emails per minute per recipient...
```

### Pattern 2: Refactoring with TAGs

**Before (Monolithic)**
```python
def authenticate(user, password, mfa=None):
    # 200 lines of validation logic
    pass
```

**After (Refactored)**
```python
def authenticate(user, password, mfa=None):
    """
    """
    if mfa:
    return True

def validate_password(user, password):
    """Extracted password validation"""
    # 50 lines

def validate_mfa(user, code):
    """Extracted MFA validation"""
    # 50 lines
```

### Pattern 3: Bug Fix with TAGs

**Issue Reported**
```
BUG-101: Password reset token expires prematurely
```

**Fix Process**
```python
title: Fix password reset token expiration

def test_token_valid_24_hours():
    """Token should remain valid for 24 hours"""
    token = generate_reset_token("user@example.com")
    assert token_valid_after(token, hours=23)
    assert not token_valid_after(token, hours=25)

def generate_reset_token(email):
    """Fixed: was using 12 hours, now 24 hours"""
    return create_token(email, expires_in_hours=24)  # Was 12

## Password Reset Token Fix
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

# 2. Find orphans
echo ""
echo "Checking for orphan TAGs..."
python << 'PYEOF'
# Implementation: Tag chain validation
import re
import subprocess


orphans = []
for spec in specs:
    # Check if any TEST, CODE, or DOC references it
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
import json
coverage_data = json.load(open('.coverage'))
for spec, coverage in coverage_data.items():
    if coverage < 0.85:
        print(f"WARNING: {spec} coverage {coverage*100:.1f}% < 85%")
PYEOF

# 5. Deprecation review
echo ""
echo "Deprecated features review:"
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

## Advanced: TAG Detection & Validation Algorithms

### 1. Missing TAG Detection Algorithm


**Detection Process**:
```python
def detect_missing_spec_tags():
    """
    """
    # 1. Find all .moai/specs/SPEC-*/spec.md files
    # 2. Read first 10 lines (header section)
```

**Fix Strategy**:
- Location: First line after title (`# SPEC-XXX`)
- Execution: Via `fix-missing-spec-tags.py --dry-run` before `--apply`

**Validation Checklist**:
- [ ] Title line exists: `# SPEC-XXX-{name}`
- [ ] TAG ID matches filename
- [ ] Format follows chain syntax

### 2. TAG Duplication Detection System


**Detection Algorithm**:
```
3. CLASSIFY:
4. RANK: By authority (topline > inline, earlier > later)
5. RECOMMEND: Keep primary, merge duplicates
```

**Unified Tool**: `tag_dedup_manager.py`
```bash
# Step 1: Scan only (no changes)
python3 .moai/scripts/validation/tag_dedup_manager.py --scan-only

# Step 2: Review plan (dry-run)
python3 .moai/scripts/validation/tag_dedup_manager.py --dry-run

# Step 3: Apply fixes
python3 .moai/scripts/validation/tag_dedup_manager.py --apply
```

**Resolution Priorities**:
1. Topline TAGs win (headers > inline)
2. Primary location wins (first occurrence)
3. Authority score wins (config > code > docs)
4. Renumber duplicates to: {type}:{domain}-{next_num}

### 3. TAG Health Monitoring Checklist

**Weekly Health Check** (Every Monday):
```python
metrics = {
    'chain_complete': 98,      # SPEC→TEST→CODE→DOC links
    'doc_sync': 92,            # SPEC updated? DOC updated?
    'definition_complete': 94, # Metadata fields filled
}
```

**Monitoring Workflow**:
```bash
# Run health check
python3 .moai/scripts/monitoring/tag_health_monitor.py --weekly

# Generate HTML dashboard
python3 .moai/scripts/monitoring/tag_health_monitor.py --html

# Send Slack alert (if configured)
python3 .moai/scripts/monitoring/tag_health_monitor.py --notify slack
```

**Health Grade Scale**:
- 🟢 95-100: Excellent (no action)
- 🟡 85-94: Good (monitor warnings)
- 🟠 75-84: Fair (improvement needed)
- 🔴 <75: Poor (immediate action required)

**Alert Triggers**:
- Orphans detected: Alert immediately
- Coverage <85%: Weekly summary
- Duplicates found: Manual review
- Chains broken: Escalate to team

### 4. 95% Traceability Verification Process

**Goal**: Ensure 95% of SPEC → TEST → CODE → DOC chains are complete

**Verification Steps**:
```
STEP 1: Coverage Analysis (5 min)
  - Count total SPEC documents

STEP 2: Chain Integrity (10 min)
  - Mark as Complete/Incomplete

STEP 3: Quality Score (5 min)
  - Complete chains / Total SPEC × 100
  - Target: ≥95%

STEP 4: Gap Analysis (5 min)
  - List missing components
  - Categorize by type
  - Assign owners for completion
```

**Execution Command**:
```bash
# Full traceability audit
python3 .moai/scripts/validation/tag_dedup_manager.py --full

# Generate audit report
python3 .moai/scripts/analysis/tag_analyzer.py --audit --report .moai/reports/tag-audit.md
```

**Pass Criteria**:
- [ ] ≥95% chains complete (SPEC→TEST→CODE→DOC)
- [ ] 0 orphan TAGs detected
- [ ] 0 duplicate TAGs found
- [ ] All TAGs have proper metadata
- [ ] Documentation synchronized

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


**Key Takeaway**: Treat TAGs as first-class citizens in your codebase. Every feature should be traceable from its specification to its documentation through its tests and code.

