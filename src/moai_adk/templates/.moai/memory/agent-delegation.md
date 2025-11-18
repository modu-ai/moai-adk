# Advanced Agent Delegation Patterns

**Advanced Task() delegation, chaining, context passing, and session management strategies for MoAI-ADK.**

> **See also**: CLAUDE.md → "Agent & Skill Orchestration (개요)" for quick overview

---

## 🔗 Agent Chaining & Orchestration

### Sequential Workflow

Use output from previous step as input to next step:

```python
# Step 1: Requirements gathering
requirements = await Task(
    subagent_type="spec-builder",
    prompt="Create SPEC for user authentication feature"
)
# Returns: SPEC-001 document with requirements

# Step 2: Implementation (depends on SPEC)
implementation = await Task(
    subagent_type="tdd-implementer",
    prompt=f"Implement {requirements.spec_id} using TDD approach"
)
# Uses SPEC from step 1

# Step 3: Database design (independent)
schema = await Task(
    subagent_type="database-expert",
    prompt="Design schema for user authentication data"
)

# Step 4: Documentation (uses all previous)
docs = await Task(
    subagent_type="docs-manager",
    prompt=f"""
    Create documentation for:
    - SPEC: {requirements.spec_id}
    - Implementation: {implementation.files}
    - Database schema: {schema.tables}
    """
)
```

### Parallel Execution (Independent tasks)

```python
import asyncio

# Run independent tasks simultaneously
results = await asyncio.gather(
    Task(
        subagent_type="frontend-expert",
        prompt="Design authentication UI component"
    ),
    Task(
        subagent_type="backend-expert",
        prompt="Design authentication API endpoints"
    ),
    Task(
        subagent_type="database-expert",
        prompt="Design user authentication schema"
    )
)

# Extract results
ui_design, api_design, db_schema = results
# All completed in parallel, much faster!
```

### Conditional Branching

```python
# Decision-based workflow
initial_analysis = await Task(
    subagent_type="plan",
    prompt="Analyze this codebase for refactoring opportunities"
)

if initial_analysis.complexity == "high":
    # Complex refactoring - use multiple agents
    spec = await Task(subagent_type="spec-builder", prompt="...")
    code = await Task(subagent_type="tdd-implementer", prompt="...")
else:
    # Simple refactoring - direct implementation
    code = await Task(
        subagent_type="frontend-expert",
        prompt="Refactor this component"
    )
```

---

## 📦 Context Passing Strategies

### Explicit Context Passing

Pass required context explicitly to each agent:

```python
# Rich context with constraints
task_context = {
    "project_type": "web_application",
    "tech_stack": ["React", "FastAPI", "PostgreSQL"],
    "constraints": ["mobile_first", "WCAG accessibility", "performance"],
    "timeline": "2 weeks",
    "budget": "limited",
    "team_size": "2 engineers"
}

result = await Task(
    subagent_type="spec-builder",
    prompt="Create SPEC for payment processing",
    context=task_context
)
# Agent tailor specifications to constraints
```

### Implicit Context (Alfred manages automatically)

Context automatically collected by Alfred:

```
✅ Project structure from .moai/config.json
✅ Language stack from pyproject.toml/package.json
✅ Existing SPEC documents
✅ Recent commits and changes
✅ Team guidelines from CLAUDE.md
✅ Project conventions and patterns
```

### Session State Management

```python
# Maintain state across multiple agent calls
session = TaskSession()

# First agent: Research phase
research = await session.execute_task(
    subagent_type="mcp-context7-integrator",
    prompt="Research React 19 patterns",
    save_session=True
)

# Second agent: Uses research context
implementation = await session.execute_task(
    subagent_type="frontend-expert",
    prompt="Implement React component",
    context_from_previous=research
)
```

---

## 🔄 Context7 MCP Agent Resume & Session Sharing

### Agent Resume Pattern

Save agent session during execution and resume from same state later:

```python
# Session 1: Start research (Day 1)
research_session = await Task(
    subagent_type="mcp-context7-integrator",
    prompt="Research authentication best practices",
    save_session=True
)
# Session saved to .moai/sessions/research-session-001

# Session 2: Resume research (Day 2)
continued_research = await Task(
    subagent_type="mcp-context7-integrator",
    prompt="Continue researching authorization patterns",
    resume_session="research-session-001"
)
# Picks up where it left off!
```

### Agent Session Sharing

Use output from one agent in another agent:

```python
# Agent 1: Research phase
research = await Task(
    subagent_type="mcp-context7-integrator",
    prompt="Research database optimization techniques",
    save_session=True
)

# Agent 2: Uses research results
optimization = await Task(
    subagent_type="database-expert",
    prompt="Based on research findings, optimize our schema",
    shared_context=research.context,
    shared_session=research.session_id
)

# Agent 3: Documentation (uses both)
docs = await Task(
    subagent_type="docs-manager",
    prompt="Document optimization process and results",
    references=[research.session_id, optimization.session_id]
)
```

### Multi-Day Project Pattern

```python
# Day 1: Planning
plan = await Task(
    subagent_type="plan",
    prompt="Plan refactoring of authentication module",
    save_session=True
)

# Day 2: Implementation (resume planning context)
code = await Task(
    subagent_type="tdd-implementer",
    prompt="Implement refactored authentication",
    resume_session=plan.session_id
)

# Day 3: Testing & Documentation
tests = await Task(
    subagent_type="quality-gate",
    prompt="Test authentication refactoring",
    references=[plan.session_id, code.session_id]
)
```

### Context7 MCP Configuration

**.claude/mcp.json**:

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"],
      "env": {
        "CONTEXT7_SESSION_STORAGE": ".moai/sessions/",
        "CONTEXT7_CACHE_SIZE": "1GB",
        "CONTEXT7_SESSION_TTL": "30d"
      }
    }
  }
}
```

---

## Best Practices

### ✅ Do's

- ✅ Break complex tasks into logical agent stages
- ✅ Use parallel execution for independent work
- ✅ Pass context explicitly when constraints matter
- ✅ Save sessions for multi-day projects
- ✅ Monitor token usage with agent delegation

### ❌ Don'ts

- ❌ Chain too many agents sequentially (overhead grows)
- ❌ Load full codebase into every agent (inefficient)
- ❌ Ignore context window warnings
- ❌ Pass unnecessary information in context
- ❌ Skip session initialization for complex workflows

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: English
**Scope**: MoAI-ADK Advanced Agent Delegation Patterns
**MCP Integration**: Context7 session management (latest)
