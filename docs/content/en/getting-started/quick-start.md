# Quick Start Guide

Experience the complete MoAI-ADK workflow in just 10 minutes. This guide walks you through creating your first API using the SPEC-First TDD methodology.

## Prerequisites

Before starting, ensure you have:

- ✅ **MoAI-ADK installed**: `uv tool install moai-adk`
- ✅ **Claude Code installed**: Available in your terminal
- ✅ **10 minutes of uninterrupted time**

## Step 0: Project Initialization (1 minute)

Create a new project and initialize it with Alfred.

```bash
# Create new project
moai-adk init hello-api
cd hello-api

# Start Claude Code
claude
```

Run the project initialization command in Claude Code:

```
/alfred:0-project
```

Alfred will ask a few questions:

- **Project name**: hello-api
- **Project goal**: Learn MoAI-ADK
- **Primary language**: python
- **Mode**: personal (for local development)

**Result**: Project is initialized with `.moai/` configuration, skills are loaded, and Alfred is ready.

## Step 1: SPEC Creation (2 minutes)

Use Alfred's spec-builder to define what to build.

```bash
/alfred:1-plan "GET /hello endpoint that accepts a name query parameter and returns a greeting"
```

### What Alfred Does

1. **Analyzes request** and extracts requirements

2. **Generates SPEC ID**: `HELLO-001`

3. **Creates EARS format specification**:

   ```yaml
   ---
   id: HELLO-001
   version: 0.0.1
   status: draft
   priority: high
   ---
   # @SPEC:EX-HELLO-001: Hello World API

   ## Ubiquitous Requirements
   - The system SHALL provide HTTP GET /hello endpoint

   ## Event-driven Requirements
   - WHEN name query parameter is provided, the system SHALL return "Hello, {name}!"
   - WHEN name is not provided, the system SHALL return "Hello, World!"

   ## Constraints
   - Name SHALL be limited to maximum 50 characters
   - Response SHALL be in JSON format
   ```

4. **Generates supporting files**:

   - `.moai/specs/SPEC-HELLO-001/spec.md` - Main specification
   - `.moai/specs/SPEC-HELLO-001/plan.md` - Implementation plan
   - Feature branch: `feature/SPEC-HELLO-001` (if in team mode)

### Verification

```bash
# Check SPEC was created
cat .moai/specs/SPEC-HELLO-001/spec.md

# Verify TAG assignment
rg '@SPEC:HELLO-001' -n
```

## Step 2: TDD Implementation (5 minutes)

Implement the API using Test-Driven Development.

```bash
/alfred:2-run HELLO-001
```

### Phase 1: 🔴 RED - Write Failing Tests

Alfred's `tdd-implementer` generates comprehensive tests first:

```python
# tests/test_hello.py
# @TEST:EX-HELLO-001 | SPEC: SPEC-HELLO-001.md

import pytest
from fastapi.testclient import TestClient
from src.hello.api import app

client = TestClient(app)

def test_hello_with_name_should_return_personalized_greeting():
    """WHEN name is provided, SHALL return 'Hello, {name}!'"""
    response = client.get("/hello?name=Alice")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, Alice!"}

def test_hello_without_name_should_return_default_greeting():
    """WHEN name is not provided, SHALL return 'Hello, World!'"""
    response = client.get("/hello")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, World!"}

def test_hello_with_long_name_should_return_400():
    """WHEN name exceeds 50 characters, SHALL return 400 error"""
    long_name = "a" * 51
    response = client.get(f"/hello?name={long_name}")
    assert response.status_code == 400
```

**Run tests** (they will fail - this is expected):

```bash
pytest tests/test_hello.py -v
# Result: FAILED - No module named 'src.hello.api'
```

**RED phase commit**:

```bash
git add tests/test_hello.py
git commit -m "🔴 test(HELLO-001): add failing hello API tests"
```

### Phase 2: 🟢 GREEN - Minimal Implementation

Alfred generates minimal code to pass tests:

```python
# src/hello/api.py
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI, HTTPException

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - Hello endpoint"""
    if len(name) > 50:
        raise HTTPException(status_code=400, detail="Name too long (max 50 chars)")
    return {"message": f"Hello, {name}!"}
```

**Run tests** (should now pass):

```bash
pytest tests/test_hello.py -v
# Result: PASSED - All 3 tests pass
```

**GREEN phase commit**:

```bash
git add src/hello/api.py
git commit -m "🟢 feat(HELLO-001): implement hello API"
```

### Phase 3: ♻️ REFACTOR - Improve Code Quality

Alfred improves the code by applying TRUST 5 principles:

```python
# src/hello/models.py
# @CODE:EX-HELLO-001:MODEL | SPEC: SPEC-HELLO-001.md

from pydantic import BaseModel, Field, validator

class HelloRequest(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - Request validation model"""
    name: str = Field(default="World", max_length=50, description="Name to greet")

    @validator('name')
    def validate_name(cls, v):
        if not v.strip():
            raise ValueError('Name cannot be empty')
        return v.strip()

class HelloResponse(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - Response model"""
    message: str = Field(description="Greeting message")
```

```python
# src/hello/api.py (refactored)
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI, HTTPException, Depends
from .models import HelloRequest, HelloResponse

app = FastAPI(title="Hello API", version="1.0.0")

@app.get("/hello", response_model=HelloResponse)
def hello(params: HelloRequest = Depends()):
    """@CODE:EX-HELLO-001:API - Hello endpoint with validation"""
    try:
        message = f"Hello, {params.name}!"
        return HelloResponse(message=message)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
```

**Verify tests still pass**:

```bash
pytest tests/test_hello.py -v
# Result: PASSED - All tests still pass
```

**REFACTOR phase commit**:

```bash
git add src/hello/models.py src/hello/api.py
git commit -m "♻️ refactor(HELLO-001): add models and improve validation"
```

## Step 3: Documentation Sync (1 minute)

Synchronize all documentation and validate the complete system.

```bash
/alfred:3-sync
```

### What Alfred Does

1. **Generates API documentation**:

   ````markdown
   # Hello API Documentation

   ## GET /hello

   Returns a personalized greeting message.

   ### Parameters
   - `name` (query, optional): Name to greet (default: "World", max 50 chars)

   ### Response
   - **200**: Success
     ```json
     {"message": "Hello, Alice!"}
   ````

   - **400**: Validation error

   ### Example

   ```bash
   curl "http://localhost:8000/hello?name=Alice"
   # → {"message": "Hello, Alice!"}
   ```

   ### Traceability

   - @SPEC:EX-HELLO-001 - Requirements
   - @TEST:EX-HELLO-001 - Tests
   - @CODE:EX-HELLO-001 - Implementation

   ```

   ```

2. **Updates README.md** (with API usage examples)

3. **Generates CHANGELOG.md** (with version history)

4. **Validates TAG chain integrity**:

   ```
   ✅ @SPEC:EX-HELLO-001 → .moai/specs/SPEC-HELLO-001/spec.md
   ✅ @TEST:EX-HELLO-001 → tests/test_hello.py
   ✅ @CODE:EX-HELLO-001 → src/hello/ (3 files)
   ✅ @DOC:EX-HELLO-001 → docs/api/hello.md (auto-generated)

   TAG chain integrity: 100%
   Orphaned TAGs: none
   ```

5. **Validates TRUST 5 compliance**:

   ```
   ✅ Test First: 100% coverage (3/3 tests passing)
   ✅ Readable: All functions < 50 lines
   ✅ Unified: Consistent FastAPI patterns
   ✅ Secured: Input validation implemented
   ✅ Trackable: All code tagged with @CODE:HELLO-001
   ```

## Step 4: Validation and Celebration (1 minute)

### Complete System Validation

```bash
# 1. Check TAG chain integrity
rg '@(SPEC|TEST|CODE|DOC):HELLO-001' -n
# All 4 TAG types should appear in output

# 2. Run tests
pytest tests/test_hello.py -v
# All tests should pass

# 3. Test API
uvicorn src.hello.api:app --reload &
curl "http://localhost:8000/hello?name=World"
# Should return: {"message": "Hello, World!"}

# 4. Check generated documentation
cat docs/api/hello.md
# Should contain complete API documentation
```

### Achievement Review

Successfully created:

```
hello-api/
├── .moai/specs/SPEC-HELLO-001/
│   ├── spec.md              ← Professional specification
│   └── plan.md              ← Implementation plan
├── tests/test_hello.py      ← 100% test coverage
├── src/hello/
│   ├── api.py               ← Production-quality implementation
│   ├── models.py            ← Data validation models
│   └── __init__.py
├── docs/api/hello.md        ← Auto-generated API docs
├── README.md                ← Updated with usage examples
├── CHANGELOG.md             ← Version history
└── .git/                    ← Clean git history with TDD commits
```

### Git History

```bash
git log --oneline | head -5
```

Expected output:

```
a1b2c3d ✅ sync(HELLO-001): update docs and changelog
d4e5f6c ♻️ refactor(HELLO-001): add models and improve validation
b2c3d4e 🟢 feat(HELLO-001): implement hello API
a3b4c5d 🔴 test(HELLO-001): add failing hello API tests
e5f6g7h 🌿 Create feature/SPEC-HELLO-001 branch
```

## What You Learned

### Concepts Experienced

✅ **SPEC-First**: Create clear requirements before coding
✅ **TDD**: RED → GREEN → REFACTOR cycle with 100% test coverage
✅ **@TAG System**: Complete traceability from requirements to documentation
✅ **TRUST 5**: Production-quality code with validation and error handling
✅ **Alfred Workflow**: Automated documentation and quality checks

### Skills Gained

- **EARS Syntax**: Writing structured requirements
- **Test Design**: Creating comprehensive test cases
- **API Development**: FastAPI best practices
- **Documentation**: Auto-generated, always-synchronized docs
- **Git Workflow**: Clean, traceable commit history

## Next Steps

### Continue Building

Add more features to your API:

```bash
# Add new endpoint
/alfred:1-plan "POST /greet endpoint that accepts JSON body"

# Or enhance existing functionality
/alfred:1-plan "Add language support to /hello endpoint"
```

### Explore Advanced Topics

- **[Project Configuration](guides/project/config.md)**: Customize project settings
- **[SPEC Writing](guides/specs/basics.md)**: Master EARS syntax
- **[TDD Patterns](guides/tdd/green.md)**: Learn advanced test strategies
- **[TAG System](reference/tags/index.md)**: Deep dive into traceability

### Join the Community

- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions and share experiences
- **Contributing**: Help improve MoAI-ADK

## Troubleshooting

### Common Issues

**Tests fail with import errors**:

```bash
# Install dependencies
uv add fastapi pytest
uv sync
```

**API doesn't start**:

```bash
# Check port and dependencies
lsof -i :8000
uvicorn src.hello.api:app --reload --port 8001
```

**Documentation not generated**:

```bash
# Run sync manually
/alfred:3-sync
```

### Get Help

```bash
# System diagnostics
moai-adk doctor

# Auto-create issue
/alfred:9-feedback
```

## Summary

In just 10 minutes, you've completed:

1. ✅ **Defined clear requirements** (using SPEC and EARS syntax)
2. ✅ **Implemented with TDD** (achieving 100% test coverage)
3. ✅ **Generated production-quality code** (with validation and error handling)
4. ✅ **Created complete documentation** (kept synchronized)
5. ✅ **Maintained full traceability** (with @TAG system)
6. ✅ **Followed best practices** (with TRUST 5 principles)

This is the power of MoAI-ADK: create reliable, maintainable, well-documented code faster than traditional methods. You're now ready to build complex applications with confidence! 🚀

Continue your journey with the [Alfred Workflow Guide](guides/alfred/index.md) or explore any specific topic that interests you.
