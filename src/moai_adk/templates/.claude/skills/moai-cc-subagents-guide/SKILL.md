---
name: moai-cc-subagents-guide
description: Complete guide to Claude Code Sub-Agents creation, configuration, invocation patterns, and advanced features. Use when designing new agents or optimizing agent workflows with official Claude Code standards.
allowed-tools: Read, WebFetch, Grep, Glob
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: subagents, cc, moai, guide  


# Claude Code Sub-Agents Complete Guide

Comprehensive guide to creating, configuring, and orchestrating Sub-Agents following official Claude Code standards.

## Quick Reference

**Sub-Agents** are specialized AI assistants with independent context windows, custom system prompts, and configurable tool access.

```yaml
---
name: agent-name
description: When/why to use this agent
tools: tool1, tool2, tool3  # Optional
model: sonnet  # Optional: sonnet/opus/haiku/inherit
permissionMode: default  # Optional
skills: skill1, skill2  # Optional
---

# Agent System Prompt
[Instructions and guidelines...]
```

## Implementation Guide

### Creation Methods (3 Approaches)

**Method 1: Interactive Command**
```bash
/agents command in Claude Code
↓
Select "Create New Agent"
↓
Define purpose, tools, system prompt
↓
Save (project or user-level)
```
- Pros: Guided interface, immediate feedback
- Cons: Requires manual input
- Best for: Quick prototyping

**Method 2: File-Based (Markdown)**
```
.claude/agents/agent-name.md  (project-level, highest priority)
~/.claude/agents/agent-name.md  (user-level, lower priority)
```
- Pros: Version-controlled, reusable
- Cons: Manual YAML editing
- Best for: Team deployment

**Method 3: CLI --agents Flag**
```bash
claude --agents '[{
  "name": "code-reviewer",
  "tools": "Read, Grep, Bash",
  "model": "sonnet"
}]'
```
- Pros: Session-specific, scripted
- Cons: JSON formatting required
- Best for: Automation, CI/CD

### Configuration Fields Reference

| Field | Required | Type | Default | Purpose |
|-------|----------|------|---------|---------|
| **name** | Yes | string | - | Lowercase kebab-case identifier (max 64 chars) |
| **description** | Yes | string | - | When/why to invoke this agent |
| **tools** | No | string | All tools | Comma-separated list (principle of least privilege) |
| **model** | No | string | inherit | sonnet/opus/haiku/inherit (context decides) |
| **permissionMode** | No | string | default | default/acceptEdits/dontAsk (permission handling) |
| **skills** | No | string | - | Comma-separated skills (auto-loaded) |

### permissionMode Options

**default** (recommended)
- Prompts user before sensitive operations
- Allows granular control
- Secure for general use
- Example: File writes, dangerous bash commands

**acceptEdits**
- Accepts modifications without prompting
- Faster execution
- Less interruption
- Use only for trusted agents

**dontAsk**
- No permission prompts
- Designed for automated workflows
- Requires high trust
- Use for read-only agents

### Model Selection

**sonnet** (Claude Sonnet 4.5)
- Complex reasoning
- Multi-step tasks
- Research & analysis
- Architecture design
- Cost: Higher

**opus** (Claude Opus)
- Not commonly used in agents
- Reserved for exceptional complexity
- Cost: Highest

**haiku** (Claude Haiku 4.5)
- Simple execution
- Formatting tasks
- Code modifications
- Cost: Lowest (70% savings)

**inherit** (recommended)
- Context decides model
- Adapts to task complexity
- Balances cost/quality
- No explicit specification needed

## Advanced Patterns

### Invocation Methods (3 Ways)

**1. Automatic Delegation**
```
User: "I need you to review my code"
↓
Claude recognizes code-reviewer agent capability
↓
Auto-invokes: Task(subagent_type='code-reviewer')
↓
Agent executes without user intervention
```
**Trigger**: Specific terminology + capability match

**2. Explicit Invocation**
```
User: "Use the code-reviewer to check my changes"
↓
Claude explicitly recognizes request
↓
Invokes: Task(subagent_type='code-reviewer')
↓
Agent executes as requested
```
**Trigger**: User explicitly names the agent

**3. Programmatic Invocation (Task API)**
```python
Task(
  subagent_type='code-reviewer',
  description='Review for security issues',
  prompt='Analyze this authentication code...'
)
```
**Trigger**: Orchestration logic via Task()

### Subagent Chaining (Sequential Workflows)

Combine multiple agents for complex tasks:

```
Task 1: Agent A (Design)
  ↓
  Output: Architecture design
  ↓
Task 2: Agent B (Implementation, receives Agent A output)
  ↓
  Output: Implemented code
  ↓
Task 3: Agent C (Security Review, receives Agent B output)
  ↓
  Output: Final approved code
```

**Pattern**: Each agent receives previous agent's output as context.

### Resumable Agents (Persistent Conversations)

Continue agent conversations across sessions:

```
# Session 1: Start analysis
Agent executes, returns result with agent_id

# Session 2: Continue analysis
resume(agent_id='...')  # Continues previous conversation
Agent can access previous context and decisions
Useful for iterative analysis
```

**Use Cases**:
- Long-running analyses
- Multi-day refactoring projects
- Iterative design reviews
- Continuous monitoring tasks

**Benefits**:
- Context preservation across sessions
- Cost efficiency (avoid re-processing)
- Better continuity in complex workflows

### Agent IDs & References

Agents can be referenced by ID:
```
Agent execution returns: agent_id='abc123'
Resume with: resume(agent_id='abc123')
Reference in workflows: uses stored agent_id
```

## Key Benefits

### 1. Context Preservation
- Separate context window per agent
- Prevents main conversation pollution
- Clean state for specialized tasks
- Efficient token usage

### 2. Specialized Expertise
- Custom system prompt per agent
- Focused toolset
- Domain-specific instructions
- Better quality results

### 3. Reusability Across Projects
- Version-controlled agents (.claude/agents/)
- Share across teams
- Standardized patterns
- Consistency across workflows

### 4. Flexible Permissions
- Minimal tool access (principle of least privilege)
- Security-focused designs
- User control via permissionMode
- Transparent capabilities

## Best Practices (6 Core Principles)

**1. Focused Responsibility**
- One primary purpose per agent
- Clear scope boundaries
- Avoid multi-purpose agents
- Example: code-reviewer (not code-reviewer-and-optimizer)

**2. Detailed System Prompts**
- Explicit instructions
- Examples for guidance
- Clear constraints
- Quality requirements

**3. Minimal Tool Access**
- Only necessary tools
- Principle of least privilege
- Security-focused configuration
- Example: code-reviewer needs only Read, Grep

**4. Specific Descriptions**
- Include triggering scenarios
- Define use cases
- Avoid vague language
- Enable auto-discovery

**5. Version Control**
- Commit agents to git (.claude/agents/)
- Document changes
- Use semantic versioning
- Track iterations

**6. Team Testing**
- Test with teammates
- Gather feedback
- Iterate before deployment
- Document improvements

## Common Issues & Solutions

**Issue: Agent Not Auto-Invoked**
- Solution: Make description more specific
- Include trigger keywords
- Test explicit invocation first

**Issue: Permissions Blocking Execution**
- Solution: Review permissionMode setting
- Adjust tool permissions
- Use acceptEdits for trusted agents

**Issue: Agent Performance Degradation**
- Solution: Check context window usage
- Split into smaller agents if chaining
- Review model selection (haiku vs sonnet)

**Issue: Tool Access Denied**
- Solution: Verify tools in configuration
- Check allowed-tools restrictions
- Ensure proper comma separation (no brackets)

## Architecture Considerations

### When to Create an Agent

Create an agent when:
- Specialized domain expertise needed
- Separate context helpful
- Reused across multiple tasks
- Tool restriction beneficial
- Team coordination required

### When NOT to Create an Agent

Don't create an agent for:
- Simple, one-time tasks
- Tasks better handled by current agent
- Operations requiring full tool access
- Low-frequency operations

### Agent Lifecycle

**Development** (Experimental)
- Test locally
- Gather feedback
- Iterate quickly

**Deployment** (Stable)
- Version control
- Team testing
- Documentation
- Deployment

**Maintenance** (Long-term)
- Monitor performance
- Gather user feedback
- Plan improvements
- Update documentation

---

**Version**: 1.0.0
**Updated**: 2025-11-22
**Reference**: https://code.claude.com/docs/en/sub-agents
