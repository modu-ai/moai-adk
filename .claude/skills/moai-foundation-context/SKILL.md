---
name: moai-foundation-context
description: >
  Manages context window optimization, session state persistence, and token budget
  allocation for multi-agent workflows.
  Use when dealing with token budget management, context window limits, session handoff,
  state persistence across agents, or /clear strategies.
  Do NOT use for agent orchestration patterns
  (use moai-foundation-core instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "3.1.0"
  category: "foundation"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "foundation, context, session, token-optimization, state-management, multi-agent"
  aliases: "moai-foundation-context"
  replaces: "moai-core-context-budget, moai-core-session-state"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    - "token"
    - "context"
    - "session"
    - "budget"
    - "optimization"
    - "handoff"
    - "state"
    - "memory"
    - "/clear"
    - "context window"
    - "token limit"
    - "session persistence"
    - "context management"
    - "multi-agent"
  agents:
    - "manager-spec"
    - "manager-ddd"
    - "manager-strategy"
    - "manager-quality"
    - "manager-docs"
    - "manager-project"
  phases:
    - "plan"
    - "run"
    - "sync"
---

## Quick Reference

Enterprise Context and Session Management - Unified context optimization and session state management for Claude Code with 200K token budget management, session persistence, and multi-agent handoff protocols.

Core Capabilities:

- 200K token budget allocation and monitoring
- Session state tracking with persistence
- Context-aware token optimization
- Multi-agent handoff protocols
- Progressive disclosure and memory management
- Session forking for parallel exploration

When to Use:

- Session initialization and cleanup
- Long-running workflows exceeding 10 minutes
- Multi-agent orchestration
- Context window approaching limits exceeding 150K tokens
- Model switches between Haiku and Sonnet
- Workflow phase transitions

Key Principles:

Avoid Last 20%: Performance degrades in final fifth of context window.

Aggressive Clearing: Execute /clear every 1-3 messages for SPEC workflows.

Lean Memory Files: Keep each file under 500 lines.

Disable Unused MCPs: Minimize tool definition overhead.

Quality Over Quantity: 10% relevant context beats 90% noise.

---

## Implementation Guide

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

Pattern 1 - Progressive File Loading:

Load files by priority tiers. Tier 1 includes CLAUDE.md and config.json which are always loaded. Tier 2 includes current SPEC and implementation files. Tier 3 includes related modules and dependencies. Tier 4 includes reference documentation loaded on-demand.

Pattern 2 - Context Checkpointing:

Monitor token usage with warning at 150K and critical at 180K. Identify essential context to preserve. Execute /clear to reset session. Reload Tier 1 and Tier 2 files automatically. Resume work with preserved context.

Pattern 3 - MCP Context Continuity:

Preserve MCP agent context across /clear by storing the agent_id. After /clear, context is restored through fresh MCP agent initialization.

## Core Patterns Detail

### Pattern 1: Token Budget Management

Concept: Strategic allocation and monitoring of 200K token context window.

Budget Breakdown: System Prompt and Instructions take approximately 15K tokens at 7.5%, including CLAUDE.md at 8K, Command definitions at 4K, and Skill metadata at 3K. Active Conversation takes approximately 80K tokens at 40%, including Recent messages at 50K, Context cache at 20K, and Active references at 10K. Reference Context with Progressive Disclosure takes approximately 50K at 25%, including Project structure at 15K, Related Skills at 20K, and Tool definitions at 15K. Reserve for Emergency Recovery takes approximately 55K tokens at 27.5%, including Session state snapshot at 10K, TAGs and cross-references at 15K, Error recovery context at 20K, and Free buffer at 10K.

Monitoring Thresholds: When usage exceeds 85%, trigger emergency compression and execute clear command. When usage exceeds 75%, defer non-critical context and warn user of approaching limit. When usage exceeds 60%, track context growth patterns.

Use Case: Prevent context overflow in long-running SPEC-First workflows.

### Pattern 2: Aggressive /clear Strategy

Concept: Proactive context clearing at strategic checkpoints to maintain efficiency.

Mandatory /clear Points: After /moai:1-plan completion to save 45-50K tokens. When context exceeds 150K tokens to prevent overflow. When conversation exceeds 50 messages to remove stale history. Before major phase transitions for clean slate. During model switches for Haiku to Sonnet handoffs.

Use Case: Maximize token efficiency across SPEC-Run-Sync cycles.

### Pattern 3: Session State Persistence

Concept: Maintain session continuity across interruptions with state snapshots.

Session State Layers: L1 is the Context-Aware Layer for Claude 4.5+ with token budget tracking, context window position, auto-summarization triggers, and model-specific optimizations. L2 is Active Context for current task, variables, and scope. L3 is Session History for recent actions and decisions. L4 is Project State for SPEC progress and milestones. L5 is User Context for preferences, language, and expertise. L6 is System State for tools, permissions, and environment.

Use Case: Resume long-running tasks after interruptions without context loss.

### Pattern 4: Multi-Agent Handoff Protocols

Concept: Seamless context transfer between agents with minimal token overhead.

Handoff Package Contents: Include handoff_id, from_agent, to_agent, session_context with session_id, model, context_position, available_tokens, and user_language, task_context with spec_id, current_phase, completed_steps, and next_step, and recovery_info with last_checkpoint, recovery_tokens_reserved, and session_fork_available.

Handoff Validation: Check token budget with minimum 30K available buffer. Verify agent compatibility. Trigger context compression if needed.

Use Case: Efficient Plan to Run to Sync workflow execution.

### Pattern 5: Progressive Disclosure and Memory Optimization

Concept: Load context progressively based on relevance and need.

Progressive Summarization: Extract key sentences to compress 50K to 15K at target ratio of 0.3. Add pointers to original content for reference. Store original in session archive for recovery. Result saves approximately 35K tokens.

Context Tagging: Avoid high token cost phrases like "The user configuration from the previous 20 messages..." and use efficient references like "Refer to @CONFIG-001 for user preferences".

Use Case: Maintain context continuity while minimizing token overhead.

---

## Advanced Documentation

For detailed patterns and implementation strategies:

- modules/token-budget-allocation.md - Budget breakdown, allocation strategies, monitoring thresholds
- modules/session-state-management.md - State layers, persistence, resumption patterns
- modules/context-optimization.md - Progressive disclosure, summarization, memory management
- modules/handoff-protocols.md - Inter-agent communication, package format, validation
- modules/memory-mcp-optimization.md - Memory file structure, MCP server configuration
- modules/reference.md - API reference, troubleshooting, best practices

---

## Best Practices

Recommended Practices:

- Execute /clear immediately after SPEC creation
- Monitor token usage and plan accordingly
- Use context-aware token budget tracking
- Create checkpoints before major operations
- Apply progressive summarization for long workflows
- Enable session persistence for recovery
- Use session forking for parallel exploration
- Keep memory files under 500 lines each
- Disable unused MCP servers to reduce overhead

Required Practices:

Maintain bounded context history with regular clearing cycles. Unbounded context accumulation degrades performance and increases token costs exponentially. This prevents context overflow, maintains consistent response quality, and reduces token waste by 60-70%.

Respond to token budget warnings immediately when usage exceeds 150K tokens. Operating in the final 20% of context window causes significant performance degradation.

Execute state validation checks during session recovery operations. Invalid state can cause workflow failures and data loss in multi-step processes.

Persist session identifiers before any context clearing operations. Session IDs are the only reliable mechanism for resuming interrupted workflows.

Execute context compression or clearing when usage reaches 85% threshold. This maintains 55K token emergency reserve and prevents forced interruptions.

---

## Works Well With

- moai-cc-memory - Memory management and context persistence
- moai-cc-configuration - Session configuration and preferences
- moai-core-workflow - Workflow state persistence and recovery
- moai-cc-agents - Agent state management across sessions
- moai-foundation-trust - Quality gate integration

---

## Workflow Integration

Session Initialization: Initialize token budget with Pattern 1, load session state with Pattern 3, setup progressive disclosure with Pattern 5, configure handoff protocols with Pattern 4.

SPEC-First Workflow: Execute /moai:1-plan, then mandatory /clear to save 45-50K tokens, then /moai:2-run SPEC-XXX, then multi-agent handoffs with Pattern 4, then /moai:3-sync SPEC-XXX, then session state persistence with Pattern 3.

Context Monitoring: Continuously track token usage with Pattern 1, apply progressive disclosure with Pattern 5, execute /clear at thresholds with Pattern 2, validate handoffs with Pattern 4.

---

## Success Metrics

- Token Efficiency: 60-70% reduction through aggressive clearing
- Context Overhead: Less than 15K tokens for system/skill metadata
- Handoff Success Rate: Greater than 95% with validation
- Session Recovery: Less than 5 seconds with state persistence
- Memory Optimization: Less than 500 lines per memory file

---

Status: Production Ready (Enterprise)
Modular Architecture: SKILL.md + 6 modules
Integration: Plan-Run-Sync workflow optimized
Generated with: MoAI-ADK Skill Factory
