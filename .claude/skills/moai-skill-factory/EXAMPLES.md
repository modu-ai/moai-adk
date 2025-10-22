# Real-World Skill Examples

This document presents four complete, production-ready Skill examples demonstrating different freedom levels, scopes, and structures.

---

## Example 1: High Freedom Skill — Architecture Design

### Metadata

```yaml
---
name: "Designing Scalable Microservice Architecture"
description: "Plan microservice strategies with service decomposition,
communication patterns, and deployment architecture. Use when architecting
distributed systems, analyzing service boundaries, scaling strategies, or
designing cloud-native applications."
allowed-tools: "Read, Glob"
---
```

### Content Organization

- **SKILL.md**: Main framework (~380 lines)
- **reference.md**: Architecture patterns reference
- **examples.md**: 3 real-world scenarios
- **No scripts/templates**: High freedom (principles-based)

### File: SKILL.md (Excerpt)

```markdown
# Designing Scalable Microservice Architecture

## Core Principles

1. **Domain-Driven Design (DDD)**
   - Service boundary = bounded context
   - Each service owns its data
   - Minimize cross-service dependencies

2. **Communication Patterns**
   - Synchronous (REST/gRPC): Tight coupling, fast, simple
   - Asynchronous (Message Queue): Loose coupling, eventual consistency
   - Choose based on consistency requirements

3. **Deployment Strategy**
   - Containers + Orchestration (Kubernetes)
   - Service mesh for observability
   - Blue-green deployment for zero downtime

## Trade-offs Analysis

| Decision | Pros | Cons | When to Use |
|----------|------|------|-------------|
| Database per Service | Independence, scaled separately | Complex transactions, data consistency | >5 services |
| Shared Database | Simple transactions | Tight coupling | MVP, <3 services |
| Message Queue | Loose coupling, scalable | Eventual consistency, debugging | High throughput |

## Decision Framework

```
1. How many services? (>5 = separate databases)
2. Consistency requirements? (Immediate = synchronous)
3. Load patterns? (Bursty = message queues)
4. Team structure? (Multiple teams = microservices)
```

## See Also
- [reference.md](reference.md) for detailed patterns
- [examples.md](examples.md) for real-world scenarios
```

### Why High Freedom Works Here

- ✅ No deterministic algorithm
- ✅ Multiple valid approaches
- ✅ Context-dependent decisions
- ✅ Principles guide, not prescribe

---

## Example 2: Medium Freedom Skill — Testing Patterns

### Metadata

```yaml
---
name: "Writing Effective Unit Tests with Pytest"
description: "Create comprehensive unit tests using Pytest with fixtures,
mocking, parametrization, and coverage reporting. Use when testing Python
code, implementing test-driven development, improving code quality, or
ensuring test coverage above 85%."
allowed-tools: "Read, Bash(python:*), Bash(pytest:*)"
---
```

### Content Organization

- **SKILL.md**: Patterns and pseudocode (~420 lines)
- **reference.md**: Pytest API reference
- **examples.md**: 4 real test examples
- **scripts/**: Test automation helpers
- **templates/**: Test templates

### File: SKILL.md (Excerpt)

```markdown
# Writing Effective Unit Tests with Pytest

## Quick Start

```bash
pip install pytest pytest-cov pytest-mock
pytest tests/ -v --cov=src
```

## Testing Pattern: Setup-Execute-Assert

```pseudocode
1. SETUP: Create fixtures
   def setup():
       db = Database()
       user = User(name="Test")
       yield db, user
       db.cleanup()

2. EXECUTE: Run the code
   def test_user_creation(setup):
       db, user = setup
       result = db.create_user(user)

3. ASSERT: Verify outcomes
       assert result.success == True
       assert db.get_user("Test") is not None
```

## Mocking Pattern

```pseudocode
When to mock:
├─ External APIs (network calls)
├─ Databases (persistence)
├─ Random/time-dependent functions
└─ Resource-intensive operations

When NOT to mock:
├─ Pure functions (no side effects)
├─ Business logic you're testing
└─ Internal helper functions
```

## Parametrized Tests

```python
@pytest.mark.parametrize("input,expected", [
    ("valid", True),
    ("", False),
    ("123", True),
])
def test_validation(input, expected):
    assert validate(input) == expected
```

## See Also
- [reference.md](reference.md) for Pytest API
- [examples.md](examples.md) for test scenarios
- `scripts/run-tests.sh` for automation
```

### Why Medium Freedom Works Here

- ✅ Standard testing patterns exist
- ✅ Pseudocode clarifies best practices
- ✅ Examples demonstrate patterns
- ✅ Some flexibility (test structure varies)

---

## Example 3: Low Freedom Skill — Git Deployment

### Metadata

```yaml
---
name: "Automating Deployment with Git Hooks"
description: "Setup and manage pre-commit, pre-push, and post-deploy Git
hooks for automated testing, linting, and deployment. Use when configuring
CI/CD pipelines, preventing broken commits, automating quality checks, or
implementing GitFlow workflows."
allowed-tools: "Read, Write, Bash(git:*), Bash(bash:*)"
---
```

### Content Organization

- **SKILL.md**: Framework + ready-to-use scripts (~380 lines)
- **reference.md**: Git hooks reference
- **scripts/**: Production-ready hook implementations
- **templates/**: .git/hooks templates

### File: SKILL.md (Excerpt)

```markdown
# Automating Deployment with Git Hooks

## Git Hooks Overview

```
.git/hooks/
├── pre-commit (runs before commit)
├── pre-push (runs before push)
├── post-checkout (runs after checkout)
└── post-merge (runs after merge)
```

## Pre-Commit Hook (Run Linters/Tests)

```bash
#!/bin/bash
set -euo pipefail

echo "Running pre-commit checks..."

# Run linter
if ! python -m ruff check .; then
  echo "❌ Rinting failed. Fix issues and retry."
  exit 1
fi

# Run tests
if ! python -m pytest tests/ -q; then
  echo "❌ Tests failed. Fix and retry."
  exit 1
fi

echo "✓ Pre-commit checks passed"
exit 0
```

## Installation

```bash
# Copy script to .git/hooks/
cp templates/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Verify
.git/hooks/pre-commit  # Should run without errors
```

See [scripts/](scripts/) for production-ready implementations.
```

### Why Low Freedom Works Here

- ✅ Deterministic steps
- ✅ Error-prone if done manually
- ✅ Standard patterns (hooks are fixed)
- ✅ Scripts reduce mistakes

---

## Example 4: Balanced Skill — Security Scanning

### Metadata

```yaml
---
name: "Security Vulnerability Scanning with OWASP Tools"
description: "Identify security vulnerabilities using OWASP Top 10
checklist, static analysis tools (SonarQube, Semgrep), and dependency
scanning. Use when auditing code, implementing security best practices,
reducing attack surface, or preparing for security reviews."
allowed-tools: "Read, Grep, Bash(curl:*), Bash(docker:*)"
---
```

### Content Organization

- **SKILL.md**: Framework + patterns (~350 lines)
- **reference.md**: OWASP Top 10 reference
- **examples.md**: 3 real security scenarios
- **scripts/**: Automated scanning scripts
- **templates/**: Scanning configurations

### File: SKILL.md (Excerpt)

```markdown
# Security Vulnerability Scanning with OWASP Tools

## OWASP Top 10 Quick Reference

1. **Broken Access Control** → Check authorization logic
2. **Cryptographic Failures** → Verify encryption usage
3. **Injection** → Sanitize user inputs
4. **Insecure Design** → Review threat model
5. **Security Misconfiguration** → Audit default settings
6. **Vulnerable Components** → Scan dependencies
7. **Authentication Failures** → Test password policies
8. **Software/Data Integrity Failures** → Verify checksums
9. **Logging/Monitoring Failures** → Check security logs
10. **SSRF** → Validate URL inputs

## Scanning Pattern: Automated + Manual

```pseudocode
1. Static Analysis (Automated)
   ├─ Run SonarQube or Semgrep
   ├─ Parse results
   └─ Flag high-severity issues

2. Dependency Scan (Automated)
   ├─ Run `pip audit` or npm audit
   ├─ Update vulnerable packages
   └─ Document exceptions

3. Manual Review (Required)
   ├─ Authentication logic
   ├─ Cryptography implementation
   └─ Input validation
```

## Running Automated Scans

```bash
# Dependency scanning
pip install safety
safety check --file requirements.txt

# Code scanning
docker run -v $(pwd):/src returntocorp/semgrep \
  semgrep --config=p/security-audit /src
```

## See Also
- [scripts/run-scan.sh](scripts/run-scan.sh) for automation
- [templates/semgrep-config.yaml](templates/semgrep-config.yaml)
- [reference.md](reference.md) for OWASP details
- [examples.md](examples.md) for real scenarios
```

### Why Balanced Works Here

- ✅ High: Principles (OWASP Top 10)
- ✅ Medium: Patterns (scanning workflow)
- ✅ Low: Scripts (automated tools)
- ✅ Combines all three levels naturally

---

## Comparison Matrix

| Aspect | Architecture | Testing | Git Hooks | Security |
|--------|--------------|---------|-----------|----------|
| **Freedom Mix** | High 70% | Medium 60% | Low 70% | Balanced 40/35/25 |
| **SKILL.md Size** | 380 lines | 420 lines | 350 lines | 380 lines |
| **Scripts** | None | 2-3 | 3-5 | 3-5 |
| **Complexity** | Conceptual | Practical | Operational | Both |
| **Use Case** | Planning | Development | Operations | Security |

---

## Key Takeaways

### When to Use High Freedom

- Conceptual, principle-based guidance
- Multiple valid approaches
- Context-dependent decisions
- No deterministic algorithm

**Examples**: Architecture, design patterns, best practices

### When to Use Medium Freedom

- Standard patterns exist
- Some flexibility in implementation
- Pseudocode clarifies approach
- Examples guide but don't prescribe

**Examples**: Testing, data validation, API design

### When to Use Low Freedom

- Deterministic, error-prone operations
- Specific steps required
- Scripts reduce mistakes
- Automation preferred

**Examples**: Deployment, automation, configuration

### When to Use Balanced

- Complex topics with multiple layers
- Mix of principles + practical patterns + automation
- Different use cases for different users

**Examples**: Security, CI/CD, cloud operations

---

## Real-World File Structures from Examples

### Example 1: High Freedom
```
moai-skill-architecture/
├── SKILL.md (380 lines)
├── reference.md (patterns reference)
└── examples.md (3 scenarios)
```

### Example 2: Medium Freedom
```
moai-skill-testing/
├── SKILL.md (420 lines)
├── reference.md (Pytest API)
├── examples.md (4 examples)
├── scripts/
│   └── run-tests.sh
└── templates/
    └── test-template.py
```

### Example 3: Low Freedom
```
moai-skill-deployment/
├── SKILL.md (350 lines)
├── reference.md (git hooks reference)
├── scripts/
│   ├── pre-commit.sh
│   ├── pre-push.sh
│   └── post-deploy.sh
└── templates/
    └── hook-template.sh
```

### Example 4: Balanced
```
moai-skill-security/
├── SKILL.md (380 lines)
├── reference.md (OWASP Top 10)
├── examples.md (3 scenarios)
├── scripts/
│   ├── run-dependency-scan.sh
│   ├── run-static-analysis.sh
│   └── parse-results.py
└── templates/
    ├── semgrep-config.yaml
    └── security-scan-config.json
```

---

## Metadata Lessons

### High Freedom Example

```yaml
description: "Plan microservice strategies...
Use when architecting distributed systems,
analyzing service boundaries, scaling strategies..."
```
↑ Trigger keywords: "architecting", "scaling", "distributed"

### Medium Freedom Example

```yaml
description: "Create comprehensive unit tests using Pytest...
Use when testing Python code, implementing test-driven
development, improving code quality..."
```
↑ Keywords: "testing", "Pytest", "test-driven", "quality"

### Low Freedom Example

```yaml
description: "Setup and manage Git hooks for automated
testing, linting, and deployment...
Use when configuring CI/CD pipelines, preventing
broken commits..."
```
↑ Keywords: "Git", "CI/CD", "automation", "deployment"

### Balanced Example

```yaml
description: "Identify security vulnerabilities using
OWASP tools...
Use when auditing code, implementing security
best practices..."
```
↑ Keywords: "security", "OWASP", "vulnerabilities", "audit"

---

## Testing Each Skill Type

### High Freedom (Architecture)
- Haiku: Can understand principles? ✓
- Sonnet: Applies framework correctly? ✓
- Opus: Extends to novel scenarios? ✓

### Medium Freedom (Testing)
- Haiku: Understands patterns? ✓
- Sonnet: Uses mocking correctly? ✓
- Opus: Creates advanced test scenarios? ✓

### Low Freedom (Git Hooks)
- Haiku: Follows script exactly? ✓
- Sonnet: Customizes hooks appropriately? ✓
- Opus: Integrates with CI/CD? ✓

### Balanced (Security)
- Haiku: Understands OWASP principles? ✓
- Sonnet: Runs all scanning tools? ✓
- Opus: Integrates results and recommends? ✓

---

**Version**: 0.3.0 (with Interactive Discovery & Web Research)
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Code Skills + skill-factory
**Related Guides**: [SKILL.md](SKILL.md), [METADATA.md](METADATA.md), [STRUCTURE.md](STRUCTURE.md), [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md)
