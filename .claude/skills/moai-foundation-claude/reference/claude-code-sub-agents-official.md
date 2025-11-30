# Claude Code Sub-agents - Official Documentation Reference

Source: https://code.claude.com/docs/en/sub-agents

## Key Concepts

### What are Claude Code Sub-agents?

Sub-agents are specialized AI assistants that can be invoked to handle specific tasks or domains. They extend Claude Code's capabilities through delegation patterns.

### Sub-agent Limitations

Critical Constraint: Sub-agents cannot spawn other sub-agents. This is a fundamental limitation to prevent infinite recursion.

Required Pattern: All sub-agent delegation must use the `Task()` function:
```python
# CORRECT - Required delegation pattern
Task(subagent_type="specialized-agent", prompt="Handle specific task")
```

### Sub-agent Configuration

Required Fields:
```yaml
---
name: agent-name
description: Clear description of agent's purpose and domain
tools: Read, Write, Bash, Grep # Required: Available tools
model: sonnet # Model choice: sonnet, opus, haiku, inherit
permissionMode: default # Permission handling strategy
skills: skill1, skill2 # Available skills for agent
---
```

### Model Selection

- sonnet: Balanced performance and quality (default)
- opus: Highest quality, higher cost
- haiku: Fastest, most cost-effective
- inherit: Use same model as calling agent

### Permission Modes

Official Claude Code permissionMode values (verified from official documentation):

- default: Standard permission prompts. Each tool usage requires user approval on first use in the session. This is the default value if permissionMode is omitted. Recommended for most agents requiring user oversight.

- acceptEdits: Automatically accepts file edit operations (Write, Edit, MultiEdit) without prompting. Other tool operations (like Bash) still require approval. Use for trusted development environments where file modifications are expected and safe.

- dontAsk: Suppresses all permission dialog prompts for this agent. All tool operations proceed without user confirmation. Use with caution - recommended ONLY for automated environments, CI/CD pipelines, or fully sandboxed contexts.

Important Notes:
- If `permissionMode` is omitted, the default value is `default`
- The values `bypassPermissions`, `plan`, and `ignore` are NOT official permissionMode values
- For maximum security, use `default` and explicitly approve each operation
- For automation workflows in safe environments, `acceptEdits` or `dontAsk` may be appropriate

### Tool Permissions

Security Principle: Apply least privilege - only grant tools necessary for agent's domain.

Common Tool Categories:
- Read Tools: `Read`, `Grep`, `Glob` (file system access)
- Write Tools: `Write`, `Edit`, `MultiEdit` (file modification)
- System Tools: `Bash` (command execution)
- Communication Tools: `AskUserQuestion`, `WebFetch` (interaction)

### Sub-agent Creation Patterns

Domain-Specific Specialization:
```yaml
---
name: security-expert
description: OWASP security analysis and vulnerability assessment
tools: Read, Grep, Bash, WebFetch
model: sonnet
skills: moai-quality-security
---
```

Task-Oriented Specialization:
```yaml
---
name: spec-builder
description: Generate EARS format specifications from requirements
tools: Read, Write, Bash
model: sonnet
skills: moai-foundation-specs
---
```

### Best Practices

1. Clear Scope Definition: Single domain or responsibility
2. Specific System Prompts: Clear instructions about agent boundaries
3. Security Boundaries: Explicit limitations on agent behavior
4. Error Handling: Robust error recovery and fallback strategies
5. Performance Considerations: Appropriate model selection and context management

### Integration Patterns

Sequential Delegation:
```python
# Phase 1: Analysis
analysis = Task(
 subagent_type="spec-builder",
 prompt="Analyze requirements for user authentication"
)

# Phase 2: Implementation (passes analysis results)
implementation = Task(
 subagent_type="backend-expert",
 prompt="Implement authentication API based on analysis",
 context={"analysis": analysis}
)
```

Parallel Delegation:
```python
# Independent execution
results = await Promise.all([
 Task(subagent_type="backend-expert", prompt="Backend implementation"),
 Task(subagent_type="frontend-expert", prompt="Frontend implementation"),
 Task(subagent_type="test-engineer", prompt="Test strategy")
])
```

Conditional Delegation:
```python
# Route based on analysis results
if analysis.has_database_issues:
 result = Task(subagent_type="database-expert", prompt="Optimize database")
elif analysis.has_api_issues:
 result = Task(subagent_type="backend-expert", prompt="Fix API issues")
```

### Context Management

Efficient Data Passing:
- Pass only essential information between agents
- Use structured data formats for complex information
- Minimize context size for performance optimization
- Include validation metadata when appropriate

Context Size Limits:
- Each Task() creates independent context window
- Recommended context size: 20K-50K tokens maximum
- Large datasets should be referenced rather than embedded

### File Storage Standards

Agent Definition Location:
```
.claude/agents/
 domain/
 agent-name.md # Agent definition
 examples.md # Usage examples
 integration.md # Integration patterns
 validation.md # Quality checks
 agent-factory.md # Agent creation utilities
```

### Naming Conventions

- Format: lowercase with hyphens (domain-function)
- Length: Maximum 64 characters
- Descriptive: Name should clearly indicate agent's domain
- Unique: Must be unique across all agents

### Testing and Validation

Test Categories:
1. Functionality Testing: Agent performs expected tasks correctly
2. Integration Testing: Agent works properly with other agents
3. Security Testing: Agent respects security boundaries
4. Performance Testing: Agent operates efficiently within token limits

Validation Steps:
1. Test agent behavior with various inputs
2. Verify tool usage respects permissions
3. Validate error handling and recovery
4. Check integration with other agents or skills

### Error Handling

Common Error Types:
- Agent Not Found: Incorrect agent name or file not found
- Permission Denied: Insufficient tool permissions
- Context Overflow: Too much context passed between agents
- Infinite Recursion: Agent tries to spawn another sub-agent

Recovery Strategies:
- Fallback to basic functionality
- User notification with clear error messages
- Graceful degradation of complex features
- Context optimization for retry attempts

### Performance Optimization

Model Selection Guidelines:
- Simple Tasks: Use `haiku` for speed and cost efficiency
- Complex Analysis: Use `sonnet` for quality and reliability
- Creative Tasks: Use `opus` for highest quality output
- Cost-Sensitive: Use `haiku` with careful prompt optimization

Context Optimization:
- Minimize passed context between agents
- Use efficient data structures
- Cache reusable results
- Avoid circular references

### Security Considerations

Access Control:
- Apply principle of least privilege
- Validate all external inputs
- Restrict file system access where appropriate
- Audit tool usage regularly

Data Protection:
- Never pass sensitive credentials
- Sanitize inputs before processing
- Use secure communication channels
- Log agent activities appropriately

### Monitoring and Observability

Activity Logging:
- Track all agent invocations
- Monitor token usage and performance
- Log error conditions and recovery attempts
- Record integration patterns and dependencies

Quality Metrics:
- Agent success rates and error frequencies
- Average response times per task type
- Token efficiency and optimization opportunities
- User satisfaction and feedback collection
