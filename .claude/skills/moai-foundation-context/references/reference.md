# moai-foundation-context Reference

## API Reference

### Token Budget Monitoring API

Functions:
- `monitor_token_budget(context_usage: int)`: Real-time usage monitoring
- `get_usage_percent()`: Returns current usage as percentage (0-100)
- `trigger_emergency_compression()`: Compress context when critical
- `defer_non_critical_context()`: Move non-essential context to cache

Thresholds:
- 60%: Monitor and track growth patterns
- 75%: Warning - start progressive disclosure
- 85%: Critical - trigger emergency compression

### Session State API

Session State Structure:
- `session_id`: Unique identifier (UUID v4)
- `model`: Current model identifier
- `created_at`: ISO 8601 timestamp
- `context_window`: Token usage statistics
- `persistence`: Recovery configuration
- `work_state`: Current task state

State Management Functions:
- `create_session_snapshot()`: Capture current state
- `restore_session_state(snapshot)`: Restore from snapshot
- `validate_session_state(state)`: Verify state integrity

### Handoff Protocol API

Handoff Package Structure:
- `handoff_id`: Unique transfer identifier
- `from_agent`: Source agent type
- `to_agent`: Destination agent type
- `session_context`: Token and model information
- `task_context`: Current work state
- `recovery_info`: Checkpoint and fork data

Validation Functions:
- `validate_handoff(package)`: Verify package integrity
- `can_agents_cooperate(from, to)`: Check compatibility

---

## Configuration Options

### Token Budget Allocation

Default Allocation (200K total):
- System Prompt and Instructions: 15K tokens (7.5%)
- Active Conversation: 80K tokens (40%)
- Reference Context: 50K tokens (25%)
- Reserve (Emergency): 55K tokens (27.5%)

Customizable Settings:
- `system_prompt_budget`: Override system allocation
- `conversation_budget`: Override conversation allocation
- `reference_budget`: Override reference allocation
- `reserve_budget`: Override emergency reserve

### Clear Execution Settings

Mandatory Clear Points:
- After `/moai:1-plan` completion
- Context exceeds 150K tokens
- Conversation exceeds 50 messages
- Before major phase transitions
- Model switches (Haiku to Sonnet)

Configuration Options:
- `auto_clear_enabled`: Enable automatic clearing
- `clear_threshold_tokens`: Token threshold for auto-clear
- `clear_threshold_messages`: Message count threshold
- `preserve_on_clear`: Context types to preserve

### Session Persistence Settings

State Layers Configuration:
- L1: Context-Aware Layer (model features)
- L2: Active Context (current task)
- L3: Session History (recent actions)
- L4: Project State (SPEC progress)
- L5: User Context (preferences)
- L6: System State (tools, permissions)

Persistence Options:
- `auto_load_history`: Restore previous context
- `context_preservation`: Preservation level
- `cache_enabled`: Enable context caching

---

## Integration Patterns

### Plan-Run-Sync Workflow Integration

Workflow Sequence:
1. `/moai:1-plan` execution
2. `/clear` (mandatory - saves 45-50K tokens)
3. `/moai:2-run SPEC-XXX`
4. Multi-agent handoffs
5. `/moai:3-sync SPEC-XXX`
6. Session state persistence

Token Savings:
- Post-plan clear: 45-50K tokens saved
- Progressive disclosure: 30-40% reduction
- Handoff optimization: 15-20K per transfer

### Multi-Agent Coordination

Handoff Workflow:
1. Source agent completes task phase
2. Create handoff package with minimal context
3. Validate handoff integrity
4. Target agent receives and validates
5. Target agent continues workflow

Context Minimization:
- Include only SPEC ID and key requirements
- Limit architecture summary to 200 characters
- Exclude background and reasoning
- Transfer critical state only

### Progressive Disclosure Integration

Loading Tiers:
- Tier 1: CLAUDE.md, config.json (always loaded)
- Tier 2: Current SPEC and implementation files
- Tier 3: Related modules and dependencies
- Tier 4: Reference documentation (on-demand)

Disclosure Triggers:
- Explicit user request
- Error recovery requirement
- Complex implementation need
- Documentation reference needed

---

## Troubleshooting

### Context Overflow Issues

Symptoms: Degraded performance, incomplete responses

Solutions:
1. Execute `/clear` immediately
2. Reduce loaded context tiers
3. Apply progressive summarization
4. Split task across multiple sessions

Prevention:
- Monitor at 60% threshold
- Clear after major milestones
- Use aggressive clearing strategy

### Session Recovery Failures

Symptoms: Lost state after interruption

Solutions:
1. Verify session ID was persisted
2. Check snapshot integrity
3. Restore from most recent checkpoint
4. Rebuild state from project files

Prevention:
- Create checkpoints before operations
- Persist session ID before clearing
- Enable auto-save for state snapshots

### Handoff Validation Errors

Symptoms: Agent transition failures

Solutions:
1. Verify available tokens exceed 30K
2. Check agent compatibility
3. Reduce handoff package size
4. Trigger context compression before transfer

Prevention:
- Validate before creating package
- Include only critical context
- Reserve tokens for handoff overhead

### Token Budget Exhaustion

Symptoms: Forced interruptions, emergency behavior

Solutions:
1. Execute immediate `/clear`
2. Resume with Tier 1 context only
3. Load additional context incrementally
4. Split remaining work across sessions

Prevention:
- Maintain 55K emergency reserve
- Execute clear at 85% threshold
- Apply progressive disclosure consistently

---

## External Resources

### Related Documentation

- Token Management Best Practices
- Session State Architecture Guide
- Multi-Agent Coordination Patterns
- Context Optimization Strategies

### Module Files

Advanced Documentation:
- `modules/token-budget-allocation.md`: Budget breakdown and strategies
- `modules/session-state-management.md`: State layers and persistence
- `modules/context-optimization.md`: Progressive disclosure and summarization
- `modules/handoff-protocols.md`: Inter-agent communication
- `modules/memory-mcp-optimization.md`: Memory file structure

### Performance Metrics

Target Metrics:
- Token Efficiency: 60-70% reduction through clearing
- Context Overhead: Less than 15K for system metadata
- Handoff Success Rate: Greater than 95%
- Session Recovery: Less than 5 seconds
- Memory Files: Less than 500 lines each

### Related Skills

- `moai-foundation-claude`: Claude Code authoring and configuration
- `moai-foundation-core`: Core execution patterns and SPEC workflow
- `moai-workflow-project`: Project management and documentation
- `moai-cc-memory`: Memory management and persistence

---

Version: 3.0.0
Last Updated: 2025-12-06
