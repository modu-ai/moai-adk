---
name: "moai-alfred-dev-guide"
version: "4.0.0"
created: 2025-11-02
updated: 2025-11-13
tier: Alfred
status: production
description: "Enterprise-grade development guide for Alfred SuperAgent orchestrating SPEC-First TDD workflow. Covers context engineering, TRUST 5 principles, EARS requirements format, and @TAG traceability for /alfred:1-plan, /alfred:2-run, /alfred:3-sync commands."
keywords: ["spec-first", "tdd", "alfred-development", "context-engineering", "trust-principles", "ears-format", "enterprise-v4", "red-green-refactor", "tag-traceability"]
allowed-tools: "Read, Bash(rg:*), Bash(grep:*), AskUserQuestion, TodoWrite"
primary-agent: "alfred"
secondary-agents: ["spec-builder", "tdd-implementer", "doc-syncer", "git-manager", "quality-gate"]
---

# Alfred Development Guide

## Quick Start

**Purpose**: Orchestrate SPEC-First TDD development workflow using Alfred SuperAgent capabilities  
**When to Use**: Any development task requiring structured planning, implementation, and documentation  
**Basic Usage**:

```bash
# Core development workflow
/alfred:1-plan "feature description"    # Create SPEC
/alfred:2-run SPEC-XXX                  # TDD implementation
/alfred:3-sync auto SPEC-XXX            # Documentation sync

# Load this skill for guidance
Skill("moai-alfred-dev-guide")
```

### Core Workflow: SPEC → TEST → CODE → DOC

**No spec, no code. No tests, no implementation.**

1. **SPEC Phase** (`/alfred:1-plan`): Author detailed specifications using EARS format
2. **TDD Phase** (`/alfred:2-run`): RED-GREEN-REFACTOR cycle with @TAG traceability
3. **SYNC Phase** (`/alfred:3-sync`): Verify integrity and synchronize documentation

### Essential Alfred Commands

```bash
/alfred:0-project        # Initialize project structure
/alfred:1-plan "task"    # Plan and create SPEC
/alfred:2-run SPEC-XXX   # Implement with TDD
/alfred:3-sync auto SPEC-XXX  # Sync documentation
```

## Implementation

### Context Engineering Strategy

**Load only what's needed, when it's needed**:

| Command | Always Load | Optional Load | Never Load |
|---------|-------------|---------------|------------|
| `/alfred:1-plan` | `.moai/project/product.md` | `.moai/project/structure.md`, `.moai/project/tech.md` | Individual SPEC files |
| `/alfred:2-run` | `.moai/specs/SPEC-{ID}/spec.md` | `development-guide.md`, related SPECs | Unrelated docs, analysis |
| `/alfred:3-sync` | Previous sync report | Modified SPEC files, TAG validation | Old reports, unrelated files |

### EARS Requirements Format

**5 Patterns for Complete Specifications**:

```markdown
### Ubiquitous Requirements (Baseline)
- The system shall provide user authentication via email and password.

### Event-driven Requirements (WHEN)
- WHEN a user submits signup form, the system shall create account.

### State-driven Requirements (WHILE)
- WHILE user is authenticated, the system shall allow access to protected resources.

### Optional Features (WHERE)
- WHERE 2FA is enabled, the system may require additional verification.

### Constraints (IF)
- IF password is invalid 3 times, the system shall lock account for 1 hour.
```

### TDD Implementation Pattern

**RED Phase**: Write failing tests first
```python
def test_signup_valid_user():
    """SPEC: The system shall create user account."""
    response = signup(email="user@example.com", password="securePass123")
    assert response["status"] == "created"
```

**GREEN Phase**: Minimal passing implementation
```python
def signup(email, password):
    # Minimal implementation to pass tests
    user = User.create(email=email, password=hash_password(password))
    return {"status": "created", "user_id": user.id}
```

**REFACTOR Phase**: Improve code quality
```python
def signup(email: str, password: str) -> Dict[str, str]:
    """Create user account with validation and security."""
    _validate_email(email)
    _validate_password(password)
    
    user = User.create(
        email=email, 
        password_hash=bcrypt.hashpw(password.encode(), bcrypt.gensalt(12))
    )
    send_verification_email(user.email)
    
    return {"status": "created", "user_id": user.id}
```

### @TAG Traceability System

**Mandatory tagging for traceability**:

- `@SPEC:ID` - Specification documents
- `@TEST:ID` - Test files and test cases  
- `@CODE:ID` - Implementation code
- `@DOC:ID` - Documentation and README files

**Validation**:
```bash
# Verify complete TAG chain
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n .moai/specs/ tests/ src/ docs/

# Expected output:
# .moai/specs/SPEC-AUTH-001/spec.md:SPEC-AUTH-001
# tests/test_auth.py:TEST-AUTH-001  
# src/auth.py:CODE-AUTH-001
# README.md:DOC-AUTH-001
```

### TRUST 5 Principles Checklist

**T - Test-Driven**: 
- [ ] SPEC written with EARS format
- [ ] Tests written before implementation (RED)
- [ ] Minimal code passes tests (GREEN)
- [ ] Code improved while maintaining tests (REFACTOR)

**R - Readable**:
- [ ] Descriptive variable and function names
- [ ] Comments explain WHY, not WHAT
- [ ] Docstrings for all public functions
- [ ] Consistent code formatting

**U - Unified**:
- [ ] Consistent naming conventions
- [ ] Same patterns across codebase
- [ ] Shared language (EARS for specs)
- [ ] Standardized documentation format

**S - Secured**:
- [ ] No hardcoded secrets
- [ ] Input validation present
- [ ] Error handling implemented
- [ ] OWASP Top 10 compliance

**E - Evaluated**:
- [ ] Test coverage ≥85%
- [ ] Performance metrics defined
- [ ] Quality gates implemented
- [ ] Code review completed

## Advanced

### Multi-Agent Orchestration

**Alfred coordinates specialist agents**:

```bash
# Alfred delegates to specialists based on task type
Task("spec-builder")     # Creates detailed SPECs
Task("tdd-implementer")  # Implements RED-GREEN-REFACTOR
Task("doc-syncer")       # Synchronizes documentation
Task("git-manager")      # Manages Git workflow
Task("quality-gate")     # Validates TRUST principles
```

### Progressive Disclosure Architecture

**Layer 1 - Quick Start** (SKILL.md):
- Essential workflow overview
- Core commands and patterns
- Quick reference examples

**Layer 2 - Implementation** (examples.md):
- Complete working examples
- Step-by-step tutorials
- Best practice patterns

**Layer 3 - Reference** (reference.md):
- Detailed technical specifications
- Command reference
- Advanced configuration options

### Enterprise Integration Patterns

**CI/CD Pipeline Integration**:
```yaml
# .github/workflows/alfred-validation.yml
name: Alfred Quality Gates
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate SPEC compliance
        run: moai-adk validate-specs
      - name: Check TRUST principles
        run: moai-adk check-trust
      - name: Verify TAG chains
        run: rg '@(SPEC|TEST|CODE|DOC):' --stats
```

**Quality Metrics Dashboard**:
```python
# Automated quality reporting
def generate_quality_report():
    return {
        "spec_coverage": calculate_spec_coverage(),
        "test_coverage": get_test_coverage(),
        "trust_compliance": validate_trust_principles(),
        "tag_integrity": verify_tag_chains(),
        "code_quality": run_static_analysis()
    }
```

### Performance Optimization

**Context Budgeting**:
- Phase 1: Load ~5MB (project overview only)
- Phase 2: Load ~200MB (specific SPEC + implementation)
- Phase 3: Load ~50MB (sync-related files only)

**Memory Management**:
- Use JIT loading for large documentation
- Cache frequently accessed SPECs
- Clean up context between phases

## Security & Compliance

### Security Considerations

**Credential Management**:
- Never commit API keys or secrets
- Use environment variables for configuration
- Implement proper secret rotation

**Input Validation**:
- Validate all user inputs in SPECs
- Implement sanitization in code
- Document security assumptions

### Enterprise Compliance

**Documentation Standards**:
- All SPECs must follow EARS format
- TAG traceability mandatory for audit
- Change logs maintained automatically

**Quality Gates**:
- Minimum 85% test coverage required
- All TRUST principles must be met
- Security review mandatory for production

## Related Skills

- **moai-alfred-spec-builder**: Create detailed SPECs with EARS format
- **moai-alfred-tdd-implementer**: Execute RED-GREEN-REFACTOR cycle
- **moai-alfred-doc-syncer**: Synchronize documentation with code
- **moai-alfred-git-manager**: Manage Git workflow and branches
- **moai-alfred-quality-gate**: Validate TRUST 5 principles
- **moai-cc-skill-factory**: Create new Skills with proper structure
- **moai-foundation-trust**: Implement trust principles in development

---

*Enterprise v4.0 Compliant*  
*Progressive Disclosure: Quick → Implementation → Advanced*  
*Context Engineering for Optimal Performance*
