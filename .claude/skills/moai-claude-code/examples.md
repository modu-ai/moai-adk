# Claude Code Examples: Real-World Scenarios

> **Version**: 2.0.0 (2025-10-22)
> **Based on**: Claude Code Official Documentation (docs.claude.com), v0.6.0+ features

This document provides concrete examples for building and managing Claude Code agents, skills, plugins, and hooks.

---

## Table of Contents

1. [Creating Your First Skill](#1-creating-your-first-skill)
2. [Building a Custom Sub-agent](#2-building-a-custom-sub-agent)
3. [Implementing Hooks for Automation](#3-implementing-hooks-for-automation)
4. [Creating a Plugin Package](#4-creating-a-plugin-package)
5. [Skill with Multiple Supporting Files](#5-skill-with-multiple-supporting-files)
6. [Agent with Task Management](#6-agent-with-task-management)
7. [Pre-commit Hook Integration](#7-pre-commit-hook-integration)
8. [Session Management Automation](#8-session-management-automation)
9. [MCP Server Integration](#9-mcp-server-integration)
10. [Production Skill Deployment](#10-production-skill-deployment)

---

## 1. Creating Your First Skill

### Scenario: PDF Processing Skill

Create a Skill that helps Claude process PDF documents.

**Directory Structure**:
```
~/.claude/skills/pdf-processor/
├── SKILL.md
├── reference.md
└── scripts/
    └── extract_text.py
```

**SKILL.md**:
```markdown
---
name: PDF Processing
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, document extraction, form filling, or PDF manipulation.
allowed-tools: Read, Bash
---

# PDF Processing Skill

## What It Does

Extract text, tables, and metadata from PDF files using Python's PyPDF2 and pdfplumber libraries.

## When Claude Should Use This

- User asks to "read a PDF", "extract PDF text", "get PDF content"
- User mentions PDF files by path
- User wants to fill PDF forms or merge PDFs
- User requests table extraction from PDFs

## Dependencies

Ensure these Python packages are installed:
```bash
pip install PyPDF2 pdfplumber tabula-py
```

## Basic Usage

### Extract Text from PDF

```bash
python scripts/extract_text.py /path/to/document.pdf
```

### Extract Tables

```python
import pdfplumber

with pdfplumber.open("document.pdf") as pdf:
    for page in pdf.pages:
        tables = page.extract_tables()
        for table in tables:
            print(table)
```

## Common Patterns

### Pattern 1: Text Extraction
```python
from PyPDF2 import PdfReader

reader = PdfReader("example.pdf")
for page in reader.pages:
    print(page.extract_text())
```

### Pattern 2: Form Filling
```python
from PyPDF2 import PdfReader, PdfWriter

reader = PdfReader("form.pdf")
writer = PdfWriter()

writer.add_page(reader.pages[0])
writer.update_page_form_field_values(
    writer.pages[0],
    {"field_name": "value"}
)

with open("filled_form.pdf", "wb") as output:
    writer.write(output)
```

## Error Handling

- **Encrypted PDFs**: Use `reader.decrypt("password")`
- **Corrupted PDFs**: Catch `PdfReadError` exceptions
- **Missing dependencies**: Provide clear installation instructions

## References

- See [reference.md](reference.md) for API details
- Scripts in [scripts/](scripts/) for advanced examples
```

**scripts/extract_text.py**:
```python
#!/usr/bin/env python3
"""Extract text from PDF file."""

import sys
from PyPDF2 import PdfReader

def extract_text(pdf_path: str) -> str:
    """Extract all text from PDF."""
    try:
        reader = PdfReader(pdf_path)
        text = []
        for page in reader.pages:
            text.append(page.extract_text())
        return "\n\n".join(text)
    except Exception as e:
        return f"Error: {e}"

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: extract_text.py <pdf_path>")
        sys.exit(1)

    result = extract_text(sys.argv[1])
    print(result)
```

**Testing the Skill**:
```bash
# Install the skill (copy to ~/.claude/skills/)
mkdir -p ~/.claude/skills/pdf-processor
cp SKILL.md ~/.claude/skills/pdf-processor/

# Test in Claude Code
claude
> "Extract text from ~/Documents/report.pdf"
# Claude should automatically use the PDF Processing Skill
```

---

## 2. Building a Custom Sub-agent

### Scenario: Database Migration Agent

Create a specialized agent for database schema management.

**Directory Structure**:
```
.claude/agents/
├── db-migrator.md
└── db-migrator-config.json
```

**db-migrator.md**:
```markdown
# Database Migrator Agent

You are a specialized database migration agent. Your role is to help create, review, and execute database migrations safely.

## Responsibilities

1. **Analyze Schema Changes**: Review proposed migrations for potential issues
2. **Generate Migrations**: Create migration scripts from schema definitions
3. **Validate Migrations**: Check for common pitfalls (data loss, downtime)
4. **Rollback Planning**: Ensure every migration has a rollback strategy

## Guidelines

### Before Creating Migration

- [ ] Check for breaking changes
- [ ] Plan for data transformation
- [ ] Consider rollback scenario
- [ ] Estimate downtime (if any)
- [ ] Test with production-like data

### Migration Best Practices

1. **Additive Changes**: Add new columns as nullable first
2. **Backfill Data**: Use separate step for large data migrations
3. **Indexes**: Create indexes concurrently (PostgreSQL)
4. **Constraints**: Add non-blocking constraints when possible

### Red Flags

- ⚠️ Dropping columns without verification
- ⚠️ Changing column types (potential data loss)
- ⚠️ Adding NOT NULL to existing column
- ⚠️ Renaming tables/columns without aliases

## Example Workflow

**User Request**: "Add email verification field to users table"

**Your Response**:
```sql
-- Migration: add_email_verified_column
-- Step 1: Add column as nullable (safe, no downtime)
ALTER TABLE users
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- Step 2: Backfill existing users
UPDATE users SET email_verified = FALSE WHERE email_verified IS NULL;

-- Step 3: Add NOT NULL constraint (after backfill)
ALTER TABLE users
ALTER COLUMN email_verified SET NOT NULL;

-- Rollback
ALTER TABLE users DROP COLUMN email_verified;
```

**Questions to Ask**:
- Should existing users be marked as verified or unverified?
- Do we need to send verification emails to existing users?
- Is there a specific deployment window?

## Integration with Project

- Read existing migration files to understand schema
- Follow project's migration naming convention
- Use project's ORM migration tool (Alembic, Django, TypeORM, etc.)
- Run tests after generating migrations

## Safety Checklist

Before executing migration:
- [ ] Reviewed by DBA (if production)
- [ ] Tested with production-like dataset
- [ ] Rollback plan documented
- [ ] Estimated execution time < 5 minutes (or scheduled maintenance)
- [ ] Backup completed

## Output Format

Provide migrations as:
1. Migration file with timestamp prefix
2. Rollback steps
3. Testing instructions
4. Risk assessment
```

**db-migrator-config.json** (optional):
```json
{
  "name": "Database Migrator",
  "description": "Specialized agent for database schema changes",
  "model": "claude-sonnet-4.5",
  "temperature": 0.2,
  "custom_instructions": "Focus on safety and reversibility. Always provide rollback plans."
}
```

**Usage**:
```bash
# Invoke the agent
claude
> "@db-migrator Add a status column to the orders table"

# Agent activates with database migration context
```

---

## 3. Implementing Hooks for Automation

### Scenario: Pre-commit Quality Checks

Implement hooks that run quality checks before allowing commits.

**Hook Configuration** (~/.claude/settings.json):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/pre-commit-check.sh"
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
            "command": "bash ~/.claude/hooks/post-write-format.sh \"{{tool_input.file_path}}\""
          }
        ]
      },
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/post-edit-lint.sh \"{{tool_input.file_path}}\""
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/session-start-status.sh"
          }
        ]
      }
    ]
  }
}
```

**~/.claude/hooks/pre-commit-check.sh**:
```bash
#!/bin/bash
set -e

# Extract git command from Claude's bash invocation
GIT_COMMAND="$1"

# Only run on git commit commands
if [[ ! "$GIT_COMMAND" =~ git[[:space:]]+commit ]]; then
    exit 0  # Allow non-commit commands
fi

echo "🔍 Running pre-commit checks..."

# 1. Check for merge conflicts
if grep -r "<<<<<<< HEAD" .; then
    echo "❌ Merge conflicts detected. Resolve before committing."
    exit 1
fi

# 2. Run linter
if command -v ruff &> /dev/null; then
    if ! ruff check .; then
        echo "❌ Linter failed. Fix issues before committing."
        exit 1
    fi
fi

# 3. Run tests
if [ -f "pyproject.toml" ] || [ -f "pytest.ini" ]; then
    if ! pytest --quiet; then
        echo "❌ Tests failed. Fix tests before committing."
        exit 1
    fi
fi

# 4. Check coverage
if command -v pytest &> /dev/null; then
    COVERAGE=$(pytest --cov=src --cov-report=term-missing | grep "TOTAL" | awk '{print $4}' | sed 's/%//')
    if [ "${COVERAGE%.*}" -lt 85 ]; then
        echo "⚠️ Warning: Coverage ${COVERAGE}% < 85% target"
        # Don't block, just warn
    fi
fi

echo "✅ Pre-commit checks passed!"
exit 0
```

**~/.claude/hooks/post-write-format.sh**:
```bash
#!/bin/bash
set -e

FILE_PATH="$1"

# Skip if not a code file
if [[ ! "$FILE_PATH" =~ \.(py|js|ts|tsx|jsx)$ ]]; then
    exit 0
fi

echo "🎨 Auto-formatting ${FILE_PATH}..."

# Python
if [[ "$FILE_PATH" =~ \.py$ ]]; then
    if command -v ruff &> /dev/null; then
        ruff format "$FILE_PATH"
        echo "✅ Formatted with ruff"
    fi
fi

# TypeScript/JavaScript
if [[ "$FILE_PATH" =~ \.(js|ts|tsx|jsx)$ ]]; then
    if command -v prettier &> /dev/null; then
        prettier --write "$FILE_PATH"
        echo "✅ Formatted with Prettier"
    fi
fi

exit 0
```

**~/.claude/hooks/session-start-status.sh**:
```bash
#!/bin/bash

echo "📊 Session Status Report"
echo "========================"
echo ""

# Git status
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo "📁 Git Branch: $(git branch --show-current)"
    echo "📝 Uncommitted Changes: $(git status --short | wc -l)"
    echo ""
fi

# Python environment
if [ -d ".venv" ]; then
    echo "🐍 Virtual Environment: .venv (active)"
elif [ -d "venv" ]; then
    echo "🐍 Virtual Environment: venv (active)"
else
    echo "⚠️ No virtual environment detected"
fi

echo ""
echo "✅ Session initialized"
exit 0
```

**Testing Hooks**:
```bash
# View configured hooks
claude /hooks

# Test hook execution
claude
> "Commit these changes with message 'Add feature'"
# Pre-commit hook runs automatically
```

---

## 4. Creating a Plugin Package

### Scenario: Testing Utilities Plugin

Bundle commands, agents, and skills into a single plugin.

**Directory Structure**:
```
my-testing-plugin/
├── plugin.json
├── commands/
│   ├── run-tests.md
│   └── coverage-report.md
├── agents/
│   └── test-analyzer.md
├── skills/
│   └── test-patterns/
│       └── SKILL.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "Testing Utilities",
  "version": "1.0.0",
  "description": "Comprehensive testing tools for Claude Code",
  "author": "Your Name",
  "repository": "https://github.com/username/testing-plugin",
  "commands": [
    {
      "name": "run-tests",
      "file": "commands/run-tests.md",
      "description": "Run test suite with coverage"
    },
    {
      "name": "coverage-report",
      "file": "commands/coverage-report.md",
      "description": "Generate detailed coverage report"
    }
  ],
  "agents": [
    {
      "name": "test-analyzer",
      "file": "agents/test-analyzer.md",
      "description": "Analyze test quality and suggest improvements"
    }
  ],
  "skills": [
    {
      "name": "test-patterns",
      "directory": "skills/test-patterns"
    }
  ]
}
```

**commands/run-tests.md**:
```markdown
# Run Tests Command

Run project test suite with coverage reporting.

## Detection

Auto-detect test framework:
- Python: pytest, unittest
- JavaScript/TypeScript: Jest, Vitest, Mocha
- Go: go test
- Rust: cargo test

## Execution

1. Detect project language and test framework
2. Run tests with coverage
3. Display summary
4. Highlight failing tests

## Example Output

```
🧪 Running Tests...
Framework: pytest
Coverage: 87% (Target: 85%) ✅

Tests:
✅ 45 passed
❌ 2 failed
⏭️  3 skipped

Failed Tests:
- test_auth.py::test_invalid_token
- test_api.py::test_rate_limiting

Next Steps:
1. Review failed tests above
2. Run with -v for verbose output
3. Check coverage report: /coverage-report
```
```

**skills/test-patterns/SKILL.md**:
```markdown
---
name: Test Patterns
description: Common testing patterns and best practices across languages. Use when writing tests, reviewing test quality, or discussing testing strategies.
allowed-tools: Read, Write, Edit
---

# Test Patterns Skill

## AAA Pattern (Arrange-Act-Assert)

```python
def test_user_login():
    # Arrange
    user = User(email="test@example.com", password="secret")
    user.save()

    # Act
    result = login(user.email, "secret")

    # Assert
    assert result.success is True
    assert result.user.email == "test@example.com"
```

## Parameterized Tests

**Python (pytest)**:
```python
@pytest.mark.parametrize("input,expected", [
    (2, 4),
    (3, 9),
    (4, 16),
])
def test_square(input, expected):
    assert square(input) == expected
```

**TypeScript (Vitest)**:
```typescript
test.each([
  [2, 4],
  [3, 9],
  [4, 16],
])('square(%i) = %i', (input, expected) => {
  expect(square(input)).toBe(expected);
});
```

## Mocking Patterns

**Dependency Injection**:
```python
def test_send_email(mocker):
    # Mock external service
    mock_smtp = mocker.patch('smtplib.SMTP')

    # Execute
    send_email("test@example.com", "Subject", "Body")

    # Verify
    mock_smtp.assert_called_once()
```

## Edge Cases Checklist

- [ ] Null/undefined inputs
- [ ] Empty collections
- [ ] Boundary values (0, -1, MAX)
- [ ] Concurrent access
- [ ] Network failures
- [ ] Invalid credentials
```

**Installation**:
```bash
# Publish to GitHub
git init
git add .
git commit -m "Initial plugin release"
git tag v1.0.0
git push origin main --tags

# Install via Claude Code
claude /plugin install https://github.com/username/testing-plugin
```

---

## 5. Skill with Multiple Supporting Files

### Scenario: Kubernetes Deployment Skill

**Directory Structure**:
```
~/.claude/skills/k8s-deploy/
├── SKILL.md
├── reference.md
├── examples.md
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
└── scripts/
    ├── deploy.sh
    └── rollback.sh
```

**SKILL.md**:
```markdown
---
name: Kubernetes Deployment
description: Deploy applications to Kubernetes clusters with best practices. Use when deploying to k8s, creating k8s manifests, or troubleshooting k8s deployments.
allowed-tools: Read, Write, Bash
---

# Kubernetes Deployment Skill

## Quick Start

Deploy an application with standard templates:

```bash
# 1. Generate manifests
bash scripts/deploy.sh generate my-app nginx:latest 3

# 2. Review manifests in generated/
# 3. Apply to cluster
kubectl apply -f generated/
```

## Templates

Standard Kubernetes manifests are available in `templates/`:
- `deployment.yaml`: Pod deployment configuration
- `service.yaml`: Service exposure
- `ingress.yaml`: External access routing

See [examples.md](examples.md) for complete deployment scenarios.

## Best Practices

1. **Resource Limits**: Always set CPU/memory limits
2. **Health Checks**: Define liveness and readiness probes
3. **Rolling Updates**: Use RollingUpdate strategy
4. **Secrets**: Use Kubernetes Secrets, not env vars
5. **Labels**: Consistent labeling for all resources

## Troubleshooting

```bash
# Check pod status
kubectl get pods

# View logs
kubectl logs <pod-name>

# Describe for events
kubectl describe pod <pod-name>

# Quick rollback
bash scripts/rollback.sh my-app
```

## References

- [reference.md](reference.md): Detailed API specs
- [examples.md](examples.md): Real-world scenarios
```

**templates/deployment.yaml**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{APP_NAME}}
  labels:
    app: {{APP_NAME}}
spec:
  replicas: {{REPLICAS}}
  selector:
    matchLabels:
      app: {{APP_NAME}}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: {{APP_NAME}}
    spec:
      containers:
      - name: {{APP_NAME}}
        image: {{IMAGE}}
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

**scripts/deploy.sh**:
```bash
#!/bin/bash
set -e

APP_NAME="$1"
IMAGE="$2"
REPLICAS="${3:-3}"

if [ "$#" -lt 2 ]; then
    echo "Usage: deploy.sh <app-name> <image> [replicas]"
    exit 1
fi

echo "🚀 Generating Kubernetes manifests..."

# Create output directory
mkdir -p generated

# Generate deployment
sed -e "s/{{APP_NAME}}/$APP_NAME/g" \
    -e "s|{{IMAGE}}|$IMAGE|g" \
    -e "s/{{REPLICAS}}/$REPLICAS/g" \
    templates/deployment.yaml > generated/deployment.yaml

# Generate service
sed -e "s/{{APP_NAME}}/$APP_NAME/g" \
    templates/service.yaml > generated/service.yaml

echo "✅ Manifests generated in generated/"
echo ""
echo "Next steps:"
echo "  1. Review: ls -la generated/"
echo "  2. Apply: kubectl apply -f generated/"
echo "  3. Verify: kubectl get pods -l app=$APP_NAME"
```

---

## 6. Agent with Task Management

### Scenario: Feature Implementation Agent

Agent that breaks down features into tasks and tracks progress.

**feature-builder.md**:
```markdown
# Feature Builder Agent

You are a feature implementation specialist. You break down feature requests into actionable tasks, track progress, and ensure quality standards.

## Workflow

### Phase 1: Planning

When user requests a feature:

1. **Clarify Requirements**
   - Ask clarifying questions
   - Identify edge cases
   - Define success criteria

2. **Create Task List** (use TodoWrite tool):
   ```
   - [ ] Write SPEC document
   - [ ] Design API endpoints
   - [ ] Create database schema
   - [ ] Implement business logic
   - [ ] Write unit tests
   - [ ] Write integration tests
   - [ ] Update documentation
   - [ ] Code review
   ```

3. **Estimate Complexity**
   - Small: < 2 hours
   - Medium: 2-8 hours
   - Large: > 8 hours (break down further)

### Phase 2: Implementation

For each task:

1. **Mark as in_progress** before starting
2. **Follow TDD**:
   - Write failing test (RED)
   - Implement feature (GREEN)
   - Refactor (REFACTOR)
3. **Mark as completed** when done

### Phase 3: Review

Before marking feature complete:

- [ ] All tests passing
- [ ] Coverage ≥ 85%
- [ ] Documentation updated
- [ ] No linter warnings
- [ ] Code reviewed

## Example Interaction

**User**: "Add user profile picture upload"

**Agent**:
```markdown
I'll help implement user profile picture upload. Let me clarify a few things:

1. **File Size Limit**: What's the maximum file size? (Recommended: 5MB)
2. **File Types**: Which formats? (Recommended: JPEG, PNG, WebP)
3. **Storage**: Where should images be stored? (S3, local filesystem)
4. **Processing**: Need thumbnails/resizing?

Based on standard practices, I'll create a task plan:

**Tasks**:
1. Add profile_picture_url column to users table
2. Create file upload endpoint (POST /api/users/:id/profile-picture)
3. Implement image validation (size, type)
4. Add S3 upload integration
5. Generate thumbnail (256x256)
6. Update user profile API response
7. Write tests for upload endpoint
8. Write tests for validation
9. Update API documentation
10. Add error handling

**Estimated Time**: Medium (4-6 hours)

Ready to proceed? Any adjustments needed?
```

## Quality Standards

Every feature must meet:

- **SOLID Principles**: No violations
- **Test Coverage**: ≥ 85%
- **Documentation**: API docs + inline comments
- **Error Handling**: Graceful degradation
- **Security**: Input validation, auth checks

## Integration with Project

- Read existing codebase patterns
- Follow project conventions
- Use project's testing framework
- Maintain consistent code style
```

**Usage**:
```bash
claude
> "@feature-builder Add password reset functionality"

# Agent creates task list and guides implementation
```

---

## 7. Pre-commit Hook Integration

### Scenario: Enforce Quality Gates Before Commit

**~/.claude/hooks/pre-commit-quality-gate.sh**:
```bash
#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔒 Quality Gate Check"
echo "===================="
echo ""

FAILED=0

# 1. Linter Check
echo "🔍 Running linter..."
if command -v ruff &> /dev/null; then
    if ruff check .; then
        echo -e "${GREEN}✅ Linter passed${NC}"
    else
        echo -e "${RED}❌ Linter failed${NC}"
        FAILED=1
    fi
elif command -v eslint &> /dev/null; then
    if eslint .; then
        echo -e "${GREEN}✅ Linter passed${NC}"
    else
        echo -e "${RED}❌ Linter failed${NC}"
        FAILED=1
    fi
fi

# 2. Type Check
echo ""
echo "🔧 Running type check..."
if [ -f "tsconfig.json" ]; then
    if tsc --noEmit; then
        echo -e "${GREEN}✅ Type check passed${NC}"
    else
        echo -e "${RED}❌ Type check failed${NC}"
        FAILED=1
    fi
elif command -v mypy &> /dev/null; then
    if mypy .; then
        echo -e "${GREEN}✅ Type check passed${NC}"
    else
        echo -e "${RED}❌ Type check failed${NC}"
        FAILED=1
    fi
fi

# 3. Test Suite
echo ""
echo "🧪 Running tests..."
if [ -f "pytest.ini" ] || [ -f "pyproject.toml" ]; then
    if pytest --quiet; then
        echo -e "${GREEN}✅ Tests passed${NC}"
    else
        echo -e "${RED}❌ Tests failed${NC}"
        FAILED=1
    fi
elif [ -f "package.json" ]; then
    if npm test -- --run; then
        echo -e "${GREEN}✅ Tests passed${NC}"
    else
        echo -e "${RED}❌ Tests failed${NC}"
        FAILED=1
    fi
fi

# 4. Coverage Check
echo ""
echo "📊 Checking coverage..."
if command -v pytest &> /dev/null; then
    COVERAGE=$(pytest --cov=src --cov-report=term-missing --quiet | grep "TOTAL" | awk '{print $4}' | sed 's/%//')
    if [ "${COVERAGE%.*}" -ge 85 ]; then
        echo -e "${GREEN}✅ Coverage: ${COVERAGE}% (≥85%)${NC}"
    else
        echo -e "${YELLOW}⚠️  Coverage: ${COVERAGE}% (<85%)${NC}"
        echo -e "${YELLOW}   This is a warning, not blocking${NC}"
    fi
fi

# 5. Security Scan
echo ""
echo "🔐 Security scan..."
if command -v bandit &> /dev/null; then
    if bandit -r . -ll --quiet; then
        echo -e "${GREEN}✅ No security issues${NC}"
    else
        echo -e "${RED}❌ Security issues found${NC}"
        FAILED=1
    fi
fi

echo ""
echo "===================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All quality gates passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Quality gates failed. Fix issues before committing.${NC}"
    exit 1
fi
```

---

## 8. Session Management Automation

### Scenario: Automatic Project Context Loading

**~/.claude/hooks/session-start-context.sh**:
```bash
#!/bin/bash

echo "📚 Loading Project Context..."
echo ""

# Detect project type
if [ -f "package.json" ]; then
    echo "🟢 Node.js Project Detected"
    echo "   Framework: $(jq -r '.dependencies | keys[]' package.json | grep -E '(react|vue|angular|next)' | head -1 || echo 'Unknown')"
    echo "   Package Manager: $([ -f 'package-lock.json' ] && echo 'npm' || echo 'yarn/pnpm')"
elif [ -f "pyproject.toml" ]; then
    echo "🐍 Python Project Detected"
    echo "   Package Manager: $(grep -q 'uv' pyproject.toml && echo 'uv' || echo 'pip')"
    if [ -f "src/__init__.py" ]; then
        echo "   Structure: src layout"
    fi
elif [ -f "go.mod" ]; then
    echo "🐹 Go Project Detected"
    echo "   Module: $(head -1 go.mod | awk '{print $2}')"
elif [ -f "Cargo.toml" ]; then
    echo "🦀 Rust Project Detected"
    echo "   Package: $(grep '^name' Cargo.toml | cut -d'"' -f2)"
fi

echo ""

# Load recent commands/patterns
if [ -f ".claude/context.md" ]; then
    echo "📝 Project Context Available: .claude/context.md"
fi

# Check for MoAI-ADK
if [ -f ".moai/config.json" ]; then
    echo "🎯 MoAI-ADK Project"
    echo "   Mode: $(jq -r '.mode' .moai/config.json)"
fi

echo ""
echo "✅ Context loaded. Type /help for available commands."
```

---

## 9. MCP Server Integration

### Scenario: Custom MCP Server for Project-Specific Tools

**mcp-server-config.json** (in ~/.claude/settings.json):
```json
{
  "mcpServers": {
    "project-tools": {
      "command": "node",
      "args": ["mcp-servers/project-tools-server.js"],
      "env": {
        "PROJECT_ROOT": "/path/to/project"
      }
    }
  }
}
```

**mcp-servers/project-tools-server.js**:
```javascript
#!/usr/bin/env node

const { MCPServer } = require('@anthropic/mcp-sdk');

const server = new MCPServer({
  name: 'Project Tools',
  version: '1.0.0'
});

// Register custom tool
server.registerTool({
  name: 'run_build',
  description: 'Run project build process',
  parameters: {
    type: 'object',
    properties: {
      mode: {
        type: 'string',
        enum: ['development', 'production'],
        description: 'Build mode'
      }
    }
  },
  handler: async ({ mode }) => {
    // Execute build
    const { execSync } = require('child_process');
    try {
      const output = execSync(`npm run build:${mode}`, {
        encoding: 'utf-8',
        cwd: process.env.PROJECT_ROOT
      });
      return {
        success: true,
        output: output
      };
    } catch (error) {
      return {
        success: false,
        error: error.message
      };
    }
  }
});

server.start();
```

---

## 10. Production Skill Deployment

### Scenario: Publishing a Skill to GitHub

**Complete Workflow**:

**1. Create Repository Structure**:
```bash
mkdir my-awesome-skill
cd my-awesome-skill

# Create files
cat > SKILL.md << 'EOF'
---
name: My Awesome Skill
description: Does amazing things for Claude Code users
allowed-tools: Read, Bash
---

# My Awesome Skill

[Your skill content]
EOF

cat > README.md << 'EOF'
# My Awesome Skill

Claude Code Skill for [purpose].

## Installation

```bash
# Install via GitHub
mkdir -p ~/.claude/skills
git clone https://github.com/username/my-awesome-skill ~/.claude/skills/my-awesome-skill
```

## Usage

[Usage instructions]

## License

MIT
EOF

cat > LICENSE << 'EOF'
MIT License
[Your license text]
EOF
```

**2. Test Locally**:
```bash
# Symlink to Claude Code skills directory
ln -s $(pwd) ~/.claude/skills/my-awesome-skill

# Test in Claude Code
claude
> "Test my awesome skill"
```

**3. Publish to GitHub**:
```bash
git init
git add .
git commit -m "Initial skill release"
git tag v1.0.0
git remote add origin https://github.com/username/my-awesome-skill.git
git push origin main --tags
```

**4. Documentation**:
Create comprehensive README with:
- Installation instructions
- Usage examples
- Configuration options
- Troubleshooting guide
- Contribution guidelines

**5. Release Notes** (CHANGELOG.md):
```markdown
# Changelog

## v1.0.0 (2025-10-22)

### Added
- Initial release
- Core functionality for [feature]
- Examples and documentation

### Known Issues
- None
```

---

## Integration with MoAI-ADK

All these examples integrate seamlessly with MoAI-ADK workflows:

```bash
# Create a skill via Alfred
/alfred:1-plan "Create a new Claude Code Skill for API testing"

# Implement via TDD
/alfred:2-run SPEC-SKILL-001

# Deploy and document
/alfred:3-sync auto
```

---

## References

- [Claude Code Official Documentation](https://docs.claude.com/en/docs/claude-code)
- [Claude Skills Guide](https://docs.claude.com/en/docs/claude-code/skills)
- [Claude Hooks Guide](https://docs.claude.com/en/docs/claude-code/hooks-guide)

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Part of**: MoAI-ADK Claude Code Operations
**Related Skills**: `moai-skill-factory`, `moai-foundation-langs`
