# Skill Update Advisor: Analysis & Update Recommendations

This guide covers how **skill-factory** analyzes existing Skills and proposes updates based on latest information, official documentation, and best practices.

---

## Overview: Keeping Skills Current

### The Challenge: Skill Drift

Skills created 6+ months ago can become outdated:

```
Original Skill (2025-04):
├─ Framework: TypeScript 4.9
├─ Best practice: Approach X
├─ Example: Code pattern A
└─ Status: ✅ Current at that time

Today (2025-10):
├─ Framework: TypeScript 5.3 (current)
├─ Best practice: Approach Y (official recommendation changed)
├─ Example: Code pattern A (now considered legacy)
└─ Status: ⚠️ Partially outdated
```

### The Solution: Automated Analysis & Recommendations

skill-factory can **analyze existing Skill folders** and propose updates:

```
Input: Path to existing Skill directory
        ↓
skill-factory Analysis:
├─ Read all files (SKILL.md, reference.md, examples.md, etc.)
├─ Extract stated versions & frameworks
├─ Research current versions
├─ Compare against current best practices
├─ Identify deprecated features
├─ Check security implications
        ↓
Output: Update Proposal Report
├─ Issues found
├─ Recommended changes
├─ Priority level (Critical/High/Medium/Low)
├─ Before/after comparisons
├─ Migration guide (if needed)
└─ Official resources to reference
```

---

## Analysis Workflow

### Phase 1: Skill Content Extraction

**Analyze existing Skill to extract:**

```yaml
Metadata:
  name: "[Extracted from frontmatter]"
  description: "[Extracted from frontmatter]"
  last_updated: "[Determined from file content or git]"

Content Inventory:
  sections_in_skill_md: "[List of headers]"
  frameworks_mentioned: "[Detect versions: v7.2, 3.9, etc.]"
  deprecated_features: "[Detect known-deprecated patterns]"
  external_links: "[Count & validate]"

Examples:
  code_samples_found: "[Count]"
  languages_used: "[Extract: Python, TypeScript, etc.]"
  working_tested: "[Try to execute or validate syntax]"
```

### Phase 2: Version Detection & Current Status

**Identify what versions are referenced:**

```
Skill states:
  SKILL.md line 42: "Using Pytest 7.2"
  examples.md line 15: "@pytest.fixture"
  reference.md line 8: "pytest==7.2"

Current reality (WebSearch):
  Latest pytest: 8.0.0 (released 2025-10)
  Previous version: 7.4.3 (will EOL 2025-12)

Analysis result:
  Status: ⚠️ Outdated (7 months behind)
  Priority: HIGH (1 minor version behind)
  Impact: Low (basic API compatible)
```

### Phase 3: Best Practices Comparison

**Check against current best practices:**

```
Skill recommends:
├─ Pattern A (setup/teardown)
├─ Pattern B (fixture usage)
└─ Pattern C (mocking approach)

Current best practices (from WebSearch):
├─ Pattern A (outdated, Pattern A' recommended)
├─ Pattern B ✓ (still current)
└─ Pattern C (replaced by Pattern C++)

Findings:
  🔴 Pattern A: DEPRECATED
     Current recommendation: Pattern A'
     Official source: https://docs.pytest.org/...

  🟢 Pattern B: CURRENT
     Still recommended
     No changes needed

  🟡 Pattern C: ENHANCED
     Pattern C still works but Pattern C++ is better
     C++ has better error handling, security
     Upgrade recommended but not critical
```

### Phase 4: Security & Deprecation Audit

**Search for security & deprecation issues:**

```
Analysis:

Security Checks:
├─ Credentials hardcoded? (RED FLAG)
├─ Outdated dependencies? (WARN)
├─ Known vulnerabilities? (URGENT)
└─ Insecure practices? (WARN)

Deprecation Audit:
├─ Deprecated syntax used?
├─ Removed in new version?
├─ Legacy patterns recommended?
└─ EOL frameworks mentioned?

Breaking Changes:
├─ API changes?
├─ Behavior changes?
├─ Removed features?
└─ Parameter changes?
```

### Phase 5: Generate Update Report

**Produce structured update proposal:**

```markdown
# Skill Update Proposal: [Skill Name]

## Summary

Status: ⚠️ UPDATE RECOMMENDED
Analyzed: 2025-10-22
Last Updated: 2025-05-15 (5 months old)
Overall Priority: HIGH

## Version Analysis

| Component | Current | Latest | Status | Update |
|-----------|---------|--------|--------|--------|
| Framework | 7.2 | 8.0 | 🔴 Critical | Recommend |
| Python | 3.9+ | 3.12 | 🟡 Warning | Optional |
| Dependency X | 2.1 | 3.0 | 🟡 Warning | Check compat |

## Issues Found

### 🔴 Critical (Update Required)
1. Pytest version outdated
   - Current: 7.2 (EOL 2025-12)
   - Recommended: 8.0
   - Impact: Performance +30%, new features
   - Effort: LOW (API compatible)

### 🟡 Warning (Recommended)
1. Python 3.9 approaching EOL
   - EOL date: Oct 2025
   - Recommended: Python 3.11+
   - Effort: MEDIUM (update examples)

### 🟢 Good (No action needed)
1. Best practices still current
2. Examples follow conventions
3. Security practices sound

## Files Requiring Updates

1. SKILL.md (420 → 500 lines)
   - Update framework version
   - Add new features section
   - Update examples

2. reference.md (300 → 370 lines)
   - API changes documented
   - Deprecations noted
   - Migration guide added

3. examples.md (200 → 280 lines)
   - Modernize code samples
   - Add new patterns
   - Update dependencies

## Specific Recommendations

### Update 1: Framework Version
Location: SKILL.md line 42, examples.md lines 15-28
Change: pytest==7.2 → pytest==8.0
Justification: Current stable, 30% performance improvement
Breaking changes: None for basic usage
Migration effort: <1 hour

### Update 2: Async Testing Support
Location: SKILL.md (NEW SECTION)
Add: pytest-asyncio patterns
Justification: Modern Python uses async/await
User impact: NEW capability, no breakage
Effort: NEW content (~100 lines)

### Update 3: Type Hints
Location: All code examples
Change: Add type annotations
Example: def test_foo(client: pytest.Client) -> None:
Justification: Official recommendation, better IDE support
Effort: MEDIUM (refactor examples)

## Migration Path for Users

If users are on Pytest 7.2:

1. Pytest 7.2 → 8.0 migration
   Time: ~15 minutes
   Breaking changes: None for basic fixtures
   New features available: Full async support

2. Step-by-step guide
   - Step 1: pip install --upgrade pytest
   - Step 2: Run existing tests (should pass)
   - Step 3: Adopt new async patterns (optional)

3. Full migration guide
   - Detailed in MIGRATION-GUIDE.md (NEW)

## Official Resources

- Pytest 8.0 Release: https://docs.pytest.org/8.0/release.html
- Async Testing: https://docs.pytest.org/asyncio.html
- Python 3.12 Support: https://python.org/dev/peps/pep-0715
- Migration Guide: https://docs.pytest.org/migration.html

## Generated Update Preview

[Show before/after code snippets for reviewers]

## Approval Workflow

Ready for update?
  ☐ Review all changes
  ☐ Test migration guide
  ☐ Update dates & versions
  ☐ Get peer review
  → Apply updates (with --apply flag)
```

---

## Analyzing Different Skill Types

### Type 1: Language/Framework Basics

```
Skill: "Testing with Python"

Analysis focuses on:
├─ Python version (3.9, 3.10, 3.11, 3.12?)
├─ Testing framework version (pytest X.Y.Z)
├─ Core library versions (unittest, mock, etc.)
├─ Best practices (fixtures, parametrization, etc.)
└─ Deprecations (nose, testtools, etc.)

Update priority:
├─ Version updates: HIGH (affects compatibility)
├─ Best practice updates: MEDIUM
├─ New patterns: LOW (backward compatible)
```

### Type 2: Advanced Patterns

```
Skill: "Performance Optimization Patterns"

Analysis focuses on:
├─ Timing/profiling tool versions
├─ Algorithm recommendations (newer = better?)
├─ Hardware considerations (changed?)
├─ Benchmark data (outdated?)
└─ Performance regressions

Update priority:
├─ Tool updates: MEDIUM
├─ Algorithmic advances: MEDIUM
├─ Benchmark data: HIGH (can be misleading if stale)
```

### Type 3: Infrastructure/DevOps

```
Skill: "Container Deployment with Docker"

Analysis focuses on:
├─ Docker version (currently 24.x+)
├─ Base images (Alpine? Ubuntu? Security patches?)
├─ Orchestration tools (Kubernetes 1.29+?)
├─ Security practices (latest vulnerabilities?)
└─ Compatibility (breaking changes?)

Update priority:
├─ Security patches: CRITICAL
├─ Version updates: HIGH
├─ Tool alternatives: MEDIUM
```

---

## Implementation: Analysis Commands

### Command Pattern

```bash
# Basic analysis
/alfred:update-skill /path/to/skill-name

# Generate report only (don't apply changes)
/alfred:update-skill --report moai-skill-testing

# Show detailed analysis
/alfred:update-skill --analyze moai-skill-testing

# Compare before/after
/alfred:update-skill --diff moai-skill-testing

# Apply recommended updates (requires approval)
/alfred:update-skill --apply moai-skill-testing

# Dry run (show what would be changed)
/alfred:update-skill --dry-run moai-skill-testing

# Force re-analysis (bypass cache)
/alfred:update-skill --force moai-skill-testing

# Generate migration guide for users
/alfred:update-skill --migration moai-skill-testing
```

---

## Automated Analysis Checks

### Version Scanning

```python
def scan_versions(skill_content):
    """Extract and validate framework versions"""

    patterns_found = {
        'pytest': ['7.2', '==7.2'],
        'python': ['3.9', '3.9+'],
        'typescript': ['4.9', 'typescript@4.9'],
    }

    for framework, versions in patterns_found.items():
        current = websearch(f"{framework} latest version 2025")
        if current != versions[-1]:
            flag_as_outdated(framework, versions[-1], current)
```

### Deprecation Checking

```python
def check_deprecations(skill_content):
    """Find deprecated patterns"""

    deprecated_patterns = {
        'pytest': ['class-based fixtures', 'nose integration'],
        'python': ['collections.abc', 'typing.Dict'],
    }

    for deprecated in deprecated_patterns:
        if deprecated in skill_content:
            flag_as_deprecated(deprecated)
```

### Security Audit

```python
def audit_security(skill_content):
    """Check for security issues"""

    security_patterns = {
        'secrets_hardcoded': [r'password.*=.*"', r'api_key.*=.*"'],
        'insecure_code': [r'eval\(', r'pickle\.loads'],
        'outdated_crypto': [r'md5', r'sha1', r'DES'],
    }

    for issue_type, patterns in security_patterns.items():
        if pattern_found in skill_content:
            flag_security_issue(issue_type)
```

### Link Validation

```python
def validate_links(skill_content):
    """Check all external links are working"""

    links_found = extract_urls(skill_content)

    for url in links_found:
        status = webfetch(url, check_only=True)
        if status != 200:
            flag_broken_link(url, status)
```

---

## Update Proposal Examples

### Example 1: Minor Version Update

```
Skill: moai-skill-logging-python
Current: Python logging module (3.9 base)
Latest: Python logging (3.12 base)
Impact: MINIMAL

Recommendation:
  Priority: LOW
  Justification: Basic logging API unchanged
  Update: Update example Python version requirement only
  Effort: 5 minutes
  Risk: Very low
```

### Example 2: Major Framework Update

```
Skill: moai-skill-web-django
Current: Django 3.2 LTS
Latest: Django 5.0 (released 2025-09)
Impact: SIGNIFICANT

Recommendation:
  Priority: HIGH
  Justification: New major version with new features
  Updates needed:
    1. Update dependency version
    2. New async views patterns
    3. New security features
    4. Breaking changes documented
  Effort: 2-3 hours
  Risk: Requires testing
  Migration guide: Essential
```

### Example 3: Best Practice Shift

```
Skill: moai-skill-error-handling-js
Current: Error handling with try/catch
Latest: Error handling with Promise chains AND try/await
Impact: MODERATE

Recommendation:
  Priority: MEDIUM
  Justification: Community evolved best practice
  Updates:
    1. Keep current patterns (still valid)
    2. Add modern async/await approach
    3. Show when to use each
  Effort: 1 hour
  Risk: Low (new content, not breaking)
  Migration: Not needed (backward compatible)
```

---

## Preventing Update Fatigue

### Update Cadence

```
Framework Type:
├─ Stable/LTS: Check quarterly (3-4 times/year)
├─ Regular release: Check monthly (12 times/year)
├─ Rapid development: Check biweekly (24 times/year)
└─ Cutting edge: Check weekly (52 times/year)
```

### Priority Filtering

```
Only flag critical updates:
├─ 🔴 Security issues (ALWAYS)
├─ 🔴 Breaking changes (ALWAYS)
├─ 🟠 Version EOL (IF actively used)
├─ 🟡 New best practices (IF significant)
└─ 🟢 Minor updates (BATCH quarterly)
```

### Grouped Updates

```
Don't update every small change:
├─ Batch 3-4 minor updates together
├─ Wait for quarterly review window
├─ Combine with community feedback

But DO update immediately if:
├─ Security vulnerability discovered
├─ Breaking change announced
├─ Official API changed
```

---

## Skill Update Metadata

Add to each Skill for tracking:

```yaml
# Update Tracking

last_analysis: 2025-10-22
next_analysis: 2025-11-22 (monthly)
framework_version: "Pytest 8.0.0"
framework_version_date: 2025-10-15

update_history:
  - date: 2025-10-22
    type: "version update"
    from: "Pytest 7.2 → 8.0"
    priority: "HIGH"
    effort: "LOW"
    status: "approved"

  - date: 2025-09-15
    type: "best practice update"
    changes: "Added async testing patterns"
    priority: "MEDIUM"
    status: "completed"

security_review: 2025-10-22 (No issues)
deprecation_review: 2025-10-22 (No deprecations found)
```

---

## Related Resources

- [SKILL.md](SKILL.md) — Main Skill framework
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Research techniques
- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — Discovery process
- [skill-factory.md](../../agents/alfred/skill-factory.md) — Orchestration

---

**Version**: 0.1.0
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Skills + skill-factory
