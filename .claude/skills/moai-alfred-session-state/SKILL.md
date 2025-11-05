---
name: moai-alfred-session-state
version: 1.1.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: Session state management, runtime state tracking, session handoff protocols, and context continuity for Alfred workflows
keywords: ['session', 'state', 'handoff', 'context', 'continuity', 'tracking']
allowed-tools:
  - Read
  - Bash
  - TodoWrite
---

# Alfred Session State Management Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-session-state |
| **Version** | 1.1.0 (2025-11-05) |
| **Status** | Active |
| **Tier** | Alfred |
| **Purpose** | Manage session state and ensure context continuity |

---

## What It Does

Provides comprehensive session state management, runtime tracking, and handoff protocols to maintain context continuity across Alfred workflows and session boundaries.

**Key capabilities**:
- ✅ Session state tracking and persistence
- ✅ Context continuity across handoffs
- ✅ Runtime state monitoring and management
- ✅ Session cleanup and optimization
- ✅ Multi-agent coordination protocols
- ✅ Memory file state synchronization

---

## When to Use

**Automatic triggers**:
- Session start/end events
- Task switches and context changes
- Multi-agent handoffs
- Long-running workflow interruptions

**Manual reference**:
- Session state debugging
- Handoff protocol design
- Context optimization strategies
- Memory management planning

---

## Session State Architecture

### State Layers

```
Session State Stack:
├── L1: Active Context (current task, variables, scope)
├── L2: Session History (recent actions, decisions, outcomes)  
├── L3: Project State (SPEC progress, milestones, blockers)
├── L4: User Context (preferences, expertise level, language)
└── L5: System State (tool availability, permissions, environment)
```

### State Persistence Pattern

**Active State** (`session-state.json`):
```json
{
  "session_id": "uuid-v4",
  "user_id": "user-context",
  "current_task": {
    "type": "alfred_command",
    "command": "/alfred:2-run",
    "spec_id": "SPEC-001",
    "status": "in_progress",
    "start_time": "2025-11-05T15:30:00Z"
  },
  "context_stack": [...],
  "memory_refs": [...],
  "agent_chain": [...]
}
```

---

## Runtime State Tracking

### Task State Management

**Task Lifecycle States**:
- `pending` - Queued but not started
- `in_progress` - Currently executing
- `blocked` - Waiting for dependencies
- `completed` - Finished successfully
- `failed` - Error occurred
- `cancelled` - User requested stop

**State Transition Rules**:
```python
def update_task_state(task_id, new_state, context):
    """Update task state with validation"""
    
    # Validate transition
    if not is_valid_transition(current_state, new_state):
        raise InvalidStateTransition(f"Cannot transition from {current_state} to {new_state}")
    
    # Update task
    task = get_task(task_id)
    task.state = new_state
    task.updated_at = timestamp()
    task.state_history.append({
        'from': current_state,
        'to': new_state,
        'timestamp': task.updated_at,
        'context': context
    })
    
    # Trigger side effects
    trigger_state_change_hooks(task, context)
```

### Context Continuity

**Context Preservation Rules**:
1. **Critical Context** - Always preserve across handoffs
   - Current task objectives and constraints
   - User preferences and expertise level
   - Recent decisions and rationale
   - Active TODO items and progress

2. **Secondary Context** - Preserve when space allows
   - Historical context and background
   - Related but inactive tasks
   - Reference material links
   - Tool availability and permissions

3. **Temporary Context** - Discard when not needed
   - Raw tool outputs
   - Intermediate calculations
   - Transient variables
   - Debug information

---

## Session Handoff Protocols

### Inter-Agent Handoff

**Handoff Package Structure**:
```json
{
  "handoff_id": "uuid-v4",
  "from_agent": "spec-builder",
  "to_agent": "tdd-implementer", 
  "timestamp": "2025-11-05T15:30:00Z",
  "session_context": {
    "user_language": "ko",
    "expertise_level": "intermediate",
    "current_project": "MoAI-ADK",
    "active_spec": "SPEC-001"
  },
  "task_context": {
    "current_phase": "implementation",
    "completed_steps": ["spec_complete", "architecture_defined"],
    "next_step": "write_tests",
    "constraints": ["must_use_pytest", "coverage_85"]
  },
  "state_snapshot": {...}
}
```

**Handoff Validation**:
```python
def validate_handoff(handoff_package):
    """Ensure handoff contains required context"""
    
    required_fields = [
        'handoff_id', 'from_agent', 'to_agent', 
        'session_context', 'task_context'
    ]
    
    for field in required_fields:
        if field not in handoff_package:
            raise HandoffError(f"Missing required field: {field}")
    
    # Validate agent compatibility
    if not can_agents_cooperate(handoff_package.from_agent, handoff_package.to_agent):
        raise AgentCompatibilityError("Agents cannot cooperate")
    
    return True
```

### Session Recovery

**Recovery Checkpoints**:
- **Task Boundaries** - Before major phase changes
- **Agent Handoffs** - During context transfers  
- **User Interruptions** - When session is paused
- **Error Conditions** - Before exception handling

**Recovery Process**:
1. **State Restoration** - Reload last valid checkpoint
2. **Context Validation** - Verify all required context available
3. **Progress Assessment** - Determine what was completed
4. **Continuation Planning** - Decide next steps
5. **User Notification** - Inform user of recovery status

---

## Memory State Synchronization

### Memory File Coordination

**Memory File States**:
- `session-summary.md` - Current session overview
- `active-tasks.md` - TodoWrite task tracking
- `context-cache.json` - Cached context for performance
- `agent-notes.md` - Agent-specific observations

**Synchronization Protocol**:
```python
def sync_memory_files(session_state):
    """Ensure memory files reflect current session state"""
    
    # Update session summary
    update_session_summary(session_state)
    
    # Sync TodoWrite tasks
    sync_todowrite_tasks(session_state.active_tasks)
    
    # Update context cache
    update_context_cache(session_state.context_stack)
    
    # Archive old context
    archive_old_context(session_state.context_history)
```

---

## State Management Best Practices

✅ **DO**:
- Always update session state on task changes
- Create checkpoints before major operations
- Validate handoff packages before transfers
- Archive old context to manage memory usage
- Monitor state consistency across agents
- Provide recovery mechanisms for failures

❌ **DON'T**:
- Lose context during agent handoffs
- Skip state validation on recovery
- Let memory files become inconsistent
- Ignore failed state transitions
- Accumulate unlimited context history
- Assume session continuity without validation

---

## Debugging Session State

### State Inspection Tools

**Session State Viewer**:
```bash
# View current session state
/alfred:debug --show-session-state

# Check context stack
/alfred:debug --show-context-stack

# Validate memory file consistency
/alfred:debug --check-memory-sync
```

**Common Issues and Solutions**:

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Lost context on handoff | Agent asks redundant questions | Verify handoff package completeness |
| Memory file drift | Inconsistent information across files | Run memory synchronization |
| State corruption | Tasks show wrong status | Restore from last checkpoint |
| Context overflow | Session performance degradation | Archive old context and clean memory |

---

## Performance Optimization

### Context Budget Management

**Optimization Strategies**:
- **Progressive Disclosure** - Load detailed context only when needed
- **Smart Caching** - Cache frequently accessed context
- **Lazy Loading** - Load reference material on demand
- **Context Summarization** - Compress historical context

**Monitoring Metrics**:
- Context usage percentage
- Memory file sizes
- Handoff success rates
- Recovery frequency
- Session performance metrics

---

Learn more in `reference.md` for detailed implementation guides, recovery procedures, and advanced coordination patterns.

**Related Skills**:
- `moai-alfred-context-budget` - Context optimization strategies
- `moai-alfred-agent-guide` - Multi-agent coordination
- `moai-alfred-workflow` - Session workflow management
- `moai-foundation-trust` - State validation principles
