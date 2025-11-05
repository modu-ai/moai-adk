# TAG System: Complete Traceability Guide

Master the @TAG system that creates complete traceability from specifications through implementation, testing, and documentation. This guide covers TAG creation, management, validation, and best practices.

## What is the @TAG System?

The @TAG system is MoAI-ADK's traceability mechanism that links every project artifact through unique identifiers. It ensures that requirements, tests, code, and documentation remain connected throughout the development lifecycle.

### Why @TAGs Matter

**Traditional Development Problems**:
- "Why was this code written?" → Lost context, forgotten requirements
- "What tests cover this feature?" → Incomplete test coverage discovery
- "Where is the documentation for this?" → Scattered or outdated docs
- "What code needs to change for this requirement update?" → Manual impact analysis

**@TAG System Solutions**:
- **Complete Traceability**: Every artifact is linked to its source
- **Impact Analysis**: Instant identification of affected code
- **Living Documentation**: Docs stay synchronized with code
- **Quality Assurance**: No orphaned code or missing tests

### @TAG Chain Concept

```
@SPEC:DOMAIN-001 (Requirements)
    ↓ Defines what to build
@TEST:DOMAIN-001 (Tests)
    ↓ Validates implementation
@CODE:DOMAIN-001:SUBTYPE (Implementation)
    ↓ Creates the solution
@DOC:DOMAIN-001 (Documentation)
    ↓ Explains the solution
```

## @TAG Format and Structure

### Basic Format

**Standard Format**: `@TYPE:DOMAIN-ID[:SUBTYPE]`

**Components**:
- **`@`**: TAG indicator (required)
- **`TYPE`**: Artifact type (SPEC, TEST, CODE, DOC)
- **`DOMAIN`**: Functional area (AUTH, USER, API, etc.)
- **`ID`**: Sequential number (001, 002, 003...)
- **`SUBTYPE`**: Optional sub-classification (MODEL, SERVICE, API, etc.)

### Type Definitions

| Type | Purpose | Examples |
|------|---------|----------|
| **SPEC** | Requirements and specifications | `@SPEC:AUTH-001` |
| **TEST** | Test cases and test suites | `@TEST:AUTH-001` |
| **CODE** | Implementation code | `@CODE:AUTH-001:SERVICE` |
| **DOC** | Documentation and guides | `@DOC:AUTH-001` |

### Domain Categories

| Domain | Description | Examples |
|--------|-------------|----------|
| **AUTH** | Authentication & authorization | `@SPEC:AUTH-001` |
| **USER** | User management & profiles | `@SPEC:USER-001` |
| **API** | REST APIs & interfaces | `@SPEC:API-001` |
| **DB** | Database & persistence | `@SPEC:DB-001` |
| **UI** | User interface & components | `@SPEC:UI-001` |
| **SEC** | Security & compliance | `@SPEC:SEC-001` |
| **PERF** | Performance & optimization | `@SPEC:PERF-001` |
| **INT** | Integrations & external systems | `@SPEC:INT-001` |
| **CONFIG** | Configuration & settings | `@SPEC:CONFIG-001` |

### Subtype Classifications

#### CODE Subtypes

| Subtype | When to Use | Examples |
|---------|-------------|----------|
| **MODEL** | Data models, schemas, classes | `@CODE:USER-001:MODEL` |
| **SERVICE** | Business logic, services | `@CODE:AUTH-001:SERVICE` |
| **API** | HTTP endpoints, controllers | `@CODE:API-001:ENDPOINT` |
| **REPO** | Repository pattern, data access | `@CODE:DB-001:REPO` |
| **UTILS** | Utility functions, helpers | `@CODE:AUTH-001:UTILS` |
| **CONFIG** | Configuration classes | `@CODE:CONFIG-001:SETTINGS` |
| **MIDDLEWARE** | Middleware, interceptors | `@CODE:API-001:MIDDLEWARE` |
| **VALIDATOR** | Validation logic | `@CODE:USER-001:VALIDATOR` |

#### TEST Subtypes

| Subtype | When to Use | Examples |
|---------|-------------|----------|
| **UNIT** | Unit tests | `@TEST:AUTH-001:UNIT` |
| **INTEGRATION** | Integration tests | `@TEST:API-001:INTEGRATION` |
| **E2E** | End-to-end tests | `@TEST:USER-001:E2E` |
| **PERFORMANCE** | Performance tests | `@TEST:API-001:PERF` |
| **SECURITY** | Security tests | `@TEST:AUTH-001:SECURITY` |

#### DOC Subtypes

| Subtype | When to Use | Examples |
|---------|-------------|----------|
| **API** | API documentation | `@DOC:API-001:API` |
| **GUIDE** | User guides, tutorials | `@DOC:USER-001:GUIDE` |
| **REFERENCE** | Technical reference | `@DOC:AUTH-001:REFERENCE` |
| **DEPLOYMENT** | Deployment guides | `@DOC:INT-001:DEPLOYMENT` |

## @TAG Creation and Assignment

### Automatic TAG Assignment

Alfred automatically assigns TAGs during the development workflow:

```bash
# Phase 1: SPEC creation
/alfred:1-plan "User authentication system"
# Alfred assigns: @SPEC:AUTH-001

# Phase 2: TDD implementation
/alfred:2-run AUTH-001
# Alfred assigns: @TEST:AUTH-001, @CODE:AUTH-001:*

# Phase 3: Documentation sync
/alfred:3-sync
# Alfred assigns: @DOC:AUTH-001
```

### Manual TAG Assignment

When creating files manually:

```python
# src/auth/service.py
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

class AuthService:
    """Authentication service implementation"""
    pass
```

```python
# tests/test_auth.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_login_with_valid_credentials():
    """Test user authentication with valid credentials"""
    pass
```

### TAG Assignment Best Practices

#### 1. Consistency

**✅ Good**:
```python
# All related code uses same DOMAIN-ID
@CODE:AUTH-001:MODEL
@CODE:AUTH-001:SERVICE
@CODE:AUTH-001:API
@CODE:AUTH-001:UTILS
```

**❌ Bad**:
```python
# Inconsistent DOMAIN-ID usage
@CODE:AUTH-001:MODEL
@CODE:AUTH-002:SERVICE  # Wrong ID
@CODE:USER-001:API     # Wrong domain
```

#### 2. Specificity

**✅ Good**:
```python
# Specific subtypes for clear organization
@CODE:AUTH-001:SERVICE
@CODE:AUTH-001:MODEL
@CODE:AUTH-001:VALIDATOR
```

**❌ Bad**:
```python
# Too generic - doesn't indicate file purpose
@CODE:AUTH-001
@CODE:AUTH-001
@CODE:AUTH-001
```

#### 3. Traceability Links

**✅ Good**:
```python
# Include links to related artifacts
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py
```

**❌ Bad**:
```python
# Missing traceability information
# @CODE:AUTH-001:SERVICE
```

## @TAG Chain Management

### Complete Chain Example

```markdown
# SPEC Document
# .moai/specs/SPEC-AUTH-001/spec.md
# @SPEC:EX-AUTH-001: User Authentication System

## Requirements
- The system SHALL provide user authentication...
```

```python
# Test File
# tests/test_auth.py
# @TEST:EX-AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_login_with_valid_credentials():
    """Test authentication with valid credentials"""
    pass
```

```python
# Implementation Files
# src/auth/models.py
# @CODE:EX-AUTH-001:MODEL | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

class User:
    """User data model"""
    pass

# src/auth/service.py
# @CODE:EX-AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

class AuthService:
    """Authentication business logic"""
    pass

# src/auth/api.py
# @CODE:EX-AUTH-001:API | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

@app.post("/auth/login")
def login():
    """Login endpoint"""
    pass
```

```markdown
# Documentation
# docs/api/auth.md
# @DOC:EX-AUTH-001 | SPEC: SPEC-AUTH-001.md

# Authentication API Documentation
...
```

### Chain Validation

Alfred automatically validates TAG chains:

```bash
/alfred:3-sync
```

**Output Example**:
```
🔍 TAG Chain Validation Report...

✅ Complete Chain: AUTH-001
   @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md
   @TEST:AUTH-001 → tests/test_auth.py (5 test functions)
   @CODE:AUTH-001:MODEL → src/auth/models.py (2 classes)
   @CODE:AUTH-001:SERVICE → src/auth/service.py (1 class)
   @CODE:AUTH-001:API → src/auth/api.py (1 endpoint)
   @DOC:AUTH-001 → docs/api/auth.md (complete API docs)

📊 Chain Integrity: 100%
🔗 Orphaned TAGs: 0
⚠️  Missing References: 0
```

### Orphaned TAG Detection

Alfred identifies and helps fix orphaned TAGs:

```bash
⚠️ Orphaned TAGs Detected:

Orphaned @CODE:AUTH-001:VALIDATOR in src/auth/validators.py
   Missing @TEST:AUTH-001:VALIDATOR
   Recommendation: Create tests for validator functions

Orphaned @TEST:AUTH-002 in tests/test_auth_advanced.py
   Missing @SPEC:AUTH-002
   Recommendation: Create specification document

🔧 Auto-fix Options:
[1] Create missing test file for @CODE:AUTH-001:VALIDATOR
[2] Create specification for @TEST:AUTH-002
[3] Manual review required
[4] Ignore warnings (not recommended)
```

## @TAG Search and Navigation

### Finding Related Artifacts

#### Basic Search

```bash
# Find all artifacts for AUTH-001
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# Find all SPECs
rg '@SPEC:' -n

# Find all CODE files
rg '@CODE:' -n
```

#### Advanced Search Patterns

```bash
# Find SPEC and related tests
rg '@SPEC:AUTH-001' -A 5 -B 5

# Find orphaned CODE (no matching TEST)
rg '@CODE:AUTH-001' --files-with-matches | while read file; do
  if ! rg -q "@TEST:AUTH-001" "$(dirname $file)/test*"; then
    echo "Orphaned CODE in: $file"
  fi
done

# Find all chains for a domain
rg '@(SPEC|TEST|CODE|DOC):AUTH-\d+' -n
```

#### Impact Analysis

```bash
# Find everything affected by SPEC change
rg '@SPEC:AUTH-001' -n
# → Shows: spec, tests, code, documentation

# Find test coverage for a feature
rg '@TEST:AUTH-001' -n
# → Shows all test files covering AUTH-001

# Find implementation status
rg '@CODE:AUTH-001' -n
# → Shows all implementation files
```

### Navigation Shortcuts

#### VS Code Integration

Create tasks in `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Find SPEC chain",
            "type": "shell",
            "command": "rg",
            "args": ["'@(SPEC|TEST|CODE|DOC):${input:domain}'", "-n"],
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ],
    "inputs": [
        {
            "id": "domain",
            "description": "Domain (e.g., AUTH-001)",
            "default": "AUTH-001",
            "type": "promptString"
        }
    ]
}
```

#### Git Aliases

Add to `.gitconfig`:

```bash
[alias]
    find-chain = "!rg '@(SPEC|TEST|CODE|DOC):' -n"
    spec-chain = "!rg '@(SPEC|TEST|CODE|DOC):' -l | xargs grep -l"
    orphaned-tags = "!rg '@CODE:' --files-with-matches | while read f; do tag=$(grep -o '@CODE:[^:]+' \"$f\"); if ! rg -q \"${tag/:CODE:/@TEST:}\" .; then echo \"Orphaned: $f ($tag)\"; fi; done"
```

## @TAG Best Practices

### 1. Consistent Formatting

**Use Standard Format**:
```python
# ✅ Correct format
@CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

# ❌ Incorrect formats
@code:auth-001:service  # Wrong case
@CODE:auth-1:SERVICE    # Wrong format
@CODE:AUTH-001         # Missing subtype
```

### 2. Complete Traceability

**Link All Related Artifacts**:
```python
# ✅ Complete traceability
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

# ❌ Missing links
# @CODE:AUTH-001:SERVICE
```

### 3. Logical Organization

**Group Related Code**:
```python
# ✅ Logical grouping
src/
├── auth/
│   ├── models.py      # @CODE:AUTH-001:MODEL
│   ├── service.py     # @CODE:AUTH-001:SERVICE
│   ├── api.py         # @CODE:AUTH-001:API
│   └── utils.py       # @CODE:AUTH-001:UTILS

# ❌ Random organization
src/
├── models.py         # @CODE:AUTH-001:MODEL
├── auth_service.py   # @CODE:AUTH-001:SERVICE
├── login_api.py      # @CODE:AUTH-001:API
└── helpers.py        # @CODE:AUTH-001:UTILS
```

### 4. Appropriate Granularity

**Right-sized Components**:
```python
# ✅ Appropriate granularity
@CODE:AUTH-001:MODEL     # User, Session models
@CODE:AUTH-001:SERVICE    # AuthService class
@CODE:AUTH-001:API        # Login endpoint

# ❌ Too granular
@CODE:AUTH-001:MODEL:USER     # User model only
@CODE:AUTH-001:MODEL:SESSION   # Session model only

# ❌ Too broad
@CODE:AUTH-001               # Everything in one file
```

### 5. Regular Maintenance

**Keep Chains Updated**:
```bash
# Regular validation
/alfred:3-sync

# Manual check
rg '@(SPEC|TEST|CODE|DOC):' -n | sort | uniq -c

# Find and fix orphans
moai-adk find-orphans
```

## @TAG in Different File Types

### Python Files

```python
# src/auth/service.py
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

"""
Authentication service implementation.

This file contains the business logic for user authentication,
including password validation, token generation, and session management.

Related files:
- Models: @CODE:AUTH-001:MODEL
- API endpoints: @CODE:AUTH-001:API
- Tests: @TEST:AUTH-001
"""

class AuthService:
    """@CODE:AUTH-001:SERVICE - Main authentication service"""

    def authenticate(self, credentials):
        """Authenticate user credentials"""
        # Implementation here
        pass
```

### JavaScript/TypeScript Files

```javascript
// src/auth/service.js
// @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/auth.test.js

/**
 * Authentication service
 *
 * Handles user authentication, token management, and session handling.
 *
 * Related files:
 * - Models: @CODE:AUTH-001:MODEL
 * - API routes: @CODE:AUTH-001:API
 * - Tests: @TEST:AUTH-001
 */

class AuthService {
  /**
   * Authenticate user credentials
   * @CODE:AUTH-001:SERVICE:METHOD
   */
  async authenticate(credentials) {
    // Implementation
  }
}
```

### Test Files

```python
# tests/test_auth.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

"""
Authentication system tests.

Tests cover:
- Login with valid credentials
- Login with invalid credentials
- Token generation and validation
- Session management

Related files:
- Specification: @SPEC:AUTH-001
- Implementation: @CODE:AUTH-001:*
- Documentation: @DOC:AUTH-001
"""

class TestAuthService:
    """Test cases for @CODE:AUTH-001:SERVICE"""

    def test_login_with_valid_credentials(self):
        """Test: @SPEC:AUTH-001 - Valid credentials should authenticate"""
        # Test implementation
        pass
```

### Documentation Files

```markdown
# docs/api/auth.md
# @DOC:AUTH-001 | SPEC: SPEC-AUTH-001.md

# Authentication API Documentation

This document describes the authentication API endpoints,
including request/response formats, authentication methods,
and security considerations.

## Related Artifacts
- Specification: @SPEC:AUTH-001
- Tests: @TEST:AUTH-001
- Implementation: @CODE:AUTH-001:*
```

## @TAG Automation and Tooling

### Git Hooks

Automated TAG validation in Git hooks:

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "🔍 Validating TAG chains..."

# Check for missing TAGs
missing_tags=$(rg -L '@(SPEC|TEST|CODE|DOC):' --files-with-matching src/ tests/ docs/)

if [ -n "$missing_tags" ]; then
    echo "❌ Files missing TAGs:"
    echo "$missing_tags"
    exit 1
fi

# Check for orphaned TAGs
orphans=$(moai-adk find-orphans)

if [ -n "$orphans" ]; then
    echo "⚠️ Orphaned TAGs detected:"
    echo "$orphans"
    echo "Consider fixing these before committing."
fi

echo "✅ TAG validation passed"
```

### CI/CD Integration

GitHub Actions workflow:

```yaml
# .github/workflows/tag-validation.yml
name: TAG Chain Validation

on: [push, pull_request]

jobs:
  validate-tags:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Setup Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.13'

    - name: Install MoAI-ADK
      run: pip install moai-adk

    - name: Validate TAG chains
      run: |
        moai-adk validate-tags
        moai-adk check-orphans

    - name: Generate TAG report
      run: |
        moai-adk tag-report > tag-report.md

    - name: Upload TAG report
      uses: actions/upload-artifact@v3
      with:
        name: tag-report
        path: tag-report.md
```

### VS Code Extensions

Recommended extensions for TAG management:

1. **TAG Highlighter**: Custom syntax highlighting for @TAGs
2. **TAG Navigator**: Quick navigation between related artifacts
3. **TAG Validator**: Real-time TAG validation and suggestions

### Custom Scripts

TAG management utilities:

```python
#!/usr/bin/env python3
# scripts/tag_manager.py

import re
import os
import sys
from pathlib import Path

class TagManager:
    def __init__(self, project_root):
        self.project_root = Path(project_root)

    def find_all_tags(self):
        """Find all @TAGs in the project"""
        pattern = r'@(SPEC|TEST|CODE|DOC):([A-Z]+-\d+)(?::([A-Z]+))?'
        tags = []

        for file_path in self.project_root.rglob('*'):
            if file_path.is_file() and file_path.suffix in ['.py', '.js', '.md', '.ts']:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    matches = re.finditer(pattern, content)
                    for match in matches:
                        tags.append({
                            'file': str(file_path),
                            'type': match.group(1),
                            'domain': match.group(2),
                            'subtype': match.group(3),
                            'full_tag': match.group(0)
                        })

        return tags

    def validate_chains(self):
        """Validate TAG chain completeness"""
        tags = self.find_all_tags()
        chains = {}

        # Group by domain
        for tag in tags:
            domain = tag['domain']
            if domain not in chains:
                chains[domain] = {'SPEC': False, 'TEST': False, 'CODE': False, 'DOC': False}
            chains[domain][tag['type']] = True

        # Find incomplete chains
        incomplete = []
        for domain, types in chains.items():
            if not all(types.values()):
                missing = [t for t, present in types.items() if not present]
                incomplete.append({'domain': domain, 'missing': missing})

        return incomplete

    def find_orphans(self):
        """Find orphaned TAGs"""
        # Implementation for finding orphaned TAGs
        pass

if __name__ == '__main__':
    manager = TagManager('.')

    if len(sys.argv) > 1:
        command = sys.argv[1]

        if command == 'validate':
            incomplete = manager.validate_chains()
            if incomplete:
                print("❌ Incomplete TAG chains:")
                for item in incomplete:
                    print(f"  {item['domain']}: missing {', '.join(item['missing'])}")
            else:
                print("✅ All TAG chains complete")

        elif command == 'list':
            tags = manager.find_all_tags()
            for tag in tags:
                print(f"{tag['file']}: {tag['full_tag']}")
```

## Troubleshooting @TAG Issues

### Common Problems

#### 1. Missing TAGs

**Symptoms**:
- Files not appearing in search results
- Incomplete traceability chains

**Solutions**:
```bash
# Find files without TAGs
find src/ tests/ docs/ -type f \( -name "*.py" -o -name "*.js" -o -name "*.md" \) -exec grep -L '@(SPEC|TEST|CODE|DOC):' {} \;

# Add missing TAGs manually
# Or use Alfred to fix automatically
/alfred:3-sync --fix-tags
```

#### 2. Incorrect TAG Format

**Symptoms**:
- TAGs not being recognized
- Validation errors

**Common Format Errors**:
```python
# ❌ Wrong case
@code:auth-001:service

# ❌ Wrong format
@CODE-AUTH-001-SERVICE

# ❌ Missing parts
@CODE:AUTH-001
```

**Solutions**:
```bash
# Find and fix format issues
rg '@[a-zA-Z]+:[a-zA-Z]+-\d+' --files-with-matching

# Use automated fixing
moai-adk fix-tag-format
```

#### 3. Orphaned TAGs

**Symptoms**:
- Broken chains in validation reports
- Missing related artifacts

**Solutions**:
```bash
# Find orphaned CODE (no matching TEST)
rg '@CODE:[A-Z]+-\d+' --files-with-matching | while read file; do
    domain=$(grep -o '@CODE:[A-Z]+-\d+' "$file" | head -1)
    test_domain="${domain/@CODE/@TEST}"
    if ! rg -q "$test_domain" tests/; then
        echo "Orphaned CODE: $file ($domain)"
    fi
done

# Auto-fix with Alfred
/alfred:3-sync --auto-fix-orphans
```

#### 4. Duplicate TAGs

**Symptoms**:
- Confusion in traceability
- Multiple files with same TAG

**Solutions**:
```bash
# Find duplicate TAGs
rg '@CODE:[A-Z]+-\d+' --files-with-matching | sort | uniq -d

# Reassign unique TAGs
moai-adk reassign-tags --domain=AUTH
```

### Recovery Procedures

#### 1. Restore from Git History

```bash
# Find last known good state
git log --oneline --grep="sync" | head -5

# Restore TAGs from previous commit
git checkout HEAD~1 -- src/ tests/ docs/

# Run validation
/alfred:3-sync
```

#### 2. Rebuild TAG Chains

```bash
# Rebuild all TAGs from SPEC files
moai-adk rebuild-tags --from-specs

# Validate rebuilt chains
moai-adk validate-tags --strict
```

#### 3. Manual Recovery

```bash
# Export current TAG state
moai-adk export-tags > tags-backup.json

# Manual cleanup and fix
# ... manual editing ...

# Import corrected TAGs
moai-adk import-tags tags-fixed.json
```

## Advanced @TAG Techniques

### 1. Hierarchical TAGs

For complex features, use hierarchical relationships:

```python
# Core authentication
@SPEC:AUTH-001
@CODE:AUTH-001:SERVICE

# Password reset (related to auth)
@SPEC:AUTH-002
@CODE:AUTH-002:SERVICE

# Link related features
# In AUTH-002 SPEC:
## Dependencies
- @SPEC:AUTH-001 (Core authentication system)
```

### 2. Cross-Domain TAGs

For features spanning multiple domains:

```python
# User profile with authentication
@CODE:USER-001:PROFILE | SPEC: SPEC-USER-001.md, SPEC-AUTH-001.md

# API with security concerns
@CODE:API-001:ENDPOINT | SPEC: SPEC-API-001.md, SPEC-SEC-001.md
```

### 3. Version TAGs

For tracking feature evolution:

```python
# Version 1 implementation
@CODE:AUTH-001:SERVICE_V1

# Version 2 implementation
@CODE:AUTH-001:SERVICE_V2

# In SPEC:
## Implementation History
- V1: @CODE:AUTH-001:SERVICE_V1 (deprecated)
- V2: @CODE:AUTH-001:SERVICE_V2 (current)
```

### 4. Environment-Specific TAGs

For environment-specific implementations:

```python
# Production implementation
@CODE:AUTH-001:SERVICE_PROD

# Development implementation
@CODE:AUTH-001:SERVICE_DEV

# Testing implementation
@CODE:AUTH-001:SERVICE_TEST
```

## @TAG Analytics and Reporting

### Chain Completeness Report

```bash
# Generate comprehensive TAG report
moai-adk tag-report --format=html > tag-report.html
```

**Report Contents**:
- Overall chain completeness statistics
- Domain-specific analysis
- Orphaned TAG identification
- Historical trends
- Recommendations for improvement

### Coverage Analysis

```python
# scripts/coverage_analysis.py

def analyze_coverage():
    """Analyze test coverage by TAG"""
    tags = find_all_tags()

    for domain in get_unique_domains(tags):
        spec_count = count_tags(tags, 'SPEC', domain)
        test_count = count_tags(tags, 'TEST', domain)

        coverage = (test_count / spec_count) * 100 if spec_count > 0 else 0

        print(f"{domain}: {test_count}/{spec_count} specs covered ({coverage:.1f}%)")
```

### Quality Metrics

Track TAG quality over time:

```bash
# Historical TAG quality
moai-adk tag-quality --since="2025-01-01" --format=trend

# Domain comparison
moai-adk tag-quality --by-domain --format=comparison
```

## Summary

The @TAG system is the backbone of MoAI-ADK's traceability and quality assurance. By maintaining complete chains from specifications through implementation, testing, and documentation, you create a development environment where:

- **Nothing gets lost** - All code is traceable to requirements
- **Impact analysis is instant** - Know exactly what to change when requirements evolve
- **Quality is assured** - No orphaned code or missing tests
- **Documentation stays current** - Automatic synchronization prevents drift

Master the @TAG system, and you'll experience a new level of confidence and control in your software development process! 🎯