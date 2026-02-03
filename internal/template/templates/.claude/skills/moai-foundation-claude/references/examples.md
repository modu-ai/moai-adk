# Claude Code Authoring Kit Examples

Multishot prompting examples demonstrating practical Claude Code patterns covering skills, sub-agents, custom slash commands, hooks, memory, and settings.

---

## Example 1: Creating a Custom Skill

**Scenario**: Create a skill for Python testing best practices.

**Input**:
```yaml
# Need: A skill for Python testing that:
# - Follows Claude Code official standards
# - Uses progressive disclosure
# - Stays under 500 lines
# - Includes practical examples
```

**Output**:
```yaml
# File: .claude/skills/python-testing/SKILL.md
---
name: python-testing
description: Python testing best practices with pytest, coverage, and TDD patterns for modern applications
version: 1.0.0
updated: 2025-11-26
status: active
tools: Read, Write, Edit, Bash, Grep, Glob
---

# Python Testing Specialist

## Quick Reference (30 seconds)

Python Testing Mastery - Comprehensive testing patterns with pytest, coverage analysis, and TDD workflows.

Core Capabilities:
- Unit Testing: pytest fixtures, parametrization, mocking
- Integration Testing: Database, API, and service testing
- Coverage: pytest-cov configuration and reporting
- TDD Workflow: RED-GREEN-REFACTOR cycle implementation

When to Use:
- Writing unit tests for Python modules
- Setting up pytest configuration
- Implementing TDD for new features
- Achieving coverage targets

Quick Start:
```python
# Basic test structure
import pytest

def test_user_creation():
    user = User(name="John", email="john@example.com")
    assert user.name == "John"
    assert user.is_valid()

@pytest.fixture
def db_session():
    session = create_test_session()
    yield session
    session.rollback()
```

---

## Implementation Guide (5 minutes)

### pytest Configuration

```toml
# pyproject.toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
python_functions = ["test_*"]
addopts = "-v --tb=short --strict-markers"
markers = [
    "slow: marks tests as slow",
    "integration: marks integration tests"
]

[tool.coverage.run]
source = ["src"]
branch = true
omit = ["*/tests/*", "*/__init__.py"]

[tool.coverage.report]
fail_under = 80
show_missing = true
```

### Fixture Patterns

```python
import pytest
from unittest.mock import Mock, patch

# Scope-based fixtures
@pytest.fixture(scope="session")
def database():
    """Create database once for all tests."""
    db = create_test_database()
    yield db
    db.cleanup()

@pytest.fixture(scope="function")
def user(database):
    """Create fresh user for each test."""
    user = User.create(database, name="Test User")
    yield user
    user.delete()

# Factory fixtures
@pytest.fixture
def user_factory(database):
    """Factory for creating multiple users."""
    created_users = []

    def _create_user(**kwargs):
        user = User.create(database, **kwargs)
        created_users.append(user)
        return user

    yield _create_user

    for user in created_users:
        user.delete()
```

---

## Advanced Patterns (10+ minutes)

### Parametrized Testing

```python
@pytest.mark.parametrize("input,expected", [
    ("hello", "HELLO"),
    ("World", "WORLD"),
    ("", ""),
    ("123", "123"),
])
def test_uppercase(input, expected):
    assert input.upper() == expected

@pytest.mark.parametrize("email,is_valid", [
    ("user@example.com", True),
    ("invalid-email", False),
    ("user@", False),
    ("@example.com", False),
])
def test_email_validation(email, is_valid):
    assert validate_email(email) == is_valid
```

### Async Testing

```python
import pytest

@pytest.mark.asyncio
async def test_async_api_call():
    result = await api_client.get_user(user_id=1)
    assert result.status == "success"
    assert result.data["id"] == 1
```

---

## Works Well With

- moai-lang-python - Python 3.13+ patterns
- moai-domain-backend - Backend testing strategies
- moai-workflow-ddd - DDD workflow integration
```

**Explanation**: This skill follows Claude Code official standards with progressive disclosure (Quick/Implementation/Advanced), stays under 500 lines, and includes practical pytest examples.

---

## Example 2: Sub-agent Delegation Patterns

**Scenario**: Implement efficient sub-agent delegation for a SPEC workflow.

**Input**:
```python
# Need: Coordinate multiple agents for SPEC implementation
# Requirements:
# - Sequential for dependent tasks
# - Parallel for independent tasks
# - Proper error handling
# - Context passing between agents
```

**Output**:
```python
# SPEC Workflow with Sub-agent Delegation

## Sequential Delegation Pattern

```python
# Phase 1: Analysis with spec-builder
analysis = Task(
    subagent_type="spec-builder",
    prompt="""
    Analyze the following requirement and create a SPEC:

    Requirement: User authentication system with JWT tokens

    Output format:
    - SPEC ID
    - Requirements list
    - Architecture overview
    - Test criteria
    """,
    context={
        "project_type": "web_api",
        "language": "python",
        "framework": "fastapi"
    }
)

# Phase 2: Implementation with ddd-implementer (depends on analysis)
implementation = Task(
    subagent_type="ddd-implementer",
    prompt=f"""
    Implement the SPEC using DDD approach:

    SPEC ID: {analysis.spec_id}
    Requirements: {analysis.requirements}

    Follow ANALYZE-PRESERVE-IMPROVE cycle:
    1. Analyze existing structure and behavior
    2. Preserve behavior with characterization tests
    3. Improve structure incrementally
    """,
    context={
        "spec_id": analysis.spec_id,
        "architecture": analysis.architecture
    }
)

# Phase 3: Validation with quality-gate (depends on implementation)
validation = Task(
    subagent_type="quality-gate",
    prompt=f"""
    Validate the implementation:

    SPEC ID: {implementation.spec_id}
    Files changed: {implementation.files}

    Check:
    - All tests pass
    - Coverage >= 80%
    - No security issues
    - Code quality standards met
    """,
    context={
        "implementation": implementation,
        "original_spec": analysis
    }
)
```

## Parallel Delegation Pattern

```python
# Independent tasks can run simultaneously
# for 3x faster execution

# All three can run in parallel
results = await Promise.all([
    # Backend implementation
    Task(
        subagent_type="backend-expert",
        prompt="Implement API endpoints for SPEC-001",
        context={"spec_id": "SPEC-001", "focus": "api"}
    ),

    # Frontend implementation
    Task(
        subagent_type="frontend-expert",
        prompt="Implement UI components for SPEC-001",
        context={"spec_id": "SPEC-001", "focus": "ui"}
    ),

    # Documentation generation
    Task(
        subagent_type="docs-manager",
        prompt="Generate API documentation for SPEC-001",
        context={"spec_id": "SPEC-001", "focus": "docs"}
    )
])

# Aggregate results
backend_result = results[0]
frontend_result = results[1]
docs_result = results[2]

# Integration validation (sequential - depends on all parallel tasks)
integration = Task(
    subagent_type="quality-gate",
    prompt="Validate integration of all components",
    context={
        "backend": backend_result,
        "frontend": frontend_result,
        "docs": docs_result
    }
)
```

## Error Handling Pattern

```python
async def execute_with_retry(task_config, max_retries=3):
    """Execute task with automatic retry on failure."""

    for attempt in range(max_retries):
        try:
            result = Task(**task_config)

            if result.success:
                return result

            # Log failure but don't raise yet
            log_warning(f"Attempt {attempt + 1} failed: {result.error}")

        except AgentTimeoutError:
            log_warning(f"Timeout on attempt {attempt + 1}")

        except AgentOverloadError:
            # Wait before retry
            await sleep(2 ** attempt)

    # All retries exhausted
    raise TaskExecutionError(f"Failed after {max_retries} attempts")
```

**Explanation**: This pattern shows sequential delegation for dependent tasks, parallel execution for independent work, and proper error handling for resilient workflows.

---

## Example 3: Custom Slash Command with Hooks

**Scenario**: Create a deployment command with pre/post hooks for validation.

**Input**:
```markdown
# Need: Deployment command that:
# - Validates before deployment
# - Runs tests before deploying
# - Notifies on completion
# - Supports rollback
```

**Output**:
```markdown
# File: .claude/commands/deploy.md
---
description: Deploy application with validation, testing, and notification
allowed-tools: Bash, Read, Write, Glob
---

# Deployment Command

Deploy the application to the specified environment.

## Arguments

- $1: Environment (staging/production)
- $2: Version tag (optional, defaults to latest)

## Execution Flow

### Step 1: Pre-deployment Validation

Verify deployment requirements:

```bash
# Check environment
if [ "$1" != "staging" ] && [ "$1" != "production" ]; then
    echo "Error: Invalid environment. Use 'staging' or 'production'"
    exit 1
fi

# Verify clean git state
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory not clean. Commit or stash changes."
    exit 1
fi
```

### Step 2: Run Tests

Execute full test suite before deployment:

```bash
# Run tests with coverage
pytest tests/ --cov=src --cov-fail-under=80

if [ $? -ne 0 ]; then
    echo "Error: Tests failed. Deployment aborted."
    exit 1
fi
```

### Step 3: Build and Deploy

Build and push to environment:

```bash
# Build Docker image
docker build -t myapp:$VERSION .

# Push to registry
docker push registry.example.com/myapp:$VERSION

# Deploy to environment
kubectl set image deployment/myapp myapp=registry.example.com/myapp:$VERSION
```

### Step 4: Health Check

Verify deployment success:

```bash
# Wait for rollout
kubectl rollout status deployment/myapp --timeout=5m

# Run health check
curl -f https://$ENVIRONMENT.example.com/health || exit 1
```

### Step 5: Notification

Notify team of deployment:

```bash
# Send Slack notification
curl -X POST -H 'Content-type: application/json' \
    --data '{"text":"Deployed v'$VERSION' to '$ENVIRONMENT'"}' \
    $SLACK_WEBHOOK_URL
```
```

```json
// File: .claude/settings.json (hooks section)
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "validate-bash-command",
            "description": "Validate bash commands before execution"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "git add $FILE",
            "description": "Auto-stage written files"
          }
        ]
      }
    ]
  }
}
```

**Explanation**: This pattern combines a custom slash command with hooks for validation, testing, and notification, creating a complete deployment workflow.

---

## Common Patterns

### Pattern 1: Memory File Organization

Organize memory for efficient context loading:

```markdown
# File: .claude/CLAUDE.md (Project-level memory)

## Project Overview
- Name: MyApp
- Type: Web API
- Stack: Python 3.13, FastAPI, PostgreSQL

## Development Guidelines
- Follow TDD for all new features
- Minimum 80% test coverage
- Use type hints everywhere

## Active SPECs
- SPEC-001: User Authentication (In Progress)
- SPEC-002: API Rate Limiting (Planned)

@import architecture.md
@import coding-standards.md
```

```markdown
# File: .claude/architecture.md

## System Architecture
- API Layer: FastAPI with automatic OpenAPI
- Database: PostgreSQL with async SQLAlchemy
- Cache: Redis for session management
- Auth: JWT with refresh tokens
```

### Pattern 2: Settings Hierarchy

Configure settings at appropriate levels:

```json
// ~/.claude/settings.json (User-level)
{
  "preferences": {
    "outputStyle": "concise",
    "codeStyle": "modern"
  },
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash"]
  }
}
```

```json
// .claude/settings.json (Project-level)
{
  "model": "claude-sonnet-4-5-20250929",
  "permissions": {
    "allow": ["Read", "Write", "Edit"],
    "deny": ["Bash dangerous commands"]
  },
  "hooks": {
    "PreToolUse": [...]
  }
}
```

### Pattern 3: IAM Permission Tiers

Define permissions based on agent role:

```markdown
## Permission Tiers

### Tier 1: Read-Only Agents
- Tools: Read, Grep, Glob
- Use for: Code analysis, documentation review
- Example agents: code-analyzer, doc-reviewer

### Tier 2: Write-Limited Agents
- Tools: Read, Write, Edit, Grep, Glob
- Restrictions: Cannot modify production files
- Use for: Code generation, refactoring
- Example agents: code-generator, refactorer

### Tier 3: Full-Access Agents
- Tools: All including Bash
- Restrictions: Dangerous commands require approval
- Use for: Deployment, system administration
- Example agents: deployer, admin

### Tier 4: Admin Agents
- Tools: All with elevated permissions
- Use for: System configuration, security
- Example agents: security-auditor, config-manager
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: Monolithic Skills

**Problem**: Skills exceeding 500 lines become hard to maintain and load.

```markdown
# Incorrect: Single 1500-line SKILL.md
---
name: everything-skill
---

## Quick Reference
[200 lines...]

## Implementation
[800 lines...]

## Advanced
[500 lines...]
```

**Solution**: Split into focused skills with cross-references.

```markdown
# Correct: Modular skills under 500 lines each

# python-testing/SKILL.md (400 lines)
# python-async/SKILL.md (350 lines)
# python-typing/SKILL.md (300 lines)

# Each references the others in "Works Well With"
```

### Anti-Pattern 2: Nested Sub-agent Spawning

**Problem**: Sub-agents spawning other sub-agents causes context issues.

```python
# Incorrect approach
def backend_agent_task():
    # Sub-agent spawning another sub-agent - BAD
    result = Task(subagent_type="database-expert", prompt="...")
    return result
```

**Solution**: All sub-agent delegation from main thread only.

```python
# Correct approach - main thread orchestrates all
analysis = Task(subagent_type="spec-builder", prompt="...")
database = Task(subagent_type="database-expert", prompt="...", context=analysis)
backend = Task(subagent_type="backend-expert", prompt="...", context=database)
```

### Anti-Pattern 3: Hardcoded Paths in Skills

**Problem**: Hardcoded paths break portability.

```markdown
# Incorrect
Load configuration from /Users/john/projects/myapp/config.yaml
```

**Solution**: Use relative paths and project references.

```markdown
# Correct
Load configuration from @config.yaml or $PROJECT_ROOT/config.yaml
```

---

## Integration Examples

### Complete SPEC Workflow

```python
# Full SPEC-First TDD Workflow

# Step 1: Plan - Create SPEC
plan_result = Task(
    subagent_type="spec-builder",
    prompt="Create SPEC for: User profile management with avatar upload",
    context={"project": "@CLAUDE.md"}
)

# Step 2: Clear context (after plan)
# /clear

# Step 3: Run - Implement with DDD
run_result = Task(
    subagent_type="ddd-implementer",
    prompt=f"Implement SPEC: {plan_result.spec_id}",
    context={"spec": plan_result}
)

# Step 4: Sync - Generate documentation
sync_result = Task(
    subagent_type="docs-manager",
    prompt=f"Generate docs for: {run_result.spec_id}",
    context={"implementation": run_result}
)
```

### Hook-Driven Quality Assurance

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "lint-check $FILE"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "validate-command $COMMAND"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "run-tests --affected $FILE"
          }
        ]
      }
    ]
  }
}
```

---

*For complete reference documentation, see the reference/ directory.*
