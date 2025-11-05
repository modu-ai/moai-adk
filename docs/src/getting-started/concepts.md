# Core Concepts

MoAI-ADK is built on five foundational concepts that work together to create a reliable, traceable, and maintainable development workflow. Understanding these concepts is key to unlocking the full potential of AI-assisted development.

## The Problem: Trust in AI Development

Modern AI-assisted development faces fundamental challenges:

- **Unclear Requirements**: "Build a login system" means different things to different people
- **Missing Tests**: AI often generates code without comprehensive test coverage
- **Documentation Drift**: Code changes but documentation becomes outdated
- **Context Loss**: Each interaction starts from scratch, losing project history
- **Untraceable Changes**: When requirements change, identifying affected code is difficult

## The Solution: SPEC-First TDD with Alfred

MoAI-ADK addresses these challenges through a systematic approach:

> **"No code without tests, no tests without specifications"**

This creates a chain of accountability: **SPEC → TEST → CODE → DOCUMENTATION**

## The 5 Core Concepts

### 1. SPEC-First Development

**What it is**: Writing clear, executable specifications before any code is written.

**Why it matters**:
- Eliminates ambiguity about what to build
- Provides a foundation for automated testing
- Ensures team alignment on requirements

**The EARS Syntax**: We use EARS (Easy Approach to Requirements Syntax) with 5 patterns:

1. **Ubiquitous** (Basic functionality): "The system SHALL provide JWT-based authentication"
2. **Event-driven** (Conditional): "**WHEN** valid credentials are provided, the system SHALL issue a token"
3. **State-driven** (State-based): "**WHILE** a user is authenticated, the system SHALL allow protected resources"
4. **Optional** (Optional features): "**WHERE** a refresh token exists, the system MAY issue a new token"
5. **Constraints** (Limitations): "Token expiration time SHALL NOT exceed 15 minutes"

**How it works**:
```bash
/alfred:1-plan "User authentication with JWT tokens"
```

Alfred's spec-builder automatically creates professional SPECs using EARS format.

### 2. Test-Driven Development (TDD)

**What it is**: Writing tests before implementation code, following the RED-GREEN-REFACTOR cycle.

**Why it matters**:
- Guarantees 85%+ test coverage
- Enables confident refactoring
- Provides living documentation of expected behavior

**The TDD Cycle**:

1. **🔴 RED**: Write failing tests first
   ```python
   def test_login_with_valid_credentials_should_return_token():
       """WHEN valid credentials are provided, system SHALL issue JWT token"""
       response = auth_client.login("user@example.com", "password123")
       assert response.status_code == 200
       assert "token" in response.json()
   ```

2. **🟢 GREEN**: Write minimal implementation to pass tests
   ```python
   def login(email: str, password: str) -> dict:
       if validate_credentials(email, password):
           return {"token": generate_jwt_token(email)}
       return {"error": "Invalid credentials"}
   ```

3. **♻️ REFACTOR**: Improve code quality while maintaining test coverage
   ```python
   class AuthService:
       def authenticate(self, email: str, password: str) -> AuthResult:
           if not self._validate_credentials(email, password):
               return AuthResult(success=False, error="Invalid credentials")

           token = self._generate_jwt_token(email)
           return AuthResult(success=True, token=token)
   ```

**How it works**:
```bash
/alfred:2-run SPEC-ID
```

Alfred automatically executes the complete TDD cycle.

### 3. @TAG System

**What it is**: A unique identifier system that links specifications, tests, code, and documentation.

**Why it matters**:
- Enables complete traceability across all project artifacts
- Makes impact analysis simple and reliable
- Prevents orphaned code and forgotten requirements

**The TAG Chain**:
```
@SPEC:EX-AUTH-001 (Requirements)
    ↓
@TEST:EX-AUTH-001 (Tests)
    ↓
@CODE:EX-AUTH-001:SERVICE (Implementation)
    ↓
@DOC:EX-AUTH-001 (Documentation)
```

**TAG Format**: `<DOMAIN>-<3-DIGIT-NUMBER>`

Examples: `AUTH-001`, `AUTH-002`, `USER-001`, `API-001`

**Usage Example**:
```bash
# Find all code related to authentication
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# Results:
# .moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:EX-AUTH-001: User Authentication
# tests/test_auth.py:3:# @TEST:EX-AUTH-001 | SPEC: SPEC-AUTH-001.md
# src/auth/service.py:5:# @CODE:EX-AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md
# docs/api/auth.md:24:- @SPEC:EX-AUTH-001
```

**How it works**:
```bash
/alfred:3-sync
```

Alfred automatically validates TAG chains and detects orphaned TAGs.

### 4. TRUST 5 Principles

**What it is**: A quality framework that ensures all code meets production standards.

**Why it matters**:
- Guarantees consistent code quality across the project
- Provides clear criteria for code reviews
- Prevents common bugs and security issues

**The 5 Principles**:

1. **🧪 Test First**
   - Test coverage ≥ 85%
   - All code protected by tests
   - Feature addition = test addition

2. **📖 Readable**
   - Functions ≤ 50 lines, files ≤ 300 lines
   - Variable names reveal intent
   - Linter compliance (ESLint/ruff/clippy)

3. **🎯 Unified**
   - SPEC-based architecture consistency
   - Repeating patterns (reduced learning curve)
   - Type safety or runtime validation

4. **🔒 Secured**
   - Input validation (XSS, SQL Injection prevention)
   - Password hashing (bcrypt, Argon2)
   - Sensitive data protection (environment variables)

5. **🔗 Trackable**
   - @TAG system usage
   - Git commits include TAG references
   - All decisions documented

**How it works**:
```bash
/alfred:3-sync
```

Alfred automatically validates TRUST 5 compliance.

### 5. Alfred SuperAgent

**What it is**: An AI orchestration system that coordinates multiple specialized agents and skills throughout the development process.

**Why it matters**:
- Eliminates prompt engineering complexity
- Maintains project context across sessions
- Provides consistent, professional-quality output

**Agent Architecture**:

```
Alfred SuperAgent (Orchestration)
    ├── Core Sub-agents (Project workflow)
    │   ├── project-manager 📋
    │   ├── spec-builder 🏗️
    │   ├── code-builder 💎
    │   ├── doc-syncer 📖
    │   └── quality-gate 🛡️
    ├── Expert Agents (Domain specialists)
    │   ├── backend-expert 🔧
    │   ├── frontend-expert 💻
    │   ├── devops-expert 🚀
    │   └── ui-ux-expert 🎨
    └── Built-in Claude Agents (General support)
        ├── Code understanding
        ├── Debugging
        └── Analysis
```

**Skills System**: 69+ production-ready Claude Skills organized in 4 tiers:

1. **Foundation**: Core principles (TRUST/TAG/SPEC/Git/EARS)
2. **Essentials**: Daily development tools (debug/perf/refactor)
3. **Alfred**: Workflow orchestration
4. **Domain**: Specialized knowledge (backend/frontend/security)
5. **Language**: Language-specific best practices (Python/TS/Go/Rust)

**How it works**:
```bash
/alfred:0-project    # Project initialization
/alfred:1-plan      # Specification creation
/alfred:2-run       # TDD implementation
/alfred:3-sync       # Documentation synchronization
```

## The Complete Workflow

### Step-by-Step Process

1. **PLAN** (2 minutes)
   ```bash
   /alfred:1-plan "User authentication with email/password"
   ```
   - Creates SPEC with @SPEC:AUTH-001
   - Defines requirements using EARS syntax
   - Status: `planning` → `draft`

2. **RUN** (5 minutes)
   ```bash
   /alfred:2-run AUTH-001
   ```
   - Executes TDD cycle (RED → GREEN → REFACTOR)
   - Creates tests with @TEST:AUTH-001
   - Creates implementation with @CODE:AUTH-001
   - Status: `draft` → `in_progress` → `testing`

3. **SYNC** (1 minute)
   ```bash
   /alfred:3-sync
   ```
   - Generates documentation with @DOC:AUTH-001
   - Validates TAG chain integrity
   - Checks TRUST 5 compliance
   - Status: `testing` → `completed`

### Result: Complete Traceability

```
@SPEC:EX-AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md
     ↓ (requirements)
@TEST:EX-AUTH-001 → tests/test_auth.py
     ↓ (validation)
@CODE:EX-AUTH-001 → src/auth/service.py
     ↓ (implementation)
@DOC:EX-AUTH-001 → docs/api/auth.md
```

## Benefits of the System

### For Individual Developers

- **Speed**: Clear requirements reduce back-and-forth time
- **Confidence**: 85%+ test coverage enables fearless refactoring
- **Clarity**: @TAG system makes code intent immediately obvious
- **Learning**: Professional patterns and best practices built-in

### For Teams

- **Consistency**: Everyone follows the same development patterns
- **Onboarding**: New team members understand code intent through SPECs
- **Quality**: TRUST 5 ensures consistent code quality
- **Collaboration**: SPECs provide clear communication about requirements

### For Projects

- **Maintainability**: Code and documentation always stay synchronized
- **Scalability**: TAG system makes impact analysis trivial
- **Reliability**: TDD ensures robust, well-tested code
- **Documentation**: Living documentation that evolves with the code

## Comparison with Traditional Development

| Aspect | Traditional Approach | MoAI-ADK Approach |
|--------|----------------------|-------------------|
| Requirements | Verbal descriptions, emails | Formal SPEC documents with EARS syntax |
| Testing | After implementation, often incomplete | First, with 85%+ coverage guaranteed |
| Documentation | Written separately, often outdated | Automatically synchronized with code |
| Traceability | Manual, often lost | @TAG system provides complete chain |
| Quality | Varies by developer | TRUST 5 principles ensure consistency |
| AI Usage | Prompt engineering, inconsistent | Standardized commands with reliable output |

## Getting Started with Concepts

1. **Experience the workflow**: Try the [Quick Start Guide](quick-start.md)
2. **Understand EARS syntax**: Learn [SPEC writing](../guides/specs/basics.md)
3. **Master TDD**: Follow the [TDD guide](../guides/tdd/red.md)
4. **Explore TAG system**: Read [TAG documentation](../reference/tags/index.md)

These concepts work together to create a development experience that is more reliable, maintainable, and enjoyable than traditional approaches. With Alfred as your guide, you'll write better code, faster, with confidence that it meets production standards.