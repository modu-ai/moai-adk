---
name: "moai-alfred-practices"
version: "4.0.0"
created: 2025-11-05
updated: 2025-11-12
status: stable
description: Enterprise practical workflows, context engineering strategies, JIT (Just-In-Time) retrieval optimization, real-world execution examples, debugging patterns, and moai-adk workflow mastery; activates for workflow pattern learning, context optimization, debugging issue resolution, feature implementation end-to-end, and team knowledge transfer
keywords: ['workflow-patterns', 'context-engineering', 'jit-retrieval', 'agent-usage', 'debugging-patterns', 'feature-implementation', 'practical-examples', 'moai-adk-mastery', 'knowledge-transfer', 'enterprise-workflows']
allowed-tools: 
  - Read
  - Glob
  - Bash
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Practical Workflows & Context Engineering v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-practices |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Focus** | Practical execution patterns, real-world scenarios |
| **Auto-load** | When workflow guidance or debugging help needed |
| **Included Patterns** | 15+ real-world scenarios |
| **Lines of Content** | 950+ with 20+ production examples |
| **Progressive Disclosure** | 3-level (quick-patterns, scenarios, advanced) |

---

## What It Does

Provides practical workflows, context engineering strategies, real-world execution examples, and debugging solutions for moai-adk. Covers JIT context management, efficient agent usage, SPEC→TDD→Sync execution, and common problem resolution.

---

## JIT (Just-In-Time) Context Strategy

### Principle: Load Only What's Needed Now

```
Traditional (overload):
  Load entire codebase
  → Context window fills immediately
  → Limited reasoning capacity
  → Slow, inefficient

JIT (optimized):
  Load core entry points
  → Identify specific function/module
  → Load only that section
  → Cache in thread context
  → Reuse for related tasks
  → Minimal context waste
```

### Practice 1: Core Module Mapping

```bash
# 1. Get high-level structure
find src/ -type f -name "*.py" | wc -l
# Output: 145 files total

# 2. Identify entry points (only 3-5 files)
find src/ -name "__main__.py" -o -name "main.py" -o -name "run.py"

# 3. Load entry point + immediate dependencies
Glob("src/{**/}*.py") 
# Load only files referenced by entry point

# 4. Cache in Task() context for reuse
Task(prompt="Task 1 using mapped modules")
Task(prompt="Task 2 reuses cached context")
```

### Practice 2: Dependency Tree Navigation

```
Project Root
├─ src/
│  ├─ __init__.py              ← Entry point #1
│  ├─ main.py                  ← Entry point #2
│  ├─ core/
│  │  ├─ domain.py             ← Core models
│  │  ├─ repository.py         ← Data access
│  │  └─ service.py            ← Business logic
│  └─ api/
│     ├─ routes.py             ← API endpoints
│     └─ handlers.py           ← Request handlers

Load strategy:
1. Load main.py + __init__.py (entry points)
2. When modifying API → Load api/ subtree
3. When fixing business logic → Load core/service.py
4. Cache all loaded files in context
5. Share context between related tasks
```

### Practice 3: Context Reuse Across Tasks

```python
# Task 1: Understand module structure
analysis = Task({
  prompt="Map src/ directory structure, identify entry points, list dependencies"
})

# Task 2: Reuse analysis for implementation
implementation = Task({
  prompt=f"""Using this structure:
{analysis}

Now implement feature X...
"""
})

# Task 3: Reuse analysis for testing
testing = Task({
  prompt=f"""Using this structure:
{analysis}

Write tests for feature X...
"""
})

# Result: No re-mapping, efficient context reuse
```

---

## SPEC → TDD → Sync Execution Pattern

### Step 1: Create SPEC with `/alfred:1-plan`

```bash
/alfred:1-plan "Add user authentication with JWT"

# This creates:
# .moai/specs/SPEC-042/spec.md (full requirements)
# feature/SPEC-042 (git branch)
# Track with TodoWrite
```

### Step 2: Implement with `/alfred:2-run SPEC-042`

```
RED:        Test agent writes failing tests
  ↓
GREEN:      Implementer agent creates minimal code
  ↓
REFACTOR:   Quality agent improves code
  ↓
Repeat TDD cycle for each feature component
  ↓
All tests passing, coverage ≥85%
```

### Step 3: Sync with `/alfred:3-sync auto SPEC-042`

```
Updates:
  ✓ Documentation
  ✓ Test coverage metrics
  ✓ Creates PR to develop
  ✓ Auto-validation of quality gates
```

---

## Debugging Pattern: Issue → Root Cause → Fix

### Step 1: Triage & Understand

```
Error message: "Cannot read property 'user_id' of undefined"

Questions:
- When does it occur? (always, intermittently, specific scenario)
- Which code path? (which endpoint/function)
- What's the state? (what data led to this)
- What changed recently? (revert to narrow down)
```

### Step 2: Isolate Root Cause

```python
# Method 1: Binary search
# Is it in API layer? → Yes
# Is it in route handler? → No
# Is it in service layer? → Yes
# Is it in this function? → Narrow down

# Method 2: Add logging
logger.debug(f"user_id = {user_id}")  # Check where it becomes undefined

# Method 3: Test locally
# Reproduce with minimal example
# Add breakpoint in debugger
# Step through execution
```

### Step 3: Fix with Tests

```python
# RED: Write failing test
def test_handles_missing_user_id():
    """Should handle case when user_id is undefined."""
    assert get_user(None) raises ValueError

# GREEN: Minimal fix
def get_user(user_id):
    if not user_id:
        raise ValueError("user_id required")
    return fetch_user(user_id)

# REFACTOR: Improve
def get_user(user_id: int) -> User:
    """Get user by ID.
    
    Args:
        user_id: User identifier
        
    Raises:
        ValueError: If user_id is None or invalid
    """
    if not user_id or user_id <= 0:
        raise ValueError(f"Invalid user_id: {user_id}")
    return self.user_repo.find(user_id)
```

---

## 5 Real-World Scenarios

### Scenario 1: Feature Implementation (2-3 hours)

```
1. Create SPEC:     /alfred:1-plan "Add user dashboard"
2. Clarify details: AskUserQuestion (which data to show?)
3. Implement:       /alfred:2-run SPEC-XXX (TDD cycle)
4. Document:        /alfred:3-sync auto SPEC-XXX
5. Result:          Production-ready feature
```

### Scenario 2: Bug Investigation (1-2 hours)

```
1. Reproduce:       Create minimal test case
2. Isolate:         Narrow down affected code
3. Debug:           Add logging, trace execution
4. Fix:             TDD RED→GREEN→REFACTOR
5. Validate:        Ensure tests pass, regression tests
```

### Scenario 3: Large Refactoring (4-8 hours)

```
1. Analyze:         Map current code structure
2. Plan:            Design new structure with trade-offs
3. Clone pattern:   Create autonomous agents for parallel refactoring
4. Integrate:       Verify all pieces work together
5. Test:            Comprehensive test coverage
```

### Scenario 4: Performance Optimization (2-4 hours)

```
1. Profile:         Identify bottleneck with profiler
2. Analyze:         Understand performance characteristics
3. Design:          Plan optimization approach
4. Implement:       TDD RED→GREEN→REFACTOR
5. Validate:        Benchmark before/after
```

### Scenario 5: Multi-Team Coordination (ongoing)

```
1. SPEC clarity:    AskUserQuestion for ambiguous requirements
2. Agent routing:   Delegate to specialist teams
3. Progress tracking: TodoWrite for coordination
4. Integration:     Verify components work together
5. Documentation:   Central SPEC as source of truth
```

---

## Context Budget Optimization

```
Typical project context:
  - Config files: ~50 tokens
  - .moai/ structure: ~100 tokens
  - Entry points (3-5 files): ~500 tokens
  - SPEC document: ~200 tokens
  → Total: ~850 tokens per session

Reusable context:
  - Load once per session
  - Share across 5-10 tasks
  - Saves: 3,500-8,500 tokens per session
  - Result: More reasoning capacity
```

---

## Best Practices

### DO
- ✅ Load entry points first (3-5 files)
- ✅ Identify dependencies before deep dive
- ✅ Reuse analyzed context across tasks
- ✅ Cache intermediate results in Task context
- ✅ Follow SPEC → TDD → Sync workflow
- ✅ Track progress with TodoWrite
- ✅ Ask for clarification (AskUserQuestion)
- ✅ Test before declaring done

### DON'T
- ❌ Load entire codebase at once
- ❌ Reanalyze same code multiple times
- ❌ Skip SPEC clarification (causes rework)
- ❌ Write code without tests
- ❌ Ignore error messages
- ❌ Assume context understanding
- ❌ Skip documentation updates
- ❌ Commit without running tests

---

## Related Skills

- `moai-alfred-agent-guide` (Agent orchestration patterns)
- `moai-alfred-clone-pattern` (Complex task delegation)
- `moai-essentials-debug` (Debugging techniques)

---

**For detailed workflow examples**: [reference.md](reference.md)  
**For real-world scenarios**: [examples.md](examples.md)  
**Last Updated**: 2025-11-12  
**Status**: Production Ready (Enterprise v4.0.0)
