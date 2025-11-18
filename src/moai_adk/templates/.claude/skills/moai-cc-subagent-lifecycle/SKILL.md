---
name: moai-cc-subagent-lifecycle
version: 4.0.0
updated: 2025-11-19
status: stable
stability: stable
description: Claude Code subagent lifecycle management including initialization, execution, error handling, and cleanup patterns. Covers full lifecycle from startup through resource cleanup.
focus_areas:
  - Subagent initialization and startup
  - Execution lifecycle management
  - Error handling and recovery
  - Resource cleanup and teardown
keywords:
  - subagent
  - lifecycle
  - initialization
  - cleanup
  - resource-management
  - error-handling
allowed-tools: Read, Glob, Bash
---

# Claude Code Subagent Lifecycle Management

## Overview

Subagent lifecycle spans initialization → execution → error handling → cleanup, with proper resource management at each stage. Enterprise-grade patterns for reliable subagent orchestration.

## Lifecycle Architecture

```
INITIALIZATION
  ├─ Context loading
  ├─ State initialization
  └─ Ready for execution
       ↓
EXECUTION
  ├─ Task processing
  ├─ Progress monitoring
  └─ State updates
       ↓
COMPLETION/ERROR
  ├─ Result collection
  ├─ Error handling
  └─ Cleanup trigger
       ↓
CLEANUP
  ├─ Resource release
  ├─ State persistence
  └─ Session closure
```

## Initialization Phase

### Subagent Startup

```python
# Initialize subagent with required context
subagent = Task(
    subagent_type="spec-builder",
    description="Create SPEC for feature",
    prompt="Create specification with EARS format"
)

# Context automatically loaded:
# ✅ Project configuration (.moai/config.json)
# ✅ Memory files (.claude/memory.md)
# ✅ SPEC documents (.moai/specs/)
# ✅ Language preferences (conversation_language)
# ✅ Project conventions (CLAUDE.md)
```

### State Management

```python
# Subagent carries state across calls
session_state = {
    "project_id": "moai-adk",
    "spec_id": "SPEC-001",
    "phase": "requirements",
    "tokens_used": 0,
    "start_time": time.time(),
    "context_loaded": ["SPEC-001.md", "memory.md"],
    "language": "ko"
}

# State persisted between calls
# Enables resumable work
```

### Context Loading Strategy

```python
# Automatic context collection:
context = {
    "project": ProjectConfig.from_path(".moai/config.json"),
    "memory": MemoryFile.load(".claude/memory.md"),
    "specs": SPEC.load_all(".moai/specs/"),
    "language": config.language.conversation_language,
    "standards": CLAUDE.md.parse()
}

# Only load relevant context
# Saves tokens, improves performance
```

## Execution Phase

### Task Execution

```python
# Subagent executes autonomously
result = await Task(
    subagent_type="tdd-implementer",
    prompt="Implement SPEC-001 using TDD cycle",
    timeout=600000,  # 10 minutes
    model="haiku"  # Optional: specify model
)

# Returns result with:
# ✅ Output content (code, docs, etc.)
# ✅ Execution time (milliseconds)
# ✅ Tokens consumed (for budgeting)
# ✅ Status (success/error/timeout)
# ✅ Logs (execution details)
```

### Progress Monitoring

```python
# Monitor subagent execution
while subagent.is_running():
    status = subagent.get_status()
    print(f"Progress: {status.percentage}%")
    print(f"Tokens used: {status.tokens_used}/{budget}")
    print(f"Current phase: {status.phase}")
    
    if status.tokens_used > budget * 0.8:
        print("Warning: Approaching token limit")
    
    time.sleep(5)  # Poll every 5 seconds
```

### Error Handling

```python
try:
    result = await Task(
        subagent_type="backend-expert",
        prompt="Design API endpoints"
    )
except TaskTimeout:
    logger.error("Task exceeded timeout")
    # Implement retry or escalation
    result = await Task(
        subagent_type="backend-expert",
        prompt="Design API endpoints (retry)",
        timeout=900000  # Increase timeout
    )
except TaskError as e:
    logger.error(f"Subagent error: {e}")
    # Implement recovery logic
    notify_user(f"Task failed: {e.message}")
    # Optionally retry with different parameters
except OutOfMemory:
    logger.error("Subagent context exceeded")
    # Break task into smaller subtasks
    subtask1 = Task(subagent_type="...", prompt="Part 1")
    subtask2 = Task(subagent_type="...", prompt="Part 2")
```

## State Persistence Phase

### Result Collection

```python
# Collect and validate results
if result.status == "success":
    # Validate output
    if result.validate():
        logger.info(f"Task completed: {result.spec_id}")
        return result
    else:
        logger.error(f"Output validation failed")
        # Implement fallback
else:
    logger.error(f"Task failed with status: {result.status}")
    # Implement error recovery
```

### Session Persistence

```python
# Save session for resumability
session_data = {
    "task_id": result.task_id,
    "session_id": result.session_id,
    "state": result.state,
    "progress": result.progress,
    "timestamp": time.time()
}

# Persist to disk
with open(f".moai/sessions/{session_id}.json", "w") as f:
    json.dump(session_data, f)
```

## Cleanup Phase

### Resource Cleanup

```python
# Explicit cleanup after task
try:
    result = await Task(...)
    logger.info("Task completed successfully")
finally:
    # Cleanup happens automatically:
    # ✅ Agent context released
    # ✅ Temporary files removed
    # ✅ Connections closed
    # ✅ State synchronized
    cleanup_resources()
```

### Session Closure

```python
# Close agent session explicitly
if subagent.session_id:
    try:
        await subagent.close_session()
        logger.info(f"Session {subagent.session_id} closed")
        # Releases all resources
    except Exception as e:
        logger.error(f"Error closing session: {e}")
```

### Memory Management

```python
# Clear temporary context
subagent.clear_temp_context()

# Persist important state
subagent.save_session_state()

# Update memory file if needed
memory.update(subagent.learnings)

# Log final metrics
logger.info(f"Tokens used: {subagent.total_tokens}")
logger.info(f"Duration: {subagent.execution_time}s")
```

## Best Practices

### Initialization

- ✅ Load project configuration first
- ✅ Validate context before execution
- ✅ Set appropriate timeouts (10-30 minutes typical)
- ✅ Specify model explicitly (Sonnet/Haiku)

### Execution

- ✅ Implement error handling and retries
- ✅ Monitor token usage continuously
- ✅ Log execution milestones
- ✅ Track progress percentage

### Cleanup

- ✅ Always cleanup resources (use try/finally)
- ✅ Persist important state
- ✅ Close sessions properly
- ✅ Update memory files
- ✅ Log final metrics

### Testing

- ✅ Test with different context sizes
- ✅ Validate error handling paths
- ✅ Verify resource cleanup
- ✅ Monitor for memory leaks

## Common Patterns

### Sequential Workflow with State Passing

```python
# Task 1: Requirements gathering
spec = await Task(
    subagent_type="spec-builder",
    prompt="Create SPEC for authentication",
    save_session=True
)

# Task 2: Implementation using spec context
code = await Task(
    subagent_type="tdd-implementer",
    prompt="Implement authentication",
    context_from_previous=spec
)

# Task 3: Testing using code context
tests = await Task(
    subagent_type="test-engineer",
    prompt="Test authentication implementation",
    context_from_previous=code
)
```

### Parallel Execution with Resource Management

```python
import asyncio

# Launch parallel tasks
tasks = [
    Task(subagent_type="frontend-expert", ...),
    Task(subagent_type="backend-expert", ...),
    Task(subagent_type="database-expert", ...)
]

# Execute in parallel
results = await asyncio.gather(*tasks)

# Cleanup all sessions
for result in results:
    if result.session_id:
        await result.close_session()
```

## TRUST 5 Compliance

- **Test-First**: Lifecycle patterns validated with real agents
- **Readable**: Clear initialization, execution, cleanup structure
- **Unified**: Consistent lifecycle across all subagent types
- **Secured**: Proper resource cleanup, no memory leaks, secure state persistence
- **Trackable**: Full execution logging and state tracking

## Related Skills

- `moai-cc-hook-model-strategy` - Hook execution during lifecycle
- `moai-cc-permission-mode` - Permission control throughout lifecycle
- `moai-core-workflow` - Overall workflow orchestration

---

**Last Updated**: 2025-11-19
**Version**: 4.0.0
**Enterprise Production Ready**: Yes ✅
**Maturity**: Stable
