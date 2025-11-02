# TodoWrite Auto-Initialization Pattern

> **Version**: 0.8.0+
> **Status**: Active
> **Updated**: 2025-11-02

## Overview

**OLD Pattern** (manual initialization):
```python
# User had to manually initialize TodoWrite
TodoWrite([
    {"content": "Analyze project", "status": "pending"},
    {"content": "Create SPEC", "status": "pending"}
])
```

**NEW Pattern** (auto-initialization from Plan agent):
```python
# Plan agent generates task list
plan = Task(
    subagent_type="Plan",
    prompt="Create structured plan for: {request}"
)

# Alfred auto-initializes TodoWrite from plan
TodoWrite(plan.tasks)  # Automatic from Plan agent output
```

## Integration with Alfred 4-Step Workflow

### Step 1: Intent Understanding
- **Goal**: Clarify user intent before any action
- **TodoWrite**: NO initialization
- **Output**: Clear understanding of requirements

### Step 2: Plan Creation (NEW: AUTO-INIT)
- **Goal**: Analyze tasks and initialize progress tracking
- **Action**: Invoke Plan Agent to decompose tasks
- **NEW**: **AUTO-INITIALIZE TodoWrite from Plan output**
- **Output**: Structured task breakdown + initialized TodoWrite

### Step 3: Task Execution
- **Goal**: Execute tasks with transparent progress tracking
- **Action**: Execute tasks in order, updating TodoWrite status
- **TodoWrite**: Update: pending → in_progress → completed
- **Output**: Completed work, all todos marked done

### Step 4: Report & Commit
- **Goal**: Document work and create git history
- **Action**: Generate report (if requested) and commit
- **TodoWrite**: All tasks should be "completed"
- **Output**: Git history, documentation

## Plan Agent Output Format

Plan agent should return tasks in TodoWrite-compatible format:

```json
{
    "tasks": [
        {
            "content": "Analyze SPEC requirements",
            "activeForm": "Analyzing SPEC requirements",
            "status": "pending"
        },
        {
            "content": "Write test cases",
            "activeForm": "Writing test cases",
            "status": "pending"
        },
        {
            "content": "Implement feature",
            "activeForm": "Implementing feature",
            "status": "pending"
        }
    ]
}
```

## Implementation Pattern

### In /alfred:1-plan Command

```python
# STEP 2: Plan Creation (with AUTO-INIT)
plan_result = Task(
    subagent_type="Plan",
    prompt=f"""Create structured plan for: {user_request}

Return tasks in this format:
{{
    "tasks": [
        {{"content": "[imperative verb] [object]", "activeForm": "[present continuous]"}},
        ...
    ]
}}
"""
)

# AUTO-INIT TodoWrite from Plan output
tasks = [
    {
        "content": task["content"],
        "activeForm": task["activeForm"],
        "status": "pending"
    }
    for task in plan_result.tasks
]

TodoWrite(tasks)

# STEP 3: Execute tasks
for task in plan_result.tasks:
    TodoWrite([{"content": task["content"], "status": "in_progress"}])
    execute_task(task)
    TodoWrite([{"content": task["content"], "status": "completed"}])
```

## Real-World Example

### User Request
```
/alfred:1-plan "Create user authentication feature"
```

### Step 1: Intent Understanding
- Clarify: Is JWT or OAuth preferred?
- Ask for scope: Frontend, backend, or both?

### Step 2: Plan Creation
Plan agent returns:
```json
{
    "tasks": [
        {"content": "Design authentication architecture", "activeForm": "Designing authentication architecture"},
        {"content": "Create database schema", "activeForm": "Creating database schema"},
        {"content": "Write API endpoints", "activeForm": "Writing API endpoints"},
        {"content": "Build frontend login form", "activeForm": "Building frontend login form"},
        {"content": "Write integration tests", "activeForm": "Writing integration tests"}
    ]
}
```

Alfred AUTO-INITIALIZES:
```
☑ Design authentication architecture [pending]
☑ Create database schema [pending]
☑ Write API endpoints [pending]
☑ Build frontend login form [pending]
☑ Write integration tests [pending]
```

### Step 3: Task Execution
```
☑ Design authentication architecture [completed]
☑ Create database schema [completed]
☑ Write API endpoints [in_progress]
...
```

## Benefits

1. **Consistency**: All workflows start with same TodoWrite structure
2. **Visibility**: Users see full task list upfront
3. **Progress tracking**: Clear pending → in_progress → completed flow
4. **Resumability**: Saved TodoWrite state enables resume feature (Feature 6)
5. **No manual work**: Plan agent output directly initializes todos

## Backward Compatibility

**Existing workflows** that manually call TodoWrite still work:
```python
# Old style still works
TodoWrite([
    {"content": "Task 1", "status": "pending"},
    {"content": "Task 2", "status": "pending"}
])
```

**Transition strategy**:
1. New commands use AUTO-INIT pattern (Week 1)
2. Existing commands continue supporting manual TodoWrite
3. Deprecation notice added in v0.9.0
4. Full migration in v1.0.0

## Testing Checklist

- [ ] Plan agent returns well-formed task list
- [ ] TodoWrite initializes with all tasks as "pending"
- [ ] Task execution updates status correctly
- [ ] Resume feature works with auto-initialized todos
- [ ] Backward compatibility maintained for manual TodoWrite calls

## Migration Guide for Users

**Before (v0.7.0)**:
- Manual TodoWrite initialization
- Inconsistent task structures
- No automatic initialization

**After (v0.8.0+)**:
- Automatic TodoWrite from Plan agent
- Consistent structure across all commands
- Better progress visibility
- Supports resume feature

**How to upgrade**:
1. No action required (automatic)
2. All 3 commands (`/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`) now support auto-init
3. Manual TodoWrite calls still work (backward compatible)
