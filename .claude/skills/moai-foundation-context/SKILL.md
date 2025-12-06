---
name: moai-foundation-context
aliases: [moai-foundation-context]
description: Enterprise context and session management with token budget optimization and state persistence
version: 3.0.0
modularized: false
category: foundation
allowed-tools: Read, Grep, Glob
replaces: moai-core-context-budget, moai-core-session-state
---

## Quick Reference (30 seconds)

# Enterprise Context & Session Management

Unified context optimization and session state management for Claude Code with 200K token budget management, session persistence, and multi-agent handoff protocols.

Core Capabilities:
- 200K token budget allocation and monitoring
- Session state tracking with persistence
- Context-aware token optimization
- Multi-agent handoff protocols
- Progressive disclosure and memory management
- Session forking for parallel exploration

When to Use:
- Session initialization and cleanup
- Long-running workflows (>10 minutes)
- Multi-agent orchestration
- Context window approaching limits (>150K tokens)
- Model switches (Haiku ↔ Sonnet)
- Workflow phase transitions

Key Principles (2025):
1. Avoid Last 20% - Performance degrades in final fifth of context
2. Aggressive Clearing - `/clear` every 1-3 messages for SPEC workflows
3. Lean Memory Files - Keep each file < 500 lines
4. Disable Unused MCPs - Minimize tool definition overhead
5. Quality > Quantity - 10% relevant context beats 90% noise

---

## Implementation Guide (5 minutes)

### Features

- Intelligent context window management for Claude Code sessions
- Progressive file loading with priority-based caching
- Token budget tracking and optimization alerts
- Selective context preservation across /clear boundaries
- MCP integration context persistence

### When to Use

- Managing large codebases exceeding 150K token limits
- Optimizing token usage in long-running development sessions
- Preserving critical context across session resets
- Coordinating multi-agent workflows with shared context
- Debugging context-related issues in Claude Code

### Core Patterns

Pattern 1: Progressive File Loading
```python
# Load files by priority tiers
Tier 1: CLAUDE.md, config.json (always loaded)
Tier 2: Current SPEC and implementation files
Tier 3: Related modules and dependencies
Tier 4: Reference documentation (on-demand)
```

Pattern 2: Context Checkpointing
1. Monitor token usage: warn at 150K, critical at 180K
2. Identify essential context to preserve
3. Execute `/clear` to reset session
4. Reload Tier 1 and Tier 2 files automatically
5. Resume work with preserved context

Pattern 3: MCP Context Continuity
```python
# Preserve MCP agent context across /clear
agent_id = mcp_agent.get_id()
# After /clear:
# Context restored through fresh MCP agent initialization
```

## 5 Core Patterns (5-10 minutes each)

### Pattern 1: Token Budget Management (200K Context)

Concept: Strategic allocation and monitoring of 200K token context window.

Allocation Strategy:
```
200K Token Budget Breakdown:
 System Prompt & Instructions: ~15K tokens (7.5%)
 CLAUDE.md: ~8K
 Command definitions: ~4K
 Skill metadata: ~3K
 Active Conversation: ~80K tokens (40%)
 Recent messages: ~50K
 Context cache: ~20K
 Active references: ~10K
 Reference Context (Progressive Disclosure): ~50K (25%)
 Project structure: ~15K
 Related Skills: ~20K
 Tool definitions: ~15K
 Reserve (Emergency Recovery): ~55K tokens (27.5%)
 Session state snapshot: ~10K
 TAGs and cross-references: ~15K
 Error recovery context: ~20K
 Free buffer: ~10K
```

Monitoring Thresholds:
```python
def monitor_token_budget(context_usage: int):
 """Real-time token budget monitoring with automatic actions."""

 usage_percent = (context_usage / 200000) * 100

 if usage_percent > 85:
 # Critical: Emergency compression
 trigger_emergency_compression()
 execute_clear_command()
 elif usage_percent > 75:
 # Warning: Start progressive disclosure
 defer_non_critical_context()
 warn_user_approaching_limit()
 elif usage_percent > 60:
 # Monitor: Track growth patterns
 track_context_growth()
```

Use Case: Prevent context overflow in long-running SPEC-First workflows.

---

### Pattern 2: Aggressive `/clear` Strategy

Concept: Proactive context clearing at strategic checkpoints to maintain efficiency.

Clear Execution Rules:
```
MANDATORY /clear Points:
 After /moai:1-plan completion (saves 45-50K tokens)
 Context > 150K tokens (prevents overflow)
 Conversation > 50 messages (removes stale history)
 Before major phase transitions (clean slate)
 Model switches (Haiku ↔ Sonnet handoffs)
```

Implementation:
```python
def should_execute_clear(context: dict) -> bool:
 """Determine if /clear should be executed."""

 triggers = {
 "post_spec_creation": context.get("spec_created", False),
 "token_threshold": context.get("token_usage", 0) > 150000,
 "message_count": context.get("message_count", 0) > 50,
 "phase_transition": context.get("phase_changed", False)
 }

 return any(triggers.values())
```

Use Case: Maximize token efficiency across SPEC-Run-Sync cycles.

---

### Pattern 3: Session State Persistence

Concept: Maintain session continuity across interruptions with state snapshots.

Session State Architecture:
```
Session State Layers:
 L1: Context-Aware Layer (Claude 4.5+ feature)
 Token budget tracking
 Context window position
 Auto-summarization triggers
 Model-specific optimizations
 L2: Active Context (current task, variables, scope)
 L3: Session History (recent actions, decisions)
 L4: Project State (SPEC progress, milestones)
 L5: User Context (preferences, language, expertise)
 L6: System State (tools, permissions, environment)
```

State Snapshot Structure:
```json
{
 "session_id": "sess_uuid_v4",
 "model": "claude-sonnet-4-5-20250929",
 "created_at": "2025-11-24T10:30:00Z",
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
 "work_state": {
 "current_spec": "SPEC-001",
 "phase": "implementation",
 "completed_steps": ["spec_complete", "architecture_defined"]
 }
}
```

Use Case: Resume long-running tasks after interruptions without context loss.

---

### Pattern 4: Multi-Agent Handoff Protocols

Concept: Seamless context transfer between agents with minimal token overhead.

Handoff Package:
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
 "last_checkpoint": "2025-11-24T10:25:00Z",
 "recovery_tokens_reserved": 55000,
 "session_fork_available": true
 }
}
```

Handoff Validation:
```python
def validate_handoff(handoff_package: dict) -> bool:
 """Validate handoff package integrity."""

 # Token budget check
 available = handoff_package['session_context']['available_tokens']
 if available < 30000: # Minimum safe buffer
 trigger_context_compression()

 # Agent compatibility check
 if not can_agents_cooperate(
 handoff_package['from_agent'],
 handoff_package['to_agent']
 ):
 raise AgentCompatibilityError("Agents cannot cooperate")

 return True
```

Use Case: Efficient Plan → Run → Sync workflow execution.

---

### Pattern 5: Progressive Disclosure & Memory Optimization

Concept: Load context progressively based on relevance and need.

Progressive Summarization:
```python
def progressive_summarization(context: str, target_ratio: float = 0.3):
 """Compress context while preserving key information."""

 # Step 1: Extract key sentences (50K → 15K)
 summary = extract_key_sentences(context, ratio=target_ratio)

 # Step 2: Add pointers to original content
 summary_with_refs = add_content_pointers(summary, context)

 # Step 3: Store original for recovery
 store_original_context(context, "session_archive")

 return summary_with_refs # 35K tokens saved
```

Context Tagging:
```python
# Bad (high token cost):
"The user configuration from the previous 20 messages..."

# Good (efficient reference):
"Refer to @CONFIG-001 for user preferences"
```

Use Case: Maintain context continuity while minimizing token overhead.

---

## Advanced Documentation

For detailed patterns and implementation strategies:

- [Token Budget Allocation](./modules/token-budget-allocation.md) - Budget breakdown, allocation strategies, monitoring thresholds
- [Session State Management](./modules/session-state-management.md) - State layers, persistence, resumption patterns
- [Context Optimization](./modules/context-optimization.md) - Progressive disclosure, summarization, memory management
- [Handoff Protocols](./modules/handoff-protocols.md) - Inter-agent communication, package format, validation
- [Memory & MCP Optimization](./modules/memory-mcp-optimization.md) - Memory file structure, MCP server configuration
- [Reference Guide](./modules/reference.md) - API reference, troubleshooting, best practices

---

## Best Practices

### DO
- Execute `/clear` immediately after SPEC creation
- Monitor token usage and plan accordingly
- Use context-aware token budget tracking
- Create checkpoints before major operations
- Apply progressive summarization for long workflows
- Enable session persistence for recovery
- Use session forking for parallel exploration
- Keep memory files < 500 lines each
- Disable unused MCP servers to reduce overhead

### REQUIREMENTS

[HARD] Maintain bounded context history with regular clearing cycles
WHY: Unbounded context accumulation degrades performance and increases token costs exponentially
IMPACT: Prevents context overflow, maintains consistent response quality, reduces token waste by 60-70%

[HARD] Respond to token budget warnings immediately when usage exceeds 150K tokens
WHY: Operating in the final 20% of context window causes significant performance degradation
IMPACT: Ensures optimal model performance, prevents context overflow failures, maintains workflow continuity

[HARD] Execute state validation checks during session recovery operations
WHY: Invalid state can cause workflow failures and data loss in multi-step processes
IMPACT: Guarantees session integrity, prevents silent failures, enables reliable recovery with >95% success rate

[HARD] Persist session identifiers before any context clearing operations
WHY: Session IDs are the only reliable mechanism for resuming interrupted workflows
IMPACT: Enables seamless workflow resumption, prevents work loss, supports multi-agent coordination

[SOFT] Establish clear session boundaries when working with multiple concurrent sessions
WHY: Session mixing causes context contamination and unpredictable agent behavior
IMPACT: Improves debugging clarity, prevents cross-session interference, maintains clean audit trails

[SOFT] Create checkpoint snapshots before assuming session continuity
WHY: Context can be lost due to network issues, timeouts, or system events
IMPACT: Provides recovery points, reduces rework time, maintains user trust in system reliability

[SOFT] Load codebase components progressively using priority tiers
WHY: Loading entire codebases exhausts token budget and includes irrelevant context
IMPACT: Optimizes token usage by 40-50%, improves response relevance, enables larger project support

[SOFT] Limit handoff packages to critical context only
WHY: Non-critical context increases handoff overhead and reduces available working tokens
IMPACT: Speeds up agent transitions by 30%, preserves token budget for actual work, reduces transfer errors

[HARD] Execute context compression or clearing when usage reaches 85% threshold
WHY: Approaching context limits triggers emergency behaviors and reduces model capabilities
IMPACT: Maintains 55K token emergency reserve, prevents forced interruptions, ensures graceful degradation

---

## Works Well With

- `moai-cc-memory` - Memory management and context persistence
- `moai-cc-configuration` - Session configuration and preferences
- `moai-core-workflow` - Workflow state persistence and recovery
- `moai-cc-agents` - Agent state management across sessions
- `moai-foundation-trust` - Quality gate integration

---

## Workflow Integration

Session Initialization:
```
1. Initialize token budget (Pattern 1)
2. Load session state (Pattern 3)
3. Setup progressive disclosure (Pattern 5)
4. Configure handoff protocols (Pattern 4)
```

SPEC-First Workflow:
```
1. /moai:1-plan execution
 ↓
2. /clear (mandatory - saves 45-50K tokens)
 ↓
3. /moai:2-run SPEC-XXX
 ↓
4. Multi-agent handoffs (Pattern 4)
 ↓
5. /moai:3-sync SPEC-XXX
 ↓
6. Session state persistence (Pattern 3)
```

Context Monitoring:
```
Continuous:
 Track token usage (Pattern 1)
 Apply progressive disclosure (Pattern 5)
 Execute /clear at thresholds (Pattern 2)
 Validate handoffs (Pattern 4)
```

---

## Success Metrics

- Token Efficiency: 60-70% reduction through aggressive clearing
- Context Overhead: <15K tokens for system/skill metadata
- Handoff Success Rate: >95% with validation
- Session Recovery: <5 seconds with state persistence
- Memory Optimization: <500 lines per memory file

---

## Changelog

- v2.0.0 (2025-11-24): Unified moai-core-context-budget and moai-core-session-state into single skill with 5 core patterns
- v1.0.0 (2025-11-22): Original individual skills

---

Status: Production Ready (Enterprise)
Modular Architecture: SKILL.md + 6 modules
Integration: Plan-Run-Sync workflow optimized
Generated with: MoAI-ADK Skill Factory
