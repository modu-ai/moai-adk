---
name: "moai-cc-subagent-lifecycle"
version: "2.0.43"
created: 2025-11-18
updated: 2025-11-18
status: stable
tier: specialization
description: "Claude Code v2.0.43 SubagentStart/Stop Hook patterns for agent lifecycle management, context optimization, and performance tracking. Complete guide for multi-agent orchestration with session management."
allowed-tools: "Read, Bash, Edit, Task"
primary-agent: "implementation-planner"
secondary-agents: ["tdd-implementer", "backend-expert", "performance-engineer"]
keywords: ["claude-code", "hooks", "agents", "lifecycle", "context-optimization", "performance-tracking"]
tags: ["claude-code-v2.0.43", "advanced"]
orchestration:
  multi_agent: true
  supports_chaining: true
can_resume: true
typical_chain_position: "early"
depends_on: ["moai-cc-hooks", "moai-cc-permission-mode"]
---

# moai-cc-subagent-lifecycle

**Claude Code v2.0.43 Agent Lifecycle Management with SubagentStart/Stop Hooks**

> **Primary Agent**: implementation-planner
> **Secondary Agents**: tdd-implementer, backend-expert, performance-engineer
> **Version**: 2.0.43
> **Keywords**: hooks, agents, lifecycle, context-optimization, performance-tracking

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (50 lines)

**Core Purpose**: Manage agent lifecycle from initialization to completion using SubagentStart and SubagentStop Hooks with automatic context optimization and performance metrics.

**Two Key Hooks**:

#### SubagentStart Hook
- **When**: Before agent execution begins
- **Parameters**: `agentId`, `agentName`, `agentTranscriptPath`
- **Purpose**: Optimize context (max_tokens, priority_files, auto_load_skills)
- **Output**: Context optimization strategy + system message

#### SubagentStop Hook
- **When**: After agent execution completes
- **Parameters**: `agentId`, `agentName`, `agentTranscriptPath`, `executionTime`, `success`
- **Purpose**: Track lifecycle, record metrics, maintain session state
- **Output**: Performance data + completion status

**Quick Hook Implementation**:

```python
# SubagentStart: Optimize context BEFORE execution
def optimize_context(agent_name: str, agent_id: str) -> dict:
    strategies = {
        "spec-builder": {"max_tokens": 20000, "priority_files": [".moai/specs/"]},
        "tdd-implementer": {"max_tokens": 30000, "priority_files": ["src/", "tests/"]},
        "backend-expert": {"max_tokens": 30000, "priority_files": ["src/", "pyproject.toml"]},
    }
    return {
        "continue": True,
        "systemMessage": f"🎯 Optimizing context for {agent_name}"
    }

# SubagentStop: Track lifecycle AFTER execution
def track_lifecycle(
    agent_id: str,
    agent_name: str,
    execution_time_ms: int,
    success: bool
) -> dict:
    metrics = {
        "agent_id": agent_id,
        "agent_name": agent_name,
        "execution_time_ms": execution_time_ms,
        "success": success,
        "completed_at": datetime.now().isoformat(),
    }
    return {
        "continue": True,
        "systemMessage": f"✅ {agent_name} completed in {execution_time_ms/1000:.1f}s"
    }
```

**Quick Usage Pattern**:

```bash
# Hook files in project:
# .claude/hooks/alfred/subagent_start__context_optimizer.py
# .claude/hooks/alfred/subagent_stop__lifecycle_tracker.py

# Automatic execution by Claude Code:
# 1. User calls /alfred:2-run SPEC-001
# 2. SubagentStart Hook triggered → Context optimized
# 3. Agent executes with optimal context
# 4. SubagentStop Hook triggered → Metrics recorded
```

---

### Level 2: Core Implementation (150 lines)

**SubagentStart Hook Deep Dive**

#### Purpose & Timing

SubagentStart Hook executes **before** agent initialization:

```
User calls /alfred:2-run SPEC-001
         ↓
SubagentStart Hook triggered
  - Load agent metadata (agentId, agentName)
  - Determine context strategy
  - Calculate max_tokens, priority_files
  - Select auto_load_skills
  - Write metadata to .moai/logs/agent-transcripts/
         ↓
Agent initializes with optimized context
         ↓
Agent executes task
```

#### Agent-Specific Context Strategies

Each agent gets tailored context optimization:

```python
context_strategies = {
    # Lightweight agents (SPEC generation, documentation)
    "spec-builder": {
        "max_tokens": 20000,
        "priority_files": [".moai/specs/", ".moai/config/config.json"],
        "auto_load_skills": True
    },

    # Code implementation (TDD cycle)
    "tdd-implementer": {
        "max_tokens": 30000,
        "priority_files": ["src/", "tests/", "pyproject.toml"],
        "auto_load_skills": True
    },

    # Architecture & design (backend/frontend)
    "backend-expert": {
        "max_tokens": 30000,
        "priority_files": ["src/", "pyproject.toml"],
        "auto_load_skills": True
    },
    "frontend-expert": {
        "max_tokens": 25000,
        "priority_files": ["src/components/", "src/pages/", "package.json"],
        "auto_load_skills": True
    },

    # Security analysis (broad scope)
    "security-expert": {
        "max_tokens": 50000,
        "priority_files": ["src/", "tests/", ".moai/config/"],
        "auto_load_skills": True
    },

    # Database design
    "database-expert": {
        "max_tokens": 20000,
        "priority_files": [".moai/docs/schema/", "migrations/", "pyproject.toml"],
        "auto_load_skills": True
    },

    # Documentation generation
    "docs-manager": {
        "max_tokens": 15000,
        "priority_files": [".moai/specs/", "README.md", "src/"],
        "auto_load_skills": True
    },
}
```

#### Metadata Recording

SubagentStart records agent initialization metadata:

```python
metadata = {
    "agent_id": "abc-123-def",          # Unique per execution
    "agent_name": "tdd-implementer",    # Agent type
    "started_at": "2025-11-18T10:30:00",
    "strategy": "TDD implementation - code/tests only",
    "max_tokens": 30000,
    "auto_load_skills": True,
    "priority_files": ["src/", "tests/"],
}

# Saved to:
# .moai/logs/agent-transcripts/agent-{agent_id}.json
```

---

**SubagentStop Hook Deep Dive**

#### Purpose & Timing

SubagentStop Hook executes **after** agent completes:

```
Agent finishes execution
         ↓
SubagentStop Hook triggered
  - Collect execution metrics (execution_time_ms)
  - Record completion status (success)
  - Update metadata with finish time
  - Append performance record to JSONL file
         ↓
Next phase/agent or workflow completion
```

#### Performance Metrics Collection

SubagentStop captures 5 key metrics:

```python
performance_record = {
    "timestamp": "2025-11-18T10:35:45",     # When completed
    "agent_id": "abc-123-def",              # Same as start
    "agent_name": "tdd-implementer",        # Agent type
    "execution_time_ms": 312456,            # Duration in milliseconds
    "success": True,                        # Success status
}

# Appended to JSONL (one record per line):
# .moai/logs/agent-performance.jsonl
```

#### Session State Persistence

Agent transcripts enable session resumption:

```python
metadata_update = {
    "agent_id": "abc-123-def",
    "agent_name": "tdd-implementer",
    "transcript_path": "/path/to/transcript.json",
    "execution_time_ms": 312456,
    "execution_time_seconds": 312.456,
    "success": True,
    "completed_at": "2025-11-18T10:35:45",
    "status": "completed"  # or "failed"
}

# Persisted to:
# .moai/logs/agent-transcripts/agent-abc-123-def.json
```

---

### Level 3: Advanced Patterns (200+ lines)

**Multi-Agent Workflows with Lifecycle Management**

#### Pattern 1: Sequential Workflow (Dependent Agents)

```python
# Day 1: SPEC Creation
spec_result = await Task(
    subagent_type="spec-builder",
    prompt="Create SPEC for user authentication"
)
# SubagentStart: spec-builder context optimized
# Agent executes
# SubagentStop: SPEC-001 completed, 45 seconds
#
# Metadata saved:
# .moai/logs/agent-transcripts/agent-spec-001.json
# Performance recorded:
# .moai/logs/agent-performance.jsonl → 45000ms

# Day 2: Implementation (depends on SPEC)
impl_result = await Task(
    subagent_type="tdd-implementer",
    prompt=f"Implement {spec_result.spec_id} using TDD",
    context={"spec_document": spec_result.content}
)
# SubagentStart: tdd-implementer with SPEC context
# Agent executes with optimized token budget
# SubagentStop: Implementation completed, 523 seconds
#
# Metadata saved:
# .moai/logs/agent-transcripts/agent-impl-001.json
# Performance recorded: 523000ms
```

#### Pattern 2: Parallel Workflow (Independent Agents)

```python
# All agents run simultaneously
results = await asyncio.gather(
    Task(subagent_type="frontend-expert", prompt="Design auth UI"),
    Task(subagent_type="backend-expert", prompt="Design auth API"),
    Task(subagent_type="database-expert", prompt="Design user schema")
)

# Parallel execution timeline:
# ┌─ SubagentStart: frontend-expert (max_tokens: 25000)
# │  └─ Agent: Design React auth component (120 seconds)
# │  └─ SubagentStop: Success, 120000ms
# │
# ├─ SubagentStart: backend-expert (max_tokens: 30000)
# │  └─ Agent: Design FastAPI auth endpoints (150 seconds)
# │  └─ SubagentStop: Success, 150000ms
# │
# └─ SubagentStart: database-expert (max_tokens: 20000)
#    └─ Agent: Design user/role schema (90 seconds)
#    └─ SubagentStop: Success, 90000ms
#
# Total time: 150 seconds (vs 360 sequential)
# 58% faster!
```

#### Pattern 3: Conditional Branching

```python
# Initial analysis
analysis = await Task(
    subagent_type="plan",
    prompt="Analyze codebase complexity"
)
# SubagentStart → plan context optimized
# Agent analyzes
# SubagentStop → Analysis recorded

if analysis.complexity == "high":
    # Complex path: Use multiple specialized agents
    spec = await Task(subagent_type="spec-builder", ...)
    impl = await Task(subagent_type="tdd-implementer", ...)
    security = await Task(subagent_type="security-expert", ...)
    # Each triggers SubagentStart/Stop lifecycle
else:
    # Simple path: Direct implementation
    impl = await Task(subagent_type="frontend-expert", ...)
    # Single agent execution
```

#### Pattern 4: Session Resume/Share

```python
# Session 1: Day 1 - Research phase
research = await Task(
    subagent_type="mcp-context7-integrator",
    prompt="Research authentication patterns",
    save_session=True
)
# SubagentStart: research context optimized
# Agent researches (300 seconds)
# SubagentStop: Research saved
# Transcript: .moai/logs/agent-transcripts/agent-research-001.json

# Session 2: Day 2 - Continue research
continued_research = await Task(
    subagent_type="mcp-context7-integrator",
    prompt="Continue researching authorization patterns",
    resume_session="agent-research-001"
)
# SubagentStart: Loads previous session context
# Agent continues from previous state (200 seconds)
# SubagentStop: Session resumed and completed
```

---

## 🎯 Real-World Examples

### Example 1: Full-Stack Feature Implementation

```python
# User command: /alfred:2-run SPEC-PAYMENT-001

# Step 1: SPEC Loading
SubagentStart: plan
  ├─ Load: SPEC-PAYMENT-001
  ├─ max_tokens: 20000
  └─ strategy: "SPEC verification"
Agent: Verify SPEC completeness
SubagentStop: Verified, 30 seconds

# Step 2: Backend Implementation
SubagentStart: backend-expert
  ├─ Load: src/, tests/, pyproject.toml
  ├─ max_tokens: 30000
  └─ strategy: "Stripe API integration"
Agent: Implement payment endpoints
SubagentStop: 6 files created, 450 seconds

# Step 3: Frontend Implementation (parallel)
SubagentStart: frontend-expert
  ├─ Load: src/components/, package.json
  ├─ max_tokens: 25000
  └─ strategy: "Payment UI components"
Agent: Implement payment form
SubagentStop: 3 components created, 280 seconds

# Step 4: Database Design (parallel)
SubagentStart: database-expert
  ├─ Load: migrations/, schema/
  ├─ max_tokens: 20000
  └─ strategy: "Transaction table design"
Agent: Design payment schema
SubagentStop: Migration created, 90 seconds

# Total: ~450 seconds (parallel) vs 820 sequential
# 45% faster with lifecycle optimization!
```

### Example 2: Performance Metrics Analysis

```python
# View agent performance across project

# .moai/logs/agent-performance.jsonl:
{"timestamp": "2025-11-18T10:30:15", "agent_name": "spec-builder", "execution_time_ms": 45000}
{"timestamp": "2025-11-18T10:31:00", "agent_name": "tdd-implementer", "execution_time_ms": 450000}
{"timestamp": "2025-11-18T10:38:30", "agent_name": "frontend-expert", "execution_time_ms": 280000}
{"timestamp": "2025-11-18T10:38:35", "agent_name": "backend-expert", "execution_time_ms": 450000}
{"timestamp": "2025-11-18T10:39:05", "agent_name": "database-expert", "execution_time_ms": 90000}

# Analysis:
# - Fastest: database-expert (90s)
# - Slowest: tdd-implementer (450s)
# - Total: 1,315 seconds sequential
# - Parallel (best): 450 seconds (66% reduction!)
# - Efficiency: Agents used = 5, Max parallel = 3
```

---

## 🔧 Configuration & Setup

### Hook File Locations

```
.claude/hooks/alfred/
├── subagent_start__context_optimizer.py   (SubagentStart)
└── subagent_stop__lifecycle_tracker.py    (SubagentStop)
```

### settings.json Hook Configuration

```json
{
  "hooks": {
    "SubagentStart": {
      "type": "command",
      "command": "python3 .claude/hooks/alfred/subagent_start__context_optimizer.py",
      "model": "haiku",
      "timeout_ms": 5000
    },
    "SubagentStop": {
      "type": "command",
      "command": "python3 .claude/hooks/alfred/subagent_stop__lifecycle_tracker.py",
      "model": "haiku",
      "timeout_ms": 5000
    }
  }
}
```

### Metadata Directory Structure

```
.moai/logs/
├── agent-transcripts/              # Agent session metadata
│   ├── agent-spec-001.json        # Spec-builder execution
│   ├── agent-impl-001.json        # TDD-implementer execution
│   └── agent-db-001.json          # Database-expert execution
├── agent-performance.jsonl        # Performance metrics (append-only)
└── agent-execution.log            # Execution log
```

---

## 📊 Best Practices

### ✅ Do's

- ✅ Use SubagentStart to optimize context **before** agent runs
- ✅ Record all metrics in SubagentStop for later analysis
- ✅ Use agent-specific max_tokens (not one-size-fits-all)
- ✅ Enable auto_load_skills for specialized agents
- ✅ Run independent agents in parallel (3-5x faster)
- ✅ Save session state for multi-day projects
- ✅ Monitor performance.jsonl for bottlenecks

### ❌ Don'ts

- ❌ Chain too many agents sequentially (overhead adds up)
- ❌ Use high max_tokens for simple agents (waste)
- ❌ Ignore agent execution metrics (miss optimization opportunities)
- ❌ Skip session persistence for complex workflows
- ❌ Load full codebase in every agent context
- ❌ Forget to use parallel execution for independent tasks

---

## 🔗 Related Skills

- **moai-cc-hooks** - General Hook architecture
- **moai-cc-permission-mode** - Agent permission strategy
- **moai-cc-hook-model-strategy** - Hook model selection (Haiku/Sonnet)
- **moai-alfred-session-management** - Session `/clear` patterns
- **moai-essentials-perf** - Performance optimization

---

**Last Updated**: 2025-11-18
**Version**: 2.0.43
**Status**: Production Ready
