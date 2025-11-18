# Skill Factory Examples

> Real-world skill creation scenarios

_Last updated: 2025-10-22_

---

## Example 1: Creating a Python Testing Skill (Complete Workflow)

### Phase 1: Discovery (15 min)

**Problem Statement**:
"Developers need guidance on Python testing best practices with pytest, coverage, and fixtures."

**Trigger Keywords**:
`python`, `pytest`, `testing`, `fixtures`, `coverage`, `unittest`, `tdd`

**Reference URLs**:
- https://docs.pytest.org/en/stable/
- https://coverage.readthedocs.io/
- https://realpython.com/pytest-python-testing/

**Target Tier**: Language (moai-lang-python-testing)

---

### Phase 2: Design (20 min)

**Metadata**:
```yaml
---
name: Testing Python Applications with pytest
description: Write and run Python tests using pytest, manage fixtures, measure coverage, and integrate with CI/CD. Use when testing Python code, setting up test environments, or when the user mentions pytest, coverage, fixtures, or TDD.
allowed-tools: Read, Bash
---
```

**Content Outline**:
```
1. Introduction (High Freedom - 10%)
   - Why pytest over unittest
   - Testing philosophy

2. Core Concepts (Medium Freedom - 40%)
   - Test discovery patterns
   - Fixture usage
   - Parametrization
   - Mocking strategies

3. Examples (Medium Freedom - 30%)
   - Simple unit test
   - Fixture example
   - Parametrized test
   - Integration test

4. Coverage & CI (Low Freedom - 20%)
   - Coverage command
   - CI configuration (GitHub Actions)
```

---

### Phase 3: Production (60 min)

**SKILL.md Excerpt**:

```markdown
# Testing Python Applications with pytest

## Quick Start

```bash
# Install pytest
pip install pytest pytest-cov

# Run tests
pytest

# With coverage
pytest --cov=src --cov-report=term-missing
```

## Core Concepts

### Test Discovery

pytest automatically discovers tests matching these patterns:
- Files: `test_*.py` or `*_test.py`
- Functions: `test_*`
- Classes: `Test*`

### Fixtures

Fixtures provide reusable test setup:

```python
import pytest

@pytest.fixture
def user_data():
    return {"name": "Alice", "age": 30}

def test_user_creation(user_data):
    assert user_data["name"] == "Alice"
```

For complete reference, see [reference.md](reference.md).
```

**reference.md Excerpt**:

```markdown
# pytest Reference

## Fixture Scopes

- `function` (default): Run once per test function
- `class`: Run once per test class
- `module`: Run once per module
- `session`: Run once per test session

## Parametrization

```python
@pytest.mark.parametrize("input,expected", [
    (2, 4),
    (3, 9),
    (4, 16),
])
def test_square(input, expected):
    assert input ** 2 == expected
```
```

**examples.md Excerpt**:

```markdown
# pytest Examples

## Example 1: API Testing with Fixtures

```python
import pytest
import requests

@pytest.fixture
def api_client():
    return requests.Session()

@pytest.fixture
def base_url():
    return "https://api.example.com"

def test_get_user(api_client, base_url):
    response = api_client.get(f"{base_url}/users/1")
    assert response.status_code == 200
    assert "id" in response.json()
```
```

---

### Phase 4: Validation (20 min)

**Checklist Results**:
- ✅ Name: 45 characters (gerund format)
- ✅ Description: 256 characters (includes 7 keywords)
- ✅ SKILL.md: 420 lines
- ✅ Examples: 4 scenarios included
- ✅ Tested with Haiku and Sonnet

---

## Example 2: Creating a Git Workflow Skill (Abbreviated)

### Discovery

**Problem**: "Developers need GitFlow automation with commit conventions and PR templates."

**Keywords**: `git`, `gitflow`, `commit`, `pr`, `pull request`, `branch`, `merge`

### Design

```yaml
name: Automating Git Workflows with GitFlow
description: Manage Git branches using GitFlow, write conventional commits, create pull requests, and automate releases. Use when working with Git repositories, feature branches, or when the user mentions GitFlow, commits, or pull requests.
allowed-tools: Read, Bash
```

### Content Structure
- 35% High Freedom: Branch strategy philosophy
- 45% Medium Freedom: Commit patterns, PR templates
- 20% Low Freedom: Git commands, automation scripts

---

## Example 3: Creating a Security Scanning Skill (Abbreviated)

### Discovery

**Problem**: "Projects need automated security scanning for dependencies and SAST."

**Keywords**: `security`, `sast`, `vulnerabilities`, `dependencies`, `owasp`, `audit`

### Design

```yaml
name: Securing Applications with SAST and Dependency Scanning
description: Run static application security testing (SAST), scan dependencies for vulnerabilities, enforce OWASP Top 10 compliance, and integrate security checks into CI/CD. Use when performing security audits, scanning for vulnerabilities, or when the user mentions SAST, OWASP, or security scanning.
allowed-tools: Read, Bash
```

### Freedom Breakdown
- 25% High: Security principles, threat modeling
- 35% Medium: Scanning patterns, report interpretation
- 40% Low: Specific scanning commands per language

---

## Common Skill Creation Mistakes

### Mistake 1: Too Generic

❌ **Bad Name**: "Python Helper"
✅ **Good Name**: "Testing Python Applications with pytest"

### Mistake 2: Missing Keywords

❌ **Bad Description**: "Helps with Python testing."
✅ **Good Description**: "Write and run Python tests using pytest, fixtures, parametrization, and coverage reporting."

### Mistake 3: No Examples

❌ **Bad Practice**: Only provide theory and no code examples
✅ **Good Practice**: Include 3-4 complete, runnable examples

### Mistake 4: Outdated Information

❌ **Bad**: "Use Python 2.7 with unittest"
✅ **Good**: "Use Python 3.11+ with pytest (latest stable version)"

---

## Skill Creation Time Estimates

| Skill Complexity | Discovery | Design | Production | Validation | Total |
|-----------------|-----------|---------|-----------|------------|-------|
| **Simple** (single tool) | 10 min | 15 min | 30 min | 10 min | ~1 hour |
| **Medium** (multi-tool, 3-5 examples) | 15 min | 20 min | 60 min | 20 min | ~2 hours |
| **Complex** (multi-domain, 5+ examples) | 30 min | 30 min | 90 min | 30 min | ~3 hours |

---

## Quick Skill Creation Checklist

- [ ] Problem statement (1-2 sentences)
- [ ] 5+ trigger keywords identified
- [ ] 3+ reference URLs collected
- [ ] Name written (gerund + domain, ≤64 chars)
- [ ] Description written (capabilities + triggers, ≤1024 chars)
- [ ] Allowed tools listed (minimal set)
- [ ] Content outline (High/Medium/Low freedom %)
- [ ] SKILL.md written (≤500 lines)
- [ ] reference.md created (detailed docs)
- [ ] examples.md created (3-4 scenarios)
- [ ] CHECKLIST.md validation passed
- [ ] Tested with Haiku and Sonnet

---

**For complete skill creation reference, see [reference.md](reference.md)**

---

## Quick Skill Description Creation Examples

### Foundation Skill Description
```yaml
description: Validates SPEC YAML frontmatter (7 required fields id, version, status, created, updated, author, priority) and HISTORY section. Use when creating SPEC documents, validating SPEC metadata, checking SPEC structure, or authoring specifications.
```

### Alfred Skill Description
```yaml
description: Generates descriptive commit messages by analyzing git diffs. Use when writing commit messages, reviewing staged changes, or summarizing code modifications. Automatically activates git-workflow skill for advanced patterns.
```

### Language Skill Description
```yaml
description: TypeScript best practices with Vitest, Biome, strict typing. Use when implementing TypeScript code, writing tests, checking code quality, or applying type safety patterns.
```

### Domain Skill Description
```yaml
description: REST API design patterns with authentication, versioning, error handling. Use when designing REST APIs, implementing authentication, building backend services, or managing API versions.
```

---

**Last Updated**: 2025-11-05
**Version**: 2.2.0 (Integrated Description Writing Standards)
**Examples**: 3 complete workflows + description writing examples + mistake patterns
**Maintained by**: MoAI-ADK Foundation Team
