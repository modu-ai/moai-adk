---
name: moai-cc-subagent-lifecycle
description: Claude Code subagent lifecycle, delegation patterns, and multi-agent coordination
allowed-tools: [Read, Bash]
---

# Claude Code Subagent Lifecycle

## Quick Reference

Claude Code subagents enable **parallel, specialized task execution** through an orchestrator-worker pattern. Each subagent runs independently with its own context window, model selection, and tool permissions - preventing context pollution while scaling efficiently from 2-3 agents (simple tasks) to 20-30 agents (complex projects).

**Lifecycle Phases** (November 2025):
- **Spawn**: Main agent creates subagent via `Task()` tool
- **Configure**: Model, context window, temperature, tool permissions
- **Execute**: Subagent performs specialized task in isolation
- **Report**: Results returned to orchestrator
- **Terminate**: Subagent context released after completion
- **SubagentStop Hook**: Fires when subagent finishes (coordination point)

**Built-in Subagents**:
- **Plan**: Multi-step task breakdown (Sonnet for planning, Haiku for execution)
- **Explore**: Fast codebase exploration (Haiku 4.5, pattern search)

**Key Constraint**: Subagents **cannot spawn other subagents** (prevents infinite nesting).

---

## Implementation Guide

### Phase 1: Subagent Architecture Basics

**Custom Subagent Definition** (`.claude/agents/security-expert.md`):
```yaml
---
name: security-expert
description: Performs OWASP Top 10 security audits and vulnerability scanning
tools: Read, Bash(bandit:*), Bash(safety:*)
color: Red
model: sonnet
disallowedTools: [Write, Edit]
---

# Security Expert Agent

## Purpose
You are a security auditing specialist focused on OWASP Top 10 compliance.

## Instructions
1. Read all Python files in the target directory
2. Run bandit for static security analysis
3. Run safety for dependency vulnerability scan
4. Generate comprehensive security report

## Report Format
```markdown
# Security Audit Report

## OWASP Top 10 Analysis
[Findings by category]

## Dependency Vulnerabilities
[CVE list with severity]

## Recommendations
[Prioritized fixes]
```

**Delegation Pattern** (main agent):
```python
# Main agent delegates security audit to subagent
Task(
    agent="security-expert",
    task="Audit /src directory for OWASP Top 10 vulnerabilities"
)
```

### Phase 2: Lifecycle Event Coordination

**SubagentStop Hook** (coordinate multi-agent workflows):
```json
{
  "hooks": {
    "SubagentStop": [
      {
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/subagent_coordination.py"
        }]
      }
    ]
  }
}
```

**Coordination Script** (`.claude/hooks/subagent_coordination.py`):
```python
#!/usr/bin/env python3
import json
import sys
from pathlib import Path

input_data = json.loads(sys.stdin.read())
agent_name = input_data.get('agent', {}).get('name', 'unknown')
result = input_data.get('result', '')
session_id = input_data.get('session', {}).get('id', 'unknown')

# Log subagent completion
log_file = Path(".claude/logs/subagent_execution.jsonl")
log_file.parent.mkdir(parents=True, exist_ok=True)

log_entry = {
    "session_id": session_id,
    "agent": agent_name,
    "status": "completed",
    "result_length": len(result)
}

with open(log_file, 'a') as f:
    f.write(json.dumps(log_entry) + '\n')

# Trigger next phase (example: notify completion)
if agent_name == "security-expert":
    print("Security audit completed - proceeding to fix recommendations", file=sys.stderr)

sys.exit(0)
```

### Phase 3: Advanced Orchestration Patterns

**Parallel Execution** (spawn multiple subagents):
```markdown
# Main agent orchestration

I'll analyze this codebase using parallel subagents:

Task(agent="backend-expert", task="Audit /src/api for performance bottlenecks")
Task(agent="frontend-expert", task="Review /src/components for accessibility")
Task(agent="security-expert", task="Scan /src for vulnerabilities")

[Waits for all 3 subagents to complete, then synthesizes results]
```

**Sequential Coordination** (phase-based workflow):
```markdown
# Phase 1: Exploration
Task(agent="code-explorer", task="Map project structure and dependencies")

# Wait for Phase 1 completion...

# Phase 2: Architecture (uses Phase 1 results)
Task(agent="code-architect", task="Design API refactoring based on exploration findings")

# Wait for Phase 2 completion...

# Phase 3: Implementation
Task(agent="backend-expert", task="Implement architecture design")
```

**Error Recovery Pattern**:
```yaml
---
name: resilient-worker
description: Worker agent with automatic retry on failure
tools: Read, Bash(*), Write
---

# Resilient Worker

## Error Handling
1. Attempt task execution
2. If failure, analyze error message
3. Retry with adjusted approach (max 3 attempts)
4. If still failing, report to orchestrator with diagnostic info

## Success Criteria
- Task completed without errors
- Output validated against acceptance criteria
- Report includes execution metrics
```

---

## Advanced Patterns

### Model Selection Strategy

**Dynamic Model Assignment**:
```yaml
# Opus for complex reasoning
---
name: architect-agent
model: opus
description: High-level architecture design requiring deep reasoning
---

# Sonnet for balanced performance
---
name: implementation-agent
model: sonnet
description: Feature implementation with moderate complexity
---

# Haiku for fast execution
---
name: test-runner-agent
model: haiku
description: Fast test execution and basic validation
---
```

**Cost Optimization** (Haiku 4.5 auto-model selection):
- Planning phase: Auto-escalates to Sonnet
- Execution phase: Runs on Haiku
- 70% cost reduction vs pure Sonnet

### Context Window Management

**Isolation Benefits**:
```
Main Agent Context (200K tokens)
├─ Project overview
├─ User conversation
└─ Coordination logic

Subagent #1 Context (50K tokens)
├─ Focused on /src/api
├─ Performance analysis only
└─ Independent of Subagent #2

Subagent #2 Context (50K tokens)
├─ Focused on /src/components
├─ Accessibility audit only
└─ Independent of Subagent #1

Result: No cross-contamination, efficient context usage
```

**Best Practice**: Limit each subagent to 1-2 focused responsibilities to prevent context bloat.

### Built-in Subagent Patterns

**Plan Mode** (automatic for complex tasks):
```bash
# User asks complex multi-step question
"Refactor authentication system to support OAuth 2.0"

# Claude automatically:
# 1. Activates Plan subagent
# 2. Breaks down into phases:
#    - Phase 1: Analyze current auth system
#    - Phase 2: Design OAuth integration
#    - Phase 3: Implement token management
#    - Phase 4: Update tests
# 3. Presents plan before execution
# 4. Spawns specialized subagents for each phase
```

**Explore Mode** (automatic for codebase queries):
```bash
# User asks: "Where are error handlers implemented?"

# Claude automatically:
# 1. Activates Explore subagent (Haiku 4.5)
# 2. Searches for error handling patterns
# 3. Returns findings without loading entire codebase
# 4. Saves ~80% context tokens
```

---

## Best Practices

### ✅ DO
- Define **single-responsibility agents** (one task per agent)
- Use **disallowedTools** to enforce least privilege
- Implement **SubagentStop hooks** for coordination
- Select **appropriate models** (Opus for reasoning, Haiku for speed)
- Limit **context window** per subagent (10-50K tokens)
- Provide **clear success criteria** in agent instructions
- Log **subagent execution metrics** for monitoring
- Use **parallel execution** for independent tasks

### ❌ DON'T
- Create agents with broad, unfocused responsibilities
- Allow subagents to spawn other subagents (prevented by system)
- Use Opus for simple tasks (cost inefficiency)
- Load entire codebase into subagent context
- Ignore SubagentStop events (missed coordination opportunities)
- Skip error handling in subagent logic
- Forget to document agent capabilities in agent description

---

## Scaling Patterns

**Small Projects** (2-3 subagents):
```
Main Agent
├─ backend-expert (API implementation)
├─ frontend-expert (UI components)
└─ test-expert (test coverage)
```

**Medium Projects** (5-10 subagents):
```
Main Agent
├─ code-explorer (codebase mapping)
├─ code-architect (design)
├─ backend-expert (API)
├─ frontend-expert (UI)
├─ database-expert (schema)
├─ security-expert (audit)
├─ test-expert (coverage)
└─ doc-expert (documentation)
```

**Large Projects** (20-30 subagents in waves):
```
Wave 1: Exploration (parallel)
├─ code-explorer-api
├─ code-explorer-frontend
└─ code-explorer-database

Wave 2: Design (sequential, uses Wave 1 results)
├─ code-architect
└─ security-architect

Wave 3: Implementation (parallel)
├─ backend-expert (5 agents for different modules)
├─ frontend-expert (3 agents for page sections)
└─ integration-expert (2 agents for API/DB)

Wave 4: Validation (parallel)
├─ test-expert
├─ security-expert
└─ performance-expert

Wave 5: Documentation
└─ doc-expert
```

---

## Monitoring & Debugging

**Subagent Execution Log**:
```bash
tail -f .claude/logs/subagent_execution.jsonl
```

**Output Format**:
```json
{
  "session_id": "sess_abc123",
  "agent": "security-expert",
  "status": "completed",
  "duration_ms": 4500,
  "tokens_used": 12000,
  "result_length": 3500
}
```

**Performance Metrics**:
- Average subagent duration: < 10 seconds (target)
- Context usage per subagent: 10-50K tokens
- Parallel speedup: 3-5x vs sequential execution

---

## Works Well With

- `moai-cc-hook-model-strategy` (SubagentStop hook integration)
- `moai-cc-permission-mode` (Tool permission inheritance)
- `moai-alfred-orchestration` (Multi-agent coordination)
- `moai-essentials-perf` (Performance monitoring)

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Status**: Production Ready (Enterprise)  
**Official Reference**: https://docs.anthropic.com/en/docs/claude-code/sub-agents  
**Community Examples**: https://github.com/VoltAgent/awesome-claude-code-subagents
