---
name: "moai-alfred-session-state"
version: "4.0.0"
tier: Alfred
created: 2025-11-05
updated: 2025-11-12
status: stable
description: Enterprise session state management, token budget optimization, runtime tracking, session handoff protocols, context continuity for Claude Sonnet 4.5 and Haiku 4.5 with context awareness features
keywords: ['session', 'state', 'context-window', 'token-budget', 'handoff', 'continuity', 'context-awareness']
allowed-tools: 
  - Read
  - Bash
  - TodoWrite
---

# Alfred Session State Management Skill (Enterprise v4.0.0)

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-session-state |
| **Version** | 4.0.0 (Enterprise) |
| **Updated** | 2025-11-12 |
| **Status** | Active |
| **Tier** | Alfred |
| **Supported Models** | Claude Sonnet 4.5, Claude Haiku 4.5 |
| **Context Window** | 200K tokens (Sonnet/Haiku), 500K tokens (Enterprise), 1M tokens (beta) |
| **Key Features** | Context Awareness, Token Budget Tracking, Session Persistence, Adaptive Recovery |

---

## What It Does

Provides enterprise-grade session state management for extended workflows, token budget optimization, runtime tracking, and handoff protocols to maintain context continuity across Alfred workflows and session boundaries.

**Enterprise v4.0.0 Capabilities**:
- ✅ Context-aware token budget management (November 2025 Claude API features)
- ✅ Session persistence with automatic history loading
- ✅ Session forking for parallel exploration
- ✅ Incremental multi-pack index optimization (Git 2.47+ integration)
- ✅ Context continuity across handoffs with state snapshots
- ✅ Progressive disclosure for memory efficiency
- ✅ Adaptive recovery checkpoints
- ✅ Multi-agent coordination protocols
- ✅ Memory file state synchronization
- ✅ Token budget awareness callbacks (Sonnet/Haiku 4.5 feature)

---

## When to Use

**Automatic triggers**:
- Session start/end events
- Long-running task execution (>10 minutes)
- Multi-agent handoffs
- Context window approaching limits
- Model switches (Haiku ↔ Sonnet)
- Workflow phase transitions

**Manual reference**:
- Session state debugging and recovery
- Token budget optimization strategies
- Handoff protocol design
- Context continuity planning
- Multi-session workflow design

---

## Token Budget Management (November 2025)

### Context Awareness Feature

Claude Sonnet 4.5 and Haiku 4.5 feature **built-in context awareness**, enabling these models to:
- Track remaining context window ("token budget") throughout conversation
- Understand current position within 200K token limit (Sonnet/Haiku)
- Execute adaptive strategies based on available tokens
- Automatically manage context without manual intervention

**Key Advantage**: Models now self-regulate context usage in real time.

### Token Budget Optimization Framework

```
Token Allocation Strategy (200K Sonnet context):
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

### Optimization Techniques (v4.0.0)

**Technique 1: Progressive Summarization**
```
Step 1: Original context (50K tokens)
Step 2: Compress to summary (15K tokens)
Step 3: Add pointers to original → 35K tokens saved
Step 4: Carry forward summary only across handoffs
```

**Technique 2: Context Tagging with Unique Identifiers**
```

❌ Bad (high token cost):
"The user configuration from the previous 20 messages..."

✅ Good (efficient reference):
```

**Technique 3: Structured Context Architecture**
```
├── Critical Context (ALWAYS keep)
│   ├── Current task objectives
│   ├── User preferences & expertise level
│   └── Active constraints
├── Supporting Context (keep if space allows)
│   ├── Related history
│   ├── Reference documentation
│   └── Tool availability
└── Temporary Context (discard when not needed)
    ├── Raw tool outputs
    ├── Intermediate calculations
    └── Debug information
```

**Technique 4: MCP Server Context Budget**
```bash
# Check MCP server context consumption
/context

# Result: Each enabled MCP server adds tool definitions
# Example: context7 MCP = ~2K tokens for tool definitions

# Optimization: Disable unused servers before critical tasks
# Typical savings: 5-10K tokens per unused MCP server
```

**Technique 5: Task-Based Session Management**
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

---

## Session State Architecture (Enterprise v4.0.0)

### State Layers

```
Session State Stack (Enterprise v4.0.0):
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

### Session Creation & Persistence

**Agent SDK Session Management** (November 2025 API):
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
    resume=session_id,  # Continue from checkpoint
    messages=[new_message]
)

# Or: Fork session for parallel exploration
response = claude.messages.create(
    model="claude-sonnet-4-5-20250929",
    fork_session=session_id,  # Branch from checkpoint
    messages=[alternative_message]
)
```

---

## Runtime State Tracking

### Task State Machine (Enterprise v4.0.0)

```
Workflow State Transitions:

pending → in_progress → blocked (waiting) → completed/failed
             ↓                                    ↓
         [monitor token budget]          [save checkpoint]
         [track elapsed time]            [update history]
         [check for recovery]            [archive state]
```

**Task Lifecycle States**:
- `pending` - Queued but not started
- `in_progress` - Currently executing (monitor tokens)
- `blocked` - Waiting for dependencies or input
- `token_warning` - Approaching context limit (v4.0.0)
- `context_switch` - Model change or session fork
- `completed` - Finished successfully
- `failed` - Error occurred, initiating recovery
- `recovered` - Resumed from checkpoint

### Token Budget Callbacks (Haiku/Sonnet 4.5 Feature)

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

---

## Session Handoff Protocols

### Inter-Agent Handoff Package (Enterprise v4.0.0)

```json
{
  "handoff_id": "uuid-v4",
  "timestamp": "2025-11-12T10:30:00Z",
  "from_agent": "spec-builder",
  "to_agent": "tdd-implementer",
  "session_context": {
    "session_id": "sess_uuid",
    "model": "claude-sonnet-4-5-20250929",
    "context_position": 42.5,
    "available_tokens": 115000,
    "user_language": "ko",
    "expertise_level": "intermediate",
    "current_project": "MoAI-ADK"
  },
  "task_context": {
    "spec_id": "SPEC-001",
    "current_phase": "implementation",
    "completed_steps": ["spec_complete", "architecture_defined"],
    "next_step": "write_tests",
    "constraints": ["must_use_pytest", "coverage_85"]
  },
  "context_snapshot": {
    "critical_context": "...compressed...",
    "session_checkpoints": [...],
    "active_todos": [...],
    "token_budget_strategy": "progressive_summarization"
  },
  "recovery_info": {
    "last_checkpoint": "2025-11-12T10:25:00Z",
    "recovery_tokens_reserved": 55000,
    "session_fork_available": true
  }
}
```

### Handoff Validation (Enterprise v4.0.0)

```python
def validate_handoff(handoff_package):
    """Enterprise validation with token budget check"""
    
    required_fields = [
        'handoff_id', 'from_agent', 'to_agent',
        'session_context', 'task_context', 'context_snapshot'
    ]
    
    for field in required_fields:
        if field not in handoff_package:
            raise HandoffError(f"Missing required field: {field}")
    
    # NEW v4.0.0: Validate token budget
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

---

## Session Recovery (Enterprise v4.0.0)

### Recovery Checkpoints

**Checkpoint Triggers**:
- Task phase boundaries (before RED, GREEN, REFACTOR)
- Agent handoffs
- User interruptions
- Token budget thresholds
- Error conditions
- Session timeouts

**Checkpoint Structure**:
```json
{
  "checkpoint_id": "ckpt_uuid",
  "timestamp": "2025-11-12T10:30:00Z",
  "phase": "GREEN",
  "token_usage": {
    "used": 85000,
    "available": 115000
  },
  "context_snapshot": "...compressed snapshot...",
  "session_id": "sess_uuid",
  "recovery_tokens_reserved": 55000
}
```

### Recovery Process (Enterprise v4.0.0)

1. **State Restoration** - Reload last valid checkpoint
2. **Context Validation** - Verify token budget sufficient
3. **Session Resumption** - Use Agent SDK resume feature
4. **Progress Assessment** - Determine what was completed
5. **Continuation Planning** - Decide next steps with updated token budget
6. **User Notification** - Inform user of recovery status

---

## Memory State Synchronization

### Memory Files (v4.0.0)

**Files**:
- `.moai/sessions/session-state.json` - Current session metadata
- `.moai/sessions/context-cache.json` - Cached context for performance
- `.moai/sessions/checkpoints/` - Saved recovery checkpoints
- `.moai/sessions/token-usage.log` - Token budget history
- `active-tasks.md` - TodoWrite task tracking

**Synchronization Protocol**:
```python
def sync_memory_files(session_state):
    """Ensure memory files reflect current session state"""
    
    # Update session metadata with token info
    update_session_metadata({
        'session_id': session_state.id,
        'token_usage': session_state.token_budget,
        'context_position': session_state.context_position,
        'last_sync': timestamp()
    })
    
    # Sync TodoWrite tasks
    sync_todowrite_tasks(session_state.active_tasks)
    
    # Update context cache (compressed)
    update_context_cache(compress_context(session_state.context))
    
    # Log token usage for analytics
    log_token_usage({
        'timestamp': timestamp(),
        'used': session_state.tokens_used,
        'available': session_state.tokens_available,
        'percent': session_state.usage_percent
    })
    
    # Archive old checkpoints (>7 days)
    archive_old_checkpoints()
```

---

## Best Practices (Enterprise v4.0.0)

### Context Management

✅ **DO**:
- Use context-aware token budget tracking (Sonnet/Haiku 4.5 feature)
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

### Token Budget Optimization

✅ **DO**:
- Start new session per distinct task (fresh 200K tokens)
- Use /context to identify expensive MCP servers
- Compress context before handoffs
- Keep reserve buffer (25-30% of tokens)
- Monitor usage percent, not absolute tokens
- Enable auto-summarization at 75% threshold

❌ **DON'T**:
- Let single conversation exceed 150K tokens
- Keep all MCP servers enabled if not needed
- Reuse sessions across fundamentally different tasks
- Ignore available_tokens feedback
- Store uncompressed context in memory files

---

## Configuration

**Location**: `.moai/config/config.json`

```json
{
  "session": {
    "persistence_enabled": true,
    "auto_checkpoint": true,
    "checkpoint_interval_minutes": 10,
    "recovery_strategy": "progressive_summarization",
    "context_budget": {
      "warning_threshold_percent": 75,
      "emergency_threshold_percent": 85,
      "reserve_tokens": 55000
    },
    "forking_enabled": true,
    "max_parallel_sessions": 3
  }
}
```

---

## Debugging & Troubleshooting

### Inspection Tools

```bash
# View current session state
/alfred:debug --show-session-state

# Check context window position
/alfred:debug --show-token-budget

# View all checkpoints
/alfred:debug --list-checkpoints

# Validate memory synchronization
/alfred:debug --check-memory-sync

# Show token usage history
/alfred:debug --show-token-usage-log
```

### Common Issues

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Lost session | Cannot resume conversation | Check .moai/sessions/ for session IDs |
| Token budget exceeded | Model stops responding | Use /context to identify heavy consumers, create new session |
| Handoff failed | Agent has wrong context | Verify handoff package completeness before transfer |
| Recovery stuck | Cannot continue after interruption | Restore from earlier checkpoint or start new session |
| Memory drift | Inconsistent information | Run sync_memory_files() or check cache integrity |

---

## Version History

| Version | Date | Key Changes |
|---------|------|-------------|
| **4.0.0** | 2025-11-12 | Enterprise context awareness, token budget tracking, Session SDK integration, Git 2.47+ support |
| 1.1.0 | 2025-11-05 | Session state management foundation |
| 1.0.0 | 2025-10-01 | Initial release |

---

## Related Skills

- `moai-alfred-context-budget` - Token optimization deep dive
- `moai-alfred-agent-guide` - Multi-agent coordination
- `moai-foundation-trust` - State validation principles
- `moai-foundation-git` - Git session state tracking

---

Learn more in `reference.md` for detailed implementation guides, recovery procedures, advanced coordination patterns, and November 2025 API examples.

**Skill Status**: Production Ready | Last Updated: 2025-11-12 | Model Support: Sonnet 4.5, Haiku 4.5 | Enterprise v4.0.0
