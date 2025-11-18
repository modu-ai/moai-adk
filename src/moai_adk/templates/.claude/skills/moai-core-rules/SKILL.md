---
name: moai-core-rules
version: 4.0.0
created: 2025-11-02
updated: 2025-11-19
status: stable
tier: Alfred
description: "Defines the essential rules for Alfred SuperAgent. Based on November 2025 enterprise standard. 3-Layer architecture, 4-Step workflow, Agent-first paradigm, Skill invocation rules, AskUserQuestion patterns, TRUST 5 quality gates, TAG chain integrity, commit message standards. Usage: Workflow rule validation, quality gate confirmation, MoAI-ADK standard compliance, architecture rule validation."
keywords:
  - rules
  - agent-first
  - skill-invocation
  - ask-user-question
  - trust-5
  - tag-chain
  - workflow-compliance
  - quality-gates
  - 4-step-workflow
  - architecture-rules
allowed-tools: Read, Glob, Grep, Bash, AskUserQuestion, TodoWrite
stability: stable
---

# Alfred SuperAgent Core Rules

## Skill Overview

**moai-core-rules** is the essential framework controlling Alfred SuperAgent's decision-making and execution.

| Item | Value |
|------|-------|
| Version | 4.0.0 (November 2025 enterprise) |
| Tier | Alfred (top layer) |
| Auto-load | Triggered when rule validation, quality gates, or architecture rules needed |
| Architecture Paradigm | Agent-First (Command → Agent → Skill → Hook) |
| Workflow Model | 4-Step ADAP Workflow |

---

## What Does It Do?

### Core Responsibilities

1. **3-Layer Architecture Definition**: Commands → Agents → Skills layer separation
2. **4-Step Workflow Rules**: ADAP (Analyze, Design, Assure, Produce) + Intent
3. **Agent-First Paradigm**: Delegate all execution work to agents
4. **Skill Invocation Rules**: 10+ mandatory patterns, invocation syntax
5. **AskUserQuestion Patterns**: 5 essential usage scenarios
6. **TRUST 5 Quality Gates**: Validation criteria for T/R/U/S/T each
7. **TAG Chain Integrity**: SPEC→TEST→CODE→DOC traceability
8. **Commit Message Standards**: TDD cycle message formats

---

## When to Use

### Essential Scenarios (MUST use)

| Situation | Usage |
|-----------|-------|
| ✅ Validate Skill() invocation rules | **Required** |
| ✅ Clarify Command vs Agent responsibilities | **Required** |
| ✅ Decide on AskUserQuestion usage | **Required** |
| ✅ Confirm TRUST 5 compliance | **Required** |
| ✅ Validate TAG chain integrity | **Required** |
| ✅ Verify commit message format | **Required** |
| ✅ Validate workflow compliance | **Required** |
| ✅ Confirm Agent delegation correctness | **Required** |
| ✅ Validate quality gates (quality gate) | **Required** |
| ✅ Detect architecture rule violations | **Required** |

---

## Rule 1: 3-Layer Architecture (November 2025 Standard)

### Layer Structure

```
┌─────────────────────────────────────┐
│ Commands (Orchestration Only)       │ ← User-facing entry points
│ /alfred:0-project, /alfred:1-plan   │   No direct execution
│ /alfred:2-run, /alfred:3-sync       │
└──────────┬──────────────────────────┘
           │ Task(subagent_type="...")
           ↓
┌─────────────────────────────────────┐
│ Agents (Domain Expertise)           │ ← Deep reasoning
│ spec-builder, tdd-implementer       │   Complex decisions
│ test-engineer, doc-syncer           │   Plan → Execute
│ git-manager, qa-validator           │
└──────────┬──────────────────────────┘
           │ Skill("skill-name")
           ↓
┌─────────────────────────────────────┐
│ Skills (Knowledge Capsules)         │ ← Reusable patterns
│ 55 specialized Skills               │   Playbooks
│ < 1000 lines each                   │   Best practices
└─────────────────────────────────────┘
```

### Rule 1.1: Commands - Orchestration ONLY

**Forbidden (❌)**:
```bash
# ❌ WRONG: Direct task execution
echo "Building application..."
python setup.py build
git commit -m "Build"

# ❌ WRONG: Direct Skill invocation
Skill("moai-core-rules")  # Forbidden in Commands!

# ❌ WRONG: Complex logic implementation
if feature_type == "backend":
  # Complex business logic...
```

**Required (✅)**:
```bash
# ✅ CORRECT: Delegate to Agent
Task(
  subagent_type="tdd-implementer",
  description="Build and test application",
  prompt="Implement feature with RED-GREEN-REFACTOR cycle"
)

# ✅ CORRECT: Decision-making only, then delegate
if user_approval_needed:
  AskUserQuestion(...)  # Confirm with user, then delegate
  Task(subagent_type="implementation-planner", ...)
```

### Rule 1.2: Agents - Domain Expertise Ownership

**Agent responsibilities**:
- ✅ Complex analysis & reasoning (deep reasoning)
- ✅ Planning (planning)
- ✅ Decision-making (decision-making)
- ✅ Skill invocation and coordination (orchestration within domain)
- ✅ Task execution (execution)

**Example: tdd-implementer Agent**:
```
Agent receives: "Implement user authentication"
  ↓
1. Analyze: Review SPEC, analyze requirements
2. Design: Architecture design
3. RED: Invoke test-engineer Skill → Write tests
4. GREEN: Implement code
5. REFACTOR: Optimize
6. Commit: Invoke git-manager → Commit
  ↓
Returns: Fully tested, documented, committed code
```

### Rule 1.3: Skills - Knowledge Capsules (Stateless)

**Skill characteristics**:
- ✅ Stateless (no state)
- ✅ Reusable (can be invoked multiple times)
- ✅ Called by agents (agents invoke Skills)
- ✅ Under 1000 lines (< 1000 lines)
- ✅ Single topic (focused expertise)

**Forbidden (❌)**:
```bash
# ❌ WRONG: Skill calling other Skill
Skill("moai-foundation-git")  # Forbidden in Skills!

# ❌ WRONG: Skill executing Task()
Task(subagent_type="...")  # Forbidden in Skills!

# ❌ WRONG: Skill maintaining state
state = {"counter": 0}  # Violates stateless principle
```

---

## Rule 2: 4-Step Agent-Based Workflow (November 2025)

### Phase Overview

```
Phase 0: INTENT                      Phase 1: ANALYZE
┌──────────────────┐               ┌──────────────────┐
│ User Request     │ ─clarity?─→   │ WebSearch        │
│ Ambiguous?       │ NO             │ WebFetch         │
│ ✅ YES: clarify  │     YES        │ Research         │
│ ✅ NO: continue  │────────────→   │ Best practices   │
└──────────────────┘               └──────────────────┘
                                          ↓
Phase 3: ASSURE                     Phase 2: DESIGN
┌──────────────────┐               ┌──────────────────┐
│ Quality Gate     │←──────────────→│ Architecture     │
│ TRUST 5          │                │ Latest info      │
│ TAG integrity    │                │ Version specs    │
│ Compliance       │                │ Design patterns  │
└──────────────────┘               └──────────────────┘
        ↓
Phase 4: PRODUCE
┌──────────────────┐
│ Skill invocation │
│ File generation  │
│ Commit           │
└──────────────────┘
```

### Phase 0: Intent (Understanding User Intent)

**Rule 0.1**: Use AskUserQuestion if intent is ambiguous

```
Situation: "Create a data processing module"

Step 1: Evaluate Intent
  ├─ Clarity: LOW (What data? What processing?)
  └─ Action: Use AskUserQuestion

Step 2: Clarification
AskUserQuestion({
  question: "What data are you processing?",
  options: [
    "CSV file",
    "Database",
    "API response"
  ]
})

Step 3: Proceed to next phase with clarified requirements
```

### Phase 1: Analyze (Information Gathering & Research)

**Tools**: WebSearch, WebFetch

```
Task 1: Research latest information
  ├─ Search: "[framework] [version] best practices 2025"
  ├─ Fetch: Official documentation URLs
  └─ Validate: Cross-check multiple sources

Task 2: Collect best practices
  ├─ Official docs
  ├─ Industry standards
  └─ Current patterns (2025)

Task 3: Identify version-specific guidance
  ├─ Current stable version
  ├─ Breaking changes
  └─ Deprecation warnings
```

### Phase 2: Design (Architecture Design)

**Input**: Phase 1 research results
**Output**: Design based on November 2025 latest information

```
Design Activities:
  ├─ Naming based on latest information
  ├─ Specify current version
  ├─ Include latest patterns
  ├─ Link to official documentation
  └─ Include last update date
```

### Phase 3: Assure (Quality Validation)

**TRUST 5 Quality Gates**:

| Gate | Validation Criteria |
|------|------------------|
| **Test** | 85%+ coverage, all code paths tested |
| **Readable** | Clean code, SOLID principles, comments included |
| **Unified** | Consistent patterns, no duplication, naming standards |
| **Secured** | OWASP Top 10 verified, secrets removed |

### Phase 4: Produce (Generation & Release)

**Responsibility**: Skill invocation (e.g., moai-skill-factory)

```
Actions:
  ├─ Apply templates
  ├─ Generate files
  ├─ Add metadata
  ├─ Embed latest information
  ├─ Include official documentation links
  └─ Include version date
```

---

## Rule 3: Agent-First Paradigm (Critical Enforcement)

### Forbidden Tasks (❌ ABSOLUTELY FORBIDDEN)

Alfred (or Command) MUST NEVER directly execute:

1. **Direct bash command execution** ❌
   ```bash
   # ❌ WRONG
   bash("git commit -m 'message'")
   os.system("python build.py")
   
   # ✅ CORRECT
   Task(subagent_type="git-manager", prompt="Commit changes")
   ```

2. **File read/write** ❌
   ```bash
   # ❌ WRONG
   with open("file.py", "w") as f:
       f.write(code)
   
   # ✅ CORRECT
   Task(subagent_type="file-manager", prompt="Create file")
   ```

3. **Direct Git manipulation** ❌
   ```bash
   # ❌ WRONG
   subprocess.run(["git", "push", "origin", "main"])
   
   # ✅ CORRECT
   Task(subagent_type="git-manager", prompt="Push changes")
   ```

4. **Direct code analysis** ❌
   ```bash
   # ❌ WRONG
   lines = len(open("file.py").readlines())
   
   # ✅ CORRECT
   Task(subagent_type="code-analyzer", prompt="Analyze code")
   ```

5. **Direct test execution** ❌
   ```bash
   # ❌ WRONG
   subprocess.run(["pytest", "tests/"])
   
   # ✅ CORRECT
   Task(subagent_type="test-engineer", prompt="Run tests")
   ```

### Mandatory Delegation (✅ MANDATORY DELEGATION)

| Task | Delegate To | Pattern |
|------|----------|---------|
| Planning | plan-agent | `Task(subagent_type="plan-agent", ...)` |
| Code development | tdd-implementer | `Task(subagent_type="tdd-implementer", ...)` |
| Test writing | test-engineer | `Task(subagent_type="test-engineer", ...)` |
| Documentation | doc-syncer | `Task(subagent_type="doc-syncer", ...)` |
| Git operations | git-manager | `Task(subagent_type="git-manager", ...)` |
| Quality validation | qa-validator | `Task(subagent_type="qa-validator", ...)` |
| User questions | ask-user-questions | `AskUserQuestion(...)` |

---

## Rule 4: 10 Mandatory Skill Invocations

### Rule 4.1: Skill Invocation Pattern

**Syntax**:
```python
Skill("skill-name")  # Explicit invocation only
```

### 10 Essential Skills

| # | Skill | Purpose | Invocation |
|---|-------|---------|-----------|
| 1 | moai-foundation-trust | TRUST 5 validation | `Skill("moai-foundation-trust")` |
| 2 | moai-foundation-tags | TAG validation & traceability | `Skill("moai-foundation-tags")` |
| 3 | moai-foundation-specs | SPEC writing & validation | `Skill("moai-foundation-specs")` |
| 4 | moai-foundation-ears | EARS requirement format | `Skill("moai-foundation-ears")` |
| 5 | moai-foundation-git | Git workflow | `Skill("moai-foundation-git")` |
| 6 | moai-foundation-langs | Language & stack detection | `Skill("moai-foundation-langs")` |
| 7 | moai-essentials-debug | Debugging & error analysis | `Skill("moai-essentials-debug")` |
| 8 | moai-essentials-refactor | Refactoring & improvement | `Skill("moai-essentials-refactor")` |
| 9 | moai-essentials-perf | Performance optimization | `Skill("moai-essentials-perf")` |
| 10 | moai-essentials-review | Code review & quality | `Skill("moai-essentials-review")` |

### Rule 4.2: Skill Invocation Examples

```python
# Context 1: TRUST 5 validation needed
if validation_required:
    Skill("moai-foundation-trust")
    # → Returns: TRUST score, violations, recommendations

# Context 2: TAG chain validation
if tag_integrity_check:
    Skill("moai-foundation-tags")
    # → Returns: orphaned TAGs, broken chains, suggestions

# Context 3: SPEC writing
if spec_needed:
    Skill("moai-foundation-specs")
    # → Returns: SPEC template, validation results

# Context 4: Git workflow
if git_decision_needed:
    Skill("moai-foundation-git")
    # → Returns: branch strategy, commit format, merge rules
```

---

## Rule 5: AskUserQuestion Patterns (5 Essential Scenarios)

### Rule 5.1: MANDATORY Scenarios

**Scenario 1: Ambiguous technology stack**
```
Situation: "Recommend a Python web framework?"

AskUserQuestion({
  question: "What type of application?",
  header: "Application Type",
  options: [
    { label: "REST API", description: "High performance APIs" },
    { label: "Web Application", description: "Traditional MVC" },
    { label: "Microservice", description: "Event-driven" }
  ]
})
```

**Scenario 2: Architecture decision**
```
Situation: "How should I design the database model?"

AskUserQuestion({
  question: "What are your data characteristics?",
  header: "Data Model",
  options: [
    { label: "Relational", description: "Structured, ACID" },
    { label: "Document", description: "Flexible schema" },
    { label: "Graph", description: "Relationships" }
  ]
})
```

**Scenario 3: Ambiguous intent**
```
Situation: "Improve the code?"

AskUserQuestion({
  question: "Which aspect to improve?",
  header: "Improvement Focus",
  options: [
    { label: "Performance", description: "Speed & efficiency" },
    { label: "Readability", description: "Code clarity" },
    { label: "Security", description: "Vulnerability fixes" }
  ],
  multiSelect: true  # Allow multiple selection
})
```

**Scenario 4: Existing component impact**
```
Situation: "Want to upgrade a package, compatibility?"

AskUserQuestion({
  question: "Need to maintain existing code compatibility?",
  header: "Compatibility",
  options: [
    { label: "Full compatibility", description: "Maintain all APIs" },
    { label: "Deprecation path", description: "Gradual migration" },
    { label: "Breaking OK", description: "Version bump allowed" }
  ]
})
```

**Scenario 5: Resource constraints**
```
Situation: "Plan a system refactoring"

AskUserQuestion({
  question: "Expected development timeline?",
  header: "Timeline",
  options: [
    { label: "1 week", description: "Focused scope" },
    { label: "2-4 weeks", description: "Medium refactor" },
    { label: "1+ months", description: "Comprehensive" }
  ]
})
```

### Rule 5.2: Correct Usage

**❌ WRONG** (Plain text question):
```
User: "What would you prefer?"
Response: "Well, thinking about it... probably..."
```

**✅ CORRECT** (AskUserQuestion):
```
AskUserQuestion({
  question: "Which approach do you prefer?",
  header: "Approach",
  multiSelect: false,
  options: [
    { label: "Option A", description: "Benefit A, Cost B" },
    { label: "Option B", description: "Benefit C, Cost D" }
  ]
})
```

---

## Rule 6: TRUST 5 Quality Gates (November 2025 Enterprise)

### Validation Criteria for Each Gate

#### T: Test First (85%+ Coverage)
```yaml
requirements:
  coverage: "≥ 85%"
  coverage_tools: ["pytest-cov", "coverage.py"]
  test_types:
    - unit_tests: "Each function/method"
    - integration_tests: "Module interactions"
    - edge_cases: "Boundary values, error handling"
  
validation:
  - All code paths executed
  - Error conditions tested
  - Mock external dependencies
  - No skipped tests (×skip, ×xfail)
```

#### R: Readable (Clean Code)
```yaml
requirements:
  code_standards:
    - SOLID principles
    - DRY (Don't Repeat Yourself)
    - KISS (Keep It Simple, Stupid)
  
  documentation:
    - Function docstrings
    - Comments for complex logic
    - Type hints (Python 3.10+)
  
  naming:
    - Descriptive variable names
    - Consistent conventions
    - No single-letter vars (except i, j, k in loops)
```

#### U: Unified (Consistent Patterns)
```yaml
requirements:
  consistency:
    - Same patterns across codebase
    - No duplicate logic
    - Shared utilities for common tasks
  
  conventions:
    - Consistent indentation (4 spaces)
    - Consistent naming (snake_case, PascalCase)
    - Consistent import order
  
  validation:
    - Linting (flake8, black, isort)
    - Static analysis (pylint, mypy)
    - No code duplication (DRY violations)
```

#### S: Secured (OWASP Top 10)
```yaml
requirements:
  security_checks:
    - No hardcoded credentials
    - No SQL injection vectors
    - No XXE vulnerabilities
    - Input validation
    - Output encoding
  
  tools:
    - bandit (Python security linter)
    - safety (dependency vulnerabilities)
    - SAST scanning
  
  validation:
    - No secrets committed
    - Dependencies scanned
    - Known CVEs checked
```

#### T: Trackable (Complete Traceability)
```yaml
requirements:
  tag_chain:
    - Complete history traceability
  
  documentation:
    - SPEC → TEST → CODE → DOC
  
  validation:
    - Bidirectional references
```

---

## Rule 7: TAG Chain Integrity Rules

### Rule 7.1: TAG Naming Convention

```
Format: @<DOMAIN>-<###>

Examples:
@AUTH-001, @PAYMENT-042, @FRONTEND-003
```

### Rule 7.2: TAG Lifecycle

```
SPEC → TEST → CODE → DOC

Complete chain:
  └─ Example: test_auth.py - test_auth_success()

  └─ Example: auth.py - authenticate_user()

  └─ Example: CHANGELOG.md
```

### Rule 7.3: TAG Validation Rules

**❌ Violations**:
```python
# 1. Orphan TAG (TEST/CODE missing)

# 2. Broken chain (missing step)

# 3. Mismatch (different numbers)
```

**✅ Correct**:
```python
# Complete chain
```

---

## Rule 8: Commit Message Standards (TDD Cycle)

### Rule 8.1: Commit Format

```
Format:
<type>(<tag>): <subject>

<body>

<footer>
```

### Rule 8.2: TDD Cycle Commits

**RED Commit** (Write tests):
```
test(@AUTH-001): Add authentication tests

- test_successful_login()
- test_invalid_credentials()
- test_expired_token()

Status: RED (Tests fail as expected)
```

**GREEN Commit** (Implement):
```
feat(@AUTH-001): Implement authentication

- Implement authenticate_user() function
- Add token generation
- Add error handling

Status: GREEN (All tests pass)
```

**REFACTOR Commit** (Optimize):
```
refactor(@AUTH-001): Improve authentication code

- Extract token validation to separate function
- Add caching for user lookups
- Improve error messages

Status: PASSING (Tests still pass, code improved)
```

### Rule 8.3: Commit Types (Conventional Commits)

| Type | Description | Example |
|------|-------------|---------|
| **chore** | Build/configuration | `chore: Update dependencies` |
| **test** | Test code | `test: Add authentication tests` |
| **feat** | New feature | `feat: Add user authentication` |
| **fix** | Bug fix | `fix: Correct login validation` |
| **refactor** | Code improvement | `refactor: Optimize token validation` |

---

## Rule 9: Workflow Compliance Validation

### Rule 9.1: Compliance Checklist

**Before Commit**:
- [ ] Tests pass (85%+ coverage)
- [ ] Linting passes (black, flake8)
- [ ] Security scan passes (bandit, safety)
- [ ] Secrets removed (no secrets)
- [ ] Commit message format correct

**Before Merge**:
- [ ] TAG chain complete
- [ ] TRUST 5 passes
- [ ] Code review done
- [ ] CI/CD passes
- [ ] PR description included

### Rule 9.2: Violation Response

**Violation Detected**:
```
1. Auto-detect (hook)
2. Notify user
3. Request fix
4. Re-validate
5. Pass/Fail decision
```

---

## 3-Level Progressive Disclosure

### Level 1: Quick Start (Beginner - 10 minutes)

**What you need to know**:
1. Command = orchestration only, Agent = execution
2. Skill = specialized tool
3. Rule violation → error

### Level 2: Practical Patterns (Intermediate - 30 minutes)

**What you should do**:
1. Decide when to use AskUserQuestion
2. Validate TRUST 5
3. Maintain TAG chain
4. Follow commit message format

### Level 3: Advanced (Advanced - 1 hour)

**What you should optimize**:
1. Agent delegation strategy
2. Skill combination usage
3. Workflow automation
4. Exception management

---

## Official References & Links

### Architecture References
- Command-Agent-Skill pattern: Internal moai-adk documentation
- ADAP Workflow: Internal workflow definition

### Quality Standards
- TRUST 5 Framework: Skill("moai-foundation-trust")
- TAG System: Skill("moai-foundation-tags")
- Commit Standards: Skill("moai-foundation-git")

### Enterprise Standards (November 2025)
- Agent-First Architecture: InfoQ (agentic-ai-architecture-framework)
- TDD Best Practices: Official pytest documentation
- Code Quality: OWASP Top 10, SOLID principles

---

**Version**: 4.0.0 (November 2025 Enterprise Standard)
**Last Updated**: 2025-11-19
**Maintainer**: GoosLab (Alfred SuperAgent Framework)
**Language**: English
**Status**: Enterprise Production Ready ✅
