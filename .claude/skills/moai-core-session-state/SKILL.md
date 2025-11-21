---
name: moai-core-session-state
description: Enterprise session state management, token budget optimization, runtime tracking
---

## Quick Reference

Enterprise-grade session state management for extended workflows with token budget optimization and handoff protocols to maintain context continuity.

**Core Capabilities**:
- Context-aware token budget management (November 2025 Claude API)
- Session persistence with automatic history loading
- Session forking for parallel exploration
- Context continuity across handoffs with state snapshots
- Progressive disclosure for memory efficiency
- Token budget awareness callbacks (Sonnet/Haiku 4.5)

**When to Use**:
- Session start/end events
- Long-running task execution (>10 minutes)
- Multi-agent handoffs
- Context window approaching limits
- Model switches (Haiku ↔ Sonnet)
- Workflow phase transitions

---

## Implementation Guide

### Token Budget Management (November 2025)

**Context Awareness Feature**:
Claude Sonnet 4.5 and Haiku 4.5 feature **built-in context awareness**, enabling these models to:
- Track remaining context window ("token budget") throughout conversation
- Understand current position within 200K token limit
- Execute adaptive strategies based on available tokens
- Automatically manage context without manual intervention

**Token Allocation Strategy (200K Sonnet context)**:
```
├── System Prompt & Instructions: ~15K tokens (7.5%)
│   ├── CLAUDE.md: ~8K
│   ├── Command definitions: ~4K
│   └── Skill metadata: ~3K
├── Active Conversation: ~80K tokens (40%)
│   ├── Recent messages: ~50K
│   ├── Context cache: ~20K
│   └── Active references: ~10K
├── Reference Context (Progressive Disclosure): ~50K (25%)
│   ├── Project structure: ~15K
│   ├── Related Skills: ~20K
│   └── Tool definitions: ~15K
└── Reserve (Emergency Recovery): ~55K tokens (27.5%)
    ├── Session state snapshot: ~10K
    ├── TAGs and cross-references: ~15K
    ├── Error recovery context: ~20K
    └── Free buffer: ~10K
```

**Optimization Techniques**:

**Technique 1: Progressive Summarization**:
```
Step 1: Original context (50K tokens)
Step 2: Compress to summary (15K tokens)
Step 3: Add pointers to original → 35K tokens saved
Step 4: Carry forward summary only across handoffs
```

**Technique 2: Context Tagging with Unique Identifiers**:
```
❌ Bad (high token cost):
"The user configuration from the previous 20 messages..."

✅ Good (efficient reference):
"Refer to @CONFIG-001 for user preferences"
```

**Technique 3: Task-Based Session Management**:
```
Strategy: Start new conversation for distinct tasks

Benefits:
- Fresh 200K token budget per task
- Eliminates stale context accumulation
- Enables parallel session forking
- Improves recovery speed

Implementation:
1. Complete current task in Session A
2. Save session snapshot to .moai/sessions/
3. Start Session B for new task with fresh context
4. Resume Session A later if needed via session ID
```

### Session State Architecture

**State Layers**:
```
Session State Stack:
├── L1: Context-Aware Layer (Claude 4.5+ feature)
│   ├── Token budget tracking
│   ├── Context window position
│   ├── Auto-summarization triggers
│   └── Model-specific optimizations
├── L2: Active Context (current task, variables, scope)
├── L3: Session History (recent actions, decisions)
├── L4: Project State (SPEC progress, milestones)
├── L5: User Context (preferences, language, expertise)
└── L6: System State (tools, permissions, environment)
```

**Session Creation & Persistence**:
```json
{
  "session_id": "sess_uuid_v4",
  "model": "claude-sonnet-4-5-20250929",
  "created_at": "2025-11-12T10:30:00Z",
  "context_window": {
    "total": 200000,
    "used": 85000,
    "available": 115000,
    "position_percent": 42.5
  },
  "persistence": {
    "auto_load_history": true,
    "context_preservation": "critical_only",
    "cache_enabled": true
  },
  "forking": {
    "enabled": true,
    "fork_session_id": "sess_fork_uuid",
    "checkpoint_timestamp": "2025-11-12T10:30:00Z"
  }
}
```

---

## Advanced Patterns

### Session Resumption Pattern

```python
# Capture session ID from initial response
session_id = extract_session_id(response)

# Save for later use
save_session_checkpoint({
    'session_id': session_id,
    'timestamp': now(),
    'model': 'claude-sonnet-4-5',
    'context_state': current_context_snapshot()
})

# Later: Resume conversation
response = claude.messages.create(
    model="claude-sonnet-4-5-20250929",
    resume=session_id,
    messages=[new_message]
)

# Or: Fork session for parallel exploration
response = claude.messages.create(
    model="claude-sonnet-4-5-20250929",
    fork_session=session_id,
    messages=[alternative_message]
)
```

### Token Budget Callbacks

```python
def token_budget_callback(context):
    """
    Called automatically when token budget changes.
    Model provides real-time context awareness.
    """
    
    remaining_tokens = context.available_tokens
    used_percent = context.token_usage_percent
    
    if used_percent > 85:
        # Activate emergency summarization
        compress_context_window()
        archive_old_context()
        
    elif used_percent > 75:
        # Start progressive disclosure
        defer_non_critical_context()
        
    elif used_percent > 60:
        # Monitor for safety
        track_context_growth()
```

### Session Handoff Protocols

**Inter-Agent Handoff Package**:
```json
{
  "handoff_id": "uuid-v4",
  "from_agent": "spec-builder",
  "to_agent": "tdd-implementer",
  "session_context": {
    "session_id": "sess_uuid",
    "model": "claude-sonnet-4-5-20250929",
    "context_position": 42.5,
    "available_tokens": 115000,
    "user_language": "ko"
  },
  "task_context": {
    "spec_id": "SPEC-001",
    "current_phase": "implementation",
    "completed_steps": ["spec_complete", "architecture_defined"],
    "next_step": "write_tests"
  },
  "recovery_info": {
    "last_checkpoint": "2025-11-12T10:25:00Z",
    "recovery_tokens_reserved": 55000,
    "session_fork_available": true
  }
}
```

**Handoff Validation**:
```python
def validate_handoff(handoff_package):
    """Enterprise validation with token budget check"""
    
    # Validate token budget
    context = handoff_package['session_context']
    available = context['available_tokens']
    if available < 30000:  # Minimum safe buffer
        trigger_context_compression()
    
    # Validate agent compatibility
    if not can_agents_cooperate(
        handoff_package['from_agent'],
        handoff_package['to_agent']
    ):
        raise AgentCompatibilityError("Agents cannot cooperate")
    
    return True
```

### Best Practices

✅ **DO**:
- Use context-aware token budget tracking
- Create checkpoints before major operations
- Apply progressive summarization for long workflows
- Enable session persistence for recovery
- Monitor token usage and plan accordingly
- Use session forking for parallel exploration

❌ **DON'T**:
- Accumulate unlimited context history
- Ignore token budget warnings
- Skip state validation on recovery
- Lose session IDs without saving
- Mix multiple sessions without clear boundaries
- Assume session continuity without checkpoint


---

## Context7 Integration

### Related Libraries & Tools
- [Redis](/redis/redis): In-memory data store
- [Memcached](/memcached/memcached): Distributed cache

### Official Documentation
- [Documentation](https://redis.io/docs/)
- [API Reference](https://redis.io/commands/)

### Version-Specific Guides
Latest stable version: 7.2
- [Release Notes](https://github.com/redis/redis/releases)
- [Migration Guide](https://redis.io/docs/about/releases/)
